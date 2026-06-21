package main

import (
	"os"
	"path/filepath"
	"testing"
)

// ----------------------------------------------------------------------------
// Tests de paridad del engine de hooks (portados del --selftest del antiguo
// specforge_enforce.py). Son la red de seguridad que respalda borrar el Python:
// si estos pasan, `sf hook` decide igual que el .py.
// ----------------------------------------------------------------------------

// writeReg escribe specforge/features.json con el contenido dado (helper local).
func writeReg(t *testing.T, proj, content string) {
	t.Helper()
	p := filepath.Join(proj, "specforge", "features.json")
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestDecidePreToolUseGateChain: la cadena de gates (downstream necesita upstream
// aprobado), gateando tanto .md como .json (JSON-first), plan bajo progress/.
func TestDecidePreToolUseGateChain(t *testing.T) {
	proj := t.TempDir()
	feat := filepath.Join(proj, "specforge", "features", "x")
	if err := os.MkdirAll(filepath.Join(feat, "progress"), 0o755); err != nil {
		t.Fatal(err)
	}
	d := func(rel string) string {
		dec, _ := decidePreToolUse(proj, filepath.Join(feat, rel))
		return dec
	}

	// sin gates → design denegado, requirements permitido
	writeReg(t, proj, `{"schema_version":"1.0","features":[{"name":"x","status":"approved","gates":[]}]}`)
	if d("design.md") != "deny" {
		t.Error("design.md debería estar denegado sin gate de requirements")
	}
	if d("requirements.md") != "allow" {
		t.Error("requirements.md debería estar permitido (sin upstream)")
	}

	// requirements aprobado → design permitido, tasks aún denegado (.md y .json)
	writeReg(t, proj, `{"schema_version":"1.0","features":[{"name":"x","gates":[{"phase":"requirements","result":"approve"}]}]}`)
	if d("design.md") != "allow" {
		t.Error("design.md debería permitirse tras el gate de requirements")
	}
	if d("tasks.md") != "deny" || d("tasks.json") != "deny" {
		t.Error("tasks.md/.json deberían estar denegados sin gate de design")
	}
	if d(filepath.Join("progress", "plan.md")) != "deny" {
		t.Error("plan debería estar denegado sin gate de tasks")
	}

	// cadena completa → plan.json permitido
	writeReg(t, proj, `{"schema_version":"1.0","features":[{"name":"x","gates":[
		{"phase":"requirements","result":"approve"},
		{"phase":"design","result":"approve"},
		{"phase":"tasks","result":"approve"}]}]}`)
	if d(filepath.Join("progress", "plan.json")) != "allow" {
		t.Error("plan.json debería permitirse tras el gate de tasks")
	}
}

// TestDecidePreToolUseSerial: con una feature en curso, arrancar otra (su
// requirements) se deniega; el requirements de la propia activa se permite.
func TestDecidePreToolUseSerial(t *testing.T) {
	proj := t.TempDir()
	zreq := filepath.Join(proj, "specforge", "features", "z", "requirements.md")
	yreq := filepath.Join(proj, "specforge", "features", "y", "requirements.md")

	writeReg(t, proj, `{"schema_version":"1.0","features":[
		{"name":"y","status":"building"},{"name":"z","status":"planned"}]}`)
	if dec, _ := decidePreToolUse(proj, zreq); dec != "deny" {
		t.Error("serial: el requirements de la 2da feature debería denegarse con otra activa")
	}
	if dec, _ := decidePreToolUse(proj, yreq); dec != "allow" {
		t.Error("serial: el requirements de la feature activa debería permitirse")
	}

	writeReg(t, proj, `{"schema_version":"1.0","features":[{"name":"z","status":"planned"}]}`)
	if dec, _ := decidePreToolUse(proj, zreq); dec != "allow" {
		t.Error("serial: requirements permitido cuando no hay ninguna activa")
	}
}

// TestDecidePreToolUseGuards: .state/ siempre denegado; fuera de specforge/ o
// fuera de un proyecto SpecForge → permitido (no-op).
func TestDecidePreToolUseGuards(t *testing.T) {
	proj := t.TempDir()
	writeReg(t, proj, `{"schema_version":"1.0","features":[]}`)

	if dec, _ := decidePreToolUse(proj, filepath.Join(proj, "specforge", ".state", "session.md")); dec != "deny" {
		t.Error(".state/ debería estar denegado")
	}
	if dec, _ := decidePreToolUse(proj, filepath.Join(proj, "src", "main.go")); dec != "allow" {
		t.Error("ruta fuera de specforge/ debería permitirse")
	}
	// proyecto sin specforge/ → no-op allow
	noSF := t.TempDir()
	if dec, _ := decidePreToolUse(noSF, filepath.Join(noSF, "specforge", "features", "y", "design.md")); dec != "allow" {
		t.Error("proyecto no-SpecForge: debería ser no-op allow")
	}
}

// TestDecideInjection: el trigger del slice (pura).
func TestDecideInjection(t *testing.T) {
	if m, n := decideInjection([]string{"f", "build", "1"}, []string{"f", "build", "1"}, 3, 10); m != "breadcrumb" || n != 4 {
		t.Errorf("sin cambio + turnos bajos → breadcrumb/4, got %s/%d", m, n)
	}
	if m, n := decideInjection([]string{"f", "design", ""}, []string{"f", "requirements", ""}, 2, 10); m != "full" || n != 0 {
		t.Errorf("on-step-change → full/0, got %s/%d", m, n)
	}
	if m, n := decideInjection([]string{"f", "build", "0"}, []string{"f", "build", "0"}, 10, 10); m != "full" || n != 0 {
		t.Errorf("backstop N turnos → full/0, got %s/%d", m, n)
	}
}

// TestPickJournalNudge: el conciliador de memoria (pura).
func TestPickJournalNudge(t *testing.T) {
	feats := []feature{
		{Name: "old", Status: "done"},
		{Name: "cur", Status: "building"},
		{Name: "add-export", Status: "archived"}, // nombre con guion
	}
	if got := pickJournalNudge(feats, map[string]bool{}, map[string]bool{}); got != "old" {
		t.Errorf("primera archivada sin journal → old, got %q", got)
	}
	if got := pickJournalNudge(feats, map[string]bool{"old": true}, map[string]bool{}); got != "add-export" {
		t.Errorf("ya journaleada se saltea → add-export, got %q", got)
	}
	if got := pickJournalNudge(feats, map[string]bool{"old": true, "add-export": true}, map[string]bool{}); got != "" {
		t.Errorf("todas journaleadas → '', got %q", got)
	}
	if got := pickJournalNudge(feats, map[string]bool{}, map[string]bool{"old": true}); got != "add-export" {
		t.Errorf("ya nudgeada se saltea → add-export, got %q", got)
	}
}

// TestForceFullSlice: PreCompact marca el flag por sesión, y el próximo
// UserPromptSubmit lo consume.
func TestForceFullSlice(t *testing.T) {
	proj := t.TempDir()
	if err := os.MkdirAll(filepath.Join(proj, "specforge"), 0o755); err != nil {
		t.Fatal(err)
	}
	forceFullSlice(proj, "sess-x")
	st := readHookState(filepath.Join(proj, "specforge"))
	if !st.Sessions["sess-x"].ForceFull {
		t.Error("force_full_slice debería marcar el flag por sesión")
	}
}

// TestFullSliceEvery: configurable por env.
func TestFullSliceEvery(t *testing.T) {
	os.Unsetenv("SPECFORGE_FULL_SLICE_EVERY")
	if fullSliceEvery() != fullSliceEveryDefault {
		t.Error("sin env → default")
	}
	t.Setenv("SPECFORGE_FULL_SLICE_EVERY", "3")
	if fullSliceEvery() != 3 {
		t.Error("env válido → override")
	}
	t.Setenv("SPECFORGE_FULL_SLICE_EVERY", "junk")
	if fullSliceEvery() != fullSliceEveryDefault {
		t.Error("env basura → default")
	}
}
