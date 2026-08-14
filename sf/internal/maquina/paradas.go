package maquina

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdbaigorria/specforge/sf/internal/constitucion"
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/frontmatter"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/revision"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
)

// ────────────────────────────────────────────────────────────────────────────
// LAS PARADAS SON DE JAVIER, Y POR ESO SON COMANDOS SEPARADOS
// ────────────────────────────────────────────────────────────────────────────
//
// Los tres comandos de este archivo tienen algo en común que los distingue de
// `done` y `context`: **los corre el orquestador, nunca el subagente**. Son
// cómo Javier le contesta a una parada.
//
// Y los seis comandos que aparecieron en la ronda de la superficie —approve,
// reject, take, new, model, dismiss— tienen todos esa forma. No salieron de
// mirar los artefactos: salieron de recorrer el bucle y preguntarse qué pasa
// cuando sf dice 🛑.

// Efecto es lo que pasó al contestarle a una parada.
type Efecto struct {
	Mensaje string
	Fallas  []string
}

func (e Efecto) Pasa() bool { return len(e.Fallas) == 0 }

func (e *Efecto) falla(formato string, args ...any) {
	e.Fallas = append(e.Fallas, fmt.Sprintf(formato, args...))
}

// Texto arma la respuesta para el orquestador.
func (e Efecto) Texto() string {
	if e.Pasa() {
		return "✓ " + e.Mensaje + "\n"
	}
	s := "✗ No pude.\n"
	for _, f := range e.Fallas {
		s += "  · " + f + "\n"
	}
	return s
}

// ────────────────────────────────────────────────────────────────────────────
// sf approve
// ────────────────────────────────────────────────────────────────────────────

// Aprobar sella la parada en la que estés — y hay cinco.
//
// ────────────────────────────────────────────────────────────────────────────
// UN COMANDO, CINCO SIGNIFICADOS, NINGUNO AMBIGUO (H5 · H16)
// ────────────────────────────────────────────────────────────────────────────
//
//	brief          → guarda el veredicto que Javier escribió en el archivo
//	constitucion   → constitucion_sellada = true
//	backlog (⏸)    → backlog_visto = true
//	planificacion  → la feature pasa a implementar          (el ⑰)
//	cierre (⏸)     → ARCHIVA: mueve la carpeta, mergea, borra la branch
//
// Las cinco son el mismo hecho —Javier aprueba lo que se produjo— y lo que
// cambia es qué se sella. Eso sf lo deduce del estado en el que está: una sola
// respuesta correcta, R1.
func Aprobar(raiz string, e *estado.Estado, r *roadmap.Roadmap) Efecto {
	var ef Efecto

	switch {
	case e.Producto.BriefSellado == "":
		// El veredicto NO lo elige sf: lo escribió Javier en el brief, y acá
		// sólo se copia al estado. `sf approve` significa "sí, sellalo con lo
		// que dice" — incluso si lo que dice es "no-lo-hagas".
		var fm struct {
			Veredicto string `yaml:"veredicto"`
		}
		if _, err := frontmatter.DeArchivo(filepath.Join(raiz, docs.Brief), &fm); err != nil {
			ef.falla("no pude leer el brief: %v", err)
			return ef
		}
		if fm.Veredicto == "" {
			ef.falla("el brief no trae veredicto: no hay qué sellar")
			return ef
		}
		e.Producto.BriefSellado = fm.Veredicto
		ef.Mensaje = "brief sellado: " + fm.Veredicto
		e.Producto.Rechazo = ""
		return ef

	case e.Producto.PrdHash == "":
		ef.falla("el ⑦ no tiene parada: corré `sf done`")
		return ef

	case !e.Producto.ConstitucionSellada:
		e.Producto.ConstitucionSellada = true
		e.Producto.Rechazo = ""
		ef.Mensaje = "constitución sellada"
		return ef

	case !e.Producto.BacklogVisto:
		e.Producto.BacklogVisto = true
		ef.Mensaje = "backlog visto — sigue el ⑩"
		return ef
	}

	// De acá para abajo es el ciclo de feature.
	f, hay := e.Actual()
	if !hay {
		ef.falla("no hay ninguna parada que aprobar")
		return ef
	}
	fr, enRoadmap := r.Buscar(e.FeatureActual)
	if !enRoadmap {
		ef.falla("%s no está en el roadmap", e.FeatureActual)
		return ef
	}

	switch f.Estado {
	case estado.Planificacion:
		// El ⑰. La puerta "otra feature" NO se elige acá: se elige con
		// `sf take f-3` después de aprobar. Cada comando hace un solo trabajo.
		f.Estado = estado.Implementar
		f.Rechazo = ""
		ef.Mensaje = fmt.Sprintf("plan de %s aprobado — a implementar", fr.ID)
		return ef

	case estado.Cierre:
		return archivar(raiz, e, f, fr)
	}

	ef.falla("%s está en %q y ahí no hay nada que aprobar", fr.ID, f.Estado)
	return ef
}

