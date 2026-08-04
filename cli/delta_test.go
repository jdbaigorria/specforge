package main

import (
	"path/filepath"
	"testing"
)

// TestDeltaLifecycle (D4'): new valida y asigna id/estado/timestamp; el ciclo
// de vida es la escalera proposed→applied→archived sin saltos ni retornos.
func TestDeltaLifecycle(t *testing.T) {
	proj := t.TempDir()
	draft := filepath.Join(proj, "specforge", "features", "f", "drafts", "delta.json")
	mustWrite(t, draft, `{"why":"R3 cambió su contrato","changes":[
		{"op":"modify","target":"requirements","ref":"R3","description":"acepta paginación"},
		{"op":"add","target":"tasks","ref":"T9","description":"task nueva para paginado"}]}`)

	if code := runDeltaNew([]string{"--feature=f", "--from=drafts/delta.json", proj}); code != 0 {
		t.Fatalf("delta new exit=%d", code)
	}
	deltas := listDeltas(proj, "f")
	if len(deltas) != 1 {
		t.Fatalf("esperaba 1 delta, hay %d", len(deltas))
	}
	d := deltas[0]
	if d.ID != "D1" || d.Status != "proposed" || d.CreatedAt == "" || d.Feature != "f" {
		t.Errorf("el CLI fija id/status/timestamps: %+v", d)
	}
	// El borrador se retiró al promover (mismo contrato que sf save --from).
	if fileExists(draft) {
		t.Error("el draft debería retirarse tras delta new")
	}

	// Saltarse "applied" es ilegal.
	if code := runDeltaSetStatus([]string{"--feature=f", "--id=D1", "--to=archived", proj}); code != 5 {
		t.Errorf("proposed → archived directo → exit 5, got %d", code)
	}
	if code := runDeltaSetStatus([]string{"--feature=f", "--id=D1", "--to=applied", proj}); code != 0 {
		t.Errorf("proposed → applied → exit 0")
	}
	if d := listDeltas(proj, "f")[0]; d.AppliedAt == "" {
		t.Error("applied debe estampar applied_at")
	}
	if code := runDeltaSetStatus([]string{"--feature=f", "--id=D1", "--to=archived", proj}); code != 0 {
		t.Errorf("applied → archived → exit 0")
	}
	// archived es terminal.
	if code := runDeltaSetStatus([]string{"--feature=f", "--id=D1", "--to=applied", proj}); code != 5 {
		t.Errorf("archived es terminal → exit 5, got %d", code)
	}
}

// TestDeltaValidation: un delta sin why o con op/target/ref inválidos no se guarda.
func TestDeltaValidation(t *testing.T) {
	var rep report
	checkDelta(deltaFile{Why: "", Changes: nil}, &rep)
	if len(rep.errors) < 2 {
		t.Errorf("sin why ni changes → ≥2 errores, got %v", rep.errors)
	}
	rep = report{}
	checkDelta(deltaFile{Why: "x", Changes: []deltaChange{
		{Op: "explode", Target: "requirements", Ref: "R1", Description: "d"},
		{Op: "add", Target: "everything", Ref: "R2", Description: "d"},
		{Op: "modify", Target: "design", Ref: "", Description: "d"},
	}}, &rep)
	if len(rep.errors) != 3 {
		t.Errorf("op+target+ref inválidos → 3 errores, got %v", rep.errors)
	}
}

// TestHookProtectsDeltas: los deltas son estado — Write directo denegado.
func TestHookProtectsDeltas(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, featureFilePath(proj, "f"), `{"name":"f","status":"done","gates":[]}`)
	dec, reason := decidePreToolUse(proj, filepath.Join(proj, "specforge", "features", "f", "deltas", "D1.json"))
	if dec != "deny" {
		t.Error("deltas/*.json debería estar protegido")
	}
	if reason == "" {
		t.Error("el deny debe enseñar el camino sf delta")
	}
}
