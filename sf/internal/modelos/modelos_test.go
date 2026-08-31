package modelos

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// salidaOpencode es lo que `opencode models` imprime de verdad, recortado.
//
// Un id por línea y nada más: ni encabezado, ni descripciones, ni pie.
const salidaOpencode = `opencode/big-pickle
opencode/ling-3.0-flash-fin-free
opencode/nemotron-3-ultra-free
openrouter/z-ai/glm-5.3
openrouter/z-ai/glm-5v-turbo
`

// salidaCommandCode es lo que `cmd --list-models` imprime de verdad, recortado.
//
// Tiene encabezado, secciones por proveedor, dos columnas alineadas con espacios
// y un pie con un link. Los tres estorbos están acá a propósito.
const salidaCommandCode = `Available models  ·  63 models

Open Source

deepseek/deepseek-v4-pro               hybrid-attention long-context reasoning
moonshotai/kimi-k3                     long-horizon coding & knowledge work with 1M context
zai-org/glm-5.2                        powerful coding with 1M context and long-horizon tasks

Frontier

claude-sonnet-5                        best combo of speed & intelligence (recommended)
gpt-5.6-sol                            frontier model for complex professional work

Pass the full id, or just the short name after the last "/":
cmd --model moonshotai/kimi-k2.5

Docs:  https://commandcode.ai/docs/reference/cli/models
`

func TestParsearOpencodeDevuelveLosIds(t *testing.T) {
	got := parsearOpencode(salidaOpencode)

	if len(got) != 5 {
		t.Fatalf("esperaba 5 modelos, hubo %d: %+v", len(got), got)
	}
	if got[0].ID != "opencode/big-pickle" {
		t.Errorf("primer id = %q", got[0].ID)
	}
	// Un id con dos barras es un id, no un error de parseo.
	if got[3].ID != "openrouter/z-ai/glm-5.3" {
		t.Errorf("id con dos barras = %q", got[3].ID)
	}
	for _, m := range got {
		if m.Para != "" {
			t.Errorf("opencode no da descripciones y %q trajo %q", m.ID, m.Para)
		}
	}
}

// El ruido de los plugins es real: `opencode models` con un plugin cargado
// imprime "[icm] plugin loaded (icm 0.10.50)". Sale por stderr —por eso Listar
// lee sólo stdout— pero el parser no puede confiar en eso: cualquier plugin
// puede escribir en stdout y ahí el ruido entra igual.
func TestParsearOpencodeIgnoraElRuidoDeLosPlugins(t *testing.T) {
	sucia := "[icm] plugin loaded (icm 0.10.50)\n" + salidaOpencode + "\nlisto\n"

	for _, m := range parsearOpencode(sucia) {
		if strings.Contains(m.ID, " ") || !strings.Contains(m.ID, "/") {
			t.Errorf("se coló una línea que no es un id: %q", m.ID)
		}
	}
	if len(parsearOpencode(sucia)) != 5 {
		t.Errorf("el ruido cambió la cuenta: %d", len(parsearOpencode(sucia)))
	}
}

func TestParsearCommandCodeSeparaIdYDescripcion(t *testing.T) {
	got := parsearCommandCode(salidaCommandCode)

	if len(got) != 5 {
		t.Fatalf("esperaba 5 modelos, hubo %d: %+v", len(got), got)
	}
	if got[0].ID != "deepseek/deepseek-v4-pro" {
		t.Errorf("primer id = %q", got[0].ID)
	}
	if got[0].Para != "hybrid-attention long-context reasoning" {
		t.Errorf("primera descripción = %q", got[0].Para)
	}
	// La descripción se copia entera, con espacios adentro y con paréntesis.
	if got[3].Para != "best combo of speed & intelligence (recommended)" {
		t.Errorf("descripción con paréntesis = %q", got[3].Para)
	}
}

// Un id de Command Code NO siempre tiene barra: `claude-sonnet-5` y `gpt-5.6-sol`
// son ids válidos. Un parser que exigiera la barra —como el de opencode, donde
// sí la tienen todos— se comería catorce modelos reales sin decir nada.
func TestParsearCommandCodeAceptaIdsSinBarra(t *testing.T) {
	got := parsearCommandCode(salidaCommandCode)

	var hay bool
	for _, m := range got {
		if m.ID == "claude-sonnet-5" {
			hay = true
		}
	}
	if !hay {
		t.Errorf("se perdió un id sin barra: %+v", got)
	}
}

