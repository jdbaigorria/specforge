package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/lanzar"
	"github.com/jdbaigorria/specforge/sf/internal/maquina"
	"github.com/jdbaigorria/specforge/sf/internal/registro"
	"github.com/jdbaigorria/specforge/sf/internal/sobre"
)

// lanzarPaso es `sf lanzar`: sf corre el paso, en vez de pedir que lo corran.
//
// ────────────────────────────────────────────────────────────────────────────
// NO DECIDE NADA NUEVO
// ────────────────────────────────────────────────────────────────────────────
//
// Le pregunta a `sf next` qué toca, resuelve el modelo con LA MISMA CADENA DE
// PRECEDENCIA DE SIEMPRE, arma la línea desde el catálogo y ejecuta. Lo que
// cambia no es la decisión: es quién aprieta el botón (headless.md §3).
//
// Y por eso son dos comandos y no uno: `sf next` sigue siendo consulta pura, y
// éste escribe. Tampoco corre `sf done` solo — lanzar es hacer y `done` es
// juzgar, y juntarlos sería que el que hace se apruebe a sí mismo.
func lanzarPaso(args []string) int {
	var (
		harness, alias, esfuerzo string
		seco                     bool
		espera                   = 30 * time.Minute
	)
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
		case a == "--seco" || a == "--dry-run":
			seco = true
		default:
			if v, ok := valor("--harness"); ok {
				harness = v
			} else if v, ok := valor("--alias"); ok {
				alias = v
			} else if v, ok := valor("--esfuerzo"); ok {
				esfuerzo = v
			} else if v, ok := valor("--espera"); ok {
				d, err := time.ParseDuration(v)
				if err != nil {
					fmt.Fprintf(os.Stderr, "sf lanzar: no entiendo la espera %q. Va como `30m`, `90s`, o `0` para sin tope.\n", v)
					return SalidaError
				}
				espera = d
			} else {
				fmt.Fprintf(os.Stderr, "sf lanzar: no conozco %q.\n", a)
				fmt.Fprintln(os.Stderr, "    Son --harness, --alias, --esfuerzo, --espera y --seco.")
				return SalidaError
			}
		}
	}
	raiz, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	// PEDIR OTRO ARNÉS ES DECIRLE A TODO SF QUE ESTAMOS AHÍ, y por eso se hace
	// con la variable y no con un parámetro suelto: así el modelo se resuelve
	// contra el catálogo de ESE arnés —no del de acá— y de paso el hijo la
	// hereda, que es literalmente para lo que `VarHarness` se creó.
	if harness != "" {
		if err := os.Setenv(global.VarHarness, harness); err != nil {
			fmt.Fprintln(os.Stderr, "sf:", err)
			return SalidaError
		}
	}

	e, r, err := cargar(raiz)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}
	g, _ := global.LeerPara(raiz)

	i := maquina.Siguiente(raiz, e, r, g)
	if err := lanzar.Puede(i); err != nil {
		// Si la máquina no está en `Trabajar`, ELLA ya explica por qué y con qué
		// comando se sale. Mostrar eso —lo mismo que imprime `sf next`— es mucho
		// más útil que un "no hay trabajo" de nuestra cosecha.
		if errors.Is(err, lanzar.ErrNoHayTrabajo) {
			fmt.Print(mostrar(i))
			if i.Tipo == maquina.Fin {
				return SalidaFin
			}
			return SalidaParada
		}
		// Y los motivos propios de lanzar —conversa, va por consola, no nombra
		// skill— no los sabe nadie más, así que se dicen acá.
		fmt.Fprintln(os.Stderr, "sf lanzar:", err)
		return SalidaParada
	}

	modelo, esf := i.Modelo, i.Esfuerzo
	if alias != "" {
		// Un alias explícito gana, y tiene que existir EN ESTE ARNÉS. No se busca
		// uno parecido ni se traduce el id a otro proveedor: es R3, y es lo mismo
		// que ya hace `sf next`.
		m, hay := g.PorAlias(alias)
		if !hay {
			fmt.Fprintf(os.Stderr, "sf lanzar: %q no está declarado en %s.\n", alias, g.EnUso())
			if otros := g.AliasDe(g.EnUso()); len(otros) > 0 {
				var nombres []string
				for _, o := range otros {
					nombres = append(nombres, o.Alias)
				}
				fmt.Fprintln(os.Stderr, "    Hay: "+strings.Join(nombres, " · "))
			}
			return SalidaParada
		}
		modelo, esf = m.ID, m.Esfuerzo
	}
	if esfuerzo != "" {
		esf = esfuerzo
	}

	// FRENAR ANTES DE GASTAR UNA CORRIDA. Sin el permiso declarado, Command Code
	// arranca, tarda medio minuto, y devuelve `exit=0` con la carta diciendo
	// `success` y ningún artefacto — que es el peor desenlace posible, porque
	// parece que anduvo.
	yolo := g.PermisoDe(g.EnUso()) == global.PermisoYolo
	if lanzar.NecesitaPermiso(g.EnUso()) && !yolo {
		fmt.Fprintf(os.Stderr, "🛑 %s no escribe archivos ni corre comandos en headless "+
			"sin que vos se lo permitas.\n", g.EnUso())
		fmt.Fprintln(os.Stderr, "   Es --yolo, que apaga TODOS sus chequeos de permisos, y la")
		fmt.Fprintln(os.Stderr, "   decisión es tuya: sf no se la toma solo.")
		fmt.Fprintln(os.Stderr)
		// El archivo se nombra por su ORIGEN y no "~/.specforge/" a secas: hay dos
		// catálogos y el del proyecto gana si existe. Mandar a editar el que no
		// rige es peor que no decir nada — el permiso se escribe, no pasa nada, y
		// nadie entiende por qué.
		donde := "~/.specforge/" + global.Archivo
		if o := g.Origen(); o != "" {
			donde = filepath.Join(o, global.Archivo)
		}
		fmt.Fprintln(os.Stderr, "   Lo declarás con `sf install` —te lo pregunta— o a mano en")
		fmt.Fprintf(os.Stderr, "   %s:\n\n     permisos:\n       %s: yolo\n", donde, g.EnUso())
		return SalidaParada
	}

	s, err := sobre.Armar(raiz, e, r, g)
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf:", err)
		return SalidaError
	}

	linea, err := lanzar.Linea(lanzar.Pedido{
		Harness:  g.EnUso(),
		Modelo:   modelo,
		Esfuerzo: esf,
		Raiz:     raiz,
		Prompt:   lanzar.Armar(i, s.Texto(raiz, false)),
		// sf NO decide esto: lo lee de donde Javier lo escribió.
		Yolo: yolo,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "sf lanzar:", err)
		return SalidaParada
	}

	if seco {
		fmt.Println(lanzar.Mostrar(linea))
		return SalidaTrabajo
	}

	fmt.Printf("→ %s · %s%s\n", g.EnUso(), modelo, conEsfuerzo(esf))
	if i.Feature != "" {
		fmt.Printf("  %s · %s · feature %s\n", i.Estado, i.Skill, i.Feature)
	} else {
		fmt.Printf("  %s · %s\n", i.Estado, i.Skill)
	}

	f, err := lanzar.Correr(linea, lanzar.Ficha{
		Estado:   i.Estado,
		Feature:  i.Feature,
		Harness:  g.EnUso(),
		Alias:    alias,
		Modelo:   modelo,
		Esfuerzo: esf,
		Skill:    i.Skill,
	}, lanzar.Opciones{Raiz: raiz, Harness: g.EnUso(), Espera: espera})

	// Ata esta invocación a la ficha de la corrida que largó. Es lo que permite
	// contestar "entre este `next` que dijo subagente y el `done` que vino
	// después, ¿hubo una corrida?".
	registro.AnotarLanzamiento(f.ID)
	registro.AnotarPaso(registro.Paso{
		Tipo: i.Tipo.String(), Estado: i.Estado, Feature: i.Feature,
		Skill: i.Skill, Perfil: i.Perfil, Modelo: modelo,
		Via: i.Via, Esfuerzo: esf,
	})

	fmt.Print(lanzar.Contar(f))

	if err != nil {
		fmt.Fprintln(os.Stderr, "sf lanzar:", err)
		return SalidaError
	}
	// EL CÓDIGO DE SALIDA NO OPINA SOBRE EL TRABAJO, sólo sobre la corrida.
	//
	// Un 0 acá quiere decir "el proceso terminó bien", NO "el artefacto está
	// bien": Command Code devolvió exit=0 y `subtype: success` sin haber hecho
	// nada. Quien decide si el trabajo vale es `sf done`, y por eso `sf lanzar`
	// no lo corre solo.
	if f.Fin != "termino" || f.Salida != 0 {
		return SalidaParada
	}
	return SalidaTrabajo
}

// conEsfuerzo arma el sufijo del encabezado, o nada si no hay esfuerzo.
func conEsfuerzo(e string) string {
	if e == "" {
		return ""
	}
	return " · esfuerzo " + e
}
