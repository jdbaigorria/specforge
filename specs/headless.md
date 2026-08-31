# Headless — que `sf` lance, en vez de pedir que lancen

**Fecha:** 2026-08-31 · **Branch:** `refundation` · **Commit:** `d1c7856`

> **Estado: SIN IMPLEMENTAR.** Es una spec de diseño, no un parte de obra. Lo que sí está hecho es
> **la medición**: los tres harness se corrieron headless en este repo el 2026-08-31 y las tablas
> de §2 son salida real, no documentación leída. Las dos veces que hoy adiviné una variable de
> entorno o un flag, estaba equivocado —está contado en
> [`agnostico-al-harness.md`](agnostico-al-harness.md)— así que acá no hay nada supuesto.
>
> **La decisión que falta es de Javier y está en §7.**

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

---

## 2. Lo que se midió — los tres, headless

Corrido el 2026-08-31 en carpetas descartables, con `git init` y nada más.

| | Claude Code | opencode | Command Code |
|---|---|---|---|
| **entrar** | `claude -p "…"` | `opencode run "…"` | `cmd -p "…"` |
| **modelo** | `--model <alias>` | `-m provider/model` | `-m <model>` |
| **esfuerzo** | **no hay flag** | `--variant high\|max\|minimal` | `--effort low\|medium\|high` |
| **stream** | `--output-format stream-json` | `--format json` | `--output-format json` (NDJSON) |
| **carpeta** | `--add-dir` | `--dir` | la cwd |
| **agentes** | `--agents <json>` | `--agent <nombre>` | (config) |
| **seguir sesión** | `--resume` | `-c` · `-s <id>` · `--fork` | — |
| **permisos** | `--permission-mode` | *nada* | ver §7 |

### Las tres pruebas de permisos, con su resultado

Cada una escribió un archivo y corrió un `echo`. Es el mínimo que necesita cualquier estado de la
máquina: sin escritura no hay artefacto, sin shell no hay `sf done`.

```
claude -p … --permission-mode bypassPermissions --model sonnet     ✓ escribió y ejecutó
opencode run --dir … -m opencode/nemotron-3.5-lightning-free …     ✓ escribió y ejecutó, SIN --auto
cmd -p … -t --permission-mode auto-accept                          ✗ "Necesito permisos para
                                                                      escribir archivos y
                                                                      ejecutar comandos shell."
cmd -p … --yolo                                                    ✓ escribió y ejecutó
```

**opencode corre permisivo sin configurar nada** — `--auto` existe pero no hizo falta. Es coherente
con lo medido en la sonda 2 el 2026-08-30.

### Lo que la medición cambia respecto de lo que yo creía

- **El `--variant` de opencode y el `--effort` de Command Code existen.** El esfuerzo, que hoy `sf`
  escribe adentro del portamodelo con un nombre de campo distinto por arnés (`reasoningEffort` en
  Command Code), pasa a ser **un flag**. Una tabla menos.
- **Claude Code no tiene flag de esfuerzo.** Es una asimetría real y hay que anotarla: en headless,
  `esfuerzo:` es expresable en dos de tres. `sf` no lo puede inventar y no debe callarlo.
- **`-p` no significa lo mismo en los tres.** En opencode `-p` es `--password`; headless ahí es el
  subcomando `run`. Un `sf lanzar` que asuma `-p` en los tres se rompe en el primer intento.

---

## 3. Qué es `sf lanzar`

Un **modo extra**, no un reemplazo. Javier lo dijo antes de que existiera la spec:

> *"lo veo como un modo extra de sf, que él corra todo desde adentro"*

```bash
sf lanzar                 # corre el paso que dice `sf next`, acá, y espera
sf lanzar --async         # lo larga y devuelve un id; el bucle sigue
sf lanzar --seco          # imprime el comando que correría y no lo corre
```

`sf lanzar` no decide nada nuevo: le pregunta a `sf next` qué toca, resuelve el modelo con **la
misma cadena de precedencia de siempre** (§H2 de `agnostico-al-harness.md`), arma la línea de
comando desde el catálogo, y ejecuta.

**Lo que cambia no es la decisión: es quién aprieta el botón.**

### Y cierra un agujero que hoy está abierto

Hoy `via: subagente` es una instrucción **sin mecanismo**. El 2026-08-31 opencode leyó la skill,
entendió, e hizo el trabajo él mismo en vez de lanzar el subagente — y `sf done` aceptó el PRD sin
chistar, porque `compuerta.PRD` sólo comprueba que el archivo exista. `sf` **no tiene forma de saber
quién escribió un artefacto**. Se enteró porque el orquestador confesó.

Con `sf lanzar`, `via: subagente` deja de ser una sugerencia y pasa a ser un hecho. Es la misma
familia de agujero que el *"You do not rewrite the roadmap"* del ⑩: `sf` da instrucciones y confía.

