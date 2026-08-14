// Package frontmatter parte un `.md` en cabecera y cuerpo.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ CASI TODO ARTEFACTO TIENE LAS DOS COSAS
// ────────────────────────────────────────────────────────────────────────────
//
//	---
//	lo que sf necesita leer     ← estructurado, cinco líneas
//	---
//	lo que el modelo lee        ← prosa
//
// La cabecera es de sf y el cuerpo es del modelo, y por eso no hacen falta
// archivos paralelos (artefactos.md §1.3). Javier: "no sólo me cierra, creo que
// es el camino correcto, porque después se puede ir hacia un LLM wiki como el
// que proponía Karpathy".
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ yaml.v3 Y NO UN PARSER PROPIO
// ────────────────────────────────────────────────────────────────────────────
//
// La cabecera parece simple —cinco claves y una lista— y escribir el parser
// sería media hora. No se hace, por la misma regla que hizo que `git` sea
// shell-out: la cabecera la escribe un LLM, y un LLM escribe YAML de verdad —
// comillas, listas en línea, valores multilínea, comentarios. Un parser
// casero acierta el 90% de los casos y falla en silencio en el 10% restante.
//
// Es la única dependencia del proyecto, y es la librería estándar de facto.
package frontmatter

import (
	"bytes"
	"errors"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// separador es la línea que abre y cierra la cabecera.
var separador = []byte("---")

// ErrSinCabecera es que el archivo no tiene frontmatter.
//
// Es un error y no un caso normal: todos los artefactos del flujo la llevan, y
// uno sin ella salió mal escrito. Decirlo con nombre propio permite que quien
// llama distinga "está mal escrito" de "el YAML tiene un error de sintaxis".
var ErrSinCabecera = errors.New("el archivo no tiene frontmatter")

// Partir separa la cabecera del cuerpo.
//
// `cabecera` se deserializa sobre `destino`, que tiene que ser un puntero a un
// struct con etiquetas `yaml:"..."`. Devuelve el cuerpo por si alguien lo
// necesita — hoy lo usa quien busca los criterios de aceptación adentro.
func Partir(contenido []byte, destino any) (cuerpo []byte, err error) {
	// El frontmatter arranca en la PRIMERA línea. Si no, no es frontmatter:
	// es un `---` cualquiera en el medio de la prosa, y tratarlo como cabecera
	// partiría el archivo por un lugar arbitrario.
	if !bytes.HasPrefix(contenido, separador) {
		return contenido, ErrSinCabecera
	}

	resto := contenido[len(separador):]

	// El cierre es el próximo "\n---" — buscarlo desde el principio del resto
	// evita que un "---" pegado a texto (como una línea de guiones larga) lo
	// confunda.
	fin := bytes.Index(resto, append([]byte("\n"), separador...))
	if fin < 0 {
		return contenido, fmt.Errorf("%w: no se cierra", ErrSinCabecera)
	}

	cabecera := resto[:fin]
	cuerpo = resto[fin+len(separador)+1:]

	if destino != nil {
		if err := yaml.Unmarshal(cabecera, destino); err != nil {
			return cuerpo, fmt.Errorf("el frontmatter no es YAML válido: %w", err)
		}
	}
	return cuerpo, nil
}

// DeArchivo es Partir sobre un archivo.
func DeArchivo(ruta string, destino any) (cuerpo []byte, err error) {
	b, err := os.ReadFile(ruta)
	if err != nil {
		return nil, err
	}
	return Partir(b, destino)
}
