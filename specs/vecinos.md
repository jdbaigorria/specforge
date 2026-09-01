# Los vecinos — cinco compuertas que faltan

**Fecha:** 2026-09-01 · **Branch:** `refundation` · **Commit:** `caf7e55`

> **Estado: SIN IMPLEMENTAR.** Las cinco salen de comparar `sf` contra dos proyectos que llegaron
> solos al mismo diagnóstico. Ninguna necesita una dependencia nueva, ninguna toca el frente de
> producto (⑥–⑩) —que es justamente lo que ninguno de los dos tiene—, y las cinco caen sobre
> pendientes que este repo ya tenía escritos.

Este documento continúa a [`arreglos.md`](arreglos.md) y a [`headless.md`](headless.md). Aquéllos
salieron de leer y de medir el propio código; éste sale de **leer el código de otros** y encontrar
qué cierran ellos que acá está abierto.

---

## 1. De dónde sale esto

Los dos los trajo Javier el 2026-09-01. Los dos se leyeron **en código**, no en su landing.

### 1.1 `specd` — el gemelo

`github.com/0xkhdr/specd`, v1.1.0 del 2026-08-01. MIT.

Se bajó del proxy de módulos (`proxy.golang.org/github.com/0xkhdr/specd/@v/v1.1.0.zip`) porque
`raw.githubusercontent.com` da 404 sobre ese repo. Lo que hay adentro:

| | |
|---|---|
| lenguaje | Go, **stdlib sola** — el `require` vacío de `go.mod` es un chequeo de release |
| tamaño | 19.6k líneas de código · 16k de tests |
| entrega | un binario estático, linux/amd64 |
| lema | *"The agent reasons. The harness enforces."* |

Y su diagnóstico, textual:

> *"An AI agent is good at reasoning and bad at remembering a process it agreed to twenty thousand
> tokens ago. specd assumes that."*

Es el §1 de `headless.md` escrito por otra persona. La coincidencia sigue: cero LLM y cero red en el
camino de enforcement, Markdown del usuario contra JSON de la herramienta, y *"every refusal carries
**exactly one legal next action**"*, que es el commit `d1c7856` de este repo.

**Que dos proyectos que no se conocen lleguen a la misma forma es evidencia de que la forma está
bien.** Lo que sigue no es "nos ganaron": es qué afiló cada uno mientras el otro miraba para otro
lado.

### 1.2 `agent-spec` — el otro eje

`ZhangHanDong/agent-spec`, v1.2.0, Rust, ~60 comandos: DSL de contrato, grafo de código (Atlas),
wiki vivo, daemon, MCP, SDK de proveedores, arnés de benchmarks.

**El tamaño es exactamente lo contrario de la disciplina de una sola dependencia de este repo, y
por eso de acá se trae UNA idea y nada más** (§6). Su Atlas es sólo-Rust, así que su parte más
vistosa no le sirve a una máquina agnóstica al lenguaje.

Dos cosas suyas que **no** son tareas y conviene anotar igual, porque validan decisiones ya tomadas:

- Su `--ai-mode caller` —la herramienta corre lo mecánico, emite `AiRequest[]` a un archivo, y el
  agente que la llamó devuelve `AiDecision[]` que se fusionan— es **estructuralmente idéntico** a
  `sf context` → subagente → `sf done`. Dos diseños independientes, la misma junta.
- Su liveness *"is derived, never stored"*. Es la regla de `sf next` como consulta pura.

### 1.3 Lo que la comparación NO dice

**No dice que estemos atrás.** `specd` es brownfield-first y arranca en *"vengo a hacer este
cambio"*: no tiene brief, ni PRD, ni backlog, ni roadmap, ni investigación. Todo el tramo ⑥–⑩ de
esta máquina no existe ahí. Y no tiene **nada** de capa de arneses ni de modelos: no hay `sf lanzar`,
no hay catálogo, no hay elegir con qué corre cada paso.

