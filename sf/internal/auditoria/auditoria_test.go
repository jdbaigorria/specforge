package auditoria

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
)

// Igual que en `maquina` y en `arranque`, los tests arman un proyecto de verdad
// en un directorio temporal. Acá pesa más que en ningún lado: la comprobación
// ④ —"¿el test sigue existiendo?"— es literalmente mirar el disco, y un mock la
// haría pasar sin probar nada.

type proyecto struct {
	raiz string
	e    *estado.Estado
	r    *roadmap.Roadmap
	t    *testing.T
}

func (p *proyecto) archivo(rel, texto string) *proyecto {
	p.t.Helper()
	ruta := filepath.Join(p.raiz, rel)
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		p.t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(texto), 0o644); err != nil {
		p.t.Fatal(err)
	}
	return p
}

// unProyectoCerrado arma el caso base: f-1 construida, revisada y archivada, con
// un criterio dado por cumplido y el test que lo probaba EN SU LUGAR.
//
// De acá salen todos los tests: cada uno rompe UNA cosa y comprueba que la
// auditoría la vea.
func unProyectoCerrado(t *testing.T) *proyecto {
	t.Helper()
	p := &proyecto{
		raiz: t.TempDir(),
		e:    &estado.Estado{Features: map[string]*estado.Feature{}},
		t:    t,
	}

	p.archivo(roadmap.Archivo, `{"features":[
		{"id":"f-1","slug":"nucleo","nombre":"núcleo","orden":1,"historias":["us-1"]}]}`)
	r, err := roadmap.Leer(p.raiz)
	if err != nil {
		t.Fatal(err)
	}
	p.r = r

	p.archivo(".docs/backlog/us-1.md", `---
tipo: us
id: us-1
---

# us-1

## Criterios de aceptación
- **CA-1** — devuelve el estado serializado
- **CA-2** — sin estado, sale con código 1
`)

	const carpeta = ".docs/archivado/f-1-nucleo"
	p.archivo(carpeta+"/spec-design.md", "# f-1\n")
	p.archivo(carpeta+"/tareas.json", `{"feature":"f-1","tareas":[
		{"id":"t-1","lote":1,"descripcion":"serializar","satisface":["us-1/CA-1"],
		 "tests":["core_test.go::TestSerializa"]},
		{"id":"t-2","lote":1,"descripcion":"el error","satisface":["us-1/CA-2"],
		 "tests":["core_test.go::TestSinEstado"]}]}`)
	p.archivo(carpeta+"/revision.json", `{"feature":"f-1","vuelta":1,"veredicto":"limpio",
		"criterios":{"us-1/CA-1":"cumple","us-1/CA-2":"cumple"},"hallazgos":[]}`)

	// Los tests, existiendo de verdad.
	p.archivo("core_test.go", "package x\n\nfunc TestSerializa(t *testing.T) {}\nfunc TestSinEstado(t *testing.T) {}\n")

	p.e.Producto = estado.Producto{
		BriefSellado: "hacelo", PrdHash: "a3f9c1",
		ConstitucionSellada: true, BacklogVisto: true,
	}
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Cerrada}
	return p
}

func (p *proyecto) auditar() *Informe {
	p.t.Helper()
	alcance, err := Alcance(p.e, p.r, nil)
	if err != nil {
		p.t.Fatalf("Alcance: %v", err)
	}
	inf, err := Auditar(p.raiz, p.e, p.r, alcance)
	if err != nil {
		p.t.Fatalf("Auditar: %v", err)
	}
	return inf
}

// texto es lo que ve el auditor, que es lo único que importa de verdad: un
// hecho que sf calcula y no imprime no le sirve a nadie.
func (p *proyecto) texto() string {
	p.t.Helper()
	return p.auditar().Texto(p.raiz, false)
}

// ────────────────────────────────────────────────────────────────────────────
// ④ — donde muere "el implementador mintió"
// ────────────────────────────────────────────────────────────────────────────

