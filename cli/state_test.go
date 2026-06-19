package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestDerivePhaseFrontier: la fase actual es la SIGUIENTE a la última aprobada
// (frontera + 1). Sin tasks.json (projectDir = dir vacío "."/inexistente),
// totalWaves devuelve -1, así que las fases pre-build se computan puras.
func TestDerivePhaseFrontier(t *testing.T) {
	cases := []struct {
		name      string
		gates     []gate
		wantPhase string
	}{
		{"sin gates → primera fase", nil, "lane"},
		{"lane aprobado → requirements", []gate{{Phase: "lane", Result: "approve"}}, "requirements"},
		{"requirements → design", []gate{
			{Phase: "lane", Result: "approve"},
			{Phase: "requirements", Result: "approve"},
		}, "design"},
		{"reject no mueve la frontera", []gate{
			{Phase: "requirements", Result: "approve"},
			{Phase: "design", Result: "reject"},
		}, "design"},
		{"tasks → plan", []gate{
			{Phase: "requirements", Result: "approve"},
			{Phase: "design", Result: "approve"},
			{Phase: "tasks", Result: "approve"},
		}, "plan"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f := &feature{Name: "x", Gates: c.gates}
			got, _ := derivePhase(f, "/nonexistent-project")
			if got != c.wantPhase {
				t.Errorf("phase=%q, want %q", got, c.wantPhase)
			}
		})
	}
}

// TestDerivePhaseBuild: tras aprobar "plan" entramos a build wave 0; tras una
// wave, vamos a la siguiente. Sin tasks.json no podemos acotar el total, así que
// build "sigue" (degradación segura).
func TestDerivePhaseBuild(t *testing.T) {
	planApproved := &feature{Gates: []gate{
		{Phase: "plan", Result: "approve"},
	}}
	phase, wave := derivePhase(planApproved, "/nonexistent-project")
	if phase != "build" || wave != 0 {
		t.Errorf("tras plan: phase=%q wave=%d, want build/0", phase, wave)
	}

	wave0Approved := &feature{Gates: []gate{
		{Phase: "plan", Result: "approve"},
		{Phase: "wave-0", Result: "approve"},
	}}
	phase, wave = derivePhase(wave0Approved, "/nonexistent-project")
	if phase != "build" || wave != 1 {
		t.Errorf("tras wave-0: phase=%q wave=%d, want build/1", phase, wave)
	}
}

// TestComputeCurrentStateSerial: con UNA feature activa, el estado la refleja;
// con ninguna o varias, devuelve una nota honesta y sin feature.
func TestComputeCurrentStateSerial(t *testing.T) {
	// Una activa (building), con un planned y un done de ruido.
	ff := featuresFile{Features: []feature{
		{Name: "old", Status: "done"},
		{Name: "cur", Status: "building", Gates: []gate{{Phase: "plan", Result: "approve"}}},
		{Name: "later", Status: "planned"},
	}}
	st := computeCurrentState(ff, "/nonexistent-project")
	if st.Feature != "cur" || st.Phase != "build" {
		t.Errorf("got feature=%q phase=%q, want cur/build", st.Feature, st.Phase)
	}
	if st.Wave == nil || *st.Wave != 0 {
		t.Errorf("wave=%v, want 0", st.Wave)
	}

	// Ninguna activa.
	none := featuresFile{Features: []feature{{Name: "old", Status: "done"}}}
	if st := computeCurrentState(none, "."); st.Feature != "" || st.Note == "" {
		t.Errorf("sin activa: got feature=%q note=%q, want vacío + nota", st.Feature, st.Note)
	}

	// Dos activas → violación serial reportada, sin elegir una.
	two := featuresFile{Features: []feature{
		{Name: "a", Status: "approved"},
		{Name: "b", Status: "building"},
	}}
	if st := computeCurrentState(two, "."); st.Feature != "" || st.Note == "" {
		t.Errorf("dos activas: got feature=%q note=%q, want vacío + nota de violación", st.Feature, st.Note)
	}
}

// TestBuildBreadcrumb: con feature menciona feature+fase (+wave si hay); sin
// feature cae a la nota.
func TestBuildBreadcrumb(t *testing.T) {
	w := 1
	withWave := buildBreadcrumb(currentState{Feature: "trunc", Phase: "build", Wave: &w}, ".")
	if !strings.Contains(withWave, "trunc") || !strings.Contains(withWave, "build") || !strings.Contains(withWave, "wave 1") {
		t.Errorf("breadcrumb=%q, want feature+fase+wave", withWave)
	}

	noWave := buildBreadcrumb(currentState{Feature: "trunc", Phase: "design"}, ".")
	if strings.Contains(noWave, "wave") {
		t.Errorf("breadcrumb=%q, no debería mencionar wave en design", noWave)
	}

	none := buildBreadcrumb(currentState{Note: "no active feature"}, ".")
	if !strings.Contains(none, "no active feature") {
		t.Errorf("breadcrumb=%q, want la nota", none)
	}
}

// TestBuildCurrentContextBuild: en fase build, el slice adjunta el wave_slice
// leído del tasks.json del proyecto.
func TestBuildCurrentContextBuild(t *testing.T) {
	dir := t.TempDir()
	featDir := filepath.Join(dir, "specforge", "features", "trunc")
	if err := os.MkdirAll(featDir, 0o755); err != nil {
		t.Fatal(err)
	}
	tasks := `{"feature":"trunc","waves":[
		{"n":0,"name":"Foundation","tasks":[{"id":"T1","status":"pending","requirement_refs":["R1"]}]}
	]}`
	if err := os.WriteFile(filepath.Join(featDir, "tasks.json"), []byte(tasks), 0o644); err != nil {
		t.Fatal(err)
	}

	w := 0
	st := currentState{Feature: "trunc", Status: "building", Phase: "build", Wave: &w}
	cc := buildCurrentContext(st, dir)

	if cc.WaveSlice == nil {
		t.Fatal("WaveSlice nil, want el slice de la wave 0")
	}
	if cc.WaveSlice.Wave.N != 0 {
		t.Errorf("wave_slice.wave.n=%d, want 0", cc.WaveSlice.Wave.N)
	}
	if indexOf(cc.Artifacts, "tasks.json") < 0 { // indexOf vive en status_test.go
		t.Errorf("available_artifacts=%v, want incluir tasks.json", cc.Artifacts)
	}
}
