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

// Modelo de plan.json (CLI-SPEC §4.5). Ahora es EL artefacto de waves: la
// membresía task→wave (campo Tasks) la COMPUTA `sf plan compute` desde el grafo
// de depends_on de tasks.json; Name/Complexity/Rationale las refina el LLM.
type planFile struct {
	SchemaVersion string     `json:"schema_version"`
	Feature       string     `json:"feature"`
	Waves         []planWave `json:"waves"`
}

type planWave struct {
	N          int      `json:"n"`
	Tasks      []string `json:"tasks"`                // IDs en esta wave (computado)
	Name       string   `json:"name,omitempty"`       // tema (LLM refina)
	Complexity string   `json:"complexity,omitempty"` // LLM refina
	Rationale  string   `json:"rationale,omitempty"`  // LLM refina
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
	case "compute":
		projectDir, feature, _, ok := parseArtifactFlags("plan", rest)
		if !ok {
			return 2
		}
		return computePlan(projectDir, feature)
	default:
		fmt.Fprintf(os.Stderr, "sf plan: unknown action %q (use render|validate|compute)\n", action)
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
	// El guard (wave > deps) y la cobertura de tasks necesitan tasks.json.
	tf, tcode := readTasksFile(projectDir, feature)
	if tcode != 0 {
		return tcode
	}
	var rep report
	checkPlan(pf, tf, &rep)

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

// checkPlanInternal valida lo que se ve SOLO en plan.json (sin tasks.json): N
// únicos y complexity válida. Es lo que puede chequear `sf save plan`.
func checkPlanInternal(pf planFile, rep *report) {
	if pf.Feature == "" {
		rep.errorf("missing `feature`")
	}
	seenN := map[int]bool{}
	for _, w := range pf.Waves {
		if seenN[w.N] {
			rep.errorf("duplicate wave number %d", w.N)
		}
		seenN[w.N] = true
		if w.Complexity != "" && !complexities[w.Complexity] {
			rep.errorf("wave %d: invalid complexity %q (low|medium|high)", w.N, w.Complexity)
		}
	}
}

// checkPlan valida el plan CONTRA tasks.json: lo interno + membresía (cada task
// en exactamente una wave, sin ids desconocidos) + el GUARD de dependencias
// (una task no puede estar en una wave anterior o igual a la de sus deps).
func checkPlan(pf planFile, tf tasksFile, rep *report) {
	checkPlanInternal(pf, rep)

	allTasks := map[string]bool{}
	for _, t := range tf.Tasks {
		allTasks[t.ID] = true
	}

	// id de task → su wave en el plan; detecta doble-membresía.
	taskWave := map[string]int{}
	for _, w := range pf.Waves {
		for _, id := range w.Tasks {
			if !allTasks[id] {
				rep.errorf("wave %d: unknown task %q", w.N, id)
				continue
			}
			if prev, dup := taskWave[id]; dup {
				rep.errorf("task %s is in waves %d and %d (must be in exactly one)", id, prev, w.N)
			}
			taskWave[id] = w.N
		}
	}

	// Toda task debe estar asignada a alguna wave.
	for _, t := range tf.Tasks {
		if _, ok := taskWave[t.ID]; !ok {
			rep.errorf("task %s is not assigned to any wave", t.ID)
		}
	}

	// GUARD (tier hermético): wave(t) > wave(d) para cada dep d de t.
	for _, t := range tf.Tasks {
		wt, ok := taskWave[t.ID]
		if !ok {
			continue
		}
		for _, d := range t.DependsOn {
			if wd, ok := taskWave[d]; ok && wt <= wd {
				rep.errorf("%s (wave %d) depends on %s (wave %d): a task must be in a later wave than its dependencies", t.ID, wt, d, wd)
			}
		}
	}
}

// computePlan es `sf plan compute`: deriva el layout de waves desde el grafo de
// depends_on de tasks.json y lo escribe a plan.json (preservando la refinación
// del LLM por wave). Acá vive la separación juicio/cálculo: propose autora las
// deps (juicio), sf computa las waves (cálculo).
func computePlan(projectDir, feature string) int {
	tf, code := readTasksFile(projectDir, feature)
	if code != 0 {
		return code
	}

	// Validar tasks ANTES de computar: si hay deps colgadas o un ciclo, no hay
	// grafo acíclico que layerizar.
	var rep report
	checkTasks(tf, &rep)
	if len(rep.errors) > 0 {
		for _, e := range rep.errors {
			fmt.Printf("  ERROR:   %s\n", e)
		}
		fmt.Printf("\nFAIL: cannot compute plan — fix tasks.json (%d error(s)).\n", len(rep.errors))
		return 2
	}

	waveOf, cycle := computeTaskWaves(tf.Tasks)
	if cycle != "" { // defensivo: checkTasks ya lo habría cazado
		fmt.Fprintf(os.Stderr, "sf plan: dependency cycle involving %q\n", cycle)
		return 2
	}

	// Agrupar tasks por wave, preservando el orden de tasks.json dentro de cada una.
	maxWave := 0
	for _, w := range waveOf {
		if w > maxWave {
			maxWave = w
		}
	}
	groups := make([][]string, maxWave+1)
	for _, t := range tf.Tasks {
		w := waveOf[t.ID]
		groups[w] = append(groups[w], t.ID)
	}

	// MERGE: conservar Name/Complexity/Rationale del plan previo por número de wave.
	prev := map[int]planWave{}
	for _, w := range readPlanQuiet(projectDir, feature).Waves {
		prev[w.N] = w
	}

	pf := planFile{SchemaVersion: tf.SchemaVersion, Feature: feature}
	for n := 0; n <= maxWave; n++ {
		if len(groups[n]) == 0 {
			continue
		}
		pw := planWave{N: n, Tasks: groups[n]}
		if p, ok := prev[n]; ok {
			pw.Name, pw.Complexity, pw.Rationale = p.Name, p.Complexity, p.Rationale
		}
		pf.Waves = append(pf.Waves, pw)
	}

	if code := writePlan(projectDir, feature, pf); code != 0 {
		return code
	}
	fmt.Printf("computed plan for %s: %d task(s) → %d wave(s)\n", feature, len(tf.Tasks), len(pf.Waves))
	return 0
}

// readPlanQuiet lee plan.json sin imprimir nada (para el merge: si no existe,
// devuelve vacío). A diferencia de readPlanFile, no es ruidoso.
func readPlanQuiet(projectDir, feature string) planFile {
	path := filepath.Join(projectDir, "specforge", "features", feature, "progress", "plan.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return planFile{}
	}
	var pf planFile
	if json.Unmarshal(data, &pf) != nil {
		return planFile{}
	}
	return pf
}

// writePlan escribe el plan.json canónico y renderiza el plan.md.
func writePlan(projectDir, feature string, pf planFile) int {
	dir := filepath.Join(projectDir, "specforge", "features", feature, "progress")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf plan: %v\n", err)
		return 1
	}
	out, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf plan: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(filepath.Join(dir, "plan.json"), append(out, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf plan: cannot write plan.json (%v)\n", err)
		return 1
	}
	var buf strings.Builder
	if err := planTmpl.Execute(&buf, pf); err != nil {
		fmt.Fprintf(os.Stderr, "sf plan: render failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(filepath.Join(dir, "plan.md"), []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf plan: cannot write plan.md (%v)\n", err)
		return 1
	}
	return 0
}
