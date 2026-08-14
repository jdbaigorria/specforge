// Package constitucion lee la cabecera de `constitucion.md`.
//
// ────────────────────────────────────────────────────────────────────────────
// ES EL ÚNICO ARTEFACTO CUYO LECTOR PRINCIPAL NO ES UN HUMANO
// ────────────────────────────────────────────────────────────────────────────
//
// Ni un humano ni el hilo de conversación: es UN SUBAGENTE FRÍO, el
// implementador del ⑲, para el que la constitución es su manual. Por eso el
// cuerpo es técnico y no filosófico (R2: un artefacto tiene el tamaño de sus
// consumidores).
//
// De la cabecera sale casi todo lo que sf necesita para comprobar:
//
//	test_cmd               cómo correr los tests → sostiene TRES compuertas
//	manifiesto             dónde mirar las dependencias → el dolor #6
//	dependencias_aprobadas la lista que se construye sola
//	mutacion               la herramienta del ㉒
//	git                    el patrón de branch y el modo de merge → el dolor #4
//
// ────────────────────────────────────────────────────────────────────────────
// EL DOLOR #6 ESTABA MAL LEÍDO, Y POR ESO NO HAY LISTA BLANCA
// ────────────────────────────────────────────────────────────────────────────
//
//	"instaló librerías fuera de la constitución SIN CONSULTAR"
//
// El problema no es la librería prohibida: es no enterarte. Nadie puede listar
// de antemano lo que va a necesitar, y a la semana la lista queda vieja. Por eso
// sf no compara contra una lista blanca escrita a mano: compara el manifiesto
// contra lo aprobado hasta ahora, y lo nuevo lo AVISA.
//
//	La lista no se escribe de antemano: se construye sola, con cada aprobación.
package constitucion

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/frontmatter"
)

// Constitucion es la cabecera. El cuerpo no se parsea: lo lee el modelo.
type Constitucion struct {
	Lenguaje string `yaml:"lenguaje"`

	// Manifiesto es dónde declara sus dependencias el proyecto: go.mod,
	// package.json, Cargo.toml. sf no lo entiende — lo compara consigo mismo.
	Manifiesto string `yaml:"manifiesto"`

	// TestCmd es el agujero que se encontró mirando el CLI construido.
	//
	// Sin él sf NO PUEDE CORRER LOS TESTS, y correr los tests es lo que sostiene
	// tres compuertas: el rojo del ⑲, el verde del ⑳ y el conteo del #8. Era el
	// campo más barato de todo el recorrido: una línea. Y no se escribe a mano
	// —detectStack() lo llena solo— pero si falta hay que decirlo fuerte.
	TestCmd string `yaml:"test_cmd"`

	// Mutacion es la herramienta del ㉒, o vacío.
	//
	// Vacío no es un error: si el stack no tiene una herramienta buena, queda
	// sólo el modelo. La herramienta es una MEJORA, no un requisito — igual que
	// los subagentes.
	Mutacion string `yaml:"mutacion"`

	DependenciasAprobadas []string `yaml:"dependencias_aprobadas"`

	Git Git `yaml:"git"`
}

// Git son las reglas de Javier, que son las mismas en todos sus proyectos.
//
// Viven en la constitución y no cableadas en un skill: es R4 — el skill deja de
// tener convenciones propias y lee el dato del proyecto. Es lo que convirtió el
// "Squash merge default" de sfx-github en un campo.
type Git struct {
	BranchPorFeature bool   `yaml:"branch_por_feature"`
	PatronBranch     string `yaml:"patron_branch"` // "feat/{feature-id}-{slug}"
	Commit           string `yaml:"commit"`        // "conventional"
	Merge            string `yaml:"merge"`         // "no-ff"
}

// ErrNoHay es que todavía no se corrió el ⑧.
var ErrNoHay = errors.New("no hay constitucion.md: falta el ⑧")

// Leer carga la constitución desde la raíz del proyecto.
func Leer(raiz string) (*Constitucion, error) {
	ruta := filepath.Join(raiz, docs.Constitucion)

	var c Constitucion
	_, err := frontmatter.DeArchivo(ruta, &c)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoHay
		}
		return nil, fmt.Errorf("%s: %w", docs.Constitucion, err)
	}
	return &c, nil
}

// Branch arma el nombre de branch de una feature.
//
//	patron_branch: "feat/{feature-id}-{slug}"  +  f-2 · verificacion
//	                        →  feat/f-2-verificacion
//
// Esto es la mitad del dolor #4, y la otra mitad es que sf la CREE en vez de
// verificarla (H9): comparar es la mitad barata del trabajo, y con el patrón y
// la feature en la mano el nombre tiene una sola respuesta correcta.
func (c *Constitucion) Branch(id, slug string) string {
	patron := c.Git.PatronBranch
	if patron == "" {
		// Un default explícito y no un error: un proyecto sin `git:` en la
		// constitución sigue siendo usable, y este patrón es el que Javier usa
		// en todos lados igual.
		patron = "feat/{feature-id}-{slug}"
	}
	return strings.NewReplacer(
		"{feature-id}", id,
		"{slug}", slug,
	).Replace(patron)
}

// DependenciasNuevas devuelve las que están en el manifiesto y no aprobadas.
//
// sf NO entiende el formato del manifiesto: recibe la lista ya extraída y sólo
// compara dos conjuntos. Quién sabe leer un go.mod o un package.json es otro
// problema, y separarlo es lo que hace que esta parte sirva para cualquier
// lenguaje sin tocarla.
func (c *Constitucion) DependenciasNuevas(enElManifiesto []string) []string {
	var nuevas []string
	for _, d := range enElManifiesto {
		if !slices.Contains(c.DependenciasAprobadas, d) {
			nuevas = append(nuevas, d)
		}
	}
	return nuevas
}
