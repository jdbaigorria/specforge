# SpecForge

Este proyecto trabaja con `sf`. **No te sepas el flujo: preguntáselo.**

## El bucle

```
1.  Corré `sf next`.

2.  Hacé lo que diga, con el skill y el modelo que diga:
      via: vos        → trabajás vos, de frente
      via: subagente  → lanzás un subagente fresco
      via: consola    → salís por CLI con el comando que te dio

3.  Cuando vuelva el control, corré `sf next` otra vez.
    NO leas lo que devolvió el que trabajó — el estado es la verdad.

4.  Si `sf` dice 🛑 o ⏸, mostráselo al usuario y esperá.
    Su respuesta es: sf approve · sf reject "motivo" · sf take <f-#> ·
                     sf model <nombre> · sf dismiss <h-#> "motivo"
```

**Eso es todo.** No hay una tabla de estados que mantener acá: `sf next` devuelve el estado, el
skill, el modelo y el `via` en cada llamada.

## Las tres cosas que no son obvias

- **El que trabaja pide su propio sobre.** Arranca con `sf context` y termina con `sf done`. Vos
  no le pasás archivos ni le explicás en qué paso está — eso lo decide `sf`.
- **Vos no leés el trabajo.** El subagente puede decir *"terminé"* y estar equivocado. `sf done`
  corre las compuertas y decide si el estado se mueve. **Volvé a `sf next` y creele a eso.**
- **Si `sf done` da ✗, no lo arreglás vos.** El que trabajó todavía tiene el contexto: dejalo
  arreglarlo ahí. Recién si `sf` dice **ME TRABÉ** entra el usuario.

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
sf init      # el andamio: 2 directorios, detecta el stack, el estado vacío
sf next      # y de acá en adelante, el bucle
```

---

> `AGENTS.md` es **este mismo archivo**, no una traducción. Lo único que cambia entre harness es
> cómo se lanza un subagente, y eso el harness ya lo sabe hacer.
