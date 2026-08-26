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

	"github.com/jdbaigorria/specforge/sf/internal/andamio"
	"github.com/jdbaigorria/specforge/sf/internal/arranque"
	"github.com/jdbaigorria/specforge/sf/internal/auditoria"
	"github.com/jdbaigorria/specforge/sf/internal/comandos"
	"github.com/jdbaigorria/specforge/sf/internal/doctor"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/git"
	"github.com/jdbaigorria/specforge/sf/internal/global"
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
// version la pone el linker en el release:
//
//	go build -ldflags "-X main.version=v2.0.1"
//
// El default dice la verdad y no una versión inventada: quien compiló desde el
// repo NO tiene versión, y decir "v0.0.0" sería peor que decirlo.
var version = "sin-versión (compilado del repo)"

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
	case "audit":
		os.Exit(auditar(os.Args[2:]))
	case "doctor":
		os.Exit(revisar())
	case "version", "--version":
		// Una sola línea y nada más: `install.sh` la lee para comprobar que
		// instaló lo que creía. El informe para humanos es `sf doctor`.
		fmt.Println(version)
		os.Exit(salidaTrabajo)
	case "install":
		os.Exit(instalar(os.Args[2:]))
	case "uninstall":
		os.Exit(desinstalar())
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
		fmt.Fprintln(os.Stderr, "    Todos: "+comandos.Lista())
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

	// El mapa de modelos puede no existir —nadie corrió `sf install`— y eso NO
	// impide trabajar: `via()` cae a subagente. Lo único que no se puede
	// resolver sin él es `consola`.
	g, _ := global.Leer()

	i := maquina.Siguiente(raiz, e, r, g)
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

	// El mapa global sólo lo toca `sf model`, y sólo cuando trae `--via`. Se lee
	// igual para todos porque leerlo es barato y el error de "no está" ya está
	// contemplado: g queda nil y Modelo() lo maneja.
	g, _ := global.Leer()

	var ef maquina.Efecto
	switch cmd {
	case "approve":
		ef = maquina.Aprobar(raiz, e, r)
	case "reject":
		ef = maquina.Rechazar(e, arg(0))
	case "take":
		ef = maquina.Tomar(e, r, arg(0))
	case "model":
		nombre, via, comando, err := flagsDeModelo(args)
		if err != nil {
			fmt.Fprintln(os.Stderr, "sf model:", err)
			return salidaError
		}
		ef = maquina.Modelo(e, g, nombre, via, comando)
	case "dismiss":
		ef = maquina.Descartar(raiz, e, r, arg(0), arg(1))
	case "lote start":
		ef = maquina.EmpezarLote(raiz, e, r)
	case "new":
		// Todo lo que venga después del comando es el texto, no flags: `sf new
		// que sf soporte brownfield` tiene que funcionar sin comillas.
		ef = maquina.Nueva(raiz, e, strings.Join(args, " "))
	}

	// El estado se guarda si el comando funcionó. Un `sf take f-99` que falla no
	// tiene que dejar rastro.
	//
	// Y `Cambio` es la excepción, que existe por un caso concreto: `archivar`
	// toca el disco —mueve la carpeta, mergea, borra la branch— y si algo falla
	// DESPUÉS de eso, no guardar deja al estado.json describiendo un repo que ya
	// no existe. Cuando el comando falló y el mundo cambió igual, lo correcto es
	// anotar lo que sí pasó.
	if ef.Pasa() || ef.Cambio {
		if err := e.Guardar(raiz); err != nil {
			fmt.Fprintln(os.Stderr, "sf: no pude guardar el estado:", err)
			return salidaError
		}
		// Y el mapa global sólo cuando algo lo cambió: escribirlo en cada
		// `sf approve` sería tocar el home de Javier para no cambiar nada.
		if ef.Global && g != nil {
			if err := g.Guardar(); err != nil {
				fmt.Fprintln(os.Stderr, "sf: no pude guardar ~/.specforge/:", err)
				return salidaError
			}
		}
		// El commit va acá y no adentro del comando, porque recién ahora el
		// estado.json está escrito y puede entrar en el mismo commit. Hoy sólo
		// lo pide `archivar`: es el único lugar donde sf mueve archivos sin
		// estar cerrando un lote.
		//
		// Que falle NO invalida lo que ya pasó —la carpeta se movió, la branch
		// se mergeó, el estado se guardó—, así que avisa y sigue. Frenar acá
		// sería reportar como error un archivado que salió bien.
		if ef.Commit != "" {
			if _, err := git.Commit(raiz, ef.Commit); err != nil {
				fmt.Fprintln(os.Stderr, "sf: quedó sin commitear:", err)
			}
		}
	}

	fmt.Print(ef.Texto())
	if ef.Pasa() {
		return salidaTrabajo
	}
	return salidaError
}

