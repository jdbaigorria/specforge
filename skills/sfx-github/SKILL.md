---
name: sfx-github
delegate: true
description: >
  Execute git workflow: branch, commit, push, PR, merge. Also the home for SpecForge↔git
  integration: feature branches, one-commit-per-wave, archive=merge, and one-way export of the
  roadmap to issues. Trigger: "/github", "/git", "create a PR", "push this", "branch for",
  "commit", "merge", "release", "export the roadmap to issues", "open issues for the features".
---

# github

Execute git operations. Return clean summary. Model: haiku.

## Workflows

| Intent | Action |
|--------|--------|
| "start working on X" | Branch from main |
| "commit this" | Stage + conventional commit |
| "push" / "create PR" | Push + PR |
| "merge" / "done" | Squash merge, delete branch |

## Conventions

Branches: `feature/`, `fix/`, `refactor/`, `docs/`, `chore/` — all from main, all via PR.
Commits: `type(scope): description` — imperative, max 72 chars. Breaking: `type(scope)!:`
Types: feat, fix, refactor, docs, test, chore, ci.

## SpecForge integration (optional, opt-in)

When the project uses SpecForge, this skill is the bridge to git and the issue
tracker. Read `references/specforge-integration.md` for the full mapping. In short:

- **One branch per feature** (`feature/<slug>`), **one commit per wave** (message
  references the wave + completed tasks), **`archive` = merge to main**. The git
  history becomes the build op-log for free.
- **The PR carries the artefacts.** `requirements.md`, `design.md`, `review.md`
  travel in the PR, so the reviewer approves code *and* spec together — **the PR
  approval IS the sf-check verdict gate** in team mode (F35).
- **Export the roadmap to issues** (`export the roadmap to issues`): read
  `specforge/roadmap.md` + `features.json` and open one issue per feature via
  `gh`, each with a back-link to its artefact folder. **One-way only** — the
  tracker is the "what to do" index, SpecForge is the detail. No bidirectional
  sync (that's the expensive, fragile trap).

## Rules

- NEVER commit directly to main. ALWAYS PR.
- Squash merge default. Delete branch after merge.
- If CI fails, report — do NOT auto-merge.
- If uncommitted changes, ask before stashing.
- Issue export is **one direction** (SpecForge → tracker). Never sync back.
