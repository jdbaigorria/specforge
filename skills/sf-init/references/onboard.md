# Onboarding (Brownfield)

## Purpose

Analyze an existing codebase to generate context documents that inform
subsequent propose/build/check phases. The agent should understand the
project as well as a new team member after a thorough onboarding.

## Analysis Steps

### 1. Project Structure
- Map the directory tree (ignore node_modules, .git, __pycache__, venv)
- Identify entry points (main.py, index.ts, app.py, etc.)
- Identify test structure (tests/, __tests__/, spec/)
- Note config files (package.json, pyproject.toml, serverless.yml, CDK, etc.)

### 2. Stack Detection
- **Language:** primary language + version
- **Framework:** web framework, CLI framework, etc.
- **Infrastructure:** AWS, GCP, Azure; specific services
- **Database:** type + ORM/client
- **Testing:** framework + coverage tool
- **Build/Deploy:** CI/CD, IaC, containerization

### 3. Code Conventions
- **Naming:** snake_case, camelCase, PascalCase — for variables, functions, classes, files
- **Documentation:** docstring style (google, numpy, sphinx), inline comments frequency
- **Error handling:** pattern (try/except, Result types, error codes)
- **Imports:** organization (stdlib → third-party → local)
- **Architecture patterns:** layered, hexagonal, MVC, serverless handlers

### 4. Domain Model (if applicable)
- Key entities and their relationships
- Business rules encoded in code
- Domain-specific vocabulary

## Output: .ai/project.md

```markdown
# Project Context

## Overview
**Name:** [from package.json/pyproject.toml or folder name]
**Language:** [lang + version]
**Framework:** [primary framework]
**Type:** [CLI / API / Web App / Library / Service]

## Stack
- **Runtime:** [e.g., Python 3.12, Node 20]
- **Framework:** [e.g., FastAPI, Typer, Next.js]
- **Infrastructure:** [e.g., AWS Lambda + API Gateway + DynamoDB]
- **Database:** [e.g., DynamoDB single-table, PostgreSQL via SQLAlchemy]
- **Testing:** [e.g., pytest + coverage.py]
- **Deploy:** [e.g., Serverless Framework, CDK, GitHub Actions]

## Architecture
[Brief description of how the code is organized — layers, modules, patterns]

## Key Files
- `[path]` — [what it does]
- `[path]` — [what it does]

## Domain
[Key entities, business rules, vocabulary if applicable]
```

## Output: .ai/conventions.md

```markdown
# Code Conventions

## Naming
- Variables: [snake_case]
- Functions: [snake_case]
- Classes: [PascalCase]
- Files: [kebab-case / snake_case]
- Constants: [UPPER_SNAKE_CASE]

## Documentation
- Style: [google docstrings]
- When required: [public functions, classes, modules]

## Error Handling
- Pattern: [custom exceptions inheriting from base, try/except at boundaries]

## Imports
- Order: [stdlib → third-party → local, one blank line between groups]

## Testing
- File naming: `test_[module].py`
- Structure: [Arrange/Act/Assert]
- Coverage target: [if identifiable]

## Git
- Branch naming: [if identifiable from .git]
- Commit style: [if identifiable]
```

## Important

- Don't guess. If a convention isn't clearly established in the code, say "not established" 
  rather than inferring from a single occurrence.
- Read at least 5-10 representative files before concluding conventions.
- If the project has a CONTRIBUTING.md or style guide, prioritize that over inference.
