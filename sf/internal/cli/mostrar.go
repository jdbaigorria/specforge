package cli

import (
	"fmt"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/maquina"
)

// mostrar arma el texto que ve el orquestador.
//
// Está separado de la decisión y devuelve un string en vez de imprimir, por una
// razón práctica: así el test compara texto en vez de capturar stdout, que en
// Go es incómodo y frágil.
//
// El formato es "campo: valor" porque el lector es un LLM y esa forma no se
// presta a confusión. Los campos vacíos no se imprimen: una línea "feature:"
// sin nada al lado es ruido que el modelo tiene que interpretar.
func mostrar(i maquina.Instruccion) string {
	var b strings.Builder

	campo := func(nombre, valor string) {
		if valor != "" {
			fmt.Fprintf(&b, "%-9s %s\n", nombre+":", valor)
		}
	}

	if i.Tipo == maquina.Trabajar {
		campo("estado", i.Estado)

		feature := i.Feature
		if i.Lote > 0 {
			// "f-2 · lote 2 de 4" — el lote no es un campo aparte porque
			// siempre se lee junto a la feature.
			feature = fmt.Sprintf("%s · lote %d de %d", feature, i.Lote, i.DeLotes)
		}
		campo("feature", feature)
		campo("skill", i.Skill)
		// Perfil y modelo son dos preguntas distintas y por eso son dos
		// renglones: qué PIDE el paso, y qué se le da en esta máquina. Con una
		// sola palabra —como era antes de H2— las dos tenían la misma respuesta
		// y no se podían distinguir al leer el log.
		campo("perfil", i.Perfil)
		campo("modelo", i.Modelo)
		// El esfuerzo sólo aparece si el alias lo declara: vacío significa "el
		// que traiga el modelo", que no es lo mismo que "bajo".
		campo("esfuerzo", i.Esfuerzo)
		// El agente sólo aparece donde hace falta: en Claude Code el modelo va
		// como parámetro de la llamada y este renglón sería ruido.
		campo("agente", i.Agente)
		campo("via", i.Via)
		// El comando sólo aparece con `via: consola`, y es lo que el
		// orquestador tiene que tipear. Sin esta línea, "consola" sería una
		// instrucción que no se puede ejecutar.
		campo("comando", i.Comando)

		// Los avisos van ANTES del mensaje: son lo que puede cambiar cómo se
		// hace el trabajo, y el mensaje es sólo qué toca hacer.
		avisar(&b, i.Avisos)

		if i.Mensaje != "" {
			fmt.Fprintf(&b, "\n%s\n", i.Mensaje)
		}
	} else {
		if i.Mensaje != "" {
			fmt.Fprintf(&b, "%s\n", i.Mensaje)
		}
		// Una PARADA también avisa, y antes esto se tragaba: los avisos sólo se
		// imprimían en la rama de `Trabajar`.
		//
		// El caso que lo destapó es el más caro de todos: la 🛑 de un perfil sin
		// declarar trae "reiniciá tu harness, los agentes se leen al arrancar".
		// Sin ese renglón, Javier declara los modelos, no reinicia, y el siguiente
		// intento falla por un motivo que sf ya sabía y no dijo.
		avisar(&b, i.Avisos)

		// El horizonte va DESPUÉS del mensaje y ANTES de los comandos, que es
		// el orden en que se lee una parada: qué pasa, qué desencadena si digo
		// que sí, y recién ahí con qué se dice.
		//
		// Existe porque `sf next` decía dónde estás y nunca qué ibas a
		// disparar. No todos los pasos frenan, así que un `sf approve` arranca
		// todos los que siguen hasta la próxima parada — y eso, sin anunciarlo,
		// se siente como que la máquina se te escapó.
		if len(i.SiApruebas) > 0 {
			fmt.Fprintf(&b, "\n   si aprobás corre:  %s\n", strings.Join(i.SiApruebas, " → "))
			if i.ProximaParada != "" {
				fmt.Fprintf(&b, "   próxima parada:    %s\n", i.ProximaParada)
			}
		}
	}

	if len(i.Sugerido) > 0 {
		b.WriteString("\n")
		for _, c := range i.Sugerido {
			fmt.Fprintf(&b, "  %s\n", c)
		}
	}
	return b.String()
}

// avisar imprime los ⚠ de una instrucción, si los hay.
//
// Existe para que las dos ramas —trabajo y parada— no puedan divergir otra vez:
// tenerlo escrito dos veces fue exactamente cómo una de las dos se quedó sin
// avisos y nadie lo notó.
func avisar(b *strings.Builder, avisos []string) {
	for _, a := range avisos {
		fmt.Fprintf(b, "\n⚠ %s\n", a)
	}
}
