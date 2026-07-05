package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// report acumula los problemas encontrados durante el lint. Los errores fallan la ejecución;
// las advertencias son informativas y nunca cambian el código de salida.
type report struct {
	errors   []string
	warnings []string
}

// errorf y warnf tienen un receptor de *puntero* (*report) porque mutan el
// struct (append en sus slices). Con un receptor por valor mutarían una copia y
// el cambio se perdería. `...any` los hace variádicos, como fmt.Printf.
func (r *report) errorf(format string, a ...any) {
	r.errors = append(r.errors, fmt.Sprintf(format, a...))
}

func (r *report) warnf(format string, a ...any) {
	r.warnings = append(r.warnings, fmt.Sprintf(format, a...))
}

// productDocs son los documentos de primer nivel que rige el lint, relativos a la raíz
// del repo. Los archivos scratch/fuente (CLI-SPEC, SPECFORGE-REVIEW, ...) se excluyen a
// propósito — mencionan legítimamente los nombres antiguos sdd-.
var productDocs = []string{
	"AGENT.md",
	"README.md",
	"README.es.md",
	"SUPPORT-SKILLS.md",
	"SUPPORT-SKILLS.es.md",
	"INSTALL.md",
}

// runLint es el punto de entrada de `sf lint`. Encuentra la raíz del repo, recopila los
// archivos markdown, ejecuta las comprobaciones, imprime el informe y devuelve 0 (limpio) o
// 1 (al menos un error).
func runLint(args []string) int {
	root := findRoot()
	if root == "" {
		fmt.Fprintln(os.Stderr, "sf lint: not inside a SpecForge repo (no skills/ dir found)")
		return 1
	}

	files, err := collectMarkdown(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf lint: %v\n", err)
		return 1
	}

	skills, err := discoverSkills(root)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf lint: %v\n", err)
		return 1
	}

	var rep report
	checkSkills(root, skills, &rep)
	checkNoSDD(files, root, &rep)
	checkSkillRefs(files, skills, root, &rep)
	checkRelativeLinks(files, root, &rep)
	checkParity(root, &rep)
	checkCLIContract(files, root, &rep)

	fmt.Printf("Linted %d skills, %d markdown files.\n", len(skills), len(files))
	for _, w := range rep.warnings {
		fmt.Printf("  warning: %s\n", w)
	}
	for _, e := range rep.errors {
		fmt.Printf("  ERROR:   %s\n", e)
	}

	if len(rep.errors) > 0 {
		fmt.Printf("\nFAIL: %d error(s), %d warning(s).\n", len(rep.errors), len(rep.warnings))
		return 1
	}
	fmt.Printf("\nOK: 0 errors, %d warning(s).\n", len(rep.warnings))
	return 0
}

// findRoot sube desde el directorio actual hasta encontrar un directorio que
// contenga el subdirectorio skills/ — esa es la raíz del repo. Devuelve "" si no encuentra ninguno.
func findRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}
	for {
		if info, err := os.Stat(filepath.Join(dir, "skills")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir { // filepath.Dir("/") == "/": llegamos a la raíz del FS
			return ""
		}
		dir = parent
	}
}

// collectMarkdown devuelve todos los archivos .md bajo skills/ más los documentos
// de producto que existan. filepath.WalkDir recorre el árbol; el callback decide qué
// conservar. Devolver un error desde el callback aborta el recorrido.
func collectMarkdown(root string) ([]string, error) {
	var files []string

	skillsDir := filepath.Join(root, "skills")
	err := filepath.WalkDir(skillsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".md") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, name := range productDocs {
		p := filepath.Join(root, name)
		if _, err := os.Stat(p); err == nil {
			files = append(files, p)
		}
	}
	return files, nil
}

// checkNoSDD señala cualquier referencia residual a `sdd-` (la suite se renombró de
// sdd- a sf-). Escaneamos línea por línea para poder reportar el número de línea.
func checkNoSDD(files []string, root string, rep *report) {
	for _, f := range files {
		fh, err := os.Open(f)
		if err != nil {
			rep.errorf("%s: cannot read (%v)", rel(f, root), err)
			continue
		}
		scanner := bufio.NewScanner(fh)
		for line := 1; scanner.Scan(); line++ {
			if strings.Contains(scanner.Text(), "sdd-") {
				rep.errorf("%s:%d: leftover `sdd-` reference", rel(f, root), line)
			}
		}
		fh.Close()
	}
}

