package lanzar

import (
	"errors"
	"fmt"

	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/maquina"
)

// ErrNoHayTrabajo es que la máquina no está en `Trabajar`: una parada, un ME
// TRABÉ o el fin del roadmap.
//
// Es un sentinel y no un texto porque QUIEN LLAMA TIENE ALGO MEJOR QUE DECIR.
// La instrucción ya trae el motivo real y el comando para salir —"el paso pide
// el perfil construir y claude-code no lo tiene declarado: `sf model construir
// --alias …`"—, así que taparlo con un "no hay trabajo" genérico sería contestar
// peor de lo que sf ya sabe contestar. El que llama lo detecta y muestra la
// instrucción tal cual.
var ErrNoHayTrabajo = errors.New("no hay trabajo para lanzar")

// Puede dice si ese paso se puede lanzar headless, y si no, por qué.
//
// Contesta ANTES de armar nada: los motivos por los que un paso no se lanza no
// son fallas, son la máquina funcionando, y el que llama tiene que poder
// decirlos con sus palabras en vez de mostrar un error de ejecución.
func Puede(i maquina.Instruccion) error {
	if i.Tipo != maquina.Trabajar {
		return ErrNoHayTrabajo
	}

	// HEADLESS SÓLO PUEDE CORRER LO QUE NO CONVERSA, y eso ya está resuelto en
	// el código: el mapa `conversan` de maquina.go. El ⑥ porque ①–⑤ es un
	// pinponeo, y el ⑧ porque la constitución NO SE DERIVA DEL PRD — arquitectura,
	// stack y convenciones son decisiones suyas.
	//
	// No es una limitación del modo: un paso que necesita la opinión de Javier no
	// se puede delegar A NADIE, ni a un subagente ni a un proceso. Es la misma
	// regla que sacó al ⑧ de ser subagente el 2026-08-31.
	if i.Via == global.Vos {
		return fmt.Errorf("el %s se hace con Javier, no se delega", nombrar(i.Estado))
	}

	// Un `via: consola` es un modelo que se invoca a mano, con el prefijo que
	// Javier declaró en el catálogo. No pasa por la línea de ningún arnés, y
	// lanzarlo con la tabla de `linea.go` sería correr otra cosa distinta de la
	// que él declaró.
	if i.Via == global.Consola {
		return fmt.Errorf("%q va por consola: lo lanza el orquestador con sus manos, no sf", i.Modelo)
	}
	if i.Skill == "" {
		return errors.New("este paso no nombra ninguna skill: no hay a qué apuntar")
	}
	return nil
}

// nombrar le pone al estado un nombre que se pueda leer en un error.
func nombrar(estado string) string {
	switch estado {
	case "brief":
		return "brief (⑥)"
	case "constitucion":
		return "paso de la constitución (⑧)"
	}
	return "paso `" + estado + "`"
}

// Armar es el prompt que recibe el arnés.
//
// SF APUNTA, NO PEGA. Alcanza con nombrar la skill: los tres arneses resuelven
// solos la composición, los `references/` y los `templates/` — tres niveles,
// medido el 2026-08-31, y sin escalar ningún permiso. Así que `sf lanzar` no
// tiene que entender la estructura interna de ninguna skill, que era el
// acoplamiento nuevo que este diseño temía.
//
// La única condición es que el arnés ENCUENTRE las skills, y eso ya es trabajo
// de `sf install`: para Command Code escribe `settings.skills`, y opencode las
// encuentra sin que nadie se lo diga.
func Armar(i maquina.Instruccion, sobre string) string {
	if i.Skill == "" {
		return sobre
	}
	return "Usá la skill " + i.Skill + ".\n\n" + sobre
}
