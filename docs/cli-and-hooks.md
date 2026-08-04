[← Back to README](../README.md)

# The deterministic layer (CLI `sf` + hooks)

The skills are the **cooperative** half: an LLM produces specs, judges code,
writes prose. But instruction-following degrades as the context fills — past
~50% an agent starts ignoring the workflow it was told to follow. So SpecForge
ships a **deterministic** half that doesn't depend on the model's goodwill: a
small Go CLI **`sf`** and a set of per-harness **hooks**.

> **The division of labor:** the LLM *produces and judges*; `sf` *persists,
> validates, computes, and renders*; the hooks *force and inject* on harness
> events. None of the three does another's job.

**JSON-first.** Each artifact is a `.json` **source** plus a `.md` **render**.
You (or the LLM) produce the JSON; `sf` validates it and generates the Markdown.
One source of truth — the Markdown can't drift from the data because it's derived
from it.

**Two tiers of enforcement** (this distinction runs through everything below):

| Tier | What it guards | How |
|------|----------------|-----|
| **Structural** (hermetic) | gate order, schema, dependencies, serial flow | a hook *denies* the tool call — it cannot be skipped |
| **Quality** (cooperative) | "is this good / minimal / aligned?" | a fresh sub-agent judges; the verdict is a *nudge*, recorded for the human |

You can make illegal states unreachable (structural). You cannot force good
content into existence (quality) — so quality is raised by a checker, not
guaranteed. SpecForge is honest about which is which.

### The `sf` CLI

`sf` is deterministic: same inputs, same outputs, no model calls. It's safe to
run anywhere and is what the hooks consult.

**Validate & render artifacts (the JSON-first write path).**

```console
$ echo '{"feature":"add-task-crud","tasks":[
    {"id":"T1","title":"TaskModel + JsonStorage","requirement_refs":["R5"],"depends_on":[]},
    {"id":"T2","title":"add command","requirement_refs":["R1"],"depends_on":["T1"]}
  ]}' | sf save tasks --feature=add-task-crud --json -
saved specforge/features/add-task-crud/tasks.json (+ rendered .../tasks.md)
```

`sf save` validates the JSON (unique ids, refs in `R#`/`C#` form, dependency
graph acyclic) and **only if it passes** writes the canonical `.json` + renders
the `.md`. Invalid input is rejected with exit 2 and **nothing touches disk** —
a malformed artifact never exists. (`sf <artifact> validate|render` do the two
halves standalone, for every artifact: constitution, requirements, design,
tasks, plan, review, and **domain** — project-level domain knowledge:
glossary + entities + business rules, validated like the rest, e.g.
`echo '{...}' | sf save domain -`.)

For large artifacts, piping a heredoc is hostile — use **promotable drafts**
instead: `specforge/features/<f>/drafts/` (and project-level
`specforge/drafts/`) is the one corner of `specforge/` the hook leaves
writable. The agent writes the draft there with its native `Write` tool, then
promotes it:

```console
$ sf save tasks --feature=add-task-crud --from=drafts/tasks.json
saved specforge/features/add-task-crud/tasks.json (+ rendered .../tasks.md)
promoted draft .../drafts/tasks.json (removed after save)
```

Same guarantee (only `sf` writes the canonical file — the draft is validated
before promotion and removed after), a fraction of the friction. An invalid
draft stays in `drafts/` to be fixed and re-promoted.

**Compute the wave layout from task dependencies.**

```console
$ sf plan compute --feature=add-task-crud
computed plan for add-task-crud: 5 task(s) → 3 wave(s)
```

Tasks are a **flat** list; each declares `depends_on`. The wave grouping is a
deterministic **topological layering** — wave 0 = tasks with no deps, wave N =
`1 + max(wave of its deps)`. So `sf` *computes* the waves; the LLM doesn't sort
them by hand. `sf plan validate` then enforces the guard `wave(task) >
wave(its deps)` as an error. (Tasks in the same wave are independent →
parallelizable.)

**Read project & gate state.**

```console
$ sf status                       # health table: every feature, phase, drift, blockers
$ sf gate status --feature=X      # the human-approval gate ledger for a feature
$ sf doctor --drift               # archived specs whose code anchor vanished (reads trace.json)
$ sf trace verify --feature=X     # the requirement → code → test matrix vs the repo
$ sf lint                         # the skill suite's own consistency
```

**Export the knowledge graph (deterministic — the opposite of RAG).**

```console
$ sf graph export                 # → specforge/graph.json + specforge/graph.mmd
exported 10 node(s), 9 edge(s) → specforge/graph.json + specforge/graph.mmd
```

`sf graph export` makes **explicit** the graph already declared *implicitly* in
the files — no LLM, no embeddings, fully reproducible. It walks the feature
state (feature nodes + `depends_on` edges) and each feature's `trace.json` (the spine:
feature → `R#` → `code:symbol` → test) and emits two files: `graph.json` for
tools/tests and `graph.mmd`, a Mermaid `flowchart LR` that GitHub and most
Markdown viewers render inline. Flags: `--feature=NAME` (scope to one feature),
`--format=json|mermaid|both`, `--stdout` (print instead of writing).

**Emit the context slices the hooks inject — the "brain".**

