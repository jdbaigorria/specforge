# `sf install` interactivo — los arneses y los modelos, de una sola vez

**Fecha:** 2026-08-31 · **Branch:** `refundation` · **Commit:** `1aa81dd`

> **Estado: DOS DE TRES IMPLEMENTADOS** (2026-08-31). El ② (`sf models`) y el ① (`sf install
> --harness=a,b,c`, más la detección de instalados) están construidos y corriendo — el detalle al
> final, en **Lo que se implementó**. Falta el ③, la pregunta, que es el único que necesitaba las
> otras dos hechas.
>
> El resto sigue siendo spec de diseño, y la **medición** —el TTY, los tres listados de modelos y
> las dependencias de `sf`— se comprobó ejecutando el 2026-08-31. Nada acá es supuesto.
>
> **No depende de [`headless.md`](headless.md), pero headless depende de ésta.** Sirve hoy, con el
> modo orquestado que ya existe, y después pasa a ser **de dónde `sf lanzar` saca el arnés, el
> modelo y el esfuerzo** — está en §7 y es de Javier. O sea que se puede hacer antes o en paralelo,
> pero ya no después.

---

## 1. El dolor

Javier, después de correr el bucle en los tres arneses el 2026-08-31:

> *"parece que sí anda cambiando de harness, lo cual es bueno pero es medio enquilombado […] y el
> tema de los modelos de cada uno, debería ser más sencillo."*

Y lo que hay que hacer hoy para dejar dos arneses andando, en orden:

```bash
# en Claude Code
sf install
sf model razonar   --alias grande --id opus   --via subagente
sf model construir --alias medio  --id sonnet --via subagente
# cerrás, abrís opencode
sf install
sf model razonar   --alias ultra  --id opencode/nemotron-3-ultra-free --via subagente
sf model construir --alias rapido --id opencode/…                     --via subagente
# ⚠ reiniciá tu harness: los agentes se leen al arrancar
# cerrás opencode, lo volvés a abrir
```

**Ocho comandos, dos reinicios y un ida y vuelta entre arneses** para decir dos cosas: cuáles usás
y con qué modelo. Y los ids hay que sacarlos de otro lado y tipearlos.

Lo que propone esta spec:

```bash
sf install
```

---

## 2. La línea: quién conversa y quién no

`sf` **nunca** conversa. Es la regla que lo hace usable por un agente: `sf next` es consulta pura,
`sf done` contesta y se va, ninguno espera a nadie.

`sf install` es la excepción, y la excepción tiene un fundamento que ya existía: **`sf install` no
es de la máquina.** No mira el estado ni lo mueve, no aparece en ningún trazado del bucle
(`andamio.go`, encabezado del paquete). Es lo que hace que SpecForge se pueda *usar*, y lo corre una
persona, una vez.

### Y no hay que elegir, porque se puede saber

**Medido el 2026-08-31**: la shell de un agente **no tiene TTY**; la de una persona sí.

```
adentro de la herramienta Bash de Claude Code   stdin: NO es TTY · stdout: NO es TTY
en una terminal de verdad                       stdin: SÍ es TTY
```

Así que la regla es una llamada, no una arquitectura:

```go
// hayPersona dice si del otro lado hay alguien a quien preguntarle.
//
// Sin esto, un `sf install` corrido por un agente —el orquestador lo nombra en
// "Empezar de cero", y `sf doctor` lo sugiere— colgaría esperando una respuesta
// que nadie va a escribir.
func hayPersona() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
```

**Con la stdlib sola, y eso importa:** `sf` tiene hoy **una** dependencia (`gopkg.in/yaml.v3`).
Meter `golang.org/x/term` para preguntar si hay una terminal la duplicaría. Probado en los dos
sentidos y anda.

| | quién lo corre | conversa |
|---|---|---|
| `sf install` | Javier, una vez | **sí, si hay TTY** |
| todo lo demás | el orquestador, siempre | no, nunca |

> **Los flags no son el salvavidas: son el mecanismo.** `sf install --harness=a,b,c` es cómo se
> scriptea y cómo debería sugerirlo `sf doctor` —que hoy dice `sf install` pelado
> (`doctor.go:209`)—. Si vienen flags, no se pregunta nada, haya TTY o no.

---

## 3. Lo que se midió — los tres listados

