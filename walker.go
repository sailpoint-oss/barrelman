package barrelman

import (
	"fmt"

	navigator "github.com/LukasParke/navigator"
)

// Visitors holds optional callback functions for each OpenAPI model element.
// Walk invokes only the non-nil visitors, handling traversal automatically.
type Visitors struct {
	Document        func(doc *navigator.Document, r *Reporter)
	Info            func(info *navigator.Info, r *Reporter)
	Path            func(path string, item *navigator.PathItem, r *Reporter)
	Operation       func(path string, method string, op *navigator.Operation, r *Reporter)
	Schema          func(name string, schema *navigator.Schema, pointer string, r *Reporter)
	RecursiveSchema func(name string, schema *navigator.Schema, pointer string, r *Reporter)
	Parameter       func(param *navigator.Parameter, r *Reporter)
	Response        func(code string, resp *navigator.Response, r *Reporter)
	Tag             func(tag *navigator.Tag, r *Reporter)
	Server          func(server *navigator.Server, r *Reporter)
	RequestBody     func(path string, method string, rb *navigator.RequestBody, r *Reporter)
	SecurityScheme  func(name string, ss *navigator.SecurityScheme, r *Reporter)
	Example         func(name string, ex *navigator.Example, r *Reporter)
	Custom          func(idx *navigator.Index, r *Reporter)
}

// Walk traverses the OpenAPI index and invokes the registered visitor
// callbacks. The traversal order is: tags, servers, paths (with nested
// operations, parameters, request bodies, responses), component schemas,
// then the custom callback.
func Walk(idx *navigator.Index, v Visitors, r *Reporter) {
	if idx == nil || idx.Document == nil {
		return
	}
	doc := idx.Document

	if v.Document != nil {
		v.Document(doc, r)
	}

	if v.Info != nil && doc.Info != nil {
		v.Info(doc.Info, r)
	}

	if v.Tag != nil {
		for i := range doc.Tags {
			v.Tag(&doc.Tags[i], r)
		}
	}

	if v.Server != nil {
		for i := range doc.Servers {
			v.Server(&doc.Servers[i], r)
		}
	}

	walkPaths(doc, idx, v, r)
	walkComponentSchemas(doc, idx, v, r)
	walkSecuritySchemes(doc, v, r)
	walkExamples(doc, v, r)

	if v.Custom != nil {
		v.Custom(idx, r)
	}
}

func walkPaths(doc *navigator.Document, idx *navigator.Index, v Visitors, r *Reporter) {
	needPaths := v.Path != nil || v.Operation != nil || v.Parameter != nil ||
		v.RequestBody != nil || v.Response != nil

	if !needPaths {
		return
	}

	for path, item := range doc.Paths {
		if v.Path != nil {
			v.Path(path, item, r)
		}

		for _, mo := range item.Operations() {
			if v.Operation != nil {
				v.Operation(path, mo.Method, mo.Operation, r)
			}

			if v.Parameter != nil {
				for _, p := range mo.Operation.Parameters {
					if p.Ref != "" {
						continue
					}
					v.Parameter(p, r)
				}
			}

			if v.RequestBody != nil && mo.Operation.RequestBody != nil {
				if mo.Operation.RequestBody.Ref == "" {
					v.RequestBody(path, mo.Method, mo.Operation.RequestBody, r)
				}
			}

			if v.Response != nil {
				for code, resp := range mo.Operation.Responses {
					if resp.Ref != "" {
						continue
					}
					v.Response(code, resp, r)
				}
			}
		}

		if v.Parameter != nil {
			for _, p := range item.Parameters {
				if p.Ref != "" {
					continue
				}
				v.Parameter(p, r)
			}
		}
	}
}

func walkSecuritySchemes(doc *navigator.Document, v Visitors, r *Reporter) {
	if v.SecurityScheme == nil || doc.Components == nil {
		return
	}
	for name, ss := range doc.Components.SecuritySchemes {
		v.SecurityScheme(name, ss, r)
	}
}

