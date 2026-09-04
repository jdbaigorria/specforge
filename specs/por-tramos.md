# Probar por tramos — el plan de secciones

**Fecha:** 2026-09-04 · **Branch:** `refundation` · **Commit:** `26aea5d`

> **Estado: PROPUESTA.** Sale del debate del 2026-09-03/04, de un diagnóstico de Javier y de cuatro
> agujeros medidos ese día. No propone reconstruir nada: propone **instrumentar y ejercitar** lo que
> ya está construido, tramo por tramo, y convertir cada sorpresa en un hallazgo escrito.

Este documento continúa a [`salir-a-la-cancha.md`](salir-a-la-cancha.md) ③ —que ya pedía la corrida
real— y a [`vecinos.md`](vecinos.md), que listó cinco compuertas esperando datos de uso. Lo que
agrega es **el método**: cómo se corre esa corrida para que produzca hechos en vez de impresiones.

---

## 1. El diagnóstico, y por qué no es una opinión

Javier, 2026-09-04:

> *"Cometimos el error de crear toda la maquinaria completa en lugar de ir probando cada parte
> individual. Cuando ejecutamos el ciclo completo salieron varios errores por no haberlos probado
> desde un principio, por ejemplo el problema que salió de la constitución."*

Está respaldado por este mismo repo. `salir-a-la-cancha.md`, textual:

> *"Lo auditado fue **el binario**. Lo que nunca se probó es que **un agente escriba lo que la
> compuerta acepta**. Los tres tests punta a punta escriben cada artefacto **a mano**."*

**456 tests verdes, y ninguno donde el artefacto lo haya producido un modelo.**

### El bug del ⑧ es el caso de manual

Está explicado en `maquina/maquina.go`, en el comentario de `conversan`:

> *"Su skill siempre dijo 'Step 1: Interview Javier — this state talks to him directly'… Lo que lo
> puso ahí fue el default de esta función. **Los modelos de Anthropic tapaban la contradicción**
> salteándose el Step 1 y escribiendo directo; **nemotron hizo lo que estaba escrito** y preguntó al
> aire."*

Leído despacio, dice tres cosas y las tres importan:

1. El defecto **no estaba en el código**: el binario hacía exactamente lo que decía hacer.
2. Estaba **entre el código y el que lo lee** — el skill pedía una cosa y la máquina asumía otra.
3. Y era **invisible con un modelo y visible con otro**.

> **Ningún test de Go podía encontrarlo.** No porque falte un test: porque el defecto no vive en el
> lugar donde los tests miran.

### Los cuatro agujeros del 2026-09-03 son de la misma familia

| Agujero | Dónde vive |
|---|---|
| el bucle de revisión no tiene corte | `done.go:cerrarRevision` no incrementa `intentos_fallidos`, y `cerrarLote` lo resetea |
| la vuelta la declara el propio revisor | `Vuelta` no lo incrementa ningún Go; lo escribe `sf-check` |
| el orquestador no delega | `via: subagente` es un string que sf imprime y nadie comprueba |
| el sello no se puede probar | medido: el agente fabrica un TTY con `script(1)` |

**Ninguno es un bug del binario. Los cuatro están en la costura.** Ese es el terreno que este plan
viene a cubrir.

### La distinción que ordena el documento entero

```
BUG DEL BINARIO      el código no hace lo que dice        →  lo caza un test de Go
BUG DE LA COSTURA    el código hace lo que dice, y lo     →  sólo lo caza una corrida
                     que dice no es lo que hacía falta
```

Los 456 tests cubren la primera columna, y la cubren bien. **La segunda está vacía.**

---

## 2. La regla de corte — los tramos ya están dibujados

La tentación es inventar secciones. **No hace falta: la máquina ya las tiene.**

> **Un tramo es lo que va entre dos paradas.**

