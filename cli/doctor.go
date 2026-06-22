package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// Modelo de trace.json — la matriz de trazabilidad que emite sf-check.
//
// "requirements" es un OBJETO json (mapa req-id → info), no un array. En Go eso
// es un map[string]traceReq. Los maps de Go iteran en orden ALEATORIO, así que
// cuando recorramos los requirements vamos a ordenar las claves para una salida
// estable.
// ----------------------------------------------------------------------------

type traceFile struct {
	Feature      string              `json:"feature"`
	Requirements map[string]traceReq `json:"requirements"`
}

type traceReq struct {
	Code   []string `json:"code"`
	Test   []string `json:"test"`
	Status string   `json:"status"`
}

// runDoctor es el punto de entrada de `sf doctor`. Por ahora el único chequeo es
// drift; cuando doctor crezca (integridad de estado, staleness, schema) el flag
// --drift servirá para aislarlo. Hoy `sf doctor` y `sf doctor --drift` hacen lo
// mismo.
func runDoctor(args []string) int {
	projectDir := "."
	quiet := false
	runTests := ""
	install := false
	global := false

	// Parseo manual de flags. i++ extra cuando un flag consume su valor.
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--drift":
			// reconocido; sin efecto distinto todavía (drift es el único chequeo)
		case a == "--install":
			install = true
		case a == "--global":
			global = true
		case a == "--quiet":
			quiet = true
		case a == "--run-tests":
			if i+1 < len(args) {
				runTests = args[i+1]
				i++
			} else {
				runTests = "pytest -q {test}"
			}
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf doctor: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	if install {
		return runDoctorInstall(projectDir, global)
	}
	return runDrift(projectDir, runTests, quiet)
}

// runDrift recorre todos los trace.json del proyecto y reporta los anchors que
// ya no existen en el código. Devuelve 1 si hay drift, 0 si todo en sync.
func runDrift(projectDir, runTests string, quiet bool) int {
	specforge := filepath.Join(projectDir, "specforge")
	if info, err := os.Stat(specforge); err != nil || !info.IsDir() {
		fmt.Printf("No specforge/ under %s — nothing to check.\n", projectDir)
		return 0
	}

	traces := loadTraces(specforge)
	var drifts []string
	for _, t := range traces {
		// El operador `...` expande el slice devuelto como argumentos variádicos
		// de append (concatena dos slices).
		drifts = append(drifts, checkFeatureDrift(projectDir, t.feature, t.path, runTests)...)
	}

	if !quiet {
		fmt.Printf("Checked %d feature(s) with trace.json.\n", len(traces))
	}
	for _, d := range drifts {
		fmt.Printf("  DRIFT: %s\n", d)
	}
	if len(drifts) > 0 {
		fmt.Printf("\nDRIFT: %d anchor(s) diverged from the spec.\n", len(drifts))
		return 1
	}
	fmt.Println("\nOK: specs and code in sync.")
	return 0
}

// traceEntry empareja una feature con la ruta de su trace.json.
type traceEntry struct {
	feature string
	path    string
}

// loadTraces busca todos los specforge/{archive,features}/*/trace.json. El
// nombre de la feature es el nombre del directorio contenedor.
func loadTraces(specforge string) []traceEntry {
	var out []traceEntry
	for _, base := range []string{"archive", "features"} {
		matches, _ := filepath.Glob(filepath.Join(specforge, base, "*", "trace.json"))
		sort.Strings(matches) // salida estable
		for _, m := range matches {
			out = append(out, traceEntry{
				feature: filepath.Base(filepath.Dir(m)),
				path:    m,
			})
		}
	}
	return out
}

