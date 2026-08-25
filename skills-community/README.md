# Community skills

Domain-generic skills that are useful but **not part of the maintained
SpecForge core** (D7' of the platform evaluation: every generic skill competes
for maintenance time against the enforcement spine nobody else has — less
surface, more spine).

They don't participate in the SDD pipeline, aren't checked by the CI that validates `skills/`, and
aren't installed by `sf install`. To use one, copy or symlink its folder into
your agent's skills directory:

```bash
ln -s "$(pwd)/skills-community/sfx-aws-architect" ~/.claude/skills/
```

| Skill | What it does |
|-------|--------------|
| `sfx-aws-architect` | Design AWS infrastructure with tradeoffs and cost analysis |
| `sfx-data-engineer` | Design data pipelines with quality gates |

Contributions of new community skills are welcome — this folder is the staging
area for what may become a registry.
