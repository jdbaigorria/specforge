// Package maquina contesta la única pregunta que importa: ¿qué sigue?
//
// ────────────────────────────────────────────────────────────────────────────
// LO QUE ESTE PAQUETE **NO** HACE, Y ES LA MITAD DEL DISEÑO
// ────────────────────────────────────────────────────────────────────────────
//
// No lanza a nadie. No escribe el estado. No corre tests. Lee tres cosas —el
// estado.json, el roadmap.json y qué archivos existen— y devuelve una
// instrucción. Se muere en microsegundos.
//
//	"sf nunca lanza a nadie. El que lanza es el orquestador. Si sf spawneara,
//	 manejaría contexto y tool-calling, y SERÍA UN HARNESS."   (session.md §6)
//
// Y no piensa en ningún lado: todas las decisiones de acá son mirar un campo,
// contar una lista o preguntar si un archivo existe.
//
// ────────────────────────────────────────────────────────────────────────────
// EL TRUCO QUE EVITA UN MONTÓN DE CAMPOS: LOS CHECKPOINTS SE DEDUCEN
// ────────────────────────────────────────────────────────────────────────────
//
// Un estado no es un subagente (maquina-estados.md §1). `planificacion` produce
// tres archivos y un subagente puede morirse en el medio. Para saber por dónde
// retomar NO hay un campo de progreso: se mira qué archivos hay.
//
//	¿hay decision.md?      no → arrancá de cero
//	¿hay spec-design.md?   no → retomá desde la spec
//	¿hay tareas.json?      no → retomá desde las tareas
//
// Es la regla 1.5 de artefactos.md aplicada al progreso: lo que se puede
// deducir, no se guarda.
package maquina

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/compuerta"
	"github.com/jdbaigorria/specforge/sf/internal/constitucion"
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/historia"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
	"github.com/jdbaigorria/specforge/sf/internal/tareas"
)

// Tipo es cuál de las cuatro paradas —o ninguna— corresponde ahora.
//
// Eran tres en artefactos.md §2 y apareció una cuarta al cerrar la máquina:
// ME TRABÉ, que estaba escondida en el ⑳ ("si falla varias veces, entro yo").
// Las otras tres salen del gusto de Javier y son configurables; esa no, porque
// es de seguridad: sin ella, el bucle "sf da rojo → el orquestador relanza"
// no termina nunca.
type Tipo int

const (
	// Trabajar es el caso normal: hay algo que hacer y sf dice quién y con qué.
	Trabajar Tipo = iota

	// Para es la 🛑 de ⑥, ⑧ y ⑰: no avanza sin Javier.
	Para

	// Barata es la ⏸ de ⑨ y ㉓: te muestra qué salió, seguís con un enter.
	Barata

	// MeTrabe es el freno del bucle. No es configurable.
	MeTrabe

	// Fin es que no queda nada: todas las features del roadmap están cerradas.
	Fin
)

// TopeIntentos es cuántos `sf done` en ✗ seguidos disparan ME TRABÉ.
//
// Tres sale del ejemplo del diseño ("el lote 2 falló 3 veces con deepseek").
// Es el único número mágico del paquete y está acá arriba, solo, para que se
// vea. Cuando haya constitución de verdad, sale de ahí.
const TopeIntentos = 3

// Instruccion es lo que `sf next` le contesta al orquestador.
//
// Es un struct de datos y no un string armado, aunque hoy el único consumidor
// lo imprima: el día que haga falta `sf next --json` —para un harness que
// parsee en vez de leer— la decisión ya está separada de cómo se muestra.
type Instruccion struct {
	Tipo Tipo

	// Estado es uno de los 9 (o "" en una parada de producto ya cerrada).
	Estado string

	Feature string // "" en los estados de producto

	// Lote y DeLotes arman el "lote 2 de 4" de `implementar`. Cero = no aplica.
	Lote, DeLotes int

	// Skill, Perfil, Modelo, Agente y Via son la mitad del hallazgo H1: si sf no
	// los devuelve, la tabla estado→skill→modelo vive en CLAUDE.md y hay que
	// mantener una copia por harness. Devolviéndolos, CLAUDE.md queda en cuatro
	// líneas y AGENTS.md es el mismo texto.
	//
	// Perfil y Modelo son dos cosas distintas y viajan juntas a propósito:
	//
	//	Perfil   qué PIDE el paso          razonar          vocabulario de sf
	//	Modelo   qué se le da en ESTA máquina  laguna/s2.1  vocabulario del harness
	//
	// Que viajen juntas es lo que hace que un `sf next` sirva de log: dentro de
	// tres meses se puede ver qué pedía el paso y qué se le dio. Con una sola
	// palabra —como era antes— las dos preguntas tenían la misma respuesta y no
	// se podían distinguir.
	Skill  string
	Perfil string
	Modelo string

	// Esfuerzo es cuánto tiene que pensar, si el alias lo declara.
	//
	// Vacío es "el que traiga el modelo por default", que NO es lo mismo que
	// "bajo": sf no elige un esfuerzo igual que no elige un modelo.
	Esfuerzo string

	Via string

	// Agente es A QUIÉN invocar para conseguir ese modelo, y viene vacío en
	// Claude Code.
	//
	// Es la otra mitad de H1b, la que sólo sf puede contestar. En Claude Code la
	// herramienta de subagente acepta el modelo como parámetro, así que alcanza
	// con `Modelo`. En opencode y Command Code NO —medido: el `model:` sale del
	// archivo del agente y no se puede pisar al invocar—, así que sf genera un
	// portamodelo por alias y acá devuelve su nombre.
	Agente string

	// Comando es con qué se sale por consola, y sólo viene con `via: consola`.
	//
	//	modelo:  deepseek
	//	via:     consola
	//	comando: deepseek exec
	//
	// Sale del mapa de modelos (~/.specforge/), que es el único lugar que sabe
	// cómo se invoca cada uno en ESTA máquina.
	Comando string

	// Mensaje es qué pasa, en una línea, para el humano o el LLM que lee.
	Mensaje string

	// Sugerido son los comandos que corresponden ahora. En una parada, son las
	// salidas que tiene Javier.
	Sugerido []string

	// Avisos son cosas que hay que saber y que NO frenan.
	//
	// La distinción es de diseño: sf frena sobre hechos y avisa sobre todo lo
	// demás, porque la última palabra es de Javier y una herramienta que frena
	// sola rompe esa regla.
	Avisos []string
}

