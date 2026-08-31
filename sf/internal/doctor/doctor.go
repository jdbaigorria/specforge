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

	"github.com/jdbaigorria/specforge/sf/internal/andamio"
	"github.com/jdbaigorria/specforge/sf/internal/comandos"
	"github.com/jdbaigorria/specforge/sf/internal/constitucion"
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
	Perfiles []Perfil
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

	// TestRequiere es lo que la constitución declara que el `test_cmd` necesita
	// para poder correr: postgres, docker, redis. Vacío es lo normal.
	TestRequiere []string
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

	// El harness es el que está EN USO, que es lo que `sf next` va a usar para
	// resolver el modelo. Reportar otra cosa sería un diagnóstico que no
	// describe al binario que corre — y es justo lo que hacía esta función
	// antes: leía el puntero ESCRITO, así que adentro de opencode contestaba
	// "claude-code" porque era lo último que habías instalado.
	g, errCat := global.LeerPara(raiz)
	i.Harness = g.EnUso()
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

	// El harness desconocido pasó de ⚠ a ✗, y el motivo es que dejó de ser
	// cosmético: desde H2 el catálogo está INDEXADO POR HARNESS, así que sin
	// saber cuál es no hay perfil que resolver y `sf next` no puede contestar
	// con qué se lanza nada.
	if i.Harness == "desconocido" || i.Harness == "" {
		i.Fallas = append(i.Fallas,
			"no sé en qué harness estás, y sin eso no puedo resolver ningún modelo. "+
				"`sf install --harness=<"+strings.Join(global.Harness, "|")+">` lo fija.")
	}

	// Los perfiles que la máquina PIDE tienen que estar declarados. Sin esto, el
	// primer `sf next` para — y enterarse acá es más barato que enterarse
	// cuando el orquestador ya arrancó el bucle.
	if errCat == nil {
		i.Perfiles = revisarPerfiles(raiz, g)
		for _, p := range i.Perfiles {
			if !p.Declarado {
				i.Fallas = append(i.Fallas, "el perfil `"+p.Nombre+"` no está declarado para "+
					i.Harness+": `sf model "+p.Nombre+" --alias <corto> --id <id> --via subagente`")
				continue
			}
			if p.SinPortamodelo {
				i.Fallas = append(i.Fallas, "el alias `"+p.Alias+"` no tiene su archivo de agente: "+
					"`sf next` va a devolver un `agente:` que "+i.Harness+" no conoce. Corré `sf install`")
			}
		}
	}
	// El orquestador viejo es invisible por todos lados menos acá: el bucle
	// sigue corriendo, y lo único que pasa es que ignora las instrucciones que
	// sf agregó después. Es como se perdió una hora con el `agente:`.
	for _, o := range andamio.OrquestadoresViejos(raiz) {
		i.Fallas = append(i.Fallas, o+" es de una versión anterior de sf y no conoce "+
			"todo lo que `sf next` contesta: `sf install --forzar` lo actualiza")
	}

	if !i.Proyecto.Andamiado {
		i.Avisos = append(i.Avisos,
			"acá no hay .docs/: es un proyecto sin `sf init`, y eso puede estar bien")
	}
	if i.Proyecto.Roto != "" {
		i.Fallas = append(i.Fallas, "el estado no se puede leer: "+i.Proyecto.Roto)
	}

	// H6: lo que el `test_cmd` necesita para poder correr.
	//
	// Es AVISO y no falla, y la distinción importa: `test_requiere: [postgres]`
	// puede estar declarado porque el CI lo levanta, y frenar el doctor de
	// Javier por algo que su CI resuelve sería frenar sobre una suposición. sf
	// no lo resuelve —no levanta contenedores— sólo dice lo que ve.
	for _, req := range i.Proyecto.TestRequiere {
		if _, err := exec.LookPath(req); err != nil {
			i.Avisos = append(i.Avisos, "la constitución declara `test_requiere: "+req+
				"` y no lo encuentro en el PATH. Los lotes que lo necesiten van a fallar acá.")
		}
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

// Perfil es el diagnóstico de UN perfil que la máquina pide.
type Perfil struct {
	Nombre string

	// Declarado es si el harness activo tiene algún modelo para este perfil.
	Declarado bool

	// Alias es el default del perfil, si hay.
	Alias string

	// SinPortamodelo es que el alias está declarado pero su archivo de agente
	// no existe — o sea que `sf next` va a nombrar un agente que el harness no
	// conoce, y eso falla en el momento más caro: cuando ya lanzó.
	SinPortamodelo bool
}

// revisarPerfiles mira sólo los que la MÁQUINA pide.
//
// `mecanico` no está y no es un olvido: ningún estado lo devuelve —lo nombran
// los sfx-*, que están fuera de los nueve— así que exigirlo sería frenar por
// algo que el bucle no va a necesitar nunca.
func revisarPerfiles(raiz string, g *global.Config) []Perfil {
	pedidos := []string{global.Razonar, global.Construir}
	faltantes := andamio.PortamodelosQueFaltan(raiz, g)

	var ps []Perfil
	for _, n := range pedidos {
		p := Perfil{Nombre: n}
		if m, hay := g.Default(n); hay {
			p.Declarado, p.Alias = true, m.Alias
			p.SinPortamodelo = slices.Contains(faltantes, m.Alias)
		}
		ps = append(ps, p)
	}
	return ps
}

// raicesDeSkills son los lugares donde un harness deja los skills.
//
// Ninguno es adivinado. Los cuatro de Claude Code salen de mirar una instalación
// real; los de los otros dos salen de MEDIRLOS (sonda del 29/08, con dos skills
// de sonda y tokens aleatorios):
//
//	~/.claude/skills/<n>/                        claude-code · opencode ✅
//	~/.claude/plugins/cache/*/*/*/skills/<n>/    el plugin instalado
//	~/.claude/plugins/marketplaces/*/skills/<n>/ el repo clonado
//	<proyecto>/.claude/skills/<n>/               del proyecto
//	~/.config/opencode/skills/<n>/               opencode, global
//	<proyecto>/.opencode/skills/<n>/             opencode, del proyecto
//	~/.commandcode/skills/<n>/                   commandcode, global
//	<proyecto>/.commandcode/skills/<n>/          commandcode, del proyecto
//	~/.agents/skills/ y <proyecto>/.agents/skills/   opencode Y commandcode ✅
//
// EL DATO QUE IMPORTA, y que sólo se supo ejecutando: **Command Code NO lee
// `~/.claude/skills/`.** Los 18 skills instalados como symlinks ahí —que es la
// instalación real de hoy— son invisibles para él, y por eso `sf install` le
// escribe la ruta en su `settings.json` (ver andamio/harness.go).
//
// No existe una raíz que lean los tres: `.agents/skills/` cubre opencode y
// Command Code, `~/.claude/skills/` cubre Claude Code y opencode.
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
			filepath.Join(h, ".config", "opencode", "skills", "*"),
			filepath.Join(h, ".commandcode", "skills", "*"),
			filepath.Join(h, ".agents", "skills", "*"),
		)
	}
	return append(r,
		filepath.Join(raiz, ".claude", "skills", "*"),
		filepath.Join(raiz, ".opencode", "skills", "*"),
		filepath.Join(raiz, ".commandcode", "skills", "*"),
		filepath.Join(raiz, ".agents", "skills", "*"),
	)
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
		// Una constitución que todavía no existe no es un problema acá: el ⑧ la
		// escribe, y quejarse antes sería quejarse del orden del flujo.
		if c, err := constitucion.Leer(raiz); err == nil {
			p.TestRequiere = c.TestRequiere
		}
	}
	return p
}

// Sano dice si se puede usar. Es lo que decide el código de salida.
func (i Informe) Sano() bool { return len(i.Fallas) == 0 }
