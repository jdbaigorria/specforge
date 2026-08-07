# Brief — el módulo Inception

**Fase:** Spark · **Fecha:** 2026-08-07 · **Decisión:** proceder
**Doble propósito:** es un brief real *y* es la primera prueba del formato de brief. Si
falta un campo o sobra otro, se ve acá y no en un debate.

---

## 1. La chispa

SpecForge arranca en `sf propose`. O sea: **asume que ya sabés qué construir.**

Falta todo lo de antes — de "se me ocurrió algo" a "esto es lo que hay que construir y por
qué". Eso es Inception, y son dos fases:

- **Spark** — *la chispa*. Decide **si lo hago y por qué**.
- **Flame** — *preparar el fuego*. Decide **qué es exactamente**.

Después recién entra Forge, que es construir.

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

**Proceder.** Con dos condiciones que salen de las quejas medidas:

1. **Inception tiene dos modos.** Proyecto nuevo (PRD del MVP + constitución + historias) y
   feature nueva sobre algo existente (sólo la rebanada; la constitución ya está). El
   *"para features chicas era overkill"* pasaba porque había un solo modo y era el pesado.
2. **Cada fase pregunta sólo lo contestable en esa fase.** Lo que hizo abandonar OpenSpec no
   fue el número de pasos: fue que preguntaba **cómo implementar** algo durante la fase de
   spec. Eso es una pregunta de Forge hecha en Inception. Lo que no toca, se difiere — no se
   inventa ni bloquea.

---

## 6. Preguntas abiertas

Cada una con **la fase que la contesta**. Inception no las resuelve: las nombra.

| # | Pregunta | La contesta |
|---|---|---|
| Q1 | ¿El pinponeo se orquesta con herramienta, o queda afuera y sólo entra su resultado? | **Spark**, al caminarlo a mano |
| Q2 | ¿Cuánto de la constitución se escribe en Flame? Hipótesis: sólo lenguaje y comando de test; el resto crece cuando haga falta | **Forge**, cuando la pida |
| Q3 | ¿Las historias de usuario son la unidad que Forge convierte en requisitos, o hay un paso intermedio? | **Forge** |
| Q4 | El PRD es el mapa y las historias el entregable. Si el PRD cambia, la IA resincroniza — pero, ¿qué pasa con las historias **ya construidas**? | **Forge** |
| Q5 | ¿El "por qué" del brief y el "por qué" del proyecto (constitución) son dos niveles? ¿Cuál manda si chocan? | **Flame** |
| Q6 | ¿Qué se registra del brief en el ledger: el documento entero, o sólo la decisión y su hash? | **capa cross**, al final |

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
