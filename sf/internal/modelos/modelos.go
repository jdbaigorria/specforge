// Package modelos lista los modelos que cada arnés dice tener.
//
// ────────────────────────────────────────────────────────────────────────────
// PARA QUÉ EXISTE
// ────────────────────────────────────────────────────────────────────────────
//
// Para llenar el catálogo. Hoy, para declarar un modelo hay que ir a buscar el
// id a otra ventana —`sf doctor` literalmente te manda a `opencode models` o a
// `/model`— y tipearlo a mano. Esto lo trae acá.
//
// Su consumidor real es `sf install` (install-interactivo.md §5②), que arma el
// menú con esto, y a través del catálogo lo termina consumiendo `sf lanzar`:
// cuando el arnés principal delega un paso, el id, el esfuerzo y el `via` ya
// están escritos porque alguien los eligió de esta lista.
//
// ────────────────────────────────────────────────────────────────────────────
// LA REGLA QUE ORDENA TODO ESTE PAQUETE
// ────────────────────────────────────────────────────────────────────────────
//
//	el listado es una COMODIDAD, nunca el mecanismo.
//
// Los dos arneses que listan imprimen para humanos: encabezados, secciones,
// columnas alineadas con espacios, un pie con un link. Ninguno ofrece JSON
// —probado: el `--json` de Command Code se ignora—. O sea que este parser se va
// a romper el día que cambien el formato, y se va a romper en silencio.
//
// La mitigación NO es parsear mejor. Es que tipear el id a mano siga siendo un
// camino de primera clase: si `sf models` falla, quien llama pide el id y sigue.
// Si sf queda inservible porque un arnés cambió una tabla, el diseño está mal.
//
// Por eso también: `Listar` prefiere FALLAR a devolver una lista dudosa. Una
// lista vacía o con basura adentro es peor que un error, porque el que la lee es
// alguien eligiendo con qué modelo va a correr un paso.
//
// ────────────────────────────────────────────────────────────────────────────
// Y POR QUÉ NO HAY UNA TABLA DE MODELOS ACÁ ADENTRO
// ────────────────────────────────────────────────────────────────────────────
//
// Claude Code no lista sus modelos. La tentación es escribirle los alias a mano
// —son cuatro— y quedar tres por tres. Es exactamente la tabla que se pudre que
// `global.Modelo` ya se prohíbe para el esfuerzo: "cuáles son válidos depende del
// modelo y del proveedor". Así que este paquete dice que no sabe (ErrSinListado)
// y el que llama ofrece pegar el id.
package modelos

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

// Modelo es una fila del listado de un arnés, ya normalizada.
//
// Los dos campos calzan con `global.Modelo`: el `ID` es el id, y `Para` es el
// campo `para` — la opinión del proveedor sobre su propio modelo, que sf
// TRANSPORTA y no interpreta (R3). "long-horizon coding" no lo dedujo sf.
type Modelo struct {
	ID string

	// Para es la descripción que da el arnés, textual. Vacía si no da ninguna.
	Para string
}

// ErrSinListado es que ese arnés no tiene forma de enumerar sus modelos.
//
// NO es una falla: es un hecho del mundo, y se distingue a propósito de "ese
// arnés no existe". El primero se contesta ofreciendo pegar el id; el segundo es
// un error de quien llamó.
var ErrSinListado = errors.New("este arnés no lista sus modelos")

// ErrArnesDesconocido es que sf no conoce ese arnés, y NO es lo mismo.
//
// La diferencia importa afuera: "claude-code no lista" se contesta ofreciendo
// pegar el id a mano, y "emacs no existe" se contesta diciendo cuáles sí. Darle
// el camino manual a un typo es un consejo inútil.
var ErrArnesDesconocido = errors.New("no conozco ese arnés")

// Espera es cuánto se le da a un arnés para contestar el listado.
//
// Existe porque un arnés puede colgarse callado: medido el 2026-08-31, una
// corrida contra un modelo gratis de opencode estuvo 7m40 sin emitir un byte.
// Eso fue lanzando trabajo y no listando, pero la lección es la misma — un
// comando ajeno sin tope es un cuelgue esperando.
const Espera = 30 * time.Second

// listados es cómo se le pregunta a cada arnés, y con qué se lee la respuesta.
//
// claude-code NO está acá, y esa ausencia ES el dato: `Listar` la traduce a
// ErrSinListado. Ver el encabezado del paquete.
var listados = map[string]struct {
	binario string
	args    []string
	parsear func(string) []Modelo
}{
	"opencode":    {"opencode", []string{"models"}, parsearOpencode},
	"commandcode": {"cmd", []string{"--list-models"}, parsearCommandCode},
}

