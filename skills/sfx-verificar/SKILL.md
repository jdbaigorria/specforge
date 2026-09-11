---
name: sfx-verificar
description: >
  The verification primitive. Generate a project-local skill that drives the REAL app the way a
  user does — launch it, exercise a feature, capture evidence — plus a feature map of what else
  there is to drive. Any language, framework or platform. Triggers: "/sfx-verificar", "make a
  verification skill", "how do we prove the app actually works", "there's no way to test this by
  hand", or a review that can only say "the tests pass". Standalone: it works in any repo, with
  or without SpecForge.
---

# sfx-verificar

**A green suite proves the tests pass. It does not prove the app runs.**

That gap has a name in the flow: the ㉑ scores every criterion on a five-rung ladder, and rung 5
is *"you reproduced it in the running app"*. Without a written way to start the app and drive it,
**nobody can reach that rung** and the honest ceiling is 4.

This skill writes that way down, once, as `.docs/verificar/`.

```
.docs/verificar/
  SKILL.md          Launch · Doctor · Drive · Evidence · Cleanup
  features/         one file per user-facing feature — the map
  pruebas/          where the evidence lands
```

> **You are writing for an agent, not for a person.** It gets read cold, mid-task, by a fresh
> subagent that has never seen this app and cannot ask you anything. Every command is exact. No
> placeholders survive.

## 1. Interview the repo, not the user

Answer these from the code. Ask only what you genuinely cannot observe.

- **Surface.** What does a user actually touch? A web UI, a CLI or TUI, a desktop app, an HTTP
  API, a mobile app, a library. A repo can have several: pick the primary one, name the rest.
- **Launch.** How does it start locally? Prefer the repo's own documented command — package
  scripts, Makefile, README quickstart. Note ports, env vars, seed data, auth.
- **Drive.** How can an agent interact with it programmatically? **Existing harnesses first**:
  Playwright or Cypress specs, expect scripts, PTY helpers, curl-able endpoints, a debug port.
  Only then reach for a generic recipe — browser/CDP for web, a tmux or PTY session for CLI/TUI,
  plain HTTP for services.
- **Observe.** What evidence can be captured? Screenshots, terminal transcripts, response bodies,
  logs, exit codes, rows in a database.
- **Isolate.** Can two instances run side by side — ports, data dirs, profiles? If not, **say so
  in the generated skill**: refusing to double-drive a shared instance beats corrupting the
  user's running session.

Read `.docs/constitucion.md` first when there is one. `test_cmd`, `lenguaje` and the conventions
are facts about this project, and **you have no conventions of your own** (R4).

**If the checkout does not build or start as-is, fix that first, or report it precisely and
stop.** A verification skill written against a broken base teaches the wrong steps, and the next
agent has no way to tell.

## 2. Write the skill

`.docs/verificar/SKILL.md`, with frontmatter (`name: verificar-<app>` and a description naming
the app, the surface, and when to reach for it) and these five sections. **Each one grounded in
what the interview actually found.**

- **Launch.** The exact command, and how to tell it is ready — a log line, a port answering, a
  prompt appearing. Include teardown. For a short-lived CLI there is no server to keep alive:
  launch means build the binary once, then start each drive in its own isolated session.
- **Doctor.** One read-only check that answers *"is this instance worth driving?"* — process up,
  right build, port owned by us, auth valid. An agent runs this first whenever anything looks off.
- **Drive.** The harness recipe with **this repo's real selectors and commands**, not examples.
  Prefer stable handles — ARIA labels, data attributes, prompt strings, route paths — over
  coordinates and tab order.
- **Evidence.** What to capture and where it goes (`.docs/verificar/pruebas/`). State the proof
  standards:
  - exercise the real user path, never an internal setter or a test-only endpoint;
  - capture the action **and** the resulting state, not just the final screen;
  - check side effects — files written, rows inserted, messages sent — alongside what is visible;
  - mock only where a production boundary already isolates the external system.
- **Cleanup.** How to tear down what the run created. **Never kill by process name; kill what you
  started.** Cleanup removes instances and scratch state and **never the evidence** — the proofs
  outlive the teardown, at the path the skill names.

Any script the skill ships is executable, and its invocation is shown in the body. A helper the
reader has to reverse-engineer is not a helper.

## 3. Seed the feature map

`.docs/verificar/features/`, one file per user-facing feature, plus a `README.md` index. Start
with the top 3 to 5 you can identify from routes, commands, menus, or the docs.

Each file answers, **from the user's point of view**:

```markdown
# <feature>

## Qué es
## Cómo se llega (desde afuera, como un usuario)
## Cómo se maneja con el harness
## Qué estado final prueba que anda
## Trampas
```

> **The map is what stops a convenient proof.** A verification that drives the one entry point
> the reviewer had at hand, and declares the feature verified, is the evidenced version of the
> same *"it works"* the ㉑ exists to kill. The map is what says which other paths there are.

## 4. Prove it before handing it over

**Run your own instructions, end to end, once.** Launch, doctor, drive ONE mapped feature (one is
enough — the map exists so later runs cover the rest), capture the evidence, clean up.

Then, **after cleanup, confirm the evidence is still there.** A cleanup that eats its own proof
fails this step.

Fix what breaks and run it again, and **run the generated cleanup after every failed attempt
too**, so broken tries do not strand processes and ports.

> **A generated skill that was never executed is a draft, not a deliverable.** This step is the
> whole difference between this skill and writing a README about testing.

## 5. Say what it is worth

Close with the honest ceiling:

```
.docs/verificar/ listo — superficie: <cuál>
Probado end to end sobre: <la feature que manejaste>
En el mapa quedan <N> features sin manejar todavía.
El ㉑ ya puede declarar escalón 5 para <esa feature>. Para el resto, 4.
```

## Where it plugs into SpecForge, and where it does not

**It is a `sfx-`: standalone.** It does not ask for an envelope, it does not call `sf done`, and
`sf next` never returns its name. It works in any repo.

What SpecForge does with its output is serve it:

- **`sf context` for `revision`** adds `.docs/verificar/SKILL.md` and the whole feature map when
  they exist. Serving a file is not thinking and not launching anyone.
- **The ㉑ points `prueba` at the evidence** it left, and that is how a criterion earns rung 5.
- **`sfp-constitucion`** can then set `verificacion.escalon_minimo`. Do not set a floor of 5
  before this skill exists, or you are asking for evidence nobody can produce.

## Keeping the map honest

**A generated skill rots**, and a lying map is worse than no map: it sends the next agent to drive
a screen that no longer exists, and they conclude the app is broken.

The cheap answer is that `sf-cierre` already knows what a feature built. When a feature adds user
surface, its file joins the map at the ㉓ — the same move the documentation already makes.

**If the map has drifted badly, regenerate rather than patch**, and if nobody is maintaining it,
delete it. An unmaintained map is a rung-5 claim with nothing behind it.

## Rules

- Interview the repo. Ask the user only what the code cannot answer.
- Real selectors and real commands. A placeholder that ships is a bug.
- Exercise the user's path, never a test-only shortcut.
- Kill what you started, never by process name.
- Cleanup never eats the evidence.
- Run the generated skill once before handing it over. Never executed means draft.
- Say the honest ceiling: which features are drivable, and which are not yet.
