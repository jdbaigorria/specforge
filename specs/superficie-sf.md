# La superficie de `sf` y el reparto

**Fecha:** 2026-08-13 · **Método:** prueba de escritorio del **bucle**, no de los artefactos.
Se recorre una vuelta de feature anotando **cada llamada**: quién la hace, con qué, y qué vuelve.

> **Por qué el inventario y el reparto son un solo tema.** Saber **qué comandos existen** y
> saber **quién los llama** se contesta con el mismo dato: si en el trazado la llamada la hace
> el orquestador, es del orquestador. **El inventario es lo que aparezca.** No se inventan
> comandos por las dudas — así se infló la versión anterior.

**Entra:** los 9 estados de [`maquina-estados.md`](maquina-estados.md), los artefactos de
[`artefactos.md`](artefactos.md) y las siete reglas de [`que-sobrevive.md`](que-sobrevive.md) §2.
**Sale:** el inventario de `sf`, y qué le toca a `CLAUDE.md` / `AGENTS.md` y a cada skill.

---

## 1. Ronda 1 — `planificacion` (⑪ → ⑰)

Se empieza por acá y no por el ⑪ literal porque **este tramo fija el vocabulario**: tiene una
parada de tres puertas, un traspaso a subagente y una compuerta de conteo. Lo que salga acá se
aplica a los otros ocho estados.

### El trazado, tal cual

Punto de partida: `f-1` cerrada, el roadmap tiene `f-2` en el orden 2.

```
ORQUESTADOR   sf next
              → estado:  planificacion
                feature: f-2
                skill:   sf-plan
                modelo:  opus
                lanzá:   subagente fresco

              lanza el subagente
   SUBAGENTE    sf context
                → .docs/constitucion.md
                  .docs/backlog/us-2.md · us-5.md
                  aprendizajes de f-1        (el journal)

                ...trabaja: escribe decision.md, spec-design.md, tareas.json...

                sf done
                → ✓ los tres archivos están
                  ✓ decision.md tiene 3 opciones
                  ✓ los 4 lotes tienen tests
                  ✓ las 12 tareas cubren los 9 criterios
                  ⚠ go.mod trae 1 dependencia nueva: yaml.v3
                  → estado: planificacion ⇒ 🛑 el ⑰

                muere

ORQUESTADOR   sf next
              → 🛑 PARÁ. El plan de f-2 lo revisa Javier.
                   .docs/features/f-2/{decision,spec-design,tareas}
                   ⚠ dependencia nueva sin aprobar: yaml.v3

              (Javier mira y dice qué hace)

              sf approve          →  f-2 aprobada
              sf next             →  "implementar f-2, lote 1, modelo deepseek"
```

Seis llamadas. **Tres las hace el orquestador, tres el subagente**, y el corte es limpio:

```
ORQUESTADOR   sf next · sf approve · sf reject · sf take      ← mueve la máquina
SUBAGENTE     sf context · sf done                            ← trabaja y se reporta
```

---

### H1 — `sf next` devuelve el skill y el modelo, no sólo el estado

**Es la decisión grande de la ronda**, porque define cuánto tiene que saber el orquestador.

Si `sf next` contesta sólo *"planificar f-2"*, el mapeo **estado → skill → modelo** vive en
`CLAUDE.md`. Y ahí hay que mantener uno por harness: `CLAUDE.md`, `AGENTS.md`, el de Codex…
**tres copias de la misma tabla, que se desincronizan** — que es exactamente el argumento por
el que SpecForge terminó siendo un binario (`session.md` §5).

Devolviéndolo:

- **no rompe la regla dura 1.** Decir *"lanzá `sf-plan` con Opus"* **no es lanzar**:
  es servir un dato, como `sf context` sirve un archivo. El que spawnea sigue siendo el
  orquestador.
- **cumple R1.** *"¿Qué skill corresponde al estado `planificacion`?"* tiene **una sola
  respuesta correcta**. Y `sf` ya conoce los nombres: es `sf install` el que los instaló.
- **el modelo ya lo tiene.** Sale de `tareas.json` (lo escribe el ⑯) y de `estado.json`
  (`modelo`, cuando Javier lo subió en el ⑳). Nadie más lo sabe.

**La consecuencia es la mitad del reparto, resuelta de una:**

> **`CLAUDE.md` no lleva el flujo. Lleva cuatro líneas: preguntale a `sf`, hacé lo que diga,
> volvé a preguntar.**

Y entonces `AGENTS.md` es **el mismo texto**, no una traducción. Lo único que cambia entre
harness es *cómo* se lanza un subagente — y eso el harness ya lo sabe hacer.

---

### H1b — Y el modelo no es un dato: es lo que el paso pide contra lo que el harness puede

**El argumento que cerró H1 es de Javier, y es más fuerte que el de las tres tablas:**

> *"si corre en Claude Code sabe que no puede usar un modelo fuera de Anthropic; pero si está
> corriendo en otro harness sabe que puede cambiar entre modelos de proveedores."*

`session.md` §6 ya había registrado la limitación —*"DeepSeek, Grok o Codex no pueden ser
subagentes; para esos el orquestador tiene que salir por consola"*— pero la había dejado como
una nota al pie. **Acá se convierte en un campo**, porque hay exactamente un lugar del diseño
que conoce las dos mitades del problema:

```
tareas.json  (el ⑯)  →  QUÉ modelo hace falta      lo escribe el que planificó
sf install           →  EN QUÉ harness estás       lo sabe el que instaló los skills
sf next              →  ⇒ CÓMO se lanza acá
```

