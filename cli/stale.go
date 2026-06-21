package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ----------------------------------------------------------------------------
// Modelo de STALE ARTIFACTS — recomendaciones-de-ia.md §7.
//
// Cada artefacto de spec puede estar en uno de estos estados, derivados SOLO de
// archivos en disco + el gate ledger (nada de memoria conversacional):
//
//   missing   el archivo no existe.
//   draft     existe pero ningún gate `approve` lo certifica todavía.
//   approved  existe, aprobado, y su fundación NO se movió.
//   stale     fue aprobado pero algo cambió después (ver abajo las 3 señales).
//   blocked   todavía no aprobado y una fase aguas arriba está missing/stale.
//
// La cadena de dependencia (cada fase confía en la anterior):
//
//   requirements → design → tasks → plan
//
// Un artefacto aprobado se vuelve STALE por tres señales, en orden de prioridad:
//   1) silent edit   : su hash actual ≠ el hash sellado en su gate (`sf gate
//                      approve` sella el hash → si el contenido cambió sin
//                      re-gatear, lo detectamos). Es la señal ROBUSTA.
//   2) upstream reabierto: una fase aguas arriba se RE-aprobó después (su gate
//                      `at` es posterior al gate de esta fase). Caso F23.
//   3) upstream stale: una fase aguas arriba está stale/missing → la staleness
//                      se PROPAGA aguas abajo.
// ----------------------------------------------------------------------------

// artifactChain es la cadena de fases que producen un artefacto hasheable, EN
// ORDEN. lane no tiene archivo (es la decisión de carril); build/verdict cuelgan
// de plan. Procesamos en este orden para que, al evaluar una fase, los estados
// de sus upstream YA estén computados (y podamos propagar staleness).
var artifactChain = []struct {
	phase string
	file  string // ruta relativa a specforge/features/<feature>/
}{
	{"requirements", "requirements.json"},
	{"design", "design.json"},
	{"tasks", "tasks.json"},
	{"plan", filepath.Join("progress", "plan.json")},
}

// artifactFileForPhase mapea una fase gateada a su artefacto de SPEC hasheable
// (ruta relativa a la carpeta de la feature). El bool es false para fases SIN
// artefacto que sellar. Lo usa `sf gate approve` para saber qué hashear.
//
// Sin artefacto (bool=false):
//   - lane   : la decisión de carril vive en el campo "lane", no en un archivo.
//   - wave-N : las waves son CHECKPOINTS DE EJECUCIÓN, no artefactos de spec. Su
//              "contenido" es código (lo gobierna el drift de trace.json/`sf
//              doctor`), y su artefacto de spec es plan.json, ya sellado en el
//              gate `plan`. Así que un wave gate se registra SIN hash.
func artifactFileForPhase(phase string) (string, bool) {
	switch phase {
	case "requirements":
		return "requirements.json", true
	case "design":
		return "design.json", true
	case "tasks":
		return "tasks.json", true
	case "plan":
		return filepath.Join("progress", "plan.json"), true
	case "verdict":
		return "review.json", true
	default:
		// lane, wave-0, wave-1, … → sin artefacto de spec que sellar.
		return "", false
	}
}

// artifactState es una fila del informe `sf status --artifacts`.
type artifactState struct {
	Phase  string `json:"phase"`
	File   string `json:"file"`
	State  string `json:"state"`            // missing|draft|approved|stale|blocked
	Reason string `json:"reason,omitempty"` // por qué (solo cuando aporta)
}

// computeArtifactStates es la función PURA-de-cómputo (lee disco para
// existencia/hash, pero no imprime ni sale): dada una feature, devuelve el estado
// de cada artefacto de la cadena. La probamos aislada en stale_test.go.
func computeArtifactStates(f *feature, projectDir string) []artifactState {
	featDir := filepath.Join(projectDir, "specforge", "features", f.Name)

	// done[phase] = estado ya computado, para consultar upstream sin recomputar.
	done := map[string]artifactState{}
	out := make([]artifactState, 0, len(artifactChain))

	for _, link := range artifactChain {
		st := artifactState{Phase: link.phase, File: link.file}
		absPath := filepath.Join(featDir, link.file)
		exists := fileExists(absPath)
		g := latestApproveGate(f, link.phase)

		switch {
		case !exists:
			st.State = "missing"

		case g == nil:
			// Presente pero sin aprobar. Si una fase aguas arriba no está sólida,
			// este artefacto no puede aprobarse válidamente todavía → blocked.
			if up, ok := firstUnsolidUpstream(link.phase, done); ok {
				st.State = "blocked"
				st.Reason = fmt.Sprintf("%s is %s", up.Phase, up.State)
			} else {
				st.State = "draft"
			}

		default:
			// Aprobado: ¿se movió la fundación? Evaluamos las 3 señales en orden.
			st.State, st.Reason = staleReason(link.phase, absPath, g, f, done)
		}

		done[link.phase] = st
		out = append(out, st)
	}
	return out
}

