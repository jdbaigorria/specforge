package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// nowUTC devuelve el instante actual en UTC con formato RFC3339 (el mismo de los
// gates en features.json, ej. "2026-06-20T14:00:00Z").
func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// runGate es el punto de entrada de `sf gate ...`. Sub-acciones: `status` (lee
// el ledger de gates humanos de features.json) y `record-verdict` (persiste el
// veredicto del auditor de fase, tier calidad).
func runGate(args []string) int {
	if len(args) == 0 {
		gateUsage()
		return 2
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "status":
		return runGateStatusCmd(rest)
	case "approve":
		return runGateApprove(rest)
	case "record-verdict":
		return runGateRecordVerdict(rest)
	default:
		fmt.Fprintf(os.Stderr, "sf gate: unknown sub-command %q\n\n", sub)
		gateUsage()
		return 2
	}
}

func gateUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf gate status [--feature=NAME] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf gate approve --feature=NAME --phase=PHASE [--by=user] [--comment=TEXT] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf gate record-verdict --feature=NAME [--phase=PHASE] [--json -|FILE] [project_dir]")
}

// ----------------------------------------------------------------------------
// `sf gate approve` — registra DETERMINÍSTICAMENTE un gate humano aprobado y, lo
// clave, SELLA el hash del artefacto. Es el primer comando del CLI que ESCRIBE
// features.json (hasta ahora el ledger lo escribía el skill a mano). Sellar el
// hash acá —y no pedírselo al LLM— es lo que hace confiable la detección de
// silent edits: el hash lo computa la máquina sobre el archivo real.
// ----------------------------------------------------------------------------

func runGateApprove(args []string) int {
	projectDir := "."
	feature, phase, by, comment := "", "", "user", ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--phase="):
			phase = strings.TrimPrefix(a, "--phase=")
		case strings.HasPrefix(a, "--by="):
			by = strings.TrimPrefix(a, "--by=")
		case strings.HasPrefix(a, "--comment="):
			comment = strings.TrimPrefix(a, "--comment=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf gate approve: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" || phase == "" {
		gateUsage()
		return 2
	}
	return gateApprove(projectDir, feature, phase, by, comment)
}

// gateApprove hace el trabajo: lee features.json, sella el hash del artefacto de
// la fase, apendea el gate y reescribe el archivo. El parámetro se llama `name`
// (no `feature`) para no tapar al tipo `feature`.
func gateApprove(projectDir, name, phase, by, comment string) int {
	path := filepath.Join(projectDir, "specforge", "features.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf gate approve: cannot read features.json under %s (%v)\n", projectDir, err)
		return 4
	}
	var ff featuresFile
	if err := json.Unmarshal(data, &ff); err != nil {
		fmt.Fprintf(os.Stderr, "sf gate approve: invalid features.json (%v)\n", err)
		return 2
	}

	// Buscamos la feature por nombre. Tomamos el puntero al elemento real del
	// slice (no una copia) para poder mutar sus Gates.
	var f *feature
	for i := range ff.Features {
		if ff.Features[i].Name == name {
			f = &ff.Features[i]
			break
		}
	}
	if f == nil {
		fmt.Fprintf(os.Stderr, "sf gate approve: feature %q not found\n", name)
		return 4
	}

	// R1 (integrity.go): no EXTENDEMOS un ledger roto. Si la cadena no valida,
	// alguien escribió gates fuera del CLI; aprobar encima legitimaría el fraude.
	if refuseOnBrokenLedger("sf gate approve", f) {
		return 5
	}

	// Capa 2: el verdict es el sello final (siguiente paso = archive). No se
	// otorga si la trazabilidad driftó, si algún requirement no nombra un test
	// real, o si no hay un resultado de test verde y fresco. El LLM no puede
	// saltearse esto editando a mano: features.json está protegido (Capa 1).
	if phase == "verdict" {
		if reasons := verdictPreconditions(projectDir, name); len(reasons) > 0 {
			// Telemetría (A7): un verdict rehusado es exactamente el dato que
			// queremos poder contar después.
			logEvent(projectDir, sfEvent{Kind: "refuse", Feature: name, Phase: "verdict",
				Detail: fmt.Sprintf("verdict refused: %d precondition(s) unmet", len(reasons))})
			fmt.Fprintf(os.Stderr, "sf gate approve: verdict refused for %q — %d precondition(s) unmet:\n", name, len(reasons))
			for _, r := range reasons {
				fmt.Fprintf(os.Stderr, "  - %s\n", r)
			}
			return 5
		}
	}

	// Sellamos el hash del artefacto de esta fase (si la fase tiene artefacto).
	// No se puede aprobar un artefacto que falta: sería certificar el vacío.
	hash := ""
	if rel, hasArtifact := artifactFileForPhase(phase); hasArtifact {
		abs := filepath.Join(projectDir, "specforge", "features", name, rel)
		h, ok := hashArtifact(abs)
		if !ok {
			fmt.Fprintf(os.Stderr, "sf gate approve: artifact for phase %q not found (%s)\n", phase, rel)
			return 4
		}
		hash = h
	}

	// Prev tiene que computarse ANTES del append (es el hash de la última
	// entrada EXISTENTE; después del append "la última" sería esta misma).
	f.Gates = append(f.Gates, gate{
		Phase:   phase,
		Result:  "approve",
		By:      by,
		At:      nowUTC(),
		Comment: comment,
		Hash:    hash,
		Prev:    nextPrev(f),
	})

	if code := writeFeaturesFile(path, ff); code != 0 {
		return code
	}
	if hash != "" {
		fmt.Printf("approved %s/%s (hash %s…)\n", name, phase, hash[:12])
	} else {
		fmt.Printf("approved %s/%s\n", name, phase)
	}
	return 0
}

