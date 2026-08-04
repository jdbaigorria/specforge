---
name: sf-init
description: >
  Initialize a SpecForge project. Works on both greenfield and brownfield codebases. Use when the
  user says "sf-init", "init specforge", "start a new project with specforge", "initialize SDD",
  or begins describing a new software project from scratch. Also triggers on brownfield signals
  like "add specforge to this project", "I have existing code and want to add specs", or
  "set up specforge here". Supports --from <path> to import an existing PRD or brief as input
  for the constitution, and --path <dir> to initialize in a specific subdirectory (useful for
  monorepos).
---

# sf-init

Initialize a SpecForge project. Detect context, scaffold structure, generate foundational artefacts.

## Step 1: Scaffold (automatic, no gate)

Create the directory structure. One visible root, `specforge/`, holds
everything; machine state hides in `specforge/.state/`:

```
specforge/
├── history.md              # Append-only project log
├── features/               # One folder per feature (each holds feature.json: status + gate ledger)
├── archive/                # Completed features
├── audits/                 # sf-audit reports
├── context/                # Project context + skill outputs (visible)
│   ├── project.md          # Stack, architecture (brownfield: inferred; greenfield: from constitution)
│   ├── conventions.md      # Code conventions (brownfield: inferred; greenfield: defined by user)
│   └── compact-rules.md    # Condensed rules for sub-agents
└── .state/                 # Hidden machine state (not human-edited)
    └── session.md          # Session cache / recovery (created on first checkpoint)
```

`constitution.md` and `roadmap.md` are added later (constitution in Step 3;
roadmap when `sf-propose --all` runs).

Feature state lives **per feature** in `specforge/features/<name>/feature.json`
(status, lane, gates) — created by `sf feature add`, never by hand. There is no
global registry file to initialize; `sf status` assembles the view on demand.

## Step 2: Detect project type

If `--path <dir>` was provided, scope all detection to that directory.

Otherwise, scan the working directory:
- If source code exists (`.py`, `.ts`, `.js`, `.go`, `src/`, `lib/`, `package.json`, etc.) → likely **brownfield**
- If empty or only config files → **greenfield**

### Ambiguous cases (monorepos, mixed workspaces)

If code is detected but the situation could be ambiguous (e.g., monorepo root with multiple
services, or a repo where the user might want to start a new sub-project), ask:

> "I see existing code in the workspace. Are you:
> 1. Adding SpecForge to this existing project
> 2. Starting a new sub-project within this workspace"

If (1) → brownfield. If (2) → ask for the sub-project path, scope detection to that directory,
and treat it as greenfield.

If detection is unambiguous (empty directory = greenfield, clear single-project structure with
package.json/pyproject.toml = brownfield), skip the question and inform:
"Detected [greenfield/brownfield] project."

## Step 3: Generate foundational artefacts

### Greenfield

1. **Constitution:** Read `references/constitution.md` for detailed instructions.
   Engage the user in a conversation about project vision (what, who, why),
   principles, constraints, and anti-goals.
   Generate `specforge/constitution.md` using `templates/constitution.tmpl.md`.
   → 🔴 **GATE**: Present to user. Wait for approval or changes.

2. **Project context:** From the constitution conversation, extract stack and
   architecture decisions. Generate `specforge/context/project.md` using `templates/project.tmpl.md`.
   Generate `specforge/context/conventions.md` with the conventions the user stated or agreed to.
   → 🔴 **GATE**: Present both documents. Wait for approval or changes.

If `--from <path>` was provided, read the file first and use it as primary input
for the constitution conversation. Pre-fill what you can extract from it, then ask
about what's missing (principles, anti-goals, constraints not covered in the document).

**From a discovery brief (greenfield, F27).** When `--from` points at an
`sfp-scout` discovery brief, it carries **product requirements with stable ids**
(`PR1`, `PR2`, …). Preserve those ids: record the `PR#` set so `sf-propose --all`
can derive the roadmap from them with traceable links (`serves: [PR#]`). This is
the top of the traceability spine — don't paraphrase the `PR#` away.

### Brownfield

1. **Onboard:** Read `references/onboard.md` for detailed instructions.
   Analyze the existing codebase: detect stack, frameworks, patterns, conventions.
   Generate `specforge/context/project.md` and `specforge/context/conventions.md` using templates.
   → 🔴 **GATE**: Present both documents. Wait for approval or changes.

2. **Constitution:** Read `references/constitution.md` for detailed instructions.
   The codebase tells you the *how* (stack, patterns) but not the *why* (principles,
   vision, anti-goals). Engage the user in the constitution conversation, using the
   inferred stack as context. The conversation can be shorter than greenfield because
   technical constraints are already known from onboard.
   Generate `specforge/constitution.md` using `templates/constitution.tmpl.md`.
   → 🔴 **GATE**: Present to user. Wait for approval or changes.

## Step 4: Confirm and log

Append to `specforge/history.md`:
```
## [date] — Project initialized
- Type: [greenfield/brownfield]
- Foundation: constitution.md + project.md + conventions.md
```

Inform the user:
"Project initialized. Use `sf-propose <feature-name>` to start specifying your first feature."
