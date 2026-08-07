# Sesión de refundación — estado y tensión abierta

**Fecha:** 2026-08-06 / 07 · **Estado:** en pausa, con **una decisión de arquitectura sin
resolver** (§3). Todo lo demás quedó cerrado o escrito.

---

## 1. Dónde quedamos

El día arrancó con la sospecha de que SpecForge había perdido el rumbo y terminó con un
contrato escrito y probado contra código real. El recorrido:

| Paso | Resultado | Dónde quedó |
|---|---|---|
| Auditoría de features | 39 comandos, 16 skills; **43% de la superficie entró por leer otros repos**, sin usuario que lo pidiera | `planning/AUDITORIA-FEATURES.es.md` |
| Lectura de los 16 frameworks en `inspiration/` | de 16, sólo **3** tienen enforcement real (nWave, gentle-ai, SpecForge). **Nadie une los módulos con un solo ledger** — ese es el hueco | `planning/SPECFORGE-V2-INSPIRACION.es.md` |
| Decisión de arquitectura | SpecForge = **conjunto de módulos** (Inception + Forge + Ship + capa cross), un instalador, módulos instalables por separado | ídem §15 |
| Contrato de la capa cross | 7 hechos, 6 estados, el juez y cómo verificarlo | `specs/contract/audit.md` · `judge.md` |
| Validación contra `examples/slugify` | **encontró un bug real** en código con 7 gates aprobados; 6 hallazgos | `specs/contract/validation/slugify.md` |
| Dos correcciones de Javier | ver §2 y §3 | aplicada la primera; **la segunda está pendiente** |

---

## 2. La corrección que ya se aplicó — no sobre-indexar en determinismo

**Lo que Javier señaló:**

> *"Creo que estás obsesionándote a full con el determinismo y fue comprobado que ni así
> podemos tener garantías. Recordá que el CLI nació como una forma de crear artifacts
> útiles para los LLM para que estos alucinaran menos."*

**El error concreto:** el contrato original restringía la **entrada** del juez (sólo veía
el "residuo": los requisitos que las verificaciones mecánicas no pudieron decidir). La
validación probó que eso **creaba el punto ciego**: en `slugify`, el parámetro `sep` sin
requisito apareció **por accidente** — porque un mutante *no relacionado* dejó a `R1` en el
residuo. Con `R1` limpio, el juez nunca veía la función.

**La lección, generalizada:** verificar la **salida** de una capa cooperativa da seguridad.
Restringir su **entrada** no — sólo le saca visión. La pregunta correcta al diseñar es
*"¿puedo verificar lo que afirma?"*, nunca *"¿puedo limitar lo que ve?"*.

**Referencia histórica que dio Javier:** en la v1 de SpecForge, antes del determinismo,
corría un Opus con *"tengo esta propuesta, verificá que el código está completo y es
correcto, corré los tests para validar, y luego que el código cubre los requerimientos"*, y
hacía una revisión punta a punta. **Ese es el nivel de análisis que el juez debe tener.**

**Aplicado:** `judge.md` §0 y §3.1 — el juez ve la feature entera, el residuo lo orienta
pero no lo limita; puede señalar sobre un `PROBADO` con cita, pero no cambiarle el estado
(su hallazgo va **al lado** del hecho, no encima). `audit.md` §0 — *"determinismo para lo
que el modelo podría fabricar, juicio para todo lo demás"*.

---

## 3. ~~LA TENSIÓN ABIERTA~~ — RESUELTA 2026-08-07

> **✅ CERRADA. Ni A ni B: el planteo mezclaba dos ejes.** Lo que sigue queda como registro
> del debate; la resolución está aplicada en `contract/audit.md` §0 y §2 y en
> `contract/judge.md` §0. En una línea:
>
> **`sf-audit` es un solo actor —un subagente limpio que revisa y firma el informe— y el
> CLI es su brazo.** La pregunta no era *quién computa F5/F6* sino **de qué boca sale cada
> línea del informe**: los hechos los cita textuales del brazo, los juicios los escribe el
> cerebro con cita. Regla que decide cualquier caso futuro: *si validar X exige recomputar
> X, X es del brazo; si validar X es resolver una cita, X es del cerebro.*
>
> Pendiente de esta sección: reescribirla como decisión en §5 (item 5 del plan).

Las dos posiciones, planteadas en serio, sin que ninguna sea el hombre de paja de la otra.

### 3.1 Perspectiva A — la que se venía construyendo

**El CLI es la capa determinista y el skill la orquesta.**

`sf audit` es un comando que **computa**: resuelve símbolos, corre la mutación, calcula
complejidad, clasifica cada requisito y emite el veredicto. El skill invoca el comando,
muestra el resultado y despacha al juez para el residuo.

**Argumentos a favor:**

- **Reproducibilidad.** El mismo árbol da el mismo veredicto siempre, sin importar el
  modelo, la temperatura ni el harness. Es lo que permite ponerlo en CI.
