[← Back to README](../README.md)

# What each harness actually guarantees

SpecForge's promise is *governance*, but governance has layers, and not every
harness can host every layer. This page is the honest degradation table: what
you keep and what you lose per setup. No setup silently pretends to guarantee
more than it can.

## The four enforcement layers

| Layer | Mechanism | Lives in |
|-------|-----------|----------|
| **L1 — Write denial** | the hook denies `Write`/`Edit`/Bash writes to SpecForge state | harness hook |
| **L2 — CLI refusal** | `sf` refuses illegal operations (verdict preconditions, broken ledger, transitions) | the `sf` binary |
| **L3 — Tamper evidence** | hash-sealed gates + hash-chained ledger; forgery is visible to every command | the state format itself |
| **L4 — Outside verification** | `sf verify` in CI: ledger + trace + schemas + freshness on every PR | your CI |

L2–L4 work **everywhere** — they don't depend on the harness at all. Only L1
(stopping the write *before it happens*) needs hook support.

## Per-harness table

| Setup | L1 write denial | Context injection | Journal nudge | Net guarantee |
|-------|-----------------|-------------------|---------------|---------------|
| **Claude Code** (`sf hook --harness=claude-code`) | ✅ PreToolUse deny (Write/Edit/Bash) | ✅ SessionStart + per-prompt breadcrumb + post-compaction re-ground | ✅ Stop hook | Full: prevention + detection + CI |
| **pi / opencode** (generic adapter) | ✅ if the harness exposes a pre-tool event | ✅ session/prompt events | ⚠️ depends on a stop event | Near-full; check your adapter's event coverage |
| **Any agent, no hooks** | ❌ direct writes possible | ❌ (agent must run `sf context` itself) | ❌ | Detection-only: forged state is caught by the next `sf` command (L3) and by CI (L4) — it cannot pass a verdict or archive |
| **A human with an editor** | ❌ | — | — | Same as above: you *can* edit `feature.json` by hand; `sf next`/`verify` will refuse to build on it |

## The design consequence

Prevention (L1) is deliberately **not** the load-bearing layer — it raises the
cost of the *cheap* shortcut for a lazy model, which is the real adversary. The
guarantees that matter are the ones nobody can switch off:

- a gate entry seals the **hash of what was approved** — silent edits flip the
  artifact to `stale` (L3);
- every ledger entry chains to the previous one — inserting or rewriting
  history breaks the chain visibly (L3);
- the verdict gate demands a **fresh green run sealed by the CLI**, with
  per-test causality when a structured report is configured (L2);
- `sf verify` re-checks all of it on a machine you don't control (L4).

So the honest one-line summary per tier: **with hooks, illegal writes don't
happen; without hooks, they happen and then fail to matter** — they're detected
at the next command and can never cross a gate.

If you're evaluating SpecForge for a team: wire L4 first (`sf verify
--init-ci`). It's the layer that works for every member regardless of their
editor, harness, or discipline.
