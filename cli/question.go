package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// `sf question` — las preguntas abiertas como objeto de primera clase (DL-13).
//
// LO ÚNICO GENUINAMENTE NUEVO DE LA CAPA DE ELICITACIÓN, y no existía en
// SpecForge en ninguna forma.
//
// Para un fundador, un desconocido es un riesgo que él mismo resuelve: se
// investiga y se sigue. Para una consultora es **una pregunta que se manda y se
// espera**, a veces semanas, mientras el trabajo sigue alrededor. Esa diferencia
// no es de tono: la segunda necesita que la incógnita SOBREVIVA fuera de la
// cabeza de quien la tuvo, con id estable, estado y una lista de qué frena.
//
// POR QUÉ ES DATO Y NO PROSA DE SKILL. Una pregunta abierta anotada en un `.md`
// se pierde en la siguiente sesión, y con ella se pierde por qué una decisión
// quedó pendiente. Como dato, el gate la puede ver, y ahí aparece la regla que
// hace que todo esto valga:
//
//	CP11 — LA EXCLUSIÓN DE ALCANCE NUNCA ES UN DEFAULT SILENCIOSO.
//
// Cuando una feature depende de una pregunta sin responder, el sistema NO decide
// por su cuenta ni "la sacamos de v1" ni "esperamos". Presenta el fork y exige
// que alguien elija. Sacar algo del alcance sin decisión consciente sale más caro
// que dejarlo — el costo aparece meses después, cuando nadie recuerda que hubo
// una pregunta.
//
// Por eso `open` bloquea y hay DOS salidas, las dos dejando rastro: contestarla
// (`answer`) o adoptar una práctica externa (`adopt`). Ninguna de las dos es
// "ignorarla", que es la única salida que el diseño no ofrece.
// ----------------------------------------------------------------------------

type questionsFile struct {
	SchemaVersion string         `json:"schema_version"`
	Questions     []openQuestion `json:"questions"`
}

// openQuestion es UNA incógnita con dueño y consecuencia.
type openQuestion struct {
	ID   string `json:"id"` // Q1, Q2, … (forma Q#, como S#/R#/E#)
	Text string `json:"text"`
	// Status: open | answered | adopted-as-external-practice.
	//
	// El tercero sale de DDA CP9 y es el que más se suele omitir: una
	// recomendación por práctica externa que el humano adoptó **no asciende a
	// hecho confirmado por el cliente**. Colapsarlo en `answered` borraría
	// exactamente la distinción que después explica por qué algo se construyó
	// así — y en una consultora esa distinción es la que se discute.
	Status string `json:"status"`
	// Impact es OBLIGATORIO: qué cambia si no lo sabemos. Una pregunta sin
	// consecuencia declarada no se puede priorizar contra las otras, así que en
	// la práctica no se prioriza ninguna: la lista se vuelve un depósito.
	Impact string `json:"impact"`
	// Blocks son los nombres de feature (o ids `R#`) que esta pregunta frena.
	Blocks  []string `json:"blocks,omitempty"`
	AskedOf string   `json:"asked_of,omitempty"` // a quién se le mandó
	Asked   string   `json:"asked"`              // YYYY-MM-DD
	Answer  string   `json:"answer,omitempty"`
	// Source es el `S#` de `sources.json` donde llegó la respuesta. Cierra el
	// circuito con RM-C3: la respuesta también tiene procedencia.
	Source   string `json:"source,omitempty"`
	Answered string `json:"answered,omitempty"`
}

var questionStatuses = map[string]bool{
	"open": true, "answered": true, "adopted-as-external-practice": true,
}

func questionsPath(projectDir string) string {
	return filepath.Join(projectDir, "specforge", "questions.json")
}

// loadQuestions: lectura QUIETA. Sin archivo, no hay preguntas — que es el
// estado correcto de un proyecto que nunca abrió ninguna.
func loadQuestions(projectDir string) questionsFile {
	qf := questionsFile{SchemaVersion: schemaVersionCurrent}
	data, err := os.ReadFile(questionsPath(projectDir))
	if err != nil {
		return qf
	}
	if json.Unmarshal(data, &qf) != nil {
		return questionsFile{SchemaVersion: schemaVersionCurrent}
	}
	return qf
}

