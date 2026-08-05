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
// constitution: a diferencia de los otros artefactos, es a NIVEL PROYECTO (no
// por-feature). Vive en specforge/constitution.json → specforge/constitution.md,
// así que NO usa --feature, solo el project dir.
// ----------------------------------------------------------------------------

// Modelo de constitution.json (CLI-SPEC §4.8). JSON-first: los principios pasan
// de strings sueltos a objetos estructurados con id + statement + applies_to. El
// `applies_to` es el mapeo fase→principio (paso 6): dice en qué fases el auditor
// de calidad debe chequear ese principio.
type constitutionFile struct {
	SchemaVersion string       `json:"schema_version"`
	IdentityMD    string       `json:"identity_md"`
	Principles    []principle  `json:"principles"`
	Constraints   []string     `json:"constraints"`
	AntiGoals     []string     `json:"anti_goals"`
	Invariants    []invariant  `json:"invariants"`
	Audit         *auditConfig `json:"audit,omitempty"` // config del auditor de fase (paso 6)
	Build         *buildConfig `json:"build,omitempty"` // config de ejecución del build
	Flow          *flowConfig  `json:"flow,omitempty"`  // serial (default) | parallel (F2)
	// Coverage declara qué código NO corresponde especificar (DL-4, cat. 4).
	Coverage *coverageConfig `json:"coverage,omitempty"`
	// Verification gradúa la severidad del gate del veredicto (RM-C2).
	Verification *verificationConfig `json:"verification,omitempty"`
}

// verificationConfig: hasta dónde bloquea el contrato de verificación.
//
// EL DEFAULT NO AFLOJA NADA. Ausente equivale a `["must","should","could"]`, o
// sea: todo bloquea, exactamente como antes de que este campo existiera. Bajar
// el listón es una decisión EXPLÍCITA del proyecto, escrita en la constitución
// —y por lo tanto sellada por gate y auditable— y no una configuración invisible
// que alguien tocó un martes.
type verificationConfig struct {
	BlockingPriorities []string `json:"blocking_priorities,omitempty"`
	// RequireSource (RM-C3) exige que todo requisito cite de dónde salió.
	// Default `false` a propósito: una idea propia no tiene fuente de cliente y
	// eso es legítimo. En greenfield esto sería burocracia; en el camino
	// consultora es el chequeo que impide que el modelo invente requisitos.
	RequireSource bool `json:"require_source,omitempty"`
	// RequireRedWitness (RM-C5) exige que cada test haya demostrado poder
	// fallar. Default `false`: encenderlo el día uno rompería todo proyecto
	// existente, porque sus testigos nunca se acumularon. Se enciende cuando el
	// registro ya tiene historia.
	RequireRedWitness bool `json:"require_red_witness,omitempty"`
}

// verificationOpts: lectura QUIETA de la config de verificación. Sin
// constitución (o inválida), todos los opt-in quedan apagados — que es el
// default correcto: nadie pidió esas garantías.
func verificationOpts(projectDir string) verificationConfig {
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "constitution.json"))
	if err != nil {
		return verificationConfig{}
	}
	var c constitutionFile
	if json.Unmarshal(data, &c) != nil || c.Verification == nil {
		return verificationConfig{}
	}
	return *c.Verification
}

func requireSourceEnabled(projectDir string) bool {
	return verificationOpts(projectDir).RequireSource
}

func requireRedWitnessEnabled(projectDir string) bool {
	return verificationOpts(projectDir).RequireRedWitness
}

// blockingPriorities devuelve el set de prioridades que BLOQUEAN el veredicto.
// Lectura QUIETA: sin constitución, sin campo o con lista vacía → todo bloquea.
//
// Que la lista vacía signifique "todo bloquea" y no "nada bloquea" es a
// propósito: es el mismo criterio fail-closed que `priorityOf`. Un
// `"blocking_priorities": []` escrito por error desactivaría el gate entero, y
// ese es justo el error que no queremos que sea silencioso.
func blockingPriorities(projectDir string) map[string]bool {
	all := map[string]bool{"must": true, "should": true, "could": true}

	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "constitution.json"))
	if err != nil {
		return all
	}
	var c constitutionFile
	if json.Unmarshal(data, &c) != nil || c.Verification == nil || len(c.Verification.BlockingPriorities) == 0 {
		return all
	}
	out := map[string]bool{}
	for _, p := range c.Verification.BlockingPriorities {
		out[strings.TrimSpace(p)] = true
	}
	return out
}

