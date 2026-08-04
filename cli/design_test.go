package main

import "testing"

// TestCheckDesign cubre la integridad referencial interna (depends_on) y la
// regla de que `chosen` sea una de las opciones.
func TestCheckDesign(t *testing.T) {
	t.Run("valid with internal depends_on", func(t *testing.T) {
		df := designFile{Feature: "f", Components: []component{
			{ID: "C1", Name: "Model"},
			{ID: "C2", Name: "Storage", DependsOn: []string{"C1"}},
		}}
		var rep report
		checkDesign(df, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("unexpected errors: %v", rep.errors)
		}
	})

	t.Run("depends_on unknown component", func(t *testing.T) {
		df := designFile{Feature: "f", Components: []component{
			{ID: "C1", Name: "Model", DependsOn: []string{"C9"}}, // C9 no existe
		}}
		var rep report
		checkDesign(df, &rep)
		if len(rep.errors) != 1 {
			t.Errorf("errors=%d (%v), want 1", len(rep.errors), rep.errors)
		}
	})

	t.Run("chosen not among options + duplicate id", func(t *testing.T) {
		df := designFile{Feature: "f",
			Components: []component{{ID: "C1", Name: "A"}, {ID: "C1", Name: "B"}}, // id dup
			Decisions:  []decision{{ID: "D1", Options: []decisionOption{{Name: "X"}}, Chosen: "Y"}},
		}
		var rep report
		checkDesign(df, &rep)
		// duplicate C1 + chosen "Y" not an option = 2
		if len(rep.errors) != 2 {
			t.Errorf("errors=%d (%v), want 2", len(rep.errors), rep.errors)
		}
	})
}
