package historia

import (
	"os"
	"path/filepath"
	"slices"
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

// ────────────────────────────────────────────────────────────────────────────
// El esqueleto de `sf new` — cuándo el ⑨ terminó, no cuándo empezó
// ────────────────────────────────────────────────────────────────────────────

// Un criterio con el id puesto y el texto vacío contaba como criterio, y eso es
// PEOR que no tener ninguno: el ⑰ lo da por cubierto y el ㉑ le pone veredicto,
// así que el mecanismo de los ids queda en pie sobre algo que nadie puede
// juzgar.
func TestSinTextoAtrapaElCriterioVacio(t *testing.T) {
	raiz := t.TempDir()
	escribirCuerpo(t, raiz, "us-1", "titulo: sumar", "## Criterios\n"+
		"- **CA-1** —\n"+
		"- **CA-2** — devuelve 5\n"+
		"- **CA-3**\n")

	h, err := Leer(raiz, "us-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(h.Criterios) != 3 {
		t.Fatalf("encontró %v, quería los tres ids", h.Criterios)
	}
	if !slices.Equal(h.SinTexto, []string{"CA-1", "CA-3"}) {
		t.Errorf("SinTexto = %v, quería [CA-1 CA-3]", h.SinTexto)
	}
}

// Y el que SÍ tiene texto no se marca, con las formas que usa un LLM: em dash,
// guión, dos puntos, con negritas y sin.
func TestSinTextoNoSeComeLosCriteriosEscritos(t *testing.T) {
	raiz := t.TempDir()
	escribirCuerpo(t, raiz, "us-1", "titulo: x", "## Criterios\n"+
		"- **CA-1** — con em dash\n"+
		"- **CA-2**: con dos puntos\n"+
		"- CA-3 - sin negritas\n"+
		"* **CA-4** — con asterisco de lista\n")

	h, err := Leer(raiz, "us-1")
	if err != nil {
		t.Fatal(err)
	}
	if len(h.SinTexto) != 0 {
		t.Errorf("marcó como vacíos a %v", h.SinTexto)
	}
}

// SinPinponear es el checkpoint del ⑨, y es lo que hace que `sf new` lleve a
// `sfp-backlog`. Antes el checkpoint preguntaba "¿hay historias?", y con un
// producto en marcha la respuesta es siempre sí: la ⏸ salía en vez del trabajo
// y lo que quedaba para aprobar era el esqueleto vacío.
func TestSinPinponearDetectaElEsqueletoDeNew(t *testing.T) {
	raiz := t.TempDir()

	// Completa: título y criterio con texto.
	escribirCuerpo(t, raiz, "us-1", "titulo: sumar dos numeros",
		"## Criterios\n- **CA-1** — Suma(2,3) da 5\n")
	// El esqueleto tal cual lo deja `sf new`.
	if err := os.WriteFile(filepath.Join(raiz, docs.Historia("us-2")),
		[]byte(Esqueleto("us-2", "abc123")), 0o644); err != nil {
		t.Fatal(err)
	}
	// Con título pero sin ningún criterio.
	escribirCuerpo(t, raiz, "us-3", "titulo: algo", "Como usuario quiero algo.\n")
	// Criterios completos, pero el hueco del esqueleto sigue en el cuerpo.
	escribirCuerpo(t, raiz, "us-4", "titulo: a medias",
		"Como **<quién>** quiero **buscar** para **encontrar**.\n\n"+
			"## Criterios\n- **CA-1** — busca\n")

	quiero := []string{"us-2", "us-3", "us-4"}
	if got := SinPinponear(raiz); !slices.Equal(got, quiero) {
		t.Errorf("SinPinponear = %v, quería %v", got, quiero)
	}
}

// Y con todo completo no queda nada por hacer: sin esto, el ⑨ mandaría a
// pinponear para siempre y la ⏸ nunca aparecería.
func TestSinPinponearNoDevuelveNadaConElBacklogCompleto(t *testing.T) {
	raiz := t.TempDir()
	escribirCuerpo(t, raiz, "us-1", "titulo: una", "## Criterios\n- **CA-1** — algo\n")
	escribirCuerpo(t, raiz, "us-2", "titulo: otra", "## Criterios\n- **CA-1** — otra cosa\n")

	if got := SinPinponear(raiz); len(got) != 0 {
		t.Errorf("SinPinponear = %v, quería vacío", got)
	}
}

// El checkpoint NO se ata a `titulo:` del frontmatter, y esto es lo que impide
// el falso positivo: una historia escrita a mano puede tener el título sólo en
// el encabezado y estar perfectamente completa. Un checkpoint que frena trabajo
// terminado es peor que no tenerlo — se aprende a ignorarlo.
func TestSinPinponearNoFrenaUnaHistoriaSinTituloEnElFrontmatter(t *testing.T) {
	raiz := t.TempDir()
	escribirCuerpo(t, raiz, "us-1", "tipo: us",
		"# us-1 — sumar dos números\n\n## Criterios\n- **CA-1** — Suma(2,3) da 5\n")

	if got := SinPinponear(raiz); len(got) != 0 {
		t.Errorf("SinPinponear = %v: la historia está completa", got)
	}
}

// escribirCuerpo deja un us-# con el frontmatter y el cuerpo que se le pidan.
func escribirCuerpo(t *testing.T, raiz, id, frontmatter, cuerpo string) {
	t.Helper()
	ruta := filepath.Join(raiz, docs.Historia(id))
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		t.Fatal(err)
	}
	texto := "---\nid: " + id + "\n" + frontmatter + "\n---\n\n" + cuerpo
	if err := os.WriteFile(ruta, []byte(texto), 0o644); err != nil {
		t.Fatal(err)
	}
}