Sale de los dos mapas que ya viven en `maquina.go` —`ordenDeEstados` y `paranAlFinal`— y es
exactamente lo que `horizonte()` recorre para anunciar *"si aprobás corre…"*. No es una estructura
nueva: es la que ya existe, usada para otra cosa.

Y tiene la propiedad que hace falta acá: **un tramo arranca con Javier decidiendo y termina con
Javier decidiendo.** O sea que es la unidad natural de una corrida — se puede empezar, terminar, y
mirar qué quedó, sin cortar nada por la mitad.

```
T1  ①–⑥    el brief                        →  🛑 el sello del ⑥
T2  ⑦⑧      el PRD y la constitución        →  🛑 el sello del ⑧
T3  ⑨⑩      el backlog y el roadmap         →  ⏸ y sigue
T4  ⑪–⑰    la planificación                →  🛑 el ⑰
T5  ⑱–⑳    implementar (N lotes)           →  el primer bucle
T6  ㉑–㉓    revisión y cierre               →  el segundo bucle · ⏸
T7  —        el ticket (entrada B y C)       →  no es un tramo, es otra puerta
```

`headless.md:294` ya usaba esta misma unidad, con otro propósito:

> *"Lo que se automatiza es **el tramo entre dos paradas**."*

---

## 3. El tablero va primero — y esto es la única corrección al plan de Javier

### El problema

Un test de tramo **no es un test de Go**: es una corrida con un modelo, y por lo tanto **no es
determinista**. Correr T1 una vez y que salga bien prueba muy poco — el bug del ⑧ *salía bien* con
Opus durante meses.

Entonces hay una pregunta que el plan tiene que contestar antes de empezar:

> **¿Qué significa "pasó" en un test de costura?**

No puede ser *"salió bien"*. Tiene que ser:

> **el estado quedó donde debía, y el camino fue el que debía.**

Y lo segundo **hoy no se puede afirmar**, porque no queda registrado en ningún lado.

### La analogía

Querés encontrar dónde el motor pierde fuerza, tramo por tramo. **Pero el auto no tiene tablero.**
Vas a terminar cada corrida diciendo *"me pareció que en la subida aflojaba"*. Primero el tablero,
después la ruta.

### Lo que hay hoy, y lo que falta

Lo único estructurado que existe es `.specforge/lanzamientos/` (`lanzar/correr.go:21`): una ficha
JSON por corrida de `sf lanzar` —modelo, esfuerzo, duración, exit code, costo, tokens, si cargó la
skill, qué herramientas usó— más el stream crudo en JSONL. **Está bien hecho**, incluida la regla
*"lo que el arnés no da, NO ESTÁ"*.

Pero cubre **sólo las corridas que lanza `sf lanzar`**. No queda registro de:

| | hoy |
|---|---|
| cada `sf next`: qué se preguntó, qué se contestó | nada |
| cada `sf done`: qué compuerta falló y por qué | nada |
| cada `sf approve` / `reject` / `dismiss`, y desde dónde | nada |
| las transiciones de estado, y cuántas veces se repitió cada una | nada |
| el trabajo que el orquestador hizo **sin** `sf lanzar` | nada |

> `estado.json` guarda **la foto**. No existe **la película**. Y las cuatro preguntas de Javier
> —dónde va un loop, cuándo corta, cuándo es interactivo, qué deja cada paso— **son preguntas sobre
> la película.**

### El mínimo, y qué NO incluye

**Un archivo, una línea por invocación de `sf`.** `.specforge/registro.jsonl`, append-only:

```json
{"t":"2026-09-04T10:31:02Z","cmd":"next","salida":0,
 "estado_antes":"revision","estado_despues":"revision","feature":"f-2",
 "tipo":"trabajar","skill":"sf-check","via":"subagente","modelo":"opus-grande",
 "quien":"orquestador","pid":12345,"ppid":12340,
 "fallas":[],"avisos":[]}
```

Dos campos merecen aclaración porque parecen lo mismo y no lo son:

