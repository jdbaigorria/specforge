package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestReadFeaturesMerge: la vista global se ensambla desde el legacy + los
// feature.json por feature; el archivo por feature PISA al legacy por nombre,
// y las features nuevas (solo por feature) se agregan en orden de seq.
func TestReadFeaturesMerge(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"),
		`{"schema_version":"1.0","features":[
			{"name":"a","status":"planned","gates":[]},
			{"name":"b","status":"planned","gates":[]}]}`)
	// Override de "b" (más nueva por feature) + una "c" que el legacy no conoce.
	mustWrite(t, featureFilePath(proj, "b"),
		`{"schema_version":"1.0","name":"b","status":"building","gates":[]}`)
	mustWrite(t, featureFilePath(proj, "c"),
		`{"schema_version":"1.0","name":"c","status":"planned","seq":3,"gates":[]}`)

	ff, err := readFeaturesFile(proj)
	if err != nil {
		t.Fatal(err)
	}
	if len(ff.Features) != 3 {
		t.Fatalf("esperaba 3 features (a, b, c), got %d", len(ff.Features))
	}
	// Orden: el del legacy primero (a, b), las nuevas después (c).
	if ff.Features[0].Name != "a" || ff.Features[1].Name != "b" || ff.Features[2].Name != "c" {
		t.Errorf("orden inesperado: %s, %s, %s", ff.Features[0].Name, ff.Features[1].Name, ff.Features[2].Name)
	}
	// El feature.json por feature gana sobre el legacy.
	if ff.Features[1].Status != "building" {
		t.Errorf("b debería venir del feature.json (building), got %q", ff.Features[1].Status)
	}
}

// TestReadFeaturesPerFeatureOnly: sin legacy, la vista sale solo de los
// feature.json, ordenada por seq (los sin seq al final, por nombre).
func TestReadFeaturesPerFeatureOnly(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, featureFilePath(proj, "zeta"),
		`{"name":"zeta","status":"planned","seq":1,"gates":[]}`)
	mustWrite(t, featureFilePath(proj, "alfa"),
		`{"name":"alfa","status":"planned","seq":2,"gates":[]}`)

	ff, err := readFeaturesFile(proj)
	if err != nil {
		t.Fatal(err)
	}
	if len(ff.Features) != 2 || ff.Features[0].Name != "zeta" || ff.Features[1].Name != "alfa" {
		t.Errorf("el orden debe salir de seq (zeta=1, alfa=2), got %+v", ff.Features)
	}

	// Sin NINGUNA fuente → error (mismo contrato que "no hay proyecto").
	if _, err := readFeaturesFile(t.TempDir()); err == nil {
		t.Error("sin estado → error")
	}
}

// TestWriteFeatureState: el escritor único escribe el feature.json canónico
// (con schema_version) y readFeaturesFile lo levanta.
func TestWriteFeatureState(t *testing.T) {
	proj := t.TempDir()
	if code := writeFeatureState(proj, feature{Name: "f", Status: "planned", Seq: 1}); code != 0 {
		t.Fatalf("write exit=%d", code)
	}
	ff, err := readFeaturesFile(proj)
	if err != nil {
		t.Fatal(err)
	}
	if len(ff.Features) != 1 || ff.Features[0].SchemaVersion != schemaVersionCurrent {
		t.Errorf("feature.json debe llevar schema_version estampado, got %+v", ff.Features)
	}
}

// TestFeatureAddWritesPerFeature (A1): `sf feature add` ya NO crea ningún
// features.json global — el estado nace por feature, con seq de creación.
func TestFeatureAddWritesPerFeature(t *testing.T) {
	proj := t.TempDir()
	if code := runFeatureAdd([]string{"--feature=uno", proj}); code != 0 {
		t.Fatal("add uno falló")
	}
	if code := runFeatureAdd([]string{"--feature=dos", proj}); code != 0 {
		t.Fatal("add dos falló")
	}
	if fileExists(filepath.Join(proj, "specforge", "features.json")) {
		t.Error("feature add no debería crear el features.json global")
	}
	ff, err := readFeaturesFile(proj)
	if err != nil {
		t.Fatal(err)
	}
	if len(ff.Features) != 2 || ff.Features[0].Name != "uno" || ff.Features[1].Name != "dos" {
		t.Errorf("orden de creación esperado (uno, dos), got %+v", ff.Features)
	}
	if ff.Features[0].Seq != 1 || ff.Features[1].Seq != 2 {
		t.Errorf("seq esperado 1,2 — got %d,%d", ff.Features[0].Seq, ff.Features[1].Seq)
	}
}

// TestMigrateSplitsLegacy: `sf migrate` parte el features.json legacy en
// feature.json por feature (numerados) y retira el legacy. Idempotente.
func TestMigrateSplitsLegacy(t *testing.T) {
	proj := t.TempDir()
	legacy := filepath.Join(proj, "specforge", "features.json")
	mustWrite(t, legacy,
		`{"schema_version":"1.0","features":[
			{"name":"a","status":"done","gates":[]},
			{"name":"b","status":"planned","gates":[]}]}`)

	if code := runMigrate([]string{proj}); code != 0 {
		t.Fatalf("migrate exit=%d", code)
	}
	if fileExists(legacy) {
		t.Error("el legacy features.json debería retirarse tras el split")
	}
	for _, name := range []string{"a", "b"} {
		if !fileExists(featureFilePath(proj, name)) {
			t.Errorf("falta features/%s/feature.json", name)
		}
	}
	ff, err := readFeaturesFile(proj)
	if err != nil {
		t.Fatal(err)
	}
	// El orden histórico del array quedó explícito en seq.
	if ff.Features[0].Name != "a" || ff.Features[0].Seq != 1 ||
		ff.Features[1].Name != "b" || ff.Features[1].Seq != 2 {
		t.Errorf("split debe preservar el orden vía seq, got %+v", ff.Features)
	}

	// Segunda corrida: nada que migrar.
	if code := runMigrate([]string{proj}); code != 0 {
		t.Errorf("migrate idempotente → exit 0, got %d", code)
	}
}

// TestHookProtectsFeatureJSON (A1): el estado por feature hereda la Capa 1 —
// feature.json no se escribe a mano.
func TestHookProtectsFeatureJSON(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, featureFilePath(proj, "x"), `{"name":"x","status":"planned","gates":[]}`)
	dec, _ := decidePreToolUse(proj, filepath.Join(proj, "specforge", "features", "x", "feature.json"))
	if dec != "deny" {
		t.Error("feature.json debería estar protegido (estado autoritativo)")
	}
}

// TestMigrateDryRunKeepsLegacy: --dry-run anuncia el split pero no toca nada.
func TestMigrateDryRunKeepsLegacy(t *testing.T) {
	proj := t.TempDir()
	legacy := filepath.Join(proj, "specforge", "features.json")
	mustWrite(t, legacy, `{"schema_version":"1.0","features":[{"name":"a","status":"planned","gates":[]}]}`)
	if code := runMigrate([]string{"--dry-run", proj}); code != 0 {
		t.Fatalf("dry-run exit=%d", code)
	}
	if !fileExists(legacy) {
		t.Error("--dry-run no debe retirar el legacy")
	}
	if fileExists(featureFilePath(proj, "a")) {
		t.Error("--dry-run no debe escribir feature.json")
	}
	_ = os.Remove(legacy)
}
