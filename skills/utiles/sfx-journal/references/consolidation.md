# Consolidation, judging, and the learning layers

The loop is `journal → judge → promote`. The naive version (one doc per session,
always injected) fails two ways: it grows without bound, and self-evaluation
invents generic lessons. The judge rubric below fixes the second; the structure
fixes the first.

## Two layers, not three

```
.docs/features/<f-#>/journal.md    written at the ㉓, archived WITH the feature
.docs/constitucion.md              the promoted rules (gated — Javier approves)
```

- **journal.md** — one per feature, evidence-anchored, with `[[wikilinks]]`. It is
  archived alongside the spec and the doc, and `sf context` serves the whole set
  to the ⑫ of every future feature. It is never injected wholesale; it is the
  trail.
- **constitucion.md** — `## Reglas de trabajo`. Only promotion-grade rules, and
  only ones a human approved.

### The middle layer is gone on purpose

There used to be a consolidated `learnings.md` between the two. It was **a second
copy of the same knowledge**, and a second copy goes stale: someone forgets to
reconcile it and it starts disagreeing with the entries it summarises.

What replaced it is not a file — it is that the archived journals are **already in
the envelope**. Counting "have I seen this three times?" means reading them, which
is slower than reading a counter and right every time.

> **What is computed cannot go stale.** The same rule that removed `estado:` from
> the `us-#` removes this file.

## The judge rubric (F32)

Borrowed from the LLM-judge pattern: evaluate quality **before** consolidating,
and ideally with a *different* evaluator than the one that wrote the note (a
fresh sub-agent where the harness supports delegation — see `delegate` skills).
A note earns a place only if all three hold:

| Test | Pass | Fail |
|------|------|------|
| **Anchored** | tied to a specific gate rejection, error→fix, or user correction this session | "it's good practice to…" with no event behind it |
| **Generalizable** | will plausibly recur on future work | a one-off quirk of this exact moment |
| **Actionable** | changes a concrete future decision | a feeling or a restatement of the obvious |

Provenance matters: a note from observed evidence (`retrieved`) outranks one from
the model's prior (`model-prior`). Discard `model-prior` notes that aren't backed
by something that actually happened.

## Consolidation rules

1. **Dedup first.** Search `learnings.md` for the same pattern. If present, bump
   its `(seen N×)` count and link the new evidence — do not add a second line.
2. **First occurrence = note.** New pattern → one curated line under the right
   heading.
3. **Recurring (~3×) = promotion candidate.** Flag it for Step 5 of the skill.
4. **Keep it short.** If `learnings.md` grows past a screen, the consolidation is
   too loose — merge related notes.

## Project vs meta-agent (don't mix layers)

Two different kinds of learning, two different homes:

- **Project-specific** (codebase quirks, this stack's gotchas, this project's
  conventions) → `learnings.md` and, when promoted, `constitution.md`.
- **Meta-agent** (how the agent works in general, universal habits) → does **not**
  belong in this project's files. If `icm` is available, store it under a
  universal topic (e.g. `preferences`); otherwise just note it for the human.

Mixing them pollutes the project constitution with universal advice — the same
mistake F28 corrects for the engineering principles.

## `[[wikilinks]]` and the vault

Journal entries link with `[[name]]`:

- `[[f-3-slug]]` — another feature whose journal saw the same thing
- `[[constitucion#reglas-de-trabajo]]` — the rule it promoted to, if any

This makes `.docs/archivado/` an Obsidian-compatible vault: open it in any
markdown graph viewer to see clusters of related lessons. The "analyze and
relate" step is consolidation expressed as link-building — surfacing which notes
cluster and which point at the same root cause. No dependency on Obsidian; it is
one possible viewer over plain files.

**And the links are what makes the missing middle layer work.** A journal that
links back to the two earlier ones that hit the same problem *is* the occurrence
count, without a file that has to be kept in sync.
