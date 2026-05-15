# Archive: Sync and Close

## Procedure

### 1. Copy to Archive
```
specforge/archive/<YYYY-MM-DD>-<feature-name>/
├── requirements.md     # Final approved version
├── design.md           # Final approved version
├── tasks.md            # All tasks marked [x]
├── review.md           # Verdict document
└── progress/           # All wave logs
    ├── plan.md
    ├── wave-0.md
    ├── wave-1.md
    └── ...
```

### 2. Update Feature Registry

In `features.json`, set:
```json
{
  "name": "<feature-name>",
  "status": "done",
  "completed": "<date>"
}
```

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
