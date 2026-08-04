package main

import (
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// Tests del contrato skills↔CLI (lint_contract.go, C5).
// ----------------------------------------------------------------------------

func TestInvocationProblemsValidCommands(t *testing.T) {
	// Invocaciones REALES que aparecen en las skills — todas deben pasar.
	valid := []string{
		"sf save requirements --feature=auth --json -",
		"sf gate approve --feature=auth --phase=verdict",
		"sf feature set-lane --feature=auth --to=lite|standard",
		"sf check run --feature=auth",
		"sf plan compute --feature=auth && sf save plan --feature=auth --json -",
		"sf next --json",
		"sf context current",
		"sf trace verify --feature=NAME --contract --wave=2",
		"sf verify --init-ci",
		"sf migrate --dry-run",
		"echo '{}' | sf gate record-verdict --feature=x --json -",
		"sf save <artifact> --feature=NAME --json -", // placeholder de docs
		"sf status --artifacts",
	}
	for _, s := range valid {
		if probs := invocationProblems(s); len(probs) != 0 {
			t.Errorf("%q should be valid, got %v", s, probs)
		}
	}
}

func TestInvocationProblemsCatchesDrift(t *testing.T) {
	cases := []struct{ span, want string }{
		{"sf regenerate --feature=x", "unknown `sf` command"},        // comando inventado
		{"sf gate aprove --feature=x", "no sub-command `aprove`"},    // typo de subcomando
		{"sf next --force", "no flag `--force`"},                     // flag inventado
		{"sf save trace --feature=x --output=md", "no flag `--output`"},
		{"sf check run --feature=x && sf archive --feature=x", "unknown `sf` command"},
	}
	for _, c := range cases {
		probs := strings.Join(invocationProblems(c.span), "\n")
		if !strings.Contains(probs, c.want) {
			t.Errorf("%q: want problem containing %q, got %q", c.span, c.want, probs)
		}
	}
}

func TestInvocationProblemsIgnoresProseAndNonSf(t *testing.T) {
	// Sin invocación `sf ` → sin problemas (aunque mencione palabras parecidas).
	for _, s := range []string{
		"the transform function",
		"sf-propose", // skill, no CLI
		"npm run sf",
	} {
		if probs := invocationProblems(s); len(probs) != 0 {
			t.Errorf("%q: expected no problems, got %v", s, probs)
		}
	}
}

// TestCLISurfaceMatchesUsage: la tabla cliSurface y el usage de main.go deben
// listar EXACTAMENTE los mismos comandos — si se agrega un comando sin
// actualizar la tabla (o viceversa), este test lo caza.
func TestCLISurfaceMatchesUsage(t *testing.T) {
	usageCmds := map[string]bool{}
	inCommands := false
	for _, line := range strings.Split(usageText, "\n") {
		if strings.HasPrefix(line, "commands:") {
			inCommands = true
			continue
		}
		if !inCommands || !strings.HasPrefix(line, "  ") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) > 0 {
			usageCmds[fields[0]] = true
		}
	}

	for cmd := range cliSurface {
		if !usageCmds[cmd] {
			t.Errorf("cliSurface has %q but usageText does not list it", cmd)
		}
	}
	for cmd := range usageCmds {
		if _, ok := cliSurface[cmd]; !ok {
			t.Errorf("usageText lists %q but cliSurface does not cover it", cmd)
		}
	}
}

// TestCheckCLIContractOnFile: integración chica — un markdown con una
// invocación rota dentro de un fence dispara el error; la prosa no.
func TestCheckCLIContractOnFile(t *testing.T) {
	dir := t.TempDir()
	md := dir + "/SKILL.md"
	mustWrite(t, md, "# demo\n\n"+
		"In prose, the sf gate command is fine to mention.\n\n"+ // prosa: ignorada
		"```bash\nsf gate approve --feature=x --seal\n```\n") // flag inventado

	var rep report
	checkCLIContract([]string{md}, dir, &rep)
	joined := strings.Join(rep.errors, "\n")
	if !strings.Contains(joined, "--seal") {
		t.Fatalf("expected a --seal contract error, got %v", rep.errors)
	}
	if strings.Contains(joined, "sub-command `command`") {
		t.Fatal("prose mention must not be validated")
	}
}
