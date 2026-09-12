<p align="center">
  <img src="assets/specforge-logo.png" alt="SpecForge" width="380" />
</p>

# SpecForge

**SpecForge no te dice cómo trabajar. Ejecuta cómo trabajás vos.**

`sf` es una máquina de estados que sabe **dónde estás**, **qué sigue** y **si podés avanzar**. Tu
agente le pregunta, hace el trabajo, y vuelve a preguntar.

**[Instalar](INSTALL.md)** · **[Documentación](docs/)** ·
**[Una vuelta completa](docs/primeros-pasos.md)** · **[Comandos](docs/comandos.md)** ·
**Licencia:** [MIT](LICENSE)

---

## La idea, en una línea

> `sf` expone la máquina de estados que guía al harness — y es el árbitro que decide si se puede
> avanzar.

Son dos verbos, y el segundo es el que importa:

| | Qué hace | Sin esto sería… |
|---|---|---|
| **expone** | *"estás en el ⑬, ahora toca el ⑭"* | una lista de tareas |
| **comprueba** | *"no terminaste: falta la branch"* | una sugerencia que el agente puede ignorar |

`sf` **no maneja el auto** —no agarra el volante ni elige la ruta— pero **da verde o rojo, y en
rojo no se pasa**. Su poder es uno solo: es el único que puede mover el estado, y sólo lo mueve
cuando lo comprobó él mismo.

## El reparto

```
sf       →  DÓNDE estás · QUÉ sigue · ¿PODÉS avanzar?
skills   →  CÓMO se hace cada paso
harness  →  lo HACE
vos      →  DECIDÍS   (en tres puntos, y sólo tres)
```

**El orquestador es el agente, no el CLI.** `sf` arranca, contesta y se muere: dura
milisegundos. Un programa muerto no puede lanzar a nadie — no tiene manos. El único vivo durante
toda la sesión es el agente, así que el que lanza es él.

Por eso el `CLAUDE.md` son **cuatro líneas**:

```
1.  Corré `sf next`.
2.  Hacé lo que diga, con el skill y el modelo que diga:
      via: vos        → trabajás vos, de frente
      via: subagente  → lanzás un subagente fresco
      via: consola    → salís por CLI con el `comando:` que te dio
3.  Cuando vuelva el control, corré `sf next` otra vez.
    NO leas lo que devolvió el que trabajó — el estado es la verdad.
4.  Si `sf` dice 🛑 o ⏸, mostrale al usuario y esperá.
```

**`AGENTS.md` es el mismo archivo, no una traducción.** Lo único que cambia entre harness es cómo
se lanza un subagente, y eso el harness ya lo sabe hacer.

## Las tres reglas duras

1. **`sf` nunca lanza a nadie.** Si spawneara, manejaría contexto y tool-calling — y **sería un
   harness**, reinventando lo que Claude Code y Codex ya hacen bien.
2. **El estado avanza con hechos comprobados, nunca con la palabra del que trabajó.** El
   subagente dice *"terminé"*; `sf` **no le cree**: corre los tests él, mira si existe el archivo,
   mira si hay branch. Si falta algo, el estado **no se mueve**.
3. **El estado vive en el repo, no en el chat.** Un `estado.json` versionado — porque cuando
   pasás a otro modelo, el que implementa no estuvo en la conversación.

## Arrancar

```bash
cd sf && go build -o ~/go/bin/sf ./cmd/sf
cp -r skills/maquina/*/ skills/utiles/*/ ~/.claude/skills/

cd <tu proyecto>
sf install    # CLAUDE.md + AGENTS.md · ~/.specforge/ con tus modelos
sf init       # el andamio: 2 directorios, detecta el stack, el estado vacío
sf next       # y de acá en adelante, el bucle
```

## Los nueve estados

Cinco corren una vez por producto, cuatro una vez por feature:

```
PRODUCTO   brief → prd → constitución → backlog → roadmap
FEATURE    planificación → implementar → revisión → cierre
```

**Tres de ellos paran y te preguntan**, y sólo tres: el sello del brief, la constitución y la
revisión del plan. Todo lo demás avanza solo.

Cada estado tiene un skill que sabe hacerlo, y `sf next` te dice cuál — así la tabla
`estado → skill → modelo` vive en un solo lugar en vez de una copia por harness.

