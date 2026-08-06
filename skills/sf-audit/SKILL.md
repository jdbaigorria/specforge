---
name: sf-audit
description: >
  Run a project-wide adversarial audit. Cross-reference all features against the constitution,
  detect contradictions between features, find accumulated drift, verify invariants across the
  entire codebase. Use when the user says "sf-audit", "audit the project", "full review",
  "check everything against the constitution", "is the project consistent", "cross-check all
  features", "adversarial review", "project health check", or after a milestone when multiple
  features have been completed and it's time to verify the whole system holds together.
  Unlike sf-check (validates one feature), sf-audit validates the entire project.
---

# sf-audit

Adversarial audit of the entire project. Cross-reference everything. Find what nobody checked.

sf-check validates a feature against its own specs. sf-audit validates the project against
itself — constitution vs reality, feature vs feature, invariants vs codebase, specs vs code
across every completed feature.

**Always produces** `specforge/audits/{date}-{scope}.md`.

## When to Run

- After a milestone (3-5 features completed)
- Before a major release
- When the user feels "something is off" but can't pinpoint it
- Periodically as hygiene (every N features)
- After a major refactor to verify nothing drifted

## Step 1: Gather Everything

Read all of these (don't skip any that exist):

```
specforge/constitution.md          ← the law
sf status --json                   ← all features and statuses, already counted
specforge/features/*/               ← active features (requirements, design, tasks)
specforge/archive/*/                ← completed features
specforge/history.md               ← project log + recurring issues
specforge/context/project.md                     ← stack and architecture
specforge/context/conventions.md                 ← coding standards
```

Scan the actual codebase for the audit checks below.

## Step 2: Constitution Compliance (project-wide)

For each principle in `constitution.md`:

1. Scan the codebase for violations — not just the latest feature, ALL code
2. Check if any archived feature's design contradicts a principle
3. Check if any principle has become stale (the project evolved past it)
4. Check if invariants (promoted via backprop) are respected everywhere

Produce a compliance table:

```markdown
| Principle | Status | Violations | Notes |
|-----------|--------|------------|-------|
| "Offline-first" | ✅ PASS | 0 | |
| "No raw print()" | ⚠️ DRIFT | 3 files | Introduced in add-export feature |
| INV1: "Error handling on interfaces" | ❌ FAIL | 5 endpoints | 2 features predate this invariant |
```

## Step 3: Cross-Feature Consistency

Check for contradictions and conflicts across features:

### Data model consistency
- Do different features define the same entity differently?
- Are there conflicting schemas or type definitions?
- Do naming conventions stay consistent across features?

### Interface conflicts
- Do any features expose contradictory API contracts?
- Are there duplicate endpoints or CLI commands with different behavior?
- Do error codes/messages stay consistent?

### Dependency conflicts
- Do features depend on incompatible versions of the same library?
- Are there circular dependencies between feature modules?

### Scope overlap
- Do any features implement the same requirement differently?
- Are there redundant implementations nobody noticed?

## Step 4: Spec-to-Code Drift

Start from `sf doctor --drift --run-tests --json`, not from reading. The engine
classifies every finding into four categories, and they are not interchangeable:

| Category | What it means | What it asks of the reader |
|---|---|---|
| **not implemented** | the anchor doesn't resolve — the code is gone or was never written | is the requirement still wanted? |
| **implemented differently** | the anchor resolves and its test is red — the code does something else | **which side was wrong?** (see below) |
| **unverified** | the requirement names no test, or the test couldn't be run | it's a gap, not a contradiction |
| **out of spec** | code that no trace anchors | adopt it or exclude it (Step 6) |

Then read, for what the engine cannot see: does the *design* still match the
architecture? Did a post-approval edit change behavior the tests don't pin down?

### The two routes — never assume the spec is the one that's wrong

For every **implemented differently** finding, both readings are live:

> **(a)** the spec went out of date → accept the implementation, amend the spec.
> **(b)** the spec was right → this is a defect; the code has to change.

Present both, with no default. The pull toward (a) is structural — it makes the
finding disappear with one edit — and a project that always picks (a) has a spec
that records what happened instead of governing what should. Route (b) is
`sf delta new --feature=<name> --kind=code-wrong` with `expected` and `observed`.

**One at a time.** An audit that ends with "approve all 12" is 12 decisions
nobody made.

## Step 5: Convention Adherence

Read `specforge/context/conventions.md` and scan the codebase:

- Naming conventions consistent?
- Error handling patterns followed everywhere?
- Test structure consistent across all features?
- Import organization consistent?
- Documentation standards followed?

Flag areas where conventions drifted — often the earliest features don't
follow conventions that were established later.

## Step 6: Health Metrics — read them, don't estimate them

**The dividing line: if two runs over an unchanged repo can produce different
numbers, it isn't a metric — it's an opinion.**

Most of this table is already computed. Run the commands and transcribe. Counting
by reading is how an audit reports 7 drifted features on Monday and 9 on Friday
with nothing having changed in between — and once the numbers move on their own,
nobody trusts any of them.

```bash
sf status --json                            # features by status
sf doctor --drift --run-tests --json        # drift in 4 categories
sf coverage --json --by-priority            # anchored ratio + unanchored files + contract by priority
sf verify --json                            # invariants and integrity checks
```

Without `--run-tests`, `implemented_differently` comes back
`{"status":"undetermined"}`. Report it as undetermined. Do **not** write 0 —
that would state a fact nobody checked.

| Row | Where it comes from |
|---|---|
| Features completed / active | `sf status --json` → `statuses` |
| Spec-to-code drift | `sf doctor --drift --json`, broken out by the 4 categories |
| Not implemented / Implemented differently | same, `not_implemented` / `implemented_differently` |
| Unverified requirements | same, `unverified` |
| Out of spec (code with no requirement) | same, `out_of_spec` — and `sf coverage --json` → `unanchored` |
| Test coverage gaps | `sf coverage --json` → `percent`, `anchored`, `total` |
| Verification contract by priority | `sf coverage --json --by-priority` → `requirement_coverage.by_priority`. **Different unit** from the row above: that one counts *files*, this one counts *requirements*. Don't merge them — an audit that adds the two totals is adding apples to oranges. The row that matters is `must`: anything short of `total` there is release-blocking work the project declared and never verified |
| Constitution principles / Invariants | `constitution.json` (count them) |
| Integrity / invariant checks | `sf verify --json` → `checks` |
| **Cross-feature conflicts** | **Judgment** — two specs have to be read and found to contradict |
| **Convention violations** | **Judgment** — `conventions.md` is prose |

Mark the last two as judgment in the artifact. They're the rows where an
adversarial reader is the only instrument that works — which is precisely why
they shouldn't be competing for attention with nine numbers the CLI already
knows.

### Category 4: turn the list into decisions

`out_of_spec` / `unanchored` is code that no requirement governs. A count there
is useless; the list is actionable. For each file, exactly two outcomes — and
say which one you're recommending and why:

- **Adopt** — it's real product code and should be specified. Feeds a retroactive
  requirement (`sf-propose` on the existing code).
- **Exclude** — it isn't product (generated, tooling, scripts). Record it, with a
  reason, under `coverage.exclude` in `constitution.json`. Excluding *raises* the
  percentage, so an exclusion without a reason is how the metric gets quietly
  dressed up.

Never propose excluding a file just to move the number.

## Step 7: Produce Verdict

### HEALTHY
No critical violations. Minor drift noted. Project is consistent.

### NEEDS ATTENTION
Some principle violations or cross-feature inconsistencies.
Specific remediation steps listed.

### AT RISK
Significant drift from constitution. Multiple feature conflicts.
Accumulated debt threatens future development. Major remediation needed.

## Step 8: Write Artifact

Generate `specforge/audits/{date}-{scope}.md` using `templates/audit.tmpl.md`.

Alongside it, write `specforge/audits/{date}-{scope}.json` with the computed
rows verbatim — the raw output of the four commands from Step 6, plus the
verdict. Two audits of a markdown file can't be compared; two JSON files can.
That's what makes "drift went from 3 to 7 since May" a sentence anyone can
check.

```json
{
  "date": "…", "scope": "…", "verdict": "HEALTHY|NEEDS_ATTENTION|AT_RISK",
  "computed": {"status": {…}, "drift": {…}, "coverage": {…}, "verify": {…}},
  "judgment": {"cross_feature_conflicts": 0, "convention_violations": 0}
}
```

`computed` and `judgment` stay separated in the artifact for the same reason
they're separated in Step 6: one is reproducible and the other is a reading.
Collapsing them lends the numbers' authority to the opinions.

For the trend, `sf coverage --history` already keeps every measurement — read it
instead of reporting a lone point.

→ 🔴 **GATE**: Present audit results. User reviews findings.

For each finding, recommend:
- **Fix now:** violations that will compound if ignored
- **Fix next sprint:** issues that are real but not urgent
- **Track:** patterns that aren't violations yet but are trending that way
- **Accept:** deliberate deviations the user is aware of (document the decision)

## Step 9: Update History

Append to `specforge/history.md`:
```markdown
## [date] — Project audit
- Scope: {full | partial — what was checked}
- Verdict: {HEALTHY | NEEDS ATTENTION | AT RISK}
- Findings: {N critical}, {N major}, {N minor}
- Actions: {summary of recommended actions}
```

## Rules

- **Never estimate a number the CLI computes.** Every row in Step 6 marked as
  computed comes from a command. Two runs over an unchanged repo must produce
  identical computed rows — if they don't, the audit is fiction and its judgment
  rows inherit that.
- **Undetermined is not zero.** If `--run-tests` wasn't run, say the category
  wasn't evaluated. Writing 0 asserts something nobody checked.
- Be adversarial. Your job is to find problems, not confirm everything is fine.
- Be specific. "Code quality could improve" is useless. "3 API endpoints in
  add-export feature lack error handling, violating principle 4 and INV1" is useful.
- Check EVERYTHING against the constitution — including features that were approved
  before a principle or invariant existed.
- Don't audit what doesn't exist. If there's no constitution, say so and recommend
  running sf-init instead of inventing violations.
- Distinguish between violations (clear breach) and drift (gradual divergence).
  Violations need fixes. Drift needs decisions — maybe the constitution should update.
- If a principle is consistently violated, maybe it's the principle that's wrong.
  Flag it both ways: "either fix the code or amend the constitution."
