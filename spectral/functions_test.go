package spectral

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func parseNode(t *testing.T, yamlStr string) *yaml.Node {
	t.Helper()
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(yamlStr), &doc); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if doc.Kind == yaml.DocumentNode && len(doc.Content) > 0 {
		return doc.Content[0]
	}
	return &doc
}

// ---------------------------------------------------------------------------
// truthy
// ---------------------------------------------------------------------------

func TestFuncTruthy_PassesOnTruthyValue(t *testing.T) {
	node := parseNode(t, `hello`)
	issues := funcTruthy(node, "", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d: %v", len(issues), issues)
	}
}

func TestFuncTruthy_FailsOnEmptyString(t *testing.T) {
	node := parseNode(t, `""`)
	issues := funcTruthy(node, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for empty string")
	}
}

func TestFuncTruthy_FailsOnNull(t *testing.T) {
	node := parseNode(t, `null`)
	issues := funcTruthy(node, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for null")
	}
}

func TestFuncTruthy_FailsOnFalse(t *testing.T) {
	node := parseNode(t, `false`)
	issues := funcTruthy(node, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for false")
	}
}

func TestFuncTruthy_FailsOnZero(t *testing.T) {
	node := parseNode(t, `0`)
	issues := funcTruthy(node, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for 0")
	}
}

func TestFuncTruthy_FailsOnEmptySequence(t *testing.T) {
	node := parseNode(t, `[]`)
	issues := funcTruthy(node, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for empty sequence")
	}
}

func TestFuncTruthy_FailsOnEmptyMapping(t *testing.T) {
	node := parseNode(t, `{}`)
	issues := funcTruthy(node, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for empty mapping")
	}
}

