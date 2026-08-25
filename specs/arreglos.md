# Arreglos — lo que la máquina promete y hoy no cumple

**Fecha de la auditoría:** 2026-08-25 · **Branch:** `refundation` · **Commit:** `cb4d905`

> **Estado: TERMINADO.** Los diez hallazgos y los tres menores están arreglados, cada uno con
> tests que fallan sin el arreglo (verificado revirtiendo el código y, en A2, reproduciendo el bug
> con el binario viejo sobre un repo real). **277 tests** contra los 213 del día de la auditoría.
> El detalle de cada PR está al final, en **Lo que se implementó**.
>
> El guion de humo del apéndice **está construido y corre en el CI**
> (`sf/cmd/sf/e2e_test.go`). Lo único que quedó afuera está anotado como **residuo conocido** al
> final de PR6, y es una aspereza, no un defecto.

Este documento no es diseño: es el **parte de daños**, escrito para implementarse. Cada hallazgo
trae el síntoma **reproducido de verdad** (no leído), la causa raíz con archivo y línea, el arreglo
propuesto, y el test que tiene que existir para que no vuelva.

> **Lo que está sano.** `go build`, `go vet` y `gofmt` limpios. **213 tests en 19 paquetes.** Los
> cuatro chequeos custom del CI pasan. 18 skills, todas con `SKILL.md` y `name:` igual a la carpeta.
> Ningún link markdown roto. Ningún `TODO`, stub ni `panic` de "no implementado". La documentación
> —`INSTALL.md`, `docs/primeros-pasos.md` y las seis guías— es completa y correcta **como
> descripción de lo que el diseño quiso**. El problema es la costura entre el diseño y el binario.

---

## La causa común, en una frase

> **El estado de una feature se interpreta en cinco lugares, y sólo dos aplican las reglas.**

`sf next` (`maquina.go:332`) y `sf done` (`done.go:144`) calculan el estado **efectivo** —el que sale
de aplicar el camino corto del bug— cada uno por su cuenta, con el mismo `if` copiado. `sf lote
start` (`lote.go:48`), `sf context` (`sobre.go:218`) y `sf approve` (`paradas.go:83`) leen `f.Estado`
**crudo**. De esa divergencia salen A1, A3 y A6.

Y hay una segunda causa, independiente: **`f.Lotes` vacío se lee como "todos los lotes cerrados"**
(`estado.go:368` devuelve `false`, y `done.go:204` interpreta ese `false` como final). De ahí salen
A0, la mitad de A1, y A5.

---

## Resumen

| # | Qué | Severidad | Bloquea |
|---|---|---|---|
| **A0** ✅ | `sf done` saltea `implementar` entero: sin branch, sin tests, sin código, sin commit | 🔴 **crítico** | la razón de ser del producto |
| **A1** ✅ | El camino corto del bug no funciona: llega a `cierre` sin escribir nada | 🔴 **crítico** | entrada C |
| **A2** ✅ | Archivar deja el repo inconsistente **e irrecuperable** | 🔴 **crítico** | el cierre de toda feature |
| **A3** ✅ | `sf approve` nunca corre compuertas | 🟠 alto | el ⑧ y el ⑨ |
| **A4** ✅ | El estado ⑧ es inalcanzable; `sfp-constitucion` es código muerto | 🟠 alto | el ⑧ |
| **A5** ✅ | `revision → implementar` es un loop muerto | 🟠 alto | el ㉑ con hallazgos |
| **A6** ✅ | `sf context` no aplica el camino corto del bug | 🟡 medio | entrada C |
| **A7** ✅ | `sf new` no lleva a `sfp-backlog` | 🟡 medio | entradas B y C |
| **A8** ✅ | `sf model` se ignora fuera de `implementar` | 🟡 medio | la salida de ME TRABÉ |
| **A9** ✅ | `compuerta.Roadmap` es código muerto | 🔵 bajo | el ⑩ |
| **M1–M3** ✅ | Números y punteros desactualizados en docs | 🔵 bajo | — |

**Orden ejecutado:** ~~A0 → A1 → A5~~ · ~~A2~~ · ~~A6~~ · ~~A3 → A4~~ · ~~A7~~ · ~~A8~~ ·
~~A9 + M~~ — **todo hecho.**

---

## A0 — `sf done` saltea `implementar` entero 🔴

**Lo que promete el README:** *"«todo verde» sin tests → `sf` tiene la lista exacta de tests que
deben existir, los corre él, y **se niega a dejarte implementar hasta haberlos visto fallar**."*

### Síntoma (reproducido)

Con el plan recién aprobado, **un solo `sf done`** —sin `sf lote start`, sin un archivo de test, sin
una línea de código, sin branch— mueve la feature a `revision`:

```
$ sf approve
✓ plan de f-1 aprobado — a implementar
$ sf done
✓ listo
→ todos los lotes cerrados — sigue la revisión (㉑)
$ git log --oneline
21ea927 init          ← no hay ningún commit de feature
$ ls
AGENTS.md  CLAUDE.md  go.mod      ← no hay código
```

Todo el mecanismo del producto —el rojo, el hash de los tests, un lote un commit, la branch— se
evita **corriendo `sf done` una vez**.

### Causa raíz

Cuando el plan se aprueba, `f.Lotes` está **vacío**: los lotes se siembran recién en
`EmpezarLote` (`lote.go:74-84`). Y `f.Lotes` vacío se lee en dos lugares como *"ya está todo"*:

- `compuerta.Implementar` (`compuerta.go:402`) — `LoteActual()` da `false`, y la función devuelve
  `Resultado{}` vacío, o sea **pasa**. El comentario dice *"todos commiteados: el estado puede
  cerrarse"*, y confunde "ninguno" con "todos".
- `cerrarLote` (`done.go:204`) — el mismo `!hayLote` dispara la rama *"todos los lotes cerrados —
  sigue la revisión"*.

`estado.Feature.LoteActual()` (`estado.go:368`) no puede distinguir los dos casos: devuelve
`(nil, false)` tanto para *"no hay lotes"* como para *"todos tienen commit"*.

### El arreglo

Distinguir **"nunca hubo lotes"** de **"todos cerrados"**, y que sólo el segundo cierre el estado.

1. En `estado`, agregar el predicado que hoy falta:

   ```go
   // SinSembrar dice que esta feature nunca pasó por `sf lote start`.
   //
   // No es lo mismo que "no queda ningún lote abierto": los lotes se siembran
   // recién cuando alguien va a trabajarlos, así que una lista vacía significa
   // que NADIE EMPEZÓ — y eso no puede leerse como "terminado".
   func (f *Feature) SinSembrar() bool { return len(f.Lotes) == 0 }
   ```

2. En `compuerta.Implementar`, frenar antes de mirar el lote actual:

   ```go
   if f.SinSembrar() {
       r.falla("todavía no empezaste ningún lote. Corré `sf lote start`.")
       return r
   }
   ```

3. En `cerrarLote` (`done.go:204`), la rama `!hayLote` queda como está — ya no se alcanza con la
   lista vacía, porque la compuerta frenó antes.

**Por qué así y no de otra forma:** es un hecho contable (`len(f.Lotes) == 0`), no un juicio. Cumple
R3 y no agrega estado nuevo: se deriva de lo que ya está en `estado.json`.

### El test que lo fija

`sf/internal/compuerta/compuerta_test.go`:

- `TestImplementarFrenaSinLotesSembrados` — una `estado.Feature` con `Lotes: nil` **no pasa**.
- `TestImplementarPasaConTodosLosLotesCommiteados` — la misma con dos lotes con `Commit != nil`
  **sí pasa** (protege que el arreglo no rompa el cierre legítimo).

`sf/internal/maquina/done_test.go`:

- `TestDoneNoMueveAImplementarSinLoteStart` — `Terminar` sobre una feature en `implementar` con
  `Lotes: nil` deja el estado en `implementar` y devuelve la falla.

---

## A1 — El camino corto del bug no funciona 🔴

**Lo que promete `docs/primeros-pasos.md`:** *"`tipo: bug` es lo que rutea. Un bug saltea
planificación y revisión y va derecho `implementar → cierre`."*

### Síntoma (reproducido)

`sf next` aplica el salteo y pide `sf lote start`. `sf lote start` lo rechaza:

```
$ sf next
estado:   implementar
feature:  f-2
skill:    sf-build
  sf lote start
$ sf lote start
✗ No pude.
  · f-2 está en "planificacion", y `sf lote start` es de implementar
```

El único comando que avanza es `sf done`, y con `f.Lotes` vacío salta directo a **cierre**:

```
$ sf done
✓ listo
→ lotes cerrados — es un bug, se saltea la revisión: sigue el cierre (㉓)
```

Un bug llega a cerrado **sin una línea de código, sin test, sin rojo y sin commit**.

### Causa raíz

Son dos, y las dos hay que arreglarlas:

1. **El salteo está copiado en dos lugares y falta en un tercero.**
   `maquina.go:332` lo calcula en una variable local; `done.go:144` lo escribe sobre `f.Estado`;
   `lote.go:48` compara contra `f.Estado` **crudo**, que sigue siendo `"planificacion"` porque nadie
   lo movió todavía.

2. **`f.Lotes` vacío se lee como "terminado"** — es exactamente A0, y por eso el `sf done` que
   destraba el punto 1 salta el estado entero.

### El arreglo

**Un solo lugar que decida el estado efectivo, y todos los comandos preguntándole ahí.**

1. Mover el predicado a donde vive el dato. En `internal/historia`:

   ```go
   // SonTodasBugs dice si todas las historias de una feature son bugs.
   //
   // Se exige que sean TODAS y no "alguna": saltear la planificación de una
   // feature que mezcla un bug con dos historias nuevas dejaría esas dos sin
   // diseño. Si están mezcladas, el agrupamiento del ⑩ ya estaba mal.
   func SonTodasBugs(raiz string, ids []string) bool
   ```

   Es literalmente `maquina.esDeBugs` (que vivía en `entradas.go`) movido, cambiando la firma de
   `roadmap.Feature` a `[]string` para que `historia` no importe `roadmap`.

