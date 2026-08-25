<p align="center">
  <img src="assets/specforge-logo.png" alt="SpecForge" width="380" />
</p>

# SpecForge

**SpecForge doesn't tell you how to work. It runs how you already work.**

`sf` is a state machine that knows **where you are**, **what comes next**, and **whether you're
allowed to advance**. Your agent asks it, does the work, and asks again.

**[Install](INSTALL.md)** · **[Docs](docs/)** · **[First run, end to end](docs/primeros-pasos.md)** ·
**[Commands](docs/comandos.md)** · **License:** [MIT](LICENSE)

> The guides are in Spanish, like `sf`'s own output and the design notes in `specs/`.

---

## The idea in one line

> `sf` exposes the state machine that guides the harness — and it is the arbiter that decides
> whether you can move forward.

Two verbs, and the second is the one that matters:

| | What it does | Without it, it would be… |
|---|---|---|
| **exposes** | *"you're at step 13, now comes 14"* | a to-do list |
| **checks** | *"you didn't finish: the branch is missing"* | a suggestion the agent can ignore |

`sf` doesn't drive the car — it doesn't take the wheel or pick the route — **but it gives green
or red, and on red you don't pass.** Its power is exactly one thing: it's the only one that can
move the state, and it only moves it when it checked for itself.

## The split

```
sf       →  WHERE you are · WHAT's next · CAN you advance?
skills   →  HOW each step is done
harness  →  DOES it
you      →  DECIDE   (three points, and only three)
```

**The orchestrator is the agent, not the CLI.** `sf` starts, answers, and dies — it lasts
milliseconds. A dead program can't launch anyone; it has no hands. The only thing alive for the
whole session is the agent, so the agent is the one that launches.

That's why `CLAUDE.md` is **four lines**:

```
1.  Run `sf next`.
2.  Do what it says, with the skill and model it says:
      via: vos        → you work, in the open
      via: subagente  → you launch a fresh subagent
      via: consola    → you shell out with the `comando:` it gave you
3.  When control comes back, run `sf next` again.
    DON'T read what the worker returned — the state is the truth.
4.  If `sf` says 🛑 or ⏸, show the user and wait.
```

**`AGENTS.md` is the same file, not a translation.** The only thing that changes between
harnesses is how a subagent is launched, and the harness already knows how to do that.

## The three hard rules

1. **`sf` never launches anyone.** If it spawned, it would manage context and tool-calling —
   and it would *be* a harness, reinventing what Claude Code and Codex already do well.
2. **State advances on checked facts, never on the worker's word.** The subagent says *"done"*;
   `sf` **doesn't believe it**: it runs the tests itself, checks the file exists, checks there's
   a branch. If something's missing, the state **doesn't move**.
3. **State lives in the repo, not in the chat.** A versioned `estado.json` — because when you
   hand off to another model, the one implementing wasn't in the conversation.

## Getting started

```bash
cd sf && go build -o ~/go/bin/sf ./cmd/sf
cp -r skills/* ~/.claude/skills/

cd <your project>
sf install    # CLAUDE.md + AGENTS.md · ~/.specforge/ with your models
sf init       # the scaffold: 2 directories, stack detection, empty state
sf next       # and from here on, the loop
```

Full instructions in **[INSTALL.md](INSTALL.md)**, and a complete walkthrough in
**[docs/primeros-pasos.md](docs/primeros-pasos.md)**.

## The nine states

Five run once per product, four run once per feature:

```
PRODUCT   brief → prd → constitución → backlog → roadmap
FEATURE   planificación → implementar → revisión → cierre
```

**Three of them stop for you**, and only three: sealing the brief, sealing the constitution, and
reviewing the plan. Everything else advances on its own.

Each state has a skill that knows how to do it, and `sf next` tells you which one — so the
`state → skill → model` table lives in one place instead of one copy per harness.

## Commands

```
sf init                       the scaffold
sf install · uninstall        the orchestrator + ~/.specforge/

sf next                       where you are · what's next · skill · model · via
sf context [--completo]       the envelope for the current state. No arguments
sf done [--msg "…"]           runs the gates and moves — or says what's missing
sf lote start                 creates the branch · demands the RED · saves the hash
sf new "…"                    puts a feature or a bug in the backlog

sf approve                    seals whatever you're looking at
sf reject "motivo"            doesn't seal, and keeps the reason for whoever redoes it
sf take <f-#>                 pulls the next one off the roadmap
sf model <name> [--via …]     raises the model — and declares it if it's new
sf dismiss <h-#> "motivo"     discards a review finding

sf status                     where everything is — the only one for humans
sf audit [f-# …]              end to end: many features against their stories
```

## What it actually prevents

Not by opinion — by arithmetic:

- **The commit doesn't get made** → closing a batch **is** committing. No `--msg`, no close.
- **Badly grouped commits** → the grouping was decided at planning time. One batch, one commit.
- **The branch never gets created** → `sf lote start` creates it and won't proceed without it.
- **"All green" with no tests** → `sf` has the exact list of tests that must exist, runs them
  itself, and **refuses to let you implement until it has seen them fail**.
- **A test loosened to pass** → the hash of the test files is taken at the red and compared at
  the green.
- **"Done" with half the story** → every acceptance criterion has an id. `sf` doesn't judge the
  review; it verifies **that the judgment happened, over all of them**.
- **Mocks at 50% context** → one fresh subagent per batch. It's the only one of these that can be
  prevented *before* the damage instead of detected after.

## What it does NOT promise

- **Code quality.** `sf` guarantees structure, sequence and evidence — not that the design is
  good. A gate stops on a **fact**; only a judge opines.
- **That the model won't hallucinate.** It produces artifacts that make it hallucinate *less*,
  and it checks the parts that are checkable.
- **Deciding for you.** The AI proposes; the last word is yours. Always. That's why an
  unapproved dependency **warns** instead of blocking — a tool that stops on its own breaks that
  rule.

## Layout

```
sf/          the binary — 18 packages, 277 tests, one dependency
skills/      18 skills: 9 for the states (sfp-* · sf-*) + 9 utilities (sfx-*)
specs/       the design, and why each decision is the way it is
```

**The prefix says something:** `sfp-` runs once per product, `sf-` once per feature, and `sfx-`
is a utility that lives **outside** the machine — standalone, usable in any project.

**And there's a smoke run behind a build tag** — `go test -tags e2e ./cmd/sf/` — that compiles the
binary and walks both loops end to end against a real repo. It's separate on purpose: the checks
that matter most here live in the **seam between commands**, and no package test can see them.

## Documentation

| | |
|---|---|
| [`INSTALL.md`](INSTALL.md) | install it |
| [`docs/primeros-pasos.md`](docs/primeros-pasos.md) | a complete run, end to end |
| [`docs/comandos.md`](docs/comandos.md) | the 15 commands |
| [`docs/estados.md`](docs/estados.md) | the 9 states and what each gate checks |
| [`docs/artefactos.md`](docs/artefactos.md) | every file and its shape |
| [`docs/skills.md`](docs/skills.md) | the 18 skills |
| [`docs/problemas.md`](docs/problemas.md) | what to do when `sf` stops you |

## The design

Everything in [`specs/`](specs/), and it is not decoration: each document says **why** and what
was rejected. Start with [`session.md`](specs/session.md), which points at the rest.
