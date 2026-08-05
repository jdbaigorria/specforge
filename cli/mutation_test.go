package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// RM-C6 — mutation testing (F8), opt-in.
//
// No se testea que sepamos mutar código: eso lo hace la herramienta del stack.
// Se testea lo que decide si el gate significa algo — que el ALCANCE salga del
// trace (la ventaja que ninguna herramienta genérica tiene), que se consuma el
// EXIT CODE y nunca un score parseado, y que los dos verdes falsos posibles
// —sin comando y sin alcance— RECHACEN en vez de dejar pasar.
// ----------------------------------------------------------------------------

func makeMutationProject(t *testing.T, trace string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"f","status":"checking","gates":[]}]}`)
	writeFile(t, filepath.Join(dir, "specforge/features/f/trace.json"),
		`{"feature":"f","requirements":`+trace+`}`)
	return dir
}

func mutationConstitution(t *testing.T, dir, cmd string, require bool) {
	t.Helper()
	req := "false"
	if require {
		req = "true"
	}
	writeFile(t, filepath.Join(dir, "specforge/constitution.json"),
		`{"identity_md":"x","build":{"mutation_cmd":`+jsonString(cmd)+`},
		  "verification":{"require_mutation":`+req+`}}`)
}

func TestMutationScope(t *testing.T) {
	t.Run("sale del trace, deduplicado y ordenado", func(t *testing.T) {
		// Ordenado importa: dos corridas con el mismo estado tienen que producir
		// el mismo comando, o el mismo proyecto daría resultados distintos sin
		// que nada haya cambiado.
		dir := makeMutationProject(t,
			`{"R1":{"code":["src/b.go:Foo","src/a.go:Bar"],"test":[],"status":"ok"},
			  "R2":{"code":["src/a.go:Baz"],"test":[],"status":"ok"}}`)

		got := mutationScope(dir, "f")
		if len(got) != 2 || got[0] != "src/a.go" || got[1] != "src/b.go" {
			t.Errorf("scope = %v, want [src/a.go src/b.go]", got)
		}
	})

	t.Run("los tests no se mutan", func(t *testing.T) {
		// Mutar un test y ver que "falla" no dice nada sobre si el suite detecta
		// defectos del producto.
		dir := makeMutationProject(t,
			`{"R1":{"code":["src/a.go:Foo","src/a_test.go:TestFoo"],"test":[],"status":"ok"}}`)

		got := mutationScope(dir, "f")
		if len(got) != 1 || got[0] != "src/a.go" {
			t.Errorf("scope = %v, want sólo [src/a.go]", got)
		}
	})

	t.Run("sin trace el alcance es vacío, no un error", func(t *testing.T) {
		dir := t.TempDir()
		if got := mutationScope(dir, "f"); len(got) != 0 {
			t.Errorf("scope = %v, want vacío", got)
		}
	})
}

func TestShellQuote(t *testing.T) {
	// Un archivo con espacio partiría el comando en dos argumentos y el mutador
	// correría sobre rutas inexistentes: VERDE POR LA RAZÓN EQUIVOCADA, que es
	// el peor resultado posible en un gate.
	cases := []struct{ in, wantContains string }{
		{"src/a.go", "src/a.go"},
		{"src/my file.go", "my file.go"},
	}
	for _, c := range cases {
		got := shellQuote(c.in)
		if !strings.Contains(got, c.wantContains) {
			t.Errorf("shellQuote(%q) = %q", c.in, got)
		}
	}
	if shellQuote("src/a.go") != "src/a.go" {
		t.Error("una ruta sin nada raro no debería entrecomillarse")
	}
	if q := shellQuote("src/my file.go"); q == "src/my file.go" {
		t.Error("una ruta con espacio DEBE entrecomillarse")
	}
}

func TestMutationGate(t *testing.T) {
	goodTrace := `{"R1":{"code":["src/a.go:Foo"],"test":[],"status":"ok"}}`

	t.Run("require_mutation ON sin mutation_cmd RECHAZA", func(t *testing.T) {
		dir := makeMutationProject(t, goodTrace)
		mutationConstitution(t, dir, "", true)

		reasons := mutationGateReasons(dir, "f")
		if len(reasons) != 1 || !strings.Contains(reasons[0], "mutation_cmd") {
			t.Fatalf("reasons = %v", reasons)
		}
	})

	t.Run("alcance vacío con {files} RECHAZA — el verde no significaría nada", func(t *testing.T) {
		// Un mutador sobre cero archivos sale verde. Ese verde dice "no se probó
		// nada", no "el suite verifica": es exactamente la confusión que el flag
		// existe para impedir.
		dir := makeMutationProject(t, `{}`)
		mutationConstitution(t, dir, "mutmut run --paths={files}", true)

		reasons := mutationGateReasons(dir, "f")
		if len(reasons) != 1 || !strings.Contains(reasons[0], "nothing to mutate") {
			t.Fatalf("reasons = %v, want el rechazo por alcance vacío", reasons)
		}
	})

	t.Run("exit 0 ⇒ pasa, y el {files} llegó sustituido", func(t *testing.T) {
		// `test -f` sobre el archivo interpolado pasa SÓLO si la sustitución
		// ocurrió: si fallara, el comando vería el literal `{files}`.
		dir := makeMutationProject(t, goodTrace)
		writeFile(t, filepath.Join(dir, "src/a.go"), "package src\n")
		mutationConstitution(t, dir, "test -f {files}", true)

		if reasons := mutationGateReasons(dir, "f"); len(reasons) != 0 {
			t.Errorf("reasons = %v, want ninguna", reasons)
		}
	})

	t.Run("exit != 0 ⇒ el suite no detecta los defectos", func(t *testing.T) {
		dir := makeMutationProject(t, goodTrace)
		mutationConstitution(t, dir, "echo '3 mutants survived'; exit 1", true)

		reasons := mutationGateReasons(dir, "f")
		if len(reasons) != 1 {
			t.Fatalf("reasons = %v", reasons)
		}
		if !strings.Contains(reasons[0], "does not detect") {
			t.Errorf("mensaje = %q", reasons[0])
		}
		// La salida de la herramienta viaja en el motivo: sin ella el usuario
		// tiene un exit code y ninguna pista de qué mutante sobrevivió.
		if !strings.Contains(reasons[0], "survived") {
			t.Errorf("la salida de la herramienta debe viajar: %q", reasons[0])
		}
	})

	t.Run("sin {files} el comando corre tal cual — placeholder OPCIONAL", func(t *testing.T) {
		// Asimetría deliberada con arch_cmd: allá omitir `{config}` invierte la
		// garantía (se verifica contra reglas que nadie aprobó); acá sólo
		// significa "el usuario define su alcance", más lento pero no incorrecto.
		dir := makeMutationProject(t, `{}`) // alcance vacío a propósito
		mutationConstitution(t, dir, "exit 0", true)

		if reasons := mutationGateReasons(dir, "f"); len(reasons) != 0 {
			t.Errorf("reasons = %v — sin {files} el alcance vacío no debe rechazar", reasons)
		}
	})

	t.Run("el flag apagado no corre nada", func(t *testing.T) {
		dir := makeMutationProject(t, goodTrace)
		mutationConstitution(t, dir, "exit 1", false)
		if requireMutationEnabled(dir) {
			t.Error("require_mutation debería estar apagado")
		}
	})
}

func TestMutationConstitutionValidation(t *testing.T) {
	errorsOf := func(cf constitutionFile) (errs, warns []string) {
		var rep report
		checkConstitution(cf, &rep)
		return rep.errors, rep.warnings
	}
	has := func(list []string, want string) bool {
		for _, e := range list {
			if strings.Contains(e, want) {
				return true
			}
		}
		return false
	}

	t.Run("mutation_cmd sin {files} AVISA, no rechaza", func(t *testing.T) {
		// Perderse el alcance por trace es la diferencia entre minutos y horas, y
		// esa es la razón nº1 por la que un gate caro se termina apagando. Pero
		// no es incorrecto, así que es warning y no error.
		errs, warns := errorsOf(constitutionFile{IdentityMD: "x",
			Build: &buildConfig{MutationCmd: "mutmut run"}})
		if has(errs, "mutation_cmd") {
			t.Errorf("no debería ser error: %v", errs)
		}
		if !has(warns, "{files}") {
			t.Errorf("warnings = %v, want el aviso de alcance", warns)
		}
	})

	t.Run("require_mutation sin mutation_cmd se rechaza en la config", func(t *testing.T) {
		errs, _ := errorsOf(constitutionFile{IdentityMD: "x",
			Verification: &verificationConfig{RequireMutation: true}})
		if !has(errs, "require_mutation") {
			t.Errorf("errors = %v", errs)
		}
	})

	t.Run("los dos opt-in son ORTOGONALES", func(t *testing.T) {
		// DEC-3: un proyecto puede querer conformidad de arquitectura (segundos)
		// sin pagar mutación (minutos). Un `lane: strict` los habría atado.
		errs, _ := errorsOf(constitutionFile{IdentityMD: "x",
			Build:        &buildConfig{ArchCmd: "go-arch-lint check --config={config}"},
			Verification: &verificationConfig{RequireArch: true}})
		if len(errs) != 0 {
			t.Errorf("require_arch solo debe validar sin mutation_cmd: %v", errs)
		}
	})
}

func TestMutationCommand(t *testing.T) {
	goodTrace := `{"R1":{"code":["src/a.go:Foo"],"test":[],"status":"ok"}}`

	t.Run("--feature obligatorio", func(t *testing.T) {
		if code := runMutation([]string{"scope"}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
	})

	t.Run("subcomando desconocido", func(t *testing.T) {
		if code := runMutation([]string{"nope"}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
	})

	t.Run("scope lista los archivos", func(t *testing.T) {
		dir := makeMutationProject(t, goodTrace)
		if code := runMutation([]string{"scope", "--feature=f", dir}); code != 0 {
			t.Errorf("exit = %d", code)
		}
	})

	t.Run("run sin mutation_cmd DEGRADA elegante", func(t *testing.T) {
		dir := makeMutationProject(t, goodTrace)
		if code := runMutation([]string{"run", "--feature=f", dir}); code != 0 {
			t.Errorf("exit = %d, want 0 — sin comando se omite, no falla", code)
		}
	})

	t.Run("run propaga el fallo", func(t *testing.T) {
		dir := makeMutationProject(t, goodTrace)
		mutationConstitution(t, dir, "exit 1", true)
		if code := runMutation([]string{"run", "--feature=f", dir}); code != 1 {
			t.Errorf("exit = %d, want 1", code)
		}
	})
}
