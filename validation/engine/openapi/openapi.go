// Package openapi wires OpenAPI-aware validation on top of the
// validation/engine. Use this package to compile OpenAPI component
// schemas, parameter schemas, and request/response/example instances.
//
// The package layers four adapters over the core engine:
//
//  1. Nullable rewrite — OAS 3.0 `nullable: true` becomes an `anyOf` with
//     a `type: null` alternative (engine handles this when
//     CompileOpts.OpenAPI.NullableRewrite is true; this package sets the
//     flag for callers).
//  2. Discriminator helper — inspects schemas with a `discriminator`
//     keyword and reshapes errors so "oneOf failed" issues are rewritten
//     as "discriminator variant 'X' failed: ..." with a Zod-style path.
//  3. x- extension keywords — declares known SailPoint extensions so
//     "unknown keyword" warnings do not fire for x-sailpoint-*.
//  4. Format registry — registers the OpenAPI-flavored formats (int32,
//     int64, float, double, byte, binary, password, date, date-time,
//     uuid). Most of these are a superset of JSON Schema formats; the
//     registry enforces them when CompileOpts.StrictFormats is true.
package openapi

import (
	"fmt"

	"github.com/sailpoint-oss/barrelman/validation/engine"
)

// CompileSchema compiles a single OpenAPI schema (a map[string]any, raw
// JSON bytes, or any JSON-marshalable value) for use with Validate or
// ValidateExample.
//
// id is the identity the underlying compiler registers the schema at and
// must be unique within a compiler. For barrelman consumers this is
// typically the navigator pointer, for example "components/schemas/User".
func CompileSchema(id string, schema any) (*engine.Schema, error) {
	if id == "" {
		id = "mem://openapi-schema"
	} else if id[0] != '/' && id[0:1] != "h" {
		id = "mem://" + id
	}
	return engine.Compile(schema, engine.CompileOpts{
		URL: id,
		OpenAPI: engine.OpenAPIOptions{
			NullableRewrite:      true,
			DiscriminatorEnforce: true,
			RecognizeExtensions:  true,
		},
	})
}

// ValidateExample validates a single example value against a compiled
// schema and tags every returned Issue with Source=SourceExample so the
// LSP/reporter layers can filter example-originated issues from
// structural OpenAPI validation issues at the same location.
func ValidateExample(schema *engine.Schema, example any) []engine.Issue {
	issues := schema.Validate(example)
	for i := range issues {
		issues[i].Source = engine.SourceExample
		if issues[i].Data == nil {
			issues[i].Data = map[string]any{}
		}
		issues[i].Data["originatesFrom"] = "example"
	}
	return issues
}

// ValidateComponent validates a map of component examples against a
// compiled schema and returns a flat list of Issues keyed in the Data
// map by the example name.
func ValidateComponent(schema *engine.Schema, examples map[string]any) []engine.Issue {
	var out []engine.Issue
	for name, example := range examples {
		issues := ValidateExample(schema, example)
		for i := range issues {
			if issues[i].Data == nil {
				issues[i].Data = map[string]any{}
			}
			issues[i].Data["exampleName"] = name
			issues[i].Message = fmt.Sprintf("example %q: %s", name, issues[i].Message)
		}
		out = append(out, issues...)
	}
	return out
}
