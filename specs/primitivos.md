# Los primitivos — la tabla, contada

**Fecha:** 2026-09-13 · **Branch:** `claude/specforge-alternatives-direction-ae84u8`

> **Qué es esto.** El criterio de `tramo-1.md` §3 —*"el primitivo es el dueño ÚNICO de su
> método"*— aplicado a los 27 skills, con los consumidores **contados**, no estimados. Es el
> insumo de la extracción, no la extracción.
>
> **Regla de lectura.** Todo número sale de un `grep` sobre `skills/*/SKILL.md` y sobre
> `sf/internal/maquina/*.go`, corrido el 2026-09-13. Lo que es lectura mía está marcado.

---

## 1. Cómo se contó

Un consumidor es una de tres cosas, y las tres se miden distinto:

| fuente | qué cuenta | cómo se detecta |
|---|---|---|
| **máquina** | un estado que corre ese skill | el nombre en `internal/maquina/*.go` |
| **orden** | un skill que lo llama ejecutablemente | `Call the Skill tool with "X"` |
| **prosa** | un skill que lo nombra sin llamarlo | el nombre, sin la orden |

**`sfx-mapa` y `sfx-skill` no cuentan como consumidores.** Los dos nombran a los 27 porque son
catálogos: uno dibuja el mapa y el otro enseña a escribir skills. Contarlos daría 27 primitivos y
cero información.

**La prosa cuenta a medias, y es el hallazgo del 2026-09-08:** *"un modelo débil lee prosa y no
llama nada"*. Una mención en prosa es un consumidor **que hoy no consume**. Se cuenta para decidir
si el método se comparte; no se cuenta como que la composición exista.

---

## 2. Lo que está medido

### 2.1 Órdenes ejecutables — sólo existen en T1

```
sfp-scout   → sfx-grilling
sfx-grill-me→ sfx-grilling
sfx-grilling→ sfx-buscar · sfx-prototipo · sfx-think · sfx-vocabulario
```

**Seis órdenes en todo el repo, y las seis son el racimo del brief.** Del ⑪ para adelante —plan,
build, check, cierre— **no hay una sola orden ejecutable**: todo lo que comparten está escrito en
prosa o duplicado. O sea: `tramo-1.md` no está a medio aplicar, está aplicado a un séptimo de la
máquina.

### 2.2 El tablero de los 27

`M` = máquina · `O` = orden ejecutable · `P` = prosa. Total excluye los dos catálogos.

| skill | M | O | P | total | qué es |
|---|---|---|---|---|---|
| `sfx-think` | 2 | 1 | 2 | **5** | primitivo ✅ |
| `sfx-prototipo` | 0 | 1 | 2 | **3** | primitivo ✅ |
| `sfx-github` | 2 | 0 | 2 | **4** | primitivo ✅ (y es el único que toca red) |
| `sfx-grilling` | 0 | 2 | 0 | **2** | primitivo ✅ — el único con composición real |
| `sfx-buscar` | 0 | 1 | 1 | **2** | primitivo ✅ |
| `sfx-vocabulario` | 0 | 1 | 1 | **2** | primitivo ✅ |
| `sfx-tdd` | 1 | 0 | 1 | **2** | primitivo ✅ — pero ver §3.1 |
| `sfx-documenter` | 1 | 0 | 1 | **2** | primitivo ✅ |
| `sfx-journal` | 1 | 0 | 1 | **2** | primitivo ✅ |
| `sfx-prosa` | 0 | 0 | 2 | **2** | primitivo ✅ |
| `sfx-verificar` | 0 | 0 | 2 | **2** | puerta de usuario + primitivo |
| `sfx-triage` | 0 | 0 | 1 | **1** | ⚠️ un solo consumidor |
| `sfx-explain` | 0 | 0 | 0 | **0** | 🔴 ver §3.3 |
| `sfp-*` · `sf-*` (9) | — | — | — | — | compositores de estado: se quedan |
| `sfx-audit` `sfx-interrogar` `sfx-grill-me` | — | — | — | — | puertas de usuario |
| `sfx-mapa` `sfx-skill` | — | — | — | — | catálogos |

---

## 3. Los cuatro métodos que están escritos dos y tres veces

