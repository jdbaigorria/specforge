package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ----------------------------------------------------------------------------
// Modelo de datos: structs que mapean features.json.
//
// El concepto nuevo es encoding/json. Las etiquetas `json:"..."` en cada campo
// le dicen al decoder qué clave del JSON corresponde a qué campo del struct.
// Sin la etiqueta, Go buscaría una clave con el nombre EXACTO del campo (y los
// campos deben ser exportados —Mayúscula inicial— para que json los vea).
// ----------------------------------------------------------------------------

type featuresFile struct {
	SchemaVersion string    `json:"schema_version"`
	Features      []feature `json:"features"`
}

type feature struct {
	Name      string   `json:"name"`
	Status    string   `json:"status"`
	Lane      string   `json:"lane"`
	DependsOn []string `json:"depends_on"`
	Gates     []gate   `json:"gates"`
}

type gate struct {
	Phase   string `json:"phase"`
	Result  string `json:"result"`
	By      string `json:"by"`
	At      string `json:"at"`
	Comment string `json:"comment"`
	// Hash es el sha256 del CONTENIDO canónico del artefacto al momento de
	// aprobarlo (lo sella `sf gate approve`). Permite detectar "silent edits":
	// si el hash actual del artefacto difiere del sellado, el gate ya no
	// certifica el contenido vigente → el artefacto está STALE. `omitempty`:
	// los gates viejos (pre-hash) y los de fases sin artefacto (lane) no lo traen.
	Hash string `json:"hash,omitempty"`
}

