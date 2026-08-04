<p align="center">
  <img src="assets/specforge-logo.png" alt="SpecForge" width="380" />
</p>

# SpecForge

Framework de Spec-Driven Development. La especificación es el producto — el código es un subproducto regenerable.

Un pipeline de 4 skills por feature + `sf-audit` para revisión transversal del proyecto + skills de soporte. Revelación progresiva. Gate humano en cada artefacto. Cero ceremonia sin propósito.

**¿Recién llegás?** Leé primero [el modelo mental](docs/mental-model.md) — una página sobre cómo piensa SpecForge.

**Instalación:** ver [INSTALL.md](INSTALL.md). **Licencia:** [MIT](LICENSE).

---

## Qué es SpecForge — y qué no es

SpecForge gobierna el *ciclo completo* de spec a código verificado, y se apoya en una
**capa determinista** (un CLI en Go + hooks por harness) para que el workflow no dependa
de que el modelo se porte bien a medida que el contexto se llena.

**Es:**
- Un pipeline de specs donde **la estructura y la secuencia se fuerzan mecánicamente** — un
  hook *deniega* escribir código antes de su gate, el CLI *rechaza* un artefacto malformado.
  Los estados ilegales son inalcanzables, no solo desaconsejados.
- **JSON-first**: cada artefacto es una fuente JSON validada que se renderiza a Markdown, así
  los docs no pueden driftear de los datos.
- **Autogobernado desde disco**: la próxima acción válida se deriva del estado
  (`sf state current`), los gates viven en un ledger, y las lecciones vuelven a la
  constitution (`backprop`). La conversación puede desaparecer y SpecForge sigue sabiendo qué hacer.

**No es:**
- Una garantía de *calidad del código* — solo de *estructura y secuencia*, más una traza para
  auditar la calidad vos mismo. SpecForge es honesto con esa línea.
- Un instalador universal de agentes, una wiki de conocimiento, ni un generador autónomo de
  código que saltea el juicio humano.

### Dónde encaja en el panorama

La frase honesta: **otras herramientas preparan o proponen; SpecForge gobierna.**

| Herramienta | Qué hace | Enforcement | Fuente de verdad |
|-------------|----------|-------------|------------------|
| **SpecForge** | Pipeline spec→build→verify con gates, trace, backprop | **Bloqueante** (hooks deniegan) | **JSON validado** → MD renderizado |
| [Kaddo](https://github.com/Kaddo-kdd/kaddo) | Prepara *conocimiento* vivo como contexto para agentes | Advisory (informa) | Markdown + front-matter |
| [OpenSpec](https://github.com/Fission-AI/OpenSpec) | Capa liviana de specs (proposal/spec/design/tasks) | Ninguno (living docs) | Markdown |
| [GitHub spec-kit](https://github.com/github/spec-kit) | Scaffolding spec-driven para agentes | Ninguno | Markdown |

Encontrar buenos vecinos acá es justamente el punto: Kaddo y SpecForge llegaron de forma
independiente a la misma apuesta de fondo — *determinismo antes que IA, conocimiento cerca del
código*. La contribución distintiva de SpecForge es la **espina de enforcement + trazabilidad**
que las demás dejan libradas a la buena voluntad.

## Tabla de contenidos

- [Qué es SpecForge — y qué no es](#qué-es-specforge--y-qué-no-es)
- [Cómo funciona](#cómo-funciona)
- [La capa determinista](#la-capa-determinista)
- [Documentación](#documentación)
- [FAQ](#faq)

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

## La capa determinista

Las skills son la mitad **cooperativa**: un LLM produce specs y juzga código. Pero el
seguimiento de instrucciones se degrada a medida que el contexto se llena — por eso
SpecForge trae una mitad **determinista** que no depende de la buena voluntad del modelo:
un CLI chico en Go, **`sf`**, y **hooks** por harness.

> **La división del trabajo:** el LLM *produce y juzga*; `sf` *persiste, valida, computa y
> renderiza*; los hooks *fuerzan e inyectan* en los eventos del harness.

| Tier | Qué protege | Cómo |
|------|-------------|------|
| **Estructural** (hermético) | orden de gates, schema, dependencias, flujo serial | un hook *deniega* la tool call — no se puede saltear |
| **Calidad** (cooperativo) | "¿esto es bueno / mínimo / alineado?" | un sub-agente fresco juzga; el veredicto es un *nudge*, registrado para el humano |

Podés volver inalcanzables los estados ilegales (estructural). No podés forzar buen contenido
a existir (calidad) — así que la calidad la sube un checker, no se garantiza. SpecForge es
honesto sobre cuál es cuál.

→ **Referencia completa: [docs/cli-and-hooks.md](docs/cli-and-hooks.md)** (en inglés) — el
surface de `sf`, la vía de escritura JSON-first, el cómputo de waves, los slices de contexto
y cada evento de hook.

## Documentación

El README es la puerta de entrada; la profundidad vive en `docs/` (en inglés).

| Página | Qué contiene |
|--------|--------------|
| [Mental model](docs/mental-model.md) | Cómo piensa SpecForge en una página. **Empezá acá.** |
| [CLI & hooks](docs/cli-and-hooks.md) | La capa determinista completa: comandos `sf`, JSON-first, hooks. |
| [Arquitectura](docs/architecture.md) | Layout de directorios, estructura de skills, modelo de ejecución. |
| [Referencia de skills](docs/skills.md) | Cada skill (pipeline + audit + soporte) y el flujo de artefactos. |
| [Conceptos](docs/concepts.md) | Notación EARS, ciclo de vida, backprop, resync. |
| [Walkthrough](docs/walkthrough.md) | Ejemplos greenfield y brownfield, de punta a punta. |
| [Skills de soporte](SUPPORT-SKILLS.es.md) | Los helpers standalone `sfx-*`. |
| [Ejemplos](examples/) | Features reales y completas para inspeccionar. |
| [Instalación](INSTALL.md) | Setup para Claude Code y otros harnesses. |

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
