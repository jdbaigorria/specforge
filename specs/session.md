# Sesión — refundación de SpecForge

**Fecha:** 2026-08-07 / 08 / 10 / 11 / 12 / 13 · **Branch:** `refundation` (sin pushear)
**Estado:** el relevamiento está **cerrado**. La pregunta de fondo está **contestada**
(*ejecuta*, §5), **la arquitectura está decidida** (§6), la **prueba de escritorio de los
artefactos** quedó **✅ completa** (①–㉓, en [`artefactos.md`](artefactos.md)), el **strawman
de los ocho dolores está ✅ firmado** (§5), **la máquina de estados está ✅ cerrada** — los 9
estados, las compuertas y el `estado.json`, en [`maquina-estados.md`](maquina-estados.md) — y
**el recorrido de lo construido está ✅ cerrado**, con veredicto por pieza en
[`que-sobrevive.md`](que-sobrevive.md). Lo que sigue es **la superficie de `sf` y el reparto
orquestador ↔ skills** — el punto de retomada está al final, en *"por dónde arrancar la
próxima sesión"*.

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
| **`artefactos.md`** | qué produce cada paso, quién lo consume y en qué formato — 7 rondas, ①–㉓ | ✅ **completo** |
| **`maquina-estados.md`** | los 9 estados, las compuertas, las 4 paradas y el `estado.json` | ✅ **completo** |
| **`que-sobrevive.md`** | el recorrido de los 9 estados contra lo construido — veredicto por pieza | ✅ **completo** |
| **`anexo-determinismo.md`** | insumo de diseño: la cita de Uncle Bob contrastada contra los datos | completo |
| `contract/audit.md` · `judge.md` | la capa cross | ❌ **descartados** — son el contrato de `sf-audit`, y ningún estado lo consume |
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

## 5. El marco de los ocho dolores — ✅ FIRMADO (2026-08-12)

Se debatía **qué tiene que hacer la herramienta** sobre cada dolor. El marco: cada uno admite
**cuatro** respuestas, y elegir mal es lo que infla las herramientas.

> **hacerlo** · **avisar** · **impedir** · **nada** (se resuelve de otra forma)

**Javier lo firmó tal cual, sin cambiar ninguna fila:**

| Dolor | Respuesta | Por qué |
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

### 📌 ~~Idea guardada para más adelante~~ → ✅ **adoptada en la ronda 7** — `sf` como proveedor de contexto

Era de Javier y estaba marcada *"no ahora"*. **La ronda 7 la ascendió a mecanismo del ⑱** (ver
[`artefactos.md`](artefactos.md) §10). Queda acá el origen:

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

1. ~~La prueba de escritorio de los artefactos~~ — ✅ **cerrada**, las 7 rondas en
   [`artefactos.md`](artefactos.md). El inventario completo está en su §11.
2. ~~Cerrar el strawman de §5~~ — ✅ **firmado el 2026-08-12**, sin cambios.
3. ~~La máquina de estados~~ — ✅ **cerrada**, en
   [`maquina-estados.md`](maquina-estados.md).
4. ~~La forma del `estado.json`~~ — ✅ **cerrada**, en `maquina-estados.md` §9.
5. ~~Qué de lo construido sobrevive~~ — ✅ **cerrado**, en
   [`que-sobrevive.md`](que-sobrevive.md).
6. **⬅ ACÁ SEGUIMOS — la superficie de `sf` y el reparto orquestador ↔ skills.**

---

## 8. La máquina de estados — cerrada

Vive en [`maquina-estados.md`](maquina-estados.md). **Acá no se copia: se apunta.**

### El saldo

| | |
|---|---|
| **23 pasos** | → **9 estados** — 5 de producto, 4 de feature |
| **3 decisiones** | ⑥ · ⑧ · ⑰ — el resto avanza solo |
| **4 tipos de parada** | apareció **ME TRABÉ**, y es la única no configurable |
| **1 config que no hizo falta** | el modo de orden sale de las tres puertas del ⑰ |
| **0 máquinas en paralelo** | es una cola: `feature_actual` alcanza |

Y **`sf` sigue sin pensar en ningún lado**: todas las compuertas son correr un comando, contar
un campo o comparar dos strings.

### Las tres correcciones de Javier, y las tres mejoraron el diseño

