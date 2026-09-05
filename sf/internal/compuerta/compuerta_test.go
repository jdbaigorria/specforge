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
			{"id":"t-1","lote":1,"satisface":["us-1/CA-1","us-1/CA-2"],
			 "tests":["a_test.go::TestX","a_test.go::TestY"]}
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
		p := nuevo(t).archivo(".docs/brief.md",
			"---\nveredicto: "+v+"\n---\n# ok\n\n`retrieved` https://github.com/x/y\n")
		if !Brief(p.raiz).Pasa() {
			t.Errorf("el veredicto %q no pasó", v)
		}
	}
}

// El caso medido el 2026-09-05: nemotron selló `hacelo` con 8 afirmaciones
// `model-prior`, CERO links, y esta compuerta lo dejó pasar. El producto siguió
// hasta el PRD apoyado en nada.
//
// Los tres veredictos, no sólo `hacelo`: el propio skill dice que la
// alucinación más cara del ⑥ es sellar `no-lo-hagas` por algo que no existe.
func TestBriefRechazaVeredictoSinUnaSolaFuente(t *testing.T) {
	for _, v := range []string{"hacelo", "pivotea", "no-lo-hagas"} {
		p := nuevo(t).archivo(".docs/brief.md",
			"---\nveredicto: "+v+"\n---\n# el panorama\n\n"+
				"| ccusage | lo mismo | nada | `model-prior` sin verificar |\n")
		exige(t, Brief(p.raiz), "no cita una sola fuente")
	}
}

// La salida ya estaba diseñada en el skill: sin los MCPs de investigación, con
// consentimiento explícito, corre una pasada degradada y el brief queda marcado.
// Pasa, pero NO en silencio.
func TestBriefAceptaLaBajaEvidenciaDeclarada(t *testing.T) {
	p := nuevo(t).archivo(".docs/brief.md",
		"---\nveredicto: hacelo\nevidencia: baja\n---\n# sin un solo link\n")

	r := Brief(p.raiz)
	if !r.Pasa() {
		t.Fatalf("con `evidencia: baja` tiene que pasar: %v", r.Fallas)
	}
	if !strings.Contains(strings.Join(r.Avisos, ""), "BAJA EVIDENCIA") {
		t.Errorf("tiene que avisar que la evidencia es baja: %v", r.Avisos)
	}
}

