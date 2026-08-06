package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ----------------------------------------------------------------------------
// `sf migrate` — R6 de EVALUACION-PLATAFORMA §3: versionado de schemas ANTES
// del primer usuario externo.
//
// El día que un artefacto cambie de forma (tasks.json v2, p.ej.), todo proyecto
// existente rompe SIN camino de migración — y retrofitear versionado con
// usuarios afuera es de las cosas imposibles-sin-dolor. Por eso este comando
// existe HOY, cuando la única migración real es trivial:
//
//   - estampar `schema_version: "1.0"` donde falte (artefactos pre-versionado)
//   - detectar versiones DESCONOCIDAS (artefacto de un CLI más nuevo → este
//     binario no debe tocarlo: upgrade de sf, no downgrade del artefacto)
//   - rellenar la cadena de integridad (`prev`, R1) en ledgers legacy — el
//     camino LEGÍTIMO para que un proyecto pre-cadena quede 100% encadenado
//
// Sutileza que hace a migrate no-trivial: estampar schema_version re-serializa
// el artefacto → cambia su hash canónico → el sello del gate que lo aprobó
// dejaría de coincidir (falso "silent edit"). migrate lo resuelve RE-SELLANDO:
// si el sello era válido ANTES de estampar, se actualiza al hash nuevo. Solo
// migrate hace esto — es el único comando autorizado a tocar sellos, porque
// migra formato, no contenido.
// ----------------------------------------------------------------------------

// schemaVersionCurrent es LA versión que escribe este binario. Cuando exista
// una "2.0", este archivo gana la lógica real de conversión 1.0→2.0.
// RM-MIG: la 2.0 es el modelo de requisito v2 — `acceptance` con id por criterio
// (RM-C1), y con él el contrato de verificación a nivel criterio. Los demás
// campos que trajo el bloque (`priority`, `kind`, `verification`, `source` como
// refs) NO necesitan migración: todos tienen default implícito y fail-closed, así
// que un artefacto que los omite ya se comporta como corresponde. Materializarlos
// sería agregar bytes a cada requisito para cero cambio de comportamiento.
const schemaVersionCurrent = "2.0"

// knownSchemaVersions: las versiones que este binario sabe leer. La 1.0 sigue
// acá porque leerla es justamente lo que permite migrarla.
var knownSchemaVersions = map[string]bool{"1.0": true, "2.0": true}

// migrateAction registra un cambio hecho (o por hacer, en --dry-run).
type migrateAction struct {
	Path   string `json:"path"`
	Action string `json:"action"` // stamp-version | reseal-gate | chain-ledger
	Detail string `json:"detail,omitempty"`
}

