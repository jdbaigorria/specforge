package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestLogEventAppends: los eventos se apendean como JSONL y se leen de vuelta
// con sus campos; fuera de un proyecto SpecForge no se escribe nada.
func TestLogEventAppends(t *testing.T) {
	proj := t.TempDir()

	// Sin specforge/ → no-op (mismo guard que el hook: nunca ensuciar un
	// proyecto ajeno).
	logEvent(proj, sfEvent{Kind: "deny", Target: "x"})
	if _, err := os.Stat(eventsPath(proj)); err == nil {
		t.Fatal("sin specforge/ no debería escribirse telemetría")
	}

	mustWrite(t, filepath.Join(proj, "specforge", "features.json"), `{"features":[]}`)
	logEvent(proj, sfEvent{Kind: "deny", Target: "specforge/features.json", Detail: "denied"})
	logEvent(proj, sfEvent{Kind: "verdict", Feature: "f", Phase: "design", Detail: "fail"})

	events := readEvents(proj)
	if len(events) != 2 {
		t.Fatalf("esperaba 2 eventos, hay %d", len(events))
	}
	if events[0].Kind != "deny" || events[0].At == "" {
		t.Errorf("el evento debe llevar kind y timestamp, got %+v", events[0])
	}
	if events[1].Feature != "f" || events[1].Detail != "fail" {
		t.Errorf("evento de verdict mal registrado: %+v", events[1])
	}
}

// TestPruneEvents: pasado el cap, se conserva la mitad MÁS RECIENTE.
func TestPruneEvents(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"), `{"features":[]}`)

	// Detail grande para superar el cap con pocas líneas… no: logEvent trunca
	// Detail. Escribimos el archivo directo (pruneEvents opera sobre el path).
	path := eventsPath(proj)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	var b strings.Builder
	for i := 0; b.Len() <= eventsMaxBytes; i++ {
		fmt.Fprintf(&b, `{"at":"t%d","kind":"deny","target":"x"}`+"\n", i)
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		t.Fatal(err)
	}
	before := len(readEvents(proj))

	pruneEvents(path)
	after := readEvents(proj)
	if len(after) >= before {
		t.Fatalf("la poda debería reducir las líneas (%d → %d)", before, len(after))
	}
	// Se conserva la cola (los más recientes): el último evento sigue siendo el último.
	if got := after[len(after)-1].At; got != fmt.Sprintf("t%d", before-1) {
		t.Errorf("la poda debe conservar los eventos MÁS RECIENTES, último=%s", got)
	}
}

// TestDenyEvent: target = ruta relativa para Write/Edit, comando recortado para Bash.
func TestDenyEvent(t *testing.T) {
	proj := t.TempDir()
	e := denyEvent(proj, filepath.Join(proj, "specforge", "features.json"), "", "s1", "reason")
	if e.Target != "specforge/features.json" || e.Kind != "deny" {
		t.Errorf("Write/Edit deny debería registrar la ruta relativa, got %+v", e)
	}
	e = denyEvent(proj, "", "echo x > specforge/features.json", "s1", "reason")
	if !strings.HasPrefix(e.Target, "echo x >") {
		t.Errorf("Bash deny debería registrar el comando, got %q", e.Target)
	}
}

// TestRunEventsSummary: el comando agrega por kind y no explota con log vacío.
func TestRunEventsSummary(t *testing.T) {
	proj := t.TempDir()
	if code := runEvents([]string{proj}); code != 0 {
		t.Errorf("sin eventos → exit 0, got %d", code)
	}
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"), `{"features":[]}`)
	logEvent(proj, sfEvent{Kind: "deny", Target: "specforge/features.json"})
	if code := runEvents([]string{"--tail=5", proj}); code != 0 {
		t.Errorf("con eventos → exit 0, got %d", code)
	}
	if code := runEvents([]string{"--json", proj}); code != 0 {
		t.Errorf("--json → exit 0, got %d", code)
	}
	if code := runEvents([]string{"--tail=nope", proj}); code != 2 {
		t.Errorf("--tail inválido → exit 2, got %d", code)
	}
}
