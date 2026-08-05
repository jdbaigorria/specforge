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

#### Batch mode — when the gaps are already known

The conversation above is for *discovering* what the feature is. Once you can
already name the open questions — after reading a PRD, or when `--from` seeded
the feature — asking them one at a time is pure fatigue. Switch to a batch:

1. **List the whole batch first** — number + topic, asking for no answers yet.
   The user sees the scope before committing to anything.
2. **Closed options, numbered, 2 to 4 per question.** This maps 1:1 to
   `AskUserQuestion`.
3. **Group only what is genuinely interdependent.** Bundling unrelated questions
   forces a single answer onto separate decisions.
4. **Open questions stay loose, in prose.** Don't force them into the selector —
   a 4-option menu on a genuinely open question just hides the real answer.

> **This is not `sfx-grill-me`, and the two rules do not conflict.** They do
> different jobs:
>
> | | `sfx-grill-me` (adaptive) | Batch mode (here) |
> |---|---|---|
> | For | Probing assumptions — each answer changes the next question | Filling gaps **known in advance** |
> | Why one-at-a-time / batched | Batching destroys the adaptation | Asking 20 known gaps one by one is fatigue |
>
> If the gaps are not yet known, you are still discovering: stay in the
> conversation above, or use `sfx-grill-me`.

### Step 2: Generate requirements

Author the draft, then promote it — **never write the artifact directly**:

```bash
# 1) Write specforge/features/<name>/drafts/requirements.json with the Write tool
sf save requirements --feature=<name> --from=drafts/requirements.json
```

`sf save` validates against the schema and only then writes both the
authoritative `requirements.json` and the rendered `requirements.md`. Writing
either by hand is denied by the hook — `drafts/` is the authoring corner, and
`sf save` is the only sanctioned writer.

Use EARS notation for system behaviors (`WHEN [trigger] THE SYSTEM SHALL
[behavior]`). Read `references/ears-notation.md` if needed. Every requirement
needs testable acceptance criteria, plus edge cases and an explicit out-of-scope
section.

**Give every acceptance criterion an id.** The id is what lets a test anchor to
*that* case — without it, the verdict gate can only ask "does R5 have a test?",
and a requirement with five criteria seals green on one:

```jsonc
{
  "id": "R5", "ears_type": "event", "priority": "must",
  "acceptance": [
    { "id": "R5.1", "text": "the order is persisted as `pending`" },
    // the long form, when the precondition carries weight:
    { "id": "R5.2", "given": "an empty cart", "when": "the user confirms",
      "then": "it is rejected with `EMPTY_CART`" }
  ]
}
```

`id` is `<requirement>.<n>`, numbered from 1. Use `text` **or** given/when/then,
never both, and never half the long form (a `when` without a `then` says less
than the one-liner). The id is permanent: a new criterion takes the next free
number, and **a retired id is never reused** — anchors point at ids, so reusing
one silently re-points a test at a different case.

**`priority` is `must` | `should` | `could`**, and omitting it means `must`. Set
it deliberately — it's how a project decides what can and can't hold up a
release (`verification.blocking_priorities` in the constitution). Ask the user
rather than guessing: a `could` you invented is a requirement nobody agreed to
deprioritise. When in doubt, leave it out and get `must`.

### Cite where each requirement came from

If the user handed you material — an email, a transcript, a ticket, a document —
declare it once, at project level, and cite its id from every requirement it
produced:

```bash
# 1) Write specforge/drafts/sources.json with the Write tool:
#    {"sources":[{"id":"S1","kind":"email","ref":"inputs/client.md",
#                 "captured":"2026-07-14","note":"original billing request"}]}
sf save sources --from=drafts/sources.json
```

Then `"source": ["S1"]` on each requirement. `ref` is a project-relative path or
a URL, and a path **must exist** — a source pointing at a missing file can't be
told apart from a fabricated one.

**Why this is not bookkeeping.** `trace.json` already stops you from claiming
"I implemented R5" with no code and no test. Nothing stopped you from *inventing
R5*. Citing the source closes that same hole at the other end of the pipe, and
`sf sources coverage` reads it both ways:

- **a requirement with no source** — you invented it;
- **a source no requirement uses** — material that was read and never made it
  into the spec.

A requirement with no source is legitimate when the idea is the user's own —
that's why nothing blocks by default. Don't invent a ref to silence the report;
if you don't know where a requirement came from, that *is* the finding. Projects
that want it enforced set `verification.require_source: true`.

→ 🔴 **GATE**: Present `requirements.md` to the user.
- "Approved" → proceed to Step 3
- Changes requested → iterate on requirements, re-present
- If user edits the file directly → acknowledge changes, proceed

### Step 3: Generate design

Same path as Step 2 — draft, then promote:

```bash
# 1) Write specforge/features/<name>/drafts/design.json with the Write tool
sf save design --feature=<name> --from=drafts/design.json
```

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

### Step 4: Generate tasks

Same path again:

```bash
# 1) Write specforge/features/<name>/drafts/tasks.json with the Write tool
sf save tasks --feature=<name> --from=drafts/tasks.json
```

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