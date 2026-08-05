package main

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
)

// mustUnmarshal parsea JSON en v, abortando el test si el literal está mal.
func mustUnmarshal(t *testing.T, data string, v any) {
	t.Helper()
	if err := json.Unmarshal([]byte(data), v); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, data)
	}
}

// ----------------------------------------------------------------------------
// RM-C1 F1b — el contrato de verificación a nivel criterio (R3, R10).
//
// El fraude que cierra, y no requiere mala fe: un requisito con cinco criterios
// y un test honesto del caso feliz pasa el contrato v1, sale verde y queda
// sellado con hash encadenado. Los otros cuatro criterios nunca se tocaron. El
// sello certifica correctamente el cumplimiento de una regla que mide poco.
// ----------------------------------------------------------------------------

// makeVerdictProject arma una feature lista para el verdict salvo por lo que
// cada subtest quiera romper: código, test, requirements.json y trace.json.
//
// `acceptance` y `trace` los pone el llamador — son justamente las dos piezas
// cuya relación estamos probando.
func makeVerdictProject(t *testing.T, acceptance, trace string) string {
	t.Helper()
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/features.json"),
		`{"schema_version":"1.0","features":[{"name":"checkout","status":"checking","lane":"standard","gates":[]}]}`)
	writeFile(t, filepath.Join(dir, "src/order.go"),
		"package src\n\nfunc CreateOrder() {}\n")
	writeFile(t, filepath.Join(dir, "src/order_test.go"),
		"package src\n\nfunc TestPersists() {}\nfunc TestEmptyCart() {}\n")
	writeFile(t, filepath.Join(dir, "specforge/features/checkout/requirements.json"),
		`{"schema_version":"1.0","feature":"checkout","requirements":[
			{"id":"R5","ears_type":"ubiquitous","behavior":"persist the order",
			 "acceptance":`+acceptance+`}]}`)
	writeFile(t, filepath.Join(dir, "specforge/features/checkout/trace.json"), trace)
	return dir
}

// reasonsFor corre las precondiciones y devuelve sólo las del contrato de tests
// (descartamos las de resultado de `sf check run`, que este test no monta).
func reasonsFor(t *testing.T, dir string) []string {
	t.Helper()
	var out []string
	for _, r := range verdictPreconditions(dir, "checkout") {
		if strings.Contains(r, "check run") || strings.Contains(r, "test result") {
			continue
		}
		out = append(out, r)
	}
	return out
}

// TestVerdictContractV1Unchanged — R4. Un requisito legado conserva la regla
// vieja, exactamente como antes.
func TestVerdictContractV1Unchanged(t *testing.T) {
	const legacy = `["persiste la orden","rechaza el carrito vacio"]`

	t.Run("un solo test alcanza — es el contrato viejo", func(t *testing.T) {
		dir := makeVerdictProject(t, legacy,
			`{"feature":"checkout","requirements":{"R5":{
				"code":["src/order.go:CreateOrder"],
				"test":["src/order_test.go:TestPersists"],"status":"ok"}}}`)
		if rs := reasonsFor(t, dir); len(rs) != 0 {
			t.Errorf("want 0 razones bajo v1, got %v", rs)
		}
	})

	t.Run("sin ningún test se rechaza, nombrando el requisito", func(t *testing.T) {
		dir := makeVerdictProject(t, legacy,
			`{"feature":"checkout","requirements":{"R5":{
				"code":["src/order.go:CreateOrder"],"test":[],"status":"ok"}}}`)
		rs := reasonsFor(t, dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "R5: names no test") {
			t.Errorf("want rechazo nombrando R5, got %v", rs)
		}
	})
}