De la mitad para atrás, acá hay algo que ellos no tienen. De la mitad para adelante, ellos afilaron
cinco cosas que acá quedaron a medias. Eso es todo lo que dice.

### 1.4 La incómoda, que también es un dato

`specd` publicó v1.1.0 con instalador, binario y sitio de documentación el 1 de agosto. La
refundación de este repo **nunca se pusheó a `main`** (`salir-a-la-cancha.md` ①). Vale como
argumento para cerrar aquel documento antes que éste.

---

## 2. ① La firma del ⑰ — quién aprueba, y por qué hoy no se sabe

### El dolor

Las tres paradas de Javier —el ⑥, el ⑧ y el ⑰— son el punto entero de la máquina. Si un agente
las puede pasar solo, no son paradas: son un comentario.

### Lo que hay hoy

`sf approve` no comprueba nada sobre quién lo corrió (`maquina/paradas.go:181`). Y el agujero ya
estaba escrito: `arreglos.md:396` dice, textual, que la máquina ofrece el ⑰ *"pero un `sf approve`
tipeado directo la saltea igual"*.

Y del otro lado: **la primitiva para cerrarlo ya está construida**. `hayPersona`
(`cmd/sf/main.go:804`) contesta *"¿hay alguien del otro lado de esta terminal?"*, salió con stdlib
sola, se midió el 2026-08-31, y ya tiene cazado el caso de `/dev/null`. Se usa **sólo** para decidir
si `sf install` pregunta.

`specd` usa esa misma comprobación —un ioctl termios sobre stdin— como **autoridad**: `approve` y
`sync` son human-only y *"there is no flag that changes this"*.

### POR QUÉ LA RESPUESTA OBVIA ROMPE SPECFORGE

Y acá está la razón por la que esto es una sección larga y no una línea.

**Copiar a `specd` tal cual —exigir terminal para aprobar— rompe el diseño de este repo.** La
plantilla del orquestador, que `sf install` escribe en cada proyecto, dice esto
(`andamio/plantilla/CLAUDE.md:30`):

```
4.  Si `sf` dice 🛑 o ⏸, mostráselo al usuario y esperá.
    Su respuesta es: sf approve · sf reject "motivo" · sf take <f-#> ·
                     sf model <nombre> · sf dismiss <h-#> "motivo"
```

O sea: **la máquina le pide EXPLÍCITAMENTE al agente que corra `sf approve`.** Javier contesta en el
chat y el agente presta las manos. Es la misma arquitectura que hace que `sf` pueda durar
milisegundos: el que está vivo es el agente.

Adentro de la herramienta Bash de Claude Code no hay terminal. O sea que una compuerta de TTY sobre
`sf approve` haría que **todas** las aprobaciones se rechacen, siempre, en el camino normal.

> **La distinción que ordena todo lo que sigue:** el problema no es *quién tipeó el comando* —eso ya
> se decidió, y la respuesta es el agente—. El problema es **si Javier estaba ahí para contestar**.

### Lo que `sf lanzar` cambió — y lo que NO cambió

**Corrección de Javier, 2026-09-01, y es la que ordena esta sección entera.** La primera versión de
este documento decía que `sf lanzar` había abierto un agujero nuevo: un hijo delegado aprobando su
propio plan. **Es falso, y se comprueba en el código.**

El prompt que recibe el hijo es, entero (`lanzar/pedido.go`, `Armar`):

```go
return "Usá la skill " + i.Skill + ".\n\n" + sobre
```

Nada más. Ni protocolo, ni bucle, ni una mención de `sf approve`. Y el skill del paso lo para donde
tiene que pararlo — `sf-plan/SKILL.md:223` le hace **imprimir** el menú de Javier y después cerrar:

```
Awaiting: sf approve  /  sf reject "motivo"  /  sf take <other feature>
```
> *"Then `sf done`."*

O sea: **el hijo headless hace su paso, deja escrito qué podría contestar Javier, corre `sf done` y
se muere.** No aprueba. Actualiza el estado, que es otra cosa.

