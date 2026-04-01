package analyzers

import (
	"fmt"
	"strings"

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
			issues = filterRedundantParentIssues(issues)
			return issuesToOAS3Diagnostics(ctx, issues)
		},
	})
}

// filterRedundantParentIssues removes generic wrapper errors (meta.ref,
// meta.group) when more specific child errors exist at the same or deeper
// JSON Pointer. This eliminates noise like "Validation failed... (see nested
// causes)" when the actual cause (e.g., "not a valid property") is also reported.
func filterRedundantParentIssues(issues []navigator.Issue) []navigator.Issue {
	redundantCodes := map[string]bool{
		"meta.ref":   true,
		"meta.group": true,
	}

	filtered := make([]navigator.Issue, 0, len(issues))
	for _, iss := range issues {
		if !redundantCodes[iss.Code] {
			filtered = append(filtered, iss)
			continue
		}
		// Keep the parent error only if no more specific error exists
		if !hasMoreSpecificIssue(iss.Pointer, iss.Code, issues) {
			filtered = append(filtered, iss)
		}
	}
	return filtered
}

// hasMoreSpecificIssue checks if any other issue exists at the same pointer
// with a non-wrapper code, or at a deeper (child) pointer.
func hasMoreSpecificIssue(parentPtr, parentCode string, issues []navigator.Issue) bool {
	prefix := parentPtr + "/"
	for _, other := range issues {
		// Child pointer: more specific location
		if strings.HasPrefix(other.Pointer, prefix) {
			return true
		}
		// Same pointer, different (non-wrapper) code
		if other.Pointer == parentPtr && other.Code != parentCode &&
			other.Code != "meta.ref" && other.Code != "meta.group" {
			return true
		}
	}
	return false
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
