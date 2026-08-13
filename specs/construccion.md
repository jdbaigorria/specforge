# La construcción

**Fecha:** 2026-08-13 · **Entra:** el diseño completo (`superficie-sf.md` §5 y §6).
**Sale:** por dónde se empieza, qué se migra y con qué criterio.

> **Esto ya no es diseño.** Las decisiones de acá se revisan contra el reloj y contra el
> compilador, no contra el flujo.

---

## 1. La decisión — híbrido, y de Javier

La pregunta abierta era *"podar el CLI actual o partir de cero con lo que sobrevive"*.

> *"yo creo que lo mejor no es ninguna de las dos sino un híbrido: comenzar de cero migrando lo
> que realmente nos sirve, y de allí construir."*

**Y disuelve el dilema en vez de elegir un lado.** Los dos costados que lo hacían difícil eran
reales y ninguno sobrevive a esta forma:

| | El problema que tenía | Por qué acá no pasa |
|---|---|---|
| **podar** | 63 archivos de test cubriendo código que se va, y los conceptos muertos —gates, hook, fases— **enredados** con lo que sirve | no se poda: se arranca vacío |
| **de cero** | tira `redwitness`, `check run` y `context for-wave`, que están bien y costaron | no se tiran: se traen |

---

## 2. El criterio que lo hace funcionar

*"Migrar lo que sirve"* es una intención, no una regla — y sin regla dura degenera en *"copiamos
todo y limpiamos después"*, **que es podar con otro nombre**. La regla:

> ### Nada se migra por existir. Se migra cuando un comando del inventario lo necesita para andar.

**La migración va tirada por la construcción, no al revés.** No se abre `cli/` a ver qué se
salva: se escribe `sf lote start`, y **cuando toca correr los tests y exigir el rojo**, ahí se
mira `redwitness.go` y se decide si entra tal cual, adaptado, o no entra.

Es la misma regla de método que abrió esta refundación (`session.md` §1): *los repos se miran con
la necesidad ya escrita, para diferenciarse, no para surtirse*. Acá se aplica al propio código.

**Y trae dos cosas gratis:**

- lo que **nadie pidió** nunca se copia, y no hay que acordarse de borrarlo después;
- **un test se migra con su pieza**, nunca antes — así no queda cobertura de código que no existe.

---

## 3. Dónde vive lo nuevo

**Recomendación: carpeta nueva en el mismo repo, y el viejo se apaga cuando sobra.**

```
sf/                      ← lo nuevo
  cmd/sf/main.go
  internal/…
cli/                     ← el viejo, intacto, hasta que se apague
skills/  examples/  docs/  specs/     ← se comparten, no se duplican
```

**Por qué no un repo nuevo:** los skills, los ejemplos y toda la doc viven acá y los dos los
comparten. Un repo aparte los duplica desde el día uno.

**Por qué no reemplazar `cli/` en el lugar:** los 51 archivos viejos son **un solo `package
main`**. Escribir un `runNext()` nuevo al lado del viejo **no compila** — colisionan los símbolos.
Con carpeta aparte hay cero colisión, el viejo **sigue compilando y corriendo** mientras se
construye (que es lo que permite comparar contra él), y el switch final es borrar una carpeta.

### Y la estructura de paquetes: los cuatro verbos ya la dibujan

El CLI viejo es `package main` plano: 51 archivos, 18.000 líneas, un `switch` de 39 casos.
**Eso es lo que lo dejó crecer sin fricción hasta duplicar el tamaño de lo que hacía falta.**

Los cuatro verbos de `sf` (`que-sobrevive.md` R1) son casi el árbol:

```
internal/estado/      el estado.json: leer, escribir, mover        ← el único que escribe
internal/roadmap/     el roadmap.json: parsear                     (H19: es un parser)
internal/sobre/       sf context: qué archivos van en cada estado  sirve
internal/compuerta/   correr tests, contar, comparar hashes        comprueba
internal/git/         branch, commit, merge --no-ff                hace
cmd/sf/               el ruteo de los 10 comandos                  expone
```

---

## 4. El orden de construcción

**La columna primero**, que ya estaba decidida: `estado.json` + `roadmap.json` + `sf next`.

