---
name: sf-amend
description: >
  Modify a feature that was already shipped and archived, without creating a parallel spec.
  Loads the archived feature and runs a delta mini-pipeline (propose-delta → gate → build →
  check) that edits the EXISTING requirements, design, tasks, and trace.json in place. Use when
  the user wants to change, extend, or fix an already-completed feature — "amend <feature>",
  "change the shipped <feature>", "the archived spec for X is out of date", "update feature X",
  "/sf-amend <feature>". Distinct from sf-propose: amend changes an existing capability; propose
  creates a new one. Promotion of changes still goes through the normal gates.
---

# sf-amend

Change a shipped feature the honest way. The archived spec is a **living
document** (F33), not a frozen snapshot — so amending edits it in place and keeps
the traceability link to code alive, instead of spawning a second, diverging
truth.

**Use amend, not a new feature, when:** you're changing an existing capability.
**Use `sf-propose`, not amend, when:** you're adding a genuinely new capability.

## Step 0: Load the archived feature

Read from `specforge/archive/<date>-<feature>/` (or `specforge/features/<feature>/`
if not yet archived):

- `requirements.md` — the current contract
- `design.md`, `tasks.md`
- `trace.json` — the structured matrix (the live link to code)
- `review.md` — historical verdict (read-only; do not rewrite history)

If the feature isn't found, stop and report — never invent the prior state.

## Step 1: Propose the delta

Don't regenerate the whole spec. State precisely what changes:

- **Requirements:** which R# are added, changed, or removed (keep IDs stable;
  new ones continue the numbering).
- **Why:** the reason for the change (new need, bug, drift found by `sf-doctor`).

Present the delta → 🔴 **GATE** (requirements). On approve, record the gate in
`features.json` (F10).

## Step 2: Cascade staleness (F23)

A change upstream reopens what depended on it:

- changed/added requirements → `design.md` gate goes **stale** → re-present design
  (delta) → 🔴 GATE.
- changed design → `tasks.md` goes stale → re-present tasks → 🔴 GATE.

Mark the reopened gates `stale` in `features.json` until re-approved. Unchanged
artefacts keep their existing gates — only what the delta touches reopens. The
enforcement hook (if installed) will hold downstream writes until each reopened
gate is re-approved.

## Step 3: Build the delta

Run the build for the changed tasks only (reuse `sf-build` semantics): edit the
existing code, don't rewrite untouched files. Gate after each wave. A failed
operation halts — same hard rules as a normal build.

## Step 4: Re-check and update the matrix in place

Run `sf-check` semantics, but **edit the existing `trace.json`** rather than
creating a parallel matrix:

- update `code`/`test` anchors for changed requirements,
- add rows for new requirements,
- remove rows for removed requirements.

Build the full traceability matrix and verdict as usual → 🔴 **GATE**. The old
`review.md` stays as history; write the new verdict as a dated amend entry.

## Step 5: Re-seal

On APPROVE:

1. Update the archived feature in place (requirements/design/tasks/trace.json).
2. Append to `specforge/history.md`:
   ```
   ## [date] — Feature amended: <name>
   - Change: [summary of the delta]
   - Requirements touched: [R#...]
   ```
3. Append the verdict gate to the feature's `gates[]` ledger (don't overwrite
   the original completion).

## Rules

- **Edit in place, never fork.** One feature, one evolving traceability matrix —
  amending must not create a second parallel spec (that's the telephone game F33
  warns about).
- **Keep requirement IDs stable.** R2 stays R2 across amends; new requirements get
  new numbers. Drift and history depend on stable IDs.
- **History is append-only.** `review.md` and `history.md` record what happened;
  don't rewrite past verdicts.
- **Same ceremony.** An amend gets the same gates as the original — a "small"
  change to shipped code is exactly where false confidence bites.
