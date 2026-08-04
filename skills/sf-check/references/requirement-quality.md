# Requirement quality review (built-in)

The rubric for auditing the **intrinsic quality of a requirement's text**, applied
in the `requirements` phase of **every** project.

Unlike `minimal-code.md`, this rubric is **not opt-in**: it does not audit a
constitution principle. Whether a requirement is ambiguous, or is secretly two
requirements, or cannot be verified at all, does not depend on the project — it
is universal, so it ships from the factory. A thin constitution must not produce
an audit that audits nothing.

`sf context for-judge --phase=requirements` names this rubric in its `rubrics`
field on every run. Apply it in addition to whatever principles are in scope.

## The six defects

Each finding uses the id in the first column, so it lands in `audit.json`
alongside the constitution's own verdicts.

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
you cannot report it — same citation rule the phase judge already enforces.

## What is NOT a finding

- **Prose style, voice, or length.** `RQ-*` is about semantics, not writing.
- **Vagueness the design phase legitimately resolves.** A requirement is not
  supposed to pin the implementation. *"Persist the session"* is not `RQ-1`
  because it doesn't name a database.
- **Terminology that differs from `domain.json` when the requirement defines it
  in this artifact.** That's new vocabulary, not inconsistency.
- **A requirement with no `source`.** Provenance is checked upstream and
  computably; the judge does not opine on it.
- **Missing requirements.** This rubric judges the requirements that are here.
  Coverage of the problem space is a different question.

## Verdict mapping

- Any `RQ-*` finding is a **fail** verdict for that rule when audited at the
  requirements gate, and a valid reason for **REVISE**.
- Because `audit.phase` defaults to `nudge`, a fail surfaces to the human and
  does not block. The human decides; the finding is already in `audit.json`.
- Recurring findings across features (e.g. the same term used two ways in three
  specs) belong in `sf journal add` — they are candidates for promotion to a
  project invariant via backprop.

## Relationship to the requirement model (forward note)

When the v2 requirement model lands, two of these six become **computable**
rather than judged: scenario-level contracts decide `RQ-3` (verifiable)
deterministically, and source refs decide provenance.

**Do not delete `RQ-3` when that happens.** It stops being the only check and
becomes the second pair of eyes on what the contract already measured — a
scenario can be mechanically well-formed and still be unverifiable in substance.
The other five stay irreducibly a matter of judgment.
