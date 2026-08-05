package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// nowUTC devuelve el instante actual en UTC con formato RFC3339 (el mismo de los
// gates en features.json, ej. "2026-06-20T14:00:00Z").
func nowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// runGate es el punto de entrada de `sf gate ...`. Sub-acciones: `status` (lee
// el ledger de gates humanos de features.json) y `record-verdict` (persiste el
// veredicto del auditor de fase, tier calidad).
func runGate(args []string) int {
	if len(args) == 0 {
		gateUsage()
		return 2
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "status":
		return runGateStatusCmd(rest)
	case "approve":
		return runGateApprove(rest)
	case "record-verdict":
		return runGateRecordVerdict(rest)
	case "show":
		return runGateShow(rest)
	default:
		fmt.Fprintf(os.Stderr, "sf gate: unknown sub-command %q\n\n", sub)
		gateUsage()
		return 2
	}
}

func gateUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf gate status [--feature=NAME] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf gate approve --feature=NAME --phase=PHASE [--by=user] [--comment=TEXT] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf gate record-verdict --feature=NAME [--phase=PHASE] [--json -|FILE] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf gate show --feature=NAME [--phase=PHASE] [--json] [project_dir]")
}

// ----------------------------------------------------------------------------
// `sf gate approve` — registra DETERMINÍSTICAMENTE un gate humano aprobado y, lo
// clave, SELLA el hash del artefacto. Es el primer comando del CLI que ESCRIBE
// features.json (hasta ahora el ledger lo escribía el skill a mano). Sellar el
// hash acá —y no pedírselo al LLM— es lo que hace confiable la detección de
// silent edits: el hash lo computa la máquina sobre el archivo real.
// ----------------------------------------------------------------------------

func runGateApprove(args []string) int {
	projectDir := "."
	feature, phase, by, comment := "", "", "user", ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--phase="):
			phase = strings.TrimPrefix(a, "--phase=")
		case strings.HasPrefix(a, "--by="):
			by = strings.TrimPrefix(a, "--by=")
		case strings.HasPrefix(a, "--comment="):
			comment = strings.TrimPrefix(a, "--comment=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf gate approve: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" || phase == "" {
		gateUsage()
		return 2
	}
	return gateApprove(projectDir, feature, phase, by, comment)
}

// gateApprove hace el trabajo: lee el estado de features, sella el hash del
// artefacto de la fase, apendea el gate y reescribe el feature.json de ESA
// feature (A1: nadie toca el estado de las demás). El parámetro se llama `name`
// (no `feature`) para no tapar al tipo `feature`.
func gateApprove(projectDir, name, phase, by, comment string) int {
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf gate approve: cannot read feature state under %s (%v)\n", projectDir, err)
		return 4
	}

	// Buscamos la feature por nombre. Tomamos el puntero al elemento real del
	// slice (no una copia) para poder mutar sus Gates.
	var f *feature
	for i := range ff.Features {
		if ff.Features[i].Name == name {
			f = &ff.Features[i]
			break
		}
	}
	if f == nil {
		fmt.Fprintf(os.Stderr, "sf gate approve: feature %q not found\n", name)
		return 4
	}

	// R1 (integrity.go): no EXTENDEMOS un ledger roto. Si la cadena no valida,
	// alguien escribió gates fuera del CLI; aprobar encima legitimaría el fraude.
	if refuseOnBrokenLedger("sf gate approve", f) {
		return 5
	}

	// Capa 2: el verdict es el sello final (siguiente paso = archive). No se
	// otorga si la trazabilidad driftó, si algún requirement no nombra un test
	// real, o si no hay un resultado de test verde y fresco. El LLM no puede
	// saltearse esto editando a mano: features.json está protegido (Capa 1).
	if phase == "verdict" {
		// Las advertencias se imprimen SIEMPRE, se bloquee o no. Bajar el listón
		// con `blocking_priorities` es decidir que algo no frena el release, no
		// decidir dejar de verlo — y una advertencia que sólo aparece cuando ya
		// hay un bloqueo es una advertencia que nadie lee nunca.
		for _, w := range verdictWarnings(projectDir, name) {
			fmt.Fprintf(os.Stderr, "  warning: %s\n", w)
		}
		if reasons := verdictPreconditions(projectDir, name); len(reasons) > 0 {
			// Telemetría (A7): un verdict rehusado es exactamente el dato que
			// queremos poder contar después.
			logEvent(projectDir, sfEvent{Kind: "refuse", Feature: name, Phase: "verdict",
				Detail: fmt.Sprintf("verdict refused: %d precondition(s) unmet", len(reasons))})
			fmt.Fprintf(os.Stderr, "sf gate approve: verdict refused for %q — %d precondition(s) unmet:\n", name, len(reasons))
			for _, r := range reasons {
				fmt.Fprintf(os.Stderr, "  - %s\n", r)
			}
			return 5
		}
	}

	// Sellamos el hash del artefacto de esta fase (si la fase tiene artefacto).
	// No se puede aprobar un artefacto que falta: sería certificar el vacío.
	hash := ""
	if rel, hasArtifact := artifactFileForPhase(phase); hasArtifact {
		abs := filepath.Join(projectDir, "specforge", "features", name, rel)
		h, ok := hashArtifact(abs)
		if !ok {
			fmt.Fprintf(os.Stderr, "sf gate approve: artifact for phase %q not found (%s)\n", phase, rel)
			return 4
		}
		hash = h
	}

	// Prev tiene que computarse ANTES del append (es el hash de la última
	// entrada EXISTENTE; después del append "la última" sería esta misma).
	f.Gates = append(f.Gates, gate{
		Phase:   phase,
		Result:  "approve",
		By:      by,
		At:      nowUTC(),
		Comment: comment,
		Hash:    hash,
		Prev:    nextPrev(f),
		// R10: sólo el verdict lleva contrato — es el único gate que verifica
		// trazabilidad. Anotarlo en `requirements` o `design` sería ruido.
		Contract: verdictContract(projectDir, name, phase),
	})

	if code := writeFeatureState(projectDir, *f); code != 0 {
		return code
	}
	if hash != "" {
		fmt.Printf("approved %s/%s (hash %s…)\n", name, phase, hash[:12])
	} else {
		fmt.Printf("approved %s/%s\n", name, phase)
	}
	if phase == "verdict" {
		fmt.Printf("  verification contract: %s\n", verdictContract(projectDir, name, phase))
	}
	return 0
}

