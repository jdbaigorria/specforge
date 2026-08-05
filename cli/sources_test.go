package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// RM-C3 — `sources.json` y `source` como refs (R2, R6).
//
// El agujero está en el otro extremo del tubo: `trace.json` impide que el agente
// declare "implementé R5" sin código ni test, pero nada impedía que el agente
// INVENTARA R5. El ref es lo que vuelve computable la procedencia.
// ----------------------------------------------------------------------------

// makeSourcesProject arma un proyecto con sources.json y una feature.
func makeSourcesProject(t *testing.T, sources string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "inputs/cliente.md"), "pedido original\n")
	writeFile(t, filepath.Join(dir, "inputs/kickoff.md"), "notas del kickoff\n")
	if sources != "" {
		writeFile(t, filepath.Join(dir, "specforge/sources.json"), sources)
	}
	return dir
}

const twoSources = `{"schema_version":"2.0","sources":[
	{"id":"S1","kind":"email","ref":"inputs/cliente.md","captured":"2026-07-14","note":"pedido original"},
	{"id":"S2","kind":"transcript","ref":"inputs/kickoff.md"}]}`

func TestSourcesValidation(t *testing.T) {
	errorsIn := func(dir, body string) []string {
		t.Helper()
		var sf sourcesFile
		mustUnmarshal(t, body, &sf)
		var rep report
		checkSourcesIn(sf, dir, &rep)
		return rep.errors
	}

	t.Run("un artefacto bien formado valida", func(t *testing.T) {
		dir := makeSourcesProject(t, twoSources)
		if errs := errorsIn(dir, twoSources); len(errs) != 0 {
			t.Errorf("want 0 errores, got %v", errs)
		}
	})

	t.Run("un ref que no existe se rechaza", func(t *testing.T) {
		// Una fuente que apunta a un archivo ausente es indistinguible de una
		// fuente inventada — que es justo lo que este artefacto viene a cazar.
		dir := makeSourcesProject(t, "")
		errs := errorsIn(dir, `{"sources":[{"id":"S1","kind":"email","ref":"inputs/fantasma.md"}]}`)
		if len(errs) != 1 || !strings.Contains(errs[0], "does not exist") {
			t.Errorf("want error de ref inexistente, got %v", errs)
		}
	})

	t.Run("una URL no se chequea contra el disco", func(t *testing.T) {
		dir := makeSourcesProject(t, "")
		errs := errorsIn(dir, `{"sources":[{"id":"S1","kind":"document","ref":"https://ejemplo.com/spec"}]}`)
		if len(errs) != 0 {
			t.Errorf("want 0 errores sobre una URL, got %v", errs)
		}
	})

	t.Run("forma del id, kind y captured", func(t *testing.T) {
		cases := []struct{ name, body string }{
			{"id sin forma S#", `{"sources":[{"id":"X1","kind":"email","ref":"inputs/cliente.md"}]}`},
			{"id duplicado", `{"sources":[
				{"id":"S1","kind":"email","ref":"inputs/cliente.md"},
				{"id":"S1","kind":"email","ref":"inputs/kickoff.md"}]}`},
			{"kind fuera del enum", `{"sources":[{"id":"S1","kind":"telepatia","ref":"inputs/cliente.md"}]}`},
			{"ref vacío", `{"sources":[{"id":"S1","kind":"email","ref":""}]}`},
			{"captured mal formada", `{"sources":[{"id":"S1","kind":"email","ref":"inputs/cliente.md","captured":"14/07/2026"}]}`},
		}
		dir := makeSourcesProject(t, "")
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				if errs := errorsIn(dir, c.body); len(errs) == 0 {
					t.Errorf("debería rechazarse: %s", c.body)
				}
			})
		}
	})

	t.Run("dos ids para el mismo material avisa", func(t *testing.T) {
		// Parte la trazabilidad en dos: la mitad de los requisitos apunta a uno
		// y la otra al otro, y ninguna consulta los ve juntos.
		dir := makeSourcesProject(t, "")
		var sf sourcesFile
		mustUnmarshal(t, `{"sources":[
			{"id":"S1","kind":"email","ref":"inputs/cliente.md"},
			{"id":"S2","kind":"document","ref":"inputs/cliente.md"}]}`, &sf)
		var rep report
		checkSourcesIn(sf, dir, &rep)
		if len(rep.warnings) == 0 {
			t.Error("want warning de ref duplicado")
		}
	})
}

