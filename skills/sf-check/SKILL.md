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
  "us-3/CA-1": "cumple",
  "us-3/CA-2": "no-cumple",
  "us-3/CA-3": "cumple"
}
```

**Do not answer "it works."** Go criterion by criterion, find the code that satisfies it and the
test that proves it. A criterion whose test you cannot find is a `no-cumple` with a finding, not
a benefit of the doubt.

## The ㉒: stress the tests

A green suite proves the tests pass. It does not prove they would **catch anything**.

- **If the constitution names a `mutacion` tool**, its run is already in your envelope. Read the
  score and the survivors: a surviving mutant is a change to the code that **no test noticed**.
- **If it does not** (the field is empty), do it by reading: pick the load-bearing assertions and
  ask what you could break without any test going red. Empty is not an error — the tool is an
  improvement, not a requirement.

Record the number, not the patches:

```json
"mutantes": {"herramienta": "gremlins", "score": 71.0, "sobrevivieron": 4}
```

> **Why the score and not the patches.** The mutants come from a different model each run — they
> are neither reproducible nor stable — and once a test is fixed the patch is never applied
> again. What is worth keeping is the number, because it is **comparable between rounds**:
> `71% (round 1) → 88% (round 2)`.

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

The gate counts three things: **every criterion has a verdict**, **zero open findings**, and it
warns if this is round 3 or later. All three are counting — none is an opinion, because a gate
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
- Record the mutation score, never the patches.
- All findings are born `abierto`. `descartado` is Javier's alone.
- Never carry findings forward. Re-derive them; the review is redone whole.
- Do not fix. Do not rewrite the spec. Report.
