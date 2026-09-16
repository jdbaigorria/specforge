# Los nodos — el modelo de skills, redefinido

**Fecha:** 2026-09-15 · **Branch:** `claude/redefinicion-desde-cero-hks3hk`

> **Qué es esto.** El modelo de la capa de skills, rearmado de abajo hacia arriba. Es el
> documento chico contra el que se escribe cada skill nuevo — no un informe de estado (eso sigue
> siendo `el-mapa.md`) ni el traspaso (`traspaso-2026-09-13.md`).
>
> **Qué NO cambia.** El binario, las compuertas y la máquina de estados se quedan. Lo que se
> rehace es cómo están escritos los skills y quién posee cada método.
>
> **Regla de lectura.** Lo que dice "medido" tiene archivo o `grep` con fecha. Lo que es lectura
> mía está marcado. Lo que viene de afuera dice de dónde.

---

## 1. Nodo, compositor, primitivo

Un **nodo** es un compositor. Un **primitivo** es un método.

```
PRIMITIVO    sabe hacer UNA cosa. Sirve solo, en cualquier proyecto.
             No sabe que existe sf, ni un estado, ni un contrato.

COMPOSITOR   no sabe hacer nada por sí mismo.
             Sabe a quién llamar, en qué orden, y qué tiene que quedar escrito.
```

**El compositor está vacío de método a propósito.** Un método metido adentro de un compositor
queda encerrado ahí y no lo puede usar nadie más. Así fue como la vara terminó enterrada adentro
de `sf-plan` y la pregunta del mutante escrita tres veces (`primitivos.md` §3).

**La prueba:** si borrás un primitivo, el compositor se queda sin poder hacer ese paso. Si borrás
un primitivo y el compositor sigue igual, ese método estaba copiado adentro.

## 2. La regla que los separa — y no es el piso

`sfx-grilling` llama a cuatro skills. `sfp-scout` llama a uno. Los dos están bien, y por eso el
modelo de "dos pisos" no alcanza. Lo que separa a un primitivo de un compositor es **qué posee**:

| | puede llamar a otros | por qué |
|---|---|---|
| **posee un MÉTODO** | **sí** | cada llamada es una rama de SU método: `grilling` llama a `buscar` porque buscar es parte de entrevistar |
| **posee un RESULTADO** | sólo delega | no puede re-explicar el método: `scout` no enseña a entrevistar, llama y le agrega el veredicto |

**Y un compositor llama a uno por cada tramo que termina en un artefacto.** Si dos llamadas no
dejan nada escrito en el medio, era una sola llamada.

## 3. El molde del compositor

Sacado de leer `sfp-scout`, que es el único compositor probado del repo. Seis partes, y **ninguna
es método**:

```
1  EL ENCUADRE     qué hace y, sobre todo, qué NO hace
2  LA DELEGACIÓN   órdenes ejecutables. "Compose, never re-explain."
3  EL ARTEFACTO    qué archivo queda, y qué NO va adentro
4  LA CABECERA     los pocos campos que una compuerta lee
5  EL ACTA         lo que se imprime al parar, con los números crudos
6  LA CICATRIZ     qué se midió, qué se rompió, y por qué el listón está donde está
```

**La orden tiene que ser ejecutable.** `Call the Skill tool with "X"`, no "acá conviene investigar".
Medido el 2026-09-08: *un modelo débil lee prosa y no llama nada*. Es la diferencia entre las seis
órdenes ejecutables del repo y las decenas de menciones en prosa (`primitivos.md` §2.1).

## 4. El eje que faltaba: quién te necesita

Cada primitivo se clasifica por si necesita a Javier o no. **Es el eje que ordena la
optimización**, y es `comprobar ≠ decidir` aplicado a los skills en vez de al motor.

```
CORRE SOLO, y en paralelo           TE NECESITA
buscar afuera                       contestar las preguntas de la entrevista
leer el repo                        aprobar que se construya un prototipo
reproducir un error                 el veredicto final
ver qué cambió en el git
grep de si algo se usa
puntuar contra la vara
```

