# Headless — que `sf` lance, en vez de pedir que lancen

**Fecha:** 2026-08-31 · **Branch:** `refundation` · **Commit:** `1aa81dd`

> **Estado: SIN IMPLEMENTAR.** Es una spec de diseño, no un parte de obra.
>
> **La medición se hizo dos veces.** La primera ronda dejó las tablas de §2. La segunda, más
> profunda y el mismo día, **corrigió tres cosas de la primera** y encontró la que más pesa de todo
> el documento: **la carta del arnés dice "éxito" aunque no haya hecho nada** (§7). Todo lo de §2,
> §7, §8 y §10 salió de ejecutar los tres arneses en carpetas descartables — nada está leído de una
> doc ni supuesto. Las dos veces que se adivinó una variable o un flag se adivinó mal, y está
> contado en [`agnostico-al-harness.md`](agnostico-al-harness.md).
>
> **La decisión que falta es de Javier y está en §8.**

---

## 1. Por qué — es la causa, no el síntoma

Después de correr el bucle en los tres harness, Javier resumió así:

> *"parece que sí anda cambiando de harness, lo cual es bueno pero es medio enquilombado […] y el
> tema de los modelos de cada uno, debería ser más sencillo. Pero sí funciona. Creo que sigue
> siendo más interesante la ejecución headless."*

Tiene razón, y la razón es más profunda que la incomodidad. **Toda esa complejidad es la factura de
que lance el arnés.** Hoy `sf` no lanza modelos: le deja al arnés un papel con el número anotado y
le pide que llame. De ahí salen, una por una:

| Lo que molesta | Por qué existe |
|---|---|
| declarar perfiles en cada arnés | el arnés resuelve el modelo, `sf` sólo le pasa un nombre |
| los portamodelo (`sf-<alias>.md`) | única forma de fijarle el modelo a un subagente ajeno |
| reiniciar después de `sf model` | las definiciones de agente se leen al arrancar (medido, sonda 1) |
| `sf install` por arnés | permisos y rutas de skills, cada uno el suyo |
| detectar dónde estoy parado | `EnUso()`, el anidamiento, las seis variables de entorno |
| que `via: subagente` sea una sugerencia | `sf` no puede verificar quién hizo el trabajo |

**Si `sf` lanza, las seis no se simplifican: dejan de existir.** No queda nada que emparchar porque
no queda el problema.

Lo único irreducible es que **un id sólo significa algo en su proveedor** — `opus` existe en Claude
Code, `opencode/nemotron-3-ultra-free` en opencode. Eso se queda, y está bien: es un hecho del
mundo, y es exactamente lo que el catálogo transporta hoy sin opinar (R3).

### El cambio conceptual

Hoy `sf next` resuelve el modelo **según dónde estás parado**. En headless, `sf` deja de
preguntarse *dónde corro* y se pregunta sólo *qué puedo correr*. El arnés pasa de ser **el suelo**
a ser **un ejecutor de modelos entre otros** — que es lo mismo que `via: consola` ya hacía con
`codex exec`, sólo que generalizado.

> Dicho de otra forma: `via: consola` no era un caso raro. Era el diseño correcto, aplicado a un
> solo modelo porque todavía no se sabía.

### Y el que elige no es el estado: es Javier, paso por paso

**Esto lo agregó Javier el 2026-08-31 y es lo que faltaba para que el modo cierre.** El recorrido
que describió, tal cual:

> *"abro claude code, inicio sf, hago el brief, y el prd como es un derivado digo mm voy a usar
> nemotron de opencode […] una vez que termina le digo vamos a crear las us, le vuelvo a decir
> hacelo con opencode y nemotron […] llega el momento de implementar, y le digo a claude quiero que
> commandcode implemente porque me es más barato […] o podría ser al revés, usar opencode como
> principal y claude code y command code como secundario."*

Hay **un arnés principal** —el que abriste, el que conversa con Javier y el que corre los pasos
que no se delegan— y **ejecutores**, que pueden ser el principal mismo o cualquier otro.

Lo que decide quién ejecuta **no es el estado ni dónde estás parado**: es Javier, en el momento, y
por el motivo que sea — costo, velocidad, o que ese modelo es mejor para ese paso.

> **Y esto ya estaba previsto en el código.** `global.VarHarness` (`SPECFORGE_HARNESS`) existe con
> este comentario escrito hace semanas: *"Existe para dos cosas concretas: los tests, y `sf lanzar`
> — cuando sf spawnea un arnés headless, el hijo tiene que saber en cuál está corriendo, y su
> variable de entorno propia puede no estar puesta en un proceso hijo."* El agujero para esta pieza
> ya estaba hecho.

