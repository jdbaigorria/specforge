// Package docs son las rutas de los artefactos, y nada más.
//
// Existe por una razón sola: las mismas strings estaban empezando a aparecer en
// dos paquetes. `maquina` pregunta "¿existe .docs/brief.md?" para decidir si
// hay que trabajarlo, y `sobre` lo sirve. Con la ruta escrita en los dos, mover
// un archivo de lugar rompe uno y el otro sigue compilando.
//
// Es la misma regla que el proyecto entero aplica a los datos —un dato en dos
// lugares es un dato que se desincroniza— aplicada a las rutas.
package docs

import "path/filepath"

// Base es la carpeta donde vive todo lo que produce el flujo.
const Base = ".docs"

// Los artefactos de producto.
const (
	Brief        = Base + "/brief.md"
	PRD          = Base + "/prd.md"
	Constitucion = Base + "/constitucion.md"
	Backlog      = Base + "/backlog"   // los us-#.md, uno por historia
	Archivado    = Base + "/archivado" // las features cerradas, con todo adentro
)

// Los tres que deja el ①–⑤, y por qué son tres y no uno.
//
// ────────────────────────────────────────────────────────────────────────────
// UN ARCHIVO, DOS LECTORES
// ────────────────────────────────────────────────────────────────────────────
//
// Los tres son markdown con frontmatter, y eso NO es decoración:
//
//	el cuerpo         lo lee Javier            prosa, rondas, links
//	el frontmatter    lo cuenta la compuerta   números, sin parsear prosa
//
// Antes de esto la única forma de comprobar la evidencia era un `grep` de
// "http" sobre el brief. Funcionaba, y no se podía extender: "¿la entrevista
// terminó?" no se contesta con un grep.
//
// MEDIDO EL 2026-09-05, y es la razón de que Evidencia exista aparte del brief:
// dos modelos, misma idea semilla. Uno citó 13 links, el otro CERO, y los dos
// pasaron. Los números existían en la conversación y no llegaban a ningún lado
// donde alguien pudiera contarlos.
const (
	// Entrevista es el registro del ①–⑤: las rondas tal cual pasaron.
	//
	// Su frontmatter lleva `abiertas`, que es la única de las tres que FRENA:
	// una entrevista con ramas sin visitar es un brief construido sobre huecos.
	Entrevista = Base + "/entrevista.md"

	// Evidencia es lo que se encontró, con procedencia por afirmación.
	//
	// Se separó del brief a propósito: el brief es el ARGUMENTO y envejece con
	// cada vuelta del ⑥; la evidencia es el HECHO y se acumula.
	Evidencia = Base + "/evidencia.md"

	// Vocabulario es el glosario del negocio, y es el único PEREZOSO de los
	// tres: no existe hasta que una palabra se tambalea. Su ausencia no es una
	// falta — es que no hizo falta.
	//
	// No se llama CONTEXT.md aunque venga de ahí: `sf context` ya es el sobre,
	// y dos cosas distintas con el mismo nombre dentro de la misma familia es
	// el choque que ya pasó una vez con `Parte`.
	Vocabulario = Base + "/vocabulario.md"
)

// Los archivos de una feature, relativos a su carpeta.
const (
	Decision = "decision.md"    // ⑫ las 3 opciones — NO va en el sobre del ⑱
	Spec     = "spec-design.md" // ⑬ lo único que el implementador necesita
	Revision = "revision.json"  // ㉑㉒ criterios, mutantes y hallazgos
	Doc      = "doc.md"         // ㉓ la mitad técnica y la funcional
	Journal  = "journal.md"     // ㉓ las lecciones durables
)

// Historia arma la ruta de un us-# a partir de su id.
//
//	Historia("us-3")  →  .docs/backlog/us-3.md
//
// El roadmap y las tareas guardan ids, nunca rutas (regla 1.4: referenciar por
// id, nunca copiar). Esta función es la única que sabe traducir de uno a otro,
// que es lo que permite mover la carpeta sin tocar ningún JSON.
func Historia(id string) string {
	return filepath.Join(Backlog, id+".md")
}

// Journals devuelve el patrón glob de los journals archivados.
//
// Son la memoria del ⑫ de una feature futura: "¿esto ya está resuelto en algún
// lado?" y "¿qué aprendimos la vez pasada?". Se archivan CON la feature, y el
// que los encuentra es sf (regla 1.4).
func Journals() string {
	return filepath.Join(Archivado, "*", Journal)
}
