package main

import (
	_ "embed" // habilita //go:embed para el template del workflow
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf verify` — R2 de EVALUACION-PLATAFORMA §3: la capa que nadie puede esquivar.
//
// Todo el enforcement local (hook deny, gate refuse) vive DENTRO del arnés: si
// el usuario apagó los hooks, si el agente es otro, o si el arnés no soporta
// hooks, esas garantías no corren. `sf verify` es el agregador con EXIT CODE
// pensado para CI: corre en otra máquina, con otro trust, en cada PR — y ahí
// ni el modelo ni el humano apurado pueden esquivarlo.
//
// Qué agrega (nada nuevo: compone verificaciones que ya existen):
//   1. ledger íntegro        → ledgerProblemsAll (integrity.go, R1)
//   2. schemas válidos       → los check* de cada artefacto (save.go los usa)
//   3. sin artefactos stale  → computeArtifactStates (stale.go)
//   4. trace sin drift       → checkFeatureDrift (doctor.go)
//   5. verdict listo         → verdictPreconditions (gate.go): contrato de
//      verificación + check verde y FRESCO (solo cuando hay verdict en juego)
//
// Además: `sf verify --init-ci` scaffoldea el workflow de GitHub Actions que
// corre `sf verify` en cada PR — el mapeo "verdict ↔ PR review" deja de ser
// prosa aspiracional en AGENT.md y pasa a ser un job rojo/verde.
// ----------------------------------------------------------------------------

// verifyCheck es el resultado de UNA verificación del agregado.
type verifyCheck struct {
	Name     string   `json:"name"`
	OK       bool     `json:"ok"`
	Problems []string `json:"problems,omitempty"`
	Note     string   `json:"note,omitempty"` // contexto ("skipped: no verdict in flight")
}

// verifyReport es el JSON de `sf verify --json` (y la fuente del render humano).
type verifyReport struct {
	OK     bool          `json:"ok"`
	Checks []verifyCheck `json:"checks"`
}

func runVerify(args []string) int {
	projectDir := "."
	feature := ""
	asJSON := false
	initCI := false
	for _, a := range args {
		switch {
		case a == "--json":
			asJSON = true
		case a == "--init-ci":
			initCI = true
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf verify: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	if initCI {
		return verifyInitCI(projectDir)
	}

	rep := buildVerifyReport(projectDir, feature)

	if asJSON {
		out, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "sf verify: marshal failed (%v)\n", err)
			return 1
		}
		fmt.Println(string(out))
	} else {
		printVerifyReport(rep)
	}

	if !rep.OK {
		return 5 // mismo código que status --artifacts / recover: CI-reaccionable
	}
	return 0
}

// buildVerifyReport compone las 5 verificaciones. Separado del print/exit para
// poder testearlo como función.
func buildVerifyReport(projectDir, featureFilter string) verifyReport {
	rep := verifyReport{OK: true}
	add := func(c verifyCheck) {
		if len(c.Problems) == 0 {
			c.OK = true
		} else {
			rep.OK = false
		}
		rep.Checks = append(rep.Checks, c)
	}

	// Sin specforge/ no hay nada que verificar: exit 0 honesto (un repo que no
	// adoptó SpecForge no debe romper su CI por instalar el workflow).
	if _, ok := specforgeRoot(projectDir); !ok {
		return verifyReport{OK: true, Checks: []verifyCheck{
			{Name: "specforge", OK: true, Note: "no specforge/ directory — nothing to verify"},
		}}
	}

	// features.json es la raíz de todo: ilegible/inválido = fail duro.
	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		return verifyReport{OK: false, Checks: []verifyCheck{
			{Name: "ledger", Problems: []string{"cannot read/parse specforge/features.json: " + err.Error()}},
		}}
	}

	// 1) Cadena de integridad (R1) sobre TODAS las features.
	add(verifyCheck{Name: "ledger", Problems: ledgerProblemsAll(ff)})

	// 2) Schemas de todos los artefactos existentes.
	add(verifyCheck{Name: "schemas", Problems: verifySchemas(projectDir, ff)})

	// 3) Staleness (features activas: las done ya no se gobiernan acá).
	add(verifyCheck{Name: "staleness", Problems: verifyStaleness(projectDir, ff)})

	// 4) Drift del trace (toda feature CON trace, incluidas las done: el legado
	//    anclado no debe pudrirse en silencio — §2.9 punto 5).
	add(verifyCheck{Name: "trace-drift", Problems: verifyTraceDrift(projectDir, ff)})

	// 5) Verdict-readiness: el chequeo caro (corre codeHash). Solo cuando hay un
	//    verdict EN JUEGO: --feature explícito, o la feature activa está en fase
	//    verdict. Fuera de eso, exigir "check fresco" sería un falso positivo
	//    permanente (cualquier commit posterior mueve el hash).
	target, note := verdictTarget(ff, projectDir, featureFilter)
	if target == "" {
		add(verifyCheck{Name: "verdict", Note: note})
	} else {
		var probs []string
		for _, r := range verdictPreconditions(projectDir, target) {
			probs = append(probs, target+": "+r)
		}
		// RM-C2: los incumplimientos no bloqueantes no rompen CI (para eso está
		// `blocking_priorities`), pero tienen que verse. Van a la nota:
		// visibilidad sin bloqueo, el mismo criterio que el resto del sistema.
		note := "feature: " + target
		if warns := verdictWarnings(projectDir, target); len(warns) > 0 {
			note += fmt.Sprintf(" — %d non-blocking contract issue(s): %s",
				len(warns), strings.Join(warns, "; "))
		}
		add(verifyCheck{Name: "verdict", Problems: probs, Note: note})
	}

	return rep
}

// verdictTarget decide sobre qué feature corre el chequeo de verdict, y si
// ninguna, la nota que lo explica.
func verdictTarget(ff featuresFile, projectDir, featureFilter string) (string, string) {
	if featureFilter != "" {
		return featureFilter, ""
	}
	st := computeCurrentState(ff, projectDir)
	if st.Feature != "" && st.Phase == "verdict" {
		return st.Feature, ""
	}
	return "", "skipped — no verdict in flight (use --feature=NAME to force)"
}

// verifySchemas valida el JSON de cada artefacto EXISTENTE contra su schema
// (los mismos check* que usa `sf save` al escribir). Detecta artefactos que
// entraron al repo por fuera del CLI (merge a mano, edición directa).
func verifySchemas(projectDir string, ff featuresFile) []string {
	var probs []string
	sf := filepath.Join(projectDir, "specforge")

	// Artefactos a nivel proyecto.
	probs = append(probs, validateArtifactFile("constitution", filepath.Join(sf, "constitution.json"))...)
	probs = append(probs, validateArtifactFile("domain", domainPath(projectDir)+".json")...)

	// Artefactos por feature.
	perFeature := []struct{ kind, rel string }{
		{"requirements", "requirements.json"},
		{"design", "design.json"},
		{"tasks", "tasks.json"},
		{"plan", filepath.Join("progress", "plan.json")},
		{"review", "review.json"},
		{"trace", "trace.json"},
	}
	for i := range ff.Features {
		fdir := filepath.Join(sf, "features", ff.Features[i].Name)
		for _, a := range perFeature {
			probs = append(probs, validateArtifactFile(a.kind, filepath.Join(fdir, a.rel))...)
		}
	}
	return probs
}

// validateArtifactFile decodifica y valida UN archivo si existe (ausente no es
// problema: la fase puede no haber llegado). Devuelve los errores prefijados
// con la ruta, para que el reporte de CI diga exactamente dónde mirar.
func validateArtifactFile(kind, absPath string) []string {
	raw, err := os.ReadFile(absPath)
	if err != nil {
		return nil // no existe → no es un problema de schema
	}

	var rep report
	ok := true
	switch kind {
	case "constitution":
		var v constitutionFile
		if ok = jsonOK(raw, &v, &rep); ok {
			checkConstitution(v, &rep)
		}
	case "domain":
		var v domainFile
		if ok = jsonOK(raw, &v, &rep); ok {
			checkDomain(v, &rep)
		}
	case "requirements":
		var v requirementsFile
		if ok = jsonOK(raw, &v, &rep); ok {
			checkRequirements(v, &rep)
		}
	case "design":
		var v designFile
		if ok = jsonOK(raw, &v, &rep); ok {
			checkDesign(v, &rep)
		}
	case "tasks":
		var v tasksFile
		if ok = jsonOK(raw, &v, &rep); ok {
			checkTasks(v, &rep)
		}
	case "plan":
		var v planFile
		if ok = jsonOK(raw, &v, &rep); ok {
			checkPlanInternal(v, &rep)
		}
	case "review":
		var v reviewFile
		if ok = jsonOK(raw, &v, &rep); ok {
			checkReview(v, &rep)
		}
	case "trace":
		var v traceFile
		if ok = jsonOK(raw, &v, &rep); ok {
			checkTrace(v, &rep)
		}
	}

	out := make([]string, 0, len(rep.errors))
	for _, e := range rep.errors {
		out = append(out, filepath.ToSlash(absPath)+": "+e)
	}
	return out
}

// jsonOK desempaqueta el JSON al tipo concreto; si falla lo registra como error
// de schema (un artefacto ilegible ES un problema, a diferencia de uno ausente).
func jsonOK(raw []byte, v any, rep *report) bool {
	if err := json.Unmarshal(raw, v); err != nil {
		rep.errorf("invalid JSON (%v)", err)
		return false
	}
	return true
}

// verifyStaleness reporta artefactos stale de las features activas (reusa el
// stale model completo: silent edits, upstream reabierto, propagación).
func verifyStaleness(projectDir string, ff featuresFile) []string {
	var probs []string
	for i := range ff.Features {
		f := &ff.Features[i]
		if f.Status == "done" {
			continue
		}
		for _, s := range computeArtifactStates(f, projectDir) {
			if s.State == "stale" {
				probs = append(probs, fmt.Sprintf("%s/%s: stale — %s", f.Name, s.Phase, s.Reason))
			}
		}
	}
	return probs
}

// verifyTraceDrift corre el chequeo estático de anchors sobre toda feature que
// tenga trace.json (activas Y archivadas — el trace de una feature done sigue
// afirmando "este código implementa esta spec", y eso tiene que seguir siendo
// verdad o detectarse).
func verifyTraceDrift(projectDir string, ff featuresFile) []string {
	var probs []string
	sf := filepath.Join(projectDir, "specforge")
	for i := range ff.Features {
		name := ff.Features[i].Name
		tracePath := findArtifact(sf, name, "trace.json")
		if tracePath == "" {
			continue
		}
		for _, d := range checkFeatureDrift(projectDir, name, tracePath, "") {
			probs = append(probs, name+": "+d)
		}
	}
	return probs
}

// printVerifyReport imprime el agregado con ✓/✗ por check.
func printVerifyReport(rep verifyReport) {
	fmt.Print("SpecForge — verify\n\n")
	for _, c := range rep.Checks {
		mark := "✓"
		if !c.OK {
			mark = "✗"
		}
		line := fmt.Sprintf("  %s %s", mark, c.Name)
		if c.Note != "" {
			line += "  (" + c.Note + ")"
		}
		fmt.Println(line)
		for _, p := range c.Problems {
			fmt.Printf("      - %s\n", p)
		}
	}
	if rep.OK {
		fmt.Println("\nOK — the ledger, schemas, staleness and trace all verify.")
	} else {
		fmt.Println("\nFAIL — fix the problems above (see `sf recover` for a plan).")
	}
}

// ── --init-ci: scaffold del workflow de GitHub Actions ───────────────────────

//go:embed templates/verify-workflow.yml
var verifyWorkflowYML string

// verifyInitCI escribe .github/workflows/specforge-verify.yml. Se rehúsa a
// pisar uno existente (puede tener personalizaciones del usuario).
func verifyInitCI(projectDir string) int {
	path := filepath.Join(projectDir, ".github", "workflows", "specforge-verify.yml")
	if fileExists(path) {
		fmt.Fprintf(os.Stderr, "sf verify --init-ci: %s already exists — not overwriting\n", path)
		return 4
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf verify --init-ci: %v\n", err)
		return 1
	}
	if err := os.WriteFile(path, []byte(verifyWorkflowYML), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf verify --init-ci: cannot write %s (%v)\n", path, err)
		return 1
	}
	fmt.Printf("wrote %s\n", filepath.ToSlash(path))
	fmt.Println("Review the `Install sf` step — it needs to match how your project distributes the binary.")
	return 0
}
