# Fundamentos — cómo funciona SpecForge

**Fecha:** 2026-09-07 · **Branch:** `refundation` · **Commit base:** `6880bd8`

> **Qué es esto.** Las cuatro cosas que quedaron claras después de escribir el código, puestas
> como modelo. No es un informe de estado (eso es `el-mapa.md`), no es un plan de corridas (eso
> es `por-tramos.md`). **Es contra lo que se mide el código de acá en adelante.**
>
> Es corto a propósito. Hay ~790K de specs en esta carpeta y ese es el problema, no la solución.
> El antídoto no es un documento más grande: es uno chico que los demás tengan que respetar.
>
> **Regla de lectura.** Todo lo que dice "el código hace X" tiene archivo y línea. Lo que es
> deducción está marcado como deducción. Nada por suposición.

---

## Por qué existe este documento

Javier, el 2026-09-07:

> *"ahora tenemos más claro varias cosas, y capaz es momento de pivotear para bajar a tierra
> mejor cosas que fueron saliendo en nuestras conversaciones."*

No es tirar el Go. Es que **el modelo mental maduró después de que se escribió el código**, y en
dos de las cuatro cosas el código todavía encodea el modelo viejo. Eso no se arregla parchando:
se arregla escribiendo primero cuál es el modelo, y después moviendo el código donde diverge.

---

## Las siete reglas — el índice

**Se citan por número en todo el repo** (`sf-build` cita R4, `sf-check` cita R3, `sfp-backlog`
cita R3 y R5), y hasta hoy vivían sólo adentro de
[`que-sobrevive.md`](que-sobrevive.md) §2, en la línea 46 de un documento de 986. El que escribe
un skill nuevo no las encontraba.

**El desarrollo de cada una sigue estando allá.** Acá está la regla, y **qué artefacto tiene que
aparecer en el diff si la citaste**.

| | La regla | Si la citaste, en el diff hay |
|---|---|---|
| **R1** | `sf` hace lo que tiene una sola respuesta correcta | un comando de `sf`, o una función determinista — no un párrafo en un `.md` |
| **R2** | Un artefacto tiene el tamaño de sus consumidores | un consumidor nombrado por su estado, o una sección **borrada** |
| **R3** | Una compuerta frena sobre un hecho; un juez opina | un contador en `compuerta.go`, o una decisión devuelta a Javier |
| **R4** | El skill no tiene convenciones propias: lee la constitución | un default cableado **borrado** del skill y un campo leído de la constitución |
| **R5** | Si no se puede escribir el test, no es un criterio de aceptación | un test nombrado, o el criterio movido a *"Qué NO entra"* |
| **R6** | Un campo deducible de otro es un campo que se desincroniza | un campo **borrado**, y la función que lo deriva |
| **R7** | Declarar el resultado o declarar la causa: gana el que se usa | los consumidores de cada campo, contados |

### La columna de la derecha es la mitad que faltaba

Sale de leer `pstack` ([`pstack.md`](pstack.md) §2.2), y es lo único que ese plugin tiene y acá no
estaba:

> *"Aplicar este principio produce un archivo. Si lo citaste y no hay codemod, script, generador
> ni skill delegado en el diff, no lo aplicaste."*

**Citar una regla sin que la cita cueste nada es un test que pasa con todo mockeado.** Cuatro de
las siete se cobran en **líneas borradas**, y eso es a propósito: R2, R4, R6 y media R5 son reglas
de sustracción, y una sustracción que no borró nada no ocurrió.

> **Y no son un principio más cada una.** Siete reglas citadas le ganan a veintitrés skills de una
> regla cada uno, por R2: un skill que nadie abre no tiene consumidores.

---

## ① La máquina de estados

### El modelo

Hay **un solo carril**. No hay máquina paralela para bugs, ni modo especial, ni carril rápido.
Hay un campo que **saltea estados**, y eso es todo:

```
tipo: us    →  planificacion → implementar → revision → cierre
tipo: bug   →                  implementar →            cierre
```

Cinco estados de producto que corren **una sola vez** (brief → prd → constitución → backlog →
roadmap), y seis de feature que corren **una vez por feature**: cuatro con trabajo
(`planificacion`, `implementar`, `revision`, `cierre`) y dos de espera, donde la feature existe
pero nadie la está tocando (`planificada`, `cerrada`).

**No existe "pendiente".** Una feature que está en el roadmap y no arrancó simplemente no aparece
en el mapa. Se deduce de la ausencia, así que no se guarda.

### Dónde está el código

**A la altura.** `internal/estado/estado.go:64-71` declara los seis. Y la regla del salteo vive en
`Efectivo()` (`estado.go:95-100`) — **una función, no un `if` repetido**, con el motivo escrito
arriba:

> *"la REGLA vivía copiada en `sf next` y en `sf done`, y faltaba en `sf lote start` y en
> `sf context` — así que los comandos se contradecían sobre la misma feature."*

Eso ya es el entendimiento maduro, escrito en el código y con la cicatriz a la vista.

**Veredicto: ✅ no necesita refundación.**

