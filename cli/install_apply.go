package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// `sf install` (apply) + `sf uninstall` — el cut de ESCRITURAS, con disciplina:
//
//   - IDEMPOTENTE: re-correr install no duplica nada (symlink ya correcto = skip;
//     bloque marcado = se actualiza en su lugar).
//   - BACKUP antes de modificar algo que NO sea SpecForge-managed.
//   - MANIFEST de lo instalado → `sf uninstall` revierte exactamente.
//   - FAIL-SOFT por acción: un error en una acción se reporta y se sigue.
//
// Alcance de este cut: symlink (skills, plugin) + merge marker-based (AGENT.md →
// CLAUDE.md/AGENTS.md/.mdc). El `wire` (settings.json de pi/claude) se saltea con
// aviso — su merge JSON va en el cut siguiente, sobre esta misma infra.
//
// Solo se aplica a los arneses DETECTADOS (configurar lo que existe).
// ----------------------------------------------------------------------------

// Marcadores de inyección en archivos markdown. El bloque entre ellos lo posee
// `sf install`; el usuario edita AGENT.md, no el bloque.
const (
	sfBegin = "<!-- SPECFORGE:BEGIN (managed by `sf install`; edit AGENT.md instead) -->"
	sfEnd   = "<!-- SPECFORGE:END -->"
)

// installStateDir: dónde viven el manifest y los backups (bajo el base elegido).
// `.specforge-install` (con punto) para no confundir con `specforge/` (artefactos).
func installStateDir(base string) string {
	return filepath.Join(base, ".specforge-install")
}

// ── Manifest ─────────────────────────────────────────────────────────────────

type manifestEntry struct {
	Harness  string    `json:"harness"`
	Kind     string    `json:"kind"`                // symlink | merge | wire
	Target   string    `json:"target"`              // qué tocamos
	Backup   string    `json:"backup,omitempty"`    // backup del original, si lo hubo
	JSONAdds []jsonAdd `json:"json_adds,omitempty"` // solo "wire": qué agregamos (para revertir)
}

type installManifest struct {
	Version     string          `json:"version"`
	InstalledAt string          `json:"installed_at"`
	SourceRoot  string          `json:"source_root"`
	Entries     []manifestEntry `json:"entries"`
}

func (m installManifest) byTarget() map[string]manifestEntry {
	out := make(map[string]manifestEntry, len(m.Entries))
	for _, e := range m.Entries {
		out[e.Target] = e
	}
	return out
}

func loadManifest(base string) installManifest {
	var m installManifest
	if data, err := os.ReadFile(filepath.Join(installStateDir(base), "manifest.json")); err == nil {
		_ = json.Unmarshal(data, &m)
	}
	return m
}

func saveManifest(base string, m installManifest) error {
	dir := installStateDir(base)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "manifest.json"), append(out, '\n'), 0o644)
}

// ── Apply ────────────────────────────────────────────────────────────────────

func applyPlan(plans []harnessPlan, base, sourceRoot string) int {
	prev := loadManifest(base).byTarget()
	merged := map[string]manifestEntry{} // target → entry final (preserva prev)
	for t, e := range prev {
		merged[t] = e
	}
	applied := 0

	for _, p := range plans {
		if !p.detected {
			fmt.Printf("%-12s skipped (not detected)\n", p.name)
			continue
		}
		fmt.Printf("%-12s\n", p.name)
		for _, a := range p.actions {
			var e manifestEntry
			var err error
			switch a.kind {
			case "symlink":
				e, err = applySymlink(p.name, a.source, a.target, base, prev[a.target])
			case "merge":
				e, err = applyMerge(p.name, a.source, a.target, base, prev[a.target])
			case "wire":
				e, err = applyWire(p.name, a.target, a.jsonAdds, base, prev[a.target])
			default: // note
				continue
			}
			if err != nil {
				fmt.Printf("  %-7s FAILED %s (%v)\n", a.kind, a.target, err)
				continue
			}
			merged[e.Target] = e
			applied++
		}
		fmt.Println()
	}

	// Persistimos el manifest (unión de lo previo + lo de esta corrida).
	m := installManifest{Version: "1", InstalledAt: time.Now().UTC().Format(time.RFC3339), SourceRoot: sourceRoot}
	for _, e := range merged {
		m.Entries = append(m.Entries, e)
	}
	if err := saveManifest(base, m); err != nil {
		fmt.Fprintf(os.Stderr, "sf install: could not write manifest (%v)\n", err)
		return 1
	}
	fmt.Printf("Applied %d item(s). Manifest + backups in %s\nUndo with `sf uninstall%s`.\n",
		applied, tildeHomeOr(base, installStateDir(base)), globalSuffix(base))
	return 0
}