// coverageConfig: la mitad "excluir" del flujo de disposición de la categoría 4
// (fuera de spec). Cada archivo de código no anclado tiene dos salidas —
// adoptarlo (escribirle el requisito que le falta) o excluirlo. Excluirlo vive
// ACÁ, en un artefacto sellado por gate, y no en la cabeza de quien decidió:
// una exclusión sin registro es indistinguible de un olvido.
type coverageConfig struct {
	Exclude []coverageExclusion `json:"exclude,omitempty"`
}

// coverageExclusion saca un archivo (o un directorio, con "/" al final) del
// denominador de `sf coverage`. El motivo es OBLIGATORIO: excluir sube el
// porcentaje, así que sin motivo la métrica se puede maquillar sin dejar rastro.
type coverageExclusion struct {
	Path   string `json:"path"`
	Reason string `json:"reason"`
}

// flowConfig (F2): el flujo serial (una feature activa a la vez, F22) es el
// default y la recomendación para solo-dev. `parallel` lo relaja — habilitado
// por A1 (estado POR feature: dos features activas ya no compiten por un
// archivo global). Pensado para equipos: cada rama trabaja su feature y los
// merges no colisionan. Opt-in explícito en la constitución: es una decisión
// de proyecto, no un default silencioso.
type flowConfig struct {
	Mode string `json:"mode"` // serial | parallel
}

var flowModes = map[string]bool{"serial": true, "parallel": true}

// parallelFlow responde si el proyecto declaró flujo paralelo. Lectura QUIETA
// (sin ruido en stderr): la consultan el hook y los guards en caminos donde
// una constitución ausente simplemente significa "default serial".
func parallelFlow(projectDir string) bool {
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "constitution.json"))
	if err != nil {
		return false
	}
	var cf constitutionFile
	if json.Unmarshal(data, &cf) != nil {
		return false
	}
	return cf.Flow != nil && cf.Flow.Mode == "parallel"
}

// auditConfig gobierna el tier calidad (opt-in). Puntero → si el proyecto no lo
// declara, el campo no aparece (y sf save no lo inventa al re-serializar).
type auditConfig struct {
	Phase string `json:"phase"` // off | nudge | block (default: off)
}

// buildConfig elige cómo ejecutar el build: inline (en el main context, lo de
// siempre), single (un subagente fresco para todo el build) o per-wave (un
// subagente fresco por wave, con checkpoint automático entre waves).
type buildConfig struct {
	Mode string `json:"mode"` // inline | single | per-wave (default: inline)
	// TestCmd es el comando determinista que `sf check run` shellea para correr la
	// suite (ej. "pytest -q", "go test ./...", "npm test"). Sin esto, no hay forma
	// de que el CLI capture un exit code real → el verdict (Capa 2) no puede exigir
	// "verde y fresco". Opcional: si falta, `sf check run` falla pidiéndolo.
	TestCmd string `json:"test_cmd,omitempty"`
	// AgentCmd (F1/R7) es el comando con el que `sf run` invoca al agente de
	// wave (ej. "claude -p --permission-mode acceptEdits"). El seed de la wave
	// entra por su stdin. Opcional: sin él, `sf run` no está disponible y el
	// build es conversacional (sf-build clásico).
	AgentCmd string `json:"agent_cmd,omitempty"`
	// Report (A2/R4) declara el formato del reporte POR TEST de la corrida:
	//   "go-json" → test_cmd emite `go test -json` por stdout
	//   "junit"   → test_cmd contiene {report} (ej. "pytest -q --junitxml={report}")
	// Con esto, el verdict puede exigir causalidad test→requirement (cada test
	// nombrado en el trace corrió y pasó). Opcional: sin él, el verdict solo
	// garantiza "suite verde y fresca".
	Report string `json:"report,omitempty"`
}

var auditPhaseModes = map[string]bool{"off": true, "nudge": true, "block": true}
var buildModes = map[string]bool{"inline": true, "single": true, "per-wave": true}

type principle struct {
	ID        string   `json:"id"`
	Statement string   `json:"statement"`
	AppliesTo []string `json:"applies_to,omitempty"` // fases donde el auditor lo chequea
}

type invariant struct {
	ID           string   `json:"id"`
	Rule         string   `json:"rule"`
	AppliesTo    []string `json:"applies_to,omitempty"`
	PromotedFrom []string `json:"promoted_from"`
	At           string   `json:"at"`
}

