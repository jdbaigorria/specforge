package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// RM-C1 F1 — `acceptance` con id (R1, R4, R8).
//
// El agujero: el contrato del veredicto se cumple a nivel `R#`, así que un
// requisito con cinco criterios sella en verde con un test. El id es lo que
// permite anclar un test a CADA criterio — el humano aprueba `R5.2` y la máquina
// exige un test para `R5.2`, que es el mismo objeto.
// ----------------------------------------------------------------------------

// TestAcceptanceDoubleRead — R4. Un proyecto puede tener las dos formas, y la
// legada tiene que seguir funcionando idéntica.
func TestAcceptanceDoubleRead(t *testing.T) {
	t.Run("forma legada: strings sin id", func(t *testing.T) {
		var a acceptanceList
		if err := json.Unmarshal([]byte(`["slugify(\"\") == \"\"", "slugify(\"!!!\") == \"\""]`), &a); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(a) != 2 {
			t.Fatalf("len=%d, want 2", len(a))
		}
		if a[0].Text != `slugify("") == ""` || a[0].ID != "" {
			t.Errorf("got %+v; el legado entra sin id", a[0])
		}
		if a.hasIDs() {
			t.Error("hasIDs=true sobre forma legada")
		}
		if !a.legacy() {
			t.Error("legacy=false sobre forma legada")
		}
	})

	t.Run("forma v2: objetos con id", func(t *testing.T) {
		var a acceptanceList
		if err := json.Unmarshal([]byte(`[
			{"id":"R4.1","text":"slugify(\"!!!\") == \"\""},
			{"id":"R4.2","given":"un carrito vacío","when":"confirma","then":"rechaza con EMPTY_CART"}]`), &a); err != nil {
			t.Fatalf("unmarshal: %v", err)
		}
		if len(a) != 2 || a[0].ID != "R4.1" || a[1].Then != "rechaza con EMPTY_CART" {
			t.Fatalf("got %+v", a)
		}
		if !a.hasIDs() || a.legacy() {
			t.Error("una lista con ids no es legada")
		}
	})

	t.Run("lista vacía y null no son error", func(t *testing.T) {
		for _, in := range []string{`[]`, `null`} {
			var a acceptanceList
			if err := json.Unmarshal([]byte(in), &a); err != nil {
				t.Errorf("unmarshal(%s): %v", in, err)
			}
			if len(a) != 0 {
				t.Errorf("unmarshal(%s) → len=%d, want 0", in, len(a))
			}
		}
	})

	t.Run("un número no es un criterio", func(t *testing.T) {
		var a acceptanceList
		if err := json.Unmarshal([]byte(`[42]`), &a); err == nil {
			t.Error("want error: 42 no es ni string ni objeto")
		}
	})
}

// TestAcceptanceRoundTripPreservesShape es el test que impide una subida de
// contrato SILENCIOSA. `sf save` re-serializa el JSON canónico; si un requisito
// legado volviera escrito como objetos, un simple save lo habría movido de
// contrato v1 a v2 y el veredicto pasaría a exigir un test por criterio sin que
// nadie lo decidiera.
func TestAcceptanceRoundTripPreservesShape(t *testing.T) {
	cases := []struct {
		name, in, want string
	}{
		{"legado sigue siendo strings", `["a","b"]`, `["a","b"]`},
		{"vacío sigue vacío", `[]`, `[]`},
		{
			"v2 sigue siendo objetos",
			`[{"id":"R1.1","text":"a"}]`,
			`[{"id":"R1.1","text":"a"}]`,
		},
		{
			"la forma rica se preserva entera",
			`[{"id":"R1.1","given":"g","when":"w","then":"t"}]`,
			`[{"id":"R1.1","given":"g","when":"w","then":"t"}]`,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var a acceptanceList
			if err := json.Unmarshal([]byte(c.in), &a); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			out, err := json.Marshal(a)
			if err != nil {
				t.Fatalf("marshal: %v", err)
			}
			if string(out) != c.want {
				t.Errorf("round-trip = %s, want %s", out, c.want)
			}
		})
	}
}

