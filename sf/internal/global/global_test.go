package global

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// SPECFORGE_HOME es lo que permite que estos tests no escriban en el home de
// verdad de quien los corre. t.Setenv lo restaura solo al terminar.
func enUnHomeDePrueba(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("SPECFORGE_HOME", dir)
	return dir
}

func TestSinInstalarNoHayMapa(t *testing.T) {
	enUnHomeDePrueba(t)

	if _, err := Leer(); !errors.Is(err, ErrNoHay) {
		t.Fatalf("Leer: %v, quería ErrNoHay", err)
	}
}

func TestLaSemillaGuardaYVuelve(t *testing.T) {
	dir := enUnHomeDePrueba(t)

	c := Semilla("claude-code")
	if err := c.Guardar(); err != nil {
		t.Fatal(err)
	}

	leido, err := Leer()
	if err != nil {
		t.Fatal(err)
	}
	if leido.Harness != "claude-code" {
		t.Errorf("harness %q", leido.Harness)
	}

	// Los nativos tienen que estar: son los que la propia máquina elige por
	// default, y un default sin declarar pararía el primer `sf next`.
	for _, n := range nativos {
		m, hay := leido.Buscar(n)
		if !hay {
			t.Errorf("falta el nativo %q", n)
			continue
		}
		if m.Via != Subagente {
			t.Errorf("%s: via %q, quería subagente", n, m.Via)
		}
	}

	// Y el archivo tiene que ser legible para un humano: lo va a abrir alguien.
	b, err := os.ReadFile(filepath.Join(dir, Archivo))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(b), "#") {
		t.Errorf("el archivo no arranca con el encabezado:\n%s", b)
	}
}

// El caso que justifica todo el paquete: declarar un modelo de otro proveedor,
// con su comando, y que sobreviva a la ida y vuelta al disco.
func TestDeclararUnModeloDeConsola(t *testing.T) {
	enUnHomeDePrueba(t)

	c := Semilla("claude-code")
	c.Declarar("deepseek", Modelo{Via: Consola, Comando: "deepseek exec"})
	if err := c.Guardar(); err != nil {
		t.Fatal(err)
	}

	leido, err := Leer()
	if err != nil {
		t.Fatal(err)
	}
	m, hay := leido.Buscar("deepseek")
	if !hay {
		t.Fatal("deepseek no sobrevivió al disco")
	}
	if m.Via != Consola || m.Comando != "deepseek exec" {
		t.Errorf("deepseek = %+v", m)
	}
}

// Guardar tiene que crear la carpeta: `sf install` corre en una máquina donde
// ~/.specforge/ todavía no existe, que es el caso normal la primera vez.
func TestGuardarCreaLaCarpeta(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "no-existe-todavia")
	t.Setenv("SPECFORGE_HOME", dir)

	if err := Semilla("x").Guardar(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, Archivo)); err != nil {
		t.Errorf("no creó la carpeta: %v", err)
	}
}

func TestDetectarHarness(t *testing.T) {
	// La detección es una PISTA: que la variable esté es evidencia; que no esté
	// no prueba nada. Por eso "desconocido" es un resultado válido y no un
	// error — y por eso el que llama puede pisarlo con --harness.
	t.Setenv("CLAUDECODE", "")
	t.Setenv("CLAUDE_CODE_ENTRYPOINT", "")
	if h := DetectarHarness(); h != "desconocido" {
		t.Errorf("sin señales detectó %q", h)
	}

	t.Setenv("CLAUDECODE", "1")
	if h := DetectarHarness(); h != "claude-code" {
		t.Errorf("con CLAUDECODE detectó %q", h)
	}
}
