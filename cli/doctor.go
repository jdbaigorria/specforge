package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// Modelo de trace.json — la matriz de trazabilidad que emite sf-check.
//
// "requirements" es un OBJETO json (mapa req-id → info), no un array. En Go eso
// es un map[string]traceReq. Los maps de Go iteran en orden ALEATORIO, así que
// cuando recorramos los requirements vamos a ordenar las claves para una salida
// estable.
// ----------------------------------------------------------------------------

type traceFile struct {
	// SchemaVersion va PRIMERO para que el JSON canónico que escribe `sf save`
	// lleve "schema_version" arriba, igual que el resto de los artefactos. El
	// resto del código (verifyTrace) lo ignora; solo round-trippea el campo.
	SchemaVersion string              `json:"schema_version,omitempty"`
	Feature       string              `json:"feature"`
	Requirements  map[string]traceReq `json:"requirements"`
}

type traceReq struct {
	Code   []string `json:"code"`
	Test   []string `json:"test"`
	Status string   `json:"status"`
}

// runDoctor es el punto de entrada de `sf doctor`. Por ahora el único chequeo es
// drift; cuando doctor crezca (integridad de estado, staleness, schema) el flag
// --drift servirá para aislarlo. Hoy `sf doctor` y `sf doctor --drift` hacen lo
// mismo.
func runDoctor(args []string) int {
	projectDir := "."
	quiet := false
	runTests := ""
	install := false
	global := false
	asJSON := false

	// Parseo manual de flags. i++ extra cuando un flag consume su valor.
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--drift":
			// reconocido; sin efecto distinto todavía (drift es el único chequeo)
		case a == "--install":
			install = true
		case a == "--global":
			global = true
		case a == "--quiet":
			quiet = true
		case a == "--json":
			asJSON = true
		case a == "--run-tests":
			if i+1 < len(args) {
				runTests = args[i+1]
				i++
			} else {
				runTests = "pytest -q {test}"
			}
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf doctor: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	if install {
		return runDoctorInstall(projectDir, global)
	}
	return runDrift(projectDir, runTests, quiet, asJSON)
}

// ----------------------------------------------------------------------------
// La ONTOLOGÍA DE DRIFT (DL-4) — cuatro categorías, no un booleano.
//
// El motor veía existencia (¿resuelve el símbolo?) y reportaba una lista plana.
// Pero "el código no está" y "el código está y hace otra cosa" son hallazgos
// distintos que piden decisiones distintas, y aplanarlos los volvía un solo
// "DRIFT" sin acción asociada.
//
// La categoría difícil es la 2. `checkAnchor` mira que el símbolo RESUELVA, no
// qué hace — así que "implementado distinto" no se puede ver estáticamente.
// Dejarla a juicio de un LLM sería repetir justo el defecto que este bloque
// viene a corregir. Tiene una definición computable:
//
//	el anchor de código resuelve Y el test que ancla ese requisito falla.
//
// Si el símbolo está y su test da rojo, el código hace algo distinto de lo que
// el requisito pide. Eso es gratis con --run-tests.
//
// La tabla de decisión, por requisito:
//
//	anchor de código │ test                    │ categoría
//	─────────────────┼─────────────────────────┼──────────────────────────
//	no resuelve      │ —                       │ 1 — no implementado
//	resuelve         │ falla                   │ 2 — implementado distinto
//	resuelve         │ no existe               │ 3 — sin verificar
//	resuelve         │ pasa                    │ (sin drift)
//
// EL LÍMITE, DICHO: sin --run-tests la categoría 2 es INDETERMINABLE, no
// ausente. El reporte lo dice con todas las letras en vez de emitir una lista
// vacía — un `[]` se lee como "no hay", y ahí es donde la falta de información
// se disfraza de limpio.
// ----------------------------------------------------------------------------

// driftItem es un hallazgo ya clasificado.
type driftItem struct {
	Feature     string `json:"feature"`
	Requirement string `json:"requirement"`
	Reason      string `json:"reason"`
}

// undetermined es lo que ocupa el lugar de una lista cuando el chequeo no se
// pudo correr. Es un OBJETO y no un array a propósito: obliga al consumidor a
// distinguir "no hay hallazgos" de "no se miró".
type undetermined struct {
	Status string `json:"status"` // siempre "undetermined"
	Reason string `json:"reason"`
}

