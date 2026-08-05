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

**This is not the same axis as a requirement's `priority`.** Both use the MoSCoW
words, so they read as one thing and they are not:

| | Question it answers | What it controls |
|---|---|---|
| Feature priority (here) | Do we build this, and when? | The order of the roadmap |
| Requirement `priority` | Within a feature we *are* building, what can hold up a release? | How hard the verdict gate bites (`verification.blocking_priorities`) |

**Do not inherit one from the other.** A `could` feature you decided to build
still has requirements that are `must` *for that feature* — "we chose to build
it, so it has to work". Copying `could` down would silently ship it ungated,
which is the opposite of what classifying it as backlog meant. The reverse is
just as wrong: a `must` feature routinely contains `could` polish.

Ask per requirement, as `SKILL.md` says. The roadmap priority is context for
that question, never the answer to it.

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

After roadmap approval, propose features in **priority order — except that
dependencies win**. A feature cannot be proposed before the ones it depends on,
whatever its priority. Record the edges with `--depends-on` when you create the
feature so the order is data, not something you re-derive each session:

```bash
sf feature add --feature=checkout --depends-on=user-auth
```

**When a `must` depends on a `could`, stop and say so.** Don't quietly pull the
`could` forward — the ordering is the symptom, not the problem. Either the
dependency is real and that feature was mis-classified (it's load-bearing for the
MVP, so it isn't backlog), or the dependency is accidental and the `must` should
not need it. Both readings change the roadmap, and only the user can pick.

For each feature: run the normal Requirements-First flow in `SKILL.md`.
Gate after each feature's 3 artefacts. User can stop at any point.

## Step A5: Check the roadmap against reality

Once features start closing, the classification stops being a plan and becomes a
claim you can check:

```bash
sf coverage --by-priority
```

It reports, per requirement priority, how many have a satisfied verification
contract — and which ones don't, by name. The unit is the **requirement**, not
the feature and not the file, so read it against Step A2's second table.

The number to look at is `must`: anything under 100% there is a requirement that
the project itself declared release-blocking and has not verified. `should` and
`could` under 100% may be entirely fine — that's what classifying them was for.
