---
name: sfx-prototipo
description: >
  The prototype primitive. Throwaway code that answers ONE question the interview could not settle
  by reading or searching. The answer survives; the code does not. Triggers: a question that only
  running code can answer, "does this state model feel right", "what should this look like", or any
  composing skill calling it — always after the user approved building it.
---

# sfx-prototipo

**A prototype is throwaway code that answers a question.** If you cannot say the question in one
line, there is nothing to prototype yet — go back to the interview.

## Before you build: it is approved, not assumed

Building costs real time, and that makes it a **decision**, not a check. The interview *proposes*
it, the user *approves* it. Never start one on your own initiative.

State the question, what you would build, and roughly how long, then wait.

## What language

1. **Read `.docs/constitucion.md`.** `sf init` already filled its header — `lenguaje`, `manifiesto`,
   `test_cmd` — by looking at the disk. If `lenguaje` is there, use it: the prototype should run
   with the toolchain that is already installed.
2. **If it is empty** (a new folder with no manifest), ask which language, offering the default
   below.
3. **The default, when there is nothing:** a single HTML file with everything inline, opened by
   double-click. It is the only thing that runs anywhere with nothing installed.

## Two shapes, and picking wrong wastes the whole thing

- **"Does this logic / state model feel right?"** → a single shareable file that pushes the model
  through the cases that are hard to reason about on paper: free-play buttons plus a couple of
  guided walkthroughs, drivable by someone who does not write code.
- **"What should this look like?"** → several **radically** different variations of the same
  screen, switchable in one click. Not three shades of the same layout.

If it is genuinely ambiguous, say which one you assumed at the top of the prototype.

## The rules

1. **Throwaway from day one, and it says so.** Lives in `.docs/prototipos/<pregunta-en-kebab>/`,
   outside the feature flow. **No branch** — there is no feature yet at this point, and inventing
   one would drag the whole per-feature ceremony into a question that has not been asked yet.
2. **Trivial to run.** One command, or a double-click. No setup step.
3. **No persistence.** State lives in memory. If the question itself is about the database, use a
   scratch file named so nobody mistakes it.
4. **No polish.** No tests, no error handling beyond making it runnable, no abstractions.
5. **Surface the state.** After every action, print or render the full relevant state. A prototype
   you cannot see the insides of answers nothing.

## What survives

**The code is disposable. The answer is not.**

When it is done, write the answer back into the interview as a settled decision, tagged with the
third provenance:

```markdown
- El modelo de estados aguanta la cancelación parcial sin un estado nuevo. [probado]
  - .docs/prototipos/cancelacion-parcial/
```

`probado` is the strongest tag there is — stronger than `retrieved`, because you did not read it
somewhere, you built it and watched it. It is also the most expensive, which is why it is approved
and not assumed.

If the prototype encoded the answer more precisely than prose can — a state machine, a schema, a
type — carry **that fragment** into the artifact, trimmed to the decision. Not the working demo.

## Rules

- One question per prototype. Two questions is two prototypes.
- Approved by the user before a line is written.
- Language from the constitution header; ask only when it is empty.
- No branch, no tests, no polish, no persistence.
- The answer goes back tagged `probado`, with a pointer to the folder.
