# El arnés ciego — lo que vio la primera corrida real

**Fecha:** 2026-09-05 · **Branch:** `refundation` · **Commit:** `4d7cc28`

> **Estado: UN ARREGLO HECHO, SEIS PROPUESTOS.** Sale de la primera corrida de T1 con dos modelos
> y la misma idea semilla. **Todo lo de §1, §2, §6 y §8 está medido** — carpetas, registros y
> salidas reales, no impresiones. Lo que es deducción está marcado como deducción.
>
> **Revisión del 2026-09-07:** §5② se amplió. Dejó de ser *"el ⑥ muestra la evidencia"* y pasó a ser
> **el contrato del acta, para toda parada** — con la distinción **comprobar ≠ decidir** que la
> explica. Lo que se agrega sobre el código está verificado leyéndolo, no supuesto.

Este documento continúa a [`por-tramos.md`](por-tramos.md) —que pidió esta corrida y escribió el
método— y a [`vecinos.md`](vecinos.md). Lo que agrega es **el resultado**: la primera vez que un
artefacto de esta máquina lo escribió un modelo y no una mano.

**Cómo está armado.** §1–§3 son el parte de la corrida y terminan en la regla que ordena todo lo
demás. §4 es lo único ya construido. §5 son las cuatro tareas, y **§6 y §7 son las respuestas a las
tres preguntas que Javier hizo el mismo día** —cómo se subdivide la fase, qué herramientas hacen
falta, qué artefactos faltan—, que resultaron ser **la misma respuesta vista de tres lados**:

```
subdividir la fase        el ② es el único paso con entrada, artefacto y compuerta propias
darle herramientas        el ② es el único que las necesita
generar artefactos        el ② es el que produce el que falta

                          → son tres caras de sacar el ② afuera
```

---

## 1. Qué se corrió, y qué salió

**El banco:** `~/projects/workspace/personal/sf-banco/`, fuera de este repo, dos copias idénticas
(mismo `estado.json`, mismo commit, registro vacío — verificado antes de arrancar).

**La idea semilla**, la misma palabra por palabra en las dos: *un CLI que lee los transcripts
locales de Claude Code y dice cuánto se gastó por sesión, proyecto y día*. Elegida **a propósito
porque es probable que ya exista** — un scout honesto tiene que poder decir "esto ya está hecho".

### El resultado

| | A — Claude Code | B — opencode / nemotron |
|---|---|---|
| **Veredicto** | `no-lo-hagas` | `hacelo` |
| Afirmaciones con fuente (`retrieved`) | 17 | 1 |
| Afirmaciones de memoria (`model-prior`) | 1 | 8 |
| **Links citados** | **13** | **0** |

**Veredictos opuestos sobre la misma idea, con el mismo texto de entrada.**

### Y la máquina aceptó los dos

`compuerta.Brief` (`compuerta.go:111`) sólo comprobaba que el veredicto fuera uno de los tres
válidos. Nada más. Así que:

```
A:  22:30:07  next  →  "El brief se selló como no-lo-hagas. No hay nada que seguir."   (exit 3)
B:  22:31:26  next  →  prd            ⏵ sfp-po · nemotron · subagente
    22:35:34  next  →  constitucion
```

> **A frenó el producto. B siguió y escribió un PRD de 74 líneas para algo que A concluyó que no
> había que construir.**

### El detalle que ordena el arreglo entero: **B no mintió**

Marcó cada afirmación como `` `model-prior — unverified` ``, que es **exactamente lo que el skill le
pide**. Y las señales de demanda también: *"(todas `model-prior — unverified`)"*.

**El skill hizo su trabajo. El modelo fue honesto. La que exigía de menos era la compuerta.**

### Dos cosas que salieron a favor de la máquina, y también son dato

**① La compuerta atrapó a un paso que dijo estar listo sin estarlo.** Del registro de B:

```
22:33:51  done  prd  ✗  falta .docs/prd.md      (salida 2)
22:35:17  done  prd  ✓
```

Es [`headless.md`](headless.md) §7 en vivo: *"la carta del arnés dice éxito aunque no haya hecho
nada"*. Lo único que lo frenó fue que quien decide es `sf done` y no el arnés.

**② El PRD de nemotron no fue basura.** 74 líneas, actores, capacidades. Y **arrastra la debilidad
aguas abajo** por su cuenta:

```
C-4 — Agregar costos por sesión  ⚠ apoyada en evidencia sin verificar (brief: model-prior)
```

