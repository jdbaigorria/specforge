# Design — slugify

Status: approved

## Approach

A single pure function in `src/texttools/slug.py`, re-exported from
`__init__.py`. Implemented with the standard library only (constitution: zero
deps), using `unicodedata` for accent folding and `re` for tokenizing.

## Algorithm

```
def slugify(text: str, sep: str = "-") -> str:
    1. NFKD-normalize, drop combining marks  → folds accents (R2)
    2. lowercase
    3. find runs of [a-z0-9]+ with a regex    → drops punctuation (R3)
    4. join the runs with `sep`               → single separators (R1, R3)
    5. result is "" when there are no runs    → empty result (R4)
```

Accent folding uses `unicodedata.normalize("NFKD", text)` then filters out
characters where `unicodedata.combining(c)` is truthy. This decomposes `é` into
`e` + combining-acute and drops the accent.

## Public API

```python
def slugify(text: str, sep: str = "-") -> str: ...
```

`sep` is exposed for callers who want `_` instead of `-`; it defaults to `-` and
is not required by any requirement (kept minimal per "no ceremony without
purpose").

<!-- Security / Performance sections omitted: pure in-memory string transform,
     no untrusted-surface or scale concerns to document. -->