Ningún skill puede contestar eso, y el orquestador tampoco: el skill es un `.md` agnóstico y el
orquestador **es** el harness, no lo sabe describir. `sf` es el único que ve las dos cosas.

```
sf next
→ modelo: deepseek
  via:    consola · deepseek exec        ← en Claude Code

sf next
→ modelo: deepseek
  via:    subagente                      ← en un harness multi-proveedor
```

**El ⑱ deja de ser un caso especial.** Era el último cabo suelto del plan de la sesión anterior
(*"¿qué pasa cuando el que corre `sf` no es un subagente de Claude?"*): resulta que **es la misma
llamada con otro `via:`**. El sobre de `sf context` es idéntico y el skill es el mismo `.md`.

#### El mapa de modelos se construye solo, igual que las dependencias

`sf` **no elige el reemplazo** cuando el modelo no está: eso sería opinar sobre qué modelo se
parece a cuál, y R3 lo prohíbe. Lo que hace es lo que ya hace con las librerías (§6 de
`artefactos.md`): **comparar contra una lista que se llena con cada aprobación.**

```
sf next
→ 🛑 PARÁ. El ⑯ pidió deepseek y acá no está declarado.
     ¿Lo agrego? [s] subagente   [c] consola: ______   [o] otro modelo
```

> **La lista no se escribe de antemano: se construye sola.** Es literalmente el mismo mecanismo
> que `dependencias_aprobadas`, aplicado a otra cosa — un patrón ya firmado, no uno nuevo.

Vive en la constitución global (`~/.specforge/`) porque los modelos de Javier son los mismos en
todos sus proyectos, igual que sus reglas de git.

#### Y afina el reparto: prestarle las manos al que no las tiene

Un modelo por consola puede no tener shell propia. Entonces **no puede correr `sf context` ni
`sf done`** — y no hace falta agregar nada, porque un CLI escribe a stdout:

```
ORQUESTADOR   sf context                     ← lo corre él
              deepseek exec "<el sobre> …"   ← y se lo pega en el prompt
              sf done --msg "…"              ← y lo cierra él cuando vuelve
```

> **`sf context` y `sf done` los corre el que trabaja. Si el que trabaja no tiene manos, el
> orquestador se las presta.**

Mismos comandos, mismo sobre. Lo único que cambia es quién los tipea.

---

### H2 — `sf context` no lleva argumentos

En los documentos anteriores aparecieron **tres formas para lo mismo**:

```
sf context implement f-1 --lote 2      (artefactos.md §10)
sf context documentar f-1              (que-sobrevive.md §11)
sf context planificar f-3              (que-sobrevive.md §11)
```

Tres vocabularios —`implement`, `documentar`, `planificar`— y ninguno coincide con los nombres
de los estados (`implementar`, `cierre`, `planificacion`).

**Pero el problema no es el nombre: es que los tres argumentos son deducibles.** `sf` ya sabe
el estado, la feature y el lote — están en `estado.json`, que es lo primero que lee.

> **R6 aplicada a la CLI: un argumento deducible es un argumento que se pasa mal.**

Un subagente que escribe `--lote 2` cuando va el 3 recibe el sobre equivocado y **no se entera**.
Sin argumento no hay forma de errarle.

```
sf context     → el sobre del estado actual, y nada más que eso
```

---

### H3 — `sf save` no existe

Había aparecido como `sf save tareas --feature=f-2`. **No hace falta**: el subagente ya sabe
escribir archivos —es lo que hace todo el día— y `sf context` ya le dijo en qué carpeta.

La única razón para tenerlo sería validar el schema al guardar. **Pero eso ya lo hace `sf done`**,
y hacerlo dos veces es dos lugares donde el schema puede quedar viejo.

Queda un caso que hay que mirar de frente, y sale limpio: **`estado.json` sólo lo toca `sf`**
(regla dura 3). ¿Necesita el subagente escribir algo en el estado durante `planificacion`? No:
`base_commit` lo pone `sf done` cuando cierra el estado, y el sello del ⑰ lo pone `sf approve`.
**El subagente no toca el estado nunca.**

---

### H4 — Las tres puertas del ⑰ son dos comandos, y uno de ellos es el ⑪

`maquina-estados.md` §6 fijó que el orden lo elige Javier y no hace falta configurarlo. Acá se
ve **con qué llamada** elige cada puerta:

```
implementar     →  sf approve
otra feature    →  sf approve   +   sf take f-3
pido cambios    →  sf reject "el diseño B no cierra: no contempla el modo lote"
```

**`sf take` es el ⑪.** La máquina lo había declarado *"no es un estado, es la transición que
entra a `planificacion`"* — y una transición que Javier elige **necesita un nombre para ser
invocada**. Acá lo consigue.

**Por qué dos comandos y no uno con bandera** (`sf approve --luego=planificar-otra`): cada
comando hace un solo trabajo, y `sf take` sirve igual **fuera** del ⑰ — es como se elige la
primera feature de todas, y como se cambia de idea a mitad de la cola.

**Y `sf reject` lleva el motivo como argumento**, porque el motivo es lo único que el
orquestador va a pasarle al subagente nuevo. Sin él, el que replanifica arranca sin saber qué
estuvo mal y **vuelve a proponer lo mismo**.

---

### H5 — `sf approve` es **uno solo** para las tres paradas

