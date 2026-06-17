# Lite lane

The fast lane for trivial, low-risk changes (F34). One combined artefact, one
spec gate, build, minimal check. It exists so the user has a *legitimate* fast
path instead of skipping the process behind the framework's back.

Lite is chosen at the lane gate in `sf-propose` Step 0 and recorded as
`"lane": "lite"` in `features.json`. If you arrived here, that gate passed.

## 1. Write the combined change artefact

Instead of three files, write one: `specforge/features/<name>/change.md`, using
`templates/change.tmpl.md`. It holds, on a single page:

- **What & why** — one paragraph.
- **Requirement** — one EARS line with an acceptance criterion (still testable —
  lite is not "untested").
- **Plan** — the handful of edits, as a short checklist.

Then:

```
───────────────────────────────────────
🔴 GATE — lite change: <feature>
Awaiting approval. Reply: approve / reject / change X
───────────────────────────────────────
```

Record the gate in `features.json` (`gates[]`, phase `change`). One gate replaces
the requirements/design/tasks trio — that is the whole point.

## 2. Build

Run the build straight from `change.md` (no separate `tasks.md`). Lite is almost
always a single wave. Same hard rules apply: verify each edit landed, a failed
operation halts and waits.

### Escape hatch — only upward (F34)

If the change turns out bigger or riskier than triaged (it sprawls across files,
touches a sensitive surface, or stops being trivially reversible), **promote to
standard**: stop, set `"lane": "standard"` in `features.json`, and run the normal
`sf-propose` flow (requirements → design → tasks) for what remains. Cheap can
become expensive; expensive never silently degrades to cheap. There is no
downward escape hatch.

## 3. Minimal check (still real, still traceable)

Run `sf-check` in minimal form — the matrix is trivial (one requirement → one
change → one test), but it is still built, and it still emits
`specforge/features/<name>/trace.json` (F33). Lite changes stay inside drift
detection; skipping the trace would let small specs rot unseen.

Verdict gate as usual → on APPROVE, archive (the lite file set is just
`change.md` + `review.md` + `trace.json` + the wave log).

## What lite never does

- Never touches a security/auth/data/money/migration surface — that is always
  standard.
- Never skips the test or the `trace.json` — lite is lighter ceremony, not
  lower integrity.
- Never auto-selects itself — the lane gate is human-approved.
