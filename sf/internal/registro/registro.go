// Package registro es la película. El `estado.json` es la foto.
//
// ────────────────────────────────────────────────────────────────────────────
// PARA QUÉ EXISTE
// ────────────────────────────────────────────────────────────────────────────
//
// Porque `sf` controla los ARTEFACTOS y no controla la CONDUCTA:
//
//	¿existe el archivo? ¿pasan los tests? ¿son tres opciones?     → sf es bueno
//	¿lanzaste el subagente? ¿cuántas vueltas diste? ¿quién aprobó? → sf no ve NADA
//
// Los cuatro agujeros medidos el 2026-09-03 —el bucle de revisión sin corte, la
// vuelta que declara el propio revisor, el subagente que nunca se lanzó, el
// sello imposible de probar— son los cuatro de conducta. Ninguno de artefacto.
// Este paquete es el órgano que le falta para verla.
//
// Y es la pieza ⓪ de `por-tramos.md`: un test de tramo es una corrida con un
// modelo, o sea que NO es determinista, y "pasó" no puede significar "salió
// bien". Sin esto, cada tramo devuelve impresiones en vez de datos.
//
// ────────────────────────────────────────────────────────────────────────────
// LA REGLA DURA, Y LAS TRES PARTES SE SOSTIENEN SOLAS
// ────────────────────────────────────────────────────────────────────────────
//
//	El registro ESCRIBE. No frena, no lo lee ninguna compuerta, y NO PUEDE
//	ROMPER UN COMANDO.
//
// La tercera es la que importa: si el disco está lleno, si `.specforge/` es de
// sólo lectura, si el JSON no serializa — `sf next` tiene que contestar
// exactamente lo mismo y con el mismo exit code. Todos los errores de acá se
// tragan, a propósito y sin excepción.
//
//	Un tablero que puede apagar el motor es peor que no tener tablero.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ HAY UNA VARIABLE DE PAQUETE Y NO UN PUNTERO QUE VIAJE
// ────────────────────────────────────────────────────────────────────────────
//
// Por un hecho del programa, no por comodidad: `sf` es UN PROCESO, UN COMANDO,
// UNA ENTRADA. No hay dos en vuelo nunca. Enhebrar un puntero por veinte
// funciones para un asunto lateral ensuciaría firmas que hoy están limpias.
package registro

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// Archivo es dónde vive el registro, relativo a la raíz del proyecto.
//
// EN `.specforge/` Y NO EN `.docs/`, y la razón es quién lee cada carpeta:
//
//	.docs/       lo que el proyecto PRODUCE Y COMMITEA — brief, PRD, estado.json
//	.specforge/  lo de ESTA máquina: el catálogo, los lanzamientos. Gitignoreado
//	             por `sf init` (arranque.go:117)
//
// Una línea por invocación de `sf`: versionarlo metería ruido en CADA commit, y
// su contenido —pids, ids de modelo— sólo significa algo en esta máquina.
//
// Se revisa el día que una compuerta lea el registro: ahí pasa de instrumento a
// evidencia, y una evidencia que no viaja con el repo no sirve.
const Archivo = global.Carpeta + "/registro.jsonl"

// Version es la forma de la línea. Cuesta un entero y evita que un lector
// futuro tenga que adivinar por qué faltan campos.
const Version = 1

// Tope es el largo máximo de una línea, en bytes.
//
// Existe para que el `O_APPEND` de un solo `Write` siga siendo la operación
// atómica que asumimos en `escribir`. Cuando una línea se pasa, se recortan los
// tres campos de largo impredecible —args, fallas, avisos— y `trunco` queda en
// true. NUNCA se tira una línea entera: una línea que falta es un hueco
// invisible en la película, y es justo lo que este archivo existe para no tener.
const Tope = 4096

// VarDelegado es la marca que `sf lanzar` le pone al hijo.
//
// Es el ÚNICO valor fiable de `quien`, porque lo puso sf mismo — a diferencia de
// `terminal` y `agente`, que se deducen de señales que el propio agente puede
// producir (ver el comentario de `quienCorrio`).
const VarDelegado = "SPECFORGE_DELEGADO"

