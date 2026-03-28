package rulesets

import "testing"

func TestMerge_Empty(t *testing.T) {
	result := Merge()
	if result == nil {
		t.Fatal("Merge() returned nil, expected non-nil")
	}
	if len(result.Rules) != 0 {
		t.Errorf("Merge() rules count = %d, want 0", len(result.Rules))
	}
}

func TestMerge_NilInputs(t *testing.T) {
	result := Merge(nil, nil, nil)
	if result == nil {
		t.Fatal("Merge(nil...) returned nil, expected non-nil")
	}
	if len(result.Rules) != 0 {
		t.Errorf("Merge(nil...) rules count = %d, want 0", len(result.Rules))
	}
}

func TestMerge_SingleRuleset(t *testing.T) {
	rs := &RuleSet{
		Name: "test",
		Rules: map[string]RuleDefinition{
			"rule-a": {Severity: "error"},
			"rule-b": {Severity: "warn"},
		},
	}
	result := Merge(rs)
	if len(result.Rules) != 2 {
		t.Fatalf("rules count = %d, want 2", len(result.Rules))
	}
	if result.Rules["rule-a"].Severity != "error" {
		t.Errorf("rule-a severity = %q, want %q", result.Rules["rule-a"].Severity, "error")
	}
	if result.Rules["rule-b"].Severity != "warn" {
		t.Errorf("rule-b severity = %q, want %q", result.Rules["rule-b"].Severity, "warn")
	}
}

func TestMerge_LaterTakesPriority(t *testing.T) {
	rs1 := &RuleSet{
		Rules: map[string]RuleDefinition{
			"rule-a": {Severity: "error"},
			"rule-b": {Severity: "warn"},
		},
	}
	rs2 := &RuleSet{
		Rules: map[string]RuleDefinition{
			"rule-a": {Severity: "info"}, // override
			"rule-c": {Severity: "hint"}, // new
		},
	}
	result := Merge(rs1, rs2)

	if len(result.Rules) != 3 {
		t.Fatalf("rules count = %d, want 3", len(result.Rules))
	}
	// rule-a should be overridden by rs2.
	if result.Rules["rule-a"].Severity != "info" {
		t.Errorf("rule-a severity = %q, want %q (rs2 should override)", result.Rules["rule-a"].Severity, "info")
	}
	// rule-b should be preserved from rs1.
	if result.Rules["rule-b"].Severity != "warn" {
		t.Errorf("rule-b severity = %q, want %q", result.Rules["rule-b"].Severity, "warn")
	}
	// rule-c should come from rs2.
	if result.Rules["rule-c"].Severity != "hint" {
		t.Errorf("rule-c severity = %q, want %q", result.Rules["rule-c"].Severity, "hint")
	}
}

func TestMerge_ThreeRulesets(t *testing.T) {
	rs1 := &RuleSet{
		Rules: map[string]RuleDefinition{
			"rule-a": {Severity: "error"},
		},
	}
	rs2 := &RuleSet{
		Rules: map[string]RuleDefinition{
			"rule-a": {Severity: "warn"},
		},
	}
	rs3 := &RuleSet{
		Rules: map[string]RuleDefinition{
			"rule-a": {Severity: "hint"},
		},
	}

	result := Merge(rs1, rs2, rs3)
	// Last one wins.
	if result.Rules["rule-a"].Severity != "hint" {
		t.Errorf("rule-a severity = %q, want %q (last ruleset should win)", result.Rules["rule-a"].Severity, "hint")
	}
}

func TestMerge_WithNilInMiddle(t *testing.T) {
	rs1 := &RuleSet{
		Rules: map[string]RuleDefinition{
			"rule-a": {Severity: "error"},
		},
	}
	rs2 := &RuleSet{
		Rules: map[string]RuleDefinition{
			"rule-b": {Severity: "warn"},
		},
	}

	result := Merge(rs1, nil, rs2)
	if len(result.Rules) != 2 {
		t.Fatalf("rules count = %d, want 2", len(result.Rules))
	}
	if result.Rules["rule-a"].Severity != "error" {
		t.Errorf("rule-a severity = %q, want %q", result.Rules["rule-a"].Severity, "error")
	}
	if result.Rules["rule-b"].Severity != "warn" {
		t.Errorf("rule-b severity = %q, want %q", result.Rules["rule-b"].Severity, "warn")
	}
}

