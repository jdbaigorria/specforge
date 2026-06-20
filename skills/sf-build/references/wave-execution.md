# Wave Execution Strategy

## Execution Loop

For each wave, follow this cycle:

### 1. Pre-wave Check

Before starting a wave:
- Verify all tasks from previous wave are marked `[x]` in `tasks.md`
- Verify previous wave's tests pass (if applicable)
- Read the wave's tasks and understand the scope

### 2. Task Execution Order

Within a wave, tasks have no dependencies on each other (that's why they're
in the same wave). Execute in the order that feels most natural:

**Recommended order within a wave:**
1. Data models / schemas first (other tasks may reference them)
2. Core logic / business rules
3. Interface / API / CLI layer
4. Tests last (they verify everything above)

### 3. Per-Task Procedure

For each task:

1. **Read** the task description, expected output, and file list from `tasks.md`
2. **Read** the relevant requirement(s) from `requirements.md`
3. **Read** the relevant component(s) from `design.md`
4. **Climb the decision ladder** (parsimony — see below) BEFORE writing any code
5. **Implement** the MINIMAL code that satisfies the task, following
   `specforge/context/conventions.md` if it exists
6. **Write tests** as specified in the task
7. **Mark** the task as `[x]` in `tasks.md`
8. **Log** what was done — and, if you wrote new code, which ladder rung you
   landed on and why the rungs above didn't serve

### Decision ladder (minimal code)

Before writing a single line for a task, climb this ladder and stop at the first
rung that works — *necessary code, not golfed, not clever*:

```
Does this even need to exist?            → if no, don't write it
  └─ Does the stdlib already do it?      → use it
       └─ Is there a native language/framework feature?  → use it
            └─ Does an ALREADY-installed dependency do it?  → use it
                 └─ Is it a one-liner?   → write the one line
                      └─ only here: write the MINIMAL code
```

Lazy-senior-dev discipline: the best code is the code you didn't write. A new
dependency, a new abstraction, or a config knob is a rung you must *justify*, not
a default. If you land on "write code", note in the wave progress which rungs you
ruled out and why — that's the audit trail `sf-check` reviews (constitution
principle `P-min`).

> Adapted from the **ponytail** project (https://github.com/DietrichGebert/ponytail,
> MIT) — we adopt the *idea* (the ladder), not its code.

### 4. Post-wave Validation

After completing all tasks in a wave:
- Run the test suite (if applicable)
- Verify the expected output matches what was produced
- Write the wave progress log using `templates/progress.tmpl.md`
- Present to the human for approval

### 5. Error Recovery

If a task fails or produces unexpected results:

**Minor issue (typo, wrong import):**
- Fix inline, note in wave progress log

**Design mismatch (task can't be implemented as specified):**
- Stop the wave
- Document the issue in the progress log
- Present to human: "T3 can't be implemented as specified because [reason].
  Options: (a) modify design.md, (b) adjust the task, (c) skip and continue"
- Wait for human decision → 🔴 GATE

**Blocking dependency (previous wave's output is wrong):**
- Stop the wave
- Document what's wrong and what needs to change
- Human decides: fix previous wave or adjust current plan

## Resume After Pause

If a build was paused mid-wave:

1. Read `specforge/features/<name>/progress/` to find last completed wave
2. Read `tasks.md` to find which tasks are `[x]` and which are `[ ]`
3. Show the user a summary: "Wave 1 is partially complete. T3 done, T4 pending."
4. Resume from the next incomplete task

## Conventions

- **Never skip a task silently.** If a task seems unnecessary, flag it.
- **Never modify specs during build.** If specs need changes, pause and
  escalate to the human. Specs are the contract; build fulfills them.
- **Always log.** Even if a wave is trivial, write the progress file.
  It's the audit trail that `sf-check` uses.
