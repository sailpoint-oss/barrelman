package analyzers

import (
	"strconv"
	"strings"

	navigator "github.com/LukasParke/navigator"
	"github.com/sailpoint-oss/barrelman"
)

var (
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
