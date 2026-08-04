[← Back to README](../README.md)

# Quickstart — first governed feature in ~10 minutes

The full SpecForge experience starts with a constitution conversation
([sf-init](skills.md)). This page is the shortcut: minimal scaffold, one small
feature, every gate real. Solo dev, existing or empty repo.

## 0. Install (once)

```bash
cd your-project
sf install            # hooks + skills for your harness
sf doctor --install   # confirm the hooks are wired
```

## 1. Scaffold (30 seconds)

```bash
sf init --minimal
```

You get `specforge/` with a **3-principle default constitution** (minimal code,
tests prove requirements, explicit scope) and — if your stack is recognizable —
`build.test_cmd` already set (`go test -json ./...`, `pytest -q
--junitxml={report}`, …). With a structured report configured, the final gate
will require every test named in the trace to have **run and passed** — you get
test→requirement causality for free.

Brownfield? Add `sf onboard scan`: the CLI computes the code inventory
(modules, symbols, test map) so the agent interprets a map instead of
free-exploring your repo.

## 2. One small feature

```bash
sf feature add --feature=hello-endpoint
```

Then ask your agent to run **sf-propose hello-endpoint** (or author the JSON
yourself). The agent drafts `requirements` → you approve → `design` → you
approve → `tasks` → you approve. Each approval is sealed:

```bash
sf gate approve --feature=hello-endpoint --phase=requirements
```

That entry records *who*, *when*, and the **hash of what you approved** — a
silent edit afterwards flips the artifact to `stale` (see `sf status
--artifacts`).

## 3. Build

```bash
sf plan compute --feature=hello-endpoint   # waves derived from task deps; gate auto-seals
```

The agent implements wave by wave, writing `trace.json` (requirement → code
symbol → test) through `sf save`. Check where you are anytime:

```bash
sf next        # the compass: exactly one next action
sf status      # the whole board
```

## 4. Close it — the part nobody can fake

```bash
sf check run --feature=hello-endpoint          # CLI runs YOUR test_cmd, seals exit code + code hash
sf gate show --feature=hello-endpoint          # evidence view: trace coverage, freshness, verdicts
sf gate approve --feature=hello-endpoint --phase=verdict
sf feature archive --feature=hello-endpoint
```

`verdict` **refuses** unless: the trace has no drift, every requirement names a
real test, the last run is green **and fresh** (code untouched since), and —
with a report configured — every named test actually passed. This is the gate
the June incident could not have survived.

## 5. Make the guarantee portable

```bash
sf verify           # the aggregate check: ledger integrity + trace + schemas + freshness
sf verify --init-ci # GitHub Action: the same check on every PR — enforcement outside your machine
sf coverage         # % of code anchored to a live spec; the ratchet only lets it go up
```

That's the whole loop. From here, read [mental-model.md](mental-model.md) for
*why* it's shaped like this, and grow the constitution as your project teaches
you its rules (`sf-check` promotes recurring issues to invariants).