// Producto son los cuatro sellos, tal cual están en el estado.
//
// Se guardan CRUDOS y no derivados a "en qué estado de producto estás", que
// sería una segunda copia de la lógica de `terminarProducto`. Un diff entre el
// antes y el después muestra exactamente qué sello se movió, sin que este
// paquete tenga que saber en qué orden van.
type Producto struct {
	Brief        string `json:"brief,omitempty"`
	Prd          string `json:"prd,omitempty"`
	Constitucion bool   `json:"constitucion,omitempty"`
	Backlog      bool   `json:"backlog,omitempty"`
}

// Foto es el estado en un instante: al abrir la entrada y al cerrarla.
//
// ────────────────────────────────────────────────────────────────────────────
// SÍ, `lote` ES DEDUCIBLE. Y SE GUARDA IGUAL.
// ────────────────────────────────────────────────────────────────────────────
//
// R6 dice "lo que se puede deducir, no se guarda", y el lote actual es el
// primero sin commit. Guardarlo acá NO viola la regla, porque R6 es una regla
// sobre el ESTADO —donde un dato repetido se desincroniza—. Una línea del
// registro es INMUTABLE y describe un instante que ya pasó: no puede
// desincronizarse con nada, y tiene que poder leerse sin el estado.json al lado.
type Foto struct {
	Feature  string   `json:"feature,omitempty"`
	Estado   string   `json:"estado,omitempty"`
	Lote     int      `json:"lote,omitempty"`
	Intentos int      `json:"intentos,omitempty"`
	Producto Producto `json:"producto,omitempty"`
}

// Entrada es una línea del registro.
type Entrada struct {
	// ① Qué se corrió.
	T      string   `json:"t"`
	V      int      `json:"v"`
	Cmd    string   `json:"cmd"`
	Args   []string `json:"args,omitempty"`
	Salida int      `json:"salida"`

	// Ms es cuánto tardó, y es lo que distingue "el ㉑ dio dos vueltas" de "el
	// ㉑ dio dos vueltas y la segunda tardó veinte minutos".
	Ms int64 `json:"ms"`

	// ② Quién lo corrió. Ver `quienCorrio` para el límite que esto tiene.
	Quien    string `json:"quien"`
	Tty      bool   `json:"tty"`
	Delegado string `json:"delegado,omitempty"`
	Arnes    string `json:"arnes,omitempty"`
	Pid      int    `json:"pid"`
	Ppid     int    `json:"ppid"`

	// ③ Qué había y qué quedó.
	Antes   Foto `json:"antes"`
	Despues Foto `json:"despues"`

	// ④ Qué dijo sf. Sólo los llena `sf next` (y Lanzamiento, `sf lanzar`).
	//
	// Perfil y Modelo van los dos y no es redundancia: es la misma distinción
	// que `maquina.Instruccion` ya hace —qué PIDE el paso contra qué se le dio
	// en ESTA máquina—. Con una sola palabra, dentro de tres meses las dos
	// preguntas tienen la misma respuesta y no se pueden separar.
	Tipo string `json:"tipo,omitempty"`

	// Estado y Feature son lo que la MÁQUINA contestó, y viven acá y no en la
	// Foto porque son otra cosa:
	//
	//	Foto.Estado      el estado de la feature, leído del estado.json
	//	Entrada.Estado   el paso que sf nombró — y en los cinco de producto
	//	                 ("brief", "prd") eso NO es un campo de ningún archivo
	//
	// Pisar la Foto con esto —que fue la primera versión— rompía dos cosas: el
	// `tomarLaProxima` del ⑪ nombra la feature SIGUIENTE y no la actual, y un
	// `sf next` sobre un paso de producto dejaba un "antes" que el estado.json
	// nunca dijo.
	Estado  string `json:"estado,omitempty"`
	Feature string `json:"feature,omitempty"`

	Skill       string `json:"skill,omitempty"`
	Via         string `json:"via,omitempty"`
	Perfil      string `json:"perfil,omitempty"`
	Modelo      string `json:"modelo,omitempty"`
	Esfuerzo    string `json:"esfuerzo,omitempty"`
	Lanzamiento string `json:"lanzamiento,omitempty"`

	// ⑤ Cómo salió.
	//
	// Fallas y Avisos son el dato MÁS VALIOSO del registro para la etapa de
	// tramos: son, literalmente, la lista de las veces que un modelo produjo
	// algo que la compuerta no aceptó. Hoy se imprimen y se pierden.
	Movio  bool     `json:"movio,omitempty"`
	Fallas []string `json:"fallas,omitempty"`
	Avisos []string `json:"avisos,omitempty"`
	Trunco bool     `json:"trunco,omitempty"`
}

