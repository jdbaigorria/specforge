# Señales de demanda — la respuesta honesta a "la IA no puede probar demanda"

La IA **no puede probar** que alguien vaya a pagar por algo. Lo que sí puede es buscar **rastros**,
en vez de adivinar. La diferencia importa: un hilo con 200 votos quejándose del problema es
`retrieved`; *"la gente seguro quiere esto"* es `model-prior`.

## Dónde están los rastros

| Fuente | Qué te dice | Cómo se llega |
|---|---|---|
| Reddit / Hacker News / foros | usuarios reales quejándose del problema, o de los que ya existen | `tavily` con la búsqueda acotada al sitio (`site:reddit.com …`), después `fetch` al hilo |
| Sitios de reviews (G2, Trustpilot, stores) | dónde son débiles los que ya existen — el hueco que podrías tomar | `tavily`; si la página bloquea, nivel 2 |
| Noticias y tendencias | si el espacio se está calentando o enfriando | `tavily` en modo noticias |
| Issues de repos parecidos | necesidades sin cubrir en las alternativas abiertas | **nivel 0**: la API pública de GitHub |

La última fila es la única que se alcanza sin llave, y por eso el nivel 0 **mapea el panorama pero
no mide la demanda**.

## El trabajo que el usuario está contratando (JTBD)

No describas el producto: describí **el trabajo para el que lo contratan**.

> Cuando **<situación>**, quiero **<motivación>**, para **<resultado esperado>**.

Esto mantiene la investigación honesta sobre *quién tiene el problema y cuándo*, y frena la solución
que anda buscando un problema.

## El hueco, y las dos mitades que casi siempre se olvidan una

- **En qué te diferenciás:** el hueco que los que ya existen **no pueden** cerrar sin romper su
  propio modelo de negocio. *"El nuestro es más lindo"* no es un hueco.
- **De qué te surtís:** lo que los repos comparables **ya resolvieron** y conviene tomar en vez de
  reescribir. Una investigación que sólo busca huecos te entrega un producto hecho desde cero.

La segunda mitad suele ser la más rentable de las dos, y es la que se saltea.

## Posicionamiento, en una línea

> Para **<quién>** que **<cuándo>**, a diferencia de **<el que ya existe>**, esto hace **<qué>**.

Si esa línea no se puede llenar con evidencia `retrieved`, el diferenciador **todavía no está
probado** — y eso se dice, no se maquilla.
