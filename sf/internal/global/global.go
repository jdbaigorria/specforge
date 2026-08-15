// Package global lee y escribe `~/.specforge/`: lo que NO es de un proyecto.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ EXISTE UN LUGAR FUERA DEL REPO
// ────────────────────────────────────────────────────────────────────────────
//
// Casi todo el diseño vive dentro del proyecto, y a propósito: el estado, el
// roadmap y la constitución van versionados, porque en el ⑱ cambiás de modelo y
// el que implementa no estuvo en la conversación.
//
// Pero hay una cosa que NO es del proyecto: QUÉ MODELOS TENÉS Y CÓMO SE INVOCA
// CADA UNO. Eso es de la máquina de Javier y es igual en sus 16 repos — meterlo
// en cada `constitucion.md` sería copiar el mismo dato dieciséis veces.
//
// ────────────────────────────────────────────────────────────────────────────
// Y POR QUÉ EL HARNESS VIVE EN EL MISMO ARCHIVO QUE LOS MODELOS
// ────────────────────────────────────────────────────────────────────────────
//
// Porque los dos existen para contestar UNA sola pregunta: "¿cómo lanzo este
// modelo acá?". El argumento es de Javier y es lo que cerró H1b:
//
//	"si corre en Claude Code sabe que no puede usar un modelo fuera de
//	 Anthropic; pero si está corriendo en otro harness sabe que puede cambiar
//	 entre modelos de proveedores."
//
// Hay exactamente un lugar del diseño que ve las dos mitades:
//
//	tareas.json  (el ⑯)  →  QUÉ modelo hace falta     lo escribe el que planificó
//	~/.specforge/        →  EN QUÉ harness estás      lo escribe `sf install`
//	sf next              →  ⇒ CÓMO se lanza acá
//
// Ningún skill puede contestarlo —es un `.md` agnóstico— y el orquestador
// tampoco: él ES el harness, no lo sabe describir.
//
// ────────────────────────────────────────────────────────────────────────────
// LA LISTA NO SE ESCRIBE DE ANTEMANO: SE CONSTRUYE SOLA
// ────────────────────────────────────────────────────────────────────────────
//
// Es literalmente el mismo mecanismo que `dependencias_aprobadas`, aplicado a
// otra cosa. Cuando el ⑯ pide un modelo que no está declarado, sf NO ELIGE UN
// REEMPLAZO —eso sería opinar sobre qué modelo se parece a cuál, y R3 lo
// prohíbe— sino que para y pregunta. La respuesta de Javier es la que lo
// declara, y queda para siempre.
package global

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Carpeta es dónde vive todo esto, relativo al home.
const Carpeta = ".specforge"

// Archivo es el mapa de modelos dentro de esa carpeta.
const Archivo = "modelos.yaml"

// Los tres valores de `via:` (H17).
//
// El tercero es el que hizo falta el mapa: DeepSeek, Grok o Codex no pueden ser
// subagentes de Claude Code, así que para esos el orquestador tiene que salir
// por consola. Y el ⑱ deja de ser un caso especial — es la misma llamada con
// otro `via:`.
const (
	Vos       = "vos"       // trabaja Javier, de frente
	Subagente = "subagente" // el harness lo lanza con sus propias manos
	Consola   = "consola"   // se sale por CLI: el orquestador le presta las manos
)

// Modelo es cómo se invoca UN modelo en esta máquina.
type Modelo struct {
	Via string `yaml:"via"`

	// Comando es con qué se lo llama, y sólo tiene sentido con `via: consola`.
	//
	//	deepseek exec
	//
	// Se guarda el prefijo y no una plantilla con huecos: el orquestador le
	// pega el prompt atrás, y una plantilla con `{prompt}` sería un formato más
	// que inventar y que documentar para ganar nada.
	Comando string `yaml:"comando,omitempty"`
}

// Config es el archivo entero.
type Config struct {
	// Harness es dónde corre sf, y lo escribe `sf install`.
	//
	// No se detecta en cada corrida a propósito: la detección es una PISTA
	// (una variable de entorno que puede estar o no), y una pista que se
	// re-evalúa cada vez daría respuestas distintas según desde dónde se
	// invoque. Se decide una vez, al instalar, y queda escrito.
	Harness string `yaml:"harness"`

	// Modelos es nombre → cómo se lanza.
	Modelos map[string]Modelo `yaml:"modelos"`
}

