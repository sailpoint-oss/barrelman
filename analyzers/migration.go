package analyzers

import (
	navigator "github.com/LukasParke/navigator"
	"github.com/sailpoint-oss/barrelman"
)

var (
	migrationNullableMeta = barrelman.RuleMeta{
		ID:          "migration-nullable",
		Description: "In OpenAPI 3.1, use type array ['string', 'null'] instead of nullable: true.",
		Severity:    barrelman.SeverityInfo,
		Category:    barrelman.CategoryTypes,
		Recommended: false,
		Formats:     []navigator.Format{navigator.Format(navigator.Version30)},
		HowToFix:    "When migrating to 3.1, replace `nullable: true` with `type: ['string', 'null']`.",
		DocURL:      barrelman.DocBaseURL + "migration-nullable",
	}
)

func registerMigrationAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("migration-nullable", migrationNullableMeta).
		Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
			if idx.Version != navigator.Version30 {
				return
			}
			if idx.Document.Components != nil {
				for _, schema := range idx.Document.Components.Schemas {
					checkNullable(schema, r)
				}
			}
			for _, item := range idx.Document.Paths {
				for _, mo := range item.Operations() {
					for _, p := range mo.Operation.Parameters {
						if p.Schema != nil {
							checkNullable(p.Schema, r)
						}
					}
					if mo.Operation.RequestBody != nil {
						for _, mt := range mo.Operation.RequestBody.Content {
							if mt.Schema != nil {
								checkNullable(mt.Schema, r)
							}
						}
					}
					for _, resp := range mo.Operation.Responses {
						for _, mt := range resp.Content {
							if mt.Schema != nil {
								checkNullable(mt.Schema, r)
							}
						}
					}
				}
			}
		}).
		Register(reg)
}

func checkNullable(schema *navigator.Schema, r *barrelman.Reporter) {
	if schema.Nullable && schema.Type != "" {
		r.At(schema.Loc,
			"OpenAPI 3.1 migration: replace `nullable: true` with `type: ['%s', 'null']`",
			schema.Type,
		)
	}
	for _, prop := range schema.Properties {
		checkNullable(prop, r)
	}
	if schema.Items != nil {
		checkNullable(schema.Items, r)
	}
}
