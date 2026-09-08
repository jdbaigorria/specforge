//go:build e2e

// El guion de humo: la máquina entera, por la superficie del binario.
//
// ────────────────────────────────────────────────────────────────────────────
// POR QUÉ ESTE ARCHIVO EXISTE, Y POR QUÉ NINGÚN TEST DE PAQUETE LO REEMPLAZA
// ────────────────────────────────────────────────────────────────────────────
//
// La auditoría de specs/arreglos.md encontró diez defectos con la suite en
// verde — 213 tests, todos pasando. Y no fue mala suerte:
//
//	Ninguno de los diez se ve desde adentro de un paquete.
//	Todos viven en la COSTURA ENTRE COMANDOS.
//
// `sf next` decía "implementar" y `sf lote start` contestaba "esto está en
// planificacion". `sf approve` sellaba lo que `sf done` rechazaba. `sf done`
// movía a revisión una feature en la que nadie escribió una línea. Cada
// comando, mirado solo, estaba bien.
//
// Por eso acá se compila el binario y se lo corre de verdad, contra un proyecto
// de juguete en disco, con git de verdad. Lo que se afirma es el EXIT CODE, que
// ya es parte de la interfaz (ver los cuatro de main.go): un CLI que siempre
// devuelve 0 obliga a parsear texto, y acá el que lee suele ser un agente.
//
// ────────────────────────────────────────────────────────────────────────────
// CADA PASO NOMBRA EL HALLAZGO QUE CUBRE
// ────────────────────────────────────────────────────────────────────────────
//
// Los pasos marcados con A#—el número del hallazgo en specs/arreglos.md— no
// están para que la vuelta sea completa: están porque ESE PASO FALLABA. Si
// alguno se cae, buscá el hallazgo en el documento antes de tocar el test.
//
// Corre aparte de la suite normal porque tarda —compila y corre git— y porque
// un `go test ./...` que necesita git instalado deja de correr en cualquier
// lado:
//
//	go test -tags e2e ./cmd/sf/
package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/comandos"
	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// Los cuatro exit codes son la interfaz. Se repiten acá con su nombre para que
// las aserciones se lean como el flujo y no como una tabla de números.
const (
	hayTrabajo  = 0
	esError     = 1
	esParada    = 2
	noQuedaNada = 3
)

// ────────────────────────────────────────────────────────────────────────────
// El andamio
// ────────────────────────────────────────────────────────────────────────────

type proyecto struct {
	t       *testing.T
	raiz    string
	bin     string
	home    string // SPECFORGE_HOME, para no tocar el home de verdad
	harness string // en cuál está parado: lo lee sf por SPECFORGE_HARNESS
}

// nuevoProyecto compila el binario y arma un repo de juguete con un commit.
func nuevoProyecto(t *testing.T) *proyecto {
	t.Helper()

	dir := t.TempDir()
	p := &proyecto{
		t:    t,
		raiz: filepath.Join(dir, "proyecto"),
		bin:  filepath.Join(dir, "sf"),
		home: filepath.Join(dir, "home"),
	}
	if err := os.MkdirAll(p.raiz, 0o755); err != nil {
		t.Fatal(err)
	}

	// El binario se compila desde el código de ESTA corrida. Usar uno del PATH
	// probaría el que está instalado, que es justo lo que no queremos saber.
	compilar := exec.Command("go", "build", "-o", p.bin, ".")
	if out, err := compilar.CombinedOutput(); err != nil {
		t.Fatalf("no pude compilar sf: %v\n%s", err, out)
	}

	// `-c` en vez de `config` global: los tests no pueden tocar la
	// configuración de quien los corra, y git exige user.name/email.
	p.git("init", "-b", "main")
	p.git("config", "user.email", "e2e@ejemplo")
	p.git("config", "user.name", "E2E")
	p.escribir("go.mod", "module juguete\n\ngo 1.22\n")
	p.git("add", "-A")
	p.git("commit", "-m", "init")

	// Todos los guiones arrancan parados en claude-code salvo que digan otra
	// cosa. Fijarlo importa: sin esto heredarían el arnés de quien corre la
	// suite, y el guion probaría algo distinto en la máquina de cada uno.
	p.harness = "claude-code"
	return p
}

// conPerfiles declara los dos perfiles que la máquina pide.
//
// ────────────────────────────────────────────────────────────────────────────
// ESTO ES EL COSTO DE H2, Y ESTÁ ACÁ PARA QUE SE VEA
// ────────────────────────────────────────────────────────────────────────────
//
// Antes el e2e no lo necesitaba, porque `sf install` sembraba tres modelos con
// nombres de Anthropic. Ya no siembra ninguno —para nadie, tampoco para Claude
// Code— así que el primer `sf next` de un proyecto nuevo PARA y los pide.
//
// Que el guion de humo tenga que hacer esto no es un defecto del guion: es el
// paso real que Javier va a dar una vez por harness, y tenerlo escrito acá es la
// única forma de que un cambio que lo rompa se note.
//
// Los alias son genéricos a propósito: si acá dijera un nombre de proveedor, el
// chequeo de CI que prohíbe justamente eso lo cazaría — y tendría razón.
func (p *proyecto) conPerfiles() *proyecto {
	p.t.Helper()
	for _, perfil := range []string{"razonar", "construir"} {
		if c, out := p.sf("model", perfil, "--alias", "a-"+perfil,
			"--id", "prov/"+perfil, "--via", "subagente"); c != 0 {
			p.t.Fatalf("sf model %s → exit %d\n%s", perfil, c, out)
		}
	}
	return p
}

