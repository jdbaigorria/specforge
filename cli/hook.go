package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// `sf hook` — el motor de enforcement (F29), portado del antiguo
// specforge_enforce.py a Go.
//
// POR QUÉ vive ahora en el CLI: el hook Python re-implementaba lógica que el CLI
// ya tenía (lectura del gate ledger, features activas, cadena de fases) → dos
// fuentes de verdad (F2). Acá reusamos directamente las funciones de Go
// (latestApproveGate, activeStatuses, artifactChain, computeCurrentState,
// buildCurrentContext): una sola verdad. Además los adapters por arnés pasan a
// ser shells finitos sobre `sf hook` — sin dependencia de Python.
//
// Contrato (igual que el .py): lee UN objeto JSON por stdin y emite UNO por
// stdout. Dos modos de salida:
//   --harness=generic     → {"decision":"deny|allow|block","reason":..,"context":..}
//   --harness=claude-code → el shape nativo de Claude (hookSpecificOutput / decision)
//
// Dos propiedades de seguridad, por diseño:
//   - FAIL OPEN: cualquier error interno → allow (nunca bloquear el editing).
//   - NO-OP fuera de SpecForge: si <project>/specforge/ no existe, todo allow.
// ----------------------------------------------------------------------------

// fullSliceEveryDefault: cada cuántos turnos SIN cambio de paso se re-inyecta el
// slice COMPLETO como backstop de saliencia (el resto va solo el breadcrumb).
// Override por env SPECFORGE_FULL_SLICE_EVERY.
const fullSliceEveryDefault = 10

// archivedStatuses: estados en que una feature está CERRADA — sus artefactos no
// se editan más por la vía normal (para eso está `sf-amend`).
//
// `"archived"` estaba acá y era un status FANTASMA: nunca existió en
// validStatuses, así que ninguna feature podía tenerlo y la rama estaba muerta.
// Se reemplaza por los dos estados reales de cierre (DL-5 F2): `retired` y
// `abandoned` cuentan como archivadas a ojos del hook — ninguno de los dos debe
// habilitar edición de artefactos.
var archivedStatuses = map[string]bool{
	"done": true, "retired": true, "abandoned": true,
}

// artifactRe reconoce la ruta-relativa de un artefacto gateable. JSON-first:
// matchea .json (la fuente) Y .md (el render); plan vive bajo progress/.
var artifactRe = regexp.MustCompile(
	`^specforge/features/([^/]+)/(?:progress/)?(requirements|design|tasks|plan)\.(?:json|md)$`,
)

// protectedJSONRe reconoce TODO el estado autoritativo en formato .json que solo
// `sf` puede escribir (Capa 1 de FIXBUGHIGH). Cubre el ledger de features, la
// constitución, el domain, las entradas de journal y los .json de artefacto por
// feature (incluido trace y review, que artifactRe NO matchea). Los .md
// renderizados quedan FUERA a propósito: son prosa, los regenera `sf save`.
var protectedJSONRe = regexp.MustCompile(
	`^specforge/(` +
		`features\.json` +
		`|constitution\.json` +
		`|sources\.json` +
		`|context/domain\.json` +
		`|journal/[^/]+\.json` +
		`|features/[^/]+/(?:progress/)?(?:requirements|design|tasks|plan|review|trace|audit|feature)\.json` +
		`|features/[^/]+/deltas/[^/]+\.json` +
		`)$`,
)

// ── Payload + entry point ────────────────────────────────────────────────────

// hookPayload son los campos que nos interesan del JSON de entrada. Claude manda
// tool_input.file_path / cwd / session_id / stop_hook_active; el contrato
// generic manda event / file_path / project_dir / session_id.
type hookPayload struct {
	Event          string `json:"event"`
	ProjectDir     string `json:"project_dir"`
	Cwd            string `json:"cwd"`
	FilePath       string `json:"file_path"`
	Command        string `json:"command"` // contrato generic: comando de Bash a inspeccionar
	SessionID      string `json:"session_id"`
	StopHookActive bool   `json:"stop_hook_active"`
	HookEventName  string `json:"hook_event_name"`
	ToolInput      struct {
		FilePath string `json:"file_path"`
		Command  string `json:"command"` // Claude Code: tool_input.command para la tool Bash
	} `json:"tool_input"`
}

