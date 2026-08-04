package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf gate show` — EVIDENCIA para el gate humano (A5).
//
// El riesgo de producto #1 es la fatiga de gates: 7+ gates por feature, cada
// uno presentando el artefacto ENTERO re-renderizado → el humano deja de leer →
// sello de goma → el sistema degrada a teatro. La respuesta no es menos gates,
// es MENOS LECTURA por gate: mostrar qué cambió y qué evidencia lo respalda,
// no el documento completo.
//
// Todo lo que se muestra acá ya existe y es determinista:
//   - el ledger (qué gates pasaron, cuándo, con qué sello)
//   - el estado de artefactos (stale model — ¿la fundación se movió?)
//   - el trace (qué requirements están anclados a código y tests reales)
//   - el último check run (verde/rojo, fresco/stale, resultados por test)
//   - los veredictos del juez para la fase (con citation_check de A8)
//   - los commits desde el último gate (git, best-effort)
//
// `sf gate show` solo lo junta en una vista para decidir en 20 segundos.
// ----------------------------------------------------------------------------

// gateEvidence es el paquete de evidencia computado (separado del render para
// poder testearlo como datos).
type gateEvidence struct {
	Feature   string          `json:"feature"`
	Status    string          `json:"status"`
	Lane      string          `json:"lane,omitempty"`
	LastGate  *gate           `json:"last_gate,omitempty"`
	Artifacts []artifactState `json:"artifacts"`
	// Trace: conteos de la matriz de trazabilidad contra el repo REAL.
	TraceTotal   int      `json:"trace_total"`
	TraceOK      int      `json:"trace_ok"`
	TraceIssues  []string `json:"trace_issues,omitempty"`
	// Check: el último resultado sellado y su frescura.
	Check      *checkResult `json:"check,omitempty"`
	CheckFresh bool         `json:"check_fresh"`
	// Veredictos del juez para la fase pedida (audit.json).
	Verdicts []auditEntry `json:"verdicts,omitempty"`
	// Commits desde el último gate (best-effort, requiere git).
	Commits []string `json:"commits,omitempty"`
}

func runGateShow(args []string) int {
	projectDir := "."
	featureName := ""
	phase := ""
	asJSON := false
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			featureName = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--phase="):
			phase = strings.TrimPrefix(a, "--phase=")
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf gate show: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if featureName == "" {
		fmt.Fprintln(os.Stderr, "sf gate show: --feature=NAME is required")
		return 2
	}

	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf gate show: cannot read feature state under %s (%v)\n", projectDir, err)
		return 4
	}
	f := findFeature(&ff, featureName)
	if f == nil {
		fmt.Fprintf(os.Stderr, "sf gate show: feature %q not found\n", featureName)
		return 4
	}

	ev := buildGateEvidence(projectDir, f, phase)
	if asJSON {
		out, _ := json.MarshalIndent(ev, "", "  ")
		fmt.Println(string(out))
		return 0
	}
	renderGateEvidence(ev, phase)
	return 0
}

// buildGateEvidence junta la evidencia determinista de una feature (y una fase,
// si se pide — filtra los veredictos del juez a esa fase).
func buildGateEvidence(projectDir string, f *feature, phase string) gateEvidence {
	ev := gateEvidence{
		Feature: f.Name,
		Status:  f.Status,
		Lane:    f.Lane,
	}

	// Último gate aprobado (el punto de referencia de "qué cambió desde").
	var last *gate
	for i := range f.Gates {
		g := &f.Gates[i]
		if approveResult(g.Result) && (last == nil || g.At > last.At) {
			last = g
		}
	}
	ev.LastGate = last

	// Estado de artefactos (stale model): ¿la fundación se movió desde su gate?
	ev.Artifacts = computeArtifactStates(f, projectDir)

	// Trace contra el repo real: por requirement, código y tests deben resolver.
	if tf, ok := readTraceFile(projectDir, f.Name); ok {
		ids := make([]string, 0, len(tf.Requirements))
		for id := range tf.Requirements {
			ids = append(ids, id)
		}
		sort.Strings(ids)
		ev.TraceTotal = len(ids)
		for _, id := range ids {
			info := tf.Requirements[id]
			issue := ""
			for _, a := range info.Code {
				if ok, why := checkAnchor(projectDir, a); !ok {
					issue = fmt.Sprintf("%s: code drift (%s)", id, why)
				}
			}
			if len(info.Test) == 0 {
				issue = id + ": names no test"
			}
			for _, tr := range info.Test {
				if ok, why := checkAnchor(projectDir, tr); !ok {
					issue = fmt.Sprintf("%s: test gone (%s)", id, why)
				}
			}
			if issue == "" {
				ev.TraceOK++
			} else {
				ev.TraceIssues = append(ev.TraceIssues, issue)
			}
		}
	}

	// Último check run + frescura (mismo criterio que el verdict).
	if res, ok := readCheckResult(projectDir, f.Name); ok {
		ev.Check = &res
		ev.CheckFresh = res.CodeHash == codeHash(projectDir)
	}

	// Veredictos del juez (audit.json), filtrados a la fase si se pidió una.
	if entries, ok := readAuditLedger(projectDir, f.Name); ok {
		for _, e := range entries {
			if phase == "" || e.Phase == phase {
				ev.Verdicts = append(ev.Verdicts, e)
			}
		}
	}

	// Commits desde el último gate (git log --since; best-effort).
	if last != nil {
		ev.Commits = gitLogSince(projectDir, last.At)
	}
	return ev
}

// readTraceFile lee el trace.json de una feature (ok=false si no hay).
func readTraceFile(projectDir, name string) (traceFile, bool) {
	var tf traceFile
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "features", name, "trace.json"))
	if err != nil || json.Unmarshal(data, &tf) != nil {
		return tf, false
	}
	return tf, true
}

// readAuditLedger lee las entradas de audit.json (ok=false si no hay).
func readAuditLedger(projectDir, name string) ([]auditEntry, bool) {
	var ledger auditLedger
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "features", name, "audit.json"))
	if err != nil || json.Unmarshal(data, &ledger) != nil {
		return nil, false
	}
	return ledger.Entries, true
}

// gitLogSince devuelve los commits (oneline) posteriores a un timestamp
// RFC3339. Sin git o sin repo → nil (la evidencia degrada, no falla).
func gitLogSince(projectDir, since string) []string {
	cmd := exec.Command("git", "-C", projectDir, "log", "--oneline", "--since="+since)
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	var lines []string
	for l := range strings.SplitSeq(strings.TrimSpace(string(out)), "\n") {
		if l != "" {
			lines = append(lines, l)
		}
	}
	return lines
}

// renderGateEvidence imprime la vista compacta para el humano del gate.
func renderGateEvidence(ev gateEvidence, phase string) {
	title := ev.Feature
	if phase != "" {
		title += " / " + phase
	}
	fmt.Printf("Gate evidence — %s\n\n", title)
	fmt.Printf("  status: %s%s\n", ev.Status, laneSuffix(ev.Lane))
	if ev.LastGate != nil {
		fmt.Printf("  last gate: %s (%s, by %s)\n", ev.LastGate.Phase, ev.LastGate.At, ev.LastGate.By)
	} else {
		fmt.Println("  last gate: none yet")
	}

	// Artefactos: solo los que NO están en orden (lo sano no necesita lectura).
	var offenders []artifactState
	for _, a := range ev.Artifacts {
		if a.State == "stale" || a.State == "blocked" {
			offenders = append(offenders, a)
		}
	}
	if len(offenders) == 0 {
		fmt.Println("  artifacts: all consistent (no stale foundations)")
	} else {
		fmt.Println("  artifacts NEEDING ATTENTION:")
		for _, a := range offenders {
			fmt.Printf("    - %s is %s (%s)\n", a.Phase, a.State, a.Reason)
		}
	}

	// Trace.
	switch {
	case ev.TraceTotal == 0:
		fmt.Println("  trace: none yet")
	case len(ev.TraceIssues) == 0:
		fmt.Printf("  trace: %d/%d requirements anchored to live code + tests\n", ev.TraceOK, ev.TraceTotal)
	default:
		fmt.Printf("  trace: %d/%d ok — issues:\n", ev.TraceOK, ev.TraceTotal)
		for _, i := range ev.TraceIssues {
			fmt.Printf("    - %s\n", i)
		}
	}

	// Check.
	if ev.Check == nil {
		fmt.Println("  tests: no `sf check run` on record")
	} else {
		verdict := "FAIL"
		if ev.Check.Passed {
			verdict = "PASS"
		}
		fresh := "STALE (code changed since — re-run)"
		if ev.CheckFresh {
			fresh = "fresh"
		}
		extra := ""
		if n := len(ev.Check.Tests); n > 0 {
			extra = fmt.Sprintf(", %d per-test results", n)
		}
		fmt.Printf("  tests: %s at %s (%s%s)\n", verdict, ev.Check.At, fresh, extra)
	}

	// Juez.
	for _, e := range ev.Verdicts {
		fmt.Printf("  judge %s@%s: %s\n", e.Phase, e.At, e.Overall)
		for _, v := range e.Verdicts {
			if v.Result == "fail" || v.CitationCheck == "not-found" {
				note := ""
				if v.CitationCheck == "not-found" {
					note = " [citation NOT found in artifact]"
				}
				fmt.Printf("    - %s: %s%s\n", v.Rule, v.Result, note)
			}
		}
	}

	// Commits.
	if len(ev.Commits) > 0 {
		fmt.Printf("  commits since last gate (%d):\n", len(ev.Commits))
		for i, c := range ev.Commits {
			if i == 10 {
				fmt.Printf("    … and %d more\n", len(ev.Commits)-10)
				break
			}
			fmt.Printf("    %s\n", c)
		}
	}
}
