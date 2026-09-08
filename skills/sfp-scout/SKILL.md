---
name: sfp-scout
description: >
  Turn a fuzzy product idea into a sealed brief with a verdict. The front-end of the SpecForge
  machine (state `brief`, steps ①–⑤): it COMPOSES the interview, the research, the glossary and
  the prototype primitives, and adds the one thing they do not have — a verdict and the ⑥ gate.
  Invoked by the orchestrator when `sf next` returns `skill: sfp-scout`, not on your own
  initiative. Also usable standalone: "scout this idea", "should I build X", "is this worth
  building", "/sfp-scout". NOT validation — AI cannot prove demand; it gathers evidence and
  surfaces risk.
---

# sfp-scout

The first state of the machine. **Fuzzy idea → evidence → a brief with a verdict.**

```bash
sf context   # your envelope. The brief is the ONE state with an empty envelope, on purpose:
             # ①–⑤ is an interview about an idea that has no shape yet. There is nothing to read.
```

**Framing, non-negotiable:** this **de-risks**, it does not **validate**. AI cannot prove market
demand or willingness to pay. Say so; never present research as proof.

**This state talks to Javier directly** (`via: vos`). It is not a subagent — the back-and-forth
*is* the step.

## This skill is a composer

It owns the **outcome** — the brief, the verdict, the ⑥. It does **not** own the methods, and it
does not repeat them. Four primitives do the work:

| when | call |
|---|---|
| the whole of ①–⑤ | `Call the Skill tool with "sfx-grilling"` |
| a fact is missing | `sfx-grilling` calls `sfx-buscar` on its own |
| a word means two things | `sfx-grilling` calls `sfx-vocabulario` on its own |
| only running code answers it | `sfx-grilling` calls `sfx-prototipo`, **after Javier approves** |

**You call `sfx-grilling` once and it drives.** Do not re-explain how to interview, do not run your
own question loop, do not batch the research into a separate phase. **Research is a branch of the
tree, not a stage** — that is the whole point of this shape, and it is why there is no loop to
close here.

## Step 1 — the interview (①–⑤)

`Call the Skill tool with "sfx-grilling"`, telling it: the raw idea, that the output file is
`.docs/entrevista.md`, and that the tree must cover these five branches before the frontier can be
empty:

1. **What it is** — one sentence Javier agrees with. If it does not fit in one, the tree is not done.
2. **Who and when** — the job being hired: *"Cuando <situación>, quiero <motivación>, para
   <resultado>"*. Not "everyone".
3. **What already exists** — the landscape, via `sfx-buscar`. Level 0 first.
4. **The gap, both halves** — where you differ (a hole the incumbents *cannot* close without
   breaking their own model) **and what you stock up on** (what comparable repos already solved and
   you should take instead of rewriting). A scout that only hunts gaps hands you a product built
   from scratch.
5. **What has to be true** — what kills this, which assumption is most fragile, what only building
   it can settle.

The interview ends when the frontier is empty. Not when it feels long enough.

## Step 2 — write the brief (⑤)

Write `.docs/brief.md` from `templates/brief.tmpl.md`.

**The brief is the ARGUMENT, not the archive.** The evidence already lives in `.docs/evidencia.md`
and the reasoning in `.docs/entrevista.md`. The brief cites them; it does not copy them. If a
section of the brief could be replaced by a pointer, make it a pointer.

Frontmatter, and `sf` reads it:

```yaml
---
veredicto: hacelo   # hacelo | pivotea | no-lo-hagas — the one YOU propose
---
```

**Write the verdict you propose.** Not empty — the same one you argue for in the body and print in
the gate line. `sf approve` means *"sí, sellalo con lo que dice"*, so the file has to say
something; an empty `veredicto` deadlocks the ⑥ and the only way out is editing the file by hand.
Javier's move is `approve` or `reject`, never filling in a blank.

## Step 3 — the ⑥, and it is Javier's

You do not seal. Print the gate and stop:

```
───────────────────────────────────────
🛑 ⑥ — brief listo: "<la idea en una línea>"
Propongo: <hacelo | pivotea | no-lo-hagas>
Evidencia: retrieved N / model-prior M / probado P  ·  links L
Ronda(s): R  ·  preguntas abiertas: 0
Diferencial: <una línea>
No pude comprobar: <lo que quedó afuera, o "nada">
Esperando: sf approve  /  sf reject "motivo"
───────────────────────────────────────
```

Then `sf done`. The gate checks the mechanical part — the three files exist, the interview closed
with no open branches, the evidence cites at least one real source, the verdict is one of three. It
does **not** check whether the brief is *good*, nor whether the evidence is *enough*. That is
judgment, and judgment is Javier's.

> **Measured, 2026-09-05.** Same idea, same seed text, two models. One returned `no-lo-hagas` with
> 13 links; the other `hacelo` with 8 `model-prior` claims and **zero** links, and the machine
> accepted both. The second was not lying — it tagged provenance correctly, exactly as asked. **The
> harness was blind and the gate asked for too little.** Both halves are fixed now; this skill is
> the half that stopped guessing.

## Rules

- Compose, never re-explain. The method belongs to the primitive.
- De-risk, never claim validation. Name what AI cannot know.
- Research is a branch of the interview, not a phase before it.
- Differentiate **and** stock up. Both halves.
- `no-lo-hagas` is a success. The value of the ⑥ is being able to say no.
- The brief argues and cites. The evidence lives in `evidencia.md`.
