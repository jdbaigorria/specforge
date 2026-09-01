package lanzar

import (
	"slices"
	"strings"
	"testing"
)

func base() Pedido {
	return Pedido{
		Raiz:   "/tmp/proy",
		Prompt: "Usá la skill sfp-po.\n\nestado: prd",
	}
}

// tiene dice si `arg` está en la línea, y con qué valor lo sigue.
func tiene(linea []string, arg string) (string, bool) {
	i := slices.Index(linea, arg)
	if i < 0 {
		return "", false
	}
	if i+1 < len(linea) {
		return linea[i+1], true
	}
	return "", true
}

func TestLineaDeClaudeCode(t *testing.T) {
	p := base()
	p.Harness, p.Modelo, p.Esfuerzo = "claude-code", "opus", "high"

	l, err := Linea(p)
	if err != nil {
		t.Fatal(err)
	}
	if l[0] != "claude" {
		t.Errorf("el binario es %q", l[0])
	}
	if v, _ := tiene(l, "--model"); v != "opus" {
		t.Errorf("--model %q", v)
	}
	if v, _ := tiene(l, "--effort"); v != "high" {
		t.Errorf("--effort %q", v)
	}
	if v, _ := tiene(l, "--output-format"); v != "stream-json" {
		t.Errorf("--output-format %q", v)
	}
	// `--verbose` no es decoración: sin él, `--output-format stream-json` no
	// emite el stream. Medido — la corrida de §2 lo llevaba.
	if _, hay := tiene(l, "--verbose"); !hay {
		t.Error("falta --verbose, sin el cual stream-json no emite")
	}
	if v, _ := tiene(l, "--permission-mode"); v != "bypassPermissions" {
		t.Errorf("--permission-mode %q", v)
	}
}

// EL ESFUERZO DE OPENCODE NO SE LLAMA --effort.
//
// Es `--variant`, y confundirlos es la clase de error que este proyecto ya
// cometió dos veces adivinando flags. Cada arnés tiene su nombre y la tabla es
// lo único que sabe cuál.
func TestLineaDeOpencodeUsaVariantYNoEffort(t *testing.T) {
	p := base()
	p.Harness, p.Modelo, p.Esfuerzo = "opencode", "opencode/nemotron-3-ultra-free", "high"

	l, err := Linea(p)
	if err != nil {
		t.Fatal(err)
	}
	if l[0] != "opencode" || l[1] != "run" {
		t.Errorf("headless en opencode es el subcomando `run`: %v", l[:2])
	}
	if v, _ := tiene(l, "--variant"); v != "high" {
		t.Errorf("--variant %q", v)
	}
	if _, hay := tiene(l, "--effort"); hay {
		t.Error("le pasó --effort a opencode, que no lo tiene")
	}
	if v, _ := tiene(l, "--dir"); v != "/tmp/proy" {
		t.Errorf("--dir %q", v)
	}
	if v, _ := tiene(l, "--format"); v != "json" {
		t.Errorf("--format %q", v)
	}
	// `-p` en opencode es --password. Un `sf lanzar` que lo asuma se rompe en el
	// primer intento, y por eso hay un test que lo mira.
	if _, hay := tiene(l, "-p"); hay {
		t.Error("le pasó -p a opencode, donde -p es --password")
	}
}

func TestLineaDeCommandCode(t *testing.T) {
	p := base()
	p.Harness, p.Modelo, p.Esfuerzo = "commandcode", "deepseek/deepseek-v4-flash", "low"

	l, err := Linea(p)
	if err != nil {
		t.Fatal(err)
	}
	if l[0] != "cmd" {
		t.Errorf("el binario es %q", l[0])
	}
	if _, hay := tiene(l, "-p"); !hay {
		t.Error("falta -p")
	}
	if v, _ := tiene(l, "--effort"); v != "low" {
		t.Errorf("--effort %q", v)
	}
	if v, _ := tiene(l, "--output-format"); v != "json" {
		t.Errorf("--output-format %q", v)
	}
	// Command Code no tiene flag de carpeta: corre en la cwd. Quien ejecute
	// tiene que poner `cmd.Dir`, y por eso Linea no puede ser lo único.
	if _, hay := tiene(l, "--dir"); hay {
		t.Error("le inventó un --dir a Command Code")
	}
}

