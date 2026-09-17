# El mercado — dónde está parado `sf`, y por qué se siente que no avanza

**Fecha:** 2026-09-12 · **Branch:** `claude/specforge-alternatives-direction-ae84u8`

> **Estado: DIAGNÓSTICO.** No propone código. `vecinos.md` salió de leer dos repos que llegaron al
> mismo diagnóstico; éste sale de mirar **la categoría entera** y contestar una pregunta distinta:
> no *"qué nos falta"* sino **"qué de todo esto hay que dejar de construir"**.

---

## 1. El síntoma, medido

Javier dice que siente que se está enroscando. No es una sensación: está en el repo.

| | |
|---|---|
| commits en `refundation` | **50**, en 15 días (primero: 2026-08-29) |
| código Go, sin tests | **15.133** líneas |
| documentos de diseño en `specs/` | **30 archivos · 17.302 líneas** |
| skills | 27 archivos · 5.948 líneas |
| corridas reales con agentes | **3 registradas** (T1 el 09-05, el 09-08, y la de 7 vueltas) |
| releases públicos | **0** |

> **Hay más líneas de diseño que de producto.** 17.302 contra 15.133. Lo mismo que mide
> `el-mapa.md` §1 con otras palabras: *"no es un proyecto que se empieza de nuevo, es un proyecto
> que perdió el índice"*.

**Corrección del 2026-09-13.** La primera versión de este documento decía *"cero corridas reales"* y
*"el bucle de feedback está cerrado sobre sí mismo"*. **Las dos son falsas**, y la evidencia está en
`el-arnes-ciego.md`: la corrida de T1 del 2026-09-05 midió dos modelos sobre la misma semilla —
`no-lo-hagas` con 13 links contra `hacelo` con 0— y `compuerta.Brief` aceptó los dos. De ahí salió
la regla del proyecto y el arreglo `4d7cc28`. Los tres hallazgos más caros del repo salieron de
**correr**, no de leer. Las corridas viven fuera del árbol (`banco/diagnosticos/` está en
`.gitignore`) y por eso no se vieron.

**Lo que sí sobrevive a la corrección** es lo de abajo: el plan T0–T7 no tiene renglón de publicar,
y `salir-a-la-cancha.md` ① quedó huérfano cuando `el-mapa.md` reemplazó a `retomar.md` como puerta
de entrada.

Y el dato que lo vuelve urgente en vez de filosófico:

```
2026-09-01 · salir-a-la-cancha.md ①:  origin/main...refundation  →  0 del lado de main
2026-09-12 · medido hoy:              origin/main...refundation  →  14 del lado de main
```

**Lo que el 1 de septiembre era un fast-forward limpio, hoy son catorce commits de divergencia.**
El bloqueo que aquel documento llamó *"el más grande y el más barato"* se encareció solo por
esperar, y se va a seguir encareciendo. Es el único costo del repo que crece sin que nadie trabaje.

### 1.1 De dónde sale cada compuerta

Las cinco de `vecinos.md` salieron de **leer el código de `specd`**. Las de `arreglos.md`, de leer
el código propio. El ⓪ de `banco/GUION.md` salió de leer una corrida que se contaminó sola.

**Ninguna salió de un agente que hizo un desastre haciendo trabajo de verdad.** El bucle de
feedback de este proyecto está cerrado sobre sí mismo: `sf` se corrige contra `sf`. Eso puede
producir un binario impecable —501 tests, `vet` limpio— que sigue sin saber si sirve.

El banco se construyó exactamente para abrir ese bucle. Y los últimos tres commits del banco son
**higiene del banco**: el vocabulario que se filtra, el directorio que envenena ICM, el exit code
que mentía. Todos correctos, todos necesarios, y todos son el instrumento — no la medición. El
banco todavía no produjo **un solo resultado**.

---

## 2. La categoría se partió en tres capas

> **Actualizado el 2026-09-17:** gentle-ai pisa las tres igual que nosotros, con diez veces el
> tamaño. Ver §2.4.

Leído el 2026-09-12. Lo que importa no es la lista: es que `sf` hoy pisa las tres, y sólo una es
suya.

### Capa 1 — el runtime: dónde viven los agentes

