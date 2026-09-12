package maquina

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/constitucion"
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
	"github.com/jdbaigorria/specforge/sf/internal/tareas"
)

// ────────────────────────────────────────────────────────────────────────────
// Andamio de los tests
// ────────────────────────────────────────────────────────────────────────────
//
// Los tests de este paquete arman un proyecto de verdad en un directorio
// temporal: archivos en disco, no mocks. Es a propósito.
//
// La mitad de las decisiones de la máquina son "¿existe este archivo?", y un
// mock de filesystem las haría pasar sin probar lo único que importa: que el
// nombre del archivo y la ruta que arma Carpeta() coincidan de verdad. Un test
// que miente sobre el disco es exactamente el dolor #8 con otro disfraz.

type proyecto struct {
	raiz string
	e    *estado.Estado
	r    *roadmap.Roadmap
	g    *global.Config
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

// conArchivo crea un archivo vacío. El contenido no importa: la máquina sólo
// pregunta si está.
func (p *proyecto) conArchivo(rel string) *proyecto {
	p.t.Helper()
	ruta := filepath.Join(p.raiz, rel)
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		p.t.Fatal(err)
	}
	if err := os.WriteFile(ruta, nil, 0o644); err != nil {
		p.t.Fatal(err)
	}
	return p
}

// conHistoria escribe un us-# COMPLETO: con título y criterios con texto.
//
// Existe porque `conArchivo` deja el archivo vacío, y desde que el checkpoint
// del ⑨ pregunta "¿qué falta completar?" en vez de "¿hay algo?", un us-# vacío
// ya no es una historia — es un esqueleto que manda a `sfp-backlog`.
func (p *proyecto) conHistoria(id string) *proyecto {
	p.t.Helper()
	return p.conArchivoConTexto(docs.Historia(id),
		"---\nid: "+id+"\ntitulo: la historia "+id+"\n---\n"+
			"## Criterios de aceptación\n- **CA-1** — hace lo que tiene que hacer\n")
}

// conRoadmap escribe un roadmap de dos features y lo deja cargado.
func (p *proyecto) conRoadmap() *proyecto {
	p.t.Helper()
	p.conArchivoConTexto(roadmap.Archivo, `{"features": [
		{"id": "f-1", "slug": "nucleo", "nombre": "núcleo", "orden": 1, "historias": ["us-1"]},
		{"id": "f-2", "slug": "verif",  "nombre": "verificación", "orden": 2, "historias": ["us-2"]}
	]}`)
	r, err := roadmap.Leer(p.raiz)
	if err != nil {
		p.t.Fatal(err)
	}
	p.r = r
	return p
}