- **`via`** — lo que `sf` le **dijo** al orquestador: `vos` · `subagente` · `consola`.
- **`quien`** — desde dónde se **corrió** `sf`: `terminal` · `orquestador` · `delegado`.

Y con eso, las preguntas se vuelven conteos:

```
¿cuántas vueltas dio la revisión de f-2?
  → contar las líneas con estado_antes=implementar, estado_despues=revision

¿se lanzó el subagente que sf pidió?
  → entre un `next` con via=subagente y el `done` siguiente, ¿hay una ficha
    de lanzamiento? ¿o el `context` vino del mismo pid?

¿algún sello salió de una terminal?
  → contar `approve` por valor de `quien`
```

**Lo que este mínimo NO trae, a propósito:**

- **No trae cadena de hash.** Un registro que nadie puede editar en silencio hace falta el día que
  **una compuerta lea el registro**. Mientras sea un instrumento de medición, nadie falsifica sus
  propios datos de prueba. Se agrega cuando se lo necesite, no antes.

> **T0 está especificado entero en [`registro.md`](registro.md)**: la forma de la línea, los cuatro
> puntos de enganche, los seis tests y los cinco diferidos con su umbral.
- **No frena nada.** En esta etapa el registro **sólo escribe**. Ninguna compuerta lo lee todavía.
- **No agrega superficie**, salvo un comando para leerlo: `sf log`. Es la única adición al
  inventario de `comandos.Todos`, y existe porque un JSONL a ojo no se lee.

> **Y esto no es una desviación del plan: es su primera pieza.** El mismo registro que sirve de
> tablero es el que después cierra tres de los cuatro agujeros.

---

## 4. Las cinco preguntas de cada tramo

Las cuatro primeras son de Javier, textuales. La quinta sale del bug del ⑧.

```
①  ¿esto es un bucle?  ¿de qué, y quién lo cierra?
②  ¿cuándo tiene que cortar, y con qué número?
③  ¿cuándo tiene que ser interactivo, y por qué no puede delegarse?
④  ¿qué artefacto deja, y qué tiene que quedar en el registro?
⑤  ¿qué pasa con un modelo que obedece LITERAL en vez de rellenar huecos?
```

### La quinta merece su párrafo

El bug del ⑧ existió porque un modelo *"rellenó el hueco"* —se salteó un paso que no tenía sentido y
siguió— y otro *"hizo lo escrito"* —preguntó al aire, sin nadie del otro lado—. **El primero tapa el
defecto; el segundo lo revela.**

> **Cada tramo se corre con dos modelos de obediencia distinta.** Uno que rellena, uno que obedece
> literal.

Y de ahí sale un criterio de "pasó" que es chequeable y no una impresión:

> **PASA si el `estado.json` quedó igual con los dos modelos.**

Distinto estado con el mismo tramo significa que **el skill y la máquina no dicen lo mismo** — que
es, exactamente, la familia de bug que ya mordió una vez.

---

## 5. Los tramos, uno por uno

Cada ficha dice: qué corre, qué produce, **qué campo del `estado.json` tiene que moverse**, qué
compuerta cierra, y las preguntas abiertas. Los campos salen de `estado/estado.go`.

### T1 · El brief — ①–⑥

| | |
|---|---|
| **estado / skill** | `brief` / `sfp-scout` |
| **vía** | `vos` — está en `conversan` (`maquina.go:914`) |
| **produce** | `.docs/brief.md`, con `veredicto` en el frontmatter |
| **compuerta** | `compuerta.Brief` — que exista y traiga uno de los tres veredictos |
| **se mueve** | `producto.brief_sellado`: `""` → `hacelo` \| `pivotea` \| `no-lo-hagas` |
| **termina en** | 🛑 ⑥, y anuncia horizonte: *"si aprobás corre ⑦ el PRD → ⑧ la constitución"* |

