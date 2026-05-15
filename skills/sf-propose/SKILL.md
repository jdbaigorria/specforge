---
name: sf-propose
description: >
  Generate feature specifications: requirements, design, and tasks. Use when the user says
  "sf-propose", "propose feature", "specify", "spec this", "add feature", or describes
  functionality they want to build. Also triggers on "sf-propose --design-first" for
  architecture-driven specification, "sf-propose --from-code" for reverse-engineering specs
  from existing code, or when the user wants to refine/clarify existing specs. Covers the
  full specification lifecycle from idea to implementation plan.
---

# sf-propose

Generate feature specifications. Produces 3 artefacts with a human gate after each one.

## Pre-flight

1. Verify `specforge/` exists. If not: "Run `sf-init` first."
2. Read `specforge/features.json` to check for naming conflicts.
3. If `specforge/constitution.md` exists, read it — principles guide spec generation.
4. If `.ai/project.md` exists, read it — stack context informs design.

## Mode Detection

- Default → **Requirements-First** (this file covers the full flow)
- `--design-first` → Read `references/design-first.md` before proceeding
- `--from-code` → Read `references/from-code.md` before proceeding

## Requirements-First Flow

### Step 1: Understand the Feature

Engage the user in conversation:
- **What** does this feature do? (user-facing behavior)
- **Why** does it exist? (problem it solves)
- **Who** uses it? (actor/persona)
- **What are the boundaries?** (what it deliberately doesn't do)

Keep the conversation focused. 3-5 exchanges should be enough for most features.
Don't interrogate — if the user gives a clear description, proceed.

### Step 2: Generate requirements.md

Create `specforge/features/<name>/requirements.md`.

Use EARS notation for system behaviors. Read `references/ears-notation.md` if needed.

Structure:
```markdown
# <Feature Name> — Requirements

## Overview
[1-2 sentence summary]

## User Stories
- As a [actor], I want to [action], so that [benefit]

## System Requirements

### R1: [Requirement name]
WHEN [trigger]
THE SYSTEM SHALL [behavior]

**Acceptance Criteria:**
- [ ] AC1: [testable criterion]
- [ ] AC2: [testable criterion]

### R2: ...

## Edge Cases
- [What happens when X fails?]
- [What happens with empty input?]

## Out of Scope
- [What this feature deliberately doesn't do]
```

→ 🔴 **GATE**: Present `requirements.md` to the user.
- "Approved" → proceed to Step 3
- Changes requested → iterate on requirements, re-present
- If user edits the file directly → acknowledge changes, proceed

### Step 3: Generate design.md

Create `specforge/features/<name>/design.md`.

If a technical decision requires investigation, read `references/research.md`.

Only include sections relevant to the feature's complexity. A CLI flag addition
doesn't need Security Considerations or Performance Considerations. A payment
system does. Include a section only if it changes a decision or informs implementation.
Sections that would say "N/A" generate noise in the context — omit them.

Core sections (always include):
- Architecture (how it fits into the system)
- Components (what gets created/modified)
- Technical Decisions (if any non-obvious choices were made)

Conditional sections (include when relevant):
- Data Model (if entities or schemas are involved)
- API / Interface Contracts (if public interfaces change)
- Sequence / Flow (if the interaction has multiple steps)
- Error Handling (if failure modes are non-trivial)
- Testing Strategy (if testing approach differs from project default)
- Security Considerations (if security-sensitive surfaces exist)
- Performance Considerations (if performance-sensitive paths exist)

→ 🔴 **GATE**: Present `design.md` to the user.
- "Approved" → proceed to Step 4
- Changes requested → iterate, re-present
- If design changes invalidate requirements → flag and offer to update requirements

### Step 4: Generate tasks.md

Create `specforge/features/<name>/tasks.md`.

Structure:
```markdown
# <Feature Name> — Tasks

## Implementation Plan

### Wave 0: [theme] (no dependencies)
- [ ] T1: [task description] → R1
  - Expected: [what exists after this task]
  - Files: [files created/modified]

### Wave 1: [theme] (depends on Wave 0)
- [ ] T2: [task description] → R2
  - Expected: [what exists after this task]
  - Files: [files created/modified]

## Traceability Matrix
| Requirement | Task(s) |
|-------------|---------|
| R1          | T1      |
| R2          | T2      |
```

Every task must trace to at least one requirement.
Every requirement must be covered by at least one task.
If any requirement has no task, add one or flag it to the user.

→ 🔴 **GATE**: Present `tasks.md` to the user.
- "Approved" → update `features.json` status to `approved`
- Changes requested → iterate, re-present

### Step 5: Update Feature Registry

Update `specforge/features.json`:
```json
{
  "name": "<feature-name>",
  "status": "approved",
  "workflow": "requirements-first",
  "created": "<date>"
}
```

Append to `specforge/history.md`:
```
## [date] — Feature proposed: <name>
- Workflow: requirements-first
- Requirements: [count] | Tasks: [count] | Waves: [count]
```

Inform the user: "Feature `<name>` approved. Use `sf-build <name>` to start implementation."

## Post-Generation: Clarify (optional)

If the user asks to refine, review, or clarify specs after generation:
read `references/clarify.md` for the refinement procedure.

## Resync Detection

If invoked on a feature that already has artefacts:
1. Check which artefacts exist
2. Compare timestamps
3. If upstream is newer than downstream → offer to regenerate
4. If user modified an artefact directly → acknowledge and cascade