func runMigrate(args []string) int {
	projectDir := "."
	dryRun := false
	for _, a := range args {
		switch {
		case a == "--dry-run":
			dryRun = true
		case strings.HasPrefix(a, "-"):
			fmt.Fprintf(os.Stderr, "sf migrate: unknown flag %q\n", a)
			return 2
		default:
			projectDir = a
		}
	}

	sf, ok := specforgeRoot(projectDir)
	if !ok {
		fmt.Println("No specforge/ directory — nothing to migrate.")
		return 0
	}

	ff, err := readFeaturesFile(projectDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf migrate: cannot read features.json (%v)\n", err)
		return 4
	}

	// Guard: NO migramos sobre un ledger con problemas de integridad reales.
	// Rellenar la cadena de un ledger forjado lavaría la forja. (Un ledger 100%
	// legacy no reporta problemas → pasa.)
	if probs := ledgerProblemsAll(ff); len(probs) > 0 {
		fmt.Println("sf migrate: REFUSED — the ledger fails integrity validation; migrating would launder it:")
		for _, p := range probs {
			fmt.Printf("  - %s\n", p)
		}
		fmt.Println("Restore the ledger first (see `sf recover`).")
		return 5
	}

	var actions []migrateAction
	var unknown []string
	// upgraded: cuántos requisitos pasaron a criterios con id. No es cosmético —
	// es la medida de una consecuencia que hay que avisar (reportContractTightened).
	upgraded := 0

	// 1) Artefactos: versión faltante → estampar; desconocida → error.
	//    Si estampamos, re-sellamos los gates cuyo sello era válido antes.
	stampArtifacts(projectDir, sf, &ff, dryRun, &actions, &unknown, &upgraded)

	// 2) Split del estado (A1/R3): si el features.json legacy todavía existe, el
	//    estado pasa a vivir POR FEATURE (features/<name>/feature.json) y el
	//    legacy se retira. También numeramos (seq) y estampamos versión donde falte.
	legacyPath := filepath.Join(sf, "features.json")
	splitLegacy := fileExists(legacyPath)
	if ff.SchemaVersion != "" && !knownSchemaVersions[ff.SchemaVersion] {
		unknown = append(unknown, "feature state: "+ff.SchemaVersion)
	}
	for i := range ff.Features {
		f := &ff.Features[i]
		if f.Seq == 0 {
			f.Seq = i + 1 // el orden histórico del array pasa a ser explícito
			if !splitLegacy {
				actions = append(actions, migrateAction{
					Path: "specforge/features/" + f.Name + "/feature.json", Action: "stamp-seq"})
			}
		}
		switch {
		case f.SchemaVersion == "":
			f.SchemaVersion = schemaVersionCurrent
			if !splitLegacy {
				actions = append(actions, migrateAction{
					Path: "specforge/features/" + f.Name + "/feature.json", Action: "stamp-version"})
			}
		case !knownSchemaVersions[f.SchemaVersion]:
			unknown = append(unknown, fmt.Sprintf("features/%s/feature.json: %s", f.Name, f.SchemaVersion))
		}
	}
	if splitLegacy {
		actions = append(actions, migrateAction{
			Path:   "specforge/features.json",
			Action: "split-state",
			Detail: fmt.Sprintf("%d feature(s) → specforge/features/<name>/feature.json; legacy file removed", len(ff.Features)),
		})
	}

	// 3) Cadena de integridad: rellenar `prev` donde falte. Como el paso 1 pudo
	//    re-sellar hashes (cambia el contenido de entradas), recomputamos la
	//    cadena ENTERA — sobre un ledger ya validado como íntegro, es seguro.
	for i := range ff.Features {
		f := &ff.Features[i]
		if chainLedger(f) {
			actions = append(actions, migrateAction{
				Path:   "specforge/features.json",
				Action: "chain-ledger",
				Detail: fmt.Sprintf("feature %q: %d gate(s) chained", f.Name, len(f.Gates)),
			})
		}
	}

	// Versiones desconocidas = este binario es viejo para el proyecto. No
	// tocamos nada más y salimos con error: la migración correcta es actualizar sf.
	if len(unknown) > 0 {
		fmt.Println("sf migrate: artifacts with UNKNOWN schema_version (written by a newer sf?):")
		for _, u := range unknown {
			fmt.Printf("  - %s\n", u)
		}
		fmt.Printf("This binary understands: %s. Upgrade sf instead of editing the artifacts.\n",
			strings.Join(sortedKeys(knownSchemaVersions), ", "))
		return 5
	}

	// Las features archivadas se saltean, Y SE DICE. Es lo correcto según la
	// estrategia de contrato versionado: una feature ya sellada no se re-evalúa
	// ni se re-sella bajo el contrato nuevo — reescribir un sello para que cumpla
	// una regla que no existía cuando se selló es exactamente lo que el sello
	// existe para impedir. Pero saltearlas EN SILENCIO haría creer que el
	// proyecto entero subió a 2.0, y no es así: sus artefactos siguen en 1.0 y
	// corren bajo contrato v1, que es como fueron aprobados.
	skipped := archivedFeatureCount(sf)

	if len(actions) == 0 {
		fmt.Println("OK — everything already at schema " + schemaVersionCurrent + ", ledger fully chained. Nothing to migrate.")
		reportArchivedSkip(skipped)
		return 0
	}

	if dryRun {
		fmt.Printf("Would apply %d change(s):\n", len(actions))
	} else {
		// Persistimos POR FEATURE (el único formato de escritura desde A1) y
		// retiramos el legacy AL FINAL: si algo falla a mitad de camino, el
		// legacy sigue ahí y readFeaturesFile ensambla la vista igual (los
		// feature.json ya escritos simplemente lo pisan por nombre).
		for i := range ff.Features {
			if code := writeFeatureState(projectDir, ff.Features[i]); code != 0 {
				return code
			}
		}
		if splitLegacy {
			if err := os.Remove(legacyPath); err != nil {
				fmt.Fprintf(os.Stderr, "sf migrate: cannot remove legacy features.json (%v)\n", err)
				return 1
			}
		}
		fmt.Printf("Applied %d change(s):\n", len(actions))
	}
	for _, a := range actions {
		line := fmt.Sprintf("  - %s: %s", a.Path, a.Action)
		if a.Detail != "" {
			line += " (" + a.Detail + ")"
		}
		fmt.Println(line)
	}
	reportContractTightened(upgraded)
	reportArchivedSkip(skipped)
	return 0
}

