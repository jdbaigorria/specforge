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
	if r.Modelos == 0 {
		t.Error("no sembró ningún modelo")
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
	c.Declarar("deepseek", global.Modelo{Via: global.Consola, Comando: "deepseek exec"})
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
	if _, hay := leido.Buscar("deepseek"); !hay {
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
