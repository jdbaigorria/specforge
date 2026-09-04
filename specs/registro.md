# El registro — T0, el tablero

**Fecha:** 2026-09-04 · **Branch:** `refundation` · **Commit:** `5b5dc29`

> **Estado: SPEC, SIN IMPLEMENTAR.** Es la pieza ⓪ de [`por-tramos.md`](por-tramos.md): lo único
> que hay que construir **antes** de correr el primer tramo. Sin esto, cada corrida devuelve
> impresiones en vez de datos.

**Entra:** las cuatro preguntas de Javier sobre cada tramo, los cuatro agujeros medidos el
2026-09-03, y la medición de TTY del 2026-09-04.
**Sale:** un paquete (`internal/registro`), un archivo (`.specforge/registro.jsonl`), un comando
(`sf log`), y cuatro puntos de enganche.

---

## 1. Para qué existe, en una línea

> **`estado.json` guarda la foto. El registro guarda la película.**

Las cuatro preguntas de cada tramo —dónde va un bucle, cuándo corta, cuándo es interactivo, qué
deja cada paso— **son preguntas sobre la película**, y hoy no hay ninguna.

Y hay una segunda razón, que es la que lo justifica más allá de la etapa de pruebas. Dicho con la
frase de Javier:

> *"sf es la máquina de estado pero **no le dimos herramientas para controlar el estado**."*

Traducido a lo que se puede verificar en el código:

```
ARTEFACTOS   ¿existe el archivo? ¿pasan los tests? ¿son tres opciones?
             ¿cambió el hash? ¿vio el rojo antes del verde?
               → acá sf es muy bueno

CONDUCTA     ¿lanzaste el subagente? ¿cuántas vueltas diste?
             ¿quién aprobó? ¿te fuiste a hacer otra cosa?
               → acá sf no ve NADA
```

Los cuatro agujeros del 2026-09-03 son **los cuatro de conducta**. El registro es el órgano que le
falta para verla.

---

## 2. La regla dura de este paquete

> **El registro ESCRIBE. No frena, no lee ninguna compuerta, y no puede romper un comando.**

Las tres partes se sostienen solas:

- **No frena** porque en esta etapa es un instrumento de medición, y un instrumento que altera lo
  que mide no sirve.
- **No lo lee ninguna compuerta** todavía. El día que una lo lea, deja de ser instrumento y pasa a
  ser evidencia — y ahí hacen falta cosas que acá se dejan afuera a propósito (§10).
- **No puede romper un comando**, y esto es lo más importante de las tres: si el disco está lleno,
  si `.specforge/` es de sólo lectura, si el JSON no serializa — `sf next` tiene que seguir
  contestando exactamente lo mismo y con el mismo exit code. **Un error del registro se traga.**

> Un tablero que puede apagar el motor es peor que no tener tablero.

---

## 3. Dónde vive el archivo

```
.specforge/registro.jsonl
```

**En `.specforge/` y no en `.docs/`, y la razón es quién lee cada carpeta** — es el mismo argumento
que ya usó `lanzar/correr.go:17` para las fichas de lanzamiento:

| carpeta | qué es | versionado |
|---|---|---|
| `.docs/` | lo que el proyecto **produce y commitea** — brief, PRD, historias, `estado.json` | **sí** |
| `.specforge/` | lo de **esta máquina**: el catálogo de modelos, los lanzamientos | **no** — `sf init` lo gitignorea |

**Verificado:** `estado.Archivo = ".docs/estado.json"` (versionado) y `arranque.go:117` gitignorea
`.specforge/` entero al sembrar el catálogo.

### Por qué gitignoreado, y cuándo se revisa

Una línea por invocación de `sf`. Versionarlo metería ruido en **cada** commit del proyecto, y su
contenido es de esta máquina (pids, rutas, ids de modelo que no existen en otro lado).

> **Y esto se revisa el día que una compuerta lea el registro.** Ahí pasa de instrumento a
> evidencia, y una evidencia que no viaja con el repo no sirve. Ver §10.

