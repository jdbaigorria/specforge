# Brief — el módulo Inception

**Fase:** Spark · **Fecha:** 2026-08-07 · **Decisión:** proceder
**Doble propósito:** es un brief real *y* es la primera prueba del formato de brief. Si
falta un campo o sobra otro, se ve acá y no en un debate.

---

## 1. La chispa

SpecForge arranca en `sf propose`. O sea: **asume que ya sabés qué construir.**

Falta todo lo de antes — de "se me ocurrió algo" a "esto es lo que hay que construir y por
qué". Eso es Inception, y son dos fases:

- **Spark** — *la chispa*. Decide **si lo hago y por qué**. → el **brief**
- **Flame** — *preparar el fuego*. Decide **qué es exactamente**. → **PRD, historias, constitución**

Después recién entra Forge, que es construir.

> **Las dos corren siempre, y en ese orden.** No son alternativas ni dependen del tipo de
> proyecto. Lo que cambia según el caso es **cuánto pesa cada una** (§5.1), nunca cuál se
> usa. Un Spark de cinco líneas sigue siendo un Spark; un Spark salteado es un agujero.

---

## 2. Por qué lo quiero — y por qué ahora

Cuatro razones, todas ancladas en algo que pasó. Ninguna es "sería lindo tener".

### 2.1 SpecForge nunca construyó la parte que lo originó

SpecForge nació porque **OpenSpec era demasiado quilombo**. Pero lo que se construyó fue de
`propose` para adelante — o sea, la parte que OpenSpec *sí* tenía resuelta. **La fase que
molestaba nunca se rehízo, se salteó.**

### 2.2 Sin una puerta que pregunte "¿por qué?", entra cualquier cosa — y está medido

La auditoría de features midió que **el 43% de la superficie de SpecForge entró por leer
otros repos**, sin ningún usuario que lo pidiera. Sólo ~10% vino de una falla observada.

No fue por descuido. Fue porque **no había ninguna puerta donde hubiera que escribir por qué**.
Un brief con el porqué obligatorio es esa puerta, y es la barandilla más barata que hay
contra la acumulación.

### 2.3 El flujo ya existe, pero no deja rastro

Hoy Spark se hace igual, a mano: Claude para pinponear, Perplexity para buscar prior art,
Claude de nuevo para comparar. Funciona. **Lo que se pierde es todo lo demás:** por qué se
decidió, qué se descartó, qué se objetó y se hizo igual.

Seis meses después no queda nada de eso, y es justo lo que hace falta para saber si una
decisión fue buena.

### 2.4 SpecForge no se usa a sí mismo

**Cero features de SpecForge construidas con SpecForge.** Este brief es el primer intento
de revertirlo: es un artefacto real del flujo nuevo, sobre una feature real.

---

## 3. Qué ya existe, y en qué me aparto

> **Los 16 repos de `inspiration/` se leyeron para diferenciarse, no para surtirse.** La
> diferencia importa: leerlos para ver *qué features agregar* es exactamente lo que produjo
> el 43% de §2.2.

### 3.1 La división de tres módulos NO es diferenciación — y hay que decirlo

**Inception → Construction → Operations es nomenclatura de AI-DLC, publicada por AWS**, y
specs.md ya la implementó. Llegar a esa división no es originalidad: es converger a donde
el mercado ya convergió. Baja el riesgo, y **no vende nada**.

**Incluso los nombres son prestados.** *Spark → Flame → Forge* es la nomenclatura de
**specs.md Ideation**, ya anotada en `planning/SPECFORGE-V2-INSPIRACION.es.md` §4.2. Se
adoptan igual, porque son buenos y porque coincidir con el vocabulario del rubro es una
ventaja, no una deuda. Pero **no son un aporte nuestro** y este brief no los va a presentar
como tal.

### 3.2 Quién cubre esta fase hoy

| Repo | Qué hace en Inception | Qué le falta |
|---|---|---|
| **DDA / inception_loop** | **Es literalmente este módulo.** Pipeline `01-inputs/ → 02-brief/ → 04-prd/ → 05-hu/ → 10-export/`. Ingesta desde repo, docs, Jira y app. Para PM/analista, sin escribir código | **cero enforcement.** Sus gates son prosa: "revisá antes de avanzar" |
| **specs.md Ideation** | Spark → Flame → Forge, con técnicas nombradas (Six Hats, Disney, rueda anti-sesgo) | **0 checkpoints, no bloquea nada.** Es un facilitador de conversación |
| **nWave DISCOVER/DIVERGE** | explora mercado y espacio del problema antes de converger | **sólo greenfield.** No sirve para agregar una feature a algo que ya existe |
| **BMAD `plan/`** | PRFAQ, product-brief, project-context | sin enforcement, sin ledger |
| **OpenSpec** | — | **no tiene esta fase.** Arranca en el change |

