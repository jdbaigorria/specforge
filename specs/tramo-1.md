# T1 refundado — el brief, en cuatro primitivos

**Fecha:** 2026-09-07 · **Branch:** `refundation` · **Commit base:** `6880bd8`

> **Qué es esto.** El replanteo del primer tramo (①–⑥) después de leer las skills de Matt Pocock.
> No reemplaza a `el-mapa.md` —ése sigue siendo el estado del proyecto— sino a la ficha de T1 que
> está ahí en §3 y a la de `por-tramos.md` §5.
>
> **Es corto a propósito.** Lo largo son las skills, y las skills son chicas.
>
> **Regla de lectura.** Lo que dice "el código hace X" tiene archivo y línea, medido el 2026-09-07.
> Lo que es deducción está marcado.

---

## 1. Por qué se replantea

T1 corrió el 2026-09-05 y falló. La causa raíz está en `el-arnes-ciego.md`: **el arnés estaba
ciego**. Pero al ir a arreglarlo aparecieron cuatro arreglos sueltos —el nivel 0, `evidencia`, el
acta, `sf doctor`— y ninguno explicaba por qué el tramo tenía esa forma.

Las skills de Matt contestan eso, y con una sola idea:

> **La entrevista tiene forma mecánica.** No es *"entrevistá bien"* — es un **árbol de decisiones**,
> una **frontera** (las preguntas cuyos prerequisitos ya están resueltos) y **rondas**. Y el corte
> no es de gusto: **termina cuando la frontera queda vacía.**

Lo que hoy dice `sfp-scout` es *"Interview properly. Do not rush this."* — una sugerencia. Y la
regla del proyecto ya está escrita:

> **TODA REGLA QUE VIVE SÓLO EN EL SKILL ES UNA SUGERENCIA. SÓLO LA COMPUERTA OBLIGA.**

**Entonces la regla de esta refundación, y es la que ordena todo lo de abajo:**

> **De Matt se copia la FORMA. Pero cada cosa que se copia tiene que terminar en un archivo que la
> compuerta pueda CONTAR.** Un árbol con la frontera vacía **es** un archivo donde no quedan
> preguntas abiertas. Eso se cuenta.

---

## 2. La forma nueva

```
HOY          ① pinponeo → ② research → ③ hueco → ④ grill → ⑤ brief.md → 🛑⑥
             una skill hace las cinco cosas · todo prosa · un solo archivo

T1           ①–⑤ = UNA entrevista por rondas de frontera
                     │
                     ├── falta un HECHO      → sfx-buscar     (no bloquea: sólo
                     │                                          espera lo que cuelga)
                     ├── falta una PALABRA   → sfx-vocabulario
                     ├── falta una PRUEBA    → sfx-prototipo  (lo aprueba Javier)
                     ├── falta una FORMA     → sfx-think      (la rama rara: buscaste,
                     │                                          encontraste, y las
                     │                                          opciones siguen empatadas)
                     │
                     ▼  frontera vacía
             entrevista.md · evidencia.md · vocabulario.md · brief.md
                     │
                     ▼
             🛑 ⑥ con acta:  COMPROBÉ / MEDÍ / NO PUEDO COMPROBAR
```

**Lo que se ganó sin pagar nada.** La pregunta ⑤① de `el-mapa.md` —*"si se parte el ①–⑤ y el grill
descubre que falta investigar, ¿el ② se reabre? ¿quién cierra ese bucle?"*— **desaparece**. No hay
un ② al que volver: la investigación es una rama del árbol, no una etapa. Una búsqueda corriendo es
*un prerequisito sin resolver*, así que sólo espera lo que cuelga de ella. **No nace ningún bucle
nuevo**, y ése era el único motivo por el que el corte del ② estaba diferido.

---

## 3. Los cuatro primitivos

**La regla de composición, copiada entera de Matt:**

1. **El primitivo es el dueño ÚNICO de su método.** El compositor no lo repite.
2. **Una skill user-invoked nunca invoca a otra user-invoked.**
3. **La dependencia se escribe como orden ejecutable** — `Call the Skill tool with "sfx-grilling"` —
   no como link ni como `/nombre` suelto en la prosa. Un modelo débil lee prosa y no llama nada.