2. En `internal/estado`, el que aplica la regla:

   ```go
   // Efectivo aplica el camino corto: un bug no pasa por planificacion ni por
   // revision. No hay carril paralelo — hay un campo que saltea dos estados.
   //
   // Lo tiene que consultar TODO el que decida por estado. Que `sf next` lo
   // aplicara y `sf lote start` no es exactamente el bug A1.
   func Efectivo(actual string, esBug bool) string {
       if esBug && actual == Planificacion {
           return Implementar
       }
       return actual
   }
   ```

3. Reemplazar los cuatro call sites por la misma línea:

   | Archivo | Hoy | Después |
   |---|---|---|
   | `maquina.go:332` | `if actual == Planificacion && esDeBugs(...)` | `est := estado.Efectivo(f.Estado, esBug)` |
   | `done.go:144` | `if f.Estado == Planificacion && esDeBugs(...)` | ídem, y persiste en `f.Estado` |
   | `lote.go:48` | `if f.Estado != estado.Implementar` | `if estado.Efectivo(f.Estado, esBug) != estado.Implementar` |
   | `sobre.go:218` | `switch f.Estado` | `switch estado.Efectivo(f.Estado, esBug)` ← esto es **A6** |

4. **`EmpezarLote` persiste el salteo.** Cuando entra por el camino corto, además de sembrar el lote
   único escribe `f.Estado = estado.Implementar`. Es el primer comando que muta el estado en ese
   camino, y dejar `estado.json` diciendo `"planificacion"` mientras el rojo ya está confirmado es
   la divergencia que causó todo esto.

   > **Trade-off, explícito:** `sf next` es consulta pura y no puede escribir; `sf lote start` sí
   > escribe (ya guarda `Rojo` y `HashTests`). Persistir acá no viola ninguna regla y hace que
   > `estado.json` deje de mentir.

5. Con A0 arreglado, `sf done` sobre un bug sin lotes ya no salta a `cierre`: frena pidiendo
   `sf lote start`. La rama relajada de `EmpezarLote` para el camino corto (`lote.go:121-136`,
   *"sin plan: alcanza con que algo falle"*) **ya está escrita y es correcta** — hoy es inalcanzable
   porque el guard de arriba la mata.

### El test que lo fija

`sf/internal/estado/estado_test.go`:
- `TestEfectivoSalteaPlanificacionSoloParaBugs` — tabla de los cuatro casos.

`sf/internal/maquina/lote_test.go`:
- `TestEmpezarLoteAceptaBugEnPlanificacion` — feature de bug con `f.Estado == "planificacion"`:
  siembra el lote único, exige el rojo, y deja `f.Estado == "implementar"`.
- `TestEmpezarLoteRechazaUsEnPlanificacion` — el guard sigue vivo para las historias normales.

`sf/internal/maquina/done_test.go`:
- `TestBugNoLlegaACierreSinLotes` — el `sf done` de un bug sin `lote start` **no** mueve a `cierre`.

---

## A2 — Archivar deja el repo inconsistente e irrecuperable 🔴

### Síntoma (reproducido)

```
$ sf approve
✗ No pude.
  · no pude pararme en main: git checkout main: error: Your local changes to the
    following files would be overwritten by checkout: .docs/estado.json
```

Y después de eso el proyecto **no sale más**:

```
$ sf done
✗ No avanzo.
  · falta .docs/features/f-1-suma/doc.md        ← la carpeta ya se movió
  · falta .docs/features/f-1-suma/journal.md
$ sf approve
✗ No pude.
  · no pude archivar la carpeta: rename ...f-1-suma → ...: no such file or directory
```

Estado final: carpeta **archivada**, feature **sin cerrar**, branch **sin mergear**, `sf next`
mandando a `cierre` sobre archivos que ya no existen. Sólo se sale editando el filesystem a mano.

### Causa raíz

`archivar` (`paradas.go:162`) hace las cuatro cosas en un orden que no aguanta un fallo:

```
paradas.go:175   os.Rename(origen, destino)    ← efecto irreversible, primero
paradas.go:189   git.Checkout(raiz, base)      ← lo que puede fallar, después
```

Y falla **siempre**, no de vez en cuando: `.docs/estado.json` está sucio en ese momento **por
construcción**. `sf` lo escribe en cada transición y sólo lo commitea dentro de `cerrarLote`
(`done.go:230`, vía `git.Commit` que hace `add -A`). Entre el último lote cerrado y el `approve` del
㉓ hay al menos dos escrituras de `estado.json` sin commit (el paso a `revision` y el paso a
`cierre`), así que `git checkout` aborta.

El remate está en `main.go:305`: `if ef.Pasa()` — como `archivar` devolvió falla, **el estado no se
guarda**, pero el `os.Rename` ya ocurrió. El disco avanzó y el `estado.json` no.

> El comentario de `paradas.go:180` dice *"si falla NO se deshace el movimiento: la carpeta archivada
> es correcta igual y el merge se puede reintentar a mano"*. La premisa es falsa: no se puede
> reintentar, porque el segundo intento falla en el `rename`.

### El arreglo

Son tres cambios y ninguno es opcional:

1. **Dejar el árbol limpio antes de tocar git.** `archivar` empieza commiteando el estado, que es
   trabajo real que hoy queda huérfano:

   ```go
   // El estado.json viene sucio por construcción: sf lo escribe en cada
   // transición y sólo lo commitea al cerrar un lote. Sin esto, el checkout
   // aborta SIEMPRE — no es un caso raro, es el caso.
   if git.EsRepo(raiz) && git.Sucio(raiz) {
       if _, err := git.Commit(raiz, "chore: cierre de "+fr.ID); err != nil { … }
   }
   ```

   Hace falta un `git.Sucio(dir) bool` nuevo (`git status --porcelain` vacío), al lado de
   `git.EsRepo` en `git.go:71`.

2. **Invertir el orden: git primero, `os.Rename` último.** Lo reversible antes que lo irreversible.
   Si el merge falla, no se movió nada y el `sf approve` se reintenta tal cual.

3. **Hacer el `rename` idempotente.** Si el destino ya existe y el origen no, el movimiento ya pasó:
   seguir en vez de fallar. Es lo que destraba un repo que ya quedó a medias con la versión de hoy.

Y como red: en `main.go:305`, guardar el estado también cuando `archivar` falló **después** de mover
la carpeta. La forma limpia es que `Efecto` tenga el mismo campo `Cambio` que ya tiene `Cierre`
(`done.go:35`), y que `parada()` guarde con `if ef.Pasa() || ef.Cambio`.

### El test que lo fija

`sf/internal/maquina/paradas_test.go`:
- `TestArchivarConEstadoSucio` — repo con `.docs/estado.json` modificado: `sf approve` **cierra la
  feature**, mergea y borra la branch. Es el caso normal, y hoy es el que falla.
- `TestArchivarNoMueveSiElMergeFalla` — con un conflicto forzado, la carpeta **sigue** en
  `.docs/features/`.
- `TestArchivarEsIdempotente` — con la carpeta ya en `.docs/archivado/`, `approve` termina bien.

---

## A3 — `sf approve` nunca corre compuertas 🟠

**Lo que promete `INSTALL.md`:** *"**No es opcional:** sin `test_cmd` caen tres compuertas —el rojo,
el verde y el conteo de tests— y **la constitución no se puede sellar**."*

### Síntoma (reproducido)

```
$ cat .docs/constitucion.md | head -4
---
lenguaje: ""
manifiesto: ""
test_cmd: ""        ← vacío
$ sf approve
✓ constitución sellada
```

Lo mismo en el ⑨: `sf done` rechaza una historia sin criterios, `sf next` igual ofrece `sf approve`,
y `approve` la sella:

```
$ sf done
✗ No avanzo.
  · us-1 no tiene criterios de aceptación (CA-1, CA-2, …)
$ sf approve
✓ backlog visto — sigue el ⑩
```

El backlog sellado sin ids desarma el mecanismo del ㉑ **entero**: sin criterios, la cobertura del ⑰
cuenta cero contra cero y pasa, y el conteo de veredictos del ㉑ también.

### Causa raíz

`Aprobar` (`paradas.go:83`) sella con una asignación directa y **no llama a ninguna compuerta**:

```go
paradas.go:110   case !e.Producto.ConstitucionSellada:
paradas.go:111       e.Producto.ConstitucionSellada = true    ← sin compuerta.Constitucion
paradas.go:116   case !e.Producto.BacklogVisto:
paradas.go:117       e.Producto.BacklogVisto = true           ← sin compuerta.Backlog
paradas.go:135   case estado.Planificacion:
paradas.go:137       f.Estado = estado.Implementar            ← sin compuerta.Planificacion
```

El único que valida algo es el `brief`, y de casualidad: lee el frontmatter para copiar el veredicto
y de paso comprueba que no esté vacío.

En `planificacion` el agujero está tapado por accidente —`sf next` corre la compuerta antes de
ofrecer el ⑰ (`maquina.go:471`)— pero un `sf approve` tipeado directo la saltea igual.

### El arreglo

`Aprobar` corre la compuerta del estado que va a sellar, y **no sella si no pasa**. Es la regla dura
del producto aplicada a `approve`: *"el estado avanza con hechos comprobados"* — y `approve` mueve
estado.

```go
case !e.Producto.ConstitucionSellada:
    if res := compuerta.Constitucion(raiz); !res.Pasa() {
        ef.Fallas = append(ef.Fallas, res.Fallas...)
        return ef
    }
    e.Producto.ConstitucionSellada = true
```

Igual para `backlog` (`compuerta.Backlog`) y para `planificacion` (`compuerta.Planificacion`).

**El `brief` queda como está**, y a propósito: su compuerta es *"trae uno de los tres veredictos"*, y
`approve` ya lo comprueba mejor —lee el valor concreto que va a sellar.

**El `cierre` también queda como está**: su compuerta (`compuerta.Cierre`) ya la corrió `sf next`
para decidir mostrar la ⏸, y correrla de nuevo acá no cambia nada. Si con A2 se toca esa función,
agregarla es barato y no molesta.

> **Nota de diseño para escribir en el comentario:** esto no convierte a `approve` en `done`. `done`
> **mueve** cuando la compuerta pasa; `approve` **sella lo que Javier decidió**, y lo único que
> cambia es que ahora no puede sellar algo que la máquina sabe que está roto. La decisión sigue
> siendo suya; deja de poder ser una decisión sobre un artefacto inválido.

