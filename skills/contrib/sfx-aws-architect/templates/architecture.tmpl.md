# Architecture: {{name}}

**Date**: {{date}}
**Status**: {{draft | approved | implemented}}
**Workload**: {{type}}
**Scale**: {{expected_metrics}}

## Summary

{{what_this_achieves_for_whom_at_what_scale}}

## Requirements

| Requirement | Value | Notes |
|-------------|-------|-------|
| Users | {{n}} | |
| RPS | {{n}} | |
| Budget | ${{n}}/mo | |
| Availability | {{sla}} | |
| Compliance | {{list}} | |

## Services

| Service | Role | Why this over alternatives | Est. $/mo |
|---------|------|---------------------------|-----------|
| {{svc}} | {{role}} | {{rationale}} | ${{est}} |

## Data Flow

```
{{ascii_or_mermaid_diagram}}
```

## Security

- **IAM**: {{approach}}
- **Network**: {{vpc_design}}
- **Encryption**: {{at_rest_and_transit}}
- **Secrets**: {{management_approach}}

## Scaling Strategy

{{how_it_handles_10x_load}}

## Reliability

- **Availability target**: {{sla}}
- **Failure modes**: {{main_ones}}
- **Recovery**: RPO {{rpo}} / RTO {{rto}}
- **Multi-AZ / Multi-region**: {{approach}}

## Cost Breakdown

| Service | Unit cost | Usage | Monthly |
|---------|-----------|-------|---------|
| {{svc}} | {{unit}} | {{usage}} | ${{monthly}} |
| **Total** | | | **${{total}}/mo** |

**Optimization opportunities**: {{list_2_3}}

## Risks

| Risk | Impact | Mitigation |
|------|--------|------------|
| {{risk}} | {{severity}} | {{mitigation}} |

## Decisions Log

| Decision | Chose | Rejected | Why |
|----------|-------|----------|-----|
| {{topic}} | {{choice}} | {{alt}} | {{rationale}} |

## Open Questions

- [ ] {{question}}
