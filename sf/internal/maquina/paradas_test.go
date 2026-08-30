package maquina

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/revision"
)

// ────────────────────────────────────────────────────────────────────────────
// sf approve — un comando, cinco significados
// ────────────────────────────────────────────────────────────────────────────

// El veredicto NO lo elige sf: lo escribió Javier en el brief. `sf approve`
// significa "sí, sellalo con lo que dice" — incluso si dice que no.
func TestApproveCopiaElVeredictoDelBrief(t *testing.T) {
	for _, v := range []string{"hacelo", "pivotea", "no-lo-hagas"} {
		t.Run(v, func(t *testing.T) {
			p := nuevo(t)
			p.conArchivoConTexto(docs.Brief, "---\nveredicto: "+v+"\n---\n# brief\n")

			ef := Aprobar(p.raiz, p.e, p.r)
			if !ef.Pasa() {
				t.Fatalf("no pasó: %v", ef.Fallas)
			}
			if p.e.Producto.BriefSellado != v {
				t.Errorf("selló %q, quería %q", p.e.Producto.BriefSellado, v)
			}
		})
	}
}

func TestApproveSinVeredictoNoSella(t *testing.T) {
	p := nuevo(t)
	p.conArchivoConTexto(docs.Brief, "---\ntipo: brief\n---\n# sin veredicto\n")

	if ef := Aprobar(p.raiz, p.e, p.r); ef.Pasa() {
		t.Error("selló un brief sin veredicto")
	}
}

// El ⑦ no tiene parada: se pasa derecho al ⑧. Aprobar ahí es un error de uso, y
// el mensaje tiene que decir cuál es el comando correcto.
func TestApproveEnElPrdMandaAlDone(t *testing.T) {
	p := nuevo(t)
	p.e.Producto.BriefSellado = "hacelo"

	ef := Aprobar(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("aprobó un estado sin parada")
	}
	if !strings.Contains(strings.Join(ef.Fallas, ""), "sf done") {
		t.Errorf("no dijo cuál es el comando correcto: %v", ef.Fallas)
	}
}

// El ⑰: aprobar el plan manda a implementar. La puerta "otra feature" NO se
// elige acá — se elige con `sf take` después. Cada comando, un trabajo.
func TestApproveDelPlanMandaAImplementar(t *testing.T) {
	p := nuevo(t).productoListo().conPlanCompleto()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Planificacion, Rechazo: "el diseño B no cierra"}

	ef := Aprobar(p.raiz, p.e, p.r)
	if !ef.Pasa() {
		t.Fatalf("no pasó: %v", ef.Fallas)
	}
	if p.e.Features["f-1"].Estado != estado.Implementar {
		t.Errorf("quedó en %q, quería implementar", p.e.Features["f-1"].Estado)
	}
	// Aprobar limpia el rechazo anterior: ya no hay nada que corregir.
	if p.e.Features["f-1"].Rechazo != "" {
		t.Error("el rechazo sobrevivió a la aprobación")
	}
}