// ────────────────────────────────────────────────────────────────────────────
// El mapa estado → skill
// ────────────────────────────────────────────────────────────────────────────
//
// Vive acá y no en CLAUDE.md, y esa es la decisión H1 hecha código. La pregunta
// "¿qué skill corresponde al estado planificacion?" tiene UNA SOLA RESPUESTA
// CORRECTA, así que la contesta sf (R1).
//
// Los nombres salen de la REGLA DE PREFIJOS, que dejó de ser decorativa
// (que-sobrevive.md §4) y que este mapa violaba en seis de nueve entradas hasta
// que la ronda de skills.md lo corrigió:
//
//	sfp-*   los 5 estados de PRODUCTO    corren una vez
//	sf-*    los 4 estados de FEATURE     corren una vez por feature
//	sfx-*   utilitarios                  fuera de la máquina — sf NO los conoce
//
// La tercera línea es la que muerde, y es el motivo por el que acá no aparece
// ningún `sfx-`: un utilitario es standalone —no pide sobre, no avisa que
// terminó—, así que devolver su nombre sería lanzar a alguien que nunca va a
// llamar a `sf done`, y EL ESTADO NO SE MOVERÍA NUNCA.
//
// El método sigue viviendo en los utilitarios: el skill de estado es delgado y
// LOS COMPONE, igual que sfp-scout ya componía sfx-think (②) y sfx-grill-me (⑤).
//
//	sf-plan    → sfx-think (⑫)
//	sf-build   → sfx-tdd (⑲) · sfx-github (el commit)
//	sf-cierre  → sfx-documenter · sfx-journal · sfx-github
//
// Los nueve existen y están adaptados: arrancan con `sf context` y terminan con
// `sf done` (skills.md §8).
var skills = map[string]string{
	"brief":              "sfp-scout",
	"prd":                "sfp-po",
	"constitucion":       "sfp-constitucion",
	"backlog":            "sfp-backlog",
	"roadmap":            "sfp-roadmap",
	estado.Planificacion: "sf-plan",
	estado.Implementar:   "sf-build",
	estado.Revision:      "sf-check",
	estado.Cierre:        "sf-cierre",
}

// modeloPorEstado es lo que el paso PIDE por default, y es el piso de tres.
//
// La cadena de precedencia, de la que más manda a la que menos:
//
//  1. sf model <nombre>   la decisión de Javier en runtime (estado.json)
//  2. tareas.json         el ⑯: "ESTA feature necesita uno más grande"
//  3. modeloPorEstado     el criterio del diseño, por estado
//  4. modeloPorDefecto
//
// El orden no es arbitrario: cada nivel sabe MENOS que el de arriba. El default
// no sabe nada de la feature; el ⑯ la planificó pero no vio fallar nada; Javier
// está mirando el bucle patinar cuando escribe `sf model`.
//
// Y la otra mitad de H1b —CÓMO se lanza el modelo que sale de acá— vive en
// `~/.specforge/` y la resuelve `via()`. Este mapa dice QUÉ hace falta; el otro
// dice cómo se invoca en esta máquina. Ninguno de los dos sabe lo del otro, y
// por eso `sf next` es el único que puede contestar la pregunta entera.
// SkillsDeEstado son los nueve, ordenados como los recorre la máquina.
//
// Existe para que `sf doctor` pueda comprobar que están instalados sin escribir
// la lista una segunda vez: el mapa de arriba es la única copia, y una lista
// suelta en el doctor se desincronizaría el día que un estado cambie de skill.
func SkillsDeEstado() []string {
	orden := []string{
		"brief", "prd", "constitucion", "backlog", "roadmap",
		estado.Planificacion, estado.Implementar, estado.Revision, estado.Cierre,
	}
	var s []string
	for _, e := range orden {
		s = append(s, skills[e])
	}
	return s
}

// perfilPorEstado es qué PIDE cada paso, y es el piso de la cadena.
//
// Dice un ROL y no un modelo, y ésa es la corrección entera de H2: "opus" no es
// una necesidad, es una respuesta — y una respuesta que sólo vale en un harness.
// Los dos que están acá son los que JUZGAN: el ⑫ compara tres opciones y elige,
// el ㉑ tiene que encontrar lo que no está. El resto escribe código contra un
// plan que ya existe, y para eso está el default.
//
// `mecanico` no aparece a propósito: ningún estado lo pide. Lo nombran los sfx-*,
// que están fuera de los nueve, y por eso su entrada en el catálogo es opcional
// y no frena.
var perfilPorEstado = map[string]string{
	estado.Planificacion: global.Razonar,
	estado.Revision:      global.Razonar,
}

const perfilPorDefecto = global.Construir

