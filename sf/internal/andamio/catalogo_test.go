package andamio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// EL CATÁLOGO QUE MANDA PARA EL ANDAMIO ES EL DEL PROYECTO, NO EL GLOBAL.
//
// `sf init` siembra un catálogo por proyecto (`<raiz>/.specforge/`) y desde ahí
// ÉSE es el que rige: `LeerPara` lo prefiere, y `sf model` escribe ahí. Pero
// `Instalar` leía siempre el global.
//
// Mientras `sf install` corría UNA sola vez y antes de `sf init`, la diferencia
// no se veía: no había catálogo de proyecto todavía. Se ve ahora, porque la
// razón de ser de `--harness=a,b,c` es volver a instalar sobre un proyecto que
// ya anda para dejar listos los otros arneses — y ahí el de proyecto ya existe.
//
// Con el global, esa reinstalación escribía CERO portamodelo y decía "0 modelos
// declarados" mientras el proyecto tenía los suyos: el "se muere el reinicio"
// de install-interactivo.md §6 no pasaba, y no pasaba en silencio.
func TestElAndamioSaleDelCatalogoDelProyectoYNoDelGlobal(t *testing.T) {
	enUnHomeDePrueba(t)
	parandoEn(t, "claude-code")

	// El global, vacío: es lo que deja un `sf install` recién corrido.
	if err := global.Semilla("claude-code").Guardar(); err != nil {
		t.Fatal(err)
	}

	// Y el del proyecto con un modelo declarado en opencode, que es lo que deja
	// un `sf model` después de `sf init`.
	raiz := t.TempDir()
	delProyecto := global.Semilla("claude-code")
	delProyecto.Harnesses["opencode"] = global.Catalogo{
		global.Construir: {{Alias: "rapido", ID: "oc/rapido", Via: global.Subagente}},
	}
	if err := delProyecto.GuardarEn(global.RutaDeProyecto(raiz)); err != nil {
		t.Fatal(err)
	}

	r, err := Instalar(raiz, Opciones{Harness: []string{"opencode"}})
	if err != nil {
		t.Fatal(err)
	}

	ruta := filepath.Join(raiz, ".opencode", "agents", "sf-rapido.md")
	if _, err := os.Stat(ruta); err != nil {
		t.Errorf("no escribió el portamodelo del catálogo del proyecto: %v", err)
	}
	if n := r.Modelos["opencode"]; n != 1 {
		t.Errorf("contó %d modelos y el proyecto declara 1", n)
	}
}

// Y el catálogo GLOBAL se sigue sembrando en el global, que es lo correcto: los
// modelos son de la máquina. Lo que cambia es de dónde sale el andamio, no dónde
// vive la semilla.
func TestLaSemillaSigueYendoAlCatalogoGlobal(t *testing.T) {
	enUnHomeDePrueba(t)
	parandoEn(t, "opencode")

	if _, err := Instalar(t.TempDir(), Opciones{}); err != nil {
		t.Fatal(err)
	}

	g, err := global.Leer()
	if err != nil {
		t.Fatalf("no sembró el catálogo global: %v", err)
	}
	if g.Harness != "opencode" {
		t.Errorf("el puntero global quedó en %q", g.Harness)
	}
}
