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

**Inspect the opt-in guarantees** (see `verification.*` below — each is off until
you turn it on, and each answers a different question):

```console
$ sf coverage --by-priority       # of my requirements, which have a satisfied contract?
$ sf arch rules --feature=X       # the dependency rules derived from the approved design
$ sf arch check --feature=X       # does the built code respect that graph?
$ sf mutation scope --feature=X   # which files would the mutator touch?
$ sf mutation run --feature=X     # does the suite actually detect injected defects?
$ sf evidence checklist --feature=X  # what a human still has to verify, and what expired
$ sf question list                # unknowns still waiting on someone
$ sf domain terms                 # the same concept named differently across features
```

**`sf domain terms`** catches the defect you cannot see by reading one feature,
because each spec is internally coherent. It surfaces when two features nobody
read side by side finally have to integrate and it turns out `Order` and `Pedido`
were the same thing — or worse, were not, and nobody noticed until the data
disagreed. In consulting work this is the single largest source of rework.

It is fully deterministic: `domain.json` already declares the ubiquitous language
with its aliases, and the requirements are already written, so *"which surface
form does each feature use for this term?"* is a regex, not an inference.

Matching uses **word boundaries** — `Order` does not match inside `Reordering` —
and is **case-insensitive**, because `order` and `Order` are the same word and
flagging that would be noise shaped like a finding.

Two findings, and only one moves the exit code:

- **inconsistent** — one glossary term appears under more than one of its
  declared forms. Exit 1. Two names for one thing is never fine.
- **unused** — a glossary term no live requirement mentions. Reported, exit
  unchanged: it is often vocabulary that has not reached a spec yet, and a check
  that goes red on every real project from day one is a check people stop running.

**Open questions block a verdict until somebody decides.** An unknown you are
waiting on — typically a client answer that takes weeks — is a first-class
object, not a note that dies with the session:

```console
$ sf question add --text="is VAT included?" --impact="changes Order and the total" --blocks=checkout
$ sf question answer --id=Q1 --answer="yes, included" --source=S3
$ sf question adopt  --id=Q1 --answer="VAT included — industry practice"
```

`--impact` is required: an unknown with no stated consequence cannot be ranked
against the others, and a list nobody ranks is a list nobody reads. A question
listing `--blocks` stops that feature's verdict, and there are exactly **two**
ways out, both leaving a record — answer it, or adopt an external practice.

The third way — quietly dropping the thing from scope because no answer came —
is the one path the gate does not offer, and that is the whole point. Removing
scope without a conscious decision costs more than leaving it in: the bill
arrives months later, when nobody remembers there was ever a question.

`adopt` is **not** a synonym for `answer`. An external practice adopted because
the client never replied is not a fact the client confirmed, and it stays
labelled that way forever. The day someone asks "why is it built like this?",
that distinction is the entire answer.

