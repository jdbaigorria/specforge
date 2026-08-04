package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeFile es un helper de test: crea los directorios necesarios y escribe el
// archivo, abortando el test si algo falla. t.Helper() hace que, si este helper
// llama t.Fatalf, el error se reporte en la línea que LO LLAMA, no acá adentro
// — así el mensaje de fallo apunta al lugar útil.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// TestCheckAnchor es un test "table-driven": una tabla de casos y un loop que
// corre cada uno como subtest (t.Run). Es el patrón idiomático en Go para
// probar muchas variantes de la misma función.
func TestCheckAnchor(t *testing.T) {
	// t.TempDir() devuelve un directorio temporal único que se borra solo al
	// terminar el test. No hace falta cleanup manual.
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "src/slug.py"), "def slugify(text):\n    return text\n")

	tests := []struct {
		name   string
		anchor string
		wantOK bool
	}{
		{"symbol present", "src/slug.py:slugify", true},
		{"symbol missing", "src/slug.py:nope", false},
		{"file missing", "src/gone.py:slugify", false},
		{"no symbol, file exists", "src/slug.py", true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ok, reason := checkAnchor(dir, tc.anchor)
			if ok != tc.wantOK {
				t.Errorf("checkAnchor(%q) ok=%v, want %v (reason=%q)", tc.anchor, ok, tc.wantOK, reason)
			}
		})
	}
}

// makeDriftProject arma un proyecto temporal con slug.py y su trace.json
// apuntando a `src/slug.py:slugify`. Devuelve la raíz del proyecto.
func makeDriftProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "src/slug.py"), "def slugify(text):\n    return text\n")
	// Test real al que apunta el trace: trace verify ahora exige que el test
	// nombrado exista (refuerzo FIXBUGHIGH), no solo el código.
	writeFile(t, filepath.Join(dir, "tests/test_slug.py"), "def test_slugify():\n    assert True\n")
	writeFile(t, filepath.Join(dir, "specforge/archive/2026-06-16-slugify/trace.json"),
		`{"feature":"slugify","requirements":{"R1":{"code":["src/slug.py:slugify"],"test":["tests/test_slug.py:test_slugify"],"status":"ok"}}}`)
	return dir
}

// TestRunDrift cubre los mismos casos que el --selftest del check-drift.py,
// verificando el exit code de runDrift (0 = en sync, 1 = drift).
func TestRunDrift(t *testing.T) {
	t.Run("clean project: no drift", func(t *testing.T) {
		dir := makeDriftProject(t)
		if code := runDrift(dir, "", true, false); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})

	t.Run("symbol removed: drift detected", func(t *testing.T) {
		dir := makeDriftProject(t)
		writeFile(t, filepath.Join(dir, "src/slug.py"), "def something_else():\n    return 1\n")
		if code := runDrift(dir, "", true, false); code != 1 {
			t.Errorf("exit=%d, want 1", code)
		}
	})

	t.Run("file removed: drift detected", func(t *testing.T) {
		dir := makeDriftProject(t)
		if err := os.Remove(filepath.Join(dir, "src/slug.py")); err != nil {
			t.Fatalf("remove: %v", err)
		}
		if code := runDrift(dir, "", true, false); code != 1 {
			t.Errorf("exit=%d, want 1", code)
		}
	})

	t.Run("non-SpecForge dir: no-op pass", func(t *testing.T) {
		if code := runDrift(t.TempDir(), "", true, false); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
}

// ---------------------------------------------------------------------------
// DL-5 F1 — el join trace↔feature.
//
// El defecto: `sf feature archive` COPIA la carpeta de la feature a archive/ sin
// borrar la original, y loadTraces globea las DOS bases. Resultado: una feature
// archivada aporta su trace dos veces, con dos nombres distintos —
// "slugify" (de features/) y "2026-06-16-slugify" (de archive/) — así que cada
// ancla divergente se reporta duplicada y la mitad bajo un nombre que no existe
// en features.json.
//
// Nota de Go: los tests de abajo llaman a loadTraces, que es una función NO
// exportada (minúscula). Se puede porque el archivo de test está en el mismo
// paquete (`package main`, no `main_test`) — el test ve todo el paquete.
// ---------------------------------------------------------------------------

// makeArchivedFeature arma un proyecto donde `slugify` ya fue archivada: su
// trace existe en features/ (la copia viva) y en archive/ (la foto histórica),
// que es exactamente el estado que deja `sf feature archive`.
func makeArchivedFeature(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "src/slug.py"), "def slugify(text):\n    return text\n")
	writeFile(t, filepath.Join(dir, "tests/test_slug.py"), "def test_slugify():\n    assert True\n")
	trace := `{"feature":"slugify","requirements":{"R1":{"code":["src/slug.py:slugify"],"test":["tests/test_slug.py:test_slugify"],"status":"ok"}}}`
	writeFile(t, filepath.Join(dir, "specforge/features/slugify/trace.json"), trace)
	writeFile(t, filepath.Join(dir, "specforge/archive/2026-06-16-slugify/trace.json"), trace)
	return dir
}