### 3.2b Cómo entra cada uno a un proyecto que ya existe

Relevado el 2026-08-07 leyendo los repos, con la necesidad ya escrita (§5.2).

| Enfoque | Quién | Estado |
|---|---|---|
| **Ingesta completa por adelantado** | DDA (4 fuentes: repo, docs, Jira, app) · BMAD `document-project` | **BMAD lo DEPRECÓ.** Su skill hoy sólo reenvía, con el motivo escrito: *"en vez de generar volumen de documentación, cura un sistema de contexto chico y verificado — un kernel siempre cargado más un bundle de conocimiento"* |
| **Auto-detección de patrones** | **specs.md**, flow FIRE | vigente. *"First-class brownfield — auto-detects existing patterns and conventions"* |
| **Nada: delta-first** | **OpenSpec** | vigente, y es su **titular** |
| **Sólo greenfield** | nWave (DISCOVER/DIVERGE) | — |
| **No lo tratan** | cavekit, Kiro, Archon, superpowers, kaddo, gentle-ai, SpecEngine | 0 menciones en el README |

**Lo que dice OpenSpec, textual** (`docs/existing-projects.md`, primera línea):

> *"You do not document your whole codebase to start. You write specs only for what you're
> about to change."* · *"'Mi app tiene 80.000 líneas, ¿tengo que escribir specs de todo antes
> de que OpenSpec sirva?' **No. Lo odiarías, y nosotros también.**"*

**El mecanismo:** su unidad es un **delta** (`ADDED`/`MODIFIED`/`REMOVED`) contra el estado
actual. Cada cambio archivado funde su delta en la base. En su propio repo eso dio
**83 cambios archivados → 36 specs**: la base no la escribió nadie de entrada, **se armó sola**.

**Lo que se paga:** la cobertura queda parcial para siempre y no hay forma de saber qué
falta. Es correcto para ellos y hay que decirlo, no disimularlo.

### 3.2c Consecuencia incómoda: la ingesta perezosa NO es diferenciación

El modelo perezoso de §5.2 es **el titular de OpenSpec**, no un hueco que dejaron. Y la
posición intermedia que adopta este brief —kernel mínimo declarado + crecimiento— es
**exactamente donde aterrizó BMAD después de retirarse** del modelo eager.

O sea: en esto no somos originales, somos los terceros en llegar. Se adopta igual, porque
es lo correcto y ahora hay evidencia de campo —**alguien construyó el barrido completo,
lo shipeó y lo deprecó**— pero **no se vende como aporte**.

> **Nota de honestidad.** El recuerdo de *"OpenSpec vendía que analizaba tu código"* no
> resistió la verificación: ni el README ni los docs lo dicen, y la guía dedicada dice lo
> contrario. Lo que sí existe con esa promesa es **specs.md** (*auto-detects existing
> patterns and conventions*). Vale anotarlo porque el error iba a entrar al brief como
> hecho.

### 3.3 El hueco, en una línea

> **Nadie enforcea Inception, y nadie une los módulos con un solo ledger.**

nWave enforcea la construcción. gentle-ai enforcea la entrega. **Nadie enforcea la fase
donde se decide qué construir** — que es justo donde entró el 43% de §2.2.

Un `sf gate` encadenado por hash que vaya desde *"aprobé el brief"* hasta *"autoricé el
release"* no existe en ninguno de los 16 repos.

**Ese es el aporte. Todo lo demás es ejecución.**

### 3.4 Por qué no uso DDA en vez de construir esto

Es la pregunta honesta, porque DDA hace el 80% de lo que quiero y ya está escrito.

| | |
|---|---|
| **Lo que me sirve** | el pipeline de artefactos, la ingesta de cuatro fuentes, `ENGRAM.md` como memoria del proyecto, la matriz de trazabilidad |
| **Por qué no alcanza** | sus gates son prosa. Nada impide avanzar con el brief a medias, y **nada ata el brief a lo que después se construye**. Sin esa costura, Inception es una carpeta de documentos que envejece |
| **Por qué no forkeo** | está armado alrededor de una práctica de consultoría —prototipo React, 25 skills, cliente externo— y el usuario objetivo acá es **el dev indie que también es PM y tester**. Forkear sería sacarle más de lo que le queda |

