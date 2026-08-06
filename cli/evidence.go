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

// ----------------------------------------------------------------------------
// `sf evidence` — el circuito de la verificación NO automática (DL-15).
//
// EL HUECO. `RM-C4` shipeó la DEMANDA sin la OFERTA: el veredicto exige
// `evidence` cuando el requisito declara `verification != test`, y para darla
// había que escribir a mano el documento entero del `trace.json` — sin nada que
// dijera qué criterios la necesitan, cuáles ya la tienen, ni cuáles envejecieron.
// El camino existía (`sf save trace`), pero no había circuito.
//
// Y un circuito es lo que hace falta, porque este es el único tramo del pipeline
// donde el verificador es UNA PERSONA. Un test lo corre la máquina y el resultado
// es un exit code; un benchmark, una auditoría de seguridad o una verificación
// manual las hace alguien, en su tiempo, fuera del repo. Sin checklist, ese
// alguien no sabe qué le toca; sin frescura, lo que verificó hace ocho meses vale
// lo mismo que lo de ayer.
//
// LAS TRES PARTES, y cada una tapa un agujero distinto:
//
//	checklist  QUÉ hay que verificar a mano  → derivado del trace, no escrito a mano
//	record     CÓMO se re-ingiere            → un comando, no editar el artefacto
//	frescura   CUÁNDO deja de valer          → `evidence_max_age_days`, opt-in
//
// LA FRESCURA ES DISTINTA DE TODO LO DEMÁS EN EL PRODUCTO, y conviene decirlo:
// es el único chequeo que puede poner en rojo un proyecto que nadie tocó. El
// tiempo pasa solo. Por eso es opt-in y por eso, cuando bloquea, el mensaje dice
// que el problema es la EDAD y no la ausencia — si dijera "falta evidencia",
// alguien la re-anotaría con la fecha de hoy sin volver a verificar nada, que es
// exactamente el fraude que el circuito viene a evitar.
// ----------------------------------------------------------------------------

// evidenceItem es una línea de la checklist: un criterio que NO se verifica con
// un test, y en qué estado está su evidencia.
type evidenceItem struct {
	Requirement string `json:"requirement"`
	Scenario    string `json:"scenario"`
	Text        string `json:"text,omitempty"`
	// Method es el `verification` declarado (benchmark, audit, manual, …). Es lo
	// que la evidencia tiene que igualar en su `kind`.
	Method string `json:"method"`
	// State: missing | recorded | stale. Tres y no dos, porque "la tengo pero
	// venció" pide una acción distinta de "nunca la tuve": una es re-verificar,
	// la otra es verificar por primera vez.
	State    string `json:"state"`
	Ref      string `json:"ref,omitempty"`
	Recorded string `json:"recorded,omitempty"`
	AgeDays  int    `json:"age_days,omitempty"`
}

// evidenceChecklist recorre los requisitos de la feature y devuelve, ordenada,
// la lista de criterios que dependen de una persona.
//
// Los requisitos con `verification: test` no aparecen: su verificación ya la
// exige el contrato y la corre la máquina. Mezclarlos acá convertiría la
// checklist en una lista de todo, que es una lista que nadie usa.
func evidenceChecklist(projectDir, feature string, maxAge int, today time.Time) []evidenceItem {
	criteria := acceptanceCriteriaOf(projectDir, feature)
	tf := loadTraceFile(projectDir, feature)

	var out []evidenceItem
	for _, r := range requirementsOf(projectDir, feature) {
		method := verificationOf(r) // ausente ⇒ test (fail-closed, RM-C4)
		if method == "test" {
			continue
		}
		texts := map[string]string{}
		for _, c := range r.Acceptance {
			texts[c.ID] = c.Text
		}
		for _, id := range criteria[r.ID] {
			it := evidenceItem{
				Requirement: r.ID, Scenario: id, Text: texts[id],
				Method: method, State: "missing",
			}
			if sc, ok := tf.Requirements[r.ID].Scenarios[id]; ok && sc.Evidence != nil {
				it.Ref = sc.Evidence.Ref
				it.Recorded = sc.Evidence.Recorded
				it.State = "recorded"
				if age, ok := evidenceAgeDays(sc.Evidence.Recorded, today); ok {
					it.AgeDays = age
					if maxAge > 0 && age > maxAge {
						it.State = "stale"
					}
				}
			}
			out = append(out, it)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Requirement != out[j].Requirement {
			return out[i].Requirement < out[j].Requirement
		}
		return out[i].Scenario < out[j].Scenario
	})
	return out
}

