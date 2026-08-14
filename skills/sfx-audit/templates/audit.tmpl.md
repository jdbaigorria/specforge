# Auditoría — <qué se auditó: f-1 · f-2 · f-3, o "todo">

**Fecha:** <YYYY-MM-DD> · **Alcance:** <N features · M criterios>

> Si la suite NO pasa, decilo acá arriba y en negrita: todo lo que sigue se
> juzgó sobre un árbol roto y hay que leerlo como provisorio.

## Lo que sf comprobó

<Pegá tal cual los hechos que imprimió `sf audit`. Son hechos, no opiniones: no
hay que re-verificarlos, hay que explicar qué significan.>

```
✓ alcance: 3 features · 14 criterios
✗ us-3/CA-2 se dio por cumplido en f-1 y su test ya no existe
✓ la suite pasa
```

### Qué significa cada ✗

| Hecho | Qué implica | Gravedad |
|---|---|---|
| `us-3/CA-2` sin su test | <la afirmación de f-1 perdió su respaldo: hoy nadie prueba que…> | <alta/media/baja> |

## Hallazgos

<Uno por bloque. Cada uno con su evidencia: archivo y línea, o no va.>

### H1 — <el hallazgo en una línea>

**Dónde:** `ruta/al/archivo.go:42`
**Historia:** us-3/CA-2 · **Feature:** f-1
**Tipo:** <la historia no se satisface | dos features no se integran | una feature
rompió a otra | la afirmación perdió su respaldo>

<Qué encontraste, concreto. Qué dice el código, qué pedía la historia, dónde está
la diferencia.>

**Cómo se ve desde afuera:** <qué le pasa a alguien que usa esto>

## Entre features — las costuras

<La sección que ninguna revisión por feature puede escribir. Qué pasa cuando f-1
y f-3 se tocan. Qué supuso una que la otra no cumple.

Si no encontraste nada acá, decilo: es información.>

## Lo que verifiqué y está bien

<Y esta sección NO es relleno. Sin ella, nadie sabe cuánto del sistema se miró de
verdad — un informe que sólo lista problemas puede ser exhaustivo o puede haber
mirado tres archivos, y se leen igual.>

- **us-1** — seguida punta a punta desde `cmd/x.go:12`. Hace lo que la historia
  pedía.
- **us-4** — <…>

## Lo que NO miré

<Explícito. Qué quedó afuera del alcance y por qué.>

## Qué hacer

<Nada de esto lo hace la auditoría. Lo que valga la pena entra por el embudo:>

```bash
sf new "<el hallazgo, como entrada al backlog>"
```

<Y si algo se repite en tres features, no es un hallazgo: es una regla que le
falta a la constitución. Proponela, no la escribas.>
