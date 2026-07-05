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
const schemaVersionCurrent = "1.0"

// knownSchemaVersions: las versiones que este binario sabe leer.
var knownSchemaVersions = map[string]bool{"1.0": true}

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

	// 1) Artefactos: versión faltante → estampar; desconocida → error.
	//    Si estampamos, re-sellamos los gates cuyo sello era válido antes.
	stampArtifacts(projectDir, sf, &ff, dryRun, &actions, &unknown)

	// 2) features.json: su propio schema_version.
	if ff.SchemaVersion == "" {
		ff.SchemaVersion = schemaVersionCurrent
		actions = append(actions, migrateAction{Path: "specforge/features.json", Action: "stamp-version"})
	} else if !knownSchemaVersions[ff.SchemaVersion] {
		unknown = append(unknown, "specforge/features.json: "+ff.SchemaVersion)
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

	if len(actions) == 0 {
		fmt.Println("OK — everything already at schema " + schemaVersionCurrent + ", ledger fully chained. Nothing to migrate.")
		return 0
	}

	if dryRun {
		fmt.Printf("Would apply %d change(s):\n", len(actions))
	} else {
		if code := writeFeaturesFile(filepath.Join(sf, "features.json"), ff); code != 0 {
			return code
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
	return 0
}

// stampArtifacts recorre todos los artefactos .json existentes; estampa la
// versión faltante y acumula las desconocidas. Con dryRun no escribe.
func stampArtifacts(projectDir, sf string, ff *featuresFile, dryRun bool, actions *[]migrateAction, unknown *[]string) {
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
		switch v, _ := doc["schema_version"].(string); {
		case v == "":
			before, _ := hashArtifact(tg.abs) // el hash canónico PRE-estampado
			doc["schema_version"] = schemaVersionCurrent
			*actions = append(*actions, migrateAction{Path: relPath, Action: "stamp-version"})
			if dryRun {
				continue
			}
			out, err := json.MarshalIndent(doc, "", "  ")
			if err != nil {
				continue
			}
			if os.WriteFile(tg.abs, append(out, '\n'), 0o644) != nil {
				continue
			}
			// Re-sellado: todo gate de la feature dueña cuyo sello coincidía con
			// el hash PRE-estampado pasa al hash nuevo. Un sello que YA no
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
		case !knownSchemaVersions[v]:
			*unknown = append(*unknown, relPath+": "+v)
		}
	}
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
