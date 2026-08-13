// Package sobre arma lo que `sf context` le sirve al que va a trabajar.
//
// ────────────────────────────────────────────────────────────────────────────
// DE DÓNDE SALE ESTA IDEA
// ────────────────────────────────────────────────────────────────────────────
//
// Era de Javier y estaba guardada como "no ahora":
//
//	"el cli no sólo indicaría dónde estás, qué sigue, podés avanzar, sino que
//	 tendría la capacidad de inyectar el contexto adecuado a los subagentes.
//	 Porque cuando un subagente arranque le pide al cli: dame la constitución
//	 del proyecto y el spec."
//
// La ronda 7 la ascendió a MECANISMO del ⑱, que es el primer traspaso a otro
// modelo. `flujo-real.md` dice que a partir de ahí "la propuesta tiene que
// bastarse sola" — y si la lista de qué leer la arma el orquestador a mano,
// cada vez se olvida algo distinto.
//
// ────────────────────────────────────────────────────────────────────────────
// SIN ARGUMENTOS, Y ES LO QUE LO HACE SEGURO (H2)
// ────────────────────────────────────────────────────────────────────────────
//
// En los documentos aparecieron tres formas para lo mismo —`sf context
// implement f-1 --lote 2`, `sf context documentar f-1`, `sf context planificar
// f-3`— y el problema no era el nombre: era que los tres argumentos son
// DEDUCIBLES. El estado, la feature y el lote están todos en estado.json.
//
//	R6 aplicada a la CLI: un argumento deducible es un argumento que se pasa mal.
//
// Un subagente que escribe `--lote 2` cuando va el 3 recibe el sobre
// equivocado y NO SE ENTERA. Sin argumento no hay forma de errarle.
//
// ────────────────────────────────────────────────────────────────────────────
// QUÉ NO ENTRA EN CADA SOBRE, QUE ES LA MITAD DEL DISEÑO
// ────────────────────────────────────────────────────────────────────────────
//
// El sobre del ⑱ NO lleva `decision.md`, y no es un olvido. Duda de Javier, y
// era buena: "¿no le sirve al implementador saber por qué se descartaron las
// otras, o le ocasionaría confusión?". Las dos cosas son ciertas a la vez, así
// que se separan:
//
//	Al implementador le sirve la RESTRICCIÓN, no la ALTERNATIVA.
//
// Sin saber que A se descartó, se le puede ocurrir A solo y hacerla. Pero con
// las tres opciones adentro, el que mezcla es él — y justo el ⑱ es donde entra
// el modelo del dolor #7. Por eso la conclusión de cada descarte baja al
// `spec-design.md` en una línea, y el debate se queda en `decision.md`.
package sobre

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
	"github.com/jdbaigorria/specforge/sf/internal/tareas"
)

// Sobre es lo que se le entrega al que va a trabajar.
type Sobre struct {
	Estado  string
	Feature string
	Lote    int
	Partes  []Parte
}

// Parte es una sección del sobre.
//
// Tiene DOS formas de traer material, y la distinción no es cosmética:
//
//	Rutas      archivos que existen → el que trabaja los abre con sus
//	           herramientas, y aprovecha el caché de su harness
//	Contenido  material DERIVADO que no es un archivo: el lote de tareas.json
//	           filtrado, el diff de git, la corrida de mutantes
//
// Servir rutas por defecto es más barato y le deja al que trabaja decidir
// cuánto lee. Pero hay cosas que no tienen ruta —no se puede apuntar a "sólo el
// lote 2 de este JSON"— y ésas viajan embebidas.
type Parte struct {
	Titulo    string
	Rutas     []string
	Contenido string

	// Falta explica por qué esta parte no se pudo servir.
	//
	// Existe para no mentir por omisión. Un sobre al que le falta el diff
	// porque el repo no es git tiene que DECIRLO: si la parte simplemente no
	// apareciera, el que trabaja creería que no hacía falta.
	Falta string
}