// runHook es el punto de entrada de `sf hook`. Parsea flags, lee el payload y
// despacha al adapter. Fail-open ante panics (defer/recover): un bug del engine
// jamás debe bloquear al usuario.
func runHook(args []string) (code int) {
	harness, event := "claude-code", ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--harness="):
			harness = strings.TrimPrefix(a, "--harness=")
		case strings.HasPrefix(a, "--event="):
			event = strings.TrimPrefix(a, "--event=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf hook: unknown flag %q\n", a)
			return 2
		}
	}

	// Leemos el payload ANTES del defer para que el recover pueda inspeccionarlo
	// (necesita saber el evento y el proyecto para decidir fail-open vs fail-closed).
	raw, _ := io.ReadAll(os.Stdin)
	var p hookPayload
	if len(strings.TrimSpace(string(raw))) > 0 {
		_ = json.Unmarshal(raw, &p) // payload inválido → struct vacío (fail open)
	}
	event = eventOr(event, p.HookEventName)

	// Política de error del engine (FIXBUGHIGH, cierre de borde):
	//   - FAIL-CLOSED dentro de una feature activa: si el panic ocurre decidiendo
	//     un PreToolUse y hay una feature en curso, DENEGAMOS. Un bug del engine no
	//     debe convertirse en una puerta abierta justo cuando hay estado que proteger.
	//   - FAIL-OPEN en cualquier otro caso (otros eventos, o sin feature activa): un
	//     bug no debe bloquear el editor cuando no hay nada que custodiar.
	defer func() {
		if r := recover(); r != nil {
			if isPreToolUseEvent(harness, event) && hasActiveFeatureSafe(p.projectDir()) {
				fmt.Fprintf(os.Stderr, "sf hook: internal error during an active feature — failing closed (%v)\n", r)
				emitDeny(harness, "SpecForge enforcement hit an internal error while a feature is active, so it is "+
					"failing closed (denying) to avoid an unguarded write. Run `sf doctor` and retry.")
				code = 0 // claude/generic comunican el deny por el JSON, no por exit code
				return
			}
			fmt.Fprintf(os.Stderr, "sf hook: internal error, allowing (%v)\n", r)
			code = 0
		}
	}()

	if harness == "generic" {
		return runHookGeneric(p)
	}
	return runHookClaude(event, p)
}

// projectDir resuelve el dir del proyecto: env de Claude, o cwd/project_dir del
// payload, o el cwd del proceso.
func (p hookPayload) projectDir() string {
	if d := os.Getenv("CLAUDE_PROJECT_DIR"); d != "" {
		return d
	}
	if p.Cwd != "" {
		return p.Cwd
	}
	if p.ProjectDir != "" {
		return p.ProjectDir
	}
	d, _ := os.Getwd()
	return d
}

func eventOr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// ── Adapters por arnés ───────────────────────────────────────────────────────

// runHookClaude emite el shape nativo de Claude Code.
func runHookClaude(event string, p hookPayload) int {
	pd := p.projectDir()
	switch event {
	case "PreToolUse":
		decision, reason := decidePreToolUseTool(pd, p.ToolInput.FilePath, p.ToolInput.Command)
		if decision == "deny" {
			logEvent(pd, denyEvent(pd, p.ToolInput.FilePath, p.ToolInput.Command, p.SessionID, reason))
			emitJSON(map[string]any{"hookSpecificOutput": map[string]any{
				"hookEventName":            "PreToolUse",
				"permissionDecision":       "deny",
				"permissionDecisionReason": reason,
			}})
		}
	case "SessionStart":
		if ctx := sessionContext(pd); ctx != "" {
			emitJSON(map[string]any{"hookSpecificOutput": map[string]any{
				"hookEventName":     "SessionStart",
				"additionalContext": ctx,
			}})
		}
	case "UserPromptSubmit":
		if ctx := userPromptContext(pd, p.SessionID); ctx != "" {
			emitJSON(map[string]any{"hookSpecificOutput": map[string]any{
				"hookEventName":     "UserPromptSubmit",
				"additionalContext": ctx,
			}})
		}
	case "Stop":
		// stop_hook_active = ya estamos en una continuación forzada por un Stop
		// hook → no re-bloquear (defensa extra contra loops; el nudge-once ya lo evita).
		if p.StopHookActive {
			return 0
		}
		if target := stopNudge(pd); target != "" {
			logEvent(pd, sfEvent{Kind: "nudge", Session: p.SessionID, Feature: target, Detail: "journal nudge on archive"})
			emitJSON(map[string]any{"decision": "block", "reason": journalNudgeReason(target)})
		}
	case "SessionEnd":
		markSession(pd, "session ended")
	case "PreCompact":
		markSession(pd, "pre-compaction checkpoint")
		forceFullSlice(pd, p.SessionID) // re-grounding post-compact
	}
	return 0
}

