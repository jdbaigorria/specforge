package main

import (
	"path/filepath"
	"testing"
)

// validDomain es un domain.json correcto que cada subtest puede mutar.
func validDomain() domainFile {
	return domainFile{
		SchemaVersion: "1.0",
		Glossary:      []glossaryTerm{{Term: "Order", Definition: "A confirmed purchase intent."}},
		Entities:      []entity{{ID: "E1", Name: "Order", Description: "..."}},
		Rules: []domainRule{
			{ID: "D1", Rule: "no duplicate orders", Entities: []string{"E1"}, AppliesTo: []string{"design", "build"}},
		},
	}
}

func TestCheckDomain(t *testing.T) {
	t.Run("valid -> no errors", func(t *testing.T) {
		var rep report
		checkDomain(validDomain(), &rep)
		if len(rep.errors) != 0 {
			t.Errorf("errors=%v, want none", rep.errors)
		}
	})

	t.Run("duplicate entity id -> error", func(t *testing.T) {
		df := validDomain()
		df.Entities = append(df.Entities, entity{ID: "E1", Name: "Dup"})
		var rep report
		checkDomain(df, &rep)
		if len(rep.errors) == 0 {
			t.Error("want an error for duplicate entity id")
		}
	})

	t.Run("rule references unknown entity -> error", func(t *testing.T) {
		df := validDomain()
		df.Rules[0].Entities = []string{"E9"} // no existe
		var rep report
		checkDomain(df, &rep)
		if len(rep.errors) == 0 {
			t.Error("want an error for unknown entity ref")
		}
	})

	t.Run("applies_to unknown phase -> error", func(t *testing.T) {
		df := validDomain()
		df.Rules[0].AppliesTo = []string{"nope"}
		var rep report
		checkDomain(df, &rep)
		if len(rep.errors) == 0 {
			t.Error("want an error for unknown applies_to phase")
		}
	})

	t.Run("id not in E#/D# form -> warning", func(t *testing.T) {
		df := validDomain()
		df.Entities[0].ID = "Order" // not E#
		df.Rules[0].Entities = nil  // avoid the now-dangling ref error
		df.Rules[0].ID = "rule-dup" // not D#
		var rep report
		checkDomain(df, &rep)
		if len(rep.warnings) < 2 {
			t.Errorf("warnings=%v, want >=2 (E# and D# form)", rep.warnings)
		}
	})

	t.Run("empty rule text -> error", func(t *testing.T) {
		df := validDomain()
		df.Rules[0].Rule = ""
		var rep report
		checkDomain(df, &rep)
		if len(rep.errors) == 0 {
			t.Error("want an error for empty rule")
		}
	})
}

func TestDomainRulesForPhase(t *testing.T) {
	df := validDomain() // D1 applies_to design+build
	if got := domainRulesForPhase(df, "design"); len(got) != 1 {
		t.Errorf("design rules=%d, want 1", len(got))
	}
	if got := domainRulesForPhase(df, "requirements"); len(got) != 0 {
		t.Errorf("requirements rules=%d, want 0", len(got))
	}
}

func TestLoadDomainQuiet(t *testing.T) {
	t.Run("missing -> false", func(t *testing.T) {
		if _, ok := loadDomainQuiet(t.TempDir()); ok {
			t.Error("want false when domain.json is absent")
		}
	})
	t.Run("present -> true", func(t *testing.T) {
		dir := t.TempDir()
		writeFile(t, filepath.Join(dir, "specforge/context/domain.json"),
			`{"schema_version":"1.0","entities":[{"id":"E1","name":"X"}]}`)
		df, ok := loadDomainQuiet(dir)
		if !ok || len(df.Entities) != 1 {
			t.Errorf("ok=%v entities=%d, want true,1", ok, len(df.Entities))
		}
	})
}

// TestForJudgeInjectsDomain comprueba que for-judge inyecta solo las reglas de la
// fase pedida (la inversión applies_to → fase, sobre domain.json).
func TestForJudgeInjectsDomain(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "specforge/context/domain.json"),
		`{"schema_version":"1.0","rules":[{"id":"D1","rule":"r","applies_to":["design"]}]}`)
	df, _ := loadDomainQuiet(dir)
	if len(domainRulesForPhase(df, "design")) != 1 {
		t.Error("D1 should apply to design")
	}
	if len(domainRulesForPhase(df, "build")) != 0 {
		t.Error("D1 should NOT apply to build")
	}
}