// driftReport es la salida de `sf doctor --drift --json`.
//
// ImplementedDifferently es `any` porque lleva []driftItem o undetermined según
// se hayan corrido los tests. Es incómodo de tipar y ese es el punto: un
// consumidor no puede leerlo como lista vacía sin darse cuenta.
type driftReport struct {
	Checked                int         `json:"checked"`
	NotImplemented         []driftItem `json:"not_implemented"`
	ImplementedDifferently any         `json:"implemented_differently"`
	Unverified             []driftItem `json:"unverified"`
	OutOfSpec              []string    `json:"out_of_spec"`
}

// diverged: ¿hay drift de COMPORTAMIENTO? Sólo las categorías 1 y 2 lo son —
// el código no está, o hace otra cosa.
//
// Las categorías 3 y 4 son HUECOS, no divergencias: un requisito sin test y un
// archivo sin spec son trabajo que falta, no una contradicción entre spec y
// código. Se reportan siempre (la regla anti-omisión), pero no mueven el exit
// code, porque `sf doctor` saliendo 1 en todo repo brownfield es un comando que
// se deja de correr en una semana — y un chequeo que nadie corre no chequea nada.
func (r driftReport) diverged() int {
	n := len(r.NotImplemented)
	if items, ok := r.ImplementedDifferently.([]driftItem); ok {
		n += len(items)
	}
	return n
}

// runDrift recorre todos los trace.json del proyecto, clasifica lo que encuentra
// y lo reporta. Devuelve 1 si hay drift de comportamiento, 0 si no.
func runDrift(projectDir, runTests string, quiet, asJSON bool) int {
	specforge := filepath.Join(projectDir, "specforge")
	if info, err := os.Stat(specforge); err != nil || !info.IsDir() {
		if asJSON {
			return printDriftJSON(emptyDriftReport(runTests))
		}
		fmt.Printf("No specforge/ under %s — nothing to check.\n", projectDir)
		return 0
	}

	rep := buildDriftReport(projectDir, runTests)
	if asJSON {
		if code := printDriftJSON(rep); code != 0 {
			return code
		}
		if rep.diverged() > 0 {
			return 1
		}
		return 0
	}
	return printDriftHuman(rep, runTests, quiet)
}

// buildDriftReport arma el reporte completo: las tres categorías por requisito
// más la cuarta, que sale de la cobertura (archivos de código que ningún trace
// cita).
func buildDriftReport(projectDir, runTests string) driftReport {
	traces := loadTraces(filepath.Join(projectDir, "specforge"))

	rep := driftReport{
		Checked:        len(traces),
		NotImplemented: []driftItem{},
		Unverified:     []driftItem{},
	}
	var differently []driftItem
	for _, t := range traces {
		c := classifyFeatureDrift(projectDir, t.feature, t.path, runTests)
		rep.NotImplemented = append(rep.NotImplemented, c.notImplemented...)
		differently = append(differently, c.implementedDifferently...)
		rep.Unverified = append(rep.Unverified, c.unverified...)
	}

	// R2: sin --run-tests la categoría 2 no se pudo mirar. El objeto lo dice.
	if runTests == "" {
		rep.ImplementedDifferently = undetermined{
			Status: "undetermined",
			Reason: "requires --run-tests",
		}
	} else {
		if differently == nil {
			differently = []driftItem{}
		}
		rep.ImplementedDifferently = differently
	}

	rep.OutOfSpec = computeSpecCoverageDetail(projectDir).Unanchored
	if rep.OutOfSpec == nil {
		rep.OutOfSpec = []string{}
	}
	return rep
}

// emptyDriftReport: el reporte de un proyecto sin specforge/. Mismo esquema —
// un consumidor no debería necesitar un caso especial para "no hay nada".
func emptyDriftReport(runTests string) driftReport {
	rep := driftReport{NotImplemented: []driftItem{}, Unverified: []driftItem{}, OutOfSpec: []string{}}
	if runTests == "" {
		rep.ImplementedDifferently = undetermined{Status: "undetermined", Reason: "requires --run-tests"}
	} else {
		rep.ImplementedDifferently = []driftItem{}
	}
	return rep
}

func printDriftJSON(rep driftReport) int {
	out, err := json.MarshalIndent(rep, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf doctor: marshal failed (%v)\n", err)
		return 1
	}
	fmt.Println(string(out))
	return 0
}