// runHookGeneric emite el contrato normalizado (lo consumen pi/cursor/opencode).
func runHookGeneric(p hookPayload) int {
	pd := p.projectDir()
	switch p.Event {
	case "pre_tool_use":
		decision, reason := decidePreToolUseTool(pd, p.FilePath, p.Command)
		if decision == "deny" {
			logEvent(pd, denyEvent(pd, p.FilePath, p.Command, p.SessionID, reason))
		}
		emitJSON(map[string]any{"decision": decision, "reason": nilIfEmpty(reason)})
	case "session_start":
		emitJSON(map[string]any{"decision": "allow", "context": sessionContext(pd)})
	case "user_prompt_submit":
		emitJSON(map[string]any{"decision": "allow", "context": userPromptContext(pd, p.SessionID)})
	case "stop":
		if target := stopNudge(pd); target != "" {
			logEvent(pd, sfEvent{Kind: "nudge", Session: p.SessionID, Feature: target, Detail: "journal nudge on archive"})
			emitJSON(map[string]any{"decision": "block", "reason": journalNudgeReason(target)})
		} else {
			emitJSON(map[string]any{"decision": "allow"})
		}
	case "session_end":
		markSession(pd, "session ended")
		emitJSON(map[string]any{"decision": "allow"})
	case "pre_compact":
		markSession(pd, "pre-compaction checkpoint")
		forceFullSlice(pd, p.SessionID)
		emitJSON(map[string]any{"decision": "allow"})
	default:
		emitJSON(map[string]any{"decision": "allow", "reason": fmt.Sprintf("unknown event %q", p.Event)})
	}
	return 0
}

// nilIfEmpty devuelve nil para "" (marshala a JSON null, igual que el .py que
// emitía reason=None en allow) o el string si tiene contenido.
func nilIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

func emitJSON(v any) {
	out, err := json.Marshal(v)
	if err != nil {
		return
	}
	fmt.Println(string(out))
}

// isPreToolUseEvent: ¿este evento es el hard-gate de escritura, en cualquiera de
// los dos contratos? (Claude usa "PreToolUse"; el generic, "pre_tool_use".)
func isPreToolUseEvent(harness, event string) bool {
	if harness == "generic" {
		return event == "pre_tool_use"
	}
	return event == "PreToolUse"
}

// emitDeny imprime un deny en la forma nativa del arnés. Lo usa el fail-closed.
func emitDeny(harness, reason string) {
	if harness == "generic" {
		emitJSON(map[string]any{"decision": "deny", "reason": reason})
		return
	}
	emitJSON(map[string]any{"hookSpecificOutput": map[string]any{
		"hookEventName":            "PreToolUse",
		"permissionDecision":       "deny",
		"permissionDecisionReason": reason,
	}})
}

// hasActiveFeatureSafe responde si hay alguna feature en curso (activeStatuses),
// blindado con su propio recover: si LEER el estado también explota, devolvemos
// false (no podemos afirmar que haya algo activo → no forzamos el fail-closed).
func hasActiveFeatureSafe(projectDir string) (active bool) {
	defer func() {
		if recover() != nil {
			active = false
		}
	}()
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		return false
	}
	for i := range ff.Features {
		if activeStatuses[ff.Features[i].Status] {
			return true
		}
	}
	return false
}

// ── Decisión: PreToolUse (el hard-gate) ──────────────────────────────────────

// decidePreToolUse devuelve ("deny", reason) o ("allow", "") para un Write/Edit
// sobre filePath. Reusa los structs y helpers del CLI — no re-deriva nada.
func decidePreToolUse(projectDir, filePath string) (string, string) {
	if filePath == "" {
		return "allow", ""
	}
	if !isDir(filepath.Join(projectDir, "specforge")) {
		return "allow", "" // no es un proyecto SpecForge → nunca interferir
	}
	rel, inside := relPosix(projectDir, filePath)
	if !inside {
		return "allow", "" // fuera del árbol del proyecto
	}

	// Capa 1 (FIXBUGHIGH): el estado autoritativo SOLO lo escribe `sf`. Cualquier
	// Write/Edit directo a features.json o a un *.json de artefacto se deniega acá,
	// ANTES de la lógica de orden-de-gates (que ya solo aplica al .md renderizado).
	if reason, deny := protectedStateDeny(rel); deny {
		return "deny", reason
	}

	feature, artefact, ok := artifactTarget(rel)
	if !ok {
		return "allow", ""
	}

	ff, _ := readFeaturesFile(projectDir) // ausente/ inválido → featuresFile vacío

	// Serial (F22): no arrancar una 2da feature mientras otra está en curso. El
	// PRIMER artefacto de una feature es requirements → lo interceptamos ahí.
	// F2: si la constitución declara flow.mode=parallel, este guard no aplica
	// (el estado por feature de A1 hace estructuralmente seguro el paralelo).
	if artefact == "requirements" {
		if other := activeOther(ff, feature); other != "" && !parallelFlow(projectDir) {
			return "deny", fmt.Sprintf(
				"Serial flow: feature '%s' is still active. Finish and archive it before "+
					"starting '%s' — one active feature at a time (F22). "+
					"(Teams can opt into parallel features via flow.mode=parallel in the constitution.)",
				other, feature)
		}
		return "allow", ""
	}

	// Cadena de gates: el downstream necesita su upstream INMEDIATO aprobado.
	if required := immediateUpstream(artefact); required != "" && !gateApprovedIn(ff, feature, required) {
		return "deny", fmt.Sprintf(
			"Cannot write %s for '%s': the %s gate is not approved in features.json. "+
				"Present %s at a 🔴 gate and get approval first — gates cannot be skipped.",
			artefact, feature, required, required)
	}
	return "allow", ""
}

