package estado

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ejemplo es, letra por letra, el estado.json de maquina-estados.md §9.
//
// Que el test use el JSON del documento y no uno inventado es a propósito: si
// alguien cambia un nombre de campo en el código, este test falla y obliga a
// mirar el diseño. Es la forma barata de que el documento y el código no se
// separen.
const ejemplo = `{
  "producto": {
    "brief_sellado": "hacelo",
    "prd_hash": "a3f9c1",
    "constitucion_sellada": true
  },
  "feature_actual": "f-2",
  "features": {
    "f-1": { "estado": "cerrada", "base_commit": "7d1b40" },
    "f-2": {
      "estado": "implementando",
      "base_commit": "9c2e1a",
      "modelo": "deepseek",
      "intentos_fallidos": 0,
      "lotes": [
        { "lote": 1, "rojo": true, "hash_tests": "b70c33", "commit": "4a8f21" },
        { "lote": 2, "rojo": true, "hash_tests": "4e91b7", "commit": null }
      ]
    }
  }
}`

// escribirEstado deja un estado.json en un directorio temporal y devuelve la
// raíz. t.TempDir() se limpia solo cuando el test termina, así que no hay que
// acordarse de borrar nada.
func escribirEstado(t *testing.T, contenido string) string {
	t.Helper() // hace que los errores se reporten en la línea que llamó, no acá

	raiz := t.TempDir()
	ruta := filepath.Join(raiz, Archivo)
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(contenido), 0o644); err != nil {
		t.Fatal(err)
	}
	return raiz
}

func TestLeerElEjemploDelDiseno(t *testing.T) {
	e, err := Leer(escribirEstado(t, ejemplo))
	if err != nil {
		t.Fatalf("Leer devolvió error: %v", err)
	}

	if e.Producto.BriefSellado != "hacelo" {
		t.Errorf("BriefSellado = %q, quería %q", e.Producto.BriefSellado, "hacelo")
	}
	if e.Producto.PrdHash != "a3f9c1" {
		t.Errorf("PrdHash = %q, quería %q", e.Producto.PrdHash, "a3f9c1")
	}
	if !e.Producto.ConstitucionSellada {
		t.Error("ConstitucionSellada = false, quería true")
	}
	if e.FeatureActual != "f-2" {
		t.Errorf("FeatureActual = %q, quería %q", e.FeatureActual, "f-2")
	}
	if len(e.Features) != 2 {
		t.Fatalf("Features tiene %d entradas, quería 2", len(e.Features))
	}

	f2 := e.Features["f-2"]
	if f2.Modelo != "deepseek" {
		t.Errorf("f-2.Modelo = %q, quería %q", f2.Modelo, "deepseek")
	}
	if len(f2.Lotes) != 2 {
		t.Fatalf("f-2 tiene %d lotes, quería 2", len(f2.Lotes))
	}
}

// El `commit: null` del lote 2 es la razón por la que Commit es *string y no
// string. Este test es el que se rompe si alguien "simplifica" el tipo.
func TestCommitNullSeDistingueDeVacio(t *testing.T) {
	e, err := Leer(escribirEstado(t, ejemplo))
	if err != nil {
		t.Fatal(err)
	}
	lotes := e.Features["f-2"].Lotes

	if lotes[0].Commit == nil {
		t.Fatal("el lote 1 tiene commit en el JSON y llegó como nil")
	}
	if *lotes[0].Commit != "4a8f21" {
		t.Errorf("lote 1 commit = %q, quería %q", *lotes[0].Commit, "4a8f21")
	}
	if lotes[1].Commit != nil {
		t.Errorf("el lote 2 tiene commit: null y llegó como %q", *lotes[1].Commit)
	}
}

func TestLeerSinArchivoDaErrNoHay(t *testing.T) {
	_, err := Leer(t.TempDir())

	// errors.Is y no `err == ErrNoHay`: si mañana Leer envuelve el error con
	// %w, la comparación directa deja de funcionar y errors.Is sigue andando.
	if !errors.Is(err, ErrNoHay) {
		t.Fatalf("Leer sin archivo devolvió %v, quería ErrNoHay", err)
	}
}

func TestLeerJsonCorruptoNoSeConfundeConNoHay(t *testing.T) {
	_, err := Leer(escribirEstado(t, `{"producto": `))
	if err == nil {
		t.Fatal("un JSON cortado por la mitad no dio error")
	}
	// La distinción importa: "no hay estado" es parte normal del flujo (recién
	// corriste sf init) y "el estado está roto" es un problema que hay que
	// mostrarle a Javier. Confundirlos haría que sf arranque de cero en
	// silencio y se pierda el sello del ⑥.
	if errors.Is(err, ErrNoHay) {
		t.Error("un estado.json corrupto se reportó como si no existiera")
	}
}