| | |
|---|---|
| **Herdr** (`herdrdev/herdr`, AGPL) | multiplexor en Rust, un binario de ~10MB. Detecta ~20 CLIs, sabe si cada agente está **trabajando / bloqueado / ocioso / listo**. Expone un socket Unix y un CLI *para que los agentes se consulten entre sí*, lancen ayudantes y esperen tareas. |
| **Orca** (`stablyai/orca`, MIT) | ADE de escritorio + móvil. 30+ CLIs en worktrees paralelos, editor, browser con element picker, simulador iOS, automatizaciones cron, Linear/Jira/Issues. |
| Conductor · Vibe Kanban · Sculptor | la misma capa con otra forma: worktrees en macOS, un kanban Apache-2.0, contenedores Docker. |

**Esta capa está llena, tiene plata y se mueve rápido.** Y la R1 de este repo —*"`sf` nunca lanza a
nadie"*— acertó antes de que existieran. Pero `sf lanzar`, el catálogo de modelos y `sf model --via`
**se metieron adentro igual**. Herdr ya hace eso mejor, con estado de proceso real, y lo regala.

> Esas piezas no son una ventaja: son superficie que hay que mantener contra gente que la construye
> a tiempo completo.

### Capa 2 — la planificación: el documento antes del código

| | |
|---|---|
| **Traycer** | plan → ejecutar → verificar, encima de Cursor / Claude Code / Copilot. Mapea archivos afectados, ordena pasos, entrega el plan de un click, y después **verifica** los diffs contra la intención. Comercial. |
| **spec-kit** (GitHub) | la referencia agnóstica de CLI. |
| **Kiro** (AWS) | IDE spec-first, sucesor de Amazon Q. Lock-in: los specs viven en `.kiro/`, los modelos ruteados por Bedrock. |
| **OpenSpec** | MIT, vive en el repo, sin API key, sin MCP. Su rasgo propio es **delta-tracking sobre código que ya existe**. |
| **Tessl** | US$125M, spec-as-source: el código es regenerable y nunca se toca a mano. Nueve meses en beta cerrada, sólo JavaScript, y salida **no determinística a partir de specs idénticos**. |

**El frente ⑥–⑩ de `sf` —brief, PRD, constitución, backlog, roadmap— es esta capa.** Es la parte
más poblada de la categoría, la más fácil de copiar (son prompts y Markdown), y la que **corre una
sola vez en la vida de un producto**. Es donde está puesto el grueso de las 5.948 líneas de skills.

Dos datos que hay que mirar de frente:

- **Thoughtworks puso spec-driven development en `Assess`, no en `Adopt`.** La crítica textual es
  ceremonia: duplica el overhead de documentación cuando cada fase se trata como obligatoria en vez
  de como una herramienta que se agarra cuando la tarea la pide. **Nueve estados, 27 skills y tres
  paradas humanas es el punto de máxima ceremonia de toda la categoría.** El `tipo: chico/bug` es
  el principio de la respuesta, no la respuesta.
- **Tessl es la fábula.** Ciento veinticinco millones para descubrir que *generar* desde el spec no
  converge. Lo que sí converge es lo de abajo.

### Capa 3 — el árbitro: el que dice que no

**Está casi vacía, y es la de `sf`.**

Traycer "verifica", pero verifica **con un modelo**: opina sobre el diff. Kiro y spec-kit producen
documentos y confían. Nadie más corre los tests él mismo, guarda el hash de los archivos de test en
el rojo y lo compara en el verde, ni se niega a mover el estado porque falta una rama.

La lista de `README.md` §*"What it actually prevents"* no es marketing: **es el producto entero**,
y son unas dos mil líneas de Go. Todo lo demás —los nueve estados, los 27 skills, el instalador, el
plugin, el catálogo de modelos, el banco— es reparto.

| lo que hace la competencia | lo que hace `sf` |
|---|---|
| un juez LLM opina sobre el diff | una compuerta se para sobre **un hecho**: el exit code, el hash, el archivo, la rama |
| el agente dice "listo" | el estado no se mueve hasta que `sf` lo comprobó solo |
| el plan vive en el chat | el estado vive versionado en el repo |

**Una frase, y es la única que hay que defender:** *el resto de la categoría te ayuda a escribir el
spec; `sf` es el único que no te deja pasar sin cumplirlo.*

---

## 2.4 gentle-ai — el que más se nos parece, y el que fue para el otro lado

