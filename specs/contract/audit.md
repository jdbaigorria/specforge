# Contrato: `sf audit` — el brazo

**Estado:** ⏸ CONGELADO — hipótesis sin consumidor validado (ver [`README.md`](README.md)) · **Fecha:** 2026-08-07 · **Contrato de salida:** [`judge.md`](judge.md)

---

## 0. La lente correcta

> **El CLI nació para producir artefactos que hagan que el LLM alucine menos.**

### 0.1 Dos nombres parecidos que son cosas distintas

Este contrato usa los dos con precisión y **no son intercambiables**:

| | Qué es | Rol |
|---|---|---|
| **`sf-audit`** (con guión) | un **skill**: un subagente de contexto limpio | **el cerebro.** Hace la revisión punta a punta y firma el informe |
| **`sf audit`** (comando) | un comando del CLI de Go | **el brazo.** Ejecuta, computa, sella y entrega material. Nunca opina |

**Este documento especifica el brazo.** El cerebro está en [`judge.md`](judge.md), que es
el contrato de salida de `sf-audit`.

### 0.2 El CLI es un brazo, no una capa que compite

`sf audit` **no es el producto**. El producto es el **informe** que firma `sf-audit`. El
brazo existe para dos cosas: **darle al cerebro material que no puede inventar**, y
**ahorrarle computar lo que se computa** — para que piense en vez de gastar atención en
orquestar pasos.

Corolario de diseño, y es normativo:

> **Cada secuencia de comandos que un skill tiene que recordar es un lugar donde el skill
> se puede equivocar.** La orquestación es del brazo; el pensamiento es del cerebro.

Por eso `sf audit` es **un solo comando que entrega todo** (§5.2), y no un conjunto de
verbos que el cerebro tiene que correr en orden.

### 0.3 La regla que decide todo: dos canales, dos procedencias

El informe de `sf-audit` mezcla dos clases de afirmación, y **cada línea DEBE ser
atribuible a una de las dos**:

| Canal | Quién lo emite | Ejemplo | Garantía |
|---|---|---|---|
| **Hecho** | el brazo — textual, sellado | `71% de mutantes muertos (5/7)` | **no fabricable**: salió de ejecutar algo |
| **Juicio** | el cerebro — prosa con cita | *"el regex puede perder dígitos y ningún test lo nota"* | **falible pero verificable**: la cita se resuelve |

> **Un número que el cerebro parafrasea deja de ser un hecho.** El cerebro **PUEDE**
> interpretar lo que el brazo devolvió; **NO DEBE** reescribirlo de memoria. Lo que va al
> informe como hecho es lo que reportó el brazo, literal.

De ahí sale el criterio que asigna cualquier responsabilidad futura, sin tener que
discutirla:

> **Si validar X exige recomputar X, X es del brazo. Si validar X es resolver una cita,
> X es del cerebro.**

Probado contra el contrato: para validar `CRAP = 2` hay que recomputar CRAP → brazo. Para
validar *"`sep` no lo reclama ningún `R#`"* alcanza un lookup en `trace.json` → cerebro
(es `judge.md` §5 condición 10).

### 0.4 Determinismo para lo fabricable, juicio para todo lo demás

> **Determinismo para lo que el modelo podría fabricar. Juicio para todo lo demás.**

Y "todo lo demás" es ancho: el cerebro revisa la feature **entera**, punta a punta
(`judge.md` §0). Lo que ningún juicio puede sustituir es **haber ejecutado algo** — un LLM
diciendo *"corrí los tests y pasaron"* es el incidente que originó SpecForge, y vale
idéntico para *"corrí la mutación y dio 85%"*. **Ejecutar es siempre del brazo.**

---

## 1. Propósito

El sistema entero —cerebro y brazo— existe para responder **una** pregunta:

> ¿Puedo aceptar este trabajo **sin leer el código**?

**La responde `sf-audit` en su informe.** `sf audit` existe para que esa respuesta **no
esté inventada**. No es un reporte de salud. No es QA. Es **la licencia para no mirar** —
y por lo tanto el informe no puede ser algo que haya que leer y evaluar para saber si
sirve.

### 1.1 El criterio de éxito, explícito

El sistema cumple su función si y sólo si el humano puede tomar una decisión **mirando
sólo el informe**. Un informe que obliga a abrir el diff para saber si es confiable es un
informe fallido, aunque su contenido sea correcto.

Corolario: **el informe no dice "está bien". Dice dónde hay que mirar.** El producto no
es la ausencia de trabajo humano, es su **concentración**: pasar de revisar 40 archivos a
revisar un requisito.

### 1.2 Las dos garantías, y cuál es alcanzable

