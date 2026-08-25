package maquina

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/compuerta"
	"github.com/jdbaigorria/specforge/sf/internal/constitucion"
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/historia"
	"github.com/jdbaigorria/specforge/sf/internal/revision"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
	"github.com/jdbaigorria/specforge/sf/internal/suite"
	"github.com/jdbaigorria/specforge/sf/internal/tareas"
)

// Cierre es lo que pasó al correr `sf done`.
type Cierre struct {
	compuerta.Resultado

	// Movio dice si el estado avanzó de verdad.
	//
	// No es lo mismo que Pasa(): las compuertas pueden pasar y el estado
	// quedarse igual, porque lo que sigue es una parada que decide Javier. El
	// ⑰ es exactamente eso.
	Movio bool

	// Cambio dice si hay que guardar el estado.json.
	//
	// Es más amplio que Movio, y la diferencia importa en un caso concreto: un
	// `sf done` que FALLA dentro de una feature igual incrementa
	// `intentos_fallidos`, que es el contador de ME TRABÉ. Si sólo se guardara
	// cuando el estado avanza, el contador nunca subiría y el bucle no tendría
	// freno.
	Cambio bool

	// Mensaje es qué pasó, para el que lo lee.
	Mensaje string
}

// Terminar corre las compuertas del estado actual y mueve lo que corresponda.
//
// ────────────────────────────────────────────────────────────────────────────
// `sf done` SIGNIFICA SIEMPRE LO MISMO: "TERMINÉ LO QUE ME TOCABA"
// ────────────────────────────────────────────────────────────────────────────
//
// Y qué le tocaba lo sabe sf, no el que trabajó (H8). Por eso no existe un
// `sf lote done` aparte: en el lote 1 de 4 cierra el lote, y en el 4 de 4
// cierra el lote Y el estado. El que trabaja nunca necesita saber si era el
// último.
//
// `msg` es el mensaje del commit y sólo se usa en `implementar`. Sin él ahí, no
// hay done — porque cerrar el lote ES commitear (H10).
func Terminar(raiz string, e *estado.Estado, r *roadmap.Roadmap, msg string) Cierre {
	if c, hay := terminarProducto(raiz, e); hay {
		return c
	}
	return terminarFeature(raiz, e, r, msg)
}

// ────────────────────────────────────────────────────────────────────────────
// Producto
// ────────────────────────────────────────────────────────────────────────────

func terminarProducto(raiz string, e *estado.Estado) (Cierre, bool) {
	switch {
	case e.Producto.BriefSellado == "":
		// Las compuertas pasan pero el estado NO se mueve: el ⑥ lo sella
		// Javier, y sellarlo es elegir uno de tres veredictos. sf no puede
		// elegir por él ni siquiera cuando todo está bien.
		return paraDeProducto(compuerta.Brief(raiz), "el brief está listo. Lo sellás vos (el ⑥)."), true

	case e.Producto.PrdHash == "":
		// El ⑦ no tiene parada: se pasa derecho al ⑧. Acá sf sí mueve, y de
		// paso guarda la huella del PRD para poder avisar después si cambia.
		r := compuerta.PRD(raiz)
		if !r.Pasa() {
			return Cierre{Resultado: r}, true
		}
		h, err := git.HashDe(raiz, []string{docs.PRD})
		if err != nil {
			// Sin git no hay hash, y eso no puede frenar el flujo: el hash es
			// para un AVISO futuro, no para una compuerta.
			h = "sin-hash"
		}
		e.Producto.PrdHash = h
		return Cierre{Resultado: r, Movio: true, Cambio: true,
			Mensaje: "prd_hash " + h + " — sigue el ⑧"}, true

	case !e.Producto.ConstitucionSellada:
		return paraDeProducto(compuerta.Constitucion(raiz),
			"la constitución está lista. La sellás vos (el ⑧)."), true

	case !e.Producto.BacklogVisto:
		return paraDeProducto(compuerta.Backlog(raiz),
			"salieron las historias. Mirá si querés y seguí (⏸)."), true
	}
	return Cierre{}, false
}

// paraDeProducto arma el cierre de un estado que termina en parada.
//
// Las compuertas corrieron y pueden haber pasado — pero el estado no se mueve,
// porque lo que sigue es una decisión de Javier. Movio queda en false a
// propósito: `sf approve` es el que mueve.
func paraDeProducto(r compuerta.Resultado, mensaje string) Cierre {
	c := Cierre{Resultado: r}
	if r.Pasa() {
		c.Mensaje = mensaje
	}
	return c
}

