package lanzar

import (
	"errors"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/maquina"
)

func instruccion() maquina.Instruccion {
	return maquina.Instruccion{
		Tipo:     maquina.Trabajar,
		Estado:   "prd",
		Skill:    "sfp-po",
		Perfil:   "razonar",
		Modelo:   "opencode/nemotron-3-ultra-free",
		Esfuerzo: "high",
		Via:      "subagente",
	}
}

// EL PROMPT APUNTA A LA SKILL, NO LA PEGA.
//
// La corrección de Javier del 2026-08-31, medida: los arneses resuelven solos la
// composición, los `references/` y los `templates/` — tres niveles, sin escalar
// permisos. Así que `sf lanzar` no tiene que entender la estructura interna de
// ninguna skill, que era el acoplamiento nuevo que este diseño temía.
func TestElPromptNombraLaSkillYLlevaElSobre(t *testing.T) {
	p := Armar(instruccion(), "el sobre entero\ncon sus renglones")

	if !strings.Contains(p, "sfp-po") {
		t.Errorf("no nombra la skill:\n%s", p)
	}
	if !strings.Contains(p, "el sobre entero") {
		t.Errorf("no lleva el sobre:\n%s", p)
	}
	// Y no pega el texto de la skill: si apareciera algo de adentro del archivo,
	// es que alguien lo leyó y lo incrustó.
	if strings.Contains(p, "SKILL.md") {
		t.Errorf("parece estar pegando el archivo:\n%s", p)
	}
}

// Sin skill no hay a qué apuntar. Pasa en las paradas, y por eso `Puede` las
// frena antes — pero si alguien llama igual, esto no puede inventar una skill.
func TestSinSkillNoSePuedeArmarUnPrompt(t *testing.T) {
	i := instruccion()
	i.Skill = ""

	if p := Armar(i, "sobre"); strings.Contains(p, "Usá la skill") {
		t.Errorf("inventó una skill:\n%s", p)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Puede: qué se puede lanzar y qué no
// ────────────────────────────────────────────────────────────────────────────

func TestPuedeLanzarUnPasoQueNoConversa(t *testing.T) {
	if err := Puede(instruccion()); err != nil {
		t.Errorf("no dejó lanzar el ⑦, que no conversa: %v", err)
	}
}

// LOS DOS QUE CONVERSAN NO SE DELEGAN A NADIE, ni a un subagente ni a un proceso.
//
// El ⑥ porque ①–⑤ es un pinponeo, y el ⑧ porque la constitución no se deriva del
// PRD: arquitectura, stack y convenciones son decisiones de Javier. Es la misma
// regla que sacó al ⑧ de ser subagente el 2026-08-31, aplicada al modo nuevo.
func TestNoLanzaLosQueConversan(t *testing.T) {
	for _, est := range []string{"brief", "constitucion"} {
		i := instruccion()
		i.Estado, i.Via = est, "vos"

		err := Puede(i)
		if err == nil {
			t.Errorf("%s conversa y lo dejó lanzar", est)
			continue
		}
		if !strings.Contains(err.Error(), "Javier") {
			t.Errorf("%s: el motivo no dice que es de Javier: %v", est, err)
		}
	}
}

// Una parada no es trabajo: no hay nada que lanzar hasta que alguien conteste.
//
// Y se distingue con un sentinel, no con un texto, porque QUIEN LLAMA TIENE ALGO
// MEJOR QUE DECIR: la máquina ya explica por qué frenó —"el paso pide el perfil
// construir y claude-code no lo tiene declarado", con el comando para
// arreglarlo—. Tapar eso con un "no hay trabajo" genérico es contestar peor de
// lo que sf ya sabe contestar.
func TestNoLanzaEnUnaParadaYSeDistingue(t *testing.T) {
	for _, tipo := range []maquina.Tipo{maquina.Para, maquina.Barata, maquina.MeTrabe, maquina.Fin} {
		i := instruccion()
		i.Tipo = tipo

		err := Puede(i)
		if err == nil {
			t.Errorf("tipo %v: lanzó sin que hubiera trabajo", tipo)
			continue
		}
		if !errors.Is(err, ErrNoHayTrabajo) {
			t.Errorf("tipo %v: no es distinguible del resto: %v", tipo, err)
		}
	}
}

// Y los motivos que SÍ son de este paquete no se confunden con aquél: de esos,
// sf lanzar es el único que sabe. La máquina no tiene nada que decir sobre por
// qué un `via: consola` no se lanza headless.
func TestLosMotivosPropiosNoSonElSentinelDeLaMaquina(t *testing.T) {
	for _, c := range []struct {
		que string
		i   maquina.Instruccion
	}{
		{"conversa", func() maquina.Instruccion { i := instruccion(); i.Via = "vos"; return i }()},
		{"consola", func() maquina.Instruccion { i := instruccion(); i.Via = "consola"; return i }()},
		{"sin skill", func() maquina.Instruccion { i := instruccion(); i.Skill = ""; return i }()},
	} {
		err := Puede(c.i)
		if err == nil {
			t.Errorf("%s: lo dejó pasar", c.que)
			continue
		}
		if errors.Is(err, ErrNoHayTrabajo) {
			t.Errorf("%s: se disfrazó de parada de la máquina: %v", c.que, err)
		}
	}
}

// Un `via: consola` es un modelo que se invoca a mano y cuyo prefijo declaró
// Javier: no pasa por la línea de un arnés. Lanzarlo con la tabla de arneses
// sería correr otra cosa distinta de la que él declaró.
func TestNoLanzaUnViaConsola(t *testing.T) {
	i := instruccion()
	i.Via = "consola"

	err := Puede(i)
	if err == nil {
		t.Fatal("lanzó un via: consola con la línea de un arnés")
	}
	if !strings.Contains(err.Error(), "consola") {
		t.Errorf("el motivo no lo nombra: %v", err)
	}
}

// FRENAR ANTES DE LANZAR VALE UNA CORRIDA.
//
// Sin permiso declarado, Command Code arranca, tarda medio minuto, gasta tokens
// y devuelve exit=0 con `subtype: success` y ningún artefacto. Enterarse antes es
// gratis.
func TestSoloCommandCodeNecesitaPermisoDeclarado(t *testing.T) {
	if !NecesitaPermiso("commandcode") {
		t.Error("Command Code no escribe ni ejecuta en headless sin --yolo")
	}
	for _, h := range []string{"claude-code", "opencode"} {
		if NecesitaPermiso(h) {
			t.Errorf("%s no necesita ningún permiso declarado", h)
		}
	}
}
