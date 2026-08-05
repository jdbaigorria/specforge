package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// RM-C7 — conformidad de arquitectura (F7), opt-in.
//
// Lo que se verifica acá NO es que sepamos analizar imports: eso lo hace la
// herramienta del stack. Lo que se verifica es lo único que SpecForge aporta y
// nadie más puede — que la regla salga del DISEÑO APROBADO y no de un archivo
// paralelo, y que cuando el mapeo no se puede derivar el sistema FALLE PIDIÉNDOLO
// en vez de inventarlo. Un mapeo inventado vuelve decorativa toda la conformidad
// que se verifique después.
// ----------------------------------------------------------------------------

// makeArchProject arma una feature con su design.json y su tasks.json.
func makeArchProject(t *testing.T, design, tasks string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`)
	writeFile(t, filepath.Join(dir, "specforge/features/f/design.json"),
		`{"feature":"f","components":`+design+`}`)
	writeFile(t, filepath.Join(dir, "specforge/features/f/tasks.json"),
		`{"feature":"f","tasks":`+tasks+`}`)
	return dir
}

// archConstitution escribe la constitución con los dos campos de C7.
func archConstitution(t *testing.T, dir, archCmd string, requireArch bool) {
	t.Helper()
	req := "false"
	if requireArch {
		req = "true"
	}
	writeFile(t, filepath.Join(dir, "specforge/constitution.json"),
		`{"identity_md":"x","build":{"arch_cmd":`+jsonString(archCmd)+`},
		  "verification":{"require_arch":`+req+`}}`)
}

func jsonString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}

func TestComponentFilesMapping(t *testing.T) {
	t.Run("deriva el mapeo desde tasks.json", func(t *testing.T) {
		dir := makeArchProject(t,
			`[{"id":"C1","name":"api","depends_on":[]},{"id":"C2","name":"store","depends_on":[]}]`,
			`[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go"]},
			  {"id":"T2","component_refs":["C1"],"files_touched":["src/api_extra.go"]},
			  {"id":"T3","component_refs":["C2"],"files_touched":["src/store.go"]}]`)

		mapping, conflicts := componentFiles(dir, "f")
		if len(conflicts) != 0 {
			t.Fatalf("conflictos inesperados: %v", conflicts)
		}
		if got := mapping["C1"]; len(got) != 2 || got[0] != "src/api.go" || got[1] != "src/api_extra.go" {
			t.Errorf("C1 = %v, want [src/api.go src/api_extra.go] ordenado", got)
		}
		if got := mapping["C2"]; len(got) != 1 || got[0] != "src/store.go" {
			t.Errorf("C2 = %v", got)
		}
	})

	t.Run("los tests NO son parte del componente", func(t *testing.T) {
		// Un test que importa a través de una frontera de componentes es normal
		// —arma un escenario de punta a punta—, así que contarlo produciría
		// violaciones falsas. Mismo criterio que `sf coverage`, que también los
		// saca. Apareció mirando la salida real contra examples/slugify, donde
		// `tests/test_slug.py` figuraba como parte de C1.
		dir := makeArchProject(t,
			`[{"id":"C1","name":"api","depends_on":[]}]`,
			`[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go","src/api_test.go","tests/test_api.py"]}]`)

		mapping, _ := componentFiles(dir, "f")
		if got := mapping["C1"]; len(got) != 1 || got[0] != "src/api.go" {
			t.Errorf("C1 = %v, want sólo [src/api.go]", got)
		}
	})

	t.Run("un archivo reclamado por dos componentes es CONFLICTO, no una elección", func(t *testing.T) {
		// Ni "gana el primero" ni "pertenece a los dos": las dos salidas son
		// inventar un mapeo que nadie declaró.
		dir := makeArchProject(t,
			`[{"id":"C1","name":"api","depends_on":[]},{"id":"C2","name":"store","depends_on":[]}]`,
			`[{"id":"T1","component_refs":["C1"],"files_touched":["src/shared.go"]},
			  {"id":"T2","component_refs":["C2"],"files_touched":["src/shared.go"]}]`)

		_, conflicts := componentFiles(dir, "f")
		if len(conflicts) != 1 || !strings.Contains(conflicts[0], "src/shared.go") {
			t.Fatalf("conflicts = %v, want uno nombrando src/shared.go", conflicts)
		}
		if !strings.Contains(conflicts[0], "C1") || !strings.Contains(conflicts[0], "C2") {
			t.Errorf("el conflicto debe nombrar los DOS componentes: %q", conflicts[0])
		}
	})
}