**Acá está la extracción de verdad.** No son skills que haya que mover: son métodos **sin dueño**,
copiados adentro de los compositores. Ninguno tiene hoy un archivo propio.

### 3.1 «¿Qué podría estar mal y aún así pasar?» — el que está escrito dos veces, casi textual

| dónde | texto |
|---|---|
| `sf-plan:174` (⑯) | *"What could be wrong in the implementation and still pass this test?"* |
| `sf-build:95` (⑲) | *"What could be wrong in this implementation and still pass the tests I have?"* |
| `sf-check` (㉒) | la corrida de mutantes — **la misma pregunta, automatizada** |

**Tres consumidores, cero dueños.** Y `sfx-tdd` **no** es el dueño: `sfx-tdd` posee el *ciclo*
(RED → GREEN → REFACTOR), que es otra cosa. Encontrar el test que falta y correr el ciclo son dos
métodos distintos que hoy viven en el mismo lugar por costumbre.

> **Es el primitivo más valioso del repo y no existe.** Nombre propuesto: `sfx-mutante`.

### 3.2 La vara antes de las opciones

`sf-plan` ⑫a lo tiene entero: 3 a 6 criterios `V-1…`, escritos **antes** que la primera opción,
con el motivo (*"una vara escrita después describe la opción que ya elegiste"*). Y:

- `tramo-1.md` le pone una vara al brief — *"la vara pasa a ser de evidencia"* (`736c92c`)
- `sfx-think` tiene el debate (Step 2) y la convergencia (Step 3) **sin la mecánica de la vara**

**Tres consumidores, y el método vive adentro de un compositor.** Es exactamente el bug que
`tramo-1.md` §3 describe para la entrevista, un tramo más adelante.

### 3.3 El criterio que se puede testear

R5 dice *"si no se puede escribir el test, no es un criterio de aceptación"*. El método está en
`sfp-backlog` (EARS + la rúbrica, Steps 2–3). Y lo consumen:

- `sf-plan` ⑭⑮ — convierte cada criterio en tests y los cuenta
- `sf-check` — la matriz camina criterio por criterio
- `sfp-backlog` — los escribe

**Quien escribe el criterio y quien lo tiene que poder testear son dos skills distintas**, y la
regla que las une vive en una sola.

### 3.4 La procedencia

`sfx-buscar` posee *"procedencia"*; `sfp-po:49` dice *"mind the provenance"*; `sf-check:52` tiene
*"la escalera: `cumple` tiene que decir CÓMO lo sabés"*; `evidencia.json` de T1 lo cuenta. Cuatro
lugares, una idea: **una afirmación vale por su fuente**. Hoy cada uno la escribe a su manera —y
es la causa raíz del arnés ciego.

---

## 4. Los dos hallazgos sueltos

**① `sfx-explain` tiene cero consumidores.** No lo corre la máquina, no lo llama nadie, no es
user-invoked. O sea: es model-invoked, su `description` está siempre en contexto, y **compite en
la carrera de ruteo sin que nadie lo necesite**. Es la misma forma del bug medido el 2026-09-08 en
opencode. Por R2 se borra o se vuelve puerta de usuario.

**② `sfx-triage` tiene uno solo** (`sfp-backlog`, y en prosa). Lectura mía, no medición: el triage
es el método natural de la puerta *"llego con un bug"*, que hoy no existe. Si esa puerta se abre,
pasa a dos y se queda. Si no, es una sección de `sfp-backlog`.

---

## 5. Lo que NO se extrae

- **Las compuertas.** El primitivo es el método; la compuerta es el árbitro. Un método que se
  mete en un skill deja de obligar: *toda regla que vive sólo en el skill es una sugerencia*.
- **Los 9 compositores de estado.** No son métodos: son el pegamento entre un estado y su
  artefacto.
- **Las reglas R1–R7.** Una regla no es un primitivo — no la podés llamar.

---

## 6. Qué queda para decidir

1. Si los cuatro métodos de §3 se extraen ya o después de que T1 vuelva a correr.
2. Si el eje del ⑪ para adelante se compone con órdenes ejecutables, como T1 — hoy son cero.
3. Si se abre la puerta *"llego con un bug"*, que es la que decide el destino de `sfx-triage`.
