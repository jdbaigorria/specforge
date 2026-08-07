# Cómo trabajo hoy — foja cero

**Fecha:** 2026-08-07 · **Regla de este documento:** acá **no existe SpecForge**. Ni
comandos, ni módulos, ni nombres de fases. Sólo qué pasa, quién lo hace y qué queda.

Cuando esto esté bien, recién ahí se piensa la herramienta.

---

## El flujo

```
  ┌─────────────────────────────────────────────────────────────────┐
  │  ①  SE ME OCURRE ALGO                                           │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ╔═════════════════════════════════════════════════════════════════╗
  ║        ACÁ SE ARMA EL BRIEF  —  y es un BUCLE, no una fila      ║
  ╠═════════════════════════════════════════════════════════════════╣
  ║                                                                 ║
  ║   ②  LE TIRO LA IDEA A LA IA Y PINPONEO                         ║
  ║      Claude, Opus, effort xhigh · "para darle forma"            ║
  ║                          │                                      ║
  ║                          ▼                                      ║
  ║   ③  BUSCO SI YA EXISTE                                         ║
  ║      Perplexity · herramientas similares                        ║
  ║                          │                                      ║
  ║                          ▼                                      ║
  ║   ④  LE PASO EL REPO A LA IA PARA QUE LO REVISE                 ║
  ║      ¿me sirve algo? ¿qué diferencia tiene lo mío?              ║
  ║      ¿se complementa? ¿por qué crearlo en vez de usarlo?        ║
  ║      ¿conviene forkearlo?                                       ║
  ║                          │                                      ║
  ║                          ▼                                      ║
  ║             ┌────────────────────────┐                          ║
  ║             │  ⑤  ¿HAY CONSENSO?     │──── no ──┐               ║
  ║             │  qué quiero crear      │          │               ║
  ║             │  y POR QUÉ             │◄─────────┘               ║
  ║             └───────────┬────────────┘   vuelve a ②/③/④         ║
  ║                         │ sí                                    ║
  ╚═════════════════════════╪═══════════════════════════════════════╝
                            │
                            ▼   ✔ el BRIEF está terminado
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑥  LA DECISIÓN  —  acá se sella el brief                       │
  │                                                                 │
  │     la IA RECOMIENDA:  hacelo / pivoteá / no lo hagas           │
  │                        …con sus razones                         │
  │                                                                 │
  │     YO SELLO.  ← la última palabra es mía, SIEMPRE              │
  │                  la IA nunca sella sola, ni frena nada sola     │
  └───────┬──────────────────────┬──────────────────────┬───────────┘
          │                      │                      │
      NO LO HAGAS            PIVOTEÁ                  HACELO
          │                      │                      │
          ▼                      ▼                      ▼
   se archiva.            vuelve al ②.              sigue
   Queda el brief         La idea cambia y             │
   diciendo POR QUÉ       el brief se reescribe        │
   NO lo hice                                          ▼
                                        ┌──────────────────────────┐
                                        │  ⑦  LE PIDO EL PRD       │
                                        │                          │
                                        │  misma conversación —    │
                                        │  la IA ya tiene el brief │
                                        │  en el contexto          │
                                        │                          │
                                        │  "generame un prd        │
                                        │   completo sobre el      │
                                        │   producto que estuvimos │
                                        │   conversando"           │
                                        │                          │
                                        │  → sale un PRD completo, │
                                        │    de una sola vez       │
                                        └────────────┬─────────────┘
                                                     ▼
                                        ┌──────────────────────────┐
                                        │  BAJO LOS .md AL REPO    │
                                        │                          │
                                        │  carpeta  .docs/         │
                                        │  adentro de un git que   │
                                        │  siempre inicializo      │
                                        │                          │
                                        │  → quedan versionados    │
                                        └────────────┬─────────────┘
                                                     ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑧  LA CONSTITUCIÓN DEL PROYECTO                                │
  │      arquitectura + stack tecnológico                           │
  │                                                                 │
  │      entra: el PRD                                              │
  │                                                                 │
  │      la IA me PREGUNTA y me SUGIERE                             │
  │        "¿qué lenguaje vas a usar?" + una sugerencia             │
  │                        │                                        │
  │                        ▼                                        │
  │      YO DECIDO.  ← la última palabra es mía, otra vez           │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑨  LAS HISTORIAS DE USUARIO → EL BACKLOG                       │
  │                                                                 │
  │      con la constitución ya armada, se las pido                 │
  │                                                                 │
  │      .docs/backlog/                                             │
  │        ├── backlog.md   ← el índice                             │
  │        ├── us-1.md                                              │
  │        ├── us-2.md                                              │
  │        └── us-#.md                                              │
  │                                                                 │
  │      ← primera salida que no es para leer sino para TRABAJAR    │
  │      ← primera vez que aparecen IDs estables                    │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑩  EL ROADMAP DE TRABAJO                                       │
  │      entra: el backlog                                          │
  │                                                                 │
  │      ORDENA  +  AGRUPA   →   prioridad: qué va antes            │
  │                                                                 │
  │      ← no agrega contenido nuevo: pone ORDEN sobre lo que ya    │
  │        existe. Es una vista del backlog, no otro backlog        │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑪  TOMO DEL ROADMAP LO QUE HAY QUE HACER                       │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑫  "BUSCÁ 3 IMPLEMENTACIONES POSIBLES Y QUEDATE CON            │
  │       LA MÁS ÓPTIMA"                                            │
  │                                                                 │
  │      ← TRES. Un número, no "varias"                             │
  │      ← segundo momento de divergir del flujo entero             │
  │        (el primero fue el pinponeo del ②)                       │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼ ya hay una solución elegida
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑬  "ESCRIBÍ LA ESPECIFICACIÓN JUNTO CON EL DISEÑO"             │
  │                                                                 │
  │      ← de LA SOLUCIÓN ELEGIDA, no del requisito                 │
  │        (el requisito ya venía del us-#)                         │
  │      ← van JUNTOS, en un solo pedido                            │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑭  "DIVIDÍ LA IMPLEMENTACIÓN EN TAREAS,                        │
  │       AGRUPADAS PARA QUE CORRAN EN PARALELO"                    │
  │                                                                 │
  │      ← el agrupamiento tiene un motivo declarado:               │
  │        PARALELIZAR. No son etapas, son lotes                    │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑮  "DECIME QUÉ TESTS HAY QUE CREAR                             │
  │       ANTES DE CUALQUIER IMPLEMENTACIÓN"                        │
  │                                                                 │
  │      ← los tests van PRIMERO. Dicho explícito                   │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑯  "SEGÚN LA COMPLEJIDAD, RECOMENDAME QUÉ MODELO USAR"         │
  │                                                                 │
  │      lee:  .docs/<modelos>.yaml   ← catálogo propio             │
  │      ← complejidad → modelo, es una cadena                      │
  │      ← primera decisión sobre QUÉ RECURSO gastar                │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑰  LO REVISO                                                   │
  │      tercer punto de decisión del flujo, y cierra el bloque      │
  │      de planificación ⑫–⑯                                       │
  └───────┬──────────────────────┬──────────────────────┬───────────┘
          │                      │                      │
     PIDO CAMBIOS          OTRA FEATURE            IMPLEMENTAR
          │                      │                      │
          ▼                      ▼                      ▼
   vuelve a ⑫–⑯         vuelve a ⑪ y agarro       sigue
   se rehace lo que      otra cosa del roadmap        │
   no me cerró           ← se pueden planificar       │
                           VARIAS features antes      │
                           de implementar ninguna     │
                                                      ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑱  LE PASO LA PROPUESTA AL MODELO RECOMENDADO                  │
  │                                                                 │
  │      ← PRIMER TRASPASO del flujo. Hasta acá era una sola        │
  │        conversación; ahora el plan va a OTRO ejecutor           │
  │      ← acá se consume la recomendación del ⑯: el yaml           │
  │        de modelos deja de ser un catálogo y decide algo         │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑲  IMPLEMENTA — y SIEMPRE en este orden                        │
  │                                                                 │
  │      1. crea los tests                                          │
  │      2. LOS EJECUTA                                             │
  │      3. COMPRUEBA QUE FALLAN        ← paso propio, no adorno    │
  │      4. recién ahí implementa la solución                       │
  │                                                                 │
  │      ← "siempre". No es condicional                             │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ⑳  "CORRÉ LOS TESTS Y VERIFICÁ QUE PASAN TODOS"                │
  │                                                                 │
  │            ┌──────────────────────────────┐                     │
  │            │   ¿pasan todos?              │                     │
  │            └────┬────────────────────┬────┘                     │
  │                 │ no                 │ sí                       │
  │                 ▼                    │                          │
  │      "revisá el fallo                │                          │
  │       y corregilo"                   │                          │
  │                 │                    │                          │
  │                 └──── vuelve ────────┤                          │
  │                    a correrlos       │                          │
  │                                      │                          │
  │   ⚠ SI FALLA VARIAS VECES, ENTRO YO  │                          │
  │      ┌───────────────────────────┐   │                          │
  │      │ miro cuál es el problema  │   │                          │
  │      └───┬───────────────────┬───┘   │                          │
  │          ▼                   ▼       │                          │
  │   le doy indicaciones   SUBO EL MODELO                          │
  │                         a uno mejor  │                          │
  │                              │       │                          │
  │                    "la recomendación │                          │
  │                     del ⑯ no alcanzó"│                          │
  │                              └───────┤                          │
  └──────────────────────────────────────┼──────────────────────────┘
                                         ▼
  ┌─────────────────────────────────────────────────────────────────┐
  │  ㉑  REVISIÓN PUNTA A PUNTA — con un MODELO MÁS GRANDE          │
  │                                                                 │
  │      recorre toda la cadena:                                    │
  │                                                                 │
  │        us-#  →  spec  →  diseño  →  código  →  tests que pasan  │
  │                                                                 │
  │      la pregunta: ¿la implementación SATISFACE                  │
  │                    la feature planteada?                        │
  │                                                                 │
  │      ← SEGUNDO TRASPASO, y a otro actor distinto del que        │
  │        implementó: acá sí hay ojos frescos                      │
  │      ← el modelo se elige GRANDE, no "el que corresponda"       │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌═════════════════════════════════════════════════════════════════┐
  ║  ㉒  ACÁ SE CORTA LO QUE SÉ                                     ║
  ║     Ya está revisada… ¿y después qué?                           ║
  └═════════════════════════════════════════════════════════════════┘
```