// Paso es lo que `sf next` sabe y el registro no puede deducir.
//
// Es un struct propio y no `maquina.Instruccion` para que este paquete NO
// dependa de `maquina`: el que traduce es main, que ya conoce a los dos.
type Paso struct {
	Tipo, Estado, Feature      string
	Skill, Perfil, Modelo, Via string
	Esfuerzo                   string
}

// enCurso es la entrada de ESTE proceso. Ver el encabezado del paquete.
var enCurso *Entrada

// raizEnCurso es dónde vive el proyecto. Se guarda al abrir para no volver a
// preguntárselo al sistema al cerrar.
var raizEnCurso string

// arranque es cuándo empezó, para poder medir Ms.
var arranque time.Time

// Abrir empieza una entrada y saca la foto del ANTES.
//
// `args` son los del comando ya sin el nombre, tal cual los recibió main.
func Abrir(raiz, cmd string, args []string) {
	arranque = time.Now()
	raizEnCurso = raiz
	enCurso = &Entrada{
		V:     Version,
		Cmd:   cmd,
		Args:  args,
		Pid:   os.Getpid(),
		Ppid:  os.Getppid(),
		Antes: foto(raiz),
	}
}

// AnotarPaso guarda lo que `sf next` contestó.
func AnotarPaso(p Paso) {
	if enCurso == nil {
		return
	}
	enCurso.Tipo, enCurso.Skill, enCurso.Via = p.Tipo, p.Skill, p.Via
	enCurso.Perfil, enCurso.Modelo, enCurso.Esfuerzo = p.Perfil, p.Modelo, p.Esfuerzo
	enCurso.Estado, enCurso.Feature = p.Estado, p.Feature
}

// DeQuienEs devuelve la feature de la que habla esta entrada, o "" si es de
// producto. Mira los tres lugares donde puede estar, en orden de confianza: lo
// que sf dijo, lo que había, y lo que quedó.
func (e Entrada) DeQuienEs() string {
	for _, f := range []string{e.Feature, e.Antes.Feature, e.Despues.Feature} {
		if f != "" {
			return f
		}
	}
	return ""
}

// AnotarVeredicto guarda lo que dijeron las compuertas, y en qué paso.
//
// `estado` viene de la máquina —`Cierre.Estado` y `Efecto.Estado`— y no se
// deduce acá: sin él, la línea de un `sf done` sobre un paso de PRODUCTO no
// puede decir de cuál era, porque "brief" o "prd" no son un valor de ningún
// campo del estado.json.
func AnotarVeredicto(est string, fallas, avisos []string, movio bool) {
	if enCurso == nil {
		return
	}
	enCurso.Fallas, enCurso.Avisos, enCurso.Movio = fallas, avisos, movio
	if est != "" {
		enCurso.Estado = est
	}
}

// AnotarLanzamiento ata esta invocación a la ficha de la corrida que largó.
//
// Es lo que permite contestar "entre este `next` que dijo subagente y el `done`
// que vino después, ¿hubo una corrida?".
func AnotarLanzamiento(id string) {
	if enCurso == nil {
		return
	}
	enCurso.Lanzamiento = id
}