**Preguntas abiertas.** ①–⑤ es **el único pinponeo declarado de la máquina**, así que es el mejor
lugar para contestar la ③. ¿Cuántas vueltas hasta que el brief está? ¿Quién decide que terminó — el
skill o Javier? ¿Y qué pasa si el veredicto es `no-lo-hagas`: la máquina lo sella igual (hoy sí) y
después qué?

**Riesgo ya conocido:** `sfp-scout` trababa el ⑥ escribiendo el veredicto vacío. Ya arreglado, y se
encontró **leyendo**, no corriendo. Si reaparece, es esto.

**Advertencia honesta:** T1 tiene **la menor maquinaria de los seis** — sin lotes, sin hallazgos,
sin vueltas. Va a contestar mucho sobre interactividad y casi nada sobre bucles. Es igual el primero,
por lo que dice §7.

---

### T2 · El PRD y la constitución — ⑦⑧

| | |
|---|---|
| **⑦** | `prd` / `sfp-po` / **subagente** → `.docs/prd.md` · sf anota `producto.prd_hash` y **no para** |
| **⑧** | `constitucion` / `sfp-constitucion` / **vos** → `.docs/constitucion.md` |
| **compuertas** | `compuerta.PRD` —**sólo comprueba que el archivo exista**— · `compuerta.Constitucion` (`test_cmd` no vacío, entre otras) |
| **se mueve** | `producto.prd_hash`, después `producto.constitucion_sellada` |
| **termina en** | 🛑 el sello del ⑧ |

**Este es el tramo donde vivía el bug**, y por dos motivos distintos que conviene no confundir:

1. **El ⑧ estaba del lado equivocado** de `conversan`. Arreglado en `4acf47f`.
2. **El encadenado sorprende.** Javier selló el brief y le salieron dos pasos de una: *"es medio
   enquilombado saber cuándo no debés aprobar porque sino salta directo a la siguiente fase"*. El
   commit `d1c7856` lo anuncia — **y anunciarlo no es lo mismo que resolverlo.** Esta corrida tiene
   que decidir si alcanza.

**Qué mirar además:** que el `test_cmd` que detectó `sf init` **sobreviva** a que el skill reescriba
el archivo (ya estaba en la lista de `salir-a-la-cancha.md`).

**Y una asimetría que conviene mirar de frente.** `compuerta.PRD` es la compuerta más floja de las
nueve: comprueba que `.docs/prd.md` exista, y nada más. Un PRD de una línea la pasa. Puede estar
bien —el que de verdad cuenta criterios es el ⑨, que se alimenta de él— o puede ser el motivo por el
que un backlog malo se descubre recién un tramo después. **Esta corrida es la que lo dice.**

---

### T3 · El backlog y el roadmap — ⑨⑩

| | |
|---|---|
| **⑨** | `backlog` / `sfp-backlog` → `.docs/backlog/us-#.md`, con `tipo: us\|bug` y criterios `CA-#` |
| | ⏸ parada barata → `producto.backlog_visto` |
| **⑩** | `roadmap` / `sfp-roadmap` → `.docs/roadmap.json` · **no deja sello** (es deducible, R6) |
| **compuertas** | `compuerta.Backlog` (los `CA-#` con texto) · `compuerta.Roadmap` (ninguna historia huérfana) |

**Preguntas abiertas.** La ⏸ del ⑨ es la **única parada que existe sólo para mirar** — no sella
nada, y por eso necesitó un campo inventado (`backlog_visto`, que no estaba en el diseño). ¿Sirve, o
es una parada de más? Y el ⑨ **corre dos veces**: la primera parte el PRD, y vuelve cada vez que
entra algo por `sf new`. Hay que ver las dos.

**Riesgo ya documentado:** el ⑩ corriendo a mitad de ciclo suma un `intentos_fallidos` de más
(`arreglos.md`). **Si aparece en la corrida, es esto y no algo nuevo.**

---

### T4 · La planificación — ⑪–⑰

