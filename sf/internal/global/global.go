// Package global lee y escribe el catálogo de modelos: qué modelos tenés y cómo
// se invoca cada uno.
//
// ────────────────────────────────────────────────────────────────────────────
// LAS TRES CAPAS, Y CADA UNA ES DE ALGUIEN DISTINTO
// ────────────────────────────────────────────────────────────────────────────
//
// Este paquete existe porque "¿con qué modelo corro este paso?" no tiene una
// respuesta: tiene tres, y cada una la contesta otro.
//
//	perfil    vocabulario de sf        en el binario           razonar · construir
//	alias     vocabulario de Javier    acá Y en tareas.json    laguna
//	id        vocabulario del harness  SÓLO acá                laguna/s2.1
//
// UN ALIAS NO ES UN MODELO: ES UNA FORMA DE CORRER UN MODELO. O sea el par
// (id, esfuerzo), y a veces más. Por eso `laguna` y `laguna-pensando` pueden ser
// dos aliases sobre el mismo id, y por eso sigue habiendo UN archivo de agente
// por alias en vez de alias × esfuerzo.
//
// Y el esfuerzo vive en el alias y no en el perfil aunque conceptualmente sea el
// paso el que lo pide, porque el alias se declara ADENTRO de un perfil: el mismo
// id puede aparecer en `construir` con esfuerzo bajo y en `razonar` con esfuerzo
// alto, y son dos aliases distintos porque son dos formas distintas de correrlo.
//
// El alias es la capa que hace que todo cierre. El ⑯ escribe un ALIAS y nunca un
// id: un id en tareas.json sería un dato de la máquina de Javier metido en un
// archivo versionado que otro harness va a leer. El alias no — está definido en
// el bloque de cada harness apuntando a ids distintos, así que el plan SOBREVIVE
// al cambio de harness, que es el escenario real.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ ESTÁ INDEXADO POR HARNESS
// ────────────────────────────────────────────────────────────────────────────
//
// Porque una máquina tiene varios harness instalados a la vez, y ése es el caso
// normal y no el raro. Con un solo bloque de perfiles, instalar en opencode deja
// `razonar: opus` —que ahí no existe— y volver a Claude Code deja
// `anthropic/claude-opus-4-1`, que ahí tampoco. El archivo no sobreviviría a
// cambiar de harness, y la forma de descubrirlo sería que el subagente se lance
// con el modelo equivocado sin decir nada.
//
// Indexado, cada bloque se llena una vez y cambiar de harness no destruye nada.
// `harness:` arriba deja de ser configuración y pasa a ser lo que siempre fue:
// un puntero.
//
// ────────────────────────────────────────────────────────────────────────────
// NO HAY SIEMBRA, Y NO HAY EXCEPCIÓN PARA NADIE
// ────────────────────────────────────────────────────────────────────────────
//
// Una versión anterior sembraba `opus`/`sonnet`/`haiku` para claude-code, con el
// argumento de que ahí no son ids de proveedor sino alias del propio harness. El
// argumento se sostenía; la consecuencia no: dejaba el camino feliz con forma de
// Anthropic y a todos los demás con una parada.
//
// Ahora el bloque nace vacío para todos por igual, y la parada la contesta Javier
// una vez por harness. Es el mismo mecanismo que `dependencias_aprobadas` —una
// lista que se llena con cada aprobación— aplicado a la única cosa que sf no
// puede saber. El invariante que compra es testeable: ningún nombre de modelo
// concreto vive en este repo.
//
// ────────────────────────────────────────────────────────────────────────────
// Y POR QUÉ sf NO ELIGE DE LA LISTA
// ────────────────────────────────────────────────────────────────────────────
//
// Un perfil tiene VARIOS modelos porque el ⑯ tiene que poder decir "este lote es
// plomería, va con el barato". Pero el que elige es el ⑯, no sf: la lista, su
// orden, `capacidad` y `para` son opinión de Javier escrita a mano. sf no los
// interpreta, no los compara y no hace aritmética con ellos — los TRANSPORTA al
// sobre y listo.
//
//	sf no puede tener una opinión sobre qué modelo se parece a cuál.
//	Sí puede llevar la de Javier hasta quien la necesita.
package global

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Carpeta es dónde vive el catálogo, tanto en el home como en el proyecto.
const Carpeta = ".specforge"

// Archivo es el catálogo dentro de esa carpeta.
const Archivo = "modelos.yaml"

// Los tres valores de `via:` (H17).
//
// El tercero es el que hizo falta el catálogo: DeepSeek, Grok o Codex no pueden
// ser subagentes de Claude Code, así que para ésos el orquestador tiene que salir
// por consola. Y el ⑱ deja de ser un caso especial — es la misma llamada con otro
// `via:`.
const (
	Vos       = "vos"       // trabaja Javier, de frente
	Subagente = "subagente" // el harness lo lanza con sus propias manos
	Consola   = "consola"   // se sale por CLI: el orquestador le presta las manos
)

