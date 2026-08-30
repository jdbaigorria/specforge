// Este archivo es lo que `sf install` escribe PARA EL HARNESS, y son tres cosas
// que se descubrieron ejecutando y no leyendo (sondas del 29/08).
//
// ────────────────────────────────────────────────────────────────────────────
// ① PERMISOS — el bucle es desatendido, y un prompt lo rompe
// ────────────────────────────────────────────────────────────────────────────
//
// El orquestador lanza, el subagente pide su sobre, trabaja y cierra con
// `sf done`. Un prompt de permisos en el medio no es una molestia: rompe la
// premisa. Y hasta ahora `sf install` —que existe para "hacer que SpecForge se
// pueda usar"— no tocaba la única cosa que lo hace usable fuera de Claude Code.
//
// Lo MEDIDO, que corrige lo que este archivo habría hecho por deducción:
//
//	opencode      corrió todo sin preguntar, SIN configuración. Su default es
//	              permisivo, así que lo que escribimos es un cinturón y no el
//	              arreglo: acota, no habilita.
//	commandcode   preguntó por todo, porque sin settings.json el baseline es
//	              `default` = "preguntá antes de cualquier cosa que cambie algo".
//
// Y el modo correcto para Command Code es `dont-ask` y NO `auto-accept`, que era
// lo primero que se propuso: su propia documentación aclara que auto-accept
// sigue preguntando por comandos de shell arbitrarios, que es todo lo que el
// bucle hace. `dont-ask` nunca pregunta: corre lo pre-aprobado y DENIEGA el
// resto.
//
//	Un bucle desatendido no puede colgarse esperando un enter.
//	`dont-ask` convierte el cuelgue en un fallo que `sf done` ve.
//
// El precio es real y está asumido: lo que falte en la lista no pregunta, falla.
// Por eso `sf init` la completa con el `test_cmd` cuando lo detecta, y por eso
// `sf doctor` comprueba que siga estando.
//
// ────────────────────────────────────────────────────────────────────────────
// ② LOS SKILLS — Command Code no lee ~/.claude/skills/
// ────────────────────────────────────────────────────────────────────────────
//
// Medido con dos skills de sonda y tokens aleatorios: opencode encuentra las de
// `~/.claude/skills/` y las de `.agents/skills/`; Command Code SÓLO las
// segundas. O sea que los 18 skills instalados como symlinks en `~/.claude/`
// —que es la instalación real de hoy— son invisibles para él.
//
// Se arregla con una línea de su `settings.json`, y eso también está medido:
// con `"skills": ["~/.claude/skills"]` y reiniciando, aparecieron las dos.
//
// No se mudan los archivos a `.agents/skills/` —que sería la raíz común a los
// otros dos— porque Claude Code no la lee: mudarlos arreglaría dos harness y
// rompería el tercero.
//
// ────────────────────────────────────────────────────────────────────────────
// ③ LOS PORTAMODELO — H8, la pieza sin la cual H2 es decoración
// ────────────────────────────────────────────────────────────────────────────
//
// En opencode y Command Code el modelo de un subagente sale del `model:` de su
// archivo de agente y NO se puede pisar al invocar. Medido: un agente con un id
// inexistente falló con "modelo desconocido" en los dos. En Claude Code sí se
// puede, y por eso ahí no se genera ninguno.
//
// El portamodelo es un archivo por ALIAS cuyo único trabajo es pinear un modelo.
// No lleva una sola instrucción de SpecForge: el skill sigue llegando por el
// prompt. Esa factorización es la que evita la explosión combinatoria —9 skills
// × N alias— y es lo que lo distingue de un subagente de verdad.
//
// Se generan TODOS de una y no bajo demanda, y eso también sale de una medición:
// los agentes se leen AL ARRANCAR. Generándolos todos, un `sf model <alias>` en
// medio del bucle apunta siempre a un archivo que ya existía y ya se cargó.
package andamio

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// permisos es la configuración que cada harness necesita para no preguntar.
//
// Se otorga lo que `sf` MISMO necesita y nada más. El `test_cmd` no entra acá y
// no es un olvido: `sf install` corre antes de `sf init`, así que todavía no hay
// constitución de donde leerlo. Lo agrega `sf init`, que es el que lo detecta.
var permisos = map[string]struct {
	Ruta      string
	Contenido string
}{
	"claude-code": {
		Ruta: filepath.Join(".claude", "settings.json"),
		Contenido: `{
  "permissions": {
    "allow": ["Bash(sf:*)", "Bash(git:*)"]
  }
}
`,
	},
	"opencode": {
		Ruta: "opencode.json",
		Contenido: `{
  "$schema": "https://opencode.ai/config.json",
  "permission": {
    "edit": "allow",
    "bash": { "*": "allow", "rm -rf *": "deny" }
  }
}
`,
	},
	"commandcode": {
		Ruta: filepath.Join(".commandcode", "settings.json"),
		Contenido: `{
  "permissions": {
    "defaultMode": "dont-ask",
    "allow": ["Shell(sf:*)", "Shell(git:*)"]
  },
  "skills": ["~/.claude/skills"]
}
`,
	},
}

