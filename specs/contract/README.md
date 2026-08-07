# Contratos de SpecForge v2

Esta carpeta define **contratos**: qué garantiza cada pieza, qué no garantiza, y qué
tiene que ser verdad para que la garantía valga. Es la fuente de verdad del diseño de v2 —
si el código y el contrato no coinciden, el que está mal es el código.

## Índice

| Contrato | Cubre | Estado |
|---|---|---|
| [`audit.md`](audit.md) | **el brazo** — `sf audit`: los hechos deterministas, la clasificación y el material | borrador |
| [`judge.md`](judge.md) | **el cerebro** — contrato de salida de `sf-audit`: qué puede afirmar y cómo se lo verifica | borrador |

Los dos describen **un solo bucle**: pedir → trabajar → entregar → validar → sellar.

## Convenciones de estos documentos

**Nivel de obligación.** Se usa el vocabulario de RFC 2119 en castellano:

- **DEBE** / **NO DEBE** — requisito absoluto. Violarlo es un bug.
- **DEBERÍA** — recomendación fuerte; apartarse exige justificación escrita.
- **PUEDE** — opcional.

**Vocabulario del dominio.** Estos términos se usan con precisión en todos los contratos y
no son intercambiables:

| Término | Definición |
|---|---|
| **Cerebro** | El subagente de contexto limpio que revisa y firma el informe. En Forge es **`sf-audit`**. Sinónimo: *juez*. |
| **Brazo** | El CLI de Go. Ejecuta, computa, sella y entrega material. **Nunca opina.** En Forge es el comando **`sf audit`**. |
| **Hecho** | Afirmación producida por el brazo de forma determinista y reproducible. No depende de ningún LLM. Se puede volver a computar y da lo mismo. |
| **Juicio** | Afirmación producida por el cerebro sobre algo que ningún hecho puede decidir. Es falible por diseño, y **siempre lleva cita `path:line`**. |
| **Informe** | El entregable. Mezcla hechos (citados textuales) y juicios (con cita). Cada línea **DEBE** ser atribuible a uno de los dos canales. |
| **Sello** | Un hecho + el `tree_hash` del código sobre el que se computó. Un sello cuyo `tree_hash` ya no coincide con el árbol actual está **stale** y no vale. |
| **Residuo** | El conjunto de requisitos sobre los que los hechos **no alcanzan** para decidir. **Orienta** al cerebro sobre por dónde empezar; **no limita** qué puede mirar. |

> **Cuidado con `sf-audit` vs `sf audit`.** Con guión es el **skill** (cerebro); con espacio
> es el **comando** (brazo). Es una distinción normativa, no cosmética.

**Las dos reglas que ordenan todo:**

> **1.** El brazo ejecuta y computa, y nunca juzga. El cerebro juzga, y nunca ejecuta nada
> cuyo resultado vaya al informe como hecho. El brazo verifica la evidencia del cerebro.
>
> **2.** Si validar X exige **recomputar** X, X es del brazo. Si validar X es **resolver una
> cita**, X es del cerebro.

La segunda es la que decide cualquier responsabilidad nueva sin tener que volver a
discutirla.

## Por qué existen estos contratos

Salen del debate de refundación (2026-08-06/07, ver `planning/AUDITORIA-FEATURES.es.md`
y `planning/SPECFORGE-V2-INSPIRACION.es.md`). La pregunta que los originó fue:

> *"¿Qué tendría que decirte `sf audit` para que aceptes no leer el código?"*

Respuesta de Javier: *"Que no haya mentido: tests verdes reales, trace completo, sin
drift, y verificar que lo implementado cubre el requerimiento. Porque de qué serviría
usar IA para implementar si tengo que sentarme a ver qué hizo."*

Estos contratos son esa respuesta, hecha especificación.
