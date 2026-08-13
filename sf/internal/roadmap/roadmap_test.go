package roadmap

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ejemplo es el roadmap.json de artefactos.md §8, con las features al revés
// para que el test de ordenado tenga algo que hacer.
const ejemplo = `{
  "actualizado": "2026-08-11",
  "features": [
    { "id": "f-2", "slug": "verificacion", "nombre": "verificación",
      "orden": 2, "historias": ["us-2", "us-5"] },
    { "id": "f-1", "slug": "nucleo-cli", "nombre": "núcleo del CLI",
      "orden": 1, "historias": ["us-1", "us-3", "us-7"] }
  ]
}`

func escribirRoadmap(t *testing.T, contenido string) string {
	t.Helper()

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
	r, err := Leer(escribirRoadmap(t, ejemplo))
	if err != nil {
		t.Fatalf("Leer devolvió error: %v", err)
	}

	if len(r.Features) != 2 {
		t.Fatalf("hay %d features, quería 2", len(r.Features))
	}
	f := r.Features[0]
	if f.ID != "f-1" {
		t.Errorf("la primera es %q, quería f-1", f.ID)
	}
	if f.Nombre != "núcleo del CLI" {
		t.Errorf("Nombre = %q", f.Nombre)
	}
	if len(f.Historias) != 3 {
		t.Errorf("f-1 tiene %d historias, quería 3", len(f.Historias))
	}
}

// El archivo puede venir en cualquier orden —lo escribe un LLM— y Leer tiene
// que devolverlo por `orden`. Si no, la cola del ⑪ sale mal.
func TestLeerOrdenaAunqueElArchivoVengaDesordenado(t *testing.T) {
	r, err := Leer(escribirRoadmap(t, ejemplo))
	if err != nil {
		t.Fatal(err)
	}

	// En el JSON f-2 está primera; después de Leer tiene que estar segunda.
	quiero := []string{"f-1", "f-2"}
	for i, id := range quiero {
		if r.Features[i].ID != id {
			t.Errorf("posición %d: %q, quería %q", i, r.Features[i].ID, id)
		}
	}
}

func TestCarpeta(t *testing.T) {
	f := Feature{ID: "f-1", Slug: "nucleo-cli"}

	// filepath.Join usa el separador del sistema, así que se compara contra un
	// Join y no contra un string con barras a mano.
	quiero := filepath.Join(".docs", "features", "f-1-nucleo-cli")
	if got := f.Carpeta(); got != quiero {
		t.Errorf("Carpeta() = %q, quería %q", got, quiero)
	}
}

func TestLeerSinArchivoDaErrNoHay(t *testing.T) {
	_, err := Leer(t.TempDir())
	if !errors.Is(err, ErrNoHay) {
		t.Fatalf("Leer sin archivo devolvió %v, quería ErrNoHay", err)
	}
}

// Las tres validaciones son las tres formas en que el roadmap puede hacer que
// sf trabaje mal EN SILENCIO. Todo lo demás no se valida a propósito.
func TestValidacionesQueImportan(t *testing.T) {
	casos := []struct {
		nombre    string
		json      string
		enElError string // un pedazo que el mensaje tiene que mencionar
	}{
		{
			nombre: "id repetido: sf trabajaría sobre la feature equivocada",
			json: `{"features": [
				{"id": "f-1", "orden": 1, "historias": ["us-1"]},
				{"id": "f-1", "orden": 2, "historias": ["us-2"]}
			]}`,
			enElError: "repetido",
		},
		{
			nombre: "orden repetido: la cola queda indeterminada",
			json: `{"features": [
				{"id": "f-1", "orden": 1, "historias": ["us-1"]},
				{"id": "f-2", "orden": 1, "historias": ["us-2"]}
			]}`,
			enElError: "mismo orden",
		},
		{
			nombre: "feature sin historias: no hay qué implementar",
			json: `{"features": [
				{"id": "f-1", "orden": 1, "historias": []}
			]}`,
			enElError: "no tiene historias",
		},
		{
			nombre:    "feature sin id",
			json:      `{"features": [{"orden": 1, "historias": ["us-1"]}]}`,
			enElError: "sin id",
		},
	}

	for _, c := range casos {
		// t.Run le da nombre propio a cada caso: si falla uno, el output dice
		// cuál y no hay que contar posiciones.
		t.Run(c.nombre, func(t *testing.T) {
			_, err := Leer(escribirRoadmap(t, c.json))
			if err == nil {
				t.Fatal("no dio error")
			}
			if !strings.Contains(err.Error(), c.enElError) {
				t.Errorf("el mensaje fue %q y no menciona %q", err, c.enElError)
			}
		})
	}
}

// Un roadmap sano no tiene que dar error, aunque le falten campos que sf no usa
// para decidir (nombre, actualizado).
func TestRoadmapMinimoEsValido(t *testing.T) {
	_, err := Leer(escribirRoadmap(t, `{"features": [
		{"id": "f-1", "orden": 1, "historias": ["us-1"]}
	]}`))
	if err != nil {
		t.Errorf("un roadmap mínimo dio error: %v", err)
	}
}

func TestBuscar(t *testing.T) {
	r, err := Leer(escribirRoadmap(t, ejemplo))
	if err != nil {
		t.Fatal(err)
	}

	f, ok := r.Buscar("f-2")
	if !ok {
		t.Fatal("Buscar(f-2) no la encontró")
	}
	if f.Slug != "verificacion" {
		t.Errorf("Slug = %q, quería verificacion", f.Slug)
	}

	if _, ok := r.Buscar("f-99"); ok {
		t.Error("Buscar(f-99) encontró algo que no existe")
	}
}

// Proxima es la mitad del ⑪. La otra mitad —qué está cerrado— la pone quien
// llama, y por eso este paquete nunca lee estado.json.
func TestProxima(t *testing.T) {
	r, err := Leer(escribirRoadmap(t, ejemplo))
	if err != nil {
		t.Fatal(err)
	}

	f, ok := r.Proxima(nil)
	if !ok || f.ID != "f-1" {
		t.Errorf("sin nada cerrado, Proxima() = %q, quería f-1", f.ID)
	}

	f, ok = r.Proxima(map[string]bool{"f-1": true})
	if !ok || f.ID != "f-2" {
		t.Errorf("con f-1 cerrada, Proxima() = %q, quería f-2", f.ID)
	}

	if _, ok := r.Proxima(map[string]bool{"f-1": true, "f-2": true}); ok {
		t.Error("con todo cerrado, Proxima() devolvió una feature")
	}
}
