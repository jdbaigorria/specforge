# Anexo — determinismo y máquina de estados

**Fecha:** 2026-08-08 · **Anexo de:** [`flujo-real.md`](flujo-real.md)
**Qué es esto:** insumo de diseño, **no** descripción del flujo. `flujo-real.md` es
deliberadamente libre de herramienta; acá sí se habla de qué construir.

**Origen:** una cita de Uncle Bob que trajo Javier, contrastada contra el flujo y los
incidentes ya relevados.

---

## 1. La cita

> *"Cuando uses agentes, asegurate de que **todo lo que pueda ser determinista se haga con
> una herramienta determinista**. No intentes hacer que los pobres agentes sigan un proceso
> determinista.*
>
> *No le digas a un agente que calcule la cobertura de código o la complejidad ciclomática.
> Hacé que el agente **ejecute herramientas deterministas** que calculen esas métricas.*
>
> *No le digas a los agentes que muten el código y busquen sobrevivientes. Ejecutá
> **probadores de mutación deterministas** que hagan eso.*
>
> *No le digas a un agente que siga un procedimiento. No le des una larga cadena de
> instrucciones if/else para seguir. Perderá el hilo de ellas. En su lugar, creá una
> **máquina de estados finita ejecutable** que dirija las acciones de los agentes."*

---

## 2. Lo que se acepta entero: la máquina de estados ejecutable

**Es la respuesta directa al dolor #1** de `flujo-real.md`: *"me cansaba de ir corriendo cada
fase"*.

> **Hoy Javier es la máquina de estados.** Él es quien mira en qué estado está, decide la
> transición y dispara el paso siguiente. Veintitrés veces por feature.

Y **la máquina ya está dibujada**: `flujo-real.md` tiene 23 pasos, 3 entradas, 3 puntos de
decisión, bucles con salida definida y un embudo. No hay que diseñarla — hay que
**ejecutarla**.

**La palabra que hace todo el trabajo es "ejecutable".** Una máquina de estados escrita en
prosa adentro de un skill **no es una máquina de estados**: es una cadena de if/else con otro
nombre, y el agente la va a perder igual. Ese es exactamente el mecanismo del cansancio.

Se registra como **aceptado sin reservas**.

---

## 3. Lo que se discute: generar mutantes

**Javier ya genera mutantes con el modelo (paso ㉒ de `flujo-real.md`) y le funciona.** El
consejo dice que no lo haga.

**La regla está mal cortada.** Lo que tiene que ser determinista es *"¿el suite mató a este
mutante?"* — eso es aplicar un parche, correr un programa y mirar un exit code. Pero
**generar el mutante es un juicio**, y verificarlo no exige regenerarlo.

### La regla más filosa

> **Si verificar X exige volver a derivar X → determinista.**
> **Si verificar X es ejecutar algo barato → lo puede generar el agente.**

Y **explica los tres ejemplos de la cita en vez de contradecirlos**:

| | Verificarlo es… | Veredicto |
|---|---|---|
| cobertura | recomputar la cobertura | **determinista** — coincide |
| complejidad ciclomática | recomputarla | **determinista** — coincide |
| *score* de mutación | volver a correr los mutantes | **determinista** — coincide |
| **generar** un mutante | aplicarlo y correr los tests | **el agente puede** — no coincide |

### Por qué el mutante generado por LLM es un caso especialmente bueno

- **Es un parche ejecutable.** No hay que creerle nada: se aplica y se corre.
- **Nace apuntado a la historia de usuario**, así que la atribución a un requisito sale
  gratis. Ninguna herramienta genérica de mutación sabe qué es una historia.
- **No necesita una herramienta por lenguaje.** Es el adaptador más caro de evitar.

**Condición innegociable:** el agente **genera**; el que **aplica, corre y reporta** no puede
ser el mismo que después afirma el resultado sin que nadie lo mire. Eso es ejecución, y la
ejecución es del lado determinista.

---

## 4. El límite: el determinismo es necesario, no suficiente

Los seis incidentes reales de `flujo-real.md`, contra la vara del determinismo:

| Incidente | ¿Lo agarra una herramienta determinista? |
|---|---|
| *"todo verde"* sin que haya tests **(Grok)** | **sí**, trivial |
| librería fuera de la constitución | **sí**, trivial |
| commits o branch que no se hicieron | **sí**, trivial |
| métodos que eran **mocks** (contexto al 50%) | **a medias** — *"¿es un stub?"* es mecánico; *"¿es un stub legítimo?"* no |
| tests salteados reportados como corridos | **sí** — el reporte por test lo dice |
| **feature a medio hacer (DeepSeek)** | **NO.** Completitud contra una historia es juicio |

**Tres de seis se resuelven con determinismo puro.** Y el que no se resuelve es del mismo
tipo que **la verificación más valiosa del flujo**: el paso ㉑ —*"¿esto satisface la
historia?"*— que ningún test verde ni ninguna métrica contesta.

> **Conclusión:** la máquina de estados resuelve el dolor #1 y **no puede reemplazar al ㉑**.
> Lo que sí puede es **garantizar que el ㉑ corra siempre, con material que no se inventó**.
> Determinismo para lo que se puede fabricar; juicio para lo demás — y el juicio con
> evidencia comprobable a la salida.

---

## 5. La restricción que el consejo no cubre: el camino corto

`flujo-real.md` registra que Javier **ya se sale del carril** cuando el peso no se justifica:

> *"a veces ni siquiera uso una propuesta, le digo a mi modelo del día a día: hay que
> arreglar tal problema, y que ejecute los tests y verifique que esté solucionado"*

**Una máquina de estados rígida que no lo contemple va a ser esquivada igual que ahora** — y
ahí se pierde el rastro entero, que es peor que no tener máquina.

> **El camino corto tiene que ser un estado de la máquina, no una excepción a la máquina.**

Con sus dos condiciones, que Javier ya mantiene por su cuenta: **correr los tests** y
**verificar que quedó resuelto**.

---

## 6. Qué queda decidido y qué queda abierto

**Decidido:**

1. La máquina de estados **ejecutable** es la respuesta al dolor #1. Prosa en un skill no
   cuenta.
2. Cobertura, complejidad y ejecución de mutantes van **deterministas**, vía herramienta
   existente. No se reimplementan.
3. **Generar** mutantes lo puede hacer el agente. Aplicarlos y correrlos, no.
4. El camino corto es **un estado**, no una excepción.

**Abierto:**

- ¿Qué estados tiene la máquina exactamente? El dibujo de `flujo-real.md` tiene 23 pasos,
  pero pasos ≠ estados: hay que decidir dónde están los cortes reales.
- ¿Qué pasa cuando la máquina detecta que un paso no ocurrió (el commit, la branch)? ¿Lo
  hace ella, avisa, o frena?
- El caso del **contexto al 50%** es un indicador **anticipado**, no una verificación
  posterior. ¿La máquina lo mira? Es lo único del relevamiento que permite actuar **antes**
  del daño, y no encaja en la dicotomía determinista/juicio: es una condición de operación.

---

## 7. Nota de procedencia

Este anexo salió de una cita traída por Javier, **contrastada contra datos propios ya
relevados** (el flujo y los seis incidentes), no adoptada por autoridad. Las dos secciones
donde se discrepa —§3 y §4— se apoyan en su práctica medida, no en una preferencia.

En `inspiration/bob_uncle.md` hay material del mismo autor de la lectura anterior, por si
conviene contrastar.
