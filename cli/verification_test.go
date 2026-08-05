package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// RM-C4 — `kind` + `verification` + evidencia no-test (R9), y el cierre de R10.
//
// EL AGUJERO: "el sistema responderá en menos de 200 ms" no se prueba con un
// test unitario, así que nunca ancla a uno. El contrato, que sólo sabía pedir
// tests, lo dejaba pasar; y como no anclaba, tampoco contaba en la cobertura. El
// requisito existía, nadie lo verificaba, y la métrica no lo reflejaba: decaía
// EN SILENCIO, que es la peor forma de decaer.
// ----------------------------------------------------------------------------

func TestKindAndVerificationDefaults(t *testing.T) {
	t.Run("ausentes ⇒ functional y test", func(t *testing.T) {
		r := requirement{ID: "R1"}
		if got := kindOf(r); got != "functional" {
			t.Errorf("kindOf = %q, want functional", got)
		}
		// Fail-closed: si el default fuera `manual`, omitir el campo sería la
		// forma más barata de salir del contrato.
		if got := verificationOf(r); got != "test" {
			t.Errorf("verificationOf = %q, want test", got)
		}
	})

	t.Run("valores fuera de los enums no validan", func(t *testing.T) {
		for _, r := range []requirement{
			{ID: "R1", EarsType: "ubiquitous", Behavior: "b", Kind: "nfr"},
			{ID: "R1", EarsType: "ubiquitous", Behavior: "b", Verification: "vibes"},
		} {
			var rep report
			checkRequirements(requirementsFile{Feature: "f", Requirements: []requirement{r}}, &rep)
			if len(rep.errors) == 0 {
				t.Errorf("%+v debería rechazarse", r)
			}
		}
	})

	t.Run("kind y ears_type son ortogonales", func(t *testing.T) {
		// Un NFR se enuncia perfectamente como ubiquitous. No se agrega ningún
		// tipo EARS nuevo: kind dice QUÉ es, ears_type dice cómo se ENUNCIA.
		var rep report
		checkRequirements(requirementsFile{Feature: "f", Requirements: []requirement{
			{ID: "R1", EarsType: "ubiquitous", Behavior: "respond in under 200 ms",
				Kind: "performance", Verification: "benchmark",
				Acceptance: acceptanceList{{ID: "R1.1", Text: "p99 < 200ms"}}},
		}}, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("un NFR ubiquitous es válido, got %v", rep.errors)
		}
	})
}