Tres consecuencias:

1. **Lo que corre solo, corre junto.** No hay motivo para hacerlo de a uno.
2. **Lo que corre solo puede ir con modelo barato.** Decidir no.
3. **Un compositor bien armado te interrumpe UNA vez**: cuando te muestra el acta.

## 5. Las tres entradas

Las tres son el mismo trabajo con distinta materia prima. **Lo único que cambia es el primer
paso.**

```
COMPOSITOR DE ENTRADA
   1  ENTENDER   un primitivo que maneja      ← LO ÚNICO distinto entre las tres
   2  DECIDIR    sfx-decidir                  ← igual en las tres
   3  EL ACTA    y la parada                  ← igual en las tres
```

| entrada | ENTENDER | decide | deja | skill |
|---|---|---|---|---|
| llego con una idea | `sfx-grilling` | ¿vale construirlo? | el brief con veredicto | `sfp-scout` (ya existe) |
| llego con una feature | `sfx-grilling`, arrancando en el repo | ¿entra? ¿cuánto proceso pide? | la historia con criterios | `sf-entrar-feature` |
| llego con un bug | `sfx-triage` | ¿cuál arreglo? | la causa + el test que falla | `sf-entrar-bug` |

**Las tres escriben lo mismo, además de su artefacto:**

```
· el veredicto         va / no va / no lo pude comprobar
· el motivo            por qué
· cuánto proceso pide  lo que viene después
```

El tercero hoy no lo escribe nadie — medido el 2026-09-13: `tipo: chico` existe en el motor, en
`sfx-mapa` y en el esqueleto del `us-#`, y **`sfp-backlog`, que es quien escribe el frontmatter, no
lo menciona nunca**. Por eso el flujo cobra peaje completo siempre.

### El proyecto que ya existe no es una cuarta entrada

Si llegás con un repo de tres años, lo que querés es meter una feature o arreglar un bug: entrás
por la ② o la ③. Lo que falta no es un nodo — es un **insumo**: esas entradas necesitan saber qué
es este producto, y hoy eso sólo puede venir de haber hecho todo el camino de la idea.

Lo resuelve `sfx-leer-repo`, que corre solo y una vez. Ver §7.

## 6. La vara

**La vara son los criterios con los que vas a decidir, escritos antes de ver las opciones.**

Una vara escrita después describe la opción que ya elegiste. Por eso el orden es la propiedad, y
sin esa propiedad la vara no vale nada.

**No lo tiene ninguno de los tres vecinos.** Se buscó `rubric`, `scorecard`, `criteria before` en
`obra/superpowers` y `mattpocock/skills` el 2026-09-13: no hay nada. Su `brainstorming` genera
opciones y los ADR de Matt registran la decisión después. **Es el único primitivo del rediseño que
no se puede copiar de nadie.**

### Dos tipos

| | cuándo | quién la escribe |
|---|---|---|
| **de catálogo** | el material se repite siempre igual | ya está escrita: se lee |
| **a medida** | el material es único | crece durante la conversación |

Tres de las cuatro varas son de catálogo (feature, bug, hallazgo). **La única a medida es la de la
idea** — y ésa no se escribe antes de la charla, porque no se puede: nace de ella.

### La regla de honestidad, afinada

No es *"antes de la conversación"*. Es:

> **Un criterio sirve para juzgar una opción sólo si se escribió antes de que esa opción estuviera
> sobre la mesa.**

Por eso cada criterio anota **en qué ronda nació**. Si el V-3 nació en la ronda 2 y la opción B
apareció en la 4, el V-3 la juzga honestamente. Si nació después, no. Y eso se comprueba mirando
el archivo, sin creerle a nadie.

### Y potencia la entrevista

La vara no le agrega trabajo al pinponeo: **es lo que ya dijiste, anotado.** Cuando en la ronda 2
decís *"esto tiene que poder probarse sin construirlo entero"*, eso es un criterio, y hoy se pierde
en el fondo de la entrevista.