---

## ② Los loops

### El modelo

**Todo bucle tiene dueño y tiene corte.** Un bucle sin número de salida no es un bucle: es una
trampa. Si algo puede volver a un estado por el que ya pasó, tiene que estar declarado:

- **cuál es** el bucle (qué estado vuelve a qué estado)
- **quién lo cierra** (qué comprobación decide que ya está)
- **con qué número** se corta si no cierra

### Dónde está el código

**A medias.** El mecanismo existe: `TopeIntentos = 3` (`internal/maquina/maquina.go:102`), y se
chequea **primero**, antes de mirar en qué estado está la feature (`maquina.go:495`). Eso está
bien pensado: si estás trabado, no importa dónde.

Pero mirá dónde vive cada pieza:

| | dónde | cuántos |
|---|---|---|
| se chequea | `maquina.go:495` | 1 |
| se incrementa | `done.go:240, 271, 279` | 3 |
| se resetea | `done.go:285`, `paradas.go:677, 754` | 3 |

**Los tres incrementos están adentro de `cerrarLote`.** O sea: el corte cubre **un solo bucle**,
el de implementar. Y el bucle de la revisión, que también existe y también puede girar para
siempre, no lo toca nadie.

Peor: `cerrarRevision` (`done.go:296-318`) manda a `implementar` y abre un lote de corrección. Ese
lote, cuando cierra bien, **resetea el contador a cero** (`done.go:285`). Entonces el bucle
`revision → implementar → revision` **vuelve a arrancar de cero cada vuelta**. ME TRABÉ no puede
dispararse ahí nunca.

**Y esto no es un olvido — es que nadie escribió nunca la lista de bucles.** El código tiene *"un
contador que algunos caminos tocan"*, no *"todo bucle tiene corte"*. Son cosas distintas.

**Veredicto: 🟡 falta declarar la lista de bucles. Escribirlo es chico; cablearlo es mediano.**

---

## ③ `sf` valida como si fuese un test

### El modelo

Ésta es **la que más cambió**, y la que explica el fracaso de la v1.

La v1 ponía al humano de gate en cada etapa, y terminabas firmando cosas que no requieren
criterio: un check después de implementar, una auditoría, una validación de seguridad. Eso no es
una decisión — **es algo que se tiene que hacer**. Firmarlo es teatro.

La distinción que resuelve eso:

> **Comprobar ≠ decidir.**
> Comprobar es mecánico: lo hace la máquina, siempre, sin preguntar.
> Decidir es criterio: lo hace Javier, sólo donde de verdad hay algo que elegir.

Y entonces: `sf` **comprueba todo** y en las paradas **entrega un acta**, como el reporte de una
corrida de tests. El humano no valida — el humano **lee la evidencia y decide**.

Un test no tiene dos salidas. Tiene **tres**:

```
✓  pasó              lo comprobé, salió bien
✗  falló             lo comprobé, salió mal
?  NO EVALUABLE      esto NO lo pude comprobar
```

La tercera es la que importa. *"No pude comprobar si esto ya existe"* es información que cambia
una decisión, y hoy no tiene dónde vivir. **Un verde que en realidad quiere decir "no miré" es
peor que un rojo.**

El acta tiene tres bloques y dos botones:

```
COMPROBÉ            los ✓, con el hecho al lado
MEDÍ                los números crudos (17 retrieved / 1 model-prior / 13 links)
NO PUEDO COMPROBAR  lo que quedó fuera del alcance de la máquina

           [ aprobar ]   [ rechazar "motivo" ]
```

Rechazar **es** rehacer: el motivo es obligatorio, porque *"sin él, el que rehaga vuelve a
proponer lo mismo"* (`paradas.go:434`).

### Dónde está el código

**Encodea el modelo viejo, y lo dice con todas las letras.** La cabecera de `internal/compuerta`:

> *"sf no maneja el auto —no agarra el volante ni elige la ruta— pero **DA VERDE O ROJO, Y EN
> ROJO NO SE PASA**."*

Ésa es la metáfora de la **compuerta**: una barrera, dos salidas. Y la estructura la sigue al pie
de la letra (`compuerta.go:59-70`):

```go
type Resultado struct {
    Fallas []string   // por qué NO se mueve
    Avisos []string   // lo que hay que saber y no frena
}
```

**No hay campo para los ✓ ni para el "no evaluable".** No es un descuido: `Texto()` tira los ✓ a
propósito, y el comentario explica por qué (`compuerta.go:85-87`):

> *"Los ✓ no se listan: sólo importa lo que falta. Un veredicto que enumera todo lo que salió
> bien es ruido, y el que lo lee tiene que buscar la ✗ entre quince ✓."*

**Ese razonamiento es correcto para el lector que tenía en mente — el subagente, que sólo necesita
saber qué le falta.** Es incorrecto para el lector nuevo, que es Javier decidiendo. **Son dos
lectores distintos y hoy hay un solo texto.**

