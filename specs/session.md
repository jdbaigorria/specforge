# Sesión — refundación de SpecForge

**Fecha:** 2026-08-07 → 14 · **Branch:** `refundation` (sin pushear)
**Estado:** el relevamiento está **cerrado**. La pregunta de fondo está **contestada**
(*ejecuta*, §5), **la arquitectura está decidida** (§6), la **prueba de escritorio de los
artefactos** quedó **✅ completa** (①–㉓, en [`artefactos.md`](artefactos.md)), el **strawman
de los ocho dolores está ✅ firmado** (§5), **la máquina de estados está ✅ cerrada** — los 9
estados, las compuertas y el `estado.json`, en [`maquina-estados.md`](maquina-estados.md) — y
**el recorrido de lo construido está ✅ cerrado**, con veredicto por pieza en
[`que-sobrevive.md`](que-sobrevive.md), y **la superficie de `sf` y el reparto están ✅
cerrados** — 10 comandos y el bucle de 4 líneas, en
[`superficie-sf.md`](superficie-sf.md).

> **El diseño está completo Y la columna del binario está construida.** `sf/` tiene los 10
> comandos del inventario **más `sf init`**, 173 tests verdes y una sola dependencia. **Los nueve
> skills están ✅ escritos** y **el orquestador ✅ también** — cuatro líneas, en
> [`plantillas/CLAUDE.md`](../plantillas/CLAUDE.md).
>
> **El bucle ya cierra de punta a punta:** `sf init` → `sf next` → skill → `sf done` → `sf next`.
> Lo que queda del diseño es **uno solo**: el mapa de modelos (`via: consola`). El punto de
> retomada está al final.

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
| **`superficie-sf.md`** | la prueba de escritorio del bucle — los 10 comandos y el reparto | ✅ **completo** |
| **`construccion.md`** | por dónde se empieza, qué se migra y con qué criterio · **+ los 7 pasos y lo que apareció construyendo** | ✅ **completo** |
| **`skills.md`** | el mapa `estado → skill` firme, los 15 veredictos y la cirugía común | ✅ **completo** |
| **`skills/`** | los 9 de estado escritos + 9 utilitarios · `sf-propose`, `sf-amend` y `sf-init` borrados | ✅ **escritos** |
| **`plantillas/CLAUDE.md`** | el orquestador — cuatro líneas. Un archivo, dos destinos | ✅ **completo** |
| **`sf/`** | el binario: 12 paquetes, 162 tests, los 10 comandos | ✅ **anda** |
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
6. ~~La superficie de `sf` y el reparto~~ — ✅ **cerrada**, en
   [`superficie-sf.md`](superficie-sf.md). **Con esto el diseño está completo.**
7. ~~La construcción del binario~~ — ✅ **los 7 pasos hechos**, en `sf/`. Ver §11.
8. ~~El mapeo `estado → skill`~~ — ✅ **cerrado el 2026-08-14**, en [`skills.md`](skills.md).
   Los nueve nombres firmes, ya renombrados en `maquina.go` y con test.
9. ~~Escribir los nueve skills~~ — ✅ **hechos**, en `skills/`.
10. ~~El `CLAUDE.md`, `sf init` y el hueco 6~~ — ✅ **los tres cerrados**. Con eso **el bucle
    corre de punta a punta desde un directorio vacío.**

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

## 9. El recorrido de lo construido cerró

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

## 10. La superficie de `sf` y el reparto cerraron

Vive en [`superficie-sf.md`](superficie-sf.md). **Acá no se copia: se apunta.**

### El saldo

```
comandos   10 de máquina  +  sf status (el único para humanos)  +  el andamio
reparto     7 los llama el orquestador · 3 el que trabaja · cero solapamiento
CLAUDE.md   4 líneas — y AGENTS.md es el MISMO texto, no una traducción
```

**El método fue la prueba de escritorio del bucle**, la quinta vez que se usa acá: cuatro
trazados —`planificacion`, `implementar`, `revision`+`cierre`, y los cinco de producto—
anotando cada llamada. **El inventario es lo que apareció**, sin inventar nada por las dudas.

### Los cuatro hallazgos que valieron la ronda

- **`sf next` devuelve el skill, el modelo y el `via`.** Es lo que deja `CLAUDE.md` en cuatro
  líneas — si no, la tabla estado→skill→modelo vive en el orquestador y hay que mantener una
  copia por harness.
