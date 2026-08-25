# Primeros pasos — una vuelta completa

Esta página recorre **un producto entero, de la idea al código archivado**. No es un resumen: es
lo que ves en pantalla, paso por paso, con lo que hace falta entender en cada parada.

**Antes de empezar:** [`../INSTALL.md`](../INSTALL.md).

---

## Cómo se lee esto

Hay dos formas de usar SpecForge y conviene entenderlas antes:

| | Quién tipea | Cuándo |
|---|---|---|
| **con el agente** | el agente corre `sf next` y sigue el bucle solo | siempre, es el modo normal |
| **a mano** | vos corrés cada comando | para entender qué pasa, o para destrabar algo |

**Este recorrido está escrito a mano** para que veas la máquina. En el uso real le decís a tu
agente *"seguí"* y hace todo esto solo — el `CLAUDE.md` que puso `sf install` le dice cómo.

---

## ① Arrancar

```bash
mkdir mi-producto && cd mi-producto
git init
sf install
sf init
```

```
+ .docs/
+ .docs/backlog/
+ .docs/constitucion.md
+ .docs/estado.json

⚠ No reconocí el stack: completá `test_cmd:` en la constitución.
```

En un repo vacío no hay manifiesto todavía, así que no puede detectar nada. **No lo arregles
ahora**: el ⑧ lo va a llenar cuando decidas el stack.

```bash
sf next
```

```
estado:   brief
skill:    sfp-scout
modelo:   sonnet
via:      vos

el ⑥ lo sellás vos cuando esté

  sf context
  sf done
```

Tres cosas que ya aparecen acá y valen para todo el recorrido:

- **`skill:`** es qué le pedís al agente. `sf` te lo dice para que el orquestador no tenga que
  saberse la tabla `estado → skill`.
- **`via: vos`** significa *trabajás de frente*, sin subagente. Es el **único** estado así, y hay
  motivo: el brief es un pinponeo, y **un subagente no te habla** — arranca, trabaja y muere.
- **`sf context` y `sf done`** los corre **el que trabaja**, no vos. Acá coinciden porque el que
  trabaja sos vos.

---

## ② El brief — de una idea difusa a un veredicto

Le decís a tu agente: **"corré `sfp-scout`"**.

El skill pide su sobre:

```bash
sf context
```

```
# Sobre — brief

## Sin sobre
El brief arranca de cero: es un pinponeo con Javier, no una lectura.
```

**Es el único sobre vacío del flujo, y está bien que lo sea.** Todavía no hay nada escrito que
leer — ése es el punto de arrancar.

El skill entonces: te entrevista hasta que la idea tenga forma, busca si ya existe, encuentra el
hueco, y estresa los supuestos. Escribe `.docs/brief.md` y para:

```
🛑 ⑥ — brief ready: "un CLI que ordena descargas por tipo"
Proposed: hacelo
Evidence: retrieved 7 / model-prior 2  ·  Differentiator: …
Awaiting: sf approve  /  sf reject "motivo"
```

**Acá decidís vos.** Es la primera de tus tres decisiones.

```bash
sf approve
```

```
✓ brief sellado: hacelo
```

> **`sf` no elige el veredicto:** lo escribiste vos en el frontmatter del brief, y `sf approve`
> significa *"sí, sellalo con lo que dice"* — incluso si dice `no-lo-hagas`. **Poder decir que no
> es el valor del ⑥.**

---

## ③ El PRD y la constitución

```bash
sf next
```

```
estado:   prd
skill:    sfp-po
modelo:   sonnet
via:      subagente

sin parada: del ⑦ se pasa derecho al ⑧
```

**`via: subagente`**, y de acá en adelante casi todo es así. Le decís a tu agente que lance uno
fresco con `sfp-po`. El subagente corre `sf context`, escribe `.docs/prd.md`, corre `sf done` y
se muere.

Después sigue el ⑧, con su propio subagente:

```
estado:   constitucion
skill:    sfp-constitucion
modelo:   sonnet
via:      subagente
```

> **`sf init` ya dejó la constitución a medias, y eso no la da por escrita.** Llenó la cabecera
> técnica —`lenguaje`, `manifiesto`, `test_cmd`— porque eso se detecta mirando el manifiesto. El
> cuerpo lo dejó con un marcador, y **ese marcador es el checkpoint**: mientras esté, `sf next` te
> manda al ⑧; cuando el ⑧ lo reemplaza, aparece la parada.

Y el ⑧ es distinto de los otros dos subagentes: **para**.

```
🛑 ⑧ — constitution written
Stack: Go 1.22 · CLI sin dependencias  ·  test_cmd: go test ./...
Awaiting: sf approve  /  sf reject "motivo"
```

