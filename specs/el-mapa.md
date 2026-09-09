# El mapa — dónde estamos, qué se descubrió, y qué falta en cada tramo

**Fecha:** 2026-09-07 · **Branch:** `refundation` · **Commit base:** `6880bd8`

> **Qué es esto y por qué existe.** Hay 792K de specs repartidos en veinte documentos y ningún
> índice. Cada uno es correcto por separado y ninguno dice **dónde estamos**. Este documento no
> agrega diseño nuevo: **consolida**. Si mañana hay que releer uno solo, es éste.
>
> **Reemplaza a `retomar.md`** como punto de entrada. Los demás documentos siguen siendo la fuente
> del detalle — acá está el mapa, no el terreno.

**Cómo leerlo.** §1 son números medidos hoy. §2 son los hallazgos, que es lo que hay que entender
antes que nada. §3 es el tablero tramo por tramo, que es el pedido. §4 son las contradicciones
abiertas entre documentos. §5 las preguntas sin dueño. §6 el orden. §7 el índice de documentos.

**La marca de honestidad:** lo que dice **medido** se verificó leyendo código o corriendo algo, con
la fecha. Lo que dice **deducción** es razonamiento y está marcado como tal. No hay una tercera
categoría.

---

## 1. Dónde estamos, en números

**Medido el 2026-09-07** corriendo `wc`, `go build` y `git`:

```
CÓDIGO       13.623 líneas de Go productivo
             11.078 líneas de tests
                 23 paquetes en sf/internal
                496 tests   (482 con `go test ./...` + 14 con `-tags e2e`)
             compila limpio

HISTORIA        331 commits · 156 de ellos en `refundation` sobre `main`

DOCUMENTOS      792K de specs en ~20 archivos
                468K de skills en 18 carpetas
```

**La lectura, y es deducción:** esto no es un proyecto que se empieza de nuevo. Es un proyecto que
**perdió el índice**. Hay el doble de spec que de código productivo y ningún lugar donde mirar el
conjunto — que es exactamente la sensación de *"se me hizo un quilombo"*.

---

## 2. Los cinco hallazgos que ordenan todo lo demás

Todo lo que sigue en este documento cuelga de estos cinco. Cada uno con su evidencia y su fecha.

### ① El arnés estaba ciego — y de ahí sale la regla del proyecto

**Medido el 2026-09-05.** Primera corrida real de T1 con dos modelos y la misma idea semilla,
palabra por palabra:

```
A (claude-code)   selló  no-lo-hagas    17 retrieved ·  1 model-prior · 13 links
B (nemotron)      selló  hacelo          1 retrieved ·  8 model-prior ·  0 links
```

**Veredictos opuestos, y la máquina aceptó los dos.** A se detuvo; B siguió y construyó un PRD de
74 líneas para un producto que A concluyó que no había que construir.

**La causa no era el modelo.** SpecForge nunca le dio herramientas de búsqueda a ninguno de los dos.
`sfp-scout` manda investigar con Tavily y el MCP de GitHub — o sea, directo al nivel que necesita
llave y configuración. **El nivel que funciona en cualquier arnés no está nombrado en ninguna parte.**

Y la medición que lo cierra: una llamada a `registry.npmjs.org` **sin llave** devolvió tres
competidores reales de la idea de prueba, al instante.

> **LA REGLA QUE ORDENA TODO EL PROYECTO:**
> **Toda regla que vive sólo en el skill es una sugerencia. Sólo la compuerta obliga.**

*Fuente: `el-arnes-ciego.md` §1–§3, §6.*

---

### ② Las compuertas están calibradas para modelo fuerte

`compuerta.Brief` validaba que el veredicto fuera uno de tres strings. **No miraba la evidencia.**
Por eso aceptó el brief de B con cero links.

**Arreglado en `4d7cc28`:** un veredicto sin una sola fuente no sella. Y con un límite dicho de
frente: la compuerta **no puede salir a internet** a verificar que los links existan (cero red y
cero LLM en el camino de enforcement), así que **nunca va a distinguir un link real de uno
inventado — sólo contar.**

**Deducción, y es la tesis del proyecto:** el multi-arnés le pone modelos débiles atrás a compuertas
escritas mirando modelos fuertes. Cada compuerta de las nueve hay que releerla con esa pregunta.

