package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf onboard scan` — inventario determinista para brownfield (A4, §2.9).
//
// El problema: onboardear SpecForge sobre un repo existente hoy significa que
// el LLM "explora" — lee archivos a discreción, alucina estructura, quema
// tokens. El incidente FIXBUGHIGH fue en brownfield y no fue casualidad: con
// 80% del código preexistente, "todo parece hecho" y el costo de fingir baja.
//
// La respuesta: el CLI computa el inventario (árbol, símbolos top-level por
// archivo, mapa test→módulo) y el LLM INTERPRETA sobre ese inventario. Menos
// alucinación, menos tokens, y una base estable para `--from-code`.
//
// Salida: specforge/context/inventory.json (+ .md legible). Determinista: dos
// corridas sobre el mismo repo dan el mismo inventario.
// ----------------------------------------------------------------------------

type inventory struct {
	SchemaVersion string         `json:"schema_version"`
	At            string         `json:"at"`
	Files         int            `json:"files"`     // archivos de código (tests incluidos)
	ByLang        map[string]int `json:"by_lang"`   // extensión → cantidad
	Modules       []moduleInfo   `json:"modules"`   // símbolos top-level por archivo
	TestMap       map[string][]string `json:"test_map"` // test → módulo(s) que parece cubrir
}

type moduleInfo struct {
	Path    string   `json:"path"`
	IsTest  bool     `json:"is_test,omitempty"`
	Symbols []string `json:"symbols,omitempty"`
}

// codeExtensions: qué consideramos "código" para el inventario y la cobertura.
var codeExtensions = map[string]bool{
	".go": true, ".py": true, ".js": true, ".ts": true, ".jsx": true, ".tsx": true,
	".rb": true, ".rs": true, ".java": true, ".kt": true, ".cs": true, ".php": true,
	".c": true, ".h": true, ".cpp": true, ".hpp": true,
}

// symbolPatterns: extracción de símbolos top-level por lenguaje. Regex simples
// a propósito — el inventario es un MAPA, no un AST; alcanza con los nombres
// que un trace usaría como anchor (path:symbol).
var symbolPatterns = map[string][]*regexp.Regexp{
	".go": {
		regexp.MustCompile(`(?m)^func (\w+)`),
		regexp.MustCompile(`(?m)^func \([^)]+\) (\w+)`),
		regexp.MustCompile(`(?m)^type (\w+)`),
	},
	".py": {
		regexp.MustCompile(`(?m)^(?:async )?def (\w+)`),
		regexp.MustCompile(`(?m)^class (\w+)`),
	},
	".js":  jsPatterns, ".jsx": jsPatterns, ".ts": jsPatterns, ".tsx": jsPatterns,
	".rb": {
		regexp.MustCompile(`(?m)^\s*def (\w+)`),
		regexp.MustCompile(`(?m)^\s*class (\w+)`),
		regexp.MustCompile(`(?m)^\s*module (\w+)`),
	},
	".rs": {
		regexp.MustCompile(`(?m)^(?:pub )?fn (\w+)`),
		regexp.MustCompile(`(?m)^(?:pub )?struct (\w+)`),
		regexp.MustCompile(`(?m)^(?:pub )?enum (\w+)`),
		regexp.MustCompile(`(?m)^(?:pub )?trait (\w+)`),
	},
}

var jsPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?m)^(?:export )?(?:default )?(?:async )?function (\w+)`),
	regexp.MustCompile(`(?m)^(?:export )?(?:default )?class (\w+)`),
	regexp.MustCompile(`(?m)^export const (\w+)`),
}

// maxSymbolsPerFile: cap para que un archivo generado gigante no infle el
// inventario (el resto se resume con un "+N more").
const maxSymbolsPerFile = 60

