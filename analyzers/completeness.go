package analyzers

import (
	"fmt"
	"strings"

	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

var (
	contactPropertiesMeta = barrelman.RuleMeta{
		ID:          "contact-properties",
		Description: "Contact object should have name, url, and email.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryDocumentation,
		Recommended: false,
		HowToFix:    "Add name, url, and email properties to the info.contact object.",
		DocURL:      barrelman.DocBaseURL + "contact-properties",
	}

	licenseURLMeta = barrelman.RuleMeta{
		ID:          "license-url",
		Description: "License object should include a url.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryDocumentation,
		Recommended: false,
		HowToFix:    "Add a 'url' property to the info.license object.",
		DocURL:      barrelman.DocBaseURL + "license-url",
	}

	missingErrorResponsesMeta = barrelman.RuleMeta{
		ID:          "missing-error-responses",
		Description: "Operations should define at least one error response (4xx or 5xx).",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryStructure,
		Recommended: false,
		HowToFix:    "Add error response definitions (e.g., 400, 404, 500) to the operation.",
		DocURL:      barrelman.GuidelineDocURL("403"),
	}

	responseBodyOnDeleteMeta = barrelman.RuleMeta{
		ID:          "response-body-on-delete",
		Description: "DELETE operations typically should not return a response body.",
		Severity:    barrelman.SeverityInfo,
		Category:    barrelman.CategoryStructure,
		Recommended: false,
		HowToFix:    "Use a 204 No Content response for DELETE operations.",
		DocURL:      barrelman.DocBaseURL + "response-body-on-delete",
	}

	requestBodyOnGetMeta = barrelman.RuleMeta{
		ID:          "no-request-body-on-get",
		Description: "GET and HEAD operations should not have request bodies.",
		Severity:    barrelman.SeverityWarning,
		Category:    barrelman.CategoryStructure,
		Recommended: true,
		HowToFix:    "Remove the request body from the GET/HEAD operation. Use query parameters instead.",
		DocURL:      barrelman.DocBaseURL + "no-request-body-on-get",
	}

	missingPaginationMeta = barrelman.RuleMeta{
		ID:          "missing-pagination",
		Description: "List endpoints returning arrays should include pagination parameters.",
		Severity:    barrelman.SeverityInfo,
		Category:    barrelman.CategoryStructure,
		Recommended: false,
		HowToFix:    "Add pagination query parameters (e.g., page, pageSize, limit, offset).",
		DocURL:      barrelman.GuidelineDocURL("602"),
	}

	inconsistentErrorShapeMeta = barrelman.RuleMeta{
		ID:          "inconsistent-error-shape",
		Description: "Error responses should use a consistent schema across operations.",
		Severity:    barrelman.SeverityInfo,
		Category:    barrelman.CategoryStructure,
		Recommended: false,
		HowToFix:    "Define a shared error schema in components and reference it in all error responses.",
		DocURL:      barrelman.DocBaseURL + "inconsistent-error-shape",
	}
)

func registerCompletenessAnalyzers(reg *barrelman.Registry) {
	barrelman.Define("missing-error-responses", missingErrorResponsesMeta).
		Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			if len(op.Responses) == 0 {
				return
			}
			hasError := false
			for code := range op.Responses {
				if strings.HasPrefix(code, "4") || strings.HasPrefix(code, "5") || code == "default" {
					hasError = true
					break
				}
			}
			if !hasError {
				r.At(navigator.LocOrFallback(op.ResponsesLoc, op.Loc), "Operation %s %s has no error responses (4xx/5xx)", strings.ToUpper(method), path)
			}
		}).
		Register(reg)

	barrelman.Define("response-body-on-delete", responseBodyOnDeleteMeta).
		Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			if strings.ToUpper(method) != "DELETE" {
				return
			}
			for code, resp := range op.Responses {
				if code == "204" || strings.HasPrefix(code, "4") || strings.HasPrefix(code, "5") || code == "default" {
					continue
				}
				if len(resp.Content) > 0 {
					r.At(resp.Loc, "DELETE %s response %s has a response body; consider using 204 No Content", path, code)
				}
			}
		}).
		Register(reg)

	barrelman.Define("no-request-body-on-get", requestBodyOnGetMeta).
		Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			m := strings.ToUpper(method)
			if m != "GET" && m != "HEAD" {
				return
			}
			if op.RequestBody != nil {
				r.At(op.RequestBody.Loc, "%s %s should not have a request body", m, path)
			}
		}).
		Register(reg)

	barrelman.Define("missing-pagination", missingPaginationMeta).
		Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
			if strings.ToUpper(method) != "GET" {
				return
			}
			returnsArray := false
			for code, resp := range op.Responses {
				if !strings.HasPrefix(code, "2") {
					continue
				}
				for _, mt := range resp.Content {
					if mt.Schema != nil && mt.Schema.Type == "array" {
						returnsArray = true
						break
					}
				}
			}
			if !returnsArray {
				return
			}
			paginationNames := map[string]bool{
				"page": true, "pagesize": true, "page_size": true,
				"limit": true, "offset": true, "cursor": true,
				"after": true, "before": true, "per_page": true,
			}
			for _, p := range op.Parameters {
				if paginationNames[strings.ToLower(p.Name)] {
					return
				}
			}
			r.At(navigator.LocOrFallback(op.ParametersLoc, op.Loc), "GET %s returns an array but has no pagination parameters", path)
		}).
		Register(reg)

	barrelman.Define("contact-properties", contactPropertiesMeta).Info(
		func(info *navigator.Info, r *barrelman.Reporter) {
			if info.Contact == nil {
				return
			}
			c := info.Contact
			if c.Name == "" {
				r.At(c.Loc, "Contact object should have a 'name' property")
			}
			if c.URL == "" {
				r.At(c.Loc, "Contact object should have a 'url' property")
			}
			if c.Email == "" {
				r.At(c.Loc, "Contact object should have an 'email' property")
			}
		},
	).Register(reg)

	barrelman.Define("license-url", licenseURLMeta).Info(
		func(info *navigator.Info, r *barrelman.Reporter) {
			if info.License == nil {
				return
			}
			if info.License.URL == "" {
				r.At(info.License.Loc, "License object should have a 'url' property")
			}
		},
	).Register(reg)

	barrelman.Define("inconsistent-error-shape", inconsistentErrorShapeMeta).
		Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
			type errorInfo struct {
				ref    string
				opDesc string
				loc    navigator.Loc
			}
			var errorSchemas []errorInfo

			for path, item := range idx.Document.Paths {
				for _, mo := range item.Operations() {
					for code, resp := range mo.Operation.Responses {
						if !strings.HasPrefix(code, "4") && !strings.HasPrefix(code, "5") {
							continue
						}
						for _, mt := range resp.Content {
							if mt.Schema != nil {
								ref := mt.Schema.Ref
								if ref == "" {
									ref = fmt.Sprintf("inline(%s)", mt.Schema.Type)
								}
								errorSchemas = append(errorSchemas, errorInfo{
									ref:    ref,
									opDesc: fmt.Sprintf("%s %s (%s)", strings.ToUpper(mo.Method), path, code),
									loc:    resp.Loc,
								})
							}
						}
					}
				}
			}

			if len(errorSchemas) < 2 {
				return
			}

			firstRef := errorSchemas[0].ref
			for _, info := range errorSchemas[1:] {
				if info.ref != firstRef {
					r.At(info.loc, "Error response in %s uses a different schema than other error responses", info.opDesc)
					return
				}
			}
		}).
		Register(reg)
}