## Lo que queda en el repo

```
proyecto/                 ← siempre con git, desde el arranque
  .docs/
    brief.md              ⑤+⑥  con el veredicto "como leyenda"
    prd.md                ⑦
    constitucion.md       ⑧    arquitectura + stack
    backlog/
      backlog.md          ⑨    índice
      us-1.md  us-2.md …  ⑨    una por historia, con ID
    roadmap.md            ⑩
    <modelos>.yaml        ⑯    catálogo de modelos de IA · NO lo genera la IA:
                               es tuyo, y es el insumo de la única decisión
                               de recursos del flujo
```

*(los nombres de archivo son mi suposición de formato; lo relevado es la **estructura**:
`.docs/`, la carpeta `backlog/` con índice y `us-#`)*

### Las tres reglas del paso ⑥

1. **El brief se sella acá, no antes.** Los pasos ② a ⑤ lo arman; ⑥ lo cierra.
2. **La IA recomienda, vos sellás.** Nunca al revés. La IA **no tiene poder de veto**: un
   "no lo hagas" es una objeción, no un freno.
3. **"No lo hagas" no tira el trabajo.** El brief queda igual, y ahora dice por qué NO lo
   hiciste — que es justo lo que hoy se pierde.

### Esto es un CAMBIO a tu flujo, no una relectura

