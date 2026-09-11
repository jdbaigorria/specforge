---
name: sfx-prosa
description: >
  The prose primitive. Cut the tells out of a written document — the sentence that would serve any
  other project, the filler connector, the passive with a known actor — against a numbered rule
  set other skills cite by number. Written for Spanish, which is what this project's artifacts are
  written in. Triggers: "/sfx-prosa", "limpiá esto", "esto suena a IA", "tighten this doc", or any
  long prose artifact before it ships. Standalone: works on any text, with or without SpecForge.
---

# sfx-prosa

**Edit the text. Do not rewrite what it says.**

Every rule below has a stable number, and other skills cite it by number (`P-1`, `P-4`). **A
removed rule leaves a gap** — the numbers never get reused, because a skill somewhere points at
one.

## Why this is not a translation of an English checklist

The well-known lists are calibrated for English: *delve*, *tapestry*, *showcase*, *boasts*. None
of those ever appear in a document written in Spanish, so a translated list is a skill that never
fires.

**These rules come from this project's own documents.** The reference corpus is `specs/`, and that
is deliberate: a prose skill that does not sound like `que-sobrevive.md` will rewrite the repo in
somebody else's voice, which is worse than leaving it alone.

## The rules

### P-1 — The sentence that would serve any other project

**The most important one, and the only one worth running alone.**

> If the sentence could appear unchanged in another project's docs, it says nothing about this
> one. Cut it.

The test is mechanical: ask what the sentence tells the reader to do or know. If you cannot
restate it as a concrete instruction, a fact, or a number, it is filler wearing a suit.

```
✗  "el sistema está diseñado para ser robusto y escalable"
✓  "una corrida de 40 features entra en 12 MB de estado.json"

✗  "la arquitectura facilita el mantenimiento"
✓  "mover un artefacto de lugar rompe un solo paquete: docs/"
```

### P-2 — The filler connector

`cabe destacar`, `es importante mencionar`, `vale la pena señalar`, `en el mundo de`, `a la hora
de`, `en este sentido`, `dicho esto`, `por otro lado` used as a paragraph starter with nothing
opposed.

Delete. The sentence after it is the sentence.

### P-3 — "No sólo X sino también Y"

State the point directly. The construction promises a contrast and almost never delivers one.

### P-4 — Passive voice with a known actor

Catch `se valida`, `se comprueba`, `es leído por`, `fue diseñado para`. **Name who does it.**

```
✗  "los criterios se cuentan en la compuerta"
✓  "la compuerta cuenta los criterios"
```

Passive is fine only when the actor is genuinely unknown or genuinely does not matter.

### P-5 — The adverb propping up a weak verb

`significativamente`, `notablemente`, `considerablemente`, `rápidamente`, `fácilmente`. Either the
measured number or a stronger verb.

```
✗  "mejora significativamente el tiempo de arranque"
✓  "el arranque pasó de 1.8s a 0.3s"
```

### P-6 — The long dash

Use periods or commas. If a thought needs separating, end the sentence. (Applies to prose. **Not**
to the ASCII diagrams and box rules this repo uses in code comments and headings — those are
structure, not punctuation.)

### P-7 — The colon as a mid-sentence connector

Fine before a list or an example. Not as a hinge in the middle of a sentence. Rewrite so the point
stands without the crutch.

### P-8 — Bold on every proper noun

Bold marks what the reader must not miss. Bolding every filename and acronym marks nothing.

### P-9 — Over-compression

Dropped articles, verbless fragments, arrows standing in for verbs. Whole sentences, with their
articles.

```
✗  "parser rechaza fecha mala → exit 2, no escribe"
✓  "el parser rechaza una fecha inválida, sale con código 2 y no escribe nada"
```

### P-10 — The generic closing

*"El futuro es prometedor"*, *"las posibilidades son amplias"*, *"queda mucho por hacer"*. Either
the specific next step with its name, or nothing.

## How to run it

1. Read the whole document first. **You are editing, not drafting** — the argument is not yours to
   change.
2. Pass rule by rule. P-1 first: it usually deletes the sentences the other rules would have spent
   time polishing.
3. Self-audit at the end: *"what in this still reads as machine-written?"* Fix what you find, even
   if no rule above names it.
4. **Report what you cut**, by rule number, so the author can push back on a specific call.

## What it does NOT touch

- **Claims, numbers, links, and citations.** If a sentence is wrong, that is a finding, not an
  edit. Say it separately.
- **The structure.** Headings, order and sections are the author's.
- **Code, diagrams, and box rules.** P-6 and P-9 are about prose. An ASCII diagram is doing a job
  no sentence does.
- **Other people's quotes.** Verbatim is verbatim, tells and all.

## Where it composes

- **`sf-cierre`** — the ㉓ documentation, the half a human reads.
- **`sfp-po`** — the PRD.
- **`sfp-constitucion`** — the body, not the frontmatter.

**Not `sf-plan`.** `spec-design.md` is read by an implementer and is already governed by its own
anti-N/A rule: a section only exists if it changes a decision. Running a style pass over it adds a
step and removes nothing.

## Rules

- Edit, never rewrite. The argument belongs to whoever wrote it.
- P-1 first. It deletes what the rest would have polished.
- A wrong claim is a finding, not an edit.
- Quotes stay verbatim.
- Report the cuts by rule number.
- Rule numbers are stable. A dead rule leaves a gap; it never gets reused.
