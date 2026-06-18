package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// runGate es el punto de entrada de `sf gate ...`. Hoy la única sub-acción es
// `status`. Reutiliza el modelo featuresFile/feature/gate definido en status.go.
func runGate(args []string) int {
	if len(args) == 0 || args[0] != "status" {
		fmt.Fprintln(os.Stderr, "usage: sf gate status [project_dir] [--feature=NAME]")
		return 2
	}
	args = args[1:] // descartamos "status"

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