**Se roban las mecánicas, no el repo.**

---

## 4. Qué reuso — no reinventar la rueda

Lo que ya está escrito y sirve tal cual o casi:

| Pieza | Dónde está | Para qué |
|---|---|---|
| **`sfp-scout`** | `skills/sfp-scout/` | **es Spark, ya escrito.** Research con MCPs, veredicto `proceed`/`pivot`/`kill`, IDs `PR#` trazables. Mejor punto de partida del módulo |
| **`sfx-grill-me`** | `skills/sfx-grill-me/` | interrogar los supuestos: a quién, por qué ahora, qué lo mata |
| **`sf sources`** | CLI (`RM-C3`) | la auditoría lo marcó **sin consumidor**. No fracasó por diseño: le faltaba el módulo que produce el material. Inception es ese módulo |
| **Preguntas abiertas** | `DL-13` | ya son objeto de primera clase. Acá se les agrega **la fase que las contesta** |
| **Ingesta multi-fuente** | DDA (`dda-learn-from-*`) | mecánica a robar: repo, docs, Jira, app |
| **Anulación registrada** | patrón del ledger | el humano anula, pero queda escrito con autor y motivo |

---

## 5. La decisión

**Proceder.** Con dos condiciones que salen de las quejas medidas.

### 5.1 Condición 1 — dos ejes cruzados, no uno

**Fase** y **modo** son ejes distintos, y confundirlos es el error más fácil de cometer acá.

- **Fase:** Spark → Flame. **Siempre las dos, siempre en ese orden.**
- **Modo:** qué tan pesada es cada fase, según de dónde venís.

|  | **Spark** — ¿lo hago y por qué? | **Flame** — ¿qué es exactamente? |
|---|---|---|
| **Proyecto nuevo** | ¿vale la pena? ¿ya existe en el mercado? ¿por qué yo y no forkear? | PRD del MVP + **constitución** + historias |
| **Feature nueva** | ¿vale la pena? **¿no lo resuelve algo que ya tengo?** | sólo las historias de la rebanada. La constitución **ya está** |

> **Spark importa MÁS en el modo feature, no menos.** El 43% de superficie sin usuario que
> la pidiera (§2.2) **no entró como proyectos: entró de a una feature por vez**, y cada una
> parecía razonable en el momento. Si Spark corriera sólo para proyectos nuevos, SpecForge
> no tendría ninguna defensa contra justo la enfermedad que se le midió.
>
> Para una feature chica Spark son cinco líneas. Cinco líneas escritas, no cero.

**El "para features chicas era overkill" se explica acá:** el flujo viejo tenía un solo
modo, el pesado, y lo hacía correr entero para cualquier cosa.

> **El eje real no es greenfield/brownfield: es "¿es la primera vez o no?".** Una vez que el
> proyecto existe, de dónde vino deja de importar — un greenfield que ya arrancó y un
> brownfield adoptado se comportan igual. **Modo proyecto pasa una sola vez; modo feature,
> todas las demás.**
>
> Prior art: **specs.md FIRE** ya gradúa por complejidad — *"Adaptive checkpoints: Autopilot
> (0), Confirm (1) o Validate (2)"*. La idea de que el peso se ajuste no es nuestra.

### 5.1b Lo que Spark tiene que preguntar en modo feature, y nadie pregunta

Además de *"¿vale la pena?"* y *"¿ya lo tengo?"*, una tercera:

> **¿Esto entra en el PRD que ya está, o lo cambia?**

- **Entra** → Flame sólo genera las historias. El PRD no se toca.
- **Lo cambia** → hay que tocar el PRD **primero**, y eso es una decisión mucho más grande
  que agregar una feature.

**Este es el detector del 43%** (§2.2). Esas features entraron de a una, cada una parecía
razonable, y **ninguna se preguntó nunca si estaba corriendo el límite del producto**. Una
feature que obliga a reescribir el PRD no es una feature: es un cambio de alcance
disfrazado.

### 5.2 Adoptar SpecForge sobre código que ya existe