- **Y el argumento que lo cerró es de Javier:** *"si corre en Claude Code sabe que no puede
  usar un modelo fuera de Anthropic; si corre en otro harness, sabe que puede cambiar entre
  proveedores"*. `sf` es el único que ve las dos mitades —lo que el paso pide y lo que el
  harness puede—, y con eso **el ⑱ deja de ser un caso especial: es la misma llamada con otro
  `via:`**. Era el último cabo suelto.
- **El commit lo hace `sf`, el mensaje lo trae el que trabajó.** Sin `--msg` no hay `done`, así
  que **cerrar el lote *es* commitear**. Los dolores #2 y #3 dejan de ser detectables para ser
  imposibles.
- **Los hallazgos de la revisión no tienen ciclo de vida.** Como la revisión se rehace entera,
  `arreglado` es *no reaparecer*. Nadie marca nada, y queda un solo estado real:
  `descartado`, que es de Javier.

### Los cinco comandos que se cayeron, y ninguno se perdió

`lote done` era `done` · `save` no hacía falta · `check run` es trabajo del que trabaja ·
`mutation` es una línea del sobre · `feature archive` es `approve`.

> **Cuando dos cosas parecen distintas y caen en el mismo lugar del trazado, es que eran una.**
> Mismo hallazgo que `un estado ≠ un subagente` en la ronda de la máquina.

---

## 11. ⬅ ACÁ QUEDAMOS — el binario está construido

El plan y el saldo viven en [`construccion.md`](construccion.md). **Acá no se copia: se apunta.**

### El saldo

```
sf/          12 paquetes · 161 tests verdes · go vet limpio
comandos     los 10 del inventario, andando y probados a mano contra repos git reales
dependencias UNA: gopkg.in/yaml.v3
migrado      NADA del CLI viejo
```

```
sf/
  cmd/sf/                el ruteo de los 10 comandos
  internal/estado/       el estado.json — el único que sf escribe
  internal/roadmap/      el orden de las features (sólo lectura)
  internal/maquina/      next · done · lote start · las 5 paradas · las entradas
  internal/compuerta/    lo que frena, y por qué
  internal/sobre/        sf context
  internal/suite/        correr los tests · ver si existen
  internal/git/          shell-out al binario git
  internal/vista/        sf status — el único para humanos
  internal/docs/ constitucion/ historia/ tareas/ revision/ frontmatter/
```

### El hallazgo de la construcción, y contradice lo que se esperaba

**No se migró NADA del CLI viejo.** La regla de `construccion.md` §2 —*se migra cuando un
comando lo necesita para andar*— resultó **más filosa de lo previsto**: al llegar cada comando,
escribir la pieza contra el diseño nuevo salió más corto que adaptar la vieja.

> `redwitness.go` eran 8.4K. Se volvió `EmpezarLote` + `internal/suite`, comentarios incluidos —
> porque la máquina ya aportaba la mitad: el estado, los lotes y la lista de tests planificados.

**No es que el código viejo fuera malo: sostenía conceptos que la máquina volvió innecesarios.**
La regla funcionó; su respuesta, casi siempre, fue *"no"*.

### Las siete cosas que aparecieron construyendo

Ninguna sale de los documentos. Están con su porqué en `construccion.md` §7; las tres que más
cambiaron el diseño:

- **`backlog_visto`** — las 🛑 dejan rastro solas (un veredicto, un bool, un cambio de estado).
  La ⏸ del ⑨ es un enter y **no sella nada**: sin campo, `sf next` la repetía para siempre.
  Su hermana la ⏸ del ㉓ no lo necesita, porque ahí `sf approve` archiva y eso sí deja huella.
- **`rechazo`** en producto y en feature — el motivo del `reject` no se imprime y ya: el que
  rehace es un subagente **nuevo**, y sin el motivo vuelve a proponer lo mismo. Viaja primero
  en el sobre.
- **`Movio` vs `Cambio`** — un `sf done` que **falla** igual incrementa `intentos_fallidos`.
  Sin la distinción, el contador de ME TRABÉ nunca subía y el bucle no tenía freno.

### Lo que se probó a mano, y no sólo con tests

La compuerta del rojo, entera, contra un proyecto Go real:

```
① el test planificado no existe        ✗ y dice cuál
② el test existe pero YA PASA          ✗ "un test que pasa antes de que exista
                                           el código es un test de mentira"
③ el test falla de verdad              ✓ crea la branch · rojo · guarda el hash
④ "arreglar" aflojando el test         ✗ "los archivos de test CAMBIARON entre
                                           el rojo y el verde"
⑤ el test intacto + el código real     ✓ lote cerrado en 8bbf5a2
```

Y la vuelta completa del ciclo: brief → approve → prd → constitución → backlog ⏸ → roadmap →
take → reject (el motivo viaja) → cierre → approve, que **archiva la carpeta, mergea `--no-ff`
y borra la branch**.