| primitivo | model-invoked | dueño único de |
|---|---|---|
| `sfx-grilling` | sí | el método de entrevista: árbol, frontera, rondas |
| `sfx-buscar` | sí | el método de investigación: nivel 0 → nivel 1, procedencia |
| `sfx-vocabulario` | sí | el glosario del negocio: `vocabulario.md` |
| `sfx-prototipo` | sí | código descartable que contesta **una** pregunta |
| `sfx-think` | sí (ya existía) | debatir cuando ningún hecho desempata las opciones |

Y arriba de ellos, los compositores:

| compositor | qué agrega | qué NO hace |
|---|---|---|
| `sfp-scout` | el brief, el veredicto y el ⑥ | no explica cómo se entrevista |
| `sfx-grill-me` | nada — es la entrevista suelta, sin producto atrás | se achica a 3 líneas |

**El bug que esto arregla, medido:** `sfp-scout` Step 1 hace una entrevista **sin llamar al
primitivo de entrevista** —`sfx-grill-me` recién aparece en el Step 4— y donde compone lo dice en
prosa. El método está duplicado, y las dos copias ya divergieron: `sfx-grill-me` dice *"ONE question
at a time. Never batch"* y Matt ya cambió a rondas.

**Los dos ejes se combinan, no se eligen:**

```
a lo ancho   composición           skill llama skill        ← de Matt
a lo alto    descubrimiento        SKILL.md → references/   ← ya era de SpecForge
             progresivo
```

El descubrimiento progresivo vive **adentro** de cada primitivo. Un `references/` no lo puede
invocar otra skill sin copiarlo — y copiarlo es el bug de arriba.

**Medido:** ninguna de las 22 skills de SpecForge usa `disable-model-invocation`. La distinción
user-invoked / model-invoked **no está encodeada hoy en ninguna parte**.

**Y no se puede copiar tal cual — acá choca.** En Matt, "user-invoked" quiere decir *sólo la
persona tipeando `/nombre`*. En SpecForge la cadena tiene un eslabón más, y hay que decirlo bien:

```
sf next      NOMBRA la skill     maquina.Instruccion tiene un campo `Skill` (maquina.go:120)
             no la invoca        es un binario que imprime y sale
   ↓
un modelo    LA CARGA            el orquestador leyendo esa salida, o el hijo de `sf lanzar`
```

`disable-model-invocation: true` bloquea **el segundo renglón**, que es el único que carga skills.
Ponérselo a `sfp-scout` cortaría el flujo aunque `sf next` la haya nombrado bien.

**Lo que sí se toma, y ya está aplicado:**

- **un compositor nunca llama a otro compositor** — es la mitad que vale de la regla
- la descripción del compositor dice **"invocado por `sf next`, no por iniciativa propia"**

**Lo que queda abierto, y es deducción sin verificar:** si un modelo puede arrancar `sfp-po` o
`sfp-backlog` por su cuenta, salteándose la máquina. **Umbral que lo despierta:** verlo pasar una
vez en una corrida. Antes de eso, poner la marca es adivinar cuál es el arreglo.

---

## 4. Los artefactos, y quién lee cada uno

**El truco que evita inventar paquetes: un archivo, dos lectores.** Markdown con frontmatter
contable. El cuerpo lo lee Javier; el frontmatter lo cuenta la compuerta con `internal/frontmatter`,
que ya existe y que es exactamente como `compuerta.Brief` lee hoy el `veredicto`.

| archivo | cuerpo (lo lee Javier) | frontmatter (lo cuenta la compuerta) | cuándo se crea |
|---|---|---|---|
| `.docs/entrevista.md` | las rondas, tal cual pasaron | `rondas` · `preguntas` · `abiertas` | siempre |
| `.docs/evidencia.md` | los hallazgos con su link | `retrieved` · `model_prior` · `probado` · `links` | siempre |
| `.docs/vocabulario.md` | el glosario del negocio | `terminos` | **sólo si hay algo que escribir** |
| `.docs/brief.md` | el argumento | `veredicto` (ya existe) | siempre |
| `.docs/prototipos/<slug>/` | código descartable | — | **sólo si una pregunta lo pide** |

