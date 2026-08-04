# SpecForge ↔ git & issue tracker

How SpecForge artefacts map to git and to an issue tracker (F35/F36). All of this
is **opt-in convention**, not enforced — document the recommended flow, automate
with hooks where it helps, never impose.

## Git mapping (F36)

| SpecForge | Git |
|-----------|-----|
| A feature | a branch `feature/<slug>` off main |
| A build wave | one commit (`feat(<slug>): wave N — <tasks done>`) |
| `sf-check` archive (on APPROVE) | merge the branch to main |

The git history then *is* the build op-log: one commit per wave, each naming the
tasks it completed. That gives an auditable record for free, complementing the
`features.json` gate ledger (F10) — no separate op-log needed.

Keep it a recommended flow. A solo user on a scratch project can ignore it; a
team turns it on. Where the enforcement hooks (F29) are installed, a wave gate
passing can optionally trigger the commit — but the convention stands on its own.

## PR = the verdict gate (F35, team mode)

Single-player: the human at each gate is the author. That's fine for the creation
gates (requirements/design/tasks/plan/wave) — those are in-progress design
decisions.

Team: the artefacts live in the repo, so they travel in the PR. The reviewer sees
`requirements.md`, `design.md`, and `review.md` alongside the diff and approves
**code and spec together**. So:

- **Creation gates** (propose/build) → the author's, on the branch.
- **Verdict gate** (`sf-check`) → the team's, and it **maps to the PR approval**.
  Don't run a separate SpecForge verdict gate *and* a PR review — the PR approval
  satisfies it. Map, don't duplicate.

## Ownership (F35)

No custom permission system — lean on what git already has:

- `owners` in `constitution.md` — who may approve changes to **invariants** (the
  highest-stakes edits).
- `CODEOWNERS` (git) — who reviews which paths for everything else. A minimal
  example at the repo root:

  ```
  # CODEOWNERS
  /specforge/constitution.md   @tech-lead
  /specforge/                  @backend-team
  *                            @backend-team
  ```

## Issue export (F36) — one way only

`sf-propose --all` produces a roadmap; export it so the team's tracker stays the
single "what to do" index without duplicating the spec detail.

- Read `specforge/roadmap.md` + `specforge/features.json`.
- For each feature, open one issue (via `gh issue create`, or the tracker's API):
  - **title:** the feature name
  - **body:** one-line summary + a **back-link** to `specforge/features/<slug>/`
    (or the archived path) + its `status`/`lane`
  - **labels:** optionally `specforge`, the lane, the status
- Record the created issue URL back into the feature (e.g. a `tracker_url` field)
  so the link is two-way as *reference*, but the **source of truth stays one
  direction**: SpecForge → tracker. Never import issue edits back into the spec.

Why one-way: bidirectional sync between a spec system and an issue tracker is the
classic expensive, fragile integration. The tracker indexes *what*; SpecForge owns
*the detail*. One direction, one truth.

GitHub is the grounded default (`gh`). Linear/Jira work the same way through their
API/MCP — same one-way mapping.
