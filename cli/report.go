package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"os"
	"sort"
	"strings"
)

// ----------------------------------------------------------------------------
// Causalidad test→requirement (A2/R4, cierra D3).
//
// El fraude que ataca: con el verdict actual, un modelo puede anclar cada
// requirement a un test REAL y trivialmente verde (test_healthcheck) y pasar.
// La mentira ya no es "los tests pasan" — es "ESTE test prueba ESTE requirement".
//
// La respuesta: `sf check run` deja de mirar solo el exit code y parsea un
// reporte ESTRUCTURADO de la corrida (qué test corrió y con qué resultado).
// El verdict entonces exige que cada test nombrado en trace.json aparezca como
// PASSED en la corrida sellada. La máquina ata el nombre al resultado; el LLM
// ya no puede narrar la relación.
//
// Formatos soportados (empezamos con los dos del roadmap):
//   - build.report = "go-json":  test_cmd emite `go test -json` por stdout.
//   - build.report = "junit":    test_cmd contiene el placeholder {report}; el
//     CLI lo sustituye por una ruta y parsea el JUnit XML resultante
//     (pytest --junitxml={report}, y casi todo runner sabe emitir JUnit).
// Sin build.report configurado, el verdict conserva las garantías previas
// (suite verde y fresca) — la causalidad es opt-in porque requiere que el
// proyecto declare cómo produce el reporte.
// ----------------------------------------------------------------------------

// buildReports: valores válidos de build.report en la constitución.
var buildReports = map[string]bool{"go-json": true, "junit": true}

// parseGoTestJSON extrae los resultados por test de un stream `go test -json`.
// Cada línea es un evento {"Action":..,"Test":..}; nos quedamos con las
// acciones TERMINALES (pass/fail/skip) de eventos con Test (los eventos de
// paquete no traen Test). Líneas que no son JSON (output del build, prints del
// propio test) se saltean — el stream real viene mezclado.
func parseGoTestJSON(out string) map[string]string {
	tests := map[string]string{}
	for line := range strings.SplitSeq(out, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var ev struct {
			Action string `json:"Action"`
			Test   string `json:"Test"`
		}
		if json.Unmarshal([]byte(line), &ev) != nil || ev.Test == "" {
			continue
		}
		switch ev.Action {
		case "pass", "fail", "skip":
			tests[ev.Test] = ev.Action
		}
	}
	return tests
}

// Tipos mínimos del schema JUnit. Los suites pueden anidar (testsuites →
// testsuite → testcase); xml.Unmarshal ignora el NOMBRE del elemento raíz, así
// que el mismo struct sirve para un archivo con raíz <testsuites> o <testsuite>.
type junitCase struct {
	Name     string     `xml:"name,attr"`
	Failures []struct{} `xml:"failure"`
	Errors   []struct{} `xml:"error"`
	Skipped  []struct{} `xml:"skipped"`
}

type junitSuite struct {
	Cases  []junitCase  `xml:"testcase"`
	Suites []junitSuite `xml:"testsuite"`
}

// parseJUnitXML lee el reporte generado por la corrida y devuelve test→estado.
// nil si el archivo no existe o no parsea (el caller decide qué significa).
func parseJUnitXML(path string) map[string]string {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var root junitSuite
	if xml.Unmarshal(data, &root) != nil {
		return nil
	}
	tests := map[string]string{}
	collectJUnit(root, tests)
	return tests
}

func collectJUnit(s junitSuite, tests map[string]string) {
	for _, c := range s.Cases {
		status := "pass"
		switch {
		case len(c.Failures) > 0 || len(c.Errors) > 0:
			status = "fail"
		case len(c.Skipped) > 0:
			status = "skip"
		}
		tests[c.Name] = status
	}
	for _, sub := range s.Suites {
		collectJUnit(sub, tests)
	}
}

// testRefLeaf normaliza una ref de test del trace al NOMBRE que aparece en el
// reporte: "tests/t.py::TestClass::test_fn" → "test_fn" (JUnit reporta el name
// del caso), "cli/gate_test.go:TestFoo" → "TestFoo" (go -json reporta el nombre
// de la función). Es la bisagra entre el formato de anchor (path:symbol) y el
// formato de los reporters.
func testRefLeaf(ref string) string {
	s := ref
	if i := strings.LastIndex(s, "::"); i >= 0 {
		s = s[i+2:]
	}
	if i := strings.LastIndex(s, ":"); i >= 0 {
		s = s[i+1:]
	}
	return s
}

// causalityReasons devuelve una razón por cada test nombrado en el trace que NO
// aparece como passed en la corrida sellada. La invocan las precondiciones del
// verdict SOLO cuando la corrida trae reporte estructurado (len(tests) > 0):
// sin reporte no hay evidencia por-test y no fabricamos garantías.
func causalityReasons(tf *traceFile, tests map[string]string) []string {
	var reasons []string
	ids := make([]string, 0, len(tf.Requirements))
	for id := range tf.Requirements {
		ids = append(ids, id)
	}
	sort.Strings(ids) // orden estable → reporte reproducible
	for _, req := range ids {
		for _, tref := range tf.Requirements[req].Test {
			leaf := testRefLeaf(tref)
			status, ran := tests[leaf]
			switch {
			case !ran:
				reasons = append(reasons, fmt.Sprintf(
					"%s: test %q did not run in the sealed `sf check run` (test→requirement causality unproven)", req, tref))
			case status != "pass":
				reasons = append(reasons, fmt.Sprintf(
					"%s: test %q ran but was %s (not passed) in the sealed run", req, tref, status))
			}
		}
	}
	return reasons
}
