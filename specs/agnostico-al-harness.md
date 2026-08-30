# Agnóstico al harness — el parte de la primera corrida afuera

**Fecha:** 2026-08-29 · **Branch:** `refundation` · **Commit:** `2a44be3`

> **Estado: PENDIENTE — diseño cerrado, con las siete preguntas de harness medidas.** Esta es la primera corrida del bucle con agentes de verdad
> ([`salir-a-la-cancha.md`](salir-a-la-cancha.md) §③) y la primera fuera de Claude Code. Corrió en
> **Command Code** sobre `crowd-qa-pro` (Vite + React + TS + Supabase + Vitest), feature `f-1`,
> lote 1. No terminó, y lo que la frenó no fue la máquina: fue **la costura con el harness**.

Este documento continúa a [`arreglos.md`](arreglos.md) —el parte de daños del binario— y a
[`salir-a-la-cancha.md`](salir-a-la-cancha.md) —el de la instalación—. Éste es el de **la costura
entre `sf` y el harness que lo ejecuta**, que es la única de las tres que no se podía ver leyendo.

---

## La causa común, en una frase

> **`sf next` habla el vocabulario de un solo proveedor, y `sf install` instala en un solo harness.**

Las dos mitades son el mismo error con distinta cara. El diseño ya había separado bien las
preguntas —`tareas.json` dice **qué** hace falta, `~/.specforge/` dice **cómo se lanza acá**
([`global.go`](../sf/internal/global/global.go), H1b)— pero después metió un nombre de Anthropic
del lado del **qué**. Y ahí `~/.specforge/` deja de poder traducir: no hay nada que traducir
cuando el que pregunta ya dijo la respuesta.

Lo mismo del otro lado: `sf install` escribe `CLAUDE.md` y `AGENTS.md` —que es correcto y funciona
en los tres harness, ver §7— pero **no escribe nada de lo que el harness necesita para no
preguntar**. En Claude Code eso nunca se notó porque un subagente hereda el modo de permisos del
padre. En los otros dos no lo hereda, y ahí el bucle se traba en el primer subagente.

---

## Resumen

| # | Qué | Severidad | Bloquea |
|---|---|---|---|
| **H1** | El ⑯ recomienda `opus` siempre, porque es el único ejemplo que ve | 🟠 alto | la elección de modelo entera |
| **H2** | `sf` devuelve un nombre de proveedor donde tenía que devolver un rol | 🔴 **crítico** | correr en cualquier harness que no sea Claude Code |
| **H3** | `sf install` no escribe permisos: los subagentes se traban pidiendo por todo | 🔴 **crítico** | la corrida entera fuera de Claude Code |
| **H4** | Sólo se detecta y se busca en Claude Code | 🟠 alto | `sf doctor`, y la resolución de modelo de H2 |
| **H5** | El ⑯ puede nombrar tests en archivos que no son tests, y nadie lo mira | 🟠 alto | `sf lote start` en cualquier proyecto con no-código |
| **H6** | La constitución no declara qué necesita el `test_cmd` para poder correr | 🟠 alto | los lotes de schema y DB — **y, con `dont-ask`, la suite entera** |
| **H7** | El ⑯ elige el modelo a ciegas: el sobre no lleva el menú | 🟠 alto | que el routing por fase sirva de algo |
| **H8** | El `modelo:` de `sf next` es inaccionable fuera de Claude Code | 🔴 **crítico** | el routing por fase, entero |

**Orden propuesto:** H8 primero —sin él, H2 y H7 son decoración— · H2 + H1 + H7 (son la misma
pieza) · H4 (lo necesita H2) · H3 · H5 · H6.

---

## Los tres síntomas, tal como aparecieron

Vale la pena dejarlos crudos antes de las causas, porque **dos de los tres no son lo que
parecían.**

1. *"El gate de `sf lote start` exige nombres con prefijo del path del SQL, no del `.test.ts` que
   los ejecuta."* Los 11 tests del lote salieron reportados como faltantes.
2. *"No hay Postgres en el host."* Los tests del plan apuntaban al `.sql`, así que ejecutarlos de
   verdad pedía `pg` como devDep + Docker, y la constitución no dice nada de eso.
3. *"`vitest` con `environment: jsdom` no expone `node:fs`."* Se resolvió con
   `// @vitest-environment node`, que ya estaba en el taste del proyecto.

Y el cuarto, el que Javier reportó de las corridas previas y **no se pudo reproducir en ésta**:
*"siempre colocaba que debía usarse opus"*. No apareció porque la corrida murió antes, en el ⑰.
Pero la causa está en el código y es H1: no hacía falta reproducirlo, hacía falta leerlo.

---

## Lo que se midió, y lo que se leyó

Este documento arrancó leyendo documentación y **se equivocó dos veces**. Las dos se corrigieron
ejecutando, con sondas de veinte minutos en los harness de verdad. Vale dejar la separación a la
vista, porque es la diferencia entre una spec y una hipótesis:

| # | Pregunta | Cómo se contestó | Resultado |
|---|---|---|---|
| 1 | ¿el `model:` del archivo del agente se aplica? | **sonda 1** — un agente con un id inexistente | ✅ los dos dieron *modelo desconocido* |
| 2 | ¿se puede invocar un subagente por nombre? | **sonda 1** | ✅ los dos |
| 3 | ¿se releen los agentes en caliente? | **sonda 1** — agregar uno con la sesión abierta | ❌ los dos lo vieron como un archivo |
| 4 | ¿los subagentes leen `~/.claude/skills/`? | **sonda 2** — tokens aleatorios | opencode ✅ · Command Code ❌ |
| 5 | ¿leen `.agents/skills/` del proyecto? | **sonda 2** | ✅ los dos |
| 6 | ¿preguntan permisos sin config? | **sonda 2** | opencode ✅ corre solo · Command Code ❌ pregunta |
| 7 | ¿`settings.skills` arregla el descubrimiento? | **sonda 2, ronda B** | ✅ Command Code dijo los dos tokens |

**Los dos errores, anotados a propósito.** Que queden es más útil que borrarlos:

- **decía que opencode se trababa pidiendo permisos y Command Code no.** Es al revés. Salió de
  leer un ejemplo de config como si fuera el default.
- **proponía `auto-accept` para Command Code.** Su propia doc aclara que ese modo sigue preguntando
  por comandos de shell arbitrarios — que es todo lo que el bucle hace.

Lo que **no** está medido y sigue siendo lectura: la fila de Claude Code de §7 (que sí está probada
por uso diario), y H5 y H6, que son de la máquina y no del harness.

---

## H1 — El ⑯ recomienda `opus` siempre, porque es el único ejemplo que ve

### Síntoma

`tareas.json` sale con `"modelo": "opus"` en features que no lo necesitan. Como
`modeloDeFeature` (`maquina.go:713`) le da al ⑯ prioridad sobre el default del estado, **cada
lote de esa feature se lanza en el modelo más caro**, incluidos los mecánicos.

### Causa raíz

`skills/sf-plan/SKILL.md:175`:

```json
{"feature": "f-1", "modelo": "opus", "tareas": [ … ]}
```

El texto de al lado dice lo correcto —*"Leave it out otherwise. The default is the default because
it is right almost always"*— pero **el único ejemplo del campo es el caso excepcional**. Un
planificador que duda copia el ejemplo, no la advertencia. El template
(`templates/tareas.tmpl.json`) hace lo correcto y no trae el campo; el que filtra es el SKILL.

