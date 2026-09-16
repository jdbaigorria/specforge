# Los vecinos, método por método — qué falta, qué sobra, qué es sólo nuestro

**Fecha:** 2026-09-13 · **Branch:** `claude/specforge-alternatives-direction-ae84u8`

> **Qué es esto.** La tabla de [`primitivos.md`](primitivos.md) cruzada contra los tres repos que
> resuelven el mismo problema sin máquina: `obra/superpowers` (14 skills), `mattpocock/skills`
> (38) y `pstack` —éste ya exprimido y cerrado en [`pstack.md`](pstack.md) el 2026-09-11—.
>
> **Se leyeron los repos, no sus landings.** Clonados el 2026-09-13. Lo que dice "tienen X" tiene
> archivo. Lo que es lectura mía está marcado.

---

## 1. La comparación no es de skills, es de métodos

Ellos tienen 52 skills entre los dos y nosotros 27, y **el número no dice nada**: superpowers
mete tres skills en lo que acá es capa de arnés (worktrees, agentes en paralelo, subagentes) y
Matt mete ocho en escritura y setup. Comparar cantidades es comparar alcances distintos.

Lo que se compara es **el método**: entrevistar, investigar, encontrar el test que falta,
diagnosticar un bug. Eso sí es la misma pregunta en los tres.

---

## 2. Lo que ya está cubierto en los tres — y no hay nada que hacer

| método | `sf` | Matt | superpowers |
|---|---|---|---|
| entrevistar / grill | `sfx-grilling` | `grilling` · `grill-me` · `grill-with-docs` | `brainstorming` |
| investigar | `sfx-buscar` | `research` | — |
| prototipo descartable | `sfx-prototipo` | `prototype` | — |
| ciclo TDD | `sfx-tdd` | `tdd` | `test-driven-development` |
| escribir skills | `sfx-skill` | `writing-for-agents` | `writing-skills` |
| router de skills | `sfx-mapa` | `ask-matt` | `using-superpowers` |
| plan → tickets | `sf-plan` · `sfp-backlog` | `to-tickets` · `to-spec` | `writing-plans` |
| triage | `sfx-triage` | `triage` | — |
| verificar antes de cerrar | `sfx-verificar` | — | `verification-before-completion` |
| retro / bitácora | `sfx-journal` | `retro` | — |
| vocabulario del dominio | `sfx-vocabulario` | `domain-modeling` | — |

**Las dos últimas adopciones (`sfx-mapa` y `sfx-skill`, del `f828ec2`) cerraron los dos huecos que
quedaban en esta tabla.** Acá no hay deuda.

---

## 3. Los cuatro métodos huérfanos de `primitivos.md`, contra ellos

### 3.1 «¿Qué test falta?» — superpowers lo tiene, y escrito al revés

`superpowers/skills/test-driven-development/writing-good-tests.md`:

> *"Antes de escribir el cuerpo del test, contestá: **¿qué cambio en producción debería hacer
> fallar este test — y ese cambio es un bug o una decisión?**"*

Es **el mismo primitivo por el otro lado**:

| | pregunta | sirve para |
|---|---|---|
| `sf-plan` ⑯ · `sf-build` ⑲ | ¿qué podría estar mal y aún así pasar? | **encontrar** el test que falta |
| superpowers | ¿qué cambio debería hacer fallar este test? | **justificar** el test que ya escribiste |

Las dos caras del mismo método, y ellos lo tienen **en un archivo aparte que se carga cuando hace
falta** — que es exactamente la forma que acá le falta.

**Y traen una mecánica concreta que no tenemos:** la *mirror assertion*. Un `expected` calculado
por el mismo código que se está probando pasa siempre, haga lo que haga el código. La regla que
sale es *"derivá la expectativa a mano, con literales"*.

> **Adopción directa.** El primitivo que falta (`sfx-mutante`) nace con las dos caras y con la
> trampa del espejo adentro. No hay que inventarlo: hay que traducirlo.

### 3.2 La vara antes de las opciones — **no la tiene nadie**

Se buscó `rubric`, `scorecard`, `criteria before`, `weigh` en los dos repos: no hay nada.
`brainstorming` de superpowers genera opciones; los ADR de Matt registran la decisión **después**.

**Ninguno de los dos escribe la vara antes que la primera opción**, que es justamente el punto de
`sf-plan` ⑫a: *"una vara escrita después describe la opción que ya elegiste"*.

> **Esto es nuestro y no está en ninguno de los tres.** Es el candidato más fuerte a primitivo
> propio, y el único de los cuatro donde no hay nada que copiar.

### 3.3 El criterio que se puede testear — parcial en Matt

`to-tickets` arma tickets *"each declaring…"*, pero la regla dura —R5, *"si no se puede escribir
el test, no es un criterio"*— con su conteo en la compuerta, no está. Acá hay más, no menos.

### 3.4 La procedencia — Matt la tiene en UN lugar; nosotros en cuatro

`research/SKILL.md:10`: *"investigá la pregunta contra **fuentes primarias** (docs oficiales,
código fuente, specs, APIs de primera mano), no una secundaria"*.

Una frase, un dueño. Acá la misma idea está repartida entre `sfx-buscar`, `sfp-po:49`,
`sf-check:52` y `evidencia.json`. **No falta el método: falta que tenga un solo dueño.**

---

## 4. Los dos agujeros — métodos que ellos tienen y acá no existen

### 4.1 🔴 Diagnosticar un bug — **lo tienen los dos, y acá no hay nada**

```
Matt          diagnosing-bugs   "loop de diagnóstico para bugs difíciles y regresiones"
superpowers   systematic-debugging   "ante cualquier bug, ANTES de proponer un arreglo"
```

