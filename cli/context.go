package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ----------------------------------------------------------------------------
// La "killer query": en vez de que el LLM lea el estado completo, le damos el
// slice MÍNIMO para ejecutar la wave N. Este comando EMITE JSON (a diferencia de
// los otros, que imprimen tablas para humanos) porque su consumidor es el LLM.
//
// En el piloto solo `tasks` es estructurado, así que el slice trae la wave + los
// IDs de requirements/components en scope + archivos + un resumen de las waves
// previas. Los CUERPOS de R#/C# siguen en markdown; eso queda para cuando
// extendamos el piloto a requirements/design.
// ----------------------------------------------------------------------------

type waveContext struct {
	Feature         string             `json:"feature"`
	Wave            wave               `json:"wave"`
	RequirementRefs []string           `json:"requirement_refs_in_scope"`
	Requirements    []requirement      `json:"requirements_in_scope,omitempty"` // cuerpos, si hay requirements.json
	ComponentRefs   []string           `json:"component_refs_in_scope"`
	Components      []component        `json:"components_in_scope,omitempty"` // cuerpos, si hay design.json
	FilesInScope    []string           `json:"files_in_scope"`
	PriorWaves      []priorWaveSummary `json:"prior_waves"`
	Note            string             `json:"note,omitempty"`
}

type priorWaveSummary struct {
	N         int            `json:"n"`
	Name      string         `json:"name"`
	TaskCount int            `json:"task_count"`
	Statuses  map[string]int `json:"statuses"` // histograma: {"done":2,"pending":1}
}

// runContext es el punto de entrada de `sf context ...`. Despacha por
// sub-comando: `for-wave` (slice de una wave puntual) y `current` (slice de la
// fase/wave ACTUAL, derivada de sf state current — vive en state.go).
func runContext(args []string) int {
	if len(args) == 0 {
		contextUsage()
		return 2
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "for-wave":
		return contextForWaveCmd(rest)
	case "current":
		return runContextCurrent(rest) // definido en state.go: cuelga de computeCurrentState
	case "for-judge":
		return contextForJudgeCmd(rest) // definido en judge.go: material del auditor de fase
	default:
		fmt.Fprintf(os.Stderr, "sf context: unknown sub-command %q\n\n", sub)
		contextUsage()
		return 2
	}
}

func contextUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf context for-wave --feature=NAME --n=N [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf context current [--breadcrumb] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf context for-judge --phase=PHASE --feature=NAME [project_dir]")
}

// contextForWaveCmd parsea los flags de `for-wave` y emite el slice de la wave N.
func contextForWaveCmd(args []string) int {
	projectDir := "."
	feature := ""
	n := -1
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--n="):
			// strconv.Atoi: string → int (devuelve error si no es número).
			v, err := strconv.Atoi(strings.TrimPrefix(a, "--n="))
			if err != nil {
				fmt.Fprintf(os.Stderr, "sf context: --n must be a number, got %q\n", strings.TrimPrefix(a, "--n="))
				return 2
			}
			n = v
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf context: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" || n < 0 {
		contextUsage()
		return 2
	}
	return contextForWave(projectDir, feature, n)
}

// contextForWave computa el slice de la wave y lo imprime como JSON.
func contextForWave(projectDir, feature string, n int) int {
	ctx, ok := loadWaveContext(projectDir, feature, n)
	if !ok {
		fmt.Fprintf(os.Stderr, "sf context: cannot build wave %d slice for %s (missing/invalid tasks.json, or wave not found)\n", n, feature)
		return 4
	}
	// MarshalIndent serializa con sangría (lo opuesto de Unmarshal). "" de prefijo
	// y "  " de sangría = JSON pretty de 2 espacios.
	out, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf context: marshal failed (%v)\n", err)
		return 1
	}
	fmt.Println(string(out))
	return 0
}

// loadWaveContext lee tasks.json (+ requirements.json/design.json OPCIONALES) y
// computa el slice de la wave n. Función reutilizable: la usan tanto `for-wave`
// como `current`. Devuelve (_, false) si no hay/no parsea tasks.json o la wave
// no existe — degradación segura para que `current` no explote en esos casos.
func loadWaveContext(projectDir, feature string, n int) (waveContext, bool) {
	jsonPath := filepath.Join(projectDir, "specforge", "features", feature, "tasks.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		return waveContext{}, false
	}
	var tf tasksFile
	if json.Unmarshal(data, &tf) != nil {
		return waveContext{}, false
	}
	// Si existen, el slice trae los CUERPOS de los R#/C# en scope; si no, solo IDs.
	reqByID := loadRequirementsMap(projectDir, feature)
	compByID := loadDesignMap(projectDir, feature)
	return buildWaveContext(tf, reqByID, compByID, feature, n)
}

