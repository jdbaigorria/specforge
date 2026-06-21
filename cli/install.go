package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf install` — instalador multi-harness.
//
// ESTE CUT es solo DETECCIÓN + PLAN (dry-run): detecta qué arneses están
// presentes y muestra QUÉ instalaría dónde, SIN tocar ningún archivo. Las
// escrituras reales (symlinks de skills, merge marker-based de AGENT.md, wiring
// de adapters) + backup + `sf uninstall` llegan en el cut siguiente. Empezar por
// el plan es seguro (riesgo cero) y valida los paths por arnés antes de escribir
// en configs del usuario (~/.claude, .pi/settings.json, etc.).
//
// Diseño: las dos partes con lógica (resolver el source root, computar el plan
// por arnés) son funciones PURAS → testeables sin tocar disco real.
// ----------------------------------------------------------------------------

// installAction es un paso del plan. kind clasifica la operación; source/target
// son rutas; note explica matices (p.ej. un gap del arnés).
type installAction struct {
	kind   string // symlink | merge | wire | note
	source string // ruta en el repo SpecForge (vacío para "note")
	target string // ruta destino en el sistema/proyecto
	desc   string // descripción legible
}

// harnessPlan es el plan para UN arnés: si está presente y qué haría.
type harnessPlan struct {
	name     string
	detected bool
	actions  []installAction
}

// runInstall parsea flags y emite el plan. Flags:
//
//	--from=PATH   source root de SpecForge (default: autodetectado)
//	--global      instalar a nivel usuario (~), en vez de project-local (default)
//	--dry-run     no-op por ahora (este cut SIEMPRE es dry-run)
func runInstall(args []string) int {
	from, global := "", false
	projectDir := "."
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--from="):
			from = strings.TrimPrefix(a, "--from=")
		case a == "--global":
			global = true
		case a == "--dry-run":
			// reconocido; este cut ya es dry-run
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf install: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	sourceRoot, err := resolveSourceRoot(from)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf install: %v\n  pass --from=/path/to/specforge\n", err)
		return 4
	}
	home, _ := os.UserHomeDir()
	projAbs, _ := filepath.Abs(projectDir)

	scope := "project"
	base := projAbs
	if global {
		scope = "global"
		base = home
	}

	fmt.Printf("SpecForge install plan\n  source: %s\n  scope:  %s (%s)\n\n", sourceRoot, scope, base)

	plans := planAll(sourceRoot, base, global, home)
	for _, p := range plans {
		status := "not detected"
		if p.detected {
			status = "detected"
		}
		fmt.Printf("%-12s %s\n", p.name, status)
		for _, a := range p.actions {
			if a.kind == "note" {
				fmt.Printf("  · %s\n", a.desc)
				continue
			}
			fmt.Printf("  %-7s %s\n", a.kind, a.desc)
		}
		fmt.Println()
	}

	fmt.Println("(dry-run: nothing was written. Apply with backup lands in the next cut.)")
	return 0
}

// resolveSourceRoot ubica la raíz del repo SpecForge (la que tiene skills/ +
// AGENT.md). Orden: --from explícito → dir del binario hacia arriba → cwd hacia
// arriba. PURA salvo el acceso a disco para verificar existencia.
func resolveSourceRoot(from string) (string, error) {
	if from != "" {
		abs, _ := filepath.Abs(from)
		if isSourceRoot(abs) {
			return abs, nil
		}
		return "", fmt.Errorf("%s is not a SpecForge source (no skills/ + AGENT.md)", from)
	}
	if exe, err := os.Executable(); err == nil {
		if r := findSourceUp(filepath.Dir(exe)); r != "" {
			return r, nil
		}
	}
	if cwd, err := os.Getwd(); err == nil {
		if r := findSourceUp(cwd); r != "" {
			return r, nil
		}
	}
	return "", fmt.Errorf("could not locate the SpecForge source")
}

// isSourceRoot: ¿este dir es la raíz del repo? (tiene skills/ y AGENT.md).
func isSourceRoot(dir string) bool {
	return isDir(filepath.Join(dir, "skills")) && fileExists(filepath.Join(dir, "AGENT.md"))
}

// findSourceUp sube desde start buscando la raíz del repo, hasta la raíz del FS.
func findSourceUp(start string) string {
	dir, _ := filepath.Abs(start)
	for {
		if isSourceRoot(dir) {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir { // llegamos a "/" sin encontrarla
			return ""
		}
		dir = parent
	}
}

// planAll computa el plan de los 4 arneses. PURA: solo arma rutas + chequea
// presencia (lectura), no escribe. base = dónde instalar (proj o home).
func planAll(sourceRoot, base string, global bool, home string) []harnessPlan {
	skills := filepath.Join(sourceRoot, "skills")
	agent := filepath.Join(sourceRoot, "AGENT.md")

	return []harnessPlan{
		planClaude(skills, agent, sourceRoot, base, global, home),
		planPi(skills, agent, base, global, home),
		planOpencode(sourceRoot, agent, base, global, home),
		planCursor(agent, base, global, home),
	}
}

// detectedAt: ¿existe alguno de estos paths? (marca de presencia del arnés).
func detectedAt(paths ...string) bool {
	for _, p := range paths {
		if pathExists(p) {
			return true
		}
	}
	return false
}

func planClaude(skills, agent, sourceRoot, base string, global bool, home string) harnessPlan {
	dir := filepath.Join(base, ".claude")
	if global {
		dir = filepath.Join(home, ".claude")
	}
	return harnessPlan{
		name:     "claude-code",
		detected: detectedAt(filepath.Join(home, ".claude"), filepath.Join(base, ".claude")),
		actions: []installAction{
			{"symlink", skills, filepath.Join(dir, "skills"), "skills/ → " + tildeHome(home, filepath.Join(dir, "skills"))},
			{"merge", agent, filepath.Join(base, "CLAUDE.md"), "AGENT.md → CLAUDE.md (marker-based)"},
			{"wire", filepath.Join(sourceRoot, "hooks", "claude-code", "hooks.json"), filepath.Join(dir, "settings.json"),
				"hooks → " + tildeHome(home, filepath.Join(dir, "settings.json")) + " (or the Claude plugin)"},
		},
	}
}

func planPi(skills, agent, base string, global bool, home string) harnessPlan {
	settings := filepath.Join(base, ".pi", "settings.json")
	if global {
		settings = filepath.Join(home, ".pi", "agent", "settings.json")
	}
	ext := filepath.Join(filepath.Dir(skills), "hooks", "pi", "specforge.js")
	return harnessPlan{
		name:     "pi",
		detected: detectedAt(filepath.Join(home, ".pi"), filepath.Join(base, ".pi")),
		actions: []installAction{
			{"wire", skills, settings, `add "skills": [skills/] → ` + tildeHome(home, settings)},
			{"wire", ext, settings, `add "extensions": [hooks/pi/specforge.js] → ` + tildeHome(home, settings)},
			{"merge", agent, settings, "AGENT.md → pi system prompt"},
		},
	}
}

func planOpencode(sourceRoot, agent, base string, global bool, home string) harnessPlan {
	pluginDir := filepath.Join(base, ".opencode", "plugin")
	if global {
		pluginDir = filepath.Join(home, ".config", "opencode", "plugin")
	}
	plugin := filepath.Join(sourceRoot, "hooks", "opencode", "specforge.js")
	return harnessPlan{
		name: "opencode",
		detected: detectedAt(
			filepath.Join(home, ".config", "opencode"),
			filepath.Join(base, ".opencode"),
			filepath.Join(base, "opencode.json"),
		),
		actions: []installAction{
			{"symlink", plugin, filepath.Join(pluginDir, "specforge.js"), "opencode plugin → " + tildeHome(home, filepath.Join(pluginDir, "specforge.js"))},
			{"merge", agent, filepath.Join(base, "AGENTS.md"), "AGENT.md → AGENTS.md (marker-based)"},
			{"note", "", "", "no skills dir on opencode; per-turn context injection is gapped (see hooks/opencode/README.md)"},
		},
	}
}

func planCursor(agent, base string, global bool, home string) harnessPlan {
	rules := filepath.Join(base, ".cursor", "rules", "specforge.mdc")
	if global {
		rules = filepath.Join(home, ".cursor", "rules", "specforge.mdc")
	}
	return harnessPlan{
		name:     "cursor",
		detected: detectedAt(filepath.Join(home, ".cursor"), filepath.Join(base, ".cursor")),
		actions: []installAction{
			{"merge", agent, rules, "AGENT.md → " + tildeHome(home, rules)},
			{"note", "", "", "no enforcement adapter yet (Cursor hooks pending); cooperative layer only"},
		},
	}
}

// pathExists: ¿existe el path (archivo o dir)?
func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

// rel acorta una ruta bajo $HOME a ~/… para que el plan se lea mejor.
func tildeHome(home, p string) string {
	if home != "" && strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}
