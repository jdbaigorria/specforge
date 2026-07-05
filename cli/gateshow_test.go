package main

import (
	"path/filepath"
	"testing"
)

// TestBuildGateEvidence: la evidencia junta trace + check + veredictos y su
// frescura, como DATOS (el render es aparte). Es lo que el humano del gate ve
// en lugar del artefacto entero.
func TestBuildGateEvidence(t *testing.T) {
	proj := t.TempDir()

	// Código + test reales para que los anchors resuelvan.
	mustWrite(t, filepath.Join(proj, "src", "a.py"), "def fn():\n    pass\n")
	mustWrite(t, filepath.Join(proj, "tests", "test_a.py"), "def test_fn():\n    pass\n")

	// Trace con un requirement sano y uno sin test.
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "trace.json"),
		`{"feature":"f","requirements":{
			"R1":{"code":["src/a.py:fn"],"test":["tests/test_a.py::test_fn"],"status":"ok"},
			"R2":{"code":["src/a.py:fn"],"test":[],"status":"no-test"}}}`)

	// Un veredicto del juez con una citation inventada (A8 la marca).
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "audit.json"),
		`{"feature":"f","entries":[{"phase":"design","at":"2026-07-01T00:00:00Z","overall":"fail",
			"verdicts":[{"rule":"P1","result":"fail","citation":"x","citation_check":"not-found"}]}]}`)

	// Un check result sellado con el codeHash ACTUAL → fresco.
	res := checkResult{Feature: "f", At: "2026-07-01T00:00:00Z", Passed: true,
		CodeHash: codeHash(proj), Tests: map[string]string{"test_fn": "pass"}}
	if code := writeCheckResult(proj, "f", res); code != 0 {
		t.Fatal("setup: check result")
	}

	f := &feature{Name: "f", Status: "checking", Gates: []gate{
		{Phase: "requirements", Result: "approve", At: "2026-06-30T00:00:00Z", By: "user"},
	}}
	ev := buildGateEvidence(proj, f, "design")

	if ev.TraceTotal != 2 || ev.TraceOK != 1 || len(ev.TraceIssues) != 1 {
		t.Errorf("trace: total=%d ok=%d issues=%v — esperaba 2/1/1", ev.TraceTotal, ev.TraceOK, ev.TraceIssues)
	}
	if ev.Check == nil || !ev.CheckFresh {
		t.Errorf("check debería estar presente y FRESCO, got %+v fresh=%v", ev.Check, ev.CheckFresh)
	}
	if len(ev.Verdicts) != 1 || ev.Verdicts[0].Overall != "fail" {
		t.Errorf("debería traer el veredicto de design, got %+v", ev.Verdicts)
	}
	if ev.LastGate == nil || ev.LastGate.Phase != "requirements" {
		t.Errorf("last gate debería ser requirements, got %+v", ev.LastGate)
	}

	// Tocar el código → la evidencia deja de ser fresca (mismo criterio que el verdict).
	mustWrite(t, filepath.Join(proj, "src", "a.py"), "def fn():\n    return 1\n")
	if ev := buildGateEvidence(proj, f, ""); ev.CheckFresh {
		t.Error("cambió el código → CheckFresh debe ser false")
	}
}

// TestRunGateShow: el comando levanta la feature del estado y sale 0; feature
// inexistente → 4; sin --feature → 2.
func TestRunGateShow(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, featureFilePath(proj, "f"),
		`{"schema_version":"1.0","name":"f","status":"building","gates":[]}`)

	if code := runGateShow([]string{"--feature=f", proj}); code != 0 {
		t.Errorf("show → exit 0, got %d", code)
	}
	if code := runGateShow([]string{"--feature=f", "--json", proj}); code != 0 {
		t.Errorf("show --json → exit 0, got %d", code)
	}
	if code := runGateShow([]string{"--feature=nope", proj}); code != 4 {
		t.Errorf("feature inexistente → exit 4, got %d", code)
	}
	if code := runGateShow([]string{proj}); code != 2 {
		t.Errorf("sin --feature → exit 2, got %d", code)
	}
}
