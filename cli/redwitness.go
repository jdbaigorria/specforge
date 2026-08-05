package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ----------------------------------------------------------------------------
// RM-C5 — el testigo RED.
//
// EL ESLABÓN QUE FALTA. Hoy el CLI verifica cuatro cosas sobre los tests, todas
// fuertes y todas de máquina:
//
//	cada requisito NOMBRA un test real      → sf trace verify --contract
//	el test CORRIÓ de verdad (exit code)    → sf check run
//	el resultado es FRESCO (hash del código) → checkResult.CodeHash
//	cada test nombrado corrió y PASÓ ahí     → causalityReasons
//
// Lo que no verifica es el ORDEN: que el test existiera antes que el código, o
// que alguna vez haya fallado.
//
// Y ésa es la garantía real de TDD. No es "hay tests" — es el paso RED: el test
// falló antes de que existiera la implementación, lo cual prueba que PUEDE
// fallar. Un test que nunca falló no demostró nada; puede ser tautológico.
//
// Un test escrito DESPUÉS del código tiende exactamente a eso: afirma lo que el
// código hace en vez de lo que el requisito pide. Es el fraude semántico en su
// forma más común y MENOS malintencionada — el agente implementa, lee su propia
// implementación, y escribe un test que la describe. Pasa el contrato, pasa la
// causalidad, se sella, y no verifica nada.
//
// EL PRINCIPIO: NO OBLIGAR LA PRÁCTICA, REGISTRAR LA EVIDENCIA.
//
// SpecForge no impone TDD. TDD es una práctica; la tesis del producto es sellar
// trabajo verificado, no un estilo de trabajo — forzarlo sería exactamente lo
// que le criticamos a los frameworks que revisamos. Pero "este test falló alguna
// vez con un hash de código anterior" es un HECHO DE MÁQUINA, de la misma
// naturaleza que el exit code: el LLM no lo puede fabricar. Eso sí entra.
//
// Por eso el registro se llena solo, sin ceremonia: quien trabaja en RED-GREEN
// lo acumula sin hacer nada especial, y quien no, no se entera.
// ----------------------------------------------------------------------------

type redWitnessFile struct {
	SchemaVersion string `json:"schema_version"`
	Feature       string `json:"feature"`
	// Witnesses: id-hoja del test → la primera vez que se lo vio fallar.
	Witnesses map[string]redWitness `json:"witnesses"`
}

type redWitness struct {
	FailedAt string `json:"failed_at"`
	// CodeHash es el hash del código CUANDO falló. Es el campo que hace útil al
	// testigo: comparado con el hash actual dice si el test falló contra un
	// código distinto del que hoy lo hace pasar.
	CodeHash string `json:"code_hash"`
	TestCmd  string `json:"test_cmd"`
}

func redWitnessPath(projectDir, feature string) string {
	return filepath.Join(projectDir, "specforge", ".state", "red-witness", feature+".json")
}

func readRedWitnesses(projectDir, feature string) redWitnessFile {
	rw := redWitnessFile{Feature: feature, Witnesses: map[string]redWitness{}}
	data, err := os.ReadFile(redWitnessPath(projectDir, feature))
	if err != nil {
		return rw
	}
	var parsed redWitnessFile
	if json.Unmarshal(data, &parsed) != nil {
		return rw // un registro corrupto se trata como vacío, nunca como error
	}
	if parsed.Witnesses == nil {
		parsed.Witnesses = map[string]redWitness{}
	}
	return parsed
}

// accumulateRedWitnesses registra los tests que fallaron en ESTA corrida y
// todavía no tenían testigo (R11).
//
// APPEND-ONLY, y es la regla que hace que el registro valga algo: un testigo
// nunca se sobrescribe ni se borra. Si se sobrescribiera, una segunda corrida
// roja movería el `code_hash` hacia adelante y el testigo terminaría apuntando
// al mismo código que hoy hace pasar el test — que es justo la condición que
// R12 rechaza. La evidencia histórica sólo sirve si es histórica.
//
// Sin reporte por-test (`build.report` sin configurar) no hay nada que atribuir:
// no se escribe y NO es error. Exigir la config acá castigaría a quien no pidió
// esta garantía; quien la quiera la pide con `require_red_witness`.
func accumulateRedWitnesses(projectDir, feature string, res checkResult) {
	if len(res.Tests) == 0 {
		return
	}
	rw := readRedWitnesses(projectDir, feature)

	added := 0
	for test, status := range res.Tests {
		if status != "fail" {
			continue
		}
		if _, exists := rw.Witnesses[test]; exists {
			continue // ya tiene testigo: se conserva el PRIMERO
		}
		rw.Witnesses[test] = redWitness{
			FailedAt: res.At,
			CodeHash: res.CodeHash,
			TestCmd:  res.TestCmd,
		}
		added++
	}
	if added == 0 {
		return
	}

	rw.SchemaVersion = schemaVersionCurrent
	rw.Feature = feature
	path := redWitnessPath(projectDir, feature)
	if os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return // best-effort, como la telemetría: no rompe la corrida
	}
	out, err := json.MarshalIndent(rw, "", "  ")
	if err != nil {
		return
	}
	if os.WriteFile(path, append(out, '\n'), 0o644) == nil {
		fmt.Printf("red witness: recorded %d test(s) that failed against this code\n", added)
	}
}

