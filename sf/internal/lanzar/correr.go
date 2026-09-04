package lanzar

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/jdbaigorria/specforge/sf/internal/registro"
	"time"
)

// Carpeta es dónde viven la ficha y el registro de cada corrida.
//
// EN `.specforge/` Y NO EN `.docs/`, y la razón es quién lee cada carpeta.
// `.docs/` es la documentación que el proyecto GUARDA Y COMMITEA — el brief, el
// PRD, las historias—. El stream de un modelo no es documentación, y tampoco es
// evidencia: es forense. Y `.specforge/` ya lo gitignorea `sf init`.
const Carpeta = ".specforge/lanzamientos"

// Los cuatro finales posibles. TODOS HABLAN DEL PROCESO Y NINGUNO DEL TRABAJO.
//
// Es la lección más cara de headless.md §7: Command Code devolvió `exit=0` y
// `subtype: success` SIN HABER HECHO NADA. Para un arnés, "éxito" significa que
// la conversación terminó bien, no que el trabajo se hizo. Quien decide si el
// trabajo vale es `sf done`.
const (
	FinTermino = "termino" // salió por su cuenta, con el código que sea
	FinTimeout = "timeout" // se agotó la espera y lo matamos
	FinError   = "error"   // no se pudo lanzar
)

// EsperaPorDefecto es cuánto se espera si nadie dice otra cosa.
//
// Sale de dos hechos medidos que tiran para lados opuestos: un modelo gratis de
// opencode estuvo 7m40 emitiendo CERO BYTES —así que un tope hace falta— y un
// lote de implementación de verdad puede tardar mucho más que eso trabajando
// bien —así que el tope no puede ser chico—. Media hora es holgado para lo
// segundo y corta lo primero.
const EsperaPorDefecto = 30 * time.Minute

// gracia es cuánto se le da al arnés entre el pedido de salida y el tiro final.
//
// Existe para que cierre su sesión y la deje resumible: los tres tienen
// `--resume`, y una sesión a medio cerrar no se recupera.
const gracia = 5 * time.Second

// Ficha es el parte de UNA corrida. Es lo que el bucle mira.
//
// ────────────────────────────────────────────────────────────────────────────
// LA REGLA DE LOS CAMPOS OPCIONALES
// ────────────────────────────────────────────────────────────────────────────
//
//	lo que el arnés no da, NO ESTÁ. No va en cero, no va en false, no va.
//
// Un costo de `0` y "este arnés no dice cuánto costó" son dos cosas distintas, y
// escribir la primera cuando pasa la segunda es mentir en un archivo que alguien
// va a leer para decidir con qué modelo sigue. Por eso son punteros: es la única
// forma de que `omitempty` distinga el cero del ausente.
type Ficha struct {
	ID       string `json:"id"`
	Estado   string `json:"estado"`
	Feature  string `json:"feature,omitempty"`
	Harness  string `json:"harness"`
	Alias    string `json:"alias,omitempty"`
	Modelo   string `json:"modelo"`
	Esfuerzo string `json:"esfuerzo,omitempty"`
	Skill    string `json:"skill,omitempty"`

	Empezo   string `json:"empezo"`
	DuroMs   int64  `json:"duro_ms"`
	Salida   int    `json:"salida"`
	Fin      string `json:"fin"`
	Registro string `json:"registro"`

	CostoUSD     *float64 `json:"costo_usd,omitempty"`
	Tokens       *Tokens  `json:"tokens,omitempty"`
	CargoLaSkill *bool    `json:"cargo_la_skill,omitempty"`
	Herramientas []string `json:"herramientas,omitempty"`

	// Dijo es la última palabra del arnés. NO ES EVIDENCIA: que diga "listo" no
	// cierra un paso.
	Dijo string `json:"dijo,omitempty"`

	// Error es lo que salió por stderr cuando algo no anduvo.
	//
	// No entra al registro porque lo rompería como JSONL, y no se tira porque es
	// lo primero que se mira cuando una corrida sale mal.
	Error string `json:"error,omitempty"`
}

// Opciones son las de la corrida.
type Opciones struct {
	Raiz    string
	Harness string

	// Espera es el tope. CERO ES SIN TOPE, no "sin tiempo".
	Espera time.Duration
}

