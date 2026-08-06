package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// DL-13 — preguntas abiertas como objeto de primera clase.
//
// La regla que hay que proteger con tests es CP11: **la exclusión de alcance
// nunca es un default silencioso**. Todo lo demás —ids, estados, listados— es
// plomería alrededor de eso. Por eso los tests que importan son los que fijan
// que `open` BLOQUEA, que hay exactamente dos salidas y las dos dejan rastro, y
// que un estado inventado a mano NO es una tercera.
// ----------------------------------------------------------------------------

func makeQuestionProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"2.0","features":[{"name":"checkout","status":"checking","gates":[]}]}`)
	return dir
}

func TestQuestionAdd(t *testing.T) {
	t.Run("el impacto es obligatorio", func(t *testing.T) {
		// Sin consecuencia declarada una pregunta no se puede rankear contra las
		// otras, y una lista que nadie prioriza es una que nadie mira.
		dir := makeQuestionProject(t)
		if code := runQuestion([]string{"add", "--text=¿IVA incluido?", dir}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
	})

	t.Run("ids consecutivos que NO se reusan", func(t *testing.T) {
		// Misma regla que los ids de criterio (RM-C1): un id reciclado re-apunta
		// en silencio referencias que ya existían.
		dir := makeQuestionProject(t)
		for i := 0; i < 3; i++ {
			if code := runQuestion([]string{"add", "--text=q", "--impact=cambia el modelo", dir}); code != 0 {
				t.Fatalf("add exit %d", code)
			}
		}
		qf := loadQuestions(dir)
		if len(qf.Questions) != 3 {
			t.Fatalf("questions = %d", len(qf.Questions))
		}
		// Se borra la del medio: la próxima debe ser Q4, no Q2.
		qf.Questions = []openQuestion{qf.Questions[0], qf.Questions[2]}
		if err := saveQuestions(dir, qf); err != nil {
			t.Fatal(err)
		}
		if got := nextQuestionID(loadQuestions(dir)); got != "Q4" {
			t.Errorf("nextQuestionID = %q, want Q4 — un id liberado nunca se reusa", got)
		}
	})
}

func TestOpenQuestionBlocksTheGate(t *testing.T) {
	t.Run("una pregunta abierta que declara blocks FRENA a esa feature", func(t *testing.T) {
		dir := makeQuestionProject(t)
		runQuestion([]string{"add", "--text=¿el IVA va incluido en el precio?",
			"--impact=cambia el modelo de Order y el cálculo del total",
			"--blocks=checkout", dir})

		reasons := openQuestionReasons(dir, "checkout")
		if len(reasons) != 1 {
			t.Fatalf("reasons = %v", reasons)
		}
		// El motivo PRESENTA EL FORK. Decir sólo "hay una pregunta abierta"
		// invitaría a la salida silenciosa, que es lo que CP11 prohíbe.
		for _, want := range []string{"answer it", "adopt an external practice", "without deciding"} {
			if !strings.Contains(reasons[0], want) {
				t.Errorf("el motivo debe ofrecer el fork (%q): %q", want, reasons[0])
			}
		}
		// Y lleva el impacto, que es lo que permite decidir sin ir a buscarlo.
		if !strings.Contains(reasons[0], "cambia el modelo de Order") {
			t.Errorf("el motivo debe llevar el impacto: %q", reasons[0])
		}
	})

	t.Run("no frena a una feature que no declara", func(t *testing.T) {
		dir := makeQuestionProject(t)
		runQuestion([]string{"add", "--text=q", "--impact=i", "--blocks=otra", dir})
		if reasons := openQuestionReasons(dir, "checkout"); len(reasons) != 0 {
			t.Errorf("reasons = %v", reasons)
		}
	})

	t.Run("contestarla la destraba", func(t *testing.T) {
		dir := makeQuestionProject(t)
		runQuestion([]string{"add", "--text=q", "--impact=i", "--blocks=checkout", dir})
		if code := runQuestion([]string{"answer", "--id=Q1", "--answer=sí, incluido", "--source=S2", dir}); code != 0 {
			t.Fatalf("answer exit %d", code)
		}
		if reasons := openQuestionReasons(dir, "checkout"); len(reasons) != 0 {
			t.Errorf("reasons = %v", reasons)
		}
		if q := loadQuestions(dir).Questions[0]; q.Status != "answered" || q.Source != "S2" {
			t.Errorf("q = %+v", q)
		}
	})

	t.Run("adoptar una práctica externa TAMBIÉN destraba, con estado distinto", func(t *testing.T) {
		// `adopt` no es alias de `answer`: una práctica externa adoptada porque
		// el cliente no contestó NO es un hecho que el cliente confirmó, y el día
		// que alguien pregunte por qué está hecho así, la diferencia es toda la
		// respuesta.
		dir := makeQuestionProject(t)
		runQuestion([]string{"add", "--text=q", "--impact=i", "--blocks=checkout", dir})
		if code := runQuestion([]string{"adopt", "--id=Q1", "--answer=IVA incluido, práctica de la industria", dir}); code != 0 {
			t.Fatalf("adopt exit %d", code)
		}
		if reasons := openQuestionReasons(dir, "checkout"); len(reasons) != 0 {
			t.Errorf("reasons = %v", reasons)
		}
		if q := loadQuestions(dir).Questions[0]; q.Status != "adopted-as-external-practice" {
			t.Errorf("status = %q — no debe colapsarse en `answered`", q.Status)
		}
	})

	t.Run("un estado INVENTADO no destraba (fail-closed)", func(t *testing.T) {
		// Sin esto, editar el archivo a mano y poner "wontfix" sería la forma más
		// barata de pasar una pregunta por el gate.
		dir := makeQuestionProject(t)
		runQuestion([]string{"add", "--text=q", "--impact=i", "--blocks=checkout", dir})
		qf := loadQuestions(dir)
		qf.Questions[0].Status = "wontfix"
		if err := saveQuestions(dir, qf); err != nil {
			t.Fatal(err)
		}
		if reasons := openQuestionReasons(dir, "checkout"); len(reasons) != 1 {
			t.Errorf("un status desconocido debe contar como abierto, got %v", reasons)
		}
	})
}

func TestQuestionListAndSurface(t *testing.T) {
	t.Run("--blocking filtra por feature", func(t *testing.T) {
		dir := makeQuestionProject(t)
		runQuestion([]string{"add", "--text=a", "--impact=i", "--blocks=checkout", dir})
		runQuestion([]string{"add", "--text=b", "--impact=i", "--blocks=otra", dir})
		if got := len(blockingQuestions(dir, "checkout")); got != 1 {
			t.Errorf("blocking = %d, want 1", got)
		}
	})

	t.Run("sin preguntas no es un error", func(t *testing.T) {
		if code := runQuestion([]string{"list", makeQuestionProject(t)}); code != 0 {
			t.Errorf("exit = %d, want 0", code)
		}
	})

	t.Run("subcomando desconocido y flags malos fallan con 2", func(t *testing.T) {
		if code := runQuestion([]string{"nope"}); code != 2 {
			t.Errorf("exit = %d", code)
		}
		if code := runQuestion(nil); code != 2 {
			t.Errorf("exit = %d", code)
		}
		if code := runQuestion([]string{"answer", "--id=Q1"}); code != 2 {
			t.Errorf("answer sin --answer debe fallar, got %d", code)
		}
	})

	t.Run("answer sobre un id inexistente sale 4", func(t *testing.T) {
		dir := makeQuestionProject(t)
		if code := runQuestion([]string{"answer", "--id=Q9", "--answer=x", dir}); code != 4 {
			t.Errorf("exit = %d, want 4", code)
		}
	})
}

// TestQuestionsJSONIsProtected — `questions.json` es estado autoritativo: si el
// agente pudiera escribirlo a mano, podría cerrar una pregunta sin pasar por la
// decisión. Es la Capa 1 aplicada al objeto nuevo.
func TestQuestionsJSONIsProtected(t *testing.T) {
	if !protectedJSONRe.MatchString("specforge/questions.json") {
		t.Error("specforge/questions.json debe estar protegido por el hook")
	}
}