**Dos de tres pueden enumerar sus modelos**, y uno regala las descripciones.

| | comando | qué devuelve |
|---|---|---|
| **opencode** | `opencode models [provider]` | **397** ids, uno por línea. Acepta filtro por proveedor. Sin JSON. |
| **Command Code** | `cmd --list-models` | **62**, agrupados por proveedor, **con descripción**. Sin JSON (`--json` se ignora). |
| **Claude Code** | *no hay* | el `--help` nombra los alias; el error de un modelo inválido no lista. |

Y así se ve el de Command Code:

```
Open Source

deepseek/deepseek-v4-flash    fast hybrid-attention reasoning (default)
moonshotai/kimi-k3            long-horizon coding & knowledge work with 1M context
zai-org/glm-5.2               powerful coding with 1M context and long-horizon tasks
```

**Esa segunda columna es `capacidad` y `para`** — los dos campos que el catálogo ya tiene para que
el ⑯ elija, y que hoy Javier escribe a mano. El arnés los está regalando.

> **Y `sf` no los interpreta.** Los copia. Es R3 sin discusión: "long-horizon coding" es la opinión
> del proveedor sobre su modelo, y `sf` la transporta igual que transporta un id. El día que el ⑯
> elija con eso, está eligiendo con lo que dijo el proveedor y lo que corrigió Javier — nunca con
> algo que `sf` dedujo.

---

## 4. La forma

```
$ sf install

¿Qué arneses vas a usar acá?
  [x] claude-code     (detectado)
  [ ] opencode
  [ ] commandcode

── opencode ────────────────────────────────────
razonar   — el que compara y juzga (⑫ y ㉑)
  1  opencode/nemotron-3-ultra-free
  2  opencode/nemotron-3.5-lightning-free
  …  397 en total — escribí para filtrar, o pegá un id
> 1
  alias corto para éste: ultra
  esfuerzo (enter = el del modelo): high

construir — el que escribe el código (el resto)
> …

── commandcode ─────────────────────────────────
razonar
  1  minimaxai/minimax-m3     frontier coding, agents & native multimodality
  2  zai-org/glm-5.2          powerful coding with 1M context and long-horizon tasks
> …

✓ 3 arneses · 6 perfiles declarados
  + .claude/settings.json
  + .opencode/agents/sf-ultra.md · sf-rapido.md
  + .commandcode/settings.json · agents/sf-…
```

Tres cosas de esa pantalla que no son cosméticas:

1. **`mecanico` no se pregunta.** Ningún estado lo pide (`perfilPorEstado`); lo nombran los `sfx-*`,
   que están fuera de los nueve. Preguntarlo sería pedir una decisión que no frena nada.
2. **El alias lo pone Javier, no `sf`.** Es su vocabulario (H3). Derivarlo del id —`nemotron-3-ultra-free`
   → `ultra`— sería `sf` opinando sobre cómo se llaman las cosas de él.
3. **El esfuerzo se pregunta pero admite enter.** Vacío es "el que traiga el modelo", que **no** es
   lo mismo que "bajo", y esa distinción ya está en el código.

---

## 5. Los tres pedazos, y sirven sueltos

### ① `sf install --harness=a,b,c`

Hoy `Opciones.Harness` es un string y `Instalar` arma el andamio de uno. Pasa a ser una lista.

**No hay cambio de esquema en el catálogo**: `harnesses:` ya está indexado por arnés desde H2. Es lo
que menos toca de los tres.

> **El hueco, y cómo quedó tapado.** `global.Config.Harness` —el singular— está documentado como
> *"el ÚLTIMO que instaló `sf install --harness=…`"*, y es el fallback de `EnUso()` cuando no hay
> variable de entorno ni detección: una terminal pelada, un cron. Instalando tres de una, **"el
> último" deja de significar algo**.
>
> **Decidido (Javier, 2026-08-31): gana donde estás parado**, si está entre los pedidos — es el
> único de la lista sobre el que hay un hecho y no una preferencia. Si no está —instalás desde
> Claude Code para dejar listos los otros dos—, cae al **primero que nombró Javier**. No se ordena
> por gusto de `sf`. Está en `andamio.puntero()` y lo fijan dos tests.

### ② `sf models [--harness=X]`

Corre el listado del arnés y lo devuelve normalizado: `id`, y `descripción` si el arnés la da.