Y no está solo: `skills/sfx-github/SKILL.md:13` dice `Model: haiku.` — un nombre de proveedor
adentro de un artefacto que se declara agnóstico.

### Arreglo

Con H2 y H7 resueltos, el ejemplo deja de nombrar un modelo suelto y pasa a mostrar lo que el ⑯
de verdad hace: **elegir del menú, por lote, con un motivo.**

```json
{"feature": "f-1",
 "modelo": "sonnet",
 "modelo_por_lote": {"1": "laguna", "3": "opus"},
 "tareas": [ … ]}
```

y el texto gana las dos frases que faltan:

- **el ⑯ nunca escribe un id de modelo, escribe un alias del menú.** Un id en `tareas.json` es un
  dato de la máquina de Javier metido en un archivo versionado que va a leer otro harness;
- **el default es no decir nada.** El menú invita a elegir, y un planificador que elige en los
  cinco lotes de una feature trivial es el mismo ruido de antes con más pasos.

El ejemplo tiene que mostrar los dos casos —un lote barato y uno caro— justamente porque **un solo
ejemplo se copia**. Ésa fue la causa raíz de este hallazgo, y un ejemplo único con un alias adentro
la reproduce igual.

### El test que tiene que existir

Un chequeo de CI —de los cuatro que ya hay— que **falle si un `SKILL.md` menciona un nombre de
modelo concreto**. La lista negra es corta y no crece sola: `opus`, `sonnet`, `haiku`, `gpt-`,
`deepseek`, `grok`, `gemini`. Es exactamente la misma clase de chequeo que el de `name:` igual a
la carpeta: barato, mecánico, y protege una regla que un humano no va a recordar.

Y cubre a los skills igual que a `sf/`: los alias del menú son de la máquina de Javier, así que un
`SKILL.md` que nombre uno está haciendo lo mismo que hacía con `opus`.

---

## H2 — `sf` devuelve un nombre de proveedor donde tenía que devolver un rol

### Síntoma

En opencode y en Command Code, `sf next` devuelve `modelo: opus` y **no significa nada**. opencode
espera `provider/model` (`anthropic/claude-opus-4-1`); Command Code espera un id de su propio
catálogo de 60+ modelos. El harness ignora el campo y corre con el modelo de la sesión — o sea que
el routing por fase, que es el diferenciador entero de `sf`, **no existe fuera de Claude Code**.

### Causa raíz

Tres lugares dicen `opus`, y ninguno tenía derecho a decirlo:

| Archivo | Línea | Qué |
|---|---|---|
| `sf/internal/maquina/maquina.go` | 208–213 | `modeloPorEstado{planificacion:"opus", revision:"opus"}` · `modeloPorDefecto = "sonnet"` |
| `sf/internal/global/global.go` | `nativos` | `[]string{"opus","sonnet","haiku"}`, sembrados **igual en cualquier harness** |
| `skills/sf-plan/SKILL.md` | 175 | H1 |

El comentario de `modeloPorEstado` describe bien la arquitectura —*"Este mapa dice QUÉ hace falta;
el otro dice cómo se invoca en esta máquina"*— pero el mapa **no dice qué hace falta: dice quién lo
hace**. "Opus" no es una necesidad, es una respuesta.

Y `Semilla()` remata: siembra los tres nativos con `via: subagente` sin mirar el harness. En
opencode eso declara tres modelos que no existen, y peor: **desactiva la única compuerta que había**
—la 🛑 de `sinDeclarar`— porque `Buscar("opus")` contesta que sí.

### El arreglo: perfiles, indexados por harness, y el YAML se llena a mano

`sf` deja de nombrar modelos y nombra **roles**. Son tres, y no cuatro ni siete, porque son los que
la máquina puede distinguir con lo que sabe:

| Perfil | Qué pide | Quién lo usa |
|---|---|---|
| `razonar` | juicio, comparación, hallar lo que no está | ⑫–⑯ `planificacion` · ㉑㉒ `revision` |
| `construir` | escribir código contra un plan que ya existe | el default: ⑥–⑩, ⑱–⑳, ㉓ |
| `mecanico` | ejecutar un procedimiento sin decidir nada | los `sfx-*` (`sfx-github`) — **fuera de la máquina** |

`~/.specforge/modelos.yaml` pasa a mapear **perfil → cómo se lanza acá**, que es lo que ese archivo
siempre quiso ser — con los perfiles **abajo de cada harness**:

```yaml
# ~/.specforge/modelos.yaml
harness: opencode          # cuál está activo — lo mueve `sf install --harness=…`

harnesses:
  opencode:
    razonar:
      - {alias: opus, id: anthropic/claude-opus-4-1, via: subagente, capacidad: 3,
         para: "lo que hay que pensar de verdad: diseño, concurrencia, hallar lo que falta"}
      - {alias: r1,   id: deepseek/deepseek-r1,      via: subagente, capacidad: 2,
         para: "razona bien y sale barato; la prosa le sale peor"}
    construir:
      - {alias: sonnet, id: anthropic/claude-sonnet-4-5, via: subagente, capacidad: 3,
         para: "el caballo de batalla; si dudás, éste"}
      - {alias: laguna, id: laguna/s2.1,                 via: subagente, capacidad: 1,
         para: "rápido y barato; bien en CRUD, plomería y refactors mecánicos"}

  claude-code:
    razonar:
      - {alias: opus,   id: opus,   via: subagente, capacidad: 3, para: "…"}
    construir:
      - {alias: sonnet, id: sonnet, via: subagente, capacidad: 3, para: "…"}
      - {alias: haiku,  id: haiku,  via: subagente, capacidad: 1, para: "…"}

# Los sueltos: el ⑳ de Javier, y los ajenos que salen por consola.
modelos:
  deepseek: {id: deepseek-reasoner, via: consola, comando: "deepseek exec"}
```

**Las tres capas, y cada una es de alguien distinto.** Ésta es la pieza que hace que todo lo
demás cierre:

| Capa | Vocabulario de | Vive en | Ejemplo |
|---|---|---|---|
| **perfil** | `sf` | el binario | `construir` |
| **alias** | **Javier** | `modelos.yaml` **y `tareas.json`** | `laguna` |
| **id** | el harness | `modelos.yaml`, y sólo ahí | `laguna/s2.1` |

El alias es la capa que faltaba, y es la que resuelve la tensión entera: **el ⑯ escribe un alias,
no un id.** Un id en `tareas.json` sería un dato de la máquina de Javier metido en un archivo
versionado que otro harness va a leer — que es exactamente H2. Un alias no: `laguna` está definido
en los dos bloques apuntando a ids distintos, así que **el plan sobrevive al cambio de harness**,
que es el escenario real.

y `sf next` devuelve **las dos cosas**:

```
estado:  planificacion    skill: sf-plan
perfil:  razonar
modelo:  anthropic/claude-opus-4-1
via:     subagente
```

El `perfil:` es para el que lee y para el que decide; el `modelo:` es para el que lanza. Que viajen
juntos es lo que hace que un `sf next` sirva de log: dentro de tres meses se puede ver **qué pedía
el paso** y **qué se le dio**, que hoy son la misma palabra y no se pueden distinguir.

### Por qué indexado por harness, y no un solo bloque de perfiles

