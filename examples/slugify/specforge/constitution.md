# Constitution — texttools

## Identity

**What:** A small, dependency-free Python library of string utilities.
**Who:** Python developers who want a few well-tested text helpers without pulling
a large framework.
**Why:** Common text transforms (slugify, truncate, normalize) are re-implemented
badly in every project. One tiny, correct, tested library fixes that.

## Principles

<!-- Project-specific only. Universal engineering principles live in AGENT.md. -->

1. **Zero runtime dependencies** — the standard library only. A utility lib that
   drags in dependencies defeats its purpose.
2. **Unicode-correct by default** — text functions handle non-ASCII input
   sensibly, never mangle it silently.
3. **Pure functions** — no global state, no I/O. Every function is input → output.

## Constraints

- Python 3.9+.
- 100% test coverage on public functions.

## Anti-goals

- Not a full NLP toolkit. No tokenization, stemming, or language models.
- No CLI. This is a library, imported in code.
