// Package roadmap lee el `roadmap.json`: el orden en que se hacen las features.
//
// ────────────────────────────────────────────────────────────────────────────
// EL "HUECO MÁS GRANDE" QUE RESULTÓ SER UN PARSER
// ────────────────────────────────────────────────────────────────────────────
//
// Este archivo figuraba como lo más grande por construir (que-sobrevive.md §13:
// "el roadmap.json en el CLI — el más grande, de ahí sale feature_actual"). Al
// trazar el bucle se desinfló solo (superficie-sf.md H19):
//
//	lo ESCRIBE el subagente del ⑩, como cualquier otro archivo
//	lo LEE sf, y nada más
//
// O sea que acá no hay un writer. Por eso este paquete es la mitad de largo que
// `estado`: sólo se lee.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ ES JSON Y NO PROSA, Y POR QUÉ ES TAN FLACO
// ────────────────────────────────────────────────────────────────────────────
//
// Es JSON porque sf tiene que consultarlo, y sf no tiene un LLM adentro: si el
// roadmap fuera prosa, marcar algo como hecho pediría un modelo que reescriba
// markdown, y ahí se corrompe (regla 1.2 de artefactos.md).
//
// Y JSON trae de regalo algo que la prosa no: OBLIGA A REFERENCIAR. No hay
// dónde copiar el texto de una historia, así que el roadmap no puede
// contradecir al us-#. Por eso guarda ids y nada más — sin títulos de historia,
// sin estados, sin descripciones (artefactos.md §8).
//
// Lo que NO está acá y podría parecer que falta:
//
//	el estado de cada feature   → estado.json. Acá se desincronizaría
//	las dependencias explícitas → el `orden` ya las expresa, y es un campo
//	                              menos que mantener
package roadmap

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// Archivo es dónde vive, relativo a la raíz del proyecto.
const Archivo = ".docs/roadmap.json"

// Roadmap es el archivo entero.
type Roadmap struct {
	// Actualizado es informativo: lo escribe el ⑩ y sf no lo usa para decidir
	// nada. Está para que un humano que abre el archivo sepa de cuándo es.
	Actualizado string `json:"actualizado"`

	// Features es slice y no mapa, al revés que en `estado`.
	//
	// La diferencia no es capricho: acá el archivo ES una secuencia y el orden
	// es su contenido principal. En estado.json el acceso siempre es por id, y
	// el orden no existe.
	Features []Feature `json:"features"`
}

// Feature es un grupo de historias que se entregan juntas.
//
// El criterio de agrupado no es "por prioridad" —eso no permite decir que un
// grupo esté mal armado— sino:
//
//	Una feature es un conjunto de historias que COMPARTEN SOLUCIÓN TÉCNICA
//	y se entregan juntas.                              (artefactos.md §3)
//
// Si tres historias necesitan tres soluciones distintas, no son una feature: el
// agrupamiento está mal.
type Feature struct {
	ID string `json:"id"` // "f-1"

	// Slug es la parte legible de la carpeta y de la branch.
	Slug string `json:"slug"` // "nucleo-cli"

	// Nombre es para mostrarle a un humano en `sf status`.
	Nombre string `json:"nombre"` // "núcleo del CLI"

	// Orden es la posición en la cola. Es el ÚNICO campo que expresa
	// dependencias: si us-5 necesita us-3, van en ese orden y listo.
	Orden int `json:"orden"`

	// Historias son ids de us-#, nunca su texto.
	Historias []string `json:"historias"`
}

// Carpeta es dónde viven los artefactos de esta feature.
//
//	.docs/features/f-1-nucleo-cli/
//	  decision.md · spec-design.md · tareas.json · revision.json · doc · journal
//
// El id solo no alcanza y el slug solo tampoco: el id ordena y es estable, el
// slug es el que hace que un `ls` se entienda. Del mismo par sale la branch
// (feat/f-1-nucleo-cli), que la arma internal/git con el patrón de la
// constitución.
func (f Feature) Carpeta() string {
	return filepath.Join(".docs", "features", f.ID+"-"+f.Slug)
}

// ErrNoHay es lo que devuelve Leer cuando todavía no se corrió el ⑩.
//
// No es un error de verdad: es el estado normal de un producto que tiene brief,
// PRD, constitución y backlog pero al que todavía nadie le puso orden. Quien
// llama lo distingue con errors.Is.
var ErrNoHay = errors.New("no hay roadmap.json: falta el ⑩")

