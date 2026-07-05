package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"
)

// ----------------------------------------------------------------------------
// `sf metrics` — mide el TAMAÑO del output que la IA consume.
//
// La IA no consume los archivos crudos; consume lo que emiten los getters de
// contexto (`sf context current`, `sf context for-wave`). Este comando mide ese
// output para tener un BASELINE: cuántos tokens/chars cuesta cada slice hoy, en
// JSON. Sirve para decidir con datos si vale la pena un formato más compacto
// (TOON) en el OUTPUT — sin tocar el storage, que sigue siendo JSON.
// ----------------------------------------------------------------------------

// runMetrics es el punto de entrada de `sf metrics ...`. Hoy la única sub-acción
// es `context`.
func runMetrics(args []string) int {
	if len(args) == 0 || args[0] != "context" {
		fmt.Fprintln(os.Stderr, "usage: sf metrics context [project_dir] [--feature=NAME]")
		return 2
	}
	args = args[1:] // descartamos "context"

	projectDir := "."
	feature := ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf metrics: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	return metricsContext(projectDir, feature)
}

// sliceMetric es una fila de la tabla: el slice medido y su tamaño.
type sliceMetric struct {
	label  string
	tokens int
	chars  int
}

// metricsContext recolecta las métricas y las imprime. La recolección vive en
// collectContextMetrics (testeable sin tocar stdout); esta función solo formatea.
func metricsContext(projectDir, feature string) int {
	rows, code := collectContextMetrics(projectDir, feature)
	if code != 0 {
		return code
	}

	header := "context output size"
	if feature != "" {
		header += " — " + feature
	}
	fmt.Printf("%s\n\n", header)

	cols := []string{"slice", "~tokens", "chars"}
	var table [][]string
	for _, r := range rows {
		table = append(table, []string{r.label, fmt.Sprintf("%d", r.tokens), fmt.Sprintf("%d", r.chars)})
	}
	renderTable(cols, table)
	fmt.Println("\n~tokens is a heuristic (alnum runs + punctuation), not a model tokenizer.")
	return 0
}

// collectContextMetrics arma cada slice REUSANDO los mismos builders que los
// getters (buildBreadcrumb / buildCurrentContext / loadWaveContext) y mide su
// tamaño. No reimplementa el slicing: mide exactamente lo que la IA recibe.
//
// Si no se pasa --feature, usa la feature activa que deriva computeCurrentState.
// Devuelve (filas, code): code 4 si falta features.json, 2 si es inválido, 0 ok.
func collectContextMetrics(projectDir, feature string) ([]sliceMetric, int) {
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf metrics: cannot read feature state under %s (%v)\n", projectDir, err)
		return nil, 4
	}
	st := computeCurrentState(ff, projectDir)
	if feature == "" {
		feature = st.Feature // fallback: la feature activa
	}

	var rows []sliceMetric
	measure := func(label, body string) {
		// len([]rune) cuenta CARACTERES (no bytes): un acento es 1 char.
		rows = append(rows, sliceMetric{label, estimateTokens(body), len([]rune(body))})
	}

	// Tier barato: el breadcrumb (texto plano, una línea — lo que un hook
	// inyectaría cada turno).
	measure("breadcrumb", buildBreadcrumb(st, projectDir))

	// Tier completo: `context current` (JSON con el slice del paso actual).
	if b, err := json.MarshalIndent(buildCurrentContext(st, projectDir), "", "  "); err == nil {
		measure("context current", string(b))
	}

	// Un slice `for-wave` por cada wave del plan de la feature (los slices "ricos"
	// que se consumen durante el build — los candidatos más fuertes a TOON).
	if feature != "" {
		pf := readPlanQuiet(projectDir, feature)
		for _, w := range pf.Waves {
			if ws, ok := loadWaveContext(projectDir, feature, w.N); ok {
				if b, err := json.MarshalIndent(ws, "", "  "); err == nil {
					measure(fmt.Sprintf("for-wave --n=%d", w.N), string(b))
				}
			}
		}
	}
	return rows, 0
}

// estimateTokens aproxima el conteo de tokens SIN un tokenizer real (SpecForge no
// agrega dependencias): agrupa cada corrida de letras/dígitos/_ como UN token y
// cuenta cada signo de puntuación como su propio token; el espacio en blanco
// separa pero no cuenta. Para JSON/specs estructurados se acerca a un BPE porque
// la puntuación ({ } : " ,) suele tokenizar ~1 cada uno. Es para COMPARAR formatos
// del mismo contenido (JSON vs TOON), no una factura exacta de tokens.
func estimateTokens(s string) int {
	n := 0
	inWord := false
	for _, r := range s {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_':
			if !inWord {
				n++ // empieza una palabra nueva
				inWord = true
			}
		case unicode.IsSpace(r):
			inWord = false // el espacio corta la palabra, pero no es un token
		default:
			n++ // puntuación = un token
			inWord = false
		}
	}
	return n
}
