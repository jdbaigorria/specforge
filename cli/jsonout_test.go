package main

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// ----------------------------------------------------------------------------
// Tests de las salidas machine-readable (F1 de CLI-COMO-FUENTE: R6 + el `--json`
// de status que R7 necesita).
//
// El test que importa de verdad no es "el JSON parsea" — es que agregar `--json`
// NO haya movido la salida humana. Un flag nuevo que de paso reformatea la tabla
// rompe a todo el que scrapeaba la anterior, y eso no lo caza ningún test de
// JSON.
// ----------------------------------------------------------------------------

// captureStdout corre fn con os.Stdout redirigido a un pipe y devuelve lo que
// escribió.
//
// Concepto Go: os.Stdout es una VARIABLE (*os.File), no una constante del
// runtime, así que se puede reapuntar. El defer restaura el original pase lo que
// pase — sin eso, un fallo dentro de fn dejaría el stdout roto para el resto de
// los tests del paquete.
//
// El pipe tiene buffer finito: si fn escribiera más que eso, w.Write bloquearía
// para siempre. Por eso la lectura va en una goroutine que drena en paralelo, y
// recién después de cerrar w esperamos el resultado por el canal.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	orig := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	done := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(r)
		done <- string(b)
	}()

	fn()
	_ = w.Close()
	out := <-done
	_ = r.Close()
	return out
}

func TestMetricsJSON(t *testing.T) {
	dir := makeMetricsProject(t)

	t.Run("--json parsea y trae los mismos slices que la tabla", func(t *testing.T) {
		out := captureStdout(t, func() {
			if code := metricsContext(dir, "widget", true); code != 0 {
				t.Errorf("code=%d, want 0", code)
			}
		})
		var got struct {
			Feature string        `json:"feature"`
			Slices  []sliceMetric `json:"slices"`
		}
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		if got.Feature != "widget" {
			t.Errorf("feature=%q, want %q", got.Feature, "widget")
		}
		// Mismos slices que mide la tabla: breadcrumb, current y la wave 0.
		labels := map[string]bool{}
		for _, s := range got.Slices {
			labels[s.Label] = true
			if s.Tokens <= 0 || s.Chars <= 0 {
				t.Errorf("slice %q con tamaño no positivo: tokens=%d chars=%d", s.Label, s.Tokens, s.Chars)
			}
		}
		for _, want := range []string{"breadcrumb", "context current", "for-wave --n=0"} {
			if !labels[want] {
				t.Errorf("falta el slice %q en el JSON (got %v)", want, labels)
			}
		}
	})

	t.Run("sin --json la salida humana no cambia", func(t *testing.T) {
		out := captureStdout(t, func() { metricsContext(dir, "widget", false) })
		if len(out) == 0 || out[0] == '{' {
			t.Fatalf("la salida humana parece JSON:\n%s", out)
		}
		// Los anclajes de la vista de siempre: encabezado, columnas y la nota
		// del heurístico. Si alguno se cae, el render cambió.
		for _, want := range []string{"context output size — widget", "slice", "~tokens", "chars", "not a model tokenizer"} {
			if !contains(out, want) {
				t.Errorf("la salida humana ya no contiene %q:\n%s", want, out)
			}
		}
	})

	t.Run("proyecto sin plan emite slices [] y no null", func(t *testing.T) {
		bare := t.TempDir()
		writeFile(t, filepath.Join(bare, "specforge/features.json"),
			`{"schema_version":"1.0","features":[{"name":"widget","status":"planned","gates":[]}]}`)
		out := captureStdout(t, func() { metricsContext(bare, "widget", true) })
		var probe map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &probe); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		if s := string(probe["slices"]); s == "null" {
			t.Error(`slices == null; el consumidor que itere sin chequear revienta — queremos []`)
		}
	})
}

func TestStatusJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"), `{"schema_version":"1.0","features":[
		{"name":"alpha","status":"done","lane":"standard","gates":[]},
		{"name":"beta","status":"building","lane":"lite","depends_on":["alpha"],"gates":[]},
		{"name":"gamma","status":"planned","lane":"lite","depends_on":["beta"],"gates":[]}]}`)

	t.Run("histograma de statuses — la fila que sf-audit estimaba", func(t *testing.T) {
		out := captureStdout(t, func() {
			if code := runStatus([]string{"--json", dir}); code != 0 {
				t.Errorf("code=%d, want 0", code)
			}
		})
		var rep statusReport
		if err := json.Unmarshal([]byte(out), &rep); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		for status, want := range map[string]int{"done": 1, "building": 1, "planned": 1} {
			if rep.Statuses[status] != want {
				t.Errorf("statuses[%q]=%d, want %d", status, rep.Statuses[status], want)
			}
		}
		if len(rep.Features) != 3 {
			t.Fatalf("features=%d, want 3", len(rep.Features))
		}
		// El orden topológico: alpha antes que beta antes que gamma.
		if rep.CriticalPath[0] != "alpha" || rep.CriticalPath[2] != "gamma" {
			t.Errorf("critical_path=%v, want alpha…gamma", rep.CriticalPath)
		}
		// gamma depende de beta, que no está done → bloqueada.
		for _, f := range rep.Features {
			if f.Feature == "gamma" && len(f.Blocked) != 1 {
				t.Errorf("gamma.blocked=%v, want [beta]", f.Blocked)
			}
			if f.Feature == "alpha" && len(f.Blocked) != 0 {
				t.Errorf("alpha.blocked=%v, want []", f.Blocked)
			}
		}
	})

	t.Run("sin --json la salida humana no cambia", func(t *testing.T) {
		out := captureStdout(t, func() { runStatus([]string{dir}) })
		for _, want := range []string{"SpecForge — project status", "Critical path (dependency order):", "alpha → beta → gamma"} {
			if !contains(out, want) {
				t.Errorf("la salida humana ya no contiene %q:\n%s", want, out)
			}
		}
	})

	t.Run("--json con --artifacts se rechaza en vez de inventar un esquema", func(t *testing.T) {
		if code := runStatus([]string{"--json", "--artifacts", dir}); code != 2 {
			t.Errorf("code=%d, want 2", code)
		}
	})

	t.Run("proyecto vacío emite features [] y no null", func(t *testing.T) {
		out := captureStdout(t, func() { runStatus([]string{"--json", t.TempDir()}) })
		var probe map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &probe); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		if s := string(probe["features"]); s == "null" {
			t.Error("features == null; queremos []")
		}
	})
}
