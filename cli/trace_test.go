package main

import (
	"path/filepath"
	"testing"
)

// TestRunTraceVerify reutiliza makeDriftProject (doctor_test.go), que crea un
// trace.json apuntando a src/slug.py:slugify.
func TestRunTraceVerify(t *testing.T) {
	t.Run("clean feature -> 0", func(t *testing.T) {
		dir := makeDriftProject(t)
		if code := runTraceVerify(dir, "slugify"); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
	t.Run("symbol removed -> 1", func(t *testing.T) {
		dir := makeDriftProject(t)
		writeFile(t, filepath.Join(dir, "src/slug.py"), "def x():\n    return 1\n")
		if code := runTraceVerify(dir, "slugify"); code != 1 {
			t.Errorf("exit=%d, want 1", code)
		}
	})
	t.Run("unknown feature -> 4", func(t *testing.T) {
		dir := makeDriftProject(t)
		if code := runTraceVerify(dir, "nope"); code != 4 {
			t.Errorf("exit=%d, want 4", code)
		}
	})
	t.Run("all features clean -> 0", func(t *testing.T) {
		dir := makeDriftProject(t)
		if code := runTraceVerify(dir, ""); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
}