// evidenceAgeDays: días entre la fecha registrada y hoy. `ok=false` si la fecha
// no parsea — ese caso ya lo rechaza `evidenceReasons` en el gate, y acá no
// corresponde inventarle una edad.
func evidenceAgeDays(recorded string, today time.Time) (int, bool) {
	d, err := time.Parse("2006-01-02", strings.TrimSpace(recorded))
	if err != nil {
		return 0, false
	}
	return int(today.Sub(d).Hours() / 24), true
}

// evidenceMaxAgeDays lee `verification.evidence_max_age_days`. Cero (o ausente)
// significa APAGADO: la evidencia no vence.
func evidenceMaxAgeDays(projectDir string) int {
	return verificationOpts(projectDir).EvidenceMaxAgeDays
}

// staleEvidenceReasons es lo que el gate del veredicto consume. Vacío ⇒ nada
// venció (o la frescura está apagada).
func staleEvidenceReasons(projectDir, feature string) []string {
	maxAge := evidenceMaxAgeDays(projectDir)
	if maxAge <= 0 {
		return nil
	}
	var reasons []string
	for _, it := range evidenceChecklist(projectDir, feature, maxAge, time.Now().UTC()) {
		if it.State != "stale" {
			continue
		}
		// El mensaje nombra la EDAD, no la ausencia. Si dijera "falta evidencia",
		// la salida barata sería re-anotar la misma con la fecha de hoy sin
		// volver a verificar nada.
		reasons = append(reasons, fmt.Sprintf(
			"%s: evidence is %d day(s) old and the project's limit is %d — it was verified against an "+
				"older system, so re-verify and record the new result (do not just re-date the old one)",
			it.Scenario, it.AgeDays, maxAge))
	}
	return reasons
}

// ----------------------------------------------------------------------------
// El comando.
// ----------------------------------------------------------------------------

func runEvidence(args []string) int {
	if len(args) == 0 {
		evidenceUsage()
		return 2
	}
	switch args[0] {
	case "checklist":
		return runEvidenceChecklist(args[1:])
	case "record":
		return runEvidenceRecord(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "sf evidence: unknown subcommand %q\n", args[0])
		evidenceUsage()
		return 2
	}
}

func evidenceUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf evidence checklist --feature=NAME [--json] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf evidence record --feature=NAME --scenario=R1.1 --kind=KIND --ref=PATH|URL [--recorded=YYYY-MM-DD] [project_dir]")
}

func runEvidenceChecklist(args []string) int {
	projectDir, feature := ".", ""
	asJSON := false
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf evidence: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf evidence checklist: --feature=NAME is required")
		return 2
	}

	maxAge := evidenceMaxAgeDays(projectDir)
	items := evidenceChecklist(projectDir, feature, maxAge, time.Now().UTC())

	if asJSON {
		if items == nil {
			items = []evidenceItem{} // siempre el array: distinguir "vacío" de "no vino"
		}
		out, _ := json.MarshalIndent(map[string]any{
			"feature": feature, "max_age_days": maxAge, "items": items,
		}, "", "  ")
		fmt.Println(string(out))
		return 0
	}

	if len(items) == 0 {
		fmt.Printf("%s: every requirement is verified by test — nothing needs a human.\n", feature)
		return 0
	}

	fmt.Printf("Evidence checklist — %s (%d criterion/criteria verified by a human)\n\n", feature, len(items))
	cols := []string{"criterion", "method", "state", "recorded", "ref"}
	var rows [][]string
	pending := 0
	for _, it := range items {
		if it.State != "recorded" {
			pending++
		}
		state := it.State
		if it.State == "stale" {
			state = fmt.Sprintf("STALE (%dd)", it.AgeDays)
		}
		rows = append(rows, []string{it.Scenario, it.Method, state, orDash(it.Recorded), orDash(it.Ref)})
	}
	renderTable(cols, rows)

	if pending > 0 {
		// `missing` y `stale` se cuentan aparte porque piden ACCIONES DISTINTAS, y
		// un pie que dijera "registrá estos" para los dos invitaría a re-anotar la
		// evidencia vencida con la fecha de hoy sin volver a verificar nada — el
		// fraude exacto que la frescura viene a evitar.
		missing, stale := 0, 0
		for _, it := range items {
			switch it.State {
			case "missing":
				missing++
			case "stale":
				stale++
			}
		}
		fmt.Println()
		if missing > 0 {
			fmt.Printf("%d never verified — verify, then record:\n", missing)
			fmt.Println("  sf evidence record --feature=" + feature + " --scenario=<id> --kind=<method> --ref=<path|url>")
		}
		if stale > 0 {
			fmt.Printf("%d past the project's age limit — RE-VERIFY against the current system and record\n", stale)
			fmt.Println("  the new result. Re-dating the old evidence would pass the gate without verifying")
			fmt.Println("  anything, which is the failure this check exists to catch.")
		}
		return 1
	}
	fmt.Println("\nOK: every human-verified criterion has current evidence.")
	return 0
}

