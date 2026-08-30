package global

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// enUnHomeFalso apunta SPECFORGE_HOME a un temporal, para que ningún test
// escriba en el home de verdad de quien lo corre.
func enUnHomeFalso(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("SPECFORGE_HOME", dir)
	return dir
}

// varsDeHarness son todas las que mira DetectarHarness.
var varsDeHarness = VarsDeHarness

// parandoEn dice en qué harness está corriendo el test.
//
// Hace falta desde que el puntero se DETECTA: esta suite corre adentro de algún
// arnés —hoy Claude Code, mañana otro—, así que sin limpiar las variables cada
// test heredaría el arnés de quien lo lanzó y probaría cualquier cosa. Que haga
// falta es, de hecho, la prueba de que el mecanismo funciona.
func parandoEn(t *testing.T, harness string) {
	t.Helper()
	for _, v := range varsDeHarness {
		t.Setenv(v, "")
	}
	if harness != "" {
		t.Setenv(VarHarness, harness)
	}
}

func TestSinInstalarNoHayCatalogo(t *testing.T) {
	enUnHomeFalso(t)
	if _, err := Leer(); !errors.Is(err, ErrNoHay) {
		t.Fatalf("esperaba ErrNoHay, vino %v", err)
	}
}

// La semilla nace VACÍA, y es el corazón de H2: ningún nombre de modelo lo pone
// sf. Si esto se rompe, volvió la siembra y con ella el "siempre opus".
func TestLaSemillaNaceVaciaYNoNombraNingunModelo(t *testing.T) {
	enUnHomeFalso(t)
	c := Semilla("opencode")
	if err := c.Guardar(); err != nil {
		t.Fatal(err)
	}

	v, err := Leer()
	if err != nil {
		t.Fatal(err)
	}
	if v.Harness != "opencode" {
		t.Errorf("harness = %q", v.Harness)
	}
	if len(v.Alias()) != 0 {
		t.Errorf("la semilla declaró %d modelos y no tenía que declarar ninguno", len(v.Alias()))
	}
	for _, p := range PerfilesConocidos() {
		if _, hay := v.Default(p); hay {
			t.Errorf("el perfil %q vino con un default y la semilla es vacía", p)
		}
	}
}

func TestDeclararYResolverPorPerfilYPorAlias(t *testing.T) {
	c := Semilla("opencode")
	if _, err := c.Declarar(Construir, Modelo{Alias: "sonnet", ID: "anthropic/x", Via: Subagente}); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Declarar(Construir, Modelo{Alias: "laguna", ID: "laguna/s2.1", Via: Subagente}); err != nil {
		t.Fatal(err)
	}

	// El PRIMERO de la lista es el default del perfil: el orden es la preferencia.
	m, hay := c.Resolver(Construir)
	if !hay || m.Alias != "sonnet" {
		t.Errorf("el default de construir = %+v, esperaba sonnet", m)
	}
	// Y un alias resuelve al suyo, no al default.
	m, hay = c.Resolver("laguna")
	if !hay || m.ID != "laguna/s2.1" {
		t.Errorf("laguna = %+v", m)
	}
	if _, hay := c.Resolver("noexiste"); hay {
		t.Error("resolvió un alias que nadie declaró")
	}
}

// El mismo alias en dos harness da ids distintos. Es la prueba de que el plan
// sobrevive al cambio de harness, que es todo el punto del alias.
func TestElMismoAliasDaIdsDistintosSegunElHarness(t *testing.T) {
	parandoEn(t, "opencode")
	c := Semilla("opencode")
	if _, err := c.Declarar(Razonar, Modelo{Alias: "grande", ID: "prov/grande-4", Via: Subagente}); err != nil {
		t.Fatal(err)
	}

	// Me mudo de arnés SIN tocar el archivo: sólo cambia dónde estoy parado.
	parandoEn(t, "claude-code")
	if _, err := c.Declarar(Razonar, Modelo{Alias: "grande", ID: "el-alias-nativo", Via: Subagente}); err != nil {
		t.Fatal(err)
	}
	if m, _ := c.Resolver("grande"); m.ID != "el-alias-nativo" {
		t.Errorf("en claude-code grande = %q", m.ID)
	}

	parandoEn(t, "opencode")
	if m, _ := c.Resolver("grande"); m.ID != "prov/grande-4" {
		t.Errorf("en opencode grande = %q", m.ID)
	}
}

