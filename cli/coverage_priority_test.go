package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// RM-C2b — cobertura de requisitos por prioridad.
//
// La propiedad que sostiene toda la métrica: LO QUE CUENTA COMO CUBIERTO ACÁ ES
// LO MISMO QUE EXIGE EL GATE. Si esta medición fuera más blanda, diría "95%" de
// un proyecto que el veredicto rechaza — y un número que no predice el gate es
// peor que no tener número, porque se planifica con él.
//
// Por eso el test central no es el conteo por bucket sino
// `TestRequirementCoverageMatchesTheGate`: un requisito con dos criterios y un
// solo test NO está cubierto, exactamente como lo rechaza RM-C1.
// ----------------------------------------------------------------------------

// makePriorityCoverageProject arma un proyecto con las features/requisitos que se le
// pidan. `reqs` es el JSON crudo del array de requirements y `trace` el del
// objeto requirements del trace, para poder ejercitar contratos v1 y v2 sin un
// constructor por forma.
func makePriorityCoverageProject(t *testing.T, features string, blocking string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"), features)
	if blocking != "" {
		writeFile(t, filepath.Join(dir, "specforge/constitution.json"),
			`{"identity_md":"x","verification":{"blocking_priorities":`+blocking+`}}`)
	}
	writeFile(t, filepath.Join(dir, "src/app.go"), "package src\n\nfunc Run() {}\n")
	writeFile(t, filepath.Join(dir, "src/app_test.go"), "package src\n\nfunc TestRun() {}\nfunc TestOther() {}\n")
	return dir
}

func addFeatureSpec(t *testing.T, dir, feature, reqs, trace string) {
	t.Helper()
	writeFile(t, filepath.Join(dir, "specforge/features", feature, "requirements.json"),
		`{"feature":"`+feature+`","requirements":`+reqs+`}`)
	writeFile(t, filepath.Join(dir, "specforge/features", feature, "trace.json"),
		`{"feature":"`+feature+`","requirements":`+trace+`}`)
}

