---
name: sfp-roadmap
description: >
  Group the backlog's stories into features and put them in order — write `.docs/roadmap.json`.
  State `roadmap` (step ⑩) of the SpecForge machine, invoked when `sf next` returns
  `skill: sfp-roadmap`, running in a fresh subagent. Runs once for a new product, and again
  whenever `sf new` leaves stories that belong to no feature. Also usable standalone: "build the
  roadmap", "group the stories", "what order should these go in", "/sfp-roadmap". Adds no
  content — only ids and order.
---

# sfp-roadmap

The ⑩. **Groups and orders. Adds nothing.**

```bash
sf context      # → .docs/backlog/us-*.md    the stories to group
                # → .docs/constitucion.md    the project's rules
```

This is the file the machine runs on. `feature_actual` comes from here; without it `sf next`
does not know what comes next. **It is one of the three JSON files where `sf` counts**, which is
why it is JSON and not a document nobody parses.

## Step 1: Group by shared technical solution

The grouping criterion is not "by priority" — priority cannot tell you a group is *wrong*:

> **A feature is a set of stories that SHARE A TECHNICAL SOLUTION and ship together.**

That definition earns its keep because it can be violated. If three stories need three different
solutions, **they are not a feature and the grouping is wrong.** Split it.

Practical consequences:

- A bug and a new capability **never** share a feature. They do not share a solution, and the
  machine routes them differently (`tipo: bug` skips two states). Mixing them means the whole
  feature loses its planning.
- Six or more stories in one feature gets a warning from `sf`. It does not block — splitting is
  judgment and judgment is Javier's — but it is a real signal: a feature that big takes too long
  to come round, and its plan ages while it waits.

## Step 2: Order them, and the order is the only thing that expresses dependencies

```json
"orden": 1, 2, 3, …
```

That is all. **There is no `depends_on`, no dependency section, no MoSCoW.**

> **Why priority tiers are gone, at all three levels.** Must/Should/Could is a **negotiation
> tool** — it exists so a team can argue about scope with a PM or a client. **Here there is
> nobody to negotiate with:** Javier decides alone, at the ⑥, the ⑧ and the ⑰. Every question the
> tiers would answer is already answered:
>
> | Level | The question | Who already answers it |
> |---|---|---|
> | feature | do we build it, and when? | **`orden`**. A "could" is `orden: 7` |
> | story | is it in the MVP? | the `orden` of its feature |
> | criterion | can it fail and still ship? | **existing or not existing** (R5) |

If `us-5` needs `us-3`, put their features in that order. Done.

## Step 3: Give each feature an id, a slug and a name

```json
{"id": "f-1", "slug": "nucleo-cli", "nombre": "núcleo del CLI", "orden": 1,
 "historias": ["us-1", "us-3"]}
```

- **`id`** — `f-1`, `f-2`, … sequential and stable. It orders and it never changes.
- **`slug`** — lowercase, hyphenated, no accents. **It is load-bearing, not decoration:** the
  feature's folder is `.docs/features/{id}-{slug}/` and its branch comes from the same pair.
  Make it short and legible — someone will read it in an `ls` and in a branch name.
- **`nombre`** — for humans, in `sf status`. This one can have accents and spaces.
- **`historias`** — **ids only, never their text.** Referencing by id is what lets a story be
  rewritten without touching this file.

## Step 4: Every story lands somewhere

**A story in no feature will be implemented by nobody, and nobody finds out.** The gate checks
exactly this and names the ones left out.

This is also why the ⑩ runs again: `sf new` adds a story to a product that already has a
roadmap, and that story is an orphan until you place it. **You do not rewrite the roadmap** —
you place what is loose. Features already closed stay as they are.

## Step 5: Write it and close

Write `.docs/roadmap.json` from `templates/roadmap.tmpl.json`, then:

```bash
sf done
```

The gate: the file parses, and every `us-#` in the backlog belongs to some feature. Then the
feature cycle begins — `sf next` will tell the orchestrator to `sf take f-1`.

**No stop here.** Getting this wrong is cheap: fixing the order is changing a number, and the ⑰
already has a door for *"do another feature first"*. The expensive mistake was the ⑨, and that
one already had its pause.

## Rules

- Group by shared technical solution. If they need different solutions, it is not one feature.
- Never mix a bug with new capabilities.
- `orden` is the only expression of dependencies. No `depends_on`, no MoSCoW.
- ids only in `historias`. Never copy story text.
- The slug is a folder and a branch. Short, lowercase, no accents.
- Every story in some feature, or the gate stops you.
- Add no content. If you are writing prose, you are doing the ⑬'s job.
