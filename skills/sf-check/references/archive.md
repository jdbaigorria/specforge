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

The verdict gate should already be sealed (`sf gate approve --feature=<name>
--phase=verdict`, run at the APPROVE gate in `SKILL.md` Step 5). If it wasn't,
seal it now — don't hand-write the entry; the CLI seals a content hash of
`review.json` that the stale model relies on.

Finish the feature through the CLI — it is the only writer of `features.json`:

```bash
sf feature archive --feature=<name>
```

This copies the feature folder to `specforge/archive/<date>-<name>/` and sets
`status` to `done` in one atomic step. It **refuses unless the verdict gate is
sealed**, so it cannot archive a feature whose code doesn't pass — that is the
whole point. Record the human-readable completion date in `history.md` (yours to
write), not in features.json.

A `revise` verdict means the feature is NOT archived — don't run `sf feature
archive`; loop back to `sf-build`.

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