Es un tercer caso, distinto de los dos del cuadro, y no estaba cubierto: **el proyecto ya
existe y nunca usó SpecForge** (es lo que ejercita `examples/brownfield-tempconv`).

**No es una fase nueva.** Es una entrada, y ocurre **una sola vez**:

| Momento | Qué pasa |
|---|---|
| **Al adoptar** (`sf init`) | se **deriva** la constitución mínima leyendo el repo: lenguaje, comando de test, poco más. Es mecánico —`go.mod`, `package.json`, el script de test— y no requiere decidir nada |
| **Después** | Spark y Flame normales, en modo *feature nueva*, una feature por vez |

**Decidido: la ingesta es perezosa. Eager sólo lo declarado.** No hace falta documentar el
proyecto entero para empezar; hace falta entender lo que **la próxima feature toca**.

| Qué | Cuándo | Costo | De dónde sale |
|---|---|---|---|
| Lenguaje, comando de test, build | al adoptar, **una vez** | **1 archivo** | está **declarado** en el CI o el manifest |
| Convenciones y patrones | **cuando vas a escribir**, por feature | los archivos vecinos | se leen donde vas a tocar |
| Arquitectura y límites de capas | **nunca automático** | — | lo dice el humano, si le importa |

**El mejor lugar para el kernel declarado es el CI, no el código.** Medido en este repo:
`.github/workflows/lint.yml`, **577 bytes**, dio el lenguaje (Go), el directorio de trabajo
(`cli/`), el comando de test (`go test ./...`) y el de build. No se infirió nada: está
declarado, y **está verificado continuamente** — si estuviera mal, el CI fallaría. Un
resumen de 18k líneas hecho por un LLM no tiene nada que lo verifique.

*(Detalle que lo confirma: el `go.mod` no está en la raíz sino en `cli/`. Buscando sólo
manifests se perdía; el CI dio el `working-directory` gratis.)*

**Por qué no eager para lo demás**, con dos evidencias y no con una opinión:

1. **Ya lo intentamos y falló.** `sf arch` está congelado porque `RM-C7b` probó que el mapeo
   componente→archivo **no funciona** en `examples/brownfield-tempconv`: *"dos componentes en
   un archivo es normal en código chico"*.
2. **Alguien más lo construyó y lo deprecó.** BMAD `document-project` → hoy sólo reenvía a un
   *"kernel chico verificado"* (§3.2b).

Y hay un tercer motivo, que es el de siempre: un resumen de todo el repo escrito por un LLM
**suena bien, está a medias, y nadie lo verifica**. Es el riesgo de fabricación aplicado al
onboarding, en el peor momento — cuando todavía no sabés si la herramienta te sirve.

**Lo que se pierde, dicho de frente:** podés escribir una feature que rompa una convención
usada en otra parte del repo y no enterarte hasta la revisión. El revisor puede marcarlo y
cada convención descubierta queda escrita, así que el agujero se achica con el uso — pero
**no se cierra**.

Y ese entendimiento por feature ya tiene dónde vivir: es la pregunta de Spark en modo
feature — *"¿no lo resuelve algo que ya tengo?"*. En greenfield se contesta buscando en el
mercado; en brownfield, **también dentro del repo**. Misma pregunta, dos espacios.

### 5.3 Condición 2 — cada fase pregunta sólo lo contestable en esa fase

Lo que hizo abandonar OpenSpec no fue el número de pasos: fue que preguntaba **cómo
implementar** algo durante la fase de spec. Eso es una pregunta de Forge hecha en Inception.
Lo que no toca, **se difiere** — no se inventa ni bloquea.

---

## 6. Preguntas abiertas

Cada una con **la fase que la contesta**. Inception no las resuelve: las nombra.

### 6.1 Cerradas — con medición, no con debate

| # | Pregunta | Respuesta |
|---|---|---|
| **Q2** | ¿Cuánto de la constitución se escribe en Flame? | **Sólo lo declarado**: lenguaje, comando de test, build. Sale del CI (§5.2). El resto crece cuando Forge lo pida |
| **Q5** | ¿El "por qué" del brief y el del proyecto son dos niveles? ¿Cuál manda? | **Sí, dos niveles.** El del proyecto vive en el PRD/constitución; el de la feature, en el brief. **El choque se detecta en Spark** (§5.1b) y manda el del proyecto — cambiarlo es una decisión aparte y más grande |
| **Q7** | ¿La ingesta es perezosa o hace falta barrido? ¿Cuánto contexto necesita Spark? | **Perezosa.** Y el costo se midió: **2 búsquedas y 70 líneas** para contestar *"¿ya existe?"* sobre un repo de 18k líneas (§6.3) |

