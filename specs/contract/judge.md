# Contrato: `sf-audit` — el cerebro

**Estado:** ⏸ CONGELADO — hipótesis sin consumidor validado (ver [`README.md`](README.md)) · **Fecha:** 2026-08-07 · **El brazo:** [`audit.md`](audit.md)

Este documento es **el contrato de salida de `sf-audit`**: qué revisa, qué puede afirmar,
con qué forma, y cómo el CLI verifica cada afirmación antes de sellarla.

---

## 0. Para qué existe el cerebro — leer antes que nada

> **El CLI nació para producir artefactos que hagan que el LLM alucine menos.** El
> entregable es el **juicio**; el determinismo es lo que hace que ese juicio no sea
> inventado — material confiable a la entrada, evidencia comprobable a la salida.

El cerebro **DEBE** hacer una revisión **punta a punta** de la feature: ¿el código está
completo?, ¿es correcto?, ¿los tests prueban lo que dicen probar?, ¿cubre los requisitos?,
¿hay algo acá que nadie pidió? Es la revisión que un ingeniero senior haría, con la
diferencia de que **cada afirmación queda atada a una cita que el brazo resuelve**.

**Lo que lo mantiene honesto NO es la ceguera. Es la verificación (§5).**

### 0.1 Un solo actor, no dos

> **`sf-audit` *es* el juez.** No hay un CLI que audita y después un juez que opina sobre
> las sobras: hay **un subagente de contexto limpio** que revisa y firma, y un **brazo**
> (`audit.md`) que le ejecuta, le computa y le verifica las citas.

El diseño anterior tenía dos actores emisores de veredicto y eso escondía un problema: si
el CLI compone el veredicto, el skill queda de envoltorio que imprime salida ajena y **la
IA deja de trabajar**. Con un solo actor, la metodología vuelve a ser lo que era:

> **La IA trabaja. El CLI no la deja mentir.**

Lo que **no** se colapsa —y por eso no es un simple renombre— son las dos garantías, que
tapan modos de falla distintos y necesitan mecanismos distintos:

| Garantía | Tapa | Mecanismo | Dónde |
|---|---|---|---|
| **Determinismo** | fabricación (*"corrí los tests y pasaron"*) | el brazo ejecuta y sella | `audit.md` §3 |
| **Contexto fresco** | razonamiento motivado (juzgar lo que uno escribió) | subagente limpio, obligatorio | §3.2 |

Fusionarlas en un actor que también escribió el código rompe la segunda. Por eso `sf-audit`
**DEBE** ser un subagente, no la sesión principal.

> **Nota de vocabulario.** En este documento **cerebro** y **juez** nombran al **mismo
> actor**: `sf-audit`. Se dice *juez* cuando lo que importa es el rol adversario (§4) y
> *cerebro* cuando lo que importa es la división de trabajo con el brazo. No son dos
> piezas.

### 0.2 Los dos canales del informe

El informe que firma `sf-audit` mezcla dos clases de afirmación y **cada línea DEBE ser
atribuible a una de las dos** (`audit.md` §0.3):

| Canal | Origen | Regla |
|---|---|---|
| **Hecho** | el brazo | se cita **textual**. El cerebro **NO DEBE** reescribirlo de memoria |
| **Juicio** | el cerebro | prosa propia, **siempre** con cita `path:line` verificable |

Un número parafraseado deja de ser un hecho. Si el informe dice *"la mutación dio ~85%"* y
el brazo devolvió `62%`, el informe está fabricando — aunque el cerebro haya "redondeado"
de buena fe.

> **Corregido 2026-08-07.** La primera versión de este contrato restringía al juez a ver
> sólo el "residuo" (lo que las verificaciones mecánicas no pudieron decidir). Era un
> error, y quedó demostrado en la propia validación: en `slugify` el juez habría
> encontrado el parámetro `sep` sin requisito — pero sólo porque un mutante **no
> relacionado** dejó a `R1` en el residuo. Con `R1` limpio, el juez nunca veía la función.
>
> Restringir la **entrada** no agrega seguridad: la seguridad la da verificar la **salida**.
> Sólo le sacaba visión. Se elimina la restricción.

---

## 1. Qué es un cerebro, y por qué es peligroso

Un **cerebro** es un subagente LLM al que se le pide un **juicio** sobre algo que ningún
hecho puede decidir. Es la única pieza cooperativa que puede afectar un veredicto.

