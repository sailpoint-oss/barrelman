package rulesets

import (
	"os"
	"reflect"
	"testing"
)

var testBuiltinCatalog = []CatalogRule{
	{ID: "info-description", Severity: SeverityWarning, Category: "documentation", Recommended: true},
	{ID: "sailpoint-operation-single-tag", Severity: SeverityError, Category: "documentation", Recommended: true},
	{ID: "owasp-no-api-keys-in-url", Severity: SeverityError, Category: "owasp", Recommended: false},
	{ID: "schema-name-capital", Severity: SeverityWarning, Category: "naming", Recommended: false},
}

func TestMain(m *testing.M) {
	SetBuiltinCatalogProvider(func() []CatalogRule {
		out := make([]CatalogRule, len(testBuiltinCatalog))
		copy(out, testBuiltinCatalog)
		return out
	})
	os.Exit(m.Run())
}

func TestGetBuiltin_Recommended(t *testing.T) {
	rs := GetBuiltin(Recommended)
	if rs == nil {
		t.Fatal("expected non-nil ruleset for barrelman:recommended")
	}
	if len(rs.Rules) == 0 {
		t.Fatal("expected recommended ruleset to contain rules")
	}

	// Every rule in the recommended set must be marked Recommended in the registry.
	for id := range rs.Rules {
		meta, ok := testCatalogRule(id)
		if !ok {
			t.Errorf("rule %q in recommended ruleset not found in registry", id)
			continue
		}
		if !meta.Recommended {
			t.Errorf("rule %q in recommended ruleset is not marked Recommended", id)
		}
	}

	// Ensure no recommended rules are missing from the set.
	for _, meta := range testBuiltinCatalog {
		if meta.Recommended {
			if _, ok := rs.Rules[meta.ID]; !ok {
				t.Errorf("recommended rule %q missing from barrelman:recommended", meta.ID)
			}
		}
	}
}

func TestGetBuiltin_All(t *testing.T) {
	rs := GetBuiltin(All)
	if rs == nil {
		t.Fatal("expected non-nil ruleset for barrelman:all")
	}
	if len(rs.Rules) == 0 {
		t.Fatal("expected all ruleset to contain rules")
	}

	// The "all" ruleset must not contain any OWASP rules.
	for id := range rs.Rules {
		meta, ok := testCatalogRule(id)
		if !ok {
			t.Errorf("rule %q in all ruleset not found in registry", id)
			continue
		}
		if meta.Category == "owasp" {
			t.Errorf("rule %q in all ruleset has OWASP category; all ruleset should exclude OWASP", id)
		}
	}
}

func TestGetBuiltin_OWASP(t *testing.T) {
	rs := GetBuiltin(OWASP)
	if rs == nil {
		t.Fatal("expected non-nil ruleset for barrelman:owasp")
	}
	if len(rs.Rules) == 0 {
		t.Fatal("expected owasp ruleset to contain rules")
	}

	// Every rule should be in the OWASP category.
	for id := range rs.Rules {
		meta, ok := testCatalogRule(id)
		if !ok {
			t.Errorf("rule %q in owasp ruleset not found in registry", id)
			continue
		}
		if meta.Category != "owasp" {
			t.Errorf("rule %q in owasp ruleset has category %q, want owasp", id, meta.Category)
		}
	}

	// Ensure all OWASP rules from the registry are present.
	for _, meta := range testBuiltinCatalog {
		if meta.Category != "owasp" {
			continue
		}
		if _, ok := rs.Rules[meta.ID]; !ok {
			t.Errorf("OWASP rule %q missing from barrelman:owasp", meta.ID)
		}
	}
}

func TestGetBuiltin_Strict(t *testing.T) {
	rs := GetBuiltin(Strict)
	if rs == nil {
		t.Fatal("expected non-nil ruleset for barrelman:strict")
	}
	if len(rs.Rules) == 0 {
		t.Fatal("expected strict ruleset to contain rules")
	}

	recommended := GetBuiltin(Recommended)
	owasp := GetBuiltin(OWASP)

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
	if Recommended != "barrelman:recommended" {
		t.Errorf("Recommended = %q, want %q", Recommended, "barrelman:recommended")
	}
	if All != "barrelman:all" {
		t.Errorf("All = %q, want %q", All, "barrelman:all")
	}
	if OWASP != "barrelman:owasp" {
		t.Errorf("OWASP = %q, want %q", OWASP, "barrelman:owasp")
	}
	if Strict != "barrelman:strict" {
		t.Errorf("Strict = %q, want %q", Strict, "barrelman:strict")
	}
	if LegacyRecommended != "telescope:recommended" {
		t.Errorf("LegacyRecommended = %q, want %q", LegacyRecommended, "telescope:recommended")
	}
	if LegacyAll != "telescope:all" {
		t.Errorf("LegacyAll = %q, want %q", LegacyAll, "telescope:all")
	}
	if LegacyOWASP != "telescope:owasp" {
		t.Errorf("LegacyOWASP = %q, want %q", LegacyOWASP, "telescope:owasp")
	}
	if LegacyStrict != "telescope:strict" {
		t.Errorf("LegacyStrict = %q, want %q", LegacyStrict, "telescope:strict")
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
	all := GetBuiltin(All)
	recommended := GetBuiltin(Recommended)

	// Every recommended non-OWASP rule should appear in the "all" set.
	for id := range recommended.Rules {
		meta, ok := testCatalogRule(id)
		if !ok {
			continue
		}
		if meta.Category == "owasp" {
			continue
		}
		if _, ok := all.Rules[id]; !ok {
			t.Errorf("recommended rule %q not found in all ruleset", id)
		}
	}
}

func testCatalogRule(id string) (CatalogRule, bool) {
	for _, rule := range testBuiltinCatalog {
		if rule.ID == id {
			return rule, true
		}
	}
	return CatalogRule{}, false
}

func TestGetBuiltin_LegacyAliasesResolveToCanonicalSets(t *testing.T) {
	tests := []struct {
		canonical string
		legacy    string
	}{
		{Recommended, LegacyRecommended},
		{All, LegacyAll},
		{OWASP, LegacyOWASP},
		{Strict, LegacyStrict},
	}

	for _, tc := range tests {
		canonical := GetBuiltin(tc.canonical)
		legacy := GetBuiltin(tc.legacy)
		if canonical == nil || legacy == nil {
			t.Fatalf("expected builtins for %q and %q", tc.canonical, tc.legacy)
		}
		if len(canonical.Rules) != len(legacy.Rules) {
			t.Fatalf("rule counts differ for %q and %q", tc.canonical, tc.legacy)
		}
		for id, def := range canonical.Rules {
			if !reflect.DeepEqual(legacy.Rules[id], def) {
				t.Fatalf("legacy alias %q differs from %q for rule %q", tc.legacy, tc.canonical, id)
			}
		}
	}
}