// ────────────────────────────────────────────────────────────────────────────
// Feature
// ────────────────────────────────────────────────────────────────────────────

func terminarFeature(raiz string, e *estado.Estado, r *roadmap.Roadmap, msg string) Cierre {
	f, hay := e.Actual()
	if !hay {
		// Sin feature en curso, el `sf done` que llega es el del ⑩: el
		// subagente del roadmap acaba de escribirlo y avisa que terminó.
		//
		// ────────────────────────────────────────────────────────────────
		// LA COMPUERTA DEL ⑩ CORRÍA EN EL ÚNICO CASO EN QUE NO SIRVE
		// ────────────────────────────────────────────────────────────────
		//
		// Estaba colgada de `r == nil`, y `r` es nil sólo cuando NO HAY
		// roadmap.json — o sea que `compuerta.Roadmap` fallaba en su primera
		// línea, al intentar leerlo. Sus dos chequeos reales —que ninguna
		// historia quedara fuera de todas las features, y el aviso de la
		// feature que junta seis— no se ejecutaban nunca.
		//
		// Y con el roadmap ya escrito el `sf done` del ⑩ caía acá abajo y
		// contestaba "no hay feature en curso: corré `sf take`", que le decía
		// al subagente que hizo bien su trabajo que se había equivocado.
		//
		// El estado no se mueve, y eso es correcto: el ⑩ no guarda sello
		// —"¿existe el roadmap.json?" es deducible (R6)— y la transición al
		// ciclo la hace `sf take`, que es una decisión de Javier (H4).
		res := compuerta.Roadmap(raiz)
		c := Cierre{Resultado: res}
		if res.Pasa() {
			c.Mensaje = "el roadmap está listo. Elegí con `sf take <feature>`."
		}
		return c
	}
	fr, enRoadmap := r.Buscar(e.FeatureActual)
	if !enRoadmap {
		var c Cierre
		c.Fallas = []string{fmt.Sprintf("%s no está en el roadmap", e.FeatureActual)}
		return c
	}

	// El mismo salteo del camino corto que hace `sf next`, y acá SÍ se escribe:
	// `sf done` es el que mueve el estado. La regla vive en estado.Efectivo —
	// tenerla copiada en cada comando fue lo que dejó a `sf lote start`
	// contradiciendo a `sf next`.
	f.Estado = estado.Efectivo(f.Estado, historia.SonTodasBugs(raiz, fr.Historias))

	switch f.Estado {
	case estado.Planificacion:
		return cerrarPlanificacion(raiz, f, fr)
	case estado.Implementar:
		return cerrarLote(raiz, f, fr, msg)
	case estado.Revision:
		return cerrarRevision(raiz, f, fr)
	case estado.Cierre:
		return cerrarCierre(raiz, f, fr)
	}

	var c Cierre
	c.Fallas = []string{fmt.Sprintf("%s está en %q y ahí no hay nada que cerrar", fr.ID, f.Estado)}
	return c
}

// cerrarPlanificacion corre las cinco compuertas y anota el base_commit.
//
// El estado NO se mueve aunque todo pase: lo que sigue es el ⑰, que es una de
// las tres decisiones de Javier. Lo que sí queda anotado es `base_commit`, y es
// justo el momento correcto para anotarlo: el plan se dibujó mirando ESTE
// commit, y compararlo después es lo que detecta el plan que envejece.
func cerrarPlanificacion(raiz string, f *estado.Feature, fr roadmap.Feature) Cierre {
	res := compuerta.Planificacion(raiz, fr)
	c := Cierre{Resultado: res}
	if !res.Pasa() {
		return c
	}

	if h, err := git.Head(raiz); err == nil {
		f.BaseCommit = h
		c.Cambio = true
	}
	c.Mensaje = fmt.Sprintf("el plan de %s está listo. Lo revisás vos (el ⑰).", fr.ID)
	return c
}

