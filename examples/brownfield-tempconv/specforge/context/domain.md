# Domain knowledge

## Glossary
- **Kelvin** — Absolute thermodynamic temperature scale (0 K = absolute zero). _(aka: K)_

## Entities
- **E1 Temperature** — A scalar in C, F or K.
  - _invariant:_ Never below absolute zero (-273.15 C)

## Business rules
- **D1** No temperature below absolute zero (-273.15 C) is ever produced or accepted. _(entities: E1)_ _(audits: design,build)_