Porque **una máquina tiene varios harness instalados a la vez**, y ése es el caso normal y no el
raro: la corrida que originó este documento pasó por Claude Code y por Command Code en la misma
semana, en la misma máquina.

Con un solo bloque, `sf install --harness=opencode` te deja `razonar: opus` —que en opencode no
existe— y al volver a Claude Code te deja `anthropic/claude-opus-4-1`, que ahí tampoco. **El
archivo no sobrevive a que cambies de harness**, y la forma de descubrirlo es que el subagente se
lance con el modelo equivocado sin decir nada.

Indexado, cada bloque se llena una vez y **cambiar de harness no destruye nada**. `harness:` arriba
deja de ser un dato de configuración y pasa a ser lo que siempre fue: un puntero.

### Y esto es lo que mata el "siempre Anthropic" de verdad

Un solo id por perfil, elegido por `sf`, siempre iba a ser el id de alguien. Con los perfiles
indexados, **la mezcla de proveedores es de Javier y es por harness**: `construir` puede ser GPT en
opencode y `sonnet` en Claude Code, y `sf` no tiene opinión sobre ninguno de los dos — sólo sabe
que el ⑱ pide *construir*.

### Por qué varios modelos por perfil, y quién elige

Un perfil con **un** modelo alcanza para que `sf` no hable Anthropic, y no alcanza para nada más:
deja al ⑯ sin nada que decidir. Y el ⑯ **es** el que decide — es el nivel 2 de la cadena de
precedencia y su única función en esa línea es ésa.

> **El chooser existía y estaba a la vista: es el ⑯.** Es un modelo, lee el sobre, y elegir con qué
> se implementa cada lote es literalmente su paso. Hoy elige a ciegas —escribe `opus` porque es la
> única palabra que conoce (H1)—, y darle el menú es lo que convierte esa elección en una decisión.
> Eso es H7.

Lo que **no** se hace es que `sf` elija de la lista. La lista está ordenada y cada entrada trae un
`capacidad:` y un `para:`, y los tres son **opinión de Javier escrita a mano**. `sf` no los
interpreta, no los compara y no hace aritmética con ellos: los **transporta** al sobre y listo. La
diferencia importa y es la línea de R3:

> **`sf` no puede tener una opinión sobre qué modelo se parece a cuál. Sí puede llevar la de Javier
> hasta quien la necesita.** Es la misma distinción que ya rige en `dependencias_aprobadas`: la
> lista es de Javier, `sf` sólo compara contra ella.

Por eso `capacidad:` es un número **para los ojos del ⑯**, no un campo de cómputo. Si `sf` eligiera
"el más barato con capacidad ≥ 2" estaría eligiendo modelo, y volveríamos a lo mismo con más pasos.

> **Lo que sigue descartado: el fallback automático.** "Si `razonar` no está disponible, usá el que
> sigue" es otra feature — `sf` no lanza modelos, así que reintentar sería enseñarle al orquestador
> un protocolo de reintento y hacer que `sf next` deje de ser determinista. No entra acá.

### Cómo se resuelve, y qué pasa cuando el alias no está

**El primero de cada lista es el default de ese perfil.** El orden es la preferencia, y no hay un
`default: true` que pueda contradecirlo: un solo hecho, en un solo lugar.

```
1. sf model <alias|perfil>     estado.json — la decisión de Javier en runtime
2. tareas.json, del lote        el ⑯ — "el lote 3 necesita otro"
3. tareas.json, de la feature   el ⑯ — "toda esta feature necesita otro"
4. perfilPorEstado              el criterio del diseño
5. perfilPorDefecto             construir
```

Los niveles 2 y 3 son **un solo rung partido en dos**, no uno nuevo: siguen siendo "lo que el ⑯
recomendó", con lo más específico arriba. La regla de siempre —cada nivel sabe menos que el de
arriba— se mantiene.

**Un alias que no existe en este harness NO es una 🛑.** Cae al default del perfil y lo dice:

```
⚠ el lote 3 pide `laguna` y `opencode` no lo tiene declarado — va con `sonnet`
  (el default de `construir`). Declararlo: sf model laguna --id … --via subagente
```

Es aviso y no parada, y la asimetría con la 🛑 del perfil vacío es deliberada: **un perfil sin
declarar no tiene salida —`sf` no sabe con qué lanzar nada—, mientras que un alias faltante sí la
tiene, y además cae para el lado seguro**: el default de un perfil es el más capaz de su lista, así
que el peor caso es gastar de más. Frenar el bucle por eso sería frenar sobre una preferencia.

### La parte que `sf` NO hace, y es la importante

**`sf` no trae una tabla de ids, ni para un harness.** Ni una. Eso sería opinar sobre qué modelo se
parece a cuál en un catálogo ajeno que cambia cada dos meses — R3 lo prohíbe, y es la misma regla
que ya existe para `dependencias_aprobadas`.

**No hay siembra, y no hay excepción para nadie.** La versión anterior de este documento le daba a
`claude-code` un pase —sembrar `opus`/`sonnet`/`haiku` porque son alias del propio harness y no ids
de proveedor— y el argumento se sostenía, pero la consecuencia no: dejaba el camino feliz con forma
de Anthropic y a todos los demás con una 🛑. **`sf install` escribe el bloque del harness vacío y
comentado, para todos por igual.**

Lo que pasa después es lo que la máquina ya sabía hacer: **para y pregunta.** La 🛑 de
`sinDeclarar` (`maquina.go:755`) está construida y no hay que inventarla, sólo hay que dejarla
llegar:

```
🛑 PARÁ. El paso pide el perfil "razonar" y el harness `opencode` no lo tiene declarado.
   sf no elige el modelo: decime vos cuál es acá.
   (los ids de tu harness: `opencode models` · `/model` en Claude Code y Command Code)

   sf model razonar --alias opus --id anthropic/claude-opus-4-1 --via subagente
   sf model razonar --alias r1   --id deepseek/deepseek-r1 --via subagente \
                    --capacidad 2 --para "razona bien y sale barato"
```

Cada `sf model` **agrega** al final de la lista del perfil, que es lo correcto: el primero es el
default, y lo primero que declarás es lo que más vas a usar. Reordenar es editar el YAML a mano,
y está bien que lo sea — es un archivo de Javier, no una base de datos.

**Dos preguntas, una vez por harness, para los dieciséis repos.** Dos y no tres: `mecanico` no lo
devuelve la máquina nunca —sólo lo nombran los `sfx-*`, que están fuera de los nueve estados—, así
que su entrada es **opcional y no frena**. Es el mismo mecanismo del mapa que se construye solo,
aplicado a la única cosa que `sf` no puede saber.

> **El invariante que esto compra, y que se puede testear:**
> `grep -rn "opus\|sonnet\|haiku" sf/` devuelve **cero**. No un default razonable, no una excepción
> documentada: cero. Es un chequeo de CI de una línea, y es la única forma de que la regla no se
> erosione la próxima vez que alguien tenga apuro.

### El precio, dicho de frente

Hoy un `sf install && sf init && sf next` en Claude Code corre sin preguntar nada. Con esto, para
en el ⑫ y pide dos ids. **Es un paso más en el primer arranque de cada harness**, y es el único
costo de todo el rediseño.

Y es el piso, no el techo: con dos entradas —una por perfil— el bucle corre igual que hoy, sólo que
sin nombres de proveedor adentro del binario. El menú de varios modelos (H7) es **opt-in**: se llena
el día que quieras que el ⑯ pueda mandar un lote de plomería a un modelo barato, y hasta ese día no
existe.