---

## 2. Lo que se midió — los tres, headless

Corrido el 2026-08-31 en carpetas descartables, con `git init` y nada más.
**Versiones:** `claude 2.1.251` · `opencode 1.18.25` · `cmd 1.38.2`.

| | Claude Code | opencode | Command Code |
|---|---|---|---|
| **entrar** | `claude -p "…"` | `opencode run "…"` | `cmd -p "…"` |
| **modelo** | `--model <alias>` | `-m provider/model` | `-m <model>` |
| **esfuerzo** | `--effort low\|medium\|high\|xhigh\|max` | `--variant high\|max\|minimal` | `--effort low\|medium\|high` |
| **stream** | `--output-format stream-json` | `--format json` | `--output-format json` (NDJSON) |
| **carpeta** | `--add-dir` | `--dir` | la cwd |
| **agentes** | `--agents <json>` (inline) · `--agent` | `--agent <nombre>` | (config) |
| **skills extra** | `--plugin-dir` | (config) | `--skill <path>` (repetible) |
| **background** | `--bg` + `agents`/`logs`/`attach`/`stop`/`rm` | — | — |
| **tope de gasto** | `--max-budget-usd` | — | — |
| **salida con forma** | `--json-schema <schema>` | — | — |
| **tope de vueltas** | — | — | `--max-turns` (exit 8 al tope) |
| **seguir sesión** | `--resume` · `--fork-session` | `-c` · `-s <id>` · `--fork` | `-c` · `-r` · `--session` · `--fork-session` |
| **permisos** | `--permission-mode` | *nada* | ver §8 |

### Las tres correcciones a la primera ronda

Las tres son de la segunda medición, y las tres iban en contra de lo que este documento decía.

**① Claude Code SÍ tiene flag de esfuerzo.** La primera ronda anotó *"no hay flag"* y construyó
sobre eso una sección entera sobre *"una asimetría real que `sf` no puede inventar y no debe
callar"*. **Es falso.** `claude --effort <low|medium|high|xhigh|max>` existe en la 2.1.251. Los
**tres** arneses expresan el esfuerzo por flag, y la asimetría no existe. Lo que sí queda en pie es
lo otro: que el esfuerzo deja de vivir adentro del portamodelo, con un nombre de campo distinto por
arnés, y pasa a ser un argumento.

**② El `--async` que §3 proponía construir, Claude Code ya lo tiene entero.** No es un flag suelto:
es un subsistema. `claude --bg` lanza y devuelve un id, y después hay `claude agents` (listar),
`claude logs <id>`, `claude attach <id>`, `claude stop <id>`, `claude rm <id>` y `claude respawn`.
Los otros dos no tienen nada equivalente. Eso cambia el cálculo de §9: no es *"construirlo"*, es
*"construirlo para dos y adoptarlo en uno"*, que es peor, porque son dos ciclos de vida distintos
conviviendo.

**③ Claude Code tiene dos cosas más que tocan este diseño de cerca.** `--agents <json>` define
agentes **en la línea de comando, sin archivo y sin reiniciar** — o sea que ahí el portamodelo
podría morir sin esperar a nada. Y `--json-schema <schema>` fuerza la salida a una forma declarada,
que es la tentación obvia para la ficha de §7 y **hay que resistirla**: sólo uno de los tres lo
tiene, y una ficha que existe en un arnés y no en los otros no es una ficha.

### Los permisos, probados

Cada prueba escribió un archivo y corrió un `echo`. Es el mínimo que necesita cualquier estado de
la máquina: sin escritura no hay artefacto, sin shell no hay `sf done`.

```
claude -p … --permission-mode bypassPermissions --model sonnet     ✓ escribió y ejecutó
opencode run --dir … -m opencode/mimo-v2.5-free …                  ✓ escribió y ejecutó, SIN --auto
cmd -p … -t --permission-mode auto-accept                          ✗
cmd -p … --tools-all                                               ✗   ← nuevo, y era el candidato
cmd -p … --tools-all -t --permission-mode auto-accept              ✗   ← nuevo
cmd -p … --yolo                                                    ✓ escribió y ejecutó
```

**opencode corre permisivo sin configurar nada** — `--auto` existe pero no hizo falta. Reconfirmado.

