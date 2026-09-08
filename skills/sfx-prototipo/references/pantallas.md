# Prototipo de pantallas — "¿cómo tendría que verse esto?"

Varias versiones **radicalmente distintas** de la misma pantalla, que se cambian con un clic. Mirás,
elegís una (o robás pedazos de cada una) y tirás el resto.

Si la pregunta es sobre reglas o estados y no sobre cómo se ve, estás en la rama equivocada:
[logica.md](logica.md).

## Cuándo es ésta

- *"¿Cómo tendría que verse esta pantalla?"*
- *"Quiero ver tres opciones del tablero antes de comprometerme."*
- Cada vez que si no lo hacés te vas a pasar un día eligiendo entre tres maquetas que sólo existen
  en tu cabeza.

## Dónde viven las versiones — y acá T1 es distinto a todo lo demás

**La regla general de Matt, que es correcta y hay que conservar:** una pantalla se juzga mucho mejor
cuando está **rozando el resto de la aplicación** — el encabezado de verdad, los datos de verdad, la
densidad de verdad. Una pantalla suelta es un vacío donde **todas las versiones parecen buenas**.

De ahí salen dos formas, y cuál te toca lo decide si ya hay aplicación:

### Si la aplicación ya existe → adentro de la página real

Las versiones se dibujan **en la misma ruta que ya existe**, elegidas por un parámetro en la URL. La
búsqueda de datos, los permisos y todo lo demás quedan igual: **lo único que cambia es lo que se
dibuja.**

Es la mejor de las dos, y siempre que haya una página donde meterlas, es ésta.

### Si todavía no hay nada → un archivo suelto, y se dice

**Éste es el caso normal en T1**, porque en el ①–⑤ el proyecto todavía no existe. Un HTML solo, todo
adentro, con las versiones y una barra para cambiar entre ellas.

**Y va con una advertencia escrita en el prototipo mismo**, porque es la trampa de esta rama:

> *Esto se está mirando en el vacío. Sin datos reales, sin el resto de la pantalla y sin densidad
> real, las tres versiones van a parecer mejores de lo que son.*

Sin esa línea, el prototipo de T1 miente por omisión.

## Los cinco pasos

### 1. La pregunta, y cuántas versiones

**Tres por defecto.** Más de cinco deja de ser variedad y pasa a ser ruido, así que ahí está el
techo.

Escribí el plan en una línea, arriba del archivo:

> *"Tres versiones de la pantalla de configuración, en un archivo suelto porque todavía no hay app."*

### 2. Que sean de verdad distintas

Cada versión tiene que **estar en desacuerdo con las otras sobre la estructura**: otro reparto del
espacio, otra jerarquía de qué se ve primero, otra acción principal.

**Tres grillas de tarjetas con distinto color no son tres versiones: son empapelado.** Si dos te
salen parecidas, rehacé una prohibiéndote explícitamente lo que usaste en la primera.

### 3. La barra para cambiar

Una barra fija abajo y al centro, con tres cosas: **flecha izquierda**, **el nombre de la versión
actual** (la letra y, si tiene nombre, el nombre: `B (barra lateral)`), **flecha derecha**. Las dos
flechas dan la vuelta.

Y tres detalles que la hacen usable:

- **las flechas del teclado** también cambian — salvo que el foco esté en un campo de texto
- **la versión queda en la URL** (o en la dirección del archivo), así el link se puede compartir y
  sobrevive a recargar
- **la barra se ve claramente ajena al diseño** — pastilla de alto contraste, sombrita— para que
  nadie la confunda con parte de lo que está evaluando

Si esto se está montando adentro de una aplicación de verdad, **la barra no puede llegar a
producción**: va detrás de una condición de entorno de desarrollo.

### 4. Mostráselo

Pasale el archivo o el link, con las letras de las versiones. La devolución que sirve casi nunca es
*"me gusta la B"*: es **"quiero el encabezado de la B con la barra lateral de la C"** — y ése es el
diseño que en realidad quería.

### 5. Guardá la respuesta, tirá el resto

Lo que sobrevive es **cuál ganó y por qué**, y entra a `.docs/entrevista.md` como decisión con
procedencia `probado`:

```markdown
- La pantalla va con barra lateral fija, no con pestañas: con 12 categorías las pestañas
  se cortan. [probado]
  - .docs/prototipos/forma-de-la-configuracion/
```

Las versiones que perdieron **no se borran y no se mudan**: se quedan en la carpeta del prototipo,
que ya está marcada como descartable. Son la única prueba de que la que ganó se comparó contra algo.

## Lo que arruina un prototipo de pantallas

| Se rompe cuando… | Lo que sí |
|---|---|
| las versiones cambian sólo el color o el texto | que estén en desacuerdo sobre la **estructura** |
| comparten demasiado código | un encabezado común está bien; un armazón común mata el punto — cada versión tiene que poder tirarlo |
| las enchufás a datos que se escriben | que sólo lean; la pregunta es cómo se ve, no si el backend anda |
| pasás la versión ganadora directo a producción | se reescribe: se dibujó sin manejo de errores y sin tests |
| la mirás en el vacío y no lo decís | la advertencia escrita, arriba de todo |
