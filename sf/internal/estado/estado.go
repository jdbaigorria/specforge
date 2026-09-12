// Package estado lee y escribe el `estado.json`: el único archivo que `sf` escribe.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ ESTE PAQUETE EXISTE, Y POR QUÉ NO SE MIGRÓ EL `state.go` VIEJO
// ────────────────────────────────────────────────────────────────────────────
//
// El CLI viejo tenía un `state.go` que decía, en su propio comentario:
//
//	"NO guarda estado nuevo: TODO se deriva de lo que ya existe en features.json"
//
// Este paquete existe por EXACTAMENTE LO CONTRARIO, y esa inversión es la razón
// por la que se escribe de cero en vez de migrarlo (construccion.md §5):
//
//	"Guardá lo que decidió Javier, y lo que sf vio pasar.
//	 Todo lo demás sale de los archivos."          (maquina-estados.md §9)
//
// Los archivos dicen QUÉ SE PRODUJO. El estado dice QUÉ SE APROBÓ — y eso no
// está escrito en ningún lado porque salió de una cabeza. El sello del ⑥, el
// `rojo` de un lote que ya pasó, el modelo que estás usando ahora: nada de eso
// se puede derivar de mirar el repo.
//
// La regla de qué campo entra acá y cuál no es dura, y ya tumbó tres campos:
//
//	R6 — un campo deducible de otro es un campo que se desincroniza.
//
// Por eso NO están: `verde` (deducible de `commit`: si sf no deja commitear en
// rojo, hay commit ⇒ estaba verde), `modelo_recomendado` (lo escribe el ⑯ en
// tareas.json) ni `paso` (cada feature tiene el suyo).
package estado

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Archivo es dónde vive el estado, relativo a la raíz del proyecto.
//
// Va en `.docs/` junto al resto de los artefactos y NO en un directorio oculto
// propio, por una razón del diseño: el estado se versiona con el repo. En el ⑱
// cambiás de modelo y el que implementa no estuvo en la conversación — si el
// estado viviera fuera del repo, no le llegaría. (session.md §6, regla dura 3)
const Archivo = ".docs/estado.json"

// ────────────────────────────────────────────────────────────────────────────
// El vocabulario de en qué anda una feature
// ────────────────────────────────────────────────────────────────────────────
//
// Son los cuatro estados de feature de la máquina (maquina-estados.md §2) más
// dos de espera, donde la feature existe pero nadie está trabajando en ella:
//
//	planificacion → implementar → revision → cierre        los 4 con trabajo
//	planificada                                            plan aprobado, esperando en la cola
//	cerrada                                                terminada y archivada
//
// `planificada` es la puerta "otra feature" del ⑰: aprobaste el plan pero
// elegiste ir a planificar otra en vez de implementar esta (maquina-estados §6).
//
// NO existe un valor "pendiente". Una feature que está en el roadmap y todavía
// no arrancó simplemente NO APARECE en el mapa `Features`. Es R6 otra vez: se
// deduce de la ausencia, así que no se guarda.
const (
	Planificacion = "planificacion"
	Planificada   = "planificada"
	Implementar   = "implementar"
	Revision      = "revision"
	Cierre        = "cierre"
	Cerrada       = "cerrada"
)

// ────────────────────────────────────────────────────────────────────────────
// Los tres caminos — cuánto proceso pide una feature
// ────────────────────────────────────────────────────────────────────────────
//
// Es UN campo que saltea estados, no tres máquinas (maquina-estados.md §8,
// H20). Los tres recorren la misma máquina; lo único que cambia es por cuántos
// de sus estados pasan:
//
//	Largo   planificacion → implementar → revision → cierre    tipo: us
//	Chico                   implementar → revision → cierre    tipo: chico
//	Corto                   implementar →            cierre    tipo: bug
//
// `Chico` es el del medio y es el que faltaba. Sin él, un cambio acotado sobre
// código que ya existe —un flag, un endpoint, un arreglo de una línea que no es
// un bug— pagaba las tres opciones, la spec-design y el tareas.json del ⑫–⑯
// para producir un solo lote. La clasificación la escribe el ⑨ en el
// frontmatter de la historia; el ⑩ la hereda.
//
// **El trinquete corre para un solo lado**, y lo aplica `Feature.Camino`: una
// feature puede subir de Chico a Largo cuando la máquina comprueba que no era
// chica, y no puede volver a bajar. Ver `Ampliada`.
type Camino int

