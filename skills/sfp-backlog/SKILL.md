---
name: sfp-backlog
description: >
  Cut the PRD into user stories with stable IDs and testable acceptance criteria — the `us-#.md`
  files that everything downstream references. State `backlog` (step ⑨) of the SpecForge machine,
  invoked when `sf next` returns `skill: sfp-backlog`, running in a fresh subagent. Also handles
  entries that arrive later through `sf new`: a new capability, or a bug (which composes
  sfx-triage). Also usable standalone: "cut this into stories", "write the backlog", "split the
  PRD", "/sfp-backlog". Produces stories only — grouping and ordering are the ⑩'s job.
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

## Step 2: Write the criteria in EARS, and this is not decoration

Read `references/ears-notation.md`. The reason EARS survives when everything around it was cut:

```
a test needs   →   trigger · state · expected behaviour
EARS asks for  →   trigger · state · behaviour
```

**It is the same triple.** That is why a criterion written in EARS converts into a test at the
⑮ without inventing anything. The failure mode is concrete:

```
CA-1  acepta --json y devuelve el estado serializado     ← ubiquitous ✔
CA-2  si no hay estado, sale con código 1 y mensaje      ← IF…THEN ✔
CA-3  --json es incompatible con --verbose               ← ⚠ and then what?
```

The third one does not say what happens, **so no test can be written against it.** EARS would
have forced it.

**Format matters, because `sf` counts these with a pattern match:**

```markdown
## Criterios de aceptación
- **CA-1** — <criterion>
- **CA-2** — <criterion>
```

## Step 3: Run the quality rubric over your own criteria

Read `references/requirement-quality.md` and apply its six defects to what you just wrote.
**`sf` cannot catch a single one of them** — ambiguity, two-criteria-in-one, unverifiable — they
are all judgment (R3). This rubric is the only thing between a vague criterion and a feature
that ships half-done.

The sharpest one is R5 in rubric form: **if you cannot write the test, it is not a criterion.**
"The UI should feel responsive" is not one. Turn it into something observable or drop it — and
if you drop it, say so, do not leave it as prose nobody will ever check.

**There is no priority field on a criterion.** A criterion either exists or it does not. There
is no "nice to have" tier — that is what deleting is for.

## Step 4: Entries that arrive later (`sf new`)

When the product already exists, new work enters here — **the backlog is the funnel**, all three
entries converge on it. `sf new` has already created the skeleton with Javier's text in
`## Contexto`. You fill in the title, the story and the criteria.

- **A new capability** — same as above.
- **A bug** — set `tipo: bug` and **compose `sfx-triage`**: find the root cause first, then write
  the criteria against the *actual* defect. Fill `relacionado_a` with the original `us-#`.

> **`tipo: bug` is what routes it.** A bug skips `planificacion` and `revision` and goes straight
> `implementar → cierre`. There is no parallel lane and no second machine — one field that skips
> two states, which is why the trail is never lost.

**A bug still needs criteria.** "It should not crash" is not one; "given input X the command
exits 0 and prints Y" is.

## Step 5: What does NOT go in the story

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

## Step 6: Write them, then the ⏸

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
- EARS, because a criterion that is not a trigger/state/behaviour cannot become a test.
- Untestable → not a criterion. Delete it or rewrite it.
- The story is text and criteria. No state, no priority, no lane.
- A bug composes `sfx-triage` and carries `relacionado_a`.