// La ⏸ del ㉓ termina en archivar, y archivar es esto: `sf feature archive` no
// existe como comando aparte (H16).
func TestApproveEnCierreArchivaLaCarpeta(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Cierre}

	carpeta := filepath.Join(".docs", "features", "f-1-nucleo")
	p.conArchivoConTexto(filepath.Join(carpeta, "doc.md"), "# doc\n")
	p.conArchivoConTexto(filepath.Join(carpeta, "journal.md"), "- lección\n")

	ef := Aprobar(p.raiz, p.e, p.r)
	if !ef.Pasa() {
		t.Fatalf("no pasó: %v", ef.Fallas)
	}

	// La carpeta se mueve ENTERA, con todo adentro.
	destino := filepath.Join(p.raiz, docs.Archivado, "f-1-nucleo", "journal.md")
	if _, err := os.Stat(destino); err != nil {
		t.Errorf("el journal no viajó con la carpeta: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p.raiz, carpeta)); err == nil {
		t.Error("la carpeta original sigue ahí")
	}

	if p.e.Features["f-1"].Estado != estado.Cerrada {
		t.Errorf("la feature quedó en %q, quería cerrada", p.e.Features["f-1"].Estado)
	}
	if p.e.FeatureActual != "" {
		t.Errorf("feature_actual quedó en %q después de cerrar", p.e.FeatureActual)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// sf reject — el motivo es lo único que viaja
// ────────────────────────────────────────────────────────────────────────────

// Sin motivo no hay reject. El que va a rehacer el trabajo arranca de cero: sin
// saber qué estuvo mal, vuelve a proponer lo mismo.
func TestRejectExigeMotivo(t *testing.T) {
	p := nuevo(t)

	ef := Rechazar(p.e, "")
	if ef.Pasa() {
		t.Fatal("rechazó sin motivo")
	}
	if !strings.Contains(strings.Join(ef.Fallas, ""), "lo mismo") {
		t.Errorf("el mensaje no explica por qué hace falta: %v", ef.Fallas)
	}
}

func TestRejectGuardaElMotivoSegunDondeEstes(t *testing.T) {
	// En producto va al producto.
	p := nuevo(t)
	if ef := Rechazar(p.e, "falta la comparación con lo que existe"); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}
	if p.e.Producto.Rechazo == "" {
		t.Error("el motivo no quedó guardado en el producto")
	}

	// En una feature va a la feature.
	p = nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Planificacion}

	if ef := Rechazar(p.e, "el diseño B no contempla el modo lote"); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}
	if p.e.Features["f-1"].Rechazo == "" {
		t.Error("el motivo no quedó guardado en la feature")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// sf take — el ⑪
// ────────────────────────────────────────────────────────────────────────────

func TestTakePrimeraVezEntraPorPlanificacion(t *testing.T) {
	p := nuevo(t).productoListo()

	if ef := Tomar(p.e, p.r, "f-1"); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}
	if p.e.FeatureActual != "f-1" {
		t.Errorf("feature_actual = %q", p.e.FeatureActual)
	}
	if p.e.Features["f-1"].Estado != estado.Planificacion {
		t.Errorf("entró en %q, quería planificacion", p.e.Features["f-1"].Estado)
	}
}

// La vuelta de la puerta "otra feature" del ⑰: el plan ya está aprobado y
// esperando, así que retomarla es implementar, no replanificar.
func TestTakeSobreUnaPlanificadaVaAImplementar(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Planificada}

	if ef := Tomar(p.e, p.r, "f-1"); !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}
	if p.e.Features["f-1"].Estado != estado.Implementar {
		t.Errorf("quedó en %q, quería implementar", p.e.Features["f-1"].Estado)
	}
}

func TestTakeRechazaLoQueNoExisteOYaCerro(t *testing.T) {
	p := nuevo(t).productoListo()

	if ef := Tomar(p.e, p.r, "f-99"); ef.Pasa() {
		t.Error("tomó una feature que no está en el roadmap")
	}

	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Cerrada}
	if ef := Tomar(p.e, p.r, "f-1"); ef.Pasa() {
		t.Error("tomó una feature ya cerrada")
	}
}

// Dejar algo a medias AVISA, no frena: el trabajo no se pierde —queda guardado
// en `features`— y frenar sería sf decidiendo por Javier.
func TestTakeAvisaSiDejasAlgoAMedias(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Implementar}

	ef := Tomar(p.e, p.r, "f-2")
	if !ef.Pasa() {
		t.Fatalf("frenó en vez de avisar: %v", ef.Fallas)
	}
	if !strings.Contains(ef.Mensaje, "f-1") || !strings.Contains(ef.Mensaje, "⚠") {
		t.Errorf("no avisó de lo que quedó a medias: %q", ef.Mensaje)
	}
	// Y el estado de f-1 sigue intacto para poder volver.
	if p.e.Features["f-1"].Estado != estado.Implementar {
		t.Error("cambiar de feature pisó el estado de la anterior")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Las dos salidas de ME TRABÉ
// ────────────────────────────────────────────────────────────────────────────

// Cambiar de modelo resetea el contador: es empezar de nuevo, no seguir
// acumulando los fracasos del modelo anterior.
func TestModelResetaElContador(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{
		Estado: estado.Implementar, Modelo: "el-viejo", IntentosFallidos: 3,
	}
	g := global.Semilla("claude-code")
	if _, err := g.Declarar(global.Construir, global.Modelo{Alias: "el-nuevo", ID: "x", Via: global.Subagente}); err != nil {
		t.Fatal(err)
	}

	ef := Modelo(p.e, g, "el-nuevo", "", "", "", "")
	if !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}
	if p.e.Features["f-1"].Modelo != "el-nuevo" {
		t.Errorf("modelo = %q", p.e.Features["f-1"].Modelo)
	}
	if p.e.Features["f-1"].IntentosFallidos != 0 {
		t.Errorf("intentos_fallidos = %d, quería 0", p.e.Features["f-1"].IntentosFallidos)
	}
	if !strings.Contains(ef.Mensaje, "el-viejo") {
		t.Errorf("el mensaje no dice de qué modelo venía: %q", ef.Mensaje)
	}
}

