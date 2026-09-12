// Package suite corre los tests del proyecto y mira si existen.
//
// ────────────────────────────────────────────────────────────────────────────
// ES LA MITAD BARATA DE LA COMPUERTA MÁS IMPORTANTE DEL DISEÑO
// ────────────────────────────────────────────────────────────────────────────
//
// El ⑲ tiene tres tiempos y el tercero es "comprobá que fallan". Hoy eso es UNA
// PROMESA DEL MODELO: contesta "sí, fallan" y nadie mira.
//
// Pero es un HECHO, y la lista exacta de tests del lote ya está en tareas.json.
// Así que sf lo mira con sus propios ojos, antes de dejar implementar:
//
//	Un test que pasa antes de que exista el código es un test de mentira.
//
// Es el dolor #5 (los mocks) y el #8 (verde sin sustancia) atacados por un EXIT
// CODE, sin juicio y sin agregar un solo artefacto.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ EL EXIT CODE ALCANZA, Y NO SE PARSEA LA SALIDA
// ────────────────────────────────────────────────────────────────────────────
//
// La tentación es leer la salida del runner y decir exactamente cuáles de los 6
// tests del lote fallaron. No se hace, y no por pereza: cada runner imprime
// distinto —go test, pytest, jest, cargo test, rspec— y un parser por runner es
// una lista que hay que mantener y que se rompe con cada versión nueva.
//
// El diseño ya lo había resuelto: "sf no deja pasar al verde sin haber visto el
// rojo con sus propios ojos, y LE ALCANZA UN EXIT CODE" (maquina-estados.md §4).
//
// Lo que sí es genérico y sí se hace: comprobar que cada test PLANIFICADO exista
// en su archivo. Eso no depende del runner — depende de que el nombre esté
// escrito en el archivo, y eso vale en cualquier lenguaje.
//
// ────────────────────────────────────────────────────────────────────────────
// Y HAY UN TERCER HECHO, QUE TAMPOCO ES UN PARSER: ¿EL ROJO ES DE ELLOS?
// ────────────────────────────────────────────────────────────────────────────
//
// El exit code dice que algo falló. NO dice que haya fallado lo que el lote
// vino a escribir, y ese hueco era real:
//
//	la suite ya venía roja por otra cosa   →  el rojo del lote sale gratis
//	un test viejo quedó fallando           →  ídem
//	un paquete no compila                  →  ídem
//
// Con eso, `Faltantes` prueba que los tests EXISTEN y el exit code prueba que
// ALGO falla — y entre las dos cosas nunca se prueba que fallen ÉSTOS. El hash
// del rojo se toma igual, y toda la cadena rojo→verde queda apoyada en un
// fallo que no era del lote.
//
// `Nombrados` lo cierra sin volverse un parser, con la misma jugada que ya
// bendijo el ② de vecinos.md: no se estructura la salida, se busca una marca.
// Y acá la marca no hay ni que tabularla por runner, porque ya la tenemos — es
// el nombre del test. TODOS los runners nombran lo que falla; ninguno dice
// "falló un test" sin decir cuál.
//
// Es el MISMO Contains de `Faltantes`, movido de blanco: allá se busca el
// nombre adentro del archivo, acá adentro de la salida.
package suite

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
)

// Resultado es lo que pasó al correr la suite.
type Resultado struct {
	// Verde es si el runner salió con 0.
	Verde bool

	// Salida es lo que imprimió, para mostrárselo a quien tenga que arreglarlo.
	Salida string
}

// Correr ejecuta el `test_cmd` de la constitución en la raíz del proyecto.
//
// El comando viene como una línea (`go test ./...`) y se ejecuta con la shell,
// no partiéndolo por espacios. Es a propósito: los test_cmd reales llevan pipes,
// variables y flags con comillas —`pytest -q 2>&1 | tail -20`— y partir por
// espacios los rompe en silencio.
func Correr(raiz, testCmd string) (Resultado, error) {
	if strings.TrimSpace(testCmd) == "" {
		return Resultado{}, errors.New("la constitución no tiene `test_cmd:`")
	}

	cmd := exec.Command("sh", "-c", testCmd)
	cmd.Dir = raiz

	// CombinedOutput junta stdout y stderr: los runners escriben los fallos en
	// cualquiera de los dos según cuál sea, y separarlos sólo daría la mitad de
	// la historia a quien tenga que arreglarlo.
	salida, err := cmd.CombinedOutput()

	r := Resultado{Salida: string(salida)}

	if err == nil {
		r.Verde = true
		return r, nil
	}
	// Un ExitError es "los tests fallaron", que es un resultado NORMAL y a veces
	// el que queremos. Cualquier otro error —no existe sh, no se pudo ejecutar—
	// sí es un problema de verdad y se propaga.
	var ee *exec.ExitError
	if errors.As(err, &ee) {
		return r, nil
	}
	return r, err
}