// ErrNoHay es que todavía no se corrió `sf install`.
var ErrNoHay = errors.New("no hay ~/.specforge/modelos.yaml: corré `sf install`")

// Ruta devuelve la carpeta global.
//
// SPECFORGE_HOME existe para los tests y para quien tenga el home en un lugar
// raro. No es una feature que haya que documentarle a nadie: es la forma de que
// un test no escriba en el home de verdad de quien lo corre.
func Ruta() (string, error) {
	if h := os.Getenv("SPECFORGE_HOME"); h != "" {
		return h, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("no pude encontrar tu home: %w", err)
	}
	return filepath.Join(home, Carpeta), nil
}

// Leer carga el mapa de modelos.
func Leer() (*Config, error) {
	dir, err := Ruta()
	if err != nil {
		return nil, err
	}

	b, err := os.ReadFile(filepath.Join(dir, Archivo))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoHay
		}
		return nil, fmt.Errorf("leyendo %s: %w", Archivo, err)
	}

	var c Config
	if err := yaml.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("%s: %w", Archivo, err)
	}
	if c.Modelos == nil {
		c.Modelos = map[string]Modelo{}
	}
	return &c, nil
}

// Guardar escribe el mapa, creando la carpeta si hace falta.
func (c *Config) Guardar() error {
	dir, err := Ruta()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creando %s: %w", dir, err)
	}

	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// El encabezado se pone a mano porque yaml.Marshal no escribe comentarios,
	// y este archivo lo va a abrir un humano: sin una línea que diga qué es y
	// quién lo escribe, parece basura de una herramienta.
	cuerpo := "# ~/.specforge/modelos.yaml — qué modelos tenés y cómo se invoca cada uno.\n" +
		"# Lo crea `sf install` y CRECE SOLO: cada `sf model` que apruebes queda acá.\n\n" +
		string(b)

	return os.WriteFile(filepath.Join(dir, Archivo), []byte(cuerpo), 0o644)
}

// Buscar dice cómo se lanza un modelo, si está declarado.
func (c *Config) Buscar(nombre string) (Modelo, bool) {
	m, hay := c.Modelos[nombre]
	return m, hay
}

// Declarar agrega un modelo al mapa. Es el "se construye sola".
func (c *Config) Declarar(nombre string, m Modelo) {
	if c.Modelos == nil {
		c.Modelos = map[string]Modelo{}
	}
	c.Modelos[nombre] = m
}

// ────────────────────────────────────────────────────────────────────────────
// La semilla
// ────────────────────────────────────────────────────────────────────────────

// nativos son los modelos que sf usa por default en cualquier harness.
//
// Se siembran SIEMPRE, y no contradice el "la lista se construye sola": son los
// que la propia máquina elige cuando nadie eligió nada (`modeloPorEstado` y
// `modeloPorDefecto` en maquina.go). Un default que no está declarado haría que
// el primer `sf next` de un proyecto nuevo pare a pedir permiso para usar lo
// que sf mismo acaba de recomendar, que es absurdo.
//
// Lo que SÍ para es un modelo ajeno —deepseek, grok, codex—, que es exactamente
// el caso que el mapa existe para resolver.
var nativos = []string{"opus", "sonnet", "haiku"}

// Semilla arma la configuración inicial.
func Semilla(harness string) *Config {
	c := &Config{Harness: harness, Modelos: map[string]Modelo{}}
	for _, n := range nativos {
		c.Modelos[n] = Modelo{Via: Subagente}
	}
	return c
}

// DetectarHarness adivina dónde estamos, y es una PISTA, no una certeza.
//
// Que la variable esté es evidencia de que sí; que NO esté no prueba nada —
// puede ser otro harness, o el mismo invocado de otra forma. Por eso el que
// llama tiene que poder pisarlo (`sf install --harness=…`) y por eso el
// resultado se ESCRIBE en vez de recalcularse en cada corrida.
func DetectarHarness() string {
	if os.Getenv("CLAUDECODE") != "" || os.Getenv("CLAUDE_CODE_ENTRYPOINT") != "" {
		return "claude-code"
	}
	return "desconocido"
}
