# Installing SpecForge

SpecForge is a pack of agent skills plus an `AGENT.md` orchestrator. The skills
are plain markdown — any agent that can load skills from a directory can run
them. Installation is just making `skills/` (and `AGENT.md`) visible to your
agent.

The suite has two tiers:

- **Pipeline + audit** (`sf-init`, `sf-propose`, `sf-build`, `sf-check`,
  `sf-audit`) — the SpecForge workflow.
- **Support skills** (`triage`, `documenter`, `explain`, `product-owner`,
  `aws-architect`, `data-engineer`, `grill-me`, `tdd`, `github`, `think`) —
  standalone helpers that don't require `specforge/` to be initialized.

You can install all of them or only the pipeline.

---

## 1. Clone the repo

```sh
git clone https://github.com/<you>/specforge.git
cd specforge
```

Everything below assumes `/path/to/specforge` is where you cloned it.

---

## 2. Make the skills visible to your agent

Pick the option that matches your harness. Symlinking (instead of copying) keeps
the skills updated when you `git pull`.

### Project-local (recommended for trying it on one repo)

```sh
cd /your/project
mkdir -p .agents
ln -s /path/to/specforge/skills .agents/skills
cp /path/to/specforge/AGENT.md .agents/AGENT.md   # or merge into your existing one
```

### Global (available in every project)

```sh
ln -s /path/to/specforge/skills ~/.agents/skills
```

Or point your agent settings at the directory directly:

```json
{ "skills": ["/path/to/specforge/skills"] }
```

### Claude Code

Copy or symlink into the Claude skills directory:

```sh
# global
ln -s /path/to/specforge/skills ~/.claude/skills
# or project-local
ln -s /path/to/specforge/skills .claude/skills
```

> Distribution as a Claude Code plugin (`.claude-plugin/`, installable via
> `/plugin marketplace add` + `/plugin install`) is planned — see the roadmap in
> `SPECFORGE-REVIEW.es.md` (v0.3, packaging).

---

## 3. Wire up the orchestrator

`AGENT.md` tells the agent how to dispatch to the right skill, where artefacts
live, and how the gate protocol works. Place its contents where your harness
reads agent instructions (e.g. `AGENT.md`, `CLAUDE.md`, or your agent's system
prompt). If you already have one, merge the **SpecForge Workflow** and **Skills**
sections in.

---

## 4. Verify

In a fresh agent session, ask it to run a SpecForge skill, e.g.:

```
sf-init
```

It should scaffold `specforge/` and start the constitution conversation. If the
agent can't find the skill, the `skills/` directory isn't on its load path —
recheck step 2.

---

## Notes

- **Portability:** the core is markdown + cooperative gates, so it runs on any
  markdown-capable agent. Optional enforcement (hooks) and delegation are
  per-harness layers added later — they don't change the install above.
- **Updating:** if you symlinked, `git pull` in `/path/to/specforge` updates the
  skills in place. If you copied, re-copy after pulling.
