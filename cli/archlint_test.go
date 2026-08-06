package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// Emisor nativo de `.go-arch-lint.yml`.
//
// El test que vale de verdad es el último: corre `go-arch-lint` DE VERDAD contra
// lo que emitimos. Todo lo de arriba es forma; ése es el único que puede
// desmentirnos, y ya lo hizo una vez — la primera versión emitía
// `mayDependOn: []` para los componentes sin dependencias y la herramienta lo
// rechaza. Ningún test de string lo habría cazado.
// ----------------------------------------------------------------------------

func TestArchLintYAML(t *testing.T) {
	rules := archRules{
		SchemaVersion: "1.0", Feature: "orders",
		Components: []archComponent{
			{ID: "handler", Files: []string{"internal/handler/handler.go"}, MayDependOn: []string{"service"}},
			{ID: "service", Files: []string{"internal/service/service.go"}, MayDependOn: []string{"storage"}},
			{ID: "storage", Files: []string{"internal/storage/storage.go"}, MayDependOn: []string{}},
		},
	}

	t.Run("los archivos se traducen a DIRECTORIOS", func(t *testing.T) {
		// Nuestro mapeo es por archivo; `go-arch-lint` mapea por PAQUETE. La
		// traducción es tomar el directorio, y es la que hace que el dato
		// nuestro sea consumible.
		out, reasons := archLintYAML(rules)
		if len(reasons) != 0 {
			t.Fatalf("reasons = %v", reasons)
		}
		if !strings.Contains(out, "      - internal/handler\n") {
			t.Errorf("falta el directorio de handler:\n%s", out)
		}
		if strings.Contains(out, "handler.go") {
			t.Errorf("no debe emitir archivos, sólo directorios:\n%s", out)
		}
	})

	t.Run("un componente SIN dependencias se OMITE de deps", func(t *testing.T) {
		// go-arch-lint rechaza `mayDependOn: []` — pide una ref o un flag. La
		// única forma de decir "no depende de nadie" es no nombrarlo.
		out, _ := archLintYAML(rules)
		if strings.Contains(out, "mayDependOn: []") {
			t.Errorf("la lista vacía es inválida para go-arch-lint:\n%s", out)
		}
		if strings.Contains(out, "  storage:\n    mayDependOn") {
			t.Errorf("storage no depende de nada: no debe aparecer en deps:\n%s", out)
		}
		// Pero sí tiene que estar declarado como COMPONENTE, o sus imports no se
		// chequean contra nada.
		if !strings.Contains(out, "  storage:\n    in:") {
			t.Errorf("storage debe existir en components:\n%s", out)
		}
	})

	t.Run("dos componentes en un mismo DIRECTORIO se rechazan", func(t *testing.T) {
		// La unidad de go-arch-lint es el paquete: dos componentes en un
		// directorio no se pueden expresar. Emitir igual verificaría una
		// arquitectura que nadie declaró. Es RM-C7b visto desde la herramienta.
		shared := archRules{Feature: "f", Components: []archComponent{
			{ID: "a", Files: []string{"internal/core/a.go"}},
			{ID: "b", Files: []string{"internal/core/b.go"}},
		}}
		_, reasons := archLintYAML(shared)
		if len(reasons) == 0 {
			t.Fatal("dos componentes en internal/core deben rechazarse")
		}
		for _, want := range []string{"internal/core", "PACKAGES", "a", "b"} {
			if !strings.Contains(reasons[0], want) {
				t.Errorf("el motivo debe contener %q: %q", want, reasons[0])
			}
		}
	})

	t.Run("un formato desconocido se rechaza", func(t *testing.T) {
		dir := makeArchProject(t,
			`[{"id":"C1","name":"api","depends_on":[]}]`,
			`[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go"]}]`)
		if _, reasons := writeArchRules(dir, "f", "nope"); len(reasons) == 0 {
			t.Error("un formato desconocido debe rechazarse, no caer en el default")
		}
	})

	t.Run("el default sigue siendo json", func(t *testing.T) {
		// Un proyecto existente no puede cambiar de comportamiento sólo por
		// actualizar el binario.
		dir := makeArchProject(t,
			`[{"id":"C1","name":"api","depends_on":[]}]`,
			`[{"id":"T1","component_refs":["C1"],"files_touched":["src/api.go"]}]`)
		if got := archFormat(dir); got != "json" {
			t.Errorf("archFormat = %q, want json", got)
		}
		path, reasons := writeArchRules(dir, "f", archFormat(dir))
		if len(reasons) != 0 {
			t.Fatalf("reasons = %v", reasons)
		}
		if filepath.Ext(path) != ".json" {
			t.Errorf("path = %q, want .json", path)
		}
	})
}

