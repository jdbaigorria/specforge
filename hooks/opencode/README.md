# SpecForge adapter for opencode

`specforge.js` is an opencode **plugin** that wires the SpecForge enforcement
engine into opencode. Thin translate pair over `sf hook --harness=generic` (see
[`../README.md`](../README.md)) — no enforcement logic lives here.

## What it does (and doesn't)

| opencode hook | SpecForge event | Status |
|---------------|-----------------|--------|
| `tool.execute.before` (`write`/`edit`) | `pre_tool_use` | ✅ hard-deny a gate-skipping write — throws; opencode passes the message to the agent as the reason |
| `experimental.session.compacting` | `pre_compact` | ✅ pushes the current slice into `output.context` so the spec survives compaction |
| per-turn slice + SessionStart resume context | `user_prompt_submit` / `session_start` | ⚠️ **not wired** — see below |

### Why per-turn injection is a gap on opencode

SpecForge's other half is **context injection** — re-grounding the agent each turn
so the spec doesn't dilute as context fills (Claude Code and pi both do this).
opencode's only injection points are the **experimental** `chat.*` hooks
(`experimental.chat.messages.transform`, `experimental.chat.system.transform`),
whose `Message`/`Part` payload shapes are unstable. Rather than fabricate a
message shape we can't verify (a wrong shape would break the request), this
adapter leaves per-turn injection unwired until opencode's chat hooks stabilize.

**Net:** on opencode the *hard gate* holds (the most important guarantee), and
the spec survives compaction; the per-turn re-grounding degrades to none. This is
the honest trade-off — opencode has the weakest injection surface of the
supported harnesses. Drift detection (`sf doctor`) and the gate ledger still work
exactly the same, since those live in `sf`, not the harness.

## Requirements

- The **`sf` binary on PATH** (or set `SPECFORGE_SF_BIN`). Fail-open on any error.

## Install

opencode auto-discovers `*.js`/`*.ts` plugins in the plugin directory:

```sh
# project-local
mkdir -p .opencode/plugin
ln -s /abs/path/to/specforge/hooks/opencode/specforge.js .opencode/plugin/specforge.js
# or global: ~/.config/opencode/plugin/
```

Or reference it from `opencode.json` `plugin` array. The SpecForge skills install
the same as any other agent (see the repo `INSTALL.md`).

## Status

The `sf hook` boundary is covered by the Go test suite and was validated against
this adapter's exact subprocess call (gate deny/allow). The opencode-side hook
wiring should be verified in a real opencode session.