Y esto es exactamente lo que Javier midió en la corrida real de T1 (2026-09-05): dos modelos
sellaron veredictos opuestos sobre la misma idea, uno con 17 fuentes recuperadas y 13 links, el
otro con 1 y **cero links** — y `compuerta.Brief` aceptó los dos, porque sólo miraba que el
veredicto fuese uno de tres. **Los números existían y no llegaban a los ojos de nadie.**

Lo que cambia:

- `Resultado` gana la tercera salida y un lugar para los ✓ y los números medidos.
- `Texto()` se parte en dos lectores: el del subagente (sólo lo que falta) y el acta de la parada
  (todo, con evidencia).
- El nombre del artefacto es **`acta`**. No "sobre" — ése ya es lo que `sf context` le sirve al
  subagente (`superficie-sf.md` H2/H13). No "parte" — ya existe `type Parte struct` en
  `internal/sobre/sobre.go:77`.

**Veredicto: 🔴 refundación real. Chica en tamaño (un paquete), grande en consecuencia.**

---

## ④ Los puntos de entrada

### El modelo

**Las puertas se nombran por lo que vas a hacer, no por dónde estás parado en el ciclo.**

La máquina sirve cuando hay que **hacer un desarrollo**. La pregunta al entrar es siempre la
misma: *¿qué vas a hacer?* — y la respuesta elige la puerta y el tramo de máquina que corre.

### Dónde está el código

**Es el más flojo de los cuatro.** `internal/maquina/entradas.go` traza bien el embudo:

> *"las tres entradas convergen en el backlog: **EL BACKLOG ES EL EMBUDO**."*

Pero las tres entradas de hoy no responden "qué vas a hacer" — responden "en qué parte del ciclo
de la entrada A estás":

| entrada | qué es en realidad |
|---|---|
| **A** producto nuevo | no es una entrada: es **el producto entero desde cero**, los cinco estados |
| **B** feature | exige `ConstitucionSellada` (`entradas.go:49`) — sólo existe *dentro* de un producto que ya pasó por A |
| **C** bug | idem B |

**Y falta la cuarta: un proyecto que ya existe.** Hoy `sf` le contesta *"este producto todavía no
está armado: seguí con `sf next` desde el brief"* — o sea, le pide a un repo con tres años de
código que primero se invente un brief y decida si vale la pena construirlo.

La pregunta abierta, que **no está resuelta y hay que resolverla**: para meter un ticket en un
proyecto que ya existe, ¿hacen falta los cuatro artefactos, o alcanza con la constitución?
*(Deducción, no verificado: la constitución sola parece alcanzar, porque es el único artefacto que
las compuertas de feature leen de verdad — pero eso hay que medirlo, no suponerlo.)*

**Veredicto: 🔴 refundación real. El más grande de los cuatro, y el que más abre.**

---

## El tablero

| | fundamento | código | qué cambia |
|---|---|---|---|
| ① | máquina de estados | ✅ a la altura | nada |
| ② | loops | 🟡 hay contador, falta la lista | declarar los bucles; cablear el corte de revisión |
| ③ | sf valida como un test | 🔴 dos salidas donde van tres | `Resultado` + `Texto()` partido en dos lectores |
| ④ | puntos de entrada | 🔴 responden la pregunta equivocada | renombrar por intención; abrir brownfield |

**El orden no es éste.** ③ antes que ④: el acta es un paquete y cambia lo que Javier ve en cada
parada; las entradas son cirugía de superficie y conviene hacerla cuando ya sabés qué muestra
cada parada. ② puede ir en paralelo — declarar la lista de bucles es escribir, no compilar.

---

## Lo que este documento NO decide

Y queda anotado para que nadie asuma que sí:

1. **Si la corrida T1 pasa o no** — hay dos criterios en conflicto entre `por-tramos.md` §4
   ("mismo `estado.json` con los dos modelos") y `retomar.md` §4 ("pasa si B **cita**"). El único
   campo que se mueve en T1 **es** el veredicto, así que el criterio genérico hace fallar T1
   siempre. Hay que escribir la excepción con nombre.
2. **Qué hace `sf` cuando la máquina no aplica** — un cambio de una línea no necesita brief, PRD
   ni roadmap. Dónde está ese piso, no está escrito.
3. **Las cinco compuertas de `vecinos.md`** — ninguna implementada. Este documento no las
   prioriza; sólo señala que ① de esa lista (decirle al hijo que es hijo) es **más barata de lo
   que ese doc estima**, porque `SPECFORGE_DELEGADO` ya existe y ya viaja
   (`internal/lanzar/correr.go:153`) — lo que falta es que alguna compuerta lo lea.

---

## Índice de los otros documentos

| documento | qué contesta |
|---|---|
| **este** | cómo funciona SpecForge |
| `el-mapa.md` | dónde estamos hoy, tramo por tramo |
| `por-tramos.md` | cómo se prueba cada tramo |
| `el-arnes-ciego.md` | qué falló en la primera corrida real y por qué |
| `vecinos.md` | las compuertas que faltan |
| `arreglos.md` | cerrado — A0–A9 y M1–M3, todo hecho |
