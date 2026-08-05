package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// RM-C5 — el testigo RED (R11 acumulación, R12 precondición).
//
// El CLI ya verificaba que el test existe, corrió, dio verde y es fresco. Lo que
// no verificaba es el ORDEN. Un test escrito DESPUÉS del código afirma lo que el
// código hace, no lo que el requisito pide — el fraude semántico en su forma más
// común y menos malintencionada: el agente implementa, lee su propia
// implementación y escribe un test que la describe. Pasa todo, se sella, y no
// verifica nada.
//
// El principio: no obligar la práctica, REGISTRAR la evidencia. "Este test falló
// alguna vez con otro hash de código" es un hecho de máquina que el LLM no puede
// fabricar.
// ----------------------------------------------------------------------------

func TestRedWitnessAccumulation(t *testing.T) {
	run := func(dir string, at, hash string, tests map[string]string) {
		t.Helper()
		accumulateRedWitnesses(dir, "f", checkResult{
			Feature: "f", At: at, CodeHash: hash, TestCmd: "go test -json ./...", Tests: tests,
		})
	}

	t.Run("un test rojo gana testigo con el hash de ESA corrida", func(t *testing.T) {
		dir := t.TempDir()
		run(dir, "2026-07-30T10:00:00Z", "hash-A", map[string]string{
			"TestEmptyCart": "fail", "TestPersists": "pass",
		})
		rw := readRedWitnesses(dir, "f")
		if len(rw.Witnesses) != 1 {
			t.Fatalf("witnesses=%+v, want sólo el que falló", rw.Witnesses)
		}
		w := rw.Witnesses["TestEmptyCart"]
		if w.CodeHash != "hash-A" || w.FailedAt != "2026-07-30T10:00:00Z" {
			t.Errorf("witness=%+v", w)
		}
	})

	t.Run("APPEND-ONLY: una segunda corrida roja NO mueve el testigo", func(t *testing.T) {
		// Es la regla que hace que el registro valga algo. Si se sobrescribiera,
		// el code_hash avanzaría hasta coincidir con el actual — que es
		// exactamente la condición que R12 rechaza.
		dir := t.TempDir()
		run(dir, "2026-07-30T10:00:00Z", "hash-A", map[string]string{"TestEmptyCart": "fail"})
		run(dir, "2026-07-31T11:00:00Z", "hash-B", map[string]string{"TestEmptyCart": "fail"})

		w := readRedWitnesses(dir, "f").Witnesses["TestEmptyCart"]
		if w.CodeHash != "hash-A" || w.FailedAt != "2026-07-30T10:00:00Z" {
			t.Errorf("el testigo se sobrescribió: %+v — la evidencia histórica sólo sirve si es histórica", w)
		}
	})

	t.Run("una corrida verde posterior no altera el registro", func(t *testing.T) {
		dir := t.TempDir()
		run(dir, "2026-07-30T10:00:00Z", "hash-A", map[string]string{"TestEmptyCart": "fail"})
		run(dir, "2026-08-01T09:00:00Z", "hash-C", map[string]string{"TestEmptyCart": "pass"})

		rw := readRedWitnesses(dir, "f")
		if len(rw.Witnesses) != 1 || rw.Witnesses["TestEmptyCart"].CodeHash != "hash-A" {
			t.Errorf("witnesses=%+v", rw.Witnesses)
		}
	})

	t.Run("sin reporte por-test no se escribe nada y NO es error", func(t *testing.T) {
		dir := t.TempDir()
		run(dir, "2026-07-30T10:00:00Z", "hash-A", nil)
		if fileExists(redWitnessPath(dir, "f")) {
			t.Error("sin build.report no hay nada que atribuir — no debería escribirse")
		}
	})

	t.Run("un registro corrupto se lee como vacío, no rompe", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, redWitnessPath(dir, "f"), "{ esto no es json")
		if len(readRedWitnesses(dir, "f").Witnesses) != 0 {
			t.Error("want registro vacío")
		}
	})
}

