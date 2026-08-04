---
name: sf-amend
description: >
  Modify a feature that was already shipped and archived, without creating a parallel spec.
  The archived spec is edited in place, so the traceability link to code stays alive instead
  of a second, diverging truth being spawned. Use when
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

## Step 1: Decide WHICH SIDE WAS WRONG — before touching anything

A spec and its code disagree. There are two ways that happens, and they are not
the same finding:

| Route | What it claims | What it edits |
|---|---|---|
| **`spec-wrong`** | the spec went out of date; the code is right | the spec |
| **`code-wrong`** | the spec was right; the code has a defect | **nothing in the spec** |

**Present both. Never suggest a default.** The pull toward `spec-wrong` is
structural — editing the spec makes the disagreement disappear immediately,
while `code-wrong` leaves you with work to do. That is exactly why the choice
has to be explicit: a spec that always yields to the code is not a contract, it
is a changelog.

Get the evidence first — `sf doctor --drift --run-tests --json`. A finding under
`implemented_differently` is the case where this decision actually bites: the
symbol resolves and its test is red, so *something* is wrong and the CLI cannot
tell you which side.

**One item at a time.** Never approve a batch of drifts in one answer — that is
choosing `spec-wrong` N times without having thought it once. If the user says
"approve everything", read back the full list and confirm it item by item before
applying anything.

### Route (b): the spec was right

```bash
# drafts/delta.json: {"why":"...", "expected":"what R3 requires",
#   "observed":"what the code does", "changes":[
#   {"op":"modify","target":"tasks","ref":"T9","description":"fix the pager"}]}
sf delta new --feature=<name> --kind=code-wrong --from=drafts/delta.json
```

`expected` and `observed` are required — a defect report without them is an
opinion. While that delta is open, saving this feature's `requirements`,
`design` or `tasks` is **refused**: you chose that the spec was right, so the
spec is read-only until the code catches up. Skip to Step 3 and fix the code;
the trace stays writable because the fix has to re-anchor.

Changed your mind? Close it explicitly:
`sf delta set-status --feature=<name> --id=D1 --to=archived`.

### Route (a): the spec went out of date

Don't regenerate the whole spec — and don't leave the delta as prose either.
Author it as a validated object (D4'): write the draft to the feature's
`drafts/` dir with your `Write` tool, then promote it:

```bash
# drafts/delta.json: {"why": "...", "changes": [
#   {"op":"modify","target":"requirements","ref":"R3","description":"..."},
#   {"op":"add","target":"tasks","ref":"T9","description":"..."}]}
sf delta new --feature=<name> --kind=spec-wrong --from=drafts/delta.json
```

Rules the object enforces: every change declares `op` (add|modify|remove),
`target` (requirements|design|tasks) and a **stable ref** (R2 stays R2 across
amends; new elements get new ids), plus a `why` — a delta without a reason is
not auditable.

Present the delta → 🔴 **GATE** (requirements). On approve, seal the gate
(`sf gate approve`) and mark it applied when the edits land:
`sf delta set-status --feature=<name> --id=D1 --to=applied`.

## Step 2: Cascade staleness (F23)

*Route (b) skips this step entirely: a `code-wrong` delta changes no spec
artifact, so there is nothing downstream to reopen.*

A change upstream reopens what depended on it:

- changed/added requirements → `design.md` gate goes **stale** → re-present design
  (delta) → 🔴 GATE.
- changed design → `tasks.md` goes stale → re-present tasks → 🔴 GATE.

The reopened artifacts show as `stale` (their sealed hash no longer matches)
until re-approved — check `sf status --artifacts`. The delta's `changes[].target`
list IS the reopen list: derive the cascade from it, not from memory. Unchanged
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
4. Archive the delta: `sf delta set-status --feature=<name> --id=<id> --to=archived`
   — the amend's audit trail (what changed, why, when) stays queryable via
   `sf delta list`.

## Rationalizations

Every one of these ends at the same place: edit the spec, make the disagreement
go away. That's the pull this skill exists to resist.

| The excuse | Why it doesn't hold |
|---|---|
| "The code is what ships, so the code is the truth" | Then nothing was ever specified — it was described after the fact. The spec's only job is to be the thing the code is checked *against*. |
| "The spec was written before we understood the problem" | Sometimes true, and that's what `spec-wrong` is for. Say it, don't assume it: the same sentence justifies every defect ever shipped. |
| "It's a small difference, not worth a defect record" | Small differences are the ones nobody re-examines. A `code-wrong` delta costs two fields; the drift it prevents compounds silently. |
| "The test is what's wrong, not the code" | Then the requirement it anchors is wrong too — that's a `spec-wrong` about the acceptance criterion, and it goes through the gate like anything else. |
| "The user said to approve everything" | Read the list back and confirm each item. "Approve everything" is consent to the *outcome* they pictured, not to N decisions they never saw. |
| "Fixing the code is out of scope for an amend" | Then the amend stops here and the `code-wrong` delta stays open. An out-of-scope fix is a scheduling problem; editing the spec to match the bug is a correctness one. |

## Rules

- **Name the wrong side before editing.** Every amend starts as a `spec-wrong`
  or `code-wrong` decision, presented with both options and no default. An amend
  that skips the question has already answered it.
- **Approve item by item.** A batch approval of N drifts is N unexamined
  decisions, all of them landing on the side that requires no work.
- **Edit in place, never fork.** One feature, one evolving traceability matrix —
  amending must not create a second parallel spec (that's the telephone game F33
  warns about).
- **Keep requirement IDs stable.** R2 stays R2 across amends; new requirements get
  new numbers. Drift and history depend on stable IDs.
- **History is append-only.** `review.md` and `history.md` record what happened;
  don't rewrite past verdicts.
- **Same ceremony.** An amend gets the same gates as the original — a "small"
  change to shipped code is exactly where false confidence bites.
