package maquina

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/compuerta"
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/revision"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
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
	if r == nil {
		// El ⑩ es el último de producto y su compuerta necesita el roadmap ya
		// escrito, así que se resuelve acá.
		res := compuerta.Roadmap(raiz)
		return Cierre{Resultado: res, Movio: res.Pasa(), Cambio: res.Pasa(),
			Mensaje: "el roadmap está listo — arranca el ciclo de feature"}
	}

	f, hay := e.Actual()
	if !hay {
		var c Cierre
		c.Fallas = []string{"no hay feature en curso: corré `sf take <feature>` primero"}
		return c
	}
	fr, enRoadmap := r.Buscar(e.FeatureActual)
	if !enRoadmap {
		var c Cierre
		c.Fallas = []string{fmt.Sprintf("%s no está en el roadmap", e.FeatureActual)}
		return c
	}

	switch f.Estado {
	case estado.Planificacion:
		return cerrarPlanificacion(raiz, f, fr)
	case estado.Implementar:
		return cerrarLote(raiz, f, msg)
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
func cerrarLote(raiz string, f *estado.Feature, msg string) Cierre {
	res := compuerta.Implementar(f)
	c := Cierre{Resultado: res}
	if !res.Pasa() {
		f.IntentosFallidos++
		c.Cambio = true
		return c
	}

	l, hayLote := f.LoteActual()
	if !hayLote {
		// Todos los lotes tienen commit: se cierra el estado entero.
		f.Estado = estado.Revision
		return Cierre{Movio: true, Cambio: true, Mensaje: "todos los lotes cerrados — sigue la revisión (㉑)"}
	}

	if strings.TrimSpace(msg) == "" {
		c.Fallas = append(c.Fallas, "falta el mensaje del commit: `sf done --msg \"…\"`. El lote no se cierra sin él.")
		return c
	}

	// ⚠ Falta correr los tests y exigir el verde, y falta comparar el hash
	// contra el que tomó `sf lote start`. Las dos son del paso 6, donde vive la
	// compuerta del rojo — sin ella no hay hash guardado contra qué comparar.
	// Mientras tanto, la compuerta de arriba ya impide llegar acá sin `rojo`.

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
			return Cierre{Movio: true, Cambio: true, Resultado: res,
				Mensaje: "hay hallazgos abiertos — vuelve a implementar"}
		}
		return c
	}

	f.Estado = estado.Cierre
	c.Movio, c.Cambio = true, true
	c.Mensaje = "revisión limpia — sigue el cierre (㉓)"
	return c
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
