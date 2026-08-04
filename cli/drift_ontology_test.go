package main

import (
	"encoding/json"
	"path/filepath"
	"runtime"
	"testing"
)

// ----------------------------------------------------------------------------
// F2 de CLI-COMO-FUENTE (R1/R2/R3): la ontología de drift.
//
// El test central es la TABLA DE DECISIÓN. Las cuatro filas de §1 de la spec no
// son variantes de lo mismo: cada una pide una acción distinta del humano, y
// aplanarlas en un booleano es exactamente lo que hacía el motor viejo.
// ----------------------------------------------------------------------------

// testCmdTemplate arma un comando de test que pasa o falla según se pida, sin
// depender de que haya pytest/go instalado. `exit 0` / `exit 1` es lo mínimo que
// entiende cualquier shell, y runTest ya shellea con sh -c / cmd /c.
func testCmdTemplate(pass bool) string {
	code := "1"
	if pass {
		code = "0"
	}
	if runtime.GOOS == "windows" {
		return "cmd /c exit " + code
	}
	return "exit " + code
}

// makeOntologyProject arma una feature con UN requisito, parametrizable en las
// dos variables que decide la tabla: ¿resuelve el anchor de código? ¿nombra un
// test?
func makeOntologyProject(t *testing.T, codeResolves, hasTest bool) string {
	t.Helper()
	dir := t.TempDir()
	body := "def slugify(text):\n    return text\n"
	if !codeResolves {
		body = "def something_else():\n    return 1\n"
	}
	writeFile(t, filepath.Join(dir, "src/slug.py"), body)
	writeFile(t, filepath.Join(dir, "tests/test_slug.py"), "def test_slugify():\n    assert True\n")

	tests := `["tests/test_slug.py::test_slugify"]`
	if !hasTest {
		tests = `[]`
	}
	writeFile(t, filepath.Join(dir, "specforge/features/slugify/trace.json"),
		`{"feature":"slugify","requirements":{"R1":{"code":["src/slug.py:slugify"],"test":`+tests+`,"status":"ok"}}}`)
	return dir
}

// TestDriftDecisionTable recorre las cuatro filas de la tabla de §1.
func TestDriftDecisionTable(t *testing.T) {
	cases := []struct {
		name         string
		codeResolves bool
		hasTest      bool
		testPasses   bool
		wantNotImpl  int
		wantDiff     int
		wantUnverif  int
	}{
		{
			name: "anchor no resuelve → 1, no implementado",
			// El test rojo NO agrega una segunda categoría: el ancla manda (R3).
			codeResolves: false, hasTest: true, testPasses: false,
			wantNotImpl: 2, wantDiff: 0, wantUnverif: 0,
		},
		{
			name:         "anchor resuelve + test falla → 2, implementado distinto",
			codeResolves: true, hasTest: true, testPasses: false,
			wantNotImpl: 0, wantDiff: 1, wantUnverif: 0,
		},
		{
			name:         "anchor resuelve + no hay test → 3, sin verificar",
			codeResolves: true, hasTest: false, testPasses: true,
			wantNotImpl: 0, wantDiff: 0, wantUnverif: 1,
		},
		{
			name:         "anchor resuelve + test pasa → sin drift",
			codeResolves: true, hasTest: true, testPasses: true,
			wantNotImpl: 0, wantDiff: 0, wantUnverif: 0,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := makeOntologyProject(t, c.codeResolves, c.hasTest)
			got := classifyFeatureDrift(dir, "slugify",
				filepath.Join(dir, "specforge/features/slugify/trace.json"),
				testCmdTemplate(c.testPasses))

			if n := len(got.notImplemented); n != c.wantNotImpl {
				t.Errorf("not_implemented=%d, want %d: %+v", n, c.wantNotImpl, got.notImplemented)
			}
			if n := len(got.implementedDifferently); n != c.wantDiff {
				t.Errorf("implemented_differently=%d, want %d: %+v", n, c.wantDiff, got.implementedDifferently)
			}
			if n := len(got.unverified); n != c.wantUnverif {
				t.Errorf("unverified=%d, want %d: %+v", n, c.wantUnverif, got.unverified)
			}
		})
	}
}

