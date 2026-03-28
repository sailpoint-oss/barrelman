package barrelman_test

import (
	"path/filepath"
	"testing"

	"github.com/sailpoint-oss/barrelman"
)

func lintFixtureFile(t *testing.T, rel string) []barrelman.Diagnostic {
	t.Helper()
	path, err := filepath.Abs(filepath.Join("testdata", "toolchain", rel))
	if err != nil {
		t.Fatalf("abs path for %q: %v", rel, err)
	}
	results, err := barrelman.LintFiles([]string{path}, barrelman.LintOptions{Rules: allRules()})
	if err != nil {
		t.Fatalf("LintFiles(%q): %v", rel, err)
	}
	if len(results) != 1 {
		t.Fatalf("LintFiles(%q) returned %d results, want 1", rel, len(results))
	}
	return results[0].Diagnostics
}

func hasBarrelCode(diags []barrelman.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func hasBarrelIssueCode(diags []barrelman.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code != "oas3-schema" {
			continue
		}
		switch data := d.Data.(type) {
		case map[string]string:
			if data["issueCode"] == code {
				return true
			}
		case map[string]any:
			if v, ok := data["issueCode"].(string); ok && v == code {
				return true
			}
		}
	}
	return false
}

func hasBarrelDocumentKind(diags []barrelman.Diagnostic, kind string) bool {
	for _, d := range diags {
		if d.Code != "oas3-schema" {
			continue
		}
		switch data := d.Data.(type) {
		case map[string]string:
			if data["documentKind"] == kind {
				return true
			}
		case map[string]any:
			if v, ok := data["documentKind"].(string); ok && v == kind {
				return true
			}
		}
	}
	return false
}

func TestToolchainFixtureMatrix_LintFiles(t *testing.T) {
	t.Run("oas30 minimal root", func(t *testing.T) {
		diags := lintFixtureFile(t, "oas30-minimal.yaml")
		if hasBarrelCode(diags, "oas3-schema") {
			t.Fatalf("expected no navigator structural/meta diagnostics, got %#v", diags)
		}
	})

	t.Run("structural missing responses", func(t *testing.T) {
		diags := lintFixtureFile(t, "structural-missing-responses.yaml")
		if !hasBarrelIssueCode(diags, "structural.missing-responses") {
			t.Fatalf("expected structural.missing-responses via oas3-schema, got %#v", diags)
		}
		if !hasBarrelDocumentKind(diags, "openapi") {
			t.Fatalf("expected documentKind=openapi in structural bridge diagnostic, got %#v", diags)
		}
	})

	t.Run("workspace multifile root", func(t *testing.T) {
		diags := lintFixtureFile(t, filepath.Join("multifile", "api.yaml"))
		if hasBarrelCode(diags, "oas3-schema") {
			t.Fatalf("expected no navigator structural/meta diagnostics, got %#v", diags)
		}
		if hasBarrelCode(diags, "unresolved-ref") {
			t.Fatalf("expected no unresolved-ref diagnostics, got %#v", diags)
		}
	})

	t.Run("workspace missing target", func(t *testing.T) {
		diags := lintFixtureFile(t, filepath.Join("multifile-broken", "api.yaml"))
		if !hasBarrelCode(diags, "unresolved-ref") {
			t.Fatalf("expected unresolved-ref diagnostic, got %#v", diags)
		}
	})

	t.Run("arazzo valid root", func(t *testing.T) {
		diags := lintFixtureFile(t, "arazzo-simple.yaml")
		if hasBarrelCode(diags, "oas3-schema") {
			t.Fatalf("expected no navigator structural/meta diagnostics for valid arazzo fixture, got %#v", diags)
		}
	})

	t.Run("arazzo structural issue bridge", func(t *testing.T) {
		diags := lintFixtureFile(t, "arazzo-invalid-missing-workflows.yaml")
		if !hasBarrelIssueCode(diags, "structural.missing-workflows") {
			t.Fatalf("expected structural.missing-workflows via oas3-schema, got %#v", diags)
		}
		if !hasBarrelDocumentKind(diags, "arazzo") {
			t.Fatalf("expected documentKind=arazzo in structural bridge diagnostic, got %#v", diags)
		}
	})
}