// Correr lanza la línea, guarda el registro y devuelve la ficha.
//
// Devuelve error SÓLO si no se pudo lanzar o no se pudo escribir. Que el arnés
// salga con un código feo, o que se agote la espera, NO son errores de sf: son
// datos de la corrida, y viajan en la ficha.
func Correr(linea []string, f Ficha, o Opciones) (Ficha, error) {
	if len(linea) == 0 {
		f.Fin = FinError
		return f, errors.New("no hay línea que correr")
	}

	empezo := time.Now()
	f.Empezo = empezo.UTC().Format(time.RFC3339)
	f.ID = empezo.Format("2006-01-02T15-04-05") + "-" + f.Estado
	f.Registro = f.ID + ".jsonl"
	f.Fin = FinError

	dir := filepath.Join(o.Raiz, Carpeta)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return f, fmt.Errorf("creando %s: %w", Carpeta, err)
	}
	reg, err := os.Create(filepath.Join(dir, f.Registro))
	if err != nil {
		return f, fmt.Errorf("creando el registro: %w", err)
	}
	defer reg.Close()

	ctx := context.Background()
	cancelar := func() {}
	if o.Espera > 0 {
		ctx, cancelar = context.WithTimeout(ctx, o.Espera)
	}
	defer cancelar()

	cmd := exec.CommandContext(ctx, linea[0], linea[1:]...)
	cmd.Dir = o.Raiz

	// LA MARCA DEL HIJO, y es el único valor de `quien` en el que sf puede
	// confiar: los otros dos —`terminal` y `agente`— se deducen de señales que
	// el propio agente puede producir (medido el 2026-09-04: `script(1)`
	// fabrica un TTY). Éste lo pone sf mismo, acá.
	//
	// Va el ID de la ficha y no un `1` para que el registro pueda decir CUÁL
	// corrida hizo qué. Y se hereda el resto del entorno a propósito: sin
	// os.Environ() el hijo perdería PATH, HOME y las claves del proveedor.
	//
	// HOY SÓLO REGISTRA. Que una parada de Javier se RECHACE cuando viene de un
	// delegado es vecinos.md §2 ①, y espera los datos de las corridas.
	cmd.Env = append(os.Environ(), registro.VarDelegado+"="+f.ID)

	// Stdin nil es el dispositivo nulo, y hace falta: Claude Code espera tres
	// segundos por stdin y avisa "no stdin data received in 3s" si no se lo dan.
	cmd.Stdin = nil

	// El registro es SÓLO stdout: es el JSONL, y mezclarle stderr —donde los
	// plugins escriben— lo rompería como formato.
	cmd.Stdout = reg
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	// Al agotarse la espera: primero se PIDE que salga, para que el arnés cierre
	// su sesión y la deje resumible; y si en `gracia` sigue vivo, se lo mata.
	//
	// El error del Signal se traga a propósito: en Windows sólo Kill está
	// soportado, y ahí el que resuelve es WaitDelay. Devolver el error haría que
	// una corrida que SÍ se cortó bien pareciera rota.
	cmd.Cancel = func() error {
		_ = cmd.Process.Signal(os.Interrupt)
		return nil
	}
	cmd.WaitDelay = gracia

	errCorrida := cmd.Run()
	f.DuroMs = time.Since(empezo).Milliseconds()
	f.Error = recortar(stderr.String())

	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded):
		f.Fin = FinTimeout
		f.Salida = -1

	case errCorrida == nil:
		f.Fin = FinTermino

	default:
		var salida *exec.ExitError
		if errors.As(errCorrida, &salida) {
			// El arnés corrió y salió mal. Es un dato, no un error de sf: los
			// tres devuelven códigos distintos —Command Code usa el 4 para un
			// modelo retirado y el 8 para el tope de vueltas— y sf los ANOTA sin
			// interpretarlos.
			f.Fin = FinTermino
			f.Salida = salida.ExitCode()
		} else {
			// No se pudo lanzar: el binario no está, no hay permisos. Acá sí no
			// hay corrida de la que informar.
			f.Fin = FinError
			return f, guardar(dir, f, errCorrida)
		}
	}

	// El registro se lee del archivo y no de memoria: una corrida larga puede
	// dejar megabytes —Command Code emite un evento por token de pensamiento— y
	// no hay motivo para tenerlos todos en RAM.
	if abierto, err := os.Open(filepath.Join(dir, f.Registro)); err == nil {
		res := Resumir(o.Harness, abierto)
		abierto.Close()

		f.CostoUSD, f.Tokens, f.CargoLaSkill = res.CostoUSD, res.Tokens, res.CargoLaSkill
		f.Herramientas = res.Herramientas
		if res.Dijo != "" {
			f.Dijo = res.Dijo
		}
	}
	return f, guardar(dir, f, nil)
}

// guardar escribe la ficha al lado de su registro.
func guardar(dir string, f Ficha, previo error) error {
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dir, f.ID+".json"), append(b, '\n'), 0o644); err != nil {
		return err
	}
	return previo
}

// recortar deja el stderr en un tamaño que se pueda leer.
//
// Un arnés puede escupir miles de líneas de aviso, y la ficha es un parte, no un
// volcado: para el volcado está el registro.
func recortar(s string) string {
	const tope = 2000
	if len(s) <= tope {
		return s
	}
	return s[:tope] + "\n… (recortado)"
}
