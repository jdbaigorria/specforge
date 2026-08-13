package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// Los tests de este paquete crean un repo git DE VERDAD en un temporal.
//
// Es la única forma honesta de probar un shell-out: un mock del comando
// probaría que sabemos escribir mocks, no que los argumentos que le pasamos a
// git son los correctos. Y son los argumentos lo único que este paquete tiene.

// repo arma un repositorio con un commit inicial y devuelve su ruta.
func repo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()

	// -c en vez de config global: no queremos tocar la configuración de quien
	// corra los tests, y git exige user.name/email para commitear.
	correrTest(t, dir, "init", "-b", "main")
	correrTest(t, dir, "config", "user.email", "test@ejemplo")
	correrTest(t, dir, "config", "user.name", "Test")

	escribir(t, dir, "uno.txt", "primero")
	correrTest(t, dir, "add", ".")
	correrTest(t, dir, "commit", "-m", "primero")

	return dir
}

func correrTest(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

func escribir(t *testing.T, dir, nombre, texto string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, nombre), []byte(texto), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestEsRepo(t *testing.T) {
	if !EsRepo(repo(t)) {
		t.Error("un repo recién creado dio que no es repo")
	}
	if EsRepo(t.TempDir()) {
		t.Error("un directorio pelado dio que es repo")
	}
}

func TestHeadDevuelveHashCorto(t *testing.T) {
	h, err := Head(repo(t))
	if err != nil {
		t.Fatal(err)
	}
	// El hash corto de git es de 7 caracteres, y puede crecer si hay colisión.
	// Lo que importa es que no sean los 40 del largo: el estado.json se lee en
	// un `git diff` y un hash largo lo hace ilegible.
	if len(h) < 7 || len(h) > 12 {
		t.Errorf("Head() = %q (%d chars), quería un hash corto", h, len(h))
	}
}

func TestBranchActual(t *testing.T) {
	b, err := BranchActual(repo(t))
	if err != nil {
		t.Fatal(err)
	}
	if b != "main" {
		t.Errorf("BranchActual() = %q, quería main", b)
	}
}

// El diff arranca en base_commit, que es el HEAD de cuando se terminó de
// planificar. Por eso muestra exactamente el trabajo de esta feature.
func TestDiffDesdeElBaseCommit(t *testing.T) {
	dir := repo(t)

	base, err := Head(dir)
	if err != nil {
		t.Fatal(err)
	}

	escribir(t, dir, "dos.txt", "el trabajo de la feature")
	correrTest(t, dir, "add", ".")
	correrTest(t, dir, "commit", "-m", "la feature")

	d, err := Diff(dir, base)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(d, "dos.txt") {
		t.Errorf("el diff no menciona el archivo nuevo:\n%s", d)
	}
	if strings.Contains(d, "uno.txt") {
		t.Errorf("el diff arrastra lo anterior al base_commit:\n%s", d)
	}
}

// El sobre sirve el resumen y no el diff completo: un diff entero de una
// feature grande es enorme, y meterlo en el contexto de un subagente es justo
// lo que el diseño evita.
func TestDiffResumenEsUnaLineaPorArchivo(t *testing.T) {
	dir := repo(t)
	base, _ := Head(dir)

	escribir(t, dir, "dos.txt", "a\nb\nc\n")
	correrTest(t, dir, "add", ".")
	correrTest(t, dir, "commit", "-m", "tres líneas")

	r, err := DiffResumen(dir, base)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r, "dos.txt") {
		t.Errorf("el resumen no nombra el archivo:\n%s", r)
	}
	// --stat trae el conteo, no el contenido.
	if strings.Contains(r, "+a") {
		t.Errorf("el resumen trae el contenido del diff:\n%s", r)
	}
}

// Sin base_commit no hay rango posible, y el mensaje tiene que decir por qué en
// vez de dejar que git falle con "ambiguous argument".
func TestDiffSinBaseCommitExplica(t *testing.T) {
	_, err := Diff(repo(t), "")
	if err == nil {
		t.Fatal("no dio error")
	}
	if !strings.Contains(err.Error(), "base_commit") {
		t.Errorf("el mensaje no explica qué falta: %v", err)
	}
}

// git escribe el motivo en stderr, y sin rescatarlo el error diría sólo
// "exit status 128", que no le sirve a nadie.
func TestElErrorDeGitLlegaConSuMensaje(t *testing.T) {
	_, err := Diff(repo(t), "noexisteestecommit")
	if err == nil {
		t.Fatal("un commit inexistente no dio error")
	}
	if !strings.Contains(err.Error(), "noexisteestecommit") {
		t.Errorf("el error perdió el mensaje de git: %v", err)
	}
}
