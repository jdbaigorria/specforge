package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// readRaw lee un archivo como string, para comparar el ANTES y el DESPUÉS de una
// operación que debe rechazar sin escribir nada.
func readRaw(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func mustRemove(t *testing.T, path string) {
	t.Helper()
	if err := os.Remove(path); err != nil {
		t.Fatalf("remove %s: %v", path, err)
	}
}

// ----------------------------------------------------------------------------
// DL-5 F2/F3 — `retired` y `abandoned`.
//
// El enum terminaba en `done`, así que no había forma de expresar "esto se
// removió a propósito" ni "esto nunca shippeó". La consecuencia no era estética:
// una feature retirada reporta drift para siempre (su código no está *por
// diseño*), el ruido crece monótonamente, y en algún momento se deja de correr
// el chequeo. Un chequeo que nadie corre no chequea nada.
// ----------------------------------------------------------------------------

// makeClosableProject arma una feature con su trace y el archivo que ancla.
func makeClosableProject(t *testing.T, status string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"slugify","status":"`+status+`","lane":"lite","gates":[]}]}`)
	writeFile(t, filepath.Join(dir, "src/slug.py"), "def slugify(text):\n    return text\n")
	writeFile(t, filepath.Join(dir, "specforge/features/slugify/trace.json"),
		`{"feature":"slugify","requirements":{"R1":{"code":["src/slug.py:slugify"],"test":["tests/t.py::test_x"],"status":"ok"}}}`)
	return dir
}

// TestTerminalStatusRequiresReason — R2.
func TestTerminalStatusRequiresReason(t *testing.T) {
	for _, to := range []string{"retired", "abandoned"} {
		t.Run(to+" sin --reason no escribe nada", func(t *testing.T) {
			from := "done"
			if to == "abandoned" {
				from = "planned"
			}
			dir := makeClosableProject(t, from)
			before := readRaw(t, filepath.Join(dir, "specforge/features.json"))

			if code := runFeatureSetStatus([]string{"--feature=slugify", "--to=" + to, dir}); code != 2 {
				t.Errorf("exit=%d, want 2", code)
			}
			if after := readRaw(t, filepath.Join(dir, "specforge/features.json")); after != before {
				t.Error("el rechazo tiene que ser previo a cualquier escritura")
			}
		})
	}

	t.Run("con --reason persiste motivo y fecha", func(t *testing.T) {
		dir := makeClosableProject(t, "done")
		if code := runFeatureSetStatus([]string{"--feature=slugify", "--to=retired", "--reason=reemplazado por Y", dir}); code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		ff, err := readFeaturesFile(dir)
		if err != nil {
			t.Fatalf("readFeaturesFile: %v", err)
		}
		f := findFeature(&ff, "slugify")
		if f.Status != "retired" {
			t.Errorf("status=%q, want retired", f.Status)
		}
		if f.ClosedReason != "reemplazado por Y" {
			t.Errorf("closed_reason=%q", f.ClosedReason)
		}
		// La fecha la pone el CLI, y el CAMPO dice cuál de los dos cierres fue.
		if f.RetiredAt == "" {
			t.Error("retired_at vacío — la fecha la estampa el CLI, no el agente")
		}
		if f.AbandonedAt != "" {
			t.Error("abandoned_at seteado en un retiro: los dos campos son excluyentes")
		}
	})

	t.Run("abandoned usa su propio campo de fecha", func(t *testing.T) {
		dir := makeClosableProject(t, "building")
		if code := runFeatureSetStatus([]string{"--feature=slugify", "--to=abandoned", "--reason=el cliente canceló el módulo", dir}); code != 0 {
			t.Fatalf("exit=%d, want 0", code)
		}
		ff, _ := readFeaturesFile(dir)
		f := findFeature(&ff, "slugify")
		if f.AbandonedAt == "" || f.RetiredAt != "" {
			t.Errorf("abandoned_at=%q retired_at=%q; el nombre del campo es el que lleva la semántica", f.AbandonedAt, f.RetiredAt)
		}
	})
}

// TestTerminalTransitions — R3. La diferencia entre los dos estados es si hubo
// código en producción, y la tabla de transiciones es donde eso se hace cumplir.
func TestTerminalTransitions(t *testing.T) {
	cases := []struct {
		from, to string
		want     int
	}{
		{"done", "retired", 0},       // se shippeó y se removió
		{"building", "retired", 5},   // nunca shippeó: no se puede "retirar"
		{"planned", "retired", 5},    //
		{"planned", "abandoned", 0},  // se especificó y no siguió
		{"approved", "abandoned", 0}, //
		{"building", "abandoned", 0}, //
		{"checking", "abandoned", 0}, //
		{"blocked", "abandoned", 0},  //
		{"done", "abandoned", 5},     // ya shippeó: lo suyo es retired
		{"retired", "abandoned", 5},  // terminal
		{"abandoned", "retired", 5},  // terminal
		{"retired", "building", 5},   // terminal: no se reabre
		{"abandoned", "planned", 5},  // terminal
	}
	for _, c := range cases {
		t.Run(c.from+"→"+c.to, func(t *testing.T) {
			dir := makeClosableProject(t, c.from)
			code := runFeatureSetStatus([]string{"--feature=slugify", "--to=" + c.to, "--reason=r", dir})
			if code != c.want {
				t.Errorf("exit=%d, want %d", code, c.want)
			}
		})
	}
}

// TestClosedFeaturesSkipDrift — R4, la mitad del drift.
func TestClosedFeaturesSkipDrift(t *testing.T) {
	// El código que la feature anclaba ya no está: es el caso que hoy reporta
	// DRIFT para siempre.
	setup := func(t *testing.T, status string) string {
		t.Helper()
		// El origen no da lo mismo: sólo `done` puede retirarse (hubo código en
		// producción) y `done` no puede abandonarse (ya shippeó).
		from := "done"
		if status == "abandoned" {
			from = "building"
		}
		dir := makeClosableProject(t, from)
		if status != from {
			if code := runFeatureSetStatus([]string{"--feature=slugify", "--to=" + status, "--reason=se removió el módulo", dir}); code != 0 {
				t.Fatalf("set-status %s→%s exit=%d", from, status, code)
			}
		}
		mustRemove(t, filepath.Join(dir, "src/slug.py"))
		return dir
	}

	t.Run("una feature viva con el ancla borrada sí reporta drift", func(t *testing.T) {
		rep := buildDriftReport(setup(t, "done"), "")
		if len(rep.NotImplemented) == 0 {
			t.Error("want drift: el control positivo del test de abajo")
		}
	})

	t.Run("retirada: no reporta drift, pero la omisión se lista", func(t *testing.T) {
		rep := buildDriftReport(setup(t, "retired"), "")
		if len(rep.NotImplemented) != 0 {
			t.Errorf("not_implemented=%+v, want vacío", rep.NotImplemented)
		}
		if rep.Checked != 0 {
			t.Errorf("checked=%d, want 0", rep.Checked)
		}
		if len(rep.Skipped) != 1 {
			t.Fatalf("skipped=%+v, want 1 — omitir en silencio es indistinguible de no encontrar nada", rep.Skipped)
		}
		s := rep.Skipped[0]
		if s.Status != "retired" || s.Reason != "se removió el módulo" || s.At == "" {
			t.Errorf("skipped[0]=%+v; falta status/motivo/fecha", s)
		}
	})

	t.Run("el exit code baja a 0 al retirar", func(t *testing.T) {
		if code := runDrift(setup(t, "retired"), "", true, false); code != 0 {
			t.Errorf("exit=%d, want 0", code)
		}
	})

	t.Run("skipped viaja en el JSON", func(t *testing.T) {
		// El setup va FUERA del capture: set-status escribe en stdout y su línea
		// se mezclaría con el JSON.
		dir := setup(t, "abandoned")
		out := captureStdout(t, func() { runDrift(dir, "", true, true) })
		var probe map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &probe); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		if _, ok := probe["skipped"]; !ok {
			t.Error("falta la clave skipped en el JSON")
		}
	})
}

// TestClosedFeaturesLeaveCoverageDenominator — R4, la mitad de la cobertura.
//
// El punto no es el número: es que el RATCHET no castigue la limpieza. Si
// retirar una feature bajara el porcentaje, la salida racional sería no retirar
// nada nunca.
func TestClosedFeaturesLeaveCoverageDenominator(t *testing.T) {
	// Proyecto: 2 archivos, 1 anclado por `slugify` → 50%.
	setup := func(t *testing.T) string {
		t.Helper()
		dir := makeClosableProject(t, "done")
		writeFile(t, filepath.Join(dir, "src/other.py"), "def other():\n    pass\n")
		return dir
	}

	t.Run("viva: 1 de 2 anclados", func(t *testing.T) {
		d := computeSpecCoverageDetail(setup(t))
		if d.Anchored != 1 || d.Total != 2 {
			t.Errorf("anchored=%d total=%d, want 1/2", d.Anchored, d.Total)
		}
	})

	t.Run("retirada: su archivo sale del denominador, no del numerador", func(t *testing.T) {
		dir := setup(t)
		if code := runFeatureSetStatus([]string{"--feature=slugify", "--to=retired", "--reason=r", dir}); code != 0 {
			t.Fatalf("set-status exit=%d", code)
		}
		d := computeSpecCoverageDetail(dir)
		if d.Total != 1 || d.Anchored != 0 || d.Retired != 1 {
			t.Errorf("anchored=%d total=%d retired=%d, want 0/1/1", d.Anchored, d.Total, d.Retired)
		}
		// Y el archivo retirado NO aparece como pendiente: no es cobertura faltante.
		for _, u := range d.Unanchored {
			if u == "src/slug.py" {
				t.Error("el archivo de una feature retirada aparece como pendiente de especificar")
			}
		}
	})

	t.Run("una spec viva gana sobre una cerrada", func(t *testing.T) {
		dir := setup(t)
		// Otra feature, viva, ancla el MISMO archivo.
		writeFile(t, filepath.Join(dir, "specforge/features.json"),
			`{"schema_version":"1.0","features":[
				{"name":"slugify","status":"retired","closed_reason":"r","retired_at":"2026-08-04T00:00:00Z","gates":[]},
				{"name":"otra","status":"building","gates":[]}]}`)
		writeFile(t, filepath.Join(dir, "specforge/features/otra/trace.json"),
			`{"feature":"otra","requirements":{"R1":{"code":["src/slug.py:slugify"],"test":["tests/t.py::t"],"status":"ok"}}}`)
		d := computeSpecCoverageDetail(dir)
		if d.Anchored != 1 || d.Retired != 0 {
			t.Errorf("anchored=%d retired=%d, want 1/0 — sigue gobernado por una spec viva", d.Anchored, d.Retired)
		}
	})
}

// TestClosedFeaturesAreArchivedToTheHook — R5.
func TestClosedFeaturesAreArchivedToTheHook(t *testing.T) {
	for _, status := range []string{"retired", "abandoned"} {
		t.Run(status+": Write a requirements.json → deny", func(t *testing.T) {
			proj := t.TempDir()
			mustWrite(t, featureFilePath(proj, "f"), `{"name":"f","status":"`+status+`","gates":[]}`)
			dec, reason := decidePreToolUse(proj, filepath.Join(proj, "specforge", "features", "f", "requirements.json"))
			if dec != "deny" {
				t.Errorf("decision=%q, want deny (una feature cerrada no habilita edición)", dec)
			}
			if reason == "" {
				t.Error("el deny debe explicar el camino")
			}
		})
	}
}
