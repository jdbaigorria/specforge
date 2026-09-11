// Package compuerta es el segundo verbo de sf, y el que importa.
//
// ────────────────────────────────────────────────────────────────────────────
// SIN ESTO, TODO LO DEMÁS ES UNA LISTA DE TAREAS
// ────────────────────────────────────────────────────────────────────────────
//
//	expone      "estás en el ⑬, ahora toca el ⑭"   → sin esto, una lista
//	comprueba   "no terminaste: falta la branch"    → sin esto, una sugerencia
//	                                                   que el agente puede ignorar
//
// sf no maneja el auto —no agarra el volante ni elige la ruta— pero DA VERDE O
// ROJO, Y EN ROJO NO SE PASA. Su poder es uno solo: es el único que puede mover
// el estado, y sólo lo mueve cuando lo comprobó él mismo.
//
// ────────────────────────────────────────────────────────────────────────────
// LA REGLA DURA
// ────────────────────────────────────────────────────────────────────────────
//
//	El estado avanza con hechos comprobados, nunca con la palabra del que
//	trabajó.
//
// El subagente dice "terminé"; sf NO LE CREE: cuenta los archivos él, corre los
// tests él, mira si hay branch él. Si falta algo, el estado no se mueve y el
// orquestador recibe QUÉ FALTA.
//
// ────────────────────────────────────────────────────────────────────────────
// Y LA QUE LA LIMITA, QUE ES IGUAL DE IMPORTANTE
// ────────────────────────────────────────────────────────────────────────────
//
//	R3 — una compuerta frena sobre un HECHO; un juez OPINA.
//
//	sf puede frenarte porque el test falló — eso no lo discute nadie.
//	NO puede frenarte porque un modelo dijo que tu diseño está flojo.
//
// Por eso acá no hay ni una decisión que necesite criterio. Todas son: contar
// una lista, comparar dos strings, o preguntar si un archivo existe.
package compuerta

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/jdbaigorria/specforge/sf/internal/constitucion"
	"github.com/jdbaigorria/specforge/sf/internal/docs"
	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/frontmatter"
	"github.com/jdbaigorria/specforge/sf/internal/historia"
	"github.com/jdbaigorria/specforge/sf/internal/revision"
	"github.com/jdbaigorria/specforge/sf/internal/roadmap"
	"github.com/jdbaigorria/specforge/sf/internal/suite"
	"github.com/jdbaigorria/specforge/sf/internal/tareas"
)

// Resultado es el veredicto de correr las compuertas de un estado.
//
// ────────────────────────────────────────────────────────────────────────────
// UN TEST NO TIENE DOS SALIDAS: TIENE TRES
// ────────────────────────────────────────────────────────────────────────────
//
//	✓  pasó              lo comprobé, salió bien
//	✗  falló             lo comprobé, salió mal
//	?  NO EVALUABLE      esto NO lo pude comprobar
//
// La tercera es la que faltaba, y es la que cambia decisiones. "No pude
// comprobar si esto ya existe" no es un verde. UN VERDE QUE EN REALIDAD QUIERE
// DECIR "NO MIRÉ" ES PEOR QUE UN ROJO, porque el que lo lee no tiene forma de
// enterarse.
//
// La regla ya estaba escrita en este repo, en `lanzar/resumen.go`:
//
//	"LO QUE EL ARNÉS NO DA, NO ESTÁ. No va en cero, no va en false, no va."
//
// Estaba en el paquete equivocado. Acá está la misma idea para las compuertas.
//
// ────────────────────────────────────────────────────────────────────────────
// Y POR QUÉ APARECEN LOS ✓, QUE ANTES SE TIRABAN A PROPÓSITO
// ────────────────────────────────────────────────────────────────────────────
//
// Porque son DOS LECTORES y hasta hoy había un solo texto:
//
//	el subagente que arregla   sólo necesita saber QUÉ FALTA   → Texto()
//	Javier que firma           necesita ver la EVIDENCIA        → Acta()
//
// El comentario viejo de Texto() —"un veredicto que enumera todo lo que salió
// bien es ruido"— sigue siendo cierto para el primero, y es exactamente falso
// para el segundo.
type Resultado struct {
	// Fallas son los motivos por los que el estado NO se mueve.
	Fallas []string

	// Avisos son cosas que hay que saber y que NO frenan.
	//
	// La distinción es de diseño, no de estilo: el dolor #6 (librerías fuera de
	// la constitución) sería trivial de impedir, y va acá igual porque Javier
	// dijo dos veces que la última palabra es suya. Una herramienta que frena
	// sola rompe esa regla.
	Avisos []string

	// Ok son las comprobaciones que SÍ se hicieron y salieron bien, con el
	// hecho al lado. No las lee el subagente; las lee Javier en el acta.
	Ok []string

	// NoSeSabe es lo que quedó fuera del alcance de la máquina.
	//
	// NO frena: no poder comprobar algo no es lo mismo que comprobar que está
	// mal. Frenar sobre lo que no se sabe sería opinar, y R3 lo prohíbe. Lo que
	// sí hace es LLEGAR A LOS OJOS del que decide, que es todo el punto.
	NoSeSabe []string

	// Medido son los números crudos, sin interpretar.
	//
	// Existe porque el 2026-09-05 los números ESTABAN —17 retrieved, 13 links
	// contra 1 y 0— y no llegaban a ninguna parte. Un número crudo al lado de
	// una decisión cambia la decisión; el mismo número adentro de una prosa no.
	Medido []Medida
}

// Medida es un número crudo con su nombre. Sin unidades mágicas ni formato:
// lo que se guarda es lo que se imprime.
type Medida struct {
	Que    string
	Cuanto string
}

// Pasa dice si el estado puede moverse.
func (r Resultado) Pasa() bool { return len(r.Fallas) == 0 }

func (r *Resultado) falla(formato string, args ...any) {
	r.Fallas = append(r.Fallas, fmt.Sprintf(formato, args...))
}

func (r *Resultado) avisa(formato string, args ...any) {
	r.Avisos = append(r.Avisos, fmt.Sprintf(formato, args...))
}

// ok anota una comprobación que salió bien, con el hecho al lado.
func (r *Resultado) ok(formato string, args ...any) {
	r.Ok = append(r.Ok, fmt.Sprintf(formato, args...))
}

