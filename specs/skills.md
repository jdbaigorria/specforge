# El mapeo estado → skill

**Fecha:** 2026-08-14 · **Método:** contrastar el mapa que ya vive en `maquina.go` contra los
veredictos de [`que-sobrevive.md`](que-sobrevive.md), estado por estado.

> **Por qué hay una ronda para esto.** El mapa `estado → skill` se escribió durante la
> construcción del binario, con tres nombres marcados ⚠ y el resto dado por bueno. Al pasarlo
> por la **regla de prefijos** —que la ronda anterior había declarado no decorativa— resultó que
> **seis de las nueve entradas estaban mal**, y las tres nuevas no eran descuidos de nombre: eran
> el mapa apuntando a **utilitarios**, que es justo lo que la regla prohíbe.

**Entra:** el mapa de `sf/internal/maquina/maquina.go`, los veredictos de `que-sobrevive.md` y la
regla de prefijos de su §4.
**Sale:** los 9 nombres firmes, quién compone a quién, los 15 veredictos completos, y la cirugía
que hay que hacerle a cada skill.

---

## 1. La regla que ordena todo

Estaba escrita en `que-sobrevive.md` §4 y no se había aplicado al código:

```
sfp-*   los 5 estados de PRODUCTO    corren una vez
sf-*    los 4 estados de FEATURE     corren una vez por feature
sfx-*   utilitarios                  fuera de la máquina — sf NO los conoce
```

**La tercera línea es la que muerde.** Un `sfx-` es standalone: no sabe que existe una máquina de
estados, no pide sobre, no avisa que terminó. Si `sf next` devuelve el nombre de un `sfx-`, el
orquestador lanza a alguien que no va a llamar a `sf done` — y **el estado no se mueve nunca**.

---

## 2. El mapa viejo, y qué tenía mal

| estado | `maquina.go` decía | veredicto |
|---|---|---|
| `brief` | `sfp-scout` | ✅ |
| `prd` | `sfp-po` | ⚠ prefijo bien, **no existe** |
| `constitucion` | `sf-init` | ❌ prefijo de feature · **choca con el comando `sf init`** |
| `backlog` | `sf-propose` | ❌ prefijo de feature |
| `roadmap` | `sf-propose` | ❌ prefijo de feature · **el mismo skill en dos estados** |
| `planificacion` | `sfx-think` | ❌ **es un utilitario** |
| `implementar` | `sfx-tdd` | ❌ **es un utilitario** |
| `revision` | `sf-check` | ✅ |
| `cierre` | `sfx-documenter` | ❌ **es un utilitario** |

**Dos de nueve estaban bien.** Y el patrón de los errores tiene una explicación única: el mapa se
escribió preguntando *"¿quién sabe hacer este trabajo?"* y la respuesta correcta a esa pregunta es
casi siempre un utilitario, porque **el método vive ahí**. Lo que faltaba era la pregunta de
arriba: *"¿quién habla con la máquina?"*.

---

## 3. El patrón que lo resuelve, y ya estaba inventado

`sfp-scout` no hace el trabajo del ② ni del ⑤: **los compone.** `que-sobrevive.md` §3 lo dice —
*"un skill delgado que invoca a los otros dos"*, y los utilitarios *"not replaced — they stay
standalone"*.

Eso es exactamente lo que falta en los otros tres estados:

```
          el skill de ESTADO                 los UTILITARIOS
          delgado · habla con sf             el método · standalone
          ───────────────────────            ─────────────────────
brief     sfp-scout           compone   →    sfx-think (②) · sfx-grill-me (⑤)
planif.   sf-plan             compone   →    sfx-think (⑫)
implem.   sf-build            compone   →    sfx-tdd (⑲) · sfx-github (el commit)
cierre    sf-cierre           compone   →    sfx-documenter · sfx-journal · sfx-github
```

**Y no cambia nada de lo ya construido:** `sf next` sigue devolviendo **un** skill, `sf context`
sigue sin llevar argumentos, y `sf` sigue sin nombrar un `sfx-` en ningún lado.

> **El corolario de método:** los `sfx-` no se tocan. La cirugía —pedí el sobre, avisá que
> terminaste— **es sólo de los nueve de estado**. Un utilitario que aprende a llamar a `sf done`
> dejaría de servir fuera de un proyecto SpecForge, que es la mitad de su valor.

---

## 4. El mapa cerrado

