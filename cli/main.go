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
  gate    show the gate ledger (gate status [--feature=NAME])
  trace   verify the traceability matrix vs code (trace verify [--feature=NAME])
  help    show this message
`)
}
