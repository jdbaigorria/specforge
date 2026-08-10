# Sesión — refundación de SpecForge

**Fecha:** 2026-08-07 / 08 / 10 · **Branch:** `refundation` (sin pushear)
**Estado:** el relevamiento está **cerrado**. La pregunta de fondo está **contestada**
(*ejecuta*, §5) y **la arquitectura está decidida** (§6). Sigue abierto el strawman de los
ocho dolores, en §5.

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

## 5. El marco de los ocho dolores — sigue abierto

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

### La pregunta de fondo — ✅ CONTESTADA: **ejecuta**

> **¿La herramienta EJECUTA tu flujo, o lo REEMPLAZA?**

**Ejecuta.** El flujo ya funciona; no hace falta que nadie proponga un método, hace falta que
alguien corra el tuyo. SpecForge deja de ser *"te dice cómo trabajar"* y pasa a ser
*"ejecuta cómo trabajás vos"*.

**Dos argumentos, y el segundo no es de gusto sino de física:**

1. *"Reemplazar"* obliga a la herramienta a tener razón sobre cómo se desarrolla software, y
   cada caso nuevo pide un concepto nuevo. Eso explica por qué todo lo anterior se infló.
2. **Multi-harness casi lo decide solo.** Un método vive en prompts, y los prompts son
   distintos en cada harness. Una máquina de estados vive en un programa, y un programa corre
   igual en todos lados. *(Y tu flujo ya cruza harnesses por su cuenta: ⑱ y ㉑ son traspasos
   a otro modelo.)*

**Consecuencia inmediata:** SpecForge tiene que ser **un binario**, no un skill ni un prompt.
Lo único que todos los harness saben hacer igual es correr un comando.

---

## 6. La arquitectura — decidida

Salió de acá, y el diseño es de Javier: **el orquestador es el agente, no el CLI.**

Razón dura: **`sf` arranca, contesta y se muere.** Dura milisegundos. Un programa muerto no
puede invocar un skill — no tiene manos. El único vivo durante toda la sesión es el agente.

### Las cuatro piezas

| Pieza | Qué sabe | Qué **no** hace |
|---|---|---|
| **`CLAUDE.md` / `AGENTS.md`** — el orquestador | qué skill invocar, cuándo lanzar un subagente | **no se sabe el flujo de memoria** — lo pregunta cada vez |
| **`sf`** (el binario) | **dónde estás** (`estado.json`) y **qué sigue**; y comprueba | no piensa · no lanza a nadie · no sabe *cómo* se hace un paso |
| **los skills** (`.md`) | **cómo** se hace cada paso: leé la constitución, el `us-#`, generá esto | no sabe en qué paso está parado |
| **el agente y sus subagentes** | hacen el trabajo | — |

**Los skills son agnósticos** (son un `.md`, cualquiera lo lee) y **conservan todo el método
que ya tienen**. Sólo se les agrega el principio y el final: preguntá dónde estás, avisá que
terminaste.

### El bucle

```
ORQUESTADOR
   ├─ sf: ¿qué sigue?              → "implementar"
   ├─ lanza subagente implementador
   │     └─ trabaja · termina · sf done · muere
   ├─ vuelve el control
   ├─ sf: ¿qué sigue?              → "check"
   ├─ lanza subagente check … vuelve
   ├─ sf: ¿qué sigue?              → "audit"
   └─ … hasta que sf diga "PARÁ, esto lo decide Javier"
```

### En una línea

> **`sf` es una tool que expone la máquina de estados que guía al harness — y es el árbitro
> que decide si se puede avanzar.**

Son **dos** verbos, y el segundo es el que importa:

| | Qué hace | Sin esto sería… |
|---|---|---|
| **expone** | *"estás en el ⑬, ahora toca el ⑭"* | una lista de tareas |
| **comprueba** | *"no terminaste: falta la branch"* | una sugerencia que el agente puede ignorar |

`sf` **no maneja el auto** —no agarra el volante ni elige la ruta—, pero **da verde o rojo, y
en rojo no se pasa**. Su poder es uno solo: **es el único que puede mover el estado, y sólo
lo mueve cuando lo comprobó él mismo.**

```
sf       →  DÓNDE estás · QUÉ sigue · ¿PODÉS avanzar?
skills   →  CÓMO se hace
harness  →  lo HACE
vos      →  DECIDÍS  (⑥, ⑧, ⑰)
```

