# SpecForge

**Spec-driven development that is actually enforced.** Most SDD tooling
prepares or proposes; SpecForge **governs**: a deterministic Go CLI (`sf`),
per-harness hooks, and a skill suite that together make illegal states
unreachable — gates can't be skipped, state can't be forged, and "done" requires
a fresh green test run sealed by the machine.

- **New here?** Read the [mental model](mental-model.md) (one page), then run
  the [quickstart](quickstart.md) — first governed feature in ~10 minutes.
- **Evaluating guarantees?** [What each harness actually guarantees](harness-guarantees.md)
  is the honest degradation table.
- **Reference:** [CLI & hooks](cli-and-hooks.md) ·
  [architecture](architecture.md) · [skills](skills.md) ·
  [concepts](concepts.md) · [walkthrough](walkthrough.md).

> The division of labor: the LLM *produces and judges*; `sf` *persists,
> validates, computes, and renders*; the hooks *force and inject*. None of the
> three does another's job.