- **Integridad del ledger.** El gate está encadenado por hash. Si el veredicto lo compone
  un LLM, el eslabón se vuelve irreproducible y la cadena pierde sentido.
- **La garantía sobrevive al harness.** Los hechos del CLI valen fuera de Claude Code; lo
  que vive en la prosa de un skill, no.
- **Testeable.** Los ~596 tests existentes prueban la capa Go. Un veredicto compuesto por
  un skill es mucho más difícil de regresionar.

**Su debilidad, admitida:** si el CLI computa todo, el skill se vuelve un envoltorio que
imprime salida de CLI. **La IA deja de trabajar**, y SpecForge termina siendo un linter con
interfaz de chat. Además empuja a reimplementar en Go cosas que ya existen (complejidad
ciclomática, mutación), contra la regla de no reinventar la rueda.

### 3.2 Perspectiva B — la de Javier

**El CLI es un conjunto de herramientas que los skills usan para minimizar la alucinación.**

Textual:

> *"Todos los repos que vimos usan skills para cada etapa, lo cual está bien. Nosotros nos
> diferenciamos en que esos skills siempre generan un artefacto que debe ser validado por
> el CLI, y cuando un skill necesita una parte o un artefacto, el CLI lo valida y se lo
> entrega."*

> *"Siento que te mandaste a querer hacer que todo sea un CLI. Me da la sensación de que
> asumiste que el CLI tiene más peso que el skill y se termina perdiendo la esencia de la
> metodología, que es que la IA trabaje."*

El bucle es: **pedir → trabajar → entregar → validar → sellar.**

**Argumentos a favor:**

- **Es la esencia de la metodología.** El trabajo lo hace la IA; el CLI existe para que no
  invente. Si el CLI hace el trabajo, no estamos haciendo SDD asistido por IA.
- **La diferenciación real está intacta.** Contra los 16 frameworks, lo distintivo no es
  *"el CLI analiza"* sino **"nada entra al estado sin validación y nada sale del estado sin
  que el CLI lo arme"**. El skill no lee archivos crudos ni escribe estado crudo. Eso es
  hermético y no requiere que el CLI piense.
- **No reinventar la rueda.** F5 y F6 los resuelven herramientas maduras por lenguaje
  (`radon`/`gocyclo`/`lizard`, `mutmut`/`go-mutesting`/`Stryker`). Escribirlas en Go es
  exactamente el error que Javier ya marcó como preferencia de ingeniería.
- **Escala a lenguajes.** Un CLI que analiza tiene que saber de cada lenguaje. Un CLI que
  valida y despacha, no.

**Su debilidad, a mirar:** si el veredicto lo compone un skill, ¿qué se sella en el ledger
y qué tan reproducible es? ¿Cómo se regresiona? El contrato ya tiene una respuesta parcial
(la cita `path:line` verificable), pero hay que confirmar que alcanza.

### 3.3 En qué coinciden — no hace falta discutirlo

- El skill **nunca** escribe estado directo; pasa por el CLI y se valida.
- El skill **nunca** lee estado crudo; el CLI se lo arma y se lo entrega.
- Toda afirmación del juez lleva cita `path:line` que el CLI resuelve; sin cita no se registra.
- **F3 (tests verdes sellados) es la excepción irreducible.** El CLI **debe ejecutar y
  sellar**. Si el skill corre los tests y el CLI valida "la salida dice PASS", el skill
  puede fabricar la salida. Es la primitiva anti-fabricación — el incidente que originó
  SpecForge.

### 3.4 Dónde difieren, concretamente

| Hecho | Perspectiva A | Perspectiva B |
|---|---|---|
| F1 anclaje | CLI computa | **CLI** — es validar que el trace dice la verdad |
| F2 prueba existe | CLI computa | **CLI** — ídem |
| F3 verde sellado | CLI ejecuta y sella | **CLI** — la excepción, ambos coinciden |
| F4 frescura | CLI computa | **CLI** — comparar hashes |
| F5 complejidad / CRAP | **CLI lo calcula en Go** | **skill** llama a `radon`/`gocyclo`; el CLI sella el resultado |
| F6 mutación | **CLI orquesta la herramienta** | **skill** llama a `mutmut`/`Stryker`; el CLI sella |
| F7 criterios | CLI computa | **CLI** — contar sobre el trace |
| **La clasificación y el veredicto** | **CLI los compone** | **el skill los compone**; el CLI valida y sella |

**El desacuerdo real son cuatro casillas: F5, F6, la clasificación y el veredicto.** El
resto ya coincide.

### 3.5 La simplificación que caería con B

Si el skill es el que trabaja, entonces **`sf-audit` y "el juez" son el mismo actor.**

Hoy están como dos: el CLI audita y después un juez opina sobre las sobras. Con B hay uno
solo — el skill auditor hace la revisión punta a punta (el Opus de la v1) y el CLI hace dos
cosas: le arma el material y le valida las citas del veredicto. `judge.md` dejaría de
describir un agente aparte y pasaría a ser **el contrato de salida del skill auditor**.