// Armar arma el sobre del estado actual.
//
// Recibe todo ya leído —estado y roadmap— en vez de leerlo: así los tests no
// necesitan escribir un estado.json para probar cada sobre, y este paquete no
// duplica el manejo de "no hay archivo" que ya está resuelto en los otros.
func Armar(raiz string, e *estado.Estado, r *roadmap.Roadmap) (*Sobre, error) {
	// Los cinco de producto no dependen de ninguna feature, así que se
	// resuelven mirando sólo los sellos. El orden es el mismo que el de
	// maquina.Siguiente, y por la misma razón: es el orden del flujo.
	if s, hay := deProducto(e); hay {
		return s, nil
	}
	return deFeature(raiz, e, r)
}

// ────────────────────────────────────────────────────────────────────────────
// Producto
// ────────────────────────────────────────────────────────────────────────────

func deProducto(e *estado.Estado) (*Sobre, bool) {
	switch {
	case e.Producto.BriefSellado == "":
		// El brief es el único sobre VACÍO del flujo, y está bien que lo sea:
		// ①–⑤ es un pinponeo con Javier sobre una idea que todavía no tiene
		// forma. No hay nada previo que leer — ése es el punto de arrancar.
		return &Sobre{
			Estado: "brief",
			Partes: []Parte{{
				Titulo:    "Sin sobre",
				Contenido: "El brief arranca de cero: es un pinponeo con Javier, no una lectura.",
			}},
		}, true

	case e.Producto.PrdHash == "":
		return &Sobre{Estado: "prd", Partes: []Parte{
			{Titulo: "De dónde sale el PRD", Rutas: []string{docs.Brief}},
		}}, true

	case !e.Producto.ConstitucionSellada:
		return &Sobre{Estado: "constitucion", Partes: []Parte{
			{Titulo: "Qué hay que construir", Rutas: []string{docs.PRD}},
			// La cabecera técnica (lenguaje, manifiesto, test_cmd) ya está
			// llena: la puso detectStack() en `sf init`. El que escribe acá
			// sólo pone el cuerpo — arquitectura, stack y por qué,
			// convenciones— que es lo único con varias respuestas (H18).
			{Titulo: "La cabecera técnica ya está llena", Rutas: []string{docs.Constitucion}},
		}}, true

	case !e.Producto.BacklogVisto:
		return &Sobre{Estado: "backlog", Partes: []Parte{
			{Titulo: "Qué hay que partir en historias", Rutas: []string{docs.PRD}},
			{Titulo: "Las reglas del proyecto", Rutas: []string{docs.Constitucion}},
		}}, true
	}
	return nil, false
}

// ────────────────────────────────────────────────────────────────────────────
// Feature
// ────────────────────────────────────────────────────────────────────────────

func deFeature(raiz string, e *estado.Estado, r *roadmap.Roadmap) (*Sobre, error) {
	// El ⑩ es el último de producto pero necesita el backlog entero, así que
	// se resuelve acá donde ya hay glob a mano.
	if r == nil {
		return &Sobre{Estado: "roadmap", Partes: []Parte{
			{Titulo: "Las historias a agrupar y ordenar", Rutas: historias(raiz)},
			{Titulo: "Las reglas del proyecto", Rutas: []string{docs.Constitucion}},
		}}, nil
	}

	f, hay := e.Actual()
	if !hay {
		return nil, fmt.Errorf("no hay feature en curso: corré `sf take <feature>` primero")
	}
	fr, enRoadmap := r.Buscar(e.FeatureActual)
	if !enRoadmap {
		return nil, fmt.Errorf("%s no está en el roadmap", e.FeatureActual)
	}
	carpeta := fr.Carpeta()

	s := &Sobre{Estado: f.Estado, Feature: fr.ID}

	switch f.Estado {
	case estado.Planificacion:
		s.Partes = []Parte{
			{Titulo: "El manual del proyecto", Rutas: []string{docs.Constitucion}},
			{Titulo: "Qué hay que resolver", Rutas: rutasDeHistorias(fr.Historias)},
			aprendizajes(raiz),
		}

	case estado.Implementar:
		s.Partes = []Parte{
			{Titulo: "El manual del proyecto", Rutas: []string{docs.Constitucion}},
			{Titulo: "Qué hay que construir y cómo", Rutas: []string{filepath.Join(carpeta, docs.Spec)}},
			// decision.md NO va acá. Ver el comentario de arriba del paquete.
		}
		s.Lote, s.Partes = loteYTareas(raiz, carpeta, f, s.Partes)
		s.Partes = append(s.Partes,
			Parte{Titulo: "Los criterios que tenés que satisfacer", Rutas: rutasDeHistorias(fr.Historias)})

	case estado.Revision:
		s.Partes = []Parte{
			{Titulo: "El manual del proyecto", Rutas: []string{docs.Constitucion}},
			{Titulo: "Qué se había planeado", Rutas: []string{filepath.Join(carpeta, docs.Spec)}},
			{Titulo: "Los criterios, y hay que opinar sobre TODOS", Rutas: rutasDeHistorias(fr.Historias)},
			diff(raiz, f.BaseCommit, "El código que quedó"),
			mutantes(),
		}

	case estado.Cierre:
		// La doc tiene dos mitades y sólo una sale del código. Por eso el sobre
		// trae las dos fuentes: el diff da la técnica, los us-# dan la
		// funcional — qué se pedía y para quién (que-sobrevive.md §11).
		s.Partes = []Parte{
			diff(raiz, f.BaseCommit, "El código que quedó → la mitad TÉCNICA"),
			{Titulo: "Qué se pedía → la mitad FUNCIONAL", Rutas: rutasDeHistorias(fr.Historias)},
			{Titulo: "Cómo se resolvió", Rutas: []string{filepath.Join(carpeta, docs.Spec)}},
			aprendizajes(raiz),
		}

	default:
		return nil, fmt.Errorf("%s está en %q y ese estado no pide sobre", fr.ID, f.Estado)
	}

	return s, nil
}