// flagsDeModelo parte `sf model <nombre> [--via …] [--comando "…"]`.
//
// Los dos flags son cómo se DECLARA un modelo nuevo, y por eso van acá y no en
// un `sf model add` aparte: aprobar es declarar (ver maquina.Modelo).
func flagsDeModelo(args []string) (nombre, via, comando string, err error) {
	for i := 0; i < len(args); i++ {
		a := args[i]
		siguiente := func() (string, bool) {
			if i+1 < len(args) {
				i++
				return args[i], true
			}
			return "", false
		}

		switch {
		case a == "--via":
			v, ok := siguiente()
			if !ok {
				return "", "", "", errors.New("`--via` sin valor. Es `subagente` o `consola`.")
			}
			via = v
		case strings.HasPrefix(a, "--via="):
			via = strings.TrimPrefix(a, "--via=")
		case a == "--comando":
			c, ok := siguiente()
			if !ok {
				return "", "", "", errors.New("`--comando` sin valor")
			}
			comando = c
		case strings.HasPrefix(a, "--comando="):
			comando = strings.TrimPrefix(a, "--comando=")
		case strings.HasPrefix(a, "-"):
			return "", "", "", fmt.Errorf("no conozco %q. Sólo --via y --comando", a)
		default:
			if nombre != "" {
				return "", "", "", fmt.Errorf("dos modelos: %q y %q", nombre, a)
			}
			nombre = a
		}
	}
	return nombre, via, comando, nil
}

// instalar es `sf install`: el andamio.
//
// No es de la máquina —no mira el estado ni lo mueve— y por eso no aparece en
// ningún trazado del bucle. Es lo que hace que SpecForge se pueda USAR: pone el
// orquestador en el proyecto y arma ~/.specforge/.
func instalar(args []string) int {
	var o andamio.Opciones
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--forzar" || a == "--force":
			o.Forzar = true
		case a == "--harness" && i+1 < len(args):
			i++
			o.Harness = args[i]
		case strings.HasPrefix(a, "--harness="):
			o.Harness = strings.TrimPrefix(a, "--harness=")
		default:
			fmt.Fprintf(os.Stderr, "sf install: no conozco %q. Sólo --harness y --forzar.\n", a)
			return salidaError
		}
	}

	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	r, err := andamio.Instalar(raiz, o)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	for _, e := range r.Escritos {
		fmt.Println("+", e)
	}
	for _, s := range r.Salteados {
		fmt.Println("·", s)
	}

	fmt.Println()
	fmt.Printf("harness: %s · %d modelos declarados\n", r.Harness, r.Modelos)
	if r.Harness == "desconocido" {
		// Se avisa fuerte porque el harness es la mitad de H1b: sin él, sf no
		// puede decidir si un modelo va por subagente o por consola.
		fmt.Println("⚠ No reconocí el harness. Corregilo con `sf install --harness=<nombre>`.")
	}
	return salidaTrabajo
}

