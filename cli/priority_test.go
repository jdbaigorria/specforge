package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// RM-C2 — `priority` tipada + `blocking_priorities` (R5).
//
// El antídoto a la fatiga de gate: un gate que bloquea por todo se termina
// salteando por todo. No por deshonestidad — un bloqueo que no discrimina deja
// de informar. Graduar la severidad es lo que hace que bloquear vuelva a
// significar algo.
//
// Las dos decisiones que sostienen el diseño son las mismas: FAIL-CLOSED.
// `priority` ausente ⇒ `must`, y `blocking_priorities` ausente o vacío ⇒ todo
// bloquea. Si fuera al revés, omitir un campo sería la forma más barata de bajar
// el listón, y esas formas un agente bajo presión de contexto las encuentra solo.
// ----------------------------------------------------------------------------

func TestPriorityDefaults(t *testing.T) {
	t.Run("ausente ⇒ must (fail-closed)", func(t *testing.T) {
		if got := priorityOf(requirement{ID: "R1"}); got != "must" {
			t.Errorf("priorityOf = %q, want must", got)
		}
	})

	t.Run("las tres del enum se respetan", func(t *testing.T) {
		for _, p := range []string{"must", "should", "could"} {
			if got := priorityOf(requirement{ID: "R1", Priority: p}); got != p {
				t.Errorf("priorityOf(%q) = %q", p, got)
			}
		}
	})

	t.Run("un valor fuera del enum no valida", func(t *testing.T) {
		rf := requirementsFile{Feature: "f", Requirements: []requirement{
			{ID: "R1", EarsType: "ubiquitous", Behavior: "b", Priority: "critical"},
		}}
		var rep report
		checkRequirements(rf, &rep)
		found := false
		for _, e := range rep.errors {
			if strings.Contains(e, "invalid priority") {
				found = true
			}
		}
		if !found {
			t.Errorf("want error de priority inválida, got %v", rep.errors)
		}
	})
}

func TestBlockingPrioritiesDefaults(t *testing.T) {
	t.Run("sin constitución bloquea todo", func(t *testing.T) {
		b := blockingPriorities(t.TempDir())
		for _, p := range []string{"must", "should", "could"} {
			if !b[p] {
				t.Errorf("%s debería bloquear por default", p)
			}
		}
	})

	t.Run("lista VACÍA bloquea todo, no nada", func(t *testing.T) {
		// Es la decisión fail-closed que más fácil se hace al revés: un `[]`
		// tecleado por error desactivaría el gate entero en silencio.
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "specforge/constitution.json"),
			`{"identity_md":"x","verification":{"blocking_priorities":[]}}`)
		b := blockingPriorities(dir)
		if !b["must"] || !b["should"] || !b["could"] {
			t.Errorf("una lista vacía debe leerse como 'todo bloquea', got %v", b)
		}
	})

	t.Run("una lista explícita se respeta", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "specforge/constitution.json"),
			`{"identity_md":"x","verification":{"blocking_priorities":["must"]}}`)
		b := blockingPriorities(dir)
		if !b["must"] || b["should"] || b["could"] {
			t.Errorf("blocking=%v, want sólo must", b)
		}
	})

	t.Run("un valor fuera del enum no valida", func(t *testing.T) {
		cf := constitutionFile{IdentityMD: "x",
			Verification: &verificationConfig{BlockingPriorities: []string{"must", "high"}}}
		var rep report
		checkConstitution(cf, &rep)
		if len(rep.errors) == 0 {
			t.Error(`"high" no es una prioridad — debería rechazarse en vez de ignorarse`)
		}
	})
}

