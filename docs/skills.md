[← Back to README](../README.md)

## Skills reference

### sfp-scout (product tier, optional)

The from-zero front-end. De-risks a fuzzy idea before it becomes a project —
research the landscape, stress-test it, decide proceed/pivot/kill. Greenfield
only; brownfield skips straight to `sf-init`.

| | |
|---|---|
| **Triggers** | `sfp-scout`, "should I build X", "is this worth building", "de-risk this idea" |
| **Needs** | the research MCPs (web/GitHub/docs) — won't fabricate research without them |
| **Produces** | `specforge/product/<slug>/` — research, discovery brief (with `PR#` ids), decision log |
| **Hands off** | on *proceed* → `sf-init --from <brief>`; `PR#` ids flow into the roadmap |

De-risk, not validation — it gathers evidence and surfaces risk; it cannot prove
demand. Its product requirements (`PR#`) sit at the top of the traceability spine:
`PR# → feature → R# → task → code → test`.

---

### sf-init

Scaffolds the project and generates foundational context.

| | |
|---|---|
| **Triggers** | `sf-init`, "start a new project", "add specforge to this project" |
| **Flags** | `--from <path>` (import PRD/brief), `--path <dir>` (monorepo sub-project) |
| **Detects** | Greenfield vs brownfield (interactive for ambiguous cases like monorepos) |

**Greenfield flow:**
1. Scaffold directories (automatic)
2. Constitution conversation: identity, principles, constraints, anti-goals → 🔴 GATE
3. Generate `specforge/context/project.md` + `specforge/context/conventions.md` from constitution context → 🔴 GATE

**Brownfield flow:**
1. Scaffold directories (automatic)
2. Onboard: analyze codebase → `specforge/context/project.md` + `specforge/context/conventions.md` → 🔴 GATE
3. Constitution conversation: principles and anti-goals (shorter, stack already known) → 🔴 GATE

**Produces:**
- `specforge/constitution.md` — project identity, principles, constraints, anti-goals
- `specforge/history.md` — project log (first entry)
- `specforge/context/project.md` — stack and architecture context
- `specforge/context/conventions.md` — coding standards

---

### sf-propose

Generates the full specification for a feature: requirements, design, tasks.

| | |
|---|---|
| **Triggers** | `sf-propose <name>`, "specify", "add feature", "spec this" |
| **Flags** | `--design-first` (architecture → requirements), `--from-code` (reverse-engineer) |
| **Modes** | Requirements-first (default), Design-first, From-code |

**Requirements-first flow (default):**
1. Conversation: what, why, who, boundaries
2. Generate `requirements.md` with EARS notation → 🔴 GATE
3. Generate `design.md` (sections conditional on complexity) → 🔴 GATE
4. Generate `tasks.json` — a flat task list with `depends_on` + traceability matrix → 🔴 GATE
5. Register the feature (`sf feature add` → `features/<name>/feature.json`, status: `approved`)

**Design-first flow** (`--design-first`):
Inverts steps 2 and 3 — design first, then derive requirements from what the
architecture can deliver. For projects where technical constraints drive scope.

**From-code flow** (`--from-code`):
Reverse-engineers specs from existing code. Creates a `_baseline` feature with
`[INFERRED]` requirements. No tasks.md (already implemented).

**Produces per feature:**
- `specforge/features/<name>/requirements.md` — EARS requirements + acceptance criteria
- `specforge/features/<name>/design.md` — architecture, components, decisions
- `specforge/features/<name>/tasks.json` (+ `.md`) — flat tasks + `depends_on` (waves are computed at build)

**Design sections are conditional.** Only include sections relevant to the feature.
A CLI flag doesn't need Security Considerations. A payment endpoint does. Sections
that would say "N/A" are omitted to avoid context noise.

**Post-generation:** User can invoke clarify at any time (reads `references/clarify.md`).
Scans for ambiguities, completeness gaps, and inconsistencies.

---

### sf-build

Plans execution and implements wave by wave.

| | |
|---|---|
| **Triggers** | `sf-build <name>`, "build", "implement", "start building" |
| **Prerequisite** | Feature status must be `approved` |
| **Resumable** | If paused, reads `progress/` to find where it left off |

**Flow:**
1. `sf plan compute` — compute the wave layout from the tasks' `depends_on`
   graph (topological layering); optionally refine → 🔴 GATE
2. For each wave (per `build.mode`: inline, or a fresh sub-agent):
   a. Execute all tasks in the wave — climbing the minimal-code ladder per task
   b. Log progress to `progress/wave-<n>.md`
   c. Present results → 🔴 GATE (inline) / automated checkpoint (sub-agent)
3. All waves complete → status becomes `checking`