Es peligroso por tres motivos, y todo este contrato existe para tapar los tres:

| Modo de falla | Qué pasa | Cómo lo tapa este contrato |
|---|---|---|
| **El que miente** | dice "todo bien" sin haber mirado nada | no computa hechos (§3) y **no tiene el verbo "aprobar"** (§4.2) |
| **El interesado** | el mismo contexto que escribió el código lo juzga y racionaliza sus propias decisiones | contexto fresco obligatorio (§3.2), material **armado por el brazo** (§3.1) |
| **El complaciente** | preguntado *"¿está bien?"*, un LLM tiende a decir que sí | encuadre adversario: se le pide **refutar**, no aprobar (§4) |

> **"Armado por el brazo" no es "recortado por el brazo".** El brazo decide el material
> para que el cerebro no elija su propio examen — y ese material es la feature **entera**
> (§3.1). Acotar lo que ve **NO DEBE** usarse como mecanismo de control.

**Principio rector:**

> El cerebro **NO DEBE** poder emitir una afirmación que el brazo no pueda respaldar.

---

## 2. Un cerebro por fase, no uno solo

`sf-audit` es **un** actor (§0.1) — pero es el actor de **una** fase. A lo largo del
producto hay **tres**, con preguntas, insumos y momentos distintos. Colapsarlos fue un
error de diseño de v1: un solo revisor para spec, código y entrega negocia consigo mismo.

| Cerebro | Pregunta | Módulo | Estado actual |
|---|---|---|---|
| **de la spec** | ¿el requisito es testeable, no ambiguo, no inventado? | **Inception** | existe en `sf-check` (rúbricas `requirement-quality`, `minimal-code`) — **mal ubicado y `default: off`** |
| **`sf-audit`** | ¿el código cumple el requisito? ¿el test prueba *eso*? ¿hay algo que nadie pidió? | **Forge** | **no existe** |
| **de la entrega** | ¿el receipt cubre lo que se revisó? ¿esto se puede publicar? | **Ship** | no existe |

Los tres comparten la mecánica de §3–§5 y cada uno tiene su brazo. Difieren en el material
de entrada y en las rúbricas. Este contrato especifica la mecánica común y detalla
**`sf-audit`**, que es el que falta y el que hace posible no leer el código.

---

## 3. Insumos: qué recibe el cerebro

### 3.1 Una sola llamada, y el brazo arma el material

```
sf audit --feature=<x> --json
```

**Eso es todo.** No hay una secuencia de comandos que el cerebro tenga que correr en
orden: `sf audit` absorbió a `sf context for-judge` y devuelve hechos **y** material en una
sola invocación (`audit.md` §2.1 y §5.2).

**Reglas:**

1. El material **DEBE** salir de **un solo comando** del CLI. El cerebro **NO DEBE** leer
   archivos por su cuenta ni elegir qué mirar: no elige su propio examen.
2. El material **DEBE** contener **la feature entera**:
   - **todos** los requisitos con **todos** sus criterios de aceptación
   - **todo** el código y los tests anclados, completos — **incluidos los `R#` en
     `PROBADO`**
   - el diff de la feature
   - los hechos F1–F7 y el estado de cada `R#`, incluido el `residue`, **como contexto de
     sólo lectura**, para saber qué ya está mecánicamente establecido y dónde concentrarse
   - los principios/invariantes de la constitución cuyo `applies_to` incluye esta fase
   - las rúbricas que aplican, nombradas por el CLI
3. El material **NO DEBE** contener:
   - la conversación del build, los logs de progreso, ni el razonamiento del implementador
     — **esto sí es load-bearing**: es lo que evita el razonamiento motivado (§3.2)
4. El cerebro **NO DEBE** ejecutar nada cuyo resultado vaya al informe como hecho — ni el
   suite, ni la mutación, ni el análisis de complejidad. Si necesita que algo se ejecute,
   lo ejecuta el brazo (`audit.md` §2 regla 3). *"Corrí los tests y pasaron"* dicho por un
   LLM es el incidente que originó SpecForge, y vale igual para cualquier otra herramienta.

**El `residue` orienta, no limita.** El juez **DEBERÍA** empezar por ahí, porque es donde
los hechos ya avisaron que hay algo. Pero **PUEDE** pronunciarse sobre cualquier parte de
la feature, incluido un `R#` en `PROBADO`, siempre con cita verificable.