// Declarar un perfil NO exige que haya una feature en curso: la 🛑 puede saltar
// en cualquiera de los cinco estados de producto, donde no hay ninguna.
// Exigirla ahí dejaría la parada sin salida.
func TestDeclararUnPerfilNoNecesitaFeature(t *testing.T) {
	e := &estado.Estado{Features: map[string]*estado.Feature{}}
	g := global.Semilla("opencode")

	ef := Modelo(e, g, global.Razonar, "grande", "prov/grande", global.Subagente, "")
	if !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}
	if m, hay := g.Default(global.Razonar); !hay || m.ID != "prov/grande" {
		t.Fatalf("no lo declaró: %+v", m)
	}
	if !ef.Global {
		t.Error("no marcó que hay que guardar el catálogo")
	}
	if !ef.Portamodelo {
		t.Error("un alias nuevo tiene que pedir regenerar los portamodelo")
	}
	// Medido: los agentes se leen al arrancar.
	if !strings.Contains(strings.Join(ef.Avisos, " "), "reiniciá") {
		t.Errorf("no avisó del reinicio: %v", ef.Avisos)
	}
}

// Y redeclarar el mismo alias NO pide reiniciar: no aparece ningún archivo
// nuevo, y avisar cuando no hace falta enseña a ignorar el aviso.
func TestRedeclararNoPideReiniciar(t *testing.T) {
	e := &estado.Estado{Features: map[string]*estado.Feature{}}
	g := global.Semilla("opencode")

	Modelo(e, g, global.Razonar, "grande", "v1", global.Subagente, "")
	ef := Modelo(e, g, global.Razonar, "grande", "v2", global.Subagente, "")
	if ef.Portamodelo {
		t.Error("redeclarar no agrega ningún portamodelo")
	}
	if strings.Contains(strings.Join(ef.Avisos, " "), "reiniciá") {
		t.Errorf("avisó de un reinicio que no hace falta: %v", ef.Avisos)
	}
}

// Declarar sin --alias se rechaza: el alias es lo que va a escribir el ⑯ en
// tareas.json, y sin él no hay forma de nombrar el modelo desde el plan.
func TestDeclararSinAliasSeRechaza(t *testing.T) {
	e := &estado.Estado{Features: map[string]*estado.Feature{}}
	ef := Modelo(e, global.Semilla("opencode"), global.Razonar, "", "prov/x", global.Subagente, "")
	if ef.Pasa() {
		t.Fatal("dejó declarar un modelo sin alias")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "--alias") {
		t.Errorf("la falla no dice qué falta: %v", ef.Fallas)
	}
}

// Descartar tapa un bucle infinito real: un falso positivo no se arregla nunca,
// porque no está roto.
func TestDismissMarcaElHallazgoYDesbloquea(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Revision, IntentosFallidos: 3}

	ruta := filepath.Join(".docs", "features", "f-1-nucleo", docs.Revision)
	p.conArchivoConTexto(ruta, `{"vuelta":3,"hallazgos":[
		{"id":"h-2","estado":"abierto","detalle":"el parser acepta frontmatter sin cerrar"}]}`)

	ef := Descartar(p.raiz, p.e, p.r, "h-2", "es intencional: el ⑨ lo tolera a propósito")
	if !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}

	// El cambio tiene que quedar EN EL ARCHIVO: si no, la vuelta siguiente lo
	// vuelve a encontrar porque el código sigue igual.
	rev, err := revision.Leer(filepath.Join(p.raiz, ruta))
	if err != nil {
		t.Fatal(err)
	}
	h, _ := rev.Buscar("h-2")
	if h.Estado != revision.Descartado {
		t.Errorf("el hallazgo quedó en %q", h.Estado)
	}
	if h.Motivo == "" {
		t.Error("no guardó el motivo")
	}
	if len(rev.Abiertos()) != 0 {
		t.Error("sigue contando como abierto")
	}
	if p.e.Features["f-1"].IntentosFallidos != 0 {
		t.Error("no reseteó el contador de ME TRABÉ")
	}
}

