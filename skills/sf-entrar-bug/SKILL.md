---
name: sf-entrar-bug
description: >
  Something is broken — turn a symptom into a named root cause, a failing test, and a chosen fix.
  THIS IS THE FIRST STOP for "esto no anda", "encontré un bug", "tira este error", "se rompió", "it
  crashes", "the test is failing", a stack trace pasted with no question attached. It reproduces,
  diagnoses and picks the fix against a rubric — it does NOT write the fix. Works on a project
  SpecForge has never seen, by reading it first.
---

# sf-entrar-bug

**A symptom → a named cause → a failing test → the chosen fix.**

**Framing, non-negotiable:** the lane for a bug skips planning, and that is correct — but skipping
planning is not the same as skipping thinking. What replaces the plan is the diagnosis, and this
door exists because the lane had no method in it.

## This skill is a composer

| when | call |
|---|---|
| this project has no `.docs/repo/` and no constitution | `Call the Skill tool with "sfx-leer-repo"` **first** |
| reproducing and finding the cause | `Call the Skill tool with "sfx-triage"` |
| choosing among the candidate fixes | `Call the Skill tool with "sfx-decidir"` |

Two calls, plus the ground if it is missing. Do not run your own investigation loop alongside the
diagnosis, and do not recommend a fix before the decision — the diagnosis deliberately hands over
candidates **without a recommendation**, so that the rubric is what picks.

## Step 1 — the cause

`Call the Skill tool with "sfx-triage"` with the symptom exactly as it arrived: the full
error, the steps, what the user was doing.

**This entire step runs without the user.** Reading the error, reproducing it, checking what
changed, tracing the bad value back to its source — none of it is a decision. This is the door that
interrupts the least of the three, and that is not an accident: nothing here needs a judgement
call.

It ends with a cause named at `file:line`, a test that fails today, and the candidate fixes.

### If it could not be reproduced

That is a result, not a failure. Stop here and say so as `?` in the acta. **Do not go on to choose
a fix for a defect nobody saw happen** — that is guessing with a rubric around it, which is worse
than guessing plainly.

## Step 2 — which fix

`Call the Skill tool with "sfx-decidir"` with the diagnosis. It reads catalog rubric ②, whose five
criteria all answer themselves without the user: does it attack the cause or cover the symptom, is
there a test that fails now and passes after, does it change one thing, does it break anything
green, and does it fix where the problem is born rather than where the error shows up.

## Step 3 — the acta, and it is the user's

```
───────────────────────────────────────
🛑 <el síntoma, en una línea>
Reproducido: <sí, N de N veces | NO>
Causa: <una línea>  —  <file:line>
Apareció en: <commit | "no lo pude ubicar">
Test que falla: <nombre>  —  <corriendo hoy: rojo>
Propongo: <el arreglo elegido>  —  <el motivo>
Otros candidatos: <A · B>
No pude comprobar: <lo que quedó afuera, o "nada">
───────────────────────────────────────
```

Then stop. You do not fix and you do not seal.

## The three-strike rule belongs to this door too

If three fixes have already failed on this defect, the problem is not the hypothesis — it is the
shape of the thing. Stop, say so, and put it to the user. Do not attempt a fourth.

> Superpowers cuts at three and hands it to the human; SpecForge's `TopeIntentos = 3` has said the
> same since 2026-09-03. Two projects that never read each other landed on the same number.

## Rules

- No fix without a cause. "It looks obvious" is not an exemption.
- Could not reproduce is a `?`, and it stops the door. Never choose a fix for a defect nobody saw.
- The failing test exists before anybody writes the fix.
- Fix where it is born, not where it shows up.
- The diagnosis hands over candidates with no recommendation; the rubric picks.
- Three failed attempts is a conversation, not a fourth attempt.
- This door ends at a chosen fix. Writing it is somebody else's job.

## Why this exists

`tipo: bug` skips planning and lands straight in a build skill, against a plan that in this lane is
never written. The diagnosis is what replaces that plan — and **it had no door**: `sfx-triage` did
the work and nothing routed a symptom to it, so a bug arriving cold landed on whoever answered
first.

> **Correction, 2026-09-16.** `vecinos-metodos.md` §4.1 said the method did not exist here and
> that *"`sfx-triage` classifies the ticket without diagnosing the defect"*. **That was false** —
> the file has the iron law, backward tracing, the working-example diff, ranked hypotheses with
> evidence, a hard limit, and the failing test before the fix. What was missing was the door and
> three details, not the method. Both neighbours (Matt's `diagnosing-bugs`, superpowers'
> `systematic-debugging`) agree debugging goes **before** touching code; so did we, and nobody had
> checked.