| | |
|---|---|
| **entra por** | `sf take <f-#>` (el ⑪ — decisión de Javier, no la toma sf) |
| **estado / skill** | `planificacion` / `sf-plan` / **un subagente de contexto amplio** |
| **produce** | `decision.md` (⑫) · `spec-design.md` (⑬) · `tareas.json` (⑭–⑯) |
| **checkpoints** | **no hay campo**: se deducen mirando qué archivos existen |
| **compuerta** | `compuerta.Planificacion` — 5 chequeos, incluido *tres opciones exactas* y *cada tarea apunta a un criterio* |
| **se mueve** | `features[f].base_commit` al cerrar · `approve` la pasa a `implementar` |
| **termina en** | 🛑 el ⑰ — **la única parada sin horizonte**, y es deliberado (`maquina-estados.md` §5.1) |

**Preguntas abiertas.** ¿El subagente de contexto amplio aguanta los cinco pasos sin degradarse? ¿El
⑯ elige bien el modelo por lote, o pone el mismo en todos? Y la del ⑰: sus **tres puertas** —pido
cambios, otra feature, implementar— ¿se usan las tres, o una sola?

---

### T5 · Implementar — ⑱–⑳

| | |
|---|---|
| **el ciclo** | `sf lote start` (ve el rojo, guarda `hash_tests`) → trabajo → `sf done --msg` (ve el verde, compara el hash, commitea) |
| **compuertas** | tres: la branch antes · **el rojo en el medio** · el verde y el hash después |
| **se mueve** | `features[f].lotes[]`: `rojo`, `hash_tests`, `commit` · `intentos_fallidos` |
| **el bucle** | **N lotes** — y es el primer bucle real de la máquina |
| **el corte** | `TopeIntentos = 3` → **ME TRABÉ**. Vive acá y **sólo** acá |

**Preguntas abiertas.** Las dos joyas de la máquina se ejercitan por primera vez con un modelo de
verdad: ¿el rojo obligatorio se puede cumplir sin fabricar un test falso? ¿El `hash_tests` atrapa a
un modelo que afloja un assert para llegar al verde? Y la del bucle: **¿tres intentos es el número,
o se descubre otro?** Hoy es el único número mágico del paquete, y está solo arriba de todo para que
se vea.

---

### T6 · Revisión y cierre — ㉑–㉓

| | |
|---|---|
| **㉑㉒** | `revision` / `sf-check` / subagente grande → `revision.json`: criterios, mutantes, hallazgos |
| **compuerta** | `compuerta.Revision` — todos los criterios con veredicto · **cero hallazgos abiertos** · avisa (no frena) en vuelta ≥3 |
| **el bucle** | un hallazgo abierto → vuelve a `implementar` y **abre un lote de corrección** |
| **㉓** | `cierre` / `sf-cierre` → `doc.md` + `journal.md` → ⏸ → `sf approve` archiva, mergea y borra la branch |

> **Éste es el tramo que más importa, porque acá está el agujero peor.**

Verificado el 2026-09-03: la vuelta de revisión **no incrementa `intentos_fallidos`**
(`done.go:cerrarRevision`), y cerrar el lote de corrección **lo resetea a cero** (`done.go:259`). O
sea que **ME TRABÉ nunca se dispara en este bucle**. El único freno es el aviso de `vuelta >= 3`, que
no frena — y ese número **lo escribe el propio revisor**, no `sf`.

**Preguntas abiertas — y son las del proyecto, no las del tramo:**

- Un revisor LLM al que se le pregunta *"¿qué está mal?"* **siempre** contesta algo. ¿Dónde está el
  punto fijo?
- `revision.json` permite hallazgos con `origen: 22` **sin criterio asociado** — opiniones de calidad
  sueltas, de las que hay infinitas. ¿Se le limita a lo que cuelga de un criterio o de un mutante
  sobreviviente?
- El lote de corrección **exige que la suite falle** (`lote.go`). Un hallazgo del tipo *"el diseño
  está flojo"* no lo puede poner en rojo nadie. ¿Qué hace el implementador ahí — se traba, o fabrica
  un test?

