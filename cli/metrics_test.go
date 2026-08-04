package main

import (
	"path/filepath"
	"testing"
)

// TestEstimateTokens fija el comportamiento del heurístico: corrida alnum = 1
// token, cada puntuación = 1 token, el espacio separa pero no cuenta.
func TestEstimateTokens(t *testing.T) {
	cases := []struct {
		in   string
		want int
	}{
		{"", 0},
		{"hello", 1},
		{"hello world", 2},
		{`{"a":1}`, 7},    // { " a " : 1 } → 7 puntuaciones+corridas (a, 1)
		{"R1, R2", 3},     // R1 | , | R2 → 3 (el espacio no cuenta)
		{"café crème", 2}, // acentos: cada corrida de letras Unicode = 1 token
	}
	for _, c := range cases {
		if got := estimateTokens(c.in); got != c.want {
			t.Errorf("estimateTokens(%q) = %d, want %d", c.in, got, c.want)
		}
	}
}

// makeMetricsProject arma una feature EN BUILD (features.json + tasks.json +
// plan.json con una wave) para ejercitar la enumeración de slices for-wave.
func makeMetricsProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"widget","status":"building","gates":[]}]}`)
	base := filepath.Join(dir, "specforge/features/widget")
	writeFile(t, filepath.Join(base, "tasks.json"),
		`{"feature":"widget","tasks":[{"id":"T1","status":"pending","requirement_refs":["R1"],"depends_on":[]}]}`)
	writeFile(t, filepath.Join(base, "progress/plan.json"),
		`{"feature":"widget","waves":[{"n":0,"name":"Foundation","tasks":["T1"]}]}`)
	return dir
}

func TestCollectContextMetrics(t *testing.T) {
	t.Run("missing features.json -> 4", func(t *testing.T) {
		if _, code := collectContextMetrics(t.TempDir(), "widget"); code != 4 {
			t.Errorf("code=%d, want 4", code)
		}
	})

	t.Run("building feature includes breadcrumb, current and for-wave rows", func(t *testing.T) {
		dir := makeMetricsProject(t)
		rows, code := collectContextMetrics(dir, "widget")
		if code != 0 {
			t.Fatalf("code=%d, want 0", code)
		}
		labels := map[string]sliceMetric{}
		for _, r := range rows {
			labels[r.Label] = r
		}
		for _, want := range []string{"breadcrumb", "context current", "for-wave --n=0"} {
			m, ok := labels[want]
			if !ok {
				t.Errorf("missing row %q (got %v)", want, labels)
				continue
			}
			if m.Tokens <= 0 || m.Chars <= 0 {
				t.Errorf("row %q has non-positive size: tokens=%d chars=%d", want, m.Tokens, m.Chars)
			}
		}
	})
}
