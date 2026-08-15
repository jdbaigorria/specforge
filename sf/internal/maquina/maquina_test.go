package maquina

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

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

// conModelos declara un mapa de modelos para este proyecto de prueba.
//
// Los tests que NO lo llaman corren con g == nil, que es el caso de "todavía no
// se corrió sf install" — y tiene que seguir funcionando: no tener el mapa no
// puede impedir trabajar.
func (p *proyecto) conModelos(m map[string]global.Modelo) *proyecto {
	p.g = &global.Config{Harness: "claude-code", Modelos: m}
	return p
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

	// ⑨ con historias: ⏸, la parada barata.
	p.conArchivo(".docs/backlog/us-1.md")
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
	p.conArchivo(".docs/backlog/us-1.md")

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
	p.conArchivoConTexto(filepath.Join(carpeta, "decision.md"),
		"## A — una\n## B — otra\n## C — otra más\n")
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
	p := implementando1(t, []estado.Lote{{Lote: 1, Rojo: true}})
	p.e.Features["f-1"].Modelo = "opus"

	if i := p.next(); i.Modelo != "opus" {
		t.Errorf("modelo %q, quería opus (el que Javier subió)", i.Modelo)
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
	if !slices.Contains(i.Sugerido, "sf model <nombre>") {
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

	if i := p.next(); i.Modelo != "opus" {
		t.Errorf("modelo %q, quería opus", i.Modelo)
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
			if modelo(est) == "" {
				t.Errorf("el estado %q no resuelve modelo", est)
			}
			// Sin mapa (nadie corrió sf install) el via igual resuelve: es
			// lo que permite trabajar antes de instalar nada.
			if v, _, hay := via(est, modelo(est), nil); !hay || v == "" {
				t.Errorf("el estado %q no resuelve via sin mapa", est)
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

// La cadena de precedencia es de tres niveles y cada uno sabe MENOS que el de
// arriba. Este test la recorre entera, y es el que se rompe si alguien invierte
// dos.
func TestLaPrecedenciaDelModelo(t *testing.T) {
	const conModelo = `{"feature":"f-1","modelo":"opus","tareas":[
		{"id":"t-1","lote":1,"descripcion":"x","satisface":["us-1/CA-1"],"tests":["a_test.go::TestX"]}]}`
	const sinModelo = `{"feature":"f-1","tareas":[
		{"id":"t-1","lote":1,"descripcion":"x","satisface":["us-1/CA-1"],"tests":["a_test.go::TestX"]}]}`

	casos := []struct {
		nombre string
		tareas string
		deSf   string // lo que puso `sf model`
		quiero string
	}{
		{"sin nada: el default del estado", sinModelo, "", modeloPorDefecto},
		{"el ⑯ le gana al default", conModelo, "", "opus"},
		{"sf model le gana al ⑯", conModelo, "haiku", "haiku"},
		{"sf model manda aunque no haya ⑯", sinModelo, "haiku", "haiku"},
		// El plan roto no puede impedir que sf conteste qué sigue: de eso se
		// queja la compuerta del ⑰, que corre antes.
		{"un tareas.json ilegible cae al default", `{roto`, "", modeloPorDefecto},
		{"sin tareas.json cae al default", "", "", modeloPorDefecto},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			p := nuevo(t).enImplementar(c.tareas)
			p.e.Features["f-1"].Modelo = c.deSf

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
	p := nuevo(t).enImplementarCon("deepseek").conModelos(map[string]global.Modelo{
		"sonnet":   {Via: global.Subagente},
		"deepseek": {Via: global.Consola, Comando: "deepseek exec"},
	})

	i := p.next()
	if i.Tipo != Trabajar {
		t.Fatalf("tipo %v, quería Trabajar", i.Tipo)
	}
	if i.Via != global.Consola {
		t.Errorf("via %q, quería consola", i.Via)
	}
	if i.Comando != "deepseek exec" {
		t.Errorf("comando %q, quería el del mapa", i.Comando)
	}
	// Y el skill es EL MISMO: lo único que cambia es cómo se lanza.
	if i.Skill != skills[estado.Implementar] {
		t.Errorf("skill %q — el ⑱ no es un skill distinto", i.Skill)
	}
}

// Un modelo que no está en el mapa PARA, y sf no elige el reemplazo: eso sería
// opinar sobre qué modelo se parece a cuál, y R3 lo prohíbe.
func TestUnModeloSinDeclararPara(t *testing.T) {
	p := nuevo(t).enImplementarCon("grok").conModelos(map[string]global.Modelo{
		"sonnet": {Via: global.Subagente},
	})

	i := p.next()
	if i.Tipo != Para {
		t.Fatalf("tipo %v, quería Para", i.Tipo)
	}
	if !strings.Contains(i.Mensaje, "grok") {
		t.Errorf("no dijo cuál falta: %q", i.Mensaje)
	}
	// Las tres salidas son las del diseño, y las tres declaran.
	if len(i.Sugerido) != 3 {
		t.Errorf("sugirió %v, querían ser tres salidas", i.Sugerido)
	}
	if !strings.Contains(strings.Join(i.Sugerido, " "), "--via consola") {
		t.Errorf("no ofreció la salida por consola: %v", i.Sugerido)
	}
}

// Sin mapa —nadie corrió `sf install`— la máquina sigue funcionando. No tener
// el mapa no puede impedir trabajar: sólo impide resolver `consola`.
func TestSinMapaCaeASubagente(t *testing.T) {
	i := nuevo(t).enImplementarCon("loquesea").next()

	if i.Tipo != Trabajar {
		t.Fatalf("tipo %v: sin sf install la máquina tiene que andar igual", i.Tipo)
	}
	if i.Via != global.Subagente {
		t.Errorf("via %q, quería subagente", i.Via)
	}
}

// El brief es el único que conversa, y eso NO depende del mapa: un subagente
// arranca, trabaja y muere — no te habla.
func TestElBriefSiempreEsVos(t *testing.T) {
	p := nuevo(t).conModelos(map[string]global.Modelo{
		"sonnet": {Via: global.Consola, Comando: "lo que sea"},
	})

	if i := p.next(); i.Via != global.Vos {
		t.Errorf("el brief salió con via %q", i.Via)
	}
}
