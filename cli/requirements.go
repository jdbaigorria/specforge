package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"text/template"
)

// ----------------------------------------------------------------------------
// Segundo artefacto del piloto JSON-first: requirements. Misma regla que tasks:
// requirements.json es la FUENTE, requirements.md el RENDER. Estructurar esto es
// lo que le permite a `sf context for-wave` incluir los CUERPOS de los R# en
// scope (no solo los IDs) — el ahorro real de tokens.
// ----------------------------------------------------------------------------

// Modelo de requirements.json (CLI-SPEC §4.2). Nota: en JSON `state`/`trigger`
// pueden venir como null; al decodificar en un string quedan "" (no
// distinguimos null de ausente, y para EARS no hace falta).
type requirementsFile struct {
	SchemaVersion string        `json:"schema_version"`
	Feature       string        `json:"feature"`
	Summary       string        `json:"summary"`
	Actors        []string      `json:"actors"`
	Requirements  []requirement `json:"requirements"`
}

type requirement struct {
	ID       string `json:"id"`
	EarsType string `json:"ears_type"`
	// Priority gradúa la severidad del gate (RM-C2). `omitempty` + default
	// implícito: los requisitos escritos antes del campo se leen como `must`.
	Priority   string         `json:"priority,omitempty"`
	Trigger    string         `json:"trigger"`
	State      string         `json:"state"`
	Behavior   string         `json:"behavior"`
	Acceptance acceptanceList `json:"acceptance"`
	// Source pasó de string libre a REFS a sources.json (RM-C3). Hasta ahora era
	// un campo muerto: existía en el struct, no lo renderizaba el template y no
	// lo leía nadie.
	Source sourceRefs `json:"source,omitempty"`
	Tags   []string   `json:"tags"`
}

// sourceRefs son las fuentes de un requisito, con doble lectura como
// `acceptance`: acepta la forma legada (un string suelto de prosa libre) y la
// v2 (una lista de refs `S#`).
//
// Un string que YA tiene forma de ref (`"S1"`) entra como ref — no hay
// ambigüedad ahí. Cualquier otro string se preserva como prosa: no lo
// convertimos ni lo tiramos. Convertir prosa en un ref sería inventar un id que
// nadie declaró, y tirarla sería perder el único rastro de procedencia que ese
// requisito tenía.
type sourceRefs struct {
	Refs  []string
	Prose string
}

func (s *sourceRefs) UnmarshalJSON(data []byte) error {
	trimmed := strings.TrimSpace(string(data))
	if trimmed == "null" {
		return nil
	}
	if strings.HasPrefix(trimmed, `"`) {
		var one string
		if err := json.Unmarshal(data, &one); err != nil {
			return err
		}
		if one = strings.TrimSpace(one); one == "" {
			return nil
		}
		if sourceIDRe.MatchString(one) {
			s.Refs = []string{one}
			return nil
		}
		s.Prose = one
		return nil
	}
	var many []string
	if err := json.Unmarshal(data, &many); err != nil {
		return fmt.Errorf("source must be a ref list or a string: %w", err)
	}
	s.Refs = many
	return nil
}

// MarshalJSON re-emite la forma que entró, por el mismo motivo que
// `acceptanceList`: un `sf save` no debe convertir prosa en refs por su cuenta.
func (s sourceRefs) MarshalJSON() ([]byte, error) {
	if s.Prose != "" && len(s.Refs) == 0 {
		return json.Marshal(s.Prose)
	}
	if s.Refs == nil {
		return json.Marshal([]string{})
	}
	return json.Marshal(s.Refs)
}

// empty: ni refs ni prosa. `omitempty` no funciona con structs, así que el
// campo se emite siempre; esto es para las decisiones de validación.
func (s sourceRefs) empty() bool { return len(s.Refs) == 0 && s.Prose == "" }

// sourceLine rinde la procedencia para el markdown. La prosa legada se marca
// como tal: que se vea que ese requisito todavía no tiene procedencia
// verificable es la mitad del punto.
func sourceLine(r requirement) string {
	if len(r.Source.Refs) > 0 {
		return strings.Join(r.Source.Refs, ", ")
	}
	if r.Source.Prose != "" {
		return r.Source.Prose + " _(free prose — not a ref)_"
	}
	return ""
}

