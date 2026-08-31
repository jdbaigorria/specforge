package andamio

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// conUnModelo siembra el catálogo con un alias declarado en cada arnés pedido.
//
// Hace falta porque el portamodelo es lo que se mira para saber si el andamio de
// un arnés se armó de verdad, y sin un alias declarado no hay ninguno que
// escribir.
func conUnModelo(t *testing.T, harnesses ...string) {
	t.Helper()
	g := global.Semilla(harnesses[0])
	if g.Harnesses == nil {
		g.Harnesses = map[string]global.Catalogo{}
	}
	// Se arma el catálogo a mano y no con Declarar() porque Declarar escribe en
	// el arnés EN USO, y acá hace falta sembrar los tres a la vez.
	for _, h := range harnesses {
		g.Harnesses[h] = global.Catalogo{
			global.Construir: {{Alias: "barato", ID: h + "/barato", Via: global.Subagente}},
		}
	}
	if err := g.Guardar(); err != nil {
		t.Fatal(err)
	}
}

// LA PRUEBA QUE MATA EL "REINICIÁ TU HARNESS".
//
// Hasta ahora `sf install` armaba el andamio de UNO, así que dejar dos arneses
// listos pedía instalar en cada uno — y como las definiciones de agente se leen
// al arrancar (sonda 1, 2026-08-29), en el medio había un cerrá-y-abrí.
//
// Declarando los tres de una, los portamodelo de opencode y de Command Code
// quedan escritos ANTES de que abras esos arneses. Cuando los abrís, ya están
// cargados: el paso no se reduce, DESAPARECE (install-interactivo.md §6).
func TestInstalarParaVariosArmaElAndamioDeCadaUno(t *testing.T) {
	enUnHomeDePrueba(t)
	parandoEn(t, "claude-code")
	conUnModelo(t, "claude-code", "opencode", "commandcode")

	raiz := t.TempDir()
	r, err := Instalar(raiz, Opciones{Harness: []string{"claude-code", "opencode", "commandcode"}})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(r.Para, []string{"claude-code", "opencode", "commandcode"}) {
		t.Errorf("no dijo para cuáles instaló: %v", r.Para)
	}

	// Los permisos, que es lo que hace que el bucle no se trabe en el primer
	// subagente: uno por arnés, y cada uno en su archivo.
	for _, ruta := range []string{
		filepath.Join(".claude", "settings.json"),
		"opencode.json", // en la raíz, no adentro de .opencode/
		filepath.Join(".commandcode", "settings.json"),
	} {
		if _, err := os.Stat(filepath.Join(raiz, ruta)); err != nil {
			t.Errorf("no escribió %s: %v", ruta, err)
		}
	}

	// Y los portamodelo, en los dos que los necesitan.
	for _, ruta := range []string{
		filepath.Join(".opencode", "agents", "sf-barato.md"),
		filepath.Join(".commandcode", "agents", "sf-barato.md"),
	} {
		if _, err := os.Stat(filepath.Join(raiz, ruta)); err != nil {
			t.Errorf("no escribió %s: %v", ruta, err)
		}
	}
}

// El puntero es UNO SOLO aunque se instalen tres, y es el fallback de EnUso():
// para qué arnés resolver modelos cuando no hay ni variable ni detección — una
// terminal pelada, un cron. Con varios instalados, "el último que instalaste"
// deja de significar algo, así que hay que elegir a propósito.
//
// Gana donde estás parado: es el único de los tres sobre el que hay un hecho.
func TestElPunteroQuedaEnDondeEstasParado(t *testing.T) {
	enUnHomeDePrueba(t)
	parandoEn(t, "opencode")
	conUnModelo(t, "claude-code", "opencode")

	r, err := Instalar(t.TempDir(), Opciones{Harness: []string{"claude-code", "opencode"}})
	if err != nil {
		t.Fatal(err)
	}
	if r.Puntero != "opencode" {
		t.Errorf("el puntero quedó en %q y estabas parado en opencode", r.Puntero)
	}

	g, err := global.Leer()
	if err != nil {
		t.Fatal(err)
	}
	if g.Harness != "opencode" {
		t.Errorf("en el catálogo quedó %q", g.Harness)
	}
}