**Ésta es tu segunda decisión**, y la constitución es el artefacto bisagra: la leen el ⑨, **el
implementador (es su manual)**, el revisor y `sf` mismo.

```bash
sf approve
```

**`sf approve` corre la compuerta antes de sellar**, así que si el `test_cmd` quedó vacío no sella:

```
✗ No pude.
  · la constitución no tiene `test_cmd:` — sin eso sf no puede correr los tests
```

Vale para las cinco paradas. No convierte a `approve` en `done` —`done` **mueve**, `approve`
**sella lo que decidiste**—: lo único que cambia es que ya no podés sellar algo que la máquina
sabe que está roto.

### Si algo salió mal

```bash
sf reject "el stack está bien pero falta la regla de errores"
```

El motivo **no se imprime y ya**: se guarda, y **viaja primero en el sobre** del que rehaga. Sin
eso, un subagente nuevo vuelve a proponer exactamente lo mismo.

---

## ④ El backlog y el roadmap

```bash
sf next     # → backlog, sfp-backlog
```

El ⑨ parte el PRD en historias con **criterios que tienen id**:

```markdown
## Criterios de aceptación
- **CA-1** — acepta --json y devuelve el estado serializado
- **CA-2** — si no hay estado, sale con código 1 y mensaje
```

> **Los ids son el mecanismo entero contra el *"terminado"* con media historia.** Con criterios
> en prosa, el revisor contesta *"anda"*. Con `CA-1 / CA-2 / CA-3` tiene que contestar **uno por
> uno**, y `sf` cuenta. **Contar no es juzgar** — es la única parte que una máquina puede hacer, y
> alcanza.

Al terminar aparece una parada distinta:

```
⏸ Salieron las historias. Mirá si querés y seguí.

  sf approve
```

**Es una ⏸, no una 🛑.** No es una de tus tres decisiones: es *"mirá esto, que acá un error sale
barato"*. Un `sf approve` y sigue.

Después el ⑩ agrupa las historias en features y las ordena. **Sólo ids y orden** — no agrega
contenido.

---

## ⑤ Planificar una feature

```bash
sf next
```

```
Sigue f-1 — núcleo del CLI (2 historias).

  sf take f-1
```

```bash
sf take f-1
sf next
```

```
estado:   planificacion
feature:  f-1
skill:    sf-plan
modelo:   opus
via:      subagente
```

**`modelo: opus`**, y no es casual: planificar es donde el contexto grande es un **activo**. El
subagente hace los cinco pasos en **una sola pasada** —tres opciones, la spec, las tareas, los
tests, el modelo— y escribe tres archivos.

Cuando termina:

```
🛑 ⑰ — f-1 planned
Option chosen: B — un pool de workers
7 tasks in 3 batches  ·  9 criteria, all covered
Awaiting: sf approve  /  sf reject "motivo"  /  sf take <other feature>
```

**Tu tercera y última decisión.** Y tiene **tres** puertas, no dos: la tercera es *"aprobalo pero
andá a planificar otra primero"*.

### Qué comprobó `sf` antes de dejarte llegar acá

Cinco cosas, **y ninguna es una opinión**:

```
✓ los tres archivos están
✓ decision.md tiene exactamente 3 opciones
✓ cada lote tiene al menos un test planificado
✓ los 9 criterios de las historias están cubiertos por alguna tarea
✓ ninguna tarea dice satisfacer un criterio que no existe
```

> **Acá empieza a morir el *"terminado"* con media historia, y todavía no se escribió una línea
> de código.** Si la feature tiene 9 criterios y las tareas cubren 7, nace a medias — y `sf` te lo
> dice ahora, no en la revisión.

```bash
sf approve
```

---

## ⑥ Implementar — donde `sf` es más duro

```bash
sf next
```

```
estado:   implementar
feature:  f-1 · lote 1 de 3
skill:    sf-build
modelo:   sonnet
via:      subagente

escribí los tests del lote y confirmá el rojo antes de implementar

  sf context
  sf lote start
```

El subagente pide su sobre y recibe **sólo el lote 1**, no la feature entera. Escribe los tests
que el plan nombró, y entonces:

```bash
sf lote start
```

Y acá `sf` hace tres cosas y **se niega si los tests no están honestamente en rojo**:

```
① el test planificado no existe        ✗ y te dice cuál
② el test existe pero YA PASA          ✗ "un test que pasa antes de que exista
                                           el código es un test de mentira"
③ el test falla de verdad              ✓ crea la branch · rojo · guarda el hash
```

Con el rojo confirmado, implementa. Y para cerrar:

```bash
sf done --msg "feat: el pool de workers"
```

Antes de commitear, `sf` comprueba dos cosas:

```
✓ los tests pasan            — los corre él, no te cree
✓ los archivos de test NO CAMBIARON entre el rojo y el verde
```

> **La segunda no es desconfianza personal.** La forma más barata de hacer pasar un test que falla
> es **aflojar el test**, y suele pasar sin que nadie lo decida. El hash lo hace visible.

Si todo está bien: **commit**, y el lote 2.

> **Cerrar el lote ES commitear.** Sin `--msg` no hay `done`. Así *"el commit no se hace"* y
> *"commits mal agrupados"* dejan de ser detectables **para ser imposibles**: el agrupamiento ya
> se decidió al planificar, un lote = un commit.

---

## ⑦ Revisar y cerrar

```bash
sf next     # → revision, sf-check, opus
```

El revisor es **un subagente fresco con un modelo grande, que no implementó esto**. Recibe el
diff, los criterios y la corrida de mutantes, y escribe `revision.json` con **un veredicto por
cada criterio**.

```
✗ No avanzo.
  · la feature tiene 9 criterios y el informe opina sobre 7.
    Sin veredicto: us-3/CA-4 · us-3/CA-7
```

> `sf` **no juzga la revisión**: comprueba que **el juicio haya ocurrido, sobre todos**.

Con un hallazgo abierto, vuelve a `implementar`. Cuando queda limpio:

```bash
sf next     # → cierre, sf-cierre
```

El ㉓ escribe la doc y el journal. Y después:

```
⏸ f-1 lista para archivar. La doc y el journal están.

  sf approve
```

```bash
sf approve
```

```
✓ f-1 archivada y cerrada
```

Eso hizo **cuatro cosas mecánicas**: commiteó lo que quedaba pendiente, mergeó la branch con el
modo de tu constitución y la borró, movió la carpeta a `.docs/archivado/`, y marcó la feature
cerrada — dejando el archivado en un commit propio.

> **El git va antes que el movimiento**, aunque se lea al revés: mover la carpeta es lo único
> irreversible. Si el merge conflictúa, no se movió nada y `sf approve` se reintenta tal cual.

```bash
sf next     # → Sigue f-2 …
```

Y vuelve a empezar. Cuando no queda ninguna:

```
🎉 Todas las features del roadmap están cerradas.
```

---

## Agregar algo después

Un producto vivo no vuelve al brief. **Las tres entradas convergen en el backlog: el backlog es
el embudo.**

```bash
sf new "el login rompe con email en mayúsculas"
```

```
us-7 creada. Ahora se pinponea: `sf next` te lleva.
```

`sf next` te manda a `sfp-backlog` a completarla. Si es un **bug**, el skill compone `sfx-triage`
para encontrar la causa raíz, marca `tipo: bug` y llena `relacionado_a` con el `us-#` original.

> **`tipo: bug` es lo que rutea.** Un bug **saltea planificación y revisión** y va derecho
> `implementar → cierre`. No hay carril paralelo ni segunda máquina: **hay un campo que saltea dos
> estados**, y por eso el rastro no se pierde.

Y como el bug rompió algo ya archivado, **el sobre del que lo arregla trae la spec archivada de
esa feature** — sin eso, arrancaría de cero sin saber qué se había decidido, que es justo donde
nacen los mocks.

---

## Cuando el bucle patina

Si `sf done` falla tres veces seguidas:

```
⚠ ME TRABÉ. implementar falló 3 veces seguidas con sonnet.
```

**Es una parada de seguridad y no es configurable:** sin ella, el bucle *"`sf` da rojo → el
orquestador relanza"* no termina nunca. Tenés tres salidas:

```bash
sf model opus                    # subí el modelo — resetea el contador
sf dismiss h-1 "falso positivo"  # el hallazgo no era real
                                 # o entrás vos y lo arreglás a mano
```

---

## Ver dónde está todo

```bash
sf status
```

Es **el único comando pensado para vos y no para el agente**. No mueve nada ni comprueba nada:
lee y arma una vista.

---

## Revisar lo construido

Cada tanto, o cuando cerraste varias features:

```bash
sf audit
```

Verifica **punta a punta** que lo construido satisfaga las historias — y encuentra lo que una
revisión por feature **estructuralmente no puede ver**: que `f-4` rompió un criterio de `f-1`, que
dos features no se integran, o que una afirmación perdió su respaldo:

```
✗ us-1/CA-2 se dio por cumplido en f-1 y su test ya no existe
```

Más en [`comandos.md`](comandos.md#sf-audit).

---

## Qué leer ahora

- **[`comandos.md`](comandos.md)** — los 16 comandos con sus flags.
- **[`problemas.md`](problemas.md)** — qué hacer con cada ✗ que te tire `sf`.
- **[`estados.md`](estados.md)** — qué exige la compuerta de cada estado.