---

## 4. Siete estados sí, dos no

**Headless sólo puede correr lo que no conversa.** Y eso ya está resuelto en el código: el mapa
`conversan` de `maquina.go`.

| | Estados | `sf lanzar` |
|---|---|---|
| conversan | ①–⑤ el brief · ⑧ la constitución | **no** — necesitan a Javier |
| no conversan | ⑦ · ⑨ · ⑩ · ⑫–⑯ · ⑱–⑳ · ㉑㉒ · ㉓ | sí |

No es una limitación del modo: es la misma regla que hizo que el ⑧ dejara de ser subagente el
2026-08-31 (ver `maquina-estados.md` §5.1 y el commit `4acf47f`). Un paso que necesita la opinión
de Javier no se puede delegar **a nadie**, ni a un subagente ni a un proceso.

> Y por eso `sf lanzar` **no es un `sf run` que corre el proyecto entero solo**. La máquina para
> donde siempre paró. Lo que se automatiza es el tramo entre dos paradas — que es exactamente el
> tramo que el horizonte de §5.1 ya anuncia.

---

## 5. El prompt — la parte que no es un flag

Acá está el trabajo de verdad, y es lo que hay que diseñar con cuidado.

Hoy el arnés carga la skill: `sf next` dice `skill: sfp-po` y el orquestador la encuentra en
`~/.claude/skills/` (o donde sea que ese arnés busque). Eso **ya trajo problemas medidos**: Command
Code no lee `~/.claude/skills/`, y hubo que escribirle `settings.skills` en `sf install`.

En headless no hay orquestador que busque nada. **`sf` tiene que mandar el prompt entero**:

```
prompt = el texto de la skill
       + el sobre (lo que hoy devuelve `sf context`)
       + qué se espera de vuelta
```

Y ahí desaparece otra clase entera de problemas: **el descubrimiento de skills deja de importar.**
No hace falta que el arnés encuentre `sfp-po`; `sf` la lee del disco y la manda. Se cae también la
mitad de `sf doctor` que hoy cuenta 9/9.

### El problema, dicho preciso: la skill es un menú, no la comida

Una skill no contiene el método entero. Contiene **punteros**:

```
sf-build/SKILL.md:82        Compose **sfx-tdd**. The cycle is unchanged: …
sfp-constitucion/SKILL.md   **Working rules** — see `references/reglas-de-trabajo.md`
```

Hoy eso funciona porque el que lee está **adentro de un arnés y puede abrir archivos**: ve el
puntero y va a buscarlo. En headless no hay quien lo resuelva, y **el que falla no avisa**: un
modelo que lee "Compose sfx-tdd" y no puede traerlo va a inventar algo parecido a TDD y devolver un
artefacto que parece bien. Es la misma familia de todo lo que mordió el 2026-08-31: silencioso.

### Medido el 2026-08-31 — el grafo es chico y no hay recursión

| | |
|---|---|
| skills que componen | 5 de 9 — `sfp-scout`, `sfp-backlog`, `sf-plan`, `sf-build`, `sf-cierre` |
| profundidad | **un nivel**: ninguna `sfx-` compone otra |
| archivos `references/` | 16, nombrados 17 veces, de 1,5 a 4,4 KB |
| el prompt más grande | `sf-build` = 9.036 B + 8.157 B compuestas ≈ **4.300 tokens** |

**El tamaño no es el problema.** Entra sobrado.

### Y "apuntar en vez de pegar" NO es una alternativa

La salida fácil sería no pegar nada: decirle al modelo *"la skill está en tal ruta, leela"*. Se
probó, y opencode la cierra:

```
$ opencode run --dir <proyecto> -m … "Leé /…/skills/sfx-tdd/SKILL.md y decime …"
  ! permission requested: external_directory (/…/skills/sfx-tdd/*); auto-rejecting
  ✗ Read … failed

$ opencode run --dir <proyecto> --auto -m … (mismo prompt)
  → Read /…/skills/sfx-tdd/SKILL.md
  Core Principle
```

O sea: **apuntar exige `--auto`**, una escalada de permisos, para algo tan inocente como leer un
`.md` propio. **Pegar el texto no necesita ningún permiso.** La decisión queda cerrada por medición
y no por gusto: **`sf` pega.**

### Lo que queda por decidir, entonces, es uno solo

**¿Cuándo viajan los `references/`?** Hoy el modelo decide si los abre — son carga perezosa, y la
mayoría de las veces no hacen falta. Pegarlos siempre engorda cada prompt con material que casi
nunca se usa; no pegarlos deja punteros colgando, que es justo el modo de fallar silencioso de
arriba.

Las dos salidas razonables:

| | |
|---|---|
| **pegar todo** | simple, sin sorpresas, ~2 a 6 KB de más por prompt |
| **pegar el que la skill nombra en el paso que toca** | más fino, pero exige que `sf` entienda la estructura interna de cada skill — y eso es acoplamiento nuevo |

