package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf delta` — el CAMBIO como objeto de primera clase (D4', idea de OpenSpec).
//
// Hasta acá, amendar una feature shipped era prosa: sf-amend decía "declará
// precisamente qué cambia" y eso vivía en la conversación. Un delta es esa
// declaración convertida en ARTEFACTO validado y versionado:
//
//   specforge/features/<f>/deltas/<id>.json
//   { id, feature, status: proposed→applied→archived, why,
//     changes: [{op: add|modify|remove, target, ref, description}] }
//
// Qué ordena esto:
//   - el gate del amend aprueba un OBJETO con hash (sellable), no un párrafo;
//   - el cascade de staleness se deriva de `changes[].target` (qué artefactos
//     reabre), no del criterio del modelo;
//   - la historia del amend queda auditable: cada delta archivado dice qué
//     cambió, por qué y cuándo se aplicó.
//
// Los IDs de elementos (R3, C2, T4) se mantienen ESTABLES a través de deltas —
// la regla de sf-amend, ahora chequeable.
// ----------------------------------------------------------------------------

type deltaFile struct {
	SchemaVersion string        `json:"schema_version"`
	ID            string        `json:"id"` // D1, D2, … (por feature)
	Feature       string        `json:"feature"`
	Status        string        `json:"status"` // proposed | applied | archived
	Why           string        `json:"why"`
	Changes       []deltaChange `json:"changes"`
	CreatedAt     string        `json:"created_at"`
	AppliedAt     string        `json:"applied_at,omitempty"`
}

type deltaChange struct {
	Op          string `json:"op"`     // add | modify | remove
	Target      string `json:"target"` // requirements | design | tasks
	Ref         string `json:"ref"`    // el id estable del elemento (R3/C2/T4; nuevo id si op=add)
	Description string `json:"description"`
}

var deltaOps = map[string]bool{"add": true, "modify": true, "remove": true}
var deltaTargets = map[string]bool{"requirements": true, "design": true, "tasks": true}

// deltaTransitions: el ciclo de vida es una escalera sin retorno (proposed →
// applied → archived). Un delta rechazado no retrocede: se archiva y se
// propone otro — historia append-only, igual que el resto del sistema.
var deltaTransitions = map[string]string{"proposed": "applied", "applied": "archived"}

func runDelta(args []string) int {
	if len(args) == 0 {
		deltaUsage()
		return 2
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "new":
		return runDeltaNew(rest)
	case "list":
		return runDeltaList(rest)
	case "set-status":
		return runDeltaSetStatus(rest)
	default:
		fmt.Fprintf(os.Stderr, "sf delta: unknown sub-command %q\n\n", sub)
		deltaUsage()
		return 2
	}
}

func deltaUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf delta new --feature=NAME [--json -|FILE | --from=DRAFT] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf delta list --feature=NAME [--json] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf delta set-status --feature=NAME --id=ID --to=applied|archived [project_dir]")
}

// deltaDir: la carpeta de deltas de una feature.
func deltaDir(projectDir, feature string) string {
	return filepath.Join(projectDir, "specforge", "features", feature, "deltas")
}

// ── new ──────────────────────────────────────────────────────────────────────

func runDeltaNew(args []string) int {
	projectDir, feature, jsonSrc, fromDraft := ".", "", "-", ""
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--json="):
			jsonSrc = strings.TrimPrefix(a, "--json=")
		case strings.HasPrefix(a, "--from="):
			fromDraft = strings.TrimPrefix(a, "--from=")
		case a == "--json":
			if i+1 < len(args) {
				jsonSrc = args[i+1]
				i++
			}
		case a != "-" && strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf delta: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf delta new: --feature=NAME is required")
		return 2
	}
	if fromDraft != "" {
		jsonSrc = draftPath(projectDir, feature, fromDraft)
	}

	raw, err := readJSONInput(jsonSrc)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf delta: cannot read input (%v)\n", err)
		return 1
	}
	var d deltaFile
	if err := json.Unmarshal(raw, &d); err != nil {
		fmt.Fprintf(os.Stderr, "sf delta: invalid JSON (%v)\n", err)
		return 2
	}

	// El CLI fija lo que es suyo: feature, id secuencial, estado inicial,
	// timestamp, versión. El productor solo declara why + changes.
	d.SchemaVersion = schemaVersionCurrent
	d.Feature = feature
	d.Status = "proposed"
	d.CreatedAt = nowUTC()
	if d.ID == "" {
		d.ID = fmt.Sprintf("D%d", len(listDeltas(projectDir, feature))+1)
	}

	var rep report
	checkDelta(d, &rep)
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

	path := filepath.Join(deltaDir(projectDir, feature), d.ID+".json")
	if fileExists(path) {
		fmt.Fprintf(os.Stderr, "sf delta: %s already exists — deltas are append-only (new id, new delta)\n", d.ID)
		return 4
	}
	if code := writeDelta(path, d); code != 0 {
		return code
	}
	if fromDraft != "" {
		_ = os.Remove(jsonSrc) // promovido: retiramos el borrador (mismo contrato que sf save)
	}
	fmt.Printf("saved delta %s for %s (proposed, %d change(s)) — gate it, then `sf delta set-status --to=applied`\n",
		d.ID, feature, len(d.Changes))
	return 0
}

// checkDelta valida el objeto: why + al menos un change, cada change con op y
// target del vocabulario y ref estable no vacío.
func checkDelta(d deltaFile, rep *report) {
	if strings.TrimSpace(d.Why) == "" {
		rep.errorf("why is required — a delta without a reason is not auditable")
	}
	if len(d.Changes) == 0 {
		rep.errorf("at least one change is required")
	}
	for i, c := range d.Changes {
		if !deltaOps[c.Op] {
			rep.errorf("change %d: op must be add|modify|remove, got %q", i+1, c.Op)
		}
		if !deltaTargets[c.Target] {
			rep.errorf("change %d: target must be requirements|design|tasks, got %q", i+1, c.Target)
		}
		if strings.TrimSpace(c.Ref) == "" {
			rep.errorf("change %d: ref is required (stable element id, e.g. R3 — new id when op=add)", i+1)
		}
		if strings.TrimSpace(c.Description) == "" {
			rep.errorf("change %d: description is required", i+1)
		}
	}
}

func writeDelta(path string, d deltaFile) int {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf delta: %v\n", err)
		return 1
	}
	out, err := json.MarshalIndent(d, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf delta: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf delta: cannot write %s (%v)\n", path, err)
		return 1
	}
	return 0
}

// listDeltas lee todos los deltas de una feature, ordenados por id.
func listDeltas(projectDir, feature string) []deltaFile {
	paths, _ := filepath.Glob(filepath.Join(deltaDir(projectDir, feature), "*.json"))
	var out []deltaFile
	for _, p := range paths {
		var d deltaFile
		if data, err := os.ReadFile(p); err == nil && json.Unmarshal(data, &d) == nil {
			out = append(out, d)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt < out[j].CreatedAt })
	return out
}

// ── list ─────────────────────────────────────────────────────────────────────

func runDeltaList(args []string) int {
	projectDir, feature, asJSON := ".", "", false
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf delta: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf delta list: --feature=NAME is required")
		return 2
	}
	deltas := listDeltas(projectDir, feature)
	if asJSON {
		out, _ := json.MarshalIndent(deltas, "", "  ")
		fmt.Println(string(out))
		return 0
	}
	if len(deltas) == 0 {
		fmt.Printf("No deltas for %s.\n", feature)
		return 0
	}
	cols := []string{"id", "status", "changes", "why", "created"}
	var rows [][]string
	for _, d := range deltas {
		rows = append(rows, []string{d.ID, d.Status, fmt.Sprintf("%d", len(d.Changes)), d.Why, d.CreatedAt})
	}
	renderTable(cols, rows)
	return 0
}

// ── set-status ───────────────────────────────────────────────────────────────

func runDeltaSetStatus(args []string) int {
	projectDir, feature, id, to := ".", "", "", ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--id="):
			id = strings.TrimPrefix(a, "--id=")
		case strings.HasPrefix(a, "--to="):
			to = strings.TrimPrefix(a, "--to=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf delta: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" || id == "" || to == "" {
		fmt.Fprintln(os.Stderr, "sf delta set-status: --feature, --id and --to are required")
		return 2
	}

	path := filepath.Join(deltaDir(projectDir, feature), id+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf delta: no delta %s for %s\n", id, feature)
		return 4
	}
	var d deltaFile
	if err := json.Unmarshal(data, &d); err != nil {
		fmt.Fprintf(os.Stderr, "sf delta: invalid %s (%v)\n", path, err)
		return 2
	}

	// La escalera: solo el paso siguiente es legal (proposed→applied→archived).
	if deltaTransitions[d.Status] != to {
		fmt.Fprintf(os.Stderr, "sf delta: illegal transition %s → %s (lifecycle: proposed → applied → archived)\n", d.Status, to)
		return 5
	}
	d.Status = to
	if to == "applied" {
		d.AppliedAt = nowUTC()
	}
	if code := writeDelta(path, d); code != 0 {
		return code
	}
	fmt.Printf("delta %s: → %s\n", id, to)
	return 0
}