**Leído el 2026-09-17**, clonando `Gentleman-Programming/gentle-ai` (último commit `9bf454d`,
2026-09-16) y `gentle-shell` (`ce47bae`, 2026-09-17). Lo que sigue sale del código; la nota de
release se usó sólo para ubicar qué mirar.

### Los números, para saber contra qué medimos

```
                        gentle-ai        specforge
Go productivo           133.832 líneas    13.623
Go de tests             231.743 líneas    11.078
prompts instalados      177 archivos · 15.785 líneas
orquestadores           12, uno por arnés
issues abiertos         809
releases                v3.0.0, brew, archivos firmados       0
```

**Son diez veces nosotros.** No es un vecino: es la versión con recursos del mismo problema.

### Su v3.0.0 hace por defecto lo que nosotros llamamos el nodo de entrada

`internal/components/agentguidance/routing.go:42-52` escribe un protocolo obligatorio en **todo**
pedido, y va en los doce orquestadores:

```
1  Authorize              ¿esto autoriza un cambio, o es sólo leer?
2  Explore                el código que ya está, antes de proponer nada
3  Resolve uncertainty    investigar sólo una incertidumbre con nombre;
                          UNA pregunta si hay una decisión de producto real
4  Classify               ¿sustancial o chico?
5  Track before the first write
6  Implement task by task
7  Close
```

**Los pasos 2-3-4 son nuestro nodo de entrada**, y en dos cosas están mejor:

**① Su `Classify` es el piso que nosotros no escribimos.** Textual (`routing.go:48`): *"el trabajo
es sustancial cuando explorar da dos o más pasos de implementación con sentido, o progreso que
valga la pena recuperar después de una interrupción. El trabajo chico y entendido se queda chico y
no crea artefactos durables."* Es nuestro `tipo: chico` —el campo que el motor rutea y ningún
skill escribe— definido y por defecto.

**② Su `Authorize` es un paso que no tenemos.** Distingue un pedido de **leer** de un pedido de
**cambiar**, y se queda read-only hasta que esté claro. Nuestras tres entradas asumen que todo lo
que llega es trabajo: un *"¿por qué esto anda así?"* entra como si fuera un ticket.

### Pero su protocolo es prosa en el prompt, no enforcement

**Medido:** las únicas dos menciones de `odd/tasks` en código Go que no es de tests
(`routing.go:49` y `:83`) están **adentro de los strings del prompt**. No hay nada que compruebe
que el archivo existe, cuente las tareas, ni se niegue a avanzar.

### Y en la misma versión sacaron las compuertas que tenían

```
sdd-verify ya no bloquea sdd-archive
se retiró el "research admission"
se retiró el governance de intentos (budget/exhaustion)
archive procede "with truthful, possibly unfinished, task state"
```

El motivo lo escriben ellos: *"ceremony that had become gatekeeping rather than value"*.

**Fueron en la dirección contraria a la nuestra.** Tenían compuertas y las sacaron.

### La lección, y es la parte que vale

Esto no nos da la razón: es un dato **en contra**, de alguien con diez veces nuestros recursos y
muchos más usuarios. Pero mirá **qué conservaron**: *"live grant integrity, artifact locators,
task-progress routing, safe archive composition"* — o sea, **qué carpetas puede tocar un cambio
autorizado sigue siendo duro.**

```
lo que se les murió    una verificación · una atestación · un presupuesto de intentos
                       → las tres frenan sobre un JUICIO

lo que conservaron     qué carpetas se pueden tocar
                       → frena sobre un HECHO
```

**Es R3 confirmada desde afuera, y con datos de campo en vez de razonamiento:** *una compuerta
frena sobre un hecho; un juez opina.* Las que se les murieron eran jueces disfrazados de
compuerta.

> **Y el aviso que trae para nosotros:** si ellos, con ese tamaño, concluyeron que su propia
> ceremonia estorbaba, **nuestro riesgo de ceremonia es peor, no menor.** Nueve estados y 27
> skills sigue siendo el punto de máxima ceremonia de la categoría.

### Lo que hay que hacer con esto