// TestSourceRefsDoubleRead: `source` acepta la forma legada y la v2, y preserva
// la que entró.
func TestSourceRefsDoubleRead(t *testing.T) {
	cases := []struct {
		name      string
		in        string
		wantRefs  []string
		wantProse string
		wantOut   string
	}{
		{"prosa libre se preserva", `"del mail de Ana"`, nil, "del mail de Ana", `"del mail de Ana"`},
		{"un string con forma de ref ES un ref", `"S1"`, []string{"S1"}, "", `["S1"]`},
		{"lista de refs", `["S1","S2"]`, []string{"S1", "S2"}, "", `["S1","S2"]`},
		{"null no rompe", `null`, nil, "", `[]`},
		{"string vacío no inventa procedencia", `""`, nil, "", `[]`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var s sourceRefs
			if err := json.Unmarshal([]byte(c.in), &s); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if strings.Join(s.Refs, ",") != strings.Join(c.wantRefs, ",") {
				t.Errorf("refs=%v, want %v", s.Refs, c.wantRefs)
			}
			if s.Prose != c.wantProse {
				t.Errorf("prose=%q, want %q", s.Prose, c.wantProse)
			}
			out, err := json.Marshal(s)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(out) != c.wantOut {
				t.Errorf("round-trip=%s, want %s — un save no debe convertir prosa en refs por su cuenta", out, c.wantOut)
			}
		})
	}
}

// TestSourceRefsValidation — R2.
func TestSourceRefsValidation(t *testing.T) {
	reqWithSource := func(src sourceRefs) requirementsFile {
		return requirementsFile{Feature: "f", Requirements: []requirement{
			{ID: "R1", EarsType: "ubiquitous", Behavior: "b",
				Acceptance: acceptanceList{{Text: "ok"}}, Source: src},
		}}
	}

	t.Run("un ref declarado pasa", func(t *testing.T) {
		dir := makeSourcesProject(t, twoSources)
		var rep report
		checkRequirementsIn(reqWithSource(sourceRefs{Refs: []string{"S1"}}), dir, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("want 0 errores, got %v", rep.errors)
		}
	})

	t.Run("un ref NO declarado se rechaza nombrándolo", func(t *testing.T) {
		dir := makeSourcesProject(t, twoSources)
		var rep report
		checkRequirementsIn(reqWithSource(sourceRefs{Refs: []string{"S9"}}), dir, &rep)
		if len(rep.errors) != 1 || !strings.Contains(rep.errors[0], "S9") {
			t.Errorf("want error nombrando S9, got %v", rep.errors)
		}
	})

	t.Run("sin sources.json, cualquier ref se rechaza", func(t *testing.T) {
		dir := makeSourcesProject(t, "")
		var rep report
		checkRequirementsIn(reqWithSource(sourceRefs{Refs: []string{"S1"}}), dir, &rep)
		if len(rep.errors) != 1 {
			t.Errorf("want 1 error, got %v", rep.errors)
		}
	})

	t.Run("sin project dir NO se inventa un error de existencia", func(t *testing.T) {
		// Reportar "S9 no existe" cuando ni siquiera se pudo abrir sources.json
		// sería afirmar una ausencia que no se midió.
		var rep report
		checkRequirements(reqWithSource(sourceRefs{Refs: []string{"S9"}}), &rep)
		if len(rep.errors) != 0 {
			t.Errorf("want 0 errores sin project dir, got %v", rep.errors)
		}
	})

	t.Run("un ref mal formado se rechaza siempre", func(t *testing.T) {
		var rep report
		checkRequirements(reqWithSource(sourceRefs{Refs: []string{"fuente-1"}}), &rep)
		if len(rep.errors) != 1 {
			t.Errorf("want 1 error de forma, got %v", rep.errors)
		}
	})

	t.Run("la prosa legada avisa pero no bloquea", func(t *testing.T) {
		dir := makeSourcesProject(t, twoSources)
		var rep report
		checkRequirementsIn(reqWithSource(sourceRefs{Prose: "del mail de Ana"}), dir, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("la prosa legada no debe bloquear, got %v", rep.errors)
		}
		if len(rep.warnings) == 0 {
			t.Error("want warning: esa procedencia no se puede chequear")
		}
	})

	t.Run("require_source convierte 'sin fuente' en error", func(t *testing.T) {
		dir := makeSourcesProject(t, twoSources)
		// Sin el flag: sin fuente es legítimo (una idea propia no tiene fuente
		// de cliente).
		var rep report
		checkRequirementsIn(reqWithSource(sourceRefs{}), dir, &rep)
		if len(rep.errors) != 0 {
			t.Errorf("sin require_source no debe bloquear, got %v", rep.errors)
		}

		writeFile(t, filepath.Join(dir, "specforge/constitution.json"),
			`{"identity_md":"x","verification":{"require_source":true}}`)
		rep = report{}
		checkRequirementsIn(reqWithSource(sourceRefs{}), dir, &rep)
		if len(rep.errors) != 1 || !strings.Contains(rep.errors[0], "invented by the model") {
			t.Errorf("con require_source want 1 error, got %v", rep.errors)
		}
	})
}