// contractV1 / contractV2: las dos reglas de verificación que pueden haber
// sellado un verdict.
const (
	contractV1 = "v1" // un test por requisito
	contractV2 = "v2" // un test por criterio de aceptación
)

// verdictContract responde bajo qué contrato se está sellando.
//
// Un proyecto puede ser MIXTO durante la transición, así que la respuesta es la
// del conjunto: alcanza con que UN requisito tenga acceptance estructurada para
// que la feature se haya verificado bajo la regla estricta. Decir `v1` en ese
// caso subvaluaría lo que el sello certifica.
func verdictContract(projectDir, feature, phase string) string {
	if phase != "verdict" {
		return ""
	}
	if len(acceptanceCriteriaOf(projectDir, feature)) > 0 {
		return contractV2
	}
	return contractV1
}

// runGateStatusCmd parsea los flags de `status`.
func runGateStatusCmd(args []string) int {
	projectDir := "."
	feature := ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf gate: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	return runGateStatus(projectDir, feature)
}

// runGateStatus lee el gate ledger. Sin --feature muestra un resumen por
// feature; con --feature muestra el ledger detallado de esa feature.
func runGateStatus(projectDir, feature string) int {
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Printf("No feature state under %s/specforge.\n", projectDir)
		return 0
	}

	if feature != "" {
		return gateLedger(ff, feature)
	}
	gateSummary(ff)
	return 0
}

// gateSummary: una fila por feature con el conteo de gates y el último estado.
func gateSummary(ff featuresFile) {
	fmt.Print("SpecForge — gate status\n\n")
	if len(ff.Features) == 0 {
		fmt.Println("No features registered yet.")
		return
	}

	cols := []string{"feature", "status", "gates", "phase", "last-result", "when"}
	var rows [][]string
	for i := range ff.Features {
		f := &ff.Features[i]
		lastResult, when := "—", "—"
		if n := len(f.Gates); n > 0 {
			lastResult = orDash(f.Gates[n-1].Result)
			when = orDash(f.Gates[n-1].At)
		}
		rows = append(rows, []string{
			f.Name,
			orDash(f.Status),
			fmt.Sprintf("%d", len(f.Gates)),
			lastPhase(f),
			lastResult,
			when,
		})
	}
	renderTable(cols, rows)
}

