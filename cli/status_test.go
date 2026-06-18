package main

import (
	"reflect"
	"testing"
)

// TestTopoOrder verifica que las dependencias salgan antes que quien depende de
// ellas, y que no se reporte un ciclo cuando no lo hay.
func TestTopoOrder(t *testing.T) {
	feats := map[string]*feature{
		"base":   {Name: "base"},
		"feat-b": {Name: "feat-b", DependsOn: []string{"base"}},
	}
	order, cycle := topoOrder(feats, []string{"base", "feat-b"})

	if len(cycle) != 0 {
		t.Errorf("cycle=%v, want none", cycle)
	}
	if indexOf(order, "base") > indexOf(order, "feat-b") {
		t.Errorf("order=%v, base should come before feat-b", order)
	}
}

// TestTopoOrderCycle verifica que un ciclo a→b→a se detecte.
func TestTopoOrderCycle(t *testing.T) {
	feats := map[string]*feature{
		"a": {Name: "a", DependsOn: []string{"b"}},
		"b": {Name: "b", DependsOn: []string{"a"}},
	}
	if _, cycle := topoOrder(feats, []string{"a", "b"}); len(cycle) == 0 {
		t.Errorf("expected a cycle, got none")
	}
}

// TestBlockers: una dep "done" no bloquea; una dep inexistente sí.
func TestBlockers(t *testing.T) {
	feats := map[string]*feature{
		"base":   {Name: "base", Status: "done"},
		"feat-b": {Name: "feat-b", Status: "approved", DependsOn: []string{"base"}},
		"feat-c": {Name: "feat-c", Status: "approved", DependsOn: []string{"missing-dep"}},
	}

	if bs := blockers(feats["feat-b"], feats); len(bs) != 0 {
		t.Errorf("feat-b blockers=%v, want none (base is done)", bs)
	}
	// reflect.DeepEqual compara slices/maps por contenido (== no funciona en slices).
	if bs := blockers(feats["feat-c"], feats); !reflect.DeepEqual(bs, []string{"missing-dep"}) {
		t.Errorf("feat-c blockers=%v, want [missing-dep]", bs)
	}
}

// TestLastPhase: devuelve la fase del ÚLTIMO gate aprobado, ignorando rechazos.
func TestLastPhase(t *testing.T) {
	f := &feature{Gates: []gate{
		{Phase: "requirements", Result: "approve"},
		{Phase: "design", Result: "reject"},
		{Phase: "tasks", Result: "approve"},
	}}
	if got := lastPhase(f); got != "tasks" {
		t.Errorf("lastPhase=%q, want tasks", got)
	}
	if got := lastPhase(&feature{}); got != "—" {
		t.Errorf("lastPhase(empty)=%q, want —", got)
	}
}

// indexOf devuelve el índice de s en xs, o -1 si no está.
func indexOf(xs []string, s string) int {
	for i, x := range xs {
		if x == s {
			return i
		}
	}
	return -1
}
