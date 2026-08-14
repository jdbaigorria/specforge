package arranque

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/constitucion"
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
)

// Los tests de este paquete escriben en un directorio temporal de verdad, igual
// que los de `maquina`, y por el mismo motivo: todo lo que hace `sf init` es
// tocar el disco. Un mock de filesystem probaría que la función llama a MkdirAll
// y no lo único que importa — que después el proyecto ARRANQUE.

func conArchivo(t *testing.T, raiz, nombre, contenido string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(raiz, nombre), []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestIniciarDejaElProyectoListo(t *testing.T) {
	raiz := t.TempDir()
	conArchivo(t, raiz, "go.mod", "module ejemplo\n")

	r, err := Iniciar(raiz)
	if err != nil {
		t.Fatalf("Iniciar: %v", err)
	}

	// Los dos directorios y los dos archivos.
	for _, ruta := range []string{docs.Base, docs.Backlog, docs.Constitucion, estado.Archivo} {
		if _, err := os.Stat(filepath.Join(raiz, ruta)); err != nil {
			t.Errorf("falta %s: %v", ruta, err)
		}
	}

	// Y NO los que aparecen sólo cuando hacen falta.
	for _, ruta := range []string{".docs/features", docs.Archivado} {
		if _, err := os.Stat(filepath.Join(raiz, ruta)); err == nil {
			t.Errorf("%s no debería existir todavía", ruta)
		}
	}

	if !r.Stack.Reconocido() {
		t.Fatal("no reconoció un proyecto con go.mod")
	}
}

// El test que importa de verdad: la constitución que escribe `sf init` tiene que
// poder LEERSE con el parser que usa el resto de sf. Comprobar que el archivo
// existe no alcanza — un frontmatter mal armado pasa esa prueba y rompe en el ⑧.
func TestLaCabeceraQueEscribeSeParsea(t *testing.T) {
	raiz := t.TempDir()
	conArchivo(t, raiz, "go.mod", "module ejemplo\n")

	if _, err := Iniciar(raiz); err != nil {
		t.Fatal(err)
	}

	c, err := constitucion.Leer(raiz)
	if err != nil {
		t.Fatalf("la constitución que escribió sf init no se puede leer: %v", err)
	}

	if c.Lenguaje != "go" {
		t.Errorf("lenguaje %q, quería go", c.Lenguaje)
	}
	if c.Manifiesto != "go.mod" {
		t.Errorf("manifiesto %q, quería go.mod", c.Manifiesto)
	}
	// Sin test_cmd caen tres compuertas. Es la única línea que sf exige.
	if c.TestCmd == "" {
		t.Error("test_cmd quedó vacío en un proyecto Go")
	}
	// Y el git: tiene que llegar entero, porque de ahí sale la branch.
	if !c.Git.BranchPorFeature || c.Git.PatronBranch == "" {
		t.Errorf("el bloque git: no llegó: %+v", c.Git)
	}
	if c.Git.Base() != "main" {
		t.Errorf("branch_base %q, quería main", c.Git.Base())
	}
}

func TestNoPisaUnProyectoYaIniciado(t *testing.T) {
	raiz := t.TempDir()
	if _, err := Iniciar(raiz); err != nil {
		t.Fatal(err)
	}

	// Se ensucia el estado para poder ver si lo pisó.
	e, err := estado.Leer(raiz)
	if err != nil {
		t.Fatal(err)
	}
	e.Producto.BriefSellado = "hacelo"
	if err := e.Guardar(raiz); err != nil {
		t.Fatal(err)
	}

	if _, err := Iniciar(raiz); !errors.Is(err, ErrYaIniciado) {
		t.Fatalf("segundo Iniciar: %v, quería ErrYaIniciado", err)
	}

	e2, err := estado.Leer(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if e2.Producto.BriefSellado != "hacelo" {
		t.Error("pisó el estado: el sello del ⑥ se perdió")
	}
}

// El caso del proyecto que perdió el estado.json pero conserva su constitución.
// Volver a correr `sf init` tiene que devolverle el estado SIN tirar el trabajo
// del ⑧.
func TestNoPisaUnaConstitucionQueYaExiste(t *testing.T) {
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, docs.Base), 0o755); err != nil {
		t.Fatal(err)
	}
	escrita := "---\nlenguaje: rust\ntest_cmd: cargo nextest run\n---\n\n# la mía\n"
	conArchivo(t, raiz, docs.Constitucion, escrita)

	if _, err := Iniciar(raiz); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(raiz, docs.Constitucion))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != escrita {
		t.Errorf("pisó la constitución existente:\n%s", b)
	}
}

func TestDetectarReconoceLosStacks(t *testing.T) {
	casos := []struct {
		archivo  string
		lenguaje string
	}{
		{"go.mod", "go"},
		{"Cargo.toml", "rust"},
		{"pyproject.toml", "python"},
		{"package.json", "node"},
		{"Gemfile", "ruby"},
		{"pom.xml", "java"},
	}

	for _, c := range casos {
		raiz := t.TempDir()
		conArchivo(t, raiz, c.archivo, "")

		s := Detectar(raiz)
		if s.Lenguaje != c.lenguaje {
			t.Errorf("%s → %q, quería %q", c.archivo, s.Lenguaje, c.lenguaje)
		}
		if s.TestCmd == "" {
			t.Errorf("%s no trajo test_cmd", c.archivo)
		}
	}
}

// El orden de la tabla es lo que hace la detección REPRODUCIBLE, y por eso se
// prueba: con un mapa, un proyecto que tiene dos manifiestos daría un lenguaje
// distinto en cada corrida.
func TestConDosManifiestosGanaSiempreElMismo(t *testing.T) {
	for i := range 20 {
		raiz := t.TempDir()
		conArchivo(t, raiz, "package.json", "{}")
		conArchivo(t, raiz, "pyproject.toml", "")

		if s := Detectar(raiz); s.Lenguaje != "python" {
			t.Fatalf("corrida %d: ganó %q, y la tabla pone python antes que node", i, s.Lenguaje)
		}
	}
}

func TestSinManifiestoAvisaEnElArchivo(t *testing.T) {
	raiz := t.TempDir()

	r, err := Iniciar(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if r.Stack.Reconocido() {
		t.Fatalf("reconoció un directorio vacío como %q", r.Stack.Lenguaje)
	}

	b, err := os.ReadFile(filepath.Join(raiz, docs.Constitucion))
	if err != nil {
		t.Fatal(err)
	}
	// El aviso va EN el archivo y no sólo en la salida del comando: el que abre
	// la constitución tres días después no vio esa salida.
	if !strings.Contains(string(b), "no reconocí el proyecto") {
		t.Errorf("no avisó que falta el test_cmd:\n%s", b)
	}
}
