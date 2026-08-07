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
  │      se lo pido a la IA                                         │
  └────────────────────────────┬────────────────────────────────────┘
                               ▼
  ┌═════════════════════════════════════════════════════════════════┐
  ║  ⑪  ACÁ SE CORTA LO QUE SÉ                                      ║
  ║     Ya hay roadmap… ¿y después qué?                             ║
  ║     Lo único que tengo es "a veces el PRD lo paso directo a     ║
  ║     propose o a un generador de historias de usuario".          ║
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

### Un patrón que ya aparece dos veces

| Paso | Qué hace la IA | Quién decide |
|---|---|---|
| ⑥ la decisión | recomienda con razones | **vos** |
| ⑧ la constitución | pregunta y sugiere | **vos** |

*"me pregunta qué lenguaje voy a usar y me sugiere alguno pero como siempre yo tengo la
última decisión"*. Se registra como observación, no como conclusión: **en los dos puntos
donde hay que elegir, la IA propone y la última palabra es tuya.**

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
