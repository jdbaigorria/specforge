# Los nueve estados

Cinco corren **una vez por producto**, cuatro **una vez por feature**.

```
PRODUCTO   brief → prd → constitución → backlog → roadmap
FEATURE    planificación → implementar → revisión → cierre
```

Y hay dos estados de espera donde la feature existe pero nadie está trabajando: `planificada`
(plan aprobado, esperando en la cola) y `cerrada` (terminada y archivada).

> **No existe un estado "pendiente".** Una feature que está en el roadmap y todavía no arrancó
> simplemente **no aparece** en el estado. Lo que se puede deducir, no se guarda.

---

## Cómo leer esta página

Cada estado tiene cuatro cosas y las cuatro importan:

| | |
|---|---|
| **el sobre** | qué recibe el que trabaja — se lo da `sf context` |
| **produce** | qué archivo aparece cuando terminó |
| **la compuerta** | **qué comprueba `sf` antes de dejar avanzar**. Nunca una opinión |
| **la parada** | 🛑 decidís vos · ⏸ mirá y seguí · nada |

---

# Los cinco de producto

## ① `brief` — de la idea a un veredicto

```
skill    sfp-scout          via: vos          (el único que conversa)
sobre    VACÍO — es un pinponeo, no una lectura
produce  .docs/brief.md · .docs/entrevista.md · .docs/evidencia.md
         (+ .docs/vocabulario.md, sólo si alguna palabra hizo falta definirla)
parada   🛑 el ⑥ · lo sellás vos, y con un acta delante
```

**Es el único `via: vos` del flujo, y hay motivo:** un subagente arranca, trabaja y muere — **no
te habla**. Los pasos ①–⑤ son una conversación donde la idea toma forma.

**La compuerta:**

```
✓ existe .docs/brief.md, y trae un veredicto válido: hacelo | pivotea | no-lo-hagas
✓ existe .docs/entrevista.md, y declara  abiertas: 0
✓ existe .docs/evidencia.md, y cita al menos UNA fuente con link
  ↳ salvo que declare `evidencia: baja`, que es una pasada degradada declarada a propósito
```

Y como toda parada, **entrega un acta**: lo que comprobó, los números crudos que midió, y —el
bloque que más importa— **lo que no pudo comprobar**.

**No comprueba que el brief sea bueno** — eso es juicio, y el juicio es tuyo. Es la diferencia
entre una compuerta y un juez.

> **`no-lo-hagas` es tan válido como los otros dos**, y si lo sellás así el flujo termina ahí. Es
> la única vez que el flujo se cierra sin haber construido nada, y está bien que exista: **el
> valor del ⑥ es poder decir que no.**

## ② `prd` — qué hay que poder hacer

```
skill    sfp-po             via: subagente
sobre    el brief
produce  .docs/prd.md
parada   ninguna — del ⑦ se pasa derecho al ⑧
```

**La compuerta:** que el archivo exista. Nada más — el ⑦ no tiene punto de decisión.

`sf` guarda un `prd_hash`. Si el PRD cambia después de que se escribieron las historias, esas
historias quedaron mirando otro producto. **Avisa, no frena.**

## ③ `constitucion` — las reglas del proyecto

```
skill    sfp-constitucion   via: subagente
sobre    el PRD + la constitución (con la cabecera YA llena por sf init)
produce  .docs/constitucion.md
parada   🛑 el ⑧ · la sellás vos
```

**El artefacto bisagra:** el primero cuyo lector principal no es un humano ni esta conversación,
sino **un subagente frío**. Lo leen el ⑨, el implementador (es su manual), el revisor y `sf`.

**La compuerta es una sola línea:**

```
✓ test_cmd:  no está vacío
```

Y no es pedantería: **sin `test_cmd` caen tres compuertas** — el rojo, el verde y el conteo de
tests. **`sf approve` la corre antes de sellar**, así que la constitución sin `test_cmd` no se
sella: es la única línea del frontmatter que se exige.

