# Prototipo de lógica — "¿este modelo aguanta los casos raros?"

La pregunta que contesta esta rama es sobre **reglas de negocio, estados o forma de los datos**: eso
que en el papel parece razonable y recién se siente mal cuando lo empujás con casos de verdad.

Si la pregunta es *"¿cómo tendría que verse esto?"*, estás en la rama equivocada:
[pantallas.md](pantallas.md).

## Cuándo es ésta

- *"No sé si esto aguanta el caso de que primero pase X y después Y."*
- *"¿Este modelo de datos me deja representar la situación de…?"*
- *"Quiero sentir cómo sería la API antes de escribirla."*
- Cualquier cosa donde alguien quiera **apretar botones y ver cambiar el estado**.

## Elegí la forma: la pantalla o el guión

Hay dos, y la elige **quién va a mirarla**:

| | La pantalla compartible | El guión que se corre |
|---|---|---|
| **quién mira** | alguien que no programa — vos, un experto del dominio | el que va a construir |
| **qué es** | un HTML solo, todo adentro, se abre con doble clic | un script en el lenguaje del proyecto |
| **cuándo** | hay que *sentir* el modelo, o todavía no hay stack | el módulo tiene que poder levantarse después |

**En T1 la pantalla gana casi siempre**, por dos motivos: todavía no hay proyecto donde levantar
nada, y la pregunta de esta etapa suele ser tuya, no del que implementa.

**El guión gana cuando `lenguaje` ya está en la cabecera de `.docs/constitucion.md`** y lo que se
está probando es un módulo que después se muda al código de verdad. Un reducer en JavaScript no se
muda a un proyecto en Go: ahí el prototipo se escribe en Go y se corre con un comando.

## Los cinco pasos

### 1. Escribí la pregunta, arriba de todo y a la vista

Antes de una línea de código: **qué se está probando y qué se quiere saber**, en un párrafo, visible
en el prototipo mismo — no en un comentario.

Un prototipo que contesta la pregunta equivocada es basura entera, y la única defensa es que la
pregunta esté escrita donde se pueda comparar después, incluso si el que lo abre no estuvo en la
charla.

### 2. Separá la lógica en un módulo puro

La parte que de verdad contesta la pregunta va aparte, escrita como un módulo chico que se pueda
levantar tal cual. **La cáscara es descartable; esto no.**

Qué forma darle depende de la pregunta:

- **una función de transición** `(estado, acción) → estado` — cuando las acciones son eventos
  sueltos y el estado es un valor
- **una máquina de estados** con sus transiciones escritas — cuando parte de la pregunta es *"¿qué
  acciones son legales acá?"*
- **un puñado de funciones puras** sobre un dato plano — cuando no hay estado actual, sólo
  transformaciones
- **un módulo con métodos** — cuando la lógica de verdad es dueña de un estado que dura

Elegí la forma que **contesta la pregunta**, no la más fácil de enchufar a la pantalla.

**Y mantenelo puro:** el módulo no toca la pantalla, no lee el DOM, no sabe que existen botones. La
cáscara lo llama a él; nunca al revés. Eso es lo que hace que sirva después de que el prototipo
muera: la pregunta se contesta, y el módulo validado se muda solo.

### 3. Armá la cáscara

Si es **pantalla**: un archivo, HTML/CSS/JS a mano, todo adentro. Sin framework, sin bundler, sin
servidor — se abre con doble clic y sobrevive a que lo manden por mail.

Si es **guión**: un comando y listo. `go run ./proto`, `python proto.py`, lo que use el proyecto.
Nada de instalar cosas para verlo correr.

En los dos casos, **de arriba hacia abajo**:

1. **El título y la pregunta** del paso 1.
2. **El estado completo**, legible — campos con nombre, no un JSON crudo— y **redibujado después de
   cada acción**, para que el cambio se vea.
3. **Botones libres**: uno por acción, siempre disponibles, para poder toquetear en cualquier orden.
4. **Recorridos guiados**: escenarios, uno por pestaña (o por bloque, si es guión). Cada uno con su
   explicación en castellano y abajo los pasos en orden, cada paso un botón. Arrancar un recorrido
   **resetea a un estado conocido**, para que dé lo mismo todas las veces.

**Los escenarios se eligen por incómodos**, no por prolijos: el camino feliz, un borde difícil de
razonar en el papel, y un intento de hacer algo que **debería ser ilegal**.

### 4. Hablá el idioma del negocio, no el del código

Si la va a manejar alguien que no programa, cada etiqueta se lee como el negocio: *"cancelar el
pedido"*, no *"dispatch CANCEL"*. Un prototipo que hay que traducir no se lo podés dar a nadie.

Lindo pero sobrio: tipografía limpia, aire, un color de acento. **Nada de animaciones**: lo único que
tiene que llamar la atención es el estado cambiando.

### 5. Guardá la respuesta, tirá el código

Lo que sobrevive es **la respuesta**, y entra a `.docs/entrevista.md` como una decisión con
procedencia `probado`, apuntando a la carpeta:

```markdown
- El modelo aguanta la cancelación parcial sin un estado nuevo. [probado]
  - .docs/prototipos/cancelacion-parcial/
```

Si el módulo puro dejó la respuesta escrita mejor de lo que la escribiría la prosa —una máquina de
estados, un esquema, un tipo— **ese pedazo** va al artefacto, recortado a la decisión. No la demo
funcionando.

## Lo que arruina un prototipo de lógica

Cada uno de éstos tiene su versión en positivo al lado, que es la que hay que hacer:

| Se rompe cuando… | Lo que sí |
|---|---|
| le ponés tests | si necesita tests, ya dejó de ser un prototipo |
| lo enchufás a la base de datos real | estado en memoria, salvo que la pregunta **sea** la persistencia |
| generalizás *"por si después…"* | contesta **una** pregunta y se muere |
| mezclás la lógica con la pantalla | el módulo puro se queda puro, o no se puede levantar |
| traés un framework o un servidor | un archivo que se abre con doble clic, o un comando |
| lo mandás a producción | la cáscara se escribió sin manejo de errores; lo que vale es el módulo |
