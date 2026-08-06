package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// PRA-3 — los tres `examples/` como suite de regresión.
//
// POR QUÉ EXISTE ESTE ARCHIVO, con nombre y apellido. Durante el bloque `RM-*`
// pasó cuatro veces lo mismo: suite verde, y correr el binario contra un ejemplo
// encontró lo que ningún test había visto.
//
//	RM-C2b   `examples/slugify` daba 0/4 teniendo los cuatro anclados — el trace
//	         de una feature `done` vive en `archive/`, y se lo buscaba en `features/`.
//	RM-C7    `may_depend_on` serializaba `null`, y los tests contaban como parte
//	         del componente.
//	RM-MIG   el salteo de features archivadas no se declaraba.
//	DOC-2    migrar endurecía el contrato sin avisar (3/3 → 0/3).
//
// La causa es estructural y vale nombrarla: los fixtures de los otros tests son
// proyectos de juguete construidos para el caso que ese test prueba. Los
// `examples/` tienen la forma de un proyecto de verdad — features archivadas,
// tres lanes, artefactos completos, historia de gates. Por eso encontraron cosas
// que un `t.TempDir()` con dos archivos no podía encontrar.
//
// Y ADEMÁS PROTEGE LO QUE YA SE PAGÓ. Los ejemplos llegaron a esquema 1.0 con
// `acceptance` legada porque **nada los chequeaba**: quedaron atrás en silencio
// mientras el modelo avanzaba nueve fases. `DOC-2` los puso al día; sin este
// archivo, el próximo cambio de esquema los deja atrás otra vez.
//
// NO NECESITA EL RUNNER DE `PRA-2`. El roadmap listaba `PRA-3` detrás de él
// ("una vez que el runner probó que sirve"), y esa dependencia era inventada:
// `PRA-2` simula COMPORTAMIENTO DE AGENTE; esto es determinista y no necesita
// ningún agente.
// ----------------------------------------------------------------------------

// exampleProjects: los tres lanes que el producto dice soportar.
var exampleProjects = []struct {
	name string
	// hasRequirements distingue el carril `lite` (un `change.md`, sin
	// requirements.json) de los otros dos. No es un detalle del fixture: es la
	// diferencia de carril que el producto promete, así que se afirma.
	hasRequirements bool
}{
	{"slugify", true},             // standard, feature archivada (`done`)
	{"lite-wordcount", false},     // lite
	{"brownfield-tempconv", true}, // brownfield, feature viva
}

// copyExample copia un ejemplo a un temp dir.
//
// Se copia y no se usa in situ porque varios comandos ESCRIBEN: `sf coverage`
// deja su baseline en `specforge/.state/`. Un test que ensucia el repo del
// proyecto es un test que alguien termina desactivando.
func copyExample(t *testing.T, name string) string {
	t.Helper()
	src := filepath.Join("..", "examples", name)
	if _, err := os.Stat(src); err != nil {
		t.Skipf("examples/%s no está disponible: %v", name, err)
	}
	dst := filepath.Join(t.TempDir(), name)

	err := filepath.WalkDir(src, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, p)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("no se pudo copiar examples/%s: %v", name, err)
	}
	return dst
}

// TestExamplesVerify — los tres ejemplos pasan la verificación de integridad.
func TestExamplesVerify(t *testing.T) {
	for _, ex := range exampleProjects {
		t.Run(ex.name, func(t *testing.T) {
			dir := copyExample(t, ex.name)
			if code := runVerify([]string{dir}); code != 0 {
				t.Errorf("sf verify exit %d — el ejemplo que enseñamos no verifica", code)
			}
		})
	}
}