**Hecho de la corrida real de Javier:** llegó a **7 vueltas**. Los artefactos no se conservaron, así
que el mecanismo exacto **no está verificado**; esta corrida es la que lo va a decir.

---

## 6. T7 · El ticket — no es un tramo, es otra puerta

Javier, 2026-09-04:

> *"La máquina debería permitir todo el ciclo de desarrollo completo pero también trabajar por
> ticket cuando un error aparece o es reportado o cuando se decide crear una feature no
> contemplada."*

**Lo que hay hoy, verificado** (`maquina/entradas.go:48`):

```go
if !e.Producto.ConstitucionSellada {
    ef.falla("este producto todavía no está armado: seguí con `sf next` desde el brief")
}
```

> **No se puede meter un ticket en un proyecto que no pasó por brief → PRD → constitución.**

El camino corto **existe** —un `tipo: bug` saltea planificación y revisión, `maquina-estados.md`
§8— pero **sólo adentro de un producto con la ceremonia entera hecha**. Llega un bug a un proyecto
que ya existe, y `sf` pide que primero se le invente un brief.

**Eso es, en una línea de Go, la queja de que la máquina *"sólo sirve cuando hay que hacer un
desarrollo"*.**

### Qué hay que contestar acá, y no antes

Este tramo **no se diseña ahora**. Se corre después de T2 —que es cuando ya hay constitución sellada
y por lo tanto `sf new` acepta algo— y las preguntas son:

1. **¿Un ticket necesita las tres cosas, o sólo la constitución?** La constitución es la que tiene el
   `test_cmd`, la branch base y las convenciones: sin eso `sf` no puede correr ni una compuerta. El
   brief y el PRD, en cambio, son *el porqué del producto* — y un bug no los necesita.
2. **¿Qué es el mínimo para que un proyecto existente entre a la máquina?** Es la pregunta brownfield,
   y está sin contestar en todo el repo.
3. **¿Una feature no contemplada entra igual que un bug?** Hoy sí: las dos por `sf new`, y el
   `tipo:` del `us-#` las rutea distinto.

> **No se toca `entradas.go` hasta que T1–T2 hayan corrido.** La restricción puede ser correcta y
> estar sólo mal explicada, y aflojar una compuerta antes de tener el dato es exactamente lo que este
> repo le prohíbe al implementador.

---

## 7. El orden, y por qué no es opcional

```
T0  el tablero        el registro mínimo + `sf log`
T1  el brief
T2  el PRD y la constitución
     └─ T7 el ticket   se habilita acá, se corre cuando haya ganas
T3  el backlog y el roadmap
T4  la planificación
T5  implementar
T6  revisión y cierre
```

**El orden está forzado por los datos, no por el gusto.**

> No se puede probar el ⑨ con un PRD escrito **a mano** — porque volvés a los artefactos a mano,
> que es justo lo que tapó los bugs. Para que la entrada del ⑨ sea de verdad, **tiene que haber
> salido de la corrida del ⑦.**

Cada tramo consume lo que produjo el anterior. Saltear uno obliga a fabricar su salida, y ahí se
pierde el punto entero del ejercicio.

**La única excepción es T0**, que no consume nada y sin el cual los demás producen impresiones en
vez de datos.

---

## 8. Dónde se corre

Ya está contestado en `salir-a-la-cancha.md` ③ y no se cambia:

- **No en el repo de SpecForge.** El `sf` bajo prueba estaría construyéndose a sí mismo, y a la
  tercera vuelta nadie sabe qué binario está mirando.
- **Sí en un proyecto chico, descartable, con una idea real** — no un "hola mundo", porque el ⑥
  necesita algo sobre lo que investigar y el ⑨ historias que se puedan partir.
- **Con los skills como symlinks**, no como plugin: un plugin es una copia congelada y cada arreglo
  obligaría a reinstalar antes de seguir.

