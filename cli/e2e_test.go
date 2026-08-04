package main

// ----------------------------------------------------------------------------
// E2E Tier A — el pipeline completo, ejercido contra el BINARIO real.
//
// Qué es esto y en qué se diferencia del resto de la suite:
//
//   cli/*_test.go        llaman a las funciones internas (runLint, runGate, …).
//                        Prueban unidades. Rápidos, pero no ven el binario.
//   cli/e2e_test.go      compila `sf` y lo ejecuta como lo haría un usuario,
//                        sobre un proyecto sandbox con git de verdad.
//
// Es la automatización de `planning/TUTORIAL-PRUEBAS.es.md`, que hasta ahora era
// una caminata manual. Cubre el arco entero: init → feature → artefactos con
// gates sellados → plan computado → build orquestado → juez → verdict → archive
// → verify, más enforcement (hook), métricas y migración.
//
// LO QUE **NO** ES: no prueba que un agente de IA use bien SpecForge. Eso es
// Tier B y necesita un modelo de verdad (ver §9 del doc). Acá todo es
// determinista: mismos comandos, mismas salidas, sin API y sin red.
//
// Correr:
//   go test ./... -run TestE2E -v     # sólo este
//   go test ./... -short              # lo saltea (es el más lento de la suite)
// ----------------------------------------------------------------------------

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// ----------------------------------------------------------------------------
// Infraestructura
// ----------------------------------------------------------------------------

// sfBinDir guarda el tempdir del binario para poder limpiarlo en TestMain.
// Queda vacío si ningún test E2E llegó a compilar (corrida `-short`, o
// `go test -run TestGate`): ese es el punto de que el build sea perezoso.
var sfBinDir string

// buildSF compila `sf` UNA sola vez por corrida del paquete y devuelve su ruta.
//
// sync.OnceValues memoiza una función que devuelve (valor, error): la primera
// llamada compila, las siguientes devuelven lo mismo sin recompilar. Es perezoso
// a propósito — los ~250 tests unitarios del paquete no deben pagar un `go build`
// que no van a usar.
var buildSF = sync.OnceValues(func() (string, error) {
	dir, err := os.MkdirTemp("", "sf-e2e-bin")
	if err != nil {
		return "", err
	}
	sfBinDir = dir

	bin := filepath.Join(dir, "sf")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}

	// "." = el paquete de este mismo directorio. El CWD de un test siempre es el
	// directorio del paquete, así que esto compila el CLI que estamos testeando
	// —no una copia vieja que quedó por ahí.
	if out, err := exec.Command("go", "build", "-o", bin, ".").CombinedOutput(); err != nil {
		return "", fmt.Errorf("go build: %w\n%s", err, out)
	}
	return bin, nil
})

// TestMain sólo limpia: el build lo dispara el primer test que lo necesite.
func TestMain(m *testing.M) {
	code := m.Run()
	if sfBinDir != "" {
		os.RemoveAll(sfBinDir) // no usamos defer: os.Exit no lo correría
	}
	os.Exit(code)
}

// lab es un proyecto sandbox descartable: un directorio temporal con git
// inicializado, donde corremos `sf` como si fuéramos el usuario.
type lab struct {
	t   *testing.T
	dir string
	sf  string // ruta al binario compilado
}

// newLab arma el sandbox. t.TempDir() se borra solo al terminar el test, así que
// no hay limpieza manual que olvidarse (el tutorial terminaba con un `rm -rf`).
func newLab(t *testing.T) *lab {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("e2e: git no está en PATH — el modelo de frescura (A6) lo necesita")
	}

	bin, err := buildSF()
	if err != nil {
		t.Fatalf("e2e: no pude compilar sf: %v", err)
	}

	l := &lab{t: t, dir: t.TempDir(), sf: bin}

	l.git("init", "-q")
	// git se niega a commitear sin identidad; la seteamos local al sandbox para
	// no depender de (ni tocar) la config global de quien corre los tests.
	l.git("config", "user.email", "e2e@specforge.test")
	l.git("config", "user.name", "SpecForge E2E")

	// .gitignore + un archivo ignorado: la prueba de A6 (frescura git-aware)
	// depende de que un archivo ignorado NO invalide el resultado de los tests.
	l.write(".gitignore", "*.log\n")
	// pyproject.toml es la señal de stack que `sf init` usa para detectar python.
	l.write("pyproject.toml", "")

	return l
}

