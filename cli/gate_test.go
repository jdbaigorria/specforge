package main

import (
	"path/filepath"
	"testing"
)

// TestRunGateStatus verifica los exit codes: 0 cuando hay datos o cuando no hay
// features.json (ausencia no es error), y 4 cuando se pide una feature que no
// existe. (writeFile vive en doctor_test.go, mismo paquete.)
func TestRunGateStatus(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"slugify","status":"done","gates":[{"phase":"requirements","result":"approve","by":"user"}]}]}`)

	t.Run("known feature -> 0", func(t *testing.T) {
		if code := runGateStatus(dir, "slugify"); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
	t.Run("unknown feature -> 4", func(t *testing.T) {
		if code := runGateStatus(dir, "nope"); code != 4 {
			t.Errorf("exit=%d, want 4", code)
		}
	})
	t.Run("summary (no feature) -> 0", func(t *testing.T) {
		if code := runGateStatus(dir, ""); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
	t.Run("no features.json -> 0", func(t *testing.T) {
		if code := runGateStatus(t.TempDir(), ""); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
}