// EL TEST DE SECCIONAR. Javier planifica en un arnés, cierra, abre otro en el
// mismo repo. El estado está en disco así que eso ya funcionaba; lo que faltaba
// era que sf resolviera los modelos DEL ARNÉS DONDE ESTÁ PARADO.
func TestAbrirOtroArnesEnElMismoRepoCambiaLosModelos(t *testing.T) {
	enUnHomeFalso(t)

	parandoEn(t, "claude-code")
	c := Semilla("claude-code")
	if _, err := c.Declarar(Construir, Modelo{Alias: "el-de-siempre", ID: "nativo", Via: Subagente}); err != nil {
		t.Fatal(err)
	}
	parandoEn(t, "commandcode")
	if _, err := c.Declarar(Construir, Modelo{Alias: "gratis", ID: "prov/gratis", Via: Subagente}); err != nil {
		t.Fatal(err)
	}
	if err := c.Guardar(); err != nil {
		t.Fatal(err)
	}

	// Mismo archivo, leído desde cada arnés.
	parandoEn(t, "claude-code")
	v, _ := Leer()
	if m, hay := v.Default(Construir); !hay || m.ID != "nativo" {
		t.Errorf("en claude-code el default es %+v", m)
	}
	parandoEn(t, "commandcode")
	w, _ := Leer()
	if m, hay := w.Default(Construir); !hay || m.ID != "prov/gratis" {
		t.Errorf("en commandcode el default es %+v", m)
	}
}

// Y si el arnés donde estás parado NO tiene bloque, la respuesta correcta es
// "no hay nada declarado acá" — no los ids del otro, que en éste no existen.
func TestUnArnesSinBloqueNoHeredaLosModelosDelOtro(t *testing.T) {
	parandoEn(t, "claude-code")
	c := Semilla("claude-code")
	if _, err := c.Declarar(Construir, Modelo{Alias: "x", ID: "nativo", Via: Subagente}); err != nil {
		t.Fatal(err)
	}

	parandoEn(t, "opencode")
	if m, hay := c.Default(Construir); hay {
		t.Errorf("opencode heredó %+v, y ese id no existe ahí", m)
	}
}

// El orden de la certeza: la variable explícita le gana a la detección, y la
// detección le gana a lo que quedó escrito.
func TestElOrdenDeLaCerteza(t *testing.T) {
	c := &Config{Harness: "el-escrito"}

	parandoEn(t, "")
	if h := c.EnUso(); h != "el-escrito" {
		t.Errorf("sin nada = %q, quería el escrito", h)
	}

	parandoEn(t, "")
	t.Setenv("OPENCODE", "1")
	if h := c.EnUso(); h != "opencode" {
		t.Errorf("con detección = %q, quería opencode", h)
	}

	t.Setenv(VarHarness, "commandcode")
	if h := c.EnUso(); h != "commandcode" {
		t.Errorf("con la variable explícita = %q", h)
	}
}

// Cambiar de harness NO puede destruir el bloque del anterior. Es el bug que
// destapó la prueba de escritorio.
func TestCambiarDeHarnessNoDestruyeElBloqueAnterior(t *testing.T) {
	dir := enUnHomeFalso(t)
	c := Semilla("opencode")
	if _, err := c.Declarar(Construir, Modelo{Alias: "laguna", ID: "laguna/s2.1", Via: Subagente}); err != nil {
		t.Fatal(err)
	}
	if err := c.Guardar(); err != nil {
		t.Fatal(err)
	}

	v, _ := Leer()
	v.Harness = "commandcode" // "me mudé de harness"
	if err := v.GuardarEn(dir); err != nil {
		t.Fatal(err)
	}

	w, _ := Leer()
	w.Harness = "opencode"
	if m, hay := w.Resolver("laguna"); !hay || m.ID != "laguna/s2.1" {
		t.Fatal("se perdió el bloque de opencode al mudarse de harness")
	}
}

