package maquina

import (
	"os/exec"
	"path/filepath"
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

// listoParaImplementar deja f-1 en implementar, con constitución, plan y tests.
//
// `testCmd` es lo que va a decidir el rojo o el verde: los tests le pasan
// `exit 1` o `exit 0` directamente, porque lo que se prueba acá es la COMPUERTA,
// no el runner.
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
	p := nuevo(t).listoParaImplementar("exit 1") // la suite falla: hay rojo

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
	p := nuevo(t).listoParaImplementar("exit 1")
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
	p := nuevo(t).listoParaImplementar("exit 0") // ya está verde

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
	p := nuevo(t).listoParaImplementar("exit 1")
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
	p := nuevo(t).listoParaImplementar("exit 1")
	p.e.Features["f-1"].Estado = estado.Planificacion

	if ef := EmpezarLote(p.raiz, p.e, p.r); ef.Pasa() {
		t.Error("dejó correr lote start en planificacion")
	}
}

// Correrlo dos veces avisa en vez de volver a exigir el rojo — que ya no
// existiría, porque el código empezó a escribirse.
func TestLoteStartDosVecesAvisa(t *testing.T) {
	p := nuevo(t).listoParaImplementar("exit 1")
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
	p := nuevo(t).listoParaImplementar("exit 1")
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
	p := nuevo(t).listoParaImplementar("exit 1")
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
	p := nuevo(t).listoParaImplementar("exit 1")
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
