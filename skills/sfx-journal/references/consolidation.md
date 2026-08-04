# Consolidation, judging, and the learning layers

The loop is `journal → judge → consolidate → promote`. Each level is smaller and
more curated than the one before. The naive version (one doc per session, always
injected) fails two ways: it grows without bound, and self-evaluation invents
generic lessons. This methodology fixes both.

## The three files

```
specforge/context/journal/<date>-<slug>.md   raw, per session, append-only
specforge/learnings.md                        consolidated, curated, small, injected
specforge/constitution.md                     promoted invariants (gated)
```

- **journal/** — the firehose. Every evidence-anchored observation, with
  `[[wikilinks]]`. Never injected wholesale; it is the audit trail.
- **learnings.md** — the curated layer. Deduplicated, one line per pattern with
  an occurrence count. This is the only learning file injected each session, so
  it must stay short. This is where volume is controlled.
- **constitution.md** — the law. Only promotion-grade, gate-approved invariants.

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

journal entries and `learnings.md` link with `[[name]]`:

- `[[feature-slug]]` — the feature the lesson came from
- `[[learnings#heading]]` — the consolidated note it rolls up into
- `[[constitution#invariant]]` — the invariant it promoted to (if any)

This makes `specforge/context/journal/` an Obsidian-compatible vault: open it in
any markdown graph viewer to see clusters of related lessons. The "analyze and
relate" step is just consolidation expressed as link-building — surfacing which
notes cluster, which point at the same root cause. No dependency on Obsidian; it
is one possible viewer over plain files.