// TestDriftCategory2Undetermined es el test que sostiene la regla anti-omisión:
// sin --run-tests la categoría 2 no se pudo mirar, y el reporte tiene que
// DECIRLO. Una lista vacía se lee como "no hay hallazgos", que es una afirmación
// que nadie verificó.
func TestDriftCategory2Undetermined(t *testing.T) {
	dir := makeOntologyProject(t, true, true)

	t.Run("sin --run-tests el campo es un objeto undetermined, no []", func(t *testing.T) {
		out := captureStdout(t, func() { runDrift(dir, "", true, true) })
		var probe map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &probe); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		raw := string(probe["implemented_differently"])
		if raw == "[]" || raw == "null" {
			t.Fatalf("implemented_differently=%s; la ausencia de información se está leyendo como limpio", raw)
		}
		var u undetermined
		if err := json.Unmarshal(probe["implemented_differently"], &u); err != nil {
			t.Fatalf("no es un objeto undetermined: %v (%s)", err, raw)
		}
		if u.Status != "undetermined" || u.Reason == "" {
			t.Errorf("got %+v, want status=undetermined con un motivo", u)
		}
	})

	t.Run("con --run-tests el campo es una lista", func(t *testing.T) {
		out := captureStdout(t, func() { runDrift(dir, testCmdTemplate(true), true, true) })
		var probe map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &probe); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		if raw := string(probe["implemented_differently"]); raw != "[]" {
			t.Errorf("implemented_differently=%s, want []", raw)
		}
	})

	t.Run("la salida humana también lo dice", func(t *testing.T) {
		out := captureStdout(t, func() { runDrift(dir, "", true, false) })
		if !contains(out, "undetermined") || !contains(out, "--run-tests") {
			t.Errorf("la salida humana no declara el límite de la categoría 2:\n%s", out)
		}
	})
}

func TestDriftReportJSON(t *testing.T) {
	t.Run("las cuatro categorías vienen siempre, incluso vacías", func(t *testing.T) {
		dir := makeOntologyProject(t, true, true)
		out := captureStdout(t, func() { runDrift(dir, testCmdTemplate(true), true, true) })
		var probe map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &probe); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		for _, k := range []string{"checked", "not_implemented", "implemented_differently", "unverified", "out_of_spec"} {
			if _, ok := probe[k]; !ok {
				t.Errorf("falta la clave %q — una sección ausente es indistinguible de una que no se miró", k)
			}
		}
	})

	t.Run("exit 1 sólo por divergencia de comportamiento", func(t *testing.T) {
		// Un requisito sin test es un HUECO (categoría 3), no una contradicción
		// entre spec y código. Si moviera el exit code, `sf doctor` saldría 1 en
		// todo repo brownfield y se dejaría de correr.
		dir := makeOntologyProject(t, true, false)
		if code := runDrift(dir, "", true, true); code != 0 {
			t.Errorf("exit=%d, want 0: la categoría 3 se reporta pero no es divergencia", code)
		}
		// La categoría 1 sí.
		broken := makeOntologyProject(t, false, true)
		if code := runDrift(broken, "", true, true); code != 1 {
			t.Errorf("exit=%d, want 1: un anchor que no resuelve es divergencia", code)
		}
	})

	t.Run("proyecto sin specforge/ emite el mismo esquema", func(t *testing.T) {
		out := captureStdout(t, func() { runDrift(t.TempDir(), "", true, true) })
		var rep driftReport
		if err := json.Unmarshal([]byte(out), &rep); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		if rep.Checked != 0 {
			t.Errorf("checked=%d, want 0", rep.Checked)
		}
	})

	t.Run("out_of_spec trae las rutas que ningún trace cita", func(t *testing.T) {
		dir := makeOntologyProject(t, true, true)
		writeFile(t, filepath.Join(dir, "src/orphan.py"), "def orphan():\n    pass\n")
		rep := buildDriftReport(dir, "")
		found := false
		for _, f := range rep.OutOfSpec {
			if f == "src/orphan.py" {
				found = true
			}
		}
		if !found {
			t.Errorf("out_of_spec=%v, want que incluya src/orphan.py", rep.OutOfSpec)
		}
	})
}

// TestCheckFeatureDriftStaysBehavioural fija el contrato con sus DOS consumidores
// existentes: `sf status` (columna drift) y `sf verify` (chequeo 4). Los dos
// preguntan "¿spec y código se contradicen?", no "¿falta trabajo?". Si los huecos
// se colaran acá, toda feature sin tests aparecería en DRIFT y la columna dejaría
// de significar algo.
func TestCheckFeatureDriftStaysBehavioural(t *testing.T) {
	dir := makeOntologyProject(t, true, false) // resuelve, sin test → categoría 3
	drifts := checkFeatureDrift(dir, "slugify", filepath.Join(dir, "specforge/features/slugify/trace.json"), "")
	if len(drifts) != 0 {
		t.Errorf("drifts=%v, want vacío: un hueco de verificación no es drift de comportamiento", drifts)
	}
}