```
PRODUCTO                                                          de dónde sale el método
  brief          sfp-scout          adaptar     sfp-scout, delgado · think + grill-me
  prd            sfp-po             CONSTRUIR   nada — el hueco 2 de §13
  constitucion   sfp-constitucion   adaptar     sf-init Step 3 (la conversación)
  backlog        sfp-backlog        adaptar     sf-propose --all · sfx-triage (entrada C)
  roadmap        sfp-roadmap        adaptar     sf-propose --all Step A3

FEATURE
  planificacion  sf-plan            adaptar     sfx-think (⑫) + sf-propose Steps 2-4
  implementar    sf-build           adaptar     sf-build (preflight) + sfx-tdd + sfx-github
  revision       sf-check           adaptar     sf-check Step 1 (la matriz) — el dolor #7
  cierre         sf-cierre          CONSTRUIR   documenter + journal + github, delgado
```

**Dos se construyen desde cero y los dos son delgados:** `sfp-po` (cinco bloques) y `sf-cierre`
(un compositor). Los otros siete tienen método adentro y lo que cambia es la cáscara.

### Tres consecuencias de nombre

- **El skill `sf-init` desaparece.** Se parte en dos, y las dos mitades ya tenían destino:
  su **Step 3** (la conversación de la constitución) es `sfp-constitucion`, y **`detectStack()`**
  es del **comando** `sf init` (el paso 3 del plan). Que el skill y el comando se llamaran igual
  y **no hicieran lo mismo** era una bomba de tiempo.
- **`sf-propose` desaparece como nombre.** Se parte en tres, que es lo que `que-sobrevive.md` §14
  ya decía: el ⑨, el ⑩ y el bloque de planificación son tres estados distintos.
- **`sf-build` se queda con su nombre pero cambia de trabajo.** Era *planificar y ejecutar con un
  gate humano por wave*; ahora es el skill del ⑱–⑳: preflight, y el bucle lo maneja el
  orquestador (regla dura 1: `sf` no lanza a nadie, y el skill tampoco).

---

## 5. Los 15 veredictos — completos

`que-sobrevive.md` §14 listaba 12 que sobreviven + 1 que se va = **13, y hay 15**. Faltaban dos, y
acá se cierran:

```
LOS 9 DE ESTADO
  sfp-scout       adaptar · brief
  sf-init         SE PARTE  → sfp-constitucion (Step 3) + el comando sf init (detectStack)
  sf-propose      SE PARTE  → sfp-backlog + sfp-roadmap + sf-plan
  sf-check        adaptar · revision
  sf-build        adaptar · implementar
  sfp-po          SE CONSTRUYE
  sf-cierre       SE CONSTRUYE

LOS 7 UTILITARIOS — no se tocan
  sfx-think       método del ② y del ⑫
  sfx-grill-me    método del ⑤
  sfx-tdd         método del ⑲
  sfx-github      el commit, la branch, el merge
  sfx-documenter  la doc del ㉓
  sfx-journal     el aprendizaje del ㉓
  sfx-triage      entrada C: bug → us-# con tipo y relacionado_a
  sfx-explain     ← FALTABA. Feynman, no requiere proyecto. Queda tal cual.

FUERA DE LA MÁQUINA — pero con comando propio
  sfx-audit       el AUDITOR. Ningún estado lo consume: lo corre Javier cuando
                  quiere. REESCRITO ENTERO — ver §10
  sf-amend        ← FALTABA. SE TIRA — ver §6
```

> **Nota de política:** *tirar* un skill significa **registrar el veredicto**, no borrar la
> carpeta todavía. Misma regla que `cli/` (session.md §12): se apaga cuando el reemplazo existe,
> no antes.

---

## 6. `sf-amend` se tira, y la pregunta que lo cerró

`sf-amend` cambiaba una feature **ya archivada**, editando su spec **in-place** para que no
hubiera *"una segunda verdad divergente"*. Javier preguntó lo que había que preguntar:

> *"en realidad sería un punto de entrada, porque cuando se encuentra un bug y fue archivada la
> feature ¿tenés que desarchivarla o no?"*

**No se desarchiva nada.** Verificado en el código, no en los documentos:

```
sf new "el login rompe con email en mayúsculas"
  → crea us-7 · tipo: bug · relacionado_a: us-3
  → BacklogVisto = false                         (reabre la ⏸ del ⑨)

⑩  el roadmap agrupa us-7 en una feature NUEVA

sf take f-4
  → tipo: bug ⇒ saltea planificacion y revision  (implementar → cierre)
```

`.docs/archivado/` **no se toca en ninguna rama del código**.

