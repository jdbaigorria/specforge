package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// `sf arch` — conformidad de arquitectura (RM-C7, F7). OPT-IN.
//
// EL HUECO QUE TAPA. `design.json` declara el grafo de arquitectura en
// `component.DependsOn`. Ese grafo pasa por un gate humano y queda sellado. Y
// después no lo mira nadie. Una wave puede implementar todos los requisitos,
// nombrar todos los tests, pasar la causalidad, salir verde y sellarse HABIENDO
// VIOLADO POR COMPLETO EL DISEÑO APROBADO. El sello certifica trazabilidad y
// verificación; no certifica que el código respete su propia estructura.
//
// Es el mismo patrón que `requirement.Source` antes de RM-C3: un campo
// estructurado, ya aprobado, que muere sin que nada lo compute.
//
// NO ESCRIBIMOS UN ANALIZADOR. Se delega a la herramienta del stack
// (`go-arch-lint`/`depguard` en Go, `import-linter` en Python,
// `dependency-cruiser` en JS), igual que `build.test_cmd` delega el runner. Lo
// que aporta SpecForge es lo que ninguna de ellas puede: **generar la regla
// desde el diseño aprobado**. La regla no la escribe el usuario en un archivo
// aparte que se desincroniza — sale de `design.json`, así que si el diseño
// cambia por `sf-amend`, la regla cambia con él y ambos quedan bajo el sello.
//
// DE DÓNDE SALE EL MAPEO COMPONENTE → ARCHIVOS. `component` no tiene campo de
// ruta, y agregárselo habría reintroducido justo el archivo paralelo que se
// desincroniza. El dato ya existe: `task` tiene `component_refs` Y
// `files_touched`, los dos aprobados en el gate de tasks. El mapeo se DERIVA de
// ahí. Cero esquema nuevo, y la fuente sigue siendo un artefacto sellado.
// ----------------------------------------------------------------------------

// archRules es el contrato NEUTRAL que `sf` le pasa a la herramienta: quién es
// cada componente (sus archivos) y a quién tiene permitido depender.
//
// Es neutral a propósito. Emitir YAML de go-arch-lint, INI de import-linter y JS
// de dependency-cruiser metería a SpecForge en el negocio de mantener tres
// formatos de terceros que cambian sin avisar. Acá se emite el HECHO —el grafo
// permitido y de quién es cada archivo— y el `arch_cmd` del proyecto lo adapta a
// su herramienta. Es la misma división que con `test_cmd`: nosotros decimos qué
// hay que verificar, el stack dice cómo.
type archRules struct {
	SchemaVersion string          `json:"schema_version"`
	Feature       string          `json:"feature"`
	Components    []archComponent `json:"components"`
}

type archComponent struct {
	ID    string   `json:"id"`
	Name  string   `json:"name,omitempty"`
	Files []string `json:"files"`
	// MayDependOn es `component.DependsOn` tal cual lo aprobó el gate de diseño:
	// las aristas DIRECTAS, sin clausura transitiva. Calcular la clausura acá
	// sería decidir por el usuario que las dependencias son transitivas, y esa
	// es una propiedad de la arquitectura que él declaró, no que nosotros
	// inferimos. Si la quiere transitiva, su herramienta la computa.
	MayDependOn []string `json:"may_depend_on"`
}

// archRulesPath: el archivo vive en `.state/`, que es estado de máquina
// protegido por el hook. Se REGENERA en cada corrida y nunca se edita a mano —
// su fuente de verdad es el diseño sellado, así que una edición manual sería una
// regla que ya no representa a ningún diseño aprobado.
func archRulesPath(projectDir, feature, format string) string {
	return filepath.Join(projectDir, "specforge", ".state", archRulesFilename(feature, format))
}

// archFormat lee `build.arch_format`. Ausente ⇒ `json`, el contrato neutral —
// que es lo que ya emitíamos, así que un proyecto existente no cambia de
// comportamiento por actualizar el binario.
func archFormat(projectDir string) string {
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "constitution.json"))
	if err != nil {
		return "json"
	}
	var c constitutionFile
	if json.Unmarshal(data, &c) != nil || c.Build == nil {
		return "json"
	}
	if f := strings.TrimSpace(c.Build.ArchFormat); f != "" {
		return f
	}
	return "json"
}