// reqWith arma un requisito ubiquitous válido con los criterios dados.
func reqWith(id string, crits ...acceptanceCriterion) requirement {
	return requirement{
		ID: id, EarsType: "ubiquitous", Behavior: "do the thing",
		Acceptance: acceptanceList(crits),
	}
}

// errorsOf corre la validación y devuelve los errores.
func errorsOf(r requirement) []string {
	var rep report
	checkAcceptance(r, &rep)
	return rep.errors
}

func warningsOf(r requirement) []string {
	var rep report
	checkAcceptance(r, &rep)
	return rep.warnings
}

// TestAcceptanceValidation — R1.
func TestAcceptanceValidation(t *testing.T) {
	t.Run("legado no produce errores", func(t *testing.T) {
		r := reqWith("R1", acceptanceCriterion{Text: "a"}, acceptanceCriterion{Text: "b"})
		if errs := errorsOf(r); len(errs) != 0 {
			t.Errorf("want 0 errores sobre legado, got %v", errs)
		}
	})

	t.Run("mitad y mitad se rechaza", func(t *testing.T) {
		r := reqWith("R1",
			acceptanceCriterion{ID: "R1.1", Text: "a"},
			acceptanceCriterion{Text: "b"}) // sin id
		errs := errorsOf(r)
		if len(errs) != 1 || !strings.Contains(errs[0], "missing id") {
			t.Errorf("want error de id faltante, got %v", errs)
		}
	})

	t.Run("el id lleva el prefijo del requisito padre", func(t *testing.T) {
		r := reqWith("R5", acceptanceCriterion{ID: "R6.1", Text: "a"})
		errs := errorsOf(r)
		if len(errs) != 1 || !strings.Contains(errs[0], "belongs to R6") {
			t.Errorf("want error de prefijo, got %v", errs)
		}
	})

	t.Run("forma del id", func(t *testing.T) {
		for _, bad := range []string{"R1", "1.1", "R1.1.1", "S1.1", "R1."} {
			r := reqWith("R1", acceptanceCriterion{ID: bad, Text: "a"})
			if errs := errorsOf(r); len(errs) == 0 {
				t.Errorf("id %q debería rechazarse", bad)
			}
		}
	})

	t.Run("ids duplicados", func(t *testing.T) {
		r := reqWith("R1",
			acceptanceCriterion{ID: "R1.1", Text: "a"},
			acceptanceCriterion{ID: "R1.1", Text: "b"})
		errs := errorsOf(r)
		if len(errs) != 1 || !strings.Contains(errs[0], "duplicate") {
			t.Errorf("want error de duplicado, got %v", errs)
		}
	})

	t.Run("la numeración arranca en 1", func(t *testing.T) {
		r := reqWith("R1", acceptanceCriterion{ID: "R1.0", Text: "a"})
		if errs := errorsOf(r); len(errs) == 0 {
			t.Error("R1.0 debería rechazarse: el 0 es un off-by-one, no un criterio")
		}
	})

	t.Run("media forma rica se rechaza", func(t *testing.T) {
		for _, c := range []acceptanceCriterion{
			{ID: "R1.1", When: "w"},             // sin then
			{ID: "R1.1", Then: "t"},             // sin when
			{ID: "R1.1", Given: "g"},            // sólo given
			{ID: "R1.1", Given: "g", When: "w"}, // sin then
		} {
			if errs := errorsOf(reqWith("R1", c)); len(errs) == 0 {
				t.Errorf("%+v debería rechazarse: media forma rica dice menos que la corta", c)
			}
		}
	})

	t.Run("text y given/when/then juntos se rechazan", func(t *testing.T) {
		r := reqWith("R1", acceptanceCriterion{ID: "R1.1", Text: "a", When: "w", Then: "t"})
		errs := errorsOf(r)
		if len(errs) != 1 || !strings.Contains(errs[0], "not both") {
			t.Errorf("want error de forma ambigua, got %v", errs)
		}
	})

	t.Run("un criterio vacío se rechaza", func(t *testing.T) {
		r := reqWith("R1", acceptanceCriterion{ID: "R1.1", Text: "   "})
		if errs := errorsOf(r); len(errs) == 0 {
			t.Error("un criterio en blanco debería rechazarse")
		}
	})

	t.Run("las dos formas válidas pasan", func(t *testing.T) {
		r := reqWith("R1",
			acceptanceCriterion{ID: "R1.1", Text: `slugify("") == ""`},
			acceptanceCriterion{ID: "R1.2", Given: "g", When: "w", Then: "t"},
			acceptanceCriterion{ID: "R1.3", When: "w", Then: "t"}) // given es opcional
		if errs := errorsOf(r); len(errs) != 0 {
			t.Errorf("want 0 errores, got %v", errs)
		}
	})

	t.Run("sin criterios avisa: nada puede anclar un test", func(t *testing.T) {
		if ws := warningsOf(reqWith("R1")); len(ws) != 1 {
			t.Errorf("want 1 warning, got %v", ws)
		}
	})
}

