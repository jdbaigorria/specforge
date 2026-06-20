package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestPrinciplesForPhase: el filtro por applies_to es la inversión del mapeo.
func TestPrinciplesForPhase(t *testing.T) {
	cf := constitutionFile{
		Principles: []principle{
			{ID: "P1", Statement: "offline-first", AppliesTo: []string{"design", "build"}},
			{ID: "P2", Statement: "todo testeable", AppliesTo: []string{"requirements"}},
			{ID: "P3", Statement: "sin applies_to"},
		},
		Invariants: []invariant{
			{ID: "I1", Rule: "error handling", AppliesTo: []string{"design", "build"}},
		},
	}

	ps, invs := principlesForPhase(cf, "design")
	if len(ps) != 1 || ps[0].ID != "P1" {
		t.Errorf("design principles=%v, want [P1]", ps)
	}
	if len(invs) != 1 || invs[0].ID != "I1" {
		t.Errorf("design invariants=%v, want [I1]", invs)
	}

	ps, _ = principlesForPhase(cf, "requirements")
	if len(ps) != 1 || ps[0].ID != "P2" {
		t.Errorf("requirements principles=%v, want [P2]", ps)
	}

	// tasks: ningún principio mapeado.
	if ps, invs := principlesForPhase(cf, "tasks"); len(ps) != 0 || len(invs) != 0 {
		t.Errorf("tasks debería no tener reglas, got %v / %v", ps, invs)
	}
}

// TestPhaseArtifactPath: cada fase de spec mapea a su .json; plan bajo progress/;
// build no tiene artefacto único.
func TestPhaseArtifactPath(t *testing.T) {
	base := filepath.Join("p", "specforge", "features", "f")
	cases := map[string]string{
		"design": filepath.Join(base, "design.json"),
		"plan":   filepath.Join(base, "progress", "plan.json"),
		"build":  "",
		"verdict": "",
	}
	for phase, want := range cases {
		if got := phaseArtifactPath("p", "f", phase); got != want {
			t.Errorf("phase %q → %q, want %q", phase, got, want)
		}
	}
}

// TestCheckAppliesTo: fase desconocida = ERROR (gate estructural); applies_to
// vacío = WARN (lint).
func TestCheckAppliesTo(t *testing.T) {
	var rep report
	cf := constitutionFile{
		IdentityMD: "x",
		Principles: []principle{
			{ID: "P1", Statement: "ok", AppliesTo: []string{"design"}},
			{ID: "P2", Statement: "fase mala", AppliesTo: []string{"nope"}},
			{ID: "P3", Statement: "sin mapeo"},
		},
	}
	checkConstitution(cf, &rep)
	if len(rep.errors) != 1 { // solo P2 (fase inexistente)
		t.Errorf("errors=%v, want 1 (unknown phase)", rep.errors)
	}
	if len(rep.warnings) == 0 { // al menos P3 sin applies_to
		t.Errorf("warnings=%v, want al menos 1 (missing applies_to)", rep.warnings)
	}
}

// TestContextForJudge: integración — escribe constitution.json + design.json en
// un dir temporal y confirma que el material del juez trae el artefacto y SOLO
// los principios de la fase.
func TestContextForJudge(t *testing.T) {
	dir := t.TempDir()
	sf := filepath.Join(dir, "specforge")
	featDir := filepath.Join(sf, "features", "trunc")
	if err := os.MkdirAll(featDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cons := `{"principles":[
		{"id":"P1","statement":"offline","applies_to":["design"]},
		{"id":"P2","statement":"otra","applies_to":["tasks"]}
	]}`
	if err := os.WriteFile(filepath.Join(sf, "constitution.json"), []byte(cons), 0o644); err != nil {
		t.Fatal(err)
	}
	design := `{"feature":"trunc","components":[{"id":"C1","name":"X"}]}`
	if err := os.WriteFile(filepath.Join(featDir, "design.json"), []byte(design), 0o644); err != nil {
		t.Fatal(err)
	}

	if cf, ok := loadConstitutionQuiet(dir); !ok {
		t.Fatal("loadConstitutionQuiet false")
	} else {
		ps, _ := principlesForPhase(cf, "design")
		if len(ps) != 1 || ps[0].ID != "P1" {
			t.Errorf("design principles=%v, want [P1]", ps)
		}
	}

	// El artefacto crudo debe parsear y traer el componente.
	raw, _ := os.ReadFile(phaseArtifactPath(dir, "trunc", "design"))
	var parsed map[string]any
	if err := json.Unmarshal(raw, &parsed); err != nil || parsed["feature"] != "trunc" {
		t.Errorf("artefacto design no parsea bien: %v", err)
	}
}
