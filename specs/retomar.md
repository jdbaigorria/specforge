# Retomar — por dónde seguir

**Fecha:** 2026-09-05, de noche · **Branch:** `refundation` · **Commit:** `0287486`

> **Qué es esto.** El traspaso de la sesión del 2026-09-05. Está escrito para arrancar en frío:
> si mañana lo lee Javier con el mate, o lo lee un agente sin contexto, los dos tienen que poder
> empezar sin releer nada más que esto y `el-arnes-ciego.md`.

---

## 1. Dónde quedamos, en cinco líneas

Se corrió **por primera vez** la máquina con un modelo de verdad escribiendo el artefacto. Dos
corridas, misma idea, dos modelos. **Salieron veredictos opuestos y la máquina aceptó los dos.**

La causa no era el modelo: **el arnés estaba ciego** — SpecForge nunca le dio herramientas de
búsqueda a ninguno de los dos. Se arregló la compuerta del brief, se escribió el parte entero, y
quedó una lista de seis cosas en orden.

**Todo está commiteado. El árbol está limpio.**

```
0287486  docs(corrida): las tres preguntas de la fase eran una sola — §6 y §7
3f8c14a  docs(corrida): el arnés estaba ciego — el parte de la primera corrida real
4d7cc28  fix(compuerta): un veredicto sin una sola fuente no sella — lo midió la corrida
e47f847  feat(registro): el tablero  ← acá estábamos anoche
```

---

## 2. Lo primero que hay que saber antes de tocar nada

**El binario de `sf` se reinstaló hoy.** El que estaba en `~/.local/bin/sf` era viejo: no tenía
`log`, ni `lanzar`, ni `models`. El viejo quedó en `~/.local/bin/sf.viejo-20260905`.

**Los tests son DOS suites, y una no corre sola:**

```bash
cd sf
go test ./...                    # 482
go test -tags e2e ./cmd/sf/      # 14   ← estos están detrás de //go:build e2e
```

Hoy **dos de esos 14 se rompieron** con el arreglo de la compuerta, y la suite normal no dijo nada.
**Correr las dos, siempre.**

**`sf install` te cambia el harness global.** Corrió una vez en el banco de pruebas y movió
`~/.specforge/modelos.yaml` de `harness: commandcode` a `harness: claude-code`. El resto del
catálogo quedó intacto. Si lo querés de vuelta en commandcode, es una línea.

---

## 3. El banco de pruebas — está armado y guardado

```
~/projects/workspace/personal/sf-banco/
  IDEA.md          la idea semilla. Es lo ÚNICO que se cambia si querés otra
  GUION.md         el paso a paso de T1 y T2
  comparar.sh      ./comparar.sh          ← ya tiene default A B
  banco/           la plantilla limpia. NO SE TOCA, se copia
  corridas/A/      Claude Code · selló no-lo-hagas · 13 links   ← el producto quedó cerrado
  corridas/B/      nemotron    · selló hacelo     ·  0 links   ← llegó hasta constitucion
```

**Las dos corridas se conservan a propósito**, con su `.docs/`, su `estado.json` y su
`registro.jsonl`. Son la evidencia. La corrida anterior de Javier llegó a 7 vueltas y no quedó nada;
ésta no se pierde.

Para volver a empezar una corrida de cero:

```bash
cd ~/projects/workspace/personal/sf-banco
rm -rf corridas/A && cp -r banco corridas/A && truncate -s 0 corridas/A/.specforge/registro.jsonl
```

---

## 4. El siguiente paso, concreto — **el nivel 0 en el skill**

> **Es lo primero de la lista, es de una tarde, no toca una línea de Go, y habría evitado el fallo
> de la corrida.** Está en [`el-arnes-ciego.md`](el-arnes-ciego.md) §6 y §9.

### El problema, en una línea

`skills/sfp-scout/references/method.md` manda investigar con **Tavily y el MCP de GitHub** — o sea
directo al nivel que necesita llave y configuración por arnés. **El nivel que funciona en cualquier
arnés no está nombrado en ninguna parte.**

### La medición que lo justifica

```bash
curl -s "https://registry.npmjs.org/-/v1/search?text=claude+code+usage+cost&size=3"
```

Devolvió **tres competidores reales** de la idea de prueba, al instante, sin llave. Si nemotron
hubiera hecho esa sola llamada, no habría inventado nada.

### Qué hay que escribir

**Archivos:** `skills/sfp-scout/references/method.md` y `references/tooling.md`.

