package spectral

import (
	"log/slog"
	"testing"

	"github.com/sailpoint-oss/barrelman"
	"gopkg.in/yaml.v3"
)

func TestNewEngine_ReturnsNonNil(t *testing.T) {
	e := NewEngine(nil, nil)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
}

func TestNewEngine_WithLogger(t *testing.T) {
	logger := slog.Default()
	e := NewEngine(nil, logger)
	if e == nil {
		t.Fatal("expected non-nil engine")
	}
	if e.logger != logger {
		t.Error("expected logger to be stored")
	}
}

func TestSetRules_RulesRoundTrip(t *testing.T) {
	e := NewEngine(nil, nil)

	rules := []Rule{
		{ID: "rule-1", Description: "first rule"},
		{ID: "rule-2", Description: "second rule"},
	}
	e.SetRules(rules)

	got := e.Rules()
	if len(got) != 2 {
		t.Fatalf("expected 2 rules, got %d", len(got))
	}
	if got[0].ID != "rule-1" {
		t.Errorf("expected rule-1, got %s", got[0].ID)
	}
	if got[1].ID != "rule-2" {
		t.Errorf("expected rule-2, got %s", got[1].ID)
	}
}

func TestRules_ReturnsCopy(t *testing.T) {
	e := NewEngine([]Rule{{ID: "original"}}, nil)
	got := e.Rules()
	got[0].ID = "mutated"

	original := e.Rules()
	if original[0].ID != "original" {
		t.Error("Rules() should return a copy; mutation should not affect engine")
	}
}

func TestExecute_WithNoRules(t *testing.T) {
	e := NewEngine(nil, nil)
	diags := e.Execute([]byte("openapi: 3.0.0"))
	if diags != nil {
		t.Fatalf("expected nil diagnostics with no rules, got %d", len(diags))
	}
}

func TestExecute_WithInvalidYAML(t *testing.T) {
	e := NewEngine([]Rule{
		{
			ID:    "test-rule",
			Given: []string{"$"},
			Then:  []FunctionCall{{Function: "truthy"}},
		},
	}, nil)

	diags := e.Execute([]byte(":\n\t:\n\t\t- [[["))
	if diags != nil {
		t.Fatalf("expected nil diagnostics for invalid YAML, got %d", len(diags))
	}
}

func TestExecute_SimpleTruthyRule(t *testing.T) {
	doc := `
openapi: "3.0.0"
info:
  title: My API
  version: "1.0"
`
	rules := []Rule{
		{
			ID:       "info-title",
			Given:    []string{"$.info"},
			Severity: barrelman.SeverityWarning,
			Then: []FunctionCall{
				{Field: "title", Function: "truthy"},
			},
		},
	}

	e := NewEngine(rules, nil)
	diags := e.Execute([]byte(doc))
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for valid doc, got %d: %v", len(diags), diags)
	}
}

func TestExecute_TruthyRuleFindsIssue(t *testing.T) {
	doc := `
openapi: "3.0.0"
info:
  version: "1.0"
`
	rules := []Rule{
		{
			ID:       "info-title",
			Given:    []string{"$.info"},
			Severity: barrelman.SeverityWarning,
			Then: []FunctionCall{
				{Field: "title", Function: "truthy"},
			},
		},
	}

	e := NewEngine(rules, nil)
	diags := e.Execute([]byte(doc))
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if diags[0].Code != "info-title" {
		t.Errorf("expected code 'info-title', got %q", diags[0].Code)
	}
	if diags[0].Source != "spectral" {
		t.Errorf("expected source 'spectral', got %q", diags[0].Source)
	}
	if diags[0].Severity != barrelman.SeverityWarning {
		t.Errorf("expected severity warning, got %d", diags[0].Severity)
	}
}

func TestExecute_CustomMessage(t *testing.T) {
	doc := `
info:
  version: "1.0"
`
	rules := []Rule{
		{
			ID:       "info-title",
			Message:  "Title is required: {{error}}",
			Given:    []string{"$.info"},
			Severity: barrelman.SeverityError,
			Then: []FunctionCall{
				{Field: "title", Function: "truthy"},
			},
		},
	}

	e := NewEngine(rules, nil)
	diags := e.Execute([]byte(doc))
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	expected := "Title is required: 'title' must be truthy"
	if diags[0].Message != expected {
		t.Errorf("expected message %q, got %q", expected, diags[0].Message)
	}
}

