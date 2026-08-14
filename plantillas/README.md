# plantillas/

Lo que `sf install` copia dentro de un proyecto. **No es documentación de SpecForge: es su
producto.**

| Archivo | Va a | Qué es |
|---|---|---|
| `CLAUDE.md` | `CLAUDE.md` **y** `AGENTS.md` del proyecto | el orquestador |

## Por qué un archivo y dos destinos

El diseño dice que `AGENTS.md` es **el mismo texto**, no una traducción
([`specs/superficie-sf.md`](../specs/superficie-sf.md) §6). Mantenerlo como dos archivos en este
repo sería tener el mismo contenido en dos lugares — que es exactamente lo que el proyecto evita
en todo lo demás.

**Se escribe una vez y se copia dos veces.** Lo único que cambia entre harness es cómo se lanza
un subagente, y eso el harness ya lo sabe hacer.

## Por qué es tan corto

Porque `sf next` devuelve **el estado, el skill, el modelo y el `via`** en cada llamada. Si el
orquestador tuviera que saber qué skill corresponde a cada estado, esa tabla habría que
mantenerla acá **y** una copia por harness.

> El orquestador no sabe el flujo: lo pregunta.