func saveQuestions(projectDir string, qf questionsFile) error {
	if qf.SchemaVersion == "" {
		qf.SchemaVersion = schemaVersionCurrent
	}
	sort.Slice(qf.Questions, func(i, j int) bool {
		return questionNum(qf.Questions[i].ID) < questionNum(qf.Questions[j].ID)
	})
	p := questionsPath(projectDir)
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(qf, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(p, append(data, '\n'), 0o644)
}

// questionNum extrae el número de `Q12`. Sirve para ordenar y para el próximo id.
func questionNum(id string) int {
	n, err := strconv.Atoi(strings.TrimPrefix(id, "Q"))
	if err != nil {
		return 0
	}
	return n
}

// nextQuestionID toma `max+1` y NUNCA reusa un id liberado — misma regla que los
// ids de criterio (RM-C1): un id que se recicla re-apunta en silencio referencias
// que ya existían.
func nextQuestionID(qf questionsFile) string {
	max := 0
	for _, q := range qf.Questions {
		if n := questionNum(q.ID); n > max {
			max = n
		}
	}
	return fmt.Sprintf("Q%d", max+1)
}

// blockingQuestions devuelve las preguntas ABIERTAS que frenan a una feature.
func blockingQuestions(projectDir, feature string) []openQuestion {
	var out []openQuestion
	for _, q := range loadQuestions(projectDir).Questions {
		// FAIL-CLOSED: un estado que este binario no conoce cuenta como abierto.
		// Si se ignorara, editar el archivo a mano y poner `"status": "wontfix"`
		// sería la forma más barata de pasar una pregunta por el gate — y el
		// punto entero de este objeto es que eso requiera una decisión con
		// rastro, no un string inventado.
		if q.Status != "open" && questionStatuses[q.Status] {
			continue
		}
		for _, b := range q.Blocks {
			if b == feature || strings.HasPrefix(b, feature+"/") {
				out = append(out, q)
				break
			}
		}
	}
	return out
}

// openQuestionReasons es lo que consume el gate del veredicto.
//
// El mensaje PRESENTA EL FORK, no lo resuelve. Decir sólo "hay una pregunta
// abierta" invitaría a la salida silenciosa —sacar el alcance y seguir—, que es
// justo lo que CP11 prohíbe. Las dos salidas que se ofrecen dejan rastro.
func openQuestionReasons(projectDir, feature string) []string {
	var reasons []string
	for _, q := range blockingQuestions(projectDir, feature) {
		reasons = append(reasons, fmt.Sprintf(
			"%s is still open and blocks this feature — %q (impact: %s). Decide explicitly: answer it "+
				"(`sf question answer --id=%s`) or adopt an external practice on the record "+
				"(`sf question adopt --id=%s`). Dropping it from scope without deciding is the one path "+
				"this gate does not offer",
			q.ID, q.Text, q.Impact, q.ID, q.ID))
	}
	return reasons
}

// ----------------------------------------------------------------------------
// El comando.
// ----------------------------------------------------------------------------

func runQuestion(args []string) int {
	if len(args) == 0 {
		questionUsage()
		return 2
	}
	switch args[0] {
	case "add":
		return runQuestionAdd(args[1:])
	case "answer":
		return runQuestionAnswer(args[1:], false)
	case "adopt":
		return runQuestionAnswer(args[1:], true)
	case "list":
		return runQuestionList(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "sf question: unknown subcommand %q\n", args[0])
		questionUsage()
		return 2
	}
}

func questionUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf question add --text=Q --impact=WHAT-CHANGES [--blocks=a,b] [--asked-of=WHO] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf question answer --id=Q1 --answer=TEXT [--source=S3] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf question adopt  --id=Q1 --answer=PRACTICE [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf question list [--json] [--blocking=FEATURE] [project_dir]")
}

func runQuestionAdd(args []string) int {
	projectDir, text, impact, askedOf := ".", "", "", ""
	var blocks []string
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--text="):
			text = strings.TrimPrefix(a, "--text=")
		case strings.HasPrefix(a, "--impact="):
			impact = strings.TrimPrefix(a, "--impact=")
		case strings.HasPrefix(a, "--asked-of="):
			askedOf = strings.TrimPrefix(a, "--asked-of=")
		case strings.HasPrefix(a, "--blocks="):
			for _, b := range strings.Split(strings.TrimPrefix(a, "--blocks="), ",") {
				if b = strings.TrimSpace(b); b != "" {
					blocks = append(blocks, b)
				}
			}
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf question: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if strings.TrimSpace(text) == "" {
		fmt.Fprintln(os.Stderr, "sf question add: --text is required")
		return 2
	}
	// El impacto es obligatorio y se valida ANTES de escribir nada. Sin él la
	// lista de preguntas se vuelve un depósito que nadie prioriza — y una lista
	// que nadie prioriza es una que nadie mira.
	if strings.TrimSpace(impact) == "" {
		fmt.Fprintln(os.Stderr, "sf question add: --impact is required — say what changes if we never find out. "+
			"A question with no declared consequence cannot be ranked against the others")
		return 2
	}

	qf := loadQuestions(projectDir)
	q := openQuestion{
		ID: nextQuestionID(qf), Text: text, Status: "open", Impact: impact,
		Blocks: blocks, AskedOf: askedOf, Asked: time.Now().UTC().Format("2006-01-02"),
	}
	qf.Questions = append(qf.Questions, q)
	if err := saveQuestions(projectDir, qf); err != nil {
		fmt.Fprintf(os.Stderr, "sf question add: %v\n", err)
		return 1
	}
	fmt.Printf("%s open — %s\n", q.ID, q.Text)
	if len(blocks) > 0 {
		fmt.Printf("  blocks: %s\n", strings.Join(blocks, ", "))
	}
	return 0
}

// runQuestionAnswer cierra una pregunta por cualquiera de las dos vías.
//
// `adopt` NO es un alias de `answer`: deja `adopted-as-external-practice`, que
// es un estado distinto a propósito. Una práctica externa que se adoptó porque
// el cliente no contestó no es un hecho que el cliente confirmó, y el día que
// alguien pregunte "¿por qué está hecho así?" la diferencia es toda la respuesta.
func runQuestionAnswer(args []string, adopt bool) int {
	projectDir, id, answer, source := ".", "", "", ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--id="):
			id = strings.TrimPrefix(a, "--id=")
		case strings.HasPrefix(a, "--answer="):
			answer = strings.TrimPrefix(a, "--answer=")
		case strings.HasPrefix(a, "--source="):
			source = strings.TrimPrefix(a, "--source=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf question: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if id == "" || strings.TrimSpace(answer) == "" {
		fmt.Fprintln(os.Stderr, "sf question: --id and --answer are required")
		return 2
	}

	qf := loadQuestions(projectDir)
	found := false
	for i := range qf.Questions {
		q := &qf.Questions[i]
		if q.ID != id {
			continue
		}
		found = true
		q.Answer = answer
		q.Answered = time.Now().UTC().Format("2006-01-02")
		q.Source = source
		if adopt {
			q.Status = "adopted-as-external-practice"
		} else {
			q.Status = "answered"
		}
	}
	if !found {
		fmt.Fprintf(os.Stderr, "sf question: no question %q\n", id)
		return 4
	}
	if err := saveQuestions(projectDir, qf); err != nil {
		fmt.Fprintf(os.Stderr, "sf question: %v\n", err)
		return 1
	}
	if adopt {
		fmt.Printf("%s adopted as external practice — NOT confirmed by the client, and it stays labelled that way.\n", id)
	} else {
		fmt.Printf("%s answered.\n", id)
	}
	return 0
}

func runQuestionList(args []string) int {
	projectDir, blocking := ".", ""
	asJSON := false
	for _, a := range args {
		switch {
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "--blocking="):
			blocking = strings.TrimPrefix(a, "--blocking=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf question: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	items := loadQuestions(projectDir).Questions
	if blocking != "" {
		items = blockingQuestions(projectDir, blocking)
	}

	if asJSON {
		if items == nil {
			items = []openQuestion{}
		}
		out, _ := json.MarshalIndent(map[string]any{"questions": items}, "", "  ")
		fmt.Println(string(out))
		return 0
	}

	if len(items) == 0 {
		fmt.Println("No open questions on record.")
		return 0
	}
	cols := []string{"id", "status", "asked", "blocks", "question"}
	var rows [][]string
	open := 0
	for _, q := range items {
		if q.Status == "open" {
			open++
		}
		rows = append(rows, []string{q.ID, q.Status, orDash(q.Asked),
			orDash(strings.Join(q.Blocks, " ")), q.Text})
	}
	renderTable(cols, rows)
	if open > 0 {
		// Se cuentan aparte las que FRENAN algo: una pregunta abierta sin
		// `blocks` es un estado legítimo —una incógnita que todavía no detiene
		// nada—, y decir que "todas bloquean" sería falso justo en el mensaje
		// que la gente usa para decidir a qué correr primero.
		blockingCount := 0
		for _, q := range items {
			if q.Status == "open" && len(q.Blocks) > 0 {
				blockingCount++
			}
		}
		fmt.Printf("\n%d open", open)
		if blockingCount > 0 {
			fmt.Printf(", %d of them blocking a verdict until answered or adopted as an external\n", blockingCount)
			fmt.Println("practice on the record — dropping one from scope silently is not an option.")
		} else {
			fmt.Println(" — none of them lists a feature it blocks, so none stops a verdict yet.")
		}
	}
	return 0
}
