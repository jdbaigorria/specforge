package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestScanInventory: el inventario es determinista — símbolos top-level por
// archivo y mapa test→módulo por convención de nombres.
func TestScanInventory(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "src", "auth.py"),
		"class AuthService:\n    pass\n\ndef login(user):\n    pass\n")
	mustWrite(t, filepath.Join(proj, "tests", "test_auth.py"),
		"def test_login():\n    pass\n")
	mustWrite(t, filepath.Join(proj, "README.md"), "# no es código\n")

	inv := scanInventory(proj)
	if inv.Files != 2 {
		t.Fatalf("esperaba 2 archivos de código, got %d", inv.Files)
	}
	if inv.ByLang[".py"] != 2 {
		t.Errorf("by_lang: %v", inv.ByLang)
	}

	// Símbolos de src/auth.py.
	var auth *moduleInfo
	for i := range inv.Modules {
		if inv.Modules[i].Path == "src/auth.py" {
			auth = &inv.Modules[i]
		}
	}
	if auth == nil {
		t.Fatal("falta src/auth.py en el inventario")
	}
	if len(auth.Symbols) != 2 { // AuthService + login
		t.Errorf("símbolos de auth.py: %v", auth.Symbols)
	}
	if auth.IsTest {
		t.Error("src/auth.py no es test")
	}

	// El test se mapea al módulo por convención (test_auth → auth).
	if mods := inv.TestMap["tests/test_auth.py"]; len(mods) != 1 || mods[0] != "src/auth.py" {
		t.Errorf("test map: %v", inv.TestMap)
	}

	// Dos corridas → mismo inventario (salvo At).
	inv2 := scanInventory(proj)
	inv.At, inv2.At = "", ""
	if len(inv.Modules) != len(inv2.Modules) || inv.Files != inv2.Files {
		t.Error("el inventario debe ser determinista")
	}
}

// TestRunOnboardScanWrites: `sf onboard scan` persiste inventory.json + .md.
func TestRunOnboardScanWrites(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "main.go"), "package main\n\nfunc main() {}\n")
	if code := runOnboard([]string{"scan", proj}); code != 0 {
		t.Fatalf("scan exit=%d", code)
	}
	for _, p := range []string{"inventory.json", "inventory.md"} {
		if _, err := os.Stat(filepath.Join(proj, "specforge", "context", p)); err != nil {
			t.Errorf("%s no escrito: %v", p, err)
		}
	}
}

// TestCoverageBadgeAndHistory (F3): --badge escribe el SVG auto-contenido y
// cada corrida deja un punto en la historia; --history no explota vacío ni lleno.
func TestCoverageBadgeAndHistory(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "trace.json"),
		`{"feature":"f","requirements":{"R1":{"code":["src/a.py:fn"],"test":["t"],"status":"ok"}}}`)
	mustWrite(t, filepath.Join(proj, "src", "a.py"), "def fn():\n    pass\n")

	// Historia vacía → mensaje amable, exit 0.
	if code := runCoverage([]string{"--history", proj}); code != 0 {
		t.Errorf("--history sin datos → exit 0, got %d", code)
	}

	if code := runCoverage([]string{"--badge", proj}); code != 0 {
		t.Fatalf("--badge exit=%d", code)
	}
	svg, err := os.ReadFile(filepath.Join(proj, "specforge", "coverage-badge.svg"))
	if err != nil {
		t.Fatal("el badge debería escribirse")
	}
	if !strings.Contains(string(svg), "spec coverage") || !strings.Contains(string(svg), "100.0%") {
		t.Errorf("badge inesperado: %s", svg)
	}

	// La corrida dejó un punto en la historia.
	if _, err := os.Stat(coverageHistoryPath(proj)); err != nil {
		t.Error("cada corrida debe registrar un punto en coverage-history.jsonl")
	}
	if code := runCoverage([]string{"--history", proj}); code != 0 {
		t.Errorf("--history con datos → exit 0, got %d", code)
	}
}

// TestCoverageRatchet: la cobertura se mide contra los traces y el ratchet
// solo permite subir.
func TestCoverageRatchet(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "trace.json"),
		`{"feature":"f","requirements":{"R1":{"code":["src/a.py:fn"],"test":["tests/test_a.py::test_fn"],"status":"ok"}}}`)
	mustWrite(t, filepath.Join(proj, "src", "a.py"), "def fn():\n    pass\n")
	mustWrite(t, filepath.Join(proj, "src", "b.py"), "def other():\n    pass\n")
	mustWrite(t, filepath.Join(proj, "tests", "test_a.py"), "def test_fn():\n    pass\n")

	// 1 de 2 archivos de producto anclados (los tests no cuentan en el total).
	anchored, total := computeSpecCoverage(proj)
	if anchored != 1 || total != 2 {
		t.Fatalf("coverage: %d/%d, esperaba 1/2", anchored, total)
	}

	// Primera corrida fija la marca (50%).
	if code := runCoverage([]string{proj}); code != 0 {
		t.Fatalf("primera corrida exit=%d", code)
	}
	// Quitar el trace → cae a 0% → ratchet violado (exit 5).
	if err := os.Remove(filepath.Join(proj, "specforge", "features", "f", "trace.json")); err != nil {
		t.Fatal(err)
	}
	if code := runCoverage([]string{proj}); code != 5 {
		t.Errorf("cobertura en caída → exit 5 (ratchet), got %d", code)
	}
}
