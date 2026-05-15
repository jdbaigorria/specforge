<p align="center">
  <img src="assets/specforge-logo.png" alt="SpecForge" width="380" />
</p>

# SpecForge

Spec-Driven Development framework. The specification is the product — code is a regenerable byproduct.

4 skills. Progressive disclosure. Human gate on every artefact. No ceremony without purpose.

---

## Table of contents

- [How it works](#how-it-works)
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

  4 skills. Each produces artefacts. Human gate 🔴 on every artefact.
```

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

## Architecture

### Directory structure

```
my-project/
├── specforge/
│   ├── features.json                # Feature registry (status tracker)
│   ├── constitution.md              # Project principles + identity
│   ├── history.md                   # Append-only project log
│   ├── features/
│   │   └── add-task-manager/
│   │       ├── requirements.md
│   │       ├── design.md
│   │       ├── tasks.md
│   │       ├── review.md
│   │       ├── decisions/           # Complex technical decisions (optional)
│   │       └── progress/
│   │           ├── plan.md
│   │           ├── wave-0.md
│   │           └── wave-1.md
│   └── archive/
│       └── 2026-05-12-add-auth/
│
├── .ai/
│   ├── project.md                   # Stack, architecture (brownfield inferred)
│   ├── conventions.md               # Code conventions
│   ├── compact-rules.md             # Condensed rules for sub-agents
│   └── session.md                   # Session continuity
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
│   ├── task-planning.md             # Organize tasks into waves
│   └── wave-execution.md            # Per-wave execution strategy
└── templates/
    └── progress.tmpl.md

sf-check/
├── SKILL.md
├── references/
│   ├── backprop.md                  # Pattern 3x → invariant promotion
│   └── archive.md                   # Sync + close procedure
└── templates/
    └── review.tmpl.md
```

### Execution model

All 4 skills run **inline** — in the main conversation context. No sub-agent
delegation for SpecForge skills because every artefact has a human gate that
requires interaction.

The anti-telephone-game principle: artefacts live on disk. When a skill needs
context from a previous artefact, it reads the file — it doesn't rely on
conversation history. This means context window carries references, not payloads.

---

## Skills reference

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
3. Generate `.ai/project.md` + `.ai/conventions.md` from constitution context → 🔴 GATE

**Brownfield flow:**
1. Scaffold directories (automatic)
2. Onboard: analyze codebase → `.ai/project.md` + `.ai/conventions.md` → 🔴 GATE
3. Constitution conversation: principles and anti-goals (shorter, stack already known) → 🔴 GATE

**Produces:**
- `specforge/constitution.md` — project identity, principles, constraints, anti-goals
- `specforge/features.json` — empty feature registry
- `specforge/history.md` — project log (first entry)
- `.ai/project.md` — stack and architecture context
- `.ai/conventions.md` — coding standards

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
4. Generate `tasks.md` with waves + traceability matrix → 🔴 GATE
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
- `specforge/features/<name>/tasks.md` — waves + traceability matrix

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
1. Generate execution plan from `tasks.md` waves → 🔴 GATE
2. For each wave:
   a. Execute all tasks in the wave
   b. Log progress to `progress/wave-<n>.md`
   c. Present results → 🔴 GATE
3. All waves complete → status becomes `checking`

**Error recovery (3 levels):**
- **Minor** (typo, wrong import): fix inline, note in progress log
- **Design mismatch** (can't implement as specified): pause wave, present options to user → 🔴 GATE
- **Blocking dependency** (previous wave output wrong): pause, escalate to user

**Key rule:** Never modify specs during build. If specs need changes, pause and
escalate. Specs are the contract; build fulfills them.

**Produces per wave:**
- `specforge/features/<name>/progress/plan.md` — execution plan
- `specforge/features/<name>/progress/wave-<n>.md` — what was done, decisions, issues
- Updated `tasks.md` — tasks marked `[x]` as completed

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
3. Constitution compliance: validate against principles
4. Verdict: APPROVE / APPROVE WITH NOTES / REVISE
5. Present review → 🔴 GATE
6. If APPROVE → archive automatically

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

## Artefact flow

```
sf-init produces:
  specforge/constitution.md
  .ai/project.md
  .ai/conventions.md

sf-propose produces (per feature):
  specforge/features/<name>/requirements.md
  specforge/features/<name>/design.md
  specforge/features/<name>/tasks.md

sf-build produces (per feature):
  specforge/features/<name>/progress/plan.md
  specforge/features/<name>/progress/wave-<n>.md

sf-check produces (per feature):
  specforge/features/<name>/review.md
  specforge/archive/<date>-<name>/        (on APPROVE)
```

**Who reads what:**

| Artefact | Created by | Read by |
|----------|-----------|---------|
| constitution.md | sf-init | sf-propose (constraints), sf-check (compliance) |
| .ai/project.md | sf-init | sf-propose (stack context), sf-build (conventions) |
| .ai/conventions.md | sf-init | sf-build (coding standards) |
| requirements.md | sf-propose | sf-build (traceability), sf-check (validation) |
| design.md | sf-propose | sf-build (architecture guide), sf-check (adherence) |
| tasks.md | sf-propose | sf-build (execution), sf-check (traceability) |
| progress/*.md | sf-build | sf-check (audit trail) |
| review.md | sf-check | archive (verdict determines archival) |
| features.json | sf-init | all skills (status gate) |
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
  ✓ .ai/ created
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

**Why only 4 skills instead of 10?**
Progressive disclosure. Capabilities that were separate skills (clarify, research,
map, archive, explore, constitute) now live as references inside the 4 core skills.
They load on demand. Less context overhead, less cognitive load.

**Can I use SpecForge with any AI agent?**
Yes. Skills are markdown files. Any agent that reads markdown can execute them.
The AGENT.md orchestrator targets Claude Code but the skills are agent-agnostic.

**What if check keeps returning REVISE?**
REVISE sends you back to sf-build with specific corrections. If the spec itself
is wrong, edit it directly and resync detection will cascade the changes.

**Can multiple features be active simultaneously?**
Yes. Each feature has its own folder. `features.json` tracks status independently.

**How is this different from OpenSpec / Spec Kit / CaveKit?**
SpecForge combines: constitution + identity from Spec Kit, change organization
from OpenSpec, wave execution + backprop from CaveKit. The 4-skill architecture,
human gates on every artefact, EARS notation, resync detection, and progressive
disclosure are unique to SpecForge.

**What about the product owner / vision phase?**
It's built into sf-init. The constitution conversation captures identity (what,
who, why), principles, and anti-goals. This is the product definition phase —
it just doesn't need a separate skill.
