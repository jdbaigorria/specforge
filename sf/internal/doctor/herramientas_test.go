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

// ────────────────────────────────────────────────────────────────────────────
// La cadena de composición — apareció con el replanteo de T1
// ────────────────────────────────────────────────────────────────────────────

// skillEn deja un SKILL.md mínimo en una carpeta con ese nombre.
func skillEn(t *testing.T, raiz, nombre, cuerpo string) {
	t.Helper()
	d := filepath.Join(raiz, ".claude", "skills", nombre)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "SKILL.md"),
		[]byte("---\nname: "+nombre+"\n---\n"+cuerpo), 0o644); err != nil {
		t.Fatal(err)
	}
}

// Un compositor instalado que llama a un primitivo que no está NO falla con un
// error: el modelo improvisa. La corrida sale igual y nadie se entera de que
// corrió sin el método. Por eso lo tiene que ver el doctor y no la corrida.
func TestElDoctorVeElPrimitivoQueFalta(t *testing.T) {
	raiz := t.TempDir()
	skillEn(t, raiz, "sfp-scout", `Call the Skill tool with "sfx-grilling"`)

	donde := map[string]string{"sfp-scout": filepath.Join(raiz, ".claude", "skills", "sfp-scout")}
	faltan := compuestasQueFaltan(donde, donde["sfp-scout"])
	if len(faltan) != 1 || faltan[0] != "sfx-grilling" {
		t.Errorf("quería [sfx-grilling], vino %v", faltan)
	}
}

// LA CADENA ES TRANSITIVA. scout llama a grilling y grilling llama a buscar:
// mirar un solo nivel dejaría el agujero un escalón más abajo, que es peor que
// no mirar, porque el informe diría que está todo bien.
func TestLaCadenaDeComposicionSeCaminaEntera(t *testing.T) {
	raiz := t.TempDir()
	skillEn(t, raiz, "sfp-scout", `Call the Skill tool with "sfx-grilling"`)
	skillEn(t, raiz, "sfx-grilling", `y si falta un hecho, Call the Skill tool with "sfx-buscar"`)

	base := filepath.Join(raiz, ".claude", "skills")
	donde := map[string]string{
		"sfp-scout":    filepath.Join(base, "sfp-scout"),
		"sfx-grilling": filepath.Join(base, "sfx-grilling"),
	}
	faltan := compuestasQueFaltan(donde, donde["sfp-scout"])
	if len(faltan) != 1 || faltan[0] != "sfx-buscar" {
		t.Errorf("tenía que encontrar sfx-buscar un nivel más abajo, vino %v", faltan)
	}
}

// Con todo instalado no dice nada. Un doctor que se queja cuando está todo bien
// se aprende a ignorar.
func TestConLaCadenaCompletaNoSeQuejaDeNada(t *testing.T) {
	raiz := t.TempDir()
	skillEn(t, raiz, "sfp-scout", `Call the Skill tool with "sfx-grilling"`)
	skillEn(t, raiz, "sfx-grilling", `Call the Skill tool with "sfx-buscar"`)
	skillEn(t, raiz, "sfx-buscar", "busca y ya")

	base := filepath.Join(raiz, ".claude", "skills")
	donde := map[string]string{
		"sfp-scout":    filepath.Join(base, "sfp-scout"),
		"sfx-grilling": filepath.Join(base, "sfx-grilling"),
		"sfx-buscar":   filepath.Join(base, "sfx-buscar"),
	}
	if faltan := compuestasQueFaltan(donde, donde["sfp-scout"]); len(faltan) != 0 {
		t.Errorf("está todo y se queja igual: %v", faltan)
	}
}