**Y `--tools-all` no es la salida de Command Code**, aunque su propia ayuda lo insinúe:
*"-p: enable every tool, **including the ones a headless run withholds**"*. Se probó solo y
combinado, y las dos veces se negó. Está en §8, que es donde vive esa decisión.

### Los códigos de salida

Con un modelo que no existe, **los tres devuelven `exit=1`** y dicen por qué. Las fallas duras se
detectan sin parsear nada.

Las blandas no. Eso es §7, y es el hallazgo grande.

### Lo que la medición cambia respecto de lo que se creía

- **`-p` no significa lo mismo en los tres.** En opencode `-p` es `--password`; headless ahí es el
  subcomando `run`. Un `sf lanzar` que asuma `-p` en los tres se rompe en el primer intento.
- **Command Code se actualiza solo en el medio de una corrida.** Una de las pruebas empezó con
  `Updated 1.38.2 → 1.39.1`. Existe `--no-auto-update` en Claude Code; en Command Code no aparece
  en la ayuda. Es ruido en el stream y una versión que cambia debajo.

---

## 3. Qué es `sf lanzar`

Un **modo extra**, no un reemplazo. Javier lo dijo antes de que existiera la spec:

> *"lo veo como un modo extra de sf, que él corra todo desde adentro"*

```bash
sf lanzar                                   # corre el paso que dice `sf next`, acá, y espera
sf lanzar --harness=opencode --alias=ultra  # el mismo paso, ejecutado por otro
sf lanzar --seco                            # imprime el comando que correría y no lo corre
```

`sf lanzar` no decide nada nuevo: le pregunta a `sf next` qué toca, resuelve el modelo con **la
misma cadena de precedencia de siempre** (§H2 de `agnostico-al-harness.md`), arma la línea de
comando desde el catálogo, y ejecuta.

**Lo que cambia no es la decisión: es quién aprieta el botón.**

### El "che, terminé" no hay que inventarlo

En el recorrido de Javier, el hijo *"le dice a claude code oye terminé"*. **Eso no necesita
mecanismo**: el principal llamó a `sf lanzar` desde su shell, se quedó esperando, y cuando el
proceso termina, terminó. No hay buzón, ni aviso, ni archivo de señal. El aviso **es** el exit.

Lo que el principal hace después son tres cosas y ninguna es nueva:

```
1. lee la ficha        →  qué modelo corrió, cuánto tardó, cuánto costó, cómo salió
2. corre `sf done`     →  la compuerta decide si eso vale
3. le dice a Javier    →  "terminó el ⑦, ¿next?"
```

El paso 2 es el que importa, y está desarrollado en §7.

### Y cierra un agujero que hoy está abierto

Hoy `via: subagente` es una instrucción **sin mecanismo**. El 2026-08-31 opencode leyó la skill,
entendió, e hizo el trabajo él mismo en vez de lanzar el subagente — y `sf done` aceptó el PRD sin
chistar, porque `compuerta.PRD` sólo comprueba que el archivo exista. `sf` **no tiene forma de saber
quién escribió un artefacto**. Se enteró porque el orquestador confesó.

Con `sf lanzar`, `via: subagente` deja de ser una sugerencia y pasa a ser un hecho. Es la misma
familia de agujero que el *"You do not rewrite the roadmap"* del ⑩: `sf` da instrucciones y confía.

---

## 4. De dónde salen el arnés, el modelo y el esfuerzo

**Del catálogo, y el catálogo lo llena `sf install`.** Es de Javier, el 2026-08-31:

> *"el tema de arnés y modelos lo podemos resolver con el install: cuando instalás seleccionás los
> arneses que vas a usar y los modelos que querés tener disponibles, entonces cuando el agente
> principal lanza el headless vas a saber el arnés, el modelo y el effort que necesitás."*

Esto no agrega una pieza: **conecta dos specs que estaban separadas y las vuelve una sola cosa.**

```
sf install   →  escribe el catálogo   →  sf lanzar lo lee
(una vez, con Javier)                    (cada paso, sin preguntar nada)
```

[`install-interactivo.md`](install-interactivo.md) ya diseñaba exactamente esa pantalla: elegís los
arneses, y por cada uno y por cada perfil elegís un modelo, le ponés **tu** alias y le ponés el
esfuerzo. Lo que no decía es **para qué termina sirviendo**. Sirve para esto: cuando el principal
quiere lanzar el ⑦ con nemotron, no tiene que saber ningún id, ningún flag ni ninguna sintaxis.
Dice `sf lanzar --harness=opencode --alias=ultra` y `sf` ya tiene todo lo demás escrito.

