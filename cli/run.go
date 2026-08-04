package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf run` — orquestación paso a paso (F1/R7, v0). La INVERSIÓN DE CONTROL.
//
// Hasta acá el LLM conduce y `sf` restringe (deny/nudge/inject): el proceso
// depende de que una conversación larga recuerde el proceso — exactamente
// donde se rompía. `sf run` invierte eso para la parte NO gateada (la
// ejecución de waves): el CLI conduce y el agente ejecuta.
//
// El loop determinista, por cada wave pendiente:
//
//   1. computar el seed exacto (loadWaveContext — el mismo slice de
//      `sf context for-wave`, ni un token más)
//   2. invocar el agente configurado (build.agent_cmd), seed por STDIN.
//      El agente es UN PROCESO FRESCO por wave: la degradación de contexto
//      no se acumula entre waves (per-wave, generalizado).
//   3. validar con la máquina, no con la narración:
//      - contrato de verificación de la wave (trace verify --contract --wave)
//      - suite verde (sf check run — exit code + hash sellados)
//   4. sellar el checkpoint (gate wave-N, by "sf run") y pasar a la siguiente.
//
// Cualquier paso falla → STOP y escalación al humano. Las fases GATEADAS
// (propose, verdict) siguen siendo conversacionales: sf run gobierna solo la
// ejecución larga. El gate de plan aprobado es la autorización de spawn
// (misma regla que sf-build).
// ----------------------------------------------------------------------------

func runRun(args []string) int {
	projectDir := "."
	featureName := ""
	dryRun := false
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			featureName = strings.TrimPrefix(a, "--feature=")
		case a == "--dry-run":
			dryRun = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf run: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if featureName == "" {
		fmt.Fprintln(os.Stderr, "sf run: --feature=NAME is required")
		return 2
	}

	// El comando del agente sale de la constitución — igual que test_cmd, es
	// configuración del PROYECTO, no un flag que el modelo pueda inventar.
	cf, code := readConstitutionFile(projectDir)
	if code != 0 {
		return code
	}
	agentCmd := ""
	if cf.Build != nil {
		agentCmd = strings.TrimSpace(cf.Build.AgentCmd)
	}
	if agentCmd == "" {
		fmt.Fprintln(os.Stderr, "sf run: no build.agent_cmd in the constitution — declare how to invoke "+
			`your agent (e.g. "claude -p --permission-mode acceptEdits"). The wave seed is piped to its stdin.`)
		return 2
	}

	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf run: cannot read feature state under %s (%v)\n", projectDir, err)
		return 4
	}
	f := findFeature(&ff, featureName)
	if f == nil {
		fmt.Fprintf(os.Stderr, "sf run: feature %q not found\n", featureName)
		return 4
	}
	// Mismas precondiciones que un build humano: ledger íntegro y plan sellado
	// (el gate de plan es la autorización de spawn — regla de sf-build).
	if refuseOnBrokenLedger("sf run", f) {
		return 5
	}
	if latestApproveGate(f, "plan") == nil {
		fmt.Fprintf(os.Stderr, "sf run: the plan gate for %q is not sealed — approve tasks and run "+
			"`sf plan compute` first (spawn authorization).\n", featureName)
		return 5
	}

	total := totalWaves(projectDir, featureName)
	if total <= 0 {
		fmt.Fprintf(os.Stderr, "sf run: no computed waves for %q (run `sf plan compute`).\n", featureName)
		return 4
	}

	// El loop. Releemos el estado en cada vuelta: cada checkpoint sellado mueve
	// la frontera, y derivePhase la recomputa del disco (no de esta memoria).
	for iter := 0; iter <= total; iter++ {
		ff, err := readFeaturesFile(projectDir)
		if err != nil {
			fmt.Fprintf(os.Stderr, "sf run: %v\n", err)
			return 1
		}
		f := findFeature(&ff, featureName)
		phase, wave := derivePhase(f, projectDir)
		if phase != "build" || wave < 0 {
			fmt.Printf("\nsf run: build complete — all %d wave(s) checkpointed. Next phase: %s "+
				"(gated — back to the conversation).\n", total, phase)
			return 0
		}

		seed, ok := buildRunSeed(projectDir, featureName, wave)
		if !ok {
			fmt.Fprintf(os.Stderr, "sf run: cannot build the wave %d seed (missing plan/tasks?)\n", wave)
			return 4
		}

		if dryRun {
			fmt.Printf("would run wave %d/%d: %s  (seed: %d bytes via stdin)\n", wave+1, total, agentCmd, len(seed))
			// En dry-run no ejecutamos ni sellamos: mostramos SOLO la próxima
			// wave (las siguientes dependen del resultado de esta).
			return 0
		}

		fmt.Printf("── wave %d/%d ── agent: %s\n", wave+1, total, agentCmd)
		if exit := runAgent(projectDir, agentCmd, seed); exit != 0 {
			fmt.Fprintf(os.Stderr, "sf run: agent exited %d on wave %d — STOP. Fix/resume manually or re-run.\n", exit, wave)
			return 1
		}

		// Checkpoint automatizado (mismo contrato que el inter-wave de sf-build):
		// 1) contrato de verificación de la wave — cada R tocado nombra un test real.
		if c := runTraceContract(projectDir, featureName, wave); c != 0 {
			fmt.Fprintf(os.Stderr, "\nsf run: wave %d verification contract UNMET — STOP. "+
				"The agent must anchor its work in trace.json before the wave can close.\n", wave)
			return 5
		}
		// 2) suite verde, sellada por la máquina.
		if c := runCheckRun([]string{"--feature=" + featureName, projectDir}); c != 0 {
			fmt.Fprintf(os.Stderr, "\nsf run: tests failed after wave %d — STOP. Fix and re-run `sf run`.\n", wave)
			return 3
		}
		// 3) sellar el checkpoint. El gate wave-N registra QUE el checkpoint
		//    automatizado pasó (by lo dice honestamente: no fue un humano).
		if c := gateApprove(projectDir, featureName, fmt.Sprintf("wave-%d", wave), "sf run",
			"auto checkpoint: verification contract met + suite green"); c != 0 {
			return c
		}
	}
	fmt.Fprintln(os.Stderr, "sf run: wave loop did not converge (more iterations than waves) — check the plan.")
	return 1
}

// buildRunSeed arma el prompt del agente de wave: el framing (contrato de
// trabajo) + el slice determinista de la wave. El framing es corto a propósito:
// las reglas duras las sostienen el hook y el checkpoint, no este texto.
func buildRunSeed(projectDir, feature string, wave int) (string, bool) {
	ctx, ok := loadWaveContext(projectDir, feature, wave)
	if !ok {
		return "", false
	}
	slice, err := json.MarshalIndent(ctx, "", "  ")
	if err != nil {
		return "", false
	}
	var b strings.Builder
	fmt.Fprintf(&b, `You are executing wave %d of feature %q under SpecForge governance.

Work ONLY on the tasks listed in the seed below. Rules:
- Implement each task; anchor your work in trace.json via
  `+"`sf save trace --feature=%s --from=drafts/trace.json`"+` (requirement → code symbol → test).
- Never write SpecForge state directly; author through drafts/ + sf commands.
- When the wave's tasks are done and their tests pass, finish. The orchestrator
  verifies the contract and runs the suite — narration is not evidence.

── WAVE SEED (deterministic, from sf context for-wave) ──
%s
`, wave, feature, feature, slice)
	return b.String(), true
}

// runAgent lanza el agente (shell nativo) con el seed por stdin, streameando su
// salida a la consola — el humano puede mirar el build en vivo.
func runAgent(projectDir, agentCmd, seed string) int {
	cmd := shellCommand(agentCmd)
	cmd.Dir = projectDir
	cmd.Stdin = strings.NewReader(seed)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(interface{ ExitCode() int }); ok {
			return ee.ExitCode()
		}
		return 127
	}
	return 0
}
