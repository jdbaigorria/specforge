---
lenguaje: go                      # lo llenó sf init (detectStack)
manifiesto: go.mod                # lo llenó sf init
test_cmd: go test ./...           # lo llenó sf init — SIN ESTO sf no sella
test_requiere: []                 # qué SERVICIO necesita el test_cmd: [postgres] [docker] [redis]
mutacion: "gremlins"              # la del ㉒ — sf init deja un ⚠ con la de tu stack
dependencias_aprobadas: []        # arranca vacía y crece con cada aprobación
git:
  branch_por_feature: true
  patron_branch: "feat/{feature-id}-{slug}"
  commit: conventional
  merge: no-ff                    # no-ff | squash | ff
  branch_base: main
---

# <proyecto> — constitución

## Arquitectura

<La forma de la cosa. Capas, límites, qué le habla a qué. Un diagrama en texto
vale más que tres párrafos.>

```
<un esquema simple: los módulos y las flechas entre ellos>
```

<Qué NO puede pasar: "la capa X nunca importa la capa Y". Los límites son lo que
un subagente frío no puede deducir mirando el código.>

## Stack y por qué

| Pieza | Elección | Por qué |
|---|---|---|
| <lenguaje / runtime> | <cuál> | <la razón, no "es popular"> |
| <persistencia> | <cuál> | <la razón> |
| <dependencias> | <cuáles> | <la razón> |

> Cada elección con su razón al lado. El *por qué* es lo que le permite a un
> subagente decidir algo que vos no anticipaste.

## Convenciones de código

- **Nombres:** <la convención>
- **Errores:** <envueltos y propagados / manejados en el borde / …>
- **Comentarios:** <densidad y en qué idioma>
- **Tests:** <table-driven o uno por caso · dónde viven · filesystem real o mocks>
- **Qué significa "terminado":** <y esta es la que más falta le hace al ⑲>

## Estructura de carpetas

```
<el árbol, con una línea por carpeta diciendo qué va adentro>
```

## Reglas de trabajo

- **Código mínimo** — subí la escalera antes de escribir código nuevo: stdlib →
  nativo del lenguaje → dependencia ya aprobada → una línea propia. Código
  necesario, no código ingenioso.
- <otra regla, si cambia lo que alguien hace>

> No repitas acá lo que la máquina ya obliga (commitear al cerrar el lote, crear
> la branch, ver el rojo antes del verde). Una regla que repite una compuerta es
> una regla que se va a desincronizar de ella.
