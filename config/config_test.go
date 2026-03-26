package config

import (
	"testing"
	"time"

	"github.com/sailpoint-oss/barrelman/rulesets"
)

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg == nil {
		t.Fatal("DefaultConfig returned nil")
	}
	if cfg.Extends != rulesets.Recommended {
		t.Errorf("Extends = %q, want %q", cfg.Extends, rulesets.Recommended)
	}
	if cfg.Output.Format != "text" {
		t.Errorf("Output.Format = %q, want %q", cfg.Output.Format, "text")
	}
	if cfg.LSP.Debounce != 300*time.Millisecond {
		t.Errorf("LSP.Debounce = %v, want 300ms", cfg.LSP.Debounce)
	}
	if cfg.LSP.MaxFileSize != 5*1024*1024 {
		t.Errorf("LSP.MaxFileSize = %d, want 5MB", cfg.LSP.MaxFileSize)
	}
}

func TestEffectiveSchemaValidationMode(t *testing.T) {
	tests := []struct {
		mode string
		want string
	}{
		{"go", "go"},
		{"bun", "go"},
		{"compare", "go"},
		{"", "go"},
		{"anything", "go"},
	}
	for _, tc := range tests {
		cfg := &Config{LSP: LSPConfig{SchemaValidation: LSPSchemaValidationSettings{Mode: tc.mode}}}
		got := cfg.EffectiveSchemaValidationMode()
		if got != tc.want {
			t.Errorf("EffectiveSchemaValidationMode(%q) = %q, want %q", tc.mode, got, tc.want)
		}
	}
}

func TestHasCustomRules(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.HasCustomRules() {
		t.Error("default config should not have custom rules")
	}

	cfg.OpenAPI.Rules = []RuleRef{{Rule: "test.ts"}}
	if !cfg.HasCustomRules() {
		t.Error("config with OpenAPI rules should have custom rules")
	}
}

func TestHasSpectralRulesets(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.HasSpectralRulesets() {
		t.Error("default config should not have spectral rulesets")
	}

	cfg.SpectralRulesets = []string{".spectral.yaml"}
	if !cfg.HasSpectralRulesets() {
		t.Error("config with spectral rulesets should report true")
	}
}

func TestNeedsBunSidecar(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.NeedsBunSidecar() {
		t.Error("default config should not need bun sidecar")
	}

	cfg.SpectralRulesets = []string{".spectral.yaml"}
	if !cfg.NeedsBunSidecar() {
		t.Error("config with spectral rulesets needs bun sidecar")
	}
}

func TestResolveRunner(t *testing.T) {
	tests := []struct {
		ref  RuleRef
		want string
	}{
		{RuleRef{Rule: "check.ts", Runner: ""}, "bun"},
		{RuleRef{Rule: "check.js", Runner: "auto"}, "bun"},
		{RuleRef{Rule: "check.mjs", Runner: ""}, "bun"},
		{RuleRef{Rule: "check.ts", Runner: "bun"}, "bun"},
		{RuleRef{Rule: "check.go", Runner: ""}, "native"},
		{RuleRef{Rule: "check.yaml", Runner: ""}, "native"},
		{RuleRef{Rule: "check.ts", Runner: "custom"}, "custom"},
	}
	for _, tc := range tests {
		got := ResolveRunner(tc.ref)
		if got != tc.want {
			t.Errorf("ResolveRunner(%+v) = %q, want %q", tc.ref, got, tc.want)
		}
	}
}