// auditablePhases: las fases que un auditor de calidad puede chequear. applies_to
// solo puede apuntar acá (lane/verdict son meta, no se auditan por principio).
var auditablePhases = map[string]bool{
	"requirements": true, "design": true, "tasks": true, "plan": true, "build": true,
}

//go:embed templates/constitution.tmpl.md
var constitutionTemplate string

var constitutionTmpl = template.Must(
	template.New("constitution").
		Funcs(template.FuncMap{"list": joinOrDash, "orDash": orDash}).
		Parse(constitutionTemplate),
)

var invIDRe = regexp.MustCompile(`^I\d+$`)

// runConstitution despacha `sf constitution <render|validate>`. Solo project dir
// (+ --stdout en render); no hay --feature.
func runConstitution(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf constitution <render|validate> [project_dir]")
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
			fmt.Fprintf(os.Stderr, "sf constitution: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	switch action {
	case "render":
		return renderConstitution(projectDir, toStdout)
	case "validate":
		return validateConstitution(projectDir)
	default:
		fmt.Fprintf(os.Stderr, "sf constitution: unknown action %q (use render|validate)\n", action)
		return 2
	}
}

func readConstitutionFile(projectDir string) (constitutionFile, int) {
	path := filepath.Join(projectDir, "specforge", "constitution.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf constitution: cannot read %s (%v)\n", path, err)
		return constitutionFile{}, 4
	}
	var cf constitutionFile
	if err := json.Unmarshal(data, &cf); err != nil {
		fmt.Fprintf(os.Stderr, "sf constitution: invalid JSON (%v)\n", err)
		return constitutionFile{}, 2
	}
	return cf, 0
}

func renderConstitution(projectDir string, toStdout bool) int {
	cf, code := readConstitutionFile(projectDir)
	if code != 0 {
		return code
	}
	var buf strings.Builder
	if err := constitutionTmpl.Execute(&buf, cf); err != nil {
		fmt.Fprintf(os.Stderr, "sf constitution: render failed (%v)\n", err)
		return 1
	}
	if toStdout {
		fmt.Print(buf.String())
		return 0
	}
	mdPath := filepath.Join(projectDir, "specforge", "constitution.md")
	if err := os.WriteFile(mdPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf constitution: cannot write %s (%v)\n", mdPath, err)
		return 1
	}
	fmt.Printf("rendered %s\n", filepath.Join("specforge", "constitution.md"))
	return 0
}

func validateConstitution(projectDir string) int {
	cf, code := readConstitutionFile(projectDir)
	if code != 0 {
		return code
	}
	var rep report
	checkConstitution(cf, &rep)

	fmt.Printf("Validated constitution: %d principle(s), %d invariant(s).\n", len(cf.Principles), len(cf.Invariants))
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
	fmt.Printf("\nOK: constitution.json valid (%d warning(s)).\n", len(rep.warnings))
	return 0
}

