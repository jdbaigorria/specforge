# Validación del contrato contra `examples/slugify`

**Fecha:** 2026-08-07 · **Contrato bajo prueba:** [`../audit.md`](../audit.md)
**Método:** aplicación **a mano** de F1–F7 sobre el ejemplo real. Sin escribir Go.
**Pregunta que responde:** ¿el contrato dice algo útil, o produce un veredicto vacío?

---

## 0. Resultado en una línea

**Sí dice algo útil.** F6 encontró un **bug real** en 53 líneas de código que ya habían
pasado **siete gates aprobados** y estaban archivadas como `done`. Y el ejercicio destapó
**cinco defectos del contrato o del repo**, cuatro de los cuales no se veían desde el papel.

---

## 1. El material

| Artefacto | Contenido |
|---|---|
| Requisitos | **4** (`R1`–`R4`), EARS, con **5 criterios de aceptación** |
| Código | `src/texttools/slug.py` · función `slugify` · 12 líneas efectivas |
| Tests | `tests/test_slug.py` · 5 tests, uno por criterio |
| Ledger | 7 gates aprobados: lane → requirements → design → tasks → plan → wave-0 → verdict |
| Estado | `"status": "done"`, archivada `2026-06-16` |

---

## 2. Los hechos, uno por uno

### F1 — ANCLAJE · verde, con reserva

4/4 requisitos resuelven a `src/texttools/slug.py:slugify`. El archivo existe, el símbolo
existe.

**La reserva:** los cuatro apuntan **al mismo símbolo**. F1 confirma existencia, no
distingue qué requisito vive dónde. En una función de 12 líneas eso es correcto y honesto,
pero significa que **cuando el ratio requisitos:símbolos es N:1, F1 no aporta información**
— pasa siempre, y un hecho que siempre pasa no discrimina nada.

### F2 — PRUEBA EXISTE · verde · 5/5

| Criterio | Test | Existe |
|---|---|---|
| R1.1 | `test_basic` | ✓ |
| R2.1 | `test_accents` | ✓ |
| R3.1 | `test_punctuation_collapse` | ✓ |
| R4.1 | `test_punctuation_only_is_empty` | ✓ |
| R4.2 | `test_empty_input_is_empty` | ✓ |

**`RM-C1` (acceptance con id) se gana el lugar acá, visible.** `R4` tiene **dos** criterios
y **dos** tests. Con granularidad de requisito, un solo test habría dado `R4` por cubierto
y el otro caso quedaba sin probar. La decisión de `RM-C1` era correcta.

### F3 — VERDE SELLADO · **MISSING**

No existe `examples/slugify/specforge/.state/check/slugify.json`.

Verificado además: **ninguno de los tres ejemplos** (`slugify`, `lite-wordcount`,
`brownfield-tempconv`) tiene directorio `specforge/.state/`. No está gitignoreado — no está.

### F4 — FRESCURA · no computable

Depende de F3. Sin sello no hay `code_hash` contra el cual comparar. Correctamente
clasificado como **`missing`**, no como `stale` — la distinción del contrato funciona.

### F5 — DRIFT Y RIESGO DE CAMBIO · verde, con ruido

- **Complejidad ciclomática de `slugify`:** 2 (base 1 + el `if` del generador).
- **Cobertura de líneas:** 100% — los 5 tests recorren todo.
- **CRAP ≈ 2.** Riesgo bajísimo.
- **Huérfano potencial:** `src/texttools/__init__.py` (4 líneas, re-export de `slugify`)
  no está reclamado por ningún `R#`.

**Observación importante:** F5 dice *"riesgo bajo"* y F6 dice *"tests débiles"* sobre el
**mismo código**. No se contradicen — miden cosas distintas — pero prueba que **CRAP solo
no alcanza**: código simple y 100% cubierto puede tener tests que no prueban nada.

### F6 — FUERZA DEL TEST · **71% · POR DEBAJO DEL UMBRAL (80%)**

Siete mutantes generados sobre `slugify`, corridos contra los 5 criterios:

| Mutante | Resultado | Criterios que lo matan |
|---|---|---|
| M1 `[a-z0-9]+` → `[a-z]+` *(pierde dígitos)* | **SOBREVIVE** | — |
| M2 `[a-z0-9]+` → `[a-z0-9]*` | MUERTO | R1.1, R2.1, R3.1, R4.1 |
| M3 `NFKD` → `NFC` | MUERTO | R2.1 |
| M4 `not combining` → `combining` | MUERTO | R1.1, R2.1, R3.1 |
| M5 quita `.lower()` | MUERTO | R1.1, R2.1, R3.1 |
| M6 ignora el parámetro `sep` | **SOBREVIVE** | — |
| M7 default `sep="-"` → `"_"` | MUERTO | R1.1, R2.1, R3.1 |

