package main

import (
	"path/filepath"
	"testing"
)

// TestFeatureAdd: crea en planned, detecta duplicados, valida lane.
func TestFeatureAdd(t *testing.T) {
	proj := t.TempDir()
	if code := runFeatureAdd([]string{"--feature=x", "--lane=lite", proj}); code != 0 {
		t.Fatalf("add → exit 0, got %d", code)
	}
	ff, _ := readFeaturesFile(proj)
	f := findFeature(&ff, "x")
	if f == nil || f.Status != "planned" || f.Lane != "lite" {
		t.Errorf("feature mal creada: %+v", f)
	}
	if code := runFeatureAdd([]string{"--feature=x", proj}); code != 4 {
		t.Errorf("duplicado → exit 4, got %d", code)
	}
	if code := runFeatureAdd([]string{"--feature=z", "--lane=huge", proj}); code != 2 {
		t.Errorf("lane inválida → exit 2, got %d", code)
	}
}

// TestFeatureSetStatusTransitions: transiciones válidas e ilegales + el atajo a
// done bloqueado + el flujo serial.
func TestFeatureSetStatusTransitions(t *testing.T) {
	proj := t.TempDir()
	_ = runFeatureAdd([]string{"--feature=x", proj})

	// `done` siempre enruta a `sf feature archive` (exit 2), incluso desde planned.
	if code := runFeatureSetStatus([]string{"--feature=x", "--to=done", proj}); code != 2 {
		t.Errorf("set-status done → exit 2 (use archive), got %d", code)
	}
	// Una transición de salto real (planned→checking) sí es ilegal (exit 5).
	if code := runFeatureSetStatus([]string{"--feature=x", "--to=checking", proj}); code != 5 {
		t.Errorf("planned→checking ilegal → exit 5, got %d", code)
	}
	// planned → approved → building OK.
	if code := runFeatureSetStatus([]string{"--feature=x", "--to=approved", proj}); code != 0 {
		t.Errorf("planned→approved → exit 0, got %d", code)
	}
	if code := runFeatureSetStatus([]string{"--feature=x", "--to=building", proj}); code != 0 {
		t.Errorf("approved→building → exit 0, got %d", code)
	}
	// building → done directo está bloqueado (use archive).
	if code := runFeatureSetStatus([]string{"--feature=x", "--to=done", proj}); code != 2 {
		t.Errorf("set-status done → exit 2 (use archive), got %d", code)
	}
	// status desconocido → exit 2.
	if code := runFeatureSetStatus([]string{"--feature=x", "--to=bogus", proj}); code != 2 {
		t.Errorf("status inválido → exit 2, got %d", code)
	}

	// Flujo serial: con x building, activar y → rechazado.
	_ = runFeatureAdd([]string{"--feature=y", proj})
	if code := runFeatureSetStatus([]string{"--feature=y", "--to=approved", proj}); code != 5 {
		t.Errorf("serial: 2da feature activa → exit 5, got %d", code)
	}
}

// TestFeatureSetLaneLock: el carril se bloquea una vez en building.
func TestFeatureSetLaneLock(t *testing.T) {
	proj := t.TempDir()
	_ = runFeatureAdd([]string{"--feature=x", proj})
	if code := runFeatureSetLane([]string{"--feature=x", "--to=standard", proj}); code != 0 {
		t.Errorf("set-lane en planned → exit 0, got %d", code)
	}
	_ = runFeatureSetStatus([]string{"--feature=x", "--to=approved", proj})
	_ = runFeatureSetStatus([]string{"--feature=x", "--to=building", proj})
	if code := runFeatureSetLane([]string{"--feature=x", "--to=lite", proj}); code != 5 {
		t.Errorf("set-lane en building → exit 5 (locked), got %d", code)
	}
}

// TestFeatureArchive: rechaza sin verdict; con verdict aprobado, copia + done.
func TestFeatureArchive(t *testing.T) {
	proj := t.TempDir()
	_ = runFeatureAdd([]string{"--feature=x", proj})
	// Un archivo dentro de la carpeta de la feature, para verificar la copia.
	mustWrite(t, filepath.Join(proj, "specforge", "features", "x", "requirements.md"), "# reqs\n")

	// Sin verdict → rechazado (exit 5).
	if code := runFeatureArchive([]string{"--feature=x", proj}); code != 5 {
		t.Errorf("archive sin verdict → exit 5, got %d", code)
	}

	// Sembramos un verdict aprobado a mano en el ledger (acá el test ES el escritor
	// legítimo; en prod lo pone `sf gate approve --phase=verdict`).
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"),
		`{"schema_version":"1.0","features":[{"name":"x","status":"checking","gates":[{"phase":"verdict","result":"approve","by":"user","at":"2026-06-30T00:00:00Z"}]}]}`)
	if code := runFeatureArchive([]string{"--feature=x", proj}); code != 0 {
		t.Errorf("archive con verdict → exit 0, got %d", code)
	}
	ff, _ := readFeaturesFile(proj)
	if f := findFeature(&ff, "x"); f == nil || f.Status != "done" {
		t.Errorf("feature debería quedar done, got %+v", f)
	}
	// La carpeta archivada debe existir con el contenido copiado.
	if !fileExists(filepath.Join(proj, "specforge", "archive")) {
		t.Error("debería existir specforge/archive/")
	}
}
