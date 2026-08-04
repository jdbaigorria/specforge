[← Back to README](../README.md)

# Concepts

## Spec format: EARS notation

SpecForge uses EARS (Easy Approach to Requirements Syntax) for unambiguous requirements:

```
WHEN [trigger] THE SYSTEM SHALL [behavior]          ← event-driven
WHILE [state] THE SYSTEM SHALL [behavior]           ← state-driven
IF [unwanted condition] THEN THE SYSTEM SHALL [fix]  ← error handling
THE SYSTEM SHALL [behavior]                          ← always active
```

One requirement per statement. Active voice. Specific and testable.
Full reference in `sf-propose/references/ears-notation.md`.

---

## Feature lifecycle

```json
// specforge/features/user-auth/feature.json
{
      "name": "user-auth",
      "status": "done",
      "workflow": "requirements-first",
      "created": "2026-05-10",
      "completed": "2026-05-12"
    }
  ]
}
```

Statuses: `pending` → `proposing` → `approved` → `building` → `checking` → `done`

Each transition is triggered by a skill completing its work:
- sf-propose sets `approved` after all 3 artefacts are gate-approved
- sf-build sets `checking` after all waves complete
- sf-check sets `done` after APPROVE + archive

---

## Backprop: how the spec learns

When the same type of issue appears across 3+ features, it becomes a project invariant.

```
Feature A: API endpoint missing error handling     ← occurrence 1
Feature B: CLI command crashes on bad input        ← occurrence 2
Feature C: Webhook handler ignores timeout         ← occurrence 3 → PROMOTE

→ New invariant in constitution.md:
  "Every external interface must have explicit error handling."
```

Tracking lives in `specforge/history.md`:

```markdown
## Recurring Issues
| Issue Pattern | Occurrences | Features | Status |
|---------------|-------------|----------|--------|
| Missing error handling | 3 | auth, cli, api | → INVARIANT |
```

After promotion, sf-check validates every future feature against the new invariant.

---

## Resync detection

When sf-propose is invoked on a feature that already has artefacts:

1. Check timestamps: if `requirements.md` is newer than `design.md` → stale
2. Prompt: "requirements.md was modified after design.md was generated. Regenerate design? [y/n]"
3. If yes → regenerate from the updated upstream artefact
4. Cascade: if design changes → offer to regenerate tasks too

This handles the common case where the user edits a spec file directly and
downstream artefacts need to catch up.
