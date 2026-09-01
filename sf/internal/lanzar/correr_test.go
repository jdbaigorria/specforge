package lanzar

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func baseFicha() Ficha {
	return Ficha{Estado: "prd", Harness: "claude-code", Modelo: "opus", Skill: "sfp-po"}
}

// eco arma una línea que escupe por stdout lo que se le pase.
func eco(texto string) []string {
	return []string{"sh", "-c", "printf '%s' " + "'" + texto + "'"}
}

func TestCorrerDejaElRegistroYLaFicha(t *testing.T) {
	raiz := t.TempDir()

	f, err := Correr(eco(streamClaude), baseFicha(), Opciones{Raiz: raiz, Harness: "claude-code"})
	if err != nil {
		t.Fatal(err)
	}

	if f.Salida != 0 || f.Fin != FinTermino {
		t.Errorf("salida=%d fin=%q", f.Salida, f.Fin)
	}
	if f.ID == "" || !strings.HasSuffix(f.ID, "-prd") {
		t.Errorf("el id no nombra el estado: %q", f.ID)
	}
	if f.DuroMs < 0 {
		t.Errorf("duró %d ms", f.DuroMs)
	}

	dir := filepath.Join(raiz, Carpeta)
	if b, err := os.ReadFile(filepath.Join(dir, f.Registro)); err != nil {
		t.Errorf("no dejó el registro: %v", err)
	} else if !strings.Contains(string(b), "total_cost_usd") {
		t.Errorf("el registro no es el stream:\n%s", b)
	}

	// Y la ficha en disco tiene que ser la misma que devolvió.
	b, err := os.ReadFile(filepath.Join(dir, f.ID+".json"))
	if err != nil {
		t.Fatalf("no dejó la ficha: %v", err)
	}
	var leida Ficha
	if err := json.Unmarshal(b, &leida); err != nil {
		t.Fatalf("la ficha no es JSON válido: %v", err)
	}
	if leida.ID != f.ID || leida.Modelo != "opus" {
		t.Errorf("la ficha en disco no coincide: %+v", leida)
	}
}

// Lo que el arnés dijo en su stream entra a la ficha, normalizado.
func TestLaFichaTraeLoQueSeSacoDelStream(t *testing.T) {
	raiz := t.TempDir()

	f, err := Correr(eco(streamClaude), baseFicha(), Opciones{Raiz: raiz, Harness: "claude-code"})
	if err != nil {
		t.Fatal(err)
	}
	if f.CostoUSD == nil || *f.CostoUSD != 0.0951614 {
		t.Errorf("costo %v", f.CostoUSD)
	}
	if f.CargoLaSkill == nil || !*f.CargoLaSkill {
		t.Errorf("cargo_la_skill %v", f.CargoLaSkill)
	}
	if f.Dijo != "Actores" {
		t.Errorf("dijo %q", f.Dijo)
	}
}

// LO QUE EL ARNÉS NO DA NO APARECE EN EL JSON, ni siquiera en cero.
//
// Es la regla de headless.md §8.4, y se comprueba sobre el archivo y no sobre el
// struct: un `omitempty` mal puesto se ve ahí y en ningún otro lado.
func TestLoAusenteNoSeSerializa(t *testing.T) {
	raiz := t.TempDir()

	// Command Code no dice cuánto costó.
	f, err := Correr(eco(streamCommandCode), baseFicha(), Opciones{Raiz: raiz, Harness: "commandcode"})
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(raiz, Carpeta, f.ID+".json"))
	if strings.Contains(string(b), "costo_usd") {
		t.Errorf("escribió un costo que el arnés no da:\n%s", b)
	}
	// Y los que sí da, están.
	if !strings.Contains(string(b), "tokens") {
		t.Errorf("perdió los tokens:\n%s", b)
	}
}