const (
	// Largo es el default y el que gana ante la duda: de más se puede saltear
	// después, de menos ya se implementó sin diseño.
	Largo Camino = iota
	Chico
	Corto
)

// Efectivo aplica el camino sobre el estado guardado.
//
// ────────────────────────────────────────────────────────────────────────────
// UN SOLO LUGAR QUE DECIDA, PORQUE TENERLO EN DOS YA COSTÓ CARO
// ────────────────────────────────────────────────────────────────────────────
//
// La REGLA vivía copiada en `sf next` y en `sf done`, y faltaba en `sf lote
// start` y en `sf context` — así que los comandos se contradecían sobre la
// misma feature: `next` decía "implementar, corré lote start" y `lote start`
// contestaba "esto está en planificacion".
//
// Por eso esto es una función y no un `if` en cada uno: el que decide por
// estado tiene que preguntar acá, y agregar un comando nuevo no puede volver a
// olvidarse de la regla.
//
// Chico y Corto saltean lo MISMO acá —la planificación—, y eso no es un
// descuido: se separan más adelante. El salteo de la revisión no se puede
// resolver mirando sólo el estado guardado —depende de que TODOS los lotes
// estén commiteados—, así que lo sigue haciendo `cerrarLote`, que es el único
// que tiene esa información, y ahí Chico y Corto se bifurcan.
func Efectivo(actual string, c Camino) string {
	if c != Largo && actual == Planificacion {
		return Implementar
	}
	return actual
}

// Estado es el archivo entero.
//
// Las etiquetas `json:"..."` le dicen a encoding/json cómo se llama cada campo
// en el archivo. Sin ellas, Go usaría el nombre del campo Go tal cual —
// `Producto`, con mayúscula— y el JSON quedaría distinto al del diseño.
//
// Los campos arrancan en mayúscula porque en Go eso es lo que los hace públicos
// (visibles desde otros paquetes). Y encoding/json SOLO ve los campos públicos:
// un campo en minúscula no se serializa nunca, ni con etiqueta.
type Estado struct {
	Producto Producto `json:"producto"`

	// FeatureActual es el id de la única feature que se está tocando.
	//
	// Es UN string y no una lista porque no hay features en paralelo: hay una
	// cola y se toca una por vez (maquina-estados.md §6). Cuando no hay
	// ninguna en curso queda en "".
	FeatureActual string `json:"feature_actual"`

	// Features es un mapa id → feature.
	//
	// Mapa y no slice porque el acceso real siempre es por id (`f-2`), nunca
	// por posición. El ORDEN no vive acá: vive en el roadmap.json, que es el
	// único que sabe en qué secuencia van (artefactos.md §8).
	//
	// Y son punteros (*Feature) por una razón de Go que muerde seguido: NO SE
	// PUEDE tomar la dirección de un valor guardado en un mapa. Con
	// map[string]Feature, esto no compila:
	//
	//	e.Features["f-2"].IntentosFallidos++    ← error
	//
	// porque el mapa devuelve una COPIA, y modificar la copia no hace nada.
	// Con punteros, el mapa guarda la dirección y mutar funciona.
	Features map[string]*Feature `json:"features"`
}

