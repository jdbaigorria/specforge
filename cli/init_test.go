package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestInitMinimal (D1'): scaffold válido — constitución de 3 principios que
// PASA la validación real (va por finishSave), stack detectado, estructura.
func TestInitMinimal(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "go.mod"), "module demo\n")

	if code := runInit([]string{"--minimal", proj}); code != 0 {
		t.Fatalf("init --minimal exit=%d", code)
	}
	for _, p := range []string{
		"specforge/constitution.json", "specforge/constitution.md",
		"specforge/history.md", "specforge/drafts", "specforge/features", "specforge/context",
	} {
		if _, err := os.Stat(filepath.Join(proj, p)); err != nil {
			t.Errorf("%s no existe: %v", p, err)
		}
	}

	// La constitución generada valida y trae el stack Go detectado.
	cf, code := readConstitutionFile(proj)
	if code != 0 {
		t.Fatal("la constitución generada debe ser leíble")
	}
	if len(cf.Principles) != 3 {
		t.Errorf("3 principios default, got %d", len(cf.Principles))
	}
	if cf.Build == nil || cf.Build.TestCmd != "go test -json ./..." || cf.Build.Report != "go-json" {
		t.Errorf("stack go detectado mal: %+v", cf.Build)
	}

	// Re-init sobre un proyecto ya inicializado → rechazado.
	if code := runInit([]string{"--minimal", proj}); code != 4 {
		t.Errorf("re-init → exit 4, got %d", code)
	}
	// Sin --minimal → uso (la conversación completa es de la skill).
	if code := runInit([]string{t.TempDir()}); code != 2 {
		t.Errorf("sin --minimal → exit 2, got %d", code)
	}
}

// TestDetectStack: la detección por archivos-señal del ecosistema.
func TestDetectStack(t *testing.T) {
	proj := t.TempDir()
	if _, _, stack := detectStack(proj); stack != "" {
		t.Errorf("dir vacío → sin stack, got %q", stack)
	}
	mustWrite(t, filepath.Join(proj, "pyproject.toml"), "[project]\n")
	cmd, report, stack := detectStack(proj)
	if stack != "python" || report != "junit" || cmd == "" {
		t.Errorf("python: %q %q %q", cmd, report, stack)
	}
}
