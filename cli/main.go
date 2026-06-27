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
	case "plan":
		return runPlan(rest)
	case "review":
		return runReview(rest)
	case "save":
		return runSave(rest)
	case "journal":
		return runJournal(rest)

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
	fmt.Fprint(os.Stderr, `sf — SpecForge helper CLI

usage: sf <command> [flags]

commands:
  lint    check the skill suite for consistency
  status  project health view, or artifact staleness with --artifacts [--feature=NAME]
  gate    gate ledger + approvals + verdicts (gate approve --feature=NAME --phase=PHASE seals the artifact hash)
  doctor  detect drift between archived specs and code (--drift), or check install health (--install [--global])
  trace   verify the traceability matrix vs code (trace verify [--feature=NAME] [--contract [--wave=N]])
  graph   export the knowledge graph from declared structure (graph export [--feature=NAME] [--format=json|mermaid|both] [--stdout])
  tasks   render/validate tasks.json (tasks render|validate --feature=NAME)
  context emit a JSON slice (context for-wave --n=N | context current [--breadcrumb] | context for-judge --phase=PHASE)
  metrics measure context-slice output size for a token baseline (metrics context [--feature=NAME])
  state   emit the current feature/phase/wave as JSON (state current) — the brain hooks consult
  next    the agent compass: next valid action + write/read contract (next [--json] [--explain])
  recover diagnose stale/inconsistent artifacts + emit an ordered recovery plan (recover [--feature=NAME] [--json])
  hook    enforcement engine: reads a hook payload on stdin, emits a decision (hook --event=E --harness=generic|claude-code)
  install detect harnesses + install skills/AGENT.md/adapters with backup (install [--from=PATH] [--global] [--dry-run])
  uninstall revert what sf install did, from the manifest (uninstall [--global])
  requirements  render/validate requirements.json (requirements render|validate --feature=NAME)
  design  render/validate design.json (design render|validate --feature=NAME)
  constitution  render/validate constitution.json (constitution render|validate)
  domain  render/validate domain.json — project domain knowledge (domain render|validate)
  plan    render/validate/compute the wave layout (plan render|validate|compute --feature=NAME)
  review  render/validate review.json (review render|validate --feature=NAME)
  save    validate JSON from stdin, then write it + render md (save <artifact> --feature=NAME --json -)
  journal persist durable lessons to specforge/journal/ + git-stage (journal add --feature=NAME --json - [--bridge-icm])
  help    show this message
`)
}