| Garantía | Alcanzable | Quién la da |
|---|---|---|
| **"no puede mentir"** — nada acá está fabricado | **Sí** | los hechos del brazo (§3) + la verificación de las citas del cerebro (`judge.md` §5) |
| **"no puede equivocarse"** — esto es correcto | **No** | ninguna herramienta la tiene, ni un humano leyendo el diff |

El informe **DEBE** garantizar la primera y **NO DEBE** insinuar la segunda en ninguna
parte de su salida.

### 1.3 Qué produce el brazo cuando no hay cerebro

`sf audit` **DEBE** poder correr sin `sf-audit` — es lo que hace que el CI funcione sin
LLM. Pero lo que produce en ese caso es **hechos y un código de salida, no un informe**:

| Corre | Produce |
|---|---|
| `sf audit` solo | los hechos de §3, la clasificación de §4, y el exit code de §5.3 |
| `sf-audit` (que invoca `sf audit`) | lo anterior **más el informe** — la revisión punta a punta |

Esto es una consecuencia aceptada del diseño, no una limitación a disimular: **el informe
requiere cerebro.** Un CI sin LLM detecta fabricación y regresión; no detecta *"esta
función no la pidió nadie"*.

---

## 2. Cerebro y brazo

El bucle es **pedir → trabajar → entregar → validar → sellar**, y el cerebro toca el CLI
en exactamente **dos** momentos:

```
   ┌─ sf-audit ────────────────────────────────── subagente de CONTEXTO LIMPIO ─┐
   │                                                                            │
   │   ① PEDIR                                                                  │
   │   sf audit --feature=x --json ──────────────► ┌──────────────────────────┐ │
   │                                               │  EL BRAZO                │ │
   │                                               │  ejecuta el suite        │ │
   │                                               │  resuelve símbolos       │ │
   │                                               │  corre mutación / CRAP   │ │
   │                                               │  clasifica cada R# (§4)  │ │
   │                                               │  sella contra tree_hash  │ │
   │   ◄─────────────── material completo (§5.2) ──┤  NUNCA invoca un LLM     │ │
   │                                               └──────────────────────────┘ │
   │   ② TRABAJAR                                                               │
   │   revisión PUNTA A PUNTA de la feature entera (judge.md §3.1)              │
   │   el `residue` ORIENTA por dónde empezar · NO limita qué puede mirar       │
   │                                                                            │
   │   ③ ENTREGAR                                                               │
   │   informe: HECHOS del brazo, textuales  +  JUICIOS propios, con cita       │
   │                                                                            │
   └────────────────────────────┬───────────────────────────────────────────────┘
                                │
   ④ VALIDAR + ⑤ SELLAR         ▼
   sf gate record-verdict ──► resuelve cada cita · fail-closed (judge.md §5)
                              rechaza entero si algo no resuelve · al ledger
```

**Reglas:**

1. El brazo **NO DEBE** invocar ningún LLM. Ni para computar, ni para redactar, ni para
   resumir. Si un LLM tocó un hecho, no es un hecho.
2. El cerebro **NO DEBE** recomputar un hecho ni reescribirlo. Recibe la salida `--json`
   como entrada de sólo lectura y la cita **textual**. **PUEDE** señalar un hallazgo sobre
   un requisito que el brazo dejó en `PROBADO` — con cita verificable — y ese hallazgo se
   registra **al lado** del hecho, nunca encima. El humano ve los dos.
3. El cerebro **NO DEBE** ejecutar nada cuyo resultado vaya al informe como hecho. Correr
   el suite, la mutación o el análisis de complejidad por fuera del brazo **NO DEBE**
   ocurrir: es el vector de fabricación que originó SpecForge (§0.4).
4. `sf audit` **DEBE** poder correr sin cerebro, produciendo hechos y exit code — **no** un
   informe (§1.3).

### 2.1 Por qué un solo comando y no siete

`sf audit` **DEBE** entregar el material **completo en una sola invocación**. Dos motivos
independientes, y cada uno alcanza:

1. **El cerebro no elige su propio examen** (`judge.md` §3.1 regla 1). Si el subagente
   decide qué comandos correr, empieza a decidir qué mira — y un juez que arma su propio
   material puede armarse uno cómodo.
2. **Una secuencia de pasos en prosa no es una garantía.** Depende de que el modelo la lea
   entera, la recuerde y no se saltee el tercero. Encapsulada en un comando, es código
   testeable. Es el corolario de §0.2.

**Regla general, aplicable a todo skill de SpecForge:**

> **La superficie de comandos que un skill invoca DEBE ser igual a sus puntos de decisión.**

`sf-audit` tiene dos puntos de decisión —*"dame todo sobre esta feature"* y *"acá está mi
juicio, validalo"*— y por lo tanto dos comandos. Ni uno (un comando que además sellara su
propio veredicto convertiría al brazo en juez de sí mismo), ni siete.