// artifactTarget extrae (feature, artefacto) si rel es un artefacto gateable.
func artifactTarget(rel string) (feature, artefact string, ok bool) {
	m := artifactRe.FindStringSubmatch(rel)
	if m == nil {
		return "", "", false
	}
	return m[1], m[2], true
}

// ── Capa 1: protección del estado autoritativo ───────────────────────────────

// decidePreToolUseTool es el dispatcher de PreToolUse: enruta según QUÉ tool se
// está por usar. La tool Bash trae `command`; Write/Edit traen `file_path`. Si
// hay comando, lo inspeccionamos como escritura por shell; si no, va por la ruta
// clásica de file_path. (Mantener decidePreToolUse con su firma de 2 args deja
// intactos los tests existentes.)
func decidePreToolUseTool(projectDir, filePath, command string) (string, string) {
	if strings.TrimSpace(command) != "" {
		return decideBashWrite(projectDir, command)
	}
	return decidePreToolUse(projectDir, filePath)
}

// protectedStateDeny decide si `rel` (ruta POSIX relativa al proyecto) es estado
// autoritativo que NO se puede escribir a mano, y devuelve un mensaje que ENSEÑA
// el camino correcto (el comando `sf` equivalente). Es la única definición del
// "conjunto protegido": la comparten la rama Write/Edit y la rama Bash.
func protectedStateDeny(rel string) (string, bool) {
	if strings.HasPrefix(rel, "specforge/.state/") {
		return "specforge/.state/ is machine state — never edit it directly. " +
			"features.json is the source of truth; let the session protocol manage state.", true
	}
	if !protectedJSONRe.MatchString(rel) {
		return "", false
	}
	// Deltas (D4'): nombre variable (D1.json, D2.json…) → se resuelve por ruta,
	// no por base. Solo `sf delta` los escribe (el borrador va en drafts/).
	if strings.Contains(rel, "/deltas/") {
		return "deltas are SpecForge state — don't write them directly. " +
			"Draft the JSON in the feature's drafts/ dir and promote it with " +
			"`sf delta new --feature=… --from=drafts/<file>.json`; lifecycle via `sf delta set-status`.", true
	}
	switch base := filepath.Base(rel); base {
	case "features.json":
		return "features.json is the SpecForge gate ledger — never edit it by hand. " +
			"Gates come only from `sf gate approve`; feature status changes via the lifecycle commands.", true
	case "feature.json":
		return "feature.json is this feature's gate ledger + lifecycle state — never edit it by hand. " +
			"Gates come only from `sf gate approve`; status/lane only via `sf feature set-status|set-lane`.", true
	case "constitution.json":
		return "constitution.json is SpecForge state — don't write it directly. " +
			"Draft it at specforge/drafts/constitution.json (writable), then promote with " +
			"`sf save constitution --from=drafts/constitution.json` (it validates, then writes).", true
	case "domain.json":
		return "domain.json is SpecForge state — don't write it directly. " +
			"Draft it at specforge/drafts/domain.json (writable), then promote with " +
			"`sf save domain --from=drafts/domain.json` (it validates, then writes).", true
	case "sources.json":
		return "sources.json is SpecForge state — don't write it directly. " +
			"Draft it at specforge/drafts/sources.json (writable), then promote with " +
			"`sf save sources --from=drafts/sources.json` (it validates, then writes).", true
	case "trace.json":
		return "trace.json is SpecForge state — don't write it directly. " +
			"Draft it under the feature's drafts/ dir and promote with " +
			"`sf save trace --feature=… --from=drafts/trace.json`, then check it with `sf trace verify`.", true
	case "audit.json":
		return "audit.json is the SpecForge quality ledger — don't write it directly. " +
			"Phase verdicts are recorded only via `sf gate record-verdict --feature=… --json -`.", true
	default:
		// requirements/design/tasks/plan/review.json, o una entrada de journal.
		if strings.HasPrefix(rel, "specforge/journal/") {
			return "journal entries are SpecForge state — don't write them directly. " +
				"Use `sf journal add --feature=… --json -`.", true
		}
		name := strings.TrimSuffix(base, ".json")
		return fmt.Sprintf("%s.json is SpecForge state — don't write it directly. "+
			"Write your draft to specforge/features/<feature>/drafts/%s.json (that dir IS writable), "+
			"then promote it with `sf save %s --feature=… --from=drafts/%s.json` (it validates against "+
			"the schema, then writes both the .json and the .md).", name, name, name, name), true
	}
}