Un componente menos y la metodología vuelve a ser *"la IA trabaja, el CLI no la deja
mentir"*.

---

## 4. Cómo decidir entre A y B — el experimento

**No decidir por debate.** El método que funcionó hoy: aplicarlo a mano contra algo real.

Propuesta: **correr `sf audit` sobre `slugify` en las dos versiones** y comparar.

| Qué medir | Por qué decide |
|---|---|
| ¿El veredicto es el mismo? | si B llega al mismo resultado, la reproducibilidad extra de A no está comprando nada |
| ¿Cuánto código Go hace falta en cada una? | A implica escribir complejidad ciclomática y orquestación de mutación por lenguaje |
| ¿Se puede regresionar el veredicto de B? | es la debilidad declarada de B; hay que verla, no suponerla |
| ¿Qué se sella en el ledger en cada una? | si en B se sella el artefacto validado + las citas resueltas, el encadenado por hash sobrevive |
| ¿Qué pasa con un lenguaje nuevo? | A necesita soporte en Go; B necesita nombrar una herramienta |

**Hipótesis de partida (mía, para falsar):** B gana en F5, F6 y en el veredicto; A retiene
F1–F4, F7 y el sellado. O sea, probablemente el resultado sea **B con la excepción F3
explícita** — pero hay que verlo, no asumirlo.

---

## 5. Decisiones cerradas — no volver a discutir

1. **SpecForge es un conjunto de módulos, no dos productos.** Inception es un módulo.
   Usuario objetivo: el dev indie/freelancer que también es PM, analista y tester. Módulos
   instalables por separado (un PM puede instalar sólo Inception) pero comparten producto,
   vocabulario y ledger. Modelo `bundle` de spec-kit, **no** modelo `flow` de specs.md.
2. **Orden de trabajo:** capa cross (hecho) → Forge → Inception → Ship. Forge segundo
   porque es lo único construido: su contrato se escribe contra código real y se testea el
   mismo día.
3. **Ship = cerrar + autorizar + documentar. NO deploy.** El deploy es infra del usuario.
4. **Un solo eje de rigor** (`lean`/`standard`/`strict`) que decide qué etapas corren.
   Reemplaza los cinco `require_*`. Una garantía de calidad es **una etapa con dueño**, no
   un campo del schema del requisito.
5. **Formato:** YAML para lo que el humano configura, JSON/JSONL para estado y ledger.
   La tercera capa —lo que va al contexto del modelo— se **mide**, no se debate
   (es `EVA-1` reformulado y sí es respondible).
6. **El juez no tiene el verbo "aprueba".** Sólo `refuta_codigo`, `refuta_spec`,
   `no_refuta`, `no_juzgable`.
7. **`no_refuta` no se renderiza como ✓.** Significa "no encontré el problema", no "no hay
   problema".

---

## 6. Pendientes concretos, además de §3

| # | Qué | Origen |
|---|---|---|
| P1 | **Ningún ejemplo tiene `specforge/.state/`** — cero resultados de test sellados en los tres. `PRA-3` los usa como suite de regresión pero regresionan artefactos, no evidencia | `validation/slugify.md` H1 |
| P2 | Agregar criterio `R1.2: slugify("Top 10 Songs") == "top-10-songs"` + su test | ídem H3 |
| P3 | Decidir si `specs/` se commitea o se gitignorea. Hoy **no** está ignorado | — |
| P4 | 50 commits sin pushear, último push 2026-05-21. Es `OPS-1` y es lo más barato del backlog | auditoría §7 |
| P5 | Las preguntas abiertas al pie de `audit.md` y `judge.md` — **se resuelven por uso, no por debate** | — |

---

## 7. Archivos

```
planning/                                  (gitignored — scratch local)
  AUDITORIA-FEATURES.es.md                 qué tenemos y de dónde vino
  SPECFORGE-V2-INSPIRACION.es.md           los 16 frameworks + arquitectura de módulos

specs/                                     (NO gitignored — ver P3)
  session.md                               este archivo
  contract/
    README.md                              índice + vocabulario normativo
    audit.md                               los 7 hechos, 6 estados, salida, exit codes
    judge.md                               los 3 jueces, encuadre adversario, verificación
    validation/
      slugify.md                           el contrato aplicado a mano · 6 hallazgos

inspiration/                               (gitignored — 16 repos + bob_uncle.md + bench_yaml_json.md)
```

---

## 8. Para retomar

Leer §3 completo, decidir si el experimento de §4 vale la pena o si con leer las dos
posiciones alcanza. Si se elige **B**, hay que reescribir `audit.md` (de "comando que
computa" a "dispensador + validador") y fundir `judge.md` en el contrato de salida del
skill auditor. Si se elige **A**, los contratos quedan como están y sólo hay que mover F5 y
F6 a shell-out de herramientas existentes.

**Lo que no cambia en ninguno de los dos casos:** el bucle pedir → trabajar → entregar →
validar → sellar, la cita verificable, y F3 como excepción irreducible.
