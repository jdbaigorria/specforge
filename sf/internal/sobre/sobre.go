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
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/constitucion"
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/historia"
	"github.com/jdbaigorria/specforge/sf/internal/revision"
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

// rechazo es el motivo del último `sf reject`, si lo hay.
//
// Va PRIMERO en el sobre y no al final, porque es lo único que el que rehace el
// trabajo tiene que leer antes que nada: sin eso arranca de cero y vuelve a
// proponer lo mismo que ya se rechazó.
func rechazo(motivo string) (Parte, bool) {
	if motivo == "" {
		return Parte{}, false
	}
	return Parte{
		Titulo:    "⚠ Esto se rechazó. El motivo",
		Contenido: motivo,
	}, true
}

// Armar arma el sobre del estado actual.
//
// Recibe todo ya leído —estado y roadmap— en vez de leerlo: así los tests no
// necesitan escribir un estado.json para probar cada sobre, y este paquete no
// duplica el manejo de "no hay archivo" que ya está resuelto en los otros.
func Armar(raiz string, e *estado.Estado, r *roadmap.Roadmap, g *global.Config) (*Sobre, error) {
	// Los cinco de producto no dependen de ninguna feature, así que se
	// resuelven mirando sólo los sellos. El orden es el mismo que el de
	// maquina.Siguiente, y por la misma razón: es el orden del flujo.
	s, err := armar(raiz, e, r, g)
	if err != nil {
		return nil, err
	}
	// El vocabulario va antes del material de trabajo: las palabras se leen
	// primero. Después del rechazo, que sigue siendo lo primero de todo.
	if p, hay := vocabulario(raiz); hay {
		s.Partes = append([]Parte{p}, s.Partes...)
	}
	// El motivo del rechazo se antepone a lo que sea que traiga el sobre: da
	// igual el estado, si algo se rechazó eso va arriba de todo.
	if p, hay := rechazo(motivoDeRechazo(e)); hay {
		s.Partes = append([]Parte{p}, s.Partes...)
	}
	return s, nil
}

// vocabulario sirve el glosario del negocio, si existe.
//
// ────────────────────────────────────────────────────────────────────────────
// VA EN TODOS LOS SOBRES, Y ES EL ÚNICO QUE PUEDE FALTAR SIN SER UNA FALTA
// ────────────────────────────────────────────────────────────────────────────
//
// Nace en el ①–⑤, cuando una palabra se tambalea, y de ahí en adelante lo lee
// todo el mundo: el que escribe el PRD, el que corta las historias, el que
// implementa. Ése es el punto — el vocabulario se construye UNA vez y los demás
// estados lo heredan en vez de reinventar las palabras.
//
// Se sirve acá y no caso por caso a propósito. Cada rama de deProducto y
// deFeature enumera sus rutas a mano, así que agregarlo ahí serían once lugares
// donde olvidarse, y olvidarse no daría error: daría un estado que usa otra
// palabra para la misma cosa, tres estados después.
//
// Y es PEREZOSO: no existe hasta que haga falta. Su ausencia no es una falta,
// así que NO lleva `Falta` — a diferencia del diff de git, que cuando no está
// hay que explicarlo. Acá no hay nada que explicar: no hizo falta.
func vocabulario(raiz string) (Parte, bool) {
	if _, err := os.Stat(filepath.Join(raiz, docs.Vocabulario)); err != nil {
		return Parte{}, false
	}
	return Parte{
		Titulo: "Las palabras de este proyecto",
		Rutas:  []string{docs.Vocabulario},
	}, true
}

// motivoDeRechazo busca el rechazo que corresponde al estado actual.
func motivoDeRechazo(e *estado.Estado) string {
	if f, hay := e.Actual(); hay && f.Rechazo != "" {
		return f.Rechazo
	}
	return e.Producto.Rechazo
}

func armar(raiz string, e *estado.Estado, r *roadmap.Roadmap, g *global.Config) (*Sobre, error) {
	if s, hay := deProducto(raiz, e); hay {
		return s, nil
	}
	return deFeature(raiz, e, r, g)
}

// ────────────────────────────────────────────────────────────────────────────
// Producto
// ────────────────────────────────────────────────────────────────────────────