func TestRequirementCoverageByPriority(t *testing.T) {
	t.Run("reparte por prioridad y cuenta cubiertos", func(t *testing.T) {
		dir := makePriorityCoverageProject(t,
			`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`, "")
		addFeatureSpec(t, dir, "f",
			`[{"id":"R1","ears_type":"ubiquitous","behavior":"a","priority":"must",
			   "acceptance":[{"id":"R1.1","text":"corre"}]},
			  {"id":"R2","ears_type":"ubiquitous","behavior":"b","priority":"should",
			   "acceptance":[{"id":"R2.1","text":"otra"}]},
			  {"id":"R3","ears_type":"ubiquitous","behavior":"c","priority":"could",
			   "acceptance":[{"id":"R3.1","text":"mas"}]}]`,
			`{"R1":{"code":[],"test":[],"scenarios":{"R1.1":{"test":["src/app_test.go:TestRun"]}},"status":"ok"},
			  "R2":{"code":[],"test":[],"scenarios":{},"status":"no-test"},
			  "R3":{"code":[],"test":[],"scenarios":{},"status":"no-test"}}`)

		rc := computeRequirementCoverage(dir)
		if rc.Total != 3 || rc.Contracted != 1 {
			t.Fatalf("total=%d contracted=%d, want 3/1", rc.Total, rc.Contracted)
		}
		if b := rc.ByPriority["must"]; b.Total != 1 || b.Contracted != 1 {
			t.Errorf("must = %d/%d, want 1/1", b.Contracted, b.Total)
		}
		if b := rc.ByPriority["should"]; b.Total != 1 || b.Contracted != 0 {
			t.Errorf("should = %d/%d, want 0/1", b.Contracted, b.Total)
		}
		if b := rc.ByPriority["could"]; b.Total != 1 || b.Contracted != 0 {
			t.Errorf("could = %d/%d, want 0/1", b.Contracted, b.Total)
		}
	})

	t.Run("prioridad ausente cae en must (fail-closed)", func(t *testing.T) {
		// Misma regla que el gate: omitir el campo NO puede ser la salida barata.
		dir := makePriorityCoverageProject(t,
			`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`, "")
		addFeatureSpec(t, dir, "f",
			`[{"id":"R1","ears_type":"ubiquitous","behavior":"a","acceptance":[{"id":"R1.1","text":"x"}]}]`,
			`{"R1":{"code":[],"test":[],"scenarios":{},"status":"no-test"}}`)

		rc := computeRequirementCoverage(dir)
		if b := rc.ByPriority["must"]; b.Total != 1 {
			t.Errorf("un requisito sin priority debe contar como must, got %+v", rc.ByPriority)
		}
	})

	t.Run("el incumplido se nombra con su feature", func(t *testing.T) {
		dir := makePriorityCoverageProject(t,
			`{"schema_version":"1.0","features":[{"name":"checkout","status":"checking","gates":[]}]}`, "")
		addFeatureSpec(t, dir, "checkout",
			`[{"id":"R3","ears_type":"ubiquitous","behavior":"a","priority":"must",
			   "acceptance":[{"id":"R3.1","text":"x"}]}]`,
			`{"R3":{"code":[],"test":[],"scenarios":{},"status":"no-test"}}`)

		rc := computeRequirementCoverage(dir)
		got := rc.ByPriority["must"].Unmet
		if len(got) != 1 || got[0] != "checkout/R3" {
			t.Errorf("unmet = %v, want [checkout/R3] — sin la feature, R3 es ambiguo entre features", got)
		}
	})

	t.Run("blocking refleja la constitución", func(t *testing.T) {
		// Dos proyectos con el mismo número significan cosas distintas si uno
		// bloquea por `could` y el otro no. El reporte tiene que decirlo.
		dir := makePriorityCoverageProject(t,
			`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`,
			`["must"]`)
		addFeatureSpec(t, dir, "f",
			`[{"id":"R1","ears_type":"ubiquitous","behavior":"a","priority":"must","acceptance":[{"id":"R1.1","text":"x"}]},
			  {"id":"R2","ears_type":"ubiquitous","behavior":"b","priority":"could","acceptance":[{"id":"R2.1","text":"y"}]}]`,
			`{"R1":{"code":[],"test":[],"scenarios":{},"status":"no-test"},
			  "R2":{"code":[],"test":[],"scenarios":{},"status":"no-test"}}`)

		rc := computeRequirementCoverage(dir)
		if !rc.ByPriority["must"].Blocking {
			t.Error("must debería figurar como bloqueante")
		}
		if rc.ByPriority["could"].Blocking {
			t.Error("could no bloquea en este proyecto y el reporte debe decirlo")
		}
	})

	t.Run("una feature retirada sale del denominador", func(t *testing.T) {
		// Mismo motivo que en la métrica de archivos (DL-5 F3): si retirar
		// bajara el número, la salida racional sería no retirar nada.
		dir := makePriorityCoverageProject(t,
			`{"schema_version":"1.0","features":[
			   {"name":"viva","status":"checking","gates":[]},
			   {"name":"muerta","status":"retired","closed_reason":"se removió","retired_at":"2026-01-01T00:00:00Z","gates":[]}]}`, "")
		addFeatureSpec(t, dir, "viva",
			`[{"id":"R1","ears_type":"ubiquitous","behavior":"a","priority":"must","acceptance":[{"id":"R1.1","text":"x"}]}]`,
			`{"R1":{"code":[],"test":[],"scenarios":{},"status":"no-test"}}`)
		addFeatureSpec(t, dir, "muerta",
			`[{"id":"R9","ears_type":"ubiquitous","behavior":"z","priority":"must","acceptance":[{"id":"R9.1","text":"x"}]}]`,
			`{"R9":{"code":[],"test":[],"scenarios":{},"status":"no-test"}}`)

		rc := computeRequirementCoverage(dir)
		if rc.Total != 1 {
			t.Errorf("total = %d, want 1 — la feature retirada no es cobertura faltante", rc.Total)
		}
		for _, id := range rc.ByPriority["must"].Unmet {
			if strings.HasPrefix(id, "muerta/") {
				t.Errorf("una feature retirada no debe aparecer en unmet: %v", rc.ByPriority["must"].Unmet)
			}
		}
	})
}