### Entonces qué queda en pie

Tres cosas, y ninguna es la que decía la primera versión.

**① El agujero es VIEJO, no nuevo.** `arreglos.md:396` lo registró antes de que `sf lanzar`
existiera: `sf approve` no comprueba nada sobre quién lo corrió, y punto. Que un hijo delegado *no
esté instruido* para aprobar no es lo mismo que que *no pueda*. Sigue sin haber guarda mecánica —
sólo que el riesgo lo corre el orquestador de siempre, no un proceso nuevo.

**② Lo que `sf lanzar` sí dejó abierto es otra cosa, y es más interesante: el hijo no sabe que es un
hijo.** Su prompt no le dice qué es. Pero los tres arneses cargan solos el `CLAUDE.md` / `AGENTS.md`
de la raíz del proyecto — que `sf install` escribió **para el orquestador**:

```
1.  Corré `sf next`.
3.  Cuando vuelva el control, corré `sf next` otra vez.
4.  Si `sf` dice 🛑 o ⏸, mostráselo al usuario y esperá.
```

Un proceso que lee eso, sin nadie a quien mostrarle nada y sin una línea que le diga *"vos hacés UN
paso"*, tiene todo para creerse el orquestador. El rol se lo da hoy el **skill**, de costado; el
lanzador no se lo dice.

> **Y el riesgo de eso no es que apruebe: es que SIGA.** Corre `sf next`, ve el paso siguiente, y lo
> hace también. Una corrida delegada que se pidió para un paso y devuelve tres.

**③ La marca sigue valiendo, con otra justificación y menos urgencia.** Deja de ser *"tapa un
agujero que la máquina ya no garantiza"* y pasa a ser *"cerca a un proceso que nadie está mirando"*.
Es defensa en profundidad, no una reparación.

### El arreglo que sale de la corrección, y es el más barato de todos

**Decirle al hijo qué es.** Una línea en `Armar`:

```go
return "Sos un trabajador de UN paso, no el orquestador. Hacé lo que dice la skill,\n" +
       "cerrá con `sf done`, y no corras `sf next` ni ninguna respuesta de Javier\n" +
       "(approve · reject · take · model · dismiss).\n\n" +
       "Usá la skill " + i.Skill + ".\n\n" + sobre
```

Ataca la causa —la confusión de rol— en vez del síntoma. Y respeta *"sf apunta, no pega"*: no
explica la skill ni la estructura interna de nada; sólo declara el rol, que es lo único que el
lanzador sabe y la skill no.

**Esto va primero, y la marca va después.** El prompt evita; la marca atrapa al que igual lo intente.

### Las tres vías, y qué hace cada una

La propuesta es que `sf` deje de tratar todas las invocaciones como la misma, y distinga tres:

| vía | cómo se detecta | qué hace `sf` | por qué |
|---|---|---|---|
| `terminal` | `hayPersona()` da true | **acepta** y lo registra | Javier lo tipeó él mismo. Es la señal más fuerte que existe. |
| `orquestador` | no hay persona **y** no hay marca de delegado | **acepta** y lo registra | Es el camino normal que la plantilla pide. Frenarlo rompe la máquina. |
| `delegado` | está la marca que puso `sf lanzar` | **RECHAZA** | Nadie está mirando. Es el caso nuevo, y es el peligroso. |

**Impedir donde es barato, registrar donde no se puede impedir.** Es la triple defensa de
`FIXBUGHIGH` otra vez: se *evita* el caso delegado, y se *visibiliza* la diferencia entre una
aprobación tipeada por Javier y una prestada por el orquestador.

> **Lo que esto NO compra, dicho antes de que alguien lo crea:** que la vía diga `orquestador` no
> prueba que Javier haya contestado. Un orquestador que fabrica una aprobación sigue pudiendo. Lo
> que cambia es que **queda escrito** que esa aprobación no vino de una terminal, y `sf status` lo
> puede mostrar. `specd` es honesto sobre el mismo límite: *"the harness can refuse an agent that
> declared itself; it cannot attest that a human is present."*

