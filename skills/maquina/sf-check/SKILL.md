---
name: sf-check
description: >
  Review a finished feature against its criteria and stress its tests — write `revision.json`
  with a verdict for EVERY criterion plus any findings. State `revision` (steps ㉑㉒) of the
  SpecForge machine, invoked when `sf next` returns `skill: sf-check`, running in a fresh
  subagent with a big model. Also usable standalone: "review this feature", "check the
  implementation", "did we actually finish this", "/sf-check". One open finding sends the
  feature back to implement; the review is then redone whole.
---

# sf-check

The ㉑ and the ㉒. **Did it actually satisfy every criterion, and do the tests actually hold?**

```bash
sf context      # → .docs/constitucion.md   the rules it was supposed to follow
                # → spec-design.md          what was planned
                # → the us-# — the criteria, and you must opine on ALL of them
                # → a summary of the code that landed (diff since base_commit)
                # → the mutation run, if the stack has a tool for it
```

**You are a fresh subagent with a big model, and you did not implement this.** That is the
design: the reviewer who wrote the code has already convinced themselves.

## The matrix is the whole job, and it is not optional

This is the step where *"done"* with half the story dies, and the mechanism is arithmetic:

> **`sf` does not verify your judgment. It verifies that the judgment HAPPENED, over all of
> them.** *"the feature has 9 criteria and the report opines on 7"* — and it will not move.

So: **every `CA-#` of every `us-#` in this feature gets an explicit verdict.** Not a summary, not
"the rest are fine". One entry each, by id.

```json
"criterios": {
  "us-3/CA-1": {"veredicto": "cumple",    "escalon": 4,
                "prueba": "internal/suite/suite_test.go::TestCierra"},
  "us-3/CA-2": {"veredicto": "no-cumple", "escalon": 2,
                "prueba": "internal/brief/brief.go:42"},
  "us-3/CA-3": {"veredicto": "cumple",    "escalon": 5,
                "prueba": ".docs/verificar/pruebas/f-1-login.txt"}
}
```

**Do not answer "it works."** Go criterion by criterion, find the code that satisfies it and the
test that proves it. A criterion whose test you cannot find is a `no-cumple` with a finding, not
a benefit of the doubt.

## The ladder: `cumple` has to say HOW you know

The verdict used to be a two-value string, and three different things fit inside `cumple`:

```
"us-3/CA-1": "cumple"   ← I ran it and watched it work
"us-3/CA-2": "cumple"   ← there is a test with that name and I assume it proves this
"us-3/CA-3": "cumple"   ← I read it and it looked right
```

**All three were the same byte.** So every verdict now declares which rung it reached:

| | `escalon` | What it means |
|---|---|---|
| 1 | you said so | worth nothing on its own |
| 2 | you pointed at the line | a real `file:line` |
| 3 | you showed the bad case cannot happen | you walked the failure and it does not reach |
| 4 | **you ran it** | a test or script that fails loud if you are wrong |
| 5 | **you reproduced it in the running app** | via the project's verification skill |

**`prueba` is a pointer, never prose.** An `archivo_test.go::TestNombre`, a `file:line`, or the
path of an evidence artifact. If it needs a paragraph, it is not a proof.

### What `sf` counts, and it is arithmetic

- **Every criterion declares `escalon`.** Missing is a failure. There is one correct answer to
  "is the field there", so it is `sf`'s (R1).
- **A `cumple` at rung 4 or above names a `prueba` that EXISTS.** This is where the cheap lie
  dies, and it is the same trick as `sf audit`'s question ④: saying "rung 4" costs two bytes,
  having the test exist does not.
- **The floor comes from the constitution, not from `sf`.** `verificacion.escalon_minimo`. With no
  floor declared, only the declaration is required.

**Reaching rung 1 or 2 is allowed, and saying so is the point.** Anything you could not get to
rung 4 gets said, not written up as settled. If the constitution sets a floor, a `cumple` below it
is a finding.

> **Rung 5 needs the project's verification skill.** If `.docs/verificar/SKILL.md` is in your
> envelope, drive the feature with it and point `prueba` at the evidence it leaves. If there is no
> such skill, rung 4 is your honest ceiling — say so, and do not claim 5.

## The ㉒: stress the tests

A green suite proves the tests pass. It does not prove they would **catch anything**.

There are **two sources, and they do not do the same job.** Use both.

### Source 1 — the tool, if the constitution names one

Its run is already in your envelope. Read the score and the survivors: a surviving mutant is a
change to the code that **no test noticed**.

A tool mutates *syntax*: it flips a `>`, deletes a line, negates an `if`. Cheap, broad, and
identical every run over the same code — which is why you do not save its mutants. It will
regenerate them.

**If the field is empty, that is not an error** — the tool is an improvement, not a requirement.
You just do not have source 1 this round.