Las tres 🛑 (⑥ el brief · ⑧ la constitución · ⑰ el plan) son el mismo hecho: **Javier aprueba
lo que se produjo**. Lo que cambia es qué se sella, y eso `sf` lo deduce del estado en el que
está — una sola respuesta correcta, R1 otra vez.

```
en brief         sf approve  →  "brief_sellado": "hacelo"
en constitucion  sf approve  →  "constitucion_sellada": true
en planificacion sf approve  →  f-2 pasa a "planificada"
```

Las ⏸ (⑨ y ㉓) se destraban con **el mismo comando**. La diferencia entre 🛑 y ⏸ no está en el
comando: está en **si `sf next` te muestra un resumen y espera, o pasa de largo** — y eso ya es
configurable (`para_en:`).

---

### H6 — El que corre `sf done` es el subagente, y el orquestador no lee su respuesta

Dos preguntas que parecían de detalle y no lo son.

**Quién lo corre: el subagente, antes de morir.** Si `sf done` da ✗ —falta un test, las tareas
no cubren un criterio— el subagente **todavía tiene el contexto para arreglarlo ahí mismo**. Si
lo corriera el orquestador, cada ✗ costaría un subagente nuevo que arranca de cero. El más
barato es el que ya está adentro.

**Y el orquestador no lee lo que el subagente devolvió: corre `sf next`.** Esto extiende la
regla dura 2 un escalón más arriba:

> `sf` no le cree al que trabajó. **Y el orquestador tampoco: le cree al estado.**

Si el subagente dice *"listo"* pero `sf done` no movió nada, `sf next` sigue diciendo
`planificacion` y el orquestador relanza. Si el subagente se murió sin correr `sf done`, pasa lo
mismo. **No hay caso donde el reporte del subagente cambie lo que hace el orquestador**, así que
no hace falta leerlo — y el orquestador se ahorra el único contexto que lo podía ensuciar.

De acá sale el contador de ME TRABÉ sin agregar nada: **`sf done` incrementa `intentos_fallidos`
cuando da ✗**, y `sf next` es quien avisa al llegar al tope.

---

### El inventario de la ronda 1

| Comando | Lo llama | Qué hace |
|---|---|---|
| **`sf next`** | orquestador | dónde estás · qué sigue · **con qué skill, qué modelo y por qué vía** · ¿podés avanzar? |
| **`sf context`** | **el que trabaja** | el sobre del estado actual. Sin argumentos |
| **`sf done`** | **el que trabaja** | corre las compuertas. Mueve el estado, o dice qué falta |
| **`sf approve`** | orquestador | sella la parada en la que estás. Uno para las cinco |
| **`sf reject "motivo"`** | orquestador | no sella, y guarda el motivo para el que rehaga |
| **`sf take <feature>`** | orquestador | el ⑪: elige la próxima de la cola |

**Seis comandos, y el reparto salió solo:** los que **mueven la máquina** son del orquestador;
los que **trabajan** son del que trabaja — que es el subagente, salvo cuando no tiene manos y el
orquestador se las presta (H1b).

---

## 2. Ronda 2 — `implementar` (⑱ → ⑳)

El estado donde muere casi todo (`maquina-estados.md` §4), y donde estaban **los dos cabos
sueltos** que la sesión anterior dejó anotados: `sf lote start` y `sf lote done`, que habían
aparecido en el ⑦ y no estaban en ningún inventario.

### El trazado, tal cual

```
ORQUESTADOR   sf next
              → estado:  implementar
                feature: f-2 · lote 1 de 4
                skill:   sf-build
                modelo:  deepseek
                via:     consola · deepseek exec

              lanza (subagente o consola, según el via)
   EL QUE       sf context
   TRABAJA      → .docs/constitucion.md
                  .docs/features/f-2/spec-design.md
                  .docs/features/f-2/tareas.json   ← sólo el lote 1
                  .docs/backlog/us-2.md · us-5.md

                ...escribe los 6 tests del lote 1...

                sf lote start
                → ✓ branch feat/f-2-verificacion creada
                  ✓ los 6 tests planificados existen
                  ✓ los 6 fallan. Rojo confirmado.  hash b70c33
                  → podés implementar

                ...implementa...

                sf done --msg "feat: verifica el frontmatter del brief"
                → ✓ los 6 tests pasan
                  ✓ los archivos de test no cambiaron desde el rojo
                  ✓ commit 4a8f21
                  → lote 1 cerrado

                muere

ORQUESTADOR   sf next   → "implementar f-2, lote 2 de 4"   … ×3
ORQUESTADOR   sf next   → "revisar f-2"
```

---

### H8 — `sf lote done` no existe: es `sf done`

De los dos cabos sueltos, **uno era real y el otro era el mismo comando con otro nombre.**

`sf lote start` **sí existe**, y es la compuerta del medio: la joya del diseño, la que exige ver
el rojo. Pero *"terminé el lote"* y *"terminé el estado"* **no son dos hechos distintos desde
afuera**: el que trabaja terminó lo que le tocaba, y **qué le tocaba lo sabe `sf`, no él.**

```
sf done   en el lote 1 de 4   →  cierra el lote, y sf next dice "lote 2"
sf done   en el lote 4 de 4   →  cierra el lote y el estado, y sf next dice "revisar"
```

> **`sf done` significa siempre lo mismo: *terminé lo que me tocaba*.** Vale igual para los
> nueve estados, y el que trabaja nunca necesita saber si era el último.

Es R6 otra vez: *"soy el último lote"* es deducible de `tareas.json`, así que no se declara.

