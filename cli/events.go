package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf events` — telemetría del enforcement (A7/D10).
//
// El problema que resuelve: hasta acá los denies del hook, los nudges y los
// veredictos fallidos NO quedaban registrados en ningún lado. Sin datos no se
// puede responder "¿cuántas veces intentó el modelo escribir estado directo
// esta semana?", ni tunear el sistema, ni demostrar su valor.
//
// Diseño:
//   - specforge/.state/events.jsonl — UN evento por línea (JSONL). Append-only
//     y best-effort: un fallo del log JAMÁS rompe el comando que lo emite (la
//     telemetría es observación, no control).
//   - Emisores: el hook (deny/nudge), record-verdict (pass/fail) y el gate
//     approve cuando REHÚSA un verdict (refuse).
//   - Poda incorporada (C6-style): si el archivo supera el cap, conservamos la
//     mitad más reciente. Un proyecto de meses no acumula megabytes de log.
// ----------------------------------------------------------------------------

// sfEvent es un registro de telemetría. Kind:
//
//	deny    el hook denegó una escritura de estado (Write/Edit o Bash)
//	nudge   el hook empujó una acción pendiente (p.ej. journal al archivar)
//	verdict record-verdict persistió un veredicto del juez (detail: pass|fail)
//	refuse  el CLI rehusó una operación (p.ej. verdict con precondiciones rotas)
type sfEvent struct {
	At      string `json:"at"`
	Kind    string `json:"kind"`
	Feature string `json:"feature,omitempty"`
	Phase   string `json:"phase,omitempty"`
	Session string `json:"session,omitempty"`
	Target  string `json:"target,omitempty"` // path o comando que gatilló el evento
	Detail  string `json:"detail,omitempty"` // razón corta / overall del veredicto
}

// eventsMaxBytes: cap del log antes de podar. 512K ≈ miles de eventos; sobra
// para tunear y no molesta en disco.
const eventsMaxBytes = 512 * 1024

// eventDetailMax: los reasons del hook son párrafos didácticos; para telemetría
// alcanza el arranque (el mensaje completo ya lo vio el agente).
const eventDetailMax = 160

func eventsPath(projectDir string) string {
	return filepath.Join(projectDir, "specforge", ".state", "events.jsonl")
}

// logEvent apendea un evento. Best-effort en TODOS los pasos: sin error de
// retorno a propósito — ningún emisor debe fallar (ni ensuciar stdout, que en
// el hook es el canal del protocolo) porque la telemetría no pudo escribirse.
func logEvent(projectDir string, e sfEvent) {
	// Solo dentro de un proyecto SpecForge real (mismo guard que el hook).
	if !isDir(filepath.Join(projectDir, "specforge")) {
		return
	}
	e.At = nowUTC()
	e.Detail = truncateDetail(e.Detail)
	line, err := json.Marshal(e)
	if err != nil {
		return
	}
	path := eventsPath(projectDir)
	if os.MkdirAll(filepath.Dir(path), 0o755) != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(line, '\n'))
	_ = f.Close()
	pruneEvents(path)
}

// pruneEvents mantiene el log acotado: pasado el cap, conserva la mitad MÁS
// RECIENTE de las líneas. (Podar de a mitades amortiza: no reescribimos el
// archivo en cada append.)
func pruneEvents(path string) {
	info, err := os.Stat(path)
	if err != nil || info.Size() <= eventsMaxBytes {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	keep := lines[len(lines)/2:]
	_ = os.WriteFile(path, []byte(strings.Join(keep, "\n")+"\n"), 0o644)
}

func truncateDetail(s string) string {
	if len(s) <= eventDetailMax {
		return s
	}
	return s[:eventDetailMax] + "…"
}

// readEvents devuelve todos los eventos del log (líneas ilegibles se saltean:
// el log es telemetría, no estado — mejor perder una línea que abortar).
func readEvents(projectDir string) []sfEvent {
	data, err := os.ReadFile(eventsPath(projectDir))
	if err != nil {
		return nil
	}
	var out []sfEvent
	for line := range strings.SplitSeq(strings.TrimRight(string(data), "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var e sfEvent
		if json.Unmarshal([]byte(line), &e) == nil {
			out = append(out, e)
		}
	}
	return out
}

// denyEvent arma el evento de un deny del hook: el target es la ruta relativa
// al proyecto (Write/Edit) o el comando recortado (Bash) — lo que haga falta
// para responder después "¿QUÉ intentó escribir el modelo?".
func denyEvent(projectDir, filePath, command, session, reason string) sfEvent {
	target := ""
	switch {
	case strings.TrimSpace(command) != "":
		target = truncateDetail(strings.TrimSpace(command))
	default:
		if rel, inside := relPosix(projectDir, filePath); inside {
			target = rel
		} else {
			target = filePath
		}
	}
	return sfEvent{Kind: "deny", Session: session, Target: target, Detail: reason}
}

// ── El comando `sf events` ───────────────────────────────────────────────────

// runEvents: `sf events [--json] [--tail=N] [project_dir]`.
// Default: resumen agregado (conteos por kind, blancos más denegados, veredictos
// fallidos por feature) + los últimos N eventos. --json: todo el log como array.
func runEvents(args []string) int {
	projectDir := "."
	asJSON := false
	tail := 15
	for _, a := range args {
		switch {
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "--tail="):
			n, err := strconv.Atoi(strings.TrimPrefix(a, "--tail="))
			if err != nil || n < 0 {
				fmt.Fprintln(os.Stderr, "sf events: --tail wants a non-negative number")
				return 2
			}
			tail = n
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf events: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	events := readEvents(projectDir)
	if asJSON {
		out, _ := json.MarshalIndent(events, "", "  ")
		fmt.Println(string(out))
		return 0
	}
	if len(events) == 0 {
		fmt.Println("No events recorded yet (specforge/.state/events.jsonl is empty).")
		fmt.Println("Denies, nudges and judge verdicts will appear here as the hooks run.")
		return 0
	}

	printEventsSummary(events)
	printEventsTail(events, tail)
	return 0
}

func printEventsSummary(events []sfEvent) {
	kinds := map[string]int{}
	denyTargets := map[string]int{}
	verdictFails := map[string]int{}
	for _, e := range events {
		kinds[e.Kind]++
		if e.Kind == "deny" && e.Target != "" {
			denyTargets[e.Target]++
		}
		if e.Kind == "verdict" && e.Detail == "fail" {
			verdictFails[e.Feature]++
		}
	}

	fmt.Printf("SpecForge — enforcement telemetry (%d event(s))\n\n", len(events))
	for _, k := range []string{"deny", "nudge", "verdict", "refuse"} {
		if kinds[k] > 0 {
			fmt.Printf("  %-8s %d\n", k, kinds[k])
		}
	}
	// Kinds futuros que este binario no conoce todavía: mostrarlos igual.
	for k, n := range kinds {
		switch k {
		case "deny", "nudge", "verdict", "refuse":
		default:
			fmt.Printf("  %-8s %d\n", k, n)
		}
	}

	if len(denyTargets) > 0 {
		fmt.Println("\n  most-denied targets:")
		for _, kv := range topN(denyTargets, 5) {
			fmt.Printf("    %3d× %s\n", kv.n, kv.k)
		}
	}
	if len(verdictFails) > 0 {
		fmt.Println("\n  failed judge verdicts per feature (REVISE pressure):")
		for _, kv := range topN(verdictFails, 5) {
			fmt.Printf("    %3d× %s\n", kv.n, kv.k)
		}
	}
	fmt.Println()
}

func printEventsTail(events []sfEvent, tail int) {
	if tail == 0 {
		return
	}
	if tail < len(events) {
		events = events[len(events)-tail:]
	}
	fmt.Printf("last %d event(s):\n", len(events))
	for _, e := range events {
		target := e.Target
		if target == "" {
			target = e.Feature
			if e.Phase != "" {
				target += "/" + e.Phase
			}
		}
		fmt.Printf("  %s  %-8s %s — %s\n", e.At, e.Kind, target, e.Detail)
	}
}

// topN ordena un mapa conteo→clave de mayor a menor y devuelve los primeros n.
type kcount struct {
	k string
	n int
}

func topN(m map[string]int, n int) []kcount {
	out := make([]kcount, 0, len(m))
	for k, v := range m {
		out = append(out, kcount{k, v})
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].n != out[j].n {
			return out[i].n > out[j].n
		}
		return out[i].k < out[j].k // desempate estable
	})
	if len(out) > n {
		out = out[:n]
	}
	return out
}
