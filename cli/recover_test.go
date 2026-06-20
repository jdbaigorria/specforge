package main

import (
	"path/filepath"
	"testing"
)

// ----------------------------------------------------------------------------
// Tests de `sf recover`. buildRecoveryReport cuelga de computeArtifactStates, así
// que reusamos los helpers de stale_test.go (writeArtifact, approvedGate) para
// armar el escenario en un t.TempDir() y verificar el PLAN resultante.
// ----------------------------------------------------------------------------

// planStepFor busca el paso del plan para una fase (helper de aserción).
func planStepFor(r recoveryReport, phase string) (recoveryStep, bool) {
	for _, p := range r.Plan {
		if p.Phase == phase {
			return p, true
		}
	}
	return recoveryStep{}, false
}

// TestRecoverClean: con todo aprobado y en sync, no hay nada que recuperar.
func TestRecoverClean(t *testing.T) {
	dir := t.TempDir()
	featDir := filepath.Join(dir, "specforge", "features", "auth")
	writeArtifact(t, featDir, "requirements.json", `{"r":[1]}`)
	writeArtifact(t, featDir, "design.json", `{"c":[]}`)

	f := &feature{Name: "auth", Gates: []gate{
		approvedGate(t, dir, "auth", "requirements", "2026-01-01T10:00:00Z"),
		approvedGate(t, dir, "auth", "design", "2026-01-01T11:00:00Z"),
	}}
	r := buildRecoveryReport(f, dir)

	if !r.OK || len(r.Plan) != 0 {
		t.Errorf("estado sano: OK=%v plan=%d, want OK=true plan=0", r.OK, len(r.Plan))
	}
}

// TestRecoverReseal: un artefacto editado después de aprobarlo (hash mismatch) se
// arregla RE-SELLANDO (reseal); el downstream que hereda la staleness se
// REGENERA. Distinguir ambos es lo valioso del plan.
func TestRecoverReseal(t *testing.T) {
	dir := t.TempDir()
	featDir := filepath.Join(dir, "specforge", "features", "auth")
	writeArtifact(t, featDir, "requirements.json", `{"r":[1]}`)
	writeArtifact(t, featDir, "design.json", `{"c":[]}`)

	gates := []gate{
		approvedGate(t, dir, "auth", "requirements", "2026-01-01T10:00:00Z"),
		approvedGate(t, dir, "auth", "design", "2026-01-01T11:00:00Z"),
	}
	// Silent edit en requirements (su propio contenido cambió):
	writeArtifact(t, featDir, "requirements.json", `{"r":[1,2]}`)

	r := buildRecoveryReport(&feature{Name: "auth", Gates: gates}, dir)

	if p, ok := planStepFor(r, "requirements"); !ok || p.Action != "reseal" {
		t.Errorf("requirements action=%q ok=%v, want reseal", p.Action, ok)
	}
	if p, ok := planStepFor(r, "design"); !ok || p.Action != "regenerate" {
		t.Errorf("design action=%q ok=%v, want regenerate (heredó staleness)", p.Action, ok)
	}
	// El comando sugerido debe ser el sf gate approve de esa fase.
	if p, _ := planStepFor(r, "requirements"); p.Command != "sf gate approve --feature=auth --phase=requirements" {
		t.Errorf("comando inesperado: %q", p.Command)
	}
}

// TestRecoverMissingRootVsTail: un missing EN EL MEDIO (con contenido aguas
// abajo) es una raíz a arreglar; un missing DE COLA (nada aprobado después) es
// trabajo pendiente, no inconsistencia → no se lista.
func TestRecoverMissingRootVsTail(t *testing.T) {
	dir := t.TempDir()
	featDir := filepath.Join(dir, "specforge", "features", "auth")
	// requirements falta; design existe y está aprobado → requirements es raíz.
	writeArtifact(t, featDir, "design.json", `{"c":[]}`)

	f := &feature{Name: "auth", Gates: []gate{
		approvedGate(t, dir, "auth", "design", "2026-01-01T11:00:00Z"),
	}}
	r := buildRecoveryReport(f, dir)

	if p, ok := planStepFor(r, "requirements"); !ok || p.Action != "generate" {
		t.Errorf("requirements (raíz missing) action=%q ok=%v, want generate", p.Action, ok)
	}
	// tasks y plan faltan pero son COLA (nada después) → no deben listarse.
	if _, ok := planStepFor(r, "tasks"); ok {
		t.Errorf("tasks es missing de cola: no debería estar en el plan")
	}
}
