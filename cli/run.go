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

	// retriedForContext recuerda qué waves ya consumieron su único reintento por
	// NEEDS_CONTEXT. Vive fuera del loop porque el loop re-deriva la wave del
	// disco en cada vuelta: sin esto, un agente que siempre pide contexto
	// loopearía hasta agotar el bound en vez de escalar.
	retriedForContext := map[int]bool{}

	// El loop. Releemos el estado en cada vuelta: cada checkpoint sellado mueve
	// la frontera, y derivePhase la recomputa del disco (no de esta memoria).
	//
	// El bound contempla los reintentos: cada wave puede consumir hasta DOS
	// vueltas (intento + reintento por NEEDS_CONTEXT), más una final para
	// detectar que ya no quedan waves.
	for iter := 0; iter <= 2*total+1; iter++ {
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
		exit := runAgent(projectDir, agentCmd, seed)
		switch classifyAgentExit(exit) {
		case outcomeDone:
			// Camino normal: al checkpoint.

		case outcomeConcerns:
			// La wave SE HIZO; el agente avisa que algo no cierra. El checkpoint
			// corre igual — si pasa, se sella y se sigue. Frenar acá castigaría
			// al agente por ser honesto, y la preocupación ya quedó escrita.
			fmt.Printf("  agent reported CONCERNS — running the checkpoint anyway; "+
				"the detail is in progress/wave-%d.md\n", wave)

		case outcomeNeedsContext:
			// No es un fallo: el agente dice que le falta material. Se relanza la
			// wave UNA vez. Dos veces seguidas significa que falta algo que el
			// seed no puede dar, y ahí decide un humano.
			if retriedForContext[wave] {
				fmt.Fprintf(os.Stderr, "\nsf run: agent still reports NEEDS_CONTEXT on wave %d after a retry — STOP.\n"+
					"Two in a row means the seed can't supply it. See progress/wave-%d.md for what it asked for.\n", wave, wave)
				return 6
			}
			retriedForContext[wave] = true
			fmt.Printf("  agent reported NEEDS_CONTEXT — relaunching wave %d once\n", wave)
			continue

		case outcomeBlocked:
			fmt.Fprintf(os.Stderr, "\nsf run: agent reports BLOCKED on wave %d — STOP.\n"+
				"It needs a decision that isn't its to make. See progress/wave-%d.md.\n", wave, wave)
			return 6

		case outcomeCrash:
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

Report how the wave ended with your EXIT CODE. The orchestrator routes on it,
so prose about being blocked changes nothing — the code is what it reads:
-   0  DONE               — tasks complete, tests pass.
-  10  DONE_WITH_CONCERNS — you finished, but something doesn't add up. The
                            checkpoint still runs; say what worries you in
                            progress/wave-%d.md.
-  11  NEEDS_CONTEXT      — you can't proceed without material the seed didn't
                            give you. Write exactly what you need in
                            progress/wave-%d.md; the wave is relaunched once.
-  12  BLOCKED            — this needs a decision that isn't yours to make (it
                            invalidates the design, or it's a product call).
                            Explain it in progress/wave-%d.md and stop.
Any other non-zero code is read as a crash, not as a status.

── WAVE SEED (deterministic, from sf context for-wave) ──
%s
`, wave, feature, feature, wave, wave, wave, slice)
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

// ----------------------------------------------------------------------------
// DL-12 — protocolo de estado del subagente.
//
// Antes el orquestador leía UN BIT: cero o no cero. "Terminé pero algo huele
// mal" y "me falta contexto" no eran expresables, así que colapsaban a "no
// cero" y `sf run` frenaba igual que ante un crash.
//
// El enum viaja por EXIT CODE, no por un archivo de estado. La razón es de
// seguridad: `decidePreToolUse` protege por NOMBRE DE ARCHIVO BASE, así que un
// `progress/wave-N.status.json` no caería en ninguna categoría protegida y
// sería escribible a mano con Write. Sería estado autoritativo sin proteger
// ruteando una decisión del CLI — exactamente el agujero de FIXBUGHIGH. El
// exit code lo produce el proceso al terminar y `sf run` ya lo leía.
//
// El DETALLE no viaja acá: el exit code lleva la decisión, y el porqué lo
// escribe el subagente en progress/wave-N.md, que ya escribe. Canal angosto
// para la decisión, artefacto para la evidencia.
//
// Esto es tier COOPERATIVO: un agente puede salir 0 sin haber hecho nada, igual
// que antes. Lo que cambia no es la garantía sino la resolución. La garantía la
// sigue dando el checkpoint (contrato + suite verde), que corre después y NO
// le cree al exit code.
// ----------------------------------------------------------------------------

type waveOutcome int

const (
	outcomeDone         waveOutcome = iota // 0
	outcomeConcerns                        // 10 — hecho, pero algo no cierra
	outcomeNeedsContext                    // 11 — no es fallo: falta material
	outcomeBlocked                         // 12 — hace falta una decisión ajena
	outcomeCrash                           // cualquier otro no-cero
)

// Rango reservado ALTO a propósito: un `claude -p` que falla por su cuenta sale
// 1 o 2, y eso tiene que seguir siendo un crash. Sin la reserva, un error del
// harness sería indistinguible de un estado del protocolo.
const (
	exitWaveConcerns     = 10
	exitWaveNeedsContext = 11
	exitWaveBlocked      = 12
)

func classifyAgentExit(exit int) waveOutcome {
	switch exit {
	case 0:
		return outcomeDone
	case exitWaveConcerns:
		return outcomeConcerns
	case exitWaveNeedsContext:
		return outcomeNeedsContext
	case exitWaveBlocked:
		return outcomeBlocked
	default:
		return outcomeCrash
	}
}