// ----------------------------------------------------------------------------
// RM-C2 — `priority` tipada.
//
// Sin este eje, todo requisito pesa lo mismo: `sf coverage` no distingue un 78%
// al que le falta todo lo `must` de uno al que le falta todo lo `could`, y el
// gate del veredicto sólo sabe bloquear o no bloquear.
//
// EL ANTÍDOTO A LA FATIGA DE GATE. Un gate que bloquea por todo se termina
// salteando por todo — no porque alguien sea deshonesto, sino porque un
// bloqueo que no discrimina deja de informar. Graduar la severidad es lo que
// hace que un bloqueo vuelva a significar algo.
// ----------------------------------------------------------------------------

// priorities: MoSCoW sin el "won't" — un requisito que no se va a hacer no se
// escribe, se borra.
var priorities = map[string]bool{"must": true, "should": true, "could": true}

// priorityOf normaliza: ausente ⇒ `must`.
//
// FAIL-CLOSED, y es deliberado. El default tenía que ser el extremo más
// exigente: si un requisito sin prioridad declarada valiera `could`, omitir el
// campo sería la forma más barata de bajar el listón, y un agente bajo presión
// de contexto encuentra esas formas solo. Acá omitirlo cuesta más, no menos.
func priorityOf(r requirement) string {
	if r.Priority == "" {
		return "must"
	}
	return r.Priority
}

// ----------------------------------------------------------------------------
// RM-C1 — el criterio de aceptación con ID.
//
// EL AGUJERO QUE TAPA. El contrato de verificación se cumple a granularidad de
// requisito (`cli/gate.go`: "R5: names no test"). O sea: un requisito con cinco
// criterios de aceptación y UN test pasa el contrato, sale verde y queda sellado
// con hash encadenado. No hace falta mala fe — el agente cumple lo que se le
// pide; la regla es la que mide poco.
//
// POR QUÉ EL ID Y NO GIVEN/WHEN/THEN OBLIGATORIO (DEC-1, 2026-08-05). Lo que
// tapa el agujero es poder ANCLAR un test a cada criterio, y para eso alcanza
// con que el criterio tenga nombre. El humano aprueba `R5.2`; la máquina exige
// un test para `R5.2`; son el mismo objeto. Given/When/Then agrega rigor de
// REDACCIÓN, no de verificación — así que entra como opción.
//
// Medido sobre los ejemplos del repo: los criterios que el proyecto escribe de
// verdad ya son asertos atómicos (`slugify("Café Olé") == "cafe-ole"`), no prosa.
// Envolverlos en Given/When/Then sería ceremonia sobre algo que ya es preciso.
// ----------------------------------------------------------------------------

// acceptanceCriterion es UN caso observable que prueba el requisito.
//
// `text` y la terna given/when/then son dos formas de escribir lo mismo: la
// corta para un aserto que ya se explica solo, la larga cuando la precondición
// aporta. Nunca las dos.
type acceptanceCriterion struct {
	ID    string `json:"id"`
	Text  string `json:"text,omitempty"`
	Given string `json:"given,omitempty"`
	When  string `json:"when,omitempty"`
	Then  string `json:"then,omitempty"`
}

// structured responde si el criterio usa la forma rica.
func (c acceptanceCriterion) structured() bool {
	return c.When != "" || c.Then != "" || c.Given != ""
}

// line devuelve el criterio como una sola línea, para el render y los mensajes.
func (c acceptanceCriterion) line() string {
	if !c.structured() {
		return c.Text
	}
	var b strings.Builder
	if c.Given != "" {
		fmt.Fprintf(&b, "Given %s, ", c.Given)
	}
	fmt.Fprintf(&b, "when %s, then %s", c.When, c.Then)
	return b.String()
}

// acceptanceList es la lista de criterios, y sabe leerse en las DOS formas:
//
//	legado: ["slugify(\"\") == \"\"", "..."]          → sin ids
//	v2:     [{"id":"R4.1","text":"..."}, ...]         → con ids
//
// Concepto Go: implementando `UnmarshalJSON` y `MarshalJSON` sobre un tipo
// propio, ese tipo toma el control total de cómo se serializa. `encoding/json`
// chequea si el valor satisface las interfaces `json.Unmarshaler` /
// `json.Marshaler` y, si las satisface, delega en ellas en vez de usar la
// reflexión por defecto. Es el gancho que permite aceptar dos esquemas distintos
// para el mismo campo sin duplicar el struct entero.
type acceptanceList []acceptanceCriterion

