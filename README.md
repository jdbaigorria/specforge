<p align="center">
  <img src="assets/specforge-logo.png" alt="SpecForge" width="380" />
</p>

# SpecForge

Spec-Driven Development framework. The specification is the product — code is a regenerable byproduct.

A 4-skill feature pipeline + `sf-audit` for project-wide review + support skills. Progressive disclosure. Human gate on every artefact. No ceremony without purpose.

**New here?** Read [the mental model](docs/mental-model.md) first — one page on how SpecForge thinks.

**Install:** see [INSTALL.md](INSTALL.md). **License:** [MIT](LICENSE).

---

## Table of contents

- [How it works](#how-it-works)
- [The deterministic layer (CLI `sf` + hooks)](#the-deterministic-layer-cli-sf--hooks)
- [Architecture](#architecture)
- [Skills reference](#skills-reference)
- [Artefact flow](#artefact-flow)
- [Spec format: EARS notation](#spec-format-ears-notation)
- [Feature lifecycle](#feature-lifecycle)
- [Backprop: how the spec learns](#backprop-how-the-spec-learns)
- [Example 1: Greenfield — CLI from scratch](#example-1-greenfield--cli-from-scratch)
- [Example 2: Brownfield — adding features to existing code](#example-2-brownfield--adding-features-to-existing-code)
- [Resync detection](#resync-detection)
- [FAQ](#faq)

---

## How it works

```
  sf-init ──▶ sf-propose ──▶ sf-build ──▶ sf-check
  scaffold    requirements    plan+execute  validate
  + context   design          wave by wave  + archive
              tasks

  The feature loop: 4 skills. Each produces artefacts. Human gate 🔴 on every artefact.
```

This is the per-feature loop. Two more pieces sit alongside it:
`sf-audit` runs a project-wide adversarial review (constitution vs reality,
cross-feature consistency), and a set of [support skills](SUPPORT-SKILLS.md)
(`sfx-think`, `sfx-triage`, `sfx-grill-me`, `sfx-tdd`, `sfx-documenter`, and more) complement the
pipeline without being part of it.

Detailed flow with gates:

```
sf-init
  ├─ scaffolding                          → automatic
  ├─ constitution (greenfield)            → 🔴 GATE
  ├─ onboard (brownfield)                 → 🔴 GATE
  └─ constitution (brownfield)            → 🔴 GATE

sf-propose
  ├─ requirements.md                      → 🔴 GATE
  ├─ design.md                            → 🔴 GATE
  └─ tasks.md                             → 🔴 GATE

sf-build
  ├─ execution plan                       → 🔴 GATE
  ├─ wave 0 execution                     → 🔴 GATE
  ├─ wave 1 execution                     → 🔴 GATE
  └─ wave N...                            → 🔴 GATE

sf-check
  ├─ review + verdict                     → 🔴 GATE
  ├─ APPROVE → archive (automatic)
  └─ REVISE  → back to sf-build with feedback
```

The REVISE loop is what makes this iterative. When check finds gaps, it sends
the feature back to build with specific corrections. If the spec itself was
wrong, the user edits the spec and resync detection cascades the changes.

---

## The deterministic layer (CLI `sf` + hooks)

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
halves standalone, for all six artifacts: constitution, requirements, design,
tasks, plan, review.)

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

**Emit the context slices the hooks inject — the "brain".**

```console
$ sf state current
{ "feature": "add-task-crud", "status": "building",
  "phase": "build", "wave": 1, "last_approved_gate": "wave-0" }
```

`sf state current` is a **pure function of `features.json`**: the active feature
(serial flow), the phase derived as *last-approved-gate + 1*, the wave from
`plan.json`. Zero stored state — nothing to keep in sync (no "second truth").

```console
$ sf context current --breadcrumb
SpecForge: feature `add-task-crud`, fase `build` (wave 1) · slice completo: `sf context current`

$ sf context for-wave --feature=add-task-crud --n=1   # the seed for a build sub-agent
$ sf context for-judge --phase=design --feature=X     # artifact + ONLY the principles that apply to that phase
```

`context current` is the minimal slice of the current step — a cheap breadcrumb,
or the full JSON. `for-wave` seeds a build sub-agent with one wave. `for-judge`
hands the quality auditor exactly the artifact plus the constitution principles
whose `applies_to` includes that phase — nothing more.

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
  "build": { "mode": "inline" | "single" | "per-wave" } // default inline
}
```

- **`audit.phase`** — `off` (skip), `nudge` (surface failures, don't block),
  `block` (stop the gate on a fail).
- **`build.mode`** — `inline` (waves run in the main context, classic), `single`
  (the whole build in one fresh sub-agent), `per-wave` (one fresh sub-agent per
  wave, for large features — each wave starts clean, with an automated checkpoint
  between waves).

Quality principles like **`P-min`** (minimal code) are just constitution
principles with `applies_to` — the phase auditor checks them for free, no special
path. This is the cooperative tier: a fresh judge raises the floor on quality; it
doesn't guarantee it.

---

## Architecture

### Directory structure

A single visible root, `specforge/`, holds everything. Review artefacts
(requirements, design, review) are visible so the human gates work; machine
state hides in `specforge/.state/`.

```
my-project/
├── specforge/                       # Single visible SpecForge root
│   ├── features.json                # Feature registry (status + gate ledger)
│   ├── constitution.json / .md      # Principles (with applies_to) + identity — JSON source + render
│   ├── history.md                   # Append-only project log
│   ├── roadmap.md                   # Feature roadmap
│   ├── learnings.md                 # Consolidated, evidence-anchored learnings (injected each session)
│   ├── features/
│   │   └── add-task-manager/
│   │       ├── requirements.json / .md   # JSON source + Markdown render (each artifact)
│   │       ├── design.json / .md
│   │       ├── tasks.json / .md          # flat tasks + depends_on
│   │       ├── trace.json                # requirement → code:symbol → test (drift anchors)
│   │       ├── review.json / .md
│   │       ├── audit.json                # phase-audit verdicts (quality tier)
│   │       ├── decisions/                # Complex technical decisions (optional)
│   │       └── progress/
│   │           ├── plan.json / .md       # computed wave layout (sf plan compute)
│   │           ├── wave-0.md
│   │           └── wave-1.md
│   ├── archive/
│   │   └── 2026-05-12-add-auth/
│   ├── audits/                      # sf-audit reports
│   ├── journal/                     # durable lessons (sf journal add) — git-tracked memory
│   ├── context/                     # Project context + skill outputs (visible)
│   │   ├── project.md               # Stack, architecture (brownfield inferred)
│   │   ├── conventions.md           # Code conventions
│   │   ├── compact-rules.md         # Condensed rules for sub-agents
│   │   └── thinks/ triages/ briefs/ grills/ …  # support-skill artefacts
│   └── .state/                      # Hidden machine state (not human-edited)
│       ├── session.md               # Session cache / recovery (not source of truth)
│       └── hook-context.json        # per-session slice-injection trigger state
│
└── src/                             # Your code
```

### Skill structure (progressive disclosure)

Each skill has a lean SKILL.md (<200 lines) and loads capabilities on demand
from `references/`. This keeps the context small — the agent only loads what
the current step needs.

```
sf-init/
├── SKILL.md
├── references/
│   ├── constitution.md              # Greenfield: conversation guide
│   └── onboard.md                   # Brownfield: codebase analysis
└── templates/
    ├── constitution.tmpl.md
    ├── project.tmpl.md
    └── conventions.tmpl.md

sf-propose/
├── SKILL.md
├── references/
│   ├── design-first.md              # Alt flow: architecture → requirements
│   ├── from-code.md                 # Brownfield: infer specs from code
│   ├── clarify.md                   # Post-generation refinement
│   ├── research.md                  # Technical decision documentation
│   └── ears-notation.md             # EARS syntax reference
└── templates/
    ├── requirements.tmpl.md
    ├── design.tmpl.md
    └── tasks.tmpl.md

sf-build/
├── SKILL.md
├── references/
│   ├── task-planning.md             # Run `sf plan compute` (waves from deps) + refine
│   └── wave-execution.md            # Per-wave execution + the minimal-code ladder
└── templates/
    └── progress.tmpl.md

sf-check/
├── SKILL.md
├── references/
│   ├── backprop.md                  # Pattern 3x → invariant promotion
│   ├── minimal-code.md              # Parsimony rubric (P-min audit)
│   └── archive.md                   # Sync + close procedure
└── templates/
    └── review.tmpl.md
```

### Execution model

The pipeline skills run **inline** by default — in the main conversation context
— because every artefact has a human gate that requires interaction, and a
sub-agent can't stop to ask for approval. The one exception is **build
execution**: once the plan gate is approved (the human authorizes it), the build
can run in a fresh sub-agent (`build.mode = single | per-wave`) — the plan is
already the contract, so execution needs no further gate until it reports back.
Build is where most tokens burn, so offloading it keeps the main context clean.
See [the deterministic layer](#the-deterministic-layer-cli-sf--hooks).

Support skills follow the same rule, decided by **interactivity, not tier**:

- **Gated or iterative** (`sfx-grill-me`, `sfx-tdd`) → inline.
- **Pure transform** — takes input, returns output, no human turn in the middle
  (`sfx-documenter`, `sfx-explain`, `sfx-aws-architect`, `sfx-data-engineer`, `sfx-triage`,
  `sfx-github`) → may be delegated to a sub-agent **where the harness supports it**,
  falling back to inline otherwise. Delegation is an optional, per-harness
  optimization, not part of the portable core.

The anti-telephone-game principle: artefacts live on disk. When a skill needs
context from a previous artefact, it reads the file — it doesn't rely on
conversation history. This means context window carries references, not payloads.

---

## Skills reference

### sfp-scout (product tier, optional)

The from-zero front-end. De-risks a fuzzy idea before it becomes a project —
research the landscape, stress-test it, decide proceed/pivot/kill. Greenfield
only; brownfield skips straight to `sf-init`.

| | |
|---|---|
| **Triggers** | `sfp-scout`, "should I build X", "is this worth building", "de-risk this idea" |
| **Needs** | the research MCPs (web/GitHub/docs) — won't fabricate research without them |
| **Produces** | `specforge/product/<slug>/` — research, discovery brief (with `PR#` ids), decision log |
| **Hands off** | on *proceed* → `sf-init --from <brief>`; `PR#` ids flow into the roadmap |

De-risk, not validation — it gathers evidence and surfaces risk; it cannot prove
demand. Its product requirements (`PR#`) sit at the top of the traceability spine:
`PR# → feature → R# → task → code → test`.

---

### sf-init

Scaffolds the project and generates foundational context.

| | |
|---|---|
| **Triggers** | `sf-init`, "start a new project", "add specforge to this project" |
| **Flags** | `--from <path>` (import PRD/brief), `--path <dir>` (monorepo sub-project) |
| **Detects** | Greenfield vs brownfield (interactive for ambiguous cases like monorepos) |

**Greenfield flow:**
1. Scaffold directories (automatic)
2. Constitution conversation: identity, principles, constraints, anti-goals → 🔴 GATE
3. Generate `specforge/context/project.md` + `specforge/context/conventions.md` from constitution context → 🔴 GATE

**Brownfield flow:**
1. Scaffold directories (automatic)
2. Onboard: analyze codebase → `specforge/context/project.md` + `specforge/context/conventions.md` → 🔴 GATE
3. Constitution conversation: principles and anti-goals (shorter, stack already known) → 🔴 GATE

**Produces:**
- `specforge/constitution.md` — project identity, principles, constraints, anti-goals
- `specforge/features.json` — empty feature registry
- `specforge/history.md` — project log (first entry)
- `specforge/context/project.md` — stack and architecture context
- `specforge/context/conventions.md` — coding standards

---

### sf-propose

Generates the full specification for a feature: requirements, design, tasks.

| | |
|---|---|
| **Triggers** | `sf-propose <name>`, "specify", "add feature", "spec this" |
| **Flags** | `--design-first` (architecture → requirements), `--from-code` (reverse-engineer) |
| **Modes** | Requirements-first (default), Design-first, From-code |

**Requirements-first flow (default):**
1. Conversation: what, why, who, boundaries
2. Generate `requirements.md` with EARS notation → 🔴 GATE
3. Generate `design.md` (sections conditional on complexity) → 🔴 GATE
4. Generate `tasks.json` — a flat task list with `depends_on` + traceability matrix → 🔴 GATE
5. Register feature in `features.json` (status: `approved`)

**Design-first flow** (`--design-first`):
Inverts steps 2 and 3 — design first, then derive requirements from what the
architecture can deliver. For projects where technical constraints drive scope.

**From-code flow** (`--from-code`):
Reverse-engineers specs from existing code. Creates a `_baseline` feature with
`[INFERRED]` requirements. No tasks.md (already implemented).

**Produces per feature:**
- `specforge/features/<name>/requirements.md` — EARS requirements + acceptance criteria
- `specforge/features/<name>/design.md` — architecture, components, decisions
- `specforge/features/<name>/tasks.json` (+ `.md`) — flat tasks + `depends_on` (waves are computed at build)

**Design sections are conditional.** Only include sections relevant to the feature.
A CLI flag doesn't need Security Considerations. A payment endpoint does. Sections
that would say "N/A" are omitted to avoid context noise.

**Post-generation:** User can invoke clarify at any time (reads `references/clarify.md`).
Scans for ambiguities, completeness gaps, and inconsistencies.

---

### sf-build

Plans execution and implements wave by wave.

| | |
|---|---|
| **Triggers** | `sf-build <name>`, "build", "implement", "start building" |
| **Prerequisite** | Feature status must be `approved` |
| **Resumable** | If paused, reads `progress/` to find where it left off |

**Flow:**
1. `sf plan compute` — compute the wave layout from the tasks' `depends_on`
   graph (topological layering); optionally refine → 🔴 GATE
2. For each wave (per `build.mode`: inline, or a fresh sub-agent):
   a. Execute all tasks in the wave — climbing the minimal-code ladder per task
   b. Log progress to `progress/wave-<n>.md`
   c. Present results → 🔴 GATE (inline) / automated checkpoint (sub-agent)
3. All waves complete → status becomes `checking`

**Error recovery (3 levels):**
- **Minor** (typo, wrong import): fix inline, note in progress log
- **Design mismatch** (can't implement as specified): pause wave, present options to user → 🔴 GATE
- **Blocking dependency** (previous wave output wrong): pause, escalate to user

**Key rule:** Never modify specs during build. If specs need changes, pause and
escalate. Specs are the contract; build fulfills them.

**Produces per wave:**
- `specforge/features/<name>/progress/plan.json` (+ `.md`) — computed wave layout
- `specforge/features/<name>/progress/wave-<n>.md` — what was done, decisions, ladder rung
- Updated `tasks.json` — tasks marked `done` as completed

---

### sf-check

Validates implementation against specs. Archives on approval.

| | |
|---|---|
| **Triggers** | `sf-check <name>`, "check", "validate", "review" |
| **Prerequisite** | Feature status must be `checking` |
| **Includes** | Traceability, gap analysis, constitution compliance, backprop |

**Flow:**
1. Traceability analysis: every R# → task → implementation → test
2. Gap analysis: missing implementations, tests, design deviations, orphan code
3. Constitution compliance: validate against principles — including **parsimony**
   (`P-min`: over-engineering is a valid REVISE reason; see `references/minimal-code.md`)
4. Verdict: APPROVE / APPROVE WITH NOTES / REVISE
5. Present review → 🔴 GATE
6. If APPROVE → archive automatically

It also runs as a **phase auditor** (`sf-check --phase=<phase>`) — a narrow,
fresh-sub-agent quality check of a single just-finished phase at its gate, instead
of the full end-of-feature pass. See [the quality tier](#the-quality-tier-phase-auditor).

**Verdicts:**
- **APPROVE** — all requirements implemented + tested, no violations
- **APPROVE WITH NOTES** — implemented, minor gaps documented for future
- **REVISE** — significant gaps, back to `sf-build` with specific corrections

**User can override verdict.** If the user disagrees with the assessment,
they can override. The override is logged in the review.

**Produces:**
- `specforge/features/<name>/review.md` — traceability matrix, gaps, verdict
- `specforge/archive/<date>-<name>/` — complete feature archive (on approve)
- Updated `specforge/history.md` — completion entry
- Updated `specforge/constitution.md` — if backprop promotes a new invariant

---

### sf-audit

Project-wide adversarial audit. Not part of the per-feature loop — it steps back
and reviews the whole project: constitution vs reality, cross-feature
consistency, drift, and accumulated gaps.

| | |
|---|---|
| **Triggers** | `sf-audit`, "audit the project", "is the constitution still true" |
| **Scope** | All features + constitution, not a single feature |
| **Produces** | `specforge/audits/<date>.md` — findings, severity, recommendations |

---

### sf-amend

Modify a feature that already shipped, without forking a parallel spec. Archived
specs are **living documents** — `sf-amend` runs a delta mini-pipeline
(propose-delta → gate → build → check) that edits the existing requirements,
design, tasks, and `trace.json` in place.

| | |
|---|---|
| **Triggers** | `sf-amend <feature>`, "change the shipped feature", "the archived spec is out of date" |
| **vs sf-propose** | amend changes an existing capability; propose creates a new one |
| **Edits in place** | one evolving traceability matrix per feature — never a second parallel spec |

Pair it with drift detection: `sf doctor --drift`
reads each archived feature's `trace.json` and flags requirements whose code
anchor vanished or whose test fails — so you find out the spec drifted before it
becomes a lie.

---

### Support skills

Standalone skills that complement the pipeline but are not part of it. They work
without `specforge/` initialized and produce artefacts in `specforge/context/`. See
[SUPPORT-SKILLS.md](SUPPORT-SKILLS.md) for the full reference: `sfx-think`,
`sfx-triage`, `sfx-grill-me`, `sfx-tdd`, `sfx-documenter`, `sfx-explain`,
`sfx-aws-architect`, `sfx-data-engineer`, `sfx-github`, `sfx-journal`.

---

## Artefact flow

Each artifact below is a `.json` **source** + a `.md` **render** (`sf` generates
the Markdown from the JSON):

```
sf-init produces:
  specforge/constitution.json / .md       (principles with applies_to)
  specforge/context/project.md
  specforge/context/conventions.md

sf-propose produces (per feature):
  specforge/features/<name>/requirements.json / .md
  specforge/features/<name>/design.json / .md
  specforge/features/<name>/tasks.json / .md          (flat tasks + depends_on)

sf-build produces (per feature):
  specforge/features/<name>/progress/plan.json / .md  (sf plan compute — waves from deps)
  specforge/features/<name>/progress/wave-<n>.md
  specforge/features/<name>/audit.json                (if phase auditor ran)

sf-check produces (per feature):
  specforge/features/<name>/review.json / .md
  specforge/features/<name>/trace.json                (requirement → code → test anchors)
  specforge/journal/<date>-<name>.json / .md          (durable lessons, on archive)
  specforge/archive/<date>-<name>/                    (on APPROVE)
```

**Who reads what:**

| Artefact | Created by | Read by |
|----------|-----------|---------|
| constitution.md | sf-init | sf-propose (constraints), sf-check (compliance) |
| specforge/context/project.md | sf-init | sf-propose (stack context), sf-build (conventions) |
| specforge/context/conventions.md | sf-init | sf-build (coding standards) |
| requirements.json | sf-propose | sf-build (traceability), sf-check (validation) |
| design.json | sf-propose | sf-build (architecture guide), sf-check (adherence) |
| tasks.json | sf-propose | `sf plan compute` (deps → waves), sf-build (execution) |
| plan.json | `sf plan compute` | sf-build (`sf context for-wave`), sf state current |
| progress/*.md | sf-build | sf-check (audit trail) |
| trace.json | sf-check | `sf doctor --drift`, `sf trace verify` |
| audit.json | `sf gate record-verdict` | the human (phase-audit verdicts) |
| review.json | sf-check | archive (verdict determines archival) |
| journal/*.json | `sf journal add` | the phase auditor (past lessons), ICM (optional) |
| features.json | sf-init | all skills + `sf` (status gate, gate ledger) |
| history.md | sf-init | sf-check (backprop pattern tracking) |

---

## Spec format: EARS notation

SpecForge uses EARS (Easy Approach to Requirements Syntax) for unambiguous requirements:

```
WHEN [trigger] THE SYSTEM SHALL [behavior]          ← event-driven
WHILE [state] THE SYSTEM SHALL [behavior]           ← state-driven
IF [unwanted condition] THEN THE SYSTEM SHALL [fix]  ← error handling
THE SYSTEM SHALL [behavior]                          ← always active
```

One requirement per statement. Active voice. Specific and testable.
Full reference in `sf-propose/references/ears-notation.md`.

---

## Feature lifecycle

```json
// specforge/features.json
{
  "features": [
    {
      "name": "user-auth",
      "status": "done",
      "workflow": "requirements-first",
      "created": "2026-05-10",
      "completed": "2026-05-12"
    }
  ]
}
```

Statuses: `pending` → `proposing` → `approved` → `building` → `checking` → `done`

Each transition is triggered by a skill completing its work:
- sf-propose sets `approved` after all 3 artefacts are gate-approved
- sf-build sets `checking` after all waves complete
- sf-check sets `done` after APPROVE + archive

---

## Backprop: how the spec learns

When the same type of issue appears across 3+ features, it becomes a project invariant.

```
Feature A: API endpoint missing error handling     ← occurrence 1
Feature B: CLI command crashes on bad input        ← occurrence 2
Feature C: Webhook handler ignores timeout         ← occurrence 3 → PROMOTE

→ New invariant in constitution.md:
  "Every external interface must have explicit error handling."
```

Tracking lives in `specforge/history.md`:

```markdown
## Recurring Issues
| Issue Pattern | Occurrences | Features | Status |
|---------------|-------------|----------|--------|
| Missing error handling | 3 | auth, cli, api | → INVARIANT |
```

After promotion, sf-check validates every future feature against the new invariant.

---

## Example 1: Greenfield — CLI from scratch

**Goal:** Task manager CLI with Python, Typer, Rich.

```
> sf-init

  Scaffolding...
  ✓ specforge/ created
  ✓ specforge/context/ created
  Detected greenfield project.

  Let's define your project.
  What are you building? → "A CLI task manager"
  Who is it for? → "Developers who want a fast terminal-based todo"
  What problem does it solve? → "Existing tools are bloated, I want something minimal"

  Principles?
  → "Speed over features — fast MVP"
  → "Terminal-native — no TUI, just commands"
  → "Local-first — JSON file, no server"

  Constraints?
  → "Python 3.12, Typer, Rich. Must work on macOS and Linux."

  Anti-goals?
  → "Not a project manager — no teams, boards, or sprints"
  → "Not synced — single device only"

  🔴 GATE: constitution.md
  User: "Approved"

  🔴 GATE: project.md + conventions.md
  User: "Approved"

  ✓ Initialized. Run sf-propose <feature-name> to start.
```

```
> sf-propose add-task-crud

  What does this feature do? → "Add, list, complete, delete tasks"
  Who uses it? → "The developer, from the terminal"
  Boundaries? → "No due dates, no priorities, no categories"

  🔴 GATE: requirements.md
  R1: Add task (WHEN user runs `tasks add "text"` THE SYSTEM SHALL create task)
  R2: List tasks (WHEN user runs `tasks list` THE SYSTEM SHALL display table)
  R3: Complete task (WHEN user runs `tasks done <id>` THE SYSTEM SHALL mark done)
  R4: Delete task (WHEN user runs `tasks rm <id>` THE SYSTEM SHALL remove task)
  R5: Persist (THE SYSTEM SHALL store tasks in ~/.tasks/tasks.json)
  User: "Approved"

  🔴 GATE: design.md
  Components: TaskModel, JsonStorage, CLI commands
  No Security/Performance sections (not relevant for this feature)
  User: "Approved"

  🔴 GATE: tasks.md
  Wave 0: T1 TaskModel + JsonStorage → R5
  Wave 1: T2 add cmd → R1, T3 list cmd → R2
  Wave 2: T4 done cmd → R3, T5 rm cmd → R4
  Traceability: 5/5 requirements covered ✓
  User: "Approved"

  ✓ Feature approved. Run sf-build add-task-crud.
```

```
> sf-build add-task-crud

  🔴 GATE: Execution plan
  Wave 0: models.py + storage.py (2 files, low complexity)
  Wave 1: add.py + list.py (2 files, medium)
  Wave 2: done.py + rm.py (2 files, low)
  User: "Approved"

  Executing Wave 0...
  ✓ T1: Created models.py, storage.py, test_storage.py
  🔴 GATE: Wave 0 results
  User: "Approved"

  Executing Wave 1...
  ✓ T2: Created commands/add.py
  ✓ T3: Created commands/list.py with Rich table
  🔴 GATE: Wave 1 results
  User: "Approved"

  Executing Wave 2...
  ✓ T4: Created commands/done.py
  ✓ T5: Created commands/rm.py
  🔴 GATE: Wave 2 results
  User: "Approved"

  ✓ Built. Run sf-check add-task-crud.
```

```
> sf-check add-task-crud

  Traceability: 5/5 requirements implemented ✓
  Tests: 5/5 requirements have tests ✓
  Constitution: all principles respected ✓
  Verdict: APPROVE

  🔴 GATE: review.md
  User: "Approved"

  ✓ Archived to specforge/archive/2026-05-12-add-task-crud/
```

---

## Example 2: Brownfield — adding features to existing code

**Goal:** Existing task manager CLI. Add priority levels (high/medium/low).

### First: initialize SpecForge on the existing project

```
> sf-init

  Detected existing code: Python 3.12, Typer, Rich, JSON storage.
  Detected brownfield project.

  Analyzing codebase...
  🔴 GATE: project.md + conventions.md
  User: "Approved, but change test convention to pytest-style"
  Updated.

  Now let's define principles for this project.
  What principles guide development?
  → "Speed over features"
  → "Local-first"
  → "Terminal-native"
  Anti-goals?
  → "Not a project manager"
  🔴 GATE: constitution.md
  User: "Approved"
```

### Optional: document what exists

```
> sf-propose --from-code

  Analyzing existing code...
  Inferred 5 requirements from codebase [INFERRED]:
  R1: Add task, R2: List tasks, R3: Complete task, R4: Delete task, R5: Persist

  🔴 GATE: requirements.md
  User: "R2 is missing the empty state message, adding it. Rest looks good."
  Updated.

  🔴 GATE: design.md (current architecture as-is)
  User: "Approved"

  ✓ Baseline registered as _baseline (status: done)
```

### Then: add the new feature

```
> sf-propose add-priorities

  What does this feature do? → "Add high/medium/low priority with color-coded
  Rich output and a --priority filter on list"

  🔴 GATE: requirements.md
  R1: Set priority on add (WHEN user runs `tasks add --priority high "text"`...)
  R2: Display priority in list (WHEN tasks are listed THE SYSTEM SHALL color-code...)
  R3: Filter by priority (WHEN user runs `tasks list --priority high`...)
  R4: Default priority (IF no priority specified THE SYSTEM SHALL assign "medium")
  User: "Approved"

  🔴 GATE: design.md
  Modifies: TaskModel (adds priority field), JsonStorage (schema migration),
  list command (color mapping), add command (new flag)
  User: "Approved"

  🔴 GATE: tasks.md
  Wave 0: T1 Update TaskModel + storage schema → R4
  Wave 1: T2 Update add command → R1, T3 Update list command → R2, R3
  Traceability: 4/4 covered ✓
  User: "Approved"

> sf-build add-priorities
  (waves execute, gates on each)

> sf-check add-priorities
  Verdict: APPROVE
  ✓ Archived
```

---

## Resync detection

When sf-propose is invoked on a feature that already has artefacts:

1. Check timestamps: if `requirements.md` is newer than `design.md` → stale
2. Prompt: "requirements.md was modified after design.md was generated. Regenerate design? [y/n]"
3. If yes → regenerate from the updated upstream artefact
4. Cascade: if design changes → offer to regenerate tasks too

This handles the common case where the user edits a spec file directly and
downstream artefacts need to catch up.

---

## FAQ

**Why is the feature pipeline only 4 skills?**
Progressive disclosure. Capabilities that were separate skills (clarify, research,
map, archive, explore, constitute) now live as references inside the 4 pipeline
skills. They load on demand. Less context overhead, less cognitive load. The
pipeline is deliberately small — but it is not the whole framework: `sf-audit`
adds project-wide review, and the [support skills](SUPPORT-SKILLS.md) cover
thinking, triage, TDD, docs, and infra design around it.

**Isn't the full pipeline overkill for a typo or a config tweak?**
That's why there are two lanes (F34). `sf-propose` opens by classifying the
change and proposing a **lite** lane for trivial, low-risk edits — one combined
`change.md`, one gate, build, a minimal check — versus the **standard** full
flow. You don't pick the lane to skip work; the framework proposes it and you
approve it at a gate, and it's recorded in `features.json`. Lite still writes a
test and a `trace.json`, so it stays inside drift detection — lighter ceremony,
not lower integrity. If a lite change turns out bigger than it looked, it's
promoted to standard mid-flight (the escape hatch only goes up).

**What happens to a spec after the feature is archived? Doesn't it go stale?**
That's the classic SDD failure, and SpecForge treats the archived spec as a
**living document**, not a frozen snapshot (the historical snapshot is just the
git commit). `archive` seals the feature with a live link to code — `trace.json`,
the structured matrix mapping each requirement to its `path:symbol` and test. To
change a shipped feature you run `sf-amend`, which edits that spec and matrix in
place instead of forking a parallel one. And `sf doctor --drift` reads
`trace.json` to tell you when code moved out from under a requirement — cheaply,
because it only checks the exact anchors, not the whole repo.

**Does this work for a team, or only solo? How does it relate to PR review?**
It works for a team without building its own permission system — it leans on
git/PR (F35/F36). The creation gates (propose/build) belong to the author on a
`feature/<slug>` branch; the **verdict gate maps to the PR approval** — the
artefacts travel in the PR, so the reviewer approves code and spec together (map,
don't duplicate). Ownership is `owners` in the constitution + git `CODEOWNERS`.
The git convention is one branch per feature, one commit per wave, `archive` =
merge. And `sfx-github` can export the roadmap to issues **one-way** (the tracker
indexes *what*, SpecForge owns the detail — no fragile bidirectional sync).

**Can I use SpecForge with any AI agent?**
Yes. Skills are markdown files. Any agent that reads markdown can execute them.
The AGENT.md orchestrator targets Claude Code but the skills are agent-agnostic.

**What if check keeps returning REVISE?**
REVISE sends you back to sf-build with specific corrections. If the spec itself
is wrong, edit it directly and resync detection will cascade the changes.

**Can multiple features be active simultaneously?**
Not yet — the flow is **serial** in this version: one active feature at a time,
the rest stay `queued` in `features.json`. This keeps gates and the session
checkpoint unambiguous. Each feature still has its own folder, so parallel
features are a planned future capability (tracked `active_feature` + per-feature
session sections); for now, archive or park the current one before starting
another.

**How is this different from OpenSpec / Spec Kit / CaveKit?**
SpecForge combines: constitution + identity from Spec Kit, change organization
from OpenSpec, wave execution + backprop from CaveKit. Human gates on every
artefact, EARS notation, resync detection, living specs with drift detection, and
progressive disclosure are unique to SpecForge. And where SDD tools (Spec Kit
included) are strongest once you *know* what to build, SpecForge also covers the
step before — `sfp-scout` de-risks a fuzzy idea from zero — so it spans the whole
arc, idea → de-risked → spec → build → check → kept alive.

**Can it help me figure out WHAT to build, or only build a known spec?**
Both. For a known idea, start at `sf-init`. For a fuzzy one, start at `sfp-scout`:
it researches the landscape (via the research MCPs), stress-tests the idea, and
returns a discovery brief with a proceed/pivot/**kill** verdict — then hands off
to `sf-init`. It de-risks; it doesn't claim to validate demand. The vision/identity
itself is still captured by `sf-init`'s constitution conversation.
