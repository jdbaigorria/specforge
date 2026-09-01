package lanzar

import (
	"strings"
	"testing"
)

func TestContarDiceComoSalioYQueSigue(t *testing.T) {
	c := 0.0951614
	si := true
	f := Ficha{
		ID: "2026-08-31T14-30-22-prd", Estado: "prd", Harness: "opencode",
		Modelo: "opencode/nemotron-3-ultra-free", Skill: "sfp-po",
		DuroMs: 26944, Salida: 0, Fin: FinTermino, Registro: "…jsonl",
		CostoUSD: &c, Tokens: &Tokens{Entrada: 4, Salida: 154}, CargoLaSkill: &si,
	}

	s := Contar(f)
	if !strings.Contains(s, "27s") && !strings.Contains(s, "26.9s") {
		t.Errorf("no dice cuánto tardó:\n%s", s)
	}
	if !strings.Contains(s, "0.095") {
		t.Errorf("no dice cuánto costó:\n%s", s)
	}
	// LO QUE NO PUEDE FALTAR: que la compuerta sigue siendo la compuerta.
	if !strings.Contains(s, "sf done") {
		t.Errorf("no dice que falta `sf done`:\n%s", s)
	}
}

// SI LA SKILL NO SE CARGÓ, SE AVISA FUERTE.
//
// `sf` apunta a la skill y apuntar es una instrucción que se puede ignorar —
// opencode ya ignoró un `via: subagente` el 2026-08-31 e hizo el trabajo él
// mismo—. Es la única señal barata de que el paso corrió como se pidió.
func TestAvisaCuandoLaSkillNoSeCargo(t *testing.T) {
	no := false
	f := Ficha{Estado: "prd", Skill: "sfp-po", Fin: FinTermino, CargoLaSkill: &no}

	if s := Contar(f); !strings.Contains(s, "skill") {
		t.Errorf("no avisó que no cargó la skill:\n%s", s)
	}
}

// Y "no sé" no se cuenta como "no": si el arnés no lo dice, no se dice nada.
func TestNoInventaCuandoNoSeSabeSiCargoLaSkill(t *testing.T) {
	f := Ficha{Estado: "prd", Skill: "sfp-po", Fin: FinTermino}

	if s := Contar(f); strings.Contains(s, "no cargó") {
		t.Errorf("afirmó algo que no sabe:\n%s", s)
	}
}

func TestContarUnTimeoutLoDiceYNoLoDisimula(t *testing.T) {
	f := Ficha{Estado: "prd", Fin: FinTimeout, DuroMs: 1800000}

	s := Contar(f)
	if !strings.Contains(s, "espera") {
		t.Errorf("no dice que se agotó la espera:\n%s", s)
	}
	if strings.Contains(s, "sf done") {
		t.Errorf("ofreció `sf done` sobre una corrida que no terminó:\n%s", s)
	}
}

// Un costo ausente no se imprime como 0: el arnés no lo dijo.
func TestUnCostoAusenteNoSeImprimeComoCero(t *testing.T) {
	f := Ficha{Estado: "prd", Fin: FinTermino, Tokens: &Tokens{Entrada: 1, Salida: 2}}

	if s := Contar(f); strings.Contains(s, "US$ 0") || strings.Contains(s, "$0.00") {
		t.Errorf("inventó un costo cero:\n%s", s)
	}
}