// loadRequirementsMap lee requirements.json y devuelve un índice id→requirement.
// Es opcional: si no existe o no parsea, devuelve nil (sin error: el slice
// degrada a solo-IDs).
func loadRequirementsMap(projectDir, feature string) map[string]requirement {
	path := filepath.Join(projectDir, "specforge", "features", feature, "requirements.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var rf requirementsFile
	if json.Unmarshal(data, &rf) != nil {
		return nil
	}
	m := make(map[string]requirement, len(rf.Requirements))
	for _, r := range rf.Requirements {
		m[r.ID] = r
	}
	return m
}

// loadDesignMap lee design.json (opcional) y devuelve un índice id→component, o
// nil si no existe / no parsea.
func loadDesignMap(projectDir, feature string) map[string]component {
	path := filepath.Join(projectDir, "specforge", "features", feature, "design.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var df designFile
	if json.Unmarshal(data, &df) != nil {
		return nil
	}
	m := make(map[string]component, len(df.Components))
	for _, c := range df.Components {
		m[c.ID] = c
	}
	return m
}

// buildWaveContext computa el slice (función pura → fácil de testear). reqByID y
// compByID pueden ser nil (no hay requirements.json/design.json): en ese caso
// van solo los IDs. Devuelve (ctx, false) si la wave N no existe.
func buildWaveContext(tf tasksFile, reqByID map[string]requirement, compByID map[string]component, feature string, n int) (waveContext, bool) {
	// Buscamos la wave objetivo.
	var target *wave
	for i := range tf.Waves {
		if tf.Waves[i].N == n {
			target = &tf.Waves[i]
			break
		}
	}
	if target == nil {
		return waveContext{}, false
	}

	// Juntamos refs y archivos de las tasks de ESTA wave. El `...` expande cada
	// slice en los argumentos de append.
	var reqs, comps, files []string
	for _, tk := range target.Tasks {
		reqs = append(reqs, tk.RequirementRefs...)
		comps = append(comps, tk.ComponentRefs...)
		files = append(files, tk.FilesTouched...)
	}

	// Resumen compacto de las waves anteriores (n menor): conteo + histograma de
	// status. El LLM no necesita el detalle de lo ya hecho, solo el panorama.
	prior := []priorWaveSummary{}
	for _, w := range tf.Waves {
		if w.N >= n {
			continue
		}
		statuses := map[string]int{}
		for _, tk := range w.Tasks {
			s := tk.Status
			if s == "" {
				s = "pending"
			}
			statuses[s]++
		}
		prior = append(prior, priorWaveSummary{N: w.N, Name: w.Name, TaskCount: len(w.Tasks), Statuses: statuses})
	}
	// sort.Slice ordena in-place con una función "less" (i va antes que j?).
	sort.Slice(prior, func(i, j int) bool { return prior[i].N < prior[j].N })

	reqRefs := uniqSorted(reqs)
	compRefs := uniqSorted(comps)

	// Si tenemos requirements.json/design.json, adjuntamos los CUERPOS de los
	// R#/C# en scope — el ahorro real: el LLM recibe solo esos, no el archivo
	// entero.
	var reqBodies []requirement
	for _, id := range reqRefs {
		if r, ok := reqByID[id]; ok {
			reqBodies = append(reqBodies, r)
		}
	}
	var compBodies []component
	for _, id := range compRefs {
		if c, ok := compByID[id]; ok {
			compBodies = append(compBodies, c)
		}
	}

	// Nota honesta listando qué artefactos faltan (solo van sus IDs).
	var missing []string
	if len(reqByID) == 0 {
		missing = append(missing, "requirements.json")
	}
	if len(compByID) == 0 {
		missing = append(missing, "design.json")
	}
	note := ""
	if len(missing) > 0 {
		note = "not structured (only IDs in scope): " + strings.Join(missing, ", ")
	}

	return waveContext{
		Feature:         feature,
		Wave:            *target,
		RequirementRefs: reqRefs,
		Requirements:    reqBodies,
		ComponentRefs:   compRefs,
		Components:      compBodies,
		FilesInScope:    uniqSorted(files),
		PriorWaves:      prior,
		Note:            note,
	}, true
}