**Lo que no cambia:** el archivo sobrevive en disco. `por-tramos.md` §8 pide conservar el proyecto
de prueba justamente por esto — la corrida de las 7 vueltas no dejó nada, y por eso hoy ese análisis
es deducción y no hecho.

---

## 4. La forma de una línea

Una línea es un objeto JSON completo, sin saltos adentro. Ejemplo real de un `sf done` que cerró un
lote y movió el estado:

```json
{"t":"2026-09-04T13:22:41.118Z","v":1,"cmd":"done","args":["--msg","el parser acepta el flag"],
 "salida":0,"ms":812,
 "quien":"agente","tty":false,"delegado":"","arnes":"claude-code","pid":94730,"ppid":94724,
 "antes":{"feature":"f-2","estado":"implementar","lote":3,"intentos":0},
 "despues":{"feature":"f-2","estado":"revision","lote":0,"intentos":0},
 "tipo":"","skill":"","via":"","perfil":"","modelo":"","esfuerzo":"","lanzamiento":"",
 "movio":true,"fallas":[],"avisos":[],"trunco":false}
```

### Los grupos, y qué contesta cada uno

**① Qué se corrió** — `t` `v` `cmd` `args` `salida` `ms`

`v` es la versión de la forma. Cuesta un entero y evita que un lector futuro tenga que adivinar por
qué faltan campos. `ms` es cuánto tardó: es lo que distingue *"el ㉑ dio dos vueltas"* de *"el ㉑ dio
dos vueltas y la segunda tardó veinte minutos"*.

**② Quién lo corrió** — `quien` `tty` `delegado` `arnes` `pid` `ppid`

Es §5, y tiene su propia sección porque acá es donde hay que ser honesto.

**③ Qué había y qué quedó** — `antes` `despues`

Los dos son la misma forma: `feature` · `estado` · `lote` · `intentos`. Se leen del
`.docs/estado.json` al abrir y al cerrar la entrada, y **el registro los lee él mismo** — no depende
de que el comando haya cargado el estado. Si el archivo no existe (un `sf init`, un proyecto sin
arrancar), los dos van vacíos y no es un error.

> **Y acá hay una objeción que conviene contestar antes de que la haga alguien.** R6 dice *"lo que se
> puede deducir, no se guarda"*, y `lote` **es** deducible (es el primero sin commit). Guardarlo acá
> no la viola, porque **R6 es una regla sobre el ESTADO**, donde un dato repetido se desincroniza.
> Una línea del registro es **inmutable y describe un instante que ya pasó**: no puede
> desincronizarse con nada, y tiene que poder leerse sin el `estado.json` al lado.

**④ Qué dijo `sf`** — `tipo` `skill` `via` `perfil` `modelo` `esfuerzo` `lanzamiento`

Sólo los llena `sf next` (y `lanzamiento` lo llena `sf lanzar`). Es la mitad que contesta *"¿el
subagente que sf pidió se lanzó?"*.

`perfil` y `modelo` van los dos, y no es redundancia: es la misma distinción que `Instruccion` ya
hace —*"qué PIDE el paso"* contra *"qué se le dio en ESTA máquina"*—. Con una sola palabra, dentro
de tres meses las dos preguntas tienen la misma respuesta y no se pueden separar.

**⑤ Cómo salió** — `movio` `fallas` `avisos` `trunco`

`fallas` y `avisos` salen tal cual de `compuerta.Resultado` / `maquina.Efecto`. **Son el dato más
valioso del registro para la etapa de tramos**: son, literalmente, la lista de las veces que un
modelo produjo algo que la compuerta no aceptó.

---

## 5. `quien` — y el límite que ya está medido

### Lo que se midió el 2026-09-04

Se corrió la misma sonda dos veces: una desde la herramienta Bash del agente, otra tipeada por
Javier con `!` en el mismo Claude Code.

| | agente | Javier con `!` |
|---|---|---|
| `[ -t 0 ]` | NO-TTY | **NO-TTY** |
| `tty(1)` | not a tty | **not a tty** |
| padre → abuelo | zsh → claude | **zsh → claude** |
| `AI_AGENT` | `claude-code_2-1-258_agent` | **idéntico** |
| `CLAUDECODE` | `1` | **idéntico** |
| `stdin` | `socket:[…]` | `/dev/null` |

