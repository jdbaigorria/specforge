#!/usr/bin/env python3
"""SpecForge enforcement engine — portable core + per-harness adapters.

This is the v0.4 "hard gate" layer (F29). The DECISION LOGIC is harness-neutral
(`decide_pre_tool_use`, `session_context`, `mark_session`). Each harness gets a
thin adapter that translates its hook event in/out of that core. Claude Code is
the first adapter; a new harness = a new translate pair, not new logic.

Portable contract (`--harness generic`): read one JSON object on stdin and emit
one JSON object on stdout.

  in:  {"event": "pre_tool_use"|"session_start"|"session_end"|"pre_compact",
        "project_dir": "/abs/path", "tool": "Write", "file_path": "/abs/or/rel"}
  out: {"decision": "deny"|"allow", "reason": "...", "context": "..."}

Design choices:
- **Fail open.** Any internal error → allow (never brick the user's editing).
  Enforcement is a safety rail, not a tripwire.
- **No-op outside SpecForge.** If `<project>/specforge/` doesn't exist, allow
  everything — the plugin can be installed globally without interfering.
- Stdlib only (no deps), so it runs wherever python3 does.

Self-test: `python3 hooks/specforge_enforce.py --selftest`
"""

from __future__ import annotations

import json
import os
import re
import sys
from datetime import datetime, timezone
from pathlib import Path

APPROVED = {"approve", "approve-with-notes"}
DOWNSTREAM = re.compile(r"specforge/features/([^/]+)/(design|tasks)\.md$")
UPSTREAM_OF = {"design": "requirements", "tasks": "design"}


# ── Portable core ────────────────────────────────────────────────────────────

def _specforge_root(project_dir: str) -> Path | None:
    sf = Path(project_dir).resolve() / "specforge"
    return sf if sf.is_dir() else None


def gate_approved(sf_dir: Path, feature: str, phase: str) -> bool:
    fj = sf_dir / "features.json"
    if not fj.exists():
        return False
    try:
        data = json.loads(fj.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError):
        return False
    for f in data.get("features", []):
        if f.get("name") == feature:
            return any(
                g.get("phase") == phase and g.get("result") in APPROVED
                for g in f.get("gates", [])
            )
    return False


def decide_pre_tool_use(project_dir: str, file_path: str | None) -> tuple[str, str | None]:
    """Return ("deny", reason) or ("allow", None) for a Write/Edit on file_path."""
    if not file_path:
        return ("allow", None)
    sf = _specforge_root(project_dir)
    if sf is None:
        return ("allow", None)  # not a SpecForge project — never interfere
    proj = Path(project_dir).resolve()
    target = Path(file_path)
    if not target.is_absolute():
        target = proj / target
    try:
        rel = target.resolve().relative_to(proj).as_posix()
    except ValueError:
        return ("allow", None)  # outside the project tree

    if rel.startswith("specforge/.state/"):
        return (
            "deny",
            "specforge/.state/ is machine state — never edit it directly. "
            "features.json is the source of truth; let the session protocol manage state.",
        )

    m = DOWNSTREAM.match(rel)
    if m:
        feature, artefact = m.group(1), m.group(2)
        required = UPSTREAM_OF[artefact]
        if not gate_approved(sf, feature, required):
            return (
                "deny",
                f"Cannot write {artefact}.md for '{feature}': the {required} gate is "
                f"not approved in features.json. Present {required} at a 🔴 gate and get "
                f"approval first — gates cannot be skipped.",
            )
    return ("allow", None)


def session_context(project_dir: str) -> str:
    """Context to re-inject on SessionStart (incl. after compaction)."""
    sf = _specforge_root(project_dir)
    if sf is None:
        return ""
    chunks: list[str] = []
    sess = sf / ".state" / "session.md"
    if sess.exists():
        chunks.append("## SpecForge session — resume from here\n\n" + sess.read_text(encoding="utf-8"))
    rules = sf / "context" / "compact-rules.md"
    if rules.exists():
        chunks.append("## SpecForge compact-rules — project invariants\n\n" + rules.read_text(encoding="utf-8"))
    return "\n\n".join(chunks)


def mark_session(project_dir: str, note: str) -> None:
    """Append a lightweight continuity marker to session.md (side effect).

    The rich session summary is the agent's job (AGENT.md session protocol);
    a command hook has no conversation access. This only timestamps lifecycle
    events so continuity is visible across end/compaction.
    """
    sf = _specforge_root(project_dir)
    if sf is None:
        return
    state = sf / ".state"
    state.mkdir(parents=True, exist_ok=True)
    ts = datetime.now(timezone.utc).isoformat(timespec="seconds")
    with (state / "session.md").open("a", encoding="utf-8") as fh:
        fh.write(f"\n<!-- {note} @ {ts} -->\n")


# ── Adapters ─────────────────────────────────────────────────────────────────

