package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// ----------------------------------------------------------------------------
// domain: conocimiento de NEGOCIO/DOMINIO, a NIVEL PROYECTO (como constitution,
// no por-feature). Vive en specforge/context/domain.json → domain.md. Las reglas
// de negocio cruzan features ("no se permiten órdenes duplicadas" aplica a toda
// feature que toque órdenes), por eso no usa --feature.
//
// El truco que lo hace barato: reusa el patrón `id` + `applies_to` de los
// principios de la constitution. Así el slicer (`sf context for-judge`) puede
// inyectar las reglas de dominio filtradas por fase con CERO lógica nueva — la
// misma inversión applies_to→fase que ya hace con los principios.
// ----------------------------------------------------------------------------

// Modelo de domain.json. Las tres capas son opcionales (omitempty): un proyecto
// puede tener solo glosario, o solo reglas, etc.
type domainFile struct {
	SchemaVersion string         `json:"schema_version"`
	Glossary      []glossaryTerm `json:"glossary,omitempty"`
	Entities      []entity       `json:"entities,omitempty"`
	Rules         []domainRule   `json:"rules,omitempty"`
}

// glossaryTerm: el lenguaje ubicuo del proyecto. Aliases son sinónimos (ej.
// "Orden" ↔ "Order" ↔ "Pedido") para que el agente reconozca el mismo concepto.
type glossaryTerm struct {
	Term       string   `json:"term"`
	Definition string   `json:"definition"`
	Aliases    []string `json:"aliases,omitempty"`
}

// entity: una entidad de dominio (Order, User, ...). Invariants son verdades de
// negocio sobre la entidad ("una orden confirmada publica OrderPlaced").
type entity struct {
	ID          string   `json:"id"` // E1, E2... (forma E#, como R#/C#)
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Invariants  []string `json:"invariants,omitempty"`
}

// domainRule: una regla de negocio invariante. Entities son refs a E# que la
// regla restringe; AppliesTo son las fases donde el auditor debe chequearla
// (mismo campo que principle.AppliesTo → reusa el slicer y checkAppliesTo).
type domainRule struct {
	ID        string   `json:"id"` // D1, D2... (forma D#)
	Rule      string   `json:"rule"`
	Entities  []string `json:"entities,omitempty"`
	AppliesTo []string `json:"applies_to,omitempty"`
}

var (
	entIDRe     = regexp.MustCompile(`^E\d+$`) // entity:      E1, E2, ...
	domRuleIDRe = regexp.MustCompile(`^D\d+$`) // domain rule: D1, D2, ...
)

//go:embed templates/domain.tmpl.md
var domainTemplate string

var domainTmpl = template.Must(
	template.New("domain").
		Funcs(template.FuncMap{"list": joinOrDash, "orDash": orDash}).
		Parse(domainTemplate),
)

// runDomain despacha `sf domain <render|validate>`. Como constitution: a nivel
// proyecto, sin --feature (solo el project dir + --stdout en render).
func runDomain(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf domain <render|validate> [project_dir]")
		return 2
	}
	action, rest := args[0], args[1:]

	projectDir := "."
	toStdout := false
	for _, a := range rest {
		switch {
		case a == "--stdout":
			toStdout = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf domain: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	switch action {
	case "render":
		return renderDomain(projectDir, toStdout)
	case "validate":
		return validateDomain(projectDir)
	default:
		fmt.Fprintf(os.Stderr, "sf domain: unknown action %q (use render|validate)\n", action)
		return 2
	}
}

// domainPath devuelve la base (sin extensión) del artefacto domain. Vive bajo
// context/ porque es conocimiento de proyecto (junto a project.md/conventions.md).
func domainPath(projectDir string) string {
	return filepath.Join(projectDir, "specforge", "context", "domain")
}

func readDomainFile(projectDir string) (domainFile, int) {
	path := domainPath(projectDir) + ".json"
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf domain: cannot read %s (%v)\n", path, err)
		return domainFile{}, 4
	}
	var df domainFile
	if err := json.Unmarshal(data, &df); err != nil {
		fmt.Fprintf(os.Stderr, "sf domain: invalid JSON (%v)\n", err)
		return domainFile{}, 2
	}
	return df, 0
}

