package main

import (
	"strings"
	"testing"
)

// ----------------------------------------------------------------------------
// Tests de la cadena de integridad del ledger (integrity.go, R1).
//
// Casi todos prueban ledgerProblems, que es PURA: armamos ledgers en memoria
// (legítimos, forjados, legacy) y verificamos qué detecta. Sin disco, sin CLI.
// ----------------------------------------------------------------------------

// chainedGate arma un gate firmado y encadenado sobre el ledger actual de f
// (mismo camino que usa gateApprove: nextPrev ANTES del append).
func chainedGate(f *feature, phase string) gate {
	return gate{
		Phase:  phase,
		Result: "approve",
		By:     "tester",
		At:     "2026-07-05T12:00:00Z",
		Prev:   nextPrev(f),
	}
}

func TestLedgerValidChain(t *testing.T) {
	f := feature{Name: "demo"}
	for _, phase := range []string{"lane", "requirements", "design"} {
		f.Gates = append(f.Gates, chainedGate(&f, phase))
	}
	if probs := ledgerProblems(&f); len(probs) != 0 {
		t.Fatalf("valid chain reported problems: %v", probs)
	}
}

func TestLedgerLegacyGatesTolerated(t *testing.T) {
	// Gates pre-cadena (sin prev): tolerados mientras estén al principio.
	f := feature{Name: "demo", Gates: []gate{
		{Phase: "lane", Result: "approve", By: "user", At: "2026-01-01T00:00:00Z"},
		{Phase: "requirements", Result: "approve", By: "user", At: "2026-01-02T00:00:00Z"},
	}}
	if probs := ledgerProblems(&f); len(probs) != 0 {
		t.Fatalf("legacy ledger reported problems: %v", probs)
	}

	// Y la transición legacy→encadenado también vale: el primer gate encadenado
	// engancha con el hash del último legacy.
	f.Gates = append(f.Gates, chainedGate(&f, "design"))
	if probs := ledgerProblems(&f); len(probs) != 0 {
		t.Fatalf("legacy→chained transition reported problems: %v", probs)
	}
}

func TestLedgerTamperedEntryBreaksChain(t *testing.T) {
	f := feature{Name: "demo"}
	for _, phase := range []string{"lane", "requirements", "design"} {
		f.Gates = append(f.Gates, chainedGate(&f, phase))
	}
	// Forjamos la entrada del medio (como haría un edit a mano del JSON):
	// cambia su contenido → cambia su hash → el prev del gate siguiente ya no
	// coincide.
	f.Gates[1].Result = "approve-with-notes"

	probs := ledgerProblems(&f)
	if len(probs) == 0 {
		t.Fatal("tampered middle entry not detected")
	}
	if !strings.Contains(strings.Join(probs, "\n"), "chain broken") {
		t.Fatalf("expected a chain-broken problem, got: %v", probs)
	}
}

func TestLedgerUnchainedAppendDetected(t *testing.T) {
	// Un gate SIN prev apendeado DESPUÉS de que la cadena empezó = escritura
	// fuera del CLI (el atajo del forjador que finge ser "legacy").
	f := feature{Name: "demo"}
	f.Gates = append(f.Gates, chainedGate(&f, "lane"))
	f.Gates = append(f.Gates, gate{
		Phase: "verdict", Result: "approve", By: "user", At: "2026-07-05T12:00:00Z",
	})

	probs := ledgerProblems(&f)
	if len(probs) == 0 {
		t.Fatal("unchained append after chained era not detected")
	}
	if !strings.Contains(strings.Join(probs, "\n"), "appended outside the CLI") {
		t.Fatalf("expected an unchained-append problem, got: %v", probs)
	}
}

func TestLedgerUnsignedChainedGateDetected(t *testing.T) {
	f := feature{Name: "demo"}
	g := chainedGate(&f, "lane")
	g.By = ""            // sin autor
	g.At = "yesterday"   // timestamp no RFC3339
	f.Gates = append(f.Gates, g)

	probs := strings.Join(ledgerProblems(&f), "\n")
	if !strings.Contains(probs, "unsigned") {
		t.Fatalf("missing `by` not detected: %v", probs)
	}
	if !strings.Contains(probs, "RFC3339") {
		t.Fatalf("invalid `at` not detected: %v", probs)
	}
}

func TestGateEntryHashSensitivity(t *testing.T) {
	// El hash debe cambiar si CUALQUIER campo cambia (incluido prev), y los
	// separadores \x00 deben impedir colisiones por corrimiento de contenido.
	base := gate{Phase: "lane", Result: "approve", By: "u", At: "2026-01-01T00:00:00Z", Prev: genesisPrev}
	h := gateEntryHash(base)

	mut := base
	mut.Comment = "x"
	if gateEntryHash(mut) == h {
		t.Fatal("hash did not change when a field changed")
	}

	// Corrimiento: "ab"+"c" vs "a"+"bc" no deben colisionar.
	a := gate{Phase: "ab", Result: "c"}
	b := gate{Phase: "a", Result: "bc"}
	if gateEntryHash(a) == gateEntryHash(b) {
		t.Fatal("field-shifting collision — separator missing")
	}
}

func TestNextPrev(t *testing.T) {
	f := feature{Name: "demo"}
	if got := nextPrev(&f); got != genesisPrev {
		t.Fatalf("empty ledger: want %q, got %q", genesisPrev, got)
	}
	f.Gates = append(f.Gates, chainedGate(&f, "lane"))
	if got := nextPrev(&f); got != gateEntryHash(f.Gates[0]) {
		t.Fatal("nextPrev must be the hash of the last entry")
	}
}
