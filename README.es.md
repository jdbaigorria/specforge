<p align="center">
  <img src="assets/specforge-logo.png" alt="SpecForge" width="380" />
</p>

# SpecForge

Framework de Spec-Driven Development. La especificación es el producto — el código es un subproducto regenerable.

Un pipeline de 4 skills por feature + `sf-audit` para revisión transversal del proyecto + skills de soporte. Revelación progresiva. Gate humano en cada artefacto. Cero ceremonia sin propósito.

**Instalación:** ver [INSTALL.md](INSTALL.md). **Licencia:** [MIT](LICENSE).

---

## Tabla de contenidos

- [Cómo funciona](#cómo-funciona)
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

- **Con gate o iterativos** (`sfx-grill-me`, `sfx-product-owner`, `sfx-tdd`) → inline.
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

### Skills de soporte

Skills standalone que complementan el pipeline pero no son parte de él. Funcionan
sin `specforge/` inicializado y producen artefactos en `specforge/context/`. Ver
[SUPPORT-SKILLS.es.md](SUPPORT-SKILLS.es.md) para la referencia completa:
`sfx-think`, `sfx-triage`, `sfx-grill-me`, `sfx-tdd`, `sfx-documenter`, `sfx-explain`, `sfx-product-owner`,
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
de OpenSpec, ejecución por waves + backprop de CaveKit. La arquitectura de 4 skills,
gates humanos en cada artefacto, notación EARS, detección de resync, y revelación
progresiva son exclusivos de SpecForge.

**¿Dónde está la fase de product owner / visión?**
Integrada en sf-init. La conversación de constitución captura identidad (qué, quién,
por qué), principios y anti-goals. Es la fase de definición de producto — simplemente
no necesita un skill separado.
