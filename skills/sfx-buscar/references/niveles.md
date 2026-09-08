# Los tres niveles de herramienta

Ordenados por lo que cuesta **arrancarlos**, no por lo que valen. El de arriba corre en cualquier
arnés sin configurar nada; el de abajo hay que pagarlo.

## Nivel 0 — sin llave, en cualquier arnés

Lo único que hace falta es poder hacer una llamada HTTP (`curl`, `fetch`, o lo que exponga el
arnés). **Nunca hay excusa para saltearlo.**

| Para qué | Cómo |
|---|---|
| ¿existe este paquete en npm? | `https://registry.npmjs.org/-/v1/search?text=<consulta>&size=10` |
| ¿en PyPI? | `https://pypi.org/pypi/<nombre>/json` · búsqueda: `https://pypi.org/search/?q=` |
| ¿en crates.io? | `https://crates.io/api/v1/crates?q=<consulta>` |
| ¿en Go? | `https://pkg.go.dev/search?q=<consulta>` |
| ¿existe este repo? ¿de qué se quejan? | `https://api.github.com/search/repositories?q=` · `.../issues` — 60 req/hora sin token |
| leer una página concreta | `curl -sL <url>` |
| documentación de una librería | el MCP `context7`, que no necesita llave |

**Lo que el nivel 0 contesta:** *"¿ya existe algo así, quién lo hizo, y qué le falta según sus
propios issues?"* — que es la mayor parte del panorama.

**Lo que NO contesta:** si alguien quiere pagar por esto, si la gente se queja del problema, si la
categoría está creciendo. Para eso hace falta el nivel 1, y si no se llegó, **se dice**.

## Nivel 1 — MCPs con llave

Vienen declarados en el `.mcp.json` del plugin y arrancan solos; los que necesitan llave no
funcionan hasta que esté la variable de entorno.

| Servidor | Para qué | Llave |
|---|---|---|
| `tavily` | búsqueda web y de noticias — el caballo de batalla para señales de demanda | `TAVILY_API_KEY` |
| `github` | búsqueda de repos y de issues a escala | `GITHUB_PERSONAL_ACCESS_TOKEN` |
| `fetch` | leer una URL concreta | ninguna |
| `context7` | docs de librerías | ninguna |

```sh
export TAVILY_API_KEY=...                 # tavily.com
export GITHUB_PERSONAL_ACCESS_TOKEN=...   # scope public_repo alcanza
```

Ojo: `github` **sin** token es nivel 0 (60 req/hora, alcanza para mirar un puñado de repos). Con
token es nivel 1 (5000 req/hora, alcanza para barrer una categoría). Es el mismo servidor en dos
niveles distintos.

## Nivel 2 — pesado, opt-in

No viene de fábrica. Para páginas de competidores detrás de anti-bot o extracción estructurada de
precios y reviews.

| Servidor | Para qué | Nota |
|---|---|---|
| Bright Data | scrapear precios, features y reviews | pago, cuenta propia |
| Playwright | manejar un browser real contra sitios que bloquean | pesado |

## Cómo se declara lo que no se pudo

Sea cual sea el nivel al que se llegó, **el nivel alcanzado se escribe en la evidencia**. Las tres
frases, y son distintas:

```
"no existe"              lo busqué y no está        → retrieved, con la búsqueda
"no lo pude comprobar"   no llegué a la herramienta → no evaluable
"no me suena"            me lo estoy acordando      → model-prior
```

Confundir la segunda con la primera es la falla que se midió el 2026-09-05.