**Y la premisa de `sf-amend` es la que está mal**, no su implementación:

> **Una spec archivada no es la documentación del sistema: es el registro de una decisión, con
> fecha.** Editarla in-place borra *por qué se hizo así*. La documentación viva la escribe el ㉓
> —`sfx-documenter` *"+ la mitad funcional, que no sale del código"*— y ésa sí se reescribe.

**Su punto de entrada ya existe y se llama `sf new`.** Las entradas B y C del flujo son exactamente
eso: el backlog es el embudo, y un bug sobre algo archivado entra por el embudo como todo lo demás.

---

## 7. El hueco que quedó al descubierto — el sobre y lo archivado ✅ **cerrado**

Tirar `sf-amend` dejó ver algo que **no está en ningún documento** y salió de leer el código:

```
historia.go:61   RelacionadoA string `yaml:"relacionado_a"`     declarado
historia.go:195  relacionado_a: null   # el us-# original…      en el esqueleto
                 ────────────────────────────────────────────
                 CERO CONSUMIDORES
```

Y su otra cara: **`sf context` no toca `.docs/archivado/` en ningún lado.**

**Las dos juntas son el dolor #5 esperando.** El subagente que arregla el bug arranca de cero, sin
la spec de la feature que rompió — y *"el implementador que no encuentra lo que la spec dice,
improvisa"*, que es exactamente donde nacen los mocks.

**Es el mismo patrón que `tipo: bug` antes de H20:** un campo que ya existía, esperando su
consumidor.

### El arreglo, y es uno solo para las dos caras

> Si la historia del lote tiene `relacionado_a`, el sobre agrega **la spec archivada de la feature
> que contenía ese `us-#`**.

No rompe ninguna regla: **servir un archivo no es pensar ni lanzar a nadie.** Y es el mismo
mecanismo que ya usa el sobre para los journals de las features cerradas (`docs.Journals()`).

### ✅ Construido — 2026-08-14, `sobre.loQueRompio`

La cadena son dos saltos y **los dos usan datos que ya existían**:

```
us-7.relacionado_a = "us-3"          →  qué historia rompió
roadmap: ¿qué feature tenía us-3?    →  f-1
.docs/archivado/f-1-nucleo/spec-design.md
```

**El segundo salto funciona porque archivar mueve la carpeta y NO toca el roadmap**, así que las
features cerradas siguen ahí con sus historias. Es una búsqueda en memoria, sin leer nada nuevo.

Cuatro tests, y los dos que más valen son los que dicen **cuándo NO servir**:

- la feature original **todavía no está archivada** — ofrecer una ruta que no existe sería mentir
  al revés;
- el bug cayó en **la misma feature** que la historia original — la spec ya está en el sobre, y
  repetirla es gastarle contexto al que trabaja.

Verificado que los tests fallan sin el arreglo.

---

## 8. La cirugía común — los nueve, todos igual

```
+  arranca con   sf context      pedí tu sobre, no busques qué leer
+  termina con   sf done         avisá que terminaste
−  las convenciones propias      salen de la constitución            (R4)
−  el modo / carril / fase       la máquina saltea estados           (R4)
```

Y las dos que no cambian:

- **El skill no sabe en qué estado está.** Por eso `sf context` no lleva argumentos: el skill dice
  *"dame mi sobre"*, y cuál es el sobre lo decide `sf`.
- **`sf done --msg` es lo que cierra el lote.** Sin mensaje no hay `done`, así que **cerrar el
  lote _es_ commitear**: los dolores #2 y #3 dejan de ser detectables para ser imposibles.

---

## 9. ✅ Los nueve escritos — 2026-08-14

```
sfp-scout        121L   adaptado: pinponeo PRIMERO · −PR# · −question · −ready/not-ready ·
                        −research.md · −decision-log.md · un archivo: brief.md
sfp-po            98L   NUEVO. El tamaño lo dan sus dos lectores: el ⑧ y el ⑨
sfp-constitucion 124L   de sf-init Step 3. Técnica, no filosófica. 3 archivos → 1
sfp-backlog      155L   de sf-propose --all. Fuente = PRD · produce us-# · EARS
sfp-roadmap      111L   de sf-propose --all Step A3. Sólo ids y orden
sf-plan          159L   ⑫–⑯ en una pasada. Compone sfx-think · el lote lo declara el modelo
sf-build         137L   reescrito. Un lote, un subagente, un commit. Compone sfx-tdd
sf-check         132L   reescrito. La matriz sobre TODOS los criterios · el score, no los parches
sf-cierre        106L   NUEVO. Delgado: compone sfx-documenter + sfx-journal
```