// gateLedger: el historial completo de gates de una feature. Devuelve exit 4
// (not found) si la feature no existe.
func gateLedger(ff featuresFile, name string) int {
	var f *feature
	for i := range ff.Features {
		if ff.Features[i].Name == name {
			f = &ff.Features[i]
			break
		}
	}
	if f == nil {
		fmt.Fprintf(os.Stderr, "sf gate: feature %q not found.\n", name)
		return 4
	}

	fmt.Printf("Gate ledger — %s  (status: %s, current phase: %s)\n\n",
		f.Name, orDash(f.Status), lastPhase(f))
	if len(f.Gates) == 0 {
		fmt.Println("  no gates recorded yet.")
		return 0
	}

	cols := []string{"phase", "result", "by", "at", "comment"}
	var rows [][]string
	for _, g := range f.Gates {
		rows = append(rows, []string{
			g.Phase,
			g.Result,
			orDash(g.By),
			orDash(g.At),
			orDash(g.Comment),
		})
	}
	renderTable(cols, rows)
	return 0
}

// ----------------------------------------------------------------------------
// `sf gate record-verdict` — persiste el veredicto del AUDITOR DE FASE.
//
// El juez (subagente fresco) produce {phase, verdicts:[{rule,result,citation}]}.
// El CLI computa el `overall` (fail si alguna regla falla), le pone timestamp y
// lo APENDEA a specforge/features/<f>/audit.json — el ledger de calidad,
// SEPARADO de los gates humanos de features.json y de la matriz trace.json.
//
// El CLI solo PERSISTE (determinista). La decisión nudge/block la toma el hook
// según el `overall`: exit 0 = pass, exit 3 = recorded-but-FAIL → señal limpia
// para que el hook nudgee sin parsear stdout.
// ----------------------------------------------------------------------------

type auditLedger struct {
	Feature string       `json:"feature"`
	Entries []auditEntry `json:"entries"`
}

type auditEntry struct {
	Phase    string        `json:"phase"`
	At       string        `json:"at"`
	Overall  string        `json:"overall"` // pass | fail
	Verdicts []ruleVerdict `json:"verdicts"`
}

type ruleVerdict struct {
	Rule     string `json:"rule"`
	Result   string `json:"result"` // pass | fail
	Citation string `json:"citation"`
	// CitationCheck (A8): el CLI verifica la citation contra el artefacto real.
	// "verified" = aparece textual (whitespace-normalizado); "not-found" = el juez
	// citó algo que NO está en el artefacto (citation inventada — la mentira
	// semántica que sigue después de cerrar la escritura de estado); "" = no hay
	// citation o la fase no tiene artefacto contra qué verificar.
	CitationCheck string `json:"citation_check,omitempty"`
}

var verdictResults = map[string]bool{"pass": true, "fail": true}

func runGateRecordVerdict(args []string) int {
	projectDir := "."
	feature := ""
	phase := ""
	jsonSrc := "-"
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--phase="):
			phase = strings.TrimPrefix(a, "--phase=")
		case strings.HasPrefix(a, "--json="):
			jsonSrc = strings.TrimPrefix(a, "--json=")
		case a == "--json":
			if i+1 < len(args) {
				jsonSrc = args[i+1]
				i++
			}
		case a != "-" && strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf gate: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf gate: --feature=NAME is required")
		return 2
	}

	raw, err := readJSONInput(jsonSrc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: cannot read input (%v)\n", err)
		return 1
	}
	// Solo nos interesan phase + verdicts del JSON del juez.
	var in struct {
		Phase    string        `json:"phase"`
		Verdicts []ruleVerdict `json:"verdicts"`
	}
	if err := json.Unmarshal(raw, &in); err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: invalid JSON (%v)\n", err)
		return 2
	}
	if phase != "" {
		in.Phase = phase // el flag manda
	}

	entry, rep := buildAuditEntry(in.Phase, in.Verdicts)
	if len(rep.errors) > 0 {
		for _, e := range rep.errors {
			fmt.Printf("  ERROR:   %s\n", e)
		}
		fmt.Printf("\nFAIL: not recorded — %d error(s).\n", len(rep.errors))
		return 2
	}

	// A8: cada citation se verifica mecánicamente contra el artefacto antes de
	// entrar al ledger. No rechazamos (el juez es nudge, no gate duro), pero lo
	// no-verificado queda MARCADO — el humano del gate lo ve.
	verifyCitations(projectDir, feature, &entry)
	for _, v := range entry.Verdicts {
		if v.CitationCheck == "not-found" {
			fmt.Printf("  warning: rule %q cites text not found in the %s artifact — recorded as citation_check=not-found\n",
				v.Rule, entry.Phase)
		}
	}

	if code := appendAuditEntry(projectDir, feature, entry); code != 0 {
		return code
	}

	// Telemetría (A7): cada veredicto del juez queda contable — los "fail" por
	// feature son la medida de presión REVISE.
	logEvent(projectDir, sfEvent{Kind: "verdict", Feature: feature, Phase: entry.Phase, Detail: entry.Overall})

	fmt.Printf("recorded %s verdict for %s/%s (%d rule(s))\n", entry.Overall, feature, entry.Phase, len(entry.Verdicts))
	if entry.Overall == "fail" {
		// D3': escalación del loop REVISE. N fails CONSECUTIVOS de la misma fase
		// ya no son "iterá de nuevo" — son señal de que la SPEC está mal. Hasta
		// ahora detectar eso era responsabilidad del humano; es un contador, y
		// los contadores son trabajo de la máquina.
		if n := consecutiveFails(projectDir, feature, entry.Phase); n >= reviseEscalationThreshold {
			fmt.Printf("ESCALATE: %d consecutive failed verdicts on %s/%s — the spec itself is likely wrong.\n",
				n, feature, entry.Phase)
			fmt.Println("Stop iterating this phase. Go back to propose: re-open the upstream artifact " +
				"(requirements/design), fix the spec, and let the resync cascade re-derive downstream.")
			logEvent(projectDir, sfEvent{Kind: "refuse", Feature: feature, Phase: entry.Phase,
				Detail: fmt.Sprintf("REVISE escalation: %d consecutive fails", n)})
		}
		return 3 // recorded, pero el veredicto es FAIL → el hook nudgea
	}
	return 0
}