**Y algo que agrega este plan:** el proyecto de prueba **se conserva**, con su `.docs/`, su
`estado.json` y su `registro.jsonl`. Es la evidencia de cada tramo. La corrida de Javier llegó a 7
vueltas y **no quedó nada** — por eso hoy esa parte del análisis es deducción y no hecho.

---

## 9. Qué produce cada tramo, y qué NO

**Produce un documento de hallazgos**, uno por tramo, con la misma pregunta en cada `✗` — la
disciplina ya está escrita en `salir-a-la-cancha.md`:

> *"¿el skill pidió mal, o la compuerta exige de más? Si es lo primero, se arregla el skill. Si es lo
> segundo, se arregla la compuerta. **Lo que no se hace es aflojar la compuerta para que pase.**"*

**Lo que este plan NO hace, dicho para que nadie lo lea como un olvido:**

1. **No reconstruye.** *"Construimos todo junto"* se convierte muy fácil en *"reconstruyámoslo por
   partes"*. 456 tests pasan y la máquina es coherente. Los tramos **no necesitan reconstrucción:
   necesitan instrumentación y uso.** El producto de un tramo es un hallazgo, no un rewrite.
2. **No cierra las cinco de `vecinos.md`.** Cuatro de ellas esperan datos de estas corridas, y la
   quinta —la aprobación que vence— el propio documento la manda después de la corrida real.
3. **No toca el orquestador conversacional.** Javier lo declaró no negociable el 2026-09-04, y con
   argumento: es el valor del producto. Todo lo de acá se hace sin moverlo.
4. **No implementa `sf run`.** Descartado el 2026-09-04 por lo mismo.

---

## 10. La única pieza nueva que este plan sí propone construir

**El registro (T0).** Y conviene decir por qué no es una feature más:

| Sirve para | Y de paso |
|---|---|
| medir los seis tramos | contar las vueltas de revisión **sin preguntarle al revisor** |
| | ver si el subagente que `sf` pidió se lanzó de verdad |
| | dejar escrito desde dónde salió cada sello |
| | sumar costo y vueltas por feature |

Es la respuesta operativa a la frase de Javier:

> *"sf es la máquina de estado pero **no le dimos herramientas para controlar el estado**."*

Dicho con precisión: **`sf` controla los artefactos y no controla la conducta.**

```
ARTEFACTOS  ¿existe el archivo? ¿pasan los tests? ¿son tres opciones?
            ¿cambió el hash? ¿vio el rojo antes del verde?
              → acá sf es muy bueno

CONDUCTA    ¿lanzaste el subagente? ¿cuántas vueltas diste?
            ¿quién aprobó? ¿te fuiste a hacer otra cosa?
              → acá sf no ve NADA
```

**Los cuatro agujeros del 2026-09-03 son los cuatro de conducta. Ninguno de artefacto.** Eso no es
casualidad: es el mapa de lo que falta.

---

## 11. Cómo se sabrá que quedó bien

**De T0:**

```bash
sf next && sf log --ultimas 1     # → una línea con estado, tipo, skill, via y quien
sf log --feature f-2 --vueltas    # → cuántas veces entró a cada estado
```

**De cada tramo:** dos corridas, dos modelos de obediencia distinta, y

```
diff estado-modelo-A.json estado-modelo-B.json    →  vacío
```

Un diff no vacío **es el hallazgo**: significa que el skill y la máquina no dicen lo mismo, que es la
familia del bug del ⑧.

**Del plan entero:** que las preguntas que hoy son deducciones pasen a ser hechos con un número al
lado. Concretamente estas cuatro, que hoy no se pueden contestar:

1. ¿en qué vuelta hay que cortar la revisión, y por qué ésa?
2. ¿qué se le permite decir a un revisor para que el bucle tenga punto fijo?
3. ¿cuánto de la ceremonia necesita de verdad un ticket?
4. ¿la ⏸ del ⑨ sirve, o es una parada de más?
