package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

// ----------------------------------------------------------------------------
// Cadena de integridad del ledger (R1 de EVALUACION-PLATAFORMA §3).
//
// El problema: la Capa 1 (hook deny) PREVIENE la escritura directa de
// features.json, pero la prevención en el borde del hook es porosa por
// construcción (shell Turing-completo, editores humanos, otros agentes sin
// hooks). La respuesta arquitectónica no es más regexes: es DETECCIÓN
// INFALSIFICABLE. Este archivo la implementa.
//
// La idea (la misma de git y de cualquier blockchain, en miniatura):
//
//   cada entrada de gate lleva `prev` = hash sha256 de la ENTRADA ANTERIOR
//   completa (incluido SU prev). La primera entrada lleva el centinela
//   "genesis".
//
// Consecuencia: forjar o editar una entrada INTERMEDIA rompe visiblemente la
// cadena — el `prev` de la entrada siguiente ya no coincide. Y como cada
// comando que avanza el pipeline (`sf gate approve`, `sf next`, `sf feature
// archive`) VALIDA la cadena antes de operar, el bypass de escritura pasa de
// "imposible de prevenir al 100%" a "inútil": se detecta al próximo comando.
//
// Compatibilidad hacia atrás: los gates creados ANTES de esta capa no traen
// `prev` (legacy). Se toleran mientras estén al PRINCIPIO del ledger. Una vez
// que aparece el primer gate encadenado, todo lo que sigue debe encadenar —
// si no, un forjador podría apendear entradas "legacy" a voluntad. `sf migrate`
// puede rellenar la cadena de un ledger legacy legítimo (C4).
// ----------------------------------------------------------------------------

// genesisPrev es el centinela del primer eslabón: la primera entrada de la
// cadena no tiene anterior, así que declara "genesis" explícitamente (un ""
// significaría "gate legacy sin cadena", que es otra cosa).
const genesisPrev = "genesis"

// gateEntryHash computa el sha256 (hex) de UNA entrada del ledger. Hashea todos
// los campos —incluido el propio Prev— separados por \x00 (un byte que no
// aparece en texto normal, así "a"+"bc" no colisiona con "ab"+"c"). Incluir el
// Prev propio es lo que forma la CADENA: el hash de la entrada N depende del de
// la N-1, que depende del de la N-2… igual que un commit de git incluye el SHA
// de su parent.
func gateEntryHash(g gate) string {
	h := sha256.New()
	for _, field := range []string{g.Phase, g.Result, g.By, g.At, g.Comment, g.Hash, g.Prev} {
		h.Write([]byte(field))
		h.Write([]byte{0})
	}
	return hex.EncodeToString(h.Sum(nil))
}

// nextPrev devuelve el `prev` que debe llevar la PRÓXIMA entrada del ledger de
// esta feature: el hash de la última entrada existente, o "genesis" si el
// ledger está vacío. Lo usa `sf gate approve` al apendear.
func nextPrev(f *feature) string {
	if len(f.Gates) == 0 {
		return genesisPrev
	}
	return gateEntryHash(f.Gates[len(f.Gates)-1])
}

// ledgerProblems valida la cadena de UNA feature y devuelve la lista de
// problemas (vacía = ledger íntegro). Es una función PURA: no lee disco, no
// imprime — fácil de testear y de llamar desde cualquier comando.
//
// Reglas:
//  1. Gates sin `prev` (legacy) se toleran SOLO antes del primer gate
//     encadenado. Después, un gate sin `prev` es un apéndice fuera del CLI.
//  2. El `prev` de cada gate encadenado debe coincidir con el hash de la
//     entrada anterior (o "genesis" si es la primera).
//  3. Un gate encadenado debe estar firmado: `by` no vacío y `at` RFC3339
//     válido. (Gate sin autor/fecha = gate inválido, EVALUACION §2.4.)
func ledgerProblems(f *feature) []string {
	var probs []string
	chained := false // ¿ya empezó la era encadenada en este ledger?

	for i := range f.Gates {
		g := &f.Gates[i]
		label := fmt.Sprintf("gate %d (%s/%s)", i+1, g.Phase, orDash(g.Result))

		if g.Prev == "" {
			if chained {
				// Regla 1: después del primer eslabón, no hay vuelta al modo legacy.
				probs = append(probs, label+": no `prev` after the chain began — appended outside the CLI")
			}
			continue // legacy tolerado (pre-cadena)
		}
		chained = true

		// Regla 2: el eslabón tiene que engancharse con la entrada anterior.
		want := genesisPrev
		if i > 0 {
			want = gateEntryHash(f.Gates[i-1])
		}
		if g.Prev != want {
			probs = append(probs, label+": `prev` does not match the previous entry — chain broken (tampered or forged)")
		}

		// Regla 3: firmado. Sin autor o sin timestamp válido no certifica nada.
		if strings.TrimSpace(g.By) == "" {
			probs = append(probs, label+": missing `by` (unsigned gate)")
		}
		if _, err := time.Parse(time.RFC3339, g.At); err != nil {
			probs = append(probs, label+": `at` is not a valid RFC3339 timestamp")
		}
	}
	return probs
}

// ledgerProblemsAll corre ledgerProblems sobre TODAS las features y devuelve
// los problemas prefijados con el nombre de la feature. La usan los comandos
// de vista global (`sf status`, `sf verify`).
func ledgerProblemsAll(ff featuresFile) []string {
	var probs []string
	for i := range ff.Features {
		f := &ff.Features[i]
		for _, p := range ledgerProblems(f) {
			probs = append(probs, f.Name+": "+p)
		}
	}
	return probs
}

// refuseOnBrokenLedger es el guard compartido por los comandos que AVANZAN el
// pipeline: imprime los problemas + la salida (sf recover / git) y devuelve
// true si hay que rehusarse. Centralizado para que el mensaje sea idéntico en
// todos los puntos de entrada.
func refuseOnBrokenLedger(cmd string, f *feature) bool {
	probs := ledgerProblems(f)
	if len(probs) == 0 {
		return false
	}
	fmt.Printf("%s: REFUSED — the gate ledger for %q fails integrity validation:\n", cmd, f.Name)
	for _, p := range probs {
		fmt.Printf("  - %s\n", p)
	}
	fmt.Println("\nThe ledger was modified outside the CLI. Restore specforge/features.json from git")
	fmt.Println("history (or re-approve the gates legitimately), then retry. See `sf recover`.")
	return true
}
