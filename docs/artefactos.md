# Los artefactos

Qué archivos existen, dónde viven, y qué forma tienen.

```
<tu proyecto>/
  CLAUDE.md · AGENTS.md            el orquestador — el mismo texto
  .docs/
    brief.md                       ⑤  el ARGUMENTO, con su veredicto
    entrevista.md                  ①–⑤ las rondas tal cual pasaron
    evidencia.md                   ①–⑤ lo que se encontró, con procedencia
    vocabulario.md                 ·  las palabras del negocio — PEREZOSO: sólo si hizo falta
    prd.md                         ⑦  qué hay que poder hacer
    constitucion.md                ⑧  las reglas — la lee TODO el mundo
    estado.json                    ·  dónde estás. LO ESCRIBE SÓLO sf
    roadmap.json                   ⑩  el orden de las features
    backlog/
      us-1.md · us-2.md · …        ⑨  las historias, con criterios con id
    features/
      f-1-nucleo/
        decision.md                ⑫  las 3 opciones y por qué se eligió una
        spec-design.md             ⑬  qué se construye y cómo
        tareas.json                ⑭⑮ las tareas, sus lotes y sus tests
        revision.json              ㉑㉒ el veredicto por criterio y los hallazgos
        doc.md                     ㉓  la documentación
        journal.md                 ㉓  las lecciones durables
    archivado/
      f-1-nucleo/                  la carpeta entera, movida al cerrar

~/.specforge/
  modelos.yaml                     qué modelos tenés y cómo se invoca cada uno
```

---

## Las cinco reglas que explican esta forma

**1. Lo que se puede deducir, no se guarda.**
No hay campo de progreso en ningún lado: para saber por dónde retomar, `sf` mira **qué archivos
existen**. ¿Hay `decision.md`? ¿Hay `spec-design.md`? Ahí está el checkpoint.

**2. Un campo deducible de otro es un campo que se desincroniza.**
Por eso no existe `verde` en el estado: si `sf` no deja commitear en rojo, **tener commit ya
significa que estaba verde**.

**3. `.json` sólo donde `sf` cuenta.**
Tres archivos son JSON —`estado`, `roadmap`, `tareas` (y `revision`)— y son exactamente donde `sf`
cuenta, filtra o compara. Todo lo demás es Markdown, porque lo lee un modelo.

**4. Referenciar por id, nunca copiar.**
El roadmap y las tareas guardan `us-3`, nunca su texto. Por eso podés reescribir una historia sin
tocar ningún JSON.

**5. Lo que falta se DICE, no se omite.**
Un sobre al que le falta el diff **lo dice**. Si la parte simplemente no apareciera, el que
trabaja creería que no hacía falta.

---

# Los de producto

## Los tres del ①–⑤ — y por qué son tres

El primer tramo deja **tres archivos, no uno**, y cada uno tiene un lector distinto:

```
brief.md        el ARGUMENTO       envejece con cada vuelta del ⑥
entrevista.md   el RAZONAMIENTO    se relee después: "¿por qué decidimos esto?"
evidencia.md    el HECHO           se acumula, no se reescribe
```

**Los tres son markdown con frontmatter, y eso no es decoración: es un archivo con dos lectores.**
El cuerpo lo leés vos; el frontmatter lo cuenta la compuerta sin parsear prosa.

### `brief.md`

```markdown
---
veredicto: hacelo   # hacelo | pivotea | no-lo-hagas — lo sellás vos en el ⑥
---

# <producto> — brief

<la idea en UNA oración>
```

El brief **cita** a los otros dos, no los copia. Si una sección se puede reemplazar por un
puntero, es un puntero.

### `entrevista.md`

```markdown
---
rondas: 4
preguntas: 23
abiertas: 0      ← el ⑥ NO sella con esto en más de cero
---
```

El ①–⑤ es un árbol de decisiones que se trabaja por rondas, y **termina cuando la frontera queda
vacía** — no cuando se hizo largo. `abiertas: 0` es esa frase, escrita como número.

