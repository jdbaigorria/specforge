---
name: sfx-mapa
disable-model-invocation: true
description: >
  Ask which skill fits your situation. A router over every skill in SpecForge and over the
  machine they hang off. Takes an optional situation ("I have a bug", "the review keeps
  bouncing", "I want to build X") and answers with the one entry point that fits.
---

# sfx-mapa

You do not remember twenty-six skills. Ask.

**This is a map, not a dispatcher.** Nothing here invokes anything: it names the entry point and
you take it. The one thing it can settle for you is *which* entry point, which is the question
that actually goes wrong — a fuzzy idea landing in a debate skill instead of in research.

---

## The one rule that outranks the rest

**In a SpecForge project, the answer is almost always `sf next`.**

The machine knows where you are, what comes next and which skill does it. If `.docs/estado.json`
exists, run `sf next` before reading any further down this page. Everything below is for the
questions `sf next` does not answer: you are outside a SpecForge project, or you are looking for
a utility that lives outside the machine, or the machine stopped and you have to decide something.

---

## The machine — nine skills you do not choose

These nine are named BY `sf next`, dispatched by the orchestrator, and they report back with
`sf done`. **You do not pick them off this list.** They are here so you can see the shape.

```
PRODUCTO   (once per product)
  brief          sfp-scout         idea → sealed brief, with a verdict          🛑 ⑥
  prd            sfp-po            brief → actors, capabilities, scope
  constitución   sfp-constitucion  architecture, stack, conventions             🛑 ⑧
  backlog        sfp-backlog       PRD → us-# with criteria that have ids       ⏸ ⑨
  roadmap        sfp-roadmap       group into features, put them in order

FEATURE    (once per feature)
  planificación  sf-plan           3 options → spec-design → tasks → tests      🛑 ⑰
  implementar    sf-build          one batch: tests first, red, code, commit
  revisión       sf-check          a verdict per criterion, plus findings
  cierre         sf-cierre         the doc and the journal                      ⏸ ㉓
```

**Three paths through it, and the middle one is the one people miss.** The path is declared in
each story's `tipo:` and the longest one in a feature wins:

| `tipo:` | skips | for |
|---|---|---|
| `us` | nothing | new behaviour, anything architectural, anything in a repo with no flow to read |
| `chico` | planificación | a bounded change to a flow that ALREADY EXISTS here: one more flag, one more field, a missing validation |
| `bug` | planificación and revisión | something that used to work and stopped |

**`chico` measures the repo, not your confidence.** If there is no existing flow to go read, it is
not small — it is `us`, however simple it sounds. Reaching for the light label to skip the
planning IS the doubt that means you should not.

**And the ratchet only turns one way.** When a `chico` turns out not to be, run
`sf ampliar [<f-#>] "why it was not small"`. It goes back to ⑫ and it never comes back down —
not by editing the story, not by running the command again.

---

## When the machine stops

`sf next` prints four kinds of stop. Two are yours to answer, two are brakes on a loop.

| | what it means | your way out |
|---|---|---|
| 🛑 | a gate: ⑥, ⑧ or ⑰ | `sf approve` · `sf reject "motivo"` |
| ⏸ | cheap: look and continue (⑨, ㉓) | `sf approve` |
| ⚠ ME TRABÉ | `sf done` failed 3 times running | `sf model <alias>`, or go in yourself |
| ⚠ ME TRABÉ EN LA REVISIÓN | the ㉑ bounced it 3 times | `sf dismiss <h-#> "motivo"` — adjudicate each finding |

**The last two are different problems and that is why the ways out differ.** The first is
capacity, and more model fixes it. The second is a disagreement between the reviewer and the
implementer about the same finding, and another attempt does not fix that — somebody has to rule
on whether the finding is load-bearing. `sfx-interrogar` is worth reaching for there: if a second
set of models finds the same thing on its own, it is real.

---

## The utilities — the ones you DO pick

Sixteen, and they split on one axis: **who can reach them.**

### The primitives — the model composes these, you rarely call them

Each one is the single owner of its method. A state skill that needs the method composes the
primitive rather than re-explaining it, which is why they stay reachable by the model.

| | the method it owns |
|---|---|
| `sfx-grilling` | the interview: decision tree, frontier, rounds |
| `sfx-buscar` | research: cheapest tier first, provenance on every claim |
| `sfx-vocabulario` | the glossary: one word, one meaning, in `vocabulario.md` |
| `sfx-prototipo` | throwaway code that answers ONE question |
| `sfx-prosa` | cutting the tells out of Spanish prose, by numbered rule |
| `sfx-think` | weighing options **already on the table** |
| `sfx-tdd` | red → green → refactor, one slice at a time |

> **`sfx-think` needs its options named before you get there.** A raw idea — "quiero armar X",
> "tengo una idea" — goes to `sfp-scout`, which starts by finding out whether the thing already
> exists. Debating an idea nobody researched is the most expensive way to be wrong.

### The workers — model-reachable, and useful on their own

| | |
|---|---|
| `sfx-triage` | a bug, down to root cause, with a fix plan |
| `sfx-documenter` | documentation out of code |
| `sfx-journal` | evidence-anchored learnings from a session |
| `sfx-explain` | a concept, Feynman-style |
| `sfx-github` | the remote: push, PRs, releases — what `sf` deliberately does not touch |
| `sfx-skill` | writing or fixing a skill, tested against a subagent before you trust it |

### User-invoked — only you can fire these

They carry `disable-model-invocation: true`, and the reason is the same in all five: each is an
**entry point a human chooses**, and a rich trigger list on an entry point wins races it should
lose. The model cannot reach them, including from inside another skill.

| | when you reach for it |
|---|---|
| `sfx-mapa` | this page |
| `sfx-grill-me` | stress-test a plan or a design with no repo under it (in a repo, the interview is already inside `sfp-scout`) |
| `sfx-verificar` | once per project: generate the skill that drives the REAL app, which is what makes rung 5 of the ㉑ reachable |
| `sfx-interrogar` | a high-risk diff before the ㉑, or a stuck review — several models on the same diff |
| `sfx-audit` | after several features closed: does the whole thing still hold together |

---

## Outside a SpecForge project

Everything under **the utilities** works anywhere. The nine state skills do not: they open with
`sf context` and close with `sf done`, and without a machine behind them there is no envelope to
ask for and nobody to report to.

The one exception worth knowing: **`sfp-scout` is the front door.** A raw product idea goes there
whether or not `sf init` has run — it interviews, researches whether the thing exists, builds the
glossary and reaches a verdict. What it cannot do outside a project is seal the ⑥.

---

## Keeping this page honest

A router that names a skill that no longer exists, or that misses one that does, is worse than no
router: it is confidently wrong at the exact moment somebody is lost.

**Whenever a skill is added, renamed, removed, or changes which flow it belongs to, this file gets
re-read and updated in the same change.** `docs/skills.md` is the human-facing catalogue and it
carries the same obligation; the two say the same thing to two different readers.
