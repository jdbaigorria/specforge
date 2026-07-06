package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// `sf feature` — el ESCRITOR ÚNICO del ciclo de vida en features.json.
//
// Por qué existe (FIXBUGHIGH, cierre de borde): la Capa 1 protege features.json
// del Write directo. Pero el ciclo de vida (crear feature, mover status, fijar
// lane, archivar) ANTES se hacía con Write a mano. Sin un camino CLI, proteger
// el archivo trabaría el flujo. Acá está ese camino: el CLI valida la transición
// y es el único que muta status/lane. Los gates siguen siendo de `sf gate`.
//
// Subacciones:
//   add         crea una feature nueva (status inicial: planned)
//   set-status  mueve el status validando la transición + el flujo serial (F22)
//   set-lane    fija el carril (lite|standard) antes de construir
//   archive     status→done + copia la carpeta a specforge/archive/<date>-<name>/
// ----------------------------------------------------------------------------

// validStatuses: el vocabulario del ciclo de vida (sale de los skills).
var validStatuses = map[string]bool{
	"planned": true, "approved": true, "building": true,
	"checking": true, "done": true, "blocked": true,
}

// statusTransitions declara, por status origen, los destinos permitidos. Forward
// es el pipeline normal; las back-edges cubren revisar (checking→building),
// despausar y bloquear/desbloquear. `done` es terminal (solo lo pone `archive`).
var statusTransitions = map[string][]string{
	"planned":  {"approved", "blocked"},
	"approved": {"building", "planned", "blocked"},
	"building": {"checking", "blocked"},
	"checking": {"done", "building", "blocked"}, // building = revisar
	"blocked":  {"planned", "approved", "building", "checking"},
	"done":     {}, // terminal
}

// validLanes: los dos carriles.
var validLanes = map[string]bool{"lite": true, "standard": true}

func runFeature(args []string) int {
	if len(args) == 0 {
		featureUsage()
		return 2
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "add":
		return runFeatureAdd(rest)
	case "set-status":
		return runFeatureSetStatus(rest)
	case "set-lane":
		return runFeatureSetLane(rest)
	case "archive":
		return runFeatureArchive(rest)
	default:
		fmt.Fprintf(os.Stderr, "sf feature: unknown sub-command %q\n\n", sub)
		featureUsage()
		return 2
	}
}

func featureUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf feature add --feature=NAME [--lane=lite|standard] [--depends-on=a,b] [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf feature set-status --feature=NAME --to=STATUS [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf feature set-lane --feature=NAME --to=lite|standard [project_dir]")
	fmt.Fprintln(os.Stderr, "       sf feature archive --feature=NAME [project_dir]")
}

// featureFlags parsea el set común de flags. Devuelve (projectDir, feature, to,
// lane, dependsOn, ok). Un flag desconocido imprime el error y deja ok=false.
func featureFlags(args []string) (projectDir, feature, to, lane string, dependsOn []string, ok bool) {
	projectDir = "."
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--to="):
			to = strings.TrimPrefix(a, "--to=")
		case strings.HasPrefix(a, "--lane="):
			lane = strings.TrimPrefix(a, "--lane=")
		case strings.HasPrefix(a, "--depends-on="):
			for _, d := range strings.Split(strings.TrimPrefix(a, "--depends-on="), ",") {
				if d = strings.TrimSpace(d); d != "" {
					dependsOn = append(dependsOn, d)
				}
			}
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf feature: unknown flag %q\n", a)
			return "", "", "", "", nil, false
		default:
			projectDir = a
		}
	}
	return projectDir, feature, to, lane, dependsOn, true
}

// findFeature devuelve el puntero al elemento real del slice (mutable), o nil.
func findFeature(ff *featuresFile, name string) *feature {
	for i := range ff.Features {
		if ff.Features[i].Name == name {
			return &ff.Features[i]
		}
	}
	return nil
}

// ── add ──────────────────────────────────────────────────────────────────────