Empezar por **pegar todo** y medir. Si duele, se afina; si no, no hay nada que afinar.

---

## 6. La ficha y el registro

Los tres arneses emiten **un stream de eventos JSON**. De ahí salen dos cosas distintas, y
confundirlas sería un error:

| | Qué es | Para quién |
|---|---|---|
| **el registro** | el stream completo, JSONL, tal cual salió | forense. Se guarda y no se lee salvo que algo salga mal |
| **la ficha** | un JSON chico: qué se lanzó, con qué modelo, cuánto tardó, cómo terminó | el bucle. Es lo que `sf next` mira |

**El registro nunca es evidencia.** Que un modelo diga "listo" en su stream no cierra nada: la
compuerta sigue siendo la compuerta. Es la misma regla que ya rige para el orquestador —*"Vos no
leés el trabajo. El subagente puede decir 'terminé' y estar equivocado."*

Con `--async`, `sf next` tiene que poder decir **⏳ hay algo corriendo** y no proponer trabajo
encima. Eso es un estado nuevo en la máquina, y es la parte de §6 que toca `estado.json`.

---

## 7. La compuerta de Command Code — la decisión que falta

**Es de Javier, no mía, y por eso está sola en su sección.**

Command Code en `-p` **no escribe archivos ni ejecuta shell**. Probado el 2026-08-31 con:

```
--permission-mode dont-ask       ✗
--permission-mode auto-accept    ✗
--permission-mode bypass         ✗
--tools-all                      ✗
-t / --trust                     ✗   (el más prometedor: "auto-trust project")
--yolo                           ✓   único que cede
```

Y `--yolo` es, en la ayuda del propio Command Code, **alias de
`--dangerously-skip-permissions`**, marcado en su doc para *"throwaway environments only"*.

Las opciones, sin recomendación de mi parte porque el riesgo lo corre él:

| | Qué implica |
|---|---|
| **A. `sf lanzar` nunca pasa `--yolo`** | Command Code queda fuera del modo headless. Los otros dos andan. Honesto y limitado. |
| **B. Lo pasa si Javier lo declara** | Un campo explícito en el catálogo (`permisos: yolo`) por arnés. `sf` no lo elige: lo transporta, igual que un id. Es R3 aplicado a los permisos. |
| **C. Lo pasa siempre** | No. `sf` estaría tomando por Javier una decisión de seguridad que su propio proveedor marca como peligrosa. |

**B es la única que cabe en el diseño existente** —el catálogo ya es el lugar donde vive lo que
Javier decidió y `sf` sólo acarrea— pero **A es una respuesta legítima** si prefiere no tener ese
flag escrito en ningún archivo suyo.

---

## 8. Lo que NO cambia

Para que quede dicho antes de que alguien lo "aproveche":

- **Las compuertas siguen en `sf`.** Headless no las relaja ni una.
- **Las paradas siguen siendo de Javier.** El ⑥, el ⑧ y el ⑰ no se automatizan.
- **`sf next` sigue siendo consulta pura.** Nunca escribe el estado. `sf lanzar` sí, y por eso son
  dos comandos y no uno.
- **R3 sigue vigente.** `sf` arma la línea de comando desde el catálogo; no elige un modelo parecido
  ni traduce un id de un proveedor a otro.
- **El modo orquestado no se borra.** Quien quiera seguir trabajando adentro del arnés, sigue.

---

## 9. Cómo se sabrá que quedó bien

```bash
# el mismo paso, el mismo estado en disco, tres ejecutores distintos
sf lanzar --seco                 # imprime la línea, no corre nada
sf lanzar                        # corre el ⑦ y deja .docs/prd.md
sf done                          # la compuerta pasa igual que si lo hubiera hecho un subagente
```

Y las tres pruebas que de verdad cierran el diseño:

1. **Sin `.opencode/agents/`.** Borrar los portamodelo y que `sf lanzar` funcione igual. Si hace
   falta alguno, el modelo no se está pasando por flag y estamos donde empezamos.
2. **Sin reiniciar nada.** Declarar un alias con `sf model` y lanzarlo **en la misma sesión**. Hoy
   eso es imposible —los agentes se leen al arrancar— y es la señal más limpia de que la causa
   desapareció.
3. **Con las skills desinstaladas.** Sacar `~/.claude/skills/` y que `sf lanzar` corra igual, porque
   el texto viaja en el prompt. Si falla, el prompt no está completo (§5).

> **La prueba de que el modo sirve no es que corra: es que el ⑦ salga bien con un modelo que no es
> de Anthropic, sin que nadie haya tenido que configurar el arnés.** Eso es lo que hoy cuesta una
> instalación por arnés, un archivo por alias y un reinicio.
