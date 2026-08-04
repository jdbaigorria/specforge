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
	// SchemaVersion viaja en el feature.json POR FEATURE (A1/R3); en el legacy
	// features.json el campo va vacío (la versión vivía en el archivo global).
	SchemaVersion string   `json:"schema_version,omitempty"`
	Name          string   `json:"name"`
	Status        string   `json:"status"`
	Lane          string   `json:"lane"`
	DependsOn     []string `json:"depends_on"`
	Gates         []gate   `json:"gates"`
	// Seq es el orden de creación (1, 2, …). Con el estado partido por feature
	// no hay array global que recuerde el orden — lo recuerda cada feature.
	Seq int `json:"seq,omitempty"`
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
	// Prev encadena esta entrada con la ANTERIOR del ledger: es el sha256 de la
	// entrada previa completa (gateEntryHash, integrity.go), o "genesis" en la
	// primera. Forjar/editar una entrada intermedia rompe la cadena de forma
	// visible — es la capa de detección que complementa la prevención del hook.
	// `omitempty`: gates legacy (pre-cadena) no lo traen.
	Prev string `json:"prev,omitempty"`
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
	asJSON := false
	featureFilter := ""
	for _, a := range args {
		switch {
		case a == "--artifacts":
			artifacts = true
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "--feature="):
			featureFilter = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf status: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	// --json cubre la vista principal. La de artefactos ya tiene su propio
	// consumidor máquina (`sf verify --json`), así que en vez de inventarle un
	// segundo esquema, rechazamos la combinación: mejor un error claro que un
	// JSON que nadie definió.
	if artifacts && asJSON {
		fmt.Fprintln(os.Stderr, "sf status: --json does not cover --artifacts — use `sf verify --json` for artifact state")
		return 2
	}

	specforge := filepath.Join(projectDir, "specforge")
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		// Igual que el Python: ausencia no es error, solo no hay nada que mostrar.
		if asJSON {
			return printStatusJSON(statusReport{Features: []statusFeature{}, Statuses: map[string]int{}})
		}
		fmt.Printf("No feature state under %s/specforge.\n", projectDir)
		return 0
	}
	if len(ff.Features) == 0 {
		if asJSON {
			return printStatusJSON(statusReport{Features: []statusFeature{}, Statuses: map[string]int{}})
		}
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

	// --json corta acá: mismo dato que la tabla, en un objeto. El histograma por
	// status es lo que `sf-audit` consume para sus filas de features (DL-9) —
	// contarlas leyendo la tabla es justo lo que el ítem viene a eliminar.
	if asJSON {
		return printStatusJSON(buildStatusReport(rows, feats, order, cycle, ff))
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

	// R1 (integrity.go): la vista de estado también AVISA si algún ledger no
	// valida (status no bloquea — los que bloquean son next/approve/archive —
	// pero el humano tiene que enterarse acá, no al intentar avanzar).
	if probs := ledgerProblemsAll(ff); len(probs) > 0 {
		fmt.Println("\n⚠ LEDGER INTEGRITY — gates modified outside the CLI:")
		for _, p := range probs {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println("  run `sf recover` for the restore plan")
	}
	return 0
}

// ----------------------------------------------------------------------------
// Fila de la tabla y su render.
// ----------------------------------------------------------------------------

type statusRow struct {
	feature, lane, status, phase, drift, gaps, blocked string
}

// ----------------------------------------------------------------------------
// `sf status --json` — la misma vista, machine-readable.
//
// Por qué existe: el Paso 6 de `sf-audit` pedía "features completed / active"
// como un {N} que estimaba un LLM leyendo la tabla. El dato ya está en
// features.json; lo que faltaba era una salida que no hubiera que parsear con
// ojos. `statuses` es el histograma que responde esa fila de una.
// ----------------------------------------------------------------------------

type statusFeature struct {
	Feature string   `json:"feature"`
	Lane    string   `json:"lane"`
	Status  string   `json:"status"`
	Phase   string   `json:"phase"`
	Drift   string   `json:"drift"` // "ok" | "DRIFT" | "—" (sin trace.json)
	Gaps    string   `json:"gaps"`
	Blocked []string `json:"blocked"`
}

type statusReport struct {
	Features []statusFeature `json:"features"`
	// Statuses: histograma {"done":2,"building":1}. Los status que valen 0 no
	// aparecen — el consumidor pregunta por el que le importa y toma el cero.
	Statuses       map[string]int `json:"statuses"`
	CriticalPath   []string       `json:"critical_path"`
	Cycle          []string       `json:"cycle,omitempty"`
	LedgerProblems []string       `json:"ledger_problems,omitempty"`
}

// buildStatusReport traduce las filas ya calculadas al objeto JSON. No recalcula
// nada: la tabla humana y el JSON salen de la MISMA fuente, así que no pueden
// discrepar (dos cálculos paralelos es cómo se llega a que la tabla diga una
// cosa y el JSON otra).
func buildStatusReport(rows []statusRow, feats map[string]*feature, order, cycle []string, ff featuresFile) statusReport {
	rep := statusReport{
		Features:     make([]statusFeature, 0, len(rows)),
		Statuses:     map[string]int{},
		CriticalPath: order,
		Cycle:        uniqSorted(cycle),
	}
	for _, r := range rows {
		var blocked []string
		if f, ok := feats[r.feature]; ok {
			blocked = blockers(f, feats)
		}
		if blocked == nil {
			blocked = []string{}
		}
		rep.Features = append(rep.Features, statusFeature{
			Feature: r.feature,
			Lane:    r.lane,
			Status:  r.status,
			Phase:   r.phase,
			Drift:   r.drift,
			Gaps:    r.gaps,
			Blocked: blocked,
		})
		rep.Statuses[r.status]++
	}
	rep.LedgerProblems = ledgerProblemsAll(ff)
	return rep
}

func printStatusJSON(rep statusReport) int {
	if rep.CriticalPath == nil {
		rep.CriticalPath = []string{}
	}
	out, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf status: marshal failed (%v)\n", err)
		return 1
	}
	fmt.Println(string(out))
	return 0
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