// reportContractTightened avisa la consecuencia que la migración NO puede
// resolver por su cuenta.
//
// Al darle ids a los criterios, el requisito pasa del contrato v1 (un test por
// requisito) al v2 (un test POR CRITERIO). Pero el `trace.json` sigue anclando a
// nivel requisito, y repartir esos tests entre criterios es CONTENIDO — decidir
// cuál test cubre cuál caso necesita un modelo, que es justo lo que migrate no
// puede hacer sin perder la autorización que tiene para re-sellar.
//
// El resultado es correcto pero contraintuitivo: justo después de migrar, la
// cobertura por contrato cae y el próximo veredicto rechaza. Descubierto
// migrando `examples/brownfield-tempconv`, que pasó de 3/3 a 0/3 sin que nada lo
// dijera. Un costo real y silencioso es cómo una migración se revierte a mano.
func reportContractTightened(n int) {
	if n == 0 {
		return
	}
	fmt.Printf("\n%d requirement(s) now carry per-criterion acceptance ids — their verification\n", n)
	fmt.Println("contract just got stricter: the verdict will ask for a test per CRITERION, not per")
	fmt.Println("requirement. Their traces still anchor at requirement level, so expect them to read")
	fmt.Println("as uncovered until each criterion gets its own anchor under")
	fmt.Println("`trace.json` → requirements.<R#>.scenarios.<R#.n>.")
	fmt.Println("Splitting those anchors is a content decision, so migrate cannot do it for you.")
	fmt.Println("See exactly which ones with `sf coverage --by-priority`.")
}

// archivedFeatureCount cuenta las features archivadas que tienen artefactos.
func archivedFeatureCount(sf string) int {
	matches, _ := filepath.Glob(filepath.Join(sf, "archive", "*", "requirements.json"))
	return len(matches)
}

// reportArchivedSkip explica la omisión en vez de dejarla implícita.
func reportArchivedSkip(n int) {
	if n == 0 {
		return
	}
	fmt.Printf("\n%d archived feature(s) left at their sealed schema — not migrated, on purpose.\n", n)
	fmt.Println("They were approved under the contract of their time, and rewriting a seal to satisfy " +
		"a rule that didn't exist then is what the seal exists to prevent. They keep running under " +
		"contract v1; `sf-amend` is the explicit way to bring one up.")
}

// stampArtifacts recorre todos los artefactos .json existentes; estampa la
// versión faltante y acumula las desconocidas. Con dryRun no escribe.
func stampArtifacts(projectDir, sf string, ff *featuresFile, dryRun bool, actions *[]migrateAction, unknown *[]string, upgraded *int) {
	type target struct {
		abs   string
		owner *feature // dueña del gate a re-sellar (nil: constitution/domain, sin sello)
	}
	var targets []target

	targets = append(targets,
		target{filepath.Join(sf, "constitution.json"), nil},
		target{domainPath(projectDir) + ".json", nil},
	)
	perFeature := []string{
		"requirements.json", "design.json", "tasks.json",
		filepath.Join("progress", "plan.json"), "review.json", "trace.json",
	}
	for i := range ff.Features {
		f := &ff.Features[i]
		for _, rel := range perFeature {
			targets = append(targets, target{filepath.Join(sf, "features", f.Name, rel), f})
		}
	}

	for _, tg := range targets {
		raw, err := os.ReadFile(tg.abs)
		if err != nil {
			continue // ausente: la fase no llegó — nada que migrar
		}
		// Decodificamos a map (no al struct tipado) para NO descartar claves que
		// este binario no conozca: migrate agrega schema_version, no reescribe.
		var doc map[string]any
		if json.Unmarshal(raw, &doc) != nil {
			continue // ilegible: eso lo reporta `sf verify`, no es trabajo de migrate
		}

		relPath := filepath.ToSlash(tg.abs)
		v, _ := doc["schema_version"].(string)

		// Una versión que este binario no conoce = el proyecto viene de un sf más
		// nuevo. No se toca: la migración correcta es actualizar sf, no degradar
		// el artefacto.
		if v != "" && !knownSchemaVersions[v] {
			*unknown = append(*unknown, relPath+": "+v)
			continue
		}
		if v == schemaVersionCurrent {
			continue // ya está donde queremos
		}

		// Los cambios de FORMATO que separan a este artefacto de la versión
		// actual. Todos deterministas: ninguno decide contenido.
		var did []string
		if v == "" || v == "1.0" {
			if n := upgradeAcceptanceToV2(doc, filepath.Base(tg.abs)); n > 0 {
				did = append(did, fmt.Sprintf("acceptance ids on %d requirement(s)", n))
				*upgraded += n
			}
		}
		doc["schema_version"] = schemaVersionCurrent

		action := migrateAction{Path: relPath, Action: "stamp-version"}
		if len(did) > 0 {
			action.Action = "upgrade-" + schemaVersionCurrent
			action.Detail = strings.Join(did, ", ")
		}
		*actions = append(*actions, action)
		if dryRun {
			continue
		}

		before, _ := hashArtifact(tg.abs) // el hash canónico PRE-cambio
		out, err := json.MarshalIndent(doc, "", "  ")
		if err != nil {
			continue
		}
		if os.WriteFile(tg.abs, append(out, '\n'), 0o644) != nil {
			continue
		}
		// Re-sellado: todo gate de la feature dueña cuyo sello coincidía con
		// el hash PRE-cambio pasa al hash nuevo. Un sello que YA no
		// coincidía (silent edit real) se deja como está: migrate no amnistía.
		if tg.owner != nil {
			after, _ := hashArtifact(tg.abs)
			for gi := range tg.owner.Gates {
				if g := &tg.owner.Gates[gi]; g.Hash != "" && g.Hash == before {
					g.Hash = after
					*actions = append(*actions, migrateAction{
						Path:   relPath,
						Action: "reseal-gate",
						Detail: fmt.Sprintf("%s/%s", tg.owner.Name, g.Phase),
					})
				}
			}
		}
	}
}