> **Corolario de poda.** Un comando del CLI que ningún skill invoca en un punto de decisión
> está muerto **o le falta estar adentro de otro comando**. No hay una tercera opción.

---

## 3. Los hechos

Cada hecho **DEBE** cumplir tres propiedades:

- **Determinista** — dos corridas sobre el mismo árbol dan el mismo resultado.
- **Sellado** — se registra junto al `tree_hash` sobre el que se computó.
- **Invalidable** — existe una condición explícita que lo vuelve falso.

### 3.0 El brazo ejecuta, nunca reimplementa

> **El brazo NO DEBE reimplementar ningún análisis que ya exista como herramienta madura.**
> Para cada hecho que requiere analizar código —resolver símbolos, medir complejidad,
> generar mutantes, correr tests— el brazo **DEBE** hacer shell-out a la mejor herramienta
> disponible y quedarse con lo que sí es suyo: **invocar, normalizar y sellar**.

Escribir complejidad ciclomática en Go es exactamente el error que este contrato quiere
evitar: código propio, peor probado, para un problema resuelto hace veinte años.

**Lo que sí es del brazo** —y por lo tanto se implementa— es la parte que ninguna
herramienta externa hace:

| El brazo hace | Por qué no lo hace la herramienta |
|---|---|
| **Acotar** el análisis a los archivos anclados a un `R#` | `mutmut` no sabe qué es un `R#` |
| **Normalizar** la salida al schema de §5.2 | cada herramienta tiene su formato |
| **Sellar** contra el `tree_hash` | ninguna herramienta ata su resultado a un árbol |
| **Clasificar** y aplicar el trinquete (§4) | es la tabla de verdad del contrato |

#### Contrato del adaptador

Cada lenguaje soportado **DEBE** declarar sus herramientas en configuración, no en código:

```yaml
# specforge/toolchain.yaml
python:
  symbols:    { cmd: "sg", args: ["scan", "--json"] }
  complexity: { cmd: "lizard", args: ["--csv"] }
  coverage:   { cmd: "coverage", args: ["json"] }
  mutation:   { cmd: "mutmut", args: ["run", "--paths-to-mutate"] }
  test:       { cmd: "pytest", args: ["-q"] }
```

**Reglas del adaptador:**

1. El adaptador **DEBE** limitarse a *invocar → parsear → normalizar*. **NO DEBE** contener
   lógica de decisión: clasificar es de §4.
2. **DEBERÍA** preferir el modo de salida estructurada de la herramienta (`--json`, `--csv`)
   antes que parsear texto para humanos. Un formato para humanos cambia entre versiones
   menores sin avisar.
3. Si la herramienta **no está instalada o no está declarada**, el hecho es
   **`NO_EVALUADO`** (§4), nunca verde y nunca `ROTO`. Una herramienta ausente **NO DEBE**
   poder parecerse a un análisis limpio.
4. Si la herramienta corre pero su salida no parsea, el hecho es **`NO_EVALUADO`** y el
   error crudo **DEBE** quedar en el ledger. Un parser roto **DEBE** fallar ruidosamente.
5. Agregar un lenguaje **DEBERÍA** ser agregar un bloque a `toolchain.yaml` más su parser.
   Si exige tocar la lógica del audit, el adaptador está mal cortado.

> **El costo declarado.** Este diseño paga un parser por herramienta, y es la desventaja
> real frente a dejar que el cerebro lea cualquier formato. Se acepta a propósito: un
> parser roto falla ruidosamente y se testea, mientras que **un cerebro que malinterpreta
> una salida no falla nunca** — reporta un número plausible y equivocado, que es
> exactamente la clase de mentira que este contrato existe para impedir (§0.4). La regla
> de §0.3 no admite excepción por comodidad.

**Preferencia de herramienta**, en orden: (1) una herramienta **multi-lenguaje** que cubra
varios de una —menos adaptadores que mantener—; (2) la herramienta de referencia del
lenguaje. Entre equivalentes, **DEBERÍA** preferirse la que tenga salida JSON estable.

### F1 — ANCLAJE

Cada `R#` de la feature resuelve a al menos un `path:symbol` que **existe** en el árbol.

- **Fuente:** `trace.json`
- **Cómo se computa:** resolución del símbolo en el archivo por **AST, no grep** — shell-out
  al motor de símbolos del `toolchain` (§3.0). Un grep da falsos positivos en comentarios,
  strings y nombres parciales, y F1 es la base de todo lo demás.
- **Invalidación:** el archivo no existe, o el símbolo no está en él → `ROTO`.
- **Lo que NO dice:** que ese símbolo *implemente* el requisito.

### F2 — PRUEBA EXISTE

Cada **criterio de aceptación** (no cada requisito) tiene un test nombrado que existe.