### El catálogo tiene las tres cosas y no le falta ninguna

El esquema de hoy (`global.Modelo`, `global.Catalogo`) ya guarda, por arnés y por perfil: el `id`,
el `alias` de Javier, el `esfuerzo`, y el `via`. **Es exactamente el juego de datos que `sf lanzar`
necesita** — un id sin un arnés no sirve, y un esfuerzo sin un modelo tampoco.

Lo único que hay que agregarle es **la tabla de cómo se escribe cada cosa en cada arnés**, que es
§2 y son seis filas:

```
claude-code   →  claude -p <prompt> --model <id> --effort <e> --output-format stream-json …
opencode      →  opencode run --dir <raiz> -m <id> --variant <e> --format json <prompt>
commandcode   →  cmd -p <prompt> -m <id> --effort <e> --output-format json …
```

Esa tabla vive en `sf`, es chica, y es lo único que hay que tocar el día que aparezca un cuarto
arnés. **No vive en el catálogo**: el catálogo es de Javier, la tabla es del programa.

### Lo que esto zanja de una vez

- **No hace falta que el arnés principal sepa nada de los otros.** Le pasa un alias y un arnés.
- **No hace falta un listado de modelos en tiempo de lanzamiento.** Ya está declarado.
- **`sf models` (el ② de `install-interactivo.md`) deja de ser una comodidad y pasa a ser cómo se
  llena el catálogo.** Sube de prioridad, y sobrevive a los dos escenarios.
- **La precedencia no cambia.** `--alias` explícito gana; sin él, el perfil del estado; sin eso, el
  default del perfil. Es la cadena de H2, aplicada donde siempre.

> **Y `sf install` sigue sin ser de la máquina.** No mira el estado ni lo mueve. Escribe el
> catálogo, y el catálogo es configuración, no estado.

---

## 5. Siete estados sí, dos no

**Headless sólo puede correr lo que no conversa.** Y eso ya está resuelto en el código: el mapa
`conversan` de `maquina.go`.

| | Estados | `sf lanzar` |
|---|---|---|
| conversan | ①–⑤ el brief · ⑧ la constitución | **no** — necesitan a Javier |
| no conversan | ⑦ · ⑨ · ⑩ · ⑫–⑯ · ⑱–⑳ · ㉑㉒ · ㉓ | sí |

No es una limitación del modo: es la misma regla que hizo que el ⑧ dejara de ser subagente el
2026-08-31 (ver `maquina-estados.md` §5.1 y el commit `4acf47f`). Un paso que necesita la opinión
de Javier no se puede delegar **a nadie**, ni a un subagente ni a un proceso.

**Y es exactamente el recorrido que describió Javier**: el brief y la constitución los hace el
principal porque *"esta fase necesita de mi intervención"*; el PRD, las historias y la
implementación se reparten. La regla ya estaba escrita; el modo la usa sin tocarla.

> Y por eso `sf lanzar` **no es un `sf run` que corre el proyecto entero solo**. La máquina para
> donde siempre paró. Lo que se automatiza es el tramo entre dos paradas — que es exactamente el
> tramo que el horizonte de §5.1 ya anuncia.

---

## 6. El prompt — la parte que no es un flag

Ésta parecía la parte difícil. **Resultó ser la más barata**, y esta sección es sobre todo el
registro de por qué.

Hoy el arnés carga la skill: `sf next` dice `skill: sfp-po` y el orquestador la encuentra donde ese
arnés busque. En headless no hay orquestador, así que la pregunta era **qué manda `sf` en el
prompt** — y en particular si tenía que mandar el método entero, resuelto por su cuenta.

La respuesta corta, medida: **no**. Alcanza con nombrar la skill.

```
prompt = "Usá la skill sfp-po." + el sobre (lo que hoy devuelve `sf context`)
```

### La duda que había acá, y por qué se cayó

> **Corrección del 2026-08-31.** Esta sección decía que `sf` tenía que **pegar** el texto de la
> skill en el prompt, y que apuntarle al modelo dónde está exigía escalar permisos. **Estaba mal, y
> lo marcó Javier**: *"a opencode le podés decir que use la skill x cuando usás headless y deberá
> leerla"*. Mi prueba había apuntado a una **ruta absoluta suelta**, que es lo que opencode bloquea;
> el mecanismo de skills del propio arnés es otra cosa y funciona. Queda escrito el error porque el
> razonamiento que llevó a él —"si una lectura externa se rechaza, todas se rechazan"— es el que
> hay que no repetir.

