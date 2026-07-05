package main

import (
	"os/exec"
	"path/filepath"
	"testing"
)

// TestCodeHashFreshness: el hash cambia cuando cambia el código, pero NO cuando
// cambia el estado de SpecForge (specforge/ está excluido). Es la invariante que
// hace confiable el "verde y fresco" de la Capa 2.
func TestCodeHashFreshness(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "src", "main.go"), "package main\n")
	h1 := codeHash(proj)

	// Escribir estado de SpecForge no debe mover el hash.
	mustWrite(t, filepath.Join(proj, "specforge", "features.json"), `{"x":1}`)
	mustWrite(t, filepath.Join(proj, "specforge", ".state", "check", "x.json"), `{}`)
	if got := codeHash(proj); got != h1 {
		t.Error("escribir estado de specforge NO debería cambiar el code_hash")
	}

	// Tocar el código sí.
	mustWrite(t, filepath.Join(proj, "src", "main.go"), "package main\n// changed\n")
	if got := codeHash(proj); got == h1 {
		t.Error("cambiar el código SÍ debería cambiar el code_hash")
	}

	// node_modules y demás dirs de deps están excluidos.
	h2 := codeHash(proj)
	mustWrite(t, filepath.Join(proj, "node_modules", "junk", "a.js"), "lots of noise")
	if got := codeHash(proj); got != h2 {
		t.Error("node_modules debería estar excluido del code_hash")
	}
}

// TestCodeHashRespectsGitignore: en un repo git, la lista de archivos sale de
// `git ls-files` → lo ignorado por .gitignore NO churnea el hash (D4: adiós
// staleness espuria por logs/archivos generados), pero un archivo nuevo NO
// ignorado sí lo mueve (--others lo ve aunque no esté trackeado).
func TestCodeHashRespectsGitignore(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, ".gitignore"), "*.log\n")
	mustWrite(t, filepath.Join(proj, "src", "main.go"), "package main\n")
	gitInit(t, proj)

	h1 := codeHash(proj)

	// Un archivo ignorado (ni trackeado ni visible para --others) no mueve el hash.
	mustWrite(t, filepath.Join(proj, "debug.log"), "ruido de runtime")
	if got := codeHash(proj); got != h1 {
		t.Error("un archivo ignorado por .gitignore NO debería cambiar el code_hash")
	}

	// Un archivo nuevo no ignorado sí, aunque todavía no esté git-addeado.
	mustWrite(t, filepath.Join(proj, "src", "extra.go"), "package main\n")
	if got := codeHash(proj); got == h1 {
		t.Error("un archivo nuevo no ignorado SÍ debería cambiar el code_hash")
	}
}

// gitInit arma un repo git mínimo en dir (sin commits: ls-files --others ya ve
// los archivos). Si no hay git instalado, salteamos el test.
func gitInit(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", "-q", dir)
	if err := cmd.Run(); err != nil {
		t.Skipf("git no disponible: %v", err)
	}
}

// TestRunTestCommand: capturamos exit code y output combinados.
func TestRunTestCommand(t *testing.T) {
	proj := t.TempDir()
	if out, exit := runTestCommand(proj, "echo hola; exit 0"); exit != 0 || out == "" {
		t.Errorf("comando verde → exit 0 + output, got exit %d out %q", exit, out)
	}
	if _, exit := runTestCommand(proj, "exit 7"); exit != 7 {
		t.Errorf("debería propagar el exit code real, got %d", exit)
	}
}

// TestCheckRunRecords: check run escribe un resultado leíble por readCheckResult,
// con el exit code real y passed acorde.
func TestCheckRunRecords(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "constitution.json"),
		`{"schema_version":"1.0","identity_md":"x","build":{"mode":"inline","test_cmd":"exit 1"}}`)

	if code := runCheckRun([]string{"--feature=demo", proj}); code != 3 {
		t.Errorf("test rojo → exit 3 (recorded-but-fail), got %d", code)
	}
	res, ok := readCheckResult(proj, "demo")
	if !ok {
		t.Fatal("debería haberse registrado un check-result")
	}
	if res.Passed || res.ExitCode != 1 {
		t.Errorf("resultado debería reflejar el fallo real, got passed=%v exit=%d", res.Passed, res.ExitCode)
	}
	if res.CodeHash == "" {
		t.Error("el resultado debería sellar el code_hash")
	}
}

// TestCheckRunNoTestCmd: sin build.test_cmd → exit 2 (no se puede correr nada).
func TestCheckRunNoTestCmd(t *testing.T) {
	proj := t.TempDir()
	mustWrite(t, filepath.Join(proj, "specforge", "constitution.json"),
		`{"schema_version":"1.0","identity_md":"x"}`)
	if code := runCheckRun([]string{"--feature=demo", proj}); code != 2 {
		t.Errorf("sin test_cmd → exit 2, got %d", code)
	}
}
