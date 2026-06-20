# Task Planning

## Purpose

Group the tasks into executable waves based on their dependencies. This is the
first step of `sf-build`, before implementation.

## The wave layout is COMPUTED, not hand-rolled

Tasks (`tasks.json`) are a **flat** list; each task declares `depends_on` (task
IDs). The wave grouping is a deterministic **topological layering** of that
graph — so `sf` computes it, you don't sort it by hand:

```
sf plan compute --feature=<name>
```

This reads `tasks.json`, validates the graph (no dangling deps, no cycles), and
writes `progress/plan.json` (+ `plan.md`) with each wave's task IDs:

- **Wave 0** = tasks with no dependencies.
- **Wave N** = `1 + max(wave of its dependencies)` (as early as possible).
- Tasks in the **same wave** are independent → parallelizable.

If `sf plan compute` reports a cycle or a dangling dependency, the spec is wrong
— fix the `depends_on` edges in `tasks.json`, don't paper over it in the plan.

## Refining the computed plan (judgment, optional)

The computed layout is the deterministic default. You MAY refine it — add a wave
`name`/`complexity`/`rationale`, or merge/split waves for review granularity
(guideline below). After any manual edit, run `sf plan validate --feature=<name>`:
it **errors** if a task lands in a wave at or before one of its dependencies (the
dependency guard) or if a task is left unassigned — so refinement can't silently
break the ordering. Re-running `sf plan compute` preserves your
`name`/`complexity`/`rationale` by wave number.

## Wave Size Guidelines (for refinement)

- Each wave should have 1-5 tasks
- If a wave has more than 5 tasks, split into sub-waves
- If a wave has only 1 task and it's trivial, merge with adjacent wave
- Total waves should be 2-5 for most features
  (1 wave = feature is too simple for SDD; 6+ waves = feature should be split)

## Scope Estimation

For each wave, estimate:
- **Files touched:** list of files created or modified
- **Complexity:** low (boilerplate/config), medium (logic), high (algorithm/integration)
- **Risk:** where things might go wrong

This helps the human decide if the plan is reasonable before execution starts.

## Traceability Validation

Before presenting the plan, verify:
- Every task maps to at least one requirement (from `tasks.md` traceability table)
- Every requirement is covered by at least one task
- If gaps exist, flag them: "R3 has no implementing task — add one?"
