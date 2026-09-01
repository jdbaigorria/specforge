package lanzar

import (
	"bufio"
	"encoding/json"
	"io"
	"slices"
)

// ────────────────────────────────────────────────────────────────────────────
// LEER EL STREAM — y por qué hay tres lectores y no uno
// ────────────────────────────────────────────────────────────────────────────
//
// Los tres arneses emiten JSON por línea y ninguno emite lo mismo. Medido el
// 2026-08-31, corriendo `Usá la skill sfp-po` en carpetas descartables:
//
//	                  la carga de skill        el total de la corrida
//	claude-code       tool_use name=Skill      type=result: total_cost_usd, usage
//	opencode          part.tool=skill          NO HAY — cost y tokens POR PASO
//	commandcode       toolName=activate_skill  type=result: usage, finalText
//
// Tres nombres para la misma cosa —`Skill`, `skill`, `activate_skill`— y uno de
// los tres sin resultado de corrida. Por eso `sf` no lee la carta del arnés:
// ESCRIBE LA SUYA, normalizando los tres.
//
// ────────────────────────────────────────────────────────────────────────────
// LA REGLA: LO QUE EL ARNÉS NO DA, NO ESTÁ
// ────────────────────────────────────────────────────────────────────────────
//
// No va en cero, no va en false, no va. Un costo de 0 y "este arnés no dice
// cuánto costó" son dos cosas distintas, y escribir la primera cuando pasa la
// segunda es mentir en un archivo que alguien va a leer para decidir con qué
// modelo sigue. Por eso los campos opcionales son punteros.
//
// Es la misma distinción que el catálogo ya hace con `esfuerzo`: vacío es "el
// que traiga el modelo", que NO es "bajo".

// Tokens son los que consumió la corrida.
type Tokens struct {
	Entrada int `json:"entrada"`
	Salida  int `json:"salida"`
}

// Resumen es lo que se pudo sacar del stream. Todo lo opcional puede faltar.
type Resumen struct {
	CostoUSD     *float64
	Tokens       *Tokens
	CargoLaSkill *bool

	// Herramientas son las que usó, sin repetir y en orden de aparición.
	Herramientas []string

	// Dijo es su última palabra. NO ES EVIDENCIA de nada: que un modelo diga
	// "listo" no cierra un paso — la compuerta sigue siendo la compuerta. Sirve
	// para contarle a Javier qué pasó, sobre todo cuando algo salió mal.
	Dijo string
}

// lectores es cómo se lee el stream de cada arnés.
var lectores = map[string]func(*Resumen, map[string]any){
	"claude-code": leerClaude,
	"opencode":    leerOpencode,
	"commandcode": leerCommandCode,
}

// Resumir recorre el stream y saca lo que ese arnés dé.
//
// Nunca falla: un stream ilegible es un resumen vacío, no un error. La corrida
// ya pasó, y no poder leerla no cambia lo que sf sí sabe —qué lanzó, cuánto
// tardó, con qué código salió—, que es lo que igual va a la ficha.
func Resumir(harness string, r io.Reader) Resumen {
	var res Resumen
	leer, hay := lectores[harness]
	if !hay {
		return res
	}

	sc := bufio.NewScanner(r)
	// Una línea de Command Code puede traer el contenido entero del mensaje
	// repetido en cada `message_update`, y pasa holgado los 64K del default.
	sc.Buffer(make([]byte, 0, 64*1024), 8*1024*1024)

	for sc.Scan() {
		var ev map[string]any
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		leer(&res, ev)
	}
	return res
}

// herramienta anota una, sin repetir, y de paso contesta lo de la skill.
func (r *Resumen) herramienta(nombre string, esSkill bool) {
	if nombre == "" {
		return
	}
	if !slices.Contains(r.Herramientas, nombre) {
		r.Herramientas = append(r.Herramientas, nombre)
	}
	if esSkill {
		si := true
		r.CargoLaSkill = &si
	}
}

// vistoUnEvento marca que el stream SÍ se pudo leer.
//
// Es lo que separa "no cargó la skill" de "no sé si la cargó": sin un solo
// evento reconocido, contestar `false` sería afirmar algo sobre una corrida que
// no pudimos mirar.
func (r *Resumen) vistoUnEvento() {
	if r.CargoLaSkill == nil {
		no := false
		r.CargoLaSkill = &no
	}
}

func numero(m map[string]any, clave string) (float64, bool) {
	v, hay := m[clave].(float64)
	return v, hay
}

func objeto(m map[string]any, clave string) map[string]any {
	o, _ := m[clave].(map[string]any)
	return o
}

// ────────────────────────────────────────────────────────────────────────────

func leerClaude(r *Resumen, ev map[string]any) {
	switch ev["type"] {
	case "assistant":
		r.vistoUnEvento()
		contenido, _ := objeto(ev, "message")["content"].([]any)
		for _, c := range contenido {
			parte, _ := c.(map[string]any)
			if parte["type"] != "tool_use" {
				continue
			}
			nombre, _ := parte["name"].(string)
			r.herramienta(nombre, nombre == "Skill")
		}

	case "result":
		r.vistoUnEvento()
		if c, hay := numero(ev, "total_cost_usd"); hay {
			r.CostoUSD = &c
		}
		if u := objeto(ev, "usage"); u != nil {
			e, _ := numero(u, "input_tokens")
			s, _ := numero(u, "output_tokens")
			r.Tokens = &Tokens{Entrada: int(e), Salida: int(s)}
		}
		if t, hay := ev["result"].(string); hay {
			r.Dijo = t
		}
	}
}

// leerOpencode suma, porque opencode NO tiene resultado de corrida: da `cost` y
// `tokens` por paso, y el total es la suma.
func leerOpencode(r *Resumen, ev map[string]any) {
	parte := objeto(ev, "part")
	if parte == nil {
		return
	}
	r.vistoUnEvento()

	switch ev["type"] {
	case "tool_use":
		nombre, _ := parte["tool"].(string)
		r.herramienta(nombre, nombre == "skill")

	case "step_finish":
		if c, hay := numero(parte, "cost"); hay {
			total := c
			if r.CostoUSD != nil {
				total += *r.CostoUSD
			}
			r.CostoUSD = &total
		}
		if t := objeto(parte, "tokens"); t != nil {
			e, _ := numero(t, "input")
			s, _ := numero(t, "output")
			if r.Tokens == nil {
				r.Tokens = &Tokens{}
			}
			r.Tokens.Entrada += int(e)
			r.Tokens.Salida += int(s)
		}

	case "text":
		// La última gana: es su palabra final.
		if t, hay := parte["text"].(string); hay && t != "" {
			r.Dijo = t
		}
	}
}

func leerCommandCode(r *Resumen, ev map[string]any) {
	switch ev["type"] {
	case "event":
		e := objeto(ev, "event")
		if e == nil {
			return
		}
		r.vistoUnEvento()
		if nombre, hay := e["toolName"].(string); hay {
			r.herramienta(nombre, nombre == "activate_skill")
		}

	case "result":
		r.vistoUnEvento()
		// Command Code NO dice cuánto costó. Queda ausente, no en cero.
		if u := objeto(ev, "usage"); u != nil {
			e, _ := numero(u, "inputTokens")
			s, _ := numero(u, "outputTokens")
			r.Tokens = &Tokens{Entrada: int(e), Salida: int(s)}
		}
		if t, hay := ev["finalText"].(string); hay {
			r.Dijo = t
		}
	}
}