// Los tres perfiles. Son el vocabulario de sf y viven en el binario.
//
// Son tres y no cuatro ni siete porque son los que la máquina puede distinguir
// con lo que sabe. Mecanico NO lo devuelve ningún estado: lo nombran los sfx-*,
// que están fuera de los nueve. Por eso su entrada es opcional y no frena.
const (
	Razonar   = "razonar"
	Construir = "construir"
	Mecanico  = "mecanico"
)

// PerfilesConocidos son los tres, de más capaz a menos.
//
// Existe para que doctor y los mensajes puedan recorrerlos sin escribir la lista
// una segunda vez.
func PerfilesConocidos() []string { return []string{Razonar, Construir, Mecanico} }

// EsPerfil dice si un nombre es uno de los tres.
func EsPerfil(n string) bool { return n == Razonar || n == Construir || n == Mecanico }

// Modelo es UNA entrada del catálogo: cómo se lanza y para qué sirve.
type Modelo struct {
	// Alias es el nombre corto de Javier, y es lo que viaja a tareas.json.
	Alias string `yaml:"alias"`

	// ID es cómo lo nombra el harness. NUNCA sale de este archivo.
	ID string `yaml:"id"`

	Via string `yaml:"via"`

	// Comando es con qué se lo llama, y sólo tiene sentido con `via: consola`.
	//
	// Se guarda el prefijo y no una plantilla con huecos: el orquestador le pega
	// el prompt atrás, y un `{prompt}` sería un formato más que documentar para no
	// ganar nada.
	Comando string `yaml:"comando,omitempty"`

	// Esfuerzo es cuánto tiene que pensar, y lo entiende el harness, no sf.
	//
	// Los tres lo exponen: Command Code con `--effort` y con `reasoningEffort:`
	// en el frontmatter del agente, los otros con sus propios nombres. sf lo
	// TRANSPORTA — no valida los valores, porque cuáles son válidos depende del
	// modelo y del proveedor, y una lista blanca acá sería una tabla que se
	// pudre igual que la de los ids.
	//
	// Vacío es lo normal: significa "el que traiga el modelo por default".
	Esfuerzo string `yaml:"esfuerzo,omitempty"`

	// Capacidad y Para son PARA LOS OJOS DEL ⑯, no campos de cómputo.
	//
	// sf los lleva al sobre textuales y no los mira. Si sf eligiera "el más barato
	// con capacidad >= 2" estaría eligiendo modelo, y eso es lo único que no puede
	// hacer.
	Capacidad int    `yaml:"capacidad,omitempty"`
	Para      string `yaml:"para,omitempty"`
}

// Catalogo es perfil → los modelos declarados para él, EN ORDEN.
//
// El primero de cada lista es el default de ese perfil. El orden ES la
// preferencia, y no hay un `default: true` que pueda contradecirlo: un solo
// hecho, en un solo lugar.
type Catalogo map[string][]Modelo

// Config es el archivo entero.
type Config struct {
	// Harness es cuál está activo. Lo mueve `sf install --harness=…`.
	//
	// No se detecta en cada corrida a propósito: la detección es una PISTA (una
	// variable de entorno que puede estar o no), y una pista que se re-evalúa
	// daría respuestas distintas según desde dónde se invoque.
	Harness string `yaml:"harness"`

	// Harnesses es harness → su catálogo.
	Harnesses map[string]Catalogo `yaml:"harnesses"`

	// Modelos son los sueltos: los que no compiten en ningún perfil.
	//
	// Es donde vive un `sf model deepseek --via consola`: un modelo ajeno que
	// Javier invoca a mano en el ⑳.
	Modelos map[string]Modelo `yaml:"modelos,omitempty"`
}

// ErrNoHay es que todavía no se corrió `sf install`.
var ErrNoHay = errors.New("no hay catálogo de modelos: corré `sf install`")

// Ruta devuelve la carpeta global.
//
// SPECFORGE_HOME existe para los tests y para quien tenga el home en un lugar
// raro. No es una feature que documentarle a nadie: es la forma de que un test no
// escriba en el home de verdad de quien lo corre.
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

// RutaDeProyecto devuelve la carpeta del catálogo DE ESTE PROYECTO.
//
// El del proyecto le gana al global, y existe porque no todos los repos quieren
// los mismos modelos: uno de CRUD puede correr con el barato y uno de
// concurrencia no. Se siembra copiando el global en `sf init` y va gitignoreado —
// es de la máquina Y del proyecto, así que versionarlo le rompería el clon a
// cualquier otro.
func RutaDeProyecto(raiz string) string { return filepath.Join(raiz, Carpeta) }