Hoy el documento se escribe **después** de decidir: *"una vez que se define que sí vamos a
implementar creamos lo que sería un brief o prd"*.

O sea que hoy, en rigor, **no hay brief**: hay un PRD, y la decisión pasa en tu cabeza y en
un chat que después se pierde. **El paso ⑥ hoy no deja nada.**

Mover el documento antes de la decisión es lo que le da a ⑥ algo que sellar.

---

## Lo que sé con certeza, porque lo dijiste vos

| # | Paso | Tus palabras |
|---|---|---|
| ② | pinponeo con IA | *"abro claude modelo opus effort xhigh, tiro mi idea para comenzar a darle forma"* |
| ③ | búsqueda | *"con perplexity busco si ya existe algo similar"* |
| ④ | revisión del repo ajeno | *"le paso el repo a claude para que lo revise, ver si me sirve algo o qué diferencia tiene mi idea con lo ya creado"* |
| ⑤ | consenso | *"hasta llegar a un consenso de lo que quiero crear y porque deseo crearlo"* |
| ⑥ | sello yo | *"la ia recomienda pero sello yo"* · *"a veces hay ideas que me las rechaza pero lo mismo quiero hacerlas"* |
| ⑦ | el documento | *"una vez que se define que si vamos a implementar creamos lo que seria un brief o prd"* — hoy va después de decidir; se mueve antes (ver arriba) |

