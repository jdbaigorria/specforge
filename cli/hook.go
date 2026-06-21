package main

import (
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

// archivedStatuses: estados que cuentan como "archivada" para el nudge del journal.
var archivedStatuses = map[string]bool{"done": true, "archived": true}

// artifactRe reconoce la ruta-relativa de un artefacto gateable. JSON-first:
// matchea .json (la fuente) Y .md (el render); plan vive bajo progress/.
var artifactRe = regexp.MustCompile(
	`^specforge/features/([^/]+)/(?:progress/)?(requirements|design|tasks|plan)\.(?:json|md)$`,
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
	SessionID      string `json:"session_id"`
	StopHookActive bool   `json:"stop_hook_active"`
	HookEventName  string `json:"hook_event_name"`
	ToolInput      struct {
		FilePath string `json:"file_path"`
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

	// Fail open: si algo explota, avisamos por stderr y devolvemos 0 (allow).
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintf(os.Stderr, "sf hook: internal error, allowing (%v)\n", r)
			code = 0
		}
	}()

	raw, _ := io.ReadAll(os.Stdin)
	var p hookPayload
	if len(strings.TrimSpace(string(raw))) > 0 {
		_ = json.Unmarshal(raw, &p) // payload inválido → struct vacío (fail open)
	}

	if harness == "generic" {
		return runHookGeneric(p)
	}
	return runHookClaude(eventOr(event, p.HookEventName), p)
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
		decision, reason := decidePreToolUse(pd, p.ToolInput.FilePath)
		if decision == "deny" {
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
		decision, reason := decidePreToolUse(pd, p.FilePath)
		emitJSON(map[string]any{"decision": decision, "reason": nilIfEmpty(reason)})
	case "session_start":
		emitJSON(map[string]any{"decision": "allow", "context": sessionContext(pd)})
	case "user_prompt_submit":
		emitJSON(map[string]any{"decision": "allow", "context": userPromptContext(pd, p.SessionID)})
	case "stop":
		if target := stopNudge(pd); target != "" {
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

	if strings.HasPrefix(rel, "specforge/.state/") {
		return "deny", "specforge/.state/ is machine state — never edit it directly. " +
			"features.json is the source of truth; let the session protocol manage state."
	}

	feature, artefact, ok := artifactTarget(rel)
	if !ok {
		return "allow", ""
	}

	ff, _ := readFeaturesFile(projectDir) // ausente/ inválido → featuresFile vacío

	// Serial (F22): no arrancar una 2da feature mientras otra está en curso. El
	// PRIMER artefacto de una feature es requirements → lo interceptamos ahí.
	if artefact == "requirements" {
		if other := activeOther(ff, feature); other != "" {
			return "deny", fmt.Sprintf(
				"Serial flow: feature '%s' is still active. Finish and archive it before "+
					"starting '%s' — one active feature at a time (F22).", other, feature)
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

// sessionContext arma lo que se re-inyecta en SessionStart (incl. post-compact):
// session.md + compact-rules.md + learnings.md, si existen.
func sessionContext(projectDir string) string {
	sf, ok := specforgeRoot(projectDir)
	if !ok {
		return ""
	}
	var chunks []string
	add := func(rel, header string) {
		if data, err := os.ReadFile(filepath.Join(sf, rel)); err == nil {
			chunks = append(chunks, header+"\n\n"+string(data))
		}
	}
	add(filepath.Join(".state", "session.md"), "## SpecForge session — resume from here")
	add(filepath.Join("context", "compact-rules.md"), "## SpecForge compact-rules — project invariants")
	add("learnings.md", "## SpecForge learnings — consolidated, evidence-anchored")
	return strings.Join(chunks, "\n\n")
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
}

type hookState struct {
	JournalNudged []string                `json:"journal_nudged,omitempty"`
	Sessions      map[string]sessionEntry `json:"sessions,omitempty"`
}

func (s *hookState) setSession(id string, e sessionEntry) {
	if s.Sessions == nil {
		s.Sessions = map[string]sessionEntry{}
	}
	s.Sessions[id] = e
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

// markSession apendea un marcador de continuidad liviano a session.md. El resumen
// rico es trabajo del agente (un command hook no tiene acceso a la conversación);
// esto solo timestampea eventos de ciclo de vida.
func markSession(projectDir, note string) {
	sf, ok := specforgeRoot(projectDir)
	if !ok {
		return
	}
	dir := filepath.Join(sf, ".state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	ts := time.Now().UTC().Format(time.RFC3339)
	f, err := os.OpenFile(filepath.Join(dir, "session.md"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "\n<!-- %s @ %s -->\n", note, ts)
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

// readFeaturesFile lee y parsea specforge/features.json.
func readFeaturesFile(projectDir string) (featuresFile, error) {
	var ff featuresFile
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "features.json"))
	if err != nil {
		return ff, err
	}
	if err := json.Unmarshal(data, &ff); err != nil {
		return ff, err
	}
	return ff, nil
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