> **`sf init` deja la constitución escrita a medias, y eso es a propósito.** Llena la cabecera
> técnica —`lenguaje`, `manifiesto`, `test_cmd`— porque eso se detecta mirando el manifiesto y
> tiene una sola respuesta correcta. El cuerpo lo deja con un marcador:
>
> ```
> <Esto lo escribe el ⑧. Corré: sf next>
> ```
>
> **Ese marcador es el checkpoint del ⑧.** El brief y el PRD se distinguen por si el archivo
> existe; la constitución no puede, porque `sf init` siempre la crea. Mientras el marcador esté,
> `sf next` te manda a escribirla; cuando el ⑧ lo reemplaza, aparece la 🛑.

**Es puramente técnica.** Cero principios abstractos: el *qué* y el *por qué* ya están escritos y
sellados arriba, en el brief y el PRD. Repetirlos acá crearía una segunda copia que se
desincroniza.

## ④ `backlog` — las historias

```
skill    sfp-backlog        via: subagente
sobre    el PRD + la constitución
produce  .docs/backlog/us-#.md
parada   ⏸ mirá si querés y seguí
```

**La compuerta**, y la corren tanto `sf done` como el `sf approve` de la ⏸:

```
✓ hay al menos una historia
✓ TODAS tienen criterios con id:  CA-1, CA-2, …
✓ y los criterios DICEN algo
```

La tercera existe por el esqueleto de `sf new`, que deja la línea puesta y vacía —`- **CA-1** —`—.
Contando sólo ids, ese esqueleto pasaba. Y **un criterio que no dice nada es peor que ninguno**: el
⑰ lo da por cubierto y el ㉑ le pone veredicto, así que el mecanismo entero queda en pie sobre algo
que nadie puede juzgar.

> **El título NO se exige.** La compuerta pregunta si el mecanismo se sostiene, y lo que el ⑰
> cuenta y el ㉑ juzga son los criterios. De la historia a medio escribir se ocupa el checkpoint,
> abajo.

**La ⏸ es barata, la compuerta no.** Que mirar el backlog sea un enter no significa que se pueda
sellar cualquier cosa: un backlog sin ids desarma el mecanismo entero — la cobertura del ⑰ cuenta
cero contra cero y pasa, y el conteo de veredictos del ㉑ también.

> ### El ⑨ corre dos veces, y el checkpoint lo sabe

La primera es partir el PRD. Las otras son completar lo que metió `sf new`, y ahí la pregunta no
puede ser *"¿hay historias?"* — con un producto en marcha la respuesta es siempre sí, y la ⏸
saldría con el esqueleto sin completar adentro.

Así que el checkpoint pregunta **qué falta**: una historia cuenta como pendiente si no tiene
criterios, si alguno está vacío, o si todavía tiene los huecos que dejó `sf new` (`<título>`,
`<quién>`, …). Es el mismo marcador que usa el ⑧, y por el mismo motivo — atarlo a `titulo:` daría
falsos positivos con una historia escrita a mano que lleva el título sólo en el encabezado.

**Los ids son el mecanismo entero contra el "terminado" con media historia.** Con criterios en
> prosa el revisor contesta *"anda"*. Con ids tiene que contestar **uno por uno**, y `sf` cuenta:
> *"la historia tiene 3 criterios y el informe habla de 2"*. **Contar no es juzgar** — es la única
> parte que una máquina puede hacer, y alcanza.

Es una **⏸ y no una 🛑**: no es una de tus tres decisiones, pero es donde **un mal corte es más
barato de arreglar** — todavía es una línea de texto.

## ⑤ `roadmap` — el orden

```
skill    sfp-roadmap        via: subagente
sobre    todas las historias + la constitución
produce  .docs/roadmap.json
parada   ninguna
```

**Ordena y agrupa. No agrega contenido** — sólo ids y orden.

**La compuerta:**

```
✓ el archivo se puede leer
✓ TODA historia está en alguna feature
⚠ una feature con 6+ historias:  "¿la partís?"   (avisa, no frena)
```

> El segundo chequeo atrapa el error caro: **una historia que quedó fuera de todas las features no
> la va a implementar nadie, y nadie se entera.**

Una feature es *"un conjunto de historias que **comparten solución técnica** y se entregan
juntas"*. Ese criterio se eligió porque **se puede violar**: si tres historias necesitan tres
soluciones distintas, el agrupamiento está mal.