func (p *proyecto) git(args ...string) {
	p.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = p.raiz
	if out, err := cmd.CombinedOutput(); err != nil {
		p.t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func (p *proyecto) escribir(rel, texto string) {
	p.t.Helper()
	ruta := filepath.Join(p.raiz, rel)
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		p.t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(texto), 0o644); err != nil {
		p.t.Fatal(err)
	}
}

// sf corre el binario y devuelve el exit code y la salida junta.
//
// stdout y stderr van juntos a propósito: el que lee esto es un agente que ve
// una sola corriente, y separar acá probaría algo que nadie ve.
// enArnes cambia en cuál está parado el proyecto de prueba.
//
// Es lo que hace posible probar SECCIONAR: el mismo repo, comandos corridos
// desde arneses distintos, y `sf` resolviendo los modelos del que está parado.
func (p *proyecto) enArnes(harness string) *proyecto {
	p.harness = harness
	return p
}

func (p *proyecto) sf(args ...string) (int, string) {
	p.t.Helper()

	cmd := exec.Command(p.bin, args...)
	cmd.Dir = p.raiz
	// SPECFORGE_HOME existe justamente para esto: `sf install` escribe el mapa
	// de modelos, y un test que toca ~/.specforge/ del que lo corre está mal.
	env := append(os.Environ(), "SPECFORGE_HOME="+p.home)
	// Las de detección se limpian SIEMPRE: la suite corre adentro de un arnés y
	// sin esto el binario bajo prueba heredaría el de quien la lanzó.
	for _, v := range global.VarsDeHarness {
		env = append(env, v+"=")
	}
	if p.harness != "" {
		env = append(env, global.VarHarness+"="+p.harness)
	}
	cmd.Env = env

	out, err := cmd.CombinedOutput()
	if err != nil {
		var ee *exec.ExitError
		if !asExitError(err, &ee) {
			p.t.Fatalf("sf %v: %v\n%s", args, err, out)
		}
		return ee.ExitCode(), string(out)
	}
	return 0, string(out)
}

func asExitError(err error, dst **exec.ExitError) bool {
	ee, ok := err.(*exec.ExitError)
	if ok {
		*dst = ee
	}
	return ok
}

// paso corre un comando y exige el exit code. `porque` es lo que se está
// probando, y aparece en la falla — un "quería 2, dio 0" sin contexto obliga a
// contar líneas para saber en qué parte de la vuelta estabas.
func (p *proyecto) paso(porque string, quiero int, args ...string) string {
	p.t.Helper()
	code, out := p.sf(args...)
	if code != quiero {
		p.t.Fatalf("%s\n  sf %s → exit %d, quería %d\n%s",
			porque, strings.Join(args, " "), code, quiero, sangrar(out))
	}
	return out
}

// dice exige que la salida mencione un pedazo. Se usa donde el exit code no
// alcanza: dos fallas distintas dan las dos exit 2, y la diferencia entre
// atrapar el bug correcto y atrapar otro está en el texto.
func (p *proyecto) dice(out, pedazo, porque string) {
	p.t.Helper()
	if !strings.Contains(out, pedazo) {
		p.t.Fatalf("%s\n  la salida no menciona %q:\n%s", porque, pedazo, sangrar(out))
	}
}

func (p *proyecto) noDice(out, pedazo, porque string) {
	p.t.Helper()
	if strings.Contains(out, pedazo) {
		p.t.Fatalf("%s\n  la salida menciona %q y no debería:\n%s", porque, pedazo, sangrar(out))
	}
}

func sangrar(s string) string {
	return "    " + strings.ReplaceAll(strings.TrimRight(s, "\n"), "\n", "\n    ")
}

// ────────────────────────────────────────────────────────────────────────────
// Vuelta A — una historia normal, de la idea al archivado
// ────────────────────────────────────────────────────────────────────────────

func TestVueltaCompletaDeUnaHistoria(t *testing.T) {
	p := nuevoProyecto(t)

	// ── El andamio ──────────────────────────────────────────────────────────
	p.paso("install tiene que dejar el orquestador y el mapa de modelos",
		hayTrabajo, "install")
	p.paso("init arma .docs/ y detecta el stack", hayTrabajo, "init")

	// Los perfiles se declaran una vez por harness. Antes de H2 esto no hacía
	// falta porque `sf install` sembraba tres modelos con nombres de Anthropic;
	// ahora no siembra ninguno y el primer `sf next` de trabajo los pide.
	p.conPerfiles()

	out := p.paso("el primer next arranca en el brief, y es el único via: vos",
		hayTrabajo, "next")
	p.dice(out, "sfp-scout", "el ⑥ lo hace sfp-scout")
	p.dice(out, "vos", "el brief es un pinponeo: un subagente no te habla")

	// ── ⑥ el brief ──────────────────────────────────────────────────────────
	p.brief()
	p.paso("el brief está y lo sella Javier", hayTrabajo, "done")
	p.paso("hasta que no lo selle, no se pasa", esParada, "next")
	p.paso("approve copia el veredicto que dice el archivo", hayTrabajo, "approve")

	// ── ⑦ el PRD ────────────────────────────────────────────────────────────
	p.escribir(".docs/prd.md", "# PRD\nSumar dos números.\n")
	p.paso("del ⑦ se pasa derecho al ⑧", hayTrabajo, "done")

	// ── ⑧ la constitución ───────────────────────────────────────────────────
	//
	// A4: acá el ⑧ era INALCANZABLE. `sf init` siempre crea constitucion.md,
	// y el checkpoint miraba si existía — así que se saltaba a la 🛑 y lo que
	// se sellaba era la plantilla. `sfp-constitucion` no lo invocaba nadie.
	out = p.paso("A4 · el ⑧ tiene que pedir trabajo, no la 🛑", hayTrabajo, "next")
	p.dice(out, "sfp-constitucion", "A4 · el skill del ⑧ existe y tiene que usarse")

	out = p.paso("el sobre del ⑧ trae el PRD y la cabecera ya llena", hayTrabajo, "context")
	p.dice(out, ".docs/prd.md", "sin el PRD no hay qué convertir en reglas")

	// El ⑧ escribe el cuerpo. El marcador que dejó `sf init` es el checkpoint.
	p.constitucion("go test ./...")
	p.paso("con la constitución escrita, la 🛑", esParada, "next")
	p.paso("y approve sella", hayTrabajo, "approve")

	// ── ⑨ el backlog ────────────────────────────────────────────────────────
	//
	// A3: `sf approve` no corría ninguna compuerta, así que sellaba un backlog
	// con historias sin criterios — y sin ids, la cobertura del ⑰ cuenta cero
	// contra cero y el conteo del ㉑ también.
	p.escribir(".docs/backlog/us-1.md",
		"---\ntipo: us\nid: us-1\ntitulo: sumar dos números\n---\n"+
			"# us-1 — sumar dos números\n\n## Criterios de aceptación\n- **CA-1** —\n")
	out = p.paso("A3 · approve no puede sellar un criterio vacío", esError, "approve")
	p.dice(out, "criterio vacío", "A7 · el esqueleto de `sf new` deja el id y no el texto")

	p.escribir(".docs/backlog/us-1.md",
		"---\ntipo: us\nid: us-1\ntitulo: sumar dos números\n---\n"+
			"# us-1 — sumar dos números\n\n## Criterios de aceptación\n"+
			"- **CA-1** — Suma(2,3) devuelve 5\n")
	p.paso("con el criterio escrito sí sella", hayTrabajo, "approve")

	// ── ⑩ el roadmap ────────────────────────────────────────────────────────
	//
	// A9: la compuerta del ⑩ estaba colgada de `r == nil`, o sea que fallaba
	// leyendo el archivo que no existe. Sus dos chequeos no corrían nunca, y
	// el `sf done` del ⑩ contestaba "no hay feature en curso".
	p.escribir(".docs/backlog/us-2.md",
		"---\ntipo: us\nid: us-2\ntitulo: restar\n---\n# us-2 — restar\n\n"+
			"## Criterios de aceptación\n- **CA-1** — Resta(5,3) devuelve 2\n")
	p.escribir(".docs/roadmap.json", `{"features":[
		{"id":"f-1","slug":"suma","nombre":"la suma","orden":1,"historias":["us-1"]}]}`)
	out = p.paso("A9 · la historia que quedó afuera no la implementa nadie", esParada, "done")
	p.dice(out, "us-2", "A9 · tiene que decir CUÁL quedó huérfana")

	p.escribir(".docs/roadmap.json", `{"features":[
		{"id":"f-1","slug":"suma","nombre":"la suma","orden":1,"historias":["us-1","us-2"]}]}`)
	out = p.paso("con el roadmap completo, cierra el ⑩", hayTrabajo, "done")
	p.dice(out, "sf take", "la transición al ciclo la elige Javier")

	// ── ⑫–⑯ la planificación ────────────────────────────────────────────────
	p.paso("tomar la feature es del ⑪", hayTrabajo, "take", "f-1")
	out = p.paso("planificar pide el perfil que razona", hayTrabajo, "next")
	p.dice(out, "sf-plan", "el ⑫ lo hace sf-plan")
	p.dice(out, "razonar", "planificar es donde el modelo que piensa paga")
	// Y el modelo concreto sale del catálogo de ESTA máquina, no del binario.
	p.dice(out, "prov/razonar", "el id lo pone el catálogo, no sf")

	p.plan(`{"tareas":[
		{"id":"t-1","lote":1,"descripcion":"Suma","satisface":["us-1/CA-1"],
		 "tests":["suma_test.go::TestSuma"]}]}`)

	// A3 otra vez: el ⑰ tipeado a mano salteaba las cinco compuertas.
	out = p.paso("A3 · el ⑰ no aprueba un plan que deja un criterio afuera",
		esError, "approve")
	p.dice(out, "us-2/CA-1", "A3 · tiene que decir cuál criterio quedó sin cubrir")

	p.plan(`{"tareas":[
		{"id":"t-1","lote":1,"descripcion":"Suma","satisface":["us-1/CA-1"],
		 "tests":["suma_test.go::TestSuma"]},
		{"id":"t-2","lote":2,"descripcion":"Resta","satisface":["us-2/CA-1"],
		 "tests":["resta_test.go::TestResta"]}]}`)
	p.paso("con los dos criterios cubiertos, el plan cierra", hayTrabajo, "done")
	p.paso("y el ⑰ lo aprueba Javier", esParada, "next")
	p.paso("approve", hayTrabajo, "approve")

	// ── ⑱–⑳ implementar ─────────────────────────────────────────────────────
	//
	// A0: EL AGUJERO MÁS GRANDE. Con el plan recién aprobado f.Lotes está
	// vacío, y eso se leía como "todos los lotes cerrados": UN SOLO `sf done`
	// movía la feature a revisión sin branch, sin tests, sin código y sin
	// commit.
	out = p.paso("A0 · sin haber empezado ningún lote no se cierra nada",
		esParada, "done", "--msg", "feat: nada")
	p.dice(out, "no empezaste ningún lote", "A0 · tiene que decir qué falta")

	// Y el sobre lo dice igual, en vez de mentir con "todos commiteados", que
	// es la otra mitad de A0.
	out = p.paso("A0 · el sobre distingue 'nadie empezó' de 'todos terminaron'",
		hayTrabajo, "context")
	p.dice(out, "sf lote start", "A0 · tiene que mandar a empezar")
	p.noDice(out, "commiteados", "A0 · no hay ningún lote commiteado todavía")

	// El rojo, y las dos formas de fallarlo.
	out = p.paso("el test planificado tiene que existir", esError, "lote", "start")
	p.dice(out, "suma_test.go", "tiene que decir cuál falta")

	// Sólo el test del lote 1: `resta_test.go` es del lote 2, y escribirlo
	// ahora dejaría la suite roja por algo que no es este lote — que es
	// exactamente lo que la lista de tests planificados existe para evitar.
	p.escribir("suma_test.go", testDe("TestSuma", "Suma(2, 3)", "5"))
	p.escribir("juguete.go", "package main\n\nfunc main() {}\n")
	out = p.paso("el rojo confirmado crea la branch", hayTrabajo, "lote", "start")
	p.dice(out, "rojo confirmado", "sin ver el rojo no hay verde que valga")

	// Y recién ahora el sobre trae el lote, y SÓLO el lote que toca.
	out = p.paso("el sobre trae sólo el lote en curso", hayTrabajo, "context")
	p.dice(out, "lote 1", "el ⑱ recibe un lote, no la feature entera")
	p.noDice(out, "t-2", "el lote 2 no es asunto del subagente del lote 1")
	if b := p.branch(); b != "feat/f-1-suma" {
		t.Fatalf("branch %q, quería feat/f-1-suma: sf la crea, no la verifica", b)
	}

	// El verde, y el agujero astuto del test aflojado.
	p.escribir("juguete.go",
		"package main\n\nfunc Suma(a, b int) int { return a + b }\n\nfunc main() {}\n")
	p.paso("cerrar el lote ES commitear: sin mensaje no hay done",
		esParada, "done")

	p.escribir("suma_test.go", "package main\n\nimport \"testing\"\n\n"+
		"func TestSuma(t *testing.T) {} // aflojado\n")
	out = p.paso("el hash atrapa el test que se aflojó entre el rojo y el verde",
		esParada, "done", "--msg", "feat: la suma")
	p.dice(out, "CAMBIARON", "un test que se afloja para llegar a verde no prueba nada")

	p.escribir("suma_test.go", testDe("TestSuma", "Suma(2, 3)", "5"))
	p.paso("con el verde y los tests intactos, commitea", hayTrabajo,
		"done", "--msg", "feat: la suma")

	// El lote 2, para que la vuelta tenga más de uno. Su rojo es suyo: el test
	// que nombra el plan todavía no existe.
	p.escribir("resta_test.go", testDe("TestResta", "Resta(5, 3)", "2"))
	p.paso("el lote 2 pide su propio rojo", hayTrabajo, "lote", "start")
	// Y se implementa con un bug de verdad —clampea en cero— que TestResta no
	// ve, porque 5-3 no es negativo. Es el que el ㉑ encuentra más abajo: un
	// hallazgo inventado no se puede reproducir, y sin rojo no hay lote.
	p.escribir("juguete.go", juguete("if a < b {\n\t\treturn 0\n\t}\n\treturn a - b"))
	p.paso("y su propio commit", hayTrabajo, "done", "--msg", "feat: la resta")
	p.paso("con los dos lotes cerrados, sigue la revisión", hayTrabajo, "done")

	// ── ㉑ la revisión ──────────────────────────────────────────────────────
	p.revision(`{"vuelta":1,"veredicto":"limpio","criterios":{"us-1/CA-1":"cumple"},
		"hallazgos":[]}`)
	out = p.paso("sf no juzga la revisión: cuenta que haya ocurrido sobre TODOS",
		esParada, "done")
	p.dice(out, "us-2/CA-1", "tiene que decir sobre cuál no opinó")

	// A5: con un hallazgo abierto volvía a `implementar` y NO HABÍA DÓNDE
	// TRABAJAR — lote start rechazaba, done rebotaba a revisión, y el arreglo
	// no tenía dónde commitearse.
	p.revision(`{"vuelta":1,"veredicto":"con-hallazgos",
		"criterios":{"us-1/CA-1":"cumple","us-2/CA-1":"no-cumple"},
		"hallazgos":[{"id":"h-1","origen":21,"criterio":"us-2/CA-1","estado":"abierto",
		 "detalle":"Resta no maneja el caso de resultado negativo"}]}`)
	out = p.paso("A5 · un hallazgo abierto vuelve a implementar", esParada, "done")
	p.dice(out, "lote 3", "A5 · y abre un lote donde arreglarlo")

	out = p.paso("A5 · el sobre del lote de corrección trae el hallazgo",
		hayTrabajo, "context")
	p.dice(out, "h-1", "A5 · sin el hallazgo, el subagente fresco improvisa")
	p.dice(out, "resultado negativo", "A5 · el detalle va embebido, no por ruta")

	out = p.paso("A5 · un hallazgo sin test que lo reproduzca no se puede verificar",
		esError, "lote", "start")
	p.dice(out, "no está reproducido", "A5 · el mensaje es el del lote de corrección")

	p.escribir("negativos_test.go", testDe("TestRestaNegativa", "Resta(3, 5)", "-2"))
	p.paso("A5 · con el rojo, el lote de corrección arranca", hayTrabajo, "lote", "start")
	p.escribir("juguete.go", juguete("return a - b")) // el arreglo: se saca el clamp
	p.paso("A5 · y termina en un commit, como cualquier otro lote", hayTrabajo,
		"done", "--msg", "fix: la resta con resultado negativo")
	p.paso("los lotes vuelven a estar todos cerrados", hayTrabajo, "done")

	p.revision(`{"vuelta":2,"veredicto":"limpio",
		"criterios":{"us-1/CA-1":"cumple","us-2/CA-1":"cumple"},"hallazgos":[]}`)
	p.paso("revisión limpia: sigue el cierre", hayTrabajo, "done")

	// ── ㉓ el cierre y el archivado ──────────────────────────────────────────
	//
	// A3: archivar es irreversible, así que la compuerta va ANTES.
	out = p.paso("A3 · no se archiva sin la doc y el journal", esError, "approve")
	p.dice(out, "doc.md", "A3 · tiene que decir qué falta")
	if p.existe(".docs/archivado/f-1-suma") {
		t.Fatal("A3 · movió la carpeta con la compuerta en rojo")
	}

	p.escribir(".docs/features/f-1-suma/doc.md", "# f-1 — la suma\nSuma y resta.\n")
	p.escribir(".docs/features/f-1-suma/journal.md", "- el ㉑ encontró los negativos\n")
	p.paso("con los dos archivos, lista para archivar", hayTrabajo, "done")

	// A2: FALLABA SIEMPRE. `.docs/estado.json` está sucio por construcción al
	// llegar acá, así que el `git checkout` abortaba — y como el movimiento de
	// la carpeta iba primero, el repo quedaba sin salida.
	if !p.sucio() {
		t.Fatal("A2 · el test se apoya en que el estado esté sucio, y no lo está")
	}
	out = p.paso("A2 · archivar con el estado sucio, que es el caso normal",
		hayTrabajo, "approve")
	p.dice(out, "archivada", "A2 · la feature se cierra")

	if b := p.branch(); b != "main" {
		t.Fatalf("A2 · quedó en %q, quería main", b)
	}
	if !p.existe(".docs/archivado/f-1-suma/journal.md") {
		t.Fatal("A2 · la carpeta no viajó entera")
	}
	if p.sucio() {
		t.Fatal("A2 · el archivado quedó sin commitear: el próximo lote se lo lleva puesto")
	}

	p.paso("y no queda nada", noQuedaNada, "next")
}

// ────────────────────────────────────────────────────────────────────────────
// Vuelta B — un bug por `sf new`, el camino corto
// ────────────────────────────────────────────────────────────────────────────

func TestVueltaDeUnBugPorElCaminoCorto(t *testing.T) {
	p := nuevoProyecto(t)
	p.productoListo()

	// ── La entrada C ────────────────────────────────────────────────────────
	//
	// A7: `sf new` reabría la ⏸ del ⑨, pero el checkpoint preguntaba "¿hay
	// historias?" — y con un producto en marcha eso es siempre sí. Salía la ⏸
	// y lo que quedaba para aprobar era el esqueleto vacío.
	p.paso("sf new mete la entrada al backlog, que es el embudo",
		hayTrabajo, "new", "Suma no valida el overflow")

	out := p.paso("A7 · next tiene que mandar a completarla, no a la ⏸",
		hayTrabajo, "next")
	p.dice(out, "sfp-backlog", "A7 · el ⑨ es quien la completa")
	p.dice(out, "us-2", "A7 · y tiene que decir CUÁL, no 'partí el PRD'")

	out = p.paso("A7 · el sobre la nombra en vez de dar el PRD entero",
		hayTrabajo, "context")
	p.dice(out, ".docs/backlog/us-2.md", "A7 · el que pinponea necesita ESA historia")

	// El ⑨ la completa, y la marca como bug con su historia original.
	p.escribir(".docs/backlog/us-2.md",
		"---\ntipo: bug\nid: us-2\ntitulo: la suma no satura\nrelacionado_a: us-1\n---\n"+
			"# us-2 — la suma no satura\n\n## Criterios de aceptación\n"+
			"- **CA-1** — Suma(MaxInt, 1) devuelve MaxInt\n")
	p.paso("con la historia completa, la ⏸", esParada, "next")
	p.paso("approve", hayTrabajo, "approve")

	p.escribir(".docs/roadmap.json", `{"features":[
		{"id":"f-1","slug":"suma","nombre":"la suma","orden":1,"historias":["us-1"]},
		{"id":"f-2","slug":"overflow","nombre":"el overflow","orden":2,"historias":["us-2"]}]}`)
	p.paso("el ⑩ vuelve a correr por la historia nueva", hayTrabajo, "done")
	p.paso("take f-2", hayTrabajo, "take", "f-2")

	// ── El camino corto ─────────────────────────────────────────────────────
	//
	// A1: `sf next` aplicaba el salteo y `sf lote start` no, así que los dos
	// comandos se contradecían sobre la misma feature. El único avance posible
	// era `sf done`, que con f.Lotes vacío mandaba el bug DIRECTO A CIERRE:
	// sin una línea de código, sin test, sin rojo y sin commit.
	out = p.paso("A1 · un bug entra directo a implementar", hayTrabajo, "next")
	p.dice(out, "sf-build", "A1 · saltea planificación")
	p.dice(out, "sf lote start", "A1 · y lo que sigue es empezar el lote")

	// A6: el sobre leía el estado crudo, así que servía el del ⑫ mientras
	// `next` ya había lanzado a sf-build.
	out = p.paso("A6 · el sobre tiene que ser el de implementar", hayTrabajo, "context")
	p.noDice(out, "Qué hay que resolver", "A6 · ése es el sobre del ⑫")
	p.dice(out, ".docs/archivado/f-1-suma/spec-design.md",
		"A6 · el bug rompió algo archivado: sin esa spec se arranca de cero")

	out = p.paso("A1 · el bug no puede llegar a cierre sin escribir nada",
		esParada, "done")
	p.dice(out, "no empezaste ningún lote", "A0+A1 · la misma puerta")

	// Sin plan, la compuerta afloja a lo comprobable: que la suite falle.
	out = p.paso("A1 · y sin el bug reproducido tampoco arranca", esError, "lote", "start")
	p.dice(out, "no está reproducido", "el camino corto exige el rojo igual")

	p.escribir("overflow_test.go", "package main\n\nimport (\n\t\"math\"\n\t\"testing\"\n)\n\n"+
		"func TestSumaSatura(t *testing.T) {\n"+
		"\tif Suma(math.MaxInt, 1) != math.MaxInt {\n\t\tt.Fatal(\"no satura\")\n\t}\n}\n")
	out = p.paso("A1 · con el rojo, el lote del bug arranca", hayTrabajo, "lote", "start")
	p.dice(out, "camino corto", "A1 · y dice que va sin plan")

	p.escribir("juguete.go", "package main\n\nimport \"math\"\n\n"+
		"func Suma(a, b int) int {\n"+
		"\tif b > 0 && a > math.MaxInt-b {\n\t\treturn math.MaxInt\n\t}\n"+
		"\treturn a + b\n}\n\nfunc main() {}\n")
	p.paso("A1 · el arreglo se commitea como cualquier lote", hayTrabajo,
		"done", "--msg", "fix: la suma satura")

	out = p.paso("A1 · un bug saltea la revisión y va derecho al cierre",
		hayTrabajo, "done")
	p.dice(out, "cierre", "A1 · el camino corto saltea DOS estados, no uno")

	p.escribir(".docs/features/f-2-overflow/doc.md", "# f-2 — el overflow\n")
	p.escribir(".docs/features/f-2-overflow/journal.md", "- int no es infinito\n")
	p.paso("el ㉓ corre igual para un bug", hayTrabajo, "done")
	p.paso("y se archiva", hayTrabajo, "approve")
	p.paso("no queda nada", noQuedaNada, "next")
}

// ────────────────────────────────────────────────────────────────────────────
// La cadena de modelo, que vive fuera de la vuelta feliz
// ────────────────────────────────────────────────────────────────────────────

// A8: `modeloDeFeature` tenía un solo consumidor —`implementando`—, así que
// `sf model` no hacía nada en los otros tres estados de feature. Es la salida
// de ME TRABÉ, o sea la que se usa mirando el bucle patinar.
func TestElModeloDeJavierMandaEnLosCuatroEstados(t *testing.T) {
	p := nuevoProyecto(t)
	p.productoListo()
	// f-2 y no f-1: la del andamio ya está archivada, y en `cierre` su doc y su
	// journal existen —la compuerta los busca también en `.docs/archivado/`—,
	// así que saldría la ⏸ en vez del trabajo. f-2 no tiene carpeta en ningún
	// lado, que es lo que hace falta para recorrer los cuatro estados.
	p.escribir(".docs/roadmap.json", `{"features":[
		{"id":"f-1","slug":"suma","nombre":"la suma","orden":1,"historias":["us-1"]},
		{"id":"f-2","slug":"otra","nombre":"otra","orden":2,"historias":["us-1"]}]}`)

	//
	// Sin `t.Run`: el andamio guarda el `*testing.T` del padre, y un Fatal
	// desde adentro de un subtest sobre ese t es un Goexit en el hilo
	// equivocado. El estado ya va nombrado en cada aserción.
	// Los dos perfiles que la máquina pide, y qué estado pide cada uno. Es el
	// mapa de `perfilPorEstado` visto desde afuera del binario.
	for _, caso := range []struct{ estado, perfil string }{
		{"planificacion", "razonar"},
		{"implementar", "construir"},
		{"revision", "razonar"},
		{"cierre", "construir"},
	} {
		p.enEstado("f-2", caso.estado, "")
		out := p.paso("el perfil de "+caso.estado, hayTrabajo, "next")
		p.dice(out, caso.perfil, "el perfil de "+caso.estado+" es "+caso.perfil)

		// A8 · la decisión de runtime le gana a todo, en los cuatro estados. Y
		// ahora se pide por ALIAS, que es el vocabulario de Javier.
		p.enEstado("f-2", caso.estado, "a-construir")
		out = p.paso("A8 · sf model en "+caso.estado, hayTrabajo, "next")
		p.dice(out, "prov/construir", "A8 · la decisión de runtime le gana a todo, en "+caso.estado)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// Ayudantes que arman artefactos del flujo
// ────────────────────────────────────────────────────────────────────────────

// constitucion escribe la constitución CON EL CUERPO, o sea sin el marcador que
// deja `sf init` — que es el checkpoint del ⑧.
// brief deja los TRES archivos del ①–⑤, que es lo que el ⑥ exige desde el
// 2026-09-07: el argumento, el registro de la entrevista y la evidencia.
//
// Están los tres y no uno porque la compuerta los mira a los tres: `abiertas: 0`
// dice que la entrevista terminó, y el link vive en la evidencia, no en el
// brief.
func (p *proyecto) brief() {
	p.t.Helper()
	p.escribir(".docs/brief.md",
		"---\nveredicto: hacelo\n---\n# Brief\nUn sumador.\n")
	p.escribir(".docs/entrevista.md",
		"---\nrondas: 2\npreguntas: 7\nabiertas: 0\n---\n"+
			"# Ronda 1\n\n**Q1 — qué es**: un sumador. ✅\n")
	p.escribir(".docs/evidencia.md",
		"---\nretrieved: 1\nmodel_prior: 0\nlinks: 1\n---\n"+
			"## ¿ya existe?\n\n- suma-cli hace lo mismo y no exporta. [retrieved]\n"+
			"  - https://github.com/x/suma-cli\n")
}

func (p *proyecto) constitucion(testCmd string) {
	p.t.Helper()
	p.escribir(".docs/constitucion.md",
		"---\nlenguaje: go\nmanifiesto: go.mod\ntest_cmd: "+testCmd+"\n"+
			"mutacion: \"\"\ndependencias_aprobadas: []\ngit:\n"+
			"  branch_por_feature: true\n  patron_branch: \"feat/{feature-id}-{slug}\"\n"+
			"  commit: conventional\n  merge: no-ff\n  branch_base: main\n---\n\n"+
			"# Constitución\n\n## Arquitectura\nUn paquete, sin capas.\n")
}

// plan escribe los tres archivos del ⑫–⑯ con el tareas.json que se le pase.
func (p *proyecto) plan(tareas string) {
	p.t.Helper()
	p.escribir(".docs/features/f-1-suma/decision.md",
		"# Decisión — f-1\n\n## A — funciones sueltas\nLo más simple. ✅\n\n"+
			"## B — un tipo Calculadora\nEstado que no hace falta.\n\n"+
			"## C — generics\nOverkill.\n")
	p.escribir(".docs/features/f-1-suma/spec-design.md",
		"# Spec — f-1\n\nfunc Suma(a, b int) int\nfunc Resta(a, b int) int\n")
	p.escribir(".docs/features/f-1-suma/tareas.json", tareas)
}

func (p *proyecto) revision(json string) {
	p.t.Helper()
	p.escribir(".docs/features/f-1-suma/revision.json", json)
}

// productoListo deja el producto sellado y f-1 cerrada y archivada, para los
// tests que empiezan a mitad del flujo. Es la vuelta A comprimida, sin
// aserciones: lo que ese camino prueba ya lo prueba su propio test.
func (p *proyecto) productoListo() {
	p.t.Helper()

	mustSf := func(args ...string) {
		p.t.Helper()
		if code, out := p.sf(args...); code != hayTrabajo {
			p.t.Fatalf("andamio: sf %v → exit %d\n%s", args, code, sangrar(out))
		}
	}

	mustSf("install")
	mustSf("init")
	p.conPerfiles()
	p.brief()
	mustSf("approve")
	p.escribir(".docs/prd.md", "# PRD\nSumar.\n")
	mustSf("done")
	p.constitucion("go test ./...")
	mustSf("approve")
	p.escribir(".docs/backlog/us-1.md",
		"---\ntipo: us\nid: us-1\ntitulo: sumar dos números\n---\n"+
			"# us-1 — sumar\n\n## Criterios de aceptación\n- **CA-1** — Suma(2,3) da 5\n")
	mustSf("approve")
	p.escribir(".docs/roadmap.json", `{"features":[
		{"id":"f-1","slug":"suma","nombre":"la suma","orden":1,"historias":["us-1"]}]}`)
	mustSf("done")
	mustSf("take", "f-1")
	p.plan(`{"tareas":[
		{"id":"t-1","lote":1,"descripcion":"Suma","satisface":["us-1/CA-1"],
		 "tests":["suma_test.go::TestSuma"]}]}`)
	mustSf("done")
	mustSf("approve")
	p.escribir("suma_test.go", testDe("TestSuma", "Suma(2, 3)", "5"))
	p.escribir("juguete.go", "package main\n\nfunc main() {}\n")
	mustSf("lote", "start")
	p.escribir("juguete.go",
		"package main\n\nfunc Suma(a, b int) int { return a + b }\n\nfunc main() {}\n")
	mustSf("done", "--msg", "feat: la suma")
	mustSf("done")
	p.revision(`{"vuelta":1,"veredicto":"limpio","criterios":{"us-1/CA-1":"cumple"},
		"hallazgos":[]}`)
	mustSf("done")
	p.escribir(".docs/features/f-1-suma/doc.md", "# f-1 — la suma\n")
	p.escribir(".docs/features/f-1-suma/journal.md", "- nada raro\n")
	mustSf("done")
	mustSf("approve")
}

// enEstado edita el estado.json a mano para pararse en un punto del ciclo.
//
// Es la única parte de este archivo que no pasa por el binario, y se hace sólo
// donde lo que se prueba es la LECTURA del estado —la cadena de modelo—, no la
// transición. Llegar ahí por el flujo costaría una vuelta entera por caso y no
// probaría nada más.
func (p *proyecto) enEstado(feature, est, modelo string) {
	p.t.Helper()

	// En `implementar` el lote tiene que estar ABIERTO. Con todos commiteados
	// la máquina contesta "falta cerrar la feature", y esa rama no lanza a
	// nadie —lo que falta es un `sf done`— así que va sin skill ni modelo: no
	// habría nada que comparar.
	commit := `"abc123"`
	if est == "implementar" {
		commit = "null"
	}

	p.escribir(".docs/estado.json", `{
	  "producto": {"brief_sellado":"hacelo","prd_hash":"abc123",
	               "constitucion_sellada":true,"backlog_visto":true},
	  "feature_actual": "`+feature+`",
	  "features": {"`+feature+`": {"estado":"`+est+`","base_commit":"",
	               "modelo":"`+modelo+`","intentos_fallidos":0,
	               "lotes":[{"lote":1,"rojo":true,"hash_tests":"h","commit":`+commit+`}]}}
	}`)
}

// juguete arma el módulo del proyecto de prueba con el cuerpo de Resta que se
// le pase. Suma no cambia; lo que varía es Resta, que es donde vive el bug que
// el ㉑ encuentra.
//
// El lote 1 no lo usa: ahí Resta todavía NO EXISTE, y adelantarla dejaría el
// lote 2 naciendo en verde — o sea sin rojo que confirmar.
func juguete(cuerpoDeResta string) string {
	return "package main\n\n" +
		"func Suma(a, b int) int { return a + b }\n\n" +
		"func Resta(a, b int) int {\n\t" + cuerpoDeResta + "\n}\n\n" +
		"func main() {}\n"
}

// testDe arma un test de Go que compara una expresión con un valor.
func testDe(nombre, expresion, esperado string) string {
	return "package main\n\nimport \"testing\"\n\n" +
		"func " + nombre + "(t *testing.T) {\n" +
		"\tif " + expresion + " != " + esperado + " {\n" +
		"\t\tt.Fatal(\"" + nombre + "\")\n\t}\n}\n"
}

// ────────────────────────────────────────────────────────────────────────────
// Preguntas sobre el repo, hechas desde AFUERA de sf
// ────────────────────────────────────────────────────────────────────────────
//
// Ninguna usa el paquete `git` de sf: un test que se apoya en la función que
// está probando no prueba nada.

func (p *proyecto) branch() string {
	p.t.Helper()
	return strings.TrimSpace(p.gitSalida("rev-parse", "--abbrev-ref", "HEAD"))
}

func (p *proyecto) sucio() bool {
	p.t.Helper()
	return strings.TrimSpace(p.gitSalida("status", "--porcelain")) != ""
}

func (p *proyecto) gitSalida(args ...string) string {
	p.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = p.raiz
	out, err := cmd.Output()
	if err != nil {
		p.t.Fatalf("git %v: %v", args, err)
	}
	return string(out)
}

func (p *proyecto) existe(rel string) bool {
	_, err := os.Stat(filepath.Join(p.raiz, rel))
	return err == nil
}

// ────────────────────────────────────────────────────────────────────────────
// El inventario de comandos y el `switch` que despacha
// ────────────────────────────────────────────────────────────────────────────

// TestElInventarioDeComandosNoMiente comprueba que la lista y el despacho digan
// lo mismo.
//
// `comandos.Todos` existe porque la lista hacía falta en tres lugares y estaba
// escrita en dos. Pero eliminar la tercera copia —el `switch` de main.go— no se
// puede sin convertir el despacho en un mapa de funciones, que es peor: se
// pierde el orden de lectura y la firma de cada comando.
//
// Lo que sí se puede es comprobar que las dos coincidan, y eso sólo se ve desde
// afuera: por adentro, la lista es un slice de strings que compila igual esté
// bien o mal. Acá se le pasa cada nombre al binario de verdad y se mira que no
// conteste "todavía no está construido".
//
// El día que alguien agregue un comando al switch y se olvide de la lista, el
// que se entera es `sf doctor`, que va a acusar a un skill correcto de nombrar
// un comando inexistente. Este test lo agarra antes.
func TestElInventarioDeComandosNoMiente(t *testing.T) {
	p := nuevoProyecto(t)

	for _, c := range comandos.Todos {
		// Se parte en palabras porque `lote start` es de dos y el binario las
		// lee como dos argumentos.
		_, out := p.sf(strings.Fields(c)...)

		if strings.Contains(out, "todavía no está construido") {
			t.Errorf("`sf %s` está en comandos.Todos y el switch de main.go no lo atiende:\n%s",
				c, sangrar(out))
		}
	}
}

// TestNingunComandoDelSwitchFaltaEnElInventario es el espejo del anterior.
//
// El agujero que cubre es el que importa para el doctor: un comando que existe
// y NO está en la lista hace que `comandos.Existe` diga que no, y el doctor
// acusa a un skill correcto de nombrar algo inexistente. Es un falso positivo
// que hace ignorar el chequeo entero.
//
// La ayuda es la fuente: `sf help` los enumera para un humano, y si un comando
// no está ahí tampoco está documentado.
func TestNingunComandoDelSwitchFaltaEnElInventario(t *testing.T) {
	p := nuevoProyecto(t)
	_, ayuda := p.sf("help")

	for _, l := range strings.Split(ayuda, "\n") {
		m := reAyuda.FindStringSubmatch(l)
		if m == nil {
			continue
		}
		if !comandos.Existe(m[1]) {
			t.Errorf("`sf %s` sale en la ayuda y no está en comandos.Todos:\n  %s", m[1], l)
		}
	}
}

// TestModelos_LosTresDesenlacesSeDistinguenDesdeAfuera
//
// `sf models` puede terminar de tres formas que se parecen y NO son lo mismo, y
// el que las lee es alguien —o un agente— tratando de llenar el catálogo:
//
//	un arnés mal escrito   error de quien llama       exit 1
//	un arnés que no lista  un hecho del mundo         exit 2 + cómo seguir a mano
//	un arnés que lista     ids por stdout             exit 0
//
// Confundir las dos primeras es lo que hacía la primera versión: contestaba
// "buscá el id a mano" ante un `--harness=emacs`, que es un consejo inútil para
// un typo. El código de salida es la mitad de la interfaz (ver el bloque de
// códigos en main.go), así que esto se prueba desde afuera y no leyendo texto.
func TestModelos_LosTresDesenlacesSeDistinguenDesdeAfuera(t *testing.T) {
	p := nuevoProyecto(t)

	code, out := p.sf("models", "--harness=emacs")
	if code != 1 {
		t.Errorf("un arnés que no existe tiene que ser error (1), fue %d:\n%s", code, sangrar(out))
	}
	if strings.Contains(out, "a mano") {
		t.Errorf("a un typo no se le ofrece el camino manual, se le dice cuáles hay:\n%s", sangrar(out))
	}
	if !strings.Contains(out, "claude-code") {
		t.Errorf("el error tiene que nombrar los que sí existen:\n%s", sangrar(out))
	}

	code, out = p.sf("models", "--harness=claude-code")
	if code != 2 {
		t.Errorf("un arnés sin listado es parada (2), no error: fue %d\n%s", code, sangrar(out))
	}
	if !strings.Contains(out, "sf model ") {
		t.Errorf("sin listado, la salida tiene que decir cómo declararlo igual:\n%s", sangrar(out))
	}
}

// TestInstall_SinTTYNoPreguntaNiCuelga
//
// EL TEST QUE NO PUEDE FALLAR NUNCA.
//
// `sf install` es el único comando que conversa, y conversa sólo si hay una
// persona: la shell de un agente no tiene TTY y la de una persona sí (medido).
// Si esa condición se rompiera, un `sf install` corrido por el orquestador —lo
// nombra en "Empezar de cero"— o por `install.sh` se quedaría esperando una
// respuesta que nadie va a escribir, y el que se cuelga es el primer comando que
// corre alguien que recién llega.
//
// Acá el binario corre con stdin conectado a un pipe, que es exactamente lo que
// ve un agente. Que el test TERMINE ya es la mitad de la prueba.
func TestInstall_SinTTYNoPreguntaNiCuelga(t *testing.T) {
	p := nuevoProyecto(t)

	code, out := p.sf("install")
	if code != 0 {
		t.Fatalf("exit %d:\n%s", code, sangrar(out))
	}
	if strings.Contains(out, "¿Qué arneses") {
		t.Errorf("preguntó sin que hubiera nadie del otro lado:\n%s", sangrar(out))
	}
	if strings.Contains(out, "alias corto") {
		t.Errorf("llegó a pedir un alias sin TTY:\n%s", sangrar(out))
	}
}

// Y con flags tampoco pregunta, haya persona o no: pasar `--harness=` ES la
// respuesta, y volver a preguntarla sería no escuchar.
func TestInstall_ConFlagsNoPregunta(t *testing.T) {
	p := nuevoProyecto(t)

	code, out := p.sf("install", "--harness=opencode,commandcode")
	if code != 0 {
		t.Fatalf("exit %d:\n%s", code, sangrar(out))
	}
	if strings.Contains(out, "¿Qué arneses") {
		t.Errorf("preguntó teniendo la respuesta en un flag:\n%s", sangrar(out))
	}
	for _, h := range []string{"opencode", "commandcode"} {
		if !strings.Contains(out, h) {
			t.Errorf("no instaló para %s:\n%s", h, sangrar(out))
		}
	}
}

// reAyuda caza los comandos de la ayuda: dos espacios, `sf`, el nombre.
var reAyuda = regexp.MustCompile(`^\s+sf (lote start|[a-z]+)`)

// ────────────────────────────────────────────────────────────────────────────
// SECCIONAR — un arnés planifica, otro implementa, sobre el mismo repo
// ────────────────────────────────────────────────────────────────────────────

// EL GUION QUE PRUEBA LA IDEA ENTERA.
//
// Javier planifica en Claude Code con el modelo de su suscripción, cierra, y
// abre Command Code —donde tiene otro modelo, gratis por unos días— en el mismo
// repo. `sf next` le contesta lo mismo, porque el estado está en disco, PERO
// con los modelos del arnés donde está parado.
//
// Esto no funcionaba mientras el puntero de harness era un campo escrito: al
// abrir el segundo arnés, sf resolvía los ids del primero — que ahí no existen.
func TestSeccionar_UnArnesPlanificaYOtroImplementa(t *testing.T) {
	p := nuevoProyecto(t)

	// ── En Claude Code: el producto entero, y una feature f-2 sin empezar ───
	p.enArnes("claude-code")
	p.productoListo() // declara sus perfiles con ids `prov/<perfil>`
	p.escribir(".docs/roadmap.json", `{"features":[
		{"id":"f-1","slug":"suma","nombre":"la suma","orden":1,"historias":["us-1"]},
		{"id":"f-2","slug":"otra","nombre":"otra","orden":2,"historias":["us-1"]}]}`)
	p.enEstado("f-2", "planificacion", "")

	out := p.paso("planificar corre con el modelo de claude-code", hayTrabajo, "next")
	p.dice(out, "prov/razonar", "el id sale del bloque de claude-code")
	p.noDice(out, "agente:", "en claude-code el modelo va en la llamada, no en un archivo")

	// ── Se muda a Command Code. MISMO REPO, sin tocar un solo archivo ───────
	p.enArnes("commandcode")

	out = p.paso("el otro arnés no hereda los modelos del primero", esParada, "next")
	p.dice(out, "PARÁ", "sin modelos declarados acá, sf tiene que frenar")
	p.dice(out, "commandcode", "y decir en qué arnés falta")
	p.noDice(out, "prov/razonar", "NO puede ofrecer un id que en este arnés no existe")
	p.dice(out, "reiniciá", "y avisar que los agentes se leen al arrancar")

	p.paso("el modelo gratis del otro arnés", hayTrabajo,
		"model", "razonar", "--alias", "el-gratis", "--id", "prov/gratis",
		"--via", "subagente", "--esfuerzo", "high")

	out = p.paso("y ahora resuelve con el suyo", hayTrabajo, "next")
	p.dice(out, "prov/gratis", "el id sale del bloque de commandcode")
	p.dice(out, "high", "y el esfuerzo con el que se declaró")
	p.dice(out, "sf-el-gratis", "acá SÍ hace falta portamodelo: el modelo va en el archivo")

	// Y el portamodelo quedó escrito de verdad, con su esfuerzo.
	b, err := os.ReadFile(filepath.Join(p.raiz, ".commandcode", "agents", "sf-el-gratis.md"))
	if err != nil {
		t.Fatalf("no escribió el portamodelo: %v", err)
	}
	if !strings.Contains(string(b), "model: prov/gratis") ||
		!strings.Contains(string(b), "reasoningEffort: high") {
		t.Errorf("el portamodelo no lleva modelo y esfuerzo:\n%s", b)
	}

	// ── Y volver al primero no perdió nada ──────────────────────────────────
	p.enArnes("claude-code")
	out = p.paso("volver a claude-code encuentra su bloque intacto", hayTrabajo, "next")
	p.dice(out, "prov/razonar", "el bloque del primero sobrevivió a la mudanza")
	p.noDice(out, "el-gratis", "y no se cruzó con el del otro")
}
