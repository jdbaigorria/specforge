---
name: sf-build
description: >
  Implement one batch of an approved plan, test-first. State `implementar` (steps ⑱–⑳) of the
  SpecForge machine, invoked when `sf next` returns `skill: sf-build`, running in a fresh
  subagent — one per batch. Composes sfx-tdd for the red-green-refactor cycle. Writes the
  planned tests, confirms they fail, implements until green, and closes the batch with a commit
  message it writes itself. Also usable standalone: "implement this batch", "build the feature",
  "/sf-build". It does not run the loop — the orchestrator does.
---

# sf-build

The ⑲. **One batch, one fresh subagent, one commit.**

```bash
sf context      # → .docs/constitucion.md   your manual
                # → spec-design.md          what to build and how
                # → your tasks (batch N of M), embedded
                # → the us-# — the criteria you have to satisfy
```

**You get one batch, not the feature.** That is deliberate: the mocks showed up *"with the
context at 50%"*, and a fresh subagent per batch is the only place in this flow where that
damage can be prevented instead of detected.

**`decision.md` is not in your envelope, and that is on purpose.** The decision is made. Two
rejected designs would only be an invitation to improvise.

## Preflight

Before writing anything:

- **Read the constitution.** It is your manual — conventions, folder layout, working rules, the
  approved dependency list. **You have no conventions of your own** (R4); everything you would
  have hardcoded lives there.
- **Check `test_cmd` runs.** If it does not, stop and say so. Everything below depends on `sf`
  being able to run the tests.
- **Do not add a dependency that is not in `dependencias_aprobadas`** without saying so out loud
  in your closing message. `sf` will warn — it will not block, because the last word is Javier's
  — but a dependency that slips in silently is exactly the pain this field exists for.

## Step ⑱: Start the batch — and this is a gate, not a formality

```bash
sf lote start
```

This does three things and **refuses to do them if the tests are not honestly red**:

1. **Creates the branch** from the constitution's pattern. You never create it by hand — that is
   how it used to get forgotten.
2. **Demands the red**, against the exact test names the plan declared. Two ways to fail:
   - *the planned test does not exist* → it tells you which one;
   - *the test exists but already passes* → **a test that passes before the code exists is a
     fake test.** Write it so it actually exercises the behaviour.
3. **Saves the hash of the test files**, which is what makes Step ⑳ possible.

**So the order is: write the tests first, run `sf lote start`, then implement.** Not the other
way round.

### If your envelope says "Qué hay que arreglar" instead of "Tus tareas"

Then this is a **correction batch**: the ㉑ found something and the feature came back to
`implementar`. Two things change, and nothing else does.

- **Your work is the finding**, not a task from the plan. The envelope has its id, the criterion
  it breaks, and the detail — that is all there is, because the plan was written before the
  finding existed.
- **`sf lote start` cannot name your tests**, so it asks for less: that the suite fails. It asks
  for it anyway. **Write the test that reproduces the finding first.** If the suite is already
  green, you have not reproduced anything — and a finding nobody can verify is a finding nobody
  can close.

The batch still ends in `sf done --msg "…"`, so the fix is one commit, like everything else.

The same relaxed gate applies on the short path of a bug (`tipo: bug`), where there is no
`tareas.json` at all: same rule, same reason.

## Step ⑲: Red → Green → Refactor

Compose **`sfx-tdd`**. The cycle is unchanged: one test, one implementation, one cycle, repeat.

**Write the tests the plan named, by name.** If a planned test turns out to be wrong or
impossible, that is a finding — say it in your closing message. Do not quietly rename it, and do
not quietly write a different one: `sf` is checking for those names.

### The plan's list is the floor, not the ceiling

**Write those, and then write every other test this batch needs.** The planner named the tests
it could see from outside; you are the one reading the code. Whatever you can see from here that
it could not, write it.

Ask the ⑲ version of the planner's question, now that the code is in front of you:

> **What could be wrong in this implementation and still pass the tests I have?**

Two timers instead of one. Something started and never stopped. State that survives a call it
should not. If you can answer that question, the answer is a test you are missing.

> **Why this is not optional.** The tool that would catch it — the ㉒ mutation run — happens
> after the whole feature is built, and every gap it finds sends the feature back here. A test
> you write now costs one line. The same test, found at the ㉒, costs a whole round.

**The planner's tests still come first and are not negotiable**, and the reason is not seniority:
they were chosen by someone who could not see your implementation, so they cannot be shaped to
fit it. Yours are the ones at risk of being written to pass. Add to the list; never trade one of
theirs for one of yours.

**All of them go in before `sf lote start`.** Not because more is better, but because `sf done`
compares the test files between the red and the green — a test written afterwards looks exactly
like a test that was softened to pass.

**Cover every criterion in your tasks' `satisface`.** They are in your envelope for this reason.

### Real code, not scaffolding

The one rule that matters more than the others here:

> **A method that returns a plausible value without doing the work is worse than a method that
> is not written yet.** The missing one fails loudly at the boundary; the mock passes the test,
> passes review, and is found in production.

If you cannot implement something in this batch, **say so and leave it unimplemented** — a
`TODO` that fails, an explicit error, anything that cannot be mistaken for working. Then name it
in your closing message.

## Step ⑳: Close the batch, and closing *is* committing

```bash
sf done --msg "feat: <what this batch did>"
```

**There is no `--msg`, no close.** The commit is not something to remember any more; it is the
only way to finish. That is how "the commit doesn't get made" and "commits badly grouped" stop
being detectable and become impossible — the batch was already the unit of grouping, decided
back at planning.

**`sf` makes the commit; you write the message.** Follow the constitution's `git.commit`
convention (`conventional`, by default) — the convention is a field there, not a rule you carry.
`sfx-github` is the reference for the mechanics if you need it.

Before it commits, `sf` checks two things:

- **the tests pass** — it runs them itself, it does not take your word;
- **the test files did not change between the red and the green.** If they did, it stops with
  *"the test files CHANGED between the red and the green"*.

> **That second check is not distrust of you specifically.** The cheapest way to make a failing
> test pass is to loosen the test, and it usually happens without anyone deciding to do it. The
> hash makes it visible instead of invisible.

If a test genuinely needs to change — the plan named it wrong — that is a real finding: say it,
and let it go back through the ⑰ rather than editing your way past the gate.

## When `sf done` refuses

Read what it says, fix that, run it again. **Every failed attempt increments a counter**, and
after enough of them `sf` stops the loop with **ME TRABÉ** and hands it to Javier, who can raise
the model (`sf model <name>`) or step in himself.

**Do not fight the gate.** If you are on your third attempt and considering weakening a test,
loosening an assertion, or deleting a case — stop and say so in plain words. Getting stuck is a
supported outcome; getting past the gate dishonestly is not.

## You do not run the loop

When your batch closes, you are done. **You die and the orchestrator calls `sf next` again** —
if a batch is left, it launches a fresh you with a fresh context. That is the whole point.

## Rules

- Tests first, then `sf lote start`, then code. Never the other way round.
- Write the tests the plan named, by their exact names — and every other one the batch needs.
- All the tests before `sf lote start`. One written later is indistinguishable from one softened.
- Never create the branch yourself. `sf lote start` does it.
- No conventions of your own. The constitution has them.
- No dependency outside `dependencias_aprobadas` without saying so.
- A mock that passes is worse than nothing. Fail loudly instead.
- Closing the batch is committing. No `--msg`, no close.
- Never edit a test to get past the gate. Report it.
- One batch. You do not continue to the next one.