func walkExamples(doc *navigator.Document, v Visitors, r *Reporter) {
	if v.Example == nil || doc.Components == nil {
		return
	}
	for name, ex := range doc.Components.Examples {
		v.Example(name, ex, r)
	}
}

func walkComponentSchemas(doc *navigator.Document, idx *navigator.Index, v Visitors, r *Reporter) {
	hasFlat := v.Schema != nil
	hasRecursive := v.RecursiveSchema != nil
	if !hasFlat && !hasRecursive {
		return
	}

	visitSchema := func(name string, schema *navigator.Schema, pointer string) {
		if hasFlat {
			v.Schema(name, schema, pointer, r)
		}
		if hasRecursive {
			visited := make(map[*navigator.Schema]bool)
			walkSchemaRecursive(schema, name, pointer, v.RecursiveSchema, r, 0, visited)
		}
	}

	if doc.Components != nil {
		for name, schema := range doc.Components.Schemas {
			visitSchema(name, schema, "components/schemas/"+name)
		}
	}

	for path, item := range doc.Paths {
		for _, mo := range item.Operations() {
			for _, p := range mo.Operation.Parameters {
				if p.Schema != nil {
					visitSchema("", p.Schema, fmt.Sprintf("paths%s/%s/parameters/%s", path, mo.Method, p.Name))
				}
			}
			if mo.Operation.RequestBody != nil {
				for mt, media := range mo.Operation.RequestBody.Content {
					if media.Schema != nil {
						visitSchema("", media.Schema, fmt.Sprintf("paths%s/%s/requestBody/%s", path, mo.Method, mt))
					}
				}
			}
			for code, resp := range mo.Operation.Responses {
				for mt, media := range resp.Content {
					if media.Schema != nil {
						visitSchema("", media.Schema, fmt.Sprintf("paths%s/%s/responses/%s/%s", path, mo.Method, code, mt))
					}
				}
			}
		}
	}
}

const maxWalkDepth = 64

func walkSchemaRecursive(schema *navigator.Schema, name, pointer string, fn func(string, *navigator.Schema, string, *Reporter), r *Reporter, depth int, visited map[*navigator.Schema]bool) {
	if schema == nil || depth > maxWalkDepth || visited[schema] {
		return
	}
	visited[schema] = true
	fn(name, schema, pointer, r)

	for propName, propSchema := range schema.Properties {
		walkSchemaRecursive(propSchema, propName, pointer+"/properties/"+propName, fn, r, depth+1, visited)
	}
	if schema.Items != nil {
		walkSchemaRecursive(schema.Items, "", pointer+"/items", fn, r, depth+1, visited)
	}
	if schema.AdditionalProperties != nil {
		walkSchemaRecursive(schema.AdditionalProperties, "", pointer+"/additionalProperties", fn, r, depth+1, visited)
	}
	if schema.UnevaluatedProperties != nil {
		walkSchemaRecursive(schema.UnevaluatedProperties, "", pointer+"/unevaluatedProperties", fn, r, depth+1, visited)
	}
	for i, sub := range schema.AllOf {
		walkSchemaRecursive(sub, "", fmt.Sprintf("%s/allOf/%d", pointer, i), fn, r, depth+1, visited)
	}
	for i, sub := range schema.AnyOf {
		walkSchemaRecursive(sub, "", fmt.Sprintf("%s/anyOf/%d", pointer, i), fn, r, depth+1, visited)
	}
	for i, sub := range schema.OneOf {
		walkSchemaRecursive(sub, "", fmt.Sprintf("%s/oneOf/%d", pointer, i), fn, r, depth+1, visited)
	}
	if schema.Not != nil {
		walkSchemaRecursive(schema.Not, "", pointer+"/not", fn, r, depth+1, visited)
	}
}
