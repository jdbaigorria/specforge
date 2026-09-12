---
name: sfx-data-engineer
delegate: true
description: >
  Design data pipelines and schemas with quality, observability, and idempotency in mind.
  Produce a pipeline design document. Trigger: "/data-engineer", "data pipeline", "ETL",
  "data model", "schema design", "data flow", "data quality", "batch processing", "streaming".
---

# data-engineer

Design data systems that are reliable, observable, and maintainable. Think in data contracts, lineage, quality gates, and idempotency.

**Always produces a written design.** Model: opus. It goes wherever the user asks — this skill is standalone and does not write inside `.docs/` uninvited.

## Step 1: Understand the Data

Clarify what's not obvious:
- **Source**: origin, format, frequency, volume, schema stability
- **Destination**: consumer, SLA, query patterns
- **Transformations**: business logic, aggregations, enrichments, joins
- **Quality**: validation rules, what makes a record invalid, handling strategy
- **Freshness**: real-time, near-real-time, hourly, daily, on-demand
- **History**: retention needs, reprocessing requirements

Read `.docs/constitucion.md` for stack context if it exists.

## Step 2: Design the Pipeline

For each pipeline, define:
- **Stages**: extract → validate → transform → load → verify
- **Idempotency**: dedup key, upsert logic, watermarks
- **Error handling**: dead letter queue, retry policy, alerting thresholds
- **Backfill**: how to reprocess historical data
- **Observability**: metrics, alerts, lineage tracking

## Step 3: Write Artifact

Write it using `templates/pipeline.tmpl.md`.

Only include sections relevant to the pipeline's complexity. A simple CSV-to-DB loader doesn't need Streaming considerations. A real-time event pipeline does.

## Step 4: Return Summary

```
## Pipeline: {name}

**Artifact**: the path the user chose
**Type**: {batch | streaming | hybrid}
**Volume**: {records/day}
**Latency**: {SLA}
**Storage**: ${N}/mo estimated
**Quality gates**: {count}
**Risks**: {count}

Next: {`sf new "…"` to put it in the backlog, or "resolve open questions"}
```

## Rules

- ALWAYS design for re-runnability. Every pipeline must be idempotent.
- ALWAYS include quality gate before loading. Bad data doesn't reach consumers.
- ALWAYS define error handling per stage.
- Schema changes are breaking changes — flag with migration plan.
- If volume doesn't justify streaming, use batch. Simple first.
- Observability is not optional. Every pipeline has metrics + alerts.
- Consider cost per GB at target volume.
- Keep artifact under 900 words. Tables over prose.
