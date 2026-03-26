package rulesets

import (
	"testing"
)

func TestSpectralToTelescopeID_MappedRules(t *testing.T) {
	tests := []struct {
		spectralID   string
		telescopeID  string
	}{
		{"info-contact", "info-contact"},
		{"info-description", "info-description"},
		{"info-license", "info-license"},
		{"oas3-unused-component", "unused-component"},
		{"no-eval-in-markdown", "description-markdown"},
		{"no-script-tags-in-markdown", "description-html"},
		{"operation-operationId", "operation-operationId"},
		{"operation-operationId-unique", "operation-operationId-unique"},
		{"oas3-schema", "oas3-schema"},
		{"oas3-valid-media-example", "oas3-valid-media-example"},
		{"oas3-valid-schema-example", "oas3-valid-schema-example"},
		{"contact-properties", "contact-properties"},
		{"license-url", "license-url"},
	}

	for _, tt := range tests {
		t.Run(tt.spectralID, func(t *testing.T) {
			got := SpectralToTelescopeID(tt.spectralID)
			if got != tt.telescopeID {
				t.Errorf("SpectralToTelescopeID(%q) = %q, want %q", tt.spectralID, got, tt.telescopeID)
			}
		})
	}
}

func TestSpectralToTelescopeID_Passthrough(t *testing.T) {
	got := SpectralToTelescopeID("unknown-rule")
	if got != "unknown-rule" {
		t.Errorf("SpectralToTelescopeID(%q) = %q, want %q", "unknown-rule", got, "unknown-rule")
	}
}

func TestTelescopeToSpectralID_MappedRules(t *testing.T) {
	tests := []struct {
		telescopeID string
		spectralID  string
	}{
		{"unused-component", "oas3-unused-component"},
		{"description-markdown", "no-eval-in-markdown"},
		{"description-html", "no-script-tags-in-markdown"},
		{"info-contact", "info-contact"},
		{"oas3-schema", "oas3-schema"},
		{"oas3-api-servers", "oas3-api-servers"},
	}

	for _, tt := range tests {
		t.Run(tt.telescopeID, func(t *testing.T) {
			got := TelescopeToSpectralID(tt.telescopeID)
			if got != tt.spectralID {
				t.Errorf("TelescopeToSpectralID(%q) = %q, want %q", tt.telescopeID, got, tt.spectralID)
			}
		})
	}
}

func TestTelescopeToSpectralID_Passthrough(t *testing.T) {
	got := TelescopeToSpectralID("completely-unknown")
	if got != "completely-unknown" {
		t.Errorf("TelescopeToSpectralID(%q) = %q, want %q", "completely-unknown", got, "completely-unknown")
	}
}

func TestIsNativeRule(t *testing.T) {
	nativeRules := []string{
		"info-contact",
		"info-description",
		"info-license",
		"operation-description",
		"operation-operationId",
		"operation-operationId-unique",
		"operation-tags",
		"path-keys-no-trailing-slash",
		"path-declarations-must-exist",
		"path-params",
		"no-eval-in-markdown",
		"no-script-tags-in-markdown",
		"oas3-api-servers",
		"oas3-schema",
		"tag-description",
		"parameter-description",
		"oas3-unused-component",
		"contact-properties",
		"license-url",
		"oas3-valid-media-example",
		"oas3-valid-schema-example",
	}

	for _, id := range nativeRules {
		t.Run(id+"_native", func(t *testing.T) {
			if !IsNativeRule(id) {
				t.Errorf("IsNativeRule(%q) = false, want true", id)
			}
		})
	}

	nonNativeRules := []string{
		"unknown",
		"some-custom-rule",
		"duplicated-entry-in-enum",
		"typed-enum",
	}

	for _, id := range nonNativeRules {
		t.Run(id+"_not_native", func(t *testing.T) {
			if IsNativeRule(id) {
				t.Errorf("IsNativeRule(%q) = true, want false", id)
			}
		})
	}
}

func TestSpectralToTelescopeID_RoundTrip(t *testing.T) {
	// For every mapped Spectral rule, converting to Telescope and back should
	// yield the original Spectral ID.
	for spectralID, telescopeID := range spectralToTelescope {
		got := TelescopeToSpectralID(telescopeID)
		if got != spectralID {
			t.Errorf("round-trip failed: Spectral %q -> Telescope %q -> Spectral %q (want %q)",
				spectralID, telescopeID, got, spectralID)
		}
	}
}

func TestGetSpectralBuiltin_OAS(t *testing.T) {
	rs := GetSpectralBuiltin("spectral:oas")
	if rs == nil {
		t.Fatal("expected non-nil ruleset for spectral:oas")
	}
	if rs.Name != "Spectral OAS" {
		t.Errorf("Name = %q, want %q", rs.Name, "Spectral OAS")
	}
	if len(rs.Rules) == 0 {
		t.Fatal("expected spectral:oas ruleset to contain rules")
	}

	// Verify all spectralOASDefaults are present.
	for ruleID, expectedSev := range spectralOASDefaults {
		def, ok := rs.Rules[ruleID]
		if !ok {
			t.Errorf("rule %q missing from spectral:oas ruleset", ruleID)
			continue
		}
		if def.Severity != expectedSev {
			t.Errorf("rule %q severity = %q, want %q", ruleID, def.Severity, expectedSev)
		}
	}

	// Verify ruleset has exactly the expected number of rules.
	if len(rs.Rules) != len(spectralOASDefaults) {
		t.Errorf("spectral:oas has %d rules, want %d", len(rs.Rules), len(spectralOASDefaults))
	}
}

func TestGetSpectralBuiltin_Unknown(t *testing.T) {
	rs := GetSpectralBuiltin("other")
	if rs != nil {
		t.Errorf("expected nil for unknown spectral builtin, got %+v", rs)
	}

	rs = GetSpectralBuiltin("spectral:unknown")
	if rs != nil {
		t.Errorf("expected nil for spectral:unknown, got %+v", rs)
	}

	rs = GetSpectralBuiltin("")
	if rs != nil {
		t.Errorf("expected nil for empty string, got %+v", rs)
	}
}

func TestSpectralOASConstant(t *testing.T) {
	if SpectralOAS != "spectral:oas" {
		t.Errorf("SpectralOAS = %q, want %q", SpectralOAS, "spectral:oas")
	}
}
