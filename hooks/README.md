# SpecForge enforcement hooks (F29)

The optional **hard-gate** layer. SpecForge's core is portable Markdown +
cooperative gates; this layer turns the gates from "the agent should" into "the
harness won't let it." It is opt-in and **per-harness** — the portable core
never depends on it.

## What it enforces

All decisions are made by the `sf` CLI's **`sf hook`** command, which is
**harness-agnostic** — the same decision engine the rest of the CLI uses, so the
gate ledger has a single source of truth (no second implementation to drift).

| Event | Behaviour |
|-------|-----------|
| `PreToolUse` (Write/Edit) | **Hard-deny** writing a downstream artifact whose upstream gate isn't approved — `design`←`requirements`, `tasks`←`design`, `plan`←`tasks` — read from the `features.json` gate ledger (F10). Gates the `.json` source **and** the `.md` render (JSON-first). Gates cannot be skipped. |
| `PreToolUse` (Write/Edit) | **Serial flow (F22).** Hard-deny creating a new feature's `requirements` while another feature is `approved`/`building`. One active feature at a time, so `sf state current` stays unambiguous. |
| `PreToolUse` (Write/Edit) | **Hard-deny** direct writes to `specforge/.state/` (machine state). Protects the source-of-truth boundary (F2/F25). |
| `SessionStart` | Inject `specforge/.state/session.md` + `specforge/context/compact-rules.md` + `specforge/learnings.md` as context — on startup, resume, **and after compaction**. Restores project state and consolidated learnings without the agent having to remember (F21/F7/F31). |
| `UserPromptSubmit` | Inject the current step's slice (computed in-process, the same logic as `sf context current`) so the spec doesn't dilute as context fills. Cheap **breadcrumb** every turn; full slice on **step-change** or every N turns (salience backstop, N = `SPECFORGE_FULL_SLICE_EVERY`, default 10). Per-session trigger state lives in `specforge/.state/hook-context.json`. |
| `Stop` | **Memory reconciler.** If a feature is archived (`done`/`archived`) with no journal entry yet, nudge **once** to extract durable lessons via `sf journal add`. Blocks a single time per feature (recorded in `.state/hook-context.json`), then stands down — nudge-once, never a loop. Honours `stop_hook_active`. |
| `PreCompact` | Append a continuity marker **and** flag the session to re-inject the **full** slice on the next `UserPromptSubmit` — after compaction the earlier slices are summarized away, so re-grounding is forced (salience). |
| `SessionEnd` | Append a timestamped continuity marker to `session.md`. (The rich summary stays the agent's job — a command hook has no conversation access.) |

Two safety properties, by design:

- **Fail open.** Any internal error → allow. Enforcement is a rail, not a tripwire.
- **No-op outside SpecForge.** If `<project>/specforge/` doesn't exist, every
  event is allowed — the plugin is safe to install globally.

## Layout (portable-first)

```
hooks/
├── claude-code/
│   └── hooks.json           # Claude Code adapter (wired via plugin.json "hooks")
├── pi/
│   ├── specforge.js         # pi (pi.dev) adapter — extension over `sf hook`
│   └── README.md            # pi install + event mapping
└── README.md                # this file
```

The decision engine lives in the Go CLI (`cli/hook.go`): pure decision functions
(`decidePreToolUse`, `decideInjection`, `pickJournalNudge`) reusing the same
gate-ledger logic as the rest of `sf`, plus a thin per-harness I/O translation
(`runHookClaude`, `runHookGeneric`). A new harness is a new translate pair, not
new logic.

**Requires the `sf` binary on PATH** (or via an absolute path in the adapter).
`sf` is already the CLI that persists/validates/renders, so the hook layer adds
no new runtime dependency — and it drops the previous `python3` requirement.

## The portable contract (`--harness generic`)

For any harness that isn't Claude Code, drive the engine through one normalized
JSON object on stdin → one JSON object on stdout:

```
in:  {"event": "pre_tool_use", "project_dir": "/abs", "tool": "Write", "file_path": "/abs/or/rel"}
out: {"decision": "deny", "reason": "..."}

in:  {"event": "session_start", "project_dir": "/abs"}
out: {"decision": "allow", "context": "...load into the model..."}

in:  {"event": "user_prompt_submit", "project_dir": "/abs", "session_id": "..."}
out: {"decision": "allow", "context": "...current-step slice (or "")..."}

in:  {"event": "stop", "project_dir": "/abs"}
out: {"decision": "block", "reason": "...journal nudge..."}  // or {"decision": "allow"}
```

`event` ∈ `pre_tool_use | session_start | user_prompt_submit | stop | session_end | pre_compact`.

To add a harness, write an adapter that (1) translates that harness's hook event
into this object, (2) calls `sf hook --harness=generic`, and (3) maps
`{"decision": "deny", "reason"}` back into the harness's block mechanism.

## Claude Code adapter

Shipped and active when the plugin is installed — `plugin.json` points its
`hooks` field at `hooks/claude-code/hooks.json`, which runs `sf hook
--harness=claude-code` and maps the result to Claude Code's
`permissionDecision: "deny"` / `additionalContext` contract.

## pi (pi.dev) adapter

`hooks/pi/specforge.js` is a pi extension that maps pi's events
(`tool_call`, `before_agent_start`, `session_start`, `session_before_compact`,
`session_shutdown`) onto `sf hook --harness=generic`. pi covers all six events
natively. See [`pi/README.md`](pi/README.md) for the event table and install.

## Test

The decision engine is covered by the Go test suite (parity tests ported from the
former Python `--selftest`):

```sh
cd cli && go test ./...
```