// TestAcceptanceGapsAreLegal — la regla de estabilidad de ids.
//
// Un hueco es la SEÑAL de que un criterio se retiró, y es legal. Renumerar para
// cerrarlo es lo que rompe anclas en silencio: `trace.json` apunta a ids, así
// que reusar un id liberado lo re-apunta a otro caso sin que nada falle.
func TestAcceptanceGapsAreLegal(t *testing.T) {
	r := reqWith("R1",
		acceptanceCriterion{ID: "R1.1", Text: "a"},
		acceptanceCriterion{ID: "R1.3", Text: "c"}) // R1.2 se retiró

	if errs := errorsOf(r); len(errs) != 0 {
		t.Errorf("un hueco NO es error — renumerar es el error, got %v", errs)
	}
	ws := warningsOf(r)
	if len(ws) != 1 || !strings.Contains(ws[0], "R1.2") {
		t.Fatalf("want warning nombrando R1.2, got %v", ws)
	}
	if !strings.Contains(ws[0], "Never renumber") {
		t.Error("el warning tiene que decir por qué no se renumera, no sólo que hay un hueco")
	}
}

// TestAcceptanceRender — R8. Given/When/Then entra como render, no como fuente.
func TestAcceptanceRender(t *testing.T) {
	cases := []struct {
		name string
		c    acceptanceCriterion
		want string
	}{
		{"forma corta", acceptanceCriterion{Text: `slugify("") == ""`}, `slugify("") == ""`},
		{"con given", acceptanceCriterion{Given: "un carrito vacío", When: "confirma", Then: "rechaza"},
			"Given un carrito vacío, when confirma, then rechaza"},
		{"sin given", acceptanceCriterion{When: "confirma", Then: "rechaza"},
			"when confirma, then rechaza"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := c.c.line(); got != c.want {
				t.Errorf("line() = %q, want %q", got, c.want)
			}
		})
	}

	t.Run("el md muestra el id", func(t *testing.T) {
		rf := requirementsFile{
			Feature: "checkout",
			Requirements: []requirement{
				reqWith("R1", acceptanceCriterion{ID: "R1.1", When: "confirma", Then: "persiste"}),
			},
		}
		var buf strings.Builder
		if err := reqTmpl.Execute(&buf, rf); err != nil {
			t.Fatalf("render: %v", err)
		}
		out := buf.String()
		for _, want := range []string{"**R1.1**", "when confirma, then persiste"} {
			if !strings.Contains(out, want) {
				t.Errorf("el render no contiene %q:\n%s", want, out)
			}
		}
	})

	t.Run("el md legado no inventa un id", func(t *testing.T) {
		rf := requirementsFile{
			Feature:      "slugify",
			Requirements: []requirement{reqWith("R1", acceptanceCriterion{Text: `slugify("") == ""`})},
		}
		var buf strings.Builder
		if err := reqTmpl.Execute(&buf, rf); err != nil {
			t.Fatalf("render: %v", err)
		}
		out := buf.String()
		if !strings.Contains(out, `- slugify("") == ""`) {
			t.Errorf("el render legado cambió:\n%s", out)
		}
		if strings.Contains(out, "**") && strings.Contains(out, "** — ") {
			t.Error("el render legado no debe inventar un id")
		}
	})
}
