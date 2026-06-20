package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
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

	f.Gates = append(f.Gates, gate{
		Phase:   phase,
		Result:  "approve",
		By:      by,
		At:      nowUTC(),
		Comment: comment,
		Hash:    hash,
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

	if code := appendAuditEntry(projectDir, feature, entry); code != 0 {
		return code
	}

	fmt.Printf("recorded %s verdict for %s/%s (%d rule(s))\n", entry.Overall, feature, entry.Phase, len(entry.Verdicts))
	if entry.Overall == "fail" {
		return 3 // recorded, pero el veredicto es FAIL → el hook nudgea
	}
	return 0
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
