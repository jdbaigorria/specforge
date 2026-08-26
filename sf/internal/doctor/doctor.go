// Package doctor contesta una sola pregunta: ¿esta instalación sirve?
//
// ────────────────────────────────────────────────────────────────────────────
// DE DÓNDE SALIÓ
// ────────────────────────────────────────────────────────────────────────────
//
// De una corrida real que no arrancó. La máquina estaba entera y auditada, y
// aun así el punta a punta era imposible por tres cosas que NINGÚN test podía
// ver, porque ninguna vive adentro del repo:
//
//	el `sf` del PATH era el viejo        se compiló, no se instaló
//	los skills no estaban instalados     el binario los nombra y no existen
//	un skill se contradecía              y trababa el ⑥
//
// Las dos primeras son mecánicas y este paquete las agarra. La tercera no: es
// una contradicción adentro de un archivo y la agarra el CI, no el doctor.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ SE INSTALA EN DOS MITADES, Y POR QUÉ ESO HAY QUE VIGILARLO
// ────────────────────────────────────────────────────────────────────────────
//
// El binario y los skills llegan por caminos distintos —`install.sh` baja uno,
// el plugin del harness trae los otros— y eso no es un defecto del instalador:
// es que son cosas distintas. Un ejecutable compilado no entra en un plugin de
// Markdown, y 18 archivos de Markdown no necesitan compilarse.
//
// El precio es que pueden quedar desalineados, y ahí aparece un error de una
// familia peor que la de los diez de la auditoría: las dos mitades están BIEN
// CADA UNA POR SU LADO. No hay archivo que leer donde se vea el problema.
//
// ────────────────────────────────────────────────────────────────────────────
// Y SE COMPRUEBA CON UN HECHO, NO CON DOS NÚMEROS DE VERSIÓN
// ────────────────────────────────────────────────────────────────────────────
//
// La salida obvia sería que el binario y los skills lleven versión y se
// comparen. Es peor de lo que parece:
//
//	el que instala desde el repo no tiene versión — trabaja sobre `main`
//	el que edita un skill a mano tampoco, y hace bien: son suyos
//	y dos versiones distintas NO prueban que algo esté roto
//
// O sea que avisaría de más justo a los que más lo usan, y se aprende a
// ignorarlo. Lo que sí es un hecho comprobable es más chico y más útil:
//
//	¿los comandos que nombran los skills instalados existen en ESTE binario?
//
// Eso no opina sobre versiones: si un skill dice `sf loquesea` y este binario
// no lo tiene, el bucle SE VA A TRABAR, y el que lo iba a descubrir es un
// agente a mitad de camino. Es la misma regla que las compuertas (R3): se frena
// sobre un hecho, no sobre un parecido.
package doctor

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/comandos"
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/maquina"
)

// Informe es todo lo que el doctor averiguó.
//
// Se separa de cómo se imprime por el mismo motivo que `vista`: lo que un test
// quiere afirmar es el hallazgo, no el dibujo.
type Informe struct {
	Binario  Binario
	Skills   []Skill
	Harness  string
	Proyecto Proyecto

	// Fallas son las que impiden usar sf. Avisos, las que no.
	Fallas []string
	Avisos []string
}

// Binario es qué ejecutable está corriendo y cuál encontraría el harness.
type Binario struct {
	Version string

	// Corriendo es el ejecutable de ESTA corrida.
	Corriendo string

	// EnPath es lo que devuelve buscar "sf" en el PATH, que es lo que va a
	// ejecutar el agente cuando el skill le diga `sf next`.
	EnPath string

	// Sombra es el caso que motivó todo esto: el que corre el doctor y el que
	// va a correr el bucle no son el mismo archivo.
	Sombra bool
}

// Skill es uno de los nueve de la máquina, y dónde se encontró.
type Skill struct {
	Nombre string

	// Ruta es dónde está instalado, vacía si no está.
	Ruta string

	// Desconocidos son los `sf <cmd>` que el skill nombra y este binario no
	// tiene. Es la comprobación de desalineación.
	Desconocidos []string
}

// Instalado dice si el harness lo va a encontrar.
func (s Skill) Instalado() bool { return s.Ruta != "" }

// Proyecto es el estado del directorio desde el que se corrió.
type Proyecto struct {
	Raiz string

	// Andamiado es si existe .docs/ — o sea, si alguien corrió `sf init`.
	Andamiado bool

	// Roto es un estado.json que no se puede leer. DÓNDE está la máquina no se
	// contesta acá: eso es `sf status`, y derivarlo de nuevo sería una segunda
	// copia de la lógica del ⑥–⑩ (R6).
	Roto string
}

// reComando caza los `sf <algo>` que un skill nombra.
//
// Sólo minúsculas y guiones: lo que sigue a `sf` en una línea de shell. Los
// `sf done --msg`, `sf reject "motivo"` y `sf take <feature>` quedan reducidos a
// su primera palabra, que es la única que el despacho mira.
var reComando = regexp.MustCompile(`\bsf ([a-z][a-z-]*)`)

