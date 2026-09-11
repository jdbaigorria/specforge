# Los skills

Veinticinco. **Nueve saben hacer un estado; dieciséis son utilitarios que no saben que la máquina
existe.**

```
sfp-*   los 5 estados de PRODUCTO    corren una vez
sf-*    los 4 estados de FEATURE     corren una vez por feature
sfx-*   utilitarios                  fuera de la máquina — sf NO los conoce
```

**El prefijo no es decorativo**, y la tercera línea es la que muerde: un `sfx-` es standalone —no
pide sobre, no avisa que terminó—, así que si `sf next` devolviera el nombre de uno, el
orquestador lanzaría a alguien que **nunca va a llamar a `sf done`**, y el estado no se movería
nunca.

---

## Los nueve de estado

| Estado | Skill | Qué hace |
|---|---|---|
| `brief` | **`sfp-scout`** | pinponea la idea, busca si existe, encuentra el hueco, la estresa |
| `prd` | **`sfp-po`** | brief → actores, capacidades, restricciones, alcance |
| `constitucion` | **`sfp-constitucion`** | arquitectura, stack y por qué, convenciones, reglas |
| `backlog` | **`sfp-backlog`** | PRD → historias con criterios que tienen id |
| `roadmap` | **`sfp-roadmap`** | agrupa en features y ordena. Sólo ids y orden |
| `planificacion` | **`sf-plan`** | 3 opciones → spec → tareas → tests → modelo, en una pasada |
| `implementar` | **`sf-build`** | un lote: tests primero, rojo, código, commit |
| `revision` | **`sf-check`** | un veredicto por **cada** criterio + los mutantes |
| `cierre` | **`sf-cierre`** | la doc (dos mitades) y el journal |

**Los nueve tienen la misma cirugía**, y es lo que los hace de la máquina:

```
+  arrancan con   sf context      pedí tu sobre, no busques qué leer
+  terminan con   sf done         avisá que terminaste
−  sin convenciones propias       salen de la constitución
−  sin modo / carril / fase       la máquina saltea estados
```

> **Y no saben en qué estado están.** Por eso `sf context` no lleva argumentos: el skill dice
> *"dame mi sobre"*, y **cuál es el sobre lo decide `sf`**. Un mismo `.md` puede servir a dos
> estados sin enterarse.

---

## Los dieciséis utilitarios

Sirven **solos**, en cualquier proyecto, con o sin SpecForge. Invocalos cuando quieras.

| Skill | Qué hace |
|---|---|
| **`sfx-grilling`** | **el primitivo de entrevista**: árbol, frontera, rondas. Dueño único del método |
| **`sfx-buscar`** | **el primitivo de investigación**: nivel 0 sin llave antes que los MCPs, con procedencia |
| **`sfx-vocabulario`** | **el primitivo del glosario**: `vocabulario.md`, perezoso |
| **`sfx-prototipo`** | **el primitivo de la prueba**: código descartable que contesta UNA pregunta |
| **`sfx-verificar`** | **el primitivo de la verificación**: genera el skill que maneja la app de verdad, y el mapa de features. Es lo que hace alcanzable el escalón 5 del ㉑ |
| **`sfx-prosa`** | **el primitivo de la prosa**: reglas `P-#` numeradas y citables, derivadas del castellano de este repo |
| **`sfx-interrogar`** | varios modelos sobre el mismo diff: lo que dos encuentran solos es la señal alta |
| **`sfx-think`** | debatir una idea, explorar opciones antes de comprometerse |
| **`sfx-grill-me`** | la entrevista suelta — compone `sfx-grilling` |
| **`sfx-tdd`** | RED → GREEN → REFACTOR, una rebanada por vez |
| **`sfx-github`** | branch, commit, push, PR, merge |
| **`sfx-documenter`** | documentación exhaustiva desde el código |
| **`sfx-journal`** | capturar aprendizajes anclados en evidencia |
| **`sfx-triage`** | investigar un bug hasta la causa raíz |
| **`sfx-explain`** | explicar un concepto con el método Feynman |
| **`sfx-audit`** | el auditor punta a punta — ver [`comandos.md`](comandos.md#sf-audit) |

---

## El patrón que los une: composición

Un skill de estado es **delgado y compone utilitarios**. El método vive en el `sfx-`; el skill de
estado sólo sabe hablar con la máquina.

```
          el skill de ESTADO             los UTILITARIOS
          delgado · habla con sf         el método · standalone
          ─────────────────────          ─────────────────────
brief     sfp-scout        compone  →    sfx-grilling (①–⑤)
                                          └ y sfx-grilling ramifica solo:
                                            sfx-buscar · sfx-vocabulario
                                            sfx-prototipo · sfx-think
prd       sfp-po           compone  →    sfx-prosa (antes de sellar)
backlog   sfp-backlog      compone  →    sfx-triage (cuando es un bug)
planif.   sf-plan          compone  →    sfx-think (⑫)
implem.   sf-build         compone  →    sfx-tdd · sfx-github
revisión  sf-check         recibe   →    el skill que dejó sfx-verificar, si existe
cierre    sf-cierre        compone  →    sfx-documenter · sfx-journal · sfx-github · sfx-prosa
```

> **El `revision` dice "recibe" y no "compone", y la diferencia importa.** `sf-check` no invoca a
> `sfx-verificar`: lo que recibe es **el archivo** que ese utilitario dejó en `.docs/verificar/`,
> servido por el sobre. `sfx-verificar` se corre una vez por proyecto, a mano, y el ㉑ de cada
> feature cosecha lo que dejó.

**Los `sfx-` no se tocan**, y hay motivo: un utilitario que aprendiera a llamar a `sf done`
dejaría de servir fuera de un proyecto SpecForge — **que es la mitad de su valor**.

---

## Cómo los invoca el orquestador

No los elige él: **se los dice `sf next`**.

```
skill:   sf-build
modelo:  sonnet
via:     subagente
```

Y con eso el orquestador lanza. **Los tres campos juntos son lo que deja `CLAUDE.md` en cuatro
líneas** — si `sf` no los devolviera, la tabla `estado → skill → modelo` viviría en el
orquestador y habría que mantener **una copia por harness**.

---

## Si escribís un skill nuevo

**Para un utilitario (`sfx-`):** hacé lo que quieras. No toca la máquina.

**Para un estado**, tres reglas y el CI las verifica:

1. **El `name:` del frontmatter tiene que ser igual al nombre de la carpeta**, o no lo encuentra
   nadie.
2. **El prefijo tiene que ser el correcto** — `sfp-` producto, `sf-` feature. **Nunca `sfx-`**.
3. **Sólo nombrá comandos que `sf` tenga.** El CI corre exactamente ese chequeo, y no es
   burocracia: es el único del proyecto que **cruza el `.md` con el binario**, y ninguna compuerta
   lo hace.

```bash
# los cuatro chequeos que corre el CI, a mano
grep -m1 '^name:' skills/mi-skill/SKILL.md      # ¿coincide con la carpeta?
grep -rn 'specforge/' skills/                    # la ruta es .docs/
```

Y acordate de la cirugía: **arranca con `sf context`, termina con `sf done`**, y no lleva
convenciones propias.
