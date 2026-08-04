package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"
	"time"
)

// ----------------------------------------------------------------------------
// `sf journal add` — el CONCILIADOR DE MEMORIA (paso 4 del debate).
//
// Cuando una feature se archiva, el LLM extrae lecciones durables y se las pasa
// al CLI por stdin; el CLI las VALIDA, las persiste a un journal PROPIO de
// SpecForge (specforge/journal/<fecha>-<feature>.md, git-trackeado, SIN depender
// de ICM) y deja los archivos STAGED — nunca commitea en silencio.
//
// Doble rol (debate): el journal no es solo memoria; alimenta a futuro las
// reglas del juez de calidad (las lecciones pasadas del proyecto).
// ----------------------------------------------------------------------------

type journalEntry struct {
	SchemaVersion string          `json:"schema_version"`
	Feature       string          `json:"feature"`
	Date          string          `json:"date"`
	Lessons       []journalLesson `json:"lessons"`
}

type journalLesson struct {
	Context string   `json:"context"`           // qué pasó (la situación concreta)
	Rule    string   `json:"rule"`              // la lección durable / regla a futuro
	Tags    []string `json:"tags,omitempty"`    // para que el juez/recall las filtre
	Anchors []string `json:"anchors,omitempty"` // evidencia opcional (archivos, commits)
}

//go:embed templates/journal.tmpl.md
var journalTemplate string

var journalTmpl = template.Must(
	template.New("journal").
		Funcs(template.FuncMap{"list": joinOrDash}).
		Parse(journalTemplate),
)

// renderMarkdown: igual que los 6 artefactos, el journal sabe renderizarse.
func (v journalEntry) renderMarkdown() (string, error) { return execTemplate(journalTmpl, v) }

// checkJournal valida la estructura. Receptor *report (muta sus slices).
func checkJournal(v journalEntry, r *report) {
	if strings.TrimSpace(v.Feature) == "" {
		r.errorf("feature is required")
	}
	if len(v.Lessons) == 0 {
		r.errorf("at least one lesson is required")
	}
	for i, l := range v.Lessons {
		if strings.TrimSpace(l.Context) == "" {
			r.errorf("lesson %d: context is required", i+1)
		}
		if strings.TrimSpace(l.Rule) == "" {
			r.errorf("lesson %d: rule is required", i+1)
		}
	}
}

// runJournal parsea `sf journal add --feature=X [--date=YYYY-MM-DD] [--json -|FILE] [dir]`.
func runJournal(args []string) int {
	if len(args) == 0 || args[0] != "add" {
		fmt.Fprintln(os.Stderr, "usage: sf journal add --feature=NAME [--date=YYYY-MM-DD] [--json -|FILE] [project_dir]")
		return 2
	}
	rest := args[1:]

	projectDir := "."
	feature := ""
	date := ""
	jsonSrc := "-" // por defecto, stdin
	bridgeICM := false
	for i := 0; i < len(rest); i++ {
		a := rest[i]
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--date="):
			date = strings.TrimPrefix(a, "--date=")
		case strings.HasPrefix(a, "--json="):
			jsonSrc = strings.TrimPrefix(a, "--json=")
		case a == "--json":
			if i+1 < len(rest) {
				jsonSrc = rest[i+1]
				i++ // consumimos el valor
			}
		case a == "--bridge-icm":
			bridgeICM = true
		case a != "-" && strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf journal: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf journal: --feature=NAME is required")
		return 2
	}

	raw, err := readJSONInput(jsonSrc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf journal: cannot read input (%v)\n", err)
		return 1
	}
	var entry journalEntry
	if err := json.Unmarshal(raw, &entry); err != nil {
		fmt.Fprintf(os.Stderr, "sf journal: invalid JSON (%v)\n", err)
		return 2
	}

	// El --feature del flag manda. La fecha: flag explícito > la del JSON > hoy.
	entry.Feature = feature
	if date != "" {
		entry.Date = date
	}
	if entry.Date == "" {
		// Layout de Go: la fecha de referencia es "Mon Jan 2 15:04:05 2006".
		// "2006-01-02" significa AAAA-MM-DD. UTC para que sea estable.
		entry.Date = time.Now().UTC().Format("2006-01-02")
	}

	code := journalAdd(projectDir, entry)
	// El journal propio (git) es la fuente; ICM es un puente OPCIONAL. Solo si se
	// pidió y el write funcionó. NO depende de ICM (el journal ya está en disco).
	if code == 0 && bridgeICM {
		bridgeToICM(entry)
	}
	return code
}

// bridgeToICM hace un store best-effort a ICM si el binario está. El journal
// propio sigue siendo la verdad; esto es un espejo opcional para recall.
func bridgeToICM(entry journalEntry) {
	icm, err := exec.LookPath("icm")
	if err != nil {
		fmt.Println("note: --bridge-icm but `icm` not found — skipped (journal still written)")
		return
	}
	var rules, tags []string
	for _, l := range entry.Lessons {
		rules = append(rules, l.Rule)
		tags = append(tags, l.Tags...)
	}
	content := fmt.Sprintf("SpecForge journal — %s: %s", entry.Feature, strings.Join(rules, "; "))
	icmArgs := []string{"store", "-t", "specforge-journal", "-c", content, "-i", "high"}
	if len(tags) > 0 {
		icmArgs = append(icmArgs, "-k", strings.Join(uniqSorted(tags), ","))
	}
	if exec.Command(icm, icmArgs...).Run() == nil {
		fmt.Println("bridged to ICM (topic specforge-journal)")
	} else {
		fmt.Println("note: ICM bridge failed — journal still written")
	}
}

// journalAdd valida, escribe el JSON canónico + el markdown, y deja staged.
func journalAdd(projectDir string, entry journalEntry) int {
	var rep report
	checkJournal(entry, &rep)
	for _, w := range rep.warnings {
		fmt.Printf("  warning: %s\n", w)
	}
	if len(rep.errors) > 0 {
		for _, e := range rep.errors {
			fmt.Printf("  ERROR:   %s\n", e)
		}
		fmt.Printf("\nFAIL: not saved — %d error(s).\n", len(rep.errors))
		return 2
	}

	dir := filepath.Join(projectDir, "specforge", "journal")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf journal: %v\n", err)
		return 1
	}
	base := filepath.Join(dir, entry.Date+"-"+entry.Feature)
	jsonPath, mdPath := base+".json", base+".md"

	canonical, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf journal: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(jsonPath, append(canonical, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf journal: cannot write %s (%v)\n", jsonPath, err)
		return 1
	}
	md, err := entry.renderMarkdown()
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf journal: render failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf journal: cannot write %s (%v)\n", mdPath, err)
		return 1
	}

	fmt.Printf("journaled %s (+ rendered %s)\n", jsonPath, mdPath)
	if gitStage(projectDir, jsonPath, mdPath) {
		fmt.Println("staged for commit (NOT committed — that's your call)")
	} else {
		fmt.Println("note: could not git-stage (not a repo / git missing) — files written anyway")
	}
	return 0
}

// gitStage corre `git add` sobre los archivos, best-effort. Devuelve si funcionó.
// REGLA del conciliador: stagear sí, COMMITEAR nunca (no auto-commit silencioso).
func gitStage(projectDir string, paths ...string) bool {
	c := exec.Command("git", append([]string{"add"}, paths...)...)
	c.Dir = projectDir
	return c.Run() == nil
}
