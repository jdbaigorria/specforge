# Qué de lo construido sobrevive

**Fecha:** 2026-08-12 · **Método:** recorrido de los 9 estados de
[`maquina-estados.md`](maquina-estados.md) contra lo que ya existe en el repo.

> **Por qué recién ahora.** Es la regla del corte de método (`session.md` §1): *no anclar en
> lo construido hasta que el flujo esté escrito y validado*. La necesidad está escrita entera
> —flujo, artefactos, máquina— así que lo ya hecho se puede mirar sin que moldee el diseño.

**Entra:** los 9 estados y sus compuertas · el inventario del repo (15 skills, 47 comandos de
CLI, 18.086 líneas de Go sin tests, la capa cross congelada).
**Sale:** un veredicto por pieza, cinco huecos, siete reglas nuevas y tres correcciones a
`artefactos.md`.

---

## 1. La vara

Se recorren **los 9 estados**, y por cada uno una sola pregunta:

> **¿hay algo ya construido que haga esto? ¿sirve tal cual, sirve cambiado, o estorba?**

**Nunca al revés.** Abrir `skills/` y preguntar *"¿esto dónde encaja?"* es lo que hundió las
versiones anteriores: eso moldea el diseño según lo que hay.

```
SOBREVIVE     hace algo que la máquina necesita, tal como lo necesita
SE ADAPTA     el método adentro sirve, la cáscara no
SE TIRA       no tiene lugar en ninguno de los 9 estados
```

**El sesgo declarado es hacia `se tira`**: todo lo que sobreviva hay que mantenerlo, y el
diagnóstico de esta refundación fue que la herramienta se había inflado.

### La excepción que hubo que declarar: el andamio

La vara mataba `sf install` — ningún estado lo consume. Pero hay **dos familias**, y se juzgan
distinto:

| | Vara | Ejemplos |
|---|---|---|
| **comandos de la MÁQUINA** | ¿lo consume un estado? | `next` · `check` · `context` · `save` |
| **comandos del ANDAMIO** | ¿lo necesita el usuario para que la máquina exista? | `install` · `uninstall` · `lint` |

---

## 2. Las siete reglas que salieron del recorrido

Son el rendimiento real de la ronda: no dependen de ninguna pieza y se aplicaron varias veces
cada una.

### R1 — `sf` hace lo que tiene una sola respuesta correcta

*(de Javier, sobre el scaffold del ③)*

```
crear las carpetas · detectar el stack · contar tres opciones · servir un archivo
   → una sola respuesta correcta → sf
elegir el stack · escribir la spec · juzgar si satisface
   → hay varias → LLM (y Javier decide)
```

**No rompe ninguna regla dura:** crear directorios no es pensar y no es lanzar a nadie — la
misma defensa que ya se aceptó para `sf context`. Y gana dos cosas: **abarata el dolor #1**
(un subagente menos por cada cosa determinista) y **elimina un error silencioso** (un LLM que
crea 7 carpetas, a veces crea 6).

**Le agrega el cuarto verbo a `sf`:**

```
expone      dónde estás · qué sigue
comprueba   ¿podés avanzar?
sirve       sf context <estado>
hace        lo mecánico y sin ambigüedad     ← nuevo
```

### R2 — Un artefacto tiene el tamaño de sus consumidores

La vara para no inflar un documento no es el gusto: **si una sección no la lee nadie, sobra.**

- El `prd.md` lo leen **dos** (⑧ y ⑨) → ~5 bloques, no las 13 secciones de un PRD formal.
- La `constitucion.md` la lee un **subagente frío** → técnica, no filosófica.
- La doc del ㉓ la leen **dos lectores distintos** → dos mitades (§11).

Es la misma regla que ya se usó para tirar `verde` y `modelo_recomendado` del `estado.json`:
no discutir si es lindo, preguntar quién lo consume.

### R3 — Una compuerta frena sobre un hecho; un juez opina

> `sf` puede frenarte porque el test falló — eso no lo discute nadie.
> **No puede frenarte porque un modelo dijo que tu diseño está flojo.**

