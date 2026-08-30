package andamio

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// Cada test corre con su propio ~/.specforge/ de mentira. Sin esto, un test que
// pase escribiría en el home de verdad de quien lo corre — y `sf install` es
// justamente el comando que toca el home.
func enUnHomeDePrueba(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("SPECFORGE_HOME", filepath.Join(dir, ".specforge"))
	return dir
}

func TestInstalarPoneElOrquestadorConLosDosNombres(t *testing.T) {
	enUnHomeDePrueba(t)
	raiz := t.TempDir()

	r, err := Instalar(raiz, Opciones{Harness: "claude-code"})
	if err != nil {
		t.Fatal(err)
	}

	// Un archivo, dos destinos: AGENTS.md es el MISMO texto, no una traducción.
	claude, err := os.ReadFile(filepath.Join(raiz, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	agents, err := os.ReadFile(filepath.Join(raiz, "AGENTS.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(claude) != string(agents) {
		t.Error("CLAUDE.md y AGENTS.md no son idénticos")
	}
	if !strings.Contains(string(claude), "sf next") {
		t.Errorf("el orquestador no trae el bucle:\n%s", claude)
	}

	if r.Harness != "claude-code" {
		t.Errorf("harness %q", r.Harness)
	}
	// Y NO siembra ningún modelo, que es el invariante de H2: un default de un
	// proveedor copiado a un skill y de ahí a tareas.json es exactamente cómo
	// nació el "siempre opus".
	if r.Modelos != 0 {
		t.Errorf("sembró %d modelos y no tiene que sembrar ninguno", r.Modelos)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// H3, H4 y H8 — lo que sf escribe PARA EL HARNESS
// ────────────────────────────────────────────────────────────────────────────

// Medido: sin settings.json, Command Code arranca en baseline `default` y
// pregunta por todo. Y el modo tiene que ser `dont-ask` y no `auto-accept`,
// que sigue preguntando por comandos de shell arbitrarios.
func TestInstalarEscribeLosPermisosDelHarness(t *testing.T) {
	enUnHomeDePrueba(t)
	raiz := t.TempDir()
	if _, err := Instalar(raiz, Opciones{Harness: "commandcode"}); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(raiz, ".commandcode", "settings.json"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), `"dont-ask"`) {
		t.Errorf("no puso dont-ask:\n%s", b)
	}
	if strings.Contains(string(b), "auto-accept") {
		t.Errorf("puso auto-accept, que sigue preguntando por shell:\n%s", b)
	}
	// H4, medido: Command Code no lee ~/.claude/skills/. Sin esta línea los 18
	// skills son invisibles para él.
	if !strings.Contains(string(b), "~/.claude/skills") {
		t.Errorf("no le dijo dónde están los skills:\n%s", b)
	}
}

// La config de permisos de un proyecto tiene cosas de Javier igual que el
// CLAUDE.md: no se pisa sin pedirlo.
func TestNoPisaLaConfigDelHarnessQueYaExiste(t *testing.T) {
	enUnHomeDePrueba(t)
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, ".commandcode"), 0o755); err != nil {
		t.Fatal(err)
	}
	ruta := filepath.Join(raiz, ".commandcode", "settings.json")
	if err := os.WriteFile(ruta, []byte(`{"mio": true}`), 0o644); err != nil {
		t.Fatal(err)
	}

	r, err := Instalar(raiz, Opciones{Harness: "commandcode"})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(ruta)
	if string(b) != `{"mio": true}` {
		t.Errorf("pisó la config de Javier: %s", b)
	}
	if !strings.Contains(strings.Join(r.Salteados, " "), "settings.json") {
		t.Errorf("la salteó sin decirlo: %v", r.Salteados)
	}
}

// H8: el portamodelo existe porque en opencode y Command Code el modelo sale
// del archivo del agente y no se puede pisar al invocar. En Claude Code sí se
// puede, así que ahí no se genera ninguno.
func TestLosPortamodeloSoloSeGeneranDondeHacenFalta(t *testing.T) {
	for _, c := range []struct {
		harness string
		dir     string
		quiero  bool
	}{
		{"opencode", filepath.Join(".opencode", "agents"), true},
		{"commandcode", filepath.Join(".commandcode", "agents"), true},
		{"claude-code", filepath.Join(".claude", "agents"), false},
	} {
		t.Run(c.harness, func(t *testing.T) {
			enUnHomeDePrueba(t)
			g := global.Semilla(c.harness)
			if _, err := g.Declarar(global.Construir, global.Modelo{
				Alias: "barato", ID: "prov/barato", Via: global.Subagente,
			}); err != nil {
				t.Fatal(err)
			}
			if err := g.Guardar(); err != nil {
				t.Fatal(err)
			}

			raiz := t.TempDir()
			if _, err := Instalar(raiz, Opciones{Harness: c.harness}); err != nil {
				t.Fatal(err)
			}

			b, err := os.ReadFile(filepath.Join(raiz, c.dir, "sf-barato.md"))
			if c.quiero {
				if err != nil {
					t.Fatalf("no generó el portamodelo: %v", err)
				}
				if !strings.Contains(string(b), "model: prov/barato") {
					t.Errorf("no pineó el id:\n%s", b)
				}
			} else if err == nil {
				t.Error("generó un portamodelo donde el modelo va en la llamada")
			}
		})
	}
}

// Lo que hace que el portamodelo NO sea una cuarta copia de cada skill: no
// lleva método adentro. El día que alguien le meta "acordate de correr sf
// context", pasa a ser una fuente que hay que mantener sincronizada.
func TestElPortamodeloNoLlevaMetodoAdentro(t *testing.T) {
	texto := Portamodelo(global.Modelo{Alias: "x", ID: "prov/x", Via: global.Subagente})
	for _, prohibido := range []string{"sf context", "sf done", "sf next", "sf-build", "lote"} {
		if strings.Contains(texto, prohibido) {
			t.Errorf("el portamodelo menciona %q: dejó de ser sólo un modelo", prohibido)
		}
	}
	if !strings.Contains(texto, "model: prov/x") {
		t.Errorf("no pineó el modelo:\n%s", texto)
	}
}

// Un `via: consola` lo ejecuta el orquestador con sus manos: el harness nunca
// lo lanza, así que un portamodelo suyo sería un archivo que nadie invoca.
func TestUnModeloDeConsolaNoLlevaPortamodelo(t *testing.T) {
	enUnHomeDePrueba(t)
	g := global.Semilla("opencode")
	if _, err := g.Declarar(global.Razonar, global.Modelo{
		Alias: "ajeno", ID: "ajeno-1", Via: global.Consola, Comando: "ajeno exec",
	}); err != nil {
		t.Fatal(err)
	}
	if err := g.Guardar(); err != nil {
		t.Fatal(err)
	}

	raiz := t.TempDir()
	if _, err := Instalar(raiz, Opciones{Harness: "opencode"}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(raiz, ".opencode", "agents", "sf-ajeno.md")); err == nil {
		t.Error("generó un portamodelo para un modelo que sale por consola")
	}
	if f := PortamodelosQueFaltan(raiz, g); len(f) != 0 {
		t.Errorf("lo reportó como faltante: %v", f)
	}
}

// `dont-ask` deniega lo que no está en la lista, así que un test_cmd afuera
// hace fallar el ⑲ de TODOS los lotes. Por eso `sf init` la completa.
func TestPermitirComandoAgregaElTestCmd(t *testing.T) {
	enUnHomeDePrueba(t)
	raiz := t.TempDir()
	if _, err := Instalar(raiz, Opciones{Harness: "commandcode"}); err != nil {
		t.Fatal(err)
	}

	ruta, toco, err := PermitirComando(raiz, "commandcode", "npm test")
	if err != nil || !toco {
		t.Fatalf("no agregó el comando: toco=%v err=%v", toco, err)
	}
	b, _ := os.ReadFile(filepath.Join(raiz, ruta))
	if !strings.Contains(string(b), `"Shell(npm:*)"`) {
		t.Errorf("no quedó la regla:\n%s", b)
	}
	// Y lo que ya está no se duplica: `sf init` corre más de una vez.
	if _, toco, _ := PermitirComando(raiz, "commandcode", "npm test"); toco {
		t.Error("duplicó una regla que ya estaba")
	}
}

// El caso que más importa del comando: un CLAUDE.md de un proyecto suele tener
// instrucciones propias de Javier. Pisarlo sin avisar sería borrarle trabajo
// para poner cuatro líneas.
func TestNoPisaUnClaudeMdQueYaExiste(t *testing.T) {
	enUnHomeDePrueba(t)
	raiz := t.TempDir()

	mio := "# lo mío\n\nno me lo toques\n"
	if err := os.WriteFile(filepath.Join(raiz, "CLAUDE.md"), []byte(mio), 0o644); err != nil {
		t.Fatal(err)
	}

	r, err := Instalar(raiz, Opciones{})
	if err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(raiz, "CLAUDE.md"))
	if err != nil {
		t.Fatal(err)
	}
	if string(b) != mio {
		t.Errorf("pisó el CLAUDE.md existente:\n%s", b)
	}
	if !strings.Contains(strings.Join(r.Salteados, " "), "CLAUDE.md") {
		t.Errorf("no avisó que lo salteó: %v", r.Salteados)
	}
	// Y AGENTS.md, que no existía, sí se escribió: las dos mitades son
	// independientes.
	if _, err := os.Stat(filepath.Join(raiz, "AGENTS.md")); err != nil {
		t.Error("no escribió AGENTS.md, que no existía")
	}
}

func TestForzarSiPisa(t *testing.T) {
	enUnHomeDePrueba(t)
	raiz := t.TempDir()
	if err := os.WriteFile(filepath.Join(raiz, "CLAUDE.md"), []byte("viejo\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := Instalar(raiz, Opciones{Forzar: true}); err != nil {
		t.Fatal(err)
	}

	b, _ := os.ReadFile(filepath.Join(raiz, "CLAUDE.md"))
	if string(b) == "viejo\n" {
		t.Error("--forzar no pisó")
	}
}

// Y el que protege lo más valioso: ~/.specforge/ tiene los modelos que Javier
// fue aprobando de a uno. Re-sembrarlo en cada `sf install` los tiraría.
func TestNoResiembraElMapaDeModelos(t *testing.T) {
	enUnHomeDePrueba(t)

	c := global.Semilla("claude-code")
	c.DeclararSuelto("deepseek", global.Modelo{ID: "ds", Via: global.Consola, Comando: "deepseek exec"})
	if err := c.Guardar(); err != nil {
		t.Fatal(err)
	}

	if _, err := Instalar(t.TempDir(), Opciones{}); err != nil {
		t.Fatal(err)
	}

	leido, err := global.Leer()
	if err != nil {
		t.Fatal(err)
	}
	if _, hay := leido.Resolver("deepseek"); !hay {
		t.Error("sf install borró un modelo aprobado")
	}
}

// Mudarse de harness es el único caso en que se toca un mapa que ya existe, y
// hay que pedirlo explícito.
func TestHarnessExplicitoActualizaElMapaExistente(t *testing.T) {
	enUnHomeDePrueba(t)
	if err := global.Semilla("claude-code").Guardar(); err != nil {
		t.Fatal(err)
	}

	if _, err := Instalar(t.TempDir(), Opciones{Harness: "codex"}); err != nil {
		t.Fatal(err)
	}

	leido, err := global.Leer()
	if err != nil {
		t.Fatal(err)
	}
	if leido.Harness != "codex" {
		t.Errorf("harness %q, quería codex", leido.Harness)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Desinstalar
// ────────────────────────────────────────────────────────────────────────────

func TestDesinstalarSacaLoQueEscribimos(t *testing.T) {
	enUnHomeDePrueba(t)
	raiz := t.TempDir()
	if _, err := Instalar(raiz, Opciones{}); err != nil {
		t.Fatal(err)
	}

	if _, err := Desinstalar(raiz); err != nil {
		t.Fatal(err)
	}

	for _, n := range destinos {
		if _, err := os.Stat(filepath.Join(raiz, n)); err == nil {
			t.Errorf("%s sigue ahí", n)
		}
	}
}

// Si alguien lo editó, tiene cosas que nosotros no pusimos. Borrarlo sería
// tirar eso.
func TestDesinstalarNoBorraLoQueEditaste(t *testing.T) {
	enUnHomeDePrueba(t)
	raiz := t.TempDir()
	if _, err := Instalar(raiz, Opciones{}); err != nil {
		t.Fatal(err)
	}

	ruta := filepath.Join(raiz, "CLAUDE.md")
	b, _ := os.ReadFile(ruta)
	if err := os.WriteFile(ruta, append(b, []byte("\n## lo mío\n")...), 0o644); err != nil {
		t.Fatal(err)
	}

	r, err := Desinstalar(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(ruta); err != nil {
		t.Error("borró un CLAUDE.md editado")
	}
	if !strings.Contains(strings.Join(r.Salteados, " "), "editaste") {
		t.Errorf("no avisó por qué no lo tocó: %v", r.Salteados)
	}
}

// Los modelos son de la máquina, no del proyecto: sacar SpecForge de un repo no
// puede borrar la lista que Javier construyó en los otros quince.
func TestDesinstalarNoTocaElMapaGlobal(t *testing.T) {
	enUnHomeDePrueba(t)
	raiz := t.TempDir()
	if _, err := Instalar(raiz, Opciones{}); err != nil {
		t.Fatal(err)
	}

	if _, err := Desinstalar(raiz); err != nil {
		t.Fatal(err)
	}

	if _, err := global.Leer(); err != nil {
		t.Errorf("desinstalar se llevó ~/.specforge/: %v", err)
	}
}
