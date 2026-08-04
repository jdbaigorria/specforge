# Project Audit: {{scope}}

**Date**: {{date}}
**Verdict**: {{HEALTHY | NEEDS_ATTENTION | AT_RISK}}
**Features audited**: {{count}} ({{completed}} completed, {{active}} active)

## Summary

{{2_3_sentence_verdict_summary}}

## Project Health

### Computed — transcribed from the CLI, not estimated

<!-- sf status --json · sf doctor --drift --run-tests --json · sf coverage --json · sf verify --json
     Two runs over an unchanged repo must produce this table identically. -->

| Metric | Value | Source | Assessment |
|--------|-------|--------|------------|
| Features completed | {{n}} | `sf status --json` | |
| Features active | {{n}} | `sf status --json` | |
| Not implemented | {{n}} requirements | `sf doctor --json` | |
| Implemented differently | {{n | undetermined}} | `sf doctor --json` | |
| Unverified | {{n}} requirements | `sf doctor --json` | |
| Out of spec | {{n}} files | `sf coverage --json` | |
| Spec coverage | {{pct}}% ({{anchored}}/{{total}}) | `sf coverage --json` | |
| Integrity checks | {{n_ok}}/{{n_total}} passing | `sf verify --json` | |
| Constitution principles | {{n}} | `constitution.json` | |
| Invariants (backprop) | {{n}} | `constitution.json` | |

<!-- "Implemented differently" is `undetermined` unless --run-tests was passed.
     Write `undetermined`, never 0 — 0 asserts a check that didn't happen. -->

### Judgment — an adversarial reading, not a measurement

<!-- These two can legitimately differ between runs. That's why they're separate:
     collapsing them into the table above lends them authority they don't have. -->

| Finding | Count | Assessment |
|---------|-------|------------|
| Principle violations | {{n}} | {{good / concerning / critical}} |
| Cross-feature conflicts | {{n}} | |
| Convention violations | {{n}} files | |

## Constitution Compliance

| Principle | Status | Violations | Location |
|-----------|--------|------------|----------|
| {{principle}} | {{PASS / DRIFT / FAIL}} | {{count}} | {{files_or_features}} |
| {{invariant}} | {{PASS / DRIFT / FAIL}} | {{count}} | {{files_or_features}} |

### Critical Violations

{{critical_violations_detail}}
<!-- Each violation: what principle, where exactly, what the code does vs what it should do -->

### Drift (not violations yet, but trending)

{{drift_observations}}

## Cross-Feature Consistency

### Conflicts Found

| Conflict | Features | Impact | Recommendation |
|----------|----------|--------|----------------|
| {{conflict}} | {{feature_a}} vs {{feature_b}} | {{impact}} | {{fix}} |

### Data Model Consistency

{{data_model_findings}}

### Interface Consistency

{{interface_findings}}

## Spec-to-Code Drift

<!-- One row per finding from `sf doctor --drift --run-tests --json`, by category. -->

| Feature | Requirement | Category | Detail |
|---------|-------------|----------|--------|
| {{feature}} | {{R#}} | {{not implemented / implemented differently / unverified}} | {{reason}} |

### Which side was wrong

<!-- Only for "implemented differently". BOTH routes, no default, one row each —
     never a batch recommendation. -->

| Feature / Req | (a) spec went stale | (b) spec was right, code has a defect | Recommendation |
|---|---|---|---|
| {{feature}} / {{R#}} | {{what amending would say}} | {{expected vs observed}} | {{a or b + why}} |

## Out of Spec — code no requirement governs

<!-- From `sf coverage --json` → unanchored. Two outcomes per file, nothing else.
     Excluding raises the percentage, so every exclusion carries a reason. -->

| File | Disposition | Reason |
|------|-------------|--------|
| {{path}} | {{adopt / exclude}} | {{why}} |

## Convention Adherence

| Convention | Status | Violations | Files |
|------------|--------|------------|-------|
| {{convention}} | {{consistent / inconsistent}} | {{count}} | {{locations}} |

## Findings by Priority

### Fix Now
<!-- Violations that compound if ignored -->
- [ ] {{finding}} — {{location}} — {{why_urgent}}

### Fix Next Sprint
<!-- Real issues, not urgent -->
- [ ] {{finding}} — {{location}}

### Track
<!-- Patterns trending toward problems -->
- [ ] {{pattern}} — {{current_count}}/3 threshold

### Accept (document decision)
<!-- Deliberate deviations -->
- [ ] {{deviation}} — {{rationale_for_accepting}}

## Recommendations

{{overall_recommendations}}
<!-- Actions to take. Specific, actionable.
     "Consider amending principle X — it's violated in 8/10 features,
      suggesting the principle doesn't match how the project evolved."
     "Run sf-propose fix-error-handling to address INV1 violations." -->

---
*Audited on {{date}}*
