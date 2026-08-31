package pregunta

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/global"
	"github.com/jdbaigorria/specforge/sf/internal/modelos"
)

// catalogoFalso devuelve modelos sin correr ningún arnés.
func catalogoFalso(porArnes map[string][]modelos.Modelo) Listador {
	return func(h string) ([]modelos.Modelo, error) {
		ms, ok := porArnes[h]
		if !ok {
			return nil, errors.New("no pude listar")
		}
		return ms, nil
	}
}

var tresDeOpencode = map[string][]modelos.Modelo{
	"opencode": {
		{ID: "opencode/nemotron-3-ultra-free"},
		{ID: "opencode/mimo-v2.5-free"},
		{ID: "opencode/big-pickle"},
	},
}

// correr hace la pregunta con un guion de respuestas, una por línea.
func correr(t *testing.T, guion string, instalados []string, l Listador) (*Respuesta, string, error) {
	t.Helper()
	var out bytes.Buffer
	r, err := Preguntar(&out, strings.NewReader(guion), instalados, l)
	return r, out.String(), err
}

func TestElijeArnesesPorNumero(t *testing.T) {
	r, _, err := correr(t, "1,3\n\n\n", global.Harness, catalogoFalso(nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Harnesses) != 2 || r.Harnesses[0] != "claude-code" || r.Harnesses[1] != "commandcode" {
		t.Errorf("eligió %v", r.Harnesses)
	}
}

// Enter en la primera pregunta es "todos los que me mostraste", y los que se
// muestran son los que están instalados. Es el camino más corto para el caso
// normal: los tenés, los vas a usar.
func TestEnterEnLosArnesesEsTodos(t *testing.T) {
	r, _, err := correr(t, "\n\n\n\n\n\n\n", global.Harness, catalogoFalso(nil))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Harnesses) != 3 {
		t.Errorf("enter tenía que tomar los tres, tomó %v", r.Harnesses)
	}
}

func TestDeclaraUnModeloConAliasYEsfuerzo(t *testing.T) {
	// arneses · razonar: elige 1, alias, esfuerzo · construir: se saltea
	r, _, err := correr(t, "1\n1\nultra\nhigh\n\n", []string{"opencode"},
		catalogoFalso(tresDeOpencode))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Declaraciones) != 1 {
		t.Fatalf("esperaba una declaración, hubo %+v", r.Declaraciones)
	}
	d := r.Declaraciones[0]
	if d.Harness != "opencode" || d.Perfil != global.Razonar {
		t.Errorf("la declaración fue a %q/%q", d.Harness, d.Perfil)
	}
	if d.ID != "opencode/nemotron-3-ultra-free" {
		t.Errorf("id %q", d.ID)
	}
	if d.Alias != "ultra" || d.Esfuerzo != "high" {
		t.Errorf("alias %q esfuerzo %q", d.Alias, d.Esfuerzo)
	}
}

// El esfuerzo admite enter, y vacío NO es "bajo": es "el que traiga el modelo".
// La distinción ya está en `global.Modelo.Esfuerzo` y acá no se puede perder.
func TestEnterEnElEsfuerzoDejaVacioYNoUnValor(t *testing.T) {
	r, _, err := correr(t, "1\n2\nrapido\n\n\n", []string{"opencode"},
		catalogoFalso(tresDeOpencode))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Declaraciones) != 1 {
		t.Fatalf("%+v", r.Declaraciones)
	}
	if e := r.Declaraciones[0].Esfuerzo; e != "" {
		t.Errorf("enter puso %q y tenía que dejar vacío", e)
	}
}

// Escribir texto filtra. Es lo que reemplaza al filtrado en vivo de una pantalla
// raw: una tecla más —el enter— y el mismo resultado, sin dependencia nueva.
func TestUnTextoFiltraLaLista(t *testing.T) {
	r, out, err := correr(t, "1\nmimo\n1\nrapido\n\n\n", []string{"opencode"},
		catalogoFalso(tresDeOpencode))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "mimo-v2.5-free") {
		t.Errorf("no mostró el filtrado:\n%s", out)
	}
	if len(r.Declaraciones) != 1 || r.Declaraciones[0].ID != "opencode/mimo-v2.5-free" {
		t.Errorf("eligió %+v", r.Declaraciones)
	}
}

