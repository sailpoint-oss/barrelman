package barrelman

import (
	"testing"

	navigator "github.com/sailpoint-oss/navigator"
)

func TestDefineAndBuild(t *testing.T) {
	meta := RuleMeta{
		ID:          "test-build",
		Description: "test builder",
		Severity:    SeverityWarning,
		Category:    CategoryDocumentation,
	}

	called := false
	rule := Define("test-build", meta).Operations(
		func(path, method string, op *navigator.Operation, r *Reporter) {
			called = true
			r.At(op.Loc, "test issue")
		},
	).Build()

	if rule.ID != "test-build" {
		t.Errorf("rule.ID = %q, want %q", rule.ID, "test-build")
	}
	if rule.Meta.Description != "test builder" {
		t.Errorf("rule.Meta.Description = %q", rule.Meta.Description)
	}

	// Run with nil index should return nil.
	diags := rule.Run(&AnalysisContext{})
	if len(diags) != 0 {
		t.Errorf("expected no diags for nil index, got %d", len(diags))
	}
	if called {
		t.Error("visitor should not be called when index is nil")
	}
}

func TestDefineAndRegister(t *testing.T) {
	reg := NewRegistry()
	meta := RuleMeta{
		ID:       "test-register",
		Severity: SeverityError,
	}
	rule := Define("test-register", meta).Custom(
		func(idx *navigator.Index, r *Reporter) {
			r.AtRange(FileStartRange, "custom issue")
		},
	).Register(reg)

	if rule.ID != "test-register" {
		t.Errorf("rule.ID = %q", rule.ID)
	}

	_, ok := reg.Get("test-register")
	if !ok {
		t.Error("rule not found in registry after Register")
	}

	rules := reg.AllRules()
	if len(rules) != 1 {
		t.Errorf("AllRules() = %d, want 1", len(rules))
	}
}

func TestBuilderMeta(t *testing.T) {
	meta := RuleMeta{ID: "meta-test", Description: "desc"}
	b := Define("meta-test", meta)
	got := b.Meta()
	if got.ID != "meta-test" || got.Description != "desc" {
		t.Errorf("Meta() = %+v, unexpected", got)
	}
}