*Fuente: `el-arnes-ciego.md` §4.*

---

### ③ Comprobar ≠ decidir — y v1 los hizo un solo trabajo

**Conversación del 2026-09-07.** Son dos trabajos distintos:

```
comprobar   trabajo mecánico   →  la máquina, SIEMPRE, sin pedir permiso
decidir     juicio             →  Javier, y SÓLO donde equivocarse sale caro
```

**El error de v1 no fue "el humano es la compuerta". Fue que el humano era la compuerta *y* el
inspector** — tenía que leer, contar, verificar *y además* decidir, veinte veces por feature. La
cuarta se convierte en apretar Enter sin mirar.

**El *dónde* ya está bien resuelto** (medido en `done.go` y `paradas.go` el 2026-09-07): paran ⑥, ⑧,
⑨/⑩, ⑰ y ㉓; no paran ⑦, los lotes ni ㉑㉒ — la revisión **decide sola**, cuenta hallazgos abiertos.

**Lo que falta es el *qué mostrar* cuando para.** Y hay un hallazgo concreto en el código: arriba de
`Resultado.Texto()` está escrito *"Los ✓ no se listan: sólo importa lo que falta"*. **Tiene razón
para el subagente que arregla y está equivocado para Javier que firma.** Un output, dos lectores
opuestos. Hoy `Resultado` guarda `Fallas` y `Avisos`, y **los ✓ se tiran** — así que el informe para
Javier no se puede armar aunque se quiera.

*Fuente: `el-arnes-ciego.md` §5②.*

---

### ④ ME TRABÉ no cubre el bucle de revisión

**Medido el 2026-09-03** leyendo `done.go`:

- la vuelta de revisión **no incrementa** `intentos_fallidos` (`cerrarRevision`)
- cerrar el lote de corrección **lo resetea a cero** (`done.go:259`)

> **O sea que el corte de tres intentos nunca se dispara en el único bucle que puede girar solo.**

El único freno es un aviso en `vuelta >= 3`, que **no frena** — y ese número **lo escribe el propio
revisor**, no `sf`. La corrida real de Javier llegó a **7 vueltas**, y los artefactos no se
conservaron, así que el mecanismo exacto **no está verificado**.

**Éste es el agujero peor del proyecto**, y sigue abierto.

*Fuente: `por-tramos.md` §5 T6.*

---

### ⑤ `sf` no sabe si quien le habla es una persona o un hijo suyo

**Medido el 2026-09-04:** `script(1)` fabrica un TTY (`/dev/pts/7`), así que `hayPersona` no alcanza
para distinguir. Un subagente headless puede parecer una persona.

**Novedad medida hoy (2026-09-07), y abarata el arreglo:** la marca ya viaja. `correr.go:153` setea
`SPECFORGE_DELEGADO=<id>` en el entorno del hijo, y `registro.go:376` la lee. Pero el comentario del
propio código lo dice: **"HOY SÓLO REGISTRA"**. Ninguna compuerta la mira.

**Deducción:** `vecinos.md` §7 estima el ① como *"la pieza ya existe, es enchufarla"*. Con la marca
ya puesta por T0, es más barato todavía de lo que ese documento supone.

*Fuente: `vecinos.md` §2, `registro.md` §5.*

---

## 3. El tablero — tramo por tramo

**La leyenda:**

```
✅  hecho y verificado          🟡  corrido y falló, con arreglos identificados
⬜  no corrido todavía          🔴  agujero conocido y abierto
```

---

### T0 · El tablero — el registro ✅

| | |
|---|---|
| **Qué es** | `internal/registro` + `sf log`. Sin esto los demás tramos producen impresiones, no datos |
| **Estado** | **HECHO** — `e47f847`. Paquete nuevo, 2 archivos, stdlib sola, cero dependencias |
| **Qué dejó** | 4 enganches (`main`, `next`, `done`/`parada`, `lanzar`) · `main()` → `despachar() int` |

**Qué se descubrió al construirlo.** Que `quien` no se puede deducir del TTY (hallazgo ⑤), y que la
única señal confiable es la que pone `sf` mismo al lanzar.

