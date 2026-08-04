package main

import (
	"path/filepath"
	"testing"
)

// runProject arma un proyecto listo para `sf run`: constitución con agent_cmd
// fake (lee el seed y sale 0), tasks de UNA wave, plan computado + gates hasta
// plan, y un trace ya anclado a código/test reales (simula que "el agente ya
// hizo el trabajo" — acá probamos el ORQUESTADOR, no al agente).
func runProject(t *testing.T) string {
	t.Helper()
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "constitution.json"),
		`{"schema_version":"1.0","identity_md":"x","build":{"mode":"per-wave",`+
			`"test_cmd":"true","agent_cmd":"cat > /dev/null"}}`)
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "tasks.json"),
		`{"schema_version":"1.0","feature":"f","tasks":[{"id":"T1","title":"a","requirement_refs":["R1"],"depends_on":[]}]}`)
	mustWrite(t, featureFilePath(proj, "f"),
		`{"schema_version":"1.0","name":"f","status":"building","gates":[]}`)
	// Código + test reales para que el contrato resuelva.
	mustWrite(t, filepath.Join(proj, "src", "a.py"), "def fn():\n    pass\n")
	mustWrite(t, filepath.Join(proj, "tests", "test_a.py"), "def test_fn():\n    pass\n")
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "trace.json"),
		`{"feature":"f","requirements":{"R1":{"code":["src/a.py:fn"],"test":["tests/test_a.py::test_fn"],"status":"ok"}}}`)

	// Gates por el camino real: tasks aprobado → compute auto-sella plan (D2').
	for _, phase := range []string{"requirements", "design", "tasks"} {
		mustWriteArtifactFor(t, proj, "f", phase)
		if code := gateApprove(proj, "f", phase, "user", ""); code != 0 {
			t.Fatalf("setup: %s approve", phase)
		}
	}
	if code := computePlan(proj, "f"); code != 0 {
		t.Fatal("setup: plan compute")
	}
	return proj
}

// mustWriteArtifactFor crea un artefacto mínimo para poder sellar su gate
// (gate approve exige que el archivo exista para hashearlo).
func mustWriteArtifactFor(t *testing.T, proj, feature, phase string) {
	t.Helper()
	switch phase {
	case "requirements":
		mustWrite(t, filepath.Join(proj, "specforge", "features", feature, "requirements.json"),
			`{"feature":"f","requirements":[{"id":"R1","statement":"x"}]}`)
	case "design":
		mustWrite(t, filepath.Join(proj, "specforge", "features", feature, "design.json"),
			`{"feature":"f","components":[{"name":"c","purpose":"x"}]}`)
	case "tasks":
		// ya escrito en runProject (tasks.json real, lo necesita plan compute)
	}
}

// TestRunOrchestratesWaves (F1): el loop ejecuta el agente, valida el contrato,
// corre la suite y sella el checkpoint wave-0; al agotarse las waves termina 0.
func TestRunOrchestratesWaves(t *testing.T) {
	proj := runProject(t)

	if code := runRun([]string{"--feature=f", proj}); code != 0 {
		t.Fatalf("sf run → exit 0, got %d", code)
	}
	ff, _ := readFeaturesFile(proj)
	f := findFeature(&ff, "f")
	g := latestApproveGate(f, "wave-0")
	if g == nil {
		t.Fatal("wave-0 debería quedar sellada por el checkpoint")
	}
	if g.By != "sf run" {
		t.Errorf("el sello debe declarar su origen automatizado, by=%q", g.By)
	}
	// El check result quedó sellado por el orquestador.
	if _, ok := readCheckResult(proj, "f"); !ok {
		t.Error("el checkpoint debe dejar un check result registrado")
	}
	// Re-run: sin waves pendientes → exit 0 inmediato (idempotente).
	if code := runRun([]string{"--feature=f", proj}); code != 0 {
		t.Errorf("re-run sin waves pendientes → exit 0, got %d", code)
	}
}

// TestRunGuards: sin agent_cmd → 2; sin plan gate → 5; agente que falla → STOP 1
// sin sellar nada.
func TestRunGuards(t *testing.T) {
	// Sin agent_cmd.
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "constitution.json"),
		`{"schema_version":"1.0","identity_md":"x","build":{"mode":"inline","test_cmd":"true"}}`)
	mustWrite(t, featureFilePath(proj, "f"), `{"name":"f","status":"building","gates":[]}`)
	if code := runRun([]string{"--feature=f", proj}); code != 2 {
		t.Errorf("sin agent_cmd → exit 2, got %d", code)
	}

	// Con agent_cmd pero sin plan sellado → 5 (spawn no autorizado).
	mustWrite(t, filepath.Join(proj, "specforge", "constitution.json"),
		`{"schema_version":"1.0","identity_md":"x","build":{"test_cmd":"true","agent_cmd":"true"}}`)
	if code := runRun([]string{"--feature=f", proj}); code != 5 {
		t.Errorf("sin plan gate → exit 5, got %d", code)
	}

	// Agente que falla → STOP sin sellar el checkpoint.
	full := runProject(t)
	mustWrite(t, filepath.Join(full, "specforge", "constitution.json"),
		`{"schema_version":"1.0","identity_md":"x","build":{"test_cmd":"true","agent_cmd":"exit 7"}}`)
	if code := runRun([]string{"--feature=f", full}); code != 1 {
		t.Errorf("agente en rojo → exit 1 (STOP), got %d", code)
	}
	ff, _ := readFeaturesFile(full)
	if latestApproveGate(findFeature(&ff, "f"), "wave-0") != nil {
		t.Error("un agente fallido NO debe dejar checkpoint sellado")
	}
}

// TestRunDryRun: muestra la próxima wave sin ejecutar ni sellar.
func TestRunDryRun(t *testing.T) {
	proj := runProject(t)
	if code := runRun([]string{"--feature=f", "--dry-run", proj}); code != 0 {
		t.Fatalf("dry-run → exit 0, got %d", code)
	}
	ff, _ := readFeaturesFile(proj)
	if latestApproveGate(findFeature(&ff, "f"), "wave-0") != nil {
		t.Error("dry-run no debe sellar checkpoints")
	}
}
