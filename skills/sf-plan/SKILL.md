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

### One test per criterion, minimum — and `sf` counts it

A task that claims to satisfy four criteria and names one test does not pass the ⑰:

```
✗ la tarea t-3 dice satisfacer 4 criterios y nombra 1 test
```

The count is not a judgment about quality — `sf` cannot read a test. It is the cheapest possible
floor against the failure that actually happens: **one happy-path test per task, and everything
else discovered three review rounds later.**

If one test genuinely covers two criteria, name two anyway and split it. A test that proves two
different things has two reasons to fail, and when it goes red you do not know which.

### How to find the tests that are missing: ask what would still pass

Naming one test per criterion is the floor. This is how you get past it — and it is one question,
not a taxonomy to fill in:

> **What could be wrong in the implementation and still pass this test?**

Worked through:

```
criterion   "the counter starts when you enter the screen"

the obvious test      enter → the counter started                    ✓
now the question      what broken implementation passes that too?
                        → one that starts TWO timers
                        → one that never stops it on the way out
                        → one that restarts it on every render
```

Those three answers are three tests, and they are the ones nobody writes.

**Do not reach for a checklist of test categories** — "one happy path, one edge, one invariant".
A category gives you the tests the category asks for. The question gives you the tests *this
criterion* needs, and for some criteria the answer is one test and for others it is five.

> **This is the ㉒ asked early.** The mutation run at the end of the feature asks exactly this
> question against the finished code, and everything it finds sends the feature back to
> `implementar`. Asking it here costs a line in a JSON file. Asking it there costs a round.

### Every criterion gets a task

`satisface` uses `us-<n>/CA-<n>`. `sf` counts both directions and both are real errors:

- a criterion with no task → **the feature is born half-done**, before a line of code;
- a task claiming a criterion that does not exist → the task is lying, and the first count would
  not have seen it.

**This is where "done" with half the story starts dying — at planning time.**

## Step ⑯: Choose the model for each batch

`sf context` handed you a **menu of the models declared on this machine** — aliases, capacities,
and one line each on what they are good for. Read it. This step is where you use it.

```json
{"feature": "f-1",
 "modelo": "medio",
 "modelo_por_lote": {"1": "barato", "3": "grande"},
 "tareas": [ … ]}
```

`modelo` is the feature-wide choice; `modelo_por_lote` overrides it for one batch. **Both take an
ALIAS from the menu — never a model id.** An id is a fact about Javier's machine, and `tareas.json`
is versioned and will be read by another harness where that id does not exist. The menu
deliberately does not show you ids, so you cannot copy one by accident.

### Leaving it out is the normal answer

The default of each profile is the default because it is right almost always. A `modelo` on every
feature is noise, and noise here means every batch gets launched on a bigger model than it needed.

**Choose only where the difference is real.** A batch that wires up CRUD, moves plumbing or does a
mechanical refactor can run on the cheap one. A batch that decides something — a concurrency
boundary, a data model, an error contract — cannot. If you cannot name what makes this batch
different, say nothing.

### It is a recommendation, not an order

The chain that resolves it has four levels and yours are the middle two:

```
sf model <alias>       Javier, in runtime, watching the loop struggle   ← wins
tareas.json, batch     you, here — "batch 3 is the hard one"
tareas.json, feature   you, here — "all of this one is hard"
the profile of the state
```

You know more than the default (you just planned this feature) and less than Javier (you have not
seen it fail yet). That is exactly where you sit.

> **If the menu was empty or missing**, say nothing at all: `sf next` will stop and ask Javier to
> declare the profiles before it launches anyone. Guessing a name would produce a plan that stops
> the loop.

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
- At least one test per criterion the task claims. `sf` counts it at the ⑰.
- For each criterion: what could be wrong and still pass? Those answers are tests too.
- Every criterion has a task; every task points at a criterion that exists.
- `decision.md` is not for the implementer. Do not restate the rejected options in the spec.
