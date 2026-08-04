package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// ----------------------------------------------------------------------------
// Estado POR FEATURE (A1/R3, ataca D2).
//
// El problema del ledger monolítico: toda feature, todo gate, toda transición
// mutaba UN solo features.json → imán de conflictos de merge en equipo, flujo
// serial casi forzado por la estructura, y un solo archivo comprometido = todo
// el historial comprometido.
//
// El modelo nuevo:
//   - specforge/features/<name>/feature.json  — la VERDAD de esa feature
//     (status, lane, depends_on, gates). Cada feature es su propio archivo:
//     dos branches que aprueban gates de features DISTINTAS ya no colisionan.
//   - El "features.json global" pasa a ser una VISTA DERIVADA: readFeaturesFile
//     la ensambla en memoria desde los archivos por feature (estilo `sf state`:
//     derivado, no almacenado — nada que sincronizar).
//
// Transición sin quiebre: readFeaturesFile sigue leyendo el features.json
// legacy si existe, y los feature.json por feature LO PISAN por nombre (el
// escritor nuevo escribe siempre por feature). `sf migrate` hace el split
// completo y retira el legacy. Así un proyecto viejo funciona igual antes,
// durante y después de migrar.
// ----------------------------------------------------------------------------

// featureFilePath: la ruta del estado propio de una feature.
func featureFilePath(projectDir, name string) string {
	return filepath.Join(projectDir, "specforge", "features", name, "feature.json")
}

// readFeaturesFile ensambla la vista global del estado de features:
//  1. base: el features.json legacy si todavía existe (define el ORDEN de las
//     features que registra — el orden histórico del array).
//  2. override/extensión: cada specforge/features/*/feature.json pisa a su
//     homónima del legacy o se agrega como nueva (ordenada por seq, luego nombre).
//
// Error solo si NO existe ninguna fuente de estado (mismo contrato que antes:
// "no hay proyecto que leer").
func readFeaturesFile(projectDir string) (featuresFile, error) {
	var ff featuresFile
	legacyPath := filepath.Join(projectDir, "specforge", "features.json")

	haveLegacy := false
	if data, err := os.ReadFile(legacyPath); err == nil {
		// Un legacy presente pero corrupto ES un error: estado roto debe verse,
		// no taparse con la vista por feature.
		if err := json.Unmarshal(data, &ff); err != nil {
			return featuresFile{}, err
		}
		haveLegacy = true
	}

	perFeature, err := readPerFeatureStates(projectDir)
	if err != nil {
		return featuresFile{}, err
	}
	if !haveLegacy && len(perFeature) == 0 {
		return featuresFile{}, fmt.Errorf(
			"no feature state under %s/specforge (neither features.json nor features/*/feature.json)", projectDir)
	}

	// Merge: por nombre, el archivo por feature es más nuevo que el legacy → gana.
	byName := map[string]int{} // nombre → índice en ff.Features
	for i := range ff.Features {
		byName[ff.Features[i].Name] = i
	}
	var extra []feature
	for _, pf := range perFeature {
		if i, ok := byName[pf.Name]; ok {
			ff.Features[i] = pf
		} else {
			extra = append(extra, pf)
		}
	}
	// Las que solo existen por feature van al final, en orden de creación (seq);
	// seq 0 (sin numerar) al fondo, desempatado por nombre — orden ESTABLE, que
	// no dependa del orden de listado del filesystem.
	sort.Slice(extra, func(i, j int) bool {
		si, sj := seqOrInf(extra[i].Seq), seqOrInf(extra[j].Seq)
		if si != sj {
			return si < sj
		}
		return extra[i].Name < extra[j].Name
	})
	ff.Features = append(ff.Features, extra...)

	if ff.SchemaVersion == "" {
		ff.SchemaVersion = schemaVersionCurrent
	}
	return ff, nil
}

func seqOrInf(s int) int {
	if s == 0 {
		return int(^uint(0) >> 1) // MaxInt: sin seq → al final
	}
	return s
}

// readPerFeatureStates lee todos los specforge/features/*/feature.json. Un
// archivo corrupto es error (estado roto debe frenar, no desaparecer de la vista).
func readPerFeatureStates(projectDir string) ([]feature, error) {
	pattern := filepath.Join(projectDir, "specforge", "features", "*", "feature.json")
	paths, _ := filepath.Glob(pattern) // el único error posible es un pattern malformado
	var out []feature
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue // desapareció entre glob y read: lo salteamos
		}
		var f feature
		if err := json.Unmarshal(data, &f); err != nil {
			return nil, fmt.Errorf("invalid %s (%v)", p, err)
		}
		if f.Name == "" {
			// El nombre canónico es el del dir que lo contiene.
			f.Name = filepath.Base(filepath.Dir(p))
		}
		out = append(out, f)
	}
	return out, nil
}

// writeFeatureState es EL escritor del estado de una feature: escribe su
// feature.json canónico (con schema_version). Todos los mutadores (gate
// approve, feature add/set-status/set-lane/archive, migrate) pasan por acá —
// nadie reescribe el estado de una feature que no tocó.
func writeFeatureState(projectDir string, f feature) int {
	if f.SchemaVersion == "" {
		f.SchemaVersion = schemaVersionCurrent
	}
	path := featureFilePath(projectDir, f.Name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		fmt.Fprintf(os.Stderr, "sf: %v\n", err)
		return 1
	}
	out, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		fmt.Fprintf(os.Stderr, "sf: marshal failed (%v)\n", err)
		return 1
	}
	if err := os.WriteFile(path, append(out, '\n'), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "sf: cannot write %s (%v)\n", path, err)
		return 1
	}
	return 0
}

// nextSeq: el orden de creación de la próxima feature (máximo seq conocido + 1;
// las legacy sin seq cuentan por su posición para no chocar).
func nextSeq(ff featuresFile) int {
	max := len(ff.Features)
	for i := range ff.Features {
		if ff.Features[i].Seq > max {
			max = ff.Features[i].Seq
		}
	}
	return max + 1
}
