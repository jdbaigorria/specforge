---
name: sfx-explain
delegate: true
disable-model-invocation: true
description: >
  Explain a concept using the Feynman method — simple language, analogies, built from basics,
  honest tradeoffs. A USER DOOR: you invoke it with `/sfx-explain`; no skill composes it and the
  machine never runs it. Works for any topic, technical or not, and needs no project context.
---

# Explain

Explain a concept using the Feynman method. Simple language, analogies, layered
complexity, concrete examples, honest tradeoffs.

This skill is for **understanding**, not doing. If the user wants to build something
after understanding it, they invoke the appropriate action skill.

> **Why it is closed to the model.** Measured 2026-09-13 (`primitivos.md` §4①): zero consumers —
> the machine does not run it, no skill composes it, nothing needs it. But its old triggers
> (*"explain", "why does", "how does X work", "I don't understand", "what is X"*) matched almost
> any question, so it competed in the routing race against skills that were actually needed and
> won races it should have lost. Same shape as the bug measured in opencode on 2026-09-08.
>
> A rich trigger list on a door nobody composes is a cost with no consumer (R2). Matt Pocock's
> equivalent (`teach`) is user-invoked for the same reason.

## Step 1: Determine Topic

- `/explain` alone → ask: "What do you want me to explain?"
- `/explain <topic>` → use the topic
- "how does X work" / "what is X" → use X
- Follow-up in same session → go deeper on the specific part, don't re-explain basics

## Step 2: Read Context (if relevant)

If the topic relates to the current project:
- `.docs/constitucion.md` for stack (so code examples match the user's world)
- Relevant source files if the question is about THIS codebase
- Any note from a previous explanation of the same topic, if the user points you at one

If topic is generic ("how does OAuth work?") → skip context, explain universally.

## Step 3: Explain

Read `references/feynman-method.md` for the detailed methodology.

**Quick summary — 6 layers:**

1. **Core idea in one sentence.** No jargon.
2. **Analogy.** From everyday life, not from other technical concepts.
3. **Build up layer by layer.** What it is → why it exists → how it works → when to use it.
4. **Concrete example.** Real code, real scenario. Use the user's stack if known.
5. **Tradeoffs.** What you gain, what you lose, what to watch out for.
6. **Summary.** One sentence. Offer to go deeper.

Present using this structure:

```markdown
## {Topic}

**In one sentence**: {core idea}

**Analogy**: {everyday comparison}

**How it works**:
{Layer by layer, short paragraphs}

**Example**:
{Real, runnable code or concrete scenario}

**Tradeoffs**:
- You gain: {benefit}
- You lose: {cost}
- Watch out for: {common pitfall}

**Summary**: {one line}
```

## Step 4: Handle Follow-ups

If the user asks to go deeper on a specific part:
- Explain that part with the same method
- Don't re-explain the basics
- If it's a substantially new concept → treat as new explanation cycle

## Step 5: Offer to Save

After the user is satisfied (no more follow-ups), ask:

> Save this explanation to a file? Where? [path / N]

Default: no.

If yes → generate using `templates/explanation.tmpl.md`. Include all layers +
follow-ups in a single cohesive document.

## Rules

- No jargon without defining it immediately in parentheses.
- No walls of text. Short paragraphs, breathing room.
- One concept at a time. Multiple concepts → ask which first, or explain sequentially.
- Code examples must be real and runnable, not pseudocode.
- Analogies from everyday life, not other technical concepts.
- Always name tradeoffs. "No downsides" means you haven't understood it.
- Always end with a summary sentence.
