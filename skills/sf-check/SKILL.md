---
name: sf-check
description: >
  Validate feature implementation against specs and archive on approval. Use when the user
  says "sf-check", "check feature", "validate", "review implementation", or after completing
  a build. Verifies traceability (every requirement has implementation + test), runs gap
  analysis, and archives the feature on APPROVE. Also handles the backprop pattern where
  recurring issues become project invariants.
---

# sf-check

Validate implementation against specs. Archive on approval.

## Invocation Modes

### `sf-check` (no argument)
List features ready to check:
- Features in status `checking` → "Ready to validate"
- No features → "Nothing to check. Complete `sf-build <name>` first."

Ask: "Which feature do you want to validate?"

### `sf-check <name>`
Normal flow: validate the named feature (full, end-of-feature). Continue below.

### `sf-check --phase=<phase> <name>`
**Phase audit (shift-left).** A narrow, *fresh-context* quality check of ONE
just-completed phase against the constitution rules mapped to it — not the full
end-of-feature validation. Run at a phase gate (e.g. after design, before build).
This is the **same engine, scoped to one phase and run in a fresh subagent**. See
"Phase audit mode" below, then stop — do not run the full flow.

## Phase audit mode (`--phase`)

Opt-in, governed by `audit.phase` in `constitution.json`:
`"audit": { "phase": "off" | "nudge" | "block" }` (default `off`).

Read `audit.phase`. If `off` or absent → do nothing, return silently. Otherwise:

1. **Fetch the judge material** (deterministic — `sf` does the slicing):
   ```
   sf context for-judge --phase=<phase> --feature=<name>
   ```
   Returns the phase artifact + ONLY the principles/invariants whose `applies_to`
   includes `<phase>`. If it returns no rules, there's nothing to audit → return.
   If a parsimony principle (`P-min` / minimal-code) is in scope, also hand the
   subagent the rubric in `references/minimal-code.md`.

2. **Spawn a FRESH subagent** (Task tool) to judge. Freshness is the whole point:
   a clean context window escapes the degradation that hits a long main session.
   Hand it ONLY the for-judge output and this narrow prompt:

   > You are a strict spec auditor. Below is one artifact and the rules that apply
   > to its phase. For EACH rule, decide `pass` or `fail` and cite the exact element
   > of the artifact that justifies it — no citation means you cannot pass it. Judge
   > nothing outside these rules. Output ONLY JSON:
   > `{"phase":"<phase>","verdicts":[{"rule":"<id>","result":"pass|fail","citation":"..."}]}`
   >
   > <paste the `sf context for-judge` output here>

3. **Persist the verdict** (deterministic — `sf` records, never judges):
   ```
   echo '<subagent JSON>' | sf gate record-verdict --feature=<name> --phase=<phase>
   ```
   Exit `0` = all pass · exit `3` = at least one `fail` (recorded in `audit.json`).

4. **Act per `audit.phase`:**
   - `nudge`: surface the failing rules + citations to the user as a heads-up.
     **Do not block** — the human decides; the fail is already visible in `audit.json`.
   - `block` (advanced): treat a `fail` as a stop — do not proceed past the gate
     until the artifact is fixed and the audit re-run passes.
   - On all-pass: a one-line confirmation is enough.

The judge is **cooperative** (an LLM can err) → best-effort: it raises the floor on
quality, it doesn't guarantee it. The structural gates (PreToolUse) stay the
hermetic layer; this is the quality layer. Do **not** run the full check steps below.

## Pre-flight

1. Read `specforge/features.json` — verify status is `checking`
2. Read the feature's `requirements.md`
3. Read the feature's `tasks.md`
4. Read the feature's `design.md`
5. Read the feature's `progress/` logs
6. Read the feature's `failures.md` if it exists — known issues from build
7. If `specforge/constitution.md` exists, read it — validate against principles

If status is not `checking`:
"Feature `<name>` is in status `<status>`. Complete `sf-build <name>` first."

**Lite lane.** If `features.json` has `"lane": "lite"`, read `change.md` instead
of requirements/design/tasks. The check is **minimal but still real**: build the
trivial matrix (one requirement → one change → one test) and **still emit
`trace.json`** so the change stays inside drift detection (F33). Lite means
lighter ceremony, not lower integrity.

## Step 1: Traceability Analysis

This step is non-negotiable. Every check MUST produce the full matrix.
"All tests pass" is not a substitute. A check without this matrix is invalid.

For each requirement in `requirements.md`:
1. Is there a task that implements it? (check traceability table in `tasks.md`)
2. Is that task marked complete? (check `[x]` in `tasks.md`)
3. **Verify the implementation exists in code** — actually locate the code that
   implements this requirement. A checked-off task without corresponding code
   in the codebase is a gap.
4. Is there a test that validates it? (check test files in codebase)

Build the matrix:

```markdown
## Traceability Matrix

| Requirement | Task(s) | Implemented | Code Location | Tested | Status |
|-------------|---------|-------------|---------------|--------|--------|
| R1          | T1, T2  | ✅          | src/foo.py:42 | ✅     | ✅ PASS |
| R2          | T3      | ✅          | src/bar.py:10 | ❌     | ⚠️ NO TEST |
| R3          | —       | ❌          | —             | ❌     | ❌ MISSING |
```

**If any requirement has status ❌ MISSING, the verdict CANNOT be APPROVE.**

