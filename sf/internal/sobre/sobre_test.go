package sobre

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
)

const carpetaF1 = ".docs/features/f-1-nucleo"

type proyecto struct {
	raiz string
	e    *estado.Estado
	r    *roadmap.Roadmap
	t    *testing.T
}

func nuevo(t *testing.T) *proyecto {
	t.Helper()
	return &proyecto{
		raiz: t.TempDir(),
		e:    &estado.Estado{Features: map[string]*estado.Feature{}},
		t:    t,
	}
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

func (p *proyecto) conRoadmap() *proyecto {
	p.archivo(roadmap.Archivo, `{"features":[
		{"id":"f-1","slug":"nucleo","nombre":"núcleo","orden":1,"historias":["us-1","us-3"]}]}`)
	r, err := roadmap.Leer(p.raiz)
	if err != nil {
		p.t.Fatal(err)
	}
	p.r = r
	return p
}

// enFeature deja el producto sellado y la feature f-1 en el estado que se pida.
func (p *proyecto) enFeature(est string) *proyecto {
	p.e.Producto = estado.Producto{
		BriefSellado: "hacelo", PrdHash: "a3f9c1",
		ConstitucionSellada: true, BacklogVisto: true,
	}
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: est}
	return p.conRoadmap()
}

// texto arma el sobre y lo renderiza, que es lo que ve el que trabaja.
func (p *proyecto) texto() string {
	p.t.Helper()
	s, err := Armar(p.raiz, p.e, p.r)
	if err != nil {
		p.t.Fatalf("Armar: %v", err)
	}
	return s.Texto(p.raiz, false)
}

// ────────────────────────────────────────────────────────────────────────────
// Producto
// ────────────────────────────────────────────────────────────────────────────

// Cada sobre de producto trae exactamente lo que su paso necesita leer, y el
// del brief no trae nada: ①–⑤ es un pinponeo sobre una idea sin forma.
func TestLosSobresDeProducto(t *testing.T) {
	casos := []struct {
		nombre string
		armar  func(*proyecto)
		quiero []string
		noVa   []string
	}{
		{
			nombre: "brief: no hay nada que leer",
			armar:  func(p *proyecto) {},
			quiero: []string{"pinponeo"},
		},
		{
			nombre: "prd: sale del brief",
			armar:  func(p *proyecto) { p.e.Producto.BriefSellado = "hacelo" },
			quiero: []string{".docs/brief.md"},
		},
		{
			nombre: "constitucion: el PRD, y la cabecera técnica ya llena",
			armar: func(p *proyecto) {
				p.e.Producto.BriefSellado = "hacelo"
				p.e.Producto.PrdHash = "a3f9c1"
			},
			quiero: []string{".docs/prd.md", ".docs/constitucion.md"},
		},
		{
			nombre: "backlog: el PRD y las reglas",
			armar: func(p *proyecto) {
				p.e.Producto = estado.Producto{
					BriefSellado: "hacelo", PrdHash: "a3f9c1", ConstitucionSellada: true,
				}
			},
			quiero: []string{".docs/prd.md", ".docs/constitucion.md"},
		},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p := nuevo(t)
			c.armar(p)
			txt := p.texto()

			for _, q := range c.quiero {
				if !strings.Contains(txt, q) {
					t.Errorf("el sobre no trae %q:\n%s", q, txt)
				}
			}
			for _, n := range c.noVa {
				if strings.Contains(txt, n) {
					t.Errorf("el sobre trae %q y no debería:\n%s", n, txt)
				}
			}
		})
	}
}

