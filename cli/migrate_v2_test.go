package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// RM-MIG — migración 1.0 → 2.0.
//
// Lo que decide si esta migración es legítima NO es que convierta bien: es que
// convierta SÓLO FORMATO. `migrate` es el único comando autorizado a tocar
// sellos, y esa autorización se sostiene enteramente en que no decide contenido.
// Numerar criterios por índice es determinista; partirlos en Given/When/Then
// requeriría un modelo. El primer test de acá fija esa frontera.
// ----------------------------------------------------------------------------

// v1Project arma un proyecto en schema 1.0 con acceptance legada ([]string) y su
// gate de requirements sellado sobre el hash actual.
func v1Project(t *testing.T, acceptance string) string {
	t.Helper()
	proj := t.TempDir()
	reqPath := filepath.Join(proj, "specforge", "features", "f", "requirements.json")
	mustWrite(t, reqPath,
		`{"schema_version":"1.0","feature":"f","requirements":[
		   {"id":"R5","ears_type":"ubiquitous","behavior":"b","acceptance":`+acceptance+`}]}`)

	sealed, ok := hashArtifact(reqPath)
	if !ok {
		t.Fatal("setup: cannot hash artifact")
	}
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"),
		`{"schema_version":"1.0","features":[{"name":"f","status":"building","gates":[`+
			`{"phase":"requirements","result":"approve","by":"user","at":"2026-01-02T00:00:00Z","hash":"`+sealed+`"}]}]}`)
	return proj
}

// readAcceptance devuelve el `acceptance` crudo del único requisito.
func readAcceptance(t *testing.T, proj string) []any {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(proj, "specforge", "features", "f", "requirements.json"))
	if err != nil {
		t.Fatalf("no se pudo leer: %v", err)
	}
	var doc map[string]any
	if err := json.Unmarshal(data, &doc); err != nil {
		t.Fatalf("json inválido tras migrar: %v", err)
	}
	reqs, _ := doc["requirements"].([]any)
	if len(reqs) != 1 {
		t.Fatalf("requirements = %v", reqs)
	}
	acc, _ := reqs[0].(map[string]any)["acceptance"].([]any)
	return acc
}

func TestMigrateAcceptanceToV2(t *testing.T) {
	t.Run("asigna ids POR ÍNDICE y preserva el texto", func(t *testing.T) {
		proj := v1Project(t, `["primero","segundo"]`)
		if code := runMigrate([]string{proj}); code != 0 {
			t.Fatalf("migrate exit %d", code)
		}

		acc := readAcceptance(t, proj)
		if len(acc) != 2 {
			t.Fatalf("acceptance = %v", acc)
		}
		want := []struct{ id, text string }{{"R5.1", "primero"}, {"R5.2", "segundo"}}
		for i, w := range want {
			c, _ := acc[i].(map[string]any)
			if c["id"] != w.id || c["text"] != w.text {
				t.Errorf("acceptance[%d] = %v, want id=%s text=%s", i, c, w.id, w.text)
			}
		}
	})

	t.Run("NO parte en given/when/then — eso sería contenido", func(t *testing.T) {
		// La frontera entera de la autorización de migrate. Si algún día esto
		// empieza a fallar porque alguien "mejoró" la conversión, el sello dejó
		// de ser confiable.
		proj := v1Project(t, `["dado un carrito vacío, cuando confirma, entonces rechaza"]`)
		if code := runMigrate([]string{proj}); code != 0 {
			t.Fatalf("migrate exit %d", code)
		}
		c, _ := readAcceptance(t, proj)[0].(map[string]any)
		if _, has := c["given"]; has {
			t.Error("migrate no puede inventar given/when/then: es decisión de contenido")
		}
		if c["text"] != "dado un carrito vacío, cuando confirma, entonces rechaza" {
			t.Errorf("el texto original debe quedar intacto en `text`: %v", c)
		}
	})

	t.Run("es determinista: dos corridas dan los mismos ids", func(t *testing.T) {
		proj := v1Project(t, `["a","b","c"]`)
		runMigrate([]string{proj})
		first := readAcceptance(t, proj)
		runMigrate([]string{proj}) // idempotente: ya está en 2.0
		second := readAcceptance(t, proj)

		if len(first) != len(second) {
			t.Fatalf("la segunda corrida cambió la lista: %v vs %v", first, second)
		}
		for i := range first {
			a, _ := first[i].(map[string]any)
			b, _ := second[i].(map[string]any)
			if a["id"] != b["id"] {
				t.Errorf("id[%d] cambió entre corridas: %v → %v", i, a["id"], b["id"])
			}
		}
	})

	t.Run("una lista MIXTA se deja como está", func(t *testing.T) {
		// Alguien ya la tocó a mano. Adivinar qué quiso hacer es justo el tipo de
		// decisión que migrate no puede tomar.
		proj := v1Project(t, `["suelto",{"id":"R5.9","text":"ya tenía id"}]`)
		if code := runMigrate([]string{proj}); code != 0 {
			t.Fatalf("migrate exit %d", code)
		}
		acc := readAcceptance(t, proj)
		if s, isStr := acc[0].(string); !isStr || s != "suelto" {
			t.Errorf("la lista mixta no debe convertirse: %v", acc)
		}
	})

	t.Run("una lista ya v2 no se toca", func(t *testing.T) {
		proj := v1Project(t, `[{"id":"R5.1","text":"ya estaba"}]`)
		if code := runMigrate([]string{proj}); code != 0 {
			t.Fatalf("migrate exit %d", code)
		}
		c, _ := readAcceptance(t, proj)[0].(map[string]any)
		if c["id"] != "R5.1" || c["text"] != "ya estaba" {
			t.Errorf("acceptance v2 alterada: %v", c)
		}
	})
}

