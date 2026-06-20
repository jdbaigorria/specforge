# Update Feature Registry (gate ledger)

`specforge/features.json` is the **source of truth** for status and gates (F10).
Each gate the user approved is recorded in the `gates[]` ledger — the auditable
record that a gate actually happened, not a claim in a markdown header.

## 1. Create/update the feature record

Write the feature's metadata (the CLI does not author this part). Start `gates`
empty — approvals are sealed by the CLI in step 2.

```json
{
  "name": "<feature-name>",
  "status": "approved",
  "workflow": "requirements-first",
  "lane": "standard",
  "depends_on": [],
  "created": "<date>",
  "completed": null,
  "gates": []
}
```

`depends_on` lists the feature names this one needs first (captured from the
roadmap during `--all`, or stated by the user). It drives the critical-path
ordering and blocker view in `/sf-status` (F37). Default `[]`.

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

Only seal gates the user actually gave. For a `reject` or `change`, hand-write
that entry with the request in `comment` (those carry no hash). If a downstream
gate reopened (F23), mark the affected feature `status` back; the upstream
re-approval (a fresh `sf gate approve`) is what flags the downstream as stale.

Append to `specforge/history.md`:

```
## [date] — Feature proposed: <name>
- Workflow: requirements-first
- Requirements: [count] | Tasks: [count]
```

Inform the user: "Feature `<name>` approved. Use `sf-build <name>` to start implementation."
