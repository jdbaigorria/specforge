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

## Pre-flight

1. Read `specforge/features.json` — verify the feature status is `approved`
2. Read the feature's `tasks.md` — this is the execution plan
3. Read the feature's `design.md` — this is the architectural guide
4. Read the feature's `requirements.md` — for traceability during implementation
5. If `.ai/project.md` and `.ai/conventions.md` exist, read them — follow conventions

If status is not `approved`: "Feature `<name>` is in status `<status>`. Run `sf-propose <name>` first."

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

Create/update `specforge/features/<name>/progress/wave-<n>.md`
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

## Step 3: Completion

After all waves complete:

1. **Self-audit**: Verify every task in `tasks.md` is marked `[x]`. If any task
   is still `[ ]`, the build is NOT complete — report the gap.
2. Update `features.json` status to `checking`
3. Append to `specforge/history.md`:
   ```
   ## [date] — Feature built: <name>
   - Waves: [count] | Tasks: [completed]/[total]
   ```
4. **Proceed directly to sf-check.** Do not ask. Do not wait. Load the sf-check
   skill and execute it. The build→check transition is automatic.

## Resuming a Paused Build

If a feature has status `building`:
1. Read `progress/` to find last completed wave
2. Show the user what's done and what remains
3. Resume from the next incomplete wave