1. **`method.md` §1 arranca por el nivel 0**, y recién después va a Tavily:

   | Para qué | Cómo, sin llave |
   |---|---|
   | ¿ya existe este paquete? | `registry.npmjs.org/-/v1/search?text=…` · PyPI · crates.io |
   | ¿existe este repo, y qué issues tiene? | la API pública de GitHub (60 req/hora sin token) |
   | leer una página concreta | `curl` |

2. **`tooling.md` gana la tabla de tres niveles** (§6 del parte), y dice cuál necesita llave.

3. **La regla, escrita en los dos:** *"empezá por lo que funciona sin configurar nada. Si el nivel 0
   ya contesta la pregunta, no hace falta el nivel 1."*

4. **La honestidad que no se puede saltear:** el nivel 0 **no** reemplaza al 1. Los registries
   contestan *"¿existe algo con este nombre?"*; **no** contestan *"¿alguien se queja de esto en
   Reddit?"*. Un brief que sólo tuvo nivel 0 puede mapear el panorama y **no** puede traer señales
   de demanda. Eso se dice en el brief, no se tapa.

### Cómo se sabe que quedó bien

Volver a correr **la corrida B**, misma idea semilla, mismo modelo:

```
B cita fuentes           →  era la herramienta. Causa raíz cerrada.
B sigue sin citar nada   →  era el modelo. Recién ahí se sabe, y es otro problema.
```

**Ojo con la vara:** pasa si B **cita**, no si B dice `hacelo`. El veredicto sigue siendo juicio.

---

## 5. La lista completa, en orden

De [`el-arnes-ciego.md`](el-arnes-ciego.md) §9. El orden no es de gusto: dos tareas se abarataron
solas al ponerlas después de otra.

| # | Qué | Dónde | Tamaño |
|---|---|---|---|
| **0** | **el nivel 0 entra en el skill** | §6 | **chico — es lo de mañana** |
| 1 | `evidencia.json` + la compuerta lee JSON en vez de prosa | §7 | mediano |
| 2 | el ② sale como paso propio, delegable | §5④ | mediano — **va con el 1, no sin él** |
| 3 | `sf doctor` ve las herramientas | §5① | mediano |
| 4 | toda parada entrega un **acta** | §5② | **casi gratis si el 1 está hecho** — creció de alcance el 07-09 |
| 5 | `sf install` garantiza el nivel 0 | §5③ | mediano — se abarató con §6 |
| 6 | el mapa entero de wayfinder | §5④ | **no ahora** — necesita datos de T3+ |

---

## 6. Las preguntas abiertas — que no se pierdan

Ninguna de éstas está contestada, y las tres pueden morder.

**① El bucle nuevo que aparece si se corta el ②.** Si el grill descubre que falta investigar algo,
¿el ② se reabre? Ahí nace un bucle, **y todo bucle nuevo necesita saber quién lo cierra.** Es la
misma trampa que ya está abierta en el bucle de revisión (㉑). No meter el corte sin contestar esto.

**② El agujero "el orquestador no delega" sigue abierto.** El tablero **no** lo cerró: el test del
pid que proponía `por-tramos.md` §3 no funciona —mide arquitectura del arnés, no delegación— y ya
quedó corregido ahí y en §8 del parte. Hay que buscarle otra medida. La candidata es
`SPECFORGE_DELEGADO` de `vecinos.md` §2, **y no está probada.**

**③ `compuerta.PRD` sigue siendo un `os.Stat`** y su comentario dice *"y tenga cuerpo"*, que es
falso. Esta corrida no lo explotó. **Lo gratis que se puede hacer ya: arreglar el comentario.**

---

## 7. Lo que NO hay que hacer mañana

Está en §10 del parte, con su umbral cada uno. En corto:

- **No apretar más la compuerta del brief.** Ya atrapó el caso real. Exigirle más antes de dar
  herramientas es castigar al modelo por algo que es culpa nuestra.
- **No construir `sf buscar`.** Buena idea, superficie grande, sin dato que la pida.
- **No traer el mapa entero de wayfinder.** Se trae el corte, no el mapa.
- **No tocar `entradas.go`** (lo del ticket / brownfield) hasta que T1 y T2 hayan corrido enteros.

---

## 8. Para arrancar en frío

```bash
cd ~/projects/workspace/personal/specforge
rtk git log --oneline -4          # dónde quedamos
sed -n '1,30p' specs/retomar.md   # esto

# el parte completo, si hace falta el detalle
sed -n '1,50p' specs/el-arnes-ciego.md

# y la evidencia de la corrida
cd ~/projects/workspace/personal/sf-banco && ./comparar.sh
```

**La frase que resume todo lo de hoy, por si se lee una sola línea:**

```
TODA REGLA QUE VIVE SÓLO EN EL SKILL ES UNA SUGERENCIA.
SÓLO LA COMPUERTA OBLIGA.
```