func TestMigrateV2Versioning(t *testing.T) {
	t.Run("1.0 sube a 2.0", func(t *testing.T) {
		proj := v1Project(t, `["a"]`)
		if code := runMigrate([]string{proj}); code != 0 {
			t.Fatalf("migrate exit %d", code)
		}
		data, _ := os.ReadFile(filepath.Join(proj, "specforge", "features", "f", "requirements.json"))
		if !strings.Contains(string(data), `"schema_version": "2.0"`) {
			t.Errorf("no se estampó 2.0:\n%s", data)
		}
	})

	t.Run("un proyecto ya en 2.0 no tiene nada que migrar", func(t *testing.T) {
		proj := v1Project(t, `["a"]`)
		runMigrate([]string{proj})
		// La segunda corrida no debe reportar acciones ni fallar.
		if code := runMigrate([]string{proj}); code != 0 {
			t.Errorf("segunda corrida exit %d, want 0", code)
		}
	})

	t.Run("una versión desconocida sigue rechazándose", func(t *testing.T) {
		proj := v1Project(t, `["a"]`)
		mustWrite(t, filepath.Join(proj, "specforge", "features", "f", "design.json"),
			`{"schema_version":"9.9","feature":"f","components":[]}`)
		if code := runMigrate([]string{proj}); code != 5 {
			t.Errorf("exit = %d, want 5 — el binario es viejo para el proyecto", code)
		}
	})
}

func TestMigrateV2Reseals(t *testing.T) {
	t.Run("el sello válido se actualiza al hash nuevo", func(t *testing.T) {
		// Convertir acceptance cambia los bytes → cambia el hash canónico → el
		// sello dejaría de coincidir y se leería como un silent edit. Re-sellar
		// es lo que migrate está autorizado a hacer, y sólo porque migra formato.
		proj := v1Project(t, `["a","b"]`)
		if code := runMigrate([]string{proj}); code != 0 {
			t.Fatalf("migrate exit %d", code)
		}

		ff, err := readFeaturesFile(proj)
		if err != nil {
			t.Fatalf("features.json ilegible: %v", err)
		}
		after, ok := hashArtifact(filepath.Join(proj, "specforge", "features", "f", "requirements.json"))
		if !ok {
			t.Fatal("no se pudo hashear el artefacto migrado")
		}
		var sealed string
		for _, g := range ff.Features[0].Gates {
			if g.Phase == "requirements" {
				sealed = g.Hash
			}
		}
		if sealed != after {
			t.Errorf("el sello no siguió al artefacto: sello=%s hash=%s", sealed, after)
		}
	})

	t.Run("un sello YA roto no se amnistía", func(t *testing.T) {
		// migrate re-sella lo que estaba bien sellado. Un silent edit real —el
		// sello no coincidía ANTES de migrar— tiene que seguir sin coincidir, o
		// migrate se volvería la forma barata de blanquear una edición a mano.
		proj := v1Project(t, `["a"]`)
		reqPath := filepath.Join(proj, "specforge", "features", "f", "requirements.json")
		mustWrite(t, filepath.Join(proj, "specforge", "features.json"),
			`{"schema_version":"1.0","features":[{"name":"f","status":"building","gates":[`+
				`{"phase":"requirements","result":"approve","by":"user","at":"2026-01-02T00:00:00Z","hash":"deadbeef"}]}]}`)

		runMigrate([]string{proj})

		ff, _ := readFeaturesFile(proj)
		after, _ := hashArtifact(reqPath)
		for _, g := range ff.Features[0].Gates {
			if g.Phase == "requirements" && g.Hash == after {
				t.Error("migrate re-selló un sello que ya estaba roto — eso es amnistiar un silent edit")
			}
		}
	})

	t.Run("--dry-run no escribe nada", func(t *testing.T) {
		proj := v1Project(t, `["a"]`)
		before, _ := os.ReadFile(filepath.Join(proj, "specforge", "features", "f", "requirements.json"))
		if code := runMigrate([]string{"--dry-run", proj}); code != 0 {
			t.Fatalf("exit %d", code)
		}
		after, _ := os.ReadFile(filepath.Join(proj, "specforge", "features", "f", "requirements.json"))
		if string(before) != string(after) {
			t.Error("--dry-run modificó el artefacto")
		}
	})
}