| # | qué | tamaño |
|---|---|---|
| 1 | robar **`Authorize`** como paso 0 de las tres entradas — hoy no existe | chico |
| 2 | robar la redacción de su **`Classify`** para el nodo de entrada | chico |
| 3 | releer **las nueve compuertas** con una sola pregunta: *¿frena sobre un hecho o sobre un juicio?* La que frene sobre un juicio se va a morir igual que se les murió a ellos | mediano |
| 4 | robar **cómo miden**: su evidencia de campo es un experimento **con control** — el protocolo corre en un pedido que nunca menciona workflow, y un control de sólo-documentación no crea nada. Nosotros no corremos controles | mediano |

### Lo que NO cambia

La vara sigue sin estar en ninguno de los cuatro repos leídos. En gentle-ai el paso 3 investiga,
pero **no escribe con qué va a elegir antes de ver las opciones**. Se buscó `rubric`, `scorecard` y
`criteria before` en superpowers y en `mattpocock/skills` el 2026-09-13, y en gentle-ai el 09-17:
nada.

**Lo que cambia es el mapa competitivo, no el diseño de `specs/nodos/`.**

---

## 3. Lo que sale de esto

Cinco cosas, en orden, y las dos primeras no son código.

### ① Pushear. Esta semana.

Es la decisión de Javier —reemplaza el producto público— y por eso sigue abierta desde el 26 de
agosto. Pero el argumento de `salir-a-la-cancha.md` sólo se fortaleció: la divergencia pasó de 0 a
14, `specd` publicó su v1.1.0 con instalador y sitio el 1 de agosto, y **nada de la lista de aquel
documento se puede verificar hasta que esto pase**. No hay que decidir si `sf` está listo: hay que
decidir si sigue siendo privado.

### ② Una corrida real, en un proyecto ajeno, antes del próximo commit de diseño.

No el banco con dos brazos. **Una.** El banco resuelve un problema de n grande —comparar con y sin
`sf`— cuando todavía no existe el n = 1. La primera corrida honesta va a producir más compuertas
nuevas que los tres documentos que faltan escribir, y es lo único que puede.

> Regla propuesta, y es incómoda a propósito: **no entra un documento nuevo a `specs/` hasta que
> haya una corrida registrada.** El repo ya demostró que puede generar diseño indefinidamente.

### ③ Sacar el árbitro solo, sin los nueve estados.

Las compuertas que son aritmética pura —el rojo antes del verde con hash de los archivos de test,
el verde vacío, un lote un commit, la cobertura criterio por criterio, el número de revisión— **no
necesitan brief, ni PRD, ni roadmap, ni ninguno de los 27 skills**. Funcionan sobre cualquier
workflow: con spec-kit, con Traycer, o sin nada.

Eso es un binario chico más un hook de Claude Code, se instala en un minuto, y es lo único de acá
que nadie más tiene. **Es el producto que se puede mostrar sin explicar una metodología.**

### ④ Conceder el frente.

⑥–⑩ es la capa más poblada, corre una vez por producto, y es donde está el grueso del mantenimiento
de skills. Las opciones, de menor a mayor costo: dejarlo opcional; o consumir los artefactos de
OpenSpec / spec-kit en vez de producir los propios. Lo que no conviene es seguir afilándolo.

Y el hueco real está del otro lado: `specd` es brownfield-first, OpenSpec vive de delta-tracking
sobre código que ya existe, y **`sf` arranca asumiendo un producto nuevo**. Casi nadie empieza un
producto; todo el mundo cambia uno que ya está andando.

### ⑤ Devolver la capa 1.

`sf lanzar`, el catálogo de modelos y `--via` compiten contra Herdr y Orca. La R1 ya decía que `sf`
no lanza; conviene que el código también lo diga. `sf` tiene que ser lo que corre **adentro** de un
panel de Herdr o de un worktree de Orca, y aburrirse con eso: un binario y un `AGENTS.md`.

---

## 4. Lo que este documento NO dice

**No dice que esté mal construido.** El binario está mejor testeado que el 90% de la categoría y el
diagnóstico —*"un agente razona bien y recuerda mal un proceso que aceptó hace veinte mil tokens"*—
es correcto, y lo confirma que `specd` llegó solo al mismo lugar.

**Dice que está mal balanceado.** El 20% que nadie más tiene está enterrado abajo del 80% que todos
tienen, y ninguno de los dos salió nunca del repo.

Y una honestidad final, porque corresponde: **este documento es análisis número treinta y uno.** No
rompe el bucle. Lo rompe la corrida.