**Error recovery (3 levels):**
- **Minor** (typo, wrong import): fix inline, note in progress log
- **Design mismatch** (can't implement as specified): pause wave, present options to user → 🔴 GATE
- **Blocking dependency** (previous wave output wrong): pause, escalate to user

**Key rule:** Never modify specs during build. If specs need changes, pause and
escalate. Specs are the contract; build fulfills them.

**Produces per wave:**
- `specforge/features/<name>/progress/plan.json` (+ `.md`) — computed wave layout
- `specforge/features/<name>/progress/wave-<n>.md` — what was done, decisions, ladder rung
- Updated `tasks.json` — tasks marked `done` as completed

---

### sf-check

Validates implementation against specs. Archives on approval.

| | |
|---|---|
| **Triggers** | `sf-check <name>`, "check", "validate", "review" |
| **Prerequisite** | Feature status must be `checking` |
| **Includes** | Traceability, gap analysis, constitution compliance, backprop |

**Flow:**
1. Traceability analysis: every R# → task → implementation → test
2. Gap analysis: missing implementations, tests, design deviations, orphan code
3. Constitution compliance: validate against principles — including **parsimony**
   (`P-min`: over-engineering is a valid REVISE reason; see `references/minimal-code.md`)
4. Verdict: APPROVE / APPROVE WITH NOTES / REVISE
5. Present review → 🔴 GATE
6. If APPROVE → archive automatically

It also runs as a **phase auditor** (`sf-check --phase=<phase>`) — a narrow,
fresh-sub-agent quality check of a single just-finished phase at its gate, instead
of the full end-of-feature pass. See [the quality tier](cli-and-hooks.md#the-quality-tier-phase-auditor).

**Verdicts:**
- **APPROVE** — all requirements implemented + tested, no violations
- **APPROVE WITH NOTES** — implemented, minor gaps documented for future
- **REVISE** — significant gaps, back to `sf-build` with specific corrections

**User can override verdict.** If the user disagrees with the assessment,
they can override. The override is logged in the review.

**Produces:**
- `specforge/features/<name>/review.md` — traceability matrix, gaps, verdict
- `specforge/archive/<date>-<name>/` — complete feature archive (on approve)
- Updated `specforge/history.md` — completion entry
- Updated `specforge/constitution.md` — if backprop promotes a new invariant

---

### sf-audit

Project-wide adversarial audit. Not part of the per-feature loop — it steps back
and reviews the whole project: constitution vs reality, cross-feature
consistency, drift, and accumulated gaps.

| | |
|---|---|
| **Triggers** | `sf-audit`, "audit the project", "is the constitution still true" |
| **Scope** | All features + constitution, not a single feature |
| **Produces** | `specforge/audits/<date>.md` — findings, severity, recommendations |

---

### sf-amend

Modify a feature that already shipped, without forking a parallel spec. Archived
specs are **living documents** — `sf-amend` runs a delta mini-pipeline
(propose-delta → gate → build → check) that edits the existing requirements,
design, tasks, and `trace.json` in place.

| | |
|---|---|
| **Triggers** | `sf-amend <feature>`, "change the shipped feature", "the archived spec is out of date" |
| **vs sf-propose** | amend changes an existing capability; propose creates a new one |
| **Edits in place** | one evolving traceability matrix per feature — never a second parallel spec |

Pair it with drift detection: `sf doctor --drift`
reads each archived feature's `trace.json` and flags requirements whose code
anchor vanished or whose test fails — so you find out the spec drifted before it
becomes a lie.

---

### Support skills

Standalone skills that complement the pipeline but are not part of it. They work
without `specforge/` initialized and produce artefacts in `specforge/context/`. See
[SUPPORT-SKILLS.md](../SUPPORT-SKILLS.md) for the full reference: `sfx-think`,
`sfx-triage`, `sfx-grill-me`, `sfx-tdd`, `sfx-documenter`, `sfx-explain`,
`sfx-aws-architect`, `sfx-data-engineer`, `sfx-github`, `sfx-journal`.

---

## Artefact flow

Each artifact below is a `.json` **source** + a `.md` **render** (`sf` generates
the Markdown from the JSON):

```
sf-init produces:
  specforge/constitution.json / .md       (principles with applies_to)
  specforge/context/project.md
  specforge/context/conventions.md

sf-propose produces (per feature):
  specforge/features/<name>/requirements.json / .md
  specforge/features/<name>/design.json / .md
  specforge/features/<name>/tasks.json / .md          (flat tasks + depends_on)

sf-build produces (per feature):
  specforge/features/<name>/progress/plan.json / .md  (sf plan compute — waves from deps)
  specforge/features/<name>/progress/wave-<n>.md
  specforge/features/<name>/audit.json                (if phase auditor ran)

sf-check produces (per feature):
  specforge/features/<name>/review.json / .md
  specforge/features/<name>/trace.json                (requirement → code → test anchors)
  specforge/journal/<date>-<name>.json / .md          (durable lessons, on archive)
  specforge/archive/<date>-<name>/                    (on APPROVE)
```

**Who reads what:**

| Artefact | Created by | Read by |
|----------|-----------|---------|
| constitution.md | sf-init | sf-propose (constraints), sf-check (compliance) |
| specforge/context/project.md | sf-init | sf-propose (stack context), sf-build (conventions) |
| specforge/context/conventions.md | sf-init | sf-build (coding standards) |
| requirements.json | sf-propose | sf-build (traceability), sf-check (validation) |
| design.json | sf-propose | sf-build (architecture guide), sf-check (adherence) |
| tasks.json | sf-propose | `sf plan compute` (deps → waves), sf-build (execution) |
| plan.json | `sf plan compute` | sf-build (`sf context for-wave`), sf state current |
| progress/*.md | sf-build | sf-check (audit trail) |
| trace.json | sf-check | `sf doctor --drift`, `sf trace verify` |
| audit.json | `sf gate record-verdict` | the human (phase-audit verdicts) |
| review.json | sf-check | archive (verdict determines archival) |
| journal/*.json | `sf journal add` | the phase auditor (past lessons), ICM (optional) |
| features/&lt;name&gt;/feature.json | `sf feature add` | all skills + `sf` (status gate, gate ledger) |
| history.md | sf-init | sf-check (backprop pattern tracking) |