### 6.2 Abiertas

| # | Pregunta | La contesta |
|---|---|---|
| Q1 | ¿El pinponeo se orquesta con herramienta, o queda afuera y sólo entra su resultado? | **Spark**, al caminarlo a mano |
| Q3 | ¿Las historias de usuario son la unidad que Forge convierte en requisitos, o hay un paso intermedio? | **Forge** |
| Q4 | Si el PRD cambia, la IA resincroniza las historias — ¿qué pasa con las **ya construidas**? | **Forge** |
| Q6 | ¿Qué se registra del brief en el ledger: el documento entero, o sólo la decisión y su hash? | **capa cross**, al final |
| Q8 | La cobertura perezosa deja el proyecto parcialmente spec'ado **para siempre**. `F7 — cobertura de criterios` entonces habla **de la rebanada construida con la herramienta, no del proyecto**. ¿Cómo se dice eso sin que se lea como "el proyecto está cubierto"? | **capa cross**, al final |

> **Q8 es la misma disciplina que `no_refuta`:** decir lo que la métrica significa de verdad
> y no lo que suena mejor. Queda anotada ahora para que no se descubra tarde.

### 6.3 Cómo se cerró Q7 — la medición

Caminado el 2026-08-07 sobre el caso brownfield más real que hay: **SpecForge sobre
SpecForge** (18k líneas, 38 comandos, 15 skills, cero estado propio).

**Feature caminada:** *"`sf init` que derive la constitución mínima leyendo el repo"*.
**Pregunta de Spark:** ¿no lo resuelve algo que ya tengo?

| Paso | Qué se hizo | Resultado |
|---|---|---|
| 1 | grep de los **títulos** de `skills/sf-init/SKILL.md` | aparece *"Step 2: Detect project type"* y una sección *"### Brownfield"* |
| 2 | grep de detección de stack en `cli/` (18k líneas) | **nada.** La detección no está en Go |
| 3 | leer 70 líneas de esa skill | ya hace todo: detecta, genera `project.md` y `conventions.md`, y hasta tiene escrita la misma idea (*"the codebase tells you the how but not the why"*) |

**Veredicto: `pivot`.** La feature ya existe. Lo que queda en pie no es construirla sino
**cambiarla de ansiosa a perezosa** — una intervención mucho más chica.

**La regla que sale, y es lo generalizable:**

> **Spark no necesita leer el repo. Necesita leer el índice de lo que el repo ya sabe hacer.**

Nunca se abrió un `.go`. La respuesta vivía en la **superficie** —los títulos de 15 skills—
no en el interior. Para otro proyecto el índice será otro (README, lista de comandos, API
pública, nombres de carpetas), pero el principio aguanta. Y un proyecto que no puede
contestar barato *"¿qué sé hacer ya?"* tiene un problema de documentación, no de Spark.

Corolario que refuerza §5.2: **"¿ya lo tengo?" se contesta en la superficie.** Un onboarding
que lee todo el código por adelantado está leyendo el interior para contestar preguntas que
viven afuera.

> **Q4 tiene un riesgo ya identificado.** Regenerar historias pisa la spec de cosas que ya
> existen, en silencio. La regla propuesta: **el resync muestra un diff y marca cuáles ya
> están construidas; nunca sobreescribe.** Es la regla de `sf-amend` — *nunca asumas que el
> que está mal es el spec*.

---

## 7. Objeciones anuladas

Lo que se objetó, qué se decidió, y por qué. Esto existe para poder mirarlo dentro de seis
meses y ver quién tenía razón.

### O1 — "El orden debe ser capa cross → Forge → Inception"

- **Objeción:** era una decisión ya cerrada, con motivo escrito: *"Forge segundo porque es
  lo único construido: su contrato se escribe contra código real y se testea el mismo día."*
- **Se anuló.** Inception primero, capa cross al final.
- **Por qué:** la capa cross verifica que el trabajo cumple los requisitos, pero **la forma
  del requisito la define Inception**. Se estaba escribiendo el verificador de un artefacto
  cuya forma nadie decidió. Síntoma medible: **F6 se reescribió tres veces en un día** sin
  poder probarse contra ningún flujo real.
