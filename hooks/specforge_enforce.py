#!/usr/bin/env python3
"""SpecForge enforcement engine — portable core + per-harness adapters.

This is the v0.4 "hard gate" layer (F29). The DECISION LOGIC is harness-neutral
(`decide_pre_tool_use`, `session_context`, `mark_session`). Each harness gets a
thin adapter that translates its hook event in/out of that core. Claude Code is
the first adapter; a new harness = a new translate pair, not new logic.

Portable contract (`--harness generic`): read one JSON object on stdin and emit
one JSON object on stdout.

  in:  {"event": "pre_tool_use"|"session_start"|"session_end"|"pre_compact"
              |"user_prompt_submit",
        "project_dir": "/abs/path", "tool": "Write", "file_path": "/abs/or/rel",
        "session_id": "..."}
  out: {"decision": "deny"|"allow", "reason": "...", "context": "..."}

Design choices:
- **Fail open.** Any internal error → allow (never brick the user's editing).
  Enforcement is a safety rail, not a tripwire.
- **No-op outside SpecForge.** If `<project>/specforge/` doesn't exist, allow
  everything — the plugin can be installed globally without interfering.
- Stdlib only (no deps), so it runs wherever python3 does. The one runtime
  dependency is the `sf` binary, used ONLY by user_prompt_submit to fetch the
  current slice; if `sf` is absent the hook injects nothing (fail open).

Self-test: `python3 hooks/specforge_enforce.py --selftest`
"""

from __future__ import annotations

import json
import os
import re
import shutil
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

APPROVED = {"approve", "approve-with-notes"}
ACTIVE_STATUSES = {"approved", "building"}  # "en curso" (serial, F22)
UPSTREAM_OF = {"design": "requirements", "tasks": "design", "plan": "tasks"}

# Ruta-relativa → (feature, artefacto). Reconoce .json Y .md (JSON-first: la
# fuente es .json, el .md es render); plan vive bajo progress/.
ARTIFACT_RE = re.compile(
    r"specforge/features/([^/]+)/(?:progress/)?(requirements|design|tasks|plan)\.(?:json|md)$"
)

# Trigger del slice (decisión #1 del debate): el slice COMPLETO se re-inyecta
# on-step-change y, como backstop de saliencia, cada N turnos sin cambio. El
# resto de los turnos va solo el breadcrumb (barato). Sin gauge de % de contexto
# en el hook, N turnos es el proxy del umbral. Configurable por env (capa
# per-harness): SPECFORGE_FULL_SLICE_EVERY.
FULL_SLICE_EVERY = 10


def _full_slice_every() -> int:
    """N de turnos para el backstop, con override por env (>0); si no, el default."""
    try:
        v = int(os.environ.get("SPECFORGE_FULL_SLICE_EVERY", ""))
        return v if v > 0 else FULL_SLICE_EVERY
    except ValueError:
        return FULL_SLICE_EVERY

# Conciliador de memoria (paso 4): qué estados cuentan como "archivada" para
# disparar el nudge del journal.
ARCHIVED_STATUSES = {"done", "archived"}


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


def _artifact_target(rel: str):
    """(feature, artefacto) si rel es un artefacto gateable, o None."""
    m = ARTIFACT_RE.match(rel)
    return (m.group(1), m.group(2)) if m else None


