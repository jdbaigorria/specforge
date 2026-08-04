# Tasks — add-kelvin

- **T1** — Add absolute-zero guard helper `[done]`
  - requirements: R3
  - components: C2
  - files: src/tempconv/convert.py
  - depends on: —
  - effort: S
- **T2** — Implement c_to_k and k_to_c `[done]`
  - requirements: R1,R2
  - components: C1
  - files: src/tempconv/convert.py,tests/test_convert.py
  - depends on: T1
  - effort: S
