# SpecForge examples

A worked, end-to-end feature so you can see what the pipeline actually produces —
not just how it's described. It doubles as a fixture for `scripts/lint-skills.py`
and the future CLI.

## `slugify/`

A trivial feature (`slugify` — turn a string into a URL slug) for a small Python
utility library, taken all the way through the pipeline and **archived**:

```
requirements → design → tasks → build → check → archive
```

What to look at:

| File | What it demonstrates |
|------|----------------------|
| `slugify/specforge/features.json` | The registry + **gate ledger** (F10): every gate that actually passed, with timestamps. Source of truth for status. |
| `slugify/specforge/archive/2026-06-16-slugify/requirements.md` | EARS notation, acceptance criteria. |
| `.../design.md` | Conditional sections only — no `N/A` noise. |
| `.../tasks.md` | Waves + the **traceability matrix** (requirement → task). |
| `.../review.md` | The full matrix (requirement → task → code → test → status) and the verdict. |
| `slugify/specforge/context/` | Project context (`project.md`, `conventions.md`). |

The feature is in `archive/` because it completed. A feature still in flight
would live under `specforge/features/<name>/` with the same file set.

> This is illustrative content, not a runnable package — the point is the shape
> of the artefacts and the traceability chain.
