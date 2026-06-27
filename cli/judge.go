package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf context for-judge --phase=X --feature=Y` — el material del AUDITOR DE FASE
// (tier calidad, paso 6 del debate).
//
// El CLI NO juzga: junta lo que el subagente-juez fresco necesita y nada más:
//   - el ARTEFACTO de la fase (lo que se juzga), crudo;
//   - SOLO los principios/invariantes cuyo applies_to incluye esa fase.
// El juez nunca ve la constitución entera ni otras features → subagente barato y
// focalizado. El veredicto (cooperativo) lo produce el modelo; sf solo provee el
// material y, después, persiste el veredicto (sf gate record-verdict, próximo).
// ----------------------------------------------------------------------------

type judgeContext struct {
	Feature     string          `json:"feature"`
	Phase       string          `json:"phase"`
	Principles  []principle     `json:"principles"`
	Invariants  []invariant     `json:"invariants"`
	DomainRules []domainRule    `json:"domain_rules,omitempty"` // reglas de negocio que aplican a esta fase
	Artifact    json.RawMessage `json:"artifact,omitempty"`     // el artefacto crudo, pasa tal cual
	Note        string          `json:"note,omitempty"`
}

// contextForJudgeCmd parsea los flags de `for-judge`.
func contextForJudgeCmd(args []string) int {
	projectDir := "."
	feature := ""
	phase := ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--phase="):
			phase = strings.TrimPrefix(a, "--phase=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf context: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" || phase == "" {
		fmt.Fprintln(os.Stderr, "usage: sf context for-judge --phase=PHASE --feature=NAME [project_dir]")
		return 2
	}
	return contextForJudge(projectDir, feature, phase)
}

func contextForJudge(projectDir, feature, phase string) int {
	jc := judgeContext{Feature: feature, Phase: phase}

	// constitution.json es la fuente de los principios. Si falta, degradamos con
	// una nota (el juez no podrá chequear principios, pero igual ve el artefacto).
	if cf, ok := loadConstitutionQuiet(projectDir); ok {
		jc.Principles, jc.Invariants = principlesForPhase(cf, phase)
	} else {
		jc.Note = appendNote(jc.Note, "no/invalid constitution.json — no principles to check")
	}

	// Conocimiento de dominio: las reglas de negocio cuyo applies_to incluye esta
	// fase. Mismo patrón que los principios; opcional (un proyecto sin domain.json
	// simplemente no suma nada acá).
	if df, ok := loadDomainQuiet(projectDir); ok {
		jc.DomainRules = domainRulesForPhase(df, phase)
	}

	// El artefacto de la fase (lo que se juzga), si la fase tiene uno y existe.
	if path := phaseArtifactPath(projectDir, feature, phase); path != "" {
		if raw, err := os.ReadFile(path); err == nil {
			jc.Artifact = json.RawMessage(raw)
		} else {
			jc.Note = appendNote(jc.Note, fmt.Sprintf("artifact for phase %q not found", phase))
		}
	}

	if len(jc.Principles) == 0 && len(jc.Invariants) == 0 && len(jc.DomainRules) == 0 {
		jc.Note = appendNote(jc.Note, fmt.Sprintf("no rules mapped to phase %q (applies_to)", phase))
	}

	out, err := json.MarshalIndent(jc, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf context: marshal failed (%v)\n", err)
		return 1
	}
	fmt.Println(string(out))
	return 0
}

// principlesForPhase filtra los principios/invariantes cuyo applies_to incluye la
// fase. Es la INVERSIÓN del mapeo: la constitución taggea por principio, acá lo
// consultamos por fase.
func principlesForPhase(cf constitutionFile, phase string) ([]principle, []invariant) {
	var ps []principle
	for _, p := range cf.Principles {
		if hasString(p.AppliesTo, phase) {
			ps = append(ps, p)
		}
	}
	var invs []invariant
	for _, inv := range cf.Invariants {
		if hasString(inv.AppliesTo, phase) {
			invs = append(invs, inv)
		}
	}
	return ps, invs
}

// phaseArtifactPath devuelve el .json del artefacto que se juzga en esa fase, o
// "" si la fase no tiene un artefacto único (build se juzga por wave, no acá).
func phaseArtifactPath(projectDir, feature, phase string) string {
	fdir := filepath.Join(projectDir, "specforge", "features", feature)
	switch phase {
	case "requirements", "design", "tasks":
		return filepath.Join(fdir, phase+".json")
	case "plan":
		return filepath.Join(fdir, "progress", "plan.json")
	default: // build, etc.
		return ""
	}
}

// hasString: ¿el slice contiene s? (membership simple para []string).
func hasString(xs []string, s string) bool {
	for _, x := range xs {
		if x == s {
			return true
		}
	}
	return false
}
