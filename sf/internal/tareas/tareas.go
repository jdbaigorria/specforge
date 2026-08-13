// Package tareas lee el `tareas.json`: las tareas de una feature y sus lotes.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ LAS TAREAS Y LOS TESTS VIVEN EN EL MISMO ARCHIVO
// ────────────────────────────────────────────────────────────────────────────
//
// El ⑭ (las tareas) y el ⑮ (los tests) son dos pasos del flujo y UN artefacto.
// El motivo es el consumidor: el ⑲ trabaja por lote —crea los tests de ese
// lote, los corre, los ve fallar, implementa— y necesita las dos cosas en la
// misma lectura. En dos archivos habría que sincronizarlos a mano y sf tendría
// que cruzarlos para chequear (artefactos.md §9).
//
// ────────────────────────────────────────────────────────────────────────────
// PLANO, NO ANIDADO: EL LOTE ES UN CAMPO
// ────────────────────────────────────────────────────────────────────────────
//
// La forma obvia sería anidar (lotes → tareas). No se hace, y la razón es que
// las dos operaciones reales son más baratas planas:
//
//	"dame las tareas del lote 2"   → un filtro
//	"mové esta tarea al lote 3"    → cambiar un número
//
// Anidado, lo segundo es mover un elemento entre dos listas.
//
// ────────────────────────────────────────────────────────────────────────────
// ESTE ARCHIVO TACHA TRES DOLORES, Y NINGUNO NECESITA QUE sf PIENSE
// ────────────────────────────────────────────────────────────────────────────
//
//	#8  "todo verde" sin tests   sf tiene la LISTA EXACTA de tests que deben
//	                             existir. Buscarlos es un rg; ver si corrieron
//	                             es leer la salida del runner.
//	#2  el commit no se hace     el lote es la unidad de commit
//	#3  commits mal agrupados    el agrupamiento YA SE DECIDIÓ acá, en la
//	                             planificación. No hay que agrupar bien: hay
//	                             que commitear en el momento correcto.
//
// Y la feature a medias se ataca ANTES de empezar: cada tarea declara qué
// criterios satisface, así que sf cuenta y avisa si quedaron sin cubrir.
package tareas

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

// Archivo es el nombre dentro de la carpeta de la feature.
const Archivo = "tareas.json"

// Plan es el archivo entero.
type Plan struct {
	Feature string  `json:"feature"`
	Tareas  []Tarea `json:"tareas"`
}

// Tarea es una unidad de trabajo con sus tests ya nombrados.
type Tarea struct {
	ID string `json:"id"` // "t-1"

	// Lote es el número de lote. Es un campo y no un nivel de anidamiento.
	Lote int `json:"lote"`

	Descripcion string `json:"descripcion"`

	// Satisface son criterios de aceptación, en formato "us-1/CA-2".
	//
	// Es lo que permite contar antes de implementar: "la feature tiene 9
	// criterios, las tareas cubren 7, sin cubrir us-7/CA-1 y us-7/CA-4".
	//
	// No se pisa con el ㉑ porque son preguntas distintas en momentos
	// distintos: acá es "¿hay tarea para cada criterio?" y lo CUENTA sf; allá
	// es "¿el código satisface cada criterio?" y lo JUZGA un modelo.
	Satisface []string `json:"satisface"`

	// Tests son los nombres exactos de los tests que tienen que existir.
	//
	//	"internal/docs/brief_test.go::TestParseFrontmatter"
	//
	// Esta lista es toda la defensa contra el dolor #8, y es la que hace que
	// `sf lote start` pueda exigir el rojo sobre tests concretos en vez de
	// sobre "los que haya".
	Tests []string `json:"tests"`
}

// ErrNoHay es que la feature todavía no tiene plan.
var ErrNoHay = errors.New("no hay tareas.json: falta el ⑭")

// Leer carga el plan desde la carpeta de una feature.
//
// Recibe la carpeta y no la raíz + el id, porque quien llama ya la tiene
// resuelta: sale de roadmap.Feature.Carpeta(), que es el único que sabe armarla
// (necesita el id Y el slug).
func Leer(raiz, carpeta string) (*Plan, error) {
	ruta := filepath.Join(raiz, carpeta, Archivo)

	b, err := os.ReadFile(ruta)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoHay
		}
		return nil, fmt.Errorf("leyendo %s: %w", ruta, err)
	}

	var p Plan
	if err := json.Unmarshal(b, &p); err != nil {
		return nil, fmt.Errorf("%s está corrupto: %w", ruta, err)
	}
	return &p, nil
}

// DelLote devuelve las tareas de un lote, en el orden en que están.
//
// Es lo que hace que `sf context` pueda servir SÓLO el lote que toca. El
// implementador no tiene por qué ver los otros tres: gasta contexto y puede
// adelantarse a trabajo que todavía no le toca.
func (p *Plan) DelLote(n int) []Tarea {
	var r []Tarea
	for _, t := range p.Tareas {
		if t.Lote == n {
			r = append(r, t)
		}
	}
	return r
}

// Lotes devuelve los números de lote que existen, ordenados y sin repetir.
//
// Sirve para dos cosas que sf hace seguido: saber cuántos son ("lote 2 de 4") y
// saber cuál es el próximo. El plan puede venir con los lotes en cualquier
// orden —lo escribe un LLM— así que se ordena acá y no se confía en el archivo.
func (p *Plan) Lotes() []int {
	var r []int
	for _, t := range p.Tareas {
		if !slices.Contains(r, t.Lote) {
			r = append(r, t.Lote)
		}
	}
	slices.Sort(r)
	return r
}

// TestsDelLote junta todos los tests planificados de un lote.
//
// Es la lista que `sf lote start` va a exigir que exista y que falle, y la que
// `sf done` va a exigir que pase. Los dos extremos de la compuerta del medio
// salen de la misma línea del archivo.
func (p *Plan) TestsDelLote(n int) []string {
	var r []string
	for _, t := range p.DelLote(n) {
		for _, test := range t.Tests {
			if !slices.Contains(r, test) {
				r = append(r, test)
			}
		}
	}
	return r
}

// CriteriosCubiertos junta los criterios que las tareas dicen satisfacer.
//
// La comparación contra los criterios que EXISTEN la hace quien llama: este
// paquete lee tareas.json y nada más. Los criterios viven en los us-#, que son
// otro archivo y otro paquete.
func (p *Plan) CriteriosCubiertos() []string {
	var r []string
	for _, t := range p.Tareas {
		for _, c := range t.Satisface {
			if !slices.Contains(r, c) {
				r = append(r, c)
			}
		}
	}
	slices.Sort(r)
	return r
}