// reviseEscalationThreshold: a partir de cuántos fails seguidos de una fase
// dejamos de sugerir "otra vuelta" y escalamos a re-abrir la spec.
const reviseEscalationThreshold = 3

// consecutiveFails cuenta los veredictos FAIL consecutivos MÁS RECIENTES de una
// fase (el fail recién grabado incluido). Un pass corta la racha.
func consecutiveFails(projectDir, feature, phase string) int {
	entries, ok := readAuditLedger(projectDir, feature)
	if !ok {
		return 0
	}
	n := 0
	for i := len(entries) - 1; i >= 0; i-- {
		if entries[i].Phase != phase {
			continue
		}
		if entries[i].Overall != "fail" {
			break
		}
		n++
	}
	return n
}

// ----------------------------------------------------------------------------
// Capa 2 de FIXBUGHIGH — el sello final (gate `verdict`) no se puede falsificar.
//
// `gate approve --phase=verdict` es el último gate antes de archivar (done). Sin
// esta guarda, el agente podía aprobarlo igual que cualquier otro y declarar
// "todos los tests pasan / 70% coverage" sin haber corrido nada. Ahora el CLI
// REHÚSA el verdict salvo que tres condiciones se cumplan, todas verificadas por
// la máquina contra el disco real (no contra la narración del LLM).
// ----------------------------------------------------------------------------

// verdictPreconditions devuelve la lista de razones por las que el verdict NO
// puede otorgarse (vacía = todo en orden):
//  1. trace.json existe y NINGUNA requirement driftó (todos sus code anchors viven).
//  2. cada requirement nombra ≥1 test y cada test ref resuelve a un test real
//     (el contrato de verificación: "test pass != done" se vuelve chequeable).
//  3. hay un resultado de test VERDE y FRESCO: `sf check run` pasó y su code_hash
//     == el code_hash actual (no se tocó el código después de correr).
//  4. (si la corrida sellada trae reporte por-test) cada test nombrado en el
//     trace CORRIÓ y PASÓ en esa corrida — causalidad test→requirement (A2/R4).
//
// verdictIssue es un incumplimiento del contrato, con su severidad.
//
// Que la severidad viaje CON la razón (y no en dos listas paralelas) es lo que
// impide el bug clásico: alguien filtra por bloqueantes, reporta esa lista, y
// las advertencias desaparecen sin que nadie note que se perdieron.
type verdictIssue struct {
	Reason   string
	Blocking bool
}

// verdictPreconditions devuelve sólo las razones BLOQUEANTES. Es el envoltorio
// que consumen `sf verify` y el gate para decidir sí/no.
func verdictPreconditions(projectDir, feature string) []string {
	var out []string
	for _, is := range verdictIssues(projectDir, feature) {
		if is.Blocking {
			out = append(out, is.Reason)
		}
	}
	return out
}

// verdictWarnings devuelve las razones NO bloqueantes: incumplimientos reales de
// requisitos cuya prioridad el proyecto declaró como no-bloqueante. Se reportan
// siempre — bajar el listón no es lo mismo que dejar de mirar.
func verdictWarnings(projectDir, feature string) []string {
	var out []string
	for _, is := range verdictIssues(projectDir, feature) {
		if !is.Blocking {
			out = append(out, is.Reason)
		}
	}
	return out
}

