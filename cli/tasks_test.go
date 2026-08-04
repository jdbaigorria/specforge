package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderTasksMarkdown prueba la plantilla directamente sobre un struct.
func TestRenderTasksMarkdown(t *testing.T) {
	tf := tasksFile{
		Feature: "demo",
		Tasks: []task{
			{ID: "T1", Title: "thing", RequirementRefs: []string{"R1"}, Status: "done"},
			{ID: "T2", Title: "other", Status: "pending", DependsOn: []string{"T1"}},
		},
	}
	out, err := renderTasksMarkdown(tf)
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	for _, want := range []string{"# Tasks — demo", "**T1**", "**T2**", "depends on: T1", "—"} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\n--- got ---\n%s", want, out)
		}
	}
}

// TestRenderTasks prueba los exit codes y que escriba el archivo.
func TestRenderTasks(t *testing.T) {
	t.Run("stdout -> 0", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "specforge/features/f/tasks.json"),
			`{"feature":"f","tasks":[{"id":"T1","title":"x","status":"done"}]}`)
		if code := renderTasks(dir, "f", true); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
	t.Run("missing tasks.json -> 4", func(t *testing.T) {
		if code := renderTasks(t.TempDir(), "nope", true); code != 4 {
			t.Errorf("exit=%d, want 4", code)
		}
	})
	t.Run("writes tasks.md -> 0", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "specforge/features/f/tasks.json"), `{"feature":"f","tasks":[]}`)
		if code := renderTasks(dir, "f", false); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
		if _, err := os.Stat(filepath.Join(dir, "specforge/features/f/tasks.md")); err != nil {
			t.Errorf("tasks.md not written: %v", err)
		}
	})
}

// TestCheckTasks prueba la consistencia interna: un archivo válido no da
// errores; las violaciones (id duplicado, status inválido, feature faltante)
// sí.
func TestCheckTasks(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tf := tasksFile{Feature: "f", Tasks: []task{
			{ID: "T1", Title: "a", RequirementRefs: []string{"R1"}, Status: "done"},
			{ID: "T2", Title: "b", DependsOn: []string{"T1"}},
		}}
		var rep report
		checkTasks(tf, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("unexpected errors: %v", rep.errors)
		}
	})

	t.Run("duplicate task id + bad status + missing feature", func(t *testing.T) {
		tf := tasksFile{Tasks: []task{
			{ID: "T1", Title: "a", Status: "done"},
			{ID: "T1", Title: "b", Status: "bogus"}, // id dup + status inválido
		}}
		var rep report
		checkTasks(tf, &rep)
		// esperamos: missing feature, duplicate T1, invalid status = 3 errores
		if len(rep.errors) != 3 {
			t.Errorf("errors=%d (%v), want 3", len(rep.errors), rep.errors)
		}
	})

	t.Run("malformed ref is a warning, not error", func(t *testing.T) {
		tf := tasksFile{Feature: "f", Tasks: []task{
			{ID: "T1", Title: "a", RequirementRefs: []string{"oops"}},
		}}
		var rep report
		checkTasks(tf, &rep)
		if len(rep.errors) != 0 || len(rep.warnings) != 1 {
			t.Errorf("errors=%d warnings=%d, want 0/1", len(rep.errors), len(rep.warnings))
		}
	})

	t.Run("unknown dep + cycle are errors", func(t *testing.T) {
		// dep colgada
		var rep report
		checkTasks(tasksFile{Feature: "f", Tasks: []task{{ID: "T1", Title: "a", DependsOn: []string{"T9"}}}}, &rep)
		if len(rep.errors) != 1 {
			t.Errorf("dep colgada: errors=%v, want 1", rep.errors)
		}
		// ciclo T1→T2→T1
		rep = report{}
		checkTasks(tasksFile{Feature: "f", Tasks: []task{
			{ID: "T1", Title: "a", DependsOn: []string{"T2"}},
			{ID: "T2", Title: "b", DependsOn: []string{"T1"}},
		}}, &rep)
		if len(rep.errors) != 1 {
			t.Errorf("ciclo: errors=%v, want 1", rep.errors)
		}
	})
}
