# La máquina de estados

**Fecha:** 2026-08-11 / 12 · **Método:** debatida sobre `flujo-real.md` y `artefactos.md`, ya
con los artefactos conocidos.

> **Por qué recién ahora.** No se podía decidir dónde están los cortes sin saber qué produce
> cada paso y quién lo consume. Esa prueba de escritorio está cerrada en
> [`artefactos.md`](artefactos.md); esto se apoya en ella y **no la repite** — la referencia.
> Es la regla 1.4 aplicada a este documento.

**Entra:** los 23 pasos de [`flujo-real.md`](flujo-real.md) y el inventario de
[`artefactos.md`](artefactos.md) §11.
**Sale:** 9 estados, las compuertas de cada uno, cuatro tipos de parada y la forma del
`estado.json`.

---

## 1. La regla de corte

Los 23 pasos **no son 23 estados**. La pregunta *"¿dónde corto?"* tiene una sola respuesta:

> **Cortá donde cambia lo que hay que tener en la cabeza.**

Un estado es **un trabajo que arranca con un contexto propio y termina en algo comprobable**.

Y la regla corta para los dos lados, que es lo que la hace útil:

- **Corta** cuando lo que sigue no necesita lo anterior → contexto fresco.
- **No corta** cuando lo anterior es justo lo que hace falta → mismo estado.

El ejemplo de las dos direcciones está en §2: partió el ciclo de feature en cuatro, y pegó el
⑯ al ⑮.

### La corrección que la fijó

El primer strawman partía el bloque de planificación (⑫–⑯) en tres estados, con el argumento
de que ningún subagente se llene de contexto. **Javier lo objetó y el argumento no se
sostiene:**

1. **El dolor #5 es un daño de código, no de planificación.** Un implementador saturado
   escribe un método vacío y sigue. Un planificador saturado no produce nada equivalente: no
   existe una *"spec mock"*.
2. **En planificar, el contexto grande es un activo.** El ⑬ escribe la spec de la solución
   elegida y le sirve acordarse de las dos que descartó. Cortando, el segundo actor tiene la
   conclusión (`decision.md`) pero perdió el razonamiento.
3. **Y el flujo ya lo decía:** el ⑰ revisa el bloque entero, y *"pido cambios"* vuelve al
   bloque entero. **Nadie entra ni sale por el medio.** Cortar donde nadie entra ni sale es
   cortar por gusto.

### La separación que destrabó todo

El strawman pegaba dos cosas que son distintas:

> **un estado ≠ un subagente**

Un estado puede tener **checkpoints internos**, y `sf` no necesita un campo para ellos:
**mira qué archivos existen.**

```
¿hay decision.md?      no  → arrancá de cero
¿hay spec-design.md?   no  → retomá desde la spec
¿hay tareas.json?      no  → retomá desde las tareas
```

Es la regla 1.5 de `artefactos.md` aplicada al progreso: **lo que se puede deducir, no se
guarda.** Un subagente que se muere a la mitad se retoma sin que el `estado.json` sepa nada.

---

## 2. Los nueve estados

**Cinco de producto** — corren una vez, y sólo en la entrada A. **Cuatro de feature** — se
repiten, una vuelta por feature.

```
PRODUCTO  (una vez · sólo entrada A)
  brief          ①–⑤   bucle con Javier   → brief.md            🛑 PARA   (el sello ⑥)
  prd            ⑦                        → prd.md
  constitucion   ⑧                        → constitucion.md     🛑 PARA
  backlog        ⑨                        → us-#.md             ⏸ enter
  roadmap        ⑩                        → roadmap.json

FEATURE  (una vuelta por feature)
  planificacion  ⑫–⑯   un subagente Opus  → decision.md
                        contexto amplio      spec-design.md
                                             tareas.json        🛑 PARA   (el ⑰)
  implementar    ⑱–⑳   subagente POR LOTE → código + commit
  revision       ㉑㉒    subagente grande   → revision.json
  cierre         ㉓     subagente          → doc + PR            ⏸ enter
```

**El ⑪ no es un estado:** es tomar la próxima del roadmap, y eso es la transición que entra a
`planificacion`.

**El ⑯ tampoco:** es la cola de `planificacion`. El que acaba de escribir las tareas conoce
la complejidad mejor que nadie — un contexto fresco acá sería *peor*. Escribe el modelo en
`tareas.json` y listo. Un dato de una línea no merece un estado (regla 1.1).

**El ⑰ no es un estado, es la parada de `planificacion`** — con tres salidas, ver §5.

---

## 3. Las dos clases de estado

Sale de una limitación real, y la máquina tiene que respetarla:

| Clase | Quién lo corre | Por qué |
|---|---|---|
| **conversa con Javier** | **el orquestador**, de frente | un subagente arranca, trabaja y muere — **no te habla** |
| **produce un artefacto** | **un subagente fresco**, con el modelo que corresponda | así el orquestador nunca se llena |

