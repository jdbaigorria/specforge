# Los artefactos del flujo — prueba de escritorio

**Fecha:** 2026-08-10 / 11 · **corregido el 13** · **Método:** paso a paso por `flujo-real.md`, preguntando en cada
uno: **qué artefacto sale · quién lo consume · qué formato · qué estructura**.

> **Por qué así.** Este tema hundió la versión anterior de SpecForge. Se estructuró de más y
> se decidió el formato antes de saber quién iba a leer cada cosa. Acá el orden es al revés:
> **primero el consumidor, después el formato.**

**Estado:** ✅ **cerrada.** Las siete rondas cubren el flujo entero, ①–㉓. El inventario
completo de artefactos está en §11, y **las correcciones posteriores están en §13**.

---

## 1. Las reglas transversales

Salieron de las rondas, no se decidieron de antemano.

### 1.1 Archivo o estado

> **Es ARCHIVO si otro lo tiene que leer. Es ESTADO si sólo hay que saber que pasó.**

*"El modelo recomendado es DeepSeek"* no necesita un `.md`: es un dato de una línea.

### 1.2 Prosa o datos

> **¿`sf` tiene que escribirlo o consultarlo? Sí → JSON. No → `.md`.**

El motivo es duro: **`sf` no tiene un LLM adentro**. No puede reescribir un markdown. Si el
roadmap es prosa y hay que marcar algo como hecho, hace falta un modelo para reescribir el
archivo — y ahí se corrompe.

### 1.3 Frontmatter: la cabecera es de `sf`, el cuerpo es del modelo

Casi todo artefacto tiene **las dos cosas**, y por eso no hacen falta archivos paralelos:

```
---
lo que sf necesita leer     ← estructurado, cinco líneas
---
lo que el modelo lee        ← prosa
```

Javier: *"no sólo me cierra, creo que es el camino correcto, porque después se puede ir
hacia un LLM wiki como el que proponía Karpathy"*.

### 1.4 Referenciar por id, nunca copiar

