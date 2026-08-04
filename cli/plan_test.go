package main

import (
	"path/filepath"
	"testing"
)

// TestAutoSealPlanGate (D2'): con tasks aprobado, `sf plan compute` auto-sella
// el gate de plan (el plan es un cálculo, no un juicio). Sin tasks aprobado, o
// con el plan ya gateado, no hace nada.
func TestAutoSealPlanGate(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "tasks.json"),
		`{"schema_version":"1.0","feature":"f","tasks":[
			{"id":"T1","title":"a","requirement_refs":["R1"],"depends_on":[]},
			{"id":"T2","title":"b","requirement_refs":["R1"],"depends_on":["T1"]}]}`)
	mustWrite(t, featureFilePath(proj, "f"),
		`{"schema_version":"1.0","name":"f","status":"approved","gates":[]}`)

	// Sin gate de tasks todavía → compute NO auto-sella.
	if code := computePlan(proj, "f"); code != 0 {
		t.Fatalf("compute exit=%d", code)
	}
	ff, _ := readFeaturesFile(proj)
	if latestApproveGate(findFeature(&ff, "f"), "plan") != nil {
		t.Fatal("sin tasks aprobado no debe auto-sellarse el plan")
	}

	// Aprobamos tasks (por el camino real: sella hash + cadena) y recomputamos.
	if code := gateApprove(proj, "f", "tasks", "user", ""); code != 0 {
		t.Fatalf("tasks approve exit=%d", code)
	}
	if code := computePlan(proj, "f"); code != 0 {
		t.Fatalf("re-compute exit=%d", code)
	}
	ff, _ = readFeaturesFile(proj)
	g := latestApproveGate(findFeature(&ff, "f"), "plan")
	if g == nil {
		t.Fatal("con tasks aprobado, compute debe auto-sellar el gate de plan")
	}
	if g.By != "sf (derived)" || g.Hash == "" {
		t.Errorf("el auto-sello debe declarar su origen y sellar hash, got %+v", g)
	}

	// Idempotente: otro compute no duplica el gate.
	if code := computePlan(proj, "f"); code != 0 {
		t.Fatalf("3er compute exit=%d", code)
	}
	ff, _ = readFeaturesFile(proj)
	count := 0
	for _, g := range findFeature(&ff, "f").Gates {
		if g.Phase == "plan" {
			count++
		}
	}
	if count != 1 {
		t.Errorf("el plan gate no debe duplicarse, hay %d", count)
	}
}