### El test que lo fija

`sf/internal/maquina/paradas_test.go`:
- `TestAprobarNoSellaConstitucionSinTestCmd`
- `TestAprobarNoSellaBacklogSinCriterios`
- `TestAprobarNoSellaPlanIncompleto` — con `decision.md` de 2 opciones.
- `TestAprobarSellaCuandoLaCompuertaPasa` — los tres casos felices, para que el arreglo no bloquee
  el flujo normal.

---

## A4 — El estado ⑧ es inalcanzable; `sfp-constitucion` es código muerto 🟠

### Síntoma (reproducido)

En un proyecto recién iniciado, la máquina **nunca** pide trabajar la constitución. Del ⑦ salta
directo a la parada:

```
$ sf done                    # cierra el prd
✓ listo
→ prd_hash 6221423cb244 — sigue el ⑧
$ sf next
🛑 PARÁ. La constitución está escrita y la sellás vos (el ⑧).
```

Y lo que hay escrito es la plantilla que puso `sf init`:

```markdown
# Constitución

<Esto lo escribe el ⑧. Corré: sf next>
```

El skill `sfp-constitucion` está en el mapa (`maquina.go:163`), está en el CI, existe en `skills/`
— y **no lo invoca nadie nunca**.

### Causa raíz

`siguienteDeProducto` distingue *"falta hacerlo"* de *"está hecho, falta que lo mires"* por la
**existencia del archivo** (`maquina.go:249`):

```go
if !existe(raiz, docs.Constitucion) {
    return trabajar("constitucion", …)     ← inalcanzable
}
return Instruccion{Tipo: Para, …}
```

El patrón funciona para el `brief` y el `prd` porque `sf init` no los crea. Pero `sf init` **sí**
crea `.docs/constitucion.md` (`arranque.go:216`), y tiene que crearlo: ahí es donde escribe el
`test_cmd` que detectó del stack. El checkpoint por existencia y el andamio se pisan.

### El arreglo

Cambiar el checkpoint de *"¿existe?"* a *"¿está escrita?"*. El marcador ya está en el archivo — la
línea `<Esto lo escribe el ⑧. Corré: sf next>` — así que es una comparación de strings, no un juicio.

1. En `internal/arranque`, exportar el marcador para que haya **una** definición:

   ```go
   // MarcaSinEscribir es lo que `sf init` deja en el cuerpo de la constitución.
   //
   // Es un checkpoint, igual que la existencia del brief: distingue "el andamio
   // la creó" de "el ⑧ la escribió". Sin esto el ⑧ es inalcanzable, porque el
   // andamio SIEMPRE crea el archivo (necesita escribir el test_cmd del stack).
   const MarcaSinEscribir = "<Esto lo escribe el ⑧. Corré: sf next>"
   ```

2. En `internal/constitucion`, el predicado:

   ```go
   // SinEscribir dice si la constitución todavía es la plantilla del andamio.
   func SinEscribir(raiz string) bool
   ```

3. En `maquina.go:249`:

   ```go
   if !existe(raiz, docs.Constitucion) || constitucion.SinEscribir(raiz) {
       return trabajar("constitucion", "", "el ⑧ lo sellás vos cuando esté", g), true
   }
   ```

**Efecto de arrastre — ya verificado:** con esto, el sobre del ⑧ pasa a pedirse de verdad, y **el
sobre ya existe**. `sobre.deProducto` (`sobre.go:173`) tiene la rama de `constitucion` y sirve el PRD
(*"Qué hay que construir"*) más la cabecera técnica que llenó `sf init`. O sea que este arreglo no
arrastra trabajo en `sobre`: hoy ese código también es inalcanzable, y con A4 se enciende solo.

### El test que lo fija

`sf/internal/maquina/maquina_test.go`:
- `TestSiguienteMandaAEscribirLaConstitucionReciénIniciada` — proyecto con `sf init` + brief sellado
  + `prd_hash`: `sf next` devuelve `Trabajar` con skill `sfp-constitucion`.
- `TestSiguienteParaCuandoLaConstitucionYaSeEscribio` — con el marcador reemplazado, vuelve la 🛑.

---

## A5 — `revision → implementar` es un loop muerto 🟠

**Lo que promete el README:** *"Con un hallazgo abierto, vuelve a `implementar`."*

### Síntoma (reproducido)

Vuelve, sí. Pero no hay dónde trabajar:

```
$ sf done                       # revision con h-1 abierto
✗ No avanzo.
  · hay 1 hallazgos abiertos: h-1
→ hay hallazgos abiertos — vuelve a implementar

$ sf context
## Tus tareas
⚠ todos los lotes están commiteados: no queda nada por implementar

$ sf lote start
✗ No pude.
  · todos los lotes de f-1 ya están commiteados

$ sf done
✓ listo
→ todos los lotes cerrados — sigue la revisión (㉑)      ← rebota
```

`revision → implementar → revision → …` sin fin. El arreglo del hallazgo **no tiene dónde
commitearse**, y la única salida es `sf dismiss` (declarar el hallazgo falso) o editar
`revision.json` a mano. El skill `sf-build` tampoco cubre el caso: asume que hay un lote.

### Causa raíz

`cerrarRevision` (`done.go:243`) mueve el estado de vuelta y **no siembra trabajo**:

```go
if hayAbiertos(raiz, fr) {
    f.Estado = estado.Implementar
    return Cierre{Movio: true, …}         ← f.Lotes sigue con todos commiteados
}
```

Y `EmpezarLote` (`lote.go:87`) rechaza explícitamente sembrar de nuevo:

```go
if !hayLote { ef.falla("todos los lotes de %s ya están commiteados", fr.ID) }
```

### El arreglo

**La vuelta del ㉑ abre un lote nuevo.** Es la unidad de trabajo que la máquina ya tiene, y el
arreglo de un hallazgo es exactamente eso: un cambio con su test, que termina en un commit.

En `cerrarRevision`, al volver a `implementar`:

```go
// El hallazgo se arregla como cualquier otra cosa: en un lote, con su rojo y
// su commit. Sin esto el estado vuelve a `implementar` y no hay dónde trabajar
// — y el bucle rebota entre los dos estados para siempre.
f.Lotes = append(f.Lotes, estado.Lote{Lote: len(f.Lotes) + 1})
```

Ese lote **no está en `tareas.json`**, así que `EmpezarLote` cae en la rama relajada que ya existe
para el camino corto (`lote.go:121`): no puede exigir *cuáles* tests, pero **sigue exigiendo que la
suite falle**. Y eso es lo correcto acá: un hallazgo del ㉑ sin un test que lo reproduzca es un
hallazgo que nadie va a poder verificar.

Hace falta un ajuste en `lote.go`: hoy decide la rama relajada por `conPlan` (si `tareas.json` se
pudo leer). Con `tareas.json` presente pero el lote fuera de él, `p.TestsDelLote(n)` devuelve vacío y
cae en `ef.falla("el lote %d no tiene tests planificados. Eso tendría que haberlo frenado el ⑰")`
(`lote.go:124`), que acá es un mensaje equivocado. La condición pasa a ser:

```go
planificados = p.TestsDelLote(l.Lote)
if len(planificados) == 0 {
    // Un lote fuera del plan: la vuelta del ㉑, o el camino corto de un bug.
    // Se afloja a lo comprobable —que la suite falle— y se dice por qué.
    pasos = append(pasos, "lote de corrección: alcanza con que algo falle")
} else { … }
```

**Alternativa considerada y descartada:** hacer que el arreglo del hallazgo no pase por lote y se
commitee suelto. Se cae porque rompe *"un lote, un commit"*, que es lo que mata los dolores #2 y #3.

**Docs a tocar:** `docs/estados.md` (el ㉑) y `skills/sf-build/SKILL.md`, que tiene que decir qué
hacer cuando el sobre trae un hallazgo en vez de un lote del plan.

### El test que lo fija

`sf/internal/maquina/done_test.go`:
- `TestRevisionConHallazgosAbreLoteNuevo` — después de `Terminar`, `f.Lotes` tiene uno más y
  `LoteActual()` lo devuelve.

`sf/internal/maquina/lote_test.go`:
- `TestEmpezarLoteDeCorreccionSinPlan` — lote que no está en `tareas.json`: no se queja del ⑰, exige
  el rojo, y confirma.

---

## A6 — `sf context` no aplica el camino corto del bug 🟡

### Síntoma (reproducido)

`sf next` y `sf context` se contradicen sobre la misma feature:

```
$ sf next
estado:   implementar          ← el orquestador lanza sf-build
$ sf context
# Sobre — planificacion · f-2  ← el subagente recibe el sobre del ⑫
```

El subagente de `sf-build` recibe *"Qué hay que resolver"* (la historia) en vez de *"Tus tareas"*,
la spec y los criterios. Y de paso se pierde `loQueRompio` (`sobre.go:473`) —la spec archivada de la
feature que el bug rompió—, que **está implementada, es buena, y sólo cuelga de la rama
`implementar`**. Es justo lo que `docs/primeros-pasos.md` promete: *"el sobre del que lo arregla trae
la spec archivada de esa feature"*.

### Causa raíz

`sobre.Armar` (`sobre.go:216-218`) usa `f.Estado` crudo.

### El arreglo

Cae solo con el helper de A1:

```go
est := estado.Efectivo(f.Estado, historia.SonTodasBugs(raiz, fr.Historias))
s := &Sobre{Estado: est, Feature: fr.ID}
switch est {
```

`sobre` ya importa `historia` (`sobre.go:494`), así que no hay dependencia nueva.

### El test que lo fija

`sf/internal/sobre/sobre_test.go`:
- `TestSobreDeBugEnPlanificacionDaElDeImplementar` — y comprueba que la parte de `loQueRompio` esté.

---

## A7 — `sf new` no lleva a `sfp-backlog` 🟡

**Lo que promete `docs/primeros-pasos.md`:** *"`sf next` te manda a `sfp-backlog` a completarla."*
**Y lo que promete el propio código** (`entradas.go:83`): *"Volver a abrir la ⏸ del ⑨ es lo que hace
que `sf next` mande al pinponeo en vez de seguir con el ciclo de feature."*

### Síntoma (reproducido)