// Cerrar saca la foto del DESPUÉS y escribe la línea.
//
// Todos los errores se tragan: ver la regla dura del encabezado.
func Cerrar(salida int) {
	if enCurso == nil {
		return
	}
	e := enCurso
	enCurso = nil // que un Cerrar repetido no escriba dos líneas

	e.T = time.Now().UTC().Format("2006-01-02T15:04:05.000Z")
	e.Ms = time.Since(arranque).Milliseconds()
	e.Salida = salida
	e.Despues = foto(raizEnCurso)
	e.Quien, e.Tty, e.Delegado = quienCorrio()
	e.Arnes = global.DetectarHarness()

	escribir(raizEnCurso, e)
}

// foto lee el estado.json y saca lo que hace falta.
//
// Lo lee ESTE paquete y no se lo pide al comando: así funciona igual en los
// comandos que nunca cargan el estado (`version`, `doctor`, `install`). Que el
// archivo no exista —un proyecto sin arrancar, un `sf init`— no es un error: la
// foto queda vacía.
func foto(raiz string) Foto {
	e, err := estado.Leer(raiz)
	if err != nil {
		return Foto{}
	}
	f := Foto{
		Feature: e.FeatureActual,
		Producto: Producto{
			Brief:        e.Producto.BriefSellado,
			Prd:          e.Producto.PrdHash,
			Constitucion: e.Producto.ConstitucionSellada,
			Backlog:      e.Producto.BacklogVisto,
		},
	}
	if fe, hay := e.Actual(); hay {
		f.Estado = fe.Estado
		f.Intentos = fe.IntentosFallidos
		if l, hayLote := fe.LoteActual(); hayLote {
			f.Lote = l.Lote
		}
	}
	return f
}

// ────────────────────────────────────────────────────────────────────────────
// QUIÉN CORRIÓ ESTO — y el límite está MEDIDO, no supuesto
// ────────────────────────────────────────────────────────────────────────────

// Los tres valores de `quien`. El tercero no prueba lo que su nombre sugiere.
const (
	// Delegado es el único FIABLE: la marca la puso sf mismo, en `sf lanzar`.
	Delegado = "delegado"

	// Terminal es que hay TTY y no hay ningún arnés alrededor.
	Terminal = "terminal"

	// Agente es todo lo demás, y NO DISTINGUE al orquestador de Javier
	// tipeando `!` en el chat. Ver `quienCorrio`.
	Agente = "agente"
)

// quienCorrio contesta desde dónde se invocó sf, y su límite es un HECHO MEDIDO.
//
// ────────────────────────────────────────────────────────────────────────────
// LA MEDICIÓN DEL 2026-09-04
// ────────────────────────────────────────────────────────────────────────────
//
// Se corrió la misma sonda dos veces: desde la herramienta Bash del agente, y
// tipeada por Javier con `!` en el MISMO Claude Code.
//
//	                     agente              Javier con `!`
//	[ -t 0 ]             NO-TTY              NO-TTY
//	tty(1)               not a tty           not a tty
//	padre → abuelo       zsh → claude        zsh → claude
//	AI_AGENT             claude-code_…       idéntico
//	CLAUDECODE           1                   idéntico
//	stdin                socket:[…]          /dev/null      ← lo único distinto
//
// Y el agente reproduce esa única diferencia con un redirect: `sf approve <
// /dev/null` da la salida idéntica. Peor: `script -qec 'sf approve' /dev/null`
// FABRICA UN TTY —probado, /dev/pts/7—, así que `hayPersona` tampoco alcanza.
//
//	Esto corrige vecinos.md §2, que daba por resuelta la firma del ⑰ con
//	`hayPersona`. No la resuelve.
//
// ────────────────────────────────────────────────────────────────────────────
// POR ESO SE GUARDAN LAS SEÑALES CRUDAS Y NO SÓLO LA ETIQUETA
// ────────────────────────────────────────────────────────────────────────────
//
//  1. La etiqueta es una INTERPRETACIÓN; las señales son el HECHO. El día que
//     la regla cambie, las líneas viejas se releen con la regla nueva. Con sólo
//     la etiqueta, no.
//  2. `arnes` atrapa lo que el TTY no: un `script` que fabrica un TTY NO
//     desactiva `CLAUDECODE=1`. Una línea con tty=true y arnes="claude-code" es
//     exactamente la firma de esa evasión. No la impide; la deja escrita.
//
// Que `agente` no distinga al orquestador de Javier NO ES UN PENDIENTE: es un
// límite. Mientras el agente controle el arranque de procesos, ningún mecanismo
// en banda puede probar que había un humano.
func quienCorrio() (etiqueta string, tty bool, delegado string) {
	delegado = os.Getenv(VarDelegado)
	tty = hayTTY()

	switch {
	case delegado != "":
		return Delegado, tty, delegado
	case tty && global.DetectarHarness() == "desconocido":
		return Terminal, tty, ""
	default:
		return Agente, tty, ""
	}
}