// componentFiles deriva el mapeo componente → archivos desde tasks.json.
//
// Devuelve también los CONFLICTOS: un archivo reclamado por tasks de dos
// componentes distintos. No se resuelve por nuestra cuenta (ni "gana el
// primero" ni "pertenece a los dos") porque las dos salidas son inventar un
// mapeo que nadie declaró, y con un mapeo inventado la conformidad que se
// verifique después no significa nada. §3.8 lo dice: debe fallar pidiendo el
// mapeo, nunca inventarlo.
func componentFiles(projectDir, feature string) (map[string][]string, []string) {
	files := map[string]map[string]bool{} // componente → set de archivos
	owner := map[string]string{}          // archivo → componente que ya lo reclamó
	var conflicts []string

	for _, t := range tasksOf(projectDir, feature) {
		for _, c := range t.ComponentRefs {
			for _, f := range t.FilesTouched {
				rel := filepath.ToSlash(f)
				// Los tests quedan afuera, igual que en `sf coverage`. La
				// conformidad es sobre la estructura del PRODUCTO: un test que
				// importa a través de una frontera de componentes es normal
				// (arma un escenario de punta a punta), así que incluirlos
				// generaría violaciones falsas — y un chequeo que grita en falso
				// es uno que se termina apagando.
				if isTestFile(rel) {
					continue
				}
				if prev, taken := owner[rel]; taken && prev != c {
					conflicts = append(conflicts, fmt.Sprintf(
						"%s is claimed by both %s and %s — a file belongs to one component; "+
							"split the task or correct component_refs in tasks.json", rel, prev, c))
					continue
				}
				owner[rel] = c
				if files[c] == nil {
					files[c] = map[string]bool{}
				}
				files[c][rel] = true
			}
		}
	}

	out := make(map[string][]string, len(files))
	for c, set := range files {
		list := make([]string, 0, len(set))
		for f := range set {
			list = append(list, f)
		}
		sort.Strings(list) // salida estable: el archivo de reglas se diffea
		out[c] = list
	}
	sort.Strings(conflicts)
	return out, conflicts
}

// tasksOf y designComponentsOf: lecturas QUIETAS de los dos artefactos, con el
// mismo resolvedor (`findArtifact`) que usa `requirementsOf` — así una feature
// `done`, cuyos artefactos viven bajo `archive/<fecha>-<nombre>/`, se resuelve
// igual que una viva.
func tasksOf(projectDir, feature string) []task {
	path := findArtifact(filepath.Join(projectDir, "specforge"), feature, "tasks.json")
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var tf tasksFile
	if json.Unmarshal(data, &tf) != nil {
		return nil
	}
	return tf.Tasks
}

func designComponentsOf(projectDir, feature string) []component {
	path := findArtifact(filepath.Join(projectDir, "specforge"), feature, "design.json")
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var df designFile
	if json.Unmarshal(data, &df) != nil {
		return nil
	}
	return df.Components
}

