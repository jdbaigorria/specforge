package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCheckJournal: feature + al menos una lección con context y rule.
func TestCheckJournal(t *testing.T) {
	ok := journalEntry{Feature: "trunc", Lessons: []journalLesson{
		{Context: "se rompió el render con tags vacíos", Rule: "siempre guardar tags como [] no null"},
	}}
	var rep report
	checkJournal(ok, &rep)
	if len(rep.errors) != 0 {
		t.Errorf("entrada válida con errores: %v", rep.errors)
	}

	// Sin feature, sin lecciones, y una lección incompleta.
	bad := journalEntry{Lessons: []journalLesson{{Context: "algo"}}} // falta rule
	rep = report{}
	checkJournal(bad, &rep)
	if len(rep.errors) < 2 { // feature faltante + rule faltante
		t.Errorf("errors=%v, want al menos 2 (feature + rule)", rep.errors)
	}

	none := journalEntry{Feature: "x"}
	rep = report{}
	checkJournal(none, &rep)
	if len(rep.errors) != 1 {
		t.Errorf("sin lecciones: errors=%v, want 1 (at least one lesson)", rep.errors)
	}
}

// TestJournalAdd: escribe json + md bajo specforge/journal/ y rinde el contenido.
// En un dir temporal (no es repo git) gitStage devuelve false, pero los archivos
// se escriben igual y el exit es 0.
func TestJournalAdd(t *testing.T) {
	dir := t.TempDir()
	entry := journalEntry{
		Feature: "trunc",
		Date:    "2026-06-19",
		Lessons: []journalLesson{
			{Context: "el slice de wave dependía de tasks.json", Rule: "derivar estado, no guardarlo", Tags: []string{"f2", "estado"}},
		},
	}
	if code := journalAdd(dir, entry); code != 0 {
		t.Fatalf("journalAdd exit=%d, want 0", code)
	}

	base := filepath.Join(dir, "specforge", "journal", "2026-06-19-trunc")
	if _, err := os.Stat(base + ".json"); err != nil {
		t.Errorf("falta el .json: %v", err)
	}
	md, err := os.ReadFile(base + ".md")
	if err != nil {
		t.Fatalf("falta el .md: %v", err)
	}
	// El render debe traer la regla, el contexto y los tags.
	for _, want := range []string{"derivar estado, no guardarlo", "el slice de wave", "f2"} {
		if !strings.Contains(string(md), want) {
			t.Errorf("md no contiene %q\n%s", want, md)
		}
	}
}