// loteYTareas embebe SÓLO el lote que toca.
//
// Es el único caso del sobre donde no se puede servir una ruta: no hay forma de
// apuntar a "el lote 2 de este JSON". Y filtrar importa — darle los cuatro
// lotes al implementador es gastarle contexto y tentarlo con trabajo que
// todavía no le toca.
func loteYTareas(raiz, carpeta string, f *estado.Feature, partes []Parte) (int, []Parte) {
	l, hayLote := f.LoteActual()
	if !hayLote {
		return 0, append(partes, Parte{
			Titulo: "Tus tareas",
			Falta:  "todos los lotes están commiteados: no queda nada por implementar",
		})
	}

	p, err := tareas.Leer(raiz, carpeta)
	if err != nil {
		return l.Lote, append(partes, Parte{
			Titulo: fmt.Sprintf("Tus tareas (lote %d)", l.Lote),
			Falta:  err.Error(),
		})
	}

	var b strings.Builder
	for _, t := range p.DelLote(l.Lote) {
		fmt.Fprintf(&b, "%s  %s\n", t.ID, t.Descripcion)
		if len(t.Satisface) > 0 {
			fmt.Fprintf(&b, "    satisface: %s\n", strings.Join(t.Satisface, " · "))
		}
		for _, test := range t.Tests {
			fmt.Fprintf(&b, "    test: %s\n", test)
		}
	}

	titulo := fmt.Sprintf("Tus tareas (lote %d de %d)", l.Lote, len(p.Lotes()))
	return l.Lote, append(partes, Parte{Titulo: titulo, Contenido: strings.TrimRight(b.String(), "\n")})
}

// diff sirve el resumen, no el diff completo.
//
// Un diff entero de una feature grande es enorme, y meterlo adentro del
// contexto de un subagente es justo lo que el diseño quiere evitar. El resumen
// le dice QUÉ MIRAR; los archivos los abre él.
func diff(raiz, base, titulo string) Parte {
	if !git.EsRepo(raiz) {
		return Parte{Titulo: titulo, Falta: "el proyecto no está bajo git"}
	}
	resumen, err := git.DiffResumen(raiz, base)
	if err != nil {
		return Parte{Titulo: titulo, Falta: err.Error()}
	}
	if resumen == "" {
		return Parte{Titulo: titulo, Falta: "no hay cambios desde " + base}
	}
	return Parte{Titulo: titulo + " (desde " + base + ")", Contenido: resumen}
}