Sirve suelto: hoy, para declarar un modelo, hay que ir a buscar el id a otra ventana — `sf doctor`
literalmente te manda a `opencode models` o a `/model`.

### ③ La pregunta

Sólo tiene sentido con ① y ② hechos. Usa ② para el menú y ① para escribir.

---

## 6. El premio que no es obvio: se muere el reinicio

**Medido el 2026-08-29 (sonda 1):** las definiciones de agente se leen **al arrancar**. No hay
recarga en caliente. Por eso `sf model` avisa *"reiniciá tu harness"*, y por eso hoy el ciclo tiene
un cierre-y-abrí en el medio.

Ese aviso existe porque hoy declarás **estando adentro** del arnés que vas a usar.

Si declarás los tres de una, **los portamodelo de opencode y de Command Code quedan escritos antes
de que abras esos arneses**. Cuando los abrís, ya están cargados.

> No es que la molestia se reduzca: **el paso desaparece.** El aviso sigue existiendo para el
> `sf model` suelto —el ⑳, cuando el bucle patina y subís el modelo en caliente—, que es el único
> caso donde de verdad estás adentro.

---

## 7. Para qué termina sirviendo — es el catálogo que `sf lanzar` va a leer

**Lo agregó Javier el 2026-08-31, y le da a esta spec un consumidor que antes no tenía.**

> *"el tema de arnés y modelos lo podemos resolver con el install: cuando instalás seleccionás los
> arneses que vas a usar y los modelos que querés tener disponibles, entonces cuando el agente
> principal lanza el headless vas a saber el arnés, el modelo y el effort que necesitás."*

Hasta acá esta spec se defendía sola por comodidad: ocho comandos y dos reinicios pasan a ser uno.
Es cierto y alcanza. Pero el motivo de fondo es otro y es más fuerte:

```
sf install   →  escribe el catálogo   →  sf lanzar lo lee
(una vez, con Javier)                    (cada paso, sin preguntar nada)
```

En el modo headless ([`headless.md`](headless.md) §4), el arnés principal delega un paso diciendo
apenas esto:

```bash
sf lanzar --harness=opencode --alias=ultra
```

y **no tiene que saber ningún id, ningún flag ni ninguna sintaxis**. El id, el esfuerzo y el `via`
ya están escritos, y los escribió esta pantalla. Sin este paso, headless obliga a tipear un id de
proveedor en cada delegación — que es el dolor de §1 mudado de lugar, no resuelto.

**Tres consecuencias concretas:**

1. **`sf models` (el ② de §5) sube de prioridad.** Deja de ser una comodidad para tipear menos y
   pasa a ser **cómo se llena** lo que headless consume. Es la pieza que sobrevive intacta a los dos
   escenarios, así que es por donde conviene empezar.
2. **El esfuerzo deja de ser decorativo.** Hoy se escribe adentro del portamodelo con un nombre de
   campo distinto por arnés; en headless es un flag (`--effort` en Claude Code y Command Code,
   `--variant` en opencode). Preguntarlo acá es lo que hace que después se pueda pasar.
3. **El catálogo no necesita campos nuevos.** `global.Modelo` ya guarda `id`, alias, `esfuerzo` y
   `via`, indexado por arnés y por perfil. Es exactamente el juego de datos que hace falta. Lo único
   que se agrega vive en `sf`, no en el catálogo: la tabla de cómo se escribe cada cosa en cada
   arnés, que son tres líneas.

> **Y esto no rompe la regla de §2.** `sf install` sigue sin ser de la máquina: escribe
> **configuración**, no estado. Que headless la lea después no la convierte en un paso del bucle.

---

## 8. Lo que NO cambia

- **`sf install` sigue sin ser de la máquina.** No mira el estado ni lo mueve.
- **Los flags mandan.** Con `--harness=`, no pregunta, haya TTY o no.
- **Sin TTY se comporta exactamente como hoy**: detecta e instala uno. Cero regresión para el
  orquestador y para `install.sh`.
- **`sf` no elige un modelo ni traduce un id de un proveedor a otro** (R3). Lista, y Javier elige.
- **No se pisa nada.** La regla de `CLAUDE.md`/`AGENTS.md` y de `~/.specforge/` sigue igual, con la
  marca de versión que se agregó en `be72b61`.