func TestMigrateWarnsContractGotStricter(t *testing.T) {
	// La consecuencia que migrate NO puede resolver: darle ids a los criterios
	// sube el requisito al contrato v2 (un test POR CRITERIO), pero el trace
	// sigue anclando a nivel requisito y repartir esos tests es CONTENIDO.
	// El resultado es correcto y contraintuitivo — justo después de migrar la
	// cobertura CAE. Apareció migrando examples/brownfield-tempconv, que pasó de
	// 3/3 a 0/3 sin que nada lo dijera. Un costo silencioso es cómo una
	// migración se revierte a mano.
	proj := v1Project(t, `["uno","dos"]`)

	out := captureStdout(t, func() { runMigrate([]string{proj}) })
	if !strings.Contains(out, "contract just got stricter") {
		t.Errorf("migrate debe avisar que el contrato se endureció:\n%s", out)
	}
	if !strings.Contains(out, "sf coverage --by-priority") {
		t.Errorf("el aviso debe decir CÓMO ver cuáles quedaron descubiertos:\n%s", out)
	}
	if !strings.Contains(out, "1 requirement(s)") {
		t.Errorf("debe contar los requisitos afectados, no sólo avisar:\n%s", out)
	}
}

func TestMigrateReportsArchivedSkip(t *testing.T) {
	// El salteo es correcto (una feature sellada no se re-sella bajo un contrato
	// que no existía cuando se aprobó), pero saltear EN SILENCIO haría creer que
	// el proyecto entero subió a 2.0. La omisión se reporta.
	proj := v1Project(t, `["a"]`)
	mustWrite(t, filepath.Join(proj, "specforge", "archive", "2026-01-01-vieja", "requirements.json"),
		`{"schema_version":"1.0","feature":"vieja","requirements":[
		   {"id":"R1","ears_type":"ubiquitous","behavior":"b","acceptance":["legado"]}]}`)

	out := captureStdout(t, func() { runMigrate([]string{proj}) })
	if !strings.Contains(out, "archived feature(s) left at their sealed schema") {
		t.Errorf("migrate debe declarar la omisión:\n%s", out)
	}

	// Y de verdad no la tocó: sigue en 1.0 con acceptance legada.
	data, _ := os.ReadFile(filepath.Join(proj, "specforge", "archive", "2026-01-01-vieja", "requirements.json"))
	if !strings.Contains(string(data), `"acceptance":["legado"]`) {
		t.Errorf("una feature archivada no debe migrarse:\n%s", data)
	}
}

// TestMigratedRequirementRunsUnderV2Contract — el efecto que justifica la
// migración entera: después de migrar, el requisito deja de correr bajo el
// contrato débil y el gate empieza a exigirle un test POR CRITERIO.
func TestMigratedRequirementRunsUnderV2Contract(t *testing.T) {
	proj := v1Project(t, `["primero","segundo"]`)

	// Antes: sin ids, el contrato v1 se conforma con un test a nivel requisito.
	if crit := acceptanceCriteriaOf(proj, "f"); len(crit) != 0 {
		t.Fatalf("antes de migrar no debería haber criterios con id: %v", crit)
	}

	if code := runMigrate([]string{proj}); code != 0 {
		t.Fatalf("migrate exit %d", code)
	}

	// Después: dos criterios con id, y el gate los exige uno por uno.
	crit := acceptanceCriteriaOf(proj, "f")
	if len(crit["R5"]) != 2 {
		t.Fatalf("criterios de R5 = %v, want 2", crit["R5"])
	}
	reasons := scenarioReasons(proj, "R5", "test", crit["R5"], traceReq{})
	if len(reasons) != 2 {
		t.Fatalf("reasons = %v, want una por criterio", reasons)
	}
	// Y nombra el CRITERIO, no el requisito: "R5 no tiene test" manda a releer
	// los dos para encontrar cuál falta.
	if !strings.Contains(reasons[0], "R5.1") {
		t.Errorf("el rechazo debe nombrar R5.1, no R5: %q", reasons[0])
	}
}
