# Los comandos

Dieciséis, y **el reparto importa tanto como la lista**: quién llama a cada uno es quién es dueño
de qué.

```
el orquestador    next · approve · reject · take · new · model · dismiss
el que trabaja    context · lote start · done
vos               status · audit · install · uninstall · init
```

**No hay solapamiento**, y no es casualidad: los siete del orquestador son *cómo se le contesta a
una parada*, y los tres del que trabaja son *pedí tu sobre, avisá que terminaste*.

---

## Los códigos de salida

Son parte de la interfaz, no un detalle. Un CLI que siempre devuelve 0 obliga a parsear texto, y
acá el que lee suele ser un agente.

```
0   hay trabajo          seguí el bucle
1   error de verdad      algo se rompió
2   parada  🛑 ⏸ ⚠       no sigue sin vos
3   no queda nada        todas las features cerradas
```

Con eso, *"¿hay algo que hacer?"* es un `if` en cualquier shell y en cualquier harness.

---

# El bucle

## `sf next`

**Dónde estás · qué sigue · con qué skill y modelo · si podés avanzar.** Es el corazón: sin él,
todo lo demás es una lista de tareas.

```bash
sf next
```

```
estado:   implementar
feature:  f-1 · lote 2 de 3
skill:    sf-build
modelo:   sonnet
via:      subagente

el rojo está confirmado: implementá

  sf context
  sf done --msg "…"
```

**Los campos, y por qué están todos:**

| Campo | Qué es |
|---|---|
| `estado` | uno de los 9 |
| `feature` | y el lote, cuando aplica |
| `skill` | **qué invocar.** Por esto el orquestador no se sabe la tabla `estado → skill` |
| `modelo` | qué modelo pide este paso |
| `via` | `vos` · `subagente` · `consola` — **cómo** se lanza |
| `comando` | sólo con `via: consola`: con qué se invoca |

> **Sin `skill`, `modelo` y `via`, esa tabla viviría en `CLAUDE.md` y habría que mantener una
> copia por harness.** Devolviéndolos, `CLAUDE.md` son cuatro líneas y `AGENTS.md` es el mismo
> texto.

**No lleva flags** — todo sale del `estado.json`.

---

## `sf context [--completo]`

**El sobre del estado actual.** Lo corre **el que trabaja**, no el orquestador.

```bash
sf context
```

```
# Sobre — implementar · f-1 · lote 2

## El manual del proyecto
.docs/constitucion.md

## Qué hay que construir y cómo
.docs/features/f-1-nucleo/spec-design.md

## Tus tareas (lote 2 de 3)
t-3  exponer el veredicto por la CLI
    satisface: us-4/CA-1
    test: cmd/sf/main_test.go::TestBriefJSON

## Los criterios que tenés que satisfacer
.docs/backlog/us-4.md
```

**No lleva argumentos, y es a propósito:** el estado, la feature y el lote **son todos deducibles**
del `estado.json`, y un argumento deducible es un argumento que se pasa mal. El skill dice *"dame
mi sobre"*; **cuál es el sobre lo decide `sf`**.

Sirve **rutas** por defecto —el que trabaja las abre con sus herramientas y aprovecha el caché de
su harness— y **embebe** sólo lo que no tiene ruta: el lote filtrado, el resumen del diff, la
corrida de mutantes.

### `--completo`

Embebe el contenido de cada archivo en vez de listar la ruta.

```bash
sf context --completo
```

**No es una comodidad:** es para `via: consola`. Un modelo por CLI **no tiene shell propia**, así
que no puede abrir archivos — el orquestador corre esto y le pega la salida en el prompt.

> **`sf context` y `sf done` los corre el que trabaja. Si el que trabaja no tiene manos, el
> orquestador se las presta.**

---

## `sf done [--msg "…"]`

**Corre las compuertas y mueve el estado — o dice qué falta.** Lo corre **el que trabaja**, antes
de morir.

```bash
sf done
```

```
✗ No avanzo.
  · falta .docs/features/f-1-nucleo/tareas.json
```

> **Por qué lo corre el subagente y no el orquestador:** si da ✗, el subagente **todavía tiene el
> contexto** para arreglarlo ahí mismo. Si lo corriera el orquestador, cada ✗ costaría un
> subagente nuevo desde cero.

### `--msg` — sólo en `implementar`

```bash
sf done --msg "feat: el pool de workers"
```

**Sin mensaje no hay `done`.** Cerrar el lote **es** commitear:

- **`sf` hace el commit**; el mensaje lo trae el que trabajó, que es el único que sabe qué hizo.
- Antes de commitear corre los tests **él** y compara el hash de los archivos de test contra el
  que guardó en el rojo.

---

## `sf lote start`

**Crea la branch · exige el ROJO · guarda el hash.** Lo corre el que trabaja, después de escribir
los tests y **antes** de implementar.

```bash
sf lote start
```

```
✓ lote 1 · branch feat/f-1-nucleo · rojo confirmado
```

Tres cosas, y **se niega si los tests no están honestamente en rojo**:

| | |
|---|---|
| **crea la branch** | con el patrón de tu constitución. Nunca la creás vos — así se olvidaba |
| **exige el rojo** | contra los tests **exactos** que el plan nombró, no contra "los que haya" |
| **guarda el hash** | de los archivos de test, para poder comparar en el verde |

Dos formas de fallar:

```
✗ el test planificado no existe: core_test.go::TestSinEstado
✗ la suite YA PASA, y todavía no se escribió el código del lote 1.
```

> **Un test que pasa antes de que exista el código es un test de mentira.**

---

# Las cinco respuestas a una parada

Las corre **el orquestador**, nunca el subagente. Son cómo le contestás a un 🛑, una ⏸ o un ME
TRABÉ.

## `sf approve`

**Sella la parada en la que estés** — y hay cinco. `sf` deduce cuál del estado:

```
brief          → guarda el veredicto que escribiste en el archivo
constitucion   → constitucion_sellada = true
backlog (⏸)    → backlog_visto = true
planificacion  → la feature pasa a implementar          (el ⑰)
cierre (⏸)     → ARCHIVA: commitea lo pendiente, mergea, borra la branch,
                 mueve la carpeta y deja el archivado en su propio commit
```

**Un comando, cinco significados, ninguno ambiguo.** Las cinco son el mismo hecho —aprobás lo que
se produjo— y qué se sella tiene **una sola respuesta correcta**.

**Y corre la compuerta del estado antes de sellar.** Eso no lo convierte en `sf done`: `done`
**mueve** cuando la compuerta pasa; `approve` **sella lo que decidiste**, y lo único que cambia es
que ya no puede sellar algo que la máquina sabe que está roto. La decisión sigue siendo tuya; deja
de poder ser una decisión sobre un artefacto inválido.

```bash
$ sf approve            # la constitución sin test_cmd
✗ No pude.
  · la constitución no tiene `test_cmd:` — sin eso sf no puede correr los tests
```

Los **avisos** de la compuerta no frenan y se muestran igual, salga bien o mal: esconderlos detrás
de un ✓ es la forma más fácil de que nadie los lea.