**Conversan:** `brief` (①–⑤ es un pinponeo, no una fila) y el pinponeo de la entrada B.
**Todo el resto va a subagente.**

**Consecuencia:** el orquestador **nunca trabaja, sólo lanza**. No ve código ni specs: sólo
*"terminó"* y *"ahora X"*. Aguanta las 23 vueltas sin llenarse. `session.md` §6 lo decía para
la implementación; acá se extiende a todo el ciclo.

*(Diseño de Javier: "la planificación la debería hacer un modelo como Opus grande con amplio
contexto; la implementación un subagente; la revisión también".)*

---

## 4. Las compuertas — la mitad que importa

Sin esto la máquina es una lista de tareas. La regla dura, ya firmada en `session.md` §6:

> **`sf` no le cree al que trabajó.** Corre, cuenta o compara él mismo.

### Salir de `planificacion`

Todo es contar. Nada es juzgar.

- ¿están los tres archivos?
- ¿`decision.md` tiene **tres** opciones? *(tres, no "varias" — es el número del ⑫)*
- ¿cada lote de `tareas.json` tiene al menos un test?
- ¿cada tarea apunta a un criterio de aceptación de la `us-#`? → empieza a morir el **#7**
- ¿hay librerías nuevas fuera de la constitución? → **avisa**, no frena *(#6)*

### `implementar` — tres compuertas, no una

Es el estado donde muere casi todo:

```
antes    →  ¿existe la branch que dice la constitución?          #4
medio    →  corre los tests del lote: TIENEN QUE FALLAR          #8 #5
después  →  corre los tests: pasan · ¿hay commit del lote?       #2 #3
```

**La del medio es la joya.** `sf` no deja pasar al verde sin haber visto el rojo con sus
propios ojos, y le alcanza un exit code:

> **Un test que pasa antes de que exista el código es un test de mentira.**

### Salir de `revision`

¿Hay hallazgos abiertos en `revision.json`? Con uno solo no avanza: vuelve a `implementar`.
**`sf` cuenta, no opina.**

Y después del arreglo se rehace la revisión **entera** (㉑ y ㉒), no sólo lo que falló — es
lo que `flujo-real.md` ya pedía.

### Salir de `cierre`

¿La carpeta se movió a archivado? ¿El `roadmap.json` la marca hecha? ¿Existe la doc?

---

## 5. Los cuatro tipos de parada

Eran tres en `artefactos.md` §2. **Apareció un cuarto, y estaba escondido en el ⑳.**

| | Dónde | Qué hace |
|---|---|---|
| **PARA** | ⑥ · ⑧ · ⑰ | no avanza sin Javier |
| **PARADA BARATA** | ⑨ · ㉓ | te muestra qué salió, seguís con un enter |
| **SIGUE** | el resto | avanza solo |
| **ME TRABÉ** | ⚠ **nueva** | N intentos fallidos → te llama |

### De dónde salió la cuarta

Cuando `sf` da rojo no alcanza con no mover el estado: le dice al orquestador **qué falta**, y
este lanza un subagente nuevo con ese motivo. Eso es un bucle, y **un bucle sin freno se
cuelga**.

El freno ya estaba en el flujo: el ⑳ dice *"si falla varias veces, entro yo"*. El contador ya
estaba en el borrador del estado (`intentos_fallidos`). Sólo faltaba **darle nombre y decir
qué pasa cuando llega al tope**.

**Y no es configurable.** Las otras tres salen del gusto de Javier (`para_en: [6, 8, 9, 17]`).
Esta es de seguridad.

---

## 5.1 El horizonte de una parada — y hasta dónde llega

**Agregado el 2026-08-31** (commit `d1c7856`), después de correr el bucle en tres harness.

Una parada contestaba *dónde estás* y nunca *qué desencadena decir que sí*. Y como **no todos los
pasos frenan**, un `sf approve` no dispara UN paso: dispara todos los que siguen hasta la próxima
parada. Hay dos tramos así:

```
⑥ sello  →  ⑦ PRD       (sin parada)  →  ⑧ constitución
⑨ ⏸     →  ⑩ roadmap   (sin parada)  →  ⑫–⑯ planificación
```

Javier selló el brief y le salieron el PRD y la constitución de una. En sus palabras: *"es medio
enquilombado saber cuándo no debés aprobar porque sino salta directo a la siguiente fase"*. Ahora
la parada lo anuncia:

```
🛑 PARÁ. El brief está escrito y lo sellás vos (el ⑥).

   si aprobás corre:  ⑦ el PRD → ⑧ la constitución
   próxima parada:    el sello del ⑧
```

**Se deriva, no se escribe.** Sale de dos mapas —`ordenDeEstados`, que ya vivía escondido adentro
de `SkillsDeEstado` y ahora es una sola copia, y `paranAlFinal`— y no de un texto a mano en cada
parada, que sería una segunda copia del flujo y se desincronizaría el día que alguien mueva un
estado. `TestElHorizonteEsCierto` camina la máquina de verdad y compara el anuncio contra el
recorrido.

### Lo que quedó afuera, a propósito: el ⑰ no anuncia horizonte

**Y no es una tarea pendiente que se olvidó: es un límite del método.**

La derivación es LINEAL — recorre `ordenDeEstados` hasta el primer estado que frena. Del ⑰ para
adelante el flujo **deja de ser una fila**:

- los lotes del ⑱–⑳ dan vueltas: cada uno es rojo y después verde, y son N
- un finding del ㉑ **manda la feature de vuelta a implementar**, y la revisión se rehace entera
- el ⑳ tiene su propia parada de seguridad, la **ME TRABÉ** de §5, que no está en ningún orden

Una derivación lineal sobre eso no se equivocaría a veces: **contestaría siempre, y con seguridad,
algo falso.** Y una parada que anuncia mal es peor que una que no anuncia nada — la primera te hace
aprobar confiado.

Así que el ⑰ se dejó sin horizonte y está dicho en el código, en el sitio mismo donde se
construye esa parada, para que nadie lo "complete" creyendo que fue un descuido.

> **Si alguna vez hace falta**, no es agregarle un caso a `horizonte()`: es **modelar los ciclos**
> —cuántos lotes, qué pasa con un finding, dónde entra el contador de intentos— y eso es un trabajo
> propio, no un remate de éste.

---

## 6. El orden lo elige Javier, y no hace falta configurarlo

**No hay features en paralelo.** Hay una **cola**: features planificadas esperando, y se toca
**una por vez**. Por eso `feature_actual` alcanza.

Lo que Javier elige es el **orden**:

> *"yo puedo decidir si implemento luego de definir una feature, o si quiero crear el spec de
> todas"*

El primer strawman iba a agregar una config (`modo: una_por_vez | planificar_todo`). **Está de
más**, porque la elección ya tiene un lugar en el flujo — las tres puertas del ⑰:

```
⑰ →  pido cambios     vuelve al bloque ⑫–⑯
   →  otra feature     vuelve al ⑪, otra del roadmap
   →  implementar      sigue
```

Elegís *"otra feature"* tres veces → planificaste todo. Elegís *"implementar"* siempre → vas
de a una.

> **El modo no se declara: sale de lo que fuiste eligiendo.** Una config menos, y es el propio
> flujo el que la hace innecesaria.

---

## 7. El plan que envejece

Es el riesgo que abre la puerta *"otra feature"*, y sale gratis cubrirlo.

**El caso.** Planificás `f-1` y `f-2` sin implementar ninguna. La spec de `f-2` dice *"usa
`leerEstado()` de `estado.go`"*, porque miró el repo. Después implementás `f-1`, y en el medio
el implementador renombra `estado.go` y parte esa función.

**La spec de `f-2` quedó mintiendo.** Y lo peor no es el error: el implementador de `f-2` la
lee, no encuentra la función, y **improvisa** — que es justo donde aparecen los mocks.

**El mecanismo, y es de pobre.** `sf` no lee la spec ni la entiende. Anota un número y lo
compara:

```
al terminar de planificar f-2   →  sf anota:  base_commit a3f9c1
al ir a implementar f-2         →  sf mira:   HEAD        7d1b40
                                   distinto   →  avisa
```

Dos strings y un `if`. No sabe *qué* cambió ni si importa: sabe que **el suelo se movió desde
que se dibujó el plano**. Te dice *"esto lo planificaste hace 4 commits, ¿lo mirás antes?"* —
**avisa, no frena**, misma regla que el #6.

Es el mismo truco que el `prd_hash`: **un dato guardado hoy, comparado mañana.** La forma más
barata que tiene `sf` de saber que algo envejeció, sin pensar.

### Autocontenida ≠ vive sola

Duda de Javier: *"pero cada feature es autocontenida"*. Las dos cosas conviven — la
autocontención es del **alcance**, no del **suelo**.

Los departamentos de un edificio son autocontenidos, y comparten los caños. `sf next` y
`sf done` son dos features distintas y **las dos leen `estado.json`**. Es más: la primera
feature de un producto **construye los cimientos** de todas las siguientes.

> **Autocontenida = se entrega sola. No = vive sola.**

Tres cosas salen de ahí:

1. **El aviso no contradice la autocontención: la protege.** Justamente porque la feature se
   planificó sola, nadie le avisa cuando el repo se movió.
2. **Si el aviso salta seguido entre dos features, eran una sola.** Detector gratis de la
   regla del ⑩ (`artefactos.md` §3): si comparten tanto código, comparten solución técnica y
   el agrupamiento estaba mal.
3. **El riesgo lo creás vos, y sólo a veces.** Planificando e implementando de a una, el aviso
   **nunca salta**. El mecanismo duerme el resto del tiempo.

---

## 8. El camino corto es un estado, no una excepción

`flujo-real.md` lo advierte: *"una herramienta que obligue a la ceremonia completa para
arreglar algo chico va a ser evitada igual que ahora, y ahí sí se pierde el rastro"*.

**No es un carril paralelo. Es la misma máquina con estados salteados.**

```
entra al backlog como bug  →  implementar  →  cierre
                              (se saltea planificacion y revision)
```

Mantiene las dos condiciones que Javier ya sostiene: **correr los tests** y **verificar que
quedó resuelto** — que son, exactamente, dos de las tres compuertas de `implementar`.

**Por qué así:** un carril paralelo sería una segunda máquina que mantener. Saltear estados es
gratis. Y el rastro no se pierde, porque el bug igual entró por el backlog.

---

## 9. El `estado.json`

La regla que lo define salió sola de todo lo anterior:

> **Guardá lo que decidió Javier, y lo que `sf` vio pasar. Todo lo demás sale de los
> archivos.**

Los archivos dicen **qué se produjo**. El estado dice **qué se aprobó** — y eso no está
escrito en ningún lado, porque salió de una cabeza.

```json
{
  "producto": {
    "brief_sellado": "hacelo",
    "prd_hash": "a3f9c1",
    "constitucion_sellada": true
  },
  "feature_actual": "f-2",
  "features": {
    "f-1": { "estado": "cerrada", "base_commit": "7d1b40" },
    "f-2": {
      "estado": "implementando",
      "base_commit": "9c2e1a",
      "modelo": "deepseek",
      "intentos_fallidos": 0,
      "lotes": [
        { "lote": 1, "rojo": true, "hash_tests": "b70c33", "commit": "4a8f21" },
        { "lote": 2, "rojo": true, "hash_tests": "4e91b7", "commit": null }
      ]
    }
  }
}
```

### Qué se le sacó al borrador de `artefactos.md` §11

| Campo | Por qué se fue |
|---|---|
| `"paso": 19` | no hay un paso global — cada feature tiene el suyo |
| `modelo_recomendado` | vive en `tareas.json`, lo escribe el ⑯. No se guarda dos veces |
| `verde` | **redundante con `commit`**: si `sf` no deja commitear en rojo, *hay commit* ya significa *estaba verde*. Un campo deducible de otro es un campo que se desincroniza |

### Qué se le agregó

- **`base_commit`** por feature → el envejecimiento del §7.
- **`producto`** → los sellos del ⑥ y del ⑧. Son decisiones de Javier: ningún archivo las
  contiene.
- **`estado`** por feature → la cola del §6.
- **`hash_tests`** por lote → el agujero astuto: `sf` ve rojo, el subagente trabaja, `sf` ve
  verde… **y lo que cambió entre medio fue el test.** Se toma en el rojo y se compara en el
  verde. Comparar dos strings.

### Los dos campos que no se pueden deducir de nada

- **`rojo`** — el momento en que los tests fallaban ya pasó y **no dejó huella**. Si `sf` no lo
  anota cuando lo vio, se pierde para siempre. Es, literalmente, la memoria de que el test era
  de verdad. *(Lo mismo vale para `hash_tests`: es esa memoria, con detalle.)*
- **`modelo`** — es el que se está usando **ahora**, no el recomendado. Cuando Javier sube el
  modelo en el ⑳, esa decisión no queda registrada en ningún archivo.

---

## 10. El saldo

| | |
|---|---|
| **23 pasos** | → **9 estados** |
| **3 decisiones** | ⑥ · ⑧ · ⑰ — el resto avanza solo |
| **4 tipos de parada** | y una sola no es configurable |
| **1 config que no hizo falta** | el modo de orden sale del ⑰ |
| **0 máquinas en paralelo** | una cola, `feature_actual` alcanza |

Y **`sf` sigue sin pensar en ningún lado**: todas las compuertas son correr un comando, contar
un campo o comparar dos strings.

---

## 11. Lo que sigue

1. **Qué de lo construido sobrevive.** Recién ahora, y en este orden — es el punto 5 de
   `session.md` §7. Los 16 repos y lo ya hecho se miran **con la necesidad ya escrita**.
2. **La superficie de `sf`** — qué comandos expone. Ya aparecieron `sf next`, `sf done`,
   `sf context <estado>`; falta el inventario completo y quién los llama.
3. **Qué le toca al orquestador** (`CLAUDE.md` / `AGENTS.md`) y qué a cada skill, con los
   estados ya fijos.
