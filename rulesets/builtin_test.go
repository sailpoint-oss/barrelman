package rulesets

import (
	"os"
	"testing"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/analyzers"
)

func TestMain(m *testing.M) {
	analyzers.RegisterAll(barrelman.DefaultRegistry)
	os.Exit(m.Run())
}

func TestGetBuiltin_Recommended(t *testing.T) {
	rs := GetBuiltin("telescope:recommended")
	if rs == nil {
		t.Fatal("expected non-nil ruleset for telescope:recommended")
	}
	if len(rs.Rules) == 0 {
		t.Fatal("expected recommended ruleset to contain rules")
	}

	// Every rule in the recommended set must be marked Recommended in the registry.
	for id := range rs.Rules {
		meta, ok := barrelman.DefaultRegistry.Get(id)
		if !ok {
			t.Errorf("rule %q in recommended ruleset not found in registry", id)
			continue
		}
		if !meta.Recommended {
			t.Errorf("rule %q in recommended ruleset is not marked Recommended", id)
		}
	}

	// Ensure no recommended rules are missing from the set.
	for _, meta := range barrelman.DefaultRegistry.All() {
		if meta.Recommended {
			if _, ok := rs.Rules[meta.ID]; !ok {
				t.Errorf("recommended rule %q missing from telescope:recommended", meta.ID)
			}
		}
	}
}

func TestGetBuiltin_All(t *testing.T) {
	rs := GetBuiltin("telescope:all")
	if rs == nil {
		t.Fatal("expected non-nil ruleset for telescope:all")
	}
	if len(rs.Rules) == 0 {
		t.Fatal("expected all ruleset to contain rules")
	}

	// The "all" ruleset must not contain any OWASP rules.
	for id := range rs.Rules {
		meta, ok := barrelman.DefaultRegistry.Get(id)
		if !ok {
			t.Errorf("rule %q in all ruleset not found in registry", id)
			continue
		}
		if meta.Category == barrelman.CategoryOWASP {
			t.Errorf("rule %q in all ruleset has OWASP category; all ruleset should exclude OWASP", id)
		}
	}
}

func TestGetBuiltin_OWASP(t *testing.T) {
	rs := GetBuiltin("telescope:owasp")
	if rs == nil {
		t.Fatal("expected non-nil ruleset for telescope:owasp")
	}
	if len(rs.Rules) == 0 {
		t.Fatal("expected owasp ruleset to contain rules")
	}

	// Every rule should be in the OWASP category.
	for id := range rs.Rules {
		meta, ok := barrelman.DefaultRegistry.Get(id)
		if !ok {
			t.Errorf("rule %q in owasp ruleset not found in registry", id)
			continue
		}
		if meta.Category != barrelman.CategoryOWASP {
			t.Errorf("rule %q in owasp ruleset has category %q, want owasp", id, meta.Category)
		}
	}

	// Ensure all OWASP rules from the registry are present.
	for _, meta := range barrelman.DefaultRegistry.ByCategory(barrelman.CategoryOWASP) {
		if _, ok := rs.Rules[meta.ID]; !ok {
			t.Errorf("OWASP rule %q missing from telescope:owasp", meta.ID)
		}
	}
}

func TestGetBuiltin_Strict(t *testing.T) {
	rs := GetBuiltin("telescope:strict")
	if rs == nil {
		t.Fatal("expected non-nil ruleset for telescope:strict")
	}
	if len(rs.Rules) == 0 {
		t.Fatal("expected strict ruleset to contain rules")
	}

	recommended := GetBuiltin("telescope:recommended")
	owasp := GetBuiltin("telescope:owasp")

	// Strict must contain all recommended rules.
	for id := range recommended.Rules {
		if _, ok := rs.Rules[id]; !ok {
			t.Errorf("recommended rule %q missing from strict ruleset", id)
		}
	}

	// Strict must contain all OWASP rules.
	for id := range owasp.Rules {
		if _, ok := rs.Rules[id]; !ok {
			t.Errorf("OWASP rule %q missing from strict ruleset", id)
		}
	}

	// Total size should be at least the union of recommended and OWASP.
	expectedMin := len(recommended.Rules)
	// Add OWASP rules that are not already in recommended.
	for id := range owasp.Rules {
		if _, ok := recommended.Rules[id]; !ok {
			expectedMin++
		}
	}
	if len(rs.Rules) < expectedMin {
		t.Errorf("strict ruleset has %d rules, want at least %d", len(rs.Rules), expectedMin)
	}
}

func TestGetBuiltin_Nonexistent(t *testing.T) {
	rs := GetBuiltin("nonexistent")
	if rs != nil {
		t.Errorf("expected nil for nonexistent builtin, got %+v", rs)
	}
}

func TestGetBuiltin_Constants(t *testing.T) {
	// Verify constants match expected values.
	if Recommended != "telescope:recommended" {
		t.Errorf("Recommended = %q, want %q", Recommended, "telescope:recommended")
	}
	if All != "telescope:all" {
		t.Errorf("All = %q, want %q", All, "telescope:all")
	}
	if OWASP != "telescope:owasp" {
		t.Errorf("OWASP = %q, want %q", OWASP, "telescope:owasp")
	}
	if Strict != "telescope:strict" {
		t.Errorf("Strict = %q, want %q", Strict, "telescope:strict")
	}
}

func TestGetBuiltin_RuleSeverities(t *testing.T) {
	// Verify that all rules in built-in rulesets have valid severity strings.
	validSeverities := map[string]bool{
		"error": true,
		"warn":  true,
		"info":  true,
		"hint":  true,
	}

	for _, name := range []string{Recommended, All, OWASP, Strict} {
		rs := GetBuiltin(name)
		if rs == nil {
			t.Fatalf("GetBuiltin(%q) returned nil", name)
		}
		for id, def := range rs.Rules {
			if !validSeverities[def.Severity] {
				t.Errorf("GetBuiltin(%q): rule %q has invalid severity %q", name, id, def.Severity)
			}
		}
	}
}

func TestGetBuiltin_AllExcludesOWASPButIncludesRecommended(t *testing.T) {
	all := GetBuiltin("telescope:all")
	recommended := GetBuiltin("telescope:recommended")

	// Every recommended non-OWASP rule should appear in the "all" set.
	for id := range recommended.Rules {
		meta, ok := barrelman.DefaultRegistry.Get(id)
		if !ok {
			continue
		}
		if meta.Category == barrelman.CategoryOWASP {
			continue
		}
		if _, ok := all.Rules[id]; !ok {
			t.Errorf("recommended rule %q not found in all ruleset", id)
		}
	}
}
