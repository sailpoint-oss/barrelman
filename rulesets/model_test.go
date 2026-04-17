package rulesets

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input   string
		wantSev Severity
		wantOK  bool
	}{
		{"error", SeverityError, true},
		{"warn", SeverityWarning, true},
		{"warning", SeverityWarning, true},
		{"info", SeverityInfo, true},
		{"information", SeverityInfo, true},
		{"hint", SeverityHint, true},
		{"off", SeverityOff, true},
		{"false", SeverityOff, true},
		{"", SeverityOff, false},
		{"invalid", SeverityOff, false},
		{"ERROR", SeverityOff, false},   // case-sensitive
		{"Warning", SeverityOff, false}, // case-sensitive
		{"unknown", SeverityOff, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			gotSev, gotOK := ParseSeverity(tt.input)
			if gotOK != tt.wantOK {
				t.Errorf("ParseSeverity(%q) ok = %v, want %v", tt.input, gotOK, tt.wantOK)
			}
			if gotSev != tt.wantSev {
				t.Errorf("ParseSeverity(%q) severity = %v, want %v", tt.input, gotSev, tt.wantSev)
			}
		})
	}
}

func TestRuleDefinition_UnmarshalYAML_StringSeverity(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantSev string
	}{
		{"error", `"error"`, "error"},
		{"warn", `"warn"`, "warn"},
		{"info", `"info"`, "info"},
		{"hint", `"hint"`, "hint"},
		{"off", `"off"`, "off"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rd RuleDefinition
			if err := yaml.Unmarshal([]byte(tt.yaml), &rd); err != nil {
				t.Fatalf("UnmarshalYAML(%q) error: %v", tt.yaml, err)
			}
			if rd.Severity != tt.wantSev {
				t.Errorf("Severity = %q, want %q", rd.Severity, tt.wantSev)
			}
		})
	}
}

func TestRuleDefinition_UnmarshalYAML_BoolFalse(t *testing.T) {
	var rd RuleDefinition
	if err := yaml.Unmarshal([]byte(`false`), &rd); err != nil {
		t.Fatalf("UnmarshalYAML(false) error: %v", err)
	}
	if rd.Severity != "off" {
		t.Errorf("Severity = %q, want %q", rd.Severity, "off")
	}
}

func TestRuleDefinition_UnmarshalYAML_IntegerSeverity(t *testing.T) {
	tests := []struct {
		input   string
		wantSev string
	}{
		{"0", "off"},
		{"1", "error"},
		{"2", "warn"},
		{"3", "info"},
		{"4", "hint"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			var rd RuleDefinition
			if err := yaml.Unmarshal([]byte(tt.input), &rd); err != nil {
				t.Fatalf("UnmarshalYAML(%q) error: %v", tt.input, err)
			}
			if rd.Severity != tt.wantSev {
				t.Errorf("Severity = %q, want %q", rd.Severity, tt.wantSev)
			}
		})
	}
}

func TestRuleDefinition_UnmarshalYAML_SequenceForm(t *testing.T) {
	// [severity, {options}] form
	input := `["error", {"description": "custom desc"}]`
	var rd RuleDefinition
	if err := yaml.Unmarshal([]byte(input), &rd); err != nil {
		t.Fatalf("UnmarshalYAML(%q) error: %v", input, err)
	}
	if rd.Severity != "error" {
		t.Errorf("Severity = %q, want %q", rd.Severity, "error")
	}
	if rd.Description != "custom desc" {
		t.Errorf("Description = %q, want %q", rd.Description, "custom desc")
	}
}

func TestRuleDefinition_UnmarshalYAML_SequenceSeverityOnly(t *testing.T) {
	input := `["warn"]`
	var rd RuleDefinition
	if err := yaml.Unmarshal([]byte(input), &rd); err != nil {
		t.Fatalf("UnmarshalYAML(%q) error: %v", input, err)
	}
	if rd.Severity != "warn" {
		t.Errorf("Severity = %q, want %q", rd.Severity, "warn")
	}
}

func TestRuleDefinition_UnmarshalYAML_EmptySequence(t *testing.T) {
	input := `[]`
	var rd RuleDefinition
	if err := yaml.Unmarshal([]byte(input), &rd); err != nil {
		t.Fatalf("UnmarshalYAML(%q) error: %v", input, err)
	}
	if rd.Severity != "" {
		t.Errorf("Severity = %q, want empty string", rd.Severity)
	}
}

func TestRuleDefinition_UnmarshalYAML_MapForm(t *testing.T) {
	input := `
severity: error
description: "A full rule definition"
message: "something is wrong"
`
	var rd RuleDefinition
	if err := yaml.Unmarshal([]byte(input), &rd); err != nil {
		t.Fatalf("UnmarshalYAML error: %v", err)
	}
	if rd.Severity != "error" {
		t.Errorf("Severity = %q, want %q", rd.Severity, "error")
	}
	if rd.Description != "A full rule definition" {
		t.Errorf("Description = %q, want %q", rd.Description, "A full rule definition")
	}
	if rd.Message != "something is wrong" {
		t.Errorf("Message = %q, want %q", rd.Message, "something is wrong")
	}
}

