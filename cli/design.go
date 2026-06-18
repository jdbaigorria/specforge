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
// Tercer artefacto del piloto: design. Misma regla (json=fuente, md=render).
// Estructurarlo le permite a `sf context for-wave` incluir los CUERPOS de los
// componentes (C#) en scope, igual que ya hace con los requirements.
// ----------------------------------------------------------------------------

// Modelo de design.json (CLI-SPEC §4.3).
type designFile struct {
	SchemaVersion    string      `json:"schema_version"`
	Feature          string      `json:"feature"`
	SummaryMD        string      `json:"summary_md"`
	Components       []component `json:"components"`
	Decisions        []decision  `json:"decisions"`
	Interfaces       []iface     `json:"interfaces"`
	SectionsIncluded []string    `json:"sections_included"`
}

type component struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	Kind             string   `json:"kind"`
	Responsibilities []string `json:"responsibilities"`
	DependsOn        []string `json:"depends_on"`
}

type decision struct {
	ID          string           `json:"id"`
	Question    string           `json:"question"`
	Options     []decisionOption `json:"options"`
	Chosen      string           `json:"chosen"`
	RationaleMD string           `json:"rationale_md"`
}

type decisionOption struct {
	Name string   `json:"name"`
	Pros []string `json:"pros"`
	Cons []string `json:"cons"`
}

type iface struct {
	Name string `json:"name"`
}

//go:embed templates/design.tmpl.md
var designTemplate string

// La func `section` decide si una sección se renderiza: si sections_included
// está vacío, se renderiza todo; si tiene elementos, solo los listados (CLI-SPEC
// §4.3: "los campos no incluidos se omiten aunque estén poblados").
var designTmpl = template.Must(
	template.New("design").
		Funcs(template.FuncMap{
			"list":    joinOrDash,
			"orDash":  orDash,
			"section": sectionIncluded,
		}).
		Parse(designTemplate),
)

func sectionIncluded(included []string, name string) bool {
	if len(included) == 0 {
		return true
	}
	for _, s := range included {
		if s == name {
			return true
		}
	}
	return false
}

var (
	compIDRe = regexp.MustCompile(`^C\d+$`)
	decIDRe  = regexp.MustCompile(`^D\d+$`)
)

// runDesign despacha `sf design <render|validate>`.
func runDesign(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf design <render|validate> --feature=NAME [project_dir]")
		return 2
	}
	action, rest := args[0], args[1:]
	switch action {
	case "render":
		projectDir, feature, toStdout, ok := parseArtifactFlags("design", rest)
		if !ok {
			return 2
		}
		return renderDesign(projectDir, feature, toStdout)
	case "validate":
		projectDir, feature, _, ok := parseArtifactFlags("design", rest)
		if !ok {
			return 2
		}
		return validateDesign(projectDir, feature)
	default:
		fmt.Fprintf(os.Stderr, "sf design: unknown action %q (use render|validate)\n", action)
		return 2
	}
}

// readDesignFile lee y parsea design.json. (code 0 = OK.)
func readDesignFile(projectDir, feature string) (designFile, int) {
	path := filepath.Join(projectDir, "specforge", "features", feature, "design.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf design: cannot read %s (%v)\n", path, err)
		return designFile{}, 4
	}
	var df designFile
	if err := json.Unmarshal(data, &df); err != nil {
		fmt.Fprintf(os.Stderr, "sf design: invalid JSON (%v)\n", err)
		return designFile{}, 2
	}
	return df, 0
}

func renderDesign(projectDir, feature string, toStdout bool) int {
	df, code := readDesignFile(projectDir, feature)
	if code != 0 {
		return code
	}
	var buf strings.Builder
	if err := designTmpl.Execute(&buf, df); err != nil {
		fmt.Fprintf(os.Stderr, "sf design: render failed (%v)\n", err)
		return 1
	}
	if toStdout {
		fmt.Print(buf.String())
		return 0
	}
	mdPath := filepath.Join(projectDir, "specforge", "features", feature, "design.md")
	if err := os.WriteFile(mdPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf design: cannot write %s (%v)\n", mdPath, err)
		return 1
	}
	fmt.Printf("rendered %s\n", filepath.Join("specforge", "features", feature, "design.md"))
	return 0
}

func validateDesign(projectDir, feature string) int {
	df, code := readDesignFile(projectDir, feature)
	if code != 0 {
		return code
	}
	var rep report
	checkDesign(df, &rep)

	fmt.Printf("Validated %s: %d component(s), %d decision(s).\n", feature, len(df.Components), len(df.Decisions))
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
	fmt.Printf("\nOK: design.json valid (%d warning(s)).\n", len(rep.warnings))
	return 0
}

// checkDesign valida en DOS pasadas: primero junta los ids de componentes,
// después verifica que los depends_on apunten a componentes que existen
// (integridad referencial INTERNA, resoluble dentro del propio archivo).
func checkDesign(df designFile, rep *report) {
	if df.Feature == "" {
		rep.errorf("missing `feature`")
	}

	compIDs := map[string]bool{}
	for _, c := range df.Components {
		switch {
		case c.ID == "":
			rep.errorf("component with empty id")
		case compIDs[c.ID]:
			rep.errorf("duplicate component id %s", c.ID)
		default:
			compIDs[c.ID] = true
		}
		if c.ID != "" && !compIDRe.MatchString(c.ID) {
			rep.warnf("%s: id not in C# form", c.ID)
		}
		if c.Name == "" {
			rep.errorf("%s: empty name", c.ID)
		}
	}

	// Segunda pasada: ahora que tenemos todos los ids, validamos los depends_on.
	for _, c := range df.Components {
		for _, dep := range c.DependsOn {
			if !compIDs[dep] {
				rep.errorf("%s: depends_on unknown component %s", c.ID, dep)
			}
		}
	}

	decIDs := map[string]bool{}
	for _, d := range df.Decisions {
		switch {
		case d.ID == "":
			rep.errorf("decision with empty id")
		case decIDs[d.ID]:
			rep.errorf("duplicate decision id %s", d.ID)
		default:
			decIDs[d.ID] = true
		}
		if d.ID != "" && !decIDRe.MatchString(d.ID) {
			rep.warnf("%s: id not in D# form", d.ID)
		}
		// `chosen` debe ser una de las opciones ofrecidas.
		if d.Chosen != "" && len(d.Options) > 0 {
			found := false
			for _, o := range d.Options {
				if o.Name == d.Chosen {
					found = true
					break
				}
			}
			if !found {
				rep.errorf("%s: chosen %q is not one of the options", d.ID, d.Chosen)
			}
		}
	}
}
