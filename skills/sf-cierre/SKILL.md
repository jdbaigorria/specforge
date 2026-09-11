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

**Then run `sfx-prosa` over the functional half.** That half has a human reader, and it is the one
place in the flow where prose quality is the deliverable rather than a side effect. Rule `P-1`
alone earns the pass: a sentence about this feature that would serve any other project is not
documentation, it is filler.

### And if this feature added user surface, the map gets its file

Only when `.docs/verificar/features/` exists — that is, when the project has a verification skill
(`/sfx-verificar` generates it). Skip it silently otherwise; this is not a reason to go build one.

**It is the same move you just made for the documentation**, one floor down: you know what this
feature built, and you are the last state that will. Add `<feature>.md` to the map with the four
headings the map uses — what it is, how a user reaches it, how to drive it with the harness, what
end state proves it works.

> **Why it hangs here and not on a maintenance skill.** A feature map drifts against the app, and
> a lying map is worse than no map: it sends the next agent to drive a screen that no longer
> exists. The ㉓ is the only state that runs exactly once per feature, right when what was built
> is still known — so the cheapest moment to keep the map true is this one.
>
> If nobody is keeping it up and it has drifted badly, say so in the journal. **An unmaintained
> map is a rung-5 claim with nothing behind it**, and deleting it beats leaving it.

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
- If the project has a feature map and this feature added user surface, the map gets its file.
- The journal is lessons, never progress. Progress lives in `estado.json`.
- Write the journal for a stranger planning a different feature.
- Propose a promotion to the constitution; never write one yourself.
- Do not archive, merge or delete a branch. `sf approve` does it.
