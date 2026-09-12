package maquina

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
)

// ────────────────────────────────────────────────────────────────────────────
// Andamio: un proyecto con git de verdad
// ────────────────────────────────────────────────────────────────────────────
//
// `sf lote start` crea branches y corre comandos: sin un repo real no se prueba
// nada. Los tests de acá arman uno con un commit inicial.

func (p *proyecto) conGit() *proyecto {
	p.t.Helper()
	for _, args := range [][]string{
		{"init", "-b", "main"},
		{"config", "user.email", "t@e"},
		{"config", "user.name", "T"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = p.raiz
		if out, err := cmd.CombinedOutput(); err != nil {
			p.t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	p.conArchivoConTexto("README.md", "# proyecto\n")
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-m", "base"}} {
		cmd := exec.Command("git", args...)
		cmd.Dir = p.raiz
		if out, err := cmd.CombinedOutput(); err != nil {
			p.t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	return p
}

// rojoReal y verdeReal son los dos `test_cmd` que usan los tests de este archivo.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ EL ROJO IMPRIME, Y NO ES COSMÉTICA
// ────────────────────────────────────────────────────────────────────────────
//
// Estos fixtures eran `exit 1` pelado: fallaban EN SILENCIO. Servía mientras la
// compuerta del rojo era sólo un exit code, y dejó de servir cuando `sf lote
// start` empezó a exigir que la salida nombre alguno de los tests del lote.
//
// Y lo que rompió es justamente lo que había que arreglar: **un `exit 1` mudo
// es indistinguible de una suite que ya venía rota**. Ningún runner de verdad
// se comporta así — go, pytest, jest y cargo nombran todos lo que falla —, así
// que el fixture que se cayó era el que no se parecía a la realidad.
//
// `TestUno` es uno de los dos tests que el plan declara más abajo.
const (
	rojoReal  = "echo FAIL TestUno; exit 1"
	verdeReal = "exit 0"
)

// listoParaImplementar deja f-1 en implementar, con constitución, plan y tests.
//
// `testCmd` es lo que va a decidir el rojo o el verde: los tests le pasan
// `rojoReal` o `verdeReal`, porque lo que se prueba acá es la COMPUERTA, no el
// runner.
func (p *proyecto) listoParaImplementar(testCmd string) *proyecto {
	p.productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Implementar}

	p.conArchivoConTexto(docs.Constitucion,
		"---\nlenguaje: go\ntest_cmd: "+testCmd+"\ngit:\n"+
			"  branch_por_feature: true\n  patron_branch: \"feat/{feature-id}-{slug}\"\n---\n# reglas\n")

	p.conArchivoConTexto(filepath.Join(".docs", "features", "f-1-nucleo", "tareas.json"),
		`{"tareas":[{"id":"t-1","lote":1,"satisface":["us-1/CA-1"],
		  "tests":["a_test.go::TestUno","a_test.go::TestDos"]}]}`)

	p.conArchivoConTexto("a_test.go", "package a\n\nfunc TestUno() {}\nfunc TestDos() {}\n")
	return p.conGit()
}

// corrergit corre un comando de git en el proyecto de prueba y falla el test si
// se rompe. Es para ARMAR el escenario; lo que se prueba es lo que hace sf.
func corrergit(t *testing.T, raiz string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = raiz
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// gitLog devuelve los mensajes de commit, uno por línea.
func gitLog(t *testing.T, raiz string) string {
	t.Helper()
	cmd := exec.Command("git", "log", "--pretty=%s")
	cmd.Dir = raiz
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func branchActual(t *testing.T, raiz string) string {
	t.Helper()
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = raiz
	out, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

// ────────────────────────────────────────────────────────────────────────────
// El camino feliz: las tres compuertas pasan
// ────────────────────────────────────────────────────────────────────────────

func TestLoteStartCreaLaBranchYConfirmaElRojo(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal) // la suite falla: hay rojo

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if !ef.Pasa() {
		t.Fatalf("no pasó: %v", ef.Fallas)
	}

	// ① la branch se CREA, no se verifica: el dolor #4 deja de existir.
	if b := branchActual(t, p.raiz); b != "feat/f-1-nucleo" {
		t.Errorf("branch = %q, quería feat/f-1-nucleo", b)
	}

	// ② y ③ quedan anotados en el estado.
	l, _ := p.e.Features["f-1"].LoteActual()
	if !l.Rojo {
		t.Error("no anotó el rojo")
	}
	if l.HashTests == "" {
		t.Error("no guardó el hash de los archivos de test")
	}
}

// Los lotes se siembran desde tareas.json la primera vez: el estado sólo guarda
// por cuál va, no cuántos hay.
func TestLoteStartSiembraLosLotesDesdeElPlan(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	if len(p.e.Features["f-1"].Lotes) != 0 {
		t.Fatal("el andamio ya traía lotes")
	}

	if ef := EmpezarLote(p.raiz, p.e, p.r); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}
	if len(p.e.Features["f-1"].Lotes) != 1 {
		t.Errorf("sembró %d lotes, quería 1", len(p.e.Features["f-1"].Lotes))
	}
}

// ────────────────────────────────────────────────────────────────────────────
// La joya: un test que ya pasa es un test de mentira
// ────────────────────────────────────────────────────────────────────────────

func TestLoteStartNoDejaEmpezarSiLaSuiteYaPasa(t *testing.T) {
	p := nuevo(t).listoParaImplementar(verdeReal) // ya está verde

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("dejó empezar con la suite en verde")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "test de mentira") {
		t.Errorf("no explicó por qué: %v", ef.Fallas)
	}

	// Y no anota nada: el rojo no ocurrió.
	if l, hay := p.e.Features["f-1"].LoteActual(); hay && l.Rojo {
		t.Error("anotó el rojo igual")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El dolor #8: los tests planificados tienen que existir
// ────────────────────────────────────────────────────────────────────────────

func TestLoteStartExigeQueLosTestsPlanificadosExistan(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	// El archivo existe pero le falta TestDos: es exactamente "escribí tests"
	// contra "escribí LOS tests que el plan pedía".
	p.conArchivoConTexto("a_test.go", "package a\n\nfunc TestUno() {}\n")

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("dejó empezar con un test planificado sin escribir")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "TestDos") {
		t.Errorf("no dijo cuál falta: %v", ef.Fallas)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Errores de uso
// ────────────────────────────────────────────────────────────────────────────

func TestLoteStartSoloAplicaAImplementar(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	p.e.Features["f-1"].Estado = estado.Planificacion

	if ef := EmpezarLote(p.raiz, p.e, p.r); ef.Pasa() {
		t.Error("dejó correr lote start en planificacion")
	}
}

// Correrlo dos veces avisa en vez de volver a exigir el rojo — que ya no
// existiría, porque el código empezó a escribirse.
func TestLoteStartDosVecesAvisa(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	if ef := EmpezarLote(p.raiz, p.e, p.r); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("dejó correrlo dos veces sobre el mismo lote")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "ya está confirmado") {
		t.Errorf("no explicó por qué: %v", ef.Fallas)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El cierre del lote: el verde y el hash
// ────────────────────────────────────────────────────────────────────────────

// No se commitea en rojo. Es la mitad obvia de la compuerta.
func TestDoneNoCommiteaEnRojo(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	if ef := EmpezarLote(p.raiz, p.e, p.r); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}

	c := Terminar(p.raiz, p.e, p.r, "feat: algo")
	if c.Pasa() {
		t.Fatal("commiteó con la suite en rojo")
	}
	if !strings.Contains(strings.Join(c.Fallas, " "), "todavía fallan") {
		t.Errorf("no dijo que siguen fallando: %v", c.Fallas)
	}
}

// EL AGUJERO ASTUTO: la suite pasa, pero lo que cambió fue el test.
//
// Un subagente que no logra implementar puede ablandar el assert y llegar a
// verde. sf vio el rojo y ve el verde; sin el hash, esto pasa sin que nadie se
// entere.
func TestDoneAtrapaElTestAblandadoEntreElRojoYElVerde(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	if ef := EmpezarLote(p.raiz, p.e, p.r); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}

	// Ahora "arregla" aflojando el test, y de paso la suite pasa.
	p.conArchivoConTexto(docs.Constitucion,
		"---\nlenguaje: go\ntest_cmd: exit 0\ngit:\n"+
			"  branch_por_feature: true\n  patron_branch: \"feat/{feature-id}-{slug}\"\n---\n# reglas\n")
	p.conArchivoConTexto("a_test.go", "package a\n\nfunc TestUno() {}\nfunc TestDos() {} // sin assert\n")

	c := Terminar(p.raiz, p.e, p.r, "feat: listo")
	if c.Pasa() {
		t.Fatal("dejó pasar un test aflojado entre el rojo y el verde")
	}
	if !strings.Contains(strings.Join(c.Fallas, " "), "CAMBIARON") {
		t.Errorf("no dijo que los tests cambiaron: %v", c.Fallas)
	}
}

// Con la suite en verde y los tests intactos, el lote cierra y sf commitea.
func TestDoneCierraElLoteConVerdeYTestsIntactos(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	if ef := EmpezarLote(p.raiz, p.e, p.r); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}

	// El código aparece y la suite pasa; los tests no se tocan.
	p.conArchivoConTexto(docs.Constitucion,
		"---\nlenguaje: go\ntest_cmd: exit 0\ngit:\n"+
			"  branch_por_feature: true\n  patron_branch: \"feat/{feature-id}-{slug}\"\n---\n# reglas\n")
	p.conArchivoConTexto("codigo.go", "package a\n\nfunc Hace() {}\n")

	c := Terminar(p.raiz, p.e, p.r, "feat: hace algo")
	if !c.Pasa() {
		t.Fatalf("no cerró: %v", c.Fallas)
	}
	l := p.e.Features["f-1"].Lotes[0]
	if l.Commit == nil {
		t.Fatal("no guardó el commit")
	}
	if p.e.Features["f-1"].IntentosFallidos != 0 {
		t.Error("no reseteó el contador al cerrar el lote")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El camino corto: sin plan, la compuerta afloja pero no se cae
// ────────────────────────────────────────────────────────────────────────────

// Un bug no pasa por planificación, así que no hay tareas.json y no hay lista
// de tests contra la cual exigir el rojo. La compuerta pasa de "¿fallan LOS 6
// planificados?" a "¿falla AL MENOS UNO?" — sigue siendo un hecho y un exit
// code, y sigue impidiendo lo que importa: dar por arreglado algo que nunca se
// vio romper.
func TestLoteStartSinPlanAflojaPeroSigueExigiendoRojo(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	// Se saca el plan: es el caso del bug.
	if err := os.Remove(filepath.Join(p.raiz, ".docs", "features", "f-1-nucleo", "tareas.json")); err != nil {
		t.Fatal(err)
	}

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if !ef.Pasa() {
		t.Fatalf("sin plan no dejó empezar: %v", ef.Fallas)
	}
	if !strings.Contains(ef.Mensaje, "camino corto") {
		t.Errorf("no dijo que iba sin plan: %q", ef.Mensaje)
	}
	// Siembra un lote único: no hay plan que diga cuántos son.
	if n := len(p.e.Features["f-1"].Lotes); n != 1 {
		t.Errorf("sembró %d lotes, quería 1", n)
	}
}

// Si la suite pasa entera, el bug no está reproducido. Es la misma regla que el
// "test de mentira", dicha para el caso del arreglo chico.
func TestLoteStartSinPlanExigeQueElBugEsteReproducido(t *testing.T) {
	p := nuevo(t).listoParaImplementar(verdeReal)
	if err := os.Remove(filepath.Join(p.raiz, ".docs", "features", "f-1-nucleo", "tareas.json")); err != nil {
		t.Fatal(err)
	}

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("dejó empezar con la suite entera en verde")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "no está reproducido") {
		t.Errorf("no explicó por qué: %v", ef.Fallas)
	}
}

// Sin lotes todavía no arrancó nada. Confundirlo con "todos commiteados"
// mandaba a cerrar una feature en la que no se escribió una línea — y en el
// camino corto pasaba siempre, porque ahí no hay planificación que los siembre.
func TestSinLotesMandaALoteStartYNoACerrar(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Implementar}

	i := p.next()
	if !slices.Contains(i.Sugerido, "sf lote start") {
		t.Errorf("sugirió %v, quería sf lote start", i.Sugerido)
	}
	if strings.Contains(i.Mensaje, "commiteados") {
		t.Errorf("dijo que estaban todos commiteados sin ningún lote: %q", i.Mensaje)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El camino corto, la otra mitad: el guard tiene que mirar el estado EFECTIVO
// ────────────────────────────────────────────────────────────────────────────

// Un bug entra a `implementar` sin pasar por `planificacion`, y el estado.json
// todavía dice "planificacion" — el que lo mueve es `sf done`, que en el camino
// corto va DESPUÉS de esto.
//
// Comparar contra el estado crudo dejaba a `sf lote start` contradiciendo a
// `sf next` sobre la misma feature: next decía "implementar, corré lote start" y
// lote start contestaba "esto está en planificacion". El bug quedaba sin ninguna
// forma de escribir código.
func TestLoteStartAceptaUnBugEnPlanificacion(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	p.conArchivoConTexto(docs.Historia("us-1"), "---\ntipo: bug\nid: us-1\n---\n- **CA-1** — no rompe\n")
	// El estado guardado es el que quedó después del `sf take`: nadie lo movió.
	p.e.Features["f-1"].Estado = estado.Planificacion

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if !ef.Pasa() {
		t.Fatalf("no dejó empezar el lote de un bug: %v", ef.Fallas)
	}

	// Y el salteo queda ESCRITO: dejar el estado.json diciendo "planificacion"
	// con el rojo ya confirmado es la divergencia que causó todo esto.
	if got := p.e.Features["f-1"].Estado; got != estado.Implementar {
		t.Errorf("el estado quedó en %q, quería implementar", got)
	}
}

// Y el guard sigue vivo para las historias normales: una us en planificación no
// puede saltar a implementar sin pasar por el ⑰.
func TestLoteStartRechazaUnaUsEnPlanificacion(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	p.conArchivoConTexto(docs.Historia("us-1"), "---\ntipo: us\nid: us-1\n---\n- **CA-1** — x\n")
	p.e.Features["f-1"].Estado = estado.Planificacion

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("dejó empezar un lote de una feature que todavía no se planificó")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "es de implementar") {
		t.Errorf("no explicó por qué: %v", ef.Fallas)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El lote de corrección — la vuelta del ㉑
// ────────────────────────────────────────────────────────────────────────────

// Un lote que NO está en tareas.json es el que abrió cerrarRevision para
// arreglar un hallazgo: el plan se escribió antes de que el hallazgo existiera.
// La compuerta afloja a lo comprobable —que la suite falle— y NO le echa la
// culpa al ⑰, que hizo bien su trabajo.
func TestLoteStartDeCorreccionNoCulpaAlPlan(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	commit := "abc123"
	f := p.e.Features["f-1"]
	f.Lotes = []estado.Lote{
		{Lote: 1, Rojo: true, HashTests: "h", Commit: &commit},
		{Lote: 2}, // el de corrección: no está en tareas.json
	}

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if !ef.Pasa() {
		t.Fatalf("no dejó empezar el lote de corrección: %v", ef.Fallas)
	}
	if !strings.Contains(ef.Mensaje, "corrección") {
		t.Errorf("no dijo que era un lote de corrección: %q", ef.Mensaje)
	}
}

// Pero un lote que SÍ está en el plan y no tiene tests sigue siendo el error del
// ⑰. Las dos preguntas dan "cero tests" y no son la misma.
func TestLoteStartDelPlanSinTestsSigueCulpandoAl17(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	p.conArchivoConTexto(filepath.Join(".docs", "features", "f-1-nucleo", "tareas.json"),
		`{"tareas":[{"id":"t-1","lote":1,"satisface":["us-1/CA-1"],"tests":[]}]}`)

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("dejó empezar un lote del plan sin ningún test planificado")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "⑰") {
		t.Errorf("no señaló al ⑰: %v", ef.Fallas)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// A0 — `sf done` no puede saltear `implementar` entero
// ────────────────────────────────────────────────────────────────────────────

// Era el agujero más grande del binario, y era una sola línea de shell.
//
// Con el plan recién aprobado f.Lotes está vacío —los lotes se siembran en
// `sf lote start`—, y tanto compuerta.Implementar como cerrarLote leían ese
// vacío como "todos los lotes están commiteados". Un solo `sf done` movía la
// feature a revisión: sin branch, sin tests, sin código y sin commit. Todo el
// mecanismo del producto se evitaba corriendo un comando una vez.
func TestDoneNoMueveARevisionSinHaberEmpezadoNingunLote(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	p.e.Features["f-1"].Lotes = nil // el plan se aprobó y nadie corrió lote start

	c := Terminar(p.raiz, p.e, p.r, "feat: nada")
	if c.Movio {
		t.Fatal("movió el estado sin que se hubiera empezado ningún lote")
	}
	if got := p.e.Features["f-1"].Estado; got != estado.Implementar {
		t.Errorf("el estado quedó en %q, quería implementar", got)
	}
	if !strings.Contains(strings.Join(c.Fallas, " "), "no empezaste ningún lote") {
		t.Errorf("no dijo qué falta: %v", c.Fallas)
	}
}

// Y la misma puerta para el camino corto, donde el salto era peor todavía: un
// bug llegaba directo a `cierre` —saltea la revisión— sin escribir una línea.
func TestDoneNoMandaUnBugACierreSinHaberEmpezadoNingunLote(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	p.conArchivoConTexto(docs.Historia("us-1"), "---\ntipo: bug\nid: us-1\n---\n- **CA-1** — x\n")
	p.e.Features["f-1"].Estado = estado.Planificacion
	p.e.Features["f-1"].Lotes = nil

	c := Terminar(p.raiz, p.e, p.r, "")
	if got := p.e.Features["f-1"].Estado; got == estado.Cierre {
		t.Fatal("un bug llegó a cierre sin escribir una línea de código")
	}
	if c.Movio && p.e.Features["f-1"].Estado == estado.Revision {
		t.Fatal("un bug no pasa por revisión, y además no empezó ningún lote")
	}
	if !strings.Contains(strings.Join(c.Fallas, " "), "no empezaste ningún lote") {
		t.Errorf("no dijo qué falta: %v", c.Fallas)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// A5 — la vuelta del ㉑ tiene que abrir dónde trabajar
// ────────────────────────────────────────────────────────────────────────────

// Sin lote nuevo, "vuelve a implementar" era un loop muerto: los lotes del plan
// ya estaban todos commiteados, así que `sf lote start` contestaba "todos ya
// están commiteados" y `sf done` rebotaba a revisión. El arreglo del hallazgo no
// tenía dónde commitearse, y la única salida era `sf dismiss` — o sea, declarar
// falso lo que el revisor encontró.
func TestRevisionConHallazgosAbreUnLoteDeCorreccion(t *testing.T) {
	p := nuevo(t).listoParaImplementar(verdeReal)
	commit := "abc123"
	f := p.e.Features["f-1"]
	f.Estado = estado.Revision
	f.Lotes = []estado.Lote{{Lote: 1, Rojo: true, HashTests: "h", Commit: &commit}}

	p.conArchivoConTexto(filepath.Join(".docs", "features", "f-1-nucleo", docs.Revision),
		`{"vuelta":1,"criterios":{"us-1/CA-1":"no-cumple"},"hallazgos":[
			{"id":"h-1","origen":21,"criterio":"us-1/CA-1","estado":"abierto","detalle":"falta el caso vacío"}]}`)
	p.conArchivoConTexto(docs.Historia("us-1"), "---\nid: us-1\n---\n- **CA-1** — x\n")

	c := Terminar(p.raiz, p.e, p.r, "")
	if !c.Movio || f.Estado != estado.Implementar {
		t.Fatalf("no volvió a implementar: estado %q", f.Estado)
	}

	// Lo que faltaba: un lote abierto donde arreglar el hallazgo.
	l, hay := f.LoteActual()
	if !hay {
		t.Fatal("volvió a implementar y no hay ningún lote donde trabajar")
	}
	if l.Lote != 2 {
		t.Errorf("el lote de corrección es el %d, quería el 2", l.Lote)
	}
	if l.Rojo {
		t.Error("el lote nuevo nace con el rojo puesto, y el rojo se ve, no se declara")
	}
}

// Y el bucle deja de rebotar: con el lote abierto, `sf lote start` tiene algo
// que hacer. Es el test que prueba que A5 se cerró de verdad y no sólo que el
// campo cambió.
func TestDespuesDeLaVueltaDel21SePuedeEmpezarElLote(t *testing.T) {
	p := nuevo(t).listoParaImplementar(rojoReal)
	commit := "abc123"
	f := p.e.Features["f-1"]
	f.Estado = estado.Revision
	f.Lotes = []estado.Lote{{Lote: 1, Rojo: true, HashTests: "h", Commit: &commit}}

	p.conArchivoConTexto(filepath.Join(".docs", "features", "f-1-nucleo", docs.Revision),
		`{"vuelta":1,"criterios":{"us-1/CA-1":"no-cumple"},"hallazgos":[
			{"id":"h-1","origen":21,"criterio":"us-1/CA-1","estado":"abierto","detalle":"x"}]}`)
	p.conArchivoConTexto(docs.Historia("us-1"), "---\nid: us-1\n---\n- **CA-1** — x\n")

	if c := Terminar(p.raiz, p.e, p.r, ""); !c.Movio {
		t.Fatalf("no volvió a implementar: %v", c.Fallas)
	}
	if ef := EmpezarLote(p.raiz, p.e, p.r); !ef.Pasa() {
		t.Fatalf("no se pudo empezar el lote de corrección: %v", ef.Fallas)
	}
	if l, _ := f.LoteActual(); !l.Rojo {
		t.Error("no confirmó el rojo del lote de corrección")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El hueco entre "existen" y "algo falla": ¿el rojo es DE ELLOS?
// ────────────────────────────────────────────────────────────────────────────

// Una suite que ya venía rota le regalaba el rojo al lote: los tests
// planificados podían no haberse ejecutado nunca, y el hash se tomaba igual
// sobre un fallo ajeno. Toda la cadena rojo→verde quedaba apoyada ahí.
func TestLoteStartRechazaUnRojoQueNoEsDelLote(t *testing.T) {
	// La suite falla, y falla por otra cosa: nombra un test que no es del lote.
	p := nuevo(t).listoParaImplementar("echo FAIL TestDeOtroPaquete; exit 1")

	ef := EmpezarLote(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("aceptó un rojo que no menciona ningún test del lote")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "no nombra NINGUNO") {
		t.Errorf("no explicó por qué: %v", ef.Fallas)
	}

	// Y no anota nada: sin rojo propio, no hay hash que valga.
	if l, hay := p.e.Features["f-1"].LoteActual(); hay && l.Rojo {
		t.Error("anotó el rojo igual")
	}
}

// Alcanza con que nombre UNO. El lote tiene dos tests planificados y es normal
// que el runner corte en el primero que falla — exigir los dos convertiría una
// compuerta en una molestia.
func TestLoteStartAlcanzaConQueElRojoNombreUnTest(t *testing.T) {
	p := nuevo(t).listoParaImplementar("echo FAIL TestDos; exit 1")

	if ef := EmpezarLote(p.raiz, p.e, p.r); !ef.Pasa() {
		t.Fatalf("no pasó nombrando uno de los dos: %v", ef.Fallas)
	}
}

// En un lote SIN plan no hay lista contra la cual comparar, así que la pregunta
// no se hace y el exit code vuelve a ser todo lo que hay. Son los dos casos
// legítimos: el camino corto de un bug y la vuelta del ㉑.
func TestUnLoteSinPlanNoExigeQueElRojoLoNombre(t *testing.T) {
	p := nuevo(t).listoParaImplementar("exit 1") // mudo, y está bien acá
	if err := os.Remove(filepath.Join(p.raiz, ".docs", "features", "f-1-nucleo", "tareas.json")); err != nil {
		t.Fatal(err)
	}

	if ef := EmpezarLote(p.raiz, p.e, p.r); !ef.Pasa() {
		t.Fatalf("un lote sin plan no puede exigir nombres: %v", ef.Fallas)
	}
}
