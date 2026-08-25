# Cuando `sf` te frena

Una guía por síntoma. **Cada ✗ de `sf` es un hecho, no una opinión** — así que siempre hay algo
concreto que arreglar.

---

## Arrancar

### `sf: command not found`

`~/go/bin` no está en tu `PATH`:

```bash
export PATH="$PATH:$(go env GOPATH)/bin"
```

### `Este proyecto todavía no tiene estado.`

```bash
sf init
```

### `Este proyecto ya está iniciado.`

Corriste `sf init` dos veces. **No pasó nada malo**: no pisa nada. Seguí con `sf next`.

### `⚠ No reconocí el stack`

`sf init` no encontró un manifiesto que conozca (`go.mod`, `Cargo.toml`, `pyproject.toml`,
`package.json`, `Gemfile`, `pom.xml`, `build.gradle`).

Abrí `.docs/constitucion.md` y completá `test_cmd:` a mano. **No es opcional:** sin eso caen tres
compuertas y la constitución no sella.

### `⚠ No reconocí el harness`

```bash
sf install --harness=claude-code
```

La detección es una **pista** —una variable de entorno que puede estar o no—, así que se escribe
una vez y se corrige a mano.

---

## Las compuertas de producto

### `el brief no trae veredicto (hacelo | pivotea | no-lo-hagas)`

El frontmatter de `.docs/brief.md` tiene `veredicto: ""`. **Lo llenás vos** — es tu decisión, no
la del modelo. Escribí uno de los tres y volvé a `sf approve`.

### `la constitución no tiene test_cmd`

La línea está vacía en `.docs/constitucion.md`. Ponela.

Es la **única** línea del frontmatter que `sf` exige, porque de ella dependen tres compuertas: el
rojo del ⑲, el verde del ⑳ y el conteo de tests.

### `us-3 no tiene criterios de aceptación (CA-1, CA-2, …)`

Una historia sin criterios es **invisible** para dos estados: el ⑯ no puede contar si están
cubiertos, y el ㉑ no puede opinar sobre ellos.

Abrí `.docs/backlog/us-3.md` y agregá criterios con el formato exacto:

```markdown
## Criterios de aceptación
- **CA-1** — <disparador · estado · comportamiento>
```

**`sf` los cuenta con un patrón**, así que la forma importa: `- **CA-<número>**` al principio de
la línea.

### `us-5 no está en ninguna feature del roadmap`

Una historia huérfana **no la va a implementar nadie, y nadie se entera**. Agregala a alguna
feature de `.docs/roadmap.json`.

Pasa siempre después de un `sf new`: la historia entró al backlog pero todavía no tiene lugar en
la cola. `sf next` te manda al ⑩ solo.

---

## Las compuertas de planificación

### `decision.md tiene 1 opciones y el ⑫ pide 3`

El formato es un encabezado por opción:

```markdown
## A — un pool de workers
## B — una cola con un solo consumidor
## C — procesar en el request
```

> **Es el anti-alucinación más barato del flujo.** Una sola opción escrita como si fuera una
> comparación es la forma que toma una corazonada segura de sí misma.

Si de verdad sólo hay un camino razonable, **escribí igual los otros dos y por qué no sirven** —
eso es información valiosa para el que lea esto en seis meses.

### `el lote 2 no tiene ningún test planificado`

Cada lote tiene que declarar al menos un test en `tareas.json`. **Sin esa lista, `sf lote start`
no tiene contra qué exigir el rojo** — y ahí se cuela el "todo verde" sin tests.

### `la feature tiene 9 criterios y las tareas cubren 7. Sin cubrir: us-7/CA-1 · us-7/CA-4`

**La feature nace a medias.** Agregá tareas que satisfagan esos criterios, o —si el criterio no
corresponde a esta feature— sacalo de la historia.

Cada tarea declara qué cubre:

```json
"satisface": ["us-7/CA-1", "us-7/CA-4"]
```

### `una tarea dice satisfacer us-7/CA-9, y ese criterio no existe`

El espejo del anterior, y atrapa el error opuesto: **una tarea que miente**. Casi siempre es un
número mal tipeado o un criterio que se borró de la historia.

---

## Las compuertas de implementación

### `el test planificado no existe: core_test.go::TestSinEstado`

`sf lote start` busca **por nombre exacto** los tests que el plan declaró. O no lo escribiste, o
lo escribiste con otro nombre.

**Si el nombre del plan estaba mal**, eso es un hallazgo: decilo y que vuelva por el ⑰. No lo
renombres en silencio — `sf` está buscando *ese* nombre.

### `la suite YA PASA, y todavía no se escribió el código del lote 1`

> **Un test que pasa antes de que exista el código es un test de mentira.**

Tu test no está ejerciendo el comportamiento que dice ejercer. Casi siempre es un assert que no
asegura nada, o que prueba algo que ya existía.

### `no vi el rojo del lote 2. Corré sf lote start antes de implementar`

El orden es: **escribís los tests → `sf lote start` → implementás**. No al revés.

Si ya implementaste, no hay atajo: `sf` **no puede** dejar pasar al verde sin haber visto el rojo
con sus propios ojos, y si nunca lo vio, no pasa.

### `falta el mensaje del commit: sf done --msg "…"`

**Cerrar el lote es commitear.** No hay forma de terminar sin mensaje, y es a propósito: así *"el
commit no se hace"* deja de ser detectable **para ser imposible**.

### `los archivos de test CAMBIARON entre el rojo y el verde`

`sf` guardó el hash de los archivos de test cuando los vio fallar, y ahora no coincide.