func (p *proyecto) conArchivoConTexto(rel, texto string) *proyecto {
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

// productoListo salta los cinco estados de producto: sella todo y deja el
// roadmap. Es el punto de partida de todos los tests del ciclo de feature.
func (p *proyecto) productoListo() *proyecto {
	p.e.Producto = estado.Producto{
		BriefSellado:        "hacelo",
		PrdHash:             "a3f9c1",
		ConstitucionSellada: true,
		BacklogVisto:        true,
	}
	return p.conRoadmap()
}

func (p *proyecto) next() Instruccion {
	return Siguiente(p.raiz, p.e, p.r, p.g)
}

// conCatalogo declara un catálogo para este proyecto de prueba.
//
// Los tests que NO lo llaman corren con g == nil, que es el caso de "todavía no
// se corrió sf install" — y tiene que seguir funcionando: no tener el catálogo no
// puede impedir trabajar.
func (p *proyecto) conCatalogo(h string, cat global.Catalogo) *proyecto {
	parandoEn(p.t, h)
	p.g = &global.Config{Harness: h, Harnesses: map[string]global.Catalogo{h: cat}}
	return p
}

// conTresPerfiles es el catálogo que usan casi todos los tests.
//
// Los alias son deliberadamente genéricos —`caro`, `medio`, `barato`— y no
// nombres de modelos reales. Es el mismo invariante que en `global`: si un test
// dijera `opus`, el día que alguien lo lea creería que sf conoce ese nombre.
func (p *proyecto) conTresPerfiles() *proyecto {
	return p.conCatalogo("claude-code", global.Catalogo{
		global.Razonar: {{Alias: "caro", ID: "id-de-caro", Via: global.Subagente}},
		global.Construir: {{Alias: "medio", ID: "id-de-medio", Via: global.Subagente},
			{Alias: "barato", ID: "id-de-barato", Via: global.Subagente}},
	})
}

// ────────────────────────────────────────────────────────────────────────────
// Los cinco estados de producto
// ────────────────────────────────────────────────────────────────────────────

// El orden de los cinco no es negociable: no se puede planificar sin
// constitución, ni partir en historias un PRD que no existe. Este test recorre
// la secuencia entera y es el que se rompe si alguien reordena los if.
func TestElOrdenDeLosCincoDeProducto(t *testing.T) {
	p := nuevo(t)

	// ① sin nada: el brief, y es el único que conversa.
	i := p.next()
	if i.Estado != "brief" || i.Tipo != Trabajar {
		t.Fatalf("con el proyecto vacío: estado %q tipo %v, quería brief/Trabajar", i.Estado, i.Tipo)
	}
	if i.Via != "vos" {
		t.Errorf("via = %q, quería vos: el brief es un pinponeo, no una fila", i.Via)
	}

	// ⑥ el brief existe pero no está sellado: 🛑, y lo sella Javier.
	p.conArchivo(".docs/brief.md")
	if i := p.next(); i.Tipo != Para {
		t.Fatalf("con brief.md escrito: tipo %v, quería Para (el ⑥)", i.Tipo)
	}

	// ⑦ sellado: el PRD.
	p.e.Producto.BriefSellado = "hacelo"
	if i := p.next(); i.Estado != "prd" {
		t.Fatalf("después del sello: %q, quería prd", i.Estado)
	}

	// ⑧ con hash de PRD: la constitución.
	p.e.Producto.PrdHash = "a3f9c1"
	if i := p.next(); i.Estado != "constitucion" {
		t.Fatalf("con prd_hash: %q, quería constitucion", i.Estado)
	}

	// ⑧ escrita pero sin sellar: 🛑.
	p.conArchivo(".docs/constitucion.md")
	if i := p.next(); i.Tipo != Para {
		t.Fatalf("con constitucion.md escrita: tipo %v, quería Para (el ⑧)", i.Tipo)
	}

	// ⑨ sellada: el backlog.
	p.e.Producto.ConstitucionSellada = true
	if i := p.next(); i.Estado != "backlog" {
		t.Fatalf("con la constitución sellada: %q, quería backlog", i.Estado)
	}

	// ⑨ con historias COMPLETAS: ⏸, la parada barata.
	p.conHistoria("us-1")
	if i := p.next(); i.Tipo != Barata {
		t.Fatalf("con us-1.md: tipo %v, quería Barata (la ⏸ del ⑨)", i.Tipo)
	}

	// ⑩ visto: el roadmap.
	p.e.Producto.BacklogVisto = true
	if i := p.next(); i.Estado != "roadmap" {
		t.Fatalf("con el backlog visto: %q, quería roadmap", i.Estado)
	}
}

// Este es el test del campo que apareció construyendo: sin BacklogVisto, la ⏸
// del ⑨ se repite para siempre porque un enter no deja rastro.
func TestLaParadaBaratasDelBacklogNoSeCuelga(t *testing.T) {
	p := nuevo(t)
	p.e.Producto = estado.Producto{BriefSellado: "hacelo", PrdHash: "x", ConstitucionSellada: true}
	p.conHistoria("us-1")

	if i := p.next(); i.Tipo != Barata {
		t.Fatalf("tipo %v, quería Barata", i.Tipo)
	}

	// El enter de Javier (sf approve) marca el campo. Sin él, next volvería a
	// decir ⏸ eternamente.
	p.e.Producto.BacklogVisto = true
	if i := p.next(); i.Tipo == Barata {
		t.Error("después del enter sigue diciendo ⏸: el bucle no sale")
	}
}

// El brief es la única parte del flujo que puede terminar sin construir nada, y
// eso no es un caso raro: es el valor del ⑥.
func TestBriefNegativoTerminaElFlujo(t *testing.T) {
	p := nuevo(t)
	p.e.Producto.BriefSellado = "no-lo-hagas"

	i := p.next()
	if i.Tipo != Fin {
		t.Errorf("con el brief en no-lo-hagas: tipo %v, quería Fin", i.Tipo)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El ⑪ — tomar de la cola
// ────────────────────────────────────────────────────────────────────────────

// sf next NO toma la feature sola: propone y deja el `sf take`. Es lo que lo
// mantiene como consulta pura — correrlo dos veces da lo mismo.
func TestSinFeatureActualProponeElTake(t *testing.T) {
	p := nuevo(t).productoListo()

	i := p.next()
	if i.Feature != "f-1" {
		t.Errorf("propuso %q, quería f-1 (la primera del roadmap)", i.Feature)
	}
	if len(i.Sugerido) == 0 || i.Sugerido[0] != "sf take f-1" {
		t.Errorf("sugirió %v, quería sf take f-1", i.Sugerido)
	}

	// Consulta pura: el estado no se tocó.
	if p.e.FeatureActual != "" {
		t.Errorf("sf next escribió feature_actual = %q", p.e.FeatureActual)
	}
}

func TestSaltaLasFeaturesCerradas(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Cerrada}

	if i := p.next(); i.Feature != "f-2" {
		t.Errorf("con f-1 cerrada propuso %q, quería f-2", i.Feature)
	}
}

func TestConTodoCerradoTermina(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Cerrada}
	p.e.Features["f-2"] = &estado.Feature{Estado: estado.Cerrada}

	if i := p.next(); i.Tipo != Fin {
		t.Errorf("tipo %v, quería Fin", i.Tipo)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// planificacion — los checkpoints deducidos
// ────────────────────────────────────────────────────────────────────────────

// Un subagente puede morirse a la mitad de los tres archivos. sf retoma sin que
// el estado.json sepa nada: mira qué hay en la carpeta.
func TestPlanificacionRetomaMirandoLosArchivos(t *testing.T) {
	carpeta := filepath.Join(".docs", "features", "f-1-nucleo")

	casos := []struct {
		nombre  string
		crear   []string
		enElMsg string
	}{
		{"sin nada", nil, "de cero"},
		{"con decision", []string{"decision.md"}, "desde la spec"},
		{"con decision y spec", []string{"decision.md", "spec-design.md"}, "desde las tareas"},
		// Con los tres archivos ya no alcanza con que existan: sf corre las
		// mismas compuertas que `sf done`. Como acá están vacíos, falla — y eso
		// es exactamente lo que tiene que pasar.
		{"con los tres vacíos", []string{"decision.md", "spec-design.md", "tareas.json"}, "pero falta"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p := nuevo(t).productoListo()
			p.e.FeatureActual = "f-1"
			p.e.Features["f-1"] = &estado.Feature{Estado: estado.Planificacion}
			for _, f := range c.crear {
				p.conArchivo(filepath.Join(carpeta, f))
			}

			i := p.next()
			if i.Estado != estado.Planificacion {
				t.Fatalf("estado %q, quería planificacion", i.Estado)
			}
			if !strings.Contains(i.Mensaje, c.enElMsg) {
				t.Errorf("mensaje %q, quería que mencione %q", i.Mensaje, c.enElMsg)
			}
		})
	}
}

// Con el plan COMPLETO y válido, sf next para: el ⑰ lo decide Javier.
//
// Es lo que distingue "falta trabajar" de "está listo, falta que lo mires", y
// se decide corriendo las compuertas — no con un campo en el estado.
func TestPlanCompletoParaEnElDiecisiete(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Planificacion}

	carpeta := filepath.Join(".docs", "features", "f-1-nucleo")
	p.conArchivoConTexto(filepath.Join(carpeta, "decision.md"), decisionCompleta)
	p.conArchivoConTexto(filepath.Join(carpeta, "spec-design.md"), "# spec\n")
	p.conArchivoConTexto(filepath.Join(carpeta, "tareas.json"),
		`{"tareas":[{"id":"t-1","lote":1,"satisface":["us-1/CA-1"],"tests":["a_test.go::TestX"]}]}`)
	p.conArchivoConTexto(".docs/backlog/us-1.md",
		"---\nid: us-1\n---\n## Criterios\n- **CA-1** — hace algo\n")

	i := p.next()
	if i.Tipo != Para {
		t.Fatalf("con el plan completo: tipo %v, quería Para (el ⑰)\n%s", i.Tipo, i.Mensaje)
	}
	if !slices.Contains(i.Sugerido, "sf approve") {
		t.Errorf("sugirió %v, quería sf approve", i.Sugerido)
	}
}

// El plan aprobado que espera en la cola es la puerta "otra feature" del ⑰.
// Volver a él es decisión de Javier, así que sf no lo retoma solo.
func TestPlanificadaEsperaUnTake(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Planificada}

	i := p.next()
	if i.Tipo != Para {
		t.Errorf("tipo %v, quería Para", i.Tipo)
	}
	if len(i.Sugerido) == 0 || i.Sugerido[0] != "sf take f-1" {
		t.Errorf("sugirió %v, quería sf take f-1", i.Sugerido)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// implementar — el rojo primero
// ────────────────────────────────────────────────────────────────────────────

func implementando1(t *testing.T, lotes []estado.Lote) *proyecto {
	t.Helper()
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Implementar, Lotes: lotes}
	return p
}

// Sin `rojo` no se implementa. Es la compuerta del medio, la joya del diseño:
// sf no deja pasar al verde sin haber visto el rojo con sus propios ojos.
func TestSinRojoMandaACorrerLoteStart(t *testing.T) {
	p := implementando1(t, []estado.Lote{{Lote: 1}, {Lote: 2}})

	i := p.next()
	if i.Lote != 1 || i.DeLotes != 2 {
		t.Errorf("lote %d de %d, quería 1 de 2", i.Lote, i.DeLotes)
	}
	if !slices.Contains(i.Sugerido, "sf lote start") {
		t.Errorf("sugirió %v, quería que incluya sf lote start", i.Sugerido)
	}
}

func TestConRojoMandaAImplementar(t *testing.T) {
	p := implementando1(t, []estado.Lote{{Lote: 1, Rojo: true}})

	i := p.next()
	if slices.Contains(i.Sugerido, "sf lote start") {
		t.Error("con el rojo confirmado sigue pidiendo sf lote start")
	}
	if !strings.Contains(i.Mensaje, "implementá") {
		t.Errorf("mensaje %q", i.Mensaje)
	}
}

// El lote actual es el primero SIN COMMIT, y no hay campo que lo diga.
func TestElLoteActualEsElPrimeroSinCommit(t *testing.T) {
	c := "4a8f21"
	p := implementando1(t, []estado.Lote{
		{Lote: 1, Rojo: true, Commit: &c},
		{Lote: 2, Rojo: true},
		{Lote: 3},
	})

	if i := p.next(); i.Lote != 2 {
		t.Errorf("lote %d, quería 2 (el 1 ya tiene commit)", i.Lote)
	}
}

func TestConTodosLosLotesCommiteadosFaltaCerrar(t *testing.T) {
	c := "4a8f21"
	p := implementando1(t, []estado.Lote{{Lote: 1, Rojo: true, Commit: &c}})

	i := p.next()
	if !slices.Contains(i.Sugerido, "sf done") {
		t.Errorf("sugirió %v, quería sf done", i.Sugerido)
	}
}

// El modelo que Javier subió en el ⑳ tiene que ganarle al default del mapa: esa
// decisión no está escrita en ningún archivo.
func TestElModeloDelEstadoPisaAlDefault(t *testing.T) {
	p := implementando1(t, []estado.Lote{{Lote: 1, Rojo: true}}).conTresPerfiles()
	p.e.Features["f-1"].Modelo = "caro"

	i := p.next()
	if i.Modelo != "id-de-caro" {
		t.Errorf("modelo %q, quería el id de caro (el que Javier subió)", i.Modelo)
	}
	// Y el perfil sigue diciendo lo que el PASO pide, que no es lo mismo que lo
	// que se le dio. Las dos cosas viajan juntas a propósito.
	if i.Perfil != global.Construir {
		t.Errorf("perfil %q, quería construir: el ⑳ cambia el modelo, no lo que el paso pide", i.Perfil)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// ME TRABÉ — el freno del bucle
// ────────────────────────────────────────────────────────────────────────────

// Se mira ANTES que el estado: si hace rato que no avanza, no importa qué
// falta. Ponerlo después dejaría que sf siga proponiendo trabajo mientras el
// bucle patina.
func TestMeTrabeGanaSobreElEstado(t *testing.T) {
	p := implementando1(t, []estado.Lote{{Lote: 1, Rojo: true}})
	p.e.Features["f-1"].IntentosFallidos = TopeIntentos
	p.e.Features["f-1"].Modelo = "deepseek"

	i := p.next()
	if i.Tipo != MeTrabe {
		t.Fatalf("tipo %v, quería MeTrabe", i.Tipo)
	}
	if !strings.Contains(i.Mensaje, "deepseek") {
		t.Errorf("el mensaje no dice con qué modelo venía fallando: %q", i.Mensaje)
	}
	// Las tres salidas: subir el modelo, descartar un hallazgo, o entrar vos
	// (que no necesita comando).
	if !slices.Contains(i.Sugerido, "sf model <alias>") {
		t.Errorf("sugirió %v, quería sf model", i.Sugerido)
	}
}

func TestPorDebajoDelTopeSigueTrabajando(t *testing.T) {
	p := implementando1(t, []estado.Lote{{Lote: 1, Rojo: true}})
	p.e.Features["f-1"].IntentosFallidos = TopeIntentos - 1

	if i := p.next(); i.Tipo != Trabajar {
		t.Errorf("con %d intentos: tipo %v, quería Trabajar", TopeIntentos-1, i.Tipo)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// revision y cierre
// ────────────────────────────────────────────────────────────────────────────

func TestRevisionYCierre(t *testing.T) {
	for _, est := range []string{estado.Revision, estado.Cierre} {
		t.Run(est, func(t *testing.T) {
			p := nuevo(t).productoListo()
			p.e.FeatureActual = "f-1"
			p.e.Features["f-1"] = &estado.Feature{Estado: est}

			i := p.next()
			if i.Estado != est || i.Tipo != Trabajar {
				t.Fatalf("estado %q tipo %v", i.Estado, i.Tipo)
			}
			if i.Skill == "" {
				t.Error("no devolvió skill, y es la mitad de H1")
			}
		})
	}
}

// La revisión la hace un modelo grande: es el ㉑, "revisá con uno grande".
func TestRevisionVaConModeloGrande(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Revision}

	if i := p.conTresPerfiles().next(); i.Perfil != global.Razonar {
		t.Errorf("perfil %q, quería razonar: el ㉑ tiene que encontrar lo que no está", i.Perfil)
	}
}

// Una feature cerrada que quedó como actual no traba la cola: sigue la próxima.
func TestFeatureCerradaSigueLaCola(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Cerrada}

	if i := p.next(); i.Feature != "f-2" {
		t.Errorf("propuso %q, quería f-2", i.Feature)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Todo estado con trabajo tiene que decir skill y modelo. Si no, el orquestador
// tiene que saberlo — y eso es exactamente lo que H1 vino a evitar.
// ────────────────────────────────────────────────────────────────────────────

func TestTodoTrabajoTraeSkillYModelo(t *testing.T) {
	for est := range skills {
		t.Run(est, func(t *testing.T) {
			if skills[est] == "" {
				t.Errorf("el estado %q no tiene skill en el mapa", est)
			}
			if perfil(est) == "" {
				t.Errorf("el estado %q no resuelve perfil", est)
			}
			// Sin catálogo (nadie corrió sf install) el via igual resuelve: es
			// lo que permite trabajar antes de instalar nada.
			if m, _, hay := lanzar(est, perfil(est), perfil(est), nil); !hay || m.Via == "" {
				t.Errorf("el estado %q no resuelve via sin catálogo", est)
			}
		})
	}
}

// Con la doc y el journal escritos, sf next para en la ⏸ del ㉓ en vez de
// mandarte a escribir algo que ya está.
func TestCierreCompletoParaEnLaPausa(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Cierre}

	carpeta := filepath.Join(".docs", "features", "f-1-nucleo")
	if i := p.next(); i.Tipo != Trabajar {
		t.Fatalf("sin la doc: tipo %v, quería Trabajar", i.Tipo)
	}

	p.conArchivo(filepath.Join(carpeta, "doc.md"))
	p.conArchivo(filepath.Join(carpeta, "journal.md"))

	i := p.next()
	if i.Tipo != Barata {
		t.Fatalf("con la doc y el journal: tipo %v, quería Barata (la ⏸ del ㉓)", i.Tipo)
	}
	if !slices.Contains(i.Sugerido, "sf approve") {
		t.Errorf("sugirió %v, quería sf approve", i.Sugerido)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// La regla de prefijos, como test
// ────────────────────────────────────────────────────────────────────────────
//
// Este test existe porque el mapa YA SE ROMPIÓ una vez: se escribió durante la
// construcción preguntando "¿quién sabe hacer este trabajo?", y la respuesta a
// esa pregunta es casi siempre un utilitario —porque el método vive ahí—, así
// que seis de nueve entradas terminaron mal (skills.md §2).
//
// La regresión es silenciosa y ésa es la parte fea: compila, los tests pasan,
// `sf next` contesta. Se rompe recién EN PRODUCCIÓN, porque un `sfx-` no llama
// a `sf done` y el estado no se mueve nunca.
func TestElMapaRespetaLaReglaDePrefijos(t *testing.T) {
	producto := []string{"brief", "prd", "constitucion", "backlog", "roadmap"}
	feature := []string{estado.Planificacion, estado.Implementar, estado.Revision, estado.Cierre}

	for _, e := range producto {
		s, hay := skills[e]
		if !hay {
			t.Errorf("el estado %q no tiene skill", e)
			continue
		}
		if !strings.HasPrefix(s, "sfp-") {
			t.Errorf("%s → %s: los estados de producto llevan sfp-", e, s)
		}
	}

	for _, e := range feature {
		s, hay := skills[e]
		if !hay {
			t.Errorf("el estado %q no tiene skill", e)
			continue
		}
		if !strings.HasPrefix(s, "sf-") {
			t.Errorf("%s → %s: los estados de feature llevan sf-", e, s)
		}
	}

	// La que muerde: sf no conoce a los utilitarios. Se chequea sobre el mapa
	// ENTERO y no sobre las dos listas de arriba, porque una entrada nueva mal
	// prefijada se colaría por el agujero de no estar en ninguna.
	for e, s := range skills {
		if strings.HasPrefix(s, "sfx-") {
			t.Errorf("%s → %s: un sfx- es standalone, no llama a `sf done`", e, s)
		}
	}

	// Y nadie comparte skill: dos estados con el mismo nombre fue justo lo que
	// pasó con sf-propose en el ⑨ y el ⑩.
	visto := map[string]string{}
	for e, s := range skills {
		if otro, repetido := visto[s]; repetido {
			t.Errorf("%s y %s comparten el skill %s", otro, e, s)
		}
		visto[s] = e
	}

	if len(skills) != len(producto)+len(feature) {
		t.Errorf("el mapa tiene %d entradas, hay %d estados", len(skills), len(producto)+len(feature))
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El ⑯: qué modelo pide esta feature
// ────────────────────────────────────────────────────────────────────────────

// enImplementar deja f-1 lista para implementar, con el tareas.json que se pida.
func (p *proyecto) enImplementar(tareasJSON string) *proyecto {
	p.t.Helper()
	p.productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Implementar}
	if tareasJSON != "" {
		p.conArchivoConTexto(filepath.Join(".docs/features/f-1-nucleo", tareas.Archivo), tareasJSON)
	}
	return p
}

// La cadena de precedencia tiene cuatro niveles y cada uno sabe MENOS que el de
// arriba. Este test la recorre entera, y es el que se rompe si alguien invierte
// dos.
//
// Los niveles 2 y 3 —el lote y la feature— son el ⑯ partido en dos: los dos son
// "lo que recomendó el que planificó", con lo más específico arriba.
func TestLaPrecedenciaDelModelo(t *testing.T) {
	const sinModelo = `{"feature":"f-1","tareas":[
		{"id":"t-1","lote":1,"descripcion":"x","satisface":["us-1/CA-1"],"tests":["a_test.go::TestX"]}]}`

	const porLote = `{"feature":"f-1","modelo":"medio","modelo_por_lote":{"1":"barato"},"tareas":[
		{"id":"t-1","lote":1,"descripcion":"x","satisface":["us-1/CA-1"],"tests":["a_test.go::TestX"]}]}`

	const conModelo = `{"feature":"f-1","modelo":"caro","tareas":[
		{"id":"t-1","lote":1,"descripcion":"x","satisface":["us-1/CA-1"],"tests":["a_test.go::TestX"]}]}`

	casos := []struct {
		nombre string
		tareas string
		deSf   string // lo que puso `sf model`
		quiero string
	}{
		{"sin nada: el default del perfil del estado", sinModelo, "", "id-de-medio"},
		{"el ⑯ de la feature le gana al default", conModelo, "", "id-de-caro"},
		{"el ⑯ del LOTE le gana al de la feature", porLote, "", "id-de-barato"},
		{"sf model le gana al ⑯", conModelo, "barato", "id-de-barato"},
		{"sf model manda aunque no haya ⑯", sinModelo, "barato", "id-de-barato"},
		// El plan roto no puede impedir que sf conteste qué sigue: de eso se
		// queja la compuerta del ⑰, que corre antes.
		{"un tareas.json ilegible cae al default", `{roto`, "", "id-de-medio"},
		{"sin tareas.json cae al default", "", "", "id-de-medio"},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p := nuevo(t).enImplementar(c.tareas).conTresPerfiles()
			p.e.Features["f-1"].Modelo = c.deSf
			// El nivel del LOTE sólo existe si hay un lote en curso: los lotes
			// se siembran en el primer `sf lote start`.
			p.e.Features["f-1"].Lotes = []estado.Lote{{Lote: 1, Rojo: true}}

			if m := p.next().Modelo; m != c.quiero {
				t.Errorf("modelo %q, quería %q", m, c.quiero)
			}
		})
	}
}

// ────────────────────────────────────────────────────────────────────────────
// H1b — el via sale del mapa, y un modelo sin declarar PARA
// ────────────────────────────────────────────────────────────────────────────

// enImplementarCon deja f-1 lista para implementar con el modelo que se pida.
func (p *proyecto) enImplementarCon(modelo string) *proyecto {
	p.t.Helper()
	p.productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Implementar, Modelo: modelo}
	return p
}

// EL TEST QUE JUSTIFICA EL MAPA. El ⑱ deja de ser un caso especial: es la misma
// llamada con otro `via:`, y el comando viene con ella — sin eso, "consola"
// sería una instrucción que el orquestador no puede ejecutar.
func TestUnModeloDeOtroProveedorSalePorConsola(t *testing.T) {
	p := nuevo(t).enImplementarCon("ajeno").conTresPerfiles()
	p.g.DeclararSuelto("ajeno", global.Modelo{ID: "ajeno-r1", Via: global.Consola, Comando: "ajeno exec"})

	i := p.next()
	if i.Tipo != Trabajar {
		t.Fatalf("tipo %v, quería Trabajar", i.Tipo)
	}
	if i.Via != global.Consola {
		t.Errorf("via %q, quería consola", i.Via)
	}
	if i.Comando != "ajeno exec" {
		t.Errorf("comando %q, quería el del catálogo", i.Comando)
	}
	// Y NO lleva agente: un `via: consola` lo ejecuta el orquestador con sus
	// manos, así que no hay portamodelo que invocar.
	if i.Agente != "" {
		t.Errorf("agente %q — consola no se lanza como subagente", i.Agente)
	}
	// Y el skill es EL MISMO: lo único que cambia es cómo se lanza.
	if i.Skill != skills[estado.Implementar] {
		t.Errorf("skill %q — el ⑱ no es un skill distinto", i.Skill)
	}
}

// Un modelo que no está en el mapa PARA, y sf no elige el reemplazo: eso sería
// opinar sobre qué modelo se parece a cuál, y R3 lo prohíbe.
// Un perfil que el harness no tiene declarado PARA, y sf no elige el reemplazo:
// eso sería opinar sobre qué modelo se parece a cuál, y R3 lo prohíbe.
func TestUnPerfilSinDeclararPara(t *testing.T) {
	p := nuevo(t).enImplementarCon("").conCatalogo("opencode", global.Catalogo{})

	i := p.next()
	if i.Tipo != Para {
		t.Fatalf("tipo %v, quería Para", i.Tipo)
	}
	// El mensaje nombra el perfil Y el harness: sin las dos cosas, quien lo lee
	// no sabe qué contestar ni dónde.
	if !strings.Contains(i.Mensaje, global.Construir) {
		t.Errorf("no dijo qué perfil falta: %q", i.Mensaje)
	}
	if !strings.Contains(i.Mensaje, "opencode") {
		t.Errorf("no dijo en qué harness falta: %q", i.Mensaje)
	}
	if !strings.Contains(strings.Join(i.Sugerido, " "), "--alias") {
		t.Errorf("no ofreció declarar con alias: %v", i.Sugerido)
	}
	// Medido: los agentes se leen al arrancar, así que declarar no alcanza.
	if !strings.Contains(strings.Join(i.Avisos, " "), "reiniciá") {
		t.Errorf("no avisó que hay que reiniciar el harness: %v", i.Avisos)
	}
}

// Un ALIAS que falta NO para: cae al default del perfil y avisa. La asimetría es
// deliberada — un perfil sin declarar no tiene salida, un alias faltante sí, y
// además cae para el lado seguro (el default es el primero de su lista).
func TestUnAliasQueFaltaAvisaYNoPara(t *testing.T) {
	p := nuevo(t).enImplementarCon("el-que-no-esta").conTresPerfiles()

	i := p.next()
	if i.Tipo != Trabajar {
		t.Fatalf("tipo %v: un alias que falta no puede frenar el bucle", i.Tipo)
	}
	if i.Modelo != "id-de-medio" {
		t.Errorf("modelo %q, quería el default de construir", i.Modelo)
	}
	if !strings.Contains(strings.Join(i.Avisos, " "), "el-que-no-esta") {
		t.Errorf("cayó al default sin decirlo: %v", i.Avisos)
	}
}

// Sin mapa —nadie corrió `sf install`— la máquina sigue funcionando. No tener
// el mapa no puede impedir trabajar: sólo impide resolver `consola`.
func TestSinCatalogoCaeASubagente(t *testing.T) {
	i := nuevo(t).enImplementarCon("loquesea").next()

	if i.Tipo != Trabajar {
		t.Fatalf("tipo %v: sin sf install la máquina tiene que andar igual", i.Tipo)
	}
	if i.Via != global.Subagente {
		t.Errorf("via %q, quería subagente", i.Via)
	}
}

// El brief conversa, y eso NO depende del mapa: un subagente arranca, trabaja y
// muere — no te habla.
func TestElBriefSiempreEsVos(t *testing.T) {
	p := nuevo(t).conTresPerfiles()

	if i := p.next(); i.Via != global.Vos {
		t.Errorf("el brief salió con via %q", i.Via)
	}
}

// El ⑧ también conversa, y por el mismo motivo: la constitución NO SE DERIVA
// del PRD — arquitectura, stack y convenciones son decisiones de Javier.
//
// Estuvo del lado equivocado y nadie lo vio: su skill dice "Step 1: Interview
// Javier" y los modelos de Anthropic se lo salteaban, así que la contradicción
// no aparecía. Nemotron hizo lo que estaba escrito y preguntó al aire.
func TestLaConstitucionTambienEsVos(t *testing.T) {
	p := nuevo(t).conTresPerfiles()
	p.conArchivo(".docs/brief.md").conArchivo(".docs/prd.md")
	p.e.Producto.BriefSellado = "hacelo"
	p.e.Producto.PrdHash = "a3f9c1"

	i := p.next()
	if i.Estado != "constitucion" {
		t.Fatalf("estado %q, quería constitucion", i.Estado)
	}
	if i.Via != global.Vos {
		t.Errorf("el ⑧ salió con via %q: no puede ser subagente, tiene que poder preguntar", i.Via)
	}
	// Y sin agente ni modelo: no hay a quién delegarle esto.
	if i.Agente != "" {
		t.Errorf("el ⑧ trajo agente %q", i.Agente)
	}
}

// Los que NO conversan siguen delegando. Es la contracara del test de arriba:
// sin esto, mover un estado a `conversan` por error no rompería nada.
func TestElPrdNoConversa(t *testing.T) {
	p := nuevo(t).conTresPerfiles()
	p.conArchivo(".docs/brief.md")
	p.e.Producto.BriefSellado = "hacelo"

	i := p.next()
	if i.Estado != "prd" {
		t.Fatalf("estado %q, quería prd", i.Estado)
	}
	if i.Via != global.Subagente {
		t.Errorf("el ⑦ salió con via %q, quería subagente", i.Via)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El plan que envejece
// ────────────────────────────────────────────────────────────────────────────

// Planificás f-2, implementás f-1, y la spec de f-2 queda mirando un repo que ya
// cambió. Avisa y NO frena: qué cambió y si importa es criterio, y el criterio
// es de Javier.
func TestAvisaCuandoElPlanSeEscribioSobreOtroCommit(t *testing.T) {
	p := nuevo(t).enImplementarCon("")
	p.e.Features["f-1"].BaseCommit = "0000000000000000000000000000000000000000"

	i := p.next()
	if i.Tipo != Trabajar {
		t.Fatalf("tipo %v: el plan viejo AVISA, no frena", i.Tipo)
	}
	// El proyecto de prueba no es un repo git, así que no hay con qué comparar:
	// sin git no se puede saber si el suelo se movió, y callarse es correcto.
	if len(i.Avisos) != 0 {
		t.Errorf("avisó sin poder comparar: %v", i.Avisos)
	}
}

// Sin base_commit no hay nada que comparar, y eso es lo normal en el camino
// corto de un bug: ahí no hay planificación que lo escriba.
func TestSinBaseCommitNoAvisa(t *testing.T) {
	i := nuevo(t).enImplementarCon("").next()

	if len(i.Avisos) != 0 {
		t.Errorf("avisó sin base_commit: %v", i.Avisos)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El ⑧ tiene que ser alcanzable
// ────────────────────────────────────────────────────────────────────────────

// `sf init` SIEMPRE crea la constitución —tiene que crearla, ahí escribe el
// test_cmd que detectó del stack—, así que el checkpoint por existencia daba
// siempre "está escrita" y del ⑦ se saltaba derecho a la 🛑. Lo que se sellaba
// era la plantilla, y `sfp-constitucion` no lo invocaba nadie nunca.
func TestElOchoEsAlcanzableConLaPlantillaDelAndamio(t *testing.T) {
	p := nuevo(t)
	p.e.Producto = estado.Producto{BriefSellado: "hacelo", PrdHash: "a3f9c1"}
	p.conArchivoConTexto(docs.Constitucion,
		"---\nlenguaje: go\ntest_cmd: go test ./...\n---\n\n# Constitución\n\n"+
			constitucion.MarcaSinEscribir+"\n\n## Arquitectura\n")

	i := p.next()
	if i.Tipo != Trabajar {
		t.Fatalf("tipo %v, quería Trabajar: la constitución es la plantilla", i.Tipo)
	}
	if i.Estado != "constitucion" || i.Skill != "sfp-constitucion" {
		t.Errorf("estado %q skill %q, quería constitucion/sfp-constitucion", i.Estado, i.Skill)
	}
}

// Y cuando el ⑧ la escribió de verdad —el marcador ya no está—, vuelve la 🛑:
// el arreglo no puede dejar la máquina pidiendo la constitución para siempre.
func TestElOchoParaCuandoLaConstitucionYaSeEscribio(t *testing.T) {
	p := nuevo(t)
	p.e.Producto = estado.Producto{BriefSellado: "hacelo", PrdHash: "a3f9c1"}
	p.conArchivoConTexto(docs.Constitucion,
		"---\nlenguaje: go\ntest_cmd: go test ./...\n---\n\n# Constitución\n\n"+
			"## Arquitectura\nHexagonal, con los puertos en internal/.\n")

	i := p.next()
	if i.Tipo != Para {
		t.Fatalf("tipo %v, quería Para: la constitución está escrita", i.Tipo)
	}
	if !slices.Contains(i.Sugerido, "sf approve") {
		t.Errorf("sugirió %v, quería sf approve", i.Sugerido)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// La cadena de modelo vale en los CUATRO estados de feature
// ────────────────────────────────────────────────────────────────────────────

// `modeloDeFeature` implementa los tres niveles y tenía un solo consumidor:
// `implementando`. Los otros tres pasaban por `trabajar`, que sólo mira el mapa
// por estado — así que `sf model opus` no hacía nada en `planificacion`,
// `revision` ni `cierre`.
//
// Es la salida de ME TRABÉ, o sea la que Javier usa mirando el bucle patinar, y
// no funcionaba en tres de los cuatro estados.
func TestElModeloDeJavierGanaEnLosCuatroEstadosDeFeature(t *testing.T) {
	for _, est := range []string{
		estado.Planificacion, estado.Implementar, estado.Revision, estado.Cierre,
	} {
		t.Run(est, func(t *testing.T) {
			p := nuevo(t).productoListo().enImplementar(`{"tareas":[{"id":"t-1","lote":1}]}`)
			p.e.Features["f-1"].Estado = est
			p.e.Features["f-1"].Modelo = "ajeno"
			p.conTresPerfiles()
			p.g.DeclararSuelto("ajeno", global.Modelo{ID: "ajeno-x", Via: global.Consola, Comando: "ajeno exec"})

			if i := p.next(); i.Modelo != "ajeno-x" {
				t.Errorf("modelo %q, quería el de `sf model`", i.Modelo)
			}
		})
	}
}

// El nivel 2 es el ⑯: "ESTA feature necesita uno más grande", escrito en
// tareas.json por el mismo que planificó. Le gana al default del estado y
// pierde contra `sf model`.
func TestElModeloDelPlanGanaAlDefaultDelEstado(t *testing.T) {
	p := nuevo(t).productoListo().
		enImplementar(`{"modelo":"caro","tareas":[{"id":"t-1","lote":1}]}`).conTresPerfiles()
	p.e.Features["f-1"].Estado = estado.Cierre

	// El default de `cierre` es el de construir, y el plan pide `caro`.
	if i := p.next(); i.Modelo != "id-de-caro" {
		t.Errorf("modelo %q, quería el de caro: lo pidió el ⑯", i.Modelo)
	}

	// Y `sf model` le gana igual, porque Javier decide en runtime y el ⑯ se
	// escribió antes de ver fallar nada.
	p.e.Features["f-1"].Modelo = "barato"
	if i := p.next(); i.Modelo != "id-de-barato" {
		t.Errorf("modelo %q, quería el de barato: le gana la decisión de runtime", i.Modelo)
	}
}

// Y sin nada pedido manda el perfil del estado: `revision` pide `razonar` porque
// juzgar es donde el modelo que piensa paga.
func TestSinDecidirNadaMandaElPerfilDelEstado(t *testing.T) {
	p := nuevo(t).productoListo().
		enImplementar(`{"tareas":[{"id":"t-1","lote":1}]}`).conTresPerfiles()
	p.e.Features["f-1"].Estado = estado.Revision

	i := p.next()
	if i.Perfil != global.Razonar {
		t.Errorf("perfil %q, quería razonar", i.Perfil)
	}
	if i.Modelo != "id-de-caro" {
		t.Errorf("modelo %q, quería el default de razonar", i.Modelo)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// H8 — el portamodelo
// ────────────────────────────────────────────────────────────────────────────

// En Claude Code el modelo va como parámetro de la llamada, así que NO hace
// falta portamodelo y `agente:` viene vacío. En los otros dos el `model:` sale
// del archivo del agente y no se puede pisar al invocar (medido), así que sf
// tiene que decir a quién invocar.
func TestElAgenteSoloApareceDondeHaceFalta(t *testing.T) {
	casos := map[string]string{
		"claude-code": "",
		"opencode":    "sf-medio",
		"commandcode": "sf-medio",
	}
	for harness, quiero := range casos {
		t.Run(harness, func(t *testing.T) {
			p := nuevo(t).productoListo().
				enImplementar(`{"tareas":[{"id":"t-1","lote":1}]}`).conTresPerfiles()
			p.g.Harnesses[harness] = p.g.Harnesses["claude-code"]
			parandoEn(t, harness)

			if i := p.next(); i.Agente != quiero {
				t.Errorf("agente %q, quería %q", i.Agente, quiero)
			}
		})
	}
}

// ME TRABÉ tiene que decir CON QUÉ MODELO se trabó, y decía "con ." — el campo
// `f.Modelo` está vacío hasta el primer `sf model`, y justo la primera vez que
// aparece esta parada es la vez que nadie lo corrió todavía.
//
// Es la mitad del dato que hace falta para contestarle: "¿subo el modelo?" no
// se puede decidir sin saber cuál está fallando.
func TestMeTrabeDiceConQueModeloSeTrabo(t *testing.T) {
	p := nuevo(t).productoListo().
		enImplementar(`{"tareas":[{"id":"t-1","lote":1}]}`).conTresPerfiles()
	p.e.Features["f-1"].IntentosFallidos = TopeIntentos

	i := p.next()
	if i.Tipo != MeTrabe {
		t.Fatalf("tipo %v, quería MeTrabe", i.Tipo)
	}
	// El ALIAS y no sólo el perfil: "falló con construir" no alcanza para
	// decidir si subirlo, porque en construir hay varios.
	if !strings.Contains(i.Mensaje, "medio") {
		t.Errorf("no dijo con qué modelo se trabó: %q", i.Mensaje)
	}
	if strings.Contains(i.Mensaje, "con .") {
		t.Errorf("el nombre del modelo salió vacío: %q", i.Mensaje)
	}
}

// parandoEn dice en qué harness está corriendo el test.
//
// Hace falta desde que el puntero de harness se DETECTA: esta suite corre
// adentro de algún arnés, así que sin limpiar las variables cada test heredaría
// el de quien lo lanzó. Que haga falta es la prueba de que el mecanismo anda.
func parandoEn(t *testing.T, harness string) {
	t.Helper()
	for _, v := range global.VarsDeHarness {
		t.Setenv(v, "")
	}
	if harness != "" {
		t.Setenv(global.VarHarness, harness)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El horizonte de las paradas
// ────────────────────────────────────────────────────────────────────────────

// Lo que una parada ANUNCIA tiene que ser lo que después PASA.
//
// Es el test que hace que el horizonte pueda vivir en dos mapas chicos en vez de
// en un texto escrito a mano en cada parada: acá se camina la máquina de verdad
// y se compara el anuncio contra el recorrido. Si alguien mueve un estado de
// lugar y no toca los mapas, esto se rompe.
func TestElHorizonteEsCierto(t *testing.T) {
	p := nuevo(t).conTresPerfiles()

	// ── el ⑥
	p.conArchivo(".docs/brief.md")
	i := p.next()
	if i.Tipo != Para {
		t.Fatalf("tipo %v, quería la 🛑 del ⑥", i.Tipo)
	}
	anunciados, parada := i.SiApruebas, i.ProximaParada
	if len(anunciados) == 0 || parada == "" {
		t.Fatalf("la 🛑 del ⑥ no anunció nada: corren=%v parada=%q", anunciados, parada)
	}

	// Apruebo y camino: cada paso anunciado tiene que ser el que sale.
	p.e.Producto.BriefSellado = "hacelo"
	if i := p.next(); comoSeLlama[i.Estado] != anunciados[0] {
		t.Errorf("anunció %q y salió %q", anunciados[0], comoSeLlama[i.Estado])
	}

	p.conArchivo(".docs/prd.md")
	p.e.Producto.PrdHash = "a3f9c1"
	i = p.next()
	if len(anunciados) < 2 {
		t.Fatalf("anunció un solo paso: entre el ⑦ y el ⑧ no hay parada, tienen que ser dos")
	}
	if comoSeLlama[i.Estado] != anunciados[1] {
		t.Errorf("anunció %q y salió %q", anunciados[1], comoSeLlama[i.Estado])
	}

	// Y ahí tiene que aparecer la parada anunciada.
	p.conArchivo(".docs/constitucion.md")
	i = p.next()
	if i.Tipo != Para {
		t.Fatalf("anunció parada en %q y no paró: tipo %v", parada, i.Tipo)
	}
	if i.Estado != "constitucion" {
		t.Errorf("paró en %q, y había anunciado %q", i.Estado, parada)
	}

	// ── el ⑧: mismo contrato, un tramo más corto.
	anunciados, parada = i.SiApruebas, i.ProximaParada
	if len(anunciados) != 1 || parada == "" {
		t.Fatalf("el ⑧ anunció corren=%v parada=%q, quería un solo paso", anunciados, parada)
	}
	p.e.Producto.ConstitucionSellada = true
	if i := p.next(); comoSeLlama[i.Estado] != anunciados[0] {
		t.Errorf("anunció %q y salió %q", anunciados[0], comoSeLlama[i.Estado])
	}
	p.conHistoria("us-1")
	if i := p.next(); i.Tipo != Barata {
		t.Fatalf("anunció parada en %q y no paró", parada)
	}
}

// La ⏸ del ⑨ es la que más sorprende: aprobás y arrancan DOS pasos.
func TestLaPausaDelNueveAvisaQueSonDos(t *testing.T) {
	p := nuevo(t).conTresPerfiles()
	p.e.Producto = estado.Producto{BriefSellado: "hacelo", PrdHash: "x", ConstitucionSellada: true}
	p.conHistoria("us-1")

	i := p.next()
	if i.Tipo != Barata {
		t.Fatalf("tipo %v, quería la ⏸", i.Tipo)
	}
	if len(i.SiApruebas) != 2 {
		t.Errorf("anunció %v: entre el ⑩ y el ⑫ no hay parada, tienen que ser dos", i.SiApruebas)
	}
	if i.ProximaParada == "" {
		t.Error("no dijo dónde se vuelve a parar")
	}
}

// Todo estado del orden tiene nombre legible. Es lo único no derivable de los
// mapas, así que es lo único que hay que acordarse de completar.
func TestTodosLosEstadosTienenNombre(t *testing.T) {
	for _, e := range ordenDeEstados {
		if comoSeLlama[e] == "" {
			t.Errorf("el estado %q no tiene nombre legible", e)
		}
	}
}

// Un estado que no está en el orden no inventa horizonte.
func TestUnEstadoDesconocidoNoTieneHorizonte(t *testing.T) {
	corren, parada := horizonte("no-existe")
	if corren != nil || parada != "" {
		t.Errorf("horizonte inventado: %v %q", corren, parada)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// ME TRABÉ EN LA REVISIÓN — el fondo del otro bucle
// ────────────────────────────────────────────────────────────────────────────

// `cerrarRevision` es la única transición que va para atrás, y hasta acá no
// tenía tope: revisor encuentra → implementador no arregla → revisor encuentra,
// para siempre. Al llegar a TopeRondas `sf next` levanta la parada en vez de
// seguir proponiendo trabajo.
func TestRondasDeRevisionLevantanLaParada(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{
		Estado:         estado.Revision,
		RondasRevision: TopeRondas,
	}

	i := p.next()
	if i.Tipo != MeTrabe {
		t.Fatalf("tipo %v, quería MeTrabe", i.Tipo)
	}
	if !strings.Contains(i.Mensaje, "REVISIÓN") {
		t.Errorf("el mensaje no distingue este bucle del otro: %q", i.Mensaje)
	}
	// La salida de ESTE bucle es adjudicar, no subir el modelo: el problema es
	// un desacuerdo sobre un hallazgo, y otro intento no lo rompe.
	if len(i.Sugerido) == 0 || i.Sugerido[0] != `sf dismiss <h-#> "motivo"` {
		t.Errorf("sugirió %v, quería sf dismiss primero", i.Sugerido)
	}
}

func TestPorDebajoDelTopeDeRondasSigueTrabajando(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{
		Estado:         estado.Revision,
		RondasRevision: TopeRondas - 1,
	}

	if i := p.next(); i.Tipo == MeTrabe {
		t.Errorf("con %d rondas ya paró: tiene que quedar una vuelta", TopeRondas-1)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El camino chico y su trinquete
// ────────────────────────────────────────────────────────────────────────────

// Una feature cuyas historias son todas `chico` no pasa por planificación: el
// ⑫–⑯ no corre y `sf next` manda a implementar directo.
func TestElCaminoChicoSalteaLaPlanificacion(t *testing.T) {
	p := nuevo(t).productoListo()
	p.conArchivoConTexto(docs.Historia("us-1"),
		"---\nid: us-1\ntipo: chico\ntitulo: un flag mas\n---\n"+
			"## Criterios de aceptación\n- **CA-1** — acepta --json\n")
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Planificacion}

	if i := p.next(); i.Estado != estado.Implementar {
		t.Errorf("estado %q, quería implementar: `tipo: chico` saltea el ⑫–⑯", i.Estado)
	}
}

// Y el trinquete: una vez ampliada, la misma feature con las mismas historias
// vuelve a pasar por planificación. Nadie editó el us-#, y tiene que alcanzar.
func TestUnaFeatureAmpliadaVuelveAPlanificar(t *testing.T) {
	p := nuevo(t).productoListo()
	p.conArchivoConTexto(docs.Historia("us-1"),
		"---\nid: us-1\ntipo: chico\ntitulo: un flag mas\n---\n"+
			"## Criterios de aceptación\n- **CA-1** — acepta --json\n")
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Planificacion, Ampliada: true}

	if i := p.next(); i.Estado != estado.Planificacion {
		t.Errorf("estado %q, quería planificacion: el trinquete no bajó", i.Estado)
	}
}
