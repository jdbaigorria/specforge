package main

import (
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ----------------------------------------------------------------------------
// DL-15 — el circuito de la verificación no automática.
//
// Lo que se prueba acá no es que sepamos leer un JSON: es que el circuito CIERRE.
// `RM-C4` dejó una demanda sin oferta —el veredicto exige evidencia y no había
// forma asistida de darla—, así que los tests que importan son los que fijan las
// tres piezas: la checklist deriva del trace (nadie la escribe a mano), `record`
// valida con las MISMAS reglas que el gate (el rechazo llega acá y no tres pasos
// después), y la frescura distingue "venció" de "falta".
// ----------------------------------------------------------------------------

// makeEvidenceCircuitProject arma una feature con un requisito verificado por benchmark
// y otro por test — la mezcla realista, que es la que hace visible el filtro.
func makeEvidenceCircuitProject(t *testing.T, maxAge string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"2.0","features":[{"name":"f","status":"checking","gates":[]}]}`)
	if maxAge != "" {
		writeFile(t, filepath.Join(dir, "specforge/constitution.json"),
			`{"identity_md":"x","verification":{"evidence_max_age_days":`+maxAge+`}}`)
	}
	writeFile(t, filepath.Join(dir, "docs/bench-2026.md"), "p99 = 180ms\n")
	writeFile(t, filepath.Join(dir, "specforge/features/f/requirements.json"),
		`{"schema_version":"2.0","feature":"f","requirements":[
		  {"id":"R1","ears_type":"ubiquitous","behavior":"responde rápido","kind":"performance",
		   "verification":"benchmark","acceptance":[{"id":"R1.1","text":"p99 < 200ms"}]},
		  {"id":"R2","ears_type":"ubiquitous","behavior":"suma","acceptance":[{"id":"R2.1","text":"1+1=2"}]}]}`)
	writeFile(t, filepath.Join(dir, "specforge/features/f/trace.json"),
		`{"schema_version":"2.0","feature":"f","requirements":{
		  "R1":{"code":[],"test":[],"scenarios":{},"status":"pending"},
		  "R2":{"code":[],"test":[],"scenarios":{},"status":"pending"}}}`)
	return dir
}

func TestEvidenceChecklist(t *testing.T) {
	now := time.Date(2026, 8, 6, 0, 0, 0, 0, time.UTC)

	t.Run("sólo lista lo que depende de una persona", func(t *testing.T) {
		// R2 se verifica con test: su verificación ya la exige el contrato y la
		// corre la máquina. Meterlo acá haría una lista de todo, que no se usa.
		dir := makeEvidenceCircuitProject(t, "")
		items := evidenceChecklist(dir, "f", 0, now)
		if len(items) != 1 {
			t.Fatalf("items = %+v, want sólo R1.1", items)
		}
		if items[0].Scenario != "R1.1" || items[0].Method != "benchmark" {
			t.Errorf("item = %+v", items[0])
		}
		if items[0].State != "missing" {
			t.Errorf("state = %q, want missing", items[0].State)
		}
	})

	t.Run("una evidencia registrada pasa a recorded", func(t *testing.T) {
		dir := makeEvidenceCircuitProject(t, "")
		if code := runEvidence([]string{"record", "--feature=f", "--scenario=R1.1",
			"--kind=benchmark", "--ref=docs/bench-2026.md", "--recorded=2026-08-01", dir}); code != 0 {
			t.Fatalf("record exit %d", code)
		}
		items := evidenceChecklist(dir, "f", 0, now)
		if items[0].State != "recorded" || items[0].Ref != "docs/bench-2026.md" {
			t.Errorf("item = %+v", items[0])
		}
	})

	t.Run("stale es un estado DISTINTO de missing", func(t *testing.T) {
		// "La tengo pero venció" pide re-verificar; "nunca la tuve" pide verificar
		// por primera vez. Colapsarlos en uno pierde la acción.
		dir := makeEvidenceCircuitProject(t, "30")
		runEvidence([]string{"record", "--feature=f", "--scenario=R1.1",
			"--kind=benchmark", "--ref=docs/bench-2026.md", "--recorded=2026-01-01", dir})

		items := evidenceChecklist(dir, "f", 30, now)
		if items[0].State != "stale" {
			t.Fatalf("state = %q, want stale", items[0].State)
		}
		if items[0].AgeDays < 200 {
			t.Errorf("age = %d días, want ~217", items[0].AgeDays)
		}
	})

	t.Run("sin límite configurado nada vence", func(t *testing.T) {
		dir := makeEvidenceCircuitProject(t, "")
		runEvidence([]string{"record", "--feature=f", "--scenario=R1.1",
			"--kind=benchmark", "--ref=docs/bench-2026.md", "--recorded=2020-01-01", dir})

		items := evidenceChecklist(dir, "f", evidenceMaxAgeDays(dir), now)
		if items[0].State != "recorded" {
			t.Errorf("state = %q — sin evidence_max_age_days no debe vencer", items[0].State)
		}
	})
}

func TestEvidenceRecordValidatesLikeTheGate(t *testing.T) {
	t.Run("un kind que no coincide se rechaza ACÁ", func(t *testing.T) {
		// El requisito declaró `benchmark`. Aceptar un `manual` haría de
		// `evidence` un campo de texto libre para pasar el gate.
		dir := makeEvidenceCircuitProject(t, "")
		if code := runEvidence([]string{"record", "--feature=f", "--scenario=R1.1",
			"--kind=manual", "--ref=docs/bench-2026.md", dir}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
	})

	t.Run("un ref que no existe se rechaza", func(t *testing.T) {
		dir := makeEvidenceCircuitProject(t, "")
		if code := runEvidence([]string{"record", "--feature=f", "--scenario=R1.1",
			"--kind=benchmark", "--ref=docs/no-existe.md", dir}); code != 2 {
			t.Errorf("exit = %d, want 2 — evidencia que nadie puede abrir es una afirmación", code)
		}
	})

	t.Run("un criterio no declarado se rechaza", func(t *testing.T) {
		// Sin esto se podría anclar evidencia a un id inventado, y una evidencia
		// que no cubre ningún criterio declarado es justo lo que el gate no puede
		// detectar después.
		dir := makeEvidenceCircuitProject(t, "")
		if code := runEvidence([]string{"record", "--feature=f", "--scenario=R1.9",
			"--kind=benchmark", "--ref=docs/bench-2026.md", dir}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
	})

	t.Run("un id que no tiene forma de criterio se rechaza", func(t *testing.T) {
		dir := makeEvidenceCircuitProject(t, "")
		if code := runEvidence([]string{"record", "--feature=f", "--scenario=R1",
			"--kind=benchmark", "--ref=docs/bench-2026.md", dir}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
	})

	t.Run("sin --recorded estampa hoy", func(t *testing.T) {
		dir := makeEvidenceCircuitProject(t, "")
		if code := runEvidence([]string{"record", "--feature=f", "--scenario=R1.1",
			"--kind=benchmark", "--ref=docs/bench-2026.md", dir}); code != 0 {
			t.Fatalf("exit %d", code)
		}
		items := evidenceChecklist(dir, "f", 0, time.Now().UTC())
		if items[0].Recorded != time.Now().UTC().Format("2006-01-02") {
			t.Errorf("recorded = %q, want hoy", items[0].Recorded)
		}
	})
}

func TestStaleEvidenceBlocksTheVerdict(t *testing.T) {
	t.Run("vencida bloquea y el motivo habla de EDAD", func(t *testing.T) {
		// Si el mensaje dijera "falta evidencia", la salida barata sería
		// re-anotar la misma con la fecha de hoy sin volver a verificar nada.
		dir := makeEvidenceCircuitProject(t, "30")
		runEvidence([]string{"record", "--feature=f", "--scenario=R1.1",
			"--kind=benchmark", "--ref=docs/bench-2026.md", "--recorded=2020-01-01", dir})

		reasons := staleEvidenceReasons(dir, "f")
		if len(reasons) != 1 {
			t.Fatalf("reasons = %v", reasons)
		}
		for _, want := range []string{"day(s) old", "re-verify", "do not just re-date"} {
			if !strings.Contains(reasons[0], want) {
				t.Errorf("el motivo debe contener %q: %q", want, reasons[0])
			}
		}
	})

	t.Run("apagado por default: el tiempo no pone en rojo un proyecto quieto", func(t *testing.T) {
		dir := makeEvidenceCircuitProject(t, "")
		runEvidence([]string{"record", "--feature=f", "--scenario=R1.1",
			"--kind=benchmark", "--ref=docs/bench-2026.md", "--recorded=2015-01-01", dir})

		if reasons := staleEvidenceReasons(dir, "f"); len(reasons) != 0 {
			t.Errorf("reasons = %v — sin opt-in nada vence", reasons)
		}
	})
}

func TestEvidenceCommandSurface(t *testing.T) {
	t.Run("checklist sale 1 mientras quede trabajo humano", func(t *testing.T) {
		// Exit code y no sólo texto: así entra en un CI o en un `&&`.
		dir := makeEvidenceCircuitProject(t, "")
		if code := runEvidence([]string{"checklist", "--feature=f", dir}); code != 1 {
			t.Errorf("exit = %d, want 1", code)
		}
	})

	t.Run("checklist sale 0 cuando está todo cubierto", func(t *testing.T) {
		dir := makeEvidenceCircuitProject(t, "")
		runEvidence([]string{"record", "--feature=f", "--scenario=R1.1",
			"--kind=benchmark", "--ref=docs/bench-2026.md", dir})
		if code := runEvidence([]string{"checklist", "--feature=f", dir}); code != 0 {
			t.Errorf("exit = %d, want 0", code)
		}
	})

	t.Run("--feature es obligatorio y el subcomando desconocido falla", func(t *testing.T) {
		if code := runEvidence([]string{"checklist"}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
		if code := runEvidence([]string{"nope"}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
		if code := runEvidence(nil); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
	})
}
