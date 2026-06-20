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
	case "tasks":
		return runTasks(rest)
	case "context":
		return runContext(rest)
	case "state":
		return runState(rest)
	case "requirements":
		return runRequirements(rest)
	case "design":
		return runDesign(rest)
	case "constitution":
		return runConstitution(rest)
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
  status  project-level health view (reads specforge/features.json)
  doctor  detect drift between archived specs and code (--drift)
  gate    gate ledger + quality verdicts (gate status [--feature=NAME] | gate record-verdict --feature=NAME)
  trace   verify the traceability matrix vs code (trace verify [--feature=NAME])
  tasks   render/validate tasks.json (tasks render|validate --feature=NAME)
  context emit a JSON slice (context for-wave --n=N | context current [--breadcrumb] | context for-judge --phase=PHASE)
  state   emit the current feature/phase/wave as JSON (state current) — the brain hooks consult
  requirements  render/validate requirements.json (requirements render|validate --feature=NAME)
  design  render/validate design.json (design render|validate --feature=NAME)
  constitution  render/validate constitution.json (constitution render|validate)
  plan    render/validate plan.json (plan render|validate --feature=NAME)
  review  render/validate review.json (review render|validate --feature=NAME)
  save    validate JSON from stdin, then write it + render md (save <artifact> --feature=NAME --json -)
  journal persist durable lessons to specforge/journal/ + git-stage (journal add --feature=NAME --json - [--bridge-icm])
  help    show this message
`)
}
