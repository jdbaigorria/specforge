package main

import (
	"path/filepath"
	"testing"
)

// ----------------------------------------------------------------------------
// F4 de CLI-COMO-FUENTE (R5): la regla de las dos rutas.
//
// El defecto que arregla: `sf-amend` asumía SIEMPRE que el spec estaba mal. Ante
// cualquier divergencia, la única jugada disponible era actualizar el spec a lo
// que el código hace — así que el spec dejaba de ser un contrato y pasaba a ser
// un registro de lo que pasó.
// ----------------------------------------------------------------------------

func TestDeltaKindDefaults(t *testing.T) {
	t.Run("un delta sin kind se lee como spec-wrong — migración cero", func(t *testing.T) {
		if got := deltaKindOf(deltaFile{}); got != deltaKindSpecWrong {
			t.Errorf("kind=%q, want %q", got, deltaKindSpecWrong)
		}
		// Y valida: los deltas escritos antes de este campo no se rompen.
		var rep report
		checkDelta(deltaFile{Why: "x", Changes: []deltaChange{
			{Op: "modify", Target: "requirements", Ref: "R1", Description: "d"},
		}}, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("un delta legado sin kind debería validar, got %v", rep.errors)
		}
	})

	t.Run("un kind desconocido no valida", func(t *testing.T) {
		var rep report
		checkDelta(deltaFile{Kind: "whatever", Why: "x", Changes: []deltaChange{
			{Op: "modify", Target: "requirements", Ref: "R1", Description: "d"},
		}}, &rep)
		if len(rep.errors) == 0 {
			t.Error("kind fuera del vocabulario debería ser error")
		}
	})

	t.Run("code-wrong exige evidencia", func(t *testing.T) {
		var rep report
		checkDelta(deltaFile{Kind: deltaKindCodeWrong, Why: "x", Changes: []deltaChange{
			{Op: "modify", Target: "requirements", Ref: "R1", Description: "d"},
		}}, &rep)
		// Faltan expected y observed: dos errores.
		if len(rep.errors) != 2 {
			t.Errorf("code-wrong sin expected/observed → 2 errores, got %v", rep.errors)
		}
	})

	t.Run("--kind gana sobre el JSON", func(t *testing.T) {
		proj := t.TempDir()
		mustWrite(t, filepath.Join(proj, "specforge/features/f/drafts/d.json"),
			`{"kind":"spec-wrong","why":"w","expected":"e","observed":"o","changes":[
				{"op":"modify","target":"requirements","ref":"R1","description":"d"}]}`)
		if code := runDeltaNew([]string{"--feature=f", "--kind=code-wrong", "--from=drafts/d.json", proj}); code != 0 {
			t.Fatalf("delta new exit=%d", code)
		}
		if got := deltaKindOf(listDeltas(proj, "f")[0]); got != deltaKindCodeWrong {
			t.Errorf("kind=%q, want %q — el flag es lo que el humano tecleó recién", got, deltaKindCodeWrong)
		}
	})
}

// TestCodeWrongFreezesSpec es el test que le da dientes a la ruta (b). Sin él,
// elegir code-wrong sería una anotación decorativa.
func TestCodeWrongFreezesSpec(t *testing.T) {
	newProject := func(t *testing.T, kind string) string {
		t.Helper()
		proj := t.TempDir()
		draft := `{"why":"el paginado no respeta el limite","changes":[
			{"op":"modify","target":"requirements","ref":"R1","description":"d"}]}`
		if kind == deltaKindCodeWrong {
			draft = `{"why":"el paginado no respeta el limite","expected":"R1 pide max 50 por pagina",
				"observed":"devuelve la coleccion entera","changes":[
				{"op":"modify","target":"requirements","ref":"R1","description":"d"}]}`
		}
		mustWrite(t, filepath.Join(proj, "specforge/features/f/drafts/d.json"), draft)
		if code := runDeltaNew([]string{"--feature=f", "--kind=" + kind, "--from=drafts/d.json", proj}); code != 0 {
			t.Fatalf("delta new exit=%d", code)
		}
		return proj
	}

	reqJSON := filepath.Join("drafts", "requirements.json")
	writeReqDraft := func(t *testing.T, proj string) {
		t.Helper()
		mustWrite(t, filepath.Join(proj, "specforge/features/f", reqJSON),
			`{"feature":"f","requirements":[{"id":"R1","ears_type":"ubiquitous","behavior":"paginate results at 50 per page","acceptance":["devuelve 50"]}]}`)
	}

	t.Run("con un code-wrong abierto, sf save requirements se rechaza", func(t *testing.T) {
		proj := newProject(t, deltaKindCodeWrong)
		writeReqDraft(t, proj)
		if code := runSave([]string{"requirements", "--feature=f", "--from=" + reqJSON, proj}); code != 5 {
			t.Errorf("exit=%d, want 5 — el spec de esa feature está congelado", code)
		}
	})

	t.Run("con un spec-wrong, el mismo save pasa", func(t *testing.T) {
		proj := newProject(t, deltaKindSpecWrong)
		writeReqDraft(t, proj)
		if code := runSave([]string{"requirements", "--feature=f", "--from=" + reqJSON, proj}); code != 0 {
			t.Errorf("exit=%d, want 0 — spec-wrong es justamente la ruta que edita el spec", code)
		}
	})

	t.Run("archivar el code-wrong descongela el spec", func(t *testing.T) {
		proj := newProject(t, deltaKindCodeWrong)
		for _, to := range []string{"applied", "archived"} {
			if code := runDeltaSetStatus([]string{"--feature=f", "--id=D1", "--to=" + to, proj}); code != 0 {
				t.Fatalf("set-status %s exit=%d", to, code)
			}
		}
		writeReqDraft(t, proj)
		if code := runSave([]string{"requirements", "--feature=f", "--from=" + reqJSON, proj}); code != 0 {
			t.Errorf("exit=%d, want 0 — cerrar el defecto es la salida explícita", code)
		}
	})

	t.Run("el trace NO se congela — arreglar el bug exige re-anclar", func(t *testing.T) {
		proj := newProject(t, deltaKindCodeWrong)
		mustWrite(t, filepath.Join(proj, "src/pager.go"), "package src\n\nfunc Paginate() {}\n")
		mustWrite(t, filepath.Join(proj, "src/pager_test.go"), "package src\n\nfunc TestPaginate() {}\n")
		mustWrite(t, filepath.Join(proj, "specforge/features/f/drafts/trace.json"),
			`{"feature":"f","requirements":{"R1":{"code":["src/pager.go:Paginate"],"test":["src/pager_test.go:TestPaginate"],"status":"ok"}}}`)
		if code := runSave([]string{"trace", "--feature=f", "--from=drafts/trace.json", proj}); code != 0 {
			t.Errorf("exit=%d, want 0 — bloquear el trace dejaría la ruta (b) sin salida", code)
		}
	})

	t.Run("otra feature no se ve afectada", func(t *testing.T) {
		proj := newProject(t, deltaKindCodeWrong)
		mustWrite(t, filepath.Join(proj, "specforge/features/otra", reqJSON),
			`{"feature":"otra","requirements":[{"id":"R1","ears_type":"ubiquitous","behavior":"log every request","acceptance":["loguea"]}]}`)
		if code := runSave([]string{"requirements", "--feature=otra", "--from=" + reqJSON, proj}); code != 0 {
			t.Errorf("exit=%d, want 0 — el congelamiento es por feature", code)
		}
	})
}