// applySymlink crea target → source. Idempotente; respalda (por rename) lo que
// hubiera ahí y no sea ya nuestro symlink.
func applySymlink(harness, source, target, base string, prev manifestEntry) (manifestEntry, error) {
	e := manifestEntry{Harness: harness, Kind: "symlink", Target: target, Backup: prev.Backup}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return e, err
	}
	if cur, err := os.Readlink(target); err == nil && cur == source {
		fmt.Printf("  symlink ok (already linked) %s\n", tildeHomeOr(base, target))
		return e, nil
	}
	if lexists(target) {
		if e.Backup == "" {
			b, err := backupByRename(target, base)
			if err != nil {
				return e, err
			}
			e.Backup = b
		} else {
			os.RemoveAll(target) // ya teníamos el original respaldado; sacamos lo actual
		}
	}
	if err := os.Symlink(source, target); err != nil {
		return e, err
	}
	fmt.Printf("  symlink → %s\n", tildeHomeOr(base, target))
	return e, nil
}

// applyMerge inyecta el contenido de source (AGENT.md) en target entre marcadores.
// Idempotente: si ya hay un bloque marcado, lo reemplaza; si el archivo existe sin
// marcadores, respalda (copia) y apendea; si no existe, lo crea.
func applyMerge(harness, source, target, base string, prev manifestEntry) (manifestEntry, error) {
	e := manifestEntry{Harness: harness, Kind: "merge", Target: target, Backup: prev.Backup}
	content, err := os.ReadFile(source)
	if err != nil {
		return e, err
	}
	block := sfBegin + "\n\n" + string(content) + "\n\n" + sfEnd + "\n"

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return e, err
	}
	if !lexists(target) {
		if err := os.WriteFile(target, []byte(block), 0o644); err != nil {
			return e, err
		}
		fmt.Printf("  merge   → created %s\n", tildeHomeOr(base, target))
		return e, nil
	}

	existing, err := os.ReadFile(target)
	if err != nil {
		return e, err
	}
	if strings.Contains(string(existing), sfBegin) {
		updated := replaceBlock(string(existing), block)
		if err := os.WriteFile(target, []byte(updated), 0o644); err != nil {
			return e, err
		}
		fmt.Printf("  merge   ok (updated block) %s\n", tildeHomeOr(base, target))
		return e, nil
	}

	// Existe sin marcadores: respaldamos (copia) y apendeamos.
	if e.Backup == "" {
		b, err := backupByCopy(target, base)
		if err != nil {
			return e, err
		}
		e.Backup = b
	}
	combined := strings.TrimRight(string(existing), "\n") + "\n\n" + block
	if err := os.WriteFile(target, []byte(combined), 0o644); err != nil {
		return e, err
	}
	fmt.Printf("  merge   → appended block (backup saved) %s\n", tildeHomeOr(base, target))
	return e, nil
}

// replaceBlock reemplaza la región sfBegin..sfEnd por block, preservando lo de
// afuera. Si por algo falta el cierre, apendea el block nuevo.
func replaceBlock(existing, block string) string {
	i := strings.Index(existing, sfBegin)
	if i < 0 {
		return strings.TrimRight(existing, "\n") + "\n\n" + block
	}
	rest := existing[i:]
	j := strings.Index(rest, sfEnd)
	if j < 0 {
		return existing[:i] + block // bloque corrupto: lo reescribimos limpio
	}
	after := rest[j+len(sfEnd):]
	return existing[:i] + strings.TrimRight(block, "\n") + after
}

