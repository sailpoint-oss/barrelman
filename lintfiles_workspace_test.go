package barrelman_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/analyzers"
)

func writeWorkspaceFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
}

func unresolvedRefRule(t *testing.T) barrelman.Rule {
	t.Helper()
	reg := barrelman.NewRegistry()
	analyzers.RegisterAll(reg)
	for _, rule := range reg.AllRules() {
		if rule.ID == "unresolved-ref" {
			return rule
		}
	}
	t.Fatal("unresolved-ref rule not registered")
	return barrelman.Rule{}
}

func TestLintFiles_ProvidesWorkspaceResolver(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceFiles(t, root, map[string]string{
		"api.yaml": `openapi: "3.0.3"
info:
  title: Workspace API
  version: "1.0.0"
paths:
  /pets:
    get:
      operationId: listPets
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "./schemas/Pet.yaml"`,
		"schemas/Pet.yaml": `type: object
properties:
  name:
    type: string`,
	})

	apiPath := filepath.Join(root, "api.yaml")
	var seenRoot string
	var seenURI string
	var sawResolver bool
	var canResolveExternal bool

	rules := []barrelman.Rule{{
		ID: "capture-workspace",
		Run: func(ctx *barrelman.AnalysisContext) []barrelman.Diagnostic {
			seenRoot = ctx.WorkspaceRoot
			seenURI = ctx.URI
			sawResolver = ctx.Resolver != nil
			if ctx.Resolver != nil {
				canResolveExternal = ctx.Resolver.CanResolve(ctx.URI, "./schemas/Pet.yaml")
			}
			return nil
		},
	}}

	results, err := barrelman.LintFiles([]string{apiPath}, barrelman.LintOptions{
		WorkspaceRoot: root,
		Rules:         rules,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 lint result, got %d", len(results))
	}
	if seenRoot != root {
		t.Fatalf("workspace root = %q, want %q", seenRoot, root)
	}
	if seenURI == "" {
		t.Fatal("expected rule to receive a file URI")
	}
	if !sawResolver {
		t.Fatal("expected resolver to be attached for LintFiles")
	}
	if !canResolveExternal {
		t.Fatal("expected resolver to resolve sibling schema from linted file")
	}
}

func TestLintFiles_UnresolvedRefRuleAllowsResolvableCrossFileRefs(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceFiles(t, root, map[string]string{
		"api.yaml": `openapi: "3.0.3"
info:
  title: Workspace API
  version: "1.0.0"
paths:
  /pets:
    get:
      operationId: listPets
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "./schemas/Pet.yaml"`,
		"schemas/Pet.yaml": `type: object
properties:
  name:
    type: string`,
	})

	results, err := barrelman.LintFiles(
		[]string{filepath.Join(root, "api.yaml")},
		barrelman.LintOptions{
			WorkspaceRoot: root,
			Rules:         []barrelman.Rule{unresolvedRefRule(t)},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 lint result, got %d", len(results))
	}
	if len(results[0].Diagnostics) != 0 {
		t.Fatalf("expected no unresolved-ref diagnostics, got %#v", results[0].Diagnostics)
	}
}

func TestLintFiles_UnresolvedRefRuleReportsMissingCrossFileRefs(t *testing.T) {
	root := t.TempDir()
	writeWorkspaceFiles(t, root, map[string]string{
		"api.yaml": `openapi: "3.0.3"
info:
  title: Workspace API
  version: "1.0.0"
paths:
  /pets:
    get:
      operationId: listPets
      responses:
        "200":
          description: OK
          content:
            application/json:
              schema:
                $ref: "./schemas/Missing.yaml"`,
	})

	results, err := barrelman.LintFiles(
		[]string{filepath.Join(root, "api.yaml")},
		barrelman.LintOptions{
			WorkspaceRoot: root,
			Rules:         []barrelman.Rule{unresolvedRefRule(t)},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 lint result, got %d", len(results))
	}
	if len(results[0].Diagnostics) != 1 {
		t.Fatalf("expected 1 unresolved-ref diagnostic, got %#v", results[0].Diagnostics)
	}
	if results[0].Diagnostics[0].Code != "unresolved-ref" {
		t.Fatalf("diagnostic code = %q, want unresolved-ref", results[0].Diagnostics[0].Code)
	}
}
