package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// ----------------------------------------------------------------------------
// constitution: a diferencia de los otros artefactos, es a NIVEL PROYECTO (no
// por-feature). Vive en specforge/constitution.json → specforge/constitution.md,
// así que NO usa --feature, solo el project dir.
// ----------------------------------------------------------------------------

// Modelo de constitution.json (CLI-SPEC §4.8).
type constitutionFile struct {
	SchemaVersion string      `json:"schema_version"`
	IdentityMD    string      `json:"identity_md"`
	Principles    []string    `json:"principles"`
	Constraints   []string    `json:"constraints"`
	AntiGoals     []string    `json:"anti_goals"`
	Invariants    []invariant `json:"invariants"`
}

type invariant struct {
	ID           string   `json:"id"`
	Rule         string   `json:"rule"`
	PromotedFrom []string `json:"promoted_from"`
	At           string   `json:"at"`
}

//go:embed templates/constitution.tmpl.md
var constitutionTemplate string

var constitutionTmpl = template.Must(
	template.New("constitution").
		Funcs(template.FuncMap{"list": joinOrDash, "orDash": orDash}).
		Parse(constitutionTemplate),
)

var invIDRe = regexp.MustCompile(`^I\d+$`)

// runConstitution despacha `sf constitution <render|validate>`. Solo project dir
// (+ --stdout en render); no hay --feature.
func runConstitution(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf constitution <render|validate> [project_dir]")
		return 2
	}
	action, rest := args[0], args[1:]

	projectDir := "."
	toStdout := false
	for _, a := range rest {
		switch {
		case a == "--stdout":
			toStdout = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf constitution: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	switch action {
	case "render":
		return renderConstitution(projectDir, toStdout)
	case "validate":
		return validateConstitution(projectDir)
	default:
		fmt.Fprintf(os.Stderr, "sf constitution: unknown action %q (use render|validate)\n", action)
		return 2
	}
}

func readConstitutionFile(projectDir string) (constitutionFile, int) {
	path := filepath.Join(projectDir, "specforge", "constitution.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf constitution: cannot read %s (%v)\n", path, err)
		return constitutionFile{}, 4
	}
	var cf constitutionFile
	if err := json.Unmarshal(data, &cf); err != nil {
		fmt.Fprintf(os.Stderr, "sf constitution: invalid JSON (%v)\n", err)
		return constitutionFile{}, 2
	}
	return cf, 0
}

func renderConstitution(projectDir string, toStdout bool) int {
	cf, code := readConstitutionFile(projectDir)
	if code != 0 {
		return code
	}
	var buf strings.Builder
	if err := constitutionTmpl.Execute(&buf, cf); err != nil {
		fmt.Fprintf(os.Stderr, "sf constitution: render failed (%v)\n", err)
		return 1
	}
	if toStdout {
		fmt.Print(buf.String())
		return 0
	}
	mdPath := filepath.Join(projectDir, "specforge", "constitution.md")
	if err := os.WriteFile(mdPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf constitution: cannot write %s (%v)\n", mdPath, err)
		return 1
	}
	fmt.Printf("rendered %s\n", filepath.Join("specforge", "constitution.md"))
	return 0
}

func validateConstitution(projectDir string) int {
	cf, code := readConstitutionFile(projectDir)
	if code != 0 {
		return code
	}
	var rep report
	checkConstitution(cf, &rep)

	fmt.Printf("Validated constitution: %d principle(s), %d invariant(s).\n", len(cf.Principles), len(cf.Invariants))
	for _, w := range rep.warnings {
		fmt.Printf("  warning: %s\n", w)
	}
	for _, e := range rep.errors {
		fmt.Printf("  ERROR:   %s\n", e)
	}
	if len(rep.errors) > 0 {
		fmt.Printf("\nFAIL: %d error(s), %d warning(s).\n", len(rep.errors), len(rep.warnings))
		return 2
	}
	fmt.Printf("\nOK: constitution.json valid (%d warning(s)).\n", len(rep.warnings))
	return 0
}

func checkConstitution(cf constitutionFile, rep *report) {
	if strings.TrimSpace(cf.IdentityMD) == "" {
		rep.warnf("identity_md is empty")
	}
	seen := map[string]bool{}
	for _, inv := range cf.Invariants {
		switch {
		case inv.ID == "":
			rep.errorf("invariant with empty id")
		case seen[inv.ID]:
			rep.errorf("duplicate invariant id %s", inv.ID)
		default:
			seen[inv.ID] = true
		}
		if inv.ID != "" && !invIDRe.MatchString(inv.ID) {
			rep.warnf("%s: id not in I# form", inv.ID)
		}
		if strings.TrimSpace(inv.Rule) == "" {
			rep.errorf("%s: empty rule", inv.ID)
		}
	}
}