func runOnboard(args []string) int {
	if len(args) == 0 || args[0] != "scan" {
		fmt.Fprintln(os.Stderr, "usage: sf onboard scan [--json] [project_dir]")
		return 2
	}
	args = args[1:]
	projectDir := "."
	asJSON := false
	for _, a := range args {
		switch {
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf onboard: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	inv := scanInventory(projectDir)

	if asJSON {
		out, _ := json.MarshalIndent(inv, "", "  ")
		fmt.Println(string(out))
		return 0
	}

	// Persistimos bajo context/ — es CONOCIMIENTO computado del proyecto, la
	// base sobre la que sf-propose --from-code interpreta.
	base := filepath.Join(projectDir, "specforge", "context", "inventory")
	if err := os.MkdirAll(filepath.Dir(base), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf onboard: %v\n", err)
		return 1
	}
	out, err := json.MarshalIndent(inv, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf onboard: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(base+".json", append(out, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf onboard: cannot write inventory (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(base+".md", []byte(renderInventoryMarkdown(inv)), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf onboard: cannot write inventory.md (%v)\n", err)
		return 1
	}

	fmt.Printf("scanned %d code file(s) → %s.json (+ .md)\n", inv.Files, filepath.ToSlash(base))
	fmt.Printf("languages: %s · %d test file(s) mapped\n", langSummary(inv.ByLang), len(inv.TestMap))
	fmt.Println("Interpret over the inventory (read inventory.md) instead of free-exploring the repo.")
	return 0
}

// scanInventory computa el inventario: lista de archivos (git si hay, walk si
// no — misma base que codeHash), símbolos top-level y mapa test→módulo.
func scanInventory(projectDir string) inventory {
	inv := inventory{
		SchemaVersion: schemaVersionCurrent,
		At:            nowUTC(),
		ByLang:        map[string]int{},
		TestMap:       map[string][]string{},
	}

	files, ok := gitListFiles(projectDir)
	if !ok {
		files = walkListFiles(projectDir)
	}
	sort.Strings(files) // orden estable → inventario reproducible

	// Índice basename-sin-marcadores → paths, para casar tests con módulos.
	byStem := map[string][]string{}

	for _, rel := range files {
		ext := filepath.Ext(rel)
		if !codeExtensions[ext] {
			continue
		}
		inv.Files++
		inv.ByLang[ext]++

		mi := moduleInfo{Path: rel, IsTest: isTestFile(rel)}
		if data, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(rel))); err == nil {
			mi.Symbols = extractSymbols(ext, data)
		}
		inv.Modules = append(inv.Modules, mi)

		if !mi.IsTest {
			byStem[stemOf(rel)] = append(byStem[stemOf(rel)], rel)
		}
	}

	// Mapa test→módulo por convención de nombres (foo_test.go→foo.go,
	// test_foo.py→foo.py, foo.test.ts→foo.ts). Heurística honesta: si no casa,
	// el test queda sin mapa — mejor incompleto que inventado.
	for _, m := range inv.Modules {
		if !m.IsTest {
			continue
		}
		if mods := byStem[testStem(m.Path)]; len(mods) > 0 {
			inv.TestMap[m.Path] = mods
		}
	}
	return inv
}

// isTestFile: convenciones de test por ecosistema.
func isTestFile(rel string) bool {
	base := filepath.Base(rel)
	switch {
	case strings.HasSuffix(base, "_test.go"),
		strings.HasPrefix(base, "test_"),
		strings.HasSuffix(base, "_test.py"),
		strings.Contains(base, ".test."),
		strings.Contains(base, ".spec."):
		return true
	}
	dir := filepath.ToSlash(filepath.Dir(rel))
	return dir == "tests" || strings.HasPrefix(dir, "tests/") ||
		strings.Contains(dir, "/tests/") || strings.Contains(dir, "/__tests__/")
}

// stemOf: nombre base sin extensión ("src/foo.py" → "foo").
func stemOf(rel string) string {
	base := filepath.Base(rel)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// testStem: el stem del MÓDULO que un archivo de test parece cubrir
// ("foo_test" → "foo", "test_foo" → "foo", "foo.test" → "foo").
func testStem(rel string) string {
	stem := stemOf(rel)
	stem = strings.TrimSuffix(stem, ".test")
	stem = strings.TrimSuffix(stem, ".spec")
	stem = strings.TrimSuffix(stem, "_test")
	stem = strings.TrimPrefix(stem, "test_")
	return stem
}

// extractSymbols corre los patrones del lenguaje y devuelve los nombres únicos
// en orden de aparición (capados).
func extractSymbols(ext string, data []byte) []string {
	pats := symbolPatterns[ext]
	if len(pats) == 0 {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, re := range pats {
		for _, m := range re.FindAllSubmatch(data, -1) {
			name := string(m[1])
			if !seen[name] {
				seen[name] = true
				out = append(out, name)
			}
		}
	}
	sort.Strings(out) // orden estable (los patrones se corren por tipo, no por línea)
	if len(out) > maxSymbolsPerFile {
		out = append(out[:maxSymbolsPerFile], fmt.Sprintf("+%d more", len(out)-maxSymbolsPerFile))
	}
	return out
}

func langSummary(byLang map[string]int) string {
	type kv struct {
		k string
		n int
	}
	var kvs []kv
	for k, n := range byLang {
		kvs = append(kvs, kv{k, n})
	}
	sort.Slice(kvs, func(i, j int) bool {
		if kvs[i].n != kvs[j].n {
			return kvs[i].n > kvs[j].n
		}
		return kvs[i].k < kvs[j].k
	})
	var parts []string
	for _, e := range kvs {
		parts = append(parts, fmt.Sprintf("%s×%d", e.k, e.n))
	}
	return strings.Join(parts, " ")
}

// renderInventoryMarkdown arma el .md que el LLM lee como mapa del territorio.
func renderInventoryMarkdown(inv inventory) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Code inventory — %s\n\n", inv.At)
	fmt.Fprintf(&b, "%d code file(s) · %s\n\n", inv.Files, langSummary(inv.ByLang))

	fmt.Fprintln(&b, "## Modules (top-level symbols)")
	fmt.Fprintln(&b, "")
	for _, m := range inv.Modules {
		if m.IsTest {
			continue
		}
		fmt.Fprintf(&b, "- `%s`", m.Path)
		if len(m.Symbols) > 0 {
			fmt.Fprintf(&b, ": %s", strings.Join(m.Symbols, ", "))
		}
		fmt.Fprintln(&b, "")
	}

	fmt.Fprintln(&b, "\n## Test → module map")
	fmt.Fprintln(&b, "")
	tests := make([]string, 0, len(inv.TestMap))
	for t := range inv.TestMap {
		tests = append(tests, t)
	}
	sort.Strings(tests)
	for _, t := range tests {
		fmt.Fprintf(&b, "- `%s` → %s\n", t, strings.Join(inv.TestMap[t], ", "))
	}
	if len(tests) == 0 {
		fmt.Fprintln(&b, "(no test files matched a module by naming convention)")
	}
	return b.String()
}
