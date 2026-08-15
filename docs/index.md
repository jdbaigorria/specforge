# Documentación de SpecForge

**SpecForge no te dice cómo trabajar: ejecuta cómo trabajás vos.**

`sf` es una máquina de estados que sabe **dónde estás**, **qué sigue** y **si podés avanzar**. Tu
agente le pregunta, hace el trabajo, y vuelve a preguntar.

---

## Por dónde empezar

| Si querés… | Leé |
|---|---|
| **instalarlo** | [`../INSTALL.md`](../INSTALL.md) |
| **entender de qué se trata en 5 minutos** | esta página, abajo |
| **hacer tu primer producto de punta a punta** | [`primeros-pasos.md`](primeros-pasos.md) |
| **saber qué hace cada comando** | [`comandos.md`](comandos.md) |
| **saber por qué `sf` te frenó** | [`problemas.md`](problemas.md) |
| **entender los 9 estados y sus compuertas** | [`estados.md`](estados.md) |
| **saber qué archivos se crean y con qué forma** | [`artefactos.md`](artefactos.md) |
| **saber qué hace cada skill** | [`skills.md`](skills.md) |
| **entender POR QUÉ está diseñado así** | [`../specs/`](../specs/) |

---

## La idea, en cinco minutos

### El problema

Trabajás con un agente. El agente es bueno, pero:

- te olvidás de lanzar cada fase a mano, y hay que empujarlo;
- el commit no se hace, o se hace mal agrupado;
- no crea la branch de la feature;
- **con el contexto al 50% empieza a escribir mocks** que parecen código;
- dice *"terminado"* con media historia hecha;
- dice *"todo verde"* y no hay un solo test.

Ninguno de esos problemas se arregla con un prompt mejor. Se arreglan con **alguien que
compruebe**.

### La solución, en una línea

> `sf` expone la máquina de estados que guía al harness — y es el árbitro que decide si se puede
> avanzar.

Son **dos verbos**, y el segundo es el que importa:

| | Qué hace | Sin esto sería… |
|---|---|---|
| **expone** | *"estás en el ⑬, ahora toca el ⑭"* | una lista de tareas |
| **comprueba** | *"no terminaste: falta la branch"* | una sugerencia que el agente puede ignorar |

`sf` **no maneja el auto** —no agarra el volante ni elige la ruta— pero **da verde o rojo, y en
rojo no se pasa**.

### El reparto

```
sf       →  DÓNDE estás · QUÉ sigue · ¿PODÉS avanzar?
skills   →  CÓMO se hace cada paso
harness  →  lo HACE
vos      →  DECIDÍS   (en tres puntos, y sólo tres)
```

**El orquestador es el agente, no el CLI.** `sf` arranca, contesta y se muere: dura
milisegundos. Un programa muerto no puede lanzar a nadie — no tiene manos.

### El bucle

```
ORQUESTADOR
   ├─ sf next                      → "implementá, con sf-build, en un subagente"
   ├─ lanza el subagente
   │     └─ sf context · trabaja · sf done · muere
   ├─ vuelve el control
   ├─ sf next                      → "ahora revisá"
   ├─ lanza el revisor … vuelve
   └─ … hasta que sf diga "PARÁ, esto lo decidís vos"
```

Y eso es **todo lo que el orquestador sabe**. No se sabe los 9 estados ni qué skill va con cuál:
lo pregunta cada vez.

---

## Las tres reglas duras

**1. `sf` nunca lanza a nadie.**
Si spawneara, manejaría contexto y tool-calling — y **sería un harness**, reinventando lo que
Claude Code y Codex ya hacen bien.

**2. El estado avanza con hechos comprobados, nunca con la palabra del que trabajó.**
El subagente dice *"terminé"*; `sf` **no le cree**: corre los tests él, mira si existe el archivo,
mira si hay branch. Si falta algo, el estado **no se mueve** y el orquestador recibe qué falta.

**3. El estado vive en el repo, no en el chat.**
Un `estado.json` versionado. Porque cuando pasás a otro modelo, el que implementa **no estuvo en
la conversación**.

---

## Qué NO promete

Esto importa tanto como lo que sí:

- **Calidad de código.** `sf` garantiza estructura, secuencia y evidencia — no que el diseño sea
  bueno. **Una compuerta frena sobre un hecho; un juez opina**, y `sf` no opina nunca.
- **Que el modelo no alucine.** Produce artefactos que lo hacen alucinar *menos*, y comprueba lo
  comprobable.
- **Decidir por vos.** La IA propone, la última palabra es tuya. Siempre. Por eso una dependencia
  fuera de la constitución **avisa** en vez de frenar: una herramienta que frena sola rompe esa
  regla.
