package main

import "testing"

// ---------------------------------------------------------------------------
// DL-2 — "la description no resume el workflow".
//
// Una description que enumera el procedimiento le da al agente la ilusión de
// que ya sabe cómo se hace, y puede seguir el resumen en vez de abrir la skill.
// En sf-check el resumen justamente OMITE la trazabilidad estructurada y el
// sellado del verdict, que son los pasos no salteables.
//
// El discriminador es lingüístico: la TERCERA PERSONA ("Reads", "Verifies",
// "Loads") describe el procedimiento; el IMPERATIVO ("Find root cause",
// "Produce a fix plan") le habla al agente, y eso está bien.
// ---------------------------------------------------------------------------

func TestWorkflowSummaryScore(t *testing.T) {
	cases := []struct {
		name string
		desc string
		warn bool
	}{
		// Las tres reales del repo, verbatim.
		{"sf-amend", "Modify a feature that was already shipped and archived, without creating a parallel spec. " +
			"Loads the archived feature and runs a delta mini-pipeline (propose-delta → gate → build → check) " +
			"that edits the EXISTING requirements, design, tasks, and trace.json in place.", true},
		{"sf-build", "Plan and execute feature implementation from approved specs. Use when the user says " +
			"\"sf-build\", \"build feature\". Reads the approved tasks.md, organizes execution by waves, and " +
			"implements with a human gate after each wave.", true},
		{"sf-check", "Validate feature implementation against specs and archive on approval. Verifies " +
			"traceability (every requirement has implementation + test), runs gap analysis, and archives the " +
			"feature on APPROVE.", true},

		// Sanas: imperativo + disparadores, que es la forma correcta.
		{"sfx-triage", "Investigate a bug systematically. Find root cause. Produce a fix plan with test " +
			"strategy. Use when the user reports a bug, error, or unexpected behavior in any codebase.", false},
		{"sfx-tdd", "Implement code using test-driven development. Red-Green-Refactor, one vertical slice " +
			"at a time. Trigger: \"/tdd\", \"tdd this\", \"use TDD\".", false},
		{"sfx-think", "Debate an idea, explore possibilities, evaluate approaches before committing to a " +
			"plan. Open-ended thinking with structured conclusions.", false},

		// Una flecha sola no alcanza: puede nombrar una transformación, no un workflow.
		{"flecha sola", "Convert a spec from md → json. Use when the user asks to migrate a spec.", false},
		// Dos verbos en tercera persona tampoco: eso es alcance, no procedimiento.
		{"dos verbos", "Reads the constitution and validates it against the schema.", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			score, signals := workflowSummaryScore(c.desc)
			if got := score >= 2; got != c.warn {
				t.Errorf("score=%d signals=%v → warn=%v, want %v", score, signals, got, c.warn)
			}
		})
	}
}