// Siguiente decide qué sigue. Es toda la lógica de `sf next`.
//
// `raiz` es la raíz del proyecto: hace falta porque varias decisiones se toman
// mirando qué archivos existen, que es el truco de los checkpoints.
func Siguiente(raiz string, e *estado.Estado, r *roadmap.Roadmap, g *global.Config) Instruccion {
	// El orden importa y es el del flujo: los cinco de producto primero, y
	// recién cuando están los cinco empieza el ciclo de feature. No hay forma
	// de planificar sin constitución.
	if i, hay := siguienteDeProducto(raiz, e, r, g); hay {
		return i
	}
	return siguienteDeFeature(raiz, e, r, g)
}

// ────────────────────────────────────────────────────────────────────────────
// Producto — los cinco que corren una vez
// ────────────────────────────────────────────────────────────────────────────

func siguienteDeProducto(raiz string, e *estado.Estado, r *roadmap.Roadmap, g *global.Config) (Instruccion, bool) {
	// ① brief. El sello es un veredicto, no un bool: vacío = no sellado.
	if e.Producto.BriefSellado == "" {
		// Acá está el patrón que se repite en los tres primeros estados, y es
		// la diferencia entre "falta hacerlo" y "está hecho, falta que lo
		// mires": lo separa la EXISTENCIA DEL ARCHIVO, no un campo.
		if !existe(raiz, docs.Brief) {
			return trabajar("brief", "", "el ⑥ lo sellás vos cuando esté", g), true
		}
		return Instruccion{
			Tipo:     Para,
			Estado:   "brief",
			Mensaje:  "🛑 PARÁ. El brief está escrito y lo sellás vos (el ⑥).",
			Sugerido: []string{"sf approve", "sf reject \"motivo\""},
		}, true
	}

	// Si el veredicto no fue "hacelo", la máquina no sigue. Es la única vez que
	// el flujo termina sin haber construido nada, y está bien que exista: el
	// valor del ⑥ es poder decir que no.
	if e.Producto.BriefSellado != "hacelo" {
		return Instruccion{
			Tipo:    Fin,
			Estado:  "brief",
			Mensaje: fmt.Sprintf("El brief se selló como %q. No hay nada que seguir.", e.Producto.BriefSellado),
		}, true
	}

	// ⑦ prd. El hash lo pone `sf done`, así que vacío = todavía no cerró.
	if e.Producto.PrdHash == "" {
		return trabajar("prd", "", "sin parada: del ⑦ se pasa derecho al ⑧", g), true
	}

	// ⑧ constitución.
	//
	// Acá el patrón de "¿existe el archivo?" no alcanza, y es el único de los
	// tres donde falla: `sf init` SIEMPRE crea la constitución —tiene que
	// crearla, ahí escribe el test_cmd que detectó del stack—, así que el
	// checkpoint por existencia daba siempre "está escrita" y el ⑧ nunca se
	// alcanzaba. Lo que se sellaba era la plantilla.
	//
	// El marcador que deja el andamio convierte la pregunta en "¿está
	// ESCRITA?", que sigue siendo comparar dos strings.
	if !e.Producto.ConstitucionSellada {
		if !existe(raiz, docs.Constitucion) || constitucion.SinEscribir(raiz) {
			return trabajar("constitucion", "", "el ⑧ lo sellás vos cuando esté", g), true
		}
		return Instruccion{
			Tipo:     Para,
			Estado:   "constitucion",
			Mensaje:  "🛑 PARÁ. La constitución está escrita y la sellás vos (el ⑧).",
			Sugerido: []string{"sf approve", "sf reject \"motivo\""},
		}, true
	}

	// ⑨ backlog. Acá el patrón cambia: la parada es ⏸ y no 🛑, y necesita el
	// campo BacklogVisto porque un enter no deja rastro (ver estado.go).
	//
	// Y el checkpoint pregunta QUÉ FALTA, no si hay algo. "¿Hay historias?"
	// funciona una sola vez —la primera—, y después `sf new` reabre esta ⏸ para
	// que se pinponee lo que entró: con el producto en marcha la respuesta es
	// siempre sí, así que la ⏸ salía en vez del trabajo y lo que quedaba para
	// aprobar era el esqueleto vacío. Es la misma forma que ya usa el ⑩ con
	// `huerfanas`.
	if !e.Producto.BacklogVisto {
		// Sin ninguna historia también hay trabajo, y no lo dice SinPinponear:
		// una lista vacía no tiene nada incompleto. Es el mismo error que el de
		// los lotes —"ninguno" no es "todos"— y por eso los dos casos se
		// preguntan por separado.
		ids := historia.Ids(raiz)
		faltan := historia.SinPinponear(raiz)
		if len(ids) == 0 || len(faltan) > 0 {
			return trabajar("backlog", "", mensajeDelNueve(ids, faltan), g), true
		}
		return Instruccion{
			Tipo:     Barata,
			Estado:   "backlog",
			Mensaje:  "⏸ Salieron las historias. Mirá si querés y seguí.",
			Sugerido: []string{"sf approve"},
		}, true
	}

	// ⑩ roadmap. Sin parada: agrupa y ordena, y de ahí arranca el ciclo.
	if !existe(raiz, roadmap.Archivo) {
		return trabajar("roadmap", "", "el ⑩ agrupa las historias en features", g), true
	}

	// Y el ⑩ vuelve a hacer falta cada vez que entra algo nuevo por `sf new`:
	// una historia que no está en ninguna feature no la implementa nadie, y sin
	// esto el ciclo seguiría como si nada.
	if sueltas := huerfanas(raiz, r); len(sueltas) > 0 {
		i := trabajar("roadmap", "", fmt.Sprintf(
			"hay %d historia(s) sin feature: %s", len(sueltas), strings.Join(sueltas, " · ")), g)
		return i, true
	}

	return Instruccion{}, false
}

