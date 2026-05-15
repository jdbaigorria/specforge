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

## Pre-flight

1. Read `specforge/features.json` — verify status is `checking`
2. Read the feature's `requirements.md`
3. Read the feature's `tasks.md`
4. Read the feature's `design.md`
5. Read the feature's `progress/` logs
6. If `specforge/constitution.md` exists, read it — validate against principles

If status is not `checking`: "Feature `<name>` is in status `<status>`. Complete `sf-build <name>` first."

## Step 1: Traceability Analysis

For each requirement in `requirements.md`:
1. Is there at least one task that implements it? (check traceability table in `tasks.md`)
2. Is that task marked as complete? (check `[x]` in `tasks.md`)
3. Is there a test that validates it? (check test files in the codebase)

Build the traceability matrix:

```markdown
## Traceability Matrix

| Requirement | Task(s) | Implemented | Tested | Status |
|-------------|---------|-------------|--------|--------|
| R1          | T1, T2  | ✅          | ✅     | ✅ PASS |
| R2          | T3      | ✅          | ❌     | ⚠️ NO TEST |
| R3          | —       | ❌          | ❌     | ❌ MISSING |
```

## Step 2: Gap Analysis

Check for:
- **Missing implementations:** requirements without completed tasks
- **Missing tests:** implemented requirements without test coverage
- **Orphan code:** code that doesn't trace to any requirement (may be fine, but flag it)
- **Design deviations:** implementation that doesn't match `design.md`

## Step 3: Constitution Compliance (if constitution exists)

For each principle in `constitution.md`:
- Does the implementation respect this principle?
- Are there violations?

Example:
```
Principle: "Privacy by default — no data leaves the device"
Check: Does the feature send data to external services? → VIOLATION
```

## Step 4: Verdict

### APPROVE
All requirements implemented + tested. No constitution violations.
No critical gaps.

### APPROVE WITH NOTES
All requirements implemented. Minor gaps (e.g., missing edge case test).
Notes document what should be addressed later.

### REVISE
Significant gaps: missing implementations, constitution violations,
or untested critical paths.

## Step 5: Present Review

Generate `specforge/features/<name>/review.md` using `templates/review.tmpl.md`.

→ 🔴 **GATE**: Present the review to the user.
- User accepts APPROVE → proceed to archive
- User accepts REVISE → indicate what to fix, return to `sf-build`
- User overrides verdict → respect the override, log it

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

## Backprop (cross-feature learning)

Read `references/backprop.md` for the full pattern.

**Quick summary:** If the same type of issue appears in 3+ features:
- Promote it to a project invariant in `constitution.md`
- Example: "Missing error handling in API endpoints" found 3 times
  → Add principle: "Every API endpoint must have explicit error handling"
