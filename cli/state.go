package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf state current` — el "¿dónde estoy?" del proyecto.
//
// Es el CEREBRO que consultan los hooks: les dice qué feature está activa y en
// qué fase/wave, para saber qué slice inyectar. La decisión de diseño (#4 del
// debate JSON-first) es que NO guarda estado nuevo: TODO se deriva de lo que ya
// existe en features.json. En particular:
//
//   feature activa = la única con status "en curso" (flujo SERIAL, F22)
//   fase actual    = frontera-aprobada + 1 (la siguiente fase tras el último
//                    gate aprobado, según el pipeline canónico)
//   wave           = derivada de qué wave-N gates están aprobados (solo en build)
//
// Al no guardar "fase actual" como campo aparte evitamos la "segunda verdad"
// (F2): el array de gates YA es el puntero de fase.
// ----------------------------------------------------------------------------

// currentState es lo que emitimos como JSON. A diferencia de status.go (que
// imprime una tabla para humanos), acá el consumidor es un hook/LLM, así que
// emitimos JSON.
type currentState struct {
	Feature          string   `json:"feature"`
	Status           string   `json:"status"`
	Phase            string   `json:"phase"`
	Wave             *int     `json:"wave,omitempty"` // puntero: nil = no aplica. Un int normal con omitempty ocultaría la wave 0, que es válida.
	LastApprovedGate string   `json:"last_approved_gate"`
	BlockedBy        []string `json:"blocked_by,omitempty"`
	Note             string   `json:"note,omitempty"`
}

// activeStatuses: los estados en que una feature cuenta como "en curso". El
// vocabulario sale de los skills: planned → approved → building → done.
// `planned` es backlog y `done` está archivada; ninguno es "en curso".
// Flujo SERIAL: a lo sumo UNA feature debería estar en estos estados a la vez.
var activeStatuses = map[string]bool{"approved": true, "building": true}

// phasePipeline: el orden canónico de fases. "build" es especial — sus gates en
// features.json son wave-0, wave-1, … (los normalizamos a "build"). La fase
// actual se computa como "la siguiente después de la última aprobada".
var phasePipeline = []string{"lane", "requirements", "design", "tasks", "plan", "build", "verdict"}

// runState es el punto de entrada de `sf state ...`. Hoy la única sub-acción es
// `current`.
func runState(args []string) int {
	if len(args) == 0 || args[0] != "current" {
		fmt.Fprintln(os.Stderr, "usage: sf state current [project_dir]")
		return 2
	}
	args = args[1:] // descartamos "current"

	projectDir := "."
	for _, a := range args {
		if strings.HasPrefix(a, "-") {
			fmt.Fprintf(os.Stderr, "sf state: unknown flag %q\n", a)
			return 2
		}
		projectDir = a
	}
	return stateCurrent(projectDir)
}

// stateCurrent lee el estado de features, computa el estado y lo imprime como JSON.
func stateCurrent(projectDir string) int {
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf state: cannot read feature state under %s (%v)\n", projectDir, err)
		return 4
	}

	st := computeCurrentState(ff, projectDir)
	out, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf state: marshal failed (%v)\n", err)
		return 1
	}
	fmt.Println(string(out))
	return 0
}

// computeCurrentState es la función PURA (fácil de testear): dado features.json
// y el projectDir (para leer tasks.json si hace falta contar waves), devuelve el
// estado actual. No imprime ni sale: solo computa.
func computeCurrentState(ff featuresFile, projectDir string) currentState {
	// 1) Feature activa: la(s) que estén en un estado "en curso".
	var active []*feature
	for i := range ff.Features {
		// &ff.Features[i]: el elemento real del slice, no una copia del range.
		if activeStatuses[ff.Features[i].Status] {
			active = append(active, &ff.Features[i])
		}
	}

	switch len(active) {
	case 0:
		// Respuesta válida, no error: no hay nada en curso.
		return currentState{Phase: "—", Note: "no active feature (flujo serial: nada en curso)"}
	case 1:
		// Caso normal: seguimos abajo.
	default:
		// El invariante serial (gate estructural) debería haberlo prevenido. Si
		// igual pasó, lo reportamos en vez de elegir uno al azar.
		names := make([]string, len(active))
		for i, f := range active {
			names[i] = f.Name
		}
		return currentState{
			Phase: "—",
			Note:  "serial violation: multiple active features: " + strings.Join(names, ", "),
		}
	}

	f := active[0]
	phase, wave := derivePhase(f, projectDir)

	st := currentState{
		Feature:          f.Name,
		Status:           f.Status,
		Phase:            phase,
		LastApprovedGate: lastPhase(f), // reutiliza status.go: el último gate con result "approve*"
		BlockedBy:        blockers(f, indexFeatures(ff)),
	}
	if wave >= 0 {
		st.Wave = &wave // tomamos la dirección de la variable local: válido en Go (escapa al heap)
	}
	return st
}

// derivePhase computa la fase actual = frontera-aprobada + 1. Devuelve la fase y
// la wave a ejecutar (o -1 si la wave no aplica a esta fase).
func derivePhase(f *feature, projectDir string) (phase string, wave int) {
	lastIdx := -1  // índice en phasePipeline de la última fase lógica aprobada
	lastWave := -1 // mayor wave-N aprobada (-1 = ninguna)

	for _, g := range f.Gates {
		if !strings.HasPrefix(g.Result, "approve") {
			continue
		}
		logical := g.Phase
		if strings.HasPrefix(g.Phase, "wave-") {
			logical = "build" // wave-0, wave-1… colapsan a la fase lógica "build"
			if n, err := strconv.Atoi(strings.TrimPrefix(g.Phase, "wave-")); err == nil && n > lastWave {
				lastWave = n
			}
		}
		if i := phaseIndex(logical); i > lastIdx {
			lastIdx = i
		}
	}

	// Sin ningún gate aprobado: estamos en la primera fase del pipeline.
	if lastIdx < 0 {
		return phasePipeline[0], -1
	}

	current := phasePipeline[lastIdx]

	// Camino a build: si lo último aprobado es "plan" (entramos a build, wave 0)
	// o una wave (vamos a la siguiente), la fase actual es "build" mientras
	// queden waves; si ya se aprobaron todas → la fase siguiente a build.
	if current == "plan" || current == "build" {
		next := lastWave + 1                      // la próxima wave a ejecutar (frontera+1)
		total := totalWaves(projectDir, f.Name)   // -1 si no hay tasks.json para acotar
		if total >= 0 && next >= total {
			return phaseAfter("build"), -1 // todas las waves aprobadas → verdict
		}
		return "build", next
	}

	// Resto de las fases: la actual es la SIGUIENTE a la última aprobada.
	return phaseAfter(current), -1
}

// phaseIndex devuelve la posición de una fase lógica en el pipeline, o -1.
func phaseIndex(p string) int {
	for i, x := range phasePipeline {
		if x == p {
			return i
		}
	}
	return -1
}

// phaseAfter devuelve la fase siguiente a p, o "done" si p es la última (o no
// está en el pipeline).
func phaseAfter(p string) string {
	i := phaseIndex(p)
	if i < 0 || i+1 >= len(phasePipeline) {
		return "done"
	}
	return phasePipeline[i+1]
}

// totalWaves devuelve cuántas waves tiene el plan COMPUTADO (plan.json), o -1 si
// no hay plan todavía. Con -1 no sabemos cuándo termina build, así que asumimos
// que sigue (degradación segura: antes de `sf plan compute` no hay waves).
func totalWaves(projectDir, feature string) int {
	pf := readPlanQuiet(projectDir, feature)
	if len(pf.Waves) == 0 {
		return -1
	}
	return len(pf.Waves)
}

// indexFeatures arma el índice nombre→*feature que necesita blockers (status.go).
func indexFeatures(ff featuresFile) map[string]*feature {
	m := make(map[string]*feature, len(ff.Features))
	for i := range ff.Features {
		m[ff.Features[i].Name] = &ff.Features[i]
	}
	return m
}

// ----------------------------------------------------------------------------
// `sf context current` — el slice de la fase ACTUAL, listo para que un hook lo
// inyecte. Cuelga de computeCurrentState: primero "¿dónde estoy?", después "acá
// está lo que necesitás para este paso".
//
// Dos tiers de inyección (decisión #1 del debate):
//   - breadcrumb (barato, cada turno): una línea que recuerda "estás en
//     SpecForge, feature X, fase Y". Combate el "me olvidé de SpecForge".
//   - slice completo (caro, solo en triggers): el contexto del paso. Hoy el
//     slice "rico" existe para build (reusa el de la wave); las fases de spec
//     traen el breadcrumb + qué artefactos hay disponibles (el cuerpo lo pide
//     el skill cuando lo necesita). Es honesto: no inventamos riqueza que no
//     diseñamos todavía.
// ----------------------------------------------------------------------------

// currentContext es el JSON que emite `sf context current`.
type currentContext struct {
	Breadcrumb string       `json:"breadcrumb"`
	Feature    string       `json:"feature,omitempty"`
	Phase      string       `json:"phase,omitempty"`
	Wave       *int         `json:"wave,omitempty"`
	WaveSlice  *waveContext `json:"wave_slice,omitempty"`          // solo en fase build
	Artifacts  []string     `json:"available_artifacts,omitempty"` // archivos .json que existen
	Note       string       `json:"note,omitempty"`
}

// runContextCurrent parsea los flags de `current` y emite el contexto actual.
// Lo llama runContext (context.go) cuando el sub-comando es "current".
func runContextCurrent(args []string) int {
	projectDir := "."
	breadcrumbOnly := false
	for _, a := range args {
		switch {
		case a == "--breadcrumb":
			breadcrumbOnly = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf context current: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	return contextCurrent(projectDir, breadcrumbOnly)
}

// contextCurrent computa el estado, arma el breadcrumb y (salvo --breadcrumb) el
// slice completo.
func contextCurrent(projectDir string, breadcrumbOnly bool) int {
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf context: cannot read feature state under %s (%v)\n", projectDir, err)
		return 4
	}
	st := computeCurrentState(ff, projectDir)

	// Tier barato: solo la línea, en texto plano (no JSON). Es lo que un hook
	// inyectaría cada turno.
	if breadcrumbOnly {
		fmt.Println(buildBreadcrumb(st, projectDir))
		return 0
	}

	// Tier completo: JSON con el slice del paso.
	cc := buildCurrentContext(st, projectDir)
	out, err := json.MarshalIndent(cc, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf context: marshal failed (%v)\n", err)
		return 1
	}
	fmt.Println(string(out))
	return 0
}

// buildCurrentContext arma el slice completo (función separada del print → fácil
// de testear: devuelve el struct sin tocar stdout).
func buildCurrentContext(st currentState, projectDir string) currentContext {
	cc := currentContext{
		Breadcrumb: buildBreadcrumb(st, projectDir),
		Feature:    st.Feature,
		Phase:      st.Phase,
		Wave:       st.Wave,
		Note:       st.Note,
	}
	if st.Feature == "" {
		return cc // sin feature activa: solo el breadcrumb/nota
	}
	cc.Artifacts = availableArtifacts(projectDir, st.Feature)
	// En build adjuntamos el slice de la wave actual (el slice "rico" que ya
	// teníamos en context.go).
	if st.Phase == "build" && st.Wave != nil {
		if ws, ok := loadWaveContext(projectDir, st.Feature, *st.Wave); ok {
			cc.WaveSlice = &ws
		} else {
			cc.Note = appendNote(cc.Note, fmt.Sprintf("wave %d slice unavailable (missing/invalid tasks.json?)", *st.Wave))
		}
	}
	return cc
}

// buildBreadcrumb arma la línea de orientación. Si no hay feature activa, la nota
// de computeCurrentState ya explica por qué.
func buildBreadcrumb(st currentState, projectDir string) string {
	if st.Feature == "" {
		return "SpecForge: " + st.Note
	}
	b := fmt.Sprintf("SpecForge: feature `%s`, fase `%s`", st.Feature, st.Phase)
	if st.Wave != nil {
		b += fmt.Sprintf(" (wave %d)", *st.Wave)
	}
	return b + " · slice completo: `sf context current`"
}

// availableArtifacts lista qué artefactos .json existen para la feature (rutas
// relativas a su carpeta). El breadcrumb orienta; esto dice qué hay para leer.
func availableArtifacts(projectDir, feature string) []string {
	dir := filepath.Join(projectDir, "specforge", "features", feature)
	var found []string
	for _, name := range []string{"requirements.json", "design.json", "tasks.json", "plan.json", "review.json"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err == nil {
			found = append(found, name)
		}
	}
	return found
}

// appendNote concatena notas con "; ", tolerando que la primera esté vacía.
func appendNote(existing, add string) string {
	if existing == "" {
		return add
	}
	return existing + "; " + add
}