// ── Capa 1: detección de escritura de estado vía Bash ─────────────────────────

// decideBashWrite inspecciona un comando de shell y deniega si alguna de sus
// partes escribe a un path protegido SIN ser una invocación de `sf`. Es una
// heurística que SUBE EL COSTO del atajo (no un sandbox): cubre los vectores
// comunes (redirects, tee, sed -i, cp, mv); un `python -c "open(...,'w')"` se le
// escapa. La garantía dura es que Write/Edit están bloqueados; Bash es best-effort.
func decideBashWrite(projectDir, command string) (string, string) {
	if !isDir(filepath.Join(projectDir, "specforge")) {
		return "allow", "" // no es un proyecto SpecForge → nunca interferir
	}
	for _, seg := range splitCommandSegments(command) {
		if reason, deny := segmentWritesProtected(projectDir, seg); deny {
			return "deny", reason
		}
	}
	return "allow", ""
}

// cmdSepRe parte un comando compuesto en segmentos por los separadores de shell
// (`;`, `&&`, `||`, `|`, y newline). Cada segmento se evalúa por separado: así un
// `sf save … && echo x > features.json` se pesca en su segunda mitad.
var cmdSepRe = regexp.MustCompile(`\|\||&&|[;|\n]`)

func splitCommandSegments(command string) []string {
	return cmdSepRe.Split(command, -1)
}

// redirectRe captura el destino de una redirección de salida: `>`, `>>`, `>|`.
var redirectRe = regexp.MustCompile(`>>?\|?\s*([^\s|&;<>]+)`)

// writeVerbs: comandos cuyo propósito es crear/sobrescribir un archivo (cuando el
// path destino aparece como argumento). sed entra solo con -i (edición in-place).
var writeVerbs = map[string]bool{
	"tee": true, "cp": true, "mv": true, "dd": true,
	"truncate": true, "install": true, "ln": true,
}

// segmentWritesProtected evalúa UN segmento de comando. Deniega si:
//   - redirige (`>`/`>>`) hacia un path protegido (vale incluso para `sf`, porque
//     redirigir DENTRO del estado nunca es legítimo), o
//   - es un write-verb (tee/cp/mv/sed -i/…) con un path protegido como argumento.
//
// Un `sf save …` sin redirect escribe internamente → no matchea nada → se permite.
func segmentWritesProtected(projectDir, seg string) (string, bool) {
	// 1) Redirecciones a un path protegido (independiente del programa).
	for _, m := range redirectRe.FindAllStringSubmatch(seg, -1) {
		if reason, deny := classifyProtectedToken(projectDir, m[1]); deny {
			return reason, true
		}
	}

	fields := strings.Fields(seg)
	if len(fields) == 0 {
		return "", false
	}
	prog := filepath.Base(fields[0]) // tolera rutas tipo /usr/bin/cp

	// 2) Write-verbs con un argumento protegido. `sed` solo si trae -i.
	isWriteVerb := writeVerbs[prog] || (prog == "sed" && hasInPlaceFlag(fields))
	if !isWriteVerb {
		return "", false
	}
	for _, tok := range fields[1:] {
		if reason, deny := classifyProtectedToken(projectDir, tok); deny {
			return reason, true
		}
	}
	return "", false
}

// hasInPlaceFlag detecta el -i de sed (edición in-place), en sus formas `-i`,
// `-i.bak`, o combinada `-ri`/`-ni`.
func hasInPlaceFlag(fields []string) bool {
	for _, f := range fields[1:] {
		if f == "-i" || strings.HasPrefix(f, "-i.") {
			return true
		}
		if strings.HasPrefix(f, "-") && !strings.HasPrefix(f, "--") && strings.Contains(f, "i") {
			return true
		}
	}
	return false
}

// classifyProtectedToken resuelve un token de comando (posible path, con o sin
// comillas) contra el proyecto y delega en protectedStateDeny. Así Bash y
// Write/Edit comparten exactamente el mismo conjunto protegido y los mismos
// mensajes didácticos.
func classifyProtectedToken(projectDir, raw string) (string, bool) {
	raw = strings.Trim(raw, `"'`)
	if raw == "" {
		return "", false
	}
	rel, inside := relPosix(projectDir, raw)
	if !inside {
		return "", false
	}
	return protectedStateDeny(rel)
}

