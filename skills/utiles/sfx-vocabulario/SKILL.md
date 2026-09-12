---
name: sfx-vocabulario
description: >
  The glossary primitive. Build and sharpen the project's business vocabulary: challenge fuzzy
  terms, resolve words doing two jobs, and keep `.docs/vocabulario.md` as the one place a word
  means one thing. This is the SINGLE OWNER of the domain glossary in SpecForge. Triggers: a term
  that means two things, a word nobody has defined, "what do we call this", or any composing skill
  calling it. Merely READING the glossary is not this skill — that is a one-line habit any skill has.
---

# sfx-vocabulario

**The problem it solves:** two artifacts using the same word for two different things. It costs
nothing while everyone is in the same conversation, and it costs a rewrite three states later, when
the backlog says "cuenta" meaning one thing and the spec says "cuenta" meaning another.

## When it runs

**Not on a schedule.** It runs the moment a word wobbles, in the middle of whatever else was
happening:

- someone uses a term that clashes with what the glossary already says → **stop and call it out**:
  *"el glosario dice que 'cancelación' es X, pero parece que estás diciendo Y. ¿Cuál es?"*
- someone uses a vague or overloaded word → propose the precise one: *"decís 'cuenta': ¿el Cliente
  o el Usuario? Son dos cosas distintas"*
- the code disagrees with what was just said → surface it, do not pick a side quietly

**Stress-test with concrete scenarios.** When two concepts touch, invent the awkward case that
forces the boundary to be said out loud. *"¿Qué pasa si cancela la mitad del pedido?"* settles more
vocabulary than any definition does.

## Where it writes

`.docs/vocabulario.md`. **One entry per term, and the entry says what it is NOT** — that half is
where the ambiguity actually dies.

```yaml
---
terminos: 7
---
```

```markdown
## Corrida
Una ejecución completa de la máquina sobre una idea, de punta a punta.

**No es** una vuelta del bucle de implementación — eso es un *lote*.
```

**Create it lazily: only when there is something to write.** No empty file, no placeholder, no
"pending terms" section. If nothing wobbled, the file does not exist, and that is correct.

**Write it the moment the term is settled**, right there in the middle of the round. Do not batch
them for the end — the reason a term got sharpened is the part that evaporates first.

## What does NOT go in here

`vocabulario.md` is **a glossary and nothing else.** Devoid of implementation detail. It is not a
spec, not a scratchpad, not a decisions log.

**And no ADRs.** Matt's `domain-modeling` writes architecture decision records next to the
glossary; SpecForge does not, on purpose: the ⑫ already produces `decision.md` per feature, with
three options mandatory. Two places for the same thing is worse than one.

## Who reads it afterwards

Everything downstream, if it exists: `sf context` serves it as one more route, the same way it
serves the constitution. That is the whole point — the vocabulary is built once, in the interview,
and every later state inherits it instead of re-inventing the words.

## Rules

- It runs when a word wobbles, not on a schedule.
- Every entry says what the term **is not**.
- Written the moment it settles, never batched.
- Lazily created. No file until there is a term.
- Glossary only. No implementation, no decisions, no ADRs.