// noSeSabe anota algo que la máquina NO pudo comprobar.
//
// Se usa donde la respuesta honesta es "no sé", no donde es "no". La diferencia
// no es de matiz: "no encontré competidores" y "no pude buscar competidores"
// llevan a decisiones opuestas.
func (r *Resultado) noSeSabe(formato string, args ...any) {
	r.NoSeSabe = append(r.NoSeSabe, fmt.Sprintf(formato, args...))
}

// mide anota un número crudo para el acta.
func (r *Resultado) mide(que, cuanto string) {
	r.Medido = append(r.Medido, Medida{Que: que, Cuanto: cuanto})
}

// Texto arma el veredicto PARA EL QUE TRABAJÓ — un subagente que va a arreglar
// lo que falta y nada más.
//
// Los ✓ no se listan: sólo importa lo que falta. Un veredicto que enumera todo
// lo que salió bien es ruido, y el que lo lee tiene que buscar la ✗ entre
// quince ✓.
//
// Ese razonamiento es correcto PARA ESTE LECTOR. Para el otro —Javier, en una
// parada, decidiendo— es exactamente al revés, y por eso existe Acta().
func (r Resultado) Texto() string {
	var b strings.Builder
	for _, a := range r.Avisos {
		fmt.Fprintf(&b, "⚠ %s\n", a)
	}
	if r.Pasa() {
		b.WriteString("✓ listo\n")
		return b.String()
	}
	b.WriteString("✗ No avanzo.\n")
	for _, f := range r.Fallas {
		fmt.Fprintf(&b, "  · %s\n", f)
	}
	return b.String()
}

// Acta arma el informe PARA EL QUE FIRMA.
//
// ────────────────────────────────────────────────────────────────────────────
// COMPROBAR ≠ DECIDIR
// ────────────────────────────────────────────────────────────────────────────
//
//	comprobar   trabajo mecánico   →  la máquina, SIEMPRE, sin pedir permiso
//	decidir     juicio             →  Javier, y SÓLO donde hay algo que elegir
//
// El error de la v1 no fue poner al humano de compuerta: fue ponerlo de
// compuerta Y DE INSPECTOR. Tenía que leer, contar, verificar y ADEMÁS decidir,
// veinte veces por feature — y la cuarta ya es apretar Enter sin mirar.
//
// El acta separa los dos trabajos: la máquina comprueba y entrega la evidencia;
// el humano no valida, LEE Y DECIDE. Es el reporte de una corrida de tests, no
// un formulario de aprobación.
//
// Los tres bloques van SIEMPRE, incluso vacíos con su "—": un acta a la que le
// falta el bloque "no pude comprobar" se lee como si no hubiera nada que no se
// pudiera comprobar, y eso es justo lo que no se quiere decir.
func (r Resultado) Acta() string {
	var b strings.Builder

	bloque := func(titulo, marca string, lineas []string) {
		fmt.Fprintf(&b, "%s\n", titulo)
		if len(lineas) == 0 {
			b.WriteString("  —\n")
		}
		for _, l := range lineas {
			fmt.Fprintf(&b, "  %s %s\n", marca, l)
		}
		b.WriteString("\n")
	}

	bloque("COMPROBÉ", "✓", r.Ok)

	fmt.Fprintf(&b, "MEDÍ\n")
	if len(r.Medido) == 0 {
		b.WriteString("  —\n")
	}
	ancho := 0
	for _, m := range r.Medido {
		if len(m.Que) > ancho {
			ancho = len(m.Que)
		}
	}
	for _, m := range r.Medido {
		fmt.Fprintf(&b, "  %-*s  %s\n", ancho, m.Que, m.Cuanto)
	}
	b.WriteString("\n")

	bloque("NO PUEDO COMPROBAR", "?", r.NoSeSabe)

	if len(r.Avisos) > 0 {
		bloque("OJO", "⚠", r.Avisos)
	}

	// Las fallas no deberían llegar acá —el acta se imprime en una parada, y a
	// una parada se llega con la compuerta en verde— pero si llegan, se
	// muestran. Un acta que esconde una falla es peor que no tener acta.
	if !r.Pasa() {
		bloque("NO AVANZO", "✗", r.Fallas)
	}

	return b.String()
}

// ────────────────────────────────────────────────────────────────────────────
// Los cinco de producto
// ────────────────────────────────────────────────────────────────────────────

// Brief comprueba que el ⑥ tenga qué sellar.
//
// No comprueba que el brief sea BUENO —eso es juicio y es de Javier— sino que
// el ①–⑤ dejó lo que tenía que dejar. Es la diferencia entre una compuerta y un
// juez.
//
// ────────────────────────────────────────────────────────────────────────────
// TRES ARCHIVOS, NO UNO — Y POR QUÉ
// ────────────────────────────────────────────────────────────────────────────
//
//	brief.md        el ARGUMENTO       veredicto           envejece cada vuelta
//	entrevista.md   el RAZONAMIENTO    abiertas: 0         se relee después
//	evidencia.md    el HECHO           tags y links        se acumula
//
// Hasta el 2026-09-08 esto era un archivo y un `grep` de "http" sobre él. El
// grep funcionaba —cero links es cero links— y no se podía extender: "¿la
// entrevista terminó?" no se contesta grepeando prosa.
func Brief(raiz string) Resultado {
	var r Resultado

	var fm struct {
		Veredicto string `yaml:"veredicto"`
	}
	if !leerCabecera(raiz, docs.Brief, &fm, &r) {
		return r
	}

	// Los tres veredictos salen del ⑥. "no-lo-hagas" es tan válido como los
	// otros dos: el valor del brief es poder decir que no.
	validos := []string{"hacelo", "pivotea", "no-lo-hagas"}
	if !slices.Contains(validos, fm.Veredicto) {
		r.falla("el brief no trae veredicto (%s)", strings.Join(validos, " | "))
	} else {
		r.ok("el brief propone un veredicto: %s", fm.Veredicto)
		r.mide("veredicto", fm.Veredicto)
	}

	revisarEntrevista(raiz, &r)
	revisarEvidencia(raiz, &r)

	// Va SIEMPRE, pase o no pase, y no es una disculpa: es el límite del
	// diseño. Una compuerta no sale a la red y no llama a un modelo (R3), así
	// que puede contar links y no puede visitarlos. Decirlo en cada acta es lo
	// que evita que "13 links" se lea como "13 fuentes verificadas".
	r.noSeSabe("si las fuentes citadas existen de verdad: la compuerta las CUENTA, no las visita")

	return r
}