**Una sola diferencia, y el agente la reproduce con un redirect:** `sf approve < /dev/null` da la
salida idéntica. Y `script -qec 'sf approve' /dev/null` **fabrica un TTY** — probado, `/dev/pts/7`.

> **Esto corrige `vecinos.md` §2**, que decía *"la primitiva para cerrarlo ya está construida"*
> refiriéndose a `hayPersona()`. **No alcanza:** `script(1)` la vence, y viene en cualquier Linux.

### La consecuencia sobre el diseño

**Tres valores, y el tercero no prueba lo que su nombre sugiere:**

| valor | cómo se decide | qué significa de verdad |
|---|---|---|
| `delegado` | `SPECFORGE_DELEGADO` está puesto | lo corrió un hijo de `sf lanzar`. **Es el único fiable**, porque `sf` mismo puso esa marca |
| `terminal` | hay TTY **y** no hay marca de arnés | se corrió desde una terminal de verdad, fuera de un agente |
| `agente` | todo lo demás | **NO distingue al orquestador de Javier tipeando `!`** |

> **El tercero es un límite medido, no un pendiente.** Nadie lo va a "completar" después: mientras el
> agente controle el arranque de procesos, ningún mecanismo en banda puede probar que había un
> humano. Está dicho acá para que quien lea `quien: "agente"` sepa exactamente qué está leyendo.

### Por eso se guardan las señales crudas y no sólo la etiqueta

`tty`, `delegado`, `arnes`, `pid` y `ppid` van al archivo **además** de `quien`. Dos motivos:

1. **La etiqueta es una interpretación; las señales son el hecho.** El día que la regla cambie, las
   líneas viejas se pueden releer con la regla nueva. Con sólo la etiqueta, no.
2. **`arnes` atrapa lo que el TTY no.** Un `script` que fabrica un TTY **no desactiva `CLAUDECODE=1`**
   — así que una línea con `tty: true` y `arnes: "claude-code"` es exactamente la firma de esa
   evasión. No la impide; la deja escrita.

**`arnes` no se construye: ya existe.** Es `global.DetectarHarness()`, que mira siete variables
medidas —las de opencode y Command Code se obtuvieron volcando el entorno de cada uno— y contesta
`"desconocido"` cuando ninguna está. Su propio comentario dice lo que hay que saber:

> *"Es una **PISTA**, no una certeza. Que la variable esté es evidencia de que sí; que NO esté no
> prueba nada."*

O sea que el campo hereda la honestidad que esa función ya tiene, y **el registro no agrega ninguna
detección nueva**. Un detalle que la mejora sin costo: `sf lanzar` le pone `SPECFORGE_HARNESS` al
hijo, y esa variable gana sobre la heurística — así que en una corrida delegada el `arnes` es un
dato declarado y no adivinado.

---

## 6. Los cuatro puntos de enganche

### ① `main()` abre y cierra — y es el único cambio estructural

Hoy cada rama del `switch` hace `os.Exit(fn())`, y un `os.Exit` se saltea todos los `defer`. Para
que el registro se escriba **siempre**, main se parte en dos:

```go
func main() {
    codigo := despachar()          // el switch de hoy, con `return` en vez de os.Exit
    registro.Cerrar(codigo)        // lee el estado de después y escribe la línea
    os.Exit(codigo)
}
```

Y `registro.Abrir` va arriba de `despachar`, con los args crudos.

**Por qué acá y no en cada comando.** Enganchar en los veinte comandos serían veinte copias de la
misma llamada, y agregar el comando veintiuno sin acordarse dejaría un hueco silencioso. Es la misma
razón por la que existe el paquete `comandos`: *"dos copias ya eran una de más"*.

**Qué cuesta:** cambiar ~20 `os.Exit(x)` por `return x`. Es mecánico y no cambia ningún
comportamiento. Y ya hay red: el e2e que le pasa cada nombre de `comandos.Todos` al binario seguiría
comprobando que ninguno conteste *"todavía no está construido"*.

