package analyzers

import (
	"fmt"
	"strconv"
	"strings"

	navigator "github.com/sailpoint-oss/navigator"
	"github.com/sailpoint-oss/barrelman"
)

var (
	oas3ValidMediaExampleMeta = barrelman.RuleMeta{
		ID:          "oas3-valid-media-example",
		Description: "Examples in media type objects must conform to their associated schema.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryTypes,
		Recommended: false,
		HowToFix:    "Update the example to match the schema type, enum, and required constraints.",
		DocURL:      barrelman.DocBaseURL + "oas3-valid-media-example",
	}

	oas3ValidSchemaExampleMeta = barrelman.RuleMeta{
		ID:          "oas3-valid-schema-example",
		Description: "Example values in schema objects must conform to the schema definition.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryTypes,
		Recommended: false,
		HowToFix:    "Update the example to match the schema type, enum, and required constraints.",
		DocURL:      barrelman.DocBaseURL + "oas3-valid-schema-example",
	}

	exampleTypeMismatchMeta = barrelman.RuleMeta{
		ID:          "example-type-mismatch",
		Description: "Example values should match the declared schema type.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryTypes,
		Recommended: true,
		HowToFix:    "Update the example value to match the schema type.",
		DocURL:      barrelman.DocBaseURL + "example-type-mismatch",
	}

	exampleEnumMismatchMeta = barrelman.RuleMeta{
		ID:          "example-enum-mismatch",
		Description: "Example values should be one of the declared enum values.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryTypes,
		Recommended: true,
		HowToFix:    "Use one of the declared enum values in the example.",
		DocURL:      barrelman.DocBaseURL + "example-enum-mismatch",
	}
)

func registerExampleValidationAnalyzers(reg *barrelman.Registry) {
	registerOAS3ValidMediaExample(reg)
	registerOAS3ValidSchemaExample(reg)

	barrelman.Define("example-type-mismatch", exampleTypeMismatchMeta).
		RecursiveSchemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.Example == nil || schema.Type == "" {
				return
			}
			val := strings.TrimSpace(schema.Example.Value)
			if val == "" {
				return
			}
			if !valueMatchesType(val, schema.Type) {
				r.At(schema.Example.Loc, "Example value '%s' does not match schema type '%s'", truncateVal(val), schema.Type)
			}
		}).
		Register(reg)

	barrelman.Define("example-enum-mismatch", exampleEnumMismatchMeta).
		RecursiveSchemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.Example == nil || len(schema.Enum) == 0 {
				return
			}
			val := strings.Trim(strings.TrimSpace(schema.Example.Value), "\"'")
			found := false
			for _, e := range schema.Enum {
				if e == val {
					found = true
					break
				}
			}
			if !found {
				r.At(schema.Example.Loc, "Example value '%s' is not in enum [%s]", truncateVal(val), strings.Join(schema.Enum, ", "))
			}
		}).
		Register(reg)
}

func valueMatchesType(val, schemaType string) bool {
	unquoted := strings.Trim(val, "\"'")

	switch schemaType {
	case "string":
		return true
	case "integer":
		_, err := strconv.ParseInt(unquoted, 10, 64)
		return err == nil
	case "number":
		_, err := strconv.ParseFloat(unquoted, 64)
		return err == nil
	case "boolean":
		lower := strings.ToLower(unquoted)
		return lower == "true" || lower == "false"
	case "array":
		return strings.HasPrefix(val, "[") || strings.HasPrefix(val, "-")
	case "object":
		return strings.HasPrefix(val, "{") || strings.Contains(val, ":")
	default:
		return true
	}
}

func truncateVal(v string) string {
	if len(v) > 40 {
		return v[:37] + "..."
	}
	return v
}

func registerOAS3ValidMediaExample(reg *barrelman.Registry) {
	barrelman.Define("oas3-valid-media-example", oas3ValidMediaExampleMeta).
		Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
			doc := idx.Document
			for path, item := range doc.Paths {
				for _, mo := range item.Operations() {
					if mo.Operation.RequestBody != nil && mo.Operation.RequestBody.Ref == "" {
						for mt, media := range mo.Operation.RequestBody.Content {
							validateMediaExample(media, fmt.Sprintf("%s %s requestBody[%s]", mo.Method, path, mt), r)
						}
					}
					for code, resp := range mo.Operation.Responses {
						if resp.Ref != "" {
							continue
						}
						for mt, media := range resp.Content {
							validateMediaExample(media, fmt.Sprintf("%s %s response %s[%s]", mo.Method, path, code, mt), r)
						}
					}
				}
			}
		}).
		Register(reg)
}

func validateMediaExample(media *navigator.MediaType, context string, r *barrelman.Reporter) {
	if media.Schema == nil {
		return
	}

	if media.Example != nil {
		validateNodeAgainstSchema(media.Example, media.Schema, context, r)
	}

	for name, ex := range media.Examples {
		if ex.Value != nil {
			validateNodeAgainstSchema(ex.Value, media.Schema, fmt.Sprintf("%s example '%s'", context, name), r)
		}
	}
}

func registerOAS3ValidSchemaExample(reg *barrelman.Registry) {
	barrelman.Define("oas3-valid-schema-example", oas3ValidSchemaExampleMeta).
		RecursiveSchemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.Example == nil || schema.Type == "" {
				return
			}
			val := strings.TrimSpace(schema.Example.Value)
			if val == "" {
				return
			}

			if !valueMatchesType(val, schema.Type) {
				r.At(schema.Example.Loc, "Schema example value '%s' does not match type '%s' at %s", truncateVal(val), schema.Type, pointer)
				return
			}

			unquoted := strings.Trim(val, "\"'")
			if len(schema.Enum) > 0 {
				found := false
				for _, e := range schema.Enum {
					if e == unquoted {
						found = true
						break
					}
				}
				if !found {
					r.At(schema.Example.Loc, "Schema example value '%s' is not in enum [%s] at %s", truncateVal(unquoted), strings.Join(schema.Enum, ", "), pointer)
					return
				}
			}

			if schema.Type == "object" {
				validateObjectExample(val, schema, pointer, r, schema.Example.Loc)
			}
		}).
		Register(reg)
}

func validateNodeAgainstSchema(node *navigator.Node, schema *navigator.Schema, context string, r *barrelman.Reporter) {
	if schema.Type == "" {
		return
	}
	val := strings.TrimSpace(node.Value)
	if val == "" {
		return
	}

	if !valueMatchesType(val, schema.Type) {
		r.At(node.Loc, "Example in %s has type mismatch: got '%s', expected '%s'", context, truncateVal(val), schema.Type)
		return
	}

	unquoted := strings.Trim(val, "\"'")
	if len(schema.Enum) > 0 {
		found := false
		for _, e := range schema.Enum {
			if e == unquoted {
				found = true
				break
			}
		}
		if !found {
			r.At(node.Loc, "Example in %s has value '%s' not in enum [%s]", context, truncateVal(unquoted), strings.Join(schema.Enum, ", "))
		}
	}
}

func validateObjectExample(val string, schema *navigator.Schema, pointer string, r *barrelman.Reporter, loc navigator.Loc) {
	if len(schema.Required) == 0 {
		return
	}
	for _, req := range schema.Required {
		if !strings.Contains(val, req) {
			r.At(loc, "Schema example at %s is missing required property '%s'", pointer, req)
		}
	}
}
