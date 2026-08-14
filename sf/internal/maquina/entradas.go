package maquina

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/historia"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
)

// Nueva es `sf new`: mete una entrada B o C al backlog.
//
// ────────────────────────────────────────────────────────────────────────────
// EL BACKLOG ES EL EMBUDO, Y ESTO NO SE HABÍA TRAZADO NUNCA
// ────────────────────────────────────────────────────────────────────────────
//
// Los cinco estados de producto corren UNA VEZ, y sólo en la entrada A (producto
// nuevo). Las otras dos —feature sobre lo existente, y bug— llegan a un producto
// que ya tiene brief, PRD, constitución y roadmap. ¿Por dónde entran?
//
// La respuesta ya estaba escrita sin saber que era esto (session.md §3):
//
//	las tres entradas convergen en el backlog: EL BACKLOG ES EL EMBUDO.
//
// `sf new` es al backlog lo que `sf take` es al roadmap: el que mete algo en la
// cola, contra el que saca.
//
// ────────────────────────────────────────────────────────────────────────────
// Y NO LLEVA BANDERA PARA EL BUG (H20)
// ────────────────────────────────────────────────────────────────────────────
//
// El us-# ya tiene `tipo: us | bug` en el frontmatter, lo escribe el que
// pinponeó, y sf lo lee para rutear. Un dato que ya existía y no tenía
// consumidor, ahora lo tiene:
//
//	tipo: us    →  planificacion → implementar → revision → cierre
//	tipo: bug   →  implementar → cierre        (se saltean dos estados)
//
// No hay carril paralelo ni segunda máquina. Hay un campo que saltea estados —
// y por eso el rastro no se pierde: el bug igual entró por el backlog.
func Nueva(raiz string, e *estado.Estado, texto string) Efecto {
	var ef Efecto

	// Sin constitución no hay producto: `sf new` es para agregarle algo a uno
	// que ya existe. En un producto nuevo la entrada es el brief.
	if !e.Producto.ConstitucionSellada {
		ef.falla("este producto todavía no está armado: seguí con `sf next` desde el brief")
		return ef
	}

	id := historia.ProximoID(raiz)
	ruta := filepath.Join(raiz, docs.Historia(id))

	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		ef.falla("%v", err)
		return ef
	}
	// O_EXCL hace que crear falle si el archivo ya existe, en vez de pisarlo.
	// No debería pasar —el id sale del máximo+1— pero si pasara, pisar una
	// historia sería perder trabajo en silencio.
	f, err := os.OpenFile(ruta, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		ef.falla("no pude crear %s: %v", docs.Historia(id), err)
		return ef
	}
	defer f.Close()

	cuerpo := historia.Esqueleto(id, e.Producto.PrdHash)
	if texto != "" {
		// Lo que Javier tipeó va al Contexto, no al título ni a los criterios:
		// esos salen del pinponeo. Perderlo sería hacerle repetir lo que ya dijo.
		cuerpo += "\n" + texto + "\n"
	}
	if _, err := f.WriteString(cuerpo); err != nil {
		ef.falla("%v", err)
		return ef
	}

	// Volver a abrir la ⏸ del ⑨ es lo que hace que `sf next` mande al pinponeo
	// en vez de seguir con el ciclo de feature.
	e.Producto.BacklogVisto = false

	ef.Mensaje = fmt.Sprintf("%s creada. Ahora se pinponea: `sf next` te lleva.", id)
	return ef
}

// huerfanas son las historias que no están en ninguna feature del roadmap.
//
// Aparece por `sf new` y no por el flujo normal: cuando agregás una historia a
// un producto que ya tiene roadmap, esa historia no está en ninguna feature — y
// sin esto, `sf next` seguiría con el ciclo como si nada y la historia nueva no
// la implementaría nadie.
//
// Es el mismo chequeo que ya hace la compuerta del ⑩; acá se usa para decidir a
// dónde mandar, no para frenar.
func huerfanas(raiz string, r *roadmap.Roadmap) []string {
	if r == nil {
		return nil
	}
	enFeatures := map[string]bool{}
	for _, f := range r.Features {
		for _, id := range f.Historias {
			enFeatures[id] = true
		}
	}

	var sueltas []string
	for _, id := range historia.Ids(raiz) {
		if !enFeatures[id] {
			sueltas = append(sueltas, id)
		}
	}
	return sueltas
}

// esDeBugs dice si TODAS las historias de una feature son bugs.
//
// Se exige que sean todas y no "alguna", porque saltear la planificación de una
// feature que mezcla un bug con dos historias nuevas sería dejar esas dos sin
// diseño. Si están mezcladas, el agrupamiento del ⑩ ya estaba mal: un bug y una
// feature nueva no comparten solución técnica (artefactos.md §3).
func esDeBugs(raiz string, f roadmap.Feature) bool {
	if len(f.Historias) == 0 {
		return false
	}
	for _, id := range f.Historias {
		h, err := historia.Leer(raiz, id)
		if err != nil || !h.EsBug() {
			return false
		}
	}
	return true
}