- **No partir el bloque de planificación (⑫–⑯).** El strawman lo cortaba en tres para que
  ningún subagente se llene. El argumento no se sostiene: **el dolor #5 es daño de código, no
  de planificación**, y ahí el contexto grande es un **activo** — el ⑬ aprovecha acordarse de
  las dos opciones que descartó. Además el ⑰ revisa el bloque entero y *"pido cambios"* vuelve
  al bloque entero: **nadie entra ni sale por el medio.**
- **De ahí salió la separación que destrabó todo: `un estado ≠ un subagente`.** Un estado
  puede tener checkpoints internos, y `sf` los deduce de **qué archivos existen** — sin un
  campo extra en el `estado.json`.
- **"Varias features" era el ORDEN, no concurrencia.** Javier elige si implementa apenas
  planifica una o si planifica todas primero. Eso ya vive en las tres puertas del ⑰ → **la
  config `modo` se cayó sola.**

### Los tres hallazgos propios de la ronda

- **El cuarto tipo de parada.** Si `sf` da rojo, el orquestador relanza — y un bucle sin freno
  se cuelga. El freno ya estaba en el ⑳ (*"si falla varias veces, entro yo"*) y el contador ya
  estaba en el borrador; sólo faltaba nombrarlo. **No es configurable: es de seguridad.**
- **El plan que envejece.** Planificar varias features por adelantado deja specs mirando un
  repo que ya cambió — y el implementador que no encuentra lo que la spec dice **improvisa**,
  que es donde nacen los mocks. `sf` guarda un `base_commit` y compara: dos strings y un `if`.
  **Avisa, no frena.**
- **`verde` era redundante con `commit`.** Si `sf` no deja commitear en rojo, *hay commit* ya
  significa *estaba verde*. Un campo deducible de otro es un campo que se desincroniza.

---

## 9. ⬅ ACÁ QUEDAMOS — el recorrido de lo construido cerró

Vive en [`que-sobrevive.md`](que-sobrevive.md). **Acá no se copia: se apunta.**

### El saldo

```
skills   15  →  sobreviven 12 (adaptados o tal cual) · sf-audit sale a utilitario ·
                se suma sfp-po · sf-propose se parte en tres estados
CLI      47  →  ~10 de máquina + 3 de andamio.  ~34 se tiran, incluidos los tres
                más grandes: gate (35K) · hook (35K) · doctor --drift (23K)
capa cross   →  DESCARTADA — es el contrato de sf-audit, y ningún estado lo consume
huecos        →  5, y cuatro son chicos. El grande: roadmap.json no existe en el CLI
```

**El diagnóstico de la refundación queda confirmado por los números.** Dos tercios del CLI
sostenían conceptos —ledger de gates, enforcement por hook, cadena de cuatro fases, MoSCoW,
carril lite, traza `PR#`— que **la máquina vuelve innecesarios**. Y lo que sobrevive tiene una
forma clara: **corre algo, cuenta algo o sirve un archivo. Nada de lo que sobrevive piensa.**

### Los rescates que valieron la ronda

- **`redwitness` ya estaba construido**, con el mejor texto del repo sobre el dolor #5. Y la
  compuerta **activa** de la máquina resultó **más simple** que su registro pasivo: `sf` ve el
  rojo y el verde con sus propios ojos, así que el `CodeHash` sobra.
- **`sf context for-wave` es el mecanismo del ⑱**, ya escrito.
- **El reparto del juez de fase** (`sf` junta material → LLM juzga en subagente fresco → `sf`
  persiste el veredicto) **es la arquitectura de §6, escrita antes de decidirla.**
- **`sf mutation` ya era un contador, no un generador** — hacía shell-out a la herramienta del
  lenguaje y consumía sólo el exit code.

### Las siete reglas nuevas

Están en `que-sobrevive.md` §2. La más productiva es de Javier:

> **`sf` hace todo lo que tiene una sola respuesta correcta; el LLM hace lo que tiene varias.**

Le agrega el cuarto verbo a `sf` — *expone · comprueba · sirve · **hace*** — y no rompe ninguna
regla dura: crear directorios no es pensar ni lanzar a nadie.

---

## ⏭ POR DÓNDE ARRANCAR LA PRÓXIMA SESIÓN

**Queda UNA ronda de diseño.** Después de ella el diseño está completo y empieza otra cosa.

### Lo que hay que leer para retomar (y nada más)