**5/7 muertos = 71%.**

### F7 — COBERTURA DE CRITERIOS · verde · 5/5

---

## 3. Clasificación bajo el contrato

### 3.1 Tal como está escrito hoy

`ROTO` tiene precedencia (§4 regla 1) y F3 está `missing` → **los cuatro requisitos son
`ROTO`** → **exit 5**.

```
sf audit --feature=slugify

  ✗ NO SELLADO       no hay resultado de test sellado para esta feature
  ✗ ROTO             R1, R2, R3, R4 — sin sello no hay hecho que sostenerlos

  VEREDICTO   0 de 4 con prueba mecánica. Corré `sf check run` primero.
```

Es **correcto por contrato** y **casi inútil como diagnóstico**: dice *"todo roto"* cuando
lo que pasa es *"falta un paso"*. Eso es el hallazgo H2.

### 3.2 Suponiendo F3/F4 verdes (simulando la corrida de tests)

| `R#` | Estado | Motivo |
|---|---|---|
| **R1** | **`DÉBIL`** | M1 sobrevive: `slugify("Top 10 Songs")` daría `"top-songs"`. R1 dice *"return the words lowercased and joined"* y un dígito es parte de una palabra |
| R2 | `PROBADO` | los tres mutantes de unicode mueren |
| R3 | `PROBADO` | los mutantes de puntuación/separador mueren |
| R4 | `PROBADO` | ambos criterios (`"!!!"` y `""`) cubiertos |
| *`sep`* | **sin clasificar** | es superficie pública sin ningún requisito que la reclame |

```
  ✓ NO FABRICADO     4/4 R# → código + test · 5/5 criterios
  ✓ NO DRIFT         0 anclas huérfanas · riesgo de cambio 2 (bajo)
  ⚠ TEST DÉBIL       R1 · sobrevive 1 de 7 mutantes (71% < 80%)
                     el regex puede perder dígitos y ningún test lo nota

  VEREDICTO   3 de 4 con prueba mecánica. Mirá R1.
```

**Eso es exactamente lo que el contrato promete: te dice dónde mirar, y son 3 líneas.**

---

## 4. Los hallazgos

### H1 — Ningún ejemplo tiene sello de test · **grave, del repo**

Los tres ejemplos vidriera tienen cadena de gates aprobada, trace completo, y `status: done`
— **sin un solo resultado de test sellado**. Son proyectos que *parecen* haber pasado por
el pipeline y nunca lo corrieron: fueron escritos a mano y regenerados (`DOC-2` ya había
anotado que un render hubo que escribirlo a mano).

**Consecuencia:** `PRA-3` puso los tres ejemplos como suite de regresión determinista, pero
regresionan **artefactos**, no **evidencia**. Toda la capa que este contrato define como
"los hechos" está ausente del material que usamos para probar que el sistema funciona.

**Acción:** correr el pipeline de verdad sobre los tres ejemplos y commitear `.state/`.
Si `.state/` no debe versionarse, entonces los ejemplos **no pueden** ser la suite de
regresión de F3–F4 y hace falta otro fixture.

### H2 — El contrato colapsa "roto" con "no evaluado" · **del contrato**

`ROTO` mezcla dos cosas que necesitan acciones opuestas:

- *"el sello existe y no coincide"* → el trabajo está mal, hay que arreglar código
- *"el sello no existe"* → el trabajo puede estar perfecto, falta correr un paso

