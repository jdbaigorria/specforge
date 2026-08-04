# Project — texttools

## Stack

- **Language:** Python 3.9+
- **Tests:** pytest
- **Packaging:** `pyproject.toml` (PEP 621), no third-party runtime deps
- **Layout:** `src/texttools/` package, `tests/` alongside

## Architecture

A flat package of pure functions, one module per concern:

```
src/texttools/
├── __init__.py        # re-exports the public API
└── slug.py            # slugify()
tests/
└── test_slug.py
```

No classes, no state. Each public function is imported directly:
`from texttools import slugify`.
