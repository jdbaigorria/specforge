---
name: sf-build
description: >
  Plan and execute feature implementation from approved specs. Use when the user says
  "sf-build", "build feature", "implement", "start building", "execute tasks", or
  references an approved feature they want to implement. Requires approved tasks; execution
  is gated, not free-running.
---

# sf-build

Plan and execute implementation from approved specs. Human gate after each wave.

## Invocation Modes

### `sf-build` (no argument)
List features ready to build:
- Features in status `approved` → "Ready to build"
- Features in status `building` → "Paused, can resume"
- No features in either → "Nothing to build. Run `sf-propose <name>` first."

Ask: "Which feature do you want to build?"

### `sf-build <name>`
Normal flow: plan + execute the named feature.

## Pre-flight

1. Run `sf status` — verify the feature status is `approved` or `building`, and check its `lane`.
2. Read the feature's `tasks.md` — this is the execution plan
3. Read the feature's `design.md` — this is the architectural guide
4. Read the feature's `requirements.md` — for traceability during implementation
5. If `specforge/context/project.md` and `specforge/context/conventions.md` exist, read them — follow conventions
6. Read `build.mode` from `constitution.json` (`inline` | `single` | `per-wave`, default `inline`) — it decides how Step 2 executes (see "Execution strategy" below)
7. Check `build.test_cmd` in `constitution.json` — the deterministic command `sf check run` uses to run the suite (e.g. `"pytest -q"`, `"go test ./..."`, `"npm test"`). If it's missing, set it now via `sf save constitution --json -` (the verdict gate later requires a fresh green run, so `sf check run` must be able to execute the tests). Also set `build.report` when the stack supports it — `"go-json"` (test_cmd runs `go test -json ./...`) or `"junit"` (test_cmd carries a `{report}` placeholder, e.g. `"pytest -q --junitxml={report}"`): with a structured report the verdict additionally requires every test named in the trace to have run and passed (test→requirement causality)

If status is not `approved` or `building`: "Feature `<name>` is in status `<status>`. Run `sf-propose <name>` first."

**Lite lane.** If the feature's `feature.json` has `"lane": "lite"` (see `sf status`), the feature has a single
`change.md` instead of requirements/design/tasks. Build straight from it
(usually one wave) per `sf-propose/references/lite-lane.md`. If the change proves
bigger or touches a sensitive surface, **promote to standard** — stop, set
`"lane": "standard"`, and route back through `sf-propose`. The escape hatch only
goes up.

## Step 1: Compute the Execution Plan

The wave layout is **computed** from the task dependency graph — you don't sort
it by hand. Read `references/task-planning.md` for the details.

```
sf plan compute --feature=<name>
```

This validates the graph (no cycles, no dangling deps) and writes
`progress/plan.json` (+ `plan.md`) with each wave's tasks. If it reports a cycle
or dangling dep, fix the `depends_on` edges in `tasks.json` first.

**Optionally refine** (judgment, whole-plan, before the gate): add wave
`name`/`complexity`/`rationale`, or merge/split waves for review granularity —
but mostly in `inline` mode (in subagent modes, prefer the raw layout; see
task-planning.md). After any edit, `sf plan validate --feature=<name>` enforces
the dependency guard.