// revisarEntrevista mira el registro del ①–⑤.
//
// ────────────────────────────────────────────────────────────────────────────
// `abiertas` AUSENTE NO ES `abiertas: 0`
// ────────────────────────────────────────────────────────────────────────────
//
// Por eso es un puntero. Con un `int` pelado, un modelo que no escribe el campo
// obtiene un cero, y el cero es justo el valor que deja pasar. O sea: olvidarse
// del campo sería MÁS FÁCIL que cerrar la entrevista, y la compuerta estaría
// premiando el olvido.
//
// Es la misma regla que `lanzar/resumen.go`: lo que no está, no está.
func revisarEntrevista(raiz string, r *Resultado) {
	var fm struct {
		Rondas    *int `yaml:"rondas"`
		Preguntas *int `yaml:"preguntas"`
		Abiertas  *int `yaml:"abiertas"`
	}
	if !leerCabecera(raiz, docs.Entrevista, &fm, r) {
		return
	}

	if fm.Rondas != nil {
		r.mide("rondas", fmt.Sprint(*fm.Rondas))
	}
	if fm.Preguntas != nil {
		r.mide("preguntas", fmt.Sprint(*fm.Preguntas))
	}

	switch {
	case fm.Abiertas == nil:
		r.falla("%s no declara `abiertas` en el frontmatter, así que no se puede saber "+
			"si la entrevista terminó.\n"+
			"→ el ①–⑤ cierra cuando la frontera queda vacía: escribí `abiertas: 0`.\n"+
			"  Si quedaron ramas sin visitar, ponelas y NO sellés: un brief sobre huecos "+
			"es un PRD sobre huecos.", docs.Entrevista)
	case *fm.Abiertas > 0:
		r.mide("preguntas abiertas", fmt.Sprint(*fm.Abiertas))
		r.falla("la entrevista cerró con %d pregunta(s) abierta(s).\n"+
			"→ volvé al ①–⑤ y visitá esas ramas, o bajá el alcance del brief "+
			"hasta que no dependa de ellas.", *fm.Abiertas)
	default:
		r.mide("preguntas abiertas", "0")
		r.ok("la entrevista cerró con la frontera vacía")
	}
}

