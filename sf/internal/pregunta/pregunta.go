// Package pregunta es la pregunta de `sf install`: el ÚNICO lugar donde sf conversa.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ ACÁ SE PUEDE Y EN NINGÚN OTRO LADO
// ────────────────────────────────────────────────────────────────────────────
//
// `sf` nunca conversa, y es la regla que lo hace usable por un agente: `sf next`
// es consulta pura, `sf done` contesta y se va, ninguno espera a nadie.
//
// `sf install` es la excepción, y la excepción tiene un fundamento que ya
// existía: NO ES DE LA MÁQUINA. No mira el estado ni lo mueve, no aparece en
// ningún trazado del bucle. Es lo que hace que SpecForge se pueda usar, y lo
// corre una persona, una vez.
//
// Y no hay que elegir entre conversar y no conversar, porque se puede SABER:
// la shell de un agente no tiene TTY y la de una persona sí (medido el
// 2026-08-31). Quien llama a este paquete ya hizo esa pregunta.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ UNA LISTA NUMERADA Y NO UNA PANTALLA CON FLECHITAS
// ────────────────────────────────────────────────────────────────────────────
//
// Decidido por Javier el 2026-08-31, y la razón no es estética.
//
// Una pantalla viva —cursor con flechas, marcar con espacio, filtrar mientras
// tipeás— pide MODO RAW: leer tecla por tecla en vez de renglón por renglón. En
// Go eso no está en la biblioteca estándar, así que sale `golang.org/x/term`, y
// `sf` tiene hoy UNA dependencia en total: sería duplicarla para dibujar un menú.
//
// Y hay algo peor que la dependencia: una pantalla raw no funciona por un pipe.
// Habría que escribir dos caminos, uno para persona y otro para tubo, y el
// segundo es el que no puede fallar nunca.
//
// Leyendo renglones hay UN SOLO camino. La diferencia de experiencia es una
// tecla: el enter después del filtro.
//
// ────────────────────────────────────────────────────────────────────────────
// LO QUE ESTE PAQUETE NO HACE
// ────────────────────────────────────────────────────────────────────────────
//
// No escribe nada. Devuelve lo que Javier contestó y quien llama decide qué
// hacer con eso. Es lo que lo deja testear con un guion de respuestas y un
// buffer, sin tocar el disco ni ejecutar un arnés.
package pregunta

import (
	"bufio"
	"fmt"
	"io"
	"slices"
	"strconv"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/modelos"
)

// Declaracion es un modelo que Javier eligió para un perfil de un arnés.
type Declaracion struct {
	Harness  string
	Perfil   string
	Alias    string
	ID       string
	Esfuerzo string
}

// Respuesta es todo lo que se contestó.
type Respuesta struct {
	Harnesses     []string
	Declaraciones []Declaracion

	// Permisos es harness → el permiso extra que Javier concedió. Ver
	// `preguntarPermiso`.
	Permisos map[string]string
}

// Listador es de dónde salen los modelos de un arnés.
//
// Es un parámetro y no una llamada directa a `modelos.Listar` para que el test
// no necesite tener los tres arneses instalados — y de paso deja probar el
// camino que importa: qué pasa cuando el listado NO anda.
type Listador func(harness string) ([]modelos.Modelo, error)

// aMostrar es cuántos modelos se listan antes de pedir que filtres.
//
// opencode devuelve 397: escupirlos todos sería inservible en una terminal.
const aMostrar = 12

// perfilesQueSePreguntan son los dos que la máquina pide.
//
// `mecanico` no está y no es un olvido: ningún estado lo pide
// (`maquina.perfilPorEstado`), lo nombran los `sfx-*` que están fuera de los
// nueve. Preguntarlo sería pedir una decisión que no frena nada.
var perfilesQueSePreguntan = []struct {
	nombre string
	para   string
}{
	{global.Razonar, "el que compara y juzga — el ⑫ elige entre tres opciones, el ㉑ busca lo que no está"},
	{global.Construir, "el que escribe el código contra un plan que ya existe — todo el resto"},
}

// Preguntar conversa y devuelve lo contestado.
//
// Nunca devuelve error por quedarse sin entrada: un Ctrl-D o un pipe que se
// acabó devuelven lo que se contestó HASTA AHÍ. Tirar respuestas que ya costó
// dar sería lo peor que puede hacer con ellas.
func Preguntar(w io.Writer, r io.Reader, instalados []string, listar Listador) (*Respuesta, error) {
	p := &sesion{w: w, sc: bufio.NewScanner(r), listar: listar}
	res := &Respuesta{}

	// Sin detección se ofrecen los tres igual. Puede que estén y no se los haya
	// visto, o que Javier quiera dejar preparado uno que va a instalar después:
	// la detección es una AYUDA, no una compuerta.
	ofrecidos := instalados
	if len(ofrecidos) == 0 {
		ofrecidos = global.Harness
	}

	res.Harnesses = p.cualesArneses(ofrecidos, len(instalados) > 0)
	res.Permisos = map[string]string{}
	for _, h := range res.Harnesses {
		if permiso := p.cualPermiso(h); permiso != "" {
			res.Permisos[h] = permiso
		}
		res.Declaraciones = append(res.Declaraciones, p.modelosDe(h)...)
	}
	return res, nil
}