**Qué NO trae, a propósito, con su umbral escrito:** cadena de hash, versionado, rotación, y —la
importante— **que una compuerta lo lea**. Ese umbral son los datos de T5 y T6: *el conteo es para
descubrir el número de corte, y el número sale de mirar las corridas*. Cablearlo antes es inventarlo.

**Qué falta:** nada para T0. Está cerrado.

---

### T1 · El brief — ①–⑥ 🟡 **corrido, falló**

> ⚠️ **Esta ficha quedó vieja el 2026-09-08.** T1 se replanteó entero después de leer las
> skills de Matt Pocock: la entrevista pasa a ser un árbol con frontera, la investigación deja
> de ser una etapa y pasa a ser una rama, y el tramo deja tres artefactos contables en vez de
> uno. **La fuente de T1 es ahora [`tramo-1.md`](tramo-1.md)** — que además cierra la
> contradicción §4① (la vara) y la pregunta §5① (el bucle sin dueño).

| | |
|---|---|
| **Qué corre** | `sfp-scout`, vía `vos` (está en `conversan`, `maquina.go:914`) |
| **Produce** | `.docs/brief.md` con `veredicto` en el frontmatter |
| **Compuerta** | `compuerta.Brief` — existe, veredicto válido, **y al menos una fuente** (desde `4d7cc28`) |
| **Se mueve** | `producto.brief_sellado`: `""` → `hacelo` \| `pivotea` \| `no-lo-hagas` |
| **Estado** | **CORRIDO EL 2026-09-05 Y FALLÓ.** Es el hallazgo ① |

**Qué se descubrió.** Todo el §2① y §2②: el arnés ciego, la compuerta que no miraba evidencia, y la
regla del proyecto. **La corrida sirvió** — encontró la causa raíz, que es para lo que se corre.

**Qué falta, en orden:**

| | Qué | Toca | Tamaño |
|---|---|---|---|
| **1** | **el nivel 0 entra en el skill** — `method.md` §1 arranca por registries/GitHub API/`curl` sin llave, y recién después Tavily. Más la regla: *"si el nivel 0 ya contesta, no hace falta el nivel 1"*. Y la honestidad: el nivel 0 mapea el panorama, **no** trae señales de demanda | `sfp-scout/references/` | **chico, sin Go** |
| **2** | **`evidencia.json`** — el brief deja el conteo como dato, y la compuerta deja de tener un `grep` en el camino principal | Go + skill | mediano |
| **3** | **`sf doctor` ve las herramientas** — bloque `herramientas` al lado de `skills`; avisa, no frena; y si el arnés no expone la lista dice *"no lo puedo saber acá"*, que **no** es lo mismo que *"no están"* | Go | mediano |
| **4** | **el acta en el ⑥** — la parada imprime tres bloques: `COMPROBÉ` / `MEDÍ` / `NO PUEDO COMPROBAR` | Go | chico si el 2 está |

**Diferido con su umbral:** partir el ①–⑤ en pasos (el corte del ②). **No entra en esta vuelta**
porque abre un bucle nuevo sin dueño — ver §5.

**Cómo se sabe que pasó:**

```
B cita fuentes           →  era la herramienta. Causa raíz cerrada.
B sigue sin citar nada   →  era el modelo. Otro problema, y recién ahí se sabe.
```

⚠️ **Ojo con la vara — hay una contradicción sin resolver acá.** Ver §4①.

---

### T2 · El PRD y la constitución — ⑦⑧ ⬜

| | |
|---|---|
| **⑦** | `sfp-po` / **subagente** → `.docs/prd.md`. `sf` anota `producto.prd_hash` y **no para** |
| **⑧** | `sfp-constitucion` / **vos** → `.docs/constitucion.md`. Es el único paso además del brief que te habla |
| **Compuertas** | `compuerta.PRD` (sólo `os.Stat`) · `compuerta.Constitucion` (`test_cmd` no vacío, entre otras) |
| **Estado** | **NO CORRIDO.** B llegó hasta acá en la corrida del 05-09, pero sobre un brief inválido |

**Qué ya se sabe, sin haberlo corrido:**

