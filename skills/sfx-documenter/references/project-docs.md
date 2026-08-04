# Project Documentation

## When to use

`/documenter --project` generates the full documentation suite for a project:
README, architecture overview, API reference for all modules, and starter guides.

## Output structure

```
docs/
├── README.md              ← project overview + quickstart
├── architecture.md        ← system design overview
├── api/
│   ├── index.md           ← module map with one-line descriptions
│   └── {module}.md        ← per-module reference (see api-docs.md)
└── guides/
    ├── getting-started.md ← first-time setup and hello world
    └── {topic}.md         ← how-to guides for common tasks
```

## README.md structure

```markdown
# {Project Name}

{One paragraph: what it is, who it's for, why it exists.}

## Quickstart

```bash
{3-5 commands to install and run the simplest useful thing}
```

## Features

- {feature 1} — {one line}
- {feature 2} — {one line}

## Installation

{Full installation instructions with prerequisites.}

## Usage

{The most common use case with a real example.}

## Documentation

- [Architecture](docs/architecture.md)
- [API Reference](docs/api/index.md)
- [Getting Started Guide](docs/guides/getting-started.md)

## Contributing

{How to set up development environment. How to run tests. How to submit changes.}

## License

{License type.}
```

## architecture.md structure

```markdown
# Architecture

## Overview

{2-3 sentences: what the system does and how it's organized.}

## System diagram

```
{ASCII or Mermaid: components and their connections}
```

## Components

### {Component 1}
- **Responsibility**: {what it does}
- **Location**: `{path/}`
- **Depends on**: {other components}
- **API**: [reference](api/{module}.md)

### {Component 2}
...

## Data flow

{How data moves through the system. Request lifecycle or pipeline stages.}

## Key decisions

| Decision | Choice | Why |
|----------|--------|-----|
| {what} | {choice} | {rationale} |
```

## Generation process

1. Read all source code to understand the project structure
2. Read `specforge/context/project.md` and `specforge/context/conventions.md` if available
3. Read existing tests to extract usage patterns
4. Read existing documentation to avoid contradictions
5. Generate README first (it frames everything else)
6. Generate architecture.md (system overview)
7. Generate api/ docs (one file per public module)
8. Generate guides/ (getting-started + common tasks identified from code)
9. Present to user for review → 🔴 GATE

## What NOT to document at project level

- Internal helper functions (unless they're part of the extension/plugin API)
- Test utilities (unless contributors need to use them)
- Generated code or vendored dependencies
- Configuration that's self-documenting (well-named env vars with comments)