```
$ sf new "el CLI rompe con números negativos"
✓ us-2 creada. Ahora se pinponea: `sf next` te lleva.
$ sf next
⏸ Salieron las historias. Mirá si querés y seguí.
  sf approve                      ← no manda a sfp-backlog
$ sf context
# Sobre — backlog
## Qué hay que partir en historias
.docs/prd.md                      ← no menciona us-2
```

`sf new` deja un esqueleto con `titulo: ""` y un `CA-1` sin texto. Como el `CA-1` existe como **id**,
`compuerta.Backlog` lo cuenta y pasa. Y con A3 sin arreglar, `sf approve` lo sella.

### Causa raíz

`Nueva` hace su parte (`entradas.go:84`: `e.Producto.BacklogVisto = false`), pero
`siguienteDeProducto` sólo manda a trabajar si **no hay ninguna historia** (`maquina.go:263`):

```go
if !e.Producto.BacklogVisto {
    if !hayHistorias(raiz) {          ← con us-1 ya existiendo, es false
        return trabajar("backlog", …)
    }
    return Instruccion{Tipo: Barata, …}    ← la ⏸
}
```

El checkpoint por existencia vuelve a fallar por la misma razón que en A4: pregunta *"¿hay algo?"*
cuando la pregunta es *"¿está terminado?"*.

### El arreglo

El checkpoint pasa a ser *"¿hay alguna historia sin terminar?"*, que es un hecho contable sobre datos
que ya existen. En `internal/historia`:

```go
// SinPinponear son las historias que todavía son un esqueleto de `sf new`.
//
// El criterio es contable y no de gusto: sin título, o con criterios que sólo
// tienen id y no texto. Es lo que distingue "el ⑨ ya corrió" de "hay algo que
// alguien tiró al embudo y nadie completó".
func SinPinponear(raiz string) []string
```

Y en `maquina.go:263`:

```go
if faltan := historia.SinPinponear(raiz); len(faltan) > 0 {
    return trabajar("backlog", "", fmt.Sprintf(
        "completá: %s", strings.Join(faltan, " · ")), g), true
}
```

Es el mismo patrón que ya usa `huerfanas` para el ⑩ (`maquina.go:283`), que sí funciona.

**Y el sobre del ⑨ tiene que nombrarlas.** Hoy apunta al PRD y a la constitución y nada más; el que
completa un `us-#` que entró por `sf new` necesita **ese archivo**, y si es un bug, el `us-#`
original que `relacionado_a` señala.

**Segundo arreglo, en la misma zona:** `compuerta.Backlog` (`compuerta.go:162`) cuenta ids de
criterio y no mira si tienen texto. Un `- **CA-1** —` vacío pasa. Extender el chequeo a *"el criterio
tiene texto"* sigue siendo contar caracteres, no juzgar.

### El test que lo fija

`sf/internal/historia/historia_test.go`:
- `TestSinPinponearDetectaElEsqueletoDeNew`

`sf/internal/maquina/maquina_test.go`:
- `TestNextMandaAlBacklogDespuesDeNew`

`sf/internal/compuerta/compuerta_test.go`:
- `TestBacklogRechazaCriterioSinTexto`

---

## A8 — `sf model` se ignora fuera de `implementar` 🟡

**Lo que promete `docs/comandos.md`:** la cadena de precedencia de tres niveles, con `sf model` arriba
de todo *"porque una decisión suya en runtime siempre le gana a una recomendación escrita antes"*.

### Síntoma (reproducido)

```
$ sf model gpt5 --via consola --comando "codex exec"
✓ modelo: opus → gpt5 (contador reseteado)
$ sf next
estado:   cierre
modelo:   sonnet          ← ignoró el sf model
via:      subagente
```

La salida de ME TRABÉ no funciona en tres de los cuatro estados de feature.

### Causa raíz

`modeloDeFeature` —que implementa la precedencia completa (`maquina.go:626`)— tiene **un solo call
site**: `implementando` (`maquina.go:498`). Los otros tres estados de feature pasan por `trabajar()`
(`maquina.go:583`), que llama a `modelo(est)` (`maquina.go:602`) y sólo mira el mapa por estado.

`planificacion` va por `trabajar` en `maquina.go:443`, `revision` en `:357`, `cierre` en `:373`.

### El arreglo

`trabajar` es el helper genérico y también lo usan los cinco estados de producto, que no tienen
feature. Así que en vez de cambiarlo, agregar el hermano que sí la tiene:

```go
// trabajarEnFeature es trabajar() con la precedencia de modelo completa.
//
// Los cinco de producto no tienen feature y usan trabajar(); los cuatro del
// ciclo sí, y por eso el modelo sale de modeloDeFeature: `sf model` y el ⑯
// sólo existen dentro de una feature.
func trabajarEnFeature(raiz, est string, fr roadmap.Feature, f *estado.Feature,
    mensaje string, g *global.Config) Instruccion
```

Y usarlo en los tres call sites (`:357`, `:373`, `:443`). `implementando` (`:498`) ya hace esto a
mano y puede colapsarse adentro.

**Revisar de paso `tomarLaProxima` (`maquina.go:397`)**: calcula el `via` con `modelo(Planificacion)`
para anticipar la próxima feature. Ahí no hay `f` todavía, así que queda como está — pero vale un
comentario que diga por qué.

### El test que lo fija

`sf/internal/maquina/maquina_test.go`:
- `TestModeloDeJavierGanaEnRevisionYCierre` — tabla sobre los cuatro estados de feature con
  `f.Modelo` seteado.
- `TestModeloDelPlanGanaAlDefaultEnPlanificacion` — el nivel 2 (`tareas.json`).

---

## A9 — `compuerta.Roadmap` es código muerto 🔵

### Causa raíz

Único call site, en `terminarFeature` (`done.go:124`):

```go
if r == nil {
    res := compuerta.Roadmap(raiz)      ← sólo corre SIN roadmap.json
    …
}
```

`r` viene de `roadmap.Leer`, que devuelve `nil` sólo cuando el archivo no existe. Y lo primero que
hace `compuerta.Roadmap` (`compuerta.go:189`) es leerlo y fallar. O sea que sus dos chequeos reales
—historias huérfanas, y el aviso de feature con ≥6 historias— **no se ejecutan nunca desde el CLI**.

Las huérfanas están cubiertas por otro lado (`huerfanas` en `sf next`, `maquina.go:282`). El aviso
de las ≥6 no lo cubre nadie.

### El arreglo

Correr la compuerta cuando el ⑩ **acaba de escribir el roadmap**, que es el momento que le
corresponde. En `terminarFeature`, antes de bajar al ciclo:

```go
// El ⑩ es el último de producto, y su compuerta corre cuando el roadmap recién
// apareció: no hay feature en curso todavía. Con feature_actual ya puesto, el
// roadmap se dio por bueno y lo que sigue es el ciclo.
if e.FeatureActual == "" {
    res := compuerta.Roadmap(raiz)
    return Cierre{Resultado: res, Movio: res.Pasa(), Cambio: res.Pasa(), …}
}
```

**Alternativa más simple si esto complica el flujo:** borrar los dos chequeos muertos de
`compuerta.Roadmap` y sus tests, y dejar el aviso de ≥6 en `sf next` al lado de `huerfanas`. Código
muerto con tests que pasan es peor que código que no existe: da confianza falsa.

### El test que lo fija

`sf/internal/maquina/done_test.go`:
- `TestDoneDelRoadmapCorreLaCompuerta` — un `roadmap.json` que deja `us-2` afuera **no** avanza.

---

## Los menores

| | Dónde | Qué dice | Qué es |
|---|---|---|---|
| **M1** | `README.md` (§ Layout) y `README.es.md` | *"17 packages, 211 tests"* | 18 paquetes en `sf/internal/`, 213 tests en 19 paquetes de test |
| **M2** | `docs/comandos.md`, y la tabla de `README.md` | *"los 16 comandos"* | documenta 15 (`next` `context` `done` `lote start` `approve` `reject` `take` `model` `dismiss` `new` `status` `audit` `init` `install` `uninstall`) |
| **M3** | `skills-community/README.md:8` | *"aren't validated by `sf lint`"* | `sf lint` no existe. El chequeo real es el del CI (`.github/workflows/lint.yml`) |

**Sugerencia para M1:** los dos números son de los que envejecen callados. Un paso en el CI que los
compare contra la salida de `go test ./... 2>&1 | grep -c ^ok` cuesta tres líneas y no vuelve a
pasar.

---

## Orden de implementación

```
① A0 · A1 · A5      un solo PR: los tres son "f.Lotes vacío" + "el estado efectivo"
                    y tocan los mismos cuatro archivos. Separarlos deja el
                    árbol roto en el medio.
                      estado/estado.go    SinSembrar · Efectivo
                      historia/historia.go SonTodasBugs (movido de maquina)
                      compuerta.go        Implementar frena sin sembrar
                      maquina/lote.go     el guard · la rama sin plan
                      maquina/done.go     cerrarRevision siembra el lote

② A2                independiente y urgente: hoy ninguna feature se archiva.
                      git/git.go          Sucio()
                      maquina/paradas.go  commit · orden invertido · idempotente
                      maquina/Efecto      campo Cambio
                      cmd/sf/main.go      guardar con Pasa() || Cambio

③ A6                una línea, pero pide el helper de ①.

④ A3 · A4           los dos son "el checkpoint pregunta lo que no es".
                      maquina/paradas.go  Aprobar corre compuertas
                      constitucion/       SinEscribir
                      maquina/maquina.go  el checkpoint del ⑧
                    ⚠ verificar antes: ¿sobre.Armar tiene rama para "constitucion"?

⑤ A7                historia.SinPinponear + el sobre del ⑨ + criterio con texto.

⑥ A8                trabajarEnFeature en los tres call sites.

⑦ A9                decidir: revivir la compuerta, o borrarla y mover el aviso.

⑧ M1–M3            docs.
```

---

## Apéndice — el guion de humo

Lo que falta en el repo es un test que recorra la máquina **entera** por la superficie del binario.
Los 213 tests de hoy son de paquete, y ninguno de los diez hallazgos de arriba se ve desde adentro de
un paquete: **todos viven en la costura entre comandos.**

