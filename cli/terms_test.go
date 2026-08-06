package main

import (
	"path/filepath"
	"testing"
)

// ----------------------------------------------------------------------------
// DL-14b — terminología inconsistente cross-feature.
//
// El defecto que busca no se ve leyendo una feature: cada spec es internamente
// coherente. Por eso los tests que importan son los que cruzan DOS features, y
// el que fija los límites de palabra — sin ellos el reporte se llena de falsos
// y un reporte con falsos no se usa.
// ----------------------------------------------------------------------------

// makeTermsProject arma un proyecto con glosario y N features, cada una con un
// requisito cuyo `behavior` es el texto dado.
func makeTermsProject(t *testing.T, glossary string, features map[string]string) string {
	t.Helper()
	dir := t.TempDir()

	names := ""
	for name := range features {
		if names != "" {
			names += ","
		}
		names += `{"name":"` + name + `","status":"checking","gates":[]}`
	}
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"2.0","features":[`+names+`]}`)
	writeFile(t, filepath.Join(dir, "specforge/context/domain.json"),
		`{"schema_version":"2.0","glossary":`+glossary+`}`)

	for name, behavior := range features {
		writeFile(t, filepath.Join(dir, "specforge/features", name, "requirements.json"),
			`{"schema_version":"2.0","feature":"`+name+`","requirements":[
			  {"id":"R1","ears_type":"ubiquitous","behavior":`+jsonString(behavior)+`,
			   "acceptance":[{"id":"R1.1","text":"ok"}]}]}`)
	}
	return dir
}

func TestCrossFeatureTerminology(t *testing.T) {
	glossary := `[{"term":"Order","definition":"una compra confirmada","aliases":["Pedido"]}]`

	t.Run("dos features nombrando distinto al mismo concepto", func(t *testing.T) {
		dir := makeTermsProject(t, glossary, map[string]string{
			"checkout": "create the Order and persist it",
			"billing":  "emit an invoice for the Pedido",
		})

		rep := computeTermsReport(dir)
		if len(rep.Inconsistent) != 1 {
			t.Fatalf("inconsistent = %+v, want 1", rep.Inconsistent)
		}
		f := rep.Inconsistent[0]
		if f.Term != "Order" || len(f.Forms) != 2 {
			t.Fatalf("finding = %+v", f)
		}
		// El mapa entero, no sólo "hay conflicto": la acción depende de quién
		// dice qué.
		if got := f.Forms["Order"]; len(got) != 1 || got[0] != "checkout" {
			t.Errorf("Forms[Order] = %v", got)
		}
		if got := f.Forms["Pedido"]; len(got) != 1 || got[0] != "billing" {
			t.Errorf("Forms[Pedido] = %v", got)
		}
	})

	t.Run("las dos usando el MISMO nombre no es hallazgo", func(t *testing.T) {
		dir := makeTermsProject(t, glossary, map[string]string{
			"checkout": "create the Order",
			"billing":  "invoice the Order",
		})
		if rep := computeTermsReport(dir); len(rep.Inconsistent) != 0 {
			t.Errorf("inconsistent = %+v", rep.Inconsistent)
		}
	})

	t.Run("una sola feature usando las dos formas TAMBIÉN es hallazgo", func(t *testing.T) {
		// Es el mismo defecto sin cruzar features, y ya está mal ahí.
		dir := makeTermsProject(t, glossary, map[string]string{
			"checkout": "create the Order, then archive the Pedido",
		})
		if rep := computeTermsReport(dir); len(rep.Inconsistent) != 1 {
			t.Errorf("inconsistent = %+v, want 1", rep.Inconsistent)
		}
	})

	t.Run("respeta los LÍMITES DE PALABRA", func(t *testing.T) {
		// Sin esto `Order` matchea dentro de `Reorder` y de `Ordering`, y el
		// reporte se llena de falsos que lo vuelven inservible.
		dir := makeTermsProject(t, glossary, map[string]string{
			"checkout": "create the Order",
			"reports":  "allow Reordering the list and reorder columns",
		})
		if rep := computeTermsReport(dir); len(rep.Inconsistent) != 0 {
			t.Errorf("`Reordering` no es una mención de `Order`: %+v", rep.Inconsistent)
		}
	})

	t.Run("una diferencia de mayúscula NO es otro nombre", func(t *testing.T) {
		dir := makeTermsProject(t, glossary, map[string]string{
			"checkout": "create the Order",
			"billing":  "invoice the order",
		})
		if rep := computeTermsReport(dir); len(rep.Inconsistent) != 0 {
			t.Errorf("`order` y `Order` son la misma palabra: %+v", rep.Inconsistent)
		}
	})

	t.Run("un término que nadie menciona es informativo, no error", func(t *testing.T) {
		// La misma lección de DL-4: si esto pusiera el comando en rojo, saldría
		// en rojo en todo proyecto real y se dejaría de correr.
		dir := makeTermsProject(t, glossary, map[string]string{
			"checkout": "do something unrelated",
		})
		rep := computeTermsReport(dir)
		if len(rep.Unused) != 1 || rep.Unused[0] != "Order" {
			t.Fatalf("unused = %v", rep.Unused)
		}
		if len(rep.Inconsistent) != 0 {
			t.Errorf("inconsistent = %+v", rep.Inconsistent)
		}
		if code := runDomain([]string{"terms", dir}); code != 0 {
			t.Errorf("exit = %d, want 0 — un término sin usar no es un error", code)
		}
	})

	t.Run("una feature CERRADA queda afuera", func(t *testing.T) {
		// Mismo motivo que en coverage y drift (DL-5 F3): una spec retirada a
		// propósito no es terminología que haya que unificar.
		dir := makeTermsProject(t, glossary, map[string]string{
			"checkout": "create the Order",
			"vieja":    "archive the Pedido",
		})
		writeFile(t, filepath.Join(dir, "specforge/features.json"),
			`{"schema_version":"2.0","features":[
			  {"name":"checkout","status":"checking","gates":[]},
			  {"name":"vieja","status":"retired","closed_reason":"se removió","retired_at":"2026-01-01T00:00:00Z","gates":[]}]}`)

		if rep := computeTermsReport(dir); len(rep.Inconsistent) != 0 {
			t.Errorf("una feature retirada no debe generar hallazgo: %+v", rep.Inconsistent)
		}
	})
}

func TestDomainTermsCommand(t *testing.T) {
	glossary := `[{"term":"Order","definition":"d","aliases":["Pedido"]}]`

	t.Run("sale 1 cuando hay inconsistencia", func(t *testing.T) {
		// Exit code y no sólo texto: así entra en CI.
		dir := makeTermsProject(t, glossary, map[string]string{
			"checkout": "create the Order",
			"billing":  "invoice the Pedido",
		})
		if code := runDomain([]string{"terms", dir}); code != 1 {
			t.Errorf("exit = %d, want 1", code)
		}
	})

	t.Run("sin glosario no es un error", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "specforge/features.json"),
			`{"schema_version":"2.0","features":[]}`)
		if code := runDomain([]string{"terms", dir}); code != 0 {
			t.Errorf("exit = %d, want 0", code)
		}
	})

	t.Run("flag desconocido sale 2", func(t *testing.T) {
		if code := runDomain([]string{"terms", "--nope"}); code != 2 {
			t.Errorf("exit = %d, want 2", code)
		}
	})
}