func TestUnAliasNoPuedeLlamarseComoUnPerfil(t *testing.T) {
	c := Semilla("opencode")
	if _, err := c.Declarar(Razonar, Modelo{Alias: Razonar, ID: "x", Via: Subagente}); err == nil {
		t.Fatal("dejó declarar un alias que se llama igual que un perfil")
	}
	if _, err := c.Declarar("inventado", Modelo{Alias: "x", ID: "y", Via: Subagente}); err == nil {
		t.Fatal("dejó declarar en un perfil que no existe")
	}
	if _, err := c.Declarar(Razonar, Modelo{ID: "y", Via: Subagente}); err == nil {
		t.Fatal("dejó declarar un modelo sin alias")
	}
}

// Redeclarar el mismo alias ACTUALIZA, no duplica, y avisa que no es nuevo —
// que es lo que `sf model` usa para decidir si hay que reiniciar el harness.
func TestRedeclararActualizaYNoEsNuevo(t *testing.T) {
	c := Semilla("opencode")
	nuevo, _ := c.Declarar(Construir, Modelo{Alias: "laguna", ID: "v1", Via: Subagente})
	if !nuevo {
		t.Error("la primera vez tenía que ser nuevo")
	}
	nuevo, _ = c.Declarar(Construir, Modelo{Alias: "laguna", ID: "v2", Via: Subagente})
	if nuevo {
		t.Error("redeclarar no es nuevo")
	}
	if len(c.Activo()[Construir]) != 1 {
		t.Errorf("quedaron %d entradas, esperaba 1", len(c.Activo()[Construir]))
	}
	if m, _ := c.Resolver("laguna"); m.ID != "v2" {
		t.Errorf("no actualizó: %q", m.ID)
	}
}

func TestDeclararUnModeloSueltoDeConsola(t *testing.T) {
	c := Semilla("opencode")
	c.DeclararSuelto("deepseek", Modelo{ID: "deepseek-reasoner", Via: Consola, Comando: "deepseek exec"})

	m, hay := c.Resolver("deepseek")
	if !hay || m.Via != Consola || m.Comando != "deepseek exec" {
		t.Fatalf("deepseek = %+v", m)
	}
	// Un suelto NO compite en ningún perfil.
	if len(c.Alias()) != 0 {
		t.Errorf("el suelto se coló en un perfil: %+v", c.Alias())
	}
}

// La migración del formato viejo se hace AL LEER, sin que nadie corra nada.
func TestUnCatalogoViejoSeMigraSolo(t *testing.T) {
	dir := enUnHomeFalso(t)
	viejo := "harness: claude-code\nmodelos:\n" +
		"  opus: {via: subagente}\n" +
		"  sonnet: {via: subagente}\n" +
		"  haiku: {via: subagente}\n" +
		"  deepseek: {via: consola, comando: deepseek exec}\n"
	if err := os.WriteFile(filepath.Join(dir, Archivo), []byte(viejo), 0o644); err != nil {
		t.Fatal(err)
	}

	c, err := Leer()
	if err != nil {
		t.Fatal(err)
	}
	for perfil, esperado := range map[string]string{Razonar: "opus", Construir: "sonnet", Mecanico: "haiku"} {
		m, hay := c.Default(perfil)
		if !hay || m.Alias != esperado || m.ID != esperado {
			t.Errorf("%s = %+v, esperaba alias e id %q", perfil, m, esperado)
		}
	}
	// El suelto sobrevive donde estaba: no es de ningún perfil.
	if m, hay := c.Resolver("deepseek"); !hay || m.Comando != "deepseek exec" {
		t.Errorf("se perdió el modelo suelto al migrar: %+v", m)
	}
}