Propuesta: `sf/cmd/sf/e2e_test.go`, con `//go:build e2e`, que compile el binario a un tempdir, arme
un proyecto de juguete y recorra las dos vueltas. Cada paso es una aserción sobre el **exit code**,
que ya es parte de la interfaz (`main.go:44`).

```
# vuelta A — una historia normal
sf install · sf init                       → 0
sf next                                    → 0   brief · sfp-scout · via vos
escribir brief.md · sf approve             → 0
escribir prd.md · sf done                  → 0
sf next                                    → 0   constitucion · sfp-constitucion     ← A4
escribir constitucion · sf approve         → 0
sf approve (sin test_cmd)                  → 1                                       ← A3
escribir us-1 con CA · sf done · approve   → 0
sf approve (us sin CA)                     → 1                                       ← A3
escribir roadmap · sf done                 → 0
sf take f-1 · escribir plan · sf done      → 0
sf approve                                 → 0
sf done              (sin lote start)      → 2   ✗ no empezaste ningún lote          ← A0
escribir el test · sf lote start           → 0   branch · rojo · hash
sf done --msg "…"    (test aflojado)       → 2   ✗ los tests CAMBIARON
sf done --msg "…"    (honesto)             → 0   commit
sf done                                    → 0   → revision
revision con h-1 abierto · sf done         → 2   → implementar
sf lote start        (lote de corrección)  → 0                                       ← A5
sf done --msg "fix: …"                     → 0
revision limpia · sf done                  → 0   → cierre
doc + journal · sf done · sf approve       → 0   archivada, mergeada, branch borrada ← A2
sf next                                    → 3   🎉

# vuelta B — un bug por sf new
sf new "…"                                 → 0
sf next                                    → 0   backlog · sfp-backlog               ← A7
completar us-2 tipo:bug · sf done·approve  → 0
roadmap con f-2 · sf take f-2
sf next                                    → 0   implementar · sf-build
sf context                                 → 0   sobre de implementar + spec archivada ← A6/A1
sf lote start                              → 0                                       ← A1
sf done --msg "fix: …"                     → 0
sf done                                    → 0   → cierre (saltea revisión)
sf approve                                 → 0
```

**Este guion es el criterio de terminado del documento entero.** Cuando pasa completo, los diez
hallazgos están cerrados — y ninguno puede volver en silencio, que es exactamente como llegaron.


---

# Lo que se implementó — PR1

**A0 · A1 · A5 · A6.** A6 entró acá y no en un PR aparte porque sin él A1 quedaba a medias:
`sf lote start` aceptaba el bug y `sf context` seguía sirviéndole el sobre del ⑫, así que el
subagente que el orquestador lanzaba con `sf-build` recibía la historia en vez de la spec.

## Las dos piezas nuevas, y por qué son dos

```go
estado.Efectivo(actual string, esBug bool) string    // la regla del camino corto, en un solo lugar
estado.Feature.SinSembrar() bool                     // "nadie empezó" ≠ "todos terminaron"
historia.SonTodasBugs(raiz string, ids []string)     // movida de maquina.esDeBugs
```

`Efectivo` tiene ahora **cuatro** consumidores —`sf next`, `sf done`, `sf lote start` y
`sf context`— donde antes había dos copias del mismo `if` y dos comandos que no lo aplicaban.
`SonTodasBugs` bajó a `historia` porque la pregunta es sobre las historias y ya la necesitan
cuatro paquetes; recibe `[]string` para que `historia` no dependa de `roadmap`.

## Los archivos

| Archivo | Qué cambió |
|---|---|
| `estado/estado.go` | `Efectivo` y `SinSembrar` |
| `historia/historia.go` | `SonTodasBugs`, movida desde `maquina/entradas.go` |
| `compuerta/compuerta.go` | `Implementar` frena con `SinSembrar` **antes** de mirar el lote |
| `maquina/lote.go` | el guard usa `Efectivo` · persiste el salteo · separa "lote del plan sin tests" de "lote fuera del plan" |
| `maquina/done.go` | usa `Efectivo` · `abrirLoteDeCorreccion` en la vuelta del ㉑ |
| `maquina/maquina.go` | usa `Efectivo` |
| `maquina/entradas.go` | se fue `esDeBugs` |
| `sobre/sobre.go` | usa `Efectivo` · `hallazgosAbiertos` · distingue sin-sembrar de todo-commiteado |

## Dos cosas que el plan no había previsto

Las dos aparecieron al correr el guion de humo con el binario, y ninguna se ve desde adentro de un
paquete.

**① El lote de corrección necesitaba un sobre.** Abrir el lote destrabó el bucle, pero
`sf context` le servía las tareas del plan —una lista vacía, y un `lote 2 de 1` que no significa
nada—. Un lote sin sobre es un subagente fresco improvisando, que es exactamente lo que la máquina
existe para evitar. Ahora el sobre trae el hallazgo embebido: id, criterio y detalle. Los
descartados no van: los descartó Javier.

**② El sobre repetía la confusión de A0.** `loteYTareas` le decía *"todos los lotes están
commiteados: no queda nada por implementar"* a una feature que nunca había empezado — o sea, a un
subagente recién nacido para implementar lo mandaba a no hacer nada.

## Un chequeo que casi se pierde

Al aflojar la compuerta para el lote de corrección, el primer intento mandaba a la misma rama
**dos casos que dan "cero tests" y no son el mismo**:

```
el lote ESTÁ en el plan y no tiene tests   →  error del ⑰, se frena
el lote NO está en el plan                 →  lote de corrección, se afloja
```

Se distinguen con `slices.Contains(p.Lotes(), l.Lote)`, y hay un test para cada uno: sin el
segundo, el arreglo de A5 se habría llevado puesta una compuerta que funcionaba bien.

## Los tests

**242 pasan** (eran 213). Los nuevos se verificaron **revirtiendo el código y dejando los tests**:
los siete grupos fallan sin el arreglo, y dos de ellos son guardas de regresión que ya pasaban.

```
estado       TestSinSembrarNoEsLoMismoQueTodosCerrados
             TestEfectivoSalteaPlanificacionSoloParaBugs        (5 subtests)
historia     TestSonTodasBugs                                   (6 subtests)
compuerta    TestImplementarFrenaSinLotesSembrados
             TestImplementarPasaConTodosLosLotesCommiteados     ← guarda
maquina      TestLoteStartAceptaUnBugEnPlanificacion
             TestLoteStartRechazaUnaUsEnPlanificacion           ← guarda
             TestLoteStartDeCorreccionNoCulpaAlPlan
             TestLoteStartDelPlanSinTestsSigueCulpandoAl17      ← guarda
             TestDoneNoMueveARevisionSinHaberEmpezadoNingunLote
             TestDoneNoMandaUnBugACierreSinHaberEmpezadoNingunLote
             TestRevisionConHallazgosAbreUnLoteDeCorreccion
             TestDespuesDeLaVueltaDel21SePuedeEmpezarElLote
sobre        TestElSobreDelLoteDeCorreccionTraeElHallazgo
             TestElSobreDelLoteDelPlanSigueTrayendoLasTareas    ← guarda
             TestElSobreDeUnBugEnPlanificacionEsElDeImplementar
             TestElSobreDeUnaUsEnPlanificacionNoSeSaltea        ← guarda
             TestElSobreDistingueSinEmpezarDeTodoCommiteado
```

## El guion de humo, corrido a mano con el binario

Las dos vueltas del apéndice, hasta donde llega PR1:

```
sf done sin lote start          → 2  ✗ no empezaste ningún lote            A0
lote start · done --msg         → 0  branch · rojo · hash · commit
revisión con h-1 abierto        → 2  → vuelve a implementar en el lote 2   A5
context                         → 0  trae h-1, su criterio y el detalle    A5
lote start con la suite verde   → 1  ✗ no está reproducido lo que arreglás A5
lote start con el rojo          → 0  "lote de corrección"                  A5
done --msg · revisión limpia    → 0  → cierre

sf take f-2 (tipo: bug) · next  → 0  implementar · sf-build                A1
context                         → 0  sobre de implementar + spec archivada A1/A6
lote start                      → 0  "sin plan (camino corto)"             A1
done --msg · done               → 0  → cierre, saltea la revisión          A1
```

**Lo que quedó probado y antes no se podía hacer:** un bug entra por `sf new`, se reproduce con un
test que falla, se arregla, y se commitea — sin poder llegar a `cierre` por el camino de no hacer
nada.

## Docs actualizadas

- `docs/estados.md` — la vuelta del ㉑ y qué cambia en el lote de corrección.
- `skills/sf-build/SKILL.md` — qué hacer cuando el sobre dice *"Qué hay que arreglar"* en vez de
  *"Tus tareas"*.

## Lo que sigue

**A2 es ahora el bloqueante que queda**, y es el más urgente de los tres críticos que había:
`sf approve` en `cierre` falla siempre —`.docs/estado.json` está sucio por construcción— y deja el
repo a medias sin forma de salir. En el guion de humo de arriba hubo que commitear a mano para
poder archivar `f-1`.


---

# Lo que se implementó — PR2

**A2**, el último de los tres críticos.

## El bug, reproducido con el binario viejo

Antes de tocar nada, se construyó el binario del commit anterior y se lo corrió contra un repo
real llevado hasta el ㉓. Falla, y deja el repo sin salida:

```
$ sf approve
✗ no pude pararme en main: Your local changes to the following files would be
  overwritten by checkout: .docs/estado.json

$ sf approve                       # el segundo intento, sobre lo que quedó
✗ no pude archivar la carpeta: rename …/features/f-2-overflow → …: no such file or directory
```

Carpeta archivada · feature en `cierre` · branch sin mergear · y ningún comando que salga de ahí.

## Los tres cambios, y por qué son tres

**① Commitear antes de tocar git.** `.docs/estado.json` está sucio *por construcción* al llegar al
㉓ —`sf` lo escribe en cada transición y sólo lo commitea al cerrar un lote—, así que el
`git checkout` abortaba **siempre**, no de vez en cuando. Y el commit no es un truco para destrabar
el checkout: la revisión, la doc y el journal son trabajo real que si no quedaba huérfano.

**② Invertir el orden: git primero, `os.Rename` último.** El movimiento es lo único irreversible y
el git es lo único que puede fallar — estaban exactamente al revés. Ahora un merge conflictivo no
movió nada y `sf approve` se reintenta tal cual.

