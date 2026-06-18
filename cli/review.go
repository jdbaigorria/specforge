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
// review: por-feature (specforge/features/<f>/review.json). Trae una validación
// LÓGICA cruzada (CLI-SPEC §4.7): un veredicto "approve" no puede convivir con
// requirements "missing" ni con violaciones a la constitución.
// ----------------------------------------------------------------------------

// Modelo de review.json (CLI-SPEC §4.7).
type reviewFile struct {
	SchemaVersion          string     `json:"schema_version"`
	Feature                string     `json:"feature"`
	Traceability           []traceRow `json:"traceability"`
	Gaps                   []string   `json:"gaps"`
	ConstitutionViolations []string   `json:"constitution_violations"`
	Verdict                string     `json:"verdict"`
	VerdictRationaleMD     string     `json:"verdict_rationale_md"`
}

type traceRow struct {
	Requirement string   `json:"requirement"`
	Tasks       []string `json:"tasks"`
	Files       []string `json:"files"`
	Tests       []string `json:"tests"`
	Status      string   `json:"status"`
}

//go:embed templates/review.tmpl.md
var reviewTemplate string

var reviewTmpl = template.Must(
	template.New("review").
		Funcs(template.FuncMap{"list": joinOrDash, "orDash": orDash}).
		Parse(reviewTemplate),
)

var (
	verdicts      = map[string]bool{"approve": true, "approve-with-notes": true, "revise": true}
	traceStatuses = map[string]bool{"implemented": true, "partial": true, "missing": true}
)

func runReview(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf review <render|validate> --feature=NAME [project_dir]")
		return 2
	}
	action, rest := args[0], args[1:]
	switch action {
	case "render":
		projectDir, feature, toStdout, ok := parseArtifactFlags("review", rest)
		if !ok {
			return 2
		}
		return renderReview(projectDir, feature, toStdout)
	case "validate":
		projectDir, feature, _, ok := parseArtifactFlags("review", rest)
		if !ok {
			return 2
		}
		return validateReview(projectDir, feature)
	default:
		fmt.Fprintf(os.Stderr, "sf review: unknown action %q (use render|validate)\n", action)
		return 2
	}
}

func readReviewFile(projectDir, feature string) (reviewFile, int) {
	path := filepath.Join(projectDir, "specforge", "features", feature, "review.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf review: cannot read %s (%v)\n", path, err)
		return reviewFile{}, 4
	}
	var rv reviewFile
	if err := json.Unmarshal(data, &rv); err != nil {
		fmt.Fprintf(os.Stderr, "sf review: invalid JSON (%v)\n", err)
		return reviewFile{}, 2
	}
	return rv, 0
}

func renderReview(projectDir, feature string, toStdout bool) int {
	rv, code := readReviewFile(projectDir, feature)
	if code != 0 {
		return code
	}
	var buf strings.Builder
	if err := reviewTmpl.Execute(&buf, rv); err != nil {
		fmt.Fprintf(os.Stderr, "sf review: render failed (%v)\n", err)
		return 1
	}
	if toStdout {
		fmt.Print(buf.String())
		return 0
	}
	mdPath := filepath.Join(projectDir, "specforge", "features", feature, "review.md")
	if err := os.WriteFile(mdPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf review: cannot write %s (%v)\n", mdPath, err)
		return 1
	}
	fmt.Printf("rendered %s\n", filepath.Join("specforge", "features", feature, "review.md"))
	return 0
}

func validateReview(projectDir, feature string) int {
	rv, code := readReviewFile(projectDir, feature)
	if code != 0 {
		return code
	}
	var rep report
	checkReview(rv, &rep)

	fmt.Printf("Validated %s: verdict %q, %d row(s).\n", feature, rv.Verdict, len(rv.Traceability))
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
	fmt.Printf("\nOK: review.json valid (%d warning(s)).\n", len(rep.warnings))
	return 0
}

func checkReview(rv reviewFile, rep *report) {
	if rv.Feature == "" {
		rep.errorf("missing `feature`")
	}
	if rv.Verdict != "" && !verdicts[rv.Verdict] {
		rep.errorf("invalid verdict %q (approve|approve-with-notes|revise)", rv.Verdict)
	}

	missing := 0
	for _, row := range rv.Traceability {
		if row.Status != "" && !traceStatuses[row.Status] {
			rep.errorf("%s: invalid status %q (implemented|partial|missing)", row.Requirement, row.Status)
		}
		if row.Status == "missing" {
			missing++
		}
	}

	// Validación LÓGICA cruzada: approve es incompatible con requirements
	// faltantes o violaciones de la constitución.
	if rv.Verdict == "approve" {
		if missing > 0 {
			rep.errorf("verdict=approve but %d requirement(s) are missing", missing)
		}
		if len(rv.ConstitutionViolations) > 0 {
			rep.errorf("verdict=approve but there are %d constitution violation(s)", len(rv.ConstitutionViolations))
		}
	}
}
