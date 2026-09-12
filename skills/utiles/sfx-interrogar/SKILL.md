---
name: sfx-interrogar
description: >
  The adversarial review primitive. Several models review the same diff independently, and what
  two of them find on their own is the high-signal finding. Produces findings in `revision.json`
  shape, ready to paste. Triggers: "/sfx-interrogar", "interrogá esto", "que lo rompan", "review
  this with several models", "find the blind spots", or a high-risk feature before the ㉑.
  Standalone: works on any diff, with or without SpecForge.
---

# sfx-interrogar

**One reviewer tells you what is wrong. You cannot tell whether it is a truth or that model's
pet peeve. Two reviewers who found it on their own can.**

That is the whole value here, and it is the only thing this skill adds over the ㉑, which already
has the two expensive halves: a fresh reviewer who did not implement, on a big model.

## Read this first: when NOT to run this

**The ㉑ already covers most of it.** Run this when at least one is true:

- the change is a one-way door — a data format, a public contract, a migration that deletes;
- the ㉑ has already gone two rounds and keeps finding new things, which means the first reviewer
  is not seeing a class of problem;
- the blast radius is wide and the diff looks small, which is the combination that hides things.

**Otherwise, do not.** This costs N models per review, and a skill that adds nothing over what is
already there is debt with a feature's face. Say you are skipping it and why.

## 1. Scope and intent

Identify what is under review: a diff, a branch (`git diff main...HEAD`), or named files.

Then **state the intent in one paragraph**, from the `us-#`, the commits, and the code. The
reviewers need to know what this was supposed to do, or they review it against their own guess and
every one of them guesses differently.

## 2. Fan out

One reviewer per model, all launched together, each with **the same prompt** — the adversarial
signal comes from the models being different, not from assigning them personas.

Take the aliases from the model menu, never ids: an id is a fact about one machine, and this skill
runs on other people's. Read-only; nobody writes anything.

Each reviewer gets the intent, the diff, the constitution when there is one, and this question:

> What is wrong here that the author would defend as intentional?

**Ask for the failing case, not the opinion.** A finding is *"this input, this state, this wrong
output"*. A finding that cannot name a case is a preference.

## 3. The consensus map

Findings come back in three piles, and they are **not** worth the same:

| | What it means | How it gets written |
|---|---|---|
| **2+ models, independently** | highest signal there is | the detail says how many found it |
| **one model** | still worth reading, weighted lower | the detail says `un solo revisor` |
| **models contradict each other** | the most interesting of the three | the detail carries both readings |

Deduplicate first: two models describe the same bug in different words more often than they find
two bugs. Merge them and say who raised it.

> **Why a contradiction is the best outcome.** Two competent readers looking at the same code and
> disagreeing about what it does means the code does not say what it does. That is a real finding
> even when neither reviewer is right.

## 4. Write them as findings, and nothing else

Output is a block in `revision.json` shape, ready to paste:

```json
{"id": "h-4", "origen": 21, "criterio": "us-3/CA-2", "estado": "abierto",
 "detalle": "2 de 3 revisores, independientes: con la conexión caída el retry corre para siempre — retry.go:88 no tiene tope"}
```

**Every one is born `abierto`.** You never write `descartado`: that one is Javier's, through
`sf dismiss`.

### What this skill deliberately does NOT do

The well-known version of this sorts findings into four buckets — act on, consider, noted,
dismissed. **That is rejected here, and not for lack of craft.**

Those buckets hand the model the decision about what matters, and this machine took that decision
away on purpose: a finding is born open, and only Javier closes it. A reviewer that arrives
pre-sorted into "you can ignore these" has already made the call.

> **Consensus is information travelling TO the decision, not a filter placed in front of it.** So
> the count goes in `detalle`, where Javier reads it, and the state stays `abierto`.

It also **does not fix anything**. Touching the code here would make the next round a review of
your own work.

## Where it sits relative to the ㉑ and the auditor

```
sfx-interrogar   varios modelos · UN diff       · cuando el riesgo lo pide
sf-check (㉑)     un revisor     · UNA feature   · siempre, al terminarla
sfx-audit        un modelo      · VARIAS features · cuando Javier quiere
```

It is a `sfx-`: it never asks for an envelope, never calls `sf done`, and `sf next` never returns
its name. Run it before the ㉑ and the findings travel in with the review; run it on a bare diff
and it is just a report.

## Rules

- Say why you are running it. "The risk asked for it" names a risk.
- The same prompt to every reviewer. Diversity comes from the models.
- A finding names a failing case, not a preference.
- Deduplicate before counting. Two wordings of one bug are one bug.
- The count goes in the detail; the state is always `abierto`.
- Do not sort findings into what matters. That is `sf dismiss`, and it is Javier's.
- Do not fix. Review, and hand back.
