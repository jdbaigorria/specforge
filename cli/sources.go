package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"
)

// ----------------------------------------------------------------------------
// RM-C3 — `sources.json`: de dónde salió cada requisito.
//
// EL AGUJERO, EN EL OTRO EXTREMO DEL TUBO. `trace.json` impide que el agente
// declare "implementé R5" sin código ni test. Nada impedía que el agente
// **inventara R5**. Es el mismo hueco de fabricación, aguas arriba: el detector
// existía sólo del lado de la salida.
//
// La versión ingenua sería "que la skill lea la carpeta del cliente". Eso es un
// cambio de prompt: no aporta ninguna garantía y empuja hacia una segunda
// metodología de elicitación. La versión con forma de SpecForge es que `source`
// deje de ser un string libre y pase a ser un REF, igual que R#, C#, E#, D#.
//
// Con eso salen dos chequeos deterministas que antes no existían:
//
//   - requisito SIN fuente  → lo inventó el modelo
//   - fuente SIN requisito  → material que se leyó y no se usó
//
// El segundo es la cobertura de la ENTRADA: `sf coverage` mide qué fracción del
// código está anclada a una spec viva; esto mide qué fracción del material del
// cliente está anclada a un requisito.
//
// Nota de diseño: `sources.json` no sabe ni le importa si la fuente es un mail
// del cliente o un hilo de HN. Por eso el mismo modelo sirve al fundador con una
// idea y a la consultora con un cliente — la decisión de posicionamiento
// (2026-08-03: los dos) no le cambia nada.
// ----------------------------------------------------------------------------

type sourcesFile struct {
	SchemaVersion string       `json:"schema_version"`
	Sources       []sourceItem `json:"sources"`
}

// sourceItem es UNA entrada de material ingerido.
type sourceItem struct {
	ID   string `json:"id"`   // S1, S2, … (forma S#, como R#/E#/D#)
	Kind string `json:"kind"` // email | transcript | interview | document | ticket | observation | other
	// Ref es una ruta relativa al proyecto o una URL. Si es ruta, tiene que
	// existir: una fuente que apunta a un archivo que no está es indistinguible
	// de una fuente inventada, que es justo lo que este artefacto viene a cazar.
	Ref      string `json:"ref"`
	Captured string `json:"captured,omitempty"` // YYYY-MM-DD
	Note     string `json:"note,omitempty"`
}

var sourceIDRe = regexp.MustCompile(`^S\d+$`)

// sourceKinds: el vocabulario de procedencia. `other` existe para que nadie
// tenga que mentir con un kind que no aplica — pero es una salida, no un default.
var sourceKinds = map[string]bool{
	"email": true, "transcript": true, "interview": true,
	"document": true, "ticket": true, "observation": true, "other": true,
}

// capturedRe: YYYY-MM-DD. No validamos que la fecha exista de verdad (31 de
// febrero pasa); alcanza con fijar la forma para que ordene lexicográficamente.
var capturedRe = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)

//go:embed templates/sources.tmpl.md
var sourcesTemplate string

var sourcesTmpl = template.Must(
	template.New("sources").
		Funcs(template.FuncMap{"list": joinOrDash, "orDash": orDash}).
		Parse(sourcesTemplate),
)

// sourcesPath: nivel PROYECTO, junto a constitution.json. El material del
// cliente cruza features — atarlo a una sola sería perder justamente la
// consulta que importa (¿qué features salieron de este mail?).
func sourcesPath(projectDir string) string {
	return filepath.Join(projectDir, "specforge", "sources")
}

