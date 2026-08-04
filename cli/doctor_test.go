package main

import (
	"os"
	"path/filepath"
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
		if code := runDrift(dir, "", true); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})

	t.Run("symbol removed: drift detected", func(t *testing.T) {
		dir := makeDriftProject(t)
		writeFile(t, filepath.Join(dir, "src/slug.py"), "def something_else():\n    return 1\n")
		if code := runDrift(dir, "", true); code != 1 {
			t.Errorf("exit=%d, want 1", code)
		}
	})

	t.Run("file removed: drift detected", func(t *testing.T) {
		dir := makeDriftProject(t)
		if err := os.Remove(filepath.Join(dir, "src/slug.py")); err != nil {
			t.Fatalf("remove: %v", err)
		}
		if code := runDrift(dir, "", true); code != 1 {
			t.Errorf("exit=%d, want 1", code)
		}
	})

	t.Run("non-SpecForge dir: no-op pass", func(t *testing.T) {
		if code := runDrift(t.TempDir(), "", true); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
}
