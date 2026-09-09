package andamio

import (
	"encoding/json"
	"slices"
	"testing"
)

// AGENTS.md tiene que quedar ENGANCHADO en la config de opencode, no suelto en
// el disco.
//
// MEDIDO EL 2026-09-08: sin esto, opencode recibe una idea, elige una skill por
// su descripción y NUNCA abre el orquestador. Las descripciones están siempre en
// contexto; un archivo que hay que ir a leer no les gana. Claude Code no tiene el
// problema porque auto-descubre CLAUDE.md.
func TestOpencodeEnganchaElOrquestador(t *testing.T) {
	cfg, hay := permisos["opencode"]
	if !hay {
		t.Fatal("opencode no tiene config de permisos")
	}
	var doc struct {
		Instructions []string `json:"instructions"`
	}
	if err := json.Unmarshal([]byte(cfg.Contenido), &doc); err != nil {
		t.Fatalf("la config de opencode no es JSON válido: %v", err)
	}
	if !slices.Contains(doc.Instructions, "AGENTS.md") {
		t.Errorf("opencode.json tiene que cargar AGENTS.md solo; trae %v", doc.Instructions)
	}
}

// Y las tres configs tienen que ser JSON válido: se escriben como string crudo,
// así que una coma de más no la atrapa el compilador.
func TestLasConfigsDeArnesSonJSONValido(t *testing.T) {
	for arnes, cfg := range permisos {
		var v any
		if err := json.Unmarshal([]byte(cfg.Contenido), &v); err != nil {
			t.Errorf("%s (%s): %v", arnes, cfg.Ruta, err)
		}
	}
}