func TestBuildEnabledMap_Nil(t *testing.T) {
	result := BuildEnabledMap(nil)
	if result != nil {
		t.Errorf("BuildEnabledMap(nil) = %v, want nil", result)
	}
}

func TestBuildEnabledMap_WithRules(t *testing.T) {
	rs := &RuleSet{
		Rules: map[string]RuleDefinition{
			"rule-enabled": {Severity: "error"},
			"rule-warn":    {Severity: "warn"},
			"rule-info":    {Severity: "info"},
			"rule-hint":    {Severity: "hint"},
			"rule-off":     {Severity: "off"},
			"rule-false":   {Severity: "false"},
		},
	}

	enabled := BuildEnabledMap(rs)
	if len(enabled) != 6 {
		t.Fatalf("enabled map size = %d, want 6", len(enabled))
	}

	expects := map[string]bool{
		"rule-enabled": true,
		"rule-warn":    true,
		"rule-info":    true,
		"rule-hint":    true,
		"rule-off":     false,
		"rule-false":   false,
	}

	for id, want := range expects {
		got, ok := enabled[id]
		if !ok {
			t.Errorf("rule %q not found in enabled map", id)
			continue
		}
		if got != want {
			t.Errorf("enabled[%q] = %v, want %v", id, got, want)
		}
	}
}

func TestBuildEnabledMap_EmptyRuleset(t *testing.T) {
	rs := &RuleSet{
		Rules: map[string]RuleDefinition{},
	}
	enabled := BuildEnabledMap(rs)
	if len(enabled) != 0 {
		t.Errorf("enabled map size = %d, want 0", len(enabled))
	}
}

func TestBuildSeverityOverrides_Nil(t *testing.T) {
	result := BuildSeverityOverrides(nil)
	if result != nil {
		t.Errorf("BuildSeverityOverrides(nil) = %v, want nil", result)
	}
}

func TestBuildSeverityOverrides_WithRules(t *testing.T) {
	rs := &RuleSet{
		Rules: map[string]RuleDefinition{
			"rule-error": {Severity: "error"},
			"rule-warn":  {Severity: "warn"},
			"rule-info":  {Severity: "info"},
			"rule-hint":  {Severity: "hint"},
			"rule-off":   {Severity: "off"},
		},
	}

	overrides := BuildSeverityOverrides(rs)

	// Build a lookup map for easier assertions.
	overrideMap := make(map[string]SeverityOverride)
	for _, o := range overrides {
		overrideMap[o.RuleID] = o
	}

	if len(overrides) != 5 {
		t.Fatalf("overrides count = %d, want 5", len(overrides))
	}

	tests := []struct {
		ruleID   string
		severity Severity
		disabled bool
	}{
		{"rule-error", SeverityError, false},
		{"rule-warn", SeverityWarning, false},
		{"rule-info", SeverityInfo, false},
		{"rule-hint", SeverityHint, false},
		{"rule-off", SeverityOff, true},
	}

	for _, tt := range tests {
		o, ok := overrideMap[tt.ruleID]
		if !ok {
			t.Errorf("override for %q not found", tt.ruleID)
			continue
		}
		if o.Severity != tt.severity {
			t.Errorf("override %q severity = %v, want %v", tt.ruleID, o.Severity, tt.severity)
		}
		if o.Disabled != tt.disabled {
			t.Errorf("override %q disabled = %v, want %v", tt.ruleID, o.Disabled, tt.disabled)
		}
	}
}

func TestBuildSeverityOverrides_InvalidSeverity(t *testing.T) {
	rs := &RuleSet{
		Rules: map[string]RuleDefinition{
			"rule-valid":   {Severity: "error"},
			"rule-invalid": {Severity: "not-a-severity"},
		},
	}

	overrides := BuildSeverityOverrides(rs)

	// Only the valid rule should produce an override.
	if len(overrides) != 1 {
		t.Fatalf("overrides count = %d, want 1 (invalid severity should be skipped)", len(overrides))
	}
	if overrides[0].RuleID != "rule-valid" {
		t.Errorf("override RuleID = %q, want %q", overrides[0].RuleID, "rule-valid")
	}
}

func TestBuildSeverityOverrides_EmptyRuleset(t *testing.T) {
	rs := &RuleSet{
		Rules: map[string]RuleDefinition{},
	}
	overrides := BuildSeverityOverrides(rs)
	if len(overrides) != 0 {
		t.Errorf("overrides count = %d, want 0", len(overrides))
	}
}