// printDriftHuman rinde el reporte para una persona. Las cuatro categorías van
// SIEMPRE, incluso vacías: ver "sin verificar: 0" es información; que la sección
// no aparezca es indistinguible de que no se haya mirado.
func printDriftHuman(rep driftReport, runTests string, quiet bool) int {
	if !quiet {
		fmt.Printf("Checked %d feature(s) with trace.json.\n", rep.Checked)
	}

	printDriftSection("1 — not implemented", rep.NotImplemented)
	if items, ok := rep.ImplementedDifferently.([]driftItem); ok {
		printDriftSection("2 — implemented differently", items)
	} else {
		fmt.Println("\n2 — implemented differently: undetermined (re-run with --run-tests)")
	}
	printDriftSection("3 — unverified", rep.Unverified)

	fmt.Printf("\n4 — out of spec: %d code file(s) anchored to no trace\n", len(rep.OutOfSpec))
	for _, f := range rep.OutOfSpec {
		fmt.Printf("  %s\n", f)
	}
	if len(rep.OutOfSpec) > 0 {
		fmt.Println("  each one: adopt it (write the requirement) or exclude it " +
			"(constitution.json coverage.exclude, with a reason)")
	}

	if n := rep.diverged(); n > 0 {
		fmt.Printf("\nDRIFT: %d requirement(s) diverged from the spec.\n", n)
		if runTests == "" {
			fmt.Println("Category 2 was not evaluated — the count is a floor, not a total.")
		}
		return 1
	}
	fmt.Println("\nOK: specs and code in sync.")
	if runTests == "" {
		fmt.Println("(category 2 not evaluated: re-run with --run-tests to check behaviour, not just existence)")
	}
	return 0
}

func printDriftSection(title string, items []driftItem) {
	fmt.Printf("\n%s: %d\n", title, len(items))
	for _, it := range items {
		fmt.Printf("  %s / %s: %s\n", it.Feature, it.Requirement, it.Reason)
	}
}

// traceEntry empareja una feature con la ruta de su trace.json.
type traceEntry struct {
	feature  string
	path     string
	archived bool // true si la ruta vive bajo archive/ (la foto histórica)
}

// archiveDatePrefix matchea el prefijo `YYYY-MM-DD-` que `sf feature archive`
// le antepone al nombre de la carpeta al copiarla a archive/.
//
// Concepto Go: `MustCompile` compila la regex al cargar el paquete y panica si
// está mal escrita — es lo que se usa para regexes literales (un error acá es
// un bug del programador, no una condición de runtime que valga la pena
// manejar). `^` ancla al principio: sólo se saca un prefijo, nunca uno del medio.
var archiveDatePrefix = regexp.MustCompile(`^\d{4}-\d{2}-\d{2}-`)

// featureNameFromDir resuelve el nombre real de la feature a partir del nombre
// del directorio que contiene su trace.json.
//
// Para las copias archivadas hay que sacar el prefijo de fecha, y hay que sacar
// EXACTAMENTE UNO: una feature que de verdad se llame `2026-06-16-slugify` y se
// archive el 2026-08-04 vive en `archive/2026-08-04-2026-06-16-slugify`, y su
// nombre real es `2026-06-16-slugify`. Sacar todos los prefijos que matcheen la
// dejaría en `slugify`, que es otra feature.
func featureNameFromDir(dir string, archived bool) string {
	if !archived {
		return dir
	}
	// ReplaceAllString con la regex anclada en ^ reemplaza como mucho una vez.
	return archiveDatePrefix.ReplaceAllString(dir, "")
}

// loadTraces busca todos los specforge/{features,archive}/*/trace.json y
// devuelve UNA entrada por feature.
//
// Por qué la deduplicación (DL-5 F1): `sf feature archive` COPIA la carpeta de
// la feature a archive/ y NO borra la original. Sin dedup, una feature archivada
// aporta su trace dos veces y cada ancla divergente se reporta duplicada — una
// vez bajo su nombre real y otra bajo `<fecha>-<nombre>`, que no existe en
// features.json y por lo tanto no se puede resolver contra ningún status.
//
// features/ se recorre PRIMERO a propósito: ante duplicado gana la copia viva,
// que es la que refleja el estado actual del proyecto. La archivada es la foto.
func loadTraces(specforge string) []traceEntry {
	var out []traceEntry
	seen := map[string]bool{} // nombre de feature ya emitido

	for _, base := range []string{"features", "archive"} {
		matches, _ := filepath.Glob(filepath.Join(specforge, base, "*", "trace.json"))
		sort.Strings(matches) // salida estable
		for _, m := range matches {
			dir := filepath.Base(filepath.Dir(m))
			archived := base == "archive"
			name := featureNameFromDir(dir, archived)
			if seen[name] {
				continue // ya la vimos viva: la copia archivada no se re-chequea
			}
			seen[name] = true
			out = append(out, traceEntry{feature: name, path: m, archived: archived})
		}
	}
	return out
}

