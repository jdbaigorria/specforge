# Research: Technical Decision Documentation

## When Triggered

During propose (either workflow), when:
- Multiple viable technical approaches exist
- The user asks "should I use X or Y?"
- A design decision has significant trade-offs
- The choice impacts architecture, cost, or maintainability

## Procedure

### 1. Frame the Decision

```markdown
## Decision: [e.g., "Message Queue Selection"]
**Context:** [why this decision matters]
**Constraints:** [from constitution or project context]
```

### 2. Evaluate Options (2-3 max)

For each option:
- **What it is:** one sentence
- **Pros:** concrete benefits for this project
- **Cons:** concrete drawbacks for this project
- **Fits when:** the scenario where this is the right choice

Don't be exhaustive. Focus on what matters for THIS project.

### 3. Recommend

State a recommendation with rationale tied to the project's 
constitution/constraints. If no strong preference, say so and
let the user decide.

### 4. Document in design.md

Add the decision to the Technical Decisions table:

```markdown
## Technical Decisions
| Decision | Choice | Rationale |
|----------|--------|-----------|
| Message queue | SQS | Serverless, no infra to manage, fits Lambda architecture |
```

If the decision was complex enough to warrant full documentation,
create `specforge/features/<name>/decisions/<decision-name>.md`
with the full analysis.

## Important

- Research is a tool within propose, not a separate phase
- Don't research decisions that are already constrained by the constitution
  (e.g., if constitution says "AWS only", don't evaluate GCP options)
- Keep it concise — a decision document longer than 1 page is too long
