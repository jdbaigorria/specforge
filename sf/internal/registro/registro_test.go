package registro

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/jdbaigorria/specforge/sf/internal/estado"
	"github.com/jdbaigorria/specforge/sf/internal/global"
)

// proyecto arma una carpeta con `.docs/estado.json` y devuelve la raíz.
func proyecto(t *testing.T, estadoJSON string) string {
	t.Helper()
	raiz := t.TempDir()
	if err := os.MkdirAll(filepath.Join(raiz, ".docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if estadoJSON != "" {
		if err := os.WriteFile(filepath.Join(raiz, estado.Archivo), []byte(estadoJSON), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return raiz
}

// leerLineas devuelve las entradas escritas, sin pasar por `Leer` — así un bug
// del lector no puede tapar un bug del escritor.
func leerLineas(t *testing.T, raiz string) []Entrada {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(raiz, Archivo))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatal(err)
	}
	var es []Entrada
	for _, l := range strings.Split(strings.TrimRight(string(b), "\n"), "\n") {
		var e Entrada
		if err := json.Unmarshal([]byte(l), &e); err != nil {
			t.Fatalf("línea no es JSON válido: %v\n%s", err, l)
		}
		es = append(es, e)
	}
	return es
}

// sinArnes limpia las variables de detección: la suite corre ADENTRO de un
// arnés, y sin esto todos los tests de `quien` heredarían el de quien la lanzó.
func sinArnes(t *testing.T) {
	t.Helper()
	for _, v := range global.VarsDeHarness {
		t.Setenv(v, "")
	}
	t.Setenv(VarDelegado, "")
}

const estadoConFeature = `{
  "producto": {"brief_sellado":"hacelo","prd_hash":"a3f9c1","constitucion_sellada":true,"backlog_visto":true},
  "feature_actual": "f-2",
  "features": {"f-2": {"estado":"implementar","intentos_fallidos":2,
    "lotes":[{"lote":1,"rojo":true,"hash_tests":"aa","commit":"c1"},
             {"lote":2,"rojo":true,"hash_tests":"bb"}]}}
}`

// ────────────────────────────────────────────────────────────────────────────
// LA REGLA DURA: EL REGISTRO NO PUEDE ROMPER UN COMANDO
// ────────────────────────────────────────────────────────────────────────────

// Es el test que más importa del archivo. Un tablero que puede apagar el motor
// es peor que no tener tablero, así que escribir donde no se puede escribir
// tiene que ser un no-evento: ni panic, ni error, ni nada.
func TestEscribirDondeNoSePuedeNoRompeNada(t *testing.T) {
	sinArnes(t)
	raiz := proyecto(t, estadoConFeature)

	// La carpeta del registro existe y es de sólo lectura: crearla adentro
	// fallaría, y abrir el archivo también.
	dir := filepath.Join(raiz, global.Carpeta)
	if err := os.MkdirAll(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(dir, 0o755) })

	Abrir(raiz, "next", nil)
	Cerrar(0) // si esto explota, el test falla solo

	if _, err := os.Stat(filepath.Join(raiz, Archivo)); err == nil {
		t.Skip("el filesystem ignoró el 0555 (suele pasar como root): el test no prueba nada acá")
	}
}

// Y el mismo caso desde el otro lado: sin proyecto no se escribe NADA, para que
// un `sf version` tipeado en una carpeta cualquiera no deje basura.
func TestSinProyectoNoEscribeNada(t *testing.T) {
	sinArnes(t)
	raiz := t.TempDir() // ni .docs/ ni .specforge/

	Abrir(raiz, "version", nil)
	Cerrar(0)

	if _, err := os.Stat(filepath.Join(raiz, Archivo)); !os.IsNotExist(err) {
		t.Error("escribió el registro en una carpeta que no es un proyecto")
	}
}

