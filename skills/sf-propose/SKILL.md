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
Then propose each feature one by one with gates. Read `references/product-decomposition.md`.

### `sf-propose --design-first <name>`
Architecture-first specification. Read `references/design-first.md` before proceeding.

### `sf-propose --from-code`
Reverse-engineer specs from existing code. Read `references/from-code.md` before proceeding.

## Pre-flight

1. Verify `specforge/` exists. If not: "Run `sf-init` first."
2. Run `sf status` to check for naming conflicts.
3. If `specforge/constitution.md` exists, read it — principles guide spec generation.
4. If `specforge/context/project.md` exists, read it — stack context informs design.

## Step 0: Lane Triage (lite vs standard)

Before specifying, classify the change and **propose a lane** through a gate —
the framework classifies, the human confirms (F34). This is the only place a fast
lane is chosen; never let momentum silently skip ceremony. Default to **standard**
whenever unsure.

Read `references/lane-triage.md` for the triage heuristics and the gate format.

- **Lite approved** → read `references/lite-lane.md` and follow it. Stop here.
- **Standard approved** → continue with the Requirements-First Flow below.

## `--all`: Product Decomposition

For `sf-propose --all`, read `references/product-decomposition.md`: decompose the
product into features from the constitution, generate `roadmap.md`, gate it, then
propose each feature with the Requirements-First Flow below.

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

Create `specforge/features/<name>/requirements.md` using `templates/requirements.tmpl.md`.

Use EARS notation for system behaviors (`WHEN [trigger] THE SYSTEM SHALL
[behavior]`). Read `references/ears-notation.md` if needed. Every requirement
needs testable acceptance criteria, plus edge cases and an explicit out-of-scope
section.

→ 🔴 **GATE**: Present `requirements.md` to the user.
- "Approved" → proceed to Step 3
- Changes requested → iterate on requirements, re-present
- If user edits the file directly → acknowledge changes, proceed

### Step 3: Generate design.md

Create `specforge/features/<name>/design.md` using `templates/design.tmpl.md`.

If a technical decision requires investigation, read `references/research.md`.

**Only include sections relevant to the feature's complexity.** A CLI flag
addition doesn't need Security Considerations; a payment system does. Include a
section only if it changes a decision or informs implementation — sections that
would say "N/A" generate noise. Always include Architecture, Components, and
Technical Decisions; add Data Model, API/Interface Contracts, Sequence/Flow,
Error Handling, Testing Strategy, Security, or Performance only when relevant.

→ 🔴 **GATE**: Present `design.md` to the user.
- "Approved" → proceed to Step 4
- Changes requested → iterate, re-present
- If design changes invalidate requirements → flag and offer to update requirements

### Step 4: Generate tasks.md

Create `specforge/features/<name>/tasks.md` using `templates/tasks.tmpl.md`.

Tasks are a **flat** list. Each task declares its dependencies (`depends_on`) —
**do NOT group tasks into waves here.** Wave grouping is an execution concern
that `sf plan compute` derives from the `depends_on` graph at build time. The
`tasks.md` (the WHAT) stays flat; the build plan (the HOW-it's-grouped) is
computed. The CLI source of truth is `tasks.json`, a flat array where each task
carries `id`, `title`, `requirement_refs`, and `depends_on`.

Be honest and minimal with `depends_on`: a missing edge schedules a task too
early; a spurious one serializes work that could have run in parallel.

Every task must trace to at least one requirement.
Every requirement must be covered by at least one task.
If any requirement has no task, add one or flag it to the user.

→ 🔴 **GATE**: Present `tasks.md` to the user.
- "Approved" → `sf feature set-status --feature=<name> --to=approved`
- Changes requested → iterate, re-present

### Step 5: Update Feature Registry

Per-feature state (`specforge/features/<name>/feature.json`) is the **source of
truth** for status and gates (F10) and is
**machine state — never edit it by hand** (the hook denies a direct write). Use
the CLI, which is the only writer:

- Seal each gate the user approved this run: `sf gate approve --feature=<name>
  --phase=<phase> [--comment=…]`. The CLI appends the gate to the ledger AND
  hashes the artifact — an auditable record that a gate happened, not a claim in a
  markdown header.
- Set lifecycle fields via `sf feature`: `sf feature set-lane --feature=<name>
  --to=lite|standard` and `sf feature set-status --feature=<name> --to=approved`.
  Create new features with `sf feature add --feature=<name> [--depends-on=a,b]`.
- Append a line to `specforge/history.md` (that file is yours to write).

Read `references/gate-ledger.md` for the full `feature.json` shape and rules.

Inform the user: "Feature `<name>` approved. Use `sf-build <name>` to start implementation."

## Post-Generation: Clarify (optional)

If the user asks to refine, review, or clarify specs after generation:
read `references/clarify.md` for the refinement procedure.

## Resync Detection

If invoked on a feature that already has artefacts, check for stale downstream
artefacts (upstream newer than downstream, or a direct user edit) and offer to
regenerate. Read `references/resync.md`.