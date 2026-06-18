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

// runContext es el punto de entrada de `sf context ...`. Hoy la única sub-acción
// es `for-wave`.
func runContext(args []string) int {
	if len(args) == 0 || args[0] != "for-wave" {
		fmt.Fprintln(os.Stderr, "usage: sf context for-wave --feature=NAME --n=N [project_dir]")
		return 2
	}
	args = args[1:] // descartamos "for-wave"

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
		fmt.Fprintln(os.Stderr, "usage: sf context for-wave --feature=NAME --n=N [project_dir]")
		return 2
	}
	return contextForWave(projectDir, feature, n)
}

// contextForWave lee tasks.json, computa el slice y lo imprime como JSON.
func contextForWave(projectDir, feature string, n int) int {
	jsonPath := filepath.Join(projectDir, "specforge", "features", feature, "tasks.json")
	data, err := os.ReadFile(jsonPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf context: cannot read %s (%v)\n", jsonPath, err)
		return 4
	}
	var tf tasksFile
	if err := json.Unmarshal(data, &tf); err != nil {
		fmt.Fprintf(os.Stderr, "sf context: invalid tasks.json (%v)\n", err)
		return 2
	}

	// requirements.json es OPCIONAL: si existe, el slice trae los cuerpos de los
	// R# en scope; si no, solo los IDs (con una nota).
	reqByID := loadRequirementsMap(projectDir, feature)

	ctx, ok := buildWaveContext(tf, reqByID, feature, n)
	if !ok {
		fmt.Fprintf(os.Stderr, "sf context: wave %d not found in %s\n", n, feature)
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

// buildWaveContext computa el slice (función pura → fácil de testear). reqByID
// puede ser nil (no hay requirements.json): en ese caso solo van los IDs.
// Devuelve (ctx, false) si la wave N no existe.
func buildWaveContext(tf tasksFile, reqByID map[string]requirement, feature string, n int) (waveContext, bool) {
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

	// Si tenemos requirements.json, adjuntamos los CUERPOS de los R# en scope —
	// el ahorro real: el LLM recibe solo esos requirements, no el archivo entero.
	var bodies []requirement
	for _, id := range reqRefs {
		if r, ok := reqByID[id]; ok {
			bodies = append(bodies, r)
		}
	}

	// Nota honesta solo cuando faltan los cuerpos.
	note := ""
	if len(reqByID) == 0 {
		note = "requirements.json not found — only IDs in scope (bodies still in markdown)."
	}

	return waveContext{
		Feature:         feature,
		Wave:            *target,
		RequirementRefs: reqRefs,
		Requirements:    bodies,
		ComponentRefs:   uniqSorted(comps),
		FilesInScope:    uniqSorted(files),
		PriorWaves:      prior,
		Note:            note,
	}, true
}