// Descartar sin decir por qué es perder la razón por la que se descartó.
func TestDismissExigeMotivo(t *testing.T) {
	p := nuevo(t).productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Revision}

	if ef := Descartar(p.raiz, p.e, p.r, "h-1", ""); ef.Pasa() {
		t.Error("descartó sin motivo")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Archivar con git de verdad — el orden importa y antes estaba al revés
// ────────────────────────────────────────────────────────────────────────────

// listoParaArchivar deja f-1 en `cierre`, con constitución, branch y los dos
// archivos del ㉓. Es el estado exacto en el que llega un `sf approve` real.
func (p *proyecto) listoParaArchivar() *proyecto {
	p.t.Helper()
	p.productoListo()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Cierre}

	p.conArchivoConTexto(docs.Constitucion,
		"---\nlenguaje: go\ntest_cmd: exit 0\ngit:\n"+
			"  branch_por_feature: true\n  patron_branch: \"feat/{feature-id}-{slug}\"\n"+
			"  merge: no-ff\n  branch_base: main\n---\n# reglas\n")

	carpeta := filepath.Join(".docs", "features", "f-1-nucleo")
	p.conArchivoConTexto(filepath.Join(carpeta, "doc.md"), "# doc\n")
	p.conArchivoConTexto(filepath.Join(carpeta, "journal.md"), "- lección\n")

	p.conGit()
	// La branch de la feature, con un commit propio: es lo que el merge trae.
	corrergit(p.t, p.raiz, "checkout", "-b", "feat/f-1-nucleo")
	p.conArchivoConTexto("codigo.go", "package a\n")
	corrergit(p.t, p.raiz, "add", "-A")
	corrergit(p.t, p.raiz, "commit", "-m", "feat: el código")
	return p
}

// EL CASO NORMAL, y era el que fallaba SIEMPRE.
//
// `.docs/estado.json` está sucio por construcción cuando se llega al ㉓ —sf lo
// escribe en cada transición y sólo lo commitea al cerrar un lote—, así que el
// `git checkout main` abortaba. Y como el movimiento de la carpeta iba PRIMERO,
// el repo quedaba con la carpeta archivada, la feature sin cerrar y la branch
// sin mergear: un estado del que no se salía, porque el segundo intento fallaba
// en el rename.
func TestArchivarConElEstadoSucioIgualCierraLaFeature(t *testing.T) {
	p := nuevo(t).listoParaArchivar()
	// Lo que sf deja sucio en la vida real, puesto a mano.
	p.conArchivoConTexto(estado.Archivo, `{"producto":{},"features":{}}`)

	ef := Aprobar(p.raiz, p.e, p.r)
	if !ef.Pasa() {
		t.Fatalf("no archivó con el estado sucio: %v", ef.Fallas)
	}
	if p.e.Features["f-1"].Estado != estado.Cerrada {
		t.Errorf("la feature quedó en %q", p.e.Features["f-1"].Estado)
	}
	if b := branchActual(t, p.raiz); b != "main" {
		t.Errorf("quedó parado en %q, quería main", b)
	}
	// El merge trajo el código de la feature, y la branch se borró.
	if _, err := os.Stat(filepath.Join(p.raiz, "codigo.go")); err != nil {
		t.Errorf("el merge no trajo el código: %v", err)
	}
	// El trabajo que estaba pendiente quedó en un commit propio: es lo que
	// destraba el checkout, y de paso es trabajo real que si no queda huérfano.
	if !strings.Contains(gitLog(t, p.raiz), "chore: cierre de f-1") {
		t.Error("no commiteó lo que estaba pendiente antes de archivar")
	}
	// Y pide el commit del archivado, que lo hace main cuando el estado.json ya
	// está escrito y puede entrar en el mismo commit. Sin esto, el `git add -A`
	// del primer lote de la feature siguiente se lleva puesto este archivado.
	if ef.Commit == "" {
		t.Error("no pidió commitear el archivado")
	}
	// Y la carpeta viajó entera.
	if _, err := os.Stat(filepath.Join(p.raiz, docs.Archivado, "f-1-nucleo", "journal.md")); err != nil {
		t.Errorf("el journal no viajó: %v", err)
	}
}

