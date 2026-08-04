package main

import (
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// Tests de `sf next`. Como buildNextContract es PURA (recibe el estado ya
// computado y no toca disco), la probamos sin tocar el filesystem: armamos un
// currentState a mano y verificamos el contrato resultante. Esto es justamente
// por qué separamos "computar" de "imprimir": el cálculo se testea aislado.
// ----------------------------------------------------------------------------

// TestNextContractPerPhase: cada fase del pipeline produce el contrato correcto
// — la instrucción adecuada, el gate siguiente, y (lo más importante) los
// permisos de escritura: la fase actual SÍ se puede escribir, las de aguas abajo
// NO. Es el corazón del runtime contract.
func TestNextContractPerPhase(t *testing.T) {
	ff := featuresFile{} // vacío: no hay dependencias que resolver acá

	cases := []struct {
		phase         string
		wantGate      string
		commandHas    string // un fragmento que DEBE aparecer en command (camino CLI)
		blockedHasSrc bool   // ¿src/** debe estar bloqueado en esta fase?
	}{
		{"requirements", "requirements", "sf save requirements", true},
		{"design", "design", "sf save design", true},
		{"tasks", "tasks", "sf save tasks", true},
		{"plan", "plan", "sf save plan", true},
		{"verdict", "verdict", "sf gate approve", true},
	}

	for _, c := range cases {
		t.Run(c.phase, func(t *testing.T) {
			st := currentState{Feature: "auth", Status: "approved", Phase: c.phase, LastApprovedGate: "x"}
			nc := buildNextContract(st, ff, "/nonexistent")

			if nc.NextGate != c.wantGate {
				t.Errorf("next_gate=%q, want %q", nc.NextGate, c.wantGate)
			}
			// JSON-first: la fase produce su estado vía un comando sf, NO escribiendo
			// el .json directo. El contrato debe nombrar ese comando…
			if !sliceContainsSub(nc.Command, c.commandHas) {
				t.Errorf("command=%v, want one containing %q", nc.Command, c.commandHas)
			}
			// …y NO debe ofrecer el .json crudo como destino de Write (lo deniega el hook).
			if len(nc.AllowedWrites) != 0 {
				t.Errorf("phase %q no debería permitir Write directo, got allowed=%v", c.phase, nc.AllowedWrites)
			}
			if c.blockedHasSrc && !sliceContainsSub(nc.BlockedWrites, "src/**") {
				t.Errorf("phase %q should block src/**, got blocked=%v", c.phase, nc.BlockedWrites)
			}
			if nc.Instruction == "" {
				t.Errorf("phase %q produced an empty instruction", c.phase)
			}
		})
	}
}

// TestNextBuildAllowsCode: build es la ÚNICA fase donde escribir código es
// correcto, y el gate es por wave.
func TestNextBuildAllowsCode(t *testing.T) {
	wave := 1
	st := currentState{Feature: "auth", Status: "building", Phase: "build", Wave: &wave, LastApprovedGate: "wave-0"}
	nc := buildNextContract(st, featuresFile{}, "/nonexistent")

	if nc.NextGate != "wave-1" {
		t.Errorf("next_gate=%q, want wave-1", nc.NextGate)
	}
	if !sliceContainsSub(nc.AllowedWrites, "src/**") {
		t.Errorf("build should allow src/**, got %v", nc.AllowedWrites)
	}
	if sliceContainsSub(nc.BlockedWrites, "src/**") {
		t.Errorf("build must NOT block src/**, got %v", nc.BlockedWrites)
	}
}

// TestNextNoActiveFeature: sin feature activa el contrato no explota — da una
// acción útil (arrancar) y propaga la nota del estado.
func TestNextNoActiveFeature(t *testing.T) {
	st := currentState{Phase: "—", Note: "no active feature"}
	nc := buildNextContract(st, featuresFile{}, "/nonexistent")

	if nc.Feature != "" {
		t.Errorf("feature should be empty, got %q", nc.Feature)
	}
	if !strings.Contains(nc.Instruction, "sf-propose") {
		t.Errorf("instruction should point to sf-propose, got %q", nc.Instruction)
	}
}

// TestNextBlockedByDependency: si la feature espera a otra, el paso es desbloquear
// la dependencia — no avanzar la fase.
func TestNextBlockedByDependency(t *testing.T) {
	st := currentState{Feature: "ui", Phase: "requirements", BlockedBy: []string{"auth"}}
	nc := buildNextContract(st, featuresFile{}, "/nonexistent")

	if !strings.Contains(nc.Instruction, "blocked") {
		t.Errorf("instruction should mention the block, got %q", nc.Instruction)
	}
	if !sliceContainsSub(nc.Why, "auth") {
		t.Errorf("why should name the blocker, got %v", nc.Why)
	}
}

// sliceContainsSub: ¿algún elemento del slice contiene el substring? Helper de
// test para no acoplarnos a rutas exactas.
func sliceContainsSub(xs []string, sub string) bool {
	for _, x := range xs {
		if strings.Contains(x, sub) {
			return true
		}
	}
	return false
}
