---
name: sf-audit
description: >
  Run a project-wide adversarial audit. Cross-reference all features against the constitution,
  detect contradictions between features, find accumulated drift, verify invariants across the
  entire codebase. Use when the user says "sf-audit", "audit the project", "full review",
  "check everything against the constitution", "is the project consistent", "cross-check all
  features", "adversarial review", "project health check", or after a milestone when multiple
  features have been completed and it's time to verify the whole system holds together.
  Unlike sf-check (validates one feature), sf-audit validates the entire project.
---

# sf-audit

Adversarial audit of the entire project. Cross-reference everything. Find what nobody checked.

sf-check validates a feature against its own specs. sf-audit validates the project against
itself — constitution vs reality, feature vs feature, invariants vs codebase, specs vs code
across every completed feature.

**Always produces** `specforge/audits/{date}-{scope}.md`.

## When to Run

- After a milestone (3-5 features completed)
- Before a major release
- When the user feels "something is off" but can't pinpoint it
- Periodically as hygiene (every N features)
- After a major refactor to verify nothing drifted

## Step 1: Gather Everything

Read all of these (don't skip any that exist):

```
specforge/constitution.md          ← the law
specforge/features.json            ← all features and statuses
specforge/features/*/               ← active features (requirements, design, tasks)
specforge/archive/*/                ← completed features
specforge/history.md               ← project log + recurring issues
specforge/context/project.md                     ← stack and architecture
specforge/context/conventions.md                 ← coding standards
```

Scan the actual codebase for the audit checks below.

## Step 2: Constitution Compliance (project-wide)

For each principle in `constitution.md`:

1. Scan the codebase for violations — not just the latest feature, ALL code
2. Check if any archived feature's design contradicts a principle
3. Check if any principle has become stale (the project evolved past it)
4. Check if invariants (promoted via backprop) are respected everywhere

Produce a compliance table:

```markdown
| Principle | Status | Violations | Notes |
|-----------|--------|------------|-------|
| "Offline-first" | ✅ PASS | 0 | |
| "No raw print()" | ⚠️ DRIFT | 3 files | Introduced in add-export feature |
| INV1: "Error handling on interfaces" | ❌ FAIL | 5 endpoints | 2 features predate this invariant |
```

## Step 3: Cross-Feature Consistency

Check for contradictions and conflicts across features:

### Data model consistency
- Do different features define the same entity differently?
- Are there conflicting schemas or type definitions?
- Do naming conventions stay consistent across features?

### Interface conflicts
- Do any features expose contradictory API contracts?
- Are there duplicate endpoints or CLI commands with different behavior?
- Do error codes/messages stay consistent?

### Dependency conflicts
- Do features depend on incompatible versions of the same library?
- Are there circular dependencies between feature modules?

### Scope overlap
- Do any features implement the same requirement differently?
- Are there redundant implementations nobody noticed?

## Step 4: Spec-to-Code Drift

For each completed feature (in archive + active):

1. Read the requirements — do they still match what the code does?
2. Read the design — does the architecture still match?
3. Were there post-approval changes to code that were never reflected in specs?

This catches the common case where someone edits code directly after a feature
is archived, and the specs become stale without anyone noticing.

## Step 5: Convention Adherence

Read `specforge/context/conventions.md` and scan the codebase:

- Naming conventions consistent?
- Error handling patterns followed everywhere?
- Test structure consistent across all features?
- Import organization consistent?
- Documentation standards followed?

Flag areas where conventions drifted — often the earliest features don't
follow conventions that were established later.

## Step 6: Health Metrics

Compute project-level metrics:

```markdown
## Project Health

| Metric | Value | Assessment |
|--------|-------|------------|
| Features completed | {N} | |
| Features active | {N} | |
| Constitution principles | {N} | |
| Invariants (backprop) | {N} | |
| Principle violations found | {N} | {good/concerning/critical} |
| Cross-feature conflicts | {N} | |
| Spec-to-code drift | {N} features | |
| Convention violations | {N} files | |
| Test coverage gaps | {N} requirements | |
```

## Step 7: Produce Verdict

### HEALTHY
No critical violations. Minor drift noted. Project is consistent.

### NEEDS ATTENTION
Some principle violations or cross-feature inconsistencies.
Specific remediation steps listed.

### AT RISK
Significant drift from constitution. Multiple feature conflicts.
Accumulated debt threatens future development. Major remediation needed.

## Step 8: Write Artifact

Generate `specforge/audits/{date}-{scope}.md` using `templates/audit.tmpl.md`.

→ 🔴 **GATE**: Present audit results. User reviews findings.

For each finding, recommend:
- **Fix now:** violations that will compound if ignored
- **Fix next sprint:** issues that are real but not urgent
- **Track:** patterns that aren't violations yet but are trending that way
- **Accept:** deliberate deviations the user is aware of (document the decision)

## Step 9: Update History

Append to `specforge/history.md`:
```markdown
## [date] — Project audit
- Scope: {full | partial — what was checked}
- Verdict: {HEALTHY | NEEDS ATTENTION | AT RISK}
- Findings: {N critical}, {N major}, {N minor}
- Actions: {summary of recommended actions}
```

## Rules

- Be adversarial. Your job is to find problems, not confirm everything is fine.
- Be specific. "Code quality could improve" is useless. "3 API endpoints in
  add-export feature lack error handling, violating principle 4 and INV1" is useful.
- Check EVERYTHING against the constitution — including features that were approved
  before a principle or invariant existed.
- Don't audit what doesn't exist. If there's no constitution, say so and recommend
  running sf-init instead of inventing violations.
- Distinguish between violations (clear breach) and drift (gradual divergence).
  Violations need fixes. Drift needs decisions — maybe the constitution should update.
- If a principle is consistently violated, maybe it's the principle that's wrong.
  Flag it both ways: "either fix the code or amend the constitution."
