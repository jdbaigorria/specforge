# fix-empty-wordcount — Lite change

> SpecForge, lite lane (`sf-propose` Step 0 classified this trivial + low-risk).
> Lane: lite · one combined artefact, one spec gate, build, minimal check.

## What & why

`word_count("")` returned `1` because the implementation used `text.split(" ")`,
which yields `['']` on empty input. Switch to `text.split()` (no argument), which
collapses whitespace and returns `[]` for empty/whitespace-only strings.
One file, no sensitive surface, trivially reversible → lite.

## Requirement

**R1** — WHEN `word_count` receives an empty or whitespace-only string, THE SYSTEM SHALL return `0`.
*Acceptance:* `word_count("") == 0` and `word_count("   ") == 0`.

## Plan

- [x] `src/textstats/count.py` — use `text.split()` instead of `text.split(" ")`
- [x] `tests/test_count.py` — add `test_empty`, `test_whitespace_only`
