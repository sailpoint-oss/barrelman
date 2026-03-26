package barrelman

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	reg := NewRegistry()
	if reg == nil {
		t.Fatal("NewRegistry returned nil")
	}
	if len(reg.All()) != 0 {
		t.Errorf("new registry should be empty, got %d", len(reg.All()))
	}
}

func TestRegistryRegisterAndGet(t *testing.T) {
	reg := NewRegistry()
	meta := RuleMeta{
		ID:          "test-rule",
		Description: "A test rule",
		Severity:    SeverityWarning,
		Category:    CategoryNaming,
		Recommended: true,
	}
	rule := Rule{
		ID:   "test-rule",
		Meta: meta,
		Run:  func(ctx *AnalysisContext) []Diagnostic { return nil },
	}
	reg.Register(rule)

	got, ok := reg.Get("test-rule")
	if !ok {
		t.Fatal("Get returned false for registered rule")
	}
	if got.ID != "test-rule" {
		t.Errorf("Get.ID = %q, want %q", got.ID, "test-rule")
	}

	_, ok = reg.Get("nonexistent")
	if ok {
		t.Error("Get returned true for nonexistent rule")
	}
}

func TestRegistryRegisterMeta(t *testing.T) {
	reg := NewRegistry()
	meta := RuleMeta{ID: "meta-only", Description: "metadata only"}
	reg.RegisterMeta(meta)

	got, ok := reg.Get("meta-only")
	if !ok {
		t.Fatal("RegisterMeta rule not found via Get")
	}
	if got.Description != "metadata only" {
		t.Errorf("Description = %q, want %q", got.Description, "metadata only")
	}

	// RegisterMeta should not add to AllRules.
	if len(reg.AllRules()) != 0 {
		t.Errorf("AllRules should be empty after RegisterMeta, got %d", len(reg.AllRules()))
	}
}

func TestRegistryAll(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Rule{ID: "a", Meta: RuleMeta{ID: "a"}, Run: func(ctx *AnalysisContext) []Diagnostic { return nil }})
	reg.Register(Rule{ID: "b", Meta: RuleMeta{ID: "b"}, Run: func(ctx *AnalysisContext) []Diagnostic { return nil }})

	all := reg.All()
	if len(all) != 2 {
		t.Errorf("All() returned %d, want 2", len(all))
	}
}

func TestRegistryAllRules(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Rule{ID: "a", Meta: RuleMeta{ID: "a"}, Run: func(ctx *AnalysisContext) []Diagnostic { return nil }})
	reg.RegisterMeta(RuleMeta{ID: "meta-only"})

	rules := reg.AllRules()
	if len(rules) != 1 {
		t.Errorf("AllRules() returned %d, want 1", len(rules))
	}
}

func TestRegistryByCategory(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Rule{ID: "a", Meta: RuleMeta{ID: "a", Category: CategoryNaming}, Run: func(ctx *AnalysisContext) []Diagnostic { return nil }})
	reg.Register(Rule{ID: "b", Meta: RuleMeta{ID: "b", Category: CategorySecurity}, Run: func(ctx *AnalysisContext) []Diagnostic { return nil }})
	reg.Register(Rule{ID: "c", Meta: RuleMeta{ID: "c", Category: CategoryNaming}, Run: func(ctx *AnalysisContext) []Diagnostic { return nil }})

	naming := reg.ByCategory(CategoryNaming)
	if len(naming) != 2 {
		t.Errorf("ByCategory(naming) returned %d, want 2", len(naming))
	}
	security := reg.ByCategory(CategorySecurity)
	if len(security) != 1 {
		t.Errorf("ByCategory(security) returned %d, want 1", len(security))
	}
}

func TestRegistryRecommended(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Rule{ID: "a", Meta: RuleMeta{ID: "a", Recommended: true}, Run: func(ctx *AnalysisContext) []Diagnostic { return nil }})
	reg.Register(Rule{ID: "b", Meta: RuleMeta{ID: "b", Recommended: false}, Run: func(ctx *AnalysisContext) []Diagnostic { return nil }})

	rec := reg.Recommended()
	if len(rec) != 1 {
		t.Errorf("Recommended() returned %d, want 1", len(rec))
	}
	if rec[0].ID != "a" {
		t.Errorf("Recommended()[0].ID = %q, want %q", rec[0].ID, "a")
	}
}

func TestRegistryIDs(t *testing.T) {
	reg := NewRegistry()
	reg.Register(Rule{ID: "alpha", Meta: RuleMeta{ID: "alpha"}, Run: func(ctx *AnalysisContext) []Diagnostic { return nil }})
	reg.Register(Rule{ID: "beta", Meta: RuleMeta{ID: "beta"}, Run: func(ctx *AnalysisContext) []Diagnostic { return nil }})

	ids := reg.IDs()
	if len(ids) != 2 {
		t.Errorf("IDs() returned %d, want 2", len(ids))
	}
}
