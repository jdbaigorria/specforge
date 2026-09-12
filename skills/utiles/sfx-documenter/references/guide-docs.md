# Guide / Tutorial Documentation

## Structure

```markdown
# How to: {accomplish a specific goal}

## Overview
What you'll build/learn. What you need before starting. Estimated time.

## Prerequisites
- {tool/library/access needed}
- {knowledge assumed}

## Steps

### Step 1: {action verb — "Create the config", "Set up the database"}

{Why this step matters — 1 sentence.}

```{language}
{actual code to run/write}
```

{Brief explanation of what the code does. Not line-by-line — highlight
the non-obvious parts only.}

### Step 2: {next action}

{Continue the pattern. Each step builds on the previous.}

### Step 3: {next action}

{If a step can go wrong, add a callout:}

> ⚠️ If you see `ErrorX`, it means {cause}. Fix: {solution}.

## Complete Example

{The full working code, assembled from all steps. Copy-paste ready.}

## What's Next
- {Natural next step or related guide}
- {Where to learn more}

## Troubleshooting
| Symptom | Cause | Fix |
|---------|-------|-----|
| {error} | {why} | {how to fix} |
```

## Principles

### 1. One goal per guide

"How to add authentication" is one guide. "How to add authentication and
set up the database and deploy to AWS" is three guides. Split them.

### 2. Show, don't tell

Wrong: "Create a configuration file with the appropriate settings."
Right: "Create `config.yml` with these contents:" + actual code block.

Every instruction that involves code must show the actual code. The reader
should never have to guess what you mean.

### 3. Working at every step

After each step, the code should be in a runnable state (or at least
compilable). Don't ask the reader to make 5 changes across 3 files
before they can verify anything works.

### 4. Explain the non-obvious

Don't explain what `import os` does. Do explain why you're using
`os.environ.get("KEY", "default")` instead of `os.environ["KEY"]` —
the fallback prevents crashes in development without the env var.

### 5. Anticipate failure

If step 3 commonly fails (wrong version, missing dependency, permissions),
add a callout BEFORE the reader hits the error, not after.

### 6. Complete example at the end

After the step-by-step, show the full assembled code. Some readers
skip the walkthrough and just want the final version. Serve both
audiences.

## Guide types

### How-to guide (task-oriented)
"How to add a new CLI command"
"How to configure the database connection"
"How to deploy to production"

Focused on accomplishing a specific task. Assumes the reader already
understands the concepts. Steps in order.

### Tutorial (learning-oriented)
"Building your first API endpoint"
"Understanding the plugin system"

Focused on teaching through building. Introduces concepts as they
become relevant. More explanation, more context.

### Architecture walkthrough
"How the authentication system works"
"Request lifecycle from HTTP to database"

Focused on understanding the system. Not step-by-step instructions
but a guided tour of how pieces connect. Code excerpts from the
actual codebase with annotations.

Choose the type based on what the user asked for. If unclear, default
to how-to guide — it's the most immediately useful.