// Leer carga el catálogo global.
func Leer() (*Config, error) {
	dir, err := Ruta()
	if err != nil {
		return nil, err
	}
	return leerDe(dir)
}

// LeerPara carga el catálogo que rige EN ESTE PROYECTO.
//
// El del proyecto gana si existe; si no, el global. NO se mergean: un catálogo a
// medias entre dos archivos sería imposible de razonar cuando algo saliera mal, y
// "cuál manda" tiene que contestarse mirando un solo lugar.
func LeerPara(raiz string) (*Config, error) {
	c, err := leerDe(RutaDeProyecto(raiz))
	switch {
	case err == nil:
		return c, nil
	case errors.Is(err, ErrNoHay):
		return Leer()
	default:
		return nil, err
	}
}

func leerDe(dir string) (*Config, error) {
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
	if c.Harnesses == nil {
		c.Harnesses = map[string]Catalogo{}
	}
	if c.Modelos == nil {
		c.Modelos = map[string]Modelo{}
	}
	c.migrar()
	return &c, nil
}

// Guardar escribe el catálogo global.
func (c *Config) Guardar() error {
	dir, err := Ruta()
	if err != nil {
		return err
	}
	return c.GuardarEn(dir)
}

// GuardarEn escribe el catálogo en una carpeta concreta.
func (c *Config) GuardarEn(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creando %s: %w", dir, err)
	}

	b, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	// El encabezado se pone a mano porque yaml.Marshal no escribe comentarios, y
	// este archivo lo va a abrir un humano: sin una línea que diga qué es y quién
	// lo escribe, parece basura de una herramienta.
	cuerpo := "# " + Carpeta + "/" + Archivo + " — qué modelos tenés y cómo se invoca cada uno.\n" +
		"# Lo crea `sf install` VACÍO y CRECE SOLO: cada `sf model` que apruebes queda acá.\n" +
		"# El primero de cada perfil es su default. `capacidad` y `para` los lee el ⑯, no sf.\n\n" +
		string(b)

	return os.WriteFile(filepath.Join(dir, Archivo), []byte(cuerpo), 0o644)
}

// ────────────────────────────────────────────────────────────────────────────
// Resolución
// ────────────────────────────────────────────────────────────────────────────

// Activo es el catálogo del harness que está activo.
//
// Devuelve un catálogo vacío y no nil cuando el harness no tiene bloque: quien
// llama pregunta "¿está declarado esto?" y la respuesta correcta es "no", no un
// panic. Es la misma lección de A9.
func (c *Config) Activo() Catalogo {
	if c == nil || c.Harnesses == nil {
		return Catalogo{}
	}
	if cat, hay := c.Harnesses[c.Harness]; hay && cat != nil {
		return cat
	}
	return Catalogo{}
}

// Default es el primero de la lista de un perfil: lo que sale cuando nadie pidió
// nada.
func (c *Config) Default(perfil string) (Modelo, bool) {
	lista := c.Activo()[perfil]
	if len(lista) == 0 {
		return Modelo{}, false
	}
	return lista[0], true
}

// PorAlias busca un alias en el catálogo del harness activo, y después entre los
// sueltos.
func (c *Config) PorAlias(alias string) (Modelo, bool) {
	for _, perfil := range PerfilesConocidos() {
		for _, m := range c.Activo()[perfil] {
			if m.Alias == alias {
				return m, true
			}
		}
	}
	if c != nil && c.Modelos != nil {
		if m, hay := c.Modelos[alias]; hay {
			return m, true
		}
	}
	return Modelo{}, false
}

// Resolver contesta "cómo se lanza esto acá", donde `que` es un perfil o un alias.
//
// El perfil se resuelve PRIMERO, y por eso Declarar rechaza un alias que se llame
// como un perfil: si los dos pudieran ganar, cuál gana dependería del orden de
// este método, y eso es la clase de detalle que aparece un martes.
func (c *Config) Resolver(que string) (Modelo, bool) {
	if EsPerfil(que) {
		return c.Default(que)
	}
	return c.PorAlias(que)
}