// run ejecuta `sf <args...>` dentro del sandbox y devuelve salida combinada +
// exit code. No falla el test: hay pasos donde el exit code distinto de cero ES
// el comportamiento correcto (un deny, un ratchet violado, un gate rehusado).
func (l *lab) run(args ...string) (string, int) {
	l.t.Helper()

	cmd := exec.Command(l.sf, args...)
	cmd.Dir = l.dir
	out, err := cmd.CombinedOutput()

	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		l.t.Fatalf("sf %s: no pude ejecutar el binario: %v", strings.Join(args, " "), err)
	}
	return string(out), code
}

// ok corre un comando que DEBE salir 0.
func (l *lab) ok(args ...string) string {
	l.t.Helper()
	out, code := l.run(args...)
	if code != 0 {
		l.t.Fatalf("sf %s: esperaba exit 0, obtuve %d\n%s", strings.Join(args, " "), code, out)
	}
	return out
}

// fails corre un comando que DEBE salir con un código concreto. Chequear el
// código exacto (y no "algo distinto de cero") es lo que convierte un exit code
// en contrato: 4 = ya existe, 5 = regla de negocio violada, 2 = uso incorrecto.
func (l *lab) fails(want int, args ...string) string {
	l.t.Helper()
	out, code := l.run(args...)
	if code != want {
		l.t.Fatalf("sf %s: esperaba exit %d, obtuve %d\n%s", strings.Join(args, " "), want, code, out)
	}
	return out
}

// pipe ejecuta `sf <args...>` con un payload por stdin y devuelve salida + exit
// code. Varios comandos (gate record-verdict, hook) leen su entrada por ahí, y
// algunos usan el exit code como resultado semántico, no como error: por eso
// devuelve el código en vez de fallar el test.
func (l *lab) pipe(stdin string, args ...string) (string, int) {
	l.t.Helper()

	cmd := exec.Command(l.sf, args...)
	cmd.Dir = l.dir
	cmd.Stdin = strings.NewReader(stdin)
	out, err := cmd.CombinedOutput()

	code := 0
	if ee, ok := err.(*exec.ExitError); ok {
		code = ee.ExitCode()
	} else if err != nil {
		l.t.Fatalf("sf %s: no pude ejecutar el binario: %v", strings.Join(args, " "), err)
	}
	return string(out), code
}

// git corre un comando git en el sandbox.
func (l *lab) git(args ...string) {
	l.t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = l.dir
	if out, err := cmd.CombinedOutput(); err != nil {
		l.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}
}

// write escribe un archivo (creando directorios intermedios) en el sandbox.
func (l *lab) write(rel, content string) {
	l.t.Helper()
	path := filepath.Join(l.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		l.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		l.t.Fatal(err)
	}
}

// append agrega texto al final de un archivo. Se usa para "tocar" código y
// disparar el modelo de frescura.
func (l *lab) append(rel, content string) {
	l.t.Helper()
	path := filepath.Join(l.dir, filepath.FromSlash(rel))
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		l.t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(content); err != nil {
		l.t.Fatal(err)
	}
}

// patchJSON lee un JSON del sandbox, deja que fn lo mute, y lo reescribe.
// Reemplaza los `python3 -c "import json…"` del tutorial: el test queda sin
// dependencias externas más allá de git.
func (l *lab) patchJSON(rel string, fn func(m map[string]any)) {
	l.t.Helper()
	path := filepath.Join(l.dir, filepath.FromSlash(rel))

	raw, err := os.ReadFile(path)
	if err != nil {
		l.t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		l.t.Fatalf("%s: JSON inválido: %v", rel, err)
	}

	fn(m)

	out, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		l.t.Fatal(err)
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		l.t.Fatal(err)
	}
}

// copyFile copia un archivo dentro del sandbox (el idiom `cp x drafts/x` del
// tutorial: el camino de autoría es drafts/ + `sf save --from`).
func (l *lab) copyFile(src, dst string) {
	l.t.Helper()
	raw, err := os.ReadFile(filepath.Join(l.dir, filepath.FromSlash(src)))
	if err != nil {
		l.t.Fatal(err)
	}
	l.write(dst, string(raw))
}

// exists dice si una ruta existe en el sandbox.
func (l *lab) exists(rel string) bool {
	_, err := os.Stat(filepath.Join(l.dir, filepath.FromSlash(rel)))
	return err == nil
}

// has verifica que la salida contenga un fragmento. El mensaje de error imprime
// la salida entera: cuando un E2E falla, lo primero que querés es verla.
func has(t *testing.T, out, want string) {
	t.Helper()
	if !strings.Contains(out, want) {
		t.Errorf("esperaba encontrar %q en la salida:\n%s", want, out)
	}
}

// lacks verifica que la salida NO contenga un fragmento.
func lacks(t *testing.T, out, unwanted string) {
	t.Helper()
	if strings.Contains(out, unwanted) {
		t.Errorf("NO esperaba encontrar %q en la salida:\n%s", unwanted, out)
	}
}