**Dos cosas que se ven solas mirando el dibujo:**

- **Del ② al ⑤ es un bucle, no una fila.** Volvés a buscar, volvés a pinponear. No es
  "paso 2, paso 3, paso 4" — es dar vueltas hasta que cierra.
- **El ⑥ es el único lugar donde hay una decisión de verdad**, y es tuya, y a veces va en
  contra de lo que la IA dice.

### El brief y el PRD sí son archivos, y viven en el chat

Tus palabras sobre el ⑦: *"le digo a la ia **quien ya tiene el brief (lo tiene entre las
conversaciones)** que arme un prd"*.

Aclarado después, y corrige una lectura equivocada:

> *"si hay un artefacto pero vive en el chat, el llm genera un md que está disponible dentro
> de su contexto y también para bajarlo, lo mismo ocurre con el prd"*

Lo que queda registrado:

- **El brief es un `.md` generado**, no texto de conversación suelto. Lo mismo el PRD.
- **Vive en dos lados a la vez:** dentro del contexto del LLM (por eso el PRD puede
  construirse encima sin volver a explicar nada) **y** disponible para bajar.
- **El brief y el PRD se generan en el mismo hilo**, uno encima del otro.

> **Corrección.** Una versión anterior de este documento decía que *"no hay artefacto, la
> continuidad es la conversación"*, y de ahí concluía que el sello del ⑥ no tenía sobre qué
> caer. **La premisa era falsa: el artefacto existe.**

### Y sí sobreviven al chat: van al repo, versionados

> *"queda en el repositorio del proyecto, siempre se inicia un git para mantener el
> versionado de todos los proyectos que inicio, suelo guardarlo en una carpeta llamada
> `.docs`"*

| Hecho | Detalle |
|---|---|
| Los `.md` se bajan | no quedan sólo en el chat |
| Van al **repo del proyecto** | carpeta `.docs/` |
| **Siempre hay git** | lo inicializás en todo proyecto que arrancás, para versionar |

**Con esto se cae también la segunda mitad de mi lectura anterior.** Yo venía asumiendo que
el brief y el PRD se perdían. **No se pierden: quedan como archivos versionados en git.**

**Y el veredicto del ⑥ también queda escrito.** Preguntado si el "hacelo / pivoteá / no lo
hagas" y sus razones sobrevivían: *"Sí quedan como leyenda"*.

> Queda pendiente saber **con qué forma** — "leyenda" puede ser una nota al pie, un
> encabezado o un párrafo suelto. No cambia el hecho de que está escrito; sí cambiaría qué
> tan fácil es encontrarlo después. No lo asumo.