func runSources(args []string) int {
	if len(args) == 0 {
		sourcesUsage()
		return 2
	}
	action, rest := args[0], args[1:]

	projectDir := "."
	toStdout := false
	asJSON := false
	for _, a := range rest {
		switch {
		case a == "--stdout":
			toStdout = true
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf sources: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	switch action {
	case "render":
		return renderSources(projectDir, toStdout)
	case "validate":
		return validateSources(projectDir)
	case "coverage":
		return sourcesCoverage(projectDir, asJSON)
	default:
		fmt.Fprintf(os.Stderr, "sf sources: unknown action %q (use render|validate|coverage)\n", action)
		return 2
	}
}

func sourcesUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf sources <render|validate|coverage> [--stdout] [--json] [project_dir]")
}

// readSourcesFile lee sources.json. El segundo retorno distingue "no existe"
// (que es legítimo: un proyecto puede no declarar fuentes) de "existe y está
// roto".
func readSourcesFile(projectDir string) (sourcesFile, bool, error) {
	data, err := os.ReadFile(sourcesPath(projectDir) + ".json")
	if err != nil {
		return sourcesFile{}, false, nil
	}
	var sf sourcesFile
	if err := json.Unmarshal(data, &sf); err != nil {
		return sourcesFile{}, true, err
	}
	return sf, true, nil
}

func renderSources(projectDir string, toStdout bool) int {
	sf, found, err := readSourcesFile(projectDir)
	if !found {
		fmt.Fprintf(os.Stderr, "sf sources: no sources.json under %s\n", projectDir)
		return 4
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf sources: invalid JSON (%v)\n", err)
		return 2
	}
	var buf strings.Builder
	if err := sourcesTmpl.Execute(&buf, sf); err != nil {
		fmt.Fprintf(os.Stderr, "sf sources: render failed (%v)\n", err)
		return 1
	}
	if toStdout {
		fmt.Print(buf.String())
		return 0
	}
	mdPath := sourcesPath(projectDir) + ".md"
	if err := os.WriteFile(mdPath, []byte(buf.String()), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf sources: cannot write %s (%v)\n", mdPath, err)
		return 1
	}
	fmt.Printf("rendered %s\n", filepath.Join("specforge", "sources.md"))
	return 0
}

func validateSources(projectDir string) int {
	sf, found, err := readSourcesFile(projectDir)
	if !found {
		fmt.Fprintf(os.Stderr, "sf sources: no sources.json under %s\n", projectDir)
		return 4
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf sources: invalid JSON (%v)\n", err)
		return 2
	}
	var rep report
	checkSourcesIn(sf, projectDir, &rep)

	fmt.Printf("Validated sources: %d source(s).\n", len(sf.Sources))
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
	fmt.Printf("\nOK: sources.json valid (%d warning(s)).\n", len(rep.warnings))
	return 0
}

// checkSources es el validador que consume `sf save` (interfaz `artifact`). No
// puede chequear que los `ref` de ruta existan porque no conoce el project dir;
// eso lo hace checkSourcesIn.
func checkSources(sf sourcesFile, rep *report) { checkSourcesIn(sf, "", rep) }

// checkSourcesIn valida el artefacto. Con projectDir != "" además verifica que
// cada `ref` de ruta apunte a un archivo real.
func checkSourcesIn(sf sourcesFile, projectDir string, rep *report) {
	if len(sf.Sources) == 0 {
		rep.warnf("no sources declared — every requirement will look model-invented")
	}
	seenID := map[string]bool{}
	seenRef := map[string]string{}
	for i, s := range sf.Sources {
		label := fmt.Sprintf("source #%d", i+1)

		switch {
		case s.ID == "":
			rep.errorf("%s: id is required (form S1, S2, …)", label)
		case !sourceIDRe.MatchString(s.ID):
			rep.errorf("%s: id %q must have the form S<n>", label, s.ID)
		case seenID[s.ID]:
			rep.errorf("%s: duplicate source id %s", label, s.ID)
		default:
			seenID[s.ID] = true
			label = s.ID
		}

		if !sourceKinds[s.Kind] {
			rep.errorf("%s: kind %q is not one of email|transcript|interview|document|ticket|observation|other",
				label, s.Kind)
		}

		ref := strings.TrimSpace(s.Ref)
		if ref == "" {
			rep.errorf("%s: ref is required — a source nobody can open is not a source", label)
		} else if prev, dup := seenRef[ref]; dup {
			// Dos ids para el mismo material parten la trazabilidad en dos: la
			// mitad de los requisitos apunta a uno y la otra mitad al otro, y
			// ninguna consulta los ve juntos.
			rep.warnf("%s: ref %q is already declared as %s — one material, one id", label, ref, prev)
		} else {
			seenRef[ref] = label
		}

		// Una ruta que no existe es indistinguible de una fuente inventada.
		if projectDir != "" && ref != "" && !isSourceURL(ref) {
			if !fileExists(filepath.Join(projectDir, filepath.FromSlash(ref))) {
				rep.errorf("%s: ref %q does not exist — a source pointing at a missing file "+
					"cannot be told apart from a fabricated one", label, ref)
			}
		}

		if s.Captured != "" && !capturedRe.MatchString(s.Captured) {
			rep.errorf("%s: captured %q must be YYYY-MM-DD", label, s.Captured)
		}
	}
}

// isSourceURL: ¿el ref es una URL y no una ruta? Se pide esquema Y host para no
// tomar por URL a un `C:\algo` de Windows ni a un `mailto:` suelto.
func isSourceURL(ref string) bool {
	u, err := url.Parse(ref)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// sourceIDSet devuelve los ids declarados. Lectura QUIETA: sin sources.json el
// set queda vacío, que es lo que hace que `require_source: false` no moleste.
func sourceIDSet(projectDir string) map[string]bool {
	out := map[string]bool{}
	sf, found, err := readSourcesFile(projectDir)
	if !found || err != nil {
		return out
	}
	for _, s := range sf.Sources {
		if s.ID != "" {
			out[s.ID] = true
		}
	}
	return out
}

// ----------------------------------------------------------------------------
// `sf sources coverage` (R6) — los dos chequeos que el ref habilita.
// ----------------------------------------------------------------------------

type sourcesCoverageReport struct {
	// Unsourced: requisitos que no citan ninguna fuente. Cada uno es un
	// candidato a "lo inventó el modelo".
	Unsourced []unsourcedRequirement `json:"unsourced_requirements"`
	// Unused: fuentes que ningún requisito referencia. Material que se ingirió
	// y no llegó a la spec — o se leyó mal, o sobraba.
	Unused []sourceItem `json:"unused_sources"`
	// Totales para la lectura rápida.
	Requirements int `json:"requirements"`
	Sources      int `json:"sources"`
}

type unsourcedRequirement struct {
	Feature     string `json:"feature"`
	Requirement string `json:"requirement"`
	// Legacy trae la prosa del campo `source` cuando había algo escrito pero no
	// era un ref. No es lo mismo "nadie declaró de dónde salió" que "se declaró
	// en prosa y todavía no se convirtió en ref".
	Legacy string `json:"legacy_source,omitempty"`
}

func sourcesCoverage(projectDir string, asJSON bool) int {
	rep := buildSourcesCoverage(projectDir)

	if asJSON {
		out, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "sf sources: marshal failed (%v)\n", err)
			return 1
		}
		fmt.Println(string(out))
		return 0
	}

	fmt.Printf("source coverage: %d requirement(s), %d source(s)\n\n", rep.Requirements, rep.Sources)

	fmt.Printf("requirements with no source: %d\n", len(rep.Unsourced))
	for _, u := range rep.Unsourced {
		if u.Legacy != "" {
			fmt.Printf("  %s / %s — free-prose source, not a ref: %q\n", u.Feature, u.Requirement, u.Legacy)
			continue
		}
		fmt.Printf("  %s / %s\n", u.Feature, u.Requirement)
	}
	if len(rep.Unsourced) > 0 {
		fmt.Println("  each one either cites where it came from, or was invented by the model")
	}

	fmt.Printf("\nsources no requirement uses: %d\n", len(rep.Unused))
	for _, s := range rep.Unused {
		fmt.Printf("  %s (%s) %s\n", s.ID, s.Kind, s.Ref)
	}
	if len(rep.Unused) > 0 {
		fmt.Println("  material that was ingested and never made it into the spec")
	}
	return 0
}

