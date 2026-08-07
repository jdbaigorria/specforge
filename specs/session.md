# Sesión de refundación — estado

**Fecha:** 2026-08-06 / 07 · **Estado:** la decisión de arquitectura que estaba abierta
**quedó resuelta y aplicada** (§3 → §5.8–5.11). Lo pendiente ya no es diseño sino
**medición**: §4 y §6.

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
| Dos correcciones de Javier | ver §2 y §3 | **las dos aplicadas** — la segunda en `contract/audit.md` §0 y §2, `contract/judge.md` §0 |

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
> **Por qué no ganó ninguna: §3.6. Las decisiones que salieron: §5.8–5.11.**

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

> **Esta tabla quedó obsoleta — se deja como registro.** Preguntaba *quién computa cada
> hecho*, y esa nunca fue la pregunta que decidía nada. Ver §3.6.

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

### 3.5 La simplificación que caería con B

Si el skill es el que trabaja, entonces **`sf-audit` y "el juez" son el mismo actor.**

Hoy están como dos: el CLI audita y después un juez opina sobre las sobras. Con B hay uno
solo — el skill auditor hace la revisión punta a punta (el Opus de la v1) y el CLI hace dos
cosas: le arma el material y le valida las citas del veredicto. `judge.md` dejaría de
describir un agente aparte y pasaría a ser **el contrato de salida del skill auditor**.

Un componente menos y la metodología vuelve a ser *"la IA trabaja, el CLI no la deja
mentir"*.

> **Aceptada, con una salvedad.** La fusión es correcta y está aplicada (`judge.md` §0.1),
> pero **no es "un componente menos"**: lo que no se puede colapsar son las dos garantías,
> porque tapan modos de falla distintos. El determinismo tapa la **fabricación**; el
> contexto fresco tapa el **razonamiento motivado**. Por eso `sf-audit` **debe** ser un
> subagente limpio y no la sesión que construyó el código.

### 3.6 Por qué no ganó ninguna de las dos

**El planteo mezclaba dos ejes independientes**, y por eso no cerraba:

| Eje | Pregunta | Dónde vive el argumento |
|---|---|---|
| **1** | ¿quién **ejecuta** la herramienta? | acá vive *"no reinventar la rueda"* |
| **2** | ¿quién **compone** el entregable? | acá vive *"que la IA trabaje"* |

A los bundleaba de un lado y B del otro. Separados, tres de las cuatro casillas se caen
solas:

- **F5 y F6 eran un falso desacuerdo.** Lo que B quería —que nadie escriba complejidad
  ciclomática en Go— se logra con `exec.Command("gocyclo", ...)` desde el CLI. Y al revés:
  **si el skill corre `mutmut` y reporta 71%, puede reportar 85%** — el mismo vector de
  fabricación que F3, sin ninguna diferencia estructural. F3 no era *la excepción*: era el
  caso general de "ejecutar y sellar".
- **La clasificación no es juicio, es una tabla de verdad** (`audit.md` §4). Para validar
  la clasificación del skill hay que aplicar la tabla, o sea **recomputarla**. Si la
  validás, ya la computaste — el skill sólo agregaba un vector de fabricación gratis.
- **El veredicto sí era de B, y ya estaba aplicado desde §2.** El día que el juez pasó a
  ver la feature entera, el entregable pasó a ser el juicio y la debilidad admitida de A
  (*"el skill es un envoltorio que imprime salida de CLI"*) dejó de aplicar.

**La resolución, en una línea:** `sf-audit` es un solo actor que revisa y firma; el CLI es
su **brazo**. Ver §5.8–5.11.

---

## 4. Lo que sí hay que medir

El experimento original de esta sección —correr `sf audit` en las dos versiones y
comparar— **se descartó**: no podía decidir nada. Las dos corren `mutmut`, las dos sacan
71%, las dos clasifican `R1` como `DÉBIL`. Lo único que separaba A de B era el riesgo de
fabricación, y eso **no aparece nunca en una corrida cooperativa** — habría salido una
falsa confirmación.

**El experimento que sí vale** sale de la nota al pie de `validation/slugify.md` §4-H4
(*"las dos capas tienen agujeros, en lugares distintos"*):

> Correr el prompt de `judge.md` §4.1 sobre `slugify` **con `R1` en `PROBADO`** —sin el
> accidente del mutante de dígitos— y ver si el cerebro encuentra **H3** (el hueco de
> dígitos) y **H4** (`sep` sin requisito) **por sentido**.

| Resultado | Qué decide |
|---|---|
| Los encuentra | la capa de juicio cubre el agujero de F6; la corrección de §2 queda validada contra algo real |
| No los encuentra | las dos capas fallan juntas ahí, y hace falta la tercera (F5 sub-símbolo, hoy descartada) |

Es falsable, es barato, ejercita `record-verdict` (`judge.md` §5) contra un veredicto real,
y responde algo que hoy **nadie sabe**. Depende de P1.

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

*Las cuatro siguientes salieron de cerrar §3, el 2026-08-07:*

