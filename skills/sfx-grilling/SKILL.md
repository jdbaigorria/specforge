---
name: sfx-grilling
description: >
  The interview primitive. Work an idea, a plan or a design as a decision tree: ask the whole
  frontier in rounds, each question numbered and carrying your recommended answer, until the
  frontier is empty. This is the SINGLE OWNER of the interview method in SpecForge — every skill
  that interviews composes this one instead of re-explaining how. Triggers: "grill", "interview
  me", "stress-test this", "poke holes in this", or any composing skill calling it.
---

# sfx-grilling

**The frontier, in one line:** every decision whose prerequisites are already settled — the
questions you can ask *now* without guessing at answers you have not heard yet.

## The method

Map the subject as a **decision tree**: every decision branches into the decisions that hang off it.
Work the tree in **rounds**.

**Ask the whole frontier in one round.** Number each question and give your recommended answer.
Then stop and wait.

```
❓ **Q1 — <title>**: <the question, may be several paragraphs, may offer choices>

➡️ <your recommended answer, with the reason in one line>

---

❓ **Q2 — <title>**: …

➡️ …
```

Each round of answers reshapes the tree: settled decisions push the frontier outward and unblock
questions that depended on them. Recompute the frontier and ask the next round.

**A question whose answer depends on another question still open in this round belongs to a LATER
round, not this one.** Say so out loud when you defer one — *"la Qn queda para la ronda que viene:
depende de la Q2"* — because that is the proof the tree is real and not a list.

## Finding facts is your job, never the user's

When a frontier question needs a fact from the world — does this package exist, what does the code
do, what do the docs say — **go get it**. Never ask the user something you could look up.

- a fact about the outside world → `Call the Skill tool with "sfx-buscar"`
- a fact about this repo → read it yourself, or dispatch a subagent
- a question that only running code can answer → `Call the Skill tool with "sfx-prototipo"`,
  **after the user approves building it** (building costs time; that is a decision, not a check)
- **a question no fact can settle** — several shapes are viable and the evidence does not separate
  them → `Call the Skill tool with "sfx-think"`. This is the rarest branch, and the tell is
  specific: you searched, you found, and the options are still tied. Do **not** reach for it to
  avoid asking a question you could just ask.

**Do not block on any of them.** A running search is an unsettled prerequisite: only the questions
downstream of it wait. Ask the rest of the frontier now.

**The decisions are the user's.** Put each one to them and wait. You bring facts and
recommendations; you do not settle a branch on your own.

## When a word is doing two jobs

If an answer reveals a term meaning two different things, stop the round and
`Call the Skill tool with "sfx-vocabulario"`. A tree built on a word that means two things branches
wrong, and every round after that is wasted.

## The cut

**The session is done when the frontier is empty**: every branch visited, nothing silently assumed.
That is the criterion — not "the user got tired", not a question count.

If the user says "enough" before that, stop immediately **and say what is still open**. An interview
cut short with three branches unvisited is a fine outcome; one that pretends it finished is not.

## What it leaves

The composing skill decides the file. If it did not name one, write `.docs/entrevista.md`:

```yaml
---
rondas: 4
preguntas: 23
abiertas: 0      # branches left unvisited. A gate counts this.
---
```

Body: the rounds as they happened — question, recommendation, answer. **Verbatim, not summarized.**
The point is that someone can come back and see why a decision was made.

Every load-bearing answer carries provenance when it came from outside the user's head:
`retrieved` (with the link), `model-prior` (unverified), `probado` (built it and saw it).

## Rules

- Ask the frontier, not one question at a time. Fewer round trips is less friction.
- Every question numbered, every question with your recommended answer.
- Facts are yours, decisions are theirs.
- A deferred question gets said out loud, with what it waits on.
- The cut is an empty frontier, not fatigue.
- Never lecture. You are extracting their thinking, not teaching them yours.
