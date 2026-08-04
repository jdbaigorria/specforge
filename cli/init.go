package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf init --minimal` — el arranque de 30 segundos (D1').
//
// El costo de entrada de SpecForge es la conversación de constitución + los
// gates. Para un solo dev que quiere probar, eso es demasiado ANTES de ver
// valor. `--minimal` scaffoldea lo justo: la estructura de dirs, una
// constitución de 3 principios default (que crece después por backprop), y el
// test_cmd DETECTADO del stack (§2.8: detección de stack automatizable).
//
// No reemplaza a la skill sf-init (la conversación de constitución sigue siendo
// el camino completo) — es la puerta de entrada para "primera feature en 10
// minutos" (docs/quickstart.md).
// ----------------------------------------------------------------------------

func runInit(args []string) int {
	projectDir := "."
	minimal := false
	for _, a := range args {
		switch {
		case a == "--minimal":
			minimal = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf init: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if !minimal {
		fmt.Fprintln(os.Stderr, "sf init: use --minimal for the quick scaffold, or run the sf-init skill "+
			"for the full constitution conversation.")
		return 2
	}

	sf := filepath.Join(projectDir, "specforge")
	if fileExists(filepath.Join(sf, "constitution.json")) {
		fmt.Fprintln(os.Stderr, "sf init: specforge/constitution.json already exists — nothing to scaffold.")
		return 4
	}

	// Estructura mínima. drafts/ es el rincón de autoría (R5).
	for _, d := range []string{sf, filepath.Join(sf, "features"), filepath.Join(sf, "context"), filepath.Join(sf, "drafts")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			fmt.Fprintf(os.Stderr, "sf init: %v\n", err)
			return 1
		}
	}

	// Stack → test_cmd + report (la causalidad A2 queda configurada de arranque
	// cuando el stack la soporta).
	testCmd, report, stack := detectStack(projectDir)

	cf := constitutionFile{
		SchemaVersion: schemaVersionCurrent,
		IdentityMD: "## Identity\n\n(fill in: one paragraph — what this project is and why it exists)\n",
		Principles: []principle{
			{ID: "P1", Statement: "Minimal code: the smallest change that satisfies the requirement — no speculative structure.",
				AppliesTo: []string{"design", "tasks", "build"}},
			{ID: "P2", Statement: "Tests prove requirements: every requirement names a real, runnable test.",
				AppliesTo: []string{"tasks", "build"}},
			{ID: "P3", Statement: "Explicit scope: what a feature deliberately does NOT do is written down.",
				AppliesTo: []string{"requirements"}},
		},
		Constraints: []string{},
		AntiGoals:   []string{},
		Invariants:  []invariant{},
	}
	if testCmd != "" {
		cf.Build = &buildConfig{Mode: "inline", TestCmd: testCmd, Report: report}
	}

	// finishSave valida y escribe json+md — el scaffold pasa por la MISMA vía
	// validada que cualquier autoría (nada especial que pueda divergir).
	jsonPath, mdPath := artifactPaths("constitution", projectDir, "")
	if code := finishSave(cf, jsonPath, mdPath); code != 0 {
		return code
	}

	// history.md: el log humano del proyecto (append-only, del usuario).
	hist := filepath.Join(sf, "history.md")
	if !fileExists(hist) {
		_ = os.WriteFile(hist, []byte("# Project history\n\n- "+nowUTC()+" — specforge initialized (`sf init --minimal`)\n"), 0o644)
	}

	fmt.Println("specforge/ initialized (minimal).")
	if stack != "" {
		fmt.Printf("stack detected: %s → build.test_cmd=%q", stack, testCmd)
		if report != "" {
			fmt.Printf(" build.report=%q", report)
		}
		fmt.Println()
	} else {
		fmt.Println("stack not detected — set build.test_cmd in the constitution before `sf check run`.")
	}
	fmt.Println("\nNext steps (first feature in ~10 minutes — docs/quickstart.md):")
	fmt.Println("  1. edit the identity paragraph:  specforge/constitution.md (via `sf save constitution`)")
	fmt.Println("  2. create a feature:             sf feature add --feature=<name>")
	fmt.Println("  3. author + gate the spec:       run the sf-propose skill (or draft JSON in drafts/ + `sf save`)")
	fmt.Println("  4. wire CI (the outside gate):   sf verify --init-ci")
	return 0
}

// detectStack infiere el comando de test del ecosistema del repo. Devuelve
// (test_cmd, report, etiqueta) — vacíos si no hay señal.
func detectStack(projectDir string) (testCmd, report, stack string) {
	has := func(name string) bool { return fileExists(filepath.Join(projectDir, name)) }
	switch {
	case has("go.mod"):
		return "go test -json ./...", "go-json", "go"
	case has("pyproject.toml") || has("setup.py") || has("pytest.ini") || has("requirements.txt"):
		return "pytest -q --junitxml={report}", "junit", "python"
	case has("package.json"):
		return "npm test", "", "node"
	case has("Cargo.toml"):
		return "cargo test", "", "rust"
	}
	return "", "", ""
}
