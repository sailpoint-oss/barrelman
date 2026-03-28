package validation

import (
	"path/filepath"
	"reflect"
	"testing"
)

func TestAdditionalValidator_MatchesFilePatternsWithFileURI(t *testing.T) {
	root := t.TempDir()
	v := NewAdditionalValidator()
	v.Configure(root, map[string]ValidationGroup{
		"github-actions": {
			Patterns: []string{".github/workflows/*.yaml"},
		},
		"typescript": {
			Patterns: []string{"**/tsconfig.*.json"},
		},
	})

	group, matched := v.MatchesFilePatterns("file://" + filepath.ToSlash(filepath.Join(root, ".github", "workflows", "ci.yaml")))
	if !matched || group != "github-actions" {
		t.Fatalf("expected github-actions match, got group=%q matched=%v", group, matched)
	}

	group, matched = v.MatchesFilePatterns("file://" + filepath.ToSlash(filepath.Join(root, "apps", "web", "tsconfig.app.json")))
	if !matched || group != "typescript" {
		t.Fatalf("expected typescript match, got group=%q matched=%v", group, matched)
	}

	if group, matched = v.MatchesFilePatterns("file://" + filepath.ToSlash(filepath.Join(root, "docs", "openapi.yaml"))); matched || group != "" {
		t.Fatalf("expected no match, got group=%q matched=%v", group, matched)
	}
}

func TestAdditionalValidator_GroupsAndSchemaDir(t *testing.T) {
	root := t.TempDir()
	groups := map[string]ValidationGroup{
		"ci": {
			Patterns: []string{".github/workflows/*.yaml"},
			Schemas:  []SchemaPatternMapping{{Schema: "github-actions.json"}},
		},
	}
	v := NewAdditionalValidator()
	v.Configure(root, groups)

	if got := v.SchemaDir(); got != filepath.Join(root, ".telescope", "schemas") {
		t.Fatalf("SchemaDir = %q", got)
	}
	if got := v.Groups(); !reflect.DeepEqual(got, groups) {
		t.Fatalf("Groups mismatch: got %+v want %+v", got, groups)
	}
}

func TestEnrichAdditionalDiagnostics_AddsContextAndFallback(t *testing.T) {
	got := EnrichAdditionalDiagnostics([]string{"missing field", ""}, "ci", "github-actions.json")
	want := []string{
		"[schema:github-actions.json group:ci] missing field",
		"[schema:github-actions.json group:ci] Schema validation failed",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("EnrichAdditionalDiagnostics mismatch:\n got: %+v\nwant: %+v", got, want)
	}
}
