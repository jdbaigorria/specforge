package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf check run` — Capa 3 de FIXBUGHIGH: mover la EJECUCIÓN de tests ADENTRO del
// CLI.
//
// El problema que resuelve: hoy "corré los tests" es prosa en una skill. El
// resultado vive en la narración del LLM, que puede mentir ("70% coverage, todo
// verde") sin haber corrido nada. Acá el CLI shellea el comando real, captura el
// EXIT CODE (que el LLM no puede falsificar) y lo registra junto a un hash del
// código. Ese registro es lo que el verdict (Capa 2) consume: sin un verde y
// FRESCO, no hay archive.
//
// "Fresco" = el hash del código al momento de correr el test == el hash actual.
// Así, "corrí una vez y después cambié todo" queda detectado: cambió el código →
// cambió el hash → el resultado guardado es stale.
// ----------------------------------------------------------------------------

// checkResult es el registro determinista de UNA corrida de tests. Vive bajo
// specforge/.state/ (estado de máquina, protegido del Write directo por la Capa 1).
type checkResult struct {
	Feature    string `json:"feature"`
	At         string `json:"at"`
	TestCmd    string `json:"test_cmd"`
	ExitCode   int    `json:"exit_code"`
	Passed     bool   `json:"passed"`
	CodeHash   string `json:"code_hash"`   // hash del código fuente al correr (freshness)
	OutputTail string `json:"output_tail"` // últimas líneas del output, para diagnóstico
}

// outputTailBytes: cuánto del final del output guardamos (para no inflar el .json).
const outputTailBytes = 4000

func runCheck(args []string) int {
	if len(args) == 0 {
		checkUsage()
		return 2
	}
	sub, rest := args[0], args[1:]
	switch sub {
	case "run":
		return runCheckRun(rest)
	default:
		fmt.Fprintf(os.Stderr, "sf check: unknown sub-command %q\n\n", sub)
		checkUsage()
		return 2
	}
}

func checkUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf check run --feature=NAME [project_dir]")
}

func runCheckRun(args []string) int {
	projectDir := "."
	feature := ""
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf check: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if feature == "" {
		fmt.Fprintln(os.Stderr, "sf check: --feature=NAME is required")
		return 2
	}

	// 1) El comando de test sale de la constitución (build.test_cmd). No se lo
	//    pedimos al LLM: si pudiera pasarlo por flag, podría pasar `true` y fingir
	//    verde. Tiene que estar declarado y versionado en el proyecto.
	cf, code := readConstitutionFile(projectDir)
	if code != 0 {
		return code
	}
	if cf.Build == nil || strings.TrimSpace(cf.Build.TestCmd) == "" {
		fmt.Fprintln(os.Stderr, "sf check: no build.test_cmd in the constitution — "+
			"declare it (e.g. \"pytest -q\") so the CLI can run the suite deterministically.")
		return 2
	}
	testCmd := cf.Build.TestCmd

	// 2) Hash del código ANTES de correr (la foto contra la que se mide frescura).
	codeHash := codeHash(projectDir)

	// 3) Ejecutamos el comando real vía `sh -c` en el dir del proyecto, capturando
	//    stdout+stderr combinados. El exit code es la verdad que no se falsifica.
	fmt.Printf("running: %s\n", testCmd)
	out, exit := runTestCommand(projectDir, testCmd)
	passed := exit == 0

	res := checkResult{
		Feature:    feature,
		At:         nowUTC(),
		TestCmd:    testCmd,
		ExitCode:   exit,
		Passed:     passed,
		CodeHash:   codeHash,
		OutputTail: tailString(out, outputTailBytes),
	}
	if c := writeCheckResult(projectDir, feature, res); c != 0 {
		return c
	}

	if passed {
		fmt.Printf("PASS — %s/%s (exit 0, code %s…)\n", feature, "tests", short(codeHash))
		return 0
	}
	// Recorded pero FAIL: exit 3 (misma convención que record-verdict) → señal
	// limpia para el hook/agente sin parsear stdout.
	fmt.Printf("FAIL — %s tests exited %d (recorded). Fix the code and re-run `sf check run`.\n", feature, exit)
	return 3
}

// runTestCommand shellea el comando con `sh -c` en projectDir y devuelve el
// output combinado + el exit code. Un comando inexistente o un error de arranque
// se reportan como exit 127 (convención de shell para "command not found").
func runTestCommand(projectDir, testCmd string) (string, int) {
	cmd := exec.Command("sh", "-c", testCmd)
	cmd.Dir = projectDir
	out, err := cmd.CombinedOutput()
	if err == nil {
		return string(out), 0
	}
	// ExitError lleva el código real del proceso; cualquier otro error (no se pudo
	// arrancar el shell) lo tratamos como 127.
	if ee, ok := err.(*exec.ExitError); ok {
		return string(out), ee.ExitCode()
	}
	return string(out) + "\n" + err.Error(), 127
}

// ── Hash del código fuente (freshness) ───────────────────────────────────────

// codeHash devuelve un sha256 (hex) determinista sobre el CÓDIGO del proyecto.
//
// Estrategia en dos niveles (D4 de la evaluación de plataforma):
//
//  1. Si el proyecto es un repo git → la lista de archivos sale de
//     `git ls-files` (tracked + untracked NO ignorados). Esto respeta
//     .gitignore de verdad: un log, un binario generado o un cache que el
//     usuario ya ignoró en git NO churnea el hash → no más staleness espuria.
//     También es más rápido: git ya tiene el índice, no recorremos node_modules.
//  2. Sin git (proyecto suelto, tests) → fallback al walk del árbol con la
//     lista fija de exclusiones de siempre. Degradación honesta, no un error.
//
// En ambos casos excluimos specforge/: registrar un verdict o aprobar un gate
// (que escriben estado) NO debe invalidar la frescura del resultado; solo un
// cambio en el CÓDIGO mueve el hash.
//
// Es content-hash (no mtime): dos ediciones distintas al mismo archivo dan
// hashes distintos.
func codeHash(projectDir string) string {
	if files, ok := gitListFiles(projectDir); ok {
		return hashFileList(projectDir, files)
	}
	return hashFileList(projectDir, walkListFiles(projectDir))
}

// gitListFiles pide a git la lista de archivos del proyecto: los trackeados
// (--cached) más los nuevos aún sin agregar (--others), excluyendo lo ignorado
// (--exclude-standard = .gitignore + excludes globales). -z separa con NUL para
// que un nombre con espacios o saltos de línea no rompa el parseo.
// ok=false si no hay git o projectDir no es un repo → el caller usa el fallback.
func gitListFiles(projectDir string) ([]string, bool) {
	cmd := exec.Command("git", "-C", projectDir, "ls-files", "-z", "--cached", "--others", "--exclude-standard")
	out, err := cmd.Output()
	if err != nil {
		return nil, false
	}
	var files []string
	for f := range strings.SplitSeq(string(out), "\x00") {
		if f == "" {
			continue
		}
		rel := filepath.ToSlash(f)
		// El estado de SpecForge está trackeado en git (las specs se commitean),
		// así que acá sí hay que filtrarlo a mano.
		if rel == "specforge" || strings.HasPrefix(rel, "specforge/") {
			continue
		}
		files = append(files, rel)
	}
	return files, true
}

// walkListFiles es el fallback sin git: recorre el árbol saltando los dirs de
// la lista fija (estado de SpecForge, VCS, dependencias, build).
func walkListFiles(projectDir string) []string {
	var files []string
	_ = filepath.WalkDir(projectDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // un archivo ilegible no debe abortar el hash entero
		}
		rel, rerr := filepath.Rel(projectDir, path)
		if rerr != nil {
			return nil
		}
		rel = filepath.ToSlash(rel)
		if d.IsDir() {
			if rel != "." && skipDirForHash(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		files = append(files, rel)
		return nil
	})
	return files
}

// hashFileList hashea ruta+contenido de cada archivo de la lista y computa el
// hash agregado. Archivos ilegibles o borrados (git --cached puede listar un
// archivo recién eliminado del working tree) se saltean en silencio.
func hashFileList(projectDir string, files []string) string {
	type entry struct {
		rel  string
		hash string
	}
	var entries []entry
	for _, rel := range files {
		data, err := os.ReadFile(filepath.Join(projectDir, filepath.FromSlash(rel)))
		if err != nil {
			continue
		}
		sum := sha256.Sum256(data)
		entries = append(entries, entry{rel: rel, hash: hex.EncodeToString(sum[:])})
	}
	// Orden estable por ruta → hash reproducible independiente del orden de
	// listado (git y el walk podrían diferir en criterio de orden).
	sort.Slice(entries, func(i, j int) bool { return entries[i].rel < entries[j].rel })
	roll := sha256.New()
	for _, e := range entries {
		fmt.Fprintf(roll, "%s\x00%s\n", e.rel, e.hash)
	}
	return hex.EncodeToString(roll.Sum(nil))
}

// skipDirsForHash: dirs que NO son "el código bajo test" — estado de SpecForge,
// VCS, dependencias y artefactos de build. Hashearlos sería ruido (y lento).
var skipDirsForHash = map[string]bool{
	"specforge": true, ".git": true,
	"node_modules": true, ".venv": true, "venv": true, "__pycache__": true,
	"target": true, "dist": true, "build": true, ".next": true,
	"vendor": true, ".idea": true, ".vscode": true, ".pytest_cache": true,
	".mypy_cache": true, ".ruff_cache": true, "coverage": true, ".tox": true,
}

func skipDirForHash(name string) bool { return skipDirsForHash[name] }

// ── Persistencia del resultado (specforge/.state/check/<feature>.json) ───────

func checkResultPath(projectDir, feature string) string {
	return filepath.Join(projectDir, "specforge", ".state", "check", feature+".json")
}

func writeCheckResult(projectDir, feature string, res checkResult) int {
	path := checkResultPath(projectDir, feature)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf check: %v\n", err)
		return 1
	}
	out, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf check: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf check: cannot write %s (%v)\n", path, err)
		return 1
	}
	return 0
}

// readCheckResult lee el último resultado registrado para una feature. La usa la
// Capa 2 (verdict gateado) para exigir verde y fresco. (found=false si no hay).
func readCheckResult(projectDir, feature string) (checkResult, bool) {
	var res checkResult
	data, err := os.ReadFile(checkResultPath(projectDir, feature))
	if err != nil {
		return res, false
	}
	if json.Unmarshal(data, &res) != nil {
		return res, false
	}
	return res, true
}

// ── Helpers chicos ───────────────────────────────────────────────────────────

// tailString devuelve los últimos n bytes de s (alineado al límite de línea si
// puede), para no guardar megabytes de output en el .json.
func tailString(s string, n int) string {
	if len(s) <= n {
		return s
	}
	tail := s[len(s)-n:]
	if i := strings.IndexByte(tail, '\n'); i >= 0 && i < len(tail)-1 {
		tail = tail[i+1:]
	}
	return "…(truncated)…\n" + tail
}

// short devuelve los primeros 12 caracteres de un hash (para logs legibles).
func short(hash string) string {
	if len(hash) <= 12 {
		return hash
	}
	return hash[:12]
}
