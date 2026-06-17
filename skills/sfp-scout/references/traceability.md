# Upward traceability — why scout belongs in SpecForge

Discovery only earns a place in SpecForge if it connects to the traceability
spine. Otherwise it's a product-management appendage bolted on the front. The
connection is **stable product-requirement IDs**.

## The extended matrix

SpecForge already traces requirement → task → code → test (and checks drift on
it, F33). Scout extends that *upward*:

```
PR-requirement → roadmap feature → feature requirement → task → code → test
```

- The discovery brief states **product requirements** with stable IDs: `PR1`,
  `PR2`, … Each is a capability the product must have, at product altitude (not
  implementation detail).
- `sf-init --from <brief>` seeds the constitution and carries the `PR#` set.
- `sf-propose --all` derives the roadmap from the `PR#` ids: each roadmap feature
  declares which `PR#` it serves (a `serves: [PR1, PR3]` link). Feature-level
  requirements (R1, R2…) then trace to those.

## What this buys

- **No telephone game.** The feature requirements aren't a re-invention of the
  PRD — they're a *derivation* of named product requirements, with the link
  recorded. Drift between product intent and shipped features becomes detectable,
  not assumed.
- **Idea-to-test line.** You can answer "which test covers product requirement
  PR2?" by walking PR2 → feature → R# → trace.json → test.
- **Kill is cheap.** If a `PR#` is dropped at discovery, nothing downstream was
  built on it yet.

## Minimal mechanics

- Brief: `## Product Requirements` section, `PR1 … PRn`, each one line + rationale.
- `features.json` feature object gains `serves: ["PR1", ...]` (optional; empty for
  brownfield features with no discovery origin).
- `/sf-status` can later roll up coverage per `PR#` (which product requirements
  have shipped features) — a natural extension, not required for v1.

Without this section, scout is just `sfx-think` + `sfx-grill-me` pointed at a
product idea. With it, scout is the top of the traceability spine.
