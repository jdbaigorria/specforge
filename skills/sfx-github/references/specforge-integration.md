# SpecForge ↔ git — what `sf` already does, and what is left for you

## The short version

**`sf` owns the local repository. You own the remote.**

| | Who does it |
|---|---|
| create the feature branch | **`sf lote start`**, from `git.patron_branch` |
| commit a batch | **`sf done --msg "…"`** — no message, no close |
| merge and delete the branch | **`sf approve`** at the ㉓, with `git.merge` |
| push · PR · release · tag | **you** — `sf` never touches a remote |

## Why `sf` took the first three

They each used to be something a skill remembered, and each was a recurring pain:

- **the branch was not created** → now `sf lote start` creates it and refuses to proceed without
  it, so forgetting is not reachable;
- **the commit did not get made** → now closing a batch *is* committing. `sf done` has no path
  that leaves work uncommitted;
- **commits were badly grouped** → the grouping was already decided at planning time, where the
  batch was cut. One batch, one commit. There is nothing left to group well.

**None of that is a criticism of this skill.** It is R4: the convention became a field
(`git:` in `.docs/constitucion.md`) and the machine reads the field. What is left here is the
half a state machine cannot own, because it involves another party.

## Reading the conventions

```yaml
git:
  branch_por_feature: true
  patron_branch: "feat/{feature-id}-{slug}"
  commit: conventional
  merge: no-ff              # no-ff | squash | ff
  branch_base: main
```

**Use these, always.** If the project says `merge: no-ff`, do not squash because squashing is
usually nicer. The field exists so the answer is the project's, not yours.

## The commit message

`sf` makes the commit; **whoever did the work writes the message**, and it travels in
`sf done --msg`. If you are asked to write one, follow `git.commit` — one line, with the batch's
actual subject, not "wave 2" or "batch 3". The reader of a git log wants to know what changed.

## PRs

`sf` merges locally and does not know what a PR is. If the project uses them, the sequence is:

1. the ㉓ closes and Javier runs `sf approve` — the merge happens locally;
2. you push the base branch, or open the PR **before** approving if the team reviews there.

**Do not run a PR review as a second verdict gate.** The ㉑ already produced `revision.json` with
a verdict per criterion. A PR that re-litigates it is duplicated work with a worse envelope.

## What does not exist any more

For anyone reading an older version of this file: **one commit per wave**, the **gate ledger**,
`features.json`, **lanes**, **team mode**, and **exporting the roadmap to issues** are all gone.
The machine replaced the concepts they served — the roadmap is `.docs/roadmap.json`, progress is
`.docs/estado.json`, and there is one gate, not five.
