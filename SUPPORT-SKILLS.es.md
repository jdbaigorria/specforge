<p align="center">
  <img src="assets/specforge-support-logo.png" alt="SpecForge" width="380" />
</p>

# SpecForge — Skills de Soporte

Skills independientes que complementan el pipeline de SpecForge. Cada uno funciona por su cuenta — ninguno requiere que `specforge/` esté inicializado. Producen artefactos en `specforge/context/` o interactúan de forma conversacional.

---

## Resumen de Skills

**El modo se decide por interactividad, no por tier.** Un sub-agente corre en un
contexto aislado y no puede frenar a preguntarle al usuario, así que cualquier
skill con gate o ida y vuelta debe correr **inline**. Los skills que son
transforms puros (entra X, sale Y, sin turno humano en el medio) se marcan
**delegado**: *pueden* delegarse a un sub-agente donde el harness lo soporte, con
fallback a inline. La delegación es una optimización opcional y per-harness — el
default portable es siempre inline.

| Skill | Modo | Propósito | Artefacto |
|-------|------|-----------|-----------|
| [sfx-think](#sfx-think) | inline | Debatir una idea, explorar opciones, llegar a una conclusión documentada | `specforge/context/thinks/{slug}.md` |
| [sfx-journal](#sfx-journal) | inline | Capturar aprendizajes anclados a evidencia, consolidar, proponer backprop | `specforge/context/journal/{date}.md` + `specforge/learnings.md` |
| [sfx-triage](#sfx-triage) | delegado | Investigar bugs, encontrar root cause, plan de fix | `specforge/context/triages/{slug}.md` |
| [sfx-documenter](#sfx-documenter) | delegado | Generar docs exhaustivos del código con ejemplos | `docs/` o inline |
| [sfx-explain](#sfx-explain) | delegado | Enseñar conceptos con método Feynman | `specforge/context/explanations/{slug}.md` (opcional) |
| [sfx-aws-architect](#sfx-aws-architect) | delegado | Diseñar infraestructura AWS con tradeoffs | `specforge/context/architectures/{slug}.md` |
| [sfx-data-engineer](#sfx-data-engineer) | delegado | Diseñar pipelines de datos con quality gates | `specforge/context/data-designs/{slug}.md` |
| [sfx-grill-me](#sfx-grill-me) | inline | Stress-test de un plan mediante entrevista implacable | `specforge/context/grills/{slug}.md` (opcional) |
| [sfx-tdd](#sfx-tdd) | inline | Implementar código con disciplina Red-Green-Refactor | código + tests |
| [sfx-github](#sfx-github) | delegado | Ejecutar workflow git: branch, commit, PR, merge | estado git |

---

## sfx-think

Debatir una idea, explorar opciones y llegar a una conclusión documentada. El
espacio entre "tengo una idea vaga" y "estoy listo para especificar" — no es
propose (sin requirements/tasks), no es explain (no enseña), no es grill-me (no
hace stress-test de un plan existente).

**Triggers:** `/sfx-think`, `/sfx-think <tema>`, "pensemos en", "uso X o Y", "estoy considerando", "pros y contras de", "evaluá este enfoque"

**Flujo:**
1. Identificar el tipo de pensamiento (decisión, exploración, validación, estrategia)
2. Debatir: abogado del diablo, ofrecer alternativas no consideradas, aterrizar en específicos
3. Converger cuando la dirección está clara (5-10 intercambios es el punto justo)
4. Resumir la conclusión, luego escribir el artefacto

**Reglas clave:**
- Desafiar la inclinación inicial del usuario — stress-testearla, no solo asentir
- Ofrecer al menos una alternativa que el usuario no consideró
- Aterrizar el debate abstracto en específicos concretos del contexto del usuario
- "Todavía no sabemos lo suficiente" es una conclusión válida — documentar qué falta para decidir

**Output:** `specforge/context/thinks/{slug}.md` — tema, opciones con pros/contras, la conclusión con rationale, alternativas rechazadas, próximos pasos. Por default se guarda (a diferencia de explain/grill-me).

---

## sfx-journal

Convertir lo que realmente pasó en una sesión en conocimiento durable y curado —
sin que crezca en ruido. Tres niveles: journal (crudo, por sesión) → consolidar
(deduplicado, chico) → promover (gateado, a la constitución).

**Triggers:** `/sfx-journal`, `/sfx-journal consolidate`, "journaleá esto", "capturá lo que aprendimos", "qué salió mal", "registrá esta lección", "consolidá aprendizajes"

**Flujo:**
1. Capturar — solo observaciones ancladas a evidencia (gate rechazado, error→fix, corrección del usuario, error repetido). Sin evidencia → sin entrada.
2. Juzgar — cada candidato debe estar anclado, ser generalizable y accionable; delegar el juez a un sub-agente fresco donde se soporte. Descartar el resto.
3. Escribir la entrada cruda en `specforge/context/journal/{date}.md` con `[[wikilinks]]`.
4. Consolidar en `specforge/learnings.md` (chico, curado, inyectado cada sesión): primera vez = nota, recurrente (~3×) = candidato a promoción, dedup siempre.
5. Promover — patrones recurrentes propuestos como invariantes de la constitución en un gate 🔴 (backprop, nunca automático).

**Reglas clave:**
- Evidencia o no pasó — anclar cada nota a un evento real.
- "Qué salió mal + cómo se resolvió" vale más que una lista simétrica bien/mal.
- `learnings.md` es curado y chico (se inyecta cada sesión); el firehose queda en `journal/`.
- Aprendizajes de proyecto vs hábitos meta-agente son capas distintas — no mezclar.
- Markdown es el source of truth; ICM es un motor opcional de recall/consolidación, nunca una segunda verdad.

**Output:** `specforge/context/journal/{date}.md` (crudo, vault compatible con Obsidian vía wikilinks) + `specforge/learnings.md` (consolidado). Las promociones van a `constitution.md` tras el gate.

**Referencias:** `references/consolidation.md` (rúbrica del juez, reglas de consolidación, capas, vault).

---

## sfx-triage

Investigar un bug sistemáticamente. Nada de fixes sin entender primero la causa raíz.

**Triggers:** `/sfx-triage`, `/sfx-triage <descripción>`, "hay un bug", "esto está roto", "por qué falla esto"

**Flujo:**
1. Recopilar síntomas (observado vs esperado, pasos de reproducción, frecuencia)
2. Explorar codebase (trazar data flow hacia atrás desde el síntoma)
3. Formar 2-3 hipótesis por probabilidad, cada una con evidencia
4. Verificar hipótesis (circuit breaker: máximo 3 — si todas se rechazan, marcar como no resuelto)
5. Escribir artefacto con diagnóstico, plan de fix, y caso de prueba

**Reglas clave:**
- Nunca aplicar un fix durante triage — solo investigación
- Cada hipótesis necesita evidencia, no suposiciones
- Root cause encontrada → escribir el test que falla ANTES de recomendar el fix
- No resuelto es un resultado válido — documentar qué se probó y qué queda pendiente

**Output:** `specforge/context/triages/{slug}.md` — síntomas, traza de investigación, hipótesis, root cause, plan de fix con test TDD, riesgos.

**Referencias:** `references/investigation.md` (metodología), `references/fix-plan.md` (estrategia de fix TDD)

**Siguiente paso:** Aplicar fix directo (si es trivial) o `sf-propose fix-{slug}` (si no lo es)

---

## sfx-documenter

Generar documentación exhaustiva del código. Cada función tiene un ejemplo. Cada edge case queda documentado.

**Triggers:** `/sfx-documenter`, "documentá esto", "generá docs", "docs de API", "guía de cómo hacer", "creá un README para"

**Modos:**
- `/sfx-documenter <path>` — Documentar un archivo o módulo específico
- `/sfx-documenter --api <path>` — Referencia de API (firmas, parámetros, retornos, ejemplos)
- `/sfx-documenter --guide <tema>` — Guía how-to (narrativa, paso a paso)
- `/sfx-documenter --project` — Docs completos del proyecto (README + arquitectura + API + guías)

**Reglas clave:**
- Cada función/clase/endpoint público tiene un ejemplo ejecutable — sin excepción
- Extraer ejemplos de tests cuando sea posible — ya están verificados
- Documentar edge cases, errores, y side effects — no solo el happy path
- Omitir secciones que dirían "N/A" — sin ruido
- Los ejemplos respetan el estilo del proyecto

**Output:**
- Archivo único → inline o `docs/{modulo}.md`
- Proyecto → directorio `docs/` (README, api/, guides/, architecture.md)

**Referencias:** `references/api-docs.md` (formato de referencia), `references/guide-docs.md` (formato tutorial), `references/project-docs.md` (estructura de proyecto)

---

## sfx-explain

Explicar cualquier concepto usando el método Feynman. Lenguaje simple, analogías de la vida cotidiana, ejemplos concretos.

**Triggers:** `/sfx-explain`, `/sfx-explain <tema>`, "cómo funciona X", "no entiendo", "enseñame", "explicame"

**Flujo:**
1. Idea central en una oración (sin jargon)
2. Analogía de la vida cotidiana (no de otros conceptos técnicos)
3. Construir capa por capa (qué → por qué → cómo → cuándo)
4. Ejemplo concreto (código real y ejecutable en el stack del usuario)
5. Tradeoffs (qué ganás, qué perdés, cuándo te juega en contra)
6. Resumen en una línea

**Reglas clave:**
- Sin jargon sin definirlo inmediatamente
- Analogías de la vida cotidiana — "como un contenedor de envío" no "como una VM"
- Siempre nombrar tradeoffs — "no tiene desventajas" significa que no lo entendiste bien
- Los ejemplos de código deben ser ejecutables, no pseudocódigo

**Output:** Conversacional. Opcionalmente guardado en `specforge/context/explanations/{slug}.md` si el usuario lo pide.

**Referencias:** `references/feynman-method.md` (metodología detallada con ejemplos y anti-patrones)


---

## sfx-aws-architect

Diseñar infraestructura AWS evaluada con el Well-Architected Framework. Cada servicio justificado con tradeoffs y costo.

**Triggers:** `/sfx-aws-architect`, "diseñá la infra", "cómo deployar", "qué servicios de AWS", "arquitectura para"

**Flujo:**
1. Clarificar requerimientos (workload, escala, presupuesto, compliance, equipo)
2. Diseñar con tradeoffs explícitos por servicio (qué, por qué no alternativas, costo, blast radius, scaling)
3. Generar documento de arquitectura

**Reglas clave:**
- Siempre estimar costos — rangos, no "depende"
- Arquitectura más simple primero — complejidad solo cuando se justifica
- Nunca recomendar un servicio sin explicar por qué no uno más simple
- Seguridad nunca es opcional — IAM, encryption, network isolation siempre

**Output:** `specforge/context/architectures/{slug}.md` — requerimientos, servicios con justificación, data flow, seguridad, scaling, desglose de costos, riesgos, log de decisiones.

**Template:** `templates/architecture.tmpl.md`

---

## sfx-data-engineer

Diseñar pipelines de datos con quality gates, idempotencia y observabilidad incluidos.

**Triggers:** `/sfx-data-engineer`, "pipeline de datos", "ETL", "modelo de datos", "diseño de schema", "calidad de datos"

**Flujo:**
1. Entender los datos (origen, destino, transformaciones, calidad, freshness)
2. Diseñar stages del pipeline (extract → validate → transform → load → verify)
3. Generar documento del pipeline

**Reglas clave:**
- Todo pipeline debe ser idempotente y re-ejecutable
- Quality gate antes de cargar — datos malos nunca llegan a los consumidores
- Manejo de errores definido por stage, no solo "va a fallar"
- Observabilidad no es opcional — métricas, alertas, lineage siempre
- Si el volumen no justifica streaming, usar batch

**Output:** `specforge/context/data-designs/{slug}.md` — data contract, schema, stages, quality gates, idempotencia, backfill, observabilidad, diseño de storage.

**Template:** `templates/pipeline.tmpl.md`

---

## sfx-grill-me

Stress-test de un plan, diseño o decisión a través de entrevista implacable. Recorrer cada rama del árbol de decisiones.

**Triggers:** `/sfx-grill-me`, `/sfx-grill-me <artefacto>`, "grill me", "hacele stress-test a este plan", "buscale agujeros"

**Flujo:**
1. Leer el target (artefacto, tema, o plan pegado)
2. Preguntar si guardar antes de arrancar
3. Explorar contexto (código, specs, docs) para no preguntar cosas que puede averiguar solo
4. Entrevistar una pregunta a la vez, cada una con respuesta recomendada
5. Trackear decisiones tomadas e issues abiertos
6. Generar resumen

**Reglas clave:**
- Una pregunta a la vez — nunca agrupar
- Siempre dar tu respuesta recomendada (el usuario acepta con "sí" o debate)
- Si se puede responder leyendo código, leerlo en vez de preguntar
- 3 "no sé" consecutivos → pausar, sugerir investigar primero
- Nunca sermonear — extraer el pensamiento del usuario, no enseñar

**Output:** Resumen siempre retornado. Transcripción completa guardada opcionalmente en `specforge/context/grills/{slug}.md`.

---

## sfx-tdd

Implementar código con Test-Driven Development estricto. Un test que falla, una implementación mínima, un ciclo.

**Triggers:** `/sfx-tdd`, `/sfx-tdd <tarea>`, "tdd esto", "usá TDD", "red-green-refactor"

**Flujo:**
1. Plan: identificar cambios de interfaz, listar 3-7 comportamientos a testear, verificar testeabilidad
2. Por cada comportamiento: RED (un test que falla) → GREEN (implementación mínima) → REFACTOR (solo cuando todo está verde)
3. Trackear tabla de evidencia (comportamiento → archivo de test → archivo de impl → refactor)

**Reglas clave:**
- UN test a la vez — slicing vertical, nunca horizontal
- El test DEBE fallar antes de implementar — si pasa, algo está mal
- Tests verifican comportamiento por interfaces públicas, no detalles de implementación
- Sin código especulativo — si ningún test lo demanda, no lo escribas
- Nunca refactorear en RED

**Output:** Código + tests + tabla de evidencia de ciclos TDD.

---

## sfx-github

Ejecutar workflow git estandarizado. Branch, commit, PR, merge — limpio y silencioso.

**Triggers:** `/sfx-github`, `/git`, "creá un PR", "pusheá esto", "commiteá", "mergeá"

**Convenciones:**
- Branches: `feature/`, `fix/`, `refactor/`, `docs/`, `chore/`
- Commits: formato convencional — `type(scope): description`
- Todos los cambios vía PR. Squash merge. Borrar branch después de merge.

**Reglas clave:**
- Nunca commitear directo a main
- Si CI falla, reportar — nunca auto-mergear
- Un PR = un cambio lógico

**Output:** Estado git (branches, commits, PRs). Retorna solo resumen limpio — sin ruido de comandos.