// archivar es lo que pasa DESPUÉS de la ⏸ del ㉓.
//
// `sf feature archive` no existe como comando: es esto (H16). Lo que destraba
// una parada ya tiene nombre, y el archivado cae después de la parada, no antes.
//
// Son cuatro cosas mecánicas, y ninguna necesita criterio:
//
//	mover la carpeta entera a .docs/archivado/
//	merge a la branch base, con el modo que diga la constitución
//	borrar la branch de la feature
//	marcar la feature como cerrada
func archivar(raiz string, e *estado.Estado, f *estado.Feature, fr roadmap.Feature) Efecto {
	var ef Efecto

	origen := filepath.Join(raiz, fr.Carpeta())
	destino := filepath.Join(raiz, docs.Archivado, filepath.Base(fr.Carpeta()))

	if err := os.MkdirAll(filepath.Dir(destino), 0o755); err != nil {
		ef.falla("%v", err)
		return ef
	}
	// Se mueve la carpeta ENTERA con todo adentro —decision, spec, tareas,
	// revision, doc, journal—. Las referencias por id siguen funcionando porque
	// sf es el que resuelve dónde vive cada cosa (regla 1.4).
	if err := os.Rename(origen, destino); err != nil {
		ef.falla("no pude archivar la carpeta: %v", err)
		return ef
	}

	// El git es lo único que puede fallar de verdad acá, y si falla NO se
	// deshace el movimiento: la carpeta archivada es correcta igual, y el merge
	// se puede reintentar a mano. Deshacer sería peor — dejaría el repo a medias
	// sin que nadie sepa en qué mitad quedó.
	if c, err := constitucion.Leer(raiz); err == nil && git.EsRepo(raiz) {
		branch := c.Branch(fr.ID, fr.Slug)
		base := c.Git.Base()

		if git.Existe(raiz, branch) {
			if err := git.Checkout(raiz, base); err != nil {
				ef.falla("no pude pararme en %s: %v", base, err)
			} else if err := git.Merge(raiz, branch, c.Git.Merge); err != nil {
				ef.falla("el merge falló: %v", err)
			} else if err := git.BorrarBranch(raiz, branch); err != nil {
				// Que no se pueda borrar no invalida el merge: se avisa y sigue.
				ef.Mensaje = "merge hecho, pero la branch no se borró: " + err.Error()
			}
		}
	}

	f.Estado = estado.Cerrada
	e.FeatureActual = ""

	if ef.Mensaje == "" {
		ef.Mensaje = fmt.Sprintf("%s archivada y cerrada", fr.ID)
	}
	return ef
}

// ────────────────────────────────────────────────────────────────────────────
// sf reject
// ────────────────────────────────────────────────────────────────────────────

