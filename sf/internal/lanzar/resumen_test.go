package lanzar

import (
	"slices"
	"strings"
	"testing"
)

// Los tres fixtures son RECORTES DE CAPTURAS REALES del 2026-08-31, no
// invenciones. Cada arnés corrió `Usá la skill sfp-po` en una carpeta
// descartable y esto es lo que emitió, con los campos largos podados.

const streamClaude = `{"type":"system","subtype":"init","session_id":"fa04"}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Skill","input":{"skill":"sfp-po"}}]}}
{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{"command":"ls"}}]}}
{"type":"result","subtype":"success","stop_reason":"end_turn","session_id":"fa04","total_cost_usd":0.0951614,"result":"Actores","usage":{"input_tokens":4,"output_tokens":154}}
`

const streamOpencode = `{"type":"step_start","part":{"type":"step-start"}}
{"type":"tool_use","part":{"type":"tool","tool":"skill","state":{"input":{"name":"sfp-po"}}}}
{"type":"step_finish","part":{"type":"step-finish","reason":"tool-calls","tokens":{"input":81,"output":39},"cost":0.002}}
{"type":"tool_use","part":{"type":"tool","tool":"write","state":{"input":{"filePath":"x"}}}}
{"type":"text","part":{"type":"text","text":"Actores"}}
{"type":"step_finish","part":{"type":"step-finish","reason":"stop","tokens":{"input":12,"output":7},"cost":0.001}}
`

const streamCommandCode = `{"type":"event","event":{"type":"run_start","sessionId":"7873"}}
{"type":"event","event":{"type":"tool_queued","toolName":"activate_skill","input":{"name":"sfp-po"}}}
{"type":"event","event":{"type":"tool_running","toolName":"activate_skill"}}
{"type":"event","event":{"type":"tool_queued","toolName":"write_file","input":{"path":"x"}}}
{"type":"result","subtype":"success","stopReason":"end_turn","usage":{"inputTokens":16648,"outputTokens":71},"durationMs":26944,"finalText":"Actores"}
`

func TestResumirClaudeCode(t *testing.T) {
	r := Resumir("claude-code", strings.NewReader(streamClaude))

	if r.CostoUSD == nil || *r.CostoUSD != 0.0951614 {
		t.Errorf("costo %v", r.CostoUSD)
	}
	if r.Tokens == nil || r.Tokens.Entrada != 4 || r.Tokens.Salida != 154 {
		t.Errorf("tokens %+v", r.Tokens)
	}
	if r.Dijo != "Actores" {
		t.Errorf("dijo %q", r.Dijo)
	}
	if !slices.Contains(r.Herramientas, "Bash") {
		t.Errorf("herramientas %v", r.Herramientas)
	}
}

func TestResumirOpencodeSumaLosPasos(t *testing.T) {
	r := Resumir("opencode", strings.NewReader(streamOpencode))

	// opencode NO tiene resultado a nivel corrida: da `cost` y `tokens` por
	// paso, así que sumarlos es la única forma de tener el total.
	if r.CostoUSD == nil || *r.CostoUSD < 0.0029 || *r.CostoUSD > 0.0031 {
		t.Errorf("costo %v — tenía que sumar los dos pasos", r.CostoUSD)
	}
	if r.Tokens == nil || r.Tokens.Entrada != 93 || r.Tokens.Salida != 46 {
		t.Errorf("tokens %+v — tenía que sumar los dos pasos", r.Tokens)
	}
	if r.Dijo != "Actores" {
		t.Errorf("dijo %q", r.Dijo)
	}
	if !slices.Contains(r.Herramientas, "write") {
		t.Errorf("herramientas %v", r.Herramientas)
	}
}

func TestResumirCommandCode(t *testing.T) {
	r := Resumir("commandcode", strings.NewReader(streamCommandCode))

	if r.Tokens == nil || r.Tokens.Entrada != 16648 || r.Tokens.Salida != 71 {
		t.Errorf("tokens %+v", r.Tokens)
	}
	if r.Dijo != "Actores" {
		t.Errorf("dijo %q", r.Dijo)
	}
	// Command Code NO dice cuánto costó. Ausente, no cero.
	if r.CostoUSD != nil {
		t.Errorf("se inventó un costo que el arnés no da: %v", *r.CostoUSD)
	}
	if !slices.Contains(r.Herramientas, "write_file") {
		t.Errorf("herramientas %v", r.Herramientas)
	}
	// Una herramienta que aparece en tool_queued Y en tool_running se cuenta una vez.
	var veces int
	for _, h := range r.Herramientas {
		if h == "activate_skill" {
			veces++
		}
	}
	if veces > 1 {
		t.Errorf("repitió una herramienta: %v", r.Herramientas)
	}
}

// LA CARGA DE LA SKILL SE LLAMA DISTINTO EN CADA UNO, y eso es exactamente lo que
// hay que medir en vez de adivinar: `Skill` en Claude Code, `skill` en opencode,
// `activate_skill` en Command Code.
//
// Es la única señal barata de que el paso corrió como se pidió: `sf` apunta a la
// skill, y apuntar es una instrucción que se puede ignorar — opencode ya ignoró
// un `via: subagente` el 2026-08-31.
func TestDiceSiSeCargoLaSkillEnLosTres(t *testing.T) {
	for _, c := range []struct {
		harness, stream string
	}{
		{"claude-code", streamClaude},
		{"opencode", streamOpencode},
		{"commandcode", streamCommandCode},
	} {
		r := Resumir(c.harness, strings.NewReader(c.stream))
		if r.CargoLaSkill == nil {
			t.Errorf("%s: no supo contestar si cargó la skill", c.harness)
			continue
		}
		if !*r.CargoLaSkill {
			t.Errorf("%s: la cargó y dijo que no", c.harness)
		}
	}
}

// Y cuando NO la cargó, dice que no. No es lo mismo que "no sé".
func TestDiceQueNoCuandoNoCargoLaSkill(t *testing.T) {
	sin := `{"type":"assistant","message":{"content":[{"type":"tool_use","name":"Bash","input":{}}]}}
{"type":"result","subtype":"success","usage":{"input_tokens":1,"output_tokens":1}}
`
	r := Resumir("claude-code", strings.NewReader(sin))
	if r.CargoLaSkill == nil {
		t.Fatal("contestó \"no sé\" sobre un stream que sí pudo leer")
	}
	if *r.CargoLaSkill {
		t.Error("dijo que la cargó y no hay ningún evento de skill")
	}
}

// UN STREAM QUE NO SE ENTIENDE NO SE INVENTA. Todo queda ausente, que es
// distinto de cero y de false: la regla de headless.md §8.4.
func TestUnStreamIlegibleDejaTodoAusente(t *testing.T) {
	for _, basura := range []string{"", "no soy json\n", "{}\n{\"type\":\"otra-cosa\"}\n"} {
		r := Resumir("claude-code", strings.NewReader(basura))
		if r.CostoUSD != nil || r.Tokens != nil || r.CargoLaSkill != nil {
			t.Errorf("%q: se inventó algo: %+v", basura, r)
		}
	}
}

// Un arnés que no sabemos leer tampoco inventa: la ficha queda con lo que sf sí
// sabe —qué lanzó, cuánto tardó, con qué código salió— y sin lo que no.
func TestUnArnesDesconocidoNoInventaNada(t *testing.T) {
	r := Resumir("emacs", strings.NewReader(streamClaude))
	if r.CostoUSD != nil || r.Tokens != nil || r.CargoLaSkill != nil || len(r.Herramientas) > 0 {
		t.Errorf("leyó un stream de un arnés que no conoce: %+v", r)
	}
}