func deProducto(raiz string, e *estado.Estado) (*Sobre, bool) {
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
		partes := []Parte{
			{Titulo: "Qué hay que partir en historias", Rutas: []string{docs.PRD}},
			{Titulo: "Las reglas del proyecto", Rutas: []string{docs.Constitucion}},
		}
		// Al ⑨ se entra dos veces por motivos distintos, y el sobre tiene que
		// notarlo: la primera es partir el PRD, y las otras son completar lo
		// que entró por `sf new`. Sin esto, el que va a pinponear un `us-7`
		// recibe el PRD entero y ninguna pista de cuál es.
		if p, hay := aPinponear(raiz); hay {
			partes = append(partes, p)
		}
		return &Sobre{Estado: "backlog", Partes: partes}, true
	}
	return nil, false
}

// aPinponear arma la parte del ⑨ cuando lo que falta es completar, no partir.
//
// ────────────────────────────────────────────────────────────────────────────
// AL ⑨ SE ENTRA DOS VECES, Y POR MOTIVOS DISTINTOS
// ────────────────────────────────────────────────────────────────────────────
//
//	la primera   partir el PRD en historias        → el PRD alcanza
//	las otras    completar lo que metió `sf new`   → hace falta CUÁL
//
// Sin esto, el que va a pinponear un `us-7` recibe el PRD entero y ninguna
// pista de cuál es la historia nueva. Y en un producto con veinte historias,
// "partí el PRD" es la instrucción equivocada: lo que falta es una.
//
// Va vacío cuando faltan TODAS —ahí sí es la primera vuelta y el PRD es todo el
// insumo— para no repetir en el sobre lo que ya dice el título de arriba.
func aPinponear(raiz string) (Parte, bool) {
	faltan := historia.SinPinponear(raiz)
	if len(faltan) == 0 || len(faltan) == len(historia.Ids(raiz)) {
		return Parte{}, false
	}

	rutas := rutasDeHistorias(faltan)

	// Y si alguna es un bug, la historia que rompió va con ella: `relacionado_a`
	// apunta al us-# original, y sin él el que pinponea el bug no sabe qué
	// comportamiento se esperaba. Es el mismo dato que después usa el sobre del
	// implementador para traer la spec archivada.
	for _, id := range faltan {
		h, err := historia.Leer(raiz, id)
		if err != nil || h.RelacionadoA == "" {
			continue
		}
		if r := docs.Historia(h.RelacionadoA); !slices.Contains(rutas, r) {
			rutas = append(rutas, r)
		}
	}

	return Parte{Titulo: "Lo que hay que completar", Rutas: rutas}, true
}

// ────────────────────────────────────────────────────────────────────────────
// Feature
// ────────────────────────────────────────────────────────────────────────────