### El contrato

**⓪ `Armar` le dice al hijo qué es.** Es el arreglo de la corrección, va primero, y es una línea.
Sin esto, lo demás atrapa un síntoma cuya causa sigue en pie.

**① `sf lanzar` marca al hijo.** En `lanzar/correr.go`, el `exec.Cmd` del hijo lleva
`SPECFORGE_DELEGADO=<id de la ficha>` además del entorno heredado. El id, y no un `1`, para que el
registro pueda decir **cuál** corrida intentó pasarse de su paso.

**② Un lugar solo que contesta la pregunta.** Función nueva —`via()`, junto a `hayPersona`— que
devuelve `terminal` | `orquestador` | `delegado`. `hayPersona` pasa a ser su detalle interno y deja
de llamarse desde dos lados.

**③ Las cinco respuestas de Javier consultan `via()`.** Son `approve`, `reject`, `take`, `model` y
`dismiss` (`cmd/sf/main.go:105`) — las cinco, no sólo `approve`: un `sf dismiss` que descarta un
hallazgo es tan decisión de Javier como un sello. Con `delegado`:

```
sf approve: esto lo corrió una tarea delegada (sf lanzar, ficha l-7).
Una parada de Javier no se contesta desde adentro de la corrida que la produjo.
→ volvé a la sesión donde lanzaste y corré `sf approve` ahí.
```

Una negativa, un motivo, **una sola acción legal siguiente**.

**④ La vía queda en el estado.** Cada sello guarda con qué vía se dio. Concretamente, un campo al
lado de cada uno de los que hoy son un `bool` o un hash pelado (`brief_sellado`,
`constitucion_sellada`, y el sello del ⑰), o —mejor— un registro por sello con `via`, `cuándo` y
`qué se selló`. Se decide al implementar; lo que **no** es opcional es que se pueda contestar
*"¿esta aprobación vino de una terminal?"* meses después.

**⑤ `sf status` y `sf doctor` lo muestran.** Un sello por `orquestador` no es un ✗, pero tampoco es
invisible.

### Los tests

- `Armar` nombra el rol y prohíbe `sf next` — comprobable sobre el string, sin correr nada.
- `sf approve` bajo `SPECFORGE_DELEGADO` → rechazo, y el estado **no** se movió.
- Las otras cuatro respuestas, lo mismo.
- Sin la marca y sin TTY → acepta, y el estado dice `via: orquestador`.
- Con TTY simulado → acepta, y dice `via: terminal`.
- `sf lanzar` pone la variable: comprobable en el `--seco`, sin correr un arnés.
- El hijo **hereda** el resto del entorno (que hoy es todo): que la marca no lo pise.

### El riesgo, dicho

Si algún día `sf lanzar` gana un modo donde el hijo *debe* poder cerrar paradas, esto lo frena. No
es un accidente: es la decisión. El día que haga falta, se discute con un caso real —que es la
disciplina de `friction` de `specd` (§8)— y no aflojando la compuerta para que pase.

---

## 3. ② El verde vacío — un test que no corrió no es un test que pasó

### El dolor

Un comando de test que **no encuentra ningún test** termina bien. Sale cero. Para `sf`, es verde.

Es pedir *"revisá que todas las ventanas estén cerradas"* en una casa sin ventanas: la respuesta es
"todas cerradas", es cierta, y no probó nada.

### Lo que hay hoy