func TestBuildArchRules(t *testing.T) {
	t.Run("may_depend_on son las aristas DIRECTAS, sin clausura", func(t *testing.T) {
		// C1→C2→C3. Si computáramos la clausura, C1 podría depender de C3 sin que
		// el diseño lo haya aprobado — estaríamos decidiendo por el usuario que
		// su arquitectura es transitiva.
		dir := makeArchProject(t,
			`[{"id":"C1","name":"api","depends_on":["C2"]},
			  {"id":"C2","name":"svc","depends_on":["C3"]},
			  {"id":"C3","name":"store","depends_on":[]}]`,
			`[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go"]},
			  {"id":"T2","component_refs":["C2"],"files_touched":["src/svc.go"]},
			  {"id":"T3","component_refs":["C3"],"files_touched":["src/store.go"]}]`)

		rules, reasons := buildArchRules(dir, "f")
		if len(reasons) != 0 {
			t.Fatalf("reasons = %v", reasons)
		}
		if len(rules.Components) != 3 {
			t.Fatalf("components = %d, want 3", len(rules.Components))
		}
		c1 := rules.Components[0]
		if c1.ID != "C1" {
			t.Fatalf("orden inestable: primero = %s", c1.ID)
		}
		if len(c1.MayDependOn) != 1 || c1.MayDependOn[0] != "C2" {
			t.Errorf("C1.may_depend_on = %v, want sólo [C2] — la clausura la computa la herramienta",
				c1.MayDependOn)
		}
	})

	t.Run("un componente sin archivos FALLA PIDIENDO el mapeo", func(t *testing.T) {
		// §3.8: si el diseño habla de componentes conceptuales sin correlato
		// físico, la generación no es posible y debe fallar. Emitir reglas igual
		// dejaría un componente que ninguna herramienta puede violar — un hueco
		// silencioso justo donde se prometió una garantía.
		dir := makeArchProject(t,
			`[{"id":"C1","name":"api","depends_on":[]},{"id":"C2","name":"fantasma","depends_on":[]}]`,
			`[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go"]}]`)

		_, reasons := buildArchRules(dir, "f")
		if len(reasons) != 1 || !strings.Contains(reasons[0], "C2") {
			t.Fatalf("reasons = %v, want uno nombrando C2", reasons)
		}
		if !strings.Contains(reasons[0], "component_refs") {
			t.Errorf("el mensaje debe decir QUÉ completar, no sólo que falta: %q", reasons[0])
		}
	})

	t.Run("una arista hacia un componente no declarado se rechaza", func(t *testing.T) {
		dir := makeArchProject(t,
			`[{"id":"C1","name":"api","depends_on":["C9"]}]`,
			`[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go"]}]`)

		_, reasons := buildArchRules(dir, "f")
		if len(reasons) != 1 || !strings.Contains(reasons[0], "C9") {
			t.Errorf("reasons = %v, want uno nombrando C9", reasons)
		}
	})

	t.Run("sin componentes no hay nada que chequear", func(t *testing.T) {
		dir := makeArchProject(t, `[]`, `[]`)
		_, reasons := buildArchRules(dir, "f")
		if len(reasons) != 1 {
			t.Errorf("reasons = %v, want uno", reasons)
		}
	})
}