- **Qué probaría que la anulación fue un error:** que Inception se diseñe sin ninguna
  restricción de verificabilidad y produzca artefactos que después no se puedan chequear.

### O2 — "Arrancar desde cero mirando repos similares es la vía que produjo el 43%"

- **Objeción mía, en esta misma sesión.**
- **Se refutó, y con razón.** Lo que produjo el 43% no fue leer repos: fue **incorporar sin
  validar ni comparar**, sólo preguntando qué más agregar.
- **Por qué importa:** el arreglo no es dejar de leer repos, es **leerlos con una necesidad
  ya escrita**. Misma entrada, dirección opuesta. Es la disciplina de §3.
- **Queda anotado** porque muestra que el mecanismo funciona en las dos direcciones: acá el
  que se equivocó fue el que objetaba.

### O3 — "Fusionar el auditor y el juez pierde una garantía"

- **Objeción mía** contra fusionarlos en un actor.
- **Se anuló.** Son un solo actor.
- **Por qué:** la objeción le pegaba a una versión donde el auditor era la sesión principal.
  En la versión real es un subagente limpio, así que la garantía de contexto fresco
  sobrevive. La objeción era correcta contra algo que nadie estaba proponiendo.

---

## 8. Qué sigue

1. **Caminar Flame a mano** para este mismo brief: sacar el PRD, las historias y la
   constitución mínima de Inception.
2. Recién ahí, mirar los repos de nuevo — con la necesidad ya escrita — para ver cómo
   resolvieron lo que nos falte.
3. La capa cross **no se diseña: se deriva**. Es la respuesta a *"¿cómo valido estos
   artefactos?"* y no puede existir antes que los artefactos.

---

## Nota sobre el formato

Este brief se escribió sin plantilla, siguiendo lo que salió del pinponeo. Lo que se puede
observar de él como prueba del formato:

- **Lo que más costó llenar honestamente fue §3.1** — admitir que la división de módulos y
  hasta los nombres son prestados. Un formato que no obligue a esa sección deja pasar
  cualquier cosa como original.
- **§4 (qué reuso) se llenó solo** y devolvió cuatro piezas ya escritas que estaban sin
  consumidor. Sospechamos que ese es el campo con mejor relación valor/esfuerzo.
- **§7 (objeciones) tuvo contenido real desde el primer día**, incluida una objeción que
  resultó equivocada. Si sólo registrara las anuladas *por el humano*, se perdería ese caso.
  El campo no es *"objeciones anuladas"* sino **"objeciones y qué pasó con ellas"**.
- **La primera versión mezcló dos ejes y se leyó mal.** Decía "Inception tiene dos modos"
  al lado de "Inception tiene dos fases", y el primer lector entendió que **Spark era para
  greenfield y Flame para brownfield**. Eran cuatro combinaciones presentadas como dos.
  **Un brief que introduce más de un eje DEBE cruzarlos en un cuadro**, no describirlos en
  prosa uno detrás del otro. Corregido en §5.1.
- **Escribir §5.2 destapó un caso entero que no estaba** — adoptar SpecForge sobre código
  que ya existe. No apareció al diseñar: apareció cuando alguien leyó y preguntó *"¿y en
  brownfield?"*. Es evidencia a favor de que el brief se lea antes de aprobarse, y de que
  la lectura la haga alguien que no lo escribió.
- **Ir a los repos DESPUÉS de escribir la necesidad funciona, y se nota.** §3.2b se relevó
  con `Q7` ya planteada, y en una pasada devolvió tres cosas que no se habrían encontrado
  buscando "qué features agregar": que **BMAD construyó el barrido eager y lo deprecó**, que
  el modelo perezoso **es el titular de OpenSpec y no un hueco**, y que nuestra posición
  intermedia **ya la ocupa BMAD post-retirada**. Las tres achican el proyecto en vez de
  agrandarlo. Es la diferencia entre leer para diferenciarse y leer para surtirse (§3).
- **El brief atajó un error factual antes de que se volviera premisa.** Entró la creencia de
  que *"OpenSpec vendía que analizaba tu código"*; verificarla contra el repo la refutó y
  además reasignó la promesa a specs.md, que sí la hace. **Un campo de prior art que no exija
  la cita deja pasar recuerdos como hechos** — y un brief es justo el lugar donde un recuerdo
  falso se fosiliza en decisión.