// El encabezado, las secciones y el pie tienen que quedar afuera. El pie es el
// que muerde: `Docs:  https://…` tiene la misma forma que un modelo —una palabra,
// dos espacios, texto— y sería el único falso positivo del formato.
func TestParsearCommandCodeDescartaEncabezadoSeccionesYPie(t *testing.T) {
	for _, m := range parsearCommandCode(salidaCommandCode) {
		switch {
		case strings.HasPrefix(m.ID, "Docs"):
			t.Errorf("el pie entró como modelo: %+v", m)
		case strings.HasPrefix(m.ID, "Available"):
			t.Errorf("el encabezado entró como modelo: %+v", m)
		case m.ID == "Open" || m.ID == "Frontier":
			t.Errorf("una sección entró como modelo: %+v", m)
		case strings.HasPrefix(m.Para, "http"):
			t.Errorf("una línea de doc entró como modelo: %+v", m)
		}
	}
}

// Un listado vacío no es un parseo exitoso de cero modelos: es que el formato
// cambió. Quien llame tiene que poder distinguirlo, y por eso Listar lo convierte
// en error en vez de devolver una lista vacía y quedarse callado.
func TestListarConSalidaQueNoSeEntiendeEsError(t *testing.T) {
	viejo := correr
	defer func() { correr = viejo }()
	correr = func(context.Context, string, ...string) ([]byte, error) {
		return []byte("Uso: opencode models [provider]\n"), nil
	}

	if _, err := Listar("opencode"); err == nil {
		t.Fatal("una salida que no trae ningún id tiene que ser error")
	}
}

func TestListarPropagaQueElArnesNoEstaInstalado(t *testing.T) {
	viejo := correr
	defer func() { correr = viejo }()
	roto := errors.New("exec: \"opencode\": executable file not found in $PATH")
	correr = func(context.Context, string, ...string) ([]byte, error) {
		return nil, roto
	}

	_, err := Listar("opencode")
	if err == nil {
		t.Fatal("si el binario no está, Listar tiene que fallar")
	}
	if !strings.Contains(err.Error(), "opencode") {
		t.Errorf("el error no nombra al arnés: %v", err)
	}
}

// Claude Code no tiene listado y nunca lo tuvo. Devolver una lista de alias
// escrita a mano acá sería sf opinando sobre qué modelos hay —la tabla que se
// pudre que `global.Modelo` ya se prohíbe—, así que dice que no sabe.
func TestListarClaudeCodeDiceQueNoHayListado(t *testing.T) {
	_, err := Listar("claude-code")
	if !errors.Is(err, ErrSinListado) {
		t.Fatalf("esperaba ErrSinListado, hubo %v", err)
	}
}

func TestListarArnesDesconocidoEsOtroError(t *testing.T) {
	_, err := Listar("emacs")
	if !errors.Is(err, ErrArnesDesconocido) {
		t.Fatalf("esperaba ErrArnesDesconocido, hubo %v", err)
	}
	if errors.Is(err, ErrSinListado) {
		t.Error("un arnés inexistente NO es lo mismo que uno sin listado: " +
			"el primero es un error de quien llama, el segundo es un hecho del mundo")
	}
	if !strings.Contains(err.Error(), "emacs") {
		t.Errorf("el error no dice cuál se pidió: %v", err)
	}
}

func TestListarNormalizaLoQueDevuelveElArnes(t *testing.T) {
	viejo := correr
	defer func() { correr = viejo }()
	correr = func(_ context.Context, nombre string, args ...string) ([]byte, error) {
		if nombre != "cmd" || len(args) != 1 || args[0] != "--list-models" {
			t.Errorf("a Command Code se le pregunta con `cmd --list-models`, no %q %v", nombre, args)
		}
		return []byte(salidaCommandCode), nil
	}

	got, err := Listar("commandcode")
	if err != nil {
		t.Fatalf("Listar: %v", err)
	}
	if len(got) != 5 {
		t.Fatalf("esperaba 5, hubo %d", len(got))
	}
	if got[0].Para == "" {
		t.Error("la descripción se perdió en el camino")
	}
}