// buildSourcesCoverage cruza los requisitos de TODAS las features contra
// sources.json. Es a nivel proyecto porque el material cruza features.
func buildSourcesCoverage(projectDir string) sourcesCoverageReport {
	rep := sourcesCoverageReport{
		Unsourced: []unsourcedRequirement{},
		Unused:    []sourceItem{},
	}

	sf, found, err := readSourcesFile(projectDir)
	if found && err == nil {
		rep.Sources = len(sf.Sources)
	}

	used := map[string]bool{}
	for _, fr := range allFeatureRequirements(projectDir) {
		for _, r := range fr.file.Requirements {
			rep.Requirements++
			if len(r.Source.Refs) == 0 {
				rep.Unsourced = append(rep.Unsourced, unsourcedRequirement{
					Feature: fr.feature, Requirement: r.ID, Legacy: r.Source.Prose,
				})
				continue
			}
			for _, ref := range r.Source.Refs {
				used[ref] = true
			}
		}
	}

	for _, s := range sf.Sources {
		if !used[s.ID] {
			rep.Unused = append(rep.Unused, s)
		}
	}
	return rep
}

// featureRequirements empareja una feature con sus requisitos parseados.
type featureRequirements struct {
	feature string
	file    requirementsFile
}

// allFeatureRequirements recorre features/ y archive/ con la MISMA regla de
// deduplicación que loadTraces (DL-5 F1): `sf feature archive` copia sin borrar,
// así que sin dedup una feature archivada contaría sus requisitos dos veces y el
// reporte diría el doble de lo que hay.
func allFeatureRequirements(projectDir string) []featureRequirements {
	specforge := filepath.Join(projectDir, "specforge")
	var out []featureRequirements
	seen := map[string]bool{}

	for _, base := range []string{"features", "archive"} {
		matches, _ := filepath.Glob(filepath.Join(specforge, base, "*", "requirements.json"))
		sort.Strings(matches)
		for _, m := range matches {
			dir := filepath.Base(filepath.Dir(m))
			name := featureNameFromDir(dir, base == "archive")
			if seen[name] {
				continue
			}
			seen[name] = true
			data, err := os.ReadFile(m)
			if err != nil {
				continue
			}
			var rf requirementsFile
			if json.Unmarshal(data, &rf) != nil {
				continue
			}
			out = append(out, featureRequirements{feature: name, file: rf})
		}
	}
	return out
}