func runFeatureAdd(args []string) int {
	projectDir, name, _, lane, dependsOn, ok := featureFlags(args)
	if !ok {
		return 2
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "sf feature add: --feature=NAME is required")
		return 2
	}
	if lane != "" && !validLanes[lane] {
		fmt.Fprintf(os.Stderr, "sf feature add: --lane must be lite|standard, got %q\n", lane)
		return 2
	}

	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		// Sin estado todavía: registro vacío (la primera feature lo estrena).
		ff = featuresFile{SchemaVersion: schemaVersionCurrent}
	}
	if findFeature(&ff, name) != nil {
		fmt.Fprintf(os.Stderr, "sf feature add: feature %q already exists\n", name)
		return 4
	}

	// A1/R3: el estado nace directamente POR FEATURE (feature.json propio). No
	// se crea ni se toca ningún features.json global.
	if code := writeFeatureState(projectDir, feature{
		Name:      name,
		Status:    "planned", // toda feature nace en el backlog
		Lane:      lane,
		DependsOn: dependsOn,
		Gates:     []gate{},
		Seq:       nextSeq(ff),
	}); code != 0 {
		return code
	}
	fmt.Printf("added feature %q (status planned%s)\n", name, laneSuffix(lane))
	return 0
}

func laneSuffix(lane string) string {
	if lane == "" {
		return ""
	}
	return ", lane " + lane
}

// ── set-status ───────────────────────────────────────────────────────────────

func runFeatureSetStatus(args []string) int {
	projectDir, name, to, _, _, ok := featureFlags(args)
	if !ok {
		return 2
	}
	if name == "" || to == "" {
		fmt.Fprintln(os.Stderr, "sf feature set-status: --feature=NAME and --to=STATUS are required")
		return 2
	}
	if !validStatuses[to] {
		fmt.Fprintf(os.Stderr, "sf feature set-status: unknown status %q (valid: planned|approved|building|checking|done|blocked)\n", to)
		return 2
	}

	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf feature set-status: cannot read feature state under %s (%v)\n", projectDir, err)
		return 4
	}
	f := findFeature(&ff, name)
	if f == nil {
		fmt.Fprintf(os.Stderr, "sf feature set-status: feature %q not found\n", name)
		return 4
	}

	from := f.Status
	if from == to {
		fmt.Printf("feature %q already in status %q (no-op)\n", name, to)
		return 0
	}

	// `done` solo lo otorga `archive` (que además copia la carpeta y exige el
	// verdict). No se puede saltar el archivado fijando done a mano. Se chequea
	// ANTES de la transición para dar siempre el mensaje útil, no "ilegal".
	if to == "done" {
		fmt.Fprintln(os.Stderr, "sf feature set-status: use `sf feature archive` to finish a feature (it seals done + archives the folder)")
		return 2
	}
	if !transitionAllowed(from, to) {
		fmt.Fprintf(os.Stderr, "sf feature set-status: illegal transition %s → %s for %q. Allowed from %s: %s\n",
			orPlanned(from), to, name, orPlanned(from), strings.Join(statusTransitions[orPlanned(from)], ", "))
		return 5
	}

	// Flujo serial (F22): no entrar a un status ACTIVO si otra feature ya lo
	// está. F2: con flow.mode=parallel en la constitución, el guard no aplica.
	if activeStatuses[to] && !parallelFlow(projectDir) {
		if other := activeOther(ff, name); other != "" {
			fmt.Fprintf(os.Stderr, "sf feature set-status: feature %q is still active. Finish it before activating %q "+
				"(one at a time, F22 — or set flow.mode=parallel in the constitution for team flow).\n", other, name)
			return 5
		}
	}

	f.Status = to
	if code := writeFeatureState(projectDir, *f); code != 0 {
		return code
	}
	fmt.Printf("feature %q: %s → %s\n", name, orPlanned(from), to)
	return 0
}

// transitionAllowed consulta la tabla. Un origen vacío se trata como "planned"
// (features viejas sin status explícito).
func transitionAllowed(from, to string) bool {
	for _, dst := range statusTransitions[orPlanned(from)] {
		if dst == to {
			return true
		}
	}
	return false
}

func orPlanned(s string) string {
	if s == "" {
		return "planned"
	}
	return s
}

// ── set-lane ─────────────────────────────────────────────────────────────────

