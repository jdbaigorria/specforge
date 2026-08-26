# Salir a la cancha — lo que falta para una corrida de verdad

**Fecha:** 2026-08-25 · **Branch:** `refundation` · **Commit:** `b850cce`

> **Estado: PENDIENTE.** La máquina está entera y auditada, el instalador está construido, y en
> esta máquina todo funciona. Lo que falta no es código: es **publicar** y **correr el bucle con
> agentes de verdad**, que es lo único que ningún test de este repo puede hacer por sí solo.

Este documento continúa a [`arreglos.md`](arreglos.md). Aquel era el parte de daños del binario;
éste es el de **la costura entre el repo y el mundo**: instalar, publicar, y la primera corrida
real.

---

## La diferencia que ordena todo lo que sigue

> **Lo auditado fue el binario. Lo que nunca se probó es que un agente escriba lo que la compuerta
> acepta.**

Los tres tests punta a punta (`sf/cmd/sf/e2e_test.go`) escriben cada artefacto **a mano**. Prueban
que la máquina no se puede evadir; no prueban que los skills produzcan lo que la máquina espera.
Esa costura tiene un solo defecto conocido y ya arreglado —`sfp-scout` trababa el ⑥— y se encontró
**leyendo**, no corriendo.

Todo lo que falta en este documento sale de ahí.

---

## Lo que YA está verificado

No es una lista de intenciones: cada línea se comprobó corriendo algo.

| Qué | Cómo se comprobó |
|---|---|
| la máquina entera | 288 tests · 5 e2e · `vet` y `gofmt` limpios · 9 chequeos de CI |
| los 9 skills de estado ↔ las compuertas | nombres de archivo, ids `us-#/CA-#`, las 5 paradas de `approve`, los enums de `revision.json` |
| los 28 templates que los skills nombran | existen todos, y los 3 JSON calzan con los structs de Go |
| `install.sh` | corrida real: detectó la plataforma, no encontró release, clonó, compiló, instaló |
| `sf doctor` | detecta el binario sombra (`exit=2`) y el skill desalineado (`exit=2`), en vivo |
| instalación limpia en proyecto nuevo | `sf install` → `sf init` → `sf doctor` → `sf next`, los cuatro `exit=0` |

**Y lo que está instalado en la máquina de Javier hoy:** el binario en `~/.local/bin/sf`
(`v2.0.0-dev`), y los 18 skills como **symlinks individuales** al repo.

---

## ① Publicar — sin esto, nada de lo demás existe para nadie más

**El bloqueo más grande y el más barato.**

```
origin/main   ← el producto VIEJO: skills sin prefijo, sin sf/, sin .claude-plugin/
refundation   ← la refundación entera, NUNCA pusheada
```

`git rev-list --left-right --count origin/main...refundation` da **`0`** del lado de `main`: no
hay divergencia, es un fast-forward limpio. (El número del otro lado crece con cada commit; el que
importa es el cero.)

Mientras eso siga así:

- `curl … install.sh | sh` baja un script que **no existe en `main`**;
- `/plugin install specforge` trae los skills de **la versión apagada**;
- y `go install …@latest` resuelve al módulo viejo.

> **Por qué esto no lo decidí yo.** Pushear la refundación entera sobre `main` reemplaza el producto público
> por otro con los mismos nombres de repo y de comando. Es la clase de acción de una sola
> dirección, y es de Javier.

**Cómo se verifica que quedó bien:** desde una carpeta cualquiera, `curl … | sh` y después
`sf version`. Si dice la versión del release y no *"sin-versión"*, el camino entero anda.

---

## ② El plugin — la única pieza que no se puede verificar sin publicar

`.claude-plugin/plugin.json` declara ahora `"skills": "./skills/"`.

**La evidencia de que ése es el campo correcto es indirecta:** el único plugin instalado en esta
máquina que funciona —`ui-ux-pro-max`— declara `"skills": "./.claude/skills/"`. O sea que el campo
existe y se respeta. Lo que **no** sé es si `skills/` en la raíz se descubriría igual sin
declararlo. La línea explícita cuesta menos que averiguarlo publicando, y de paso deja afuera
`skills-community/`, que es lo correcto: no son del núcleo y no los valida el CI.

**Qué comprobar después del push, en una sesión limpia:**

```
/plugin marketplace add jdbaigorria/specforge
/plugin install specforge
```

y después:

| Pregunta | Cómo se contesta |
|---|---|
| ¿aparecen los 18? | `sf doctor` — tiene que decir `9/9 de la máquina` |
| ¿desde dónde? | el doctor imprime la ruta de cada uno |
| ¿convive con los symlinks? | si el doctor sigue mostrando `~/.claude/skills/…`, el symlink gana |
| ¿entraron los de `skills-community/`? | no deberían |

> **Ojo con el orden.** Si el plugin y los symlinks conviven, `sf doctor` reporta el **primero**
> que encuentra, y su lista de raíces pone `~/.claude/skills/` antes que las del plugin. O sea que
> mientras existan los symlinks, el plugin queda tapado y no se prueba nada.

