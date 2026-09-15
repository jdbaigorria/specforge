---
name: sfx-decidir
description: >
  Decide between options against a rubric written first, and leave the reason in writing. ONE
  skill, four materials — an idea to seal, a feature to let in, a bug fix to pick, a review finding
  to accept or push back on ("is this finding right", "should I dismiss this"). This is where a
  verdict is produced in SpecForge; the entry composers call it instead of each inventing their own
  way to choose. Triggers: "help me decide", "which of these", "is this worth it", "is the reviewer
  right", or any composing skill calling it.
---

# sfx-decidir

**Given something already understood, with a rubric written before the options: choose, and say
why.**

## This skill is a composer

It owns the **outcome** — the verdict, the three exits, the acta. It does **not** own the methods
and it does not repeat them:

| when | call |
|---|---|
| there is no rubric yet, or a criterion surfaces | `Call the Skill tool with "sfx-vara"` |
| a criterion needs a fact from the world | `Call the Skill tool with "sfx-buscar"` |
| the rubric leaves two options tied | `Call the Skill tool with "sfx-think"` |
| only running code separates them | `Call the Skill tool with "sfx-prototipo"`, **after the user approves building it** |

`sfx-vara` drives the rubric and the scoring. You add what it cannot have: the verdict, and the
stop.

## The four materials

Same method, different material. The rubric comes from `sfx-vara/references/catalogo.md` for three
of them, and grows during the interview for the fourth.

| material | the question | rubric |
|---|---|---|
| an idea, interviewed | ¿vale construirlo? | bespoke, grown in the interview |
| a request | ¿entra, y cuánto proceso pide? | catalog ① |
| a defect, already diagnosed | ¿cuál arreglo? | catalog ② |
| a review finding | ¿el hallazgo es correcto? | catalog ③ |

**If you were handed a material that has not been understood yet, stop and say so.** Deciding about
an idea nobody interviewed, or a bug nobody reproduced, is guessing with a table around it. The
understanding comes first: `sfx-grilling` for the first two, `sfx-diagnosticar` for a defect.

## Step 1 — the rubric, before the options

If a rubric already exists for this decision, **read it, do not rewrite it**. Rewriting a rubric
once the options are visible is the exact failure this skill exists to prevent.

If there is none, `Call the Skill tool with "sfx-vara"` before putting any option on the table.

## Step 2 — at least two options

**Two is the floor, and "do it" / "don't do it" already counts as two.** A decision with one option
is not a decision, it is a plan with a ribbon on it.

Everything that can be settled without the user runs **now, and together**: search, grep, git log,
running the suite, reading the repo. Do not walk them one at a time and do not ask the user
anything you could go look up.

## Step 3 — the verdict

The rubric orders the options. **You pick, and you write why.**

Three exits, never two:

```
✓  va                   the reason, in one line
✗  no va                the reason, in one line
?  NO LO PUDE COMPROBAR what stayed outside what could be checked
```

The third is not a missing answer. *"I could not tell whether this already exists"* is information
that changes a decision, and it has to reach the person deciding instead of being rounded to green.

**Going against the rubric is allowed and gets written down.** If the ranking says A and you
propose B, say so out loud with the reason. What is forbidden is silently rewriting the rubric
until B wins.

**And `no va` is a success.** The value of this skill is being able to stop something. A decider
that never says no is not deciding.

## Step 4 — the acta, and the stop

Print it and stop. You do not seal:

```
───────────────────────────────────────
⚖️  <la decisión, en una línea>
Vara: N criterios (<catálogo | a medida, R rondas>)
Opciones: <A · B · C>
Propongo: <la elegida>  —  <el motivo, en una línea>
Medí: ✓N / ✗N / ?N
No pude comprobar: <lo que quedó afuera, o "nada">
Contra la vara: <"coincide" | "voy contra el puntaje porque …">
───────────────────────────────────────
```

The composing skill decides what comes after the stop — a gate, a question, or nothing.

## What it leaves

The composing skill names the file. Alongside it, the rubric and its scoring stay in `vara.md`:
this skill writes the verdict, not the table.

```yaml
---
veredicto: va | no-va | no-evaluable
motivo: "<one line, never empty>"
contra_la_vara: false
---
```

**`motivo` is mandatory, and that is not bureaucracy.** Without it, whoever redoes this proposes
the same thing again (`paradas.go:434`).

## Rules

- Compose, never re-explain. The rubric method belongs to `sfx-vara`.
- No verdict on material nobody understood. Say so and stop.
- Two options minimum. "No" is one of them.
- Everything autonomous runs first, and together. Interrupt the user once.
- Three exits. `?` is a result.
- Going against the rubric is written down; rewriting the rubric is not allowed.
- `no va` is a success.

## Why this exists

Four skills were each writing their own way of deciding, and one of them was writing none at all:

- the brief seal, the triage, choosing a fix — three shapes for the same thing
  (`primitivos.md` §3.2);
- **and dismissing a review finding, which had the stop and the command and no criterion.** The
  `f828ec2` ③ built the ME TRABÉ EN LA REVISIÓN stop whose only exit is `sf dismiss <h-#>
  "motivo"`, and its own commit message says it: *"acá es un desacuerdo sobre el mismo hallazgo…
  lo rompe alguien que decida si el hallazgo es portante."* A `sf dismiss` with no method behind it
  is what R3 forbids from the other side: a decision with no written criterion.

> **Measured, 2026-09-05.** Same idea, same seed text, two models. One sealed `no-lo-hagas` with
> 17 sources and 13 links; the other `hacelo` with 1 source and **zero** links, and the machine
> took both, because it only checked that the verdict was one of three strings. The numbers
> existed and never reached anybody's eyes. That is why the acta prints what was measured, and why
> `?` is a first-class exit instead of a green with a shrug behind it.
