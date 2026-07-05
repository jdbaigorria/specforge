package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf coverage` — spec coverage con RATCHET (A4, §2.9).
//
// La métrica: qué fracción de los archivos de código está ANCLADA a algún
// trace.json (activo o archivado). Es "qué % de tu código está gobernado por
// una spec viva" — el code coverage de la era agéntica. Sin esto no hay noción
// de progreso en una adopción brownfield.
//
// El ratchet: como los ratchets de types/coverage — el % solo puede subir.
// specforge/.state/coverage.json guarda la marca; si una corrida da MENOS que
// la marca, exit 5 (algo des-ancló código: trace borrado, archivo renombrado
// sin resync). Si da más, la marca sube sola. Adopción incremental MEDIBLE,
// sin big-bang de especificar todo el repo (que nadie va a hacer).
// ----------------------------------------------------------------------------

type coverageBaseline struct {
	SchemaVersion string  `json:"schema_version"`
	Percent       float64 `json:"percent"`
	Anchored      int     `json:"anchored"`
	Total         int     `json:"total"`
	At            string  `json:"at"`
}

func coverageBaselinePath(projectDir string) string {
	return filepath.Join(projectDir, "specforge", ".state", "coverage.json")
}

func runCoverage(args []string) int {
	projectDir := "."
	asJSON := false
	for _, a := range args {
		switch {
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf coverage: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if !isDir(filepath.Join(projectDir, "specforge")) {
		fmt.Fprintln(os.Stderr, "sf coverage: no specforge/ directory — nothing to measure")
		return 4
	}

	anchored, total := computeSpecCoverage(projectDir)
	percent := 0.0
	if total > 0 {
		percent = 100 * float64(anchored) / float64(total)
	}

	baseline, hadBaseline := readCoverageBaseline(projectDir)

	if asJSON {
		out, _ := json.MarshalIndent(map[string]any{
			"percent": percent, "anchored": anchored, "total": total,
			"baseline": baseline.Percent, "had_baseline": hadBaseline,
		}, "", "  ")
		fmt.Println(string(out))
	} else {
		fmt.Printf("spec coverage: %.1f%% (%d/%d code files anchored to a trace)\n", percent, anchored, total)
	}

	// El ratchet. Tolerancia mínima (0.01) para no fallar por redondeo.
	switch {
	case !hadBaseline:
		writeCoverageBaseline(projectDir, percent, anchored, total)
		if !asJSON {
			fmt.Printf("baseline set at %.1f%% — from here it only goes up.\n", percent)
		}
	case percent < baseline.Percent-0.01:
		fmt.Fprintf(os.Stderr, "RATCHET VIOLATED: coverage fell %.1f%% → %.1f%% (baseline %s).\n",
			baseline.Percent, percent, baseline.At)
		fmt.Fprintln(os.Stderr, "Something un-anchored code: a deleted trace, a renamed file without resync. "+
			"Fix the trace (or re-anchor) instead of lowering the bar.")
		return 5
	case percent > baseline.Percent+0.01:
		writeCoverageBaseline(projectDir, percent, anchored, total)
		if !asJSON {
			fmt.Printf("ratchet raised: %.1f%% → %.1f%%\n", baseline.Percent, percent)
		}
	}
	return 0
}

// computeSpecCoverage cuenta (anclados, totales): archivos de código del repo
// vs archivos citados por algún anchor de algún trace.json (features activas y
// archivadas — el legado anclado cuenta, de eso se trata la adopción).
func computeSpecCoverage(projectDir string) (anchored, total int) {
	files, ok := gitListFiles(projectDir)
	if !ok {
		files = walkListFiles(projectDir)
	}

	// El conjunto de archivos anclados por los traces.
	anchoredSet := map[string]bool{}
	for _, pattern := range []string{
		filepath.Join(projectDir, "specforge", "features", "*", "trace.json"),
		filepath.Join(projectDir, "specforge", "archive", "*", "trace.json"),
	} {
		paths, _ := filepath.Glob(pattern)
		for _, p := range paths {
			var tf traceFile
			data, err := os.ReadFile(p)
			if err != nil || json.Unmarshal(data, &tf) != nil {
				continue
			}
			for _, req := range tf.Requirements {
				for _, a := range req.Code {
					pathPart, _ := splitAnchor(a)
					anchoredSet[filepath.ToSlash(pathPart)] = true
				}
			}
		}
	}

	for _, rel := range files {
		if !codeExtensions[filepath.Ext(rel)] || isTestFile(rel) {
			continue // la métrica es sobre código de producto, no sobre tests
		}
		total++
		if anchoredSet[rel] {
			anchored++
		}
	}
	return anchored, total
}

func readCoverageBaseline(projectDir string) (coverageBaseline, bool) {
	var b coverageBaseline
	data, err := os.ReadFile(coverageBaselinePath(projectDir))
	if err != nil || json.Unmarshal(data, &b) != nil {
		return b, false
	}
	return b, true
}

// writeCoverageBaseline persiste la marca (best-effort: un fallo de escritura
// no rompe la medición, solo pierde el ratchet de esta corrida).
func writeCoverageBaseline(projectDir string, percent float64, anchored, total int) {
	b := coverageBaseline{
		SchemaVersion: schemaVersionCurrent,
		Percent:       percent,
		Anchored:      anchored,
		Total:         total,
		At:            nowUTC(),
	}
	path := coverageBaselinePath(projectDir)
	if os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return
	}
	out, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return
	}
	_ = os.WriteFile(path, append(out, '\n'), 0o644)
}