// buildArchRules arma las reglas desde design.json + el mapeo. Devuelve las
// razones por las que NO se pudo: son las que §3.8 exige que fallen pidiendo el
// mapeo en vez de inventarlo.
func buildArchRules(projectDir, feature string) (archRules, []string) {
	rules := archRules{SchemaVersion: "1.0", Feature: feature}

	comps := designComponentsOf(projectDir, feature)
	if len(comps) == 0 {
		return rules, []string{
			"no components declared in design.json — architecture conformance has nothing to check",
		}
	}

	mapping, conflicts := componentFiles(projectDir, feature)
	if len(conflicts) > 0 {
		return rules, conflicts
	}

	declared := map[string]bool{}
	for _, c := range comps {
		declared[c.ID] = true
	}

	var reasons []string
	for _, c := range comps {
		f := mapping[c.ID]
		if len(f) == 0 {
			// Un componente sin correlato físico: exactamente el caso que §3.8
			// manda rechazar. Generar reglas igual dejaría un componente sin
			// archivos, que ninguna herramienta puede violar — o sea, un hueco
			// silencioso justo donde se prometió una garantía.
			reasons = append(reasons, fmt.Sprintf(
				"%s (%s) maps to no production files — no task in tasks.json lists it in component_refs "+
					"with non-test files_touched, so its dependencies cannot be checked", c.ID, c.Name))
			continue
		}
		// Una arista hacia un componente que el diseño no declara: el grafo
		// aprobado está roto y generar la regla la daría por buena.
		for _, dep := range c.DependsOn {
			if !declared[dep] {
				reasons = append(reasons, fmt.Sprintf(
					"%s depends_on %q, which no component declares", c.ID, dep))
			}
		}
		// `[]` y no `null`: este archivo lo parsea una herramienta de terceros, y
		// un campo que a veces es lista y a veces null obliga a todo consumidor a
		// manejar los dos casos. Mismo criterio que `unanchored` en
		// `sf coverage --json`.
		deps := []string{}
		deps = append(deps, c.DependsOn...)
		sort.Strings(deps)
		rules.Components = append(rules.Components, archComponent{
			ID: c.ID, Name: c.Name, Files: f, MayDependOn: deps,
		})
	}
	sort.Slice(rules.Components, func(i, j int) bool { return rules.Components[i].ID < rules.Components[j].ID })
	return rules, reasons
}

// writeArchRules genera el archivo en el formato pedido y devuelve su ruta.
func writeArchRules(projectDir, feature, format string) (string, []string) {
	if !archFormats[format] {
		return "", []string{fmt.Sprintf("unknown arch format %q — known: %s",
			format, strings.Join(sortedKeys(archFormats), ", "))}
	}
	rules, reasons := buildArchRules(projectDir, feature)
	if len(reasons) > 0 {
		return "", reasons
	}

	var body []byte
	switch format {
	case "go-arch-lint":
		yaml, why := archLintYAML(rules)
		if len(why) > 0 {
			return "", why
		}
		body = []byte(yaml)
	default:
		data, err := json.MarshalIndent(rules, "", "  ")
		if err != nil {
			return "", []string{fmt.Sprintf("cannot encode arch rules: %v", err)}
		}
		body = append(data, '\n')
	}

	path := archRulesPath(projectDir, feature, format)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", []string{fmt.Sprintf("cannot create .state directory: %v", err)}
	}
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", []string{fmt.Sprintf("cannot write arch rules: %v", err)}
	}
	return path, nil
}

// archCmd lee `build.arch_cmd`. Lectura QUIETA, igual que el resto de la config.
func archCmd(projectDir string) string {
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "constitution.json"))
	if err != nil {
		return ""
	}
	var c constitutionFile
	if json.Unmarshal(data, &c) != nil || c.Build == nil {
		return ""
	}
	return strings.TrimSpace(c.Build.ArchCmd)
}

// requireArchEnabled: ¿el proyecto pidió la garantía? Default `false` — igual
// que los otros opt-in, encenderlo el día uno rompería todo proyecto existente.
func requireArchEnabled(projectDir string) bool {
	return verificationOpts(projectDir).RequireArch
}

// runArchTool shellea la herramienta con `{config}` sustituido y CONSUME EL EXIT
// CODE. No se parsea la salida: el umbral y el formato del reporte son de la
// herramienta, y parsearlos nos ataría a su versión. Mismo criterio que C6.
func runArchTool(projectDir, template, configPath string) (bool, string) {
	cmd := strings.ReplaceAll(template, "{config}", configPath)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	sh, flag := shellArgs()
	c := exec.CommandContext(ctx, sh, flag, cmd)
	c.Dir = projectDir
	out, err := c.CombinedOutput()
	if err == nil {
		return true, ""
	}
	// Igual que runTest: distinguir "corrió y encontró una violación" de "no se
	// pudo ni lanzar". Confundirlos haría que una herramienta mal instalada se
	// reportara como código que viola su diseño.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false, strings.TrimSpace(string(out))
	}
	return false, fmt.Sprintf("could not run arch_cmd: %v", err)
}