**Y los nueve del mapa existen como carpeta, con el `name:` igual al directorio.** Verificado.

### Lo que se borró, ahora que el reemplazo existe

```
sf-propose   BORRADO   se partió en sfp-backlog + sfp-roadmap + sf-plan
sf-amend     BORRADO   su punto de entrada es `sf new`  (§6)
sf-init      BORRADO   Step 3 → sfp-constitucion. Y era ACTIVAMENTE dañino: escribía
                       constitution.json y context/*.md, rutas que ya no existen
sf-audit  →  sfx-audit RENOMBRADO. Ningún estado lo consume
```

> **El material de brownfield (`sf-init/references/onboard.md`) queda en la historia de git.**
> El debate sigue aparcado en `que-sobrevive.md` §5 y lo nombra por su nombre.

### Tres cosas que aparecieron escribiendo, y no estaban en el plan

- **`sfx-github` tenía convenciones propias** —*"Squash merge default"*, `feature/<slug>`,
  *"NEVER commit to main, ALWAYS PR"*— y R4 lo nombra por su nombre. Ahora lee `git:` de la
  constitución, y su `specforge-integration.md` decía una máquina que no existe (waves, gate
  ledger, `features.json`, lanes, team mode). **Reescrito: `sf` es dueño del repo local, el
  skill es dueño del remoto.**
- **Once punteros colgados a skills borrados** en cinco utilitarios (`sfx-triage`, `sfx-audit`,
  `sfx-grill-me`, `sfx-think` y sus plantillas): todos decían *"corré `sf-propose fix-…`"*. En
  la máquina nueva eso es **`sf new`** — el backlog es el embudo. Arreglados uno por uno.
- **Dos `references/` cambiaron de dueño y mejoraron al hacerlo.** `requirement-quality.md` pasó
  de `sf-check` a `sfp-backlog`: **un criterio defectuoso es baratísimo de arreglar cuando es una
  línea de texto**, y carísimo tres estados después. Y `backprop.md` pasó a `sf-cierre` — con el
  contador a mano borrado: los journals archivados ya son la fuente, y **un contador que se
  mantiene a mano se desincroniza la primera vez que alguien se olvida**.

---

## 10. El auditor — reescrito, y ganó un comando

Javier lo definió en una línea y eso rescató un skill que estaba muerto:

> *"El auditor es quien verifica el punta a punta de la solución. Que un modelo grande verifique
> que toda la solución construida corresponda y satisfaga a las historias de usuario que
> involucra. **Y que el modelo implementador no haya mentido.**"*

### Por qué estaba muerto, y no era el prefijo

`sf-audit` eran 261 líneas construidas **enteramente** sobre cuatro comandos que se tiraron
—`sf doctor --drift`, `sf coverage`, `sf verify`, `sf delta`— y su premisa explícita era
*"arrancá de `sf doctor --drift`, **no** de leer"*. Sin esos motores, el audit **era** leer: justo
lo que decía no hacer. Renombrarlo no lo arreglaba.

### Lo que lo separa del ㉑, y es de posición, no de rigor

```
sf-check (㉑)   UNA feature · contra SUS criterios · JUSTO al terminarla
sfx-audit       VARIAS features · punta a punta · CUANDO JAVIER QUIERE
```

El ㉑ revisa `f-2` el día que `f-2` termina, y después **nadie la vuelve a mirar nunca**. De ahí
salen tres cosas que estructuralmente no puede ver, y son el trabajo del auditor:

- `f-4` **rompió** un criterio de `f-1`;
- `f-2` y `f-3` pasaron solas y **no se integran**;
- un criterio marcado `cumple` que **hoy ya no es cierto**.

### El hallazgo: "no mintió" se puede CONTAR

Es lo que hace que esto no sea *"leé todo con un modelo grande"*. **Cuatro de las cinco preguntas
del auditor tienen una sola respuesta correcta**, así que las contesta `sf` (R1):

```
① ¿qué features entran?      roadmap.json + estado.json
② ¿qué criterios tienen?      los us-#
③ ¿qué dijo la revisión?      los revision.json archivados
④ ¿los tests que lo probaban  tareas.json + buscarlos en el repo
   SIGUEN EXISTIENDO?
⑤ ¿la suite pasa hoy?         test_cmd
```

**La ④ es la que atrapa la mentira, y no es una opinión:**