- **El bug del ⑧ vivía acá.** Estaba del lado equivocado de `conversan`. Arreglado en `4acf47f`.
- **El encadenado sorprende.** Javier selló el brief y le salieron dos pasos de una: *"es medio
  enquilombado saber cuándo no debés aprobar porque si no salta directo a la siguiente fase"*. El
  commit `d1c7856` lo **anuncia** — y anunciar no es resolver. **Esta corrida decide si alcanza.**
- 🔴 **`compuerta.PRD` es la más floja de las nueve.** Es un `os.Stat`: un PRD de una línea la pasa.
  Y su comentario dice *"y tenga cuerpo"*, **que es falso**.

**Qué falta:** nada de código antes de correrlo. **Arreglar el comentario mentiroso de
`compuerta.PRD` es gratis y se puede hacer ya.** Apretar o aflojar la compuerta, no: sin un caso
real es adivinar.

**Qué mirar en la corrida:** que el `test_cmd` que detectó `sf init` **sobreviva** a que el skill
reescriba el archivo. Y si un PRD flojo pasa el ⑦ y el problema recién aparece en el ⑨.

---

### T3 · El backlog y el roadmap — ⑨⑩ ⬜

| | |
|---|---|
| **⑨** | `sfp-backlog` → `.docs/backlog/us-#.md` con `tipo: us\|bug` y criterios `CA-#` · ⏸ → `producto.backlog_visto` |
| **⑩** | `sfp-roadmap` → `.docs/roadmap.json` · **no deja sello** (es deducible, R6) |
| **Compuertas** | `compuerta.Backlog` (los `CA-#` con texto) · `compuerta.Roadmap` (ninguna historia huérfana) |
| **Estado** | **NO CORRIDO** |

**Qué ya se sabe:**

- La ⏸ del ⑨ es **la única parada que existe sólo para mirar**. No sella nada, y necesitó un campo
  inventado (`backlog_visto`) que no estaba en el diseño. **¿Sirve, o es una parada de más?**
- El ⑨ **corre dos veces**: parte el PRD, y vuelve cada vez que entra algo por `sf new`. Hay que ver
  las dos.
- **Riesgo ya documentado:** el ⑩ corriendo a mitad de ciclo suma un `intentos_fallidos` de más
  (`arreglos.md`). **Si aparece en la corrida, es esto y no algo nuevo.**

---

### T4 · La planificación — ⑪–⑰ ⬜

| | |
|---|---|
| **Entra por** | `sf take <f-#>` — el ⑪, decisión de Javier |
| **Corre** | `sf-plan` en **un subagente de contexto amplio**, cinco pasos de un saque |
| **Produce** | `decision.md` (⑫) · `spec-design.md` (⑬) · `tareas.json` (⑭–⑯) |
| **Compuerta** | `compuerta.Planificacion` — 5 chequeos, incluido *tres opciones exactas* y *cada tarea apunta a un criterio* |
| **Termina en** | 🛑 el ⑰ — **la única parada sin horizonte**, y es deliberado |
| **Estado** | **NO CORRIDO** |

**Qué ya se sabe:** los checkpoints **no tienen campo** — se deducen mirando qué archivos existen.

**Preguntas de la corrida:** ¿el subagente aguanta los cinco pasos sin degradarse? ¿El ⑯ elige bien
el modelo por lote o pone el mismo en todos? ¿Se usan las **tres** puertas del ⑰ —pido cambios, otra
feature, implementar— o una sola?

---

### T5 · Implementar — ⑱–⑳ ⬜

| | |
|---|---|
| **El ciclo** | `sf lote start` (ve el rojo, guarda `hash_tests`) → trabajo → `sf done --msg` (ve el verde, compara el hash, commitea) |
| **Compuertas** | tres: la branch antes · **el rojo en el medio** · el verde y el hash después |
| **El bucle** | **N lotes** — el primer bucle real de la máquina |
| **El corte** | `TopeIntentos = 3` → ME TRABÉ. Vive acá y **sólo** acá |
| **Estado** | **NO CORRIDO** |

**Preguntas de la corrida — son las dos joyas de la máquina, sin ejercitar:** ¿el rojo obligatorio se
puede cumplir sin fabricar un test falso? ¿El `hash_tests` atrapa a un modelo que afloja un assert
para llegar al verde? Y **¿tres intentos es el número, o se descubre otro?** Hoy es el único número
mágico del paquete.