// TestArchLintAgainstRealTool — el único test que puede desmentirnos.
//
// Se saltea si `go-arch-lint` no está instalado: un test que exige una
// herramienta externa y rompe el build de quien no la tiene se termina
// borrando, y entonces no protege nada. Cuando SÍ está, verifica las dos mitades
// que ningún test de string alcanza: que la herramienta ACEPTE nuestra config, y
// que con ella DETECTE una violación real del grafo aprobado.
func TestArchLintAgainstRealTool(t *testing.T) {
	bin, err := exec.LookPath("go-arch-lint")
	if err != nil {
		if home, e := os.UserHomeDir(); e == nil {
			candidate := filepath.Join(home, ".go", "bin", "go-arch-lint")
			if _, e := os.Stat(candidate); e == nil {
				bin = candidate
			}
		}
	}
	if bin == "" {
		t.Skip("go-arch-lint no está instalado — este test necesita la herramienta real")
	}

	// Fixture Go con tres paquetes y un grafo real: handler → service → storage.
	proj := t.TempDir()
	write := func(rel, content string) {
		p := filepath.Join(proj, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("go.mod", "module example.com/archfx\n\ngo 1.26\n")
	write("internal/storage/storage.go", "package storage\n\nfunc Save(id string) error { return nil }\n")
	write("internal/service/service.go",
		"package service\n\nimport \"example.com/archfx/internal/storage\"\n\nfunc Place(id string) error { return storage.Save(id) }\n")
	write("specforge/features.json",
		`{"schema_version":"2.0","features":[{"name":"orders","status":"checking","gates":[]}]}`)
	write("specforge/features/orders/design.json",
		`{"schema_version":"2.0","feature":"orders","components":[
		  {"id":"handler","name":"h","depends_on":["service"]},
		  {"id":"service","name":"s","depends_on":["storage"]},
		  {"id":"storage","name":"st","depends_on":[]}]}`)
	write("specforge/features/orders/tasks.json",
		`{"schema_version":"2.0","feature":"orders","tasks":[
		  {"id":"T1","component_refs":["handler"],"files_touched":["internal/handler/handler.go"]},
		  {"id":"T2","component_refs":["service"],"files_touched":["internal/service/service.go"]},
		  {"id":"T3","component_refs":["storage"],"files_touched":["internal/storage/storage.go"]}]}`)

	cfg, reasons := writeArchRules(proj, "orders", "go-arch-lint")
	if len(reasons) != 0 {
		t.Fatalf("no se pudo emitir: %v", reasons)
	}
	run := func() int {
		c := exec.Command(bin, "check", "--arch-file", cfg)
		c.Dir = proj
		_ = c.Run()
		return c.ProcessState.ExitCode()
	}

	t.Run("el código que respeta el diseño pasa", func(t *testing.T) {
		write("internal/handler/handler.go",
			"package handler\n\nimport \"example.com/archfx/internal/service\"\n\nfunc Handle(id string) error { return service.Place(id) }\n")
		if code := run(); code != 0 {
			t.Errorf("exit = %d, want 0 — go-arch-lint debe aceptar nuestra config y no ver violaciones", code)
		}
	})

	t.Run("un import que viola el grafo aprobado se detecta", func(t *testing.T) {
		// `handler` declara depender de `service`, no de `storage`. El código
		// COMPILA igual: es exactamente el agujero que RM-C7 existe para tapar.
		write("internal/handler/handler.go",
			"package handler\n\nimport (\n\t\"example.com/archfx/internal/service\"\n\t\"example.com/archfx/internal/storage\"\n)\n\n"+
				"func Handle(id string) error {\n\t_ = storage.Save(id)\n\treturn service.Place(id)\n}\n")
		if code := run(); code == 0 {
			t.Error("handler importa storage saltándose el diseño y no se detectó")
		}
	})
}