func TestExecute_UnknownFunction(t *testing.T) {
	doc := `
info:
  title: Test
`
	logger := slog.Default()
	rules := []Rule{
		{
			ID:    "unknown-fn",
			Given: []string{"$.info"},
			Then: []FunctionCall{
				{Function: "nonexistent_function"},
			},
		},
	}

	e := NewEngine(rules, logger)
	diags := e.Execute([]byte(doc))
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for unknown function, got %d", len(diags))
	}
}

func TestExecute_MultipleGivenPaths(t *testing.T) {
	doc := `
info:
  title: ""
paths:
  /users: {}
`
	rules := []Rule{
		{
			ID:    "multi-given",
			Given: []string{"$.info", "$.paths"},
			Then: []FunctionCall{
				{Function: "truthy"},
			},
		},
	}

	e := NewEngine(rules, nil)
	diags := e.Execute([]byte(doc))
	// Both info and paths are non-empty mappings, so truthy should pass for both.
	// However info has content so it's truthy, and paths has content so it's truthy.
	if len(diags) != 0 {
		t.Fatalf("expected 0 diagnostics, got %d: %v", len(diags), diags)
	}
}

// ---------------------------------------------------------------------------
// nodeToRange
// ---------------------------------------------------------------------------

func TestNodeToRange_NilNode(t *testing.T) {
	r := nodeToRange(nil)
	if r.Start.Line != 0 || r.Start.Character != 0 {
		t.Errorf("expected zero range for nil node, got %+v", r)
	}
}

func TestNodeToRange_ScalarNode(t *testing.T) {
	node := &yaml.Node{
		Kind:   yaml.ScalarNode,
		Value:  "hello",
		Line:   3,
		Column: 5,
	}
	r := nodeToRange(node)
	// Line and Column are 1-based in yaml.Node, 0-based in our range.
	if r.Start.Line != 2 {
		t.Errorf("expected start line 2, got %d", r.Start.Line)
	}
	if r.Start.Character != 4 {
		t.Errorf("expected start char 4, got %d", r.Start.Character)
	}
	// "hello" is 5 UTF-16 code units
	if r.End.Character != 9 {
		t.Errorf("expected end char 9 (4+5), got %d", r.End.Character)
	}
}

func TestNodeToRange_MappingNode(t *testing.T) {
	node := &yaml.Node{
		Kind:   yaml.MappingNode,
		Line:   1,
		Column: 1,
	}
	r := nodeToRange(node)
	// For non-scalar, end should equal start
	if r.End.Character != r.Start.Character {
		t.Errorf("expected end == start for mapping node, got start=%d end=%d",
			r.Start.Character, r.End.Character)
	}
}

func TestNodeToRange_ZeroLineColumn(t *testing.T) {
	node := &yaml.Node{
		Kind:   yaml.ScalarNode,
		Value:  "x",
		Line:   0,
		Column: 0,
	}
	r := nodeToRange(node)
	// Should clamp to 0
	if r.Start.Line != 0 || r.Start.Character != 0 {
		t.Errorf("expected clamped to 0, got line=%d char=%d", r.Start.Line, r.Start.Character)
	}
}

// ---------------------------------------------------------------------------
// spectralUTF16Len
// ---------------------------------------------------------------------------

func TestSpectralUTF16Len_ASCII(t *testing.T) {
	if got := spectralUTF16Len("hello"); got != 5 {
		t.Errorf("expected 5, got %d", got)
	}
}

func TestSpectralUTF16Len_Empty(t *testing.T) {
	if got := spectralUTF16Len(""); got != 0 {
		t.Errorf("expected 0, got %d", got)
	}
}

func TestSpectralUTF16Len_Emoji(t *testing.T) {
	// U+1F600 (😀) requires a surrogate pair in UTF-16, so length = 2
	if got := spectralUTF16Len("😀"); got != 2 {
		t.Errorf("expected 2 for surrogate-pair emoji, got %d", got)
	}
}

func TestSpectralUTF16Len_Mixed(t *testing.T) {
	// "a😀b" = 1 + 2 + 1 = 4
	if got := spectralUTF16Len("a😀b"); got != 4 {
		t.Errorf("expected 4, got %d", got)
	}
}

func TestSpectralUTF16Len_Multibyte(t *testing.T) {
	// "café" = c(1) + a(1) + f(1) + é(1) = 4; é is U+00E9, fits in BMP
	if got := spectralUTF16Len("café"); got != 4 {
		t.Errorf("expected 4, got %d", got)
	}
}
