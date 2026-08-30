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
	"github.com/jdbaigorria/specforge/sf/internal/tareas"
)

// Resultado es el veredicto de correr las compuertas de un estado.
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
}

// Pasa dice si el estado puede moverse.
func (r Resultado) Pasa() bool { return len(r.Fallas) == 0 }

func (r *Resultado) falla(formato string, args ...any) {
	r.Fallas = append(r.Fallas, fmt.Sprintf(formato, args...))
}

func (r *Resultado) avisa(formato string, args ...any) {
	r.Avisos = append(r.Avisos, fmt.Sprintf(formato, args...))
}

// Texto arma el veredicto para que lo lea el que trabajó.
//
// Los ✓ no se listan: sólo importa lo que falta. Un veredicto que enumera todo
// lo que salió bien es ruido, y el que lo lee tiene que buscar la ✗ entre
// quince ✓.
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

// ────────────────────────────────────────────────────────────────────────────
// Los cinco de producto
// ────────────────────────────────────────────────────────────────────────────

// Brief comprueba que el ⑥ tenga qué sellar.
//
// No comprueba que el brief sea BUENO —eso es juicio y es de Javier— sino que
// exista y traiga un veredicto. Es la diferencia entre una compuerta y un juez.
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
	}
	return r
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
	} else if n := len(reOpcion.FindAll(b, -1)); n != 3 {
		r.falla("decision.md tiene %d opciones y el ⑫ pide 3", n)
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
