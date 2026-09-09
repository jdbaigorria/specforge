# El banco — cómo se corre un tramo

**Fecha:** 2026-09-08 · Lo piden `specs/tramo-1.md` y `specs/por-tramos.md`.

---

## La regla de higiene, primero, porque ya se rompió dos veces

> **Adentro de la corrida va SÓLO el proyecto.**
> Todo lo que habla del experimento vive acá, en el repo, fuera del árbol que el
> agente camina.

**Lo que pasó el 2026-09-08.** El banco tenía, uno al lado del otro:

- las corridas anteriores renombradas (`corridas/A-20260905/`), con su `brief.md` entero
- un `IDEA.md` que explicaba que la idea *"seguro ya existe"* — o sea, **la respuesta**
- un `GUION.md` contando qué se probaba y cómo se pasa

**El agente las encontró y concluyó solo que ya era momento de implementar.**

Y antes de eso, el 2026-09-08 a la mañana, la otra: la semilla vieja y **su resultado**
estaban guardados en ICM. `icm recall` devolvía *"ccusage, 18.4k estrellas, cubre 100%
del spec"*. El `CLAUDE.md` global manda usar ICM, así que el agente **tiene la orden**
de ir a buscar la respuesta que debería investigar.

**Las dos veces la corrida salía verde y medía cero.** No probás el scout: probás lo que
el agente encontró tirado.

### Las tres reglas que salen de ahí

```
①  la semilla y su resultado NO se guardan en ninguna memoria
②  las corridas viejas NO se archivan adentro del banco
③  nada que explique el experimento entra al directorio de la corrida
④  cada corrida va en un DIRECTORIO NUEVO
```

**La ④ es la que costó más cara, y es la menos obvia.** ICM inventa un proyecto de
memoria **con el nombre del directorio**: correr en `corridas/A` creó el tópico
`context-sf-banco-A`, y ahí quedó guardado *"la corrida A NO llegó al PRD, no queda
nada"*. O sea que **cada corrida en `corridas/A` envenena la siguiente corrida en
`corridas/A`** — para siempre, aunque cambies la semilla y aunque borres la carpeta.
El veneno no está en el disco: está indexado bajo el nombre.

`nuevo.sh` le pone fecha y hora al nombre justamente por eso. Verificado el
2026-09-08: en una corrida nueva el wake-up dice `project: a-20260908-2341` y no
inyecta nada del banco.

**El resto que queda, y es tuyo decidirlo:** el path dice `sf-banco/corridas/A`. Eso no
da la respuesta, pero le avisa al agente que está en una prueba. Si te parece que
importa, se renombra el árbol a algo neutro.

---

## Armar una corrida

```bash
cd ~/projects/workspace/personal/specforge
./banco/nuevo.sh A --rehacer
./banco/nuevo.sh B --rehacer
```

Deja el proyecto limpio y nada más. El `sf doctor` de arranque se guarda en
`banco/diagnosticos/`, **afuera**, porque nombra el banco.

**Mirá el diagnóstico antes de arrancar.** Dos cosas:

- `skills 9/9` **sin ningún "le falta"** — un compositor sin su primitivo no falla:
  el modelo improvisa y la corrida sale igual, sin método.
- el bloque `buscar` — con qué tier vas a correr.

**Medido el 2026-09-08: no hay ninguna llave puesta, así que la corrida es de nivel 0
puro.** Y está bien: es exactamente la hipótesis. El fallo del 05-09 fue que el nivel 0
—registries, la API pública de GitHub, `curl`— no estaba nombrado en ningún lado. Ahora
sí lo está en `sfx-buscar`.

### Y el chequeo de la semilla

```bash
icm recall "<las palabras clave de la semilla>"
```

Tiene que volver **ruido de otros proyectos**. Referencia medida el 2026-09-08:

```
semilla quemada    0.628 · 0.603   traía la respuesta escrita
semilla limpia     0.57–0.60       ruido de picode y fastro
```

---

## T1 — el brief (①–⑥)

