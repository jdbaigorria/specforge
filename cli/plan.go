package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ----------------------------------------------------------------------------
// plan: por-feature, vive en specforge/features/<f>/progress/plan.json. Es la
// estimación de complejidad por wave que produce sf-build antes de ejecutar.
// ----------------------------------------------------------------------------

// Modelo de plan.json (CLI-SPEC §4.5).
type planFile struct {
	SchemaVersion string     `json:"schema_version"`
	Feature       string     `json:"feature"`
	Waves         []planWave `json:"waves"`
}

type planWave struct {
	N          int    `json:"n"`
	Complexity string `json:"complexity"`
	Rationale  string `json:"rationale"`
}

//go:embed templates/plan.tmpl.md
var planTemplate string

var planTmpl = template.Must(
	template.New("plan").
		Funcs(template.FuncMap{"list": joinOrDash, "orDash": orDash}).
		Parse(planTemplate),
)

// complexities es el enum válido, como set.
var complexities = map[string]bool{"low": true, "medium": true, "high": true}

func runPlan(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf plan <render|validate> --feature=NAME [project_dir]")
		return 2
	}
	action, rest := args[0], args[1:]
	switch action {
	case "render":
		projectDir, feature, toStdout, ok := parseArtifactFlags("plan", rest)
		if !ok {
			return 2
		}
		return renderPlan(projectDir, feature, toStdout)
	case "validate":
		projectDir, feature, _, ok := parseArtifactFlags("plan", rest)
		if !ok {
			return 2
		}
		return validatePlan(projectDir, feature)
	default:
		fmt.Fprintf(os.Stderr, "sf plan: unknown action %q (use render|validate)\n", action)
		return 2
	}
}

func readPlanFile(projectDir, feature string) (planFile, int) {
	path := filepath.Join(projectDir, "specforge", "features", feature, "progress", "plan.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf plan: cannot read %s (%v)\n", path, err)
		return planFile{}, 4
	}
	var pf planFile
	if err := json.Unmarshal(data, &pf); err != nil {
		fmt.Fprintf(os.Stderr, "sf plan: invalid JSON (%v)\n", err)
		return planFile{}, 2
	}
	return pf, 0
}

func renderPlan(projectDir, feature string, toStdout bool) int {
	pf, code := readPlanFile(projectDir, feature)
	if code != 0 {
		return code
	}
	var buf strings.Builder
	if err := planTmpl.Execute(&buf, pf); err != nil {
		fmt.Fprintf(os.Stderr, "sf plan: render failed (%v)\n", err)
		return 1
	}
	if toStdout {
		fmt.Print(buf.String())
		return 0
	}
	mdPath := filepath.Join(projectDir, "specforge", "features", feature, "progress", "plan.md")
	if err := os.WriteFile(mdPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf plan: cannot write %s (%v)\n", mdPath, err)
		return 1
	}
	fmt.Printf("rendered %s\n", filepath.Join("specforge", "features", feature, "progress", "plan.md"))
	return 0
}

func validatePlan(projectDir, feature string) int {
	pf, code := readPlanFile(projectDir, feature)
	if code != 0 {
		return code
	}
	var rep report
	checkPlan(pf, &rep)

	fmt.Printf("Validated %s: %d wave(s).\n", feature, len(pf.Waves))
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
	fmt.Printf("\nOK: plan.json valid (%d warning(s)).\n", len(rep.warnings))
	return 0
}

func checkPlan(pf planFile, rep *report) {
	if pf.Feature == "" {
		rep.errorf("missing `feature`")
	}
	seen := map[int]bool{}
	for _, w := range pf.Waves {
		if seen[w.N] {
			rep.errorf("duplicate wave number %d", w.N)
		}
		seen[w.N] = true
		if w.Complexity != "" && !complexities[w.Complexity] {
			rep.errorf("wave %d: invalid complexity %q (low|medium|high)", w.N, w.Complexity)
		}
	}
}