def _project_dir(payload: dict) -> str:
    return os.environ.get("CLAUDE_PROJECT_DIR") or payload.get("cwd") or payload.get("project_dir") or os.getcwd()


def run_claude_code(event: str, payload: dict) -> int:
    if event == "PreToolUse":
        file_path = (payload.get("tool_input") or {}).get("file_path")
        decision, reason = decide_pre_tool_use(_project_dir(payload), file_path)
        if decision == "deny":
            print(json.dumps({"hookSpecificOutput": {
                "hookEventName": "PreToolUse",
                "permissionDecision": "deny",
                "permissionDecisionReason": reason,
            }}))
        return 0
    if event == "SessionStart":
        ctx = session_context(_project_dir(payload))
        if ctx:
            print(json.dumps({"hookSpecificOutput": {
                "hookEventName": "SessionStart",
                "additionalContext": ctx,
            }}))
        return 0
    if event == "SessionEnd":
        mark_session(_project_dir(payload), "session ended")
        return 0
    if event == "PreCompact":
        mark_session(_project_dir(payload), "pre-compaction checkpoint")
        return 0
    return 0


def run_generic(payload: dict) -> int:
    event = payload.get("event")
    pd = _project_dir(payload)
    if event == "pre_tool_use":
        decision, reason = decide_pre_tool_use(pd, payload.get("file_path"))
        print(json.dumps({"decision": decision, "reason": reason}))
    elif event == "session_start":
        print(json.dumps({"decision": "allow", "context": session_context(pd)}))
    elif event == "session_end":
        mark_session(pd, "session ended")
        print(json.dumps({"decision": "allow"}))
    elif event == "pre_compact":
        mark_session(pd, "pre-compaction checkpoint")
        print(json.dumps({"decision": "allow"}))
    else:
        print(json.dumps({"decision": "allow", "reason": f"unknown event {event!r}"}))
    return 0


# ── Entry point ──────────────────────────────────────────────────────────────

def main(argv: list[str]) -> int:
    if "--selftest" in argv:
        return selftest()
    harness = _flag(argv, "--harness", "claude-code")
    event = _flag(argv, "--event", "")
    try:
        raw = sys.stdin.read()
        payload = json.loads(raw) if raw.strip() else {}
    except json.JSONDecodeError:
        payload = {}
    try:
        if harness == "generic":
            return run_generic(payload)
        return run_claude_code(event or payload.get("hook_event_name", ""), payload)
    except Exception as exc:  # fail open — never block on an internal error
        print(f"specforge-enforce: internal error, allowing ({exc})", file=sys.stderr)
        return 0


def _flag(argv: list[str], name: str, default: str) -> str:
    if name in argv:
        i = argv.index(name)
        if i + 1 < len(argv):
            return argv[i + 1]
    return default


def selftest() -> int:
    import tempfile

    failures = 0

    def check(cond: bool, label: str) -> None:
        nonlocal failures
        print(("PASS" if cond else "FAIL") + f"  {label}")
        failures += 0 if cond else 1

    with tempfile.TemporaryDirectory() as d:
        proj = Path(d)
        feats = proj / "specforge" / "features" / "x"
        feats.mkdir(parents=True)
        reg = proj / "specforge" / "features.json"

        # no gates yet → design.md denied, requirements.md allowed
        reg.write_text(json.dumps({"schema_version": "1.0", "features": [
            {"name": "x", "status": "approved", "gates": []}
        ]}))
        check(decide_pre_tool_use(str(proj), str(feats / "design.md"))[0] == "deny",
              "design.md denied without requirements gate")
        check(decide_pre_tool_use(str(proj), str(feats / "requirements.md"))[0] == "allow",
              "requirements.md allowed (no upstream)")

        # requirements approved → design.md allowed, tasks.md still denied
        reg.write_text(json.dumps({"schema_version": "1.0", "features": [
            {"name": "x", "gates": [{"phase": "requirements", "result": "approve"}]}
        ]}))
        check(decide_pre_tool_use(str(proj), str(feats / "design.md"))[0] == "allow",
              "design.md allowed after requirements gate")
        check(decide_pre_tool_use(str(proj), str(feats / "tasks.md"))[0] == "deny",
              "tasks.md denied without design gate")

        # .state/ is always denied
        check(decide_pre_tool_use(str(proj), str(proj / "specforge" / ".state" / "session.md"))[0] == "deny",
              ".state/ write denied")

        # outside specforge/ → allowed
        check(decide_pre_tool_use(str(proj), str(proj / "src" / "main.py"))[0] == "allow",
              "non-specforge path allowed")

        # no specforge/ project → allowed (no interference)
        with tempfile.TemporaryDirectory() as d2:
            check(decide_pre_tool_use(d2, str(Path(d2) / "specforge" / "features" / "y" / "design.md"))[0] == "allow",
                  "non-SpecForge project: no-op allow")

    print(f"\n{'OK' if not failures else 'FAILED'}: {failures} failure(s).")
    return 1 if failures else 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
