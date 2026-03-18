// Package btesting provides test helpers for validating barrelman rules.
//
// Usage:
//
//	func TestMyRule(t *testing.T) {
//	    btesting.Run(t, myRule, btesting.Case{
//	        Name: "catches missing field",
//	        Spec: `openapi: "3.1.0"
//	info:
//	  title: Test
//	  version: "1.0"`,
//	        Expect: []btesting.Diag{
//	            {Line: 1, Code: "my-rule-id", Severity: btesting.Warn},
//	        },
//	    })
//	}
package btesting

import (
	"strings"
	"testing"

	navigator "github.com/LukasParke/navigator"
	tree_sitter "github.com/tree-sitter/go-tree-sitter"

	"github.com/sailpoint-oss/barrelman"
)

// Severity constants for test expectations.
const (
	Error = barrelman.SeverityError
	Warn  = barrelman.SeverityWarning
	Info  = barrelman.SeverityInfo
	Hint  = barrelman.SeverityHint
)

// Diag describes an expected diagnostic.
type Diag struct {
	Line     uint32             // 0-based line
	Col      uint32             // 0-based character (optional, 0 means any)
	Code     string             // rule ID
	Severity barrelman.Severity // expected severity
	Message  string             // optional substring match
}

// Case is a single test scenario for a rule.
type Case struct {
	Name   string
	Spec   string // YAML content
	Expect []Diag // expected diagnostics
}

// Run executes a rule against the given test cases.
func Run(t *testing.T, rule barrelman.Rule, cases ...Case) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()
			ctx := buildTestContext(t, tc.Spec)
			diags := rule.Run(ctx)
			assertDiagnostics(t, diags, tc.Expect)
		})
	}
}

// RunVisitors executes a rule's visitors directly against test cases.
func RunVisitors(t *testing.T, ruleID string, severity barrelman.Severity, v barrelman.Visitors, cases ...Case) {
	t.Helper()

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Helper()
			ctx := buildTestContext(t, tc.Spec)
			r := barrelman.NewReporter(ruleID, severity)
			barrelman.Walk(ctx.Index, v, r)
			assertDiagnostics(t, r.Diagnostics(), tc.Expect)
		})
	}
}

func buildTestContext(t *testing.T, spec string) *barrelman.AnalysisContext {
	t.Helper()

	content := []byte(spec)
	lang := navigator.YAMLLanguage()
	parser := tree_sitter.NewParser()
	defer parser.Close()
	if err := parser.SetLanguage(lang); err != nil {
		t.Fatalf("set language: %v", err)
	}

	tree := parser.Parse(content, nil)
	if tree == nil {
		t.Fatal("tree-sitter produced nil tree for spec")
	}
	t.Cleanup(func() { tree.Close() })

	uri := "file:///test/spec.yaml"
	idx := navigator.ParseTree(tree, content, uri, navigator.FormatYAML)
	if idx == nil {
		t.Fatal("navigator.ParseTree returned nil index")
	}

	return &barrelman.AnalysisContext{
		Index:    idx,
		Tree:     tree,
		Language: lang,
		Content:  content,
		URI:      uri,
	}
}

func assertDiagnostics(t *testing.T, actual []barrelman.Diagnostic, expected []Diag) {
	t.Helper()

	if len(expected) == 0 {
		if len(actual) > 0 {
			t.Errorf("expected no diagnostics, got %d:", len(actual))
			for _, d := range actual {
				t.Errorf("  L%d:%d [%s] %s",
					d.Range.Start.Line, d.Range.Start.Character,
					d.Code, d.Message)
			}
		}
		return
	}

	for i, exp := range expected {
		if i >= len(actual) {
			t.Errorf("missing expected diagnostic #%d: L%d code=%s", i, exp.Line, exp.Code)
			continue
		}
		d := actual[i]

		if exp.Line != d.Range.Start.Line {
			t.Errorf("diagnostic #%d line: got %d, want %d (code=%s msg=%q)",
				i, d.Range.Start.Line, exp.Line, d.Code, d.Message)
		}
		if exp.Col > 0 && exp.Col != d.Range.Start.Character {
			t.Errorf("diagnostic #%d col: got %d, want %d",
				i, d.Range.Start.Character, exp.Col)
		}
		if exp.Code != "" && exp.Code != d.Code {
			t.Errorf("diagnostic #%d code: got %q, want %q",
				i, d.Code, exp.Code)
		}
		if exp.Severity != 0 && exp.Severity != d.Severity {
			t.Errorf("diagnostic #%d severity: got %d, want %d",
				i, d.Severity, exp.Severity)
		}
		if exp.Message != "" && !strings.Contains(d.Message, exp.Message) {
			t.Errorf("diagnostic #%d message: %q does not contain %q",
				i, d.Message, exp.Message)
		}
	}

	if len(actual) > len(expected) {
		t.Errorf("got %d extra diagnostic(s):", len(actual)-len(expected))
		for i := len(expected); i < len(actual); i++ {
			d := actual[i]
			t.Errorf("  L%d:%d [%s] %s",
				d.Range.Start.Line, d.Range.Start.Character,
				d.Code, d.Message)
		}
	}
}
