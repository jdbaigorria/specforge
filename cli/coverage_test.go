package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
)

// ----------------------------------------------------------------------------
// F3 de CLI-COMO-FUENTE (R4): la categoría 4 del drift, "fuera de spec", pasa
// de número a LISTA.
//
// El diagnóstico que motiva esto: `sf coverage` decía "78%" y nadie sabía qué
// hacer con el 22% restante. Un porcentaje no es accionable; una lista de rutas
// sí — cada una se adopta (se le escribe el requisito) o se excluye (con motivo
// en constitution.json).
// ----------------------------------------------------------------------------

// makeCoverageProject arma un repo con 3 archivos de código, 1 anclado por un
// trace. `sf coverage` debería dar 33.3% con dos rutas sin anclar.
//
// No inicializa git a propósito: sin repo, computeSpecCoverageDetail cae en
// walkListFiles, que es el camino determinista para un test.
func makeCoverageProject(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "src/anchored.go"), "package src\n\nfunc Parse() {}\n")
	writeFile(t, filepath.Join(dir, "src/loose.go"), "package src\n\nfunc Loose() {}\n")
	// scripts/ y no vendor/: skipDirsForHash ya saltea vendor, node_modules y
	// compañía antes de que la cobertura los vea. El caso que queremos ejercitar
	// es el del código PROPIO que nadie especificó.
	writeFile(t, filepath.Join(dir, "scripts/release.go"), "package scripts\n")
	// Un test no cuenta en el denominador (la métrica es sobre código de producto).
	writeFile(t, filepath.Join(dir, "src/anchored_test.go"), "package src\n")
	writeFile(t, filepath.Join(dir, "specforge/features/widget/trace.json"),
		`{"feature":"widget","requirements":{"R1":{"code":["src/anchored.go:Parse"],"test":["src/anchored_test.go:TestParse"],"status":"covered"}}}`)
	return dir
}

func TestCoverageUnanchored(t *testing.T) {
	t.Run("enumera las rutas, no sólo las cuenta", func(t *testing.T) {
		d := computeSpecCoverageDetail(makeCoverageProject(t))
		if d.Anchored != 1 || d.Total != 3 {
			t.Fatalf("anchored=%d total=%d, want 1/3", d.Anchored, d.Total)
		}
		want := []string{"scripts/release.go", "src/loose.go"}
		if len(d.Unanchored) != len(want) {
			t.Fatalf("unanchored=%v, want %v", d.Unanchored, want)
		}
		// Ordenadas: la lista se diffea entre corridas, así que el orden importa.
		for i, w := range want {
			if d.Unanchored[i] != w {
				t.Errorf("unanchored[%d]=%q, want %q", i, d.Unanchored[i], w)
			}
		}
	})

	t.Run("el ratio no se movió — R4 exige que percent no cambie", func(t *testing.T) {
		dir := makeCoverageProject(t)
		// El envoltorio viejo (el que usan status/verify) tiene que seguir dando
		// exactamente lo mismo que antes de partir la función.
		anchored, total := computeSpecCoverage(dir)
		d := computeSpecCoverageDetail(dir)
		if anchored != d.Anchored || total != d.Total {
			t.Errorf("computeSpecCoverage=(%d,%d) ≠ detail=(%d,%d)", anchored, total, d.Anchored, d.Total)
		}
	})

	t.Run("--json trae unanchored y sigue trayendo lo de antes", func(t *testing.T) {
		dir := makeCoverageProject(t)
		out := captureStdout(t, func() {
			if code := runCoverage([]string{"--json", dir}); code != 0 {
				t.Errorf("code=%d, want 0", code)
			}
		})
		var got struct {
			Percent    float64  `json:"percent"`
			Anchored   int      `json:"anchored"`
			Total      int      `json:"total"`
			Unanchored []string `json:"unanchored"`
			Excluded   int      `json:"excluded"`
		}
		if err := json.Unmarshal([]byte(out), &got); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		if got.Anchored != 1 || got.Total != 3 {
			t.Errorf("anchored=%d total=%d, want 1/3", got.Anchored, got.Total)
		}
		if len(got.Unanchored) != 2 {
			t.Errorf("unanchored=%v, want 2 rutas", got.Unanchored)
		}
	})

	t.Run("sin pendientes emite [] y no null", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "specforge/features/w/trace.json"),
			`{"feature":"w","requirements":{"R1":{"code":["x.go:F"],"test":[],"status":"covered"}}}`)
		out := captureStdout(t, func() { runCoverage([]string{"--json", dir}) })
		var probe map[string]json.RawMessage
		if err := json.Unmarshal([]byte(out), &probe); err != nil {
			t.Fatalf("salida no parseable: %v\n%s", err, out)
		}
		if s := string(probe["unanchored"]); s == "null" {
			t.Error("unanchored == null; queremos []")
		}
	})
}

func TestCoverageExclusions(t *testing.T) {
	t.Run("un prefijo de directorio saca los archivos del denominador", func(t *testing.T) {
		dir := makeCoverageProject(t)
		writeFile(t, filepath.Join(dir, "specforge/constitution.json"), `{
			"schema_version":"1.0","identity_md":"x","principles":[],"constraints":[],"anti_goals":[],"invariants":[],
			"coverage":{"exclude":[{"path":"scripts/","reason":"tooling de release, no es producto"}]}}`)
		d := computeSpecCoverageDetail(dir)
		if d.Total != 2 || d.Excluded != 1 {
			t.Fatalf("total=%d excluded=%d, want 2/1", d.Total, d.Excluded)
		}
		for _, u := range d.Unanchored {
			if u == "scripts/release.go" {
				t.Error("el archivo excluido sigue apareciendo como pendiente")
			}
		}
	})

	t.Run("una ruta exacta excluye ese archivo y nada más", func(t *testing.T) {
		ex := []coverageExclusion{{Path: "src/loose.go", Reason: "script"}}
		if !coverageExcluded("src/loose.go", ex) {
			t.Error("la ruta exacta debería excluirse")
		}
		if coverageExcluded("src/loose_other.go", ex) {
			t.Error("un prefijo de NOMBRE no debe excluir — sólo ruta exacta o directorio con /")
		}
	})

	t.Run("sin constitución no hay exclusiones — el default es medir todo", func(t *testing.T) {
		if got := coverageExclusions(t.TempDir()); got != nil {
			t.Errorf("exclusions=%v, want nil", got)
		}
	})

	t.Run("una exclusión sin motivo no valida", func(t *testing.T) {
		cf := constitutionFile{
			IdentityMD: "x",
			Coverage:   &coverageConfig{Exclude: []coverageExclusion{{Path: "vendor/"}}},
		}
		var rep report
		checkConstitution(cf, &rep)
		if len(rep.errors) == 0 {
			t.Fatal("una exclusión sin reason debería ser error: así se maquilla la métrica sin dejar rastro")
		}
	})

	t.Run("dos exclusiones de la misma ruta no validan", func(t *testing.T) {
		cf := constitutionFile{
			IdentityMD: "x",
			Coverage: &coverageConfig{Exclude: []coverageExclusion{
				{Path: "vendor/", Reason: "a"}, {Path: "vendor/", Reason: "b"},
			}},
		}
		var rep report
		checkConstitution(cf, &rep)
		if len(rep.errors) == 0 {
			t.Error("path duplicado debería ser error (dos motivos para lo mismo = ninguno es el motivo)")
		}
	})
}
