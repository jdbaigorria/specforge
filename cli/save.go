package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ----------------------------------------------------------------------------
// `sf save`: la VÍA DE ESCRITURA validada. El LLM manda el JSON por stdin
// (--json -) o un archivo (--json path); el CLI valida (reusando los check*),
// y solo si pasa, escribe el .json canónico y renderiza el .md. Nada se guarda
// sin validar.
//
// Concepto nuevo: INTERFACES. En vez de repetir el flujo por cada artefacto,
// definimos un contrato `artifact` que los 6 tipos cumplen, y un finishSave
// genérico que trabaja contra ese contrato.
// ----------------------------------------------------------------------------

// artifact es el contrato: "sé validarme y sé renderizarme a markdown". Cualquier
// tipo con estos dos métodos lo cumple — en Go las interfaces se satisfacen de
// forma IMPLÍCITA (no se declara "implements").
type artifact interface {
	validate(*report)
	renderMarkdown() (string, error)
}

// execTemplate corre una plantilla a string (lo que necesitan los métodos render).
func execTemplate(t *template.Template, data any) (string, error) {
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// --- Métodos adaptadores: cada tipo cumple `artifact` delegando en las
// funciones que ya escribimos. Receptor por VALOR → el valor (y el puntero)
// satisfacen la interfaz.

func (v constitutionFile) validate(r *report)              { checkConstitution(v, r) }
func (v constitutionFile) renderMarkdown() (string, error) { return execTemplate(constitutionTmpl, v) }

func (v requirementsFile) validate(r *report)              { checkRequirements(v, r) }
func (v requirementsFile) renderMarkdown() (string, error) { return execTemplate(reqTmpl, v) }

func (v designFile) validate(r *report)              { checkDesign(v, r) }
func (v designFile) renderMarkdown() (string, error) { return execTemplate(designTmpl, v) }

func (v tasksFile) validate(r *report)              { checkTasks(v, r) }
func (v tasksFile) renderMarkdown() (string, error) { return execTemplate(tasksTmpl, v) }

func (v planFile) validate(r *report)              { checkPlan(v, r) }
func (v planFile) renderMarkdown() (string, error) { return execTemplate(planTmpl, v) }

func (v reviewFile) validate(r *report)              { checkReview(v, r) }
func (v reviewFile) renderMarkdown() (string, error) { return execTemplate(reviewTmpl, v) }

// runSave parsea `sf save <artifact> [--feature=X] [--json -|FILE] [dir]`.
func runSave(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf save <artifact> [--feature=NAME] [--json -|FILE] [project_dir]")
		return 2
	}
	name, rest := args[0], args[1:]

	projectDir := "."
	feature := ""
	jsonSrc := "-" // por defecto, stdin
	for i := 0; i < len(rest); i++ {
		a := rest[i]
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--json="):
			jsonSrc = strings.TrimPrefix(a, "--json=")
		case a == "--json":
			if i+1 < len(rest) {
				jsonSrc = rest[i+1]
				i++ // consumimos el valor
			}
		case a != "-" && strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf save: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	// constitution es a nivel proyecto; el resto necesita --feature.
	if name != "constitution" && feature == "" {
		fmt.Fprintln(os.Stderr, "sf save: --feature=NAME is required")
		return 2
	}

	raw, err := readJSONInput(jsonSrc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf save: cannot read input (%v)\n", err)
		return 1
	}

	jsonPath, mdPath := artifactPaths(name, projectDir, feature)

	// Decodificamos en el tipo concreto según el artefacto, y de ahí en más todo
	// es genérico vía la interfaz.
	switch name {
	case "constitution":
		var v constitutionFile
		if !decodeInto(raw, &v) {
			return 2
		}
		return finishSave(v, jsonPath, mdPath)
	case "requirements":
		var v requirementsFile
		if !decodeInto(raw, &v) {
			return 2
		}
		return finishSave(v, jsonPath, mdPath)
	case "design":
		var v designFile
		if !decodeInto(raw, &v) {
			return 2
		}
		return finishSave(v, jsonPath, mdPath)
	case "tasks":
		var v tasksFile
		if !decodeInto(raw, &v) {
			return 2
		}
		return finishSave(v, jsonPath, mdPath)
	case "plan":
		var v planFile
		if !decodeInto(raw, &v) {
			return 2
		}
		return finishSave(v, jsonPath, mdPath)
	case "review":
		var v reviewFile
		if !decodeInto(raw, &v) {
			return 2
		}
		return finishSave(v, jsonPath, mdPath)
	default:
		fmt.Fprintf(os.Stderr, "sf save: unknown artifact %q\n", name)
		return 2
	}
}

// decodeInto desempaqueta el JSON e informa el error (devuelve false si falla).
func decodeInto(raw []byte, v any) bool {
	if err := json.Unmarshal(raw, v); err != nil {
		fmt.Fprintf(os.Stderr, "sf save: invalid JSON (%v)\n", err)
		return false
	}
	return true
}

// finishSave es el corazón genérico: valida, y solo si pasa, escribe el JSON
// canónico + el markdown. Trabaja contra la INTERFAZ, no contra un tipo concreto.
func finishSave(a artifact, jsonPath, mdPath string) int {
	var rep report
	a.validate(&rep)
	for _, w := range rep.warnings {
		fmt.Printf("  warning: %s\n", w)
	}
	if len(rep.errors) > 0 {
		for _, e := range rep.errors {
			fmt.Printf("  ERROR:   %s\n", e)
		}
		fmt.Printf("\nFAIL: not saved — %d error(s).\n", len(rep.errors))
		return 2
	}

	if err := os.MkdirAll(filepath.Dir(jsonPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf save: %v\n", err)
		return 1
	}

	// Re-serializamos desde el struct → JSON canónico (formato consistente,
	// descarta claves desconocidas). Marshal funciona sobre el valor que lleva
	// adentro la interfaz.
	canonical, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf save: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(jsonPath, append(canonical, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf save: cannot write %s (%v)\n", jsonPath, err)
		return 1
	}

	md, err := a.renderMarkdown()
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf save: render failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf save: cannot write %s (%v)\n", mdPath, err)
		return 1
	}

	fmt.Printf("saved %s (+ rendered %s)\n", jsonPath, mdPath)
	return 0
}

// readJSONInput lee de stdin ("-") o de un archivo.
func readJSONInput(src string) ([]byte, error) {
	if src == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(src)
}

// artifactPaths devuelve (jsonPath, mdPath) según el artefacto. constitution es
// a nivel proyecto; plan vive bajo progress/; el resto bajo features/<f>/.
func artifactPaths(name, projectDir, feature string) (string, string) {
	fdir := filepath.Join(projectDir, "specforge", "features", feature)
	switch name {
	case "constitution":
		base := filepath.Join(projectDir, "specforge", "constitution")
		return base + ".json", base + ".md"
	case "plan":
		base := filepath.Join(fdir, "progress", "plan")
		return base + ".json", base + ".md"
	default: // requirements, design, tasks, review
		base := filepath.Join(fdir, name)
		return base + ".json", base + ".md"
	}
}