// TestSourcesCoverage — R6, los dos chequeos que el ref habilita.
func TestSourcesCoverage(t *testing.T) {
	setup := func(t *testing.T) string {
		t.Helper()
		dir := makeSourcesProject(t, twoSources)
		writeFile(t, filepath.Join(dir, "specforge/features/checkout/requirements.json"),
			`{"feature":"checkout","requirements":[
				{"id":"R1","ears_type":"ubiquitous","behavior":"a","source":["S1"],"acceptance":["x"]},
				{"id":"R2","ears_type":"ubiquitous","behavior":"b","acceptance":["x"]},
				{"id":"R3","ears_type":"ubiquitous","behavior":"c","source":"del mail de Ana","acceptance":["x"]}]}`)
		return dir
	}

	t.Run("requisito sin fuente = inventado por el modelo", func(t *testing.T) {
		rep := buildSourcesCoverage(setup(t))
		if rep.Requirements != 3 || rep.Sources != 2 {
			t.Fatalf("requirements=%d sources=%d, want 3/2", rep.Requirements, rep.Sources)
		}
		if len(rep.Unsourced) != 2 {
			t.Fatalf("unsourced=%+v, want R2 y R3", rep.Unsourced)
		}
		// Y distingue "nadie declaró" de "se declaró en prosa": no es lo mismo.
		byID := map[string]unsourcedRequirement{}
		for _, u := range rep.Unsourced {
			byID[u.Requirement] = u
		}
		if byID["R2"].Legacy != "" {
			t.Errorf("R2 no tenía fuente declarada: legacy=%q", byID["R2"].Legacy)
		}
		if byID["R3"].Legacy != "del mail de Ana" {
			t.Errorf("R3 tenía prosa y se perdió: %+v", byID["R3"])
		}
	})

	t.Run("fuente sin requisito = material leído y no usado", func(t *testing.T) {
		rep := buildSourcesCoverage(setup(t))
		if len(rep.Unused) != 1 || rep.Unused[0].ID != "S2" {
			t.Errorf("unused=%+v, want sólo S2", rep.Unused)
		}
	})

	t.Run("una feature archivada no cuenta doble", func(t *testing.T) {
		// Misma regla de dedup que loadTraces (DL-5 F1): `sf feature archive`
		// copia sin borrar, así que sin dedup el reporte diría el doble.
		dir := setup(t)
		writeFile(t, filepath.Join(dir, "specforge/archive/2026-08-01-checkout/requirements.json"),
			`{"feature":"checkout","requirements":[
				{"id":"R1","ears_type":"ubiquitous","behavior":"a","source":["S1"],"acceptance":["x"]}]}`)
		rep := buildSourcesCoverage(dir)
		if rep.Requirements != 3 {
			t.Errorf("requirements=%d, want 3 — la copia archivada se contó de nuevo", rep.Requirements)
		}
	})

	t.Run("sin sources.json la salida es vacía y exit 0, nunca error", func(t *testing.T) {
		dir := t.TempDir()
		out := captureStdout(t, func() {
			if code := sourcesCoverage(dir, true); code != 0 {
				t.Errorf("exit=%d, want 0", code)
			}
		})
		var rep sourcesCoverageReport
		if err := json.Unmarshal([]byte(out), &rep); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		if rep.Sources != 0 || len(rep.Unused) != 0 {
			t.Errorf("want reporte vacío, got %+v", rep)
		}
	})
}

// TestSourcesHookProtection: sources.json es estado — Write directo denegado.
func TestSourcesHookProtection(t *testing.T) {
	// El hook sale temprano si no hay specforge/ — nunca interfiere fuera de un
	// proyecto SpecForge. Hay que montarlo para ejercitar la protección.
	proj := t.TempDir()
	writeFile(t, filepath.Join(proj, "specforge", "features.json"), `{"features":[]}`)
	dec, reason := decidePreToolUse(proj, filepath.Join(proj, "specforge", "sources.json"))
	if dec != "deny" {
		t.Errorf("decision=%q, want deny", dec)
	}
	if !strings.Contains(reason, "sf save sources") {
		t.Errorf("el deny debe enseñar el camino: %q", reason)
	}
}

// TestSourceRender: la procedencia se ve, y la prosa legada se marca como tal.
func TestSourceRender(t *testing.T) {
	render := func(src sourceRefs) string {
		t.Helper()
		var buf strings.Builder
		rf := requirementsFile{Feature: "f", Requirements: []requirement{
			{ID: "R1", EarsType: "ubiquitous", Behavior: "b", Source: src},
		}}
		if err := reqTmpl.Execute(&buf, rf); err != nil {
			t.Fatalf("render: %v", err)
		}
		return buf.String()
	}

	if out := render(sourceRefs{Refs: []string{"S1", "S2"}}); !strings.Contains(out, "**Source:** S1, S2") {
		t.Errorf("el render no muestra los refs:\n%s", out)
	}
	if out := render(sourceRefs{Prose: "del mail de Ana"}); !strings.Contains(out, "not a ref") {
		t.Errorf("la prosa legada debe marcarse como no verificable:\n%s", out)
	}
	if out := render(sourceRefs{}); strings.Contains(out, "**Source:**") {
		t.Errorf("sin fuente no debe aparecer la sección:\n%s", out)
	}
}
