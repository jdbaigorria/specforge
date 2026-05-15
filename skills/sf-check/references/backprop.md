# Backprop: Cross-Feature Learning

## Pattern

When the same type of issue appears across multiple features,
it indicates a systemic gap rather than a one-off mistake.

## Rule: 3x → Invariant

If an issue type is found in **3 or more features**:
1. Identify the pattern
2. Formulate a principle or check
3. Add it to `specforge/constitution.md` as a new principle
4. Future `sdd-check` runs validate against this invariant

## Examples

### Issue: Missing error handling
- Feature A: API endpoint returns 500 on invalid input
- Feature B: CLI crashes on malformed JSON
- Feature C: Webhook handler doesn't catch timeout
→ **Invariant:** "Every external interface must have explicit error handling
  with user-friendly messages."

### Issue: Missing input validation  
- Feature A: No length check on username
- Feature B: No type check on config values
- Feature C: No range check on pagination params
→ **Invariant:** "All user input must be validated at the boundary
  before reaching business logic."

### Issue: Hardcoded configuration
- Feature A: API URL hardcoded
- Feature B: Timeout value hardcoded
- Feature C: Feature flag hardcoded
→ **Invariant:** "Configuration values must be externalized
  (environment variables or config file)."

## Tracking

Maintain a counter in `specforge/history.md`:

```markdown
## Recurring Issues
| Issue Pattern | Occurrences | Features | Status |
|---------------|-------------|----------|--------|
| Missing error handling | 2 | auth, cli | tracking |
| Hardcoded config | 3 | auth, cli, api | → INVARIANT |
```

When count reaches 3, promote and update status.

## Important

- Only promote patterns, not specific bugs
- The invariant should be actionable and testable
- After promotion, re-check recent features against the new invariant
  (they might need fixes too)
