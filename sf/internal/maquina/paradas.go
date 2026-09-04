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
	"github.com/jdbaigorria/specforge/sf/internal/frontmatter"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/global"
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

	// Avisos son cosas que hay que saber y que NO frenan.
	//
	// Vienen de las compuertas que `approve` corre antes de sellar, y la
	// distinción es la misma que en `compuerta.Resultado`: sf frena sobre
	// hechos y avisa sobre todo lo demás, porque la última palabra es de
	// Javier. Tragárselos acá sería silenciarlos justo cuando está mirando.
	Avisos []string

	// Global es que además del estado del proyecto hay que guardar
	// ~/.specforge/. Sólo lo enciende `sf model` al declarar uno nuevo.
	//
	// Es un campo y no "guardá siempre las dos cosas" porque escribir el mapa
	// global en cada `sf approve` sería tocar el home de Javier veinte veces
	// por feature para no cambiar nada.
	Global bool

	// Portamodelo es que se declaró un alias NUEVO y hay que regenerar los
	// archivos de agente del harness.
	//
	// Está separado de Global porque redeclarar un alias que ya existía toca el
	// catálogo pero no agrega ningún archivo — y avisar "reiniciá el harness"
	// cuando no hace falta enseña a ignorar el aviso.
	Portamodelo bool

	// Cambio es que hay que guardar el estado AUNQUE el comando haya fallado.
	//
	// ────────────────────────────────────────────────────────────────────
	// LA REGLA NORMAL ES NO GUARDAR, Y ESTÁ BIEN — SALVO EN UN CASO
	// ────────────────────────────────────────────────────────────────────
	//
	// Un `sf take f-99` que falla no tiene que dejar rastro, y por eso el
	// default es guardar sólo si pasó. Pero `archivar` toca el DISCO —mueve
	// una carpeta, mergea, borra una branch— y si falla después de eso, no
	// guardar el estado deja al estado.json describiendo un repo que ya no
	// existe. Ahí el comando falló y el mundo cambió igual, y la única
	// respuesta correcta es anotar lo que sí pasó.
	//
	// Es el mismo campo que ya tenía `Cierre` (done.go) por la misma razón:
	// un `sf done` que falla igual incrementa `intentos_fallidos`.
	Cambio bool

	// Estado es en qué paso se contestó la parada, y existe para el registro.
	//
	// Mismo motivo que el de `Cierre`: un `sf approve` sobre el ⑥ y uno sobre
	// el ⑧ son la misma línea vistos desde afuera, y cuál era lo sabe este
	// switch y nadie más.
	Estado string

	// Commit es "cuando el estado esté guardado, commiteá con este mensaje".
	//
	// ────────────────────────────────────────────────────────────────────
	// POR QUÉ NO LO COMMITEA EL QUE LO PIDE
	// ────────────────────────────────────────────────────────────────────
	//
	// Archivar mueve una carpeta entera, y eso tiene que quedar en un commit
	// suyo. Pero el estado.json lo escribe main DESPUÉS de que esto vuelve,
	// así que un commit hecho acá adentro dejaría el estado afuera.
	//
	// Y dejarlo sin commitear no es neutral: el `git add -A` del primer lote
	// de la feature SIGUIENTE se lo lleva puesto, y ahí el archivado de f-1
	// termina adentro del commit de f-2. Eso es exactamente el dolor #3
	// —commits mal agrupados— colándose por el único lugar donde sf mueve
	// archivos sin cerrar un lote.
	Commit string
}

func (e Efecto) Pasa() bool { return len(e.Fallas) == 0 }

func (e *Efecto) falla(formato string, args ...any) {
	e.Fallas = append(e.Fallas, fmt.Sprintf(formato, args...))
}

// compuerta corre una compuerta antes de sellar, y dice si se puede seguir.
//
// Los avisos NO frenan y viajan igual: son lo que hay que saber y no lo que
// impide avanzar, y perderlos acá sería silenciarlos justo cuando Javier está
// mirando.
func (e *Efecto) compuerta(r compuerta.Resultado) bool {
	e.Avisos = append(e.Avisos, r.Avisos...)
	if r.Pasa() {
		return true
	}
	e.Fallas = append(e.Fallas, r.Fallas...)
	return false
}