## Los comandos

```
sf init                       el andamio
sf install · uninstall        el orquestador + ~/.specforge/

sf next                       dónde estás · qué sigue · skill · modelo · via
sf context [--completo]       el sobre del estado actual. Sin argumentos
sf done [--msg "…"]           corre las compuertas y mueve — o dice qué falta
sf lote start                 crea la branch · exige el ROJO · guarda el hash
sf new "…"                    mete una feature o un bug al backlog

sf approve                    sella lo que estés mirando
sf reject "motivo"            no sella, y guarda el motivo para el que rehaga
sf take <f-#>                 saca la próxima del roadmap
sf model <nombre> [--via …]   sube el modelo — y lo declara si es nuevo
sf dismiss <h-#> "motivo"     descarta un hallazgo de la revisión

sf status                     dónde está todo — el único para humanos
sf audit [f-# …]              el punta a punta: varias features contra sus historias
```

## Qué impide de verdad

No por opinión, por aritmética:

- **El commit no se hace** → cerrar el lote **es** commitear. Sin `--msg` no hay `done`.
- **Commits mal agrupados** → el agrupamiento se decidió al planificar. Un lote, un commit.
- **No se crea la branch** → `sf lote start` la crea, y sin ella no sigue.
- **"Todo verde" sin tests** → `sf` tiene la lista exacta de tests que deben existir, los corre
  él, y **no te deja implementar hasta haberlos visto fallar**.
- **Un test aflojado para que pase** → el hash de los archivos de test se toma en el rojo y se
  compara en el verde.
- **"Terminado" con media historia** → cada criterio de aceptación tiene id. `sf` no juzga la
  revisión: comprueba **que el juicio haya ocurrido, sobre todos**.
- **Mocks con el contexto al 50%** → un subagente fresco por lote. Es el único de todos que se
  puede atacar *antes* del daño en vez de detectarlo después.

## Qué NO promete

- **Calidad de código.** `sf` garantiza estructura, secuencia y evidencia — no que el diseño sea
  bueno. Una compuerta frena sobre un **hecho**; un juez opina.
- **Que el modelo no alucine.** Produce artefactos que lo hacen alucinar *menos*, y comprueba lo
  comprobable.
- **Decidir por vos.** La IA propone, la última palabra es tuya. Siempre. Por eso una dependencia
  no aprobada **avisa** en vez de frenar: una herramienta que frena sola rompe esa regla.

## La estructura

```
sf/          el binario — 24 paquetes, 529 tests, una sola dependencia
skills/
  maquina/   los 9 de estado (sfp-* · sf-*) — sf doctor los EXIGE
  utiles/    los 16 utilitarios (sfx-*) — fuera de la máquina, standalone
  contrib/   skills de la comunidad — en el repo, NO los publica el plugin
specs/       el diseño, y por qué cada decisión es como es
```

**El prefijo dice algo:** `sfp-` corre una vez por producto, `sf-` una vez por feature, y `sfx-`
es un utilitario que vive **fuera** de la máquina — standalone, usable en cualquier proyecto.

**Y hay un guion de humo detrás de un build tag** —`go test -tags e2e ./cmd/sf/`— que compila el
binario y recorre las dos vueltas punta a punta contra un repo de verdad. Está aparte a propósito:
lo que más importa comprobar acá vive en la **costura entre comandos**, y ningún test de paquete
lo ve.

## La documentación

| | |
|---|---|
| [`INSTALL.md`](INSTALL.md) | instalarlo |
| [`docs/primeros-pasos.md`](docs/primeros-pasos.md) | una vuelta completa, de punta a punta |
| [`docs/comandos.md`](docs/comandos.md) | los 15 comandos |
| [`docs/estados.md`](docs/estados.md) | los 9 estados y qué exige cada compuerta |
| [`docs/artefactos.md`](docs/artefactos.md) | cada archivo y su forma |
| [`docs/skills.md`](docs/skills.md) | los 25 skills |
| [`docs/problemas.md`](docs/problemas.md) | qué hacer cuando `sf` te frena |

## El diseño

Todo en [`specs/`](specs/), y no es decoración: cada documento dice **por qué** y qué se
descartó. Empezá por [`session.md`](specs/session.md), que apunta al resto.