Una skill no contiene el método entero: contiene **punteros**.

```
sf-build/SKILL.md:82        Compose **sfx-tdd**. The cycle is unchanged: …
sfp-constitucion/SKILL.md   **Working rules** — see `references/reglas-de-trabajo.md`
```

La preocupación era que en headless nadie los resolviera, y que **fallara en silencio**: un modelo
que lee "Compose sfx-tdd" y no puede traerlo inventa algo parecido a TDD y devuelve un artefacto que
parece bien.

**No pasa. Los arneses resuelven los punteros solos, headless, sin escalar nada.**

### Lo que se midió — la cadena entera

Con dos skills de sonda, `sf-sonda-comp` → compone `sf-sonda-ref` → que manda a su
`references/reglas.md` con un token único adentro:

```
$ opencode run --dir <proyecto> -m …  "Usá la skill sf-sonda-comp y seguí lo que diga."
  → Skill "sf-sonda-comp"                              el nombre lo dio sf
  → Skill "sf-sonda-ref"                               la composición resolvió sola
  → Read …/skills/sf-sonda-ref/references/reglas.md    y el reference también
  TOKEN-REF-9F3A
```

**Tres niveles, sin `--auto`**, y la ruta del `references/` está FUERA del `--dir`. Cargar una skill
por nombre habilita su carpeta; leer una ruta suelta no. Son dos permisos distintos y sólo el
segundo está cerrado.

Y con las skills de verdad, en los dos arneses que no son Claude Code:

```
$ opencode run --dir … -m …   "Usá la skill sfp-po. Decime el primer '## '."
  → Skill "sfp-po"
  → Read ~/.claude/skills/sfp-po/templates/prd.tmpl.md      ← hasta los templates
  Actores

$ cmd -p "Usá la skill sfp-po. Decime el primer '## '."     (SIN --yolo)
  The size is set by its two readers, and nothing else      = sfp-po/SKILL.md:24
```

Dos cosas más que salieron de ahí:

- **opencode encuentra `~/.claude/skills/` sin que nadie se lo diga.** No hace falta instalarle nada.
- **La compuerta de Command Code (§8) es sólo de escritura y shell.** Leer una skill anda sin
  `--yolo`. Achica el problema: lo que falta ahí no es *entender*, es *actuar*.

### Entonces `sf` apunta, no pega

```
prompt = "Usá la skill sfp-po." + el sobre (lo que hoy devuelve `sf context`)
```

Y todo lo demás —la composición, los `references/`, los `templates/`— lo resuelve el arnés, igual que
hoy adentro del TUI. **`sf lanzar` no tiene que entender la estructura interna de ninguna skill**,
que era el acoplamiento nuevo que esta sección temía.

La única condición es que el arnés **encuentre** las skills, y eso ya es trabajo de `sf install`:
para Command Code escribe `settings.skills`, y opencode no necesita nada. Es una instalación por
arnés, una vez — no un archivo por alias con reinicio, que es lo que headless viene a matar.

### Lo que sí queda por decidir

**¿Y si el modelo decide no cargar la skill?** Apuntar es una instrucción, y una instrucción se
puede ignorar — es exactamente lo que hizo opencode el 2026-08-31 con `via: subagente`. La ficha de
§7 tiene que poder contestar **si la skill se cargó**, porque el stream lo dice (`→ Skill "sfp-po"`)
y es la única señal barata de que el paso corrió como se pidió.

> Pegar el texto sigue existiendo como **plan B por arnés**, no como diseño. Si aparece uno que no
> tenga mecanismo de skills, ese arnés recibe el texto y listo.

---

## 7. La ficha y el registro — y por qué la carta del arnés miente

Los tres arneses emiten **un stream de eventos JSON**. De ahí salen dos cosas distintas, y
confundirlas sería un error:

| | Qué es | Para quién |
|---|---|---|
| **el registro** | el stream completo, JSONL, tal cual salió | forense. Se guarda y no se lee salvo que algo salga mal |
| **la ficha** | un JSON chico: qué se lanzó, con qué modelo, cuánto tardó, cómo terminó | el bucle. Es lo que `sf next` mira |

### El hallazgo: `success` no quiere decir que se hizo el trabajo

**Medido, y es el dato más importante de este documento.**

Se le pidió a Command Code, headless, la tarea más chica que existe: escribir un archivo y correr un
`echo`. **Se negó** —es §8— y no escribió nada. Su última línea fue:

