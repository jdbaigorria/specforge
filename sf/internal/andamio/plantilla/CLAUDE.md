# SpecForge

Este proyecto trabaja con `sf`. **No te sepas el flujo: preguntáselo.**

## El bucle

```
1.  Corré `sf next`.

2.  Hacé lo que diga, con el skill y el modelo que diga:
      via: vos        → trabajás vos, de frente
      via: subagente  → lanzás un subagente fresco
      via: consola    → salís por CLI con el `comando:` que te dio, y le
                        prestás las manos: corrés vos `sf context` y `sf done`

3.  Cuando vuelva el control, corré `sf next` otra vez.
    NO leas lo que devolvió el que trabajó — el estado es la verdad.

4.  Si `sf` dice 🛑 o ⏸, mostráselo al usuario y esperá.
    Su respuesta es: sf approve · sf reject "motivo" · sf take <f-#> ·
                     sf model <nombre> · sf dismiss <h-#> "motivo"
```

**Eso es todo.** No hay una tabla de estados que mantener acá: `sf next` devuelve el estado, el
skill, el modelo, el `via` y el `comando` en cada llamada.

## Las cuatro cosas que no son obvias

- **El que trabaja pide su propio sobre.** Arranca con `sf context` y termina con `sf done`. Vos
  no le pasás archivos ni le explicás en qué paso está — eso lo decide `sf`.
- **Vos no leés el trabajo.** El subagente puede decir *"terminé"* y estar equivocado. `sf done`
  corre las compuertas y decide si el estado se mueve. **Volvé a `sf next` y creele a eso.**
- **Si `sf done` da ✗, no lo arreglás vos.** El que trabajó todavía tiene el contexto: dejalo
  arreglarlo ahí. Recién si `sf` dice **ME TRABÉ** entra el usuario.
- **En `via: consola`, el que trabaja no tiene manos.** Un modelo por CLI no puede correr
  `sf context` ni `sf done`, así que los corrés vos: `sf context --completo` le da el sobre
  embebido para pegarle en el prompt, y `sf done` lo cerrás cuando vuelva. **Mismos comandos,
  mismo sobre; lo único que cambia es quién los tipea.**

## Los códigos de salida

Para no tener que leer el texto:

```
0   hay trabajo          seguí el bucle
1   algo se rompió       mostralo
2   🛑 ⏸ ⚠               es del usuario, esperá
3   no queda nada
```

## Empezar de cero

```bash
sf install   # esto: CLAUDE.md y AGENTS.md · ~/.specforge/ con tus modelos
sf init      # el andamio: 2 directorios, detecta el stack, el estado vacío
sf next      # y de acá en adelante, el bucle
```

## Si `sf` pide un modelo que no conoce

Para y te pregunta — **no elige el reemplazo solo**, porque eso sería opinar sobre qué modelo se
parece a cuál. La respuesta lo declara para siempre, en todos tus proyectos:

```bash
sf model deepseek --via consola --comando "deepseek exec"
sf model o3 --via subagente
```

**La lista no se escribe de antemano: crece con cada aprobación.**

---

> `AGENTS.md` es **este mismo archivo**, no una traducción. Lo único que cambia entre harness es
> cómo se lanza un subagente, y eso el harness ya lo sabe hacer.
