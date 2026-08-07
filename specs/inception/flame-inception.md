# Flame — el módulo Inception

**Fase:** Flame · **Fecha:** 2026-08-07 · **Viene de:** [`brief-inception.md`](brief-inception.md)
**Doble propósito:** es un Flame real *y* es la primera prueba del formato de Flame.

---

## 0. Hallazgo al arrancar: el modo no se podía determinar

Inception es una feature de SpecForge, que ya existe → **modo feature**. Pero §5.1b del
brief dice que Spark debe preguntar *"¿esto entra en el PRD que ya está?"*, y **SpecForge no
tiene PRD**: nunca corrió Flame.

> **Modo feature presupone que modo proyecto corrió alguna vez.** En una adopción brownfield,
> la primera feature **no puede** correr en modo feature: no hay contra qué chequear.

**Cómo se resuelve, sin volver al barrido eager:**

El PRD de una adopción brownfield **no se deriva del código**. El código dice el *cómo*
(stack, patrones); sólo el humano puede decir el *qué* y el *por qué*. Es exactamente lo que
`skills/sf-init/SKILL.md` ya tiene escrito:

> *"The codebase tells you the how (stack, patterns) but not the why (principles, vision,
> anti-goals)."*

Y es **corto**: el PRD que §5.1b necesita no es una spec del codebase, es **el límite del
alcance** — qué es el producto y qué no es. Eso el humano ya lo sabe y se escribe en una
página. No hay que leer 18k líneas para escribirlo.

**Regla que queda:** al adoptar, el humano escribe un PRD de límite (corto). De ahí en
adelante, modo feature funciona. **Este documento lo escribe por primera vez** (§1).

---

## 1. PRD

### 1.1 Qué es

**Inception es el módulo de SpecForge que va de "se me ocurrió algo" a "esto es lo que hay
que construir y por qué".** Dos fases: **Spark** decide *si* y *por qué*; **Flame** decide
*qué exactamente*.

### 1.2 Qué NO es

Declarado para que §5.1b tenga contra qué chequear en el futuro:

- **No es un generador de documentos.** Si el resultado es un PDF lindo que nadie mira, falló.
- **No construye nada.** Todo lo de *cómo implementar* es de Forge. Preguntarlo acá es el
  error que hizo abandonar OpenSpec.
- **No documenta el codebase existente.** La ingesta es perezosa (brief §5.2).
- **No decide por el humano.** Objeta; el humano decide y la anulación queda escrita.

### 1.3 Para quién

**El dev indie que además es PM, analista y tester.** Es una sola persona con cuatro
sombreros, sin nadie a quien delegarle el "¿esto vale la pena?".

Consecuencia de diseño: **no hay handoffs entre roles.** Todo artefacto que exista sólo para
comunicar entre personas es overhead acá.

### 1.4 El problema

Hoy SpecForge arranca en `sf propose`: **asume que ya sabés qué construir.** Eso deja tres
agujeros medidos:

| Agujero | Evidencia |
|---|---|
| No hay puerta que pregunte **por qué** | **43%** de la superficie entró sin usuario que la pidiera; sólo ~10% de una falla observada |
| El trabajo de decidir **no deja rastro** | el flujo ya se hace a mano (Claude + Perplexity) y se pierde el porqué, lo descartado y lo objetado |
| No hay defensa de **alcance** | ninguna feature se preguntó nunca si corría el límite del producto |

### 1.5 Alcance del MVP

**El MVP es Spark solo. Flame queda fuera.**

Por qué esta línea y no otra:

- **Spark es donde está el valor medido.** El detector del 43% vive ahí, no en Flame.
- **Spark ya está validado**: se caminó dos veces (el propio brief, y `sf init` brownfield con
  veredicto `pivot`). Flame se está caminando **ahora, por primera vez**.
- **La salida de Spark ya tiene consumidor manual.** Javier hoy pasa el brief/PRD directo a
  `propose`. El MVP conserva ese paso manual y agrega **la única pieza que falta: el gate**.