// EL TEST QUE JUSTIFICA EL COMANDO. La revisión de f-1 dijo que us-1/CA-2 se
// cumple; el test que lo probaba se borró después. El ㉑ de f-1 ya pasó y nadie
// vuelve a mirarla — sin esto, la afirmación queda en pie para siempre.
func TestDetectaElCriterioCumplidoSinSuTest(t *testing.T) {
	p := unProyectoCerrado(t)
	p.archivo("core_test.go", "package x\n\nfunc TestSerializa(t *testing.T) {}\n") // se fue TestSinEstado

	txt := p.texto()
	if !strings.Contains(txt, "us-1/CA-2") || !strings.Contains(txt, "ya no existe") {
		t.Errorf("no detectó el test borrado:\n%s", txt)
	}
	// Y el que sigue estando no se reporta: un informe que se queja de todo no
	// se lee.
	if strings.Contains(txt, "us-1/CA-1 se dio por cumplido") {
		t.Errorf("se quejó de un criterio cuyo test SÍ está:\n%s", txt)
	}
}

// El caso limpio tiene que salir limpio. Es la otra mitad del anterior: una
// auditoría que siempre encuentra algo no distingue nada.
func TestUnProyectoSanoNoTieneFallas(t *testing.T) {
	inf := unProyectoCerrado(t).auditar()

	if n := inf.Fallas(); n != 0 {
		t.Errorf("%d fallas en un proyecto sano:\n%s", n, inf.Texto("", false))
	}
}

func TestDetectaElCriterioSinVeredicto(t *testing.T) {
	p := unProyectoCerrado(t)
	p.archivo(".docs/archivado/f-1-nucleo/revision.json",
		`{"feature":"f-1","vuelta":1,"veredicto":"limpio",
		  "criterios":{"us-1/CA-1":"cumple"},"hallazgos":[]}`) // falta CA-2

	if txt := p.texto(); !strings.Contains(txt, "us-1/CA-2 no tiene veredicto") {
		t.Errorf("no detectó el criterio sin veredicto:\n%s", txt)
	}
}

// Un criterio dado por cumplido que ninguna tarea se comprometió a testear: la
// revisión opinó sobre algo que el plan nunca planeó probar.
func TestDetectaElCumplidoQueNadieSeComprometioATestear(t *testing.T) {
	p := unProyectoCerrado(t)
	p.archivo(".docs/archivado/f-1-nucleo/tareas.json", `{"feature":"f-1","tareas":[
		{"id":"t-1","lote":1,"descripcion":"serializar","satisface":["us-1/CA-1"],
		 "tests":["core_test.go::TestSerializa"]}]}`) // CA-2 sin tarea

	if txt := p.texto(); !strings.Contains(txt, "ninguna tarea declaró un test") {
		t.Errorf("no detectó el criterio sin test planificado:\n%s", txt)
	}
}

// Lo que Javier descartó en el ㉑ vuelve a la superficie. No es una falla —él lo
// decidió— pero el auditor mira otra pregunta, y un descarte repetido es un
// patrón.
func TestLosHallazgosDescartadosVuelvenComoAviso(t *testing.T) {
	p := unProyectoCerrado(t)
	p.archivo(".docs/archivado/f-1-nucleo/revision.json",
		`{"feature":"f-1","vuelta":1,"veredicto":"con-hallazgos",
		  "criterios":{"us-1/CA-1":"cumple","us-1/CA-2":"cumple"},
		  "hallazgos":[{"id":"h-1","origen":22,"criterio":"","estado":"descartado",
		                "motivo":"el mutante no representa un caso real","detalle":"x"}]}`)

	txt := p.texto()
	if !strings.Contains(txt, "h-1 se descartó") || !strings.Contains(txt, "no representa un caso real") {
		t.Errorf("no trajo de vuelta el hallazgo descartado:\n%s", txt)
	}
	if n := p.auditar().Fallas(); n != 0 {
		t.Errorf("un descarte de Javier contó como falla (%d)", n)
	}
}

