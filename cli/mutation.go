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
// `sf mutation` — mutation testing (RM-C6, F8). OPT-IN.
//
// QUÉ PRUEBA QUE C5 NO. El testigo RED demuestra que un test PUDO fallar alguna
// vez. La mutación demuestra que el suite DETECTA DEFECTOS AHORA: se inyectan
// fallas en el código (invertir un operador, negar una condición, mover un
// límite) y se exige que algún test se ponga rojo. Si el suite sigue verde con
// el código roto, el suite no verifica — sólo acompaña.
//
// Y VUELVE SECUNDARIO EL DEBATE SOBRE TDD. Todo C5 es un rodeo para llegar a una
// propiedad: que los tests detecten defectos reales. TDD es una práctica que
// TIENDE a producirla; el testigo RED es una prueba DÉBIL de ella. La mutación
// la mide directamente. Con un score alto no importa si hubo TDD ni en qué orden
// se escribió cada cosa. Eso no debilita "no imponer prácticas, verificar
// propiedades" — lo confirma, con mejor instrumento.
//
// CAMINO A, Y NO SE NEGOCIA. Se configura un comando y se consume su EXIT CODE.
// El Camino B —parsear y normalizar scores entre herramientas, atribuir mutantes
// a requisitos— salta el costo a ALTO y nos ata a la versión y al formato de
// salida de cada herramienta. El umbral es config de la herramienta, no nuestra.
// Es además lo único honesto frente a los MUTANTES EQUIVALENTES: algunos son
// semánticamente idénticos al original y ningún test puede matarlos nunca, así
// que el score es imperfecto de forma permanente y exigir 100% sería exigir lo
// imposible. Quien decide dónde poner la vara es la herramienta.
//
// EL ALCANCE POR TRACE — la ventaja que ninguna herramienta genérica tiene.
// `trace.json` sabe qué símbolos ancló cada requisito, así que podemos pasarle al
// mutador SÓLO los archivos de la feature en curso en vez del repo entero. Ataca
// la única objeción seria (el tiempo de corrida) SIN entrar en Camino B: se
// sigue consumiendo nada más que el exit code.
// ----------------------------------------------------------------------------

// mutationScope devuelve los archivos de producción anclados por la feature,
// deduplicados y ordenados.
//
// Sale de los anclas `code` del trace, que es el código que la feature declara
// como suyo. Ordenado para que el comando sea reproducible: dos corridas con el
// mismo estado tienen que producir la misma línea, o el mismo proyecto daría
// resultados distintos sin que nada haya cambiado.
func mutationScope(projectDir, feature string) []string {
	tf := loadTraceFile(projectDir, feature)
	set := map[string]bool{}
	for _, req := range tf.Requirements {
		for _, a := range req.Code {
			path, _ := splitAnchor(a)
			rel := filepath.ToSlash(path)
			// Los tests no se mutan: mutar un test y ver que "falla" no dice
			// nada sobre si el suite detecta defectos del producto. Por contrato
			// los anclas `code` son de producción, pero filtrarlo cuesta una
			// línea y evita que un trace mal escrito arruine la corrida.
			if rel == "" || isTestFile(rel) {
				continue
			}
			set[rel] = true
		}
	}
	out := make([]string, 0, len(set))
	for f := range set {
		out = append(out, f)
	}
	sort.Strings(out)
	return out
}

// mutationCmd lee `build.mutation_cmd`. Lectura QUIETA, como el resto.
func mutationCmd(projectDir string) string {
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "constitution.json"))
	if err != nil {
		return ""
	}
	var c constitutionFile
	if json.Unmarshal(data, &c) != nil || c.Build == nil {
		return ""
	}
	return strings.TrimSpace(c.Build.MutationCmd)
}

// requireMutationEnabled: default `false`, y ORTOGONAL a require_arch. Un
// proyecto puede querer conformidad de arquitectura (segundos) sin pagar
// mutación (minutos); atarlos en un solo `lane: strict` era justo lo que DEC-3
// descartó.
func requireMutationEnabled(projectDir string) bool {
	return verificationOpts(projectDir).RequireMutation
}

// shellQuote entrecomilla una ruta para que sobreviva al shell.
//
// Hace falta acá y no en `runTest` porque acá se interpola una LISTA: un solo
// archivo con espacio en el nombre partiría el comando en dos argumentos y el
// mutador correría sobre rutas que no existen — verde por la razón equivocada,
// que es el peor resultado posible en un gate.
func shellQuote(path string) string {
	if !strings.ContainsAny(path, " \t\"'$`\\") {
		return path
	}
	// `cmd /c` no entiende comillas simples; `sh -c` sí y son las seguras.
	if isWindowsShell() {
		return `"` + strings.ReplaceAll(path, `"`, `""`) + `"`
	}
	return `'` + strings.ReplaceAll(path, `'`, `'\''`) + `'`
}

func isWindowsShell() bool {
	sh, _ := shellArgs()
	return sh == "cmd"
}