### Las tres reglas duras

1. **`sf` nunca lanza a nadie.** El que lanza es el orquestador. Si `sf` spawneara, manejaría
   contexto y tool-calling, y **sería un harness** — reinventando lo que Claude Code y Codex
   ya hacen bien.
2. **El estado avanza con hechos comprobados, nunca con la palabra del que trabajó.** El
   subagente dice *"terminé"*; `sf` **no le cree**: corre los tests él, mira si existe el
   archivo, mira si hay branch. Si falta algo, el estado **no se mueve** y el orquestador
   recibe qué falta. *(Esto solo tacha los dolores 2, 3, 4 y 8.)*
3. **El estado vive en el repo, no en el chat.** Un `estado.json` versionado. Porque en ⑱
   cambiás de modelo y el que implementa no estuvo en la conversación.

### Por qué los subagentes, y no es un detalle técnico

**El subagente muere y se lleva su contexto.** Eso arregla dos cosas:

- **El orquestador nunca se llena.** No ve código ni tests: sólo *"terminó"* y *"ahora X"*.
  Aguanta los 23 pasos.
- **Mata el dolor #5.** Los mocks salían *"con el contexto al 50%"*; cada subagente **arranca
  de cero** y hace un lote. Era el único dolor atacable *antes* del daño, y el diseño lo
  ataca sin agregarle nada.

### Limitaciones registradas

- **Los subagentes son de Claude Code.** Sin ellos el diseño funciona igual, pero el contexto
  se llena. → **mejora, no requisito.**
- **Elegir modelo en runtime sí se puede, pero sólo entre modelos de Claude.** La herramienta
  de subagentes acepta un `model` que pisa el `model:` del frontmatter — alcanza para el ㉑
  (*"revisá con uno grande"*). **DeepSeek, Grok o Codex no pueden ser subagentes**: para esos
  el orquestador tiene que salir por consola (`deepseek exec "…"`). Sigue lanzando el
  orquestador, no `sf`.

---

### 📌 Idea guardada para más adelante — `sf` como proveedor de contexto

De Javier, explícitamente **no ahora**:

> *"el cli no sólo indicaría dónde estás, qué sigue, podés avanzar, sino que tendría la
> capacidad de inyectar el contexto adecuado a los subagentes. Porque cuando un subagente
> arranque le pide al cli: dame la constitución del proyecto y el spec."*

```
subagente:  sf context implement
sf:         → .docs/constitucion.md
            → .docs/features/us-3/spec.md
            → .docs/features/us-3/tareas.md (lote 2)
```

**Por qué encaja bien:** `flujo-real.md` ya dice que el ⑱ es el primer traspaso y que a
partir de ahí *"la propuesta tiene que bastarse sola"*. Esto sería **el mecanismo** de eso —
y de paso el subagente no gasta contexto buscando qué leer.

**Sigue sin romper ninguna regla:** servir archivos no es pensar ni lanzar a nadie.

---

## 7. Qué sigue

1. **⬅ ACÁ SEGUIMOS — revisar el flujo para formalizar los artefactos.** En qué puntos exactos
   se crea cada uno. La mayoría ya está en `flujo-real.md`; falta el ciclo por feature
   (⑪–㉓), que hoy no tiene artefactos declarados.
2. **Cerrar el strawman de §5** — las 8 filas, esperando corrección.
2. **La máquina de estados.** Los 23 pasos **no son 23 estados**. Hay que decidir dónde están
   los cortes, qué transiciones son automáticas, y **dónde para y te espera**.
   - Hipótesis: para en ⑥, ⑧ y ⑰ y en ningún otro lado → de empujar 23 pasos a decidir 3.
   - **El camino corto tiene que ser un estado, no una excepción**, o va a ser esquivado
     igual que ahora y ahí sí se pierde el rastro.
   - **Pregunta sin contestar:** cuando ㉑ o ㉒ encuentran algo y el implementador lo arregla
     solo, ¿te enterás? ¿O la máquina corrige en silencio si terminó bien?
3. **La forma del `estado.json`.** Qué guarda exactamente, y qué se registra de cada paso.
4. Recién después: qué de lo construido sobrevive. **En ese orden, no al revés.**

---

## 8. Pendiente de infraestructura

- La branch `refundation` **no está pusheada**.
- Siguen los ~155 commits viejos sin subir (`OPS-1`).