| Archivo | Para qué |
|---|---|
| **este `session.md`, §6** | la arquitectura y las tres reglas duras |
| **[`maquina-estados.md`](maquina-estados.md)** | los 9 estados y las compuertas |
| **[`que-sobrevive.md`](que-sobrevive.md)** §2, §13 y §14 | las reglas nuevas, los huecos y el saldo |

### ~~Paso 0 — las tres correcciones a `artefactos.md`~~ → ✅ **hechas el 2026-08-13**

Estaban listadas en [`que-sobrevive.md`](que-sobrevive.md) §15 y ya están aplicadas; el resumen
quedó en [`artefactos.md`](artefactos.md) **§13**. El saldo:

- el **`estado:` del `us-#` salió** — el avance vive en `estado.json`, por feature, y el
  roadmap, el backlog y el índice son **vistas que genera `sf`**. Ningún `.md` se reescribe.
- el **㉓ agrega dos archivos**: la doc (con su mitad funcional, que no sale del código) **y el
  journal**. El `us-#` gana ahí un consumidor.
- los dos campos nuevos: **`test_cmd:`** en la constitución y **`hash_tests`** por lote.

**Y aparecieron dos incoherencias más, del mismo tipo, al tocar esos párrafos:** el
`estado.json` de `artefactos.md` §11 era el **borrador viejo** (`paso`, `verde`,
`modelo_recomendado`) — ahora **apunta** a `maquina-estados.md` §9 en vez de copiarlo —, y
`hash_tests` se agregó también ahí, que es donde el `estado.json` está cerrado.

### ⬅ Paso 1 — la ronda: la superficie de `sf` **y** el reparto, juntos

`que-sobrevive.md` §16 los lista como dos temas. **Son uno solo**, y conviene tratarlos así:

> Saber **qué comandos existen** y saber **quién los llama** se contesta con el mismo dato.

**El método es el que ya funcionó cuatro veces acá: la prueba de escritorio.** Así salieron los
artefactos (7 rondas) y así salió la máquina. Ahora se aplica **al bucle**: recorrer una vuelta
completa de feature, del ⑪ al ㉓, anotando **cada llamada**.

```
ORQUESTADOR  sf next                      → "planificar f-2"
             lanza subagente Opus
   SUBAGENTE   sf context planificar f-2  → constitución · us-# · aprendizajes
               ...trabaja...
               sf save tareas --feature=f-2
               sf done                    → ✓ / ✗ y qué falta
             vuelve el control
ORQUESTADOR  sf next                      → "🛑 PARÁ: el ⑰ lo decide Javier"
```

**De ese trazado caen las dos cosas solas:** el inventario **es lo que aparezca**, y el reparto
**es quién lo llamó**. Sin inventar comandos por las dudas — que es exactamente como se infló la
versión anterior.

**Dos cabos sueltos que la prueba va a levantar igual:**

- `sf lote start` y `sf lote done` aparecieron en el ⑦ y no están en ningún inventario.
- El ⑱ es traspaso a otro modelo: hay que ver **qué pasa cuando el que corre `sf` no es un
  subagente de Claude** sino `deepseek exec "…"` — el orquestador sigue lanzando, pero el sobre
  del `sf context` es lo único que viaja.

### Después de esa ronda: la construcción, y es una decisión de otro tipo

**No tratarla ahora.** Cuando llegue, hay dos preguntas y la primera ya tiene respuesta:

1. **Por dónde arrancar** — `roadmap.json` + `estado.json` + `sf next` son **la columna**.
   Nada funciona sin eso, y el `roadmap.json` es el hueco grande (`que-sobrevive.md` §13).
2. **Podar el CLI actual, o partir de cero con lo que sobrevive.** Sobreviven ~10 comandos de
   47, así que la respuesta no es obvia: podar deja 63 archivos de test que cubren código que
   se va, y partir de cero tira `redwitness`, `check run` y `context for-wave`, que están bien.

### Y sigue aparcado

**Brownfield** — el debate de `que-sobrevive.md` §5. Decide el destino de `sf onboard scan` y
de la detección de `sf-init` Step 2. **No bloquea nada de lo de arriba.**

> **Y la advertencia que sigue vigente** (`decisions-specforge`, 2026-08-07): no sobre-indexar
> en determinismo. `sf` nació para producir artefactos útiles que hagan alucinar menos al
> modelo, no para garantizar lo que no se puede garantizar.

---

## 10. Pendiente de infraestructura

- La branch `refundation` **no está pusheada**.
- Siguen los ~155 commits viejos sin subir (`OPS-1`).