`suite.Verde` es literalmente el exit code (`suite/suite.go:47`: *"Verde es si el runner salió con
0"*). Y las dos defensas que sí existen son otras dos, distintas y buenas:

| defensa | dónde | qué caza |
|---|---|---|
| los tests que el plan nombró **existen** | `suite.go:96`, un Contains sobre el archivo | el lote que dice "listo" sin escribir el test |
| los archivos de test **no cambiaron** entre el rojo y el verde | `maquina/done.go:387` contra `lote.go:203` | el test que se afloja para llegar a verde |

**Ninguna de las dos caza la vacuidad.** Que el test exista y que nadie lo haya aflojado no dice que
haya **corrido**. Un `t.Skip()`, un filtro que no matchea, un archivo de test en el paquete
equivocado: las tres pasan las dos defensas y salen cero.

`specd` cierra eso: detecta la corrida que no ejecutó nada, la marca `zero_match: true`, y
**reescribe el exit code a 126**. Su `Passed` exige `NonVacuous`. Y lo dice con todas las letras:

> *"A test command that matches zero tests exits 0. It is a green result that asserted nothing, and
> it is the easiest way to fake progress without lying."*

### El contrato

**① `suite.Resultado` gana un campo `Vacuo bool`.** No se toca `Verde`: son dos hechos distintos y
colapsarlos es el error que se está arreglando.

**② La detección es por runner y por texto, con el mismo criterio que ya rige el paquete.** El doc
de `suite.go:24` ya dice que un parser por runner no paga; acá tampoco hace falta uno: alcanza con
las marcas que cada runner imprime cuando no corrió nada.

| runner | la marca |
|---|---|
| `go test` | `no test files` / `no tests to run` |
| `pytest` | `no tests ran` / `collected 0 items` |
| `jest` / `vitest` | `No tests found` |
| `cargo test` | `running 0 tests` |
| desconocido | **no se opina**: `Vacuo` queda false |

El último renglón es R3: para un runner que no conocemos, `sf` no adivina.

**③ El ⑳ trata `Vacuo` como rojo**, con su propio mensaje:

```
los tests corrieron y no ejecutaron NINGUNO. Un verde vacío no prueba nada.
→ revisá que los tests del lote estén donde `test_cmd` los busca.
```

**④ `Vacuo` viaja al ㉑.** El sobre del revisor ya lleva la salida de mutantes; que lleve también si
la última corrida fue vacua. Un revisor que opina sobre cobertura necesita ese dato **como insumo**,
por el mismo argumento que puso a los mutantes en el sobre.

### El riesgo

Un falso positivo acá **frena un lote legítimo**. Por eso la lista de marcas es exacta y el
desconocido no opina. Y por eso `Vacuo` es un campo aparte: si la detección se equivoca, se apaga
sin tocar el camino del verde.

---

## 4. ③ El número de fila — la guarda de revisión

### El dolor

`headless.md §11④`, textual: *"No hay candado sobre el estado. Si dos procesos corren `sf done`
sobre el mismo repo al mismo tiempo, gana el último y en silencio."* Lo único parecido a un candado
es un `O_EXCL` en `entradas.go:64`, y es para crear entradas.

Con `--async` (§10 de aquel documento) esto pasa de aspereza a requisito.

### Lo que hacen ellos, y por qué es mejor que un candado

`specd` **no tiene lock**. Tiene un número. `state.json` lleva una `revision` monotónica que sube en
cada transición, y el que quiere avanzar **declara qué número vio**:

```
$ specd complete --revision 7
error: state moved; observed 7, current 8
```

Es la etiqueta de la fila del banco: nadie te impide acercarte al mostrador, pero **tu número ya no
es el que están llamando**, y eso se nota solo.

Un candado hay que tomarlo, soltarlo, y decidir qué pasa cuando el que lo tiene se muere. Un número
no se toma ni se suelta. Y —esto es lo que lo hace encajar acá— **no necesita que `sf` viva más de
lo que vive**: se lee al principio, se compara al escribir, y listo.

### El contrato

**① `estado.json` gana `revision int`.** Sube en **toda** escritura del estado, sin excepción. Es
un campo nuevo en `estado.Estado` (`estado/estado.go:104`), al lado de `producto`.

**② `sf next` y `sf context` la imprimen.** `next` ya devuelve el estado y la instrucción; que
devuelva también el número que leyó.

**③ `sf done`, `sf lote start` y las cinco respuestas de Javier aceptan `--revision N`.**
Si viene y no coincide → se rechaza, no se mezcla:

```
sf done: el estado se movió mientras trabajabas (viste 7, ahora va 8).
→ corré `sf next` y mirá qué cambió antes de cerrar.
```

**④ El flag es opcional, y ahí está el matiz que hay que respetar.** Sin `--revision`, todo se
comporta como hoy. Un `sf` que de golpe exige un número rompe cada script, cada test y la plantilla
del orquestador. La guarda **se ofrece**; el que la usa es quien puede perder algo — que es,
justamente, `sf lanzar` y su hijo.

**⑤ `sf lanzar` la usa siempre.** El padre leyó el estado antes de lanzar: que se lo pase al hijo, y
que el `sf done` del hijo la lleve. Ahí la guarda deja de ser opcional en la práctica sin ser
obligatoria en la superficie.

### Un migrado que hay que decidir

Un `estado.json` viejo no tiene `revision`. El cero de Go es `0`, y eso **funciona**: la primera
escritura lo pone en 1 y de ahí en adelante sube. No hace falta migración, y conviene que quede
escrito para que nadie la agregue.

---

## 5. ④ La aprobación que vence

### El dolor

Javier aprueba un plan el lunes. El martes alguien edita `tareas.json`. **La aprobación del lunes
sigue valiendo.**

### Lo que hay hoy

La idea ya está en el repo, **aplicada dos veces a mano y sin generalizar**:

| hash | dónde | qué protege |
|---|---|---|
| `prd_hash` | `estado/estado.go:151` | que el backlog salga del PRD que se aprobó |
| `hash_tests` | `estado/estado.go:257` | que los tests no cambien entre el rojo y el verde |

Dos usos del mismo mecanismo, en dos lugares puntuales. Lo que no hay es la regla general.

`specd` la tiene: `approve` hashea **cada artefacto cubierto** y guarda los hashes por archivo más
uno agregado. Después, ocho razones nombradas para que una aprobación quede rancia
(`approval_missing`, `state_revision_changed`, `artifact_set_changed`, hash que no da, …). Y escriben
la consecuencia, que es la mejor frase del proyecto:

> *"You cannot quietly widen a task's declared files mid-implementation. Scope creep costs a human
> round trip, on purpose."*

### El contrato

**① Cada sello guarda el hash de lo que selló.** El ⑰ es el que importa —sella `decision.md`,
`spec-design.md` y `tareas.json`—, y el ⑥ y el ⑧ entran por el mismo precio.

**② Un hash por archivo, más el agregado.** Por archivo para poder **decir cuál** cambió; agregado
para comparar barato.

**③ La compuerta del ⑱ comprueba la vigencia antes de empezar un lote.** Si un archivo del plan
cambió después del sello:

```
el plan cambió después de que lo aprobaste: tareas.json.
Una aprobación vale sobre bytes, no sobre intenciones.
→ `sf next` te devuelve al ⑰ para que lo mires de nuevo.
```

**④ La razón se nombra.** No alcanza con "rancia": tiene que decir **qué** cambió, porque la
respuesta de Javier es distinta si se movió `tareas.json` (ensanchamiento de alcance) o
`spec-design.md` (cambió el diseño).

### El riesgo, y por qué éste va último de los cuatro mecánicos

**Éste es el que más va a molestar.** Un sistema que te manda a re-aprobar cada vez que tocás una
coma del plan es un sistema que vas a querer apagar. Y "que cueste una vuelta humana" es la
intención, no un efecto colateral — pero la intención hay que probarla en una corrida real antes de
cablearla.

Por eso: **se implementa después de `salir-a-la-cancha.md` ③**, con la experiencia de haber corrido
el bucle de verdad. Es la misma disciplina que sacó a `--async` de la primera versión de headless.

---

## 6. ⑤ Saltear no es aprobar

### El dolor

En el ㉑, el revisor tiene que opinar sobre **cada** criterio, y la compuerta comprueba que no falte
ninguno (`compuerta/compuerta.go:416`). Perfecto.

**Pero no comprueba qué dice la opinión.** `Criterios` es `map[string]string`
(`revision/revision.go:68`) y `Leer` no valida el valor. Entonces esto pasa la compuerta hoy:

```json
"criterios": { "us-1/CA-1": "no lo pude verificar" }
```

Opinó. Hay una línea. Pasa. **La planilla de control de calidad está completa y el control no se
hizo.**

Y no es hipotético: es el desenlace natural de un revisor honesto que se topa con un criterio que no
puede comprobar. Hace lo correcto —lo dice— y la máquina lo premia dejándolo pasar.

### Lo que hace `agent-spec`

Cinco veredictos cerrados, y una regla de una línea: **`skip ≠ pass`**.

| veredicto | qué significa |
|---|---|
| `pass` | verificado |
| `fail` | violación concreta |
| `skip` | ningún verificador lo cubrió |
| `uncertain` | la IA miró y no está segura |
| `pending_review` | hace falta juicio humano |

### El contrato

**① `Criterios` pasa de string libre a una lista cerrada.** Tres valores, no cinco — los otros dos
de `agent-spec` presuponen su pirámide de verificadores, que acá no existe:

| valor | qué hace la máquina |
|---|---|
| `cumple` | avanza |
| `no-cumple` | es un hallazgo: vuelve a implementar |
| `no-verificable` | **para y se lo muestra a Javier** |

**② El tercero es el que vale, y no puede colapsarse con ninguno de los otros dos.** No es `cumple`
—no se verificó nada— y no es `no-cumple` —no se encontró ninguna falla—. Es *"esto necesita a un
humano"*, y mandarlo a Javier es lo mismo que hace la máquina con todo lo demás que no puede decidir
sola. Es R3.

**③ Un valor fuera de la lista es un `revision.json` corrupto**, con el mismo trato que un JSON roto:
se nombra la lista en el error.

**④ `sf dismiss` sirve para el `no-verificable`.** Ya existe, ya es el comando con el que Javier
descarta un hallazgo con motivo, y ya está en la plantilla del orquestador. No hace falta un verbo
nuevo.

**⑤ El skill `sf-check` se actualiza en el mismo commit.** Si la compuerta cambia y el skill sigue
diciendo lo de antes, el desalineado lo va a encontrar `sf doctor` — y peor, lo va a sufrir la
primera corrida.

### El costo

~20 líneas de Go y un enum. Es el mejor cociente valor/costo de los cinco.

---

## 7. El orden, y por qué ése

```
⓪  decirle al hijo qué es    una línea en `Armar`. Ataca la causa, no el síntoma
⑤  saltear no es aprobar     ~20 líneas y un enum. El mejor cociente valor/costo
①  la firma del ⑰            la pieza YA existe (hayPersona); es enchufarla
③  el número de fila         media tarde, y cierra headless §11④
②  el verde vacío            media tarde, con la lista de marcas por runner
④  la aprobación que vence   DESPUÉS de la corrida real — es el que molesta
```

> **El ⓪ y el ① se separaron por la corrección del 2026-09-01** (§2). Antes eran uno solo y el ①
> iba primero porque se creía que tapaba un agujero activo. No lo tapa: el agujero es viejo y el
> hijo headless no aprueba. Lo que sí quedó al descubierto es que el hijo **no sabe que es un
> hijo**, y eso se arregla en el prompt, no en la compuerta.

Los cuatro primeros son enforcement puro, stdlib, y ninguno toca el frente de producto. El quinto
es el único que cambia cómo se siente usar la máquina todos los días, y por eso espera evidencia de
uso.

**Y ninguno de los cinco es previo a `salir-a-la-cancha.md` ③.** La corrida real sigue siendo lo que
más vale, y no depende de esto.

---

## 8. Lo que se decide NO hacer, y qué lo despertaría

`specd` tiene dos cosas más que acá **no** se traen, y conviene dejar escrito por qué para que nadie
las agregue creyendo que fue un olvido.

### 8.1 El alcance de archivos declarado por tarea — se trae a medias, y a propósito

En `specd`, cada tarea declara qué archivos puede tocar, y **un diff fuera de esa lista se rechaza**.
Los gitignorados cuentan, con un argumento que vale copiar tal cual: *"honoring `.gitignore` would
let an agent write anywhere by adding an ignore rule."*

Acá, `Tarea` (`tareas/tareas.go:134`) tiene `id`, `lote`, `descripcion`, `satisface` y `tests`. **No
tiene `archivos` ni `depends-on`.**

> **Y no se copia entero porque contradice una decisión ya firmada.** El dolor #6 de la refundación
> se cerró como **AVISAR, no impedir**, con el argumento de Javier: *"la última palabra es de Javier,
> una herramienta que frena sola rompe esa regla."*

Lo que sí se puede traer, el día que se traiga: **el campo y el aviso, sin la prohibición.** La
tarea declara los archivos, `sf done` compara contra el diff, y si se salió del alcance **lo dice**
y deja pasar. Eso respeta la decisión firmada y gana la visibilidad.

**Umbral que lo despertaría:** una corrida real donde un lote toque archivos de otra feature y nadie
se entere hasta el ㉑.

### 8.2 El resto de `agent-spec`

Grafo de código, wiki vivo, daemon, MCP propio, SDK de proveedores, jerarquía `org.spec` /
`project.spec` / `task.spec`. Nada de eso entra: la superficie de `sf` se defiende sola, el grafo de
código ya lo cubre `codebase-memory-mcp` fuera de la máquina, y su Atlas es sólo-Rust.

Dos que quedan anotadas por si algún día pagan:

- **`lint --min-score`** — puntuar la *vaguedad* de una spec (verbos vagos, criterios sin binding).
  Las compuertas de acá chequean estructura, no vaguedad. Son cosas distintas.
- **"exception paths ≥ happy paths"** — regla sobre los criterios de aceptación del ⑨. Barata,
  chequeable, y ataca un vicio conocido.

### 8.3 La disciplina que sí se copia entera

`specd` tiene un comando `friction`: registrar **qué te bloqueó una capacidad ausente, en el momento
en que te bloqueó**, como única vía legítima para pedir superficie nueva. *"Each of these is deferred
behind a recorded threshold, not forgotten."*

No hace falta el comando —`sfx-journal` es eso—, pero sí la regla: **cada cosa diferida en este
documento tiene un umbral escrito.** Ver 8.1.

---

## 9. Cómo se sabrá que quedó bien

```bash
# ① la firma del ⑰
SPECFORGE_DELEGADO=l-7 sf approve     # → rechaza, y el estado NO se movió
sf approve                            # → acepta, y el estado dice via: orquestador

# ⑤ saltear no es aprobar
# un revision.json con "no lo pude verificar" en un criterio  → ✗ y nombra la lista
# uno con "no-verificable"                                     → para en Javier, no avanza

# ③ el número de fila
sf next                               # imprime revision: 7
sf done --revision 6                  # → rechaza: el estado se movió

# ② el verde vacío
# un test_cmd que no matchea ningún test  → ✗ "corrieron y no ejecutaron NINGUNO"
```

Y la prueba que de verdad cierra el ⓪ + ①, corregida después de la observación de Javier:

> **Lanzar la planificación de una feature con `sf lanzar` y mirar qué hace el hijo cuando termina.**
> Lo esperado —y lo que el skill ya manda hacer— es que imprima el menú del ⑰ y cierre con
> `sf done`. Lo que hay que confirmar es que **no siga**: que no corra `sf next`, no levante el paso
> siguiente, y no toque ninguna de las cinco respuestas de Javier.
>
> Es la única de las seis que se comprueba **observando una corrida real**, no con un test. Y hasta
> que esa corrida ocurra, el ⓪ es una precaución razonable y no un arreglo medido — dicho así para
> que nadie lo lea como un hecho.
