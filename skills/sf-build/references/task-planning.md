# Task Planning

## Purpose

Organize tasks from `tasks.md` into executable waves based on dependencies.
This is the first step of `sf-build`, always executed before implementation.

## Wave Organization Rules

### Wave 0: Foundation
Tasks with no dependencies on other tasks.
Typically: data models, schemas, configuration, base utilities.

### Wave N: Depends on Wave N-1
Tasks that require output from the previous wave.
Order by: data layer → logic layer → interface layer → integration.

### Parallel Tasks Within a Wave
Tasks in the same wave can be executed in any order.
If task A and task B are both in Wave 1 and don't depend on each other,
they can be implemented in parallel (or any order).

## Dependency Detection

For each task, check:
1. Does it reference entities/schemas from another task? → depends on that task
2. Does it import/use functions defined in another task? → depends on that task
3. Does it test behavior built in another task? → same wave or later

## Wave Size Guidelines

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
