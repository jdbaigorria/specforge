# SpecForge adapter for pi (pi.dev)

`specforge.js` is a pi coding-agent **extension** that wires the SpecForge
enforcement engine into pi. It is a thin translate pair over `sf hook
--harness=generic` (see [`../README.md`](../README.md)) — no enforcement logic
lives here.

## What it does

| pi event | SpecForge event | Effect |
|----------|-----------------|--------|
| `tool_call` (`write`/`edit`) | `pre_tool_use` | hard-deny a gate-skipping write (returns `{block, reason}`) |
| `session_start` | `session_start` | fetches resume/compact-rules/learnings context (injected next turn — pi's `session_start` can't inject) |
| `before_agent_start` | `user_prompt_submit` | injects the current-step slice (breadcrumb or full) + the deferred session context + journal nudge, as a `message` |
| `session_before_compact` | `pre_compact` | flags a full re-ground on the next turn |
| `session_shutdown` | `session_end` | appends a continuity marker |

Of the non-Claude harnesses, pi has the closest match to SpecForge's needs: it
covers all six events natively.

## Requirements

- The **`sf` binary on PATH** (or set `SPECFORGE_SF_BIN` to an absolute path).
  `sf` is the decision engine; the extension only translates I/O.
- Fail-open: any adapter error → allow / no injection, never a block.

## Install

Register the extension in pi's settings — global `~/.pi/agent/settings.json` or
per-project `.pi/settings.json`:

```json
{
  "extensions": ["/abs/path/to/specforge/hooks/pi/specforge.js"]
}
```

To also load the SpecForge skills as `/skill:...` commands, add the `skills` key:

```json
{
  "extensions": ["/abs/path/to/specforge/hooks/pi/specforge.js"],
  "skills": ["/abs/path/to/specforge/skills"]
}
```

Project-local resources require trust on first load (pi's `defaultProjectTrust`).

## Status

The `sf hook` boundary is covered by the Go test suite and was validated against
this adapter's exact subprocess call. The pi-side event wiring (`pi.on(...)`)
should be verified in a real pi session — pi's extension API is the moving part,
not the SpecForge contract.
