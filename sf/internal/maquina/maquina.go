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
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/global"
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

	// Skill, Modelo y Via son la mitad del hallazgo H1: si sf no los devuelve,
	// la tabla estado→skill→modelo vive en CLAUDE.md y hay que mantener una
	// copia por harness. Devolviéndolos, CLAUDE.md queda en cuatro líneas y
	// AGENTS.md es el mismo texto.
	Skill  string
	Modelo string
	Via    string

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
var modeloPorEstado = map[string]string{
	estado.Planificacion: "opus",
	estado.Revision:      "opus",
}

const modeloPorDefecto = "sonnet"

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
	if !e.Producto.ConstitucionSellada {
		if !existe(raiz, docs.Constitucion) {
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
	if !e.Producto.BacklogVisto {
		if !hayHistorias(raiz) {
			return trabajar("backlog", "", "el ⑨ parte el PRD en historias", g), true
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
		return tomarLaProxima(e, r, g)
	}

	// ME TRABÉ va PRIMERO, antes de mirar en qué estado está: si el contador
	// llegó al tope, no importa qué falta — importa que hace rato que no
	// avanza. Ponerlo después dejaría que sf siga proponiendo trabajo mientras
	// el bucle patina.
	if f.IntentosFallidos >= TopeIntentos {
		return Instruccion{
			Tipo:    MeTrabe,
			Estado:  f.Estado,
			Feature: e.FeatureActual,
			Mensaje: fmt.Sprintf("⚠ ME TRABÉ. %s falló %d veces seguidas con %s.\n"+
				"   ¿Subo el modelo, o entrás vos?",
				e.FeatureActual, f.IntentosFallidos, f.Modelo),
			Sugerido: []string{"sf model <nombre>", "sf dismiss <h-#> \"motivo\""},
		}
	}

	fr, _ := r.Buscar(e.FeatureActual)

	// El camino corto (maquina-estados.md §8): un bug no pasa por planificación
	// ni por revisión. No es un carril paralelo —eso sería una segunda máquina
	// que mantener— sino la MISMA máquina con estados salteados.
	//
	// Se calcula en una variable local y NO se escribe en f: `sf next` es
	// consulta pura y correrlo dos veces tiene que dar lo mismo. El que mueve el
	// estado es `sf done`, y allá el mismo salteo se aplica de nuevo.
	actual := f.Estado
	if actual == estado.Planificacion && esDeBugs(raiz, fr) {
		actual = estado.Implementar
	}

	switch actual {
	case estado.Planificacion:
		return planificando(raiz, e.FeatureActual, fr, g)

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
		return trabajar(estado.Revision, e.FeatureActual,
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
		return trabajar(estado.Cierre, e.FeatureActual,
			"el ㉓ escribe la doc y el journal", g)

	case estado.Cerrada:
		// La feature actual quedó cerrada y nadie tomó otra: sigue la cola.
		return tomarLaProxima(e, r, g)
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
func tomarLaProxima(e *estado.Estado, r *roadmap.Roadmap, g *global.Config) Instruccion {
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
	m := modelo(estado.Planificacion)
	v, cmd, declarado := via(estado.Planificacion, m, g)
	if !declarado {
		return sinDeclarar(estado.Planificacion, prox.ID, m)
	}

	return Instruccion{
		Tipo:     Trabajar,
		Estado:   estado.Planificacion,
		Feature:  prox.ID,
		Skill:    skills[estado.Planificacion],
		Modelo:   m,
		Via:      v,
		Comando:  cmd,
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
func planificando(raiz, id string, fr roadmap.Feature, g *global.Config) Instruccion {
	i := trabajar(estado.Planificacion, id, "", g)

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
	// El modelo se resuelve UNA VEZ acá y no en cada rama: las tres salidas de
	// esta función que lanzan trabajo usan el mismo, y leer tareas.json tres
	// veces sería tocar el disco de más para obtener siempre lo mismo.
	m := modeloDeFeature(raiz, fr, f, estado.Implementar)

	// Y el `via` sale del mapa, igual que en `trabajar`. Acá el caso pesa más
	// que en ningún otro estado: es JUSTO donde el ⑯ recomienda un modelo
	// distinto, y donde Javier sube el modelo con `sf model` cuando el bucle
	// patina. Los dos caminos pueden traer un nombre que el mapa no conoce.
	v, cmd, declarado := via(estado.Implementar, m, g)
	if !declarado {
		return sinDeclarar(estado.Implementar, id, m)
	}

	// Sin lotes todavía no arrancó nada: los lotes se siembran en el primer
	// `sf lote start`, que es también donde se exige el rojo. Confundir esto con
	// "todos commiteados" mandaba a cerrar una feature en la que no se escribió
	// una línea — y en el camino corto de un bug pasaba siempre, porque ahí no
	// hay planificación que los siembre antes.
	if len(f.Lotes) == 0 {
		return Instruccion{
			Tipo:     Trabajar,
			Estado:   estado.Implementar,
			Feature:  id,
			Skill:    skills[estado.Implementar],
			Modelo:   m,
			Via:      v,
			Comando:  cmd,
			Mensaje:  "escribí los tests y confirmá el rojo antes de implementar",
			Sugerido: []string{"sf context", "sf lote start"},
		}
	}

	l, hay := f.LoteActual()
	if !hay {
		// Todos los lotes tienen commit y el estado no se movió: falta cerrar.
		return Instruccion{
			Tipo:     Trabajar,
			Estado:   estado.Implementar,
			Feature:  id,
			DeLotes:  len(f.Lotes),
			Mensaje:  "todos los lotes están commiteados. Falta cerrar la feature.",
			Sugerido: []string{"sf done"},
		}
	}

	i := Instruccion{
		Tipo:    Trabajar,
		Estado:  estado.Implementar,
		Feature: id,
		Lote:    l.Lote,
		DeLotes: len(f.Lotes),
		Skill:   skills[estado.Implementar],
		Modelo:  m,
		Via:     v,
		Comando: cmd,
	}

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
func trabajar(est, feature, mensaje string, g *global.Config) Instruccion {
	m := modelo(est)
	v, cmd, hay := via(est, m, g)
	if !hay {
		return sinDeclarar(est, feature, m)
	}
	return Instruccion{
		Tipo:     Trabajar,
		Estado:   est,
		Feature:  feature,
		Skill:    skills[est],
		Modelo:   m,
		Via:      v,
		Comando:  cmd,
		Mensaje:  mensaje,
		Sugerido: []string{"sf context", "sf done"},
	}
}

func modelo(est string) string {
	if m, hay := modeloPorEstado[est]; hay {
		return m
	}
	return modeloPorDefecto
}

// modeloDeFeature resuelve la cadena de precedencia de tres niveles.
//
//  1. estado.json     `sf model <nombre>` — la decisión de Javier, en runtime
//  2. tareas.json     el ⑯ — "esta feature necesita uno más grande"
//  3. el default del estado
//
// El nivel 1 es el lazo del ⑳: cuando Javier sube el modelo porque el bucle
// está patinando, esa decisión no está escrita en ningún archivo del plan y
// tiene que ganarle a todo — incluso a lo que el ⑯ recomendó, que se escribió
// antes de ver fallar nada.
//
// El nivel 2 es el ⑯, y es un campo y no un comando (regla 1.1): lo escribe el
// mismo que planificó, en el mismo archivo, en la misma pasada.
//
// Un tareas.json ilegible NO es un error acá: quien se queja de eso es la
// compuerta del ⑰, que corre antes. Si llegamos hasta acá con el archivo roto,
// caer al default es mejor que no poder decir qué sigue.
func modeloDeFeature(raiz string, fr roadmap.Feature, f *estado.Feature, est string) string {
	if f.Modelo != "" {
		return f.Modelo
	}
	if p, err := tareas.Leer(raiz, fr.Carpeta()); err == nil && p.Modelo != "" {
		return p.Modelo
	}
	return modelo(est)
}

// via es quién hace el trabajo, y tiene tres valores (H17).
//
// `brief` es el único que conversa: ①–⑤ es un pinponeo con Javier, y un
// subagente arranca, trabaja y muere — NO TE HABLA. El resto sale del mapa.
//
// El segundo retorno es la parada: `false` significa que el modelo que hace
// falta NO ESTÁ DECLARADO, y ahí sf no elige un reemplazo (eso sería opinar
// sobre qué modelo se parece a cuál, y R3 lo prohíbe): para y pregunta.
//
// Con `g == nil` —todavía no se corrió `sf install`— se cae a `subagente`. Es a
// propósito: no tener el mapa no puede impedir trabajar, sólo impide resolver
// `consola`. La máquina funcionaba así antes de que el mapa existiera y sigue
// funcionando igual.
func via(est, modelo string, g *global.Config) (string, string, bool) {
	if est == "brief" {
		return global.Vos, "", true
	}
	if g == nil {
		return global.Subagente, "", true
	}
	m, hay := g.Buscar(modelo)
	if !hay {
		return "", "", false
	}
	return m.Via, m.Comando, true
}

// sinDeclarar es la 🛑 de un modelo que no está en el mapa.
//
// Las tres salidas son las del diseño, y las tres declaran: contestar es lo que
// construye la lista. Es el mismo mecanismo que `dependencias_aprobadas`
// —comparar contra una lista que se llena con cada aprobación— aplicado a otra
// cosa. Un patrón ya firmado, no uno nuevo.
func sinDeclarar(est, feature, modelo string) Instruccion {
	return Instruccion{
		Tipo:    Para,
		Estado:  est,
		Feature: feature,
		Modelo:  modelo,
		Mensaje: fmt.Sprintf(
			"🛑 PARÁ. Hace falta %q y no está declarado en ~/.specforge/modelos.yaml.\n"+
				"   sf no elige el reemplazo: decime vos cómo se lanza acá.", modelo),
		Sugerido: []string{
			fmt.Sprintf("sf model %s --via subagente", modelo),
			fmt.Sprintf("sf model %s --via consola --comando \"…\"", modelo),
			"sf model <otro>",
		},
	}
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
func hayHistorias(raiz string) bool {
	m, err := filepath.Glob(filepath.Join(raiz, docs.Backlog, "us-*.md"))
	return err == nil && len(m) > 0
}
