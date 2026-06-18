# sf — SpecForge helper CLI

A small, dependency-free Go CLI that **reads what the SpecForge skills already
produce** (skill files, `specforge/features.json`, `trace.json`) and
validates / computes / reports. It does **not** replace the markdown workflow —
the LLM produces and judges; `sf` persists, validates, computes and enforces.

This is the deterministic slice of SpecForge: the work that is mechanical and
reproducible (same input → same output). It subsumes the maintainer Python
scripts that used to live in `scripts/`.

## Commands

| Command | What it does |
|---|---|
| `sf lint` | Check the skill suite for consistency (frontmatter, naming, no `sdd-` leftovers, unknown skill refs, broken relative links, EN/ES section parity). Exit 1 on any error. |
| `sf status [dir]` | Project-level health view from `specforge/features.json`: per-feature lane / status / phase / drift / gaps / blockers, plus a dependency-ordered critical path. |
| `sf doctor [--drift] [dir] [--run-tests "CMD {test}"]` | Detect drift between archived specs and code: for each `trace.json`, check every code anchor (`path:symbol`) still exists. `--run-tests` also runs the referenced tests. Exit 1 on drift. |
| `sf gate status [dir] [--feature=NAME]` | Show the gate ledger from `features.json`: a per-feature summary, or the full gate history of one feature. |
| `sf trace verify [dir] [--feature=NAME]` | Verify the traceability matrix against the code (recorded vs live status per requirement). Exit 1 on drift. |

`dir` defaults to the current directory. Commands that read a project expect a
`specforge/` directory there (e.g. `examples/slugify`).

## Build & run

```sh
cd cli
go build -o sf .      # produces ./sf
./sf help

# run from the repo root against a project
./cli/sf lint
./cli/sf status examples/slugify
./cli/sf doctor examples/slugify
```

## Develop

```sh
cd cli
go test ./...         # unit tests (table-driven, use t.TempDir)
go vet ./...          # static checks
gofmt -l *.go         # formatting (empty output = clean)
```

## Layout

Flat `package main`, standard library only (no third-party deps):

- `main.go` — subcommand dispatch and exit codes
- `lint.go` — `sf lint`
- `status.go` — `sf status` (+ shared `featuresFile`/`feature`/`gate` model)
- `doctor.go` — `sf doctor` (+ `trace.json` model and the drift engine)
- `gate.go` — `sf gate status`
- `trace.go` — `sf trace verify`
- `table.go` — shared aligned-table renderer
- `*_test.go` — unit tests

## Exit codes

`0` ok · `1` error / drift found · `2` usage error · `4` not found.
