package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// ----------------------------------------------------------------------------
// `sf save`: la VÍA DE ESCRITURA validada. El LLM manda el JSON por stdin
// (--json -) o un archivo (--json path); el CLI valida (reusando los check*),
// y solo si pasa, escribe el .json canónico y renderiza el .md. Nada se guarda
// sin validar.
//
// Concepto nuevo: INTERFACES. En vez de repetir el flujo por cada artefacto,
// definimos un contrato `artifact` que los 6 tipos cumplen, y un finishSave
// genérico que trabaja contra ese contrato.
// ----------------------------------------------------------------------------

// artifact es el contrato: "sé validarme y sé renderizarme a markdown". Cualquier
// tipo con estos dos métodos lo cumple — en Go las interfaces se satisfacen de
// forma IMPLÍCITA (no se declara "implements").
type artifact interface {
	validate(*report)
	renderMarkdown() (string, error)
}

// execTemplate corre una plantilla a string (lo que necesitan los métodos render).
func execTemplate(t *template.Template, data any) (string, error) {
	var buf strings.Builder
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// --- Métodos adaptadores: cada tipo cumple `artifact` delegando en las
// funciones que ya escribimos. Receptor por VALOR → el valor (y el puntero)
// satisfacen la interfaz.

func (v constitutionFile) validate(r *report)              { checkConstitution(v, r) }
func (v constitutionFile) renderMarkdown() (string, error) { return execTemplate(constitutionTmpl, v) }

func (v domainFile) validate(r *report)              { checkDomain(v, r) }
func (v domainFile) renderMarkdown() (string, error) { return execTemplate(domainTmpl, v) }

func (v sourcesFile) validate(r *report)              { checkSources(v, r) }
func (v sourcesFile) renderMarkdown() (string, error) { return execTemplate(sourcesTmpl, v) }

func (v requirementsFile) validate(r *report)              { checkRequirements(v, r) }
func (v requirementsFile) renderMarkdown() (string, error) { return execTemplate(reqTmpl, v) }

func (v designFile) validate(r *report)              { checkDesign(v, r) }
func (v designFile) renderMarkdown() (string, error) { return execTemplate(designTmpl, v) }

func (v tasksFile) validate(r *report)              { checkTasks(v, r) }
func (v tasksFile) renderMarkdown() (string, error) { return execTemplate(tasksTmpl, v) }

func (v planFile) validate(r *report)              { checkPlanInternal(v, r) }
func (v planFile) renderMarkdown() (string, error) { return execTemplate(planTmpl, v) }

func (v reviewFile) validate(r *report)              { checkReview(v, r) }
func (v reviewFile) renderMarkdown() (string, error) { return execTemplate(reviewTmpl, v) }

func (v traceFile) validate(r *report)              { checkTrace(v, r) }
func (v traceFile) renderMarkdown() (string, error) { return renderTraceMarkdown(v) }

// runSave parsea `sf save <artifact> [--feature=X] [--json -|FILE] [--from=DRAFT] [dir]`.
func runSave(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: sf save <artifact> [--feature=NAME] [--json -|FILE] [--from=DRAFT] [project_dir]")
		return 2
	}
	name, rest := args[0], args[1:]

	projectDir := "."
	feature := ""
	jsonSrc := "-" // por defecto, stdin
	fromDraft := ""
	for i := 0; i < len(rest); i++ {
		a := rest[i]
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--json="):
			jsonSrc = strings.TrimPrefix(a, "--json=")
		case strings.HasPrefix(a, "--from="):
			fromDraft = strings.TrimPrefix(a, "--from=")
		case a == "--json":
			if i+1 < len(rest) {
				jsonSrc = rest[i+1]
				i++ // consumimos el valor
			}
		case a != "-" && strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf save: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	// constitution, domain y sources son a nivel proyecto; el resto necesita --feature.
	if name != "constitution" && name != "domain" && name != "sources" && feature == "" {
		fmt.Fprintln(os.Stderr, "sf save: --feature=NAME is required")
		return 2
	}

	// --from: el CAMINO CÓMODO de autoría (R5/D8). El agente escribe el borrador
	// con su tool Write en drafts/ (dir NO protegido por el hook) y acá lo
	// promovemos: validar → escribir el canónico → borrar el borrador. La
	// garantía es la misma (solo `sf` escribe el .json canónico) sin el escaping
	// frágil de heredocs por stdin.
	if fromDraft != "" {
		if jsonSrc != "-" {
			fmt.Fprintln(os.Stderr, "sf save: --from and --json are mutually exclusive — pick one input")
			return 2
		}
		jsonSrc = draftPath(projectDir, feature, fromDraft)
	}

	// La ruta (b) de DL-4, con dientes: si hay un defecto de código abierto
	// contra esta feature, alguien ya decidió que el SPEC TENÍA RAZÓN. Editarlo
	// ahora sería deshacer esa decisión sin declararlo.
	if deltaSpecEditGuard(projectDir, feature, name) {
		return 5
	}

	raw, err := readJSONInput(jsonSrc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf save: cannot read input (%v)\n", err)
		return 1
	}

	jsonPath, mdPath := artifactPaths(name, projectDir, feature)

	// Decodificamos en el tipo concreto según el artefacto, y de ahí en más todo
	// es genérico vía la interfaz. Cada rama deja el valor en `a`; el final
	// (guardar + promover el draft) es común a todos.
	var a artifact
	switch name {
	case "constitution":
		var v constitutionFile
		if !decodeInto(raw, &v) {
			return 2
		}
		if v.SchemaVersion == "" {
			v.SchemaVersion = schemaVersionCurrent
		}
		a = v
	case "domain":
		var v domainFile
		if !decodeInto(raw, &v) {
			return 2
		}
		if v.SchemaVersion == "" {
			v.SchemaVersion = schemaVersionCurrent
		}
		a = v
	case "sources":
		var v sourcesFile
		if !decodeInto(raw, &v) {
			return 2
		}
		if v.SchemaVersion == "" {
			v.SchemaVersion = schemaVersionCurrent
		}
		a = v
	case "requirements":
		var v requirementsFile
		if !decodeInto(raw, &v) {
			return 2
		}
		if v.SchemaVersion == "" {
			v.SchemaVersion = schemaVersionCurrent
		}
		// R2: los refs de fuente se validan contra sources.json, que la
		// interfaz `artifact` no puede ver (no conoce el project dir). Se
		// chequea acá, ANTES de escribir nada.
		var rep report
		checkRequirementsIn(v, projectDir, &rep)
		if len(rep.errors) > 0 {
			for _, e := range rep.errors {
				fmt.Printf("  ERROR:   %s\n", e)
			}
			fmt.Printf("\nFAIL: not saved — %d error(s).\n", len(rep.errors))
			return 2
		}
		a = v
	case "design":
		var v designFile
		if !decodeInto(raw, &v) {
			return 2
		}
		if v.SchemaVersion == "" {
			v.SchemaVersion = schemaVersionCurrent
		}
		a = v
	case "tasks":
		var v tasksFile
		if !decodeInto(raw, &v) {
			return 2
		}
		if v.SchemaVersion == "" {
			v.SchemaVersion = schemaVersionCurrent
		}
		a = v
	case "plan":
		var v planFile
		if !decodeInto(raw, &v) {
			return 2
		}
		if v.SchemaVersion == "" {
			v.SchemaVersion = schemaVersionCurrent
		}
		a = v
	case "review":
		var v reviewFile
		if !decodeInto(raw, &v) {
			return 2
		}
		if v.SchemaVersion == "" {
			v.SchemaVersion = schemaVersionCurrent
		}
		a = v
	case "trace":
		var v traceFile
		if !decodeInto(raw, &v) {
			return 2
		}
		// Si el agente no mandó schema_version, lo fijamos (canónico, como el resto).
		if v.SchemaVersion == "" {
			v.SchemaVersion = schemaVersionCurrent
		}
		a = v
	default:
		fmt.Fprintf(os.Stderr, "sf save: unknown artifact %q\n", name)
		return 2
	}

	code := finishSave(a, jsonPath, mdPath)
	// Promoción: el borrador ya vive validado en el canónico → lo retiramos para
	// que no queden dos copias divergentes dando vueltas. Solo si el save fue OK.
	if code == 0 && fromDraft != "" {
		if err := os.Remove(jsonSrc); err == nil {
			fmt.Printf("promoted draft %s (removed after save)\n", jsonSrc)
		}
	}
	return code
}

// specArtifacts: los artefactos que SON el contrato de la feature. Un delta
// code-wrong los congela. `trace` queda afuera a propósito — arreglar el defecto
// implica re-anclar código y tests, y bloquear eso dejaría la ruta (b) sin
// salida. `review` y `plan` tampoco son el contrato.
var specArtifacts = map[string]bool{"requirements": true, "design": true, "tasks": true}

// deltaSpecEditGuard imprime el rechazo y devuelve true si hay que abortar.
func deltaSpecEditGuard(projectDir, feature, artifact string) bool {
	if feature == "" || !specArtifacts[artifact] {
		return false
	}
	d, blocked := deltaBlockingSpecEdits(projectDir, feature)
	if !blocked {
		return false
	}
	fmt.Fprintf(os.Stderr, "sf save: refusing to edit %s — delta %s (%s) is open on %s\n",
		artifact, d.ID, deltaKindCodeWrong, feature)
	fmt.Fprintf(os.Stderr, "  %s declares the spec was RIGHT and the code is wrong:\n", d.ID)
	fmt.Fprintf(os.Stderr, "    expected: %s\n", d.Expected)
	fmt.Fprintf(os.Stderr, "    observed: %s\n", d.Observed)
	fmt.Fprintln(os.Stderr, "  Fix the code, or — if you changed your mind about which side was right —")
	fmt.Fprintf(os.Stderr, "  close it explicitly: sf delta set-status --feature=%s --id=%s --to=archived\n", feature, d.ID)
	return true
}

// draftPath resuelve el nombre de un borrador contra su dir de drafts (el ÚNICO
// rincón de specforge/ donde el hook permite Write directo): artefactos de
// feature → specforge/features/<f>/drafts/; los de nivel proyecto (constitution,
// domain) → specforge/drafts/. Aceptamos "tasks.json" y "drafts/tasks.json"
// (como lo teclearía el agente); un path absoluto se usa tal cual.
func draftPath(projectDir, feature, from string) string {
	if filepath.IsAbs(from) {
		return from
	}
	rel := strings.TrimPrefix(filepath.ToSlash(from), "drafts/")
	base := filepath.Join(projectDir, "specforge", "drafts")
	if feature != "" {
		base = filepath.Join(projectDir, "specforge", "features", feature, "drafts")
	}
	return filepath.Join(base, filepath.FromSlash(rel))
}

// decodeInto desempaqueta el JSON e informa el error (devuelve false si falla).
func decodeInto(raw []byte, v any) bool {
	if err := json.Unmarshal(raw, v); err != nil {
		fmt.Fprintf(os.Stderr, "sf save: invalid JSON (%v)\n", err)
		return false
	}
	return true
}

// finishSave es el corazón genérico: valida, y solo si pasa, escribe el JSON
// canónico + el markdown. Trabaja contra la INTERFAZ, no contra un tipo concreto.
func finishSave(a artifact, jsonPath, mdPath string) int {
	var rep report
	a.validate(&rep)
	for _, w := range rep.warnings {
		fmt.Printf("  warning: %s\n", w)
	}
	if len(rep.errors) > 0 {
		for _, e := range rep.errors {
			fmt.Printf("  ERROR:   %s\n", e)
		}
		fmt.Printf("\nFAIL: not saved — %d error(s).\n", len(rep.errors))
		return 2
	}

	if err := os.MkdirAll(filepath.Dir(jsonPath), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf save: %v\n", err)
		return 1
	}

	// Re-serializamos desde el struct → JSON canónico (formato consistente,
	// descarta claves desconocidas). Marshal funciona sobre el valor que lleva
	// adentro la interfaz.
	canonical, err := json.MarshalIndent(a, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf save: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(jsonPath, append(canonical, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf save: cannot write %s (%v)\n", jsonPath, err)
		return 1
	}

	md, err := a.renderMarkdown()
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf save: render failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf save: cannot write %s (%v)\n", mdPath, err)
		return 1
	}

	fmt.Printf("saved %s (+ rendered %s)\n", jsonPath, mdPath)
	return 0
}

// readJSONInput lee de stdin ("-") o de un archivo.
func readJSONInput(src string) ([]byte, error) {
	if src == "-" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(src)
}

// artifactPaths devuelve (jsonPath, mdPath) según el artefacto. constitution es
// a nivel proyecto; plan vive bajo progress/; el resto bajo features/<f>/.
func artifactPaths(name, projectDir, feature string) (string, string) {
	fdir := filepath.Join(projectDir, "specforge", "features", feature)
	switch name {
	case "constitution":
		base := filepath.Join(projectDir, "specforge", "constitution")
		return base + ".json", base + ".md"
	case "domain":
		base := domainPath(projectDir) // specforge/context/domain
		return base + ".json", base + ".md"
	case "sources":
		base := sourcesPath(projectDir) // specforge/sources
		return base + ".json", base + ".md"
	case "plan":
		base := filepath.Join(fdir, "progress", "plan")
		return base + ".json", base + ".md"
	default: // requirements, design, tasks, review
		base := filepath.Join(fdir, name)
		return base + ".json", base + ".md"
	}
}