// Faltantes devuelve los tests planificados que NO existen.
//
// Los ids vienen de tareas.json con la forma `ruta/al/archivo_test.go::NombreDelTest`.
// Existe significa dos cosas, y las dos son verificables sin saber el lenguaje:
//
//	el archivo está
//	el nombre del test aparece adentro
//
// Buscar el nombre es un Contains sobre un archivo que ya sabemos cuál es. Acá
// NO se usa ripgrep, y la distinción es real: rg gana cuando hay que barrer un
// repo entero; para mirar dentro de un archivo conocido es una dependencia
// externa que puede no estar instalada, a cambio de nada.
func Faltantes(raiz string, tests []string) []string {
	var faltan []string

	// Los archivos se leen una sola vez aunque tengan varios tests: un lote
	// suele concentrar sus tests en dos o tres archivos.
	cache := map[string]string{}

	for _, t := range tests {
		archivo, nombre := Partir(t)
		if archivo == "" {
			faltan = append(faltan, t+" (no tiene la forma archivo::Test)")
			continue
		}

		contenido, leido := cache[archivo]
		if !leido {
			b, err := os.ReadFile(filepath.Join(raiz, archivo))
			if err != nil {
				faltan = append(faltan, t+" (no existe el archivo)")
				cache[archivo] = ""
				continue
			}
			contenido = string(b)
			cache[archivo] = contenido
		}
		if contenido == "" {
			faltan = append(faltan, t+" (no existe el archivo)")
			continue
		}

		// Sin nombre, la tarea declaró el archivo entero: alcanza con que exista.
		if nombre != "" && !strings.Contains(contenido, nombre) {
			faltan = append(faltan, t)
		}
	}
	return faltan
}

// Nombrados devuelve los tests planificados que la salida MENCIONA.
//
// Es la contracara de Faltantes: aquélla busca el nombre adentro del archivo,
// ésta adentro de lo que imprimió el runner. Las dos son el mismo Contains
// sobre un string que ya tenemos, y ninguna sabe en qué lenguaje está escrito
// el proyecto.
//
// ────────────────────────────────────────────────────────────────────────────
// QUÉ SIGNIFICA UN RESULTADO VACÍO, Y QUÉ NO
// ────────────────────────────────────────────────────────────────────────────
//
// Devolver vacío significa "la salida no habla de estos tests". En el rojo eso
// es fuerte: si el runner nombra lo que falla —y los nombra todos— una salida
// que no menciona ninguno de los planificados está contando OTRO fallo.
//
// Lo que NO significa es "estos tests pasaron". Hay una manera legítima de caer
// acá con el lote en orden, y es que la salida venga recortada: un `test_cmd`
// con `| tail -20` puede dejar afuera justo las líneas que los nombran. Por eso
// el que llama decide qué hacer con el vacío; esta función sólo cuenta.
//
// Al revés no es simétrico y conviene decirlo: que un nombre SÍ aparezca no
// prueba que ese test haya fallado —un runner verboso también imprime los que
// pasan—. Sirve igual, porque el caso que se quiere atajar es el otro: el rojo
// que no tiene nada que ver con el lote.
//
// Un test declarado sin nombre (`archivo_test.go`, sin `::`) se busca por su
// ruta y también por el nombre del archivo solo: los runners imprimen las rutas
// con separadores distintos, y el basename es la parte que sobrevive a todos.
func Nombrados(salida string, tests []string) []string {
	var hay []string
	for _, t := range tests {
		archivo, nombre := Partir(t)

		var marca string
		switch {
		case nombre != "":
			marca = nombre
		case archivo != "":
			marca = filepath.Base(archivo)
		default:
			continue
		}

		if strings.Contains(salida, marca) && !slices.Contains(hay, t) {
			hay = append(hay, t)
		}
	}
	return hay
}

// Archivos devuelve los archivos de test de una lista, sin repetir y ordenados.
//
// Es lo que se hashea para tapar el agujero astuto del #8: sf ve rojo, el
// subagente trabaja, sf ve verde… y lo que cambió entre medio fue EL TEST.
func Archivos(tests []string) []string {
	var r []string
	for _, t := range tests {
		if a, _ := Partir(t); a != "" && !slices.Contains(r, a) {
			r = append(r, a)
		}
	}
	slices.Sort(r)
	return r
}

// Partir separa `archivo::Test` en sus dos mitades.
//
// Si no trae `::`, se asume que todo es el archivo y no hay nombre. Eso permite
// que una tarea declare "este archivo entero" sin inventar sintaxis nueva, y
// que Faltantes lo trate igual: el archivo tiene que existir.
func Partir(id string) (archivo, nombre string) {
	archivo, nombre, hay := strings.Cut(id, "::")
	if !hay {
		return archivo, ""
	}
	return archivo, nombre
}