// ────────────────────────────────────────────────────────────────────────────
// Feature — los cuatro que se repiten, una vuelta por feature
// ────────────────────────────────────────────────────────────────────────────

func siguienteDeFeature(raiz string, e *estado.Estado, r *roadmap.Roadmap, g *global.Config) Instruccion {
	// Sin roadmap no hay cola. No debería pasar (siguienteDeProducto ya lo
	// habría atajado), pero un nil acá sería un panic y no un mensaje.
	if r == nil {
		return Instruccion{Tipo: Fin, Mensaje: "No hay roadmap.json: falta el ⑩."}
	}

	f, hay := e.Actual()
	if !hay {
		return tomarLaProxima(raiz, e, r, g)
	}

	// ME TRABÉ va PRIMERO, antes de mirar en qué estado está: si el contador
	// llegó al tope, no importa qué falta — importa que hace rato que no
	// avanza. Ponerlo después dejaría que sf siga proponiendo trabajo mientras
	// el bucle patina.
	fr, _ := r.Buscar(e.FeatureActual)

	if f.IntentosFallidos >= TopeIntentos {
		// El modelo sale de la cadena y no de `f.Modelo` crudo: ese campo está
		// vacío hasta el primer `sf model`, y justo la primera vez que aparece
		// esta parada es la vez que nadie lo corrió todavía. El mensaje decía
		// "falló 3 veces seguidas con ." — que es la mitad del dato que hace
		// falta para contestarle.
		return Instruccion{
			Tipo:    MeTrabe,
			Estado:  f.Estado,
			Feature: e.FeatureActual,
			Mensaje: fmt.Sprintf("⚠ ME TRABÉ. %s falló %d veces seguidas con %s.\n"+
				"   ¿Subo el modelo, o entrás vos?",
				e.FeatureActual, f.IntentosFallidos,
				conQueSeTrabo(raiz, fr, f, g)),
			Sugerido: []string{"sf model <alias>", "sf dismiss <h-#> \"motivo\""},
		}
	}

	// El camino corto (maquina-estados.md §8): un bug no pasa por planificación
	// ni por revisión. No es un carril paralelo —eso sería una segunda máquina
	// que mantener— sino la MISMA máquina con estados salteados.
	//
	// Se calcula en una variable local y NO se escribe en f: `sf next` es
	// consulta pura y correrlo dos veces tiene que dar lo mismo. Los que mueven
	// el estado son `sf done` y `sf lote start`, y allá se aplica la MISMA
	// función — que exista una sola es el arreglo: tenerla copiada acá y en
	// `done` mientras faltaba en `lote start` dejaba a los comandos
	// contradiciéndose sobre la misma feature.
	actual := estado.Efectivo(f.Estado, historia.SonTodasBugs(raiz, fr.Historias))

	switch actual {
	case estado.Planificacion:
		return planificando(raiz, fr, f, g)

	case estado.Planificada:
		// El plan está aprobado y esperando. Es la puerta "otra feature" del
		// ⑰: aprobaste y te fuiste a planificar otra. Volver a ésta es una
		// decisión de Javier, así que sf no la retoma sola.
		return Instruccion{
			Tipo:     Para,
			Estado:   estado.Planificada,
			Feature:  e.FeatureActual,
			Mensaje:  fmt.Sprintf("El plan de %s está aprobado y esperando.", e.FeatureActual),
			Sugerido: []string{"sf take " + e.FeatureActual},
		}

	case estado.Implementar:
		return implementando(raiz, fr, f, g)

	case estado.Revision:
		return trabajarEnFeature(raiz, estado.Revision, fr, f,
			"el ㉑ juzga los criterios y el ㉒ mira los mutantes, en la misma pasada", g)

	case estado.Cierre:
		// Igual que en `planificacion`: la compuerta es barata —mirar dos
		// archivos— así que sf next la corre para distinguir "falta escribir la
		// doc" de "está lista, mirá y archivá".
		if res := compuerta.Cierre(raiz, fr); res.Pasa() {
			return Instruccion{
				Tipo:     Barata,
				Estado:   estado.Cierre,
				Feature:  e.FeatureActual,
				Mensaje:  fmt.Sprintf("⏸ %s lista para archivar. La doc y el journal están.", e.FeatureActual),
				Sugerido: []string{"sf approve"},
			}
		}
		return trabajarEnFeature(raiz, estado.Cierre, fr, f,
			"el ㉓ escribe la doc y el journal", g)

	case estado.Cerrada:
		// La feature actual quedó cerrada y nadie tomó otra: sigue la cola.
		return tomarLaProxima(raiz, e, r, g)
	}

	return Instruccion{
		Tipo:    Fin,
		Mensaje: fmt.Sprintf("%s tiene un estado que no conozco: %q", e.FeatureActual, f.Estado),
	}
}