// TestVerdictContractV2 — R3, el corazón de RM-C1.
func TestVerdictContractV2(t *testing.T) {
	const structured = `[
		{"id":"R5.1","text":"la orden queda persistida"},
		{"id":"R5.2","text":"el carrito vacio se rechaza"}]`

	t.Run("EL FRAUDE: un test del caso feliz ya no alcanza", func(t *testing.T) {
		// Exactamente el escenario de §1: cinco criterios (acá dos), un test
		// honesto del primero, y bajo v1 esto sellaba en verde.
		dir := makeVerdictProject(t, structured,
			`{"feature":"checkout","requirements":{"R5":{
				"code":["src/order.go:CreateOrder"],
				"test":["src/order_test.go:TestPersists"],
				"scenarios":{"R5.1":{"test":["src/order_test.go:TestPersists"]}},
				"status":"ok"}}}`)
		rs := reasonsFor(t, dir)
		if len(rs) != 1 {
			t.Fatalf("want 1 razón (R5.2 sin test), got %v", rs)
		}
		// Y el mensaje nombra EL CRITERIO, no el requisito: "R5 no tiene test"
		// manda a releer cinco criterios para encontrar cuál falta.
		if !strings.HasPrefix(rs[0], "R5.2:") {
			t.Errorf("el rechazo debe nombrar R5.2, no R5: %q", rs[0])
		}
	})

	t.Run("un test por criterio habilita", func(t *testing.T) {
		dir := makeVerdictProject(t, structured,
			`{"feature":"checkout","requirements":{"R5":{
				"code":["src/order.go:CreateOrder"],"test":[],
				"scenarios":{
					"R5.1":{"test":["src/order_test.go:TestPersists"]},
					"R5.2":{"test":["src/order_test.go:TestEmptyCart"]}},
				"status":"ok"}}}`)
		if rs := reasonsFor(t, dir); len(rs) != 0 {
			t.Errorf("want 0 razones, got %v", rs)
		}
	})

	t.Run("un test a nivel requisito NO satisface a un criterio", func(t *testing.T) {
		// §3.3: `test` a nivel requisito es para cobertura TRANSVERSAL. Un test
		// que prueba "algo de R5" no prueba R5.2, y confundirlos reabre el
		// agujero por la puerta de atrás.
		dir := makeVerdictProject(t, structured,
			`{"feature":"checkout","requirements":{"R5":{
				"code":["src/order.go:CreateOrder"],
				"test":["src/order_test.go:TestPersists","src/order_test.go:TestEmptyCart"],
				"status":"ok"}}}`)
		rs := reasonsFor(t, dir)
		if len(rs) != 2 {
			t.Fatalf("want 2 razones (R5.1 y R5.2), got %v", rs)
		}
	})

	t.Run("un test que no resuelve se rechaza nombrando el criterio", func(t *testing.T) {
		dir := makeVerdictProject(t, structured,
			`{"feature":"checkout","requirements":{"R5":{
				"code":["src/order.go:CreateOrder"],"test":[],
				"scenarios":{
					"R5.1":{"test":["src/order_test.go:TestPersists"]},
					"R5.2":{"test":["src/order_test.go:TestNoExiste"]}},
				"status":"ok"}}}`)
		rs := reasonsFor(t, dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "R5.2") || !strings.Contains(rs[0], "does not resolve") {
			t.Errorf("want rechazo de R5.2 por ancla rota, got %v", rs)
		}
	})

	t.Run("un escenario huérfano delata un renumerado", func(t *testing.T) {
		// El trace ancla R5.3, que el spec ya no declara. O el criterio se
		// retiró y el ancla quedó colgando, o alguien renumeró — y renumerar
		// re-apunta anclas a otro caso sin que nada falle.
		dir := makeVerdictProject(t, structured,
			`{"feature":"checkout","requirements":{"R5":{
				"code":["src/order.go:CreateOrder"],"test":[],
				"scenarios":{
					"R5.1":{"test":["src/order_test.go:TestPersists"]},
					"R5.2":{"test":["src/order_test.go:TestEmptyCart"]},
					"R5.3":{"test":["src/order_test.go:TestPersists"]}},
				"status":"ok"}}}`)
		rs := reasonsFor(t, dir)
		if len(rs) != 1 || !strings.Contains(rs[0], "R5.3") || !strings.Contains(rs[0], "renumbered") {
			t.Errorf("want rechazo del huérfano R5.3, got %v", rs)
		}
	})

	t.Run("sin requirements.json todo cae a v1 — lectura quieta", func(t *testing.T) {
		dir := makeVerdictProject(t, structured,
			`{"feature":"checkout","requirements":{"R5":{
				"code":["src/order.go:CreateOrder"],
				"test":["src/order_test.go:TestPersists"],"status":"ok"}}}`)
		mustRemove(t, filepath.Join(dir, "specforge/features/checkout/requirements.json"))
		if rs := reasonsFor(t, dir); len(rs) != 0 {
			t.Errorf("want 0 razones (v1 por ausencia de spec), got %v", rs)
		}
	})
}