func TestWriteArchRules(t *testing.T) {
	dir := makeArchProject(t,
		`[{"id":"C1","name":"api","depends_on":[]}]`,
		`[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go"]}]`)

	path, reasons := writeArchRules(dir, "f")
	if len(reasons) != 0 {
		t.Fatalf("reasons = %v", reasons)
	}
	// Vive en .state/: estado de máquina, protegido por el hook. No es un
	// artefacto que el humano edite — su fuente es el diseño sellado.
	if !strings.Contains(filepath.ToSlash(path), "specforge/.state/") {
		t.Errorf("path = %q, want dentro de specforge/.state/", path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("no se escribió: %v", err)
	}
	var rules archRules
	if err := json.Unmarshal(data, &rules); err != nil {
		t.Fatalf("json inválido: %v", err)
	}
	if rules.Feature != "f" || len(rules.Components) != 1 {
		t.Errorf("rules = %+v", rules)
	}
	// `[]`, nunca `null`: lo parsea una herramienta de terceros y un campo que a
	// veces es lista y a veces null obliga a manejar los dos casos.
	if !strings.Contains(string(data), `"may_depend_on": []`) {
		t.Errorf("may_depend_on vacío debe serializar como [], no null:\n%s", data)
	}
}

func TestArchGate(t *testing.T) {
	goodDesign := `[{"id":"C1","name":"api","depends_on":[]}]`
	goodTasks := `[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go"]}]`

	t.Run("require_arch ON sin arch_cmd RECHAZA, no degrada", func(t *testing.T) {
		// Mismo criterio que R12 con build.report: aprobar sería peor que
		// bloquear, porque el proyecto creería tener una garantía de conformidad
		// que nunca se evaluó.
		dir := makeArchProject(t, goodDesign, goodTasks)
		archConstitution(t, dir, "", true)

		reasons := archGateReasons(dir, "f")
		if len(reasons) != 1 || !strings.Contains(reasons[0], "arch_cmd") {
			t.Fatalf("reasons = %v, want uno pidiendo build.arch_cmd", reasons)
		}
	})

	t.Run("la herramienta sale 0 ⇒ conforme, y recibe el {config} sustituido", func(t *testing.T) {
		// `test -f {config}` pasa SÓLO si el placeholder se reemplazó por la ruta
		// real del archivo generado. Si la sustitución fallara, el comando vería
		// el literal `{config}`, no lo encontraría y saldría != 0 — así este test
		// cubre la conformidad y el cableado de una sola vez.
		dir := makeArchProject(t, goodDesign, goodTasks)
		archConstitution(t, dir, "test -f {config}", true)

		if reasons := archGateReasons(dir, "f"); len(reasons) != 0 {
			t.Errorf("reasons = %v, want ninguna", reasons)
		}
	})

	t.Run("la herramienta sale != 0 ⇒ violación", func(t *testing.T) {
		// Se consume el EXIT CODE, no se parsea un score: el umbral lo configura
		// la herramienta y parsear su salida nos ataría a su versión.
		dir := makeArchProject(t, goodDesign, goodTasks)
		archConstitution(t, dir, "test -f {config} && { echo 'C1 -> C2 forbidden'; exit 1; }", true)

		reasons := archGateReasons(dir, "f")
		if len(reasons) != 1 {
			t.Fatalf("reasons = %v, want una", reasons)
		}
		if !strings.Contains(reasons[0], "violates the approved design graph") {
			t.Errorf("mensaje = %q", reasons[0])
		}
		if !strings.Contains(reasons[0], "forbidden") {
			t.Errorf("la salida de la herramienta debe viajar en el motivo: %q", reasons[0])
		}
	})

	t.Run("el flag apagado no corre nada", func(t *testing.T) {
		dir := makeArchProject(t, goodDesign, goodTasks)
		// Un comando que fallaría si se corriera: si el flag apagado igual lo
		// ejecutara, este test lo cazaría.
		archConstitution(t, dir, "exit 1 # {config}", false)

		if requireArchEnabled(dir) {
			t.Fatal("require_arch debería estar apagado")
		}
	})

	t.Run("sin constitución el opt-in queda apagado", func(t *testing.T) {
		dir := makeArchProject(t, goodDesign, goodTasks)
		if requireArchEnabled(dir) {
			t.Error("sin constitución, require_arch debe ser false")
		}
	})
}

func TestArchCommand(t *testing.T) {
	goodDesign := `[{"id":"C1","name":"api","depends_on":[]}]`
	goodTasks := `[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go"]}]`

	t.Run("--feature es obligatorio", func(t *testing.T) {
		if code := runArch([]string{"rules"}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
	})

	t.Run("subcomando desconocido", func(t *testing.T) {
		if code := runArch([]string{"nope"}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
	})

	t.Run("rules genera y reporta la ruta", func(t *testing.T) {
		dir := makeArchProject(t, goodDesign, goodTasks)
		if code := runArch([]string{"rules", "--feature=f", dir}); code != 0 {
			t.Errorf("exit = %d, want 0", code)
		}
	})

	t.Run("rules falla cuando el mapeo no se puede derivar", func(t *testing.T) {
		dir := makeArchProject(t,
			`[{"id":"C1","name":"api","depends_on":[]},{"id":"C2","name":"fantasma","depends_on":[]}]`,
			goodTasks)
		if code := runArch([]string{"rules", "--feature=f", dir}); code != 1 {
			t.Errorf("exit = %d, want 1", code)
		}
	})

	t.Run("check sin arch_cmd DEGRADA elegante — a diferencia del gate", func(t *testing.T) {
		// La asimetría es deliberada: correr `sf arch check` a mano es una
		// pregunta legítima ("¿tengo esto cableado?"), no un gate que alguien
		// esté por cruzar. El que bloquea es el veredicto.
		dir := makeArchProject(t, goodDesign, goodTasks)
		if code := runArch([]string{"check", "--feature=f", dir}); code != 0 {
			t.Errorf("exit = %d, want 0 — sin arch_cmd se omite, no falla", code)
		}
	})

	t.Run("check propaga la violación", func(t *testing.T) {
		dir := makeArchProject(t, goodDesign, goodTasks)
		archConstitution(t, dir, "test -f {config} && exit 1", true)
		if code := runArch([]string{"check", "--feature=f", dir}); code != 1 {
			t.Errorf("exit = %d, want 1", code)
		}
	})
}

func TestArchConstitutionValidation(t *testing.T) {
	errorsOf := func(cf constitutionFile) []string {
		var rep report
		checkConstitution(cf, &rep)
		return rep.errors
	}
	has := func(errs []string, want string) bool {
		for _, e := range errs {
			if strings.Contains(e, want) {
				return true
			}
		}
		return false
	}

	t.Run("arch_cmd sin {config} se rechaza", func(t *testing.T) {
		// Sin el placeholder la herramienta corre con SU propio archivo de
		// reglas, escrito a mano: la conformidad se verificaría contra algo que
		// nadie aprobó. Es la garantía dada vuelta, no una config incompleta.
		errs := errorsOf(constitutionFile{IdentityMD: "x",
			Build: &buildConfig{ArchCmd: "go-arch-lint check"}})
		if !has(errs, "{config}") {
			t.Errorf("errors = %v, want uno pidiendo el placeholder", errs)
		}
	})

	t.Run("arch_cmd con {config} valida", func(t *testing.T) {
		errs := errorsOf(constitutionFile{IdentityMD: "x",
			Build: &buildConfig{ArchCmd: "go-arch-lint check --config={config}"}})
		if has(errs, "arch_cmd") {
			t.Errorf("errors = %v, no debería quejarse", errs)
		}
	})

	t.Run("require_arch sin arch_cmd se rechaza en la config, no recién en el gate", func(t *testing.T) {
		errs := errorsOf(constitutionFile{IdentityMD: "x",
			Verification: &verificationConfig{RequireArch: true}})
		if !has(errs, "require_arch") {
			t.Errorf("errors = %v, want uno — si no, la sorpresa llega el día que alguien sella", errs)
		}
	})
}

// TestArchResolvesArchivedFeature — la lección de RM-C2b: los artefactos de una
// feature cerrada viven bajo archive/, y buscarlos sólo en features/ los pierde
// en silencio.
func TestArchResolvesArchivedFeature(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"f","status":"done","gates":[]}]}`)
	writeFile(t, filepath.Join(dir, "specforge/archive/2026-01-01-f/design.json"),
		`{"feature":"f","components":[{"id":"C1","name":"api","depends_on":[]}]}`)
	writeFile(t, filepath.Join(dir, "specforge/archive/2026-01-01-f/tasks.json"),
		`{"feature":"f","tasks":[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go"]}]}`)

	rules, reasons := buildArchRules(dir, "f")
	if len(reasons) != 0 {
		t.Fatalf("reasons = %v — los artefactos viven en archive/, no en features/", reasons)
	}
	if len(rules.Components) != 1 {
		t.Errorf("components = %d, want 1", len(rules.Components))
	}
}