```json
{"type":"result","subtype":"success","sessionId":"430e25dc-…","stopReason":"end_turn",
 "usage":{…},"durationMs":26944,
 "finalText":"No puedo completar la tarea: las herramientas requieren permisos que no están
              habilitados en esta sesión (necesitas `--yolo` …)"}
```

**`subtype: success`. Y `exit=0`.**

No es un bug del arnés. Para él, *"success"* significa **que la conversación terminó bien**, no que
el trabajo se hizo. Es un empleado que vuelve puntual: eso no dice nada sobre la tarea.

Y no es un caso raro de Command Code — es la forma de todos. Los tres devuelven `exit=1` con un
modelo que no existe (falla dura), y **ninguno tiene forma de devolver "el modelo no hizo lo que le
pediste"** (falla blanda), que son casi todas las que importan: el que se negó, el que hizo la
mitad, el que entendió otra cosa.

> **Consecuencia de diseño, y es una sola frase:** el principal **nunca** le cree a la carta. Cuando
> el hijo termina, corre `sf done`, y **la compuerta decide**. La carta cuenta qué pasó; no decide
> si se avanza.
>
> Esto no es una regla nueva. Es la que ya rige para el orquestador —*"Vos no leés el trabajo. El
> subagente puede decir 'terminé' y estar equivocado"*— aplicada a un proceso en vez de a un
> subagente. **El diseño de Javier encaja sin tocar nada; lo único que hay que no hacer es
> confundir la carta con la evidencia.**

### Las tres cartas son distintas, y una no existe

| | Qué deja al final | Sirve como ficha |
|---|---|---|
| **Claude Code** | `type:result` con `total_cost_usd`, `usage` completo, `session_id`, `stop_reason`, `duration_api_ms`, `modelUsage` | la más rica |
| **Command Code** | `type:result` con `subtype`, `sessionId`, `stopReason`, `usage`, `durationMs`, `finalText` | alcanza |
| **opencode** | **nada a nivel corrida.** Su último evento es `step_finish` — de *un paso*, con `tokens` y `cost` | **no** |

Los eventos de opencode son `step_start` / `tool_use` / `step_finish`, y se repiten por paso. No hay
un evento de cierre que hable de la corrida entera.

**Entonces `sf` no lee la carta del arnés: escribe la suya.** Toma los tres streams, que son
distintos, y produce **una** ficha con la misma forma siempre:

```
qué estado se lanzó · qué arnés · qué modelo e id · qué esfuerzo
cuánto tardó · cuánto costó (si el arnés lo dice) · con qué código salió
si la skill se cargó (el stream lo dice, §6) · qué herramientas usó
```

Es trabajo mecánico y acotado —tres normalizadores chicos— pero **es trabajo, y no estaba en la
primera versión de esta spec**, que asumía que la carta venía hecha.

> **Y la tentación que hay que resistir:** Claude Code tiene `--json-schema`, que forzaría la salida
> a la forma que uno declare. Es la solución elegante para uno de los tres. Una ficha que existe en
> un arnés y no en los otros no es una ficha — es una excepción con buena prensa.

---

## 8. La compuerta de Command Code — la decisión que falta

**Es de Javier, no mía, y por eso está sola en su sección.**

Command Code en `-p` **no escribe archivos ni ejecuta shell**. Probado el 2026-08-31 con:

```
--permission-mode dont-ask       ✗
--permission-mode auto-accept    ✗
--permission-mode bypass         ✗
-t / --trust                     ✗
--tools-all                      ✗   ← el candidato de la segunda ronda
--tools-all -t --permission-mode auto-accept   ✗
--yolo                           ✓   único que cede
```

**`--tools-all` era el candidato bueno y no alcanza**, y entender por qué achica la pregunta. Su
ayuda dice *"enable every tool, including the ones a headless run withholds"* — o sea que suena
exactamente a lo que falta. Pero son **dos compuertas distintas**: *tener* la herramienta disponible
y *tener permiso* de usarla. `--tools-all` abre la primera. La que está cerrada es la segunda.

Y no hay que deducirlo: **lo dice el propio Command Code** cuando se niega.

> *"las herramientas requieren permisos que no están habilitados en esta sesión (necesitás `--yolo`
> o `--dangerously-skip-permissions`)"*

O sea que la pregunta ya no es *"¿falta encontrar el flag?"*. No falta. Es `--yolo` o nada, y
`--yolo` es, en la ayuda del propio Command Code, **alias de `--dangerously-skip-permissions`**,
marcado en su doc para *"throwaway environments only"*.

