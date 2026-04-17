// Package examples provides an engine-backed example-against-schema
// validator for OpenAPI documents. It replaces the string-heuristic
// implementations in analyzers/examples.go and spectral/functions.go.
//
// The package is a thin adapter: it takes a navigator schema plus an
// example node (or raw example string), compiles the schema with the
// OpenAPI semantic adapters, and validates the example using the
// shared engine.
package examples

import (
	"encoding/json"
	"fmt"

	navigator "github.com/sailpoint-oss/navigator"
	"gopkg.in/yaml.v3"

	"github.com/sailpoint-oss/barrelman/validation/engine"
	"github.com/sailpoint-oss/barrelman/validation/engine/openapi"
)

// ValidateSchemaExample validates an inline schema.Example against the
// schema definition and returns engine Issues. Returns nil when the
// schema has no example or when the example could not be parsed (the
// caller is expected to emit a parse-time diagnostic separately).
func ValidateSchemaExample(schema *navigator.Schema) ([]engine.Issue, error) {
	if schema == nil || schema.Example == nil {
		return nil, nil
	}
	value, err := parseExampleValue(schema.Example.Value)
	if err != nil {
		return nil, err
	}
	compiled, err := openapi.CompileSchema("mem://inline-schema", schemaToMap(schema))
	if err != nil {
		return nil, fmt.Errorf("compile inline schema: %w", err)
	}
	return openapi.ValidateExample(compiled, value), nil
}

// ValidateMediaExamples iterates over all examples attached to a media
// type (the inline Example and the Examples map) and returns a flat list
// of Issues annotated with the example name in Issue.Data.
func ValidateMediaExamples(media *navigator.MediaType) ([]engine.Issue, error) {
	if media == nil || media.Schema == nil {
		return nil, nil
	}
	compiled, err := openapi.CompileSchema("mem://media-schema", schemaToMap(media.Schema))
	if err != nil {
		return nil, fmt.Errorf("compile media schema: %w", err)
	}
	var out []engine.Issue
	if media.Example != nil {
		value, err := parseExampleValue(media.Example.Value)
		if err == nil {
			out = append(out, openapi.ValidateExample(compiled, value)...)
		}
	}
	for name, ex := range media.Examples {
		if ex == nil || ex.Value == nil {
			continue
		}
		value, err := parseExampleValue(ex.Value.Value)
		if err != nil {
			continue
		}
		issues := openapi.ValidateExample(compiled, value)
		for i := range issues {
			if issues[i].Data == nil {
				issues[i].Data = map[string]any{}
			}
			issues[i].Data["exampleName"] = name
		}
		out = append(out, issues...)
	}
	return out, nil
}

// parseExampleValue turns the raw example text (as captured by navigator)
// into a decoded Go value. Examples may be YAML or JSON; we try JSON
// first because it round-trips the common cases verbatim, then fall
// back to YAML which tolerates unquoted strings and block scalars.
func parseExampleValue(raw string) (any, error) {
	if raw == "" {
		return nil, nil
	}
	var via any
	if err := json.Unmarshal([]byte(raw), &via); err == nil {
		return via, nil
	}
	if err := yaml.Unmarshal([]byte(raw), &via); err != nil {
		return nil, fmt.Errorf("parse example: %w", err)
	}
	return normalizeYAMLMaps(via), nil
}

// normalizeYAMLMaps converts yaml.v3's `map[any]any` into `map[string]any`
// recursively so the engine (which expects JSON-shaped instances) can
// validate without type assertions failing.
func normalizeYAMLMaps(v any) any {
	switch m := v.(type) {
	case map[any]any:
		out := make(map[string]any, len(m))
		for k, val := range m {
			out[fmt.Sprint(k)] = normalizeYAMLMaps(val)
		}
		return out
	case []any:
		for i, item := range m {
			m[i] = normalizeYAMLMaps(item)
		}
		return m
	case map[string]any:
		for k, val := range m {
			m[k] = normalizeYAMLMaps(val)
		}
		return m
	default:
		return v
	}
}

// schemaToMap converts a navigator.Schema into the map[string]any shape
// the engine expects. Only the fields meaningful for instance validation
// are carried; documentation-only fields are omitted.
func schemaToMap(s *navigator.Schema) map[string]any {
	if s == nil {
		return nil
	}
	m := map[string]any{}
	if s.Type != "" {
		m["type"] = s.Type
	}
	if s.Format != "" {
		m["format"] = s.Format
	}
	if s.Ref != "" {
		m["$ref"] = s.Ref
	}
	if s.Nullable {
		m["nullable"] = true
	}
	if len(s.Enum) > 0 {
		enum := make([]any, 0, len(s.Enum))
		for _, e := range s.Enum {
			enum = append(enum, e)
		}
		m["enum"] = enum
	}
	if len(s.Required) > 0 {
		req := make([]any, 0, len(s.Required))
		for _, r := range s.Required {
			req = append(req, r)
		}
		m["required"] = req
	}
	if len(s.Properties) > 0 {
		props := make(map[string]any, len(s.Properties))
		for name, prop := range s.Properties {
			props[name] = schemaToMap(prop)
		}
		m["properties"] = props
	}
	if s.Items != nil {
		m["items"] = schemaToMap(s.Items)
	}
	if s.MinLength != nil {
		m["minLength"] = *s.MinLength
	}
	if s.MaxLength != nil {
		m["maxLength"] = *s.MaxLength
	}
	if s.Pattern != "" {
		m["pattern"] = s.Pattern
	}
	return m
}
