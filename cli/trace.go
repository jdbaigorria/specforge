package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// runTrace es el punto de entrada de `sf trace ...`. Hoy la única sub-acción es
// `verify`. Reutiliza el modelo traceFile/traceReq y checkAnchor de doctor.go, y
// findArtifact de status.go.
func runTrace(args []string) int {
	if len(args) == 0 || args[0] != "verify" {
		fmt.Fprintln(os.Stderr, "usage: sf trace verify [project_dir] [--feature=NAME]")
		return 2
	}
	args = args[1:] // descartamos "verify"

	projectDir := "."
	feature := ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf trace: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	return runTraceVerify(projectDir, feature)
}

// runTraceVerify valida la matriz de trazabilidad contra el disco. Con --feature
// valida esa feature (exit 4 si no tiene trace.json); sin --feature valida todas
// las que tengan trace.json. Exit 1 si alguna requirement divergió.
func runTraceVerify(projectDir, feature string) int {
	specforge := filepath.Join(projectDir, "specforge")

	if feature != "" {
		tracePath := findArtifact(specforge, feature, "trace.json")
		if tracePath == "" {
			fmt.Fprintf(os.Stderr, "sf trace: no trace.json for feature %q.\n", feature)
			return 4
		}
		return verifyTrace(projectDir, feature, tracePath)
	}

	traces := loadTraces(specforge)
	if len(traces) == 0 {
		fmt.Printf("No trace.json found under %s/specforge.\n", projectDir)
		return 0
	}
	exit := 0
	for i, t := range traces {
		if i > 0 {
			fmt.Println() // separación entre matrices
		}
		if verifyTrace(projectDir, t.feature, t.path) != 0 {
			exit = 1
		}
	}
	return exit
}

// verifyTrace imprime la matriz de una feature y devuelve 1 si alguna
// requirement divergió (algún anchor de código ya no existe), 0 si todo en sync.
// Columnas: la "recorded" es el status que grabó sf-check; la "live" es lo que
// verificamos ahora contra el código real.
func verifyTrace(projectDir, feature, tracePath string) int {
	data, err := os.ReadFile(tracePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf trace: cannot read %s (%v)\n", tracePath, err)
		return 1
	}
	var tf traceFile
	if err := json.Unmarshal(data, &tf); err != nil {
		fmt.Fprintf(os.Stderr, "sf trace: invalid trace.json for %s (%v)\n", feature, err)
		return 1
	}

	// Orden estable de requirements (los maps de Go iteran aleatorio).
	reqIDs := make([]string, 0, len(tf.Requirements))
	for id := range tf.Requirements {
		reqIDs = append(reqIDs, id)
	}
	sort.Strings(reqIDs)

	fmt.Printf("Trace matrix — %s\n\n", feature)

	cols := []string{"requirement", "code", "tests", "recorded", "live"}
	var rows [][]string
	drifted := 0
	for _, req := range reqIDs {
		info := tf.Requirements[req]
		live := "ok"
		for _, anchor := range info.Code {
			if ok, _ := checkAnchor(projectDir, anchor); !ok {
				live = "DRIFT"
			}
		}
		if live == "DRIFT" {
			drifted++
		}
		rows = append(rows, []string{
			req,
			joinOrDash(info.Code),
			joinOrDash(info.Test),
			orDash(info.Status),
			live,
		})
	}
	renderTable(cols, rows)

	if drifted > 0 {
		fmt.Printf("\nDRIFT: %d requirement(s) diverged.\n", drifted)
		return 1
	}
	fmt.Println("\nOK: all requirements traced to live code.")
	return 0
}
