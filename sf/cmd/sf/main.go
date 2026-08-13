// Command sf es la máquina de estados que guía al harness.
//
// ────────────────────────────────────────────────────────────────────────────
// LO QUE ES, EN UNA LÍNEA
// ────────────────────────────────────────────────────────────────────────────
//
//	sf es una tool que expone la máquina de estados que guía al harness —
//	y es el árbitro que decide si se puede avanzar.       (session.md §6)
//
// Son dos verbos y el segundo es el que importa. Sin "expone" sería una lista
// de tareas; sin "comprueba" sería una sugerencia que el agente puede ignorar.
//
//	sf       →  DÓNDE estás · QUÉ sigue · ¿PODÉS avanzar?
//	skills   →  CÓMO se hace
//	harness  →  lo HACE
//	Javier   →  DECIDE  (⑥, ⑧, ⑰)
//
// sf arranca, contesta y se muere. Dura milisegundos. Por eso NO puede lanzar a
// nadie: un programa muerto no tiene manos. El único vivo durante toda la
// sesión es el agente, y por eso el orquestador es él.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/maquina"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
)

// Los códigos de salida son parte de la interfaz, no un detalle.
//
// Un CLI que siempre devuelve 0 obliga a parsear texto para saber qué pasó, y
// acá el que lee suele ser un agente. Con esto, "¿hay trabajo?" es un if en
// cualquier shell y en cualquier harness.
const (
	salidaTrabajo = 0 // hay algo que hacer
	salidaError   = 1 // algo se rompió de verdad
	salidaParada  = 2 // 🛑 ⏸ ⚠ — no sigue sin Javier
	salidaFin     = 3 // no queda nada
)

func main() {
	if len(os.Args) < 2 {
		uso()
		os.Exit(salidaError)
	}

	switch os.Args[1] {
	case "next":
		os.Exit(next())
	case "-h", "--help", "help":
		uso()
		os.Exit(salidaTrabajo)
	default:
		// Los otros nueve comandos del inventario todavía no existen. Decirlo
		// con el nombre del que falta es más útil que un "comando desconocido":
		// el que lo lee suele ser un agente siguiendo el bucle.
		fmt.Fprintf(os.Stderr, "sf: %q todavía no está construido.\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "    Por ahora sólo: sf next")
		os.Exit(salidaError)
	}
}

func next() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	e, err := estado.Leer(raiz)
	if err != nil {
		// "No hay estado" no es un error: es un proyecto sin arrancar, y la
		// respuesta útil es decir cómo se arranca. Cualquier otro error sí es
		// un problema y se muestra tal cual.
		if errors.Is(err, estado.ErrNoHay) {
			fmt.Println("Este proyecto todavía no tiene estado.")
			fmt.Println()
			fmt.Println("  sf init")
			return salidaParada
		}
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	// El roadmap puede no existir todavía —los cinco estados de producto corren
	// antes que él— y eso no es un error: se le pasa nil a la máquina, que sabe
	// que hasta el ⑩ no hay cola.
	var r *roadmap.Roadmap
	if r, err = roadmap.Leer(raiz); err != nil {
		if !errors.Is(err, roadmap.ErrNoHay) {
			fmt.Fprintln(os.Stderr, "sf:", err)
			return salidaError
		}
		r = nil
	}

	i := maquina.Siguiente(raiz, e, r)
	fmt.Print(mostrar(i))

	switch i.Tipo {
	case maquina.Trabajar:
		return salidaTrabajo
	case maquina.Fin:
		return salidaFin
	default:
		return salidaParada
	}
}

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
		campo("modelo", i.Modelo)
		campo("via", i.Via)

		if i.Mensaje != "" {
			fmt.Fprintf(&b, "\n%s\n", i.Mensaje)
		}
	} else if i.Mensaje != "" {
		fmt.Fprintf(&b, "%s\n", i.Mensaje)
	}

	if len(i.Sugerido) > 0 {
		b.WriteString("\n")
		for _, c := range i.Sugerido {
			fmt.Fprintf(&b, "  %s\n", c)
		}
	}
	return b.String()
}

func uso() {
	fmt.Fprintln(os.Stderr, `sf — la máquina de estados de SpecForge

  sf next     dónde estás · qué sigue · con qué skill y modelo

salidas:
  0  hay trabajo     2  parada (🛑 ⏸ ⚠)
  1  error           3  no queda nada`)
}