// revisarEvidencia mira lo que se encontró, y cuenta los tags y los links de
// verdad — los cuatro contadores, no sólo `links`.
//
// ────────────────────────────────────────────────────────────────────────────
// EL FRONTMATTER DECLARA; EL CUERPO ES EL HECHO
// ────────────────────────────────────────────────────────────────────────────
//
// `links: 13` es una línea que el modelo tipea. Trece "https://" en el cuerpo
// son trece bytes que están o no están. Lo mismo vale para los otros tres: un
// `[retrieved` en el cuerpo es un byte, `retrieved: 5` es una opinión sobre esos
// bytes. La compuerta frena sobre los segundos y AVISA cuando los primeros no
// coinciden — porque un número declarado que no cierra
// con lo que hay es exactamente la clase de cosa que alguien quiere ver antes de
// firmar, y no es motivo para frenar a nadie.
//
// # POR QUÉ EL UMBRAL ES UNO Y NO ES UN NÚMERO MÁGICO
//
// Cero es una categoría —"no investigó"—. Cualquier número mayor es un juicio
// sobre CUÁNTO alcanza, y el juicio es de Javier. Uno es el borde entre nada y
// algo, que es la única línea que una compuerta puede trazar sin opinar.
//
// MEDIDO EL 2026-09-05, y es la razón de que esto exista:
//
//	A (claude-code)  no-lo-hagas   17 retrieved · 1 model-prior · 13 links
//	B (nemotron)     hacelo         1 retrieved · 8 model-prior ·  0 links
//
// Veredictos OPUESTOS, y la compuerta aceptó los dos. Y el detalle que decide
// el diseño: B NO MINTIÓ. Marcó cada afirmación como `model-prior — unverified`,
// que es exactamente lo que el skill le pide. El skill hizo su trabajo. La que
// exigía de menos era la compuerta.
func revisarEvidencia(raiz string, r *Resultado) {
	var fm struct {
		Retrieved  *int   `yaml:"retrieved"`
		ModelPrior *int   `yaml:"model_prior"`
		Probado    *int   `yaml:"probado"`
		Links      *int   `yaml:"links"`
		Consultado *int   `yaml:"consultado"`
		SinAcceso  *int   `yaml:"sin_acceso"`
		Evidencia  string `yaml:"evidencia"`
	}
	cuerpo, err := frontmatter.DeArchivo(filepath.Join(raiz, docs.Evidencia), &fm)
	if err != nil {
		if os.IsNotExist(err) {
			r.falla("falta %s.\n"+
				"→ el ①–⑤ deja lo que encontró en su propio archivo, con procedencia por "+
				"afirmación: `retrieved` con link, `model-prior` sin verificar, `probado` "+
				"si lo construiste y lo viste.", docs.Evidencia)
		} else {
			r.falla("%s: %v", docs.Evidencia, err)
		}
		return
	}

	// LOS CUATRO CONTADORES SE MIDEN IGUAL, Y POR LA MISMA RAZÓN
	//
	// MEDIDO EL 2026-09-09, corrida B: el modelo declaró `retrieved: 5` con
	// SIETE tags `[retrieved]` en el cuerpo, y nadie chistó — porque hasta acá
	// `links` era el único que se contrastaba y los otros tres se repetían tal
	// como venían. Un número declarado que nadie mira es un número que se puede
	// inventar, que es justo lo que esta compuerta existe para no dejar pasar.
	//
	// Se cuenta por el ABRE-TAG y no por el tag entero: el skill escribe
	// `[model-prior — sin verificar]`, y el sufijo es prosa libre. El prefijo es
	// lo único estable.
	for _, m := range []struct {
		que       string
		abreTag   string
		declarado *int
	}{
		{"retrieved", "[retrieved", fm.Retrieved},
		{"model-prior", "[model-prior", fm.ModelPrior},
		{"probado", "[probado", fm.Probado},
		{"consultado", "[consultado", fm.Consultado},
		{"sin-acceso", "[sin-acceso", fm.SinAcceso},
	} {
		reales := bytes.Count(cuerpo, []byte(m.abreTag))
		r.mide(m.que+" en el cuerpo", fmt.Sprint(reales))
		if m.declarado != nil && *m.declarado != reales {
			r.avisa("%s declara `%s: %d` y en el cuerpo hay %d tag(s).",
				docs.Evidencia, m.que, *m.declarado, reales)
		}
	}

	reales := bytes.Count(cuerpo, []byte("http://")) + bytes.Count(cuerpo, []byte("https://"))
	r.mide("links en el cuerpo", fmt.Sprint(reales))
	if fm.Links != nil && *fm.Links != reales {
		r.avisa("%s declara `links: %d` y en el cuerpo hay %d.\n"+
			"→ `links` cuenta TODA url del cuerpo, no sólo las de competidores: "+
			"la compuerta cuenta bytes y no puede distinguirlas.", docs.Evidencia, *fm.Links, reales)
	}

	// EL ROSTER DE FUENTES — la procedencia por afirmación no dice qué NO se miró
	//
	// Los cuatro contadores de arriba miden lo que se encontró. Ninguno mide lo
	// que no se buscó: una corrida que sólo tocó el nivel 0 y una que además
	// intentó el nivel 1 y rebotó por falta de llave se ven IGUAL desde acá.
	//
	// El bloque `## Fuentes` lo arregla con una línea por nivel, y las dos
	// etiquetas son las dos respuestas posibles: `[consultado]` y `[sin-acceso]`.
	// Es la misma regla que el skill ya aplica a las afirmaciones —"lo que no
	// encontraste se escribe"— corrida un nivel para arriba: a las fuentes.
	//
	// AVISA Y NO FRENA, y no es blandura: ninguna `evidencia.md` escrita antes de
	// hoy tiene el bloque, y una compuerta nueva que frena sobre artefactos
	// viejos convierte en roja toda corrida archivada. Sube a falla cuando el
	// skill lleve una vuelta de uso y los artefactos del repo ya lo traigan.
	if bytes.Count(cuerpo, []byte("[consultado"))+bytes.Count(cuerpo, []byte("[sin-acceso")) == 0 {
		r.avisa("%s no tiene el roster `## Fuentes`: una línea por nivel, "+
			"`[consultado]` o `[sin-acceso]`.\n"+
			"→ sin él no se distingue \"el nivel 1 no tenía nada\" de \"al nivel 1 "+
			"no llegué\", y son cosas distintas.", docs.Evidencia)
	}

	// LA SALIDA, QUE YA ESTABA DISEÑADA
	//
	// El skill contempla la corrida degradada: sin herramientas de
	// investigación, con consentimiento explícito, todo queda `model-prior` y
	// la evidencia se marca de baja. Tiene que ser DELIBERADA — declarar que no
	// se investigó cuesta escribir una línea a propósito, que es justo lo que
	// se quiere que cueste.
	if fm.Evidencia == "baja" {
		r.avisa("la evidencia está declarada BAJA: el veredicto no se apoya en investigación " +
			"verificada.")
		r.noSeSabe("el panorama real: se declaró una corrida degradada y no se investigó")
		return
	}

	if reales == 0 {
		r.falla("%s no cita una sola fuente, y sellar un veredicto sin evidencia "+
			"es lo único que ningún paso posterior puede corregir.\n"+
			"→ empezá por el nivel 0, que no necesita ninguna llave: los registries "+
			"(npm, PyPI, crates.io), la API pública de GitHub y `curl`.\n"+
			"  Si de verdad no se pudo investigar, se declara: `evidencia: baja`.", docs.Evidencia)
		return
	}
	r.ok("la evidencia cita %d fuente(s) con link", reales)
}

// PRD comprueba que exista y tenga cuerpo.
func PRD(raiz string) Resultado {
	var r Resultado
	if _, err := os.Stat(filepath.Join(raiz, docs.PRD)); err != nil {
		r.falla("falta %s", docs.PRD)
	}
	return r
}

// Constitucion comprueba lo único sin lo cual sf queda ciego.
func Constitucion(raiz string) Resultado {
	var r Resultado

	var fm struct {
		TestCmd  string `yaml:"test_cmd"`
		Lenguaje string `yaml:"lenguaje"`
		Mutacion string `yaml:"mutacion"`
	}
	if !leerCabecera(raiz, docs.Constitucion, &fm, &r) {
		return r
	}

	// Sin test_cmd sf no puede correr los tests, y sin eso caen TRES compuertas:
	// el rojo del ⑲, el verde del ⑳ y el conteo del #8. Es la única línea del
	// frontmatter que se exige.
	if fm.TestCmd == "" {
		r.falla("la constitución no tiene `test_cmd:` — sin eso sf no puede correr los tests")
	}

	// `mutacion` vacío puede ser la respuesta correcta, así que esto AVISA y no
	// frena: si el stack no tiene herramienta, el ㉒ lo hace el modelo y está
	// bien. Lo que sf sí puede afirmar es el hecho de al lado —"para node
	// existe Stryker"—, y decirlo acá cuesta una línea y ahorra una feature
	// entera de mutantes hechos a mano que no se pueden comparar entre vueltas.
	if strings.TrimSpace(fm.Mutacion) == "" {
		if h := constitucion.HerramientaSugerida(fm.Lenguaje); h != "" {
			r.avisa("`mutacion:` está vacío y para %s existe %s. Vacío es válido, pero que sea una decisión.", fm.Lenguaje, h)
		}
	}
	return r
}