// El ⑩ es el único que mira todas las historias juntas: tiene que agruparlas.
func TestSobreDelRoadmapTraeTodasLasHistorias(t *testing.T) {
	p := nuevo(t)
	p.e.Producto = estado.Producto{
		BriefSellado: "hacelo", PrdHash: "x", ConstitucionSellada: true, BacklogVisto: true,
	}
	p.archivo(".docs/backlog/us-1.md", "").archivo(".docs/backlog/us-2.md", "")

	txt := p.texto()
	for _, q := range []string{"us-1.md", "us-2.md"} {
		if !strings.Contains(txt, q) {
			t.Errorf("falta %q:\n%s", q, txt)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Feature
// ────────────────────────────────────────────────────────────────────────────

func TestSobreDePlanificacion(t *testing.T) {
	txt := nuevo(t).enFeature(estado.Planificacion).texto()

	// El manual, las historias a resolver, y la memoria de lo anterior.
	for _, q := range []string{"constitucion.md", "us-1.md", "us-3.md", "Aprendizajes"} {
		if !strings.Contains(txt, q) {
			t.Errorf("falta %q:\n%s", q, txt)
		}
	}
}

// ESTE es el test que protege la decisión más fina del sobre: el implementador
// NO ve decision.md. Le sirve la restricción, no la alternativa — con las tres
// opciones adentro, el que mezcla es él.
func TestElImplementadorNoVeDecisionMd(t *testing.T) {
	p := nuevo(t).enFeature(estado.Implementar)
	p.e.Features["f-1"].Lotes = []estado.Lote{{Lote: 1}}
	p.archivo(filepath.Join(carpetaF1, "decision.md"), "A, B y C")

	txt := p.texto()
	if strings.Contains(txt, "decision.md") {
		t.Errorf("el sobre del ⑱ trae decision.md:\n%s", txt)
	}
	if !strings.Contains(txt, "spec-design.md") {
		t.Errorf("el sobre del ⑱ no trae la spec:\n%s", txt)
	}
}

// El lote se embebe filtrado porque no hay forma de apuntar a "el lote 2 de
// este JSON" con una ruta.
func TestElSobreDelLoteTraeSoloSuLote(t *testing.T) {
	p := nuevo(t).enFeature(estado.Implementar)
	p.e.Features["f-1"].Lotes = []estado.Lote{{Lote: 1}, {Lote: 2}}
	p.archivo(filepath.Join(carpetaF1, "tareas.json"), `{"tareas":[
		{"id":"t-1","lote":1,"descripcion":"la del lote uno","tests":["a_test.go::TestUno"]},
		{"id":"t-9","lote":2,"descripcion":"la del lote dos","tests":["b_test.go::TestDos"]}
	]}`)

	txt := p.texto()
	if !strings.Contains(txt, "la del lote uno") {
		t.Errorf("no trae la tarea del lote 1:\n%s", txt)
	}
	if strings.Contains(txt, "la del lote dos") {
		t.Errorf("trae la tarea del lote 2, y va el 1:\n%s", txt)
	}
	// Los tests planificados tienen que viajar: son lo que el ⑲ va a escribir.
	if !strings.Contains(txt, "a_test.go::TestUno") {
		t.Errorf("no trae los tests planificados:\n%s", txt)
	}
}

// El lote actual es el primero sin commit, igual que en todos lados.
func TestElSobreSigueAlLoteActual(t *testing.T) {
	c := "4a8f21"
	p := nuevo(t).enFeature(estado.Implementar)
	p.e.Features["f-1"].Lotes = []estado.Lote{{Lote: 1, Commit: &c}, {Lote: 2}}
	p.archivo(filepath.Join(carpetaF1, "tareas.json"), `{"tareas":[
		{"id":"t-1","lote":1,"descripcion":"ya commiteada"},
		{"id":"t-9","lote":2,"descripcion":"la que toca"}
	]}`)

	txt := p.texto()
	if !strings.Contains(txt, "la que toca") || strings.Contains(txt, "ya commiteada") {
		t.Errorf("el sobre no siguió al lote actual:\n%s", txt)
	}
}

// La doc del ㉓ tiene dos mitades y sólo una sale del código. El sobre tiene
// que traer las dos fuentes.
func TestSobreDeCierreTraeLasDosMitades(t *testing.T) {
	txt := nuevo(t).enFeature(estado.Cierre).texto()

	if !strings.Contains(txt, "TÉCNICA") {
		t.Errorf("falta la mitad técnica:\n%s", txt)
	}
	if !strings.Contains(txt, "FUNCIONAL") || !strings.Contains(txt, "us-1.md") {
		t.Errorf("falta la mitad funcional, que no sale del código:\n%s", txt)
	}
}

// El revisor tiene que opinar sobre TODOS los criterios, así que los us-# van
// enteros: es el mecanismo con el que sf después cuenta y no lo deja pasar.
func TestSobreDeRevisionTraeLosCriterios(t *testing.T) {
	txt := nuevo(t).enFeature(estado.Revision).texto()

	for _, q := range []string{"us-1.md", "us-3.md", "spec-design.md"} {
		if !strings.Contains(txt, q) {
			t.Errorf("falta %q:\n%s", q, txt)
		}
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Lo que falta se DICE, no se omite
// ────────────────────────────────────────────────────────────────────────────

// Un sobre al que le falta una parte tiene que decirlo. Si la parte
// simplemente no apareciera, el que trabaja creería que no hacía falta — y ése
// es exactamente el mecanismo del dolor #7.
func TestLoQueFaltaSeDice(t *testing.T) {
	// El proyecto de prueba no es un repo git, así que el diff no se puede.
	txt := nuevo(t).enFeature(estado.Revision).texto()

	if !strings.Contains(txt, "⚠") {
		t.Errorf("no avisó de ninguna parte faltante:\n%s", txt)
	}
	if !strings.Contains(txt, "git") {
		t.Errorf("no dijo que falta el diff:\n%s", txt)
	}
	// Los mutantes todavía no están, y también tiene que decirse.
	if !strings.Contains(txt, "mutacion") {
		t.Errorf("no avisó que falta la corrida de mutantes:\n%s", txt)
	}
}

func TestSinTareasJsonAvisaEnVezDeRomper(t *testing.T) {
	p := nuevo(t).enFeature(estado.Implementar)
	p.e.Features["f-1"].Lotes = []estado.Lote{{Lote: 1}}

	txt := p.texto()
	if !strings.Contains(txt, "⚠") {
		t.Errorf("sin tareas.json no avisó:\n%s", txt)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// --completo: el caso del que no tiene manos
// ────────────────────────────────────────────────────────────────────────────

// Cuando el que trabaja es un modelo por consola sin shell, no puede abrir
// archivos: el orquestador corre esto y le pega la salida en el prompt.
func TestCompletoEmbebeElContenido(t *testing.T) {
	p := nuevo(t)
	p.e.Producto.BriefSellado = "hacelo"
	p.archivo(".docs/brief.md", "el contenido del brief")

	s, err := Armar(p.raiz, p.e, p.r)
	if err != nil {
		t.Fatal(err)
	}

	if txt := s.Texto(p.raiz, false); strings.Contains(txt, "el contenido del brief") {
		t.Error("sin --completo embebió el contenido: tendría que listar la ruta")
	}
	if txt := s.Texto(p.raiz, true); !strings.Contains(txt, "el contenido del brief") {
		t.Error("con --completo no embebió el contenido")
	}
}

func TestCompletoAvisaSiUnArchivoNoEstá(t *testing.T) {
	p := nuevo(t)
	p.e.Producto.BriefSellado = "hacelo" // el sobre pide brief.md, que no existe

	s, err := Armar(p.raiz, p.e, p.r)
	if err != nil {
		t.Fatal(err)
	}
	if txt := s.Texto(p.raiz, true); !strings.Contains(txt, "⚠") {
		t.Errorf("no avisó que el archivo no se pudo leer:\n%s", txt)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Errores
// ────────────────────────────────────────────────────────────────────────────

func TestSinFeatureEnCursoDaError(t *testing.T) {
	p := nuevo(t).conRoadmap()
	p.e.Producto = estado.Producto{
		BriefSellado: "hacelo", PrdHash: "x", ConstitucionSellada: true, BacklogVisto: true,
	}

	_, err := Armar(p.raiz, p.e, p.r)
	if err == nil {
		t.Fatal("sin feature en curso no dio error")
	}
	if !strings.Contains(err.Error(), "sf take") {
		t.Errorf("el error no dice cómo salir: %v", err)
	}
}
