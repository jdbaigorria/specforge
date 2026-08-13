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
	"github.com/jdbaigorria/specforge/sf/internal/sobre"
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
	case "context":
		os.Exit(contexto(os.Args[2:]))
	case "-h", "--help", "help":
		uso()
		os.Exit(salidaTrabajo)
	default:
		// Los otros ocho comandos del inventario todavía no existen. Decirlo
		// con el nombre del que falta es más útil que un "comando desconocido":
		// el que lo lee suele ser un agente siguiendo el bucle.
		fmt.Fprintf(os.Stderr, "sf: %q todavía no está construido.\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "    Por ahora: sf next · sf context")
		os.Exit(salidaError)
	}
}

func next() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	e, r, err := cargar(raiz)
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

// contexto es `sf context`: el sobre del estado actual.
//
// El único flag es `--completo`, y no es una comodidad: es el caso de H1b.
// Cuando el que trabaja es un modelo por consola sin shell propia, no puede
// abrir archivos — así que el orquestador corre esto y le pega la salida en el
// prompt. Mismo comando, mismo sobre; lo único que cambia es quién lo tipea.
//
// No hay ningún otro argumento a propósito: el estado, la feature y el lote son
// todos deducibles del estado.json, y un argumento deducible es un argumento
// que se pasa mal (H2).
func contexto(args []string) int {
	completo := false
	for _, a := range args {
		switch a {
		case "--completo", "--full":
			completo = true
		default:
			fmt.Fprintf(os.Stderr, "sf context: no conozco %q.\n", a)
			fmt.Fprintln(os.Stderr, "    El estado, la feature y el lote salen del estado.json.")
			return salidaError
		}
	}

	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	e, r, err := cargar(raiz)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	s, err := sobre.Armar(raiz, e, r)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	fmt.Print(s.Texto(raiz, completo))
	return salidaTrabajo
}

// cargar lee el estado y el roadmap, que es lo que necesitan los dos comandos.
//
// El roadmap puede no existir todavía —los cinco estados de producto corren
// antes que él— y eso NO es un error: se devuelve nil y quien lo use sabe que
// hasta el ⑩ no hay cola. El estado, en cambio, sí hace falta siempre.
func cargar(raiz string) (*estado.Estado, *roadmap.Roadmap, error) {
	e, err := estado.Leer(raiz)
	if err != nil {
		return nil, nil, err
	}

	r, err := roadmap.Leer(raiz)
	if err != nil {
		if !errors.Is(err, roadmap.ErrNoHay) {
			return nil, nil, err
		}
		r = nil
	}
	return e, r, nil
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

  sf next       dónde estás · qué sigue · con qué skill y modelo
  sf context    el sobre del estado actual  (--completo lo embebe)

salidas:
  0  hay trabajo     2  parada (🛑 ⏸ ⚠)
  1  error           3  no queda nada`)
}