Se paga porque la alternativa —sembrar los alias de Claude Code— es la que produjo este documento:
un default de un proveedor, copiado a un skill, copiado a `tareas.json`, ejecutado en un harness
que no lo entiende. **Dos preguntas contestadas una vez valen menos que eso.**

Lo que sí hay que hacer es que la 🛑 sea *útil*: nombra el perfil, nombra el harness, y dice **cómo
averiguar los ids en ese harness**. Una parada que te deja sin saber qué contestar es una parada
mal escrita.

### Compatibilidad hacia atrás

Un `modelos.yaml` viejo tiene `harness:` y `modelos:` con los tres nativos, y no tiene `harnesses:`.
Al leerlo, `sf` los mueve al bloque del harness que el archivo declara, cada uno como **único
elemento de la lista de su perfil**, con `alias` e `id` iguales al nombre viejo:

```yaml
harnesses:
  claude-code:
    razonar:   [{alias: opus,   id: opus,   via: subagente}]
    construir: [{alias: sonnet, id: sonnet, via: subagente}]
    mecanico:  [{alias: haiku,  id: haiku,  via: subagente}]
```

Sin `capacidad:` ni `para:`, que son opcionales — un menú de un elemento no necesita describirse.
**Nadie tiene que borrar nada, y el `sf model deepseek` que alguien haya aprobado sigue vivo**:
para eso `modelos:` no se va, es donde viven los sueltos, que no son de ningún harness en
particular.

La migración se escribe una vez al leer y se persiste al primer `Guardar()`. No hay comando de
migración y no debería haberlo: un archivo que se arregla solo la primera vez que se lo toca es
mejor que uno que exige acordarse de un comando.

La cadena de precedencia es la de `modeloDeFeature` con el nivel 2 partido en dos, y está escrita
arriba, en **Cómo se resuelve**.

### Los tests que tienen que existir

- `Siguiente` en `planificacion` devuelve `perfil: razonar` y el `id` que declara el bloque del
  harness activo — **no** la palabra `opus`.
- El mismo `modelos.yaml` con dos bloques (`claude-code` y `opencode`) devuelve **ids distintos**
  para el mismo perfil según qué diga `harness:`.
- `sf install --harness=commandcode` sobre un YAML que ya tiene el bloque de `opencode` **conserva
  ese bloque entero** y sólo mueve el puntero.
- Con el bloque del harness vacío, el primer `sf next` de `planificacion` devuelve `Tipo: Para`, y
  el mensaje nombra el perfil **y** el harness.
- Con `razonar` y `construir` declarados y `mecanico` ausente, **ningún** `sf next` de los nueve
  estados para.
- Un `modelos.yaml` legacy (sólo `modelos:`) resuelve los tres perfiles bajo su `harness:` sin que
  nadie corra nada, y al guardarse queda migrado.
- `sf model razonar --id X --via subagente` escribe el perfil **en el bloque del harness activo**,
  no en `modelos:` ni en otro harness.
- El **primero** de la lista de un perfil es el que sale cuando nadie pidió nada.
- Un `tareas.json` que nombra un alias declarado devuelve **ese** id, no el del default.
- Un `tareas.json` que nombra un alias **no** declarado devuelve el default del perfil, `Tipo:
  Trabajar` (no `Para`), y un aviso que nombra el alias que faltó.
- El alias del **lote** le gana al de la feature, y el `sf model` de Javier le gana a los dos.
- Ningún archivo bajo `sf/` menciona `opus`, `sonnet` ni `haiku` — ni siquiera `global.go`.

---

## H3 — `sf install` no escribe permisos, y el bucle se traba en el primer subagente

### Síntoma

*"Cuando se lanzaban los subagentes siempre cortaban porque pedían permisos por todo."*

### Causa raíz

`andamio.Instalar` hace exactamente dos cosas: escribe el orquestador en los dos nombres, y arma
`~/.specforge/`. **No escribe configuración de permisos de ningún harness, porque nunca hizo falta:**
en Claude Code el subagente hereda el modo del padre, y Javier corre el padre en un modo permisivo.

Fuera de ahí — y esta tabla es **medida**, no deducida (sonda 2, 29/08):

| Harness | Config que había | Qué hizo el subagente |
|---|---|---|
| **Claude Code** | hereda del padre | nunca se vio |
| **opencode** | **ninguna** — el `opencode.json` global no tiene bloque `permission` | ✅ **corrió todo sin preguntar**: shell y escritura de archivos |
| **Command Code** | **ninguna** — `~/.commandcode/settings.json` no existía, o sea baseline `default` = *"preguntá antes de cualquier cosa que cambie algo"* | ❌ **pidió confirmación** |

**La deducción de la primera versión de este documento estaba al revés**, y vale dejarlo escrito:
decía que opencode era el que se trababa y que Command Code dependía del baseline. Es exactamente
al revés. Lo primero salió de leer que su doc *muestra* `{"*": "ask"}` en un ejemplo —pero eso es
lo que ponés, no lo que trae— y lo segundo, de leer que sus subagentes auto-aprueban donde un
humano vería prompt, sin notar que el baseline `default` es más fuerte que esa regla.

El bucle de `sf` es **desatendido por diseño**: el orquestador lanza, el subagente pide su sobre,
trabaja y cierra con `sf done`. Un prompt de permisos en el medio no es una molestia — **rompe la
premisa**. Y sin embargo `sf install`, que existe justamente para "hacer que SpecForge se pueda
usar", no toca la única cosa que lo hace usable.

### Arreglo

`sf install --harness=X` escribe la config de permisos de X, con la **misma regla que ya usa para
`CLAUDE.md`**: si el archivo existe no se pisa, y `--forzar` lo pisa. Si existe pero le falta lo de
`sf`, se **mergea** la parte de `sf` y se deja el resto intacto — la config de permisos de un
proyecto tiene cosas de Javier igual que el `CLAUDE.md`.

Lo que se otorga es **lo que `sf` mismo necesita, y nada más**:

```jsonc
// claude-code — .claude/settings.json
{"permissions": {"allow": ["Bash(sf:*)", "Bash(git:*)"]}}
```

```jsonc
// opencode — opencode.json   ·   NO HACE FALTA para que el bucle corra.
// Corre permisivo por default (medido). Esto es un cinturón opcional: acota
// lo que un subagente puede hacer, no lo habilita.
{"$schema": "https://opencode.ai/config.json",
 "permission": {"edit": "allow",
                "bash": {"*": "allow", "rm -rf *": "deny"}}}
```

```jsonc
// commandcode — .commandcode/settings.json   ·   ESTE SÍ HACE FALTA.
{"permissions": {"defaultMode": "dont-ask",
                 "allow": ["Shell(sf:*)", "Shell(git:*)"]},
 "skills": ["~/.claude/skills"]}
```

### Por qué `dont-ask` y no `auto-accept`

Porque **`auto-accept` sigue preguntando por comandos de shell arbitrarios**, y eso lo dice su
propia documentación:

> `auto-accept` — *"Do normal edits (and safe file commands) without asking"* … pero sigue
> preguntando por borrados recursivos, escrituras sensibles **y comandos de shell arbitrarios**.

La primera versión de este documento proponía `auto-accept`, y la sonda 2 lo desmintió: siguió
pidiendo confirmación. Los cinco baselines, y por qué sólo uno sirve acá:

| Modo | Qué hace | ¿Sirve para el bucle? |
|---|---|---|
| `default` | pregunta antes de todo lo que cambia | ❌ es lo que te trabó |
| `auto-accept` | edita sin preguntar, **pero pregunta por shell** | ❌ el bucle corre `sf`, `git` y los tests |
| `plan` | sólo lectura | ❌ |
| `bypass` / `--yolo` | hace todo | ❌ su doc dice *"throwaway environments only"*, y `sf` no le va a pedir eso a nadie sobre su repo |
| **`dont-ask`** | **nunca pregunta; corre lo pre-aprobado y deniega el resto** | ✅ |

**Un bucle desatendido no puede colgarse esperando un enter.** `dont-ask` convierte ese cuelgue en
un fallo — el comando se deniega, `sf done` ve el ✗, y el que trabaja lo reporta. **Fallar es un
hecho; colgarse no es nada.** Es la misma vara que rige todas las compuertas de este proyecto.

> **El precio, y es real:** con `dont-ask`, lo que falte en la lista **no pregunta, falla.** O sea
> que la lista tiene que estar completa, y eso empuja lo del `test_cmd` de abajo de "aviso" a
> "hace falta de verdad". La contrapartida es que falla **la primera vez y con un mensaje**, en vez
> de colgarse en el lote 3 a las dos de la mañana.

### Lo que `sf` NO otorga, y por qué

**El `test_cmd` no entra acá, y no es un olvido.** `sf install` corre **antes** de `sf init`, así
que en ese momento no hay constitución y no se sabe con qué se corren los tests. Pero además:
autorizar `npm *` o `pytest *` en nombre de Javier es otorgar permiso sobre un comando que él no
escribió y que `sf` no leyó.

La respuesta es la que el diseño ya tiene para esta clase de cosa: **`sf` frena sobre hechos y avisa
sobre todo lo demás.** `sf doctor` gana un chequeo:

```
permisos  ⚠ el `test_cmd` de la constitución (`npm test`) no está permitido en opencode.json
             agregá:  "npm *": "allow"   dentro de  permission.bash
```

Es un **aviso**, no un ✗: el bucle puede correr y trabarse una vez, y Javier decide. Frenar la
instalación porque falta un permiso que sólo él puede otorgar sería frenar sobre una opinión.

### Los tests que tienen que existir

- `Instalar` con `--harness=opencode` en un directorio limpio escribe `opencode.json` con el bloque
  `permission`, y el resultado lo lista en `Escritos`.
- Con un `opencode.json` que ya existe y tiene otras claves: las conserva **todas** y agrega sólo
  las de `sf`.
- Con `--harness=claude-code` **no** escribe `opencode.json` ni `.commandcode/`.
- `doctor` con un `test_cmd` que no está en la allowlist devuelve un aviso y **exit 0**, no 2.

---

## H4 — Sólo se detecta y se busca en Claude Code

### Síntoma

`sf doctor` reporta `harness: desconocido` corriendo dentro de Command Code. En la corrida no
apareció nada del harness en la salida del gate, así que el aviso no llegó a ningún lado.

### Causa raíz

Dos funciones, las dos con la lista de un solo harness:

- `global.DetectarHarness()` mira `CLAUDECODE` y `CLAUDE_CODE_ENTRYPOINT`, y si no, `desconocido`.
- `doctor.raicesDeSkills()` (`doctor.go:215`) busca en cuatro rutas y las cuatro son `~/.claude/`
  o `<proyecto>/.claude/`.

La segunda tiene una consecuencia que no es obvia y que hay que **verificar en la próxima corrida**:

Y acá está el hallazgo más limpio de la sonda 2, porque **estaba predicho antes de correrla** y
salió tal cual. Dos skills de sonda con tokens aleatorios de 16 hex —imposibles de adivinar—, una
sólo en `~/.claude/skills/` y otra sólo en `<proyecto>/.agents/skills/`:

| Harness | `~/.claude/skills/` | `.agents/skills/` | Cómo encuentra los 18 |
|---|---|---|---|
| **Claude Code** | ✅ | — | como hoy |
| **opencode** | ✅ **dijo el token** | ✅ **dijo el token** | como hoy, sin tocar nada |
| **Command Code** | ❌ **la skill no existía para él** | ✅ **dijo el token** | ver abajo |

**En Command Code los 18 skills de hoy no se encuentran.** Son symlinks en `~/.claude/skills/` y
esa raíz no la lee. O sea que la corrida de `crowd-qa-pro` que originó este documento llegó al ⑰
**sin los skills instalados** — algo los suplió, y lo más probable es que el material haya entrado
por el prompt sin que nadie lo notara.

> **Eso reencuadra la corrida entera.** Que el bucle haya llegado al ⑰ y que la compuerta lo haya
> frenado por el motivo correcto, **con el harness a medio instalar**, dice más a favor de la
> máquina que si hubiera estado todo bien puesto.

**Y el arreglo se verificó, no se propuso.** Con `"skills": ["~/.claude/skills"]` en
`.commandcode/settings.json` y reiniciando, Command Code dijo **los dos tokens**. Una línea de
configuración, y el descubrimiento queda resuelto sin copiar ni symlinkear nada a una segunda raíz.

> **`.agents/skills/` es la raíz que leen opencode y Command Code**, pero no Claude Code. No existe
> una sola que lean los tres, y por eso el camino es la clave `skills` de la config y no mudar los
> archivos: mudarlos arreglaría dos harness y rompería el tercero.

### Arreglo

- `DetectarHarness` aprende los otros dos por variable de entorno, y **sigue siendo una pista**: el
  comentario de `global.go` ya explica por qué se escribe una vez y no se recalcula, y eso no cambia.
- `raicesDeSkills` gana las rutas de opencode y de Command Code. Que `sf doctor` imprima **de dónde
  salió cada skill** deja de ser un lujo y pasa a ser el diagnóstico principal.
- `sf install --harness=commandcode` agrega `~/.claude/skills` al array `skills` de
  `.commandcode/settings.json` — el mismo archivo de H3, la misma pasada.
- **`harness: desconocido` pasa a ser 🛑 y no ⚠.** Hoy es un aviso porque no bloqueaba nada; con H2
  bloquea todo, porque sin harness no hay perfil que resolver. El mensaje ya está escrito:
  `sf install --harness=<nombre>` lo fija.

---

## H5 — El ⑯ puede nombrar tests en archivos que no son tests, y nadie lo mira

### Síntoma

`tareas.json` traía los 11 tests del lote como `supabase/migrations/<archivo>.sql::<nombre>`. El
implementador los escribió en `src/test/migration-f1.test.ts` con esos nombres exactos, y
`sf lote start` reportó **los 11 como faltantes**.

### Causa raíz

**La compuerta hizo lo correcto.** `suite.Faltantes` busca el nombre adentro del archivo que el plan
declaró, y no parsea la salida del runner a propósito —está argumentado en el encabezado de
`suite/suite.go`: cada runner imprime distinto y un parser por runner es una lista que se pudre—.
Si el plan dice `.sql`, mira el `.sql`.

El error está arriba, en el ⑯. `sf-plan/SKILL.md:105` dice:

```
"tests": ["path/to/file_test.go::TestName"]
```