// SabeListar dice si de ese arnés se puede sacar un listado.
//
// Es para que quien arma un menú pueda ofrecer "elegí de la lista" o "pegá el
// id" ANTES de correr nada, en vez de descubrirlo por un error.
func SabeListar(harness string) bool {
	_, ok := listados[harness]
	return ok
}

// correr ejecuta el listado. Es una variable para que los tests no dependan de
// tener los tres arneses instalados en la máquina donde corre el CI.
var correr = func(ctx context.Context, binario string, args ...string) ([]byte, error) {
	// Output y no CombinedOutput, y es deliberado: los plugins del arnés
	// escriben en stderr. `opencode models` con el plugin de ICM cargado imprime
	// "[icm] plugin loaded (icm 0.10.50)" antes de los ids, y por stdout no viene.
	// Los parsers igual descartan lo que no tiene forma de id, porque nada impide
	// que otro plugin escriba en stdout.
	return exec.CommandContext(ctx, binario, args...).Output()
}

// Listar corre el listado de ese arnés y lo devuelve normalizado.
func Listar(harness string) ([]Modelo, error) {
	l, ok := listados[harness]
	if !ok {
		if harness == "claude-code" {
			return nil, ErrSinListado
		}
		return nil, fmt.Errorf("%w: %q", ErrArnesDesconocido, harness)
	}

	ctx, cancelar := context.WithTimeout(context.Background(), Espera)
	defer cancelar()

	salida, err := correr(ctx, l.binario, l.args...)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("`%s %s` no contestó en %s",
				l.binario, strings.Join(l.args, " "), Espera)
		}
		return nil, fmt.Errorf("`%s %s`: %w", l.binario, strings.Join(l.args, " "), err)
	}

	ms := l.parsear(string(salida))
	if len(ms) == 0 {
		// Cero modelos no es un listado vacío: los dos arneses que listan tienen
		// cientos. Es que la salida cambió de forma y este parser ya no la
		// entiende. Decirlo es lo único honesto — devolver `nil, nil` haría que
		// un menú se dibuje vacío como si el arnés no tuviera modelos.
		return nil, fmt.Errorf("`%s %s` contestó, pero no reconocí ningún modelo "+
			"en lo que devolvió: puede haber cambiado el formato",
			l.binario, strings.Join(l.args, " "))
	}
	return ms, nil
}

// idOpencode es un id de opencode: sin espacios y con al menos una barra.
//
// La barra no es un capricho: los 397 ids que devuelve el listado son
// `proveedor/modelo` —algunos con dos barras, `openrouter/z-ai/glm-5.3`— y
// exigirla es lo que deja afuera cualquier línea de ruido de un plugin.
var idOpencode = regexp.MustCompile(`^[^\s]+/[^\s]+$`)

// parsearOpencode lee la salida de `opencode models`: un id por línea, sin más.
func parsearOpencode(salida string) []Modelo {
	var ms []Modelo
	for _, l := range strings.Split(salida, "\n") {
		l = strings.TrimSpace(l)
		if idOpencode.MatchString(l) {
			ms = append(ms, Modelo{ID: l})
		}
	}
	return ms
}

// dosColumnas parte una línea de Command Code en id y descripción.
//
// El separador son DOS O MÁS espacios, que es como alinea sus columnas. Un solo
// espacio no alcanza: las descripciones tienen espacios adentro.
var dosColumnas = regexp.MustCompile(`^(\S+)\s{2,}(\S.*)$`)

// parsearCommandCode lee la salida de `cmd --list-models`.
//
// El formato trae tres cosas que no son modelos y hay que descartar:
//
//	Available models  ·  63 models      el encabezado
//	Open Source                         una sección por proveedor
//	Docs:  https://…                    el pie
//
// El encabezado y las secciones caen solos: su primera columna tiene un espacio
// adentro ("Available models", "Open Source") y `dosColumnas` exige que no lo
// tenga. El pie NO cae solo —"Docs:" es una palabra sola seguida de dos
// espacios, o sea la misma forma que un modelo— y es el único falso positivo del
// formato, así que se descarta a mano.
//
// La otra tentación acá sería exigir la barra, como en opencode. Sería un error:
// `claude-sonnet-5` y `gpt-5.6-sol` son ids válidos de Command Code, y exigirla
// se comería catorce modelos reales sin decir una palabra.
func parsearCommandCode(salida string) []Modelo {
	var ms []Modelo
	for _, l := range strings.Split(salida, "\n") {
		p := dosColumnas.FindStringSubmatch(strings.TrimRight(l, " \t\r"))
		if p == nil {
			continue
		}
		id, para := p[1], strings.TrimSpace(p[2])

		// El pie. Un id no termina en ":" y una descripción no es una URL.
		if strings.HasSuffix(id, ":") || strings.HasPrefix(para, "http") {
			continue
		}
		ms = append(ms, Modelo{ID: id, Para: para})
	}
	return ms
}
