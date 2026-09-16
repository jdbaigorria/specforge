---
name: sfx-triage
delegate: true
description: >
  Investigate a bug systematically. Find root cause. Hand over the failing test and the candidate
  fixes — WITHOUT recommending one, because choosing is a decision against a rubric and that is
  `sfx-decidir`'s. This is the SINGLE OWNER of the diagnosis method in SpecForge.
  Use when the user reports a bug, error, or unexpected behavior in any codebase — with or
  without SpecForge. Triggers: "/triage", "/triage <bug-description>", "there's a bug",
  "this is broken", "triage this", "investigate", "why is this failing", "debug this",
  "something is wrong with", "I'm getting an error", or any description of unexpected behavior
  paired with a desire to understand why. Works on any project — does not require SpecForge
  to be initialized.
---

# Triage

Investigate a bug. Find root cause. Produce fix plan with test strategy.

**Always produces a fix plan.** Composed by `sfp-backlog` it becomes the `us-#` (with `tipo: bug` and `relacionado_a`); standalone it goes wherever the user asks.

## The Iron Law

No fixes without investigation first. Do not patch, guess, or "try things".
Understand first, recommend second. The actual fix is a separate step — either
trivial (apply directly) or planned (`sf new "…"`, which enters the backlog as a bug).

## Step 1: Gather Symptoms

If the user gave a clear description → use it and proceed.
If vague or just "/triage" → ask:
- What did you expect to happen?
- What's happening instead?
- How do you reproduce it?
- How often? (always, sometimes, once)

Extract these five dimensions:
1. **Observed behavior** — what's actually happening
2. **Expected behavior** — what should happen
3. **Reproduction steps** — how to trigger it
4. **Environment** — where it happens (dev, staging, specific OS/browser)
5. **Frequency** — always, sometimes, once

## Step 2: Explore

Read `references/investigation.md` for detailed methodology.

**Quick summary:** Read relevant source code. Trace the data flow backward from
the symptom until you find where actual behavior diverges from expected. Check
recent git changes in affected areas — regressions are common.

**When the system has several layers, instrument the seams instead of guessing which one broke.**
Log what goes into each component and what comes out — the workflow, the build script, the
handler, the query — run it **once**, and read where the chain breaks. One run of evidence beats
four hypotheses, and it turns "which layer" from a guess into an observation.

Context to read if available (don't require any of these):
- `.docs/constitucion.md` — stack, architecture, conventions

- `.docs/features/*/spec-design.md` — expected behaviour, if SpecForge is initialized
- Source code in the affected area

## Step 3: Pattern Analysis

Before hypothesizing, find something analogous that *works* in this same
codebase. In a mature codebase this is the highest-yield move available, and
it turns hypothesis 1 from an intuition into an observation.

1. **Find the working example** — same layer, same kind of operation. Another
   endpoint on the same router, another handler of the same type, the same test
   against a different fixture.
2. **Diff it against the broken one.** Don't read them side by side; actually
   compare them.
3. **Isolate the minimal difference.** That difference is hypothesis 1, and it
   arrives with evidence attached.

If no analogous working case exists, say so explicitly and move on. The absence
is information too: it may mean this path never worked, in which case you're not
looking at a regression and the git-history angle from Step 2 is a dead end.

## Step 4: Hypothesize

Read `references/investigation.md` → Hypothesis section.

Form the top 2-3 possible root causes ranked by likelihood. Each needs:
- Evidence for (observations that support)
- Evidence against (observations that contradict)
- Verification step (specific check to confirm or reject)

## Step 5: Verify (Circuit Breaker)

Test each hypothesis starting with most likely:
- Execute the verification step (read code, check logs, trace execution)
- Confirmed → root cause found, go to Step 6
- Rejected → next hypothesis
- **HARD LIMIT: 3 hypotheses.** If all rejected → mark `unresolved`, document
  what was tested, recommend next steps. Do NOT force a conclusion.

### The other cut: three FAILED FIXES is the architecture

The 3-hypothesis limit above is not the only one, and they count different things:

```
3 hipótesis rechazadas   →  unresolved. No sabés la causa.
3 ARREGLOS que fallaron  →  no es la hipótesis, es la forma de la cosa.
```

The tell for the second: each fix uncovers a new problem somewhere else, or every fix needs "a
massive refactor". **That is not a failed hypothesis — it is a wrong shape**, and the move is a
conversation with the user, never a fourth attempt.

> Superpowers cuts at three failed fixes and hands it to the human; SpecForge's
> `TopeIntentos = 3` has said the same since 2026-09-03. Two projects that never read each other
> landed on the same number.

## Step 6: Write Artifact

Read `references/fix-plan.md` for fix strategy and TDD approach.

Use `templates/triage.tmpl.md`. Composed by `sfp-backlog`, this feeds the `us-#` instead of a file of its own.

The artifact covers:
- Symptoms (the 5 dimensions from Step 1)
- Investigation trace (data flow followed, with file:line references)
- Hypotheses tested (table: hypothesis, result, evidence)
- Root cause (if found: location, what's wrong, why)
- Fix plan (approach, failing test FIRST, then fix, regression prevention)
- Risks and affected files

→ 🔴 **GATE**: Present summary to user. If root cause found, recommend next step.

## Step 7: Hand over the candidates — without picking one

List the candidate fixes you found. **Do not rank them and do not recommend one.**

That is not modesty, it is the design: choosing a fix is a decision against a written rubric —
does it attack the cause or cover the symptom, is there a test that fails now and passes after,
does it change one thing, does it break anything green, does it fix where the problem is born —
and that rubric belongs to `sfx-decidir` (catalog ②). **Whoever investigated is the worst placed
to score their own hypothesis.**

`Call the Skill tool with "sfx-decidir"` with the diagnosis, or hand it to the composer that
called you (`sf-entrar-bug` does exactly this).

If `unresolved` → document what is still unknown and what would help (add logging, reproduce in
a debugger, ask the original author). **An `unresolved` is a `?`, not a red** — it says you could
not check, which is information, not failure.

If the project uses SpecForge and a pattern is emerging (the same kind of bug across several
features), note it — `sf-check`'s backprop may promote it to an invariant.

## Rules

- NEVER apply a fix during triage. Investigation only.
- NEVER pick the fix either. You hand over candidates; the rubric chooses.
- NEVER guess. Hypothesize with evidence, verify, conclude.
- Read actual code. Don't assume behavior from names.
- Trace backward from symptom. Don't start from the fix.
- Every hypothesis needs evidence. "Maybe X" is not a hypothesis.
- 3 hypotheses max. If all fail → unresolved. Stop.
- 3 failed FIXES is a different cut: that is the architecture, and it is a conversation.
- Several layers → instrument the seams and run once. Do not guess which one broke.
- Root cause found → write the failing test BEFORE the fix recommendation.
- Can't reproduce → mark `cannot-reproduce`, document what's needed.
- Check git history — regressions are the most common root cause.
- Keep artifact under 700 words. Evidence over prose.
