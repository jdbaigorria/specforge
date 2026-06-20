package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf next` — la BRÚJULA del agente.
//
// `sf state current` responde "¿dónde estoy?" (feature/fase/wave). `sf next` da
// el paso siguiente: el CONTRATO DE RUNTIME que pide recomendaciones-de-ia.md §1.
// En vez de obligar al agente a interpretar reglas, le decimos en seco:
//
//   - qué puede ESCRIBIR ahora       (allowed_writes)
//   - qué NO puede tocar             (blocked_writes)
//   - qué tiene que LEER antes       (required_reads)
//   - cuál es el próximo gate humano (next_gate)
//   - la instrucción de una línea    (instruction)
//
// Decisión de diseño clave: `sf next` NO calcula la fase por su cuenta. Reusa
// computeCurrentState (state.go), que ya deriva todo desde features.json. Acá
// solo TRADUCIMOS esa fase a un contrato de permisos. Una sola fuente de verdad
// para "¿en qué fase estoy?"; este archivo solo agrega "¿y entonces qué hago?".
// ----------------------------------------------------------------------------

// nextContract es el JSON que emite `sf next --json`. Es el contrato que un hook
// o un LLM puede consumir sin ambigüedad.
type nextContract struct {
	Feature       string   `json:"feature,omitempty"`
	Phase         string   `json:"phase"`
	Wave          *int     `json:"wave,omitempty"`           // puntero: nil = no aplica (igual que en state.go)
	AllowedWrites []string `json:"allowed_writes"`
	BlockedWrites []string `json:"blocked_writes"`
	RequiredReads []string `json:"required_reads"`
	NextGate      string   `json:"next_gate,omitempty"`
	Instruction   string   `json:"instruction"`
	Why           []string `json:"why,omitempty"`  // solo se llena con --explain / --json
	Note          string   `json:"note,omitempty"` // p.ej. "no active feature" o la violación serial
}

// runNext es el punto de entrada de `sf next`. Modos:
//
//	sf next             → texto humano corto (la acción siguiente)
//	sf next --explain   → texto + el "por qué" derivado del estado de gates
//	sf next --json      → el contrato completo en JSON (incluye el why)
func runNext(args []string) int {
	projectDir := "."
	asJSON := false
	explain := false
	for _, a := range args {
		switch {
		case a == "--json":
			asJSON = true
		case a == "--explain":
			explain = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf next: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	// Leemos features.json y derivamos el estado igual que state.go. Si no se
	// puede leer, es el mismo error/exit-code que usa `sf state`.
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "features.json"))
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf next: cannot read features.json under %s (%v)\n", projectDir, err)
		return 4
	}
	var ff featuresFile
	if err := json.Unmarshal(data, &ff); err != nil {
		fmt.Fprintf(os.Stderr, "sf next: invalid features.json (%v)\n", err)
		return 2
	}

	st := computeCurrentState(ff, projectDir)
	nc := buildNextContract(st, ff, projectDir)

	// --json: el contrato completo. El why ya viene incluido.
	if asJSON {
		out, err := json.MarshalIndent(nc, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "sf next: marshal failed (%v)\n", err)
			return 1
		}
		fmt.Println(string(out))
		return 0
	}

	// Salida humana: la instrucción siempre; el "why" solo con --explain.
	printNextHuman(nc, explain)
	return 0
}

// buildNextContract es la función PURA (sin I/O, fácil de testear): dado el
// estado ya computado, devuelve el contrato. La fase manda: un switch traduce
// cada fase del pipeline a permisos + instrucción.
func buildNextContract(st currentState, ff featuresFile, projectDir string) nextContract {
	// Sin feature activa: no hay paso de pipeline, pero igual damos una acción
	// útil (arrancar o reanudar) en vez de un contrato vacío.
	if st.Feature == "" {
		return nextContract{
			Phase:       st.Phase, // "—"
			Instruction: "No active feature. Run `sf-propose <name>` to specify one, or resume an existing feature.",
			Note:        st.Note,
		}
	}

	// Bloqueo por dependencias: si la feature espera a otra (depends_on), el paso
	// real es desbloquear esa dependencia, no avanzar esta fase.
	if len(st.BlockedBy) > 0 {
		return nextContract{
			Feature:     st.Feature,
			Phase:       st.Phase,
			Instruction: fmt.Sprintf("Feature `%s` is blocked. Finish its dependencies first: %s.", st.Feature, strings.Join(st.BlockedBy, ", ")),
			Why:         []string{"blocked_by: " + strings.Join(st.BlockedBy, ", ")},
			Note:        st.Note,
		}
	}

	// feat(file) arma la ruta del artefacto de ESTA feature, relativa al repo.
	// La usamos para allowed/blocked/required, así el agente recibe rutas exactas.
	feat := func(file string) string {
		return filepath.ToSlash(filepath.Join("specforge", "features", st.Feature, file))
	}
	const constitution = "specforge/constitution.json"

	nc := nextContract{
		Feature: st.Feature,
		Phase:   st.Phase,
		Wave:    st.Wave,
		Note:    st.Note,
	}

	// El corazón: traducir fase → contrato. Cada rama es "qué toca ahora".
	// allowed_writes = lo único que esta fase produce; blocked_writes = lo aguas
	// abajo que todavía no corresponde tocar; required_reads = lo aguas arriba
	// que hay que leer para no inventar.
	switch st.Phase {
	case "lane":
		nc.NextGate = "lane"
		nc.RequiredReads = []string{constitution}
		nc.AllowedWrites = []string{"specforge/features.json"}
		nc.BlockedWrites = []string{feat("requirements.json"), feat("design.json"), feat("tasks.json"), "src/**"}
		nc.Instruction = "Run `sf-propose` to triage the lane (lite vs standard) and get it approved at the gate."

	case "requirements":
		nc.NextGate = "requirements"
		nc.RequiredReads = []string{constitution}
		nc.AllowedWrites = []string{feat("requirements.json")}
		nc.BlockedWrites = []string{feat("design.json"), feat("tasks.json"), feat("plan.json"), "src/**"}
		nc.Instruction = "Generate requirements.json only. Do not design or implement code."

	case "design":
		nc.NextGate = "design"
		nc.RequiredReads = []string{feat("requirements.json"), constitution}
		nc.AllowedWrites = []string{feat("design.json")}
		nc.BlockedWrites = []string{feat("tasks.json"), feat("plan.json"), "src/**"}
		nc.Instruction = "Generate design.json only. Do not write tasks or implement code."

	case "tasks":
		nc.NextGate = "tasks"
		nc.RequiredReads = []string{feat("requirements.json"), feat("design.json")}
		nc.AllowedWrites = []string{feat("tasks.json")}
		nc.BlockedWrites = []string{feat("plan.json"), "src/**"}
		nc.Instruction = "Generate a FLAT tasks.json with depends_on. Do not group into waves — `sf plan compute` does that."

	case "plan":
		nc.NextGate = "plan"
		nc.RequiredReads = []string{feat("tasks.json")}
		nc.AllowedWrites = []string{feat("plan.json")}
		nc.BlockedWrites = []string{"src/**"}
		nc.Instruction = "Run `sf plan compute`, then refine wave names/complexity. Do not implement code yet."

	case "build":
		// La única fase donde tocar código es lo correcto. El gate es por wave.
		if st.Wave != nil {
			nc.NextGate = fmt.Sprintf("wave-%d", *st.Wave)
			nc.Instruction = fmt.Sprintf("Implement wave %d only. Run `sf context current` for the exact task slice, then build + test those tasks.", *st.Wave)
		} else {
			nc.NextGate = "wave-0"
			nc.Instruction = "Implement the current wave. Run `sf context current` for the task slice."
		}
		nc.RequiredReads = []string{feat("tasks.json"), feat("plan.json")}
		nc.AllowedWrites = []string{"src/**", "tests/**", feat("progress/")}
		nc.BlockedWrites = []string{feat("requirements.json"), feat("design.json")} // no reescribir spec aprobada durante build

	case "verdict":
		nc.NextGate = "verdict"
		nc.RequiredReads = []string{feat("requirements.json"), feat("design.json"), feat("tasks.json"), feat("plan.json")}
		nc.AllowedWrites = []string{feat("review.json"), feat("trace.json")}
		nc.BlockedWrites = []string{"src/**"} // el review audita; no sigue codeando
		nc.Instruction = "Run `sf-check`: produce the traceability matrix and a verdict. Do not add features."

	case "done":
		nc.Instruction = fmt.Sprintf("Feature `%s` is complete/archived. Nothing to do — start a new feature with `sf-propose`.", st.Feature)

	default:
		// Defensa: si apareciera una fase que no mapeamos, lo decimos en vez de
		// emitir un contrato engañosamente vacío.
		nc.Instruction = fmt.Sprintf("Unknown phase %q — run `sf status` to inspect the feature manually.", st.Phase)
		nc.Note = appendNote(nc.Note, "no playbook for this phase")
	}

	nc.Why = buildWhy(st)
	return nc
}

// buildWhy arma el "por qué" del paso, derivado del ledger de gates — nunca de
// memoria conversacional (recomendaciones-de-ia.md §10). Son razones legibles.
func buildWhy(st currentState) []string {
	var why []string
	if st.LastApprovedGate == "" {
		why = append(why, "no gate approved yet — starting at the first pipeline phase")
	} else {
		why = append(why, "last approved gate: "+st.LastApprovedGate)
	}
	if st.Phase != "done" && st.Phase != "—" {
		why = append(why, fmt.Sprintf("phase `%s` is the next step; its gate (`%s`) is still pending", st.Phase, st.Phase))
	}
	return why
}

// printNextHuman imprime la versión legible. Con explain agregamos el "Why".
func printNextHuman(nc nextContract, explain bool) {
	fmt.Printf("Next: %s\n", nc.Instruction)
	if nc.Feature != "" {
		line := fmt.Sprintf("Phase: %s (feature `%s`)", nc.Phase, nc.Feature)
		if nc.Wave != nil {
			line += fmt.Sprintf(" · wave %d", *nc.Wave)
		}
		fmt.Println(line)
	}
	if nc.NextGate != "" {
		fmt.Printf("Next gate: %s\n", nc.NextGate)
	}
	if nc.Note != "" {
		fmt.Printf("Note: %s\n", nc.Note)
	}
	if explain && len(nc.Why) > 0 {
		fmt.Println("\nWhy:")
		for _, w := range nc.Why {
			fmt.Printf("  - %s\n", w)
		}
	}
}
