# Project Audit: {{scope}}

**Date**: {{date}}
**Verdict**: {{HEALTHY | NEEDS_ATTENTION | AT_RISK}}
**Features audited**: {{count}} ({{completed}} completed, {{active}} active)

## Summary

{{2_3_sentence_verdict_summary}}

## Project Health

| Metric | Value | Assessment |
|--------|-------|------------|
| Features completed | {{n}} | |
| Features active | {{n}} | |
| Constitution principles | {{n}} | |
| Invariants (backprop) | {{n}} | |
| Principle violations | {{n}} | {{good / concerning / critical}} |
| Cross-feature conflicts | {{n}} | |
| Spec-to-code drift | {{n}} features | |
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

| Feature | Spec Status | Drift Found | Detail |
|---------|-------------|-------------|--------|
| {{feature}} | {{current / stale}} | {{yes / no}} | {{what_drifted}} |

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
