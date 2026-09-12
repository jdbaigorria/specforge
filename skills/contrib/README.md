# Community skills

Domain-generic skills that are useful but **not part of the maintained
SpecForge core** (D7' of the platform evaluation: every generic skill competes
for maintenance time against the enforcement spine nobody else has — less
surface, more spine).

They don't participate in the SDD pipeline and **aren't published by the plugin** —
`plugin.json` enumerates the 25 skills of `maquina/` and `utiles/`, and this folder is not
among them. `sf install` doesn't install them either.

They ARE checked by the CI, on two points. A skill here fails the build if it names an `sf`
command this binary doesn't have, or if it names a concrete provider model where a model
will read it. Broken is broken, and somebody is going to copy it — that second check is
why: these are the skills most likely to be copied out of the repo by hand.

To use one, copy or symlink its folder into your agent's skills directory:

```bash
ln -s "$(pwd)/skills/contrib/sfx-aws-architect" ~/.claude/skills/
```

| Skill | What it does |
|-------|--------------|
| `sfx-aws-architect` | Design AWS infrastructure with tradeoffs and cost analysis |
| `sfx-data-engineer` | Design data pipelines with quality gates |

Contributions of new community skills are welcome — this folder is the staging
area for what may become a registry.