// immediateUpstream devuelve la fase ANTERIOR a `phase` en la cadena (su upstream
// inmediato), o "" si es la primera. Deriva de artifactChain (stale.go) → una
// sola definición del orden del pipeline, sin tabla duplicada.
func immediateUpstream(phase string) string {
	for i, link := range artifactChain {
		if link.phase == phase {
			if i == 0 {
				return ""
			}
			return artifactChain[i-1].phase
		}
	}
	return ""
}

// gateApprovedIn: ¿la fase tiene un gate aprobado en la feature? Reusa
// latestApproveGate (stale.go).
func gateApprovedIn(ff featuresFile, feature, phase string) bool {
	for i := range ff.Features {
		if ff.Features[i].Name == feature {
			return latestApproveGate(&ff.Features[i], phase) != nil
		}
	}
	return false
}

// activeOther devuelve el nombre de OTRA feature en curso (activeStatuses, state.go),
// distinta de `feature`, o "" si no hay. Es el guard del flujo serial.
func activeOther(ff featuresFile, feature string) string {
	for i := range ff.Features {
		f := &ff.Features[i]
		if f.Name != "" && f.Name != feature && activeStatuses[f.Status] {
			return f.Name
		}
	}
	return ""
}

// ── Inyección de contexto (SessionStart / UserPromptSubmit / PreCompact) ──────

// injectBudgetBytes: presupuesto POR CHUNK de lo que se inyecta cada sesión
// (C6 de EVALUACION-PLATAFORMA: learnings.md/session.md crecen sin techo en un
// proyecto largo; inyectarlos enteros come contexto sin límite). ~8KB ≈ 2K
// tokens por chunk. El archivo completo sigue en disco — el marcador de
// truncado le dice al agente dónde leer el resto si lo necesita.
const injectBudgetBytes = 8192

// sessionContext arma lo que se re-inyecta en SessionStart (incl. post-compact):
// session.md + compact-rules.md + learnings.md, si existen — cada uno recortado
// a su presupuesto. session.md conserva la COLA (lo reciente es lo que orienta);
// los curados (compact-rules, learnings) conservan la CABEZA (su orden es
// importancia).
func sessionContext(projectDir string) string {
	sf, ok := specforgeRoot(projectDir)
	if !ok {
		return ""
	}
	var chunks []string
	add := func(rel, header string, keepTail bool) {
		data, err := os.ReadFile(filepath.Join(sf, rel))
		if err != nil {
			return
		}
		body := budgetTrim(string(data), injectBudgetBytes, keepTail, "specforge/"+filepath.ToSlash(rel))
		chunks = append(chunks, header+"\n\n"+body)
	}
	add(filepath.Join(".state", "session.md"), "## SpecForge session — resume from here", true)
	add(filepath.Join("context", "compact-rules.md"), "## SpecForge compact-rules — project invariants", false)
	add("learnings.md", "## SpecForge learnings — consolidated, evidence-anchored", false)
	return strings.Join(chunks, "\n\n")
}

// budgetTrim recorta s al presupuesto (alineado a línea) y anota QUÉ se cortó y
// dónde leer el resto. keepTail=true conserva el final; false, el principio.
func budgetTrim(s string, budget int, keepTail bool, path string) string {
	if len(s) <= budget {
		return s
	}
	note := fmt.Sprintf("[truncated to %dB — full file: %s]", budget, path)
	if keepTail {
		cut := s[len(s)-budget:]
		if i := strings.IndexByte(cut, '\n'); i >= 0 && i < len(cut)-1 {
			cut = cut[i+1:] // arrancar en línea completa
		}
		return "…" + note + "\n" + cut
	}
	cut := s[:budget]
	if i := strings.LastIndexByte(cut, '\n'); i > 0 {
		cut = cut[:i] // terminar en línea completa
	}
	return cut + "\n…" + note
}

// userPromptContext decide y arma lo que se inyecta en UserPromptSubmit. "" = nada.
// A diferencia del .py, NO shell-ea a `sf context current`: computa el slice en
// proceso (somos sf) vía computeCurrentState + buildCurrentContext.
func userPromptContext(projectDir, sessionID string) string {
	sf, ok := specforgeRoot(projectDir)
	if !ok {
		return ""
	}
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		return "" // sin features.json no hay slice
	}
	cc := buildCurrentContext(computeCurrentState(ff, projectDir), projectDir)
	if cc.Feature == "" {
		return cc.Breadcrumb // sin feature activa: solo la línea
	}

	key := []string{cc.Feature, cc.Phase, waveKey(cc.Wave)}
	st := readHookState(sf)
	entry := st.Sessions[sessionID]

	// force_full lo setea PreCompact: tras compactar, los slices previos se
	// resumieron → forzamos un slice completo este turno.
	if entry.ForceFull {
		st.setSession(sessionID, sessionEntry{Key: key, TurnsSinceFull: 0})
		writeHookState(sf, st)
		return renderFullSlice(cc)
	}

	mode, turnsNext := decideInjection(key, entry.Key, entry.TurnsSinceFull, fullSliceEvery())
	st.setSession(sessionID, sessionEntry{Key: key, TurnsSinceFull: turnsNext})
	writeHookState(sf, st)

	if mode == "full" {
		return renderFullSlice(cc)
	}
	return cc.Breadcrumb
}