// TestRequirementCoverageMatchesTheGate — la propiedad central.
func TestRequirementCoverageMatchesTheGate(t *testing.T) {
	t.Run("dos criterios y un solo test NO está cubierto", func(t *testing.T) {
		// Es exactamente lo que RM-C1 arregló en el gate. Si la métrica contara
		// esto como cubierto, volvería a medir el proxy débil que C1 eliminó: un
		// requisito de 5 criterios dándose por cubierto con 1 test.
		dir := makePriorityCoverageProject(t,
			`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`, "")
		addFeatureSpec(t, dir, "f",
			`[{"id":"R1","ears_type":"ubiquitous","behavior":"a","priority":"must",
			   "acceptance":[{"id":"R1.1","text":"uno"},{"id":"R1.2","text":"dos"}]}]`,
			`{"R1":{"code":[],"test":[],"scenarios":{"R1.1":{"test":["src/app_test.go:TestRun"]}},"status":"partial"}}`)

		rc := computeRequirementCoverage(dir)
		if rc.Contracted != 0 {
			t.Errorf("contracted = %d, want 0 — R1.2 no tiene test", rc.Contracted)
		}
	})

	t.Run("los dos criterios cubiertos SÍ cuenta", func(t *testing.T) {
		dir := makePriorityCoverageProject(t,
			`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`, "")
		addFeatureSpec(t, dir, "f",
			`[{"id":"R1","ears_type":"ubiquitous","behavior":"a","priority":"must",
			   "acceptance":[{"id":"R1.1","text":"uno"},{"id":"R1.2","text":"dos"}]}]`,
			`{"R1":{"code":[],"test":[],"scenarios":{
			   "R1.1":{"test":["src/app_test.go:TestRun"]},
			   "R1.2":{"test":["src/app_test.go:TestOther"]}},"status":"ok"}}`)

		rc := computeRequirementCoverage(dir)
		if rc.Contracted != 1 {
			t.Errorf("contracted = %d, want 1", rc.Contracted)
		}
	})

	t.Run("un ancla de código rota descalifica aunque sea could", func(t *testing.T) {
		// R5 gradúa el CONTRATO DE VERIFICACIÓN, no la integridad del trace. Un
		// ancla rota miente sobre dónde vive lo que describe, y eso no se gradúa.
		dir := makePriorityCoverageProject(t,
			`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`, "")
		addFeatureSpec(t, dir, "f",
			`[{"id":"R1","ears_type":"ubiquitous","behavior":"a","priority":"could",
			   "acceptance":[{"id":"R1.1","text":"uno"}]}]`,
			`{"R1":{"code":["src/borrado.go:Fantasma"],"test":[],
			   "scenarios":{"R1.1":{"test":["src/app_test.go:TestRun"]}},"status":"ok"}}`)

		rc := computeRequirementCoverage(dir)
		if rc.Contracted != 0 {
			t.Errorf("contracted = %d, want 0 — el ancla de código no resuelve", rc.Contracted)
		}
	})

	t.Run("una feature ARCHIVADA resuelve su trace", func(t *testing.T) {
		// Regresión. `done` dejó de ser terminal (DL-5), así que una feature
		// terminada sigue contando — pero sus artefactos viven bajo
		// `archive/<fecha>-<nombre>/`. La primera versión buscaba el trace sólo
		// en `features/<nombre>/`, encontraba los requisitos y NO sus anclas: cada
		// feature cerrada se reportaba 0% cubierta. Apareció corriendo el binario
		// contra `examples/slugify`, que daba 0/4 teniendo los cuatro anclados.
		dir := makePriorityCoverageProject(t,
			`{"schema_version":"1.0","features":[{"name":"slugify","status":"done","gates":[]}]}`, "")
		writeFile(t, filepath.Join(dir, "specforge/archive/2026-06-16-slugify/requirements.json"),
			`{"feature":"slugify","requirements":[{"id":"R1","ears_type":"ubiquitous","behavior":"a",
			  "priority":"must","acceptance":[{"id":"R1.1","text":"x"}]}]}`)
		writeFile(t, filepath.Join(dir, "specforge/archive/2026-06-16-slugify/trace.json"),
			`{"feature":"slugify","requirements":{"R1":{"code":[],"test":[],
			  "scenarios":{"R1.1":{"test":["src/app_test.go:TestRun"]}},"status":"ok"}}}`)

		rc := computeRequirementCoverage(dir)
		if rc.Total != 1 || rc.Contracted != 1 {
			t.Errorf("archivada = %d/%d, want 1/1 — el trace vive en archive/, no en features/",
				rc.Contracted, rc.Total)
		}
	})

	t.Run("contrato v1 legado: un test a nivel requisito alcanza", func(t *testing.T) {
		dir := makePriorityCoverageProject(t,
			`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`, "")
		addFeatureSpec(t, dir, "f",
			`[{"id":"R1","ears_type":"ubiquitous","behavior":"a","priority":"must","acceptance":["prosa suelta"]}]`,
			`{"R1":{"code":[],"test":["src/app_test.go:TestRun"],"status":"ok"}}`)

		rc := computeRequirementCoverage(dir)
		if rc.Contracted != 1 {
			t.Errorf("contracted = %d, want 1 — bajo contrato v1 el test a nivel requisito cuenta", rc.Contracted)
		}
	})
}

