---
name: grill-me
description: >
  Interview the user relentlessly about a plan, design, or artifact until reaching shared
  understanding. Walk down every branch of the decision tree. Trigger: "/grill-me",
  "/grill-me <artifact-path>", "grill me", "stress-test this plan", "interview me about",
  "poke holes in this", "challenge my design".
---

# grill-me

Interview relentlessly about every aspect until you and the user reach a shared understanding.

## The 5 Principles

1. **Interview relentlessly** about every aspect until shared understanding.
2. **Walk down each branch** of the decision tree, resolving dependencies one-by-one.
3. **For each question, provide your recommended answer**. User can accept with "yes" or push back.
4. **Ask one question at a time**. Never batch questions.
5. **If a question can be answered by exploring the codebase, explore instead** — don't waste user's time.

## Step 1: Determine Target

- `/grill-me` no arg → ask: "What do you want to stress-test?"
- `/grill-me <path>` → read the artifact
- `/grill-me <topic>` → ask for 2-3 sentences of context

## Step 2: Save Prompt

Before starting:

> Ready to grill. This session might last 15-45 minutes with 15-50 questions.
> Save transcript to .ai/grills/{slug}.md when done? [y/N]

Default: no. Remember the answer.

## Step 3: Explore Context First

Before asking questions, read:
- The artifact (if given)
- `.ai/project.md` (if exists)
- Related code or specs
- `specforge/` specs if relevant

This prevents asking questions you can answer yourself (principle 5).

## Step 4: Interview

For each question:

```markdown
**Q{N}**: {question}

Recommended: {your recommended answer with brief rationale}
```

Wait for response. Adapt next question based on answer.

Continue until:
- User says "stop" / "done" / "enough"
- All branches resolved
- 3 consecutive "I don't know" → pause, suggest research first

Track:
- Decisions made
- Open issues
- New branches to explore

## Step 5: Generate Summary

### If user said "save"

Write to `.ai/grills/{slug}.md`:

```markdown
# Grill: {topic}

**Date**: {YYYY-MM-DD}
**Questions**: {count}

## Summary

### Decisions Made
- {decision}: {choice + rationale}

### Open Issues
- {issue}: {why unresolved}

### Key Insights
- {insight surfaced during conversation}

## Full Transcript
### Q1: {question}
**Recommended**: {recommendation}
**User**: {response}
...
```

### If "no save" → don't write any file.

## Step 6: Return Summary

```
## Grill Complete: {topic}

**Questions**: {count}
**Saved**: {path | not saved}

### Decisions Made ({count})
- {decision, one line}

### Open Issues ({count})
- {issue, one line}

### Recommended Next Step
{sf-propose, sf-research, or "resolve open issues first"}
```

## Rules

- ONE question at a time. Never batch.
- ALWAYS provide recommended answer with each question.
- If answerable by reading code/artifacts, do that FIRST.
- Adapt depth to responses. Quick "yes" → move faster. Long deliberation → dig deeper.
- 3+ "I don't know" → pause, suggest research.
- NEVER lecture. Extract the user's thinking, don't teach.
- Respect "enough" / "stop" — end immediately with summary.
