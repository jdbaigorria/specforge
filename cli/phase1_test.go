package main

import "testing"

// Tests de los artefactos restantes de la Fase 1: constitution, plan, review.

func TestCheckConstitution(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cf := constitutionFile{
			IdentityMD: "A CLI tool.",
			Invariants: []invariant{{ID: "I1", Rule: "every interface handles errors"}},
		}
		var rep report
		checkConstitution(cf, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("unexpected errors: %v", rep.errors)
		}
	})

	t.Run("duplicate invariant id + empty rule", func(t *testing.T) {
		cf := constitutionFile{IdentityMD: "x", Invariants: []invariant{
			{ID: "I1", Rule: "a"},
			{ID: "I1", Rule: ""}, // id dup + rule vacío
		}}
		var rep report
		checkConstitution(cf, &rep)
		if len(rep.errors) != 2 {
			t.Errorf("errors=%d (%v), want 2", len(rep.errors), rep.errors)
		}
	})

	t.Run("empty identity is a warning", func(t *testing.T) {
		var rep report
		checkConstitution(constitutionFile{}, &rep)
		if len(rep.errors) != 0 || len(rep.warnings) != 1 {
			t.Errorf("errors=%d warnings=%d, want 0/1", len(rep.errors), len(rep.warnings))
		}
	})
}

func TestCheckPlanInternal(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		pf := planFile{Feature: "f", Waves: []planWave{{N: 0, Complexity: "low"}, {N: 1, Complexity: "high"}}}
		var rep report
		checkPlanInternal(pf, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("unexpected errors: %v", rep.errors)
		}
	})

	t.Run("duplicate wave + invalid complexity", func(t *testing.T) {
		pf := planFile{Feature: "f", Waves: []planWave{{N: 0, Complexity: "low"}, {N: 0, Complexity: "huge"}}}
		var rep report
		checkPlanInternal(pf, &rep)
		if len(rep.errors) != 2 {
			t.Errorf("errors=%d (%v), want 2", len(rep.errors), rep.errors)
		}
	})
}

// TestCheckPlan: el cruce con tasks.json — cobertura (toda task en una wave) y
// el GUARD wave(t) > wave(deps).
func TestCheckPlan(t *testing.T) {
	tf := tasksFile{Feature: "f", Tasks: []task{
		{ID: "T1", Title: "a"},
		{ID: "T2", Title: "b", DependsOn: []string{"T1"}},
	}}

	t.Run("valid: T1 wave0, T2 wave1", func(t *testing.T) {
		pf := planFile{Feature: "f", Waves: []planWave{
			{N: 0, Tasks: []string{"T1"}}, {N: 1, Tasks: []string{"T2"}},
		}}
		var rep report
		checkPlan(pf, tf, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("unexpected errors: %v", rep.errors)
		}
	})

	t.Run("guard: dep en la misma wave → error", func(t *testing.T) {
		pf := planFile{Feature: "f", Waves: []planWave{{N: 0, Tasks: []string{"T1", "T2"}}}}
		var rep report
		checkPlan(pf, tf, &rep)
		if len(rep.errors) != 1 {
			t.Errorf("errors=%v, want 1 (guard)", rep.errors)
		}
	})

	t.Run("task sin wave → error de cobertura", func(t *testing.T) {
		pf := planFile{Feature: "f", Waves: []planWave{{N: 0, Tasks: []string{"T1"}}}} // falta T2
		var rep report
		checkPlan(pf, tf, &rep)
		if len(rep.errors) != 1 {
			t.Errorf("errors=%v, want 1 (T2 sin asignar)", rep.errors)
		}
	})
}

func TestCheckReview(t *testing.T) {
	t.Run("valid approve", func(t *testing.T) {
		rv := reviewFile{Feature: "f", Verdict: "approve", Traceability: []traceRow{
			{Requirement: "R1", Status: "implemented"},
		}}
		var rep report
		checkReview(rv, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("unexpected errors: %v", rep.errors)
		}
	})

	t.Run("approve contradicted by missing row + violation", func(t *testing.T) {
		rv := reviewFile{Feature: "f", Verdict: "approve",
			Traceability:           []traceRow{{Requirement: "R1", Status: "missing"}},
			ConstitutionViolations: []string{"broke I1"},
		}
		var rep report
		checkReview(rv, &rep)
		// missing-while-approve + violation-while-approve = 2
		if len(rep.errors) != 2 {
			t.Errorf("errors=%d (%v), want 2", len(rep.errors), rep.errors)
		}
	})

	t.Run("invalid verdict + invalid status", func(t *testing.T) {
		rv := reviewFile{Feature: "f", Verdict: "maybe", Traceability: []traceRow{{Requirement: "R1", Status: "halfway"}}}
		var rep report
		checkReview(rv, &rep)
		if len(rep.errors) != 2 {
			t.Errorf("errors=%d (%v), want 2", len(rep.errors), rep.errors)
		}
	})
}
