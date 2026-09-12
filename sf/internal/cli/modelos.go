package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/modelos"
)

// listarModelos es `sf models`: los ids que el harness dice tener.
//
// Como `sf install` y `sf doctor`, NO es de la máquina: no mira el estado ni lo
// mueve. Existe porque hoy, para declarar un modelo, hay que ir a buscar el id a
// otra ventana — `sf doctor` literalmente te manda a `opencode models` o a
// `/model`— y tipearlo. Su consumidor final es el catálogo, que es de donde
// `sf next` (y mañana `sf lanzar`) sacan con qué correr cada paso.
//
// LA SALIDA ES UNA TABLA PARA UN PIPE, NO UNA PANTALLA: `id`, y si el arnés da
// una descripción, un tab y la descripción. Command Code regala esa segunda
// columna —"long-horizon coding & knowledge work with 1M context"— y es
// exactamente el campo `para` que el catálogo ya tiene para que el ⑯ elija. sf la
// COPIA, no la interpreta (R3): es la opinión del proveedor sobre su modelo.
//
// No poder listar NO es un error del programa, y por eso sale con `SalidaParada`
// y no con `SalidaError`: no se rompió nada, hay que hacerlo a mano. Tipear el id
// tiene que seguir siendo un camino de primera clase — si sf queda inservible
// porque un arnés cambió una tabla, el diseño está mal.
func listarModelos(args []string) int {
	var harness string
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch {
		case a == "--harness" && i+1 < len(args):
			i++
			harness = args[i]
		case strings.HasPrefix(a, "--harness="):
			harness = strings.TrimPrefix(a, "--harness=")
		default:
			fmt.Fprintf(os.Stderr, "sf models: no conozco %q. Sólo --harness.\n", a)
			return SalidaError
		}
	}

	// Sin --harness, el de esta corrida. Es EnUso() y no `c.Harness` a propósito:
	// la pregunta acá es "¿qué ids me sirven DONDE ESTOY?", que es un hecho del
	// proceso, no lo último que alguien instaló.
	if harness == "" {
		raiz, err := os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "sf:", err)
			return SalidaError
		}
		g, _ := global.LeerPara(raiz)
		harness = g.EnUso()
	}
	if harness == "desconocido" || harness == "" {
		fmt.Fprintln(os.Stderr, "sf models: no sé en qué harness estás.")
		fmt.Fprintln(os.Stderr, "    sf models --harness="+strings.Join(global.Harness, "|"))
		return SalidaParada
	}

	ms, err := modelos.Listar(harness)
	if err != nil {
		// Un arnés que sf no conoce es un error de QUIEN LLAMÓ, y se contesta
		// distinto: ofrecerle "buscá el id a mano" a un typo no lo ayuda en nada.
		if errors.Is(err, modelos.ErrArnesDesconocido) {
			fmt.Fprintf(os.Stderr, "sf models: no conozco el arnés %q.\n", harness)
			fmt.Fprintln(os.Stderr, "    Son: "+strings.Join(global.Harness, " · "))
			return SalidaError
		}

		if errors.Is(err, modelos.ErrSinListado) {
			fmt.Fprintf(os.Stderr, "%s no lista sus modelos.\n", harness)
		} else {
			fmt.Fprintln(os.Stderr, "sf models:", err)
		}
		// Y acá el mismo cierre para los dos que quedan —el que no lista y el que
		// falló—, porque es el punto de todo el comando: que el listado no ande no
		// te frena, sólo te cuesta un copiar y pegar.
		fmt.Fprintln(os.Stderr)
		fmt.Fprintf(os.Stderr, "    Buscá el id a mano y declaralo igual:\n")
		fmt.Fprintf(os.Stderr, "    sf model <perfil> --alias <corto> --id <id> --via subagente\n")
		return SalidaParada
	}

	for _, m := range ms {
		if m.Para == "" {
			fmt.Println(m.ID)
			continue
		}
		fmt.Printf("%s\t%s\n", m.ID, m.Para)
	}
	return SalidaTrabajo
}
