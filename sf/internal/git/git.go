// Package git es shell-out a git, y nada más.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ SHELL-OUT Y NO UNA LIBRERÍA
// ────────────────────────────────────────────────────────────────────────────
//
// El CLI viejo trae go-git como dependencia. Acá se llama al binario, y es una
// decisión, no pereza:
//
//  1. **git ya está instalado.** Si no lo está, SpecForge no tiene nada que
//     hacer en ese proyecto — el flujo entero asume branches y commits.
//  2. **Una dependencia menos que migrar.** go-git son ~30 MB de código
//     transitivo para leer un hash.
//  3. **Es el mismo git que usa Javier.** Una librería reimplementa el
//     comportamiento; el binario ES el comportamiento. Cuando la constitución
//     diga `merge: no-ff`, se corre el mismo `--no-ff` que correría él.
//
// Es la preferencia de ingeniería ya establecida: usar la mejor
// implementación existente y probada en vez de escribir la propia.
//
// ────────────────────────────────────────────────────────────────────────────
// LO QUE ESTE PAQUETE NO HACE
// ────────────────────────────────────────────────────────────────────────────
//
// Por ahora sólo lee. Crear la branch (H9) y commitear (H10) son del paso 6 y
// del 4, y van a entrar acá — pero recién cuando el comando que los necesita se
// escriba, que es la regla de la construcción (construccion.md §2).
package git

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"slices"
	"strings"
)

// ErrNoEsRepo es que el directorio no está bajo control de git.
//
// No es un error fatal para todo sf: los cinco estados de producto funcionan
// igual en un directorio sin inicializar. Recién el ciclo de feature lo
// necesita, así que quien llama decide si le importa.
var ErrNoEsRepo = errors.New("no es un repositorio git")

// correr ejecuta git en un directorio y devuelve stdout limpio.
//
// Centraliza dos cosas que si no se repetirían en cada función: el manejo del
// error (git escribe el motivo en stderr, y sin esto se pierde) y el
// TrimSpace, porque git termina casi todo con un \n que nunca queremos.
func correr(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir

	// Output() captura stdout. El stderr viene aparte dentro del ExitError, y
	// es donde git explica qué pasó — sin rescatarlo, el error diría sólo
	// "exit status 128", que no le sirve a nadie.
	salida, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) && len(ee.Stderr) > 0 {
			return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), strings.TrimSpace(string(ee.Stderr)))
		}
		return "", fmt.Errorf("git %s: %w", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(salida)), nil
}

// EsRepo dice si el directorio está bajo git.
func EsRepo(dir string) bool {
	_, err := correr(dir, "rev-parse", "--git-dir")
	return err == nil
}

// Head devuelve el hash del commit actual, corto.
//
// Corto y no largo porque es lo que se guarda en el estado.json y lo que se le
// muestra a un humano. Los siete caracteres de git son suficientes para
// identificar un commit en cualquier repo de este tamaño, y un hash largo hace
// ilegible el `git diff` del estado.
func Head(dir string) (string, error) {
	return correr(dir, "rev-parse", "--short", "HEAD")
}

// BranchActual devuelve el nombre de la branch.
//
// Es la mitad de la compuerta del dolor #4: la otra mitad es el patrón de la
// constitución, y compararlos es un string contra otro.
func BranchActual(dir string) (string, error) {
	return correr(dir, "rev-parse", "--abbrev-ref", "HEAD")
}

// Diff devuelve el diff entre un commit y HEAD.
//
// Es el insumo de dos estados y en los dos sirve para lo mismo: mostrar EL
// CÓDIGO QUE QUEDÓ, no el que se planeó.
//
//	㉑ revision   ¿esto satisface los criterios?
//	㉓ cierre     la mitad técnica de la doc sale de acá
//
// El rango arranca en `base_commit`, que es el HEAD de cuando se terminó de
// planificar. Por eso el diff es exactamente el trabajo de esta feature y no
// arrastra lo que hicieron las otras.
func Diff(dir, desde string) (string, error) {
	if desde == "" {
		return "", errors.New("no hay base_commit: ¿se cerró la planificación?")
	}
	return correr(dir, "diff", desde+"..HEAD")
}

// Commit agrega todo lo que cambió y commitea con ese mensaje.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ LO HACE sf Y NO EL QUE IMPLEMENTÓ (H10)
// ────────────────────────────────────────────────────────────────────────────
//
// El dolor #2 es "el commit no se hace, hay que vigilarlo". Si el que
// implementa commitea y sf VERIFICA, el dolor sobrevive: se puede terminar sin
// commitear y sf lo único que hace es retarte después.
//
//	Hay que romper la posibilidad, no detectarla.
//
// R1 parte el trabajo en dos mitades limpias:
//
//	qué archivos, y cuándo   → una sola respuesta: lo del lote, ahora  → sf
//	qué dice el mensaje      → muchas: es prosa                        → el LLM
//
// Y así cerrar el lote ES commitear — no son dos cosas de las que una se puede
// olvidar. Los dolores #2 y #3 mueren juntos, porque el agrupamiento no lo
// decide nadie en este momento: se decidió en la planificación.
func Commit(dir, mensaje string) (string, error) {
	if strings.TrimSpace(mensaje) == "" {
		return "", errors.New("falta el mensaje del commit")
	}

	if _, err := correr(dir, "add", "-A"); err != nil {
		return "", err
	}

	// Si no quedó nada staged, no hay nada que commitear. Es un caso real —el
	// subagente dijo que terminó y no tocó un archivo— y hay que distinguirlo
	// del error de git, porque el mensaje útil es otro.
	if _, err := correr(dir, "diff", "--cached", "--quiet"); err == nil {
		return "", errors.New("no hay nada que commitear: ningún archivo cambió")
	}

	if _, err := correr(dir, "commit", "-m", mensaje); err != nil {
		return "", err
	}
	return Head(dir)
}

// HashDe devuelve un hash del contenido de varios archivos.
//
// Es lo que tapa el agujero astuto del dolor #8: sf ve rojo, el subagente
// trabaja, sf ve verde… y lo que cambió entre medio fue EL TEST. Se toma en el
// rojo y se compara en el verde.
//
// Usa `git hash-object`, que es exactamente para esto y ya está instalado: da
// el mismo hash que git le daría al archivo, sin que tengamos que elegir un
// algoritmo ni escribir el bucle.
func HashDe(dir string, rutas []string) (string, error) {
	if len(rutas) == 0 {
		return "", nil
	}
	// Ordenar antes de hashear: el mismo conjunto de archivos tiene que dar el
	// mismo hash aunque la lista venga en otro orden.
	orden := slices.Clone(rutas)
	slices.Sort(orden)

	salida, err := correr(dir, append([]string{"hash-object", "--"}, orden...)...)
	if err != nil {
		return "", err
	}
	// git devuelve un hash por línea; los juntamos y hasheamos el conjunto para
	// tener un solo string comparable.
	suma := sha256.Sum256([]byte(salida))
	return hex.EncodeToString(suma[:])[:12], nil
}

// DiffResumen es el mismo diff en una línea por archivo.
//
// Existe para el sobre: un diff completo de una feature grande puede ser
// enorme, y meterlo entero adentro del contexto de un subagente es justo lo que
// el diseño quiere evitar. El resumen le dice QUÉ MIRAR; los archivos los abre
// él con sus propias herramientas.
func DiffResumen(dir, desde string) (string, error) {
	if desde == "" {
		return "", errors.New("no hay base_commit: ¿se cerró la planificación?")
	}
	return correr(dir, "diff", "--stat", desde+"..HEAD")
}
