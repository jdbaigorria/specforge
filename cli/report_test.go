package main

import (
	"path/filepath"
	"testing"
)

// TestParseGoTestJSON: extrae pass/fail/skip por test de un stream `go test
// -json`, ignorando eventos de paquete y líneas que no son JSON.
func TestParseGoTestJSON(t *testing.T) {
	out := `ok some build output
{"Action":"run","Test":"TestFoo"}
{"Action":"output","Test":"TestFoo","Output":"=== RUN TestFoo\n"}
{"Action":"pass","Test":"TestFoo","Elapsed":0.01}
{"Action":"fail","Test":"TestBar"}
{"Action":"skip","Test":"TestBaz"}
{"Action":"pass","Elapsed":0.5}
not json at all
`
	tests := parseGoTestJSON(out)
	want := map[string]string{"TestFoo": "pass", "TestBar": "fail", "TestBaz": "skip"}
	if len(tests) != len(want) {
		t.Fatalf("tests=%v, want %v", tests, want)
	}
	for k, v := range want {
		if tests[k] != v {
			t.Errorf("%s=%q, want %q", k, tests[k], v)
		}
	}
}

// TestParseJUnitXML: soporta raíz <testsuites> con suites anidados y clasifica
// failure/error/skipped; ausencia de todos = pass.
func TestParseJUnitXML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.xml")
	mustWrite(t, path, `<?xml version="1.0"?>
<testsuites>
  <testsuite name="s1">
    <testcase classname="tests.t" name="test_ok"/>
    <testcase classname="tests.t" name="test_broken"><failure message="boom"/></testcase>
    <testcase classname="tests.t" name="test_skipped"><skipped/></testcase>
  </testsuite>
</testsuites>`)

	tests := parseJUnitXML(path)
	want := map[string]string{"test_ok": "pass", "test_broken": "fail", "test_skipped": "skip"}
	for k, v := range want {
		if tests[k] != v {
			t.Errorf("%s=%q, want %q", k, tests[k], v)
		}
	}

	// Raíz <testsuite> directa (pytest emite así a veces) también parsea.
	mustWrite(t, path, `<testsuite><testcase name="test_direct"/></testsuite>`)
	if got := parseJUnitXML(path); got["test_direct"] != "pass" {
		t.Errorf("raíz testsuite directa: %v", got)
	}

	// Archivo inexistente → nil (el caller decide).
	if got := parseJUnitXML(filepath.Join(dir, "nope.xml")); got != nil {
		t.Errorf("reporte inexistente → nil, got %v", got)
	}
}

// TestTestRefLeaf: la normalización de refs del trace al nombre del reporte.
func TestTestRefLeaf(t *testing.T) {
	cases := []struct{ ref, want string }{
		{"tests/test_x.py::test_fn", "test_fn"},
		{"tests/test_x.py::TestClass::test_method", "test_method"},
		{"cli/gate_test.go:TestFoo", "TestFoo"},
		{"TestBare", "TestBare"},
	}
	for _, c := range cases {
		if got := testRefLeaf(c.ref); got != c.want {
			t.Errorf("testRefLeaf(%q)=%q, want %q", c.ref, got, c.want)
		}
	}
}

// TestCausalityReasons: el corazón de A2 — un test que no corrió o no pasó en
// la corrida sellada produce una razón de rechazo del verdict.
func TestCausalityReasons(t *testing.T) {
	tf := &traceFile{Feature: "f", Requirements: map[string]traceReq{
		"R1": {Test: []string{"tests/t.py::test_r1"}},
		"R2": {Test: []string{"tests/t.py::test_missing"}},
		"R3": {Test: []string{"tests/t.py::test_skipped"}},
	}}
	tests := map[string]string{"test_r1": "pass", "test_skipped": "skip"}

	reasons := causalityReasons(tf, tests)
	if len(reasons) != 2 {
		t.Fatalf("esperaba 2 razones (missing + skipped), got %v", reasons)
	}
	// Orden estable por requirement: R2 (no corrió) antes que R3 (skip).
	if want := "R2"; len(reasons) > 0 && reasons[0][:2] != want {
		t.Errorf("primera razón debería ser de R2, got %q", reasons[0])
	}

	// Todo passed → sin razones.
	tf2 := &traceFile{Feature: "f", Requirements: map[string]traceReq{
		"R1": {Test: []string{"tests/t.py::test_r1"}},
	}}
	if got := causalityReasons(tf2, tests); len(got) != 0 {
		t.Errorf("todo passed → sin razones, got %v", got)
	}
}

// TestCheckRunJUnitReport: e2e chico — un test_cmd que ESCRIBE el JUnit XML en
// {report}; el resultado sellado trae el mapa por-test.
func TestCheckRunJUnitReport(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "constitution.json"),
		`{"schema_version":"1.0","identity_md":"x","build":{"mode":"inline",`+
			`"test_cmd":"printf '<testsuite><testcase name=\"test_a\"/></testsuite>' > {report}",`+
			`"report":"junit"}}`)

	if code := runCheckRun([]string{"--feature=demo", proj}); code != 0 {
		t.Fatalf("corrida verde → exit 0, got %d", code)
	}
	res, ok := readCheckResult(proj, "demo")
	if !ok {
		t.Fatal("debería haberse registrado un check-result")
	}
	if res.Tests["test_a"] != "pass" {
		t.Errorf("el resultado sellado debería traer test_a=pass, got %v", res.Tests)
	}
}

// TestCheckRunJUnitRequiresPlaceholder: junit sin {report} en test_cmd → exit 2.
func TestCheckRunJUnitRequiresPlaceholder(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "constitution.json"),
		`{"schema_version":"1.0","identity_md":"x","build":{"mode":"inline","test_cmd":"true","report":"junit"}}`)
	if code := runCheckRun([]string{"--feature=demo", proj}); code != 2 {
		t.Errorf("junit sin {report} → exit 2, got %d", code)
	}
}
