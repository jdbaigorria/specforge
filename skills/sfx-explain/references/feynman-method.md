# Feynman Method — Detailed Guide

## The Core Principle

If you can't explain it simply, you don't understand it well enough.
Your job is to make the user understand, not to demonstrate your knowledge.

## Layer 1: Core Idea (one sentence)

Strip away everything until you have one sentence a non-technical person
would understand. This is the hardest part — it forces you to find the
essence.

**Good:** "A database index is a shortcut that lets the database find data
without reading every row."

**Bad:** "A B-tree index is a balanced tree data structure that maintains
sorted data and allows searches, insertions, and deletions in O(log n)."

The second is accurate but useless for building understanding. Accuracy
comes later, in deeper layers.

## Layer 2: Analogy

Connect to something the user already knows. Rules:

- **Everyday life, not technical.** "Like a shipping container" not "like a VM."
  You don't know what the user knows about other technical concepts.
- **Acknowledge the limits.** "This analogy breaks down when..." — every
  analogy is incomplete, and saying so builds trust.
- **One analogy per concept.** Multiple analogies confuse more than they clarify.

### Good analogies by domain

- **Caching:** "Like keeping frequently used ingredients on the counter instead
  of walking to the pantry every time."
- **Load balancer:** "Like a host at a restaurant who sends each party to the
  least busy waiter."
- **API:** "Like a waiter — you don't go to the kitchen. You tell the waiter
  what you want, they bring it in a standard plate."
- **Git branch:** "Like making a photocopy of a document to try edits without
  risking the original."

## Layer 3: Build Up

Start simple. Add complexity only after the base is solid.

```
Layer 3a: What it is (definition, but in simple terms)
Layer 3b: Why it exists (what problem it solved)
Layer 3c: How it works (the mechanism, simplified)
Layer 3d: When to use it / when NOT to (decision criteria)
```

At each sub-layer, check: "Would someone who only read the previous layers
understand this?" If not, you skipped a step.

## Layer 4: Concrete Example

Abstract explanations fail. The example IS the explanation for most people.

### Rules for good examples

- **Real, not toy.** Don't explain database indexing with a 3-row table.
  Use a scenario with 10 million rows where it actually matters.
- **Runnable.** If it's code, it should work. Pseudocode is a cop-out.
- **In the user's stack.** If `specforge/context/project.md` says Python, show Python.
  If it says TypeScript, show TypeScript. Unfamiliar syntax adds cognitive
  load that competes with the concept you're teaching.
- **Show the before AND after.** "Without this concept" vs "with this concept"
  makes the value visceral.

### Example structure

```
Here's a problem:
{scenario without the concept — show the pain}

Here's how {concept} solves it:
{same scenario with the concept — show the relief}
```

## Layer 5: Tradeoffs

This is where real understanding lives. If something "has no downsides,"
the explanation is incomplete.

Every concept has:
- **What you gain** (the benefit everyone talks about)
- **What you lose** (the cost nobody mentions)
- **When it backfires** (the scenario where it's the wrong choice)

Example for caching:
- Gain: speed — reads are orders of magnitude faster
- Lose: freshness — cached data can be stale
- Backfires: when data changes frequently and staleness is unacceptable
  (stock prices, real-time location)

## Layer 6: Summary

One sentence that captures the essence. The user should be able to repeat
this sentence to someone else and convey the key insight.

After the summary, offer: "Want me to go deeper on any part?"

## Anti-patterns

- **The Wikipedia dump.** Giving a comprehensive, accurate, boring overview.
  Understanding is not completeness.
- **The jargon spiral.** Explaining jargon with more jargon. "A monad is a
  monoid in the category of endofunctors" explains nothing.
- **The hedge.** "It depends" without then explaining WHAT it depends on.
  Name the variables that determine the answer.
- **The code-only explanation.** Dropping a code block without context. Code
  shows HOW, not WHY. Explain the why, then show the code.
- **Over-analogizing.** Stretching an analogy past its breaking point. Say
  "this is where the analogy stops being useful" and switch to direct
  explanation.