Las opciones, sin recomendación de mi parte porque el riesgo lo corre él:

| | Qué implica |
|---|---|
| **A. `sf lanzar` nunca pasa `--yolo`** | Command Code queda fuera del modo headless. Los otros dos andan. Honesto y limitado. |
| **B. Lo pasa si Javier lo declara** | Un campo explícito en el catálogo (`permisos: yolo`) por arnés. `sf` no lo elige: lo transporta, igual que un id. Es R3 aplicado a los permisos. |
| **C. Lo pasa siempre** | No. `sf` estaría tomando por Javier una decisión de seguridad que su propio proveedor marca como peligrosa. |

**B es la única que cabe en el diseño existente** —el catálogo ya es el lugar donde vive lo que
Javier decidió y `sf` sólo acarrea, y con §4 eso es más cierto que antes— pero **A es una respuesta
legítima** si prefiere no tener ese flag escrito en ningún archivo suyo.

> **El dato que ordena la comparación:** opencode escribe y ejecuta sin pedir permiso y sin ningún
> flag. O sea que el permiso que `--yolo` concede en Command Code es el que opencode ya da de
> arranque. La diferencia real entre los dos no es cuánto se arriesga: es que uno lo hace explícito
> y con un nombre feo.

---

## 9. El alcance de la primera versión — `--async` queda afuera

La versión anterior de §3 ofrecía `sf lanzar --async` ("lo larga y devuelve un id; el bucle sigue")
como parte del modo. **Va afuera de la primera versión, y el argumento es del propio recorrido de
Javier.**

Leído en orden, su pseudo-recorrido es: brief → PRD con nemotron → **vuelve y avisa** → constitución
con el principal → historias con nemotron → **vuelve** → implementar con Command Code → **vuelve**.

**Cada paso espera al anterior. No hay nada corriendo en paralelo.** El modo que describió es
sincrónico de punta a punta.

Y `--async` es la parte cara del proyecto:

- **estado nuevo en `estado.json`** — `sf next` tiene que poder decir *⏳ hay algo corriendo* y no
  proponer trabajo encima;
- **ciclo de vida** — quién cosecha el proceso, qué pasa si queda colgado, qué pasa si `sf` muere
  con el hijo vivo;
- **y ahora, además, dos implementaciones** — Claude Code ya trae la suya entera (`--bg` + `agents`
  / `logs` / `stop` / `attach` / `rm`, §2) y los otros dos no tienen nada. O `sf` construye la suya
  e ignora la que ya existe, o convive con dos ciclos de vida distintos. Las dos son caras.

> Sacarlo no lo mata: lo pone después, con el bucle ya corriendo en modo headless sincrónico y con
> la experiencia de haberlo usado. Que es exactamente el orden en el que este proyecto acertó las
> otras veces.

### Y hay un comentario en el código que hay que retirar a propósito

`sf/cmd/sf/main.go` abre con esto, y es el argumento fundacional del binario:

> *"sf arranca, contesta y se muere. Dura milisegundos. **Por eso NO puede lanzar a nadie**: un
> programa muerto no tiene manos. El único vivo durante toda la sesión es el agente, y por eso el
> orquestador es él."*

`sf lanzar` lo contradice de frente: `sf` se queda vivo mientras el hijo trabaja. **Cambiar de idea
está bien; dejar las dos escritas no.** El día que esto se implemente, ese párrafo se reescribe en
el mismo commit, diciendo qué cambió y por qué — no se borra en silencio.

Y hay un matiz que sobrevive y conviene conservar: `sf next`, `sf done` y todo el resto **siguen**
durando milisegundos. El único que se queda vivo es `lanzar`, y sólo mientras espera.

---

## 10. Las asperezas medidas

Chicas, todas comprobadas, y todas muerden en la primera corrida si no están anotadas.

**① Los modelos gratis cuelgan, y cuelgan callados.** Una corrida con
`opencode/nemotron-3.5-lightning-free` estuvo **7 minutos 40 segundos emitiendo cero bytes** y hubo
que matarla. La misma tarea con `opencode/mimo-v2.5-free` tardó **19 segundos**. No es que sea lento:
es que no hay señal de vida.

> **`sf lanzar` necesita un tiempo máximo**, y el default no puede ser "para siempre". Y como el
> stream no llega progresivamente en ese caso, tampoco sirve mirar si hay actividad.

