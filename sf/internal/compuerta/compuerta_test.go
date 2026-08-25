package compuerta

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
)

const carpetaF1 = ".docs/features/f-1-nucleo"

var featureF1 = roadmap.Feature{
	ID: "f-1", Slug: "nucleo", Orden: 1, Historias: []string{"us-1"},
}

type proyecto struct {
	raiz string
	t    *testing.T
}

func nuevo(t *testing.T) *proyecto {
	t.Helper()
	return &proyecto{raiz: t.TempDir(), t: t}
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

// planCompleto deja una planificación que pasa todas las compuertas. Cada test
// después rompe UNA cosa, que es lo que hace evidente qué está probando.
func (p *proyecto) planCompleto() *proyecto {
	return p.
		archivo(".docs/backlog/us-1.md",
			"---\nid: us-1\n---\n## Criterios\n- **CA-1** — uno\n- **CA-2** — dos\n").
		archivo(carpetaF1+"/decision.md", "## A — una\n## B — otra\n## C — la tercera\n").
		archivo(carpetaF1+"/spec-design.md", "# spec\n").
		archivo(carpetaF1+"/tareas.json", `{"tareas":[
			{"id":"t-1","lote":1,"satisface":["us-1/CA-1","us-1/CA-2"],"tests":["a_test.go::TestX"]}
		]}`)
}

func exige(t *testing.T, r Resultado, pedazo string) {
	t.Helper()
	if r.Pasa() {
		t.Fatalf("pasó y no debería (esperaba %q)", pedazo)
	}
	if !strings.Contains(strings.Join(r.Fallas, "\n"), pedazo) {
		t.Errorf("las fallas fueron %v y no mencionan %q", r.Fallas, pedazo)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Producto
// ────────────────────────────────────────────────────────────────────────────

// El brief tiene que traer veredicto — y "no-lo-hagas" es tan válido como los
// otros dos: el valor del ⑥ es poder decir que no.
func TestBriefExigeVeredicto(t *testing.T) {
	exige(t, Brief(nuevo(t).raiz), "falta")

	p := nuevo(t).archivo(".docs/brief.md", "---\ntipo: brief\n---\n# sin veredicto\n")
	exige(t, Brief(p.raiz), "veredicto")

	for _, v := range []string{"hacelo", "pivotea", "no-lo-hagas"} {
		p := nuevo(t).archivo(".docs/brief.md", "---\nveredicto: "+v+"\n---\n# ok\n")
		if !Brief(p.raiz).Pasa() {
			t.Errorf("el veredicto %q no pasó", v)
		}
	}
}

// Sin test_cmd sf queda ciego: caen el rojo del ⑲, el verde del ⑳ y el conteo
// del #8. Es la única línea del frontmatter que se exige.
func TestConstitucionExigeTestCmd(t *testing.T) {
	p := nuevo(t).archivo(".docs/constitucion.md", "---\nlenguaje: go\n---\n# reglas\n")
	exige(t, Constitucion(p.raiz), "test_cmd")

	p = nuevo(t).archivo(".docs/constitucion.md", "---\nlenguaje: go\ntest_cmd: go test ./...\n---\n# ok\n")
	if !Constitucion(p.raiz).Pasa() {
		t.Error("con test_cmd no pasó")
	}
}

// Sin criterios con id, el revisor contesta "anda" y no hay nada que contar
// después. Ésta es la compuerta que HABILITA el mecanismo del ㉑.
func TestBacklogExigeCriteriosConID(t *testing.T) {
	p := nuevo(t).archivo(".docs/backlog/us-1.md", "---\nid: us-1\n---\nQuiero algo.\n")
	exige(t, Backlog(p.raiz), "criterios")

	p = nuevo(t).archivo(".docs/backlog/us-1.md", "---\nid: us-1\n---\n- **CA-1** — algo\n")
	if !Backlog(p.raiz).Pasa() {
		t.Error("con CA-1 no pasó")
	}
}

// Una historia que quedó fuera de todas las features no la implementa nadie, y
// nadie se entera. Ése es el error que esta compuerta atrapa.
func TestRoadmapAtrapaHistoriasHuerfanas(t *testing.T) {
	p := nuevo(t).
		archivo(".docs/backlog/us-1.md", "---\nid: us-1\n---\n- **CA-1** — x\n").
		archivo(".docs/backlog/us-9.md", "---\nid: us-9\n---\n- **CA-1** — y\n").
		archivo(".docs/roadmap.json", `{"features":[{"id":"f-1","orden":1,"historias":["us-1"]}]}`)

	exige(t, Roadmap(p.raiz), "us-9")
}

// Una feature grande es un aviso, no un freno: partirla o no es criterio, y el
// criterio es de Javier.
func TestRoadmapAvisaFeatureGrandeSinFrenar(t *testing.T) {
	p := nuevo(t).archivo(".docs/roadmap.json",
		`{"features":[{"id":"f-1","orden":1,"historias":["us-1","us-2","us-3","us-4","us-5","us-6"]}]}`)

	r := Roadmap(p.raiz)
	if !r.Pasa() {
		t.Errorf("una feature grande FRENÓ, y sólo tiene que avisar: %v", r.Fallas)
	}
	if len(r.Avisos) == 0 {
		t.Error("no avisó de la feature de 6 historias")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// planificacion — las cinco compuertas
// ────────────────────────────────────────────────────────────────────────────

func TestPlanificacionCompletaPasa(t *testing.T) {
	p := nuevo(t).planCompleto()
	if r := Planificacion(p.raiz, featureF1); !r.Pasa() {
		t.Errorf("un plan completo no pasó: %v", r.Fallas)
	}
}

func TestPlanificacionExigeLosTresArchivos(t *testing.T) {
	p := nuevo(t).planCompleto()
	os.Remove(filepath.Join(p.raiz, carpetaF1, "spec-design.md"))

	exige(t, Planificacion(p.raiz, featureF1), "spec-design.md")
}

// Tres, no "varias": tres es el número del ⑫, y contar hasta tres es
// exactamente lo que sf puede hacer sin opinar.
func TestPlanificacionExigeTresOpciones(t *testing.T) {
	p := nuevo(t).planCompleto().
		archivo(carpetaF1+"/decision.md", "## A — una\n## B — la elegida\n")

	exige(t, Planificacion(p.raiz, featureF1), "2 opciones")
}

// Sin la lista de tests, `sf lote start` no tiene contra qué exigir el rojo —
// que es toda la defensa contra el dolor #8.
func TestPlanificacionExigeTestsEnCadaLote(t *testing.T) {
	p := nuevo(t).planCompleto().
		archivo(carpetaF1+"/tareas.json", `{"tareas":[
			{"id":"t-1","lote":1,"satisface":["us-1/CA-1","us-1/CA-2"],"tests":["a_test.go::TestX"]},
			{"id":"t-2","lote":2,"satisface":["us-1/CA-1"]}
		]}`)

	exige(t, Planificacion(p.raiz, featureF1), "lote 2 no tiene ningún test")
}

// Acá empieza a morir el dolor #7, y ANTES de escribir una línea de código: si
// las tareas no cubren todos los criterios, la feature ya nace a medias.
func TestPlanificacionCuentaLosCriteriosSinCubrir(t *testing.T) {
	p := nuevo(t).planCompleto().
		archivo(carpetaF1+"/tareas.json", `{"tareas":[
			{"id":"t-1","lote":1,"satisface":["us-1/CA-1"],"tests":["a_test.go::TestX"]}
		]}`)

	r := Planificacion(p.raiz, featureF1)
	exige(t, r, "us-1/CA-2")
	if !strings.Contains(strings.Join(r.Fallas, ""), "2 criterios") {
		t.Errorf("no dijo cuántos criterios hay: %v", r.Fallas)
	}
}

// El espejo del anterior, y atrapa el error opuesto: una tarea que dice
// satisfacer un criterio que no existe está mintiendo, y el conteo de arriba no
// lo vería.
func TestPlanificacionAtrapaCriteriosInventados(t *testing.T) {
	p := nuevo(t).planCompleto().
		archivo(carpetaF1+"/tareas.json", `{"tareas":[
			{"id":"t-1","lote":1,"satisface":["us-1/CA-1","us-1/CA-2","us-1/CA-9"],
			 "tests":["a_test.go::TestX"]}
		]}`)

	exige(t, Planificacion(p.raiz, featureF1), "us-1/CA-9")
}

// ────────────────────────────────────────────────────────────────────────────
// revision — sf cuenta, no opina
// ────────────────────────────────────────────────────────────────────────────

func revisionCon(p *proyecto, json string) *proyecto {
	return p.archivo(".docs/backlog/us-1.md",
		"---\nid: us-1\n---\n- **CA-1** — uno\n- **CA-2** — dos\n").
		archivo(carpetaF1+"/revision.json", json)
}

// Acá muere el dolor #7. sf no juzga la revisión: comprueba que haya ocurrido
// SOBRE TODOS, y dice cuáles faltan.
func TestRevisionCuentaLosCriteriosSinVeredicto(t *testing.T) {
	p := revisionCon(nuevo(t), `{"vuelta":1,"criterios":{"us-1/CA-1":"cumple"}}`)

	r := Revision(p.raiz, featureF1)
	exige(t, r, "us-1/CA-2")
	if !strings.Contains(strings.Join(r.Fallas, ""), "opina sobre 1") {
		t.Errorf("no dijo sobre cuántos opinó: %v", r.Fallas)
	}
}

// Con un hallazgo abierto no avanza. Es un conteo, no un juicio.
func TestRevisionFrenaConUnHallazgoAbierto(t *testing.T) {
	p := revisionCon(nuevo(t), `{"vuelta":1,
		"criterios":{"us-1/CA-1":"cumple","us-1/CA-2":"no-cumple"},
		"hallazgos":[{"id":"h-1","estado":"abierto","detalle":"x"}]}`)

	exige(t, Revision(p.raiz, featureF1), "h-1")
}

// `descartado` SÍ pasa: es una decisión de Javier, y una compuerta no puede
// frenar sobre algo que él ya resolvió (R3).
func TestRevisionDejaPasarLosDescartados(t *testing.T) {
	p := revisionCon(nuevo(t), `{"vuelta":1,
		"criterios":{"us-1/CA-1":"cumple","us-1/CA-2":"cumple"},
		"hallazgos":[{"id":"h-1","estado":"descartado","motivo":"es intencional"}]}`)

	if r := Revision(p.raiz, featureF1); !r.Pasa() {
		t.Errorf("un hallazgo descartado frenó: %v", r.Fallas)
	}
}

// Si el ㉑ va por la cuarta vuelta, eso no es ruido: es que la planificación se
// quedó corta. Avisa, no frena.
func TestRevisionAvisaCuandoVaPorLaTerceraVuelta(t *testing.T) {
	p := revisionCon(nuevo(t), `{"vuelta":3,
		"criterios":{"us-1/CA-1":"cumple","us-1/CA-2":"cumple"}}`)

	r := Revision(p.raiz, featureF1)
	if !r.Pasa() {
		t.Errorf("la vuelta 3 frenó, y sólo tiene que avisar: %v", r.Fallas)
	}
	if len(r.Avisos) == 0 {
		t.Error("no avisó de la tercera vuelta")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// implementar y cierre
// ────────────────────────────────────────────────────────────────────────────

// Sin haber visto el rojo no hay verde que valga. Es la compuerta del medio, y
// funciona incluso antes de que `sf lote start` exista: si nunca vio el rojo,
// no pasa.
func TestImplementarNoPasaSinRojo(t *testing.T) {
	f := &estado.Feature{Lotes: []estado.Lote{{Lote: 1, Rojo: false}}}
	exige(t, Implementar(f), "no vi el rojo")

	f = &estado.Feature{Lotes: []estado.Lote{{Lote: 1, Rojo: true}}}
	if !Implementar(f).Pasa() {
		t.Error("con el rojo confirmado no pasó")
	}
}

// Una lista de lotes VACÍA no es una lista terminada.
//
// Era el agujero más grande del binario: con el plan recién aprobado f.Lotes
// está vacío —los lotes se siembran en `sf lote start`—, y esta compuerta
// devolvía "pasa" porque LoteActual() da false igual que cuando están todos
// commiteados. Un solo `sf done` movía la feature a revisión sin branch, sin
// tests, sin código y sin commit.
func TestImplementarFrenaSinLotesSembrados(t *testing.T) {
	exige(t, Implementar(&estado.Feature{}), "no empezaste ningún lote")
	exige(t, Implementar(&estado.Feature{Lotes: []estado.Lote{}}), "no empezaste ningún lote")
}

// Y la otra mitad: con TODOS commiteados sí se puede cerrar el estado. Sin este
// test, el arreglo de arriba podría frenar el cierre legítimo y nadie se
// enteraría — los dos casos entran por el mismo `false` de LoteActual().
func TestImplementarPasaConTodosLosLotesCommiteados(t *testing.T) {
	uno, dos := "abc123", "def456"
	f := &estado.Feature{Lotes: []estado.Lote{
		{Lote: 1, Rojo: true, Commit: &uno},
		{Lote: 2, Rojo: true, Commit: &dos},
	}}
	if !Implementar(f).Pasa() {
		t.Errorf("con todos los lotes commiteados no pasó: %v", Implementar(f).Fallas)
	}
}

// El ㉓ agrega DOS archivos, no uno: la doc y el journal.
func TestCierreExigeLaDocYElJournal(t *testing.T) {
	p := nuevo(t).archivo(carpetaF1+"/doc.md", "# doc\n")
	exige(t, Cierre(p.raiz, featureF1), "journal.md")

	p.archivo(carpetaF1+"/journal.md", "- no uses regex\n")
	if !Cierre(p.raiz, featureF1).Pasa() {
		t.Error("con la doc y el journal no pasó")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El texto del veredicto
// ────────────────────────────────────────────────────────────────────────────

// Los ✓ no se listan. Un veredicto que enumera todo lo que salió bien obliga a
// buscar la ✗ entre quince líneas.
func TestElTextoSoloMuestraLoQueFalta(t *testing.T) {
	var r Resultado
	r.falla("falta X")
	r.avisa("ojo con Y")

	txt := r.Texto()
	if !strings.Contains(txt, "✗ No avanzo") || !strings.Contains(txt, "falta X") {
		t.Errorf("el veredicto no dice qué falta:\n%s", txt)
	}
	if !strings.Contains(txt, "⚠ ojo con Y") {
		t.Errorf("el aviso no aparece:\n%s", txt)
	}

	var ok Resultado
	if !strings.Contains(ok.Texto(), "✓") {
		t.Errorf("un resultado limpio no dice que está listo:\n%s", ok.Texto())
	}
}
