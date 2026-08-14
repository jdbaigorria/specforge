# Provenance — how scout stays honest

The single thing that separates real discovery from a confident hallucination is
**where each claim came from**. Every line of evidence and every load-bearing
statement in the brief carries a provenance tag.

## Two tags, never blurred

- **`retrieved`** — backed by a source pulled this session (a search result, a
  repo, a page). **Always include the link.**
- **`model-prior`** — from the model's training. Plausible, but **unverified**.
  Allowed, but it must say so.

Format in markdown:

```
- Competitors cluster around calendar sync, not task capture. [retrieved]
  - https://example.com/landscape-article
- The CLI-task-manager space feels crowded. [model-prior — unverified]
```

## Rules

- A claim with no tag is a bug. Tag it or cut it.
- `retrieved` without a link is just `model-prior` wearing a costume — demote it.
- The brief's verdict (proceed/pivot/kill) must rest mainly on `retrieved`
  evidence. If it rests on `model-prior`, the verdict is "insufficient evidence,"
  not "proceed."
- Counts surface in the gate line (`Evidence: retrieved N / model-prior M`) so
  the human sees how grounded the recommendation actually is.

## Why this matters

AI is fluent enough to invent a competitive landscape that reads like a McKinsey
slide. Provenance is the discipline that stops scout from doing that — it makes
the difference between "here's what I found" and "here's what I'd guess" explicit
and auditable.
