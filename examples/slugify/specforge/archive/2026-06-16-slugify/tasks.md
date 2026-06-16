# Tasks — slugify

Status: approved

## Wave 0 (all tasks independent of each other except T1 → rest)

- [x] **T1** — Create `src/texttools/slug.py` with `slugify` signature + NFKD
  accent folding and lowercase. *(R1, R2)*
- [x] **T2** — Add regex tokenizing + `sep` join; collapse separators, no
  leading/trailing hyphen. *(R1, R3)*
- [x] **T3** — Handle the empty/no-slug-able case to return `""`. *(R4)*
- [x] **T4** — Re-export `slugify` from `src/texttools/__init__.py`; add
  `tests/test_slug.py` covering R1–R4. *(R1, R2, R3, R4)*

## Traceability (requirement → task)

| Requirement | Task(s) |
|-------------|---------|
| R1 — Basic slug | T1, T2, T4 |
| R2 — Accent folding | T1, T4 |
| R3 — Punctuation/collapsing | T2, T4 |
| R4 — Empty result | T3, T4 |
