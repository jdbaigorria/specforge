# Archive: Sync and Close

## Procedure

### 1. Copy to Archive
```
specforge/archive/<YYYY-MM-DD>-<feature-name>/
├── requirements.md     # Living contract (NOT frozen — see below)
├── design.md           # Final approved version
├── tasks.md            # All tasks marked [x]
├── review.md           # Verdict document (historical)
├── trace.json          # Structured matrix — the live link to code (F33)
└── progress/           # All wave logs
    ├── plan.md
    ├── wave-0.md
    ├── wave-1.md
    └── ...
```

**Archive seals with a live link, it does not freeze (F33).** Git already holds
the historical snapshot (this archive commit). So the archived spec stays a
*living document*: `review.md` is the historical verdict, but `requirements.md`
remains the current contract and `trace.json` is its live link to the code. To
change a shipped feature, use `sf-amend <name>` — never hand-edit the archive
into a parallel truth. Drift detection (`sf doctor --drift`) reads
`trace.json` to catch the spec and code diverging over time.

### 2. Update Feature Registry

In `features.json`, set the final status, completion date, and append the
verdict gate to the `gates[]` ledger (F10). Don't overwrite earlier gates —
append:
```json
{
  "name": "<feature-name>",
  "status": "done",
  "completed": "<date>",
  "gates": [
    "... earlier requirements/design/tasks/plan/wave gates ...",
    { "phase": "verdict", "result": "approve", "by": "user", "at": "<iso-8601>", "comment": "<verdict notes if APPROVE WITH NOTES>" }
  ]
}
```

`result` records the real verdict reply: `approve`, `approve-with-notes`, or
`revise`. A `revise` verdict means the feature is NOT archived — it loops back to
`sf-build`, so no `completed` date is set.

### 3. Append to History

In `specforge/history.md`:
```markdown
## [date] — Feature completed: <name>
- Verdict: [verdict]
- Workflow: [requirements-first / design-first / from-code]
- Requirements: [count]
- Tasks: [completed]/[total]
- Waves: [count]
- Duration: [created → completed]
```

### 4. Clean Up (optional)

The feature folder in `specforge/features/<name>/` can be:
- **Kept** as the active reference (recommended for recent features)
- **Removed** if the archive is sufficient (saves clutter for old features)

Let the user decide. Default: keep.

### 5. Check for Backprop

After archiving, check `specforge/history.md` recurring issues table.
If any pattern reached 3 occurrences during this check, trigger backprop
promotion (see `references/backprop.md`).
