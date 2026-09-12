# <f-#> — <nombre de la feature>

<Qué se construye, en una oración.>

> **Regla anti-N/A:** poné una sección SÓLO si cambia una decisión o informa la
> implementación. Una sección que diría "N/A" es ruido, y el ruido acá se paga
> en el sobre de cada lote.

## La forma de la solución

<Cómo queda armado. La opción elegida del decision.md, ya sin comparar: acá se
describe lo que se va a construir, no lo que se descartó.>

## Límites y quién le habla a quién

<Qué módulo toca qué. Qué NO puede pasar. Un subagente frío no puede deducir un
límite mirando el código: si "la capa X nunca importa la capa Y", decilo.>

## Los datos

<Qué se guarda, con qué forma, dónde vive, y qué le pasa con el tiempo.>

## Errores y bordes

<Qué pasa cuando falla. Qué pasa con la entrada vacía, la duplicada, la enorme.
Los criterios de tipo SI…ENTONCES del us-# se contestan acá.>

## Lo que esta feature NO hace

<Explícito. Es lo que frena que alguien "aproveche el viaje".>

## Supuestos que tuve que hacer

<Lo que faltaba en el us-# y resolví por mi cuenta. Escribirlo acá es lo que
permite que el ⑰ lo agarre. Si no hay ninguno, borrá la sección.>