// UnmarshalJSON acepta las dos formas. El array vacío y `null` se leen como lista
// vacía, sin error.
//
// Concepto Go: el receptor es un PUNTERO (`*acceptanceList`) porque el método
// tiene que MODIFICAR el valor. Con receptor por valor escribiríamos sobre una
// copia y el llamador no vería nada — y json ni siquiera lo tomaría como
// Unmarshaler.
func (a *acceptanceList) UnmarshalJSON(data []byte) error {
	// json.RawMessage difiere el parseo: guardamos los bytes crudos de cada
	// elemento y decidimos elemento por elemento cómo interpretarlos.
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("acceptance must be a list: %w", err)
	}

	out := make(acceptanceList, 0, len(raw))
	for i, item := range raw {
		trimmed := strings.TrimSpace(string(item))
		if strings.HasPrefix(trimmed, `"`) {
			// Forma legada: un string suelto. Entra SIN id, y esa ausencia es
			// exactamente lo que después distingue el contrato v1 del v2.
			var s string
			if err := json.Unmarshal(item, &s); err != nil {
				return fmt.Errorf("acceptance[%d]: %w", i, err)
			}
			out = append(out, acceptanceCriterion{Text: s})
			continue
		}
		var c acceptanceCriterion
		if err := json.Unmarshal(item, &c); err != nil {
			return fmt.Errorf("acceptance[%d]: expected a string or an object with an id: %w", i, err)
		}
		out = append(out, c)
	}
	*a = out
	return nil
}

// MarshalJSON re-emite la MISMA forma que entró.
//
// Por qué importa: `sf save` re-serializa el JSON canónico. Si un requisito
// legado (strings sin id) volviera escrito como objetos, un simple save lo
// habría subido de contrato v1 a v2 — y con eso el veredicto pasaría a exigir un
// test por criterio sin que nadie lo haya decidido. Subir de contrato es un acto
// EXPLÍCITO (`sf migrate`, que asigna los ids, o escribir los objetos a mano),
// nunca el efecto colateral de guardar.
//
// El discriminador sale del dato, no de un flag escondido: sin ids y sin forma
// rica ⇒ es legado.
func (a acceptanceList) MarshalJSON() ([]byte, error) {
	if a.legacy() {
		out := make([]string, len(a))
		for i, c := range a {
			out[i] = c.Text
		}
		return json.Marshal(out)
	}
	// Alias de tipo para evitar la recursión infinita: si hiciéramos
	// json.Marshal(a) acá, json volvería a llamar a este mismo método. `plain`
	// tiene la misma representación pero NO hereda los métodos, así que json usa
	// su reflexión normal.
	type plain []acceptanceCriterion
	return json.Marshal(plain(a))
}

// legacy: ningún criterio tiene id ni forma rica. Una lista vacía cuenta como
// legada — no hay nada que distinga una cosa de la otra, y emitir `[]` es lo
// mismo en las dos formas.
func (a acceptanceList) legacy() bool {
	for _, c := range a {
		if c.ID != "" || c.structured() {
			return false
		}
	}
	return true
}

// hasIDs: ¿el requisito corre bajo el contrato v2? Basta UN id para que el
// requisito entre al contrato estricto; los que falten los caza la validación.
func (a acceptanceList) hasIDs() bool {
	for _, c := range a {
		if c.ID != "" {
			return true
		}
	}
	return false
}

// acceptanceIDRe valida la forma de un id de criterio: el id del requisito
// padre, un punto, y un entero. El prefijo hace evidente la pertenencia sin
// necesidad de un índice aparte.
var acceptanceIDRe = regexp.MustCompile(`^(R\d+)\.(\d+)$`)

//go:embed templates/requirements.tmpl.md
var reqTemplate string

// reqTmpl agrega la función `ears`, que arma la oración EARS según el tipo.
var reqTmpl = template.Must(
	template.New("requirements").
		Funcs(template.FuncMap{
			"list":   joinOrDash,
			"orDash": orDash,
			"ears":   earsSentence,
			// `criterion` rinde el criterio en una línea. Given/When/Then entra
			// como RENDER, igual que `ears` arma la oración EARS desde campos
			// estructurados: la fuente sigue siendo JSON.
			"criterion": acceptanceCriterion.line,
			// `source` rinde la procedencia. Hasta RM-C3 el campo existía y el
			// template no lo mostraba — un dato muerto en los dos extremos.
			"source": sourceLine,
		}).
		Parse(reqTemplate),
)

