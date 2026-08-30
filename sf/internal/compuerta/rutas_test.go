package compuerta

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// conArchivos crea un proyecto de mentira con los archivos que se le pidan.
func conArchivos(t *testing.T, rutas ...string) string {
	t.Helper()
	raiz := t.TempDir()
	for _, r := range rutas {
		ruta := filepath.Join(raiz, r)
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, nil, 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return raiz
}

// EL CASO DE LA CORRIDA REAL. El plan nombró los tests dentro del `.sql` que
// probaban; el implementador los escribió en el `.test.ts` que los ejecuta, y
// `sf lote start` reportó los once como faltantes.
func TestUnTestPlanificadoEnUnSqlNoPasaElPlan(t *testing.T) {
	raiz := conArchivos(t,
		"src/lib/sala.ts",
		"src/test/sala.test.ts",
		"supabase/migrations/001_init.sql",
	)

	var r Resultado
	revisarRutasDeTest(raiz, []tareaConTests{{
		ID:    "t-3",
		Tests: []string{"supabase/migrations/001_init.sql::rlsBloqueaAlAnonimo"},
	}}, &r)

	if r.Pasa() {
		t.Fatal("dejó pasar un test planificado en un .sql")
	}
	f := strings.Join(r.Fallas, " ")
	if !strings.Contains(f, "001_init.sql") {
		t.Errorf("no dijo cuál archivo: %v", r.Fallas)
	}
	if !strings.Contains(f, ".ts") {
		t.Errorf("no dijo cuál es la extensión buena de este proyecto: %v", r.Fallas)
	}
}

// En un proyecto que todavía no tiene tests no se puede saber, y callarse es lo
// correcto: es la primera feature y no hay contra qué comparar.
func TestSinTestsEnElRepoNoDiceNada(t *testing.T) {
	raiz := conArchivos(t, "src/lib/sala.ts", "supabase/migrations/001_init.sql")

	var r Resultado
	revisarRutasDeTest(raiz, []tareaConTests{{
		ID: "t-1", Tests: []string{"supabase/migrations/001_init.sql::x"},
	}}, &r)

	if !r.Pasa() {
		t.Errorf("opinó sin tener con qué comparar: %v", r.Fallas)
	}
}

// El camino feliz: los tests del plan van donde van los tests del proyecto.
func TestUnTestEnLaExtensionDelProyectoPasa(t *testing.T) {
	raiz := conArchivos(t, "src/test/sala.test.ts")

	var r Resultado
	revisarRutasDeTest(raiz, []tareaConTests{{
		ID: "t-1", Tests: []string{"src/test/migration-f1.test.ts::rlsBloquea"},
	}}, &r)

	if !r.Pasa() {
		t.Errorf("frenó un test que va donde tiene que ir: %v", r.Fallas)
	}
}

// Once tests en el mismo `.sql` son UNA falla y no once: el que la lee ya sabe
// dónde tocar, y once líneas iguales son ruido.
func TestOnceTestsEnElMismoArchivoDanUnaSolaFalla(t *testing.T) {
	raiz := conArchivos(t, "a.test.ts")

	var tests []string
	for i := 0; i < 11; i++ {
		tests = append(tests, "supabase/migrations/001.sql::test")
	}
	var r Resultado
	revisarRutasDeTest(raiz, []tareaConTests{{ID: "t-1", Tests: tests}}, &r)

	if len(r.Fallas) != 1 {
		t.Errorf("%d fallas, quería 1: %v", len(r.Fallas), r.Fallas)
	}
}

// Un archivo declarado entero, sin `::nombre`, recibe el mismo trato: lo que se
// mira es la extensión, no el sufijo.
func TestUnArchivoSinNombreDeTestRecibeElMismoTrato(t *testing.T) {
	raiz := conArchivos(t, "a.test.ts")

	var r Resultado
	revisarRutasDeTest(raiz, []tareaConTests{{
		ID: "t-1", Tests: []string{"supabase/migrations/001.sql"},
	}}, &r)

	if r.Pasa() {
		t.Error("dejó pasar un archivo entero con la extensión equivocada")
	}
}

// Un proyecto Go reconoce `_test.go`, uno de JS `.test.ts` y `.spec.js`. La
// lista no pretende ser exhaustiva: alcanza con reconocer los del proyecto que
// se está mirando.
func TestReconoceLasTresMarcasDeTest(t *testing.T) {
	for _, c := range []struct{ archivo, ext string }{
		{"internal/x/y_test.go", ".go"},
		{"src/a.test.ts", ".ts"},
		{"src/a.spec.js", ".js"},
		{"spec/models/user_spec.rb", ".rb"},
	} {
		t.Run(c.archivo, func(t *testing.T) {
			exts := extensionesDeTest(conArchivos(t, c.archivo))
			if len(exts) != 1 || exts[0] != c.ext {
				t.Errorf("= %v, quería [%s]", exts, c.ext)
			}
		})
	}
}

// node_modules no se recorre: una compuerta que tarda segundos es una compuerta
// que alguien va a querer apagar.
func TestNoSeMeteEnNodeModules(t *testing.T) {
	raiz := conArchivos(t, "node_modules/pkg/index.test.js", "src/a.test.ts")

	exts := extensionesDeTest(raiz)
	for _, e := range exts {
		if e == ".js" {
			t.Errorf("se metió en node_modules: %v", exts)
		}
	}
}
