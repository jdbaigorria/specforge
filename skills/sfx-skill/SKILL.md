---
name: sfx-skill
description: >
  Write, fix or review a skill — and prove it works before trusting it. Use when creating a new
  SKILL.md, editing an existing one, deciding whether something should be a skill at all, or when
  a skill exists and the agent keeps not using it (or uses it when it should not). Triggers:
  "/sfx-skill", "write a skill for", "this skill is not firing", "fix this SKILL.md", "should
  this be a skill". Works on any repo's skills, with or without SpecForge — the SpecForge-specific
  conventions are in their own section and you skip them elsewhere.
---

# sfx-skill

**A skill is code that shapes behaviour, not a document about behaviour.** It ships to a reader
who is fast, literal, under time pressure, and who will take any available reading that lets it
skip work. That reader is not hostile — it is optimising, and your prose is the only thing
standing between the optimisation and the shortcut.

So the only question that matters about a skill is empirical: **did the agent do the thing?** You
cannot answer that by reading your own draft. You answer it by watching one fail.

---

## The cycle — it is `sfx-tdd` with prose as the code

| `sfx-tdd` | here |
|---|---|
| write the failing test | run the scenario on a subagent **without** the skill |
| watch it fail (RED) | write down the exact excuse it used |
| minimal code | write the skill against **that excuse**, not against excuses in general |
| watch it pass (GREEN) | same scenario, skill present, fresh subagent |
| refactor | find the next excuse, close it, re-run |

**If you did not watch it fail first, you do not know what the skill has to teach.** You know what
you would have written anyway — which is the thing that was already not working.

### Running the red

Fresh subagent. Give it the situation and nothing else: no hint that a skill exists, no framing
that gives away the answer. Then read what it did, and copy the **rationalisation verbatim** into
your notes. Not "it skipped the tests" — the actual sentence: *"the change is small enough that a
test would be ceremony."* That sentence is your target. A skill that argues against a paraphrase
of it will lose to the original.

### Running the green

Fresh subagent again, same scenario, skill present. Two ways this passes badly, and you check for
both:

- **It complied because you told it the answer.** If the scenario names the skill, you tested
  nothing.
- **It complied once.** Run it again with the pressure up: a deadline in the prompt, a user who
  already said "just do it quickly", a task where the shortcut is genuinely tempting.

---

## The description is a routing rule, and it runs before anything else

> **Measured in this repo, 2026-09-08.** A fuzzy product idea arrived at opencode. The trace was
> `Read . → Skill "sfx-think"`. It never opened `AGENTS.md`, whose step 1 says to run `sf next`.
>
> It was not disobedience. **Skill descriptions are always in context; an instruction file has to
> be opened.** A file you must go read never beats text that is already there.

The description is therefore not a summary. It is the only part of your skill that competes for
the job, and it competes against every other description at once.

**Write it as the conditions under which this skill wins**, not as what it does:

- Lead with the situation, not the capability.
- Name the **precondition** if there is one. `sfx-think` needs its options already on the table,
  and saying so is what keeps raw ideas out of it.
- Where two skills sit close together, **name the other one and the line between them, inside the
  description**. A negative clause alone is weak; a negative clause plus the correct destination
  is a redirect.
- **Cut the triggers that belong to the neighbour.** This is the one that gets skipped. A trigger
  list is not free: every fuzzy phrase widens your net over somebody else's territory. If "what if
  we" could belong to two skills, it belongs to neither.

### The invocation axis — who is allowed to reach this

Two kinds, and picking wrong is what causes the races above.

**Model-invoked** (the default: omit the field). The model fires it on its own. Right for a
**primitive** — something other skills compose — and for a worker the model should reach for
autonomously. Keep the trigger phrasing rich: auto-invocation is the point.

**User-invoked** (`disable-model-invocation: true`). Only a human typing its name reaches it. No
other skill can call it, including by naming it to the Skill tool. Right for an **entry point a
person chooses**: expensive, slow, or once-per-project work. Here the description is
**human-facing** — somebody scrolling a list of slash commands — so strip the trigger lists out
entirely.

> **The test is not "is it useful".** It is: *could the model usefully reach for this on its own,
> and would it be right more often than wrong?* Reuse is a reason to extract a skill, never a
> reason to make it model-invoked.