func runFeatureSetLane(args []string) int {
	projectDir, name, to, _, _, ok := featureFlags(args)
	if !ok {
		return 2
	}
	// set-lane usa --to= para el carril (uniforme con set-status).
	if name == "" || to == "" {
		fmt.Fprintln(os.Stderr, "sf feature set-lane: --feature=NAME and --to=lite|standard are required")
		return 2
	}
	if !validLanes[to] {
		fmt.Fprintf(os.Stderr, "sf feature set-lane: --to must be lite|standard, got %q\n", to)
		return 2
	}

	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf feature set-lane: cannot read feature state under %s (%v)\n", projectDir, err)
		return 4
	}
	f := findFeature(&ff, name)
	if f == nil {
		fmt.Fprintf(os.Stderr, "sf feature set-lane: feature %q not found\n", name)
		return 4
	}
	// El carril decide la forma del pipeline; cambiarlo después de empezar a
	// construir sería incoherente. Solo antes de `building`.
	if f.Status == "building" || f.Status == "checking" || f.Status == "done" {
		fmt.Fprintf(os.Stderr, "sf feature set-lane: feature %q is past planning (status %s) — lane is locked\n", name, f.Status)
		return 5
	}

	f.Lane = to
	if code := writeFeatureState(projectDir, *f); code != 0 {
		return code
	}
	fmt.Printf("feature %q: lane → %s\n", name, to)
	return 0
}

// ── archive ──────────────────────────────────────────────────────────────────

func runFeatureArchive(args []string) int {
	projectDir, name, _, _, _, ok := featureFlags(args)
	if !ok {
		return 2
	}
	if name == "" {
		fmt.Fprintln(os.Stderr, "sf feature archive: --feature=NAME is required")
		return 2
	}

	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf feature archive: cannot read feature state under %s (%v)\n", projectDir, err)
		return 4
	}
	f := findFeature(&ff, name)
	if f == nil {
		fmt.Fprintf(os.Stderr, "sf feature archive: feature %q not found\n", name)
		return 4
	}
	if f.Status == "done" {
		fmt.Printf("feature %q is already archived\n", name)
		return 0
	}

	// R1 (integrity.go): un verdict que vive en un ledger forjado no cuenta.
	// Validamos la cadena ANTES de mirar el gate — si alguien escribió el
	// verdict a mano (fuera del arnés, sin hooks), acá se detecta y se rehúsa.
	if refuseOnBrokenLedger("sf feature archive", f) {
		return 5
	}

	// El archivado exige el verdict aprobado: y `sf gate approve --phase=verdict`
	// ya verifica trace limpio + test verde y fresco (Capa 2). Así, archivar una
	// feature cuyo código no anda es imposible: sin verdict no hay archive.
	if latestApproveGate(f, "verdict") == nil {
		fmt.Fprintf(os.Stderr, "sf feature archive: the verdict gate for %q is not approved. "+
			"Run `sf gate approve --feature=%s --phase=verdict` first (it requires a clean trace and a fresh green test run).\n", name, name)
		return 5
	}

	// Copiamos la carpeta de la feature al archivo, fechada.
	src := filepath.Join(projectDir, "specforge", "features", name)
	dateStr := time.Now().UTC().Format("2006-01-02")
	dst := filepath.Join(projectDir, "specforge", "archive", dateStr+"-"+name)
	if isDir(src) {
		if err := copyTree(src, dst); err != nil {
			fmt.Fprintf(os.Stderr, "sf feature archive: copy failed (%v)\n", err)
			return 1
		}
	}

	f.Status = "done"
	if code := writeFeatureState(projectDir, *f); code != 0 {
		return code
	}
	fmt.Printf("archived feature %q → %s (status done)\n", name, filepath.ToSlash(dst))
	return 0
}

// copyTree copia recursivamente src → dst (archivos y subdirs). No sigue
// symlinks (los omite). Preserva el modo de cada archivo.
func copyTree(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(src, path)
		if rerr != nil {
			return rerr
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return nil // no copiamos symlinks
		}
		// copyFile (install_apply.go) crea el archivo; el dir padre ya existe
		// porque copyTree crea cada subdir antes de llegar a sus archivos.
		return copyFile(path, target)
	})
}
