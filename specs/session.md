# Sesión — refundación de SpecForge

**Fecha:** 2026-08-07 / 08 · **Branch:** `refundation` (sin pushear)
**Estado:** el relevamiento está **cerrado**. Quedó abierta **una sola pregunta**, en §5.

> La sesión anterior —el contrato de la capa cross— quedó en
> [`session-capa-cross.md`](session-capa-cross.md). Está **congelada**, no descartada.

---

## 1. El corte de método

A mitad de camino Javier frenó todo con dos críticas, ambas correctas:

1. *"Siempre te forzaste en contrastar con lo ya creado."* Cada decisión se anclaba en la
   auditoría, en `sfp-scout`, en `sf-init`, en los 16 repos. **Eso moldea el diseño según lo
   que hay, no según lo que hace falta.**
2. Se había acordado escribir el flujo punta a punta **antes** de diseñar módulos, y no se
   hizo: se fue derecho a Inception.

**Decisión: foja cero.** Primero el flujo real de trabajo, sin SpecForge adentro. Recién
después, la herramienta.

**Regla que queda:** no anclar en lo construido ni en los repos hasta que el flujo esté
escrito y validado. Los repos se miran **con la necesidad ya escrita** — para diferenciarse,
no para surtirse.

---

## 2. Lo que se produjo

| Archivo | Qué es | Estado |
|---|---|---|
| **`flujo-real.md`** | el flujo real, 23 pasos, 3 entradas, sin herramienta adentro | **completo** |
| **`anexo-determinismo.md`** | insumo de diseño: la cita de Uncle Bob contrastada contra los datos | completo |
| `contract/audit.md` · `judge.md` | la capa cross | ⏸ **congelados** — hipótesis sin consumidor |
| `inception/brief-inception.md` · `flame-inception.md` | Spark y Flame caminados a mano | quedaron de la etapa anterior al corte |

---

## 3. Lo que sabemos del flujo

**Es el flujo NUEVO de Javier, ya rediseñado por él, corrido a mano, y funcionó.**
No hay que arreglarlo. El problema es **quién lo empuja**.

- **23 pasos**, **3 entradas** (producto nuevo · feature sobre lo existente · bug), y todas
  convergen en el backlog: **el backlog es el embudo**.
- **Sólo 3 puntos de decisión reales:** ⑥ el sello del brief · ⑧ la constitución · ⑰ la
  revisión del plan. Todo lo demás lo empuja sin decidir nada.
- **Existe un camino corto** que ya usa para arreglos chicos, con dos condiciones que
  mantiene: correr los tests y verificar que quedó resuelto.
- **Patrón constante:** la IA propone, la última palabra es de Javier. Siempre.

---

## 4. Los ocho dolores

| # | Duele | Cuándo |
|---|---|---|
| 1 | **lanzar cada fase a mano** | siempre — repetido 3 veces en la sesión |
| 2 | el commit no se hace, hay que vigilarlo | seguido |
| 3 | commits mal agrupados | seguido |
| 4 | no crea la branch por feature | seguido |
| 5 | métodos que son mocks | **con el contexto al 50%** |
| 6 | librerías fuera de la constitución | sin dueño identificado |
| 7 | *"terminado"* con media historia | **DeepSeek** |
| 8 | *"todo verde"* sin que haya tests | **Grok** |

Los cuatro primeros son de **operación** y pasan siempre. Del 5 al 8, tres tienen dueño o
condición conocida — eso cambia cómo se atacan.

---

## 5. ⬅ ACÁ QUEDAMOS — la pregunta abierta

Se estaba debatiendo **qué tiene que hacer la herramienta** sobre cada dolor. El marco: cada
uno admite **cuatro** respuestas, y elegir mal es lo que infla las herramientas.

> **hacerlo** · **avisar** · **impedir** · **nada** (se resuelve de otra forma)

**El strawman sobre la mesa, esperando que Javier lo corrija:**

| Dolor | Respuesta propuesta | Por qué |
|---|---|---|
| 1 lanzar cada fase | **hacerlo** | no hay nada que avisar ni impedir |
| 2 el commit | **hacerlo** | en un momento definido, no "cada tanto" |
| 3 agrupar commits | **hacerlo** | **es el mismo problema que el 2** |
| 4 la branch | **hacerlo** | mecánico, sin ambigüedad |
| 5 mocks a 50% de contexto | **anticipar** | no entra en el marco — ver abajo |
| 6 librerías fuera de la constitución | **avisar**, no impedir | ver abajo |
| 7 media historia | **avisar** | juicio puro |
| 8 verde sin tests | **impedir** | el más barato de los ocho |

**Tres cosas que salieron de aplicarlo:**

- **2 y 3 son un solo problema.** Si se sabe qué lote terminó, el agrupamiento sale solo: un
  commit por lote. No hay que agrupar bien, hay que commitear en el momento correcto.
- **El 5 no entra en el marco.** No se detecta ni se impide: **se anticipa**. Es el único de
  los ocho donde se puede actuar **antes** del daño.
- **El 6 va en avisar y no en impedir**, aunque impedirlo sería trivial: Javier dijo dos
  veces que la última palabra es suya. **Una herramienta que frena sola rompe esa regla.**

**Y el hallazgo grande:** ninguno de los ocho requiere que la herramienta **piense**. Todos
son ejecutar, comparar, o disparar a alguien que piense. El único que necesita juicio (el 7)
se resuelve **garantizando que el juicio ocurra**, no reemplazándolo.

### La pregunta de fondo, también abierta

> **¿La herramienta EJECUTA tu flujo, o lo REEMPLAZA?**

El flujo ya funciona. Si eso es cierto, la herramienta no propone una metodología: **corre la
tuya**. Deja de ser *"SpecForge te dice cómo trabajar"* y pasa a ser *"SpecForge ejecuta cómo
trabajás vos"*. Más chico, más honesto, y explicaría por qué todo lo anterior se infló.

---

## 6. Qué sigue, después de cerrar §5

1. **La máquina de estados.** Los 23 pasos **no son 23 estados**. Hay que decidir dónde están
   los cortes, qué transiciones son automáticas, y **dónde para y te espera**.
   - Hipótesis: para en ⑥, ⑧ y ⑰ y en ningún otro lado → de empujar 23 pasos a decidir 3.
   - **El camino corto tiene que ser un estado, no una excepción**, o va a ser esquivado
     igual que ahora y ahí sí se pierde el rastro.
   - **Pregunta sin contestar:** cuando ㉑ o ㉒ encuentran algo y el implementador lo arregla
     solo, ¿te enterás? ¿O la máquina corrige en silencio si terminó bien?
2. Recién después: qué de lo construido sobrevive. **En ese orden, no al revés.**

---

## 7. Pendiente de infraestructura

- La branch `refundation` **no está pusheada**.
- Siguen los ~155 commits viejos sin subir (`OPS-1`).