**③ Hacer el `rename` idempotente.** Es lo único que destraba a un repo que ya quedó a medias con
la versión vieja. Se comprobó de verdad: el repo que el binario viejo dejó trabado se recuperó
corriendo `sf approve` con el binario nuevo, sin tocar el filesystem a mano.

## Una cosa que el plan no había previsto

**El archivado entero quedaba sin commitear**, y eso no es neutral: el `git add -A` del primer lote
de la feature SIGUIENTE se lo lleva puesto, y el cierre de `f-1` termina adentro del commit de
`f-2`. Es el **dolor #3 —commits mal agrupados— colándose por el único lugar donde `sf` mueve
archivos sin estar cerrando un lote.**

No se podía commitear dentro de `archivar`, porque `main` escribe el `estado.json` *después* de que
la función vuelve. Así que `Efecto` ganó un campo `Commit string` —"cuando el estado esté guardado,
commiteá con este mensaje"— y `main` lo ejecuta al final. El archivado queda en un commit propio,
con el estado adentro.

> Esto excede lo que el plan pedía para A2. Se hizo igual porque es el mismo defecto —archivar deja
> el repo en un estado que después muerde— y porque el arreglo son diez líneas.

## Los archivos

| Archivo | Qué cambió |
|---|---|
| `git/git.go` | `Sucio()` — `git status --porcelain`, incluidos los sin trackear |
| `maquina/paradas.go` | `archivar` reordenado · `archivarCarpeta` idempotente · `Efecto.Cambio` y `Efecto.Commit` |
| `cmd/sf/main.go` | guarda con `Pasa() \|\| Cambio` · ejecuta `ef.Commit` después de guardar |

`Efecto.Cambio` existe porque `archivar` toca el disco: si algo falla *después* de eso, no guardar
el estado dejaría al `estado.json` describiendo un repo que ya no existe. Es el mismo campo que
`Cierre` ya tenía en `done.go`, por la misma razón.

## Los tests

**246 pasan** (eran 242).

```
git        TestSucio                                        (limpio · modificado · sin trackear)
maquina    TestArchivarConElEstadoSucioIgualCierraLaFeature  ← el caso normal, que fallaba siempre
           TestArchivarNoMueveLaCarpetaSiElMergeFalla        ← lo irreversible va último
           TestArchivarEsIdempotenteSiLaCarpetaYaEstaba      ← destraba lo ya roto
```

El del merge fallado fuerza un conflicto de verdad —`main` y la branch tocan el mismo archivo— en
vez de simular el error.

## El archivado, corrido de punta a punta

```
sucio antes del approve   4 archivos
sf approve                → 0   ✓ f-1 archivada y cerrada
branch                    → main
sucio después             0 archivos
sf next                   → 3   🎉
```

```
* chore: f-1 archivada
*   Merge branch 'feat/f-1-x'
|\
| * chore: cierre de f-1
| * feat: x
|/
* init
```

## Docs actualizadas

`docs/estados.md` (el ㉓: el orden y por qué), `docs/comandos.md` (`sf approve`) y
`docs/primeros-pasos.md` (las cuatro cosas mecánicas, ahora en el orden real).

## Lo que sigue

**No queda ningún crítico.** Lo próximo es **A3 + A4**, que comparten `Aprobar`: hoy `sf approve`
no corre ninguna compuerta —sella la constitución sin `test_cmd` y el backlog sin criterios—, y el
estado ⑧ es inalcanzable porque `sf init` siempre crea `constitucion.md`.


---

# Lo que se implementó — PR3

**A3 y A4**, los dos altos. Van juntos porque los dos son el mismo error de forma: **un checkpoint
que pregunta lo que no es.**

```
A4   "¿existe el archivo?"   cuando la pregunta era "¿está escrito?"
A3   "¿Javier dijo que sí?"  cuando la pregunta era "¿y además está bien?"
```

## A4 — el ⑧ vuelve a existir

`sf init` **siempre** crea `.docs/constitucion.md`, y tiene que crearla: ahí escribe el `test_cmd`
que detectó del stack. Pero el checkpoint miraba si el archivo existía, así que del ⑦ se saltaba
derecho a *"🛑 la constitución está escrita, sellala vos"* — y lo que se sellaba era la plantilla.
El skill `sfp-constitucion` estaba en el mapa, en el CI y en el disco, y **no lo invocaba nadie
nunca**.

El arreglo es el marcador que el andamio ya dejaba, ascendido a constante con dueño:

```go
constitucion.MarcaSinEscribir = "<Esto lo escribe el ⑧. Corré: sf next>"
constitucion.SinEscribir(raiz) bool
```

`arranque` la usa para escribir la plantilla y `maquina` para leerla, así que **el marcador tiene
una sola definición**: si alguien cambia el texto, no puede desincronizar las dos puntas. Y la
pregunta sigue siendo comparar dos strings — un hecho, no un juicio.

```
$ sf next
estado:   constitucion
skill:    sfp-constitucion      ← antes: 🛑 sellá la plantilla
```

## A3 — `approve` corre la compuerta antes de sellar

Sellaba con una asignación directa. O sea que la regla dura del producto —*"el estado avanza con
hechos comprobados"*— valía para `sf done` y **no para el otro comando que mueve el estado**:

```
$ sf approve      # constitución con test_cmd vacío
✓ constitución sellada

$ sf approve      # backlog con historias sin criterios
✓ backlog visto — sigue el ⑩
```

Lo primero contradecía a `INSTALL.md` textual. Lo segundo es lo grave: **un backlog sellado sin ids
desarma el mecanismo entero** — sin criterios, la cobertura del ⑰ cuenta cero contra cero y pasa, y
el conteo de veredictos del ㉑ también.

Ahora las cuatro paradas que sellan corren su compuerta:

| Parada | Compuerta |
|---|---|
| `constitucion` | `compuerta.Constitucion` — `test_cmd` no vacío |
| `backlog` (⏸) | `compuerta.Backlog` — todas las historias con ids |
| `planificacion` (⑰) | `compuerta.Planificacion` — las cinco |
| `cierre` (⏸) | `compuerta.Cierre` — la doc y el journal |

**El `brief` queda como estaba**, y no por olvido: su compuerta es *"trae uno de los tres
veredictos"*, y `approve` ya lee el valor concreto que va a sellar — el mismo chequeo, hecho mejor.

> **Esto no convierte a `approve` en `done`.** `done` **mueve** cuando la compuerta pasa; `approve`
> **sella lo que Javier decidió**, y lo único que cambia es que ya no puede sellar algo que la
> máquina sabe que está roto. La decisión sigue siendo suya; deja de poder ser una decisión sobre
> un artefacto inválido.

`Efecto` gana `Avisos`, para no tragarse los de la compuerta. Van primero y en los dos casos —salga
bien o mal—: esconderlos detrás de un ✓ es la forma más fácil de que nadie los lea.

## Dos cosas que aparecieron al implementar

**① El plan decía que la compuerta del ㉓ "queda como está", y estaba mal.** Se agregó igual —
archivar es irreversible, y hacerlo sin doc ni journal deja la carpeta en `.docs/archivado/` sin lo
único que alguien va a leer seis meses después. Pero al agregarla **rompió el test de idempotencia
de A2**: `compuerta.Cierre` buscaba en `.docs/features/`, y en el caso de recuperación la carpeta
ya está archivada.

El arreglo no fue sacar la compuerta sino **destaparla**: ahora busca en la carpeta viva y, si no
está, en la archivada. La pregunta es *"¿el ㉓ produjo sus dos archivos?"*, no *"¿dónde está la
carpeta hoy?"* — y atarla al lugar la hacía fallar justo cuando más se la necesita.

**② Un test existente estaba asegurando el bug.** `TestApproveDelPlanMandaAImplementar` aprobaba un
plan sin ninguno de los tres archivos y esperaba que pasara. No falló por casualidad: **falló
porque el arreglo es correcto.** Se le agregó el plan completo al andamio.

## Los archivos

| Archivo | Qué cambió |
|---|---|
| `constitucion/constitucion.go` | `MarcaSinEscribir` · `SinEscribir` |
| `arranque/arranque.go` | la plantilla usa la constante en vez del literal |
| `maquina/maquina.go` | el checkpoint del ⑧ pregunta si está **escrita** |
| `maquina/paradas.go` | `Aprobar` corre la compuerta · `Efecto.Avisos` · `Efecto.compuerta` |
| `compuerta/compuerta.go` | `Cierre` busca también en la carpeta archivada |

## Los tests

**252 pasan** (eran 246). Verificados revirtiendo `paradas.go` y `maquina.go` —dejando el resto—
para ver las fallas de comportamiento y no de compilación: los cuatro fallan.

```
maquina    TestElOchoEsAlcanzableConLaPlantillaDelAndamio
           TestElOchoParaCuandoLaConstitucionYaSeEscribio     ← guarda
           TestApproveNoSellaLaConstitucionSinTestCmd
           TestApproveNoSellaElBacklogSinCriterios
           TestApproveNoApruebaUnPlanIncompleto
           TestApproveMuestraLosAvisosDeLaCompuerta
```

Los tres de `approve` comprueban además **el caso feliz**: el arreglo no puede frenar el flujo
normal, y sin eso una compuerta demasiado estricta se descubre en producción.

## Corrido con el binario

```
sf next    (repo vacío, sin stack)  → 0   constitucion · sfp-constitucion      A4
sf approve (test_cmd vacío)         → 1   ✗ la constitución no tiene test_cmd  A3
sf approve (con test_cmd)           → 0   ✓ constitución sellada
sf approve (us-1 sin criterios)     → 1   ✗ us-1 no tiene criterios            A3
sf approve (con criterios)          → 0   ✓ backlog visto
sf approve (plan sin CA-2)          → 1   ✗ cubren 1 de 2. Sin cubrir: CA-2    A3
sf approve (plan completo)          → 0   ✓ plan aprobado
sf approve (cierre sin doc)         → 1   ✗ falta doc.md · falta journal.md    A3
                                          y la carpeta NO se movió
```

## Docs actualizadas