// UN ESFUERZO VACÍO NO EMITE NADA.
//
// Vacío es "el que traiga el modelo", que NO es lo mismo que "bajo". Emitir
// `--effort ""` le estaría pasando un valor que Javier no declaró, y sería sf
// eligiendo un esfuerzo — lo mismo que tiene prohibido hacer con un modelo.
func TestSinEsfuerzoNoSeEmiteElFlag(t *testing.T) {
	for _, h := range []string{"claude-code", "opencode", "commandcode"} {
		p := base()
		p.Harness, p.Modelo = h, "un/modelo"

		l, err := Linea(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, flag := range []string{"--effort", "--variant"} {
			if _, hay := tiene(l, flag); hay {
				t.Errorf("%s: emitió %s sin que nadie declarara un esfuerzo: %v", h, flag, l)
			}
		}
		if slices.Contains(l, "") {
			t.Errorf("%s: la línea tiene un argumento vacío: %v", h, l)
		}
	}
}

// EL PROMPT VIAJA COMO UN SOLO ARGUMENTO, con espacios y saltos de línea adentro.
//
// Partirlo por espacios es exactamente el bug que `suite.Correr` ya documenta
// para los `test_cmd`, y acá sería peor: el sobre entero llegaría hecho pedazos
// y el modelo trabajaría con la mitad.
func TestElPromptEsUnSoloArgumento(t *testing.T) {
	for _, h := range []string{"claude-code", "opencode", "commandcode"} {
		p := base()
		p.Harness, p.Modelo = h, "un/modelo"

		l, err := Linea(p)
		if err != nil {
			t.Fatal(err)
		}
		if !slices.Contains(l, p.Prompt) {
			t.Errorf("%s: el prompt no viaja entero: %v", h, l)
		}
	}
}

// `--yolo` SÓLO en Command Code, SÓLO si Javier lo declaró, y NUNCA por default.
//
// Es la decisión de §9 de headless.md llevada al código: sf lo transporta, no lo
// elige. Los otros dos no lo necesitan —opencode es permisivo de arranque y
// Claude Code usa bypassPermissions— así que pedirlo ahí no hace nada.
func TestYoloSoloEnCommandCodeYSoloSiSePide(t *testing.T) {
	p := base()
	p.Harness, p.Modelo = "commandcode", "un/modelo"

	sin, _ := Linea(p)
	if slices.Contains(sin, "--yolo") {
		t.Error("pasó --yolo sin que nadie lo declarara")
	}

	p.Yolo = true
	con, _ := Linea(p)
	if !slices.Contains(con, "--yolo") {
		t.Error("Javier lo declaró y no lo pasó")
	}

	// Y en los otros dos no aparece aunque se pida: no tienen ese flag.
	for _, h := range []string{"claude-code", "opencode"} {
		p.Harness = h
		l, _ := Linea(p)
		if slices.Contains(l, "--yolo") {
			t.Errorf("%s: le pasó --yolo, que no existe ahí", h)
		}
	}
}

func TestLineaDeUnArnesQueNoSabemosLanzarEsError(t *testing.T) {
	p := base()
	p.Harness, p.Modelo = "emacs", "un/modelo"

	if _, err := Linea(p); err == nil {
		t.Fatal("armó una línea para un arnés que no conoce")
	}
}

// Sin modelo no se lanza. Un arnés sin `-m` corre con SU default, que es
// exactamente lo que este proyecto viene evitando desde H1: el modelo lo elige
// el catálogo de Javier, nunca el proveedor.
func TestSinModeloNoSeLanza(t *testing.T) {
	p := base()
	p.Harness = "opencode"

	if _, err := Linea(p); err == nil {
		t.Fatal("armó una línea sin modelo: el arnés correría con su default")
	}
}

// Mostrar es para `--seco`, y tiene que poder pegarse en una terminal.
func TestMostrarEntrecomillaLoQueLoNecesita(t *testing.T) {
	p := base()
	p.Harness, p.Modelo = "opencode", "un/modelo"
	l, _ := Linea(p)

	s := Mostrar(l)
	if !strings.Contains(s, "'Usá la skill sfp-po.") {
		t.Errorf("no entrecomilló el prompt:\n%s", s)
	}
	if !strings.HasPrefix(s, "opencode run ") {
		t.Errorf("no empieza por el comando:\n%s", s)
	}
}
