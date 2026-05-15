---
name: triage
description: >
  Investigate a bug systematically. Find root cause. Produce a fix plan with test strategy.
  Use when the user reports a bug, error, or unexpected behavior in any codebase — with or
  without SpecForge. Triggers: "/triage", "/triage <bug-description>", "there's a bug",
  "this is broken", "triage this", "investigate", "why is this failing", "debug this",
  "something is wrong with", "I'm getting an error", or any description of unexpected behavior
  paired with a desire to understand why. Works on any project — does not require specforge/
  to be initialized.
---

# Triage

Investigate a bug. Find root cause. Produce fix plan with test strategy.

**Always produces an artifact** at `.ai/triages/{slug}.md`.

## The Iron Law

No fixes without investigation first. Do not patch, guess, or "try things".
Understand first, recommend second. The actual fix is a separate step — either
trivial (apply directly) or planned (`sf-propose fix-{slug}`).

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

Context to read if available (don't require any of these):
- `.ai/project.md` — stack, architecture
- `.ai/conventions.md` — patterns
- `specforge/` specs — expected behavior per spec (if SpecForge is initialized)
- Source code in the affected area

## Step 3: Hypothesize

Read `references/investigation.md` → Hypothesis section.

Form the top 2-3 possible root causes ranked by likelihood. Each needs:
- Evidence for (observations that support)
- Evidence against (observations that contradict)
- Verification step (specific check to confirm or reject)

## Step 4: Verify (Circuit Breaker)

Test each hypothesis starting with most likely:
- Execute the verification step (read code, check logs, trace execution)
- Confirmed → root cause found, go to Step 5
- Rejected → next hypothesis
- **HARD LIMIT: 3 hypotheses.** If all rejected → mark `unresolved`, document
  what was tested, recommend next steps. Do NOT force a conclusion.

## Step 5: Write Artifact

Read `references/fix-plan.md` for fix strategy and TDD approach.

Generate `.ai/triages/{slug}.md` using `templates/triage.tmpl.md`.

The artifact covers:
- Symptoms (the 5 dimensions from Step 1)
- Investigation trace (data flow followed, with file:line references)
- Hypotheses tested (table: hypothesis, result, evidence)
- Root cause (if found: location, what's wrong, why)
- Fix plan (approach, failing test FIRST, then fix, regression prevention)
- Risks and affected files

→ 🔴 **GATE**: Present summary to user. If root cause found, recommend next step.

## Step 6: Next Steps

Based on fix scope:
- **Trivial** (<10 lines, no architectural change) → "Apply directly, here's the fix"
- **Small/Medium** → "Run `sf-propose fix-{slug}` to plan the implementation"
- **Large** (systemic issue) → "This may need a design change. Review the fix plan first"

If `unresolved` → document what's still unknown and what would help
(add logging, reproduce in debugger, consult original author).

If the project uses SpecForge and a pattern is emerging (same type of bug in
multiple features), note it — sf-check's backprop may promote it to invariant.

## Rules

- NEVER apply a fix during triage. Investigation only.
- NEVER guess. Hypothesize with evidence, verify, conclude.
- Read actual code. Don't assume behavior from names.
- Trace backward from symptom. Don't start from the fix.
- Every hypothesis needs evidence. "Maybe X" is not a hypothesis.
- 3 hypotheses max. If all fail → unresolved. Stop.
- Root cause found → write the failing test BEFORE the fix recommendation.
- Can't reproduce → mark `cannot-reproduce`, document what's needed.
- Check git history — regressions are the most common root cause.
- Keep artifact under 700 words. Evidence over prose.