> **Sobre el nombre `vocabulario.md` y no `CONTEXT.md`.** Matt lo llama `CONTEXT.md`. Acá choca:
> `sf context` **ya es** el sobre que se le sirve al subagente. Dos cosas distintas con el mismo
> nombre, dentro de la misma familia — es el mismo choque que `Parte` vs el acta, y se resuelve
> igual: se renombra el que llega después.

**La regla de la creación perezosa, textual de Matt (`domain-modeling`):** *"Create files lazily:
only when you have something to write."* El glosario y el prototipo **se crean cuando hay algo que
escribir, no porque el tramo los pida.** Sin esto T1 pasa de un artefacto a cuatro y se convierte
en ceremonia, que es exactamente lo que hizo falta escribir `FUNDAMENTOS.md` para frenar.

### La procedencia gana una tercera etiqueta

```
retrieved      lo leí, con link                    ← ya existía
model-prior    me lo acuerdo, sin verificar        ← ya existía
probado        lo construí y lo vi funcionar       ← nueva: la deja sfx-prototipo
```

`probado` es la más fuerte de las tres y es lo único que sobrevive de un prototipo: **el código se
tira, la respuesta queda** — entra a `entrevista.md` como una decisión con procedencia `probado`.

---

## 5. Qué exige el ⑥, y qué no

**Exige (mecánico, contable, sin red y sin LLM):**

```
entrevista.md    existe · abiertas: 0
evidencia.md     existe · al menos una fuente retrieved CON link
brief.md         existe · veredicto ∈ {hacelo, pivotea, no-lo-hagas}
```

**No exige, y se dice de frente:**

- que el brief sea **bueno** — eso es juicio, y el juicio es de Javier
- que los links **existan** — la compuerta no sale a internet: **cuenta, no verifica**
- `vocabulario.md` ni el prototipo — son perezosos por diseño

**La excepción sigue siendo `evidencia: baja`**, y sigue siendo una declaración deliberada, nunca un
atajo.

---

## 6. La tercera salida ya existe — y está en el paquete equivocado

**Medido, `internal/lanzar/resumen.go` cabecera:**

> *"LO QUE EL ARNÉS NO DA, NO ESTÁ. No va en cero, no va en false, no va. Un costo de 0 y 'este
> arnés no dice cuánto costó' son dos cosas distintas, y escribir la primera cuando pasa la segunda
> es mentir en un archivo que alguien va a leer para decidir."*

**Eso ES el `NO PUEDO COMPROBAR` del acta**, ya escrito y funcionando con punteros. `lanzar` lo
tiene; `compuerta` no. No hay que inventar el concepto: hay que mudarlo.

**Y el otro hallazgo del mismo paquete:** `Resumen.Herramientas` (`resumen.go:48`) ya guarda las
herramientas que usó el modelo, sin repetir. Hoy eso sólo se imprime en `sf lanzar`
(`main.go:839`). O sea que *"¿el que investigó buscó de verdad, o contestó de memoria?"* **ya es
medible por `sf`**.

**La decisión, y es de Javier (Q8):** el que obliga es **el artefacto**, no el método de
lanzamiento. `evidencia.md` es obligatorio y la compuerta lo lee, se haya lanzado con `sf lanzar` o
con el subagente nativo del arnés. `lanzar.Resumir` es un **extra gratis** donde el arnés lo
permita, nunca un requisito — y donde no se pueda, el acta dice `NO PUEDO COMPROBAR qué
herramientas usó`, que **no** es lo mismo que "no usó ninguna".

---

## 7. La vara de T1 — cierra la contradicción de `el-mapa.md` §4①

```
por-tramos §4   "PASA si el estado.json quedó igual con los dos modelos"
retomar   §4    "pasa si B CITA, no si B dice hacelo"
```

**Se resuelve así, y es la excepción con nombre que faltaba:**

