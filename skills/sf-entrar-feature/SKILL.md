---
name: sf-entrar-feature
description: >
  A request arrives for a product that already exists — turn it into a story with testable criteria
  and a verdict on whether it gets in. THIS IS THE FIRST STOP for "quiero agregar X", "necesito que
  también haga Y", "podemos sumar", "new feature", "add support for". It decides three things: does
  this belong, and how much process does it need, and what has to be true for it to be done. Works
  on a project SpecForge has never seen, by reading it first. NOT planning the change — that comes
  after, and only if the verdict says it gets in.
---

# sf-entrar-feature

**A request → what it really is → does it get in, and how much process does it need.**

## This skill is a composer

It owns the **outcome** — the story, the verdict, how much process. It does **not** own the
methods:

| when | call |
|---|---|
| this project has no `.docs/repo/` and no constitution | `Call the Skill tool with "sfx-leer-repo"` **first** |
| understanding what the request really is | `Call the Skill tool with "sfx-grilling"` |
| does it get in, and how much process | `Call the Skill tool with "sfx-decidir"` |
| writing the acceptance criteria | `Call the Skill tool with "sfx-criterio"` |

Three calls, three artefacts. Nothing else: the interview opens its own branches (search,
glossary, prototype) and the decision reads its own rubric.

## Step 0 — the ground, only if it is missing

A request cannot be judged against a product nobody described. If `.docs/repo/` does not exist and
there is no sealed constitution, `Call the Skill tool with "sfx-leer-repo"` before anything else.

**This is what lets a three-year-old repository accept a ticket** without first inventing a brief
for a product that is already built. It runs once, without the user, and the next request skips
this step.

## Step 1 — what the request really is

`Call the Skill tool with "sfx-grilling"`, telling it: the request as it arrived, that the output
file is the story's interview, and that **the tree starts in the repo, not in the abstract**. Four
branches have to be covered before the frontier can be empty:

1. **What it is** — one sentence the user agrees with. Not a title: the behaviour that will exist
   afterwards and does not exist now.
2. **The situation** — *"cuando \<situación\>, quiero \<motivación\>, para \<resultado\>"*. A
   concrete situation, not "the users".
3. **What already exists here** — is there a flow in this repo doing something similar? Go read
   it. This is the branch that decides how much process the request needs, and it is answered by
   reading, not by feeling confident.
4. **What has to be true for this to be done** — the raw material for the criteria. Do not write
   them here: that method has an owner (Step 3).

**The rubric grows here.** When the user says what matters, `sfx-grilling` hands it to `sfx-vara`
with the round it was born in. Do not interrogate them for a rubric; catch what they already said.

## Step 2 — the verdict

`Call the Skill tool with "sfx-decidir"` with the interviewed request. It reads catalog rubric ①
and answers two things at once:

```
¿entra?              va / no va / no lo pude comprobar
¿cuánto proceso?     largo | acotado
```

**How much process is a fact about the repo**: if there is an existing flow to go read, the change
is bounded; if there is nothing to read, it is not. And the ratchet turns one way only — a bounded
change that turns out not to be bounded goes up, and never comes back down.

## Step 3 — the criteria

`Call the Skill tool with "sfx-criterio"` with what came out of the interview. It owns EARS, the
test-name check and the six-defect rubric, and it is the same primitive `sfp-backlog` composes —
so a story that entered through this door and one that came out of the PRD are judged by the same
rule, instead of by two copies of it that drift.

What comes back: criteria with ids, each with a named test, and what was moved to *"what does NOT
get in"* with the reason.

## Step 4 — the story

Write the story with its frontmatter filled in. It carries the verdict, how much process, and the
criteria with ids.

```yaml
---
tipo: us | chico          # ← written HERE, by whoever decided. Not later, by somebody else.
veredicto: va
---
```

> **`tipo` is written here and that is the point.** Measured 2026-09-13: the machine routes on
> `tipo: chico`, `sfx-mapa` explains it, the story skeleton offers it — and `sfp-backlog`, the
> skill that actually writes the frontmatter, never mentions it. **A field the machine enforces
> and no skill writes.** How much process a change needs is the *output* of deciding, not a
> checkbox somebody remembers later.

## Step 5 — the acta, and it is the user's

```
───────────────────────────────────────
🛑 <el pedido, en una línea>
Propongo: <va | no va>  —  <el motivo>
Proceso: <largo | acotado>  —  <el flujo que se leyó, o "no hay ninguno">
Criterios: N  ·  todos escribibles como test: <sí | no, N quedaron afuera>
Medí: ✓N / ✗N / ?N
No pude comprobar: <lo que quedó afuera, o "nada">
───────────────────────────────────────
```

Then stop. You do not seal.

## Rules

- Compose, never re-explain. Three calls, and each method belongs to its owner.
- The tree starts in the repo. "Is there an existing flow to read" is answered by reading.
- The criteria belong to `sfx-criterio`. Compose it; never restate the rule here.
- `tipo` is written here, by whoever decided.
- `no va` is a success. A door that never refuses anything is not a door.
- One interruption: the acta.
