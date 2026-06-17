---
name: sf-propose
description: >
  Generate feature specifications: requirements, design, and tasks. Use when the user says
  "sf-propose", "propose feature", "specify", "spec this", "add feature", or describes
  functionality they want to build. Also triggers on "sf-propose --design-first" for
  architecture-driven specification, "sf-propose --from-code" for reverse-engineering specs
  from existing code, "sf-propose --all" to decompose the entire product into a roadmap and
  propose each feature, or when the user wants to refine/clarify existing specs. Covers the
  full specification lifecycle from idea to implementation plan.
---

# sf-propose

Generate feature specifications. Produces 3 artefacts with a human gate after each one.

## Invocation Modes

### `sf-propose` (no argument)
Ask the user: "What feature do you want to specify?" Then proceed as `sf-propose <name>`.

### `sf-propose <name>`
Normal flow: specify a single feature (requirements → design → tasks).

### `sf-propose --all`
Decompose the entire product into features based on the constitution. Generate a roadmap.
Then propose each feature one by one with gates. See [Product Decomposition](#--all-product-decomposition) below.

### `sf-propose --design-first <name>`
Architecture-first specification. Read `references/design-first.md` before proceeding.

### `sf-propose --from-code`
Reverse-engineer specs from existing code. Read `references/from-code.md` before proceeding.

## Pre-flight

1. Verify `specforge/` exists. If not: "Run `sf-init` first."
2. Read `specforge/features.json` to check for naming conflicts.
3. If `specforge/constitution.md` exists, read it — principles guide spec generation.
4. If `specforge/context/project.md` exists, read it — stack context informs design.

## --all: Product Decomposition

Read `specforge/constitution.md` — identity, principles, constraints, anti-goals.

### Step A1: Decompose into features

Based on the product identity, break it down into logical features:
- Group by user-facing capability (not by technical layer)
- Each feature should be independently deliverable
- Name in kebab-case: `user-auth`, `task-crud`, `export-csv`

### Step A2: Prioritize

Classify each feature:

| Priority | Meaning |
|----------|---------|
| **Must** | Product doesn't work without it. MVP. |
| **Should** | Important but can launch without. v1.1. |
| **Could** | Nice to have. Backlog. |

### Step A3: Generate roadmap

Create `specforge/roadmap.md`:

```markdown
# Product Roadmap

Generated from constitution on [date].

## Must (MVP)
1. [feature-name] — [one-line description]
2. [feature-name] — [one-line description]

## Should (v1.1)
3. [feature-name] — [one-line description]

## Could (backlog)
4. [feature-name] — [one-line description]

## Dependencies
[feature-B depends on feature-A because...]
```

**Trace to product requirements (F27).** If the project was seeded from an
`sfp-scout` discovery brief, derive each roadmap feature from the product
requirements: tag it with which `PR#` it serves, e.g. `1. auth — login/logout
(serves: PR1, PR3)`. Record `"serves": ["PR1","PR3"]` on the feature in
`features.json`. This extends the traceability spine upward —
`PR# → feature → R# → task → code → test` — so product intent stays linked to
shipped code (no telephone game). Features with no discovery origin (brownfield)
just omit `serves`.

→ 🔴 **GATE**: Present roadmap. User approves, reorders, or adjusts.

### Step A4: Propose each feature

After roadmap approval, start proposing features in priority order.
For each: run the normal Requirements-First flow below.
Gate after each feature's 3 artefacts. User can stop at any point.

---

## Step 0: Lane Triage (lite vs standard)

Before specifying, classify the change and **propose a lane**. The framework
classifies; the human confirms (F34). This is the only place a fast lane is
chosen — never let momentum silently skip ceremony.

Estimate from the user's description:

- **Surface:** touches one file/string vs many?
- **Sensitivity:** any security/auth/data/money/migration surface? (if yes → standard)
- **Reversibility:** trivially revertible vs not?

Propose a lane:

- **Lite** — typos, copy, config, a one-file change with no sensitive surface.
  One combined artefact, one spec gate, build, minimal check.
- **Standard** — everything else. The full requirements → design → tasks flow.

```
───────────────────────────────────────
🔴 GATE — lane: proposed **{lite|standard}** for "<feature>"
Reason: <1 file / no sensitive surface / reversible — or why standard>
Awaiting approval. Reply: approve / use {other lane} / change X
───────────────────────────────────────
```

Record the approved lane in `features.json` (`"lane": "lite"|"standard"`). It is
a gated, auditable decision — not the agent deciding to go fast.

- **Lite approved** → read `references/lite-lane.md` and follow it. Stop here.
- **Standard approved** → continue with the Requirements-First Flow below.

When the lane is obvious and the user invoked plainly, you may still propose —
but always through this gate. Default to **standard** whenever unsure.

## Requirements-First Flow

### Step 1: Understand the Feature

Engage the user in conversation:
- **What** does this feature do? (user-facing behavior)
- **Why** does it exist? (problem it solves)
- **Who** uses it? (actor/persona)
- **What are the boundaries?** (what it deliberately doesn't do)

Keep the conversation focused. 3-5 exchanges should be enough for most features.
Don't interrogate — if the user gives a clear description, proceed.

### Step 2: Generate requirements.md

Create `specforge/features/<name>/requirements.md`.

Use EARS notation for system behaviors. Read `references/ears-notation.md` if needed.

Structure:
```markdown
# <Feature Name> — Requirements

## Overview
[1-2 sentence summary]

## User Stories
- As a [actor], I want to [action], so that [benefit]

## System Requirements

### R1: [Requirement name]
WHEN [trigger]
THE SYSTEM SHALL [behavior]

**Acceptance Criteria:**
- [ ] AC1: [testable criterion]
- [ ] AC2: [testable criterion]

### R2: ...

## Edge Cases
- [What happens when X fails?]
- [What happens with empty input?]

## Out of Scope
- [What this feature deliberately doesn't do]
```

→ 🔴 **GATE**: Present `requirements.md` to the user.
- "Approved" → proceed to Step 3
- Changes requested → iterate on requirements, re-present
- If user edits the file directly → acknowledge changes, proceed

### Step 3: Generate design.md

Create `specforge/features/<name>/design.md`.

If a technical decision requires investigation, read `references/research.md`.

Only include sections relevant to the feature's complexity. A CLI flag addition
doesn't need Security Considerations. A payment system does. Include a section
only if it changes a decision or informs implementation. Sections that would say
"N/A" generate noise — omit them.

Core sections (always include):
- Architecture (how it fits into the system)
- Components (what gets created/modified)
- Technical Decisions (if any non-obvious choices were made)

Conditional sections (include when relevant):
- Data Model (if entities or schemas are involved)
- API / Interface Contracts (if public interfaces change)
- Sequence / Flow (if the interaction has multiple steps)
- Error Handling (if failure modes are non-trivial)
- Testing Strategy (if testing approach differs from project default)
- Security Considerations (if security-sensitive surfaces exist)
- Performance Considerations (if performance-sensitive paths exist)

Structure:
```markdown
# <Feature Name> — Design

## Architecture
[How this feature fits into the existing system]
[Component diagram or description]

## Components
### [Component 1]
- **Responsibility:** [what it does]
- **Interface:** [inputs/outputs]
- **Dependencies:** [what it needs]

### [Component 2]
...

## Data Model
[Entities, schemas, or state shape]

## API / Interface Contracts
[Endpoints, CLI commands, function signatures]

## Technical Decisions
| Decision | Choice | Rationale |
|----------|--------|-----------|
| [e.g., storage] | [e.g., DynamoDB] | [why] |

## Error Handling
[Strategy for failures, retries, user-facing errors]

## Testing Strategy
[Unit, integration, e2e — what gets tested and how]
```

→ 🔴 **GATE**: Present `design.md` to the user.
- "Approved" → proceed to Step 4
- Changes requested → iterate, re-present
- If design changes invalidate requirements → flag and offer to update requirements

### Step 4: Generate tasks.md

Create `specforge/features/<name>/tasks.md`.

Structure:
```markdown
# <Feature Name> — Tasks

## Implementation Plan

### Wave 0: [theme] (no dependencies)
- [ ] T1: [task description] → R1
  - Expected: [what exists after this task]
  - Files: [files created/modified]

- [ ] T2: [task description] → R1, R2
  - Expected: [what exists after this task]
  - Files: [files created/modified]

### Wave 1: [theme] (depends on Wave 0)
- [ ] T3: [task description] → R3
  - Expected: ...
  - Files: ...

### Wave 2: [theme] (depends on Wave 1)
- [ ] T4: [task description] → R2, R4
  - Expected: ...
  - Files: ...

## Traceability
| Requirement | Tasks |
|-------------|-------|
| R1          | T1, T2 |
| R2          | T2, T4 |
| R3          | T3     |
| R4          | T4     |
```

Every task must trace to at least one requirement.
Every requirement must be covered by at least one task.
If any requirement has no task, add one or flag it to the user.

→ 🔴 **GATE**: Present `tasks.md` to the user.
- "Approved" → update `features.json` status to `approved`
- Changes requested → iterate, re-present

### Step 5: Update Feature Registry

Update `specforge/features.json`. The registry is the **source of truth** for
status and gates (F10). Each gate the user approved in this run is appended to
the `gates[]` ledger — this is the auditable record that a gate actually
happened, not a claim in a markdown header.

```json
{
  "name": "<feature-name>",
  "status": "approved",
  "workflow": "requirements-first",
  "lane": "standard",
  "depends_on": [],
  "created": "<date>",
  "completed": null,
  "gates": [
    { "phase": "lane",         "result": "approve", "by": "user", "at": "<iso-8601>", "comment": "standard" },
    { "phase": "requirements", "result": "approve", "by": "user", "at": "<iso-8601>", "comment": null },
    { "phase": "design",       "result": "approve", "by": "user", "at": "<iso-8601>", "comment": null },
    { "phase": "tasks",        "result": "approve", "by": "user", "at": "<iso-8601>", "comment": null }
  ]
}

For a **lite** feature, `"lane": "lite"` and the gates are just `lane` → `change`
→ `plan`/`wave` → `verdict` (no separate requirements/design/tasks gates).

`depends_on` lists the feature names this one needs first (captured from the
roadmap during `--all`, or stated by the user). It drives the critical-path
ordering and blocker view in `/sf-status` (F37). Default `[]`.
```

Record the real result of each gate: `approve`, `reject`, or `change` (with the
request in `comment`). Never write a gate entry the user did not actually give.
If a downstream gate reopened (F23), mark the affected feature `status` back and
the stale gate is re-recorded on re-approval.

Append to `specforge/history.md`:
```
## [date] — Feature proposed: <name>
- Workflow: requirements-first
- Requirements: [count] | Tasks: [count] | Waves: [count]
```

Inform the user: "Feature `<name>` approved. Use `sf-build <name>` to start implementation."

## Post-Generation: Clarify (optional)

If the user asks to refine, review, or clarify specs after generation:
read `references/clarify.md` for the refinement procedure.

## Resync Detection

If invoked on a feature that already has artefacts:
1. Check which artefacts exist
2. Compare timestamps
3. If upstream is newer than downstream → offer to regenerate
4. If user modified an artefact directly → acknowledge and cascade