package doctor

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
)

// ────────────────────────────────────────────────────────────────────────────
// LAS HERRAMIENTAS DEL QUE INVESTIGA — Y EL LÍMITE DE LO QUE SF PUEDE VER
// ────────────────────────────────────────────────────────────────────────────
//
// MEDIDO EL 2026-09-05. Primera corrida real del ⑥ con dos modelos y la misma
// idea semilla: uno citó 13 links y el otro CERO. La causa no era el modelo —
// SPECFORGE NUNCA LE DIO HERRAMIENTAS DE BÚSQUEDA A NINGUNO DE LOS DOS, y nadie
// se enteró hasta leer los dos briefs uno al lado del otro.
//
// Este bloque existe para que eso se vea ANTES, no después.
//
// ────────────────────────────────────────────────────────────────────────────
// Y LO QUE NO PUEDE, DICHO DE FRENTE
// ────────────────────────────────────────────────────────────────────────────
//
// `sf` es un binario que se ejecuta y termina. Los servidores MCP viven adentro
// del proceso del arnés, y desde acá NO SE VEN. Lo único que sf puede mirar es
// lo que está en el disco y en el entorno:
//
//	el binario de curl        está o no está        → nivel 0, sin llave
//	el .mcp.json              declara servidores    → la INTENCIÓN
//	la variable de entorno    está puesta o no      → si la llave llegó
//
// Un `.mcp.json` con tavily declarado y su llave puesta NO PRUEBA que el arnés
// lo haya levantado. Por eso el bloque termina con una línea que dice
// exactamente eso, y no con un ✓.
//
//	"no lo puedo saber acá" ≠ "no están"
//
// Es la misma regla de `lanzar/resumen.go` —lo que no se puede ver no se
// afirma— y la misma tercera salida que `compuerta.Resultado`.
//
// AVISA, NO FRENA. Que falte una llave no impide usar sf: impide investigar
// bien, y eso lo decide el que está por correr el ⑥.

// Herramientas es lo que el que investiga tiene a mano, hasta donde se ve.
type Herramientas struct {
	// Curl es la ruta del binario, vacía si no está.
	//
	// Es el único de todo el bloque que alcanza SOLO: con curl se llega a los
	// registries (npm, PyPI, crates.io) y a la API pública de GitHub, que son
	// el nivel 0 y no necesitan ninguna llave.
	Curl string

	// MCP es la ruta del .mcp.json que se encontró, vacía si no hay ninguno.
	MCP string

	// Servidores son los que ese archivo declara, ordenados por nombre.
	Servidores []Servidor
}

// Servidor es un MCP declarado y si su llave está puesta.
type Servidor struct {
	Nombre string

	// Llave es el nombre de la variable de entorno que necesita, vacía si no
	// necesita ninguna. Nunca se guarda el valor: acá no viaja ningún secreto.
	Llave string

	// Puesta es si esa variable tiene algo en el entorno de ESTA corrida.
	Puesta bool
}

// Nivel0 dice si se puede investigar sin ninguna llave.
//
// Es la pregunta que importa, porque es la que tiene respuesta en cualquier
// arnés: si esto es false, el ⑥ va a salir de la imaginación del modelo.
func (h Herramientas) Nivel0() bool { return h.Curl != "" }

// reLlave saca el nombre de la variable de un "${TAVILY_API_KEY}".
var reLlave = regexp.MustCompile(`^\$\{([A-Za-z_][A-Za-z0-9_]*)\}$`)

// revisarHerramientas mira el disco y el entorno. Nunca sale a la red.
func revisarHerramientas(raiz string) Herramientas {
	var h Herramientas
	h.Curl, _ = exec.LookPath("curl")

	h.MCP = buscarMCP(raiz)
	if h.MCP == "" {
		return h
	}

	b, err := os.ReadFile(h.MCP)
	if err != nil {
		return h
	}
	var doc struct {
		MCPServers map[string]struct {
			Env map[string]string `json:"env"`
		} `json:"mcpServers"`
	}
	if json.Unmarshal(b, &doc) != nil {
		return h
	}

	for nombre, s := range doc.MCPServers {
		sv := Servidor{Nombre: nombre}
		// Se toma la PRIMERA que sea una referencia `${VAR}`. Un servidor con
		// dos llaves es raro y no cambia el diagnóstico: lo que se contesta es
		// "¿le falta una llave?", no "¿cuáles?".
		for _, v := range ordenado(s.Env) {
			if m := reLlave.FindStringSubmatch(s.Env[v]); m != nil {
				sv.Llave = m[1]
				sv.Puesta = os.Getenv(m[1]) != ""
				break
			}
		}
		h.Servidores = append(h.Servidores, sv)
	}
	sort.Slice(h.Servidores, func(a, b int) bool {
		return h.Servidores[a].Nombre < h.Servidores[b].Nombre
	})
	return h
}

// ordenado devuelve las claves de un mapa en orden, para que dos corridas con
// el mismo archivo den el mismo informe.
func ordenado(m map[string]string) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// buscarMCP encuentra el .mcp.json que el arnés va a levantar.
//
// Se mira primero el del proyecto y después el del plugin, en ese orden, porque
// es el mismo orden en el que gana: lo que el proyecto declara pisa lo que trae
// el plugin. Se devuelve el PRIMERO que exista y no se mezclan — mezclarlos
// diría que hay servidores que en la corrida real no van a convivir.
func buscarMCP(raiz string) string {
	candidatos := []string{filepath.Join(raiz, ".mcp.json")}
	for _, patron := range raicesDeSkills(raiz) {
		// raicesDeSkills apunta a `…/skills/*`; el .mcp.json del plugin vive
		// dos niveles arriba, al lado de la carpeta skills.
		candidatos = append(candidatos, filepath.Join(filepath.Dir(filepath.Dir(patron)), ".mcp.json"))
	}
	for _, c := range candidatos {
		m, _ := filepath.Glob(c)
		sort.Strings(m)
		for _, ruta := range m {
			if _, err := os.Stat(ruta); err == nil {
				return ruta
			}
		}
	}
	return ""
}