```
✗ us-1/CA-2 se dio por cumplido en f-1 y su test ya no existe: core_test.go::TestSinEstado
```

Nadie más en el flujo puede verlo: el ㉑ de esa feature ya pasó, y el de la siguiente mira otros
criterios. **Y el modelo grande queda libre para la única pregunta que sí es juicio:** ¿el código
hace de verdad lo que la historia pedía?

### Las dos decisiones de Javier

- **`sf audit` es un comando nuevo**, y no rompe H2. `sf context` sigue sin argumentos porque el
  estado, la feature y el lote **son deducibles**; acá el alcance **no lo es**: lo elige Javier.
- **El módulo no es un concepto nuevo.** El ⑩ ya agrupa por *"comparten solución técnica"*, así
  que **un módulo es un conjunto de features** — nombrarlas es nombrarlo. Un campo `modulo` en el
  roadmap sería un concepto que mantener y que ningún estado consume.

```bash
sf audit                 # todo lo que se construyó
sf audit f-1 f-2 f-3     # estas tres
sf audit --completo      # embebe el material
```

**Los hechos van ARRIBA del material**, y es a propósito: el que lee es un modelo que va a gastar
contexto abriendo archivos, y saber qué buscar cambia qué abre. Probado a mano contra un proyecto
Go real, en los dos sentidos — con la mentira (`exit 2`) y sin ella (`exit 0`).

---

## 11. Qué sigue

1. ~~El `CLAUDE.md` de cuatro líneas~~ — ✅ **en [`plantillas/CLAUDE.md`](../plantillas/CLAUDE.md)**.
   Un archivo, dos destinos: `sf install` lo copia como `CLAUDE.md` **y** como `AGENTS.md`.
2. ~~`sf init`~~ — ✅ **construido**, en `sf/internal/arranque/`. Probado a mano contra un
   proyecto Go real y contra uno sin manifiesto.
3. ~~El hueco 6~~ — ✅ **cerrado** (§7).
4. ~~El ⑯ — recomendar modelo~~ — ✅ **`Plan.Modelo` en `tareas.json`**, con la cadena de
   precedencia de tres niveles. Era el hueco 4 de `que-sobrevive.md` §13 y **el skill ya lo
   prometía sin que existiera**.
5. ~~El auditor~~ — ✅ **`sf audit` + `sfx-audit` reescrito** (§10).
6. **El mapa de modelos** (`~/.specforge/`), que cierra H1b y el `via: consola`. Es lo único que
   queda del diseño, y arrastra `sobre.mutantes()`.
7. **`sf install`** — el andamio. Hoy `plantillas/CLAUDE.md` existe y **nadie lo copia**.
8. **Apagar el producto viejo**: `cli/`, `AGENT.md`, `README*.md`, `INSTALL.md`, `docs/`,
   `examples/`. **Se referencian entre sí**, así que van juntos o no van.

---

## 12. La auditoría de promesas — un método que faltaba

Después del corte de luz, Javier pidió revisar qué faltaba de verdad. **No se había cortado nada a
mitad**, pero la revisión encontró **cinco inconsistencias** que la ronda anterior había dejado, y
todas del mismo tipo: **skills que prometen cosas que el binario no tiene.**

```
grep de los COMANDOS que los skills mandan a correr   →  contra los que main.go rutea
grep de las RUTAS que nombran                         →  contra las que define docs.go
```

Se habían auditado los punteros a **skills** borrados, pero no a **comandos** ni a **rutas**.

| Lo que se encontró | |
|---|---|
| `sf-plan` §⑯ pedía escribir en un campo **que no existía** | ✅ el campo existe |
| Los 9 utilitarios leían `specforge/context/*.md` — los tres archivos **que se fusionaron** | ✅ `.docs/constitucion.md` |
| `sfx-journal` tenía un `learnings.md` en el medio: **una segunda copia** | ✅ borrado — la fuente son los journals archivados |
| `sf journal add` · `sf check run` — comandos muertos | ✅ reemplazados |
| Las salidas propias de los utilitarios apuntaban a `specforge/context/{thinks,grills,triages}/` | ✅ compuesto → el artefacto del estado; standalone → donde el usuario pida |

> **La regla que queda:** un skill nombra un comando o una ruta, y eso es una **promesa
> verificable**. Después de escribir skills, los dos greps de arriba se corren siempre — es el
> único chequeo del proyecto que cruza el `.md` con el binario, y ninguna compuerta lo hace.
