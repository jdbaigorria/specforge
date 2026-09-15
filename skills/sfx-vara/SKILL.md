---
name: sfx-vara
description: >
  The rubric primitive — the criteria you will decide with, written BEFORE the options are on the
  table. This is the SINGLE OWNER of the rubric method in SpecForge: every skill that chooses
  between options composes this one instead of inventing its own scorecard. Triggers: "what should
  I decide this with", "score these options", "write the criteria", or any composing skill calling
  it — above all `sfx-decidir`, and `sfx-grilling` when a criterion surfaces mid-interview.
---

# sfx-vara

**In one line:** a rubric written after you have seen the options describes the option you already
picked. The order *is* the method.

## What a criterion is

A criterion is a question **answered by a fact**, not by an opinion.

```
✅  "Is there an existing flow in this repo to go read?"      you can go look
❌  "Is this well designed?"                                  that is a vote
```

If you cannot say what you would go look at to answer it, it is not a criterion — drop it or
rewrite it until it is.

Three to six. Never more. A rubric with eleven rows is a rubric where nothing discriminates.

## Every criterion carries the round it was born in

This is the whole point, so it is a field and not a convention:

> **A criterion may judge an option only if it was written before that option was on the table.**

For a rubric written up front, every criterion is born in round 0. For one that grows during an
interview, each carries the round it appeared in. Then anybody can check the order by reading the
file, instead of taking the writer's word for it.

## Two kinds

### From the catalog — it is already written, you read it

Three materials repeat identically every time, so their rubric does not get rewritten. They live
in `references/catalogo.md`:

- **a feature** — does this belong, and how much process does it need?
- **a bug** — which of these fixes?
- **a review finding** — is the reviewer right?

Read the catalog rubric, add at most one or two criteria specific to this case, and say which ones
you added.

### Bespoke — it grows during the conversation

Only one material needs this: **an idea**. You cannot know the criteria before talking about it,
because they come out of the talking.

So the rubric is not a step before the interview. **It is fed by the interview.** When the user
says something like *"for me this has to be testable without building the whole thing"* — that is
a criterion. Write it down, with the round.

This is the same gesture `sfx-grilling` already uses when a word wobbles:

```
a word is doing two jobs   →  sfx-vocabulario
a criterion surfaces       →  sfx-vara
```

**You are not adding work to the interview. You are catching what the user already said** before it
sinks to the bottom of the transcript.

## Scoring

When the options are on the table, score each one against each criterion. Three answers, never two:

```
✓   met          with the fact next to it
✗   not met      with the fact next to it
?   COULD NOT CHECK
```

**The third is the one that matters.** *"I could not tell whether this already exists"* changes a
decision, and a green that really means "I did not look" is worse than a red.

Every ✓ and ✗ carries where it came from: `retrieved` (with the link), `read` (with `file:line`),
`ran` (with the command), or `model-prior` — unverified, which is a `?` wearing a costume.

**Do not average the scores into a number.** The rubric orders the options and exposes the
trade-off; who wins is a decision, and the decision belongs to `sfx-decidir` and to the human.

## What it leaves

The composing skill names the file. If it did not, write `vara.md` next to the decision:

```yaml
---
tipo: catalogo | a-medida
criterios: 5
nacidos_en_ronda: [0, 0, 2, 2, 4]   # one per criterion, in order
---
```

Body, one block per criterion:

```
V-1 · <the criterion, as a question answerable by a fact>     (ronda 0)
      cómo se contesta: <what you would go look at>
```

And, once there are options, the table: one row per criterion, one column per option, each cell
`✓ / ✗ / ?` with its fact.

## Rules

- A criterion is answered by a fact. If it is answered by a vote, it is not a criterion.
- Three to six. A rubric where everything matters discriminates nothing.
- Every criterion carries its round. That is what makes "written before" checkable.
- Three answers, not two. `?` is a result, not a missing value.
- Never average into a single number. The rubric orders; it does not decide.
- For an idea, catch the criteria from what the user said — do not interrogate them for a rubric.

## Why this exists at all

`sf-plan` ⑫a already had it — 3 to 6 criteria `V-1…` written before the first option, with the
reason spelled out: *"una vara escrita después describe la opción que ya elegiste"*. But it lived
**inside a composer**, so `sfx-think` got the debate without the rubric and the brief got a
different one again (`primitivos.md` §3.2: three consumers, no owner).

And it is the one method in this redesign with nothing to copy: `rubric`, `scorecard` and
`criteria before` were searched in `obra/superpowers` and `mattpocock/skills` on 2026-09-13 and
returned nothing. Their `brainstorming` generates options; the ADRs record the decision afterwards.
Neither writes the rubric first.
