package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/andamio"
	"github.com/jdbaigorria/specforge/sf/internal/arranque"
	"github.com/jdbaigorria/specforge/sf/internal/auditoria"
	"github.com/jdbaigorria/specforge/sf/internal/doctor"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/registro"
	"github.com/jdbaigorria/specforge/sf/internal/vista"
)

// verRegistro es `sf log`: la película, para leerla.
//
// ────────────────────────────────────────────────────────────────────────────
// ES EL ÚNICO COMANDO NUEVO, Y EXISTE PORQUE UN JSONL A OJO NO SE LEE
// ────────────────────────────────────────────────────────────────────────────
//
//	sf log                  las últimas 20
//	sf log --ultimas N      cuántas
//	sf log --feature f-2    sólo las de esa feature
//	sf log --cmd done       sólo ese comando
//	sf log --vueltas        el conteo por estado — la pregunta ① de cada tramo
//	sf log --json           crudo, para jq
//
// SALE SIEMPRE CON 0, y no es un descuido: es un LECTOR. Un 2 significaría
// "parada, es de Javier" y confundiría al orquestador que lo corriera adentro
// del bucle; un 1 haría que un registro vacío pareciera un error.
func verRegistro(args []string) int {
	f := registro.Filtro{Ultimas: 20}
	vueltas, crudo := false, false

	for i := 0; i < len(args); i++ {
		a := args[i]
		valor := func(flag string) (string, bool) {
			if a == flag && i+1 < len(args) {
				i++
				return args[i], true
			}
			if v, ok := strings.CutPrefix(a, flag+"="); ok {
				return v, true
			}
			return "", false
		}
		switch {
		case a == "--vueltas":
			vueltas = true
		case a == "--json":
			crudo = true
		default:
			if v, ok := valor("--feature"); ok {
				f.Feature = v
			} else if v, ok := valor("--cmd"); ok {
				f.Cmd = v
			} else if v, ok := valor("--ultimas"); ok {
				n, err := strconv.Atoi(v)
				if err != nil || n < 1 {
					fmt.Fprintf(os.Stderr, "sf log: --ultimas va con un número, no %q\n", v)
					return SalidaError
				}
				f.Ultimas = n
			} else {
				fmt.Fprintf(os.Stderr, "sf log: no conozco %q.\n", a)
				fmt.Fprintln(os.Stderr, "    Son --ultimas, --feature, --cmd, --vueltas y --json.")
				return SalidaError
			}
		}
	}

	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}
	todas, err := registro.Leer(raiz)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf log:", err)
		return SalidaError
	}

	// El conteo mira TODO el registro y no las últimas veinte: contar vueltas
	// sobre una ventana daría un número más chico que el real, y ese número es
	// justo el que se vino a buscar.
	if vueltas {
		fmt.Print(registro.TextoVueltas(registro.Vueltas(
			registro.Filtrar(todas, registro.Filtro{Feature: f.Feature, Cmd: f.Cmd}))))
		return SalidaTrabajo
	}

	es := registro.Filtrar(todas, f)
	if crudo {
		for _, e := range es {
			b, err := json.Marshal(e)
			if err != nil {
				continue
			}
			fmt.Println(string(b))
		}
		return SalidaTrabajo
	}
	fmt.Print(registro.Texto(es))
	return SalidaTrabajo
}