// Rechazar no sella, y guarda el motivo.
//
// ────────────────────────────────────────────────────────────────────────────
// EL MOTIVO NO ES DECORACIÓN: ES LO ÚNICO QUE VIAJA
// ────────────────────────────────────────────────────────────────────────────
//
// El orquestador va a lanzar un subagente nuevo para rehacer el trabajo, y ese
// subagente arranca DE CERO. Sin el motivo no sabe qué estuvo mal y vuelve a
// proponer lo mismo.
//
// Por eso se guarda en el estado y no se imprime y ya: `sf context` lo mete en
// el sobre, que es exactamente donde el que rehace lo va a leer.
func Rechazar(e *estado.Estado, motivo string) Efecto {
	var ef Efecto
	if motivo == "" {
		ef.falla("falta el motivo: `sf reject \"por qué\"`. Sin él, el que rehaga vuelve a proponer lo mismo.")
		return ef
	}

	switch {
	case e.Producto.BriefSellado == "":
		e.Producto.Rechazo = motivo
		ef.Mensaje = "brief rechazado — se rehace el ①–⑤"
	case !e.Producto.ConstitucionSellada:
		e.Producto.Rechazo = motivo
		ef.Mensaje = "constitución rechazada — se rehace el ⑧"
	default:
		f, hay := e.Actual()
		if !hay || f.Estado != estado.Planificacion {
			ef.falla("no hay ninguna parada que rechazar")
			return ef
		}
		// "Pido cambios" vuelve al bloque ⑫–⑯ ENTERO, no a un pedazo. Nadie
		// entra ni sale por el medio (maquina-estados.md §1), así que no hay
		// que borrar archivos: el que rehaga los pisa.
		f.Rechazo = motivo
		ef.Mensaje = fmt.Sprintf("plan de %s rechazado — vuelve al bloque ⑫–⑯", e.FeatureActual)
	}
	return ef
}

// ────────────────────────────────────────────────────────────────────────────
// sf take
// ────────────────────────────────────────────────────────────────────────────

// Tomar es el ⑪: elige de la cola.
//
// La máquina lo había declarado "no es un estado, es la transición que entra a
// planificacion" — y una transición que Javier elige necesita un nombre para
// ser invocada. Acá lo consigue.
//
// Sirve en tres momentos distintos, y por eso no es una bandera de otro
// comando: elegir la primera feature de todas, elegir "otra feature" en el ⑰
// después de aprobar, y cambiar de idea a mitad de la cola.
func Tomar(e *estado.Estado, r *roadmap.Roadmap, id string) Efecto {
	var ef Efecto

	if r == nil {
		ef.falla("todavía no hay roadmap: falta el ⑩")
		return ef
	}
	if _, hay := r.Buscar(id); !hay {
		ef.falla("%s no está en el roadmap", id)
		return ef
	}

	if f, hay := e.Features[id]; hay && f.Estado == estado.Cerrada {
		ef.falla("%s ya está cerrada", id)
		return ef
	}

	// Dejar algo a medias avisa, no frena. El trabajo no se pierde: el estado
	// de la feature anterior queda guardado en `features` y volver con
	// `sf take` la retoma donde estaba. Frenar sería sf decidiendo por Javier.
	if anterior, hay := e.Actual(); hay && e.FeatureActual != id {
		enMarcha := anterior.Estado == estado.Implementar ||
			anterior.Estado == estado.Revision ||
			anterior.Estado == estado.Cierre
		if enMarcha {
			ef.Mensaje = fmt.Sprintf("⚠ dejás %s en %q. Volvés con `sf take %s`.\n",
				e.FeatureActual, anterior.Estado, e.FeatureActual)
		}
	}

	e.FeatureActual = id

	f, existia := e.Features[id]
	if !existia {
		// Primera vez: entra por planificación, que es el principio del ciclo.
		e.Features[id] = &estado.Feature{Estado: estado.Planificacion}
		ef.Mensaje += fmt.Sprintf("%s tomada — a planificar", id)
		return ef
	}

	if f.Estado == estado.Planificada {
		// El plan ya está aprobado y esperando: retomarla es implementar. Es la
		// vuelta de la puerta "otra feature" del ⑰.
		f.Estado = estado.Implementar
		ef.Mensaje += fmt.Sprintf("%s retomada — el plan ya estaba aprobado, a implementar", id)
		return ef
	}

	ef.Mensaje += fmt.Sprintf("%s retomada en %q", id, f.Estado)
	return ef
}

