package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func conMCP(t *testing.T, contenido string) string {
	t.Helper()
	raiz := t.TempDir()
	if err := os.WriteFile(filepath.Join(raiz, ".mcp.json"), []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
	return raiz
}

const mcpDeSpecForge = `{"mcpServers":{
  "context7": {"command":"npx"},
  "fetch":    {"command":"uvx"},
  "tavily":   {"command":"npx","env":{"TAVILY_API_KEY":"${TAVILY_API_KEY}"}},
  "github":   {"command":"npx","env":{"GITHUB_PERSONAL_ACCESS_TOKEN":"${GITHUB_PERSONAL_ACCESS_TOKEN}"}}
}}`

// Lo que el 2026-09-05 nadie vio hasta leer los dos briefs: un MCP declarado
// SIN su llave arranca y no puede buscar. Eso tiene que salir en el informe.
func TestElDoctorVeLaLlaveQueFalta(t *testing.T) {
	t.Setenv("TAVILY_API_KEY", "hay-una")
	t.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", "")

	h := revisarHerramientas(conMCP(t, mcpDeSpecForge))

	por := map[string]Servidor{}
	for _, s := range h.Servidores {
		por[s.Nombre] = s
	}
	if len(por) != 4 {
		t.Fatalf("quería los 4 servidores, vinieron %d: %+v", len(por), h.Servidores)
	}
	if por["context7"].Llave != "" {
		t.Errorf("context7 no necesita llave: %+v", por["context7"])
	}
	if !por["tavily"].Puesta {
		t.Errorf("la llave de tavily está puesta y no lo vio: %+v", por["tavily"])
	}
	if por["github"].Puesta {
		t.Errorf("la llave de github NO está y dijo que sí: %+v", por["github"])
	}
	if por["github"].Llave != "GITHUB_PERSONAL_ACCESS_TOKEN" {
		t.Errorf("tiene que nombrar la variable que falta: %+v", por["github"])
	}
}

// Nunca se guarda el VALOR de una llave: acá no viaja ningún secreto, y el
// informe se imprime en una terminal que alguien puede estar compartiendo.
func TestElInformeNoGuardaElValorDeLaLlave(t *testing.T) {
	t.Setenv("TAVILY_API_KEY", "tvly-secreto-de-verdad")

	h := revisarHerramientas(conMCP(t, mcpDeSpecForge))
	i := Informe{Herramientas: h}
	if strings.Contains(i.Texto(), "tvly-secreto-de-verdad") {
		t.Error("el valor de la llave llegó al informe")
	}
	for _, s := range h.Servidores {
		if strings.Contains(s.Llave, "tvly-") {
			t.Errorf("se guardó el valor y no el nombre: %+v", s)
		}
	}
}

// Sin .mcp.json no hay drama y no hay servidores. El nivel 0 no depende de eso.
func TestSinMCPNoHaySevidoresYNoEsUnaFalla(t *testing.T) {
	h := revisarHerramientas(t.TempDir())
	if h.MCP != "" || len(h.Servidores) != 0 {
		t.Errorf("no hay .mcp.json y encontró algo: %+v", h)
	}
}

// LA LÍNEA QUE NO SE PUEDE PERDER. Sin ella, cuatro ✓ se leen como "las
// herramientas andan", y lo comprobado es que están DECLARADAS. Levantarlas es
// del arnés, y el arnés no se ve desde un binario que se ejecuta y termina.
func TestElInformeDiceQueNoPuedeVerElArnes(t *testing.T) {
	i := Informe{Herramientas: revisarHerramientas(conMCP(t, mcpDeSpecForge))}
	if !strings.Contains(i.Texto(), "no se sabe desde acá") {
		t.Errorf("falta el límite declarado:\n%s", i.Texto())
	}
}

// Avisa, NO frena: que falte una llave no impide usar sf. Impide investigar
// bien, y con qué evidencia se sella el ⑥ lo decide Javier.
func TestLaLlaveQueFaltaAvisaYNoFrena(t *testing.T) {
	t.Setenv("TAVILY_API_KEY", "")
	t.Setenv("GITHUB_PERSONAL_ACCESS_TOKEN", "")

	i := Revisar(conMCP(t, mcpDeSpecForge), "test")
	if strings.Contains(strings.Join(i.Fallas, "\n"), "TAVILY") {
		t.Errorf("una llave que falta no es una falla: %v", i.Fallas)
	}
	if !strings.Contains(strings.Join(i.Avisos, "\n"), "TAVILY_API_KEY") {
		t.Errorf("tiene que avisar cuál falta: %v", i.Avisos)
	}
}