// auditar es `sf audit`: el punta a punta sobre varias features.
//
// Es el único comando que NO es de la máquina y que igual sirve un sobre. No
// rompe H2 —`sf context` sigue sin argumentos— porque son preguntas distintas:
// `sf context` pregunta "¿qué necesito para el paso en el que estoy?", y eso es
// deducible. Acá el alcance NO es deducible: lo elige Javier.
//
//	sf audit                 todo lo que se construyó
//	sf audit f-1 f-2 f-3     estas tres — un "módulo" es un conjunto de features
//	sf audit --completo      embebe el material, para un modelo sin shell
func auditar(args []string) int {
	completo := false
	var ids []string
	for _, a := range args {
		switch {
		case a == "--completo" || a == "--full":
			completo = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf audit: no conozco %q. Sólo --completo.\n", a)
			return SalidaError
		default:
			ids = append(ids, a)
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

	alcance, err := auditoria.Alcance(e, r, ids)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	inf, err := auditoria.Auditar(raiz, e, r, alcance)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	fmt.Print(inf.Texto(raiz, completo))

	// Con fallas sale por parada y no por error: no se rompió nada: sf encontró
	// contradicciones y alguien tiene que mirarlas. Es la misma distinción que
	// hace `sf done`.
	if inf.Fallas() > 0 {
		return SalidaParada
	}
	return SalidaTrabajo
}

// revisar es `sf doctor`: ¿esta instalación sirve?
//
// Como `sf install`, no es de la máquina: no mira el estado ni lo mueve. Mira
// la MÁQUINA DE JAVIER, que es lo único que ningún test del repo puede ver.
//
// El código de salida sigue la misma convención que todo lo demás, y por eso
// devuelve `SalidaParada` y no `SalidaError` cuando encuentra algo: no se rompió
// nada, hace falta que alguien intervenga. Un `sf doctor && ...` en un script de
// instalación distingue los tres casos sin parsear texto.
func revisar() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	i := doctor.Revisar(raiz, Version)
	fmt.Print(i.Texto())
	if !i.Sano() {
		return SalidaParada
	}
	return SalidaTrabajo
}

// estadoActual es `sf status`: la única salida de sf pensada para un humano.
//
// No mueve nada ni comprueba nada — lee tres fuentes y arma una vista. Por eso
// no aparece en ningún trazado del bucle: el bucle no lo necesita nunca.
func estadoActual() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	e, r, err := cargar(raiz)
	if err != nil {
		if errors.Is(err, estado.ErrNoHay) {
			fmt.Println("Este proyecto todavía no tiene estado.")
			fmt.Println()
			fmt.Println("  sf init")
			return SalidaParada
		}
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	fmt.Print(vista.Estado(raiz, e, r))
	return SalidaTrabajo
}

// iniciar es `sf init`: el andamio, y lo único que se corre antes que nada.
//
// No es un comando de la máquina —no mira el estado ni lo mueve— y por eso no
// aparece en ningún trazado del bucle. Es lo que hace que el bucle PUEDA
// empezar: sin estado.json, `sf next` sólo sabe decir "corré sf init".
func iniciar() int {
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	r, err := arranque.Iniciar(raiz)
	if err != nil {
		// Ya iniciado NO es un error del usuario: es alguien que corrió el
		// comando dos veces. La respuesta útil es decirle por dónde seguir, no
		// retarlo — y por eso sale por stdout con código de parada.
		if errors.Is(err, arranque.ErrYaIniciado) {
			fmt.Println("Este proyecto ya está iniciado.")
			fmt.Println()
			fmt.Println("  sf next")
			return SalidaParada
		}
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	for _, c := range r.Creados {
		fmt.Println("+", c)
	}
	fmt.Println()

	// El test_cmd entra a la allowlist del harness ACÁ y no en `sf install`,
	// y el motivo es que `sf install` corre antes: en ese momento no hay
	// constitución de donde leerlo.
	//
	// Y desde que el modo de Command Code es `dont-ask`, esto dejó de ser un
	// lujo: lo que no está en la lista no pregunta, FALLA. Un `test_cmd` afuera
	// hace fallar el ⑲ de todos los lotes, siempre.
	if r.Stack.Reconocido() {
		if g, err := global.LeerPara(raiz); err == nil {
			if ruta, toco, err := andamio.PermitirComando(raiz, g.EnUso(), r.Stack.TestCmd); err == nil && toco {
				fmt.Printf("+ %s (permití `%s`)\n", ruta, strings.Fields(r.Stack.TestCmd)[0])
			}
		}
	}

	if r.Stack.Reconocido() {
		fmt.Printf("%s (%s) · test_cmd: %s\n", r.Stack.Lenguaje, r.Stack.Manifiesto, r.Stack.TestCmd)
	} else {
		// Se avisa fuerte porque sin test_cmd la constitución NO SELLA, y ese
		// freno aparecería recién en el ⑧ sin decir de dónde viene.
		fmt.Println("⚠ No reconocí el stack: completá `test_cmd:` en la constitución.")
		fmt.Println("  Sin eso el ⑧ no sella y no se puede correr ningún test.")
	}

	fmt.Println()
	fmt.Println("  sf next")
	return SalidaTrabajo
}
