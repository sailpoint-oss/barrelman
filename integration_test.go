package barrelman_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/analyzers"
	"github.com/sailpoint-oss/barrelman/checks"
)

func allRules() []barrelman.Rule {
	reg := barrelman.NewRegistry()
	analyzers.RegisterAll(reg)
	checks.RegisterAll(reg)
	return reg.AllRules()
}

func lintYAML(t *testing.T, spec string) []barrelman.Diagnostic {
	t.Helper()
	diags, err := barrelman.LintContent(
		"file:///test/spec.yaml", []byte(spec),
		barrelman.LintOptions{Rules: allRules()},
	)
	if err != nil {
		t.Fatalf("LintContent error: %v", err)
	}
	return diags
}

func hasDiagCode(diags []barrelman.Diagnostic, code string) bool {
	for _, d := range diags {
		if d.Code == code {
			return true
		}
	}
	return false
}

func requireDiag(t *testing.T, diags []barrelman.Diagnostic, code string) {
	t.Helper()
	if !hasDiagCode(diags, code) {
		t.Errorf("expected diagnostic %q not found in %d diagnostics", code, len(diags))
		for _, d := range diags {
			t.Logf("  got: [%s] sev=%d L%d: %s", d.Code, d.Severity, d.Range.Start.Line, d.Message)
		}
	}
}

func requireDiagWithSeverity(t *testing.T, diags []barrelman.Diagnostic, code string, sev barrelman.Severity) {
	t.Helper()
	for _, d := range diags {
		if d.Code == code && d.Severity == sev {
			return
		}
	}
	t.Errorf("expected diagnostic %q with severity %d not found", code, sev)
	for _, d := range diags {
		t.Logf("  got: [%s] sev=%d L%d: %s", d.Code, d.Severity, d.Range.Start.Line, d.Message)
	}
}

func countDiagCode(diags []barrelman.Diagnostic, code string) int {
	n := 0
	for _, d := range diags {
		if d.Code == code {
			n++
		}
	}
	return n
}

func isZeroRange(r barrelman.Range) bool {
	return r.Start.Line == 0 && r.Start.Character == 0 &&
		r.End.Line == 0 && r.End.Character == 0
}

// TestIntegration_LintContent_ValidSpec verifies the full pipeline runs on a
// well-formed OpenAPI 3.0 spec. Structural validation (oas3-schema) is excluded
// from the error check because the embedded JSON Schema has known false
// positives around license/contact property validation.
func TestIntegration_LintContent_ValidSpec(t *testing.T) {
	spec := `openapi: "3.0.3"
info:
  title: Petstore
  version: "1.0.0"
  description: A sample Petstore API for integration testing.
tags:
  - name: Pets
    description: Pet operations
servers:
  - url: https://api.example.com
    description: Production server
paths:
  /pets:
    get:
      operationId: listPets
      description: Returns a list of pets.
      tags:
        - Pets
      responses:
        "200":
          description: A list of pets
        "401":
          description: Unauthorized
        "500":
          description: Internal server error
components:
  schemas:
    Pet:
      type: object
      description: A pet in the store.
      properties:
        name:
          type: string
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
security:
  - BearerAuth: []`

	diags := lintYAML(t, spec)

	var errors []barrelman.Diagnostic
	for _, d := range diags {
		if d.Severity == barrelman.SeverityError && d.Code != "oas3-schema" {
			errors = append(errors, d)
		}
	}
	if len(errors) > 0 {
		t.Errorf("expected no error-level diagnostics (excluding oas3-schema), got %d:", len(errors))
		for _, d := range errors {
			t.Errorf("  [%s] L%d: %s", d.Code, d.Range.Start.Line, d.Message)
		}
	}

	for _, d := range diags {
		if d.Source == "" {
			t.Errorf("diagnostic [%s] has empty Source", d.Code)
		}
	}
}

// TestIntegration_LintContent_MissingDescription verifies the info-description
// rule fires when info.description is absent.
func TestIntegration_LintContent_MissingDescription(t *testing.T) {
	spec := `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}`

	diags := lintYAML(t, spec)
	requireDiagWithSeverity(t, diags, "info-description", barrelman.SeverityWarning)

	for _, d := range diags {
		if d.Code == "info-description" && isZeroRange(d.Range) {
			t.Error("info-description diagnostic has zero range")
		}
	}
}

// TestIntegration_LintContent_SecurityIssues verifies that an API key passed
// via query param triggers both the core security rule and OWASP variant.
func TestIntegration_LintContent_SecurityIssues(t *testing.T) {
	spec := `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  description: Security test spec.
paths:
  /pets:
    get:
      operationId: listPets
      description: List pets.
      tags:
        - Pets
      responses:
        "200":
          description: ok
components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: query
      name: api_key`

	diags := lintYAML(t, spec)
	requireDiag(t, diags, "no-api-key-in-query")
	requireDiag(t, diags, "owasp-no-api-keys-in-url")

	for _, d := range diags {
		if d.Code == "no-api-key-in-query" {
			if d.Severity != barrelman.SeverityWarning {
				t.Errorf("no-api-key-in-query severity: got %d, want %d", d.Severity, barrelman.SeverityWarning)
			}
			if isZeroRange(d.Range) {
				t.Error("no-api-key-in-query diagnostic has zero range")
			}
			break
		}
	}
}