| # | Qué | Por qué acá |
|---|---|---|
| **1** | `internal/estado` + `internal/roadmap` | nada funciona sin esto. Son dos parsers y un writer |
| **2** | **`sf next`** | ⬅ **el primer hito real.** Con esto el bucle se puede *mirar* andar aunque todo lo demás sea a mano |
| 3 | `sf context` | el sobre. Es lo que hace útil al subagente |
| 4 | `sf done` + las compuertas de conteo | acá `sf` empieza a **frenar**, que es su segundo verbo |
| 5 | `sf approve` · `reject` · `take` | las paradas: la máquina ya da la vuelta entera |
| 6 | `sf lote start` | el rojo y el hash. Es la joya, y necesita todo lo anterior |
| 7 | `new` · `model` · `dismiss` · `status` | los que contestan a una parada, y la vista |

**El hito del paso 2 es el que importa**: `sf next` diciendo *"planificar f-1, con Opus"* ya se
puede usar en un proyecto real con el resto manual. **Dogfooding desde el segundo paso**, no al
final.

---

## 5. Los candidatos a migrar

**Candidatos, no lista de trabajo** — cada uno se mira recién cuando el paso que lo necesita
llega (§2).

| Archivo del CLI viejo | Lo pide | Estado |
|---|---|---|
| `redwitness.go` | `sf lote start` — el rojo | el mejor texto del repo sobre el dolor #5. **Sin el `CodeHash`**: la compuerta activa lo hace innecesario |
| `context.go` (`for-wave`) | `sf context` | *es el mismo comando con otro nombre*. Cambia qué entra en el sobre |
| `check.go` (`run`) | las compuertas de 4 y 6 | **no como subcomando** (H12): como motor adentro de `done` y `lote start` |
| `mutation.go` | el sobre de `revision` | ya era un contador, no un generador: shell-out + exit code |
| `init.go` (`detectStack`) | `sf init` | llena `lenguaje`, `manifiesto` y `test_cmd` solo (H18) |
| `install.go` | `sf install` | andamio, + las reglas globales |
| `journal.go` | el ㉓ | **sin el rol de progreso** — eso lo hace `estado.json` |
| `feature.go` (archive) | `sf approve` en `cierre` | mover la carpeta entera |

### Y uno que suena a candidato y no lo es: `state.go`

Es el que más se parece por el nombre, y **su premisa está invertida**. Lo dice su propio
comentario:

> *"NO guarda estado nuevo: TODO se deriva de lo que ya existe en `features.json`."*

El `estado.json` del diseño nuevo existe **exactamente por lo contrario** (`maquina-estados.md`
§9): guarda *lo que decidió Javier y lo que `sf` vio pasar* — **porque eso no está escrito en
ningún archivo y no se puede derivar de nada**. El sello del ⑥, el `rojo`, el modelo que estás
usando ahora.

> **Derivar todo era la idea que hacía falta cuando el estado era redundante. Ahora el estado es
> el único que sabe.** Se escribe de cero.

---

## 6. Los tests

11.000 líneas en 63 archivos. **Se migran con su pieza, nunca antes** (§2).

Y hay uno que hay que mirar de frente porque costó y es bueno: **`examples_test.go`**, la suite de
regresión de los tres `examples/`. Cubre los carriles `lite` · `standard` · `brownfield` — y
**los tres son conceptos que la máquina de estados vuelve innecesarios** (R4: *"`lane: lite |
standard` no existe: la máquina saltea estados"*).

**La idea sobrevive, el archivo no:** un `examples/` que corra una vuelta completa de la máquina
—de `sf init` a `sf approve` en `cierre`— es exactamente el mismo activo, contra el flujo nuevo.
Se rehace cuando exista una vuelta que correr, o sea después del paso 5.

---

## 7. Lo que queda por decidir

1. **El nombre y el lugar de la carpeta nueva** — la recomendación es `sf/` en la raíz (§3).
2. **Modo de trabajo.** Javier está aprendiendo Go, y en este repo alternó entre *"escribo yo y
   vos revisás"* y *"implementá con comentarios didácticos y yo lo leo"*. Hay que confirmarlo
   antes del paso 1, porque cambia cómo se escribe cada archivo.
3. **Brownfield** sigue aparcado (`que-sobrevive.md` §5) y sigue sin bloquear nada.
