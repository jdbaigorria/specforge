# Instalar SpecForge

Son **tres cosas** y sólo la primera se hace una vez en la vida:

```
① el binario `sf`          una vez por máquina
② los skills               una vez por máquina (o por harness)
③ sf install + sf init     una vez por proyecto
```

---

## ① El binario

Una línea:

```bash
curl -fsSL https://raw.githubusercontent.com/jdbaigorria/specforge/main/install.sh | sh
```

Baja el binario de tu plataforma, lo verifica contra el checksum del release y lo deja en
`~/.local/bin`. Si no hay release para tu máquina pero tenés Go, compila. Dos variables si querés
mandar vos:

```bash
SF_VERSION=v2.0.0 SF_BIN_DIR=/usr/local/bin sh install.sh
```

Y avisa de las dos cosas que se rompen en silencio: que el directorio no esté en tu `PATH`, y que
haya **otro `sf` adelante** que le gane al recién instalado.

> **Por qué esto es un shell script y no un comando de `sf`.** Porque `sf` no puede instalar a
> `sf`: lo que ponga el binario tiene que correr *antes* de que el binario exista.

### A mano, desde el código

```bash
git clone https://github.com/jdbaigorria/specforge
cd specforge/sf
go build -o ~/.local/bin/sf ./cmd/sf
```

> **Una sola dependencia.** El binario usa `gopkg.in/yaml.v3` y nada más. Todo lo demás es
> librería estándar de Go, y el shell-out a `git` y al runner de tests de tu proyecto.

### Compilarlo para otra máquina

Go cruza-compila sin toolchain extra:

```bash
GOOS=darwin GOARCH=arm64 go build -o sf-mac ./cmd/sf
GOOS=linux  GOARCH=amd64 go build -o sf-linux ./cmd/sf
```

---

## ② Los skills

Los skills son `.md` — **cualquier harness que sepa leer un archivo los puede usar.** Se instalan
copiándolos a donde tu harness los busque.

### Claude Code

```
/plugin marketplace add jdbaigorria/specforge
/plugin install specforge
```

El plugin trae los 18 y sabe actualizarlos. Es el camino recomendado.

Si preferís no pasar por el marketplace —porque estás editándolos, por ejemplo— un symlink por
skill al repo:

```bash
for d in "$PWD"/skills/*/; do ln -s "${d%/}" ~/.claude/skills/"$(basename "$d")"; done
```

> **Uno por skill, no uno a la carpeta.** Un symlink único a `skills/` deja cada `SKILL.md` un
> nivel más abajo de donde el harness los busca, y no aparece ninguno.

### Otro harness

Copiá `skills/` a donde tu harness lea sus skills. **No hay nada específico de Claude Code
adentro**: son 18 archivos Markdown con frontmatter.

### Qué se instala

```
skills/
  sfp-scout · sfp-po · sfp-constitucion · sfp-backlog · sfp-roadmap    los 5 de PRODUCTO
  sf-plan · sf-build · sf-check · sf-cierre                            los 4 de FEATURE
  sfx-think · sfx-grill-me · sfx-tdd · sfx-github · sfx-documenter
  sfx-journal · sfx-triage · sfx-explain · sfx-audit                   los 9 UTILITARIOS
```

**El prefijo dice algo:** `sfp-` corre una vez por producto, `sf-` una vez por feature, y `sfx-`
es un utilitario que **no sabe que la máquina existe** — sirve solo, en cualquier proyecto, con o
sin SpecForge.

---

## ③ Por proyecto

Dos comandos, en este orden:

```bash
cd <tu proyecto>
sf install
sf init
```

### `sf install` — el orquestador y tus modelos

```
+ CLAUDE.md
+ AGENTS.md
+ ~/.specforge/modelos.yaml

harness: claude-code · 3 modelos declarados
```

Hace dos cosas:

- **Pone el orquestador** en el proyecto, con los dos nombres. `CLAUDE.md` y `AGENTS.md` son **el
  mismo texto**, no una traducción: lo único que cambia entre harness es cómo se lanza un
  subagente, y eso el harness ya lo sabe hacer.
- **Arma `~/.specforge/`**, que guarda qué modelos tenés y cómo se invoca cada uno. Vive en tu
  home y no en el proyecto porque **tus modelos son los mismos en todos tus repos**.

#### Si ya tenías un `CLAUDE.md`

**No lo pisa.** Te avisa y sigue:

```
· CLAUDE.md (ya existe — `--forzar` lo pisa)
+ AGENTS.md
```

Tenés dos caminos: pegarle a mano las cuatro líneas del bucle (están en
[`docs/comandos.md`](docs/comandos.md)), o `sf install --forzar` si el que tenías no te importa.

#### Si no reconoce el harness

```
⚠ No reconocí el harness. Corregilo con `sf install --harness=<nombre>`.
```

La detección es una **pista** —una variable de entorno que puede estar o no—, así que se escribe
una vez y se corrige a mano cuando hace falta. **Importa** porque es la mitad de la decisión de
*cómo* se lanza cada modelo: en Claude Code no podés usar un modelo que no sea de Anthropic como
subagente, y en otro harness quizá sí.

