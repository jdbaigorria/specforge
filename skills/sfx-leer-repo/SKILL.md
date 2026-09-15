---
name: sfx-leer-repo
description: >
  Derive what a project already IS by reading its code — for a repository that exists before
  SpecForge does. Produces `.docs/repo/`: what this is, how it is tested, the conventions actually
  in use, the flows that already exist, and the decisions the code already made and nobody wrote
  down. Use when arriving at an unfamiliar or pre-existing codebase, when a feature or a bug has to
  enter a project with no constitution, or on "what are the conventions here", "read this repo",
  "onboard me". NOT a code review and not an opinion about the architecture: it reports what IS,
  with provenance.
---

# sfx-leer-repo

**A project with three years of code should not have to invent a brief before it can accept a
ticket.** What it already is can be read.

**Framing, non-negotiable:** this reports **what is**, not what should be. The moment you write
"this should use X", you left this skill and entered an opinion — and an opinion dressed as a
derived rule makes the project worse with authority.

**It runs without the user.** Every line below is reading, counting or running. The user reads the
result once, corrects it, and does not get asked again.

## The four files

```
.docs/repo/
   ├── repo.md            what this is · how it is tested · how it is built
   ├── convenciones.md    how things are done here
   ├── flujos.md          what already exists
   └── decidido.md        what the code already settled and nobody wrote down
```

Four files, not one document — a rule nobody can find is a rule nobody follows.

### repo.md — the ground

What the project is for in one line (from its README, its entry point, its CLI help), the
languages and manifests, **the test command and whether it passes today**, how it builds, and how
it is run. Run the tests. A test command that nobody verified is a guess.

### convenciones.md — how things are done here

The patterns that actually repeat: naming, directory layout, error handling, logging, how tests
are named and where they live, what the commit messages look like.

**A convention is a count, not an impression.** Do not write "errors are wrapped here" — write it
with the number and a couple of places to go look.

### flujos.md — what already exists

The map of what the project already does, entry point by entry point. This is what makes a request
answerable: if there is an existing flow to go read, a change is bounded; if there is nothing, it
is not. That is the V-2 of the feature rubric, and it is a fact about the repo, not a feeling
about yourself.

### decidido.md — what the code already settled

The one that is worth the most and the one that is easiest to get wrong. Decisions the code has
already taken and nobody wrote down: a library deliberately not used, a pattern abandoned halfway
and why, a compatibility floor, something that looks wrong and has a reason in the history.

`git log` and `git blame` are the source here. A decision you cannot date is a guess.

## Provenance — every single line carries it

This is the difference between a derived rule and an invented one:

```
doc        a document in the repo says it                    strong
codigo:N   it is in the code, in N places — with 2 examples  strong, with the count
commit     git log / blame explains it — with the sha        strong
una-vez    seen once, no pattern                             weak
supuesto   I am guessing                                     NOT VERIFIED — it is a ?
```

**A `supuesto` is not a rule.** It goes in its own section, marked, and the user resolves it or it
gets dropped. Never mix it in with the rest.

> Read `commandcode.ai/docs/taste` on 2026-09-15: their learnings carry a confidence number
> between 0 and 1. A confidence score assigned by the same model that wrote the rule **cannot be
> checked** — and that failure has already been measured here (2026-09-05: a model sealed a verdict
> with total confidence and zero sources). So: not how sure you feel, but **how you know**.

## Staleness — say when the picture was taken

Every file records the commit it was derived from:

```yaml
---
derivado_de: <sha>
fecha: 2026-09-15
reglas: 14
supuestos: 2
---
```

A rule derived from code that has changed a lot since is a rule that is now lying, which is worse
than no rule. Recording the commit is what makes "this needs another look" answerable without a
model watching over your shoulder in the background.

## The acta

```
───────────────────────────────────────
📖 <el proyecto>, leído
Tests: <el comando>  —  <pasan | fallan N | no los pude correr>
Convenciones: N reglas  ·  Flujos: N  ·  Decisiones: N
Supuestos sin confirmar: N   ← esto lo mirás vos
Derivado del commit <sha>
───────────────────────────────────────
```

Then stop. The user reads it once and corrects the guesses. That is the only interruption this
skill is allowed.

## Rules

- Report what IS, never what should be. An opinion here is a bug.
- Every line carries provenance. A `supuesto` is not a rule and never mixes in.
- A convention is a count with examples, not an impression.
- Run the test command. Do not copy it out of a README and call it verified.
- Record the commit. A picture with no date cannot be known to be stale.
- Four files. Do not write a fifth because something did not fit — make it fit or drop it.
- One interruption: the acta.
