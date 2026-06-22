package main

import (
	"path/filepath"
	"testing"
)

// TestRunTraceVerify reutiliza makeDriftProject (doctor_test.go), que crea un
// trace.json apuntando a src/slug.py:slugify.
func TestRunTraceVerify(t *testing.T) {
	t.Run("clean feature -> 0", func(t *testing.T) {
		dir := makeDriftProject(t)
		if code := runTraceVerify(dir, "slugify"); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
	t.Run("symbol removed -> 1", func(t *testing.T) {
		dir := makeDriftProject(t)
		writeFile(t, filepath.Join(dir, "src/slug.py"), "def x():\n    return 1\n")
		if code := runTraceVerify(dir, "slugify"); code != 1 {
			t.Errorf("exit=%d, want 1", code)
		}
	})
	t.Run("unknown feature -> 4", func(t *testing.T) {
		dir := makeDriftProject(t)
		if code := runTraceVerify(dir, "nope"); code != 4 {
			t.Errorf("exit=%d, want 4", code)
		}
	})
	t.Run("all features clean -> 0", func(t *testing.T) {
		dir := makeDriftProject(t)
		if code := runTraceVerify(dir, ""); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
}

// TestSplitAnchor cubre el parseo de anchors, incluido el caso "::" de pytest que
// el split por último ":" rompía (era el bug que impedía verificar test anchors).
func TestSplitAnchor(t *testing.T) {
	cases := []struct {
		anchor   string
		wantPath string
		wantSym  string
	}{
		{"tests/foo.py::test_x", "tests/foo.py", "test_x"},
		{"tests/foo.py::Cls::test_y", "tests/foo.py", "test_y"},
		{"tests/foo.py::test_z[case-1]", "tests/foo.py", "test_z"},
		{"foo_test.go:TestParse", "foo_test.go", "TestParse"},
		{"src/bar.py:Klass.method", "src/bar.py", "Klass.method"},
		{"src/only_path.go", "src/only_path.go", ""},
	}
	for _, c := range cases {
		gotPath, gotSym := splitAnchor(c.anchor)
		if gotPath != c.wantPath || gotSym != c.wantSym {
			t.Errorf("splitAnchor(%q) = (%q,%q), want (%q,%q)",
				c.anchor, gotPath, gotSym, c.wantPath, c.wantSym)
		}
	}
}

// makeContractProject arma una feature EN BUILD (features/, no archive) con código
// y un test reales, más plan.json (wave 0=T1→R1, wave 1=T2→R2) y tasks.json. Cada
// subtest escribe su propio trace.json para ejercitar un veredicto distinto.
func makeContractProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	base := filepath.Join(dir, "specforge/features/widget")
	writeFile(t, filepath.Join(dir, "src/widget.py"), "def build():\n    return 1\n")
	writeFile(t, filepath.Join(dir, "tests/test_widget.py"), "def test_build():\n    assert build()\n")
	writeFile(t, filepath.Join(base, "tasks.json"),
		`{"schema_version":"1.0","feature":"widget","tasks":[`+
			`{"id":"T1","title":"build","requirement_refs":["R1"],"depends_on":[]},`+
			`{"id":"T2","title":"polish","requirement_refs":["R2"],"depends_on":["T1"]}]}`)
	writeFile(t, filepath.Join(base, "progress/plan.json"),
		`{"schema_version":"1.0","feature":"widget","waves":[`+
			`{"n":0,"tasks":["T1"]},{"n":1,"tasks":["T2"]}]}`)
	return dir
}

func TestRunTraceContract(t *testing.T) {
	tracePath := "specforge/features/widget/trace.json"

	t.Run("wave 0 with a test that exists -> 0", func(t *testing.T) {
		dir := makeContractProject(t)
		writeFile(t, filepath.Join(dir, tracePath),
			`{"feature":"widget","requirements":{"R1":{"code":["src/widget.py:build"],"test":["tests/test_widget.py::test_build"],"status":"ok"}}}`)
		if code := runTraceContract(dir, "widget", 0); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})

	t.Run("wave 0 requirement with no test -> 1", func(t *testing.T) {
		dir := makeContractProject(t)
		writeFile(t, filepath.Join(dir, tracePath),
			`{"feature":"widget","requirements":{"R1":{"code":["src/widget.py:build"],"test":[],"status":"no-test"}}}`)
		if code := runTraceContract(dir, "widget", 0); code != 1 {
			t.Errorf("exit=%d, want 1", code)
		}
	})

	t.Run("wave 0 test anchor does not exist -> 1", func(t *testing.T) {
		dir := makeContractProject(t)
		writeFile(t, filepath.Join(dir, tracePath),
			`{"feature":"widget","requirements":{"R1":{"code":["src/widget.py:build"],"test":["tests/test_widget.py::test_missing"],"status":"ok"}}}`)
		if code := runTraceContract(dir, "widget", 0); code != 1 {
			t.Errorf("exit=%d, want 1", code)
		}
	})

	t.Run("wave 1 requirement not yet declared -> 1", func(t *testing.T) {
		dir := makeContractProject(t)
		// trace solo tiene R1; la wave 1 toca R2 → contrato incumplido.
		writeFile(t, filepath.Join(dir, tracePath),
			`{"feature":"widget","requirements":{"R1":{"code":["src/widget.py:build"],"test":["tests/test_widget.py::test_build"],"status":"ok"}}}`)
		if code := runTraceContract(dir, "widget", 1); code != 1 {
			t.Errorf("exit=%d, want 1", code)
		}
	})

	t.Run("missing trace.json -> 4", func(t *testing.T) {
		dir := makeContractProject(t)
		if code := runTraceContract(dir, "widget", 0); code != 4 {
			t.Errorf("exit=%d, want 4", code)
		}
	})

	t.Run("whole-feature scope all covered -> 0", func(t *testing.T) {
		dir := makeContractProject(t)
		writeFile(t, filepath.Join(dir, "tests/test_widget.py"),
			"def test_build():\n    assert build()\n\ndef test_polish():\n    assert True\n")
		writeFile(t, filepath.Join(dir, tracePath),
			`{"feature":"widget","requirements":{`+
				`"R1":{"code":["src/widget.py:build"],"test":["tests/test_widget.py::test_build"],"status":"ok"},`+
				`"R2":{"code":["src/widget.py:build"],"test":["tests/test_widget.py::test_polish"],"status":"ok"}}}`)
		if code := runTraceContract(dir, "widget", -1); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
}