### ② `next()` anota lo que `sf` dijo

Después de `maquina.Siguiente()`, en una sola llamada:

```go
i := maquina.Siguiente(raiz, e, r, g)
registro.Instruccion(i)            // tipo, estado, feature, skill, perfil, modelo, esfuerzo, via
fmt.Print(mostrar(i))
```

### ③ `terminar()` y `parada()` anotan el veredicto

Las fallas y los avisos son el dato más valioso del registro, y hoy se imprimen y se pierden:

```go
c := maquina.Terminar(raiz, e, r, msg)
registro.Veredicto(c.Fallas, c.Avisos, c.Movio)
```

Lo mismo en `parada()` con el `Efecto`. Son dos sitios y cada uno tiene un motivo obvio.

### ④ `lanzarPaso()` ata la corrida a su ficha

```go
registro.Lanzamiento(ficha.ID)
```

Es lo que permite contestar *"entre este `next` que dijo subagente y el `done` que vino después,
¿hubo una corrida?"*.

### Por qué una variable de paquete y no un puntero que viaje

`registro` guarda **la entrada en curso** en una variable de paquete. Es aceptable acá por un hecho
del programa, no por comodidad: **`sf` es un proceso, un comando, una entrada.** No hay dos en
vuelo. Enhebrar un puntero por veinte funciones para un asunto lateral ensuciaría firmas que hoy
están limpias.

---

## 7. Cómo se escribe — las dos decisiones que no son obvias

### ① Una sola llamada a `Write`, con la línea entera

```go
f, err := os.OpenFile(ruta, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0o644)
// … un único f.Write(append(linea, '\n'))
```

`O_APPEND` hace que el sistema mueva el puntero al final **y** escriba en una sola operación, así que
dos `sf` corriendo a la vez no se pisan a mitad de línea.

> **No lo medí, y lo digo:** esto es lo que dice POSIX para `O_APPEND` y lo que hace Linux en la
> práctica. No hice el experimento de dos procesos escribiendo en paralelo. Si alguna vez aparece una
> línea partida en el archivo, **esta es la primera sospechosa**.

Y de ahí sale el tope del punto siguiente, que existe justamente para que esa suposición no se
estire más de lo que aguanta.

### ② El tope de 4 KB, y qué se corta cuando se pasa

Una línea no puede pasar de **4096 bytes**. Los campos de largo impredecible son tres —`args` (un
`sf done --msg "…"` o un `sf new "…"` pueden ser largos), `fallas` y `avisos`— y cuando la línea se
pasa, **se recortan esos tres, en ese orden, y `trunco` queda en `true`**.

Nunca se recorta ni se tira una línea entera: **una línea que falta es un hueco invisible en la
película**, y es justo lo que este archivo existe para no tener.

### ③ Cuándo NO se escribe nada

Si al cerrar no existe **ni `.docs/` ni `.specforge/`**, no se escribe. Es la forma de que un
`sf version` tipeado en una carpeta cualquiera no deje basura.

**Notar que la comprobación es al CERRAR y no al abrir**, y eso resuelve solo el caso de `sf init`:
cuando la entrada se cierra, las carpetas ya las creó el propio comando, así que **`sf init` queda
registrado** — que es lo correcto, porque es la primera línea de la película.

---

## 8. `sf log` — la superficie de lectura

Es **el único comando nuevo**, y existe porque un JSONL a ojo no se lee. Va a `comandos.Todos`
—hay un test que compara esa lista contra el `switch`— y a la ayuda.

```
sf log                      las últimas 20, una línea legible cada una
sf log --ultimas N          cuántas
sf log --feature f-2        sólo las de esa feature
sf log --cmd done           sólo ese comando
sf log --vueltas            el conteo por estado — la pregunta ① de cada tramo
sf log --json               crudo, para jq
```

Formato legible, una línea por entrada:

```
13:22:41  done      f-2  implementar → revision   ✓  812ms   agente
13:31:07  next      f-2  revision                 ⏵  sf-check · opus-grande · subagente
13:58:02  done      f-2  revision → implementar   ✗  hay 2 hallazgos abiertos: h-1 · h-3
```