// El catálogo del PROYECTO le gana al global. Es la elección de Javier: cada
// repo puede querer modelos distintos.
func TestElCatalogoDelProyectoLeGanaAlGlobal(t *testing.T) {
	enUnHomeFalso(t)
	g := Semilla("opencode")
	if _, err := g.Declarar(Construir, Modelo{Alias: "global", ID: "id/global", Via: Subagente}); err != nil {
		t.Fatal(err)
	}
	if err := g.Guardar(); err != nil {
		t.Fatal(err)
	}

	raiz := t.TempDir()
	// Sin catálogo propio, el proyecto usa el global.
	c, err := LeerPara(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if m, _ := c.Default(Construir); m.Alias != "global" {
		t.Fatalf("sin catálogo propio tenía que caer al global, vino %+v", m)
	}

	p := Semilla("opencode")
	if _, err := p.Declarar(Construir, Modelo{Alias: "propio", ID: "id/propio", Via: Subagente}); err != nil {
		t.Fatal(err)
	}
	if err := p.GuardarEn(RutaDeProyecto(raiz)); err != nil {
		t.Fatal(err)
	}

	c, err = LeerPara(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if m, _ := c.Default(Construir); m.Alias != "propio" {
		t.Fatalf("el del proyecto tenía que ganar, vino %+v", m)
	}
	// Y no se mergean: el global no se cuela.
	if _, hay := c.Resolver("global"); hay {
		t.Error("se mergearon los dos catálogos, y no se tienen que mergear")
	}
}

func TestGuardarCreaLaCarpeta(t *testing.T) {
	dir := enUnHomeFalso(t)
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
	if err := Semilla("claude-code").Guardar(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, Archivo)); err != nil {
		t.Fatal(err)
	}
}

// Un Config nil no puede hacer panic: es la lección de A9.
func TestUnConfigNilNoRompe(t *testing.T) {
	var c *Config
	if _, hay := c.Resolver(Razonar); hay {
		t.Error("un catálogo nil no puede resolver nada")
	}
	if len(c.Alias()) != 0 {
		t.Error("un catálogo nil no tiene alias")
	}
}

func TestDetectarHarness(t *testing.T) {
	for _, c := range []struct{ env, esperado string }{
		{"CLAUDECODE", "claude-code"},
		{"OPENCODE", "opencode"},
		{"COMMANDCODE_SCRATCHPAD", "commandcode"},
	} {
		t.Run(c.esperado, func(t *testing.T) {
			for _, v := range VarsDeHarness {
				t.Setenv(v, "")
			}
			t.Setenv(c.env, "1")
			if h := DetectarHarness(); h != c.esperado {
				t.Errorf("= %q, esperaba %q", h, c.esperado)
			}
		})
	}
}

// El invariante de H2: ni un nombre de modelo concreto en el paquete.
func TestElPaqueteNoNombraNingunModelo(t *testing.T) {
	b, err := os.ReadFile("global.go")
	if err != nil {
		t.Fatal(err)
	}
	// El encabezado del paquete SÍ los menciona: está explicando por qué ya no
	// se siembran. Lo que no puede haber es un nombre en el código.
	codigo := string(b)
	if i := strings.Index(codigo, "\npackage global"); i > 0 {
		codigo = codigo[i:]
	}
	for _, prohibido := range []string{"opus", "sonnet", "haiku", "gpt-", "gemini", "grok"} {
		if strings.Contains(codigo, prohibido) {
			t.Errorf("el código nombra %q — la siembra volvió", prohibido)
		}
	}
}

// UN ARNÉS ADENTRO DE OTRO. Medido: un `opencode run` lanzado desde Claude Code
// trae las variables de LOS DOS. Sin este orden, sf contestaba "claude-code"
// estando parado en opencode, y resolvía ids que ahí no existen.
func TestElArnesDeAdentroLeGanaAlDeAfuera(t *testing.T) {
	for _, c := range []struct{ v, quiero string }{
		{"OPENCODE", "opencode"},
		{"COMMANDCODE_SCRATCHPAD", "commandcode"},
	} {
		t.Run(c.quiero, func(t *testing.T) {
			parandoEn(t, "")
			t.Setenv("CLAUDECODE", "1") // el de AFUERA
			t.Setenv(c.v, "1")          // el de ADENTRO
			if h := DetectarHarness(); h != c.quiero {
				t.Errorf("= %q, quería %q: el de adentro es el que corre", h, c.quiero)
			}
		})
	}
}

// Y la variable explícita le gana a cualquier heurística: es lo que `sf lanzar`
// le va a poner a cada hijo, donde no hay nada que adivinar.
func TestLaVariableExplicitaLeGanaALaHeuristica(t *testing.T) {
	parandoEn(t, "")
	t.Setenv("CLAUDECODE", "1")
	t.Setenv("OPENCODE", "1")
	t.Setenv(VarHarness, "commandcode")
	if h := DetectarHarness(); h != "commandcode" {
		t.Errorf("= %q, quería la explícita", h)
	}
}