// runEvidenceRecord escribe UNA evidencia en el trace.
//
// Existe para que re-ingerir no sea "editá el trace.json a mano": ese archivo es
// estado autoritativo, el hook lo protege, y pedirle a alguien que reescriba un
// documento entero para anotar tres campos es cómo se termina anotando cualquier
// cosa con tal de que pase.
func runEvidenceRecord(args []string) int {
	projectDir, feature, scenario, kind, ref, recorded := ".", "", "", "", "", ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--scenario="):
			scenario = strings.TrimPrefix(a, "--scenario=")
		case strings.HasPrefix(a, "--kind="):
			kind = strings.TrimPrefix(a, "--kind=")
		case strings.HasPrefix(a, "--ref="):
			ref = strings.TrimPrefix(a, "--ref=")
		case strings.HasPrefix(a, "--recorded="):
			recorded = strings.TrimPrefix(a, "--recorded=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf evidence: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" || scenario == "" || kind == "" || ref == "" {
		fmt.Fprintln(os.Stderr, "sf evidence record: --feature, --scenario, --kind and --ref are all required")
		return 2
	}
	if recorded == "" {
		recorded = time.Now().UTC().Format("2006-01-02")
	}

	// El requisito dueño del criterio sale del id (`R5.2` → `R5`), no de un flag:
	// pedir los dos permitiría declararlos inconsistentes.
	req, _, found := strings.Cut(scenario, ".")
	if !found || req == "" {
		fmt.Fprintf(os.Stderr, "sf evidence record: %q is not a criterion id (expected <requirement>.<n>, e.g. R5.2)\n", scenario)
		return 2
	}

	tracePath := findArtifact(filepath.Join(projectDir, "specforge"), feature, "trace.json")
	if tracePath == "" {
		fmt.Fprintf(os.Stderr, "sf evidence record: no trace.json for %q\n", feature)
		return 4
	}
	data, err := os.ReadFile(tracePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf evidence record: %v\n", err)
		return 1
	}
	var tf traceFile
	if err := json.Unmarshal(data, &tf); err != nil {
		fmt.Fprintf(os.Stderr, "sf evidence record: invalid trace.json (%v)\n", err)
		return 1
	}

	// El criterio tiene que EXISTIR en el spec. Sin este chequeo se podría anclar
	// evidencia a un id inventado, y una evidencia que no cubre ningún criterio
	// declarado es exactamente lo que el gate no puede detectar después.
	declared := acceptanceCriteriaOf(projectDir, feature)
	ok := false
	for _, id := range declared[req] {
		if id == scenario {
			ok = true
		}
	}
	if !ok {
		fmt.Fprintf(os.Stderr, "sf evidence record: %s declares no criterion %q — evidence must attach to a declared criterion\n", req, scenario)
		return 2
	}

	if tf.Requirements == nil {
		tf.Requirements = map[string]traceReq{}
	}
	entry := tf.Requirements[req]
	if entry.Scenarios == nil {
		entry.Scenarios = map[string]traceScenario{}
	}
	sc := entry.Scenarios[scenario]
	sc.Evidence = &scenarioEvidence{Kind: kind, Ref: ref, Recorded: recorded}
	entry.Scenarios[scenario] = sc
	tf.Requirements[req] = entry
	if tf.SchemaVersion == "" {
		tf.SchemaVersion = schemaVersionCurrent
	}
	if tf.Feature == "" {
		tf.Feature = feature
	}

	// Se valida ANTES de escribir, con las mismas reglas que aplica el gate: si
	// el `kind` no coincide con lo declarado o el `ref` no resuelve, esto se
	// rechaza acá y no dentro de un veredicto tres pasos después.
	method := "test"
	for _, r := range requirementsOf(projectDir, feature) {
		if r.ID == req {
			method = verificationOf(r)
		}
	}
	if reasons := evidenceReasons(projectDir, scenario, method, *sc.Evidence); len(reasons) > 0 {
		fmt.Fprintln(os.Stderr, "sf evidence record: refused —")
		for _, r := range reasons {
			fmt.Fprintf(os.Stderr, "  - %s\n", r)
		}
		return 2
	}

	out, err := json.MarshalIndent(tf, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf evidence record: %v\n", err)
		return 1
	}
	if err := os.WriteFile(tracePath, append(out, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf evidence record: %v\n", err)
		return 1
	}
	fmt.Printf("recorded %s evidence for %s (%s) → %s\n", kind, scenario, recorded, filepath.ToSlash(tracePath))
	return 0
}
