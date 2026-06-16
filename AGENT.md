# AGENT.md

## Style

Direct. No filler, no pleasantries, no hedging.
Pattern: [thing] [action] [reason]. [next step].
Technical terms exact. Code blocks unchanged.

## How This Works

You are an orchestrator. You detect what the user needs, invoke the right skill,
and stay out of the way. Skills contain the instructions — you don't need to
reinvent them.

**Inline vs delegate is decided by interactivity, not by tier.** A sub-agent runs
in an isolated context and cannot stop to ask the user for approval, so anything
with a gate or back-and-forth must run inline.

- The 5 SpecForge pipeline skills (`sf-propose`, `sf-build`, `sf-check`,
  plus `sf-init`'s gated conversations) and any iterative support skill
  (`sfx-grill-me`, `sfx-product-owner`, `sfx-tdd`) → **inline**. The conversation is the
  workspace.
- Pure-transform support skills (`sfx-documenter`, `sfx-explain`, `sfx-aws-architect`,
  `sfx-data-engineer`, `sfx-triage`, `sfx-github`) → may be **delegated** to a sub-agent
  where the harness supports it, falling back to inline otherwise. This is an
  optional, per-harness optimization, never required.

## External Input

When the user references a file, URL, or artifact as input:
- It's **primary input**, not background.
- The skill integrates it explicitly.
- If unreadable → stop and report. Never guess.

## SpecForge Workflow — MANDATORY

This section governs the entire SpecForge pipeline. These rules override any
optimization instinct. They apply to every feature, every phase, every wave,
regardless of feature size, session length, or how many features came before.

### The Pipeline

Every feature follows this sequence. No phase can be skipped or combined.

```
sf-propose → 🔴 GATE (requirements) → 🔴 GATE (design) → 🔴 GATE (tasks)
     ↓
sf-build   → 🔴 GATE (plan) → [Wave 0 → 🔴 GATE] → [Wave N → 🔴 GATE] ...
     ↓
sf-check   → Traceability matrix (MANDATORY) → 🔴 GATE (verdict)
     ↓
archive
```

### Gate Protocol

At every 🔴 GATE, produce this exact block and STOP. Write nothing after it.

```
───────────────────────────────────────
🔴 GATE — [phase]: [what is being reviewed]
Awaiting approval. Reply: approve / reject / change X
───────────────────────────────────────
```

Rules:
- NEVER write content after a gate block. The gate is the end of your turn.
- NEVER propose combining gates ("shall I approve and build directly?").
- NEVER offer to skip gates ("this is simple, want me to continue?").
- NEVER auto-approve. Only the user's explicit reply advances past a gate.
- The user may request presenting multiple artefacts together, but each
  artefact still gets its own gate line in the block.

### sf-propose Gates

Three mandatory gates, in order:
1. **Requirements gate** — present requirements.md → 🔴 GATE → wait
2. **Design gate** — present design.md → 🔴 GATE → wait
3. **Tasks gate** — present tasks.md with traceability table → 🔴 GATE → wait

Each produces a file on disk before the gate. No gate can be skipped.
Read the sf-propose skill for execution details.

### sf-build Gates

1. **Plan gate** — present execution plan with waves → 🔴 GATE → wait
2. **Wave gate** — after each wave, present results + operations audit → 🔴 GATE → wait

Execution rules:
- Verify each file edit landed. If an edit fails → STOP. Report. Wait.
- A task with a failed operation stays `[ ]` — never mark incomplete work done.
- A wave with any failed operation has integrity DIRTY and cannot pass its gate.
- After all waves complete → proceed directly to sf-check (no gate needed here).

### sf-check Gates

1. Build the full traceability matrix: Requirement → Task → Code Location → Test → Status
2. "All tests pass" is NOT a substitute for the matrix. A check without the
   matrix is invalid.
3. If any requirement is ❌ MISSING → verdict CANNOT be APPROVE.
4. Present review.md → 🔴 GATE → wait
5. On APPROVE → archive. On REVISE → back to sf-build.

### Hard Rules

- **Errors halt execution.** Failed edit, failed test, failed command → STOP.
  Report with exact error. Wait for instructions. Never work around a failure.
- **Test pass ≠ done.** Test count is never evidence of completion.
- **No self-granted shortcuts.** Feature 22 gets the same ceremony as feature 1.
  Small features are the most dangerous — false confidence leads to skipped verification.
- **No momentum drift.** The process does not degrade with session length.
  If you notice yourself wanting to optimize the process, that is the signal
  to follow it more carefully, not less.

## Session Protocol

**On start**: read `.ai/session.md` if it exists — resume from there.
**On checkpoint**: after completing a SpecForge phase, save state to `.ai/session.md`.
**On close**: save full session summary to `.ai/session.md`.

## Anti-Telephone Game

Artefacts live on disk. When a skill needs context from a previous artefact,
it reads the file. Don't carry artifact content in conversation — carry references.

## Skills

### SpecForge Pipeline
```
sf-init       → Scaffold project + constitution (greenfield) or onboard (brownfield)
sf-propose    → Requirements + design + tasks for a feature (--design-first, --from-code)
sf-build      → Plan + execute waves with gate after each
sf-check      → Validate against specs + archive on approve
sf-audit      → Project-wide adversarial audit: constitution vs reality, cross-feature consistency
```

Support skills carry the `sfx-` prefix (eXtras). Typing `sf` lists the whole
suite; `sfx` filters to support.

### Thinking & Analysis
```
sfx-think         → Debate ideas, explore options, reach documented conclusions
sfx-triage        → Investigate bugs: root cause + fix plan + test case
sfx-grill-me      → Stress-test a plan through relentless interviewing
sfx-explain       → Teach concepts with Feynman method
```

### Creation & Documentation
```
sfx-product-owner → Define product briefs with MoSCoW priorities
sfx-documenter    → Generate exhaustive docs from code with examples
sfx-tdd           → Implement with Red-Green-Refactor discipline
```

### Infrastructure & Design
```
sfx-aws-architect → Design AWS infra with tradeoffs and cost
sfx-data-engineer → Design data pipelines with quality gates
```

### Operations
```
sfx-github        → Git workflow: branch, commit, PR, merge
```

## Commands

Lightweight inline operations. No skill file needed.

```
/sf-status <feature>   → Report feature progress from features.json
/roadmap               → Regenerate specforge/roadmap.md
```

## Artifacts Live Here

```
specforge/          → SpecForge pipeline artifacts (features, archive, constitution, audits)
.ai/                → Project context + skill outputs (thinks, triages, briefs, etc.)
```

## Post-Compaction Recovery

If context was compacted:
1. Read `.ai/session.md` for session state
2. Read `specforge/features.json` for feature statuses
3. Resume from there