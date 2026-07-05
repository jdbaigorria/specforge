package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// Tests de `sf migrate` (migrate.go, R6).
// ----------------------------------------------------------------------------

// legacyProject arma un proyecto PRE-versionado y PRE-cadena: artefacto sin
// schema_version, gate legacy (sin prev) que SELLA ese artefacto.
func legacyProject(t *testing.T) string {
	t.Helper()
	proj := t.TempDir()
	reqPath := filepath.Join(proj, "specforge", "features", "f", "requirements.json")
	mustWrite(t, reqPath, `{"feature":"f","requirements":[{"id":"R1","text":"x","criteria":["c"]}]}`)

	// El gate legacy sella el hash canónico ACTUAL del artefacto (sello válido).
	sealed, ok := hashArtifact(reqPath)
	if !ok {
		t.Fatal("setup: cannot hash artifact")
	}
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"),
		`{"features":[{"name":"f","status":"building","gates":[`+
			`{"phase":"lane","result":"approve","by":"user","at":"2026-01-01T00:00:00Z"},`+
			`{"phase":"requirements","result":"approve","by":"user","at":"2026-01-02T00:00:00Z","hash":"`+sealed+`"}]}]}`)
	return proj
}

func TestMigrateLegacyProject(t *testing.T) {
	proj := legacyProject(t)

	if code := runMigrate([]string{proj}); code != 0 {
		t.Fatalf("migrate failed (exit %d)", code)
	}

	// El artefacto quedó versionado…
	data, _ := os.ReadFile(filepath.Join(proj, "specforge", "features", "f", "requirements.json"))
	if !strings.Contains(string(data), `"schema_version": "`+schemaVersionCurrent+`"`) {
		t.Fatal("schema_version not stamped on the artifact")
	}

	ff, err := readFeaturesFile(proj)
	if err != nil {
		t.Fatalf("features.json unreadable after migrate: %v", err)
	}
	// …features.json también…
	if ff.SchemaVersion != schemaVersionCurrent {
		t.Fatalf("features.json schema_version = %q", ff.SchemaVersion)
	}
	f := &ff.Features[0]
	// …el ledger quedó 100% encadenado y VÁLIDO…
	if probs := ledgerProblems(f); len(probs) != 0 {
		t.Fatalf("migrated ledger must validate, got %v", probs)
	}
	if f.Gates[0].Prev != genesisPrev || f.Gates[1].Prev == "" {
		t.Fatal("prev chain not backfilled")
	}
	// …y el sello del gate fue RE-sellado al hash post-estampado (el artefacto
	// no debe figurar stale por la migración).
	states := computeArtifactStates(f, proj)
	for _, s := range states {
		if s.Phase == "requirements" && s.State != "approved" {
			t.Fatalf("requirements should stay approved after migrate, got %s (%s)", s.State, s.Reason)
		}
	}

	// Idempotencia: una segunda corrida no tiene nada que hacer.
	if code := runMigrate([]string{proj}); code != 0 {
		t.Fatalf("second migrate failed (exit %d)", code)
	}
}

func TestMigrateDryRunWritesNothing(t *testing.T) {
	proj := legacyProject(t)
	before, _ := os.ReadFile(filepath.Join(proj, "specforge", "features.json"))

	if code := runMigrate([]string{"--dry-run", proj}); code != 0 {
		t.Fatalf("dry-run failed (exit %d)", code)
	}
	after, _ := os.ReadFile(filepath.Join(proj, "specforge", "features.json"))
	if string(before) != string(after) {
		t.Fatal("--dry-run must not modify features.json")
	}
}

func TestMigrateRefusesUnknownVersion(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"), `{"schema_version":"1.0","features":[]}`)
	mustWrite(t, filepath.Join(proj, "specforge", "constitution.json"),
		`{"schema_version":"9.9","identity_md":"x"}`)

	if code := runMigrate([]string{proj}); code != 5 {
		t.Fatalf("unknown schema_version must exit 5, got %d", code)
	}
}

func TestMigrateRefusesBrokenLedger(t *testing.T) {
	proj := t.TempDir()
	// Cadena rota: prev que no coincide con nada.
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"),
		`{"schema_version":"1.0","features":[{"name":"f","status":"building","gates":[`+
			`{"phase":"lane","result":"approve","by":"u","at":"2026-01-01T00:00:00Z","prev":"deadbeef"}]}]}`)

	if code := runMigrate([]string{proj}); code != 5 {
		t.Fatalf("broken ledger must refuse migrate (exit 5), got %d", code)
	}
}
