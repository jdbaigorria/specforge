package main

import (
	"path/filepath"
	"testing"
)

// setupVerdictProject arma un proyecto con un requirement trazado a código y test
// reales, y devuelve el projectDir. El test luego lo degrada para cada caso.
func setupVerdictProject(t *testing.T) string {
	t.Helper()
	proj := t.TempDir()
	// Código y test reales a los que apunta el trace.
	mustWrite(t, filepath.Join(proj, "src", "app.py"), "def encrypt_with_kms():\n    return 1\n")
	mustWrite(t, filepath.Join(proj, "tests", "test_app.py"), "def test_encrypt():\n    assert True\n")
	// trace.json: R1 → código vivo + test vivo.
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "trace.json"),
		`{"feature":"f","requirements":{"R1":{"code":["src/app.py:encrypt_with_kms"],"test":["tests/test_app.py:test_encrypt"],"status":"verified"}}}`)
	// constitution con test_cmd verde.
	mustWrite(t, filepath.Join(proj, "specforge", "constitution.json"),
		`{"schema_version":"1.0","identity_md":"x","build":{"mode":"inline","test_cmd":"true"}}`)
	return proj
}

// TestVerdictPreconditions cubre las 3 condiciones de la Capa 2.
func TestVerdictPreconditions(t *testing.T) {
	// Caso OK: trace limpio + check verde y fresco → sin razones.
	proj := setupVerdictProject(t)
	if code := runCheckRun([]string{"--feature=f", proj}); code != 0 {
		t.Fatalf("setup: check run debería pasar (exit 0), got %d", code)
	}
	if rs := verdictPreconditions(proj, "f"); len(rs) != 0 {
		t.Errorf("proyecto sano → sin razones, got %v", rs)
	}

	// Caso STALE: tocar el código después de correr el test invalida la frescura.
	mustWrite(t, filepath.Join(proj, "src", "app.py"), "def encrypt_with_kms():\n    return 2  # changed\n")
	if rs := verdictPreconditions(proj, "f"); !hasReasonContaining(rs, "STALE") {
		t.Errorf("código cambiado tras el test → razón STALE, got %v", rs)
	}

	// Caso CODE DRIFT: el anchor de código ya no existe (símbolo que no importa).
	proj2 := setupVerdictProject(t)
	_ = runCheckRun([]string{"--feature=f", proj2})
	mustWrite(t, filepath.Join(proj2, "src", "app.py"), "def something_else():\n    return 1\n")
	if rs := verdictPreconditions(proj2, "f"); !hasReasonContaining(rs, "drift") {
		t.Errorf("símbolo de código ausente → razón drift, got %v", rs)
	}

	// Caso SIN CHECK: trace limpio pero nunca se corrió `sf check run`.
	proj3 := setupVerdictProject(t)
	if rs := verdictPreconditions(proj3, "f"); !hasReasonContaining(rs, "no test result") {
		t.Errorf("sin check run → razón 'no test result', got %v", rs)
	}

	// Caso TEST GONE: el test nombrado no resuelve a un test real.
	proj4 := setupVerdictProject(t)
	_ = runCheckRun([]string{"--feature=f", proj4})
	mustWrite(t, filepath.Join(proj4, "specforge", "features", "f", "trace.json"),
		`{"feature":"f","requirements":{"R1":{"code":["src/app.py:encrypt_with_kms"],"test":["tests/test_ghost.py:test_nope"],"status":"verified"}}}`)
	if rs := verdictPreconditions(proj4, "f"); !hasReasonContaining(rs, "does not resolve") {
		t.Errorf("test ref inexistente → razón 'does not resolve', got %v", rs)
	}
}

// TestGateApproveVerdictRefused: el caso real de FIXBUGHIGH — intentar archivar
// una feature cuyo código no importa → gate approve --phase=verdict RECHAZADO.
func TestGateApproveVerdictRefused(t *testing.T) {
	proj := setupVerdictProject(t)
	// review.json (artefacto de la fase verdict) + ledger con la feature.
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "review.json"), `{"feature":"f"}`)
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"),
		`{"schema_version":"1.0","features":[{"name":"f","status":"building","gates":[]}]}`)

	// Sin check run ni trace verde-fresco → verdict rechazado (exit 5).
	if code := gateApprove(proj, "f", "verdict", "user", ""); code != 5 {
		t.Errorf("verdict sin precondiciones → exit 5, got %d", code)
	}

	// Ahora lo dejamos sano: corremos el check → verdict otorgado (exit 0).
	if c := runCheckRun([]string{"--feature=f", proj}); c != 0 {
		t.Fatalf("check run debería pasar, got %d", c)
	}
	if code := gateApprove(proj, "f", "verdict", "user", ""); code != 0 {
		t.Errorf("verdict con precondiciones cumplidas → exit 0, got %d", code)
	}
}

func hasReasonContaining(reasons []string, sub string) bool {
	for _, r := range reasons {
		if contains(r, sub) {
			return true
		}
	}
	return false
}

// contains: substring sin importar el paquete strings en el test (legibilidad).
func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOfStr(s, sub) >= 0)
}

func indexOfStr(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