// TestVerdictContractRecorded — R10.
func TestVerdictContractRecorded(t *testing.T) {
	t.Run("acceptance estructurada → v2", func(t *testing.T) {
		dir := makeVerdictProject(t, `[{"id":"R5.1","text":"a"}]`,
			`{"feature":"checkout","requirements":{"R5":{"code":["src/order.go:CreateOrder"],"test":[],"status":"ok"}}}`)
		if got := verdictContract(dir, "checkout", "verdict"); got != contractV2 {
			t.Errorf("contract=%q, want %q", got, contractV2)
		}
	})

	t.Run("acceptance legada → v1", func(t *testing.T) {
		dir := makeVerdictProject(t, `["a"]`,
			`{"feature":"checkout","requirements":{"R5":{"code":["src/order.go:CreateOrder"],"test":[],"status":"ok"}}}`)
		if got := verdictContract(dir, "checkout", "verdict"); got != contractV1 {
			t.Errorf("contract=%q, want %q", got, contractV1)
		}
	})

	t.Run("sólo el verdict lleva contrato", func(t *testing.T) {
		dir := makeVerdictProject(t, `[{"id":"R5.1","text":"a"}]`, `{"feature":"checkout","requirements":{}}`)
		if got := verdictContract(dir, "checkout", "requirements"); got != "" {
			t.Errorf("contract=%q sobre la fase requirements, want vacío", got)
		}
	})

	t.Run("un verdict sellado antes del campo se lee como v1", func(t *testing.T) {
		f := feature{Gates: []gate{{Phase: "verdict", Result: "approve"}}}
		if got := f.sealedContract(); got != contractV1 {
			t.Errorf("sealedContract=%q, want %q", got, contractV1)
		}
	})

	t.Run("sin verdict no hay contrato", func(t *testing.T) {
		f := feature{Gates: []gate{{Phase: "requirements", Result: "approve"}}}
		if got := f.sealedContract(); got != "" {
			t.Errorf("sealedContract=%q, want vacío", got)
		}
	})

	t.Run("tras un amend vale el ÚLTIMO verdict", func(t *testing.T) {
		f := feature{Gates: []gate{
			{Phase: "verdict", Result: "approve", Contract: contractV1},
			{Phase: "requirements", Result: "approve"},
			{Phase: "verdict", Result: "approve", Contract: contractV2},
		}}
		if got := f.sealedContract(); got != contractV2 {
			t.Errorf("sealedContract=%q, want %q — el ledger es append-only, los previos son historia", got, contractV2)
		}
	})

	t.Run("el contrato NO entra en la cadena de integridad", func(t *testing.T) {
		// Meterlo cambiaría el hash de todas las entradas existentes y rompería
		// el ledger de cualquier proyecto vivo. Y no hace falta: el contrato se
		// computa del dato al aprobar, así que el campo es registro, no llave.
		a := gate{Phase: "verdict", Result: "approve", At: "t"}
		b := gate{Phase: "verdict", Result: "approve", At: "t", Contract: contractV2}
		if gateEntryHash(a) != gateEntryHash(b) {
			t.Error("agregar `contract` movió el hash del ledger — rompe todo proyecto existente")
		}
	})
}

// TestTraceScenarioValidation: los ids de escenario del trace pertenecen a su
// requisito.
func TestTraceScenarioValidation(t *testing.T) {
	check := func(scenarios string) []string {
		var tf traceFile
		mustUnmarshal(t, `{"feature":"f","requirements":{"R5":{
			"code":["src/x.go:F"],"test":[],"scenarios":`+scenarios+`,"status":"ok"}}}`, &tf)
		var rep report
		checkTrace(tf, &rep)
		return rep.errors
	}

	t.Run("id ajeno se rechaza", func(t *testing.T) {
		errs := check(`{"R6.1":{"test":["t.go:T"]}}`)
		if len(errs) != 1 || !strings.Contains(errs[0], "belongs to R6") {
			t.Errorf("want error de pertenencia, got %v", errs)
		}
	})

	t.Run("forma inválida se rechaza", func(t *testing.T) {
		if errs := check(`{"R5":{"test":["t.go:T"]}}`); len(errs) != 1 {
			t.Errorf("want 1 error de forma, got %v", errs)
		}
	})

	t.Run("id propio y bien formado pasa", func(t *testing.T) {
		if errs := check(`{"R5.1":{"test":["t.go:T"]}}`); len(errs) != 0 {
			t.Errorf("want 0 errores, got %v", errs)
		}
	})
}