`compuerta.PRD` sigue siendo un `os.Stat` y nada más (`compuerta.go:131`, y su comentario dice *"y
tenga cuerpo"*, que es falso). **Que no se haya explotado esta vez no dice que esté bien**: dice que
hoy no fue el eslabón que se rompió.

---

## 2. La causa raíz — el arnés estaba ciego

**Nemotron no falló por flojo. Falló porque no tenía con qué buscar.**

Los tres hechos, verificados:

| | |
|---|---|
| ¿el repo de SpecForge tiene `.mcp.json`? | **sí** — `tavily`, `github`, `fetch`, `context7` |
| ¿lo tuvieron las corridas A y B? | **no.** El archivo no existe en ninguna de las dos |
| ¿`sf install` hace algo con MCPs? | **nada.** Ni una mención en `andamio/` ni en `arranque/` |

Y el renglón que lo cierra, de `sfp-scout/references/tooling.md`, textual:

> *"SpecForge ships an opinionated default set in the **plugin's** `.mcp.json`. They start
> automatically **when the plugin is enabled**."*

**Los skills se instalaron por symlink, no como plugin.** Ese archivo nunca se cargó.

### Lo peor del hallazgo

**A encontró 13 links por accidente.** El Claude Code de Javier tiene `tavily`, `github` y `fetch`
declarados a nivel global, de antes y por otra razón. **SpecForge no le dio nada a ninguno de los
dos.** Uno tenía internet por su cuenta y el otro no.

> Le pedimos a los dos que investigaran, a uno le dimos con qué y al otro no, y después nos
> sorprendió que uno inventara.

---

## 3. La regla que ordena todo lo que sigue

El skill **ya tenía** la compuerta que hacía falta. Está en el Step 2 de `sfp-scout`, textual:

> *"If the research MCPs are unavailable, do **not** silently invent competitors from training data
> — that is the false-validation trap. Either (a) name the MCPs to enable and stop, or (b) with
> explicit consent run a **degraded pass**."*

**B no hizo ninguna de las dos.** Inventó, selló `hacelo`, y siguió.

Es la misma familia que el bug del ⑧: una instrucción que un modelo fuerte respeta y uno débil
atropella. Y de ahí sale la frase que vale para el proyecto entero:

```
TODA REGLA QUE VIVE SÓLO EN EL SKILL ES UNA SUGERENCIA.
SÓLO LA COMPUERTA OBLIGA.
```

Y ya mordió **tres veces**, no una:

| Regla | Dónde vive | Qué pasó |
|---|---|---|
| *"el ⑧ te entrevista"* | el skill | nemotron preguntó al aire; los de Anthropic se lo saltearon |
| *"sin MCPs no inventes"* | el skill | nemotron inventó |
| *"imprimí el menú del ⑥ con la evidencia"* | el skill | **§5②** — `sf` no lo imprime, y hoy Javier decide a ciegas |

**Ninguna de las tres es un bug del binario. Las tres son reglas escritas en el lugar donde no
obligan.**

---

## 4. Lo que ya se arregló — `4d7cc28`

**Un veredicto sin una sola fuente citada no sella.**

### Por qué se cuentan links y no la palabra `retrieved`

`retrieved` es una palabra que el modelo tipea; puede tipearla sin haber recuperado nada. Un link
tampoco prueba que la fuente exista —se puede alucinar una URL— **pero cero links sí prueba algo**:
un modelo que inventa competidores de memoria no tiene links que escribir, porque no los tiene.

Es un hecho sobre bytes, no una opinión sobre calidad. Por eso lo puede afirmar una compuerta (R3).

### Por qué el umbral es uno, y no es un número mágico

**Cero es una categoría** —"no investigó"—. Cualquier número mayor es un juicio sobre **cuánto**
alcanza, y el juicio es de Javier. Uno es el borde entre nada y algo, que es la única línea que una
compuerta puede trazar sin opinar.

### Aplica a los tres veredictos

No sólo a `hacelo`. Lo dice el propio skill: *"el ⑥ es el único punto donde una alucinación cuesta
el producto entero: sellar `no-lo-hagas` porque existe algo que en realidad no existe."* **Matar una
buena idea por una alucinación es el mismo fallo que construir una mala.**

### La salida, que ya estaba diseñada

`evidencia: baja` en el frontmatter. No se inventó: es la *pasada degradada* que el skill ya
contemplaba. **Pasa, pero avisa** — y hay que escribirla a propósito, que es justo lo que se quiere
que cueste.

### Verificado contra los briefs reales de la corrida

```
A — no-lo-hagas · 13 links  →  ✓ pasa       (exit 0)
B — hacelo      ·  0 links  →  ✗ se frena   (exit 2)
```

Y el mensaje de B trae **una sola acción legal siguiente**.

### Lo que NO compra

No prueba que las fuentes sean reales, ni que el modelo las haya visitado, ni que la investigación
sea buena. **Cierra el agujero de CERO evidencia. El de POCA evidencia es juicio, y el juicio es del
⑥.**

> **Y hay un límite duro que conviene aceptar ya:** la compuerta **no puede salir a internet** a
> verificar que los links existan. La regla de cero red y cero LLM en el camino de enforcement se
> respeta. Así que este chequeo nunca va a distinguir un link real de uno inventado — sólo contar.
> Por eso §5① y §5③ no son opcionales: **es ahí donde de verdad se gana.**

---

## 5. Los cuatro que faltan

### ① `sf doctor` ve las herramientas de investigación

**El dolor.** Hoy alguien parado exactamente donde estábamos nosotros **no recibe ni un aviso**.

**Lo que hay hoy** (verificado corriendo `sf doctor`): mira el binario, el harness, los modelos por
perfil y los 9 skills de la máquina. **De herramientas de investigación no dice una palabra.**

**El contrato.**

1. `doctor` gana un bloque `herramientas`, al lado de `skills`.
2. Lee qué declara el arnés en el que está parado — la misma disciplina de `sf models`: **lo que el
   arnés no da, NO ESTÁ**.
3. Compara contra la lista que `sfp-scout/references/tooling.md` ya nombra: `tavily`, `github`,
   `fetch`, `context7`.
4. **Avisa, no frena.** Un proyecto que nunca va a correr el ⑥ no necesita `tavily`.

```
herramientas   investigación
       ✓  fetch      (sin key)
       ✓  context7   (sin key)
       ✗  tavily     declarado, sin TAVILY_API_KEY
       ✗  github     no está en este arnés
   ⚠ el ⑥ va a correr en modo degradado: sin búsqueda web no hay panorama.
```

**El riesgo.** Que `sf` no pueda ver los MCPs de un arnés que no expone esa lista. Ahí **no se
adivina** (R3): se dice *"no lo puedo saber en este arnés"*, que es distinto de *"no están"*.

**Los tests.** Los tres desenlaces —está / no está / no se puede saber— distinguibles desde afuera,
igual que `sf models`.

---

### ② Toda parada entrega un **acta** — Javier decide con datos

> **Ampliado el 2026-09-07.** Nació como *"el ⑥ muestra la evidencia"*. La conversación con Javier
> lo generalizó: no es del ⑥ ni es de la evidencia, **es de toda parada**. El ⑥ queda como el
> primer caso, no como el caso.

**El dolor.** Esto es lo que la máquina le muestra hoy a Javier antes de que selle el producto
entero (salida real):

```
🛑 PARÁ. El brief está escrito y lo sellás vos (el ⑥).

   si aprobás corre:  ⑦ el PRD → ⑧ la constitución
   próxima parada:    el sello del ⑧
```

**Ni el veredicto. Ni las fuentes. Ni el diferenciador.**

El mensaje bueno —con `Proposed:`, `Evidence: retrieved N / model-prior M` y `Differentiator:`—
existe, **pero en el skill**. O sea que depende de que el modelo se acuerde de imprimirlo. **Es §3
otra vez.**

---

#### La distinción que ordena todo esto: **comprobar ≠ decidir**

Son dos trabajos, y v1 de SpecForge los hizo uno solo:

```
comprobar   trabajo mecánico   →  la máquina, SIEMPRE, en todos los pasos, sin pedir permiso
decidir     juicio             →  Javier, y SÓLO donde equivocarse sale caro
```

**El error de v1 no fue "el humano es la compuerta". Fue que el humano era la compuerta *y* el
inspector.** Tenía que leer, contar, verificar *y además* decidir. Cuando alguien hace las cuatro
cosas veinte veces por feature, la cuarta se convierte en apretar Enter sin mirar. Eso no es
control: es la ceremonia del control.

**Un check después de implementar, una auditoría, una validación de seguridad — ésos no son
decisiones.** Son cosas que se tienen que hacer. Pedir permiso para correrlas es gastar la atención
de Javier en el único lugar donde no cambia nada.

**Y el *dónde* ya está bien resuelto** (verificado en el código, no deducido):

| Paso | ¿Para? | Dónde vive |
|---|---|---|
| ⑥ brief | **sí** | `done.go` → `terminarProducto` |
| ⑦ PRD | no — pasa derecho | `done.go` |
| ⑧ constitución | **sí** | `done.go` |
| ⑨/⑩ backlog | **sí** (⏸ blanda: "visto") | `done.go` |
| ⑰ plan | **sí** | `paradas.go` → `Aprobar`, `case estado.Planificacion` |
| ⑱–⑳ lotes | no | `cerrarLote` |
| ㉑㉒ revisión | **no — decide sola**: cuenta hallazgos abiertos y vuelve a implementar | `done.go` → `cerrarRevision` |
| ㉓ cierre | **sí** (y recién ahí archiva, que es irreversible) | `paradas.go` → `Aprobar` |

**Lo que falta no es dónde parar. Es qué mostrar cuando se para.**

---

#### El hallazgo: la compuerta le habla al que trabaja, no al que firma

Está escrito en el código, arriba de `Resultado.Texto()`
(`sf/internal/compuerta/compuerta.go`):

> *Los ✓ no se listan: sólo importa lo que falta. Un veredicto que enumera todo lo que salió bien
> es ruido, y el que lo lee tiene que buscar la ✗ entre quince ✓.*

**Ese comentario tiene razón — para el subagente que tiene que arreglar algo. Y está equivocado —
para Javier parado en una parada.** Un solo output, dos lectores con necesidades opuestas.

Es la misma diferencia que un test: cuando corre en CI y falla querés la ✗ y nada más; cuando vas a
firmar el release querés ver los 340 verdes. **Misma corrida, dos informes.**

Y hoy el segundo informe **no se puede armar**: `Resultado` guarda `Fallas` y `Avisos`, y **los ✓ se
tiran**. Cuando llega el momento de mostrárselos a Javier, el dato ya no existe.

---

#### El contrato

**Tres bloques, y el tercero es el que hoy no existe en ninguna parte de la máquina.**

```
🛑 ⑥ brief — te toca a vos

COMPROBÉ              ✓ existe .docs/brief.md
                      ✓ veredicto válido: "hacelo"
                      ✓ trae sección de fuentes

MEDÍ                  1 retrieved · 8 model-prior · 0 links
(cuento, no juzgo)

NO PUEDO COMPROBAR    si 1 fuente alcanza para un "hacelo"
                      si las 8 afirmaciones del modelo son ciertas

  sf approve   ·   sf reject "por qué"
```

1. **`COMPROBÉ`** son los ✓ que hoy se tiran. Se guardan en `Resultado` y se imprimen sólo acá.
2. **`MEDÍ`** son números crudos del artefacto. **Ninguno se interpreta** (R1): si el brief no trae
   diferenciador, se dice que no lo trae — no se resume el brief.
3. **`NO PUEDO COMPROBAR`** es la frontera: dónde termina lo que la máquina sabe y arranca lo que
   sólo puede saber Javier. **Es el bloque que le da sentido a la parada.** Sin él, una parada donde
   todo pasó parece un trámite; con él, es una pregunta concreta.

**Y esto solo hubiera atajado el caso B de §1.** La compuerta no lo tenía que frenar —juzgar si una
fuente alcanza es criterio, y R3 dice que una compuerta frena sobre hechos—, pero podía **ponerlo
adelante en vez de esconderlo**. Javier leía `0 links · veredicto hacelo` y rechazaba en dos
segundos. **La compuerta no reemplaza al humano: le da con qué.**

---

#### Los botones ya están, y son dos

No hace falta un tercer verbo para "rehacer": **`sf reject` YA es rehacer.**

```
sf approve            sella y avanza                       paradas.go → Aprobar
sf reject "motivo"    no sella, guarda el motivo,          paradas.go → Rechazar
                      y el motivo viaja en el sobre de
                      `sf context` al que rehace de cero
```

El motivo es obligatorio y no es decoración: el subagente que rehace **arranca en frío**, y sin el
motivo vuelve a proponer lo mismo.

---

#### Un choque de nombres que hay que resolver antes de escribir código

**`el sobre` ya está tomado**: es lo que `sf context` le sirve al **subagente**
(`superficie-sf.md` H2, H13). Lo de acá va para el lado contrario.

```
el sobre   →  va al que TRABAJA   ·  lo sirve `sf context`
el acta    →  va al que FIRMA     ·  lo imprime la parada
```

Se propone **`el acta`**.

**Y `parte` tampoco sirve, aunque sea vocabulario del repo:** ya existe
`type Parte struct` en `internal/sobre/sobre.go:77` — *"una sección del sobre"*. Chocaría adentro
de la misma familia de conceptos. `acta` está libre: se verificó por palabra entera sobre todo el
Go y no aparece.

---

**Por qué sigue siendo de los baratos.** Los datos ya están: la compuerta abre el artefacto y ya
cuenta. Es guardar lo que ya se calculó y agregar una segunda salida al lado de `Texto()`. Toca un
solo paquete.

**El riesgo.** Que `sf` empiece a resumir el artefacto — no es su trabajo (R1). **Se imprimen campos
y ninguno se interpreta.** Y el segundo riesgo, más callado: que `NO PUEDO COMPROBAR` se llene de
todo lo imaginable y se vuelva ruido. **Va sólo lo que la compuerta rozó y no pudo cerrar.**

**Los tests.** Tres desenlaces distinguibles desde afuera: una parada con todo en verde imprime los
tres bloques; una compuerta que falla sigue imprimiendo sólo la ✗ (el lector es otro); y un
artefacto sin el campo medible dice *"no lo trae"* y no cero.

---

### ③ `sf install` le da al arnés lo que la máquina le va a pedir

**El dolor.** Es §2 entero. La máquina exige investigación y no reparte herramientas.

**Lo que hay hoy.** El `.mcp.json` existe en el repo y **sólo lo carga el camino del plugin**.
Cualquier instalación por symlink —que es la que `por-tramos.md` §8 **manda usar para probar**—
corre sin herramientas y sin decirlo.

**El contrato.**

1. `sf install` escribe la declaración de MCPs **en el formato del arnés que le toca**, con la misma
   estructura que ya usa para las skills y los agentes.
2. **Transporta, no opina** — es la regla del catálogo de modelos otra vez: `sf` no elige tu
   proveedor de búsqueda, copia lo que el default declara y te deja cambiarlo.
3. **No escribe secretos.** Se referencian variables (`${TAVILY_API_KEY}`), como ya hace el
   `.mcp.json` del repo.
4. Si el arnés no soporta MCPs, se dice y no se inventa un archivo que nadie lee.

**El riesgo, dicho de frente.** **Cada arnés declara MCPs distinto**, igual que declara modelos
distinto. Éste es el más caro de los cuatro y es donde la promesa multi-arnés se paga.

> **Y es la pelea que este repo ya ganó una vez.** El catálogo de modelos existe exactamente por
> esto. La conclusión es la misma: **`sf` transporta, no traduce.**

**El umbral que despertaría hacerlo bien y no a medias:** un segundo arnés donde el formato no se
parezca en nada al primero. Hasta entonces, dos formatos y una tabla.

---

### ④ El ①–⑤ se parte — un corte, no cinco

**El dolor.** `sfp-scout` hace **cinco pasos de un saque, en una sola cabeza, sin frenar en el
medio**. El único control está al final, en el ⑥. Si se degrada en el paso ②, **nadie se entera
hasta el final** — que es literalmente lo que pasó en B.

#### La regla que decide dónde cortar

> **Un paso merece ser un paso propio cuando tiene entrada propia, artefacto propio y compuerta
> propia.** Si le falta una de las tres, es conversación, y una conversación cortada en pedazos no
> gana control: gana ceremonia.

Aplicada al ①–⑤:

| paso | ¿entrada? | ¿artefacto? | ¿compuerta? | |
|---|---|---|---|---|
| ① entender la idea | la idea cruda | no — es charla | no | conversación |
| **② investigar** | **la idea en una frase** | **la evidencia** | **¿hay fuentes?** | **← las tres** |
| ③ diferenciarse | la evidencia | no | no | conversación |
| ④ grill | la evidencia | no | no | conversación |
| ⑤ escribir | todo | `brief.md` | el ⑥ | ya la tiene |

**Sólo el ② califica**, y no por casualidad: es el único que corre **sin Javier**, el único que
**necesita herramientas**, y **el que falló** en la corrida.

```
hoy:        ① ② ③ ④ ⑤ ──────────────────────►  🛑 ⑥
                    una sola cabeza, sin frenar

propuesta:  ① ──►  ⏸  ② investigar  ⏸  ──►  ③ ④ ⑤ ──►  🛑 ⑥
              charla    delegable · con herramientas    charla
                        · deja evidencia.json (§7)
```

#### Los dos que ya hacen exactamente este corte

**`wayfinder`** marca su ticket `research` como el único *"AFK, agent-driven"*; los otros tres
—`grilling`, `prototype`, `task`— son con humano. **El corte ya está hecho en su taxonomía.**

**`research`** (mismo autor) **no es un paso de una conversación: es un agente que se lanza aparte**
y devuelve un archivo con citas. Su regla:

> *"Investigá contra **fuentes primarias** —docs oficiales, código fuente, specs, APIs de primera
> mano—, no un resumen de ellas. Seguí cada afirmación hasta la fuente que la posee."*
>
> *"Escribí los hallazgos en un solo archivo Markdown, citando la fuente de cada afirmación."*

Y sus tipos de ticket **ya existen en esta máquina**:

| wayfinder | SpecForge | quién |
|---|---|---|
| `research` — sin humano | `via: subagente` | **acá va el modelo barato, y acá hacen falta las herramientas** |
| `grilling` — conversación | `via: vos` | Javier |
| `prototype` · `task` | (no existen) | — |

> **No hace falta el mapa entero de wayfinder para ganar esto.** Alcanza con reconocer que el ② es
> un ticket de investigación y sacarlo de la charla. El mapa completo se discute cuando haya datos
> de T3 en adelante.

#### El contrato

1. **El ② sale como paso propio**, `via: subagente`, delegable a un modelo barato.
2. **Deja `evidencia.json`** (§7) — sin artefacto no hay corte, sólo una pausa.
3. **Su compuerta es la del §4, movida de lugar**: al menos una fuente, o `evidencia: baja`
   declarada. Deja de correr al final y corre **donde el fallo ocurre**.
4. **La parada es ⏸, no 🛑.** Javier mira la evidencia si quiere; no sella nada. El único sello del
   tramo sigue siendo el ⑥.

#### Lo que este documento NO decide, a propósito

1. ¿el ② es **un estado nuevo** o **una parada dentro de `brief`**? La máquina hoy no tiene forma de
   *"el brief va por la mitad"*, y agregar un estado toca `ordenDeEstados` y `paranAlFinal`.
2. ¿el ② puede **volver a correr**? Si el ④ (grill) descubre que falta investigar algo, ¿se
   reabre? Ahí aparece un bucle nuevo, y **todo bucle nuevo necesita saber quién lo cierra** — la
   pregunta del ㉑ otra vez.
3. ¿dónde vive un mapa de decisiones abiertas, si algún día se trae el wayfinder entero? `.docs/` se
   commitea, `.specforge/` es forense. Un mapa **no es ninguno de los dos**.

---

## 6. Las herramientas — nivel 0 primero

§5① y §5③ dicen *que* hay que darle herramientas al arnés. Esta sección dice **cuáles**, y la
respuesta cambió después de una medición de un segundo.

### El experimento que reordena la lista

Corrido el 2026-09-05, sin MCP, sin llave, sin configurar nada:

```bash
curl -s "https://registry.npmjs.org/-/v1/search?text=claude+code+usage+cost&size=3"
```

```
@cliftonc/finius   Local-first Claude Code usage & cost tracker — OTLP + transcript
claude-cup         Track your Value/Tokens…
tokenfin           TokenFin CLI — one command to auto-record Claude Code usage
```

**Tres competidores reales de la idea de prueba, al instante, con una llamada HTTP sin llave.**

> **A nemotron no le faltaba un MCP caro. Le faltaba `curl`.**
>
> Y el skill no lo sabe: `references/method.md` va directo a Tavily y al MCP de GitHub — o sea
> directo al nivel que necesita configuración y llave. **El nivel que funciona en cualquier arnés no
> está nombrado en ninguna parte.**

### La decisión de fondo: MCP o CLI

No es una preferencia de gusto. Para un producto multi-arnés hay una asimetría dura:

| | MCP | CLI (`curl`, `npx`, …) |
|---|---|---|
| declararlo | **distinto en cada arnés** | ninguno — está en el `PATH` |
| que `sf` compruebe que **está** | lee un config → sabe lo **declarado**, no lo **cargado** | `command -v` → **la verdad** |
| que `sf` compruebe que se **usó** | **imposible** | posible, si pasa por un envoltorio (§7) |

La última fila es la que decide, y conviene decirla sin vueltas: **`sf` nunca va a poder ver una
llamada MCP.** No es una limitación de esta versión: es que el MCP vive entre el modelo y el
servidor, y `sf` no está en el medio.

**Y medido:** los cuatro MCPs del `.mcp.json` de este repo corren con `npx` y `uvx`, que ya están en
la máquina. La dependencia real es la misma; lo único que cambia es cómo se le avisa al arnés.

### Los tres niveles

**Nivel 0 — sin llave, sin MCP, funciona en cualquier arnés con shell**

| Para qué | Cómo |
|---|---|
| ¿ya existe este paquete? | los registries de npm · PyPI · crates.io, por HTTP |
| ¿existe este repo, y qué issues tiene? | la API pública de GitHub (60 req/hora sin token) |
| leer una página concreta | `curl` |

**Nivel 1 — con llave, y es donde vive la búsqueda web de verdad**

`tavily` (o brave/exa). Sin esto **no hay señales de demanda** —Reddit, HN, sitios de reviews—, que
es la mitad del método de `method.md`. Es el nivel que hay que declarar por arnés.

**Nivel 2 — opcional**

`context7` para docs de librerías. `fetch` como MCP **sólo si el arnés no da shell**; con shell,
`curl` hace lo mismo sin declarar nada.

### La regla

```
NIVEL 0 PRIMERO, SIEMPRE.
Una herramienta que funciona en los tres arneses sin configurar nada vale más
que una mejor que hay que declarar tres veces.
```

Y eso le pone contenido concreto a los otros dos:

- **§5① (`sf doctor`)** — el nivel 0 se comprueba con `command -v`, que es un hecho. El nivel 1 sólo
  se puede leer de un config, y ahí `doctor` dice *"declarado"*, **nunca *"funciona"***. Son dos
  columnas distintas y colapsarlas sería mentir.
- **§5③ (`sf install`)** — no tiene que cablear cuatro MCPs para que la máquina sirva. Tiene que
  **garantizar el nivel 0 y avisar del nivel 1**. Eso baja muchísimo el costo de esa tarea.

### El riesgo, dicho

**El nivel 0 no reemplaza al 1.** Los registries contestan *"¿existe algo con este nombre?"*; no
contestan *"¿alguien se queja de esto en Reddit?"*. Un brief con nivel 0 solo puede mapear el
panorama y **no** puede traer señales de demanda. Eso hay que decirlo en el brief, no taparlo.

---

## 7. Los artefactos que faltan

### El grande: `evidencia.json`, el hermano que le falta al brief

El patrón ya existe en la máquina, y el brief es el único que queda afuera:

| fase | lo que lee el humano | lo que lee la máquina |
|---|---|---|
| planificación | `spec-design.md` | **`tareas.json`** |
| revisión | el detalle de los hallazgos | **`revision.json`** |
| **brief** | `brief.md` | **nada** |

**Ésa es la razón por la que la compuerta del §4 tiene que grepear `http` sobre prosa.** Es la parte
débil de ese arreglo, y se dijo cuando se hizo.

**El contrato.**

1. El ② escribe `.docs/evidencia.json`: una fila por afirmación.

```json
{"afirmacion":"ccusage hace exactamente esto",
 "procedencia":"retrieved",
 "fuente":"https://github.com/ryoppippi/ccusage",
 "consultado":"2026-09-05T22:14:03Z",
 "para":"panorama"}
```

2. `procedencia` es una **lista cerrada**: `retrieved` | `model-prior`. Un valor fuera de la lista
   es un archivo corrupto, con el mismo trato que un JSON roto — es el ⑤ de `vecinos.md` aplicado
   acá.
3. `para` dice a qué sirve la fila: `panorama` | `demanda` | `diferenciador`. **Sin esto no se puede
   distinguir un brief que mapeó competidores de uno que además encontró demanda**, que es
   justamente lo que el nivel 0 solo no puede hacer (§6).
4. **La compuerta lee el JSON, no la prosa.** El `grep` del §4 queda como el camino de compatibilidad
   para briefs viejos, no como el mecanismo.
5. **`brief.md` no se toca.** El JSON es lo que lee la máquina; la tabla del `## Panorama` es lo que
   lee un humano. Exactamente la relación que ya tienen `revision.json` y el resto.

**Lo que compra, y son cuatro cosas de una:**

| | |
|---|---|
| la compuerta frena sobre un **hecho** | y no sobre un patrón de texto (R3) |
| el ⑥ imprime conteos reales | **§5② sale gratis**: los datos ya están estructurados |
| `sf log` puede decir *"el ⑥ consultó 13 fuentes"* | la película gana una línea que hoy no tiene |
| es la entrada de cualquier chequeo futuro | fuentes distintas, dominios repetidos, fechas viejas |

> **Y es lo que hace que el corte del §5④ sea un corte de verdad.** Sin artefacto, sacar el ② afuera
> es poner una pausa; con artefacto, es poner una compuerta.

**El riesgo.** Duplicar información entre el JSON y la tabla del brief, y que se desincronicen (R6).
Se acepta con el mismo argumento que ya se aceptó para `revision.json`: **el JSON es la fuente y el
Markdown es su rendering**, no al revés. Si divergen, manda el JSON.

### El chico: el hueco negro entre dos `sf`

Lo que el tablero ve hoy, de la corrida real:

```
22:31:26  next  prd
22:33:51  done  prd  ✗ falta .docs/prd.md
```

**Dos minutos y medio sin absolutamente nada.** El registro tiene la película **de los estados**, no
la del trabajo.

**No se propone llenarlo todavía** — es un agujero grande y no está claro con qué. Lo que sí queda
escrito, porque es el marco de todo lo demás:

> **`sf` mira la máquina, no al que trabaja.** Todo artefacto nuevo que lo haga ver al que trabaja
> vale más que uno que le dé más detalle de lo que ya ve.

`evidencia.json` es el primero que cruza esa línea: es la primera cosa que `sf` va a saber sobre
**qué hizo** el modelo, y no sólo sobre en qué estado quedó.

### La opción grande, que NO se hace ahora: `sf buscar`

Si la búsqueda pasara por `sf`, entonces **el log lo escribe `sf`** — y deja de ser una afirmación
del modelo para ser un hecho. Eso cerraría el agujero que el §4 declara que no puede cerrar: **la
URL inventada**.

Y no rompe la regla de cero red en el enforcement: la red estaría en la **herramienta**, no en la
**compuerta**; la compuerta seguiría leyendo un archivo local.

**Pero es superficie nueva grande** —`comandos.Todos` se defiende sola— y hoy no hay dato que la
justifique.

> **Umbral que la despertaría:** una corrida **con las herramientas puestas** donde el brief cite una
> URL que no existe. Hasta entonces, es una idea con fecha, no una tarea.

---

## 8. La corrección al plan — el pid no mide delegación

**`por-tramos.md` §3 dice que el registro contestaría esto:**

> *"¿se lanzó el subagente que sf pidió? → entre un `next` con via=subagente y el `done` siguiente,
> ¿hay una ficha de lanzamiento? ¿o el `context` vino del mismo pid?"*

**La segunda mitad de ese test no funciona.** Medido:

| | ppid de las invocaciones de `sf` |
|---|---|
| **B** (opencode) | `29493` en las 8, sin excepción |
| **A** (Claude Code) | `30582`, `30957`, `5963`, `6191`, `19330`… todos distintos |

Leído en aislado, B parece "nadie delegó". Puesto al lado de A, se ve qué mide de verdad:
**cómo cada arnés abre la terminal.** Claude Code levanta un shell nuevo por cada comando; opencode
reusa el mismo padre. **Eso es arquitectura del arnés, no delegación.**

**Lo único que sí quedó probado:** no se usó `sf lanzar` en ninguna de las dos:
`.specforge/lanzamientos/` no existe en ninguna.

> **Consecuencia:** el agujero *"el orquestador no delega"* —uno de los cuatro del 2026-09-03—
> **sigue abierto, y el registro no lo cerró.** Hay que buscarle otra medida, y `por-tramos.md` §3
> hay que corregirlo.
>
> **Deducción, no hecho:** una medida que probablemente sí funcione es la marca del entorno que
> [`vecinos.md`](vecinos.md) §2 ya propone (`SPECFORGE_DELEGADO`), porque la pone `sf` y no depende
> de cómo el arnés maneje procesos. **No está probada.**

---

## 9. El orden, y por qué ése

```
✓  la compuerta del brief          HECHO — 4d7cc28

0  el nivel 0 entra en el skill    §6 · CHICO · ya habría evitado el fallo de B
1  evidencia.json + la compuerta   §7 · el artefacto que le falta al brief
   lee JSON en vez de prosa
2  el ② sale como paso propio      §5④ · el corte. Va con el 1, no sin él
3  sf doctor ve las herramientas   §5① · las dos columnas: `está` y `declarado`
4  toda parada entrega un acta     §5② · el ⑥ primero; sale casi gratis si el 1 está hecho
5  sf install garantiza el nivel 0 §5③ · mucho más barato que "cablear los MCPs"
6  el mapa entero de wayfinder     §5④ · sólo con datos de T3 en adelante
```

**Por qué el 0 va primero, y sale de la medición de §6.** Es un cambio de texto en un skill, no toca
Go, y **habría evitado el fallo de esta corrida**: una llamada a un registry sin llave devolvió tres
competidores reales al instante. Cuesta una tarde y tapa la causa raíz.

**Por qué el 1 va antes que el 2.** Sin `evidencia.json`, sacar el ② afuera es poner **una pausa**;
con él, es poner **una compuerta**. Un paso sin compuerta propia no es un paso: es una parada más, y
paradas de más ya hay una en discusión (la ⏸ del ⑨).

**Por qué el 4 bajó de puesto.** En la versión anterior de este documento el ⑥ mostrando evidencia
iba segundo, y era caro porque había que parsear prosa. **Con `evidencia.json` los conteos ya están
contados.** Es la misma tarea, después, y más barata: hacer una cosa en el orden correcto la volvió
casi gratis.

**Y el 4 creció de alcance sin encarecerse (2026-09-07).** Dejó de ser *"el ⑥ imprime tres campos"*
y pasó a ser *"toda parada entrega un acta de tres bloques"*. El costo no cambió —los datos ya
están contados y es una segunda salida al lado de `Texto()`— pero **lo que compra sí**: el bloque
`NO PUEDO COMPROBAR` es hoy el único lugar de la máquina donde se dice dónde termina lo que ella
sabe. Se implementa primero en el ⑥ y se extiende a las otras cuatro paradas.

**Por qué el 5 se abarató.** §5③ se escribió como *"cablear los MCPs por arnés"*, que es la tarea más
cara del documento. §6 la parte en dos: **garantizar el nivel 0** (barato, universal, `command -v`)
y **avisar del nivel 1** (leer un config y no prometer más que eso). Sólo la segunda mitad es cara,
y ya no es bloqueante.

---

## 10. Lo que se decide NO hacer, y qué lo despertaría

**No se aprieta más la compuerta del brief.** La que quedó ya atrapó el caso real. Exigirle más
—N fuentes, fuentes distintas, que el veredicto "coincida" con el panorama— es **castigar al modelo
por algo que es culpa nuestra**: no le dimos la herramienta. Primero el ① y el ③.

> **Umbral que lo despertaría:** una corrida **con las herramientas puestas** donde el brief cite
> una fuente sola y pase igual.

**No se toca `compuerta.PRD`.** Sigue siendo un `os.Stat` y su comentario sigue mintiendo. Esta
corrida **no lo explotó**, y sin un caso real aflojar o apretar es adivinar. Lo que sí se puede
hacer gratis: **arreglar el comentario**, que hoy dice *"y tenga cuerpo"* y es falso.

**No se construye `sf buscar`.** El razonamiento entero está en §7: cerraría el agujero de la URL
inventada, no rompe la regla de cero red en el enforcement, **y es superficie nueva grande sin un
dato que la pida.**

> **Umbral que lo despertaría:** una corrida con las herramientas puestas donde el brief cite una URL
> que no existe.

**No se trae el mapa de decisiones entero de wayfinder.** Se trae **el corte** (§5④), que es la parte
que esta corrida justificó. El mapa —tickets, dependencias, frontera, "listo cuando no queda nada por
decidir"— toca la forma de la máquina y arrastra un bucle nuevo sin dueño.

> **Umbral que lo despertaría:** un tramo de T3 en adelante donde el problema sea *"no sabemos lo
> suficiente para planificar"* y no *"planificamos mal"*. Son dolores distintos y hoy sólo se midió
> el segundo.

**No se implementa `--async` ni se toca el orquestador conversacional.** Sin cambios respecto de
`por-tramos.md` §9.

---

## 11. Cómo se sabrá que quedó bien

**Del ① y el ③, la prueba es la corrida de nuevo:**

```bash
# el mismo banco, la misma idea semilla, el mismo modelo débil
cd ~/projects/workspace/personal/sf-banco && ./comparar.sh
```

> **PASA si B cita fuentes.** No si B dice `hacelo` — eso sigue siendo juicio. Si con herramientas
> nemotron sigue sin citar nada, el problema **sí** es el modelo, y recién ahí se sabe.

**Del ②:** que `sf next` en el ⑥ imprima el veredicto, el conteo y el diferenciador **sin que ningún
modelo haya tenido que acordarse de hacerlo.**

**Del nivel 0 (§6):** la prueba es la de arriba, y tiene una forma más exigente que conviene escribir
porque distingue las dos hipótesis:

```
B cita fuentes           →  era la herramienta. Causa raíz confirmada y cerrada.
B sigue sin citar nada   →  era el modelo. Recién ahí se sabe, y es otro problema.
```

**De `evidencia.json` (§7):** que la compuerta del brief **deje de tener un `grep`** en el camino
principal, y que borrar la tabla del `## Panorama` del Markdown **no** cambie el veredicto de la
compuerta. Si lo cambia, el JSON no es la fuente y el §7 quedó a medias.

**Del corte del ② (§5④):** que un `sf next` pueda contestar *"el brief va por la mitad: la
investigación está, falta el grill"*. Hoy `estado.json` sólo sabe `brief_sellado: ""`, que es lo
mismo para *"no empecé"* y para *"me falta lo último"*.

---

## Apéndice — cosas medidas que no entran en ninguna sección

**Los tests punta a punta no corren con `go test ./...`.** Están detrás de `//go:build e2e`
(`cmd/sf/e2e_test.go:1`). Son **dos suites**, y hay que pedirlas por separado:

```bash
go test ./...                    # 482
go test -tags e2e ./cmd/sf/      # 14
```

Importó de verdad: **dos de esos 14 se rompieron** con el arreglo del `4d7cc28`, y ninguna suite
normal lo habría avisado. Conviene confirmar que CI corre las dos.

**`sf install` cambia el harness global.** Corrió una vez en el banco y movió
`~/.specforge/modelos.yaml` de `harness: commandcode` a `harness: claude-code`. El resto del
catálogo quedó intacto. **No es un bug** —es lo que el comando dice hacer— pero conviene saberlo
antes de instalar en una carpeta de prueba.
