package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestReplaceBlock: actualizar el bloque marcado preserva lo de afuera.
func TestReplaceBlock(t *testing.T) {
	existing := "before\n\n" + sfBegin + "\nOLD\n" + sfEnd + "\n\nafter\n"
	block := sfBegin + "\nNEW\n" + sfEnd + "\n"
	got := replaceBlock(existing, block)
	if !strings.Contains(got, "before") || !strings.Contains(got, "after") {
		t.Errorf("replaceBlock perdió el texto de afuera: %q", got)
	}
	if strings.Contains(got, "OLD") || !strings.Contains(got, "NEW") {
		t.Errorf("replaceBlock no reemplazó el contenido: %q", got)
	}
	if strings.Count(got, sfBegin) != 1 {
		t.Errorf("replaceBlock duplicó el marcador: %q", got)
	}
}

// TestStripBlock: quitar el bloque deja el resto intacto.
func TestStripBlock(t *testing.T) {
	s := "keep me\n\n" + sfBegin + "\nstuff\n" + sfEnd + "\n"
	got := stripBlock(s)
	if strings.Contains(got, sfBegin) || strings.Contains(got, "stuff") {
		t.Errorf("stripBlock no quitó el bloque: %q", got)
	}
	if !strings.Contains(got, "keep me") {
		t.Errorf("stripBlock se comió el contenido del usuario: %q", got)
	}
}

// TestApplyMergeRoundTrip: merge sobre un archivo existente respalda + apendea;
// re-merge es idempotente; revert restaura el original EXACTO.
func TestApplyMergeRoundTrip(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "AGENT.md")
	target := filepath.Join(base, "CLAUDE.md")
	original := "# My rules\nkeep tidy\n"
	mustWrite(t, src, "# AGENT\nv1\n")
	mustWrite(t, target, original)

	e, err := applyMerge("claude-code", src, target, base, manifestEntry{})
	if err != nil {
		t.Fatal(err)
	}
	got := mustRead(t, target)
	if !strings.Contains(got, "keep tidy") || !strings.Contains(got, sfBegin) {
		t.Errorf("merge debería preservar el original y agregar el bloque: %q", got)
	}
	if e.Backup == "" {
		t.Error("merge sobre archivo existente debería respaldar")
	}

	// Re-merge idempotente: un solo bloque, mismo backup.
	e2, _ := applyMerge("claude-code", src, target, base, e)
	if strings.Count(mustRead(t, target), sfBegin) != 1 {
		t.Error("re-merge duplicó el bloque")
	}
	if e2.Backup != e.Backup {
		t.Error("re-merge no debería cambiar el backup")
	}

	// Revert restaura el original exacto.
	if err := revertEntry(e2); err != nil {
		t.Fatal(err)
	}
	if mustRead(t, target) != original {
		t.Errorf("revert no restauró el original: %q", mustRead(t, target))
	}
}

// TestApplyMergeCreateThenRevert: si el target no existía, lo creamos solo con el
// bloque; revert lo borra (no dejamos basura).
func TestApplyMergeCreateThenRevert(t *testing.T) {
	base := t.TempDir()
	src := filepath.Join(base, "AGENT.md")
	target := filepath.Join(base, "sub", "AGENTS.md")
	mustWrite(t, src, "# AGENT\n")

	e, err := applyMerge("opencode", src, target, base, manifestEntry{})
	if err != nil {
		t.Fatal(err)
	}
	if e.Backup != "" {
		t.Error("crear un archivo nuevo no debería generar backup")
	}
	if !lexists(target) {
		t.Fatal("el target debería existir tras el merge")
	}
	if err := revertEntry(e); err != nil {
		t.Fatal(err)
	}
	if lexists(target) {
		t.Error("revert de un archivo que creamos debería borrarlo")
	}
}