// Revisar arma el informe. `raiz` es desde dónde se corrió.
func Revisar(raiz, version string) Informe {
	var i Informe
	i.Binario = revisarBinario(version)
	i.Harness = global.DetectarHarness()
	i.Skills = revisarSkills(raiz)
	i.Proyecto = revisarProyecto(raiz)

	if i.Binario.Sombra {
		i.Fallas = append(i.Fallas,
			"el `sf` del PATH no es éste: el agente va a correr "+i.Binario.EnPath)
	}
	if i.Binario.EnPath == "" {
		i.Fallas = append(i.Fallas,
			"`sf` no está en el PATH: el agente no lo va a poder llamar")
	}

	var faltan []string
	for _, s := range i.Skills {
		if !s.Instalado() {
			faltan = append(faltan, s.Nombre)
			continue
		}
		for _, c := range s.Desconocidos {
			i.Fallas = append(i.Fallas,
				s.Nombre+" nombra `sf "+c+"`, y este binario no lo tiene")
		}
	}
	if len(faltan) > 0 {
		// Uno solo alcanza para trabar el bucle: `sf next` va a devolver un
		// skill que el harness no encuentra, y ahí se termina.
		i.Fallas = append(i.Fallas,
			"faltan instalar "+plural(len(faltan))+": "+strings.Join(faltan, " · "))
	}

	if i.Harness == "desconocido" {
		i.Avisos = append(i.Avisos,
			"no reconocí el harness. `sf install --harness=<nombre>` lo fija.")
	}
	if !i.Proyecto.Andamiado {
		i.Avisos = append(i.Avisos,
			"acá no hay .docs/: es un proyecto sin `sf init`, y eso puede estar bien")
	}
	if i.Proyecto.Roto != "" {
		i.Fallas = append(i.Fallas, "el estado no se puede leer: "+i.Proyecto.Roto)
	}
	return i
}

func plural(n int) string {
	if n == 1 {
		return "1 skill"
	}
	return strconv.Itoa(n) + " skills"
}

func revisarBinario(version string) Binario {
	b := Binario{Version: version}

	if e, err := os.Executable(); err == nil {
		if r, err := filepath.EvalSymlinks(e); err == nil {
			e = r
		}
		b.Corriendo = e
	}
	if p, err := exec.LookPath("sf"); err == nil {
		if r, err := filepath.EvalSymlinks(p); err == nil {
			p = r
		}
		b.EnPath = p
	}

	// La comparación es entre los dos resueltos: un symlink de ~/.local/bin/sf
	// al binario que está corriendo NO es una sombra, es una instalación.
	b.Sombra = b.EnPath != "" && b.Corriendo != "" && b.EnPath != b.Corriendo
	return b
}

// raicesDeSkills son los lugares donde un harness deja los skills.
//
// Son cuatro y ninguno es adivinado: los tres primeros salen de mirar una
// instalación real de Claude Code, y el cuarto es el proyecto.
//
//	~/.claude/skills/<n>/                                    a mano o symlink
//	~/.claude/plugins/cache/*/*/*/skills/<n>/                el plugin instalado
//	~/.claude/plugins/marketplaces/*/skills/<n>/             el repo clonado
//	<proyecto>/.claude/skills/<n>/                           del proyecto
//
// Se busca en todos y se reporta el PRIMERO que aparece, con su ruta. Que la
// ruta se muestre no es adorno: la mitad de los problemas de instalación se
// entienden viendo de dónde salió el archivo.
func raicesDeSkills(raiz string) []string {
	var r []string
	if h, err := os.UserHomeDir(); err == nil {
		r = append(r,
			filepath.Join(h, ".claude", "skills", "*"),
			filepath.Join(h, ".claude", "plugins", "cache", "*", "*", "*", "skills", "*"),
			filepath.Join(h, ".claude", "plugins", "marketplaces", "*", "skills", "*"),
		)
	}
	return append(r, filepath.Join(raiz, ".claude", "skills", "*"))
}

func revisarSkills(raiz string) []Skill {
	// Un solo glob por raíz y después se busca en el mapa: nueve skills por
	// cuatro raíces serían 36 accesos a disco para contestar lo mismo.
	donde := map[string]string{}
	for _, patron := range raicesDeSkills(raiz) {
		m, _ := filepath.Glob(patron)
		sort.Strings(m)
		for _, d := range m {
			n := filepath.Base(d)
			if _, ya := donde[n]; ya {
				continue
			}
			if _, err := os.Stat(filepath.Join(d, "SKILL.md")); err == nil {
				donde[n] = d
			}
		}
	}

	var out []Skill
	for _, n := range maquina.SkillsDeEstado() {
		s := Skill{Nombre: n, Ruta: donde[n]}
		if s.Ruta != "" {
			s.Desconocidos = comandosDesconocidos(filepath.Join(s.Ruta, "SKILL.md"))
		}
		out = append(out, s)
	}
	return out
}

