# Conventions — texttools

## Style

- Format with `black`, lint with `ruff`.
- Full type hints on every public function.
- Docstrings on every public function, with at least one example.

## Naming

- Functions: `snake_case`, verbs (`slugify`, `truncate`).
- Private helpers: leading underscore (`_strip_accents`).

## Tests

- One test module per source module: `slug.py` → `test_slug.py`.
- Test through the public API, not private helpers.
- Each acceptance criterion from the spec maps to at least one test.