**② Hay que pasarle `< /dev/null`.** Claude Code, sin stdin, avisa *"no stdin data received in 3s,
proceeding without it"* y espera esos tres segundos. Detalle mínimo, pero son tres segundos por
lanzamiento y una línea de ruido en el registro.

**③ Un id puede resolverse a otro proveedor.** Pidiendo `opencode/mimo-v2.5-free`, opencode reportó
`> kiro · mimo-v2.5-free`. **No está confirmado** si es sólo cómo lo muestra o si de verdad enrutó a
otro lado. Hay que averiguarlo antes de que el catálogo prometa un id que no es el que corre — es
R3 mirado desde el otro lado: `sf` no traduce ids, pero el arnés sí puede.

**④ No hay candado sobre el estado.** `sf` no tiene ningún lock sobre `estado.json` (lo único
parecido es un `O_EXCL` en `entradas.go:64`, y es para crear entradas, no para escribir el estado).
Si dos procesos corren `sf done` sobre el mismo repo al mismo tiempo, **gana el último y en
silencio**.

Esto **no lo introduce headless** —ya es cierto hoy con tres arneses abiertos en el mismo repo—,
pero conviene ver que el diseño de Javier lo *mejora*:

| | Qué es | Riesgo |
|---|---|---|
| tres arneses abiertos en paralelo | tres conductores, un volante | real, y existe hoy |
| un principal que reparte | un conductor que pide trabajos | **bajo** — todo pasa por un lugar y de a uno |

El paralelo vuelve a aparecer sólo con `--async`, que es §9, y es una razón más para dejarlo para
después. Cuando llegue, el candado es requisito, no adorno.

---

## 11. Lo que NO cambia

Para que quede dicho antes de que alguien lo "aproveche":

- **Las compuertas siguen en `sf`.** Headless no las relaja ni una, y con §7 son más importantes que
  antes, no menos.
- **Las paradas siguen siendo de Javier.** El ⑥, el ⑧ y el ⑰ no se automatizan.
- **`sf next` sigue siendo consulta pura.** Nunca escribe el estado. `sf lanzar` sí, y por eso son
  dos comandos y no uno.
- **R3 sigue vigente.** `sf` arma la línea de comando desde el catálogo; no elige un modelo parecido
  ni traduce un id de un proveedor a otro.
- **El modo orquestado no se borra.** Quien quiera seguir trabajando adentro del arnés, sigue.
- **`sf install` sigue sin ser de la máquina**, aunque §4 le dé un consumidor nuevo. Escribe
  configuración, no estado.

---

## 12. Cómo se sabrá que quedó bien

```bash
# el mismo paso, el mismo estado en disco, tres ejecutores distintos
sf lanzar --seco                            # imprime la línea, no corre nada
sf lanzar --harness=opencode --alias=ultra  # corre el ⑦ y deja .docs/prd.md
sf done                                     # la compuerta pasa igual que si lo hubiera hecho un subagente
```

Y las pruebas que de verdad cierran el diseño:

1. **Sin `.opencode/agents/`.** Borrar los portamodelo y que `sf lanzar` funcione igual. Si hace
   falta alguno, el modelo no se está pasando por flag y estamos donde empezamos.
2. **Sin reiniciar nada.** Declarar un alias con `sf model` y lanzarlo **en la misma sesión**. Hoy
   eso es imposible —los agentes se leen al arrancar— y es la señal más limpia de que la causa
   desapareció.
3. **Que un `success` mentiroso no avance el estado.** Lanzar un paso con Command Code **sin** el
   permiso de §8, o con un modelo que se niega, y comprobar que la ficha lo registra como terminado
   y **`sf done` igual dice ✗**. Es la prueba de §7, y es la única que verifica que la carta no se
   confundió con la evidencia.
4. **Un solo comando para dos arneses.** Que el principal lance el ⑦ en opencode y el ⑱ en Command
   Code sin que Javier haya escrito un id, un flag ni una ruta — sólo `--harness` y `--alias`. Es la
   prueba de §4.
5. **Que un modelo colgado no cuelgue el bucle.** Lanzar contra un modelo que no contesta y que
   `sf lanzar` corte solo, con la ficha diciendo que se agotó el tiempo. Es §10.①.

> **La prueba de que el modo sirve no es que corra: es que el ⑦ salga bien con un modelo que no es
> de Anthropic, sin que nadie haya tenido que configurar el arnés.** Eso es lo que hoy cuesta una
> instalación por arnés, un archivo por alias y un reinicio.
