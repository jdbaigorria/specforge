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

1. Read `specforge/features.json` — verify the feature status is `approved` or `building`
2. Read the feature's `tasks.md` — this is the execution plan
3. Read the feature's `design.md` — this is the architectural guide
4. Read the feature's `requirements.md` — for traceability during implementation
5. If `specforge/context/project.md` and `specforge/context/conventions.md` exist, read them — follow conventions

If status is not `approved` or `building`: "Feature `<name>` is in status `<status>`. Run `sf-propose <name>` first."

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
- "Approved" → proceed to Wave 0
- Changes requested → adjust plan, re-present

## Step 2: Execute Wave by Wave

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