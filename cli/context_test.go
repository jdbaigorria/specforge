package main

import (
	"reflect"
	"testing"
)

// TestBuildWaveContext prueba la función pura: la membresía sale del plan, los
// cuerpos de tasksByID; dedup de refs, resumen de waves previas, y "no
// encontrada" para una wave inexistente.
func TestBuildWaveContext(t *testing.T) {
	// plan: wave 0 = [T1], wave 1 = [T2, T3].
	pf := planFile{Feature: "f", Waves: []planWave{
		{N: 0, Name: "Foundation", Tasks: []string{"T1"}},
		{N: 1, Name: "Cmds", Tasks: []string{"T2", "T3"}},
	}}
	tasksByID := map[string]task{
		"T1": {ID: "T1", Status: "done", RequirementRefs: []string{"R5"}, ComponentRefs: []string{"C1"}},
		"T2": {ID: "T2", RequirementRefs: []string{"R1", "R1"}, ComponentRefs: []string{"C3"}, FilesTouched: []string{"a.go"}},
		"T3": {ID: "T3", RequirementRefs: []string{"R2"}},
	}

	// reqByID con el cuerpo de R1 → el slice debe traerlo en Requirements.
	reqByID := map[string]requirement{
		"R1": {ID: "R1", EarsType: "event", Trigger: "user runs add", Behavior: "create a task"},
	}
	// compByID con C3 (el componente referenciado por la task T2 de la wave 1).
	compByID := map[string]component{
		"C3": {ID: "C3", Name: "AddCommand"},
	}

	ctx, ok := buildWaveContext(pf, tasksByID, reqByID, compByID, "f", 1)
	if !ok {
		t.Fatal("wave 1 should be found")
	}
	// R1 está en reqByID → su cuerpo viaja; R2 no → solo queda como ref.
	if len(ctx.Requirements) != 1 || ctx.Requirements[0].ID != "R1" {
		t.Errorf("Requirements=%v, want [R1 body]", ctx.Requirements)
	}
	// C3 está en compByID → su cuerpo viaja.
	if len(ctx.Components) != 1 || ctx.Components[0].ID != "C3" {
		t.Errorf("Components=%v, want [C3 body]", ctx.Components)
	}
	// R1 aparece dos veces → debe deduplicarse, y quedar ordenado.
	if !reflect.DeepEqual(ctx.RequirementRefs, []string{"R1", "R2"}) {
		t.Errorf("RequirementRefs=%v, want [R1 R2]", ctx.RequirementRefs)
	}
	// Solo la wave 0 es previa a la 1.
	if len(ctx.PriorWaves) != 1 || ctx.PriorWaves[0].N != 0 {
		t.Errorf("PriorWaves=%v, want one wave n=0", ctx.PriorWaves)
	}
	if ctx.PriorWaves[0].Statuses["done"] != 1 {
		t.Errorf("status histogram=%v, want done:1", ctx.PriorWaves[0].Statuses)
	}

	if _, ok := buildWaveContext(pf, tasksByID, nil, nil, "f", 9); ok {
		t.Errorf("wave 9 should not be found")
	}
}
