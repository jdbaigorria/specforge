# Investigation Methodology

## Data Flow Tracing

Start from the symptom. Work backward through the code:

1. **Entry point:** Where does the user action hit the code? (route handler, CLI command, event listener)
2. **Data transformation:** How does input get processed? Follow the data through each layer.
3. **Decision points:** Where does the code branch? (if/else, switch, try/catch, guard clauses)
4. **External calls:** Does it call a database, API, file system? What could go wrong there?
5. **Output:** Where does the response/result get assembled?

At each step, ask: "Is the actual behavior here consistent with the expected behavior?"
The point where they diverge is your root cause candidate.

## What to Read

Read in this order (stop when you have enough context):

1. **The failing code path** — entry point through to the error
2. **Tests for that path** — what's tested tells you what was expected
3. **Recent changes** — `git log --since="2 weeks ago" -- <affected-files>`
4. **Related code** — similar paths that DON'T fail (what's different?)
5. **Configuration** — env vars, config files, feature flags that might affect behavior

## Forming Hypotheses

A good hypothesis has three parts:

```
Hypothesis: "The JWT validation fails because the token includes a trailing newline
             from the environment variable, and the validator doesn't trim whitespace."

Evidence for: The error only happens in production (env vars from secrets manager).
              The same token works in dev (hardcoded in .env without newline).

Evidence against: None yet.

Verification: Read the token from the env var, check for whitespace characters.
              Check if the validator trims input.
```

A bad hypothesis: "Maybe something is wrong with the JWT library." (no evidence, no verification path)

### Common Root Cause Patterns

- **Type mismatch:** String where number expected, null where object expected
- **Boundary condition:** Empty array, zero, negative, very large, special characters (ñ, emoji)
- **State timing:** Race condition, stale cache, session expired, async not awaited
- **Environment delta:** Works locally, fails in CI/prod (env vars, paths, permissions, versions)
- **Regression:** Recent commit changed shared code that this path depends on
- **Configuration:** Feature flag, env var, default value wrong or missing
- **Dependency:** Library updated, API contract changed, service down

## Circuit Breaker

After 3 rejected hypotheses:

1. STOP investigating
2. Document what you tested and why each was rejected
3. Document what you still don't know
4. Recommend specific diagnostic actions:
   - "Add logging at X to capture Y"
   - "Reproduce with debugger attached, breakpoint at file:line"
   - "Check production logs for the specific request ID"
   - "Ask the author of commit ABC about the intent of the change"

The goal of an unresolved triage is to leave the next investigator (human or agent)
with a clear picture of what was already checked and what to try next. Never repeat
work that's already documented.
