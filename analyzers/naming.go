package analyzers

import (
	"sort"

	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

// naming.go registers generic naming rules that are not backed by a
// specific SailPoint guideline. Rules that previously lived here but now
// belong to sailpoint-operation-id-* live in analyzers/sailpoint_naming.go.

var (
	schemaNameCapitalMeta = barrelman.RuleMeta{
		ID:          "schema-name-capital",
		Description: "Schema names should start with an uppercase letter.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryNaming,
		Recommended: true,
		HowToFix:    "Rename the schema to start with a capital letter (e.g., 'pet' → 'Pet').",
		DocURL:      barrelman.DocBaseURL + "schema-name-capital",
	}

	exampleNameCapitalMeta = barrelman.RuleMeta{
		ID:          "example-name-capital",
		Description: "Example names should start with an uppercase letter.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryNaming,
		Recommended: true,
		HowToFix:    "Rename the example to start with a capital letter.",
		DocURL:      barrelman.DocBaseURL + "example-name-capital",
	}

	tagsFormatMeta = barrelman.RuleMeta{
		ID:          "tags-format",
		Description: "Tags should follow a consistent naming format.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryNaming,
		Recommended: true,
		HowToFix:    "Use a consistent casing style for tag names.",
		DocURL:      barrelman.DocBaseURL + "tags-format",
	}
)

func registerNamingAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("schema-name-capital", schemaNameCapitalMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if name != "" && !isCapitalized(name) {
				r.At(schema.NameLoc, "Schema name '%s' should start with an uppercase letter", name)
			}
		},
	).Register(reg)

	barrelman.Define("example-name-capital", exampleNameCapitalMeta).Examples(
		func(name string, ex *navigator.Example, r *barrelman.Reporter) {
			if !isCapitalized(name) {
				r.At(ex.Loc, "Example name '%s' should start with an uppercase letter", name)
			}
		},
	).Register(reg)

	barrelman.Define("tags-format", tagsFormatMeta).Tags(
		func(tag *navigator.Tag, r *barrelman.Reporter) {
			if tag.Name == "" {
				r.At(tag.Loc, "Tag name should not be empty")
			}
		},
	).Register(reg)
}

func sortedPathKeys(paths map[string]*navigator.PathItem) []string {
	keys := make([]string, 0, len(paths))
	for path := range paths {
		keys = append(keys, path)
	}
	sort.Strings(keys)
	return keys
}