🔴 **Agujero conocido que se cruza acá:** `vecinos.md` §3 — *el verde vacío*. Un `test_cmd` que no
matchea ningún test devuelve exit 0, y **`sf` lo lee como verde**. Un test que no corrió no es un
test que pasó.

---

### T6 · Revisión y cierre — ㉑–㉓ ⬜ 🔴 **el que más importa**

| | |
|---|---|
| **㉑㉒** | `sf-check` / subagente grande → `revision.json`: criterios, mutantes, hallazgos |
| **Compuerta** | `compuerta.Revision` — todos los criterios con veredicto · **cero hallazgos abiertos** · avisa (no frena) en vuelta ≥3 |
| **El bucle** | un hallazgo abierto → vuelve a `implementar` y abre un lote de corrección |
| **㉓** | `sf-cierre` → `doc.md` + `journal.md` → ⏸ → `sf approve` archiva, mergea y borra la branch |
| **Estado** | **NO CORRIDO** — y es donde está el hallazgo ④ |

**Las tres preguntas abiertas, y son del proyecto, no del tramo:**

1. Un revisor LLM al que se le pregunta *"¿qué está mal?"* **siempre** contesta algo. **¿Dónde está
   el punto fijo?**
2. `revision.json` permite hallazgos con `origen: 22` **sin criterio asociado** — opiniones de
   calidad sueltas, de las que hay infinitas. ¿Se limita a lo que cuelga de un criterio o de un
   mutante sobreviviente?
3. El lote de corrección **exige que la suite falle**. Un hallazgo del tipo *"el diseño está flojo"*
   no lo puede poner en rojo nadie. **¿Qué hace el implementador ahí — se traba, o fabrica un test?**

**Dato duro:** la corrida real de Javier llegó a **7 vueltas**.

---

### T7 · El ticket — no es un tramo, es otra puerta ⬜

**Lo que hay hoy, medido** (`maquina/entradas.go:48`):

```go
if !e.Producto.ConstitucionSellada {
    ef.falla("este producto todavía no está armado: seguí con `sf next` desde el brief")
}
```

> **No se puede meter un ticket en un proyecto que no pasó por brief → PRD → constitución.**

El camino corto existe —un `tipo: bug` saltea planificación y revisión— **pero sólo dentro de un
producto con la ceremonia entera hecha**. Llega un bug a un proyecto que ya existe, y `sf` pide que
primero se le invente un brief.

**Eso es, en una línea de Go, la queja de que la máquina *"sólo sirve cuando hay que hacer un
desarrollo"*.**

**Qué falta contestar, y no ahora:** ¿un ticket necesita las tres cosas o sólo la constitución (que
es la que tiene el `test_cmd`, la branch base y las convenciones)? ¿Cuál es el mínimo para que un
proyecto existente entre? — **la pregunta brownfield, sin contestar en todo el repo.**

> **No se toca `entradas.go` hasta que T1 y T2 hayan corrido.** La restricción puede ser correcta y
> estar sólo mal explicada.

---

## 4. Las contradicciones abiertas — dos documentos que dicen distinto

Éstas hay que resolverlas **antes** de escribir código, porque cada una cambia lo que se escribe.

### ① La vara de T1 — `por-tramos.md` §4 vs `retomar.md` §4

```
por-tramos §4   "PASA si el estado.json quedó igual con los dos modelos"
retomar   §4    "Ojo con la vara: pasa si B CITA, no si B dice hacelo.
                 El veredicto sigue siendo juicio."
```

**Se contradicen, y sólo en T1.** El único campo que se mueve en T1 es `producto.brief_sellado`, que
**es** el veredicto: juicio puro. Con la vara de §4, T1 **falla siempre** — dos modelos honestos
pueden leer la misma evidencia y decidir distinto. En T2–T6 los campos son mecánicos (`prd_hash`,
`constitucion_sellada`, el estado de la feature) y ahí el criterio genérico funciona bien.

**Propuesta (sin confirmar):** para T1 la vara es de **evidencia**, no de veredicto, y se escribe en
`por-tramos.md` §4 como excepción con nombre. Si no, la próxima corrida se declara fallida por la
razón equivocada y se pierde el dato.

