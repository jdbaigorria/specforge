# SpecForge enforcement hooks (F29)

The optional **hard-gate** layer. SpecForge's core is portable Markdown +
cooperative gates; this layer turns the gates from "the agent should" into "the
harness won't let it." It is opt-in and **per-harness** — the portable core
never depends on it.

## What it enforces

All decisions are made by `specforge_enforce.py`, which is **harness-agnostic**.

| Event | Behaviour |
|-------|-----------|
| `PreToolUse` (Write/Edit) | **Hard-deny** writing `design.md` without the `requirements` gate approved, or `tasks.md` without the `design` gate — read from the `features.json` gate ledger (F10). Gates cannot be skipped. |
| `PreToolUse` (Write/Edit) | **Hard-deny** direct writes to `specforge/.state/` (machine state). Protects the source-of-truth boundary (F2/F25). |
| `SessionStart` | Inject `specforge/.state/session.md` + `specforge/context/compact-rules.md` + `specforge/learnings.md` as context — on startup, resume, **and after compaction**. Restores project state and consolidated learnings without the agent having to remember (F21/F7/F31). |
| `UserPromptSubmit` | Inject the current step's slice via `sf context current` so the spec doesn't dilute as context fills. Cheap **breadcrumb** every turn; full slice on **step-change** or every `FULL_SLICE_EVERY` turns (salience backstop). Per-session trigger state lives in `specforge/.state/hook-context.json`. Requires the `sf` binary — absent ⇒ injects nothing (fail open). |
| `Stop` | **Memory reconciler.** If a feature is archived (`done`/`archived`) with no journal entry yet, nudge **once** to extract durable lessons via `sf journal add`. Blocks a single time per feature (recorded in `.state/hook-context.json`), then stands down — nudge-once, never a loop. Honours `stop_hook_active`. |
| `SessionEnd` / `PreCompact` | Append a timestamped continuity marker to `session.md`. (The rich summary stays the agent's job — a command hook has no conversation access.) |

Two safety properties, by design:

- **Fail open.** Any internal error → allow. Enforcement is a rail, not a tripwire.
- **No-op outside SpecForge.** If `<project>/specforge/` doesn't exist, every
  event is allowed — the plugin is safe to install globally.

## Layout (portable-first)

```
hooks/
├── specforge_enforce.py     # PORTABLE CORE — decision logic, no harness knowledge
├── claude-code/
│   └── hooks.json           # Claude Code adapter (wired via plugin.json "hooks")
└── README.md                # this file
```

`specforge_enforce.py` separates the decision functions (`decide_pre_tool_use`,
`session_context`, `mark_session`) from the thin per-harness I/O translation. A
new harness is a new translate pair, not new logic.

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
into this object, (2) calls `python3 specforge_enforce.py --harness generic`, and
(3) maps `{"decision": "deny", "reason"}` back into the harness's block mechanism.

## Claude Code adapter

Shipped and active when the plugin is installed — `plugin.json` points its
`hooks` field at `hooks/claude-code/hooks.json`, which runs the engine with
`--harness claude-code` and maps the result to Claude Code's
`permissionDecision: "deny"` / `additionalContext` contract.

Requires `python3` on PATH (already true for any machine running this lint/CI).

## Test

```sh
python3 hooks/specforge_enforce.py --selftest
```
