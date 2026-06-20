# Update Feature Registry (gate ledger)

Update `specforge/features.json`. The registry is the **source of truth** for
status and gates (F10). Each gate the user approved in this run is appended to
the `gates[]` ledger — this is the auditable record that a gate actually
happened, not a claim in a markdown header.

```json
{
  "name": "<feature-name>",
  "status": "approved",
  "workflow": "requirements-first",
  "lane": "standard",
  "depends_on": [],
  "created": "<date>",
  "completed": null,
  "gates": [
    { "phase": "lane",         "result": "approve", "by": "user", "at": "<iso-8601>", "comment": "standard" },
    { "phase": "requirements", "result": "approve", "by": "user", "at": "<iso-8601>", "comment": null },
    { "phase": "design",       "result": "approve", "by": "user", "at": "<iso-8601>", "comment": null },
    { "phase": "tasks",        "result": "approve", "by": "user", "at": "<iso-8601>", "comment": null }
  ]
}
```

For a **lite** feature, `"lane": "lite"` and the gates are just `lane` → `change`
→ `plan`/`wave` → `verdict` (no separate requirements/design/tasks gates).

`depends_on` lists the feature names this one needs first (captured from the
roadmap during `--all`, or stated by the user). It drives the critical-path
ordering and blocker view in `/sf-status` (F37). Default `[]`.

Record the real result of each gate: `approve`, `reject`, or `change` (with the
request in `comment`). Never write a gate entry the user did not actually give.
If a downstream gate reopened (F23), mark the affected feature `status` back and
the stale gate is re-recorded on re-approval.

Append to `specforge/history.md`:

```
## [date] — Feature proposed: <name>
- Workflow: requirements-first
- Requirements: [count] | Tasks: [count]
```

Inform the user: "Feature `<name>` approved. Use `sf-build <name>` to start implementation."
