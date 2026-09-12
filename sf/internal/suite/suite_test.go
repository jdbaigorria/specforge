package suite

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func conArchivo(t *testing.T, rel, texto string) string {
	t.Helper()
	raiz := t.TempDir()
	ruta := filepath.Join(raiz, rel)
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(texto), 0o644); err != nil {
		t.Fatal(err)
	}
	return raiz
}

// ────────────────────────────────────────────────────────────────────────────
// Correr — el exit code es toda la compuerta
// ────────────────────────────────────────────────────────────────────────────

func TestCorrerDistingueVerdeDeRojo(t *testing.T) {
	raiz := t.TempDir()

	r, err := Correr(raiz, "exit 0")
	if err != nil {
		t.Fatal(err)
	}
	if !r.Verde {
		t.Error("exit 0 no dio verde")
	}

	r, err = Correr(raiz, "exit 1")
	// Que los tests fallen NO es un error del programa: es un resultado normal,
	// y a veces es el que queremos (el rojo del ⑲).
	if err != nil {
		t.Fatalf("un exit 1 se reportó como error: %v", err)
	}
	if r.Verde {
		t.Error("exit 1 dio verde")
	}
}

// El test_cmd se ejecuta con la shell y no partido por espacios: los reales
// llevan pipes y comillas, y partirlos los rompe en silencio.
func TestCorrerRespetaLaShell(t *testing.T) {
	r, err := Correr(t.TempDir(), `echo "uno dos" | tr ' ' '-'`)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(r.Salida, "uno-dos") {
		t.Errorf("el pipe no se ejecutó: %q", r.Salida)
	}
}

func TestCorrerJuntaStdoutYStderr(t *testing.T) {
	r, _ := Correr(t.TempDir(), "echo salida; echo error >&2")
	if !strings.Contains(r.Salida, "salida") || !strings.Contains(r.Salida, "error") {
		t.Errorf("perdió una de las dos salidas: %q", r.Salida)
	}
}

func TestCorrerSinTestCmdAvisa(t *testing.T) {
	if _, err := Correr(t.TempDir(), "  "); err == nil {
		t.Error("sin test_cmd no dio error")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Faltantes — lo genérico que sí se puede comprobar
// ────────────────────────────────────────────────────────────────────────────

// Existe = el archivo está Y el nombre aparece adentro. Las dos cosas se
// verifican sin saber el lenguaje.
func TestFaltantesEncuentraLosQueEstan(t *testing.T) {
	raiz := conArchivo(t, "internal/docs/brief_test.go",
		"package docs\n\nfunc TestParseFrontmatter(t *testing.T) {}\n")

	faltan := Faltantes(raiz, []string{"internal/docs/brief_test.go::TestParseFrontmatter"})
	if len(faltan) != 0 {
		t.Errorf("dijo que falta uno que está: %v", faltan)
	}
}

// El caso que ataca el dolor #8: el archivo existe pero el test planificado no
// se escribió. Sin este chequeo, la suite falla por otra cosa y el lote pasa.
func TestFaltantesAtrapaElTestQueNoSeEscribio(t *testing.T) {
	raiz := conArchivo(t, "a_test.go", "package a\n\nfunc TestUno(t *testing.T) {}\n")

	faltan := Faltantes(raiz, []string{"a_test.go::TestUno", "a_test.go::TestDos"})
	if len(faltan) != 1 || !strings.Contains(faltan[0], "TestDos") {
		t.Errorf("faltantes = %v, quería sólo TestDos", faltan)
	}
}

func TestFaltantesAtrapaElArchivoQueNoExiste(t *testing.T) {
	faltan := Faltantes(t.TempDir(), []string{"no/existe_test.go::TestX"})
	if len(faltan) != 1 || !strings.Contains(faltan[0], "no existe el archivo") {
		t.Errorf("faltantes = %v", faltan)
	}
}

// Sin `::`, la tarea declaró el archivo entero: alcanza con que exista.
func TestFaltantesAceptaUnArchivoSinNombreDeTest(t *testing.T) {
	raiz := conArchivo(t, "a_test.go", "package a\n")

	if faltan := Faltantes(raiz, []string{"a_test.go"}); len(faltan) != 0 {
		t.Errorf("faltantes = %v, quería ninguno", faltan)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Archivos y Partir
// ────────────────────────────────────────────────────────────────────────────

// Es lo que se hashea, así que tiene que ser estable: sin repetir y ordenado.
func TestArchivosSinRepetirYOrdenado(t *testing.T) {
	got := Archivos([]string{
		"b_test.go::TestZ",
		"a_test.go::TestX",
		"a_test.go::TestY",
	})
	if !slices.Equal(got, []string{"a_test.go", "b_test.go"}) {
		t.Errorf("Archivos() = %v", got)
	}
}

func TestPartir(t *testing.T) {
	a, n := Partir("internal/docs/brief_test.go::TestParseFrontmatter")
	if a != "internal/docs/brief_test.go" || n != "TestParseFrontmatter" {
		t.Errorf("Partir() = %q, %q", a, n)
	}

	a, n = Partir("a_test.go")
	if a != "a_test.go" || n != "" {
		t.Errorf("sin :: → %q, %q", a, n)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Nombrados — ¿la salida habla de ESTOS tests?
// ────────────────────────────────────────────────────────────────────────────

// Las salidas son las de verdad, recortadas: lo que se prueba es que el Contains
// sobrevive a cómo imprime cada runner, y eso no se prueba con strings inventados.
func TestNombradosReconoceLosRunnersReales(t *testing.T) {
	plan := []string{"a_test.go::TestUno", "a_test.go::TestDos"}

	casos := []struct {
		runner string
		salida string
		quiero int
	}{
		{"go test", "--- FAIL: TestUno (0.00s)\n    a_test.go:12: quería 5\nFAIL", 1},
		{"pytest", "FAILED a_test.py::TestUno - AssertionError: assert 4 == 5", 1},
		{"jest", "  ✕ TestUno (3 ms)\n  ✓ otra cosa", 1},
		{"cargo", "failures:\n    TestUno\n\ntest result: FAILED", 1},
		{"los dos", "--- FAIL: TestUno\n--- FAIL: TestDos", 2},

		// El caso que motiva todo: la suite falla por otra cosa.
		{"un fallo ajeno", "--- FAIL: TestDeOtroPaquete (0.01s)\nFAIL\texit status 1", 0},
		{"no compila", "# proyecto/otro\notro.go:9:2: undefined: Foo", 0},
		{"salida recortada", "", 0},
	}
	for _, c := range casos {
		t.Run(c.runner, func(t *testing.T) {
			if got := Nombrados(c.salida, plan); len(got) != c.quiero {
				t.Errorf("Nombrados = %v (%d), quería %d", got, len(got), c.quiero)
			}
		})
	}
}

// Un test declarado sin `::` es el archivo entero, y se busca por su basename:
// los runners imprimen las rutas con separadores distintos y el nombre del
// archivo es la parte que sobrevive a todos.
func TestNombradosBuscaPorArchivoCuandoNoHayNombre(t *testing.T) {
	plan := []string{"tests/unit/a_test.go"}

	if got := Nombrados(`--- FAIL: tests\unit\a_test.go`, plan); len(got) != 1 {
		t.Errorf("Nombrados = %v, quería encontrarlo por basename", got)
	}
	if got := Nombrados("--- FAIL: b_test.go", plan); len(got) != 0 {
		t.Errorf("Nombrados = %v, quería vacío", got)
	}
}
