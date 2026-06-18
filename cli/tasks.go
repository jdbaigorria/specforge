package main

import (
	_ "embed" // habilita la directiva //go:embed de abajo
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// ----------------------------------------------------------------------------
// PILOTO JSON-first (release/v1.1): para el artefacto `tasks`, el JSON es la
// FUENTE y el markdown es un RENDER generado. No se editan los dos. Esto NO toca
// requirements/design/constitution — es aditivo y reversible.
// ----------------------------------------------------------------------------

// Modelo de tasks.json (CLI-SPEC §4.4).
type tasksFile struct {
	SchemaVersion string `json:"schema_version"`
	Feature       string `json:"feature"`
	Waves         []wave `json:"waves"`
}

type wave struct {
	N     int    `json:"n"`
	Name  string `json:"name"`
	Tasks []task `json:"tasks"`
}

type task struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	RequirementRefs []string `json:"requirement_refs"`
	ComponentRefs   []string `json:"component_refs"`
	FilesTouched    []string `json:"files_touched"`
	EstimatedEffort string   `json:"estimated_effort"`
	Status          string   `json:"status"`
}

// go:embed mete el contenido del archivo DENTRO del binario en tiempo de
// compilación. tasksTemplate queda como un string normal; no hace falta leer el
// archivo en runtime (el binario es autocontenido). La directiva debe ir pegada
// a la variable.
//
//go:embed templates/tasks.tmpl.md
var tasksTemplate string

// tasksTmpl se parsea una sola vez al arrancar. template.Must envuelve Parse y
// hace panic si la plantilla tiene un error de sintaxis — correcto para una
// plantilla constante (un error ahí es un bug, no algo de runtime).
//
// Funcs registra funciones que la plantilla puede llamar (text/template no trae
// un "join" propio): `list` une un slice o muestra "—", `orDash` igual para un
// string suelto. Ambas las reusamos de status.go.
var tasksTmpl = template.Must(
	template.New("tasks").
		Funcs(template.FuncMap{
			"list":   joinOrDash,
			"orDash": orDash,
		}).
		Parse(tasksTemplate),
)

// runTasks es el punto de entrada de `sf tasks ...`. Despacha a la sub-acción
// (render | validate).
func runTasks(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf tasks <render|validate> --feature=NAME [project_dir]")
		return 2
	}
	action, rest := args[0], args[1:]
	switch action {
	case "render":
		return tasksRenderCmd(rest)
	case "validate":
		return tasksValidateCmd(rest)
	default:
		fmt.Fprintf(os.Stderr, "sf tasks: unknown action %q (use render|validate)\n", action)
		return 2
	}
}

// tasksRenderCmd parsea los flags de `sf tasks render`.
func tasksRenderCmd(args []string) int {
	projectDir := "."
	feature := ""
	toStdout := false
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case a == "--stdout":
			toStdout = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf tasks: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf tasks render: --feature=NAME is required")
		return 2
	}
	return renderTasks(projectDir, feature, toStdout)
}

// tasksValidateCmd parsea los flags de `sf tasks validate`.
func tasksValidateCmd(args []string) int {
	projectDir := "."
	feature := ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf tasks: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf tasks validate: --feature=NAME is required")
		return 2
	}
	return validateTasks(projectDir, feature)
}

// renderTasks lee tasks.json y escribe (o imprime) tasks.md. Con --stdout no
// escribe nada: imprime el render, útil para previsualizar o testear.
func renderTasks(projectDir, feature string, toStdout bool) int {
	dir := filepath.Join(projectDir, "specforge", "features", feature)
	jsonPath := filepath.Join(dir, "tasks.json")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf tasks: cannot read %s (%v)\n", jsonPath, err)
		return 4
	}
	var tf tasksFile
	if err := json.Unmarshal(data, &tf); err != nil {
		fmt.Fprintf(os.Stderr, "sf tasks: invalid tasks.json (%v)\n", err)
		return 2
	}

	out, err := renderTasksMarkdown(tf)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf tasks: render failed (%v)\n", err)
		return 1
	}

	if toStdout {
		fmt.Print(out)
		return 0
	}
	mdPath := filepath.Join(dir, "tasks.md")
	if err := os.WriteFile(mdPath, []byte(out), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf tasks: cannot write %s (%v)\n", mdPath, err)
		return 1
	}
	fmt.Printf("rendered %s\n", filepath.Join("specforge", "features", feature, "tasks.md"))
	return 0
}

