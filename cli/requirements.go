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
// Segundo artefacto del piloto JSON-first: requirements. Misma regla que tasks:
// requirements.json es la FUENTE, requirements.md el RENDER. Estructurar esto es
// lo que le permite a `sf context for-wave` incluir los CUERPOS de los R# en
// scope (no solo los IDs) — el ahorro real de tokens.
// ----------------------------------------------------------------------------

// Modelo de requirements.json (CLI-SPEC §4.2). Nota: en JSON `state`/`trigger`
// pueden venir como null; al decodificar en un string quedan "" (no
// distinguimos null de ausente, y para EARS no hace falta).
type requirementsFile struct {
	SchemaVersion string        `json:"schema_version"`
	Feature       string        `json:"feature"`
	Summary       string        `json:"summary"`
	Actors        []string      `json:"actors"`
	Requirements  []requirement `json:"requirements"`
}

type requirement struct {
	ID         string   `json:"id"`
	EarsType   string   `json:"ears_type"`
	Trigger    string   `json:"trigger"`
	State      string   `json:"state"`
	Behavior   string   `json:"behavior"`
	Acceptance []string `json:"acceptance"`
	Source     string   `json:"source"`
	Tags       []string `json:"tags"`
}

//go:embed templates/requirements.tmpl.md
var reqTemplate string

// reqTmpl agrega la función `ears`, que arma la oración EARS según el tipo.
var reqTmpl = template.Must(
	template.New("requirements").
		Funcs(template.FuncMap{
			"list":   joinOrDash,
			"orDash": orDash,
			"ears":   earsSentence,
		}).
		Parse(reqTemplate),
)

// earsTypes es el enum válido de tipos EARS, como set.
var earsTypes = map[string]bool{
	"event": true, "state": true, "error": true, "ubiquitous": true,
}

// earsSentence arma la oración EARS canónica. Le saca un "shall " inicial al
// behavior para no duplicarlo ("the system shall shall ...").
func earsSentence(r requirement) string {
	b := strings.TrimPrefix(r.Behavior, "shall ")
	switch r.EarsType {
	case "event":
		return fmt.Sprintf("When %s, the system shall %s.", r.Trigger, b)
	case "state":
		return fmt.Sprintf("While %s, the system shall %s.", r.State, b)
	case "error":
		return fmt.Sprintf("If %s, then the system shall %s.", r.Trigger, b)
	default: // ubiquitous
		return fmt.Sprintf("The system shall %s.", b)
	}
}

// runRequirements despacha `sf requirements <render|validate>`.
func runRequirements(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf requirements <render|validate> --feature=NAME [project_dir]")
		return 2
	}
	action, rest := args[0], args[1:]
	switch action {
	case "render":
		return reqRenderCmd(rest)
	case "validate":
		return reqValidateCmd(rest)
	default:
		fmt.Fprintf(os.Stderr, "sf requirements: unknown action %q (use render|validate)\n", action)
		return 2
	}
}

// parseReqFlags extrae --feature, --stdout y el project dir (compartido por
// render y validate; validate ignora stdout).
func parseReqFlags(args []string) (projectDir, feature string, toStdout bool, ok bool) {
	projectDir = "."
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case a == "--stdout":
			toStdout = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf requirements: unknown flag %q\n", a)
			return "", "", false, false
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf requirements: --feature=NAME is required")
		return "", "", false, false
	}
	return projectDir, feature, toStdout, true
}

func reqRenderCmd(args []string) int {
	projectDir, feature, toStdout, ok := parseReqFlags(args)
	if !ok {
		return 2
	}
	return renderRequirements(projectDir, feature, toStdout)
}

func reqValidateCmd(args []string) int {
	projectDir, feature, _, ok := parseReqFlags(args)
	if !ok {
		return 2
	}
	return validateRequirements(projectDir, feature)
}

// readRequirementsFile lee y parsea requirements.json. Devuelve (rf, exitCode);
// exitCode 0 si está OK.
func readRequirementsFile(projectDir, feature string) (requirementsFile, int) {
	path := filepath.Join(projectDir, "specforge", "features", feature, "requirements.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf requirements: cannot read %s (%v)\n", path, err)
		return requirementsFile{}, 4
	}
	var rf requirementsFile
	if err := json.Unmarshal(data, &rf); err != nil {
		fmt.Fprintf(os.Stderr, "sf requirements: invalid JSON (%v)\n", err)
		return requirementsFile{}, 2
	}
	return rf, 0
}

func renderRequirements(projectDir, feature string, toStdout bool) int {
	rf, code := readRequirementsFile(projectDir, feature)
	if code != 0 {
		return code
	}

	var buf strings.Builder
	if err := reqTmpl.Execute(&buf, rf); err != nil {
		fmt.Fprintf(os.Stderr, "sf requirements: render failed (%v)\n", err)
		return 1
	}

	if toStdout {
		fmt.Print(buf.String())
		return 0
	}
	mdPath := filepath.Join(projectDir, "specforge", "features", feature, "requirements.md")
	if err := os.WriteFile(mdPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf requirements: cannot write %s (%v)\n", mdPath, err)
		return 1
	}
	fmt.Printf("rendered %s\n", filepath.Join("specforge", "features", feature, "requirements.md"))
	return 0
}

func validateRequirements(projectDir, feature string) int {
	rf, code := readRequirementsFile(projectDir, feature)
	if code != 0 {
		return code
	}

	var rep report
	checkRequirements(rf, &rep)

	fmt.Printf("Validated %s: %d requirement(s).\n", feature, len(rf.Requirements))
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
	fmt.Printf("\nOK: requirements.json valid (%d warning(s)).\n", len(rep.warnings))
	return 0
}

// checkRequirements valida la gramática EARS (CLI-SPEC §4.2): cada tipo exige
// ciertos campos y prohíbe otros.
func checkRequirements(rf requirementsFile, rep *report) {
	if rf.Feature == "" {
		rep.errorf("missing `feature`")
	}

	seen := map[string]bool{}
	for _, r := range rf.Requirements {
		switch {
		case r.ID == "":
			rep.errorf("requirement with empty id")
		case seen[r.ID]:
			rep.errorf("duplicate requirement id %s", r.ID)
		default:
			seen[r.ID] = true
		}

		if !earsTypes[r.EarsType] {
			rep.errorf("%s: invalid ears_type %q", r.ID, r.EarsType)
			continue // sin tipo válido no tiene sentido chequear los campos
		}

		// Reglas EARS por tipo.
		switch r.EarsType {
		case "event":
			if r.Trigger == "" {
				rep.errorf("%s: event requires a trigger", r.ID)
			}
			if r.State != "" {
				rep.errorf("%s: event must not set state", r.ID)
			}
		case "state":
			if r.State == "" {
				rep.errorf("%s: state requires a state", r.ID)
			}
			if r.Trigger != "" {
				rep.errorf("%s: state must not set trigger", r.ID)
			}
		case "error":
			if r.Trigger == "" {
				rep.errorf("%s: error requires a trigger (condition)", r.ID)
			}
			if r.Behavior == "" {
				rep.errorf("%s: error requires a behavior (response)", r.ID)
			}
		case "ubiquitous":
			if r.Trigger != "" || r.State != "" {
				rep.errorf("%s: ubiquitous must not set trigger/state", r.ID)
			}
			if r.Behavior == "" {
				rep.errorf("%s: ubiquitous requires a behavior", r.ID)
			}
		}
	}
}
