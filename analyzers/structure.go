package analyzers

import (
	navigator "github.com/sailpoint-oss/navigator"
	"github.com/sailpoint-oss/barrelman"
)

var (
	additionalPropertiesMeta = barrelman.RuleMeta{
		ID:          "additional-properties",
		Description: "Object schemas should define additionalProperties explicitly.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryStructure,
		Recommended: true,
		HowToFix:    "Add 'additionalProperties: false' or define allowed additional properties.",
		DocURL:      barrelman.DocBaseURL + "additional-properties",
	}

	allOfMixedTypesMeta = barrelman.RuleMeta{
		ID:          "allof-mixed-types",
		Description: "allOf should not combine schemas of different types.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryStructure,
		Recommended: true,
		HowToFix:    "Ensure all schemas in allOf have compatible types.",
		DocURL:      barrelman.DocBaseURL + "allof-mixed-types",
	}

	allOfStructureMeta = barrelman.RuleMeta{
		ID:          "allof-structure",
		Description: "allOf schemas must be structurally valid.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryStructure,
		Recommended: true,
		HowToFix:    "Review the allOf composition for structural issues.",
		DocURL:      barrelman.DocBaseURL + "allof-structure",
	}

	arrayItemsMeta = barrelman.RuleMeta{
		ID:          "array-items",
		Description: "Array schemas must define items.",
		Severity:    barrelman.SeverityError,
		Category:    barrelman.CategoryStructure,
		Recommended: true,
		HowToFix:    "Add an 'items' definition to the array schema.",
		DocURL:      barrelman.DocBaseURL + "array-items",
	}

	discriminatorMappingMeta = barrelman.RuleMeta{
		ID:          "discriminator-mapping",
		Description: "Discriminator mapping values must reference valid schemas.",
		Severity:    barrelman.SeverityError,
		Category:    barrelman.CategoryStructure,
		Recommended: true,
		HowToFix:    "Ensure each discriminator mapping value references an existing schema.",
		DocURL:      barrelman.DocBaseURL + "discriminator-mapping",
	}

	requestBodyContentMeta = barrelman.RuleMeta{
		ID:          "request-body-content",
		Description: "Request bodies must have content defined.",
		Severity:    barrelman.SeverityError,
		Category:    barrelman.CategoryStructure,
		Recommended: true,
		HowToFix:    "Add a 'content' section to the request body.",
		DocURL:      barrelman.DocBaseURL + "request-body-content",
	}

	typeRequiredMeta = barrelman.RuleMeta{
		ID:          "type-required",
		Description: "Schemas should have a 'type' field defined.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryStructure,
		Recommended: true,
		HowToFix:    "Add the 'type' field to the schema.",
		DocURL:      barrelman.DocBaseURL + "type-required",
	}
)

func registerStructureAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("additional-properties", additionalPropertiesMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if name != "" && schema.Type == "object" && len(schema.Properties) > 0 && schema.AdditionalProperties == nil {
				r.At(schema.Loc, "Schema '%s' should define additionalProperties", name)
			}
		},
	).Register(reg)

	barrelman.Define("allof-mixed-types", allOfMixedTypesMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if name == "" || len(schema.AllOf) < 2 {
				return
			}
			var types []string
			for _, sub := range schema.AllOf {
				if sub.Type != "" {
					types = append(types, sub.Type)
				}
			}
			if len(types) >= 2 {
				first := types[0]
				for _, t := range types[1:] {
					if t != first {
						r.At(schema.Loc, "Schema '%s' allOf mixes types: %s and %s", name, first, t)
						break
					}
				}
			}
		},
	).Register(reg)

	barrelman.Define("allof-structure", allOfStructureMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if name != "" && len(schema.AllOf) == 1 && schema.AllOf[0].Ref == "" {
				r.At(schema.Loc, "Schema '%s' uses allOf with a single non-$ref item; consider inlining", name)
			}
		},
	).Register(reg)

	barrelman.Define("array-items", arrayItemsMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if schema.Type == "array" && schema.Items == nil && schema.Ref == "" {
				r.At(schema.Loc, "Array schema at %s must define 'items'", pointer)
			}
		},
	).Register(reg)

	barrelman.Define("discriminator-mapping", discriminatorMappingMeta).Custom(
		func(idx *navigator.Index, r *barrelman.Reporter) {
			if idx.Document.Components == nil {
				return
			}
			for name, schema := range idx.Document.Components.Schemas {
				if schema.Discriminator == nil || schema.Discriminator.Mapping == nil {
					continue
				}
				for key, ref := range schema.Discriminator.Mapping {
					if _, err := idx.Resolve(ref); err != nil {
						r.At(schema.Discriminator.Loc, "Discriminator mapping '%s' in '%s' references unresolvable: %s", key, name, ref)
					}
				}
			}
		},
	).Register(reg)

	barrelman.Define("request-body-content", requestBodyContentMeta).RequestBodies(
		func(path, method string, rb *navigator.RequestBody, r *barrelman.Reporter) {
			if rb.Ref == "" && len(rb.Content) == 0 {
				r.At(rb.Loc, "Request body for %s %s must define content", method, path)
			}
		},
	).Register(reg)

	barrelman.Define("type-required", typeRequiredMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if name != "" && schema.Type == "" && schema.Ref == "" &&
				len(schema.AllOf) == 0 && len(schema.AnyOf) == 0 && len(schema.OneOf) == 0 {
				r.At(schema.Loc, "Schema '%s' should define a 'type'", name)
			}
		},
	).Register(reg)
}