### Source 2 — yours, always, tool or no tool

A tool cannot think of *"and what if it starts two timers"*, or *"and what if it never stops it
on the way out"*. Those are mutations of **intent**, not syntax — the plausible-but-wrong
implementation a person would actually write. A model thinks of those or nobody does.

So write them yourself: pick the load-bearing assertions and ask what you could break without a
single test going red.

**Save them.** One file per mutant in `.docs/<feature>/mutantes/`, from
`templates/mutante.tmpl.md`. The folder is archived with the feature, so they travel with it.

The format is fixed for one reason: **the next round has to be able to re-run them.** `historial`
is append-only — one line per round, never rewritten — so the file itself carries whether this
mutant used to die.

> **Why saving them is the whole point.** Your mutants are a different set every round — 56 one
> time, 82 the next — so `73% → 85%` across them is **not an improvement, it is a different exam
> with different questions.** Saving them turns the number into a fact: next round runs *the same
> mutants*.

### On a later round, re-run the saved ones first

Four outcomes, and they are not worth the same:

| Before | Now | What it means |
|---|---|---|
| survived | survives | known debt, still open |
| survived | dies | progress — this is what you are trying to buy |
| **died** | **lives** | 🔴 **regression: a test used to catch this and no longer does** |
| — | does not apply | the code it patched changed. Retire it; not a failure |

**Every resurrection is a finding, with no exceptions and no benefit of the doubt.** A survivor
is a hole that was never covered; a resurrection is a hole that *was* covered and got uncovered —
somebody softened a test. `sf` counts them: report resurrections with no `origen: 22` finding and
the gate refuses.

### Record it

```json
"mutantes": {
  "herramienta": "gremlins", "score": 71.0, "sobrevivieron": 4,
  "propios": {"corridos": 18, "sobrevivieron": 3, "resucitados": 1, "viejos": 2}
}
```

`herramienta` empty and the three numbers at zero is the honest first round of a stack with no
tool. **What is never acceptable is inventing a percentage** for a run that did not happen: a
number invites comparison, and comparing two hand-made estimates is how a feature spends three
rounds chasing a figure that measured nothing.

## Findings

Anything worth acting on becomes a finding with an id:

```json
{"id": "h-1", "origen": 21, "criterio": "us-3/CA-2", "estado": "abierto",
 "detalle": "no hay ningún test que ejerza el camino de error"}
```

- **`origen`** — `21` if it came from the criteria, `22` if from the mutants.
- **`criterio`** — the criterion it hangs off, or empty for a ㉒ finding.
- **`estado`** — always `abierto` when you write it. **You never write `descartado`** — that one
  is Javier's, through `sf dismiss`.

**One open finding and the feature goes back to implement.** Write findings you would defend.

### Findings have no lifecycle, and this changes how you write them

There is no `arreglado` state, and nobody ever marks one fixed.

> **The review is redone whole.** Round 2 writes a brand new `revision.json`, and `h-1` simply
> **does not reappear**. *That* is what being fixed means.

Two consequences for you:

- **Never copy findings forward from a previous round.** Re-derive them. If it is still true it
  will come back on its own; if you copy it you may be reporting something already fixed.
- **Bump `vuelta`.** If it reaches 3, `sf` warns that the planning fell short — which is
  information about the ⑰, not about the implementer.

## The verdict

```json
"veredicto": "limpio"          // or "con-hallazgos"
```

Then write `.docs/features/<f-#-slug>/revision.json` from `templates/revision.tmpl.json` and:

```bash
sf done
```

The gate counts five things: **every criterion has a verdict**, **every verdict declares its
rung**, **every `cumple` at rung 4+ names a `prueba` that exists**, **zero open findings**, and it
warns if this is round 3 or later. All five are counting — none is an opinion, because a gate
stops on a **fact** and only a judge opines (R3). **You are the judge here; `sf` is only
checking you showed up for all of it.**

## What is not your job

- **Deciding whether a finding matters.** That is `sf dismiss`, and it is Javier's.
- **Fixing anything.** You review; the ⑲ fixes. Touching the code here would make the next round
  a review of your own work.
- **Judging the plan.** If the spec-design was wrong, say it as a finding — do not rewrite it.

## Rules

- Every criterion gets an explicit verdict, by id. No summaries.
- "It works" is not a review. Find the code and find the test.
- A criterion with no test that proves it is `no-cumple`.
- Every verdict declares its rung. What you could not get to rung 4, say so — do not write it up
  as settled.
- `prueba` is a pointer, never prose. At rung 4+ it has to exist in the repo.
- Record the mutation score, never the patches.
- All findings are born `abierto`. `descartado` is Javier's alone.
- Never carry findings forward. Re-derive them; the review is redone whole.
- Do not fix. Do not rewrite the spec. Report.