// La comprobación de "¿esto es un proyecto?" va al CERRAR y no al abrir, y este
// test es el que fija esa decisión: `sf init` empieza sin carpetas y las crea, y
// su línea —la primera de la película— tiene que quedar igual.
func TestInitQuedaRegistradoAunqueAlAbrirNoHubieraProyecto(t *testing.T) {
	sinArnes(t)
	raiz := t.TempDir()

	Abrir(raiz, "init", nil)
	// Lo que hace `sf init`: crear el andamio.
	if err := os.MkdirAll(filepath.Join(raiz, ".docs"), 0o755); err != nil {
		t.Fatal(err)
	}
	Cerrar(0)

	es := leerLineas(t, raiz)
	if len(es) != 1 || es[0].Cmd != "init" {
		t.Fatalf("el `sf init` no quedó registrado: %+v", es)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// LA FOTO
// ────────────────────────────────────────────────────────────────────────────

func TestLaFotoTraeElEstadoYElLoteEnCurso(t *testing.T) {
	sinArnes(t)
	raiz := proyecto(t, estadoConFeature)

	f := foto(raiz)
	if f.Feature != "f-2" || f.Estado != "implementar" {
		t.Errorf("feature/estado = %q/%q, quería f-2/implementar", f.Feature, f.Estado)
	}
	// El lote en curso es el primero SIN commit: el 1 ya está cerrado.
	if f.Lote != 2 {
		t.Errorf("lote = %d, quería 2", f.Lote)
	}
	if f.Intentos != 2 {
		t.Errorf("intentos = %d, quería 2", f.Intentos)
	}
	if !f.Producto.Constitucion || f.Producto.Brief != "hacelo" {
		t.Errorf("los sellos de producto no viajaron: %+v", f.Producto)
	}
}

// Sin estado.json la foto queda vacía y no es un error: es un proyecto sin
// arrancar, y el registro tiene que poder escribir igual.
func TestSinEstadoLaFotoQuedaVaciaYNoFalla(t *testing.T) {
	sinArnes(t)
	raiz := proyecto(t, "")

	Abrir(raiz, "next", nil)
	Cerrar(2)

	es := leerLineas(t, raiz)
	if len(es) != 1 {
		t.Fatalf("quería 1 línea, hay %d", len(es))
	}
	if es[0].Antes.Feature != "" || es[0].Antes.Estado != "" {
		t.Errorf("la foto no quedó vacía: %+v", es[0].Antes)
	}
}

// El antes y el después son dos lecturas distintas del archivo, y este test es
// el que lo comprueba: si `Cerrar` reusara la foto de `Abrir`, una transición
// sería invisible — y las transiciones son el punto del paquete.
func TestElDespuesSeLeeDeNuevoYVeLaTransicion(t *testing.T) {
	sinArnes(t)
	raiz := proyecto(t, estadoConFeature)

	Abrir(raiz, "done", nil)

	// Lo que haría `sf done`: mover la feature a revisión.
	movido := strings.Replace(estadoConFeature, `"estado":"implementar"`, `"estado":"revision"`, 1)
	if err := os.WriteFile(filepath.Join(raiz, estado.Archivo), []byte(movido), 0o644); err != nil {
		t.Fatal(err)
	}
	Cerrar(0)

	es := leerLineas(t, raiz)
	if len(es) != 1 {
		t.Fatalf("quería 1 línea, hay %d", len(es))
	}
	if es[0].Antes.Estado != "implementar" || es[0].Despues.Estado != "revision" {
		t.Errorf("antes/después = %q/%q, quería implementar/revision",
			es[0].Antes.Estado, es[0].Despues.Estado)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// QUIÉN CORRIÓ — los tres valores, y el límite medido
// ────────────────────────────────────────────────────────────────────────────

func TestQuienCorrio(t *testing.T) {
	casos := []struct {
		nombre   string
		delegado string
		tty      bool
		harness  string
		quiero   string
	}{
		// El único fiable: la marca la puso sf mismo, en `sf lanzar`.
		{"con la marca del hijo", "l-7", false, "", Delegado},
		// Y le gana al TTY: un delegado que además tenga terminal sigue siendo
		// un delegado. La marca es más fuerte que la señal.
		{"la marca le gana al tty", "l-7", true, "", Delegado},

		{"terminal de verdad", "", true, "", Terminal},

		// Adentro de un arnés NO es una terminal aunque haya TTY: es
		// exactamente la firma de un `script(1)` que lo fabricó.
		{"tty fabricado adentro de un arnés", "", true, "claude-code", Agente},

		// Y el caso normal, que es el que NO distingue al orquestador de Javier
		// tipeando `!`. Está medido y es un límite, no un pendiente.
		{"sin tty y sin marca", "", false, "", Agente},
	}

	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			sinArnes(t)
			t.Setenv(VarDelegado, c.delegado)
			if c.harness != "" {
				t.Setenv(global.VarHarness, c.harness)
			}
			viejo := hayTTY
			hayTTY = func() bool { return c.tty }
			t.Cleanup(func() { hayTTY = viejo })

			q, tty, del := quienCorrio()
			if q != c.quiero {
				t.Errorf("quien = %q, quería %q", q, c.quiero)
			}
			if tty != c.tty {
				t.Errorf("la señal cruda del tty se perdió: %v", tty)
			}
			if del != c.delegado {
				t.Errorf("delegado = %q, quería %q", del, c.delegado)
			}
		})
	}
}

// La etiqueta es una interpretación; las señales son el hecho. Este test fija
// que las señales viajan al archivo ADEMÁS de la etiqueta, que es lo que permite
// releer las líneas viejas el día que la regla cambie.
func TestLasSenalesCrudasViajanAdemasDeLaEtiqueta(t *testing.T) {
	sinArnes(t)
	t.Setenv(VarDelegado, "l-9")
	raiz := proyecto(t, estadoConFeature)

	Abrir(raiz, "approve", nil)
	Cerrar(0)

	es := leerLineas(t, raiz)
	if len(es) != 1 {
		t.Fatalf("quería 1 línea, hay %d", len(es))
	}
	e := es[0]
	if e.Quien != Delegado || e.Delegado != "l-9" {
		t.Errorf("quien/delegado = %q/%q", e.Quien, e.Delegado)
	}
	if e.Pid == 0 || e.Ppid == 0 {
		t.Error("los pids no viajaron")
	}
	if e.Arnes == "" {
		t.Error("arnes vacío: DetectarHarness siempre contesta algo, aunque sea desconocido")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// EL TOPE
// ────────────────────────────────────────────────────────────────────────────

// Una línea recortada sigue siendo una línea, y un hueco no. Lo que se comprueba
// acá es que el recorte pase Y que la línea exista igual.
func TestElTopeRecortaYNoTiraLaLinea(t *testing.T) {
	sinArnes(t)
	raiz := proyecto(t, estadoConFeature)

	Abrir(raiz, "new", []string{strings.Repeat("x", 8000)})
	Cerrar(0)

	b, err := os.ReadFile(filepath.Join(raiz, Archivo))
	if err != nil {
		t.Fatal(err)
	}
	if len(b) > Tope {
		t.Errorf("la línea salió de %d bytes, el tope es %d", len(b), Tope)
	}

	es := leerLineas(t, raiz)
	if len(es) != 1 {
		t.Fatalf("la línea se perdió: hay %d", len(es))
	}
	if !es[0].Trunco {
		t.Error("recortó y no marcó `trunco`")
	}
	if es[0].Cmd != "new" {
		t.Errorf("cmd = %q: el recorte se llevó puesto lo que no debía", es[0].Cmd)
	}
}

// El orden del recorte tampoco es casual: las FALLAS son lo último que se toca,
// porque son el motivo por el que el estado no se movió.
func TestLasFallasSonLoUltimoQueSeRecorta(t *testing.T) {
	sinArnes(t)
	raiz := proyecto(t, estadoConFeature)

	Abrir(raiz, "done", []string{"--msg", strings.Repeat("y", 5000)})
	AnotarVeredicto("implementar", []string{"falta el test del lote 2"}, nil, false)
	Cerrar(1)

	es := leerLineas(t, raiz)
	if len(es) != 1 {
		t.Fatalf("quería 1 línea, hay %d", len(es))
	}
	if len(es[0].Fallas) == 0 {
		t.Fatal("se recortaron las fallas antes que los args")
	}
	if es[0].Fallas[0] != "falta el test del lote 2" {
		t.Errorf("la falla llegó cambiada: %q", es[0].Fallas[0])
	}
	if es[0].Args != nil {
		t.Error("no recortó los args, que es lo primero que se va")
	}
}

// ────────────────────────────────────────────────────────────────────────────
// LEER Y CONTAR
// ────────────────────────────────────────────────────────────────────────────

// El archivo lo escriben procesos distintos que pueden morirse en el medio.
// Negarse a leer las buenas por una mala sería tirar justo la evidencia que uno
// vino a buscar.
func TestUnaLineaRotaNoRompeLaLectura(t *testing.T) {
	raiz := proyecto(t, "")
	ruta := filepath.Join(raiz, Archivo)
	if err := os.MkdirAll(filepath.Dir(ruta), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(ruta, []byte(
		`{"cmd":"next","v":1}`+"\n"+
			`{esto no es json`+"\n"+
			`{"cmd":"done","v":1}`+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	es, err := Leer(raiz)
	if err != nil {
		t.Fatal(err)
	}
	if len(es) != 2 {
		t.Fatalf("leyó %d entradas, quería 2 (saltea la rota)", len(es))
	}
}

// Sin archivo no hay error: es un proyecto donde sf todavía no corrió.
func TestSinArchivoLeerNoEsUnError(t *testing.T) {
	es, err := Leer(proyecto(t, ""))
	if err != nil || es != nil {
		t.Errorf("Leer sin archivo → %v, %v", es, err)
	}
}

// El conteo es el punto del paquete: hoy "¿cuántas vueltas dio la revisión?" lo
// contesta el campo que escribe el propio sf-check. Acá lo cuenta sf.
func TestVueltasCuentaLasEntradasYDeDondeVinieron(t *testing.T) {
	es := []Entrada{
		// El camino normal: entra a implementar desde planificación.
		{Antes: Foto{Feature: "f-1", Estado: "planificacion"}, Despues: Foto{Feature: "f-1", Estado: "implementar"}},
		{Antes: Foto{Feature: "f-1", Estado: "implementar"}, Despues: Foto{Feature: "f-1", Estado: "revision"}},
		// Y el bucle: dos vueltas de revisión → implementar → revisión.
		{Antes: Foto{Feature: "f-1", Estado: "revision"}, Despues: Foto{Feature: "f-1", Estado: "implementar"}},
		{Antes: Foto{Feature: "f-1", Estado: "implementar"}, Despues: Foto{Feature: "f-1", Estado: "revision"}},
		{Antes: Foto{Feature: "f-1", Estado: "revision"}, Despues: Foto{Feature: "f-1", Estado: "implementar"}},
		{Antes: Foto{Feature: "f-1", Estado: "implementar"}, Despues: Foto{Feature: "f-1", Estado: "revision"}},
		// Un `sf next` no mueve nada: no puede contar como vuelta.
		{Antes: Foto{Feature: "f-1", Estado: "revision"}, Despues: Foto{Feature: "f-1", Estado: "revision"}},
	}

	vs := Vueltas(es)
	por := map[string]Vuelta{}
	for _, v := range vs {
		por[v.Estado] = v
	}

	if got := por["revision"].Entradas; got != 3 {
		t.Errorf("entradas a revision = %d, quería 3", got)
	}
	if got := por["implementar"].Entradas; got != 3 {
		t.Errorf("entradas a implementar = %d, quería 3", got)
	}
	// La mitad que importa: DOS de las tres entradas a implementar vinieron del
	// bucle de revisión, y una del camino normal.
	if got := por["implementar"].Desde["revision"]; got != 2 {
		t.Errorf("entradas a implementar desde revision = %d, quería 2", got)
	}
	if got := por["implementar"].Desde["planificacion"]; got != 1 {
		t.Errorf("entradas a implementar desde planificacion = %d, quería 1", got)
	}
}

func TestFiltrarPorFeatureYComando(t *testing.T) {
	es := []Entrada{
		{Cmd: "next", Antes: Foto{Feature: "f-1"}},
		{Cmd: "done", Antes: Foto{Feature: "f-1"}},
		{Cmd: "done", Antes: Foto{Feature: "f-2"}},
	}
	if got := len(Filtrar(es, Filtro{Feature: "f-1"})); got != 2 {
		t.Errorf("por feature: %d, quería 2", got)
	}
	if got := len(Filtrar(es, Filtro{Cmd: "done"})); got != 2 {
		t.Errorf("por comando: %d, quería 2", got)
	}
	if got := len(Filtrar(es, Filtro{Ultimas: 1})); got != 1 {
		t.Errorf("por últimas: %d, quería 1", got)
	}
}

// ────────────────────────────────────────────────────────────────────────────
// LO QUE ANOTAN LOS COMANDOS
// ────────────────────────────────────────────────────────────────────────────

// El estado de un paso de PRODUCTO no vive en ningún campo del estado.json —lo
// deduce la máquina—, así que sin esto una línea del ⑦ no diría en qué paso
// estaba.
func TestAnotarPasoLlenaElEstadoQueLaFotoNoPuedeVer(t *testing.T) {
	sinArnes(t)
	raiz := proyecto(t, `{"producto":{"brief_sellado":"hacelo"},"feature_actual":"","features":{}}`)

	Abrir(raiz, "next", nil)
	AnotarPaso(Paso{Tipo: "trabajar", Estado: "prd", Skill: "sfp-po",
		Perfil: "razonar", Modelo: "un-modelo", Via: "subagente"})
	Cerrar(0)

	es := leerLineas(t, raiz)
	if len(es) != 1 {
		t.Fatalf("quería 1 línea, hay %d", len(es))
	}
	e := es[0]
	// Va en Entrada.Estado y NO en la Foto: la foto lee el estado.json, donde
	// "prd" no existe como valor de ningún campo.
	if e.Estado != "prd" {
		t.Errorf("estado = %q, quería prd", e.Estado)
	}
	if e.Antes.Estado != "" {
		t.Errorf("AnotarPaso pisó la foto: antes.estado = %q", e.Antes.Estado)
	}
	if e.Skill != "sfp-po" || e.Via != "subagente" {
		t.Errorf("skill/via = %q/%q", e.Skill, e.Via)
	}
	// Perfil y modelo van los DOS: son "qué pide el paso" y "qué se le dio", y
	// con una sola palabra las dos preguntas tienen la misma respuesta.
	if e.Perfil != "razonar" || e.Modelo != "un-modelo" {
		t.Errorf("perfil/modelo = %q/%q", e.Perfil, e.Modelo)
	}
}

// Anotar sin haber abierto no puede explotar: `sf` puede fallar al leer el cwd y
// no abrir nunca la entrada, y los comandos anotan igual.
func TestAnotarSinAbrirNoExplota(t *testing.T) {
	enCurso = nil
	AnotarPaso(Paso{Skill: "x"})
	AnotarVeredicto("x", []string{"x"}, nil, false)
	AnotarLanzamiento("l-1")
	Cerrar(0)
}

// Y un Cerrar repetido no puede escribir dos líneas para la misma invocación.
func TestCerrarDosVecesEscribeUnaSolaLinea(t *testing.T) {
	sinArnes(t)
	raiz := proyecto(t, estadoConFeature)

	Abrir(raiz, "status", nil)
	Cerrar(0)
	Cerrar(0)

	if es := leerLineas(t, raiz); len(es) != 1 {
		t.Errorf("hay %d líneas, quería 1", len(es))
	}
}

// El conteo tiene que ver los tramos de PRODUCTO, que son los primeros que se
// van a correr y donde no hay estado de feature que mirar. Lo que se mueve ahí
// es un sello, y se cuenta comparándolo — nunca derivándolo del orden del flujo.
func TestVueltasVeLasTransicionesDeProducto(t *testing.T) {
	es := []Entrada{
		// El sello del ⑥.
		{Cmd: "approve",
			Antes:   Foto{Producto: Producto{}},
			Despues: Foto{Producto: Producto{Brief: "hacelo"}}},
		// El ⑦ anota el prd_hash y no para.
		{Cmd: "done",
			Antes:   Foto{Producto: Producto{Brief: "hacelo"}},
			Despues: Foto{Producto: Producto{Brief: "hacelo", Prd: "a3f9c1"}}},
		// Un `sf next` en el medio no mueve ningún sello.
		{Cmd: "next",
			Antes:   Foto{Producto: Producto{Brief: "hacelo", Prd: "a3f9c1"}},
			Despues: Foto{Producto: Producto{Brief: "hacelo", Prd: "a3f9c1"}}},
	}

	por := map[string]Vuelta{}
	for _, v := range Vueltas(es) {
		if v.Feature != "" {
			t.Errorf("un sello de producto quedó atribuido a la feature %q", v.Feature)
		}
		por[v.Estado] = v
	}
	if por["brief"].Entradas != 1 {
		t.Errorf("el sello del brief se contó %d veces, quería 1", por["brief"].Entradas)
	}
	if por["prd"].Entradas != 1 {
		t.Errorf("el prd_hash se contó %d veces, quería 1", por["prd"].Entradas)
	}
	if len(por) != 2 {
		t.Errorf("contó %d sellos, quería 2: el `sf next` no mueve ninguno", len(por))
	}
}