func TestFuncTruthy_PassesOnNonEmptySequence(t *testing.T) {
	node := parseNode(t, `[a, b]`)
	issues := funcTruthy(node, "", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncTruthy_WithField(t *testing.T) {
	node := parseNode(t, "title: hello\ndescription: world")
	issues := funcTruthy(node, "title", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for truthy field, got %d", len(issues))
	}
}

func TestFuncTruthy_WithMissingField(t *testing.T) {
	node := parseNode(t, "title: hello")
	issues := funcTruthy(node, "missing", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for missing field")
	}
}

// ---------------------------------------------------------------------------
// falsy
// ---------------------------------------------------------------------------

func TestFuncFalsy_PassesOnFalsyValue(t *testing.T) {
	cases := []string{`""`, `false`, `0`, `null`, `[]`, `{}`}
	for _, c := range cases {
		node := parseNode(t, c)
		issues := funcFalsy(node, "", nil)
		if len(issues) != 0 {
			t.Errorf("expected no issues for %q, got %d", c, len(issues))
		}
	}
}

func TestFuncFalsy_FailsOnTruthyValue(t *testing.T) {
	node := parseNode(t, `hello`)
	issues := funcFalsy(node, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for truthy value")
	}
}

func TestFuncFalsy_PassesOnNilTarget(t *testing.T) {
	node := parseNode(t, "title: hello")
	issues := funcFalsy(node, "missing", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for nil target, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// defined / undefined
// ---------------------------------------------------------------------------

func TestFuncDefined_PassesWhenFieldExists(t *testing.T) {
	node := parseNode(t, "title: hello")
	issues := funcDefined(node, "title", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncDefined_FailsWhenFieldMissing(t *testing.T) {
	node := parseNode(t, "title: hello")
	issues := funcDefined(node, "description", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for missing field")
	}
}

func TestFuncDefined_NoFieldNonNilNode(t *testing.T) {
	node := parseNode(t, `hello`)
	issues := funcDefined(node, "", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for non-nil node without field, got %d", len(issues))
	}
}

func TestFuncDefined_NoFieldNilNode(t *testing.T) {
	issues := funcDefined(nil, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for nil node without field")
	}
}

func TestFuncUndefined_PassesWhenFieldMissing(t *testing.T) {
	node := parseNode(t, "title: hello")
	issues := funcUndefined(node, "description", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncUndefined_FailsWhenFieldExists(t *testing.T) {
	node := parseNode(t, "title: hello")
	issues := funcUndefined(node, "title", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue when field exists")
	}
}

func TestFuncUndefined_NoFieldNilNode(t *testing.T) {
	issues := funcUndefined(nil, "", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for nil node, got %d", len(issues))
	}
}

func TestFuncUndefined_NoFieldNonNilNode(t *testing.T) {
	node := parseNode(t, `hello`)
	issues := funcUndefined(node, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for non-nil node without field")
	}
}

// ---------------------------------------------------------------------------
// pattern
// ---------------------------------------------------------------------------

func TestFuncPattern_MatchPasses(t *testing.T) {
	node := parseNode(t, `hello-world`)
	opts := map[string]interface{}{"match": `^hello`}
	issues := funcPattern(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncPattern_MatchFails(t *testing.T) {
	node := parseNode(t, `goodbye`)
	opts := map[string]interface{}{"match": `^hello`}
	issues := funcPattern(node, "", opts)
	if len(issues) == 0 {
		t.Fatal("expected issue for non-matching pattern")
	}
}

func TestFuncPattern_NotMatchPasses(t *testing.T) {
	node := parseNode(t, `hello`)
	opts := map[string]interface{}{"notMatch": `^bye`}
	issues := funcPattern(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncPattern_NotMatchFails(t *testing.T) {
	node := parseNode(t, `goodbye`)
	opts := map[string]interface{}{"notMatch": `^good`}
	issues := funcPattern(node, "", opts)
	if len(issues) == 0 {
		t.Fatal("expected issue for notMatch that matches")
	}
}

func TestFuncPattern_NilNodeReturnsNil(t *testing.T) {
	node := parseNode(t, "title: hello")
	opts := map[string]interface{}{"match": `.*`}
	issues := funcPattern(node, "missing", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for nil target, got %d", len(issues))
	}
}

func TestFuncPattern_NonScalarReturnsNil(t *testing.T) {
	node := parseNode(t, "[a, b]")
	opts := map[string]interface{}{"match": `.*`}
	issues := funcPattern(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for sequence node, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// casing
// ---------------------------------------------------------------------------

func TestFuncCasing(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		casing  string
		wantErr bool
	}{
		// kebab-case
		{"kebab pass", "my-property", "kebab", false},
		{"kebab fail underscore", "my_property", "kebab", true},
		{"kebab fail uppercase", "My-Property", "kebab", true},
		// camelCase
		{"camel pass", "myProperty", "camel", false},
		{"camel fail uppercase start", "MyProperty", "camel", true},
		{"camel fail hyphen", "my-Property", "camel", true},
		// PascalCase
		{"pascal pass", "MyProperty", "pascal", false},
		{"pascal fail lowercase start", "myProperty", "pascal", true},
		{"pascal fail underscore", "My_Property", "pascal", true},
		// snake_case
		{"snake pass", "my_property", "snake", false},
		{"snake fail hyphen", "my-property", "snake", true},
		{"snake fail uppercase", "My_Property", "snake", true},
		// flat
		{"flat pass", "myproperty", "flat", false},
		{"flat fail hyphen", "my-property", "flat", true},
		{"flat fail underscore", "my_property", "flat", true},
		{"flat fail uppercase", "MyProperty", "flat", true},
		// empty value always passes
		{"empty value", "", "kebab", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := parseNode(t, tt.value)
			opts := map[string]interface{}{"type": tt.casing}
			issues := funcCasing(node, "", opts)
			if tt.wantErr && len(issues) == 0 {
				t.Errorf("expected issue for %q with casing %q", tt.value, tt.casing)
			}
			if !tt.wantErr && len(issues) != 0 {
				t.Errorf("unexpected issue for %q with casing %q: %v", tt.value, tt.casing, issues[0].Message)
			}
		})
	}
}

func TestFuncCasing_NoCasingType(t *testing.T) {
	node := parseNode(t, `hello`)
	issues := funcCasing(node, "", map[string]interface{}{})
	if len(issues) != 0 {
		t.Fatalf("expected no issues when no casing type given, got %d", len(issues))
	}
}

func TestFuncCasing_NonScalarReturnsNil(t *testing.T) {
	node := parseNode(t, `[a, b]`)
	opts := map[string]interface{}{"type": "kebab"}
	issues := funcCasing(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for non-scalar, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// length
// ---------------------------------------------------------------------------

func TestFuncLength_ScalarMinPass(t *testing.T) {
	node := parseNode(t, `hello`)
	opts := map[string]interface{}{"min": 3}
	issues := funcLength(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncLength_ScalarMinFail(t *testing.T) {
	node := parseNode(t, `hi`)
	opts := map[string]interface{}{"min": 5}
	issues := funcLength(node, "", opts)
	if len(issues) == 0 {
		t.Fatal("expected issue for short string")
	}
}

func TestFuncLength_ScalarMaxPass(t *testing.T) {
	node := parseNode(t, `hi`)
	opts := map[string]interface{}{"max": 5}
	issues := funcLength(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncLength_ScalarMaxFail(t *testing.T) {
	node := parseNode(t, `hello world`)
	opts := map[string]interface{}{"max": 5}
	issues := funcLength(node, "", opts)
	if len(issues) == 0 {
		t.Fatal("expected issue for long string")
	}
}

func TestFuncLength_SequenceMin(t *testing.T) {
	node := parseNode(t, "[a, b, c]")
	opts := map[string]interface{}{"min": 2}
	issues := funcLength(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncLength_SequenceMinFail(t *testing.T) {
	node := parseNode(t, "[a]")
	opts := map[string]interface{}{"min": 3}
	issues := funcLength(node, "", opts)
	if len(issues) == 0 {
		t.Fatal("expected issue for short sequence")
	}
}

func TestFuncLength_MappingMax(t *testing.T) {
	node := parseNode(t, "a: 1\nb: 2")
	opts := map[string]interface{}{"max": 5}
	issues := funcLength(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncLength_MappingMaxFail(t *testing.T) {
	node := parseNode(t, "a: 1\nb: 2\nc: 3")
	opts := map[string]interface{}{"max": 2}
	issues := funcLength(node, "", opts)
	if len(issues) == 0 {
		t.Fatal("expected issue for large mapping")
	}
}

func TestFuncLength_NilTarget(t *testing.T) {
	node := parseNode(t, "title: hello")
	opts := map[string]interface{}{"min": 1}
	issues := funcLength(node, "missing", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for nil target, got %d", len(issues))
	}
}

func TestFuncLength_WithField(t *testing.T) {
	node := parseNode(t, "title: hello")
	opts := map[string]interface{}{"min": 3}
	issues := funcLength(node, "title", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// enumeration
// ---------------------------------------------------------------------------

func TestFuncEnumeration_AllowedValue(t *testing.T) {
	node := parseNode(t, `hello`)
	opts := map[string]interface{}{
		"values": []interface{}{"hello", "world"},
	}
	issues := funcEnumeration(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncEnumeration_DisallowedValue(t *testing.T) {
	node := parseNode(t, `foo`)
	opts := map[string]interface{}{
		"values": []interface{}{"hello", "world"},
	}
	issues := funcEnumeration(node, "", opts)
	if len(issues) == 0 {
		t.Fatal("expected issue for disallowed value")
	}
}

func TestFuncEnumeration_NonScalar(t *testing.T) {
	node := parseNode(t, `[a, b]`)
	opts := map[string]interface{}{
		"values": []interface{}{"a"},
	}
	issues := funcEnumeration(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for non-scalar, got %d", len(issues))
	}
}

func TestFuncEnumeration_NoValuesOption(t *testing.T) {
	node := parseNode(t, `hello`)
	issues := funcEnumeration(node, "", map[string]interface{}{})
	if len(issues) != 0 {
		t.Fatalf("expected no issues when no values option, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// alphabetical
// ---------------------------------------------------------------------------

func TestFuncAlphabetical_SortedPasses(t *testing.T) {
	node := parseNode(t, "[alpha, beta, gamma]")
	issues := funcAlphabetical(node, "", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncAlphabetical_UnsortedFails(t *testing.T) {
	node := parseNode(t, "[gamma, alpha, beta]")
	issues := funcAlphabetical(node, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issue for unsorted sequence")
	}
}

func TestFuncAlphabetical_NonSequenceReturnsNil(t *testing.T) {
	node := parseNode(t, `hello`)
	issues := funcAlphabetical(node, "", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for non-sequence, got %d", len(issues))
	}
}

func TestFuncAlphabetical_KeyedBy(t *testing.T) {
	sorted := `
- name: alpha
- name: beta
- name: gamma
`
	node := parseNode(t, sorted)
	opts := map[string]interface{}{"keyedBy": "name"}
	issues := funcAlphabetical(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}

	unsorted := `
- name: gamma
- name: alpha
- name: beta
`
	node2 := parseNode(t, unsorted)
	issues2 := funcAlphabetical(node2, "", opts)
	if len(issues2) == 0 {
		t.Fatal("expected issue for unsorted keyedBy")
	}
}

// ---------------------------------------------------------------------------
// or
// ---------------------------------------------------------------------------

func TestFuncOr_PassesWhenOnePresent(t *testing.T) {
	node := parseNode(t, "title: hello")
	opts := map[string]interface{}{
		"properties": []interface{}{"title", "description"},
	}
	issues := funcOr(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncOr_PassesWhenBothPresent(t *testing.T) {
	node := parseNode(t, "title: hello\ndescription: world")
	opts := map[string]interface{}{
		"properties": []interface{}{"title", "description"},
	}
	issues := funcOr(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncOr_FailsWhenNonePresent(t *testing.T) {
	node := parseNode(t, "other: value")
	opts := map[string]interface{}{
		"properties": []interface{}{"title", "description"},
	}
	issues := funcOr(node, "", opts)
	if len(issues) == 0 {
		t.Fatal("expected issue when no properties present")
	}
}

func TestFuncOr_TooFewProperties(t *testing.T) {
	node := parseNode(t, "title: hello")
	opts := map[string]interface{}{
		"properties": []interface{}{"title"},
	}
	issues := funcOr(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues with <2 properties, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// xor
// ---------------------------------------------------------------------------

func TestFuncXor_PassesWithExactlyOne(t *testing.T) {
	node := parseNode(t, "title: hello")
	opts := map[string]interface{}{
		"properties": []interface{}{"title", "description"},
	}
	issues := funcXor(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncXor_FailsWithBoth(t *testing.T) {
	node := parseNode(t, "title: hello\ndescription: world")
	opts := map[string]interface{}{
		"properties": []interface{}{"title", "description"},
	}
	issues := funcXor(node, "", opts)
	if len(issues) == 0 {
		t.Fatal("expected issue when both properties present")
	}
}

func TestFuncXor_FailsWithNone(t *testing.T) {
	node := parseNode(t, "other: value")
	opts := map[string]interface{}{
		"properties": []interface{}{"title", "description"},
	}
	issues := funcXor(node, "", opts)
	if len(issues) == 0 {
		t.Fatal("expected issue when no properties present")
	}
}

func TestFuncXor_TooFewProperties(t *testing.T) {
	node := parseNode(t, "title: hello")
	opts := map[string]interface{}{
		"properties": []interface{}{"title"},
	}
	issues := funcXor(node, "", opts)
	if len(issues) != 0 {
		t.Fatalf("expected no issues with <2 properties, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// typedEnum
// ---------------------------------------------------------------------------

func TestFuncTypedEnum_MatchingTypes(t *testing.T) {
	doc := `
type: string
enum:
  - alpha
  - beta
`
	node := parseNode(t, doc)
	issues := funcTypedEnum(node, "", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got %d", len(issues))
	}
}

func TestFuncTypedEnum_MismatchedTypes(t *testing.T) {
	doc := `
type: integer
enum:
  - alpha
  - beta
`
	node := parseNode(t, doc)
	issues := funcTypedEnum(node, "", nil)
	if len(issues) == 0 {
		t.Fatal("expected issues for string values with integer type")
	}
	if len(issues) != 2 {
		t.Fatalf("expected 2 issues (one per value), got %d", len(issues))
	}
}

func TestFuncTypedEnum_NonMappingReturnsNil(t *testing.T) {
	node := parseNode(t, `hello`)
	issues := funcTypedEnum(node, "", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for non-mapping, got %d", len(issues))
	}
}

func TestFuncTypedEnum_NoEnumReturnsNil(t *testing.T) {
	node := parseNode(t, "type: string")
	issues := funcTypedEnum(node, "", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues without enum, got %d", len(issues))
	}
}

func TestFuncTypedEnum_IntegerEnumPasses(t *testing.T) {
	doc := `
type: integer
enum:
  - 1
  - 2
  - 3
`
	node := parseNode(t, doc)
	issues := funcTypedEnum(node, "", nil)
	if len(issues) != 0 {
		t.Fatalf("expected no issues for integer enum, got %d", len(issues))
	}
}

// ---------------------------------------------------------------------------
// compileSpectralRegex
// ---------------------------------------------------------------------------

func TestCompileSpectralRegex_BarePattern(t *testing.T) {
	re, err := compileSpectralRegex(`^hello`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !re.MatchString("hello world") {
		t.Error("expected match")
	}
	if re.MatchString("goodbye") {
		t.Error("unexpected match")
	}
}

func TestCompileSpectralRegex_SlashSyntaxCaseInsensitive(t *testing.T) {
	re, err := compileSpectralRegex(`/^hello/i`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !re.MatchString("Hello World") {
		t.Error("expected case-insensitive match")
	}
	if !re.MatchString("hello") {
		t.Error("expected lowercase match")
	}
}

func TestCompileSpectralRegex_SlashSyntaxCaseSensitive(t *testing.T) {
	re, err := compileSpectralRegex(`/^hello/`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !re.MatchString("hello") {
		t.Error("expected match for lowercase")
	}
	if re.MatchString("Hello") {
		t.Error("unexpected match for uppercase (should be case-sensitive)")
	}
}

func TestCompileSpectralRegex_InvalidPattern(t *testing.T) {
	_, err := compileSpectralRegex(`[invalid`)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
}

func TestCompileSpectralRegex_InvalidSlashPattern(t *testing.T) {
	_, err := compileSpectralRegex(`/[invalid/`)
	if err == nil {
		t.Fatal("expected error for invalid regex in slash syntax")
	}
}

// ---------------------------------------------------------------------------
// isFalsy helper (via funcTruthy/funcFalsy indirectly, but also test edge)
// ---------------------------------------------------------------------------

func TestIsFalsy_NilNode(t *testing.T) {
	if !isFalsy(nil) {
		t.Error("nil should be falsy")
	}
}

// ---------------------------------------------------------------------------
// BuiltinFunctions map
// ---------------------------------------------------------------------------

func TestBuiltinFunctions_ContainsExpectedKeys(t *testing.T) {
	expected := []string{
		"truthy", "falsy", "defined", "undefined",
		"pattern", "casing", "length", "enumeration",
		"alphabetical", "or", "xor", "typedEnum",
		"schema", "unreferencedReusableObject",
	}
	for _, name := range expected {
		if _, ok := BuiltinFunctions[name]; !ok {
			t.Errorf("missing built-in function %q", name)
		}
	}
}