// Un link es un link esté donde esté: no se parsea el cuerpo ni se exige que
// viva en la tabla de procedencia. Es un hecho sobre bytes.
func TestBriefCuentaLinksEnCualquierParte(t *testing.T) {
	for _, cuerpo := range []string{
		"# ok\n\nver http://example.com\n",
		"# ok\n\n[la fuente](https://example.com)\n",
		"# ok\n\n> https://news.ycombinator.com/item?id=1\n",
	} {
		p := nuevo(t).archivo(".docs/brief.md", "---\nveredicto: hacelo\n---\n"+cuerpo)
		if !Brief(p.raiz).Pasa() {
			t.Errorf("no encontró el link en %q", cuerpo)
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

// `mutacion:` vacío puede ser la respuesta correcta —hay stacks sin herramienta
// buena— así que esto avisa y NO frena. Lo que sf sí puede afirmar es el hecho
// de al lado: para este lenguaje existe una.
func TestConstitucionAvisaSiFaltaLaHerramientaDeMutacion(t *testing.T) {
	p := nuevo(t).archivo(".docs/constitucion.md",
		"---\nlenguaje: node\ntest_cmd: npm test\nmutacion: \"\"\n---\n# reglas\n")

	r := Constitucion(p.raiz)
	if !r.Pasa() {
		t.Fatalf("frenó, y sólo tiene que avisar: %v", r.Fallas)
	}
	if !strings.Contains(strings.Join(r.Avisos, ""), "Stryker") {
		t.Errorf("el aviso tiene que nombrar la herramienta: %v", r.Avisos)
	}
}

func TestConstitucionNoAvisaSiLaHerramientaEstaDeclarada(t *testing.T) {
	p := nuevo(t).archivo(".docs/constitucion.md",
		"---\nlenguaje: node\ntest_cmd: npm test\nmutacion: \"npx stryker run\"\n---\n# reglas\n")

	if r := Constitucion(p.raiz); len(r.Avisos) > 0 {
		t.Errorf("la declaró y avisó igual: %v", r.Avisos)
	}
}

// Un lenguaje que sf no tiene en la tabla no genera ruido: no sabe si existe
// una herramienta, y un aviso que no se puede accionar se deja de leer.
func TestConstitucionNoAvisaDeUnLenguajeQueNoConoce(t *testing.T) {
	p := nuevo(t).archivo(".docs/constitucion.md",
		"---\nlenguaje: cobol\ntest_cmd: hacelo\nmutacion: \"\"\n---\n# reglas\n")

	if r := Constitucion(p.raiz); len(r.Avisos) > 0 {
		t.Errorf("avisó de un lenguaje que no conoce: %v", r.Avisos)
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
			{"id":"t-1","lote":1,"satisface":["us-1/CA-1","us-1/CA-2"],
			 "tests":["a_test.go::TestX","a_test.go::TestY"]},
			{"id":"t-2","lote":2,"satisface":["us-1/CA-1"]}
		]}`)

	exige(t, Planificacion(p.raiz, featureF1), "lote 2 no tiene ningún test")
}

// Un piso de un test por LOTE se vuelve techo de un test: la tarea nombra el
// camino feliz, y todo lo demás aparece tres vueltas de revisión más tarde,
// cuando el ㉒ lo encuentra. Contar por tarea es lo más barato que lo frena.
func TestPlanificacionExigeUnTestPorCriterio(t *testing.T) {
	p := nuevo(t).planCompleto().
		archivo(carpetaF1+"/tareas.json", `{"tareas":[
			{"id":"t-1","lote":1,"satisface":["us-1/CA-1","us-1/CA-2"],"tests":["a_test.go::TestX"]}
		]}`)

	r := Planificacion(p.raiz, featureF1)
	exige(t, r, "la tarea t-1")
	if !strings.Contains(strings.Join(r.Fallas, ""), "2 criterios y 1 tests") {
		t.Errorf("el mensaje tiene que decir los dos números: %v", r.Fallas)
	}
}

// Repetir el mismo nombre de test es la forma obvia de cumplir una regla de
// contar sin cumplir la regla.
func TestPlanificacionNoCuentaDosVecesElMismoTest(t *testing.T) {
	p := nuevo(t).planCompleto().
		archivo(carpetaF1+"/tareas.json", `{"tareas":[
			{"id":"t-1","lote":1,"satisface":["us-1/CA-1","us-1/CA-2"],
			 "tests":["a_test.go::TestX","a_test.go::TestX"]}
		]}`)

	exige(t, Planificacion(p.raiz, featureF1), "la tarea t-1")
}

// Y el espejo: dos criterios con dos tests distintos pasa. Sin esto, los dos
// de arriba pasarían igual con una compuerta que rechace todo.
func TestPlanificacionPasaConUnTestPorCriterio(t *testing.T) {
	p := nuevo(t).planCompleto()

	if r := Planificacion(p.raiz, featureF1); !r.Pasa() {
		t.Fatalf("dos criterios con dos tests tiene que pasar: %v", r.Fallas)
	}
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

// Un mutante que murió antes y vive ahora es una regresión de cobertura: había
// un test que lo agarraba y ya no. Reportarlo sin abrir nada lo pierde.
func TestRevisionFrenaSiUnMutanteResucitaYNadieAbreNada(t *testing.T) {
	p := revisionCon(nuevo(t), `{"vuelta":2,
		"criterios":{"us-1/CA-1":"cumple","us-1/CA-2":"cumple"},
		"mutantes":{"propios":{"corridos":18,"sobrevivieron":0,"resucitados":2}}}`)

	exige(t, Revision(p.raiz, featureF1), "resucitaron")
}

// Con el hallazgo del ㉒ abierto sí avanza... hasta la compuerta de los
// hallazgos abiertos, que es la que lo manda de vuelta. Se prueba con uno
// descartado para aislar ESTA compuerta de aquélla.
func TestRevisionPasaSiLaResurreccionTieneHallazgo(t *testing.T) {
	p := revisionCon(nuevo(t), `{"vuelta":2,
		"criterios":{"us-1/CA-1":"cumple","us-1/CA-2":"cumple"},
		"mutantes":{"propios":{"corridos":18,"resucitados":2}},
		"hallazgos":[{"id":"h-1","origen":22,"estado":"descartado","motivo":"el test se borró a propósito"}]}`)

	if r := Revision(p.raiz, featureF1); !r.Pasa() {
		t.Errorf("la resurrección tenía su hallazgo y frenó igual: %v", r.Fallas)
	}
}

// Un hallazgo del ㉑ no cubre una resurrección: son dos preguntas distintas, y
// sin esto la compuerta se cumpliría con cualquier hallazgo de cualquier lado.
func TestRevisionNoAceptaUnHallazgoDelVeintiunoPorUnaResurreccion(t *testing.T) {
	p := revisionCon(nuevo(t), `{"vuelta":2,
		"criterios":{"us-1/CA-1":"cumple","us-1/CA-2":"cumple"},
		"mutantes":{"propios":{"resucitados":1}},
		"hallazgos":[{"id":"h-1","origen":21,"estado":"descartado","motivo":"x"}]}`)

	exige(t, Revision(p.raiz, featureF1), "resucitaron")
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

// Un criterio con id y sin texto contaba como criterio, y el esqueleto de
// `sf new` deja exactamente eso: `- **CA-1** —`.
//
// Es PEOR que no tener ninguno: el ⑰ lo da por cubierto y el ㉑ le pone
// veredicto, así que el mecanismo de los ids queda en pie sobre algo que nadie
// puede juzgar.
func TestBacklogRechazaElCriterioSinTexto(t *testing.T) {
	p := nuevo(t).archivo(".docs/backlog/us-1.md",
		"---\nid: us-1\ntitulo: sumar\n---\n## Criterios\n- **CA-1** —\n")
	exige(t, Backlog(p.raiz), "criterio vacío")

	p.archivo(".docs/backlog/us-1.md",
		"---\nid: us-1\ntitulo: sumar\n---\n## Criterios\n- **CA-1** — Suma(2,3) da 5\n")
	if !Backlog(p.raiz).Pasa() {
		t.Errorf("no pasó con el criterio escrito: %v", Backlog(p.raiz).Fallas)
	}
}

// Y el título NO se exige acá, a propósito: la compuerta pregunta si el
// MECANISMO se sostiene, y lo que el ⑰ cuenta y el ㉑ juzga son los criterios.
// De la historia sin pinponear se ocupa el checkpoint del ⑨.
func TestBacklogNoExigeTitulo(t *testing.T) {
	p := nuevo(t).archivo(".docs/backlog/us-1.md",
		"---\nid: us-1\n---\n## Criterios\n- **CA-1** — algo concreto\n")
	if !Backlog(p.raiz).Pasa() {
		t.Errorf("frenó por el título: %v", Backlog(p.raiz).Fallas)
	}
}
