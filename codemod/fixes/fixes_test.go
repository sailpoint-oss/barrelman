package fixes_test

import (
	"testing"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/analyzers"
	"github.com/sailpoint-oss/barrelman/codemod"
	navigator "github.com/sailpoint-oss/navigator"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// fixCase is the shape of a single Phase 2 golden test: a minimal
// input spec that violates a rule, and the expected source bytes
// after applying the rule's auto-fix. Each case runs through four
// assertions: diagnostic emitted, fix produced patches, output bytes
// match expected, and idempotence (second run is a no-op).
type fixCase struct {
	name    string
	rule    string
	input   string
	expect  string
	wantMsg string // optional substring expected on the diagnostic
}

func TestFixes_Golden(t *testing.T) {
	cases := []fixCase{
		{
			name: "parameter description insertion",
			rule: "sailpoint-parameter-description",
			input: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        "200":
          description: OK
`,
			expect: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
          description: TODO
      responses:
        "200":
          description: OK
`,
			wantMsg: "must include a description",
		},
		{
			name: "path param required insertion",
			rule: "sailpoint-path-param-required",
			input: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users/{userId}:
    get:
      operationId: getUser
      parameters:
        - name: userId
          in: path
          schema:
            type: string
      responses:
        "200":
          description: OK
`,
			expect: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users/{userId}:
    get:
      operationId: getUser
      parameters:
        - name: userId
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          description: OK
`,
			wantMsg: "must set required: true",
		},
		{
			name: "operation single tag insertion",
			rule: "sailpoint-operation-single-tag",
			input: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        "200":
          description: OK
`,
			expect: `openapi: "3.1.0"
info:
  title: Test
  version: "1.0"
paths:
  /users:
    get:
      operationId: listUsers
      responses:
        "200":
          description: OK
      tags: [TODO]
`,
			wantMsg: "must declare exactly one tag",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			runFixCase(t, tc)
		})
	}
}

func runFixCase(t *testing.T, tc fixCase) {
	t.Helper()

	reg := barrelman.NewRegistry()
	analyzers.RegisterAll(reg)

	rule, ok := findRule(reg, tc.rule)
	if !ok {
		t.Fatalf("rule %q not registered", tc.rule)
	}
	if rule.Fix == nil {
		t.Fatalf("rule %q has no Fix attached", tc.rule)
	}

	idx, tree, content := parseSource(t, tc.input)
	defer tree.Close()

	// 1. Emit diagnostics.
	diags := rule.Run(&barrelman.AnalysisContext{
		Index:    idx,
		Tree:     tree,
		Language: navigator.YAMLLanguage(),
		Content:  content,
		URI:      "file:///test/spec.yaml",
	})
	if len(diags) == 0 {
		t.Fatalf("expected at least one diagnostic for rule %q", tc.rule)
	}
	if tc.wantMsg != "" && !containsAny(diags, tc.wantMsg) {
		t.Fatalf("no diagnostic contained %q; got %d diag(s)", tc.wantMsg, len(diags))
	}

	// 2. Produce patches via the rule's Fix.
	fixCtx := &codemod.FixContext{
		Index:  idx,
		Source: content,
		URI:    "file:///test/spec.yaml",
	}
	var patches []codemod.Patch
	for _, d := range diags {
		ps, err := rule.Fix(fixCtx, d)
		if err != nil {
			t.Fatalf("Fix returned error: %v", err)
		}
		patches = append(patches, ps...)
	}
	if len(patches) == 0 {
		t.Fatalf("expected at least one patch for rule %q", tc.rule)
	}

	// 3. Apply patches and compare bytes.
	got, err := (&codemod.Driver{}).Apply(content, patches)
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if string(got) != tc.expect {
		t.Fatalf("post-fix bytes mismatch\n--- got\n%s\n--- want\n%s", got, tc.expect)
	}

	// 4. Idempotence: re-parse, re-run the rule, re-fix. Should yield
	//    zero patches because the fix's precondition is now satisfied.
	idx2, tree2, content2 := parseSource(t, string(got))
	defer tree2.Close()
	diags2 := rule.Run(&barrelman.AnalysisContext{
		Index:    idx2,
		Tree:     tree2,
		Language: navigator.YAMLLanguage(),
		Content:  content2,
		URI:      "file:///test/spec.yaml",
	})
	if len(diags2) == 0 {
		// No second-run diagnostics means the fix worked and
		// idempotence is trivially satisfied.
		return
	}
	fixCtx2 := &codemod.FixContext{Index: idx2, Source: content2, URI: "file:///test/spec.yaml"}
	for _, d := range diags2 {
		ps, err := rule.Fix(fixCtx2, d)
		if err != nil {
			t.Fatalf("Fix on re-parsed source: %v", err)
		}
		if len(ps) != 0 {
			t.Fatalf("idempotence broken: %d patches on re-run", len(ps))
		}
	}
}

func findRule(reg *barrelman.Registry, id string) (barrelman.Rule, bool) {
	for _, r := range reg.AllRules() {
		if r.ID == id {
			return r, true
		}
	}
	return barrelman.Rule{}, false
}

func parseSource(t *testing.T, src string) (*navigator.Index, *ts.Tree, []byte) {
	t.Helper()
	content := []byte(src)
	lang := navigator.YAMLLanguage()
	parser := ts.NewParser()
	t.Cleanup(parser.Close)
	if err := parser.SetLanguage(lang); err != nil {
		t.Fatalf("set language: %v", err)
	}
	tree := parser.Parse(content, nil)
	if tree == nil {
		t.Fatal("parse returned nil tree")
	}
	idx := navigator.ParseTree(tree, content, "file:///test/spec.yaml", navigator.FormatYAML)
	if idx == nil {
		t.Fatal("ParseTree returned nil index")
	}
	return idx, tree, content
}

func containsAny(diags []barrelman.Diagnostic, needle string) bool {
	for _, d := range diags {
		if needle == "" {
			return true
		}
		if contains(d.Message, needle) {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	if len(sub) == 0 {
		return true
	}
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
