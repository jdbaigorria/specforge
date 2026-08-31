---
name: sfp-constitucion
description: >
  Write the project's technical constitution — architecture, stack and why, code conventions,
  folder layout, working rules — and fill the frontmatter that `sf` reads. State `constitucion`
  (step ⑧) of the SpecForge machine, invoked when `sf next` returns `skill: sfp-constitucion` —
  and, unlike every other state, it runs WITH Javier and not in a fresh subagent: besides the
  brief it is the only step that talks to him, because architecture, stack and conventions do
  not derive from the PRD. This is the hinge artifact: its readers are a cold subagent, the
  implementer, the reviewer and `sf` itself, and its frontmatter is where the branch-per-feature
  and unapproved-dependency pains die. Also usable standalone: "write the constitution", "set the project rules",
  "/sfp-constitucion". Purely technical — the vision and the why already live in the brief and
  the PRD.
---

# sfp-constitucion

The ⑧, and the hinge of the whole flow.

```bash
sf context      # → .docs/prd.md            what has to be built
                # → .docs/constitucion.md   the technical header, ALREADY FILLED
```

> **The first artifact whose main reader is not a human and not this conversation: it is a cold
> subagent.** The ⑨ reads it, the ⑲ implementer reads it as their manual, the ㉑ reviewer reads
> it, and `sf` parses its header. Everything you write here gets read by someone who was never
> in this room.

## What this is NOT

**Zero abstract principles.** No identity section, no "what problem does this solve", no
anti-goals, no vision.

> **The reason is structural, not stylistic.** An older version of this artifact carried the
> *why*, because back then nothing came before it. Now it arrives **third** — the what and the
> why are already written and sealed above it, in the brief and the PRD. Repeating them here
> creates a second copy that drifts. This one is purely technical.

It is also **one file, not three.** Architecture, stack and conventions used to live in
`constitution.md` + `context/project.md` + `context/conventions.md`. Three files to maintain,
three that age, three that the envelope would have to serve to answer one question.

## Step 0: The header is already filled — check it, don't rewrite it

`sf init` ran `detectStack()` and wrote `lenguaje`, `manifiesto` and `test_cmd`. **Read them and
verify they are right**, then leave them alone. What you own is the body plus the fields that
need judgment (`mutacion`, `git`, `dependencias_aprobadas`).

**`test_cmd` is the one line that cannot be empty.** `sf` will not seal the state without it,
and it is not pedantry: three gates hang off being able to run the tests — the red of the ⑲, the
green of the ⑳, and the count that kills "all green" with no tests. If `detectStack()` guessed
wrong, fix it now.

### Then ask the two questions that `detectStack()` cannot answer

Both of these come from a real run where the loop stalled, and both are one line each.

**`test_requiere` — what does `test_cmd` need in order to run at all?** A suite that talks to
Postgres, spins a container, or needs Redis is not the same as one that runs on a bare machine.
`sf` will not provide it — it does not start containers, and it should not — but the ⑫ reads this
before cutting batches, and a planner who knows there is no database will not plan a batch that
needs one. Empty is the normal answer; do not fill it because the field exists.

**And whatever the runner needs that nobody would guess** goes in the body, in the conventions
section. The real case: a `vitest` project on `environment: jsdom` where any test touching
`node:fs` must be marked `// @vitest-environment node`. That was already the house style and
nobody had written it down — so the implementer met it as a failure instead of as a rule.

> If the runner needs a pragma, an environment, a build tag or an env var to run a given kind of
> test, **it goes here.** The implementer arrives cold and will not guess it; they will fight it,
> and the fight will look like a broken test.

## Step 1: Interview Javier — this state talks to him directly

> You can actually do this: `sf next` gives the ⑧ `via: vos`, so you are in the conversation
> with him, not in a subagent. Do not "optimise" this away by deciding everything yourself and
> writing — the ⑧ used to be dispatched as a subagent by mistake, and the models that papered
> over it by skipping this step are the reason nobody noticed for months.

Cover, in this order, and **skip nothing silently**:

1. **Architecture** — the shape of the thing. Layers, boundaries, what talks to what.
2. **Stack, and why** — every choice with the reason next to it. "Because it's popular" is not a
   reason. The *why* is what lets a subagent three states later decide something you did not
   anticipate.
3. **Code conventions** — naming, error handling, comment density, what "done" looks like.
4. **Folder layout** — where things go, so nobody has to guess or invent a parallel structure.
5. **Working rules** — see `references/reglas-de-trabajo.md`.

## Step 2: Fill the fields that carry the pains

These are not decoration; each one exists because something used to go wrong.

```yaml
git:
  branch_por_feature: true
  patron_branch: "feat/{feature-id}-{slug}"
  commit: conventional
  merge: no-ff              # no-ff | squash | ff
  branch_base: main
```

> **This is where "it never creates the branch per feature" dies.** The rule stops being
> something a skill remembers and becomes a field `sf` reads and acts on. No skill carries git
> conventions of its own any more (R4) — they read this.

```yaml
manifiesto: go.mod
dependencias_aprobadas: []
```

> **And this is where "libraries outside the constitution" dies — by warning, not blocking.**
> `sf` compares the manifest against the approved list and says something. It does **not** stop
> you: the last word is Javier's, and a tool that blocks on its own breaks that rule.
>
> **Start the list empty.** It is not written up front — it builds itself, one approval at a
> time, exactly like the model map.

### `mutacion` — you have to choose, and the skeleton will not choose for you

```yaml
mutacion: "npx stryker run"       # or gremlins · cargo-mutants · mutmut · mutant · PIT
```

**Go look for the tool for this stack.** `sf init` leaves the line with a `⚠` naming the one it
knows about for your language; if it does, that is the fact, not a suggestion to skim past.

**Empty is a legitimate answer, and it is the one thing you may not leave by default.** If you
decide there is no good tool, say so in the body and say why:

```markdown
## Reglas de trabajo
No hay herramienta de mutación: <razón>. El ㉒ lo hace el revisor leyendo,
y sus mutantes se guardan en `.docs/<feature>/mutantes/`.
```

> **Why this is not a formality.** With a tool, the ㉒ mutates the same code the same way every
> round and the score is comparable. Without one, the reviewer writes the mutants by hand and the
> set is different every round — 56 one time, 82 the next — so `73% → 85%` looks like progress
> and is a different exam. That is why the empty case has to be a decision on the record and not
> a field nobody read: whoever reviews needs to know which of the two numbers they are holding.

## Step 3: Write it, then stop — the ⑧ is Javier's

Write `.docs/constitucion.md` from `templates/constitucion.tmpl.md`, print the gate, and stop:

```
───────────────────────────────────────
🛑 ⑧ — constitution written
Stack: <one line>  ·  test_cmd: <the command>
Awaiting: sf approve  /  sf reject "motivo"
───────────────────────────────────────
```

Then `sf done`. The gate checks one thing — that `test_cmd` exists. **Everything else is
judgment, and judgment is Javier's.**

## Rules

- Technical only. The why lives in the brief and the PRD; do not copy it here.
- One file. Never split architecture, stack and conventions again.
- `test_cmd` is mandatory. Three gates depend on it.
- Every stack choice carries its reason. A cold subagent reads the reason, not your intent.
- `dependencias_aprobadas` starts empty and grows by approval.
- Include a section only if it changes a decision. A section that would say "N/A" is noise.
- You do not seal. The ⑧ is one of Javier's three decisions.
