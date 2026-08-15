# plantilla/

**`CLAUDE.md` de acá es el orquestador que `sf install` pone en un proyecto.** No es
documentación de SpecForge: es su producto.

## Por qué vive adentro del paquete y no en la raíz del repo

Porque `go:embed` **no puede salir del directorio de su paquete**. El archivo se compila adentro
del binario, y la alternativa —buscarlo en disco relativo al ejecutable— se cae sola: un `sf`
instalado con `go install` vive en `~/go/bin` y no tiene el repo al lado.

Estaba en `plantillas/` en la raíz y se movió acá cuando se construyó `sf install`. **Sigue
habiendo una sola copia**, que era lo importante.

## Un archivo, dos destinos

`sf install` lo escribe como **`CLAUDE.md` y como `AGENTS.md`** del proyecto.

El diseño dice que `AGENTS.md` es **el mismo texto**, no una traducción
([`specs/superficie-sf.md`](../../../../specs/superficie-sf.md) §6). Mantenerlo como dos archivos
acá sería el mismo contenido en dos lugares — que es exactamente lo que el proyecto evita en todo
lo demás. **Se escribe una vez y se copia dos veces.**

Lo único que cambia entre harness es cómo se lanza un subagente, y eso el harness ya lo sabe
hacer.

## Por qué es tan corto

Porque `sf next` devuelve **el estado, el skill, el modelo, el `via` y el comando** en cada
llamada. Si el orquestador tuviera que saber qué skill corresponde a cada estado, esa tabla habría
que mantenerla acá **y** una copia por harness.

> El orquestador no sabe el flujo: lo pregunta.

## Si lo editás

El build lo embebe tal cual. Y `sf uninstall` **sólo borra el archivo si es idéntico a éste** —
si alguien lo editó en su proyecto, no lo toca.
