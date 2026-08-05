package main

import (
	"fmt"
	"os"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {

	if len(args) == 0 {
		usage()
		return 2
	}
	cmd, rest := args[0], args[1:]

	switch cmd {
	case "lint":
		return runLint(rest)
	case "status":
		return runStatus(rest)
	case "doctor":
		return runDoctor(rest)
	case "gate":
		return runGate(rest)
	case "check":
		return runCheck(rest)
	case "feature":
		return runFeature(rest)
	case "trace":
		return runTrace(rest)
	case "graph":
		return runGraph(rest)
	case "tasks":
		return runTasks(rest)
	case "context":
		return runContext(rest)
	case "metrics":
		return runMetrics(rest)
	case "state":
		return runState(rest)
	case "next":
		return runNext(rest)
	case "recover":
		return runRecover(rest)
	case "verify":
		return runVerify(rest)
	case "migrate":
		return runMigrate(rest)
	case "hook":
		return runHook(rest)
	case "install":
		return runInstall(rest)
	case "uninstall":
		return runUninstall(rest)
	case "requirements":
		return runRequirements(rest)
	case "design":
		return runDesign(rest)
	case "constitution":
		return runConstitution(rest)
	case "domain":
		return runDomain(rest)
	case "sources":
		return runSources(rest)
	case "plan":
		return runPlan(rest)
	case "review":
		return runReview(rest)
	case "save":
		return runSave(rest)
	case "journal":
		return runJournal(rest)
	case "events":
		return runEvents(rest)
	case "onboard":
		return runOnboard(rest)
	case "init":
		return runInit(rest)
	case "delta":
		return runDelta(rest)
	case "run":
		return runRun(rest)
	case "coverage":
		return runCoverage(rest)
	case "arch":
		return runArch(rest)
	case "mutation":
		return runMutation(rest)

	case "help", "-h", "--help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "sf: unknown command %q\n\n", cmd)
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprint(os.Stderr, usageText)
}

// usageText es una const (no un literal inline) para que el test de paridad de
// lint_contract pueda verificar que cliSurface documenta EXACTAMENTE los
// comandos listados acá — una sola fuente de verdad, verificada.
const usageText = `sf — SpecForge helper CLI

usage: sf <command> [flags]

commands:
  lint    check the skill suite for consistency
  status  project health view [--json], or artifact staleness with --artifacts [--feature=NAME]
  gate    gate ledger + approvals + verdicts + evidence (gate approve|status|record-verdict|show --feature=NAME)
  check   run the project test suite via build.test_cmd and record a deterministic result (check run --feature=NAME)
  feature feature lifecycle — the sole writer of status/lane in features.json (feature add|set-status|set-lane|archive --feature=NAME [--reason=WHY])
  doctor  classify spec-to-code drift in 4 categories (--drift [--json] [--run-tests]), or check install health (--install [--global])
  trace   verify the traceability matrix vs code (trace verify [--feature=NAME] [--contract [--wave=N]])
  graph   export or query the knowledge graph from declared structure (graph export|query <term> [--json])
  tasks   render/validate tasks.json (tasks render|validate --feature=NAME)
  context emit a JSON slice (context for-wave --n=N | context current [--breadcrumb] | context for-judge --phase=PHASE)
  metrics measure context-slice output size for a token baseline (metrics context [--feature=NAME] [--json])
  state   emit the current feature/phase/wave as JSON (state current) — the brain hooks consult
  next    the agent compass: next valid action + write/read contract (next [--json] [--explain])
  recover diagnose stale/inconsistent artifacts + emit an ordered recovery plan (recover [--feature=NAME] [--json])
  verify  aggregate integrity check with exit code, for CI (verify [--feature=NAME] [--json] | verify --init-ci)
  migrate stamp missing schema_version + chain legacy ledgers; refuses unknown versions (migrate [--dry-run])
  hook    enforcement engine: reads a hook payload on stdin, emits a decision (hook --event=E --harness=generic|claude-code)
  install detect harnesses + install skills/AGENT.md/adapters with backup (install [--from=PATH] [--global] [--dry-run])
  uninstall revert what sf install did, from the manifest (uninstall [--global])
  requirements  render/validate requirements.json (requirements render|validate --feature=NAME)
  design  render/validate design.json (design render|validate --feature=NAME)
  constitution  render/validate constitution.json (constitution render|validate)
  domain  render/validate domain.json — project domain knowledge (domain render|validate)
  sources render/validate the ingested material + its coverage (sources render|validate|coverage [--json])
  plan    render/validate/compute the wave layout (plan render|validate|compute --feature=NAME)
  review  render/validate review.json (review render|validate --feature=NAME)
  save    validate JSON, then write it + render md (save constitution|domain|sources|requirements|design|tasks|plan|review|trace [--feature=NAME] --json -|--from=drafts/FILE)
  journal persist durable lessons to specforge/journal/ + git-stage (journal add --feature=NAME --json - [--bridge-icm])
  events  enforcement telemetry — denies, nudges, judge verdicts (events [--json] [--tail=N])
  onboard deterministic brownfield inventory — tree, symbols, test map (onboard scan [--json])
  init    minimal scaffold — 3-principle constitution + detected test_cmd (init --minimal)
  delta   first-class change objects for amends (delta new|list|set-status --feature=NAME [--kind=spec-wrong|code-wrong])
  run     CLI-driven wave execution — fresh agent per wave, machine-verified checkpoints (run --feature=NAME [--dry-run])
  coverage spec coverage with ratchet — % of code anchored to a live trace (coverage [--json] [--by-priority])
  arch    architecture conformance vs the approved design graph (arch rules|check --feature=NAME)
  mutation does the suite detect injected defects? (mutation scope|run --feature=NAME)
  help    show this message
`
