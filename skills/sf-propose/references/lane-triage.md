# Lane Triage (lite vs standard)

Before specifying, classify the change and **propose a lane**. The framework
classifies; the human confirms (F34). This is the only place a fast lane is
chosen — never let momentum silently skip ceremony.

Estimate from the user's description:

- **Surface:** touches one file/string vs many?
- **Sensitivity:** any security/auth/data/money/migration surface? (if yes → standard)
- **Reversibility:** trivially revertible vs not?

Propose a lane:

- **Lite** — typos, copy, config, a one-file change with no sensitive surface.
  One combined artefact, one spec gate, build, minimal check.
- **Standard** — everything else. The full requirements → design → tasks flow.

```
───────────────────────────────────────
🔴 GATE — lane: proposed **{lite|standard}** for "<feature>"
Reason: <1 file / no sensitive surface / reversible — or why standard>
Awaiting approval. Reply: approve / use {other lane} / change X
───────────────────────────────────────
```

Record the approved lane in `features.json` (`"lane": "lite"|"standard"`). It is
a gated, auditable decision — not the agent deciding to go fast.

- **Lite approved** → read `references/lite-lane.md` and follow it. Stop here.
- **Standard approved** → continue with the Requirements-First Flow in `SKILL.md`.

When the lane is obvious and the user invoked plainly, you may still propose —
but always through this gate. Default to **standard** whenever unsure.