// earsTypes es el enum válido de tipos EARS, como set.
var earsTypes = map[string]bool{
	"event": true, "state": true, "error": true, "ubiquitous": true,
}

// earsSentence arma la oración EARS canónica. Le saca un "shall " inicial al
// behavior para no duplicarlo ("the system shall shall ...").
func earsSentence(r requirement) string {
	b := strings.TrimPrefix(r.Behavior, "shall ")
	switch r.EarsType {
	case "event":
		return fmt.Sprintf("When %s, the system shall %s.", r.Trigger, b)
	case "state":
		return fmt.Sprintf("While %s, the system shall %s.", r.State, b)
	case "error":
		return fmt.Sprintf("If %s, then the system shall %s.", r.Trigger, b)
	default: // ubiquitous
		return fmt.Sprintf("The system shall %s.", b)
	}
}

// runRequirements despacha `sf requirements <render|validate>`.
func runRequirements(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf requirements <render|validate> --feature=NAME [project_dir]")
		return 2
	}
	action, rest := args[0], args[1:]
	switch action {
	case "render":
		return reqRenderCmd(rest)
	case "validate":
		return reqValidateCmd(rest)
	default:
		fmt.Fprintf(os.Stderr, "sf requirements: unknown action %q (use render|validate)\n", action)
		return 2
	}
}

// parseArtifactFlags extrae --feature, --stdout y el project dir. Es compartido
// por los artefactos JSON-first (requirements, design, ...). `cmd` solo se usa
// para los mensajes de error.
func parseArtifactFlags(cmd string, args []string) (projectDir, feature string, toStdout bool, ok bool) {
	projectDir = "."
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case a == "--stdout":
			toStdout = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf %s: unknown flag %q\n", cmd, a)
			return "", "", false, false
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintf(os.Stderr, "sf %s: --feature=NAME is required\n", cmd)
		return "", "", false, false
	}
	return projectDir, feature, toStdout, true
}

func reqRenderCmd(args []string) int {
	projectDir, feature, toStdout, ok := parseArtifactFlags("requirements", args)
	if !ok {
		return 2
	}
	return renderRequirements(projectDir, feature, toStdout)
}

func reqValidateCmd(args []string) int {
	projectDir, feature, _, ok := parseArtifactFlags("requirements", args)
	if !ok {
		return 2
	}
	return validateRequirements(projectDir, feature)
}

// readRequirementsFile lee y parsea requirements.json. Devuelve (rf, exitCode);
// exitCode 0 si está OK.
func readRequirementsFile(projectDir, feature string) (requirementsFile, int) {
	path := filepath.Join(projectDir, "specforge", "features", feature, "requirements.json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf requirements: cannot read %s (%v)\n", path, err)
		return requirementsFile{}, 4
	}
	var rf requirementsFile
	if err := json.Unmarshal(data, &rf); err != nil {
		fmt.Fprintf(os.Stderr, "sf requirements: invalid JSON (%v)\n", err)
		return requirementsFile{}, 2
	}
	return rf, 0
}

func renderRequirements(projectDir, feature string, toStdout bool) int {
	rf, code := readRequirementsFile(projectDir, feature)
	if code != 0 {
		return code
	}

	var buf strings.Builder
	if err := reqTmpl.Execute(&buf, rf); err != nil {
		fmt.Fprintf(os.Stderr, "sf requirements: render failed (%v)\n", err)
		return 1
	}

	if toStdout {
		fmt.Print(buf.String())
		return 0
	}
	mdPath := filepath.Join(projectDir, "specforge", "features", feature, "requirements.md")
	if err := os.WriteFile(mdPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf requirements: cannot write %s (%v)\n", mdPath, err)
		return 1
	}
	fmt.Printf("rendered %s\n", filepath.Join("specforge", "features", feature, "requirements.md"))
	return 0
}

func validateRequirements(projectDir, feature string) int {
	rf, code := readRequirementsFile(projectDir, feature)
	if code != 0 {
		return code
	}

	var rep report
	checkRequirements(rf, &rep)

	fmt.Printf("Validated %s: %d requirement(s).\n", feature, len(rf.Requirements))
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
	fmt.Printf("\nOK: requirements.json valid (%d warning(s)).\n", len(rep.warnings))
	return 0
}