// Y si donde estás parado NO está entre los que pediste, cae al primero de la
// lista. No se inventa nada: el primero es el que Javier nombró primero.
func TestElPunteroCaeAlPrimeroSiNoEstasEnNingunoDeLosPedidos(t *testing.T) {
	enUnHomeDePrueba(t)
	parandoEn(t, "claude-code")
	conUnModelo(t, "opencode", "commandcode")

	r, err := Instalar(t.TempDir(), Opciones{Harness: []string{"opencode", "commandcode"}})
	if err != nil {
		t.Fatal(err)
	}
	if r.Puntero != "opencode" {
		t.Errorf("esperaba el primero de la lista (opencode), quedó %q", r.Puntero)
	}
}

// LA PRUEBA DE NO REGRESIÓN, Y ES LA QUE NO PUEDE FALLAR NUNCA.
//
// `sf install` pelado es lo que corre el orquestador —lo nombra en "Empezar de
// cero"— y lo que corre `install.sh`. Tiene que seguir haciendo exactamente lo
// de antes: armar el andamio de UNO, el de donde estás.
//
// La detección de instalados NO cambia esto: alimenta el MENÚ de la pregunta
// interactiva, que es otra cosa. Que en esta máquina estén los tres no puede
// hacer que un `sf install` sin flags escriba .opencode/ en el repo de alguien.
func TestSinFlagsSigueArmandoUnoSolo(t *testing.T) {
	enUnHomeDePrueba(t)
	parandoEn(t, "claude-code")
	conUnModelo(t, "claude-code", "opencode")

	raiz := t.TempDir()
	r, err := Instalar(raiz, Opciones{})
	if err != nil {
		t.Fatal(err)
	}

	if !slices.Equal(r.Para, []string{"claude-code"}) {
		t.Errorf("sin flags instaló para %v, tenía que ser sólo claude-code", r.Para)
	}
	if _, err := os.Stat(filepath.Join(raiz, ".opencode")); err == nil {
		t.Error("armó el andamio de opencode sin que nadie se lo pidiera")
	}
}

// Un arnés que sf no sabe preparar SE ACEPTA —eso es de antes y es
// deliberado— pero deja de hacerlo en silencio.
//
// `--harness=codex` existe para un arnés que todavía no está soportado acá: sf
// transporta lo que Javier declara y no le pone una lista blanca (R3). Lo que
// estaba mal era que `escribirPermisos` y `escribirPortamodelos` no hacen NADA
// con un nombre que no está en sus mapas, así que la instalación terminaba con
// un ✓ y el arnés sin permisos. El bucle se traba después, en el primer
// subagente, que es el peor momento para enterarse.
func TestUnArnesQueSfNoSabePrepararSeAvisaYNoSeCalla(t *testing.T) {
	enUnHomeDePrueba(t)
	parandoEn(t, "claude-code")

	raiz := t.TempDir()
	r, err := Instalar(raiz, Opciones{Harness: []string{"claude-code", "codex"}})
	if err != nil {
		t.Fatalf("rechazarlo sería romper `--harness=codex`, que es deliberado: %v", err)
	}

	if !strings.Contains(strings.Join(r.Salteados, "\n"), "codex") {
		t.Errorf("no avisó que no sabe prepararlo: %v", r.Salteados)
	}
	// Y el que sí conoce se preparó igual: uno raro en la lista no arrastra al resto.
	if _, err := os.Stat(filepath.Join(raiz, ".claude", "settings.json")); err != nil {
		t.Error("un arnés desconocido en la lista frenó al que sí conocía")
	}
}