// agentes es dónde deja sus definiciones de subagente cada harness.
//
// Claude Code no está y no es un olvido: ahí el modelo va como parámetro de la
// llamada, así que no hay nada que generar.
var agentes = map[string]string{
	"opencode":    filepath.Join(".opencode", "agents"),
	"commandcode": filepath.Join(".commandcode", "agents"),
}

// escribirPermisos deja la config del harness, sin pisar la que ya esté.
//
// No se mergea JSON a mano y no se va a hacer: parsear, fusionar y reescribir la
// configuración de otra herramienta es cómo se le rompe la suya a alguien. Si el
// archivo existe, se saltea y se dice — que es la misma regla que ya rige para
// `CLAUDE.md`.
func escribirPermisos(raiz, harness string, forzar bool, r *Resultado) error {
	p, hay := permisos[harness]
	if !hay {
		return nil
	}
	ruta := filepath.Join(raiz, p.Ruta)

	if _, err := os.Stat(ruta); err == nil && !forzar {
		r.Salteados = append(r.Salteados, p.Ruta+" (ya existe — `--forzar` lo pisa)")
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(ruta, []byte(p.Contenido), 0o644); err != nil {
		return fmt.Errorf("escribiendo %s: %w", p.Ruta, err)
	}
	r.Escritos = append(r.Escritos, p.Ruta)
	return nil
}

// Portamodelo es el contenido del archivo de agente de UN alias.
//
// Cinco líneas y ni una instrucción de SpecForge. Que no las tenga no es
// minimalismo: es lo que garantiza que esto NO sea una cuarta copia de cada
// skill. El día que alguien le agregue "y acordate de correr sf context", este
// archivo pasa a ser una fuente que hay que mantener sincronizada con
// `skills/`, y ahí empieza el problema que evitamos.
func Portamodelo(m global.Modelo, harness string) string {
	// La descripción tampoco explica cuándo usarlo, y eso es deliberado: el
	// orquestador NO elige este agente por su descripción, lo recibe por nombre
	// exacto. Una descripción que explicara el método sería método adentro del
	// archivo, que es justo lo que no puede haber.
	//
	// El esfuerzo se emite con el nombre que usa CADA harness, y sólo si está
	// declarado. sf no inventa un default: vacío significa "el que traiga el
	// modelo", que es una respuesta distinta de "bajo".
	esfuerzo := ""
	if m.Esfuerzo != "" {
		if clave, hay := claveDeEsfuerzo[harness]; hay {
			esfuerzo = fmt.Sprintf("%s: %s\n", clave, m.Esfuerzo)
		}
	}
	return fmt.Sprintf(`---
name: %s
description: SpecForge — un modelo. Se invoca por nombre exacto, nunca por descripción.
mode: subagent
tools: "*"
model: %s
%s---

Seguí exactamente las instrucciones que te dé quien te invocó. No pidas
confirmación y no cambies de tema.

Este archivo no aporta método: lo único que aporta es el modelo. Todo lo demás
llega en el prompt.
`, global.NombreDeAgente(m.Alias), m.ID, esfuerzo)
}

// claveDeEsfuerzo es cómo se llama el esfuerzo en el frontmatter de cada harness.
//
// opencode NO está, y no es un olvido: su frontmatter de agente no expone un
// campo de esfuerzo —lo expresa con opciones del proveedor, que son distintas
// para cada uno— y sf no va a inventar una traducción. Ahí el esfuerzo viaja por
// la línea de comandos cuando se lanza headless, o no viaja.
var claveDeEsfuerzo = map[string]string{
	"commandcode": "reasoningEffort",
}

// escribirPortamodelos genera uno por alias declarado.
//
// Se PISAN siempre, y ésa es la diferencia con todo lo demás que escribe este
// paquete: un portamodelo es un derivado del catálogo, no un archivo de Javier.
// Si el catálogo dice que `laguna` es otro id, el archivo tiene que decir lo
// mismo — respetarle una edición manual sería dejar que el subagente corra con
// un modelo que el catálogo ya no declara.
func escribirPortamodelos(raiz string, g *global.Config, r *Resultado) error {
	dir, hay := agentes[g.Harness]
	if !hay || !global.NecesitaPortamodelo(g.Harness) {
		return nil
	}
	alias := g.Alias()
	if len(alias) == 0 {
		return nil
	}
	if err := os.MkdirAll(filepath.Join(raiz, dir), 0o755); err != nil {
		return err
	}
	for _, m := range alias {
		if m.Via != global.Subagente {
			// Un `via: consola` lo ejecuta el orquestador con sus manos: el
			// harness nunca lo lanza, así que un portamodelo suyo sería un
			// archivo que nadie invoca.
			continue
		}
		nombre := global.NombreDeAgente(m.Alias) + ".md"
		ruta := filepath.Join(raiz, dir, nombre)
		if err := os.WriteFile(ruta, []byte(Portamodelo(m, g.Harness)), 0o644); err != nil {
			return fmt.Errorf("escribiendo %s: %w", nombre, err)
		}
		r.Escritos = append(r.Escritos, filepath.Join(dir, nombre))
	}
	return nil
}

// PortamodelosQueFaltan son los alias declarados que no tienen su archivo.
//
// Lo usa `sf doctor`: un alias sin portamodelo es un `sf next` que va a devolver
// un `agente:` que el harness no conoce, y eso falla en el momento más caro —
// cuando el orquestador ya lanzó.
func PortamodelosQueFaltan(raiz string, g *global.Config) []string {
	dir, hay := agentes[g.Harness]
	if !hay || !global.NecesitaPortamodelo(g.Harness) {
		return nil
	}
	var faltan []string
	for _, m := range g.Alias() {
		if m.Via != global.Subagente {
			continue
		}
		ruta := filepath.Join(raiz, dir, global.NombreDeAgente(m.Alias)+".md")
		if _, err := os.Stat(ruta); err != nil {
			faltan = append(faltan, m.Alias)
		}
	}
	return faltan
}

// PermitirComando agrega un comando a la allowlist del harness.
//
// Es lo que `sf init` usa con el `test_cmd`, y existe porque `dont-ask` cambió
// el costo de una lista incompleta: antes el harness preguntaba y alcanzaba con
// un aviso; ahora deniega, y un `test_cmd` afuera de la lista hace fallar el ⑲
// de todos los lotes.
//
// Sólo se agrega el PRIMER token —`npm` de `npm test`— porque es lo que las dos
// sintaxis de allowlist saben expresar, y porque autorizar la línea entera sería
// más frágil sin ser más seguro: quien puede correr `npm` puede correr cualquier
// script del package.json.
//
// Devuelve si hizo falta tocar algo. No parsea el JSON: busca el texto de la
// regla y, si no está, lo inserta. Es feo y es a propósito — un parser de la
// config de otra herramienta es una promesa de mantenerla, y esto se puede
// deshacer a mano en diez segundos.
func PermitirComando(raiz, harness, cmd string) (string, bool, error) {
	p, hay := permisos[harness]
	if !hay || strings.TrimSpace(cmd) == "" {
		return "", false, nil
	}
	prog := strings.Fields(cmd)[0]

	var regla string
	switch harness {
	case "claude-code":
		regla = fmt.Sprintf("\"Bash(%s:*)\"", prog)
	case "commandcode":
		regla = fmt.Sprintf("\"Shell(%s:*)\"", prog)
	default:
		// opencode ya tiene `bash: {"*": "allow"}`: no hay nada que agregar, y
		// agregarlo igual sería ruido en la config de alguien.
		return "", false, nil
	}

	ruta := filepath.Join(raiz, p.Ruta)
	b, err := os.ReadFile(ruta)
	if err != nil {
		return "", false, nil // sin config no hay lista que completar
	}
	if strings.Contains(string(b), regla) {
		return p.Ruta, false, nil
	}

	// Se inserta después del primer `"allow": [`, que es el único lugar donde
	// esta clave existe en los archivos que escribimos nosotros.
	const ancla = `"allow": [`
	i := strings.Index(string(b), ancla)
	if i < 0 {
		return "", false, nil
	}
	j := i + len(ancla)
	nuevo := string(b[:j]) + regla + ", " + string(b[j:])
	if err := os.WriteFile(ruta, []byte(nuevo), 0o644); err != nil {
		return "", false, err
	}
	return p.Ruta, true, nil
}
