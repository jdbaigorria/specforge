package andamio

import (
	"context"
	"errors"
	"os/exec"
	"slices"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// sonda arma un par buscar/preguntar de mentira.
//
// enPath son los binarios que "están"; dice es qué contesta cada uno.
func sonda(t *testing.T, enPath []string, dice map[string]string) {
	t.Helper()
	viejoB, viejoP := buscar, preguntar
	t.Cleanup(func() { buscar, preguntar = viejoB, viejoP })

	buscar = func(n string) (string, error) {
		if slices.Contains(enPath, n) {
			return "/usr/bin/" + n, nil
		}
		return "", exec.ErrNotFound
	}
	preguntar = func(_ context.Context, n string, _ ...string) ([]byte, error) {
		s, ok := dice[n]
		if !ok {
			return nil, errors.New("no se pudo ejecutar")
		}
		return []byte(s), nil
	}
}

func TestInstaladosDevuelveLosQueEstanYSeIdentifican(t *testing.T) {
	sonda(t,
		[]string{"claude", "opencode", "cmd"},
		map[string]string{
			"claude":   "2.1.252 (Claude Code)",
			"opencode": "1.18.25",
			"cmd":      "\nCommand Code v1.39.2\nCoding agent that…",
		})

	got := Instalados()
	if len(got) != 3 {
		t.Fatalf("esperaba los tres, hubo %v", got)
	}
	// El orden es el de global.Harness y no el del mapa, que en Go es aleatorio.
	// Un menú que cambia de orden en cada corrida es un menú donde te equivocás.
	if !slices.Equal(got, global.Harness) {
		t.Errorf("el orden no es el de global.Harness: %v", got)
	}
}

func TestInstaladosSalteaElQueNoEstaEnElPath(t *testing.T) {
	sonda(t,
		[]string{"claude"},
		map[string]string{"claude": "2.1.252 (Claude Code)"})

	got := Instalados()
	if !slices.Equal(got, []string{"claude-code"}) {
		t.Errorf("esperaba sólo claude-code, hubo %v", got)
	}
}

// EL CASO QUE JUSTIFICA QUE HAYA UNA CONFIRMACIÓN.
//
// El binario de Command Code se llama `cmd`, que es un nombre carísimo de
// genérico: en Windows es la shell, y en cualquier máquina puede haber otro. Que
// exista un archivo llamado `cmd` NO prueba que sea Command Code, y ofrecerlo en
// el menú haría que `sf install` arme `.commandcode/` para algo que no existe.
func TestInstaladosDescartaUnBinarioQueNoSeIdentifica(t *testing.T) {
	sonda(t,
		[]string{"claude", "cmd"},
		map[string]string{
			"claude": "2.1.252 (Claude Code)",
			"cmd":    "Microsoft Windows [Version 10.0]",
		})

	if got := Instalados(); slices.Contains(got, "commandcode") {
		t.Errorf("un `cmd` que no es Command Code entró igual: %v", got)
	}
}

// Y la asimetría, que es real y está medida: `opencode --version` imprime
// "1.18.25" y nada más — no dice su nombre. Así que para opencode la
// confirmación es que el comando ANDE, y no lo que conteste. Es más flojo, y es
// lo único honesto: inventarle una marca que no emite sería adivinar.
func TestInstaladosAceptaOpencodeAunqueNoDigaSuNombre(t *testing.T) {
	sonda(t,
		[]string{"opencode"},
		map[string]string{"opencode": "1.18.25"})

	if got := Instalados(); !slices.Contains(got, "opencode") {
		t.Errorf("opencode quedó afuera por no decir su nombre: %v", got)
	}
}

func TestInstaladosDescartaElQueEstaPeroNoCorre(t *testing.T) {
	sonda(t,
		[]string{"claude", "opencode"},
		map[string]string{"claude": "2.1.252 (Claude Code)"}) // opencode no contesta

	if got := Instalados(); slices.Contains(got, "opencode") {
		t.Errorf("un binario que no ejecuta entró igual: %v", got)
	}
}

func TestInstaladosSinNingunoDevuelveVacio(t *testing.T) {
	sonda(t, nil, nil)

	if got := Instalados(); len(got) != 0 {
		t.Errorf("esperaba vacío, hubo %v", got)
	}
}
