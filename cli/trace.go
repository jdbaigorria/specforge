package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf save trace` — trace.json es estado protegido (Capa 1), así que el único
// camino para escribirlo es vía `sf save`, que valida + escribe. checkTrace y
// renderTraceMarkdown son las dos mitades del contrato `artifact` (ver save.go).
// ----------------------------------------------------------------------------

// checkTrace valida la matriz: feature + al menos una requirement, cada id no
// vacío con al menos un anchor de código. La ausencia de test es WARNING (el
// verdict, Capa 2, sí la exige; acá no rechazamos guardar un trace en progreso).
func checkTrace(t traceFile, rep *report) {
	if strings.TrimSpace(t.Feature) == "" {
		rep.errorf("feature is required")
	}
	if len(t.Requirements) == 0 {
		rep.errorf("at least one requirement is required")
	}
	ids := make([]string, 0, len(t.Requirements))
	for id := range t.Requirements {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		info := t.Requirements[id]
		if len(info.Code) == 0 {
			rep.errorf("%s: at least one code anchor (path:symbol) is required", id)
		}
		if len(info.Test) == 0 && len(info.Scenarios) == 0 {
			rep.warnf("%s: names no test — the verdict gate will reject this until it does", id)
		}

		// RM-C1: los ids de escenario tienen que pertenecer a SU requisito. Un
		// `R6.1` colgando de `R5` anclaría tests a un requisito ajeno, y el
		// contrato del veredicto los daría por buenos.
		scenarioIDs := make([]string, 0, len(info.Scenarios))
		for sid := range info.Scenarios {
			scenarioIDs = append(scenarioIDs, sid)
		}
		sort.Strings(scenarioIDs) // los mapas de Go iteran al azar
		for _, sid := range scenarioIDs {
			m := acceptanceIDRe.FindStringSubmatch(sid)
			switch {
			case m == nil:
				rep.errorf("%s: scenario id %q must have the form <requirement>.<n>, e.g. %s.1", id, sid, id)
			case m[1] != id:
				rep.errorf("%s: scenario %s belongs to %s, not to %s", id, sid, m[1], id)
			default:
				sc := info.Scenarios[sid]
				// RM-C4: un escenario se verifica con tests O con evidencia.
				// Los dos juntos no es un error (un benchmark puede tener además
				// un test de humo), pero ninguno de los dos sí.
				if len(sc.Test) == 0 && sc.Evidence == nil {
					rep.warnf("%s: names no test and declares no evidence — the verdict gate will reject this until it does", sid)
				}
				if sc.Evidence != nil {
					if !verificationMethods[sc.Evidence.Kind] {
						rep.errorf("%s: evidence kind %q is not one of test|benchmark|audit|manual|analysis", sid, sc.Evidence.Kind)
					}
					if strings.TrimSpace(sc.Evidence.Ref) == "" {
						rep.errorf("%s: evidence ref is required — evidence nobody can open is an assertion", sid)
					}
					if !capturedRe.MatchString(sc.Evidence.Recorded) {
						rep.errorf("%s: evidence recorded %q must be YYYY-MM-DD", sid, sc.Evidence.Recorded)
					}
				}
			}
		}
	}
}

// renderTraceMarkdown arma la matriz legible (sin template: es una tabla simple).
func renderTraceMarkdown(t traceFile) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "# Traceability — %s\n\n", t.Feature)
	fmt.Fprintln(&b, "| Requirement | Code | Tests | Status |")
	fmt.Fprintln(&b, "|---|---|---|---|")
	ids := make([]string, 0, len(t.Requirements))
	for id := range t.Requirements {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		info := t.Requirements[id]
		fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
			id, joinOrDash(info.Code), joinOrDash(info.Test), orDash(info.Status))
	}
	return b.String(), nil
}

// runTrace es el punto de entrada de `sf trace ...`. Hoy la única sub-acción es
// `verify`. Reutiliza el modelo traceFile/traceReq y checkAnchor de doctor.go, y
// findArtifact de status.go.
func runTrace(args []string) int {
	if len(args) == 0 || args[0] != "verify" {
		fmt.Fprintln(os.Stderr, "usage: sf trace verify [project_dir] [--feature=NAME] [--contract] [--wave=N]")
		return 2
	}
	args = args[1:] // descartamos "verify"

	projectDir := "."
	feature := ""
	contract := false
	wave := -1 // -1 = sin --wave (scope = la feature entera)
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case a == "--contract":
			contract = true
		case strings.HasPrefix(a, "--wave="):
			n, err := strconv.Atoi(strings.TrimPrefix(a, "--wave="))
			if err != nil || n < 0 {
				fmt.Fprintln(os.Stderr, "sf trace: --wave must be a non-negative integer")
				return 2
			}
			wave = n
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf trace: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	// El modo --contract es la verificación build-time (cobertura declarada).
	// El modo normal (sin --contract) es la verificación de drift de código.
	if contract {
		if feature == "" {
			fmt.Fprintln(os.Stderr, "sf trace verify --contract: --feature=NAME is required")
			return 2
		}
		return runTraceContract(projectDir, feature, wave)
	}
	if wave >= 0 {
		fmt.Fprintln(os.Stderr, "sf trace verify: --wave only applies with --contract")
		return 2
	}
	return runTraceVerify(projectDir, feature)
}

// runTraceVerify valida la matriz de trazabilidad contra el disco. Con --feature
// valida esa feature (exit 4 si no tiene trace.json); sin --feature valida todas
// las que tengan trace.json. Exit 1 si alguna requirement divergió.
func runTraceVerify(projectDir, feature string) int {
	specforge := filepath.Join(projectDir, "specforge")

	if feature != "" {
		tracePath := findArtifact(specforge, feature, "trace.json")
		if tracePath == "" {
			fmt.Fprintf(os.Stderr, "sf trace: no trace.json for feature %q.\n", feature)
			return 4
		}
		return verifyTrace(projectDir, feature, tracePath)
	}

	traces := loadTraces(specforge)
	if len(traces) == 0 {
		fmt.Printf("No trace.json found under %s/specforge.\n", projectDir)
		return 0
	}
	exit := 0
	for i, t := range traces {
		if i > 0 {
			fmt.Println() // separación entre matrices
		}
		if verifyTrace(projectDir, t.feature, t.path) != 0 {
			exit = 1
		}
	}
	return exit
}

// verifyTrace imprime la matriz de una feature y devuelve 1 si alguna
// requirement divergió (algún anchor de código ya no existe), 0 si todo en sync.
// Columnas: la "recorded" es el status que grabó sf-check; la "live" es lo que
// verificamos ahora contra el código real.
func verifyTrace(projectDir, feature, tracePath string) int {
	data, err := os.ReadFile(tracePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf trace: cannot read %s (%v)\n", tracePath, err)
		return 1
	}
	var tf traceFile
	if err := json.Unmarshal(data, &tf); err != nil {
		fmt.Fprintf(os.Stderr, "sf trace: invalid trace.json for %s (%v)\n", feature, err)
		return 1
	}

	// Orden estable de requirements (los maps de Go iteran aleatorio).
	reqIDs := make([]string, 0, len(tf.Requirements))
	for id := range tf.Requirements {
		reqIDs = append(reqIDs, id)
	}
	sort.Strings(reqIDs)

	fmt.Printf("Trace matrix — %s\n\n", feature)

	cols := []string{"requirement", "code", "tests", "recorded", "live"}
	var rows [][]string
	drifted := 0
	for _, req := range reqIDs {
		info := tf.Requirements[req]
		live := "ok"
		// Drift de CÓDIGO: algún anchor de implementación ya no existe.
		for _, anchor := range info.Code {
			if ok, _ := checkAnchor(projectDir, anchor); !ok {
				live = "DRIFT"
			}
		}
		// Drift de TEST (refuerzo FIXBUGHIGH): el requirement debe nombrar al
		// menos un test, y cada test ref debe resolver a un test real. En el caso
		// real, requirements "verificados" citaban tests en archivos inexistentes
		// y verify no lo veía porque solo miraba el código.
		if len(info.Test) == 0 {
			live = "NO TEST"
		}
		for _, tref := range info.Test {
			if ok, _ := checkAnchor(projectDir, tref); !ok {
				live = "TEST GONE"
			}
		}
		if live != "ok" {
			drifted++
		}
		rows = append(rows, []string{
			req,
			joinOrDash(info.Code),
			joinOrDash(info.Test),
			orDash(info.Status),
			live,
		})
	}
	renderTable(cols, rows)

	if drifted > 0 {
		fmt.Printf("\nDRIFT: %d requirement(s) diverged (code drift, missing or vanished tests).\n", drifted)
		return 1
	}
	fmt.Println("\nOK: all requirements traced to live code and existing tests.")
	return 0
}

// runTraceContract valida el CONTRATO DE VERIFICACIÓN en build-time: cada
// requirement EN SCOPE debe estar declarado en trace.json con al menos un test, y
// ese test debe EXISTIR de verdad en el código. Ojo: "existir", no "pasar" — que
// el test pase es drift (verify normal / --run-tests); el contrato exige que el
// agente NOMBRE un test real para cada R que implementa. Eso convierte la regla
// blanda "test pass != done" en un artefacto chequeable ANTES de cerrar la wave.
//
// Scope: con --wave=N el contrato cubre solo los requirements que tocan las tasks
// de esa wave (vía plan.json + tasks.json), porque durante el build trace.json
// está parcial — pedir cobertura de R todavía no construidos daría falsos
// faltantes. Sin --wave cubre todos los R ya declarados en trace.json (contrato
// de feature completa, útil al cierre).
//
// Exit: 0 contrato cumplido, 1 incumplido, 4 falta trace.json.
func runTraceContract(projectDir, feature string, wave int) int {
	specforge := filepath.Join(projectDir, "specforge")
	tracePath := findArtifact(specforge, feature, "trace.json")
	if tracePath == "" {
		fmt.Fprintf(os.Stderr, "sf trace: no trace.json for feature %q (build must declare coverage first).\n", feature)
		return 4
	}
	data, err := os.ReadFile(tracePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf trace: cannot read %s (%v)\n", tracePath, err)
		return 1
	}
	var tf traceFile
	if err := json.Unmarshal(data, &tf); err != nil {
		fmt.Fprintf(os.Stderr, "sf trace: invalid trace.json for %s (%v)\n", feature, err)
		return 1
	}

	// Armamos el conjunto de requirements en scope.
	var scope []string
	if wave >= 0 {
		s, code := contractScope(projectDir, feature, wave)
		if code != 0 {
			return code
		}
		scope = s
	} else {
		// Feature entera: todos los requirements ya declarados en trace.json.
		for id := range tf.Requirements {
			scope = append(scope, id)
		}
	}
	sort.Strings(scope) // orden estable de salida

	scopeLabel := "feature"
	if wave >= 0 {
		scopeLabel = fmt.Sprintf("wave %d", wave)
	}
	fmt.Printf("Verification contract — %s (%s)\n\n", feature, scopeLabel)

	cols := []string{"requirement", "test", "exists", "verdict"}
	var rows [][]string
	unmet := 0
	for _, req := range scope {
		info, declared := tf.Requirements[req]
		verdict := "ok"
		existsCol := "—"
		switch {
		case !declared || len(info.Test) == 0:
			// No hay test nombrado para este R → contrato incumplido.
			verdict = "NO TEST"
			unmet++
		default:
			// Hay test(s) declarados: cada uno debe resolver a código real.
			missing := false
			for _, t := range info.Test {
				if ok, _ := checkAnchor(projectDir, t); !ok {
					missing = true
				}
			}
			if missing {
				verdict = "TEST GONE"
				existsCol = "no"
				unmet++
			} else {
				existsCol = "yes"
			}
		}
		rows = append(rows, []string{req, joinOrDash(info.Test), existsCol, verdict})
	}
	renderTable(cols, rows)

	if len(scope) == 0 {
		fmt.Println("\nNo requirements in scope.")
		return 0
	}
	if unmet > 0 {
		fmt.Printf("\nCONTRACT UNMET: %d requirement(s) without a verifiable test. Wave gate blocked.\n", unmet)
		return 1
	}
	fmt.Println("\nOK: every requirement in scope names a test that exists.")
	return 0
}

// contractScope devuelve los requirement-ids que el contrato debe cubrir para una
// wave: la unión de los requirement_refs de las tasks de esa wave. Lee el plan
// APROBADO (plan.json) en vez de recomputar desde tasks.json, para respetar los
// refinamientos de wave (merge/split) que el humano aprobó en el gate del plan.
// Reusa readPlanFile y readTasksFile (mismo paquete).
func contractScope(projectDir, feature string, wave int) (reqs []string, code int) {
	pf, c := readPlanFile(projectDir, feature)
	if c != 0 {
		return nil, c
	}
	var taskIDs []string
	found := false
	for _, w := range pf.Waves {
		if w.N == wave {
			taskIDs = w.Tasks
			found = true
			break
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "sf trace: plan.json has no wave %d for feature %q.\n", wave, feature)
		return nil, 2
	}

	tf, c := readTasksFile(projectDir, feature)
	if c != 0 {
		return nil, c
	}
	refsByID := map[string][]string{} // task-id → requirement_refs
	for _, t := range tf.Tasks {
		refsByID[t.ID] = t.RequirementRefs
	}

	seen := map[string]bool{} // dedup: dos tasks de la wave pueden tocar el mismo R
	for _, id := range taskIDs {
		for _, r := range refsByID[id] {
			if !seen[r] {
				seen[r] = true
				reqs = append(reqs, r)
			}
		}
	}
	return reqs, 0
}