8. **`sf-audit` es un solo actor: un subagente limpio que revisa punta a punta y firma el
   informe. El CLI es su brazo.** `judge.md` deja de describir un componente aparte y pasa
   a ser el contrato de salida de `sf-audit`. Lo que **no** se colapsa son las dos
   garantías: determinismo (tapa fabricación) y contexto fresco (tapa razonamiento
   motivado). Por eso el cerebro **debe** ser subagente, no la sesión principal.
9. **El informe tiene dos canales, y cada línea es atribuible a uno.** Hechos del brazo
   (sellados, citados **textuales**) y juicios del cerebro (prosa con cita `path:line`).
   **Un número que el cerebro parafrasea deja de ser un hecho.**
10. **La regla que asigna cualquier responsabilidad futura, sin volver a debatirla:**
    *si validar X exige **recomputar** X, X es del brazo; si validar X es **resolver una
    cita**, X es del cerebro.* Deriva la condición 10 de `judge.md` §5, a la que se había
    llegado por intuición — buena señal de que la regla es la correcta.
11. **Un solo comando, y el brazo nunca reimplementa.** `sf audit --json` entrega hechos y
    material de una sola vez (absorbe a `sf context for-judge`): el cerebro no elige su
    propio examen, y una secuencia de pasos en prosa no es una garantía —depende de que el
    modelo la lea entera y no se saltee el tercero—. Todo análisis es **shell-out** a
    herramientas maduras; el brazo acota, normaliza y sella.
    - Corolario de poda: **un comando que ningún skill invoca en un punto de decisión está
      muerto, o le falta estar adentro de otro comando.** No hay tercera opción.
    - El perfil de rigor (§5.4) es lo que hace seguro el un-solo-comando: sin un eje que
      apague F6, bundlear todo sería un comando-dios sin freno.

---

## 6. Pendientes concretos

| # | Qué | Estado | Origen |
|---|---|---|---|
| **P1** | **Ningún ejemplo tiene `specforge/.state/`** — cero resultados de test sellados en los tres. `PRA-3` los usa como suite de regresión pero regresionan **artefactos, no evidencia**. Bloquea F3, F4 y el experimento de §4 | **abierto — es lo primero** | `validation/slugify.md` H1 |
| P2 | Agregar criterio `R1.2: slugify("Top 10 Songs") == "top-10-songs"` + su test | abierto | ídem H3 |
| P3 | Decidir si `specs/` se commitea o se gitignorea | **cerrado 2026-08-07: se commitea.** `planning/` sigue siendo scratch ignorado; `specs/` es la spec del producto | — |
| P4 | 155 commits sin pushear, último push 2026-05-21. Es `OPS-1` y es lo más barato del backlog | abierto | auditoría §7 |
| P5 | Las preguntas abiertas al pie de `audit.md` y `judge.md` — **se resuelven por uso, no por debate** | abierto por diseño | — |
| **P6** | Implementar el `toolchain` de `audit.md` §3.0/§9.2 — **Python primero** (es `slugify`), Go segundo (es SpecForge sobre SpecForge). No declarar soporte de un lenguaje sin un ejemplo que lo ejercite | abierto | §5.11 |

> **P1 y P6 son la misma puerta.** Sin `.state/` no hay hechos que regresionar, y sin
> toolchain no hay cómo producirlos. Es también el primer paso hacia lo que la auditoría
> marcó como el defecto estructural: **0 features de SpecForge hechas con SpecForge.**

---

## 7. Archivos

```
planning/                                  (gitignored — scratch local)
  AUDITORIA-FEATURES.es.md                 qué tenemos y de dónde vino
  SPECFORGE-V2-INSPIRACION.es.md           los 16 frameworks + arquitectura de módulos

specs/                                     (COMMITEADO 2026-08-07 — P3 cerrado)
  session.md                               este archivo
  contract/
    README.md                              índice + vocabulario normativo (cerebro/brazo)
    audit.md                               EL BRAZO · 7 hechos, 6 estados, toolchain, exit codes
    judge.md                               EL CEREBRO · contrato de salida de sf-audit
    validation/
      slugify.md                           el contrato aplicado a mano · 6 hallazgos

inspiration/                               (gitignored — 16 repos + bob_uncle.md + bench_yaml_json.md)
```

---

## 8. Para retomar

**Ya no hay nada de diseño abierto en la capa cross.** Los contratos están escritos,
validados a mano contra código real y libres de la contradicción que arrastraban. Lo que
sigue es medición y código:

1. **P1** — correr el pipeline de verdad sobre los tres ejemplos y commitear `.state/`.
   Habilita F3/F4 y el experimento de §4. Es también la primera vez que SpecForge corre
   sobre algo real de punta a punta.
2. **§4** — el experimento del cerebro sobre `slugify` con `R1` limpio. Responde si la
   capa de juicio tapa el agujero de F6, que es la última incógnita del contrato.
3. **P6** — el `toolchain` en Python, después Go.
4. **P4** — pushear. No es un debate, es un `git push`.

Recién después: **Forge**, y luego Inception y Ship (§5.2).

**Las invariantes que no cambian, pase lo que pase:** el bucle pedir → trabajar → entregar
→ validar → sellar; toda afirmación del cerebro con cita verificable; y **ejecutar es
siempre del brazo** — F3 no era la excepción, era el caso general.
