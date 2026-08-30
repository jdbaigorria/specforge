package sobre

import (
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// catalogoDePrueba es un menú con dos perfiles y tres modelos.
//
// Los alias son genéricos —`grande`, `medio`, `barato`— y no nombres de modelos
// reales: si un test dijera un nombre de proveedor, el que lo lea dentro de un
// año creería que sf conoce ese nombre. Es el mismo invariante que en `global`.
func catalogoDePrueba() *global.Config {
	c := global.Semilla("opencode")
	c.Declarar(global.Razonar, global.Modelo{
		Alias: "grande", ID: "prov/grande-4", Via: global.Subagente,
		Capacidad: 3, Para: "diseño, concurrencia, hallar lo que falta",
	})
	c.Declarar(global.Construir, global.Modelo{
		Alias: "medio", ID: "prov/medio-2", Via: global.Subagente,
		Capacidad: 3, Para: "el caballo de batalla",
	})
	c.Declarar(global.Construir, global.Modelo{
		Alias: "barato", ID: "prov/chico-1", Via: global.Subagente,
		Capacidad: 1, Para: "CRUD, plomería y refactors mecánicos",
	})
	return c
}

// enPlanificacion deja el proyecto listo para pedir el sobre del ⑫.
func enPlanificacion(t *testing.T) *proyecto {
	t.Helper()
	return nuevo(t).enFeature("planificacion")
}

// EL TEST DE H7. El ⑯ elige con qué se implementa cada lote, y hasta que el menú
// entró al sobre lo hacía sin ver el catálogo: no elegía mal, no tenía de dónde
// elegir.
func TestElSobreDelPlanTraeElMenuDeModelos(t *testing.T) {
	p := enPlanificacion(t)
	p.g = catalogoDePrueba()
	txt := p.texto()

	if !strings.Contains(txt, TituloMenu) {
		t.Fatalf("el sobre del ⑫ no trae el menú:\n%s", txt)
	}
	for _, alias := range []string{"grande", "medio", "barato"} {
		if !strings.Contains(txt, alias) {
			t.Errorf("el menú no lista %q", alias)
		}
	}
	// El `para:` viaja TEXTUAL: es la opinión de Javier llegando entera, que es
	// lo único que R3 le permite a sf hacer con ella.
	if !strings.Contains(txt, "CRUD, plomería y refactors mecánicos") {
		t.Error("el menú no lleva el `para:` de Javier")
	}
	if !strings.Contains(txt, "capacidad 1") {
		t.Error("el menú no lleva la capacidad")
	}
}

// LA GARANTÍA MÁS BARATA QUE HAY de que el ⑯ no escriba un id en tareas.json:
// que no los vea. Un id en un archivo versionado es un dato de esta máquina que
// otro harness va a leer, y eso es H2 otra vez.
func TestElMenuNoLlevaNingunId(t *testing.T) {
	p := enPlanificacion(t)
	p.g = catalogoDePrueba()
	txt := p.texto()

	for _, id := range []string{"prov/grande-4", "prov/medio-2", "prov/chico-1"} {
		if strings.Contains(txt, id) {
			t.Errorf("el menú filtró el id %q — el ⑯ lo va a copiar a tareas.json", id)
		}
	}
}

// El menú marca cuál es el default de cada perfil. Sin eso, "no digo nada"
// parece azar en vez de una elección, y el ⑯ elige por las dudas.
func TestElMenuMarcaElDefaultDeCadaPerfil(t *testing.T) {
	p := enPlanificacion(t)
	p.g = catalogoDePrueba()

	txt := p.texto()
	i, j := strings.Index(txt, "→ medio"), strings.Index(txt, "barato")
	if i < 0 {
		t.Fatalf("no marcó el default de construir:\n%s", txt)
	}
	if j > 0 && strings.Contains(txt[i:j], "→ barato") {
		t.Error("marcó dos defaults en el mismo perfil")
	}
}

// El ⑱ NO recibe el menú: ahí ya está decidido, y ofrecerle el catálogo al que
// implementa es invitarlo a cambiar una decisión que no es suya.
func TestElSobreDeImplementarNoTraeElMenu(t *testing.T) {
	p := nuevo(t).enFeature("implementar")
	p.g = catalogoDePrueba()

	if txt := p.texto(); strings.Contains(txt, TituloMenu) {
		t.Errorf("el ⑱ recibió el menú de modelos:\n%s", txt)
	}
}

// Sin catálogo el sobre NO falla: lo dice. Es la regla de no mentir por omisión
// que `Parte.Falta` ya tenía — si la parte no apareciera, el ⑯ creería que no
// hacía falta elegir.
func TestSinCatalogoElMenuLoDiceYNoFalla(t *testing.T) {
	p := enPlanificacion(t)
	p.g = nil

	txt := p.texto()
	if !strings.Contains(txt, TituloMenu) {
		t.Fatalf("la parte desapareció en vez de explicarse:\n%s", txt)
	}
	if !strings.Contains(txt, "sf install") {
		t.Errorf("no dijo por qué falta:\n%s", txt)
	}
}

// Con el catálogo vacío —el harness declarado pero sin modelos— tampoco falla,
// y avisa que `sf next` va a parar antes de lanzar a nadie.
func TestConCatalogoVacioElMenuAvisa(t *testing.T) {
	p := enPlanificacion(t)
	p.g = global.Semilla("opencode")

	txt := p.texto()
	if !strings.Contains(txt, "ningún modelo declarado") {
		t.Errorf("no avisó que el catálogo está vacío:\n%s", txt)
	}
}