// TestIntegration_LintContent_NamingViolation verifies that a lowercase schema
// name triggers the schema-name-capital rule.
func TestIntegration_LintContent_NamingViolation(t *testing.T) {
	spec := `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  description: Naming test.
paths: {}
components:
  schemas:
    pet:
      type: object
      properties:
        name:
          type: string`

	diags := lintYAML(t, spec)
	requireDiagWithSeverity(t, diags, "schema-name-capital", barrelman.SeverityWarning)

	for _, d := range diags {
		if d.Code == "schema-name-capital" {
			if d.Message == "" {
				t.Error("schema-name-capital diagnostic has empty message")
			}
			break
		}
	}
}

// TestIntegration_LintContent_DuplicateKeys verifies that duplicate YAML
// mapping keys are detected via the tree-sitter based syntactic check.
func TestIntegration_LintContent_DuplicateKeys(t *testing.T) {
	spec := `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
paths: {}
paths: {}`

	diags := lintYAML(t, spec)
	requireDiagWithSeverity(t, diags, "duplicate-keys", barrelman.SeverityError)

	for _, d := range diags {
		if d.Code == "duplicate-keys" {
			if d.Range.Start.Line != 5 {
				t.Errorf("duplicate-keys line: got %d, want 5", d.Range.Start.Line)
			}
			if isZeroRange(d.Range) {
				t.Error("duplicate-keys diagnostic has zero range")
			}
			return
		}
	}
}

// TestIntegration_LintContent_ContactProperties exercises info.contact
// detection. The navigator tree-sitter parser does not populate Contact on the
// Info object, so the contact-properties rule (which requires Contact != nil)
// cannot fire through LintContent. Instead, the info-contact rule fires because
// it sees Contact as nil. This test documents the current pipeline behavior;
// when navigator adds Contact parsing, contact-properties will become testable.
func TestIntegration_LintContent_ContactProperties(t *testing.T) {
	spec := `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  description: Contact test.
paths: {}`

	diags := lintYAML(t, spec)

	requireDiagWithSeverity(t, diags, "info-contact", barrelman.SeverityWarning)

	if hasDiagCode(diags, "contact-properties") {
		t.Log("contact-properties fired — navigator now parses Contact; update this test")
	}
}

// TestIntegration_LintContent_LicenseUrl exercises info.license detection. The
// navigator tree-sitter parser does not populate License on the Info object, so
// the license-url rule cannot fire through LintContent. Instead, info-license
// fires because it sees License as nil. This documents the current pipeline
// behavior.
func TestIntegration_LintContent_LicenseUrl(t *testing.T) {
	spec := `openapi: "3.0.3"
info:
  title: Test
  version: "1.0"
  description: License test.
paths: {}`

	diags := lintYAML(t, spec)

	requireDiagWithSeverity(t, diags, "info-license", barrelman.SeverityWarning)

	if hasDiagCode(diags, "license-url") {
		t.Log("license-url fired — navigator now parses License; update this test")
	}
}

// TestIntegration_LintContent_FullPipeline uses a realistic multi-issue spec
// and verifies that diagnostics from multiple rule categories appear.
func TestIntegration_LintContent_FullPipeline(t *testing.T) {
	spec := `openapi: "3.0.3"
info:
  title: Broken API
  version: "0.1"
paths:
  /items:
    get:
      responses:
        "200":
          description: ok
          content:
            application/json:
              schema:
                type: array
                items:
                  type: object
    post:
      requestBody:
        content:
          application/json:
            schema:
              type: object
      responses:
        "201":
          description: created
components:
  schemas:
    badName:
      type: object
      properties:
        id:
          type: integer
        name:
          type: string
  securitySchemes:
    ApiKey:
      type: apiKey
      in: query
      name: key`

	diags := lintYAML(t, spec)

	expected := []string{
		"info-description",
		"info-contact",
		"info-license",
		"schema-name-capital",
		"no-api-key-in-query",
		"owasp-no-api-keys-in-url",
		"operation-operationId",
		"operation-description",
		"operation-tags",
		"missing-error-responses",
		"security-global-or-operation",
	}

	codes := make(map[string]bool)
	for _, d := range diags {
		codes[d.Code] = true
	}

	for _, exp := range expected {
		if !codes[exp] {
			t.Errorf("expected rule %q to fire, but it did not", exp)
		}
	}

	for _, d := range diags {
		if d.Source == "" {
			t.Errorf("diagnostic [%s] has empty Source", d.Code)
		}
	}

	if len(diags) < len(expected) {
		t.Errorf("expected at least %d diagnostics, got %d", len(expected), len(diags))
	}
}