> **En T1 la vara es de EVIDENCIA, no de veredicto.** Pasa si, con los dos modelos:
> `entrevista.md` cierra con `abiertas: 0` **y** `evidencia.md` trae al menos una `retrieved` con
> link. **El veredicto no es la vara** — dos modelos honestos pueden leer lo mismo y decidir
> distinto, porque el único campo que se mueve en T1 (`producto.brief_sellado`) **es** juicio puro.
>
> De T2 en adelante los campos son mecánicos (`prd_hash`, `constitucion_sellada`) y ahí el criterio
> genérico de `por-tramos.md` §4 vuelve a servir.

---

## 8. El cableado en Go — HECHO el 2026-09-07

Las skills no obligan. Esto sí. **Los seis están hechos**, en `96f2dec` (①–⑤) y el commit del
`sf doctor` (⑥).

| # | Qué | Dónde | Tamaño |
|---|---|---|---|
| 1 | `docs.Entrevista`, `docs.Evidencia`, `docs.Vocabulario` | `internal/docs/docs.go` | trivial |
| 2 | `compuerta.Brief` lee los tres frontmatters en vez de un `grep` en la prosa | `internal/compuerta` | chico |
| 3 | `Resultado` gana la tercera salida y un lugar para los ✓ y los números | `internal/compuerta` | mediano |
| 4 | `Texto()` se parte en dos lectores: el subagente (sólo lo que falta) y **el acta** | `internal/compuerta` | chico si el 3 está |
| 5 | el sobre inyecta `vocabulario.md` como una ruta más, **si existe** | `internal/sobre` | chico |
| 6 | `sf doctor` ve las herramientas; si no las puede ver, lo dice | `internal/doctor` | mediano |

**El 5 tenía una trampa medida:** `sobre.go:160-170` — el sobre del brief está **vacío a propósito**
y de ahí en adelante cada estado enumera sus rutas a mano. Un `vocabulario.md` **no llega solo** a
los demás tramos. Se resolvió poniéndolo en `Armar()` y no caso por caso: agregarlo en cada rama
serían once lugares donde olvidarse, y olvidarse no daría error — daría un estado usando otra
palabra para la misma cosa, tres estados después.

**Lo que apareció al construirlo, y no estaba previsto:**

| | |
|---|---|
| `abiertas` tiene que ser un **puntero** | con un `int` pelado, no escribir el campo da cero, y cero es el valor que pasa. Olvidarse saldría **más barato** que cerrar la entrevista |
| el frontmatter **declara**, el cuerpo es el **hecho** | la compuerta frena sobre los `http` que hay y avisa cuando el número declarado no cierra. En la primera corrida a mano ya atrapó un `links: 13` con 2 links reales |
| el acta se engancha **sin preguntar el paso** | un `sf done` que pasa y NO mueve es, por definición, la máquina esperando una decisión que no puede tomar |
| los e2e se rompieron **otra vez** | escribían sólo `brief.md`, y la suite normal no los corre. Ya había pasado el 05-09. Ahora hay un helper `p.brief()` |

---

## 9. Lo que se decide NO hacer, con su umbral

| No se hace | Por qué | Qué lo despierta |
|---|---|---|
| ADRs junto al glosario | el `decision.md` del ⑫ ya es eso, y con tres opciones obligatorias | una decisión de producto (no de feature) que haya que justificar |
| que la compuerta verifique que los links existen | cero red y cero LLM en el camino de enforcement | nunca — es un límite de diseño, no una deuda |
| `sf buscar` como comando | superficie nueva grande; el primitivo `sfx-buscar` alcanza | una corrida donde el skill no baste y haga falta enforcement |
| exigir `vocabulario.md` en el ⑥ | perezoso por diseño; exigirlo es ceremonia | dos artefactos usando la misma palabra para dos cosas |
| que `sfp-scout` componga `sfx-think` como paso fijo | el ⑤ viejo lo llamaba **siempre**; en el árbol la mayoría de las preguntas las desempata un hecho, no un debate. Queda como **rama** de `sfx-grilling` | — ya resuelto: es rama, no paso |
| tocar `entradas.go` (brownfield) | sigue en pie de `el-mapa.md` §7 | T1 y T2 corridos enteros |

---

## 10. Dos observaciones que quedan anotadas — no son de T1, son de la máquina