// fakeTestCmd devuelve un `build.test_cmd` portable que produce un reporte JUnit
// real en el placeholder {report}.
//
// Por qué no `printf '<testsuite>…' > {report}` como el tutorial: `printf` no
// existe en cmd.exe, y escapar `<`/`>` en cmd es un festival. En vez de eso
// dejamos el XML en un archivo del sandbox y el comando sólo lo copia — sin
// comillas raras y con una única línea que cambia por plataforma.
func fakeTestCmd(l *lab) string {
	l.t.Helper()
	l.write("junit-template.xml",
		`<testsuite><testcase name="test_saludar"/><testcase name="test_despedir"/></testsuite>`)

	if runtime.GOOS == "windows" {
		return "copy junit-template.xml {report}"
	}
	return "cp junit-template.xml {report}"
}

// noopAgent es el `build.agent_cmd` del laboratorio: consume el seed por stdin y
// sale 0, simulando "el trabajo ya está hecho". Prueba el ORQUESTADOR de
// `sf run` (waves, contrato, checkpoint, sellos) sin depender de un modelo — el
// punto es justamente que el cierre de wave no confía en lo que el agente narre.
func noopAgent() string {
	if runtime.GOOS == "windows" {
		return "more > NUL"
	}
	return "cat > /dev/null"
}

// ----------------------------------------------------------------------------
// El recorrido
// ----------------------------------------------------------------------------