// tomarLaProxima es el ⑪, y termina en un `sf take` y no en un trabajo.
//
// Podría tomarla sola: cuál es la próxima tiene una sola respuesta correcta (la
// primera no cerrada), así que R1 lo permitiría. NO lo hace, por dos razones:
//
//  1. `sf next` es CONSULTA PURA. No escribe el estado nunca. Es lo que lo hace
//     imposible de romper: correrlo dos veces da lo mismo.
//  2. El diseño le dio nombre propio a esa transición justamente porque Javier
//     puede querer otra: "sirve igual fuera del ⑰ — es como se elige la primera
//     feature de todas, y como se cambia de idea a mitad de la cola" (H4).
func tomarLaProxima(raiz string, e *estado.Estado, r *roadmap.Roadmap, g *global.Config) Instruccion {
	cerradas := map[string]bool{}
	for id, f := range e.Features {
		if f.Estado == estado.Cerrada {
			cerradas[id] = true
		}
	}

	prox, hay := r.Proxima(cerradas)
	if !hay {
		return Instruccion{
			Tipo:    Fin,
			Mensaje: "🎉 Todas las features del roadmap están cerradas.",
		}
	}

	// Acá todavía no se lanza a nadie —lo sugerido es `sf take`— pero el `via`
	// igual sale del mapa: es un anticipo de la próxima vuelta, y un anticipo
	// que dice "subagente" cuando en realidad va por consola desinforma.
	//
	// Y el modelo se anticipa con la misma cadena que se va a usar de verdad,
	// si la feature ya está en el estado: puede estar `planificada` —la puerta
	// "otra feature" del ⑰— y traer un `sf model` de una vuelta anterior.
	// Anticipar el default cuando el real es otro es la misma desinformación
	// que el `via`, un renglón más abajo.
	base := perfil(estado.Planificacion)
	pedido := base
	if f, hay := e.Features[prox.ID]; hay {
		pedido = pedidoDeFeature(raiz, prox, f, estado.Planificacion)
	}
	m, aviso, declarado := lanzar(estado.Planificacion, pedido, base, g)
	if !declarado {
		return sinDeclarar(estado.Planificacion, prox.ID, pedido, g)
	}

	return Instruccion{
		Tipo:     Trabajar,
		Estado:   estado.Planificacion,
		Feature:  prox.ID,
		Skill:    skills[estado.Planificacion],
		Perfil:   base,
		Modelo:   m.ID,
		Esfuerzo: m.Esfuerzo,
		Via:      m.Via,
		Comando:  m.Comando,
		Agente:   agenteDe(m, g),
		Avisos:   avisos(aviso),
		Mensaje:  fmt.Sprintf("Sigue %s — %s (%d historias).", prox.ID, prox.Nombre, len(prox.Historias)),
		Sugerido: []string{"sf take " + prox.ID},
	}
}

// planificando deduce por dónde retomar mirando qué archivos hay.
//
// Los tres se producen en ESTE orden y por el mismo subagente, en una sola
// conversación. El strawman quería partirlos en tres estados y Javier lo
// objetó: el dolor #5 es daño de código, no de planificación, y acá el contexto
// grande es un ACTIVO — el ⑬ aprovecha acordarse de las dos opciones que
// descartó (maquina-estados.md §1).
func planificando(raiz string, fr roadmap.Feature, f *estado.Feature, g *global.Config) Instruccion {
	id := fr.ID
	i := trabajarEnFeature(raiz, estado.Planificacion, fr, f, "", g)
	if i.Tipo != Trabajar {
		// El modelo que hace falta no está declarado para este harness. El
		// guard existía en `implementando` y faltaba acá, y el síntoma era
		// silencioso: la 🛑 salía con sus comandos sugeridos pero con el mensaje
		// del ⑫ pisado encima —"arrancá de cero: las 3 opciones"—, así que quien
		// la leía veía dos cosas que no tienen nada que ver.
		//
		// Lo encontró el guion de seccionar: abrir otro arnés en el mismo repo
		// es la primera vez que esta parada salta en `planificacion`.
		return i
	}

	// La carpeta sale del roadmap (id + slug). Si la feature no está en el
	// roadmap, fr viene en cero y Carpeta() daría una ruta rara: en ese caso no
	// se puede deducir nada y se arranca de cero, que es lo seguro.
	if fr.ID == "" {
		i.Mensaje = "arrancá de cero: las 3 opciones del ⑫"
		return i
	}
	carpeta := fr.Carpeta()

	switch {
	case !existe(raiz, filepath.Join(carpeta, docs.Decision)):
		i.Mensaje = "arrancá de cero: las 3 opciones del ⑫"
	case !existe(raiz, filepath.Join(carpeta, docs.Spec)):
		i.Mensaje = "retomá desde la spec: decision.md ya está"
	case !existe(raiz, filepath.Join(carpeta, tareas.Archivo)):
		i.Mensaje = "retomá desde las tareas: la spec ya está"
	default:
		// Los tres archivos existen. Acá `sf next` corre las MISMAS compuertas
		// que `sf done` para saber si esto ya está listo para el ⑰ o si todavía
		// le falta algo.
		//
		// Puede hacerlo porque las compuertas de `planificacion` son baratas:
		// contar archivos, contar opciones, comparar dos listas de criterios.
		// En `implementar` no se hace, y no por gusto — ahí la compuerta es
		// correr los tests, y un `sf next` que corre la suite cada vez sería
		// insoportable. Allá no hace falta igual: el rastro es el commit.
		if res := compuerta.Planificacion(raiz, fr); !res.Pasa() {
			i.Mensaje = "los tres archivos están, pero falta:\n   " +
				strings.Join(res.Fallas, "\n   ")
			i.Sugerido = []string{"sf context", "sf done"}
		} else {
			return Instruccion{
				Tipo:     Para,
				Estado:   estado.Planificacion,
				Feature:  id,
				Mensaje:  fmt.Sprintf("🛑 PARÁ. El plan de %s lo revisás vos (el ⑰).", id),
				Sugerido: []string{"sf approve", "sf reject \"motivo\"", "sf take <otra>"},
			}
		}
	}
	return i
}

