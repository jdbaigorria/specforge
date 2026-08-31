# `sf install` interactivo — los arneses y los modelos, de una sola vez

**Fecha:** 2026-08-31 · **Branch:** `refundation` · **Commit:** `b93fdc3`

> **Estado: SIN IMPLEMENTAR.** Spec de diseño. Lo que sí está hecho es la **medición**: el TTY, los
> tres listados de modelos y las dependencias de `sf` se comprobaron ejecutando el 2026-08-31, y
> están abajo con su salida. Nada acá es supuesto.
>
> **No depende de [`headless.md`](headless.md).** Sirve hoy, con el modo orquestado que ya existe,
> y sigue sirviendo después. Se puede hacer antes, después o en paralelo.

---

## 1. El dolor

Javier, después de correr el bucle en los tres arneses el 2026-08-31:

> *"parece que sí anda cambiando de harness, lo cual es bueno pero es medio enquilombado […] y el
> tema de los modelos de cada uno, debería ser más sencillo."*

Y lo que hay que hacer hoy para dejar dos arneses andando, en orden:

```bash
# en Claude Code
sf install
sf model razonar   --alias grande --id opus   --via subagente
sf model construir --alias medio  --id sonnet --via subagente
# cerrás, abrís opencode
sf install
sf model razonar   --alias ultra  --id opencode/nemotron-3-ultra-free --via subagente
sf model construir --alias rapido --id opencode/…                     --via subagente
# ⚠ reiniciá tu harness: los agentes se leen al arrancar
# cerrás opencode, lo volvés a abrir
```

**Ocho comandos, dos reinicios y un ida y vuelta entre arneses** para decir dos cosas: cuáles usás
y con qué modelo. Y los ids hay que sacarlos de otro lado y tipearlos.

Lo que propone esta spec:

```bash
sf install
```

---

## 2. La línea: quién conversa y quién no

`sf` **nunca** conversa. Es la regla que lo hace usable por un agente: `sf next` es consulta pura,
`sf done` contesta y se va, ninguno espera a nadie.

`sf install` es la excepción, y la excepción tiene un fundamento que ya existía: **`sf install` no
es de la máquina.** No mira el estado ni lo mueve, no aparece en ningún trazado del bucle
(`andamio.go`, encabezado del paquete). Es lo que hace que SpecForge se pueda *usar*, y lo corre una
persona, una vez.

### Y no hay que elegir, porque se puede saber

**Medido el 2026-08-31**: la shell de un agente **no tiene TTY**; la de una persona sí.

```
adentro de la herramienta Bash de Claude Code   stdin: NO es TTY · stdout: NO es TTY
en una terminal de verdad                       stdin: SÍ es TTY
```

Así que la regla es una llamada, no una arquitectura:

```go
// hayPersona dice si del otro lado hay alguien a quien preguntarle.
//
// Sin esto, un `sf install` corrido por un agente —el orquestador lo nombra en
// "Empezar de cero", y `sf doctor` lo sugiere— colgaría esperando una respuesta
// que nadie va a escribir.
func hayPersona() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
```

**Con la stdlib sola, y eso importa:** `sf` tiene hoy **una** dependencia (`gopkg.in/yaml.v3`).
Meter `golang.org/x/term` para preguntar si hay una terminal la duplicaría. Probado en los dos
sentidos y anda.

| | quién lo corre | conversa |
|---|---|---|
| `sf install` | Javier, una vez | **sí, si hay TTY** |
| todo lo demás | el orquestador, siempre | no, nunca |

> **Los flags no son el salvavidas: son el mecanismo.** `sf install --harness=a,b,c` es cómo se
> scriptea y cómo debería sugerirlo `sf doctor` —que hoy dice `sf install` pelado
> (`doctor.go:209`)—. Si vienen flags, no se pregunta nada, haya TTY o no.

---

## 3. Lo que se midió — los tres listados

**Dos de tres pueden enumerar sus modelos**, y uno regala las descripciones.

| | comando | qué devuelve |
|---|---|---|
| **opencode** | `opencode models [provider]` | **397** ids, uno por línea. Acepta filtro por proveedor. Sin JSON. |
| **Command Code** | `cmd --list-models` | **62**, agrupados por proveedor, **con descripción**. Sin JSON (`--json` se ignora). |
| **Claude Code** | *no hay* | el `--help` nombra los alias; el error de un modelo inválido no lista. |

Y así se ve el de Command Code:

```
Open Source

deepseek/deepseek-v4-flash    fast hybrid-attention reasoning (default)
moonshotai/kimi-k3            long-horizon coding & knowledge work with 1M context
zai-org/glm-5.2               powerful coding with 1M context and long-horizon tasks
```

**Esa segunda columna es `capacidad` y `para`** — los dos campos que el catálogo ya tiene para que
el ⑯ elija, y que hoy Javier escribe a mano. El arnés los está regalando.

> **Y `sf` no los interpreta.** Los copia. Es R3 sin discusión: "long-horizon coding" es la opinión
> del proveedor sobre su modelo, y `sf` la transporta igual que transporta un id. El día que el ⑯
> elija con eso, está eligiendo con lo que dijo el proveedor y lo que corrigió Javier — nunca con
> algo que `sf` dedujo.

---

## 4. La forma