**Ojo con el campo ausente:** no escribirlo **no** vale como cero. La compuerta frena y lo dice,
porque si valiera cero, olvidarse sería más barato que terminar la entrevista.

### `evidencia.md`

```markdown
---
retrieved: 17
model_prior: 1
probado: 0
links: 13
# evidencia: baja   ← SÓLO en una pasada degradada, y es una declaración
---

- Hay tres paquetes publicados que hacen esto. [retrieved]
  - https://registry.npmjs.org/-/v1/search?text=…
- La categoría se siente saturada. [model-prior — sin verificar]
- Busqué quejas en Reddit y no llegué. [no evaluable]
```

**La procedencia es lo mejor que aporta este artefacto**, porque **el ⑥ es el único punto del flujo
donde una alucinación cuesta el producto entero**: sellar *"no lo hagas"* porque existe algo que no
existe. Tres etiquetas, y la tercera es nueva:

```
retrieved      lo leí, con link
model-prior    me lo acuerdo, sin verificar
probado        lo construí y lo vi          ← la deja un prototipo, y es la más fuerte
```

**El frontmatter DECLARA y el cuerpo es el HECHO.** La compuerta cuenta los `http` del cuerpo; si
el número declarado no cierra con lo que hay, avisa — y no frena, porque frenar por eso sería
opinar sobre prolijidad.

Y lo que **no** puede: visitar los links. Una compuerta no sale a la red, así que cuenta y no
verifica. Eso sale escrito en cada acta, para que "13 links" no se lea como "13 fuentes
verificadas".

### `vocabulario.md`

El glosario del negocio, y el único **perezoso** de todos: no existe hasta que una palabra se
tambalea, y su ausencia no es una falta. Cuando existe, `sf context` lo sirve en **todos** los
sobres de ahí en adelante — se construye una vez y los demás estados heredan las palabras.

```markdown
---
terminos: 7
---

## Corrida
Una ejecución completa de la máquina sobre una idea, de punta a punta.

**No es** una vuelta del bucle de implementación — eso es un *lote*.
```

Cada entrada dice **qué no es**, que es la mitad donde muere la ambigüedad. No se llama
`CONTEXT.md` aunque venga de ahí: `sf context` ya es el sobre.

## `prd.md`

Actores · capacidades · restricciones no funcionales · alcance y **no**-alcance · dependencias
externas.

**Y nada más**, porque tiene exactamente **dos lectores**: el ⑧ y el ⑨. Ni personas con foto, ni
journeys, ni métricas de negocio, ni pricing — **una sección que nadie consume es una sección que
envejece y contradice al código.**

## `constitucion.md`

```yaml
---
lenguaje: go                      # lo llenó sf init
manifiesto: go.mod                # lo llenó sf init
test_cmd: go test ./...           # lo llenó sf init — SIN ESTO no sella
mutacion: "gremlins"              # la del ㉒ — sf init deja un ⚠ con la de tu stack
dependencias_aprobadas: []        # arranca vacía y CRECE con cada aprobación
git:
  branch_por_feature: true
  patron_branch: "feat/{feature-id}-{slug}"
  commit: conventional
  merge: no-ff                    # no-ff | squash | ff
  branch_base: main
---
```

**El frontmatter es donde mueren dos dolores:**