// implementando decide entre "exigí el rojo" e "implementá".
//
// El lote actual es el primero sin commit y no hay campo que lo diga (R6). Y
// dentro del lote, lo que separa las dos mitades es `rojo`: sin él no se puede
// implementar, porque sf todavía no vio fallar los tests con sus propios ojos.
func implementando(raiz string, fr roadmap.Feature, f *estado.Feature, g *global.Config) Instruccion {
	id := fr.ID

	// La base la arma el mismo helper que los otros tres estados de feature, y
	// de ahí salen el modelo, el via y el skill. Lo que sigue sólo ajusta el
	// mensaje y los sugeridos según por qué lote va.
	i := trabajarEnFeature(raiz, estado.Implementar, fr, f, "", g)
	if i.Tipo != Trabajar {
		return i // el modelo que hace falta no está declarado
	}
	// EL PLAN QUE ENVEJECE, y es el aviso más barato del diseño.
	//
	// Si planificaste f-2 y después implementaste f-1, la spec de f-2 quedó
	// mirando un repo QUE YA CAMBIÓ. Y el implementador que no encuentra lo que
	// la spec dice IMPROVISA — que es exactamente donde nacen los mocks.
	//
	// El mecanismo es de pobre a propósito: guardar un número al planificar y
	// compararlo al implementar. sf no sabe QUÉ cambió ni si importa; sabe que
	// el suelo se movió desde que se dibujó el plano. Avisa, no frena.
	if f.BaseCommit != "" && git.EsRepo(raiz) {
		if h, err := git.Head(raiz); err == nil && h != f.BaseCommit {
			i.Avisos = append(i.Avisos, fmt.Sprintf(
				"el plan de %s se escribió sobre %s y el repo ya se movió. Si la spec no coincide con lo que ves, DECILO — no improvises.",
				id, f.BaseCommit[:min(7, len(f.BaseCommit))]))
		}
	}

	// Sin lotes todavía no arrancó nada: los lotes se siembran en el primer
	// `sf lote start`, que es también donde se exige el rojo. Confundir esto con
	// "todos commiteados" mandaba a cerrar una feature en la que no se escribió
	// una línea — y en el camino corto de un bug pasaba siempre, porque ahí no
	// hay planificación que los siembre antes.
	if f.SinSembrar() {
		i.Mensaje = "escribí los tests y confirmá el rojo antes de implementar"
		i.Sugerido = []string{"sf context", "sf lote start"}
		return i
	}

	l, hay := f.LoteActual()
	if !hay {
		// Todos los lotes tienen commit y el estado no se movió: falta cerrar.
		// Acá no se lanza a nadie —lo que falta es un `sf done`—, así que la
		// instrucción va sin skill ni modelo.
		return Instruccion{
			Tipo:     Trabajar,
			Estado:   estado.Implementar,
			Feature:  id,
			DeLotes:  len(f.Lotes),
			Mensaje:  "todos los lotes están commiteados. Falta cerrar la feature.",
			Sugerido: []string{"sf done"},
		}
	}

	i.Lote = l.Lote
	i.DeLotes = len(f.Lotes)

	if !l.Rojo {
		i.Mensaje = "escribí los tests del lote y confirmá el rojo antes de implementar"
		i.Sugerido = []string{"sf context", "sf lote start"}
	} else {
		i.Mensaje = "el rojo está confirmado: implementá"
		i.Sugerido = []string{"sf context", "sf done --msg \"…\""}
	}
	return i
}

// ────────────────────────────────────────────────────────────────────────────
// Ayudantes
// ────────────────────────────────────────────────────────────────────────────

// trabajar arma la instrucción de "hay algo que hacer".
//
// Devuelve la 🛑 de modelo sin declarar cuando corresponde: es el único lugar
// donde el mapa puede frenar, y frenar acá —antes de decir "trabajá"— es lo que
// evita que el orquestador lance a alguien que no sabe invocar.
// trabajar arma la instrucción de los cinco estados de PRODUCTO.
//
// El modelo sale del mapa por estado y nada más, porque los dos niveles de
// arriba de la cadena —`sf model` y el ⑯ de tareas.json— viven DENTRO de una
// feature, y acá no hay ninguna. Los cuatro del ciclo usan
// `trabajarEnFeature`, que sí los aplica.
func trabajar(est, feature, mensaje string, g *global.Config) Instruccion {
	base := perfil(est)
	m, aviso, hay := lanzar(est, base, base, g)
	if !hay {
		return sinDeclarar(est, feature, base, g)
	}
	return Instruccion{
		Tipo:     Trabajar,
		Estado:   est,
		Feature:  feature,
		Skill:    skills[est],
		Perfil:   base,
		Modelo:   m.ID,
		Esfuerzo: m.Esfuerzo,
		Via:      m.Via,
		Comando:  m.Comando,
		Agente:   agenteDe(m, g),
		Avisos:   avisos(aviso),
		Mensaje:  mensaje,
		Sugerido: []string{"sf context", "sf done"},
	}
}

