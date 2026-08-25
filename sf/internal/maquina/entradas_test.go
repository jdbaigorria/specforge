package maquina

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/historia"
)

// ────────────────────────────────────────────────────────────────────────────
// sf new — el backlog es el embudo
// ────────────────────────────────────────────────────────────────────────────

func TestNewCreaLaProximaHistoria(t *testing.T) {
	p := nuevo(t).productoListo()
	p.conArchivoConTexto(docs.Historia("us-1"), "---\nid: us-1\n---\n- **CA-1** — x\n")
	p.conArchivoConTexto(docs.Historia("us-3"), "---\nid: us-3\n---\n- **CA-1** — y\n")

	ef := Nueva(p.raiz, p.e, "que sf soporte proyectos brownfield")
	if !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}

	// El id sale del MÁXIMO + 1, no de la cantidad + 1: una historia archivada
	// puede haber dejado un hueco, y reusar su id pisaría referencias viejas.
	b, err := os.ReadFile(filepath.Join(p.raiz, docs.Historia("us-4")))
	if err != nil {
		t.Fatalf("no creó us-4: %v", err)
	}
	if !strings.Contains(string(b), "id: us-4") {
		t.Errorf("el frontmatter no trae el id:\n%s", b)
	}
	// Lo que Javier tipeó no se pierde: va al cuerpo para que el pinponeo
	// arranque de ahí en vez de hacerle repetir lo que ya dijo.
	if !strings.Contains(string(b), "brownfield") {
		t.Errorf("perdió el texto que se le pasó:\n%s", b)
	}
	// Y arranca como `us`: si es un bug, lo cambia el que pinponea (H20).
	if !strings.Contains(string(b), "tipo: us") {
		t.Errorf("no arrancó como us:\n%s", b)
	}
}

// Volver a abrir la ⏸ del ⑨ es lo que manda al pinponeo en vez de seguir con el
// ciclo de feature como si nada hubiera entrado.
func TestNewReabreElBacklog(t *testing.T) {
	p := nuevo(t).productoListo()

	if !p.e.Producto.BacklogVisto {
		t.Fatal("el andamio no dejó el backlog visto")
	}
	if ef := Nueva(p.raiz, p.e, "algo"); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}
	if p.e.Producto.BacklogVisto {
		t.Error("no reabrió la ⏸ del ⑨: sf next seguiría con el ciclo")
	}
}

