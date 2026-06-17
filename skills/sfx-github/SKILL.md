---
name: sfx-github
delegate: true
description: >
  Execute git workflow: branch, commit, push, PR, merge. Trigger: "/github", "/git",
  "create a PR", "push this", "branch for", "commit", "merge", "release".
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

## Rules

- NEVER commit directly to main. ALWAYS PR.
- Squash merge default. Delete branch after merge.
- If CI fails, report — do NOT auto-merge.
- If uncommitted changes, ask before stashing.
