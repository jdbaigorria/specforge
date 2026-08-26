package doctor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// escribirSkill deja un SKILL.md en el home falso del test.
func escribirSkill(t *testing.T, home, nombre, cuerpo string) {
	t.Helper()
	d := filepath.Join(home, ".claude", "skills", nombre)
	if err := os.MkdirAll(d, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(d, "SKILL.md"), []byte(cuerpo), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// soloCodigo — el filtro que decide qué cuenta como un comando
// ────────────────────────────────────────────────────────────────────────────

func TestLaProsaNoEsUnComandoAunqueEsteEnUnCartel(t *testing.T) {
	// Este caso es literal: es la línea de sfp-backlog que hizo que el doctor
	// acusara a un skill correcto la primera vez que corrió.
	texto := "## Why\n\n```\nsf does not verify the judgment. It verifies THAT IT HAPPENED.\n```\n"

	if c := soloCodigo(texto); strings.Contains(c, "does") {
		t.Errorf("un ``` pelado es una cita, no shell.\nquedó: %q", c)
	}
}

func TestLoQueEstaEnBashSiEsUnComando(t *testing.T) {
	texto := "```bash\nsf context\nsf done\n```\n"

	c := soloCodigo(texto)
	for _, q := range []string{"sf context", "sf done"} {
		if !strings.Contains(c, q) {
			t.Errorf("%q se perdió.\nquedó: %q", q, c)
		}
	}
}

func TestLosBackticksSueltosCuentan(t *testing.T) {
	// La mayoría de las menciones viven acá: `sf approve` en medio de una
	// oración. Si se perdieran, el chequeo no vería casi nada.
	texto := "Then run `sf approve` and the state moves.\n"

	if c := soloCodigo(texto); !strings.Contains(c, "sf approve") {
		t.Errorf("se perdió el inline.\nquedó: %q", c)
	}
}

func TestLaProsaAlrededorDeUnBacktickNoEntra(t *testing.T) {
	texto := "sf does not care. Run `sf next` when ready.\n"

	c := soloCodigo(texto)
	if strings.Contains(c, "does") {
		t.Errorf("entró la prosa.\nquedó: %q", c)
	}
	if !strings.Contains(c, "sf next") {
		t.Errorf("se perdió el comando.\nquedó: %q", c)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// comandosDesconocidos — la desalineación, que es para lo que existe
// ────────────────────────────────────────────────────────────────────────────

func TestUnSkillQueNombraUnComandoInexistenteSeDetecta(t *testing.T) {
	// El caso real: un skill de una versión más nueva que el binario. El bucle
	// se traba, y el que lo descubre hoy es un agente a mitad de camino.
	home := t.TempDir()
	escribirSkill(t, home, "sfp-scout", "```bash\nsf context\nsf sellalo\n```\n")

	d := comandosDesconocidos(filepath.Join(home, ".claude", "skills", "sfp-scout", "SKILL.md"))
	if len(d) != 1 || d[0] != "sellalo" {
		t.Fatalf("esperaba [sellalo], vino %v", d)
	}
}

func TestUnSkillAlineadoNoReportaNada(t *testing.T) {
	home := t.TempDir()
	escribirSkill(t, home, "sf-build", "```bash\nsf context\nsf lote start\nsf done --msg \"x\"\n```\n")

	if d := comandosDesconocidos(filepath.Join(home, ".claude", "skills", "sf-build", "SKILL.md")); len(d) > 0 {
		t.Fatalf("no esperaba nada, vino %v", d)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Revisar — el informe entero
// ────────────────────────────────────────────────────────────────────────────

func TestSinSkillsInstaladosNoEstaSano(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	i := Revisar(t.TempDir(), "v-test")

	if i.Sano() {
		t.Fatal("sin ningún skill instalado el bucle no puede correr, y el doctor tiene que decirlo")
	}
	if len(i.Skills) != 9 {
		t.Fatalf("son nueve estados, vinieron %d", len(i.Skills))
	}
	for _, s := range i.Skills {
		if s.Instalado() {
			t.Fatalf("%s no debería estar instalado en un home vacío (%s)", s.Nombre, s.Ruta)
		}
	}
	if !strings.Contains(strings.Join(i.Fallas, "\n"), "9 skills") {
		t.Errorf("la falla tiene que nombrar cuántos faltan: %v", i.Fallas)
	}
}

func TestElInformeDiceDondeEncontroCadaSkill(t *testing.T) {
	// Que la ruta se muestre no es adorno: la mitad de los problemas de
	// instalación se entienden viendo de dónde salió el archivo.
	home := t.TempDir()
	t.Setenv("HOME", home)
	escribirSkill(t, home, "sfp-scout", "```bash\nsf done\n```\n")

	i := Revisar(t.TempDir(), "v-test")

	var scout Skill
	for _, s := range i.Skills {
		if s.Nombre == "sfp-scout" {
			scout = s
		}
	}
	if !scout.Instalado() {
		t.Fatal("sfp-scout estaba en ~/.claude/skills y no se encontró")
	}
	if !strings.Contains(scout.Ruta, filepath.Join(".claude", "skills", "sfp-scout")) {
		t.Errorf("la ruta no dice dónde: %q", scout.Ruta)
	}
}

func TestElDoctorImprimeAlgoAunqueTodoEsteBien(t *testing.T) {
	// Un doctor que no imprime nada cuando todo anda deja al que lo corrió sin
	// saber si funcionó o si el comando no hizo nada.
	home := t.TempDir()
	t.Setenv("HOME", home)

	txt := Revisar(t.TempDir(), "v9.9.9").Texto()

	for _, q := range []string{"binario", "v9.9.9", "skills", "harness", "proyecto"} {
		if !strings.Contains(txt, q) {
			t.Errorf("el informe no dice %q:\n%s", q, txt)
		}
	}
}

func TestCuandoFallaDiceComoSeArregla(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)

	txt := Revisar(t.TempDir(), "v-test").Texto()

	if !strings.Contains(txt, "cómo se arregla") {
		t.Fatalf("un diagnóstico sin salida obliga a ir a buscar el README:\n%s", txt)
	}
	if !strings.Contains(txt, "/plugin install specforge") {
		t.Errorf("faltan los nueve skills y no dice cómo instalarlos:\n%s", txt)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Los skills DEL REPO, no los instalados
// ────────────────────────────────────────────────────────────────────────────

// TestLosSkillsDelRepoEstanAlineados es el chequeo que le faltaba al CI.
//
// Cuando `sfp-scout` se contradijo —"dejá el veredicto vacío" arriba, "la
// compuerta exige un veredicto válido" abajo— y trabó el ⑥ sin salida, el CI
// estaba en verde. Validaba tres cosas de los skills:
//
//	que el `name:` coincida con la carpeta
//	que ninguno nombre un comando muerto del producto viejo
//	que los nueve del mapa existan como carpeta
//
// Las tres miran el NOMBRE del skill. Ninguna mira lo que el skill PRODUCE
// contra lo que la máquina ACEPTA, que es donde vivía el bug.
//
// Éste no lo agarra tampoco —una contradicción en prosa no la ve ningún
// parser— pero sí agarra a su primo mecánico, que es el que se va a repetir:
// un skill que le pide al agente correr un comando que este binario no tiene.
// Y a diferencia del doctor, que mira la máquina de quien lo corre, éste mira
// el repo: falla en el PR, antes de que nadie instale nada.
func TestLosSkillsDelRepoEstanAlineados(t *testing.T) {
	raiz := filepath.Join("..", "..", "..", "skills")

	dirs, err := os.ReadDir(raiz)
	if err != nil {
		t.Fatalf("no encontré skills/ desde %s: %v", raiz, err)
	}

	var vistos int
	for _, d := range dirs {
		if !d.IsDir() {
			continue
		}
		ruta := filepath.Join(raiz, d.Name(), "SKILL.md")
		if _, err := os.Stat(ruta); err != nil {
			continue
		}
		vistos++

		if c := comandosDesconocidos(ruta); len(c) > 0 {
			t.Errorf("%s le pide al agente `sf %s`, y este binario no lo tiene",
				d.Name(), strings.Join(c, "` y `sf "))
		}
	}

	// Sin esta línea el test pasaría en verde si la ruta a skills/ cambiara y
	// el ReadDir devolviera una carpeta vacía — que es la forma exacta en que
	// un chequeo deja de chequear sin que nadie se entere.
	if vistos < 9 {
		t.Fatalf("esperaba al menos los 9 de la máquina y encontré %d en %s", vistos, raiz)
	}
}