Es la regla de `session.md` §5 (*"una herramienta que frena sola rompe la regla de que la
última palabra es tuya"*) precisada. Mató la config `audit.phase: block` y obligó a agregar
`descartado` como estado de hallazgo (§10).

### R4 — El skill no tiene convenciones propias: lee la constitución

Donde el skill traía un default cableado, ahora hay un dato del proyecto.

```
sfx-github  "squash merge default"  →  git.merge de la constitución
sf-build    build.mode              →  un solo modo: un subagente por lote
sf-propose  lane: lite | standard   →  no existe: la máquina saltea estados
```

### R5 — Si no se puede escribir el test, no es un criterio de aceptación

Es una **intención**, y va a la sección *"Qué NO entra"* con su porqué. Mató el eje
`verification: test | benchmark | manual` y es coherente con la muerte de MoSCoW: *un criterio
que se puede no cumplir no es un criterio*.

### R6 — Un campo deducible de otro es un campo que se desincroniza

Ya existía (fue lo que mató `verde`). Se aplicó **tres veces más** en este recorrido:

```
estado:    del frontmatter del us-#     → sale de roadmap.json + estado.json
depends_on de tareas.json               → sólo servía para derivar el lote
CodeHash   del testigo rojo             → la compuerta activa lo hace innecesario
```

### R7 — Declarar el resultado o declarar la causa: gana el que se usa

Cuando dos campos dicen casi lo mismo, sobrevive **el que tiene consumidores**, no el que es
más elegante. `lote` se usa en tres lugares (commit, subagente, filtro de `sf context`);
`depends_on` en ninguno — sólo existía para calcular `lote`.

---

## 3. Estado ① — `brief` (①–⑤ + la parada ⑥)

`sfp-scout` no se parece al estado: **es el estado**, con los seis pasos en el mismo orden.

| Tu flujo | `sfp-scout` |
|---|---|
| ② pinponeo, Opus xhigh | Step 1 — capture the idea |
| ③ busco si ya existe | Step 2 — research vía MCPs |
| ④ ¿me sirve? ¿en qué me diferencio? | Step 2 — comparable GitHub repos |
| ⑤ ¿hay consenso? | Step 3 — de-risk, **componiendo think + grill-me** |
| ✔ el brief | Step 4 — discovery brief |
| ⑥ hacelo / pivoteá / no lo hagas | Step 5 — **proceed / pivot / kill** + gate |

No es casualidad: es el único skill escrito **después** de que Javier empezara a trabajar así.

**Lo que le sobra, y todo es de un problema que no existe acá:**

| Sobra | Por qué |
|---|---|
| los `PR#` y la traza `PR → roadmap → requisito → tarea → código → test` | la cadena arranca en `us-#`. Es un piso de más |
| `sf question add` | es de la **vía consultoría**: un cliente que tarda semanas. Acá el que contesta es Javier, y contesta ahora |
| el eje `ready / not ready` | por lo mismo. El ⑥ tiene **tres** salidas, no seis |
| `research.md` + `decision-log.md` | el inventario dice **un** archivo: `brief.md` |
| *"greenfield only"* | se cae solo: el estado ya es sólo la entrada A |

**Se le da vuelta el orden:** Step 1 dice *"don't over-interview here"*; el ② dice lo
contrario — el pinponeo es donde la idea toma forma, **antes** de buscar nada.

**Se le deja la procedencia,** y es lo mejor que aporta. Cada afirmación marcada `retrieved`
(con link) o `model-prior` (sin verificar). El ⑥ es el único punto del flujo donde una
alucinación cuesta el producto entero: sellar *"no lo hagas"* porque existe algo que no
existe.

### Los otros dos no se duplican: uno abre y otro cierra

```
sfx-think     ABRE    explora el espacio, trae alternativas que no pensaste   → el ②
sfx-grill-me  CIERRA  estresa lo que ya hay, una pregunta por vez             → el ⑤
```

`sfp-scout` ya los compone (*"not replaced — they stay standalone"*). **Divulgación
progresiva:** un skill delgado que invoca a los otros dos.

**Y `sfx-think` tiene un segundo lugar, que es un hueco: el ⑫.** Su plantilla —*opciones con
pros y contras · el argumento que inclinó la balanza · la conclusión · lo descartado y por
qué*— **es** `decision.md`.

```
sfp-scout      SE ADAPTA    delgado. −PR#, −question, −ready/not-ready, −research.md,
                            −decision-log.md, −"greenfield only". Se da vuelta ②↔③.
                            Se le deja la procedencia. Se le agrega el ④ completo:
                            diferenciarme Y surtirme
sfx-think      SOBREVIVE    utilitario · método del ② · **método del ⑫**
sfx-grill-me   SOBREVIVE    utilitario · método del ⑤
```

**Compuerta de salida:** `¿existe brief.md?`. El sello lo pone Javier — es 🛑.

### El ④ precisa la regla del corte de método

Javier: *"puede ser también para surtirse"*. La regla nunca fue *"no te surtas"*, era
**cuándo**:

> **Escribí primero qué necesitás; recién ahí mirá lo que hay.**

En el ④ la necesidad ya salió del ②. Es exactamente lo que hace este documento con el CLI.

---

## 4. Estado ② — `prd` (⑦)

**No hay nada.** `sf-init --from` sabe *importar* un PRD; nadie sabe *escribirlo*. Y `sfp-scout`
lo dejó afuera **a propósito**, con un argumento que era bueno para el modelo viejo:

> *"el PRD pesado sería opt-in… no debe volverse una tercera capa de traducción entre el brief
> y `sf-propose`"*

**Acá no es una capa de traducción: tiene dos consumidores reales** (`artefactos.md` §11) — el
⑧ y el ⑨. De ahí salen las historias.

```
sfp-po    SE CONSTRUYE    nuevo, delgado
```

**Por qué se gana un skill propio** (contra el sesgo a `se tira`):

1. **Divulgación progresiva** — el skill del `brief` no carga las instrucciones del PRD hasta
   que hacen falta.
2. **Cambió de dueño.** El `brief` lo conversa el orquestador de frente; el PRD **produce un
   artefacto → va a subagente fresco** (§3 de la máquina). Un subagente no hereda instrucciones
   que viven en la conversación de otro: **necesita su propio `.md`.** No es preferencia, es la
   regla de las dos clases de estado.

**El tamaño lo dan sus dos lectores** (R2):

```
entra   brief.md
sale    prd.md — actores · capacidades · restricciones no funcionales ·
                 alcance y no-alcance · dependencias externas
fuera   personas con foto · journeys · métricas de negocio · go-to-market ·
        pricing · wireframes        (no los lee ni el ⑧ ni el ⑨)
```

**Sin compuerta y sin parada:** el ⑦ no tiene punto de decisión.

**`prd_hash` confirmado:** es el mismo truco que `base_commit`. Si el PRD cambia después de que
se escribieron las US, esas US quedaron mirando otro producto. **Avisa, no frena.**

### El prefijo dejó de ser decorativo

```
sfp-*   los 5 estados de PRODUCTO    scout · po · constitución · backlog · roadmap
sf-*    los 4 estados de FEATURE     planificación · implementar · revisión · cierre
sfx-*   utilitarios                  fuera de la máquina — sf no los conoce
```

Ahora dice algo: **si el estado corre una vez o una vez por feature.**

---

## 5. Estado ③ — `constitucion` (⑧) 🛑

El estado bisagra: *"primer artefacto cuyo consumidor principal no es un humano ni el hilo de
conversación: **es un subagente frío**"*. Lo leen el ⑨, **el implementador ⑲ (es su manual)**,
el revisor ㉑ y `sf`. Su frontmatter es donde mueren los dolores **#4** (`git:`) y **#6**
(`manifiesto:` + `dependencias_aprobadas:`).

### El recorte más grande: tres archivos son uno

| `sf-init` genera | Contenido | Va a |
|---|---|---|
| `constitution.md` | visión, principios, anti-goals | *(ver abajo)* |
| `context/project.md` | **stack, arquitectura** | `## Arquitectura` · `## Stack y por qué` |
| `context/conventions.md` | **convenciones de código** | `## Convenciones de código` |

**Son la misma cosa partida en tres.** Tres archivos que mantener, tres que envejecen, y tres
que `sf context` tendría que servir para una sola pregunta.

### Y cambia de contenido, no sólo de forma

La de `sf-init` es filosófica; la nueva es técnica: *Arquitectura · Stack y por qué ·
Convenciones · Estructura de carpetas · Reglas de trabajo*. **Cero principios abstractos.**

**La razón es dura:** `sf-init` no tenía brief ni PRD antes, así que la constitución cargaba
también con el *por qué*. Ahora llega **tercera** — el qué y el porqué ya están escritos y
sellados arriba. Es R2.

### El choque `.json` → `.md`, resuelto

`constitution.json` trae `Principles · Constraints · AntiGoals · Invariants` **más cinco
bloques de config del verificador viejo** (`Audit · Build · Flow · Coverage · Verification`).
Con la regla de la ronda 7 —`.json` sólo donde `sf` cuenta— la constitución es **`.md` con
frontmatter**:

```
constitution.json + sf constitution render|validate   SE TIRAN
lo que sf necesita en su lugar                        un parser de frontmatter
```

> **La regla de la ronda 7 es el cuchillo más grande del recorrido.** Se lleva también
> `sf requirements`, `sf design`, el esquema de `sf review` y la mitad de `sf save`.

### Un agujero encontrado mirando el CLI: falta `test_cmd`

El frontmatter tiene `lenguaje`, `manifiesto`, `mutacion`, `dependencias_aprobadas`, `git`.
**Le falta el comando de tests** — y sin él la compuerta del medio de `implementar` no puede
correr. El preflight de `sf-build` ya lo sabía:

> *"si falta, ponelo ahora — el gate de veredicto requiere una corrida verde fresca, así que
> `sf check run` tiene que poder ejecutar los tests"*

Y sale detectado: **`detectStack()` de `sf init --minimal` llena `lenguaje`, `manifiesto` y
`test_cmd` solo.**

```
sf-init Step 3 (la conversación)      SE ADAPTA    el ⑧, con contenido técnico
project.md + conventions.md           SE FUSIONAN  → constitucion.md. 3 archivos → 1
Step 1 — el scaffold                  PASA A sf    determinista (R1). 2-3 dirs, no 7
detectStack()                         SOBREVIVE    lenguaje · manifiesto · test_cmd
compact-rules.md                      SE TIRA      lo reemplaza sf context
history.md · .state/session.md        SE TIRAN     lo reemplaza estado.json
constitution.json + sf constitution   SE TIRAN     regla ronda 7
sf install · sf uninstall             SOBREVIVEN   andamio, + las reglas globales
```

**La constitución global** (`~/.specforge/constitucion.yaml`, cada proyecto hereda y pisa) **la
instala `sf install`.** Es lo que le da al andamio una responsabilidad más.

**Compuerta de salida:** ninguna automática. `constitucion_sellada: true` lo pone Javier.

### ⏸ Aparcado: brownfield

`sf-init` Step 2 (detección greenfield/brownfield + monorepo) y **`sf onboard scan`** (el
inventario determinista: árbol, símbolos, mapa de tests) quedan **explícitamente en debate**.
El flujo escrito no lo cubre: las entradas B y C *"no pasan por brief, PRD ni constitución"* —
asumen que el producto ya pasó por A. **Un repo que ya existe y nunca vio SpecForge no tiene
entrada.** `sf onboard` es lo único del CLI que espera esa decisión.

---

## 6. Estado ④ — `backlog` (⑨) ⏸

*"La primera salida que no es para leer sino para TRABAJAR"* y *"la primera vez que aparecen
IDs estables"*. El mecanismo que importa mata el **#7**:

> **`sf` no verifica el juicio. Verifica que el juicio haya ocurrido, y sobre todos.**
> *"la historia tiene 3 criterios, el informe habla de 2"*

### `sf-propose --all` — dos costuras cambiadas

| | Construido | El ⑨ |
|---|---|---|
| **De dónde sale** | *"from the constitution"* | del **PRD** — la constitución es contexto |
| **Qué produce** | **features** directamente | **historias**. Las features aparecen en el ⑩ |

La segunda es de fondo: **separar *qué hay que poder hacer* de *en qué orden y agrupado cómo***
es lo que hace que el ⑩ sea *"una vista del backlog, no otro backlog"*.

### EARS sobrevive; `requirements.json` no

Son dos cosas distintas y en la primera pasada se tiraron juntas. **EARS es una notación**:
disparador · estado · comportamiento.

```
un test necesita  →  disparador · estado · comportamiento esperado
EARS pide         →  disparador · estado · comportamiento
```

**Son la misma terna**, y por eso EARS no es adorno: es lo que hace que un criterio se pueda
convertir en test sin inventar nada — que es exactamente el ⑮. Los criterios reales lo
muestran:

```
CA-1  acepta --json y devuelve el estado serializado     ← ubicuo ✔
CA-2  si no hay estado, sale con código 1 y mensaje      ← SI…ENTONCES ✔
CA-3  --json es incompatible con --verbose               ← ⚠ ¿y qué pasa?
```

El tercero no dice qué ocurre, y **contra eso no se puede escribir un test**. EARS lo hubiera
obligado.

**`sf` cuenta los `CA-#` con un `rg '^\- \*\*CA-'`** — cuenta líneas que matchean un patrón,
no entiende la prosa. Sigue sin pensar, aunque los criterios vivan en el cuerpo y no en el
frontmatter.

### Se cae un concepto entero: el carril lite

`Step 0: Lane Triage` clasifica lite vs standard con gate, campo `lane` en `feature.json`,
comando `sf feature set-lane` y un `lite-lane.md` aparte. **La máquina ya lo resolvió mejor**
(§8 de `maquina-estados.md`): *"no es un carril paralelo, es la misma máquina con estados
salteados"*.

> **Un carril paralelo es una segunda máquina que mantener. Saltear estados es gratis.**

### El campo `estado:` se va del `us-#`

Los cinco valores son deducibles (R6):

```
pendiente    existe el us-#, no está en roadmap.json
planificada  está en roadmap.json
en-curso     su feature es feature_actual
hecha        su feature figura cerrada en estado.json
archivada    la carpeta se movió
```

**Se ve en el índice que genera `sf`**, en un solo lugar y siempre bien. Hoy, si un LLM se
olvida de tocar el frontmatter al cerrar la feature, el archivo miente para siempre y nadie se
entera. **Lo que se calcula no puede quedar viejo.**

El `us-#.md` queda **100% humano**: el texto y los criterios.

```
sf-propose --all (decomposición)   SE ADAPTA   fuente = PRD · produce us-#
EARS + criterios con ID            SOBREVIVE   → CA-# del us-#.md
requirements.json (por feature)    SE TIRA     el contenido se parte: us-# + spec-design
Lane triage · lane · lite-lane     SE TIRAN
sf status (el índice on-demand)    SOBREVIVE   derivado, no almacenado
sf graph · sf question             SE TIRAN
sfx-triage                         SE ADAPTA   entrada C → us-# tipo bug + relacionado_a
```

---

## 7. Estado ⑤ — `roadmap` (⑩)

*Ordena y agrupa. No agrega contenido.* Sólo ids y orden.

| | Construido (Step A3) | El ⑩ |
|---|---|---|
| **Orden** | MoSCoW: Must / Should / Could | `orden: 1, 2, 3` |
| **Dependencias** | sección en prosa + `--depends-on` | ninguna — *"el `orden` ya las expresa"* |
| **Formato** | `.md` que nadie parsea | `.json`, porque `sf` cuenta ahí |

### MoSCoW se delata solo, y muere en los tres niveles

`product-decomposition.md` gasta **media página** explicando que hay **dos ejes MoSCoW
distintos** que usan las mismas tres palabras, advirtiendo *"no heredes uno del otro"*, y un
párrafo más en el caso *"cuando un `must` depende de un `could`"*.

> **Un concepto que necesita media página para no confundirse consigo mismo es un concepto
> caro.**

| Nivel | Qué preguntaría | Quién ya lo contesta |
|---|---|---|
| feature | ¿la construimos, y cuándo? | **`orden`**. Un `could` es `orden: 7` |
| us | ¿va en el MVP? | el `orden` de su feature |
| criterio | ¿puede fallar y salir igual? | **existir o no existir** (R5) |

**Y la razón de fondo es estructural:** MoSCoW es una herramienta de **negociación**, para que
un equipo discuta alcance con un PM o un cliente. **Acá no hay con quién negociar** — Javier
decide, solo, en el ⑥, el ⑧ y el ⑰. Por eso el CLI necesitaba `verification.blocking_priorities`
y la compuerta de `revision` es binaria.

*(Contra anotado: el día que quieras cerrar una feature con un hueco conocido y sin
importancia, esto vuelve. Hoy ningún paso lo pide.)*

### 🕳 El CLI no conoce el roadmap

```
$ grep roadmap cli/*.go
cli/report.go:24: // Formatos soportados (empezamos con los dos del roadmap):
```

Una línea, y es un comentario sobre otra cosa. **`roadmap.json` es uno de los tres JSON donde
`sf` cuenta, y es de donde sale `feature_actual`.** Sin él `sf next` no sabe qué sigue: es el
archivo del que depende el dolor #1 entero.

### `feature.json` resuelve un problema que no existe

> *"el `features.json` monolítico = imán de conflictos de merge **en equipo**, flujo serial
> casi forzado por la estructura"*

La máquina dice lo contrario **a propósito**: *"0 máquinas en paralelo — es una cola,
`feature_actual` alcanza"*. Y hay un choque de dueño peor: **`feature.json` guarda `status`, y
eso vive en `estado.json`.**

**Pero del mismo comentario se rescata la regla, que ya estaba escrita ahí:**

> *"la vista global es **derivada, no almacenada** — nada que sincronizar"*

Es R6. El CLI la tenía; la aplicó a `features.json` y no al `estado:` del `us-#`.

```
sf-propose --all Step A3        SE ADAPTA   orden numérico · JSON
MoSCoW · dependencias · A5      SE TIRAN
serves: [PR#]                   SE TIRA
roadmap.json en el CLI          🕳 HUECO
feature.json                    SE TIRA     → estado.json + roadmap.json
sf feature archive              SOBREVIVE   es del ⑨
sf feature set-status/set-lane  SE TIRAN
sf migrate                      SE TIRA     migra un formato que deja de existir
```

### La parada barata va en el ⑨ **y** en el ⑩

La vara no es *"¿dónde se decide algo?"* sino **¿dónde un error que no ves ahora te sale caro
después?**

```
⑨  falta una historia · un criterio ambiguo
    el daño BAJA: mala planificación, mala implementación, mala revisión
    → arreglarlo después = rehacer el ciclo

⑩  el orden está mal
    → arreglarlo = cambiar un número. Y el ⑰ ya tiene la puerta "otra feature"
```

**El error caro está en el ⑨.** Pero miran cosas distintas y no se pueden hacer en el mismo
momento: el ⑨ mira **contenido**; el ⑩ mira **forma**, y en el ⑨ todavía no hay features. Los
dos, y `para_en` es configurable — equivocarse acá es barato.

---

## 8. Estado ⑥ — `planificacion` (⑫–⑯) 🛑

El más grande. Cinco pasos, **una sola conversación**, un subagente Opus con contexto amplio,
tres archivos, **una sola revisión al final**.

### Lo primero que se cae es el Step 1 completo

`sf-propose` arranca conversando: *"what does this feature do? why? who uses it? 3-5
exchanges"*. **En la máquina eso ya ocurrió, en el ⑨:**

> *"el planificador (⑫–⑯): **es su entrada — el *qué* ya viene resuelto**"*

**Y eso es exactamente por qué `planificacion` puede ser un subagente.** Un subagente no te
habla; si tuviera que preguntar *"¿qué hace esta feature?"*, no podría serlo. **El `us-#` es lo
que lo hace posible** — no es una simplificación, es un requisito de la arquitectura.

### Tres gates → uno

```
sf-propose   🔴 requirements → 🔴 design → 🔴 tasks     tres paradas
la máquina                  🛑 ⑰                        una, al final
```

Ya argumentado al corregir el strawman: *"el ⑰ revisa el bloque entero, y 'pido cambios' vuelve
al bloque entero. **Nadie entra ni sale por el medio.**"* Con los dos gates se va también el
caso *"si el design invalida los requirements, avisá y ofrecé actualizarlos"*.

### ⑫ `decision.md` — hueco, con candidato

Nada construido. `--design-first` elige **una** arquitectura, no compara **tres**.
**`sfx-think` es el método**, con una restricción agregada: **tres, no "las que salgan"**.

### ⑬ `spec-design.md` — se funden dos artefactos

`requirements.json` + `design.json` → **un archivo**. Se rescata una regla de Step 3, que es R2
aplicada a secciones:

> *"incluí una sección sólo si cambia una decisión o informa la implementación — las que dirían
> 'N/A' generan ruido"*

### ⑭⑮ `tareas.json` — el choque de diseño de la ronda

El CLI es explícito y dice lo contrario:

> *"**do NOT group tasks into waves here.** El agrupamiento en olas es una preocupación de
> ejecución que `sf plan compute` deriva del grafo de `depends_on`."*

**El argumento es bueno** —lo deducible no se guarda, y un orden topológico es determinista, o
sea de `sf` (R1)—. **Pero se cae por lo que el lote *es* en esta máquina:**

```
un lote  =  la unidad de COMMIT        "un lote terminado = un commit"   (#2, #3)
         =  la unidad de SUBAGENTE     uno por lote, contexto fresco     (#5)
         =  el filtro de sf context    "dame las tareas del lote 2"
```

**Un orden topológico agrupa por *cuándo se puede*; hace falta agrupar por *qué va junto*.** Dos
tareas sin dependencia entre sí pueden caer en la misma ola sin tener nada que ver — y ese lote
produce un commit incoherente y un subagente con dos temas en la cabeza. **Eso es el dolor #3.**

El desempate es R7: `lote` se usa en tres lugares, `depends_on` en ninguno.

### El hook — 35.000 caracteres que la máquina no necesita

Es el mecanismo que hace cumplir *"el skill nunca escribe el artefacto a mano"*: se escribe en
`drafts/`, se corre `sf save`, y un hook **deniega** cualquier `Write`/`Edit` sobre el archivo
real.

1. **Duplica las compuertas.** Si un skill escribe `tareas.json` con basura, la compuerta ya lo
   agarra (*¿cada lote tiene al menos un test? ¿cada tarea apunta a un CA?*) y el estado no se
   mueve. **El control ya existe, en el lugar correcto.**
2. **Es específico del harness.** `--harness=generic|claude-code`, con matchers `Write|Edit|Bash`
   para tres harness. **Es justo lo que `session.md` §5 dijo que había que evitar:** *"una
   máquina de estados vive en un programa, y un programa corre igual en todos lados"*.

```
sf-propose Step 1                 DESAPARECE   el us-# ya la contestó
los 3 gates                       SE FUNDEN    → un solo 🛑 en el ⑰
sfx-think                         SOBREVIVE    método del ⑫ · "tres, no varias"
sf-propose Steps 2-4              SE ADAPTA    → un solo spec-design.md
"sección sólo si cambia algo"     SOBREVIVE    regla anti-N/A
lote declarado por el modelo      SOBREVIVE    commit + subagente + filtro
depends_on · sf plan compute      SE TIRAN     sólo existen para derivar el lote
sf tasks render|validate          SOBREVIVE    tareas.json es uno de los 3 JSON
sf design · sf requirements       SE TIRAN
sf hook (35K) + adaptadores       SE TIRA      duplica compuertas · no es portable
sf save (3 JSON) + drafts/        SE ADAPTA
sf sources + sources.json         SE TIRAN     no hay material de cliente
--design-first · --from-code      SE TIRAN
⑯ el modelo recomendado           🕳 HUECO      un campo en tareas.json (regla 1.1)
```

---

## 9. Estado ⑦ — `implementar` (⑱–⑳)

> *"Es el estado donde muere casi todo."*

```
antes    ¿existe la branch que dice la constitución?        #4
medio    corre los tests del lote: TIENEN QUE FALLAR        #8 #5
después  corre los tests: pasan · ¿hay commit del lote?     #2 #3
```

### `redwitness` — la joya ya estaba construida

Su comentario diagnostica el dolor #5 mejor que ningún otro texto del repo:

> *"Un test escrito DESPUÉS del código tiende a afirmar lo que el código hace en vez de lo que
> el requisito pide. Es el **fraude semántico en su forma más común y menos malintencionada** —
> el agente implementa, lee su propia implementación, y escribe un test que la describe. Pasa el
> contrato, pasa la causalidad, se sella, y **no verifica nada**."*

**Pero el mecanismo se da vuelta, y sale más simple.** El CLI eligió ser **pasivo** a propósito:

> *"NO OBLIGAR LA PRÁCTICA, REGISTRAR LA EVIDENCIA. SpecForge no impone TDD… forzarlo sería
> exactamente lo que le criticamos a los frameworks que revisamos."*

**Era correcto para una herramienta que no sabe cómo trabajás.** Esta máquina sí sabe: **ejecuta
tu flujo**, y el ⑲ dice *"SIEMPRE en este orden… 'siempre', no es condicional"*. Imponerte tu
propio método no es imponer nada.

```
CLI      PASIVO   se llena solo · CodeHash para reconstruir contra qué código falló
máquina  ACTIVO   sf lote start 2 → corre los tests él mismo y exige rojo
```

**Y la compuerta activa hace innecesario el `CodeHash`** (R6): existía porque `sf` no había visto
el rojo y lo tenía que reconstruir. Ahora **ve el rojo y el verde con sus propios ojos, en
orden**.

### El agujero que ninguno de los dos tapaba

`sf` ve rojo → el subagente trabaja → `sf` ve verde. **¿Y si lo que cambió entre medio fue el
test?** Un subagente que no logra implementarlo puede aflojar el test, y `sf` no puede saberlo.
Es el dolor #8 en su forma más astuta.

**Se tapa con un hash de los archivos de test**, tomado en el rojo y comparado en el verde:

```
sf lote start 2   ✓ los 6 tests fallan. Rojo confirmado.   → rojo:true + hash de los tests
   ...el subagente implementa...
sf lote done 2    ✗ El test TestFrontmatterInvalido cambió entre el rojo y el verde. No avanzo.
```

### `sf context implement` ya existe

```
sf context for-wave --n=N          ← construido
sf context implement f-1 --lote 2  ← lo que pide el ⑱
```

**Es el mismo comando con otro nombre.** El mecanismo del ⑱ —la idea que estaba guardada *"para
más adelante"* y la ronda 7 ascendió— ya está escrito. Sólo cambia qué entra en el sobre.

### `sf gate` — 35K contra tres booleanos

Un ledger de aprobaciones por feature con sellado de hash y timestamps por fase (requirements →
design → tasks → plan → build → verdict). **La máquina tiene tres aprobaciones en total**, y
ninguna es por feature: los dos sellos del `producto` más el ⑰, que ya no sella un archivo —
elige una de tres puertas. Fundidos los tres gates de planificación (⑥) y muerto `lane` (④),
**quedan cero gates por feature.**

### `sf stale` — la señal 1 es el mecanismo; el motor sobra

> *"silent edit: su hash actual ≠ el hash sellado en su gate. **Es la señal ROBUSTA.**"*

Eso es `prd_hash`, y `base_commit` es lo mismo aplicado al repo. Las señales 2 y 3 (upstream
reabierto, propagación) existen por la cadena `requirements → design → tasks → plan`. **Esa
cadena se fundió en un archivo: sin cadena no hay propagación.**

```
redwitness (el diagnóstico)       SOBREVIVE   el mejor texto del repo sobre el #5
redwitness (el registro pasivo)   SE TIRA     la compuerta activa es más simple
sf check run                      SOBREVIVE   es la compuerta: corre y devuelve exit code
sf context for-wave               SOBREVIVE   → sf context implement. El mecanismo del ⑱
sfx-tdd                           SOBREVIVE   RED→GREEN→REFACTOR es el ⑲ palabra por palabra
sfx-github                        SOBREVIVE   el skill HACE y escribe el mensaje;
                                              sf comprueba que el commit exista. Nunca lo lee
sf stale (señal 1)                SOBREVIVE   → prd_hash + base_commit
el motor de staleness (~10K)      SE TIRA     → dos comparaciones de string y un if
sf gate · sf gateshow (35K)       SE TIRAN    cero gates por feature
sf run · build.mode               SE TIRAN    sf no lanza (regla dura 1) · un solo modo
sf recover · sf verify            SE TIRAN    el retome sale de "¿qué archivos existen?"
sf trace (como artefacto)         SE TIRA     la lista de tests ya está en tareas.json
sf trace (como chequeo)           SOBREVIVE   dentro de la compuerta: rg + contar
sf coverage · sf delta            SE TIRAN
sf-build                          SE ADAPTA   queda el preflight; el bucle es del orquestador
```

---

## 10. Estado ⑧ — `revision` (㉑㉒)

### `sf mutation` ya era un contador, no un generador

En la primera pasada se anotó *"choque frontal"* por leer mal el comando. Su código dice:

> *"El umbral es **config de la herramienta**, no nuestra… se sigue consumiendo **nada más que
> el exit code**."*

Lee `build.mutation_cmd` y hace **shell-out a la herramienta del lenguaje**. Es exactamente el
diseño acordado, y es la regla de no reinventar la rueda ya aplicada:

```
el modelo grande       genera mutantes semánticos
gremlins / mutmut / …  genera mutantes mecánicos
sf mutation            corre y cuenta        ← ni genera ni opina
```

**Y sale una mejora gratis:** `mutationScope()` saca los archivos a mutar de `trace.json`, que
se tira. El reemplazo es mejor y ya existe:

```
git diff base_commit..HEAD   →  los archivos que la feature tocó DE VERDAD
```

`trace.json` decía lo que la feature **declaraba** como suyo; el diff dice lo que **es**.

### El juez de fase: el mecanismo sobrevive, la configuración no

`sf context for-judge` + `sf gate record-verdict` implementan un reparto que **es la arquitectura
de `session.md` §6, escrita antes de decidirla:**

```
sf     junta el material            determinista
LLM    juzga, en subagente fresco   cooperativo
sf     persiste el veredicto        determinista
```

Su propio código: *"El CLI **NO juzga**: junta lo que el subagente-juez fresco necesita y nada
más."* **Tiene un consumidor: el ㉑.**

Se cae el `--phase` (existía para juzgar cada fase por separado; las fases se fundieron) y se
cae `audit.phase: off | nudge | block` por **R3**.

### `sf-check` Step 1 — el mecanismo del #7 ya construido

> *"Este paso es **no negociable**. Todo check DEBE producir la matriz completa. **'Todos los
> tests pasan' no es un sustituto.**"*

Es la misma idea que `criterios: { "us-1/CA-1": "cumple" }`: obligar a contestar uno por uno en
vez de dejar decir *"anda"*. **El nuevo es mejor porque es JSON:** `sf` puede contar *"la
historia tiene 3 criterios, el informe habla de 2"*. Sobre una matriz en prosa, no.

### `revision.json` gana dos campos

**`descartado`, y sólo lo pone Javier.** Sin eso, un hallazgo exagerado por el revisor traba la
feature — y `sf` estaría frenando sobre la **opinión** de un modelo (R3).

**`tipo`, porque el ㉑ absorbe el chequeo de arquitectura.** `sf arch check` se tira: exigía que
el diseño declarara sus reglas en formato máquina, y **el auditor lee prosa** — ya tiene el
`spec-design.md` y el código abiertos. Pero un hallazgo de arquitectura no cuelga de ningún `CA`:

```json
{ "id": "h-2", "origen": 21, "tipo": "diseño", "criterio": null, "estado": "abierto",
  "detalle": "check.go escribe estado.json; el diseño dice que sólo estado.go" }
```

Ahí se rescata el `ConstitutionViolations []string` del esquema viejo: **era la categoría
correcta, en el archivo equivocado.**

```
hallazgo.estado    abierto | arreglado | descartado
hallazgo.tipo      criterio | diseño | constitución
```

```
sf mutation                      SOBREVIVE    ya era un contador
  └ scope por trace.json         SE ADAPTA    → git diff base_commit..HEAD. Más preciso
sf context for-judge             SOBREVIVE    → sf context revisar
sf gate record-verdict           SE ADAPTA    → sf save review
sf-check Step 1 (matriz oblig.)  SOBREVIVE    es el dolor #7
sf review (validación cruzada)   SOBREVIVE    es la compuerta de salida
reviewFile (el esquema)          SE TIRA      → hallazgos[] con estado y tipo
--phase · audit.phase off/block  SE TIRAN     una sola instancia del juez · R3
sf arch (el comando)             SE TIRA      la pregunta pasa al ㉑
sf evidence                      SE TIRA      murió con R5
contract/audit.md · judge.md     DESCARTADOS  son el contrato de sf-audit, que ningún
                                              estado consume
```

---

## 11. Estado ⑨ — `cierre` (㉓) ⏸

El ㉓ produce **tres cosas**, no una: la doc, el aprendizaje, y el movimiento de la carpeta.

### La doc tiene dos mitades, dos lectores y dos fuentes

```
TÉCNICA     cómo está hecho        ← del CÓDIGO           el que mantiene · el ⑫ futuro
FUNCIONAL   qué hace y para quién  ← del us-# y la spec    Javier · el usuario
```

**Y ahí está el límite de `sfx-documenter`:** su método es *"documentación **desde el código**"*.
La mitad técnica le sale nativa; **la mitad funcional no está en el código.** El sobre del ⑨
tiene que traer las dos:

```
sf context documentar f-1
  → git diff base_commit..HEAD   el código que quedó    → la mitad técnica
  → us-1.md · us-3.md            qué se pedía           → la mitad funcional
  → spec-design.md               cómo se resolvió
```

**El `us-#` gana un consumidor que no estaba en el inventario: el ㉓.**

### El journal entra a la máquina

Su gancho ya era el correcto (*"cuando una feature se archiva, el LLM extrae lecciones
durables"*), y tenía un doble rol del que sobra la mitad: **memoria sí, progreso no** — el
progreso vive en `estado.json`.

**Dónde vive:** se archiva **con la feature**, y el que lo encuentra es `sf` (regla 1.4). Su
lector es el mismo que el de la doc — **el ⑫ de una feature futura**:

```
sf context planificar f-3
  → constitucion.md · us-#
  → aprendizajes de features anteriores        ← el journal, servido
```

**El *backprop* no necesita una parada nueva.** Cuando una lección se repite, se propone subirla
a la constitución — y eso cabe en la ⏸ que el ⑨ ya tiene:

```
sf:  f-3 lista para archivar.
     El ㉑ fue 1 vuelta, 2 hallazgos, los dos arreglados.
     2 aprendizajes nuevos.
     ⚠ Uno se repite por tercera vez:
       "los tests de tabla en Go tienen que nombrar el caso"
       ¿lo subo a la constitución?
     [enter] archivo   [c] subir a la constitución   [v] ver el detalle
```

### `sfx-github`: dos defaults cableados que ahora tienen dueño

| El skill | La regla |
|---|---|
| *"Squash merge default"* | **`merge: no-ff`** |
| *"NEVER commit directly to main. ALWAYS PR"* | ㉓: *"se manda el PR, **o se pushea**"* |

No son errores: es **R4** — el skill deja de tener convenciones propias y lee el bloque `git:`
de la constitución. Y `archive = merge a main` **sobrevive tal cual**: es literalmente el ㉓.

```
sf feature archive              SOBREVIVE   mueve la carpeta entera
sfx-documenter                  SOBREVIVE   + la mitad funcional, que no sale del código
sfx-github (branch/commit/PR)   SOBREVIVE   lee git: de la constitución
sf journal + sfx-journal        SOBREVIVEN  el aprendizaje del ⑨
  └ el rol de progreso          SE TIRA     lo hace estado.json
backprop a la constitución      SOBREVIVE   sobre la ⏸ que ya existe
modo equipo · export a issues   SE TIRAN    no hay equipo · leen archivos que ya no existen
la parada barata del ⑨          🕳 falta     el resumen de vueltas, hallazgos y aprendizajes
```

---

## 12. Barrido final — lo que ningún estado reclamó

| Comando | | Por qué |
|---|---|---|
| **`sf next`** | **SOBREVIVE** | el corazón: *dónde estás · qué sigue · ¿podés avanzar?* |
| `sf state current` | **SE FUNDE** | en `sf next` |
| `sf status` | **SOBREVIVE** | adelgazado: genera las vistas (backlog, roadmap) |
| `sf lint` · `sf doctor --install` | **ANDAMIO** | mantienen `sf` misma |
| `sf doctor --drift` (23K) | **SE TIRA** | clasificar deriva spec↔código lo hace el ㉑, con juicio |
| `sf domain` | **SE TIRA** | ningún estado lo consume |
| `sf events` | **SE TIRA** | telemetría del hook — muere con el hook |
| `sf metrics` | **SE TIRA** | mide el tamaño de las slices: herramienta de desarrollo |
| `sf report` · `terms` · `archlint` · `integrity` | **SE TIRAN** | internos de comandos que ya cayeron |

---

## 13. Los seis huecos

Lo único que hay que construir desde cero. **Cinco de los seis son chicos.**

| # | Hueco | Tamaño |
|---|---|---|
| 1 | **`roadmap.json` en el CLI** | el más grande. De ahí sale `feature_actual` |
| 2 | **`sfp-po`** — el skill del PRD (⑦) | un `.md` delgado, 5 bloques |
| 3 | **`decision.md`** (⑫) | ninguno: es `sfx-think` con "tres" |
| 4 | **el ⑯** — recomendar modelo | un campo en `tareas.json` (regla 1.1) |
| 5 | **la parada barata del ⑨** | una vista |
| 6 | **el sobre y `relacionado_a`** | un `if` y una ruta más |

> **El hueco 6 apareció en la ronda de skills** ([`skills.md`](skills.md) §7), tirando
> `sf-amend`, y salió de leer el código y no los documentos: `relacionado_a` está declarado en
> `historia.go:61` y **no lo lee nadie**, y `sf context` **no toca `.docs/archivado/` en ningún
> lado**. Las dos caras del mismo agujero: el subagente que arregla un bug **no ve la spec de la
> feature que rompió**, y ahí es donde nacen los mocks. **Arreglo:** si la historia del lote
> tiene `relacionado_a`, el sobre agrega la spec archivada de esa feature.

---

## 14. El saldo

```
skills      15  →   SOBREVIVEN  scout(adaptado) · think · grill-me · tdd · github ·
                                documenter · journal · triage · propose(partido) ·
                                init(partido) · build(preflight) · check(partido) ·
                                explain (utilitario, tal cual)
                    SE VAN      audit → sfx-audit, fuera de la máquina
                                amend → SE TIRA (skills.md §6)
                    SE SUMAN    sfp-po · sf-cierre
                    SE PARTE    propose → el ⑨, el ⑩ y el ⑫-⑯ son tres estados distintos
                                init   → sfp-constitucion (Step 3) + el COMANDO sf init

CLI         47  →   ~10 de máquina + 3 de andamio
                    ~34 se tiran, incluidos los tres más grandes:
                        gate (35K) · hook (35K) · doctor --drift (23K)

capa cross      →   DESCARTADA. audit.md + judge.md son el contrato de sf-audit,
                    y ningún estado lo consume
```

> ### ✏️ CORREGIDO — 2026-08-14, en la ronda de [`skills.md`](skills.md)
>
> **Este saldo no cerraba: listaba 12 que sobreviven + 1 que se va = 13, y los skills son 15.**
> Faltaban `sfx-explain` (utilitario, queda tal cual) y `sf-amend` (**se tira** — su punto de
> entrada ya existe y se llama `sf new`; ver `skills.md` §6).
>
> Y los nombres de arriba son **el método**, no el skill que lo ejecuta. **El mapa
> `estado → skill` firme vive en [`skills.md`](skills.md) §4** y en `sf/internal/maquina/maquina.go`.

**El diagnóstico de la refundación queda confirmado por los números:** dos tercios del CLI
existían para sostener conceptos —el ledger de gates, el enforcement por hook, la cadena de
cuatro fases, MoSCoW, el carril lite, la traza `PR#`— que **la máquina de estados vuelve
innecesarios**, no porque estuvieran mal hechos sino porque resolvían un problema que la máquina
resuelve de otra forma o que no existe en el flujo de Javier.

**Y lo que sobrevive tiene una forma clara:** casi todo lo que sobrevive **corre algo, cuenta
algo o sirve un archivo**. Nada de lo que sobrevive piensa.

---

## 15. Correcciones a `artefactos.md`

Tres, todas consecuencia de decisiones de este recorrido:

1. **§7 y §10 — el `estado:` del `us-#` sale.** §10 dice *"el estado vive en el frontmatter del
   `us-#` y en ningún otro lado; se marca una vez y `sf` genera las dos vistas"*. Ya no se marca
   **ni una vez**: `sf` cierra la feature en `estado.json` y las tres vistas salen de ahí.
2. **§10 — el ㉓ no agrega un archivo, agrega dos.** La doc **y** el journal.
3. **§11 — el `us-#` gana un consumidor:** el ㉓, por la mitad funcional de la doc.

Y dos campos nuevos que no estaban en la ronda 7:

- **`constitucion.md`** → `test_cmd:` en el frontmatter (§5).
- **`estado.json`** → el hash de los archivos de test por lote (§9).

> ### ✅ APLICADAS — 2026-08-13
>
> Las tres y los dos campos están en `artefactos.md`, con el resumen en su **§13**. Salieron dos
> arreglos de coherencia que no estaban en la lista y que se veían al tocar los mismos párrafos:
>
> - el `estado.json` de `artefactos.md` §11 era **el borrador viejo** (`paso`, `verde`,
>   `modelo_recomendado`) → se reemplazó por un puntero a `maquina-estados.md` §9 y la lista de
>   qué cambió. **El estado.json vive en un solo documento.**
> - `hash_tests` se agregó también a `maquina-estados.md` §9, que es donde el `estado.json`
>   está cerrado — si no, `artefactos.md` apuntaba a un campo que la fuente no tenía.

---

## 16. Lo que sigue

1. **La superficie de `sf`.** Ya aparecieron `next`, `done`, `context <estado>`, `lote start`,
   `lote done`, `status`, `save`, `check run`, `mutation`, `feature archive`, `install`. Falta
   el inventario completo y **quién llama a cada uno**.
2. **El reparto orquestador ↔ skills**, con los 9 estados fijos y los veredictos de acá.
3. **Brownfield** — el debate aparcado en §5. Decide el destino de `sf onboard scan`.

> **Y la advertencia que sigue vigente** (`decisions-specforge`, 2026-08-07): no sobre-indexar
> en determinismo. `sf` nació para producir artefactos útiles que hagan alucinar menos al
> modelo, no para garantizar lo que no se puede garantizar.