// featureDrift son los hallazgos de UNA feature, ya repartidos por categoría.
// (La 4 es a nivel proyecto, no por feature: sale de la cobertura.)
type featureDrift struct {
	notImplemented         []driftItem
	implementedDifferently []driftItem
	unverified             []driftItem
}

// classifyFeatureDrift aplica la tabla de decisión a cada requisito del trace.
func classifyFeatureDrift(projectDir, feature, tracePath, runTests string) featureDrift {
	var out featureDrift
	add := func(bucket *[]driftItem, req, reason string) {
		*bucket = append(*bucket, driftItem{Feature: feature, Requirement: req, Reason: reason})
	}

	data, err := os.ReadFile(tracePath)
	if err != nil {
		add(&out.notImplemented, "—", fmt.Sprintf("unreadable trace.json (%v)", err))
		return out
	}
	var tf traceFile
	if err := json.Unmarshal(data, &tf); err != nil {
		add(&out.notImplemented, "—", fmt.Sprintf("unreadable trace.json (%v)", err))
		return out
	}

	// Orden estable de los requirements.
	reqIDs := make([]string, 0, len(tf.Requirements))
	for id := range tf.Requirements {
		reqIDs = append(reqIDs, id)
	}
	sort.Strings(reqIDs)

	for _, req := range reqIDs {
		info := tf.Requirements[req]

		// Categoría 1: el código no está. Un requisito sin ningún anchor de
		// código es el caso extremo — `sf save trace` ya lo rechaza, así que
		// sólo aparece en traces legados o escritos a mano.
		codeResolves := len(info.Code) > 0
		if len(info.Code) == 0 {
			add(&out.notImplemented, req, "no code anchor — nothing implements this requirement")
		}
		for _, anchor := range info.Code {
			if ok, reason := checkAnchor(projectDir, anchor); !ok {
				add(&out.notImplemented, req, reason)
				codeResolves = false
			}
		}

		// Categoría 3, mitad estática: el requisito no nombra ningún test. Esto
		// se sabe sin correr nada y es un hueco real de la matriz.
		if len(info.Test) == 0 {
			add(&out.unverified, req, "names no test — nothing verifies this requirement")
			continue
		}

		// Sin --run-tests no hay nada más que mirar: el requisito tiene test
		// declarado, pero si pasa o falla es justamente lo indeterminado (R2).
		if runTests == "" {
			continue
		}

		for _, testID := range info.Test {
			ok, reason := runTest(projectDir, testID, runTests)
			switch {
			case ok:
				// nada: el requisito está implementado y verificado.
			case !isTestFailure(reason):
				// El test no se pudo ni lanzar (binario ausente, timeout). No es
				// un veredicto sobre el código: es que no se pudo verificar.
				add(&out.unverified, req, reason)
			case codeResolves:
				// R3, el caso que da sentido a la categoría: el símbolo está y su
				// test da rojo → el código hace algo distinto de lo que se pidió.
				add(&out.implementedDifferently, req, reason)
			default:
				// R3, la otra mitad: EL ANCLA MANDA. Si el símbolo no resuelve, el
				// test rojo no agrega una categoría nueva — es la misma ausencia
				// vista desde otro lado.
				add(&out.notImplemented, req, reason)
			}
		}
	}
	return out
}

// isTestFailure distingue "el test corrió y falló" de "no se pudo correr". La
// distinción viene de runTest, que ya separa el ExitError del resto; acá la
// leemos del prefijo del mensaje para no cambiarle la firma.
func isTestFailure(reason string) bool {
	return strings.HasPrefix(reason, "test failed:")
}