// Guardar y volver a Leer tiene que devolver lo mismo. Es el test que cubre el
// par entero: si una etiqueta json está mal escrita, el dato se pierde en la
// vuelta y esto lo caza.
func TestGuardarYLeerConservaTodo(t *testing.T) {
	raiz := t.TempDir()
	commit := "4a8f21"

	original := &Estado{
		Producto: Producto{
			BriefSellado:        "hacelo",
			PrdHash:             "a3f9c1",
			ConstitucionSellada: true,
		},
		FeatureActual: "f-2",
		Features: map[string]*Feature{
			"f-1": {Estado: Cerrada, BaseCommit: "7d1b40"},
			"f-2": {
				Estado:           Implementar,
				BaseCommit:       "9c2e1a",
				Modelo:           "deepseek",
				IntentosFallidos: 2,
				Lotes: []Lote{
					{Lote: 1, Rojo: true, HashTests: "b70c33", Commit: &commit},
					{Lote: 2, Rojo: true, HashTests: "4e91b7", Commit: nil},
				},
			},
		},
	}

	if err := original.Guardar(raiz); err != nil {
		t.Fatalf("Guardar: %v", err)
	}

	vuelta, err := Leer(raiz)
	if err != nil {
		t.Fatalf("Leer después de Guardar: %v", err)
	}

	if vuelta.Producto != original.Producto {
		t.Errorf("Producto cambió en la vuelta:\n  guardé %+v\n  leí    %+v",
			original.Producto, vuelta.Producto)
	}
	if vuelta.FeatureActual != "f-2" {
		t.Errorf("FeatureActual = %q, quería %q", vuelta.FeatureActual, "f-2")
	}

	f2 := vuelta.Features["f-2"]
	if f2.IntentosFallidos != 2 {
		t.Errorf("IntentosFallidos = %d, quería 2", f2.IntentosFallidos)
	}
	if f2.Lotes[1].Commit != nil {
		t.Error("el lote sin commit volvió con commit")
	}
	if *f2.Lotes[0].Commit != commit {
		t.Errorf("commit del lote 1 = %q, quería %q", *f2.Lotes[0].Commit, commit)
	}
}

// Guardar tiene que crear .docs/ si no existe: es lo que pasa la primera vez,
// cuando sf init todavía no corrió o corrió a medias.
func TestGuardarCreaElDirectorio(t *testing.T) {
	raiz := t.TempDir()

	if err := (&Estado{}).Guardar(raiz); err != nil {
		t.Fatalf("Guardar en un directorio vacío: %v", err)
	}
	if _, err := os.Stat(filepath.Join(raiz, Archivo)); err != nil {
		t.Fatalf("el archivo no quedó escrito: %v", err)
	}
}

// El temporal de la escritura atómica no tiene que quedar tirado.
func TestGuardarNoDejaBasura(t *testing.T) {
	raiz := t.TempDir()
	if err := (&Estado{}).Guardar(raiz); err != nil {
		t.Fatal(err)
	}

	entradas, err := os.ReadDir(filepath.Join(raiz, ".docs"))
	if err != nil {
		t.Fatal(err)
	}
	if len(entradas) != 1 {
		var nombres []string
		for _, e := range entradas {
			nombres = append(nombres, e.Name())
		}
		t.Errorf("quedaron %d archivos en .docs/, quería 1: %v", len(entradas), nombres)
	}
}

// Un estado.json sin la clave "features" no tiene que dejar el mapa en nil:
// escribir en un mapa nil es panic, y quien llame no tiene por qué chequearlo.
func TestLeerSinFeaturesDejaElMapaUsable(t *testing.T) {
	e, err := Leer(escribirEstado(t, `{"feature_actual": ""}`))
	if err != nil {
		t.Fatal(err)
	}
	// Si el mapa viniera nil, esta línea sería un panic y el test se caería.
	e.Features["f-1"] = &Feature{Estado: Planificacion}
}