// hayTTY es una variable y no una función a secas para que los tests puedan
// simular la terminal. Es el mismo recurso que `modelos.correr`, por el mismo
// motivo: el CI no tiene un TTY que prestarle a un test.
var hayTTY = HayPersona

// HayPersona dice si del otro lado de stdin hay alguien.
//
// ────────────────────────────────────────────────────────────────────────────
// VIVE ACÁ Y NO EN main.go, Y ES A PROPÓSITO
// ────────────────────────────────────────────────────────────────────────────
//
// Estaba en main.go y la usaba un solo comando (`sf install`, para decidir si
// pregunta). Desde que el registro tiene que contestar la misma pregunta en
// general, tenerla en los dos lados serían dos copias de quince líneas — y "un
// dato en dos lugares es un dato que se desincroniza" es la regla del proyecto.
//
// MEDIDO el 2026-08-31: la shell de un agente NO tiene TTY y la de una persona
// en una terminal de verdad sí.
//
// Sale con la STDLIB SOLA, y eso importa: `sf` tiene hoy UNA dependencia
// (`gopkg.in/yaml.v3`), y meter `golang.org/x/term` para preguntar si hay una
// terminal la duplicaría.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ NO ALCANZA CON ModeCharDevice, QUE ES LO QUE DECÍA LA SPEC
// ────────────────────────────────────────────────────────────────────────────
//
// Porque /dev/null TAMBIÉN es un dispositivo de caracteres. La versión de
// `install-interactivo.md` §2 miraba sólo ese bit, y con eso un
// `sf install < /dev/null` decía "hay alguien".
//
// No colgaba, porque leer de /dev/null da EOF enseguida. Hacía algo peor y más
// callado: la pregunta se contestaba sola con "enter" en todo, y "enter" en la
// primera es TODOS LOS ARNESES. O sea que un `sf install` desatendido pasaba de
// armar el andamio de uno a armar el de los tres, en silencio. Lo encontró
// `TestInstall_SinTTYNoPreguntaNiCuelga`, no la lectura.
//
// ────────────────────────────────────────────────────────────────────────────
// LO QUE NO CUBRE, dicho para que nadie lo descubra solo
// ────────────────────────────────────────────────────────────────────────────
//
// Un dispositivo de caracteres que no sea /dev/null ni una terminal —/dev/zero,
// por ejemplo— sigue dando true. Cerrar eso pide un ioctl por plataforma o
// `golang.org/x/term`, y ninguna paga para un caso que no existe: los tres
// reales son una terminal, un pipe y /dev/null.
//
// Y desde la medición del 2026-09-04 se sabe algo más grande: esto NO SIRVE
// COMO AUTORIDAD, porque `script(1)` fabrica un TTY. Sirve como SEÑAL, que es
// para lo que este paquete la usa.
func HayPersona() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	// Un pipe o un archivo redirigido: no hay nadie.
	if fi.Mode()&os.ModeCharDevice == 0 {
		return false
	}
	// Y /dev/null es un dispositivo de caracteres sin nadie atrás. `os.DevNull`
	// vale también en Windows ("NUL"), así que esto no necesita build tags.
	if dn, err := os.Stat(os.DevNull); err == nil && os.SameFile(fi, dn) {
		return false
	}
	return true
}

