# Criterion quality — the rubric for the ⑨

The rubric for the **intrinsic quality of a criterion's text**. Run it over your own `CA-#`
before closing the ⑨.

**It is universal, not project-specific.** Whether a criterion is ambiguous, or is secretly two
criteria, or cannot be verified at all, does not depend on what is being built — so it applies
everywhere, no matter how thin the constitution is.

**Why here and not at review time.** A defective criterion is cheapest to fix now, while it is
one line of text. Three states later it has grown a task, a test name and an implementation, and
the ㉑ reviewer is stuck asking a question that cannot be answered. `sf` cannot catch any of
these — every one of them is judgment (R3) — so this rubric is the only thing standing between a
vague criterion and a feature that ships half-done.

## The six defects

| id | Defect | It's a finding when |
|---|---|---|
| `RQ-1` | **Ambiguous** | The text admits two readings that would produce different implementations. Unquantified qualifiers (*"fast"*, *"appropriate"*, *"if needed"*, *"as required"*), or a pronoun with no single antecedent. |
| `RQ-2` | **Not singular** | It is really two requirements: an `and`/`or` joining two behaviors that could be accepted or rejected independently. |
| `RQ-3` | **Not verifiable** | No procedure could decide pass/fail. *"The system must be maintainable"* with no metric and no scenario. |
| `RQ-4` | **Contradiction** | Two requirements in this artifact cannot both hold. |
| `RQ-5` | **Inconsistent terminology** | The same concept under two names, or the same name for two concepts, within this artifact. |
| `RQ-6` | **Incomplete** | A trigger with no response, a condition with no alternative branch, a verb with no object. |

## How to run it — anti-omission rules

1. **Run all six, every time.** An item that does not apply is reported as
   `N/A — {reason}`. It is never silently skipped.
2. **The output is always a report.** If there are no findings, say so
   explicitly — *after* having looked for each of the six. "Looks good" without
   the six-item breakdown is not a valid verdict; it is indistinguishable from
   not having checked.

A finding must **quote the exact requirement text** it applies to. No quote means
you cannot report it: a finding you cannot point at is an opinion.

## What is NOT a finding

- **Prose style, voice, or length.** `RQ-*` is about semantics, not writing.
- **Vagueness the ⑬ legitimately resolves.** A criterion is not
  supposed to pin the implementation. *"Persist the session"* is not `RQ-1`
  because it doesn't name a database.
- **New vocabulary the story itself defines.** That is not inconsistency.
- **Missing criteria.** This rubric judges the criteria that are here. Whether the story covers
  its whole capability is a different question — and the ⏸ is where Javier answers it.

## What to do with a finding

**Fix it now.** You are writing these criteria; there is no gate to escalate to and no ticket to
open. A criterion with an `RQ-*` defect either gets rewritten or gets deleted.

`RQ-3` (not verifiable) deserves the sharpest hand, because it is R5 in rubric form: **if you
cannot write the test, it is not a criterion.** Rewrite it into something observable, or drop it
and say you dropped it — never leave it as prose that nobody can ever check.

If the same defect keeps appearing across features — the same term used two ways in three specs
— that is not a criterion problem, it is a constitution problem. It belongs in the journal at
the ㉓, as a candidate for promotion to a project rule.
