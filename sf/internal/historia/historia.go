// Package historia lee un `us-#.md`: el requisito y sus criterios.
//
// ────────────────────────────────────────────────────────────────────────────
// LOS CRITERIOS CON ID SON EL MECANISMO DEL ㉑, Y LA JUGADA ES SUTIL
// ────────────────────────────────────────────────────────────────────────────
//
// La completitud contra una historia ES JUICIO, y ninguna herramienta
// determinista la contesta. Eso sigue siendo cierto. Pero hay una jugada
// intermedia, y es toda la defensa contra el dolor #7:
//
//	sf no verifica el juicio. Verifica que el juicio HAYA OCURRIDO, y sobre
//	TODOS.
//
// Con criterios sueltos en prosa, el revisor contesta "anda" y ahí se cuela el
// "terminado" con media historia. Con CA-1 / CA-2 / CA-3, tiene que contestar
// por cada uno — y sf cuenta:
//
//	"la historia tiene 3 criterios, el informe habla de 2"
//
// Contar no es juzgar. Es la única parte de este problema que una máquina puede
// hacer, y alcanza.
//
// ────────────────────────────────────────────────────────────────────────────
// SIN CAMPO `estado:`
// ────────────────────────────────────────────────────────────────────────────
//
// La primera versión del diseño ponía el estado de la historia acá y lo
// declaraba la fuente de verdad. Se cayó por dos razones: el ciclo corre por
// FEATURE y no por historia, y marcar un `.md` exigiría un LLM que reescriba
// prosa. El avance vive en estado.json (artefactos.md §7).
//
//	El us-# guarda el requisito, que no cambia porque el trabajo avance.
package historia

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/frontmatter"
)

// Historia es un us-# ya leído.
type Historia struct {
	Tipo   string `yaml:"tipo"` // "us" | "bug"
	ID     string `yaml:"id"`
	Titulo string `yaml:"titulo"`

	// DerivaDe y PrdVersion son la traza al PRD: si el PRD cambia después de
	// que las historias salieron de él, sf compara y avisa cuáles quedaron
	// viejas. Avisa, no frena.
	DerivaDe   string `yaml:"deriva_de"`
	PrdVersion string `yaml:"prd_version"`

	// RelacionadoA apunta al us-# original cuando esto es un bug (entrada C).
	RelacionadoA string `yaml:"relacionado_a"`

	// Criterios son los ids que aparecen en el cuerpo: CA-1, CA-2, …
	//
	// No vienen del frontmatter: se extraen de la prosa. El texto del criterio
	// lo lee el modelo; sf sólo necesita SABER CUÁNTOS SON Y CÓMO SE LLAMAN.
	Criterios []string
}

// EsBug decide si esta entrada va por el camino corto.
//
// El ruteo lo hace este campo y nada más (H20): un bug saltea `planificacion` y
// `revision`. No hay carril paralelo ni segunda máquina — hay un campo que
// saltea dos estados, y por eso el rastro no se pierde: el bug igual entró por
// el backlog, que es el embudo.
func (h *Historia) EsBug() bool { return h.Tipo == "bug" }

// reCriterio caza las líneas de criterio del cuerpo.
//
//   - **CA-1** — acepta `--json` y devuelve el estado serializado
//
// Se acepta con o sin negritas y con cualquier separador después del id, porque
// el que escribe esto es un LLM y el formato exacto varía. Lo que NO varía es el
// id: es lo único que sf necesita, y es lo que el revisor va a tener que
// nombrar uno por uno.
var reCriterio = regexp.MustCompile(`(?m)^\s*[-*]\s*\**\s*(CA-\d+)\b`)