// ────────────────────────────────────────────────────────────────────────────
// ESCRIBIR
// ────────────────────────────────────────────────────────────────────────────

// escribir agrega la línea al archivo. Se traga todos los errores.
//
// ① UN SOLO Write, CON LA LÍNEA ENTERA
//
// `O_APPEND` hace que el sistema mueva el puntero al final Y escriba en una sola
// operación, así que dos `sf` corriendo a la vez no se pisan a mitad de línea.
//
//	NO ESTÁ MEDIDO. Es lo que dice POSIX para O_APPEND y lo que hace Linux en
//	la práctica; no se hizo el experimento de dos procesos en paralelo. Si
//	alguna vez aparece una línea partida en el archivo, ESTA es la primera
//	sospechosa — y el Tope existe para que la suposición no se estire más de
//	lo que aguanta.
//
// ② CUÁNDO NO SE ESCRIBE NADA
//
// Si no existe ni `.docs/` ni `.specforge/`, no se escribe: es lo que evita que
// un `sf version` tipeado en una carpeta cualquiera deje basura.
//
// Y la comprobación es acá —al CERRAR— y no al abrir, lo que resuelve solo el
// caso de `sf init`: cuando la entrada se cierra, las carpetas ya las creó el
// propio comando, así que `sf init` QUEDA REGISTRADO. Que es lo correcto: es la
// primera línea de la película.
func escribir(raiz string, e *Entrada) {
	if !esProyecto(raiz) {
		return
	}

	linea, ok := serializar(e)
	if !ok {
		return
	}

	ruta := filepath.Join(raiz, Archivo)
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(ruta, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(linea)
}

// esProyecto dice si acá hay algo de SpecForge como para dejar un registro.
func esProyecto(raiz string) bool {
	for _, d := range []string{".docs", global.Carpeta} {
		if fi, err := os.Stat(filepath.Join(raiz, d)); err == nil && fi.IsDir() {
			return true
		}
	}
	return false
}

// serializar arma la línea y la recorta si se pasa del Tope.
//
// El orden del recorte no es casual: primero `args` —que es lo que el usuario
// tipeó y suele ser lo más largo (`sf new "…"`, `sf done --msg "…"`)—, después
// los avisos, y las fallas AL FINAL. Las fallas son el motivo por el que el
// estado no se movió: es lo último que uno quiere perder.
func serializar(e *Entrada) ([]byte, bool) {
	b, err := json.Marshal(e)
	if err != nil {
		return nil, false
	}
	if len(b)+1 <= Tope {
		return append(b, '\n'), true
	}

	e.Trunco = true
	for _, sacar := range []func(){
		func() { e.Args = nil },
		func() { e.Avisos = recortarLista(e.Avisos) },
		func() { e.Fallas = recortarLista(e.Fallas) },
	} {
		sacar()
		b, err = json.Marshal(e)
		if err != nil {
			return nil, false
		}
		if len(b)+1 <= Tope {
			return append(b, '\n'), true
		}
	}

	// Ni así entró. Se escribe el esqueleto: una línea recortada sigue siendo
	// una línea, y un hueco no.
	minima := Entrada{
		T: e.T, V: e.V, Cmd: e.Cmd, Salida: e.Salida, Ms: e.Ms,
		Quien: e.Quien, Tty: e.Tty, Delegado: e.Delegado, Arnes: e.Arnes,
		Pid: e.Pid, Ppid: e.Ppid,
		Antes: e.Antes, Despues: e.Despues, Trunco: true,
	}
	b, err = json.Marshal(&minima)
	if err != nil {
		return nil, false
	}
	return append(b, '\n'), true
}

// recortarLista deja la primera y dice cuántas se fueron.
//
// Deja UNA y no cero porque el motivo por el que algo no avanzó, aunque sea
// parcial, vale mucho más que un contador solo.
func recortarLista(l []string) []string {
	if len(l) <= 1 {
		return l
	}
	return []string{l[0], "…y " + itoa(len(l)-1) + " más (recortado)"}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
