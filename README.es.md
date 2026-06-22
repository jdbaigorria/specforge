<p align="center">
  <img src="assets/specforge-logo.png" alt="SpecForge" width="380" />
</p>

# SpecForge

Framework de Spec-Driven Development. La especificación es el producto — el código es un subproducto regenerable.

Un pipeline de 4 skills por feature + `sf-audit` para revisión transversal del proyecto + skills de soporte. Revelación progresiva. Gate humano en cada artefacto. Cero ceremonia sin propósito.

**¿Recién llegás?** Leé primero [el modelo mental](docs/mental-model.md) — una página sobre cómo piensa SpecForge.

**Instalación:** ver [INSTALL.md](INSTALL.md). **Licencia:** [MIT](LICENSE).

---

## Tabla de contenidos

- [Cómo funciona](#cómo-funciona)
- [La capa determinista (CLI `sf` + hooks)](#la-capa-determinista-cli-sf--hooks)
- [Arquitectura](#arquitectura)
- [Referencia de skills](#referencia-de-skills)
- [Flujo de artefactos](#flujo-de-artefactos)
- [Formato de specs: notación EARS](#formato-de-specs-notación-ears)
- [Ciclo de vida de una feature](#ciclo-de-vida-de-una-feature)
- [Backprop: cómo el spec aprende](#backprop-cómo-el-spec-aprende)
- [Ejemplo 1: Greenfield — CLI desde cero](#ejemplo-1-greenfield--cli-desde-cero)
- [Ejemplo 2: Brownfield — agregar features a código existente](#ejemplo-2-brownfield--agregar-features-a-código-existente)
- [Detección de resync](#detección-de-resync)
- [FAQ](#faq)

---

## Cómo funciona

```
  sf-init ──▶ sf-propose ──▶ sf-build ──▶ sf-check
  scaffold    requirements    plan+ejecutar  validar
  + contexto  diseño          wave por wave  + archivar
              tareas

  El loop por feature: 4 skills. Cada uno produce artefactos. Gate humano 🔴 en cada artefacto.
```

Este es el loop por feature. Al lado hay dos piezas más: `sf-audit` corre una
revisión adversarial de todo el proyecto (constitución vs realidad, consistencia
entre features), y un conjunto de [skills de soporte](SUPPORT-SKILLS.es.md)
(`sfx-think`, `sfx-triage`, `sfx-grill-me`, `sfx-tdd`, `sfx-documenter`, y más) complementan el
pipeline sin ser parte de él.

Flujo detallado con gates:

```
sf-init
  ├─ scaffolding                          → automático
  ├─ constitución (greenfield)            → 🔴 GATE
  ├─ onboard (brownfield)                 → 🔴 GATE
  └─ constitución (brownfield)            → 🔴 GATE

sf-propose
  ├─ requirements.md                      → 🔴 GATE
  ├─ design.md                            → 🔴 GATE
  └─ tasks.md                             → 🔴 GATE

sf-build
  ├─ plan de ejecución                    → 🔴 GATE
  ├─ wave 0 ejecución                     → 🔴 GATE
  ├─ wave 1 ejecución                     → 🔴 GATE
  └─ wave N...                            → 🔴 GATE

sf-check
  ├─ review + veredicto                   → 🔴 GATE
  ├─ APPROVE → archivo (automático)
  └─ REVISE  → vuelve a sf-build con feedback
```

El loop REVISE es lo que hace el flujo iterativo, no waterfall. Cuando check
encuentra gaps, envía la feature de vuelta a build con correcciones específicas.
Si el spec estaba mal, el usuario lo edita directamente y la detección de resync
propaga los cambios en cascada.

---

## La capa determinista (CLI `sf` + hooks)

Los skills son la mitad **cooperativa**: un LLM produce specs, juzga código,
escribe prosa. Pero el instruction-following se degrada a medida que el contexto
se llena — pasado ~50% el agente empieza a ignorar el workflow que le dijeron que
siga. Por eso SpecForge trae una mitad **determinista** que no depende de la buena
voluntad del modelo: un CLI chico en Go, **`sf`**, y un set de **hooks** por harness.

> **El reparto de tareas:** el LLM *produce y juzga*; `sf` *persiste, valida,
> computa y renderiza*; los hooks *fuerzan e inyectan* en eventos del harness.
> Ninguno hace el trabajo del otro.

**JSON-first.** Cada artefacto es un `.json` **fuente** más un `.md` **render**.
Vos (o el LLM) producís el JSON; `sf` lo valida y genera el Markdown. Una sola
fuente de verdad — el Markdown no puede divergir de los datos porque deriva de ellos.

**Dos tiers de enforcement** (esta distinción recorre todo lo de abajo):

| Tier | Qué protege | Cómo |
|------|-------------|------|
| **Estructural** (hermético) | orden de gates, schema, dependencias, flujo serial | un hook *deniega* la tool call — no se puede saltar |
| **Calidad** (cooperativo) | "¿esto está bien / es mínimo / está alineado?" | un subagente fresco juzga; el veredicto es un *nudge*, registrado para el humano |

Podés volver inalcanzables los estados ilegales (estructural). No podés forzar
buen contenido a existir (calidad) — así que la calidad la sube un verificador,
no se garantiza. SpecForge es honesto sobre cuál es cuál.

### El CLI `sf`

`sf` es determinista: mismas entradas, mismas salidas, sin llamadas al modelo.
Es seguro de correr en cualquier lado y es lo que consultan los hooks.

**Validar y renderizar artefactos (la vía de escritura JSON-first).**

```console
$ echo '{"feature":"add-task-crud","tasks":[
    {"id":"T1","title":"TaskModel + JsonStorage","requirement_refs":["R5"],"depends_on":[]},
    {"id":"T2","title":"add command","requirement_refs":["R1"],"depends_on":["T1"]}
  ]}' | sf save tasks --feature=add-task-crud --json -
saved specforge/features/add-task-crud/tasks.json (+ rendered .../tasks.md)
```

`sf save` valida el JSON (ids únicos, refs en forma `R#`/`C#`, grafo de
dependencias acíclico) y **solo si pasa** escribe el `.json` canónico + renderiza
el `.md`. La entrada inválida se rechaza con exit 2 y **nada toca el disco** — un
artefacto malformado nunca existe. (`sf <artefacto> validate|render` hacen las dos
mitades sueltas, para los seis artefactos: constitution, requirements, design,
tasks, plan, review.)

**Computar el layout de waves desde las dependencias de las tasks.**

```console
$ sf plan compute --feature=add-task-crud
computed plan for add-task-crud: 5 task(s) → 3 wave(s)
```

Las tasks son una lista **plana**; cada una declara `depends_on`. La agrupación
en waves es un **layering topológico** determinista — wave 0 = tasks sin deps,
wave N = `1 + max(wave de sus deps)`. Así que `sf` *computa* las waves; el LLM no
las ordena a mano. `sf plan validate` después enforza el guard `wave(task) >
wave(sus deps)` como error. (Tasks en la misma wave son independientes →
paralelizables.)

**Leer el estado del proyecto y de los gates.**

```console
$ sf status                       # tabla de salud: cada feature, fase, drift, bloqueos
$ sf gate status --feature=X      # el ledger de gates de aprobación humana de una feature
$ sf doctor --drift               # specs archivadas cuyo anchor de código desapareció (lee trace.json)
$ sf trace verify --feature=X     # la matriz requirement → código → test vs el repo
$ sf lint                         # la consistencia de la propia suite de skills
```

**Emitir los slices de contexto que inyectan los hooks — el "cerebro".**

```console
$ sf state current
{ "feature": "add-task-crud", "status": "building",
  "phase": "build", "wave": 1, "last_approved_gate": "wave-0" }
```

`sf state current` es una **función pura de `features.json`**: la feature activa
(flujo serial), la fase derivada como *último-gate-aprobado + 1*, la wave desde
`plan.json`. Cero estado guardado — nada que mantener en sync (sin "segunda verdad").

```console
$ sf context current --breadcrumb
SpecForge: feature `add-task-crud`, fase `build` (wave 1) · slice completo: `sf context current`

$ sf context for-wave --feature=add-task-crud --n=1   # la semilla para un subagente de build
$ sf context for-judge --phase=design --feature=X     # artefacto + SOLO los principios que aplican a esa fase
```

`context current` es el slice mínimo del paso actual — un breadcrumb barato, o el
JSON completo. `for-wave` siembra un subagente de build con una wave. `for-judge`
le da al auditor de calidad exactamente el artefacto más los principios de la
constitución cuyo `applies_to` incluye esa fase — nada más.

**Registrar veredictos de calidad y lecciones durables.**

```console
$ echo '{"phase":"design","verdicts":[
    {"rule":"P1","result":"pass","citation":"ningún componente hace llamadas de red"},
    {"rule":"I1","result":"fail","citation":"el endpoint X no maneja errores"}
  ]}' | sf gate record-verdict --feature=X --phase=design
recorded fail verdict for X/design (2 rule(s))      # exit 3 = FAIL registrado → el hook puede nudgear

$ echo '{"feature":"X","lessons":[
    {"context":"reimplementamos parseo de fechas","rule":"usar la stdlib"}
  ]}' | sf journal add --feature=X --json - --bridge-icm
journaled specforge/journal/2026-06-20-X.json (+ rendered .md)
staged for commit (NOT committed — that's your call)
bridged to ICM (topic specforge-journal)
```

`sf gate record-verdict` apendea un veredicto de auditoría de fase a `audit.json`
(separado de los gates humanos). `sf journal add` persiste lecciones durables a un
journal git-trackeado — su propia memoria, no una dependencia dura de ninguna
herramienta externa; `--bridge-icm` opcionalmente espeja a
[ICM](https://github.com/rtk-ai/icm) si está presente. **Stagea pero nunca
commitea** — el commit es tu decisión.

### Los hooks (por harness — adapter de Claude Code)

Los hooks disparan en los eventos del harness y llaman a `sf`. Esta es la capa que
hace al enforcement independiente del modelo. El core (`sf`) es portable; el
adapter de hooks es por-harness (un harness nuevo = un adapter nuevo, no lógica
nueva). Cada hook es **fail-open** (cualquier error → permitir) y un **no-op fuera
de un proyecto SpecForge**.

| Evento | Qué hace el hook |
|--------|------------------|
| `SessionStart` | inyecta la sesión retomada + invariantes del proyecto + learnings consolidados |
| `UserPromptSubmit` | inyecta el slice del paso actual (`sf context current`) para que la spec no se diluya — breadcrumb barato cada turno, slice completo on-step-change |
| `PreToolUse` (Write/Edit) | **gates estructurales:** deniega escribir `design` antes de aprobar `requirements`, `tasks` antes de `design`, `plan` antes de `tasks`; deniega el `requirements` de una 2da feature mientras otra está activa (serial); deniega editar estado de máquina |
| `Stop` | **conciliador de memoria:** si una feature se archivó sin entrada en el journal, nudgea una vez para capturar lecciones |
| `PreCompact` | marca la sesión para re-inyectar el slice completo el próximo turno — tras compactar, los slices previos se resumieron |
| `SessionEnd` | timestampea un marker de continuidad |

### El tier calidad: auditor de fase

Los gates estructurales son herméticos pero solo chequean *orden y schema*, no
*calidad*. Para calidad, `sf-check --phase=<fase>` corre un **juez adversarial en
un subagente fresco** — el contexto limpio escapa la degradación que sufre una
sesión larga. Lo alimenta `sf context for-judge`, juzga cada principio mapeado
`pass`/`fail` con cita, y persiste el veredicto vía `sf gate record-verdict`.
Opt-in y gobernado por config:

```jsonc
// constitution.json
{
  "principles": [
    { "id": "P-min", "statement": "Código mínimo: subir la escalera antes de escribir código nuevo",
      "applies_to": ["design", "build"] }
  ],
  "audit": { "phase": "off" | "nudge" | "block" },   // default off
  "build": { "mode": "inline" | "single" | "per-wave" } // default inline
}
```

- **`audit.phase`** — `off` (saltea), `nudge` (muestra los fallos, no bloquea),
  `block` (frena el gate ante un fail).
- **`build.mode`** — `inline` (las waves corren en el contexto principal, clásico),
  `single` (todo el build en un subagente fresco), `per-wave` (un subagente fresco
  por wave, para features grandes — cada wave arranca limpia, con checkpoint
  automático entre waves).

Los principios de calidad como **`P-min`** (código mínimo) son simplemente
principios de la constitución con `applies_to` — el auditor de fase los chequea
gratis, sin vía especial. Este es el tier cooperativo: un juez fresco sube el piso
de calidad; no lo garantiza.

---

## Arquitectura

### Estructura de directorios

Una sola raíz visible, `specforge/`, contiene todo. Los artefactos de revisión
(requirements, design, review) quedan visibles para que funcionen los gates
humanos; el estado de máquina se oculta en `specforge/.state/`.

```
mi-proyecto/
├── specforge/                       # Raíz única visible de SpecForge
│   ├── features.json                # Registro de features (tracker de estado)
│   ├── constitution.md              # Principios del proyecto + identidad
│   ├── history.md                   # Log del proyecto (append-only)
│   ├── roadmap.md                   # Roadmap de features
│   ├── learnings.md                 # Aprendizajes consolidados y anclados a evidencia (inyectados cada sesión)
│   ├── features/
│   │   └── agregar-tareas/
│   │       ├── requirements.md
│   │       ├── design.md
│   │       ├── tasks.md
│   │       ├── review.md
│   │       ├── decisions/           # Decisiones técnicas complejas (opcional)
│   │       └── progress/
│   │           ├── plan.md
│   │           ├── wave-0.md
│   │           └── wave-1.md
│   ├── archive/
│   │   └── 2026-05-12-agregar-auth/
│   ├── audits/                      # Reportes de sf-audit
│   ├── context/                     # Contexto del proyecto + outputs de skills (visible)
│   │   ├── project.md               # Stack, arquitectura
│   │   ├── conventions.md           # Convenciones de código
│   │   ├── compact-rules.md         # Reglas condensadas para sub-agentes
│   │   └── thinks/ triages/ briefs/ grills/ journal/ …  # artefactos de skills de soporte
│   └── .state/                      # Estado de máquina oculto (no se edita a mano)
│       └── session.md               # Caché/recovery de sesión (no source of truth)
│
└── src/                             # Tu código
```

### Estructura de skills (revelación progresiva)

Cada skill tiene un SKILL.md lean (<200 líneas) y carga capacidades bajo demanda
desde `references/`. Esto mantiene el contexto pequeño — el agente solo carga lo
que el paso actual necesita.

```
sf-init/
├── SKILL.md
├── references/
│   ├── constitution.md              # Greenfield: guía de conversación
│   └── onboard.md                   # Brownfield: análisis del codebase
└── templates/
    ├── constitution.tmpl.md
    ├── project.tmpl.md
    └── conventions.tmpl.md

sf-propose/
├── SKILL.md
├── references/
│   ├── design-first.md              # Flujo alternativo: arquitectura → requirements
│   ├── from-code.md                 # Brownfield: inferir specs del código
│   ├── clarify.md                   # Refinamiento post-generación
│   ├── research.md                  # Documentación de decisiones técnicas
│   └── ears-notation.md             # Referencia de sintaxis EARS
└── templates/
    ├── requirements.tmpl.md
    ├── design.tmpl.md
    └── tasks.tmpl.md

sf-build/
├── SKILL.md
├── references/
│   ├── task-planning.md             # Organizar tareas en waves
│   └── wave-execution.md            # Estrategia de ejecución por wave
└── templates/
    └── progress.tmpl.md

sf-check/
├── SKILL.md
├── references/
│   ├── backprop.md                  # Patrón 3x → promoción a invariante
│   └── archive.md                   # Procedimiento de sync + cierre
└── templates/
    └── review.tmpl.md
```

### Modelo de ejecución

Los 4 skills del pipeline corren **inline** — en el contexto principal de la
conversación. No se delegan a sub-agentes porque cada artefacto tiene un gate
humano que requiere interacción. Un sub-agente corre en un contexto aislado y no
puede frenar a pedir aprobación, así que todo lo que tiene gate debe ser inline.

Los skills de soporte siguen la misma regla, decidida por **interactividad, no
por tier**:

- **Con gate o iterativos** (`sfx-grill-me`, `sfx-tdd`) → inline.
- **Transform puro** — entra X, sale Y, sin turno humano en el medio
  (`sfx-documenter`, `sfx-explain`, `sfx-aws-architect`, `sfx-data-engineer`, `sfx-triage`,
  `sfx-github`) → pueden delegarse a un sub-agente **donde el harness lo soporte**,
  con fallback a inline. La delegación es una optimización opcional y per-harness,
  no parte del core portable.

Principio anti-teléfono-descompuesto: los artefactos viven en disco. Cuando un
skill necesita contexto de un artefacto anterior, lee el archivo — no depende
del historial de conversación. El context window lleva referencias, no payloads.

---

## Referencia de skills

### sfp-scout (tier de producto, opcional)

El front-end desde-cero. De-riskea una idea difusa antes de que sea un proyecto —
investiga el landscape, la stress-testea, decide proceed/pivot/kill. Solo
greenfield; brownfield va directo a `sf-init`.

| | |
|---|---|
| **Triggers** | `sfp-scout`, "¿debería construir X?", "¿vale la pena esto?", "de-riskeá esta idea" |
| **Necesita** | los MCPs de research (web/GitHub/docs) — sin ellos no fabrica research |
| **Produce** | `specforge/product/<slug>/` — research, discovery brief (con ids `PR#`), decision log |
| **Handoff** | en *proceed* → `sf-init --from <brief>`; los ids `PR#` fluyen al roadmap |

De-risk, no validación — junta evidencia y expone riesgo; no puede probar demanda.
Sus product requirements (`PR#`) están en la cima de la espina de trazabilidad:
`PR# → feature → R# → task → code → test`.

---

### sf-init

Scaffoldea el proyecto y genera el contexto fundacional.

| | |
|---|---|
| **Triggers** | `sf-init`, "arrancar proyecto nuevo", "agregar specforge a este proyecto" |
| **Flags** | `--from <path>` (importar PRD/brief), `--path <dir>` (sub-proyecto en monorepo) |
| **Detecta** | Greenfield vs brownfield (interactivo para casos ambiguos como monorepos) |

**Flujo greenfield:**
1. Scaffolding de directorios (automático)
2. Conversación de constitución: identidad, principios, constraints, anti-goals → 🔴 GATE
3. Generar `specforge/context/project.md` + `specforge/context/conventions.md` del contexto → 🔴 GATE

**Flujo brownfield:**
1. Scaffolding de directorios (automático)
2. Onboard: analizar codebase → `specforge/context/project.md` + `specforge/context/conventions.md` → 🔴 GATE
3. Conversación de constitución: principios y anti-goals (más corta, stack ya conocido) → 🔴 GATE

**Produce:**
- `specforge/constitution.md` — identidad del proyecto, principios, constraints, anti-goals
- `specforge/features.json` — registro vacío
- `specforge/history.md` — log del proyecto (primera entrada)
- `specforge/context/project.md` — contexto de stack y arquitectura
- `specforge/context/conventions.md` — estándares de código

---

### sf-propose

Genera la especificación completa de una feature: requirements, diseño, tareas.

| | |
|---|---|
| **Triggers** | `sf-propose <nombre>`, "especificá", "agregar feature", "spec esto" |
| **Flags** | `--design-first` (arquitectura → requirements), `--from-code` (reverse-engineer) |
| **Modos** | Requirements-first (default), Design-first, From-code |

**Flujo requirements-first (default):**
1. Conversación: qué, por qué, quién, límites
2. Generar `requirements.md` con notación EARS → 🔴 GATE
3. Generar `design.md` (secciones condicionales según complejidad) → 🔴 GATE
4. Generar `tasks.md` con waves + matriz de trazabilidad → 🔴 GATE
5. Registrar feature en `features.json` (status: `approved`)

**Flujo design-first** (`--design-first`):
Invierte pasos 2 y 3 — diseño primero, luego derivar requirements de lo que la
arquitectura puede entregar. Para proyectos donde las restricciones técnicas definen el alcance.

**Flujo from-code** (`--from-code`):
Reverse-engineerea specs del código existente. Crea una feature `_baseline` con
requirements `[INFERRED]`. Sin tasks.md (ya está implementado).

**Produce por feature:**
- `specforge/features/<nombre>/requirements.md` — requirements EARS + criterios de aceptación
- `specforge/features/<nombre>/design.md` — arquitectura, componentes, decisiones
- `specforge/features/<nombre>/tasks.md` — waves + matriz de trazabilidad

**Las secciones del design son condicionales.** Solo se incluyen las relevantes a la
complejidad de la feature. Un flag de CLI no necesita Security Considerations. Un
endpoint de pagos sí. Las secciones que dirían "N/A" se omiten para evitar ruido.

**Post-generación:** El usuario puede invocar clarify en cualquier momento
(carga `references/clarify.md`). Escanea ambigüedades, gaps de completitud e inconsistencias.

---

### sf-build

Planifica la ejecución e implementa wave por wave.

| | |
|---|---|
| **Triggers** | `sf-build <nombre>`, "construí", "implementá", "empezá a buildear" |
| **Prerequisito** | La feature debe estar en status `approved` |
| **Resumible** | Si se pausó, lee `progress/` para retomar donde quedó |

**Flujo:**
1. Generar plan de ejecución desde las waves de `tasks.md` → 🔴 GATE
2. Por cada wave:
   a. Ejecutar todas las tareas de la wave
   b. Logear progreso en `progress/wave-<n>.md`
   c. Presentar resultados → 🔴 GATE
3. Todas las waves completas → status pasa a `checking`

**Recuperación de errores (3 niveles):**
- **Menor** (typo, import mal): fix inline, anotar en el log
- **Incompatibilidad con diseño** (no se puede implementar como está especificado): pausar wave, presentar opciones al usuario → 🔴 GATE
- **Dependencia bloqueante** (output de wave anterior está mal): pausar, escalar al usuario

**Regla clave:** Nunca modificar specs durante build. Si los specs necesitan cambios,
pausar y escalar. Los specs son el contrato; build lo cumple.

**Produce por wave:**
- `specforge/features/<nombre>/progress/plan.md` — plan de ejecución
- `specforge/features/<nombre>/progress/wave-<n>.md` — qué se hizo, decisiones, issues
- `tasks.md` actualizado — tareas marcadas `[x]` al completarse

---

### sf-check

Valida la implementación contra los specs. Archiva si aprueba.

| | |
|---|---|
| **Triggers** | `sf-check <nombre>`, "check", "validá", "review" |
| **Prerequisito** | La feature debe estar en status `checking` |
| **Incluye** | Trazabilidad, gap analysis, compliance con constitución, backprop |

**Flujo:**
1. Análisis de trazabilidad: cada R# → tarea → implementación → test
2. Gap analysis: implementaciones faltantes, tests, desviaciones del diseño, código huérfano
3. Compliance con constitución: validar contra principios
4. Veredicto: APPROVE / APPROVE WITH NOTES / REVISE
5. Presentar review → 🔴 GATE
6. Si APPROVE → archivar automáticamente

**Veredictos:**
- **APPROVE** — todo implementado + testeado, sin violaciones
- **APPROVE WITH NOTES** — implementado, gaps menores documentados para el futuro
- **REVISE** — gaps significativos, vuelve a `sf-build` con correcciones específicas

**El usuario puede anular el veredicto.** Si no está de acuerdo con la evaluación,
puede anular. La anulación queda logueada en el review.

**Produce:**
- `specforge/features/<nombre>/review.md` — matriz de trazabilidad, gaps, veredicto
- `specforge/archive/<fecha>-<nombre>/` — archivo completo de la feature (si approve)
- `specforge/history.md` actualizado — entrada de completitud
- `specforge/constitution.md` actualizado — si backprop promueve un nuevo invariante

---

### sf-audit

Auditoría adversarial de todo el proyecto. No es parte del loop por feature —
toma distancia y revisa el proyecto entero: constitución vs realidad,
consistencia entre features, drift y gaps acumulados.

| | |
|---|---|
| **Triggers** | `sf-audit`, "auditá el proyecto", "¿sigue siendo cierta la constitución?" |
| **Alcance** | Todas las features + constitución, no una feature sola |
| **Produce** | `specforge/audits/<fecha>.md` — hallazgos, severidad, recomendaciones |

---

### sf-amend

Modificar una feature ya enviada sin forkear un spec paralelo. Los specs
archivados son **documentos vivos** — `sf-amend` corre un mini-pipeline delta
(propose-delta → gate → build → check) que edita los requirements, design, tasks
y `trace.json` existentes en su lugar.

| | |
|---|---|
| **Triggers** | `sf-amend <feature>`, "cambiar la feature enviada", "el spec archivado quedó viejo" |
| **vs sf-propose** | amend cambia una capacidad existente; propose crea una nueva |
| **Edita en su lugar** | una sola matriz de trazabilidad que evoluciona por feature — nunca un spec paralelo |

Combinalo con detección de drift: `sf doctor --drift`
lee el `trace.json` de cada feature archivada y marca los requirements
cuyo anclaje de código desapareció o cuyo test falla — así te enterás de que el
spec divergió antes de que se vuelva mentira.

---

### Skills de soporte

Skills standalone que complementan el pipeline pero no son parte de él. Funcionan
sin `specforge/` inicializado y producen artefactos en `specforge/context/`. Ver
[SUPPORT-SKILLS.es.md](SUPPORT-SKILLS.es.md) para la referencia completa:
`sfx-think`, `sfx-triage`, `sfx-grill-me`, `sfx-tdd`, `sfx-documenter`, `sfx-explain`,
`sfx-aws-architect`, `sfx-data-engineer`, `sfx-github`, `sfx-journal`.

---

## Flujo de artefactos

**Quién produce qué:**

```
sf-init produce:
  specforge/constitution.md
  specforge/context/project.md
  specforge/context/conventions.md

sf-propose produce (por feature):
  specforge/features/<nombre>/requirements.md
  specforge/features/<nombre>/design.md
  specforge/features/<nombre>/tasks.md

sf-build produce (por feature):
  specforge/features/<nombre>/progress/plan.md
  specforge/features/<nombre>/progress/wave-<n>.md

sf-check produce (por feature):
  specforge/features/<nombre>/review.md
  specforge/archive/<fecha>-<nombre>/        (si APPROVE)
```

**Quién lee qué:**

| Artefacto | Creado por | Leído por |
|-----------|-----------|-----------|
| constitution.md | sf-init | sf-propose (constraints), sf-check (compliance) |
| specforge/context/project.md | sf-init | sf-propose (contexto de stack), sf-build (convenciones) |
| specforge/context/conventions.md | sf-init | sf-build (estándares de código) |
| requirements.md | sf-propose | sf-build (trazabilidad), sf-check (validación) |
| design.md | sf-propose | sf-build (guía de arquitectura), sf-check (adherencia) |
| tasks.md | sf-propose | sf-build (ejecución), sf-check (trazabilidad) |
| progress/*.md | sf-build | sf-check (audit trail) |
| review.md | sf-check | archivo (el veredicto determina si se archiva) |
| features.json | sf-init | todos los skills (gate de estado) |
| history.md | sf-init | sf-check (tracking de patrones para backprop) |

---

## Formato de specs: notación EARS

SpecForge usa EARS (Easy Approach to Requirements Syntax) para requirements inequívocos:

```
WHEN [trigger] THE SYSTEM SHALL [comportamiento]          ← event-driven
WHILE [estado] THE SYSTEM SHALL [comportamiento]          ← state-driven
IF [condición no deseada] THEN THE SYSTEM SHALL [acción]  ← manejo de errores
THE SYSTEM SHALL [comportamiento]                         ← siempre activo
```

Un requirement por statement. Voz activa. Específico y testeable.
Referencia completa en `sf-propose/references/ears-notation.md`.

---

## Ciclo de vida de una feature

```json
// specforge/features.json
{
  "features": [
    {
      "name": "user-auth",
      "status": "done",
      "workflow": "requirements-first",
      "created": "2026-05-10",
      "completed": "2026-05-12"
    }
  ]
}
```

Estados: `pending` → `proposing` → `approved` → `building` → `checking` → `done`

Cada transición la dispara un skill al completar su trabajo:
- sf-propose pone `approved` después de que los 3 artefactos pasan el gate
- sf-build pone `checking` después de completar todas las waves
- sf-check pone `done` después de APPROVE + archivo

---

## Backprop: cómo el spec aprende

Cuando el mismo tipo de problema aparece en 3+ features, se convierte en un invariante del proyecto.

```
Feature A: endpoint API sin manejo de errores      ← ocurrencia 1
Feature B: comando CLI crashea con input malo       ← ocurrencia 2
Feature C: webhook handler ignora timeout           ← ocurrencia 3 → PROMOVER

→ Nuevo invariante en constitution.md:
  "Toda interfaz externa debe tener manejo de errores explícito."
```

El tracking vive en `specforge/history.md`:

```markdown
## Issues Recurrentes
| Patrón del Issue | Ocurrencias | Features | Estado |
|------------------|-------------|----------|--------|
| Falta manejo de errores | 3 | auth, cli, api | → INVARIANTE |
```

Después de la promoción, sf-check valida cada feature futura contra el nuevo invariante.

---

## Ejemplo 1: Greenfield — CLI desde cero

**Objetivo:** CLI de gestión de tareas con Python, Typer, Rich.

```
> sf-init

  Scaffolding...
  ✓ specforge/ creado
  ✓ specforge/context/ creado
  Proyecto greenfield detectado.

  Definamos tu proyecto.
  ¿Qué estás construyendo? → "Un CLI para gestionar tareas"
  ¿Para quién? → "Desarrolladores que quieren un todo rápido en la terminal"
  ¿Qué problema resuelve? → "Las herramientas existentes son pesadas, quiero algo mínimo"

  ¿Principios?
  → "Velocidad sobre features — MVP rápido"
  → "Terminal-native — solo comandos, nada de TUI"
  → "Local-first — archivo JSON, sin servidor"

  ¿Constraints?
  → "Python 3.12, Typer, Rich. Debe funcionar en macOS y Linux."

  ¿Anti-goals?
  → "No es un project manager — nada de equipos, boards ni sprints"
  → "No sincroniza — un solo dispositivo"

  🔴 GATE: constitution.md
  Usuario: "Aprobado"

  🔴 GATE: project.md + conventions.md
  Usuario: "Aprobado"

  ✓ Inicializado. Ejecutá sf-propose <nombre-feature> para arrancar.
```

```
> sf-propose agregar-tareas-crud

  ¿Qué hace esta feature? → "Agregar, listar, completar y borrar tareas"
  ¿Quién la usa? → "El desarrollador, desde la terminal"
  ¿Límites? → "Sin fechas límite, sin prioridades, sin categorías"

  🔴 GATE: requirements.md
  R1: Agregar tarea (WHEN el usuario ejecuta `tasks add "texto"` THE SYSTEM SHALL crear tarea)
  R2: Listar tareas (WHEN el usuario ejecuta `tasks list` THE SYSTEM SHALL mostrar tabla)
  R3: Completar tarea (WHEN el usuario ejecuta `tasks done <id>` THE SYSTEM SHALL marcar)
  R4: Borrar tarea (WHEN el usuario ejecuta `tasks rm <id>` THE SYSTEM SHALL eliminar)
  R5: Persistir (THE SYSTEM SHALL almacenar tareas en ~/.tasks/tasks.json)
  Usuario: "Aprobado"

  🔴 GATE: design.md
  Componentes: TaskModel, JsonStorage, comandos CLI
  Sin secciones de Security/Performance (no relevantes para esta feature)
  Usuario: "Aprobado"

  🔴 GATE: tasks.md
  Wave 0: T1 TaskModel + JsonStorage → R5
  Wave 1: T2 comando add → R1, T3 comando list → R2
  Wave 2: T4 comando done → R3, T5 comando rm → R4
  Trazabilidad: 5/5 requirements cubiertos ✓
  Usuario: "Aprobado"

  ✓ Feature aprobada. Ejecutá sf-build agregar-tareas-crud.
```

```
> sf-build agregar-tareas-crud

  🔴 GATE: Plan de ejecución
  Wave 0: models.py + storage.py (2 archivos, complejidad baja)
  Wave 1: add.py + list.py (2 archivos, media)
  Wave 2: done.py + rm.py (2 archivos, baja)
  Usuario: "Aprobado"

  Ejecutando Wave 0...
  ✓ T1: Creados models.py, storage.py, test_storage.py
  🔴 GATE: Resultados Wave 0
  Usuario: "Aprobado"

  Ejecutando Wave 1...
  ✓ T2: Creado commands/add.py
  ✓ T3: Creado commands/list.py con tabla Rich
  🔴 GATE: Resultados Wave 1
  Usuario: "Aprobado"

  Ejecutando Wave 2...
  ✓ T4: Creado commands/done.py
  ✓ T5: Creado commands/rm.py
  🔴 GATE: Resultados Wave 2
  Usuario: "Aprobado"

  ✓ Construido. Ejecutá sf-check agregar-tareas-crud.
```

```
> sf-check agregar-tareas-crud

  Trazabilidad: 5/5 requirements implementados ✓
  Tests: 5/5 requirements con tests ✓
  Constitución: todos los principios respetados ✓
  Veredicto: APPROVE

  🔴 GATE: review.md
  Usuario: "Aprobado"

  ✓ Archivado en specforge/archive/2026-05-12-agregar-tareas-crud/
```

---

## Ejemplo 2: Brownfield — agregar features a código existente

**Objetivo:** CLI de tareas existente. Agregar niveles de prioridad (alta/media/baja).

### Primero: inicializar SpecForge en el proyecto existente

```
> sf-init

  Código existente detectado: Python 3.12, Typer, Rich, JSON storage.
  Proyecto brownfield detectado.

  Analizando codebase...
  🔴 GATE: project.md + conventions.md
  Usuario: "Aprobado, pero cambiar convención de tests a estilo pytest"
  Actualizado.

  Ahora definamos los principios del proyecto.
  ¿Qué principios guían el desarrollo?
  → "Velocidad sobre features"
  → "Local-first"
  → "Terminal-native"
  ¿Anti-goals?
  → "No es un project manager"
  🔴 GATE: constitution.md
  Usuario: "Aprobado"
```

### Opcional: documentar lo que existe

```
> sf-propose --from-code

  Analizando código existente...
  5 requirements inferidos del codebase [INFERRED]:
  R1: Agregar tarea, R2: Listar, R3: Completar, R4: Borrar, R5: Persistir

  🔴 GATE: requirements.md
  Usuario: "A R2 le falta el mensaje de estado vacío, lo agrego. El resto OK."
  Actualizado.

  🔴 GATE: design.md (arquitectura actual tal como está)
  Usuario: "Aprobado"

  ✓ Baseline registrado como _baseline (status: done)
```

### Después: agregar la nueva feature

```
> sf-propose agregar-prioridades

  ¿Qué hace esta feature? → "Agregar prioridad alta/media/baja con output
  Rich con colores y un filtro --priority en list"

  🔴 GATE: requirements.md
  R1: Setear prioridad al agregar (WHEN el usuario ejecuta `tasks add --priority high "texto"`...)
  R2: Mostrar prioridad en el listado (WHEN se listan tareas THE SYSTEM SHALL colorear...)
  R3: Filtrar por prioridad (WHEN el usuario ejecuta `tasks list --priority high`...)
  R4: Prioridad default (IF no se especifica prioridad THE SYSTEM SHALL asignar "medium")
  Usuario: "Aprobado"

  🔴 GATE: design.md
  Modifica: TaskModel (agrega campo priority), JsonStorage (migración de schema),
  comando list (mapeo de colores), comando add (nuevo flag)
  Usuario: "Aprobado"

  🔴 GATE: tasks.md
  Wave 0: T1 Actualizar TaskModel + schema de storage → R4
  Wave 1: T2 Actualizar comando add → R1, T3 Actualizar comando list → R2, R3
  Trazabilidad: 4/4 cubiertos ✓
  Usuario: "Aprobado"

> sf-build agregar-prioridades
  (waves se ejecutan, gate en cada una)

> sf-check agregar-prioridades
  Veredicto: APPROVE
  ✓ Archivado
```

---

## Detección de resync

Cuando sf-propose se invoca en una feature que ya tiene artefactos:

1. Comparar timestamps: si `requirements.md` es más nuevo que `design.md` → stale
2. Preguntar: "requirements.md fue modificado después de design.md. ¿Regenerar? [s/n]"
3. Si sí → regenerar el artefacto downstream
4. Cascada: si el diseño cambia → ofrecer regenerar las tareas también

Esto cubre el caso común donde el usuario edita un archivo de spec directamente
y los artefactos downstream necesitan actualizarse.

---

## FAQ

**¿Por qué el pipeline de features es solo 4 skills?**
Revelación progresiva. Las capacidades que antes eran skills separados (clarify,
research, map, archive, explore, constitute) ahora viven como references dentro
de los 4 skills del pipeline. Se cargan bajo demanda. Menos overhead de contexto,
menos carga cognitiva. El pipeline es deliberadamente chico — pero no es todo el
framework: `sf-audit` agrega revisión transversal, y los
[skills de soporte](SUPPORT-SKILLS.es.md) cubren pensamiento, triage, TDD, docs
y diseño de infra alrededor.

**¿No es overkill el pipeline completo para un typo o un ajuste de config?**
Para eso hay dos carriles (F34). `sf-propose` arranca clasificando el cambio y
proponiendo un carril **lite** para ediciones triviales y de bajo riesgo — un
`change.md` combinado, un gate, build, un check mínimo — versus el carril
**standard** completo. No elegís el carril para saltarte trabajo; el framework lo
propone y vos lo aprobás en un gate, y queda registrado en `features.json`. Lite
igual escribe un test y un `trace.json`, así que sigue dentro de drift detection
— menos ceremonia, no menos integridad. Si un cambio lite resulta más grande de
lo que parecía, se promueve a standard en pleno vuelo (el escape hatch solo va
hacia arriba).

**¿Qué pasa con un spec después de archivar la feature? ¿No envejece?**
Ese es el modo de falla clásico de SDD, y SpecForge trata el spec archivado como
**documento vivo**, no como snapshot congelado (el snapshot histórico ya lo da el
commit de git). `archive` sella la feature con un vínculo vivo al código —
`trace.json`, la matriz estructurada que mapea cada requirement a su `path:símbolo`
y test. Para cambiar una feature enviada corrés `sf-amend`, que edita ese spec y
esa matriz en su lugar en vez de forkear uno paralelo. Y `sf doctor --drift`
lee el `trace.json` para avisarte cuando el código se movió de abajo de un
requirement — barato, porque solo chequea los anclajes exactos, no el repo entero.

**¿Funciona para un equipo, o solo individual? ¿Cómo se relaciona con el code review del PR?**
Funciona en equipo sin construir un sistema de permisos propio — se apoya en
git/PR (F35/F36). Los gates de creación (propose/build) son del autor en una
branch `feature/<slug>`; el **gate de veredicto se mapea al approve del PR** —
los artefactos viajan en el PR, así que el reviewer aprueba código y spec juntos
(mapear, no duplicar). El ownership es `owners` en la constitución + `CODEOWNERS`
de git. La convención git es una branch por feature, un commit por wave,
`archive` = merge. Y `sfx-github` puede exportar el roadmap a issues **en una
dirección** (el tracker indexa el *qué*, SpecForge tiene el detalle — sin sync
bidireccional frágil).

**¿Puedo usar SpecForge con cualquier agente de IA?**
Sí. Los skills son archivos markdown. Cualquier agente que lea markdown puede
ejecutarlos. El AGENT.md está pensado para Claude Code pero los skills son agnósticos.

**¿Qué pasa si check sigue devolviendo REVISE?**
REVISE te devuelve a sf-build con correcciones específicas. Si el spec en sí estaba
mal, editalo directamente y la detección de resync propagará los cambios.

**¿Puedo tener múltiples features activas a la vez?**
Todavía no — en esta versión el flujo es **serial**: una feature activa a la vez,
el resto queda `queued` en `features.json`. Así los gates y el checkpoint de
sesión quedan sin ambigüedad. Cada feature igual tiene su propia carpeta, así que
el paralelismo es una capacidad futura planificada (`active_feature` rastreado +
secciones de sesión por feature); por ahora, archivá o aparcá la feature actual
antes de arrancar otra.

**¿En qué se diferencia de OpenSpec / Spec Kit / CaveKit?**
SpecForge combina: constitución + identidad de Spec Kit, organización por cambios
de OpenSpec, ejecución por waves + backprop de CaveKit. Gates humanos en cada
artefacto, notación EARS, detección de resync, specs vivas con detección de drift,
y revelación progresiva son exclusivos de SpecForge. Y donde las herramientas SDD
(Spec Kit incluido) son más fuertes una vez que ya *sabés* qué construir,
SpecForge además cubre el paso anterior — `sfp-scout` de-riskea una idea difusa
desde cero — abarcando el arco completo: idea → de-riskeada → spec → build → check
→ mantenida viva.

**¿Me ayuda a descubrir QUÉ construir, o solo a construir un spec ya conocido?**
Ambos. Para una idea clara, arrancás en `sf-init`. Para una difusa, arrancás en
`sfp-scout`: investiga el landscape (vía los MCPs de research), stress-testea la
idea, y devuelve un discovery brief con veredicto proceed/pivot/**kill** — y hace
handoff a `sf-init`. De-riskea; no pretende validar demanda. La visión/identidad
en sí la sigue capturando la conversación de constitución de `sf-init`.