// trabajarEnFeature es `trabajar` con la cadena de modelo COMPLETA.
//
// ────────────────────────────────────────────────────────────────────────────
// LA PRECEDENCIA EXISTÍA Y SE APLICABA EN UN SOLO ESTADO DE CUATRO
// ────────────────────────────────────────────────────────────────────────────
//
// `modeloDeFeature` implementa los tres niveles —`sf model`, el ⑯ de
// tareas.json, el default del estado— y tenía UN solo consumidor:
// `implementando`. Los otros tres estados de feature pasaban por `trabajar`,
// que sólo mira el mapa por estado.
//
// O sea que `sf model opus` no hacía nada en `planificacion`, `revision` ni
// `cierre`: la salida de ME TRABÉ —la que Javier usa cuando está mirando el
// bucle patinar— no funcionaba en tres de los cuatro estados.
//
// `trabajar` se queda para los cinco de PRODUCTO, que no tienen feature: ahí no
// hay `sf model` ni ⑯ que aplicar, porque los dos viven dentro de una feature.
func trabajarEnFeature(raiz, est string, fr roadmap.Feature, f *estado.Feature,
	mensaje string, g *global.Config,
) Instruccion {
	base := perfil(est)
	pedido := pedidoDeFeature(raiz, fr, f, est)
	m, aviso, hay := lanzar(est, pedido, base, g)
	if !hay {
		return sinDeclarar(est, fr.ID, pedido, g)
	}
	return Instruccion{
		Tipo:     Trabajar,
		Estado:   est,
		Feature:  fr.ID,
		Skill:    skills[est],
		Perfil:   base,
		Modelo:   m.ID,
		Esfuerzo: m.Esfuerzo,
		Via:      m.Via,
		Comando:  m.Comando,
		Agente:   agenteDe(m, g),
		Avisos:   avisos(aviso),
		Mensaje:  mensaje,
		Sugerido: []string{"sf context", "sf done"},
	}
}

// perfil es qué pide un estado cuando nadie pidió otra cosa.
func perfil(est string) string {
	if p, hay := perfilPorEstado[est]; hay {
		return p
	}
	return perfilPorDefecto
}

// pedidoDeFeature resuelve la cadena de precedencia. Devuelve un PEDIDO —un
// perfil o un alias—, no un modelo: traducirlo es trabajo de `lanzar`.
//
//  1. estado.json            `sf model <x>` — la decisión de Javier, en runtime
//  2. tareas.json, del lote   el ⑯ — "el lote 3 necesita otro"
//  3. tareas.json, de la feature   el ⑯ — "toda esta feature necesita otro"
//  4. perfilPorEstado         el criterio del diseño
//
// Los niveles 2 y 3 son UN SOLO rung partido en dos, no uno nuevo: los dos son
// "lo que el ⑯ recomendó", con lo más específico arriba. La regla de siempre
// —cada nivel sabe menos que el de arriba— se mantiene.
//
// El nivel 1 es el lazo del ⑳: cuando Javier sube el modelo porque el bucle está
// patinando, esa decisión no está escrita en ningún archivo del plan y tiene que
// ganarle a todo — incluso a lo que el ⑯ recomendó, que se escribió antes de ver
// fallar nada.
//
// El nivel 2 nació de que el menú existe. `tareas.go` argumentaba que un modelo
// por lote "sería un campo repetido con el mismo valor", y razonaba bien desde lo
// que sabía: con UN modelo por perfil, sí. Con varios los valores son DISTINTOS
// —el lote de plomería va con el barato y el de concurrencia no— y esa diferencia
// es el feature entero.
//
// Un tareas.json ilegible NO es un error acá: de eso se queja la compuerta del
// ⑰, que corre antes. Si llegamos hasta acá con el archivo roto, caer al default
// es mejor que no poder decir qué sigue.
func pedidoDeFeature(raiz string, fr roadmap.Feature, f *estado.Feature, est string) string {
	if f.Modelo != "" {
		return f.Modelo
	}
	if p, err := tareas.Leer(raiz, fr.Carpeta()); err == nil {
		// El del lote primero: es el más específico de los dos que escribió el ⑯.
		if est == estado.Implementar {
			if l, hay := f.LoteActual(); hay {
				if a := p.ModeloDeLote(l.Lote); a != "" {
					return a
				}
			}
		}
		if p.Modelo != "" {
			return p.Modelo
		}
	}
	return perfil(est)
}

// lanzar traduce un pedido al modelo concreto de ESTA máquina.
//
// `brief` es el único que conversa: ①–⑤ es un pinponeo con Javier, y un subagente
// arranca, trabaja y muere — NO TE HABLA.
//
// El tercer retorno es la parada: `false` significa que el perfil que hace falta
// NO ESTÁ DECLARADO para este harness, y ahí sf no elige un reemplazo —eso sería
// opinar sobre qué modelo se parece a cuál, y R3 lo prohíbe—: para y pregunta.
//
// El segundo es un AVISO, y la asimetría con la parada es deliberada:
//
//	perfil sin declarar  →  🛑   no hay con qué lanzar nada, no hay salida
//	alias que falta      →  ⚠   sí la hay: el default del perfil, que es el
//	                            PRIMERO de su lista, o sea el más capaz
//
// Un alias faltante cae para el lado seguro —el peor caso es gastar de más— y
// frenar el bucle por una preferencia sería frenar sobre una opinión.
//
// Con `g == nil` —todavía no se corrió `sf install`— se cae a `subagente`. Es a
// propósito: no tener el catálogo no puede impedir trabajar, sólo impide resolver
// `consola`. La máquina funcionaba así antes de que el catálogo existiera.
func lanzar(est, pedido, base string, g *global.Config) (global.Modelo, string, bool) {
	if est == "brief" {
		return global.Modelo{Via: global.Vos}, "", true
	}
	if g == nil {
		return global.Modelo{Via: global.Subagente}, "", true
	}
	if m, hay := g.Resolver(pedido); hay {
		return m, "", true
	}
	// El pedido era un alias que acá no está declarado.
	if !global.EsPerfil(pedido) {
		if m, hay := g.Default(base); hay {
			return m, fmt.Sprintf(
				"el plan pide %q y el harness %q no lo tiene declarado — va con %q, el default de %s. "+
					"Declaralo con: sf model %s --alias %s --id … --via subagente",
				pedido, g.EnUso(), m.Alias, base, base, pedido), true
		}
	}
	return global.Modelo{}, "", false
}