- **Spark tiene base construida**: `sfp-scout` ya hace research con veredicto
  `proceed`/`pivot`/`kill` e IDs trazables.

Lo que entra:

| # | Qué | Nota |
|---|---|---|
| 1 | Un skill que conduce Spark y produce el **brief** | base: `sfp-scout` |
| 2 | **Campos obligatorios** del brief: qué · por qué anclado · prior art con cita · preguntas abiertas con fase · objeciones | el formato es la barandilla |
| 3 | Búsqueda de prior art **adentro** (índice de capacidades) y **afuera** (web) | brief §6.3 |
| 4 | Veredicto **`proceed` / `pivot` / `kill`** | reusa `sfp-scout` |
| 5 | Chequeo de límite: **¿entra en el PRD o lo cambia?** | §5.1b — sólo si hay PRD |
| 6 | **Gate**: el humano aprueba el brief y **entra al ledger encadenado** | ⭐ **es la diferenciación** |

**El 6 no es negociable.** Sin gate somos otro generador de briefs. Con gate somos lo único
que enforcea Inception y lo encadena con el resto (brief §3.3).

### 1.6 Fuera del MVP, explícito

Flame entero · orquestar el pinponeo (`Q1`) · ingesta multi-fuente tipo DDA · el PRD de
límite automático · plantillas de brief por tipo de proyecto.

### 1.7 Cómo sé que funcionó

| Criterio | Cómo se mide | Por qué es honesto |
|---|---|---|
| **C1** | Toda feature nueva de SpecForge tiene brief aprobado en el ledger | binario, verificable |
| **C2** | Un brief de feature chica se escribe en **≤ 10 min** | si cuesta más, es el overkill de nuevo |
| **C3** | En los primeros **10 briefs**, al menos **1** termina en `kill` o `pivot` | ⭐ **si Spark nunca dice que no, no está filtrando nada** |
| **C4** | Todo `proceed` tiene el porqué anclado a algo observado, no a "sería lindo" | es el 43% mirado de frente |

> **C3 es el criterio que puede fallar**, y por eso es el que vale. Los otros tres se
> cumplen escribiendo documentos. C3 sólo se cumple si el gate **rechaza algo**.
> `sf init` brownfield ya dio `pivot` en el primer intento (brief §6.3): la señal existe.

---

## 2. Historias de usuario

Escritas para **una sola persona con cuatro sombreros** (§1.3): el rol nombra el sombrero,
no a otra persona.

| # | Historia | Criterios de aceptación |
|---|---|---|
| **H1** | Como **dueño del producto**, quiero que al tirar una idea me pidan el **porqué anclado a algo que pasó**, para no meter features porque sí | · el brief no se puede aprobar sin porqué<br>· un porqué que no cite un hecho, un dolor o una medición se marca como débil<br>· *"sería lindo tener"* no pasa |
| **H2** | Como **analista**, quiero saber si **ya lo tengo resuelto** antes de construir, para no duplicar | · se busca en el índice de capacidades del repo<br>· si hay un candidato, se cita `path` y se pide comparar<br>· el resultado entra al brief |
| **H3** | Como **analista**, quiero saber si **ya existe afuera**, para decidir usar, forkear o construir | · búsqueda web con fuentes citadas<br>· el brief dice en qué me diferencio, o dice que no me diferencio |
| **H4** | Como **dueño del producto**, quiero un veredicto **`proceed`/`pivot`/`kill`** explícito, para que "no hacerlo" sea un resultado válido | · el veredicto es obligatorio<br>· `kill` y `pivot` se registran igual que `proceed`<br>· ningún veredicto se emite sin evidencia citada |
| **H5** | Como **dueño del producto**, quiero poder **hacerlo igual** cuando me objetan, con la objeción registrada | · el humano siempre puede seguir<br>· la anulación guarda qué se objetó, qué se decidió y por qué<br>· nunca se puede anular sin dejar rastro |
| **H6** | Como **analista**, quiero anotar lo que **todavía no sé** con la fase que lo contesta, para no trabarme ni inventar | · toda pregunta abierta lleva fase<br>· una pregunta abierta **no bloquea** el brief<br>· una pregunta de "cómo implementar" se difiere a Forge automáticamente |
| **H7** | Como **dueño del producto**, quiero que me avisen si una feature **cambia el alcance** del producto, para no correr el límite sin darme cuenta | · si hay PRD, se compara contra él<br>· *"lo cambia"* se marca como decisión aparte y más grande<br>· si no hay PRD, se dice — no se inventa uno |
| **H8** | Como **responsable**, quiero que el brief aprobado **quede sellado en el ledger**, para que la cadena arranque acá y no en `propose` | · aprobar es un acto explícito del humano<br>· el sello encadena por hash con lo que venga después<br>· un brief editado después de aprobado se detecta |