y **nunca dice que la ruta es dónde vive el test, no qué cosa testea.** Frente a una migración SQL
—donde el sujeto es un archivo y el test es otro— el planificador escribió la ruta del sujeto. Es
la lectura natural de un ejemplo donde las dos cosas coinciden.

Y no lo atrapa nadie: la compuerta del ⑰ cuenta tests por criterio (`compuerta.go:333`) pero no
mira **dónde** dicen que van a estar.

### Arreglo — dos mitades, y las dos hacen falta

**En el skill**, la regla explícita que falta:

> La ruta es **el archivo donde va a vivir el test**, no el archivo que el test prueba. Tiene que
> ser un archivo que corra tu `test_cmd`. Si estás testeando una migración, un schema o un `.sql`,
> la ruta es la del test que lo lee o lo ejecuta — nunca la del `.sql`.

**En la compuerta del ⑰**, un chequeo barato y sin juicio: **el archivo de test declarado tiene que
tener la extensión de algún archivo de test que ya exista en el proyecto.** No inventa una tabla de
lenguajes, no adivina el runner: mira lo que el repo ya tiene. En un proyecto sin ningún test
todavía no dice nada — no puede, y callarse es lo correcto.

```
✗ t-3 nombra tests en `supabase/migrations/001_init.sql`, y en este proyecto
  los tests son `.ts`. La ruta es dónde vive el test, no qué cosa prueba.
```

### Los tests que tienen que existir

- Con `.test.ts` en el repo, un `tareas.json` que nombra tests en un `.sql` **no pasa el ⑰**.
- El mismo `tareas.json` en un repo **sin** archivos de test pasa, sin aviso.
- Un `.sql` declarado como archivo entero (sin `::nombre`) recibe el mismo trato: la extensión es lo
  que se mira, no el sufijo.

---

## H6 — La constitución no declara qué necesita el `test_cmd` para poder correr

### Síntoma

Los tests del lote apuntaban al `.sql`, o sea que ejecutarlos de verdad pedía un Postgres. **No hay
Postgres en el host** —ni `psql`, ni `pg_isready`, ni el servicio— así que la salida habría sido
`pg` como devDep + Docker. Se resolvió parseando el `.sql` como texto, que funciona, pero no es lo
que el plan quiso.

Y el primo del mismo problema: `vitest` con `environment: jsdom` no expone `node:fs`, y hace falta
`// @vitest-environment node`. **Eso ya estaba en el taste del proyecto**, y la compuerta no lo
sabe porque nadie se lo dijo.

### Causa raíz

`constitucion.Constitucion` tiene el `test_cmd` y nada sobre **qué hace falta para que ese comando
pueda correr**. El `test_cmd` se trata como un string ejecutable, y a veces es la punta de un
servicio.

Con H5 arreglado, la mitad de esto se evapora: los tests dejan de apuntar al `.sql` y pasan a ser
TypeScript que lee el `.sql`, que es lo que el implementador terminó escribiendo. Pero la otra
mitad queda, y es real: **hay lotes que necesitan un servicio, y el que planifica tiene que poder
saberlo antes de escribir el plan.**

### Arreglo

Un campo en el frontmatter de la constitución, opcional y en la misma clase que
`dependencias_aprobadas`:

```yaml
test_cmd: npm test
test_requiere: []          # o: [postgres] · [docker] · [redis]
```

`sf` **no lo resuelve** —no levanta contenedores, y no debería: eso es infraestructura del proyecto
y no de la máquina de estados— pero lo **muestra**: entra en el sobre del ⑫ (así el que planifica lo
lee antes de cortar los lotes) y `sf doctor` avisa si está declarado y no está disponible.

Y el segundo caso —el pragma de `vitest`— **no es de `sf`**: es una convención del proyecto y su
lugar es el cuerpo de la constitución, que el implementador lee entero. Lo que falta ahí es que
`sfp-constitucion` pregunte por ello. Una línea en el skill: *"si el runner necesita un pragma, un
`environment`, o una variable para correr, va acá — el implementador no lo va a adivinar."*

> **Este hallazgo es el que más fácil se sobre-diseña.** La tentación es que `sf` levante el
> Postgres. No: la vara es que el que planifica **sepa**, no que la herramienta **provea**.

### Lo que cambió cuando H3 pasó a `dont-ask`

Este hallazgo nació 🟡 y subió a 🟠 **por una consecuencia que no era suya**.

H3 decía que el `test_cmd` no entraba en la allowlist —que `sf` no autoriza un comando que no
leyó— y que `sf doctor` avisaría. Con `auto-accept` eso era defendible: en el peor caso el harness
**preguntaba**, y Javier apretaba enter.

Con **`dont-ask` no hay enter**: lo que no está en la lista se **deniega**. O sea que un `test_cmd`
fuera de la allowlist ya no es una molestia — **hace fallar el ⑲ de todos los lotes**, siempre, en
Command Code.

Así que el arreglo de H3 deja de ser cosa de `sf install` solo:

```
sf install   escribe la allowlist con lo que sabe: `sf *`, `git *`
sf init      detecta el stack y el test_cmd  →  AGREGA el test_cmd a la allowlist
sf doctor    comprueba que sigan estando, y lo dice si no
```

`sf init` es el momento correcto y no una excepción: es el único que ya mira el proyecto para
deducir el `test_cmd`, así que escribirlo donde hace falta es la misma pasada. Y sigue sin
autorizar nada que Javier no haya escrito — el `test_cmd` lo puso él en la constitución.

> **La lección, que vale más que el arreglo:** elegir `dont-ask` para que el bucle no se cuelgue
> movió el costo a otro lado. No lo eliminó. Un modo que nunca pregunta exige que la lista esté
> completa, y completarla es trabajo de `sf`, no del que la sufre.

---

## H7 — El ⑯ elige el modelo a ciegas: el sobre no lleva el menú

### Síntoma

El ⑯ escribe `"modelo": "opus"` en features que no lo necesitan (H1). El ejemplo del skill explica
*por qué escribe siempre lo mismo*; esto explica **por qué no podría escribir otra cosa aunque
quisiera**: el planificador no sabe qué modelos hay en esta máquina. No es que elija mal — no
tiene de dónde elegir.

### Causa raíz

`sobre.go:278`, el sobre del ⑫:

```go
case estado.Planificacion:
    s.Partes = []Parte{
        {Titulo: "El manual del proyecto", Rutas: []string{docs.Constitucion}},
        {Titulo: "Qué hay que resolver",   Rutas: rutasDeHistorias(fr.Historias)},
        aprendizajes(raiz),
    }
```

Tres partes y **las tres son rutas de archivos del repo**. El menú de modelos no es un archivo del
repo: vive en `~/.specforge/modelos.yaml`, es de la máquina y no del proyecto — que es la decisión
correcta y está argumentada en el encabezado de `global.go`. Pero la consecuencia nadie la sacó:
**ninguna ruta puede apuntar ahí, así que el menú no entra al sobre por el camino normal.**

Y sin embargo el ⑯ existe y tiene que decidir. Un paso que decide sobre un catálogo que no ve es
un paso que va a inventar, y lo que inventa es lo único que leyó: el ejemplo del skill.

### Arreglo

El mecanismo ya está construido y no hay que inventarlo. `Parte` tiene dos formas de traer material
y la segunda es exactamente ésta (`sobre.go:82`):

> `Contenido` — material **derivado** que no es un archivo: el lote de `tareas.json` filtrado, el
> diff de git, la corrida de mutantes.

