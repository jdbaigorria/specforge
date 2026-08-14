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

	"github.com/jdbaigorria/specforge/sf/internal/arranque"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/maquina"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
	"github.com/jdbaigorria/specforge/sf/internal/sobre"
	"github.com/jdbaigorria/specforge/sf/internal/vista"
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
	case "init":
		os.Exit(iniciar())
	case "next":
		os.Exit(next())
	case "context":
		os.Exit(contexto(os.Args[2:]))
	case "done":
		os.Exit(terminar(os.Args[2:]))
	case "new":
		os.Exit(parada("new", os.Args[2:]))
	case "status":
		os.Exit(estadoActual())
	case "approve", "reject", "take", "model", "dismiss":
		os.Exit(parada(os.Args[1], os.Args[2:]))
	case "lote":
		// `sf lote start` es el único comando de dos palabras del inventario.
		// Se mantiene así porque "lote" nombra la unidad de trabajo, y el día
		// que haga falta otra operación sobre el lote ya tiene dónde colgarse.
		if len(os.Args) < 3 || os.Args[2] != "start" {
			fmt.Fprintln(os.Stderr, "sf: el único es `sf lote start`")
			os.Exit(salidaError)
		}
		os.Exit(parada("lote start", nil))
	case "-h", "--help", "help":
		uso()
		os.Exit(salidaTrabajo)
	default:
		// Los diez del inventario están. Decirlo
		// con el nombre del que falta es más útil que un "comando desconocido":
		// el que lo lee suele ser un agente siguiendo el bucle.
		fmt.Fprintf(os.Stderr, "sf: %q todavía no está construido.\n", os.Args[1])
		fmt.Fprintln(os.Stderr, "    Todos: init · next · context · done · lote start · new · status · approve · reject · take · model · dismiss")
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

// terminar es `sf done`: corre las compuertas y mueve lo que corresponda.
//
// Lo llama EL QUE TRABAJA, antes de morir — no el orquestador. Si `sf done` da
// ✗, el subagente todavía tiene el contexto para arreglarlo ahí mismo; si lo
// corriera el orquestador, cada ✗ costaría un subagente nuevo desde cero (H6).
//
// El único flag es `--msg`, y sólo aplica a `implementar`: sin él no se cierra
// el lote, porque cerrar el lote ES commitear.
func terminar(args []string) int {
	var msg string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--msg" && i+1 < len(args):
			i++
			msg = args[i]
		case strings.HasPrefix(args[i], "--msg="):
			msg = strings.TrimPrefix(args[i], "--msg=")
		default:
			fmt.Fprintf(os.Stderr, "sf done: no conozco %q. Sólo --msg \"…\"\n", args[i])
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

	c := maquina.Terminar(raiz, e, r, msg)

	// El estado se guarda SÓLO si algo se movió. Un `sf done` que falla no
	// tiene que dejar rastro en el archivo... salvo el contador de intentos,
	// que es justamente el que cuenta los fracasos: por eso también se guarda
	// cuando falla dentro de una feature.
	if c.Movio || c.Cambio {
		if err := e.Guardar(raiz); err != nil {
			fmt.Fprintln(os.Stderr, "sf: no pude guardar el estado:", err)
			return salidaError
		}
	}

	fmt.Print(c.Texto())
	if c.Mensaje != "" {
		fmt.Println("→ " + c.Mensaje)
	}

	if c.Pasa() {
		return salidaTrabajo
	}
	return salidaParada
}