// Producto son los sellos que sólo existen porque Javier dijo que sí.
//
// Los cinco estados de producto corren una vez, y de los cinco sólo tres dejan
// rastro acá. Los otros dos NO se guardan porque se deducen de los archivos
// (regla 1.5 de artefactos.md):
//
//	backlog  →  ¿existen los us-#.md?
//	roadmap  →  ¿existe el roadmap.json?
type Producto struct {
	// BriefSellado es el veredicto del ⑥: "hacelo" | "pivotea" | "no-lo-hagas".
	// Vacío = todavía no se selló.
	BriefSellado string `json:"brief_sellado"`

	// PrdHash es la huella del prd.md cuando salió el backlog.
	//
	// Sirve para un aviso: si el PRD cambia después de que las historias
	// salieron de él, las us-# quedaron viejas y nadie se entera. sf compara
	// dos strings y avisa — NO frena (artefactos.md §5).
	PrdHash string `json:"prd_hash"`

	// ConstitucionSellada es el sello del ⑧.
	ConstitucionSellada bool `json:"constitucion_sellada"`

	// Rechazo es el motivo del último `sf reject` sobre un artefacto de producto.
	//
	// Se guarda y no se imprime y ya, porque el que va a rehacer el trabajo es
	// un subagente NUEVO que arranca de cero: sin el motivo vuelve a proponer lo
	// mismo. `sf context` lo mete en el sobre, que es donde lo va a leer.
	//
	// Se limpia solo cuando el artefacto se aprueba.
	Rechazo string `json:"rechazo,omitempty"`

	// BacklogVisto es el enter de la parada barata del ⑨.
	//
	// ────────────────────────────────────────────────────────────────────
	// ESTE CAMPO NO ESTABA EN EL DISEÑO. Apareció al construir `sf next`.
	// ────────────────────────────────────────────────────────────────────
	//
	// Las paradas 🛑 dejan rastro solas: el sello del ⑥ es un veredicto, el
	// del ⑧ es un bool, el del ⑰ mueve la feature a `planificada`. Pero la
	// PARADA BARATA del ⑨ no sella nada — es un enter, "mirá y seguí".
	//
	// Y sin rastro se cuelga: `sf next` diría ⏸, Javier haría enter, y el
	// próximo `sf next` volvería a decir ⏸ para siempre.
	//
	// Su hermana, la ⏸ del ㉓, NO necesita campo: ahí `sf approve` archiva la
	// carpeta y mergea, y eso sí deja huella. Por eso hay uno solo y no dos.
	//
	// No rompe R6 (no es deducible de nada) y encaja exacto en la regla del
	// estado: "guardá lo que decidió Javier". Que haya mirado el backlog es,
	// literalmente, una decisión suya que ningún archivo registra.
	BacklogVisto bool `json:"backlog_visto"`
}