// TestExamplesAreOnCurrentSchema — el ejemplo se queda atrás EN SILENCIO, y esa
// es exactamente la forma en que llegaron a 1.0 mientras el modelo iba por 2.0.
//
// El oráculo es nuestra propia herramienta: si `sf migrate --dry-run` encuentra
// algo que hacer, el ejemplo está desactualizado. Se elige eso en vez de
// comparar contra una constante porque se actualiza solo — el día que exista una
// 3.0, este test falla pidiendo regenerar los ejemplos sin que nadie lo toque.
func TestExamplesAreOnCurrentSchema(t *testing.T) {
	for _, ex := range exampleProjects {
		t.Run(ex.name, func(t *testing.T) {
			dir := copyExample(t, ex.name)

			out := captureStdout(t, func() {
				if code := runMigrate([]string{"--dry-run", dir}); code != 0 {
					t.Fatalf("sf migrate --dry-run exit %d", code)
				}
			})
			if !strings.Contains(out, "Nothing to migrate") {
				t.Errorf("examples/%s quedó atrás del esquema actual — regeneralo con `sf migrate`:\n%s",
					ex.name, out)
			}
		})
	}
}

// TestExamplesHaveNoLegacyAcceptance — ningún criterio suelto sin id.
//
// Es la regresión directa de `DOC-2`. Un `acceptance: []string` en un ejemplo no
// es sólo un formato viejo: es la superficie de onboarding enseñando el contrato
// DÉBIL, el que `RM-C1` existe para eliminar.
func TestExamplesHaveNoLegacyAcceptance(t *testing.T) {
	for _, ex := range exampleProjects {
		t.Run(ex.name, func(t *testing.T) {
			dir := copyExample(t, ex.name)

			err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() || filepath.Base(p) != "requirements.json" {
					return err
				}
				data, err := os.ReadFile(p)
				if err != nil {
					return err
				}
				var doc struct {
					Requirements []struct {
						ID         string            `json:"id"`
						Acceptance []json.RawMessage `json:"acceptance"`
					} `json:"requirements"`
				}
				if json.Unmarshal(data, &doc) != nil {
					return nil // ilegible lo reporta `sf verify`, no este test
				}
				for _, r := range doc.Requirements {
					for i, raw := range r.Acceptance {
						if strings.HasPrefix(strings.TrimSpace(string(raw)), `"`) {
							t.Errorf("%s: %s.acceptance[%d] es un string suelto — el ejemplo enseña el contrato v1",
								filepath.Base(filepath.Dir(p)), r.ID, i)
						}
					}
				}
				return nil
			})
			if err != nil {
				t.Fatalf("recorriendo el ejemplo: %v", err)
			}
		})
	}
}

// TestExamplesSatisfyTheirContract — el ejemplo cumple lo que el producto exige.
//
// Es la regresión de `RM-C2b`: acá `examples/slugify` daba 0/4 con los cuatro
// requisitos anclados. Un ejemplo que no cumple su propio contrato de
// verificación le enseña al usuario que el gate es decorativo.
func TestExamplesSatisfyTheirContract(t *testing.T) {
	for _, ex := range exampleProjects {
		if !ex.hasRequirements {
			continue
		}
		t.Run(ex.name, func(t *testing.T) {
			dir := copyExample(t, ex.name)

			rc := computeRequirementCoverage(dir)
			if rc.Total == 0 {
				t.Fatal("el ejemplo no expone requisitos — el fixture dejó de tener sentido")
			}
			if rc.Contracted != rc.Total {
				t.Errorf("contrato %d/%d — sin cumplir: %v",
					rc.Contracted, rc.Total, rc.ByPriority["must"].Unmet)
			}
		})
	}
}

// TestExampleLiteLaneHasNoRequirements — el carril `lite` es una PROMESA del
// producto (un `change.md` en vez del pipeline entero), no una casualidad del
// fixture. Si un día `lite-wordcount` gana un requirements.json, o el carril
// cambió de forma o el ejemplo dejó de ilustrarlo.
func TestExampleLiteLaneHasNoRequirements(t *testing.T) {
	dir := copyExample(t, "lite-wordcount")
	matches, _ := filepath.Glob(filepath.Join(dir, "specforge", "features", "*", "requirements.json"))
	if len(matches) > 0 {
		t.Errorf("el carril lite no debería tener requirements.json: %v", matches)
	}
	if rc := computeRequirementCoverage(dir); rc.Total != 0 {
		t.Errorf("cobertura de requisitos = %d, want 0 en carril lite", rc.Total)
	}
}