// Declarar agrega un modelo al catálogo del harness activo. Es el "crece solo".
//
// Se agrega al FINAL de la lista del perfil, que es lo correcto: el primero es el
// default, y lo primero que declarás es lo que más vas a usar. Reordenar es
// editar el YAML a mano, y está bien que lo sea — es un archivo de Javier, no una
// base de datos.
//
// Devuelve si el modelo es NUEVO en el archivo. Eso lo necesita `sf model` para
// saber si tiene que avisar que hay que reiniciar el harness: los portamodelo se
// leen al arrancar (medido), así que un alias nuevo a mitad de sesión no se ve.
func (c *Config) Declarar(perfil string, m Modelo) (nuevo bool, err error) {
	if !EsPerfil(perfil) {
		return false, fmt.Errorf("%q no es un perfil: son %v", perfil, PerfilesConocidos())
	}
	if m.Alias == "" {
		return false, errors.New("un modelo sin alias no se puede nombrar desde tareas.json")
	}
	if EsPerfil(m.Alias) {
		return false, fmt.Errorf("el alias %q se llama igual que un perfil: elegí otro", m.Alias)
	}
	if c.Harnesses == nil {
		c.Harnesses = map[string]Catalogo{}
	}
	if c.Harnesses[c.Harness] == nil {
		c.Harnesses[c.Harness] = Catalogo{}
	}

	cat := c.Harnesses[c.Harness]
	for i, y := range cat[perfil] {
		if y.Alias == m.Alias {
			cat[perfil][i] = m // redeclarar es actualizar, no duplicar
			return false, nil
		}
	}
	cat[perfil] = append(cat[perfil], m)
	return true, nil
}

// DeclararSuelto agrega un modelo que no es de ningún perfil.
func (c *Config) DeclararSuelto(nombre string, m Modelo) {
	if c.Modelos == nil {
		c.Modelos = map[string]Modelo{}
	}
	m.Alias = nombre
	c.Modelos[nombre] = m
}

// Alias son todos los alias declarados para el harness activo, sin repetir.
//
// Es lo que `sf install` recorre para generar los portamodelo, y lo que
// `sf doctor` recorre para comprobar que no falte ninguno.
func (c *Config) Alias() []Modelo {
	var todos []Modelo
	visto := map[string]bool{}
	for _, perfil := range PerfilesConocidos() {
		for _, m := range c.Activo()[perfil] {
			if !visto[m.Alias] {
				visto[m.Alias] = true
				todos = append(todos, m)
			}
		}
	}
	return todos
}

// ────────────────────────────────────────────────────────────────────────────
// La semilla y el harness
// ────────────────────────────────────────────────────────────────────────────

// Semilla arma la configuración inicial: el bloque del harness, VACÍO.
//
// Vacío a propósito. Ver el encabezado del paquete: no hay siembra y no hay
// excepción para nadie. El primer `sf next` para y pregunta, y esas dos
// respuestas valen para todos los repos de esa máquina.
func Semilla(harness string) *Config {
	return &Config{
		Harness:   harness,
		Harnesses: map[string]Catalogo{harness: {}},
		Modelos:   map[string]Modelo{},
	}
}

// PrefijoDeAgente es cómo se llaman los portamodelo.
const PrefijoDeAgente = "sf-"

// NombreDeAgente es a quién invocar para conseguir un alias.
func NombreDeAgente(alias string) string { return PrefijoDeAgente + alias }

// NecesitaPortamodelo dice si este harness exige un archivo de agente para poder
// elegir el modelo del subagente.
//
// Claude Code NO: su herramienta de subagente acepta el modelo como parámetro de
// la llamada. opencode y Command Code SÍ, y eso está MEDIDO y no deducido —
// sonda del 29/08: un agente con `model:` apuntando a un id inexistente falló con
// "modelo desconocido" en los dos. O sea que el modelo sale del archivo y no hay
// forma de pisarlo al invocar.
//
// De ahí sale todo el portamodelo: un archivo por alias, de cinco líneas, sin una
// sola instrucción de SpecForge. El modelo va en el archivo; el skill sigue
// llegando por el prompt.
func NecesitaPortamodelo(harness string) bool {
	return harness != "" && harness != "claude-code" && harness != "desconocido"
}

// Harness son los que sf sabe instalar.
var Harness = []string{"claude-code", "opencode", "commandcode"}

// DetectarHarness adivina dónde estamos, y es una PISTA, no una certeza.
//
// Que la variable esté es evidencia de que sí; que NO esté no prueba nada — puede
// ser otro harness, o el mismo invocado de otra forma. Por eso el que llama tiene
// que poder pisarlo (`sf install --harness=…`) y por eso el resultado se ESCRIBE
// en vez de recalcularse en cada corrida.
func DetectarHarness() string {
	switch {
	case os.Getenv("CLAUDECODE") != "", os.Getenv("CLAUDE_CODE_ENTRYPOINT") != "":
		return "claude-code"
	case os.Getenv("OPENCODE") != "", os.Getenv("OPENCODE_BIN_PATH") != "":
		return "opencode"
	case os.Getenv("COMMANDCODE") != "", os.Getenv("COMMAND_CODE_ENTRYPOINT") != "":
		return "commandcode"
	}
	return "desconocido"
}