// applyWire agrega pares clave→valor a un JSON (settings.json de pi): para cada
// add, apendea Value al array bajo Key (sin duplicar). Idempotente; respalda el
// archivo original (copia) la primera vez. Si el settings.json existe pero NO es
// JSON válido, abortamos esta acción (no lo pisamos).
func applyWire(harness, target string, adds []jsonAdd, base string, prev manifestEntry) (manifestEntry, error) {
	e := manifestEntry{Harness: harness, Kind: "wire", Target: target, Backup: prev.Backup, JSONAdds: adds}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return e, err
	}

	obj := map[string]any{}
	existed := lexists(target)
	if existed {
		data, err := os.ReadFile(target)
		if err != nil {
			return e, err
		}
		if len(strings.TrimSpace(string(data))) > 0 {
			if err := json.Unmarshal(data, &obj); err != nil {
				return e, fmt.Errorf("%s is not valid JSON; left untouched", target)
			}
		}
		if e.Backup == "" {
			b, err := backupByCopy(target, base)
			if err != nil {
				return e, err
			}
			e.Backup = b
		}
	}

	for _, ad := range adds {
		addToJSONArray(obj, ad.Key, ad.Value)
	}
	out, err := json.MarshalIndent(obj, "", "  ")
	if err != nil {
		return e, err
	}
	if err := os.WriteFile(target, append(out, '\n'), 0o644); err != nil {
		return e, err
	}
	fmt.Printf("  wire    → %s\n", tildeHomeOr(base, target))
	return e, nil
}

// addToJSONArray agrega value al array bajo key (creándolo si falta), sin
// duplicar. Si key existe y NO es un array, no lo toca (seguridad).
func addToJSONArray(obj map[string]any, key, value string) {
	cur, ok := obj[key]
	if !ok || cur == nil {
		obj[key] = []any{value}
		return
	}
	arr, ok := cur.([]any)
	if !ok {
		return // ya hay algo no-array bajo esa clave: no lo pisamos
	}
	for _, v := range arr {
		if s, ok := v.(string); ok && s == value {
			return // ya está
		}
	}
	obj[key] = append(arr, value)
}

// removeFromJSONArray quita value del array bajo key; si queda vacío, borra la
// clave. Lo usa uninstall.
func removeFromJSONArray(obj map[string]any, key, value string) {
	arr, ok := obj[key].([]any)
	if !ok {
		return
	}
	var kept []any
	for _, v := range arr {
		if s, ok := v.(string); ok && s == value {
			continue
		}
		kept = append(kept, v)
	}
	if len(kept) == 0 {
		delete(obj, key)
		return
	}
	obj[key] = kept
}

// ── Backups ──────────────────────────────────────────────────────────────────

// backupDir: subcarpeta de backups con timestamp único de la sesión de install.
func backupDir(base string) string {
	return filepath.Join(installStateDir(base), "backups")
}

// sanitizeForBackup convierte una ruta absoluta en un nombre de archivo plano.
func sanitizeForBackup(target string) string {
	s := strings.TrimPrefix(filepath.ToSlash(target), "/")
	return strings.ReplaceAll(s, "/", "_")
}

// backupByCopy copia target a la carpeta de backups (deja el original en su lugar).
func backupByCopy(target, base string) (string, error) {
	dir := backupDir(base)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, sanitizeForBackup(target))
	if err := copyFile(target, dst); err != nil {
		return "", err
	}
	return dst, nil
}

// backupByRename mueve target a la carpeta de backups (lo saca de su lugar).
func backupByRename(target, base string) (string, error) {
	dir := backupDir(base)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	dst := filepath.Join(dir, sanitizeForBackup(target))
	if err := os.Rename(target, dst); err != nil {
		// fallback cross-device: copiar + borrar
		if cerr := copyFile(target, dst); cerr != nil {
			return "", err
		}
		os.RemoveAll(target)
	}
	return dst, nil
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}