// archGateReasons aplica RM-C7 en el gate del veredicto. Vacío ⇒ conforme.
func archGateReasons(projectDir, feature string) []string {
	tmpl := archCmd(projectDir)
	if tmpl == "" {
		// El flag encendido sin comando configurado RECHAZA, no degrada. Aprobar
		// sería peor que bloquear: el proyecto creería tener una garantía de
		// conformidad que nunca se evaluó. Mismo criterio que R12 con
		// build.report — y la degradación elegante aplica al flag APAGADO, que
		// ni llega hasta acá.
		return []string{
			"require_arch is on but build.arch_cmd is not configured — nothing can check the " +
				"approved dependency graph. Set build.arch_cmd or turn the flag off",
		}
	}
	path, reasons := writeArchRules(projectDir, feature, archFormat(projectDir))
	if len(reasons) > 0 {
		return reasons
	}
	ok, detail := runArchTool(projectDir, tmpl, path)
	if ok {
		return nil
	}
	msg := "architecture conformance FAILED — the built code violates the approved design graph"
	if detail != "" {
		msg += ": " + detail
	}
	return []string{msg}
}

// ----------------------------------------------------------------------------
// El comando.
// ----------------------------------------------------------------------------

func runArch(args []string) int {
	if len(args) == 0 {
		archUsage()
		return 2
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "rules":
		return runArchRules(rest)
	case "check":
		return runArchCheck(rest)
	default:
		fmt.Fprintf(os.Stderr, "sf arch: unknown subcommand %q\n", sub)
		archUsage()
		return 2
	}
}

func archUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf arch rules --feature=NAME [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf arch check --feature=NAME [project_dir]")
}

// parseArchArgs comparten las dos subcomandos.
func parseArchArgs(args []string) (projectDir, feature string, code int) {
	projectDir = "."
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf arch: unknown flag %q\n", a)
			return "", "", 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf arch: --feature=NAME is required")
		return "", "", 2
	}
	return projectDir, feature, 0
}

// runArchRules genera el archivo de reglas y lo reporta. Sirve para inspeccionar
// qué se le va a pasar a la herramienta ANTES de cablearla — sin esto, la
// primera experiencia con C7 sería un exit code sin explicación.
func runArchRules(args []string) int {
	format := ""
	var rest []string
	for _, a := range args {
		if strings.HasPrefix(a, "--format=") {
			format = strings.TrimPrefix(a, "--format=")
			continue
		}
		rest = append(rest, a)
	}
	projectDir, feature, code := parseArchArgs(rest)
	if code != 0 {
		return code
	}
	if format == "" {
		format = archFormat(projectDir)
	}
	path, reasons := writeArchRules(projectDir, feature, format)
	if len(reasons) > 0 {
		fmt.Fprintf(os.Stderr, "sf arch rules: cannot derive the component→files mapping (%d issue(s)):\n", len(reasons))
		for _, r := range reasons {
			fmt.Fprintf(os.Stderr, "  - %s\n", r)
		}
		return 1
	}
	fmt.Printf("arch rules → %s\n", filepath.ToSlash(path))
	return 0
}

// runArchCheck genera las reglas y corre la herramienta.
func runArchCheck(args []string) int {
	projectDir, feature, code := parseArchArgs(args)
	if code != 0 {
		return code
	}
	if archCmd(projectDir) == "" {
		// Acá sí degrada elegante: `sf arch check` a mano sin comando
		// configurado es una pregunta legítima ("¿tengo esto cableado?"), no un
		// gate que alguien esté por cruzar. El que bloquea es el veredicto.
		fmt.Println("build.arch_cmd is not configured — architecture conformance skipped.")
		return 0
	}
	reasons := archGateReasons(projectDir, feature)
	if len(reasons) > 0 {
		for _, r := range reasons {
			fmt.Fprintf(os.Stderr, "%s\n", r)
		}
		return 1
	}
	fmt.Printf("architecture conformance OK — %s respects its approved dependency graph.\n", feature)
	return 0
}