// writeFeaturesFile reescribe features.json con indent de 2 espacios + newline
// final (mismo estilo que el resto de los .json del repo). Centralizado acá para
// que cualquier futuro comando que mute el registro escriba igual.
func writeFeaturesFile(path string, ff featuresFile) int {
	out, err := json.MarshalIndent(ff, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf gate approve: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf gate approve: write failed (%v)\n", err)
		return 1
	}
	return 0
}

// runGateStatusCmd parsea los flags de `status`.
func runGateStatusCmd(args []string) int {
	projectDir := "."
	feature := ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf gate: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	return runGateStatus(projectDir, feature)
}

// runGateStatus lee el gate ledger de features.json. Sin --feature muestra un
// resumen por feature; con --feature muestra el ledger detallado de esa feature.
func runGateStatus(projectDir, feature string) int {
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "features.json"))
	if err != nil {
		fmt.Printf("No specforge/features.json under %s.\n", projectDir)
		return 0
	}
	var ff featuresFile
	if err := json.Unmarshal(data, &ff); err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: invalid features.json: %v\n", err)
		return 1
	}

	if feature != "" {
		return gateLedger(ff, feature)
	}
	gateSummary(ff)
	return 0
}

// gateSummary: una fila por feature con el conteo de gates y el último estado.
func gateSummary(ff featuresFile) {
	fmt.Print("SpecForge — gate status\n\n")
	if len(ff.Features) == 0 {
		fmt.Println("No features registered yet.")
		return
	}

	cols := []string{"feature", "status", "gates", "phase", "last-result", "when"}
	var rows [][]string
	for i := range ff.Features {
		f := &ff.Features[i]
		lastResult, when := "—", "—"
		if n := len(f.Gates); n > 0 {
			lastResult = orDash(f.Gates[n-1].Result)
			when = orDash(f.Gates[n-1].At)
		}
		rows = append(rows, []string{
			f.Name,
			orDash(f.Status),
			fmt.Sprintf("%d", len(f.Gates)),
			lastPhase(f),
			lastResult,
			when,
		})
	}
	renderTable(cols, rows)
}