Lo que devuelve *(lectura mía, no medido)*:

- **poda el árbol** — una pregunta que no toca ningún criterio no vale la pena hacerla;
- **muestra el desacuerdo temprano** — en la ronda 2 y no en la 6;
- **le da un corte mejor** — hoy la entrevista termina cuando la frontera está vacía; con vara
  termina cuando cada criterio tiene con qué contestarse. Es un corte sobre la decisión, no sobre
  el árbol.

El disparador es el mismo gesto que `sfx-grilling` ya usa con `sfx-vocabulario`:

```
una palabra significa dos cosas  →  sfx-vocabulario
aparece un criterio              →  sfx-vara
```

## 7. El insumo: leer el repo

`sfx-leer-repo` corre solo, una vez, y deja `.docs/repo/`. Lo leen las entradas ② y ③.

La forma es de `commandcode.ai/docs/taste` (leído el 2026-09-15): un archivo por tema en vez de un
documento gigante, la unidad se cuenta, y un validador que revisa el formato.

**Lo que NO se copia, y es el fondo del asunto:** Taste aprende de tu conducta en segundo plano y
le pone a cada aprendizaje un número de confianza entre 0 y 1. Un número de confianza que se pone
el mismo modelo que escribió la regla **no se puede comprobar** — y eso es exactamente lo que ya
nos mordió el 2026-09-05, cuando un modelo selló un veredicto con total seguridad y cero fuentes.

```
ellos      ¿cuánta confianza tenés?    0.8         no se puede verificar
nosotros   ¿CÓMO lo sabés?             file:line   se puede ir a mirar
```

Así que se copia la forma y se cambia la fuente: **la fuente es el código, y el respaldo es la
procedencia.**

**Y se pone viejo.** Es el problema que ellos resuelven aprendiendo siempre y nosotros no. La
salida: cada regla anota de qué commit salió, y si esa parte del código cambió mucho desde
entonces, queda marcada para volver a mirarla. *(Idea mía, no de ellos.)*

## 8. La cuenta de reuso

| pieza | la usan | estado |
|---|---|---|
| `sfx-decidir` | las 3 entradas **+ recibir una revisión** = 4 | nuevo, compositor |
| `sfx-vara` | las 4 decisiones | nuevo, primitivo |
| `sfx-grilling` | idea + feature = 2 | existe, se le agrega la vara |
| `sfx-buscar` | adentro de `grilling` y de `diagnosticar` | existe |
| `sfx-triage` | bug, un test que se pone en rojo, un hallazgo que resulta ser un bug | **existe** — se le agregaron 3 cosas, ver §11 |
| `sfx-criterio` | `sfp-backlog` · `sf-entrar-feature` · `sf-plan` ⑭⑮ · `sf-check` ㉑ = 4 | nuevo, extraído de `sfp-backlog` |
| `sfx-leer-repo` | entradas ② y ③ en un repo que ya existe | nuevo, primitivo |

**Dos primitivos nuevos, uno extraído, uno afilado, y un compositor. El resto es reuso.**

Y tres cosas dejan de ser un skill aparte:

- **recibir una revisión** = `sfx-decidir` con un hallazgo adelante — tapa el hueco que abrió el
  `f828ec2` ③: la parada ME TRABÉ EN LA REVISIÓN existe, `sf dismiss` existe, y el criterio para
  decidir no;
- **triage** = `sfx-decidir` con un pedido adelante;
- **el sello del brief** deja de ser un caso especial.

**Agregar un caso de uso nuevo no agrega un skill.** Eso ataca la ceremonia por la raíz.

## 9. De dónde salió cada cosa

