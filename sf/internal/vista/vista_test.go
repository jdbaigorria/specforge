package vista

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
)

func armar(t *testing.T) (string, *estado.Estado, *roadmap.Roadmap) {
	t.Helper()
	raiz := t.TempDir()

	escribir := func(rel, texto string) {
		ruta := filepath.Join(raiz, rel)
		if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(ruta, []byte(texto), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	escribir(docs.Historia("us-1"), "---\nid: us-1\ntitulo: \"inicializar el proyecto\"\n---\n")
	escribir(docs.Historia("us-2"), "---\nid: us-2\ntitulo: \"el sobre por estado\"\n---\n")
	escribir(roadmap.Archivo, `{"features":[
		{"id":"f-1","slug":"nucleo","nombre":"núcleo del CLI","orden":1,"historias":["us-1"]},
		{"id":"f-2","slug":"verif","nombre":"verificación","orden":2,"historias":["us-2"]}
	]}`)

	r, err := roadmap.Leer(raiz)
	if err != nil {
		t.Fatal(err)
	}

	e := &estado.Estado{
		Producto: estado.Producto{
			BriefSellado: "hacelo", PrdHash: "a3f9c1",
			ConstitucionSellada: true, BacklogVisto: true,
		},
		FeatureActual: "f-2",
		Features: map[string]*estado.Feature{
			"f-1": {Estado: estado.Cerrada},
			"f-2": {
				Estado: estado.Implementar, Modelo: "deepseek",
				Lotes: []estado.Lote{{Lote: 1, Commit: ptr("abc")}, {Lote: 2}, {Lote: 3}},
			},
		},
	}
	return raiz, e, r
}

func ptr(s string) *string { return &s }

// La vista tiene que decir de un vistazo qué está cerrado, qué se está tocando
// y qué falta — sin que haya un archivo que mantener con eso.
func TestLaVistaMuestraLoEsencial(t *testing.T) {
	txt := Estado(armar(t))

	for _, q := range []string{
		"✓", "núcleo del CLI", // lo cerrado
		"▸", "verificación", "lote 2 de 3", "deepseek", // lo que se está tocando
		"us-1", "inicializar el proyecto", // los títulos salen de los us-#
	} {
		if !strings.Contains(txt, q) {
			t.Errorf("la vista no muestra %q:\n%s", q, txt)
		}
	}
}

// Los títulos NO están en el roadmap: ahí sólo hay ids, para que no pueda
// contradecir al us-#. Ese diseño se paga acá, leyendo un archivo por historia.
func TestLosTitulosSalenDeLosUsNoDelRoadmap(t *testing.T) {
	raiz, e, r := armar(t)

	// Se borra el us-2: el título tiene que desaparecer, no salir del roadmap.
	if err := os.Remove(filepath.Join(raiz, docs.Historia("us-2"))); err != nil {
		t.Fatal(err)
	}

	txt := Estado(raiz, e, r)
	if strings.Contains(txt, "el sobre por estado") {
		t.Error("el título salió de algún lado que no es el us-#")
	}
	if !strings.Contains(txt, "us-2") {
		t.Error("dejó de mostrar la historia entera en vez de sólo el título")
	}
}

// El contador de intentos sólo aparece cuando hay fracasos: en cero es ruido, y
// en dos es la señal de que ME TRABÉ está cerca.
func TestElContadorSoloApareceSiHayFracasos(t *testing.T) {
	raiz, e, r := armar(t)

	if strings.Contains(Estado(raiz, e, r), "intentos") {
		t.Error("mostró el contador en cero")
	}

	e.Features["f-2"].IntentosFallidos = 2
	if !strings.Contains(Estado(raiz, e, r), "⚠ 2 intentos") {
		t.Error("no mostró el contador con 2 fracasos")
	}
}

// Un producto ya armado no necesita que le recuerden que el brief está sellado.
// Uno a medias, sí.
func TestElBloqueDeProductoSoloApareceMientrasFalta(t *testing.T) {
	raiz, e, r := armar(t)

	if strings.Contains(Estado(raiz, e, r), "producto") {
		t.Error("mostró el bloque de producto con todo sellado")
	}

	e.Producto.ConstitucionSellada = false
	txt := Estado(raiz, e, r)
	if !strings.Contains(txt, "constitución") {
		t.Errorf("no mostró lo que falta del producto:\n%s", txt)
	}
}

func TestSinRoadmapLoDice(t *testing.T) {
	raiz, e, _ := armar(t)

	if txt := Estado(raiz, e, nil); !strings.Contains(txt, "Todavía no hay roadmap") {
		t.Errorf("no avisó que falta el ⑩:\n%s", txt)
	}
}

// Una feature que no arrancó no está en el mapa del estado (R6: se deduce de la
// ausencia). La vista tiene que mostrarla igual.
func TestUnaFeatureQueNoArrancoSeMuestraComoPendiente(t *testing.T) {
	raiz, e, r := armar(t)
	delete(e.Features, "f-2")

	if txt := Estado(raiz, e, r); !strings.Contains(txt, "pendiente") {
		t.Errorf("no mostró la feature sin empezar:\n%s", txt)
	}
}