Y `--vueltas`, que es la que contesta la pregunta que hoy le creemos al revisor:

```
f-2   planificacion  1
      implementar    8      ← 5 de ellas entrando desde revision
      revision       7      ← el bucle
      cierre         0
```

> **Ese `7` es el número que hoy escribe el propio `sf-check` en el campo `vuelta`, y que `sf` nunca
> comprueba.** Con esto, `sf` lo cuenta él. Es la regla del proyecto —*"sf no le cree al que
> trabajó"*— aplicada al único lugar donde todavía le cree.

**Exit code: siempre 0.** Es un lector. Un `2` significaría *"parada, es de Javier"* y confundiría al
orquestador que lo corriera dentro del bucle.

---

## 9. Los tests

Seis, y ninguno necesita un modelo:

1. **Una línea por invocación**, para cada nombre de `comandos.Todos` — reusando el e2e que ya los
   recorre.
2. **El registro no puede romper un comando.** Con `.specforge/` en sólo lectura, `sf next` devuelve
   el mismo texto y el mismo exit code que sin registro. **Es el test que más importa** (§2).
3. **`antes` ≠ `despues`** en un `sf done` que mueve el estado, e **iguales** en un `sf next`, que es
   consulta pura.
4. **`quien`**: con `SPECFORGE_DELEGADO` puesto → `delegado`. Sin TTY y sin marca → `agente`. Con TTY
   simulado y sin marca de arnés → `terminal`.
5. **El tope**: un `sf new` con un texto de 8 KB produce una línea de ≤ 4096 bytes, con
   `trunco: true`, **y la línea existe**.
6. **`sf log --vueltas`** cuenta lo mismo que el e2e hizo pasar — o sea, se compara contra un
   recorrido conocido y no contra sí mismo.

---

## 10. Lo que T0 NO trae, y qué lo despertaría

Cada cosa diferida con su umbral escrito, que es la disciplina que `vecinos.md` §8.3 tomó de `specd`.

| Diferido | Por qué | Qué lo despierta |
|---|---|---|
| **cadena de hash** | mientras sea instrumento, nadie falsifica sus propios datos de prueba | **la primera compuerta que lea el registro.** Ahí pasa a ser evidencia, y hace falta el encadenado + un ancla del último hash en `estado.json` |
| **versionar el archivo** | ruido en cada commit, y contenido de esta máquina | lo mismo de arriba: una evidencia que no viaja con el repo no sirve |
| **rotación** | ~50–200 líneas por feature. Es nada | un proyecto real donde el archivo moleste |
| **que una compuerta lo lea** | es el paso siguiente, no éste | los datos de T5 y T6: **recién ahí se sabe en qué vuelta hay que cortar**, y cablear un número antes de tenerlo es inventarlo |
| **`--async` / dos `sf` en paralelo** | fuera de alcance en `headless.md` §10 | si aparece, trae consigo el número de fila de `vecinos.md` ③ |

> **La más importante es la cuarta.** La tentación va a ser meter el corte del bucle ahora, porque ya
> tenemos el conteo. **No:** el conteo es para *descubrir* el número, y el número sale de mirar las
> corridas de T6. Ponerlo antes es exactamente el error de método que `por-tramos.md` §1 viene a
> arreglar.

---

## 11. Cómo se sabrá que quedó bien

```bash
sf next && sf log --ultimas 1
#   → una línea con estado, tipo, skill, via, modelo y quien

chmod -w .specforge && sf next
#   → contesta lo mismo, mismo exit code, y NO rompe

SPECFORGE_DELEGADO=l-7 sf status && sf log --ultimas 1 --json | grep '"quien":"delegado"'
#   → la marca del hijo queda escrita
```

Y la prueba que de verdad lo cierra, que es la que no se puede hacer con un test:

> **Correr T1 completo y contestar, mirando sólo `sf log`:** cuántas vueltas dio el pinponeo del
> brief, cuánto tardó cada una, qué compuerta falló y cuántas veces, y con qué modelo.
>
> Si esas cuatro se pueden contestar sin abrir otra cosa, el tablero está.