Y el ⑩ **vuelve a hacer falta** cada vez que `sf new` deja una historia huérfana.

**El `sf done` del ⑩ no mueve ningún sello**, y es correcto: *"¿existe el `roadmap.json`?"* se
deduce del archivo, así que no hay nada que guardar. Lo que sigue —elegir cuál feature— es tuyo:

```
$ sf done
⚠ f-1 junta 6 historias. ¿La partís?
✓ listo
→ el roadmap está listo. Elegí con `sf take <feature>`.
```

---

# Los cuatro de feature

## ⑥ `planificacion` — el más grande

```
skill    sf-plan            via: subagente · modelo: opus
sobre    la constitución + los us-# de la feature + los journals archivados
produce  decision.md · spec-design.md · tareas.json
parada   🛑 el ⑰ · con TRES puertas
```

**Cinco pasos, una sola conversación, un subagente con contexto amplio, una sola revisión al
final.** No se parte, y el argumento es concreto: el ⑬ aprovecha acordarse de las dos opciones que
el ⑫ descartó.

**La compuerta son seis chequeos, y ninguno es una opinión:**

```
① los tres archivos existen
② decision.md tiene EXACTAMENTE 3 opciones      (## A — … / ## B — … / ## C — …)
③ cada lote tiene al menos un test planificado
④ cada tarea nombra al menos un test POR CRITERIO que promete
⑤ cada criterio de las historias está cubierto por alguna tarea
⑥ ninguna tarea dice satisfacer un criterio que no existe
```

**El ④ existe porque un piso se vuelve techo.** Con sólo el ③, la tarea nombraba el camino feliz
y nada más; todo lo demás aparecía tres vueltas de revisión después, cuando el ㉒ lo encontraba.
No dice nada sobre la calidad del test —`sf` no lo puede leer— pero una tarea que promete cuatro
criterios y nombra un test nace corta, y eso sí se cuenta.

Se cuenta **por tarea** y no por lote: el mensaje dice cuál, y una tarea generosa no puede tapar
a la tacaña de al lado.

**El ② es el anti-alucinación más barato del flujo:** una sola opción escrita como si fuera una
comparación es la forma que toma una corazonada segura de sí misma. No puede contar como tres si
nunca se pensaron tres.

**El ④ y el ⑤ son espejos**, y atrapan errores opuestos: una feature que nace a medias, y una
tarea que miente sobre qué satisface.

> **Acá empieza a morir el "terminado" con media historia — y todavía no se escribió una línea de
> código.**

**Las tres puertas del ⑰:**

```bash
sf approve                 # a implementar
sf reject "motivo"         # rehacé el bloque entero
sf take f-3                # aprobado, pero planificá otra primero
```

Nadie entra ni sale por el medio: la revisión es del bloque entero, y *"pido cambios"* vuelve al
bloque entero.

## ⑦ `implementar` — donde `sf` es más duro

```
skill    sf-build           via: subagente · UNO POR LOTE
sobre    la constitución + spec-design.md + SÓLO tu lote + los criterios
         (+ la spec archivada de lo que rompiste, si es un bug)
produce  código, tests y un commit por lote
parada   ninguna — salvo ME TRABÉ
```

**`decision.md` NO va en el sobre, a propósito.** La decisión ya se tomó; darle dos diseños
rechazados al implementador es una invitación a improvisar.

**Un subagente fresco por lote**, y es el único punto del flujo donde el daño se puede **prevenir
en vez de detectar**: los mocks salían *"con el contexto al 50%"*.

**Hay dos compuertas, y están en momentos distintos:**

**`sf lote start`** — antes de implementar:

```
✓ el test que el plan nombró EXISTE
✓ la suite FALLA de verdad
✓ y el fallo NOMBRA al menos uno de los tests del lote
→ crea la branch · marca el rojo · guarda el hash de los archivos de test
```