---

## ⏭ POR DÓNDE ARRANCAR LA PRÓXIMA SESIÓN

**El diseño está completo y el binario anda.** Lo que falta ya no es `sf`: es lo que lo rodea.

### Lo que hay que leer para retomar (y nada más)

| Archivo | Para qué |
|---|---|
| **este `session.md`, §6 y §11** | la arquitectura, y qué está construido |
| **[`superficie-sf.md`](superficie-sf.md)** §5 y §6 | los 10 comandos y **el reparto** — es lo que viene |
| **[`construccion.md`](construccion.md)** §7 y §9 | lo que apareció construyendo, y lo que falta |
| `sf/internal/maquina/maquina.go` | el mapa `estado → skill`, con los ⚠ de lo provisional |

**Probar el binario primero cuesta un minuto y orienta todo lo demás:**

```bash
cd sf && go build -o /tmp/sf ./cmd/sf && cd <un proyecto> && /tmp/sf next
```

### Paso 1 — los skills · ✅ **CERRADO** — mapeados y escritos

Vive en [`skills.md`](skills.md). **Acá no se copia: se apunta.**

**Los nueve están escritos** (2026-08-14), con el `name:` igual a su carpeta y verificados contra
el mapa. Dos nuevos (`sfp-po`, `sf-cierre`), tres reescritos, cuatro adaptados. Se borraron
`sf-propose`, `sf-amend` y `sf-init` —el reemplazo ya existía— y `sf-audit` pasó a `sfx-audit`.

**El mapeo cerró el 2026-08-14** y lo que encontró no eran los tres ⚠: la **regla de prefijos**
—`sfp-` producto · `sf-` feature · `sfx-` utilitario, y **`sf` no conoce a los `sfx-`**— dejaba
**seis de nueve entradas mal**. Tres apuntaban a un utilitario desde la máquina, y eso rompe en
producción sin hacer ruido: **un `sfx-` no llama a `sf done`, y el estado no se mueve nunca.**

```
PRODUCTO   brief sfp-scout · prd sfp-po · constitucion sfp-constitucion ·
           backlog sfp-backlog · roadmap sfp-roadmap
FEATURE    planificacion sf-plan · implementar sf-build · revision sf-check ·
           cierre sf-cierre
```

**El mapa está renombrado en `maquina.go`, con un test que fija la regla** (se verificó que falla
si se vuelve a poner un `sfx-`).

**El patrón que lo destrabó ya estaba inventado:** el skill de estado es **delgado y compone
utilitarios**, igual que `sfp-scout` componía `sfx-think` y `sfx-grill-me`. Así `sf next` sigue
devolviendo **un** skill y **los `sfx-` no se tocan** — si aprendieran a llamar a `sf done`
dejarían de servir fuera de un proyecto SpecForge, que es la mitad de su valor.

La cirugía, la misma para los nueve:

```
+  arranca con   sf context      pedí tu sobre, no busques qué leer
+  termina con   sf done         avisá que terminaste
−  las convenciones propias      salen de la constitución            (R4)
−  el modo / carril / fase       la máquina saltea estados           (R4)
```

> **Y los skills no saben en qué estado están** — por eso `sf context` no lleva argumentos. El
> skill dice *"dame mi sobre"*; cuál es el sobre lo decide `sf`.

**Dos cosas más que cerró la ronda:**

- **`sf-amend` se tira.** Su punto de entrada ya existe y se llama `sf new` — un bug sobre una
  feature archivada **no la desarchiva**: entra al backlog como `us-#` con `tipo: bug`. Y su
  premisa era la equivocada: **una spec archivada no es la doc del sistema, es el registro de
  una decisión con fecha.**
- **Apareció el hueco 6, y ✅ se cerró** (`que-sobrevive.md` §13): `relacionado_a` estaba
  declarado y **no lo leía nadie**, y el sobre **no tocaba `.docs/archivado/`**. El que arregla el
  bug no veía la spec de lo que rompió — el dolor #5 esperando. Es `sobre.loQueRompio`, cuatro
  tests.

### Paso 2 — el `CLAUDE.md` de cuatro líneas · ✅ **CERRADO**

Vive en [`plantillas/CLAUDE.md`](../plantillas/CLAUDE.md). **Un archivo, dos destinos:**
`sf install` lo va a copiar como `CLAUDE.md` **y** como `AGENTS.md` — mantenerlo como dos
archivos acá sería el mismo contenido en dos lugares, que es lo que el proyecto evita en todo lo
demás.