---

### H9 — `sf` **crea** la branch, no la verifica

`artefactos.md` §6 decía: *"`sf` compara la branch actual contra el patrón y, si no coincide, el
estado no avanza"*. **Comparar es la mitad barata del trabajo.** El patrón está en la
constitución y la feature está en el estado:

```
patron_branch: "feat/{feature-id}-{slug}"  +  f-2 · verificacion  =  feat/f-2-verificacion
```

**Una sola respuesta correcta → R1: la hace `sf`.** Y el dolor #4 (*"no crea la branch por
feature"*) deja de ser un aviso para dejar de existir: no hay forma de implementar sin branch,
porque `sf lote start` la crea antes de dejarte empezar. Es idempotente — del lote 2 en adelante
ya está.

---

### H10 — El commit lo hace `sf`, y el mensaje lo trae el que trabajó

Es la decisión más grande de la ronda, y sale de mirar el dolor #2 de frente:
*"el commit no se hace, hay que vigilarlo"*.

Si el que implementa commitea y `sf done` **verifica**, el dolor sobrevive: se puede terminar sin
commitear, y `sf` lo único que hace es retarte después. **Hay que romper la posibilidad, no
detectarla.**

R1 parte el trabajo en dos mitades limpias:

| | Respuestas | Quién |
|---|---|---|
| **qué archivos, y cuándo** | una sola: lo del lote, ahora | **`sf`** |
| **qué dice el mensaje** | muchas: es prosa | **el que trabajó** |

```
sf done --msg "feat: verifica el frontmatter del brief"
```

**Sin `--msg` no hay `done`.** Y entonces cerrar el lote **es** commitear — no son dos cosas de
las que una se puede olvidar:

```
sf:  ✗ Falta el mensaje del commit. El lote no se cierra sin él.
```

**Los dolores #2 y #3 mueren juntos acá**, que era lo que §5 de `session.md` ya había previsto
(*"si se sabe qué lote terminó, el agrupamiento sale solo"*). El agrupamiento no lo decide nadie
en este momento: **se decidió en la planificación**, cuando se armaron los lotes.

Y no rompe ninguna regla: correr `git commit` no es pensar, y el LLM sigue poniendo lo único que
tiene varias respuestas.

---

### H11 — `sf model <nombre>`: la única forma de cambiar el modelo, y es de Javier

El ⑳ dice *"si falla varias veces, entro yo"*, y `maquina-estados.md` §5 lo convirtió en la
parada **ME TRABÉ**. Acá aparece con qué llamada se destraba.

```
ORQUESTADOR   sf done   ✗ 2 de los 6 tests siguen fallando
                        → intentos_fallidos: 3

              sf next
              → ⚠ ME TRABÉ. El lote 2 de f-2 falló 3 veces con deepseek.
                   La estimación de complejidad del ⑯ se quedó corta.
                   ¿Subo el modelo, o entrás vos?

              sf model opus        ← Javier decide
              sf next              → "implementar f-2, lote 2, opus, via subagente"
```

**Por qué necesita comando propio y no sale de ningún archivo:** `maquina-estados.md` §9 ya lo
había marcado como uno de los dos campos indeducibles — *"`modelo` es el que se está usando
**ahora**, no el recomendado; cuando Javier lo sube en el ⑳, esa decisión no queda registrada en
ningún lado"*. El comando es ese registro.

Y el contador se resetea solo: si el lote cierra, `intentos_fallidos` vuelve a 0.

---

### H12 — `sf check run` no sobrevive como comando expuesto

Venía marcado **SOBREVIVE** en `que-sobrevive.md`, pero el trazado no lo llama nunca. Y se ve
por qué al preguntarse quién lo consumiría: **el que implementa, mientras itera.**

Ese no lo necesita. Ya tiene `test_cmd` servido en el sobre (`constitucion.md`), y correr
`go test ./...` nativo le da mejor salida para depurar que cualquier envoltorio.

> **`sf` corre los tests para juzgar, no para servir.** Los corre dentro de `sf lote start` y de
> `sf done`, que es donde el resultado mueve el estado. Fuera de ahí, correr tests es trabajo del
> que trabaja.

*(El código no se tira: es lo que usan las dos compuertas. Lo que se cae es el subcomando.)*

---

### El inventario después de la ronda 2

| Comando | Lo llama | Nuevo en esta ronda |
|---|---|---|
| `sf next` | orquestador | ahora también dice **qué lote** |
| `sf context` | el que trabaja | |
| **`sf lote start`** | **el que trabaja** | ⬅ crea la branch · exige el rojo · guarda el hash |
| `sf done --msg "…"` | el que trabaja | ⬅ **commitea**, y sin `--msg` no cierra |
| `sf approve` · `sf reject` · `sf take` | orquestador | |
| **`sf model <nombre>`** | orquestador | ⬅ la salida de ME TRABÉ |