| qué | de dónde | fecha |
|---|---|---|
| el molde del compositor | `sfp-scout`, leído | 2026-09-15 |
| las dos caras del test que falta + la trampa del espejo | `superpowers/test-driven-development` | 2026-09-13 |
| la vara de catálogo del hallazgo | `superpowers/receiving-code-review`, sus seis razones | 2026-09-15 |
| las 3 cosas que le faltaban a `sfx-triage` | `superpowers/systematic-debugging` | 2026-09-15 |
| la forma de `.docs/repo/` y el validador | `commandcode.ai/docs/taste` | 2026-09-15 |
| la vara | **de nadie** — no está en ninguno de los tres | — |

Y dos cosas que ellos tienen y confirman lo nuestro sin habernos leído:

- `systematic-debugging` corta a los **tres** intentos fallidos y manda a hablar con el humano.
  Es nuestro `TopeIntentos = 3` inventado aparte.
- `receiving-code-review` dice *"si no lo podés verificar, decilo"*. Es la tercera salida —
  el `?` de `FUNDAMENTOS.md` ③ — llegando desde otro lado.

## 10. Lo que se corrigió el 16-09, y por qué queda escrito

Cuatro cosas de la primera pasada estaban mal. Quedan acá porque el error es fácil de repetir.

| se dijo el 15-09 | la verdad | por qué se falló |
|---|---|---|
| hacía falta un `sfx-diagnosticar` nuevo | **`sfx-triage` ya era el método de diagnóstico** — ley de hierro, rastreo hacia atrás, diff contra lo que funciona, hipótesis con evidencia, límite duro, test antes del arreglo | se leyó `vecinos-metodos.md` §4.1 sin abrir el archivo del skill |
| `sfx-triage` se disuelve en `sfx-decidir` | **no**: es el ENTENDER del nodo del bug. Triage diagnostica, decidir elige | mismo origen: se lo creyó clasificador |
| faltaba el método del bug | **faltaba la puerta**, más tres detalles | ídem |
| `sfp-backlog` era un compositor sano | tenía **un primitivo atrapado adentro** — el criterio, con tres consumidores y ningún dueño (`primitivos.md` §3.3) | no se lo había mirado con la regla nueva |

**El patrón del error es uno solo: leer un documento del repo sin cruzarlo con el código.** El
traspaso del 13-09 §3 ya lo había anotado como la forma de sus cuatro correcciones anteriores.
Volvió a pasar nueve días después.

Lo que se hizo: se borró `sfx-diagnosticar`; a `sfx-triage` se le agregaron las tres cosas que
aportaba (instrumentar las costuras, el corte de 3 **arreglos** fallidos, y entregar candidatos
sin recomendar) y se le sacó el Step 7 que recomendaba el alcance; se extrajo `sfx-criterio` de
`sfp-backlog`; y se corrigió `vecinos-metodos.md` §4.1, que es de donde salió todo.

> **Y `sfx-criterio` resolvió gratis el choque que quedaba abierto.** `sfp-backlog` y
> `sf-entrar-feature` competían porque los dos escribían criterios. Ahora los dos llaman al mismo
> primitivo y cada uno se queda con lo suyo: `sfp-backlog` sólo el camino del producto (PRD →
> historias), `sf-entrar-feature` sólo lo que llega después.

## 11. Lo que queda abierto

1. Si `sfx-decidir` aguanta las cuatro materias en un solo skill o hay que partirlo. Se prueba
   corriéndolo, no discutiéndolo — y si se parte, los primitivos no se tocan.
2. Si `sf-plan` ⑭⑮ y `sf-check` ㉑ tienen que llamar a `sfx-criterio` cuando un criterio se les
   resiste, o les alcanza con consumirlo. Hoy no lo llaman: sólo lo hacen los dos que ESCRIBEN
   criterios.
2. El validador de `.docs/repo/` (el `lint` de ellos). Está decidido que va; no está escrito.
3. **Resuelto el 16-09** — ver §10. `sfx-triage` NO se disuelve: es el ENTENDER del nodo del bug.
   `sfx-explain` quedó cerrado al modelo (`disable-model-invocation: true`): es puerta de usuario,
   como el `teach` de Matt.
4. Nada de esto está cableado al motor. Es la capa de skills sola, a propósito.
