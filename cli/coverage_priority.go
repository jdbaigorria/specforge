package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ----------------------------------------------------------------------------
// `sf coverage --by-priority` — cobertura de REQUISITOS partida por prioridad
// (RM-C2b, el downstream que RM-C2 habilitó y nadie consumió).
//
// POR QUÉ ES UNA MÉTRICA APARTE Y NO UNA PARTICIÓN DE LA DE ARRIBA.
// `sf coverage` mide ARCHIVOS: qué fracción del código está anclada a algún
// trace. `priority` vive en el REQUISITO. Y un archivo sin anclar no tiene
// ningún requisito detrás — por definición, si lo tuviera estaría anclado. O
// sea: el numerador del ratio de archivos se podría bucketear por prioridad,
// pero la mitad que falta del denominador NO. Partir ese porcentaje por
// prioridad daría 100% en todos los buckets, que es peor que no darlo.
//
// Así que la unidad acá es otra: el REQUISITO. La pregunta que contesta es la
// que RM-C2 planteó y quedó sin responder — "un 78% que falta todo `must` es
// indistinguible de uno que falta todo `could`". Este comando los distingue.
//
// QUÉ CUENTA COMO CUBIERTO. Exactamente lo mismo que exige el gate del
// veredicto, reusando `scenarioReasons` (gate.go): con contrato v2, un test (o
// una evidencia, si `verification != test`) POR CRITERIO; con contrato v1
// legado, un test a nivel requisito. Reusarlo no es ahorro de código, es la
// propiedad que hace que el número valga: una métrica que midiera algo más
// blando que el gate diría "95%" de un proyecto que el gate rechaza.
// ----------------------------------------------------------------------------

// priorityBucket es la cobertura de UNA prioridad.
type priorityBucket struct {
	Total      int `json:"total"`
	Contracted int `json:"contracted"`
	// Unmet: los requisitos de esta prioridad cuyo contrato NO se cumple, con
	// su feature (`checkout/R3`). Mismo criterio que `unanchored` en la métrica
	// de archivos: un porcentaje que baja no le dice a nadie qué hacer, una
	// lista sí. Va ordenada para que se diffee entre corridas.
	Unmet []string `json:"unmet"`
	// Blocking: si un incumplimiento de esta prioridad bloquea el veredicto en
	// ESTE proyecto (`verification.blocking_priorities`). Sin esto, dos números
	// idénticos significan cosas distintas en dos proyectos y no hay forma de
	// saberlo desde el reporte.
	Blocking bool `json:"blocking"`
}

// requirementCoverage es la medición completa por prioridad.
type requirementCoverage struct {
	ByPriority map[string]priorityBucket `json:"by_priority"`
	Total      int                       `json:"total"`
	Contracted int                       `json:"contracted"`
}

// priorityOrder fija el orden de salida. Un map de Go itera aleatorio y estas
// tres claves tienen un orden natural (severidad) que alfabéticamente se
// rompería: `could, must, should` no es cómo nadie lee esta tabla.
var priorityOrder = []string{"must", "should", "could"}

// computeRequirementCoverage recorre las features VIVAS y clasifica cada
// requisito por prioridad, marcando si su contrato de verificación se cumple.
//
// Las features cerradas (retired/abandoned) se saltean por el mismo motivo que
// en la métrica de archivos (DL-5 F3): si los requisitos de una feature
// retirada siguieran contando como no cubiertos, retirar bajaría el número y la
// salida racional sería no retirar nada. Una spec cerrada a propósito no es
// cobertura faltante.
func computeRequirementCoverage(projectDir string) requirementCoverage {
	rc := requirementCoverage{ByPriority: map[string]priorityBucket{}}

	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		return rc // lectura QUIETA: sin estado no hay nada que medir
	}
	blocks := blockingPriorities(projectDir)

	for _, f := range ff.Features {
		if terminalStatuses[f.Status] {
			continue
		}
		reqs := requirementsOf(projectDir, f.Name)
		if len(reqs) == 0 {
			continue // feature sin requisitos todavía: no es cobertura faltante
		}
		criteria := acceptanceCriteriaOf(projectDir, f.Name)
		tf := loadTraceFile(projectDir, f.Name)

		for _, r := range reqs {
			p := priorityOf(r) // ausente ⇒ must (fail-closed, igual que el gate)
			b := rc.ByPriority[p]
			b.Total++
			rc.Total++

			if requirementContracted(projectDir, r, criteria, tf) {
				b.Contracted++
				rc.Contracted++
			} else {
				b.Unmet = append(b.Unmet, f.Name+"/"+r.ID)
			}
			b.Blocking = blocks[p]
			rc.ByPriority[p] = b
		}
	}

	for p, b := range rc.ByPriority {
		sort.Strings(b.Unmet)
		rc.ByPriority[p] = b
	}
	return rc
}

