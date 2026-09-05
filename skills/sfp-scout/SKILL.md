---
name: sfp-scout
description: >
  Turn a fuzzy product idea into a sealed brief with a verdict. The front-end of the SpecForge
  machine (state `brief`, steps ①–⑤): ping-pong the idea until it has shape, research whether it
  already exists, find the gap, stress-test it, and write `.docs/brief.md` with a
  hacelo/pivotea/no-lo-hagas verdict for Javier to seal. Invoked by the orchestrator when
  `sf next` returns `skill: sfp-scout`. Also usable standalone: "scout this idea", "should I
  build X", "is this worth building", "de-risk this", "/sfp-scout". NOT validation — AI cannot
  prove demand; it gathers evidence and surfaces risk. Needs the research MCPs; without them it
  will not fabricate research.
---

# sfp-scout

The first state of the machine. **Fuzzy idea → evidence → a brief with a verdict.**

```bash
sf context      # your envelope. The brief is the ONE state with an empty envelope —
                # ①–⑤ is a ping-pong about an idea that has no shape yet.
```

**Framing, non-negotiable:** this **de-risks**, it does not **validate**. AI cannot prove market
demand or willingness to pay. It can gather evidence, compare alternatives, find gaps, and
stress-test assumptions. Say so; never present research as proof.

**This state talks to Javier directly** (`via: vos`). It is not a subagent — the whole point of
①–⑤ is the back-and-forth.

## Step 1: Ping-pong the idea (⑫ this is where it takes shape)

**Interview properly. Do not rush this.** The idea arrives fuzzy and leaves with edges: what it
is, who it is for, what problem, what outcome, what it explicitly is *not*.

> **This step comes first on purpose.** An earlier version of this skill said *"don't
> over-interview here — the grilling comes after there's evidence."* That is backwards for this
> flow: **researching a shapeless idea returns shapeless results.** You cannot search for
> competitors to something you cannot yet describe in one sentence.

Leave when you can state the idea in **one sentence** and Javier agrees with it.

## Step 2: Research — does it already exist? (③)

Use the MCPs to map the landscape: similar products, comparable GitHub repos, what they solve,
what they miss, demand and saturation signals. Hunt for **demand evidence** — real complaints,
weaknesses in incumbent reviews, trends — per `references/method.md`.

**Every claim carries provenance** (`references/provenance.md`), and this is the best thing this
skill contributes:

- `retrieved` — backed by a real source, **with the link**.
- `model-prior` — from training, **unverified**, flagged as such.

> **Why provenance matters more here than anywhere else in the flow.** The ⑥ is the one point
> where a hallucination costs the whole product: sealing `no-lo-hagas` because something exists
> that does not actually exist. There is no later step that catches it.

**Tooling gate.** If the research MCPs are unavailable (`references/tooling.md`), do **not**
silently invent competitors from training data — that is the false-validation trap. Either
(a) name the MCPs to enable and stop, or (b) with explicit consent run a **degraded pass** where
every claim is `model-prior` and the brief is stamped low-evidence — which means writing
`evidencia: baja` in the frontmatter. That line is the ONLY way a brief with no cited sources
gets through the ⑥, and it must be a deliberate declaration, never a shortcut.

## Step 3: Does it help me, and where do I differ? (④)

Two halves, and **both are required** — this step is not only about differentiating:

- **Differentiate** — the gap the incumbents structurally skip. Not "ours is nicer": a reason
  they *cannot* close it without breaking their own model.
- **Stock up** — what the comparable repos already solved that you should take instead of
  rewriting. A scout that only looks for gaps hands you a product built from scratch.

## Step 4: Is there consensus? (⑤ — compose think + grill-me)

This is where `sfx-think` and `sfx-grill-me` are **composed** — not replaced, they stay
standalone:

- **`sfx-think`** the solution space: the JTBD, the MVP boundary, the shape of the thing.
- **`sfx-grill-me`** the assumptions: who exactly, why now, what kills this, what has to be true.
  **Ground every challenge in the Step 2 evidence** — grilling against opinion is theatre.

## Step 5: Write the brief

Write `.docs/brief.md` from `templates/brief.tmpl.md`. **One file** — no separate research doc,
no decision log. What survives of the research is what the brief cites.

The frontmatter carries the verdict, and `sf` reads it:

```yaml
---
veredicto: hacelo   # hacelo | pivotea | no-lo-hagas — the one YOU propose
evidencia: baja     # ONLY on a degraded pass (Step 2). Omit it otherwise.
---
```

**Cite your sources in the body.** The ⑥ gate counts links: a brief that carries a verdict and
not one single URL does **not** seal. This is not about how much research is enough — that is
Javier's call — it is about the difference between *some* and *none*.

> **Measured, 2026-09-05.** The same idea, the same seed text, two models. One returned
> `no-lo-hagas` with 13 cited links; the other returned `hacelo` with 8 `model-prior` claims and
> **zero** links, and the machine accepted both. The second one was not lying — it marked every
> claim `model-prior — unverified`, exactly as this skill asks. The gate was the part that
> asked for too little. It no longer does.

**Write the verdict you propose.** Not empty — one of the three, the same one you argue for in
the body and print in the gate message. `sf approve` means *"yes, seal it with what it says"*, so
the file has to say something; an empty `veredicto` deadlocks the ⑥ — `sf done` refuses to move,
`sf approve` has nothing to seal, and the only way out is editing the file by hand.

Javier's decision is `approve` or `reject`, not filling in the blank. If he wants a different
verdict than the one you propose, he rejects and the ①–⑤ runs again.

## Step 6: The ⑥ — stop, and it is Javier's

You do not seal. Print the gate and stop:

```
───────────────────────────────────────
🛑 ⑥ — brief ready: "<idea in one line>"
Proposed: <hacelo | pivotea | no-lo-hagas>
Evidence: <retrieved N / model-prior M>  ·  Differentiator: <one line>
Awaiting: sf approve  /  sf reject "motivo"
───────────────────────────────────────
```

Then `sf done`. The gate runs: the brief must exist, carry a valid `veredicto`, and cite at
least one source — unless it declares `evidencia: baja`. It does **not** check whether the brief
is *good*, nor whether the sources are strong enough — that is judgment, and judgment is
Javier's.

## Rules

- De-risk, never claim validation. Name what AI cannot know.
- No evidence tools → no fabricated research. Gate on tooling.
- Provenance on every claim. `model-prior` is allowed but always labelled.
- `no-lo-hagas` is a success. The value of the ⑥ is being able to say no.
- Ping-pong **before** research. A shapeless idea returns shapeless results.
- Differentiate **and** stock up. Both halves of the ④.
- One file: `.docs/brief.md`. Everything else is conversation.