```bash
cd ~/projects/workspace/personal/sf-banco/corridas/A && claude     # pegá banco/IDEA.md tal cual
cd ~/projects/workspace/personal/sf-banco/corridas/B && opencode   # el mismo texto
```

**Por qué opencode del otro lado.** Es el que **obedece literal**. El bug del ⑧ vivió
meses porque los modelos fuertes **tapaban** la contradicción rellenando el hueco, y el
que obedece literal la reveló. Uno tapa el defecto, el otro lo muestra.

### La vara — y cambió

```
ANTES   "pasa si el estado.json quedó igual con los dos modelos"
AHORA   la vara es de EVIDENCIA, no de veredicto
```

El único campo que se mueve en T1 es `brief_sellado`, y ése **es** el veredicto: juicio
puro. Con la vara vieja **T1 fallaba siempre**.

> **T1 PASA si, con los dos modelos:** `entrevista.md` cierra con `abiertas: 0` **y**
> `evidencia.md` cita al menos un link.
>
> **Que uno diga `hacelo` y el otro `no-lo-hagas` NO es un fallo.** Es el ⑥ haciendo su
> trabajo.

### Lo que vas a ver en el ⑥

Un **acta**, no un `✓ listo`:

```
COMPROBÉ / MEDÍ / NO PUEDO COMPROBAR
```

Leelo antes de aprobar. Ése es el cambio: la máquina comprueba y te entrega la
evidencia; vos no validás, **leés y decidís**.

### El comparador

```bash
./banco/comparar.sh A B
```

### Anotá

- [ ] ¿el ①–⑤ corrió **por rondas**, o volvió a ser una charla suelta? Si preguntó de a
      una, `sfx-grilling` no se cargó.
- [ ] ¿investigó con nivel 0, o dijo que no podía? Si escribió todo `model-prior`
      teniendo `curl`, **era el modelo**.
- [ ] ¿apareció `vocabulario.md`? No es obligatorio. Si apareció, ¿hacía falta?
- [ ] ¿propuso un prototipo? ¿lo aprobaste vos o lo construyó solo?
- [ ] **¿la ventana se llenó antes del ⑥?** Es la observación §10① de `tramo-1.md`: no
      hay regla de corte y el replanteo alargó el tramo. **Si pasa, anotalo.**

---

## T2 — el PRD y la constitución (⑦⑧)

Es el que importa: primer paso **delegado**, y donde vivía el bug.

```bash
cd ~/projects/workspace/personal/specforge
cp -r ~/projects/workspace/personal/sf-banco/corridas/A ~/projects/workspace/personal/sf-banco/corridas/T2-A
cd ~/projects/workspace/personal/sf-banco/corridas/T2-A
sf next && sf lanzar
```

Y la variante headless, que es la razón de ser del multi-arnés — modelo barato en lo
mecánico, los grosos para planificar:

```bash
sf lanzar --seco --harness=opencode --alias=ultra   # PRIMERO mirá el comando
sf lanzar --harness=opencode --alias=ultra
```

### Anotá

- [ ] **¿el hijo delegado se quedó en su paso, o siguió?** No se puede probar con un
      test: sólo se ve mirando una corrida.
- [ ] **¿pasa un PRD flojo?** `compuerta.PRD` es un `os.Stat` (`compuerta.go:444`). Si
      pasa un PRD basura, **ése es el hallazgo más importante de la corrida.**
- [ ] **¿el sobre del ⑦ trajo `vocabulario.md`?** (si existe)
- [ ] **¿la carta del arnés mintió?** Ya pasó: `exit=0` y `success` sin haber hecho nada.
      Mirá `.specforge/lanzamientos/` contra lo que hay en `.docs/`.
- [ ] ¿los dos `estado.json` quedaron iguales? **Acá el diff SÍ es la vara**: se mueven
      `prd_hash` y `constitucion_sellada`, que son mecánicos.

---

## Al terminar

**Guardá la corrida FUERA del banco.** No renombrada al lado — afuera, o se convierte en
la migaja de la próxima.

```bash
mv ~/projects/workspace/personal/sf-banco/corridas/A ~/archivo-corridas/T1-A-$(date +%Y%m%d)
```
