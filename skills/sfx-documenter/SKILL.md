---
name: sfx-documenter
description: >
  Generate exhaustive documentation from code with examples, type signatures, edge cases, and
  usage guides. Use when the user says "document this", "generate docs", "write documentation",
  "API docs", "reference docs", "how-to guide", "document the module", "create a README for",
  "I need docs for this code", or any request to produce human-readable documentation from
  existing code. Also triggers on "document the project", "generate developer guide", or
  "onboarding docs". Works on any codebase — does not require SpecForge.
---

# Documenter

Generate exhaustive documentation from code. Real examples, accurate signatures,
edge cases documented, nothing left to guesswork.

## Mode Detection

- `/documenter <path>` → Document that specific file, module, or directory
- `/documenter --api <path>` → API/reference documentation (signatures, params, returns, examples)
- `/documenter --guide <topic>` → How-to guide (narrative, step-by-step, with code examples)
- `/documenter --project` → Full project documentation (README + architecture + API + guides)
- No flag → infer from context. Single file → API mode. Directory → project mode.

## Step 1: Read and Analyze

Before writing a single line of documentation:

1. **Read the code.** All of it in scope. Not just the public functions — read the internals
   to understand what the code actually does, not what it looks like it does.
2. **Read existing docs.** If there are docstrings, README, comments, CHANGELOG — read them.
   Don't duplicate, don't contradict. Extend and improve.
3. **Read tests.** Tests are documentation. They show intended usage, edge cases, and expected
   behavior. Extract examples from real test cases when possible.
4. **Read `.ai/project.md` and `.ai/conventions.md`** if they exist — match the project's
   documentation style.

## Step 2: Generate Documentation

Read the appropriate reference for the mode:
- API/reference → `references/api-docs.md`
- Guide/tutorial → `references/guide-docs.md`
- Project → both, plus `references/project-docs.md`

### Core principles (all modes)

**Examples are mandatory.** Every function, every endpoint, every class gets at least one
real usage example. Not pseudocode — runnable code that a developer can copy-paste and adapt.
Extract from tests when possible; they're already verified.

**Show the edge cases.** Don't just document the happy path. What happens with empty input?
Null? Invalid types? Concurrent access? The edge cases are where developers actually get stuck.

**Accurate signatures.** Types, defaults, optionals — all explicit. If the language has type
hints, use them. If it doesn't, document the expected types in text. A wrong type in docs
is worse than no docs.

**Progressive depth.** Start with a one-liner (what it does). Then a paragraph (how to use it).
Then examples (show me). Then edge cases (what to watch out for). A developer skimming should
get value from every level of depth.

## Step 3: Output

### Single file/module → inline or save

If documenting a single file, present inline and offer to save:
> Save to `docs/{module}.md`? [y/N]

### Directory/project → always save

Create documentation files:
```
docs/
├── README.md              ← project overview + quickstart
├── api/
│   ├── {module}.md        ← one file per module
│   └── {module}.md
├── guides/
│   └── {topic}.md         ← one file per how-to
└── architecture.md        ← system overview (if --project)
```

→ 🔴 **GATE**: Present generated documentation. User reviews before finalizing.

## Rules

- EVERY public function/class/endpoint gets an example. No exceptions.
- Examples must be RUNNABLE. Copy-paste-able. Using the project's actual imports and patterns.
- Don't document implementation details unless they affect usage.
- Don't document the obvious. `get_name() → returns the name` wastes everyone's time.
  Document the non-obvious: "returns None if the user hasn't set a display name, falls back
  to email prefix".
- If a function has side effects, document them prominently.
- If a function can throw/raise, list the exceptions and when they occur.
- If parameters have valid ranges or constraints, state them.
- Match the project's code style in examples (from `.ai/conventions.md`).
- Don't generate docs for private/internal APIs unless the user asks.
- Keep each doc file focused. One module = one file. One guide = one file.