// checkFeatureDrift devuelve los mensajes de drift de COMPORTAMIENTO de una
// feature (categorías 1 y 2). runTests vacío = solo chequeo estático.
//
// Sigue existiendo con esta firma porque la consumen `sf status` (columna drift)
// y `sf verify` (chequeo 4), y las dos preguntan lo mismo: ¿spec y código se
// contradicen? Los huecos (categoría 3) los reporta `sf doctor`, que es donde se
// va a hacer algo con ellos — meterlos acá pondría en DRIFT a toda feature sin
// tests y volvería inútil la columna.
func checkFeatureDrift(projectDir, feature, tracePath, runTests string) []string {
	c := classifyFeatureDrift(projectDir, feature, tracePath, runTests)
	var drifts []string
	for _, it := range append(append([]driftItem{}, c.notImplemented...), c.implementedDifferently...) {
		drifts = append(drifts, fmt.Sprintf("%s / %s: %s", it.Feature, it.Requirement, it.Reason))
	}
	return drifts
}

// splitAnchor separa un anchor en (ruta, símbolo). Soporta dos formas:
//   - path::nodeid  → id de test estilo pytest (tests/foo.py::Clase::test_x).
//     El símbolo es el ÚLTIMO segmento del nodeid, sin el sufijo de
//     parametrización: test_x[caso-1] → test_x.
//   - path:symbol   → anchor de código (src/foo.go:Parse, src/bar.py:Klass.method)
//     o de test estilo Go (foo_test.go:TestParse).
//
// Si no hay separador, el anchor es solo una ruta y el símbolo queda "".
//
// Concepto Go: hay que chequear "::" ANTES que ":" porque "::" contiene ":".
// Partir por el último ":" cortaba mal un nodeid de pytest (tests/foo.py: +
// :test_x) — ese era el bug que impedía verificar test anchors.
func splitAnchor(anchor string) (path, symbol string) {
	// strings.Cut parte por la PRIMERA aparición de "::" y devuelve
	// (antes, despues, encontrado). Es justo lo que queremos para separar la ruta.
	if before, node, found := strings.Cut(anchor, "::"); found {
		// pytest anida con "::" (archivo::Clase::test). El símbolo real es el
		// último tramo, así que si quedan más "::" nos quedamos con lo de después.
		if i := strings.LastIndex(node, "::"); i >= 0 {
			node = node[i+2:]
		}
		// pytest parametriza con sufijo "[caso]"; nos quedamos con el nombre.
		if k := strings.IndexByte(node, '['); k >= 0 {
			node = node[:k]
		}
		return before, node
	}
	if i := strings.LastIndex(anchor, ":"); i >= 0 {
		return anchor[:i], anchor[i+1:]
	}
	return anchor, ""
}

// checkAnchor (estático): ¿sigue existiendo `path:símbolo`? Devuelve (ok, razón).
// Si no hay símbolo, el anchor es solo un path. El símbolo se busca por su
// identificador final (de `Clase.metodo` toma `metodo`) con límites de palabra.
func checkAnchor(projectDir, anchor string) (bool, string) {
	pathPart, symbol := splitAnchor(anchor)

	data, err := os.ReadFile(filepath.Join(projectDir, pathPart))
	if err != nil {
		return false, fmt.Sprintf("file gone: %s", pathPart)
	}
	if symbol != "" {
		parts := strings.Split(symbol, ".")
		ident := parts[len(parts)-1]
		// QuoteMeta escapa cualquier carácter especial de regex en el símbolo.
		re := regexp.MustCompile(`\b` + regexp.QuoteMeta(ident) + `\b`)
		if !re.Match(data) {
			return false, fmt.Sprintf("symbol gone: %s not in %s", symbol, pathPart)
		}
	}
	return true, ""
}

// runTest corre un test (modo dinámico, opt-in). El template trae "{test}" que
// reemplazamos por el id. Usamos el shell nativo (shellArgs: sh -c / cmd /c —
// D6') para soportar pipes/flags. context.WithTimeout corta a los 300s.
func runTest(projectDir, testID, template string) (bool, string) {
	cmd := strings.ReplaceAll(template, "{test}", testID)

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel() // libera el timer pase lo que pase

	sh, flag := shellArgs()
	c := exec.CommandContext(ctx, sh, flag, cmd)
	c.Dir = projectDir

	err := c.Run()
	if err == nil {
		return true, ""
	}
	// errors.As distingue "el test corrió y falló" (ExitError, exit != 0) de "no
	// se pudo ni lanzar" (binario ausente, timeout, etc.).
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		return false, fmt.Sprintf("test failed: %s", testID)
	}
	return false, fmt.Sprintf("could not run test: %v", err)
}
