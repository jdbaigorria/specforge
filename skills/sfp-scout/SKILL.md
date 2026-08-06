---
name: sfp-scout
description: >
  De-risk a product idea before it becomes a project. The "from zero" front-end of SpecForge:
  research the landscape (web + GitHub + docs via MCP), stress-test the idea, and produce a
  visible discovery brief with a proceed/pivot/kill verdict and traceable product-requirement
  IDs that feed sf-init. Use for greenfield only, when the idea is still fuzzy: "scout this idea",
  "should I build X", "is this worth building", "de-risk this", "research this product idea",
  "/sfp-scout", "explore this product". NOT validation — AI cannot prove demand; it gathers
  evidence and surfaces risk. Requires the research MCPs (see references/tooling.md); without
  them it will not fabricate research.
---

# sfp-scout

The product-creation companion (`sfp-` = product tier). SDD is strong once you
know what to build; `sfp-scout` covers the step before that — **fuzzy idea →
evidence → de-risked brief** — so SpecForge spans the whole arc from zero to a
maintained feature, not just delivery.

**Framing, non-negotiable:** this **de-risks**, it does not **validate**. AI
cannot prove market demand or willingness to pay. It can gather evidence, compare
alternatives, find gaps, and stress-test assumptions. Say so; never present
research as proof.

**Scope:** greenfield only. Brownfield goes straight to `sf-init`.

## Step 0: Tooling check (gate on evidence)

`sfp-scout` runs on real evidence, so it needs the research MCPs (web search,
GitHub, docs — see `references/tooling.md`). Check they're available:

- **Available** → proceed with retrieved evidence.
- **Missing** → do NOT silently invent competitors from training data (that is
  the false-validation trap). Either: (a) tell the user which MCPs to enable and
  stop, or (b) with explicit consent, run a **degraded pass** where every claim
  is marked `model-prior` (unverified) and the brief is stamped "low-evidence."

## Step 1: Capture the idea

One short exchange: what is it, who is it for, what problem, what outcome. Don't
over-interview here — the grilling comes after there's evidence to grill against.

## Step 2: Research (provenance per claim)

Use the MCPs to map the landscape: similar products, comparable GitHub repos,
what they solve, what they miss, demand/saturation signals. Hunt for **demand
evidence** (real complaints on Reddit/HN, weaknesses in incumbent reviews, news
trends) per `references/method.md` — that is the closest thing to proof of demand,
and it must be `retrieved`, not imagined. Write `specforge/product/<slug>/research.md`.

**Every claim carries provenance** (`references/provenance.md`):

- `retrieved` — backed by a real source, with the link.
- `model-prior` — from training, **unverified**, flagged as such.

When the material was **handed to you** rather than found — a client's emails,
meeting notes, screenshots, an existing prototype — the same axis widens:

- `sourced` — a literal quote, with the `sources.json` ref it came from (`S3`).
- `inferred` — deduced from that material, **not confirmed** by anyone.
- `unanswered` — you asked and nobody has answered yet. Open it with
  `sf question add` so it survives this session; see below.

It is one axis, not two systems: *where did this come from and how much do I
believe it*. A research doc that blurs these lies with confidence.

**The hard rule, in three categories:**

1. **Analogy** with something already approved in this same project — allowed,
   labelled as a hypothesis.
2. **External practice** — offered with its reason, and **the human decides**.
   Even once adopted it stays labelled *external practice, not confirmed by the
   client*. It never graduates to settled fact. `sf question adopt` records
   exactly this state.
3. **Unmarked invention** — forbidden. This is the same prohibition as "no
   fabricated research" above, extended to material you were handed.

**Unknowns are objects, not notes.** For a founder an unknown is a risk you go
resolve yourself; on the consulting path it is a question you send and then wait
on, sometimes for weeks, while work continues around it. Record each one:

```bash
sf question add --text="is VAT included in the price?" \
  --impact="changes the Order model and how the total is computed" \
  --blocks=checkout --asked-of="client finance lead"
```

`--impact` is required — an unknown with no stated consequence cannot be ranked
against the others. A question that lists `--blocks` will **stop that feature's
verdict** until someone answers it or adopts an external practice on the record.
That is deliberate: dropping something from scope because an answer never arrived
must be a decision someone made, never a silent default.

## Step 3: De-risk (compose think + grill-me)

This is where `sfx-think` and `sfx-grill-me` are composed (not replaced — they
stay standalone), using the method in `references/method.md` (JTBD, gap analysis,
positioning, blind spots):

- **think** the solution space: frame the **JTBD**, the differentiator, the gap
  the incumbents structurally skip, the MVP boundary.
- **grill-me** the assumptions: who exactly, why now, what kills this, what has
  to be true. Ground every challenge in the Step 2 evidence.

Track decisions and rejected directions in
`specforge/product/<slug>/decision-log.md`.

## Step 4: Synthesize the discovery brief

Write `specforge/product/<slug>/discovery-brief.md` (template
`templates/discovery-brief.tmpl.md`). It is the **lightweight** product doc — no
13-section PRD. Crucially, it states **product requirements with stable IDs**
(`PR1`, `PR2`, …) so traceability extends upward (see
`references/traceability.md`):

```
PR-requirement → roadmap feature → feature requirement → task → code → test
```

## Step 5: Verdict + gate

The verdict is first-class and includes **kill**:

- **proceed** — evidence supports building; differentiator is real.
- **pivot** — the gap is elsewhere; reframe the idea.
- **kill** — crowded/served/weak; the best outcome is not to build. Say it plainly.

**A second, independent axis: is this enough to start specifying?**

`proceed / pivot / kill` answers *should this be built* — a founder's question.
When the decision to build is already made and someone handed you the material,
that axis is nearly always `proceed`, and it stops carrying information. The one
that matters then is a binary:

- **ready** — the material covers enough to write requirements against.
- **not ready** — there are blocking gaps. Name them, and open each one with
  `sf question add`. "Not ready" with no questions on record is just an opinion.

The two axes coexist and do not replace each other: an idea can be `proceed` and
`not ready` at the same time, and saying both is more useful than picking one.

```
───────────────────────────────────────
🔴 GATE — discovery: <proceed | pivot | kill> for "<idea>"
Evidence: <retrieved N / model-prior M>  ·  Differentiator: <one line>
Awaiting approval. Reply: approve / reject / change X
───────────────────────────────────────
```

## Step 6: Handoff

On approved **proceed** → next step is `sf-init --from
specforge/product/<slug>/discovery-brief.md`. `sf-init` seeds the constitution
from the brief; `sf-propose --all` derives the roadmap from the `PR#` ids,
carrying the traceability into the spec pipeline.

## Rules

- De-risk, never claim validation. Name what AI cannot know.
- No evidence tools → no fabricated research. Gate on tooling.
- Provenance on every claim. `model-prior` is allowed but always labelled.
- `kill` is a success. The point is to avoid building the wrong thing.
- Greenfield only. The brief is lightweight and feeds `sf-init --from`.

<!-- Possible future upgrade (not in scope): an OPTIONAL sfp-po skill that turns
     the brief into a formal 13-section PRD (personas, journeys, success metrics…)
     for large/stakeholder products. Deliberately kept out of scout to preserve
     "no ceremony without purpose": the lightweight brief + PR# ids are the SDD
     input. The heavy PRD would be opt-in, never the default, and must not become
     a third translation layer between brief and sf-propose. -->