**No es desconfianza personal:** la forma más barata de hacer pasar un test que falla es **aflojar
el test**, y suele pasar sin que nadie lo decida.

Si el test **de verdad** tenía que cambiar —el plan lo nombró mal, o el criterio se entendió
distinto—, eso es un hallazgo real: **decilo y que vuelva por el ⑰**, en vez de editar para pasar
la compuerta.

### `no hay nada que commitear: ningún archivo cambió`

El subagente dijo que terminó y no tocó un archivo. Suele ser que se quedó sin contexto o que se
confundió de directorio.

---

## Las compuertas de revisión y cierre

### `la feature tiene 9 criterios y el informe opina sobre 7`

El revisor tiene que dar un veredicto por **cada** `CA-#`, uno por uno. Faltan dos en el mapa
`criterios` de `revision.json`.

> `sf` **no juzga la revisión**: comprueba que el juicio **haya ocurrido, y sobre todos**.

### `hay 2 hallazgos abiertos: h-1 · h-3`

La feature vuelve a `implementar`. Dos caminos:

```bash
# arreglarlo → la revisión se rehace entera y el hallazgo no reaparece
# o, si es un falso positivo:
sf dismiss h-1 "el mutante no representa un caso real"
```

**No existe marcar un hallazgo como "arreglado"**: la revisión se rehace entera, y **no reaparecer
es** estar arreglado.

### `⚠ el ㉑ va por la vuelta 3. La planificación se quedó corta.`

**Es información sobre el ⑰, no sobre el implementador.** Si una feature va por la tercera
revisión, el problema suele estar en el plan, no en el código.

### `falta .docs/features/f-1-nucleo/journal.md`

El ㉓ produce **dos** archivos, no uno: la doc **y** el journal. El journal es el que va a leer el
⑫ de una feature futura.

---

## Cuando el bucle patina

### `⚠ ME TRABÉ. implementar falló 3 veces seguidas con sonnet.`

Es una parada de seguridad, y **no es configurable**: sin ella, el bucle *"`sf` da rojo → el
orquestador relanza"* no termina nunca.

**El nombre del modelo es la mitad del mensaje**, y por eso está: *"¿subo el modelo?"* no se puede
contestar sin saber cuál está fallando. Sale de la cadena completa, así que dice el que se está
usando de verdad — no el default del estado.

Tres salidas:

```bash
sf model opus                      # subí el modelo — resetea el contador
sf dismiss h-1 "falso positivo"    # si lo que traba es un hallazgo irreal
                                   # o entrás vos y lo arreglás a mano
```

`sf model` vale en los **cuatro** estados de feature, y ME TRABÉ puede aparecer en cualquiera de
ellos: la parada la dispara el contador de `sf done` fallidos, y `done` corre en los cuatro.

### `🛑 Hace falta "deepseek" y no está declarado`

`sf` **no elige el reemplazo** — eso sería opinar sobre qué modelo se parece a cuál. Decile cómo
se lanza, y queda declarado para siempre:

```bash
sf model deepseek --via consola --comando "deepseek exec"
sf model o3 --via subagente
```

---

## Avisos: no frenan, pero decilos en serio

### `⚠ f-2 junta 6 historias. ¿La partís?`

Una feature muy grande **tarda demasiado en dar la vuelta**, y su plan envejece mientras espera.
Partirla o no es criterio, y el criterio es tuyo.

### Dependencias fuera de la constitución — ⚠ TODAVÍA NO AVISA

**`sf` todavía no comprueba esto**, y conviene saberlo en vez de confiarse.

La mitad que compara ya está (`DependenciasNuevas` en `constitucion.go`) y la lista vive en
`dependencias_aprobadas:`. **Lo que falta es quien lea el manifiesto** — y es un parser por
lenguaje (`go.mod`, `package.json`, `Cargo.toml`…), que es trabajo real.

Mientras tanto lo cubre el skill: `sf-build` tiene la instrucción de **decirlo en voz alta** en su
mensaje de cierre cuando agrega una dependencia que no está aprobada.

Cuando exista, va a **avisar y no frenar**, aunque frenar sería trivial:

> La última palabra es tuya, y **una herramienta que frena sola rompe esa regla.**

### `⚠ el plan de f-2 se escribió sobre otro commit`

Planificaste `f-2`, implementaste `f-1`, y ahora la spec de `f-2` mira un repo que ya cambió.
**Avisa, no frena** — pero es exactamente donde el implementador que no encuentra lo que la spec
dice **improvisa**, y ahí nacen los mocks.

---

## Emergencias

### El estado quedó mal y quiero arreglarlo a mano

**Podés** editar `.docs/estado.json` — es un JSON legible. Pero pensalo dos veces: tres de sus
campos son **indeducibles**, y si los rompés no hay de dónde reconstruirlos:

- `rojo` — el momento en que los tests fallaban **ya pasó** y no dejó huella;
- `modelo` — la decisión que tomaste con `sf model`;
- `brief_sellado` / `constitucion_sellada` — tus aprobaciones.

Todo lo demás se deduce de los archivos, así que es recuperable.

### Quiero empezar de nuevo esta feature

Borrá su carpeta en `.docs/features/` y sacala del mapa `features` del `estado.json`. `sf next` la
va a proponer de nuevo desde el ⑫ — **el retome sale de "¿qué archivos existen?"**, así que
alcanza con que no existan.

### Quiero sacar SpecForge de este proyecto

```bash
sf uninstall     # saca CLAUDE.md y AGENTS.md si no los editaste
rm -rf .docs     # y esto es tu trabajo: borralo vos, a conciencia
```

`~/.specforge/` no se toca nunca: **tus modelos son de la máquina, no del proyecto.**