// ────────────────────────────────────────────────────────────────────────────
// sf model y sf dismiss — las dos salidas de ME TRABÉ
// ────────────────────────────────────────────────────────────────────────────

// Modelo sube (o baja) el modelo de la feature actual.
//
// Necesita comando propio porque no sale de ningún archivo: `maquina-estados.md`
// §9 lo marcó como uno de los dos campos indeducibles — "es el que se está
// usando AHORA, no el recomendado; cuando Javier lo sube en el ⑳, esa decisión
// no queda registrada en ningún lado". Este comando es ese registro.
//
// Y resetea el contador: cambiar de modelo es empezar de nuevo, no seguir
// acumulando los fracasos del anterior.
func Modelo(e *estado.Estado, nombre string) Efecto {
	var ef Efecto
	if nombre == "" {
		ef.falla("falta el modelo: `sf model <nombre>`")
		return ef
	}
	f, hay := e.Actual()
	if !hay {
		ef.falla("no hay feature en curso")
		return ef
	}

	anterior := f.Modelo
	f.Modelo = nombre
	f.IntentosFallidos = 0

	if anterior == "" {
		ef.Mensaje = "modelo: " + nombre
	} else {
		ef.Mensaje = fmt.Sprintf("modelo: %s → %s (contador reseteado)", anterior, nombre)
	}
	return ef
}

// Descartar marca un hallazgo como decidido por Javier.
//
// Es la tercera salida de ME TRABÉ y tapa un bucle infinito real: salir de
// `revision` es automático, así que un FALSO POSITIVO no se cierra nunca — el
// revisor lo encuentra, el implementador no lo puede arreglar porque no está
// roto, y vuelta.
//
// `descartado` es el único estado de hallazgo que no se deduce de nada (H14):
// sin persistirlo, la vuelta siguiente lo vuelve a encontrar porque el código
// sigue igual.
func Descartar(raiz string, e *estado.Estado, r *roadmap.Roadmap, id, motivo string) Efecto {
	var ef Efecto
	if id == "" {
		ef.falla("falta el hallazgo: `sf dismiss <h-#> \"motivo\"`")
		return ef
	}
	if motivo == "" {
		ef.falla("falta el motivo. Descartar sin decir por qué es perder la razón.")
		return ef
	}
	if r == nil {
		ef.falla("todavía no hay roadmap")
		return ef
	}
	fr, hay := r.Buscar(e.FeatureActual)
	if !hay {
		ef.falla("no hay feature en curso")
		return ef
	}

	ruta := filepath.Join(raiz, fr.Carpeta(), docs.Revision)
	rev, err := revision.Leer(ruta)
	if err != nil {
		ef.falla("%v", err)
		return ef
	}

	h, encontrado := rev.Buscar(id)
	if !encontrado {
		ef.falla("%s no está en la revisión de %s", id, fr.ID)
		return ef
	}

	h.Estado = revision.Descartado
	h.Motivo = motivo

	if err := rev.Guardar(ruta); err != nil {
		ef.falla("%v", err)
		return ef
	}

	// El contador se resetea igual que con `sf model`: el bucle se destrabó, y
	// lo que venía fallando ya no va a fallar por esto.
	if f, hayF := e.Actual(); hayF {
		f.IntentosFallidos = 0
	}

	ef.Mensaje = fmt.Sprintf("%s descartado: %s", id, motivo)
	return ef
}