// UN EXIT DISTINTO DE CERO NO ES UN ERROR DE SF: es un dato de la corrida.
//
// Y `fin` sigue diciendo "termino", porque habla del PROCESO y no del trabajo:
// el proceso salió por su cuenta. Si el trabajo vale lo dice `sf done`.
func TestUnaSalidaDistintaDeCeroSeRegistraYNoRompe(t *testing.T) {
	raiz := t.TempDir()

	f, err := Correr([]string{"sh", "-c", "exit 4"}, baseFicha(), Opciones{Raiz: raiz, Harness: "commandcode"})
	if err != nil {
		t.Fatalf("un exit != 0 del arnés no es un error de sf: %v", err)
	}
	if f.Salida != 4 {
		t.Errorf("salida %d, esperaba 4", f.Salida)
	}
	if f.Fin != FinTermino {
		t.Errorf("fin %q: el proceso terminó por su cuenta", f.Fin)
	}
}

// EL TOPE DE ESPERA, Y QUE EL REGISTRO PARCIAL SE QUEDE.
//
// El momento en que querés el stream es justamente cuando la corrida NO terminó.
// Borrarlo al matar el proceso sería tirar la única evidencia que queda.
func TestSeAgotaLaEsperaYElRegistroParcialQueda(t *testing.T) {
	raiz := t.TempDir()

	inicio := time.Now()
	f, err := Correr(
		[]string{"sh", "-c", `printf '{"type":"system"}\n'; sleep 30`},
		baseFicha(),
		Opciones{Raiz: raiz, Harness: "claude-code", Espera: 500 * time.Millisecond},
	)
	if err != nil {
		t.Fatalf("agotar la espera no es un error de sf: %v", err)
	}
	if d := time.Since(inicio); d > 20*time.Second {
		t.Errorf("no lo mató: tardó %s", d)
	}
	if f.Fin != FinTimeout {
		t.Errorf("fin %q, esperaba %q", f.Fin, FinTimeout)
	}

	b, err := os.ReadFile(filepath.Join(raiz, Carpeta, f.Registro))
	if err != nil {
		t.Fatalf("borró el registro parcial: %v", err)
	}
	if !strings.Contains(string(b), "system") {
		t.Errorf("el registro parcial quedó vacío:\n%s", b)
	}
}

// Espera 0 es "sin tope", no "sin tiempo". Sin esto, un `--espera=0` mataría el
// proceso instantáneamente, que es lo contrario de lo que pide.
func TestEsperaCeroEsSinTope(t *testing.T) {
	raiz := t.TempDir()

	f, err := Correr(eco("hola"), baseFicha(), Opciones{Raiz: raiz, Harness: "claude-code", Espera: 0})
	if err != nil {
		t.Fatal(err)
	}
	if f.Fin != FinTermino {
		t.Errorf("fin %q: espera 0 tiene que dejarlo correr", f.Fin)
	}
}

// Un binario que no existe SÍ es un error de sf: no se pudo lanzar nada, y no
// hay corrida de la que informar.
func TestUnBinarioQueNoExisteEsError(t *testing.T) {
	raiz := t.TempDir()

	f, err := Correr([]string{"sf-no-existe-xyz"}, baseFicha(), Opciones{Raiz: raiz, Harness: "claude-code"})
	if err == nil {
		t.Fatal("no falló al lanzar un binario inexistente")
	}
	if f.Fin != FinError {
		t.Errorf("fin %q, esperaba %q", f.Fin, FinError)
	}
}

// El stderr no entra al registro —lo rompería como JSONL— pero tampoco se tira:
// es lo primero que se mira cuando algo salió mal.
func TestElStderrNoEnsuciaElRegistroPeroSeGuarda(t *testing.T) {
	raiz := t.TempDir()

	f, err := Correr(
		[]string{"sh", "-c", `printf '{"type":"system"}\n'; echo "algo se rompio" >&2; exit 1`},
		baseFicha(), Opciones{Raiz: raiz, Harness: "claude-code"},
	)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(filepath.Join(raiz, Carpeta, f.Registro))
	if strings.Contains(string(b), "algo se rompio") {
		t.Errorf("mezcló stderr en el JSONL:\n%s", b)
	}
	if !strings.Contains(f.Error, "algo se rompio") {
		t.Errorf("perdió el stderr: %q", f.Error)
	}
}