// gateLedger: el historial completo de gates de una feature. Devuelve exit 4
// (not found) si la feature no existe.
func gateLedger(ff featuresFile, name string) int {
	var f *feature
	for i := range ff.Features {
		if ff.Features[i].Name == name {
			f = &ff.Features[i]
			break
		}
	}
	if f == nil {
		fmt.Fprintf(os.Stderr, "sf gate: feature %q not found.\n", name)
		return 4
	}

	fmt.Printf("Gate ledger — %s  (status: %s, current phase: %s)\n\n",
		f.Name, orDash(f.Status), lastPhase(f))
	if len(f.Gates) == 0 {
		fmt.Println("  no gates recorded yet.")
		return 0
	}

	cols := []string{"phase", "result", "by", "at", "comment"}
	var rows [][]string
	for _, g := range f.Gates {
		rows = append(rows, []string{
			g.Phase,
			g.Result,
			orDash(g.By),
			orDash(g.At),
			orDash(g.Comment),
		})
	}
	renderTable(cols, rows)
	return 0
}

// ----------------------------------------------------------------------------
// `sf gate record-verdict` — persiste el veredicto del AUDITOR DE FASE.
//
// El juez (subagente fresco) produce {phase, verdicts:[{rule,result,citation}]}.
// El CLI computa el `overall` (fail si alguna regla falla), le pone timestamp y
// lo APENDEA a specforge/features/<f>/audit.json — el ledger de calidad,
// SEPARADO de los gates humanos de features.json y de la matriz trace.json.
//
// El CLI solo PERSISTE (determinista). La decisión nudge/block la toma el hook
// según el `overall`: exit 0 = pass, exit 3 = recorded-but-FAIL → señal limpia
// para que el hook nudgee sin parsear stdout.
// ----------------------------------------------------------------------------

type auditLedger struct {
	Feature string       `json:"feature"`
	Entries []auditEntry `json:"entries"`
}

type auditEntry struct {
	Phase    string        `json:"phase"`
	At       string        `json:"at"`
	Overall  string        `json:"overall"` // pass | fail
	Verdicts []ruleVerdict `json:"verdicts"`
}

type ruleVerdict struct {
	Rule     string `json:"rule"`
	Result   string `json:"result"` // pass | fail
	Citation string `json:"citation"`
	// CitationCheck (A8): el CLI verifica la citation contra el artefacto real.
	// "verified" = aparece textual (whitespace-normalizado); "not-found" = el juez
	// citó algo que NO está en el artefacto (citation inventada — la mentira
	// semántica que sigue después de cerrar la escritura de estado); "" = no hay
	// citation o la fase no tiene artefacto contra qué verificar.
	CitationCheck string `json:"citation_check,omitempty"`
}

var verdictResults = map[string]bool{"pass": true, "fail": true}

