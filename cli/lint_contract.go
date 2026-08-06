package main

import (
	"os"
	"regexp"
	"strings"
)

// ----------------------------------------------------------------------------
// Chequeo 6: CONTRATO skills ↔ CLI (C5 de EVALUACION-PLATAFORMA §4).
//
// El cierre de borde de FIXBUGHIGH ("prompt y enforcement tienen que decir lo
// mismo") se hizo a mano una vez — pero nada impedía que la próxima edición de
// una skill volviera a inventar un comando o un flag que no existe. La deriva
// prompt/CLI es recurrente POR CONSTRUCCIÓN: las skills son texto libre.
//
// Este chequeo extrae cada invocación `sf …` de los markdown (solo de CONTEXTO
// DE CÓDIGO: bloques cercados y spans `backtick` — la prosa se ignora para no
// cazar falsos positivos tipo "the sf gate command") y la valida contra
// cliSurface, la tabla de superficie real de comandos/subcomandos/flags.
//
// cliSurface se mantiene A MANO junto a main.go — y para que ESA dupla tampoco
// derive, lint_contract_test.go verifica que sus claves coincidan 1:1 con los
// comandos listados en usageText.
// ----------------------------------------------------------------------------

// cmdSurface describe la superficie de UN comando: sus subcomandos válidos
// (nil = no tiene) y sus flags válidos (agregados sobre todos los subcomandos:
// granularidad suficiente para cazar drift sin duplicar el parseo real).
type cmdSurface struct {
	subs  map[string]bool
	flags map[string]bool
}

// set es un helper de literales: set("a","b") → map[string]bool{"a":…,"b":…}.
func set(xs ...string) map[string]bool {
	m := make(map[string]bool, len(xs))
	for _, x := range xs {
		m[x] = true
	}
	return m
}

// cliSurface: la superficie completa de `sf`. Mantener en sync con main.go
// (el test de paridad contra usageText lo exige).
var cliSurface = map[string]cmdSurface{
	"lint":         {},
	"status":       {flags: set("--artifacts", "--feature", "--json")},
	"doctor":       {flags: set("--drift", "--install", "--global", "--quiet", "--run-tests", "--json")},
	"gate":         {subs: set("status", "approve", "record-verdict", "show"), flags: set("--feature", "--phase", "--by", "--comment", "--json")},
	"check":        {subs: set("run"), flags: set("--feature")},
	"feature":      {subs: set("add", "set-status", "set-lane", "archive"), flags: set("--feature", "--lane", "--depends-on", "--to", "--reason")},
	"trace":        {subs: set("verify"), flags: set("--feature", "--contract", "--wave")},
	"graph":        {subs: set("export", "query"), flags: set("--feature", "--format", "--stdout", "--json")},
	"tasks":        {subs: set("render", "validate"), flags: set("--feature", "--stdout")},
	"context":      {subs: set("for-wave", "current", "for-judge"), flags: set("--n", "--feature", "--phase", "--breadcrumb")},
	"metrics":      {subs: set("context"), flags: set("--feature", "--json")},
	"state":        {subs: set("current")},
	"next":         {flags: set("--json", "--explain")},
	"recover":      {flags: set("--feature", "--json")},
	"verify":       {flags: set("--feature", "--json", "--init-ci")},
	"migrate":      {flags: set("--dry-run")},
	"hook":         {flags: set("--event", "--harness")},
	"install":      {flags: set("--from", "--global", "--dry-run")},
	"uninstall":    {flags: set("--global")},
	"requirements": {subs: set("render", "validate"), flags: set("--feature", "--stdout")},
	"design":       {subs: set("render", "validate"), flags: set("--feature", "--stdout")},
	"constitution": {subs: set("render", "validate"), flags: set("--stdout")},
	"domain":       {subs: set("render", "validate", "terms"), flags: set("--stdout", "--json")},
	"sources":      {subs: set("render", "validate", "coverage"), flags: set("--stdout", "--json")},
	"plan":         {subs: set("render", "validate", "compute"), flags: set("--feature", "--stdout")},
	"review":       {subs: set("render", "validate"), flags: set("--feature", "--stdout")},
	"save":         {subs: set("constitution", "domain", "sources", "requirements", "design", "tasks", "plan", "review", "trace"), flags: set("--feature", "--json", "--from")},
	"journal":      {subs: set("add"), flags: set("--feature", "--json", "--bridge-icm", "--date")},
	"events":       {flags: set("--json", "--tail")},
	"onboard":      {subs: set("scan"), flags: set("--json")},
	"init":         {flags: set("--minimal")},
	"delta":        {subs: set("new", "list", "set-status"), flags: set("--feature", "--json", "--from", "--id", "--to", "--kind")},
	"run":          {flags: set("--feature", "--dry-run")},
	"coverage":     {flags: set("--json", "--badge", "--history", "--by-priority")},
	"arch":         {subs: set("rules", "check"), flags: set("--feature")},
	"mutation":     {subs: set("scope", "run"), flags: set("--feature")},
	"evidence":     {subs: set("checklist", "record"), flags: set("--feature", "--json", "--scenario", "--kind", "--ref", "--recorded")},
	"question":     {subs: set("add", "answer", "adopt", "list"), flags: set("--text", "--impact", "--blocks", "--asked-of", "--id", "--answer", "--source", "--json", "--blocking")},
	"help":         {},
}