### El patrón, ya con tres casos

| Paso | Qué hace la IA | Quién decide | Salidas |
|---|---|---|---|
| ⑥ la decisión | recomienda con razones | **vos** | hacelo / pivoteá / no lo hagas |
| ⑧ la constitución | pregunta y sugiere | **vos** | — |
| ⑰ la planificación | produce todo el plan | **vos** | cambios / otra feature / implementar |

**En los tres puntos donde hay que elegir, la IA propone y la última palabra es tuya.**

Y se aclara algo que venía anotado como duda: **⑨, ⑩ y ⑫–⑯ no tienen decisión propia
porque la decisión está al final del bloque, no en cada paso.** No es que falte — está
agrupada. Revisás una vez, sobre todo junto.

### El ⑱ es el primer traspaso, y cambia lo que los artefactos tienen que aguantar

Del ① al ⑰ **es una sola conversación**: cada paso se apoya en el contexto del anterior sin
volver a explicar nada. El ⑱ rompe eso — el plan se le pasa a **otro modelo**.

Consecuencia directa, y es la primera vez que aparece en el flujo: **la propuesta tiene que
bastarse sola.** Mientras todo pasaba en un hilo, un documento incompleto se salvaba con el
contexto de la charla. Acá no hay charla previa que lo salve.

### El ⑲ tiene tres tiempos, no dos

*"siempre creando primero los test **y los ejecute, para comprobar que fallan**, y luego
comience a implementar"*.

Lo habitual sería "escribí los tests, después implementá". Vos pedís algo más:

| | |
|---|---|
| 1 | crear los tests |
| 2 | **ejecutarlos** |
| 3 | **ver que fallan** |
| 4 | implementar |

**El 2 y el 3 son un paso propio.** No alcanza con que los tests existan: hay que verlos
fallar. Y lo pedís con *"siempre"*, no como recomendación.

### El ciclo de tests aparece dos veces, con signo opuesto

| Momento | Qué se busca | Quién lo hace |
|---|---|---|
| ⑲ antes de implementar | que los tests **fallen** | el modelo que va a implementar |
| ⑳ después de implementar | que los tests **pasen** | el mismo modelo |

Tres cosas registradas, sin conclusiones:

- **El mismo actor escribe el código, corre los tests y reporta el resultado.** Es la
  topología del paso, no un juicio sobre ella.
- **El bucle sí tiene salida, y sos vos.** *"Si fallan varias veces verifico cuál es el
  problema y le doy indicaciones, o si no elevo el modelo a uno mejor."* Corrige una
  observación anterior de este documento que decía que era el único bucle donde no
  aparecías: **aparecés, pero recién después de varios intentos fallidos.**

### El fracaso repetido es información sobre la estimación, no sólo sobre el código

Ésta es la salida que no esperaba:

> *"elevo el modelo a uno mejor **porque significa que la recomendación no fue suficiente**"*

Cuando el modelo no puede, no lo leés como "el código está mal": lo leés como **"la
complejidad estaba mal estimada en el ⑯"**. Y actuás sobre el recurso, no sobre el trabajo.

Eso convierte la cadena del ⑯ en un lazo cerrado:

```
   complejidad estimada  →  modelo elegido  →  ¿pudo?
          ▲                                      │
          └────── no: la estimación era baja ────┘
```

**La recomendación de modelo es una hipótesis, y la implementación es su prueba.** Es el
único lugar del flujo donde una decisión anterior se corrige con evidencia de ejecución.

### El ㉑ hace una pregunta distinta de la del ⑳

*"cuando los tests pasan uso un modelo más grande y le pido que haga una cobertura e2e,
desde la historia de usuario hasta los tests que pasan, para que verifique si la
implementación satisface la feature planteada"*.

| | Pregunta | Quién |
|---|---|---|
| ⑳ | ¿el código hace lo que los tests dicen? | **el mismo que implementó** |
| ㉑ | ¿lo implementado **satisface la historia**? | **otro modelo, más grande** |

