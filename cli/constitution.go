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

// Modelo de constitution.json (CLI-SPEC §4.8). JSON-first: los principios pasan
// de strings sueltos a objetos estructurados con id + statement + applies_to. El
// `applies_to` es el mapeo fase→principio (paso 6): dice en qué fases el auditor
// de calidad debe chequear ese principio.
type constitutionFile struct {
	SchemaVersion string      `json:"schema_version"`
	IdentityMD    string      `json:"identity_md"`
	Principles    []principle `json:"principles"`
	Constraints   []string    `json:"constraints"`
	AntiGoals     []string    `json:"anti_goals"`
	Invariants    []invariant `json:"invariants"`
}

type principle struct {
	ID        string   `json:"id"`
	Statement string   `json:"statement"`
	AppliesTo []string `json:"applies_to,omitempty"` // fases donde el auditor lo chequea
}

type invariant struct {
	ID           string   `json:"id"`
	Rule         string   `json:"rule"`
	AppliesTo    []string `json:"applies_to,omitempty"`
	PromotedFrom []string `json:"promoted_from"`
	At           string   `json:"at"`
}

// auditablePhases: las fases que un auditor de calidad puede chequear. applies_to
// solo puede apuntar acá (lane/verdict son meta, no se auditan por principio).
var auditablePhases = map[string]bool{
	"requirements": true, "design": true, "tasks": true, "plan": true, "build": true,
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

	// Principios: id único + statement + applies_to válido.
	seenP := map[string]bool{}
	for i, p := range cf.Principles {
		label := p.ID
		if label == "" {
			label = fmt.Sprintf("principle #%d", i+1)
		}
		switch {
		case p.ID == "":
			rep.errorf("principle #%d: empty id", i+1)
		case seenP[p.ID]:
			rep.errorf("duplicate principle id %s", p.ID)
		default:
			seenP[p.ID] = true
		}
		if strings.TrimSpace(p.Statement) == "" {
			rep.errorf("%s: empty statement", label)
		}
		checkAppliesTo(label, p.AppliesTo, rep)
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
		checkAppliesTo(orDash(inv.ID), inv.AppliesTo, rep)
	}
}

// checkAppliesTo valida el mapeo fase→regla. Dos roles del debate:
//   - WARN si no hay applies_to → la regla no la chequea ningún auditor de fase
//     (solo la atrapa sf-audit). Es el "lint WARN" sobre la constitución.
//   - ERROR si apunta a una fase inexistente → gate estructural (determinista)
//     sobre la constitución misma: lo hermético envuelve lo cooperativo.
func checkAppliesTo(label string, appliesTo []string, rep *report) {
	if len(appliesTo) == 0 {
		rep.warnf("%s: no applies_to — won't be checked by any phase auditor (only by sf-audit)", label)
		return
	}
	for _, ph := range appliesTo {
		if !auditablePhases[ph] {
			rep.errorf("%s: applies_to references unknown phase %q", label, ph)
		}
	}
}

// loadConstitutionQuiet lee constitution.json SIN imprimir (a diferencia de
// readConstitutionFile). Devuelve (cf, false) si no existe o no parsea — lo usa
// for-judge para degradar con una nota en vez de fallar.
func loadConstitutionQuiet(projectDir string) (constitutionFile, bool) {
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "constitution.json"))
	if err != nil {
		return constitutionFile{}, false
	}
	var cf constitutionFile
	if json.Unmarshal(data, &cf) != nil {
		return constitutionFile{}, false
	}
	return cf, true
}