// upgradeAcceptanceToV2 convierte `acceptance: []string` en criterios con id,
// asignados POR ÍNDICE: `acceptance[0]` → `R5.1`, `acceptance[1]` → `R5.2`, y el
// string original pasa a `text` sin tocarse.
//
// POR QUÉ ESTO SÍ ES FORMATO Y PARTIRLO EN GIVEN/WHEN/THEN NO LO SERÍA. `migrate`
// está autorizado a re-sellar precisamente porque migra formato; hacerle tocar
// contenido rompería esa justificación y con ella la validez del sello. Decidir
// cuál parte de una frase es la precondición y cuál el resultado requiere un
// modelo — es contenido. Numerar por índice no requiere ninguno: dos corridas
// sobre el mismo archivo producen exactamente los mismos ids.
//
// Es lo que `DEC-1` destrabó. Con `when`/`then` obligatorios la migración era
// imposible y el sistema quedaba BIMODAL PARA SIEMPRE (legado con contrato débil,
// v2 con contrato fuerte, y cada feature subiendo a mano por `sf-amend`). Con id
// obligatorio y G/W/T opcional, todo el legado sube solo.
// Devuelve CUÁNTOS requisitos convirtió, no un booleano: ese número es el que
// mide la consecuencia que hay que avisar (ver reportContractTightened).
func upgradeAcceptanceToV2(doc map[string]any, base string) int {
	if base != "requirements.json" {
		return 0
	}
	reqs, ok := doc["requirements"].([]any)
	if !ok {
		return 0
	}
	changed := 0
	for _, r := range reqs {
		req, ok := r.(map[string]any)
		if !ok {
			continue
		}
		id, _ := req["id"].(string)
		list, ok := req["acceptance"].([]any)
		if !ok || id == "" || len(list) == 0 {
			continue
		}
		// Se convierte SÓLO si TODOS los elementos son strings. Una lista mixta
		// significa que alguien ya la tocó a mano, y adivinar qué quiso hacer es
		// exactamente el tipo de decisión que migrate no puede tomar: se deja
		// como está y el gate de su feature la reportará.
		allStrings := true
		for _, it := range list {
			if _, isStr := it.(string); !isStr {
				allStrings = false
				break
			}
		}
		if !allStrings {
			continue
		}
		out := make([]any, 0, len(list))
		for i, it := range list {
			out = append(out, map[string]any{
				"id":   fmt.Sprintf("%s.%d", id, i+1),
				"text": it.(string),
			})
		}
		req["acceptance"] = out
		changed++
	}
	return changed
}

// chainLedger recomputa `prev` de TODAS las entradas de una feature (genesis,
// luego hash-de-la-anterior). Devuelve true si algo cambió. Solo se llama tras
// validar la integridad — sobre un ledger legítimo es una operación de formato.
func chainLedger(f *feature) bool {
	changed := false
	for i := range f.Gates {
		want := genesisPrev
		if i > 0 {
			want = gateEntryHash(f.Gates[i-1])
		}
		if f.Gates[i].Prev != want {
			f.Gates[i].Prev = want
			changed = true
		}
	}
	return changed
}