An unrecognised `status` in `questions.json` counts as **open** — fail-closed, so
hand-editing a `"wontfix"` in is not a cheaper way past the gate.

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
  },
  "verification": {                                   // default: everything blocks
    "blocking_priorities": ["must", "should"],
    "require_arch": false,                            // opt-in, all three
    "require_mutation": false                         // independent of each other
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
- **`verification.blocking_priorities`** — which requirement priorities
  (`must` / `should` / `could`) can *block* the verdict gate. **Omitting it
  blocks on all three**, exactly as before the field existed; an empty list means
  the same thing, not "block on nothing" — a `[]` typed by mistake would
  otherwise disable the gate silently. Lowering the bar is an explicit project
  decision, written in a gated artifact and therefore auditable.

  Non-blocking violations are still **reported** every time — by `sf gate
  approve` on stderr and in `sf verify`'s note. Deciding something shouldn't
  stop a release is not deciding to stop looking at it.

  Priority grades the *verification contract* only. A broken code anchor blocks
  regardless: an anchor that no longer resolves isn't a minor requirement
  without a test, it's an artifact lying about where its subject lives, and that
  corrupts drift detection for the whole project.
- **`verification.require_source`** — default `false`. With `true`, every
  requirement must cite a `sources.json` id. Off by default because a founder's
  own idea has no client source and that's legitimate; on, it's the check that
  stops a model from inventing requirements. Either way, `sf sources coverage`
  reports both directions — requirements with no source, and sources no
  requirement uses.
- **`verification.require_red_witness`** — default `false`. With `true`, the
  verdict requires every test named in `trace.json` to have been **observed
  failing against different code** than the code that passes it today.

  The gap it closes: the CLI already proves a test exists, ran, went green and is
  fresh — but not the **order**. A test written after the code tends to assert
  what the code does instead of what the requirement asks. That's the most common
  and least malicious form of semantic fraud: the agent implements, reads its own
  implementation, and writes a test describing it.

  **This does not impose TDD.** The witness accumulates on its own — `sf check
  run` records any test it sees fail, and whoever works RED-GREEN builds the
  record without doing anything special. Turning the flag on is what makes it a
  requirement, and you turn it on once the record already has history; on day one
  it would reject every existing project.

  **Honest limits.** It proves there was a state of the code where the test
  failed — not strict ordering. A test written afterwards that failed because of
  a bug also counts. It's a lower bound, not a demonstration of TDD. And it needs
  `build.report` (`go-json` | `junit`): without a per-test report no failure can
  be attributed to a test, so the flag **rejects asking for the configuration**
  rather than approving in silence. Requirements whose `verification` isn't
  `test` are exempt — they never failed as a test because they never were one.
- **`verification.require_arch`** — default `false`. With `true`, the verdict
  requires the built code to respect the dependency graph `design.json` declared
  and the design gate approved.

  The gap it closes: `component.depends_on` passes a human gate, gets sealed, and
  then **nobody looks at it again**. A wave could implement every requirement,
  name every test, pass causality, go green and seal **having violated the
  approved design outright**. The seal certified traceability and verification; it
  never certified that the code respects its own declared structure.

  SpecForge doesn't implement the analyzer — it shells out to your stack's tool
  (`go-arch-lint` / `depguard`, `import-linter`, `dependency-cruiser`) via
  `build.arch_cmd` and consumes the **exit code**. What it adds is what none of
  them can: **the rule is generated from the approved design**, so when the design
  changes through `sf-amend` the rule changes with it and both stay under the same
  seal.

  ```jsonc
  "build": { "arch_cmd": "go-arch-lint check --config={config}" },
  "verification": { "require_arch": true }
  ```

  **`{config}` is required.** Without it the tool would check a hand-written
  config instead of the approved graph — that's the guarantee inverted, not a
  config that's merely incomplete, so it's a validation error. `sf` writes the
  rules to `specforge/.state/arch-rules-<feature>.json` and regenerates them every
  run; inspect them with `sf arch rules --feature=NAME` before wiring the tool.

  **Where the component→files mapping comes from.** `tasks.json` already carries
  it: `component_refs` × `files_touched`, both approved at the tasks gate. No new
  schema, and the source stays a sealed artifact. When it can't be derived —a file
  claimed by two components, a component with no production files, an edge to a
  component nobody declared— it **stops and asks for the mapping** instead of
  guessing. An invented mapping makes every conformance check afterwards
  decorative.

  **Honest limits.** It checks the dependency graph, which is the mechanizable
  part of an architecture — cohesion, naming and well-split responsibilities are
  the phase auditor's job, not this one. Test files are excluded: a test crossing
  component boundaries is normal, and counting it would produce false violations.
- **`verification.require_mutation`** — default `false`. With `true`, the verdict
  requires your mutation tool to pass its own threshold.

  What it proves that the RED witness doesn't: the witness shows a test *could*
  fail once; mutation shows the suite **detects defects now**. Defects get injected
  into the code (flip an operator, negate a condition, move a boundary) and some
  test must go red. If the suite stays green with the code broken, the suite
  doesn't verify — it accompanies.

  ```jsonc
  "build": { "mutation_cmd": "gremlins unleash --tags={files}" },
  "verification": { "require_mutation": true }
  ```

  `{files}` is substituted with the production files this feature's trace anchors
  — **the advantage no generic tool has**, because none of them knows which code
  belongs to which requirement. It's what makes the runtime tolerable. Unlike
  `{config}` above, it's **optional**: omitting it just means the command defines
  its own scope, which is slower but not wrong, so it's a warning. Inspect the
  scope with `sf mutation scope --feature=NAME`.

  **Only the exit code is consumed, never a parsed score.** The threshold is your
  tool's configuration. This is also the only honest stance on **equivalent
  mutants**: some are semantically identical to the original and no test can ever
  kill them, so the score is permanently imperfect and demanding 100% would demand
  the impossible.

  **Honest limits.** It's slow — it runs the suite once per mutant, so it's a
  feature-close gate, never something to run per wave or per PR. It runs last in
  the verdict for that reason: a broken anchor should surface in seconds, not
  after twenty minutes of mutation that was going to reject anyway. And not every
  stack has a tool; without `mutation_cmd` the check is skipped, not failed.

- **`verification.evidence_max_age_days`** — default `0` (off). With a number,
  evidence recorded further back than that many days **stops counting** and the
  verdict blocks.

  It closes the other half of `verification` ≠ `test`. `RM-C4` made a
  non-automated requirement declare its evidence; this asks whether that evidence
  still describes *this* system. A security audit from eight months and forty
  merges ago is not a current statement about the code.

  ```jsonc
  "verification": { "evidence_max_age_days": 180 }
  ```

  Work the circuit with `sf evidence`:

  ```console
  $ sf evidence checklist --feature=X   # what a human still has to verify
  $ sf evidence record --feature=X --scenario=R1.1 --kind=benchmark --ref=docs/bench.md
  ```

  The checklist is **derived from the spec**, never hand-kept: it lists exactly
  the criteria whose requirement declares a non-test `verification`, and marks
  each `missing`, `recorded` or `stale`. `record` writes into `trace.json` for
  you — that file is authoritative state the hook protects, and asking someone to
  rewrite a whole document to note three fields is how people end up noting
  whatever makes it pass. It validates with the **same rules as the gate**, so a
  mismatched `kind` or an unopenable `ref` is refused there and then.

  **`stale` and `missing` are different states on purpose**, and the tool never
  merges them: one means re-verify, the other means verify for the first time. If
  an expired item just read as "evidence missing", the cheap way out would be
  re-dating the old evidence without re-checking anything — precisely the fraud
  this exists to catch. Every message says so.

  **Honest limit — this is the one check that can turn a green project red with
  nobody touching it.** Time passes on its own. That's why it's off by default:
  deciding that a manual verification expires is a project's call about its own
  rate of change, not a default anyone can pick for it.

**The three opt-ins are orthogonal on purpose.** `require_red_witness`,
`require_arch` and `require_mutation` are independent flags, not steps on a
ladder: a project can demand architecture conformance (seconds) without paying
for mutation testing (minutes). Bundling them behind a single "strict mode" would
have made the cheap guarantee cost as much as the expensive one, and the
predictable result is that nobody turns any of them on.

**All three reject when switched on without their command configured**, instead
of degrading quietly. Approving would be worse than blocking: the project would
believe it holds a guarantee that was never evaluated. Running the commands by
hand (`sf arch check`, `sf mutation run`) *does* degrade gracefully — that's a
question, not a gate somebody is about to cross.

Quality principles like **`P-min`** (minimal code) are just constitution
principles with `applies_to` — the phase auditor checks them for free, no special
path. This is the cooperative tier: a fresh judge raises the floor on quality; it
doesn't guarantee it.
