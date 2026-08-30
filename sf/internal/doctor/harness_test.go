package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// conCatalogo escribe un catálogo en el proyecto y devuelve la raíz.
//
// Va en el PROYECTO y no en el home porque es lo que `LeerPara` mira primero, y
// así el test no toca el home de quien lo corre ni depende del orden en que
// corren los demás.
func conCatalogo(t *testing.T, c *global.Config) string {
	t.Helper()
	parandoEn(t, c.Harness)
	raiz := t.TempDir()
	if err := c.GuardarEn(global.RutaDeProyecto(raiz)); err != nil {
		t.Fatal(err)
	}
	return raiz
}

// El harness desconocido pasó de ⚠ a ✗, y no es un endurecimiento gratuito:
// desde que el catálogo está indexado por harness, sin saber cuál es NO HAY
// perfil que resolver y `sf next` no puede decir con qué se lanza nada.
func TestUnHarnessDesconocidoAhoraFrena(t *testing.T) {
	for _, v := range []string{"CLAUDECODE", "CLAUDE_CODE_ENTRYPOINT", "OPENCODE", "OPENCODE_BIN_PATH", "COMMANDCODE", "COMMAND_CODE_ENTRYPOINT"} {
		t.Setenv(v, "")
	}
	t.Setenv("SPECFORGE_HOME", t.TempDir())

	i := Revisar(t.TempDir(), "v0")
	if len(i.Fallas) == 0 {
		t.Fatal("un harness desconocido tiene que ser falla, no aviso")
	}
	if !strings.Contains(strings.Join(i.Fallas, " "), "--harness") {
		t.Errorf("la falla no dice cómo arreglarlo: %v", i.Fallas)
	}
}

// El harness sale del catálogo ESCRITO y no de la detección: es lo que va a
// usar `sf next`, y reportar otra cosa sería un diagnóstico que no describe al
// binario que corre.
func TestElHarnessSaleDelCatalogoYNoDeLaDeteccion(t *testing.T) {
	t.Setenv("CLAUDECODE", "1") // la detección diría claude-code
	t.Setenv("SPECFORGE_HOME", t.TempDir())

	raiz := conCatalogo(t, global.Semilla("opencode"))
	if i := Revisar(raiz, "v0"); i.Harness != "opencode" {
		t.Errorf("harness %q, quería el del catálogo", i.Harness)
	}
}

// Los perfiles que la máquina PIDE tienen que estar declarados. Enterarse acá
// es más barato que enterarse cuando el orquestador ya arrancó el bucle.
func TestUnPerfilSinDeclararEsFalla(t *testing.T) {
	t.Setenv("SPECFORGE_HOME", t.TempDir())
	parandoEn(t, "opencode")
	raiz := conCatalogo(t, global.Semilla("opencode"))

	i := Revisar(raiz, "v0")
	if len(i.Perfiles) != 2 {
		t.Fatalf("revisó %d perfiles, quería los 2 que pide la máquina", len(i.Perfiles))
	}
	f := strings.Join(i.Fallas, " ")
	for _, p := range []string{global.Razonar, global.Construir} {
		if !strings.Contains(f, p) {
			t.Errorf("no se quejó de que falta %q: %v", p, i.Fallas)
		}
	}
	// `mecanico` NO se exige: ningún estado lo devuelve.
	if strings.Contains(f, global.Mecanico) {
		t.Errorf("exigió mecanico, que ningún estado pide: %v", i.Fallas)
	}
}

// Un alias declarado sin su portamodelo es un `sf next` que va a nombrar un
// agente que el harness no conoce — y eso falla cuando el orquestador ya lanzó.
func TestUnAliasSinPortamodeloEsFalla(t *testing.T) {
	t.Setenv("SPECFORGE_HOME", t.TempDir())
	parandoEn(t, "opencode")
	c := global.Semilla("opencode")
	for _, p := range []string{global.Razonar, global.Construir} {
		if _, err := c.Declarar(p, global.Modelo{
			Alias: "a-" + p, ID: "prov/" + p, Via: global.Subagente,
		}); err != nil {
			t.Fatal(err)
		}
	}
	raiz := conCatalogo(t, c)

	i := Revisar(raiz, "v0")
	if !strings.Contains(strings.Join(i.Fallas, " "), "archivo de agente") {
		t.Fatalf("no detectó los portamodelo que faltan: %v", i.Fallas)
	}

	// Y con los archivos puestos, deja de quejarse.
	dir := filepath.Join(raiz, ".opencode", "agents")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{global.Razonar, global.Construir} {
		if err := os.WriteFile(filepath.Join(dir, "sf-a-"+p+".md"), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if i := Revisar(raiz, "v0"); strings.Contains(strings.Join(i.Fallas, " "), "archivo de agente") {
		t.Errorf("sigue quejándose con los archivos puestos: %v", i.Fallas)
	}
}

// En Claude Code no hay portamodelo que exigir: ahí el modelo va como parámetro
// de la llamada.
func TestEnClaudeCodeNoSeExigePortamodelo(t *testing.T) {
	t.Setenv("SPECFORGE_HOME", t.TempDir())
	parandoEn(t, "claude-code")
	c := global.Semilla("claude-code")
	for _, p := range []string{global.Razonar, global.Construir} {
		if _, err := c.Declarar(p, global.Modelo{Alias: "a-" + p, ID: p, Via: global.Subagente}); err != nil {
			t.Fatal(err)
		}
	}

	i := Revisar(conCatalogo(t, c), "v0")
	if strings.Contains(strings.Join(i.Fallas, " "), "archivo de agente") {
		t.Errorf("exigió portamodelo en claude-code: %v", i.Fallas)
	}
}

// Medido: Command Code no lee ~/.claude/skills/, pero sí `.agents/skills/`.
// Si el doctor no mira esa raíz, reporta como faltantes skills que están.
func TestElDoctorMiraLasRaicesDeLosTresHarness(t *testing.T) {
	raiz := t.TempDir()
	patrones := strings.Join(raicesDeSkills(raiz), " ")
	for _, r := range []string{".agents", ".opencode", ".commandcode", ".claude"} {
		if !strings.Contains(patrones, r) {
			t.Errorf("no mira %s: %s", r, patrones)
		}
	}
}

// parandoEn dice en qué harness está corriendo el test.
//
// Hace falta desde que el puntero de harness se DETECTA: esta suite corre
// adentro de algún arnés, así que sin limpiar las variables cada test heredaría
// el de quien lo lanzó. Que haga falta es la prueba de que el mecanismo anda.
func parandoEn(t *testing.T, harness string) {
	t.Helper()
	for _, v := range global.VarsDeHarness {
		t.Setenv(v, "")
	}
	if harness != "" {
		t.Setenv(global.VarHarness, harness)
	}
}