func TestActual(t *testing.T) {
	e, err := Leer(escribirEstado(t, ejemplo))
	if err != nil {
		t.Fatal(err)
	}

	f, ok := e.Actual()
	if !ok {
		t.Fatal("Actual() dijo que no hay feature en curso, y feature_actual es f-2")
	}
	if f.Modelo != "deepseek" {
		t.Errorf("Actual() devolvió la feature equivocada: modelo %q", f.Modelo)
	}

	e.FeatureActual = ""
	if _, ok := e.Actual(); ok {
		t.Error("Actual() devolvió una feature con feature_actual vacío")
	}
}

// El lote actual es el primero sin commit. No hay campo que lo diga: se deduce.
func TestLoteActualEsElPrimeroSinCommit(t *testing.T) {
	e, err := Leer(escribirEstado(t, ejemplo))
	if err != nil {
		t.Fatal(err)
	}
	f := e.Features["f-2"]

	l, ok := f.LoteActual()
	if !ok {
		t.Fatal("LoteActual() no encontró ninguno, y el lote 2 no tiene commit")
	}
	if l.Lote != 2 {
		t.Errorf("LoteActual() devolvió el %d, quería el 2", l.Lote)
	}

	// El puntero tiene que apuntar al lote de verdad, no a una copia: si
	// devolviera la variable del range, esta escritura se perdería.
	l.Rojo = false
	if f.Lotes[1].Rojo {
		t.Error("LoteActual() devolvió una copia: mutarla no cambió el estado")
	}
}

func TestLoteActualCuandoEstanTodosCerrados(t *testing.T) {
	c := "abc123"
	f := &Feature{Lotes: []Lote{{Lote: 1, Commit: &c}}}

	if _, ok := f.LoteActual(); ok {
		t.Error("LoteActual() encontró uno abierto y están todos commiteados")
	}
	if f.Cerrados() != 1 {
		t.Errorf("Cerrados() = %d, quería 1", f.Cerrados())
	}
}

func TestCerrados(t *testing.T) {
	e, err := Leer(escribirEstado(t, ejemplo))
	if err != nil {
		t.Fatal(err)
	}
	// Es la mitad de "lote 2 de 4" que sf next le muestra al orquestador.
	if n := e.Features["f-2"].Cerrados(); n != 1 {
		t.Errorf("Cerrados() = %d, quería 1", n)
	}
}

// SinSembrar separa los dos casos que LoteActual() mete en la misma bolsa.
//
// "Nadie empezó" y "todos terminaron" devuelven las dos (nil, false), y
// confundirlas dejaba pasar una feature sin una sola línea de código. El test
// pone los tres estados posibles al lado para que la diferencia se vea.
func TestSinSembrarNoEsLoMismoQueTodosCerrados(t *testing.T) {
	c := "abc123"

	nadie := &Feature{}
	empezada := &Feature{Lotes: []Lote{{Lote: 1, Rojo: true}}}
	terminada := &Feature{Lotes: []Lote{{Lote: 1, Rojo: true, Commit: &c}}}

	if !nadie.SinSembrar() {
		t.Error("una feature sin lotes tiene que dar SinSembrar()")
	}
	if empezada.SinSembrar() || terminada.SinSembrar() {
		t.Error("una feature con lotes NO está sin sembrar")
	}

	// Y la trampa, explícita: los dos extremos dan el mismo LoteActual().
	_, hayNadie := nadie.LoteActual()
	_, hayTerminada := terminada.LoteActual()
	if hayNadie || hayTerminada {
		t.Fatal("el test se apoya en que los dos den false")
	}
	if nadie.SinSembrar() == terminada.SinSembrar() {
		t.Error("SinSembrar() existe justamente para distinguirlos")
	}
}

// El camino corto tiene que dar lo mismo desde donde se lo mire, y esa es toda
// la razón por la que Efectivo existe: la regla vivía copiada en `sf next` y en
// `sf done`, y faltaba en `sf lote start`.
func TestEfectivoSalteaPlanificacionSoloParaBugs(t *testing.T) {
	casos := []struct {
		nombre string
		actual string
		esBug  bool
		quiero string
	}{
		{"un bug saltea la planificación", Planificacion, true, Implementar},
		{"una us no la saltea", Planificacion, false, Planificacion},
		{"un bug ya en implementar no se mueve", Implementar, true, Implementar},
		{"el cierre no se toca nunca", Cierre, true, Cierre},
		{"la revisión no se saltea acá: depende de los lotes", Revision, true, Revision},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := Efectivo(c.actual, c.esBug); got != c.quiero {
				t.Errorf("Efectivo(%q, %v) = %q, quería %q", c.actual, c.esBug, got, c.quiero)
			}
		})
	}
}