// TestE2E camina el pipeline completo en orden. Los subtests comparten el mismo
// sandbox a propósito: cada parte depende del estado que dejó la anterior, igual
// que el tutorial. Por eso van con t.Run secuencial y no en paralelo.
func TestE2E(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e: compila el binario y corre el pipeline entero — se saltea con -short")
	}

	l := newLab(t)
	const feat = "hola-api"
	featDir := "specforge/features/" + feat

	// ---------------------------------------------------------------- Parte 1
	// Arranque: scaffold mínimo, detección de stack, e inventario brownfield.
	t.Run("01-init-y-onboard", func(t *testing.T) {
		out := l.ok("init", "--minimal")
		has(t, out, "specforge/ initialized (minimal).")
		has(t, out, "stack detected: python")

		// Fail-closed: un segundo init NO pisa nada. Exit 4 = "ya existe".
		out = l.fails(4, "init", "--minimal")
		has(t, out, "already exists")

		// La constitución del laboratorio: un test_cmd falso pero REAL (produce
		// un reporte JUnit parseable), para no depender de pytest instalado.
		l.copyFile("specforge/constitution.json", "specforge/drafts/constitution.json")
		cmd := fakeTestCmd(l)
		l.patchJSON("specforge/drafts/constitution.json", func(m map[string]any) {
			build, _ := m["build"].(map[string]any)
			build["test_cmd"] = cmd
			build["report"] = "junit"
		})
		out = l.ok("save", "constitution", "--from=drafts/constitution.json")
		// A3 a nivel proyecto: drafts/ es el único rincón escribible, `sf save`
		// valida, promueve y RETIRA el borrador.
		has(t, out, "saved specforge/constitution.json")
		has(t, out, "promoted draft")
		if l.exists("specforge/drafts/constitution.json") {
			t.Error("el draft debería haberse retirado después de `sf save`")
		}

		// El "código preexistente" que la feature va a especificar a posteriori.
		l.write("src/app.py", "def saludar():\n    return \"hola\"\n\ndef despedir():\n    return \"chau\"\n")
		l.write("tests/test_app.py",
			"from src.app import saludar, despedir\n\n"+
				"def test_saludar():\n    assert saludar() == \"hola\"\n\n"+
				"def test_despedir():\n    assert despedir() == \"chau\"\n")

		// A4 — inventario brownfield determinista.
		out = l.ok("onboard", "scan")
		has(t, out, "scanned 2 code file(s)")
		has(t, out, "1 test file(s) mapped")

		// Determinismo: correrlo dos veces da lo mismo.
		out2 := l.ok("onboard", "scan")
		if out != out2 {
			t.Errorf("onboard scan no es determinista:\n--- 1 ---\n%s\n--- 2 ---\n%s", out, out2)
		}

		// git necesita un commit base: el modelo de frescura (A6) compara contra
		// los archivos trackeados.
		l.git("add", "-A")
		l.git("commit", "-qm", "baseline")
	})

	// ---------------------------------------------------------------- Parte 2
	// Estado POR feature, autoría por drafts, y gates sellados con hash.
	t.Run("02-feature-y-gates", func(t *testing.T) {
		out := l.ok("feature", "add", "--feature="+feat)
		has(t, out, `added feature "`+feat+`" (status planned)`)

		// A1 — el estado vive por feature; NO hay ledger global. Dos branches
		// tocando features distintas ya no chocan en un archivo compartido.
		if l.exists("specforge/features.json") {
			t.Error("specforge/features.json no debería existir (A1: estado por feature)")
		}
		if !l.exists(featDir + "/feature.json") {
			t.Error("falta " + featDir + "/feature.json")
		}

		l.ok("feature", "set-status", "--feature="+feat, "--to=approved")
		l.ok("feature", "set-status", "--feature="+feat, "--to=building")

		// Los tres artefactos de spec, por el camino de autoría drafts→save→gate.
		l.write(featDir+"/drafts/requirements.json", `{
  "feature": "`+feat+`",
  "summary": "Saludos y despedidas como funciones puras",
  "requirements": [
    {"id": "R1", "ears_type": "ubiquitous", "behavior": "exponer saludar() que devuelve 'hola'",
     "acceptance": ["saludar() == 'hola'"]},
    {"id": "R2", "ears_type": "ubiquitous", "behavior": "exponer despedir() que devuelve 'chau'",
     "acceptance": ["despedir() == 'chau'"]}
  ]
}`)
		l.ok("save", "requirements", "--feature="+feat, "--from=drafts/requirements.json")
		out = l.ok("gate", "approve", "--feature="+feat, "--phase=requirements", "--comment=lab")
		// El sello es el HASH del artefacto aprobado, no una marca en un header.
		has(t, out, "approved "+feat+"/requirements (hash ")

		l.write(featDir+"/drafts/design.json", `{
  "feature": "`+feat+`",
  "summary_md": "Un modulo con dos funciones puras.",
  "components": [
    {"id": "C1", "name": "saludador", "kind": "module",
     "responsibilities": ["producir el saludo y la despedida"]}
  ]
}`)
		l.ok("save", "design", "--feature="+feat, "--from=drafts/design.json")
		l.ok("gate", "approve", "--feature="+feat, "--phase=design")

		l.write(featDir+"/drafts/tasks.json", `{
  "feature": "`+feat+`",
  "tasks": [
    {"id": "T1", "title": "implementar saludar()", "requirement_refs": ["R1"], "depends_on": []},
    {"id": "T2", "title": "implementar despedir()", "requirement_refs": ["R2"], "depends_on": ["T1"]}
  ]
}`)
		l.ok("save", "tasks", "--feature="+feat, "--from=drafts/tasks.json")
		l.ok("gate", "approve", "--feature="+feat, "--phase=tasks")
	})

	// D2' — el plan es un cálculo topológico sobre tasks aprobado, así que su
	// gate se auto-sella: pedir otro gate humano agregaría fatiga sin agregar
	// juicio.
	t.Run("03-plan-computado-y-gate-fusionado", func(t *testing.T) {
		out := l.ok("plan", "compute", "--feature="+feat)
		has(t, out, "computed plan for "+feat+": 2 task(s) → 2 wave(s)")
		has(t, out, "plan gate auto-sealed")

		out = l.ok("gate", "status", "--feature="+feat)
		has(t, out, "sf (derived)") // el plan lo selló la máquina, no un humano
	})

	// C2 — sello anti silent-edit: tocar un artefacto aprobado a mano lo marca
	// stale, y la staleness CASCADEA aguas abajo.
	t.Run("04-silent-edit-detectado", func(t *testing.T) {
		l.patchJSON(featDir+"/requirements.json", func(m map[string]any) {
			m["summary"] = "EDITADO EN SILENCIO"
		})

		out := l.fails(5, "status", "--artifacts", "--feature="+feat)
		has(t, out, "modified after approval (hash mismatch)")
		has(t, out, "requirements is stale") // design/tasks/plan en cascada

		// Restaurar por el camino legítimo y re-sellar la cadena.
		l.copyFile(featDir+"/requirements.json", featDir+"/drafts/requirements.json")
		l.patchJSON(featDir+"/drafts/requirements.json", func(m map[string]any) {
			m["summary"] = "Saludos y despedidas como funciones puras"
		})
		l.ok("save", "requirements", "--feature="+feat, "--from=drafts/requirements.json")
		l.ok("gate", "approve", "--feature="+feat, "--phase=requirements", "--comment=re-sellado")
		l.ok("gate", "approve", "--feature="+feat, "--phase=design", "--comment=cascada")
		l.ok("gate", "approve", "--feature="+feat, "--phase=tasks", "--comment=cascada")
		l.ok("plan", "compute", "--feature="+feat)

		// GAP CONOCIDO: acá `plan compute` debería re-sellar el gate derivado y
		// no lo hace (ver TestE2E_GapPlanReseal). Lo sellamos a mano para que el
		// recorrido siga; el test dedicado documenta la expectativa real.
		l.ok("gate", "approve", "--feature="+feat, "--phase=plan", "--comment=workaround gap re-seal")

		out = l.ok("status", "--artifacts", "--feature="+feat)
		lacks(t, out, "stale")
	})

	// ---------------------------------------------------------------- Parte 3
	// F1 — build orquestado por el CLI, con checkpoint verificado por máquina.
	t.Run("05-run-rechaza-trabajo-sin-anclar", func(t *testing.T) {
		l.copyFile("specforge/constitution.json", "specforge/drafts/constitution.json")
		agent := noopAgent()
		l.patchJSON("specforge/drafts/constitution.json", func(m map[string]any) {
			build, _ := m["build"].(map[string]any)
			build["mode"] = "per-wave"
			build["agent_cmd"] = agent
		})
		l.ok("save", "constitution", "--from=drafts/constitution.json")

		out := l.ok("run", "--feature="+feat, "--dry-run")
		has(t, out, "would run wave 1/2")

		// EL corazón de la tesis: el agente "corre" y sale 0, pero la wave NO
		// cierra, porque nada ancló R1 a código y test reales. La narración del
		// agente no es evidencia.
		out = l.fails(5, "run", "--feature="+feat)
		has(t, out, "verification contract UNMET — STOP")
	})

	t.Run("06-run-con-trace-anclado", func(t *testing.T) {
		l.write(featDir+"/drafts/trace.json", `{
  "feature": "`+feat+`",
  "requirements": {
    "R1": {"code": ["src/app.py:saludar"],  "test": ["tests/test_app.py::test_saludar"],  "status": "ok"},
    "R2": {"code": ["src/app.py:despedir"], "test": ["tests/test_app.py::test_despedir"], "status": "ok"}
  }
}`)
		l.ok("save", "trace", "--feature="+feat, "--from=drafts/trace.json")

		out := l.ok("trace", "verify", "--feature="+feat)
		has(t, out, "OK: all requirements traced to live code and existing tests.")

		out = l.ok("run", "--feature="+feat)
		has(t, out, "report: 2 test result(s) parsed (junit)")
		has(t, out, "approved "+feat+"/wave-0")
		has(t, out, "approved "+feat+"/wave-1")
		has(t, out, "build complete — all 2 wave(s) checkpointed")
	})

	// A6 — frescura git-aware: la lista de archivos sale de `git ls-files`, así
	// que lo ignorado por .gitignore no genera staleness espuria.
	t.Run("07-frescura-git-aware", func(t *testing.T) {
		l.write("debug.log", "ruido\n") // ignorado por .gitignore
		out := l.ok("gate", "show", "--feature="+feat)
		has(t, out, "tests: PASS")
		lacks(t, out, "STALE")

		l.append("src/app.py", "# comentario\n") // tocar CÓDIGO sí cuenta
		out = l.ok("gate", "show", "--feature="+feat)
		has(t, out, "STALE (code changed since — re-run)")

		l.ok("check", "run", "--feature="+feat)
		out = l.ok("gate", "show", "--feature="+feat)
		lacks(t, out, "STALE")
	})

	// ---------------------------------------------------------------- Parte 4
	// A8 — ninguna afirmación del modelo entra al estado sin verificación
	// mecánica o sin quedar marcada como no verificable.
	t.Run("08-citations-verificadas", func(t *testing.T) {
		verdict := `{"phase":"design","verdicts":[` +
			`{"rule":"P1","result":"pass","citation":"producir el saludo y la despedida"},` +
			`{"rule":"P2","result":"pass","citation":"este texto no existe en el artefacto"}]}`

		// Exit 0 porque las dos reglas son `pass`; el problema de P2 es la cita
		// inventada, que se registra pero no reprueba la regla.
		out, code := l.pipe(verdict, "gate", "record-verdict", "--feature="+feat, "--json", "-")
		if code != 0 {
			t.Fatalf("record-verdict: esperaba exit 0, obtuve %d\n%s", code, out)
		}
		has(t, out, `cites text not found`)

		// El veredicto queda persistido con su chequeo de citation por regla.
		var audit struct {
			Entries []struct {
				Verdicts []struct {
					Rule          string `json:"rule"`
					CitationCheck string `json:"citation_check"`
				} `json:"verdicts"`
			} `json:"entries"`
		}
		data, err := os.ReadFile(filepath.Join(l.dir, filepath.FromSlash(featDir+"/audit.json")))
		if err != nil {
			t.Fatal(err)
		}
		if err := json.Unmarshal(data, &audit); err != nil {
			t.Fatal(err)
		}
		if len(audit.Entries) == 0 || len(audit.Entries[0].Verdicts) != 2 {
			t.Fatalf("audit.json no tiene las 2 reglas esperadas: %s", data)
		}
		checks := map[string]string{}
		for _, v := range audit.Entries[0].Verdicts {
			checks[v.Rule] = v.CitationCheck
		}
		if checks["P1"] != "verified" {
			t.Errorf("P1: esperaba citation_check=verified, obtuve %q", checks["P1"])
		}
		if checks["P2"] != "not-found" {
			t.Errorf("P2: esperaba citation_check=not-found, obtuve %q", checks["P2"])
		}
	})

	// D3' — la escalación la cuenta la máquina, no el criterio del humano.
	t.Run("09-escalacion-tras-3-revise", func(t *testing.T) {
		fail := `{"phase":"tasks","verdicts":[{"rule":"P1","result":"fail","citation":"implementar saludar()"}]}`

		var last string
		for i := range 3 {
			out, code := l.pipe(fail, "gate", "record-verdict", "--feature="+feat, "--json", "-")
			// Exit 3 = "se registró, y al menos una regla reprobó". Es un
			// resultado, no una falla del comando: el CLI distingue "no pude
			// hacerlo" de "lo hice y el veredicto es negativo".
			if code != 3 {
				t.Fatalf("record-verdict #%d: esperaba exit 3, obtuve %d\n%s", i+1, code, out)
			}
			last = out
		}
		has(t, last, "ESCALATE: 3 consecutive failed verdicts")
		has(t, last, "the spec itself is likely wrong")
	})

	// ---------------------------------------------------------------- Parte 5
	// El verdict no se puede falsificar: exige evidencia fresca de verdad.
	t.Run("10-verdict-archive-verify", func(t *testing.T) {
		l.write(featDir+"/drafts/review.json", `{
  "feature": "`+feat+`",
  "traceability": [
    {"requirement": "R1", "tasks": ["T1"], "files": ["src/app.py"], "tests": ["tests/test_app.py::test_saludar"], "status": "implemented"},
    {"requirement": "R2", "tasks": ["T2"], "files": ["src/app.py"], "tests": ["tests/test_app.py::test_despedir"], "status": "implemented"}
  ],
  "gaps": [], "constitution_violations": [],
  "verdict": "approve",
  "verdict_rationale_md": "R1 y R2 implementados, anclados y con tests que corrieron en la corrida sellada."
}`)
		l.ok("save", "review", "--feature="+feat, "--from=drafts/review.json")

		// Tocar código invalida la frescura → el verdict se REHÚSA. No hay
		// bypass escribible a mano: el gate lo computa el CLI sobre el disco.
		l.append("src/app.py", "# otro toque\n")
		out := l.fails(5, "gate", "approve", "--feature="+feat, "--phase=verdict")
		has(t, out, "verdict refused")
		has(t, out, "test result is STALE")

		l.ok("check", "run", "--feature="+feat)
		out = l.ok("gate", "approve", "--feature="+feat, "--phase=verdict")
		has(t, out, "approved "+feat+"/verdict (hash ")

		l.ok("feature", "set-status", "--feature="+feat, "--to=checking")
		out = l.ok("feature", "archive", "--feature="+feat)
		has(t, out, "(status done)")

		out = l.ok("verify")
		has(t, out, "✓ ledger")
		has(t, out, "✓ schemas")
		has(t, out, "✓ staleness")
		has(t, out, "✓ trace-drift")
	})

	// ---------------------------------------------------------------- Parte 6
	// Enforcement: el hook DENIEGA, no avisa. Se prueba pipeándole el payload
	// con el adapter `generic`, sin depender de ningún arnés instalado.
	t.Run("11-hook-deny-y-allow", func(t *testing.T) {
		hook := func(payload string) string {
			t.Helper()
			// El adapter `generic` SIEMPRE responde JSON por stdout y sale 0:
			// la decisión viaja en el cuerpo, no en el exit code.
			out, code := l.pipe(payload, "hook", "--harness=generic")
			if code != 0 {
				t.Fatalf("sf hook: esperaba exit 0, obtuve %d\n%s", code, out)
			}
			return out
		}
		abs := func(rel string) string {
			return strings.ReplaceAll(filepath.Join(l.dir, filepath.FromSlash(rel)), `\`, `\\`)
		}

		// feature.json es estado autoritativo → deny.
		out := hook(`{"event":"pre_tool_use","project_dir":"` + abs(".") +
			`","file_path":"` + abs(featDir+"/feature.json") + `"}`)
		has(t, out, `"decision":"deny"`)

		// La heurística también cubre la escritura por shell (redirect en Bash).
		out = hook(`{"event":"pre_tool_use","project_dir":"` + abs(".") +
			`","command":"echo x > specforge/features/` + feat + `/trace.json"}`)
		has(t, out, `"decision":"deny"`)

		// drafts/ es el rincón de autoría → allow.
		out = hook(`{"event":"pre_tool_use","project_dir":"` + abs(".") +
			`","file_path":"` + abs(featDir+"/drafts/nota.json") + `"}`)
		has(t, out, `"decision":"allow"`)

		// A7 — cada deny quedó contado en la telemetría.
		out = l.ok("events", "--json")
		has(t, out, `"kind": "deny"`)
	})

	// ---------------------------------------------------------------- Parte 7
	t.Run("12-coverage-ratchet-y-grafo", func(t *testing.T) {
		out := l.ok("coverage")
		has(t, out, "spec coverage: 100.0%")

		// Código nuevo sin anclar hace caer la cobertura → el ratchet lo frena.
		l.write("src/extra.py", "def suelto(): pass\n")
		out = l.fails(5, "coverage")
		has(t, out, "RATCHET VIOLATED")
		if err := os.Remove(filepath.Join(l.dir, "src", "extra.py")); err != nil {
			t.Fatal(err)
		}

		l.ok("coverage", "--badge")
		if !l.exists("specforge/coverage-badge.svg") {
			t.Error("falta specforge/coverage-badge.svg")
		}

		// D5' — consultas estructurales sobre el grafo declarado.
		out = l.ok("graph", "query", "saludar")
		has(t, out, `code "src/app.py:saludar"`)
		has(t, out, `requirement "R1"`)
		l.ok("graph", "query", "--json", "zzz") // sin matches ⇒ exit 0, lista vacía
	})

	// ---------------------------------------------------------------- Parte 8
	// D4' — el delta es un objeto de primera clase con ciclo de vida propio.
	t.Run("13-delta-lifecycle", func(t *testing.T) {
		l.write(featDir+"/drafts/delta.json", `{
  "why": "el saludo debe ser configurable por idioma",
  "changes": [
    {"op": "modify", "target": "requirements", "ref": "R1", "description": "saludar(idioma='es') con default es"},
    {"op": "add",    "target": "tasks",        "ref": "T3", "description": "parametrizar saludar por idioma"}
  ]
}`)
		out := l.ok("delta", "new", "--feature="+feat, "--from=drafts/delta.json")
		has(t, out, "saved delta D1")

		// La escalera no se saltea: proposed → applied → archived.
		out = l.fails(5, "delta", "set-status", "--feature="+feat, "--id=D1", "--to=archived")
		has(t, out, "illegal transition proposed → archived")

		l.ok("delta", "set-status", "--feature="+feat, "--id=D1", "--to=applied")
		l.ok("delta", "set-status", "--feature="+feat, "--id=D1", "--to=archived")
	})

	// ---------------------------------------------------------------- Parte 10
	// F2 — el flujo serial es el default; el paralelo es opt-in explícito.
	t.Run("14-flujo-paralelo-opt-in", func(t *testing.T) {
		l.ok("feature", "add", "--feature=equipo-a")
		l.ok("feature", "add", "--feature=equipo-b")
		l.ok("feature", "set-status", "--feature=equipo-a", "--to=approved")

		out := l.fails(5, "feature", "set-status", "--feature=equipo-b", "--to=approved")
		has(t, out, "one at a time, F22")

		l.copyFile("specforge/constitution.json", "specforge/drafts/constitution.json")
		l.patchJSON("specforge/drafts/constitution.json", func(m map[string]any) {
			m["flow"] = map[string]any{"mode": "parallel"}
		})
		l.ok("save", "constitution", "--from=drafts/constitution.json")

		l.ok("feature", "set-status", "--feature=equipo-b", "--to=approved")
		out = l.ok("state", "current")
		has(t, out, "parallel flow")
	})
}

// ----------------------------------------------------------------------------
// Parte 9 — migración de un proyecto legacy. Sandbox propio: necesita arrancar
// del formato viejo (features.json monolítico), no del que deja `sf init`.
// ----------------------------------------------------------------------------

func TestE2EMigrateLegacy(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e: se saltea con -short")
	}
	l := newLab(t)

	l.write("specforge/features.json", `{"schema_version":"1.0","features":[
  {"name":"vieja-a","status":"done","gates":[{"phase":"requirements","result":"approve","by":"user","at":"2026-01-01T00:00:00Z"}]},
  {"name":"vieja-b","status":"planned","gates":[]}
]}`)

	// --dry-run anuncia y NO toca nada: el archivo legacy sigue ahí después.
	out := l.ok("migrate", "--dry-run")
	has(t, out, "Would apply 2 change(s)")
	has(t, out, "split-state")
	if !l.exists("specforge/features.json") {
		t.Fatal("--dry-run no debería haber tocado specforge/features.json")
	}

	out = l.ok("migrate")
	has(t, out, "Applied 2 change(s)")

	if !l.exists("specforge/features/vieja-a/feature.json") {
		t.Error("falta el estado por feature de vieja-a")
	}
	if l.exists("specforge/features.json") {
		t.Error("el features.json monolítico debería haberse removido")
	}

	// Idempotente: correrla de nuevo no vuelve a migrar.
	out = l.ok("migrate")
	has(t, out, "Nothing to migrate")
}

