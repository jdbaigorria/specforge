package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf graph export` — el grafo de conocimiento DETERMINÍSTICO.
//
// Idea (tomada de Kaddo, pero sobre NUESTRA estructura): no inventamos relaciones
// ni llamamos a un LLM ni usamos embeddings. Solo hacemos EXPLÍCITO el grafo que
// YA existe implícito en los archivos que el proyecto declara:
//
//     features.json   → nodos de feature + aristas feature→feature (depends_on)
//     <feat>/trace.json → nodos R# / code / test + aristas de la espina:
//                          feature → R (has_requirement)
//                          R → code:symbol (implemented_by)
//                          R → test (verified_by)
//
// Es lo OPUESTO a RAG: RAG recupera por similitud; esto exporta la estructura
// declarada, sin modelo, reproducible. Salida: graph.json (para tools/tests) y
// graph.mmd (Mermaid, se renderiza solo en GitHub/Markdown).
// ----------------------------------------------------------------------------

// knowledgeGraph es el documento de salida. SchemaVersion + Scope lo hacen
// autodescriptivo (igual que los otros artefactos JSON-first).
type knowledgeGraph struct {
	SchemaVersion string      `json:"schema_version"`
	Scope         string      `json:"scope"` // "all" o el nombre de la feature filtrada
	Nodes         []graphNode `json:"nodes"`
	Edges         []graphEdge `json:"edges"`
}

// graphNode: Type es el "color" del nodo (feature|requirement|code|test); Label
// es lo que se muestra. ID es único y seguro para Mermaid (lo sanitizamos).
// Group es el id del nodo-feature "dueño" (para agrupar en subgraphs al renderizar);
// queda vacío en los propios nodos de feature. omitempty: no ensucia el JSON de
// los features. Un nodo compartido (code/test referido por varias features) se
// asigna a la PRIMERA que lo introduce — un nodo solo puede vivir en una caja.
type graphNode struct {
	ID    string `json:"id"`
	Type  string `json:"type"`
	Label string `json:"label"`
	Group string `json:"group,omitempty"`
}

type graphEdge struct {
	From string `json:"from"`
	To   string `json:"to"`
	Type string `json:"type"` // depends_on | has_requirement | implemented_by | verified_by
}

// runGraph despacha `sf graph <export|query>`.
func runGraph(args []string) int {
	if len(args) == 0 {
		graphUsage()
		return 2
	}
	switch args[0] {
	case "export":
		return graphExportCmd(args[1:])
	case "query":
		return graphQueryCmd(args[1:])
	default:
		graphUsage()
		return 2
	}
}

func graphUsage() {
	fmt.Fprintln(os.Stderr, "usage: sf graph export [project_dir] [--feature=NAME] [--format=json|mermaid|both] [--stdout]")
	fmt.Fprintln(os.Stderr, "       sf graph query <term> [project_dir] [--json]")
}

// ----------------------------------------------------------------------------
// `sf graph query` — el grafo CONSULTABLE (D5').
//
// El export es un snapshot; la pregunta real del día a día es puntual: "¿qué
// features tocan auth?", "¿qué tests cubren R3?". RAG respondería por
// similitud; acá respondemos por ESTRUCTURA: matcheamos el término contra los
// nodos declarados y devolvemos sus vecinos (aristas entrantes y salientes).
// Determinista, sin modelo, sin índice — el grafo se reconstruye del disco en
// cada consulta (es chico y siempre fresco).
// ----------------------------------------------------------------------------

func graphQueryCmd(args []string) int {
	projectDir := "."
	term := ""
	asJSON := false
	for _, a := range args {
		switch {
		case a == "--json":
			asJSON = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf graph: unknown flag %q\n", a)
			return 2
		case term == "":
			term = a
		default:
			projectDir = a
		}
	}
	if term == "" {
		fmt.Fprintln(os.Stderr, "sf graph query: a search term is required (e.g. `sf graph query auth`)")
		return 2
	}

	g, code := buildGraph(projectDir, "")
	if code != 0 {
		return code
	}

	matches, neighbors := queryGraph(g, term)
	if asJSON {
		out, _ := json.MarshalIndent(map[string]any{"term": term, "matches": matches, "edges": neighbors}, "", "  ")
		fmt.Println(string(out))
		return 0
	}
	if len(matches) == 0 {
		fmt.Printf("no node matches %q (searched %d node(s): features, requirements, code symbols, tests)\n",
			term, len(g.Nodes))
		return 0
	}
	label := map[string]string{}
	for _, n := range g.Nodes {
		label[n.ID] = fmt.Sprintf("%s %q", n.Type, n.Label)
	}
	fmt.Printf("%d node(s) match %q:\n\n", len(matches), term)
	for _, m := range matches {
		fmt.Printf("● %s\n", label[m.ID])
		for _, e := range neighbors {
			switch {
			case e.From == m.ID:
				fmt.Printf("    —%s→ %s\n", e.Type, label[e.To])
			case e.To == m.ID:
				fmt.Printf("    ←%s— %s\n", e.Type, label[e.From])
			}
		}
	}
	return 0
}

// queryGraph matchea el término (case-insensitive) contra label e id de cada
// nodo, y devuelve también todas las aristas que tocan algún match.
func queryGraph(g knowledgeGraph, term string) ([]graphNode, []graphEdge) {
	needle := strings.ToLower(term)
	matched := map[string]bool{}
	var matches []graphNode
	for _, n := range g.Nodes {
		if strings.Contains(strings.ToLower(n.Label), needle) || strings.Contains(strings.ToLower(n.ID), needle) {
			matched[n.ID] = true
			matches = append(matches, n)
		}
	}
	var neighbors []graphEdge
	for _, e := range g.Edges {
		if matched[e.From] || matched[e.To] {
			neighbors = append(neighbors, e)
		}
	}
	return matches, neighbors
}

func graphExportCmd(args []string) int {
	projectDir := "."
	feature := ""
	format := "both" // json | mermaid | both
	toStdout := false
	for _, a := range args {
		switch {
		case strings.HasPrefix(a, "--feature="):
			feature = strings.TrimPrefix(a, "--feature=")
		case strings.HasPrefix(a, "--format="):
			format = strings.TrimPrefix(a, "--format=")
		case a == "--stdout":
			toStdout = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf graph: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}
	if format != "json" && format != "mermaid" && format != "both" {
		fmt.Fprintf(os.Stderr, "sf graph: --format must be json|mermaid|both, got %q\n", format)
		return 2
	}

	g, code := buildGraph(projectDir, feature)
	if code != 0 {
		return code
	}

	// --stdout imprime y no escribe nada (útil para preview/test). Si se pidió
	// "both" con --stdout, mostramos el JSON (el formato canónico).
	if toStdout {
		if format == "mermaid" {
			fmt.Print(renderGraphMermaid(g))
		} else {
			out, _ := json.MarshalIndent(g, "", "  ")
			fmt.Println(string(out))
		}
		return 0
	}

	specforge := filepath.Join(projectDir, "specforge")
	wrote := []string{}
	if format == "json" || format == "both" {
		out, err := json.MarshalIndent(g, "", "  ")
		if err != nil {
			fmt.Fprintf(os.Stderr, "sf graph: marshal failed (%v)\n", err)
			return 1
		}
		p := filepath.Join(specforge, "graph.json")
		if err := os.WriteFile(p, append(out, '\n'), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "sf graph: cannot write %s (%v)\n", p, err)
			return 1
		}
		wrote = append(wrote, filepath.Join("specforge", "graph.json"))
	}
	if format == "mermaid" || format == "both" {
		p := filepath.Join(specforge, "graph.mmd")
		if err := os.WriteFile(p, []byte(renderGraphMermaid(g)), 0o644); err != nil {
			fmt.Fprintf(os.Stderr, "sf graph: cannot write %s (%v)\n", p, err)
			return 1
		}
		wrote = append(wrote, filepath.Join("specforge", "graph.mmd"))
	}
	fmt.Printf("exported %d node(s), %d edge(s) → %s\n", len(g.Nodes), len(g.Edges), strings.Join(wrote, " + "))
	return 0
}

// buildGraph arma el grafo recorriendo features.json y los trace.json. Es casi
// "puro": lee del disco, pero no muta nada. code != 0 solo si falta/rompe
// features.json (sin proyecto no hay grafo). Las features sin trace.json quedan
// como nodos sueltos — es correcto, no un error (todavía no construidas).
func buildGraph(projectDir, featureFilter string) (knowledgeGraph, int) {
	specforge := filepath.Join(projectDir, "specforge")

	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf graph: no feature state under %s (%v)\n", projectDir, err)
		return knowledgeGraph{}, 4
	}

	scope := "all"
	if featureFilter != "" {
		scope = featureFilter
	}

	// addNode deduplica por ID: un mismo símbolo de código puede ser referido por
	// varias R (incluso de distintas features); lo queremos una sola vez. Cierra
	// sobre `seen` y `nodes` (closure) para no pasarlos en cada llamada.
	seen := map[string]bool{}
	var nodes []graphNode
	addNode := func(id, typ, label, group string) {
		if seen[id] {
			return // ya existe: conserva el group de la PRIMERA vez que se agregó
		}
		seen[id] = true
		nodes = append(nodes, graphNode{ID: id, Type: typ, Label: label, Group: group})
	}
	var edges []graphEdge

	// 1) Nodos de feature. Filtramos acá si vino --feature.
	inScope := map[string]bool{}
	for _, f := range ff.Features {
		if featureFilter != "" && f.Name != featureFilter {
			continue
		}
		inScope[f.Name] = true
		// El nodo de feature es el "header" de su caja → sin group.
		addNode(featNodeID(f.Name), "feature", "feature: "+f.Name, "")
	}

	// 2) Aristas feature→feature (depends_on). Solo si ambos extremos están en
	// scope, para no dibujar aristas que cuelgan fuera del grafo filtrado.
	for _, f := range ff.Features {
		if !inScope[f.Name] {
			continue
		}
		for _, dep := range f.DependsOn {
			if inScope[dep] {
				edges = append(edges, graphEdge{From: featNodeID(f.Name), To: featNodeID(dep), Type: "depends_on"})
			}
		}
	}

	// 3) Espina de trazabilidad por feature: leemos su trace.json (si existe) y
	// expandimos R → code / test. findArtifact ya busca en archive/ y features/.
	for _, f := range ff.Features {
		if !inScope[f.Name] {
			continue
		}
		tracePath := findArtifact(specforge, f.Name, "trace.json")
		if tracePath == "" {
			continue // feature sin trace todavía: queda como nodo suelto
		}
		tf, ok := readTraceQuiet(tracePath)
		if !ok {
			continue // trace ilegible: lo saltamos sin romper el resto del grafo
		}

		// Orden estable: los maps de Go iteran aleatorio y queremos salida
		// reproducible (importante para tests y diffs de git).
		reqIDs := make([]string, 0, len(tf.Requirements))
		for id := range tf.Requirements {
			reqIDs = append(reqIDs, id)
		}
		sort.Strings(reqIDs)

		owner := featNodeID(f.Name) // todos los nodos de esta feature se agrupan bajo su caja
		for _, rid := range reqIDs {
			req := tf.Requirements[rid]
			reqNode := reqNodeID(f.Name, rid) // R# es único solo DENTRO de la feature → lo prefijamos
			addNode(reqNode, "requirement", rid, owner)
			edges = append(edges, graphEdge{From: featNodeID(f.Name), To: reqNode, Type: "has_requirement"})

			for _, code := range req.Code {
				cid := "code_" + mermaidID(code)
				addNode(cid, "code", code, owner)
				edges = append(edges, graphEdge{From: reqNode, To: cid, Type: "implemented_by"})
			}
			for _, test := range req.Test {
				tid := "test_" + mermaidID(test)
				addNode(tid, "test", test, owner)
				edges = append(edges, graphEdge{From: reqNode, To: tid, Type: "verified_by"})
			}
		}
	}

	return knowledgeGraph{SchemaVersion: "1.0", Scope: scope, Nodes: nodes, Edges: edges}, 0
}

// readTraceQuiet lee y parsea un trace.json sin imprimir (a diferencia de las
// rutas de `sf trace`, que reportan). Devuelve (_, false) si falla — degradación
// segura: una feature con trace roto no rompe el grafo entero.
func readTraceQuiet(path string) (traceFile, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return traceFile{}, false
	}
	var tf traceFile
	if json.Unmarshal(data, &tf) != nil {
		return traceFile{}, false
	}
	return tf, true
}

// ----------------------------------------------------------------------------
// IDs y render Mermaid.
// ----------------------------------------------------------------------------

func featNodeID(name string) string     { return "feat_" + mermaidID(name) }
func reqNodeID(feat, rid string) string { return "req_" + mermaidID(feat) + "_" + mermaidID(rid) }

// mermaidID convierte un string arbitrario (un path:símbolo, un nombre) en un id
// seguro para Mermaid: solo [A-Za-z0-9_]. Mermaid rompe con `/`, `:`, `.`, etc.
// en los identificadores de nodo, así que los mapeamos a `_`.
func mermaidID(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteByte('_')
		}
	}
	return b.String()
}

// mermaidLabel escapa el texto que va DENTRO de las comillas de un nodo. Mermaid
// no admite comillas dobles ahí; las pasamos a simples.
func mermaidLabel(s string) string {
	return strings.ReplaceAll(s, "\"", "'")
}

// mermaidShape devuelve la declaración Mermaid de un nodo con la FORMA según su
// tipo, para leerse de un vistazo:
//
//	feature      ([...])  estadio
//	requirement  [...]    rectángulo
//	code         [/.../]  paralelogramo
//	test         {{...}}  hexágono
func mermaidShape(n graphNode) string {
	label := mermaidLabel(n.Label)
	switch n.Type {
	case "feature":
		return fmt.Sprintf("%s([\"%s\"])", n.ID, label)
	case "code":
		return fmt.Sprintf("%s[/\"%s\"/]", n.ID, label)
	case "test":
		return fmt.Sprintf("%s{{\"%s\"}}", n.ID, label)
	default: // requirement (y fallback)
		return fmt.Sprintf("%s[\"%s\"]", n.ID, label)
	}
}

// renderGraphMermaid emite un `flowchart LR` donde cada feature es una CAJA
// (subgraph) que agrupa su nodo de feature + sus R/code/test. Los hijos se
// reparten por su campo Group (= id del nodo-feature dueño). Las aristas se emiten
// al final, en el nivel raíz: Mermaid las dibuja cruzando las cajas sin problema,
// así que las dependencias feature→feature conectan una caja con otra.
func renderGraphMermaid(g knowledgeGraph) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%%%% SpecForge knowledge graph — scope: %s (generated by `sf graph export`)\n", g.Scope)
	b.WriteString("flowchart LR\n")

	// Separamos los nodos de feature (headers de caja) de sus hijos, indexando
	// los hijos por su feature dueña. Preservamos el orden de g.Nodes (estable).
	var featureNodes []graphNode
	childrenByOwner := map[string][]graphNode{}
	for _, n := range g.Nodes {
		if n.Type == "feature" {
			featureNodes = append(featureNodes, n)
		} else {
			childrenByOwner[n.Group] = append(childrenByOwner[n.Group], n)
		}
	}

	// Una caja por feature: título = nombre; adentro el nodo de feature (ancla de
	// las aristas) y sus hijos.
	for _, fn := range featureNodes {
		title := mermaidLabel(strings.TrimPrefix(fn.Label, "feature: "))
		fmt.Fprintf(&b, "  subgraph sg_%s [\"%s\"]\n", fn.ID, title)
		fmt.Fprintf(&b, "    %s\n", mermaidShape(fn))
		for _, c := range childrenByOwner[fn.ID] {
			fmt.Fprintf(&b, "    %s\n", mermaidShape(c))
		}
		b.WriteString("  end\n")
	}

	// Salvaguarda: hijos cuyo dueño no quedó en el grafo (no debería pasar — solo
	// agregamos hijos de features en scope) se emiten sueltos para no perderlos.
	for owner, kids := range childrenByOwner {
		if seenOwner(featureNodes, owner) {
			continue
		}
		for _, c := range kids {
			fmt.Fprintf(&b, "  %s\n", mermaidShape(c))
		}
	}

	for _, e := range g.Edges {
		fmt.Fprintf(&b, "  %s -->|%s| %s\n", e.From, e.Type, e.To)
	}
	return b.String()
}

// seenOwner indica si `owner` corresponde a un nodo de feature presente.
func seenOwner(featureNodes []graphNode, owner string) bool {
	for _, fn := range featureNodes {
		if fn.ID == owner {
			return true
		}
	}
	return false
}
