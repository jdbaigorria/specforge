# Working rules — the section a cold subagent actually obeys

`## Reglas de trabajo` is the part of the constitution that changes what the ⑲ implementer does.
Architecture tells them where things go; this tells them how to behave when the spec runs out.

Keep it short. **A rule nobody could violate is not a rule, it is a comment.**

## The one worth seeding by default: minimal code

Unless Javier opts out, propose this rule. It is the code-parsimony counterpart of "no ceremony
without purpose", and it is the single rule that most reduces what gets written:

> **Código mínimo — subí la escalera antes de escribir código nuevo:**
> stdlib → algo nativo del lenguaje → una dependencia ya aprobada → una línea propia.
> Código necesario, no código ingenioso.

*(Credit: the **ponytail** project, https://github.com/DietrichGebert/ponytail, MIT.)*

The ladder matters more than the slogan. A subagent that reads "keep it simple" writes whatever
it was going to write; a subagent that reads "check the stdlib first, then a native construct,
then an approved dependency" has an order to follow.

**It composes with `dependencias_aprobadas`:** the third rung is *approved* dependency, so the
rule and the field reinforce each other instead of competing.

## Others worth asking about

Ask, do not assume. Each of these is a real fork where projects differ:

- **Errors** — wrapped and propagated, or handled at the boundary? Panics allowed anywhere?
- **Comments** — density, and in what language. This project comments the *why*, at length,
  in Spanish; another might want none. A subagent guesses wrong 50% of the time.
- **Tests** — table-driven or one per case? Real filesystem or mocks? Where do they live?
- **Concurrency** — allowed by default, or only where measured?
- **What "done" looks like** — this is the one people skip, and it is the one the ⑲ needs most.

## What does not belong here

- **Anything the machine already enforces.** Do not write "commit when the batch is done" — `sf`
  does the commit and there is no way to close a batch without one. A rule restating a gate is a
  rule that will drift from it.
- **Anything that is a decision, not a rule.** "We use Postgres" belongs in `## Stack y por qué`.
- **Abstract principles.** "Quality first." Nobody has ever changed a line of code because of it.
