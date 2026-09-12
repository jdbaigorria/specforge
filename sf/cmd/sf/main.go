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
	"os"

	"github.com/jdbaigorria/specforge/sf/internal/cli"
	"github.com/jdbaigorria/specforge/sf/internal/registro"
)

// version la pone el linker en el release:
//
//	go build -ldflags "-X main.version=v2.0.1"
//
// El default dice la verdad y no una versión inventada: quien compiló desde el
// repo NO tiene versión, y decir "v0.0.0" sería peor que decirlo.
//
// Vive acá y no en `cli` porque `-X` sólo sabe escribir en el paquete que le
// nombran, y el que se linkea es éste. `main` la pasa, `cli` la lee.
var version = "sin-versión (compilado del repo)"

// main abre el registro, despacha, y cierra el registro. Nada más.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ EL SWITCH NO ESTÁ ACÁ, Y POR QUÉ DEVUELVE EN VEZ DE SALIR
// ────────────────────────────────────────────────────────────────────────────
//
// Lo segundo primero, porque es lo que hace posible lo primero: `os.Exit` se
// saltea TODOS los defer. Con el switch haciendo `os.Exit(fn())` en cada rama
// no había ningún lugar donde colgar el cierre del registro. Partirlo en dos es
// lo que hace que la línea se escriba SIEMPRE, sin importar por qué rama salió.
//
// Y engancha acá y no en los veinte comandos a propósito: veinte llamadas
// serían veinte copias, y agregar el comando veintiuno sin acordarse dejaría un
// hueco silencioso. Es la misma razón por la que existe el paquete `comandos`.
//
// El switch se mudó a `internal/cli` cuando este archivo llegó a las 1458
// líneas. Lo que quedó acá es lo que no puede estar en otro lado: la versión
// que inyecta el linker, y este par abrir/cerrar alrededor del despacho.
func main() {
	cli.Version = version

	if len(os.Args) < 2 {
		cli.Uso()
		os.Exit(cli.SalidaError)
	}

	raiz, err := os.Getwd()
	if err == nil {
		registro.Abrir(raiz, os.Args[1], os.Args[2:])
	}

	codigo := cli.Despachar(os.Args)

	registro.Cerrar(codigo)
	os.Exit(codigo)
}
