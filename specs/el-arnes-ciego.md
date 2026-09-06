# El arnés ciego — lo que vio la primera corrida real

**Fecha:** 2026-09-05 · **Branch:** `refundation` · **Commit:** `4d7cc28`

> **Estado: UN ARREGLO HECHO, CUATRO PROPUESTOS.** Sale de la primera corrida de T1 con dos modelos
> y la misma idea semilla. **Todo lo de §1, §2 y §6 está medido** — carpetas, registros y salidas
> reales, no impresiones. Lo que es deducción está marcado como deducción.

Este documento continúa a [`por-tramos.md`](por-tramos.md) —que pidió esta corrida y escribió el
método— y a [`vecinos.md`](vecinos.md). Lo que agrega es **el resultado**: la primera vez que un
artefacto de esta máquina lo escribió un modelo y no una mano.

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

### ② El ⑥ muestra la evidencia — Javier decide con datos

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

**El contrato.** `sf` ya tiene el archivo en la mano cuando corre la compuerta. Que lo diga él:

```
🛑 ⑥ — el brief está. Lo sellás vos.
   propone:        no-lo-hagas
   evidencia:      13 fuentes citadas · 1 afirmación sin verificar
   se diferencia:  <la línea que el brief ya escribe>

   sf approve  ·  sf reject "motivo"
```

**Por qué éste es el más barato de los cuatro.** Los tres datos ya están en `brief.md` y la
compuerta ya lo abre para contar links. Es imprimir lo que ya se leyó.

**El riesgo.** Que `sf` empiece a resumir el brief — no es su trabajo (R1). **Se imprimen tres
campos y ninguno se interpreta.** Si el brief no trae diferenciador, se dice que no lo trae.

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

### ④ El ①–⑤ se parte en decisiones — lo que se trae de wayfinder

**El dolor.** `sfp-scout` hace **cinco pasos de un saque, en una sola cabeza, sin frenar en el
medio**. El único control está al final, en el ⑥. Si se degrada en el paso ②, **nadie se entera
hasta el final** — que es literalmente lo que pasó en B.

**De dónde sale.** `mattpocock/skills`, skill `wayfinder`. Planifica trabajo que no entra en una
sesión como **un mapa de decisiones**, no de tareas. Sus cuatro reglas:

| Regla | Qué dice |
|---|---|
| *"Plan, don't do"* | el mapa produce **decisiones**, no entregables |
| **el test de la niebla** | *"la prueba es si podés enunciar la pregunta con precisión ahora, no si podés contestarla ahora"* |
| **un ticket por sesión** | salvo los de investigar |
| **el destino fija el alcance** | lo que queda más allá, está afuera |

**Y sus tipos de ticket ya existen en esta máquina:**

| wayfinder | SpecForge | quién |
|---|---|---|
| `research` — sin humano, lo hace el agente | `via: subagente` | **acá va el modelo barato, y acá hacen falta los MCPs** |
| `grilling` — conversación | `via: vos` | Javier |
| `prototype` · `task` | (no existen todavía) | — |

**La observación que lo justifica.** [`por-tramos.md`](por-tramos.md) **es** un mapa de wayfinder
escrito a mano: destino, T0..T7 con dependencias forzadas, cada tramo produce **hallazgos y no
features**, y su §11 lista *"las cuatro preguntas que hoy no se pueden contestar"*.

> **Cuando la máquina no alcanzó, se escribió un wayfinder a mano.** Ésa es la señal más fuerte de
> que falta la pieza.

**El contrato — deliberadamente incompleto.** Éste **no se diseña en este documento**. Lo que sí
queda escrito es qué tiene que contestar el diseño:

1. ¿el ①–⑤ pasa a ser **un estado con vueltas** o **N estados chicos**? La máquina hoy no tiene
   forma de "el brief va por la mitad".
2. ¿dónde vive el mapa? `.docs/` es lo que se commitea; `.specforge/` es forense. Un mapa de
   decisiones abiertas no es ninguno de los dos todavía.
3. ¿cada ticket tiene compuerta propia? Si no la tiene, es §3 otra vez con más pasos.
4. **la del corte**: el mapa termina *"cuando no queda nada por decidir"*. ¿Quién lo declara? Si lo
   declara el mismo que abre los tickets, no hay punto fijo — es el mismo problema del ㉑.

**Por qué va último.** Es el único de los cuatro que **cambia la forma de la máquina**. Los otros
tres hacen visible lo que hoy es invisible; éste mueve estados. Y `por-tramos.md` §9 es explícito:
no se construyen piezas nuevas hasta que la corrida dé datos. **Ahora hay datos de T1 y de medio T2
— y de T3 en adelante, ninguno.**

---

## 6. La corrección al plan — el pid no mide delegación

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

## 7. El orden, y por qué ése

```
✓  la compuerta del brief        HECHO — 4d7cc28
①  sf doctor ve las herramientas  diagnóstico antes que tratamiento
②  el ⑥ muestra la evidencia      media tarde, y ya está todo leído
③  sf install cablea los MCPs     la palanca grande, y la más cara
④  el ①–⑤ en decisiones           cambia la forma. Necesita diseño, no improvisación
```

**El ① va antes que el ③** porque no se arregla lo que no se ve, y porque un `doctor` que avisa ya
alcanza para que nadie más corra a ciegas mientras el ③ se piensa.

**El ① y el ② son el mismo día de trabajo**, y los dos son lo mismo: **hacer visible lo que hoy es
invisible.** Cierran la mayor parte del dolor que encontró esta corrida.

---

## 8. Lo que se decide NO hacer, y qué lo despertaría

**No se aprieta más la compuerta del brief.** La que quedó ya atrapó el caso real. Exigirle más
—N fuentes, fuentes distintas, que el veredicto "coincida" con el panorama— es **castigar al modelo
por algo que es culpa nuestra**: no le dimos la herramienta. Primero el ① y el ③.

> **Umbral que lo despertaría:** una corrida **con las herramientas puestas** donde el brief cite
> una fuente sola y pase igual.

**No se toca `compuerta.PRD`.** Sigue siendo un `os.Stat` y su comentario sigue mintiendo. Esta
corrida **no lo explotó**, y sin un caso real aflojar o apretar es adivinar. Lo que sí se puede
hacer gratis: **arreglar el comentario**, que hoy dice *"y tenga cuerpo"* y es falso.

**No se implementa `--async` ni se toca el orquestador conversacional.** Sin cambios respecto de
`por-tramos.md` §9.

---

## 9. Cómo se sabrá que quedó bien

**Del ① y el ③, la prueba es la corrida de nuevo:**

```bash
# el mismo banco, la misma idea semilla, el mismo modelo débil
cd ~/projects/workspace/personal/sf-banco && ./comparar.sh
```

> **PASA si B cita fuentes.** No si B dice `hacelo` — eso sigue siendo juicio. Si con herramientas
> nemotron sigue sin citar nada, el problema **sí** es el modelo, y recién ahí se sabe.

**Del ②:** que `sf next` en el ⑥ imprima el veredicto, el conteo y el diferenciador **sin que ningún
modelo haya tenido que acordarse de hacerlo.**

**Del ④:** que se pueda contestar *"el brief de f-X está por la mitad, falta resolver la decisión
d-2"* mirando el estado, y no leyendo un Markdown.

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