// LO IRREVERSIBLE VA ÚLTIMO: si el git falla, la carpeta NO se movió y
// `sf approve` se puede reintentar tal cual.
//
// Es la mitad del arreglo que no se ve cuando todo sale bien, y la que convertía
// un error recuperable en un repo trabado.
func TestArchivarNoMueveLaCarpetaSiElMergeFalla(t *testing.T) {
	p := nuevo(t).listoParaArchivar()

	// Un conflicto de verdad: main toca el mismo archivo que la branch.
	corrergit(t, p.raiz, "checkout", "main")
	p.conArchivoConTexto("codigo.go", "package a // otra cosa\n")
	corrergit(t, p.raiz, "add", "-A")
	corrergit(t, p.raiz, "commit", "-m", "otro cambio")
	corrergit(t, p.raiz, "checkout", "feat/f-1-nucleo")

	ef := Aprobar(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("archivó con el merge fallado")
	}
	if _, err := os.Stat(filepath.Join(p.raiz, ".docs", "features", "f-1-nucleo")); err != nil {
		t.Errorf("movió la carpeta aunque el merge falló: %v", err)
	}
	if _, err := os.Stat(filepath.Join(p.raiz, docs.Archivado, "f-1-nucleo")); err == nil {
		t.Error("la carpeta terminó archivada con el merge fallado")
	}
	if p.e.Features["f-1"].Estado != estado.Cierre {
		t.Errorf("movió el estado: quedó en %q", p.e.Features["f-1"].Estado)
	}
}