// rel acorta una ruta absoluta a una relativa a root, para una salida ordenada.
func rel(path, root string) string {
	if r, err := filepath.Rel(root, path); err == nil {
		return r
	}
	return path
}

// nameRe es la forma permitida del nombre del skill: sf-, sfx- o sfp- seguido de un
// token kebab en minúsculas. MustCompile hace panic con un patrón inválido al arrancar, que
// es lo deseable para una regex constante (un bug, no una condición de runtime).
var nameRe = regexp.MustCompile(`^sf[xp]?-[a-z][a-z0-9-]*$`)

// descMax es el presupuesto de longitud de la descripción en SKILL.md.
const descMax = 1024

// discoverSkills devuelve los nombres de los directorios de skills bajo skills/.
func discoverSkills(root string) ([]string, error) {
	entries, err := os.ReadDir(filepath.Join(root, "skills"))
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// checkSkills valida el frontmatter de SKILL.md de cada skill: debe existir, su
// `name` debe coincidir con el nombre de la carpeta y cumplir nameRe, y `description` debe
// estar presente y dentro del presupuesto.
func checkSkills(root string, skills []string, rep *report) {
	for _, name := range skills {
		skillMD := filepath.Join(root, "skills", name, "SKILL.md")
		data, err := os.ReadFile(skillMD)
		if err != nil {
			rep.errorf("%s: missing SKILL.md", name)
			continue
		}
		fm, ok := parseFrontmatter(string(data))
		if !ok {
			rep.errorf("%s/SKILL.md: missing or malformed frontmatter", name)
			continue
		}

		switch fmName := fm["name"]; {
		case fmName == "":
			rep.errorf("%s/SKILL.md: frontmatter has no `name`", name)
		case fmName != name:
			rep.errorf("%s/SKILL.md: name %q != folder %q", name, fmName, name)
		case !nameRe.MatchString(fmName):
			rep.errorf("%s/SKILL.md: name %q fails sf-/sfx- regex", name, fmName)
		}

		switch desc := fm["description"]; {
		case desc == "":
			rep.errorf("%s/SKILL.md: frontmatter has no `description`", name)
		case len(desc) > descMax:
			rep.errorf("%s/SKILL.md: description %d chars > %d", name, len(desc), descMax)
		}
	}
}

// keyRe coincide con una línea de frontmatter `key: value`. \w incluye [A-Za-z0-9_].
var keyRe = regexp.MustCompile(`^([a-zA-Z_][\w-]*):\s*(.*)$`)

// parseFrontmatter extrae el bloque inicial --- ... --- en un mapa plano.
// Hecho a mano (sin dependencia de YAML, igual que el linter en Python) para manejar las
// líneas simples `key: value` y el escalar de bloque plegado `key: >` que usan las
// descripciones de SKILL.md. Devuelve ok=false cuando no hay frontmatter.
func parseFrontmatter(text string) (map[string]string, bool) {
	if !strings.HasPrefix(text, "---") {
		return nil, false
	}
	end := strings.Index(text[3:], "\n---")
	if end == -1 {
		return nil, false
	}
	block := strings.Trim(text[3:3+end], "\n")

	fields := map[string]string{}
	var key string
	var buf []string

	// flush guarda el valor acumulado de la clave actual. Es un closure para poder
	// leer/resetear las vars key y buf declaradas arriba.
	flush := func() {
		if key != "" {
			fields[key] = strings.TrimSpace(strings.Join(buf, " "))
		}
	}

	for _, line := range strings.Split(block, "\n") {
		m := keyRe.FindStringSubmatch(line)
		if m != nil && !strings.HasPrefix(line, " ") {
			// Nueva clave de primer nivel: terminar la anterior y empezar de nuevo.
			flush()
			key = m[1]
			buf = nil
			if rest := strings.TrimSpace(m[2]); rest != "" && rest != ">" && rest != "|" {
				buf = append(buf, rest)
			}
		} else if key != "" {
			// Línea de continuación de un escalar plegado.
			buf = append(buf, strings.TrimSpace(line))
		}
	}
	flush()
	return fields, true
}

// ----------------------------------------------------------------------------
// Helpers compartidos por los chequeos "doc-wide" (operan sobre el cuerpo del
// markdown, no skill por skill).
// ----------------------------------------------------------------------------

// stripCodeBlocks blanquea las líneas dentro de bloques de código cercados con
// ```. Así los ejemplos de código (que mencionan skills, links rotos a
// propósito, etc.) no disparan falsos positivos. Mantenemos la MISMA cantidad
// de líneas (reemplazamos por "") para que los números de línea sigan siendo
// correctos al reportar.
func stripCodeBlocks(text string) string {
	var out []string
	inFence := false
	for _, line := range strings.Split(text, "\n") {
		// Una cerca abre o cierra el bloque. La propia línea de cerca también se
		// blanquea.
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inFence = !inFence
			out = append(out, "")
			continue
		}
		if inFence {
			out = append(out, "")
		} else {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

// ----------------------------------------------------------------------------
// Chequeo 3: referencias a skills inexistentes (warning, no error).
// ----------------------------------------------------------------------------

// tokenRe captura tokens tipo `sf-foo`, `sfx-bar`, `sfp-baz` en cualquier parte
// de una línea. \b son límites de palabra (word boundaries): evitan matchear
// `xsf-foo` o `sf-foo-` parcialmente.
var tokenRe = regexp.MustCompile(`\bsf[xp]?-[a-z][a-z0-9-]*\b`)

// nonSkillTokens es la allowlist: tokens con forma de skill que NO son skills
// (comandos del CLI, skills planeados/futuros). Referenciarlos no cuenta como
// referencia rota.
//
// El tipo es map[string]struct{}: el idioma de Go para un "set". struct{} es un
// tipo vacío que ocupa 0 bytes, así que el set guarda solo las claves. Se
// consulta con `_, ok := set[k]`.
var nonSkillTokens = map[string]struct{}{
	"sf-status":       {}, // comando
	"sf-doctor":       {}, // comando CLI planeado
	"sf-amend":        {}, // planeado (F33)
	"sf-discover":     {}, // planeado (F27)
	"sf-prd":          {}, // planeado (F27)
	"sfp-po":          {}, // posible skill futuro: PRD formal opcional
	"sfp-prd":         {}, // posible skill futuro: PRD formal opcional
	"sf-lint":         {}, // comando CLI planeado
	"sf-gate":         {}, // comando CLI planeado
	"sf-traceability": {},
	"sf-save":         {},
	"sf-op":           {},
	"sf-sync":         {},
	"sf-init":         {}, // también es skill; listado por las dudas
}

// checkSkillRefs avisa (warning) cuando un doc referencia un token con forma de
// skill que no existe ni en los skills reales ni en la allowlist.
func checkSkillRefs(files []string, skills []string, root string, rep *report) {
	// known = skills reales ∪ community ∪ allowlist. Las de skills-community/
	// (D7') no participan del pipeline ni las valida el resto del lint, pero
	// SÍ son nombres legítimos para referenciar desde los docs.
	known := make(map[string]struct{}, len(skills)+len(nonSkillTokens))
	for _, s := range skills {
		known[s] = struct{}{}
	}
	if entries, err := os.ReadDir(filepath.Join(root, "skills-community")); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				known[e.Name()] = struct{}{}
			}
		}
	}
	for t := range nonSkillTokens {
		known[t] = struct{}{}
	}

	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			rep.errorf("%s: cannot read (%v)", rel(f, root), err)
			continue
		}
		body := stripCodeBlocks(string(data))
		for i, line := range strings.Split(body, "\n") {
			// FindAllString devuelve todos los matches de la línea (-1 = sin
			// límite). range sobre []string nos da índice y valor; acá solo el
			// valor (el token).
			for _, tok := range tokenRe.FindAllString(line, -1) {
				if _, ok := known[tok]; !ok {
					rep.warnf("%s:%d: reference to unknown skill `%s`", rel(f, root), i+1, tok)
				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Chequeo 4: links relativos rotos (error).
// ----------------------------------------------------------------------------

// linkRe captura links markdown `[texto](destino)` y extrae el destino en el
// grupo 1 — los paréntesis en `([^)]+)` definen ese grupo de captura.
var linkRe = regexp.MustCompile(`\[[^\]]+\]\(([^)]+)\)`)

// checkRelativeLinks verifica que cada link relativo apunte a un archivo que
// existe. Los links externos (http, ancla #, mailto) se ignoran.
func checkRelativeLinks(files []string, root string, rep *report) {
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			rep.errorf("%s: cannot read (%v)", rel(f, root), err)
			continue
		}
		body := stripCodeBlocks(string(data))
		for i, line := range strings.Split(body, "\n") {
			// FindAllStringSubmatch devuelve, por cada match, un slice donde el
			// índice 0 es el match completo y el 1 el primer grupo (el destino).
			for _, m := range linkRe.FindAllStringSubmatch(line, -1) {
				target := strings.TrimSpace(m[1])
				if target == "" ||
					strings.HasPrefix(target, "http://") ||
					strings.HasPrefix(target, "https://") ||
					strings.HasPrefix(target, "#") ||
					strings.HasPrefix(target, "mailto:") {
					continue
				}
				// Sacar el ancla (`archivo.md#seccion` → `archivo.md`).
				pathPart := target
				if idx := strings.Index(pathPart, "#"); idx >= 0 {
					pathPart = pathPart[:idx]
				}
				if pathPart == "" {
					continue
				}
				// Los links relativos se resuelven contra el directorio del
				// archivo que los contiene, no contra la raíz del repo.
				abs := filepath.Join(filepath.Dir(f), pathPart)
				if _, err := os.Stat(abs); err != nil {
					rep.errorf("%s:%d: broken relative link `%s`", rel(f, root), i+1, target)
				}
			}
		}
	}
}

// ----------------------------------------------------------------------------
// Chequeo 5: paridad de secciones entre los docs ES y EN (error).
// ----------------------------------------------------------------------------

// headers devuelve los encabezados de nivel 2 (`## ...`) de un archivo, ya
// recortados. Devuelve error si el archivo no se puede leer.
func headers(path string) ([]string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var hs []string
	for _, line := range strings.Split(stripCodeBlocks(string(data)), "\n") {
		if strings.HasPrefix(line, "## ") {
			hs = append(hs, strings.TrimSpace(line))
		}
	}
	return hs, nil
}

// countPrefix cuenta cuántos strings del slice empiezan con prefix.
func countPrefix(hs []string, prefix string) int {
	n := 0
	for _, h := range hs {
		if strings.HasPrefix(h, prefix) {
			n++
		}
	}
	return n
}

// checkParity compara los pares de docs bilingües: deben tener la misma
// cantidad de secciones `##` y la misma cantidad de skills `## sfx-`
// documentados. Es la red de seguridad contra drift entre EN y ES.
func checkParity(root string, rep *report) {
	// [2]string es un array de tamaño fijo 2 (no un slice). Útil para pares.
	pairs := [][2]string{
		{"README.md", "README.es.md"},
		{"SUPPORT-SKILLS.md", "SUPPORT-SKILLS.es.md"},
	}
	for _, p := range pairs {
		en, es := p[0], p[1]
		enH, errEn := headers(filepath.Join(root, en))
		esH, errEs := headers(filepath.Join(root, es))
		if errEn != nil || errEs != nil {
			rep.errorf("parity: missing one of %s / %s", en, es)
			continue
		}
		if len(enH) != len(esH) {
			rep.errorf("parity: %s has %d `##` sections, %s has %d", en, len(enH), es, len(esH))
		}
		if a, b := countPrefix(enH, "## sfx-"), countPrefix(esH, "## sfx-"); a != b {
			rep.errorf("parity: %s documents %d sfx- skills, %s %d", en, a, es, b)
		}
	}
}