```
$ sf install

¿Qué arneses vas a usar acá?
  [x] claude-code     (detectado)
  [ ] opencode
  [ ] commandcode

── opencode ────────────────────────────────────
razonar   — el que compara y juzga (⑫ y ㉑)
  1  opencode/nemotron-3-ultra-free
  2  opencode/nemotron-3.5-lightning-free
  …  397 en total — escribí para filtrar, o pegá un id
> 1
  alias corto para éste: ultra
  esfuerzo (enter = el del modelo): high

construir — el que escribe el código (el resto)
> …

── commandcode ─────────────────────────────────
razonar
  1  minimaxai/minimax-m3     frontier coding, agents & native multimodality
  2  zai-org/glm-5.2          powerful coding with 1M context and long-horizon tasks
> …

✓ 3 arneses · 6 perfiles declarados
  + .claude/settings.json
  + .opencode/agents/sf-ultra.md · sf-rapido.md
  + .commandcode/settings.json · agents/sf-…
```

Tres cosas de esa pantalla que no son cosméticas:

1. **`mecanico` no se pregunta.** Ningún estado lo pide (`perfilPorEstado`); lo nombran los `sfx-*`,
   que están fuera de los nueve. Preguntarlo sería pedir una decisión que no frena nada.
2. **El alias lo pone Javier, no `sf`.** Es su vocabulario (H3). Derivarlo del id —`nemotron-3-ultra-free`
   → `ultra`— sería `sf` opinando sobre cómo se llaman las cosas de él.
3. **El esfuerzo se pregunta pero admite enter.** Vacío es "el que traiga el modelo", que **no** es
   lo mismo que "bajo", y esa distinción ya está en el código.

---

## 5. Los tres pedazos, y sirven sueltos

### ① `sf install --harness=a,b,c`

Hoy `Opciones.Harness` es un string y `Instalar` arma el andamio de uno. Pasa a ser una lista.

**No hay cambio de esquema en el catálogo**: `harnesses:` ya está indexado por arnés desde H2. Es lo
que menos toca de los tres.

### ② `sf models [--harness=X]`

Corre el listado del arnés y lo devuelve normalizado: `id`, y `descripción` si el arnés la da.

Sirve suelto: hoy, para declarar un modelo, hay que ir a buscar el id a otra ventana — `sf doctor`
literalmente te manda a `opencode models` o a `/model`.

### ③ La pregunta

Sólo tiene sentido con ① y ② hechos. Usa ② para el menú y ① para escribir.

---

## 6. El premio que no es obvio: se muere el reinicio

**Medido el 2026-08-29 (sonda 1):** las definiciones de agente se leen **al arrancar**. No hay
recarga en caliente. Por eso `sf model` avisa *"reiniciá tu harness"*, y por eso hoy el ciclo tiene
un cierre-y-abrí en el medio.

Ese aviso existe porque hoy declarás **estando adentro** del arnés que vas a usar.

Si declarás los tres de una, **los portamodelo de opencode y de Command Code quedan escritos antes
de que abras esos arneses**. Cuando los abrís, ya están cargados.

> No es que la molestia se reduzca: **el paso desaparece.** El aviso sigue existiendo para el
> `sf model` suelto —el ⑳, cuando el bucle patina y subís el modelo en caliente—, que es el único
> caso donde de verdad estás adentro.

---

## 7. Lo que NO cambia

- **`sf install` sigue sin ser de la máquina.** No mira el estado ni lo mueve.
- **Los flags mandan.** Con `--harness=`, no pregunta, haya TTY o no.
- **Sin TTY se comporta exactamente como hoy**: detecta e instala uno. Cero regresión para el
  orquestador y para `install.sh`.
- **`sf` no elige un modelo ni traduce un id de un proveedor a otro** (R3). Lista, y Javier elige.
- **No se pisa nada.** La regla de `CLAUDE.md`/`AGENTS.md` y de `~/.specforge/` sigue igual, con la
  marca de versión que se agregó en `be72b61`.

---

## 8. El riesgo, y cómo se acota

**Parsear una salida hecha para humanos es frágil.** `cmd --list-models` tiene encabezado,
secciones por proveedor, dos columnas alineadas con espacios y un pie con un link. Ninguno de los
dos ofrece JSON. Eso se rompe el día que cambien el formato, y se rompe en silencio.

La mitigación no es parsear mejor: es **que el listado sea una comodidad y nunca el mecanismo.**

```
el menú falla  →  "no pude listar los modelos de opencode. Pegá el id:"  →  seguís
```

Tipear el id tiene que funcionar siempre, y tiene que ser un camino de primera clase, no un
fallback vergonzante. Si `sf` queda inservible porque un arnés cambió una tabla, el diseño está mal.

> **Corolario para `sf doctor`:** que sepa decir "el listado de X no se pudo leer" como **aviso**,
> nunca como falla. No poder mostrar un menú no impide trabajar.

---

## 9. Cómo se sabrá que quedó bien

```bash
# la instalación entera, en un solo comando y sin cambiar de ventana
sf install --harness=claude-code,opencode,commandcode
sf models --harness=opencode | head          # ids, y descripción si la hay
```

Y las cuatro pruebas que cierran el diseño:

1. **Sin TTY se porta como hoy.** `sf install < /dev/null` en un pipe: detecta uno, no pregunta
   nada, no cuelga. Es el caso del orquestador y el que no puede fallar nunca.
2. **Con flags no pregunta**, aunque haya una persona mirando.
3. **Sin reinicio.** Declarar los tres arneses desde Claude Code, abrir opencode **por primera vez**
   y que `sf next` resuelva su modelo y su `agente:` **sin haber reiniciado nada**. Es la prueba de
   §6, y es la que se ve.
4. **Con el listado roto igual se puede.** Simular que `opencode models` falla y comprobar que la
   instalación sigue pidiendo el id a mano y termina bien.

> **La prueba de que quedó bien no es que la pantalla sea linda: es el punto 3.** Ocho comandos, dos
> reinicios y un ida y vuelta entre ventanas tienen que quedar en uno.