**A user-invoked skill may call model-invoked ones. It can never call another user-invoked one.**
So before gating something, check that nobody composes it. A mention in prose ("this is what
`/sfx-verificar` generates") is fine — that is a label for a human. An operative step ("call the
Skill tool with X") is not.

---

## What earns a skill, and what does not

**Write one when** the technique was not obvious to you, you would reach for it again across
projects, and it turns on judgement.

**Do not write one when:**

- **It is enforceable.** If a regex, a schema or a gate can decide it, build that instead — a
  check that fails is worth more than a paragraph that asks. In SpecForge this is the house rule:
  prose is for judgement calls, compuertas are for facts.
- **It is a story about solving something once.** That is a journal entry (`sfx-journal`).
- **It is specific to one project.** That belongs in the constitution, or in `CLAUDE.md`.

---

## Shape

```text
skills/<name>/
  SKILL.md              required
  references/           heavy reference — 100+ lines you do not need every time
  templates/            what the skill fills in
```

**Keep inline:** the principle, the method, anything under ~50 lines, and every red-flags table.
**Split out:** long reference material and anything needed only sometimes. The split is about what
has to be in context to do the work, not about tidiness.

Frontmatter carries `name` and `description`, plus `disable-model-invocation: true` when it is an
entry point and `delegate: true` when the work belongs in a subagent.

---

## Testing a skill against pressure

Once the green passes, go hunting for the loophole. The reliable pressures:

1. **Time.** "we ship in an hour."
2. **Permission.** the user already said the shortcut is fine.
3. **Size.** the task is genuinely tiny and the process genuinely looks like overhead.
4. **Competence.** the agent knows the concept and concludes it does not need the skill.

For each one that breaks compliance, add the excuse to a red-flags table **in the agent's own
words** — the sentence it actually produced, answered in one line. A table of excuses you invented
reads as generic and gets skimmed. A table quoting the exact thought stops it, because the agent
recognises itself mid-sentence.

---

## In SpecForge specifically

**The prefix is a contract, not a naming scheme.**

```text
sfp-   a PRODUCT state.  Runs once. Opens with `sf context`, closes with `sf done`.
sf-    a FEATURE state.  Runs once per feature. Same two ends.
sfx-   a utility. Lives OUTSIDE the machine — sf does not know it exists.
```

**An `sfx-` must never learn to call `sf done`.** The day it does, it stops working outside a
SpecForge project, which is half of why it exists. The inverse bites harder: if `sf next` ever
named an `sfx-`, the orchestrator would dispatch somebody who never reports back, and the state
would sit still forever.

**State skills stay thin and compose the method.** The method lives in exactly one primitive, and
every skill that needs it composes that one:

```text
sfp-scout   →  sfx-grilling      sf-plan     →  sfx-think
sfp-po      →  sfx-prosa         sf-build    →  sfx-tdd · sfx-github
sfp-backlog →  sfx-triage        sf-cierre   →  sfx-documenter · sfx-journal · sfx-prosa
```

If you catch yourself re-explaining an interview inside a state skill, the method now has two
owners and they will drift. Move it into the primitive.

**Touching the catalogue has two obligations.** Adding, renaming, removing or re-homing a skill
means updating `sfx-mapa` (the router) and `docs/skills.md` (the human catalogue) in the same
change. A router still pointing at a skill that moved is confidently wrong exactly when somebody
is lost.

---

## Testing hygiene — four rules the bench paid for

From `banco/GUION.md`, where two runs came back green and measured nothing.

```text
①  the scenario and its answer NEVER go into any memory the agent can recall
②  old runs are NOT archived inside the bench directory
③  nothing that explains the experiment goes in the directory the agent walks
④  every run gets a NEW directory
```

**What went wrong is the part worth carrying.** Previous runs sat next to the scenario with their
answers in them; the agent found them and concluded on its own that it was time to implement.
Another time the answer was in `icm recall` — and the global `CLAUDE.md` **orders** the agent to
use ICM, so it had standing instructions to go fetch the answer it was supposed to investigate.

Both runs passed. Both measured nothing. **You were not testing the skill; you were testing what
the agent found lying around.** Rule ④ is the least obvious and cost the most: ICM names a memory
topic after the directory, so re-running in the same folder leaks the previous verdict into the
next run.