// checkCLIContract recorre los markdown y valida cada invocación encontrada.
func checkCLIContract(files []string, root string, rep *report) {
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			continue // ya reportado por otros chequeos
		}
		inFence := false
		for i, line := range strings.Split(string(data), "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "```") {
				inFence = !inFence
				continue
			}
			// Contexto de código: la línea entera (dentro de un fence) o cada
			// span `…` (fuera). La prosa no se valida.
			var spans []string
			if inFence {
				spans = []string{line}
			} else {
				spans = backtickSpans(line)
			}
			for _, span := range spans {
				for _, prob := range invocationProblems(span) {
					rep.errorf("%s:%d: %s", rel(f, root), i+1, prob)
				}
			}
		}
	}
}

// backtickRe captura el contenido de spans `…` (sin los backticks).
var backtickRe = regexp.MustCompile("`([^`]+)`")

func backtickSpans(line string) []string {
	var out []string
	for _, m := range backtickRe.FindAllStringSubmatch(line, -1) {
		out = append(out, m[1])
	}
	return out
}

// segmentSplitRe corta una línea de shell en segmentos por operadores de
// composición: cada segmento puede empezar su propia invocación `sf …`.
var segmentSplitRe = regexp.MustCompile(`&&|\|\||;`)

// sfCmdRe reconoce el ARRANQUE de una invocación dentro de un segmento:
// `sf` como palabra (inicio, espacio o `$(`) seguida del comando en kebab.
var sfCmdRe = regexp.MustCompile(`(?:^|\s|\$\()sf\s+([a-z][a-z-]*)\b(.*)$`)

// invocationProblems valida las invocaciones `sf …` de un span de código y
// devuelve los problemas (vacío = todo legal o no hay invocación).
func invocationProblems(span string) []string {
	var probs []string
	for _, seg := range segmentSplitRe.Split(span, -1) {
		m := sfCmdRe.FindStringSubmatch(seg)
		if m == nil {
			continue
		}
		cmd, rest := m[1], m[2]
		surf, ok := cliSurface[cmd]
		if !ok {
			probs = append(probs, "unknown `sf` command `"+cmd+"` (skills↔CLI contract)")
			continue
		}
		probs = append(probs, tokenProblems(cmd, surf, strings.Fields(rest))...)
	}
	return probs
}

// tokenProblems valida subcomando + flags de una invocación ya identificada.
func tokenProblems(cmd string, surf cmdSurface, tokens []string) []string {
	var probs []string
	subChecked := false
	for _, tok := range tokens {
		// Fin de la parte "comando": redirecciones, comentarios, pipes sueltos.
		if tok == "#" || tok == ">" || tok == ">>" || tok == "|" {
			break
		}
		if strings.HasPrefix(tok, "--") {
			// Flag: validamos el NOMBRE (lo previo al `=`; el valor es libre).
			name := tok
			if i := strings.IndexByte(name, '='); i >= 0 {
				name = name[:i]
			}
			if !surf.flags[name] {
				probs = append(probs, "`sf "+cmd+"` has no flag `"+name+"` (skills↔CLI contract)")
			}
			continue
		}
		// Primer token posicional: candidato a subcomando (si el comando tiene).
		// Placeholders (<…>, MAYÚSCULAS, $VAR, […]) no se validan — son docs.
		if !subChecked && surf.subs != nil {
			subChecked = true
			if isPlaceholderToken(tok) {
				continue
			}
			if !surf.subs[tok] {
				probs = append(probs, "`sf "+cmd+"` has no sub-command `"+tok+"` (skills↔CLI contract)")
			}
		}
		// Posicionales posteriores (project_dir, valores): libres.
	}
	return probs
}

// isPlaceholderToken: tokens de documentación, no literales a validar.
func isPlaceholderToken(tok string) bool {
	if strings.ContainsAny(tok, "<>[]{}$") {
		return true
	}
	return strings.ToLower(tok) != tok // tiene mayúsculas → NAME, PHASE, etc.
}
