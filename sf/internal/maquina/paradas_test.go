package maquina

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
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
	p := nuevo(t).productoListo()
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
		Estado: estado.Implementar, Modelo: "deepseek", IntentosFallidos: 3,
	}

	ef := Modelo(p.e, nil, "opus", "", "")
	if !ef.Pasa() {
		t.Fatalf("%v", ef.Fallas)
	}
	if p.e.Features["f-1"].Modelo != "opus" {
		t.Errorf("modelo = %q", p.e.Features["f-1"].Modelo)
	}
	if p.e.Features["f-1"].IntentosFallidos != 0 {
		t.Errorf("intentos_fallidos = %d, quería 0", p.e.Features["f-1"].IntentosFallidos)
	}
	if !strings.Contains(ef.Mensaje, "deepseek") {
		t.Errorf("el mensaje no dice de qué modelo venía: %q", ef.Mensaje)
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
