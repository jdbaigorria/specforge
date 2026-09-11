# pstack — las ocho, y qué queda de ellas al mirarlas de cerca

**Fecha:** 2026-09-11 · **Fuente:** `inspiration/plugins/pstack` (MIT, de
[poteto](https://x.com/poteto), Cursor / React core team).

> **Estado: ✅ IMPLEMENTADO el 2026-09-11** — las seis de código y las dos de documento. El
> registro de lo construido, con lo que apareció en el camino, está en el §11.
>
> Este documento continúa a [`vecinos.md`](vecinos.md), que salió de
> leer el código de otros. Éste sale de lo mismo, con una diferencia: pstack **no tiene binario**.
> Es 100% prosa, y por eso lo que se puede robar no son compuertas — son **métodos**, que es
> exactamente lo que vive en los `sfx-`.

**Entra:** las ocho adopciones que salieron del análisis del 2026-09-11, contrastadas contra el
código y los skills que ya existen.
**Sale:** seis que se construyen, dos que ya estaban hechas, el orden, y la compuerta de cada una.

---

## 1. La regla que ordena las ocho, y hay que decirla primero

pstack y SpecForge apuestan al revés:

| | SpecForge | pstack |
|---|---|---|
| quién manda | un binario que **frena** | prosa: *"el agente debería"* |
| dónde vive el estado | `estado.json`, versionado | la todolist del chat |
| el skill | delgado, habla con la máquina | gordo, contiene el método |

**Y eso decide qué se puede robar.** pstack es bueno en el *cómo*: el método de una revisión, de
una comparación de opciones, de una prueba. Es malo en el *cuándo*, porque no tiene quién lo haga
cumplir — y se le nota en la prosa. Del playbook `feature`, paso 4:

> *"Mandatory: no skip-with-reason escape, and Laziness Protocol does not override it. (…) 'The
> app is small' and 'a subagent cannot spawn one' are both wrong."*

Eso es un parche de prompt tapando un agujero de compuerta. Es, literal, *"a suggestion the agent
can ignore"* — la frase del README de este repo.

**La regla que sale de ahí, y vale para las ocho:**

```
lo que se adopta va a un sfx-  o  a un skill de estado que ya existe
NUNCA a la máquina
    ↳ salvo UNA excepción, declarada y contada: la #1 toca revision.json
```

---

## 2. El resultado de mirarlas de cerca: dos ya estaban hechas

Antes de las seis que se construyen, las dos que se caen. **Las dos se caen por la misma razón:
el repo ya llegó solo al mismo diagnóstico, y en un caso llegó más lejos.**

### 2.1 La #6 — el mapa de cobertura de fuentes → **casi hecha**

Dije que faltaba *"documentá el nulo, no saltees la búsqueda"*. Es falso.
[`sfx-buscar`](../skills/sfx-buscar/SKILL.md) ya lo tiene, y con compuerta:

```
- Provenance on every claim. `retrieved` without a link gets demoted.
- What you did not find gets written down.
- No tools → no invented landscape. Stop, or declare `evidencia: baja`.
```

Y `no evaluable` ya es un resultado escribible: *"A green that means 'I did not look' is worse
than a red."* Eso **es** la regla de pstack, escrita antes y con el frontmatter contándola.

**Lo único que falta es más chico de lo que dije**, y es el roster: el `why` de pstack lista las
siete categorías de fuente y obliga a nombrar **la que no se consultó porque no había MCP**.
`sfx-buscar` tiene procedencia por afirmación, pero no un listado por nivel de *qué se intentó*.

> **El delta real, y son dos líneas:** que `evidencia.md` cierre con un bloque `## Fuentes` de una
> línea por nivel —incluidos los que devolvieron nada y los que no había llave para tocar— y que
> el frontmatter gane sus contadores.
>
> **Construido como dos y no como uno:** `consultado` y `sin_acceso`, cada uno con su etiqueta en
> el cuerpo (`[consultado]`, `[sin-acceso — razón]`). Un solo contador de ausencias no distingue
> "miré tres niveles y uno estaba vacío" de "miré uno solo", y ésa es justo la diferencia que el
> roster existe para mostrar.

**Veredicto: no es una adopción, es un ajuste de una línea en un skill que ya funciona.** Va al
final de la fila.

### 2.2 La #8 — principios citables por nombre → **hecha, y mejor**

Dije que `FUNDAMENTOS.md` era prosa y pstack tenía reglas direccionables. Media verdad: las reglas
direccionables existen hace rato, con nombre y número, en
[`que-sobrevive.md`](que-sobrevive.md) §2 — **y son siete, no seis. La R7 se me pasó leyendo, y se
cita una sola vez en todo el repo:**

```
R1  sf hace lo que tiene una sola respuesta correcta
R2  un artefacto tiene el tamaño de sus consumidores
R3  una compuerta frena sobre un hecho; un juez opina
R4  el skill no tiene convenciones propias: lee la constitución
R5  si no se puede escribir el test, no es un criterio de aceptación
R6  un campo deducible de otro es un campo que se desincroniza
R7  declarar el resultado o declarar la causa: gana el que se usa   ← 1 sola cita
```

Y las otras **se citan de verdad**, no de adorno: `sf-build` cita R4, `sf-check` cita R3,
`sfp-backlog` cita R3 y R5, `sfp-roadmap` cita R5, `sfp-po` cita R5. En `specs/` hay 22 citas
sólo en `que-sobrevive.md` y 14 en `superficie-sf.md`.

**Siete reglas citadas le ganan a veintitrés skills de una regla cada uno**, y por R2: veintitrés
carpetas nuevas contradicen el conteo de skills como característica del producto.

> **Que la R7 tuviera una sola cita es, en sí mismo, el argumento del hueco 1.** Una regla que
> vive en la línea 123 de un documento de 986 no se cita porque nadie la encuentra, no porque no
> sirva.

Quedan **dos huecos reales**, y los dos son chicos:

1. **Las R viven enterradas** en un documento de diseño de 986 líneas. El que escribe un skill
   nuevo no las encuentra. `FUNDAMENTOS.md` —el documento que *suena* como la casa de los
   principios— no las tiene.
2. **Falta la cláusula de falsabilidad**, que es lo único que pstack tiene y acá no. De
   `principle-build-the-lever`:

   > *"Applying this principle produces a file. If you cited it and there is no codemod, script,
   > generator, or delegate skill in the diff, you didn't apply it."*

   Citar una regla sin que la cita cueste nada es como un test que pasa con todo mockeado.

> **El arreglo entero:** mover las seis R al tope de `FUNDAMENTOS.md` como índice de seis líneas,
> con puntero al desarrollo en `que-sobrevive.md`, y agregarle a cada una **qué artefacto tiene
> que aparecer en el diff si la citaste**. R1 → un comando de `sf`. R3 → un contador en
> `compuerta.go`. R5 → un test nombrado. R2, R4, R6 → una línea borrada.

**Veredicto: una edición de documento, no una adopción.** No necesita código.

---

### El tablero, entonces

| # | Qué era | Veredicto | Toca |
|---|---|---|---|
| **1** | escalera de evidencia | **se construye** | `sf-check` · `revision.go` · `compuerta.go` |
| **2** | `arena` en el ⑫ | **se construye, en variante barata** | `sf-plan` · `compuerta.go` |
| **3** | skill de verificación generado | **se construye** | `sfx-verificar` (nuevo) · `sfp-constitucion` |
| **4** | `interrogate` multi-modelo | **se construye último, y es opcional** | `sfx-interrogar` (nuevo) |
| **5** | protocolo de ciego del banco | **se construye** | `banco/comparar.sh` · `banco/GUION.md` |
| **6** | mapa de cobertura de fuentes | ya estaba — queda un ajuste | `sfx-buscar` |
| **7** | unslop en español | **se construye** | `sfx-prosa` (nuevo) · `sf-cierre` |
| **8** | principios citables | ya estaba, y mejor | `FUNDAMENTOS.md` |

**Seis de código, dos de documento.** Y las seis tienen una dependencia que no es obvia: **la #3
es lo que hace alcanzable el escalón 5 de la #1.** No son ocho cosas sueltas.

---

## 3. La #1 — la escalera de evidencia

### 3.1 Por qué: `cumple` hoy no dice **cómo lo sabés**

[`sf-check`](../skills/sf-check/SKILL.md) ya mata el dolor más grande, y lo mata contando:

> *"`sf` no verifica tu juicio. Verifica que el juicio OCURRIÓ, sobre todos."*

Pero el juicio que ocurrió es un string de dos valores:

```go
// revision.go:68
Criterios map[string]string `json:"criterios"`   //  "us-3/CA-1" → "cumple"
```

Y ahí adentro caben cosas que no valen lo mismo:

```
"us-3/CA-1": "cumple"   ← lo corrí y lo vi andar
"us-3/CA-2": "cumple"   ← hay un test con ese nombre y asumo que prueba esto
"us-3/CA-3": "cumple"   ← lo leí y me pareció bien
```

**Las tres son el mismo byte.** El skill ya dice *"un criterio cuyo test no encontrás es
`no-cumple`"*, lo cual sube el piso — pero no distingue el techo, y el techo es donde vive la
mentira cara: el `cumple` de la tercera línea.

`blast-radius` de pstack tiene la escalera, y es lo mejor del plugin:

```
1  lo dijiste                          vale cero
2  señalaste la línea                  file:line real
3  mostraste que el caso malo no llega  caminaste la falla y no alcanza
4  lo corriste                          un script que falla fuerte si te equivocás
5  lo reprodujiste en la app corriendo
```

Con la regla que la hace servir: *"cualquier hecho que no llegue al 4, decilo. No lo escribas como
cerrado."*

### 3.2 Qué se adopta: el escalón como campo, no como prosa

```json
"criterios": {
  "us-3/CA-1": {"veredicto": "cumple",    "escalon": 4,
                "prueba": "internal/suite/suite_test.go::TestCierra"},
  "us-3/CA-2": {"veredicto": "cumple",    "escalon": 2,
                "prueba": "internal/brief/brief.go:42"},
  "us-3/CA-3": {"veredicto": "no-cumple", "escalon": 1, "prueba": ""}
}
```

**`prueba` es un puntero, nunca prosa** — la misma regla que `evidencia` en el TSV de
`show-me-your-work`. Un `file:line`, un `archivo_test.go::TestNombre`, o un comando.

### 3.3 Cómo: la migración no rompe lo archivado

`.docs/archivado/*/revision.json` tiene revisiones viejas con el formato string, y
`auditoria.go:289` las lee (`rev.Criterios[c] != "cumple"`). Romperlas sería romper al auditor.

**`UnmarshalJSON` propio en el tipo nuevo, que acepta las dos formas:**

```go
// revision.go
type Criterio struct {
    Veredicto string `json:"veredicto"`
    Escalon   int    `json:"escalon"`   // 0 = no declarado (formato viejo)
    Prueba    string `json:"prueba"`
}

// UnmarshalJSON acepta el string pelado de las revisiones archivadas:
//     "cumple"                          →  {Veredicto:"cumple", Escalon:0}
//     {"veredicto":"cumple","escalon":4} →  tal cual
//
// Escalon 0 NO es "escalón bajo": es "esta revisión es anterior a la escalera".
// La compuerta lo distingue; el auditor no necesita distinguirlo.
func (c *Criterio) UnmarshalJSON(b []byte) error { … }
```

Y `Criterios map[string]Criterio`. Los dos consumidores se tocan en una línea cada uno:

| archivo | hoy | queda |
|---|---|---|
| `auditoria.go:260` | `_, hay := rev.Criterios[c]` | igual, la existencia no cambia |
| `auditoria.go:289` | `rev.Criterios[c] != "cumple"` | `rev.Criterios[c].Veredicto != "cumple"` |

### 3.4 La compuerta: cuenta, no opina (R3)

`compuerta.go:733` cuenta hoy tres cosas. Gana **dos, y las dos son aritmética**:

```
④  todo criterio declara `escalon` ≥ 1
       → falla:  "us-3/CA-2 no declara escalón — no se sabe cómo lo sabés"

⑤  un `cumple` con escalon ≥ 4 nombra una `prueba` QUE EXISTE
       → falla:  "us-3/CA-1 dice escalón 4 y su prueba no existe:
                  internal/suite/suite_test.go::TestCierra"
```

**El ⑤ es el que atrapa la mentira**, y no es una opinión: es el mismo truco que ya usa
`sfx-audit` en su pregunta ④ (*"¿los tests que lo probaban SIGUEN EXISTIENDO?"*). El código para
buscar un test por nombre en el repo **ya está escrito** — se reusa, no se inventa.

### 3.5 El piso lo pone la constitución, no el binario

Cuál es el escalón mínimo aceptable **no es una sola respuesta correcta**, así que no es de `sf`
(R1): en un CLI un criterio llega al 5 barato, en una librería de tipos el 3 es el techo honesto.

Va al frontmatter que `sfp-constitucion` llena, donde ya viven `test_cmd` y `mutacion:`:

```yaml
verificacion:
  escalon_minimo: 3        # un `cumple` por debajo es un hallazgo del ㉑
```

Ausente ⇒ sólo se exige que esté declarado (la ④). **Que el piso lo elija el proyecto es lo que
evita el fracaso obvio:** una vara de 4 en un repo sin forma de correr nada convierte todas las
revisiones en rojas y a los dos días alguien afloja la compuerta.

### 3.6 Cómo se prueba

Cuatro tests en `compuerta_test.go`, y los dos que valen son los de no-frenar:

- una `revision.json` **archivada, formato viejo** ⇒ pasa. La escalera no es retroactiva.
- un `cumple` en escalón 2 con `escalon_minimo` ausente ⇒ pasa. Declarado alcanza.
- un `cumple` en escalón 4 cuyo test no existe ⇒ **falla, con el nombre del test en el mensaje.**
- un criterio sin `escalon` ⇒ falla.

---

## 4. La #2 — el rubro antes de las opciones

### 4.1 Por qué: tres opciones no garantizan tres opciones

[`sf-plan`](../skills/sf-plan/SKILL.md) ya tiene la compuerta más barata del flujo:

> *"`sf` cuenta los encabezados y exige exactamente 3. (…) una sola opción escrita como si fuera
> una comparación es la forma que toma una corazonada confiada."*

Pero contar encabezados ataca **el síntoma más grosero**, no el que queda. El que queda es peor
porque pasa la compuerta sin despeinarse:

```
el modelo ya decidió B mientras leía el us-#
   → escribe B bien
   → escribe A y C como espantapájaros, creíbles y peores
   → tres encabezados ✓  ·  el argumento que inclinó ✓  ·  compuerta verde
```

`arena` de pstack tiene el antídoto, y **no es el fan-out**:

> *"Derive the rubric. (…) The rubric is the picker's tool in Phase D. **Candidates only see the
> task.**"*

El rúbrico se escribe **antes** de ver las opciones. Es el mismo principio del ciego del playbook
`eval`, aplicado a una decisión de diseño en vez de a un experimento.

### 4.2 Qué se adopta, y qué explícitamente NO

**NO se adopta el fan-out.** `sf-plan` dice, con razón:

> *"No partas este trabajo. Los cinco pasos son una pasada con contexto ancho a propósito. (…)
> Nadie entra ni sale por el medio."*

Un `arena` real necesita N subagentes anidados adentro de un subagente. Eso depende de una
capacidad del arnés que [`agnostico-al-harness.md`](agnostico-al-harness.md) no puede asumir, y
rompe el bloque ⑫–⑯. **Queda fuera.**

Se adopta la mitad barata, que son tres cosas y ninguna cuesta un modelo más:

1. **El rúbrico primero.** `decision.md` arranca con `## Vara` — de 3 a 6 criterios evaluables,
   escritos antes de la primera opción.
2. **La tabla de puntaje**, criterio por criterio, no por impresión general.
3. **La regla de divergencia**, que es la que nadie tiene:

   > Si las tres opciones convergen en la misma forma, eso es señal fuerte: decilo y seguí.
   > Si las tres **divergen salvajemente, el ⑫ estaba mal planteado.** Replanteá el problema. No
   > promedies la divergencia.

### 4.3 Cómo: el template y un contador más

`templates/decision.tmpl.md` gana la sección de arriba:

```markdown
## Vara
<!-- ANTES de escribir la primera opción. Si la escribís después, la vara
     va a describir la opción que ya elegiste y esto no mide nada. -->
| criterio | cómo se evalúa |
|---|---|
| V-1 · … | … |

## A — <nombre>
## B — <nombre>
## C — <nombre>

## El puntaje
| | V-1 | V-2 | V-3 |
|---|---|---|---|
| A | … | … | … |

## El argumento que inclinó
```

Y el ⑰ gana un sexto chequeo, del mismo tipo que los cinco que ya corre:

```
⑥  decision.md tiene una sección `## Vara` con entre 3 y 6 filas `V-#`,
    y una tabla `## El puntaje` con una fila por opción
       → falla:  "decision.md declara 3 opciones y puntúa 2"
```

**Sigue sin opinar.** Cuenta filas, igual que cuenta encabezados.

### 4.4 El límite honesto de esto

Una vara escrita por el mismo modelo que después elige **no es un ciego de verdad**. Nada impide
escribirla ya sabiendo la respuesta.

Lo que sí hace, y alcanza para justificarla: **deja el fraude por escrito.** Una vara que sólo
premia lo que hace B queda en `decision.md`, versionada, y el ⑰ es un humano leyendo. Hoy no hay
ni eso. Y el día que el arnés soporte anidado, la vara ya está escrita y el fan-out es aditivo.

---

## 5. La #3 — el skill de verificación, y por qué es el más importante

### 5.1 Por qué: el escalón 5 hoy no lo puede alcanzar nadie

La escalera de la #1 tiene cinco peldaños y **el repo hoy sólo puede llegar al 4**, porque el 4 es
"corrí un test" y el 5 es "manejé la app de verdad". Nada en el flujo sabe arrancar la app,
ejercerla como la ejerce un usuario y capturar la prueba.

Y eso es exactamente el agujero que `sf-check` tiene declarado y no puede cerrar:

> *"Una suite verde prueba que los tests pasan. No prueba que agarrarían algo."*

Los mutantes (㉒) atacan eso desde adentro del test. **Nadie lo ataca desde afuera.** Un proyecto
puede tener 100% de mutantes muertos y una app que no arranca.

### 5.2 Qué se adopta

`create-verification-skill` de pstack genera un skill **project-local** con cinco secciones, y las
cinco están bien elegidas:

```
Launch     el comando exacto, y cómo sabés que está listo
Doctor     un chequeo de sólo lectura: ¿vale la pena manejar esta instancia?
Drive      el harness con los selectores REALES de este repo, no ejemplos
Evidence   qué se captura y dónde queda
Cleanup    matás lo que vos arrancaste, nunca por nombre de proceso
```

Más un **mapa de features** (uno por feature de usuario: qué es, cómo se llega, cómo se maneja,
qué estado final prueba que anda).

**Y la regla que lo separa de un README, que es el paso 4:**

> *"Corré sus propias instrucciones punta a punta una vez antes de entregarlo. (…) Un skill
> generado que nunca se ejecutó es un borrador, no un entregable."*

### 5.3 Cómo: un `sfx-` nuevo que genera otro skill

**`sfx-verificar`** — utilitario, standalone, no sabe que la máquina existe. Se invoca a mano.

```
sfx-verificar
   ├─ ① entrevista al REPO, no al usuario
   │     superficie · cómo arranca · cómo se maneja · qué evidencia deja · ¿aísla?
   │     (lee la constitución primero: test_cmd, stack, convenciones — R4)
   │
   ├─ ② escribe   .docs/verificar/SKILL.md         las cinco secciones
   │              .docs/verificar/features/*.md    el mapa
   │
   ├─ ③ LO CORRE UNA VEZ, entero, sobre UNA feature del mapa
   │     y si falla, lo arregla y lo vuelve a correr
   │
   └─ ④ deja la evidencia en  .docs/verificar/pruebas/
```

**Dónde se conecta con la máquina, que es la parte que importa:**

- **`sfp-constitucion` lo ofrece.** Es el estado donde se declara con qué se prueba este proyecto.
  Ofrecer, no obligar: un proyecto sin superficie manejable es legítimo.
- **El sobre del `revision` lo sirve.** Si `.docs/verificar/SKILL.md` existe, `sf context` lo
  agrega para `sf-check`. **Servir un archivo no es pensar ni lanzar a nadie** — es el mismo
  mecanismo que ya usa el sobre para los journals archivados y para `loQueRompio`.
- **Y ahí recién existe el escalón 5.** `sf-check` puede escribir `"escalon": 5, "prueba":
  ".docs/verificar/pruebas/f-2-login.txt"`, y la compuerta ⑤ de la #1 verifica que ese archivo
  exista. La cadena cierra.

### 5.4 El riesgo, y es el más grande de los seis

**Un skill generado se pudre.** El mapa de features envejece contra la app, y un mapa mentiroso es
peor que no tener mapa: manda a manejar una pantalla que ya no existe y el que lo lee concluye que
la app está rota.

pstack lo sabe y le dedica un skill entero (`maintain-verification-skill`). Acá hay una respuesta
más barata **porque la máquina ya lleva la cuenta**: el ㉓ (`sf-cierre`) cierra una feature, y ahí
sabe qué se construyó. Una línea en `sf-cierre`:

> Si `.docs/verificar/features/` existe y esta feature agregó superficie de usuario, agregá su
> archivo al mapa. Es el mismo movimiento que ya hacés con la doc.

**Aun así, éste es el único de los seis que puede fallar por abandono.** Va después de la #1 y la
#2 justamente por eso: si las dos primeras andan, la #3 tiene un consumidor esperándola (el
escalón 5) y no se seca sola.

---

## 6. La #5 — el protocolo de ciego del banco

### 6.1 Por qué: el banco ya se contaminó dos veces, y lo tiene escrito

[`banco/GUION.md`](../banco/GUION.md) es el documento más honesto del repo:

> *"El agente las encontró y concluyó solo que ya era momento de implementar. (…) **Las dos veces
> la corrida salía verde y medía cero.** No probás el scout: probás lo que el agente encontró
> tirado."*

Y de ahí salieron cuatro reglas de higiene: la semilla no se guarda en ninguna memoria, las
corridas viejas no se archivan adentro, nada que explique el experimento entra al directorio, cada
corrida va en un directorio nuevo.

**Eso es el protocolo de ciego descubierto a los golpes.** pstack lo tiene escrito completo, y
tiene tres cosas que el banco no:

| pstack | el banco hoy |
|---|---|
| lista negra de palabras en todo lo que el candidato ve (`eval`, `test`, `judge`, `rubric`, `arena`, `candidate`) | ✗ |
| slugs sanitizados: el directorio se llama como lo llamaría un usuario | parcial — `corridas/A` y `B` gritan experimento |
| el juez ve etiquetas, **nunca nombres de modelo** | ✗ — `comparar.sh` imprime `A` y `B`, que Javier ya sabe qué son |
| se califica desde el transcript, **no desde lo que el candidato dice que hizo** | parcial — §③ ya compara frontmatter contra cuerpo |

La cuarta es la que el banco ya inventó a medias y es la mejor: *"el frontmatter DECLARA; el
cuerpo es el HECHO. Si no cierran, el que vale es el cuerpo — y la diferencia es en sí misma un
hallazgo."*

### 6.2 Cómo: una sección ⓪ en `comparar.sh`, y es un lever

`comparar.sh` tiene siete secciones y ninguna pregunta **si la corrida vale**. Gana una, arriba de
todo, porque una corrida contaminada no se compara: se tira.

```bash
════════════════════════════════════════════
 ⓪ ¿ESTA CORRIDA MIDE ALGO?   (si no, nada de lo de abajo importa)
════════════════════════════════════════════
  ① ninguna palabra de la lista negra adentro del árbol de la corrida
       eval · test · judge · rubric · score · benchmark · candidate · arena
       → grep -ril, y cualquier hit es un ✗ rojo

  ② la respuesta de la semilla no aparece en ningún archivo del árbol
       (el que corre el banco declara la frase en $SEMILLA_SPOILER)

  ③ el directorio de la corrida no tiene nombre de experimento
       corridas/A → ✗   ·   corridas/costos-cli → ✓
```

**Es `principle-build-the-lever` aplicado tal cual:** la regla ya estaba escrita en `GUION.md` y
se rompió dos veces igual. Una regla que se rompe dos veces no necesita otra oración, necesita un
`grep` en el script. Es también R3 — la contaminación es **un hecho contable**, no un juicio.

### 6.3 Y la que no es del script: el juez ciego

`comparar.sh` §② imprime los veredictos rotulados `A` y `B`. Cuando el que compara es Javier, que
armó las dos corridas, eso está bien: ya sabe cuál es cuál.

**Cuando el comparador es un modelo, no.** Si alguna vez el banco crece a un juez automático, la
regla de pstack entra entera: el juez ve `corrida-1` / `corrida-2` barajadas, nunca el nombre del
modelo, y el mapeo se revela después del veredicto. Va escrito en `GUION.md` ahora aunque el juez
no exista todavía — cuesta tres líneas y evita que el día que exista nazca viendo los nombres.

---

## 7. La #7 — la prosa, y por qué `unslop` no se copia

### 7.1 Por qué

El repo escribe mucho documento largo en castellano: `sf-cierre` produce la doc del ㉓, `sfp-po` el
PRD, `sfp-constitucion` la constitución. Y ninguno tiene una vara de prosa. La única regla es R2
(el tamaño lo dan los consumidores), que es de **cuánto**, no de **cómo**.

### 7.2 Qué se adopta: la forma, no el contenido

`unslop` tiene 33 reglas y **están calibradas para inglés**. "Delve", "tapestry", "showcase",
"boasts" no aparecen nunca en un documento en castellano. Copiarlas sería un skill que no dispara.

**Lo que se adopta es el mecanismo, y son dos decisiones de diseño:**

1. **Números de regla estables**, que otros skills citan por número. `unslop` lo dice explícito:
   *"Rule numbers are stable ids that other skills cite. A removed rule leaves a gap."* Eso es
   diseño de API aplicado a prosa, y es lo que permite que `sf-cierre` diga "regla 14" en vez de
   repetir el párrafo.
2. **La regla 27, que es la única universal de las 33** y no depende del idioma:

   > *"Si la oración pudiera aparecer igual en la documentación de otro proyecto, no dice nada
   > sobre éste. Borrala."*

   Ésa sola justifica el skill.

### 7.3 Cómo: `sfx-prosa`, con vocabulario propio

Utilitario nuevo. Las reglas se derivan **de los documentos de este repo**, no se traducen:

```
P-1   la oración que serviría igual en otro proyecto        (la 27 de unslop)
P-2   el conector de relleno: "cabe destacar", "es importante mencionar",
      "en el mundo de", "a la hora de"
P-3   "no sólo X sino también Y"                            (la 9)
P-4   la voz pasiva con actor conocido: "se valida" → "la compuerta cuenta"
P-5   el adverbio apuntalando un verbo flojo                (la 30)
P-6   la raya larga                                          (la 13)
P-7   los dos puntos como conector a mitad de oración        (la 14)
P-8   el bold en cada sustantivo propio                      (la 15)
```

**Y el corolario de higiene, que es de este repo y no de pstack:** los `specs/` existentes son el
corpus de referencia. Un skill de prosa que no suena como `que-sobrevive.md` va a reescribir el
repo en otra voz, que es peor que no tener skill.

Se compone donde ya hay prosa larga: `sf-cierre` (la doc del ㉓) y `sfp-po` (el PRD). **No en
`sf-plan`** — `spec-design.md` lo lee un implementador y la regla anti-N/A ya lo gobierna.

---

## 8. La #4 — `interrogate`, y por qué va última

### 8.1 Por qué va última: el ㉑ ya es casi esto

`interrogate` lanza N revisores en N modelos distintos sobre el mismo diff y sintetiza por
consenso. Y `sf-check` **ya tiene las dos mitades caras**:

```
✓  revisor fresco que no implementó      "el que escribió el código ya se convenció"
✓  modelo grande                          sf next lo devuelve en `modelo:`
✓  matriz sobre TODOS los criterios       la compuerta lo cuenta
✓  mutantes de intención                  lo que una herramienta no piensa
✗  varios modelos, en desacuerdo
```

Lo único que falta es la diversidad de modelo. **Y el valor de eso es real pero acotado:**
`interrogate` dice que un hallazgo levantado por 2+ modelos de forma independiente es la señal más
alta. Contra un solo revisor, no sabés si un hallazgo es una verdad o una manía de ese modelo.

### 8.2 Qué se adopta, y qué se rechaza explícitamente

**Se rechazan los cuatro baldes** (`Act on / Consider / Noted / Dismissed`). Chocan de frente con
una decisión ya tomada y bien tomada: un hallazgo nace `abierto` y sólo Javier lo pasa a
`descartado`, por `sf dismiss`. Los cuatro baldes le devuelven al modelo la decisión de qué
importa, que es justo la que el diseño le sacó.

**Se adopta el mecanismo de consenso**, y se escribe en el vocabulario que ya existe:

```
2+ modelos lo levantaron solos  →  hallazgo, y el detalle lo dice
1 modelo lo levantó             →  hallazgo igual, y el detalle dice "un solo revisor"
contradicción entre modelos     →  hallazgo, y es el más interesante de los tres
```

Los tres nacen `abierto`. **La decisión sigue siendo de Javier; el consenso es información que le
llega, no un filtro que le sacan.**

### 8.3 Cómo

**`sfx-interrogar`** — utilitario, se invoca a mano sobre una feature de alto riesgo antes del ㉑,
o sobre un diff cualquiera fuera de SpecForge. No sabe que la máquina existe.

Salida: un bloque de hallazgos con el formato de `revision.json` (`origen: 21`, `criterio`,
`estado: abierto`, `detalle`), **listo para pegar**. Si el ㉑ lo consume, lo consume como pega el
resto; si se corrió standalone, es un informe.

### 8.4 Por qué es opcional, dicho sin vueltas

Es el más caro de los seis (N modelos por revisión) y el que menos agujero tapa, porque el ㉑ ya
cubre casi todo. **Se construye último, y si después de la #1 y la #3 el ㉑ deja de encontrar
cosas, no se construye.** Un skill que no gana nada sobre lo que ya hay es deuda con cara de
feature.

---

## 9. El orden, y no es el de la lista

```
①  la escalera        sf-check · revision.go · compuerta.go
      ↓ abre el hueco del escalón 5, que hoy nadie puede llenar
③  sfx-verificar      el que lo llena
      ↓
②  la vara del ⑫      independiente — puede ir en paralelo, no depende de nada
⑤  el ciego del banco independiente — y es el más barato de todos
      ↓
⑥  el roster de sfx-buscar     dos líneas
⑧  las R al frente             una edición de documento
      ↓
⑦  sfx-prosa
      ↓
④  sfx-interrogar     último, y sólo si el ㉑ todavía se pierde cosas
```

**Tres razones de orden, y ninguna es de tamaño:**

1. **La ① antes que la ③** porque la ③ sin la ① genera evidencia que nadie consume, y lo que nadie
   consume se pudre. Al revés, la ① sin la ③ igual sirve: el escalón 4 ya es alcanzable hoy.
2. **La ⑤ se puede hacer cualquier día** y es media hora. Es el mejor candidato para arrancar si
   se quiere una victoria corta: ataca un fallo que **ya ocurrió dos veces documentadas**.
3. **La ④ al final, con permiso de no hacerse.** Es la única de las seis que puede resultar
   innecesaria, y decidirlo mirando el ㉑ funcionando es más barato que decidirlo ahora.

---

## 10. Lo que no se adopta, y por qué

| de pstack | por qué no |
|---|---|
| `poteto-mode` entero | es un orquestador que compite con `sf next`, y parte de *"no creo en planificar, la mejor spec es el código"*. Incompatible con el producto. |
| los 23 `principle-*` | R2. Seis R citadas le ganan a veintitrés carpetas. §2.2 |
| `recall` | mina transcripts de chat para reconstruir contexto. SpecForge ya resolvió eso poniendo el estado en el repo — es su tesis. |
| `why` con siete investigadores MCP | 7 subagentes por pregunta. `sfx-buscar` cubre la parte barata con procedencia y compuerta. |
| `babysit` · `shipping` · `autopilot-*` | `sfx-github` + la máquina ya son el flujo de entrega. |
| `arena` como fan-out | rompe *"el bloque ⑫–⑯ es una pasada"* y asume anidado en el arnés. §4.2 |
| `make-bot-ui` · `benny` · `watch-pr` | infra de Cursor. Cero superficie acá. |
| los cuatro baldes de `interrogate` | le devuelven al modelo la decisión que `sf dismiss` le sacó. §8.2 |

**Y una nota de método sobre todo el plugin:** los slugs de modelo de pstack
(`grok-4.6-fast-xhigh`, `gpt-5.6-sol-max`, `claude-fable-5-1-thinking-max`) son de su entorno.
Acá los modelos se nombran por **alias** desde el menú, nunca por id — la regla del ⑯ de
`sf-plan`, y ninguna de las seis adopciones la toca.

---

## 11. ✅ Construido — 2026-09-11

Las seis de código y las dos de documento, en el orden del §9.

```
⑤  banco            comparar.sh §⓪ · GUION.md          4 casos probados a mano
⑧  las R al frente  FUNDAMENTOS.md · siete, no seis    + la cláusula de falsabilidad
⑥  roster fuentes   sfx-buscar · compuerta · 3 tests   [consultado] · [sin-acceso]
①  la escalera      revision.go · compuerta · sf-check 6 + 8 tests · avisa lo viejo
②  la vara          sf-plan · decision.tmpl · compuerta 5 tests · el ⑰ pasa a 6
③  sfx-verificar    skill nuevo · sobre · sf-cierre    3 tests
⑦  sfx-prosa        skill nuevo · sf-cierre · sfp-po   P-1…P-10
④  sfx-interrogar   skill nuevo                        sin los cuatro baldes
```

**Los cuatro chequeos del CI corridos a mano:** los 25 `name:` coinciden con su carpeta, ninguna
ruta `specforge/` quedó viva, y **los 12 comandos que los skills nombran existen los 12 en el
router de `main.go`**.

### Seis cosas que aparecieron escribiendo, y no estaban en el plan

- **`grep -rilF` aborta en el grep de Git-Bash** —exit 134, sin salida— así que el chequeo del
  spoiler **daba verde sin haber mirado**. Es exactamente la clase de falla que la sección ⓪
  existe para atrapar, encontrada adentro de la sección ⓪. Se cambió por `-ril` con los
  metacaracteres escapados, y el porqué quedó escrito en el script para que nadie lo
  "simplifique" de vuelta.
- **`suite.Faltantes` ya existía**, así que la compuerta ⑤ de la escalera no inventó nada: usa la
  misma función que la pregunta ④ de `sf audit`. La reutilización que la spec prometía era real.
- **Hizo falta `esArchivada()`**, que la spec no previó. Una revisión vieja y una nueva que no
  declaró el escalón **se ven igual desde el JSON**: las dos tienen escalón 0. Distinguirlas
  mirando si la feature está en `.docs/archivado/` es lo único que evita volver roja toda feature
  cerrada.
- **`## Vara` casi choca con `reOpcion`.** El regex es `^##\s+([A-Z])\s`, y `Vara` empieza con `V`
  mayúscula: se salva sólo porque después de la `V` viene una `a` y no un espacio. Quedó un test
  (`TestLaVaraNoSeCuentaComoUnaOpcion`) que lo fija, porque si el regex se afloja la vara pasaría
  a contar como una cuarta opción y el test de las tres fallaría por una razón sin relación.
- **`GUION.md` decía "las tres reglas" sobre un bloque de cuatro.** Misma familia que el
  "Dieciocho": un encabezado que quedó de una versión anterior del bloque de abajo. Arreglado.
- **Y el `decision.tmpl.md` iba a repetir el bug de `mutacion`.** El primer borrador traía
  `escalon_minimo: 3` pre-completado — *"una casilla ya completada con una respuesta válida"*, que
  es literalmente el fallo que el ⑧ ya había arreglado una vez. Quedó comentado.

### Cinco tests que ya fallaban en `HEAD`, y no son de esto

Comprobado en un worktree limpio sobre `47e6dcc`, antes de tocar nada:

```
TestElSobreApuntaALoArchivado            auditoria   compara rutas con "/" y Windows da "\"
TestElSobreDelBugTraeLaSpecDeLoQueRompio sobre       la misma, en otro test
TestLaPlantillaLlevaLaMarca              andamio     la plantilla no empieza con "# SpecForge"
TestElInformeDiceDondeEncontroCadaSkill  doctor      no encuentra los skills en ~/.claude/skills
TestSeAgotaLaEsperaYElRegistroParcialQueda lanzar    se agota la espera (7s)
```

**Los dos primeros son el mismo bug, y muerde en la máquina de Javier**, que es Windows: el sobre
arma las rutas con `filepath.Join` —y ahí Windows pone `\`— mientras el test las busca con barras
normales. Es un arreglo de una línea por test, o de una función si se quiere arreglar bien.

No se tocaron acá —son otro trabajo, con su propia razón— pero quedan nombrados para que no se
confundan con la cuenta de esta ronda. **La suite de esta ronda no agregó ningún rojo.**

---

## 12. Los riesgos, dichos antes de empezar

1. **La ① agranda `revision.json`.** De un mapa de strings a un mapa de objetos. El
   `UnmarshalJSON` doble (§3.3) es lo único que impide que las revisiones archivadas se vuelvan
   ilegibles — **y si ese test no existe, el auditor se rompe en silencio y nadie lo ve hasta la
   próxima `sf audit`.** Es el primer test que hay que escribir, antes que el tipo.

2. **La ③ se pudre sola.** §5.4. El anclaje en `sf-cierre` es lo único que la mantiene viva, y es
   una línea de prosa, no una compuerta. Si a las tres features el mapa ya miente, **se apaga**:
   un mapa mentiroso es peor que ninguno.

3. **La ② puede degenerar en teatro.** Una vara escrita por el que ya eligió no es un ciego.
   §4.4. Se acepta porque deja el fraude por escrito y versionado, no porque lo impida.

4. **La ⑦ puede reescribir el repo en otra voz.** Un skill de prosa entrenado contra reglas
   traducidas del inglés va a sonar a otro proyecto. Por eso el corpus de referencia son los
   `specs/` de acá, y por eso no toca `spec-design.md`.

5. **Seis adopciones no son seis skills.** Son tres nuevos (`sfx-verificar`, `sfx-prosa`,
   `sfx-interrogar`) y tres ediciones de lo que ya hay.

   **Y al ir a actualizar el conteo apareció que ya está mal.** `docs/skills.md` se contradice
   solo:

   ```
   línea  3   "Dieciocho.  Nueve saben hacer un estado; nueve son utilitarios"
   línea 48   "Los trece utilitarios"          ← y la tabla lista trece
   el repo     9 de estado + 13 sfx- = 22
   ```

   La cabecera quedó de cuando había nueve utilitarios y nadie la movió al agregar los cuatro.
   **Se arregla antes de tocar nada** —pasa a veintidós— y después las tres altas lo llevan a
   veinticinco. Arreglarlo después sería escribir el número nuevo sobre uno que ya mentía.

6. **Y el chequeo del CI que cruza el `.md` con el binario vale para las seis.** De
   [`el-mapa.md`](el-mapa.md) §12: *"un skill nombra un comando o una ruta, y eso es una promesa
   verificable."* Cada adopción que nombre `.docs/verificar/`, un campo de `revision.json` o un
   comando nuevo pasa por los dos greps antes de darse por terminada.

6. **Y el chequeo del CI que cruza el `.md` con el binario vale para las seis.** De
   [`el-mapa.md`](el-mapa.md) §12: *"un skill nombra un comando o una ruta, y eso es una promesa
   verificable."* Cada adopción que nombre `.docs/verificar/`, un campo de `revision.json` o un
   comando nuevo pasa por los dos greps antes de darse por terminada.
