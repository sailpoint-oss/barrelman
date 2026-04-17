package engine

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCompileAndValidate_TypeMismatch(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"age": map[string]any{"type": "integer"},
		},
	}
	s, err := Compile(schema, CompileOpts{})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	instance := map[string]any{"age": "not-a-number"}
	issues := s.Validate(instance)
	if len(issues) == 0 {
		t.Fatal("expected at least one issue for type mismatch")
	}
	var foundType bool
	for _, is := range issues {
		if is.Code == "type" {
			foundType = true
			if is.Pointer != "/age" {
				t.Errorf("Pointer = %q, want /age", is.Pointer)
			}
			if is.Path != "age" {
				t.Errorf("Path = %q, want age", is.Path)
			}
			if is.Expected == "" || is.Received == "" {
				t.Errorf("expected/received missing on type issue: %+v", is)
			}
		}
	}
	if !foundType {
		t.Errorf("no `type` issue in: %+v", issues)
	}
}

func TestCompileAndValidate_MissingRequired(t *testing.T) {
	schema := map[string]any{
		"type":     "object",
		"required": []any{"name"},
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
	}
	s, err := Compile(schema, CompileOpts{})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	issues := s.Validate(map[string]any{})
	if len(issues) == 0 {
		t.Fatal("expected issue for missing required")
	}
}

func TestCompileAndValidate_EnumMismatch(t *testing.T) {
	schema := map[string]any{
		"type": "string",
		"enum": []any{"RED", "GREEN", "BLUE"},
	}
	s, err := Compile(schema, CompileOpts{})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	issues := s.Validate("MAUVE")
	if len(issues) == 0 {
		t.Fatal("expected enum issue")
	}
	var found bool
	for _, is := range issues {
		if is.Code == "enum" {
			found = true
			if is.Expected == "" || !strings.Contains(is.Expected, "RED") {
				t.Errorf("Expected should include enum choices, got %q", is.Expected)
			}
		}
	}
	if !found {
		t.Errorf("no enum issue in: %+v", issues)
	}
}

func TestCompile_NullableRewrite(t *testing.T) {
	// OpenAPI 3.0 schema: boolean, nullable: true.
	schema := map[string]any{
		"type":     "boolean",
		"nullable": true,
	}
	s, err := Compile(schema, CompileOpts{
		OpenAPI: OpenAPIOptions{NullableRewrite: true},
	})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	// Null should validate.
	if issues := s.Validate(nil); len(issues) != 0 {
		t.Errorf("nullable boolean should accept null, got %+v", issues)
	}
	// Boolean should validate.
	if issues := s.Validate(true); len(issues) != 0 {
		t.Errorf("nullable boolean should accept true, got %+v", issues)
	}
	// Non-boolean non-null should fail.
	if issues := s.Validate("hello"); len(issues) == 0 {
		t.Errorf("nullable boolean should reject string")
	}
}

func TestValidateBytes_RoundTrip(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"ok": map[string]any{"type": "boolean"},
		},
	}
	s, err := Compile(schema, CompileOpts{})
	if err != nil {
		t.Fatalf("Compile: %v", err)
	}
	raw, _ := json.Marshal(map[string]any{"ok": "yes"})
	issues, err := s.ValidateBytes(raw)
	if err != nil {
		t.Fatalf("ValidateBytes: %v", err)
	}
	if len(issues) == 0 {
		t.Fatal("expected validation failure for non-boolean value")
	}
}

func TestIssue_HumanPath(t *testing.T) {
	cases := map[string]string{
		"":                   "<root>",
		"/":                  "<root>",
		"/components":        "components",
		"/components/schemas/User/age": "components.schemas.User.age",
		"/~1weird":           "/weird",
	}
	for ptr, want := range cases {
		issue := Issue{Pointer: ptr}
		if got := issue.HumanPath(); got != want {
			t.Errorf("HumanPath(%q) = %q, want %q", ptr, got, want)
		}
	}
}
