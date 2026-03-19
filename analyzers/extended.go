package analyzers

import (
	navigator "github.com/sailpoint-oss/navigator"
	"github.com/sailpoint-oss/barrelman"
)

var (
	operationDescriptionMeta = barrelman.RuleMeta{ID: "operation-description", Description: "Operations should have descriptions.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "operation-description"}
	operationTagsMeta        = barrelman.RuleMeta{ID: "operation-tags", Description: "Operations should have at least one tag.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "operation-tags"}
	operationOperationIDMeta = barrelman.RuleMeta{ID: "operation-operationId", Description: "Operations should have operationId.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "operation-operationId"}
	infoDescriptionMeta      = barrelman.RuleMeta{ID: "info-description", Description: "Info should have a description.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "info-description"}
	infoContactMeta          = barrelman.RuleMeta{ID: "info-contact", Description: "Info should have contact information.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "info-contact"}
	infoLicenseMeta          = barrelman.RuleMeta{ID: "info-license", Description: "Info should have license information.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "info-license"}
	tagDescriptionMeta       = barrelman.RuleMeta{ID: "tag-description", Description: "Tags should have descriptions.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: true, DocURL: barrelman.DocBaseURL + "tag-description"}
	parameterDescriptionMeta = barrelman.RuleMeta{ID: "parameter-description", Description: "Parameters should have descriptions.", Severity: barrelman.SeverityWarning, Category: barrelman.CategoryDocumentation, Recommended: false, DocURL: barrelman.DocBaseURL + "parameter-description"}
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

	barrelman.Define("operation-tags", operationTagsMeta).Operations(
		func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			if len(op.Tags) == 0 {
				r.At(navigator.LocOrFallback(op.TagsLoc, op.Loc), "Operation %s %s should have at least one tag", method, path)
			}
		},
	).Register(reg)

	barrelman.Define("operation-operationId", operationOperationIDMeta).Operations(
		func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			if op.OperationID == "" {
				r.At(op.Loc, "Operation %s %s should have an operationId", method, path)
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

	barrelman.Define("tag-description", tagDescriptionMeta).Tags(
		func(tag *navigator.Tag, r *barrelman.Reporter) {
			if tag.Description.Text == "" {
				r.At(tag.Loc, "Tag '%s' should have a description", tag.Name)
			}
		},
	).Register(reg)

	barrelman.Define("parameter-description", parameterDescriptionMeta).Operations(
		func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			for _, p := range op.Parameters {
				if p.Description.Text == "" && p.Ref == "" {
					r.At(p.Loc, "Parameter '%s' in %s %s should have a description", p.Name, method, path)
				}
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