// Leer carga una historia por id.
func Leer(raiz, id string) (*Historia, error) {
	ruta := filepath.Join(raiz, docs.Historia(id))

	var h Historia
	cuerpo, err := frontmatter.DeArchivo(ruta, &h)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no existe %s", docs.Historia(id))
		}
		return nil, fmt.Errorf("%s: %w", docs.Historia(id), err)
	}

	// FindAllSubmatch con -1 devuelve todas las coincidencias; el índice 1 es
	// el primer grupo de captura, que es el id.
	for _, m := range reCriterio.FindAllSubmatch(cuerpo, -1) {
		h.Criterios = append(h.Criterios, string(m[1]))
	}

	// El id del frontmatter puede faltar o estar mal escrito; el del nombre de
	// archivo es el que usaron el roadmap y las tareas para apuntarle. Gana ése.
	if h.ID == "" {
		h.ID = id
	}
	return &h, nil
}

// Criterios devuelve los criterios de varias historias, con el prefijo puesto.
//
//	us-1/CA-1 · us-1/CA-2 · us-3/CA-1
//
// Ese formato es el que usan `tareas.json` (campo `satisface`) y
// `revision.json` (campo `criterios`), y es lo que permite compararlos sin
// ambigüedad: dos historias pueden tener las dos un CA-1.
func Criterios(raiz string, ids []string) ([]string, error) {
	var todos []string
	for _, id := range ids {
		h, err := Leer(raiz, id)
		if err != nil {
			return nil, err
		}
		for _, c := range h.Criterios {
			todos = append(todos, h.ID+"/"+c)
		}
	}
	return todos, nil
}

// Ids devuelve los ids de todas las historias del backlog, ordenados.
//
// Ordenados por NÚMERO, no alfabéticamente: `us-10` va después de `us-9`, y un
// sort de strings los pondría al revés. Es el tipo de detalle que no se nota
// hasta la décima historia, y ahí se nota mucho.
func Ids(raiz string) []string {
	m, _ := filepath.Glob(filepath.Join(raiz, docs.Backlog, "us-*.md"))

	ids := make([]string, 0, len(m))
	for _, ruta := range m {
		ids = append(ids, strings.TrimSuffix(filepath.Base(ruta), ".md"))
	}
	slices.SortFunc(ids, func(a, b string) int { return numero(a) - numero(b) })
	return ids
}

// ProximoID es el id que le toca a la próxima historia.
//
// Qué id sigue tiene UNA SOLA RESPUESTA CORRECTA, así que lo hace sf y no el
// modelo (R1). Y no es sólo comodidad: dos historias con el mismo id romperían
// las referencias del roadmap y de `satisface`, que apuntan por id.
//
// Se toma el máximo + 1 y no la cantidad + 1, porque una historia archivada
// puede haber dejado un hueco y reusar su id pisaría las referencias viejas.
func ProximoID(raiz string) string {
	max := 0
	for _, id := range Ids(raiz) {
		if n := numero(id); n > max {
			max = n
		}
	}
	return fmt.Sprintf("us-%d", max+1)
}

// numero saca el entero de "us-12". Devuelve 0 si no tiene forma de id.
func numero(id string) int {
	_, num, hay := strings.Cut(id, "-")
	if !hay {
		return 0
	}
	n, err := strconv.Atoi(num)
	if err != nil {
		return 0
	}
	return n
}

// Esqueleto es el us-# que crea `sf new`: la cabecera puesta y el cuerpo vacío.
//
// sf pone lo que tiene una sola respuesta —el id, el tipo, de dónde deriva— y
// deja el cuerpo para el pinponeo. Los criterios NO se inventan acá: son juicio,
// y salen de la conversación con Javier.
func Esqueleto(id, prdHash string) string {
	return fmt.Sprintf(`---
tipo: us                    # us | bug
id: %s
titulo: ""
deriva_de: prd
prd_version: %s
relacionado_a: null         # el us-# original, cuando esto es un bug
---

# %s — <título>

Como **<quién>** quiero **<qué>** para **<por qué>**.

## Criterios de aceptación
- **CA-1** —

## Contexto
`, id, prdHash, id)
}
