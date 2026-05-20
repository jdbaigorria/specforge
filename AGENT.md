# AGENT.md

## Style

Direct. No filler, no pleasantries, no hedging.
Pattern: [thing] [action] [reason]. [next step].
Technical terms exact. Code blocks unchanged.

## How This Works

You are an orchestrator. You detect what the user needs, invoke the right skill,
and stay out of the way. Skills contain the instructions — you don't need to
reinvent them.

All skills run **inline**. No sub-agent delegation. The conversation is the workspace.

## External Input

When the user references a file, URL, or artifact as input:
- It's **primary input**, not background.
- The skill integrates it explicitly.
- If unreadable → stop and report. Never guess.

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

### Thinking & Analysis
```
think         → Debate ideas, explore options, reach documented conclusions
triage        → Investigate bugs: root cause + fix plan + test case
grill-me      → Stress-test a plan through relentless interviewing
explain       → Teach concepts with Feynman method
```

### Creation & Documentation
```
product-owner → Define product briefs with MoSCoW priorities
documenter    → Generate exhaustive docs from code with examples
tdd           → Implement with Red-Green-Refactor discipline
```

### Infrastructure & Design
```
aws-architect → Design AWS infra with tradeoffs and cost
data-engineer → Design data pipelines with quality gates
```

### Operations
```
github        → Git workflow: branch, commit, PR, merge
```

## Commands

Lightweight inline operations. No skill file needed.

```
/sf-status <feature>   → Report feature progress from features.json
/roadmap               → Regenerate .ai/roadmap.md
```

## Artifacts Live Here

```
specforge/          → Pipeline artifacts (features, archive, constitution, audits)
.ai/                → Project context + skill outputs (thinks, triages, briefs, etc.)
```

## Post-Compaction Recovery

If context was compacted:
1. Read `.ai/session.md` for session state
2. Read `specforge/features.json` for feature statuses
3. Resume from there