**Sobre un `R#` en `PROBADO`, el juez PUEDE señalar un caso no cubierto, y NO PUEDE
cambiarle el estado.** `PROBADO` es un hecho sobre lo que las verificaciones mecánicas
encontraron; el señalamiento del juez se registra **al lado**, como hallazgo, no encima.
El humano ve los dos.

### 3.2 Contexto fresco, obligatorio

El juez **DEBE** correr en un subagente con ventana limpia. **NO DEBE** ser la sesión
principal ni un agente que participó del build.

No es una preferencia de rendimiento. Un agente que escribió el código tiene **razonamiento
motivado**: recuerda por qué tomó cada decisión y las va a racionalizar. La independencia
es lo que hace que el juicio valga algo.

### 3.3 Separación de restricciones (patrón adversario)

Cuando corre más de un juez en el mismo perfil de rigor, cada uno **DEBE** recibir un
conjunto de reglas **disjunto**. Un juez que tiene dos criterios en conflicto negocia
consigo mismo y afloja los dos.

---

## 4. El encuadre: refutar, no aprobar

### 4.1 El prompt

El prompt del juez **DEBE**:

- pedir **refutación**, no aprobación
- exigir **cita** para toda afirmación
- prohibir explícitamente pronunciarse fuera del material recibido
- exigir salida **sólo JSON**, validable contra el schema de §4.3

Forma del prompt del juez de implementación:

> Sos un auditor adversario. Abajo tenés **una feature completa**: sus requisitos con sus
> criterios de aceptación, el código que los implementa, sus tests, el diff, y el resultado
> de las verificaciones mecánicas que ya corrieron.
>
> Hacé una revisión **punta a punta**: ¿el código está completo?, ¿es correcto?, ¿los tests
> prueban lo que dicen probar?, ¿cada criterio de aceptación está realmente cubierto?, ¿hay
> código acá que ningún requisito pidió?
>
> Tu trabajo **no** es aprobar. Es **intentar refutar**. Buscá casos concretos —una
> entrada, un estado, una rama— donde el código no cumpla un criterio, o donde un test pase
> sin ejercitar lo que el criterio pide.
>
> Las verificaciones mecánicas ya marcaron dónde hay problemas conocidos: **empezá por
> ahí**, pero no te limites a eso. Un requisito marcado como probado puede tener un caso
> que nadie cubrió — si lo encontrás, señalalo.
>
> Toda afirmación **debe** citar `path:line` del material que te dieron. Una afirmación sin
> cita verificable no se registra. No opines sobre estilo, performance ni arquitectura
> salvo que una regla explícita lo pida.
>
> Salida: SÓLO el JSON del schema. Sin prosa antes ni después.

### 4.2 Los cuatro veredictos posibles

El juez **DEBE** emitir exactamente uno por requisito de su jurisdicción:

| Veredicto | Significa | Atado a un `R#` | Requiere cita |
|---|---|---|---|
| `refuta_codigo` | encontré un caso concreto en que el código no cumple el requisito | sí | **Sí**, obligatoria |
| `refuta_spec` | hay código que **ningún requisito reclama** → la spec está incompleta | **no** | **Sí**, obligatoria |
| `no_refuta` | busqué y no encontré motivo para refutar | sí | **Sí** — debe citar el elemento que examinó |
| `no_juzgable` | necesito evidencia que no tengo | sí | No, pero **DEBE** nombrar qué evidencia falta |

**No existe `aprueba`.** Es deliberado y es la pieza central de este contrato.

> **`refuta_spec` agregado 2026-08-07 — ver [`validation/slugify.md`](validation/slugify.md) H4.**
> El contrato original ataba los tres veredictos a un `requirement`, así que el juez no
> tenía forma de reportar *"hay código que ningún requisito cubre"* — eso no es una
> afirmación sobre un requisito, es sobre **la ausencia** de uno. En `slugify`, el juez
> habría visto la firma `slugify(text, sep="-")` junto a un `R1` que no menciona `sep`, y
> habría tenido que meterlo a la fuerza en un `refuta` falso o callarse.
>
> El juez sigue sin poder aprobar: ahora **refuta en dos direcciones**. Es la regla de las
> dos rutas de `sf-amend` (*"nunca asumas que el que está mal es el spec"*) aplicada al
> revés: a veces el incompleto es el spec.
>
> **Diagnósticos distintos, arreglos distintos.** `refuta_codigo` → escribí un test o
> arreglá el código. `refuta_spec` → escribí el requisito, o borrá el código. Confundirlos
> manda al humano a hacer el trabajo equivocado, y eso destruye la confianza en el
> veredicto — que es todo el punto.

