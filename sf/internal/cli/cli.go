// Package cli es la superficie de sf: un comando por función, y el switch que
// los despacha.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ NO ESTÁ EN `main`, QUE ES DONDE ESTABA
// ────────────────────────────────────────────────────────────────────────────
//
// Estaba, y eran 1458 líneas: los quince handlers, la vista de la instrucción,
// el parseo de flags y la ayuda, todo en el mismo archivo. Nada de eso se podía
// testear desde un paquete —`main` no se importa desde ningún lado— así que la
// única red era el guion de humo.
//
// Esa red sigue siendo la que importa y no se reemplaza: lo que prueba son las
// COSTURAS ENTRE COMANDOS, y ésas sólo se ven corriendo el binario de verdad.
// Mudar el switch acá no le quita trabajo; le suma la posibilidad de que un
// handler se pruebe solo cuando haga falta.
//
// Lo que queda en `main` es lo único que no puede vivir en otro lado: `version`
// —el linker la inyecta con `-X main.version`, y eso sólo apunta a `main`— y el
// par abrir/cerrar del registro alrededor del despacho.
package cli

import (
	"fmt"
	"os"

	"github.com/jdbaigorria/specforge/sf/internal/comandos"
)

// Los códigos de salida son parte de la interfaz, no un detalle.
//
// Un CLI que siempre devuelve 0 obliga a parsear texto para saber qué pasó, y
// acá el que lee suele ser un agente. Con esto, "¿hay trabajo?" es un if en
// cualquier shell y en cualquier harness.
//
// Se exportan porque `main` los necesita para el `os.Exit` final: son lo único
// de este paquete que cruza la frontera, además de `Despachar` y `Uso`.
const (
	SalidaTrabajo = 0 // hay algo que hacer
	SalidaError   = 1 // algo se rompió de verdad
	SalidaParada  = 2 // 🛑 ⏸ ⚠ — no sigue sin Javier
	SalidaFin     = 3 // no queda nada
)

// Version es la del binario, y la escribe `main`: el linker sólo sabe inyectar
// en `main.version`. Acá la leen `sf version` y `sf doctor`.
var Version = "sin-versión (compilado del repo)"

func Despachar(args []string) int {
	switch args[1] {
	case "init":
		return iniciar()
	case "next":
		return next()
	case "context":
		return contexto(args[2:])
	case "done":
		return terminar(args[2:])
	case "new":
		return parada("new", args[2:])
	case "status":
		return estadoActual()
	case "audit":
		return auditar(args[2:])
	case "doctor":
		return revisar()
	case "version", "--version":
		// Una sola línea y nada más: `install.sh` la lee para comprobar que
		// instaló lo que creía. El informe para humanos es `sf doctor`.
		fmt.Println(Version)
		return SalidaTrabajo
	case "install":
		return instalar(args[2:])
	case "uninstall":
		return desinstalar()
	case "models":
		return listarModelos(args[2:])
	case "lanzar":
		return lanzarPaso(args[2:])
	case "log":
		return verRegistro(args[2:])
	case "approve", "reject", "take", "model", "dismiss":
		return parada(args[1], args[2:])
	case "lote":
		// `sf lote start` es el único comando de dos palabras del inventario.
		// Se mantiene así porque "lote" nombra la unidad de trabajo, y el día
		// que haga falta otra operación sobre el lote ya tiene dónde colgarse.
		if len(args) < 3 || args[2] != "start" {
			fmt.Fprintln(os.Stderr, "sf: el único es `sf lote start`")
			return SalidaError
		}
		return parada("lote start", nil)
	case "-h", "--help", "help":
		Uso()
		return SalidaTrabajo
	default:
		// Los diez del inventario están. Decirlo
		// con el nombre del que falta es más útil que un "comando desconocido":
		// el que lo lee suele ser un agente siguiendo el bucle.
		fmt.Fprintf(os.Stderr, "sf: %q todavía no está construido.\n", args[1])
		fmt.Fprintln(os.Stderr, "    Todos: "+comandos.Lista())
		return SalidaError
	}
}

func Uso() {
	fmt.Fprintln(os.Stderr, `sf — la máquina de estados de SpecForge

  sf install    pone CLAUDE.md y AGENTS.md · arma ~/.specforge/  (una vez por proyecto)
  sf uninstall  saca el orquestador. NO toca ~/.specforge/
  sf models     los ids que tu harness dice tener, para declararlos (--harness=<nombre>)
  sf init       el andamio: 2 directorios · detecta el stack · el estado vacío
  sf next       dónde estás · qué sigue · con qué skill y modelo
  sf lanzar     corre acá el paso que sigue  (--seco lo imprime y no lo corre)
  sf context    el sobre del estado actual  (--completo lo embebe)
  sf done       corre las compuertas y mueve  (--msg "…" en implementar)

  sf lote start   crea la branch · exige el ROJO · guarda el hash
  sf new "…"      mete una feature o un bug al backlog (entradas B y C)
  sf status       dónde está todo — el único para vos, no para el agente
  sf log          la película: qué pasó, en qué orden  (--vueltas · --feature)
  sf audit [f-#…]  el punta a punta: varias features contra sus historias
  sf doctor       ¿esta instalación sirve? binario · skills · harness
  sf version      la versión, en una línea

las cinco respuestas a una parada:
  sf approve              sella lo que estés mirando
  sf reject "motivo"      no sella, y el motivo viaja en el sobre
  sf take <feature>       el ⑪: elige de la cola
  sf model <alias>        sube el modelo        \
  sf dismiss <h-#> "…"    descarta un hallazgo  /  las salidas de ME TRABÉ

salidas:
  0  hay trabajo     2  parada (🛑 ⏸ ⚠)
  1  error           3  no queda nada`)
}