func deFeature(raiz string, e *estado.Estado, r *roadmap.Roadmap, g *global.Config) (*Sobre, error) {
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

	// El estado EFECTIVO, igual que `sf next` y `sf done`. Sin esto, el sobre
	// de un bug era el del ⑫ —"qué hay que resolver"— mientras `sf next` ya
	// había lanzado a sf-build a implementar: el subagente recibía la historia
	// en vez de la spec, las tareas y los criterios.
	//
	// Y se perdía loQueRompio, que cuelga de la rama `implementar`: la spec
	// archivada de la feature que el bug rompió es justo lo que evita que el
	// que lo arregla arranque de cero sin saber qué se había decidido.
	est := estado.Efectivo(f.Estado, historia.SonTodasBugs(raiz, fr.Historias))
	s := &Sobre{Estado: est, Feature: fr.ID}

	switch est {
	case estado.Planificacion:
		s.Partes = []Parte{
			{Titulo: "El manual del proyecto", Rutas: []string{docs.Constitucion}},
			{Titulo: "Qué hay que resolver", Rutas: rutasDeHistorias(fr.Historias)},
			aprendizajes(raiz),
			// El ⑯ decide con qué se implementa cada lote, y hasta acá lo hacía
			// sin ver el catálogo. Ver menu.go.
			menu(g),
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
		// Y si esto es un bug, la spec de lo que rompió. Ver loQueRompio.
		if p, hay := loQueRompio(raiz, fr, r); hay {
			s.Partes = append(s.Partes, p)
		}

	case estado.Revision:
		s.Partes = []Parte{
			{Titulo: "El manual del proyecto", Rutas: []string{docs.Constitucion}},
			{Titulo: "Qué se había planeado", Rutas: []string{filepath.Join(carpeta, docs.Spec)}},
			{Titulo: "Los criterios, y hay que opinar sobre TODOS", Rutas: rutasDeHistorias(fr.Historias)},
			diff(raiz, f.BaseCommit, "El código que quedó"),
			mutantes(raiz),
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
	// Los dos casos que LoteActual() mete en la misma bolsa, otra vez — y acá
	// mentir cuesta caro: decirle "no queda nada por implementar" a un
	// subagente que acaba de nacer para implementar es mandarlo a no hacer nada.
	if f.SinSembrar() {
		return 0, append(partes, Parte{
			Titulo: "Tus tareas",
			Falta:  "todavía no empezaste ningún lote: corré `sf lote start`",
		})
	}

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

	// Un lote que NO está en el plan es el que abrió la vuelta del ㉑ para
	// arreglar un hallazgo. Servirle las tareas del plan sería servirle una
	// lista vacía —y un "lote 2 de 1" que no significa nada—, así que en su
	// lugar va LO QUE HAY QUE ARREGLAR.
	//
	// Sin esto el arreglo de A5 quedaba a mitad de camino: el lote existía, el
	// rojo se podía confirmar, y el subagente fresco no tenía cómo saber qué
	// venía a arreglar. Un lote sin sobre es un subagente improvisando, que es
	// exactamente lo que la máquina existe para evitar.
	if !slices.Contains(p.Lotes(), l.Lote) {
		return l.Lote, append(partes, hallazgosAbiertos(raiz, carpeta, l.Lote))
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

// hallazgosAbiertos arma la parte del lote de corrección.
//
// Va el detalle ENTERO y no una ruta, al revés que casi todo el resto del
// sobre: son dos o tres líneas que ya están escritas y que el que las lee
// necesita sí o sí. Mandarlo a abrir el revision.json para leer un renglón
// sería gastar una tool-call en algo que entra en el prompt.
//
// Y el criterio va adelante del detalle porque es lo que ata el arreglo a lo
// que hay que volver a probar: el ㉑ de la vuelta siguiente opina sobre ese
// mismo id.
func hallazgosAbiertos(raiz, carpeta string, lote int) Parte {
	titulo := fmt.Sprintf("Qué hay que arreglar (lote %d, la vuelta del ㉑)", lote)

	rev, err := revision.Leer(filepath.Join(raiz, carpeta, docs.Revision))
	if err != nil {
		return Parte{Titulo: titulo, Falta: err.Error()}
	}

	abiertos := rev.Abiertos()
	if len(abiertos) == 0 {
		// No debería pasar —el lote lo abrió cerrarRevision justamente porque
		// había hallazgos—, pero un sobre incompleto es mejor que uno que miente.
		return Parte{Titulo: titulo, Falta: "no quedan hallazgos abiertos en revision.json"}
	}

	var b strings.Builder
	for _, h := range abiertos {
		fmt.Fprintf(&b, "%s", h.ID)
		if h.Criterio != "" {
			fmt.Fprintf(&b, "  (%s)", h.Criterio)
		}
		fmt.Fprintf(&b, "\n    %s\n", h.Detalle)
	}
	return Parte{Titulo: titulo, Contenido: strings.TrimRight(b.String(), "\n")}
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

// mutantes es la corrida de la herramienta del ㉒.
//
// Va en el sobre y no en un comando aparte: `sf mutation` no existe (H13). El
// modelo necesita los sobrevivientes como INSUMO —si la corrida fuera después,
// habría opinado a ciegas— y sf necesita el score igual, porque 71% → 88% es un
// número que compara entre vueltas. Pedírselo al modelo sería creerle al que
// trabajó.
//
// ────────────────────────────────────────────────────────────────────────────
// VACÍO NO ES UN ERROR
// ────────────────────────────────────────────────────────────────────────────
//
// `mutacion:` puede estar vacío y está bien: si el stack no tiene una
// herramienta buena, el ㉒ lo hace el modelo leyendo el código. La herramienta
// es una MEJORA, no un requisito — igual que los subagentes.
//
// Y por eso esto no frena nunca: pase lo que pase, la parte se sirve diciendo
// qué ocurrió. Un sobre al que le falta el dato tiene que DECIRLO; si la parte
// simplemente no apareciera, el revisor creería que no hacía falta.
func mutantes(raiz string) Parte {
	const titulo = "Los mutantes que sobrevivieron"

	c, err := constitucion.Leer(raiz)
	if err != nil {
		return Parte{Titulo: titulo, Falta: "no pude leer la constitución: " + err.Error()}
	}
	if strings.TrimSpace(c.Mutacion) == "" {
		return Parte{Titulo: titulo, Falta: "el proyecto no declaró `mutacion:` — el ㉒ lo hacés leyendo el código"}
	}

	// Se corre con la shell por lo mismo que `test_cmd`: las líneas reales
	// llevan flags con comillas y a veces un pipe, y partir por espacios las
	// rompe en silencio.
	cmd := exec.Command("sh", "-c", c.Mutacion)
	cmd.Dir = raiz
	salida, err := cmd.CombinedOutput()

	// El exit code NO se mira: una corrida de mutantes sale distinto de cero
	// cuando sobrevive alguno, que es justamente el caso interesante. Lo que
	// importa es la salida, y de ahí el score lo lee el modelo — sf no parsea
	// una herramienta por lenguaje (mismo argumento que suite.Correr).
	texto := strings.TrimRight(string(salida), "\n")
	if texto == "" {
		if err != nil {
			return Parte{Titulo: titulo, Falta: c.Mutacion + " no se pudo correr: " + err.Error()}
		}
		return Parte{Titulo: titulo, Falta: c.Mutacion + " no imprimió nada"}
	}
	return Parte{Titulo: titulo + " (" + c.Mutacion + ")", Contenido: texto}
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

// loQueRompio sirve la spec ARCHIVADA de la feature que este bug rompió.
//
// ────────────────────────────────────────────────────────────────────────────
// EL CAMPO ESTABA DECLARADO Y NO LO LEÍA NADIE
// ────────────────────────────────────────────────────────────────────────────
//
// `relacionado_a` vivía en historia.go y en el esqueleto del us-# desde el
// principio, con cero consumidores. Es el mismo patrón que `tipo: bug` antes de
// H20: un dato que ya existía, esperando el suyo.
//
// El agujero que tapa es concreto y es el dolor #5:
//
//	un bug NO desarchiva nada — entra como us-# nuevo, en una feature nueva —
//	así que el subagente que lo arregla NO VE la spec de lo que rompió, y
//	"el implementador que no encuentra lo que la spec dice, IMPROVISA".
//
// ────────────────────────────────────────────────────────────────────────────
// Y NO ROMPE NINGUNA REGLA: SERVIR UN ARCHIVO NO ES PENSAR NI LANZAR A NADIE
// ────────────────────────────────────────────────────────────────────────────
//
// La cadena son dos saltos, y los dos usan datos que ya existen:
//
//	us-7.relacionado_a = "us-3"      →  qué historia rompió
//	roadmap: ¿qué feature tenía us-3?→  f-1
//	.docs/archivado/f-1-slug/spec-design.md
//
// El roadmap conserva las features cerradas —archivar mueve la carpeta, no toca
// el roadmap— así que el segundo salto es una búsqueda en memoria.
func loQueRompio(raiz string, fr roadmap.Feature, r *roadmap.Roadmap) (Parte, bool) {
	// Un mapa us-# → feature, armado una vez. Es más barato que recorrer el
	// roadmap entero por cada historia, y sobre todo se lee mejor.
	deQuienEs := map[string]roadmap.Feature{}
	for _, f := range r.Features {
		for _, id := range f.Historias {
			deQuienEs[id] = f
		}
	}

	var rutas []string
	visto := map[string]bool{}

	for _, id := range fr.Historias {
		h, err := historia.Leer(raiz, id)
		// Si la historia no se puede leer, el que se queja es la compuerta del
		// ⑨, no el sobre. Acá se saltea: un sobre incompleto es mejor que un
		// sobre que no sale.
		if err != nil || h.RelacionadoA == "" {
			continue
		}

		original, hay := deQuienEs[h.RelacionadoA]
		if !hay || original.ID == fr.ID {
			// No está en el roadmap, o el bug cayó en la MISMA feature que la
			// historia original. En el segundo caso la spec ya está en el sobre
			// —es la de esta feature— y repetirla sería gastar contexto.
			continue
		}

		// La carpeta archivada conserva el mismo nombre que tenía viva: lo que
		// cambia es el padre. Por eso alcanza con filepath.Base.
		ruta := filepath.Join(docs.Archivado, filepath.Base(original.Carpeta()), docs.Spec)
		if visto[ruta] {
			continue
		}
		// Se comprueba que exista antes de ofrecerla: la feature original puede
		// no estar archivada todavía (un bug sobre algo que se cerró en esta
		// misma corrida). Ofrecer una ruta que no está sería mentir por
		// omisión al revés.
		if _, err := os.Stat(filepath.Join(raiz, ruta)); err != nil {
			continue
		}
		visto[ruta] = true
		rutas = append(rutas, ruta)
	}

	if len(rutas) == 0 {
		return Parte{}, false
	}
	return Parte{Titulo: "Cómo se había resuelto lo que este bug rompió", Rutas: rutas}, true
}

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