// renderTasksMarkdown ejecuta la plantilla contra los datos. strings.Builder es
// un acumulador de strings eficiente que implementa io.Writer, así que
// Execute escribe directo en él.
func renderTasksMarkdown(tf tasksFile) (string, error) {
	var buf strings.Builder
	if err := tasksTmpl.Execute(&buf, tf); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// ----------------------------------------------------------------------------
// Validación de tasks.json.
//
// Como en el piloto solo `tasks` es JSON-first (requirements/design siguen en
// markdown), validamos la CONSISTENCIA INTERNA del archivo. El cruce de R#/C#
// contra requirements.json/design.json queda para cuando esos también sean
// estructurados; por ahora solo chequeamos el FORMATO de las refs (warning).
// ----------------------------------------------------------------------------

// taskStatuses es el enum válido de status (CLI-SPEC §4.4), como set.
var taskStatuses = map[string]bool{
	"pending": true, "in_progress": true, "done": true, "failed": true, "skipped": true,
}

var (
	reqRefRe  = regexp.MustCompile(`^R\d+$`) // requirement: R1, R2, ...
	compRefRe = regexp.MustCompile(`^C\d+$`) // component:   C1, C2, ...
)

// validateTasks lee tasks.json y reporta problemas. Exit 2 si hay errores
// (schema inválido), 4 si el archivo no existe, 0 si está OK.
func validateTasks(projectDir, feature string) int {
	jsonPath := filepath.Join(projectDir, "specforge", "features", feature, "tasks.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf tasks: cannot read %s (%v)\n", jsonPath, err)
		return 4
	}
	var tf tasksFile
	if err := json.Unmarshal(data, &tf); err != nil {
		fmt.Fprintf(os.Stderr, "sf tasks: invalid JSON (%v)\n", err)
		return 2
	}

	// Reusamos el struct `report` de lint.go (mismo paquete): errores + warnings.
	var rep report
	checkTasks(tf, &rep)

	fmt.Printf("Validated %s: %d wave(s).\n", feature, len(tf.Waves))
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
	fmt.Printf("\nOK: tasks.json valid (%d warning(s)).\n", len(rep.warnings))
	return 0
}

// checkTasks corre los chequeos de consistencia interna sobre tasks.json.
func checkTasks(tf tasksFile, rep *report) {
	if tf.Feature == "" {
		rep.errorf("missing `feature`")
	}

	seenWave := map[int]bool{}
	seenTask := map[string]bool{} // ids únicos a través de TODAS las waves

	for _, w := range tf.Waves {
		if seenWave[w.N] {
			rep.errorf("duplicate wave number %d", w.N)
		}
		seenWave[w.N] = true

		for _, tk := range w.Tasks {
			switch {
			case tk.ID == "":
				rep.errorf("wave %d: task with empty id", w.N)
			case seenTask[tk.ID]:
				rep.errorf("duplicate task id %s", tk.ID)
			default:
				seenTask[tk.ID] = true
			}

			if tk.Title == "" {
				rep.errorf("%s: empty title", tk.ID)
			}
			if tk.Status != "" && !taskStatuses[tk.Status] {
				rep.errorf("%s: invalid status %q", tk.ID, tk.Status)
			}

			// Refs: solo chequeo de formato (warning), no de existencia.
			for _, r := range tk.RequirementRefs {
				if !reqRefRe.MatchString(r) {
					rep.warnf("%s: requirement_ref %q is not in R# form", tk.ID, r)
				}
			}
			for _, c := range tk.ComponentRefs {
				if !compRefRe.MatchString(c) {
					rep.warnf("%s: component_ref %q is not in C# form", tk.ID, c)
				}
			}
		}
	}
}