// revisar es `sf doctor`: ¿esta instalación sirve?
//
// Como `sf install`, no es de la máquina: no mira el estado ni lo mueve. Mira
// la MÁQUINA DE JAVIER, que es lo único que ningún test del repo puede ver.
//
// El código de salida sigue la misma convención que todo lo demás, y por eso
// devuelve `salidaParada` y no `salidaError` cuando encuentra algo: no se rompió
// nada, hace falta que alguien intervenga. Un `sf doctor && ...` en un script de
// instalación distingue los tres casos sin parsear texto.
func revisar() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	i := doctor.Revisar(raiz, version)
	fmt.Print(i.Texto())
	if !i.Sano() {
		return salidaParada
	}
	return salidaTrabajo
}

// desinstalar es `sf uninstall`: saca el orquestador y NO toca ~/.specforge/.
func desinstalar() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	r, err := andamio.Desinstalar(raiz)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	for _, e := range r.Escritos {
		fmt.Println("-", e)
	}
	for _, s := range r.Salteados {
		fmt.Println("·", s)
	}
	if len(r.Escritos) == 0 && len(r.Salteados) == 0 {
		fmt.Println("No había nada que sacar.")
	}
	fmt.Println()
	fmt.Println("~/.specforge/ no se toca: tus modelos son de la máquina, no de este proyecto.")
	return salidaTrabajo
}

// auditar es `sf audit`: el punta a punta sobre varias features.
//
// Es el único comando que NO es de la máquina y que igual sirve un sobre. No
// rompe H2 —`sf context` sigue sin argumentos— porque son preguntas distintas:
// `sf context` pregunta "¿qué necesito para el paso en el que estoy?", y eso es
// deducible. Acá el alcance NO es deducible: lo elige Javier.
//
//	sf audit                 todo lo que se construyó
//	sf audit f-1 f-2 f-3     estas tres — un "módulo" es un conjunto de features
//	sf audit --completo      embebe el material, para un modelo sin shell
func auditar(args []string) int {
	completo := false
	var ids []string
	for _, a := range args {
		switch {
		case a == "--completo" || a == "--full":
			completo = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf audit: no conozco %q. Sólo --completo.\n", a)
			return salidaError
		default:
			ids = append(ids, a)
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

	alcance, err := auditoria.Alcance(e, r, ids)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	inf, err := auditoria.Auditar(raiz, e, r, alcance)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return salidaError
	}

	fmt.Print(inf.Texto(raiz, completo))

	// Con fallas sale por parada y no por error: no se rompió nada: sf encontró
	// contradicciones y alguien tiene que mirarlas. Es la misma distinción que
	// hace `sf done`.
	if inf.Fallas() > 0 {
		return salidaParada
	}
	return salidaTrabajo
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
		// El comando sólo aparece con `via: consola`, y es lo que el
		// orquestador tiene que tipear. Sin esta línea, "consola" sería una
		// instrucción que no se puede ejecutar.
		campo("comando", i.Comando)

		// Los avisos van ANTES del mensaje: son lo que puede cambiar cómo se
		// hace el trabajo, y el mensaje es sólo qué toca hacer.
		for _, a := range i.Avisos {
			fmt.Fprintf(&b, "\n⚠ %s\n", a)
		}

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

  sf install    pone CLAUDE.md y AGENTS.md · arma ~/.specforge/  (una vez por proyecto)
  sf uninstall  saca el orquestador. NO toca ~/.specforge/
  sf init       el andamio: 2 directorios · detecta el stack · el estado vacío
  sf next       dónde estás · qué sigue · con qué skill y modelo
  sf context    el sobre del estado actual  (--completo lo embebe)
  sf done       corre las compuertas y mueve  (--msg "…" en implementar)

  sf lote start   crea la branch · exige el ROJO · guarda el hash
  sf new "…"      mete una feature o un bug al backlog (entradas B y C)
  sf status       dónde está todo — el único para vos, no para el agente
  sf audit [f-#…]  el punta a punta: varias features contra sus historias
  sf doctor       ¿esta instalación sirve? binario · skills · harness
  sf version      la versión, en una línea

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
