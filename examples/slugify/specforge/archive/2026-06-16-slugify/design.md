# Design — slugify

A single pure function in `src/texttools/slug.py`, re-exported from `__init__.py`. Standard library only (`unicodedata` for accent folding, `re` for tokenizing).

**Algorithm:** NFKD-normalize and drop combining marks (folds accents, R2); lowercase; find runs of `[a-z0-9]+` with a regex (drops punctuation, R3); join the runs with `sep` (single separators, R1/R3); the result is `""` when there are no runs (R4).

**Public API:** `def slugify(text: str, sep: str = "-") -> str`.

## Components

### C1 — slugify function (function)
- normalize + fold accents
- tokenize alnum runs
- join with separator
- return empty on no runs

## Decisions

### D1 — Should the separator be configurable?
- **Hard-code "-"** (pros: simplest) (cons: callers wanting "_" must post-process)
- **Optional sep param defaulting to "-"** (pros: flexible,no required arg) (cons: one extra parameter)
**Chosen:** Optional sep param defaulting to "-" — Not required by any requirement, defaults to `-`; kept minimal per "no ceremony without purpose".
