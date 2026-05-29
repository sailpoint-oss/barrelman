package analyzers

import (
	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

// extended.go registers generic documentation rules.

var (
	operationDescriptionMeta = barrelman.RuleMeta{ID: "operation-description", Description: "Operations should have descriptions.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "operation-description"}
	infoDescriptionMeta      = barrelman.RuleMeta{ID: "info-description", Description: "Info should have a description.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "info-description"}
	infoContactMeta          = barrelman.RuleMeta{ID: "info-contact", Description: "Info should have contact information.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "info-contact"}
	infoLicenseMeta          = barrelman.RuleMeta{ID: "info-license", Description: "Info should have license information.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "info-license"}
	responseDescriptionMeta  = barrelman.RuleMeta{ID: "response-description", Description: "Responses should have descriptions.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "response-description"}
	schemaDescriptionMeta    = barrelman.RuleMeta{ID: "schema-description", Description: "Component schemas should have descriptions.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: false, DocURL: barrelman.DocBaseURL + "schema-description"}
)

func registerExtendedAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("operation-description", operationDescriptionMeta).Operations(
		func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			if op.Description.Text == "" {
				r.At(op.Loc, "Operation %s %s should have a description", method, path)
			}
		},
	).Register(reg)

	barrelman.Define("info-description", infoDescriptionMeta).Info(
		func(info *navigator.Info, r *barrelman.Reporter) {
			if info.Description.Text == "" {
				r.At(info.Loc, "Info should have a description")
			}
		},
	).Register(reg)

	barrelman.Define("info-contact", infoContactMeta).Info(
		func(info *navigator.Info, r *barrelman.Reporter) {
			if info.Contact == nil {
				r.At(info.Loc, "Info should have contact information")
			}
		},
	).Register(reg)

	barrelman.Define("info-license", infoLicenseMeta).Info(
		func(info *navigator.Info, r *barrelman.Reporter) {
			if info.License == nil {
				r.At(info.Loc, "Info should have license information")
			}
		},
	).Register(reg)

	barrelman.Define("response-description", responseDescriptionMeta).Operations(
		func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			for code, resp := range op.Responses {
				if resp.Description.Text == "" && resp.Ref == "" {
					r.At(resp.Loc, "Response '%s' for %s %s should have a description", code, method, path)
				}
			}
		},
	).Register(reg)

	barrelman.Define("schema-description", schemaDescriptionMeta).Schemas(
		func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
			if name != "" && schema.Description.Text == "" && schema.Ref == "" {
				r.At(schema.Loc, "Schema '%s' should have a description", name)
			}
		},
	).Register(reg)
}