- **Fuente:** `trace.json` + el contrato de verificación.
- **Granularidad:** a nivel criterio. Un requisito con cinco criterios necesita cinco
  anclas de test. Cubrir uno **NO DEBE** dar el requisito por cubierto.
- **Invalidación:** el test nombrado no existe en el árbol → `ROTO`.
- **Lo que NO dice:** que el test corra, ni que pase, ni que sea sobre eso.

### F3 — VERDE SELLADO

El suite corrió **adentro del CLI**, dio verde, y el resultado quedó atado al árbol.

- **Cómo se computa:** `sf check run` ejecuta el comando de test del proyecto, captura
  código de salida y salida, y sella.
- **El sello DEBE incluir:** `tree_hash` de **los archivos anclados en `trace.json` más
  los archivos de test anclados**, el comando exacto, el timestamp, y el código de salida.
- **Invalidación:** cualquier byte de esos archivos cambió → sello **stale** (ver F4).
- **Lo que NO dice:** que los tests prueben algo. Un suite de `assert true` sella verde.

> **Corregido 2026-08-07 tras validar contra `slugify` — ver [`validation/slugify.md`](validation/slugify.md) H5.**
> Antes decía que el binding al contenido era el único cambio de mecánica nuevo. **Es falso:
> ya existe.** `cli/check.go` guarda `code_hash` (*"'corrí una vez y después cambié todo'
> queda detectado"*) y además `tests map[string]string` con el resultado **por test**, lo
> que permite exigir que cada test nombrado en el trace haya corrido y pasado **en esa
> corrida** — causalidad test→requisito, no sólo "la suite dio verde". F3 y F4 son
> documentación de lo construido más el ruteo de su resultado al veredicto, no trabajo nuevo.

### F4 — FRESCURA

`tree_hash` actual de los archivos sellados == `tree_hash` del sello.

- **Invalidación:** difieren → el sello es **stale**, y todos los hechos que dependen de
  él pasan a `ROTO`. **NO DEBE** degradarse a advertencia.
- `stale` y `missing` **NO DEBEN** colapsarse nunca: son causas distintas y arreglos
  distintos.

### F5 — DRIFT Y RIESGO DE CAMBIO

- **Anclas huérfanas:** símbolos anclados que ya no existen, o código de producción sin
  ningún `R#` que lo reclame.
- **Riesgo de cambio:** por componente, la métrica **CRAP** — combinación de complejidad
  ciclomática y cobertura de tests.
- **Cómo se computa:** dos shell-outs y una multiplicación (§3.0). La **complejidad** sale
  del analizador del `toolchain`; la **cobertura por función** sale de la herramienta de
  cobertura nativa del lenguaje. El brazo las cruza y aplica la fórmula de CRAP. **NO DEBE**
  implementar ninguna de las dos mediciones.

> **Por qué CRAP y no % de líneas cubiertas.** CRAP tiene **dos salidas**: se baja
> testeando más *o* simplificando el código. Un % de cobertura tiene una sola, y empuja a
> escribir tests para código que habría que partir. Referencia: complejidad ciclomática
> > 6 se considera alta.

- **Trinquete:** el riesgo agregado **NO DEBE** subir entre corridas sin una anulación
  explícita registrada en el ledger.
- **Invalidación:** ancla huérfana → `ROTO`. Trinquete violado → el veredicto global no
  puede ser limpio.

### F6 — FUERZA DEL TEST (mutación)

Por cada `R#`: se generan mutantes sobre el código anclado y se mide cuántos mata el suite.

- **Cuándo corre:** sólo en perfil de rigor `strict` (§7). Es caro en tiempo real.
- **Cómo se computa:** shell-out a la herramienta de mutación del `toolchain` (§3.0),
  **acotada por el brazo a los archivos anclados a ese `R#`** — no al repo entero. Ese
  acotado es la parte que ninguna herramienta de mutación puede hacer sola: no sabe qué es
  un `R#`, y es lo que vuelve el resultado **atribuible a un requisito**.
- **Quién la corre:** el brazo, **nunca el cerebro**. Un LLM diciendo *"la mutación dio
  85%"* es la misma fabricación que *"corrí los tests y pasaron"* (§0.4).
- **Umbral:** configurable, por defecto **80% de mutantes muertos**.
- **Invalidación:** por debajo del umbral → el requisito pasa a `DÉBIL`.
- **Por qué existe:** es el **único proxy mecánico** de *"¿el test prueba el requisito?"*.
  Un test que sobrevive a la mutación no está testeando; está pasando. Sin F6, F2 y F3 son
  verificables y vacíos.

### F7 — COBERTURA DE CRITERIOS

Porcentaje de criterios de aceptación de la feature con F1..F3 en verde.

- Es la métrica de completitud. **NO DEBE** mezclarse con la cobertura de líneas: unidades
  distintas, y sumarlas es sumar peras con manzanas.

---

## 4. Clasificación: los seis estados de un requisito

`sf audit` **DEBE** asignar a cada `R#` exactamente uno de estos estados. **La
clasificación es del brazo**, no del cerebro: es una tabla de verdad sobre §3, y validarla
exigiría recomputarla (§0.3).

La última columna dice **qué significa el estado para el cerebro** — no si puede mirarlo.
El cerebro mira todo (§4 regla 2).

| Estado | Condición | Qué significa para el cerebro |
|---|---|---|
| **`PROBADO`** | F1–F5 verdes, y F6 sobre el umbral si corrió | las verificaciones mecánicas no encontraron nada. **PUEDE** señalar un caso no cubierto con cita; **NO PUEDE** cambiarle el estado |
| **`DÉBIL`** | F1–F5 verdes, F6 por debajo del umbral | **acá hay algo**: empezar por aquí |
| **`SIN_PRUEBA_MECANICA`** | el requisito declara verificación no automatizable (NFR, requiere medición o inspección externa) | ningún hecho lo alcanza; requiere juicio + evidencia externa |
| **`ROTO`** | falla cualquiera de F1–F5 **habiendo podido computarse** | no hay nada que juzgar: hay que arreglar |
| **`NO_EVALUADO`** | un hecho requerido **no se pudo computar** (sello ausente, herramienta no configurada) | no hay nada que juzgar: falta correr un paso |
| **`NO_APLICA`** | el requisito quedó fuera de alcance en esta feature, con motivo registrado | fuera de alcance |

> **`NO_EVALUADO` agregado 2026-08-07 tras validar contra `slugify` — ver
> [`validation/slugify.md`](validation/slugify.md) H2.** El contrato original colapsaba
> *"el sello existe y no coincide"* con *"el sello no existe"*, que piden acciones opuestas:
> arreglar código vs. correr un paso. Es el mismo error que F4 prohíbe explícitamente
> (`stale` ≠ `missing`), cometido un nivel más arriba.

**Reglas de clasificación:**

1. `ROTO` **DEBE** tener precedencia sobre `DÉBIL`, `SIN_PRUEBA_MECANICA` y `PROBADO`. Un
   requisito con sello **stale** es `ROTO`, no `DÉBIL`.
   `NO_EVALUADO` **DEBE** tener precedencia sobre `ROTO`: si un hecho no se pudo computar,
   el veredicto **NO DEBE** afirmar que algo falló.
2. El **residuo** = `DÉBIL` ∪ `SIN_PRUEBA_MECANICA`. **Orienta, no limita.** El cerebro
   **DEBERÍA** empezar por ahí, porque es donde los hechos ya avisaron que hay algo, y
   **PUEDE** pronunciarse sobre cualquier parte de la feature — incluido un `R#` en
   `PROBADO` — siempre con cita verificable.
3. El material que se le entrega al cerebro **DEBE** contener la feature **entera**,
   incluidos los `R#` en `PROBADO` (`judge.md` §3.1). Recortar el material **NO DEBE**
   usarse como mecanismo de control: el control es la verificación de la salida
   (`judge.md` §5), no la ceguera de la entrada.

> **Reglas 2 y 3 corregidas 2026-08-07.** Decían lo contrario: que el residuo era *"lo
> único que el juez puede tocar"* y que un `PROBADO` **NO DEBÍA** aparecer en el material,
> *"no como optimización de tokens sino como restricción de autoridad"*.
>
> Era el mismo error que `judge.md` §0 ya había corregido, sobreviviendo acá — y con él,
> el punto ciego entero. En `slugify`, el parámetro `sep` sin requisito apareció **por
> accidente**: un mutante *no relacionado* (dígitos) dejó a `R1` en el residuo. Con `R1`
> limpio, el cerebro nunca veía la función (`validation/slugify.md` §4-H4).
>
> **Restringir la entrada de una capa cooperativa no agrega seguridad — sólo le saca
> visión.** La seguridad la da verificar la salida. La pregunta correcta al diseñar es
> *"¿puedo verificar lo que afirma?"*, nunca *"¿puedo limitar lo que ve?"*.

---

## 5. Salida

### 5.1 Humana (default)

Es **la salida del brazo**, no el informe. Sirve para el humano que corre el comando a
mano y para el CI; el informe lo firma `sf-audit` (§1.3).

```
sf audit --feature=auth-login

  ✓ NO FABRICADO     12/12 R# → código + test
                     suite sellada · tree 3f2a1b · go test ./... · exit 0
  ✓ NO STALE         0 sellos rotos
  ✓ NO DRIFT         0 anclas huérfanas · riesgo de cambio 84 → 79 (trinquete ok)
  ✓ CRITERIOS        27/27 criterios de aceptación con prueba

  ⚠ TEST DÉBIL       R7 · sobreviven 3 de 8 mutantes (62% < 80%)
                     → el test pasa pero no prueba
  ? NO MECÁNICO      R4 · NFR "p95 < 200ms" · requiere medición externa
                     → evidencia: pendiente

  HECHOS      10 de 12 con prueba mecánica.
              Mirá R7. R4 necesita tu medición.
              Perfil de rigor: strict
  SIN INFORME no corrió `sf-audit` sobre este árbol.
              Nadie revisó si hay código que ningún requisito pidió.
```

**Reglas de la salida humana:**

1. **NO DEBE** contener las palabras "todo bien", "OK", "aprobado" ni equivalentes.
2. **NO DEBE** titularse "VEREDICTO". El brazo reporta hechos; el veredicto es del informe.
3. **DEBE** nombrar explícitamente qué requisitos requieren atención humana, y **sólo** esos.
4. **DEBE** decir el perfil de rigor con el que corrió — un `strict` y un `lean` limpios
   no significan lo mismo, y ocultar cuál corrió es una forma de mentir.
5. **DEBE** decir si hay informe de `sf-audit` registrado para ese `tree_hash` o no, y
   cuando no lo hay **DEBE** nombrar qué clase de hallazgo queda sin cubrir. Unos hechos
   limpios sin informe **NO DEBEN** poder confundirse con una revisión.

### 5.2 `--json`

Es la **entrada del cerebro** y del CI. Forma:

```json
{
  "feature": "auth-login",
  "rigor": "strict",
  "tree_hash": "3f2a1b...",
  "generated_at": "2026-08-07T14:22:03Z",
  "facts": {
    "F1_anchoring":  {"ok": true,  "total": 12, "failed": []},
    "F2_test_exists":{"ok": true,  "total": 27, "failed": []},
    "F3_sealed_green":{"ok": true, "command": "go test ./...", "exit": 0,
                       "sealed_tree": "3f2a1b..."},
    "F4_freshness":  {"ok": true,  "stale": [], "missing": []},
    "F5_drift":      {"ok": true,  "orphans": [], "crap": {"before": 84, "after": 79}},
    "F6_mutation":   {"ok": false, "ran": true, "threshold": 80,
                      "per_requirement": {"R7": {"killed": 5, "total": 8, "score": 62}}},
    "F7_criteria":   {"ok": true,  "covered": 27, "total": 27}
  },
  "requirements": [
    {"id": "R7", "state": "DEBIL",
     "reason": "F6 62% < 80%",
     "anchors": [{"path": "src/auth/login.go", "symbol": "Login"}],
     "tests":   [{"path": "src/auth/login_test.go", "symbol": "TestLoginRejectsExpired"}],
     "judge_verdict": null},
    {"id": "R4", "state": "SIN_PRUEBA_MECANICA",
     "reason": "verification: external",
     "judge_verdict": null, "evidence": null}
  ],
  "residue": ["R7", "R4"],
  "material": {
    "requirements": "...todos, con todos sus criterios de aceptación...",
    "code":         [{"path": "src/auth/login.go", "content": "..."}],
    "tests":        [{"path": "src/auth/login_test.go", "content": "..."}],
    "diff":         "...",
    "constitution": ["...principios cuyo applies_to incluye esta fase..."],
    "rubrics":      ["implementation-review"]
  },
  "exit": 3
}
```

**`material` es lo que hace que esto sea un solo comando.** Contiene la feature entera
según `judge.md` §3.1 regla 2 — todos los requisitos con todos sus criterios, todo el
código y los tests anclados **completos**, el diff, la constitución aplicable y las
rúbricas. `sf audit --json` **absorbe a `sf context for-judge`**: el cerebro hace una
llamada y tiene todo (§2.1).

El brazo **DEBE** decidir el contenido de `material`. El cerebro **NO DEBE** leer archivos
por su cuenta ni elegir qué mirar: el juez no elige su propio examen.

**`residue` es orientativo, no normativo.** Le dice al cerebro **por dónde empezar** —es
donde los hechos ya avisaron que hay algo— y **no** define su jurisdicción. El cerebro
**PUEDE** pronunciarse sobre cualquier `R#` del `trace.json`, incluido uno en `PROBADO`,
siempre con cita verificable. `record-verdict` **NO DEBE** rechazar un veredicto por estar
fuera del `residue`; rechaza por cita irresoluble (`judge.md` §5).

> **Corregido 2026-08-07.** Este párrafo decía *"`residue` es normativo… define la
> jurisdicción del juez… `record-verdict` **DEBE** rechazar"*, y citaba a `judge.md` §5
> para una condición de rechazo **que `judge.md` §5 nunca tuvo**. Es la misma corrección
> que §4 reglas 2–3.

### 5.3 Códigos de salida

| Código | Significado |
|---:|---|
| `0` | Todo `PROBADO`, o el residuo tiene un veredicto `no_refuta` registrado y válido |
| `3` | Hay residuo **sin** informe de `sf-audit` registrado para este árbol |
| `4` | El informe **refutó** al menos un requisito |
| `5` | Hay al menos un `ROTO` |
| `6` | El propio audit no pudo correr (config faltante, comando de test no definido) |
| `7` | Hay al menos un `NO_EVALUADO` — falta correr un paso, no hay nada fallado |

**Regla fail-closed:** ante cualquier ambigüedad, el código de salida **DEBE** ser el más
alto que aplique. `0` es la única afirmación positiva y sólo se emite cuando nada la
contradice.

---

## 6. Alcance: proyecto vs. feature

| Invocación | Alcance | Uso |
|---|---|---|
| `sf audit --feature=<x>` | una feature | gate de `check`, CI por PR |
| `sf audit` | todas las features vivas + consistencia cruzada | auditoría de proyecto |

En alcance de proyecto se agregan hechos que sólo existen entre features —
inconsistencia de modelo de datos, conflictos de interfaz, terminología divergente,
solapamiento de alcance. Esos hechos siguen las mismas tres propiedades de §3: si no se
pueden computar de forma determinista, **no son hechos** y van al residuo.

---

## 7. Perfiles de rigor

El rigor decide **qué hechos computa el brazo y qué cerebros corren**. Es **un solo eje**,
persistente entre sesiones (`sf rigor <perfil>`).

| Perfil | Hechos | Cerebro | Cuándo |
|---|---|---|---|
| `lean` | F1–F5 | ninguno | spikes, config, docs |
| `standard` *(default)* | F1–F5, F7 | `sf-audit` (implementación) | la mayoría del trabajo |
| `strict` | F1–F7 | `sf-audit` + revisión de entrega (Ship) | core de producción |

> **El perfil de rigor es lo que hace seguro el un-solo-comando de §2.1.** F6 es caro en
> tiempo real; sin un eje que lo apague, un comando que bundlea todo bloquearía minutos en
> cada corrida. En `standard` no corre mutación y `sf audit` es rápido; en `strict` es
> lento **a propósito**. Sin §7, §2.1 sería un comando-dios sin freno.

**Regla:** una garantía de calidad **DEBE** ser una etapa del pipeline con un dueño, y
**NO DEBE** ser un campo del schema del requisito.

> **Por qué esto reemplaza a `require_mutation` / `require_arch` / `require_red_witness` /
> `require_source` / `blocking_priorities`.** Los cinco flags gradúan garantías desde el
> schema del requisito, lo que obliga a que cada requisito cargue con la política del
> proyecto. El rigor la pone donde va: en el proyecto. Los cinco flags se colapsan en un
> campo.

---

## 8. Lo que este contrato NO garantiza

Declarado acá para que no haya que deducirlo:

1. **Que el requisito fuera el correcto.** Un `R#` puede estar `PROBADO` de punta a punta
   y no ser lo que el usuario necesitaba. Eso no se arregla en audit: se arregla en el
   módulo **Inception**, aprobando el requisito. `sf audit` **NO DEBE** insinuar que
   cubre esto.
2. **Que el cerebro no se equivoque.** Es cooperativo por diseño (`judge.md` §1). Lo que
   sí se garantiza es que **no puede fabricar evidencia**.
3. **Que la garantía sobreviva fuera del harness.** Los hechos F1–F7 son del CLI y valen
   en cualquier lado. La *imposibilidad de escribir código fuera de orden* depende de los
   hooks, y los hooks dependen del harness. Fuera de un harness soportado, `sf audit`
   sigue diciendo la verdad sobre lo que encuentra, pero deja de haber quien impida el
   desorden. **Esto DEBE estar en el README del producto, no escondido en un doc interno.**

---

## 9. Mapa de reúso

### 9.1 Qué de lo construido sirve

| Hecho / pieza | Hoy | Acción |
|---|---|---|
| F1 anclaje | `sf trace verify` | **reusar el mecanismo, cambiar el motor** — la resolución de símbolos pasa a shell-out AST (§3.0) |
| F2 prueba existe | `sf trace verify --contract` + `acceptance` con id (`RM-C1`) | **reusar.** `RM-C1` es correcto y necesario: sin id por criterio, F2 no se puede computar |
| F3 verde sellado | `sf check run` | **reusar** — `code_hash` y `tests` por test ya existen (H5). Falta rutear su resultado al veredicto |
| F4 frescura | ya existe (`stale`/`missing` separados) | **reusar** |
| F5 drift | `sf doctor --drift` + `sf coverage` + trinquete | **reusar el mecanismo, cambiar la unidad y el motor** — CRAP en vez de % de líneas, vía shell-out |
| F6 mutación | `sf mutation` (`RM-C6`) | **reusar, mover de lugar** — de flag `require_mutation` a etapa del perfil `strict` |
| F7 criterios | `sf coverage` | **reusar** |
| Material del cerebro | `sf context for-judge` (en `sf-check`) | **absorber en `sf audit --json`** (§2.1) |
| Conformidad de arquitectura | `sf arch` (`RM-C7`) | **congelado.** No entra a F1–F7: `RM-C7b` probó que no funciona en `examples/brownfield-tempconv` (dos componentes en un archivo es normal en código chico). Vuelve como etapa de `strict` cuando el mapeo componente→archivo funcione en los tres ejemplos |
| Evidencia externa | `sf evidence` (`DL-15`) | **reusar** — es el canal de `SIN_PRUEBA_MECANICA`. La auditoría lo marcó sin consumidor; ahora lo tiene |
| Fuentes del requisito | `sf sources` (`RM-C3`) | **mover a Inception.** No fracasó por diseño: le faltaba el módulo que produce el material |
| Prioridad | `priority` + `blocking_priorities` (`RM-C2`) | **colapsar en el perfil de rigor** |
| Grafo, métricas, eventos | `sf graph`, `sf metrics`, `sf events` | **fuera.** Sin consumidor antes y sin consumidor en este contrato — y bajo el corolario de poda de §2.1, ningún skill los invoca en un punto de decisión |

### 9.2 El `toolchain` de arranque

**Esto es configuración, no contrato.** Las herramientas se cambian editando
`toolchain.yaml`; que una quede obsoleta **NO DEBE** requerir tocar este documento. Va acá
sólo para que la primera implementación no tenga que investigar de cero.

| Hecho | Multi-lenguaje *(preferido)* | Go | Python | JS/TS | Rust |
|---|---|---|---|---|---|
| **F1** símbolos (AST) | `ast-grep` | — | — | — | — |
| **F5** complejidad | `rust-code-analysis` · `lizard` | `gocyclo` | `radon` | — | — |
| **F5** cobertura por función | — | `go test -coverprofile` | `coverage.py` | `c8` | `cargo-llvm-cov` |
| **F6** mutación | — | `gremlins` · `go-mutesting` | `mutmut` · `cosmic-ray` | `Stryker` | `cargo-mutants` |
| **F3** tests | el comando del proyecto | — | — | — | — |

**Notas de selección:**

- **F1 con `ast-grep`** cubre Go, Python, JS/TS, Rust, Java y varios más **con un solo
  binario** — o sea, un adaptador en vez de seis. Es el caso más claro de la preferencia
  multi-lenguaje de §3.0.
- **F5 no tiene una sola herramienta que dé las dos mitades.** La complejidad sí es
  multi-lenguaje; la **cobertura no lo es** y siempre va a ser nativa del lenguaje. Es el
  adaptador más caro de los tres, y hay que contarlo como tal.
- **F6 es irreduciblemente per-lenguaje.** No existe un mutador multi-lenguaje serio: mutar
  exige entender la semántica, no sólo la sintaxis. Cada lenguaje nuevo cuesta un adaptador
  entero — razón de más para que F6 viva sólo en `strict` (§7).

> **Primer lenguaje: Python.** Es el de `examples/slugify`, que es donde el contrato ya se
> validó a mano (`validation/slugify.md`) y donde está el único resultado de F6 medido.
> Segundo: **Go**, porque es SpecForge mismo y es lo que habilita usar SpecForge sobre
> SpecForge. **NO DEBERÍA** declararse soporte de un lenguaje sin un ejemplo que lo ejercite
> punta a punta.

---

## 10. Preguntas abiertas

1. **El trinquete de CRAP, ¿por componente o agregado?** Por componente es más útil y más
   ruidoso; agregado es más fácil de sostener y esconde regresiones locales.
2. **¿F6 acotado al código anclado a un solo `R#`, o al conjunto de la feature?** Acotado
   por `R#` da un veredicto atribuible; por feature es mucho más barato de correr.
3. **¿`sf audit` en alcance de proyecto debe fallar el CI, o sólo reportar?** Hoy `sf verify`
   es el agregado de CI; hay que decidir si se fusionan.
4. **El binding a `tree_hash` de F3: ¿qué pasa con un cambio que sólo toca comentarios?**
   Invalidar el sello es correcto y estricto, pero puede volverse insoportable en la
   práctica. La alternativa (hashear el AST y no los bytes) es más amable y más cara.