// Backlog cuenta que todas las historias tengan criterios con id Y CON TEXTO.
//
// Es la compuerta que HABILITA el mecanismo del ㉑: sin ids, el revisor
// contesta "anda" y no hay nada que contar después.
//
// Lo del texto no es un extra: el esqueleto de `sf new` deja la línea puesta y
// vacía —`- **CA-1** —`— así que contando sólo ids, un esqueleto sin completar
// pasaba la compuerta. Y un criterio que no dice nada es PEOR que ninguno: el
// ⑰ lo da por cubierto y el ㉑ le pone veredicto, o sea que el mecanismo entero
// queda en pie sobre algo que nadie puede juzgar.
//
// Sigue sin ser un juicio: no se mira si el texto es bueno, se mira si hay.
func Backlog(raiz string) Resultado {
	var r Resultado

	m, _ := filepath.Glob(filepath.Join(raiz, docs.Backlog, "us-*.md"))
	if len(m) == 0 {
		r.falla("no hay ninguna historia en %s", docs.Backlog)
		return r
	}

	for _, ruta := range m {
		id := strings.TrimSuffix(filepath.Base(ruta), ".md")
		h, err := historia.Leer(raiz, id)
		if err != nil {
			r.falla("%v", err)
			continue
		}
		if len(h.Criterios) == 0 {
			r.falla("%s no tiene criterios de aceptación (CA-1, CA-2, …)", id)
			continue
		}
		// El título NO se exige acá, y la diferencia importa: la compuerta
		// pregunta si el MECANISMO se sostiene, y lo que el ⑰ cuenta y el ㉑
		// juzga son los criterios. Un título flojo no rompe nada.
		//
		// De que la historia esté sin pinponear se ocupa el checkpoint del ⑨
		// (historia.SinPinponear), que sí lo mira: ahí la pregunta es otra
		// —"¿hay algo que completar?"— y ahí un título vacío es la señal.
		if len(h.SinTexto) > 0 {
			r.falla("%s tiene el id puesto y el criterio vacío: %s",
				id, strings.Join(h.SinTexto, " · "))
		}
	}
	return r
}

// Roadmap comprueba que se pueda leer y que cubra el backlog.
//
// El segundo chequeo es el que atrapa el error real: una historia que quedó
// fuera de todas las features no la va a implementar nadie, y nadie se entera.
func Roadmap(raiz string) Resultado {
	var r Resultado

	rm, err := roadmap.Leer(raiz)
	if err != nil {
		r.falla("%v", err)
		return r
	}

	enFeatures := map[string]bool{}
	for _, f := range rm.Features {
		for _, id := range f.Historias {
			enFeatures[id] = true
		}
		// Aviso, no freno: una feature de 8 historias tarda demasiado en dar la
		// vuelta, pero partirla o no es criterio, y el criterio es de Javier.
		if len(f.Historias) >= 6 {
			r.avisa("%s junta %d historias. ¿La partís?", f.ID, len(f.Historias))
		}
	}

	m, _ := filepath.Glob(filepath.Join(raiz, docs.Backlog, "us-*.md"))
	for _, ruta := range m {
		id := strings.TrimSuffix(filepath.Base(ruta), ".md")
		if !enFeatures[id] {
			r.falla("%s no está en ninguna feature del roadmap", id)
		}
	}
	return r
}

// ────────────────────────────────────────────────────────────────────────────
// planificacion — todo es contar. Nada es juzgar.
// ────────────────────────────────────────────────────────────────────────────

// reOpcion caza los títulos de opción de decision.md.
//
//	## A — un pool de workers
//	## B — …  ✅
//
// Se exigen TRES, no "varias": tres es el número del ⑫, y contar hasta tres es
// exactamente el tipo de comprobación que sf puede hacer sin opinar.
var reOpcion = regexp.MustCompile(`(?m)^##\s+([A-Z])\s`)

// Planificacion corre las cinco compuertas de salida del bloque ⑫–⑯.
func Planificacion(raiz string, f roadmap.Feature) Resultado {
	var r Resultado
	carpeta := f.Carpeta()

	// ① los tres archivos
	for _, n := range []string{docs.Decision, docs.Spec, tareas.Archivo} {
		if _, err := os.Stat(filepath.Join(raiz, carpeta, n)); err != nil {
			r.falla("falta %s", filepath.Join(carpeta, n))
		}
	}
	if !r.Pasa() {
		// Sin los archivos no tiene sentido seguir contando adentro: los errores
		// que salieran serían consecuencia del primero y taparían la causa.
		return r
	}

	// ② decision.md tiene TRES opciones
	b, err := os.ReadFile(filepath.Join(raiz, carpeta, docs.Decision))
	if err != nil {
		r.falla("%v", err)
	} else {
		if n := len(reOpcion.FindAll(b, -1)); n != 3 {
			r.falla("decision.md tiene %d opciones y el ⑫ pide 3", n)
		}
		// ② bis — la vara, y el puntaje contra la vara
		revisarVara(b, &r)
	}

	p, err := tareas.Leer(raiz, carpeta)
	if err != nil {
		r.falla("%v", err)
		return r
	}

	// ③ cada lote tiene al menos un test
	//
	// Esta es la que hace posible el dolor #8: sin la lista de tests, `sf lote
	// start` no tiene contra qué exigir el rojo.
	if len(p.Lotes()) == 0 {
		r.falla("tareas.json no tiene ninguna tarea")
	}
	for _, l := range p.Lotes() {
		if len(p.TestsDelLote(l)) == 0 {
			r.falla("el lote %d no tiene ningún test planificado", l)
		}
	}

	// ④ cada tarea nombra al menos un test por criterio que promete
	//
	// La ③ pide UN test por lote, y un piso de uno se vuelve techo de uno: la
	// tarea nombra el camino feliz y todo lo demás aparece tres vueltas de
	// revisión después, cuando el ㉒ lo encuentra. Contra eso alcanza con
	// contar, que es lo único que una máquina puede hacer acá.
	//
	// NO dice nada sobre la calidad del test —sf no lo puede leer—, y no
	// pretende: dice que una tarea que promete cuatro criterios y nombra un
	// test nace corta, y lo dice ANTES de que se escriba una línea de código,
	// igual que la ⑤.
	//
	// Se cuenta por TAREA y no por lote a propósito. Por lote el mensaje sería
	// "al lote 2 le faltan tests" y habría que ir a buscar cuál; por tarea el
	// que lo lee ya sabe dónde tocar. Y el lote no es la unidad honesta acá:
	// una tarea generosa taparía a una tacaña de al lado.
	//
	// Si un test cubre de verdad dos criterios, el plan nombra dos y lo parte.
	// Un test que prueba dos cosas tiene dos motivos para ponerse rojo, y
	// cuando se pone no dice cuál de los dos fue.
	for _, t := range p.Tareas {
		criterios := len(unicos(t.Satisface))
		tests := len(unicos(t.Tests))
		if criterios > 0 && tests < criterios {
			r.falla("la tarea %s: %d criterios y %d tests — falta al menos un test por criterio",
				t.ID, criterios, tests)
		}
	}

	// ④ bis — la ruta del test es la del TEST, no la del sujeto
	//
	// Sale de la primera corrida real: el plan nombró los tests dentro del `.sql`
	// que probaban, `sf lote start` los reportó todos como faltantes, y la
	// compuerta del ⑰ no lo había mirado porque cuenta tests pero no mira DÓNDE
	// dicen que van a estar. Ver rutas_de_test.go.
	{
		conTests := make([]tareaConTests, 0, len(p.Tareas))
		for _, t := range p.Tareas {
			conTests = append(conTests, tareaConTests{ID: t.ID, Tests: t.Tests})
		}
		revisarRutasDeTest(raiz, conTests, &r)
	}

	// ⑤ cada criterio de las historias está cubierto por alguna tarea
	//
	// Acá empieza a morir el #7, y ANTES de escribir una línea de código: si
	// una historia tiene 9 criterios y las tareas cubren 7, la feature ya nace
	// a medias.
	deLasHistorias, err := historia.Criterios(raiz, f.Historias)
	if err != nil {
		r.falla("%v", err)
		return r
	}
	cubiertos := p.CriteriosCubiertos()

	var sinCubrir []string
	for _, c := range deLasHistorias {
		if !slices.Contains(cubiertos, c) {
			sinCubrir = append(sinCubrir, c)
		}
	}
	if len(sinCubrir) > 0 {
		r.falla("la feature tiene %d criterios y las tareas cubren %d. Sin cubrir: %s",
			len(deLasHistorias), len(deLasHistorias)-len(sinCubrir), strings.Join(sinCubrir, " · "))
	}

	// ⑥ tareas que apuntan a criterios que no existen
	//
	// El espejo de la anterior, y atrapa el error opuesto: una tarea que dice
	// satisfacer us-7/CA-9 cuando us-7 tiene tres criterios está mintiendo, y
	// el conteo de arriba no lo vería.
	for _, c := range cubiertos {
		if !slices.Contains(deLasHistorias, c) {
			r.falla("una tarea dice satisfacer %s, y ese criterio no existe", c)
		}
	}

	return r
}

