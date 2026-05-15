<p align="center">
  <img src="assets/specforge-support-logo.png" alt="SpecForge" width="380" />
</p>

# SpecForge — Support Skills

Standalone skills that complement the SpecForge pipeline. Each works independently — none require `specforge/` to be initialized. They produce artifacts in `.ai/` or interact conversationally.

---

## Skills Overview

| Skill | Mode | Purpose | Artifact |
|-------|------|---------|----------|
| [triage](#triage) | inline | Investigate bugs, find root cause, produce fix plan | `.ai/triages/{slug}.md` |
| [documenter](#documenter) | inline | Generate exhaustive docs from code with examples | `docs/` or inline |
| [explain](#explain) | inline | Teach concepts with Feynman method | `.ai/explanations/{slug}.md` (optional) |
| [product-owner](#product-owner) | inline | Define product briefs with MoSCoW priorities | `.ai/briefs/{slug}.md` |
| [aws-architect](#aws-architect) | delegate | Design AWS infrastructure with tradeoffs | `.ai/architectures/{slug}.md` |
| [data-engineer](#data-engineer) | delegate | Design data pipelines with quality gates | `.ai/data-designs/{slug}.md` |
| [grill-me](#grill-me) | inline | Stress-test a plan through relentless interviewing | `.ai/grills/{slug}.md` (optional) |
| [tdd](#tdd) | inline | Implement code with Red-Green-Refactor discipline | code + tests |
| [github](#github) | delegate | Execute git workflow: branch, commit, PR, merge | git state |

---

## triage

Investigate a bug systematically. No fixes without understanding the root cause first.

**Triggers:** `/triage`, `/triage <description>`, "there's a bug", "this is broken", "why is this failing"

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

**Output:** `.ai/triages/{slug}.md` — symptoms, investigation trace, hypotheses tested, root cause, fix plan with TDD test, risks.

**References:** `references/investigation.md` (methodology), `references/fix-plan.md` (TDD fix strategy)

**Next step after triage:** Apply fix directly (if trivial) or `sf-propose fix-{slug}` (if non-trivial)

---

## documenter

Generate exhaustive documentation from code. Every function gets an example. Every edge case gets documented.

**Triggers:** `/documenter`, "document this", "generate docs", "API docs", "how-to guide", "create a README for"

**Modes:**
- `/documenter <path>` — Document a specific file or module
- `/documenter --api <path>` — API reference (signatures, params, returns, examples)
- `/documenter --guide <topic>` — How-to guide (narrative, step-by-step)
- `/documenter --project` — Full project docs (README + architecture + API + guides)

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

## explain

Explain any concept using the Feynman method. Simple language, analogies from everyday life, concrete examples.

**Triggers:** `/explain`, `/explain <topic>`, "how does X work", "I don't understand", "teach me", "break it down"

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

**Output:** Conversational. Optionally saved to `.ai/explanations/{slug}.md` if user requests.

**References:** `references/feynman-method.md` (detailed teaching methodology with examples and anti-patterns)

---

## product-owner

Define product requirements. Translate vague ideas into testable user stories with MoSCoW priorities.

**Triggers:** `/product-owner`, `/brief`, "user story", "requirements", "what should we build", "define MVP"

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

**Output:** `.ai/briefs/{slug}.md` — problem, metric, persona, MoSCoW scope, user stories with acceptance criteria.

**Template:** `templates/brief.tmpl.md`

---

## aws-architect

Design AWS infrastructure evaluated through the Well-Architected Framework. Every service choice justified with tradeoffs and cost.

**Triggers:** `/aws-architect`, "design the infra", "how should I deploy", "what AWS services", "architecture for"

**Flow:**
1. Clarify requirements (workload, scale, budget, compliance, team)
2. Design with explicit tradeoffs per service (what, why not alternatives, cost, blast radius, scaling)
3. Generate architecture document

**Key rules:**
- Always state cost estimates — ranges, not "it depends"
- Simplest architecture first — add complexity only when justified
- Never recommend a service without explaining why not a simpler alternative
- Security is never optional — IAM, encryption, network isolation always included

**Output:** `.ai/architectures/{slug}.md` — requirements, services with rationale, data flow, security, scaling, cost breakdown, risks, decisions log.

**Template:** `templates/architecture.tmpl.md`

---

## data-engineer

Design data pipelines with quality gates, idempotency, and observability built in.

**Triggers:** `/data-engineer`, "data pipeline", "ETL", "data model", "schema design", "data quality"

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

**Output:** `.ai/data-designs/{slug}.md` — data contract, schema, stages, quality gates, idempotency, backfill, observability, storage design.

**Template:** `templates/pipeline.tmpl.md`

---

## grill-me

Stress-test a plan, design, or decision through relentless interviewing. Walk down every branch of the decision tree.

**Triggers:** `/grill-me`, `/grill-me <artifact>`, "grill me", "stress-test this plan", "poke holes in this"

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

**Output:** Summary always returned. Full transcript optionally saved to `.ai/grills/{slug}.md`.

---

## tdd

Implement code using strict Test-Driven Development. One failing test, one minimal implementation, one cycle.

**Triggers:** `/tdd`, `/tdd <task>`, "tdd this", "use TDD", "red-green-refactor"

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

## github

Execute standardized git workflow. Branch, commit, PR, merge — clean and quiet.

**Triggers:** `/github`, `/git`, "create a PR", "push this", "commit", "merge"

**Conventions:**
- Branches: `feature/`, `fix/`, `refactor/`, `docs/`, `chore/`
- Commits: conventional format — `type(scope): description`
- All changes via PR. Squash merge. Delete branch after merge.

**Key rules:**
- Never commit directly to main
- If CI fails, report — never auto-merge
- One PR = one logical change

**Output:** Git state (branches, commits, PRs). Returns clean summary only — no command output noise.
