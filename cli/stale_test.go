package main

import (
	"os"
	"path/filepath"
	"testing"
)

// ----------------------------------------------------------------------------
// Tests del stale model. Como computeArtifactStates lee disco (existencia +
// hash), usamos t.TempDir(): un directorio temporal que Go limpia solo al
// terminar el test. Armamos los artefactos y el ledger a mano y verificamos los
// estados resultantes.
// ----------------------------------------------------------------------------

// writeArtifact es un helper de test: escribe contenido en featDir/rel creando
// los directorios intermedios. Falla el test si no puede. (No reusamos writeFile
// de doctor_test.go porque ese toma un path absoluto, no featDir+rel.)
func writeArtifact(t *testing.T, featDir, rel, content string) {
	t.Helper()
	p := filepath.Join(featDir, rel)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// approvedGate arma un gate aprobado para `phase`, sellando el hash REAL del
// artefacto en disco (como haría `sf gate approve`). Así el baseline parte de
// "approved" y los tests provocan staleness editando después.
func approvedGate(t *testing.T, projectDir, feature, phase, at string) gate {
	t.Helper()
	g := gate{Phase: phase, Result: "approve", By: "user", At: at}
	if rel, has := artifactFileForPhase(phase); has {
		abs := filepath.Join(projectDir, "specforge", "features", feature, rel)
		h, ok := hashArtifact(abs)
		if !ok {
			t.Fatalf("approvedGate: no pude hashear %s", abs)
		}
		g.Hash = h
	}
	return g
}

// stateOf busca el estado de una fase en el resultado (helper de aserción).
func stateOf(states []artifactState, phase string) artifactState {
	for _, s := range states {
		if s.Phase == phase {
			return s
		}
	}
	return artifactState{Phase: phase, State: "<absent>"}
}

// TestArtifactStatesBaseline: con artefactos presentes y gates que sellan sus
// hashes actuales, todo está approved; lo que no existe es missing.
func TestArtifactStatesBaseline(t *testing.T) {
	dir := t.TempDir()
	featDir := filepath.Join(dir, "specforge", "features", "auth")
	writeArtifact(t, featDir, "requirements.json", `{"r":[1]}`)
	writeArtifact(t, featDir, "design.json", `{"c":[]}`)

	f := &feature{Name: "auth", Gates: []gate{
		approvedGate(t, dir, "auth", "requirements", "2026-01-01T10:00:00Z"),
		approvedGate(t, dir, "auth", "design", "2026-01-01T11:00:00Z"),
	}}
	states := computeArtifactStates(f, dir)

	if s := stateOf(states, "requirements"); s.State != "approved" {
		t.Errorf("requirements=%q, want approved (%s)", s.State, s.Reason)
	}
	if s := stateOf(states, "design"); s.State != "approved" {
		t.Errorf("design=%q, want approved (%s)", s.State, s.Reason)
	}
	if s := stateOf(states, "tasks"); s.State != "missing" {
		t.Errorf("tasks=%q, want missing", s.State)
	}
}

// TestArtifactDraftAndBlocked: un artefacto presente sin gate es draft; si
// además su upstream no está sólido, es blocked.
func TestArtifactDraftAndBlocked(t *testing.T) {
	dir := t.TempDir()
	featDir := filepath.Join(dir, "specforge", "features", "auth")
	writeArtifact(t, featDir, "requirements.json", `{"r":[1]}`)
	writeArtifact(t, featDir, "design.json", `{"c":[]}`)

	// Sin ningún gate: requirements existe pero no aprobado → draft;
	// design existe pero su upstream (requirements) no está sólido → blocked.
	f := &feature{Name: "auth"}
	states := computeArtifactStates(f, dir)

	if s := stateOf(states, "requirements"); s.State != "draft" {
		t.Errorf("requirements=%q, want draft", s.State)
	}
	if s := stateOf(states, "design"); s.State != "blocked" {
		t.Errorf("design=%q, want blocked (%s)", s.State, s.Reason)
	}
}

// TestArtifactStaleHashMismatch: editar el artefacto después de aprobarlo (sin
// re-gatear) lo vuelve stale, y la staleness se propaga aguas abajo.
func TestArtifactStaleHashMismatch(t *testing.T) {
	dir := t.TempDir()
	featDir := filepath.Join(dir, "specforge", "features", "auth")
	writeArtifact(t, featDir, "requirements.json", `{"r":[1]}`)
	writeArtifact(t, featDir, "design.json", `{"c":[]}`)

	gates := []gate{
		approvedGate(t, dir, "auth", "requirements", "2026-01-01T10:00:00Z"),
		approvedGate(t, dir, "auth", "design", "2026-01-01T11:00:00Z"),
	}
	// Silent edit DESPUÉS de sellar el hash:
	writeArtifact(t, featDir, "requirements.json", `{"r":[1,2]}`)

	f := &feature{Name: "auth", Gates: gates}
	states := computeArtifactStates(f, dir)

	if s := stateOf(states, "requirements"); s.State != "stale" {
		t.Errorf("requirements=%q, want stale", s.State)
	}
	if s := stateOf(states, "design"); s.State != "stale" {
		t.Errorf("design=%q, want stale (propagación)", s.State)
	}
}

// TestArtifactStaleUpstreamReopened: re-aprobar un upstream DESPUÉS deja stale a
// los downstream aunque su propio contenido no haya cambiado.
func TestArtifactStaleUpstreamReopened(t *testing.T) {
	dir := t.TempDir()
	featDir := filepath.Join(dir, "specforge", "features", "auth")
	writeArtifact(t, featDir, "requirements.json", `{"r":[1]}`)
	writeArtifact(t, featDir, "design.json", `{"c":[]}`)

	f := &feature{Name: "auth", Gates: []gate{
		approvedGate(t, dir, "auth", "design", "2026-01-01T11:00:00Z"),
		// requirements re-aprobado MÁS TARDE que design:
		approvedGate(t, dir, "auth", "requirements", "2026-01-01T12:00:00Z"),
	}}
	states := computeArtifactStates(f, dir)

	if s := stateOf(states, "design"); s.State != "stale" {
		t.Errorf("design=%q, want stale (upstream reabierto)", s.State)
	}
}

// TestCanonicalHashIgnoresWhitespace: un cambio SOLO de formato no debe disparar
// stale — el hash es del contenido canónico, no de los bytes crudos.
func TestCanonicalHashIgnoresWhitespace(t *testing.T) {
	dir := t.TempDir()
	featDir := filepath.Join(dir, "specforge", "features", "auth")
	writeArtifact(t, featDir, "requirements.json", `{"a":1,"b":2}`)

	g := approvedGate(t, dir, "auth", "requirements", "2026-01-01T10:00:00Z")
	// Reescribimos con MISMO contenido pero distinto formato/orden:
	writeArtifact(t, featDir, "requirements.json", "{\n  \"b\": 2,\n  \"a\": 1\n}\n")

	f := &feature{Name: "auth", Gates: []gate{g}}
	if s := stateOf(computeArtifactStates(f, dir), "requirements"); s.State != "approved" {
		t.Errorf("requirements=%q, want approved (cambio solo de formato) — reason=%q", s.State, s.Reason)
	}
}
