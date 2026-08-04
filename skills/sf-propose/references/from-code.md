# From-Code: Reverse-Engineer Specs from Existing Code

## When to Use

- Brownfield project with code but no specs
- Need a spec baseline before adding new features
- Want to document what already exists for traceability

## Prerequisites

- `specforge/context/project.md` and `specforge/context/conventions.md` must exist (run `sf-init` first)
- The codebase must be readable and reasonably organized

## Flow

### Step 0: Scan the Inventory (deterministic)

```bash
sf onboard scan
```

The CLI computes the map — code files, top-level symbols per module, and the
test → module map — into `specforge/context/inventory.json` (+ `.md`).
**Interpret over that inventory instead of free-exploring the repo**: fewer
hallucinated modules, fewer tokens, and the symbol names are the exact anchors
(`path:symbol`) the trace will use later.

### Step 1: Analyze Code

Read `specforge/context/inventory.md` first, then, for each module/component
that matters:
1. Read entry points and public interfaces
2. Identify user-facing behaviors (CLI commands, API endpoints, UI actions)
3. Map data models and state management
4. Note error handling and edge cases
5. Identify test coverage (what's tested = confirmed behavior) — the inventory's
   test → module map is the starting point

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

Register through the CLI (the only writer of feature state):

```bash
sf feature add --feature=_baseline
```

### Step 5: Measure Adoption (spec coverage)

```bash
sf coverage
```

Reports the % of code files anchored to a live trace and sets the **ratchet
baseline** — from here, coverage only goes up (each new feature anchors what it
touches). Run it after every from-code baseline and in CI: incremental,
*measurable* adoption instead of a big-bang spec-everything.

## Important

- **Don't over-spec.** Only document behaviors that are user-facing or 
  architecturally significant. Internal helper functions don't need specs.
- **Mark uncertainty.** If a behavior's intent is unclear from the code alone,
  mark it with `[UNCERTAIN]` and ask the user to confirm.
- **Respect existing tests.** If a test exists for a behavior, the spec should
  match what the test verifies, not what you think the code should do.