// makeEvidenceProject arma una feature con UN requisito no-test y el trace que
// el llamador quiera probar.
func makeEvidenceProject(t *testing.T, verification, scenarios string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`)
	writeFile(t, filepath.Join(dir, "src/api.go"), "package src\n\nfunc Handle() {}\n")
	writeFile(t, filepath.Join(dir, "bench/latency.json"), `{"p99_ms":180}`)
	writeFile(t, filepath.Join(dir, "specforge/features/f/requirements.json"),
		`{"feature":"f","requirements":[{"id":"R9","ears_type":"ubiquitous",
		  "behavior":"respond in under 200 ms","kind":"performance","verification":"`+verification+`",
		  "acceptance":[{"id":"R9.1","text":"p99 under 200ms"}]}]}`)
	writeFile(t, filepath.Join(dir, "specforge/features/f/trace.json"),
		`{"feature":"f","requirements":{"R9":{"code":["src/api.go:Handle"],"scenarios":`+scenarios+`,"status":"ok"}}}`)
	return dir
}

// contractReasons deja sólo las razones del contrato (descarta las de proceso).
func contractReasons(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	for _, is := range verdictIssues(dir, "f") {
		if strings.Contains(is.Reason, "check run") || strings.Contains(is.Reason, "test result") {
			continue
		}
		out = append(out, is.Reason)
	}
	return out
}

// TestNonTestVerificationRequiresEvidence — R9.
func TestNonTestVerificationRequiresEvidence(t *testing.T) {
	t.Run("EL AGUJERO: sin evidencia ya no pasa en silencio", func(t *testing.T) {
		dir := makeEvidenceProject(t, "benchmark", `{"R9.1":{}}`)
		rs := contractReasons(t, dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "no evidence is declared") {
			t.Fatalf("want rechazo por evidencia faltante, got %v", rs)
		}
		if !strings.HasPrefix(rs[0], "R9.1:") {
			t.Errorf("el rechazo debe nombrar el criterio: %q", rs[0])
		}
	})

	t.Run("una evidencia completa y coherente habilita", func(t *testing.T) {
		dir := makeEvidenceProject(t, "benchmark",
			`{"R9.1":{"evidence":{"kind":"benchmark","ref":"bench/latency.json","recorded":"2026-07-30"}}}`)
		if rs := contractReasons(t, dir); len(rs) != 0 {
			t.Errorf("want 0 razones, got %v", rs)
		}
	})

	t.Run("el kind de la evidencia tiene que coincidir con lo declarado", func(t *testing.T) {
		// Es el chequeo que evita que `evidence` sea un campo de texto libre que
		// se llena para pasar el gate: si el requisito dice benchmark y la
		// evidencia dice test, alguien anotó lo que tenía a mano.
		dir := makeEvidenceProject(t, "benchmark",
			`{"R9.1":{"evidence":{"kind":"test","ref":"bench/latency.json","recorded":"2026-07-30"}}}`)
		rs := contractReasons(t, dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "does not match the declared verification") {
			t.Errorf("want rechazo por kind incoherente, got %v", rs)
		}
	})

	t.Run("un ref que no existe se rechaza", func(t *testing.T) {
		dir := makeEvidenceProject(t, "audit",
			`{"R9.1":{"evidence":{"kind":"audit","ref":"informes/fantasma.pdf","recorded":"2026-07-30"}}}`)
		rs := contractReasons(t, dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "an assertion") {
			t.Errorf("want rechazo por ref inexistente, got %v", rs)
		}
	})

	t.Run("una URL vale como ref", func(t *testing.T) {
		dir := makeEvidenceProject(t, "audit",
			`{"R9.1":{"evidence":{"kind":"audit","ref":"https://ejemplo.com/informe","recorded":"2026-07-30"}}}`)
		if rs := contractReasons(t, dir); len(rs) != 0 {
			t.Errorf("want 0 razones, got %v", rs)
		}
	})

	t.Run("sin fecha no se puede envejecer la evidencia", func(t *testing.T) {
		dir := makeEvidenceProject(t, "manual",
			`{"R9.1":{"evidence":{"kind":"manual","ref":"bench/latency.json"}}}`)
		rs := contractReasons(t, dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "cannot be aged") {
			t.Errorf("want rechazo por fecha faltante, got %v", rs)
		}
	})

	t.Run("un requisito `test` sigue pidiendo test, no evidencia", func(t *testing.T) {
		dir := makeEvidenceProject(t, "test",
			`{"R9.1":{"evidence":{"kind":"benchmark","ref":"bench/latency.json","recorded":"2026-07-30"}}}`)
		rs := contractReasons(t, dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "names no test") {
			t.Errorf("una evidencia no reemplaza al test cuando se declaró test, got %v", rs)
		}
	})

	t.Run("un legado no-test se manda a ponerle ids a sus criterios", func(t *testing.T) {
		// Sin escenarios no hay dónde poner la evidencia. Pedirle un test sería
		// contradecir lo que el requisito declaró.
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "specforge/features.json"),
			`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`)
		writeFile(t, filepath.Join(dir, "src/api.go"), "package src\n\nfunc Handle() {}\n")
		writeFile(t, filepath.Join(dir, "specforge/features/f/requirements.json"),
			`{"feature":"f","requirements":[{"id":"R9","ears_type":"ubiquitous","behavior":"fast",
			  "verification":"benchmark","acceptance":["p99 under 200ms"]}]}`)
		writeFile(t, filepath.Join(dir, "specforge/features/f/trace.json"),
			`{"feature":"f","requirements":{"R9":{"code":["src/api.go:Handle"],"test":[],"status":"ok"}}}`)
		rs := contractReasons(t, dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "no acceptance ids") {
			t.Errorf("want el mensaje que enseña la salida, got %v", rs)
		}
	})
}

// TestEvidenceTraceValidation: el trace valida la evidencia al guardarse.
func TestEvidenceTraceValidation(t *testing.T) {
	check := func(scenarios string) report {
		var tf traceFile
		mustUnmarshal(t, `{"feature":"f","requirements":{"R9":{
			"code":["src/x.go:F"],"scenarios":`+scenarios+`,"status":"ok"}}}`, &tf)
		var rep report
		checkTrace(tf, &rep)
		return rep
	}

	t.Run("evidencia bien formada valida", func(t *testing.T) {
		rep := check(`{"R9.1":{"evidence":{"kind":"benchmark","ref":"bench/x.json","recorded":"2026-07-30"}}}`)
		if len(rep.errors) != 0 {
			t.Errorf("want 0 errores, got %v", rep.errors)
		}
	})

	t.Run("kind fuera del enum, ref vacío y fecha mal formada", func(t *testing.T) {
		for _, s := range []string{
			`{"R9.1":{"evidence":{"kind":"telepatia","ref":"bench/x.json","recorded":"2026-07-30"}}}`,
			`{"R9.1":{"evidence":{"kind":"audit","ref":"","recorded":"2026-07-30"}}}`,
			`{"R9.1":{"evidence":{"kind":"audit","ref":"x.pdf","recorded":"30/07/2026"}}}`,
		} {
			if rep := check(s); len(rep.errors) == 0 {
				t.Errorf("debería rechazarse: %s", s)
			}
		}
	})

	t.Run("ni test ni evidencia avisa", func(t *testing.T) {
		rep := check(`{"R9.1":{}}`)
		if len(rep.warnings) == 0 {
			t.Error("want warning: el gate lo va a rechazar")
		}
	})
}

// TestRequirementRenderComplete — R10: prioridad, kind, verificación, fuentes y
// escenarios, todos en el markdown.
func TestRequirementRenderComplete(t *testing.T) {
	rf := requirementsFile{Feature: "api", Requirements: []requirement{
		{
			ID: "R9", EarsType: "ubiquitous", Behavior: "respond in under 200 ms",
			Priority: "should", Kind: "performance", Verification: "benchmark",
			Source:     sourceRefs{Refs: []string{"S1"}},
			Acceptance: acceptanceList{{ID: "R9.1", Text: "p99 under 200ms"}},
		},
		// Y un legado, que no debe ganar campos que nadie escribió.
		{ID: "R1", EarsType: "event", Trigger: "the user submits", Behavior: "persist it",
			Acceptance: acceptanceList{{Text: "it is persisted"}}},
	}}
	var buf strings.Builder
	if err := reqTmpl.Execute(&buf, rf); err != nil {
		t.Fatalf("render: %v", err)
	}
	out := buf.String()

	for _, want := range []string{
		"## R9 (ubiquitous · should · performance · verified by benchmark)",
		"**Source:** S1",
		"**R9.1** — p99 under 200ms",
		"## R1 (event)",
		"- it is persisted",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("el render no contiene %q:\n%s", want, out)
		}
	}
	// El legado no inventa nada.
	if strings.Contains(out, "## R1 (event ·") {
		t.Errorf("el requisito legado ganó campos que nadie escribió:\n%s", out)
	}
}