// Feature es una vuelta del ciclo ⑪–㉓.
type Feature struct {
	// Estado es una de las constantes de arriba.
	Estado string `json:"estado"`

	// BaseCommit es el HEAD del repo cuando se terminó de planificar.
	//
	// Existe para el plan que envejece (maquina-estados.md §7): si planificás
	// f-2 y después implementás f-1, la spec de f-2 quedó mirando un repo que
	// ya cambió — y el implementador que no encuentra lo que la spec dice
	// IMPROVISA, que es donde nacen los mocks.
	//
	// El mecanismo es de pobre a propósito: guardar un número hoy y compararlo
	// mañana. sf no sabe QUÉ cambió ni si importa; sabe que el suelo se movió
	// desde que se dibujó el plano. Avisa, no frena.
	BaseCommit string `json:"base_commit"`

	// Modelo es el que se está usando AHORA, no el recomendado.
	//
	// Es uno de los dos campos que no se pueden deducir de ningún archivo: el
	// recomendado lo escribió el ⑯ en tareas.json, pero cuando Javier lo sube
	// en el ⑳ ("elevo el modelo porque la recomendación no fue suficiente"),
	// esa decisión no queda registrada en ningún otro lado.
	Modelo string `json:"modelo"`

	// IntentosFallidos cuenta los `sf done` que dieron ✗ seguidos.
	//
	// Es el freno de la cuarta parada, ME TRABÉ (maquina-estados.md §5).
	// Cuando sf da rojo el orquestador relanza, y un bucle sin freno se cuelga.
	// El freno ya estaba en el ⑳ ("si falla varias veces, entro yo"); esto lo
	// convierte de sensación en número. Se resetea a 0 cuando el lote cierra.
	IntentosFallidos int `json:"intentos_fallidos"`

	// RondasRevision cuenta las veces que el ㉑ mandó esta feature de vuelta.
	//
	// ────────────────────────────────────────────────────────────────────
	// EL BUCLE QUE NO TENÍA FONDO
	// ────────────────────────────────────────────────────────────────────
	//
	// `cerrarRevision` es la única transición que va PARA ATRÁS: con un
	// hallazgo abierto, la feature vuelve a implementar y la revisión se
	// rehace entera. Eso no tenía tope. Un revisor que encuentra el mismo
	// problema cinco veces y un implementador que no lo entiende cinco veces
	// son un bucle que gasta plata y no termina — y nadie se entera hasta la
	// factura.
	//
	// Es el mismo remedio que `IntentosFallidos` le dio al ⑳, aplicado al
	// otro bucle de la máquina, y por la misma razón: el freno ya existía
	// como sensación ("esto ya lo vimos") y esto lo convierte en número.
	//
	// Se resetea cuando la revisión sale limpia, cuando `sf model` sube el
	// modelo y cuando `sf dismiss` descarta un hallazgo — los tres son
	// "el bucle se destrabó", igual que para IntentosFallidos.
	RondasRevision int `json:"rondas_revision,omitempty"`

	// Ampliada es el trinquete: esta feature se declaró chica y no lo era.
	//
	// ────────────────────────────────────────────────────────────────────
	// POR QUÉ ES UN CAMPO Y NO SE DEDUCE (R6 NO APLICA)
	// ────────────────────────────────────────────────────────────────────
	//
	// La historia sigue diciendo `tipo: chico` — el que la escribió no se
	// enteró de nada. Lo único que sabe que no era chica es sf, que vio a la
	// feature pedir un segundo lote. Ese hecho no está escrito en ningún
	// archivo, así que si sf no lo anota se pierde, y en el próximo `sf next`
	// el camino corto vuelve a aplicarse y la feature queda dando vueltas.
	//
	// Que sea de una sola vía es el punto: `Camino` lo lee y nunca lo apaga.
	// Una feature que se amplió no vuelve a ser chica ni editando la historia.
	Ampliada bool `json:"ampliada,omitempty"`

	// Rechazo es el motivo del "pido cambios" del ⑰.
	//
	// Igual que el de producto: viaja en el sobre para que el que rehaga el
	// bloque ⑫–⑯ sepa qué estuvo mal.
	Rechazo string `json:"rechazo,omitempty"`

	// Lotes son las unidades de trabajo y de commit.
	//
	// Acá SÍ es slice y no mapa, al revés que Features: el lote se identifica
	// por un número consecutivo y el orden importa (el lote 2 va después del
	// 1). Lo que nunca hay que guardar es "cuál es el lote actual": es el
	// primero sin commit, y eso se deduce. R6.
	Lotes []Lote `json:"lotes,omitempty"`
}

// Lote es un semáforo, no una bitácora.
//
// Su único consumidor es sf, en el instante en que decide si te deja pasar. Por
// eso no tiene descripción, ni fecha, ni quién lo hizo: nadie lee eso nunca.
// El registro del trabajo es el commit, que ya existe (artefactos.md §10).
type Lote struct {
	Lote int `json:"lote"`

	// Rojo es la memoria de que los tests fallaban antes de escribir el código.
	//
	// Es el otro campo indeducible, y de los dos es el más frágil: el momento
	// en que los tests fallaban YA PASÓ y no dejó huella en ningún lado. Si sf
	// no lo anota cuando lo ve, se pierde para siempre.
	//
	//	"Un test que pasa antes de que exista el código es un test de mentira."
	Rojo bool `json:"rojo"`

	// HashTests es la huella de los archivos de test, tomada en el rojo.
	//
	// Tapa el agujero más astuto del dolor #8: sf ve rojo, el subagente
	// trabaja, sf ve verde… y lo que cambió entre medio fue EL TEST. Un
	// subagente que no logra implementar puede ablandar el assert y llegar a
	// verde. Se compara en `sf done` y son dos strings (artefactos.md §10).
	HashTests string `json:"hash_tests"`

	// Commit es el hash del commit del lote. Puntero para que pueda ser null.
	//
	// Acá el tipo carga significado, y es la razón por la que NO existe un
	// campo `verde`: si sf no deja commitear en rojo, TENER COMMIT YA SIGNIFICA
	// QUE ESTABA VERDE. Un campo deducible de otro es un campo que se
	// desincroniza — R6, y fue el primer campo que esa regla tumbó.
	//
	// *string y no string porque hay que distinguir "todavía no hay commit"
	// (null) de "hay commit" (un hash). Con un string pelado, el cero de Go es
	// "" y no se puede saber si es que falta o si alguien guardó vacío.
	Commit *string `json:"commit"`
}