// staleReason decide si una fase APROBADA sigue "approved" o pasó a "stale", y
// con qué razón. Devuelve (estado, razón).
func staleReason(phase, absPath string, g *gate, f *feature, done map[string]artifactState) (string, string) {
	// Señal 1 — silent edit (la robusta): hash sellado vs hash actual.
	if g.Hash != "" {
		if cur, ok := hashArtifact(absPath); ok && cur != g.Hash {
			return "stale", "modified after approval (hash mismatch)"
		}
	}

	// Señal 2 — upstream reabierto: alguna fase aguas arriba se re-aprobó después.
	for _, up := range upstreamPhases(phase) {
		if ug := latestApproveGate(f, up); ug != nil && ug.At > g.At {
			return "stale", fmt.Sprintf("%s re-approved after this", up)
		}
	}

	// Señal 3 — propagación: alguna fase aguas arriba está missing/stale/blocked.
	if up, ok := firstUnsolidUpstream(phase, done); ok {
		return "stale", fmt.Sprintf("%s is %s", up.Phase, up.State)
	}

	// Sin hash sellado no podemos descartar un silent edit: lo decimos honestamente
	// pero NO marcamos stale (no hay evidencia de cambio).
	if g.Hash == "" {
		return "approved", "approved before hash stamping; edits unverifiable"
	}
	return "approved", ""
}

// firstUnsolidUpstream devuelve el primer artefacto aguas arriba que NO esté
// "approved" (missing/draft/stale/blocked). "sólido" = approved.
func firstUnsolidUpstream(phase string, done map[string]artifactState) (artifactState, bool) {
	for _, up := range upstreamPhases(phase) {
		if s, ok := done[up]; ok && s.State != "approved" {
			return s, true
		}
	}
	return artifactState{}, false
}

// upstreamPhases devuelve las fases de la cadena ANTERIORES a `phase` (las de las
// que depende), en orden.
func upstreamPhases(phase string) []string {
	var ups []string
	for _, link := range artifactChain {
		if link.phase == phase {
			break
		}
		ups = append(ups, link.phase)
	}
	return ups
}

// latestApproveGate devuelve el ÚLTIMO gate `approve*` de una fase (nil si
// ninguno). "último" = el de mayor `at`; recorremos en orden y nos quedamos con
// el más reciente para tolerar re-aprobaciones (varias entradas de la misma fase).
func latestApproveGate(f *feature, phase string) *gate {
	var latest *gate
	for i := range f.Gates {
		g := &f.Gates[i]
		if g.Phase != phase || !approveResult(g.Result) {
			continue
		}
		if latest == nil || g.At > latest.At {
			latest = g
		}
	}
	return latest
}

// approveResult: un resultado cuenta como aprobación si empieza con "approve"
// (mismo criterio que el resto del CLI: derivePhase, lastPhase).
func approveResult(result string) bool {
	return len(result) >= 7 && result[:7] == "approve"
}

// hashArtifact devuelve el sha256 del contenido CANÓNICO del artefacto (hex), y
// si pudo leerlo. Canónico = re-serializamos el JSON con claves ordenadas, así un
// cambio SOLO de formato/espacios no dispara un falso stale; cambia el hash solo
// si cambia el contenido real. Si el archivo no es JSON válido, hasheamos los
// bytes crudos (degradación segura).
func hashArtifact(absPath string) (string, bool) {
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return "", false
	}
	canonical := raw
	var v interface{}
	if json.Unmarshal(raw, &v) == nil {
		if c, err := json.Marshal(v); err == nil {
			canonical = c // json.Marshal ordena las claves de los maps → estable
		}
	}
	sum := sha256.Sum256(canonical)
	return hex.EncodeToString(sum[:]), true
}

// fileExists es un helper chico: ¿existe el path (y no es un error de lectura)?
func fileExists(absPath string) bool {
	_, err := os.Stat(absPath)
	return err == nil
}

// ----------------------------------------------------------------------------
// Render de `sf status --artifacts`.
// ----------------------------------------------------------------------------

// runStatusArtifacts imprime la vista de artefactos. Sin filtro: una tabla por
// feature (saltea las archivadas/done, que ya no se gobiernan). Con --feature:
// solo esa. Devuelve exit 5 si alguna feature mostrada tiene artefactos stale,
// para que CI/hooks puedan reaccionar (mismo patrón que doctor con el drift).
func runStatusArtifacts(ff featuresFile, projectDir, featureFilter string) int {
	matched := false
	anyStale := false
	for i := range ff.Features {
		f := &ff.Features[i]
		if featureFilter != "" && f.Name != featureFilter {
			continue
		}
		if featureFilter == "" && f.Status == "done" {
			continue // archivadas: nada que gobernar
		}
		matched = true
		states := computeArtifactStates(f, projectDir)
		printArtifactStates(f.Name, states)
		for _, s := range states {
			if s.State == "stale" {
				anyStale = true
			}
		}
		fmt.Println()
	}
	if !matched {
		if featureFilter != "" {
			fmt.Printf("No feature %q.\n", featureFilter)
		} else {
			fmt.Println("No active features (all done/archived).")
		}
		return 0
	}
	if anyStale {
		fmt.Println("⚠ stale artifacts detected — re-approve the affected phase(s) with `sf gate approve`.")
		return 5
	}
	return 0
}

var artifactCols = []string{"artifact", "state", "reason"}

// printArtifactStates imprime la tabla de estados de UNA feature, reutilizando el
// renderTable de table.go (mismo look que `sf status`).
func printArtifactStates(feature string, states []artifactState) {
	fmt.Printf("SpecForge — artifacts: %s\n\n", feature)
	cells := make([][]string, len(states))
	for i, s := range states {
		cells[i] = []string{s.Phase, s.State, orDash(s.Reason)}
	}
	renderTable(artifactCols, cells)
}
