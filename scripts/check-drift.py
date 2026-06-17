#!/usr/bin/env python3
"""Detect drift between archived SpecForge specs and the code (F33).

A spec that isn't kept alive becomes a lie. This reads each feature's
`trace.json` (the structured matrix emitted by sf-check) and verifies, for every
requirement, that its code anchor still exists. Cheap by construction: the matrix
already holds the exact `path:symbol` pointers, so there is no whole-repo
re-analysis.

Two depths:
- **static** (default): does the `path:symbol` anchor still exist in the code?
  Runs anywhere, no test runner needed — good for CI.
- **dynamic** (`--run-tests`): also run each referenced test id. Opt-in because it
  needs the project's test command.

Maintainer/CI tool, stdlib only. Reabsorbed as `sf doctor --drift` when the Go
CLI lands.

Usage:
    python3 scripts/check-drift.py [PROJECT_DIR]      # static drift, exit 1 if drift
    python3 scripts/check-drift.py --run-tests "pytest -q {test}"
    python3 scripts/check-drift.py --selftest
"""

from __future__ import annotations

import json
import re
import subprocess
import sys
from pathlib import Path


def load_traces(specforge: Path) -> list[tuple[str, Path]]:
    """Return (feature_name, trace_path) for every trace.json under archive/ and features/."""
    out: list[tuple[str, Path]] = []
    for base in (specforge / "archive", specforge / "features"):
        if base.is_dir():
            for tj in sorted(base.glob("*/trace.json")):
                out.append((tj.parent.name, tj))
    return out


def check_anchor(project_dir: Path, anchor: str) -> tuple[bool, str]:
    """Static check: does `path:symbol` still exist? Returns (ok, reason)."""
    if ":" not in anchor:
        path_part, symbol = anchor, None
    else:
        path_part, symbol = anchor.rsplit(":", 1)
    f = project_dir / path_part
    if not f.exists():
        return (False, f"file gone: {path_part}")
    if symbol:
        ident = re.escape(symbol.split(".")[-1])  # final identifier of Klass.method
        if not re.search(rf"\b{ident}\b", f.read_text(encoding="utf-8", errors="ignore")):
            return (False, f"symbol gone: {symbol} not in {path_part}")
    return (True, "")


def run_test(project_dir: Path, test_id: str, template: str) -> tuple[bool, str]:
    cmd = template.replace("{test}", test_id)
    try:
        r = subprocess.run(cmd, shell=True, cwd=project_dir, capture_output=True, text=True, timeout=300)
    except (subprocess.SubprocessError, OSError) as exc:
        return (False, f"could not run test: {exc}")
    return (r.returncode == 0, "" if r.returncode == 0 else f"test failed: {test_id}")


def check_feature(project_dir: Path, feature: str, trace_path: Path, run_tests: str | None) -> list[str]:
    drifts: list[str] = []
    try:
        data = json.loads(trace_path.read_text(encoding="utf-8"))
    except (json.JSONDecodeError, OSError) as exc:
        return [f"{feature}: unreadable trace.json ({exc})"]
    for req, info in (data.get("requirements") or {}).items():
        for anchor in info.get("code", []):
            ok, reason = check_anchor(project_dir, anchor)
            if not ok:
                drifts.append(f"{feature} / {req}: {reason}")
        if run_tests:
            for test_id in info.get("test", []):
                ok, reason = run_test(project_dir, test_id, run_tests)
                if not ok:
                    drifts.append(f"{feature} / {req}: {reason}")
    return drifts


def run(project_dir: Path, run_tests: str | None, quiet: bool) -> int:
    specforge = project_dir / "specforge"
    if not specforge.is_dir():
        print(f"No specforge/ under {project_dir} — nothing to check.")
        return 0
    traces = load_traces(specforge)
    drifts: list[str] = []
    for feature, tj in traces:
        drifts.extend(check_feature(project_dir, feature, tj, run_tests))
    if not quiet:
        print(f"Checked {len(traces)} feature(s) with trace.json.")
    for d in drifts:
        print(f"  DRIFT: {d}")
    if drifts:
        print(f"\nDRIFT: {len(drifts)} anchor(s) diverged from the spec.")
        return 1
    print("\nOK: specs and code in sync.")
    return 0


def selftest() -> int:
    import tempfile

    failures = 0

    def check(cond: bool, label: str) -> None:
        nonlocal failures
        print(("PASS" if cond else "FAIL") + f"  {label}")
        failures += 0 if cond else 1

    with tempfile.TemporaryDirectory() as d:
        proj = Path(d)
        (proj / "src").mkdir(parents=True)
        (proj / "src" / "slug.py").write_text("def slugify(text):\n    return text\n")
        arch = proj / "specforge" / "archive" / "2026-06-16-slugify"
        arch.mkdir(parents=True)
        (arch / "trace.json").write_text(json.dumps({
            "feature": "slugify",
            "requirements": {"R1": {"code": ["src/slug.py:slugify"], "test": [], "status": "ok"}},
        }))

        check(run(proj, None, quiet=True) == 0, "clean project: no drift")

        # remove the symbol → drift
        (proj / "src" / "slug.py").write_text("def something_else():\n    return 1\n")
        check(run(proj, None, quiet=True) == 1, "symbol removed: drift detected")

        # remove the file → drift
        (proj / "src" / "slug.py").unlink()
        check(run(proj, None, quiet=True) == 1, "file removed: drift detected")

        # no specforge/ → no-op pass
        with tempfile.TemporaryDirectory() as d2:
            check(run(Path(d2), None, quiet=True) == 0, "non-SpecForge dir: no-op")

    print(f"\n{'OK' if not failures else 'FAILED'}: {failures} failure(s).")
    return 1 if failures else 0


def main(argv: list[str]) -> int:
    if "--selftest" in argv:
        return selftest()
    run_tests = None
    if "--run-tests" in argv:
        i = argv.index("--run-tests")
        run_tests = argv[i + 1] if i + 1 < len(argv) else "pytest -q {test}"
        argv = argv[:i] + argv[i + 2:]
    positionals = [a for a in argv if not a.startswith("-")]
    project_dir = Path(positionals[0]).resolve() if positionals else Path.cwd()
    return run(project_dir, run_tests, quiet="--quiet" in argv)


if __name__ == "__main__":
    raise SystemExit(main(sys.argv[1:]))
