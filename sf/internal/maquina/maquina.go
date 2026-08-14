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
// ⚠ Tres de estos nombres son provisionales, y no es descuido: los skills
// todavía no se adaptaron a la máquina (que-sobrevive.md §14 los tiene con
// veredicto pero sin tocar).
//
//	sfp-po           NO EXISTE todavía — es el hueco 2 de que-sobrevive §13
//	sf-propose       cubre tres estados y hay que PARTIRLO (⑥, ⑨, ⑩)
//	sfx-think        es el ⑫; falta ver si `planificacion` entero es un skill
var skills = map[string]string{
	"brief":              "sfp-scout",
	"prd":                "sfp-po",
	"constitucion":       "sf-init",
	"backlog":            "sf-propose",
	"roadmap":            "sf-propose",
	estado.Planificacion: "sfx-think",
	estado.Implementar:   "sfx-tdd",
	estado.Revision:      "sf-check",
	estado.Cierre:        "sfx-documenter",
}

// modeloPorEstado es lo que el paso PIDE, todavía sin cruzar con lo que el
// harness PUEDE.
//
// ⚠ Esto es la mitad de H1b y está a propósito incompleto. La versión completa
// necesita dos cosas que hoy no existen:
//
//	tareas.json     el ⑯ escribe qué modelo hace falta para esta feature
//	~/.specforge/   qué modelos tiene Javier y cómo se invoca cada uno
//
// Hasta que existan, se devuelve el criterio del diseño —"la planificación la
// debería hacer un modelo como Opus grande con amplio contexto; la
// implementación un subagente; la revisión también"— y `via: subagente`.
//
// Cuando `tareas.json` exista, el modelo de la feature lo pisa; cuando exista
// el mapa de modelos, sale el `via` de verdad (subagente | consola).
var modeloPorEstado = map[string]string{
	estado.Planificacion: "opus",
	estado.Revision:      "opus",
}

const modeloPorDefecto = "sonnet"

// Siguiente decide qué sigue. Es toda la lógica de `sf next`.
//
// `raiz` es la raíz del proyecto: hace falta porque varias decisiones se toman
// mirando qué archivos existen, que es el truco de los checkpoints.
func Siguiente(raiz string, e *estado.Estado, r *roadmap.Roadmap) Instruccion {
	// El orden importa y es el del flujo: los cinco de producto primero, y
	// recién cuando están los cinco empieza el ciclo de feature. No hay forma
	// de planificar sin constitución.
	if i, hay := siguienteDeProducto(raiz, e); hay {
		return i
	}
	return siguienteDeFeature(raiz, e, r)
}

// ────────────────────────────────────────────────────────────────────────────
// Producto — los cinco que corren una vez
// ────────────────────────────────────────────────────────────────────────────

func siguienteDeProducto(raiz string, e *estado.Estado) (Instruccion, bool) {
	// ① brief. El sello es un veredicto, no un bool: vacío = no sellado.
	if e.Producto.BriefSellado == "" {
		// Acá está el patrón que se repite en los tres primeros estados, y es
		// la diferencia entre "falta hacerlo" y "está hecho, falta que lo
		// mires": lo separa la EXISTENCIA DEL ARCHIVO, no un campo.
		if !existe(raiz, docs.Brief) {
			return trabajar("brief", "", "el ⑥ lo sellás vos cuando esté"), true
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
		return trabajar("prd", "", "sin parada: del ⑦ se pasa derecho al ⑧"), true
	}

	// ⑧ constitución.
	if !e.Producto.ConstitucionSellada {
		if !existe(raiz, docs.Constitucion) {
			return trabajar("constitucion", "", "el ⑧ lo sellás vos cuando esté"), true
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
			return trabajar("backlog", "", "el ⑨ parte el PRD en historias"), true
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
		return trabajar("roadmap", "", "el ⑩ agrupa las historias en features"), true
	}

	return Instruccion{}, false
}

// ────────────────────────────────────────────────────────────────────────────
// Feature — los cuatro que se repiten, una vuelta por feature
// ────────────────────────────────────────────────────────────────────────────

func siguienteDeFeature(raiz string, e *estado.Estado, r *roadmap.Roadmap) Instruccion {
	// Sin roadmap no hay cola. No debería pasar (siguienteDeProducto ya lo
	// habría atajado), pero un nil acá sería un panic y no un mensaje.
	if r == nil {
		return Instruccion{Tipo: Fin, Mensaje: "No hay roadmap.json: falta el ⑩."}
	}

	f, hay := e.Actual()
	if !hay {
		return tomarLaProxima(e, r)
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

	switch f.Estado {
	case estado.Planificacion:
		return planificando(raiz, e.FeatureActual, fr)

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
		return implementando(e.FeatureActual, f)

	case estado.Revision:
		return trabajar(estado.Revision, e.FeatureActual,
			"el ㉑ juzga los criterios y el ㉒ mira los mutantes, en la misma pasada")

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
			"el ㉓ escribe la doc y el journal")

	case estado.Cerrada:
		// La feature actual quedó cerrada y nadie tomó otra: sigue la cola.
		return tomarLaProxima(e, r)
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
func tomarLaProxima(e *estado.Estado, r *roadmap.Roadmap) Instruccion {
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

	return Instruccion{
		Tipo:     Trabajar,
		Estado:   estado.Planificacion,
		Feature:  prox.ID,
		Skill:    skills[estado.Planificacion],
		Modelo:   modelo(estado.Planificacion),
		Via:      "subagente",
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
func planificando(raiz, id string, fr roadmap.Feature) Instruccion {
	i := trabajar(estado.Planificacion, id, "")

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
func implementando(id string, f *estado.Feature) Instruccion {
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
		Modelo:  modeloDeFeature(f, estado.Implementar),
		Via:     "subagente",
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

func trabajar(est, feature, mensaje string) Instruccion {
	return Instruccion{
		Tipo:     Trabajar,
		Estado:   est,
		Feature:  feature,
		Skill:    skills[est],
		Modelo:   modelo(est),
		Via:      via(est),
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

// modeloDeFeature deja que el modelo guardado en el estado pise al del mapa.
//
// Es el lazo del ⑳: cuando Javier sube el modelo con `sf model`, esa decisión
// no está escrita en ningún archivo y tiene que ganarle al default.
func modeloDeFeature(f *estado.Feature, est string) string {
	if f.Modelo != "" {
		return f.Modelo
	}
	return modelo(est)
}

// via es quién hace el trabajo, y tiene tres valores (H17).
//
// `brief` es el único que conversa: ①–⑤ es un pinponeo con Javier, y un
// subagente arranca, trabaja y muere — NO TE HABLA. El resto va a subagente.
//
// El tercer valor —`consola`, para modelos de otro proveedor— todavía no se
// puede decidir acá: necesita saber en qué harness corre y qué modelos hay
// declarados, y ninguna de las dos cosas existe. Ver modeloPorEstado.
func via(est string) string {
	if est == "brief" {
		return "vos"
	}
	return "subagente"
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
