---
name: sfx-product-owner
description: >
  Define product requirements through structured thinking. Generate a brief with user stories,
  priorities, and acceptance criteria. Trigger: "/product-owner", "/brief", "user story",
  "requirements", "what should we build", "define MVP", "product brief". Supports --from <path>
  to import existing PRD/doc as primary input.
---

# product-owner

Translate vague ideas into testable requirements. Think in problems, not solutions. Prioritize with MoSCoW.

**Always produces** `.ai/briefs/{slug}.md`.

## Arguments

| Flag | Required | Description |
|------|----------|-------------|
| `--from <path>` | No | Import existing doc as primary input |

If `--from` file can't be read → STOP, report. Never guess content.

## Step 1: Understand the Problem

### With --from <path>
1. Read document in full
2. Extract: problem, personas, metric, scope, features, constraints
3. Map to standard structure (Problem, Metric, Persona, MoSCoW, Stories)
4. Ask only about gaps — don't re-ask what the document covers
5. Present extraction for confirmation before writing

### Conversational (no --from)
Ask (skip what's already clear):
1. **Who** has this problem? (specific persona)
2. **What** are they trying to accomplish? (job-to-be-done)
3. **Why** can't they do it today? (current pain)
4. **How** will we measure success? (one concrete metric)

Max 2-3 exchanges. If enough info, proceed.

## Step 2: Scope with MoSCoW

| Priority | Meaning | Limit |
|----------|---------|-------|
| **Must** | Launch blocker | 3-5 items |
| **Should** | Important, workaround exists | any |
| **Could** | Nice to have | any |
| **Won't** | Explicitly out of scope | always list |

## Step 3: User Stories

For each Must, write story with acceptance criteria:

```markdown
### Story: {name}
**As a** {persona} **I want to** {action} **So that** {outcome}

#### Acceptance Criteria
- [ ] GIVEN {precondition} WHEN {action} THEN {result}
- [ ] GIVEN {edge case} WHEN {action} THEN {result}
```

## Step 4: Write Artifact

Generate `.ai/briefs/{slug}.md` using `templates/brief.tmpl.md`.

## Rules

- Start with the problem, not the solution.
- If user describes a solution, ask what problem it solves.
- Metrics must be measurable numbers.
- ALWAYS include "Won't" items — prevents scope creep.
- Acceptance criteria in GIVEN/WHEN/THEN — they become test cases.
- Each story independently implementable.
- Keep brief under 600 words.
- With --from: extract what exists, ask only what's missing.
- With --from: flag ambiguities in Open Questions, don't pick interpretations silently.