---

## ③ La corrida real — lo que de verdad falta

**Nadie corrió el bucle con agentes.** Esto es lo único de esta lista que no es infraestructura.

### Dónde NO probarlo

**No en el repo de SpecForge.** Dos motivos, y el segundo es el que muerde:

- `.docs/` sería la documentación del propio SpecForge, mezclada con `specs/`;
- el `sf` bajo prueba estaría construyéndose a sí mismo, y a la tercera vuelta nadie sabe qué
  binario está mirando.

### Dónde sí

Un proyecto **chico, descartable, y con una idea real** — no un "hola mundo", porque el ⑥ necesita
algo sobre lo que se pueda investigar y opinar, y el ⑨ necesita historias que se puedan partir.

```bash
mkdir prueba && cd prueba
git init && <el manifiesto de tu stack>
sf install
sf init
sf next          # → estado: brief · skill: sfp-scout · via: vos
```

### Qué mirar mientras corre

Esta es la lista de sospechosos, y sale de dónde viven las costuras que nunca se ejercitaron:

| Estado | Lo que hay que ver |
|---|---|
| **⑥ brief** | que `sfp-scout` escriba el `veredicto` PROPUESTO, no vacío — es el bug que se arregló leyendo |
| **⑧ constitución** | que el `test_cmd` que detectó `sf init` sobreviva a que el skill reescriba el archivo |
| **⑨ backlog** | que los `CA-#` salgan **con texto**, y que las marcas del esqueleto no queden |
| **⑩ roadmap** | que ninguna historia quede fuera de una feature |
| **⑫–⑰ plan** | tres opciones exactas, y que `satisface` cubra **todos** los criterios en los dos sentidos |
| **⑱–⑳ build** | que escriba los tests **con los nombres que el plan dijo**, y vea el rojo antes |
| **㉑ revisión** | que opine sobre **todos** los criterios, y que un hallazgo abierto mande de vuelta |
| **㉓ cierre** | `doc.md` y `journal.md`, y que `sf approve` archive sin dejar nada a medias |

**Cada `✗` de una compuerta durante la corrida es un dato, no un fracaso.** La pregunta a
contestar en cada uno es la misma:

> ¿el skill pidió mal, o la compuerta exige de más?

Si es lo primero, se arregla el skill. Si es lo segundo, se arregla la compuerta. Lo que **no** se
hace es aflojar la compuerta para que pase — es exactamente lo que `sf-build` le prohíbe al
implementador.

### El residuo que ya se conoce

El ⑩ corriendo a mitad de ciclo: un `sf new` deja una historia huérfana, y el `sf done` del
subagente del roadmap cae en el camino de la feature y suma un `intentos_fallidos` de más. Ya está
documentado en `arreglos.md` con el motivo de por qué las dos salidas obvias son peores, y
`sfp-roadmap` ahora avisa. **Si aparece en la corrida, es esto y no algo nuevo.**

---

## ④ Lo que conviene NO hacer todavía

**No cambiar los symlinks por el plugin mientras estés arreglando cosas.**

Hoy los skills son symlinks al repo: editás `sfp-scout` y el cambio vale en la próxima invocación.
Una instalación por plugin es una copia congelada, y cada arreglo que salga de la corrida te
obligaría a reinstalar antes de seguir.

> El plugin se prueba **una vez**, para saber que funciona (②). Después se vuelve al symlink hasta
> que la etapa de arreglos termine.

---

## ⑤ Cuando todo esto pase — el primer release

El workflow está construido y espera un tag `v*`. Los tags que hay (`v0.6` … `v1.0`) son **del
producto viejo**: `v1.0` apunta a `e3f2a71`, *"sf CLI — deterministic helper over Markdown"*. La
refundación arranca en **`v2.0.0`**.

```bash
git tag v2.0.0 && git push origin v2.0.0
```

El workflow corre `vet`, los tests, el e2e, compila las cinco plataformas, comprueba que
`sf version` diga el tag, y publica con checksums.

**Después del primer release, verificar el camino que hoy no se puede:** en una máquina limpia (o
un contenedor), `curl … | sh` tiene que bajar el binario **del release** y no caer al plan B de
compilar. Si compila, el nombre del asset y el que `install.sh` busca no coinciden — hay un
chequeo de CI que los compara, pero sólo compara los **patrones**, no un release de verdad.

---

## Resumen, en orden

```
①  push de refundation a main          la única acción irreversible, y es de Javier
②  probar el plugin, una vez           lo único que no se puede verificar sin publicar
③  la corrida real, en un proyecto     lo único que no es infraestructura
     descartable                        ← acá está el valor
④  volver al symlink y arreglar        lo que salga de ③
⑤  tag v2.0.0                          cuando ③ y ④ hayan cerrado
```

**③ no depende de ① ni de ②.** Se puede hacer hoy, con lo que ya está instalado. Los otros cuatro
son para que lo pueda hacer alguien más.