func verdictIssues(projectDir, feature string) []verdictIssue {
	var issues []verdictIssue

	// block: incumplimientos que NO son atribuibles a un requisito con
	// prioridad — falta el trace, no se corrió la suite, el resultado está
	// stale. Nada de eso se gradúa: son condiciones del proceso, no del
	// contenido de un requisito.
	block := func(format string, a ...any) {
		issues = append(issues, verdictIssue{Reason: fmt.Sprintf(format, a...), Blocking: true})
	}

	// traced queda apuntando al trace parseado OK — lo necesita la condición 4
	// (causalidad test→requirement contra el reporte sellado).
	var traced *traceFile

	specforge := filepath.Join(projectDir, "specforge")
	tracePath := findArtifact(specforge, feature, "trace.json")
	if tracePath == "" {
		block("no trace.json — produce it with `sf save trace --feature=%s --json -`", feature)
	} else if data, err := os.ReadFile(tracePath); err != nil {
		block("trace.json is unreadable")
	} else {
		var tf traceFile
		if json.Unmarshal(data, &tf) != nil {
			block("trace.json is invalid JSON")
		} else if len(tf.Requirements) == 0 {
			block("trace.json declares no requirements")
		} else {
			traced = &tf
			// Orden estable para que el reporte sea reproducible.
			ids := make([]string, 0, len(tf.Requirements))
			for id := range tf.Requirements {
				ids = append(ids, id)
			}
			sort.Strings(ids)
			// RM-C1: qué criterios declara cada requisito. Un requisito con
			// `acceptance` estructurada corre bajo el contrato v2 (un test POR
			// CRITERIO); uno legado conserva la regla vieja.
			criteria := acceptanceCriteriaOf(projectDir, feature)
			// RM-C2: la prioridad de cada requisito y hasta dónde bloquea el
			// proyecto. `blocks` viene con TODO en true salvo que la constitución
			// diga otra cosa — el default no afloja nada.
			// RM-C4: y CÓMO se verifica cada uno (test, benchmark, audit, …).
			prio := map[string]string{}
			method := map[string]string{}
			for _, r := range requirementsOf(projectDir, feature) {
				prio[r.ID] = priorityOf(r)
				method[r.ID] = verificationOf(r)
			}
			blocks := blockingPriorities(projectDir)

			for _, req := range ids {
				info := tf.Requirements[req]

				// Código que ya no existe: BLOQUEA siempre, sin importar la
				// prioridad. Un ancla rota no es "un requisito menor sin test":
				// es un artefacto que miente sobre dónde vive lo que describe, y
				// eso corrompe el drift de todo el proyecto. R5 gradúa el
				// CONTRATO DE VERIFICACIÓN, no la integridad del trace.
				for _, anchor := range info.Code {
					if ok, why := checkAnchor(projectDir, anchor); !ok {
						block("%s: code drift (%s)", req, why)
					}
				}

				// De acá para abajo sí se gradúa: son incumplimientos del
				// contrato de verificación, atribuibles a UN requisito.
				p := prio[req]
				if p == "" {
					p = "must" // sin spec que lo declare: fail-closed
				}
				blocking := blocks[p]
				add := func(reason string) {
					if !blocking {
						reason = fmt.Sprintf("%s [priority: %s — reported, not blocking]", reason, p)
					}
					issues = append(issues, verdictIssue{Reason: reason, Blocking: blocking})
				}

				m := method[req]
				if m == "" {
					m = "test" // sin spec que lo declare: fail-closed
				}
				if wanted, v2 := criteria[req]; v2 {
					for _, r := range scenarioReasons(projectDir, req, m, wanted, info) {
						add(r)
					}
				} else if len(info.Test) == 0 {
					// R4: contrato v1 intacto para los requisitos legados. Un
					// legado que declara `verification != test` no tiene dónde
					// poner la evidencia (no hay escenarios), así que se le pide
					// lo único que puede dar: subir a acceptance con ids.
					if m != "test" {
						add(fmt.Sprintf("%s: verification is %q but the requirement has no acceptance ids — "+
							"give its criteria ids so the evidence has somewhere to live", req, m))
					} else {
						add(fmt.Sprintf("%s: names no test (verification contract unmet)", req))
					}
				}
				for _, tref := range info.Test {
					if ok, why := checkAnchor(projectDir, tref); !ok {
						add(fmt.Sprintf("%s: test %q does not resolve (%s)", req, tref, why))
					}
				}
			}
		}
	}

	// Resultado de test verde y FRESCO (Capa 3 lo produjo; acá lo exigimos).
	res, ok := readCheckResult(projectDir, feature)
	switch {
	case !ok:
		block("no test result on record — run `sf check run --feature=%s`", feature)
	case !res.Passed:
		block("last `sf check run` FAILED (exit %d) — fix the code and re-run", res.ExitCode)
	case res.CodeHash != codeHash(projectDir):
		block("test result is STALE — code changed since the last `sf check run`; re-run it")
	default:
		// 4. Causalidad test→requirement (A2/R4): con reporte estructurado en la
		//    corrida sellada, cada test nombrado en el trace debe haber CORRIDO y
		//    PASADO en esa corrida. Sin reporte (len==0) no hay evidencia por-test
		//    y no inventamos la garantía — quedan las condiciones 1-3.
		if traced != nil && len(res.Tests) > 0 {
			for _, r := range causalityReasons(traced, res.Tests) {
				block("%s", r)
			}
		}
	}

	// 5. Testigo RED (RM-C5 / R12), OPT-IN. Bloquea siempre que esté encendido:
	//    el proyecto que lo enciende está pidiendo exactamente esta garantía, así
	//    que graduarla por prioridad la vaciaría de sentido.
	if requireRedWitnessEnabled(projectDir) {
		for _, r := range redWitnessGateReasons(projectDir, feature, traced) {
			block("%s", r)
		}
	}

	// 6. Conformidad de arquitectura (RM-C7), OPT-IN. Bloquea siempre que esté
	//    encendido, por el mismo motivo que el testigo RED: el proyecto que lo
	//    enciende está pidiendo exactamente esta garantía. Y no se gradúa por
	//    prioridad porque no es atribuible a UN requisito — una violación del
	//    grafo es una propiedad de la feature entera.
	if requireArchEnabled(projectDir) {
		for _, r := range archGateReasons(projectDir, feature) {
			block("%s", r)
		}
	}
	return issues
}