// mutantes es la corrida de la herramienta del ㉒, y TODAVÍA NO ESTÁ.
//
// Va acá y no en un comando aparte: `sf mutation` no existe (H13). El modelo
// necesita los sobrevivientes como INSUMO —si corre después, inventó a ciegas—
// y sf necesita el score igual, porque 71% → 88% es un número que compara entre
// vueltas. Si se lo pidiera al modelo le estaría creyendo al que trabajó.
//
// Falta parsear el frontmatter de la constitución para saber qué herramienta
// declaró el proyecto (`mutacion: gremlins`). Es del paso 4.
func mutantes() Parte {
	return Parte{
		Titulo: "Los mutantes que sobrevivieron",
		Falta:  "todavía no: falta leer `mutacion:` de la constitución (paso 4)",
	}
}

// aprendizajes son los journals de las features ya cerradas.
//
// Su lector es el mismo que el de la doc: el ⑫ de una feature futura. Y lo que
// NO llevan es progreso — eso vive en estado.json. Memoria sí, progreso no.
func aprendizajes(raiz string) Parte {
	m, _ := filepath.Glob(filepath.Join(raiz, docs.Journals()))
	if len(m) == 0 {
		return Parte{
			Titulo: "Aprendizajes de features anteriores",
			Falta:  "todavía no hay ninguna feature cerrada",
		}
	}
	return Parte{Titulo: "Aprendizajes de features anteriores", Rutas: relativas(raiz, m)}
}

// ────────────────────────────────────────────────────────────────────────────
// Render
// ────────────────────────────────────────────────────────────────────────────

// Texto arma el sobre para que lo lea un modelo.
//
// Si `completo` es true, embebe el contenido de cada archivo en vez de listar
// la ruta. Eso es para el caso de H1b: cuando el que trabaja es un modelo por
// consola sin shell propia, no puede abrir archivos — así que el orquestador
// corre `sf context --completo` y se lo pega en el prompt.
//
//	sf context y sf done los corre el que trabaja. Si el que trabaja no tiene
//	manos, el orquestador se las presta.
func (s *Sobre) Texto(raiz string, completo bool) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# Sobre — %s", s.Estado)
	if s.Feature != "" {
		fmt.Fprintf(&b, " · %s", s.Feature)
	}
	if s.Lote > 0 {
		fmt.Fprintf(&b, " · lote %d", s.Lote)
	}
	b.WriteString("\n")

	for _, p := range s.Partes {
		fmt.Fprintf(&b, "\n## %s\n", p.Titulo)

		if p.Falta != "" {
			fmt.Fprintf(&b, "⚠ %s\n", p.Falta)
			continue
		}
		if p.Contenido != "" {
			fmt.Fprintf(&b, "%s\n", p.Contenido)
		}
		for _, r := range p.Rutas {
			if !completo {
				fmt.Fprintf(&b, "%s\n", r)
				continue
			}
			contenido, err := os.ReadFile(filepath.Join(raiz, r))
			if err != nil {
				fmt.Fprintf(&b, "⚠ %s no se pudo leer: %v\n", r, err)
				continue
			}
			fmt.Fprintf(&b, "\n### %s\n```\n%s\n```\n", r, strings.TrimRight(string(contenido), "\n"))
		}
	}
	return b.String()
}

// ────────────────────────────────────────────────────────────────────────────
// Ayudantes
// ────────────────────────────────────────────────────────────────────────────

func rutasDeHistorias(ids []string) []string {
	r := make([]string, 0, len(ids))
	for _, id := range ids {
		r = append(r, docs.Historia(id))
	}
	return r
}

// historias lista todos los us-# del backlog. Sólo lo usa el ⑩, que es el único
// que los mira todos juntos.
func historias(raiz string) []string {
	m, _ := filepath.Glob(filepath.Join(raiz, docs.Backlog, "us-*.md"))
	return relativas(raiz, m)
}

// relativas convierte rutas absolutas a relativas a la raíz.
//
// El sobre siempre muestra rutas relativas: el que las lee corre en el mismo
// proyecto, y una ruta absoluta con el home de Javier adentro es ruido que
// además no sirve si el sobre viaja a otra máquina.
func relativas(raiz string, abs []string) []string {
	r := make([]string, 0, len(abs))
	for _, a := range abs {
		if rel, err := filepath.Rel(raiz, a); err == nil {
			r = append(r, rel)
		}
	}
	return r
}
