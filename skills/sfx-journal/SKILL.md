---
name: sfx-journal
description: >
  Capture evidence-anchored learnings from a session, judge them for quality, consolidate them,
  and propose promotion of recurring patterns to project invariants. Use at session close, after
  a rejected gate, after an error→fix cycle, or when the user says "journal this", "capture what
  we learned", "what went wrong", "log this lesson", "consolidate learnings", "/journal", or
  "/sfx-journal". Markdown-first (works without ICM); uses ICM as an optional enhancement if
  present. Promotion to the constitution is gated (backprop).
---

# sfx-journal

Turn what actually happened in a session into durable, curated knowledge — without
letting it grow into noise. Three levels: **journal** (raw, per session) →
**consolidate** (dedup, curated) → **promote** (gated, into the constitution).

The whole point is signal. A journal full of invented "lessons" is worse than no
journal: it costs context and lies with confidence. So everything here is
**anchored to evidence**.

## When to run

- At session close (capture).
- Right after a rejected gate, an error→fix cycle, or an explicit user correction.
- `/sfx-journal consolidate` — run consolidation only, no new capture.

## Step 1: Capture (evidence-anchored only)

Collect observations that are tied to **real evidence from this session**:

- a gate the user **rejected** (and why),
- an **error → fix** cycle (what failed, what fixed it),
- an explicit **user correction** ("no, do X instead"),
- a **repeated** mistake within the session.

Rules:

- **No evidence → no entry.** If nothing above happened, write nothing. Do not
  manufacture lessons to fill a template.
- **"What went wrong + how it resolved" beats "what went well."** The strong
  signal is corrections and error→fix cycles, not a symmetric good/bad list.
- **Separate project from meta-agent** (see `references/consolidation.md`):
  codebase quirks → project learnings; how you work in general → note it as
  meta, it does not belong in this project's files.

## Step 2: Judge (quality gate, before consolidating)

Each candidate must pass a quality check **before** it earns a place. Where the
harness has sub-agents, **delegate the judge to a fresh sub-agent** (`delegate`
pattern) so the evaluation is genuinely separate from the writer; otherwise do an
explicit adversarial re-read. Apply the rubric in `references/consolidation.md`:

- Is it **anchored** to specific evidence (not a model prior)?
- Is it **generalizable** beyond this one moment?
- Is it **actionable** (changes a future decision)?

Drop anything that fails. A vague or ungrounded note is discarded, not softened.

## Step 3: Write the raw journal entry

For each surviving observation, append to today's entry using
`templates/journal-entry.tmpl.md`:

```
specforge/context/journal/<YYYY-MM-DD>-<slug>.md
```

Use `[[wikilinks]]` to relate it to the feature, to entries in
`specforge/learnings.md`, and to constitution invariants. The journal directory
is plain markdown with wikilinks — navigable as an Obsidian-style vault, with no
dependency on Obsidian (it is just a viewer).

## Step 4: Consolidate

Reconcile the new entries against `specforge/learnings.md` (the small, curated,
always-injected file — see `references/consolidation.md`):

- **First occurrence** → add a one-line note under the right heading.
- **Recurring** (seen ~3×) → mark it a **promotion candidate**.
- **Dedup**: never add a second copy; bump the occurrence count and link the
  evidence instead.

This is where volume is controlled. `learnings.md` stays short and curated; the
raw per-session detail lives in `journal/`.

## Step 5: Promote (gated — backprop)

If a pattern reached promotion-candidate status, propose it as a new invariant in
`specforge/constitution.md`:

```
───────────────────────────────────────
🔴 GATE — backprop: proposed constitution invariant
"<the invariant>"  (seen <N>× — <links to evidence>)
Awaiting approval. Reply: approve / reject / change X
───────────────────────────────────────
```

Promotion is **never automatic** (F15). On approve, add the invariant to the
constitution and note the promotion in `learnings.md`. On reject, keep it in
`learnings.md` as a recurring note without promoting.

## ICM (optional enhancement)

If the `icm` CLI is available, mirror consolidated learnings to it for cross-session
recall and decay — `icm store -t learnings-<project> ...` on consolidate, and
`icm recall` when useful. **Markdown is the source of truth**; ICM is a projection
and recall engine, never a second truth. If `icm` is absent, everything above
still works.

## Rules

- Evidence or it didn't happen. Anchor every note to a gate/error/correction.
- `learnings.md` is curated and small — it is injected every session, so it must
  earn its tokens. The raw firehose stays in `journal/`.
- Promotion is gated. The constitution never changes as a silent side effect.
- Project learnings vs meta-agent habits are different layers — don't mix them.