// ── Uninstall ────────────────────────────────────────────────────────────────

func runUninstall(args []string) int {
	global := false
	projectDir := "."
	for _, a := range args {
		switch {
		case a == "--global":
			global = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf uninstall: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	home, _ := os.UserHomeDir()
	base := home
	if !global {
		base, _ = filepath.Abs(projectDir)
	}

	m := loadManifest(base)
	if len(m.Entries) == 0 {
		fmt.Printf("Nothing installed under %s.\n", tildeHomeOr(base, base))
		return 0
	}

	removed := 0
	for _, e := range m.Entries {
		if err := revertEntry(e); err != nil {
			fmt.Printf("  FAILED revert %s (%v)\n", e.Target, err)
			continue
		}
		fmt.Printf("  reverted %s\n", tildeHomeOr(base, e.Target))
		removed++
	}

	// Limpiamos el estado de install (manifest + backups consumidos).
	os.RemoveAll(installStateDir(base))
	fmt.Printf("Uninstalled %d item(s).\n", removed)
	return 0
}

// revertEntry deshace una entrada del manifest.
func revertEntry(e manifestEntry) error {
	switch e.Kind {
	case "symlink":
		if lexists(e.Target) {
			if err := os.Remove(e.Target); err != nil {
				return err
			}
		}
		if e.Backup != "" { // restauramos el original que habíamos movido
			return os.Rename(e.Backup, e.Target)
		}
		return nil
	case "merge":
		if e.Backup != "" { // restauramos el archivo original exacto
			return copyFile(e.Backup, e.Target)
		}
		// Lo creamos nosotros: quitamos el bloque; si queda vacío, borramos.
		data, err := os.ReadFile(e.Target)
		if err != nil {
			return nil // ya no está: nada que hacer
		}
		stripped := stripBlock(string(data))
		if strings.TrimSpace(stripped) == "" {
			return os.Remove(e.Target)
		}
		return os.WriteFile(e.Target, []byte(stripped), 0o644)
	case "wire":
		if e.Backup != "" { // restauramos el settings.json original exacto
			return copyFile(e.Backup, e.Target)
		}
		// Lo creamos nosotros: quitamos nuestros valores; si queda {} borramos.
		data, err := os.ReadFile(e.Target)
		if err != nil {
			return nil
		}
		obj := map[string]any{}
		if json.Unmarshal(data, &obj) != nil {
			return nil // ya no es nuestro/JSON: no tocar
		}
		for _, ad := range e.JSONAdds {
			removeFromJSONArray(obj, ad.Key, ad.Value)
		}
		if len(obj) == 0 {
			return os.Remove(e.Target)
		}
		out, _ := json.MarshalIndent(obj, "", "  ")
		return os.WriteFile(e.Target, append(out, '\n'), 0o644)
	}
	return nil
}

// stripBlock quita la región sfBegin..sfEnd (y el separador previo) del texto.
func stripBlock(s string) string {
	i := strings.Index(s, sfBegin)
	if i < 0 {
		return s
	}
	rest := s[i:]
	j := strings.Index(rest, sfEnd)
	if j < 0 {
		return strings.TrimRight(s[:i], "\n") + "\n"
	}
	after := rest[j+len(sfEnd):]
	return strings.TrimRight(s[:i], "\n") + strings.TrimLeft(after, "\n")
}

// ── Helpers ──────────────────────────────────────────────────────────────────

// lexists: ¿existe el path SIN seguir symlinks? (Lstat, para detectar symlinks
// rotos o el propio symlink).
func lexists(path string) bool {
	_, err := os.Lstat(path)
	return err == nil
}

// tildeHomeOr acorta bajo $HOME a ~/… (envuelve tildeHome resolviendo el home).
func tildeHomeOr(base, p string) string {
	home, _ := os.UserHomeDir()
	return tildeHome(home, p)
}

// globalSuffix sugiere el flag correcto para uninstall según el base.
func globalSuffix(base string) string {
	home, _ := os.UserHomeDir()
	if base == home {
		return " --global"
	}
	return ""
}
