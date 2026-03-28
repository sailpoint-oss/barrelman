package analyzers

import (
	"fmt"

	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

var structuralMeta = barrelman.RuleMeta{
	ID:          "oas3-schema",
	Description: "Surfaces Navigator structural and meta-schema issues for the current API-description document.",
	Severity:    barrelman.SeverityError,
	Category:    barrelman.CategoryStructure,
	Recommended: true,
	HowToFix:    "Fix the issues reported for the current OpenAPI or Arazzo document.",
	DocURL:      barrelman.DocBaseURL + "oas3-schema",
}

const maxOAS3SchemaDiagnostics = 200

func registerStructuralValidation(reg *barrelman.Registry) {
	reg.Register(barrelman.Rule{
		ID:   "oas3-schema",
		Meta: structuralMeta,
		Run: func(ctx *barrelman.AnalysisContext) []barrelman.Diagnostic {
			if ctx == nil || ctx.Index == nil {
				return nil
			}
			issues := ctx.Index.Issues
			if ctx.TargetVersion != navigator.VersionUnknown {
				issues = ctx.Index.Revalidate(navigator.ValidationOptions{
					TargetVersion: ctx.TargetVersion,
				})
			}
			return issuesToOAS3Diagnostics(ctx, issues)
		},
	})
}

func issuesToOAS3Diagnostics(ctx *barrelman.AnalysisContext, issues []navigator.Issue) []barrelman.Diagnostic {
	if len(issues) == 0 {
		return nil
	}
	out := make([]barrelman.Diagnostic, 0, len(issues))
	for i, iss := range issues {
		if i >= maxOAS3SchemaDiagnostics {
			break
		}
		data := map[string]string{
			"issueCode": iss.Code,
		}
		if iss.Pointer != "" {
			data["pointer"] = iss.Pointer
		}
		data["category"] = issueCategoryString(iss.Category)
		data["documentKind"] = iss.DocumentKind.String()

		out = append(out, barrelman.Diagnostic{
			URI:             ctx.URI,
			Range:           iss.Range,
			Severity:        navigatorSeverityToBarrelman(iss.Severity),
			Code:            "oas3-schema",
			CodeDescription: structuralMeta.DocURL,
			Source:          barrelman.Source,
			Message:         iss.Message,
			Data:            data,
		})
	}
	return out
}

func issueCategoryString(c navigator.IssueCategory) string {
	switch c {
	case navigator.CategorySyntax:
		return "syntax"
	case navigator.CategoryStructural:
		return "structural"
	case navigator.CategorySchema:
		return "schema"
	case navigator.CategoryMeta:
		return "meta"
	default:
		return fmt.Sprintf("category_%d", int(c))
	}
}

func navigatorSeverityToBarrelman(s navigator.Severity) barrelman.Severity {
	switch s {
	case navigator.SeverityWarning:
		return barrelman.SeverityWarning
	case navigator.SeverityInfo:
		return barrelman.SeverityInfo
	default:
		return barrelman.SeverityError
	}
}
