package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// makeGraphProject arma un proyecto con dos features (auth depende de base) y un
// trace.json para `base` donde R1 y R2 COMPARTEN el mismo símbolo de código
// (para ejercitar la deduplicación de nodos).
func makeGraphProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[`+
			`{"name":"base","status":"done","depends_on":[]},`+
			`{"name":"auth","status":"planned","depends_on":["base"]}]}`)
	writeFile(t, filepath.Join(dir, "specforge/features/base/trace.json"),
		`{"feature":"base","requirements":{`+
			`"R1":{"code":["src/m.py:f"],"test":["tests/t.py::test_a"],"status":"ok"},`+
			`"R2":{"code":["src/m.py:f"],"test":["tests/t.py::test_b"],"status":"ok"}}}`)
	return dir
}

// hasEdge busca una arista (from,to,type) en el grafo.
func hasEdge(g knowledgeGraph, from, to, typ string) bool {
	for _, e := range g.Edges {
		if e.From == from && e.To == to && e.Type == typ {
			return true
		}
	}
	return false
}

// countType cuenta los nodos de un tipo dado.
func countType(g knowledgeGraph, typ string) int {
	n := 0
	for _, nd := range g.Nodes {
		if nd.Type == typ {
			n++
		}
	}
	return n
}

func TestBuildGraph(t *testing.T) {
	t.Run("full graph: nodes, edges, dedup", func(t *testing.T) {
		dir := makeGraphProject(t)
		g, code := buildGraph(dir, "")
		if code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		// 2 features, 2 requirements, 1 code (compartido → dedup), 2 tests.
		if got := countType(g, "feature"); got != 2 {
			t.Errorf("feature nodes=%d, want 2", got)
		}
		if got := countType(g, "code"); got != 1 {
			t.Errorf("code nodes=%d, want 1 (dedup)", got)
		}
		if got := countType(g, "test"); got != 2 {
			t.Errorf("test nodes=%d, want 2", got)
		}
		// la arista de dependencia entre features.
		if !hasEdge(g, "feat_auth", "feat_base", "depends_on") {
			t.Error("missing depends_on edge auth→base")
		}
		// la espina: feature→R, R→code, R→test.
		if !hasEdge(g, "feat_base", "req_base_R1", "has_requirement") {
			t.Error("missing has_requirement edge")
		}
		if !hasEdge(g, "req_base_R1", "code_src_m_py_f", "implemented_by") {
			t.Error("missing implemented_by edge")
		}
		if !hasEdge(g, "req_base_R2", "test_tests_t_py__test_b", "verified_by") {
			t.Error("missing verified_by edge")
		}
	})

	t.Run("feature filter scopes out the rest", func(t *testing.T) {
		dir := makeGraphProject(t)
		g, _ := buildGraph(dir, "auth")
		if g.Scope != "auth" {
			t.Errorf("scope=%q, want auth", g.Scope)
		}
		if countType(g, "feature") != 1 {
			t.Errorf("feature nodes=%d, want 1 (only auth)", countType(g, "feature"))
		}
		// auth no tiene trace → sin nodos de requirement/code/test.
		if countType(g, "requirement") != 0 {
			t.Error("auth should have no requirement nodes")
		}
		// la dep apunta a base, que está FUERA de scope → no debe dibujarse.
		if hasEdge(g, "feat_auth", "feat_base", "depends_on") {
			t.Error("depends_on edge should be omitted when target is out of scope")
		}
	})

	t.Run("missing features.json -> 4", func(t *testing.T) {
		if _, code := buildGraph(t.TempDir(), ""); code != 4 {
			t.Errorf("exit=%d, want 4", code)
		}
	})

	t.Run("feature without trace is a bare node", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "specforge/features.json"),
			`{"schema_version":"1.0","features":[{"name":"solo","status":"planned","depends_on":[]}]}`)
		g, code := buildGraph(dir, "")
		if code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		if len(g.Nodes) != 1 || g.Nodes[0].Type != "feature" {
			t.Errorf("want 1 bare feature node, got %d nodes", len(g.Nodes))
		}
	})
}

func TestRenderGraphMermaid(t *testing.T) {
	dir := makeGraphProject(t)
	g, _ := buildGraph(dir, "")
	out := renderGraphMermaid(g)
	if !strings.HasPrefix(out, "%%") || !strings.Contains(out, "flowchart LR") {
		t.Error("mermaid output missing header / flowchart declaration")
	}
	// el nodo feature usa forma de estadio ([...]) y la arista lleva su tipo.
	if !strings.Contains(out, `feat_base(["feature: base"])`) {
		t.Error("feature node not rendered as a stadium shape")
	}
	if !strings.Contains(out, "-->|depends_on|") {
		t.Error("edge type label missing")
	}
}

func TestMermaidID(t *testing.T) {
	cases := map[string]string{
		"src/m.py:f":         "src_m_py_f",
		"tests/t.py::test_a": "tests_t_py__test_a",
		"add-kelvin":         "add_kelvin",
		"already_safe":       "already_safe",
	}
	for in, want := range cases {
		if got := mermaidID(in); got != want {
			t.Errorf("mermaidID(%q)=%q, want %q", in, got, want)
		}
	}
}
