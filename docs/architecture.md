[← Back to README](../README.md)

## Architecture

### Directory structure

A single visible root, `specforge/`, holds everything. Review artefacts
(requirements, design, review) are visible so the human gates work; machine
state hides in `specforge/.state/`.

```
my-project/
├── specforge/                       # Single visible SpecForge root
│   ├── features.json                # Feature registry (status + gate ledger)
│   ├── constitution.json / .md      # Principles (with applies_to) + identity — JSON source + render
│   ├── history.md                   # Append-only project log
│   ├── roadmap.md                   # Feature roadmap
│   ├── learnings.md                 # Consolidated, evidence-anchored learnings (injected each session)
│   ├── features/
│   │   └── add-task-manager/
│   │       ├── requirements.json / .md   # JSON source + Markdown render (each artifact)
│   │       ├── design.json / .md
│   │       ├── tasks.json / .md          # flat tasks + depends_on
│   │       ├── trace.json                # requirement → code:symbol → test (drift anchors)
│   │       ├── review.json / .md
│   │       ├── audit.json                # phase-audit verdicts (quality tier)
│   │       ├── decisions/                # Complex technical decisions (optional)
│   │       └── progress/
│   │           ├── plan.json / .md       # computed wave layout (sf plan compute)
│   │           ├── wave-0.md
│   │           └── wave-1.md
│   ├── archive/
│   │   └── 2026-05-12-add-auth/
│   ├── audits/                      # sf-audit reports
│   ├── journal/                     # durable lessons (sf journal add) — git-tracked memory
│   ├── context/                     # Project context + skill outputs (visible)
│   │   ├── project.md               # Stack, architecture (brownfield inferred)
│   │   ├── conventions.md           # Code conventions
│   │   ├── compact-rules.md         # Condensed rules for sub-agents
│   │   └── thinks/ triages/ briefs/ grills/ …  # support-skill artefacts
│   └── .state/                      # Hidden machine state (not human-edited)
│       ├── session.md               # Session cache / recovery (not source of truth)
│       └── hook-context.json        # per-session slice-injection trigger state
│
└── src/                             # Your code
```

### Skill structure (progressive disclosure)

Each skill has a lean SKILL.md (<200 lines) and loads capabilities on demand
from `references/`. This keeps the context small — the agent only loads what
the current step needs.

```
sf-init/
├── SKILL.md
├── references/
│   ├── constitution.md              # Greenfield: conversation guide
│   └── onboard.md                   # Brownfield: codebase analysis
└── templates/
    ├── constitution.tmpl.md
    ├── project.tmpl.md
    └── conventions.tmpl.md

sf-propose/
├── SKILL.md
├── references/
│   ├── design-first.md              # Alt flow: architecture → requirements
│   ├── from-code.md                 # Brownfield: infer specs from code
│   ├── clarify.md                   # Post-generation refinement
│   ├── research.md                  # Technical decision documentation
│   └── ears-notation.md             # EARS syntax reference
└── templates/
    ├── requirements.tmpl.md
    ├── design.tmpl.md
    └── tasks.tmpl.md

sf-build/
├── SKILL.md
├── references/
│   ├── task-planning.md             # Run `sf plan compute` (waves from deps) + refine
│   └── wave-execution.md            # Per-wave execution + the minimal-code ladder
└── templates/
    └── progress.tmpl.md

sf-check/
├── SKILL.md
├── references/
│   ├── backprop.md                  # Pattern 3x → invariant promotion
│   ├── minimal-code.md              # Parsimony rubric (P-min audit)
│   └── archive.md                   # Sync + close procedure
└── templates/
    └── review.tmpl.md
```

### Execution model

The pipeline skills run **inline** by default — in the main conversation context
— because every artefact has a human gate that requires interaction, and a
sub-agent can't stop to ask for approval. The one exception is **build
execution**: once the plan gate is approved (the human authorizes it), the build
can run in a fresh sub-agent (`build.mode = single | per-wave`) — the plan is
already the contract, so execution needs no further gate until it reports back.
Build is where most tokens burn, so offloading it keeps the main context clean.
See [the deterministic layer](cli-and-hooks.md#the-deterministic-layer-cli-sf--hooks).

Support skills follow the same rule, decided by **interactivity, not tier**:

- **Gated or iterative** (`sfx-grill-me`, `sfx-tdd`) → inline.
- **Pure transform** — takes input, returns output, no human turn in the middle
  (`sfx-documenter`, `sfx-explain`, `sfx-aws-architect`, `sfx-data-engineer`, `sfx-triage`,
  `sfx-github`) → may be delegated to a sub-agent **where the harness supports it**,
  falling back to inline otherwise. Delegation is an optional, per-harness
  optimization, not part of the portable core.

The anti-telephone-game principle: artefacts live on disk. When a skill needs
context from a previous artefact, it reads the file — it doesn't rely on
conversation history. This means context window carries references, not payloads.