def _active_features(sf_dir: Path) -> list:
    """Nombres de features 'en curso' (status approved/building) — para el serial."""
    try:
        data = json.loads((sf_dir / "features.json").read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return []
    return [f.get("name", "") for f in data.get("features", []) if f.get("status") in ACTIVE_STATUSES]


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

    target = _artifact_target(rel)
    if target is None:
        return ("allow", None)
    feature, artefact = target

    # Serial (F22 / #4): no arrancar una 2da feature mientras otra está en curso.
    # El PRIMER artefacto de una feature es requirements → lo interceptamos ahí.
    if artefact == "requirements":
        others = [n for n in _active_features(sf) if n and n != feature]
        if others:
            return (
                "deny",
                f"Serial flow: feature '{others[0]}' is still active. Finish and archive "
                f"it before starting '{feature}' — one active feature at a time (F22).",
            )
        return ("allow", None)

    # Cadena de gates: el artefacto downstream necesita su upstream aprobado.
    required = UPSTREAM_OF.get(artefact)
    if required and not gate_approved(sf, feature, required):
        return (
            "deny",
            f"Cannot write {artefact} for '{feature}': the {required} gate is not "
            f"approved in features.json. Present {required} at a 🔴 gate and get "
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
    learnings = sf / "learnings.md"
    if learnings.exists():
        chunks.append("## SpecForge learnings — consolidated, evidence-anchored\n\n" + learnings.read_text(encoding="utf-8"))
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


# ── Slice injection (UserPromptSubmit) ───────────────────────────────────────
#
# El hook NO computa el slice: llama a `sf context current` (el cerebro). Así no
# duplicamos la lógica de estado en Python (evita "segunda verdad", F2). Si `sf`
# no está instalado, no inyectamos nada (fail open).


def _sf_bin() -> str | None:
    """Ubica el binario sf: override por env, o en el PATH."""
    return os.environ.get("SPECFORGE_SF_BIN") or shutil.which("sf")


def _run_sf(args: list[str]) -> str | None:
    """Corre `sf <args>` y devuelve stdout, o None ante cualquier fallo."""
    sf = _sf_bin()
    if not sf:
        return None
    try:
        res = subprocess.run([sf, *args], capture_output=True, text=True, timeout=5)
    except (OSError, subprocess.SubprocessError):
        return None
    return res.stdout if res.returncode == 0 else None


def _hook_state(sf_dir: Path) -> dict:
    """Estado del hook por sesión, en specforge/.state/ (la máquina lo posee; el
    LLM tiene prohibido escribir ahí)."""
    p = sf_dir / ".state" / "hook-context.json"
    try:
        return json.loads(p.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return {}


def _save_hook_state(sf_dir: Path, data: dict) -> None:
    state = sf_dir / ".state"
    try:
        state.mkdir(parents=True, exist_ok=True)
        (state / "hook-context.json").write_text(
            json.dumps(data, indent=2), encoding="utf-8"
        )
    except OSError:
        pass  # fail open: si no podemos persistir, el peor caso es re-inyectar de más


def decide_injection(key: list, last_key: list, turns: int, full_every: int) -> tuple[str, int]:
    """PURA (testeable): dado el estado actual vs el último inyectado y el contador
    de turnos sin cambio, decide qué inyectar este turno.

    Devuelve ("full"|"breadcrumb", turns_para_guardar). "full" en on-step-change
    o al llegar al backstop de N turnos; reinicia el contador. Si no, "breadcrumb"
    y suma 1.
    """
    changed = list(key) != list(last_key)
    if changed or turns >= full_every:
        return ("full", 0)
    return ("breadcrumb", turns + 1)


def _render_full(cc: dict) -> str:
    """Envuelve el slice como contexto inyectable, rotulado como autoritativo
    (decisión #1: el más reciente reemplaza a los anteriores del transcript)."""
    return (
        "## SpecForge — current step (re-grounding; supersedes earlier slices)\n\n"
        + cc.get("breadcrumb", "")
        + "\n\n```json\n"
        + json.dumps(cc, indent=2, ensure_ascii=False)
        + "\n```"
    )


def user_prompt_context(project_dir: str, session_id: str) -> str:
    """Decide e arma lo que se inyecta en UserPromptSubmit. "" = nada (fail open)."""
    sf_dir = _specforge_root(project_dir)
    if sf_dir is None:
        return ""  # no es un proyecto SpecForge
    out = _run_sf(["context", "current", project_dir])
    if not out:
        return ""  # sf ausente o error → no inyectamos
    try:
        cc = json.loads(out)
    except json.JSONDecodeError:
        return ""
    breadcrumb = cc.get("breadcrumb", "")
    if not cc.get("feature"):
        return breadcrumb  # sin feature activa: solo la línea, sin gastar en slice

    # Clave de estado: feature/phase/wave. Si cambió → on-step-change.
    key = [cc.get("feature"), cc.get("phase"), cc.get("wave")]
    st = _hook_state(sf_dir)
    entry = st.get(session_id) or {}

    # force_full: lo setea PreCompact. Tras compactar, los slices previos del
    # transcript se resumieron/perdieron → forzamos un slice completo este turno.
    if entry.get("force_full"):
        st[session_id] = {"key": key, "turns_since_full": 0}  # consume el flag
        _save_hook_state(sf_dir, st)
        return _render_full(cc)

    mode, turns_next = decide_injection(
        key, entry.get("key", []), int(entry.get("turns_since_full", 0)), _full_slice_every()
    )
    st[session_id] = {"key": key, "turns_since_full": turns_next}
    _save_hook_state(sf_dir, st)

    return _render_full(cc) if mode == "full" else breadcrumb


def force_full_slice(project_dir: str, session_id: str) -> None:
    """Marca la sesión para que el próximo UserPromptSubmit inyecte el slice
    COMPLETO (no solo el breadcrumb). Lo llama PreCompact."""
    sf_dir = _specforge_root(project_dir)
    if sf_dir is None or not session_id:
        return
    st = _hook_state(sf_dir)
    entry = st.get(session_id) or {}
    entry["force_full"] = True
    st[session_id] = entry
    _save_hook_state(sf_dir, st)


# ── Memory reconciler (Stop nudge) ───────────────────────────────────────────
#
# Cuando una feature queda archivada SIN lecciones en el journal, el hook Stop
# nudgea UNA vez: "extraé lecciones y guardá con sf journal add". Decisión del
# debate: nudge-once, no hard-block. Como Stop solo puede comunicar bloqueando,
# bloqueamos a lo sumo una vez por feature (registrado en .state) y después
# stand-down — recuerda una vez, nunca atrapa en loop.


def _journaled_features(sf_dir: Path) -> set:
    """Features que YA tienen una entrada en specforge/journal/ (leemos el campo
    `feature` del .json, robusto ante nombres con guiones)."""
    out: set = set()
    jdir = sf_dir / "journal"
    if jdir.is_dir():
        for p in jdir.glob("*.json"):
            try:
                out.add(json.loads(p.read_text(encoding="utf-8")).get("feature", ""))
            except (OSError, json.JSONDecodeError):
                pass
    return out


def pick_journal_nudge(features: list, journaled: set, nudged: set) -> str | None:
    """PURA (testeable): primera feature archivada que no esté journaleada ni ya
    nudgeada, o None."""
    for f in features:
        name = f.get("name", "")
        if f.get("status") in ARCHIVED_STATUSES and name and name not in journaled and name not in nudged:
            return name
    return None


def stop_nudge(project_dir: str) -> str | None:
    """Decide a qué feature nudgear en Stop (una sola vez), y lo registra."""
    sf = _specforge_root(project_dir)
    if sf is None:
        return None
    try:
        features = json.loads((sf / "features.json").read_text(encoding="utf-8")).get("features", [])
    except (OSError, json.JSONDecodeError):
        return None
    st = _hook_state(sf)
    nudged = set(st.get("journal_nudged", []))
    target = pick_journal_nudge(features, _journaled_features(sf), nudged)
    if target is None:
        return None
    nudged.add(target)  # nudge-once: lo marcamos antes de devolver
    st["journal_nudged"] = sorted(nudged)
    _save_hook_state(sf, st)
    return target


def _journal_nudge_reason(feature: str) -> str:
    return (
        f"Feature `{feature}` quedó archivada sin lecciones registradas. "
        f"Extraé las lecciones durables y guardalas con:\n"
        f"  sf journal add --feature={feature} --json -\n"
        f"(Recordatorio único — no volverá a aparecer.)"
    )


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
    if event == "UserPromptSubmit":
        ctx = user_prompt_context(_project_dir(payload), payload.get("session_id", ""))
        if ctx:
            print(json.dumps({"hookSpecificOutput": {
                "hookEventName": "UserPromptSubmit",
                "additionalContext": ctx,
            }}))
        return 0
    if event == "Stop":
        # stop_hook_active = ya estamos en una continuación forzada por un Stop
        # hook → no volver a bloquear (evita loops); nuestro nudge-once por feature
        # ya lo evita, esto es defensa extra.
        if payload.get("stop_hook_active"):
            return 0
        target = stop_nudge(_project_dir(payload))
        if target:
            print(json.dumps({"decision": "block", "reason": _journal_nudge_reason(target)}))
        return 0
    if event == "SessionEnd":
        mark_session(_project_dir(payload), "session ended")
        return 0
    if event == "PreCompact":
        pd = _project_dir(payload)
        mark_session(pd, "pre-compaction checkpoint")
        force_full_slice(pd, payload.get("session_id", "")) # re-grounding post-compact
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
    elif event == "user_prompt_submit":
        print(json.dumps({"decision": "allow", "context": user_prompt_context(pd, payload.get("session_id", ""))}))
    elif event == "stop":
        target = stop_nudge(pd)
        if target:
            print(json.dumps({"decision": "block", "reason": _journal_nudge_reason(target)}))
        else:
            print(json.dumps({"decision": "allow"}))
    elif event == "session_end":
        mark_session(pd, "session ended")
        print(json.dumps({"decision": "allow"}))
    elif event == "pre_compact":
        mark_session(pd, "pre-compaction checkpoint")
        force_full_slice(pd, payload.get("session_id", ""))
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
        # JSON-first: el .json (la fuente) se gatea igual que el .md
        check(decide_pre_tool_use(str(proj), str(feats / "tasks.json"))[0] == "deny",
              "tasks.json denied too (json-first source)")
        # plan necesita el gate de tasks (y vive bajo progress/)
        check(decide_pre_tool_use(str(proj), str(feats / "progress" / "plan.md"))[0] == "deny",
              "plan denied without tasks gate")

        # cadena completa aprobada → plan permitido
        reg.write_text(json.dumps({"schema_version": "1.0", "features": [
            {"name": "x", "gates": [
                {"phase": "requirements", "result": "approve"},
                {"phase": "design", "result": "approve"},
                {"phase": "tasks", "result": "approve"},
            ]}
        ]}))
        check(decide_pre_tool_use(str(proj), str(feats / "progress" / "plan.json"))[0] == "allow",
              "plan.json allowed after tasks gate")

        # Serial (F22): con 'y' en curso, arrancar 'z' (su requirements) se deniega;
        # el requirements de la propia feature activa sí se permite.
        reg.write_text(json.dumps({"schema_version": "1.0", "features": [
            {"name": "y", "status": "building"},
            {"name": "z", "status": "planned"},
        ]}))
        zreq = proj / "specforge" / "features" / "z" / "requirements.md"
        yreq = proj / "specforge" / "features" / "y" / "requirements.md"
        check(decide_pre_tool_use(str(proj), str(zreq))[0] == "deny",
              "serial: 2nd feature's requirements denied while another active")
        check(decide_pre_tool_use(str(proj), str(yreq))[0] == "allow",
              "serial: active feature's own requirements allowed")
        reg.write_text(json.dumps({"schema_version": "1.0", "features": [
            {"name": "z", "status": "planned"},
        ]}))
        check(decide_pre_tool_use(str(proj), str(zreq))[0] == "allow",
              "serial: requirements allowed when none active")

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

    # ── decide_injection (trigger del slice, pura) ──
    # estado igual y contador bajo → solo breadcrumb, contador++
    check(decide_injection(["f", "build", 1], ["f", "build", 1], 3, 10) == ("breadcrumb", 4),
          "sin cambio + turnos bajos → breadcrumb")
    # estado cambió → slice completo, contador a 0
    check(decide_injection(["f", "design", None], ["f", "requirements", None], 2, 10) == ("full", 0),
          "on-step-change → full")
    # sin cambio pero alcanzó el backstop de N turnos → full, reset
    check(decide_injection(["f", "build", 0], ["f", "build", 0], 10, 10) == ("full", 0),
          "backstop N turnos → full")

    # ── pick_journal_nudge (conciliador, pura) ──
    feats = [
        {"name": "old", "status": "done"},
        {"name": "cur", "status": "building"},
        {"name": "add-export", "status": "archived"},  # nombre con guion
    ]
    check(pick_journal_nudge(feats, set(), set()) == "old",
          "primera archivada sin journal → nudge")
    check(pick_journal_nudge(feats, {"old"}, set()) == "add-export",
          "ya journaleada se saltea (respeta nombres con guion)")
    check(pick_journal_nudge(feats, {"old", "add-export"}, set()) is None,
          "todas journaleadas → None")
    check(pick_journal_nudge(feats, set(), {"old"}) == "add-export",
          "ya nudgeada se saltea (nudge-once)")

    # ── force_full_slice (PreCompact → re-grounding) ──
    with tempfile.TemporaryDirectory() as d3:
        (Path(d3) / "specforge").mkdir(parents=True)
        force_full_slice(d3, "sess-x")
        st = _hook_state(Path(d3) / "specforge")
        check(st.get("sess-x", {}).get("force_full") is True,
              "force_full_slice marca el flag por sesión")

    # ── _full_slice_every (configurable por env) ──
    os.environ.pop("SPECFORGE_FULL_SLICE_EVERY", None)
    check(_full_slice_every() == FULL_SLICE_EVERY, "sin env → default")
    os.environ["SPECFORGE_FULL_SLICE_EVERY"] = "3"
    check(_full_slice_every() == 3, "env válido → override")
    os.environ["SPECFORGE_FULL_SLICE_EVERY"] = "junk"
    check(_full_slice_every() == FULL_SLICE_EVERY, "env basura → default")
    os.environ.pop("SPECFORGE_FULL_SLICE_EVERY", None)

    print(f"\n{'OK' if not failures else 'FAILED'}: {failures} failure(s).")
    return 1 if failures else 0


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