// requirementContracted responde la única pregunta del bucket: ¿este requisito
// cumple su contrato de verificación?
//
// Delega en `scenarioReasons`, que es LO MISMO que corre el gate del veredicto
// (gate.go). Cero razones ⇒ contrato cumplido.
func requirementContracted(projectDir string, r requirement, criteria map[string][]string, tf traceFile) bool {
	info, inTrace := tf.Requirements[r.ID]
	method := verificationOf(r) // ausente ⇒ test (fail-closed, RM-C4)

	// Un ancla de código rota descalifica siempre, sin importar la prioridad —
	// misma regla que el gate (gate.go:580): R5 gradúa el CONTRATO, no la
	// integridad del trace.
	for _, anchor := range info.Code {
		if ok, _ := checkAnchor(projectDir, anchor); !ok {
			return false
		}
	}

	if wanted, v2 := criteria[r.ID]; v2 {
		return len(scenarioReasons(projectDir, r.ID, method, wanted, info)) == 0
	}

	// Contrato v1 (legado, sin ids de criterio): un test a nivel requisito, y
	// que resuelva. Un legado que declara `verification != test` no tiene dónde
	// poner la evidencia, así que no puede estar cubierto hasta que suba a v2 —
	// misma salida que le da el gate.
	if !inTrace || len(info.Test) == 0 || method != "test" {
		return false
	}
	for _, tref := range info.Test {
		if ok, _ := checkAnchor(projectDir, tref); !ok {
			return false
		}
	}
	return true
}

// loadTraceFile lee el trace.json de una feature. Lectura QUIETA: sin trace
// devuelve un traceFile vacío, que es lo correcto — significa que ningún
// requisito está anclado, no que haya un error.
//
// Resuelve por `findArtifact` y NO por la ruta directa, que es el mismo
// resolvedor que usa `requirementsOf`. Tiene que ser el mismo: una feature
// `done` vive bajo `archive/<fecha>-<nombre>/`, así que buscar el trace sólo en
// `features/<nombre>/` encontraría los requisitos pero no sus anclas, y cada
// feature cerrada se reportaría como 0% cubierta. Es el bug que apareció
// corriendo esto contra `examples/slugify`.
func loadTraceFile(projectDir, feature string) traceFile {
	var tf traceFile
	path := findArtifact(filepath.Join(projectDir, "specforge"), feature, "trace.json")
	if path == "" {
		return traceFile{Requirements: map[string]traceReq{}}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return traceFile{Requirements: map[string]traceReq{}}
	}
	if json.Unmarshal(data, &tf) != nil || tf.Requirements == nil {
		tf.Requirements = map[string]traceReq{}
	}
	return tf
}

// printRequirementCoverage imprime la tabla por prioridad.
func printRequirementCoverage(rc requirementCoverage) {
	if rc.Total == 0 {
		fmt.Println("\nrequirement coverage: no requirements in any live feature.")
		return
	}

	// El encabezado dice la UNIDAD a propósito. Arriba de esta tabla hay un
	// porcentaje de archivos; sin el rótulo, dos números distintos en la misma
	// pantalla se leen como si midieran lo mismo.
	fmt.Printf("\nrequirement coverage by priority (unit: requirement, not file)\n\n")

	cols := []string{"priority", "covered", "total", "%", "blocking"}
	var rows [][]string
	for _, p := range priorityOrder {
		b, ok := rc.ByPriority[p]
		if !ok {
			continue
		}
		pct := 100 * float64(b.Contracted) / float64(b.Total)
		blocking := "yes"
		if !b.Blocking {
			blocking = "no"
		}
		rows = append(rows, []string{
			p,
			fmt.Sprintf("%d", b.Contracted),
			fmt.Sprintf("%d", b.Total),
			fmt.Sprintf("%.0f%%", pct),
			blocking,
		})
	}
	renderTable(cols, rows)

	// Los incumplidos se listan por prioridad, empezando por la más severa: es
	// el orden en que alguien los va a arreglar.
	for _, p := range priorityOrder {
		b, ok := rc.ByPriority[p]
		if !ok || len(b.Unmet) == 0 {
			continue
		}
		note := ""
		if !b.Blocking {
			note = " — reported, not blocking in this project"
		}
		fmt.Printf("\n%s without a satisfied verification contract%s:\n", p, note)
		for _, id := range b.Unmet {
			fmt.Printf("  %s\n", id)
		}
	}
}