// Leer carga el roadmap desde la raíz del proyecto, ya ordenado.
//
// Devolverlo ordenado acá y no en cada consulta es a propósito: el orden es la
// razón de ser de este archivo, y si cada uno que lo lee tuviera que acordarse
// de ordenarlo, alguna vez alguien no lo va a hacer.
func Leer(raiz string) (*Roadmap, error) {
	ruta := filepath.Join(raiz, Archivo)

	b, err := os.ReadFile(ruta)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNoHay
		}
		return nil, fmt.Errorf("leyendo %s: %w", ruta, err)
	}

	var r Roadmap
	if err := json.Unmarshal(b, &r); err != nil {
		return nil, fmt.Errorf("%s está corrupto: %w", ruta, err)
	}

	if err := r.validar(); err != nil {
		return nil, fmt.Errorf("%s: %w", ruta, err)
	}

	// sort.Slice ordena en el lugar con la función de comparación que le pasás.
	// El "menor" va primero, así que `Orden` creciente es lo que queremos.
	sort.Slice(r.Features, func(i, j int) bool {
		return r.Features[i].Orden < r.Features[j].Orden
	})

	return &r, nil
}

// validar chequea lo poco que puede romper a sf, y nada más.
//
// La tentación acá es validar todo —que el nombre no esté vacío, que las
// historias existan como archivo, que el slug no tenga mayúsculas— y es la
// tentación que infló la versión anterior. La vara:
//
//	se valida lo que hace que sf se comporte MAL EN SILENCIO.
//
// Un nombre vacío se ve feo en `sf status` y ya. Un id repetido, en cambio,
// hace que sf trabaje sobre la feature equivocada sin que nadie se entere.
func (r *Roadmap) validar() error {
	vistos := map[string]bool{}
	ordenes := map[int]string{}

	for _, f := range r.Features {
		if f.ID == "" {
			return errors.New("hay una feature sin id")
		}
		if vistos[f.ID] {
			return fmt.Errorf("el id %q está repetido", f.ID)
		}
		vistos[f.ID] = true

		// Dos features con el mismo `orden` dejan la cola indeterminada: sort
		// no promete nada sobre elementos que empatan, así que sf podría tomar
		// una distinta en cada corrida. Eso es exactamente "mal en silencio".
		if otro, hay := ordenes[f.Orden]; hay {
			return fmt.Errorf("%q y %q tienen el mismo orden (%d)", otro, f.ID, f.Orden)
		}
		ordenes[f.Orden] = f.ID

		// Una feature sin historias no tiene qué implementar. No es un error de
		// forma, es una feature que no existe.
		if len(f.Historias) == 0 {
			return fmt.Errorf("la feature %q no tiene historias", f.ID)
		}
	}
	return nil
}

// Buscar devuelve la feature con ese id.
func (r *Roadmap) Buscar(id string) (Feature, bool) {
	// Un roadmap que no existe no tiene features, así que la respuesta honesta
	// es "no está" y no un panic.
	//
	// El nil llega desde `Leer`, que lo devuelve cuando todavía no se corrió el
	// ⑩, y hay un estado que lo alcanza: un `estado.json` con `feature_actual`
	// puesto y el `roadmap.json` borrado a mano. Es raro, pero los dos comandos
	// que llaman acá —`sf done` y `sf approve`— ya saben contestar "no está en
	// el roadmap", y eso es exactamente lo que pasa.
	if r == nil {
		return Feature{}, false
	}
	for _, f := range r.Features {
		if f.ID == id {
			return f, true
		}
	}
	return Feature{}, false
}

// Proxima devuelve la primera feature de la cola que todavía no está cerrada.
//
// Es la mitad del ⑪ ("tomo del roadmap lo que hay que hacer"). La otra mitad la
// pone el estado, y por eso este paquete NO lee estado.json: recibe el conjunto
// de ids cerrados y no sabe de dónde salió.
//
// Mantenerlo así tiene un motivo concreto y no es purismo: los dos archivos se
// escriben en momentos distintos y por actores distintos —el roadmap lo escribe
// un subagente en el ⑩, el estado lo escribe sf— y si este paquete leyera los
// dos, cualquier cambio en el estado obligaría a tocarlo.
//
// El cruce de los dos vive en quien los tiene a ambos: `sf next`.
func (r *Roadmap) Proxima(cerradas map[string]bool) (Feature, bool) {
	for _, f := range r.Features {
		if !cerradas[f.ID] {
			return f, true
		}
	}
	return Feature{}, false
}
