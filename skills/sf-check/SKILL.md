---
name: sf-check
description: >
  Validate feature implementation against specs and archive on approval. Use when the user
  says "sf-check", "check feature", "validate", "review implementation", or after completing
  a build. The verdict is gated on machine-verified evidence, not on a reading — every step
  in the skill body is load-bearing. Also covers backprop, where recurring issues become
  project invariants.
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
   includes `<phase>`, plus a `rubrics` array naming the **built-in rubrics** that
   apply to this phase.

   For every name in `rubrics`, hand the subagent `references/<name>.md` as well.
   The CLI decides which apply — do not decide it yourself:
   - `requirement-quality` — returned on **every** `requirements` phase, with or
     without a constitution. Requirement quality is universal, so it ships from
     the factory; a thin constitution must not yield an audit that audits nothing.
   - `minimal-code` — returned only when a parsimony principle (`P-min`) is in
     scope for this phase.

   Return early ONLY if `principles`, `invariants`, `domain_rules` **and**
   `rubrics` are all empty — that is what "nothing to audit" means.

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

1. Run `sf status` — verify the feature's status is `checking`
2. Read the feature's `requirements.md`
3. Read the feature's `tasks.md`
4. Read the feature's `design.md`
5. Read the feature's `progress/` logs
6. Read the feature's `failures.md` if it exists — known issues from build
7. If `specforge/constitution.md` exists, read it — validate against principles

If status is not `checking`:
"Feature `<name>` is in status `<status>`. Complete `sf-build <name>` first."

**Lite lane.** If the feature's `feature.json` has `"lane": "lite"` (see `sf status`), read `change.md` instead
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

### Complete and audit the structured matrix (`trace.json`)

`trace.json` is **not generated from scratch here** — `sf-build` already wrote it
during the build as the verification contract (each requirement with its code +
test anchors). sf-check's job is to **complete and audit** it: fill any gaps,
correct anchors that drifted during the build, set the final `status` per
requirement, and confirm it matches the markdown matrix above. The build declares;
check seals.

The markdown matrix above is the human-readable contract. Keep the same anchors in
machine-readable form so drift can be checked later without re-analyzing the repo
(F33). The file lives at `specforge/features/<name>/trace.json`:

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

#### If the requirement's `acceptance` criteria have ids, anchor per criterion

Open `requirements.json`. When a requirement's `acceptance` entries carry ids
(`R5.1`, `R5.2`, …), that requirement runs under the **v2 verification
contract**: every criterion needs its own test, anchored under `scenarios`.

```json
"R5": {
  "code": ["src/order.go:CreateOrder"],
  "scenarios": {
    "R5.1": { "test": ["order_test.go:TestCreateOrder_Persists"] },
    "R5.2": { "test": ["order_test.go:TestCreateOrder_EmptyCart"] }
  },
  "test": [],
  "status": "ok"
}
```

Why the extra level: without it, a requirement with five acceptance criteria
seals green on **one** test of the happy path, and the other four are never
touched. No bad faith required — the rule was just measuring the wrong thing.
The gate now rejects naming `R5.2`, not `R5`, so you know which case is missing
instead of re-reading five criteria to find out.

Two rules that are easy to get wrong:

- **A requirement-level `test` does not satisfy a criterion.** That list is for
  tests covering the requirement *transversally*. A test that proves "something
  about R5" does not prove `R5.2`.
- **Never renumber criterion ids to close a gap.** Anchors point at ids, so
  reusing a freed id silently re-points a test at a different case. `R5.1, R5.3`
  with no `R5.2` is correct and means a criterion was retired.

Requirements whose `acceptance` is still a plain list of strings keep the old
rule (one test per requirement) — nothing to do for those.

**Persist it through the CLI — never write `trace.json` by hand.** It is
SpecForge state: the hook denies a direct `Write`/`Edit`, and the verdict gate
hashes it. Write the draft with your `Write` tool to the feature's `drafts/`
dir (the one writable corner of `specforge/`), then promote it — `sf save`
validates, writes the canonical file, and removes the draft:

