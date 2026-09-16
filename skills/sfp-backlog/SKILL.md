---
name: sfp-backlog
description: >
  Cut the PRD into user stories with stable IDs and testable acceptance criteria — the `us-#.md`
  files that everything downstream references. State `backlog` (step ⑨) of the SpecForge machine,
  invoked when `sf next` returns `skill: sfp-backlog`, running in a fresh subagent. It composes
  sfx-criterio for the criteria. Also usable standalone: "cut this into stories", "write the
  backlog", "split the PRD", "/sfp-backlog". Produces stories only — grouping and ordering are
  the ⑩'s job, and work that arrives LATER does not come through here: that is an entry node
  (`sf-entrar-feature`, `sf-entrar-bug`).
---

# sfp-backlog

The ⑨. **PRD → user stories.** The first output that is not for reading but for **working**, and
the first time stable IDs appear.

```bash
sf context      # → .docs/prd.md            what to cut up
                # → .docs/constitucion.md   the project's rules (context, not source)
```

**The source is the PRD.** The constitution is context — it tells you what the project is like,
not what it has to do.

**You produce stories, not features.** Grouping and ordering happen in the ⑩. Keeping *what has
to be possible* apart from *in what order and grouped how* is what lets the roadmap be a view of
the backlog rather than a second backlog.

## Why the criteria are the whole point

```
sf does not verify the judgment. It verifies THAT THE JUDGMENT HAPPENED, over all of them.
     "the story has 3 criteria, the report talks about 2"
```

With criteria in loose prose, a reviewer answers *"works"* and half a story ships as done. With
`CA-1 / CA-2 / CA-3` they have to answer **one by one**, and `sf` counts. **Counting is not
judging** — it is the only part of this problem a machine can do, and it is enough.

Two states depend on the IDs you write here: the ⑯ counts whether every criterion has a task,
and the ㉑ counts whether every criterion got a verdict. **A story without IDs is invisible to
both.**

## Step 1: One story per thing someone wants to be able to do

Cut the PRD's capabilities into stories. Each one:

```markdown
Como **<rol>** quiero **<qué>** para **<por qué>**.
```

The role comes from the PRD's actor list — do not invent a new cast.

**Size heuristic:** a story that cannot be finished inside one feature is two stories. A story
with one criterion is usually a criterion of another story.

## Step 2: The criteria — compose, do not re-explain

`Call the Skill tool with "sfx-criterio"`, telling it the stories you just cut and that the
criteria go in each `us-#.md` as `- **CA-1** — …`.

**It owns the method** — EARS, the test-name check, and the six-defect rubric. Do not restate the
rule here and do not run your own quality pass: the skill that writes a criterion and the skills
that have to turn it into a test (`sf-plan` ⑭⑮) and walk it with a verdict (`sf-check` ㉑) are
three different skills, and the rule that binds them has one owner so the other two are not
improvising.

What comes back is criteria with ids, each with a named test, and a list of what was moved to
*"what does NOT get in"* with the reason.

## Step 3: Work that arrives later does NOT come through here

When the product already exists and something new arrives — a capability, a bug — it enters
through its own door: `sf-entrar-feature` or `sf-entrar-bug`. Those understand it, decide whether
it gets in, and write `tipo` themselves, because **how much process something needs is the output
of deciding, not a checkbox somebody fills in afterwards.**

> **This used to live here, and that was the bug.** Measured 2026-09-13: the machine routes on
> `tipo: chico`, `sfx-mapa` explains it and the story skeleton offers it — and this skill, the one
> that actually writes the frontmatter, only ever knew `us` and `bug`. **A field the machine
> enforces and no skill writes.** Whoever decided how much process it needs is the one who can
> write it.

This skill is now only the ⑨: the PRD, cut into stories.

## Step 4: What does NOT go in the story

The `us-#.md` stays **100% human: the text and the criteria.** Everything else was removed
because it can be derived, and what is derived cannot go stale:

```
estado:      pendiente · planificada · en-curso · hecha · archivada
             → all five come from roadmap.json + estado.json, and `sf status` shows them
prioridad:   → the feature's `orden` in the roadmap says it
lane / modo: → the machine skips states; there is no second lane
```

> If a model forgets to update a frontmatter field when the feature closes, the file lies
> forever and nobody notices. **What is computed cannot go stale.**

## Step 5: Write them, then the ⏸

Write `.docs/backlog/us-<n>.md` — one file per story, from `templates/us.tmpl.md` — then:

```bash
sf done
```

The gate counts: **at least one story, and every story has criteria with IDs.** If one is
missing, it names it.

Then `sf` prints a **⏸ cheap pause**: Javier looks if he wants and moves on with `sf approve`.
It is not a 🛑 — nothing here is one of his three decisions — but it is where a bad cut is
cheapest to fix. **A wrong story is expensive; a wrong order is one number.**

## Rules

- Source is the PRD. The constitution is context.
- Stories, never features. Grouping is the ⑩.
- Every story has at least one `CA-#`. `sf` will not move without them.
- The criteria are `sfx-criterio`'s. Compose it; never restate EARS or the rubric here.
- The story is text and criteria. No state, no priority, no lane.
- Work that arrives later is an entry node, not this skill.
