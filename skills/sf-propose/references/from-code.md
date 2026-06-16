# From-Code: Reverse-Engineer Specs from Existing Code

## When to Use

- Brownfield project with code but no specs
- Need a spec baseline before adding new features
- Want to document what already exists for traceability

## Prerequisites

- `specforge/context/project.md` and `specforge/context/conventions.md` must exist (run `sf-init` first)
- The codebase must be readable and reasonably organized

## Flow

### Step 1: Analyze Code

For each module/component in the codebase:
1. Read entry points and public interfaces
2. Identify user-facing behaviors (CLI commands, API endpoints, UI actions)
3. Map data models and state management
4. Note error handling and edge cases
5. Identify test coverage (what's tested = confirmed behavior)

### Step 2: Generate Inferred Specs

For each identified behavior, create a requirement in EARS format.
Mark each requirement as `[INFERRED]` to distinguish from human-authored specs.

Create a single feature folder for the baseline:

```
specforge/features/_baseline/
├── requirements.md    # All inferred requirements marked [INFERRED]
├── design.md          # Current architecture as-is
└── tasks.md           # Empty — baseline is already implemented
```

### Step 3: Present for Validation

→ 🔴 **GATE**: Present `requirements.md` to the user.

The user validates:
- Are the inferred behaviors correct?
- Are any behaviors missing?
- Are any behaviors listed that shouldn't be specs (implementation details vs features)?

→ 🔴 **GATE**: Present `design.md` to the user.

### Step 4: Register Baseline

Add to `features.json`:
```json
{
  "name": "_baseline",
  "status": "done",
  "workflow": "from-code",
  "created": "<date>",
  "note": "Inferred from existing codebase"
}
```

## Important

- **Don't over-spec.** Only document behaviors that are user-facing or 
  architecturally significant. Internal helper functions don't need specs.
- **Mark uncertainty.** If a behavior's intent is unclear from the code alone,
  mark it with `[UNCERTAIN]` and ask the user to confirm.
- **Respect existing tests.** If a test exists for a behavior, the spec should
  match what the test verifies, not what you think the code should do.