func checkConstitution(cf constitutionFile, rep *report) {
	if strings.TrimSpace(cf.IdentityMD) == "" {
		rep.warnf("identity_md is empty")
	}

	// Principios: id único + statement + applies_to válido.
	seenP := map[string]bool{}
	for i, p := range cf.Principles {
		label := p.ID
		if label == "" {
			label = fmt.Sprintf("principle #%d", i+1)
		}
		switch {
		case p.ID == "":
			rep.errorf("principle #%d: empty id", i+1)
		case seenP[p.ID]:
			rep.errorf("duplicate principle id %s", p.ID)
		default:
			seenP[p.ID] = true
		}
		if strings.TrimSpace(p.Statement) == "" {
			rep.errorf("%s: empty statement", label)
		}
		checkAppliesTo(label, p.AppliesTo, rep)
	}

	seen := map[string]bool{}
	for _, inv := range cf.Invariants {
		switch {
		case inv.ID == "":
			rep.errorf("invariant with empty id")
		case seen[inv.ID]:
			rep.errorf("duplicate invariant id %s", inv.ID)
		default:
			seen[inv.ID] = true
		}
		if inv.ID != "" && !invIDRe.MatchString(inv.ID) {
			rep.warnf("%s: id not in I# form", inv.ID)
		}
		if strings.TrimSpace(inv.Rule) == "" {
			rep.errorf("%s: empty rule", inv.ID)
		}
		checkAppliesTo(orDash(inv.ID), inv.AppliesTo, rep)
	}

	// Config del auditor de fase: si está, el modo debe ser off|nudge|block.
	if cf.Audit != nil && cf.Audit.Phase != "" && !auditPhaseModes[cf.Audit.Phase] {
		rep.errorf("audit.phase must be off|nudge|block, got %q", cf.Audit.Phase)
	}
	// Config del build: si está, el modo debe ser inline|single|per-wave.
	if cf.Build != nil && cf.Build.Mode != "" && !buildModes[cf.Build.Mode] {
		rep.errorf("build.mode must be inline|single|per-wave, got %q", cf.Build.Mode)
	}
	// Flujo (F2): serial | parallel.
	if cf.Flow != nil && cf.Flow.Mode != "" && !flowModes[cf.Flow.Mode] {
		rep.errorf("flow.mode must be serial|parallel, got %q", cf.Flow.Mode)
	}
	// blocking_priorities: sólo valores del enum. Un typo acá ("MUST", "high")
	// no afloja el gate por accidente — pero sí lo dejaría sin ese valor, así
	// que se rechaza en vez de ignorarse.
	if cf.Verification != nil {
		for i, p := range cf.Verification.BlockingPriorities {
			if !priorities[strings.TrimSpace(p)] {
				rep.errorf("verification.blocking_priorities[%d]: %q is not a priority (must|should|could)", i, p)
			}
		}
	}

	// Exclusiones de cobertura: cada una con ruta y MOTIVO. El motivo no es
	// ceremonia — excluir sube el porcentaje, y una exclusión sin razón escrita
	// es exactamente cómo se maquilla la métrica sin que quede rastro.
	if cf.Coverage != nil {
		seenPath := map[string]bool{}
		for i, e := range cf.Coverage.Exclude {
			label := fmt.Sprintf("coverage.exclude #%d", i+1)
			path := strings.TrimSpace(e.Path)
			switch {
			case path == "":
				rep.errorf("%s: path is required", label)
			case seenPath[path]:
				rep.errorf("%s: duplicate path %q", label, path)
			default:
				seenPath[path] = true
			}
			if strings.TrimSpace(e.Reason) == "" {
				rep.errorf("%s (%s): reason is required — an exclusion without a reason is indistinguishable from an oversight", label, orDash(path))
			}
		}
	}
	// Reporte por-test (A2): formato conocido, y junit exige el placeholder en
	// test_cmd (sin él, `sf check run` no tendría dónde leer el reporte).
	if cf.Build != nil && cf.Build.Report != "" {
		switch {
		case !buildReports[cf.Build.Report]:
			rep.errorf("build.report must be go-json|junit, got %q", cf.Build.Report)
		case cf.Build.Report == "junit" && !strings.Contains(cf.Build.TestCmd, "{report}"):
			rep.errorf(`build.report=junit requires a {report} placeholder in build.test_cmd (e.g. "pytest -q --junitxml={report}")`)
		case cf.Build.Report == "go-json" && !strings.Contains(cf.Build.TestCmd, "-json"):
			rep.warnf("build.report=go-json but test_cmd doesn't mention -json — the report will come back empty")
		}
	}
}

// checkAppliesTo valida el mapeo fase→regla. Dos roles del debate:
//   - WARN si no hay applies_to → la regla no la chequea ningún auditor de fase
//     (solo la atrapa sf-audit). Es el "lint WARN" sobre la constitución.
//   - ERROR si apunta a una fase inexistente → gate estructural (determinista)
//     sobre la constitución misma: lo hermético envuelve lo cooperativo.
func checkAppliesTo(label string, appliesTo []string, rep *report) {
	if len(appliesTo) == 0 {
		rep.warnf("%s: no applies_to — won't be checked by any phase auditor (only by sf-audit)", label)
		return
	}
	for _, ph := range appliesTo {
		if !auditablePhases[ph] {
			rep.errorf("%s: applies_to references unknown phase %q", label, ph)
		}
	}
}

// loadConstitutionQuiet lee constitution.json SIN imprimir (a diferencia de
// readConstitutionFile). Devuelve (cf, false) si no existe o no parsea — lo usa
// for-judge para degradar con una nota en vez de fallar.
func loadConstitutionQuiet(projectDir string) (constitutionFile, bool) {
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "constitution.json"))
	if err != nil {
		return constitutionFile{}, false
	}
	var cf constitutionFile
	if json.Unmarshal(data, &cf) != nil {
		return constitutionFile{}, false
	}
	return cf, true
}