// Y si un repo YA quedó a medias con la versión vieja —carpeta archivada,
// feature sin cerrar—, `sf approve` tiene que poder terminar el trabajo en vez
// de fallar en el rename. Es lo único que destraba a los que ya se comieron el
// bug.
func TestArchivarEsIdempotenteSiLaCarpetaYaEstaba(t *testing.T) {
	p := nuevo(t).listoParaArchivar()

	// El estado en que quedaba la versión vieja: la carpeta ya movida.
	origen := filepath.Join(p.raiz, ".docs", "features", "f-1-nucleo")
	destino := filepath.Join(p.raiz, docs.Archivado, "f-1-nucleo")
	if err := os.MkdirAll(filepath.Dir(destino), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(origen, destino); err != nil {
		t.Fatal(err)
	}

	ef := Aprobar(p.raiz, p.e, p.r)
	if !ef.Pasa() {
		t.Fatalf("no pudo terminar de archivar algo ya movido: %v", ef.Fallas)
	}
	if p.e.Features["f-1"].Estado != estado.Cerrada {
		t.Errorf("la feature quedó en %q", p.e.Features["f-1"].Estado)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// sf approve corre la compuerta antes de sellar
// ────────────────────────────────────────────────────────────────────────────

// conPlanCompleto deja una planificación que pasa las cinco compuertas del ⑰.
func (p *proyecto) conPlanCompleto() *proyecto {
	p.t.Helper()
	carpeta := filepath.Join(".docs", "features", "f-1-nucleo")
	p.conArchivoConTexto(docs.Historia("us-1"),
		"---\nid: us-1\n---\n## Criterios\n- **CA-1** — uno\n- **CA-2** — dos\n")
	p.conArchivoConTexto(filepath.Join(carpeta, docs.Decision),
		"## A — una\n## B — otra\n## C — la tercera\n")
	p.conArchivoConTexto(filepath.Join(carpeta, docs.Spec), "# spec\n")
	p.conArchivoConTexto(filepath.Join(carpeta, "tareas.json"),
		`{"tareas":[{"id":"t-1","lote":1,"satisface":["us-1/CA-1","us-1/CA-2"],
		  "tests":["a_test.go::TestX","a_test.go::TestY"]}]}`)
	return p
}

// La constitución sin `test_cmd` no se puede sellar, y ESO ES LO QUE DICE LA
// DOCUMENTACIÓN — pero `sf approve` la sellaba igual, porque no corría ninguna
// compuerta. Sin test_cmd caen tres: el rojo, el verde y el conteo de tests.
func TestApproveNoSellaLaConstitucionSinTestCmd(t *testing.T) {
	p := nuevo(t)
	p.e.Producto = estado.Producto{BriefSellado: "hacelo", PrdHash: "a3f9c1"}
	p.conArchivoConTexto(docs.Constitucion, "---\nlenguaje: go\ntest_cmd: \"\"\n---\n# reglas\n")

	ef := Aprobar(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("selló la constitución sin test_cmd")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "test_cmd") {
		t.Errorf("no dijo qué falta: %v", ef.Fallas)
	}
	if p.e.Producto.ConstitucionSellada {
		t.Error("marcó el sello igual")
	}

	// Y con test_cmd sí sella: el arreglo no puede frenar el flujo normal.
	p.conArchivoConTexto(docs.Constitucion, "---\nlenguaje: go\ntest_cmd: go test ./...\n---\n# reglas\n")
	if ef := Aprobar(p.raiz, p.e, p.r); !ef.Pasa() {
		t.Errorf("no selló una constitución completa: %v", ef.Fallas)
	}
}

// Y el peor de los dos: un backlog sellado sin ids de criterio DESARMA EL
// MECANISMO ENTERO. Sin ids, la cobertura del ⑰ cuenta cero contra cero y pasa,
// y el conteo de veredictos del ㉑ también.
func TestApproveNoSellaElBacklogSinCriterios(t *testing.T) {
	p := nuevo(t)
	p.e.Producto = estado.Producto{
		BriefSellado: "hacelo", PrdHash: "a3f9c1", ConstitucionSellada: true,
	}
	p.conArchivoConTexto(docs.Historia("us-1"), "---\nid: us-1\n---\nComo usuario quiero algo.\n")

	ef := Aprobar(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("selló un backlog con una historia sin criterios")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "criterios de aceptación") {
		t.Errorf("no dijo qué falta: %v", ef.Fallas)
	}
	if p.e.Producto.BacklogVisto {
		t.Error("marcó el backlog como visto igual")
	}

	p.conArchivoConTexto(docs.Historia("us-1"),
		"---\nid: us-1\n---\n## Criterios\n- **CA-1** — algo\n")
	if ef := Aprobar(p.raiz, p.e, p.r); !ef.Pasa() {
		t.Errorf("no selló un backlog con criterios: %v", ef.Fallas)
	}
}

// El ⑰ tipeado directo también pasa por las cinco. `sf next` ya las corre antes
// de ofrecer la 🛑, pero un `sf approve` a mano las salteaba — y aprobar un plan
// que cubre 7 de 9 criterios es justo lo que el ⑰ existe para impedir.
func TestApproveNoApruebaUnPlanIncompleto(t *testing.T) {
	p := nuevo(t).productoListo().conPlanCompleto()
	p.e.FeatureActual = "f-1"
	p.e.Features["f-1"] = &estado.Feature{Estado: estado.Planificacion}

	// Una tarea que cubre CA-1 y deja CA-2 afuera.
	p.conArchivoConTexto(filepath.Join(".docs", "features", "f-1-nucleo", "tareas.json"),
		`{"tareas":[{"id":"t-1","lote":1,"satisface":["us-1/CA-1"],"tests":["a_test.go::TestX"]}]}`)

	ef := Aprobar(p.raiz, p.e, p.r)
	if ef.Pasa() {
		t.Fatal("aprobó un plan que deja un criterio sin cubrir")
	}
	if !strings.Contains(strings.Join(ef.Fallas, " "), "us-1/CA-2") {
		t.Errorf("no dijo cuál falta: %v", ef.Fallas)
	}
	if p.e.Features["f-1"].Estado != estado.Planificacion {
		t.Error("movió el estado igual")
	}
}

// Los avisos de la compuerta NO frenan y viajan igual: esconderlos detrás de un
// ✓ es la forma más fácil de que nadie los lea.
func TestApproveMuestraLosAvisosDeLaCompuerta(t *testing.T) {
	p := nuevo(t)
	p.e.Producto = estado.Producto{
		BriefSellado: "hacelo", PrdHash: "a3f9c1", ConstitucionSellada: true,
	}
	// Seis historias en una feature: la compuerta del roadmap avisa, no frena.
	for _, id := range []string{"us-1", "us-2", "us-3"} {
		p.conArchivoConTexto(docs.Historia(id),
			"---\nid: "+id+"\n---\n## Criterios\n- **CA-1** — algo\n")
	}

	ef := Aprobar(p.raiz, p.e, p.r)
	if !ef.Pasa() {
		t.Fatalf("no selló: %v", ef.Fallas)
	}
	// Sin avisos que mostrar, el texto arranca con el ✓ y no con un ⚠ vacío.
	if !strings.HasPrefix(ef.Texto(), "✓ ") {
		t.Errorf("el texto arranca raro: %q", ef.Texto())
	}
}
