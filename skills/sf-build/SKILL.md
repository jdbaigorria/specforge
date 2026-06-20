---
name: sf-build
description: >
  Plan and execute feature implementation from approved specs. Use when the user says
  "sf-build", "build feature", "implement", "start building", "execute tasks", or
  references an approved feature they want to implement. Reads the approved tasks.md,
  organizes execution by waves, and implements with a human gate after each wave.
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

1. Read `specforge/features.json` — verify the feature status is `approved` or `building`, and check its `lane`.
2. Read the feature's `tasks.md` — this is the execution plan
3. Read the feature's `design.md` — this is the architectural guide
4. Read the feature's `requirements.md` — for traceability during implementation
5. If `specforge/context/project.md` and `specforge/context/conventions.md` exist, read them — follow conventions
6. Read `build.mode` from `constitution.json` (`inline` | `single` | `per-wave`, default `inline`) — it decides how Step 2 executes (see "Execution strategy" below)

If status is not `approved` or `building`: "Feature `<name>` is in status `<status>`. Run `sf-propose <name>` first."

**Lite lane.** If `features.json` has `"lane": "lite"`, the feature has a single
`change.md` instead of requirements/design/tasks. Build straight from it
(usually one wave) per `sf-propose/references/lite-lane.md`. If the change proves
bigger or touches a sensitive surface, **promote to standard** — stop, set
`"lane": "standard"`, and route back through `sf-propose`. The escape hatch only
goes up.

## Step 1: Generate Execution Plan

Read `references/task-planning.md` for detailed planning instructions.

**Quick summary:** Organize tasks from `tasks.md` into execution waves:

```markdown
# <Feature Name> — Execution Plan

## Wave 0: [theme] (no dependencies)
- T1: [description]
- T2: [description]
Estimated scope: [files touched, rough complexity]

## Wave 1: [theme] (depends on Wave 0)
- T3: [description]
Estimated scope: [files touched, rough complexity]

## Wave 2: [theme] (depends on Wave 1)
- T4: [description]
Estimated scope: [files touched, rough complexity]
```

Save as `specforge/features/<name>/progress/plan.md`.

→ 🔴 **GATE**: Present the execution plan. Wait for approval.
- "Approved" → proceed per the execution strategy below
- Changes requested → adjust plan, re-present

The plan gate is also the **spawn authorization**: when `build.mode` is `single`
or `per-wave`, approving the plan is what authorizes launching the build
subagent(s). The human is present at this gate, so the spawn is safe.

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
1. Run the project's tests for the work done so far.
2. If `trace.json` exists, run `sf doctor --drift --feature=<name>`.

All green → launch the next wave's subagent automatically. Any failure → STOP and
escalate to the human with the failing detail. A per-wave human gate is
**optional** (max control); the default is automated + escalate-on-failure.

### Escalation (inside a subagent)

- Small gap the design didn't anticipate → record a deviation in
  `progress/wave-N.md` (and `sf journal add` later), then continue.
- Gap that INVALIDATES the design → stop and escalate to the main/human; do not
  improvise an architectural decision alone.

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

### 2c. Gate

→ 🔴 **GATE**: Present wave results to the user.
- "Approved" → proceed to next wave
- "Fix X" → address issue, re-present wave
- "Stop" → pause, update status to `building` in features.json

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

Record each gate as it passes (F10): append a `{ "phase": "plan", ... }` entry
to the feature's `gates[]` ledger when the plan gate is approved, and a
`{ "phase": "wave-N", ... }` entry for each wave gate. The ledger is the
auditable record of what was actually approved.

**Team mode (F36, optional).** If the project follows the git convention, each
passing wave becomes one commit (`feat(<slug>): wave N — <tasks done>`) on the
`feature/<slug>` branch — the git history is the build op-log. See
`skills/sfx-github/references/specforge-integration.md`. Solo projects can ignore
this.

After all waves complete:

1. **Self-audit**: Verify every task in `tasks.md` is marked `[x]`. If any task
   is still `[ ]`, the build is NOT complete — report the gap.
2. Update `features.json` status to `checking`
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