// El chequeo de H5: que la ruta de un test planificado sea la del TEST y no la
// de lo que el test prueba.
//
// ────────────────────────────────────────────────────────────────────────────
// LA COMPUERTA HIZO BIEN Y EL ⑯ PIDIÓ MAL
// ────────────────────────────────────────────────────────────────────────────
//
// En la primera corrida real fuera de Claude Code, el plan trajo los 11 tests
// del lote como:
//
//	supabase/migrations/001_init.sql::elNombreDelTest
//
// El implementador los escribió en `src/test/migration-f1.test.ts` con esos
// nombres exactos, y `sf lote start` reportó los 11 como faltantes.
//
// `suite.Faltantes` había hecho exactamente lo correcto: busca el nombre adentro
// del archivo que el plan declaró. Si el plan dice `.sql`, mira el `.sql`. Y no
// parsea la salida del runner a propósito —cada runner imprime distinto y un
// parser por runner es una lista que se pudre—.
//
// El error estaba arriba. `sf-plan` decía `"tests": ["path/to/file_test.go::TestName"]`
// y nunca decía que la ruta es DÓNDE VA A VIVIR el test, no qué cosa prueba.
// Frente a una migración —donde el sujeto es un archivo y el test es otro— el
// planificador escribió la ruta del sujeto. Es la lectura natural de un ejemplo
// donde las dos cosas coinciden.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ SE MIRA LA EXTENSIÓN Y NADA MÁS
// ────────────────────────────────────────────────────────────────────────────
//
// La tentación es saber qué runner usa el proyecto y validar contra eso. No se
// hace: sería una tabla de lenguajes y runners que hay que mantener, y es la
// misma tentación que `suite.go` ya rechazó por el mismo motivo.
//
// Lo que se hace es mirar LO QUE EL REPO YA TIENE. Si en este proyecto los tests
// son `.ts`, un test planificado en un `.sql` no va a correr, y eso se sabe sin
// saber nada de TypeScript ni de SQL. En un proyecto que todavía no tiene ningún
// test, no se dice nada — no se puede, y callarse es lo correcto.
package compuerta

import (
	"io/fs"
	"path/filepath"
	"slices"
	"strings"
)

// marcasDeTest son los infijos que hacen que un archivo sea "de test".
//
// Es una lista y no una regex porque son literalmente estos: los tres que usan
// los ecosistemas que existen. No pretende ser exhaustiva y no hace falta que lo
// sea — con que reconozca los archivos de test del proyecto que se está mirando,
// alcanza.
var marcasDeTest = []string{"_test.", ".test.", ".spec.", "_spec."}

// ignoradas son las carpetas que no se recorren.
//
// Sin esto, un `node_modules` con veinte mil archivos hace que una compuerta que
// tiene que contestar en milisegundos se tome segundos.
var ignoradas = []string{"node_modules", "vendor", ".git", "dist", "build", "target", ".docs"}

// extensionesDeTest son las extensiones de los archivos de test que YA existen
// en el proyecto.
//
// Devuelve vacío cuando no hay ninguno, y ése es un resultado legítimo y no un
// error: un proyecto en su primera feature todavía no tiene tests.
func extensionesDeTest(raiz string) []string {
	var exts []string
	// El recorrido se corta a las 400 entradas: es una compuerta, no un
	// indexador. Con 400 archivos ya se vio de qué se trata el proyecto, y si el
	// primer test aparece más allá de eso, callarse es mejor que tardar.
	visto := 0
	filepath.WalkDir(raiz, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil //nolint:nilerr // un directorio ilegible no es asunto de esta compuerta
		}
		if d.IsDir() {
			if slices.Contains(ignoradas, d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if visto++; visto > 400 {
			return filepath.SkipAll
		}
		nombre := d.Name()
		for _, m := range marcasDeTest {
			if strings.Contains(nombre, m) {
				if e := filepath.Ext(nombre); e != "" && !slices.Contains(exts, e) {
					exts = append(exts, e)
				}
				break
			}
		}
		return nil
	})
	return exts
}

// revisarRutasDeTest agrega una falla por cada test planificado en un archivo
// que en este proyecto no podría ser un test.
//
// Se reporta por TAREA y no por test, igual que el conteo de criterios: si una
// tarea nombra once tests en el mismo `.sql`, once fallas idénticas son ruido y
// el que lo lee ya sabe dónde tocar con una.
func revisarRutasDeTest(raiz string, tareas []tareaConTests, r *Resultado) {
	exts := extensionesDeTest(raiz)
	if len(exts) == 0 {
		return // proyecto sin tests todavía: no hay contra qué comparar
	}

	for _, t := range tareas {
		malas := map[string]bool{}
		for _, id := range t.Tests {
			archivo, _, _ := strings.Cut(id, "::")
			if archivo == "" {
				continue
			}
			if e := filepath.Ext(archivo); e != "" && !slices.Contains(exts, e) {
				malas[archivo] = true
			}
		}
		if len(malas) == 0 {
			continue
		}
		nombres := make([]string, 0, len(malas))
		for a := range malas {
			nombres = append(nombres, a)
		}
		slices.Sort(nombres)
		r.falla("la tarea %s nombra tests en %s, y en este proyecto los tests son %s. "+
			"La ruta es dónde VIVE el test, no qué cosa prueba",
			t.ID, strings.Join(nombres, ", "), strings.Join(exts, ", "))
	}
}

// tareaConTests es lo mínimo que este chequeo necesita de una tarea.
//
// Se declara acá en vez de recibir `[]tareas.Tarea` para que el chequeo se pueda
// testear sin construir un plan entero, que es lo que hace que un test de esto
// quepa en diez líneas.
type tareaConTests struct {
	ID    string
	Tests []string
}