**Plan gate (fused into tasks, D2').** If tasks is already approved and you did
NOT refine the plan, `sf plan compute` auto-seals the plan gate — the plan is a
deterministic function of the approved tasks, so a separate human approval adds
fatigue without adding judgment. Just show the wave layout to the user and move
on. **If you DID refine** (merge/split waves, edited names), the manual edit
makes the sealed hash stale → present the refined plan and get explicit
approval: `sf gate approve --feature=<name> --phase=plan`.

→ 🔴 **GATE** (only when refined): Present the execution plan. Wait for approval.
- "Approved" → proceed per the execution strategy below
- Changes requested → adjust `depends_on`/plan, re-compute, re-present

The plan gate is also the **spawn authorization**: when `build.mode` is `single`
or `per-wave`, the sealed plan gate (auto or manual) is what authorizes launching
the build subagent(s).

## Execution strategy (`build.mode`)

Choose how to execute the approved waves based on `build.mode` (default `inline`):

- **`inline`** (default) — execute waves in THIS context, human gate after each
  wave. This is **Step 2** below; classic sf-build, unchanged.
- **`single`** — run the WHOLE build in one fresh subagent. Keeps the main
  context clean (build is where most tokens burn). Recommended once you trust the
  design + tasks to be a complete handoff.
- **`per-wave`** — run EACH wave in its own fresh subagent, sequentially. For
  large features (**> 3-4 waves**) or high-risk designs: each wave starts fresh,
  so context degradation never accumulates across waves. If `mode` is `single`
  but the plan has > 4 waves, suggest `per-wave` to the user.

> **CLI-driven alternative (`sf run`).** If the constitution declares
> `build.agent_cmd`, the whole per-wave loop can be driven by the CLI instead of
> this conversation: `sf run --feature=<name>` computes each wave's seed, spawns
> a fresh agent process (seed via stdin), verifies the contract + runs the suite,
> seals the `wave-N` checkpoint, and moves on — stopping on any failure. The
> process no longer depends on conversational memory at all. Gated phases
> (propose, verdict) stay conversational.

**Sequencing — the wave layout is fixed before sharding.** `sf plan compute` and
any plan refinement (Step 1) happen **before** the plan gate; the execution
subagents come **after** it and run their assigned wave. They do **not** re-layer
the plan — refinement is a whole-plan, pre-gate concern (it can't be per-wave; a
wave subagent only sees its own wave). And note: in subagent modes, prefer the
**raw computed layout** — splitting waves for review granularity only throws away
parallelism, and there's no per-wave human to fatigue (see `references/task-planning.md`).

> **Evolution — per-task within a wave (swarm).** Because `depends_on` makes
> parallelism explicit (same wave = independent tasks, guaranteed), a wave can be
> sharded further: one subagent **per task** within the wave, the wave acting as a
> sync barrier (all tasks done → checkpoint → next wave). `per-wave` is the
> conservative step; `per-task-within-wave` is the swarm the dependency graph now
> makes safe. Not built yet — tracked as the next granularity.

### Why a subagent (single / per-wave)

Freshness is the point: a build subagent starts at ~0% context, escaping the
degradation that hits a long main session. The main agent stays a thin conductor.

- **Seed (what the subagent gets).** For `single`: design + tasks + conventions +
  project context. For `per-wave`, the slice for wave N — deterministic:
  ```
  sf context for-wave --feature=<name> --n=<N>
  ```
  It already carries the wave's tasks, the requirement/component refs in scope,
  the files, and a summary of prior waves (the handoff). The CODE is NOT in the
  seed — the subagent has tools and reads the repo; it only needs to know where.

- **Handoff (how waves communicate).** Through ARTIFACTS, never the conversation.
  Each wave subagent marks tasks `[x]` in `tasks.md`, writes `progress/wave-N.md`,
  and leaves code on disk. Wave N+1 is seeded with `sf context for-wave --n=N+1`,
  which reports waves `0..N` as done. No subagent stays alive to carry state — the
  state lives in `tasks.json` + git.

- **Return (what comes back).** A THIN summary only — tasks completed, test
  results, deviations. NOT the subagent's transcript (that would refill the
  context we're keeping clean). The detail stays in `progress/wave-N.md` + git.

### Inter-wave checkpoint (`per-wave`)

Between waves run an AUTOMATED checkpoint — do NOT ask the human every wave (that
trains rubber-stamping):
1. **Contract**: `sf trace verify --contract --feature=<name> --wave=<n>` for the
   wave just finished — every requirement it touched must name a test that exists.
2. Run the project's tests for the work done so far.
3. If `trace.json` exists, run `sf doctor --drift` (checks every trace, the
   wave's feature included).

All green → launch the next wave's subagent automatically. Any failure → STOP and
escalate to the human with the failing detail. A per-wave human gate is
**optional** (max control); the default is automated + escalate-on-failure.

### Escalation (inside a subagent)

- Small gap the design didn't anticipate → record a deviation in
  `progress/wave-N.md` (and `sf journal add` later), then continue.
- Gap that INVALIDATES the design → stop and escalate to the main/human; do not
  improvise an architectural decision alone.

**Under `sf run`, say it with your exit code.** The orchestrator routes on the
code, so prose about being blocked changes nothing:

| Code | Meaning | What the orchestrator does |
|---|---|---|
| `0` | DONE | Runs the checkpoint, seals, next wave |
| `10` | DONE_WITH_CONCERNS | Runs the checkpoint anyway — finishing honestly isn't punished. If it passes, seals and continues with the concern surfaced |
| `11` | NEEDS_CONTEXT | Not a failure. Relaunches the wave **once**; twice in a row escalates, because the seed clearly can't supply it |
| `12` | BLOCKED | Stops and escalates |

Any other non-zero code reads as a crash, not a status — the range is high on
purpose so a harness that dies on its own stays distinguishable.

The code carries the *decision*; the *reason* goes in `progress/wave-N.md`,
which you're writing anyway. Narrow channel for the routing, artifact for the
evidence.

### Observability

The main agent loses the play-by-play. Mitigate it like the phase judge: the
subagent logs its decisions to `progress/wave-N.md` (and `trace.json` at check
time) as it goes — visibility deferred to the artifact, not the conversation.

For `single`/`per-wave`, the subagent(s) still follow Step 2's per-task and
failure protocols internally; when they finish, resume at **Step 4: Completion**
with the thin summary, recording each `wave-N` gate in `gates[]`.

## Step 2: Execute Wave by Wave (`inline` mode)

Read `references/wave-execution.md` for detailed execution strategy.

For each wave:

### 2a. Execute Tasks

For each task in the wave:
1. Implement the code changes
2. **Verify each change landed** — if a file edit fails (old text not found,
   wrong file, merge conflict), STOP. Report the failure to the user immediately.
   Do not proceed to the next task. Do not work around the failure.
3. Mark the task as `[x]` in `tasks.md` ONLY after verified implementation
4. Note what was done in the wave progress log
5. **Declare verification coverage (the contract)** — for each requirement this
   task implements, add/update its entry in `specforge/features/<name>/trace.json`
   with the code anchor(s) (`path:symbol`) and the **exact test** that proves it
   (`path::test_name` for pytest, `path:TestName` for Go). You must name a *real*
   test for every requirement you build — don't defer it to sf-check. This is the
   build-time half of traceability; sf-check later audits and seals it.

**⛔ FAILURE PROTOCOL**: If any operation fails during a wave:
- Log the failure in the wave progress with exact error details
- Report to the user: "Task T[n] failed: [error]. Wave [n] cannot continue."
- Wait for user decision: fix and retry, or restructure the wave
- A task with a failed operation stays `[ ]` — never mark incomplete work as done

### 2b. Log Progress

Create/update `specforge/features/<name>/progress/wave-<n>.md`:
using `templates/progress.tmpl.md`:

```markdown
# Wave <n> — [theme]

## Completed Tasks
- [x] T1: [what was done, files created/modified]
- [x] T2: [what was done, files created/modified]

## Decisions Made
- [any implementation decisions not in design.md]

## Issues Encountered
- [any problems, deviations from plan]

## Operations Audit
- File edits attempted: [count] | Succeeded: [count] | Failed: [count]
- Files created: [list]
- Tests run: [pass/fail/skip counts]
- **Wave integrity: [CLEAN — all operations succeeded | DIRTY — see failures above]**
```

**⛔ A wave with integrity DIRTY cannot pass the gate.**

### 2c. Verify the contract (gate precondition)

Before presenting the gate, the verification contract must hold for this wave:

```
sf trace verify --contract --feature=<name> --wave=<n>
```

This checks that every requirement the wave's tasks touch is declared in
`trace.json` with a test that **exists** in the code (it checks existence, not
that the test passes — passing is drift, checked separately). Exit non-zero → the
wave is **not** done: some requirement has no test named, or names a test that
isn't there. Fix it (write the missing test, correct the anchor in trace.json)
and re-run before the gate. This is what turns "test pass ≠ done" from a soft
rule into a checkable artifact — **"no test named ≠ done"** is now enforced.

### 2d. Gate

→ 🔴 **GATE**: Present wave results to the user.
- "Approved" → proceed to next wave
- "Fix X" → address issue, re-present wave
- "Stop" → pause: `sf feature set-status --feature=<name> --to=building` (never edit features.json by hand)

## Step 3: Handle Failures

When a task fails and can't be resolved within the wave:

### Log the failure

Create or update `specforge/features/<name>/failures.md`:

```markdown
# <Feature Name> — Failures

## F1: [failure title]
**Date**: [date]
**Task**: T[n] (Wave [n])
**Requirement**: R[n]

### What failed
[Specific behavior that didn't work]

### Why it failed
[Root cause — spec issue, design mismatch, technical limitation, or unknown]

### What was attempted
1. [Fix attempt 1 — what and result]
2. [Fix attempt 2 — what and result]

### Impact
- [What downstream tasks are blocked]
- [What requirements are affected]

### Recommended action
- [ ] [Specific action: amend spec / change design / investigate further]
```

Failures that are fixed within the wave (minor issues) go in the wave progress
log, not in failures.md. Only persistent failures that affect the build's
completeness get their own entry.

→ 🔴 **GATE**: Present failure to user. User decides:
- Fix and retry
- Skip and continue (task stays `[ ]`)
- Pause build
- Return to propose to amend specs

## Step 4: Completion

Record each gate as it passes (F10) with `sf gate approve` — **don't hand-write
the entries.** When the plan gate is approved:

```bash
sf gate approve --feature=<name> --phase=plan
```

This seals a content hash of `progress/plan.json` (the CLI computes it over the
real file, never the LLM), so the stale model can later detect if the plan was
edited after approval.

Then for each wave gate as it passes:

```bash
sf gate approve --feature=<name> --phase=wave-0
sf gate approve --feature=<name> --phase=wave-1
```

Wave gates carry **no hash** — a wave is an execution checkpoint, not a spec
artifact. Its spec artifact is `plan.json` (sealed above), and the built code is
governed by drift detection (`sf doctor --drift` / `trace.json`), not by a
content hash. The ledger is the auditable record of what was actually approved.

**Team mode (F36, optional).** If the project follows the git convention, each
passing wave becomes one commit (`feat(<slug>): wave N — <tasks done>`) on the
`feature/<slug>` branch — the git history is the build op-log. See
`skills/sfx-github/references/specforge-integration.md`. Solo projects can ignore
this.

After all waves complete:

1. **Self-audit**: Verify every task in `tasks.md` is marked `[x]`. If any task
   is still `[ ]`, the build is NOT complete — report the gap.
2. `sf feature set-status --feature=<name> --to=checking` (the CLI is the only writer of features.json)
3. Append to `specforge/history.md`:
   ```
   ## [date] — Feature built: <name>
   - Waves: [count] | Tasks: [completed]/[total]
   - Failures: [count unresolved, or "none"]
   ```
4. If `failures.md` exists with unresolved entries:
   ```
   ⚠ Feature built with [N] unresolved failures. See failures.md.
   sf-check will flag these in the gap analysis.
   ```
5. If unresolved failures exist:
   Inform: "Feature <name> built with [N] unresolved failures. Proceeding to
   sf-check — it will flag these in the gap analysis."
   Proceed to sf-check anyway. The failures will surface in the review.

## Resuming a Paused Build

If a feature has status `building`:
1. Read `progress/` to find last completed wave
2. Read `failures.md` if it exists — show unresolved failures
3. Show the user what's done and what remains
4. Resume from the next incomplete wave
## Rationalizations

The failure and escalation protocols above are imperatives. These are the
specific ways an agent talks itself out of them.

| The excuse | Why it doesn't hold |
|---|---|
| "This wave is small, I'll record it at the end" | This is the FIXBUGHIGH failure verbatim: state written from memory instead of as it happens. By the end you're reconstructing, and reconstruction is where invented anchors come from. Record per wave, via `sf save`. |
| "I'll complete the trace once everything works" | The trace is the checkpoint's input, not its output. A wave with no trace can't pass `sf trace verify --contract`, so "later" means the checkpoint never ran. |
| "The deviation is small, no need to write it down" | Small deviations are exactly the ones nobody remembers deciding. Write it in `progress/wave-N.md`; it costs a line and it's the only record that survives the subagent. |
| "I can decide this architectural point myself" | If it invalidates the design, it's not yours to decide alone — that's the escalation rule. Improvising here is how a build ends up correct against a design nobody approved. |
| "The test runner is flaky, I'll mark it green" | A flaky test is a finding, not a rounding error. Report it as a failure with the evidence; a green you asserted rather than observed is the one thing the whole verification tier exists to prevent. |
