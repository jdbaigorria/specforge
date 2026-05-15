# Clarify: Post-Generation Refinement

## When Triggered

User asks to refine, review, or clarify after specs are generated:
- "Review the requirements for gaps"
- "Check for ambiguities"
- "What did I miss?"
- "Analyze the specs"
- Direct edits to an artefact followed by "update the rest"

## Refinement Procedure

### 1. Ambiguity Scan

Review each requirement for:
- **Vague language:** "should handle errors appropriately" → what does "appropriately" mean?
- **Missing defaults:** "user can set priority" → what's the default?
- **Undefined boundaries:** "supports large files" → how large?
- **Implicit assumptions:** "sends notification" → via what channel?

### 2. Completeness Check

- Does every user story have acceptance criteria?
- Are all edge cases covered? (empty input, concurrent access, failure scenarios)
- Are error states defined? (not just happy path)
- Do requirements cover non-functional aspects? (performance, security)

### 3. Consistency Check

- Do requirements contradict each other?
- Does the design support all requirements?
- Do tasks cover all requirements? (traceability matrix)
- Are naming conventions consistent across artefacts?

### 4. Present Findings

Format as a checklist:
```markdown
## Refinement Report

### Ambiguities Found
- [ ] R3: "handles errors appropriately" — define error handling strategy
- [ ] R5: default value for priority not specified

### Missing Coverage
- [ ] No requirement for authentication timeout
- [ ] No edge case for concurrent edits

### Inconsistencies
- [ ] R2 says "email notification" but design.md says "push notification"

### Suggestions
- Consider adding rate limiting (not currently specified)
```

→ 🔴 **GATE**: User reviews findings and decides which to address.

### 5. Apply Changes

For each accepted finding:
1. Update the relevant artefact
2. Check if downstream artefacts need resync
3. Present updated artefact for approval
