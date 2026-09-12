package cli

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/maquina"
	"github.com/jdbaigorria/specforge/sf/internal/registro"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
	"github.com/jdbaigorria/specforge/sf/internal/sobre"
)

func next() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	e, r, err := cargar(raiz)
	if err != nil {
		// "No hay estado" no es un error: es un proyecto sin arrancar, y la
		// respuesta útil es decir cómo se arranca. Cualquier otro error sí es
		// un problema y se muestra tal cual.
		if errors.Is(err, estado.ErrNoHay) {
			fmt.Println("Este proyecto todavía no tiene estado.")
			fmt.Println()
			fmt.Println("  sf init")
			return SalidaParada
		}
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	// El mapa de modelos puede no existir —nadie corrió `sf install`— y eso NO
	// impide trabajar: `via()` cae a subagente. Lo único que no se puede
	// resolver sin él es `consola`.
	g, _ := global.LeerPara(raiz)

	i := maquina.Siguiente(raiz, e, r, g)

	// Lo que sf CONTESTÓ es la mitad que el registro no puede deducir del
	// estado.json: en los cinco pasos de producto no hay feature, y "brief" o
	// "prd" no son un campo de ningún archivo — los deduce la máquina.
	registro.AnotarPaso(registro.Paso{
		Tipo: i.Tipo.String(), Estado: i.Estado, Feature: i.Feature,
		Skill: i.Skill, Perfil: i.Perfil, Modelo: i.Modelo,
		Via: i.Via, Esfuerzo: i.Esfuerzo,
	})

	fmt.Print(mostrar(i))

	switch i.Tipo {
	case maquina.Trabajar:
		return SalidaTrabajo
	case maquina.Fin:
		return SalidaFin
	default:
		return SalidaParada
	}
}

// contexto es `sf context`: el sobre del estado actual.
//
// El único flag es `--completo`, y no es una comodidad: es el caso de H1b.
// Cuando el que trabaja es un modelo por consola sin shell propia, no puede
// abrir archivos — así que el orquestador corre esto y le pega la salida en el
// prompt. Mismo comando, mismo sobre; lo único que cambia es quién lo tipea.
//
// No hay ningún otro argumento a propósito: el estado, la feature y el lote son
// todos deducibles del estado.json, y un argumento deducible es un argumento
// que se pasa mal (H2).
func contexto(args []string) int {
	completo := false
	for _, a := range args {
		switch a {
		case "--completo", "--full":
			completo = true
		default:
			fmt.Fprintf(os.Stderr, "sf context: no conozco %q.\n", a)
			fmt.Fprintln(os.Stderr, "    El estado, la feature y el lote salen del estado.json.")
			return SalidaError
		}
	}

	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	e, r, err := cargar(raiz)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	// El catálogo entra al sobre porque el ⑫ tiene que ver el menú de modelos.
	// Que no esté NO es un error: el sobre lo dice con `Falta` y se sigue.
	gCat, _ := global.LeerPara(raiz)

	s, err := sobre.Armar(raiz, e, r, gCat)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	fmt.Print(s.Texto(raiz, completo))
	return SalidaTrabajo
}

// terminar es `sf done`: corre las compuertas y mueve lo que corresponda.
//
// Lo llama EL QUE TRABAJA, antes de morir — no el orquestador. Si `sf done` da
// ✗, el subagente todavía tiene el contexto para arreglarlo ahí mismo; si lo
// corriera el orquestador, cada ✗ costaría un subagente nuevo desde cero (H6).
//
// El único flag es `--msg`, y sólo aplica a `implementar`: sin él no se cierra
// el lote, porque cerrar el lote ES commitear.
func terminar(args []string) int {
	var msg string
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--msg" && i+1 < len(args):
			i++
			msg = args[i]
		case strings.HasPrefix(args[i], "--msg="):
			msg = strings.TrimPrefix(args[i], "--msg=")
		default:
			fmt.Fprintf(os.Stderr, "sf done: no conozco %q. Sólo --msg \"…\"\n", args[i])
			return SalidaError
		}
	}

	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	e, r, err := cargar(raiz)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	c := maquina.Terminar(raiz, e, r, msg)

	// Las fallas y los avisos son el dato más valioso del registro para la etapa
	// de tramos: son, literalmente, la lista de las veces que un modelo produjo
	// algo que la compuerta no aceptó. Hoy se imprimen y se pierden.
	registro.AnotarVeredicto(c.Estado, c.Fallas, c.Avisos, c.Movio)

	// El estado se guarda SÓLO si algo se movió. Un `sf done` que falla no
	// tiene que dejar rastro en el archivo... salvo el contador de intentos,
	// que es justamente el que cuenta los fracasos: por eso también se guarda
	// cuando falla dentro de una feature.
	if c.Movio || c.Cambio {
		if err := e.Guardar(raiz); err != nil {
			fmt.Fprintln(os.Stderr, "sf: no pude guardar el estado:", err)
			return SalidaError
		}
	}

	// ────────────────────────────────────────────────────────────────────
	// DOS LECTORES, Y LA DIFERENCIA ES MECÁNICA
	// ────────────────────────────────────────────────────────────────────
	//
	//	pasó y NO movió   es una PARADA        la lee Javier    → Acta()
	//	pasó y movió      la máquina sigue     la lee el orq.   → Texto()
	//	no pasó           hay que arreglar     la lee el subag. → Texto()
	//
	// No hace falta preguntar en qué paso estamos: un `done` que pasa y deja
	// el estado quieto es, por definición, la máquina esperando una decisión
	// que no puede tomar. Ahí es donde el acta vale, y sólo ahí.
	if c.Pasa() && !c.Movio {
		fmt.Print(c.Acta())
	} else {
		fmt.Print(c.Texto())
	}
	if c.Mensaje != "" {
		fmt.Println("→ " + c.Mensaje)
	}

	if c.Pasa() {
		return SalidaTrabajo
	}
	return SalidaParada
}

// cargar lee el estado y el roadmap, que es lo que necesitan los dos comandos.
//
// El roadmap puede no existir todavía —los cinco estados de producto corren
// antes que él— y eso NO es un error: se devuelve nil y quien lo use sabe que
// hasta el ⑩ no hay cola. El estado, en cambio, sí hace falta siempre.
func cargar(raiz string) (*estado.Estado, *roadmap.Roadmap, error) {
	e, err := estado.Leer(raiz)
	if err != nil {
		return nil, nil, err
	}

	r, err := roadmap.Leer(raiz)
	if err != nil {
		if !errors.Is(err, roadmap.ErrNoHay) {
			return nil, nil, err
		}
		r = nil
	}
	return e, r, nil
}