Es el mismo error que el contrato **prohíbe explícitamente** en F4 (*"`stale` y `missing`
NO DEBEN colapsarse nunca"*)… y que comete un nivel más arriba, en la clasificación.

**Acción:** agregar un sexto estado **`NO_EVALUADO`**, con su propio código de salida
(`7`), distinto de `ROTO`. Un requisito sin evaluar no es un requisito fallado.

### H3 — F6 encontró un bug real en la feature vidriera · **del código**

`slugify("Top 10 Songs")` → el regex `[a-z0-9]+` cubre dígitos, pero **ningún test los
ejercita**. Un cambio que los pierda pasa la suite entera y los siete gates.

No es un bug hoy —el código está bien— es un **agujero de red**: la protección contra esa
regresión no existe. Que es exactamente lo que F6 está para detectar.

**Acción:** agregar el criterio `R1.2: slugify("Top 10 Songs") == "top-10-songs"` y su test.

### H4 — `sep` es superficie pública sin requisito · **del contrato**

`slugify(text, sep="-")` expone un parámetro que **ningún requisito menciona**.

- **F5 no lo ve** — su detección de huérfanos es a granularidad de **símbolo**, y el
  símbolo `slugify` sí está anclado.
- **F6 sí lo ve** (M6 sobrevive) pero **lo reporta mal**: lo cuenta como *test débil*
  cuando en realidad es *requisito faltante*. Son diagnósticos distintos con arreglos
  distintos — escribir un test vs. escribir un requisito.

**Resuelto 2026-08-07 — dos capas que se suman, no compiten:**

| Capa | Qué hace | Límite |
|---|---|---|
| **F6 con alcance** *(elegido)* | cuando un mutante sobrevive, mirar en `trace.json` si ese código lo reclama algún `R#`. Sí → *test débil*. No → *código sin requisito* | sólo dispara donde cae un mutante |
| **El cerebro, con `refuta_spec`** | ve la firma completa y el requisito al lado; puede señalar el hueco por sentido, no por mutación | depende de que mire ahí — no hay garantía mecánica de que lo encuentre |

Descartado por ahora: F5 con detección **sub-símbolo** (parámetros, exports, campos). Da
el diagnóstico más directo pero es la más cara y la que más ruido hace — no todo parámetro
necesita su propio requisito. Queda guardada por si el límite de las otras dos molesta.

**Nota sobre el hallazgo del hallazgo:** que `sep` apareciera dependió de un accidente —
que sobreviviera un mutante de **dígitos**, que no tiene nada que ver con `sep`. Con el
contrato original, sin ese accidente `R1` salía `PROBADO`, el juez nunca veía la función y
F6 no corría ahí: las dos capas fallaban **juntas**.

> **Actualizado 2026-08-07.** Ese punto ciego era del contrato, no de la realidad, y ya
> está corregido: el cerebro ve la feature **entera** y el `residue` sólo lo orienta
> (`audit.md` §4 reglas 2–3, `judge.md` §3.1). Con `R1` limpio, el cerebro **igual ve**
> `slugify(text, sep="-")` con `R1` al lado.
>
> Lo que queda abierto ya no es de diseño sino **empírico**: ¿lo señala? Es el experimento
> pendiente — correr el prompt de `judge.md` §4.1 sobre `slugify` **con `R1` en `PROBADO`**
> y ver si encuentra H3 y H4 por sentido. Si los encuentra, la capa de juicio cubre el
> agujero de F6 y la corrección queda validada contra algo real. Si no, hace falta la
> tercera capa (F5 sub-símbolo, descartada arriba).

**Las dos capas tienen agujeros, en lugares distintos.** Está bien: el contrato no promete
ser exhaustivo (§1.1), promete **acotar dónde mirar**.

### H5 — F3 ya está construido · **corrección al contrato**

`audit.md` §3-F3 dice que el binding a `tree_hash` es *"el único cambio de mecánica nuevo"*.
**Es falso.** Ya existe en `cli/check.go`:

```go
CodeHash   string            `json:"code_hash"`      // hash del código al correr (freshness)
Tests      map[string]string `json:"tests,omitempty"` // resultado POR TEST
```

Y el comentario del archivo ya explica la intención completa: *"'corrí una vez y después
cambié todo' queda detectado"*. Más aún, `Tests` guarda el resultado **por test**, lo que
permite exigir que **cada test nombrado en el trace haya corrido y pasado en esa corrida**
— causalidad test→requisito, no sólo "la suite dio verde". Eso es **más** de lo que el
contrato le acreditaba.

**Acción:** corregir `audit.md` §3-F3. F3 y F4 no son trabajo nuevo: son **documentación de
lo que ya existe**, más el ruteo de su resultado al veredicto.

### H6 — Ruido de huérfanos en archivos barrel · **menor, del contrato**

`__init__.py` (4 líneas de re-export) queda sin reclamar. Marcarlo entrena a ignorar la
categoría entera.

**Acción:** F5 necesita exenciones declaradas (barrel/re-export/generado) o detección de
"archivo sin lógica propia".

---

## 5. Veredicto sobre el contrato

| Pregunta | Respuesta |
|---|---|
| ¿Produce un veredicto accionable? | **Sí** — 3 líneas y un requisito a mirar |
| ¿Encontró algo que los 7 gates no vieron? | **Sí** — H3 |
| ¿Encontró algo que los 596 tests no ven? | **Sí** — H1 |
| ¿Sirve para features chicas? | **Parcialmente** — F1 no discrimina con N:1 (§2-F1) |
| ¿Está listo para implementarse? | **No** — H2 y H4 son cambios de diseño, no de código |

**El contrato se ganó el lugar, con dos correcciones antes de escribir Go.**

Y el resultado meta que importa: **este ejercicio costó una hora y encontró seis cosas.**
Es el paso que no dimos con ninguna de las 24 features que la auditoría está sacando.