---

## 9. El riesgo, y cómo se acota

**Parsear una salida hecha para humanos es frágil.** `cmd --list-models` tiene encabezado,
secciones por proveedor, dos columnas alineadas con espacios y un pie con un link. Ninguno de los
dos ofrece JSON. Eso se rompe el día que cambien el formato, y se rompe en silencio.

La mitigación no es parsear mejor: es **que el listado sea una comodidad y nunca el mecanismo.**

```
el menú falla  →  "no pude listar los modelos de opencode. Pegá el id:"  →  seguís
```

Tipear el id tiene que funcionar siempre, y tiene que ser un camino de primera clase, no un
fallback vergonzante. Si `sf` queda inservible porque un arnés cambió una tabla, el diseño está mal.

> **Corolario para `sf doctor`:** que sepa decir "el listado de X no se pudo leer" como **aviso**,
> nunca como falla. No poder mostrar un menú no impide trabajar.

### Y el otro riesgo, que es de alcance y no de parseo

El dibujo de §4 muestra casillitas `[x]` que se marcan y un *"escribí para filtrar"* sobre 397 ids.
**Eso es una pantalla viva**, que reacciona tecla por tecla, y en Go eso pide **modo raw de
terminal** — o sea `golang.org/x/term`, que es exactamente la dependencia que §2 se enorgullece de
no agregar cuando justifica `hayPersona()` con la stdlib sola.

No se puede tener las dos. Y la salida barata existe y no es fea:

```
te muestro la lista numerada  →  escribís un número, o un texto para filtrar  →  enter
```

Con eso alcanza `bufio.Scanner`, cero dependencias, y **funciona igual por un pipe**, que es lo que
la prueba 1 de §10 exige. La pantalla viva es más linda; la pantalla numerada es la que cabe en las
reglas que esta misma spec se puso.

> **Decidido (Javier, 2026-08-31): la numerada.** La diferencia real de experiencia es una tecla
> —el enter después del filtro— y lo que se compra es que haya UN SOLO camino: el mismo código
> atiende a una persona y a un pipe. El dibujo de §4 hay que leerlo como el contenido de la
> pregunta, no como su forma.

---

## 10. Cómo se sabrá que quedó bien

```bash
# la instalación entera, en un solo comando y sin cambiar de ventana
sf install --harness=claude-code,opencode,commandcode
sf models --harness=opencode | head          # ids, y descripción si la hay
```

Y las cuatro pruebas que cierran el diseño:

1. **Sin TTY se porta como hoy.** `sf install < /dev/null` en un pipe: detecta uno, no pregunta
   nada, no cuelga. Es el caso del orquestador y el que no puede fallar nunca.
2. **Con flags no pregunta**, aunque haya una persona mirando.
3. **Sin reinicio.** Declarar los tres arneses desde Claude Code, abrir opencode **por primera vez**
   y que `sf next` resuelva su modelo y su `agente:` **sin haber reiniciado nada**. Es la prueba de
   §6, y es la que se ve.
4. **Con el listado roto igual se puede.** Simular que `opencode models` falla y comprobar que la
   instalación sigue pidiendo el id a mano y termina bien.

> **La prueba de que quedó bien no es que la pantalla sea linda: es el punto 3.** Ocho comandos, dos
> reinicios y un ida y vuelta entre ventanas tienen que quedar en uno.

---

## 11. Lo que se implementó

**2026-08-31.** Dos de los tres pedazos de §5, más la detección que pidió Javier. 403 tests contra
los 380 del día anterior, más los 7 e2e; `gofmt`, `vet` y `build` limpios.

### ② `sf models` — commit `23bfa52`

Paquete nuevo `sf/internal/modelos`. Un parser por arnés porque los formatos son distintos, y los
dos verificados **contra la salida real y no contra un fixture**: opencode 397 ids, Command Code
61 — y su propio encabezado dice "61 models".

Tres decisiones que quedaron en el código, con su motivo:

- **Sin tabla de modelos para Claude Code.** Son cuatro alias y era tentador escribirlos. Es la
  tabla que se pudre que `global.Modelo` ya se prohíbe. Devuelve `ErrSinListado` y ofrece pegar el id.