// redWitnessGateReasons aplica R12 con sus dos exenciones y su caso de
// configuración faltante.
func redWitnessGateReasons(projectDir, feature string, traced *traceFile) []string {
	// Sin reporte por-test no hay forma de atribuir un fallo a un test, así que
	// los testigos nunca se acumulan. Aprobar en silencio sería lo peor: el
	// proyecto creería tener una garantía que nunca se evaluó.
	if !perTestReportConfigured(projectDir) {
		return []string{
			"require_red_witness is on but build.report is not configured — without a per-test report " +
				"no witness can ever be attributed. Set build.report (go-json|junit) or turn the flag off",
		}
	}

	// RM-C4: un requisito verificado por benchmark o auditoría no tiene testigo
	// RED y no debe exigírsele — no falló nunca como test porque nunca fue uno.
	exempt := map[string]bool{}
	for _, r := range requirementsOf(projectDir, feature) {
		if verificationOf(r) != "test" {
			exempt[r.ID] = true
		}
	}
	return redWitnessReasons(projectDir, feature, traced, exempt)
}

// perTestReportConfigured: ¿la constitución declara build.report?
func perTestReportConfigured(projectDir string) bool {
	data, err := os.ReadFile(filepath.Join(projectDir, "specforge", "constitution.json"))
	if err != nil {
		return false
	}
	var c constitutionFile
	if json.Unmarshal(data, &c) != nil || c.Build == nil {
		return false
	}
	return strings.TrimSpace(c.Build.Report) != ""
}

// ----------------------------------------------------------------------------
// El contrato de verificación v2 (RM-C1 / R3).
//
// v1 exige "el requisito nombra al menos un test". v2 exige "CADA criterio de
// aceptación nombra al menos un test". La diferencia no es de grado: bajo v1, un
// requisito con cinco criterios sella en verde con un test del caso feliz y los
// otros cuatro quedan sin tocar, con el sello certificando correctamente el
// cumplimiento de una regla que mide poco.
//
// Qué requisito corre bajo qué contrato lo dice EL DATO, no una perilla: si su
// `acceptance` tiene ids, corre bajo v2. Los legados conservan v1 (R4), y esa
// ventana se cierra sola cuando `sf migrate` les asigne ids.
// ----------------------------------------------------------------------------

