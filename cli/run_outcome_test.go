package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// DL-12 — enum de estado del subagente.
//
// Antes `sf run` leía UN BIT del agente: cero o no cero. DONE_WITH_CONCERNS y
// NEEDS_CONTEXT no eran expresables — colapsaban a "no cero" y el orquestador
// frenaba igual que ante un crash.
//
// El enum viaja por EXIT CODE en un rango reservado, no por un archivo de
// estado: el hook protege por NOMBRE DE ARCHIVO BASE, así que un
// `progress/wave-N.status.json` sería escribible a mano — estado autoritativo
// no protegido ruteando una decisión del CLI, que es el agujero de FIXBUGHIGH.
// ---------------------------------------------------------------------------

func TestClassifyAgentExit(t *testing.T) {
	cases := []struct {
		exit int
		want waveOutcome
	}{
		{0, outcomeDone},
		{10, outcomeConcerns},
		{11, outcomeNeedsContext},
		{12, outcomeBlocked},
		// El rango reservado es alto a propósito: un `claude -p` que falla solo
		// sale 1 o 2, y eso tiene que seguir siendo un crash, no un estado.
		{1, outcomeCrash},
		{2, outcomeCrash},
		{127, outcomeCrash},
	}
	for _, c := range cases {
		if got := classifyAgentExit(c.exit); got != c.want {
			t.Errorf("classifyAgentExit(%d)=%v, want %v", c.exit, got, c.want)
		}
	}
}

// runProjectWithAgent es runProject con un agent_cmd a medida, para poder
// simular cada estado del protocolo.
func runProjectWithAgent(t *testing.T, agentCmd string) string {
	t.Helper()
	proj := runProject(t)
	mustWrite(t, filepath.Join(proj, "specforge", "constitution.json"),
		`{"schema_version":"1.0","identity_md":"x","build":{"mode":"per-wave",`+
			`"test_cmd":"true","agent_cmd":`+quoteJSON(agentCmd)+`}}`)
	return proj
}

func quoteJSON(s string) string {
	out := []byte{'"'}
	for i := 0; i < len(s); i++ {
		if s[i] == '"' || s[i] == '\\' {
			out = append(out, '\\')
		}
		out = append(out, s[i])
	}
	return string(append(out, '"'))
}

func TestRunRoutesWaveOutcomes(t *testing.T) {
	// DONE_WITH_CONCERNS: la wave se hizo. El checkpoint corre igual y, si
	// pasa, se sella y se sigue — la preocupación queda visible, no bloquea.
	t.Run("concerns sella y sigue", func(t *testing.T) {
		proj := runProjectWithAgent(t, "sh -c 'cat > /dev/null; exit 10'")
		if code := runRun([]string{"--feature=f", proj}); code != 0 {
			t.Fatalf("exit=%d, want 0 (concerns no bloquea si el checkpoint pasa)", code)
		}
		ff, _ := readFeaturesFile(proj)
		if latestApproveGate(findFeature(&ff, "f"), "wave-0") == nil {
			t.Error("wave-0 debería quedar sellada")
		}
	})

	// BLOCKED: para y escala. Nada se sella.
	t.Run("blocked para sin sellar", func(t *testing.T) {
		proj := runProjectWithAgent(t, "sh -c 'cat > /dev/null; exit 12'")
		if code := runRun([]string{"--feature=f", proj}); code == 0 {
			t.Fatal("blocked debería frenar con exit != 0")
		}
		ff, _ := readFeaturesFile(proj)
		if latestApproveGate(findFeature(&ff, "f"), "wave-0") != nil {
			t.Error("blocked NO debe sellar la wave")
		}
	})

	// NEEDS_CONTEXT dos veces seguidas: escala en vez de loopear.
	t.Run("needs-context escala tras un reintento", func(t *testing.T) {
		proj := runProjectWithAgent(t, "sh -c 'cat > /dev/null; exit 11'")
		if code := runRun([]string{"--feature=f", proj}); code == 0 {
			t.Fatal("needs-context persistente debería escalar con exit != 0")
		}
	})

	// Regresión: un crash común sigue siendo un crash.
	t.Run("exit 1 sigue siendo crash", func(t *testing.T) {
		proj := runProjectWithAgent(t, "sh -c 'cat > /dev/null; exit 1'")
		if code := runRun([]string{"--feature=f", proj}); code != 1 {
			t.Errorf("exit=%d, want 1 (comportamiento actual intacto)", code)
		}
	})
}

// TestSeedTeachesExitProtocol: el seed tiene que ENSEÑARLE el protocolo al
// agente. Sin esto el enum no existe en la práctica: el agente nunca sabría
// que 11 significa algo.
func TestSeedTeachesExitProtocol(t *testing.T) {
	proj := runProject(t)
	seed, ok := buildRunSeed(proj, "f", 0)
	if !ok {
		t.Fatal("buildRunSeed falló")
	}
	for _, want := range []string{"DONE_WITH_CONCERNS", "NEEDS_CONTEXT", "BLOCKED", "progress/wave-0.md"} {
		if !strings.Contains(seed, want) {
			t.Errorf("el seed no menciona %q", want)
		}
	}
	if strings.Contains(seed, "MISSING") || strings.Contains(seed, "%!") {
		t.Errorf("el seed tiene un placeholder sin argumento:\n%s", seed)
	}
}
