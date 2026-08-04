# Update Feature Registry (gate ledger)

`specforge/features/<name>/feature.json` is the **source of truth** for that
feature's status and gates (F10) — one file per feature, so parallel branches
don't collide on a global ledger.
Each gate the user approved is recorded in the `gates[]` ledger — the auditable
record that a gate actually happened, not a claim in a markdown header.

## 1. Create/update the feature record

`feature.json` is **machine state — never write it by hand** (the hook denies a
direct write). The CLI is the only writer. Create the record and set its
lifecycle fields with `sf feature`:

```bash
sf feature add --feature=<name> [--lane=lite|standard] [--depends-on=a,b]
# …after the user approves the spec:
sf feature set-status --feature=<name> --to=approved
```

`sf feature add` creates the record in status `planned` with an empty `gates[]`
ledger (approvals are sealed by the CLI in step 2). `--depends-on` lists the
feature names this one needs first (captured from the roadmap during `--all`, or
stated by the user); it drives the critical-path ordering and blocker view in
`/sf-status` (F37). `set-status`/`set-lane` move the lifecycle; both validate the
transition and enforce the serial flow (one active feature at a time, F22).

## 2. Seal each approved gate with `sf gate approve`

**Do not hand-write `approve` gate entries.** For each phase the user approved,
run:

```bash
sf gate approve --feature=<name> --phase=requirements
sf gate approve --feature=<name> --phase=design
sf gate approve --feature=<name> --phase=tasks
```

This appends the gate entry AND **seals a content hash** of the artifact. The
hash is what lets `sf status --artifacts` detect a *silent edit* later (an
artifact changed after approval without re-gating → it goes **stale**). Letting
the CLI compute the hash over the real file is the whole point — an LLM-written
hash would not be trustworthy.

The lane gate (no artifact) is also recorded this way:
`sf gate approve --feature=<name> --phase=lane --comment=standard`.

For a **lite** feature, `"lane": "lite"` and the gates are just `lane` → `change`
→ `plan`/`wave` → `verdict` (no separate requirements/design/tasks gates).

Only seal gates the user actually gave. A `reject` or `change` is **not** an
approval — don't record it as a gate; just iterate on the artifact and re-present
(the gate is only sealed once the user approves). If a downstream gate reopened
(F23), move the affected feature's status back with `sf feature set-status`; the
upstream re-approval (a fresh `sf gate approve`) is what flags the downstream as
stale.

Append to `specforge/history.md`:

```
## [date] — Feature proposed: <name>
- Workflow: requirements-first
- Requirements: [count] | Tasks: [count]
```

Inform the user: "Feature `<name>` approved. Use `sf-build <name>` to start implementation."