- **Dos errores distintos**, y la distinción salió corriendo el binario, no leyéndolo: la primera
  versión le contestaba *"buscá el id a mano"* a un `--harness=emacs`, que es un consejo inútil para
  un typo. Ahora `ErrSinListado` sale con 2 y el camino manual, y `ErrArnesDesconocido` con 1 y la
  lista de los que hay.
- **`Listar` prefiere fallar a devolver una lista dudosa.** Cero modelos no es un listado vacío
  —los dos que listan tienen cientos—: es que el formato cambió.

De paso quedó medido que **el listado cambia solo**: Command Code se autoactualizó `1.38.2 → 1.39.2`
durante las pruebas y pasó de 63 modelos a 61. Es el argumento de §8, ahora con un caso.

### ① `sf install --harness=a,b,c` + la detección

`Opciones.Harness` pasó a ser una lista; el orquestador (①) y la semilla del catálogo (②) siguen
corriendo **una vez**, y sólo el andamio por arnés (③) se repite. `Resultado` cambió de
`Harness string` + `Modelos int` a `Para []string` + `Modelos map[string]int` + `Puntero string`,
que es la distinción que el comando ahora necesita nombrar.

`--harness` acepta coma y además es repetible, porque las dos formas aparecen solas: una la tipeás
y la otra sale de un script que arma la lista en un bucle.

**`andamio.Instalados()`** mira el PATH y le pregunta a cada binario quién es. Y ahí apareció una
asimetría que hubo que medir: **sólo Claude Code se identifica en `--version`.**

```
claude --version     "2.1.252 (Claude Code)"    dice su nombre
cmd --help           "Command Code v1.39.2"     dice su nombre, pero en --help
opencode --version   "1.18.25"                  NO dice su nombre
```

Por eso la confirmación de opencode es más floja —que el comando ande— y no una marca en la salida.
Inventarle una marca que no emite sería adivinar. El que justifica que haya confirmación es Command
Code: su binario se llama `cmd`, que en Windows es la shell.

Hoy la detección se usa en un solo lugar y ya paga: `sf install` pelado te dice qué otros arneses
tenés y te da el comando exacto.

```
También tenés opencode y commandcode en esta máquina.
    sf install --harness=claude-code,opencode,commandcode
```

**No se instalan solos**, y eso es §7: sin flags, `sf install` sigue armando el andamio de UNO. Es
lo que corre el orquestador y lo que corre `install.sh`, y un test de no regresión lo fija.

### Dos cosas que aparecieron construyendo, y ninguna se buscó

**Un arnés que `sf` no sabe preparar se aceptaba en silencio.** `--harness=codex` es deliberado
—`sf` transporta lo que Javier declara y no le pone una lista blanca (R3)— pero `escribirPermisos` y
`escribirPortamodelos` **no hacen nada** con un nombre que no está en sus mapas: la instalación
terminaba con un ✓ y el arnés sin permisos, y el bucle se trababa después, en el primer subagente.
Se aceptó igual, pero ahora lo dice. Casi se arregla rompiéndolo —rechazándolo—, y lo frenó un test
que ya existía y decía que aceptarlo era a propósito.

**El andamio salía del catálogo equivocado, y eso rompía la promesa de §6.** `Instalar` leía siempre
el catálogo GLOBAL, pero `sf init` siembra uno **por proyecto** y desde ahí ése es el que rige
(`LeerPara` lo prefiere, `sf model` escribe ahí). No se veía mientras `sf install` corría una vez y
antes de `sf init`. Se ve con `--harness=a,b,c`, que existe justamente para **reinstalar** sobre un
proyecto que ya anda: escribía cero portamodelo y decía "0 modelos declarados" mientras el proyecto
tenía los suyos. O sea que el "se muere el reinicio" no pasaba, y no pasaba en silencio.

> Se encontró corriendo el binario contra un proyecto de verdad, no en la suite. Es el mismo tipo de
> hallazgo que `arreglos.md` §A2, y la misma lección: la costura entre comandos no se ve desde
> adentro de un paquete.

Con las dos arregladas, la prueba 3 de §9 **pasa de verdad**: desde una sesión de Claude Code, un
solo comando dejó escritos los permisos de opencode y de Command Code y sus tres portamodelo, sin
abrir ninguno de los dos.

### Lo que falta

El ③, la pregunta. Ahora sí tiene con qué: `sf models` para el menú y `--harness=a,b,c` para
escribir.