Son preguntas distintas: que los tests pasen es un hecho sobre el código; que la
implementación cumpla la historia es otra cosa, y ningún test verde la contesta.

**Y acá sí hay ojos frescos.** En el ⑳ el que implementó se verifica a sí mismo; en el ㉑
interviene un actor que no escribió el código.

### Dos momentos de elegir modelo, con lógicas opuestas

| Momento | Cómo se elige | Criterio |
|---|---|---|
| ⑯ para implementar | **el que corresponda** | ajustado a la complejidad estimada |
| ㉑ para revisar | **uno grande** | sin escala: siempre el mejor |

Para construir, el recurso se ajusta al trabajo. Para revisar, no se ajusta nada.

> **Nota de vocabulario.** *"Cobertura e2e"* acá **no** significa tests end-to-end. Significa
> **recorrer la cadena entera de artefactos** —de la historia de usuario hasta los tests— y
> ver si se sostiene. Conviene no confundir los dos sentidos más adelante.

**Es la primera vez que la cadena de artefactos se USA en vez de producirse.** Todo lo
anterior los fabricaba; el ㉑ los recorre.

### El ⑰ revela que el flujo no es de a una feature

*"puedo solicitar cambios, continuar definiendo otras features, o comenzar la fase de
implementación"*.

Esa segunda salida cambia la forma del flujo: **se pueden planificar varias features antes
de implementar ninguna.** No es una tubería de a una — la planificación se puede acumular.

### El ⑨ cambia de naturaleza, y no tiene punto de decisión declarado

*"luego con la constitución armada le pido las historias de usuario para convertirlas en mi
backlog"*.

Dos observaciones, sin conclusiones:

- **Es la primera salida que no es para leer, es para trabajar.** Brief, PRD y constitución
  son documentos que se consultan. El backlog es una **cola**: se consume.
- **Es el primer paso donde no mencionaste un "yo decido".** En ⑥ y ⑧ lo dijiste explícito.
  Acá pedís y recibís. **No concluyo que no exista** — puede que sea tan obvio que no valga
  mencionarlo, o que revises las historias antes de aceptarlas. Queda como pregunta.
- **Es la primera vez que aparecen identificadores.** *"una carpeta backlog con un
  backlog.md como índice y los `us-#`"*. Todo lo anterior eran documentos de prosa; acá hay
  **unidades direccionables**, cada una con su archivo y su número.
- **El patrón índice + ítems aparece por primera vez:** un archivo que lista, N archivos con
  el contenido.

### El ⑩ no agrega material, agrega orden

*"el roadmap las ordena y agrupa para darle la prioridad de qué debería hacerse antes"*.

- **Es el único artefacto derivado del anterior sin contenido propio.** Todos los demás
  agregan algo: el PRD agrega detalle al brief, la constitución agrega el cómo técnico, las
  historias parten el PRD. El roadmap **no agrega nada** — reordena lo que ya está.
- **Es prioridad, no calendario.** Dijiste *"qué debería hacerse antes"*, no *"para cuándo"*.
  Es secuencia.
- **Queda por saber si el roadmap referencia los `us-#` o los copia.** Existiendo los IDs, lo
  natural sería referenciar — pero no lo asumo.
- **Segundo paso seguido sin "yo decido" declarado.** ⑨ y ⑩ los pedís y los recibís. Lo
  registro como acumulación de la misma pregunta abierta del ⑨, no como conclusión.

### Alerta de método del tramo ⑫–⑯ — resuelta

Este tramo estaba descrito con vocabulario de SpecForge (*propuesta*, *spec*, *design*,
*task*, *wave*), y eso amenazaba con describir la herramienta en vez del trabajo.

**Resuelto: lo haría igual sin la herramienta.** *"Lo haría igual porque uso la IA para que
realice el trabajo."* El tramo se reescribió con las palabras del pedido real, y ahí quedó a
la vista lo que el vocabulario tapaba:

| El nombre de la herramienta decía | Lo que en realidad pedís |
|---|---|
| "planificá el wave" | **agrupá las tareas para que corran en paralelo** |
| "creá el spec y el design" | **escribí la especificación de la solución que elegiste, con su diseño** |
| "los tests" | **decime qué tests hay que crear ANTES de cualquier implementación** |

**El caso de "wave" es el más claro:** el nombre no dice nada, y el pedido real tiene un
motivo — paralelizar. Son lotes, no etapas.

### Corrección: el brainstorming va antes de la especificación

Yo había leído que primero se especifica el **qué** y después se baraja el **cómo**. **Es al
revés.**

*"Le digo que haga un brainstorming para que vea 3 posibles implementaciones y se quede con
la más óptima. Una vez que encontró la solución le digo que cree una especificación junto
con el diseño **de la misma**."*

**Y es coherente, no un desorden:** el *qué* ya estaba resuelto — viene del `us-#`. Lo que se
especifica acá es **la solución elegida**, no el requisito. Por eso especificación y diseño
van juntos: son dos vistas de la misma decisión técnica.

### Tres cosas más de este tramo

- **"3 implementaciones", con número.** No *"varias"*, no *"las que se te ocurran"*. Tres.
- **Los tests van primero, dicho explícito.** *"antes de cualquier implementación"*.
- **Aparece un artefacto nuevo que no estaba en el mapa: un `.yaml` de modelos en `.docs/`.**
  Es un catálogo propio, y es el insumo de la única decisión de recursos del flujo.

### Huecos pendientes, para volver después

Dijiste *"un prd completo sobre el **producto**"*. Eso describe el caso de **un producto
nuevo**. Falta saber qué pasa cuando lo que arranca no es un producto sino **una feature**
sobre algo que ya existe. No lo asumo — queda anotado para preguntarlo cuando toque.

### Un tipo de "por qué" que hay que dejar entrar

Cuando sellaste en contra de la recomendación, la razón fue esta, textual:

> *"lo mismo lo hice porque quería aprender x cosa y venía al pelo el proyecto"*

**Aprender algo es un motivo válido para construir**, y ningún marco de producto lo aceptaría
— no hay usuario, no hay dolor medido, no hay mercado. Pero es real, lo usaste varias veces,
y las decisiones salieron bien.

Queda anotado acá para que cualquier regla futura del estilo *"el porqué tiene que estar
anclado a algo observado"* **no lo rechace por accidente**.

---

## Lo que NO sé, y necesito de vos

Sin esto el flujo está cortado a la mitad. Son cuatro preguntas y ninguna necesita que
pienses en herramientas — sólo contame qué hacés.

**1. Después del brief, ¿qué pasa exactamente?**
No lo que una herramienta te pedía. Lo que hacés vos cuando estás solo. ¿Abrís el editor y
arrancás? ¿Escribís una lista de tareas? ¿Le pedís a la IA que lo parta en pedazos?

**2. ¿Cómo decidís que algo está listo?**
¿Mirás que pasen los tests? ¿Lo probás a mano? ¿Lo leés? ¿Todo junto?

**3. ¿En qué momento te cagás solo?**
O sea: ¿dónde te pasó que algo salió mal y dijiste "esto me lo tendría que haber avisado
algo"? Ese momento es más valioso que todo el resto del flujo.

**4. ¿Qué parte de esto te da fiaca hacer?**
La que postergás, la que hacés a desgano, la que salteás cuando estás apurado.

---

## Por qué las preguntas 3 y 4 son las importantes

Todo lo demás del flujo lo hacés bien y no necesita herramienta.

**Donde una herramienta sirve es en dos lugares nada más:** donde te cagás solo (③) y donde
te da fiaca (④). Lo primero necesita una baranda; lo segundo necesita que alguien lo haga
por vos.

Si un paso no es ninguna de las dos cosas, **no hace falta que la herramienta lo toque.**