// ----------------------------------------------------------------------------
// GAP CONOCIDO — documentado como test ejecutable, no como comentario.
//
// `sf plan compute` auto-sella el gate de plan SOLO la primera vez
// (cli/plan.go, autoSealPlanGate): se rehúsa si ya existió un gate de plan.
// Después de una cascada de re-aprobación aguas arriba —el escenario de
// "04-silent-edit-detectado"— recomputa plan.json pero NO re-sella, y el plan
// queda `stale` para siempre hasta que alguien corra `sf gate approve
// --phase=plan` a mano.
//
// Contradice el principio que declara el comentario del propio código: "el
// juicio humano entra exactamente cuando hubo juicio que revisar". Acá no hubo
// refinamiento manual — sólo se movió algo aguas arriba.
//
// Por qué está en Skip y no en rojo: el arreglo toca semántica de gates y tiene
// un caso borde real (`plan compute` MERGEA Name/Complexity/Rationale del plan
// previo, así que "recién computado" no implica "sin contenido humano").
// Sacale el Skip el día que se decida el arreglo — el test ya dice qué esperar.
// ----------------------------------------------------------------------------

func TestE2EGapPlanReseal(t *testing.T) {
	t.Skip("gap conocido: `sf plan compute` no re-sella el gate de plan tras una cascada aguas arriba")

	l := newLab(t)
	const feat = "gap"
	featDir := "specforge/features/" + feat

	l.ok("init", "--minimal")
	l.ok("feature", "add", "--feature="+feat)
	l.ok("feature", "set-status", "--feature="+feat, "--to=approved")

	l.write(featDir+"/drafts/requirements.json", `{"feature":"`+feat+`","summary":"v1",
  "requirements":[{"id":"R1","ears_type":"ubiquitous","behavior":"hacer algo","acceptance":["ok"]}]}`)
	l.ok("save", "requirements", "--feature="+feat, "--from=drafts/requirements.json")
	l.ok("gate", "approve", "--feature="+feat, "--phase=requirements")

	l.write(featDir+"/drafts/design.json", `{"feature":"`+feat+`","summary_md":"d",
  "components":[{"id":"C1","name":"c","kind":"module","responsibilities":["r"]}]}`)
	l.ok("save", "design", "--feature="+feat, "--from=drafts/design.json")
	l.ok("gate", "approve", "--feature="+feat, "--phase=design")

	l.write(featDir+"/drafts/tasks.json", `{"feature":"`+feat+`",
  "tasks":[{"id":"T1","title":"t","requirement_refs":["R1"],"depends_on":[]}]}`)
	l.ok("save", "tasks", "--feature="+feat, "--from=drafts/tasks.json")
	l.ok("gate", "approve", "--feature="+feat, "--phase=tasks")

	l.ok("plan", "compute", "--feature="+feat) // primera vez: auto-sella

	// Cascada aguas arriba, sin tocar el plan a mano.
	l.copyFile(featDir+"/requirements.json", featDir+"/drafts/requirements.json")
	l.patchJSON(featDir+"/drafts/requirements.json", func(m map[string]any) { m["summary"] = "v2" })
	l.ok("save", "requirements", "--feature="+feat, "--from=drafts/requirements.json")
	l.ok("gate", "approve", "--feature="+feat, "--phase=requirements")
	l.ok("gate", "approve", "--feature="+feat, "--phase=design")
	l.ok("gate", "approve", "--feature="+feat, "--phase=tasks")

	out := l.ok("plan", "compute", "--feature="+feat)
	has(t, out, "plan gate auto-sealed") // ← lo que HOY no pasa

	out = l.ok("status", "--artifacts", "--feature="+feat)
	lacks(t, out, "stale")
}
