---
name: aws-architect
description: >
  Design AWS infrastructure with cost, security, reliability, and scalability in mind.
  Produce an architecture document. Trigger: "/aws-architect", "design the infra",
  "how should I deploy", "what AWS services", "architecture for", "cloud design".
---

# aws-architect

Design AWS infrastructure. Evaluate through Well-Architected Framework lens. Justify every choice with tradeoffs.

**Always produces** `.ai/architectures/{slug}.md`. Model: opus.

## Step 1: Clarify Requirements

Ask only what's not clear:
- **Workload type**: API, batch, data pipeline, static site, ML inference
- **Scale**: requests/sec, data volume, concurrent users, growth
- **Budget**: monthly ceiling, pay-per-use vs reserved
- **Compliance**: GDPR, HIPAA, SOC2, data residency
- **Team**: size, AWS experience, on-call capacity

Read `.ai/project.md` for stack context if exists.

## Step 2: Design with Tradeoffs

For each service choice, state:
- **What**: the service and its role
- **Why not alternatives**: what you rejected and why
- **Cost model**: how it charges, estimated monthly
- **Blast radius**: what fails if this service goes down
- **Scaling**: behavior at 10x load

## Step 3: Write Artifact

Generate `.ai/architectures/{slug}.md` using `templates/architecture.tmpl.md`.

Only include sections relevant to the design's complexity. A static site doesn't need Multi-region Reliability. A HIPAA-compliant API does.

## Step 4: Return Summary

```
## Architecture: {name}

**Artifact**: .ai/architectures/{slug}.md
**Pattern**: {e.g., "Serverless API", "Event-driven microservices"}
**Services**: {count}
**Est. cost**: ${N}/month
**Risks**: {count} ({highest severity})
**Open questions**: {count}

Next: {sf-propose or "resolve open questions"}
```

## Rules

- ALWAYS state cost estimates. Give a range, not "it depends."
- Simplest architecture first. Add complexity only when justified.
- NEVER recommend a service without explaining why not a simpler alternative.
- Security is NOT optional — IAM, encryption, network isolation always.
- If scale doesn't justify complexity (EKS for 100 req/day), say so.
- Flag AWS lock-in decisions explicitly.
- Keep artifact under 800 words. Tables over prose.