// decideInjection es PURA (testeable): dado el estado actual vs el último
// inyectado y el contador de turnos sin cambio, decide qué inyectar. "full" en
// on-step-change o al llegar al backstop de N turnos (y resetea); si no,
// "breadcrumb" y suma 1.
func decideInjection(key, lastKey []string, turns, fullEvery int) (string, int) {
	if !equalStrs(key, lastKey) || turns >= fullEvery {
		return "full", 0
	}
	return "breadcrumb", turns + 1
}

// renderFullSlice envuelve el slice como contexto inyectable, rotulado como
// autoritativo (el más reciente reemplaza a los anteriores del transcript).
func renderFullSlice(cc currentContext) string {
	body, _ := json.MarshalIndent(cc, "", "  ")
	return "## SpecForge — current step (re-grounding; supersedes earlier slices)\n\n" +
		cc.Breadcrumb + "\n\n```json\n" + string(body) + "\n```"
}

// forceFullSlice marca la sesión para que el próximo UserPromptSubmit inyecte el
// slice COMPLETO. Lo llama PreCompact.
func forceFullSlice(projectDir, sessionID string) {
	sf, ok := specforgeRoot(projectDir)
	if !ok || sessionID == "" {
		return
	}
	st := readHookState(sf)
	entry := st.Sessions[sessionID]
	entry.ForceFull = true
	st.setSession(sessionID, entry)
	writeHookState(sf, st)
}

func fullSliceEvery() int {
	if v, err := strconv.Atoi(os.Getenv("SPECFORGE_FULL_SLICE_EVERY")); err == nil && v > 0 {
		return v
	}
	return fullSliceEveryDefault
}

// waveKey formatea la wave (puntero) para la clave de estado: "" si es nil.
func waveKey(w *int) string {
	if w == nil {
		return ""
	}
	return strconv.Itoa(*w)
}

// ── Conciliador de memoria (Stop nudge) ──────────────────────────────────────

// stopNudge decide a qué feature nudgear en Stop (una sola vez) y lo registra.
func stopNudge(projectDir string) string {
	sf, ok := specforgeRoot(projectDir)
	if !ok {
		return ""
	}
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		return ""
	}
	st := readHookState(sf)
	nudged := toSet(st.JournalNudged)
	target := pickJournalNudge(ff.Features, journaledFeatures(sf), nudged)
	if target == "" {
		return ""
	}
	nudged[target] = true // nudge-once: lo marcamos antes de devolver
	st.JournalNudged = sortedKeys(nudged)
	writeHookState(sf, st)
	return target
}

// pickJournalNudge es PURA (testeable): primera feature archivada que no esté
// journaleada ni ya nudgeada, o "".
func pickJournalNudge(features []feature, journaled, nudged map[string]bool) string {
	for _, f := range features {
		if f.Name != "" && archivedStatuses[f.Status] && !journaled[f.Name] && !nudged[f.Name] {
			return f.Name
		}
	}
	return ""
}

// journaledFeatures: features que YA tienen entrada en specforge/journal/ (leemos
// el campo `feature` del .json, robusto ante nombres con guiones).
func journaledFeatures(sf string) map[string]bool {
	out := map[string]bool{}
	matches, _ := filepath.Glob(filepath.Join(sf, "journal", "*.json"))
	for _, p := range matches {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var j struct {
			Feature string `json:"feature"`
		}
		if json.Unmarshal(data, &j) == nil && j.Feature != "" {
			out[j.Feature] = true
		}
	}
	return out
}

func journalNudgeReason(feature string) string {
	return fmt.Sprintf(
		"Feature `%s` quedó archivada sin lecciones registradas. Extraé las lecciones "+
			"durables y guardalas con:\n  sf journal add --feature=%s --json -\n"+
			"(Recordatorio único — no volverá a aparecer.)", feature, feature)
}

// ── Estado del hook por sesión (specforge/.state/hook-context.json) ───────────

type sessionEntry struct {
	Key            []string `json:"key,omitempty"`
	TurnsSinceFull int      `json:"turns_since_full"`
	ForceFull      bool     `json:"force_full,omitempty"`
	// LastSeen (RFC3339) habilita la poda por TTL: sin esto, hook-context.json
	// acumulaba una entrada POR SESIÓN para siempre (C6).
	LastSeen string `json:"last_seen,omitempty"`
}

