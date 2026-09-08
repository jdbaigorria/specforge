---
name: sfx-grill-me
description: >
  The interview, standalone — no product state behind it, no artifact required. Use when the user
  wants to stress-test a plan, a design or an idea on its own. Triggers: "/grill-me",
  "/grill-me <artifact-path>", "grill me", "stress-test this plan", "interview me about",
  "poke holes in this", "challenge my design".
---

# sfx-grill-me

The interview with nothing behind it. **The method lives in `sfx-grilling`** — this skill only
figures out what is being grilled and whether to keep a record.

## 1. What is the target

- no argument → ask: *"¿qué querés poner a prueba?"*
- a path → read the artifact first
- a topic → ask for two or three sentences of context

Then read whatever context is at hand — `.docs/constitucion.md`, `.docs/vocabulario.md`, related
specs or code — **before** the first round. Every fact you can find yourself is a question you do
not spend on the user.

## 2. Keep a record?

> Esto puede durar un rato. ¿Guardo la entrevista en un archivo? [ruta / N]

Default: no. Remember the answer.

## 3. Run it

`Call the Skill tool with "sfx-grilling"`, telling it the target, the context you gathered, and the
file to write — or that there is none.

## 4. Close

```
## Grill terminado: {tema}

Rondas: {n} · preguntas: {n} · abiertas: {n}
Guardado: {ruta | no}

### Decisiones ({n})
- {una línea cada una}

### Lo que quedó abierto ({n})
- {una línea cada una, con por qué}

### Siguiente paso sugerido
{`sf new "…"`, más investigación, o "resolvé lo abierto primero"}
```

**If `abiertas` is not zero, say it out loud.** An interview cut short is a fine outcome; one that
pretends it finished is not.
