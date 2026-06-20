# Minimal-code review (parsimony)

The rubric for auditing the constitution principle **`P-min`** ("minimal code:
climb the ladder before writing new code"). SpecForge guarantees the code *matches
the spec*; this is the orthogonal axis — is it the *minimal* code that does so?

> Adapted from the **ponytail** project
> (https://github.com/DietrichGebert/ponytail, MIT) — we adopt the *idea* (the
> decision ladder), not its code.

## The ladder (what build was supposed to climb)

```
Does this even need to exist?            → if no, it shouldn't have been written
  └─ Does the stdlib already do it?      → should have used it
       └─ Is there a native language/framework feature?  → should have used it
            └─ Does an ALREADY-installed dependency do it?  → should have used it
                 └─ Is it a one-liner?   → should be one line
                      └─ only here: minimal hand-written code is justified
```

## What counts as a finding (over-engineering)

For each new/changed unit of code, ask:

- **Reinvented the wheel** — hand-rolled something stdlib / an installed dep
  already provides (date parsing, JSON, sorting, HTTP, path handling…).
- **Premature abstraction** — an interface/factory/base-class with a single
  implementation; a generic for one type; indirection no caller needs yet.
- **Speculative generality** — config knobs, hooks, or parameters nothing uses;
  "we might need it later".
- **Unjustified dependency** — a new dependency added for what the stdlib or an
  installed dep already does.
- **Dead/duplicated code** — code no requirement traces to, or a second
  implementation of something that already exists in the repo.
- **Ceremony** — wrappers, managers, and layers that add naming, not behavior.

A finding must be **specific and cite the code** (file:symbol) and name the
ladder rung that was skipped and the minimal alternative — not "could be simpler".

## What is NOT a finding

- Style or naming differences. Parsimony ≠ taste.
- Code the spec genuinely requires (don't cut to hit a metric — `P-min` never
  overrides correctness, EARS, or security).
- Abstractions with **two or more** real current uses.
- Tests. Thorough tests are not over-engineering.

## Verdict mapping

- A clear `P-min` violation is a valid reason for **REVISE** in `sf-check`, and a
  **fail** verdict when audited at a phase gate (`sf-check --phase=build`).
- Net-new lessons (e.g. "we keep reimplementing X that stdlib gives") belong in
  `sf journal add` — they feed future audits (the judge's rules grow over time).