// `sf new` es para agregarle algo a un producto que YA existe. En uno nuevo la
// entrada es el brief, y decirlo es más útil que crear una historia suelta.
func TestNewNecesitaUnProductoArmado(t *testing.T) {
	p := nuevo(t)

	if ef := Nueva(p.raiz, p.e, "algo"); ef.Pasa() {
		t.Error("creó una historia en un producto sin constitución")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Las historias huérfanas
// ────────────────────────────────────────────────────────────────────────────

// Una historia que entró por `sf new` no está en ninguna feature: sin este
// chequeo, sf seguiría con el ciclo y esa historia no la implementaría nadie.
func TestUnaHistoriaSinFeatureMandaAlRoadmap(t *testing.T) {
	p := nuevo(t).productoListo()
	// El roadmap del andamio tiene us-1 y us-2; ésta queda suelta.
	p.conArchivoConTexto(docs.Historia("us-1"), "---\nid: us-1\n---\n- **CA-1** — x\n")
	p.conArchivoConTexto(docs.Historia("us-8"), "---\nid: us-8\n---\n- **CA-1** — nueva\n")

	i := p.next()
	if i.Estado != "roadmap" {
		t.Fatalf("estado %q, quería roadmap", i.Estado)
	}
	if !strings.Contains(i.Mensaje, "us-8") {
		t.Errorf("no dijo cuál quedó suelta: %q", i.Mensaje)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El camino corto — un campo que saltea dos estados
// ────────────────────────────────────────────────────────────────────────────

// El ruteo lo hace `tipo: bug` y nada más. No hay carril paralelo: hay estados
// salteados, y el rastro no se pierde porque el bug igual entró por el backlog.
func TestUnBugEntraDirectoAImplementar(t *testing.T) {
	p := nuevo(t).productoListo()
	p.conArchivoConTexto(docs.Historia("us-1"), "---\ntipo: bug\nid: us-1\n---\n- **CA-1** — no rompe\n")
	p.conArchivoConTexto(docs.Historia("us-2"), "---\nid: us-2\n---\n- **CA-1** — x\n")
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Planificacion}

	if i := p.next(); i.Estado != estado.Implementar {
		t.Errorf("estado %q, quería implementar: es un bug", i.Estado)
	}

	// Y `sf next` sigue siendo consulta pura: no escribió el salteo.
	if p.e.Features["f-1"].Estado != estado.Planificacion {
		t.Error("sf next movió el estado, y tiene que ser consulta pura")
	}
}

// Se exige que TODAS sean bugs: saltear la planificación de una feature que
// mezcla un bug con historias nuevas dejaría esas historias sin diseño.
func TestUnaFeatureMixtaNoSalteaLaPlanificacion(t *testing.T) {
	p := nuevo(t).productoListo()
	p.conArchivoConTexto(docs.Historia("us-1"), "---\ntipo: bug\nid: us-1\n---\n- **CA-1** — x\n")
	p.conArchivoConTexto(docs.Historia("us-2"), "---\nid: us-2\n---\n- **CA-1** — y\n")

	// La f-2 del andamio junta us-2, que no es bug.
	p.e.FeatureActual = "f-2"
	p.e.Features["f-2"] = &estado.Feature{Estado: estado.Planificacion}

	if i := p.next(); i.Estado != estado.Planificacion {
		t.Errorf("estado %q, quería planificacion", i.Estado)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Los ids
// ────────────────────────────────────────────────────────────────────────────

// us-10 va DESPUÉS de us-9. Un sort de strings los pone al revés, y eso no se
// nota hasta la décima historia.
func TestLosIdsSeOrdenanPorNumeroNoPorTexto(t *testing.T) {
	p := nuevo(t)
	for _, id := range []string{"us-1", "us-9", "us-10", "us-2"} {
		p.conArchivoConTexto(docs.Historia(id), "---\nid: "+id+"\n---\n")
	}

	ids := historia.Ids(p.raiz)
	quiero := []string{"us-1", "us-2", "us-9", "us-10"}
	for i := range quiero {
		if ids[i] != quiero[i] {
			t.Fatalf("Ids() = %v, quería %v", ids, quiero)
		}
	}
	if got := historia.ProximoID(p.raiz); got != "us-11" {
		t.Errorf("ProximoID() = %q, quería us-11", got)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// `sf new` tiene que LLEVAR a sfp-backlog, no sólo reabrir la ⏸
// ────────────────────────────────────────────────────────────────────────────

// El propio código decía que `Nueva` reabre la ⏸ "para que sf next mande al
// pinponeo", y no lo hacía: el checkpoint preguntaba "¿hay historias?" y con un
// producto en marcha la respuesta es siempre sí. Así que salía la ⏸ y lo que
// quedaba para aprobar era el esqueleto con el título vacío.
func TestDespuesDeNewElNueveMandaACompletarLaHistoria(t *testing.T) {
	p := nuevo(t).productoListo()
	p.conArchivoConTexto(docs.Historia("us-1"),
		"---\nid: us-1\ntitulo: la primera\n---\n## Criterios\n- **CA-1** — algo\n")

	if ef := Nueva(p.raiz, p.e, "el login rompe con mayúsculas"); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}

	i := p.next()
	if i.Tipo != Trabajar {
		t.Fatalf("tipo %v, quería Trabajar: us-2 es un esqueleto", i.Tipo)
	}
	if i.Estado != "backlog" || i.Skill != "sfp-backlog" {
		t.Errorf("estado %q skill %q, quería backlog/sfp-backlog", i.Estado, i.Skill)
	}
	// Y dice CUÁL, que es la diferencia entre una instrucción y una consigna:
	// "partí el PRD" con veinte historias ya escritas es lo contrario de útil.
	if !strings.Contains(i.Mensaje, "us-2") {
		t.Errorf("no dijo cuál completar: %q", i.Mensaje)
	}
}

// Y con el backlog completo vuelve la ⏸: el arreglo no puede dejar la máquina
// pidiendo pinponeo para siempre.
func TestConElBacklogCompletoVuelveLaParadaBarata(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.Producto.BacklogVisto = false
	p.conArchivoConTexto(docs.Historia("us-1"),
		"---\nid: us-1\ntitulo: la primera\n---\n## Criterios\n- **CA-1** — algo\n")

	if i := p.next(); i.Tipo != Barata {
		t.Errorf("tipo %v, quería Barata: no hay nada que completar", i.Tipo)
	}
}

// La primera vuelta sigue diciendo "partí el PRD": ahí faltan TODAS, y nombrar
// veinte historias que no existen no ayuda a nadie.
func TestLaPrimeraVueltaDelNueveSigueSiendoPartirElPRD(t *testing.T) {
	p := nuevo(t)
	p.e.Producto = estado.Producto{
		BriefSellado: "hacelo", PrdHash: "x", ConstitucionSellada: true,
	}

	i := p.next()
	if i.Estado != "backlog" || i.Tipo != Trabajar {
		t.Fatalf("estado %q tipo %v", i.Estado, i.Tipo)
	}
	if !strings.Contains(i.Mensaje, "parte el PRD") {
		t.Errorf("mensaje %q, quería el de partir el PRD", i.Mensaje)
	}
}
