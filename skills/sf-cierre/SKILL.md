---
name: sf-cierre
description: >
  Close a reviewed feature — write the documentation someone will actually read and the journal
  the next feature will learn from. State `cierre` (step ㉓) of the SpecForge machine, invoked
  when `sf next` returns `skill: sf-cierre`, running in a fresh subagent. Composes sfx-documenter
  for the doc and sfx-journal for the learnings. Also usable standalone: "close this feature",
  "document and wrap up", "/sf-cierre". Produces two files; the archiving, merge and branch
  deletion are `sf approve`'s job, not yours.
---

# sf-cierre

The ㉓. **Two files: the doc and the journal.** Thin on purpose — the method lives in the two
utilities this composes.

```bash
sf context      # → the code that landed  → the TECHNICAL half of the doc
                # → the us-#              → the FUNCTIONAL half
                # → spec-design.md        → how it was solved
                # → the journals of closed features
```

## Why the envelope brings two sources for one document

**The doc has two halves and only one of them comes out of the code.**

```
technical    what it does, how it is wired, where things live     ← the diff
functional   what was asked for, and for whom                     ← the us-#
```

A documenter given only the code writes an accurate description of *what exists* and cannot say
**why anyone wanted it**. That second half is the part a human reads six months later, and it is
the only reason the `us-#` is in your envelope at this state.

## Step 1: The doc — compose `sfx-documenter`

Use `sfx-documenter`'s method for the technical half: real signatures, real examples, edge cases
that actually exist in the code. **Then add the functional half**, which is yours:

- what capability this feature added, in the language of whoever asked for it;
- which criteria it satisfies, and what someone can now do that they could not before;
- what it deliberately does not do (the spec-design's non-scope survives into the doc).

Write `doc.md` in the feature's folder.

**Write what is true now, not the story of getting here.** The journey is the journal's job.

## Step 2: The journal — compose `sfx-journal`

Write `journal.md`: the durable lessons, **anchored in evidence** from this feature. Not "we
learned to be careful" — what specifically went wrong or right, and what to do differently.

> **The journal is not a progress log.** How far along the feature got lives in `estado.json`,
> where it cannot go stale. This file is only for what would be **worth knowing before starting
> a different feature**.

This is why journals are archived **with** the feature and served to the ⑫ of future features.
Two questions they will ask your file:

```
is this already solved somewhere?
what did we learn last time?
```

**Write for that reader** — someone with no memory of this work, deciding how to build something
else.

### Backprop: what should stop being a lesson and become a rule

A lesson that keeps reappearing across features is not a lesson any more — it is a missing rule.
Read `references/backprop.md`.

When you find one, **propose it** for the constitution's `## Reglas de trabajo`. Do not edit the
constitution: **name the proposal in your closing message** and let Javier decide at the pause.
The last word is his, and a promoted rule changes every feature that follows.

## Step 3: Close, then the ⏸

```bash
sf done
```

The gate checks both files exist. Then `sf` prints a **⏸ pause**: the feature is ready to
archive.

**You do not archive.** `sf approve` does all of it, mechanically:

```
move the whole folder → .docs/archivado/<f-#-slug>/
merge the branch      → --no-ff into the constitution's branch_base
delete the branch
```

If the project pushes or opens PRs, that part is `sfx-github` and it happens around this — `sf`
does not touch a remote.

## Rules

- The doc has two halves. Code alone gives you one of them.
- Document what is true now. The journey goes in the journal.
- The journal is lessons, never progress. Progress lives in `estado.json`.
- Write the journal for a stranger planning a different feature.
- Propose a promotion to the constitution; never write one yourself.
- Do not archive, merge or delete a branch. `sf approve` does it.
