---
name: sfx-tdd
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

Read: `specforge/context/compact-rules.md`, `specforge/context/project.md` (test runner), specs (acceptance criteria), design.md, existing code.

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

**A valid RED meets all three. Two of three is a broken test, not a RED:**

1. **It fails, it doesn't error.** A missing import, a typo, an undefined
   symbol — those are broken tests. The test must run and fail on its
   *assertion*.
2. **The failure message is the one you expected.** Write down the expected
   failure before running. A different message means the test isn't probing
   what you think.
3. **It fails because the feature is missing** — not because the test has a bug.

> **Statically-typed languages (Go, Rust, Java, TS):** testing a function that
> doesn't exist yet is *always* a compile error, so condition 1 fails by
> construction. Write the stub first — the signature with a trivial wrong return
> — then the RED is a real assertion failure. This is not optional there; it's
> the only way to get a valid RED.

Condition 3 is the one a CLI cannot check for you: judging *why* a test failed
needs reading the message. That's why it lives here and not in a gate — the
verification tier records that a RED happened, this rubric raises what that
record is worth.

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

## The Iron Law has a closed exit

"Test MUST fail first" is only half a rule until it says what happens when you
already wrote the code. It does now:

> **If you wrote the implementation before the test, delete it.**
> Not stashed on a branch. Not commented out. Not left open in another tab to
> consult while writing the test.

A test written while looking at the implementation asserts what the code *does*,
not what the requirement *asks for*. That is the exact failure the RED witness
exists to catch, and it passes every mechanical check: the test is real, it runs,
it goes green. Deleting and rewriting costs minutes. A suite that validates its
own reflection costs you the reason you were testing at all.

## Rationalizations

You will want to say these. Here is why they don't hold.

| The excuse | Why it doesn't hold |
|---|---|
| "The test is trivial, no need to watch it fail" | A test that never failed proved nothing — it may be asserting `true == true`. Watching it fail costs one run. |
| "I already know this works" | Then the test passes on the first run and you've lost nothing. If it doesn't, you just learned you were wrong. |
| "I'll write both together and split the commit after" | The commit split is cosmetic; the order you *thought* in is what the test records. Splitting after produces a test shaped by the code. |
| "The code was already there when I picked up the task" | Then it's covered by an existing test or it isn't. If it isn't, this is the moment — write the test against the requirement, not against the code you just read. |
| "Deleting working code is wasteful" | The code isn't the asset; the verified behavior is. You'll rewrite it in minutes with a test that actually constrains it. |
