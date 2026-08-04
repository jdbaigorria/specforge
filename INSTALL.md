# Installing SpecForge

SpecForge is a pack of agent skills plus an `AGENT.md` orchestrator. The skills
are plain markdown — any agent that can load skills from a directory can run
them. Installation is just making `skills/` (and `AGENT.md`) visible to your
agent.

The suite has three tiers:

- **Pipeline + audit** (`sf-init`, `sf-propose`, `sf-build`, `sf-check`,
  `sf-amend`, `sf-audit`) — the SpecForge SDD workflow.
- **Product, from-zero** (`sfp-scout`) — the optional discovery front-end. Needs
  the research MCPs (see `skills/sfp-scout/references/tooling.md`).
- **Support skills** (`sfx-triage`, `sfx-documenter`, `sfx-explain`, `sfx-think`,
  `sfx-grill-me`, `sfx-tdd`, `sfx-journal`, `sfx-github`) — standalone helpers
  that don't require `specforge/` to be initialized.
- **Community skills** (`skills-community/`: `sfx-aws-architect`,
  `sfx-data-engineer`) — domain-generic helpers kept OUTSIDE the maintained
  core (D7': less surface, more spine). Install by copying/symlinking the
  folder into your agent's skills directory if you want them.

You can install all of them or only the pipeline.

---

## 1. Clone the repo

```sh
git clone https://github.com/jdbaigorria/specforge.git
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

> **Plugin install (Claude Code).** This repo ships a plugin skeleton under
> `.claude-plugin/` (`plugin.json` + `marketplace.json`), so it can be added as a
> marketplace and installed directly:
>
> ```
> /plugin marketplace add jdbaigorria/specforge
> /plugin install specforge
> ```
>
> Validate the manifest with `/plugin` before publishing — the schema may evolve.

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
  per-harness layers — they don't change the install above.
- **Hard gates (optional, F29):** installing via the Claude Code plugin also
  activates the enforcement hooks in `hooks/` — `PreToolUse` denies skipping a
  gate or editing `specforge/.state/`, and `SessionStart` restores session +
  compact-rules context (including after compaction). The decision engine is the
  `sf` CLI itself (`sf hook`), so it needs the `sf` binary on PATH (no Python),
  fails open on error, and is a no-op outside a SpecForge project. Other harnesses
  drive the same engine through the portable contract in `hooks/README.md`.
- **Updating:** if you symlinked, `git pull` in `/path/to/specforge` updates the
  skills in place. If you copied, re-copy after pulling.