// acceptanceCriteriaOf devuelve, por requisito, los ids de sus criterios — pero
// SÓLO para los requisitos con acceptance estructurada. Un requisito ausente del
// mapa corre bajo v1.
//
// Lectura QUIETA: sin requirements.json (o inválido) no hay criterios que
// exigir y todo cae a v1. El artefacto ya lo validó su propio gate; acá no es
// lugar para volver a quejarse de él.
func acceptanceCriteriaOf(projectDir, feature string) map[string][]string {
	out := map[string][]string{}
	for _, r := range requirementsOf(projectDir, feature) {
		if !r.Acceptance.hasIDs() {
			continue
		}
		ids := make([]string, 0, len(r.Acceptance))
		for _, c := range r.Acceptance {
			if c.ID != "" {
				ids = append(ids, c.ID)
			}
		}
		if len(ids) > 0 {
			sort.Strings(ids) // reporte reproducible
			out[r.ID] = ids
		}
	}
	return out
}

// requirementsOf lee los requisitos de una feature. Lectura QUIETA: sin spec o
// con spec inválida, la lista vacía — el artefacto ya lo validó su propio gate y
// acá no es lugar para volver a quejarse de él.
func requirementsOf(projectDir, feature string) []requirement {
	path := findArtifact(filepath.Join(projectDir, "specforge"), feature, "requirements.json")
	if path == "" {
		return nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var rf requirementsFile
	if json.Unmarshal(data, &rf) != nil {
		return nil
	}
	return rf.Requirements
}

// scenarioReasons exige verificación por criterio y nombra EL CRITERIO al
// rechazar, no el requisito. Que el mensaje diga `R5.2` y no `R5` es la mitad
// del valor: "R5 no tiene test" manda a releer cinco criterios para encontrar
// cuál falta.
//
// `method` es el `verification` del requisito (RM-C4): con `test` se exige un
// test anclado; con cualquier otro, una evidencia declarada.
func scenarioReasons(projectDir, req, method string, wanted []string, info traceReq) []string {
	var reasons []string
	for _, id := range wanted {
		sc, ok := info.Scenarios[id]

		if method != "test" {
			// R9: el requisito declaró que NO se verifica con un test, así que
			// exigirle uno sería absurdo — pero dejarlo pasar sin nada es
			// exactamente el agujero de C4. Se le exige la evidencia que dijo
			// que iba a tener.
			if !ok || sc.Evidence == nil {
				reasons = append(reasons, fmt.Sprintf(
					"%s: verification is %q but no evidence is declared — record it under "+
						"trace.json requirements.%s.scenarios.%s.evidence", id, method, req, id))
				continue
			}
			reasons = append(reasons, evidenceReasons(projectDir, id, method, *sc.Evidence)...)
			continue
		}

		if !ok || len(sc.Test) == 0 {
			reasons = append(reasons, fmt.Sprintf(
				"%s: names no test (verification contract unmet) — anchor one under trace.json requirements.%s.scenarios.%s",
				id, req, id))
			continue
		}
		for _, tref := range sc.Test {
			if ok, why := checkAnchor(projectDir, tref); !ok {
				reasons = append(reasons, fmt.Sprintf("%s: test %q does not resolve (%s)", id, tref, why))
			}
		}
	}
	// Un escenario en el trace que ya no existe en el spec: el criterio se
	// retiró y su ancla quedó colgando. Es exactamente el síntoma que delata un
	// renumerado — el ancla sigue apuntando a un id que ahora nombra otra cosa,
	// o nada.
	declared := map[string]bool{}
	for _, id := range wanted {
		declared[id] = true
	}
	orphans := make([]string, 0, len(info.Scenarios))
	for id := range info.Scenarios {
		if !declared[id] {
			orphans = append(orphans, id)
		}
	}
	sort.Strings(orphans)
	for _, id := range orphans {
		reasons = append(reasons, fmt.Sprintf(
			"%s: trace anchors a scenario that %s no longer declares — the criterion was removed, or its ids were renumbered",
			id, req))
	}
	return reasons
}

// evidenceReasons valida UNA evidencia contra lo que el requisito prometió.
//
// El chequeo que importa es el primero: si el requisito declara `benchmark` y la
// evidencia dice `test`, alguien anotó lo que tenía a mano en vez de lo que hacía
// falta. Sin ese contraste, `evidence` sería un campo de texto libre que se
// llena para pasar el gate — o sea, nada.
func evidenceReasons(projectDir, id, method string, e scenarioEvidence) []string {
	var reasons []string

	if e.Kind != method {
		reasons = append(reasons, fmt.Sprintf(
			"%s: evidence kind %q does not match the declared verification %q", id, e.Kind, method))
	}

	ref := strings.TrimSpace(e.Ref)
	switch {
	case ref == "":
		reasons = append(reasons, fmt.Sprintf("%s: evidence has no ref — nothing to open", id))
	case !isSourceURL(ref) && !fileExists(filepath.Join(projectDir, filepath.FromSlash(ref))):
		reasons = append(reasons, fmt.Sprintf(
			"%s: evidence ref %q does not exist — evidence nobody can open is an assertion", id, ref))
	}

	// La fecha no se compara contra nada todavía: la frescura de la verificación
	// manual es su propio ítem (DL-15). Acá se exige que EXISTA y esté bien
	// formada, porque una evidencia sin fecha no se puede evaluar después.
	if !capturedRe.MatchString(e.Recorded) {
		reasons = append(reasons, fmt.Sprintf(
			"%s: evidence recorded %q must be YYYY-MM-DD — undated evidence cannot be aged", id, e.Recorded))
	}
	return reasons
}

// buildAuditEntry valida los verdicts y computa el overall (función pura).
func buildAuditEntry(phase string, verdicts []ruleVerdict) (auditEntry, report) {
	var rep report
	if strings.TrimSpace(phase) == "" {
		rep.errorf("phase is required")
	}
	if len(verdicts) == 0 {
		rep.errorf("at least one rule verdict is required")
	}
	overall := "pass"
	for i, v := range verdicts {
		if strings.TrimSpace(v.Rule) == "" {
			rep.errorf("verdict %d: empty rule", i+1)
		}
		if !verdictResults[v.Result] {
			rep.errorf("verdict %d (%s): result must be pass|fail, got %q", i+1, v.Rule, v.Result)
		}
		if v.Result == "fail" {
			overall = "fail"
		}
	}
	return auditEntry{
		Phase:    phase,
		At:       nowUTC(),
		Overall:  overall,
		Verdicts: verdicts,
	}, rep
}

// verifyCitations (A8) aplica la regla que ordena el backlog de integridad:
// "ninguna afirmación del modelo entra al estado sin verificación mecánica o
// sin marcarse como no-verificada". Una citation del juez ES una afirmación
// ("esto está en el artefacto"); acá la chequeamos con un substring-match
// whitespace-normalizado contra el .json canónico Y el .md renderizado de la
// fase (el juez suele citar del render). Barato y determinista.
func verifyCitations(projectDir, feature string, entry *auditEntry) {
	rel, ok := artifactFileForPhase(entry.Phase)
	if !ok {
		return // fase sin artefacto (lane, wave-N) → nada contra qué verificar
	}
	base := filepath.Join(projectDir, "specforge", "features", feature)
	var haystacks []string
	for _, p := range []string{rel, strings.TrimSuffix(rel, ".json") + ".md"} {
		if data, err := os.ReadFile(filepath.Join(base, p)); err == nil {
			haystacks = append(haystacks, normalizeWS(string(data)))
		}
	}
	if len(haystacks) == 0 {
		return // artefacto ausente: lo reporta el flujo de gates, no esta marca
	}
	for i := range entry.Verdicts {
		v := &entry.Verdicts[i]
		if strings.TrimSpace(v.Citation) == "" {
			continue
		}
		needle := normalizeWS(v.Citation)
		v.CitationCheck = "not-found"
		for _, h := range haystacks {
			if strings.Contains(h, needle) {
				v.CitationCheck = "verified"
				break
			}
		}
	}
}

// normalizeWS colapsa todo whitespace (saltos de línea, tabs, espacios
// repetidos) a un espacio simple: una citation que envuelve línea en el .md
// sigue matcheando.
func normalizeWS(s string) string { return strings.Join(strings.Fields(s), " ") }

// appendAuditEntry lee el ledger existente (si hay), apendea la entrada y lo
// reescribe canónico.
func appendAuditEntry(projectDir, feature string, entry auditEntry) int {
	path := filepath.Join(projectDir, "specforge", "features", feature, "audit.json")

	ledger := auditLedger{Feature: feature}
	if data, err := os.ReadFile(path); err == nil {
		// Si existe pero no parsea, preferimos fallar antes que pisar el historial.
		if err := json.Unmarshal(data, &ledger); err != nil {
			fmt.Fprintf(os.Stderr, "sf gate: existing audit.json is invalid (%v) — not overwriting\n", err)
			return 1
		}
	}
	ledger.Feature = feature
	ledger.Entries = append(ledger.Entries, entry)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: %v\n", err)
		return 1
	}
	out, err := json.MarshalIndent(ledger, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf gate: cannot write %s (%v)\n", path, err)
		return 1
	}
	return 0
}
