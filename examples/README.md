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

It also ships the **real code** the spec describes (`src/texttools/slug.py` +
`tests/test_slug.py`), so the traceability chain is complete and verifiable:

```sh
# project health view — per-feature status, drift, blockers, critical path (F37)
python3 ../../scripts/sf-status.py .

# drift check — every requirement's code anchor still exists (F33)
python3 ../../scripts/check-drift.py .

# run the spec's acceptance tests
PYTHONPATH=src python3 -m doctest src/texttools/slug.py -v
```

`archive/2026-06-16-slugify/trace.json` is the structured matrix that links each
requirement (R1–R4) to its `path:symbol` and test — the live link drift detection
reads.
