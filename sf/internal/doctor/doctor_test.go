package doctor

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/maquina"
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

	// Se CAMINA el árbol en vez de leer un nivel: los skills están agrupados
	// (`skills/maquina/`, `skills/utiles/`, `skills/contrib/`), y un `ReadDir`
	// del nivel de arriba devuelve las tres carpetas, ninguna con `SKILL.md`.
	// Eso no falla: pasa en verde contando cero, que es la forma exacta en que
	// un chequeo deja de chequear sin que nadie se entere — el mismo riesgo que
	// el `vistos < 9` de abajo ya cubría, y que agrupar volvió real.
	//
	// `contrib/` entra a propósito: no se publica en `plugin.json`, pero si
	// nombra un `sf` que no existe está roto igual, y alguien lo va a copiar.
	var vistos int
	err := filepath.WalkDir(raiz, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "SKILL.md" {
			return nil
		}
		vistos++

		if c := comandosDesconocidos(ruta); len(c) > 0 {
			t.Errorf("%s le pide al agente `sf %s`, y este binario no lo tiene",
				filepath.Base(filepath.Dir(ruta)), strings.Join(c, "` y `sf "))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("no pude caminar skills/ desde %s: %v", raiz, err)
	}

	// Sin esta línea el test pasaría en verde si la ruta a skills/ cambiara y
	// el walk no encontrara nada — que es la forma exacta en que un chequeo
	// deja de chequear sin que nadie se entere.
	if vistos < 9 {
		t.Fatalf("esperaba al menos los 9 de la máquina y encontré %d en %s", vistos, raiz)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// LA CARPETA COMO HECHO, NO COMO CONVENCIÓN
// ────────────────────────────────────────────────────────────────────────────
//
// Agrupar los skills en `maquina/`, `utiles/` y `contrib/` sirve para leer el
// repo, y eso es todo lo que sirve mientras la agrupación sea una costumbre.
// Una costumbre se rompe sin ruido: alguien agrega un skill de estado en
// `utiles/`, o publica uno y se olvida del `plugin.json`, y nada falla hasta
// que alguien instala.
//
// Los dos tests de abajo le ponen contrato a las dos mitades:
//
//	qué hay en maquina/     tiene que ser EXACTAMENTE lo que la máquina nombra
//	qué hay en disco        tiene que ser EXACTAMENTE lo que el plugin publica
//
// Es la misma jugada de `internal/comandos` con la lista de comandos: el dato
// vive en un lugar, y un test comprueba que la copia no se le escapó. La
// diferencia es que acá el lugar donde vive es `maquina.SkillsDeEstado()`, que
// ya existía y ya era la única copia.

// skillsEn devuelve los nombres de skill que hay bajo una carpeta del repo.
//
// Un nombre es el de la carpeta que tiene el SKILL.md, que es de donde lo
// toma el harness cuando el frontmatter no declara `name`.
func skillsEn(t *testing.T, rel string) []string {
	t.Helper()
	var n []string
	base := filepath.Join("..", "..", "..", rel)
	err := filepath.WalkDir(base, func(ruta string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == "SKILL.md" {
			n = append(n, filepath.Base(filepath.Dir(ruta)))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("no pude caminar %s: %v", base, err)
	}
	slices.Sort(n)
	return n
}

func TestLaCarpetaMaquinaEsExactamenteLosNueve(t *testing.T) {
	quiere := maquina.SkillsDeEstado()
	slices.Sort(quiere)

	hay := skillsEn(t, filepath.Join("skills", "maquina"))

	if !slices.Equal(hay, quiere) {
		t.Errorf("skills/maquina/ no es lo que la máquina nombra.\n  en disco: %v\n  la máquina pide: %v",
			hay, quiere)
	}
}

func TestElPluginPublicaLoQueHayEnDisco(t *testing.T) {
	b, err := os.ReadFile(filepath.Join("..", "..", "..", ".claude-plugin", "plugin.json"))
	if err != nil {
		t.Fatal(err)
	}
	var m struct {
		Skills []string `json:"skills"`
	}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}

	// El array tiene que ser explícito. Con `"skills": "./skills/"` —un string,
	// que es lo que había antes de agrupar— el harness ESCANEA, y un escaneo
	// plano encuentra `maquina/` en vez de los nueve. El tipo del campo es
	// parte del contrato.
	if len(m.Skills) == 0 {
		t.Fatal("plugin.json no publica un array de rutas: con los skills agrupados, el escaneo por defecto no los encuentra")
	}

	publica := make([]string, 0, len(m.Skills))
	for _, r := range m.Skills {
		publica = append(publica, filepath.Base(r))
	}
	slices.Sort(publica)

	// `contrib/` queda afuera a propósito: está en el repo y NO se publica,
	// que es exactamente lo que era `skills-community/` antes de mudarse.
	hay := append(skillsEn(t, filepath.Join("skills", "maquina")),
		skillsEn(t, filepath.Join("skills", "utiles"))...)
	slices.Sort(hay)

	if !slices.Equal(publica, hay) {
		t.Errorf("plugin.json y el disco no coinciden.\n  publica (%d): %v\n  en disco (%d): %v",
			len(publica), publica, len(hay), hay)
	}

	for _, r := range m.Skills {
		if _, err := os.Stat(filepath.Join("..", "..", "..", r, "SKILL.md")); err != nil {
			t.Errorf("plugin.json publica %s y ahí no hay SKILL.md", r)
		}
		if strings.Contains(r, "/contrib/") {
			t.Errorf("plugin.json publica %s, y contrib/ no se publica", r)
		}
	}
}