> **El tercero tapa el hueco que quedaba entre los otros dos.** El exit code dice que *algo* falló;
> no dice que haya fallado lo tuyo. Con una suite que **ya venía roja** —un test viejo que quedó
> fallando, un paquete que no compila— el lote recibía el rojo de regalo: los tests planificados
> podían no haberse ejecutado nunca, y el hash se tomaba igual. Toda la cadena rojo→verde quedaba
> apoyada sobre un fallo ajeno.
>
> No es un parser: es el **mismo `Contains`** que ya comprueba que el test exista adentro del
> archivo, apuntado a la salida. Todos los runners nombran lo que falla — ninguno dice *"falló un
> test"* sin decir cuál.
>
> **Alcanza con que nombre uno**, porque es normal que el runner corte en el primero. Y si te
> frena sin motivo, la causa es una sola: un `test_cmd` que recorta la salida (`| tail`, `-q`).
> Sacale el recorte — esa salida la leen también el ㉑ y el que tiene que arreglar.
>
> **En un lote SIN plan la pregunta no se hace**, porque no hay lista contra la cual comparar: el
> camino corto de un bug y la vuelta del ㉑ vuelven a tener el exit code como todo lo que hay.

**`sf done --msg`** — para cerrar:

```
✓ hay mensaje de commit          sin él no se cierra el lote
✓ los tests PASAN                los corre sf, no te cree
✓ los archivos de test NO CAMBIARON entre el rojo y el verde
→ commit
```

> **El tercero no es desconfianza personal.** La forma más barata de hacer pasar un test que falla
> es **aflojar el test**, y suele pasar sin que nadie lo decida. El hash lo hace visible.

## ⑧ `revision` — donde muere el "terminado" a medias

```
skill    sf-check           via: subagente · modelo: opus
sobre    la constitución + spec-design.md + los criterios + el diff
         + la corrida de mutantes
produce  revision.json
parada   ninguna — vuelve a implementar si hay hallazgos
```

**La compuerta:**

```
✓ CADA criterio tiene un veredicto en el informe
✓ CERO hallazgos abiertos
✓ si algún mutante resucitó, hay al menos un hallazgo del ㉒
⚠ si es la vuelta 3 o más: "la planificación se quedó corta"
```

> `sf` **no juzga la revisión**: comprueba que **el juicio haya ocurrido, y sobre todos**. Si la
> feature tiene 9 criterios y el informe opina sobre 7, no avanza — y dice cuáles faltan.

**Los hallazgos no tienen ciclo de vida**, y eso simplifica todo: la revisión **se rehace
entera**, así que `h-1` simplemente **no reaparece** — *eso* es estar arreglado. Nadie marca nada.

Queda **un solo estado que no se deduce**: `descartado`, y lo ponés vos con `sf dismiss`.

### Con un hallazgo abierto: la vuelta abre un lote

Volver a `implementar` sin más sería mandar al que arregla a un estado sin trabajo. Así que la
vuelta **abre un lote nuevo**, y el arreglo pasa por el mismo carril que todo lo demás:

```
✗ No avanzo.
  · hay 1 hallazgos abiertos: h-1
→ hay hallazgos abiertos — vuelve a implementar en el lote 2
```

Ese lote **no está en `tareas.json`** —el plan se escribió antes de que el hallazgo existiera—, y
eso cambia dos cosas:

- **el sobre trae el hallazgo** en vez de las tareas del plan: el id, el criterio y el detalle,
  embebidos, porque el que los lee es un subagente nuevo que no estuvo en la revisión;
- **`sf lote start` afloja a lo comprobable**: no puede exigir *cuáles* tests, pero **sigue
  exigiendo que la suite falle**. Un hallazgo sin un test que lo reproduzca es un hallazgo que
  nadie va a poder verificar.

Y el arreglo termina en `sf done --msg "…"`, o sea en un commit: **un lote, un commit**, también
acá.

### Y a la tercera vuelta, el bucle tiene fondo

Ésta es la única transición que va **para atrás**, y sin tope no termina: el revisor encuentra, el
implementador no arregla, el revisor encuentra. Al llegar a **tres rondas**, `sf next` levanta una
parada en vez de seguir proponiendo trabajo:

```
⚠ ME TRABÉ EN LA REVISIÓN. f-2 volvió del ㉑ 3 veces y los hallazgos siguen abiertos.
   Adjudicá uno por uno: ¿es portante, o no?
   h-1 el reintento no cubre el timeout del socket
```