// checkRequirements valida la gramática EARS (CLI-SPEC §4.2): cada tipo exige
// ciertos campos y prohíbe otros.
//
// Es el validador que consume `sf save` vía la interfaz `artifact`, y no conoce
// el project dir — por eso los refs de `source` se chequean aparte, en
// checkRequirementsIn.
func checkRequirements(rf requirementsFile, rep *report) {
	checkRequirementsIn(rf, "", rep)
}

// checkRequirementsIn agrega, cuando hay project dir, la validación cruzada
// contra sources.json (R2).
func checkRequirementsIn(rf requirementsFile, projectDir string, rep *report) {
	if rf.Feature == "" {
		rep.errorf("missing `feature`")
	}

	// R2: los refs de fuente tienen que existir. Un ref colgando es peor que no
	// declarar fuente — declara procedencia y no la tiene, así que pasaría el
	// chequeo de "requisito sin fuente" sin haberlo cumplido.
	var declared map[string]bool
	if projectDir != "" {
		declared = sourceIDSet(projectDir)
	}
	requireSource := projectDir != "" && requireSourceEnabled(projectDir)

	seen := map[string]bool{}
	for _, r := range rf.Requirements {
		switch {
		case r.ID == "":
			rep.errorf("requirement with empty id")
		case seen[r.ID]:
			rep.errorf("duplicate requirement id %s", r.ID)
		default:
			seen[r.ID] = true
		}

		if !priorities[priorityOf(r)] {
			rep.errorf("%s: invalid priority %q (must|should|could)", r.ID, r.Priority)
		}

		if !earsTypes[r.EarsType] {
			rep.errorf("%s: invalid ears_type %q", r.ID, r.EarsType)
			continue // sin tipo válido no tiene sentido chequear los campos
		}

		// Reglas EARS por tipo.
		switch r.EarsType {
		case "event":
			if r.Trigger == "" {
				rep.errorf("%s: event requires a trigger", r.ID)
			}
			if r.State != "" {
				rep.errorf("%s: event must not set state", r.ID)
			}
		case "state":
			if r.State == "" {
				rep.errorf("%s: state requires a state", r.ID)
			}
			if r.Trigger != "" {
				rep.errorf("%s: state must not set trigger", r.ID)
			}
		case "error":
			if r.Trigger == "" {
				rep.errorf("%s: error requires a trigger (condition)", r.ID)
			}
			if r.Behavior == "" {
				rep.errorf("%s: error requires a behavior (response)", r.ID)
			}
		case "ubiquitous":
			if r.Trigger != "" || r.State != "" {
				rep.errorf("%s: ubiquitous must not set trigger/state", r.ID)
			}
			if r.Behavior == "" {
				rep.errorf("%s: ubiquitous requires a behavior", r.ID)
			}
		}

		checkAcceptance(r, rep)
		checkSourceRefs(r, declared, requireSource, rep)
	}
}

// checkSourceRefs valida la procedencia de UN requisito (R2).
//
// `declared == nil` significa "no sabemos qué fuentes existen" (validación sin
// project dir): en ese caso sólo se chequea la forma del ref, nunca su
// existencia. Inventar un error de "S9 no existe" cuando ni siquiera se pudo
// abrir sources.json sería reportar una ausencia que no se midió.
func checkSourceRefs(r requirement, declared map[string]bool, requireSource bool, rep *report) {
	if r.Source.Prose != "" {
		rep.warnf("%s: source is free prose, not a ref (%q) — declare it in sources.json and cite its id, "+
			"otherwise nothing can check where this requirement came from", r.ID, r.Source.Prose)
	}
	if r.Source.empty() && requireSource {
		rep.errorf("%s: no source — with verification.require_source enabled, a requirement must cite "+
			"where it came from (a requirement with no source was invented by the model)", r.ID)
	}

	seen := map[string]bool{}
	for _, ref := range r.Source.Refs {
		ref = strings.TrimSpace(ref)
		switch {
		case !sourceIDRe.MatchString(ref):
			rep.errorf("%s: source ref %q must have the form S<n>", r.ID, ref)
			continue
		case seen[ref]:
			rep.warnf("%s: source ref %s listed twice", r.ID, ref)
			continue
		}
		seen[ref] = true
		if declared != nil && !declared[ref] {
			rep.errorf("%s: source ref %s is not declared in sources.json", r.ID, ref)
		}
	}
}

