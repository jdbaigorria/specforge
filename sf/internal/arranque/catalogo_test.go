package arranque

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// El catálogo del proyecto se siembra del global: en el repo número diecisiete
// no hay que volver a contestar las mismas preguntas.
func TestInitSiembraElCatalogoDelProyecto(t *testing.T) {
	t.Setenv("SPECFORGE_HOME", t.TempDir())
	g := global.Semilla("opencode")
	if _, err := g.Declarar(global.Construir, global.Modelo{
		Alias: "medio", ID: "prov/medio", Via: global.Subagente,
	}); err != nil {
		t.Fatal(err)
	}
	if err := g.Guardar(); err != nil {
		t.Fatal(err)
	}

	raiz := t.TempDir()
	if _, err := Iniciar(raiz); err != nil {
		t.Fatal(err)
	}

	c, err := global.LeerPara(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if m, hay := c.Default(global.Construir); !hay || m.ID != "prov/medio" {
		t.Fatalf("no sembró el catálogo del proyecto: %+v", m)
	}
}

// Y VA GITIGNOREADO: contiene ids que sólo existen en esta máquina, así que
// versionarlo le rompería el clon a cualquier otro.
func TestElCatalogoDelProyectoVaGitignoreado(t *testing.T) {
	t.Setenv("SPECFORGE_HOME", t.TempDir())
	if err := global.Semilla("opencode").Guardar(); err != nil {
		t.Fatal(err)
	}

	raiz := t.TempDir()
	if err := os.WriteFile(filepath.Join(raiz, ".gitignore"), []byte("node_modules/\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Iniciar(raiz); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(filepath.Join(raiz, ".gitignore"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), global.Carpeta+"/") {
		t.Errorf("no lo gitignoreó:\n%s", b)
	}
	// Y no pisó lo que ya estaba: el .gitignore de un proyecto tiene un orden
	// que alguien eligió.
	if !strings.Contains(string(b), "node_modules/") {
		t.Errorf("pisó el .gitignore que había:\n%s", b)
	}
}

// Correr `sf init` dos veces no duplica la línea.
func TestNoDuplicaLaLineaDelGitignore(t *testing.T) {
	raiz := t.TempDir()
	for i := 0; i < 3; i++ {
		if err := gitignorar(raiz, global.Carpeta+"/"); err != nil {
			t.Fatal(err)
		}
	}
	b, _ := os.ReadFile(filepath.Join(raiz, ".gitignore"))
	if n := strings.Count(string(b), global.Carpeta+"/"); n != 1 {
		t.Errorf("la línea aparece %d veces:\n%s", n, b)
	}
}

// Un catálogo propio que ya existe NO se pisa: ahí viven las decisiones que
// Javier tomó para ESTE repo.
func TestNoPisaElCatalogoDelProyectoQueYaExiste(t *testing.T) {
	t.Setenv("SPECFORGE_HOME", t.TempDir())
	if err := global.Semilla("opencode").Guardar(); err != nil {
		t.Fatal(err)
	}

	raiz := t.TempDir()
	propio := global.Semilla("commandcode")
	if _, err := propio.Declarar(global.Razonar, global.Modelo{
		Alias: "el-del-repo", ID: "prov/repo", Via: global.Subagente,
	}); err != nil {
		t.Fatal(err)
	}
	if err := propio.GuardarEn(global.RutaDeProyecto(raiz)); err != nil {
		t.Fatal(err)
	}

	if _, err := Iniciar(raiz); err != nil {
		t.Fatal(err)
	}
	c, _ := global.LeerPara(raiz)
	if m, hay := c.Default(global.Razonar); !hay || m.Alias != "el-del-repo" {
		t.Fatalf("pisó el catálogo del repo: %+v", m)
	}
}

// Sin catálogo global, `sf init` sigue funcionando: no tener modelos no puede
// impedir iniciar un proyecto. El primer `sf next` va a parar y pedirlos igual.
func TestSinCatalogoGlobalInitAndaIgual(t *testing.T) {
	t.Setenv("SPECFORGE_HOME", t.TempDir())
	raiz := t.TempDir()

	if _, err := Iniciar(raiz); err != nil {
		t.Fatalf("init falló sin catálogo global: %v", err)
	}
	if _, err := os.Stat(filepath.Join(raiz, ".docs")); err != nil {
		t.Error("no armó el andamio")
	}
}