**Es el hermano de ME TRABÉ sobre el otro bucle, y las salidas son distintas a propósito:**

| | qué falló | qué lo destraba |
|---|---|---|
| **ME TRABÉ** | 3 `sf done` en ✗ seguidos | más modelo: `sf model <alias>` |
| **ME TRABÉ EN LA REVISIÓN** | 3 vueltas del ㉑ | adjudicar: `sf dismiss <h-#> "motivo"` |

En el primero el problema es **capacidad**, y otro intento con un modelo mejor lo puede resolver.
En el segundo es un **desacuerdo** entre el revisor y el implementador sobre el mismo hallazgo — y
eso no lo rompe otro intento, lo rompe alguien que decida si el hallazgo es portante. Por eso el
mensaje lista los hallazgos con su id en vez de contar fallas: la salida es `sf dismiss`, y sin
los números habría que ir a abrir `revision.json` para saber qué escribir.

El contador se resetea cuando la revisión sale limpia, con `sf model` y con `sf dismiss` — los
tres son *"el bucle se destrabó"*. Y **no mueve el estado**: la feature se queda en `revision`,
que es donde la parada la va a encontrar. Moverla a `implementar` y parar después dejaría al
estado diciendo *"implementá"* mientras la máquina dice que no.

> **Hay dos contadores de vueltas y no son el mismo.** `revision.json` lleva un `vuelta`, y
> `estado.json` lleva `rondas_revision`:
>
> | | quién lo escribe | qué es |
> |---|---|---|
> | `vuelta` (en `revision.json`) | el revisor | una **declaración**: por cuál vuelta cree que va |
> | `rondas_revision` (en `estado.json`) | `sf` | un **hecho**: cuántas veces contó él mismo |
>
> La parada sale del segundo, y tenía que salir de ahí: la regla dura es que el estado avanza —y
> frena— con hechos comprobados, nunca con la palabra del que trabajó. Un revisor que se olvida de
> incrementar su `vuelta`, o que arranca de cero porque es un subagente nuevo, no desarma el freno.

## ⑨ `cierre` — la doc y el aprendizaje

```
skill    sf-cierre          via: subagente
sobre    el diff (la mitad TÉCNICA) + los us-# (la mitad FUNCIONAL)
         + spec-design.md + los journals anteriores
produce  doc.md · journal.md
parada   ⏸ lista para archivar
```

**El sobre trae dos fuentes para un documento, y es la clave del estado:** un documentador que
sólo ve el código escribe una descripción exacta de *qué existe* y **no puede decir por qué
alguien lo quería**. Esa mitad es la que un humano lee seis meses después.

**La compuerta:** que existan los dos archivos. Son dos y no uno — **la doc y el journal**.

Y después de la ⏸, `sf approve` hace **cuatro cosas mecánicas**, en este orden:

```
commitea lo que quedaba pendiente   → "chore: cierre de f-#"
mergea la branch y la borra         → con el modo de tu constitución
mueve la carpeta entera             → .docs/archivado/f-#-slug/
marca la feature cerrada            → "chore: f-# archivada"
```

**El orden es al revés de lo que parece, y a propósito.** Mover la carpeta es lo único
irreversible; el git es lo único que puede fallar. Con el movimiento primero, un merge conflictivo
te dejaba la carpeta archivada, la feature sin cerrar y el ㉓ apuntando a archivos que ya no
estaban ahí — y no se salía, porque el segundo intento fallaba al mover algo que ya no existía.
Así, un merge que falla **no movió nada** y `sf approve` se reintenta tal cual.

**Y empieza commiteando** porque `.docs/estado.json` está sucio *por construcción* cuando llegás
acá: `sf` lo escribe en cada transición y sólo lo commitea al cerrar un lote. Sin ese commit,
`git checkout` aborta **siempre**. De paso, ese trabajo —la revisión, la doc, el journal— deja de
quedar huérfano.