func runGateRecordVerdict(args []string) int {
	projectDir := "."
	feature := ""
	phase := ""
	jsonSrc := "-"
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--phase="):
			phase = strings.TrimPrefix(a, "--phase=")
		case strings.HasPrefix(a, "--json="):
			jsonSrc = strings.TrimPrefix(a, "--json=")
		case a == "--json":
			if i+1 < len(args) {
				jsonSrc = args[i+1]
				i++
			}
		case a != "-" && strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf gate: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf gate: --feature=NAME is required")
		return 2
	}

	raw, err := readJSONInput(jsonSrc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: cannot read input (%v)\n", err)
		return 1
	}
	// Solo nos interesan phase + verdicts del JSON del juez.
	var in struct {
		Phase    string        `json:"phase"`
		Verdicts []ruleVerdict `json:"verdicts"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: invalid JSON (%v)\n", err)
		return 2
	}
	if phase != "" {
		in.Phase = phase // el flag manda
	}

	entry, rep := buildAuditEntry(in.Phase, in.Verdicts)
	if len(rep.errors) > 0 {
		for _, e := range rep.errors {
			fmt.Printf("  ERROR:   %s\n", e)
		}
		fmt.Printf("\nFAIL: not recorded — %d error(s).\n", len(rep.errors))
		return 2
	}

	// A8: cada citation se verifica mecánicamente contra el artefacto antes de
	// entrar al ledger. No rechazamos (el juez es nudge, no gate duro), pero lo
	// no-verificado queda MARCADO — el humano del gate lo ve.
	verifyCitations(projectDir, feature, &entry)
	for _, v := range entry.Verdicts {
		if v.CitationCheck == "not-found" {
			fmt.Printf("  warning: rule %q cites text not found in the %s artifact — recorded as citation_check=not-found\n",
				v.Rule, entry.Phase)
		}
	}

	if code := appendAuditEntry(projectDir, feature, entry); code != 0 {
		return code
	}

	// Telemetría (A7): cada veredicto del juez queda contable — los "fail" por
	// feature son la medida de presión REVISE.
	logEvent(projectDir, sfEvent{Kind: "verdict", Feature: feature, Phase: entry.Phase, Detail: entry.Overall})

	fmt.Printf("recorded %s verdict for %s/%s (%d rule(s))\n", entry.Overall, feature, entry.Phase, len(entry.Verdicts))
	if entry.Overall == "fail" {
		return 3 // recorded, pero el veredicto es FAIL → el hook nudgea
	}
	return 0
}

// ----------------------------------------------------------------------------
// Capa 2 de FIXBUGHIGH — el sello final (gate `verdict`) no se puede falsificar.
//
// `gate approve --phase=verdict` es el último gate antes de archivar (done). Sin
// esta guarda, el agente podía aprobarlo igual que cualquier otro y declarar
// "todos los tests pasan / 70% coverage" sin haber corrido nada. Ahora el CLI
// REHÚSA el verdict salvo que tres condiciones se cumplan, todas verificadas por
// la máquina contra el disco real (no contra la narración del LLM).
// ----------------------------------------------------------------------------

// verdictPreconditions devuelve la lista de razones por las que el verdict NO
// puede otorgarse (vacía = todo en orden):
//  1. trace.json existe y NINGUNA requirement driftó (todos sus code anchors viven).
//  2. cada requirement nombra ≥1 test y cada test ref resuelve a un test real
//     (el contrato de verificación: "test pass != done" se vuelve chequeable).
//  3. hay un resultado de test VERDE y FRESCO: `sf check run` pasó y su code_hash
//     == el code_hash actual (no se tocó el código después de correr).
func verdictPreconditions(projectDir, feature string) []string {
	var reasons []string

	specforge := filepath.Join(projectDir, "specforge")
	tracePath := findArtifact(specforge, feature, "trace.json")
	if tracePath == "" {
		reasons = append(reasons, "no trace.json — produce it with `sf save trace --feature="+feature+" --json -`")
	} else if data, err := os.ReadFile(tracePath); err != nil {
		reasons = append(reasons, "trace.json is unreadable")
	} else {
		var tf traceFile
		if json.Unmarshal(data, &tf) != nil {
			reasons = append(reasons, "trace.json is invalid JSON")
		} else if len(tf.Requirements) == 0 {
			reasons = append(reasons, "trace.json declares no requirements")
		} else {
			// Orden estable para que el reporte sea reproducible.
			ids := make([]string, 0, len(tf.Requirements))
			for id := range tf.Requirements {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			for _, req := range ids {
				info := tf.Requirements[req]
				for _, anchor := range info.Code {
					if ok, why := checkAnchor(projectDir, anchor); !ok {
						reasons = append(reasons, fmt.Sprintf("%s: code drift (%s)", req, why))
					}
				}
				if len(info.Test) == 0 {
					reasons = append(reasons, fmt.Sprintf("%s: names no test (verification contract unmet)", req))
				}
				for _, tref := range info.Test {
					if ok, why := checkAnchor(projectDir, tref); !ok {
						reasons = append(reasons, fmt.Sprintf("%s: test %q does not resolve (%s)", req, tref, why))
					}
				}
			}
		}
	}

	// Resultado de test verde y FRESCO (Capa 3 lo produjo; acá lo exigimos).
	res, ok := readCheckResult(projectDir, feature)
	switch {
	case !ok:
		reasons = append(reasons, "no test result on record — run `sf check run --feature="+feature+"`")
	case !res.Passed:
		reasons = append(reasons, fmt.Sprintf("last `sf check run` FAILED (exit %d) — fix the code and re-run", res.ExitCode))
	case res.CodeHash != codeHash(projectDir):
		reasons = append(reasons, "test result is STALE — code changed since the last `sf check run`; re-run it")
	}
	return reasons
}

// buildAuditEntry valida los verdicts y computa el overall (función pura).
func buildAuditEntry(phase string, verdicts []ruleVerdict) (auditEntry, report) {
	var rep report
	if strings.TrimSpace(phase) == "" {
		rep.errorf("phase is required")
	}
	if len(verdicts) == 0 {
		rep.errorf("at least one rule verdict is required")
	}
	overall := "pass"
	for i, v := range verdicts {
		if strings.TrimSpace(v.Rule) == "" {
			rep.errorf("verdict %d: empty rule", i+1)
		}
		if !verdictResults[v.Result] {
			rep.errorf("verdict %d (%s): result must be pass|fail, got %q", i+1, v.Rule, v.Result)
		}
		if v.Result == "fail" {
			overall = "fail"
		}
	}
	return auditEntry{
		Phase:    phase,
		At:       nowUTC(),
		Overall:  overall,
		Verdicts: verdicts,
	}, rep
}

// verifyCitations (A8) aplica la regla que ordena el backlog de integridad:
// "ninguna afirmación del modelo entra al estado sin verificación mecánica o
// sin marcarse como no-verificada". Una citation del juez ES una afirmación
// ("esto está en el artefacto"); acá la chequeamos con un substring-match
// whitespace-normalizado contra el .json canónico Y el .md renderizado de la
// fase (el juez suele citar del render). Barato y determinista.
func verifyCitations(projectDir, feature string, entry *auditEntry) {
	rel, ok := artifactFileForPhase(entry.Phase)
	if !ok {
		return // fase sin artefacto (lane, wave-N) → nada contra qué verificar
	}
	base := filepath.Join(projectDir, "specforge", "features", feature)
	var haystacks []string
	for _, p := range []string{rel, strings.TrimSuffix(rel, ".json") + ".md"} {
		if data, err := os.ReadFile(filepath.Join(base, p)); err == nil {
			haystacks = append(haystacks, normalizeWS(string(data)))
		}
	}
	if len(haystacks) == 0 {
		return // artefacto ausente: lo reporta el flujo de gates, no esta marca
	}
	for i := range entry.Verdicts {
		v := &entry.Verdicts[i]
		if strings.TrimSpace(v.Citation) == "" {
			continue
		}
		needle := normalizeWS(v.Citation)
		v.CitationCheck = "not-found"
		for _, h := range haystacks {
			if strings.Contains(h, needle) {
				v.CitationCheck = "verified"
				break
			}
		}
	}
}

// normalizeWS colapsa todo whitespace (saltos de línea, tabs, espacios
// repetidos) a un espacio simple: una citation que envuelve línea en el .md
// sigue matcheando.
func normalizeWS(s string) string { return strings.Join(strings.Fields(s), " ") }

// appendAuditEntry lee el ledger existente (si hay), apendea la entrada y lo
// reescribe canónico.
func appendAuditEntry(projectDir, feature string, entry auditEntry) int {
	path := filepath.Join(projectDir, "specforge", "features", feature, "audit.json")

	ledger := auditLedger{Feature: feature}
	if data, err := os.ReadFile(path); err == nil {
		// Si existe pero no parsea, preferimos fallar antes que pisar el historial.
		if err := json.Unmarshal(data, &ledger); err != nil {
			fmt.Fprintf(os.Stderr, "sf gate: existing audit.json is invalid (%v) — not overwriting\n", err)
			return 1
		}
	}
	ledger.Feature = feature
	ledger.Entries = append(ledger.Entries, entry)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: %v\n", err)
		return 1
	}
	out, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: cannot write %s (%v)\n", path, err)
		return 1
	}
	return 0
}