// checkFeatureDrift devuelve los mensajes de drift de una feature. runTests
// vacío = solo chequeo estático (no corre tests). Esta función la reutiliza
// `sf status` para su columna drift.
func checkFeatureDrift(projectDir, feature, tracePath, runTests string) []string {
	data, err := os.ReadFile(tracePath)
	if err != nil {
		return []string{fmt.Sprintf("%s: unreadable trace.json (%v)", feature, err)}
	}
	var tf traceFile
	if err := json.Unmarshal(data, &tf); err != nil {
		return []string{fmt.Sprintf("%s: unreadable trace.json (%v)", feature, err)}
	}

	// Orden estable de los requirements.
	reqIDs := make([]string, 0, len(tf.Requirements))
	for id := range tf.Requirements {
		reqIDs = append(reqIDs, id)
	}
	sort.Strings(reqIDs)

	var drifts []string
	for _, req := range reqIDs {
		info := tf.Requirements[req]
		for _, anchor := range info.Code {
			if ok, reason := checkAnchor(projectDir, anchor); !ok {
				drifts = append(drifts, fmt.Sprintf("%s / %s: %s", feature, req, reason))
			}
		}
		if runTests != "" {
			for _, testID := range info.Test {
				if ok, reason := runTest(projectDir, testID, runTests); !ok {
					drifts = append(drifts, fmt.Sprintf("%s / %s: %s", feature, req, reason))
				}
			}
		}
	}
	return drifts
}

// splitAnchor separa un anchor en (ruta, símbolo). Soporta dos formas:
//   - path::nodeid  → id de test estilo pytest (tests/foo.py::Clase::test_x).
//     El símbolo es el ÚLTIMO segmento del nodeid, sin el sufijo de
//     parametrización: test_x[caso-1] → test_x.
//   - path:symbol   → anchor de código (src/foo.go:Parse, src/bar.py:Klass.method)
//     o de test estilo Go (foo_test.go:TestParse).
//
// Si no hay separador, el anchor es solo una ruta y el símbolo queda "".
//
// Concepto Go: hay que chequear "::" ANTES que ":" porque "::" contiene ":".
// Partir por el último ":" cortaba mal un nodeid de pytest (tests/foo.py: +
// :test_x) — ese era el bug que impedía verificar test anchors.
func splitAnchor(anchor string) (path, symbol string) {
	// strings.Cut parte por la PRIMERA aparición de "::" y devuelve
	// (antes, despues, encontrado). Es justo lo que queremos para separar la ruta.
	if before, node, found := strings.Cut(anchor, "::"); found {
		// pytest anida con "::" (archivo::Clase::test). El símbolo real es el
		// último tramo, así que si quedan más "::" nos quedamos con lo de después.
		if i := strings.LastIndex(node, "::"); i >= 0 {
			node = node[i+2:]
		}
		// pytest parametriza con sufijo "[caso]"; nos quedamos con el nombre.
		if k := strings.IndexByte(node, '['); k >= 0 {
			node = node[:k]
		}
		return before, node
	}
	if i := strings.LastIndex(anchor, ":"); i >= 0 {
		return anchor[:i], anchor[i+1:]
	}
	return anchor, ""
}

// checkAnchor (estático): ¿sigue existiendo `path:símbolo`? Devuelve (ok, razón).
// Si no hay símbolo, el anchor es solo un path. El símbolo se busca por su
// identificador final (de `Clase.metodo` toma `metodo`) con límites de palabra.
func checkAnchor(projectDir, anchor string) (bool, string) {
	pathPart, symbol := splitAnchor(anchor)

	data, err := os.ReadFile(filepath.Join(projectDir, pathPart))
	if err != nil {
		return false, fmt.Sprintf("file gone: %s", pathPart)
	}
	if symbol != "" {
		parts := strings.Split(symbol, ".")
		ident := parts[len(parts)-1]
		// QuoteMeta escapa cualquier carácter especial de regex en el símbolo.
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(ident) + `\b`)
		if !re.Match(data) {
			return false, fmt.Sprintf("symbol gone: %s not in %s", symbol, pathPart)
		}
	}
	return true, ""
}

// runTest corre un test (modo dinámico, opt-in). El template trae "{test}" que
// reemplazamos por el id. Usamos `sh -c` para soportar pipes/flags, igual que
// shell=True en Python. context.WithTimeout corta a los 300s.
func runTest(projectDir, testID, template string) (bool, string) {
	cmd := strings.ReplaceAll(template, "{test}", testID)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel() // libera el timer pase lo que pase

	c := exec.CommandContext(ctx, "sh", "-c", cmd)
	c.Dir = projectDir

	err := c.Run()
	if err == nil {
		return true, ""
	}
	// errors.As distingue "el test corrió y falló" (ExitError, exit != 0) de "no
	// se pudo ni lanzar" (binario ausente, timeout, etc.).
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false, fmt.Sprintf("test failed: %s", testID)
	}
	return false, fmt.Sprintf("could not run test: %v", err)
}