Y el `AGENT.md` de la raíz —321 líneas del orquestador viejo, que nombra skills borrados y decide
*inline vs delegate* por su cuenta— quedó **marcado con un banner**, no borrado: lo referencian
`README*.md`, `INSTALL.md` y `docs/`, así que se apaga con ellos (§12).

### Paso 3 — `sf init` · ✅ **CERRADO**

Vive en `sf/internal/arranque/`. **Con esto el bucle arranca de cero.** Probado a mano:

```
sf next  en un proyecto vacío   → "corré sf init"          (exit 2)
sf init                         → 2 dirs · constitución · estado.json
                                  go (go.mod) · test_cmd: go test ./...
sf init  otra vez               → "ya está iniciado"       (exit 2, no pisa nada)
sf next                         → estado: brief · skill: sfp-scout · via: vos
```

**Tres decisiones que valen más que el código:**

- **Escribe la constitución A MEDIAS, y es el reparto de R1 hecho archivo.** *"Hay un `go.mod`"*
  tiene una sola respuesta correcta → `sf`. *"¿Qué arquitectura?"* tiene varias → el ⑧. Esto es
  lo que hace verdadera la promesa de `sfp-constitucion` (*"la cabecera ya está llena"*).
- **Dos directorios, no siete.** `features/` y `archivado/` los crean el ⑫ y `sf approve` cuando
  hacen falta. **Un directorio vacío desde el día uno es una promesa que el proyecto todavía no
  puede cumplir:** el que lo abre no sabe si está roto o si no llegó.
- **La tabla de detección es una LISTA, no un mapa** — y hay un test que lo prueba 20 veces. Un
  proyecto con dos manifiestos (Python + un `package.json` para el front) daría un lenguaje
  distinto en cada corrida: los mapas de Go se recorren en orden aleatorio.

Y no pisa nada: ni el `estado.json` (perderías los sellos del ⑥ y del ⑧, que no se deducen de
ningún archivo) ni una constitución que ya exista.

### Paso 4 — el mapa de modelos, que cierra H1b ⬅ **ES LO ÚLTIMO DEL DISEÑO**

`~/.specforge/`, con qué modelos hay y cómo se invoca cada uno. Es lo único que le falta al
campo `via:` para resolver `consola` — hoy sólo resuelve `vos` y `subagente`, y está marcado con
⚠ en `maquina.go`.

**Y el mecanismo ya está decidido:** la lista **no se escribe de antemano, se construye sola con
cada aprobación**, igual que `dependencias_aprobadas`. Es un patrón ya firmado, no uno nuevo.

**Y arrastra un segundo pendiente**, que es el mismo ⚠: `sobre.mutantes()` devuelve
*"todavía no: falta leer `mutacion:` de la constitución"*. El ㉒ funciona igual —el modelo lee el
código— pero sin la corrida de la herramienta.

### Paso 5 — `sf install`, el andamio

`plantillas/CLAUDE.md` existe y **nadie lo copia**. `sf install` es lo que lo pone en un proyecto
—como `CLAUDE.md` y como `AGENTS.md`— y, según `que-sobrevive.md` §5, también instala la
constitución global de `~/.specforge/`. Es el mismo lugar donde va a vivir el mapa de modelos del
paso 4: **conviene hacerlos juntos.**

### Y sigue aparcado

**Brownfield** — el debate de `que-sobrevive.md` §5. Decide el destino de `sf onboard scan` y de
la detección de `sf-init` Step 2. **No bloquea nada.**

> Ojo con un detalle nuevo: `sf-init` **ya no existe como skill** (se borró con los nueve). Su
> `references/onboard.md` —el material de brownfield— **vive en la historia de git**, y es el
> insumo de este debate cuando se retome.

> **Y la advertencia que sigue vigente** (`decisions-specforge`, 2026-08-07): no sobre-indexar
> en determinismo. `sf` nació para producir artefactos útiles que hagan alucinar menos al
> modelo, no para garantizar lo que no se puede garantizar.

---

## 12. Pendiente de infraestructura

- La branch `refundation` **no está pusheada**.
- Siguen los ~155 commits viejos sin subir (`OPS-1`).
- **El producto viejo sigue entero, y se apaga TODO JUNTO.** No es sólo `cli/`: son
  `AGENT.md` (marcado con un banner), `README.md`, `README.es.md`, `INSTALL.md`, `docs/` y
  `examples/`. **Se referencian entre sí**, así que borrar uno solo deja enlaces rotos y ningún
  reemplazo. Mientras tanto `cli/` compila y corre, que es lo que permitió comparar contra él
  (`construccion.md` §3).
