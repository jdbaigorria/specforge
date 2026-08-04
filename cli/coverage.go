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
	badge := false
	history := false
	for _, a := range args {
		switch {
		case a == "--json":
			asJSON = true
		case a == "--badge":
			badge = true
		case a == "--history":
			history = true
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

	// --history: solo la tendencia, sin medir de nuevo (lo medido ya está).
	if history {
		return printCoverageHistory(projectDir)
	}

	anchored, total := computeSpecCoverage(projectDir)
	percent := 0.0
	if total > 0 {
		percent = 100 * float64(anchored) / float64(total)
	}

	// F3: cada medición queda en la historia (la TENDENCIA es la métrica que
	// cuenta en adopción brownfield — el número de hoy importa menos que la
	// pendiente del mes).
	appendCoverageHistory(projectDir, percent, anchored, total)

	// F3: badge SVG auto-contenido (estilo shields, flat) — commiteable y
	// referenciable desde el README: ![spec coverage](specforge/coverage-badge.svg)
	if badge {
		p := filepath.Join(projectDir, "specforge", "coverage-badge.svg")
		if err := os.WriteFile(p, []byte(coverageBadgeSVG(percent)), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "sf coverage: cannot write badge (%v)\n", err)
			return 1
		}
		fmt.Printf("badge → %s\n", filepath.ToSlash(p))
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

// ── F3: historia + badge (spec coverage como métrica pública) ────────────────

// coveragePoint es una medición histórica (una línea del JSONL).
type coveragePoint struct {
	At       string  `json:"at"`
	Percent  float64 `json:"percent"`
	Anchored int     `json:"anchored"`
	Total    int     `json:"total"`
}

func coverageHistoryPath(projectDir string) string {
	return filepath.Join(projectDir, "specforge", ".state", "coverage-history.jsonl")
}

// coverageHistoryMax: cap de líneas — misma filosofía de estado acotado que C6.
const coverageHistoryMax = 500

// appendCoverageHistory registra la medición (best-effort, como la telemetría).
func appendCoverageHistory(projectDir string, percent float64, anchored, total int) {
	path := coverageHistoryPath(projectDir)
	if os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return
	}
	line, err := json.Marshal(coveragePoint{At: nowUTC(), Percent: percent, Anchored: anchored, Total: total})
	if err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(line, '\n'))
	_ = f.Close()
	// Poda: conservamos la mitad más reciente al superar el cap.
	if data, err := os.ReadFile(path); err == nil {
		lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
		if len(lines) > coverageHistoryMax {
			_ = os.WriteFile(path, []byte(strings.Join(lines[len(lines)/2:], "\n")+"\n"), 0o644)
		}
	}
}

// printCoverageHistory muestra la tendencia: cada punto con una barra
// proporcional — un sparkline honesto en texto plano.
func printCoverageHistory(projectDir string) int {
	data, err := os.ReadFile(coverageHistoryPath(projectDir))
	if err != nil {
		fmt.Println("No coverage history yet — run `sf coverage` to record the first point.")
		return 0
	}
	var points []coveragePoint
	for line := range strings.SplitSeq(strings.TrimRight(string(data), "\n"), "\n") {
		var p coveragePoint
		if json.Unmarshal([]byte(line), &p) == nil {
			points = append(points, p)
		}
	}
	if len(points) == 0 {
		fmt.Println("No coverage history yet.")
		return 0
	}
	fmt.Printf("spec coverage trend (%d point(s)):\n\n", len(points))
	show := points
	if len(show) > 20 {
		show = show[len(show)-20:] // los últimos 20 alcanzan para ver la pendiente
	}
	for _, p := range show {
		bar := strings.Repeat("█", int(p.Percent/4)) // 25 chars = 100%
		fmt.Printf("  %s  %5.1f%%  %s\n", p.At, p.Percent, bar)
	}
	first, last := points[0], points[len(points)-1]
	fmt.Printf("\n%+.1f%% since %s (%d → %d anchored files)\n",
		last.Percent-first.Percent, first.At, first.Anchored, last.Anchored)
	return 0
}

// coverageBadgeSVG genera el badge (estilo shields flat, auto-contenido: sin
// fuentes ni requests externos). El color sube con la cobertura — la señal
// visual del ratchet.
func coverageBadgeSVG(percent float64) string {
	color := "#e05d44" // rojo
	switch {
	case percent >= 75:
		color = "#4c1" // verde
	case percent >= 50:
		color = "#dfb317" // amarillo
	case percent >= 25:
		color = "#fe7d37" // naranja
	}
	label := "spec coverage"
	value := fmt.Sprintf("%.1f%%", percent)
	// Ancho aproximado: ~6.5px por carácter + padding (suficiente para un badge).
	lw := len(label)*7 + 10
	vw := len(value)*7 + 10
	w := lw + vw
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="20" role="img" aria-label="%s: %s">
  <rect width="%d" height="20" fill="#555"/>
  <rect x="%d" width="%d" height="20" fill="%s"/>
  <g fill="#fff" text-anchor="middle" font-family="Verdana,Geneva,DejaVu Sans,sans-serif" font-size="11">
    <text x="%d" y="14">%s</text>
    <text x="%d" y="14">%s</text>
  </g>
</svg>
`, w, label, value, lw, lw, vw, color, lw/2, label, lw+vw/2, value)
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
