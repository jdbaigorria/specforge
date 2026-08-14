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

FUERA
  sf-audit        → sfx-audit. Ningún estado lo consume
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

## 7. El hueco que quedó al descubierto — el sobre y lo archivado

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

**Queda como hueco 6** de `que-sobrevive.md` §13.

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

## 9. Qué sigue

1. **Renombrar el mapa en `maquina.go`** y sacar los tres ⚠ que este documento contesta.
2. **Escribir los nueve.** Dos desde cero (`sfp-po`, `sf-cierre`), siete adaptando.
3. **El hueco 6** — el sobre y `relacionado_a`.
4. Recién ahí, **el `CLAUDE.md` de cuatro líneas**: cae solo cuando los nombres son firmes.