// runMutationTool shellea el mutador y consume el EXIT CODE.
//
// El timeout es de 30 minutos y no de los 300 s del resto: la mutación corre el
// suite una vez por mutante, así que es una operación de minutos por diseño.
// Dejarle el timeout corto la haría fallar siempre por "no se pudo correr", que
// se lee como un problema de cableado y no como lo que sería — una herramienta
// haciendo exactamente su trabajo.
func runMutationTool(projectDir, template string, files []string) (bool, string) {
	cmd := template
	if strings.Contains(template, "{files}") {
		quoted := make([]string, 0, len(files))
		for _, f := range files {
			quoted = append(quoted, shellQuote(f))
		}
		cmd = strings.ReplaceAll(template, "{files}", strings.Join(quoted, " "))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	sh, flag := shellArgs()
	c := exec.CommandContext(ctx, sh, flag, cmd)
	c.Dir = projectDir
	out, err := c.CombinedOutput()
	if err == nil {
		return true, ""
	}
	// Igual que runTest y runArchTool: "corrió y el suite no mató mutantes
	// suficientes" no es lo mismo que "no se pudo ni lanzar". Confundirlos haría
	// que una herramienta ausente se reportara como un suite que no verifica.
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false, strings.TrimSpace(string(out))
	}
	return false, fmt.Sprintf("could not run mutation_cmd: %v", err)
}

// mutationGateReasons aplica RM-C6 en el gate del veredicto. Vacío ⇒ pasa.
func mutationGateReasons(projectDir, feature string) []string {
	tmpl := mutationCmd(projectDir)
	if tmpl == "" {
		// Encendido sin comando RECHAZA, igual que C7 y que R12: aprobar sería
		// dejar al proyecto creyendo que tiene una garantía que nunca se evaluó.
		return []string{
			"require_mutation is on but build.mutation_cmd is not configured — nothing can check that " +
				"the suite detects defects. Set build.mutation_cmd or turn the flag off",
		}
	}

	files := mutationScope(projectDir, feature)
	// Sin alcance no hay nada que mutar, y un mutador sobre cero archivos sale
	// verde. Ese verde no significa "el suite verifica", significa "no se probó
	// nada" — exactamente la confusión que el flag existe para impedir.
	if strings.Contains(tmpl, "{files}") && len(files) == 0 {
		return []string{
			"require_mutation is on and mutation_cmd is scoped by {files}, but no requirement anchors " +
				"any production code — there is nothing to mutate, so nothing would be verified",
		}
	}

	ok, detail := runMutationTool(projectDir, tmpl, files)
	if ok {
		return nil
	}
	msg := "mutation testing FAILED — the suite does not detect the injected defects (its threshold was not met)"
	if detail != "" {
		msg += ": " + detail
	}
	return []string{msg}
}

// ----------------------------------------------------------------------------
// El comando.
// ----------------------------------------------------------------------------

func runMutation(args []string) int {
	if len(args) == 0 {
		mutationUsage()
		return 2
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "scope":
		return runMutationScope(rest)
	case "run":
		return runMutationRun(rest)
	default:
		fmt.Fprintf(os.Stderr, "sf mutation: unknown subcommand %q\n", sub)
		mutationUsage()
		return 2
	}
}

func mutationUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf mutation scope --feature=NAME [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf mutation run --feature=NAME [project_dir]")
}

func parseMutationArgs(args []string) (projectDir, feature string, code int) {
	projectDir = "."
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf mutation: unknown flag %q\n", a)
			return "", "", 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf mutation: --feature=NAME is required")
		return "", "", 2
	}
	return projectDir, feature, 0
}

// runMutationScope muestra qué archivos se le van a pasar al mutador. Sin esto,
// la primera experiencia con C6 sería esperar veinte minutos para descubrir que
// el alcance no era el esperado.
func runMutationScope(args []string) int {
	projectDir, feature, code := parseMutationArgs(args)
	if code != 0 {
		return code
	}
	files := mutationScope(projectDir, feature)
	if len(files) == 0 {
		fmt.Printf("no production code anchored by %s — nothing to mutate.\n", feature)
		return 0
	}
	fmt.Printf("mutation scope — %s (%d file(s)):\n", feature, len(files))
	for _, f := range files {
		fmt.Printf("  %s\n", f)
	}
	return 0
}

func runMutationRun(args []string) int {
	projectDir, feature, code := parseMutationArgs(args)
	if code != 0 {
		return code
	}
	if mutationCmd(projectDir) == "" {
		// Degradación elegante, igual que `sf arch check`: correrlo a mano es
		// una pregunta legítima, no un gate que alguien esté por cruzar.
		fmt.Println("build.mutation_cmd is not configured — mutation testing skipped.")
		return 0
	}
	reasons := mutationGateReasons(projectDir, feature)
	if len(reasons) > 0 {
		for _, r := range reasons {
			fmt.Fprintf(os.Stderr, "%s\n", r)
		}
		return 1
	}
	fmt.Printf("mutation testing OK — %s's suite meets its mutation threshold.\n", feature)
	return 0
}
