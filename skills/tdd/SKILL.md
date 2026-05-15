---
name: tdd
description: >
  Implement code using test-driven development. Red-Green-Refactor, one vertical slice at a
  time. Trigger: "/tdd", "/tdd <task>", "tdd this", "use TDD", "red-green-refactor".
---

# tdd

RED → GREEN → REFACTOR. One test, one implementation, one cycle. Repeat.

## Core Principle

Tests verify behavior through public interfaces, not implementation details.
Good: "user can checkout with valid cart". Bad: "checkout calls paymentService.process".

## The Iron Cycle

1. Write ONE failing test. It MUST fail.
2. Write MINIMAL code to pass ONLY that test.
3. NEVER refactor while RED. GREEN first.
4. Refactor only when all tests pass.
5. Next behavior. Repeat.

Vertical (RIGHT): RED→GREEN per behavior, one at a time.
Horizontal (WRONG): all tests first, all impl second.

## Step 1: Input

- `/tdd` alone → ask what to build
- `/tdd <change>` → read `specforge/features/{change}/tasks.md`
- `/tdd <description>` → use as feature description

## Step 2: Context

Read: `.ai/compact-rules.md`, `.ai/project.md` (test runner), specs (acceptance criteria), design.md, existing code.

## Step 3: Plan

```markdown
## TDD Plan
### Interface: {what changes}
### Behaviors (in order):
1. {tracer bullet — simplest happy path}
2. {next behavior}
3. {next behavior}
### Testability: {concerns or "none"}
```

→ 🔴 GATE: "Ready to start? [y/N]"

## Step 4: Execute Cycles

### RED
Write ONE test. Public interface only. Must FAIL.
If it passes → behavior exists (skip) or test is wrong (rewrite).

### GREEN
MINIMAL code to pass. No speculation. No "while I'm here..."
Run all tests. This test passes ✅. Previous tests still pass ✅.

### REFACTOR (optional)
Only when ALL GREEN. Duplication, naming, shallow abstractions.
Run tests after every change. Must remain GREEN.

## Step 5: Evidence Table (required)

```markdown
| # | Behavior | RED (test) | GREEN (impl) | REFACTOR |
|---|----------|-----------|--------------|----------|
| 1 | user can login | tests/auth.test.ts:12 | src/auth.ts:5-18 | extracted validateEmail |
| 2 | login fails wrong pw | tests/auth.test.ts:28 | src/auth.ts:20-25 | — |
```

## Step 6: Mark Tasks (if from tasks.md)

Update `[  ]` → `[x]` in tasks.md for completed tasks.

## Rules

- ONE test at a time. ONE impl at a time.
- Test MUST fail first. No exceptions.
- Public interfaces only in tests. No internal mocks.
- No speculative code. No test demands it → don't write it.
- NEVER refactor while RED.
- Prefer real over mocked. Integration-style when possible.
- Start simplest happy path. Edge cases come later.
