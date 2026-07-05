package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// ----------------------------------------------------------------------------
// Tests de higiene de estado del hook (C6): TTL de sesiones, cap de session.md,
// presupuesto de inyección.
// ----------------------------------------------------------------------------

func TestPruneStaleSessions(t *testing.T) {
	now := time.Date(2026, 7, 5, 12, 0, 0, 0, time.UTC)
	st := hookState{Sessions: map[string]sessionEntry{
		"fresh":  {LastSeen: now.Add(-time.Hour).Format(time.RFC3339)},
		"edge":   {LastSeen: now.Add(-sessionTTL + time.Minute).Format(time.RFC3339)},
		"stale":  {LastSeen: now.Add(-sessionTTL - time.Hour).Format(time.RFC3339)},
		"legacy": {}, // sin last_seen (pre-C6): se poda
	}}
	st.pruneStaleSessions(now)

	for _, keep := range []string{"fresh", "edge"} {
		if _, ok := st.Sessions[keep]; !ok {
			t.Errorf("session %q should survive the prune", keep)
		}
	}
	for _, gone := range []string{"stale", "legacy"} {
		if _, ok := st.Sessions[gone]; ok {
			t.Errorf("session %q should be pruned", gone)
		}
	}
}

func TestSetSessionStampsLastSeen(t *testing.T) {
	var st hookState
	st.setSession("s1", sessionEntry{TurnsSinceFull: 3})
	if st.Sessions["s1"].LastSeen == "" {
		t.Fatal("setSession must stamp last_seen")
	}
}

func TestHookStatePrunesOnWrite(t *testing.T) {
	proj := t.TempDir()
	sf := filepath.Join(proj, "specforge")
	mustWrite(t, filepath.Join(sf, ".keep"), "")

	old := time.Now().UTC().Add(-sessionTTL - time.Hour).Format(time.RFC3339)
	writeHookState(sf, hookState{Sessions: map[string]sessionEntry{
		"ancient": {LastSeen: old},
		"current": {LastSeen: time.Now().UTC().Format(time.RFC3339)},
	}})

	got := readHookState(sf)
	if _, ok := got.Sessions["ancient"]; ok {
		t.Error("expired session must be pruned on write")
	}
	if _, ok := got.Sessions["current"]; !ok {
		t.Error("live session must survive the write")
	}
}

func TestCapSessionMD(t *testing.T) {
	proj := t.TempDir()
	dir := filepath.Join(proj, "specforge", ".state")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, "session.md")

	// Un session.md muy por encima del techo, con líneas numeradas para poder
	// verificar que sobrevive la COLA (lo reciente).
	var b strings.Builder
	for i := 0; b.Len() < sessionMDMaxBytes*2; i++ {
		fmt := "<!-- marker %06d -->\n"
		b.WriteString(strings.Replace(fmt, "%06d", padInt(i), 1))
	}
	mustWrite(t, path, b.String())

	markSession(proj, "one more")

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) > sessionMDMaxBytes {
		t.Fatalf("session.md not capped: %d bytes", len(data))
	}
	s := string(data)
	if !strings.Contains(s, "session.md capped") {
		t.Fatal("missing truncation marker")
	}
	if !strings.Contains(s, "one more") {
		t.Fatal("the freshly appended marker must survive the cap")
	}
	if strings.Contains(s, "marker 000000") {
		t.Fatal("oldest content should have been dropped")
	}
}

// padInt: entero a 6 dígitos sin fmt.Sprintf en el hot loop del test.
func padInt(i int) string {
	s := "000000" + itoa(i)
	return s[len(s)-6:]
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for ; i > 0; i /= 10 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
	}
	return string(digits)
}

func TestBudgetTrim(t *testing.T) {
	// Bajo presupuesto: intacto.
	if got := budgetTrim("short", 100, false, "x.md"); got != "short" {
		t.Fatalf("under-budget content must pass through, got %q", got)
	}

	long := strings.Repeat("line-head\n", 50) + strings.Repeat("line-tail\n", 50)

	// keepTail=false (curados): sobrevive la cabeza, marcador al final.
	head := budgetTrim(long, 200, false, "learnings.md")
	if !strings.HasPrefix(head, "line-head") || !strings.Contains(head, "truncated") {
		t.Fatalf("head trim wrong: %q", head)
	}
	if strings.Contains(head, "line-tail") {
		t.Fatal("head trim must drop the tail")
	}

	// keepTail=true (session.md): sobrevive la cola, marcador al principio.
	tail := budgetTrim(long, 200, true, "session.md")
	if !strings.HasSuffix(strings.TrimRight(tail, "\n"), "line-tail") || !strings.Contains(tail, "truncated") {
		t.Fatalf("tail trim wrong: %q", tail)
	}
	if strings.Contains(tail, "line-head") {
		t.Fatal("tail trim must drop the head")
	}
}

func TestSessionContextRespectsBudget(t *testing.T) {
	proj := t.TempDir()
	sf := filepath.Join(proj, "specforge")
	// learnings.md gigante: la inyección debe venir recortada con marcador.
	mustWrite(t, filepath.Join(sf, "learnings.md"), strings.Repeat("- lesson\n", injectBudgetBytes))

	out := sessionContext(proj)
	if len(out) > injectBudgetBytes*2 {
		t.Fatalf("injected context way over budget: %d bytes", len(out))
	}
	if !strings.Contains(out, "truncated") {
		t.Fatal("oversized learnings.md must carry the truncation marker")
	}
}
