package main

import "testing"

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
