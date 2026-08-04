package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// DL-1 — rúbricas built-in en `sf context for-judge`.
//
// El defecto: el juez de fase sólo recibe los principios que el usuario escribió
// en su constitución. Un proyecto sin ningún principio con applies_to incluyendo
// `requirements` hace que for-judge no devuelva reglas, y sf-check retorna antes
// de auditar — o sea, la constitución flaca produce una auditoría que no audita.
//
// La calidad intrínseca del requisito (¿ambiguo? ¿singular? ¿verificable?) NO
// depende del proyecto: es universal y viene de fábrica. El CLI no la evalúa —
// sólo NOMBRA qué rúbrica aplica; el contenido vive en la skill.
// ---------------------------------------------------------------------------

func TestRubricsForPhase(t *testing.T) {
	pmin := []principle{{ID: "P-min", Statement: "minimal code"}}
	other := []principle{{ID: "P-1", Statement: "algo"}}

	cases := []struct {
		name       string
		phase      string
		principles []principle
		want       []string
	}{
		// El caso que motiva el ítem: sin principios, la rúbrica igual aplica.
		{"requirements sin constitución", "requirements", nil, []string{"requirement-quality"}},
		{"requirements con otro principio", "requirements", other, []string{"requirement-quality"}},
		{"requirements + P-min", "requirements", pmin, []string{"requirement-quality", "minimal-code"}},
		{"build con P-min", "build", pmin, []string{"minimal-code"}},
		{"build sin P-min", "build", other, nil},
		{"design sin P-min", "design", nil, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := rubricsForPhase(c.phase, c.principles)
			if len(got) != len(c.want) {
				t.Fatalf("rubrics=%v, want %v", got, c.want)
			}
			for i := range got {
				if got[i] != c.want[i] {
					t.Errorf("rubrics[%d]=%q, want %q", i, got[i], c.want[i])
				}
			}
		})
	}
}

// TestJudgeContextRubricsWithoutConstitution es el caso end-to-end del defecto:
// un proyecto SIN constitution.json pidiendo la fase requirements igual tiene
// material que auditar, y la nota no debe decir que no hay reglas.
func TestJudgeContextRubricsWithoutConstitution(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features/x/requirements.json"),
		`{"feature":"x","requirements":[{"id":"R1","statement":"el sistema debe andar"}]}`)

	jc := buildJudgeContext(dir, "x", "requirements")

	if len(jc.Principles) != 0 {
		t.Errorf("principles=%d, want 0 (no hay constitución)", len(jc.Principles))
	}
	if len(jc.Rubrics) != 1 || jc.Rubrics[0] != "requirement-quality" {
		t.Fatalf("rubrics=%v, want [requirement-quality]", jc.Rubrics)
	}
	// La nota de "no rules mapped" es la señal que sf-check usa para retornar
	// temprano. Con una rúbrica en juego, SÍ hay algo que auditar.
	if contains := jc.Note != "" && strings.Contains(jc.Note, "no rules mapped"); contains {
		t.Errorf("note=%q — no debería decir 'no rules mapped' con una rúbrica activa", jc.Note)
	}
}