### ② `Parte` choca con `sobre.Parte`

`el-arnes-ciego.md` §5② bautizó **`el parte`** al informe de la parada. Pero `Parte` ya existe:
`sobre.go:77`, *"Parte es una sección del sobre"*. **Y choca dentro de la misma familia conceptual.**

**Propuesta (sin confirmar):** **`el acta`** — verificado libre en todo el código, y más preciso: un
acta se levanta, se lee y **se firma**, que es lo que hace `sf approve`.

```
el sobre   →  va al que TRABAJA   ·  lo sirve `sf context`
el acta    →  va al que FIRMA     ·  la imprime la parada
```

### ③ El comentario de `compuerta.PRD` miente

Dice *"y tenga cuerpo"*. **Es un `os.Stat` y nada más.** No es una decisión de diseño pendiente: es
un comentario falso, y arreglarlo es gratis.

---

## 5. Las preguntas sin dueño

No están contestadas, ninguna bloquea hoy, y las cinco pueden morder mañana.

| | La pregunta | Dónde muerde |
|---|---|---|
| **①** | Si se parte el ①–⑤ y el grill descubre que falta investigar, **¿el ② se reabre? ¿Quién cierra ese bucle?** | T1. **Es la razón por la que el corte no entra en esta vuelta** |
| **②** | **¿Dónde está el punto fijo de un revisor LLM?** Siempre encuentra algo | T6 — el agujero ④ |
| **③** | **¿Cómo se mide que el orquestador delega?** El test del pid no sirve: mide arquitectura del arnés, no delegación | transversal. Candidata: `SPECFORGE_DELEGADO`, que **ya existe pero sólo registra** |
| **④** | **¿Un ticket necesita las tres cosas, o sólo la constitución?** | T7 / brownfield |
| **⑤** | **¿La ⏸ del ⑨ sirve, o es una parada de más?** | T3 |
| **⑥** | **La ventana se llena en el ①–⑤ y no hay regla de corte.** Matt la tiene (`smart zone`); nosotros no, y este replanteo alargó ese tramo | transversal — el detalle en [`tramo-1.md`](tramo-1.md) §10① |
| **⑦** | **Matt da cinco opciones en el borde de fase; nosotros una y clavada en Go.** No es defecto, es ser máquina — pero cuando la clavada no sirve no hay salida | transversal — [`tramo-1.md`](tramo-1.md) §10② |
| **⑧** | **La memoria del arnés contamina el banco.** `icm recall` devolvía la respuesta de la corrida anterior, y el `CLAUDE.md` global manda usarla. Hoy se tapó cambiando la semilla; falta la pregunta de fondo: ¿cómo se corre en un cuarto limpio? | transversal — [`tramo-1.md`](tramo-1.md) §10③ |

---

## 6. El orden — y por qué no es opcional

```
T0  el tablero          ✅ HECHO
T1  el brief            🟡 corrido y falló → 4 arreglos identificados
T2  el PRD y la constitución
     └─ T7 el ticket    se HABILITA acá; se corre cuando haya ganas
T3  el backlog y el roadmap
T4  la planificación
T5  implementar
T6  revisión y cierre   ← el que más importa, y el que tiene el agujero peor
```

> **El orden está forzado por los datos, no por el gusto.** No se puede probar el ⑨ con un PRD
> escrito **a mano**, porque volvés a los artefactos a mano — que es justo lo que tapó los bugs. Para
> que la entrada del ⑨ sea de verdad, tiene que haber salido de la corrida del ⑦.

**Y las cinco compuertas de `vecinos.md` NO son previas a nada.** Son enforcement puro, stdlib, y
no bloquean ninguna corrida. Su orden propio: ⓪ decirle al hijo qué es → ⑤ saltear no es aprobar →
① la firma del ⑰ → ③ el número de fila → ② el verde vacío → ④ la aprobación que vence (ésta,
**después** de la corrida real). **Ninguna está implementada.** El ② se cruza con T5 y el ① se
abarató con T0 (§2⑤).

### Lo inmediato, concreto