// Una feature sin revisión es un AVISO y no una falla, porque tiene un caso
// legítimo: un bug saltea el ㉑ entero.
func TestSinRevisionAvisaEnVezDeFallar(t *testing.T) {
	p := unProyectoCerrado(t)
	if err := os.Remove(filepath.Join(p.raiz, ".docs/archivado/f-1-nucleo/revision.json")); err != nil {
		t.Fatal(err)
	}

	inf := p.auditar()
	if txt := inf.Texto(p.raiz, false); !strings.Contains(txt, "no tiene revision.json") {
		t.Errorf("no avisó de la feature sin revisar:\n%s", txt)
	}
	if n := inf.Fallas(); n != 0 {
		t.Errorf("la feature sin revisión contó como falla (%d): un bug saltea el ㉑", n)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El alcance
// ────────────────────────────────────────────────────────────────────────────

func TestElAlcanceSaleEnOrdenDeRoadmap(t *testing.T) {
	p := unProyectoCerrado(t)
	p.archivo(".docs/backlog/us-2.md", "---\nid: us-2\n---\n\n- **CA-1** — x\n")
	p.archivo(".docs/backlog/us-3.md", "---\nid: us-3\n---\n\n- **CA-1** — x\n")
	p.archivo(roadmap.Archivo, `{"features":[
		{"id":"f-1","slug":"a","nombre":"a","orden":1,"historias":["us-1"]},
		{"id":"f-2","slug":"b","nombre":"b","orden":2,"historias":["us-2"]},
		{"id":"f-3","slug":"c","nombre":"c","orden":3,"historias":["us-3"]}]}`)
	r, err := roadmap.Leer(p.raiz)
	if err != nil {
		t.Fatal(err)
	}

	// Se piden desordenadas a propósito: el orden es información —f-1 se
	// construyó antes que f-3— y el auditor la necesita para saber quién pudo
	// romper a quién.
	got, err := Alcance(p.e, r, []string{"f-3", "f-1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != "f-1" || got[1] != "f-3" {
		t.Errorf("alcance %v, quería [f-1 f-3]", got)
	}
}

func TestSinArgumentosAuditaSoloLoConstruido(t *testing.T) {
	p := unProyectoCerrado(t)
	p.archivo(".docs/backlog/us-2.md", "---\nid: us-2\n---\n\n- **CA-1** — x\n")
	p.archivo(roadmap.Archivo, `{"features":[
		{"id":"f-1","slug":"a","nombre":"a","orden":1,"historias":["us-1"]},
		{"id":"f-2","slug":"b","nombre":"b","orden":2,"historias":["us-2"]}]}`)
	r, err := roadmap.Leer(p.raiz)
	if err != nil {
		t.Fatal(err)
	}
	// f-2 está en el roadmap pero nunca arrancó: no aparece en el mapa (R6).

	got, err := Alcance(p.e, r, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0] != "f-1" {
		t.Errorf("alcance %v, quería sólo [f-1]: f-2 no se construyó", got)
	}
}

func TestUnaFeatureQueNoEstaEnElRoadmapEsError(t *testing.T) {
	p := unProyectoCerrado(t)
	if _, err := Alcance(p.e, p.r, []string{"f-99"}); err == nil {
		t.Error("aceptó una feature que no existe")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El sobre
// ────────────────────────────────────────────────────────────────────────────

// El material tiene que apuntar a la carpeta ARCHIVADA, no a la viva: una
// feature cerrada ya se movió, y servir la ruta vieja sería servir nada.
func TestElSobreApuntaALoArchivado(t *testing.T) {
	txt := unProyectoCerrado(t).texto()

	for _, quiero := range []string{
		".docs/archivado/f-1-nucleo/spec-design.md",
		".docs/archivado/f-1-nucleo/revision.json",
		".docs/backlog/us-1.md",
		".docs/constitucion.md",
	} {
		if !strings.Contains(txt, quiero) {
			t.Errorf("el sobre no trae %s:\n%s", quiero, txt)
		}
	}
}

// Los hechos van ARRIBA del material: el que lee es un modelo que va a gastar
// contexto abriendo archivos, y saber qué buscar cambia qué abre.
func TestLosHechosVanAntesQueElMaterial(t *testing.T) {
	txt := unProyectoCerrado(t).texto()

	hechos := strings.Index(txt, "Lo que sf comprobó")
	material := strings.Index(txt, "Las reglas que todo esto")
	if hechos < 0 || material < 0 {
		t.Fatalf("falta una de las dos secciones:\n%s", txt)
	}
	if hechos > material {
		t.Errorf("el material quedó antes que los hechos:\n%s", txt)
	}
}