func renderDomain(projectDir string, toStdout bool) int {
	df, code := readDomainFile(projectDir)
	if code != 0 {
		return code
	}
	var buf strings.Builder
	if err := domainTmpl.Execute(&buf, df); err != nil {
		fmt.Fprintf(os.Stderr, "sf domain: render failed (%v)\n", err)
		return 1
	}
	if toStdout {
		fmt.Print(buf.String())
		return 0
	}
	mdPath := domainPath(projectDir) + ".md"
	if err := os.WriteFile(mdPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf domain: cannot write %s (%v)\n", mdPath, err)
		return 1
	}
	fmt.Printf("rendered %s\n", filepath.Join("specforge", "context", "domain.md"))
	return 0
}

func validateDomain(projectDir string) int {
	df, code := readDomainFile(projectDir)
	if code != 0 {
		return code
	}
	var rep report
	checkDomain(df, &rep)

	fmt.Printf("Validated domain: %d term(s), %d entity(ies), %d rule(s).\n",
		len(df.Glossary), len(df.Entities), len(df.Rules))
	for _, w := range rep.warnings {
		fmt.Printf("  warning: %s\n", w)
	}
	for _, e := range rep.errors {
		fmt.Printf("  ERROR:   %s\n", e)
	}
	if len(rep.errors) > 0 {
		fmt.Printf("\nFAIL: %d error(s), %d warning(s).\n", len(rep.errors), len(rep.warnings))
		return 2
	}
	fmt.Printf("\nOK: domain.json valid (%d warning(s)).\n", len(rep.warnings))
	return 0
}

// checkDomain valida la consistencia interna (espejo de checkConstitution): ids
// únicos + con forma E#/D#, refs de entidad que existen, y applies_to válido
// (reusa checkAppliesTo, definido en constitution.go).
func checkDomain(df domainFile, rep *report) {
	// Glosario: término y definición no vacíos.
	for i, g := range df.Glossary {
		if strings.TrimSpace(g.Term) == "" {
			rep.errorf("glossary entry #%d: empty term", i+1)
		}
		if strings.TrimSpace(g.Definition) == "" {
			rep.errorf("glossary term %q: empty definition", g.Term)
		}
	}

	// Entidades: id único + forma E# + name. Juntamos los ids para validar refs.
	entIDs := map[string]bool{}
	for i, e := range df.Entities {
		switch {
		case e.ID == "":
			rep.errorf("entity #%d: empty id", i+1)
		case entIDs[e.ID]:
			rep.errorf("duplicate entity id %s", e.ID)
		default:
			entIDs[e.ID] = true
		}
		if e.ID != "" && !entIDRe.MatchString(e.ID) {
			rep.warnf("%s: id not in E# form", e.ID)
		}
		if strings.TrimSpace(e.Name) == "" {
			rep.errorf("%s: empty name", orDash(e.ID))
		}
	}

	// Reglas: id único + forma D# + rule; refs a entidades que existen; applies_to.
	ruleIDs := map[string]bool{}
	for i, r := range df.Rules {
		switch {
		case r.ID == "":
			rep.errorf("rule #%d: empty id", i+1)
		case ruleIDs[r.ID]:
			rep.errorf("duplicate rule id %s", r.ID)
		default:
			ruleIDs[r.ID] = true
		}
		if r.ID != "" && !domRuleIDRe.MatchString(r.ID) {
			rep.warnf("%s: id not in D# form", r.ID)
		}
		if strings.TrimSpace(r.Rule) == "" {
			rep.errorf("%s: empty rule", orDash(r.ID))
		}
		for _, ref := range r.Entities {
			if !entIDs[ref] {
				rep.errorf("%s: references unknown entity %s", orDash(r.ID), ref)
			}
		}
		checkAppliesTo(orDash(r.ID), r.AppliesTo, rep)
	}
}

// loadDomainQuiet lee domain.json SIN imprimir (para el slicer). Devuelve
// (df, false) si no existe o no parsea — degradación segura: un proyecto sin
// dominio simplemente no inyecta nada (mismo patrón que loadConstitutionQuiet).
func loadDomainQuiet(projectDir string) (domainFile, bool) {
	data, err := os.ReadFile(domainPath(projectDir) + ".json")
	if err != nil {
		return domainFile{}, false
	}
	var df domainFile
	if json.Unmarshal(data, &df) != nil {
		return domainFile{}, false
	}
	return df, true
}

// domainRulesForPhase filtra las reglas de dominio cuyo applies_to incluye la
// fase. Es idéntico a principlesForPhase (judge.go), sobre df.Rules.
func domainRulesForPhase(df domainFile, phase string) []domainRule {
	var out []domainRule
	for _, r := range df.Rules {
		if hasString(r.AppliesTo, phase) {
			out = append(out, r)
		}
	}
	return out
}
