package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf recover` — el comando de recuperación (recomendaciones-de-ia.md §6).
//
// Cuando algo quedó inconsistente (un artefacto editado después de aprobarlo, un
// upstream reabierto, una fundación que se movió), `sf recover` DIAGNOSTICA y
// propone un PLAN ORDENADO de recuperación. Cuelga del stale model: no recalcula
// nada, lee computeArtifactStates y lo traduce a acciones.
//
// Decisión de diseño: recover es ADVISORY. NO muta. No puede:
//   - re-aprobar por su cuenta → sería forjar un gate humano.
//   - regenerar contenido      → eso lo produce el LLM, no el CLI.
// Entonces hace lo único determinístico y honesto: detectar el problema y emitir
// los pasos concretos (con el comando exacto) para que humano+LLM los ejecuten.
// Misma filosofía del resto: el CLI computa, el humano aprueba, el LLM produce.
// ----------------------------------------------------------------------------

// recoveryStep es un paso del plan. Action clasifica el tipo de arreglo; Command
// (cuando aplica) es el comando exacto a correr; Hint explica el contexto.
type recoveryStep struct {
	Step    int    `json:"step"`
	Phase   string `json:"phase"`
	Action  string `json:"action"` // reseal | regenerate | generate | resolve-upstream
	Command string `json:"command,omitempty"`
	Hint    string `json:"hint,omitempty"`
}

// recoveryReport es lo que emite recover para UNA feature.
type recoveryReport struct {
	Feature  string          `json:"feature"`
	Detected []artifactState `json:"detected"`        // los artefactos problemáticos
	Plan     []recoveryStep  `json:"plan"`            // pasos en orden upstream→downstream
	OK       bool            `json:"ok"`              // true si no hay nada que recuperar
	Note     string          `json:"note,omitempty"`
}

// runRecover parsea flags y emite el/los reporte(s). Sin --feature: todas las
// features activas (no done), igual que `sf status --artifacts`.
func runRecover(args []string) int {
	projectDir := "."
	feature := ""
	asJSON := false
	for _, a := range args {
		switch {
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf recover: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "features.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf recover: cannot read features.json under %s (%v)\n", projectDir, err)
		return 4
	}
	var ff featuresFile
	if err := json.Unmarshal(data, &ff); err != nil {
		fmt.Fprintf(os.Stderr, "sf recover: invalid features.json (%v)\n", err)
		return 2
	}

	// Recolectamos los reportes de las features que correspondan.
	var reports []recoveryReport
	for i := range ff.Features {
		f := &ff.Features[i]
		if feature != "" && f.Name != feature {
			continue
		}
		if feature == "" && f.Status == "done" {
			continue
		}
		reports = append(reports, buildRecoveryReport(f, projectDir))
	}

	if len(reports) == 0 {
		if feature != "" {
			fmt.Printf("No feature %q.\n", feature)
		} else {
			fmt.Println("No active features (all done/archived).")
		}
		return 0
	}

	if asJSON {
		out, _ := json.MarshalIndent(reports, "", "  ")
		fmt.Println(string(out))
	} else {
		for _, r := range reports {
			printRecovery(r)
		}
	}

	// Exit 5 (mismo código que `sf status --artifacts`) si ALGUNA feature tiene
	// inconsistencias: así un hook/CI puede reaccionar uniformemente.
	for _, r := range reports {
		if !r.OK {
			return 5
		}
	}
	return 0
}

// buildRecoveryReport es la función PURA-de-cómputo: dado el estado de los
// artefactos, arma el diagnóstico + el plan. Camina la cadena en orden, así los
// pasos quedan upstream→downstream (arreglar primero la raíz).
func buildRecoveryReport(f *feature, projectDir string) recoveryReport {
	states := computeArtifactStates(f, projectDir)
	rep := recoveryReport{Feature: f.Name}

	step := 0
	for i, s := range states {
		if !isRecoverable(states, i) {
			continue
		}
		rep.Detected = append(rep.Detected, s)
		step++
		rep.Plan = append(rep.Plan, recoveryStepFor(step, f.Name, s))
	}

	if len(rep.Plan) == 0 {
		rep.OK = true
		rep.Note = "all artifacts approved and in sync"
	}
	return rep
}

// isRecoverable decide si el artefacto en la posición i es un PROBLEMA que
// recover deba listar. stale y blocked siempre lo son. missing solo si es una
// raíz (hay contenido aguas abajo): un missing "de cola" (p.ej. plan todavía no
// hecho) NO es inconsistencia, es trabajo pendiente — eso lo cubre `sf next`.
func isRecoverable(states []artifactState, i int) bool {
	switch states[i].State {
	case "stale", "blocked":
		return true
	case "missing":
		return downstreamHasContent(states, i)
	default: // approved, draft
		return false
	}
}

// downstreamHasContent: ¿algún artefacto POSTERIOR a i ya tiene contenido (no
// está missing)? Si sí, el gap en i es una raíz a arreglar, no una cola.
func downstreamHasContent(states []artifactState, i int) bool {
	for j := i + 1; j < len(states); j++ {
		if states[j].State != "missing" {
			return true
		}
	}
	return false
}

// recoveryStepFor traduce un artefacto problemático a un paso accionable.
func recoveryStepFor(step int, feature string, s artifactState) recoveryStep {
	approveCmd := fmt.Sprintf("sf gate approve --feature=%s --phase=%s", feature, s.Phase)
	rs := recoveryStep{Step: step, Phase: s.Phase}

	switch {
	case s.State == "blocked":
		rs.Action = "resolve-upstream"
		rs.Hint = "fix the upstream artifact first (" + s.Reason + "), then approve this one"

	case s.State == "missing":
		rs.Action = "generate"
		rs.Command = approveCmd
		rs.Hint = "regenerate this artifact (re-run sf-propose), then seal it with the command"

	case strings.Contains(s.Reason, "hash mismatch"):
		// El artefacto se editó después de aprobarlo: la acción es ACEPTAR la
		// edición re-sellando el hash (o revertir el archivo si fue sin querer).
		rs.Action = "reseal"
		rs.Command = approveCmd
		rs.Hint = "content changed after approval — re-approve to accept it (or revert the file if the edit was unintended)"

	default:
		// stale por upstream reabierto / propagación: la fundación cambió, así que
		// hay que REVISAR/regenerar contra el upstream nuevo antes de re-aprobar.
		rs.Action = "regenerate"
		rs.Command = approveCmd
		rs.Hint = "foundation moved (" + s.Reason + ") — review/regenerate against the updated upstream, then re-approve"
	}
	return rs
}

// printRecovery imprime el reporte legible de una feature.
func printRecovery(r recoveryReport) {
	fmt.Printf("SpecForge — recover: %s\n\n", r.Feature)
	if r.OK {
		fmt.Printf("✓ No inconsistencies — %s.\n\n", r.Note)
		return
	}

	fmt.Println("Detected:")
	for _, s := range r.Detected {
		fmt.Printf("  - %s: %s — %s\n", s.Phase, s.State, orDash(s.Reason))
	}

	fmt.Println("\nRecommended (in order):")
	for _, p := range r.Plan {
		fmt.Printf("  %d. %s [%s] — %s\n", p.Step, p.Phase, p.Action, p.Hint)
		if p.Command != "" {
			fmt.Printf("     $ %s\n", p.Command)
		}
	}
	fmt.Println()
}
