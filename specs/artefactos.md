# Los artefactos del flujo — prueba de escritorio

**Fecha:** 2026-08-10 / 11 · **Método:** paso a paso por `flujo-real.md`, preguntando en cada
uno: **qué artefacto sale · quién lo consume · qué formato · qué estructura**.

> **Por qué así.** Este tema hundió la versión anterior de SpecForge. Se estructuró de más y
> se decidió el formato antes de saber quién iba a leer cada cosa. Acá el orden es al revés:
> **primero el consumidor, después el formato.**

**Estado:** rondas 1 a 6 cerradas (①–⑯). Falta el cierre del ciclo por feature, ⑰–㉓.

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
| `sf` | estado, relaciones, y el conteo de criterios |

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
estado: pendiente           # pendiente | planificada | en-curso | hecha | archivada
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

### Dónde vive cada dato — para que nada se contradiga

| Dato | Vive en | Y en ningún otro lado |
|---|---|---|
| la historia y sus criterios | `us-#.md` | — |
| el **estado** de la historia | frontmatter del `us-#.md` | **no** se copia al roadmap |
| el **orden** y el agrupamiento | `roadmap.json` | **sólo ids** |
| el índice | **lo genera `sf`** | no existe como archivo a mano |

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
"f-1": { "modelo_recomendado": "deepseek", "modelo_actual": "deepseek", "intentos_fallidos": 2 }
```

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

## 10. Lo que falta

**Ronda 7 — el traspaso, la implementación y el cierre: ⑱–㉓.**

- ⑱ el traspaso al modelo frío — qué se le entrega exactamente, y con qué mecanismo
- ⑲–⑳ la implementación por lote — qué queda registrado de cada lote
- ㉑ el informe de la revisión — hoy vive en el chat y se pierde
- ㉒ los mutantes — ¿se guardan los parches o sólo el resultado?
- ㉓ el cierre: documentación y archivado

Y una pregunta de `session.md` §7 que la ronda 7 tiene que contestar: **cuando ㉑ o ㉒
encuentran algo y el implementador lo arregla solo, ¿te enterás, o la máquina corrige en
silencio si terminó bien?**