type hookState struct {
	JournalNudged []string                `json:"journal_nudged,omitempty"`
	Sessions      map[string]sessionEntry `json:"sessions,omitempty"`
}

// sessionTTL: cuánto vive una entrada de sesión sin actividad. Dos semanas
// cubre de sobra cualquier "retomar la sesión del viernes".
const sessionTTL = 14 * 24 * time.Hour

func (s *hookState) setSession(id string, e sessionEntry) {
	if s.Sessions == nil {
		s.Sessions = map[string]sessionEntry{}
	}
	e.LastSeen = nowUTC() // toda escritura refresca el TTL de SU sesión
	s.Sessions[id] = e
}

// pruneStaleSessions borra las sesiones vencidas (o legacy, sin last_seen: si
// no sabemos cuándo se usaron, ya no orientan a nadie). Se llama al escribir el
// estado, así el archivo se auto-limpia con el uso normal — sin cron ni comando.
func (s *hookState) pruneStaleSessions(now time.Time) {
	for id, e := range s.Sessions {
		t, err := time.Parse(time.RFC3339, e.LastSeen)
		if err != nil || now.Sub(t) > sessionTTL {
			delete(s.Sessions, id)
		}
	}
}

func readHookState(sf string) hookState {
	var st hookState
	data, err := os.ReadFile(filepath.Join(sf, ".state", "hook-context.json"))
	if err == nil {
		_ = json.Unmarshal(data, &st) // inválido → estado vacío (fail open)
	}
	if st.Sessions == nil {
		st.Sessions = map[string]sessionEntry{}
	}
	return st
}

func writeHookState(sf string, st hookState) {
	st.pruneStaleSessions(time.Now().UTC()) // C6: el archivo se poda al escribirse
	dir := filepath.Join(sf, ".state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return // fail open: peor caso, re-inyectamos de más
	}
	out, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(filepath.Join(dir, "hook-context.json"), out, 0o644)
}

// sessionMDMaxBytes: techo de session.md (C6). El archivo apendea marcadores
// por evento de ciclo de vida; sin cap crece para siempre. 32KB de marcadores
// recientes orientan igual que 3MB de historia.
const sessionMDMaxBytes = 32 * 1024

// markSession apendea un marcador de continuidad liviano a session.md. El resumen
// rico es trabajo del agente (un command hook no tiene acceso a la conversación);
// esto solo timestampea eventos de ciclo de vida. Si el archivo supera el techo,
// se recorta conservando la MITAD más reciente (los marcadores viejos no
// orientan a nadie; git tiene la historia si importara).
func markSession(projectDir, note string) {
	sf, ok := specforgeRoot(projectDir)
	if !ok {
		return
	}
	dir := filepath.Join(sf, ".state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	path := filepath.Join(dir, "session.md")
	ts := time.Now().UTC().Format(time.RFC3339)
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	fmt.Fprintf(f, "\n<!-- %s @ %s -->\n", note, ts)
	f.Close()
	capSessionMD(path)
}

// capSessionMD recorta session.md a la mitad del techo cuando lo supera
// (histéresis: recortar a maxBytes exacto haría un rewrite por append).
func capSessionMD(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() <= sessionMDMaxBytes {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	tail := data[len(data)-sessionMDMaxBytes/2:]
	if i := bytes.IndexByte(tail, '\n'); i >= 0 {
		tail = tail[i+1:] // arrancar en línea completa
	}
	out := append([]byte("<!-- session.md capped — older markers dropped -->\n"), tail...)
	_ = os.WriteFile(path, out, 0o644)
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// specforgeRoot devuelve la ruta de <project>/specforge si es un directorio.
func specforgeRoot(projectDir string) (string, bool) {
	sf := filepath.Join(projectDir, "specforge")
	if isDir(sf) {
		return sf, true
	}
	return "", false
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// relPosix devuelve la ruta de filePath relativa a projectDir, con separadores
// "/", y si está DENTRO del árbol del proyecto.
func relPosix(projectDir, filePath string) (string, bool) {
	projAbs, err := filepath.Abs(projectDir)
	if err != nil {
		return "", false
	}
	target := filePath
	if !filepath.IsAbs(target) {
		target = filepath.Join(projAbs, target)
	}
	rel, err := filepath.Rel(projAbs, target)
	if err != nil {
		return "", false
	}
	rel = filepath.ToSlash(rel)
	if rel == ".." || strings.HasPrefix(rel, "../") {
		return "", false // fuera del árbol
	}
	return rel, true
}

func equalStrs(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func toSet(xs []string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

func sortedKeys(m map[string]bool) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}
