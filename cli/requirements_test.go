package main

import "testing"

// TestCheckRequirements cubre la gramática EARS: un set válido no da errores;
// las violaciones por tipo sí.
func TestCheckRequirements(t *testing.T) {
	t.Run("valid event + ubiquitous", func(t *testing.T) {
		rf := requirementsFile{Feature: "f", Requirements: []requirement{
			{ID: "R1", EarsType: "event", Trigger: "user acts", Behavior: "do x"},
			{ID: "R2", EarsType: "ubiquitous", Behavior: "persist data"},
		}}
		var rep report
		checkRequirements(rf, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("unexpected errors: %v", rep.errors)
		}
	})

	t.Run("event without trigger + event with state", func(t *testing.T) {
		rf := requirementsFile{Feature: "f", Requirements: []requirement{
			{ID: "R1", EarsType: "event", Behavior: "do x"},                           // falta trigger
			{ID: "R2", EarsType: "event", Trigger: "t", State: "s", Behavior: "do y"}, // event con state
		}}
		var rep report
		checkRequirements(rf, &rep)
		if len(rep.errors) != 2 {
			t.Errorf("errors=%d (%v), want 2", len(rep.errors), rep.errors)
		}
	})

	t.Run("invalid ears_type + duplicate id + missing feature", func(t *testing.T) {
		rf := requirementsFile{Requirements: []requirement{
			{ID: "R1", EarsType: "bogus", Behavior: "x"},
			{ID: "R1", EarsType: "ubiquitous", Behavior: "y"}, // id duplicado
		}}
		var rep report
		checkRequirements(rf, &rep)
		// missing feature + invalid ears_type + duplicate id = 3
		if len(rep.errors) != 3 {
			t.Errorf("errors=%d (%v), want 3", len(rep.errors), rep.errors)
		}
	})
}
