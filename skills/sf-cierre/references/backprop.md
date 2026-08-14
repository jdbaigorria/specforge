# Backprop — when a lesson should become a rule

## The pattern

The same kind of problem showing up in several features is not several mistakes. It is **one
missing rule**, and every feature after this one will hit it too.

## The rule: 3× → promote

If the same issue type appears in **three or more features**:

1. Name the pattern — the shared shape, not the three symptoms.
2. Write it as a rule someone could **violate**. "Be careful with errors" cannot be violated;
   "every external interface handles its errors explicitly" can.
3. **Propose it** for `## Reglas de trabajo` in `.docs/constitucion.md`, in your closing message.
4. **You do not write it.** Javier decides at the ⏸ — a promoted rule changes every feature that
   follows, so it is his call, not yours.

## Examples

**Missing error handling**
- f-1: the endpoint returns 500 on invalid input
- f-3: the CLI panics on malformed JSON
- f-4: the webhook handler swallows a timeout

→ **Rule:** every external interface handles its errors explicitly, with a message that says
what to do.

**Missing input validation**
- f-2: no length check on the username
- f-3: no type check on config values
- f-5: no range check on the pagination params

→ **Rule:** user input is validated at the boundary, before it reaches any logic.

**Hardcoded configuration**
- f-1: the API URL is hardcoded
- f-2: the timeout is hardcoded
- f-4: the flag is hardcoded

→ **Rule:** configuration lives in the environment or a config file, never in the code.

## How you count to three

**There is no counter file, and there deliberately is not one.** The evidence is the archived
journals, which `sf context` already puts in your envelope. Read them and see whether what you
just hit is already written there twice.

> A counter maintained by hand is a counter that goes stale the first time somebody forgets. The
> journals are written anyway, so they are the source of truth for this.

## What not to promote

- **Specific bugs.** A rule is a pattern; a bug is an instance.
- **Anything the machine already enforces.** A rule that restates a gate will drift from it.
- **Anything you have only seen twice.** Two is a coincidence you are pattern-matching on. Note
  it in the journal and let the third occurrence make the case.

## After a promotion

Say so plainly: the features already closed were built **without** this rule and may not follow
it. That is not a crisis and not a reason to reopen them — it is context for whoever finds the
inconsistency later.