> **El archivado hace el git antes de mover la carpeta**, que es al revés de como se lee. El
> movimiento es lo único irreversible y el git es lo único que puede fallar: si el merge conflictúa,
> no se movió nada y volvés a correr `sf approve`. Detalle completo en
> [`estados.md`](estados.md#-cierre--la-doc-y-el-aprendizaje).

## `sf reject "motivo"`

**No sella, y guarda el motivo.**

```bash
sf reject "el stack está bien pero falta la regla de errores"
```

**El motivo es obligatorio**, y no es burocracia:

> El que rehace es un subagente **nuevo**. Sin el motivo vuelve a proponer exactamente lo mismo.

Por eso **viaja primero en el sobre** — arriba de todo, antes que cualquier otra cosa.

## `sf take <feature>`

**Saca una feature de la cola.**

```bash
sf take f-2
```

Es el ⑪, y también la tercera puerta del ⑰: *"aprobé el plan de f-1 pero voy a planificar f-2
antes de implementar"*.

## `sf model <nombre> [--via …] [--comando "…"]`

**Sube (o baja) el modelo de la feature actual, y resetea el contador de ME TRABÉ.**

```bash
sf model opus
```

Cambiar de modelo es **empezar de nuevo**, no seguir acumulando los fracasos del anterior.

**Vale en los cuatro estados de feature** —`planificacion`, `implementar`, `revision` y `cierre`—,
que es lo mínimo para que sirva: ME TRABÉ puede aparecer en cualquiera de ellos, y una salida que
funciona en uno solo no es una salida.

### De dónde sale el modelo, y quién le gana a quién

```
1. sf model <nombre>   tu decisión en runtime          estado.json
2. tareas.json         el ⑯: "ESTA feature es difícil"  lo escribió el que planificó
3. el default del estado    opus en planificacion y revision
4. el default general       sonnet
```

**El orden no es arbitrario: cada nivel sabe menos que el de arriba.** El default no sabe nada de
la feature; el ⑯ la planificó pero no vio fallar nada; vos estás mirando el bucle patinar cuando
escribís `sf model`.

Los dos primeros niveles **sólo existen dentro de una feature**, así que en los cinco estados de
producto manda el default del estado y listo.

### Declarar un modelo nuevo

Cuando `sf` para porque el modelo no está en tu mapa, la respuesta **lo declara**:

```bash
sf model deepseek --via consola --comando "deepseek exec"
sf model o3 --via subagente
```

```
✓ modelo: deepseek (contador reseteado)
   declarado en ~/.specforge/: via consola · deepseek exec
```

**No hay `sf model add` aparte, a propósito:** separar *declarar* de *usar* crearía un estado
intermedio —declarado y nunca usado— que no le sirve a nadie. **Aprobar es declarar**, y la lista
crece sola.

| Flag | |
|---|---|
| `--via subagente` | el harness lo lanza con sus propias manos |
| `--via consola` | se sale por CLI. **Exige `--comando`** |
| `--comando "…"` | con qué se invoca. Sólo tiene sentido con `--via consola` |

## `sf dismiss <h-#> "motivo"`

**Descarta un hallazgo de la revisión.**

```bash
sf dismiss h-2 "el mutante no representa un caso real"
```

Tapa un bucle infinito real: salir de `revision` es automático, así que **un falso positivo no se
cierra nunca** — el revisor lo encuentra, el implementador no lo puede arreglar porque no está
roto, y vuelta.

`descartado` es **el único estado de hallazgo que no se deduce de nada**: sin persistirlo, la
vuelta siguiente lo vuelve a encontrar porque el código sigue igual.

**El motivo es obligatorio:** descartar sin decir por qué es perder la razón.

---

# Entradas

## `sf new "…"`

**Mete una feature o un bug al backlog.**

```bash
sf new "el login rompe con email en mayúsculas"
```

```
us-7 creada. Ahora se pinponea: `sf next` te lleva.
```

**Y te lleva de verdad:** reabre la ⏸ del ⑨, y el próximo `sf next` te manda a `sfp-backlog`
nombrando la historia que falta (`completá: us-7`). El esqueleto que deja `sf new` **no pasa la
compuerta** — tiene el `CA-1` puesto y vacío, y un criterio sin texto es peor que ninguno: el ⑰ lo
daría por cubierto y el ㉑ le pondría veredicto.

**No lleva bandera para el bug.** El `us-#` ya tiene `tipo: us | bug` en el frontmatter, lo
escribe el que lo completa, y `sf` lo lee para rutear:

```
tipo: us    →  planificacion → implementar → revision → cierre
tipo: bug   →  implementar → cierre        (saltea dos)
```

> **No hay carril paralelo ni segunda máquina.** Hay **un campo que saltea estados** — y por eso
> el rastro no se pierde: el bug igual entró por el backlog, que es el embudo.

Todo lo que venga después del comando es el texto, así que **no hacen falta comillas** si no
querés: `sf new el login rompe con mayúsculas` funciona igual.

---

# Fuera de la máquina

## `sf status`

**Dónde está todo.** El único comando pensado para vos y no para el agente.

```bash
sf status
```

No mueve nada ni comprueba nada: **lee tres fuentes y arma una vista**. Por eso no aparece en
ningún trazado del bucle — el bucle no lo necesita nunca.

> Es **derivado, no almacenado**. Un índice mantenido a mano siempre queda viejo; **lo que se
> calcula no puede quedar viejo.**

## `sf audit [f-# …] [--completo]`

**El punta a punta: varias features contra las historias que debían satisfacer.**

```bash
sf audit                 # todo lo que se construyó
sf audit f-1 f-2 f-3     # estas tres — un "módulo" es un conjunto de features
```

### Qué ve que `sf-check` no puede ver

```
sf-check (㉑)   UNA feature · contra SUS criterios · JUSTO al terminarla
sf audit        VARIAS features · punta a punta · CUANDO VOS QUERÉS
```

El ㉑ revisa `f-2` el día que termina, y **después nadie la vuelve a mirar nunca**. De ahí salen
tres cosas que estructuralmente no puede ver:

- `f-4` **rompió** un criterio de `f-1`;
- `f-2` y `f-3` pasaron solas y **no se integran**;
- un criterio marcado `cumple` que **hoy ya no es cierto**.

### Lo que `sf` comprueba solo

```
✓ alcance: 3 features · 14 criterios
✗ us-3/CA-2 se dio por cumplido en f-1 y su test ya no existe: core_test.go::TestSinEstado
✗ us-5/CA-4 no tiene veredicto en la revisión de f-2
⚠ f-2/h-1 se descartó: el mutante no representa un caso real
✓ la suite pasa (go test ./...)
```

**El primero es el que atrapa la mentira, y no es una opinión:** la revisión afirmó que el
criterio se cumple, y el test que lo probaba **ya no está**. Nadie más en el flujo puede verlo.

Los hechos van **arriba** del material a propósito: el que lee es un modelo que va a gastar
contexto abriendo archivos, y **saber qué buscar cambia qué abre**.

Devuelve **2** si encontró contradicciones, **0** si está limpio.

---

# Instalación

## `sf init`

**El andamio del proyecto**: 2 directorios, detecta el stack, escribe la constitución con la
cabecera llena y el `estado.json` vacío.

**Nunca pisa un `estado.json` ni una constitución que ya existan.** Ver
[`../INSTALL.md`](../INSTALL.md).

## `sf install [--harness=<n>] [--forzar]`

**Pone `CLAUDE.md` y `AGENTS.md` y arma `~/.specforge/`.**

| Flag | |
|---|---|
| `--harness=<nombre>` | pisa la detección. Es la única forma de corregirla |
| `--forzar` | pisa un `CLAUDE.md` que ya exista |

## `sf uninstall`

**Saca el orquestador del proyecto.** Sólo borra los archivos **si no los editaste**.

**Nunca toca `~/.specforge/`** —tus modelos son de la máquina, no del proyecto— ni `.docs/`, que
es tu trabajo.
