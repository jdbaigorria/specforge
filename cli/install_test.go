package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// mkSourceRoot crea un dir que parece la raíz del repo (skills/ + AGENT.md).
func mkSourceRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "skills"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "AGENT.md"), []byte("# AGENT"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// TestIsSourceRootAndFindUp: reconoce la raíz y la encuentra subiendo desde un
// subdirectorio.
func TestIsSourceRootAndFindUp(t *testing.T) {
	root := mkSourceRoot(t)
	if !isSourceRoot(root) {
		t.Error("la raíz con skills/+AGENT.md debería reconocerse")
	}
	sub := filepath.Join(root, "skills", "sf-build")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if got := findSourceUp(sub); got != root {
		t.Errorf("findSourceUp(%q)=%q, want %q", sub, got, root)
	}
	// un dir sin skills/AGENT no es raíz
	if isSourceRoot(t.TempDir()) {
		t.Error("un dir vacío no debería ser raíz")
	}
}

// TestResolveSourceRootFrom: --from válido se acepta; inválido da error.
func TestResolveSourceRootFrom(t *testing.T) {
	root := mkSourceRoot(t)
	got, err := resolveSourceRoot(root)
	if err != nil || got != root {
		t.Errorf("--from válido: got %q err %v, want %q", got, err, root)
	}
	if _, err := resolveSourceRoot(t.TempDir()); err == nil {
		t.Error("--from inválido (sin skills/AGENT) debería dar error")
	}
}

// findAction busca una acción por kind cuyo target contenga sub (helper).
func findAction(p harnessPlan, kind, sub string) bool {
	for _, a := range p.actions {
		if a.kind == kind && strings.Contains(a.target, sub) {
			return true
		}
	}
	return false
}

func planByName(plans []harnessPlan, name string) harnessPlan {
	for _, p := range plans {
		if p.name == name {
			return p
		}
	}
	return harnessPlan{name: "<absent>"}
}

// TestPlanAllProjectScope: en scope proyecto los targets cuelgan de base, y cada
// arnés tiene las acciones esperadas.
func TestPlanAllProjectScope(t *testing.T) {
	root := mkSourceRoot(t)
	base := "/proj"
	home := "/home/u"
	plans := planAll(root, base, false, home)

	claude := planByName(plans, "claude-code")
	if !findAction(claude, "symlink", filepath.Join("/proj", ".claude", "skills")) {
		t.Error("claude: falta el symlink de skills a .claude/skills del proyecto")
	}
	pi := planByName(plans, "pi")
	if !findAction(pi, "wire", filepath.Join("/proj", ".pi", "settings.json")) {
		t.Error("pi: falta el wire a .pi/settings.json del proyecto")
	}
	oc := planByName(plans, "opencode")
	if !findAction(oc, "symlink", filepath.Join("/proj", ".opencode", "plugin")) {
		t.Error("opencode: falta el symlink del plugin")
	}
	cur := planByName(plans, "cursor")
	if !findAction(cur, "merge", filepath.Join("/proj", ".cursor", "rules")) {
		t.Error("cursor: falta el merge del .mdc")
	}
}

// TestPlanAllGlobalScope: en scope global los targets cuelgan de home.
func TestPlanAllGlobalScope(t *testing.T) {
	root := mkSourceRoot(t)
	home := "/home/u"
	plans := planAll(root, home, true, home)

	claude := planByName(plans, "claude-code")
	if !findAction(claude, "symlink", filepath.Join(home, ".claude", "skills")) {
		t.Error("claude global: skills debería ir a ~/.claude/skills")
	}
	pi := planByName(plans, "pi")
	if !findAction(pi, "wire", filepath.Join(home, ".pi", "agent", "settings.json")) {
		t.Error("pi global: settings debería ir a ~/.pi/agent/settings.json")
	}
}

// TestTildeHome: acorta rutas bajo $HOME a ~/…
func TestTildeHome(t *testing.T) {
	if got := tildeHome("/home/u", "/home/u/.claude/skills"); got != "~/.claude/skills" {
		t.Errorf("tildeHome=%q, want ~/.claude/skills", got)
	}
	if got := tildeHome("/home/u", "/proj/.pi"); got != "/proj/.pi" {
		t.Errorf("ruta fuera de home no debería cambiar: %q", got)
	}
}
