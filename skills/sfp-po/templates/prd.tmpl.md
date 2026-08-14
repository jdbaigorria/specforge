# <producto> — PRD

<Qué es, en una oración. Sale del brief y no se reescribe "mejor".>

## Actores

| Actor | Rol | Qué puede hacer |
|---|---|---|
| <nombre del rol> | <qué es en el sistema> | <su autoridad, en una línea> |

> Son **roles**, no personas. El ⑨ escribe `Como **<quién>**…` desde esta lista.

## Capacidades

Qué tiene que **poder hacer** el producto. Una por línea, en lenguaje llano.

- **C-1** — <capacidad>
- **C-2** — <capacidad>
- **C-3** — <capacidad>

<Si una capacidad se apoya en una afirmación que el brief marcó `model-prior`,
decilo acá: "C-4 — … ⚠ apoyada en evidencia sin verificar".>

## Restricciones no funcionales

Cada una escrita de forma que **se pueda medir**.

| # | Restricción | Cómo se mide |
|---|---|---|
| RNF-1 | <p. ej. latencia bajo 200ms en p95> | <la medición concreta> |
| RNF-2 | <p. ej. los datos en reposo van cifrados> | <la verificación concreta> |

## Alcance

**Adentro:**
- <lo que entra>

**Afuera — y es explícito:**
- <lo que NO entra, y por qué>

> Las dos mitades son obligatorias. El no-alcance es lo que frena el arrastre de
> alcance tres estados más adelante.

## Dependencias externas

| Qué | Para qué | Riesgo si no está |
|---|---|---|
| <servicio / API / dato / cuenta> | <para qué se usa> | <qué se cae> |

## Preguntas abiertas

<Lo que hacía falta y el brief no traía. Se pregunta, no se inventa.
Si no hay ninguna, borrá la sección — no dejes "N/A".>

- <pregunta> — **impacto:** <qué decisión depende de esto>
