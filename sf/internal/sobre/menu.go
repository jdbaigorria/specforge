// El menú de modelos: la parte del sobre del ⑫ que hace que el ⑯ deje de elegir
// a ciegas.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ NO PODÍA ENTRAR POR EL CAMINO NORMAL
// ────────────────────────────────────────────────────────────────────────────
//
// El sobre del ⑫ eran tres partes y las tres RUTAS de archivos del repo. El
// catálogo de modelos no es un archivo del repo: vive en `.specforge/`, es de la
// máquina y no del proyecto — que es la decisión correcta y está argumentada en
// `global`. Pero la consecuencia nadie la había sacado: ninguna ruta puede
// apuntar ahí, así que el menú no entraba.
//
// Y sin embargo el ⑯ existe y tiene que decidir con qué se implementa cada lote.
// Un paso que decide sobre un catálogo que no ve es un paso que va a inventar, y
// lo que inventa es lo único que leyó: el ejemplo del skill. Ése fue el "siempre
// opus" (H1), y ésta es su otra mitad.
//
// El mecanismo ya estaba construido: `Parte.Contenido` existe para "material
// DERIVADO que no es un archivo". El menú es exactamente eso.
//
// ────────────────────────────────────────────────────────────────────────────
// LO QUE EL MENÚ NO LLEVA
// ────────────────────────────────────────────────────────────────────────────
//
// No lleva los `id`. El alias es vocabulario de Javier y sobrevive al cambio de
// harness; el id es del harness y no tiene por qué entrar a un archivo
// versionado. Que el ⑯ no vea los ids es la garantía más barata que hay de que
// no los pueda escribir.
//
// Y el `para:` viaja TEXTUAL, tal como Javier lo escribió. sf no lo resume, no lo
// reordena y no lo interpreta: es su opinión llegando entera a quien la necesita,
// que es lo único que R3 le permite hacer con ella.
package sobre

import (
	"fmt"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// TituloMenu es cómo se llama la parte. Lo usan los tests y el ㉑ para no
// escribirlo dos veces.
const TituloMenu = "Con qué se puede implementar esta feature"

// menu arma la parte del catálogo, o explica por qué no está.
//
// Nunca devuelve error: un catálogo que no se puede leer no puede impedir
// planificar. Lo que hace es DECIRLO, con `Falta`, que existe justamente para no
// mentir por omisión — si la parte simplemente no apareciera, el ⑯ creería que
// no hacía falta elegir y volvería a escribir lo primero que se le ocurra.
func menu(g *global.Config) Parte {
	p := Parte{Titulo: TituloMenu}

	if g == nil {
		p.Falta = "no hay catálogo de modelos todavía (nadie corrió `sf install`). " +
			"No elijas modelo: dejá que el default del perfil decida."
		return p
	}

	var b strings.Builder
	fmt.Fprintf(&b, "Estos son los modelos declarados en esta máquina para el harness `%s`.\n", g.Harness)
	b.WriteString("Elegí POR LOTE: no todos los lotes de una feature necesitan lo mismo.\n\n")

	hubo := false
	for _, perfil := range global.PerfilesConocidos() {
		lista := g.Activo()[perfil]
		if len(lista) == 0 {
			continue
		}
		hubo = true
		fmt.Fprintf(&b, "  perfil `%s`%s\n", perfil, paraQueSirve(perfil))
		for i, m := range lista {
			// El primero de la lista es el default del perfil, y decirlo importa:
			// sin eso, "no digo nada" parece azar en vez de una elección.
			marca := " "
			if i == 0 {
				marca = "→"
			}
			// La capacidad y el esfuerzo van juntos porque son la misma
			// pregunta vista de dos lados: cuánto puede pensar este alias, y
			// cuánto se le pidió que piense. Un alias no es un modelo — es una
			// forma de correr un modelo.
			como := ""
			if m.Capacidad > 0 {
				como = fmt.Sprintf("capacidad %d", m.Capacidad)
			}
			if m.Esfuerzo != "" {
				if como != "" {
					como += " · "
				}
				como += "esfuerzo " + m.Esfuerzo
			}
			fmt.Fprintf(&b, "   %s %-12s %-24s %s\n", marca, m.Alias, como, m.Para)
		}
		b.WriteString("\n")
	}

	if !hubo {
		p.Falta = "no hay ningún modelo declarado para este harness. " +
			"No elijas modelo: `sf next` va a parar y pedirlo antes de lanzar a nadie."
		return p
	}

	b.WriteString("Escribí el ALIAS —la primera columna—, nunca un id de modelo. " +
		"Un id acá\nsería un dato de esta máquina metido en un archivo versionado " +
		"que otro\nharness va a leer.\n\n")
	b.WriteString("El `→` es el default de cada perfil: es lo que va si no decís nada, " +
		"y no\ndecir nada es lo normal. Elegí sólo donde la diferencia sea real.\n")

	p.Contenido = b.String()
	return p
}

// paraQueSirve es la única línea que sf pone de su cosecha sobre un perfil.
//
// Y no es una opinión sobre modelos —eso sería R3—: es qué PIDE cada paso de la
// máquina, que es un dato del diseño y lo sabe sf y nadie más.
func paraQueSirve(perfil string) string {
	switch perfil {
	case global.Razonar:
		return "  — pedilo cuando el lote DECIDE algo, no cuando lo escribe"
	case global.Construir:
		return "  — el default de implementar"
	case global.Mecanico:
		return "  — procedimiento sin decisiones"
	}
	return ""
}