// makePriorityProject arma una feature con UN requisito de la prioridad dada,
// cuyo trace no nombra ningún test: el incumplimiento que R5 gradúa.
func makePriorityProject(t *testing.T, priority, blocking string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`)
	if blocking != "" {
		writeFile(t, filepath.Join(dir, "specforge/constitution.json"),
			`{"identity_md":"x","verification":{"blocking_priorities":`+blocking+`}}`)
	}
	writeFile(t, filepath.Join(dir, "src/app.go"), "package src\n\nfunc Run() {}\n")
	writeFile(t, filepath.Join(dir, "specforge/features/f/requirements.json"),
		`{"feature":"f","requirements":[{"id":"R1","ears_type":"ubiquitous","behavior":"run",
		  "priority":"`+priority+`","acceptance":["corre"]}]}`)
	writeFile(t, filepath.Join(dir, "specforge/features/f/trace.json"),
		`{"feature":"f","requirements":{"R1":{"code":["src/app.go:Run"],"test":[],"status":"no-test"}}}`)
	return dir
}

// TestPriorityGatesTheContract — R5, el corazón de RM-C2.
func TestPriorityGatesTheContract(t *testing.T) {
	// contractIssues descarta las razones de proceso (no hay `check run` en
	// estos fixtures) y deja sólo las del contrato de verificación.
	contractIssues := func(dir string) (blocking, warning []string) {
		for _, is := range verdictIssues(dir, "f") {
			if strings.Contains(is.Reason, "check run") || strings.Contains(is.Reason, "test result") {
				continue
			}
			if is.Blocking {
				blocking = append(blocking, is.Reason)
			} else {
				warning = append(warning, is.Reason)
			}
		}
		return
	}

	t.Run("un could sin test NO bloquea si el proyecto lo declaró", func(t *testing.T) {
		dir := makePriorityProject(t, "could", `["must"]`)
		blk, warn := contractIssues(dir)
		if len(blk) != 0 {
			t.Errorf("want 0 bloqueantes, got %v", blk)
		}
		if len(warn) != 1 {
			t.Fatalf("want 1 advertencia, got %v", warn)
		}
		// Y la advertencia dice POR QUÉ no bloqueó, si no parece un chequeo que
		// se olvidó de correr.
		if !strings.Contains(warn[0], "priority: could") || !strings.Contains(warn[0], "not blocking") {
			t.Errorf("la advertencia no explica la degradación: %q", warn[0])
		}
	})

	t.Run("el mismo caso con must sí bloquea", func(t *testing.T) {
		dir := makePriorityProject(t, "must", `["must"]`)
		blk, _ := contractIssues(dir)
		if len(blk) != 1 {
			t.Errorf("want 1 bloqueante, got %v", blk)
		}
	})

	t.Run("sin constitución, un could bloquea igual — el default no afloja nada", func(t *testing.T) {
		dir := makePriorityProject(t, "could", "")
		blk, warn := contractIssues(dir)
		if len(blk) != 1 || len(warn) != 0 {
			t.Errorf("blocking=%v warning=%v; el default debe bloquear todo", blk, warn)
		}
	})

	t.Run("un requisito sin priority declarada bloquea (fail-closed)", func(t *testing.T) {
		dir := makePriorityProject(t, "could", `["must"]`)
		// Le sacamos el campo: ahora es un requisito legado.
		writeFile(t, filepath.Join(dir, "specforge/features/f/requirements.json"),
			`{"feature":"f","requirements":[{"id":"R1","ears_type":"ubiquitous","behavior":"run","acceptance":["corre"]}]}`)
		blk, _ := contractIssues(dir)
		if len(blk) != 1 {
			t.Errorf("want 1 bloqueante: omitir priority no puede ser la forma barata de bajar el listón, got %v", blk)
		}
	})

	t.Run("sin requirements.json todo cae a must", func(t *testing.T) {
		dir := makePriorityProject(t, "could", `["must"]`)
		mustRemove(t, filepath.Join(dir, "specforge/features/f/requirements.json"))
		blk, _ := contractIssues(dir)
		if len(blk) != 1 {
			t.Errorf("want 1 bloqueante sin spec que declare prioridad, got %v", blk)
		}
	})
}

// TestCodeDriftIgnoresPriority fija el límite de R5: la prioridad gradúa el
// CONTRATO DE VERIFICACIÓN, no la integridad del trace.
//
// Un ancla de código rota no es "un requisito menor sin test": es un artefacto
// que miente sobre dónde vive lo que describe, y eso corrompe el drift de todo
// el proyecto — incluida la métrica de cobertura que otras features usan.
func TestCodeDriftIgnoresPriority(t *testing.T) {
	dir := makePriorityProject(t, "could", `["must"]`)
	// El símbolo desaparece: el ancla de código ya no resuelve.
	writeFile(t, filepath.Join(dir, "src/app.go"), "package src\n\nfunc Otra() {}\n")

	var drift []verdictIssue
	for _, is := range verdictIssues(dir, "f") {
		if strings.Contains(is.Reason, "code drift") {
			drift = append(drift, is)
		}
	}
	if len(drift) != 1 {
		t.Fatalf("want 1 razón de code drift, got %v", drift)
	}
	if !drift[0].Blocking {
		t.Error("code drift debe bloquear aunque el requisito sea `could`")
	}
}

// TestVerdictWarningsAreVisible: bajar el listón no es dejar de mirar.
func TestVerdictWarningsAreVisible(t *testing.T) {
	dir := makePriorityProject(t, "could", `["must"]`)
	warns := verdictWarnings(dir, "f")
	if len(warns) != 1 {
		t.Fatalf("want 1 advertencia expuesta, got %v", warns)
	}
	// Y no se cuelan entre las bloqueantes.
	for _, r := range verdictPreconditions(dir, "f") {
		if strings.Contains(r, "names no test") {
			t.Error("una advertencia se filtró a la lista de bloqueantes")
		}
	}
}

// TestPriorityRender: la prioridad se ve en el md, y su ausencia no ensucia el
// render legado.
func TestPriorityRender(t *testing.T) {
	render := func(r requirement) string {
		t.Helper()
		var buf strings.Builder
		if err := reqTmpl.Execute(&buf, requirementsFile{Feature: "f", Requirements: []requirement{r}}); err != nil {
			t.Fatalf("render: %v", err)
		}
		return buf.String()
	}

	out := render(requirement{ID: "R1", EarsType: "ubiquitous", Behavior: "b", Priority: "should"})
	if !strings.Contains(out, "## R1 (ubiquitous · should)") {
		t.Errorf("el render no muestra la prioridad:\n%s", out)
	}

	legacy := render(requirement{ID: "R1", EarsType: "ubiquitous", Behavior: "b"})
	if !strings.Contains(legacy, "## R1 (ubiquitous)") {
		t.Errorf("el render legado cambió:\n%s", legacy)
	}
}
