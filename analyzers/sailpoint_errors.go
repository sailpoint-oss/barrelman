package analyzers

import (
	"strings"

	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

// #403 (split 1/4) - Operations must use standard HTTP status codes appropriate to the method.
func registerSailpointOperationMethodStatusCodes(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-operation-method-status-codes",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Operations must use IANA-standard HTTP status codes appropriate to the method.",
		"Return IANA standard status codes; GET must declare 200, DELETE must declare 200 or 204, collection-creating POST should declare 201.",
	)

	barrelman.Define(meta.ID, meta).Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
		has200, has201, has204 := false, false, false
		responsesLoc := navigator.LocOrFallback(op.ResponsesLoc, op.Loc)
		for code := range op.Responses {
			if code == "default" {
				continue
			}
			if !knownStatusCodes[code] {
				r.At(responsesLoc, "Response code '%s' for %s %s is not a standard HTTP status code", code, strings.ToUpper(method), path)
			}
			switch code {
			case "200":
				has200 = true
			case "201":
				has201 = true
			case "204":
				has204 = true
			}
		}
		switch strings.ToUpper(method) {
		case "GET":
			if !has200 {
				r.At(responsesLoc, "GET %s must declare a 200 response", path)
			}
		case "DELETE":
			if !has200 && !has204 {
				r.At(responsesLoc, "DELETE %s must declare a 200 or 204 response", path)
			}
		case "POST":
			if isLikelyCollectionCreate(path) && !has201 {
				r.At(responsesLoc, "POST %s should declare a 201 response for resource creation", path)
			}
		}
	}).Register(reg)
}

// #403 (split 2/4) - Operations must declare at least one 4xx response.
func registerSailpointOperation4xxResponse(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-operation-4xx-response",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Operations must declare at least one 4xx response.",
		"Document a 4xx response describing the client error contract.",
	)

	barrelman.Define(meta.ID, meta).Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
		has4xx := false
		for code := range op.Responses {
			if strings.HasPrefix(code, "4") {
				has4xx = true
				break
			}
		}
		if !has4xx {
			r.At(navigator.LocOrFallback(op.ResponsesLoc, op.Loc),
				"Operation %s %s must declare at least one 4xx response", strings.ToUpper(method), path)
		}
	}).Register(reg)
}

// #403 (split 3/4) - Operations must declare a 401 response.
func registerSailpointOperation401Response(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-operation-401-response",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Operations must declare a 401 Unauthorized response.",
		"Document a 401 response for every operation.",
	)

	barrelman.Define(meta.ID, meta).Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
		if _, ok := op.Responses["401"]; ok {
			return
		}
		r.At(navigator.LocOrFallback(op.ResponsesLoc, op.Loc),
			"Operation %s %s must declare a 401 response", strings.ToUpper(method), path)
	}).Register(reg)
}

// #403 (split 4/4) - Operations must declare a 403 response.
func registerSailpointOperation403Response(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-operation-403-response",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Operations must declare a 403 Forbidden response.",
		"Document a 403 response for every operation.",
	)

	barrelman.Define(meta.ID, meta).Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
		if _, ok := op.Responses["403"]; ok {
			return
		}
		r.At(navigator.LocOrFallback(op.ResponsesLoc, op.Loc),
			"Operation %s %s must declare a 403 response", strings.ToUpper(method), path)
	}).Register(reg)
}

// #404 (split 1/4) - Error responses must use application/problem+json.
func registerSailpointErrorProblemDetailsMediaType(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-error-problem-details-media-type",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Error responses must use application/problem+json.",
		"Return application/problem+json on every 4xx and 5xx response.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					if !strings.HasPrefix(code, "4") && !strings.HasPrefix(code, "5") {
						continue
					}
					if _, ok := resp.Content["application/problem+json"]; !ok {
						r.At(resp.Loc, "Response %s for %s %s must use application/problem+json", code, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Register(reg)
}

// #404 (split 2/4) - Problem Details responses must declare the Problem Details fields.
func registerSailpointErrorProblemDetailsSchema(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-error-problem-details-schema",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Problem Details responses must declare the RFC 7807 fields.",
		"Ensure every Problem Details schema defines type, title, status, detail, and instance.",
	)

	requiredProps := []string{"type", "title", "status", "detail", "instance"}
	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					if !strings.HasPrefix(code, "4") && !strings.HasPrefix(code, "5") {
						continue
					}
					media := resp.Content["application/problem+json"]
					if media == nil {
						continue
					}
					if !hasMediaExample(media) {
						r.At(resp.Loc, "Problem Details response %s for %s %s must include an example", code, strings.ToUpper(mo.Method), path)
					}
					schema := resolveSchema(idx, media.Schema)
					if schema == nil {
						r.At(resp.Loc, "Problem Details response %s for %s %s must declare a schema", code, strings.ToUpper(mo.Method), path)
						continue
					}
					if schema.Type != "object" {
						r.At(resp.Loc, "Problem Details response %s for %s %s must use an object schema", code, strings.ToUpper(mo.Method), path)
						continue
					}
					for _, prop := range requiredProps {
						if _, ok := schema.Properties[prop]; !ok {
							r.At(resp.Loc, "Problem Details response %s for %s %s must define property '%s'", code, strings.ToUpper(mo.Method), path, prop)
						}
					}
				}
			}
		}
	}).Register(reg)
}

// #404 (split 3/4) - Components must define a shared ProblemDetails schema and Problem Details responses must reference it.
func registerSailpointErrorProblemDetailsSharedComponent(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-error-problem-details-shared-component",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Problem Details responses must reference a shared components.schemas.ProblemDetails schema.",
		"Define components.schemas.ProblemDetails and $ref it from every application/problem+json response.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		if !hasProblemDetailsComponent(idx.Document) {
			r.AtRange(barrelman.FileStartRange, "Components must define a shared ProblemDetails schema")
		}
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					if !strings.HasPrefix(code, "4") && !strings.HasPrefix(code, "5") {
						continue
					}
					media := resp.Content["application/problem+json"]
					if media == nil {
						continue
					}
					if media.Schema == nil || media.Schema.Ref != "#/components/schemas/ProblemDetails" {
						r.At(resp.Loc, "Problem Details response %s for %s %s must reference #/components/schemas/ProblemDetails", code, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Register(reg)
}

// #404 (split 4/4) - Problem Details responses must include a correlationId property.
func registerSailpointErrorCorrelationID(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-error-correlation-id",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Problem Details responses must declare a correlationId property.",
		"Add correlationId (string) to the shared ProblemDetails schema.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					if !strings.HasPrefix(code, "4") && !strings.HasPrefix(code, "5") {
						continue
					}
					media := resp.Content["application/problem+json"]
					if media == nil {
						continue
					}
					schema := resolveSchema(idx, media.Schema)
					if schema == nil || schema.Type != "object" {
						continue
					}
					if _, ok := schema.Properties["correlationId"]; !ok {
						r.At(resp.Loc, "Problem Details response %s for %s %s must define property 'correlationId'", code, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Register(reg)
}
