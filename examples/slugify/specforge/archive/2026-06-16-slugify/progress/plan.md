# Execution plan — slugify

Status: approved

## Waves

**Wave 0** — single wave; the whole feature is one small module.

| Task | Depends on | Files |
|------|-----------|-------|
| T1 | — | `src/texttools/slug.py` (new) |
| T2 | T1 | `src/texttools/slug.py` |
| T3 | T1 | `src/texttools/slug.py` |
| T4 | T1–T3 | `src/texttools/__init__.py` (new), `tests/test_slug.py` (new) |

One wave is enough: the tasks all touch one module and a test file, with a simple
linear dependency (T1 first, then T2/T3, then T4). No parallel waves needed.

## Test strategy

TDD per task: write the failing test for the requirement, then implement. Final
`tests/test_slug.py` covers R1–R4.