```bash
# 1) Write specforge/features/<name>/drafts/trace.json with the Write tool
# 2) Promote it:
sf save trace --feature=<name> --from=drafts/trace.json
```

(For small payloads, piping still works: `echo '<json>' | sf save trace --feature=<name> --json -`.)

Then confirm it holds up against the real repo — this is what the verdict gate
checks, so do it now, not after:

```bash
sf trace verify --feature=<name>   # every requirement's code AND named test must resolve
```

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

Before presenting, run `sf gate show --feature=<name>` and lead with that
evidence view (trace coverage, test freshness, judge verdicts, commits since the
last gate) — the human decides from evidence, not from re-reading the whole
artifact. Gate fatigue is how governance degrades into theater.

→ 🔴 **GATE**: Present the review to the user.
- User accepts APPROVE → seal the verdict gate, then proceed to archive:
  ```bash
  sf check run --feature=<name>        # run the suite; records a fresh, real exit code
  sf gate approve --feature=<name> --phase=verdict
  ```
  `sf gate approve --phase=verdict` **refuses** unless `sf trace verify` is clean,
  every requirement names a test that resolves, AND `sf check run` recorded a
  fresh green result (the code hash at test time matches the current code). It
  seals a content hash of `review.json` over the real file — you cannot hand-write
  the entry (features.json is protected). If it refuses, the gaps are real: fix
  them and re-run, don't try to bypass.
- User accepts REVISE → follow the recommended path (no verdict gate sealed)
- User overrides verdict → an override still goes through `sf gate approve
  --phase=verdict --comment="override: <reason>"`. It will only seal once the
  preconditions hold — there is no hand-written bypass.

**Team mode (F35).** When the feature is on a `feature/<slug>` branch with a PR,
this verdict gate **maps to the PR approval** — the reviewer approves code and
spec together. Don't run a separate verdict gate *and* a PR review; the PR
approval satisfies it, and `archive` corresponds to the merge. See
`skills/sfx-github/references/specforge-integration.md`.

## Step 6: Archive (on APPROVE)

Read `references/archive.md` for detailed procedure.

**Quick summary:**
1. `sf feature archive --feature=<name>` — this copies the feature folder to
   `specforge/archive/<date>-<name>/` AND sets status `done` in one step. It
   refuses unless the verdict gate is sealed, so do Step 5 first. (Never copy the
   folder or set `done` by hand — features.json is protected.)
2. Append to `specforge/history.md`:
   ```
   ## [date] — Feature completed: <name>
   - Verdict: [APPROVE / APPROVE WITH NOTES]
   - Requirements: [count] | Tasks: [count] | Coverage: [%]
   ```
3. Update `specforge/roadmap.md` if it exists (mark feature as completed)

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

## Rationalizations

The Rules above are imperatives. These are the specific ways an agent talks
itself out of them.

| The excuse | Why it doesn't hold |
|---|---|
| "The AC is *substantially* covered" | Substantially is not a verdict. Either a test names the acceptance criterion or it doesn't — and "Don't be lenient with ACs" means exactly this case, not a hypothetical one. |
| "The test covers the happy path, the edge case is obvious" | Obvious to whom, at what hour? The edge case is where the defect lives. An untested edge is a gap, and gaps go in the review. |
| "I'll fix that in the next feature" | Then it's an unresolved gap in *this* review, recorded as such. A promise made during check is not evidence, and nothing carries it forward. |
| "The trace is basically right, one anchor is stale" | A stale anchor is drift and `sf trace verify` will say so. Fixing it now costs one edit; shipping it moves the cost to whoever runs `sf doctor` next month and can't tell which anchors to trust. |
| "It's REJECT-worthy but REVISE is less disruptive" | Backwards: REJECT is the escape hatch, REVISE is the workhorse. If it's fixable, REVISE — and say precisely what to fix. |
