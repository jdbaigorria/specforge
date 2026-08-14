---
name: sfx-github
delegate: true
description: >
  Execute git workflow: branch, commit, push, PR, merge. Also the home for SpecForge↔git
  integration — in a SpecForge project `sf` already does the branch, the commit and the merge,
  so this covers what it deliberately does not touch: the remote, PRs and releases. Trigger:
  "/github", "/git", "create a PR", "push this", "branch for", "commit", "merge", "release".
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

**In a SpecForge project the conventions are not here — they are in the constitution.** Read
`git:` from `.docs/constitucion.md` (`patron_branch`, `commit`, `merge`, `branch_base`) and
follow it. A skill with git conventions of its own is a second source of truth that drifts from
the project's.

**Standalone**, with no constitution to read, these are the defaults:

Branches: `feature/`, `fix/`, `refactor/`, `docs/`, `chore/` — all from the base branch.
Commits: `type(scope): description` — imperative, max 72 chars. Breaking: `type(scope)!:`
Types: feat, fix, refactor, docs, test, chore, ci.

## SpecForge integration

Read `references/specforge-integration.md`. In short: **`sf` already does the branch, the commit
and the merge**, so most of what this skill would do is taken. What is left for you is
everything `sf` deliberately does not touch — **the remote**: push, PRs, releases, tags.

## Rules

- In a SpecForge project, read `git:` from the constitution. Never assume a convention.
- Never invent a branch or a merge strategy the project did not declare.
- If CI fails, report — do NOT auto-merge.
- If there are uncommitted changes, ask before stashing.