func TestLoadTracesDedupesArchivedCopy(t *testing.T) {
	dir := makeArchivedFeature(t)
	traces := loadTraces(filepath.Join(dir, "specforge"))

	if len(traces) != 1 {
		// %+v imprime el struct con nombres de campo — útil para ver QUÉ vino.
		t.Fatalf("loadTraces devolvió %d entradas, want 1: %+v", len(traces), traces)
	}
	if traces[0].feature != "slugify" {
		t.Errorf("feature=%q, want %q", traces[0].feature, "slugify")
	}
	// La entrada que sobrevive tiene que ser la VIVA (features/), no la foto
	// archivada: es la que refleja el estado actual del proyecto.
	if !strings.Contains(filepath.ToSlash(traces[0].path), "/features/slugify/") {
		t.Errorf("path=%q, want la copia viva bajo features/", traces[0].path)
	}
}

func TestLoadTracesStripsArchiveDatePrefix(t *testing.T) {
	cases := []struct {
		name string // nombre del directorio bajo archive/
		want string // nombre de feature esperado
	}{
		{"2026-06-16-slugify", "slugify"},
		// Se quita UN solo prefijo de fecha. Una feature que de verdad se llame
		// "2026-06-16-slugify" y se archive el 2026-08-04 queda como
		// "2026-08-04-2026-06-16-slugify" y tiene que resolver a su nombre real.
		{"2026-08-04-2026-06-16-slugify", "2026-06-16-slugify"},
		// Sin prefijo de fecha (carpeta legada) → el nombre queda intacto.
		{"slugify", "slugify"},
		// Prefijo con forma parecida pero inválido → no se toca.
		{"2026-6-16-slugify", "2026-6-16-slugify"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFile(t, filepath.Join(dir, "specforge/archive", c.name, "trace.json"),
				`{"feature":"x","requirements":{}}`)
			traces := loadTraces(filepath.Join(dir, "specforge"))
			if len(traces) != 1 {
				t.Fatalf("len=%d, want 1", len(traces))
			}
			if traces[0].feature != c.want {
				t.Errorf("feature=%q, want %q", traces[0].feature, c.want)
			}
		})
	}
}

// TestDriftReportsArchivedAnchorOnce es la regresión del defecto tal como se
// observó con el binario: UN ancla borrada producía DOS líneas DRIFT.
func TestDriftReportsArchivedAnchorOnce(t *testing.T) {
	dir := makeArchivedFeature(t)
	if err := os.Remove(filepath.Join(dir, "src/slug.py")); err != nil {
		t.Fatalf("remove: %v", err)
	}

	// Recomponemos lo que hace runDrift: por cada trace, juntar sus drifts.
	var drifts []string
	for _, tr := range loadTraces(filepath.Join(dir, "specforge")) {
		drifts = append(drifts, checkFeatureDrift(dir, tr.feature, tr.path, "")...)
	}
	if len(drifts) != 1 {
		t.Errorf("drifts=%d, want 1 (un ancla, un reporte): %v", len(drifts), drifts)
	}
}
