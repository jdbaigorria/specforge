package registro

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Filtro es qué se quiere ver de la película.
//
// El cero vale: sin filtros, `Leer` devuelve todo y `Texto` muestra las últimas
// Ultimas entradas.
type Filtro struct {
	Feature string
	Cmd     string
	Ultimas int
}

// Leer carga el registro entero, en orden de escritura.
//
// ────────────────────────────────────────────────────────────────────────────
// UNA LÍNEA ROTA NO ROMPE LA LECTURA
// ────────────────────────────────────────────────────────────────────────────
//
// Se saltea y se sigue. El archivo lo escriben procesos distintos que pueden
// morirse en el medio, y negarse a leer las 400 líneas buenas por una mala
// sería tirar justo la evidencia que se vino a buscar.
//
// Que no exista el archivo tampoco es un error: es un proyecto donde `sf`
// todavía no corrió.
func Leer(raiz string) ([]Entrada, error) {
	f, err := os.Open(filepath.Join(raiz, Archivo))
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var todas []Entrada
	s := bufio.NewScanner(f)
	// Una línea nunca pasa de Tope, pero el buffer por default de Scanner es de
	// 64 KB y no hay motivo para dejarlo más chico que eso.
	s.Buffer(make([]byte, 0, Tope), 64*1024)
	for s.Scan() {
		var e Entrada
		if err := json.Unmarshal(s.Bytes(), &e); err != nil {
			continue
		}
		todas = append(todas, e)
	}
	return todas, s.Err()
}

// Filtrar aplica el filtro y recorta a las últimas.
func Filtrar(todas []Entrada, f Filtro) []Entrada {
	var r []Entrada
	for _, e := range todas {
		if f.Feature != "" && e.DeQuienEs() != f.Feature {
			continue
		}
		if f.Cmd != "" && e.Cmd != f.Cmd {
			continue
		}
		r = append(r, e)
	}
	if f.Ultimas > 0 && len(r) > f.Ultimas {
		r = r[len(r)-f.Ultimas:]
	}
	return r
}

// Texto arma la vista legible: una línea por entrada.
//
//	13:22:41  done    f-2  implementar → revision  ✓  812ms  agente
//	13:31:07  next    f-2  revision                ⏵  sf-check · opus · subagente
//	13:58:02  done    f-2  revision → implementar  ✗  hay 2 hallazgos abiertos: h-1 · h-3
func Texto(es []Entrada) string {
	if len(es) == 0 {
		return "el registro está vacío.\n"
	}
	var b strings.Builder
	for _, e := range es {
		fmt.Fprintf(&b, "%s  %-9s %s %s %s\n",
			hora(e.T), e.Cmd, transicion(e), marca(e), cola(e))
	}
	return b.String()
}

// hora se queda con HH:MM:SS del timestamp ISO. La fecha entera es ruido cuando
// se están mirando veinte líneas de la misma tarde; el `--json` la tiene igual.
func hora(t string) string {
	if len(t) >= 19 {
		return t[11:19]
	}
	return t
}

// transicion dice de dónde a dónde se movió, o dónde se quedó.
func transicion(e Entrada) string {
	f := e.DeQuienEs()
	desde, hasta := e.Antes.Estado, e.Despues.Estado

	// El estado que la máquina nombró le gana al de la foto como PUNTO DE
	// PARTIDA: es el mismo dato cuando hay feature, y es el único que existe en
	// los cinco pasos de producto.
	if desde == "" && e.Estado != "" {
		desde = e.Estado
	}

	var donde string
	switch {
	case desde == "" && hasta == "":
		// Sin estado de feature —los cinco pasos de producto— lo que se muestra
		// es el paso que sf nombró, y si tampoco hay, el sello que se movió.
		donde = e.Estado
		if donde == "" {
			donde = sellosQueSeMovieron(e.Antes.Producto, e.Despues.Producto)
		}
		if donde == "" {
			donde = "—"
		}
	case desde == hasta || hasta == "":
		donde = desde
	case desde == "":
		donde = hasta
	default:
		donde = desde + " → " + hasta
	}
	if f == "" {
		return fmt.Sprintf("%-28s", donde)
	}
	return fmt.Sprintf("%-4s %-23s", f, donde)
}

// marca traduce el exit code a un símbolo, con el mismo vocabulario que el resto
// de sf: ✓ pasó · ✗ no avanzó · ⏵ hay trabajo · ⚠ parada · ● fin.
func marca(e Entrada) string {
	switch {
	case len(e.Fallas) > 0:
		return "✗"
	case e.Salida == 1:
		return "✗"
	case e.Salida == 2:
		return "⚠"
	case e.Salida == 3:
		return "●"
	case e.Movio:
		return "✓"
	default:
		return "⏵"
	}
}

// cola es lo más informativo que tenga esa línea, y no todo lo que tiene.
//
// El orden es por valor: si algo falló, eso es lo que hay que ver; si sf indicó
// un paso, eso; y si no, cuánto tardó.
func cola(e Entrada) string {
	if len(e.Fallas) > 0 {
		return strings.Join(e.Fallas, " · ")
	}
	if e.Skill != "" {
		partes := []string{e.Skill}
		if e.Modelo != "" {
			partes = append(partes, e.Modelo)
		}
		if e.Via != "" {
			partes = append(partes, e.Via)
		}
		if e.Lanzamiento != "" {
			partes = append(partes, "lanzó "+e.Lanzamiento)
		}
		return strings.Join(partes, " · ")
	}
	return fmt.Sprintf("%dms  %s", e.Ms, e.Quien)
}

// ────────────────────────────────────────────────────────────────────────────
// VUELTAS — la pregunta ① de cada tramo
// ────────────────────────────────────────────────────────────────────────────

