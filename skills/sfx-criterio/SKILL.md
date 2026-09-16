---
name: sfx-criterio
description: >
  The acceptance-criterion primitive — write criteria that can become tests, and catch the ones
  that cannot. Owns EARS notation and the six-defect quality rubric. This is the SINGLE OWNER of
  the criterion method in SpecForge: every skill that writes, receives or judges a `CA-#` composes
  this one instead of restating the rule. Triggers: "write the acceptance criteria", "is this
  testable", "turn this into criteria", or any composing skill calling it — `sfp-backlog`,
  `sf-entrar-feature`, and `sf-plan` when a criterion refuses to become a test.
---

# sfx-criterio

**The rule, and it is R5:**

> **If you cannot write the test, it is not an acceptance criterion.**

Not a warning. A criterion that fails it gets rewritten or moved to *"what does NOT get in"* —
never left as prose somebody will "check by eye" later.

## Why this is one method and not a paragraph in three skills

A criterion has three different skills around it and the rule that binds them lived in only one:

```
who writes it        sfp-backlog · sf-entrar-feature
who turns it into a test     sf-plan ⑭⑮, and counts them
who walks it with a verdict  sf-check ㉑
```

**Whoever writes the criterion and whoever has to be able to test it are different skills.** When
the rule lives in only one of them, the other two get a criterion they cannot use and have to
improvise (`primitivos.md` §3.3: three consumers, no owner).

## Step 1 — write it in EARS

Read `references/ears-notation.md`. The reason EARS survives when everything else was cut:

```
a test needs   →   trigger · state · expected behaviour
EARS asks for  →   trigger · state · behaviour
```

**It is the same triple.** A criterion written in EARS converts into a test without inventing
anything. One that is not, does not — and the failure is concrete:

```
CA-1  acepta --json y devuelve el estado serializado     ← ubiquitous ✔
CA-2  si no hay estado, sale con código 1 y mensaje      ← IF…THEN ✔
CA-3  --json es incompatible con --verbose               ← ⚠ ¿y entonces qué pasa?
```

The third does not say what happens, **so no test can be written against it.** EARS would have
forced it.

Format, because `sf` counts these with a pattern match:

```markdown
## Criterios de aceptación
- **CA-1** — <the criterion>
- **CA-2** — <the criterion>
```

## Step 2 — the test-name check

For each criterion, **name the test that would prove it.** Out loud, before moving on.

If you cannot name it, you have found a defective criterion — not a hard criterion. Go back to
Step 1, or move it out.

This is the cheapest check in the whole method and it is the one that gets skipped.

## Step 3 — the six-defect rubric

Read `references/requirement-quality.md` and run all six over what you just wrote: ambiguous, not
singular, not verifiable, contradiction, inconsistent terminology, incomplete.

**All six, every time.** One that does not apply is reported `N/A — <reason>`, never silently
skipped. And a finding **quotes the exact text** it applies to: a finding you cannot point at is
an opinion.

`sf` catches none of these — every one is judgment (R3). This rubric is the only thing between a
vague criterion and a feature that ships half-done.

## What does NOT belong on a criterion

```
prioridad     a criterion exists or it does not. There is no "nice to have" tier —
              that is what deleting is for
estado        derived from roadmap.json + estado.json
```

A field a model can forget to update is a file that lies forever and nobody notices.

## What it leaves

The composing skill owns the file — this primitive writes criteria into whatever it was given, and
reports:

```
Criterios: N
Con test nombrado: N
Movidos a "qué NO entra": N   ← con el motivo
Rúbrica: RQ-1…RQ-6 corridas  ·  hallazgos: N
```

## Rules

- If you cannot write the test, it is not a criterion. Rewrite it or move it out.
- EARS, because trigger/state/behaviour is the same triple a test needs.
- Name the test before you accept the criterion.
- All six defects, every time. `N/A` is an answer; silence is not.
- A finding quotes the text. No quote, no finding.
- A criterion with an `and` joining two independently acceptable behaviours is two criteria.
- No priority, no state. The criterion is text.
