package spectral

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func mustParseDoc(t *testing.T, yamlStr string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	return &doc
}

// ---------------------------------------------------------------------------
// EvaluateJSONPath
// ---------------------------------------------------------------------------

func TestEvaluateJSONPath_Root(t *testing.T) {
	doc := mustParseDoc(t, "title: hello")
	results, err := EvaluateJSONPath(doc, "$")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected at least one result for $")
	}
}

func TestEvaluateJSONPath_InfoMapping(t *testing.T) {
	doc := mustParseDoc(t, `
info:
  title: My API
  version: "1.0"
`)
	results, err := EvaluateJSONPath(doc, "$.info")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Kind != yaml.MappingNode {
		t.Errorf("expected mapping node, got kind %d", results[0].Kind)
	}
}

func TestEvaluateJSONPath_InfoNestedField(t *testing.T) {
	doc := mustParseDoc(t, `
info:
  title: My API
  version: "1.0"
`)
	results, err := EvaluateJSONPath(doc, "$.info.title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Kind != yaml.ScalarNode {
		t.Errorf("expected scalar node, got kind %d", results[0].Kind)
	}
	if results[0].Value != "My API" {
		t.Errorf("expected 'My API', got %q", results[0].Value)
	}
}

func TestEvaluateJSONPath_Wildcard(t *testing.T) {
	doc := mustParseDoc(t, `
paths:
  /users:
    get:
      summary: List users
  /items:
    get:
      summary: List items
`)
	results, err := EvaluateJSONPath(doc, "$.paths.*")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results for wildcard, got %d", len(results))
	}
}

func TestEvaluateJSONPath_Nonexistent(t *testing.T) {
	doc := mustParseDoc(t, "title: hello")
	results, err := EvaluateJSONPath(doc, "$.nonexistent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for nonexistent path, got %d", len(results))
	}
}

func TestEvaluateJSONPath_InvalidExpression(t *testing.T) {
	doc := mustParseDoc(t, "title: hello")
	_, err := EvaluateJSONPath(doc, "$[invalid[")
	if err == nil {
		t.Fatal("expected error for invalid JSONPath expression")
	}
}

func TestEvaluateJSONPath_DeepNested(t *testing.T) {
	doc := mustParseDoc(t, `
a:
  b:
    c:
      d: found
`)
	results, err := EvaluateJSONPath(doc, "$.a.b.c.d")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Value != "found" {
		t.Errorf("expected 'found', got %q", results[0].Value)
	}
}

// ---------------------------------------------------------------------------
// nodeValue
// ---------------------------------------------------------------------------

func TestNodeValue_Scalar(t *testing.T) {
	node := parseNode(t, `hello`)
	val := nodeValue(node)
	if s, ok := val.(string); !ok || s != "hello" {
		t.Errorf("expected string 'hello', got %T %v", val, val)
	}
}

func TestNodeValue_Integer(t *testing.T) {
	node := parseNode(t, `42`)
	val := nodeValue(node)
	if n, ok := val.(int); !ok || n != 42 {
		t.Errorf("expected int 42, got %T %v", val, val)
	}
}

func TestNodeValue_Boolean(t *testing.T) {
	node := parseNode(t, `true`)
	val := nodeValue(node)
	if b, ok := val.(bool); !ok || !b {
		t.Errorf("expected bool true, got %T %v", val, val)
	}
}

func TestNodeValue_Sequence(t *testing.T) {
	node := parseNode(t, "[a, b, c]")
	val := nodeValue(node)
	items, ok := val.([]interface{})
	if !ok {
		t.Fatalf("expected []interface{}, got %T", val)
	}
	if len(items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(items))
	}
}

func TestNodeValue_Mapping(t *testing.T) {
	node := parseNode(t, "key: value")
	val := nodeValue(node)
	m, ok := val.(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", val)
	}
	if m["key"] != "value" {
		t.Errorf("expected key='value', got %v", m["key"])
	}
}

func TestNodeValue_Nil(t *testing.T) {
	val := nodeValue(nil)
	if val != nil {
		t.Errorf("expected nil, got %T %v", val, val)
	}
}

// ---------------------------------------------------------------------------
// nodeField
// ---------------------------------------------------------------------------

func TestNodeField_Found(t *testing.T) {
	node := parseNode(t, "title: hello\nversion: 1.0")
	child := nodeField(node, "title")
	if child == nil {
		t.Fatal("expected to find 'title' field")
	}
	if child.Value != "hello" {
		t.Errorf("expected 'hello', got %q", child.Value)
	}
}

func TestNodeField_NotFound(t *testing.T) {
	node := parseNode(t, "title: hello")
	child := nodeField(node, "missing")
	if child != nil {
		t.Fatalf("expected nil for missing field, got %v", child)
	}
}

func TestNodeField_NilNode(t *testing.T) {
	child := nodeField(nil, "field")
	if child != nil {
		t.Fatalf("expected nil for nil node, got %v", child)
	}
}

func TestNodeField_NonMapping(t *testing.T) {
	node := parseNode(t, `hello`)
	child := nodeField(node, "anything")
	if child != nil {
		t.Fatalf("expected nil for scalar node, got %v", child)
	}
}

func TestNodeField_DocumentNode(t *testing.T) {
	// mustParseDoc returns the document node (not unwrapped)
	var doc yaml.Node
	err := yaml.Unmarshal([]byte("title: hello"), &doc)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	// doc is a DocumentNode wrapping a MappingNode
	child := nodeField(&doc, "title")
	if child == nil {
		t.Fatal("expected to find 'title' through document node")
	}
	if child.Value != "hello" {
		t.Errorf("expected 'hello', got %q", child.Value)
	}
}

// ---------------------------------------------------------------------------
// nodeHasField
// ---------------------------------------------------------------------------

func TestNodeHasField_True(t *testing.T) {
	node := parseNode(t, "title: hello")
	if !nodeHasField(node, "title") {
		t.Error("expected true for existing field")
	}
}

func TestNodeHasField_False(t *testing.T) {
	node := parseNode(t, "title: hello")
	if nodeHasField(node, "missing") {
		t.Error("expected false for missing field")
	}
}
