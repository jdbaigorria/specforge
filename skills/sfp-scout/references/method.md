# De-risking method

The methodology scout uses in Steps 2–3. It internalizes the useful parts of
product-discovery practice (JTBD, gap analysis, positioning, demand evidence) so
scout doesn't depend on external PM skills — same "compose, don't depend"
principle used for `sfx-think`/`sfx-grill-me`. Every output of these steps still
carries provenance (`references/provenance.md`).

## 1. Demand evidence — the honest answer to "AI can't prove demand"

AI cannot prove demand, but it can find **traces** of it. Go looking for them
with the research tools (`references/tooling.md`) instead of guessing:

| Source | What it tells you | How to reach it |
|--------|-------------------|-----------------|
| Reddit / Hacker News / forums | real users complaining about the problem, or about the incumbents | Tavily search scoped to the site (`site:reddit.com …`), then `fetch` the threads |
| Review sites (G2, Trustpilot, app stores) | where incumbents are weak — the gap you could own | Tavily search; for anti-bot pages use the heavy tier (Bright Data / Playwright) |
| News / trends | is the space heating up or cooling down | Tavily news search |
| GitHub issues on similar repos | unmet needs in open-source alternatives | `github` search → `list_issues` |

A complaint thread with 200 upvotes is `retrieved` demand evidence. "People
probably want this" is `model-prior`. Treat them differently.

## 2. Jobs-to-be-Done (JTBD) — frame the real problem

Don't describe the product; describe the **job the user is hiring it to do**.

> When [situation], I want to [motivation], so I can [expected outcome].

This keeps scout honest about *who has the problem and when*, and stops a
solution-in-search-of-a-problem. The main JTBD becomes the spine of the
differentiator and the `PR#` requirements.

## 3. Gap analysis & positioning — where you could win

From the landscape (Step 2 research):

- **Feature/coverage map:** what the incumbents do; what they all skip.
- **Blind spots:** segments or jobs the incumbents structurally ignore (their
  business model makes them unable to serve it well).
- **Positioning:** the one-line "for X who Y, unlike Z, this does W." If you can't
  fill it from *retrieved* evidence, the differentiator isn't proven yet.

A real gap backed by `retrieved` evidence → `proceed`. A crowded space with no
defensible gap → `kill`. A gap that's elsewhere than you thought → `pivot`.

## 4. From method to artefacts

- Demand evidence + landscape → `research.md` (tagged).
- JTBD + gap + positioning → the differentiator and MVP boundary in the brief.
- The prioritized jobs → product requirements `PR1…PRn`.

Optional rigor (don't force it — "no ceremony without purpose"): if there are
many candidate problems, a quick RICE-style sort (reach × impact × confidence ÷
effort) can order them — but a one-line rationale per `PR#` is usually enough.