// Texto arma la respuesta para el orquestador.
//
// Los avisos van primero y en los dos casos —salga bien o mal—: son cosas que
// hay que saber, y esconderlos detrás de un ✓ es la forma más fácil de que
// nadie los lea.
func (e Efecto) Texto() string {
	var b strings.Builder
	for _, a := range e.Avisos {
		fmt.Fprintf(&b, "⚠ %s\n", a)
	}
	if e.Pasa() {
		fmt.Fprintf(&b, "✓ %s\n", e.Mensaje)
		return b.String()
	}
	b.WriteString("✗ No pude.\n")
	for _, f := range e.Fallas {
		fmt.Fprintf(&b, "  · %s\n", f)
	}
	return b.String()
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
//
// ────────────────────────────────────────────────────────────────────────────
// Y CORRE LA COMPUERTA ANTES DE SELLAR, QUE NO ES LO MISMO QUE SER `done`
// ────────────────────────────────────────────────────────────────────────────
//
// Sellaba con una asignación directa, sin comprobar nada. O sea que la regla
// dura del producto —"el estado avanza con hechos comprobados"— valía para
// `sf done` y no para el otro comando que mueve el estado:
//
//	sf approve   →  ✓ constitución sellada        con test_cmd vacío
//	sf approve   →  ✓ backlog visto — sigue el ⑩  con historias sin criterios
//
// Lo segundo es lo grave: sin ids de criterio, la cobertura del ⑰ cuenta cero
// contra cero y pasa, y el conteo de veredictos del ㉑ también. Un backlog
// sellado sin ids desarma el mecanismo entero.
//
// Esto NO convierte a `approve` en `done`. `done` mueve cuando la compuerta
// pasa; `approve` sella lo que Javier decidió, y lo único que cambia es que ya
// no puede sellar algo que la máquina sabe que está roto. La decisión sigue
// siendo suya; deja de poder ser una decisión sobre un artefacto inválido.
//
// El `brief` es la excepción, y no por olvido: su compuerta es "trae uno de los
// tres veredictos", y acá se lee el valor concreto que se va a sellar — que es
// el mismo chequeo, hecho mejor.
func Aprobar(raiz string, e *estado.Estado, r *roadmap.Roadmap) Efecto {
	var ef Efecto

	switch {
	case e.Producto.BriefSellado == "":
		ef.Estado = "brief"
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
		ef.Estado = "prd"
		ef.falla("el ⑦ no tiene parada: corré `sf done`")
		return ef

	case !e.Producto.ConstitucionSellada:
		ef.Estado = "constitucion"
		if !ef.compuerta(compuerta.Constitucion(raiz)) {
			return ef
		}
		e.Producto.ConstitucionSellada = true
		e.Producto.Rechazo = ""
		ef.Mensaje = "constitución sellada"
		return ef

	case !e.Producto.BacklogVisto:
		ef.Estado = "backlog"
		if !ef.compuerta(compuerta.Backlog(raiz)) {
			return ef
		}
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

	ef.Estado = f.Estado

	switch f.Estado {
	case estado.Planificacion:
		// El ⑰. La puerta "otra feature" NO se elige acá: se elige con
		// `sf take f-3` después de aprobar. Cada comando hace un solo trabajo.
		//
		// `sf next` ya corre estas cinco antes de ofrecer el ⑰, así que en el
		// bucle normal esto no cambia nada. Cambia para el `sf approve` tipeado
		// directo, que las salteaba: aprobar un plan que cubre 7 de 9 criterios
		// es exactamente lo que el ⑰ existe para impedir.
		if !ef.compuerta(compuerta.Planificacion(raiz, fr)) {
			return ef
		}
		f.Estado = estado.Implementar
		f.Rechazo = ""
		ef.Mensaje = fmt.Sprintf("plan de %s aprobado — a implementar", fr.ID)
		return ef

	case estado.Cierre:
		// Archivar es irreversible —mueve la carpeta, mergea, borra la branch—,
		// así que la compuerta va antes: archivar una feature sin doc y sin
		// journal deja la carpeta en `.docs/archivado/` sin lo único que alguien
		// va a leer seis meses después, y de ahí no se vuelve.
		if !ef.compuerta(compuerta.Cierre(raiz, fr)) {
			return ef
		}
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
//	merge a la branch base, con el modo que diga la constitución
//	borrar la branch de la feature
//	mover la carpeta entera a .docs/archivado/
//	marcar la feature como cerrada
//
// ────────────────────────────────────────────────────────────────────────────
// EL ORDEN NO ES EL DEL DISEÑO, Y ES A PROPÓSITO
// ────────────────────────────────────────────────────────────────────────────
//
// El diseño las lista con el movimiento primero, y así estaba escrito. Pero el
// movimiento es lo ÚNICO irreversible de los cuatro, y el git es lo único que
// puede fallar — o sea que estaban exactamente al revés:
//
//	mover · fallar el checkout   →  carpeta archivada, feature sin cerrar,
//	                                branch sin mergear, y `sf next` mandando al
//	                                ㉓ sobre archivos que ya no están ahí
//
// De ese estado no se sale: el segundo `sf approve` falla en el `rename` porque
// el origen ya no existe. Había que arreglarlo a mano.
//
// Con git primero, un merge que falla no movió nada y `sf approve` se reintenta
// tal cual. Lo reversible antes que lo irreversible.
func archivar(raiz string, e *estado.Estado, f *estado.Feature, fr roadmap.Feature) Efecto {
	var ef Efecto

	// ① El estado.json, que viene sucio POR CONSTRUCCIÓN.
	//
	// sf lo escribe en cada transición y sólo lo commitea al cerrar un lote, así
	// que entre el último lote y el ㉓ hay al menos dos escrituras sin commit —
	// y `git checkout` aborta si tiene que pisar un archivo modificado. Esto no
	// fallaba a veces: fallaba SIEMPRE.
	//
	// Y commitearlo no es un truco para destrabar el checkout: es trabajo real
	// que si no queda huérfano. El estado se versiona con el repo justamente
	// para que viaje (regla dura 3).
	if git.EsRepo(raiz) {
		sucio, err := git.Sucio(raiz)
		if err != nil {
			ef.falla("no pude ver si hay cambios sin commitear: %v", err)
			return ef
		}
		if sucio {
			if _, err := git.Commit(raiz, "chore: cierre de "+fr.ID); err != nil {
				ef.falla("no pude commitear lo que quedaba antes de archivar: %v", err)
				return ef
			}
		}
	}

	// ② El git: mergear y borrar la branch. Es lo único que puede fallar de
	// verdad, así que va antes de tocar el disco.
	if c, err := constitucion.Leer(raiz); err == nil && git.EsRepo(raiz) {
		branch := c.Branch(fr.ID, fr.Slug)
		base := c.Git.Base()

		if git.Existe(raiz, branch) {
			if err := git.Checkout(raiz, base); err != nil {
				ef.falla("no pude pararme en %s: %v", base, err)
				return ef
			}
			if err := git.Merge(raiz, branch, c.Git.Merge); err != nil {
				ef.falla("el merge falló: %v", err)
				return ef
			}
			if err := git.BorrarBranch(raiz, branch); err != nil {
				// Que no se pueda borrar no invalida el merge: se avisa y sigue.
				// Es el único de los cuatro pasos que puede quedar a medias sin
				// dejar nada inconsistente.
				ef.Mensaje = "merge hecho, pero la branch no se borró: " + err.Error()
			}
		}
	}

	// ③ La carpeta. A partir de acá el disco cambió, así que todo lo que siga
	// enciende `Cambio`: el comando puede fallar y el estado igual tiene que
	// anotar lo que ya pasó.
	if err := archivarCarpeta(raiz, fr); err != nil {
		ef.falla("%v", err)
		return ef
	}

	// ④ El sello, y el commit del archivado — que lo hace main, cuando el
	// estado.json ya esté escrito y pueda entrar en el mismo commit.
	f.Estado = estado.Cerrada
	e.FeatureActual = ""
	ef.Cambio = true
	if git.EsRepo(raiz) {
		ef.Commit = "chore: " + fr.ID + " archivada"
	}

	if ef.Mensaje == "" {
		ef.Mensaje = fmt.Sprintf("%s archivada y cerrada", fr.ID)
	}
	return ef
}

// archivarCarpeta mueve la carpeta de la feature, y es idempotente.
//
// Se mueve ENTERA con todo adentro —decision, spec, tareas, revision, doc,
// journal—. Las referencias por id siguen funcionando porque sf es el que
// resuelve dónde vive cada cosa (regla 1.4).
//
// Lo de idempotente no es elegancia: es lo que destraba un repo que quedó a
// medias con la versión vieja, donde el movimiento pasó y el merge falló. Si el
// destino ya está y el origen no, el trabajo ya se hizo — decirlo "no existe el
// archivo" sería frenar por algo que está bien.
func archivarCarpeta(raiz string, fr roadmap.Feature) error {
	origen := filepath.Join(raiz, fr.Carpeta())
	destino := filepath.Join(raiz, docs.Archivado, filepath.Base(fr.Carpeta()))

	if _, err := os.Stat(destino); err == nil {
		if _, err := os.Stat(origen); os.IsNotExist(err) {
			return nil // ya estaba archivada
		}
		return fmt.Errorf("%s existe en las dos: archivada y sin archivar. Mirá cuál querés",
			filepath.Base(fr.Carpeta()))
	}

	if err := os.MkdirAll(filepath.Dir(destino), 0o755); err != nil {
		return err
	}
	if err := os.Rename(origen, destino); err != nil {
		return fmt.Errorf("no pude archivar la carpeta: %w", err)
	}
	return nil
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

// capacidadDe deja el campo vacío a propósito.
//
// `capacidad` es opinión de Javier y sf no la puede inventar: inventarla sería
// exactamente opinar sobre qué modelo se parece a cuál (R3). Se declara en cero,
// que en el YAML se omite, y Javier la escribe a mano el día que quiera que el ⑯
// la lea. Existe como función y no como un `0` suelto para que quede dicho POR
// QUÉ está vacío, y nadie lo "arregle" más adelante.
func capacidadDe(string) int { return 0 }

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
//
// ────────────────────────────────────────────────────────────────────────────
// APROBAR ES DECLARAR, Y POR ESO EL MAPA SE LLENA SOLO
// ────────────────────────────────────────────────────────────────────────────
//
// Este comando es también la respuesta a la 🛑 de "ese modelo no está
// declarado". Cuando trae `--via`, además de usar el modelo LO DECLARA en
// ~/.specforge/, y queda para todos los proyectos.
//
//	sf model deepseek --via consola --comando "deepseek exec"
//
// No hay un `sf model add` aparte a propósito: separar "declarar" de "usar"
// crearía un estado intermedio —declarado pero nunca usado— que no le sirve a
// nadie. Es el mismo mecanismo que `dependencias_aprobadas`: la lista no se
// escribe de antemano, se construye con cada aprobación.
// Declaracion son los datos de `sf model`, tal como los tipeó Javier.
//
// Es un struct y no seis parámetros porque son todos strings: en fila, invertir
// dos es un error que el compilador no puede ver.
type Declaracion struct {
	Nombre, Alias, ID, Via, Comando, Esfuerzo string
}

func Modelo(e *estado.Estado, g *global.Config, d Declaracion) Efecto {
	nombre, alias, id, via, comando := d.Nombre, d.Alias, d.ID, d.Via, d.Comando
	var ef Efecto
	if nombre == "" {
		ef.falla("falta qué: `sf model <perfil|alias>`")
		return ef
	}

	// La validación va ANTES de tocar nada: un `--via consola` sin comando
	// dejaría un modelo declarado que nadie puede invocar, y eso se descubriría
	// recién cuando el orquestador lo intente.
	if via != "" {
		if via != global.Subagente && via != global.Consola {
			ef.falla("`--via %s` no existe. Es `subagente` o `consola`.", via)
			return ef
		}
		if via == global.Consola && comando == "" {
			ef.falla("`--via consola` necesita `--comando \"…\"`: sin eso nadie sabe cómo invocarlo.")
			return ef
		}
		if via == global.Subagente && comando != "" {
			ef.falla("`--comando` sólo tiene sentido con `--via consola`.")
			return ef
		}
	}

	// ── ① declarar un modelo EN UN PERFIL ────────────────────────────────
	//
	// Es la respuesta a la 🛑, y NO exige que haya una feature en curso: la
	// parada puede saltar en cualquiera de los cinco estados de producto, donde
	// no hay ninguna. Exigirla ahí dejaría la 🛑 sin salida.
	if global.EsPerfil(nombre) {
		if id == "" {
			ef.falla("declarar un perfil necesita `--id <el-id-de-tu-harness>`.")
			return ef
		}
		if alias == "" {
			ef.falla("declarar un modelo necesita `--alias <corto>`: es el nombre que va a escribir el ⑯ en tareas.json, y no puede ser el id.")
			return ef
		}
		if g == nil {
			ef.falla("no hay catálogo todavía: corré `sf install`.")
			return ef
		}
		if via == "" {
			via = global.Subagente
		}
		nuevo, err := g.Declarar(nombre, global.Modelo{
			Alias: alias, ID: id, Via: via, Comando: comando,
			Esfuerzo: d.Esfuerzo, Capacidad: capacidadDe(nombre), Para: "",
		})
		if err != nil {
			ef.falla("%s", err)
			return ef
		}
		ef.Global = true
		ef.Portamodelo = nuevo
		ef.Mensaje += fmt.Sprintf("%q declarado en %s · perfil %s", alias, g.EnUso(), nombre)
		if nuevo && global.NecesitaPortamodelo(g.EnUso()) {
			// Medido: los agentes se leen al arrancar. Si esto no se dice acá, se
			// descubre fallando — que es la peor forma de enterarse de algo que
			// ya se sabía.
			ef.Avisos = append(ef.Avisos,
				"reiniciá tu harness: los agentes se leen al arrancar y esta sesión no va a ver el nuevo.")
		}
		return ef
	}

	// ── ② declarar un SUELTO ─────────────────────────────────────────────
	//
	// Un modelo que no compite en ningún perfil: el ajeno que sale por consola.
	// No genera portamodelo —no lo lanza el harness— y por eso no avisa nada.
	if via != "" && id != "" || via == global.Consola {
		if g == nil {
			ef.falla("no hay catálogo todavía: corré `sf install`.")
			return ef
		}
		if id == "" {
			id = nombre
		}
		g.DeclararSuelto(nombre, global.Modelo{ID: id, Via: via, Comando: comando, Esfuerzo: d.Esfuerzo})
		ef.Global = true
		ef.Mensaje += fmt.Sprintf("%q declarado como suelto", nombre)
		return ef
	}

	// ── ③ el ⑳: usar algo ya declarado en la feature en curso ────────────
	f, hay := e.Actual()
	if !hay {
		ef.falla("no hay feature en curso")
		return ef
	}

	// Tiene que estar declarado: cambiar a uno que sf no sabe invocar es cambiar
	// a nada, y el próximo `sf next` pararía igual. Decirlo acá ahorra esa vuelta.
	if g != nil {
		if _, declarado := g.Resolver(nombre); !declarado {
			ef.falla("%q no está declarado en %s. Declaralo primero: `sf model <perfil> --alias %s --id … --via subagente`.", nombre, g.EnUso(), nombre)
			return ef
		}
	}

	anterior := f.Modelo
	f.Modelo = nombre
	f.IntentosFallidos = 0

	// El "a → b" sólo si de verdad cambió. El caso `a → a` es real y frecuente:
	// el ⑯ recomendó un modelo, la 🛑 pidió declararlo, y `sf model` lo declara
	// sin cambiar nada. Imprimir "deepseek → deepseek" ahí es ruido que hace
	// dudar de si el comando hizo algo.
	switch {
	case anterior == "":
		ef.Mensaje = "modelo: " + nombre
	case anterior == nombre:
		ef.Mensaje = "modelo: " + nombre + " (contador reseteado)"
	default:
		ef.Mensaje = fmt.Sprintf("modelo: %s → %s (contador reseteado)", anterior, nombre)
	}
	if ef.Global {
		ef.Mensaje += fmt.Sprintf("\n   declarado en ~/.specforge/: via %s", via)
		if comando != "" {
			ef.Mensaje += " · " + comando
		}
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