// parada corre los cinco comandos con los que Javier le contesta a una parada.
//
// Van juntos en una función porque comparten el esqueleto entero —cargar,
// ejecutar, guardar el estado, imprimir— y lo único que cambia es qué llaman.
// Separarlos sería copiar cinco veces las mismas veinte líneas.
//
// Y comparten algo más importante: LOS CINCO LOS CORRE EL ORQUESTADOR, nunca el
// subagente. Son la mitad del reparto que la ronda de la superficie descubrió.
func parada(cmd string, args []string) int {
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

	arg := func(i int) string {
		if i < len(args) {
			return args[i]
		}
		return ""
	}

	var ef maquina.Efecto
	switch cmd {
	case "approve":
		ef = maquina.Aprobar(raiz, e, r)
	case "reject":
		ef = maquina.Rechazar(e, arg(0))
	case "take":
		ef = maquina.Tomar(e, r, arg(0))
	case "model":
		ef = maquina.Modelo(e, arg(0))
	case "dismiss":
		ef = maquina.Descartar(raiz, e, r, arg(0), arg(1))
	case "lote start":
		ef = maquina.EmpezarLote(raiz, e, r)
	case "new":
		// Todo lo que venga después del comando es el texto, no flags: `sf new
		// que sf soporte brownfield` tiene que funcionar sin comillas.
		ef = maquina.Nueva(raiz, e, strings.Join(args, " "))
	}

	// El estado se guarda sólo si el comando funcionó. Un `sf take f-99` que
	// falla no tiene que dejar rastro.
	if ef.Pasa() {
		if err := e.Guardar(raiz); err != nil {
			fmt.Fprintln(os.Stderr, "sf: no pude guardar el estado:", err)
			return salidaError
		}
	}

	fmt.Print(ef.Texto())
	if ef.Pasa() {
		return salidaTrabajo
	}
	return salidaError
}

// iniciar es `sf init`: el andamio, y lo único que se corre antes que nada.
//
// No es un comando de la máquina —no mira el estado ni lo mueve— y por eso no
// aparece en ningún trazado del bucle. Es lo que hace que el bucle PUEDA
// empezar: sin estado.json, `sf next` sólo sabe decir "corré sf init".
func iniciar() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	r, err := arranque.Iniciar(raiz)
	if err != nil {
		// Ya iniciado NO es un error del usuario: es alguien que corrió el
		// comando dos veces. La respuesta útil es decirle por dónde seguir, no
		// retarlo — y por eso sale por stdout con código de parada.
		if errors.Is(err, arranque.ErrYaIniciado) {
			fmt.Println("Este proyecto ya está iniciado.")
			fmt.Println()
			fmt.Println("  sf next")
			return salidaParada
		}
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	for _, c := range r.Creados {
		fmt.Println("+", c)
	}
	fmt.Println()

	if r.Stack.Reconocido() {
		fmt.Printf("%s (%s) · test_cmd: %s\n", r.Stack.Lenguaje, r.Stack.Manifiesto, r.Stack.TestCmd)
	} else {
		// Se avisa fuerte porque sin test_cmd la constitución NO SELLA, y ese
		// freno aparecería recién en el ⑧ sin decir de dónde viene.
		fmt.Println("⚠ No reconocí el stack: completá `test_cmd:` en la constitución.")
		fmt.Println("  Sin eso el ⑧ no sella y no se puede correr ningún test.")
	}

	fmt.Println()
	fmt.Println("  sf next")
	return salidaTrabajo
}

// estadoActual es `sf status`: la única salida de sf pensada para un humano.
//
// No mueve nada ni comprueba nada — lee tres fuentes y arma una vista. Por eso
// no aparece en ningún trazado del bucle: el bucle no lo necesita nunca.
func estadoActual() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	e, r, err := cargar(raiz)
	if err != nil {
		if errors.Is(err, estado.ErrNoHay) {
			fmt.Println("Este proyecto todavía no tiene estado.")
			fmt.Println()
			fmt.Println("  sf init")
			return salidaParada
		}
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	fmt.Print(vista.Estado(raiz, e, r))
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

  sf init       el andamio: 2 directorios · detecta el stack · el estado vacío
  sf next       dónde estás · qué sigue · con qué skill y modelo
  sf context    el sobre del estado actual  (--completo lo embebe)
  sf done       corre las compuertas y mueve  (--msg "…" en implementar)

  sf lote start   crea la branch · exige el ROJO · guarda el hash
  sf new "…"      mete una feature o un bug al backlog (entradas B y C)
  sf status       dónde está todo — el único para vos, no para el agente

las cinco respuestas a una parada:
  sf approve              sella lo que estés mirando
  sf reject "motivo"      no sella, y el motivo viaja en el sobre
  sf take <feature>       el ⑪: elige de la cola
  sf model <nombre>       sube el modelo        \
  sf dismiss <h-#> "…"    descarta un hallazgo  /  las salidas de ME TRABÉ

salidas:
  0  hay trabajo     2  parada (🛑 ⏸ ⚠)
  1  error           3  no queda nada`)
}