El menú es material derivado del mismo tipo, así que viaja embebido:

```
── Con qué se puede implementar esta feature ──

Estos son los modelos declarados en esta máquina para el harness `opencode`.
Elegí por LOTE: no todos los lotes de una feature necesitan lo mismo.

  perfil `construir`  — el default de implementar
    sonnet   capacidad 3   el caballo de batalla; si dudás, éste
    laguna   capacidad 1   rápido y barato; bien en CRUD, plomería y refactors mecánicos

  perfil `razonar`    — pedilo cuando el lote decide algo, no cuando lo escribe
    opus     capacidad 3   diseño, concurrencia, hallar lo que falta
    r1       capacidad 2   razona bien y sale barato; la prosa le sale peor

Escribí el ALIAS (`laguna`), nunca el id. Si no decís nada, va el primero del perfil.
```

**Lo que el sobre NO lleva es el `id`.** El alias es vocabulario de Javier y sobrevive al cambio de
harness; el id es del harness y no tiene por qué entrar a un archivo versionado. Que el ⑯ no vea
los ids es la garantía barata de que no los pueda escribir.

Y el `para:` viaja **textual**, tal como Javier lo escribió. `sf` no lo resume, no lo reordena y no
lo interpreta: es la opinión de Javier llegando entera a quien la necesita, que es lo único que R3
le permite hacer con ella.

### Y por eso el modelo pasa a ser por lote

`tareas.go:57` argumenta que el modelo es de la feature y no del lote, y el argumento es bueno:

> *"Y por lote tampoco, porque el que decide el lote es el mismo que decidiría el modelo y lo haría
> en la misma pasada: sería un campo repetido con el mismo valor."*

**Es correcto, y la premisa que lo sostiene desaparece con el menú.** "El mismo valor" sólo es
cierto cuando hay un modelo por perfil: entonces sí, repetirlo por lote es ruido. Con varios, los
valores son **distintos**, y esa diferencia es el feature entero — el lote de plomería va con
`laguna` y el de concurrencia con `opus`, decididos en la misma pasada por el mismo que cortó los
lotes, que es justamente quien mejor sabe cuál es cuál.

El campo de la feature **no se va**: sigue siendo la forma de decir "toda esta feature es más
difícil de lo normal" sin repetirlo tres veces. Queda como el nivel 3 de la cadena, debajo del
lote y arriba del default.

> **El comentario de `tareas.go:57` hay que reescribirlo, no borrarlo.** Razonaba bien desde lo que
> sabía. Dejar el argumento viejo al lado del campo nuevo es cómo se fabrica el próximo malentendido.

### Los tests que tienen que existir

- El sobre de `planificacion` trae una parte con el menú, y su `Contenido` **no está vacío**.
- El menú lista los alias y las capacidades del **bloque del harness activo**, y de ningún otro.
- El menú **no contiene ningún `id`** — ni `anthropic/`, ni `laguna/s2.1`.
- Sin `~/.specforge/` (nadie corrió `sf install`), la parte aparece con `Falta` explicando por qué,
  y el sobre **no** falla. Es la regla de no mentir por omisión que `Parte.Falta` ya tiene.
- El sobre de `implementar` **no** trae el menú: en el ⑱ ya está decidido, y ofrecerle el catálogo
  al que implementa es invitarlo a cambiar una decisión que no es suya.

---

## H8 — El `modelo:` de `sf next` es inaccionable fuera de Claude Code

> **Este hallazgo no salió de leer: salió de la prueba de escritorio del 29/08, y después se
> comprobó ejecutando.** Es el único de los ocho con evidencia de las dos clases.

### Síntoma

`sf next` calcula bien qué modelo hace falta y lo imprime bien. Y el orquestador **no tiene con qué
obedecerlo**: en opencode y en Command Code el modelo de un subagente sale del `model:` de su
archivo de agente, decidido de antes. No hay parámetro de modelo en la invocación.

En Claude Code sí lo hay —la herramienta Task acepta `model`— y por eso nunca se vio.

**Sin esto, H2 y H7 son decoración.** `sf` puede tener los perfiles más limpios del mundo y el menú
más informativo: si el número que sale por stdout no puede llegar al subagente, el routing por fase
—que es el diferenciador— no existe fuera de Claude Code.

### La evidencia

Dos agentes de sonda en un directorio vacío, uno de ellos con un id de modelo **que no existe**:

```markdown
---
description: SpecForge sonda
mode: subagent
model: sonda/no-existe-xyz      ← a propósito
---
Contestá únicamente: SONDA-ROTA-RESPONDIO
```

La señal es binaria y no se puede falsear: si el harness ignorara el `model:` del archivo, el
subagente heredaría el modelo de la sesión y **contestaría normalmente**.

| Harness | Resultado |
|---|---|
| **opencode** | ✅ error de modelo desconocido |
| **Command Code** | ✅ error de modelo desconocido |

O sea que las dos mitades están: **el orquestador puede invocar un subagente por nombre, y el
`model:` del archivo se aplica.**

Y la segunda mitad del experimento —agregar un agente con la sesión ya abierta— dio que **no**: los
dos lo tomaron como un archivo cualquiera, no como un agente. **Las definiciones se leen al
arrancar.**

### Arreglo: el portamodelo

Un archivo por **alias declarado**, generado por `sf install` y regenerado por `sf model`. No
contiene una sola instrucción de SpecForge:

```markdown
---
description: SpecForge — construir · laguna
mode: subagent
model: laguna/s2.1
---
Seguí las instrucciones que te dé quien te invocó.
```

Y `sf next` gana un cuarto campo, **`agente:`** — a quién invocar para conseguir ese modelo:

```
estado:  implementar · lote 1 de 3      skill:  sf-build
perfil:  construir                      modelo: laguna/s2.1
agente:  sf-laguna                      via:    subagente
```

En Claude Code `agente:` viene **vacío** y el orquestador usa `modelo:` directo. Ésa es la mitad de
H1b que sólo `sf` puede contestar: *"¿cómo lanzo este modelo acá?"* tiene una respuesta distinta
por harness, y ningún skill ni el orquestador la pueden saber.

### Por qué un archivo por alias y no un agente de verdad por skill

Un agente por skill (`sf-build`, `sf-check`, …) tiene una ventaja real y hay que decirla: el
frontmatter admite `permission:`, o sea que un `sf-check` podría tener **`edit: deny`** y el revisor
no podría escribir código. Estructural le gana a instructivo, y hoy eso se previene sólo con prosa.

Se descarta igual, por tres razones y la primera es la que decide:

1. **El modelo y el skill son ortogonales, y pinearlos juntos los multiplica.** Si `sf-build` tiene
   el modelo fijo, el menú por lote muere: el lote 1 en `laguna` y el 3 en `opus` exigirían
   `sf-build-laguna` y `sf-build-opus`. Y como el `sf model` del ⑳ puede apuntar a cualquier alias
   en cualquier estado (A8 de `arreglos.md`), son **9 skills × N alias**. Con el portamodelo son
   **N**, porque el skill llega por el prompt. La factorización correcta es **el modelo en el
   archivo, el skill en el prompt**.