```
1  arreglar el comentario de compuerta.PRD          gratis, ahora
2  resolver §4① (la vara de T1) y §4② (el nombre)   una decisión, no código
3  el nivel 0 en sfp-scout                          una tarde, sin Go
4  evidencia.json                                   desbloquea el 5 casi gratis
5  el acta en el ⑥
6  sf doctor ve las herramientas
7  RE-CORRER T1  ← acá se cierra o se reabre la causa raíz
```

---

## 7. Lo que se decide NO hacer, con su umbral

La disciplina es de `specd`: **cada cosa diferida tiene un umbral escrito.** No se hace, y se sabe
qué la despertaría.

| No se hace | Por qué | Qué lo despierta |
|---|---|---|
| apretar más `compuerta.Brief` | ya atrapó el caso real; exigir más es castigar al modelo por no darle herramientas | una corrida **con herramientas** donde un brief cite una fuente sola y pase |
| tocar `compuerta.PRD` (más allá del comentario) | esta corrida no lo explotó; sin caso real es adivinar | un backlog malo cuya causa se rastree a un PRD flojo |
| construir `sf buscar` | superficie nueva grande sin un dato que la pida | una corrida con herramientas donde el brief cite una URL que no existe |
| partir el ①–⑤ ahora | abre un bucle sin dueño (§5①) | contestar §5① primero |
| el mapa entero de wayfinder | toca la forma de la máquina | un tramo T3+ donde el problema sea *"no sabemos lo suficiente para planificar"* |
| que una compuerta lea el registro | el conteo es para **descubrir** el número de corte | los datos de T5 y T6 |
| tocar `entradas.go` (brownfield) | la restricción puede ser correcta y estar mal explicada | T1 y T2 corridos enteros |
| `--async` / dos `sf` en paralelo | fuera de alcance | — |
| **empezar de cero al estilo Matt Pocock** | sus skills son markdown sin runtime: **es "reglas que viven sólo en el skill"**, y el hallazgo ① dice que eso no obliga | nada. Lo que **sí** se le copia: skills chicas, user-invoked vs model-invoked, y *"nada se actualiza a tus espaldas"* |

---

## 8. El índice de documentos — qué hay en cada uno

Para no releer 792K. **Ordenados por cuánto valen hoy.**

| Documento | Qué contiene | Sigue vivo |
|---|---|---|
| **`el-mapa.md`** | **este** — el estado, los hallazgos, los tramos | ✅ punto de entrada |
| **`tramo-1.md`** | **T1 replanteado: los 4 primitivos, los artefactos, la vara** | ✅ **la fuente de T1** |
| `el-arnes-ciego.md` | el parte de la corrida A/B + las 4 tareas + el orden | ✅ el diagnóstico que lo originó |
| `por-tramos.md` | el método de prueba por tramos y las fichas T1–T7 | ✅ la fuente del plan |
| `vecinos.md` | las 5 compuertas que faltan, con contrato y tests | ✅ ninguna implementada |
| `registro.md` | T0, campo por campo | ✅ construido; útil de referencia |
| `maquina-estados.md` | la máquina: estados, transiciones, quién entra y sale | ✅ referencia |
| `superficie-sf.md` | los comandos de `sf` y por qué cada uno (H1–H16) | ✅ referencia |
| `headless.md` | el contrato de `sf lanzar` | ✅ referencia |
| `arreglos.md` | A0–A9 + M1–M3 | ⚠️ **cerrado — todo hecho.** Histórico |
| `retomar.md` | el traspaso del 05-09 | ⚠️ **lo reemplaza este documento** |
| `artefactos.md`, `flujo-real.md`, `que-sobrevive.md`, `agnostico-al-harness.md`, `session*.md`, `skills.md`, `construccion.md`, `install-interactivo.md`, `salir-a-la-cancha.md`, `anexo-determinismo.md` | diseño anterior a las corridas | ⚠️ **sin auditar contra el código.** Ver abajo |

> **La deuda que este índice deja al descubierto, y es honesta:** los diez de la última fila suman
> ~400K y **nadie verificó si siguen siendo ciertos** después de A0–A9, T0 y `4d7cc28`. No es urgente
> —ninguno bloquea una corrida— pero es exactamente el terreno donde nace la sensación de quilombo.
> **Umbral que lo despertaría:** la primera vez que uno de ellos contradiga al código en una
> decisión real. Ya pasó una vez hoy, con `Parte`.
