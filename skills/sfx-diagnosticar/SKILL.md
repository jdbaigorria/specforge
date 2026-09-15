---
name: sfx-diagnosticar
description: >
  The diagnosis primitive — find the ROOT CAUSE of a bug before anybody proposes a fix. Use for any
  technical failure: a bug report, a test that went red, a build that broke, unexpected behaviour,
  a review finding that turns out to be a real defect. This is the SINGLE OWNER of the diagnosis
  method in SpecForge. Triggers: "this is broken", "why does this fail", "the test went red",
  "debug this", "find the bug", or any composing skill calling it — above all `sf-entrar-bug`.
  NOT the fix: it ends at a named cause and a failing test.
---

# sfx-diagnosticar

## The iron law

```
NO FIX WITHOUT A ROOT CAUSE FIRST
```

Fixing the symptom is failing. If you have not finished Phase 1, you do not get to propose a fix —
and "the issue looks simple" is not an exemption, because simple bugs have root causes too.

**This whole skill runs without the user.** Reading the error, reproducing it, looking at what
changed, tracing the bad value back: none of it needs a decision. Do not ask; go look.

## Phase 1 — the cause

1. **Read the error whole.** The complete stack trace, the line numbers, the error codes. Not the
   first line. The answer is often literally in there.

2. **Reproduce it.** Exact steps, and does it happen every time? **If you cannot reproduce it, that
   is a result, not a failure** — say so as `?` and gather more data. Do not guess a cause for
   something you never saw happen.

3. **Look at what changed.** `git log`, `git diff`, recent commits, new dependencies, config. A
   bug that appeared has a commit where it appeared.

4. **When the system has layers, instrument the seams.** Log what goes into each component and
   what comes out, run it **once**, and read where it breaks. One run of evidence beats four
   hypotheses.

5. **Trace the bad value backwards.** Where did it come from? Who called this with it? Keep going
   up until you reach the source. **Fix at the source, not where the error surfaced** — that is the
   difference between a fix and a bandage with a nicer name.

Need a fact from outside — does this library really behave like that, is this a known issue?
`Call the Skill tool with "sfx-buscar"`. Level 0 first, provenance on every claim.

## Phase 2 — compare against what works

1. **Find something similar that works** in this same repo.
2. **Read the working one completely.** Not skimmed. Every line.
3. **List every difference**, however small. Do not decide in advance that one "cannot matter" —
   that judgement is exactly what you do not have yet.
4. **Check the dependencies**: what config, what environment, what assumptions.

## Phase 3 — one hypothesis at a time

1. **State it out loud:** *"I think X is the cause, because Y."* Written, specific.
2. **Test it with the smallest possible change.** One variable.
3. **Confirmed → Phase 4. Not confirmed → a NEW hypothesis.** Never stack a second fix on top of
   the first.
4. **If you do not know, say you do not know.** Do not narrate confidence you do not have.

## Phase 4 — the failing test, and the stop

1. **Write the test that fails today.** The simplest reproduction there is, automated if the repo
   has a framework. `Call the Skill tool with "sfx-tdd"` for the red → green cycle. **You must
   have it before anybody fixes anything**: a bug without a failing test comes back.
2. **Stop here.** Choosing *which* fix is not this skill's job — that is a decision against a
   rubric, and it belongs to `sfx-decidir`. You hand over the cause, the failing test, and the
   candidate fixes you found.

### The three-strike rule

Count the fixes attempted. **At three failures, stop and question the architecture.** The tell:
each fix uncovers a new problem somewhere else, or every fix needs "a massive refactor".

That is not a wrong hypothesis — it is a wrong shape, and it is a conversation with the user, not
a fourth attempt.

> Superpowers arrived at the same number and the same exit independently, and SpecForge's
> `TopeIntentos = 3` has said it since 2026-09-03. Three sources, one number.

## What it leaves

The composing skill names the file. If it did not, write `diagnostico.md`:

```yaml
---
reproducido: si | no
causa: "<one line, with file:line>"
test_que_falla: "<the test's name>"
arreglos_candidatos: 2
intentos: 0
---
```

Body: the error as it came, the steps to reproduce, what changed, the differences against the
working example, the hypothesis and how it was confirmed, and the candidate fixes — **with no
recommendation**. That is `sfx-decidir`'s job and it has a rubric for it.

Every load-bearing claim carries provenance: `read` (`file:line`), `ran` (the command and its
output), `retrieved` (the link), or `model-prior` — unverified, which is a `?`.

## Traps — all of these mean go back to Phase 1

| what you catch yourself thinking | what it really is |
|---|---|
| "quick fix now, investigate later" | the first fix sets the pattern |
| "it's probably X, let me change it" | seeing the symptom is not understanding the cause |
| "several changes at once saves time" | you cannot tell which one worked |
| "the test after, once the fix works" | a fix with no test does not stick |
| "it's an emergency, no time for process" | guess-and-check is SLOWER, always |
| "one more attempt" (after two) | three failures is the architecture, not the hypothesis |
| "it's flaky" | that is a conclusion, and you have not earned it |

## When there really is no cause

If the investigation genuinely lands on environmental, timing-dependent or external: you finished
the process. Write down what you looked at, handle it explicitly (retry, timeout, a real error
message), and leave something logged for next time.

**But say it as a `?`, not as a green.** And be honest about the base rate: most "no root cause"
is an unfinished investigation.

## Rules

- No fix without a cause. No exceptions for "simple".
- It runs without the user. Facts are yours; do not ask what you can look up.
- Could not reproduce is a result, and it is a `?`.
- One hypothesis at a time, the smallest possible change.
- The failing test comes before the fix.
- Fix at the source, never where the error surfaced.
- Three failed attempts is an architecture conversation, not a fourth attempt.
- This skill ends at the cause. It does not choose the fix.

## Why this exists

Translated from `obra/superpowers/skills/systematic-debugging`, read 2026-09-15. Matt Pocock has
`diagnosing-bugs` for the same thing. **SpecForge had the lane and not the method**: `tipo: bug`
skips planning and lands straight in a build skill, against a plan that in this lane was never
written — and the only nearby skill, `sfx-triage`, classifies the ticket without diagnosing the
defect (`vecinos-metodos.md` §4.1, the biggest gap of the three repos compared).