| Campo | Qué mata |
|---|---|
| `git:` | *"no crea la branch por feature"* — deja de ser algo que un skill recuerda y pasa a ser un dato que `sf` lee |
| `manifiesto:` + `dependencias_aprobadas:` | *"librerías fuera de la constitución"* — ⚠ **todavía sólo lo cubre el skill**: falta el parser de manifiestos ([`problemas.md`](problemas.md#dependencias-fuera-de-la-constitución--️-todavía-no-avisa)) |

> **Avisa, no frena**, aunque frenar sería trivial: la última palabra es tuya, y **una herramienta
> que frena sola rompe esa regla.**

El cuerpo es técnico: *Arquitectura · Stack y por qué · Convenciones de código · Estructura de
carpetas · Reglas de trabajo*. **Cero principios abstractos.**

`mutacion:` vacío **no es un error**, pero tiene que ser una **decisión**: si tu stack no tiene
una herramienta buena, el ㉒ lo hace el modelo leyendo el código, y la constitución dice por qué.

La diferencia importa porque cambia lo que significa el número del ㉒. Con herramienta, la corrida
muta el mismo código igual todas las vueltas y el score se compara. Sin ella, el revisor escribe
los mutantes a mano y el set es distinto cada vez — 56 una vuelta, 82 la siguiente — así que
`73% → 85%` parece progreso y es **otro examen**. Por eso los mutantes del modelo se guardan
(`.docs/<feature>/mutantes/`) y la vuelta siguiente corre los mismos.

> Antes esta línea venía escrita en `""` con un comentario al lado que decía *"o vacío"*. El ⑧
> leía una casilla ya completada con una respuesta válida y seguía de largo: nadie elegía nunca.
> Hoy `sf init` deja un `⚠` nombrando la herramienta de tu lenguaje, y la compuerta del ⑧ **avisa**
> —no frena— si quedó vacía en un stack que tiene una.

## `us-#.md` — las historias

```markdown
---
tipo: us                    # us | bug
id: us-7
titulo: "…"
deriva_de: prd
prd_version: a3f9c1
relacionado_a: us-3         # el us-# original, cuando es un bug
---

# us-7 — <título>

Como **<rol>** quiero **<qué>** para **<por qué>**.

## Criterios de aceptación
- **CA-1** — acepta --json y devuelve el estado serializado
- **CA-2** — si no hay estado, sale con código 1 y mensaje
```

**El `us-#` es 100% humano: el texto y los criterios.** No tiene campo `estado`, y los cinco
valores que tendría son todos deducibles:

```
pendiente    existe el us-#, no está en roadmap.json
planificada  está en roadmap.json
en-curso     su feature es feature_actual
hecha        su feature figura cerrada en estado.json
archivada    la carpeta se movió
```

> Si un modelo se olvida de tocar el frontmatter al cerrar la feature, **el archivo miente para
> siempre y nadie se entera**. Lo que se calcula no puede quedar viejo.

**`relacionado_a` tiene un consumidor concreto:** cuando implementás un bug, el sobre te trae **la
spec archivada de la feature que rompió**. Sin eso arrancarías de cero sin saber qué se había
decidido — que es donde nacen los mocks.

## `roadmap.json`

```json
{
  "actualizado": "2026-08-15",
  "features": [
    {"id": "f-1", "slug": "nucleo-cli", "nombre": "núcleo del CLI",
     "orden": 1, "historias": ["us-1", "us-3"]}
  ]
}
```

**`orden` es el único campo que expresa dependencias.** Si `us-5` necesita `us-3`, sus features van
en ese orden y listo. No hay `depends_on`, ni MoSCoW, ni prioridades.

> Must/Should/Could es una herramienta de **negociación**, y acá **no hay con quién negociar**.
> Cada pregunta que contestaría ya está contestada: *¿la construimos y cuándo?* → `orden`. *¿puede
> fallar y salir igual?* → **existir o no existir**.

**El `slug` es load-bearing:** de él salen la carpeta (`.docs/features/f-1-nucleo-cli/`) y la
branch (`feat/f-1-nucleo-cli`).

---

# Los de feature

## `decision.md` — el ⑫

Exactamente **tres** opciones, con encabezados `## A — …`, `## B — …`, `## C — …`, y después la
elegida con **el argumento que inclinó la balanza** y qué se descartó.

`sf` cuenta los encabezados y **exige tres**. Y lo descartado es lo más valioso del archivo para el
que lo lea en seis meses: dice **qué ya se pensó y no hace falta volver a pensar**.

> **No va en el sobre del implementador**, a propósito. La decisión ya se tomó.

## `spec-design.md` — el ⑬

Un archivo, con la regla anti-N/A:

> Poné una sección **sólo si cambia una decisión o informa la implementación**. Una que diría
> "N/A" es ruido — y el ruido acá se paga en el sobre de **cada lote**.

## `tareas.json` — el ⑭⑮

```json
{
  "feature": "f-1",
  "modelo": "opus",
  "tareas": [
    {"id": "t-1", "lote": 1,
     "descripcion": "parsear el frontmatter y devolver el veredicto",
     "satisface": ["us-3/CA-1", "us-3/CA-2"],
     "tests": ["internal/docs/brief_test.go::TestParseFrontmatter"]}
  ]
}
```

**Las tareas y los tests viven en el mismo archivo** porque el ⑲ trabaja por lote y necesita las
dos cosas en la misma lectura.

**Es plano, no anidado**, y el `lote` es un campo. Las dos operaciones reales son más baratas
así: *"dame las tareas del lote 2"* es un filtro, y *"mové esta tarea al lote 3"* es cambiar un
número.

**El lote es tres cosas a la vez**, y por eso lo declara el modelo y no un orden topológico:

```
un lote  =  la unidad de COMMIT        un lote terminado = un commit
         =  la unidad de SUBAGENTE     uno por lote, contexto fresco
         =  el filtro de sf context    "dame las tareas del lote 2"
```

> Un orden topológico agrupa por *cuándo se puede*; hace falta agrupar por *qué va junto*. Dos
> tareas sin dependencia entre sí pueden caer en la misma ola **sin tener nada que ver** — y ese
> lote produce un commit incoherente y un subagente con dos temas en la cabeza.

**`tests` es toda la defensa contra el "todo verde" sin tests**: es la lista exacta que `sf lote
start` va a exigir en rojo, **por nombre**.

**`modelo` es opcional** y es el ⑯: *"esta feature necesita uno más grande"*. Vacío es lo normal.

## `revision.json` — el ㉑㉒

```json
{
  "feature": "f-1", "vuelta": 1, "veredicto": "con-hallazgos",
  "criterios": {"us-3/CA-1": "cumple", "us-3/CA-2": "no-cumple"},
  "mutantes": {
    "herramienta": "gremlins", "score": 71.0, "sobrevivieron": 4,
    "propios": {"corridos": 18, "sobrevivieron": 3, "resucitados": 1, "viejos": 2}
  },
  "hallazgos": [
    {"id": "h-1", "origen": 21, "criterio": "us-3/CA-2", "estado": "abierto",
     "detalle": "el camino de error no tiene ningún test que lo ejerza"}
  ]
}
```

**Va en JSON y no en Markdown**, y el argumento es duro: el hallazgo nace `abierto` y **después
cambia de estado**, y ese cambio ocurre **después** de que el revisor escribió el archivo. En
markdown, marcarlo exigiría un modelo que reescriba prosa — y ahí se corrompe.

**`criterios` es un mapa** porque la pregunta real siempre es *"¿opinó sobre ESTE?"*.

**El ㉒ tiene dos fuentes y no hacen lo mismo.** Una herramienta muta *sintaxis* —da vuelta un
`>`, borra una línea, niega un `if`—: es barata, amplia, y genera lo mismo siempre sobre el mismo
código. Lo que no se le ocurre es *"y si arranca dos timers"*, *"y si nunca lo apaga al salir"*:
eso es mutar la **intención**, y lo piensa un modelo o no lo piensa nadie.

**Los de la herramienta no se guardan; los del modelo sí**, y es la misma razón leída de los dos
lados. La herramienta se regenera igual sola. El modelo genera un set distinto cada vuelta, y ahí
el score deja de ser comparable aunque parezca un número:

```
73% → 85%   sobre exámenes distintos NO es una mejora, es otra pregunta
```

Guardarlos en `.docs/<feature>/mutantes/` —una carpeta que se archiva con la feature— es lo que
convierte el número en un hecho: la vuelta siguiente corre **los mismos**.

**Y ahí aparece el caso que sin esto no ve nadie.** Un mutante que **murió** en una vuelta y
**vive** en ésta significa que había un test que lo agarraba y ya no: una regresión de cobertura.
Un sobreviviente es un agujero que nunca se tapó; un resucitado es uno que se tapó y se destapó.
Es lo mismo que `sf done` vigila comparando los archivos de test entre el rojo y el verde, pero
DENTRO de un lote — esto lo ve entre vueltas y entre lotes. **Cada resurrección es un hallazgo**,
y `sf` frena si se reportan resurrecciones sin ningún hallazgo del ㉒.

**Sólo hay dos estados de hallazgo**, `abierto` y `descartado`. No existe `arreglado`: la revisión
se rehace entera, y **no reaparecer** *es* estar arreglado.

## `doc.md` y `journal.md` — el ㉓

**La doc tiene dos mitades y sólo una sale del código:** la técnica sale del diff, la funcional de
los `us-#`. Un documentador que sólo ve el código no puede decir **por qué alguien lo quería**.

**El journal son lecciones, nunca progreso** — el progreso vive en `estado.json`. Se archiva
**con** la feature, y `sf context` se lo sirve al ⑫ de cada feature futura: *"¿esto ya está
resuelto en algún lado?"* y *"¿qué aprendimos la vez pasada?"*.

---

# Los dos que escribe `sf`

## `estado.json` — el único que `sf` escribe

```json
{
  "producto": {
    "brief_sellado": "hacelo",
    "prd_hash": "a3f9c1",
    "constitucion_sellada": true,
    "backlog_visto": true
  },
  "feature_actual": "f-1",
  "features": {
    "f-1": {
      "estado": "implementar",
      "base_commit": "8bbf5a2",
      "modelo": "",
      "intentos_fallidos": 0,
      "lotes": [
        {"lote": 1, "rojo": true, "hash_tests": "a3f9…", "commit": "8bbf5a2"},
        {"lote": 2, "rojo": true, "hash_tests": "c71b…", "commit": null}
      ]
    }
  }
}
```

> **Los archivos dicen QUÉ SE PRODUJO. El estado dice QUÉ SE APROBÓ** — y eso no se puede deducir
> de ningún archivo.

**Tres campos son indeducibles y por eso existen:**

| Campo | Por qué no se puede deducir |
|---|---|
| `rojo` | **el momento en que los tests fallaban YA PASÓ** y no dejó huella. Si `sf` no lo anota cuando lo ve, se pierde para siempre |
| `modelo` | es el que se usa **ahora**, no el recomendado. Cuando lo subís con `sf model`, eso no queda en ningún archivo |
| `backlog_visto` | la ⏸ del ⑨ es un enter y **no sella nada**. Sin el campo, `sf next` la repetiría para siempre |

**Y `commit` es un puntero (`null` o un hash) por una razón que carga significado:** es la razón
por la que **no existe** un campo `verde`. Si `sf` no deja commitear en rojo, **tener commit ya
significa que estaba verde**.

**No lo edites a mano.** Es el único archivo que `sf` escribe, y se escribe entero y atómico.

## `~/.specforge/modelos.yaml`

```yaml
harness: claude-code
modelos:
  opus:     {via: subagente}
  sonnet:   {via: subagente}
  haiku:    {via: subagente}
  deepseek: {via: consola, comando: deepseek exec}
```

**Vive en tu home y no en el proyecto** porque tus modelos son los mismos en todos tus repos —
igual que tus reglas de git.

**El harness va en el mismo archivo** porque los dos contestan **una sola pregunta**: *¿cómo lanzo
este modelo acá?* En Claude Code no podés usar un modelo que no sea de Anthropic como subagente;
en otro harness quizá sí. **`sf` es el único que ve las dos mitades.**

**Y la lista crece sola**, una aprobación a la vez. Ver [`comandos.md`](comandos.md#sf-model-nombre---via---comando-).

---

## Qué se versiona

**`.docs/` va al repo. Todo.** No lo pongas en `.gitignore`:

> El estado tiene que **viajar con el repo**, porque cuando pasás a otro modelo el que implementa
> **no estuvo en la conversación**.

`~/.specforge/` no, obviamente: es de tu máquina.