// checkAcceptance valida los criterios de un requisito (RM-C1 / R1).
//
// La regla de entrada: o TODOS los criterios tienen id, o NINGUNO. Una lista
// mitad y mitad es la peor de las tres — el contrato del veredicto pregunta
// "¿este requisito corre bajo v2?" y una respuesta ambigua ahí significa
// criterios que nadie verifica y nadie ve.
func checkAcceptance(r requirement, rep *report) {
	if len(r.Acceptance) == 0 {
		rep.warnf("%s: no acceptance criteria — nothing can anchor a test to it", r.ID)
		return
	}
	if !r.Acceptance.hasIDs() {
		return // legado íntegro: conserva la regla vieja, sin ruido (R4)
	}

	seen := map[string]bool{}
	for i, c := range r.Acceptance {
		label := fmt.Sprintf("%s: acceptance #%d", r.ID, i+1)

		// 1) id presente y bien formado.
		switch {
		case c.ID == "":
			rep.errorf("%s: missing id — all criteria of a requirement must have one, or none may "+
				"(a half-structured requirement runs under neither contract)", label)
			continue
		case seen[c.ID]:
			rep.errorf("%s: duplicate criterion id %s", label, c.ID)
			continue
		}
		seen[c.ID] = true

		m := acceptanceIDRe.FindStringSubmatch(c.ID)
		if m == nil {
			rep.errorf("%s: id %q must have the form <requirement>.<n>, e.g. %s.1", label, c.ID, r.ID)
			continue
		}
		// 2) el prefijo tiene que ser el requisito padre. Un `R6.1` colgando de
		//    `R5` ancla tests a un requisito que no es el suyo.
		if m[1] != r.ID {
			rep.errorf("%s: id %s belongs to %s, not to %s", label, c.ID, m[1], r.ID)
			continue
		}
		// 3) el índice arranca en 1. El 0 no es un criterio, es un off-by-one.
		if n, err := strconv.Atoi(m[2]); err == nil && n < 1 {
			rep.errorf("%s: id %s — criterion numbering starts at 1", label, c.ID)
		}

		// 4) contenido: la forma corta o la larga, nunca media larga.
		switch {
		case c.structured() && c.Text != "":
			rep.errorf("%s (%s): set either `text` or given/when/then, not both", label, c.ID)
		case c.structured():
			if c.When == "" || c.Then == "" {
				rep.errorf("%s (%s): given/when/then requires both `when` and `then` "+
					"(half the rich form says less than the short one)", label, c.ID)
			}
		case strings.TrimSpace(c.Text) == "":
			rep.errorf("%s (%s): empty — needs `text`, or `when` + `then`", label, c.ID)
		}
	}

	// 5) LA REGLA DE ESTABILIDAD, dicha donde se lee. No se puede hacer cumplir
	//    acá (la validación no tiene memoria del archivo anterior), así que la
	//    hacemos visible: los huecos son LEGALES y son la señal de que un id se
	//    retiró. Renumerar para taparlos es lo que rompe anclas en silencio.
	if gaps := acceptanceGaps(r); len(gaps) > 0 {
		rep.warnf("%s: criterion ids skip %s — that is fine if a criterion was retired. "+
			"Never renumber to close a gap: trace.json anchors point at ids, and re-using a freed id "+
			"silently re-points them at a different case", r.ID, strings.Join(gaps, ", "))
	}
}

// acceptanceGaps devuelve los índices faltantes entre 1 y el máximo declarado.
func acceptanceGaps(r requirement) []string {
	present := map[int]bool{}
	max := 0
	for _, c := range r.Acceptance {
		m := acceptanceIDRe.FindStringSubmatch(c.ID)
		if m == nil || m[1] != r.ID {
			continue
		}
		n, err := strconv.Atoi(m[2])
		if err != nil || n < 1 {
			continue
		}
		present[n] = true
		if n > max {
			max = n
		}
	}
	var gaps []string
	for n := 1; n < max; n++ {
		if !present[n] {
			gaps = append(gaps, fmt.Sprintf("%s.%d", r.ID, n))
		}
	}
	return gaps
}