### `sf init` — el andamio del proyecto

```
+ .docs/
+ .docs/backlog/
+ .docs/constitucion.md
+ .docs/estado.json

go (go.mod) · test_cmd: go test ./...
```

Hace tres cosas, **y las tres son mecánicas**:

- **Dos directorios, no siete.** `.docs/features/` y `.docs/archivado/` aparecen cuando hacen
  falta. Un directorio vacío desde el día uno es una promesa que el proyecto todavía no puede
  cumplir.
- **Detecta el stack** y llena `lenguaje`, `manifiesto` y `test_cmd` en la constitución. Reconoce
  Go, Rust, Python, Node, Ruby y Java mirando el manifiesto.
- **Crea el `estado.json` vacío**, que es lo que hace que `sf next` deje de decir *"corré
  `sf init`"*.

#### Si no reconoce tu stack

```
⚠ No reconocí el stack: completá `test_cmd:` en la constitución.
  Sin eso el ⑧ no sella y no se puede correr ningún test.
```

Abrí `.docs/constitucion.md` y completá la línea, o dejásela al ⑧ — llega igual. **No es
opcional:** sin `test_cmd` caen tres compuertas —el rojo, el verde y el conteo de tests— y
`sf approve` **no sella**:

```
✗ No pude.
  · la constitución no tiene `test_cmd:` — sin eso sf no puede correr los tests
```

#### Correrlo dos veces no rompe nada

```
Este proyecto ya está iniciado.

  sf next
```

`sf init` **nunca pisa un `estado.json` que ya existe** — ahí viven los sellos que aprobaste, y
ésos no se pueden reconstruir mirando los archivos. Tampoco pisa una constitución escrita.

---

## Verificar que quedó bien

```bash
sf doctor
```

Contesta por las tres cosas que se rompen en silencio, y por eso existe:

```
binario   v2.0.0
          ~/.local/bin/sf

harness   claude-code

skills    9/9 de la máquina
       ✓  sfp-scout          ~/.claude/skills/sfp-scout
       ✓  sfp-po             ~/.claude/skills/sfp-po
       …

proyecto  ~/proyectos/lo-que-sea
          .docs/ está — `sf status` dice dónde va
```

Y cuando algo falta, dice **cómo se arregla** en vez de mandarte al README:

```
✗ el `sf` del PATH no es éste: el agente va a correr /usr/local/bin/sf
✗ faltan instalar 9 skills: sfp-scout · sfp-po · …

cómo se arregla:
  el binario   curl -fsSL https://raw.githubusercontent.com/…/install.sh | sh
  los skills   /plugin marketplace add jdbaigorria/specforge
               /plugin install specforge
```

> **Las tres que agarra son las tres que le pasaron a este proyecto.** La instalación se hace en
> dos mitades —el binario por un lado, los skills por otro— y las dos pueden quedar bien *cada
> una por su lado* mientras no se ven entre ellas. El doctor no compara versiones: pregunta si
> los comandos que los skills instalados nombran existen en **este** binario, que es un hecho, no
> un parecido.

El código de salida es el de siempre: `0` anda, `2` hay algo que arreglar.

Después, el bucle:

```bash
sf next
```

En un proyecto recién iniciado tiene que contestar:

```
estado:   brief
skill:    sfp-scout
modelo:   sonnet
via:      vos

el ⑥ lo sellás vos cuando esté
```

Si ves eso, **está listo**. Seguí por [`docs/primeros-pasos.md`](docs/primeros-pasos.md).

---

## Qué se escribió, y dónde

```
<tu proyecto>/
  CLAUDE.md · AGENTS.md      el orquestador — cuatro líneas
  .docs/
    constitucion.md          las reglas del proyecto (la cabecera ya llena)
    estado.json              dónde estás. LO ESCRIBE SÓLO sf
    backlog/                 los us-#.md

~/.specforge/
  modelos.yaml               qué modelos tenés y cómo se invoca cada uno
```

**`.docs/` se versiona con el proyecto.** No lo pongas en `.gitignore`: el estado tiene que
viajar con el repo, porque cuando pasás a otro modelo el que implementa no estuvo en la
conversación.

---

## Actualizar

El binario:

```bash
curl -fsSL https://raw.githubusercontent.com/jdbaigorria/specforge/main/install.sh | sh
```

Los skills, si los instalaste por el plugin:

```
/plugin update specforge
```

**Actualizá los dos juntos.** Es el escenario para el que existe `sf doctor`: un skill nuevo que
le pide al binario viejo algo que todavía no sabe hacer traba el bucle, y el que se lo encuentra
es un agente a mitad de camino.

Y en cada proyecto, si cambió el orquestador:

```bash
sf install --forzar
```

## Desinstalar

```bash
sf uninstall
```

Saca `CLAUDE.md` y `AGENTS.md` **sólo si no los editaste** — si les agregaste algo, no los toca y
te lo dice.

**Nunca toca `~/.specforge/`**: tus modelos son de la máquina, no del proyecto. Sacar SpecForge
de un repo no puede borrarte la lista que construiste en los otros quince.

Y **no toca `.docs/`**: ahí está tu trabajo. Si querés que se vaya, borralo vos.
