---
name: sfp-po
description: >
  Turn a sealed brief into a PRD: actors, capabilities, non-functional constraints, scope and
  non-scope, external dependencies. State `prd` (step ⑦) of the SpecForge machine — invoked by
  the orchestrator when `sf next` returns `skill: sfp-po`, and it runs in a fresh subagent.
  Deliberately thin: it has exactly two readers, the constitution (⑧) and the backlog (⑨), and
  writes only what those two consume. Also usable standalone: "write the PRD", "turn this brief
  into a PRD", "/sfp-po".
---

# sfp-po

The ⑦. **Brief → PRD.** One input, one output, two readers.

```bash
sf context      # → .docs/brief.md
```

**You run in a fresh subagent.** You were not in the conversation that produced the brief, so
the brief is all you have — and that is the point. If something load-bearing is missing from it,
say so in the PRD as an open question instead of inventing it.

## The size is set by its two readers, and nothing else

```
in     .docs/brief.md
out    .docs/prd.md
       · actores                        who touches this, and in what role
       · capacidades                    what the product must be able to do
       · restricciones no funcionales   performance, security, scale, compliance
       · alcance / no-alcance           the boundary, both halves
       · dependencias externas          what this needs from outside itself

NOT    personas with a photo · user journeys · business metrics · go-to-market ·
       pricing · wireframes · competitive analysis (that was the brief's job)
```

> **Why the exclusions are hard rules and not taste.** The ⑧ reads this to decide the technical
> constitution; the ⑨ reads it to cut user stories. **Neither of them reads a persona or a
> pricing table.** A section no downstream step consumes is a section that ages, contradicts the
> code, and costs context in every envelope that carries it.

An earlier design kept the heavy PRD out on purpose, arguing it would become *"a third
translation layer between the brief and the specs"*. That argument held while nothing consumed
it. **It has two consumers now**, so the layer is real work, not translation — but it stays
exactly as thin as those two consumers need.

## Step 1: Read the brief, and mind the provenance

The brief tags its claims `retrieved` or `model-prior`. **Carry that distinction forward.** A
capability that rests on an unverified claim is a capability worth flagging — do not launder a
`model-prior` into a flat requirement by rewriting it in the imperative.

## Step 2: Extract the actors

Who interacts with this, in what role, and with what authority. Not demographics — **roles**,
because the ⑨ writes `Como <quién>…` from exactly this list.

## Step 3: Write the capabilities

What the product must be **able to do**, at the product level. One capability per line, in
plain language, and each one has to be the kind of thing you could later cut into stories with
acceptance criteria.

**Say WHAT, never HOW.** The how is the ⑧ (technical constitution) and the ⑬ (spec-design).
A PRD that names a library has skipped two states and pre-empted a decision that is not its own.

## Step 4: Constraints, scope, dependencies

- **Non-functional constraints** — and each one **stated so it could be measured**. "Fast" is
  not a constraint; "under 200ms at p95" is. If it cannot be measured it will not survive to a
  criterion (R5).
- **Scope and non-scope** — write **both**. The non-scope is the half that stops scope creep
  three states later, and it is the half people skip.
- **External dependencies** — services, APIs, data, accounts. What this needs that it does not
  own.

## Step 5: Write it and close

Write `.docs/prd.md` from `templates/prd.tmpl.md`, then:

```bash
sf done
```

**No gate and no stop.** The ⑦ has no decision point — `sf` checks the file exists and moves
straight to the ⑧. The hash it stores (`prd_hash`) is what later tells the ⑨'s stories they are
looking at a PRD that has since changed. It warns; it does not block.

## Rules

- Two readers only: the ⑧ and the ⑨. Write for them, not for a stakeholder.
- WHAT, never HOW. No libraries, no architecture, no file layouts.
- Every non-functional constraint stated so it could be measured.
- Non-scope is mandatory, not optional.
- Missing information becomes an open question in the PRD. Never an invention.
- No personas, journeys, metrics, pricing, or wireframes. Nothing downstream reads them.
- Run `sfx-prosa` before sealing. `P-1` especially: a capability sentence that would serve any
  other product has not described this one.