// reVara caza las filas de la vara: `| V-1 | … |`.
var reVara = regexp.MustCompile(`(?m)^\|\s*(V-\d+)\s*\|`)

// rePuntaje caza las filas del puntaje: la primera celda es A, B o C.
//
// La cabecera (`| | V-1 | …`) no matchea porque su primera celda está vacía, y
// el separador (`|---|---|`) tampoco. Se apoya en eso a propósito: contar una
// tabla con un regex es frágil, y las dos filas que podrían confundirla son
// justo las dos que no empiezan con una letra sola.
var rePuntaje = regexp.MustCompile(`(?m)^\|\s*([A-C])\s*\|`)

// revisarVara cuenta la vara y el puntaje de decision.md.
//
// ────────────────────────────────────────────────────────────────────────────
// CONTAR TRES ENCABEZADOS ATACA EL SÍNTOMA GROSERO, NO EL QUE QUEDA
// ────────────────────────────────────────────────────────────────────────────
//
// La compuerta ② exige tres opciones, y mata la corazonada escrita como si
// fuera una comparación. Lo que NO ve es el fallo que pasa sin despeinarse:
//
//	el modelo ya decidió B mientras leía el us-#
//	  → escribe B bien
//	  → escribe A y C como espantapájaros, creíbles y peores
//	  → tres encabezados ✓ · el argumento que inclinó ✓ · compuerta verde
//
// El antídoto es de `arena`, y NO es el fan-out: es que la vara se escriba
// ANTES de ver las opciones. Es el mismo principio del ciego del banco, aplicado
// a una decisión de diseño en vez de a un experimento.
//
// ────────────────────────────────────────────────────────────────────────────
// Y EL LÍMITE, DICHO ACÁ PARA QUE NADIE LO VENDA DE MÁS
// ────────────────────────────────────────────────────────────────────────────
//
// Una vara escrita por el mismo modelo que después elige NO es un ciego de
// verdad: nada le impide escribirla ya sabiendo la respuesta. Lo que sí hace es
// dejar el fraude POR ESCRITO y versionado — una vara que sólo premia lo que
// hace B queda en decision.md, y el ⑰ es un humano leyendo. Hoy no hay ni eso.
func revisarVara(b []byte, r *Resultado) {
	criterios := reVara.FindAllSubmatch(b, -1)
	vistos := map[string]bool{}
	for _, m := range criterios {
		vistos[string(m[1])] = true
	}

	switch n := len(vistos); {
	case n == 0:
		r.falla("decision.md no tiene la sección `## Vara`.\n" +
			"→ de 3 a 6 criterios `V-#`, escritos ANTES de la primera opción. " +
			"Una vara escrita después describe la opción que ya elegiste.")
	case n < 3:
		r.falla("la vara tiene %d criterio(s) y el ⑫ pide entre 3 y 6", n)
	case n > 6:
		r.falla("la vara tiene %d criterios: más de 6 no es más rigor, es una "+
			"lista donde cada opción gana en algo", n)
	}

	if filas := len(rePuntaje.FindAll(b, -1)); filas != 3 {
		r.falla("decision.md declara 3 opciones y el `## El puntaje` tiene %d fila(s).\n"+
			"→ una fila por opción, criterio por criterio. Puntuar de a dos es "+
			"elegir y después justificar.", filas)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// revision — sf cuenta, no opina
// ────────────────────────────────────────────────────────────────────────────

// Revision es donde muere el dolor #7, y el mecanismo es contar.
//
// sf no juzga la revisión: comprueba que HAYA OCURRIDO SOBRE TODOS. Si la
// feature tiene 9 criterios y el informe opina sobre 7, no avanza — y dice
// cuáles faltan.
func Revision(raiz string, f roadmap.Feature) Resultado {
	var r Resultado
	ruta := filepath.Join(raiz, f.Carpeta(), docs.Revision)

	rev, err := revision.Leer(ruta)
	if err != nil {
		r.falla("%v", err)
		return r
	}

	deLasHistorias, err := historia.Criterios(raiz, f.Historias)
	if err != nil {
		r.falla("%v", err)
		return r
	}

	var sinVeredicto []string
	for _, c := range deLasHistorias {
		if _, hay := rev.Criterios[c]; !hay {
			sinVeredicto = append(sinVeredicto, c)
		}
	}
	if len(sinVeredicto) > 0 {
		r.falla("la feature tiene %d criterios y el informe opina sobre %d. Sin veredicto: %s",
			len(deLasHistorias), len(deLasHistorias)-len(sinVeredicto), strings.Join(sinVeredicto, " · "))
	}

	// ────────────────────────────────────────────────────────────────────────
	// LA ESCALERA — "cumple" tiene que decir CÓMO SE SABE
	// ────────────────────────────────────────────────────────────────────────
	//
	// Las dos comprobaciones de abajo son aritmética, igual que la de arriba.
	// La primera cuenta que el escalón esté declarado; la segunda comprueba que
	// un escalón que dice "lo corrí" tenga algo que exista detrás.
	//
	// LA SEGUNDA ES LA QUE ATRAPA LA MENTIRA, y no es una opinión: es el mismo
	// truco que ya usa `sf audit` en su pregunta ④ —"¿los tests que lo probaban
	// SIGUEN EXISTIENDO?"—, con la misma función. Decir "escalón 4" cuesta dos
	// bytes; que el test exista, no.
	revisarEscalera(raiz, rev, deLasHistorias, &r)

	// Los hallazgos abiertos mandan de vuelta a implementar. Con uno solo no
	// avanza — y `descartado` sí pasa, porque es una decisión de Javier (R3).
	var abiertos []string
	for _, h := range rev.Hallazgos {
		if h.Estado == revision.Abierto {
			abiertos = append(abiertos, h.ID)
		}
	}
	if len(abiertos) > 0 {
		r.falla("hay %d hallazgos abiertos: %s", len(abiertos), strings.Join(abiertos, " · "))
	}

	// Un mutante que murió en una vuelta y vive en ésta es una regresión de
	// cobertura: había un test que lo agarraba y ya no. Reportarlo y no abrir
	// nada es la única forma de que se pierda, así que se cuenta.
	//
	// No juzga la revisión —sf no sabe si el hallazgo es bueno— y por eso pide
	// UNO y no uno por resucitado: dos resurrecciones pueden ser el mismo test
	// aflojado, y exigir dos hallazgos fabricaría el segundo.
	if rev.Mutantes.Propios.Resucitados > 0 && !hayHallazgoDelVeintidos(rev) {
		r.falla("%d mutantes resucitaron —murieron antes y viven ahora— y no hay ningún hallazgo del ㉒",
			rev.Mutantes.Propios.Resucitados)
	}

	// Y `vuelta:` es el mismo truco que `intentos_fallidos`: si el ㉑ va por la
	// cuarta, eso no es ruido — es que la planificación se quedó corta.
	if rev.Vuelta >= 3 {
		r.avisa("el ㉑ va por la vuelta %d. La planificación se quedó corta.", rev.Vuelta)
	}

	return r
}

// revisarEscalera comprueba que cada veredicto diga cómo se sabe.
//
// ────────────────────────────────────────────────────────────────────────────
// TRES REGLAS, Y NINGUNA OPINA
// ────────────────────────────────────────────────────────────────────────────
//
//	① todo criterio declara `escalon` ≥ 1        cuenta que el campo esté
//	② un `cumple` en escalón ≥ 4 nombra una       comprueba que el archivo
//	  `prueba` QUE EXISTE                         esté en el repo
//	③ un `cumple` por debajo del piso de la       compara dos números
//	  constitución es un hallazgo
//
// LA ① NO CORRE SOBRE LO ARCHIVADO. Un criterio en `SinEscalera` es una
// revisión escrita antes del 2026-09-11, no una afirmación sin respaldo, y
// frenar sobre artefactos viejos volvería roja toda feature cerrada. Se
// distingue por el 0, que el Unmarshal reserva justamente para eso.
//
// LA ③ SÓLO CORRE SI LA CONSTITUCIÓN DECLARÓ UN PISO (R1): cuál es el escalón
// aceptable depende del proyecto, y sf no lo elige. Sin `escalon_minimo`, la
// escalera informa y no frena.
func revisarEscalera(raiz string, rev *revision.Revision, criterios []string, r *Resultado) {
	piso := 0
	if c, err := constitucion.Leer(raiz); err == nil {
		piso = c.Verificacion.EscalonMinimo
	}

	var sinEscalon, sinPrueba, bajoElPiso []string

	for _, c := range criterios {
		v, hay := rev.Criterios[c]
		if !hay {
			continue // ya lo reportó el conteo de veredictos, y decirlo dos veces tapa
		}

		// Un criterio del formato viejo no entra a ninguna de las tres: la
		// escalera no es retroactiva.
		if v.Escalon == revision.SinEscalera {
			if !esArchivada(raiz, rev) {
				sinEscalon = append(sinEscalon, c)
			}
			continue
		}

		if !v.Cumple() {
			continue // un `no-cumple` ya es un hallazgo; no le pedimos respaldo
		}

		if v.Escalon >= revision.EscalonCorrido {
			if v.Prueba == "" {
				sinPrueba = append(sinPrueba, c+" (escalón "+fmt.Sprint(v.Escalon)+", sin `prueba`)")
			} else if faltan := suite.Faltantes(raiz, []string{v.Prueba}); len(faltan) > 0 {
				sinPrueba = append(sinPrueba, c+" → "+faltan[0])
			}
		}

		if piso > 0 && v.Escalon < piso {
			bajoElPiso = append(bajoElPiso, fmt.Sprintf("%s (escalón %d)", c, v.Escalon))
		}
	}

	if len(sinEscalon) > 0 {
		r.falla("%d criterio(s) no declaran `escalon` — no se sabe cómo lo sabés: %s\n"+
			"→ 1 lo dijiste · 2 señalaste la línea · 3 mostraste que el caso malo no llega · "+
			"4 lo corriste · 5 lo reprodujiste en la app.",
			len(sinEscalon), strings.Join(sinEscalon, " · "))
	}
	if len(sinPrueba) > 0 {
		r.falla("%d criterio(s) dicen haberse corrido y su `prueba` no existe: %s\n"+
			"→ del escalón 4 para arriba, `prueba` apunta a algo que está en el repo.",
			len(sinPrueba), strings.Join(sinPrueba, " · "))
	}
	if len(bajoElPiso) > 0 {
		r.falla("la constitución pide escalón %d y %d criterio(s) cumplen por debajo: %s\n"+
			"→ o se sube la evidencia, o el criterio es un hallazgo.",
			piso, len(bajoElPiso), strings.Join(bajoElPiso, " · "))
	}
}

// esArchivada dice si esta revisión ya está en `.docs/archivado/`.
//
// Es la única forma de distinguir "revisión vieja, la escalera no existía" de
// "revisión nueva que no declaró el escalón", y las dos se ven igual desde el
// JSON: las dos tienen escalón 0.
func esArchivada(raiz string, rev *revision.Revision) bool {
	_, err := os.Stat(filepath.Join(raiz, docs.Archivado))
	if err != nil {
		return false
	}
	entradas, err := os.ReadDir(filepath.Join(raiz, docs.Archivado))
	if err != nil {
		return false
	}
	for _, e := range entradas {
		if strings.HasPrefix(e.Name(), rev.Feature+"-") || e.Name() == rev.Feature {
			return true
		}
	}
	return false
}

// hayHallazgoDelVeintidos dice si el informe abrió algo desde los mutantes.
//
// Cuenta los descartados también: `descartado` es una decisión de Javier con
// motivo escrito (R3), y eso no es lo mismo que no haber mirado.
func hayHallazgoDelVeintidos(rev *revision.Revision) bool {
	for _, h := range rev.Hallazgos {
		if h.Origen == 22 {
			return true
		}
	}
	return false
}

// ────────────────────────────────────────────────────────────────────────────
// cierre
// ────────────────────────────────────────────────────────────────────────────

// Cierre comprueba los dos archivos del ㉓.
//
// Son dos y no uno: la doc Y el journal. El journal entró tarde al diseño —ya
// estaba construido y ningún estado lo había reclamado— y lleva la memoria que
// el ⑫ de una feature futura va a leer.
//
// Busca en la carpeta viva y, si no está, en la archivada. La pregunta es "¿el
// ㉓ produjo sus dos archivos?", no "¿dónde está la carpeta hoy?" — y atarla al
// lugar hace que la compuerta falle cuando alguien está terminando un archivado
// que quedó a medias, que es justo cuando más se la necesita.
func Cierre(raiz string, f roadmap.Feature) Resultado {
	var r Resultado

	carpeta := f.Carpeta()
	if _, err := os.Stat(filepath.Join(raiz, carpeta)); os.IsNotExist(err) {
		if archivada := filepath.Join(docs.Archivado, filepath.Base(carpeta)); existe(raiz, archivada) {
			carpeta = archivada
		}
	}

	for _, n := range []string{docs.Doc, docs.Journal} {
		if _, err := os.Stat(filepath.Join(raiz, carpeta, n)); err != nil {
			r.falla("falta %s", filepath.Join(carpeta, n))
		}
	}
	return r
}

// unicos saca los repetidos sin ordenar.
//
// Sin esto, una tarea que nombra el mismo test dos veces pasaría la ④ contando
// dos, que es exactamente la forma de cumplir una regla de contar sin cumplir
// la regla.
func unicos(xs []string) []string {
	var r []string
	for _, x := range xs {
		if !slices.Contains(r, x) {
			r = append(r, x)
		}
	}
	return r
}

func existe(raiz, rel string) bool {
	_, err := os.Stat(filepath.Join(raiz, rel))
	return err == nil
}

// ────────────────────────────────────────────────────────────────────────────
// implementar — dos preguntas, y la primera parecía no hacer falta
// ────────────────────────────────────────────────────────────────────────────

// Implementar comprueba lo que se puede sin haber visto el rojo.
//
// La compuerta del medio —"corré los tests, TIENEN QUE FALLAR"— es de `sf lote
// start`. Acá se cierran las dos puntas que esa compuerta deja abiertas:
//
//	¿empezó?     f.Lotes vacío = nadie corrió `sf lote start`
//	¿vio rojo?   sin `rojo` en el lote en curso, NO AVANZA
//
// ────────────────────────────────────────────────────────────────────────────
// LA PRIMERA FALTABA, Y ERA EL AGUJERO MÁS GRANDE DEL BINARIO
// ────────────────────────────────────────────────────────────────────────────
//
// LoteActual() devuelve false tanto para "no hay lotes" como para "todos
// commiteados", y esta función devolvía "pasa" en los dos casos. Con el plan
// recién aprobado f.Lotes está vacío —los lotes se siembran en `sf lote
// start`—, así que UN SOLO `sf done` movía la feature a revisión sin branch,
// sin tests, sin código y sin commit. Todo el mecanismo del producto se
// evitaba corriendo un comando una vez.
//
// El arreglo es contar (R3): una lista vacía no es una lista terminada.
func Implementar(f *estado.Feature) Resultado {
	var r Resultado

	if f.SinSembrar() {
		r.falla("no empezaste ningún lote. Corré `sf lote start` antes de implementar")
		return r
	}

	l, hay := f.LoteActual()
	if !hay {
		return r // todos commiteados: el estado puede cerrarse
	}

	if !l.Rojo {
		r.falla("no vi el rojo del lote %d. Corré `sf lote start` antes de implementar", l.Lote)
	}
	return r
}

// ────────────────────────────────────────────────────────────────────────────
// Ayudantes
// ────────────────────────────────────────────────────────────────────────────

// leerCabecera lee el frontmatter de un artefacto y anota la falla si no puede.
//
// Devuelve si se pudo, para que quien llama corte antes de seguir chequeando
// campos de algo que no existe.
func leerCabecera(raiz, rel string, destino any, r *Resultado) bool {
	if _, err := frontmatter.DeArchivo(filepath.Join(raiz, rel), destino); err != nil {
		if os.IsNotExist(err) {
			r.falla("falta %s", rel)
		} else {
			r.falla("%s: %v", rel, err)
		}
		return false
	}
	return true
}
