package main

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFinishSave verifica el corazón genérico: válido escribe json+md y devuelve
// 0; inválido NO escribe nada y devuelve 2.
func TestFinishSave(t *testing.T) {
	t.Run("valid writes json + md", func(t *testing.T) {
		dir := t.TempDir()
		jsonPath := filepath.Join(dir, "specforge/features/f/tasks.json")
		mdPath := filepath.Join(dir, "specforge/features/f/tasks.md")

		v := tasksFile{Feature: "f", Tasks: []task{
			{ID: "T1", Title: "x", Status: "done"},
		}}
		if code := finishSave(v, jsonPath, mdPath); code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		// finishSave crea los directorios (MkdirAll) y ambos archivos.
		for _, p := range []string{jsonPath, mdPath} {
			if _, err := os.Stat(p); err != nil {
				t.Errorf("%s not written: %v", p, err)
			}
		}
	})

	t.Run("invalid writes nothing", func(t *testing.T) {
		dir := t.TempDir()
		jsonPath := filepath.Join(dir, "tasks.json")
		// id duplicado + feature faltante → la validación falla.
		v := tasksFile{Tasks: []task{
			{ID: "T1", Title: "a"},
			{ID: "T1", Title: "b"},
		}}
		if code := finishSave(v, jsonPath, filepath.Join(dir, "tasks.md")); code != 2 {
			t.Errorf("exit=%d, want 2", code)
		}
		if _, err := os.Stat(jsonPath); err == nil {
			t.Error("invalid artifact must not be written")
		}
	})
}

// TestSaveTrace: trace.json es estado protegido (Capa 1) → su único camino de
// escritura es `sf save trace`. Verifica validación + escritura json+md.
func TestSaveTrace(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "specforge/features/f/trace.json")
	mdPath := filepath.Join(dir, "specforge/features/f/trace.md")

	// Válido: feature + un requirement con anchor de código → escribe json + md.
	v := traceFile{Feature: "f", Requirements: map[string]traceReq{
		"R1": {Code: []string{"src/a.py:fn"}, Test: []string{"tests/t.py::test_fn"}, Status: "ok"},
	}}
	if code := finishSave(v, jsonPath, mdPath); code != 0 {
		t.Fatalf("trace válido → exit 0, got %d", code)
	}
	for _, p := range []string{jsonPath, mdPath} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s no escrito: %v", p, err)
		}
	}

	// Inválido: sin requirements → falla y no escribe.
	bad := filepath.Join(dir, "bad.json")
	if code := finishSave(traceFile{Feature: "f"}, bad, filepath.Join(dir, "bad.md")); code != 2 {
		t.Errorf("trace sin requirements → exit 2, got %d", code)
	}
	if _, err := os.Stat(bad); err == nil {
		t.Error("un trace inválido no debería escribirse")
	}
}

// TestSaveFromDraft: el camino cómodo de autoría (R5/D8). El agente escribe el
// borrador en drafts/ (dir no protegido), `sf save --from=` lo valida, escribe
// el canónico y RETIRA el borrador (para que no queden dos copias divergentes).
func TestSaveFromDraft(t *testing.T) {
	proj := t.TempDir()
	draft := filepath.Join(proj, "specforge", "features", "f", "drafts", "tasks.json")
	mustWrite(t, draft, `{"feature":"f","tasks":[{"id":"T1","title":"x","status":"done"}]}`)

	if code := runSave([]string{"tasks", "--feature=f", "--from=drafts/tasks.json", proj}); code != 0 {
		t.Fatalf("draft válido → exit 0, got %d", code)
	}
	// El canónico existe (json + md render), el borrador ya no.
	for _, p := range []string{"tasks.json", "tasks.md"} {
		if _, err := os.Stat(filepath.Join(proj, "specforge", "features", "f", p)); err != nil {
			t.Errorf("%s no escrito: %v", p, err)
		}
	}
	if _, err := os.Stat(draft); err == nil {
		t.Error("el borrador debería retirarse tras promoverse")
	}

	// Un draft INVÁLIDO no se promueve y el borrador queda (para corregirlo ahí).
	mustWrite(t, draft, `{"tasks":[{"id":"T1"},{"id":"T1"}]}`)
	if code := runSave([]string{"tasks", "--feature=f", "--from=drafts/tasks.json", proj}); code != 2 {
		t.Errorf("draft inválido → exit 2, got %d", code)
	}
	if _, err := os.Stat(draft); err != nil {
		t.Error("un borrador inválido debe quedar en drafts/ para iterarlo")
	}

	// --from y --json a la vez es ambiguo → exit 2.
	if code := runSave([]string{"tasks", "--feature=f", "--from=x.json", "--json=y.json", proj}); code != 2 {
		t.Error("--from + --json deberían ser mutuamente excluyentes")
	}
}

// TestDraftPath: resolución del borrador según nivel (feature vs proyecto) y
// tolerancia al prefijo "drafts/".
func TestDraftPath(t *testing.T) {
	cases := []struct{ feature, from, want string }{
		{"f", "tasks.json", "specforge/features/f/drafts/tasks.json"},
		{"f", "drafts/tasks.json", "specforge/features/f/drafts/tasks.json"},
		{"", "constitution.json", "specforge/drafts/constitution.json"},
	}
	for _, c := range cases {
		if got := draftPath(".", c.feature, c.from); got != filepath.FromSlash(c.want) {
			t.Errorf("draftPath(%q,%q)=%q, want %q", c.feature, c.from, got, c.want)
		}
	}
}

// TestArtifactPaths verifica el mapeo de rutas por artefacto.
func TestArtifactPaths(t *testing.T) {
	cases := []struct{ name, wantJSON string }{
		{"constitution", "specforge/constitution.json"},
		{"plan", "specforge/features/f/progress/plan.json"},
		{"tasks", "specforge/features/f/tasks.json"},
	}
	for _, c := range cases {
		gotJSON, _ := artifactPaths(c.name, ".", "f")
		if gotJSON != filepath.FromSlash(c.wantJSON) {
			t.Errorf("%s: json path=%q, want %q", c.name, gotJSON, c.wantJSON)
		}
	}
}
