// Package lanzar corre un paso de la máquina en un arnés headless.
//
// ────────────────────────────────────────────────────────────────────────────
// QUÉ CAMBIA CUANDO SF LANZA
// ────────────────────────────────────────────────────────────────────────────
//
// Hoy `sf` no lanza modelos: le deja al arnés un papel con el número anotado y
// le pide que llame. De ahí sale todo lo que molesta —los portamodelo, el
// reinicio, la instalación por arnés, que `via: subagente` sea una sugerencia
// que nadie puede verificar—. Si `sf` marca, esas seis no se simplifican:
// DEJAN DE EXISTIR (headless.md §1).
//
// Lo que NO cambia: las compuertas siguen en `sf`, las paradas siguen siendo de
// Javier, y `sf next` sigue siendo consulta pura.
//
// ────────────────────────────────────────────────────────────────────────────
// LA REGLA QUE ORDENA ESTE ARCHIVO
// ────────────────────────────────────────────────────────────────────────────
//
//	sf no elige nada. Arma la línea con lo que el catálogo ya dice.
//
// El id, el esfuerzo y el permiso salen del catálogo, que llenó Javier con
// `sf install` (headless.md §4). Acá lo único que se sabe es CÓMO SE ESCRIBE
// cada cosa en cada arnés, y eso son tres filas medidas ejecutando el
// 2026-08-31 — no leídas de una doc. Las dos veces que en este proyecto se
// adivinó un flag, se adivinó mal.
package lanzar

import (
	"errors"
	"fmt"
	"strings"
)

// Pedido es todo lo que hace falta para armar una corrida.
type Pedido struct {
	Harness string

	// Modelo es el id tal cual sale del catálogo. sf NUNCA lo traduce (R3): un
	// id sólo significa algo en su proveedor.
	Modelo string

	// Esfuerzo es opcional de verdad. Vacío significa "el que traiga el modelo",
	// que NO es lo mismo que "bajo", y por eso vacío no emite ningún flag.
	Esfuerzo string

	// Raiz es la carpeta del proyecto. Va como flag donde el arnés tiene uno, y
	// SIEMPRE como `cmd.Dir` — Command Code no tiene flag y corre en la cwd.
	Raiz string

	Prompt string

	// Yolo pasa `--dangerously-skip-permissions` en Command Code.
	//
	// Sale del catálogo porque LO DECIDE JAVIER, no sf: es la opción B de
	// headless.md §9. Medido cinco veces el 2026-08-31 —con y sin
	// `.commandcode/settings.json`, con `--tools-all`, con `auto-accept`, con
	// reglas `allow` para `write_file` y `shell_command`— y en `-p` Command Code
	// no escribe ni ejecuta sin esto. No hay una puerta más angosta.
	Yolo bool
}

// arneses es CÓMO SE ESCRIBE cada cosa en cada arnés.
//
// Es la única tabla del paquete y es lo único que hay que tocar el día que
// aparezca un cuarto. No vive en el catálogo a propósito: el catálogo es de
// Javier, esta tabla es del programa.
var arneses = map[string]func(Pedido) []string{
	// `--verbose` no es decoración: sin él, `stream-json` no emite el stream.
	// `bypassPermissions` es el único de los seis modos que escribe y ejecuta.
	"claude-code": func(p Pedido) []string {
		l := []string{"claude", "-p", p.Prompt, "--model", p.Modelo}
		l = conEsfuerzo(l, "--effort", p.Esfuerzo)
		return append(l,
			"--output-format", "stream-json", "--verbose",
			"--permission-mode", "bypassPermissions",
			"--add-dir", p.Raiz)
	},

	// Headless en opencode es el subcomando `run` y NO `-p`: ahí `-p` es
	// `--password`. Un `sf lanzar` que asuma `-p` en los tres se rompe en el
	// primer intento. Y el esfuerzo se llama `--variant`.
	//
	// No lleva `--auto`: opencode escribe y ejecuta permisivo sin configurar
	// nada, medido dos veces.
	"opencode": func(p Pedido) []string {
		l := []string{"opencode", "run", "--dir", p.Raiz, "-m", p.Modelo}
		l = conEsfuerzo(l, "--variant", p.Esfuerzo)
		return append(l, "--format", "json", p.Prompt)
	},

	// Command Code no tiene flag de carpeta: corre en la cwd, así que quien
	// ejecute TIENE que poner `cmd.Dir`. Su `--output-format json` es NDJSON con
	// una línea de resultado al final.
	"commandcode": func(p Pedido) []string {
		l := []string{"cmd", "-p", p.Prompt, "-m", p.Modelo}
		l = conEsfuerzo(l, "--effort", p.Esfuerzo)
		l = append(l, "--output-format", "json")
		if p.Yolo {
			l = append(l, "--yolo")
		}
		return l
	},
}

// conEsfuerzo agrega el flag SÓLO si hay un esfuerzo declarado.
//
// Sin esto, un esfuerzo vacío emitiría `--effort ""` y le estaría pasando al
// arnés un valor que Javier no eligió. Es la misma distinción que el catálogo ya
// hace: vacío es "el del modelo", no "bajo".
func conEsfuerzo(l []string, flag, esfuerzo string) []string {
	if esfuerzo == "" {
		return l
	}
	return append(l, flag, esfuerzo)
}

// SabeLanzar dice si sf sabe correr ese arnés headless.
func SabeLanzar(harness string) bool {
	_, hay := arneses[harness]
	return hay
}

// Linea arma la línea de comando de ese arnés.
func Linea(p Pedido) ([]string, error) {
	armar, hay := arneses[p.Harness]
	if !hay {
		return nil, fmt.Errorf("no sé lanzar %q headless", p.Harness)
	}
	if p.Modelo == "" {
		// Sin `-m` el arnés corre con SU default, que es justo lo que este
		// proyecto viene evitando desde H1: el modelo lo elige el catálogo de
		// Javier, nunca el proveedor.
		return nil, errors.New("no hay modelo: el arnés correría con su default")
	}
	if p.Prompt == "" {
		return nil, errors.New("no hay prompt")
	}
	return armar(p), nil
}

// Mostrar arma la línea para que un humano la lea — es lo que imprime `--seco`.
//
// Entrecomilla lo que lo necesita para que se pueda pegar en una terminal tal
// cual. NO se usa para ejecutar: ejecutar toma el []string, que no pasa por
// ninguna shell y por lo tanto no tiene nada que escapar.
func Mostrar(linea []string) string {
	partes := make([]string, len(linea))
	for i, a := range linea {
		if strings.ContainsAny(a, " \t\n'\"$`\\") {
			partes[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
			continue
		}
		partes[i] = a
	}
	return strings.Join(partes, " ")
}
