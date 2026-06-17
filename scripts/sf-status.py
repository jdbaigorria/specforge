#!/usr/bin/env python3
"""Project-level health view for a SpecForge project (F37).

Not a dashboard — a deterministic *render of data you already have*: the
features.json gate ledger (F10), the lane (F34), the trace.json drift link
(F33), and review verdicts. Answers "what's the state of the whole project?" at a
glance, which `/sf-status <feature>` (single feature) and `sf-audit` (point-in-time
adversarial) don't.

Per feature it shows: lane, status, current phase (last approved gate), drift
(archived specs vs code), gaps (from the review verdict), and what it's blocked
on (`depends_on`). It also prints a dependency-ordered critical path.

Maintainer/CI tool, stdlib only. Reabsorbed as `sf status` when the Go CLI lands.

Usage:
    python3 scripts/sf-status.py [PROJECT_DIR]
    python3 scripts/sf-status.py --selftest
"""

from __future__ import annotations

import importlib.util
import json
import sys
from pathlib import Path

DONE = {"done"}


def _load_drift():
    """Reuse the drift engine from check-drift.py (same dir), or None if absent."""
    p = Path(__file__).resolve().parent / "check-drift.py"
    if not p.exists():
        return None
    spec = importlib.util.spec_from_file_location("_drift", p)
    mod = importlib.util.module_from_spec(spec)
    try:
        spec.loader.exec_module(mod)
        return mod
    except Exception:
        return None


def _trace_path(specforge: Path, name: str) -> Path | None:
    for base in (specforge / "archive", specforge / "features"):
        if base.is_dir():
            for tj in base.glob(f"*{name}/trace.json"):
                if tj.parent.name == name or tj.parent.name.endswith(name):
                    return tj
    return None


def _review_path(specforge: Path, name: str) -> Path | None:
    for base in (specforge / "archive", specforge / "features"):
        if base.is_dir():
            for rv in base.glob(f"*{name}/review.md"):
                if rv.parent.name == name or rv.parent.name.endswith(name):
                    return rv
    return None


def _last_phase(feature: dict) -> str:
    gates = feature.get("gates") or []
    approved = [g for g in gates if g.get("result", "").startswith("approve")]
    return approved[-1]["phase"] if approved else "—"


def _gaps(specforge: Path, name: str) -> str:
    """Read the authoritative verdict line, not stray words in prose."""
    rv = _review_path(specforge, name)
    if not rv:
        return "—"
    for line in rv.read_text(encoding="utf-8", errors="ignore").splitlines():
        if "verdict" in line.lower():
            up = line.upper()
            if "APPROVE WITH NOTES" in up:
                return "notes"
            if "REVISE" in up:
                return "gaps"
            if "APPROVE" in up:
                return "ok"
    return "—"


def _drift_state(drift_mod, project_dir: Path, specforge: Path, name: str) -> str:
    tj = _trace_path(specforge, name)
    if tj is None:
        return "—"
    if drift_mod is None:
        return "?"
    issues = drift_mod.check_feature(project_dir, name, tj, None)
    return "DRIFT" if issues else "ok"


def _topo_order(features: dict[str, dict]) -> tuple[list[str], list[str]]:
    """Return (ordered_names, cycle_names). Stable, dependency-first."""
    order: list[str] = []
    temp: set[str] = set()
    done: set[str] = set()
    cycle: list[str] = []

    def visit(n: str, stack: set[str]) -> None:
        if n in done or n not in features:
            return
        if n in stack:
            cycle.append(n)
            return
        stack = stack | {n}
        for dep in features[n].get("depends_on") or []:
            visit(dep, stack)
        done.add(n)
        order.append(n)

    for name in features:
        visit(name, set())
    return order, cycle


def render(project_dir: Path) -> int:
    specforge = project_dir / "specforge"
    fj = specforge / "features.json"
    if not fj.exists():
        print(f"No specforge/features.json under {project_dir}.")
        return 0
    data = json.loads(fj.read_text(encoding="utf-8"))
    feats = {f["name"]: f for f in data.get("features", [])}
    if not feats:
        print("No features registered yet.")
        return 0

    drift_mod = _load_drift()
    order, cycle = _topo_order(feats)

    rows = []
    for name in order:
        f = feats[name]
        blockers = [d for d in (f.get("depends_on") or []) if feats.get(d, {}).get("status") not in DONE]
        rows.append({
            "feature": name,
            "lane": f.get("lane") or "—",
            "status": f.get("status") or "—",
            "phase": _last_phase(f),
            "drift": _drift_state(drift_mod, project_dir, specforge, name),
            "gaps": _gaps(specforge, name),
            "blocked": ",".join(blockers) if blockers else "—",
        })

    cols = ["feature", "lane", "status", "phase", "drift", "gaps", "blocked"]
    widths = {c: max(len(c), *(len(str(r[c])) for r in rows)) for c in cols}
    line = lambda r: "  ".join(str(r[c]).ljust(widths[c]) for c in cols)
    print("SpecForge — project status\n")
    print(line({c: c for c in cols}))
    print("  ".join("-" * widths[c] for c in cols))
    for r in rows:
        print(line(r))

    print("\nCritical path (dependency order):")
    print("  " + " → ".join(order))
    if cycle:
        print(f"  ⚠ dependency cycle involving: {', '.join(sorted(set(cycle)))}")

    drifted = [r["feature"] for r in rows if r["drift"] == "DRIFT"]
    blocked = [r["feature"] for r in rows if r["blocked"] != "—"]
    if drifted:
        print(f"\n⚠ drifted (spec ≠ code): {', '.join(drifted)} — run sf-amend")
    if blocked:
        print(f"⚠ blocked: " + "; ".join(f"{r['feature']}←{r['blocked']}" for r in rows if r['blocked'] != '—'))
    return 0


def selftest() -> int:
    import io
    import tempfile
    from contextlib import redirect_stdout

    failures = 0

    def check(cond: bool, label: str) -> None:
        nonlocal failures
        print(("PASS" if cond else "FAIL") + f"  {label}")
        failures += 0 if cond else 1

    with tempfile.TemporaryDirectory() as d:
        proj = Path(d)
        sf = proj / "specforge"
        sf.mkdir()
        (sf / "features.json").write_text(json.dumps({"schema_version": "1.0", "features": [
            {"name": "base", "status": "done", "lane": "standard", "depends_on": [],
             "gates": [{"phase": "verdict", "result": "approve"}]},
            {"name": "feat-b", "status": "approved", "lane": "lite", "depends_on": ["base"],
             "gates": [{"phase": "lane", "result": "approve"}]},
            {"name": "feat-c", "status": "approved", "lane": "standard", "depends_on": ["missing-dep"],
             "gates": []},
        ]}))
        buf = io.StringIO()
        with redirect_stdout(buf):
            render(proj)
        out = buf.getvalue()
        check("base" in out and "feat-b" in out, "renders all features")
        check(out.index("base") < out.index("feat-b"), "dependency order: base before feat-b")
        check("feat-c" in out and "missing-dep" in out, "shows unmet dependency as blocker")
        check("lite" in out, "shows lane")
        check("Critical path" in out, "prints critical path")

    print(f"\n{'OK' if not failures else 'FAILED'}: {failures} failure(s).")
    return 1 if failures else 0


def main(argv: list[str]) -> int:
    if "--selftest" in argv:
        return selftest()
    positionals = [a for a in argv if not a.startswith("-")]
    project_dir = Path(positionals[0]).resolve() if positionals else Path.cwd()
    return render(project_dir)


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