// TestExampleArchRulesAreDerivable — RM-C7 de punta a punta sobre datos reales.
//
// `slugify` es el caso favorable: un componente, archivos propios. Además está
// ARCHIVADA, así que de paso ejercita que el mapeo se resuelva por `findArtifact`
// — el bug que RM-C2b encontró en este mismo ejemplo.
func TestExampleArchRulesAreDerivable(t *testing.T) {
	dir := copyExample(t, "slugify")

	rules, reasons := buildArchRules(dir, "slugify")
	if len(reasons) > 0 {
		t.Fatalf("no se pudo derivar el mapeo sobre una feature archivada: %v", reasons)
	}
	if len(rules.Components) == 0 {
		t.Fatal("cero componentes — el ejemplo dejó de declarar arquitectura")
	}
	for _, c := range rules.Components {
		if len(c.Files) == 0 {
			t.Errorf("%s sin archivos: el mapeo se derivó vacío", c.ID)
		}
		for _, f := range c.Files {
			if isTestFile(f) {
				t.Errorf("%s incluye el test %q — los tests no son parte del componente", c.ID, f)
			}
		}
	}
}

// TestExampleArchRulesRefuseAmbiguousMapping — EL LÍMITE REAL DE RM-C7, con un
// caso que no inventamos: lo trajo `examples/brownfield-tempconv`.
//
// Sus dos componentes —`C1` conversiones kelvin, `C2` guard de cero absoluto—
// viven los dos en `src/tempconv/convert.py`. C7 mapea componente→ARCHIVO (vía
// `tasks.component_refs` × `files_touched`), así que ese archivo lo reclaman dos
// componentes y el mapeo es ambiguo.
//
// Lo que este test fija es la propiedad de seguridad: ante la ambigüedad, C7
// **rechaza nombrando a los dos** en vez de adivinar. Con un mapeo inventado,
// toda la conformidad posterior sería decorativa.
//
// Lo que este test DOCUMENTA, y no hay que leerlo como un detalle del fixture:
// **la granularidad de archivo es demasiado gruesa para código chico.** Dos
// componentes lógicos en un mismo archivo es lo normal en un proyecto pequeño,
// no una rareza. Mientras el mapeo sea por archivo, `require_arch` sólo sirve en
// proyectos donde cada componente ya es su propio módulo o carpeta. Está
// registrado como `RM-C7b` en el roadmap; el camino es la granularidad de
// SÍMBOLO, que el trace ya tiene (`path:symbol`).
func TestExampleArchRulesRefuseAmbiguousMapping(t *testing.T) {
	dir := copyExample(t, "brownfield-tempconv")

	_, reasons := buildArchRules(dir, "add-kelvin")
	if len(reasons) == 0 {
		t.Fatal("dos componentes comparten convert.py: derivar un mapeo acá sería inventarlo")
	}
	joined := strings.Join(reasons, " ")
	for _, want := range []string{"src/tempconv/convert.py", "C1", "C2", "component_refs"} {
		if !strings.Contains(joined, want) {
			t.Errorf("el rechazo debe nombrar %q para ser accionable: %v", want, reasons)
		}
	}
}

// TestExampleMutationScopeIsProductionOnly — RM-C6 sobre datos reales: el
// alcance sale del trace y no arrastra tests. Mutar un test no dice nada sobre
// si el suite detecta defectos del producto.
func TestExampleMutationScopeIsProductionOnly(t *testing.T) {
	dir := copyExample(t, "brownfield-tempconv")

	files := mutationScope(dir, "add-kelvin")
	if len(files) == 0 {
		t.Fatal("alcance vacío sobre un ejemplo con trace completo")
	}
	for _, f := range files {
		if isTestFile(f) {
			t.Errorf("el alcance incluye el test %q", f)
		}
	}
}
