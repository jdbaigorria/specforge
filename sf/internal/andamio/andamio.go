// Package andamio es `sf install`: lo que hace que SpecForge se pueda usar.
//
// ────────────────────────────────────────────────────────────────────────────
// LA DIFERENCIA ENTRE "FUNCIONA" Y "SE PUEDE INSTALAR"
// ────────────────────────────────────────────────────────────────────────────
//
// El bucle ya cerraba de punta a punta antes de este paquete, pero sólo si
// alguien armaba las dos cosas a mano:
//
//	el orquestador   CLAUDE.md y AGENTS.md en el proyecto
//	~/.specforge/    qué modelos hay y cómo se invoca cada uno
//
// `sf install` es lo que las pone. Y no es de la máquina —no mira el estado ni
// lo mueve—: es andamio, la categoría que `que-sobrevive.md` §1 tuvo que
// declarar aparte porque la vara "¿lo consume alguno de los 9 estados?" la
// dejaba afuera sin ser inútil.
//
// ────────────────────────────────────────────────────────────────────────────
// EL ORQUESTADOR VIAJA EMBEBIDO EN EL BINARIO
// ────────────────────────────────────────────────────────────────────────────
//
// `plantillas/CLAUDE.md` se compila ADENTRO de `sf` con go:embed. La
// alternativa —buscar el archivo en disco relativo al binario— se cae sola: un
// `sf` instalado con `go install` vive en ~/go/bin y no tiene el repo al lado.
//
// Un archivo, dos destinos: se escribe UNA vez y se copia DOS veces, porque el
// diseño dice que AGENTS.md es el mismo texto y no una traducción
// (superficie-sf.md §6). Mantenerlo como dos archivos en el repo sería el mismo
// contenido en dos lugares — justo lo que el proyecto evita en todo lo demás.
package andamio

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// orquestador es plantillas/CLAUDE.md, embebido en tiempo de compilación.
//
// La ruta sube dos niveles porque este paquete vive en sf/internal/andamio/ y
// la plantilla en la raíz del repo. Si alguien mueve una de las dos, esto NO
// compila — que es exactamente lo que uno quiere: un puntero roto a la
// plantilla del orquestador tiene que romper el build, no descubrirse en la
// máquina de un usuario.
//
//go:embed plantilla/CLAUDE.md
var orquestador []byte

// Los dos nombres del mismo texto.
//
// Son dos harness distintos leyendo lo mismo: lo único que cambia entre ellos
// es cómo se lanza un subagente, y eso el harness ya lo sabe hacer.
var destinos = []string{"CLAUDE.md", "AGENTS.md"}

// Resultado es qué hizo la instalación.
type Resultado struct {
	// Escritos son los archivos que se crearon o actualizaron.
	Escritos []string

	// Salteados son los que ya existían y NO se pisaron, con su motivo.
	Salteados []string

	// Harness es dónde se detectó (o lo que se pidió con --harness).
	Harness string

	// Modelos es cuántos quedaron declarados.
	Modelos int
}

// Opciones son las del comando.
type Opciones struct {
	// Harness pisa la detección. Vacío = detectar.
	Harness string

	// Forzar pisa los archivos del proyecto que ya existan.
	//
	// Existe porque el caso normal es NO pisar —el CLAUDE.md de un proyecto
	// suele tener cosas de Javier— pero actualizar el orquestador cuando cambia
	// tiene que ser posible sin borrar a mano.
	Forzar bool
}

// ErrSinProyecto es que no se puede instalar el orquestador acá.
var ErrSinProyecto = errors.New("no encuentro el proyecto")

// Instalar pone el orquestador en el proyecto y arma ~/.specforge/.
//
// Las dos mitades son independientes a propósito: se puede correr en una
// máquina nueva sobre un proyecto que ya tenía CLAUDE.md, o en un proyecto
// nuevo con el ~/.specforge/ ya armado. Ninguna pisa lo que la otra hizo.
func Instalar(raiz string, o Opciones) (*Resultado, error) {
	r := &Resultado{}

	// ① el orquestador, en los dos nombres
	for _, nombre := range destinos {
		ruta := filepath.Join(raiz, nombre)

		if _, err := os.Stat(ruta); err == nil && !o.Forzar {
			// NO se pisa, y no es timidez: el CLAUDE.md de un proyecto suele
			// tener instrucciones propias de Javier. Pisarlo sin avisar sería
			// borrarle trabajo para poner cuatro líneas.
			r.Salteados = append(r.Salteados, nombre+" (ya existe — `--forzar` lo pisa)")
			continue
		}

		if err := os.WriteFile(ruta, orquestador, 0o644); err != nil {
			return nil, fmt.Errorf("escribiendo %s: %w", nombre, err)
		}
		r.Escritos = append(r.Escritos, nombre)
	}

	// ② ~/.specforge/
	//
	// Si ya existe NO se toca: ahí viven los modelos que Javier fue aprobando
	// una aprobación a la vez, y re-sembrarlo sería tirar esa lista.
	g, err := global.Leer()
	switch {
	case errors.Is(err, global.ErrNoHay):
		harness := o.Harness
		if harness == "" {
			harness = global.DetectarHarness()
		}
		g = global.Semilla(harness)
		if err := g.Guardar(); err != nil {
			return nil, err
		}
		r.Escritos = append(r.Escritos, "~/.specforge/"+global.Archivo)

	case err != nil:
		return nil, err

	default:
		// Ya existía. Lo único que se actualiza es el harness, y sólo si lo
		// pidieron explícito: es el caso de "me mudé de harness", y es la única
		// forma de corregir una detección que salió mal.
		if o.Harness != "" && o.Harness != g.Harness {
			g.Harness = o.Harness
			if err := g.Guardar(); err != nil {
				return nil, err
			}
			r.Escritos = append(r.Escritos, "~/.specforge/"+global.Archivo+" (harness → "+o.Harness+")")
		} else {
			r.Salteados = append(r.Salteados, "~/.specforge/"+global.Archivo+" (ya existe — tus modelos quedan)")
		}
	}

	r.Harness = g.Harness
	r.Modelos = len(g.Modelos)
	return r, nil
}

// Desinstalar saca el orquestador del proyecto.
//
// NO toca ~/.specforge/, y es a propósito: los modelos son de la máquina, no
// del proyecto. Sacar SpecForge de un repo no puede borrarle a Javier la lista
// que fue construyendo en los otros quince.
func Desinstalar(raiz string) (*Resultado, error) {
	r := &Resultado{}
	for _, nombre := range destinos {
		ruta := filepath.Join(raiz, nombre)

		// Sólo se borra lo que escribimos nosotros. Un CLAUDE.md que alguien
		// editó tiene cosas que no pusimos, y borrarlo sería tirar eso.
		b, err := os.ReadFile(ruta)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		if string(b) != string(orquestador) {
			r.Salteados = append(r.Salteados, nombre+" (lo editaste — no lo toco)")
			continue
		}
		if err := os.Remove(ruta); err != nil {
			return nil, err
		}
		r.Escritos = append(r.Escritos, nombre+" (borrado)")
	}
	return r, nil
}