`INSTALL.md` (deja de prometer lo que el binario no hacía), `docs/estados.md` (el ⑧ y su marcador ·
la ⏸ del ⑨ es barata pero la compuerta no), `docs/comandos.md` (`sf approve` corre la compuerta) y
`docs/primeros-pasos.md` (el ⑧ aparece en el recorrido, que antes se lo salteaba igual que la
máquina).

## Lo que sigue

**A7** — `sf new` no lleva a `sfp-backlog`, aunque el propio código dice que sí. Y de paso la
compuerta del backlog cuenta ids de criterio sin mirar si tienen texto, así que el esqueleto vacío
que deja `sf new` la pasa. Después **A8** (`sf model` ignorado fuera de `implementar`) y **A9**
(`compuerta.Roadmap`, código muerto).


---

# Lo que se implementó — PR4

**A7**, y es el tercer caso del mismo error de forma que ya apareció en el ⑧ y en los lotes:

```
"¿hay historias?"   funciona UNA vez — la primera
"¿qué falta?"       funciona siempre
```

## El bug era doble

**El enrutado.** El propio código decía lo que no hacía (`entradas.go`): *"volver a abrir la ⏸ del
⑨ es lo que hace que `sf next` mande al pinponeo"*. `Nueva` reabría la ⏸, sí — pero el checkpoint
preguntaba si había historias, y con un producto en marcha eso es siempre sí. Salía la ⏸ y lo que
quedaba para aprobar era el esqueleto.

**La compuerta.** Y el esqueleto **pasaba**, porque contaba ids y el esqueleto trae `- **CA-1** —`
con el id puesto y el texto no. Un criterio que no dice nada es **peor que ninguno**: el ⑰ lo da
por cubierto y el ㉑ le pone veredicto, así que el mecanismo de los ids queda en pie sobre algo que
nadie puede juzgar.

## Lo que quedó

```go
historia.Historia.SinTexto      los criterios con id y sin texto
historia.SinPinponear(raiz)     qué historias siguen siendo un esqueleto
historia.marcasDelEsqueleto     los huecos que deja `sf new`, compartidos
sobre.aPinponear(raiz)          la parte del ⑨ que dice CUÁL
```

El mensaje del ⑨ ahora distingue las dos entradas: con todas incompletas sigue siendo *"el ⑨ parte
el PRD en historias"*; con algunas, es `completá: us-7`. Y el sobre las nombra en vez de servir el
PRD entero — más la historia que el bug rompió, si `relacionado_a` apunta a una.

## Tres correcciones sobre la marcha

**① Exigí el título en la compuerta, y era de más.** Rompió cinco tests, y tenían razón: la
compuerta pregunta si el **mecanismo** se sostiene, y lo que el ⑰ cuenta y el ㉑ juzga son los
criterios. Un título flojo no rompe nada. El título quedó sólo en el checkpoint, donde la pregunta
es otra.

**② El checkpoint atado a `titulo:` daba falsos positivos.** Lo delató el binario: en un proyecto
real, una historia completa escrita a mano —con el título sólo en el encabezado— se marcaba como
pendiente, y el ⑨ decía *"partí el PRD"* cuando lo que faltaba era una sola historia.

El arreglo es **el mismo marcador que el ⑧**: `sf new` deja `<título>`, `<quién>`, `<qué>`,
`<por qué>`, y ahora esos huecos son constantes que **escribe y busca el mismo paquete**, así que
no se pueden desincronizar. Un marcador no se equivoca: o está, o no.

> **Un checkpoint que frena trabajo terminado es peor que no tenerlo** — se aprende a ignorarlo.

**③ Con CERO historias, `SinPinponear` no devuelve nada** —una lista vacía no tiene nada
incompleto—, así que el backlog vacío se iba a la ⏸ en vez de mandar a escribirlas. Es literalmente
A0 otra vez: *"ninguno" no es "todos"*. Los dos casos se preguntan por separado.

## Los archivos

| Archivo | Qué cambió |
|---|---|
| `historia/historia.go` | `SinTexto` · `SinPinponear` · las marcas del esqueleto, compartidas con `Esqueleto` |
| `compuerta/compuerta.go` | `Backlog` exige que el criterio diga algo |
| `maquina/maquina.go` | el checkpoint del ⑨ pregunta qué falta · `mensajeDelNueve` |
| `sobre/sobre.go` | `aPinponear` — el sobre nombra la historia y la que el bug rompió |

## Los tests

**265 pasan** (eran 252). Verificados revirtiendo `maquina.go`, `compuerta.go` y `sobre.go`.

```
historia   TestSinTextoAtrapaElCriterioVacio
           TestSinTextoNoSeComeLosCriteriosEscritos          ← em dash · dos puntos · sin negritas
           TestSinPinponearDetectaElEsqueletoDeNew
           TestSinPinponearNoDevuelveNadaConElBacklogCompleto   ← guarda
           TestSinPinponearNoFrenaUnaHistoriaSinTituloEnElFrontmatter ← el falso positivo ②
compuerta  TestBacklogRechazaElCriterioSinTexto
           TestBacklogNoExigeTitulo                          ← la corrección ①
maquina    TestDespuesDeNewElNueveMandaACompletarLaHistoria
           TestConElBacklogCompletoVuelveLaParadaBarata      ← guarda
           TestLaPrimeraVueltaDelNueveSigueSiendoPartirElPRD ← la corrección ③
sobre      TestElSobreDelNueveNombraLoQueHayQueCompletar
           TestElSobreDelNueveTraeLaHistoriaQueElBugRompio
           TestElSobreDelNueveNoAgregaNadaEnLaPrimeraVuelta  ← guarda
```

Y dos tests existentes usaban un `us-1.md` **vacío** para llegar a la ⏸. Ahora un archivo vacío es
—correctamente— una historia sin pinponear, así que se les dio una historia completa: el helper
`conHistoria` existe para eso.

## Corrido con el binario

```
sf new "…"                      → 0   us-2 creada
sf next                         → 0   backlog · sfp-backlog · "completá: us-2"
sf context                      → 0   ## Lo que hay que completar → us-2.md
sf approve  (el esqueleto)      → 1   ✗ us-2 tiene el id puesto y el criterio vacío
sf done     (el esqueleto)      → 2   ✗ lo mismo
sf next     (ya completa)       → 2   ⏸ Salieron las historias
sf approve                      → 0   ✓ backlog visto
```

## Docs actualizadas

`docs/estados.md` (el ⑨ corre dos veces y el checkpoint lo sabe · la compuerta exige texto y no
título), `docs/comandos.md` (`sf new` lleva de verdad) y `docs/primeros-pasos.md` (el ⑨ nombra la
historia, y el sobre del bug trae la que rompió).

## Lo que sigue

**A8** — `modeloDeFeature` se usa sólo en `implementar`, así que `sf model` se ignora en
`planificacion`, `revision` y `cierre`: la salida de ME TRABÉ no funciona en tres de los cuatro
estados de feature. Después **A9** y los menores.


---

# Lo que se implementó — PR5

**A8.** Es el más chico de los nueve y el que tiene la consecuencia más incómoda: la salida de
emergencia no funcionaba donde más se la necesita.

## El bug

`modeloDeFeature` implementa la cadena de tres niveles y tenía **un solo consumidor**:
`implementando`. Los otros tres estados de feature pasaban por `trabajar`, que sólo mira el mapa
por estado.

```
$ sf model gpt5 --via consola --comando "codex exec"
✓ modelo: opus → gpt5 (contador reseteado)
$ sf next
estado:   cierre
modelo:   sonnet          ← lo ignoró
```

O sea que **`sf model` —la salida de ME TRABÉ, la que Javier usa mirando el bucle patinar— no hacía
nada en tres de los cuatro estados de feature.** Y ME TRABÉ puede aparecer en cualquiera de ellos:
la dispara el contador de `sf done` fallidos, y `done` corre en los cuatro.

## El arreglo

`trabajarEnFeature`, hermano de `trabajar` con la cadena completa, en los cuatro call sites:

| Antes | Ahora |
|---|---|
| `revision` → `trabajar` | `trabajarEnFeature` |
| `cierre` → `trabajar` | `trabajarEnFeature` |
| `planificacion` → `trabajar` | `trabajarEnFeature` |
| `implementar` → la cadena a mano | `trabajarEnFeature`, y se le borraron 12 líneas |

`trabajar` **se queda**, y para lo que corresponde: los cinco estados de producto, donde los dos
niveles de arriba de la cadena no existen porque los dos viven dentro de una feature. Después del
cambio, sus únicos seis call sites son ésos.

Y `tomarLaProxima` —el anticipo de la próxima feature— aplica la cadena si esa feature ya está en
el estado: puede estar `planificada` (la puerta *"otra feature"* del ⑰) y traer un `sf model` de
una vuelta anterior. Anticipar el default cuando el real es otro es la misma desinformación que el
`via` ya evitaba un renglón más abajo.

## Un defecto que apareció al mirar la salida

```
⚠ ME TRABÉ. f-1 falló 3 veces seguidas con .
```

El mensaje leía `f.Modelo` **crudo**, y ese campo está vacío hasta el primer `sf model` — o sea que
la **primera** vez que aparece la parada es exactamente la vez que sale sin el nombre. Y el nombre
es la mitad del mensaje: *"¿subo el modelo?"* no se puede contestar sin saber cuál está fallando.

Ahora sale de la cadena, igual que todo lo demás:

```
⚠ ME TRABÉ. f-1 falló 3 veces seguidas con opus.
```

Hubo que subir el `r.Buscar` de la feature por encima del bloque de ME TRABÉ. Es una lectura pura,
así que mover el orden no cambia nada más.

## Los archivos

Uno: `maquina/maquina.go`. Es el arreglo más contenido de los cinco PRs.

## Los tests

**273 pasan** (eran 265).

```
TestElModeloDeJavierGanaEnLosCuatroEstadosDeFeature   (4 subtests, uno por estado)
TestElModeloDelPlanGanaAlDefaultDelEstado             el nivel 2, y que el 1 le gana
TestSinDecidirNadaMandaElDefaultDelEstado             ← guarda
TestMeTrabeDiceConQueModeloSeTrabo
```

El primero es una tabla sobre los cuatro estados a propósito: revirtiendo `maquina.go` fallan
`planificacion`, `revision` y `cierre`, y **`implementar` pasa en las dos versiones**. Eso es la
demostración exacta de cuál era el bug.

