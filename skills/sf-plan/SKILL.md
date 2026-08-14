---
name: sf-plan
description: >
  Plan one feature end to end — compare three options and pick one, write the spec-design, cut
  the tasks into batches and name the tests each batch must produce. State `planificacion`
  (steps ⑫–⑯) of the SpecForge machine, invoked when `sf next` returns `skill: sf-plan`, running
  in one wide-context subagent. Composes sfx-think for the decision. Also usable standalone:
  "plan this feature", "how should we build this", "/sf-plan". Produces three files —
  decision.md, spec-design.md, tareas.json — and stops at the ⑰ for Javier.
---

# sf-plan

The biggest state. **Five steps, ONE conversation, one subagent, three files, one review at the
end.**

```bash
sf context      # → .docs/constitucion.md   your manual
                # → the us-# of this feature — what has to be solved
                # → the journals of closed features — what we learned last time
```

## Read this before anything else: you do not ask what the feature does

The ⑨ already answered that. **The `us-#` is your input, and the *what* arrives resolved.**

> **This is exactly why this state can be a subagent.** A subagent starts, works and dies — it
> does not talk to you. If it had to ask *"what does this feature do?"*, it could not be one.
> The `us-#` is not a simplification, it is an architectural requirement.

If something is genuinely missing from the story, that is not a question to ask — it is a
finding. Write it in the spec-design as an assumption you made explicit, and let the ⑰ catch it.

**Do not split this work.** All five steps are one pass with wide context on purpose: the ⑬
benefits from remembering the two options the ⑫ discarded. The ⑰ reviews the whole block, and
"I want changes" comes back to the whole block. **Nobody enters or leaves through the middle.**

## Step ⑫: Three options, and exactly three

Compose **`sfx-think`** with one added constraint: **three, not "as many as come out"**.

- Each option gets a heading `## A — <name>`, `## B — …`, `## C — …`
- Each one: what it is, what it costs, what it buys, what it forecloses.
- Then **the argument that tipped it**, and what got discarded and why.

Write `decision.md` from `templates/decision.tmpl.md`.

> **`sf` counts the headings and requires exactly 3.** This is the cheapest anti-hallucination
> gate in the flow: a single option written up as if it were a comparison is the shape a
> confident guess takes. It cannot count as three if only one was ever considered.

**The discarded options are the most valuable part of this file for a future reader.** They are
also why `decision.md` is deliberately **not** in the implementer's envelope — by the ⑱ the
decision is made, and handing them two rejected designs is an invitation to improvise.

## Step ⑬: The spec-design

One file, `spec-design.md`. It fuses what used to be two artifacts (requirements + design),
because they had the same single reader.

**The anti-N/A rule, and it is a real rule:**

> Include a section **only if it changes a decision or informs the implementation.** A section
> that would say "N/A" is noise, and noise in this file costs context in every batch's envelope.

Cover, when they apply: the shape of the solution · the boundaries and who talks to whom · the
data and its lifecycle · error and edge behaviour · what this deliberately does not do ·
assumptions you had to make.

**This is the file the ⑲ implementer works from.** Write it for someone who was never here.

## Step ⑭⑮: Tasks and their tests — one artifact, not two

Write `tareas.json` from `templates/tareas.tmpl.json`. Tasks and tests live in the same file
because the ⑲ works batch by batch and needs both in the same read.

```json
{"id": "t-1", "lote": 1, "descripcion": "…",
 "satisface": ["us-3/CA-1"],
 "tests": ["internal/docs/brief_test.go::TestParseFrontmatter"]}
```

### The batch (`lote`) is declared by you, and it is not a topological layer

A dependency graph groups by **when a task can run**. You need to group by **what goes
together**, because in this machine a batch is three things at once:

```
one batch  =  the unit of COMMIT       one finished batch = one commit
           =  the unit of SUBAGENT     one per batch, fresh context
           =  the filter of sf context "give me the tasks of batch 2"
```

Two tasks with no dependency between them can land in the same topological layer while having
nothing to do with each other — and that batch produces an incoherent commit and a subagent
holding two subjects at once. **That is the badly-grouped-commit pain, created by the tool.**

**How to cut batches, then:** each one is a coherent slice you would want to read as a single
commit message. Small enough that a fresh subagent can hold all of it; large enough to be worth
a commit.

### Every batch names its tests, exactly

```
"tests": ["path/to/file_test.go::TestName"]
```

**This list is the entire defence against "all green" with no tests.** `sf lote start` demands
the red against *these* tests, by name — not against "whatever tests exist". Without the list
there is nothing to demand.

Name tests that do not exist yet. That is the point: the ⑲ writes them, `sf` checks they showed
up and that they **failed first**.

### Every criterion gets a task

`satisface` uses `us-<n>/CA-<n>`. `sf` counts both directions and both are real errors:

- a criterion with no task → **the feature is born half-done**, before a line of code;
- a task claiming a criterion that does not exist → the task is lying, and the first count would
  not have seen it.

**This is where "done" with half the story starts dying — at planning time.**

## Step ⑯: Say which model this needs

If this feature needs a bigger model than the default to implement, say so — it is one field at
the top of `tareas.json`:

```json
{"feature": "f-1", "modelo": "opus", "tareas": [ … ]}
```

**Leave it out otherwise.** The default is the default because it is right almost always; a
`"modelo"` on every feature is noise, and noise here means every batch gets launched on a bigger
model than it needed.

It is a **recommendation, not an order.** The chain that resolves it has three levels, and yours
is the middle one:

```
sf model <nombre>   Javier, in runtime, watching the loop struggle   ← wins
tareas.json         you, here, before anything has failed
the state default
```

You know more than the default (you just planned this feature) and less than Javier (you have
not seen it fail yet). That is exactly where you sit.

## Step ⑰: Stop. This is one of Javier's three decisions

Print the summary and stop:

```
───────────────────────────────────────
🛑 ⑰ — <f-#> planned
Option chosen: <A|B|C> — <one line>
<N> tasks in <M> batches  ·  <K> criteria, all covered
Awaiting: sf approve  /  sf reject "motivo"  /  sf take <other feature>
───────────────────────────────────────
```

Then `sf done`. The gate runs five checks — the three files exist, `decision.md` has three
options, every batch has at least one test, every criterion is covered, and no task points at a
criterion that does not exist. **All five are counting. None of them is an opinion.**

If it comes back rejected, the motive travels first in your next envelope. **Read it before
anything else** — you are a fresh subagent and without it you will propose the same thing again.

## Rules

- Never ask what the feature does. The `us-#` answered it.
- Exactly three options. Not one dressed up as three.
- One pass, wide context. Do not split the block.
- Anti-N/A: a section only if it changes a decision.
- The batch is a commit you would want to read. Not a topological layer.
- Every batch names its tests, by exact name, before they exist.
- Every criterion has a task; every task points at a criterion that exists.
- `decision.md` is not for the implementer. Do not restate the rejected options in the spec.
