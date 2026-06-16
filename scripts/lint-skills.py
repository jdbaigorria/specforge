#!/usr/bin/env python3
"""Lint the SpecForge skill suite for consistency.

Maintainer/CI tool — NOT part of the runtime. End users who only consume the
markdown skills never run this. Stdlib only (no PyYAML) so it runs on any
python3 with zero install. The checks here are the asset; when the Go CLI lands
they get reabsorbed as `sf lint` / `sf doctor`.

Usage:
    python3 scripts/lint-skills.py          # lint, exit 1 on any error
    python3 scripts/lint-skills.py --quiet   # only print problems

Exit code: 0 = clean, 1 = at least one error. Warnings never fail the build.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SKILLS_DIR = ROOT / "skills"

# Product docs the lint governs. Scratch/source files (CLI-SPEC, update-sf,
# swarm, SPECFORGE-REVIEW) are intentionally excluded — the review legitimately
# documents the old `sdd-` names.
PRODUCT_DOCS = [
    "AGENT.md",
    "README.md",
    "README.es.md",
    "SUPPORT-SKILLS.md",
    "SUPPORT-SKILLS.es.md",
    "INSTALL.md",
]

NAME_RE = re.compile(r"^sfx?-[a-z][a-z0-9-]*$")
DESC_MAX = 1024

# `sf-*` / `sfx-*` tokens that are NOT skills (commands, planned/future skills).
# Referencing one of these does not count as a broken skill reference.
NON_SKILL_TOKENS = {
    "sf-status",      # command
    "sf-doctor",      # planned CLI command
    "sf-amend",       # planned (F33)
    "sf-discover",    # planned (F27)
    "sf-prd",         # planned (F27)
    "sf-lint",        # planned CLI command
    "sf-gate",        # planned CLI command
    "sf-traceability",
    "sf-save",
    "sf-op",
    "sf-sync",
    "sf-init",        # also a skill, listed for safety
}

TOKEN_RE = re.compile(r"\bsfx?-[a-z][a-z0-9-]*\b")
LINK_RE = re.compile(r"\[[^\]]+\]\(([^)]+)\)")


class Report:
    def __init__(self) -> None:
        self.errors: list[str] = []
        self.warnings: list[str] = []

    def error(self, msg: str) -> None:
        self.errors.append(msg)

    def warn(self, msg: str) -> None:
        self.warnings.append(msg)


def parse_frontmatter(text: str) -> dict[str, str] | None:
    """Return the YAML-ish frontmatter as a flat dict, or None if absent.

    Hand-rolled to avoid a PyYAML dependency. Handles `key: value` and the
    folded `key: >` block scalar used by SKILL.md descriptions.
    """
    if not text.startswith("---"):
        return None
    end = text.find("\n---", 3)
    if end == -1:
        return None
    block = text[3:end].strip("\n")
    fields: dict[str, str] = {}
    key: str | None = None
    buf: list[str] = []
    for line in block.splitlines():
        m = re.match(r"^([a-zA-Z_][\w-]*):\s*(.*)$", line)
        if m and not line.startswith(" "):
            if key is not None:
                fields[key] = " ".join(buf).strip()
            key, rest = m.group(1), m.group(2).strip()
            buf = []
            if rest and rest != ">" and rest != "|":
                buf.append(rest)
        elif key is not None:
            buf.append(line.strip())
    if key is not None:
        fields[key] = " ".join(buf).strip()
    return fields


def strip_code_blocks(text: str) -> str:
    """Blank out fenced code blocks so examples inside them are ignored."""
    out, in_fence = [], False
    for line in text.splitlines():
        if line.lstrip().startswith("```"):
            in_fence = not in_fence
            out.append("")
            continue
        out.append("" if in_fence else line)
    return "\n".join(out)


def discover_skills() -> list[Path]:
    return sorted(p for p in SKILLS_DIR.iterdir() if p.is_dir())


def check_skills(skills: list[Path], names: set[str], rep: Report) -> None:
    for d in skills:
        skill_md = d / "SKILL.md"
        if not skill_md.exists():
            rep.error(f"{d.name}: missing SKILL.md")
            continue
        text = skill_md.read_text(encoding="utf-8")
        fm = parse_frontmatter(text)
        if fm is None:
            rep.error(f"{d.name}/SKILL.md: missing or malformed frontmatter")
            continue
        name = fm.get("name", "")
        desc = fm.get("description", "")
        if not name:
            rep.error(f"{d.name}/SKILL.md: frontmatter has no `name`")
        else:
            if name != d.name:
                rep.error(f"{d.name}/SKILL.md: name `{name}` != folder `{d.name}`")
            if not NAME_RE.match(name):
                rep.error(f"{d.name}/SKILL.md: name `{name}` fails sf-/sfx- regex")
        if not desc:
            rep.error(f"{d.name}/SKILL.md: frontmatter has no `description`")
        elif len(desc) > DESC_MAX:
            rep.error(f"{d.name}/SKILL.md: description {len(desc)} chars > {DESC_MAX}")


def check_no_sdd(files: list[Path], rep: Report) -> None:
    for f in files:
        for i, line in enumerate(f.read_text(encoding="utf-8").splitlines(), 1):
            if "sdd-" in line:
                rep.error(f"{f.relative_to(ROOT)}:{i}: leftover `sdd-` reference")


def check_skill_refs(files: list[Path], names: set[str], rep: Report) -> None:
    known = names | NON_SKILL_TOKENS
    for f in files:
        body = strip_code_blocks(f.read_text(encoding="utf-8"))
        for i, line in enumerate(body.splitlines(), 1):
            for tok in TOKEN_RE.findall(line):
                if tok not in known:
                    rep.warn(f"{f.relative_to(ROOT)}:{i}: reference to unknown skill `{tok}`")


def check_relative_links(files: list[Path], rep: Report) -> None:
    for f in files:
        body = strip_code_blocks(f.read_text(encoding="utf-8"))
        for i, line in enumerate(body.splitlines(), 1):
            for target in LINK_RE.findall(line):
                target = target.strip()
                if target.startswith(("http://", "https://", "#", "mailto:")):
                    continue
                path_part = target.split("#", 1)[0]
                if not path_part:
                    continue
                if not (f.parent / path_part).exists():
                    rep.error(f"{f.relative_to(ROOT)}:{i}: broken relative link `{target}`")


def headers(path: Path) -> list[str]:
    return [
        ln.strip()
        for ln in strip_code_blocks(path.read_text(encoding="utf-8")).splitlines()
        if ln.startswith("## ")
    ]


def check_parity(rep: Report) -> None:
    pairs = [("README.md", "README.es.md"), ("SUPPORT-SKILLS.md", "SUPPORT-SKILLS.es.md")]
    for en_name, es_name in pairs:
        en, es = ROOT / en_name, ROOT / es_name
        if not (en.exists() and es.exists()):
            rep.error(f"parity: missing one of {en_name} / {es_name}")
            continue
        ne, nes = len(headers(en)), len(headers(es))
        if ne != nes:
            rep.error(f"parity: {en_name} has {ne} `##` sections, {es_name} has {nes}")
        skills_en = sum(1 for h in headers(en) if h.startswith("## sfx-"))
        skills_es = sum(1 for h in headers(es) if h.startswith("## sfx-"))
        if skills_en != skills_es:
            rep.error(
                f"parity: {en_name} documents {skills_en} sfx- skills, {es_name} {skills_es}"
            )


def main() -> int:
    quiet = "--quiet" in sys.argv
    rep = Report()

    skills = discover_skills()
    names = {d.name for d in skills}
    md_files = sorted(SKILLS_DIR.rglob("*.md")) + [ROOT / d for d in PRODUCT_DOCS if (ROOT / d).exists()]

    check_skills(skills, names, rep)
    check_no_sdd(md_files, rep)
    check_skill_refs(md_files, names, rep)
    check_relative_links(md_files, rep)
    check_parity(rep)

    if not quiet:
        print(f"Linted {len(skills)} skills, {len(md_files)} markdown files.")
    for w in rep.warnings:
        print(f"  warning: {w}")
    for e in rep.errors:
        print(f"  ERROR:   {e}")

    if rep.errors:
        print(f"\nFAIL: {len(rep.errors)} error(s), {len(rep.warnings)} warning(s).")
        return 1
    print(f"\nOK: 0 errors, {len(rep.warnings)} warning(s).")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