func TestRuleDefinition_UnmarshalYAML_MapWithFormats(t *testing.T) {
	input := `
severity: warn
formats:
  - oas3
  - oas3.1
`
	var rd RuleDefinition
	if err := yaml.Unmarshal([]byte(input), &rd); err != nil {
		t.Fatalf("UnmarshalYAML error: %v", err)
	}
	if rd.Severity != "warn" {
		t.Errorf("Severity = %q, want %q", rd.Severity, "warn")
	}
	if len(rd.Formats) != 2 {
		t.Fatalf("Formats length = %d, want 2", len(rd.Formats))
	}
	if rd.Formats[0] != "oas3" {
		t.Errorf("Formats[0] = %q, want %q", rd.Formats[0], "oas3")
	}
	if rd.Formats[1] != "oas3.1" {
		t.Errorf("Formats[1] = %q, want %q", rd.Formats[1], "oas3.1")
	}
}

func TestRuleDefinition_UnmarshalYAML_SequenceWithSeverityOverride(t *testing.T) {
	// The sequence form should preserve the severity from the first element
	// even when the second element map contains a severity field.
	input := `["error", {"severity": "warn", "description": "overridden"}]`
	var rd RuleDefinition
	if err := yaml.Unmarshal([]byte(input), &rd); err != nil {
		t.Fatalf("UnmarshalYAML error: %v", err)
	}
	// Severity from first element takes precedence.
	if rd.Severity != "error" {
		t.Errorf("Severity = %q, want %q (first element should take precedence)", rd.Severity, "error")
	}
	if rd.Description != "overridden" {
		t.Errorf("Description = %q, want %q", rd.Description, "overridden")
	}
}

func TestRuleDefinition_UnmarshalYAML_InRulesMap(t *testing.T) {
	// Test unmarshalling within a full rules map context, as it would appear
	// in a ruleset YAML file.
	input := `
rules:
  info-contact: error
  info-description: false
  sailpoint-operation-single-tag:
    severity: warn
    description: "Provide a tag for every operation"
  custom-rule: ["hint", {"message": "check this"}]
`
	var rs RuleSet
	if err := yaml.Unmarshal([]byte(input), &rs); err != nil {
		t.Fatalf("UnmarshalYAML error: %v", err)
	}

	if def, ok := rs.Rules["info-contact"]; !ok {
		t.Error("info-contact not found in rules")
	} else if def.Severity != "error" {
		t.Errorf("info-contact severity = %q, want %q", def.Severity, "error")
	}

	if def, ok := rs.Rules["info-description"]; !ok {
		t.Error("info-description not found in rules")
	} else if def.Severity != "off" {
		t.Errorf("info-description severity = %q, want %q", def.Severity, "off")
	}

	if def, ok := rs.Rules["sailpoint-operation-single-tag"]; !ok {
		t.Error("sailpoint-operation-single-tag not found in rules")
	} else {
		if def.Severity != "warn" {
			t.Errorf("sailpoint-operation-single-tag severity = %q, want %q", def.Severity, "warn")
		}
		if def.Description != "Provide a tag for every operation" {
			t.Errorf("sailpoint-operation-single-tag description = %q, want %q", def.Description, "Provide a tag for every operation")
		}
	}

	if def, ok := rs.Rules["custom-rule"]; !ok {
		t.Error("custom-rule not found in rules")
	} else {
		if def.Severity != "hint" {
			t.Errorf("custom-rule severity = %q, want %q", def.Severity, "hint")
		}
		if def.Message != "check this" {
			t.Errorf("custom-rule message = %q, want %q", def.Message, "check this")
		}
	}
}

func TestIntToSeverity(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"0", "off"},
		{"1", "error"},
		{"2", "warn"},
		{"3", "info"},
		{"4", "hint"},
		{"5", "5"}, // unknown int passes through
		{"99", "99"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := intToSeverity(tt.input)
			if got != tt.want {
				t.Errorf("intToSeverity(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSeverityOverrideStruct(t *testing.T) {
	so := SeverityOverride{
		RuleID:   "test-rule",
		Severity: SeverityError,
		Disabled: false,
	}
	if so.RuleID != "test-rule" {
		t.Errorf("RuleID = %q, want %q", so.RuleID, "test-rule")
	}
	if so.Severity != SeverityError {
		t.Errorf("Severity = %v, want %v", so.Severity, SeverityError)
	}
	if so.Disabled {
		t.Error("Disabled = true, want false")
	}
}