Consecuencia directa del wiki. Ningún artefacto repite el contenido de otro: lo apunta por
`id`. Cierra la pregunta que `flujo-real.md` había dejado abierta (*"¿el roadmap referencia
los `us-#` o los copia?"*) → **referencia**.

### 1.5 Lo que se puede generar, se genera

Un índice mantenido a mano siempre queda viejo. **`sf` genera las vistas**; los archivos
guardan los hechos.

### 1.6 Regla defensiva

> **Sólo va a JSON lo que `sf` toca de verdad.**

Cada JSON es un schema que hay que mantener y migrar. Los `.md` no tienen ese costo.

---

## 2. Los tres tipos de parada

La máquina no tiene dos modos, tiene tres:

| | Qué hace | Dónde |
|---|---|---|
| **PARA** | no avanza sin Javier | ⑥ · ⑧ · ⑰ |
| **PARADA BARATA** | muestra qué salió, seguís con un enter | ⑨ |
| **SIGUE** | avanza solo | todo lo demás |

La parada barata salió de *"a veces reviso las us y a veces no"*. Se resuelve con un enter y
queda **configurable** en la constitución global (`para_en: [6, 8, 9, 17]`).

**Hipótesis sin confirmar:** el *"a veces"* no es al azar — en la entrada A el ⑨ escupe 7
historias de un saque y no estuviste en ninguna; en la entrada B la historia sale de un
pinponeo en el que sí estuviste.

---

## 3. La jerarquía: historia vs feature

**Decidido, y cambió el diseño.** El grupo del roadmap **es** una feature.

```
us-#      →  el QUÉ    · el requisito y sus criterios de aceptación
feature   →  el CÓMO y el CUÁNDO · la solución, el trabajo, la entrega
```

**Todo el ciclo ⑪–㉓ corre por feature, no por historia.** Argumento de Javier:

> *"si una feature involucra 3 us, no vas a hacer dos y dejar una colgada — la idea siempre
> es completar una feature"*

Y confirmado por `flujo-real.md`, que en el ㉓ dice *"**las historias** de usuario también se
archivan"*, en plural.

### La regla que le faltaba al ⑩

De la duda *"¿una sola solución técnica cubre 3 historias?"* salió el criterio de agrupado:

> **Una feature es un conjunto de historias que comparten solución técnica y se entregan
> juntas.**

Si tres historias necesitan tres soluciones distintas, **no son una feature**: el
agrupamiento está mal. Antes el ⑩ agrupaba *"por prioridad"*, que no permite decir que un
grupo esté mal armado. Ahora sí.

### El riesgo, y su antídoto barato

Features grandes: si junta 8 historias, la vuelta dura demasiado. → en el ⑩, `sf` avisa
*"esta feature tiene 8 historias, ¿la partís?"*. **Aviso, no freno.**

---

## 4. Ronda 1 — el brief (①–⑥)

Se **arma** en el bucle ②–⑤ y se **sella** en ⑥.

| Consumidor | Para qué |
|---|---|
| el modelo del ⑦ | generar el PRD encima |
| Javier, meses después | *"¿por qué hice esto?"* / *"¿por qué NO?"* |
| **`sf`** | **saber si está sellado**, para dejar pasar al ⑦ |

Ese tercer consumidor es el que obliga al frontmatter.

```markdown
---
tipo: brief
id: brief
estado: sellado          # borrador | sellado
veredicto: hacelo        # hacelo | pivotea | no-lo-hagas
fecha_sello: 2026-08-10
---

# <producto>

## Qué quiero crear
## Por qué                  ← acá entra "quería aprender X"
## Qué ya existe            ← la tabla del ③ y ④
| herramienta | qué hace | por qué no me sirve |
## En qué se diferencia lo mío
## El veredicto
> **hacelo** — porque …
```

**La tabla "qué ya existe" va.** Los pasos ③ y ④ son medio bucle —Perplexity, revisar el repo
ajeno, *"¿conviene forkearlo?"*— y hoy se pierde entero: queda sólo la conclusión.

**Si el veredicto es "pivoteá", se reescribe el mismo archivo.** Git es el versionador. Nada
de `brief-v2.md`. Y **se commitea en cada ciclo de iteración**.

---

## 5. Ronda 2 — el PRD (⑦)

| Consumidor | Para qué |
|---|---|
| el modelo del ⑧ | armar la constitución |
| **el modelo del ⑨** | **partirlo en historias** ← el consumidor fuerte |
| `sf` | que exista, y **que se sepa cuándo cambió** |

**Es un archivo vivo**, y eso trae un problema que hoy no mira nadie: el PRD cambia después
de que las historias salieron de él, y **las `us-#` quedan viejas sin que nadie se entere**.

`sf` guarda el hash; cada `us-#` anota de qué versión salió:

```
⚠ El PRD cambió. 4 historias salieron de la versión anterior: us-2, us-5, us-7, us-9.
```

**Avisa, no frena.**

```markdown
---
tipo: prd
id: prd
estado: vivo
deriva_de: brief
actualizado: 2026-08-10
---

# <producto>

## El problema
## Para quién es
## Qué hace          (alcance)
## Qué NO hace       (fuera de alcance)
## Requisitos funcionales
## Requisitos no funcionales
## Cómo sé que salió bien
```

**Sin IDs de requisito, por ahora.** `flujo-real.md` dice que los identificadores aparecen
recién en el ⑨; no los adelantamos. Donde sí valen la pena es en los criterios de aceptación
(ronda 4).

---

## 6. Ronda 3 — la constitución (⑧)

Primer artefacto cuyo consumidor principal **no es un humano ni el hilo de conversación: es
un subagente frío**.

| Consumidor | Para qué |
|---|---|
| el modelo del ⑨ | las historias salen con la constitución armada |
| **el implementador (⑲)** | **es su manual** |
| el revisor (㉑) | ¿se respetó? |
| **`sf`** | **el dolor #6** |

### El dolor #6 estaba mal leído

*"instaló librerías fuera de la constitución **sin consultar**"*. El problema **no es la
librería prohibida: es no enterarte.**

Por eso **no hay lista blanca** —nadie puede listar de antemano lo que va a necesitar, y a la
semana la lista queda vieja—. `sf` compara el manifiesto contra el estado anterior:

```
⚠ Aparecieron 2 dependencias nuevas en go.mod: viper, yaml.v3
  No estaban aprobadas. ¿Las agrego a la constitución?
```

> **La lista no se escribe de antemano: se construye sola, con cada aprobación.**

```markdown
---
tipo: constitucion
id: constitucion
estado: vivo
deriva_de: prd
actualizado: 2026-08-11

lenguaje: go
manifiesto: go.mod              # dónde mira sf
test_cmd: go test ./...         # cómo corre los tests sf
mutacion: gremlins              # herramienta del ㉒; vacío = sólo el modelo

dependencias_aprobadas:
  - github.com/spf13/cobra
  - github.com/go-git/go-git/v5

git:
  branch_por_feature: true
  patron_branch: "feat/{feature-id}-{slug}"
  commit: conventional
  merge: no-ff
---

# Constitución

## Arquitectura
## Stack, y por qué se eligió
## Convenciones de código
## Estructura de carpetas
## Reglas de trabajo
```

**El bloque `git:` mata el dolor #4** (*"no crea la branch por feature"*): `sf` compara la
branch actual contra el patrón y, si no coincide, **el estado no avanza**.

### `test_cmd:` — el campo que faltaba

Se descubrió mirando el CLI construido (`que-sobrevive.md` §5). Sin él **`sf` no puede correr
los tests**, y correr los tests es lo que sostiene tres compuertas: el rojo del ⑲, el verde del
⑳ y el conteo del #8. Era el agujero más barato de todo el recorrido: **una línea**.

Y no hay que escribirla a mano — el `detectStack()` que ya existe llena `lenguaje`,
`manifiesto` y `test_cmd` solo.

### Constitución global

Las reglas de git de Javier son las mismas en todos sus proyectos. Van en
`~/.specforge/constitucion.yaml` y **cada proyecto hereda y puede pisar**.

---

## 7. Ronda 4 — las historias (⑨)

| Consumidor | Para qué |
|---|---|
| el modelo del ⑩ | ordenarlas y agruparlas en features |
| el planificador (⑫–⑯) | es su entrada: el *qué* ya viene resuelto |
| **el revisor (㉑)** | **¿la implementación satisface esta historia?** |
| **el documentador (㉓)** | **la mitad funcional de la doc** — qué se pedía y para quién |
| `sf` | relaciones y el conteo de criterios |

### Los criterios de aceptación con ID — el mecanismo del ㉑

El anexo dice que la completitud contra una historia **es juicio** y ninguna herramienta
determinista la contesta. Sigue siendo cierto. Pero hay una jugada intermedia:

> **`sf` no verifica el juicio. Verifica que el juicio haya ocurrido, y sobre todos.**

Con criterios sueltos en prosa, el revisor contesta *"anda"* y ahí se cuela DeepSeek. Con
`CA-1 / CA-2 / CA-3`, tiene que contestar por cada uno — y `sf` cuenta: *"la historia tiene 3
criterios, el informe habla de 2"*.

```markdown
---
tipo: us                    # us | bug
id: us-3
titulo: "estado en formato json"
deriva_de: prd
prd_version: a3f9c1
relacionado_a: null         # us-# cuando es un bug (entrada C)
---

# us-3 — estado en formato json

Como **<quién>** quiero **<qué>** para **<por qué>**.

## Criterios de aceptación
- **CA-1** — acepta `--json` y devuelve el estado serializado
- **CA-2** — si no hay estado, sale con código 1 y mensaje
- **CA-3** — `--json` es incompatible con `--verbose`

## Contexto
```

### El `us-#` no lleva estado — corregido en el recorrido de lo construido

La versión anterior de esta ronda ponía `estado: pendiente | planificada | … | archivada` en el
frontmatter y lo declaraba *la* fuente de verdad. **Se cayó**, y por dos razones que sólo se
vieron después:

1. **El ciclo corre por feature, no por historia** (§3). Una historia no está *"en curso"* por
   su cuenta: lo está la feature que la contiene.
2. **Marcarlo exige reescribir un `.md`, y `sf` no tiene un LLM adentro** (regla 1.2). Cerrar
   una feature sería reescribir tres archivos de prosa; en `estado.json` es un campo.

> **El avance vive en `estado.json` y en ningún otro lado. El `us-#` guarda el requisito, que
> no cambia porque el trabajo avance.**

### Dónde vive cada dato — para que nada se contradiga

| Dato | Vive en | Y en ningún otro lado |
|---|---|---|
| la historia y sus criterios | `us-#.md` | — |
| el **avance** | `estado.json` — por **feature** | **ni** en el `us-#` **ni** en el roadmap |
| el **orden** y el agrupamiento | `roadmap.json` | **sólo ids** |
| el índice y el backlog | **los genera `sf`** | no existen como archivo a mano |

---

## 8. Ronda 5 — el roadmap (⑩)

| Consumidor | Para qué |
|---|---|
| **Javier, en el ⑪** | *"tomo del roadmap lo que hay que hacer"* |
| **`sf`** | qué feature toca y cuáles quedan |
| el modelo del ⑩ | lo escribe, y lo reordena cuando se lo piden |

**Formato JSON**, decidido por Javier: *"es más fácil seccionarlo"*. Y hay un motivo extra —
**JSON obliga a referenciar**: no hay dónde copiar el texto de la historia, así que el
roadmap no puede contradecir al `us-#`.

```json
{
  "actualizado": "2026-08-11",
  "features": [
    { "id": "f-1", "slug": "nucleo-cli", "nombre": "núcleo del CLI",
      "orden": 1, "historias": ["us-1", "us-3", "us-7"] },
    { "id": "f-2", "slug": "verificacion", "nombre": "verificación",
      "orden": 2, "historias": ["us-2", "us-5"] }
  ]
}
```

**Sin títulos, sin estados, sin descripciones.** Sólo ids y orden.

**Sin dependencias explícitas** (*"us-5 necesita us-3"*): el `orden` ya las expresa, y es un
campo menos que mantener. Si algún día falta, se agrega.

El archivo es para la máquina; **la vista es para Javier**, y la genera `sf`:

```
① núcleo del CLI
   ✓ us-1  inicializar el proyecto
   ▸ us-3  estado en formato json      ← en curso
     us-7  camino corto
```

De acá salen la carpeta y la branch:

```
.docs/features/f-1-nucleo-cli/
branch:  feat/f-1-nucleo-cli
```

### Dos agrupamientos distintos, que no hay que mezclar

| | Qué agrupa | Para qué | Nivel |
|---|---|---|---|
| **⑩ roadmap** | historias | prioridad y entrega | producto |
| **⑭ lotes** | tareas de una feature | correr en paralelo | implementación |

Si se mezclan, el `roadmap.json` empieza a saber de tareas y se vuelve un monstruo.

---

## 9. Ronda 6 — el bloque de planificación (⑫–⑯)

Los cinco pasos pasan en **una sola conversación** y se revisan **una sola vez** en el ⑰.
Después se le pasan a un modelo frío. La pregunta de la ronda: **cuántos archivos salen, y
para quién es cada uno.** Respuesta: **tres**.

### ⑫ — las 3 implementaciones: `decision.md`

| Consumidor | Necesita esto |
|---|---|
| **Javier, en el ⑰** | ver las 3 para juzgar si la elegida es la buena |
| **Javier, meses después** | *"¿por qué no hicimos la otra?"* |
| **el implementador (⑱)** | **NO. Le sobra.** |

→ **archivo propio, separado del spec.** Si las 3 opciones van dentro del `spec-design.md`,
el modelo frío del ⑱ se come dos soluciones que no tiene que hacer: gasta contexto y puede
mezclarlas — o "corregir" hacia la que le parece más linda.

```markdown
---
tipo: decision
feature: f-1
elegida: B
---

## A — <nombre>      qué es · a favor · en contra
## B — <nombre>  ✅  qué es · a favor · en contra
## C — <nombre>      qué es · a favor · en contra

## Por qué B
```

Es el mismo valor que el *"no lo hagas"* del brief: **guardar lo descartado y su porqué**.

#### El veredicto baja al spec, el debate se queda acá

Duda de Javier, y era buena: *"¿no le sirve al implementador saber por qué se descartaron las
otras, o le ocasionaría confusión?"* **Las dos cosas son ciertas a la vez**, así que se
separan:

> **Al implementador le sirve la restricción, no la alternativa.**

Sin saber que A se descartó, se le puede ocurrir A solo y hacerla. Pero con las tres opciones
adentro, el que mezcla es él — y justo el ⑱ es donde entra el modelo del dolor #7.

**La conclusión de cada descarte baja al `spec-design.md`, a la sección "Qué NO entra", en una
línea y con puntero:**

```markdown
## Qué NO entra
- Un pool de workers → descartado: sf muere en milisegundos, no hay a quién poolear.
  (el análisis completo: decision.md)
```

Le alcanza para no reincidir y no le alcanza para confundirse. **Lo descartado viaja como
límite, no como opción viva.**

*(Y si algún día conviene que el `decision.md` completo le llegue igual, es agregarlo a la
lista que devuelva `sf context implement` — ver la idea guardada en `session.md`.)*

### ⑬ — `spec-design.md`

Un solo archivo (ya decidido). **Es lo único que el implementador necesita leer.**

```markdown
---
tipo: spec-design
feature: f-1
historias: [us-1, us-3, us-7]
estado: borrador        # borrador | aprobado  ← lo sella el ⑰
deriva_de: decision
---

## Qué hay que construir
## El diseño            (módulos, tipos, quién llama a quién)
## Interfaces / contratos
## Qué NO entra         ← acá bajan los veredictos del ⑫, una línea cada uno
```

### ⑭ + ⑮ — las tareas y los tests, **en el mismo archivo**

El ⑲ trabaja **por lote**: crea los tests de ese lote, los corre, los ve fallar, implementa.
Necesita tareas y tests **juntos**, en la misma lectura. En dos archivos habría que
sincronizarlos a mano y `sf` tendría que cruzarlos para chequear.

Y **plano, no anidado** — el lote es un campo, no un nivel. Pedir *"las tareas del lote 2"* es
un filtro, y mover una tarea de lote es cambiar un número:

```json
{
  "feature": "f-1",
  "tareas": [
    {
      "id": "t-1",
      "lote": 1,
      "descripcion": "parsear el frontmatter del brief",
      "satisface": ["us-1/CA-1", "us-1/CA-2"],
      "tests": [
        "internal/docs/brief_test.go::TestParseFrontmatter",
        "internal/docs/brief_test.go::TestFrontmatterInvalido"
      ]
    }
  ]
}
```

Este archivo tacha tres dolores, y ninguno necesita que `sf` piense.

**#8 — *"todo verde" sin que haya tests*.** `sf` tiene la **lista exacta** de tests que deben
existir. Buscarlos es un `rg`; ver si corrieron es leer la salida del runner:

```
sf:  ✗ No avanzo.
     Planificados 6 tests para el lote 1. Existen 4.
     Faltan: TestFrontmatterInvalido · TestBriefSinSello
```

**#2 y #3 — el commit y su agrupamiento.** El lote es la unidad de commit: **un lote terminado
= un commit.** No hay que agrupar bien — el agrupamiento **ya se decidió en la planificación**.

**Y la feature a medias, atacada antes de empezar.** Cada tarea declara qué criterios de
aceptación satisface, así que `sf` cuenta:

```
⚠ La feature tiene 9 criterios. Las tareas cubren 7.
  Sin cubrir: us-7/CA-1, us-7/CA-4.
```

No se pisa con el ㉑, porque son preguntas distintas en momentos distintos:

| | La pregunta | Cuándo · quién |
|---|---|---|
| **⑭** | ¿hay tarea para cada criterio? | **antes** de implementar · lo **cuenta** `sf` |
| **㉑** | ¿el código satisface cada criterio? | **después** · lo **juzga** un modelo |

### ⑯ — el modelo: **estado, no archivo**

Es un dato de una línea (regla 1.1), pero el estado guarda más, porque el ⑯ es un lazo
cerrado:

> *"elevo el modelo a uno mejor **porque significa que la recomendación no fue suficiente**"*

```json
"f-1": { "modelo": "deepseek", "intentos_fallidos": 2 }
```

**Un solo campo, no dos.** El recomendado ya está escrito por el ⑯ en `tareas.json`; guardarlo
también en el estado sería guardarlo dos veces. Al estado va **el que se está usando ahora** —
que es lo único que ningún archivo contiene, porque cambia cuando Javier lo sube en el ⑳.

**Así "varias veces" deja de ser una sensación y pasa a ser un número**, y el lazo lo cierra
la máquina en vez de depender de que Javier note que ese lote viene fallando:

```
sf:  El lote 2 falló 3 veces con deepseek.
     La estimación de complejidad del ⑯ se quedó corta.
     ¿Subo el modelo, o entrás vos?
```

### La carpeta de la feature queda así

```
.docs/features/f-1-nucleo-cli/
  decision.md        ⑫   las 3 opciones y por qué B
  spec-design.md     ⑬   lo único que lee el implementador
  tareas.json        ⑭⑮  lotes, tareas, CA que satisfacen, tests planificados
```

El ⑯ y el ⑰ no crean archivos: van al estado.

---

## 10. Ronda 7 — el traspaso, la implementación y el cierre (⑱–㉓)

La última. Y la que menos archivos agrega: **tres** — `revision.json` en el ㉑㉒, y en el ㉓ **la
doc y el journal**. *(El journal entró después, en el recorrido de lo construido: ya estaba
escrito y ningún estado lo había reclamado.)*

### ⑱ — no sale archivo, sale un comando

Acá no se crea nada: se **entrega** lo que ya existe. Lo único a decidir era **qué entra en el
sobre**.

```
sf context implement f-1 --lote 2
  → .docs/constitucion.md                  el manual
  → .docs/features/f-1/spec-design.md      el qué y el cómo
  → .docs/features/f-1/tareas.json         sólo el lote 2
  → us-1.md · us-3.md                      por los criterios de aceptación
```

**Sin `decision.md`** — regla del ⑫: la restricción sí, la alternativa no.

Esto asciende la idea que estaba guardada *"para más adelante"* en `session.md` §6 y la
convierte en **el mecanismo del ⑱**. El motivo es la frase de `flujo-real.md`: acá *"la
propuesta tiene que bastarse sola"*. Si la lista de qué leer la arma el orquestador a mano,
cada vez se olvida algo distinto.

### ⑲ — el rojo es comprobable, y es lo más barato del diseño

El ⑲ tiene tres tiempos y el tercero es *"comprobá que fallan"*. Hoy eso es **una promesa del
modelo**: contesta *"sí, fallan"* y nadie mira.

Pero es un **hecho**, y la lista exacta de tests del lote ya está en `tareas.json` (ronda 6).
Así que `sf` lo mira con sus propios ojos, antes de dejar implementar:

```
sf lote start 2
  ✓ los 6 tests planificados fallan. Rojo confirmado. Podés implementar.
```

```
sf: ✗ TestParseFrontmatter YA PASA, y todavía no se escribió el código.
    Ese test no prueba nada. No avanzo.
```

> **Un test que pasa antes de que exista el código es un test de mentira.**

Es el **#5** (los mocks) y el **#8** (verde sin sustancia) atacados por un *exit code*, sin
juicio y sin agregar un solo artefacto.

#### El agujero astuto: aflojar el test entre el rojo y el verde

`sf` ve rojo → el subagente trabaja → `sf` ve verde. **¿Y si lo que cambió entre medio fue el
test?** Un subagente que no logra implementar puede ablandar el assert y llegar a verde. Es el
#8 en su forma más difícil de ver, y ninguna de las dos compuertas lo tapaba.

**Se tapa con un hash de los archivos de test**, tomado en el rojo y comparado en el verde:

```
sf lote start 2   ✓ los 6 tests fallan. Rojo confirmado.   → rojo: true + hash de los tests
   ...el subagente implementa...
sf lote done 2    ✗ TestFrontmatterInvalido cambió entre el rojo y el verde. No avanzo.
```

Cuesta un campo por lote en el `estado.json`, y no necesita que `sf` piense: es comparar dos
strings.

### ⑲–⑳ — semáforo, no registro

Del par no sale archivo. El registro del trabajo **es el commit del lote**, que ya existe. Al
estado va una línea:

```json
{ "lote": 2, "rojo": true, "hash_tests": "4e91b7", "commit": "a3f9c1e" }
```

**Sin campo `verde`.** Se cayó en la máquina de estados: si `sf` no deja commitear en rojo,
*hay commit* ya significa *estaba verde*. Un campo deducible de otro es un campo que se
desincroniza.

**Su único consumidor es `sf`, en ese instante.** Es un semáforo: su trabajo es no dejar
pasar. Una vez que el lote está verde y commiteado, que el rojo se haya visto no le cambia
nada a nadie.

Queda en el estado igual por una sola razón: cuando algo explota, poder preguntar *"¿este lote
llegó a estar en rojo alguna vez?"*. Cuesta una línea. **Lo que no se hace es darle archivo ni
bitácora — no tiene lector** (regla 1.6).

### ㉑ + ㉒ — un solo artefacto, y en JSON

**Son un archivo, no dos.** Sale de leer el flujo: los hace **el mismo actor** (el modelo
grande), en **la misma pasada**, y cuando hay un arreglo **se rehacen los dos juntos**
(*"se vuelve a correr los mutantes y la verificación"*).

**Y va en JSON, no en markdown** — por la regla 1.2, y el argumento es duro: el hallazgo nace
`abierto` y después pasa a `arreglado`. **Ese cambio ocurre después de que el revisor escribió
el archivo, y no lo hace él: lo hace el flujo.** En markdown, marcar un hallazgo como arreglado
exige un modelo que reescriba prosa — y ahí se corrompe. En JSON es un campo.

```json
{
  "feature": "f-1",
  "vuelta": 2,
  "veredicto": "con-hallazgos",
  "criterios": {
    "us-1/CA-1": "cumple",
    "us-3/CA-2": "no-cumple"
  },
  "mutantes": { "herramienta": "gremlins", "score": 0.88, "sobrevivieron": 1 },
  "hallazgos": [
    { "id": "h-1", "origen": 21, "criterio": "us-3/CA-2", "estado": "arreglado",
      "detalle": "el parser acepta frontmatter sin cerrar; CA-2 pide que falle" }
  ]
}
```

La prosa no se pierde: **es un campo**. Y para que lo lea un humano, `sf` genera la vista
(regla 1.5).

**Acá muere el dolor #7.** `sf` no juzga la revisión — **cuenta que haya ocurrido sobre
todos**, que es el mecanismo de la ronda 4:

```
sf:  La feature tiene 9 criterios. El informe opina sobre 7. No avanzo.
     Sin veredicto: us-7/CA-1, us-7/CA-4.
```

**Y `vuelta:` es el mismo truco que `intentos_fallidos`:** si el ㉑ va por la cuarta, eso no es
ruido — es que la planificación se quedó corta.

### ㉒ — herramienta **y** modelo, en ese orden

No compiten: **miran cosas distintas.**

| | Qué muta | Qué aporta |
|---|---|---|
| **la herramienta** | la **sintaxis** — un `>` por `>=`, borrar una línea, invertir un booleano | exhaustiva, determinista, barata, y da un **número reproducible** |
| **el modelo** | el **sentido** — *"¿y si el frontmatter trae el campo pero vacío?"* | el agujero conceptual, que no es un cambio de operador |

```
1. corre la herramienta                    → barata, exhaustiva, da el score
2. el modelo mira los que SOBREVIVIERON    → decide cuáles importan de verdad
3. y agrega los suyos, los semánticos
```

El modelo deja de inventar a ciegas y trabaja sobre evidencia: gasta menos y apunta mejor. Y
`sf` gana algo que con el modelo solo no tenía — **un número comparable entre vueltas**:

```
sf: mutation score 71% (vuelta 1) → 88% (vuelta 2)
```

**Se declara en la constitución**, al lado del manifiesto (§6): `mutacion: gremlins`. Si el
stack no tiene una herramienta buena, el campo va vacío y queda sólo el modelo — **la
herramienta es una mejora, no un requisito**, igual que los subagentes.

**De los mutantes se guarda el resultado, no los parches.** Los genera un modelo distinto cada
vez: no son reproducibles ni estables, y una vez arreglado el test el parche no se vuelve a
aplicar nunca. Lo que importa entra en la línea `mutantes` de arriba.

### ㉓ — el cierre

Salen **dos** archivos, y por poco se ve uno solo.

#### La doc — y tiene dos mitades

**La documentación es archivo**, y tiene un consumidor que no es obvio: **el ⑫ de una feature
futura**, cuando la pregunta sea *"¿esto ya está resuelto en algún lado?"*. Se escribe al
final, sobre lo que realmente quedó.

Pero *"documentación desde el código"* alcanza para la mitad:

```
TÉCNICA     cómo está hecho        ← del CÓDIGO        el que mantiene · el ⑫ futuro
FUNCIONAL   qué hace y para quién  ← del us-# y la spec  Javier · el usuario
```

**La mitad funcional no está en el código.** Por eso el sobre del ㉓ trae las dos fuentes —
y **el `us-#` gana ahí un consumidor que el inventario no tenía**:

```
sf context documentar f-1
  → git diff base_commit..HEAD    el código que quedó   → la mitad técnica
  → us-1.md · us-3.md             qué se pedía          → la mitad funcional
  → spec-design.md                cómo se resolvió
```

#### El journal — memoria sí, progreso no

El segundo archivo. Cuando la feature se archiva, un modelo extrae **las lecciones durables**
de la vuelta. Su lector es el mismo que el de la doc: **el ⑫ de una feature futura**.

```
sf context planificar f-3
  → constitucion.md · us-#
  → aprendizajes de features anteriores      ← el journal, servido
```

**Se archiva con la feature**, y quien lo encuentra es `sf` (regla 1.4). Lo que **no** hace es
llevar progreso: eso vive en `estado.json` y en ningún otro lado.

#### Marcar como terminada: no se marca en ningún archivo

El ㉓ pide marcarla *"en el roadmap y en el backlog"*. **No se hace ni una vez**: `sf` cierra la
feature en `estado.json` —un campo— y **el roadmap, el backlog y el índice son vistas que
genera él** (regla 1.5). El `roadmap.json` tiene sólo ids; ningún `.md` se reescribe. Cero
puntos de desincronización, y ningún LLM tocando prosa.

**Y "archivar" — la pregunta que `flujo-real.md` había dejado abierta:**

```
.docs/features/f-1-nucleo-cli/   →   .docs/archivado/f-1-nucleo-cli/
```

Se mueve **la carpeta entera con todo adentro** (`decision`, `spec-design`, `tareas`,
`revision`, la doc, el journal), y la feature queda `"estado": "cerrada"` en el `estado.json`.
Los `us-#` **no se tocan**: el requisito no cambia porque el trabajo terminó. Las referencias
por id siguen funcionando porque `sf` es el que resuelve dónde vive cada cosa (regla 1.4).

### La pregunta grande: ¿Javier se entera si lo arregla solo?

> *Cuando ㉑ o ㉒ encuentran algo y el implementador lo arregla, ¿te enterás, o la máquina
> corrige en silencio si terminó bien?*

**Sí, pero al final, y sin frenar.** Si `sf` para en cada hallazgo, vuelve a ser Javier el que
empuja — justo el dolor #1. Si no avisa nunca, el ㉑ se vuelve un trámite. El punto medio ya
existía en el diseño: **la parada barata** de §2.

```
sf:  f-1 lista para archivar.
     El ㉑ fue 2 vueltas. 2 hallazgos, los dos arreglados:
       h-1  us-3/CA-2 no se cumplía
       h-2  un mutante sobrevivió en brief_test.go
     2 aprendizajes nuevos.
     ⚠ Uno se repite por tercera vez:
       "los tests de tabla en Go tienen que nombrar el caso"
       ¿lo subo a la constitución?
     [enter] archivo   [c] subir a la constitución   [v] ver el detalle
```

Y el rastro queda en `revision.json` y en el journal, esté o no ese enter.

**El *backprop* no necesita una parada nueva.** Que una lección repetida ascienda a la
constitución es una decisión de Javier — y cabe entera en la ⏸ que el ㉓ ya tenía.

---

## 11. El inventario completo

Todo lo que el flujo produce, y nada más que eso.

### Los archivos

| Paso | Artefacto | Formato | Lo lee |
|---|---|---|---|
| ①–⑥ | `brief.md` | md + frontmatter | el ⑦ · Javier meses después · `sf` (el sello) |
| ⑦ | `prd.md` | md + frontmatter | el ⑧ y el **⑨** · `sf` (el hash) |
| ⑧ | `constitucion.md` | md + frontmatter | el ⑨ · **el implementador** · el ㉑ · `sf` (deps, git, `test_cmd`, mutación) |
| ⑨ | `us-#.md` | md + frontmatter | el ⑩ · el planificador · el ㉑ · **el ㉓** · `sf` (los CA) |
| ⑩ | `roadmap.json` | **json** | `sf` · el modelo del ⑩ |
| ⑫ | `decision.md` | md + frontmatter | **sólo Javier** — el implementador no |
| ⑬ | `spec-design.md` | md + frontmatter | **el implementador**, y es lo único que necesita |
| ⑭⑮ | `tareas.json` | **json** | el ⑲ · `sf` (tests, CA, lotes) |
| ㉑㉒ | `revision.json` | **json** | el implementador (a arreglar) · `sf` (cuenta) · Javier (vista) |
| ㉓ | la doc de la feature | md | Javier · **el ⑫ de una feature futura** |
| ㉓ | `journal.md` | md | **el ⑫ de una feature futura**, vía `sf context planificar` |

Los tres JSON son exactamente los tres lugares donde `sf` **escribe o consulta datos**. El
resto es prosa con una cabecera.

### El estado — lo único que no es de nadie más que de `sf`

**El borrador que estaba acá quedó viejo.** La forma cerrada vive en
[`maquina-estados.md`](maquina-estados.md) §9 — **acá no se copia: se apunta.** Lo que cambió,
en tres líneas:

| | |
|---|---|
| se fue `"paso"` | no hay un paso global: cada feature tiene el suyo |
| se fue `modelo_recomendado` | lo escribe el ⑯ en `tareas.json`. No se guarda dos veces |
| se fue `verde` | deducible de `commit` (§10) |
| se sumó `base_commit` · `producto` · `estado` por feature | el envejecimiento, los sellos del ⑥ y el ⑧, y la cola |
| se sumó `hash_tests` por lote | el agujero astuto del ⑲ (§10) |

### Los ocho dolores, y dónde muere cada uno

| # | Dolor | Muere en |
|---|---|---|
| 1 | lanzar cada fase a mano | **la máquina de estados** — `sf` dice qué sigue, el orquestador lanza |
| 2 | el commit no se hace | **el lote** (⑭) — un lote terminado, un commit |
| 3 | commits mal agrupados | **el lote** (⑭) — el agrupamiento ya se decidió en la planificación |
| 4 | no crea la branch | **`git:` en la constitución** (⑧) — `sf` compara y no avanza |
| 5 | métodos que son mocks | **el subagente por lote** + **el rojo del ⑲** |
| 6 | librerías fuera de la constitución | **el manifiesto** (⑧) — compara contra el estado anterior, y **avisa** |
| 7 | *"terminado"* con media historia | **`revision.json`** — `sf` cuenta los criterios sin veredicto |
| 8 | *"todo verde"* sin tests | **`tareas.json`** — la lista exacta de tests; buscarlos es un `rg` |

**Ninguno necesita que `sf` piense.** Todos son comparar, contar o correr algo.

---

## 12. Lo que sigue

La prueba de escritorio está cerrada, y las cuatro cosas que seguían **ya se hicieron**: el
strawman de los ocho dolores (`session.md` §5), la máquina de estados y el `estado.json`
([`maquina-estados.md`](maquina-estados.md)) y el recorrido de lo construido
([`que-sobrevive.md`](que-sobrevive.md)).

**Queda una sola ronda de diseño:** la superficie de `sf` y el reparto orquestador ↔ skills.
El punto de retomada está al final de [`session.md`](session.md).

---

## 13. Las correcciones del 2026-08-13

Vinieron del recorrido de lo construido (`que-sobrevive.md` §15) y son **de coherencia**: sin
ellas, dos documentos decían cosas distintas.

| Qué decía antes | Qué dice ahora | Por qué |
|---|---|---|
| el `estado:` del `us-#` es la fuente de verdad del avance (§7, §10) | **el `us-#` no lleva estado**; el avance vive en `estado.json`, por feature | el ciclo corre por feature (§3), y marcar un `.md` exige un LLM (regla 1.2) |
| el ㉓ agrega un archivo | **agrega dos: la doc y el journal** | el journal ya estaba construido y ningún estado lo había reclamado |
| el `us-#` lo leen el ⑩, el planificador y el ㉑ | **+ el ㉓** | la mitad funcional de la doc no sale del código |

Y **dos campos nuevos** que la ronda 7 no tenía:

- **`test_cmd:`** en la constitución (§6) — sin él `sf` no puede correr los tests, y tres
  compuertas dependen de eso.
- **`hash_tests`** por lote en el `estado.json` (§10) — tapa el aflojado del test entre el
  rojo y el verde.
