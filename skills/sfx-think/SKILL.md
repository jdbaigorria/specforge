---
name: sfx-think
description: >
  Weigh two or more NAMED options that are already on the table: compare them, follow each to
  its consequences, reach a documented conclusion. The options are the precondition — if you
  cannot name them yet, this is not the skill. Triggers: "/sfx-think", "/think <topic>",
  "/debate", "should I use X or Y", "pros and cons of X", "evaluate this approach". A raw
  product idea is NOT one of them: "quiero armar X", "tengo una idea", "is this worth
  building", "what if we", "I'm considering" all go to `sfp-scout`, which starts by researching
  whether the thing already exists — arriving here instead skips that and debates an idea
  nobody checked. Composed by `sf-plan` for the ⑫ decision, and by `sfx-grilling` when no fact
  separates the options.
---

# Think

Debate ideas freely. Explore options. Reach a conclusion. Document it.

This skill is the space between "I have a vague idea" and "I'm ready to specify."
It's not propose (no requirements, no tasks). It's not explain (not teaching).
It's not grill-me (not stress-testing an existing plan). It's thinking out loud
with a partner who pushes back, offers alternatives, and helps reach clarity.

**Always produces a structured conclusion.** Composed by `sf-plan` it becomes `decision.md`; standalone it goes wherever the user asks.

## Step 1: Understand What to Think About

- `/think <topic>` → use the topic
- `/think` alone → ask: "What's on your mind?"
- User describes a dilemma → use it

Identify the type of thinking needed:

- **Decision:** "Should I use X or Y?" → evaluate options against criteria
- **Exploration:** "What if we did X?" → follow the idea to its consequences
- **Validation:** "Is this approach sound?" → stress-test the reasoning
- **Strategy:** "How should I approach X?" → map the solution space

## Step 2: Debate

This is conversational. No rigid structure. But follow these principles:

### Play devil's advocate

When the user leans toward an option, challenge it. Not to be difficult — to
make sure they've considered the costs. "You're leaning toward GraphQL. What
happens when you need real-time subscriptions with 10k concurrent users on
your current infra?"

### Bring alternatives the user hasn't considered

Don't just evaluate what the user proposes. Offer at least one option they
didn't think of. "Have you considered doing neither? A simple REST endpoint
with sparse fieldsets might give you 80% of GraphQL's benefit with 0% of
the complexity."

### Ground in specifics

Steer away from abstract debate. "GraphQL is more flexible" is abstract.
"GraphQL lets your mobile client fetch user + posts + comments in one request
instead of three, saving ~400ms on 3G" is specific.

### Know when to stop

The goal is a conclusion, not exhaustive analysis. If after 5-8 exchanges
the direction is clear, move to conclusions. Don't keep debating for the
sake of thoroughness.

If after 10+ exchanges there's no convergence, say so: "We've been going
back and forth. Let me summarize where we are and what's still unresolved."

### Read context when relevant

If the debate touches on the current project:
- `.docs/constitucion.md` — stack constraints
- `.docs/constitucion.md` — principles that might settle the debate
- Existing code — reality check against abstractions

Don't force context reading for purely conceptual debates.

## Step 3: Converge

When the debate reaches a natural conclusion (or the user says "enough"):

Summarize the outcome. Present it before writing the artifact:

> Here's where we landed:
> - **Conclusion:** {what we decided/concluded}
> - **Key reason:** {the argument that tipped the scale}
> - **What we ruled out:** {and why}
> - **Open questions:** {if any remain}
>
> Save this to a file? Where? [path / N]

Default: yes (unlike explain/grill-me where default is no). The whole point
of think is to capture the reasoning for future reference.

## Step 4: Write Artifact

Use `templates/think.tmpl.md`. Composed by `sf-plan`, this IS `decision.md` — and there it needs exactly three options.

The artifact captures:
- The question/topic
- Options explored with pros/cons
- Arguments that shaped the conclusion
- The conclusion itself with rationale
- Alternatives rejected and why
- Next steps (what to do with this conclusion)

## Rules

- This is a conversation, not a lecture. Listen more than you talk.
- Always challenge the user's initial leaning — not to change their mind but
  to make sure they've stress-tested it.
- Offer at least one alternative the user didn't consider.
- Ground abstract debates in concrete specifics from the user's context.
- Don't force a conclusion. "We don't know enough yet" is valid — document
  what's needed to decide.
- Keep the debate focused. If it branches into 3 topics, pick one, finish it,
  then move to the next.
- 5-10 exchanges is the sweet spot. Under 5 is too shallow. Over 10 usually
  means you're going in circles.
- The artifact is the deliverable — make it useful enough that someone reading
  it 3 months later understands the reasoning without reading the full debate.
