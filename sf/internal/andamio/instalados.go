package andamio

import (
	"context"
	"os/exec"
	"strings"
	"time"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// ────────────────────────────────────────────────────────────────────────────
// QUÉ ARNESES HAY EN ESTA MÁQUINA
// ────────────────────────────────────────────────────────────────────────────
//
// Es una pregunta DISTINTA de las otras dos que este repo ya contesta, y las
// tres se confunden fácil:
//
//	global.DetectarHarness()   en cuál estoy corriendo AHORA    hecho del proceso
//	global.Config.Harness      cuál instalé la última vez       configuración
//	andamio.Instalados()       cuáles hay en esta máquina       hecho del disco
//
// La tercera existe para que `sf install` no pregunte a ciegas. Preguntar "¿qué
// arneses vas a usar?" con los tres siempre en la lista es hacerle elegir a
// Javier entre cosas que capaz no tiene; mirando el PATH, el menú ofrece lo que
// de verdad puede correr.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ NO ALCANZA CON QUE EL BINARIO EXISTA
// ────────────────────────────────────────────────────────────────────────────
//
// Porque el de Command Code se llama `cmd`, que es un nombre carísimo de
// genérico: en Windows es la shell. Encontrar un archivo llamado `cmd` NO prueba
// que sea Command Code, y ofrecerlo haría que la instalación arme `.commandcode/`
// para algo que no existe.
//
// Así que se le pregunta, y ACÁ HAY UNA ASIMETRÍA REAL Y MEDIDA (2026-08-31):
//
//	claude --version     "2.1.252 (Claude Code)"    dice su nombre
//	cmd --help           "Command Code v1.39.2"     dice su nombre (en --help, no en --version)
//	opencode --version   "1.18.25"                  NO dice su nombre
//
// De los tres, `opencode --version` es el único que imprime un número pelado. Por
// eso su confirmación es más floja —que el comando ANDE— y no una marca en la
// salida. Inventarle una marca que no emite sería adivinar, y las dos veces que
// en este proyecto se adivinó un flag se adivinó mal.

// arneses es con qué se llama a cada uno, y cómo se confirma que es él.
//
// `dice` vacío significa "que ande alcanza": ver la asimetría de arriba.
var arneses = map[string]struct {
	binario string
	args    []string
	dice    string
}{
	"claude-code": {"claude", []string{"--version"}, "Claude Code"},
	"opencode":    {"opencode", []string{"--version"}, ""},
	"commandcode": {"cmd", []string{"--help"}, "Command Code"},
}

// esperaIdentidad es cuánto se le da a un arnés para decir quién es.
//
// Medido: el más lento de los tres es `cmd --help` con 0,7s. Cinco segundos es
// holgado y sigue siendo un tope — un binario colgado no puede colgar un
// `sf install`, que es lo primero que corre alguien que recién llega.
const esperaIdentidad = 5 * time.Second

// buscar y preguntar son variables para que los tests no dependan de qué haya
// instalado en la máquina donde corre el CI.
var (
	buscar    = exec.LookPath
	preguntar = func(ctx context.Context, binario string, args ...string) ([]byte, error) {
		// CombinedOutput y no Output: `cmd --help` escribe por stdout, pero un
		// binario que no es el que buscamos bien puede quejarse por stderr, y
		// para descartarlo da igual por dónde hable.
		return exec.CommandContext(ctx, binario, args...).CombinedOutput()
	}
)

// Instalados dice qué arneses hay en esta máquina, en el orden de global.Harness.
//
// El orden importa y no es el del mapa: en Go recorrer un mapa da un orden
// distinto cada vez, y un menú que se reordena solo es un menú donde te
// equivocás de número.
func Instalados() []string {
	var hay []string
	for _, h := range global.Harness {
		if Hay(h) {
			hay = append(hay, h)
		}
	}
	return hay
}

// Hay dice si ESE arnés está instalado y se identifica.
func Hay(harness string) bool {
	a, ok := arneses[harness]
	if !ok {
		return false
	}
	if _, err := buscar(a.binario); err != nil {
		return false
	}

	ctx, cancelar := context.WithTimeout(context.Background(), esperaIdentidad)
	defer cancelar()

	salida, err := preguntar(ctx, a.binario, a.args...)
	if err != nil {
		return false
	}
	if a.dice == "" {
		return true
	}
	return strings.Contains(string(salida), a.dice)
}
