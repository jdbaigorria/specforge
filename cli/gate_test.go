package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestRunGateStatus verifica los exit codes: 0 cuando hay datos o cuando no hay
// features.json (ausencia no es error), y 4 cuando se pide una feature que no
// existe. (writeFile vive en doctor_test.go, mismo paquete.)
func TestRunGateStatus(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"slugify","status":"done","gates":[{"phase":"requirements","result":"approve","by":"user"}]}]}`)

	t.Run("known feature -> 0", func(t *testing.T) {
		if code := runGateStatus(dir, "slugify"); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
	t.Run("unknown feature -> 4", func(t *testing.T) {
		if code := runGateStatus(dir, "nope"); code != 4 {
			t.Errorf("exit=%d, want 4", code)
		}
	})
	t.Run("summary (no feature) -> 0", func(t *testing.T) {
		if code := runGateStatus(dir, ""); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
	t.Run("no features.json -> 0", func(t *testing.T) {
		if code := runGateStatus(t.TempDir(), ""); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})
}

// TestBuildAuditEntry: overall = fail si alguna regla falla; valida campos.
func TestBuildAuditEntry(t *testing.T) {
	// Todo pass → overall pass.
	e, rep := buildAuditEntry("design", []ruleVerdict{
		{Rule: "P1", Result: "pass", Citation: "ok"},
	})
	if len(rep.errors) != 0 || e.Overall != "pass" {
		t.Errorf("overall=%q errors=%v, want pass / sin errores", e.Overall, rep.errors)
	}
	if e.At == "" {
		t.Errorf("falta el timestamp At")
	}

	// Una falla → overall fail.
	e, _ = buildAuditEntry("design", []ruleVerdict{
		{Rule: "P1", Result: "pass"}, {Rule: "I1", Result: "fail"},
	})
	if e.Overall != "fail" {
		t.Errorf("overall=%q, want fail", e.Overall)
	}

	// Validaciones: phase vacía, sin verdicts, result inválido.
	if _, rep := buildAuditEntry("", nil); len(rep.errors) < 2 {
		t.Errorf("errors=%v, want phase + at-least-one", rep.errors)
	}
	if _, rep := buildAuditEntry("design", []ruleVerdict{{Rule: "P1", Result: "maybe"}}); len(rep.errors) != 1 {
		t.Errorf("errors=%v, want 1 (result inválido)", rep.errors)
	}
}

// TestVerifyCitations (A8): una citation que existe en el artefacto queda
// "verified"; una inventada queda "not-found"; sin citation no se marca nada.
// Es la regla de integridad aplicada al juez: sus afirmaciones se verifican
// mecánicamente o quedan marcadas como no-verificadas.
func TestVerifyCitations(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "design.json"),
		`{"feature":"f","components":[{"name":"parser","purpose":"parses EARS requirements"}]}`)

	entry := auditEntry{Phase: "design", Verdicts: []ruleVerdict{
		{Rule: "P1", Result: "pass", Citation: "parses EARS requirements"},
		{Rule: "P2", Result: "pass", Citation: "this text exists nowhere"},
		{Rule: "P3", Result: "pass"}, // sin citation → sin marca
	}}
	verifyCitations(proj, "f", &entry)

	if got := entry.Verdicts[0].CitationCheck; got != "verified" {
		t.Errorf("citation real → verified, got %q", got)
	}
	if got := entry.Verdicts[1].CitationCheck; got != "not-found" {
		t.Errorf("citation inventada → not-found, got %q", got)
	}
	if got := entry.Verdicts[2].CitationCheck; got != "" {
		t.Errorf("sin citation → sin marca, got %q", got)
	}

	// Whitespace-normalizado: la misma citation partida en líneas matchea igual.
	entry = auditEntry{Phase: "design", Verdicts: []ruleVerdict{
		{Rule: "P1", Result: "pass", Citation: "parses\n  EARS   requirements"},
	}}
	verifyCitations(proj, "f", &entry)
	if got := entry.Verdicts[0].CitationCheck; got != "verified" {
		t.Errorf("citation con whitespace distinto → verified, got %q", got)
	}

	// Fase sin artefacto (wave-1) → no marca nada.
	entry = auditEntry{Phase: "wave-1", Verdicts: []ruleVerdict{
		{Rule: "P1", Result: "pass", Citation: "whatever"},
	}}
	verifyCitations(proj, "f", &entry)
	if got := entry.Verdicts[0].CitationCheck; got != "" {
		t.Errorf("fase sin artefacto → sin marca, got %q", got)
	}
}

// TestConsecutiveFails (D3'): cuenta la racha de fails de UNA fase; un pass la
// corta; otras fases no interfieren.
func TestConsecutiveFails(t *testing.T) {
	proj := t.TempDir()
	seed := func(phase, overall string) {
		e, _ := buildAuditEntry(phase, []ruleVerdict{{Rule: "P1", Result: map[bool]string{true: "fail", false: "pass"}[overall == "fail"]}})
		if code := appendAuditEntry(proj, "f", e); code != 0 {
			t.Fatal("seed")
		}
	}
	seed("design", "fail")
	seed("tasks", "pass") // otra fase: no corta la racha de design
	seed("design", "fail")
	seed("design", "fail")
	if n := consecutiveFails(proj, "f", "design"); n != 3 {
		t.Errorf("racha de design = %d, want 3", n)
	}
	seed("design", "pass")
	if n := consecutiveFails(proj, "f", "design"); n != 0 {
		t.Errorf("un pass corta la racha, got %d", n)
	}
}

// TestAppendAuditEntry: append-only — dos entradas quedan acumuladas en audit.json.
func TestAppendAuditEntry(t *testing.T) {
	dir := t.TempDir()
	e1, _ := buildAuditEntry("design", []ruleVerdict{{Rule: "P1", Result: "pass"}})
	e2, _ := buildAuditEntry("tasks", []ruleVerdict{{Rule: "P2", Result: "fail"}})
	if code := appendAuditEntry(dir, "trunc", e1); code != 0 {
		t.Fatalf("append #1 exit=%d", code)
	}
	if code := appendAuditEntry(dir, "trunc", e2); code != 0 {
		t.Fatalf("append #2 exit=%d", code)
	}

	data, err := os.ReadFile(filepath.Join(dir, "specforge", "features", "trunc", "audit.json"))
	if err != nil {
		t.Fatal(err)
	}
	var ledger auditLedger
	if err := json.Unmarshal(data, &ledger); err != nil {
		t.Fatal(err)
	}
	if ledger.Feature != "trunc" || len(ledger.Entries) != 2 {
		t.Errorf("ledger feature=%q entries=%d, want trunc / 2", ledger.Feature, len(ledger.Entries))
	}
	if ledger.Entries[1].Overall != "fail" {
		t.Errorf("2da entrada overall=%q, want fail", ledger.Entries[1].Overall)
	}
}