## Corrido con el binario

```
sf next     (cierre)              → modelo: sonnet
sf model opus                     → ✓
sf next                           → modelo: opus                      A8

sf next     (revision, 3 fallos)  → ⚠ ME TRABÉ … con opus             el defecto de arriba
sf model gpt5 --via consola …     → ✓ declarado en ~/.specforge/
sf next                           → modelo: gpt5 · via: consola
                                    comando: codex exec
```

La última línea es el punto: la salida de ME TRABÉ funciona en `revision` **y** resuelve el `via`
de un modelo que no es de Anthropic, que era el caso H1b entero.

## Docs actualizadas

`docs/comandos.md` (la cadena de precedencia escrita, y que `sf model` vale en los cuatro estados)
y `docs/problemas.md` (por qué el nombre del modelo está en el mensaje de ME TRABÉ).

## Lo que sigue

**A9**, el último, y es una decisión más que un arreglo: `compuerta.Roadmap` sólo se llama con
`r == nil` —o sea sin `roadmap.json`—, y ahí falla en la primera línea. Sus dos chequeos reales no
corren nunca. Hay que revivirla o borrarla; **código muerto con tests que pasan es peor que código
que no existe**, porque da confianza falsa. Y después los tres menores de docs.


---

# Lo que se implementó — PR6

**A9** y los tres menores. El último, y el único que era una decisión antes que un arreglo.

## A9 — la compuerta corría en el único caso en que no sirve

`compuerta.Roadmap` estaba colgada de `r == nil`, y `r` es nil **sólo cuando no hay
`roadmap.json`** — o sea que la función fallaba en su primera línea, al intentar leerlo. Sus dos
chequeos reales no se ejecutaban nunca.

Y había un segundo síntoma que el plan no había visto: con el roadmap ya escrito, el `sf done` del
⑩ caía en el camino de feature y contestaba

```
✗ no hay feature en curso: corré `sf take <feature>` primero
```

o sea que **le decía al subagente que hizo bien su trabajo que se había equivocado.** Los dos
síntomas eran la misma línea.

### La decisión: revivirla, no borrarla

Las dos opciones estaban abiertas. Se revivió por el chequeo que **no cubre nadie más**:

| Chequeo | ¿Lo cubre otro? |
|---|---|
| historia huérfana | sí — `huerfanas` en `sf next` |
| **feature con 6+ historias** | **no** |

Borrarla habría perdido ese aviso en silencio, que es justo la forma de pérdida que esta auditoría
encontró diez veces.

### Dónde corre ahora

Cuando **no hay feature en curso**, que es exactamente el `sf done` del ⑩. Y no mueve el estado, a
propósito: el ⑩ no guarda sello —*"¿existe el `roadmap.json`?"* es deducible (R6)— y la transición
al ciclo la hace `sf take`, que es una decisión de Javier (H4).

```
$ sf done
⚠ f-1 junta 6 historias. ¿La partís?
✓ listo
→ el roadmap está listo. Elegí con `sf take <feature>`.
```

### Un panic que estaba esperando

Al sacar el `if r == nil` del principio quedó expuesto `r.Buscar(...)`, que hace
`range r.Features` sobre un puntero nil. El estado que lo alcanza es raro pero existe: un
`estado.json` con `feature_actual` puesto y el `roadmap.json` borrado a mano.

`Buscar` es ahora nil-safe —un roadmap que no existe no tiene features, así que *"no está"* es la
respuesta honesta— y eso **protege también a `Aprobar`**, que tenía el mismo riesgo desde antes y
nadie lo había mirado.

## M1–M3, y un chequeo para que no vuelvan

| | Decía | Es |
|---|---|---|
| **M1** | *"17 packages, 211 tests"* | **18 paquetes, 277 tests** |
| **M2** | *"los 16 comandos"* | 15 |
| **M3** | *"aren't validated by `sf lint`"* | `sf lint` no existe — es el CI |

**Y el CI ahora los compara.** Los dos números de M1 son de los que envejecen callados: nadie los
mira al agregar un paquete o un test, y quedaron mintiendo por 7 paquetes y 66 tests hasta que
alguien los contó a mano. Tres líneas en `lint.yml` y no vuelve a pasar.

## Los tests

**277 pasan** (eran 273).

```
TestElDoneDelDiezAtrapaLaHistoriaHuerfana
TestElDoneDelDiezCierraSinMoverElEstado              ← y que NO mueve el estado
TestElDoneDelDiezAvisaDeLaFeatureGrandeSinFrenar     ← el chequeo que no cubre nadie más
TestSinRoadmapNoRevienta                             ← el panic expuesto
```

Los tres primeros fallan revirtiendo `done.go`, y con el mensaje viejo — *"no hay feature en curso"*
donde tendría que decir cuál historia quedó afuera.

## Residuo conocido

**El ⑩ puede correr a mitad del ciclo** —`sf new` deja una historia huérfana y `sf next` prioriza
el ⑩ aunque haya una feature en curso—, y ahí el `sf done` del subagente del roadmap cae en el
camino de feature: intenta cerrar el lote de la feature en curso y suma un `intentos_fallidos`
espurio.

No se arregló, y la razón es que las dos salidas obvias son peores que el problema:

- **condicionar por `huerfanas`** no funciona: cuando el subagente del ⑩ llama a `sf done` ya las
  arregló, así que la lista está vacía justo en ese momento;
- **pasarle un argumento a `sf done`** rompe H2 —el estado, la feature y el lote son deducibles, y
  un argumento deducible es un argumento que se pasa mal.

El costo real es un contador que sube de más; el trabajo no se pierde y `sf next` sigue diciendo
la verdad. Queda anotado acá porque **un residuo escrito es distinto de un residuo olvidado**, que
es la moraleja de A9.

---

# El cierre

## Los seis PRs

| PR | Qué | Tests |
|---|---|---|
| 1 | A0 · A1 · A5 · A6 — los lotes y el camino corto | 213 → 242 |
| 2 | A2 — archivar | 242 → 246 |
| 3 | A3 · A4 — `approve` y el ⑧ | 246 → 252 |
| 4 | A7 — `sf new` y el ⑨ | 252 → 265 |
| 5 | A8 — la cadena de modelo | 265 → 273 |
| 6 | A9 · M1–M3 — la compuerta del ⑩ y los números | 273 → 277 |

## El patrón, que es lo que conviene recordar

**Nueve de los diez hallazgos eran la misma pregunta mal hecha**, en cuatro lugares distintos:

```
¿existe el archivo?     cuando la pregunta era ¿está escrito?      A4 · A7
¿hay algún lote?        cuando la pregunta era ¿empezó alguno?     A0 · A1 · A5
¿aprobó Javier?         cuando la pregunta era ¿y además está bien? A3
¿en qué estado está?    leído crudo en vez de efectivo             A1 · A6
```

Y todos comparten una propiedad que explica por qué sobrevivieron a 213 tests:

> **Ninguno se ve desde adentro de un paquete. Todos viven en la costura entre comandos.**

Por eso el guion de humo del apéndice es el entregable más importante de este documento, y por eso
**dos de los defectos de esta ronda —el sobre vacío del lote de corrección y el `con .` de ME
TRABÉ— los encontró el binario y no los tests.**

## El guion de humo, construido

`sf/cmd/sf/e2e_test.go`, detrás de `//go:build e2e`, y en el CI como paso propio:

```bash
go test -tags e2e ./cmd/sf/
```

Compila el binario de ESTA corrida, arma un proyecto de juguete con git de verdad, y recorre las
dos vueltas afirmando **exit codes** — que ya son parte de la interfaz. Tres tests:

| | Qué recorre | Hallazgos que cubre |
|---|---|---|
| `TestVueltaCompletaDeUnaHistoria` | ⑥ → archivado, con dos lotes y una vuelta del ㉑ | A0 · A2 · A3 · A4 · A5 · A7 · A9 |
| `TestVueltaDeUnBugPorElCaminoCorto` | `sf new` → bug → cierre, salteando dos estados | A1 · A6 · A7 |
| `TestElModeloDeJavierMandaEnLosCuatroEstados` | la cadena de modelo, estado por estado | A8 |

**Cada paso nombra el hallazgo que cubre.** Los marcados con `A#` no están para que la vuelta sea
completa: están porque **ese paso fallaba**. Si alguno se cae, el comentario manda a buscar el
hallazgo en este documento antes de tocar el test.

### Comprobado contra el código roto

No alcanza con que pase: tiene que **fallar cuando corresponde**. Se trajeron los diez archivos de
`cb4d905` —el commit del día de la auditoría— y se corrió:

```
sf (0 passed, 3 failed)
  [FAIL] TestVueltaCompletaDeUnaHistoria       A4 · el ⑧ tiene que pedir trabajo, no la 🛑
  [FAIL] TestVueltaDeUnBugPorElCaminoCorto     el andamio ya no llega
  [FAIL] TestElModeloDeJavierMandaEnLosCuatroEstados
```

Las tres caen en el primer hallazgo que tocan.

### Tres cosas que el guion enseñó al escribirlo

Las tres son la máquina teniendo razón sobre un escenario mal armado, y valen más que el test:

**① Un hallazgo inventado no se puede reproducir.** El primer intento le hacía encontrar al ㉑ un
bug que el código no tenía, y `sf lote start` contestó *"la suite PASA ENTERA: todavía no está
reproducido lo que venís a arreglar"*. Hubo que darle a `Resta` un bug **de verdad** —clampear en
cero— que el test del lote no ve porque `5-3` no es negativo.

**② El lote 2 no puede nacer en verde.** El primer intento implementaba `Suma` y `Resta` juntas en
el lote 1, y el lote 2 arrancaba sin rojo que confirmar. El plan decía dos lotes; el código tenía
que respetarlo.

**③ La rama "falta cerrar la feature" no lanza a nadie**, así que va sin skill ni modelo — y el
test de la cadena lo descubrió pidiendo un modelo donde no hay ninguno. Es correcto: lo que falta
ahí es un `sf done`, no un subagente.

> Que el guion de humo se pelee con quien lo escribe es exactamente lo que se le pide a la máquina.
> Las tres veces, el que estaba equivocado era el test.
