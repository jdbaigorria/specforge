package tareas

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// ejemplo sale de artefactos.md §9, extendido a dos lotes para que los filtros
// tengan algo que separar.
const ejemplo = `{
  "feature": "f-1",
  "tareas": [
    {
      "id": "t-1", "lote": 1,
      "descripcion": "parsear el frontmatter del brief",
      "satisface": ["us-1/CA-1", "us-1/CA-2"],
      "tests": [
        "internal/docs/brief_test.go::TestParseFrontmatter",
        "internal/docs/brief_test.go::TestFrontmatterInvalido"
      ]
    },
    {
      "id": "t-2", "lote": 1,
      "descripcion": "el sello",
      "satisface": ["us-1/CA-2"],
      "tests": ["internal/docs/brief_test.go::TestBriefSinSello"]
    },
    {
      "id": "t-3", "lote": 2,
      "descripcion": "serializar el estado",
      "satisface": ["us-3/CA-1"],
      "tests": ["internal/estado/estado_test.go::TestJson"]
    }
  ]
}`

func escribir(t *testing.T, contenido string) (raiz, carpeta string) {
	t.Helper()

	raiz = t.TempDir()
	carpeta = filepath.Join(".docs", "features", "f-1-nucleo")
	if err := os.MkdirAll(filepath.Join(raiz, carpeta), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(raiz, carpeta, Archivo), []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
	return raiz, carpeta
}

func leerEjemplo(t *testing.T) *Plan {
	t.Helper()
	p, err := Leer(escribir(t, ejemplo))
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func TestLeer(t *testing.T) {
	p := leerEjemplo(t)

	if p.Feature != "f-1" {
		t.Errorf("Feature = %q, quería f-1", p.Feature)
	}
	if len(p.Tareas) != 3 {
		t.Fatalf("hay %d tareas, quería 3", len(p.Tareas))
	}
	if p.Tareas[0].Lote != 1 {
		t.Errorf("t-1 está en el lote %d, quería 1", p.Tareas[0].Lote)
	}
}

func TestLeerSinArchivoDaErrNoHay(t *testing.T) {
	_, err := Leer(t.TempDir(), "no/existe")
	if !errors.Is(err, ErrNoHay) {
		t.Fatalf("devolvió %v, quería ErrNoHay", err)
	}
}

// Filtrar por lote es lo que permite que `sf context` sirva sólo el lote que
// toca. Darle los cuatro al implementador le gasta contexto y lo tienta con
// trabajo que todavía no le toca.
func TestDelLote(t *testing.T) {
	p := leerEjemplo(t)

	if n := len(p.DelLote(1)); n != 2 {
		t.Errorf("el lote 1 tiene %d tareas, quería 2", n)
	}
	if n := len(p.DelLote(2)); n != 1 {
		t.Errorf("el lote 2 tiene %d tareas, quería 1", n)
	}
	if n := len(p.DelLote(9)); n != 0 {
		t.Errorf("un lote que no existe devolvió %d tareas", n)
	}
}

// El archivo lo escribe un LLM y puede venir con los lotes en cualquier orden.
func TestLotesVieneOrdenadoYSinRepetir(t *testing.T) {
	p, err := Leer(escribir(t, `{"tareas": [
		{"id":"t-1","lote":3},{"id":"t-2","lote":1},{"id":"t-3","lote":3},{"id":"t-4","lote":2}
	]}`))
	if err != nil {
		t.Fatal(err)
	}

	if got := p.Lotes(); !slices.Equal(got, []int{1, 2, 3}) {
		t.Errorf("Lotes() = %v, quería [1 2 3]", got)
	}
}

// Esta lista es toda la defensa contra el dolor #8: es lo que `sf lote start`
// va a exigir que falle y `sf done` que pase.
func TestTestsDelLoteJuntaYNoRepite(t *testing.T) {
	p := leerEjemplo(t)

	tests := p.TestsDelLote(1)
	if len(tests) != 3 {
		t.Fatalf("el lote 1 tiene %d tests, quería 3: %v", len(tests), tests)
	}
	if !slices.Contains(tests, "internal/docs/brief_test.go::TestBriefSinSello") {
		t.Errorf("falta el test de t-2: %v", tests)
	}
}

func TestTestsDelLoteNoRepiteElMismoTest(t *testing.T) {
	// Dos tareas del mismo lote pueden nombrar el mismo test: si las dos lo
	// necesitan, existe una sola vez.
	p, err := Leer(escribir(t, `{"tareas": [
		{"id":"t-1","lote":1,"tests":["a_test.go::TestX"]},
		{"id":"t-2","lote":1,"tests":["a_test.go::TestX","a_test.go::TestY"]}
	]}`))
	if err != nil {
		t.Fatal(err)
	}

	if got := p.TestsDelLote(1); len(got) != 2 {
		t.Errorf("TestsDelLote = %v, quería 2 sin repetir", got)
	}
}

// Contar criterios cubiertos es lo que ataca la feature a medias ANTES de
// empezar: "la feature tiene 9 criterios, las tareas cubren 7".
func TestCriteriosCubiertos(t *testing.T) {
	p := leerEjemplo(t)

	quiero := []string{"us-1/CA-1", "us-1/CA-2", "us-3/CA-1"}
	if got := p.CriteriosCubiertos(); !slices.Equal(got, quiero) {
		t.Errorf("CriteriosCubiertos() = %v, quería %v", got, quiero)
	}
}
