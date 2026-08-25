package historia

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/docs"
)

// escribirHistoria deja un us-# en disco. Archivos de verdad y no mocks: media
// decisión de este paquete es "¿se puede leer este archivo?", y un mock las
// haría pasar todas sin probar que la ruta que arma docs.Historia sea la real.
func escribirHistoria(t *testing.T, raiz, id, tipo string) {
	t.Helper()
	ruta := filepath.Join(raiz, docs.Historia(id))
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		t.Fatal(err)
	}
	cuerpo := "---\ntipo: " + tipo + "\nid: " + id + "\n---\n- **CA-1** — x\n"
	if err := os.WriteFile(ruta, []byte(cuerpo), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// El camino corto — quién decide que una feature es de bugs
// ────────────────────────────────────────────────────────────────────────────

// SonTodasBugs exige TODAS y no "alguna": saltear la planificación de una
// feature que mezcla un bug con historias nuevas dejaría esas historias sin
// diseño (artefactos.md §3).
func TestSonTodasBugs(t *testing.T) {
	raiz := t.TempDir()
	escribirHistoria(t, raiz, "us-1", "bug")
	escribirHistoria(t, raiz, "us-2", "bug")
	escribirHistoria(t, raiz, "us-3", "us")

	casos := []struct {
		nombre string
		ids    []string
		quiero bool
	}{
		{"todas bugs", []string{"us-1", "us-2"}, true},
		{"un solo bug", []string{"us-1"}, true},
		{"mezcladas no saltean", []string{"us-1", "us-3"}, false},
		{"ninguna es bug", []string{"us-3"}, false},
		{"sin historias no saltea", nil, false},
		// Ante la duda, el camino largo: de más se puede saltear después, de
		// menos ya se implementó sin diseño.
		{"una que no se puede leer no saltea", []string{"us-1", "us-99"}, false},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			if got := SonTodasBugs(raiz, c.ids); got != c.quiero {
				t.Errorf("SonTodasBugs(%v) = %v, quería %v", c.ids, got, c.quiero)
			}
		})
	}
}
