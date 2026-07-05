package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// Tests de `sf verify` (verify.go, R2): el agregador de CI.
// Reusamos setupVerdictProject (gate_verdict_test.go) como proyecto base sano.
// ----------------------------------------------------------------------------

// verifyProject: proyecto sano CON ledger encadenado (gates escritos por
// gateApprove, el camino real).
func verifyProject(t *testing.T) string {
	t.Helper()
	proj := setupVerdictProject(t)
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"),
		`{"schema_version":"1.0","features":[{"name":"f","status":"building","gates":[]}]}`)
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "review.json"), `{"feature":"f"}`)
	return proj
}

func TestVerifyCleanProject(t *testing.T) {
	proj := verifyProject(t)
	rep := buildVerifyReport(proj, "")
	if !rep.OK {
		t.Fatalf("clean project must verify OK, got %+v", rep)
	}
}

func TestVerifyNoSpecforgeIsOK(t *testing.T) {
	// Repo sin specforge/: el workflow instalado no debe romper el CI.
	rep := buildVerifyReport(t.TempDir(), "")
	if !rep.OK {
		t.Fatalf("repo without specforge/ must verify OK, got %+v", rep)
	}
}

func TestVerifyDetectsForgedLedger(t *testing.T) {
	proj := verifyProject(t)
	// Un gate aprobado legítimamente…
	if code := gateApprove(proj, "f", "lane", "user", ""); code != 0 {
		t.Fatalf("setup: lane approve failed (%d)", code)
	}
	// …y un verdict FORJADO a mano (el incidente FIXBUGHIGH, ahora vía archivo):
	// apendeamos la entrada directo al JSON, sin pasar por gateApprove — como
	// haría un editor humano u otro agente sin hooks. Sin `prev` después de un
	// gate encadenado, la cadena lo delata.
	ff, err := readFeaturesFile(proj)
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	ff.Features[0].Gates = append(ff.Features[0].Gates, gate{
		Phase: "verdict", Result: "approve", By: "user", At: "2026-07-05T12:00:00Z",
	})
	if code := writeFeatureState(proj, ff.Features[0]); code != 0 {
		t.Fatalf("setup: write failed (%d)", code)
	}

	rep := buildVerifyReport(proj, "")
	if rep.OK {
		t.Fatal("forged ledger must fail verify")
	}
	if !checkFailed(rep, "ledger") {
		t.Fatalf("expected the ledger check to fail, got %+v", rep)
	}
}

func TestVerifyDetectsInvalidSchema(t *testing.T) {
	proj := verifyProject(t)
	// Un tasks.json que jamás pasaría por `sf save` (JSON roto).
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "tasks.json"), `{not json`)

	rep := buildVerifyReport(proj, "")
	if rep.OK || !checkFailed(rep, "schemas") {
		t.Fatalf("invalid artifact JSON must fail the schemas check, got %+v", rep)
	}
}

func TestVerifyDetectsTraceDrift(t *testing.T) {
	proj := verifyProject(t)
	// El código al que ancla el trace desaparece → drift.
	mustWrite(t, filepath.Join(proj, "src", "app.py"), "def renamed():\n    return 1\n")

	rep := buildVerifyReport(proj, "")
	if rep.OK || !checkFailed(rep, "trace-drift") {
		t.Fatalf("dead trace anchor must fail the trace-drift check, got %+v", rep)
	}
}

func TestVerifyVerdictCheckWithExplicitFeature(t *testing.T) {
	proj := verifyProject(t)
	// Sin `sf check run` el verdict no está listo; con --feature lo forzamos.
	rep := buildVerifyReport(proj, "f")
	if rep.OK || !checkFailed(rep, "verdict") {
		t.Fatalf("verdict check without a test run must fail, got %+v", rep)
	}

	// Corremos el check real → ahora sí.
	if code := runCheckRun([]string{"--feature=f", proj}); code != 0 {
		t.Fatalf("check run should pass, got %d", code)
	}
	if rep := buildVerifyReport(proj, "f"); !rep.OK {
		t.Fatalf("after a fresh green run, verify --feature must pass, got %+v", rep)
	}
}

func TestVerifyInitCI(t *testing.T) {
	proj := t.TempDir()
	if code := verifyInitCI(proj); code != 0 {
		t.Fatalf("init-ci failed (%d)", code)
	}
	path := filepath.Join(proj, ".github", "workflows", "specforge-verify.yml")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("workflow not written: %v", err)
	}
	if !strings.Contains(string(data), "sf verify") {
		t.Fatal("workflow must invoke `sf verify`")
	}
	// Re-scaffold sobre uno existente: rehusado (exit 4), no pisado.
	if code := verifyInitCI(proj); code != 4 {
		t.Fatalf("existing workflow must not be overwritten, got exit %d", code)
	}
}

// checkFailed busca el check por nombre y devuelve true si falló.
func checkFailed(rep verifyReport, name string) bool {
	for _, c := range rep.Checks {
		if c.Name == name {
			return !c.OK
		}
	}
	return false
}