// PEGAR UN ID TIENE QUE FUNCIONAR SIEMPRE, y es la regla que ordena §8 de la
// spec: el listado es una comodidad, nunca el mecanismo. Un id que no está en la
// lista se toma tal cual — puede ser uno nuevo que el arnés todavía no enumera.
func TestUnIdQueNoEstaEnLaListaSeTomaIgual(t *testing.T) {
	r, _, err := correr(t, "1\nprov/recien-salido\nnuevo\n\n\n", []string{"opencode"},
		catalogoFalso(tresDeOpencode))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Declaraciones) != 1 {
		t.Fatalf("%+v", r.Declaraciones)
	}
	if r.Declaraciones[0].ID != "prov/recien-salido" {
		t.Errorf("no tomó el id pegado: %+v", r.Declaraciones[0])
	}
}

// Y si el arnés no sabe listar —Claude Code, o uno cuyo listado se rompió—, la
// pregunta sigue: pide el id a mano. Que no haya menú no puede frenar a nadie.
func TestSinListadoIgualSePuedeDeclarar(t *testing.T) {
	r, out, err := correr(t, "1\nopus\ngrande\n\n\n", []string{"claude-code"},
		catalogoFalso(nil)) // ninguno lista
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "Pegá el id") {
		t.Errorf("no ofreció el camino manual:\n%s", out)
	}
	if len(r.Declaraciones) != 1 || r.Declaraciones[0].ID != "opus" {
		t.Errorf("no tomó el id tipeado: %+v", r.Declaraciones)
	}
}

// `mecanico` NO se pregunta: ningún estado lo pide (`perfilPorEstado`), lo
// nombran los sfx-* que están fuera de los nueve. Preguntarlo sería pedir una
// decisión que no frena nada.
func TestNoPreguntaPorMecanico(t *testing.T) {
	_, out, err := correr(t, "1\n\n\n", []string{"opencode"}, catalogoFalso(tresDeOpencode))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, global.Mecanico) {
		t.Errorf("preguntó por mecanico:\n%s", out)
	}
	for _, p := range []string{global.Razonar, global.Construir} {
		if !strings.Contains(out, p) {
			t.Errorf("no preguntó por %s:\n%s", p, out)
		}
	}
}

// Un alias vacío no se acepta: sin alias el modelo no se puede nombrar desde
// tareas.json, que es para lo único que existe el alias.
func TestNoAceptaUnAliasVacio(t *testing.T) {
	r, out, err := correr(t, "1\n1\n\n\nultra\n\n\n", []string{"opencode"},
		catalogoFalso(tresDeOpencode))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "alias") {
		t.Errorf("no volvió a pedirlo:\n%s", out)
	}
	if len(r.Declaraciones) != 1 || r.Declaraciones[0].Alias != "ultra" {
		t.Errorf("%+v", r.Declaraciones)
	}
}

// SE QUEDA SIN ENTRADA A LA MITAD: no es un error, es lo que hay.
//
// Pasa con un Ctrl-D, y pasa con un pipe que se acabó. Devolver error tiraría lo
// que Javier YA contestó, que es lo peor que se puede hacer con respuestas que
// costó dar.
func TestSeQuedaSinEntradaYDevuelveLoContestado(t *testing.T) {
	r, _, err := correr(t, "1\n1\nultra\n", []string{"opencode"}, catalogoFalso(tresDeOpencode))
	if err != nil {
		t.Fatal(err)
	}
	if len(r.Harnesses) != 1 || r.Harnesses[0] != "opencode" {
		t.Errorf("perdió los arneses ya elegidos: %v", r.Harnesses)
	}
	if len(r.Declaraciones) != 1 {
		t.Errorf("perdió la declaración ya completa: %+v", r.Declaraciones)
	}
}

// Sin nada instalado se ofrecen los tres igual: puede que estén y la detección
// no los haya visto, o que Javier quiera dejar preparado uno que va a instalar
// después. La detección es una AYUDA, no una compuerta.
func TestSinDeteccionOfreceLosTresIgual(t *testing.T) {
	_, out, err := correr(t, "\n\n\n\n\n\n\n", nil, catalogoFalso(nil))
	if err != nil {
		t.Fatal(err)
	}
	for _, h := range global.Harness {
		if !strings.Contains(out, h) {
			t.Errorf("no ofreció %s:\n%s", h, out)
		}
	}
}
