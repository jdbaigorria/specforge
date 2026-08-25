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

	// SinTexto son los criterios que tienen id y no dicen nada.
	//
	// ────────────────────────────────────────────────────────────────────
	// UN ID VACÍO CUENTA IGUAL, Y ESO ERA UN AGUJERO
	// ────────────────────────────────────────────────────────────────────
	//
	// El esqueleto que deja `sf new` trae la línea puesta y el texto no:
	//
	//	- **CA-1** —
	//
	// Como el id existe, la compuerta del ⑨ lo contaba y el backlog pasaba.
	// Y contar un criterio que no dice nada es peor que no contarlo: el ⑰ lo
	// da por cubierto y el ㉑ le pone veredicto, así que el mecanismo de los
	// ids sigue en pie sobre algo que nadie puede juzgar.
	//
	// Sigue siendo contar caracteres, no leer: sf no opina sobre si el texto
	// es bueno — mira si hay texto.
	SinTexto []string
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
// El segundo grupo es el resto de la línea, y sirve para saber si el criterio
// dice algo o es el hueco que dejó `sf new`.
var reCriterio = regexp.MustCompile(`(?m)^\s*[-*]\s*\**\s*(CA-\d+)\b(.*)$`)

// separadores es lo que va entre el id y el texto, y no es texto.
//
// Varía porque el que escribe esto es un LLM: em dash, guión, dos puntos, y las
// negritas que cierran. Lo que queda después de sacarlos es el criterio.
const separadores = "*—–-:. \t"

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
	// el primer grupo de captura, que es el id, y el 2 es el resto de la línea.
	for _, m := range reCriterio.FindAllSubmatch(cuerpo, -1) {
		id := string(m[1])
		h.Criterios = append(h.Criterios, id)
		if strings.Trim(string(m[2]), separadores) == "" {
			h.SinTexto = append(h.SinTexto, id)
		}
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

// Las marcas del esqueleto son el checkpoint del ⑨, igual que
// constitucion.MarcaSinEscribir lo es del ⑧.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ UN MARCADOR Y NO "¿TIENE TÍTULO?"
// ────────────────────────────────────────────────────────────────────────────
//
// El primer intento miraba `titulo:` en el frontmatter, y da falsos positivos:
// una historia escrita a mano puede tener el título sólo en el encabezado y
// estar perfectamente completa. Un checkpoint que frena trabajo terminado es
// peor que no tenerlo — se aprende a ignorarlo.
//
// El marcador no puede equivocarse: o está el hueco que dejó `sf new`, o no.
// Y las escribe el MISMO paquete que después las busca, así que no se pueden
// desincronizar.
const (
	marcaTitulo = "<título>"
	marcaQuien  = "<quién>"
	marcaQue    = "<qué>"
	marcaPorQue = "<por qué>"
)

var marcasDelEsqueleto = []string{marcaTitulo, marcaQuien, marcaQue, marcaPorQue}

// SinPinponear son las historias que todavía son el esqueleto de `sf new`.
//
// ────────────────────────────────────────────────────────────────────────────
// EL ⑨ TAMBIÉN NECESITA SABER CUÁNDO TERMINÓ, NO SÓLO CUÁNDO EMPEZÓ
// ────────────────────────────────────────────────────────────────────────────
//
// `sf new` mete una entrada al backlog y reabre la ⏸ del ⑨ para que se
// pinponee. Pero el checkpoint preguntaba "¿hay alguna historia?", y con un
// producto en marcha la respuesta es siempre sí — así que `sf next` mostraba la
// ⏸ y nunca mandaba a `sfp-backlog`. Lo que quedaba para aprobar era un
// esqueleto con el título vacío y un `CA-1` que no dice nada.
//
// Es el mismo patrón que `huerfanas` para el ⑩ —preguntar qué falta, no si hay
// algo— y es la misma forma que fallaba en el ⑧: un checkpoint que pregunta si
// existe cuando la pregunta es si está terminado.
//
// El criterio es contable y no de gusto: sin criterios, con criterios que sólo
// tienen id, o con alguno de los huecos que dejó el esqueleto todavía puesto.
// sf no opina sobre si la historia está BIEN escrita — mira si está escrita.
func SinPinponear(raiz string) []string {
	var faltan []string
	for _, id := range Ids(raiz) {
		h, err := Leer(raiz, id)
		if err != nil {
			// Una historia ilegible es un problema del ⑨ igual: mandarla a
			// pinponear es la respuesta útil, y de lo que esté roto se queja la
			// compuerta, que corre después.
			faltan = append(faltan, id)
			continue
		}
		if len(h.Criterios) == 0 || len(h.SinTexto) > 0 || tieneMarcas(raiz, id) {
			faltan = append(faltan, id)
		}
	}
	return faltan
}

// tieneMarcas dice si el cuerpo todavía tiene alguno de los huecos del esqueleto.
func tieneMarcas(raiz, id string) bool {
	b, err := os.ReadFile(filepath.Join(raiz, docs.Historia(id)))
	if err != nil {
		return false // ya se contó como faltante arriba
	}
	texto := string(b)
	for _, m := range marcasDelEsqueleto {
		if strings.Contains(texto, m) {
			return true
		}
	}
	return false
}

// SonTodasBugs dice si un conjunto de historias va por el camino corto.
//
// ────────────────────────────────────────────────────────────────────────────
// TODAS, Y NO "ALGUNA"
// ────────────────────────────────────────────────────────────────────────────
//
// Saltear la planificación de una feature que mezcla un bug con dos historias
// nuevas dejaría esas dos sin diseño. Y si están mezcladas, el agrupamiento del
// ⑩ ya estaba mal: un bug y una feature nueva no comparten solución técnica
// (artefactos.md §3).
//
// Vive acá y no en `maquina` porque la pregunta es sobre las historias, y
// porque el que la necesita ya no es uno solo: `sf next`, `sf done`, `sf lote
// start` y `sf context` tienen que contestarla igual. Recibe []string y no
// roadmap.Feature para que `historia` no dependa de `roadmap`.
//
// Una historia ilegible cuenta como "no es bug": el que se queja de eso es la
// compuerta del ⑨, y ante la duda conviene el camino largo — de más se puede
// saltear después, de menos ya se implementó sin diseño.
func SonTodasBugs(raiz string, ids []string) bool {
	if len(ids) == 0 {
		return false
	}
	for _, id := range ids {
		h, err := Leer(raiz, id)
		if err != nil || !h.EsBug() {
			return false
		}
	}
	return true
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

# %s — %s

Como **%s** quiero **%s** para **%s**.

## Criterios de aceptación
- **CA-1** —

## Contexto
`, id, prdHash, id,
		marcaTitulo, marcaQuien, marcaQue, marcaPorQue)
}