// ErrNoHay es lo que devuelve Leer cuando el proyecto todavía no tiene estado.
//
// Es un error CON NOMBRE y no un `errors.New` anónimo adentro de la función,
// para que quien llama pueda preguntarle con errors.Is en vez de comparar el
// texto del mensaje:
//
//	if errors.Is(err, estado.ErrNoHay) { … }
var ErrNoHay = errors.New("no hay estado.json: ¿corriste sf init?")

// Leer carga el estado desde la raíz del proyecto.
//
// Usa os.ReadFile y no os.Open porque el archivo es chico y se lee entero:
// ReadFile abre, lee todo y CIERRA SOLO. Con os.Open habría que acordarse del
// defer f.Close(), y no gana nada acá.
func Leer(raiz string) (*Estado, error) {
	ruta := filepath.Join(raiz, Archivo)

	b, err := os.ReadFile(ruta)
	if err != nil {
		// os.IsNotExist distingue "el archivo no está" de "está pero no lo
		// puedo leer" (permisos, disco). El primero es un estado normal del
		// flujo; el segundo es un problema de verdad y hay que dejarlo pasar
		// tal cual, sin disfrazarlo.
		if os.IsNotExist(err) {
			return nil, ErrNoHay
		}
		return nil, fmt.Errorf("leyendo %s: %w", ruta, err)
	}

	var e Estado
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, fmt.Errorf("%s está corrupto: %w", ruta, err)
	}

	// Un JSON sin la clave "features" deja el mapa en nil, y escribir en un
	// mapa nil es panic en Go (leer, en cambio, funciona y devuelve el cero).
	// Inicializarlo acá le evita el chequeo a todos los que llamen.
	if e.Features == nil {
		e.Features = map[string]*Feature{}
	}

	return &e, nil
}

// Guardar escribe el estado, y lo hace de forma atómica.
//
// Escribe primero a un archivo temporal y recién después lo renombra encima del
// bueno. El rename dentro del mismo directorio es atómico a nivel del sistema
// operativo: o quedó el viejo o quedó el nuevo, nunca medio archivo.
//
// Parece paranoia para un JSON de 20 líneas, y no lo es: este archivo es la
// única fuente de verdad de dónde estás. Si un ctrl-C lo parte por la mitad, se
// pierde el sello del ⑥, el `rojo` de los lotes y el modelo que estabas usando
// — todo lo que NO se puede recuperar mirando el repo, que es justamente el
// criterio por el que están acá.
func (e *Estado) Guardar(raiz string) error {
	ruta := filepath.Join(raiz, Archivo)

	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		return fmt.Errorf("creando %s: %w", filepath.Dir(ruta), err)
	}

	// MarshalIndent y no Marshal: el estado se versiona con el repo, así que un
	// `git diff` tiene que ser legible. Con Marshal sería una sola línea y
	// cualquier cambio se vería como "toda la línea cambió".
	b, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return fmt.Errorf("serializando el estado: %w", err)
	}
	b = append(b, '\n') // los archivos de texto terminan en newline

	// CreateTemp en el MISMO directorio que el destino, no en /tmp: el rename
	// atómico sólo funciona dentro del mismo filesystem, y /tmp puede estar en
	// otro. El "*" del patrón lo reemplaza Go por algo aleatorio.
	tmp, err := os.CreateTemp(filepath.Dir(ruta), ".estado-*.json")
	if err != nil {
		return fmt.Errorf("creando el temporal: %w", err)
	}
	// Si algo falla más abajo, este defer borra el temporal. Cuando todo sale
	// bien el rename ya se lo llevó y el Remove falla en silencio, que es lo
	// que queremos.
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(b); err != nil {
		tmp.Close()
		return fmt.Errorf("escribiendo el temporal: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("cerrando el temporal: %w", err)
	}

	if err := os.Rename(tmp.Name(), ruta); err != nil {
		return fmt.Errorf("moviendo el temporal a %s: %w", ruta, err)
	}
	return nil
}