```console
$ sf state current
{ "feature": "add-task-crud", "status": "building",
  "phase": "build", "wave": 1, "last_approved_gate": "wave-0" }
```

`sf state current` is a **pure function of the per-feature state**
(`features/<name>/feature.json`): the active feature
(serial flow), the phase derived as *last-approved-gate + 1*, the wave from
`plan.json`. Zero stored state — nothing to keep in sync (no "second truth").

```console
$ sf context current --breadcrumb
SpecForge: feature `add-task-crud`, fase `build` (wave 1) · slice completo: `sf context current`

$ sf context for-wave --feature=add-task-crud --n=1   # the seed for a build sub-agent
$ sf context for-judge --phase=design --feature=X     # artifact + ONLY the principles that apply to that phase
```

`context current` is the minimal slice of the current step — a cheap breadcrumb,
or the full JSON. `for-wave` seeds a build sub-agent with one wave (and injects
the project's `domain.json` if present). `for-judge` hands the quality auditor
exactly the artifact plus the constitution principles **and the domain business
rules** whose `applies_to` includes that phase — nothing more.

**Record quality verdicts & durable lessons.**

```console
$ echo '{"phase":"design","verdicts":[
    {"rule":"P1","result":"pass","citation":"no network calls in any component"},
    {"rule":"I1","result":"fail","citation":"endpoint X lacks error handling"}
  ]}' | sf gate record-verdict --feature=X --phase=design
recorded fail verdict for X/design (2 rule(s))      # exit 3 = recorded FAIL → hook can nudge

$ echo '{"feature":"X","lessons":[
    {"context":"reimplemented date parsing","rule":"use the stdlib"}
  ]}' | sf journal add --feature=X --json - --bridge-icm
journaled specforge/journal/2026-06-20-X.json (+ rendered .md)
staged for commit (NOT committed — that's your call)
bridged to ICM (topic specforge-journal)
```

`sf gate record-verdict` appends a phase-audit verdict to `audit.json` (separate
from human gates). `sf journal add` persists durable lessons to a git-tracked
journal — its own memory, not a hard dependency on any external tool; `--bridge-icm`
optionally mirrors to [ICM](https://github.com/rtk-ai/icm) if present. It
**stages but never commits** — the commit is your call.

### The hooks (per-harness — Claude Code adapter)

Hooks fire on the harness's events and call `sf`. This is the layer that makes
enforcement independent of the model. The core (`sf`) is portable; the hook
adapter is per-harness (a new harness = a new adapter, not new logic). Every hook
is **fail-open** (any error → allow) and a **no-op outside a SpecForge project**.

| Event | What the hook does |
|-------|--------------------|
| `SessionStart` | inject the resumed session + project invariants + consolidated learnings |
| `UserPromptSubmit` | inject the current-step slice (`sf context current`) so the spec doesn't dilute — cheap breadcrumb every turn, full slice on step-change |
| `PreToolUse` (Write/Edit) | **structural gates:** deny writing `design` before `requirements` is approved, `tasks` before `design`, `plan` before `tasks`; deny a 2nd feature's `requirements` while another is active (serial); deny edits to machine state |
| `Stop` | **memory reconciler:** if a feature was archived with no journal entry, nudge once to capture lessons |
| `PreCompact` | flag the session to re-inject the full slice next turn — after compaction the earlier slices were summarized away |
| `SessionEnd` | timestamp a continuity marker |

### The quality tier: phase auditor

The structural gates are hermetic but can only check *order and schema*, not
*quality*. For quality, `sf-check --phase=<phase>` runs an **adversarial judge in
a fresh sub-agent** — fresh context escapes the degradation a long session
suffers. It's fed by `sf context for-judge`, judges each mapped principle
`pass`/`fail` with a citation, and persists the verdict via `sf gate
record-verdict`. Opt-in and governed by config:

```jsonc
// constitution.json
{
  "principles": [
    { "id": "P-min", "statement": "Minimal code: climb the ladder before writing new code",
      "applies_to": ["design", "build"] }
  ],
  "audit": { "phase": "off" | "nudge" | "block" },   // default off
  "build": { "mode": "inline" | "single" | "per-wave" }, // default inline
  "coverage": {                                       // default: measure everything
    "exclude": [{ "path": "scripts/", "reason": "release tooling, not product" }]
  }
}
```

- **`audit.phase`** — `off` (skip), `nudge` (surface failures, don't block),
  `block` (stop the gate on a fail).
- **`build.mode`** — `inline` (waves run in the main context, classic), `single`
  (the whole build in one fresh sub-agent), `per-wave` (one fresh sub-agent per
  wave, for large features — each wave starts clean, with an automated checkpoint
  between waves).
- **`coverage.exclude`** — code that shouldn't be specified at all (generated,
  tooling, scripts) comes out of `sf coverage`'s denominator. A `path` is an
  exact file or a directory prefix ending in `/` — no globs, because a stray `*`
  silently excludes half a repo. `reason` is required: excluding *raises* the
  percentage, so an unexplained exclusion is how the metric gets dressed up
  without anyone noticing.

Quality principles like **`P-min`** (minimal code) are just constitution
principles with `applies_to` — the phase auditor checks them for free, no special
path. This is the cooperative tier: a fresh judge raises the floor on quality; it
doesn't guarantee it.
