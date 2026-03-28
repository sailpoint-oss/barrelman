package barrelman_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sailpoint-oss/barrelman"
)

func TestLintFiles_ConfigPathAppliesOverrides(t *testing.T) {
	dir := t.TempDir()
	specPath := filepath.Join(dir, "api.yaml")
	configPath := filepath.Join(dir, ".barrelman.yaml")

	spec := `openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths:
  /pets:
    get: {}`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatalf("write spec: %v", err)
	}
	if err := os.WriteFile(configPath, []byte(`extends: barrelman:recommended
rules:
  info-description: error
  oas3-schema: off
`), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	results, err := barrelman.LintFiles([]string{specPath}, barrelman.LintOptions{
		Rules:      allRules(),
		ConfigPath: configPath,
	})
	if err != nil {
		t.Fatalf("LintFiles: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("result count = %d, want 1", len(results))
	}

	diags := results[0].Diagnostics
	if hasDiagCode(diags, "oas3-schema") {
		t.Fatalf("expected oas3-schema to be disabled by config override, got %#v", diags)
	}
	requireDiagWithSeverity(t, diags, "info-description", barrelman.SeverityError)
}

func TestLintContent_RulesetPathAppliesOverrides(t *testing.T) {
	dir := t.TempDir()
	rulesetPath := filepath.Join(dir, "ruleset.yaml")
	if err := os.WriteFile(rulesetPath, []byte(`extends: barrelman:recommended
rules:
  info-description: error
`), 0o644); err != nil {
		t.Fatalf("write ruleset: %v", err)
	}

	spec := []byte(`openapi: "3.0.3"
info:
  title: Test
  version: "1.0.0"
paths: {}`)
	diags, err := barrelman.LintContent("file:///test/spec.yaml", spec, barrelman.LintOptions{
		Rules:       allRules(),
		RulesetPath: rulesetPath,
	})
	if err != nil {
		t.Fatalf("LintContent: %v", err)
	}

	requireDiagWithSeverity(t, diags, "info-description", barrelman.SeverityError)
}