**Ocho comandos.** Y los cuatro dolores de operación (#1 #2 #3 #4) ya están muertos, más el #8
por partida doble — el rojo del `lote start` y el hash del `done`.

---

## 3. Ronda 3 — `revision` (㉑㉒) y `cierre` (㉓)

Los dos estados que cierran la vuelta. Salieron **cuatro comandos menos** de los que había
anotados, y ninguno se perdió: los cuatro eran otra cosa con otro nombre.

### El trazado de `revision`

```
ORQUESTADOR   sf next
              → estado:  revision · f-2 · vuelta 1
                skill:   sf-check
                modelo:  opus            ← el ㉑ es "revisá con uno grande"
                via:     subagente

   TRABAJA      sf context
                → .docs/constitucion.md
                  .docs/features/f-2/spec-design.md
                  .docs/backlog/us-2.md · us-5.md      los 9 criterios
                  git diff 9c2e1a..HEAD                el código que quedó
                  mutantes: gremlins · score 0.71 · sobrevivieron 4   ← ya corrido

                ...juzga los 9 criterios · mira los 4 sobrevivientes · agrega los suyos...

                escribe revision.json
                sf done
                → ✓ opina sobre los 9 criterios
                  ✗ 2 hallazgos abiertos
                  → vuelve a implementar

ORQUESTADOR   sf next   → "implementar f-2 · arreglar h-1, h-2"
              ...la vuelta 2 rehace la revisión entera...
```

---

### H13 — La corrida de mutantes va **en el sobre**, y `sf mutation` se cae

El ㉒ tiene tres tiempos (`artefactos.md` §10): corre la herramienta → el modelo mira los
sobrevivientes → agrega los suyos. El primero es determinista y la herramienta está declarada
en la constitución (`mutacion: gremlins`). **R1: lo hace `sf`.**

Lo que decide *cuándo* no es el gusto, son dos cosas que apuntan al mismo lado:

1. **El modelo necesita el resultado como insumo**, no como salida. Si corre después, el modelo
   inventó a ciegas.
2. **`sf` necesita el score igual**, porque `71% → 88%` es un número que compara entre vueltas y
   lo guarda él. Si se lo pide al modelo, le está creyendo al que trabajó — **regla dura 2**.

Como lo corre igual, servírselo es gratis. Entonces **no hay comando `sf mutation`: hay una
línea más en el sobre de `revision`.**

*(Es la única vez que `sf context` **hace** algo además de servir. No rompe nada —correr una
herramienta no es pensar ni lanzar a nadie— y el resultado se cachea por commit, que es lo que
evita pagarlo dos veces cuando el subagente reintenta.)*

**Y no contradice a H12**, que sacó `sf check run`. La diferencia es de quién es el número:

```
los tests    →  el que trabaja los corre para SÍ MISMO, mientras itera   → no es de sf
los mutantes →  el score queda guardado y se compara entre vueltas       → es de sf
```

---

### H14 — Los hallazgos no tienen ciclo de vida: cada vuelta es una foto nueva

`artefactos.md` §10 decía que el hallazgo *"nace `abierto` y después pasa a `arreglado`; ese
cambio no lo hace el revisor: lo hace el flujo"*. **Y dejaba sin contestar quién es "el flujo".**

Las dos respuestas posibles eran malas: si lo marca el implementador, `sf` le cree al que
trabajó; si lo marca `sf`, está juzgando si un arreglo arregló — que es exactamente lo que no
puede hacer.

**La respuesta ya estaba escrita en otro lado.** `maquina-estados.md` §4: *"después del arreglo
se rehace la revisión **entera**, no sólo lo que falló"*.

> **Entonces nadie marca nada.** La vuelta 2 escribe un `revision.json` nuevo, y `h-1`
> simplemente **no reaparece**. Eso *es* estar arreglado.

R6 otra vez, y en el lugar menos esperado: `arreglado` era deducible de la vuelta siguiente.

**Queda un solo estado de hallazgo, y no es deducible de nada: `descartado`.** Lo agregó R3
—*una compuerta frena sobre un hecho, un juez opina*— y es una decisión de Javier: *"esto el
revisor lo marcó y no me importa"*. Sin persistirlo, la vuelta 3 lo vuelve a encontrar, porque
el código sigue igual.

---

### H15 — `sf dismiss` es la tercera salida de ME TRABÉ, y no hace falta una parada nueva

Salir de `revision` es automático (*"¿hay hallazgos abiertos? con uno solo no avanza"*). Javier
no está mirando. **Entonces un falso positivo es un bucle infinito**: el revisor lo encuentra, el
implementador no lo puede arreglar porque no está roto, y vuelta.

**El freno ya existe y no hay que agregar nada:** es el contador de ME TRABÉ. El mismo que cubre
*"el lote no compila hace 3 intentos"* cubre *"este hallazgo no se cierra nunca"*.

```
sf next
→ ⚠ ME TRABÉ. h-2 sigue abierto después de 3 vueltas.
     "el parser acepta frontmatter sin cerrar"
     ¿Lo descarto, subo el modelo, o entrás vos?

sf dismiss h-2 "es intencional: el ⑨ lo tolera a propósito"
```

**Las tres salidas de ME TRABÉ quedan cubiertas por lo que ya hay:** `sf model` (H11),
`sf dismiss`, y *"entro yo"* — que no necesita comando, porque Javier trabaja y después el
estado se cierra igual.

Y esto contesta de paso la pregunta grande que `artefactos.md` §10 había dejado planteada
(*"¿Javier se entera si lo arregla solo?"*) por el otro lado: **si lo arregla solo, se entera al
final en la ⏸ del ㉓; si NO lo arregla solo, se entera antes por ME TRABÉ.** No hay forma de que
la máquina forcejee en silencio.

---

### El trazado de `cierre`

```
ORQUESTADOR   sf next   → cierre · f-2 · sfx-documenter · sonnet

   TRABAJA      sf context
                → git diff 9c2e1a..HEAD          la mitad técnica
                  us-2.md · us-5.md              la mitad funcional
                  spec-design.md                 cómo se resolvió
                  aprendizajes de f-1            para no repetir los que ya están

                ...escribe la doc y el journal...

                sf done --msg "docs: cierra f-2"
                → ✓ existe la doc
                  ✓ existe el journal
                  → 🅿 f-2 lista para archivar

ORQUESTADOR   sf next
              → ⏸ f-2 lista. 1 vuelta · 2 hallazgos arreglados · 2 aprendizajes.
                   ⚠ uno se repite por 3ª vez:
                     "los tests de tabla en Go tienen que nombrar el caso"
                   [enter] archivo   [c] subir a la constitución   [v] detalle

              sf approve
              → carpeta movida a .docs/archivado/f-2-verificacion/
                merge --no-ff a main · branch borrada
                f-2: "cerrada"
```

---

### H16 — `sf feature archive` no existe: es `sf approve` en `cierre`

Venía marcado **SOBREVIVE** (*"mueve la carpeta entera"*). Pero mirá **dónde cae en el trazado**:
el archivado pasa **después** de la ⏸, no antes. Y lo que destraba una parada ya tiene nombre.

```
sf done      →  ¿está la doc? ¿está el journal?   →  frena en la ⏸
sf approve   →  mueve la carpeta · merge --no-ff · borra la branch · f-2 cerrada
```

Es H5 llevado hasta el final: **`sf approve` sella la parada en la que estás, y qué significa
"sellar" lo decide el estado.** En `brief` es un veredicto; en `cierre` es mover una carpeta y
mergear. Un comando, cinco significados, ninguno ambiguo.

**Y el merge se parte igual que el commit** (H10): lo mecánico es de `sf` —`merge --no-ff`,
borrar la branch, y ambos salen del bloque `git:` de la constitución—; el PR, si la constitución
lo pide, lo arma el skill, **porque un PR lleva título y descripción, que son prosa**.

**La opción `[c]` no necesita comando.** Subir un aprendizaje a la constitución es escribir prosa
en un `.md` → es del LLM, y el orquestador lanza al skill. Y no hay que marcar que ya se subió:
**una vez que está en la constitución, el próximo journal no lo repite**, porque el planificador
la lee. Se resuelve solo.

---

### El inventario después de la ronda 3

```
sf next · sf approve · sf reject · sf take · sf model · sf dismiss     ORQUESTADOR
sf context · sf lote start · sf done                                   EL QUE TRABAJA
```

**Nueve comandos**, y el bucle de feature está trazado entero. La ronda cerró **cuatro** que
estaban anotados como sobrevivientes:

| Estaba anotado | Resultó ser |
|---|---|
| `sf lote done` | `sf done` (H8) |
| `sf check run` | trabajo del que trabaja, no de `sf` (H12) |
| `sf mutation` | una línea del sobre de `revision` (H13) |
| `sf feature archive` | `sf approve` en `cierre` (H16) |

> **Ninguno se perdió: los cuatro eran otro comando con otro nombre.** Es el mismo hallazgo que
> `un estado ≠ un subagente` en la ronda de la máquina — cuando dos cosas parecen distintas y
> caen en el mismo lugar del trazado, es que eran una.

---

## 4. Ronda 4 — los cinco estados de producto (① → ⑩) y las tres entradas

Corren **una vez**, y sólo en la entrada A. Son los más parecidos entre sí —sobre, subagente,
archivo, compuerta de conteo— así que la ronda es corta. Lo que sí aparece es **cómo entran las
otras dos entradas**, que hasta acá no se había trazado nunca.

### El trazado

```
sf init                          scaffold · detectStack() · estado.json vacío

ORQUESTADOR   sf next  → brief · sfx-scout · via: VOS
              (pinponeo ②③④⑤ con Javier, de frente)
              sf done         → ✓ brief.md con veredicto   → 🛑 el ⑥
              sf next  → 🛑 PARÁ. El brief lo sella Javier.
              sf approve      → "brief_sellado": "hacelo"

              sf next  → prd · sfp-po · via: subagente
   TRABAJA      sf context → brief.md
                sf done    → ✓ prd.md · prd_hash a3f9c1

              sf next  → constitucion · sf-init Step 3 · via: subagente
   TRABAJA      sf context → prd.md · ~/.specforge/ ya fundido · el frontmatter técnico ya lleno
                sf done    → ✓ constitucion.md          → 🛑 el ⑧
              sf approve      → "constitucion_sellada": true

              sf next  → backlog · sfx-propose(⑨) · via: subagente
   TRABAJA      sf context → prd.md · constitucion.md
                sf done    → ✓ 7 us-# · ✓ todas con criterios CA-n   → ⏸
              sf approve

              sf next  → roadmap · sfx-propose(⑩) · via: subagente
   TRABAJA      sf context → los 7 us-#
                sf done    → ✓ roadmap.json · 3 features
                             ⚠ f-1 junta 6 historias, ¿la partís?

              sf next  → "planificar f-1"      ← empieza el bucle de la ronda 1
```

---

### H17 — `via:` tiene tres valores, y el tercero es **`vos`**

H1b le había dado dos (`subagente` · `consola`) para resolver el multi-proveedor. El `brief` pide
el tercero, porque es **el único estado que conversa** (`maquina-estados.md` §3): un subagente
arranca, trabaja y muere — **no te habla**, y ①–⑤ es un pinponeo, no una fila.

```
via: vos          el orquestador trabaja de frente. No lanza a nadie
via: subagente    contexto fresco, muere al terminar
via: consola      otro proveedor, por CLI
```

**Y así `via:` deja de significar *"cómo lanzo el modelo"* para significar *"quién hace el
trabajo"***, que es más general y cubre los tres casos con un campo. El orquestador no tiene que
saber cuáles estados conversan: **se lo dice `sf`.**

Cuando `via: vos`, `sf context` y `sf done` los corre el orquestador — **y no es una excepción**:
es la misma regla de H1b (*los corre el que trabaja*), sólo que acá el que trabaja es él.

---

### H18 — `sf init` llena el frontmatter técnico: `constitucion` no arranca de cero

`que-sobrevive.md` §5 había rescatado `detectStack()` porque llena `lenguaje`, `manifiesto` y
`test_cmd` solo. En el trazado se ve **cuándo** conviene que corra: en `sf init`, no en el ⑧.

Entonces el estado `constitucion` arranca con **un archivo a medio escribir**: la cabecera
técnica ya está, y el subagente escribe el cuerpo —arquitectura, stack y por qué, convenciones—
que es lo único que tiene varias respuestas.

**Lo mismo con la constitución global.** Fundir `~/.specforge/` con la del proyecto tiene una
sola respuesta correcta (el proyecto pisa) → R1 → lo hace `sf`, y el sobre la sirve **ya
fundida**. Nadie más resuelve herencia.

---

### H19 — El hueco grande del `roadmap.json` no es un comando: es un parser

`que-sobrevive.md` §13 lo marcaba como *"el hueco más grande — de ahí sale `feature_actual`"*.
En el trazado se desinfla: **lo escribe el subagente del ⑩, como cualquier otro archivo** (H3),
y `sf` sólo lo **lee**.

> El hueco no está en la superficie: está adentro. Es código, no comandos.

Y lo que `sf` lee de ahí es poco: ids, orden y qué historias tiene cada feature. Es el archivo
más chico del inventario (`artefactos.md` §8: *sin títulos, sin estados, sin descripciones*).

---

### H20 — Las entradas B y C entran por el backlog, y el ruteo lo hace el `tipo:`

**Esto no se había trazado nunca.** Los cinco estados de producto corren una vez; las entradas
B (*feature sobre lo existente*) y C (*bug*) llegan a un producto que ya tiene brief, PRD,
constitución y roadmap. ¿Por dónde entran?

`session.md` §3 ya lo tenía escrito sin saber que era esto: **el backlog es el embudo.**

```
sf new "que sf soporte proyectos brownfield"
→ estado: backlog · via: vos        ← el pinponeo de la entrada B
   (conversás, sale us-8.md con su tipo)
sf next  →  "roadmap: ubicá us-8"   →  y de ahí la máquina sigue igual
```

**`sf new` es al backlog lo que `sf take` es al roadmap:** el que mete algo en la cola, contra el
que saca. Es el único comando nuevo de la ronda, y sin él `sf next` no tendría cómo saber que
empezó una entrada.

**Y no lleva bandera para el bug.** El `us-#` ya tiene `tipo: us | bug` en el frontmatter
(`artefactos.md` §7), lo escribe el que pinponeó, y `sf` lo lee para rutear:

```
tipo: us    →  planificacion → implementar → revision → cierre
tipo: bug   →  implementar → cierre            (el camino corto de maquina-estados §8)
```

**Un dato que ya existía y no tenía consumidor, ahora lo tiene.** No hay carril paralelo ni
segunda máquina: hay un campo que saltea dos estados.

---

### H21 — El camino corto no afloja ninguna regla: afloja **una compuerta**, y por falta de plan

Un bug no pasa por `planificacion`, así que **no hay `tareas.json`** — y `sf lote start` exigía
el rojo contra *"los 6 tests planificados"*. Sin plan no hay lista.

La compuerta no se cae, se afloja a lo que sí se puede comprobar:

```
con plan   →  ¿fallan LOS 6 tests planificados?
sin plan   →  ¿falla AL MENOS UNO?        si pasa todo, el bug no está reproducido
```

Sigue siendo un hecho y sigue siendo un exit code. Y **las dos condiciones que Javier ya sostiene
para el camino corto** —correr los tests, verificar que quedó resuelto— son exactamente las dos
compuertas que quedan: el rojo de `lote start` y el verde + hash de `done`.

*(Y el `hash_tests` sigue funcionando igual, que es lo que impide "arreglar" el bug ablandando el
test que lo reproducía.)*

---

### H22 — `sf status` sobrevive, pero no lo llama la máquina

No aparece en ningún trazado, y por un rato pareció que se caía. **No: tiene un consumidor, y no
es un agente.** Es Javier, cuando quiere mirar dónde está todo sin preguntarle a nadie.

```
sf status
→ ① núcleo del CLI          ✓ cerrada
  ② verificación            ▸ implementando · lote 2 de 4
  ③ brownfield                planificada
```

Es la regla 1.5 (*lo que se puede generar, se genera*) con forma de comando: el backlog, el
roadmap y el índice **no existen como archivo** — son esta vista.

> **Es el único comando de `sf` cuyo consumidor es un humano.** Por eso no aparece en el bucle:
> el bucle no lo necesita nunca.

---

## 5. El inventario completo

Todo lo que apareció en los cuatro trazados, y nada más que eso.

### La máquina — 10 comandos

| Comando | Lo llama | Qué hace | Dónde salió |
|---|---|---|---|
| **`sf next`** | orquestador | dónde estás · qué sigue · skill · modelo · `via` · ¿podés avanzar? | R1 |
| **`sf approve`** | orquestador | sella la parada en la que estás — las cinco | H5 · H16 |
| **`sf reject "motivo"`** | orquestador | no sella, y guarda el motivo para el que rehaga | H4 |
| **`sf take <feature>`** | orquestador | el ⑪: saca la próxima del roadmap | H4 |
| **`sf new "…"`** | orquestador | mete una entrada B o C al backlog | H20 |
| **`sf model <nombre>`** | orquestador | sube el modelo. Salida de ME TRABÉ | H11 |
| **`sf dismiss <h-#> "…"`** | orquestador | descarta un hallazgo. Salida de ME TRABÉ | H15 |
| **`sf context`** | el que trabaja | el sobre del estado actual. Sin argumentos | H2 |
| **`sf lote start`** | el que trabaja | crea la branch · exige el rojo · guarda el hash | H9 · H21 |
| **`sf done [--msg]`** | el que trabaja | compuertas · commit · mueve el estado, o dice qué falta | H8 · H10 |

### Y tres más, que no son de la máquina

```
sf status                        el único cuyo consumidor es un humano   (H22)
sf init                          el arranque: scaffold · detectStack     (H18)
sf install · uninstall · lint    andamio: mantienen sf misma
```

### El saldo contra la estimación previa

`que-sobrevive.md` §14 había estimado **"~10 de máquina + 3 de andamio"** mirando los 47 del CLI
construido. El trazado dio **exactamente 10 + 3**.

**Pero no son los mismos diez.** La lista previa tenía once candidatos:

```
sobreviven  next · done · context · lote start · status · install
se caen     lote done · save · check run · mutation · feature archive     5
aparecen    approve · reject · take · new · model · dismiss               6
```

**Los cinco que se cayeron no se perdieron: cada uno resultó ser otro comando** (H3 · H8 · H12 ·
H13 · H16). **Y los seis que aparecieron tienen todos la misma forma:** son cómo Javier le
contesta a una parada. Ninguna lista los podía anticipar, porque **no salen de los artefactos:
salen del bucle.**

> El número se podía estimar mirando el CLI viejo. **El reparto no** — y era la mitad que
> faltaba.

---

## 6. El reparto

Cae solo del inventario: **quién llama a qué es quién es dueño de qué.**

### `CLAUDE.md` / `AGENTS.md` — el orquestador, y son el mismo texto

**No lleva el flujo.** No sabe los 9 estados, ni los 23 pasos, ni qué skill va con cuál. Lleva
**un bucle de cuatro líneas**:

```
1.  Corré `sf next`.
2.  Hacé lo que diga, con el skill y el modelo que diga:
      via: vos        → trabajás vos, de frente
      via: subagente  → lanzás un subagente fresco
      via: consola    → salís por CLI con el comando que te dio
3.  Cuando vuelva el control, corré `sf next` otra vez.
    NO leas lo que devolvió el que trabajó — el estado es la verdad.
4.  Si `sf` dice 🛑 o ⏸, mostrale a Javier y esperá.
    Su respuesta es sf approve · sf reject · sf take · sf model · sf dismiss.
```

**Eso es todo.** Y por eso `AGENTS.md` es el **mismo archivo**, no una traducción: lo único que
cambia entre harness es cómo se lanza un subagente, y eso el harness ya lo sabe hacer.

> Esta es la mitad del reparto que H1 resolvió de una: **el orquestador no sabe el flujo, lo
> pregunta.** Era literalmente lo que `session.md` §6 pedía —*"no se sabe el flujo de memoria"*—
> y el trazado muestra que alcanza con cuatro líneas.

### Los skills — el **cómo**, y siguen sin saber dónde están

Conservan todo el método que ya tienen (`session.md` §6). Se les agrega el principio y el final,
y se les saca lo que ahora tiene dueño:

```
+  arranca con   sf context      pedí tu sobre, no busques qué leer
+  termina con   sf done         avisá que terminaste
−  las convenciones propias      salen de la constitución            (R4)
−  el modo / carril / fase       la máquina saltea estados           (R4)
```

**Y no saben en qué estado están** — por eso `sf context` no lleva argumentos (H2). El skill dice
*"dame mi sobre"*; **cuál es el sobre lo decide `sf`.** Un mismo `.md` puede servir a dos estados
sin enterarse.

### En una tabla, y no hay solapamiento

| | Sabe | No sabe |
|---|---|---|
| **orquestador** | el bucle de 4 líneas · cómo lanzar en su harness | el flujo · los estados · los skills |
| **`sf`** | dónde estás · qué sigue · si podés avanzar · qué servir | cómo se hace nada |
| **skill** | cómo se hace **su** paso | en qué paso está · qué modelo lo corre |
| **el que trabaja** | — | — · hace |
| **Javier** | decide en ⑥ ⑧ ⑰ y en las dos ⏸ | — |

---

## 7. Lo que queda

**El diseño está completo.** Lo que sigue ya no es diseñar:

1. **Brownfield** — sigue aparcado (`que-sobrevive.md` §5). Decide el destino de
   `sf onboard scan` y de la detección de `sf-init` Step 2. **No bloquea nada.**
2. **La construcción**, y su primera pregunta ya tiene respuesta: `roadmap.json` +
   `estado.json` + `sf next` son la columna. La segunda —podar el CLI actual o partir de
   cero— sigue abierta.