// comandosDesconocidos lee un skill y devuelve los `sf <cmd>` que no existen.
func comandosDesconocidos(ruta string) []string {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return nil
	}

	var out []string
	for _, m := range reComando.FindAllStringSubmatch(soloCodigo(string(b)), -1) {
		c := m[1]
		// `sf lote start` es de dos palabras y el regex sólo trae la primera.
		// Distinguirlo importa: `lote` suelto no existe y `lote start` sí.
		if c == "lote" {
			c = "lote start"
		}
		if comandos.Existe(c) || slices.Contains(out, c) {
			continue
		}
		out = append(out, c)
	}
	sort.Strings(out)
	return out
}

// soloCodigo se queda con los bloques cercados y los backticks sueltos.
//
// ────────────────────────────────────────────────────────────────────────────
// ESTO LO ENCONTRÓ EL DOCTOR EN SU PRIMERA CORRIDA, SOBRE SÍ MISMO
// ────────────────────────────────────────────────────────────────────────────
//
// La primera versión buscaba `sf <algo>` en el archivo entero y filtraba con una
// heurística —"que esté entre backticks, o solo al principio de una línea"—. La
// segunda mitad de esa heurística estaba mal, y sfp-backlog lo demostró al
// instante:
//
//	sf does not verify the judgment. It verifies THAT THE JUDGMENT HAPPENED.
//
// Es prosa, arranca la línea, y el doctor reportó que el skill nombraba un
// comando inexistente llamado `does`. Un chequeo que acusa a un archivo correcto
// es peor que no tenerlo: se aprende a ignorarlo, y el día que acierte también
// se lo va a ignorar.
//
// Los skills están escritos en inglés y `sf` es el sujeto de media docena de
// oraciones —"sf checks", "sf counts", "sf will not let"—, así que la heurística
// no se puede afinar: hay que cambiar la pregunta. Y la correcta es exacta, no
// aproximada:
//
//	un comando que el agente va a EJECUTAR está escrito COMO CÓDIGO.
//
// Fuera del código, `sf` es una palabra. Adentro, es el programa.
//
// ────────────────────────────────────────────────────────────────────────────
// Y "ADENTRO DE UN ``` " TAMPOCO ALCANZA
// ────────────────────────────────────────────────────────────────────────────
//
// La frase de sfp-backlog está adentro de un bloque cercado. No es shell: es una
// cita, puesta en un ``` para que se lea como un cartel. Los skills usan el ```
// pelado para eso —los carteles del 🛑, los diagramas, las citas— y ```bash para
// lo que se ejecuta. O sea que el lenguaje del bloque ES la marca, y estaba ahí
// desde el principio.
//
// Esto deja un falso NEGATIVO posible: un comando de verdad escrito en un ```
// pelado no se vería. Es la dirección correcta para equivocarse — un chequeo que
// se calla de más se sigue mirando; uno que acusa de más se deja de mirar.
func soloCodigo(texto string) string {
	var b strings.Builder
	dentro := false

	for _, l := range strings.Split(texto, "\n") {
		if t := strings.TrimSpace(l); strings.HasPrefix(t, "```") {
			if dentro {
				dentro = false
			} else {
				dentro = esShell(strings.TrimPrefix(t, "```"))
			}
			continue
		}
		if dentro {
			b.WriteString(l)
			b.WriteString("\n")
			continue
		}
		// Afuera, sólo lo que esté entre backticks. Los pares se toman en
		// orden: un backtick impar suelto deja el resto de la línea afuera, que
		// es la interpretación conservadora — de más no acusa.
		partes := strings.Split(l, "`")
		for i := 1; i < len(partes); i += 2 {
			b.WriteString(partes[i])
			b.WriteString("\n")
		}
	}
	return b.String()
}

// esShell dice si el info string de un bloque cercado es una shell.
func esShell(lang string) bool {
	return slices.Contains(
		[]string{"bash", "sh", "shell", "zsh", "console"},
		strings.ToLower(strings.TrimSpace(lang)))
}

func revisarProyecto(raiz string) Proyecto {
	p := Proyecto{Raiz: raiz}
	if _, err := os.Stat(filepath.Join(raiz, docs.Base)); err == nil {
		p.Andamiado = true
	}
	if p.Andamiado {
		if _, err := estado.Leer(raiz); err != nil && !errors.Is(err, estado.ErrNoHay) {
			p.Roto = err.Error()
		}
	}
	return p
}

// Sano dice si se puede usar. Es lo que decide el código de salida.
func (i Informe) Sano() bool { return len(i.Fallas) == 0 }