2. **Un permiso en el frontmatter es una compuerta viviendo en el harness**, y este proyecto ya
   decidió dónde viven las compuertas. No corre en un harness que no tenga la feature, no se puede
   testear desde Go, y no aparece en `sf done`. La misma garantía se consigue donde están todas las
   demás: **`sf done` del ㉑ mira si el árbol de trabajo cambió** — un `git status` sucio en
   `revision` es un hecho, es verificable, y vale en los tres. *(Eso es un hallazgo aparte y vale
   por sí solo, aunque todo esto se descarte.)*
3. **Meter permisos en el archivo lo volvería mutable durante el bucle.** Hoy el portamodelo es
   función del catálogo y de nada más; con permisos pasaría a depender de la fase, y habría que
   reescribirlo en cada paso — que es justo lo que el punto siguiente prohíbe.

### La consecuencia de que no haya recarga en caliente

**Los portamodelo se generan TODOS en `sf install`, no bajo demanda.** Es la diferencia entre que el
⑳ funcione y que no:

- `sf model <alias>` en medio del bucle apunta a un alias que **ya está en el catálogo** —si no
  estuviera sería la 🛑 de `sinDeclarar`—, así que su archivo ya existe y se cargó al arrancar.
  **El caso que importa está cubierto sin reiniciar nada.**
- El único que pide reinicio es **declarar un modelo nuevo** con la sesión abierta. `sf model`
  tiene que decirlo explícito, y no como una nota al pie:

```
✓ `laguna` declarado en `opencode` · perfil `construir`
⚠ reiniciá tu harness: los agentes se leen al arrancar y esta sesión no lo va a ver.
```

- Y el arranque de cero paga ese reinicio una vez: catálogo vacío → 🛑 → declarás dos → reiniciás.
  **La 🛑 tiene que avisarlo ahí**, no dejar que lo descubra fallando.

### Los tests que tienen que existir

- `sf install --harness=opencode` con dos alias en el catálogo escribe **dos** portamodelo en
  `.opencode/agents/`, y su `model:` es el `id` del alias.
- Con `--harness=claude-code` **no escribe ninguno**, y `sf next` devuelve `agente:` vacío.
- `sf model` con un alias nuevo escribe el portamodelo **y** devuelve el aviso de reinicio.
- `sf doctor` falla si un alias del catálogo no tiene su portamodelo.
- El portamodelo generado **no menciona ningún skill de SpecForge**: es un chequeo de una línea y es
  lo que garantiza que no se vuelva una cuarta copia.

---

## §7 — La matriz, verificada

Todo lo que sigue sale de la documentación de cada harness, leída el 2026-08-29. Es lo que
`sf install` tiene que saber, y es la única tabla de este documento que va a envejecer.

| | **Claude Code** | **opencode** | **Command Code** |
|---|---|---|---|
| orquestador | `CLAUDE.md` | `AGENTS.md` **o** `CLAUDE.md` — gana `AGENTS.md` | `AGENTS.md`, **no** `CLAUDE.md` |
| id de modelo | `opus` · `sonnet` · `haiku` | `provider/model` | id de su catálogo |
| permisos | `.claude/settings.json` | `opencode.json` → `permission` · o `--auto` | `.commandcode/settings.json` → `permissions` |
| modo desatendido | hereda del padre | `--auto` | `defaultMode: "auto-accept"` |
| subagentes | Task + `.claude/agents/` | Task + `.opencode/agents/*.md` | `.commandcode/agents/*.md` |
| modelo por subagente | **parámetro en la llamada** | frontmatter `model` — **fijo, verificado** | frontmatter `model` — **fijo, verificado** |
| ¿recarga agentes en caliente? | — | **no, verificado** | **no, verificado** |
| skills | `~/.claude/skills/` + plugins | `~/.claude/skills/` **y** `.agents/skills/` — **medido** | `.agents/skills/` sí, **`~/.claude/skills/` NO** — **medido** |
| permisos por default | hereda del padre | **permisivo** — corre sin preguntar (medido) | **`default`: pregunta por todo lo que cambia** (medido) |
| ¿lee los 18 de hoy? | sí | **sí**, sin tocar nada | **no** — se arregla con `settings.skills` (**verificado**) |

**Lo que esta tabla dice, y es la buena noticia del documento:** `sf install` ya escribe los dos
nombres del orquestador y eso **funciona en los tres**. La decisión de que `AGENTS.md` es el mismo
texto y no una traducción (`superficie-sf.md` §6) resultó ser exactamente correcta. Lo que falta no
es el orquestador: son los permisos y los ids.

---

## Lo que este documento NO propone, y por qué

- **Generar `.opencode/agents/*.md` o `.commandcode/agents/*.md` desde los 18 skills.** Sería una
  cuarta copia de cada skill, versionada por harness, que se desincroniza en la primera edición. Los
  tres leen `SKILL.md`; el problema es sólo **dónde buscan**, y eso se arregla con una ruta en un
  array (H4), no con archivos generados.
- **Un adapter, un plugin o una extensión por harness.** Es lo que propone
  `planning/SF-HARNESS-ARCH.es.md` (local, no versionado), y es otro producto. `sf`
  hoy es una tool que contesta preguntas por stdout, y **esa es la razón por la que corre en tres
  harness que no se conocen entre sí.** Lo que falta acá cabe entero en las respuestas que ya da.
- **Que `sf` traiga el catálogo de modelos de cada proveedor.** R3. Dos preguntas contestadas una
  vez por harness son más baratas que una tabla que hay que mantener para siempre.
- **Que `sf` elija de la lista.** La lista, el orden, el `capacidad:` y el `para:` son opinión de
  Javier. `sf` los transporta al sobre y el ⑯ decide. En el momento en que `sf` compare dos
  capacidades para elegir un modelo, volvió a opinar — con más pasos y peor informado que el ⑯,
  que al menos leyó la feature.
- **El fallback automático cuando un modelo no está disponible.** Exigiría que `sf` lance modelos,
  y no los lanza. El argumento entero está en H2.

---

## Cómo se sabrá que quedó bien

La misma corrida, en el mismo proyecto, en los tres harness:

```bash
sf install --harness=opencode    # AGENTS.md · opencode.json · el bloque `opencode:` vacío
sf init
sf doctor                        # harness: opencode · 9/9 · permisos ✓ · exit 0
sf next                          # 🛑 el perfil `razonar` no está declarado para `opencode`
sf model razonar   --alias opus   --id … --via subagente
sf model construir --alias sonnet --id … --via subagente
sf next                          # perfil: razonar · modelo: … · via: subagente
```

Y la prueba de H7, que es la que no se puede ver leyendo el estado: **`sf context` en el ⑫ tiene que
traer el menú, y el `tareas.json` que sale tiene que nombrar alias distintos en lotes distintos.**
Si los cinco lotes salen con el mismo alias, el menú llegó pero no se usó — y ahí la pregunta es la
de siempre: ¿el skill pidió mal, o la feature de verdad era pareja?

Y de ahí en adelante, el bucle: **ningún prompt de permisos, y `perfil:` distinto en ⑫ que en ⑱.**
Si esas dos cosas se ven, los seis hallazgos están cerrados.

**La prueba de que los perfiles quedaron bien es la cuarta corrida, no la primera.** Repetir lo de
arriba con `--harness=claude-code` tiene que pedir los dos ids **otra vez** —son otros, es otro
catálogo— y al volver a `sf install --harness=opencode`, **no tiene que preguntar nada**: el bloque
sigue ahí. Si vuelve a preguntar, el archivo no está indexado y estamos donde empezamos.