// ----------------------------------------------------------------------------
// R12 — la precondición opcional del veredicto.
// ----------------------------------------------------------------------------

// redWitnessReasons devuelve por qué el veredicto NO puede otorgarse bajo
// `require_red_witness`. Vacía = todos los tests demostraron poder fallar.
//
// La condición no es "tiene testigo": es "tiene un testigo con un `code_hash`
// DISTINTO del actual". Un testigo con el mismo hash diría que el test falla
// contra el mismo código que hoy lo hace pasar — o sea, que nunca falló contra
// otra cosa, y no prueba nada sobre el orden.
func redWitnessReasons(projectDir, feature string, tf *traceFile, exempt map[string]bool) []string {
	if tf == nil {
		return nil
	}
	rw := readRedWitnesses(projectDir, feature)
	current := codeHash(projectDir)

	var reasons []string
	for _, ref := range traceTestRefs(tf, exempt) {
		leaf := testRefLeaf(ref)
		w, ok := rw.Witnesses[leaf]
		switch {
		case !ok:
			reasons = append(reasons, fmt.Sprintf(
				"%s: no RED witness — this test was never observed failing, so nothing shows it CAN fail", ref))
		case w.CodeHash == current:
			reasons = append(reasons, fmt.Sprintf(
				"%s: RED witness was recorded against the current code — it never failed against a different one", ref))
		}
	}
	return reasons
}

// tracedTest es un test nombrado por el trace, junto al requisito que lo
// nombra. El `owner` importa: un mensaje que dice sólo "el test X no corrió"
// obliga a buscar a mano qué requisito queda sin cubrir.
type tracedTest struct {
	owner string // R# (o R#.N cuando viene de un escenario)
	ref   string
}

// traceTests enumera TODOS los tests que nombra un trace: los de nivel requisito
// (contrato v1, y la cobertura transversal del v2) y los de cada escenario
// (contrato v2, RM-C1). Orden estable.
//
// Recorrer los dos niveles arregla un hueco que introdujo C1: los consumidores
// que sólo miraban `info.Test` dejaban de ver los tests de un proyecto v2 — y
// "no encontré tests" se leía como "todo en orden". El agujero más silencioso
// posible: una garantía que deja de aplicarse sin que nada falle.
//
// `exempt` son los requisitos cuya verificación no es un test (RM-C4): pedirles
// un testigo RED sería pedirle a un benchmark que fallara como test.
func traceTests(tf *traceFile, exempt map[string]bool) []tracedTest {
	var out []tracedTest

	ids := make([]string, 0, len(tf.Requirements))
	for id := range tf.Requirements {
		ids = append(ids, id)
	}
	sort.Strings(ids) // los mapas de Go iteran al azar

	for _, id := range ids {
		if exempt[id] {
			continue
		}
		info := tf.Requirements[id]
		for _, ref := range info.Test {
			out = append(out, tracedTest{owner: id, ref: ref})
		}
		scenarioIDs := make([]string, 0, len(info.Scenarios))
		for sid := range info.Scenarios {
			scenarioIDs = append(scenarioIDs, sid)
		}
		sort.Strings(scenarioIDs)
		for _, sid := range scenarioIDs {
			for _, ref := range info.Scenarios[sid].Test {
				out = append(out, tracedTest{owner: sid, ref: ref})
			}
		}
	}
	return out
}

// traceTestRefs devuelve los refs SIN repetir. El testigo RED es una propiedad
// del test, no del requisito: si dos requisitos nombran el mismo test, exigirlo
// dos veces reportaría el mismo hecho dos veces.
func traceTestRefs(tf *traceFile, exempt map[string]bool) []string {
	seen := map[string]bool{}
	var out []string
	for _, t := range traceTests(tf, exempt) {
		if t.ref == "" || seen[t.ref] {
			continue
		}
		seen[t.ref] = true
		out = append(out, t.ref)
	}
	return out
}