El veredicto positivo del juez es `no_refuta`, que es lo único epistémicamente honesto que
un LLM puede decir: *"no encontré el problema"*, no *"no hay problema"*. Al no tener el
verbo, el juez **estructuralmente no puede** funcionar como sello de aprobación general —
que es exactamente la mentira que originó SpecForge.

Consecuencia práctica: `no_refuta` **NO DEBE** renderizarse como ✓ ni como "aprobado" en
ninguna salida. Se renderiza como lo que es.

### 4.3 Schema de salida

```json
{
  "kind": "implementation",
  "feature": "auth-login",
  "audit_tree_hash": "3f2a1b...",
  "verdicts": [
    {
      "requirement": "R7",
      "result": "refuta_codigo",
      "claim": "El test sólo ejercita el camino de token expirado; AC-3 exige rechazar también el token con firma inválida, y esa rama no tiene assert.",
      "citation": {"path": "src/auth/login_test.go", "line": 42},
      "criterion": "AC-3"
    },
    {
      "requirement": "R4",
      "result": "no_juzgable",
      "claim": "El criterio es una latencia p95; no hay medición en el material.",
      "missing_evidence": "medición de p95 bajo carga representativa"
    }
  ]
}
```

**Reglas del schema:**

1. `audit_tree_hash` **DEBE** coincidir con el del `sf audit --json` que originó el
   material. Un veredicto emitido sobre un árbol y registrado sobre otro es inválido.
2. `claim` **DEBE** ser **falsable en una mirada**: un caso concreto que el humano pueda
   comprobar abriendo un archivo. *"El manejo de errores podría mejorar"* no es un claim.
3. `criterion` **DEBERÍA** nombrar el criterio de aceptación específico cuando el
   requisito tiene más de uno.

---

## 5. Cómo se verifica al juez

Esta es la sección que hace que el juez no sea un oráculo nuevo al que hay que creerle.

```
echo '<json del juez>' | sf gate record-verdict --kind=implementation --feature=<x>
```

`record-verdict` **DEBE** rechazar el veredicto —entero, no parcialmente— si:

| # | Condición de rechazo | Por qué |
|---:|---|---|
| 1 | el JSON no valida contra el schema de §4.3 | un veredicto malformado nunca debe existir |
| 2 | `audit_tree_hash` ≠ el del audit vigente | se juzgó otro árbol |
| 3 | un `refuta_codigo` pretende **cambiar el estado** de un `R#` que el audit dejó en `PROBADO` | el juez no revierte hechos: se registra como hallazgo al lado, no encima (§3.1) |
| 4 | un `requirement` no existe en el `trace.json` | requisito fabricado |
| 5 | una `citation.path` no existe en el árbol | **evidencia fabricada** |
| 6 | una `citation.line` está fuera del rango del archivo | evidencia fabricada |
| 7 | una `citation.path` está fuera del material que el CLI le entregó | citó algo que no vio |
| 8 | un `criterion` no existe entre los criterios de ese requisito | criterio fabricado |
| 9 | falta `citation` en un `refuta_codigo`, `refuta_spec` o `no_refuta` | afirmación sin evidencia |
| 10 | en un `refuta_spec`, la `citation` **sí** está reclamada por algún `R#` del `trace.json` | el juez afirmó que no hay requisito y sí lo hay — **mecánicamente comprobable** |

> La condición 10 es lo que hace que `refuta_spec` sea tan verificable como los demás: la
> afirmación *"esto no lo cubre nadie"* se comprueba contra el `trace.json` sin entender
> una palabra del claim. Misma jaula, veredicto nuevo.

**Rechazo = fail-closed.** El veredicto no se registra, el gate no avanza, y el motivo del
rechazo **DEBE** quedar en el ledger. Un juez que fabrica evidencia **DEBE** ser visible
como tal, no silenciado.

> **Esto es el detector de fabricación de `RM-C3` apuntado al juez.** La condición 5 es la
> importante: un LLM que inventa una cita es **detectable a máquina**. Por eso la cita es
> `path:line` y no prosa: para que el verificador no tenga que entenderla, sólo resolverla.

### 5.1 Qué queda garantizado y qué no

