# Wave 0 — slugify

Integrity: **CLEAN** (all operations landed, all tests green)

## Operations

| Task | Operation | File | Result |
|------|-----------|------|--------|
| T1 | create | `src/texttools/slug.py` | ✅ |
| T2 | edit | `src/texttools/slug.py` | ✅ |
| T3 | edit | `src/texttools/slug.py` | ✅ |
| T4 | create | `src/texttools/__init__.py` | ✅ |
| T4 | create | `tests/test_slug.py` | ✅ |

## Tests

```
tests/test_slug.py::test_basic                PASSED
tests/test_slug.py::test_accents              PASSED
tests/test_slug.py::test_punctuation_collapse PASSED
tests/test_slug.py::test_empty                PASSED
4 passed
```

All four tasks complete and marked `[x]` in tasks.md. Wave integrity CLEAN →
gate passed → proceed to sf-check.