El archivado termina en **un commit propio**. Si quedara suelto, el `git add -A` del primer lote
de la feature siguiente se lo llevaría puesto, y el cierre de `f-1` terminaría adentro del commit
de `f-2` — que es el dolor #3 entrando por la única puerta que quedaba abierta.

---

## Los tres caminos

Cuánto proceso pide una feature lo declara el `tipo:` de sus historias:

```
tipo: us      →  planificacion → implementar → revision → cierre
tipo: chico   →                  implementar → revision → cierre     saltea 1
tipo: bug     →                  implementar →            cierre     saltea 2
```

**No es un carril paralelo ni una segunda máquina** — un carril paralelo sería una segunda máquina
que mantener. Es **un campo que saltea estados**, y por eso el rastro no se pierde: los tres
entraron por el backlog, que es el embudo.

Se exige que **todas** las historias de la feature pidan lo mismo, no *"alguna"*, y ante la mezcla
**gana el camino más largo**: saltear la planificación de una feature que mezcla un bug con dos
historias nuevas dejaría esas dos sin diseño. Un bug y un chico juntos dan chico — el bug no
necesita revisión, pero tampoco la estorba.

### `chico` mide el repo, no tu confianza

El del medio es el que más fácil se usa mal. **`chico` es un cambio acotado sobre un flujo que ya
está escrito acá**: un flag más en un comando que existe, un campo más en un endpoint que existe,
una validación que falta.

**Entender de qué clase de app se trata NO alcanza.** Si no hay un flujo acá para ir a leer, no es
chico: es largo, aunque suene simple. Un proyecto nuevo no tiene ningún flujo, así que **nada en
él es chico**.

> Y el que declara chico para ahorrarse la planificación ya contestó que no lo es: **buscar la
> etiqueta liviana ES la duda**. Ante la duda, el largo — de más se puede saltear después; de
> menos ya se implementó sin diseño.

### El trinquete: sube y no baja

La complejidad escondida aparece **implementando**, o sea después de que el camino ya se eligió.
Cuando pasa, el que está adentro del ⑱ corre:

```bash
sf ampliar "resultó que toca el parser, no sólo el flag"
```

La feature vuelve al ⑫ con el motivo puesto, y **no vuelve a bajar nunca**: el campo `ampliada`
del estado no se apaga desde ningún lado — ni editando la historia de vuelta a `tipo: chico`, ni
corriendo el comando dos veces. El mismo optimismo que erró la clasificación la primera vez no
tiene una segunda oportunidad de errarla.

Los lotes ya cerrados quedan: su commit existe y su rojo se vio. Ver
[`sf ampliar`](comandos.md#sf-ampliar-f--motivo).

---

## Las cinco paradas

| | Cuándo | Cómo se destraba |
|---|---|---|
| **🛑 decisión** | ⑥ el brief · ⑧ la constitución · ⑰ el plan | `sf approve` / `sf reject` / `sf take` |
| **⏸ barata** | ⑨ el backlog · ㉓ el cierre | `sf approve` |
| **⚠ aviso** | dependencia nueva · el plan que envejeció · 6+ historias | nada: seguís |
| **ME TRABÉ** | 3 `sf done` fallidos seguidos | `sf model` · entrás vos |
| **ME TRABÉ EN LA REVISIÓN** | 3 vueltas del ㉑ con hallazgos abiertos | `sf dismiss` · `sf model` · entrás vos |

**Las tres primeras son de gusto y son configurables. Las dos últimas no**, porque son de
seguridad: cada una le pone fondo a un bucle que sin ella no termina nunca.

```
sf da rojo → el orquestador relanza → sf da rojo …            ← ME TRABÉ
el ㉑ encuentra → vuelve a implementar → el ㉑ encuentra …     ← ME TRABÉ EN LA REVISIÓN
```

---

## Y ninguna compuerta piensa

Es la restricción que ordena el diseño entero:

> **Una compuerta frena sobre un HECHO; un juez OPINA.**
>
> `sf` puede frenarte porque el test falló — eso no lo discute nadie. **No** puede frenarte porque
> un modelo dijo que tu diseño está flojo.

Por eso todas las compuertas de esta página son **contar una lista, comparar dos strings, o
preguntar si un archivo existe**. Nunca un criterio.