func TestCoverageByPriorityJSON(t *testing.T) {
	dir := makePriorityCoverageProject(t,
		`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`, "")
	// DOS requisitos sobre UN archivo de producto, a propósito: así los dos
	// `total` del payload no pueden coincidir por casualidad y la aserción de
	// abajo prueba una propiedad en vez de un accidente del fixture.
	addFeatureSpec(t, dir, "f",
		`[{"id":"R1","ears_type":"ubiquitous","behavior":"a","priority":"must","acceptance":[{"id":"R1.1","text":"x"}]},
		  {"id":"R2","ears_type":"ubiquitous","behavior":"b","priority":"must","acceptance":[{"id":"R2.1","text":"y"}]}]`,
		`{"R1":{"code":[],"test":[],"scenarios":{},"status":"no-test"},
		  "R2":{"code":[],"test":[],"scenarios":{},"status":"no-test"}}`)

	t.Run("sin --by-priority la clave no viaja", func(t *testing.T) {
		out := captureStdout(t, func() { runCoverage([]string{"--json", dir}) })
		var payload map[string]any
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatalf("json inválido: %v", err)
		}
		if _, ok := payload["requirement_coverage"]; ok {
			t.Error("requirement_coverage no debería estar sin el flag")
		}
	})

	t.Run("con --by-priority viaja ANIDADA, no mezclada", func(t *testing.T) {
		// El anidamiento no es cosmético: `total` de arriba cuenta archivos y
		// `total` de adentro cuenta requisitos. Aplanarlas invita a sumarlas.
		out := captureStdout(t, func() { runCoverage([]string{"--json", "--by-priority", dir}) })
		var payload map[string]any
		if err := json.Unmarshal([]byte(out), &payload); err != nil {
			t.Fatalf("json inválido: %v", err)
		}
		rc, ok := payload["requirement_coverage"].(map[string]any)
		if !ok {
			t.Fatalf("falta requirement_coverage en %v", payload)
		}
		if _, ok := rc["by_priority"]; !ok {
			t.Errorf("falta by_priority en %v", rc)
		}
		// La unidad: 1 archivo de producto en el denominador de arriba, 2
		// requisitos en el de adentro. Dos claves con el mismo nombre y
		// significados distintos — de ahí el anidamiento.
		if payload["total"] != float64(1) {
			t.Errorf("total de archivos = %v, want 1", payload["total"])
		}
		if rc["total"] != float64(2) {
			t.Errorf("total de requisitos = %v, want 2", rc["total"])
		}
	})
}

func TestCoverageByPriorityUnknownFlagStillRejected(t *testing.T) {
	// El switch de flags es un lugar clásico para romper el caso default al
	// agregar un case.
	if code := runCoverage([]string{"--nope", t.TempDir()}); code != 2 {
		t.Errorf("exit = %d, want 2 para un flag desconocido", code)
	}
}