// cualPermiso pregunta por el permiso extra de un arnés, si tiene sentido.
//
// ────────────────────────────────────────────────────────────────────────────
// SÓLO COMMAND CODE, Y SÓLO PORQUE NO HAY OTRA PUERTA
// ────────────────────────────────────────────────────────────────────────────
//
// En modo headless Command Code no escribe archivos ni corre comandos salvo con
// `--yolo`. Probado cinco veces el 2026-08-31 —con y sin su `settings.json`, con
// `--tools-all`, con `auto-accept`, con reglas `allow` para `write_file` y
// `shell_command`— y no hay puerta más angosta.
//
// A los otros dos NO se les pregunta: opencode escribe y ejecuta permisivo sin
// ningún flag, y Claude Code usa `bypassPermissions`. Preguntarlo ahí sería pedir
// una decisión que no cambia nada, y de más es lo peor que se puede pedir en una
// pregunta sobre seguridad.
//
// EL DEFAULT ES QUE NO. La respuesta que se da sin leer es el enter, así que el
// enter tiene que ser la conservadora.
func (p *sesion) cualPermiso(harness string) string {
	if harness != "commandcode" {
		return ""
	}
	p.di("\n── %s %s", harness, strings.Repeat("─", max(0, 60-len(harness))))
	p.di("  En headless, Command Code no escribe archivos ni corre comandos salvo")
	p.di("  con --yolo, que apaga TODOS sus chequeos de permisos. Sin eso, sf no le")
	p.di("  puede delegar pasos — como arnés principal lo seguís usando igual.")
	fmt.Fprint(p.w, "  ¿Se lo paso? (s/N): ")

	switch strings.ToLower(p.leer()) {
	case "s", "si", "sí", "y", "yes":
		return global.PermisoYolo
	}
	return ""
}

// sesion es el estado de una conversación: por dónde escribir y de dónde leer.
type sesion struct {
	w      io.Writer
	sc     *bufio.Scanner
	listar Listador
	fin    bool // se acabó la entrada: de acá en adelante todo es "enter"
}

// leer devuelve la próxima línea, o "" para siempre cuando se acabó.
func (p *sesion) leer() string {
	if p.fin || !p.sc.Scan() {
		p.fin = true
		return ""
	}
	return strings.TrimSpace(p.sc.Text())
}

func (p *sesion) di(formato string, a ...any) { fmt.Fprintf(p.w, formato+"\n", a...) }

func (p *sesion) cualesArneses(ofrecidos []string, detectados bool) []string {
	p.di("\n¿Qué arneses vas a usar acá?\n")
	for i, h := range ofrecidos {
		marca := ""
		if detectados {
			marca = "  (instalado)"
		}
		p.di("  %d  %s%s", i+1, h, marca)
	}
	p.di("\n  Números separados por coma, o enter para todos.")
	fmt.Fprint(p.w, "> ")

	elegidos := numeros(p.leer(), len(ofrecidos))
	if len(elegidos) == 0 {
		return slices.Clone(ofrecidos)
	}
	var out []string
	for _, i := range elegidos {
		out = append(out, ofrecidos[i])
	}
	return out
}

func (p *sesion) modelosDe(h string) []Declaracion {
	p.di("\n── %s %s", h, strings.Repeat("─", max(0, 60-len(h))))

	// El listado puede no existir (Claude Code) o haberse roto (cambió el
	// formato, el binario no está). En los dos casos la pregunta SIGUE: lo único
	// que se pierde es el menú, y pegar el id es un camino de primera clase.
	catalogo, err := p.listar(h)
	if err != nil {
		p.di("   (no pude listar los modelos de %s — %v)", h, err)
	}

	var out []Declaracion
	for _, perfil := range perfilesQueSePreguntan {
		p.di("\n%s — %s", perfil.nombre, perfil.para)

		id := p.cualModelo(catalogo)
		if id == "" {
			continue
		}
		alias := p.cualAlias()
		if alias == "" {
			continue
		}
		out = append(out, Declaracion{
			Harness: h, Perfil: perfil.nombre, ID: id, Alias: alias,
			Esfuerzo: p.cualEsfuerzo(),
		})
	}
	return out
}