// cerrarLote es donde mueren los dolores #2 y #3.
//
// Tres cosas pasan acá, y el orden importa:
//
//  1. ¿vi el rojo?      si no, no hay verde que valga (compuerta.Implementar)
//  2. ¿cambió el test?  el agujero astuto del #8
//  3. commit            lo hace sf, con el mensaje que trajo el que trabajó
//
// Y cerrar el lote ES commitear: si falta el mensaje, no hay done.
func cerrarLote(raiz string, f *estado.Feature, fr roadmap.Feature, msg string) Cierre {
	res := compuerta.Implementar(f)
	c := Cierre{Resultado: res}
	if !res.Pasa() {
		f.IntentosFallidos++
		c.Cambio = true
		return c
	}

	l, hayLote := f.LoteActual()
	if !hayLote {
		// Todos los lotes tienen commit: se cierra el estado entero. Y acá está
		// la otra mitad del camino corto — un bug tampoco pasa por revisión.
		//
		// Que acá "no hay lote en curso" signifique de verdad "todos cerrados"
		// y no "nunca empezó" lo garantiza compuerta.Implementar, que corrió
		// arriba y frena con f.Lotes vacío.
		if historia.SonTodasBugs(raiz, fr.Historias) {
			f.Estado = estado.Cierre
			return Cierre{Movio: true, Cambio: true,
				Mensaje: "lotes cerrados — es un bug, se saltea la revisión: sigue el cierre (㉓)"}
		}
		f.Estado = estado.Revision
		return Cierre{Movio: true, Cambio: true, Mensaje: "todos los lotes cerrados — sigue la revisión (㉑)"}
	}

	if strings.TrimSpace(msg) == "" {
		c.Fallas = append(c.Fallas, "falta el mensaje del commit: `sf done --msg \"…\"`. El lote no se cierra sin él.")
		return c
	}

	// El verde y el hash: las dos mitades que cierran la compuerta que
	// `sf lote start` abrió.
	if fallas := verdeYHash(raiz, fr, l); len(fallas) > 0 {
		c.Fallas = append(c.Fallas, fallas...)
		f.IntentosFallidos++
		c.Cambio = true
		return c
	}

	hash, err := git.Commit(raiz, msg)
	if err != nil {
		c.Fallas = append(c.Fallas, err.Error())
		f.IntentosFallidos++
		c.Cambio = true
		return c
	}

	l.Commit = &hash
	f.IntentosFallidos = 0 // el lote cerró: el contador de ME TRABÉ se resetea

	c.Movio, c.Cambio = true, true
	c.Mensaje = fmt.Sprintf("lote %d cerrado en %s", l.Lote, hash)
	return c
}

// cerrarRevision decide si vuelve a implementar o sigue al cierre.
//
// Es la única transición del ciclo que puede ir PARA ATRÁS, y la decisión es un
// conteo: con un hallazgo abierto no avanza. sf cuenta, no opina.
func cerrarRevision(raiz string, f *estado.Feature, fr roadmap.Feature) Cierre {
	res := compuerta.Revision(raiz, fr)
	c := Cierre{Resultado: res}

	if !res.Pasa() {
		// Si lo que falta son hallazgos abiertos, el camino es volver a
		// implementar — no es un error del revisor. Se distingue mirando el
		// archivo, no el texto de la falla.
		if hayAbiertos(raiz, fr) {
			f.Estado = estado.Implementar
			l := abrirLoteDeCorreccion(f)
			return Cierre{Movio: true, Cambio: true, Resultado: res,
				Mensaje: fmt.Sprintf(
					"hay hallazgos abiertos — vuelve a implementar en el lote %d", l)}
		}
		return c
	}

	f.Estado = estado.Cierre
	c.Movio, c.Cambio = true, true
	c.Mensaje = "revisión limpia — sigue el cierre (㉓)"
	return c
}