// ────────────────────────────────────────────────────────────────────────────
// Consultas — lo que va a preguntar `sf next`
// ────────────────────────────────────────────────────────────────────────────

// Actual devuelve la feature en curso, o nil si no hay ninguna.
//
// Devuelve (*Feature, bool) en vez de sólo el puntero para poder distinguir
// "no hay feature actual" de "hay una y está vacía". Es el mismo idioma que ya
// usás sin saberlo cada vez que hacés `v, ok := m[k]` sobre un mapa.
func (e *Estado) Actual() (*Feature, bool) {
	if e.FeatureActual == "" {
		return nil, false
	}
	f, ok := e.Features[e.FeatureActual]
	return f, ok
}

// LoteActual devuelve el primer lote sin commit — el que está en curso.
//
// Esto es R6 hecho código: NO hay un campo "lote_actual" en el estado, porque
// se deduce. Un campo que se puede calcular es un campo que en algún momento
// va a quedar desincronizado con lo que dice la realidad.
func (f *Feature) LoteActual() (*Lote, bool) {
	for i := range f.Lotes {
		if f.Lotes[i].Commit == nil {
			// &f.Lotes[i] y no &lote de un `for _, lote := range`: la variable
			// del range es una COPIA, y devolver su dirección daría un puntero
			// a algo que nadie más ve. Sobre un slice sí se puede indexar y
			// tomar la dirección del elemento real.
			return &f.Lotes[i], true
		}
	}
	return nil, false
}

// Camino aplica el trinquete sobre lo que declararon las historias.
//
// ────────────────────────────────────────────────────────────────────────────
// SUBE, NO BAJA — Y POR ESO NO ES EL CLASIFICADOR
// ────────────────────────────────────────────────────────────────────────────
//
// `historia.CaminoDe` contesta qué se DECLARÓ; ésta contesta por dónde va de
// verdad. Son dos preguntas distintas y la segunda tiene memoria: una vez que
// la máquina comprobó que una feature chica necesitaba más, editar la historia
// para volver a `tipo: chico` no la baja.
//
// La regla sale de mirar cómo se rompe la clasificación en la práctica: nadie
// declara chico algo que sabe grande. Se declara chico lo que PARECE chico, y
// la complejidad escondida aparece implementando — o sea, después de que el
// camino ya se eligió. Un trinquete de dos vías dejaría que el mismo optimismo
// que erró la primera vez vuelva a errar; de una vía, el error se paga una vez.
func (f *Feature) Camino(declarado Camino) Camino {
	if f.Ampliada {
		return Largo
	}
	return declarado
}

// SinSembrar dice que esta feature nunca pasó por `sf lote start`.
//
// ────────────────────────────────────────────────────────────────────────────
// NO ES LO MISMO QUE "NO QUEDA NINGÚN LOTE ABIERTO"
// ────────────────────────────────────────────────────────────────────────────
//
// LoteActual() devuelve (nil, false) en DOS situaciones que no tienen nada que
// ver entre sí:
//
//	f.Lotes vacío              nadie empezó         ← acá
//	todos con Commit != nil    todos terminaron
//
// Confundirlas era el agujero más grande del binario: los lotes se siembran
// recién cuando alguien va a trabajarlos, así que una lista vacía significa que
// NADIE EMPEZÓ — y eso se leía como "terminado". Con eso, un solo `sf done`
// después de aprobar el plan movía la feature a revisión sin branch, sin tests,
// sin código y sin commit.
//
// Es un hecho contable, no un juicio (R3): `len(f.Lotes) == 0`. Y no agrega
// estado nuevo — se deriva de lo que ya está en el estado.json (R6).
func (f *Feature) SinSembrar() bool { return len(f.Lotes) == 0 }

// Cerrados cuenta los lotes que ya tienen commit.
//
// Es la mitad de "lote 2 de 4", que es lo que sf next le muestra al
// orquestador (superficie-sf.md §2).
func (f *Feature) Cerrados() int {
	n := 0
	for _, l := range f.Lotes {
		if l.Commit != nil {
			n++
		}
	}
	return n
}