**De dónde salen.** Del cotejo de T1 ya construido contra el flujo de `planning/matt.md` §4 y §7,
el 2026-09-07. **No se arreglan ahora, y el motivo es el mismo para las dos: son decisiones de la
máquina entera y todavía no hay una corrida larga que diga cuál es el problema real.** Inventarles
la solución ahora sería adivinar.

### ① La pizarra se llena, y no hay regla

`matt.md` §4, sobre la higiene de contexto:

> *"Los pasos 1 a 3 van en una sola ventana sin cortar… El límite es la **smart zone** (~150k
> tokens). Si la sesión se acerca antes de `/to-tickets`, no se sigue empujando en degradado: se
> compacta en el borde de fase más cercano."*

**Nuestro ①–⑤ es exactamente eso: una sola ventana sin cortar.** Y no hay ninguna regla escrita
para cuando se hace larga — ni en el skill, ni en `sf`.

**Y esta vuelta lo agrandamos nosotros.** El ①–⑤ pasó de una charla a un árbol con cuatro ramas,
hasta cuatro artefactos, y prototipos que se construyen en el medio. El riesgo no es teórico: es
consecuencia directa de este replanteo.

**Dónde muerde:** un ①–⑤ que se degrada al final produce las últimas ramas del árbol razonadas en
degradado — y `abiertas: 0` **no lo nota**, porque cuenta ramas visitadas, no la calidad con que se
visitaron. Es un verde que puede querer decir "las contesté cansado".

**Umbral que lo despierta:** la primera corrida de T1 donde la ventana se llene antes del ⑥. Ahí se
sabe si el corte va en el skill (comprimir y seguir), en `sf` (partir el ①–⑤ en dos paradas) o en
ningún lado.

### ② Los cortes entre etapas: Matt da cinco opciones, nosotros una y clavada

`matt.md` §7 pone cinco salidas en el borde entre dos fases —continuar, `/clear`, `/handoff`,
subagente, `/compact`— y dice que elegir entre ellas *"es la decisión más difusa del mapa"*.

Nosotros tenemos **una sola, escrita en Go**: el ⑥ para, y el ⑦ arranca en un subagente fresco con
su sobre.

**Y esto NO es un defecto — es la diferencia entre una colección de skills y una máquina.** Matt le
deja la decisión al humano en cada borde y paga en carga mental; SpecForge la toma una vez y la
cablea. Es la misma elección que el proyecto ya hizo en todos lados.

**El precio, dicho igual:** el día que la opción cableada no sea la correcta, no hay salida. Y hay
un caso concreto donde ya se ve: `matt.md` §4 envuelve el prototipo en `/handoff → /prototype →
/handoff de vuelta`. Nosotros tenemos **la mitad de vuelta** —la respuesta entra a `entrevista.md`
con procedencia `probado`— y **no dijimos nada de la de ida**: si `sfx-prototipo` corre en un
subagente, no tiene sobre. Corriendo inline no pasa nada; ahí está el límite.

**Umbral que lo despierta:** un tramo donde la parada cableada sea claramente la equivocada, o el
primer prototipo que haya que correr en un subagente.

---

## 11. Las decisiones de esta sesión

Las once, para que no haya que reconstruirlas de la conversación.

| | Decisión |
|---|---|
| Q1 | La entrevista **deja dato**, y queda registro para volver a leerlo |
| Q2 | La investigación es una **rama** de la entrevista, no una etapa — disuelve el bucle de ⑤① |
| Q3 | **Rondas de frontera**, no una pregunta por vez. Minimizar fricción |
| Q4 | **Sí al glosario** — y sirve a todos los tramos, no sólo a T1 |
| Q5 | **Sí al prototipo** — no sólo averigua: refuerza la respuesta |
| Q6 | Primitivo = **skill chica model-invoked**; el descubrimiento progresivo vive adentro |
| Q7 | **Un archivo, dos lectores**: markdown con frontmatter contable |
| Q8 | **El artefacto obliga**, no el método de lanzamiento. `Resumir` es un extra gratis |
| Q9 | `vocabulario.md` se inyecta como la constitución, **si existe**. Sin ADRs |
| Q10 | El prototipo es efímero, sin branch. El lenguaje sale de la cabecera que dejó `sf init` |
| Q11 | La vara de T1 es de **evidencia**, no de veredicto |