### Emit the structured matrix (`trace.json`)

The markdown matrix above is the human-readable contract. Also write the same
anchors in machine-readable form so drift can be checked later without
re-analyzing the repo (F33). Write `specforge/features/<name>/trace.json`:

```json
{
  "schema_version": "1.0",
  "feature": "<name>",
  "requirements": {
    "R1": { "code": ["src/foo.py:funcname"], "test": ["tests/test_foo.py::test_case"], "status": "ok" },
    "R2": { "code": ["src/bar.py:Klass.method"], "test": [], "status": "no-test" }
  }
}
```

- `code` entries are `path:symbol` — the exact anchor, **not a line number**
  (lines drift, symbols are stable).
- `test` entries are runnable test ids.
- `status`: `ok` | `no-test` | `missing`.

This file travels with the feature into the archive and is what drift detection
(`sf doctor --drift`) reads. Keep it consistent with the markdown matrix —
they describe the same thing.

## Step 2: Gap Analysis

Check for:
- **Missing implementations:** requirements without completed tasks
- **Missing tests:** implemented requirements without test coverage
- **Orphan code:** code that doesn't trace to any requirement (may be fine, but flag it)
- **Design deviations:** implementation that doesn't match `design.md`
- **Failed operations in build logs:** review each `progress/wave-<n>.md` for
  operations that failed during build. Any wave marked DIRTY is a red flag —
  verify the failed operation was resolved, not skipped.

### Cross-reference with failures.md

If `failures.md` exists, verify:
- Were any failures resolved during later waves? Update their status.
- Are unresolved failures consistent with gaps found in traceability?
- Do failure root causes point to spec issues (→ REVISE) or implementation
  issues (→ fix and re-check)?

## Step 3: Constitution Compliance (if constitution exists)

For each principle in `constitution.md`:
- Does the implementation respect this principle?
- Are there violations?

For a **parsimony** principle (`P-min` / minimal-code), apply the rubric in
`references/minimal-code.md`: over-engineering (reinvented stdlib, premature
abstraction, speculative generality, unjustified dependency) is a valid reason
for **REVISE**. `P-min` never overrides correctness or security.

## Step 4: Verdict

### APPROVE
All requirements implemented + tested. No constitution violations. No critical gaps.

### APPROVE WITH NOTES
All requirements implemented. Minor gaps documented for future.

### REVISE
Significant gaps. Produces structured failure analysis.

## Step 5: Present Review

Generate `specforge/features/<name>/review.md` using `templates/review.tmpl.md`.

### When verdict is REVISE

The review must include a structured **Failure Analysis** section:

```markdown
## Failure Analysis

### Why This Feature Didn't Pass

**Root cause category**: [spec gap | design mismatch | implementation error | external dependency]

### Specific Failures

#### F1: [title]
- **Requirement**: R[n]
- **What was expected**: [from spec]
- **What happened**: [actual behavior]
- **Why**: [root cause]
- **Fix category**: [amend spec | change design | fix implementation | needs research]
- **Recommended action**: [specific, actionable step]

#### F2: ...

### Recommended Path Forward

1. [Most impactful fix first]
2. [Next fix]
3. [If spec needs amending: "Edit requirements.md → resync will cascade to design and tasks"]
```

This gives the user a clear map of what went wrong, why, and exactly what to do
about it — not just "gaps found, go back to build."

→ 🔴 **GATE**: Present the review to the user.
- User accepts APPROVE → seal the verdict gate, then proceed to archive:
  ```bash
  sf gate approve --feature=<name> --phase=verdict
  ```
  This appends the gate AND seals a content hash of `review.json` (the CLI
  computes it over the real file — don't hand-write the entry).
- User accepts REVISE → follow the recommended path (no verdict gate sealed)
- User overrides verdict → respect the override, log it with a hand-written gate
  entry carrying the override in `comment`

**Team mode (F35).** When the feature is on a `feature/<slug>` branch with a PR,
this verdict gate **maps to the PR approval** — the reviewer approves code and
spec together. Don't run a separate verdict gate *and* a PR review; the PR
approval satisfies it, and `archive` corresponds to the merge. See
`skills/sfx-github/references/specforge-integration.md`.

## Step 6: Archive (on APPROVE)

Read `references/archive.md` for detailed procedure.

**Quick summary:**
1. Copy feature folder to `specforge/archive/<date>-<name>/`
2. Update `features.json` status to `done`
3. Append to `specforge/history.md`:
   ```
   ## [date] — Feature completed: <name>
   - Verdict: [APPROVE / APPROVE WITH NOTES]
   - Requirements: [count] | Tasks: [count] | Coverage: [%]
   ```
4. Update `specforge/roadmap.md` if it exists (mark feature as completed)

## Step 7: Backprop (cross-feature learning)

Read `references/backprop.md` for the full pattern.

If the same type of issue appears in 3+ features → promote to invariant
in `constitution.md`.

Check `failures.md` from this and previous features for recurring patterns.

## Rules

- Be adversarial during check. Your job is to find problems.
- REVISE must include actionable failure analysis — not just "stuff is missing."
- Cross-reference failures.md with gap analysis — they should tell the same story.
- Don't be lenient with ACs. Partial ≠ pass.
- Don't invent gaps. Style differences aren't gaps.
- Don't use REJECT as escape. If fixable with REVISE, use REVISE.
- Don't skip backprop. Cumulative value.
