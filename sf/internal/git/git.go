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
	"errors"
	"fmt"
	"os/exec"
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
