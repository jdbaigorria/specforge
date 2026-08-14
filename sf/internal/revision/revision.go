// Package revision lee el `revision.json`: el informe del ㉑ y del ㉒.
//
// ────────────────────────────────────────────────────────────────────────────
// SON UN ARCHIVO, NO DOS
// ────────────────────────────────────────────────────────────────────────────
//
// El ㉑ (¿satisface los criterios?) y el ㉒ (los mutantes) parecen dos cosas y
// salen de leer el flujo que son una: los hace EL MISMO ACTOR, en LA MISMA
// PASADA, y cuando hay un arreglo SE REHACEN LOS DOS JUNTOS.
//
// ────────────────────────────────────────────────────────────────────────────
// Y VA EN JSON, NO EN MARKDOWN
// ────────────────────────────────────────────────────────────────────────────
//
// El argumento es duro: el hallazgo nace `abierto` y después cambia de estado.
// Ese cambio ocurre DESPUÉS de que el revisor escribió el archivo, y no lo hace
// él. En markdown, marcar un hallazgo exigiría un modelo que reescriba prosa —
// y ahí se corrompe. En JSON es un campo (regla 1.2).
//
// ────────────────────────────────────────────────────────────────────────────
// LOS HALLAZGOS NO TIENEN CICLO DE VIDA (H14)
// ────────────────────────────────────────────────────────────────────────────
//
// La primera versión del diseño les daba tres estados: abierto → arreglado →
// descartado. Y dejaba sin contestar quién marca `arreglado`. Las dos respuestas
// posibles eran malas: si lo marca el implementador, sf le cree al que trabajó;
// si lo marca sf, está juzgando si un arreglo arregló.
//
// La respuesta ya estaba escrita en otro lado: la revisión SE REHACE ENTERA.
//
//	Entonces nadie marca nada. La vuelta 2 escribe un revision.json nuevo, y
//	h-1 simplemente NO REAPARECE. Eso *es* estar arreglado.
//
// Queda un solo estado que no se deduce de nada: `descartado`. Lo pone Javier
// —"esto el revisor lo marcó y no me importa"— y sin persistirlo la vuelta 3 lo
// vuelve a encontrar, porque el código sigue igual.
package revision

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
)

// Abierto y Descartado son los dos estados reales de un hallazgo.
const (
	Abierto    = "abierto"
	Descartado = "descartado"
)

// Revision es el archivo entero.
type Revision struct {
	Feature string `json:"feature"`

	// Vuelta es cuántas veces se rehizo la revisión de esta feature.
	//
	// Es el mismo truco que `intentos_fallidos`: si el ㉑ va por la cuarta, eso
	// no es ruido — es que la planificación se quedó corta.
	Vuelta int `json:"vuelta"`

	Veredicto string `json:"veredicto"` // "limpio" | "con-hallazgos"

	// Criterios es el veredicto por criterio: "us-1/CA-1" → "cumple".
	//
	// Mapa y no lista porque la pregunta real siempre es "¿opinó sobre ESTE?",
	// y eso es lo que sf cuenta para que muera el dolor #7.
	Criterios map[string]string `json:"criterios"`

	Mutantes  Mutantes   `json:"mutantes"`
	Hallazgos []Hallazgo `json:"hallazgos"`
}

// Mutantes es el resultado del ㉒: el número, no los parches.
//
// De los mutantes se guarda el RESULTADO y no los parches, y hay razón: los
// genera un modelo distinto cada vez —no son reproducibles ni estables— y una
// vez arreglado el test, el parche no se vuelve a aplicar nunca.
//
// Lo que sí importa es el score, porque es un número comparable entre vueltas:
//
//	mutation score 71% (vuelta 1) → 88% (vuelta 2)
type Mutantes struct {
	Herramienta   string  `json:"herramienta"`
	Score         float64 `json:"score"`
	Sobrevivieron int     `json:"sobrevivieron"`
}

// Hallazgo es algo que el revisor encontró.
type Hallazgo struct {
	ID string `json:"id"` // "h-1"

	// Origen es 21 o 22: de cuál de los dos pasos salió.
	Origen int `json:"origen"`

	Criterio string `json:"criterio"` // "us-3/CA-2", o vacío si es del ㉒

	// Estado es Abierto o Descartado. No existe "arreglado": eso es no
	// reaparecer en la vuelta siguiente.
	Estado string `json:"estado"`

	Detalle string `json:"detalle"`

	// Motivo es por qué Javier lo descartó. Sólo tiene sentido con Descartado.
	Motivo string `json:"motivo,omitempty"`
}

// ErrNoHay es que la revisión todavía no se escribió.
var ErrNoHay = errors.New("no hay revision.json: falta el ㉑")

// Leer carga la revisión desde su ruta completa.
func Leer(ruta string) (*Revision, error) {
	b, err := os.ReadFile(ruta)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoHay
		}
		return nil, fmt.Errorf("leyendo %s: %w", ruta, err)
	}

	var r Revision
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("%s está corrupto: %w", ruta, err)
	}
	return &r, nil
}

// Guardar escribe la revisión. Lo usa `sf dismiss`, que es el único comando que
// la toca después de que el revisor la escribió.
func (r *Revision) Guardar(ruta string) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(ruta, append(b, '\n'), 0o644)
}

// Abiertos son los hallazgos que todavía frenan.
func (r *Revision) Abiertos() []Hallazgo {
	var a []Hallazgo
	for _, h := range r.Hallazgos {
		if h.Estado == Abierto {
			a = append(a, h)
		}
	}
	return a
}

// Buscar devuelve un hallazgo por id, para poder mutarlo.
//
// Devuelve puntero al elemento del slice —no una copia— porque quien lo llama
// (`sf dismiss`) necesita cambiarle el estado y que eso quede en el archivo.
func (r *Revision) Buscar(id string) (*Hallazgo, bool) {
	for i := range r.Hallazgos {
		if r.Hallazgos[i].ID == id {
			return &r.Hallazgos[i], true
		}
	}
	return nil, false
}
