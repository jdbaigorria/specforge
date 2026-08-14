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

## 3. Dónde vive lo nuevo — ✅ `sf/` en la raíz

**Carpeta nueva en el mismo repo, y el viejo se apaga cuando sobra.**

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

## 4. El orden de construcción — ✅ los siete, hechos

**La columna primero**, que ya estaba decidida: `estado.json` + `roadmap.json` + `sf next`.

| # | Qué | Por qué acá |
|---|---|---|
| ✅ **1** | `internal/estado` + `internal/roadmap` | dos parsers y un writer |
| ✅ **2** | **`sf next`** | el primer hito usable |
| ✅ 3 | `sf context` | el sobre, con el material derivado embebido |
| ✅ 4 | `sf done` + las compuertas | acá `sf` empezó a **frenar** |
| ✅ 5 | `approve` · `reject` · `take` · `model` · `dismiss` | las paradas — la vuelta entera |
| ✅ 6 | `sf lote start` | el rojo, la branch y el hash |
| ✅ 7 | `sf new` · `sf status` | las entradas B y C, y la vista |

**El hito del paso 2 es el que importa**: `sf next` diciendo *"planificar f-1, con Opus"* ya se
puede usar en un proyecto real con el resto manual. **Dogfooding desde el segundo paso**, no al
final.

---

### El saldo de los siete

```
sf/                 12 paquetes · 161 tests verdes · go vet limpio
comandos            los 10 del inventario + status + el andamio pendiente
dependencias        UNA: gopkg.in/yaml.v3
migrado del viejo   NADA todavía — cada pieza se escribió al llegar el comando
                    que la pedía, y ninguna pidió el código anterior
```

**Y ése es el hallazgo de la construcción**, porque contradice lo que se esperaba: la
regla de §2 —*se migra cuando un comando lo necesita*— resultó **más filosa de lo previsto**.
Al llegar cada comando, escribir la pieza contra el diseño nuevo salió más corto que adaptar
la vieja. `redwitness.go` (8.4K) se volvió `EmpezarLote` + `internal/suite` (300 líneas con
comentarios), porque la máquina ya aportaba la mitad: el estado, los lotes y la lista de tests
planificados.

> **No es que el código viejo fuera malo: es que sostenía conceptos que la máquina volvió
> innecesarios.** La regla funcionó — sólo que su respuesta, casi siempre, fue "no".

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

## 7. Lo que apareció construyendo, y no estaba en el diseño

Nada de esto sale de los documentos: salió de escribir el código y toparse con el caso.

| Qué | Por qué | Dónde |
|---|---|---|
| **`backlog_visto`** | las 🛑 dejan rastro solas; la ⏸ del ⑨ es un enter y no sella nada. Sin campo, `sf next` la repite para siempre | `estado.go` |
| **`rechazo`** en producto y feature | el motivo del `reject` no se imprime y ya: el que rehace es un subagente NUEVO. Viaja en el sobre | `estado.go` · `sobre.go` |
| **`branch_base`** | adivinar la branch base necesita un remoto, y el flujo funciona sin pushear | `constitucion.go` |
| **`Movio` vs `Cambio`** | un `sf done` que FALLA igual incrementa `intentos_fallidos`. Sin la distinción, el contador de ME TRABÉ nunca subía | `done.go` |
| **la quinta compuerta del ⑰** | una tarea que dice satisfacer `us-1/CA-9` cuando hay dos criterios está mintiendo — el conteo de "sin cubrir" no lo veía | `compuerta.go` |
| **la ⏸ del ㉓ en `sf next`** | mandaba a escribir una doc que ya estaba escrita | `maquina.go` |
| **sin lotes ≠ todos commiteados** | mandaba a cerrar una feature sin una línea escrita, y en el camino corto pasaba **siempre** | `maquina.go` |

---

## 8. Las dos decisiones de forma — ✅ tomadas

1. **`sf/` en la raíz** (§3). Cero colisión de símbolos, el viejo sigue corriendo para comparar,
   y `skills/` · `examples/` · `docs/` · `specs/` se comparten. El switch final es borrar `cli/`.
2. **Modo de trabajo: implemento yo, con comentarios didácticos**, y Javier los lee para
   aprender Go. Es el mismo modo que venía usando desde 2026-06-18.

   **Y acá cambia algo respecto de junio:** entonces se explicaba *Go*. Ahora la mitad de lo que
   hay que explicar es **por qué el diseño quedó así** —por qué `estado.json` no deriva nada, por
   qué `sf context` no lleva argumentos— y eso vive en `specs/`. El comentario apunta al
   documento en vez de repetirlo:

   ```go
   // El estado.json es el ÚNICO archivo que sf escribe.
   // No deriva nada de otros archivos, al revés que el state.go viejo:
   // guarda lo que decidió Javier y lo que sf vio pasar, porque eso no
   // está escrito en ningún lado.   (maquina-estados.md §9)
   ```

### Y sigue aparcado

**Brownfield** (`que-sobrevive.md` §5). Sigue sin bloquear nada.

---

## 9. Lo que falta

**El binario está completo; lo que falta es lo que lo rodea.**

1. **Los skills.** Tres nombres del mapa `estado → skill` son provisionales y están marcados
   con ⚠ en `maquina.go`: `sfp-po` **no existe**, `sf-propose` hay que **partirlo en tres**, y
   falta ver si `sfx-think` cubre `planificacion` entero. Y a los doce que sobreviven hay que
   agregarles el principio y el final (`sf context` · `sf done`) y sacarles las convenciones
   propias (R4).
2. **El `CLAUDE.md` de cuatro líneas**, que es la otra mitad del reparto.
3. **`sf init`** — el scaffold y `detectStack()`, que es lo único del andamio que la máquina
   necesita para arrancar un proyecto de cero.
4. **El mapa de modelos** (`~/.specforge/`), que es lo que le falta al `via:` para resolver
   `consola` y cerrar H1b del todo.
