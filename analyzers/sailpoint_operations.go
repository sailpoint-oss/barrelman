package analyzers

import (
	"strings"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/codemod/fixes"
	navigator "github.com/sailpoint-oss/navigator"
)

// #124 - Path parameters must declare x-sailpoint-resource-operation-id
// referencing an existing operationId.
func registerSailpointPathParamResourceOperationLink(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-path-param-resource-operation-link",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Path parameters must declare x-sailpoint-resource-operation-id referencing an existing operationId.",
		"Add x-sailpoint-resource-operation-id to every path parameter using a valid lowerCamelCase operationId.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		operationIDs := make(map[string]bool)
		for _, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				if strings.TrimSpace(mo.Operation.OperationID) != "" {
					operationIDs[mo.Operation.OperationID] = true
				}
			}
		}

		seen := make(map[string]bool)
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					resolved := resolveParameter(idx, param)
					if resolved == nil || resolved.In != "path" {
						continue
					}
					key := pathParameterRuleKey(param, resolved, path)
					if seen[key] {
						continue
					}
					seen[key] = true

					ext := resolved.Extensions["x-sailpoint-resource-operation-id"]
					if ext == nil || strings.TrimSpace(ext.Value) == "" {
						r.At(navigator.LocOrFallback(resolved.NameLoc, resolved.Loc),
							"Path parameter '%s' in %s must declare x-sailpoint-resource-operation-id", resolved.Name, path)
						continue
					}
					value := strings.TrimSpace(ext.Value)
					if !isLowerCamelCase(value) {
						r.At(ext.Loc,
							"x-sailpoint-resource-operation-id '%s' for path parameter '%s' in %s must use lowerCamelCase", value, resolved.Name, path)
					}
					if !operationIDs[value] {
						r.At(ext.Loc,
							"x-sailpoint-resource-operation-id '%s' for path parameter '%s' in %s must reference an existing operationId", value, resolved.Name, path)
					}
				}
			}
		}
	}).Register(reg)
}

// #602 (split 1/2) - Collection GET operations must use limit/offset pagination.
func registerSailpointCollectionOffsetPagination(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-collection-offset-pagination",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Collection GET operations must declare limit and offset query parameters.",
		"Add limit and offset query parameters to collection operations.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				if strings.ToUpper(mo.Method) != "GET" {
					continue
				}
				respSchema := firstSuccessSchema(idx, mo.Operation)
				if respSchema == nil {
					continue
				}
				isCollection := respSchema.Type == "array" || (respSchema.Type == "object" && respSchema.Properties["items"] != nil)
				if !isCollection {
					continue
				}
				params := operationParameters(item, mo.Operation)
				if !hasParam(params, "query", "limit") || !hasParam(params, "query", "offset") {
					r.At(navigator.LocOrFallback(mo.Operation.ParametersLoc, mo.Operation.Loc),
						"Collection operation %s %s must declare query parameters 'limit' and 'offset'", strings.ToUpper(mo.Method), path)
				}
			}
		}
	}).Fix(fixes.CollectionOffsetPagination).Register(reg)
}

// #602 (split 2/2) - Collection GET operations must wrap items in an object
// response that includes limit/offset/count metadata.
func registerSailpointCollectionWrappedResponse(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-collection-wrapped-response",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Collection GET operations must wrap items in a top-level object response.",
		"Return an object containing items, limit, offset, and count instead of a top-level array.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				if strings.ToUpper(mo.Method) != "GET" {
					continue
				}
				respSchema := firstSuccessSchema(idx, mo.Operation)
				if respSchema == nil {
					continue
				}
				isCollection := respSchema.Type == "array" || (respSchema.Type == "object" && respSchema.Properties["items"] != nil)
				if !isCollection {
					continue
				}
				responsesLoc := navigator.LocOrFallback(mo.Operation.ResponsesLoc, mo.Operation.Loc)
				if respSchema.Type == "array" {
					r.At(responsesLoc, "Collection operation %s %s must wrap items in a top-level object response", strings.ToUpper(mo.Method), path)
					continue
				}
				for _, prop := range []string{"items", "limit", "offset", "count"} {
					if _, ok := respSchema.Properties[prop]; !ok {
						r.At(responsesLoc, "Collection operation %s %s response schema must define property '%s'", strings.ToUpper(mo.Method), path, prop)
					}
				}
			}
		}
	}).Register(reg)
}

// #903 (split 1/3) - Responses must declare the X-Request-Id header.
func registerSailpointXRequestIDHeader(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-x-request-id-header",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Responses must declare the X-Request-Id header.",
		"Add an X-Request-Id response header to every response definition.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					if responseHeader(resp, "X-Request-Id") != nil {
						continue
					}
					r.At(headerDiagLoc(resp), "Response %s for %s %s must declare the X-Request-Id header", code, strings.ToUpper(mo.Method), path)
				}
			}
		}
	}).Fix(fixes.XRequestIDHeader).Register(reg)
}

// #903 (split 2/3) - Components must define a shared X-Request-Id header.
func registerSailpointXRequestIDSharedComponent(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-x-request-id-shared-component",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Components must define a shared X-Request-Id header.",
		"Declare components.headers.X-Request-Id with type string and format uuid.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		if hasSharedRequestIDHeaderComponent(idx.Document) {
			return
		}
		r.AtRange(barrelman.FileStartRange, "Components must define a shared X-Request-Id header with type string and format uuid")
	}).Fix(fixes.XRequestIDSharedComponent).Register(reg)
}

// #903 (split 3/3) - X-Request-Id headers must use type string with format uuid.
func registerSailpointXRequestIDUUID(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-x-request-id-uuid",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"X-Request-Id headers must use type string with format uuid.",
		"Set type: string and format: uuid on every X-Request-Id response header.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					header := responseHeader(resp, "X-Request-Id")
					if header == nil {
						continue
					}
					if !requestIDHeaderHasUUIDFormat(idx, header) {
						r.At(headerDiagLoc(resp), "X-Request-Id header on response %s for %s %s must use type string with format uuid", code, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Fix(fixes.XRequestIDUUID).Register(reg)
}