> **Sobre la forma.** El *"como X quiero Y para Z"* es ceremonia de equipo y acá el usuario
> es uno solo. Se mantiene por un motivo puntual: **obliga a nombrar a quién le sirve**, que
> es justo la pregunta que no se hizo en el 43%. Si en el uso resulta ruido, se corta — pero
> se corta después de probarlo, no antes.

---

## 3. Constitución mínima

**Sólo lo declarado** (brief `Q2`). Derivado de `.github/workflows/lint.yml`, 577 bytes:

```yaml
lenguaje:        Go
directorio:      cli/
test:            go test ./...
build:           go build -o sf .
vet:             go vet ./...
lint del suite:  ./cli/sf lint
```

Más lo que se sabe sin inferir nada:

```yaml
skills:          Markdown en skills/<nombre>/SKILL.md
reporte por test: build.report = go-json | junit   # ya existe en cli/constitution.go
```

**Y nada más.** Convenciones, arquitectura y límites de capas **no se escriben acá**: se
descubren cuando se va a escribir código y se agregan entonces. Ya hay evidencia propia de
que derivarlas automáticamente falla (`sf arch` congelado por `RM-C7b`).

---

## 4. Preguntas abiertas nuevas

| # | Pregunta | La contesta |
|---|---|---|
| **Q9** | El "índice de capacidades" (brief §6.3) funcionó acá porque las skills tienen títulos descriptivos. ¿Qué pasa en un repo sin ese índice? ¿Se construye uno, o Spark degrada y lo dice? | **Spark**, en un repo ajeno |
| **Q10** | El PRD de límite (§0) lo escribe el humano al adoptar. ¿Es un paso de `sf init`, o el primer Spark lo pide cuando lo necesita? | **Spark** |
| **Q11** | `C3` dice que si Spark nunca rechaza, no filtra. ¿Qué se hace si a los 10 briefs sigue en cero — se afloja el criterio o se endurece el gate? | **el uso**, a los 10 briefs |

---

## 5. Nota sobre el formato de Flame

- **El primer problema no fue de contenido sino de modo** (§0): Flame no podía arrancar
  porque el modo era indeterminable. **Un Flame DEBE empezar declarando su modo**, y si el
  modo no se puede determinar, eso es el primer hallazgo — no un detalle a resolver de paso.
- **§1.2 (qué NO es) es la sección que hace trabajar al PRD.** Sin ella, §5.1b del brief no
  tiene contra qué comparar y el detector de alcance queda decorativo. Un PRD sin
  anti-alcance no sirve para lo único que se le pide.
- **§1.5 (alcance del MVP) fue donde más se recortó.** Se entró a Flame para escribir
  Spark + Flame y se sale con **Spark solo**. El recorte lo forzó una pregunta simple:
  *¿qué de esto ya está validado?* Flame se está caminando por primera vez ahora mismo —
  meterlo en el MVP sería construir sobre algo que no se probó.
- **Los criterios de éxito se dividen solos en dos clases**: los que se cumplen escribiendo
  documentos (C1, C2, C4) y **el que sólo se cumple si el sistema rechaza algo (C3)**. Un
  Flame **DEBERÍA** exigir al menos un criterio de la segunda clase. Si todos los criterios
  se pueden satisfacer produciendo artefactos, el módulo no tiene forma de fallar — y algo
  que no puede fallar tampoco puede estar funcionando.
