---
name: sfx-buscar
description: >
  The research primitive. Answer a factual question against real sources, cheapest tier first, and
  leave the findings as evidence with provenance on every claim. Level 0 needs no key and works in
  any harness; level 1 is the MCPs. This is the SINGLE OWNER of the research method in SpecForge —
  every skill that needs a fact composes this one. Triggers: "research this", "does this already
  exist", "find out whether", or any composing skill calling it.
---

# sfx-buscar

**One job: turn a question into evidence.** Not an opinion, not a summary of what you remember —
evidence, with a link, or an honest label saying there is none.

## Start at level 0. Always.

**Measured 2026-09-05.** SpecForge went blind on its first real run because `sfp-scout` sent the
model straight to Tavily and the GitHub MCP — the tier that needs a key and per-harness setup. The
tier that works **everywhere, with no key**, was not named anywhere. One call to
`registry.npmjs.org` returned three real competitors, instantly.

| The question | Level 0 — no key, any harness |
|---|---|
| does this package already exist? | `registry.npmjs.org/-/v1/search?text=…` · PyPI · crates.io · pkg.go.dev |
| does this repo exist, and what do its issues say? | the public GitHub API (60 req/hour, no token) |
| what does this specific page say? | `curl` |
| what do these library docs say? | `context7` (keyless MCP) |

> **The rule: start with what works with nothing configured. If level 0 already answers the
> question, level 1 is not needed.**

**And the honesty that cannot be skipped:** level 0 does **not** replace level 1. Registries answer
*"does something with this name exist?"*. They do **not** answer *"is anybody complaining about
this on Reddit?"*. Research that only had level 0 can map the landscape and **cannot** bring demand
signals. **That goes in the evidence file, not swept under the rug.**

The full tier table, what needs a key, and how to set it up: [references/niveles.md](references/niveles.md).
How to hunt demand signals once you are at level 1: [references/senales.md](references/senales.md).

## Provenance on every claim — no exceptions

```
retrieved      pulled a source this session      ALWAYS with the link
model-prior    from training, unverified         allowed, but it must say so
probado        built it and saw it work          left by sfx-prototipo
```

- **A claim with no tag is a bug.** Tag it or cut it.
- **`retrieved` without a link is `model-prior` wearing a costume.** Demote it.
- **What you did NOT find is data too.** *"Searched npm for X, nothing came back"* is a finding —
  write it. An empty search is not a failed search.

## Never fabricate a landscape

If no research tool is reachable at all — not even `curl` — **stop and say so**. Do not fill the
gap from training data.

> A landscape built from the model's imagination is worse than none, because it *feels* like
> research.

Two honest ways out: name what needs enabling and stop, or — with explicit consent — run a
**degraded pass** where every claim is `model-prior` and the evidence is stamped `evidencia: baja`.
That stamp is the only way evidence with no cited source gets past the ⑥, and it is a declaration,
never a shortcut.

## What it leaves

`.docs/evidencia.md`, appended to across a session — one file, not one per question:

```yaml
---
retrieved: 17
model_prior: 1
probado: 0
links: 13
# evidencia: baja   ← ONLY on a degraded pass
---
```

Body, one block per question asked:

```markdown
## ¿Ya existe un CLI que calcule costos de Claude Code?

- Hay tres paquetes publicados que hacen exactamente esto. [retrieved]
  - https://registry.npmjs.org/-/v1/search?text=claude+code+usage+cost
- El más usado no separa costo por sesión. [retrieved]
  - https://github.com/<owner>/<repo>/issues/12
- La categoría se siente saturada. [model-prior — sin verificar]
- Busqué quejas en Reddit y no llegué: hace falta nivel 1. [no evaluable]
```

**`no evaluable` is a real outcome and it has to be written.** *"I could not check this"* changes a
decision. A green that means "I did not look" is worse than a red.

## Rules

- Level 0 first, always. Level 1 only if level 0 did not answer it.
- Provenance on every claim. `retrieved` without a link gets demoted.
- What you did not find gets written down.
- No tools → no invented landscape. Stop, or declare `evidencia: baja`.
- One file: `.docs/evidencia.md`. Append, never overwrite someone else's block.