// abrirLoteDeCorreccion le da al ㉑ dónde arreglar lo que encontró.
//
// ────────────────────────────────────────────────────────────────────────────
// SIN ESTO, LA VUELTA DEL ㉑ ERA UN LOOP MUERTO
// ────────────────────────────────────────────────────────────────────────────
//
// El estado volvía a `implementar` y no había nada que implementar: los lotes
// del plan ya estaban todos commiteados, así que `sf lote start` contestaba
// "todos los lotes ya están commiteados" y `sf done` rebotaba a `revision`.
// El arreglo del hallazgo no tenía dónde commitearse, y la única salida era
// `sf dismiss` — o sea, declarar falso lo que el revisor encontró.
//
// El lote es la unidad de trabajo que la máquina ya tiene, y el arreglo de un
// hallazgo es exactamente eso: un cambio con su test, que termina en un commit.
// Abrir uno nuevo mantiene "un lote, un commit" (los dolores #2 y #3) en vez de
// inventar un carril donde el fix se commitea suelto.
//
// El lote nuevo NO está en tareas.json —el plan se escribió antes de que el
// hallazgo existiera—, así que `sf lote start` cae en la rama del lote de
// corrección: no puede exigir CUÁLES tests, pero sigue exigiendo que la suite
// falle. Y eso es lo correcto acá: un hallazgo sin un test que lo reproduzca es
// un hallazgo que nadie va a poder verificar.
//
// Devuelve el número, que es lo único que el mensaje necesita.
func abrirLoteDeCorreccion(f *estado.Feature) int {
	n := 1
	for _, l := range f.Lotes {
		if l.Lote >= n {
			n = l.Lote + 1
		}
	}
	f.Lotes = append(f.Lotes, estado.Lote{Lote: n})
	return n
}

func hayAbiertos(raiz string, fr roadmap.Feature) bool {
	rev, err := revision.Leer(filepath.Join(raiz, fr.Carpeta(), docs.Revision))
	return err == nil && len(rev.Abiertos()) > 0
}

// cerrarCierre deja la feature lista para archivar, y no la archiva.
//
// Archivar —mover la carpeta, mergear, borrar la branch— es de `sf approve`
// (H16), porque pasa DESPUÉS de la ⏸. Acá sólo se comprueba que los dos
// archivos del ㉓ existan.
func cerrarCierre(raiz string, f *estado.Feature, fr roadmap.Feature) Cierre {
	res := compuerta.Cierre(raiz, fr)
	c := Cierre{Resultado: res}
	if res.Pasa() {
		c.Mensaje = fmt.Sprintf("%s lista para archivar (⏸).", fr.ID)
	}
	return c
}

// verdeYHash cierra la compuerta que `sf lote start` abrió.
//
// ────────────────────────────────────────────────────────────────────────────
// SON DOS CHEQUEOS Y EL SEGUNDO ES EL QUE NADIE ESPERA
// ────────────────────────────────────────────────────────────────────────────
//
//	¿la suite pasa?              lo obvio
//	¿los tests son LOS MISMOS?   el agujero astuto del #8
//
// El segundo existe porque un subagente que no logra implementar puede ablandar
// el assert y llegar a verde. sf vio el rojo, vio el verde, y entre medio lo que
// cambió fue el test — sin el hash, eso pasa sin que nadie se entere.
//
// Devuelve las fallas; vacío significa que se puede commitear.
func verdeYHash(raiz string, fr roadmap.Feature, l *estado.Lote) []string {
	c, err := constitucion.Leer(raiz)
	if err != nil {
		return []string{err.Error()}
	}

	res, err := suite.Correr(raiz, c.TestCmd)
	if err != nil {
		return []string{"no pude correr los tests: " + err.Error()}
	}
	if !res.Verde {
		return []string{"los tests todavía fallan. No commiteo en rojo.", ultimasLineas(res.Salida, 15)}
	}

	// El hash sólo se compara si `sf lote start` llegó a tomarlo. Sin git no lo
	// hay, y eso ya se avisó en su momento: no se puede exigir acá algo que no
	// se pudo guardar allá.
	if l.HashTests == "" {
		return nil
	}

	p, err := tareas.Leer(raiz, fr.Carpeta())
	if err != nil {
		return []string{err.Error()}
	}
	ahora, err := git.HashDe(raiz, suite.Archivos(p.TestsDelLote(l.Lote)))
	if err != nil {
		return nil // sin git no hay comparación posible; el verde ya se vio
	}
	if ahora != l.HashTests {
		return []string{
			"los archivos de test CAMBIARON entre el rojo y el verde.",
			"    Un test que se afloja para llegar a verde no prueba nada.",
			"    Si el cambio es legítimo, volvé a correr `sf lote start`.",
		}
	}
	return nil
}

// ultimasLineas recorta la salida del runner.
//
// Una suite grande escupe cientos de líneas y el que lee esto es un modelo con
// contexto acotado. Las últimas son donde los runners ponen el resumen y los
// fallos.
func ultimasLineas(s string, n int) string {
	lineas := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lineas) > n {
		lineas = lineas[len(lineas)-n:]
	}
	return "    " + strings.Join(lineas, "\n    ")
}