| | Garantía |
|---|---|
| **Puede equivocarse** | Sí. El juicio es falible por diseño. `refuta` puede ser un falso positivo; `no_refuta` puede pasar por alto un bug real |
| **Puede mentir** | **No.** Toda evidencia es verificable por la capa determinista, y el que la verifica no es otro LLM |

Esa asimetría es todo el valor del juez, y es la misma que `audit.md` §1.2 declara para el
sistema entero.

---

## 6. Registro y efecto

1. El veredicto aceptado se persiste en el ledger **encadenado por hash**, junto al
   `audit_tree_hash`, el perfil de rigor, y el modelo que lo emitió.
2. **Efecto sobre el gate**, por perfil de rigor:

| Perfil | `refuta` | `no_juzgable` |
|---|---|---|
| `lean` | (no corre juez) | — |
| `standard` | **nudge** — se le muestra al humano, no bloquea | se muestra |
| `strict` | **block** — el gate no avanza | escala a `sf evidence`; el gate espera |

3. Un `no_juzgable` **DEBE** abrir un ítem de `sf evidence` con el texto de
   `missing_evidence`. Ese es el canal por donde entra a la cadena de confianza lo que la
   máquina no puede probar.
4. El humano **PUEDE** anular cualquier veredicto. La anulación **DEBE** entrar al ledger
   como entrada propia, con autor y motivo. Anular sin dejar rastro **NO DEBE** ser posible.

---

## 7. Rendición de cuentas

Como el veredicto se guarda junto al `audit_tree_hash`, el material que el brazo entregó es
**reconstruible**: `sf audit --feature=<x> --json` sobre ese árbol devuelve lo mismo.

Consecuencia: cuando un juez se equivoca, se puede ver **exactamente qué vio**. Eso hace
que los jueces se puedan evaluar como se evalúa cualquier otra cosa — con casos y con
regresión — en vez de discutirlos por impresión.

Es también el insumo de la deuda que la auditoría marcó como `PRA-5`: **hoy hay ~650 tests
para la mitad determinista y 0 para la mitad cooperativa**, y la mitad cooperativa es la
que decide. Un juez sin suite de casos es la pieza más consecuente y menos probada del
sistema.

---

## 8. Lo que se reusa de lo construido

| Pieza | Hoy | Acción |
|---|---|---|
| Mecánica juez-fresco + `for-judge` + `record-verdict` | existe en `sf-check` | **reusar entera.** Está bien diseñada: el CLI hace el slicing, contexto fresco, cita obligatoria, `sf` registra y nunca juzga |
| `sf context for-judge` como comando propio | existe en `sf-check` | **absorber en `sf audit --json`.** No se pierde nada —el material sigue saliendo del CLI— y el cerebro pasa de dos llamadas a una (§3.1) |
| Rúbricas `requirement-quality`, `minimal-code` | en `sf-check` | **mover a Inception.** Juzgan la spec, no la implementación. Están en la fase equivocada |
| `audit.phase: off \| nudge \| block` | `default: off` | **reusar el mecanismo, cambiar el default.** Pasa a estar gobernado por el perfil de rigor (§6.2). Un juez apagado de fábrica es el eslabón que hace aceptable no leer el código, apagado |
| Verificación de la cita | **no existe** | **nuevo.** Es §5, y es lo que vuelve auditable al cerebro |
| **`sf-audit`** (cerebro de Forge) | **no existe** | **nuevo.** Es el que faltaba, y el que hace posible no leer el código |
| Cerebro de entrega | no existe | Ship, más adelante |

---

## 9. Preguntas abiertas

1. **¿El juez de implementación ve el diff o el archivo completo?** El diff es más barato
   y enfoca; el archivo completo permite detectar que la implementación rompe algo que ya
   estaba. Probablemente: diff + los símbolos anclados completos.
2. **¿`no_refuta` sin haber encontrado nada cuenta como evidencia de calidad?** Hoy el
   contrato dice que no bloquea y no aprueba. La pregunta es si acumular `no_refuta` a lo
   largo del tiempo dice algo o es ruido.
3. **¿Cuántos jueces en `strict`?** Uno adversario alcanza, o conviene el patrón de Uncle
   Bob (architect + hardener con reglas disjuntas). Más jueces = más costo real, medido:
   SwarmForge quemó 93% de una sesión en una kata.
4. **¿El juez puede pedir más material?** Un `no_juzgable` por falta de contexto es
   distinto de uno por falta de evidencia externa. Permitir un segundo turno lo hace más
   útil y más caro, y abre la puerta a que elija su propio examen.