// Vuelta es cuántas veces una feature entró a un estado, y desde dónde.
type Vuelta struct {
	Feature  string
	Estado   string
	Entradas int

	// Desde dice de qué estado vino cada entrada, y es la mitad que importa:
	// entrar a `implementar` desde `planificacion` es el camino normal; entrar
	// desde `revision` es el bucle.
	Desde map[string]int
}

// Vueltas cuenta las ENTRADAS a cada estado, por feature.
//
// ────────────────────────────────────────────────────────────────────────────
// ESTE CONTEO ES EL PUNTO DEL PAQUETE
// ────────────────────────────────────────────────────────────────────────────
//
// Hoy "¿cuántas vueltas dio la revisión?" lo contesta el campo `vuelta` de
// revision.json — que escribe el PROPIO sf-check y que sf nunca comprueba
// (verificado el 2026-09-03: nada en Go lo incrementa). Con esto lo cuenta sf.
//
//	Es la regla del proyecto —"sf no le cree al que trabajó"— aplicada al
//	único lugar donde todavía le cree.
//
// Y NO decide nada con el número. Contar no es cortar: en qué vuelta hay que
// frenar sale de mirar las corridas de T5 y T6, y cablearlo antes de tener el
// dato es inventarlo (por-tramos.md §10).
func Vueltas(es []Entrada) []Vuelta {
	type clave struct{ feature, estado string }
	acc := map[clave]*Vuelta{}

	sumar := func(feature, estado, desde string) {
		k := clave{feature, estado}
		v, hay := acc[k]
		if !hay {
			v = &Vuelta{Feature: feature, Estado: estado, Desde: map[string]int{}}
			acc[k] = v
		}
		v.Entradas++
		if desde == "" {
			desde = "—"
		}
		v.Desde[desde]++
	}

	for _, e := range es {
		// ① Los cinco pasos de PRODUCTO no tienen estado de feature: lo que se
		//    mueve ahí es un SELLO. Se cuenta el sello que cambió entre el antes
		//    y el después.
		//
		//    Y se COMPARA en vez de DERIVAR: la alternativa —"si brief_sellado
		//    está vacío estás en el brief"— sería una segunda copia del switch
		//    de `terminarProducto`, que se desincronizaría el día que alguien
		//    mueva un paso. Comparando, esto reporta lo que de verdad pasó,
		//    aunque el orden de la máquina cambie.
		if s := sellosQueSeMovieron(e.Antes.Producto, e.Despues.Producto); s != "" {
			sumar("", s, "")
		}

		// ② Y las transiciones de FEATURE, que son las de los cuatro estados
		//    que se repiten. Sólo cuenta cuando el estado CAMBIÓ: un `sf next`
		//    repetido no es una vuelta, es haber mirado dos veces.
		if e.Despues.Estado == "" || e.Despues.Estado == e.Antes.Estado {
			continue
		}
		sumar(e.DeQuienEs(), e.Despues.Estado, e.Antes.Estado)
	}

	r := make([]Vuelta, 0, len(acc))
	for _, v := range acc {
		r = append(r, *v)
	}
	sort.Slice(r, func(i, j int) bool {
		if r[i].Feature != r[j].Feature {
			return r[i].Feature < r[j].Feature
		}
		return r[i].Estado < r[j].Estado
	})
	return r
}

// TextoVueltas arma la vista del conteo.
func TextoVueltas(vs []Vuelta) string {
	if len(vs) == 0 {
		return "todavía no hay ninguna transición de estado registrada.\n"
	}
	var b strings.Builder
	feature := ""
	for _, v := range vs {
		nombre := v.Feature
		if nombre == "" {
			nombre = "(producto)"
		}
		if nombre != feature {
			fmt.Fprintf(&b, "%s\n", nombre)
			feature = nombre
		}
		fmt.Fprintf(&b, "  %-16s %d%s\n", v.Estado, v.Entradas, desdeDonde(v))
	}
	return b.String()
}

// desdeDonde arma el "← 5 desde revision", y sólo cuando hay más de un origen.
//
// Con un solo origen el dato es obvio y ocuparía una columna para no decir nada.
// Con dos o más, es exactamente lo que separa el camino normal del bucle.
func desdeDonde(v Vuelta) string {
	if len(v.Desde) < 2 {
		return ""
	}
	orígenes := make([]string, 0, len(v.Desde))
	for d := range v.Desde {
		orígenes = append(orígenes, d)
	}
	sort.Strings(orígenes)

	partes := make([]string, 0, len(orígenes))
	for _, d := range orígenes {
		partes = append(partes, fmt.Sprintf("%d desde %s", v.Desde[d], d))
	}
	return "      ← " + strings.Join(partes, " · ")
}

// sellosQueSeMovieron nombra los sellos de producto que cambiaron, o "" si
// ninguno.
//
// Es lo que le permite al conteo ver los tramos T1–T3, que son justamente los
// primeros que se van a correr y donde NO hay estado de feature que mirar.
//
// Los nombres son los de los cuatro campos y nada más. NO se traduce a "en qué
// paso de producto estás": eso lo decide la máquina, y copiar su switch acá
// sería tener el orden del flujo escrito en dos lados.
func sellosQueSeMovieron(a, d Producto) string {
	var movidos []string
	if a.Brief != d.Brief {
		movidos = append(movidos, "brief")
	}
	if a.Prd != d.Prd {
		movidos = append(movidos, "prd")
	}
	if a.Constitucion != d.Constitucion {
		movidos = append(movidos, "constitucion")
	}
	if a.Backlog != d.Backlog {
		movidos = append(movidos, "backlog")
	}
	return strings.Join(movidos, "+")
}