// agenteDe es a quién invocar para conseguir este modelo, o "" si no hace falta.
//
// Vacío en Claude Code —ahí el modelo va como parámetro de la llamada— y vacío
// también cuando el trabajo no va por subagente: un `via: consola` lo ejecuta el
// orquestador con sus manos, y un `via: vos` lo hace Javier.
func agenteDe(m global.Modelo, g *global.Config) string {
	if g == nil || m.Via != global.Subagente || m.Alias == "" {
		return ""
	}
	if !global.NecesitaPortamodelo(g.EnUso()) {
		return ""
	}
	return global.NombreDeAgente(m.Alias)
}

// conQueSeTrabo nombra el modelo que estuvo fallando, para poder contestarle.
//
// Dice el pedido Y el alias que salió de él, porque desde que hay catálogo no
// son lo mismo: "falló con construir" no alcanza para decidir si conviene
// subirlo —hay varios modelos en construir— y saber cuál estuvo fallando es
// justo la mitad del dato que falta.
//
// Cuando los dos coinciden —el pedido ya era un alias— se dice uno solo: repetir
// la misma palabra dos veces es ruido.
func conQueSeTrabo(raiz string, fr roadmap.Feature, f *estado.Feature, g *global.Config) string {
	pedido := pedidoDeFeature(raiz, fr, f, f.Estado)
	m, hay := g.Resolver(pedido)
	if !hay || m.Alias == "" || m.Alias == pedido {
		return pedido
	}
	return fmt.Sprintf("%s (%s)", m.Alias, pedido)
}

// avisos envuelve un aviso que puede estar vacío.
//
// Existe para que los tres sitios que arman una Instruccion no repitan el mismo
// `if aviso != ""`: un aviso vacío tiene que dar una lista vacía y no una lista
// con un string vacío adentro, que es lo que se imprimiría como un renglón en
// blanco.
func avisos(a string) []string {
	if a == "" {
		return nil
	}
	return []string{a}
}

// sinDeclarar es la 🛑 de un perfil que este harness no tiene declarado.
//
// Las salidas declaran: contestar es lo que construye el catálogo. Es el mismo
// mecanismo que `dependencias_aprobadas` —comparar contra una lista que se llena
// con cada aprobación— aplicado a otra cosa. Un patrón ya firmado, no uno nuevo.
//
// El mensaje nombra el perfil Y el harness, y dice CÓMO averiguar los ids en ese
// harness. Una parada que te deja sin saber qué contestar es una parada mal
// escrita: son dos preguntas que se contestan una vez en la vida de la máquina,
// y no tienen que costar una búsqueda.
func sinDeclarar(est, feature, pedido string, g *global.Config) Instruccion {
	harness := "este harness"
	if g != nil && g.EnUso() != "" {
		harness = g.EnUso()
	}
	i := Instruccion{
		Tipo:    Para,
		Estado:  est,
		Feature: feature,
		Perfil:  pedido,
		Mensaje: fmt.Sprintf(
			"🛑 PARÁ. El paso pide el perfil %q y el harness %s no lo tiene declarado.\n"+
				"   sf no elige el modelo: decime vos cuál es acá.\n"+
				"   (los ids de tu harness: `opencode models` · `/model` en Claude Code y Command Code)",
			pedido, harness),
		Sugerido: []string{
			fmt.Sprintf("sf model %s --alias <corto> --id <id-del-harness> --via subagente", pedido),
			fmt.Sprintf("sf model %s --alias <corto> --id <id> --via consola --comando \"…\"", pedido),
		},
	}
	if global.NecesitaPortamodelo(harness) {
		// Medido: los agentes se leen al arrancar. Si no se dice acá, se descubre
		// fallando, que es la peor forma de enterarse de algo que ya se sabía.
		i.Avisos = append(i.Avisos,
			"después de declararlos, reiniciá tu harness: los agentes se leen al arrancar.")
	}
	return i
}

// existe pregunta por un archivo relativo a la raíz.
//
// Devuelve bool y se traga el error a propósito: para el flujo, "no lo puedo
// leer" y "no está" llevan al mismo lado (hay que producirlo). La distinción sí
// importa al LEER un archivo —y ahí está hecha, en estado.Leer— pero no al
// preguntar si hace falta trabajarlo.
func existe(raiz, rel string) bool {
	_, err := os.Stat(filepath.Join(raiz, rel))
	return err == nil
}

// hayHistorias mira si el ⑨ ya produjo algo.
//
// Un glob y no un contador: no importa cuántas hay, importa si hay. El conteo
// de criterios lo hace `sf done`, que es quien tiene que frenar.
// mensajeDelNueve distingue las dos veces que se entra al ⑨.
//
// La primera es partir el PRD y no hay nada escrito. Las otras son completar lo
// que entró por `sf new`, y ahí decir "el ⑨ parte el PRD en historias" mandaría
// a rehacer el backlog entero: lo que falta es UNA historia, y nombrarla es la
// diferencia entre una instrucción y una consigna.
//
// El corte es "¿falta alguna, o faltan todas?". Con todas incompletas todavía
// no hay backlog del cual completar nada, así que sigue siendo partir el PRD.
func mensajeDelNueve(ids, faltan []string) string {
	if len(faltan) == len(ids) {
		return "el ⑨ parte el PRD en historias"
	}
	return "completá: " + strings.Join(faltan, " · ")
}