// TestApplySymlinkRoundTrip: symlink sobre un archivo existente lo respalda (move);
// revert quita el symlink y restaura el original.
func TestApplySymlinkRoundTrip(t *testing.T) {
	base := t.TempDir()
	source := filepath.Join(base, "skills")
	if err := os.MkdirAll(source, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(base, ".claude", "skills")
	mustWrite(t, target, "preexisting file content")

	e, err := applySymlink("claude-code", source, target, base, manifestEntry{})
	if err != nil {
		t.Fatal(err)
	}
	if dst, _ := os.Readlink(target); dst != source {
		t.Errorf("el symlink debería apuntar a la fuente, got %q", dst)
	}
	if e.Backup == "" {
		t.Error("symlink sobre algo existente debería respaldarlo")
	}

	if err := revertEntry(e); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Readlink(target); err == nil {
		t.Error("revert debería quitar el symlink")
	}
	if mustRead(t, target) != "preexisting file content" {
		t.Error("revert debería restaurar el archivo original")
	}
}

// TestJSONArrayHelpers: add no duplica y crea; remove filtra y borra clave vacía.
func TestJSONArrayHelpers(t *testing.T) {
	obj := map[string]any{}
	addToJSONArray(obj, "skills", "/a")
	addToJSONArray(obj, "skills", "/a") // duplicado: no debe agregarse
	addToJSONArray(obj, "skills", "/b")
	if arr := obj["skills"].([]any); len(arr) != 2 {
		t.Errorf("skills debería tener 2, got %v", arr)
	}
	// no pisa un valor no-array
	obj["theme"] = "dark"
	addToJSONArray(obj, "theme", "x")
	if obj["theme"] != "dark" {
		t.Error("addToJSONArray no debería pisar una clave no-array")
	}
	removeFromJSONArray(obj, "skills", "/a")
	removeFromJSONArray(obj, "skills", "/b")
	if _, ok := obj["skills"]; ok {
		t.Error("removeFromJSONArray debería borrar la clave cuando queda vacía")
	}
}

// TestApplyWireRoundTrip: wire sobre settings.json existente preserva lo del
// usuario, agrega lo nuestro, es idempotente, y revert restaura el original.
func TestApplyWireRoundTrip(t *testing.T) {
	base := t.TempDir()
	target := filepath.Join(base, ".pi", "settings.json")
	original := "{\n  \"theme\": \"dark\"\n}\n"
	mustWrite(t, target, original)
	adds := []jsonAdd{{Key: "skills", Value: "/s"}, {Key: "extensions", Value: "/e"}}

	e, err := applyWire("pi", target, adds, base, manifestEntry{})
	if err != nil {
		t.Fatal(err)
	}
	got := mustRead(t, target)
	if !strings.Contains(got, "dark") || !strings.Contains(got, "/s") || !strings.Contains(got, "/e") {
		t.Errorf("wire debería preservar theme y agregar skills+extensions: %q", got)
	}
	if e.Backup == "" {
		t.Error("wire sobre archivo existente debería respaldar")
	}

	// Idempotente.
	if _, err := applyWire("pi", target, adds, base, e); err != nil {
		t.Fatal(err)
	}
	if strings.Count(mustRead(t, target), "/s") != 1 {
		t.Error("re-wire duplicó el valor")
	}

	// Revert restaura el original.
	if err := revertEntry(e); err != nil {
		t.Fatal(err)
	}
	if mustRead(t, target) != original {
		t.Errorf("revert no restauró el original: %q", mustRead(t, target))
	}
}

// TestApplyWireCreateThenRevert: si no existía, lo creamos; revert lo borra
// (quedó solo con nuestras claves).
func TestApplyWireCreateThenRevert(t *testing.T) {
	base := t.TempDir()
	target := filepath.Join(base, ".pi", "settings.json")
	adds := []jsonAdd{{Key: "extensions", Value: "/e"}}

	e, err := applyWire("pi", target, adds, base, manifestEntry{})
	if err != nil {
		t.Fatal(err)
	}
	if e.Backup != "" {
		t.Error("crear settings.json no debería respaldar")
	}
	if err := revertEntry(e); err != nil {
		t.Fatal(err)
	}
	if lexists(target) {
		t.Error("revert debería borrar el settings.json que creamos")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func mustRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