Y en `sf` el camino del bug es un carril entero de la máquina:

```
tipo: bug  →  implementar  →  cierre
```

Se salta la planificación —correcto— y se cae directo en `sf-build`, que es un skill de
implementar contra un plan que en este camino no existe. Lo que reemplaza a ese plan es el
diagnóstico.

> ## 🔴 CORRECCIÓN — 2026-09-16
>
> **Este apartado decía que el método no existía acá, y que `sfx-triage` "clasifica el ticket; no
> diagnostica el defecto". Es falso**, y basta abrir `skills/sfx-triage/SKILL.md` para verlo:
> tiene la ley de hierro (*"No fixes without investigation first"*), el rastreo hacia atrás desde
> el síntoma, el diff contra un caso que sí funciona, hipótesis rankeadas con evidencia a favor y
> en contra, un límite duro de 3, el test que falla **antes** del arreglo, `cannot-reproduce` como
> resultado válido, y la prohibición de aplicar el arreglo durante el triage.
>
> **El error se cometió leyendo `primitivos.md` §4② —que habla del triage como clasificador— sin
> cruzarlo con el archivo.** Es la misma forma de error que este repo ya se había anotado el
> 2026-09-13 en el traspaso §3: leer un documento sin verificarlo contra el código.
>
> **Y costó caro:** el 2026-09-15 se escribió un `sfx-diagnosticar` entero que duplicaba a
> `sfx-triage`. Se borró el 09-16 y lo que aportaba de nuevo se le agregó al que ya estaba.
>
> **Lo que sí faltaba**, y ahora está: (a) una puerta que rutee un síntoma al triage
> (`sf-entrar-bug`); (b) instrumentar las costuras en un sistema de varias capas; (c) el segundo
> corte —3 **arreglos** fallidos es la arquitectura, distinto de 3 hipótesis rechazadas—; y (d)
> que el triage entregue candidatos **sin recomendar**, porque elegir es decidir contra una vara.
>
> Lo que quedó en pie del apartado: los dos vecinos coincidieron solos en que el debugging es un
> método propio y que va **antes** de tocar código. Nosotros también lo teníamos. Nadie lo había
> comprobado.

Y no es casual: es la puerta *"llego con un bug"* que `FUNDAMENTOS.md` ④ dejó sin abrir. Abrirla
le da a `sfx-triage` su segundo consumidor y resuelve su ⚠️ de `primitivos.md` §4②.

### 4.2 🔴 Recibir una revisión — la salida nueva no tiene método

`superpowers/receiving-code-review`: *"cuando recibís feedback de revisión, **antes** de
implementar las sugerencias, sobre todo si el feedback parece poco claro o técnicamente
cuestionable"*. O sea: el método para decidir si un hallazgo **es correcto**, y cómo se contesta
si no lo es.

Y acá pasó esto, ayer:

> El `f828ec2` ③ construyó la parada **ME TRABÉ EN LA REVISIÓN** a las tres rondas, y su salida es
> `sf dismiss <h-#> "motivo"`. El commit lo dice con todas las letras: *"acá es un desacuerdo
> sobre el mismo hallazgo… lo rompe alguien que decida si el hallazgo es portante."*

**La parada existe, el comando existe, y el método para decidir no existe.** Un `sf dismiss` sin
un método detrás es exactamente lo que R3 prohíbe del otro lado: una decisión sin criterio escrito.

> Es la adopción más barata de todas y la más urgente, porque tapa un hueco que se abrió **ayer**.

---

## 5. Lo que tienen y NO va — con el motivo

| ellos | por qué no |
|---|---|
| `using-git-worktrees` · `dispatching-parallel-agents` · `subagent-driven-development` | capa de arnés. Herdr y Orca lo hacen mejor, y R1 dice que `sf` no lanza |
| `handoff` · `claude-handoff` (Matt, dos skills) | **no nos hace falta, y el motivo es la tesis del repo**: el estado vive en `estado.json`, no en el chat. Ellos necesitan un skill para comprimir la conversación porque no tienen dónde dejarla |
| `codebase-design` · `improve-codebase-architecture` | son opinión de diseño. R3: la compuerta frena sobre un hecho, un juez opina |
| `resolving-merge-conflicts` · `setup-*` · escritura | fuera de alcance |

**La fila del `handoff` es la más informativa de la tabla.** Dos de sus 38 skills existen para
resolver un problema que acá no se puede tener. Eso es lo que compra la máquina de estados.

---

## 6. Y un destino para `sfx-explain`

`primitivos.md` §4① lo midió con **cero consumidores**. Matt tiene `teach` —*"enseñale al usuario
una skill o un concepto"*— y es **user-invoked**: nadie la compone, la elige una persona.

Ése es el destino correcto de `sfx-explain`: puerta de usuario, sin gatillos en la descripción.
Hoy es model-invoked y compite en la carrera de ruteo sin que nadie lo necesite.

---

## 7. El saldo

```
YA CUBIERTO      11 métodos · nada que hacer
ADOPTAR          2  · la doble cara del test que falta (+ mirror assertion)
                    · recibir una revisión  ← tapa un hueco abierto AYER
ABRIR            1  · diagnosticar un bug — el carril y EL MÉTODO existían;
                    faltaba la puerta. Ver la corrección del 09-16 en §4.1
UN DUEÑO SOLO    1  · la procedencia, hoy repartida en cuatro
SÓLO NUESTRO     1  · la vara antes de las opciones — no está en ninguno de los tres
DESCARTAR        7  · con el motivo escrito, arriba
```

**Tres adopciones y una apertura.** Y de los cuatro huérfanos de `primitivos.md`, dos se resuelven
copiando, uno es organización interna, y uno hay que inventarlo porque no existe afuera.
