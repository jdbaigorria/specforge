# Pipeline: {{name}}

**Date**: {{date}}
**Status**: {{draft | approved | implemented}}
**Type**: {{batch | streaming | hybrid}}
**Frequency**: {{schedule}}

## Summary

{{what_this_pipeline_does_why_it_matters_sla}}

## Data Contract

### Source
- **Origin**: {{system}}
- **Format**: {{json_csv_parquet_avro}}
- **Volume**: {{records_per_day}}
- **Schema stability**: {{stable | evolving | unstable}}

### Destination
- **System**: {{warehouse_db_bucket}}
- **Format**: {{final_format}}
- **Consumers**: {{who_reads_how}}
- **SLA**: {{freshness_availability}}

## Schema

| Field | Type | Source | Validation | Required | Notes |
|-------|------|--------|-----------|----------|-------|
| {{field}} | {{type}} | {{source_field}} | {{rule}} | {{y_n}} | |

## Data Flow

```
{{ascii_or_mermaid_diagram}}
```

## Stages

| Stage | Input | Output | Logic | Error Handling |
|-------|-------|--------|-------|----------------|
| Extract | {{source}} | raw | {{how}} | {{retry_dlq}} |
| Validate | raw | validated | {{rules}} | {{reject_quarantine}} |
| Transform | validated | transformed | {{logic}} | {{fallback}} |
| Load | transformed | {{dest}} | {{upsert_append}} | {{retry}} |

## Quality Gates

Records MUST pass all gates before loading:
- [ ] {{validation_rule_1}}
- [ ] {{validation_rule_2}}
- [ ] Schema match against destination
- [ ] Dedup check

Bad records: {{strategy_quarantine_dlq_alert_skip}}

## Idempotency

- **Dedup key**: {{fields}}
- **Upsert strategy**: {{how_reruns_avoid_duplicates}}
- **Watermark**: {{incremental_tracking}}

## Backfill

- **Trigger**: {{how_to_invoke}}
- **Parameters**: {{date_range_partition}}
- **Safety**: {{avoid_overwriting_good_data}}

## Observability

### Metrics
| Metric | Measures |
|--------|----------|
| Records processed | Throughput |
| Records rejected | Quality |
| Pipeline latency | Freshness |
| Error rate | Reliability |

### Alerts
| Alert | Threshold | Action |
|-------|-----------|--------|
| No records in {{N}} min | {{threshold}} | {{action}} |
| Error rate > {{X}}% | {{threshold}} | {{action}} |

### Lineage
{{how_to_trace_record_to_source}}

## Storage Design

- **Partitioning**: {{strategy}}
- **Retention**: {{how_long}}
- **Compression**: {{format_codec}}
- **Est. cost**: ${{n}}/month at target volume

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| {{risk}} | {{severity}} | {{mitigation}} |

## Open Questions

- [ ] {{question}}
