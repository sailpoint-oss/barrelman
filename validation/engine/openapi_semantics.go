package engine

import (
	"fmt"
	"strings"

	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	jsonschemakind "github.com/santhosh-tekuri/jsonschema/v6/kind"
)

// OpenAPIOptions configures OpenAPI-specific semantic adapters applied at
// compile time. The adapters rewrite the source schema so santhosh v6 can
// validate OpenAPI-flavored schemas out of the box while preserving the
// rewrites in error messages where possible.
type OpenAPIOptions struct {
	// NullableRewrite converts `"nullable": true` on OAS 3.0 schemas into
	// the JSON Schema equivalent (`anyOf: [{type: null}, ...]`).
	// Required for validating instances drawn from OAS 3.0 documents;
	// harmless for OAS 3.1 schemas which use `type: [T, "null"]` natively.
	NullableRewrite bool

	// DiscriminatorEnforce, when true, makes the engine emit an issue
	// when a discriminator value does not resolve to a known mapping
	// entry.
	DiscriminatorEnforce bool

	// RecognizeExtensions, when true, silences "unknown keyword" style
	// diagnostics for `x-` extensions (the OpenAPI spec allows them
	// everywhere).
	RecognizeExtensions bool
}

// rewriteOpenAPINullable walks the schema (map[string]any) and replaces
// every occurrence of `"nullable": true` alongside a `"type"` keyword
// with an equivalent `"anyOf": [{"type": "null"}, {...original type}]`
// tree. No-op for non-map inputs (bool schemas and scalars).
func rewriteOpenAPINullable(schema any) any {
	return rewriteNullableNode(schema)
}

func rewriteNullableNode(node any) any {
	m, ok := node.(map[string]any)
	if !ok {
		// Arrays, scalars, bool schemas pass through untouched.
		if arr, ok := node.([]any); ok {
			for i, item := range arr {
				arr[i] = rewriteNullableNode(item)
			}
			return arr
		}
		return node
	}
	for k, v := range m {
		m[k] = rewriteNullableNode(v)
	}
	nullable, present := m["nullable"].(bool)
	if !present || !nullable {
		delete(m, "nullable")
		return m
	}
	delete(m, "nullable")
	if _, hasAnyOf := m["anyOf"]; hasAnyOf {
		// Already expressed as composition — leave alone, just ensure
		// a type: null alternative is present.
		return ensureNullAlternative(m, "anyOf")
	}
	if _, hasOneOf := m["oneOf"]; hasOneOf {
		return ensureNullAlternative(m, "oneOf")
	}
	// Wrap the existing declaration in anyOf.
	wrapped := map[string]any{
		"anyOf": []any{
			map[string]any{"type": "null"},
			m,
		},
	}
	return wrapped
}

func ensureNullAlternative(schema map[string]any, key string) map[string]any {
	alternatives, _ := schema[key].([]any)
	for _, alt := range alternatives {
		if mm, ok := alt.(map[string]any); ok && mm["type"] == "null" {
			return schema
		}
	}
	alternatives = append([]any{map[string]any{"type": "null"}}, alternatives...)
	schema[key] = alternatives
	return schema
}

// expectedReceived derives Zod-style expected/received snapshots from a
// santhosh ValidationError kind when possible. Returns ok=false when the
// error kind does not carry enough structure to produce a useful pair.
func expectedReceived(ve *jsonschema.ValidationError, instance any) (string, string, bool) {
	if ve == nil {
		return "", "", false
	}
	switch k := ve.ErrorKind.(type) {
	case *jsonschemakind.Type:
		expected := strings.Join(normalizeTypes(k.Want), " | ")
		received := describeValueAtPointer(instance, instancePointer(ve))
		return expected, received, true
	case *jsonschemakind.Enum:
		expected := renderEnum(k.Want)
		received := describeValueAtPointer(instance, instancePointer(ve))
		return expected, received, true
	case *jsonschemakind.Const:
		expected := fmt.Sprintf("%v", k.Want)
		received := describeValueAtPointer(instance, instancePointer(ve))
		return expected, received, true
	}
	return "", "", false
}

func normalizeTypes(types []string) []string {
	out := make([]string, 0, len(types))
	for _, t := range types {
		out = append(out, t)
	}
	return out
}

func renderEnum(values []any) string {
	parts := make([]string, 0, len(values))
	for _, v := range values {
		parts = append(parts, fmt.Sprintf("%v", v))
	}
	return strings.Join(parts, " | ")
}

// describeValueAtPointer renders the instance value at the given pointer
// as a short human string (type + truncated literal for scalars). Used
// to populate Issue.Received.
func describeValueAtPointer(root any, pointer string) string {
	v := resolvePointer(root, pointer)
	return describeValue(v)
}

func describeValue(v any) string {
	switch val := v.(type) {
	case nil:
		return "null"
	case string:
		if len(val) > 32 {
			return fmt.Sprintf("string %q…", val[:32])
		}
		return fmt.Sprintf("string %q", val)
	case bool:
		return fmt.Sprintf("boolean %v", val)
	case float64, int, int32, int64:
		return fmt.Sprintf("number %v", val)
	case []any:
		return fmt.Sprintf("array of length %d", len(val))
	case map[string]any:
		return fmt.Sprintf("object with %d keys", len(val))
	default:
		return fmt.Sprintf("%T", v)
	}
}

// resolvePointer walks an RFC 6901 JSON pointer into a generic decoded
// instance (the shape produced by json.Unmarshal into interface{}).
// Returns nil for unresolvable pointers.
func resolvePointer(root any, pointer string) any {
	if pointer == "" || pointer == "/" {
		return root
	}
	segs := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	cur := root
	for _, seg := range segs {
		seg = strings.ReplaceAll(seg, "~1", "/")
		seg = strings.ReplaceAll(seg, "~0", "~")
		switch c := cur.(type) {
		case map[string]any:
			next, ok := c[seg]
			if !ok {
				return nil
			}
			cur = next
		case []any:
			idx, err := atoiSafe(seg)
			if err != nil || idx < 0 || idx >= len(c) {
				return nil
			}
			cur = c[idx]
		default:
			return nil
		}
	}
	return cur
}

func atoiSafe(s string) (int, error) {
	if s == "" {
		return 0, fmt.Errorf("empty index")
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, fmt.Errorf("invalid index %q", s)
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}
