# Triage: {{bug_title}}

**Date**: {{date}}
**Status**: {{root-cause-found | unresolved | cannot-reproduce}}
**Severity**: {{critical | high | medium | low}}
**Fix scope**: {{trivial | small | medium | large}}

## Symptoms

**Observed**: {{what_happens}}
**Expected**: {{what_should_happen}}
**Reproduction**:
1. {{step_1}}
2. {{step_2}}

**Environment**: {{where_it_happens}}
**Frequency**: {{always | sometimes | once}}

## Investigation

### Data Flow Traced

{{data_flow_description}}
<!-- Brief path followed from symptom to root cause, with file:line refs -->

### Hypotheses Tested

| # | Hypothesis | Result | Evidence |
|---|-----------|--------|----------|
| 1 | {{hypothesis_1}} | {{confirmed | rejected}} | {{evidence_1}} |
| 2 | {{hypothesis_2}} | {{confirmed | rejected}} | {{evidence_2}} |

## Root Cause

<!-- Only if status = root-cause-found -->

**Location**: `{{file_path}}:{{line}}`
**What's wrong**: {{specific_code_behavior}}
**Why it's wrong**: {{why_this_produces_the_symptom}}

```{{language}}
{{problematic_code_snippet}}
```

### Why This Wasn't Caught

{{why_not_caught}}
<!-- No test coverage? Edge case? Regression? Environment delta? -->

## Fix Plan

### Approach

{{fix_strategy}}
<!-- 2-3 sentences. High-level strategy. -->

### Test First (write this — it should FAIL before the fix)

```{{language}}
{{failing_test}}
```

### Fix

{{fix_description}}
<!-- What to change, which file(s), what logic -->

```{{language}}
{{fix_snippet}}
```

### Regression Prevention

{{regression_test_or_check}}
<!-- What test or check would have caught this earlier -->

## Files Affected

- `{{file_1}}` — {{what_changes}}
- `{{file_2}}` — {{what_changes}}

## Risks

| Risk | Mitigation |
|------|------------|
| {{risk}} | {{mitigation}} |

## Next Step

{{next_step_recommendation}}
<!-- Trivial: "Apply directly"
     Small/Medium: "Run sf-propose fix-{slug}"
     Large: "Review fix plan, consider sf-propose --design-first fix-{slug}"
     Unresolved: "Add logging at X, reproduce with debugger, consult author of commit Y" -->

---
<!-- If status = unresolved -->
## Unresolved: What Remains

**Tested and rejected**: {{rejected_hypotheses_summary}}
**Still unknown**: {{what_information_is_missing}}
**Recommended diagnostic**: {{next_diagnostic_steps}}

---
*Investigated on {{date}}*
