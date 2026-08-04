# Design — add-kelvin

Two pure functions plus a shared absolute-zero guard, added to the existing `convert` module.

## Components

### C1 — kelvin conversions (functions)
- c_to_k
- k_to_c
depends on: C2

### C2 — absolute-zero guard (helper)
- reject temperatures below -273.15 C

## Decisions

### D1 — Where does the absolute-zero check live?
- **Inside each conversion** (pros: simple) (cons: duplicated)
- **Shared private helper** (pros: single source) (cons: one extra function)
**Chosen:** Shared private helper — Keeps the rule in one place (P1: pure helper, no state).