// makeRedWitnessProject arma una feature con trace y constitución configurables.
func makeRedWitnessProject(t *testing.T, constitution, trace string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`)
	writeFile(t, filepath.Join(dir, "src/order.go"), "package src\n\nfunc CreateOrder() {}\n")
	writeFile(t, filepath.Join(dir, "src/order_test.go"),
		"package src\n\nfunc TestPersists() {}\nfunc TestEmptyCart() {}\n")
	writeFile(t, filepath.Join(dir, "specforge/constitution.json"), constitution)
	writeFile(t, filepath.Join(dir, "specforge/features/f/trace.json"), trace)
	return dir
}

const redWitnessOn = `{"identity_md":"x","build":{"report":"go-json","test_cmd":"go test -json ./..."},
	"verification":{"require_red_witness":true}}`

const v1Trace = `{"feature":"f","requirements":{"R1":{
	"code":["src/order.go:CreateOrder"],
	"test":["src/order_test.go:TestPersists"],"status":"ok"}}}`

// redReasons deja sólo las razones del testigo RED.
func redReasons(dir string) []string {
	var out []string
	for _, is := range verdictIssues(dir, "f") {
		if strings.Contains(is.Reason, "RED witness") || strings.Contains(is.Reason, "require_red_witness") {
			out = append(out, is.Reason)
		}
	}
	return out
}

// TestRedWitnessPrecondition — R12.
func TestRedWitnessPrecondition(t *testing.T) {
	t.Run("apagado por default: no cambia nada para nadie", func(t *testing.T) {
		dir := makeRedWitnessProject(t, `{"identity_md":"x","build":{"report":"go-json"}}`, v1Trace)
		if rs := redReasons(dir); len(rs) != 0 {
			t.Errorf("want 0 razones con el flag apagado, got %v", rs)
		}
	})

	t.Run("encendido, un test sin testigo se rechaza nombrándolo", func(t *testing.T) {
		dir := makeRedWitnessProject(t, redWitnessOn, v1Trace)
		rs := redReasons(dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "TestPersists") {
			t.Fatalf("want rechazo nombrando el test, got %v", rs)
		}
		if !strings.Contains(rs[0], "never observed failing") {
			t.Errorf("el mensaje debe decir qué falta: %q", rs[0])
		}
	})

	t.Run("un testigo contra el código ACTUAL no vale", func(t *testing.T) {
		// Diría que el test falla contra el mismo código que hoy lo hace pasar:
		// nunca falló contra otra cosa, así que no prueba nada sobre el orden.
		dir := makeRedWitnessProject(t, redWitnessOn, v1Trace)
		accumulateRedWitnesses(dir, "f", checkResult{
			At: "2026-07-30T10:00:00Z", CodeHash: codeHash(dir), TestCmd: "go test",
			Tests: map[string]string{"TestPersists": "fail"},
		})
		rs := redReasons(dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "against the current code") {
			t.Errorf("want rechazo por hash igual, got %v", rs)
		}
	})

	t.Run("un testigo contra OTRO código habilita", func(t *testing.T) {
		dir := makeRedWitnessProject(t, redWitnessOn, v1Trace)
		accumulateRedWitnesses(dir, "f", checkResult{
			At: "2026-07-30T10:00:00Z", CodeHash: "hash-de-antes", TestCmd: "go test",
			Tests: map[string]string{"TestPersists": "fail"},
		})
		if rs := redReasons(dir); len(rs) != 0 {
			t.Errorf("want 0 razones, got %v", rs)
		}
	})

	t.Run("sin build.report rechaza pidiendo la config, nunca aprueba en silencio", func(t *testing.T) {
		// Sin reporte por-test los testigos NUNCA se acumulan. Aprobar sería lo
		// peor: el proyecto creería tener una garantía que nunca se evaluó.
		dir := makeRedWitnessProject(t,
			`{"identity_md":"x","verification":{"require_red_witness":true}}`, v1Trace)
		rs := redReasons(dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "build.report is not configured") {
			t.Fatalf("want rechazo pidiendo la config, got %v", rs)
		}
	})

	t.Run("un requisito verificado por benchmark queda EXENTO", func(t *testing.T) {
		// RM-C4: no falló nunca como test porque nunca fue uno.
		dir := makeRedWitnessProject(t, redWitnessOn,
			`{"feature":"f","requirements":{"R9":{"code":["src/order.go:CreateOrder"],
			  "test":["src/order_test.go:TestPersists"],"status":"ok"}}}`)
		writeFile(t, filepath.Join(dir, "specforge/features/f/requirements.json"),
			`{"feature":"f","requirements":[{"id":"R9","ears_type":"ubiquitous","behavior":"fast",
			  "verification":"benchmark","acceptance":["p99"]}]}`)
		if rs := redReasons(dir); len(rs) != 0 {
			t.Errorf("want exento, got %v", rs)
		}
	})

	t.Run("los tests de escenario también cuentan", func(t *testing.T) {
		dir := makeRedWitnessProject(t, redWitnessOn,
			`{"feature":"f","requirements":{"R1":{"code":["src/order.go:CreateOrder"],
			  "scenarios":{"R1.1":{"test":["src/order_test.go:TestEmptyCart"]}},"status":"ok"}}}`)
		rs := redReasons(dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "TestEmptyCart") {
			t.Errorf("un test anclado en un escenario no puede quedar fuera: %v", rs)
		}
	})
}

// TestTraceTestsWalksBothLevels es la regresión del hueco que introdujo C1: los
// consumidores que sólo miraban `info.Test` dejaban de ver los tests de un
// proyecto v2, y la ausencia se leía como "todo en orden".
func TestTraceTestsWalksBothLevels(t *testing.T) {
	var tf traceFile
	mustUnmarshal(t, `{"feature":"f","requirements":{"R1":{
		"code":["src/x.go:F"],
		"test":["t.go:TestTransversal"],
		"scenarios":{
			"R1.1":{"test":["t.go:TestUno"]},
			"R1.2":{"test":["t.go:TestDos"]}},
		"status":"ok"}}}`, &tf)

	got := traceTests(&tf, nil)
	if len(got) != 3 {
		t.Fatalf("traceTests=%+v, want 3 (1 transversal + 2 de escenario)", got)
	}
	// El owner distingue de dónde viene cada uno: un mensaje que no lo dice
	// obliga a buscar a mano qué requisito queda sin cubrir.
	wantOwners := []string{"R1", "R1.1", "R1.2"}
	for i, w := range wantOwners {
		if got[i].owner != w {
			t.Errorf("got[%d].owner=%q, want %q", i, got[i].owner, w)
		}
	}

	t.Run("los refs se deduplican: el testigo es del test, no del requisito", func(t *testing.T) {
		var dup traceFile
		mustUnmarshal(t, `{"feature":"f","requirements":{
			"R1":{"code":["src/x.go:F"],"test":["t.go:TestMismo"],"status":"ok"},
			"R2":{"code":["src/x.go:G"],"test":["t.go:TestMismo"],"status":"ok"}}}`, &dup)
		if refs := traceTestRefs(&dup, nil); len(refs) != 1 {
			t.Errorf("refs=%v, want 1 — exigirlo dos veces reportaría el mismo hecho dos veces", refs)
		}
	})

	t.Run("exempt saltea el requisito entero", func(t *testing.T) {
		if got := traceTests(&tf, map[string]bool{"R1": true}); len(got) != 0 {
			t.Errorf("got=%+v, want vacío", got)
		}
	})
}