// runStatus es el punto de entrada de `sf status [project_dir]`. Renderiza el
// estado de TODO el proyecto desde specforge/features.json. Determinista: es un
// render de datos que ya existen, no un dashboard.
func runStatus(args []string) int {
	// Dir de proyecto: el primer argumento posicional, o el cwd por defecto.
	// --artifacts cambia la vista a los estados de artefactos (stale model);
	// --feature la acota a una feature.
	projectDir := "."
	artifacts := false
	featureFilter := ""
	for _, a := range args {
		switch {
		case a == "--artifacts":
			artifacts = true
		case strings.HasPrefix(a, "--feature="):
			featureFilter = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf status: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	specforge := filepath.Join(projectDir, "specforge")
	data, err := os.ReadFile(filepath.Join(specforge, "features.json"))
	if err != nil {
		// Igual que el Python: ausencia no es error, solo no hay nada que mostrar.
		fmt.Printf("No specforge/features.json under %s.\n", projectDir)
		return 0
	}

	// Unmarshal copia el JSON dentro del struct. Le pasamos &ff (un puntero)
	// para que pueda escribir en él. Devuelve error si el JSON está malformado.
	var ff featuresFile
	if err := json.Unmarshal(data, &ff); err != nil {
		fmt.Fprintf(os.Stderr, "sf status: invalid features.json: %v\n", err)
		return 1
	}
	if len(ff.Features) == 0 {
		fmt.Println("No features registered yet.")
		return 0
	}

	// Vista de artefactos (stale model): una tabla por feature con el estado de
	// cada artefacto de la cadena. --feature la acota a una sola.
	if artifacts {
		return runStatusArtifacts(ff, projectDir, featureFilter)
	}

	// Indexamos por nombre para lookups, y guardamos el orden original aparte
	// (el orden de un map en Go es aleatorio; no podemos confiar en él).
	feats := make(map[string]*feature, len(ff.Features))
	var names []string
	for i := range ff.Features {
		// &ff.Features[i] apunta al elemento real del slice. Si usáramos
		// `for _, f := range ...` tomaríamos la dirección de una copia.
		f := &ff.Features[i]
		feats[f.Name] = f
		names = append(names, f.Name)
	}

	order, cycle := topoOrder(feats, names)

	// Armamos las filas en orden topológico (dependencias primero).
	var rows []statusRow
	for _, n := range order {
		f := feats[n]
		rows = append(rows, statusRow{
			feature: n,
			lane:    orDash(f.Lane),
			status:  orDash(f.Status),
			phase:   lastPhase(f),
			drift:   driftState(projectDir, specforge, n),
			gaps:    gapsState(specforge, n),
			blocked: joinOrDash(blockers(f, feats)),
		})
	}

	printStatusTable(rows)

	fmt.Println("\nCritical path (dependency order):")
	fmt.Println("  " + strings.Join(order, " → "))
	if len(cycle) > 0 {
		fmt.Printf("  ⚠ dependency cycle involving: %s\n", strings.Join(uniqSorted(cycle), ", "))
	}

	// Avisos finales. (drifted queda vacío hasta que portemos el motor de drift
	// en `sf doctor --drift`; driftState devuelve "?"/"—" por ahora.)
	var drifted, blockedMsgs []string
	for _, r := range rows {
		if r.drift == "DRIFT" {
			drifted = append(drifted, r.feature)
		}
		if r.blocked != "—" {
			blockedMsgs = append(blockedMsgs, r.feature+"←"+r.blocked)
		}
	}
	if len(drifted) > 0 {
		fmt.Printf("\n⚠ drifted (spec ≠ code): %s — run sf-amend\n", strings.Join(drifted, ", "))
	}
	if len(blockedMsgs) > 0 {
		fmt.Printf("⚠ blocked: %s\n", strings.Join(blockedMsgs, "; "))
	}
	return 0
}

// ----------------------------------------------------------------------------
// Fila de la tabla y su render.
// ----------------------------------------------------------------------------

type statusRow struct {
	feature, lane, status, phase, drift, gaps, blocked string
}

var statusCols = []string{"feature", "lane", "status", "phase", "drift", "gaps", "blocked"}

// cells devuelve los valores de la fila en el mismo orden que statusCols.
func (r statusRow) cells() []string {
	return []string{r.feature, r.lane, r.status, r.phase, r.drift, r.gaps, r.blocked}
}

// printStatusTable imprime el título y delega la tabla en renderTable (table.go).
func printStatusTable(rows []statusRow) {
	fmt.Print("SpecForge — project status\n\n")
	cells := make([][]string, len(rows))
	for i, r := range rows {
		cells[i] = r.cells()
	}
	renderTable(statusCols, cells)
}

// ----------------------------------------------------------------------------
// Cálculos por feature.
// ----------------------------------------------------------------------------

// lastPhase devuelve la fase del último gate aprobado, o "—" si ninguno.
func lastPhase(f *feature) string {
	phase := "—"
	for _, g := range f.Gates {
		if strings.HasPrefix(g.Result, "approve") {
			phase = g.Phase
		}
	}
	return phase
}

// blockers devuelve las dependencias que aún no están "done" (o que no existen).
func blockers(f *feature, feats map[string]*feature) []string {
	var bs []string
	for _, dep := range f.DependsOn {
		d, ok := feats[dep]
		if !ok || d.Status != "done" {
			bs = append(bs, dep)
		}
	}
	return bs
}

// driftState evalúa el drift de una feature reutilizando el motor de doctor.go.
// "—" = sin trace.json, "ok" = anchors en sync, "DRIFT" = algún anchor divergió.
// Solo chequeo estático (runTests vacío): status debe ser rápido.
func driftState(projectDir, specforge, name string) string {
	tracePath := findArtifact(specforge, name, "trace.json")
	if tracePath == "" {
		return "—"
	}
	if len(checkFeatureDrift(projectDir, name, tracePath, "")) > 0 {
		return "DRIFT"
	}
	return "ok"
}

// gapsState lee la línea de veredicto del review.md (la autoridad), no palabras
// sueltas en la prosa.
func gapsState(specforge, name string) string {
	rv := findArtifact(specforge, name, "review.md")
	if rv == "" {
		return "—"
	}
	data, err := os.ReadFile(rv)
	if err != nil {
		return "—"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if !strings.Contains(strings.ToLower(line), "verdict") {
			continue
		}
		up := strings.ToUpper(line)
		switch {
		case strings.Contains(up, "APPROVE WITH NOTES"):
			return "notes"
		case strings.Contains(up, "REVISE"):
			return "gaps"
		case strings.Contains(up, "APPROVE"):
			return "ok"
		}
	}
	return "—"
}

// findArtifact busca specforge/{archive,features}/*<name>/<filename> y devuelve
// la primera coincidencia cuyo dir contenedor sea o termine en <name>.
func findArtifact(specforge, name, filename string) string {
	for _, base := range []string{"archive", "features"} {
		matches, _ := filepath.Glob(filepath.Join(specforge, base, "*"+name, filename))
		for _, m := range matches {
			parent := filepath.Base(filepath.Dir(m))
			if parent == name || strings.HasSuffix(parent, name) {
				return m
			}
		}
	}
	return ""
}

// ----------------------------------------------------------------------------
// Orden topológico (critical path) + detección de ciclos.
// ----------------------------------------------------------------------------

// topoOrder devuelve los nombres en orden dependencia-primero, y los nombres
// involucrados en algún ciclo. Es un DFS clásico con un "stack" de la rama
// actual para detectar back-edges (ciclos).
func topoOrder(feats map[string]*feature, names []string) (order []string, cycle []string) {
	done := map[string]bool{}

	// visit es recursiva. En Go una función anónima no puede referenciarse a sí
	// misma en su propia expresión de inicialización, así que la declaramos
	// primero (var visit func(...)) y la asignamos después.
	var visit func(n string, stack map[string]bool)
	visit = func(n string, stack map[string]bool) {
		if done[n] {
			return
		}
		if _, ok := feats[n]; !ok {
			return // dependencia hacia una feature inexistente: la ignoramos acá
		}
		if stack[n] {
			cycle = append(cycle, n)
			return
		}
		// Copiamos el stack y agregamos n, para que cada rama tenga el suyo.
		next := map[string]bool{n: true}
		for k := range stack {
			next[k] = true
		}
		for _, dep := range feats[n].DependsOn {
			visit(dep, next)
		}
		done[n] = true
		order = append(order, n)
	}

	for _, n := range names {
		visit(n, map[string]bool{})
	}
	return order, cycle
}

// ----------------------------------------------------------------------------
// Utilidades chicas.
// ----------------------------------------------------------------------------

// orDash devuelve "—" si el string está vacío.
func orDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}

// joinOrDash une con coma, o "—" si el slice está vacío.
func joinOrDash(xs []string) string {
	if len(xs) == 0 {
		return "—"
	}
	return strings.Join(xs, ",")
}

// uniqSorted deduplica y ordena un slice de strings.
func uniqSorted(xs []string) []string {
	set := map[string]bool{}
	for _, x := range xs {
		set[x] = true
	}
	out := make([]string, 0, len(set))
	for x := range set {
		out = append(out, x)
	}
	sort.Strings(out)
	return out
}