// cualModelo pide un modelo y devuelve su id, o "" si se saltea el perfil.
//
// Un número elige de lo mostrado. Un texto FILTRA y vuelve a preguntar. Y un
// texto que no filtra nada se toma como un id pegado — que es la regla de §8 de
// la spec: el listado es una comodidad, nunca el mecanismo, así que un id que el
// arnés todavía no enumera tiene que poder declararse igual.
func (p *sesion) cualModelo(catalogo []modelos.Modelo) string {
	visibles := catalogo

	for {
		switch {
		case len(catalogo) == 0:
			fmt.Fprint(p.w, "  Pegá el id (enter lo saltea): ")

		case len(visibles) == 0:
			// El filtro no encontró nada. No es un callejón: lo tipeado puede ser
			// el id que se quiere, y volver a la lista entera sería hacerle
			// repetir el filtro a alguien que ya sabe lo que busca.
			p.di("  (nada con ese texto)")
			visibles = catalogo
			p.mostrar(visibles, len(catalogo))
			fmt.Fprint(p.w, "> ")

		default:
			p.mostrar(visibles, len(catalogo))
			fmt.Fprint(p.w, "> ")
		}

		linea := p.leer()
		if linea == "" {
			return ""
		}
		if n, err := strconv.Atoi(linea); err == nil && n >= 1 && n <= len(visibles) {
			return visibles[n-1].ID
		}

		// Un id exacto es un pegado, aunque esté en la lista: si tipeaste el id
		// entero, no querías filtrar.
		if i := slices.IndexFunc(catalogo, func(m modelos.Modelo) bool { return m.ID == linea }); i >= 0 {
			return catalogo[i].ID
		}
		if len(catalogo) == 0 {
			return linea // no hay con qué filtrar: es un id y listo
		}

		if f := filtrar(catalogo, linea); len(f) > 0 {
			visibles = f
			continue
		}
		// Filtró a cero. Lo tipeado es el id.
		return linea
	}
}

func (p *sesion) mostrar(visibles []modelos.Modelo, total int) {
	for i, m := range visibles {
		if i >= aMostrar {
			break
		}
		if m.Para == "" {
			p.di("  %2d  %s", i+1, m.ID)
			continue
		}
		p.di("  %2d  %-42s %s", i+1, m.ID, m.Para)
	}
	// Se dice cuántos hay SIEMPRE que no entren todos, porque elegir entre doce
	// creyendo que son doce, cuando son cuatrocientos, es elegir mal.
	if len(visibles) > aMostrar {
		p.di("      … %d más. Escribí para filtrar, o pegá un id.", len(visibles)-aMostrar)
	} else if total > len(visibles) {
		p.di("      (%d de %d — escribí para filtrar de nuevo)", len(visibles), total)
	}
}

// cualAlias pide el nombre corto, que lo pone Javier y no sf.
//
// Derivarlo del id —`nemotron-3-ultra-free` → `ultra`— sería sf opinando sobre
// cómo se llaman las cosas de él, y el alias es SU vocabulario: es lo que
// después viaja a tareas.json y lo que tipea en `sf model <alias>`.
// HAY UN SOLO LUGAR PARA SALTEAR UN PERFIL, y es la pregunta del modelo.
//
// Una vez que elegiste modelo, el alias es obligatorio: sin él el modelo no se
// puede nombrar desde tareas.json, y nombrarlo es para lo único que el alias
// existe. Dos puntos de escape para lo mismo harían que un enter de más tire la
// elección que acabás de hacer, sin decir nada.
func (p *sesion) cualAlias() string {
	for {
		fmt.Fprint(p.w, "  alias corto: ")
		a := p.leer()
		switch {
		case a == "":
			if p.fin {
				return "" // se acabó la entrada, no es que no quiso
			}
			p.di("  el alias es obligatorio: es cómo lo vas a nombrar en tareas.json.")
			p.di("  (para saltear el perfil, enter en la pregunta del modelo)")
		case global.EsPerfil(a):
			p.di("  \"%s\" se llama igual que un perfil. Elegí otro.", a)
		case strings.ContainsAny(a, " \t/"):
			p.di("  el alias es una palabra corta, sin espacios ni barras.")
		default:
			return a
		}
	}
}

// cualEsfuerzo pide el esfuerzo, que es opcional de verdad.
//
// Enter deja VACÍO, y vacío no es "bajo": es "el que traiga el modelo". La
// distinción ya vive en `global.Modelo.Esfuerzo` y acá no se puede perder.
func (p *sesion) cualEsfuerzo() string {
	fmt.Fprint(p.w, "  esfuerzo (enter = el que traiga el modelo): ")
	return p.leer()
}

// numeros lee "1,3" y devuelve los índices válidos, sin repetir y en orden.
func numeros(linea string, tope int) []int {
	var out []int
	for _, t := range strings.FieldsFunc(linea, func(r rune) bool { return r == ',' || r == ' ' }) {
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err != nil || n < 1 || n > tope {
			continue
		}
		if !slices.Contains(out, n-1) {
			out = append(out, n-1)
		}
	}
	slices.Sort(out)
	return out
}

// filtrar son los modelos cuyo id o descripción contienen el texto.
func filtrar(ms []modelos.Modelo, texto string) []modelos.Modelo {
	texto = strings.ToLower(texto)
	var out []modelos.Modelo
	for _, m := range ms {
		if strings.Contains(strings.ToLower(m.ID), texto) ||
			strings.Contains(strings.ToLower(m.Para), texto) {
			out = append(out, m)
		}
	}
	return out
}
