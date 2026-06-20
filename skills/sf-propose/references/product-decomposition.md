# Product Decomposition (`sf-propose --all`)

Decompose the entire product into features, generate a roadmap, then propose each
feature one by one with gates.

Read `specforge/constitution.md` first — identity, principles, constraints,
anti-goals.

## Step A1: Decompose into features

Based on the product identity, break it down into logical features:
- Group by user-facing capability (not by technical layer)
- Each feature should be independently deliverable
- Name in kebab-case: `user-auth`, `task-crud`, `export-csv`

## Step A2: Prioritize

Classify each feature:

| Priority | Meaning |
|----------|---------|
| **Must** | Product doesn't work without it. MVP. |
| **Should** | Important but can launch without. v1.1. |
| **Could** | Nice to have. Backlog. |

## Step A3: Generate roadmap

Create `specforge/roadmap.md`:

```markdown
# Product Roadmap

Generated from constitution on [date].

## Must (MVP)
1. [feature-name] — [one-line description]
2. [feature-name] — [one-line description]

## Should (v1.1)
3. [feature-name] — [one-line description]

## Could (backlog)
4. [feature-name] — [one-line description]

## Dependencies
[feature-B depends on feature-A because...]
```

**Trace to product requirements (F27).** If the project was seeded from an
`sfp-scout` discovery brief, derive each roadmap feature from the product
requirements: tag it with which `PR#` it serves, e.g. `1. auth — login/logout
(serves: PR1, PR3)`. Record `"serves": ["PR1","PR3"]` on the feature in
`features.json`. This extends the traceability spine upward —
`PR# → feature → R# → task → code → test` — so product intent stays linked to
shipped code (no telephone game). Features with no discovery origin (brownfield)
just omit `serves`.

→ 🔴 **GATE**: Present roadmap. User approves, reorders, or adjusts.

## Step A4: Propose each feature

After roadmap approval, start proposing features in priority order.
For each: run the normal Requirements-First flow in `SKILL.md`.
Gate after each feature's 3 artefacts. User can stop at any point.
