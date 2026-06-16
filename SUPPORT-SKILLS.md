<p align="center">
  <img src="assets/specforge-support-logo.png" alt="SpecForge" width="380" />
</p>

# SpecForge — Support Skills

Standalone skills that complement the SpecForge pipeline. Each works independently — none require `specforge/` to be initialized. They produce artifacts in `specforge/context/` or interact conversationally.

---

## Skills Overview

**Mode is decided by interactivity, not by tier.** A sub-agent runs in an
isolated context and cannot stop to ask the user, so any skill with a gate or
back-and-forth must run **inline**. Skills that are pure transforms (take input,
return output, no human turn in the middle) are marked **delegate**: they *may*
be delegated to a sub-agent where the harness supports it, and fall back to
inline otherwise. Delegation is an optional, per-harness optimization — the
portable default is always inline.

| Skill | Mode | Purpose | Artifact |
|-------|------|---------|----------|
| [sfx-triage](#sfx-triage) | delegate | Investigate bugs, find root cause, produce fix plan | `specforge/context/triages/{slug}.md` |
| [sfx-documenter](#sfx-documenter) | delegate | Generate exhaustive docs from code with examples | `docs/` or inline |
| [sfx-explain](#sfx-explain) | delegate | Teach concepts with Feynman method | `specforge/context/explanations/{slug}.md` (optional) |
| [sfx-product-owner](#sfx-product-owner) | inline | Define product briefs with MoSCoW priorities | `specforge/context/briefs/{slug}.md` |
| [sfx-aws-architect](#sfx-aws-architect) | delegate | Design AWS infrastructure with tradeoffs | `specforge/context/architectures/{slug}.md` |
| [sfx-data-engineer](#sfx-data-engineer) | delegate | Design data pipelines with quality gates | `specforge/context/data-designs/{slug}.md` |
| [sfx-grill-me](#sfx-grill-me) | inline | Stress-test a plan through relentless interviewing | `specforge/context/grills/{slug}.md` (optional) |
| [sfx-tdd](#sfx-tdd) | inline | Implement code with Red-Green-Refactor discipline | code + tests |
| [sfx-github](#sfx-github) | delegate | Execute git workflow: branch, commit, PR, merge | git state |

---

## sfx-triage

Investigate a bug systematically. No fixes without understanding the root cause first.

**Triggers:** `/sfx-triage`, `/sfx-triage <description>`, "there's a bug", "this is broken", "why is this failing"

**Flow:**
1. Gather symptoms (observed vs expected, reproduction steps, frequency)
2. Explore codebase (trace data flow backward from symptom)
3. Form 2-3 hypotheses ranked by likelihood, each with evidence
4. Verify hypotheses (circuit breaker: 3 max — if all rejected, mark unresolved)
5. Write artifact with diagnosis, fix plan, and failing test case

**Key rules:**
- Never apply a fix during triage — investigation only
- Every hypothesis needs evidence, not guesses
- Root cause found → write the failing test BEFORE the fix recommendation
- Unresolved is a valid outcome — document what was tested and what remains

**Output:** `specforge/context/triages/{slug}.md` — symptoms, investigation trace, hypotheses tested, root cause, fix plan with TDD test, risks.

**References:** `references/investigation.md` (methodology), `references/fix-plan.md` (TDD fix strategy)

**Next step after triage:** Apply fix directly (if trivial) or `sf-propose fix-{slug}` (if non-trivial)

---

## sfx-documenter

Generate exhaustive documentation from code. Every function gets an example. Every edge case gets documented.

**Triggers:** `/sfx-documenter`, "document this", "generate docs", "API docs", "how-to guide", "create a README for"

**Modes:**
- `/sfx-documenter <path>` — Document a specific file or module
- `/sfx-documenter --api <path>` — API reference (signatures, params, returns, examples)
- `/sfx-documenter --guide <topic>` — How-to guide (narrative, step-by-step)
- `/sfx-documenter --project` — Full project docs (README + architecture + API + guides)

**Key rules:**
- Every public function/class/endpoint gets a runnable example — no exceptions
- Extract examples from tests when possible — they're already verified
- Document edge cases, error conditions, and side effects — not just happy paths
- Omit sections that would say "N/A" — no noise
- Match the project's code style in examples

**Output:**
- Single file → inline or `docs/{module}.md`
- Project → `docs/` directory (README, api/, guides/, architecture.md)

**References:** `references/api-docs.md` (reference format), `references/guide-docs.md` (tutorial format), `references/project-docs.md` (full project structure)

---

## sfx-explain

Explain any concept using the Feynman method. Simple language, analogies from everyday life, concrete examples.

**Triggers:** `/sfx-explain`, `/sfx-explain <topic>`, "how does X work", "I don't understand", "teach me", "break it down"

**Flow:**
1. Core idea in one sentence (no jargon)
2. Analogy from everyday life (not from other technical concepts)
3. Build up layer by layer (what → why → how → when)
4. Concrete example (real, runnable code in the user's stack)
5. Tradeoffs (what you gain, what you lose, when it backfires)
6. One-line summary

**Key rules:**
- No jargon without defining it immediately
- Analogies from everyday life — "like a shipping container" not "like a VM"
- Always name tradeoffs — "no downsides" means you haven't understood it
- Code examples must be runnable, not pseudocode

**Output:** Conversational. Optionally saved to `specforge/context/explanations/{slug}.md` if user requests.

**References:** `references/feynman-method.md` (detailed teaching methodology with examples and anti-patterns)

---

## sfx-product-owner

Define product requirements. Translate vague ideas into testable user stories with MoSCoW priorities.

**Triggers:** `/sfx-product-owner`, `/brief`, "user story", "requirements", "what should we build", "define MVP"

**Flags:** `--from <path>` to import an existing PRD or document as primary input

**Flow:**
1. Understand the problem (who, what, why, how to measure)
2. Scope with MoSCoW (Must 3-5 items, Should, Could, Won't)
3. Write user stories with GIVEN/WHEN/THEN acceptance criteria
4. Generate brief

**Key rules:**
- Start with the problem, not the solution
- Metrics must be measurable numbers
- Always include "Won't" items — explicit exclusion prevents scope creep
- With --from: extract what exists, ask only what's missing

**Output:** `specforge/context/briefs/{slug}.md` — problem, metric, persona, MoSCoW scope, user stories with acceptance criteria.

**Template:** `templates/brief.tmpl.md`

---

## sfx-aws-architect

Design AWS infrastructure evaluated through the Well-Architected Framework. Every service choice justified with tradeoffs and cost.

**Triggers:** `/sfx-aws-architect`, "design the infra", "how should I deploy", "what AWS services", "architecture for"

**Flow:**
1. Clarify requirements (workload, scale, budget, compliance, team)
2. Design with explicit tradeoffs per service (what, why not alternatives, cost, blast radius, scaling)
3. Generate architecture document

**Key rules:**
- Always state cost estimates — ranges, not "it depends"
- Simplest architecture first — add complexity only when justified
- Never recommend a service without explaining why not a simpler alternative
- Security is never optional — IAM, encryption, network isolation always included

**Output:** `specforge/context/architectures/{slug}.md` — requirements, services with rationale, data flow, security, scaling, cost breakdown, risks, decisions log.

**Template:** `templates/architecture.tmpl.md`

---

## sfx-data-engineer

Design data pipelines with quality gates, idempotency, and observability built in.

**Triggers:** `/sfx-data-engineer`, "data pipeline", "ETL", "data model", "schema design", "data quality"

**Flow:**
1. Understand the data (source, destination, transformations, quality, freshness)
2. Design pipeline stages (extract → validate → transform → load → verify)
3. Generate pipeline document

**Key rules:**
- Every pipeline must be idempotent and re-runnable
- Quality gate before loading — bad data never reaches consumers
- Error handling defined per stage, not just "it'll fail"
- Observability is not optional — metrics, alerts, lineage always included
- If volume doesn't justify streaming, use batch

**Output:** `specforge/context/data-designs/{slug}.md` — data contract, schema, stages, quality gates, idempotency, backfill, observability, storage design.

**Template:** `templates/pipeline.tmpl.md`

---

## sfx-grill-me

Stress-test a plan, design, or decision through relentless interviewing. Walk down every branch of the decision tree.

**Triggers:** `/sfx-grill-me`, `/sfx-grill-me <artifact>`, "grill me", "stress-test this plan", "poke holes in this"

**Flow:**
1. Read the target (artifact, topic, or pasted plan)
2. Ask save preference before starting
3. Explore context (code, specs, project docs) to avoid asking answerable questions
4. Interview one question at a time, each with a recommended answer
5. Track decisions made and open issues
6. Generate summary

**Key rules:**
- One question at a time — never batch
- Always provide your recommended answer (user can accept with "yes" or push back)
- If answerable by reading code, read it instead of asking
- 3 consecutive "I don't know" → pause, suggest research first
- Never lecture — extract the user's thinking, don't teach

**Output:** Summary always returned. Full transcript optionally saved to `specforge/context/grills/{slug}.md`.

---

## sfx-tdd

Implement code using strict Test-Driven Development. One failing test, one minimal implementation, one cycle.

**Triggers:** `/sfx-tdd`, `/sfx-tdd <task>`, "tdd this", "use TDD", "red-green-refactor"

**Flow:**
1. Plan: identify interface changes, list 3-7 behaviors to test, check testability
2. For each behavior: RED (one failing test) → GREEN (minimal implementation) → REFACTOR (only when all green)
3. Track evidence table (behavior → test file → impl file → refactor)

**Key rules:**
- ONE test at a time — vertical slicing, never horizontal
- Test MUST fail before implementation — if it passes, something's wrong
- Tests verify behavior through public interfaces, not implementation details
- No speculative code — if no test demands it, don't write it
- Never refactor while RED

**Output:** Code + tests + TDD cycle evidence table.

---

## sfx-github

Execute standardized git workflow. Branch, commit, PR, merge — clean and quiet.

**Triggers:** `/sfx-github`, `/git`, "create a PR", "push this", "commit", "merge"

**Conventions:**
- Branches: `feature/`, `fix/`, `refactor/`, `docs/`, `chore/`
- Commits: conventional format — `type(scope): description`
- All changes via PR. Squash merge. Delete branch after merge.

**Key rules:**
- Never commit directly to main
- If CI fails, report — never auto-merge
- One PR = one logical change

**Output:** Git state (branches, commits, PRs). Returns clean summary only — no command output noise.
