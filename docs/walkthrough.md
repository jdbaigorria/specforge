[← Back to README](../README.md)

# Walkthrough

Two end-to-end runs of the pipeline.

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
