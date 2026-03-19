package analyzers

import (
	"fmt"
	"strings"

	navigator "github.com/sailpoint-oss/navigator"
	"github.com/sailpoint-oss/barrelman"
)

var (
	deprecatedDescriptionMeta = barrelman.RuleMeta{
		ID:          "deprecated-description",
		Description: "Deprecated items should include a description explaining the deprecation.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryDocumentation,
		Recommended: true,
		HowToFix:    "Add a description to the deprecated item explaining why it is deprecated and what to use instead.",
		DocURL:      barrelman.DocBaseURL + "deprecated-description",
	}

	enumDescriptionMeta = barrelman.RuleMeta{
		ID:          "enum-description",
		Description: "Enum schemas should include a description.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryDocumentation,
		Recommended: true,
		HowToFix:    "Add a description explaining the enum values.",
		DocURL:      barrelman.DocBaseURL + "enum-description",
	}

	deprecatedOperationMeta = barrelman.RuleMeta{
		ID:          "deprecated-operation",
		Description: "Deprecated operations are marked with strikethrough in the IDE.",
		Severity:    barrelman.SeverityHint,
		Category:    barrelman.CategoryDocumentation,
		Recommended: true,
		DocURL:      barrelman.DocBaseURL + "deprecated-operation",
	}

	deprecatedSchemaMeta = barrelman.RuleMeta{
		ID:          "deprecated-schema",
		Description: "Deprecated schemas are marked with strikethrough in the IDE.",
		Severity:    barrelman.SeverityHint,
		Category:    barrelman.CategoryDocumentation,
		Recommended: true,
		DocURL:      barrelman.DocBaseURL + "deprecated-schema",
	}

	deprecatedRefUsageMeta = barrelman.RuleMeta{
		ID:          "deprecated-ref-usage",
		Description: "References to deprecated components are flagged.",
		Severity:    barrelman.SeverityInfo,
		Category:    barrelman.CategoryDocumentation,
		Recommended: true,
		HowToFix:    "Consider migrating to a non-deprecated alternative.",
		DocURL:      barrelman.DocBaseURL + "deprecated-ref-usage",
	}
)

func registerDocumentationAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("deprecated-description", deprecatedDescriptionMeta).
		Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			if op.Deprecated && op.Description.Text == "" {
				r.At(op.Loc, "Deprecated operation %s %s should have a description", method, path)
			}
		}).
		Schemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if name != "" && schema.Deprecated && schema.Description.Text == "" {
				r.At(schema.Loc, "Deprecated schema '%s' should have a description", name)
			}
		}).
		Register(reg)

	barrelman.Define("enum-description", enumDescriptionMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if name != "" && len(schema.Enum) > 0 && schema.Description.Text == "" {
				r.At(schema.Loc, "Enum schema '%s' should have a description", name)
			}
		},
	).Register(reg)

	barrelman.Define("deprecated-operation", deprecatedOperationMeta).
		Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			if op.Deprecated {
				r.WithTags(barrelman.DiagnosticTagDeprecated).
					At(op.Loc, "Operation %s %s is deprecated", strings.ToUpper(method), path)
			}
		}).
		Register(reg)

	barrelman.Define("deprecated-schema", deprecatedSchemaMeta).
		Schemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if name != "" && schema.Deprecated {
				r.WithTags(barrelman.DiagnosticTagDeprecated).
					At(schema.NameLoc, "Schema '%s' is deprecated", name)
			}
		}).
		Register(reg)

	barrelman.Define("deprecated-ref-usage", deprecatedRefUsageMeta).
		Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
			for target, usages := range idx.Refs {
				resolved, err := idx.Resolve(target)
				if err != nil {
					continue
				}

				var isDeprecated bool
				var replacement string
				switch t := resolved.(type) {
				case *navigator.Schema:
					isDeprecated = t.Deprecated
					if ext, ok := t.Extensions["x-telescope-replacement"]; ok {
						replacement = ext.Value
					}
				case *navigator.Parameter:
					isDeprecated = t.Deprecated
				}

				if !isDeprecated {
					continue
				}

				for _, usage := range usages {
					msg := fmt.Sprintf("References deprecated component '%s'", refBaseName(target))
					if replacement != "" {
						msg += fmt.Sprintf(". Consider using '%s' instead", replacement)
					}
					r.WithTags(barrelman.DiagnosticTagDeprecated).At(usage.Loc, "%s", msg)
				}
			}
		}).
		Register(reg)
}

func refBaseName(ref string) string {
	parts := strings.Split(ref, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ref
}
