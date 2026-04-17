package analyzers

import (
	"strings"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/codemod/fixes"
	navigator "github.com/sailpoint-oss/navigator"
)

// #104 - JSON object property names must use lowerCamelCase.
func registerSailpointPropertyCamelCase(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-property-camel-case",
		barrelman.SeverityError,
		barrelman.CategoryNaming,
		"JSON object property names must use lowerCamelCase.",
		"Rename schema properties to lowerCamelCase.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		walkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
			for propName, prop := range schema.Properties {
				if isExtensionName(propName) || isLowerCamelCase(propName) {
					continue
				}
				loc := propertyLoc(prop, schema)
				if sn := SchemaNameFromPointer(pointer); sn != "" {
					r.At(loc, "Property '%s' in Schema Object '%s' must use lowerCamelCase", propName, sn)
				} else {
					r.At(loc, "Property '%s' at %s must use lowerCamelCase", propName, pointerForProperty(pointer, propName))
				}
			}
		})
	}).Register(reg)
}

// #107 (split 1/2) - Path segments must use lowercase, hyphen-separated form.
func registerSailpointPathKebabCase(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-path-kebab-case",
		barrelman.SeverityError,
		barrelman.CategoryPaths,
		"Path segments must use lowercase hyphenated form.",
		"Use lowercase hyphenated resource segments in path templates.",
	)

	barrelman.Define(meta.ID, meta).Paths(func(path string, item *navigator.PathItem, r *barrelman.Reporter) {
		for _, seg := range barrelman.NonParamSegments(path) {
			if barrelman.IsKebabCase(seg) {
				continue
			}
			r.At(item.PathLoc, "Path segment '%s' in %s must use lowercase hyphenated form", seg, path)
		}
	}).Register(reg)
}

// #107 (split 2/2) - Path parameters must use lowerCamelCase.
func registerSailpointPathParamCamelCase(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-path-param-camel-case",
		barrelman.SeverityError,
		barrelman.CategoryPaths,
		"Path parameters must use lowerCamelCase.",
		"Rename path parameters to lowerCamelCase.",
	)

	barrelman.Define(meta.ID, meta).Paths(func(path string, item *navigator.PathItem, r *barrelman.Reporter) {
		for _, param := range barrelman.ExtractPathParams(path) {
			if isLowerCamelCase(param) {
				continue
			}
			r.At(item.PathLoc, "Path parameter '{%s}' in %s must use lowerCamelCase", param, path)
		}
	}).Register(reg)
}

// #108 - Query parameter names must use lowerCamelCase.
func registerSailpointQueryParamCamelCase(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-query-param-camel-case",
		barrelman.SeverityError,
		barrelman.CategoryPaths,
		"Query parameter names must use lowerCamelCase.",
		"Rename query parameters to lowerCamelCase.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					if param == nil || param.Ref != "" || param.In != "query" || isLowerCamelCase(param.Name) {
						continue
					}
					r.At(param.NameLoc, "Query parameter '%s' in %s %s must use lowerCamelCase", param.Name, strings.ToUpper(mo.Method), path)
				}
			}
		}
	}).Register(reg)
}

// #112 - String enum values must use UPPER_SNAKE_CASE.
func registerSailpointEnumScreamingSnakeCase(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-enum-screaming-snake-case",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"String enum values must use UPPER_SNAKE_CASE.",
		"Rename string enum values to UPPER_SNAKE_CASE.",
	)

	barrelman.Define(meta.ID, meta).RecursiveSchemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
		if schema.Type != "string" {
			return
		}
		for _, value := range schema.Enum {
			if upperSnakeRe.MatchString(value) {
				continue
			}
			if sn := SchemaNameFromPointer(pointer); sn != "" {
				r.At(schema.Loc, "Enum value '%s' in Schema Object '%s' must use UPPER_SNAKE_CASE", value, sn)
			} else {
				r.At(schema.Loc, "Enum value '%s' at %s must use UPPER_SNAKE_CASE", value, pointer)
			}
			return
		}
	}).Register(reg)
}

// #122 (split 1/2) - Operations must declare a lowerCamelCase operationId.
func registerSailpointOperationIDCamelCase(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-operation-id-camel-case",
		barrelman.SeverityError,
		barrelman.CategoryNaming,
		"Operations must declare a lowerCamelCase operationId.",
		"Add an operationId to every operation and use lowerCamelCase.",
	)

	barrelman.Define(meta.ID, meta).Operations(func(path, method string, op *navigator.Operation, r *barrelman.Reporter) {
		upper := strings.ToUpper(method)
		if strings.TrimSpace(op.OperationID) == "" {
			r.At(op.Loc, "Operation %s %s must declare an operationId", upper, path)
			return
		}
		if !isLowerCamelCase(op.OperationID) {
			r.At(op.OperationIDLoc, "operationId '%s' for %s %s must use lowerCamelCase", op.OperationID, upper, path)
		}
	}).Register(reg)
}

// #122 (split 2/2) - operationIds must be unique across the API.
func registerSailpointOperationIDUnique(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-operation-id-unique",
		barrelman.SeverityError,
		barrelman.CategoryNaming,
		"operationIds must be unique across the API.",
		"Give each operation a unique operationId.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		type location struct {
			loc  navigator.Loc
			desc string
		}
		seen := make(map[string]location)
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				op := mo.Operation
				if strings.TrimSpace(op.OperationID) == "" {
					continue
				}
				method := strings.ToUpper(mo.Method)
				here := method + " " + path
				if first, ok := seen[op.OperationID]; ok {
					r.WithRelated(first.loc, "", "First defined here at %s", first.desc).
						At(op.OperationIDLoc, "operationId '%s' is already used at %s", op.OperationID, first.desc)
					continue
				}
				seen[op.OperationID] = location{loc: op.OperationIDLoc, desc: here}
			}
		}
	}).Register(reg)
}

// #123 (split 1/2) - Each operation must declare exactly one tag.
func registerSailpointOperationSingleTag(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-operation-single-tag",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Each operation must declare exactly one tag.",
		"Declare exactly one root tag on each operation.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		rootTags := make(map[string]bool, len(idx.Document.Tags))
		for _, tag := range idx.Document.Tags {
			rootTags[tag.Name] = true
		}
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				op := mo.Operation
				method := strings.ToUpper(mo.Method)
				if len(op.Tags) != 1 {
					r.At(navigator.LocOrFallback(op.TagsLoc, op.Loc), "Operation %s %s must declare exactly one tag", method, path)
					continue
				}
				tagName := op.Tags[0].Name
				if !rootTags[tagName] {
					r.At(op.Tags[0].Loc, "Operation tag '%s' on %s %s must be declared in the root tags section", tagName, method, path)
				}
			}
		}
	}).Fix(fixes.OperationSingleTag).Register(reg)
}

// #123 (split 2/2) - Every root tag must include a description.
func registerSailpointTagDocumented(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-tag-documented",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Every root tag must include a description.",
		"Add a description to every root tag.",
	)

	barrelman.Define(meta.ID, meta).Tags(func(tag *navigator.Tag, r *barrelman.Reporter) {
		if strings.TrimSpace(tag.Description.Text) == "" {
			r.At(tag.Loc, "Root tag '%s' must include a description", tag.Name)
		}
	}).Fix(fixes.TagDocumented).Register(reg)
}

// #115 (split 1/2) - Parameters must include descriptions.
func registerSailpointParameterDescription(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-parameter-description",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Parameters must include descriptions.",
		"Add a description to every parameter.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					if param == nil || param.Ref != "" || strings.TrimSpace(param.Description.Text) != "" {
						continue
					}
					r.At(param.Loc, "Parameter '%s' in %s %s must include a description", param.Name, strings.ToUpper(mo.Method), path)
				}
			}
		}
	}).Fix(fixes.ParameterDescription).Register(reg)
}

// #115 (split 2/2) - Schema properties must include descriptions.
func registerSailpointPropertyDescription(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-property-description",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Schema properties must include descriptions.",
		"Add a description to every schema property.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		walkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
			for propName, prop := range schema.Properties {
				if prop == nil || prop.Ref != "" || strings.TrimSpace(prop.Description.Text) != "" {
					continue
				}
				loc := propertyLoc(prop, schema)
				if sn := SchemaNameFromPointer(pointer); sn != "" {
					r.At(loc, "Property '%s' in Schema Object '%s' must include a description", propName, sn)
				} else {
					r.At(loc, "Property '%s' at %s must include a description", propName, pointerForProperty(pointer, propName))
				}
			}
		})
	}).Fix(fixes.PropertyDescription).Register(reg)
}

// #116 (split 1/3) - Parameters must include examples.
func registerSailpointParameterExample(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-parameter-example",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Parameters must include examples.",
		"Add a representative example to every parameter.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					if param == nil || param.Ref != "" || hasParameterExample(param) {
						continue
					}
					r.At(param.Loc, "Parameter '%s' in %s %s must include an example", param.Name, strings.ToUpper(mo.Method), path)
				}
			}
		}
	}).Fix(fixes.ParameterExample).Register(reg)
}

// #116 (split 2/3) - Schema properties must include examples.
func registerSailpointPropertyExample(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-property-example",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Schema properties must include examples.",
		"Add a representative example to every schema property.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		walkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
			for propName, prop := range schema.Properties {
				if prop == nil || prop.Ref != "" || prop.Example != nil {
					continue
				}
				loc := propertyLoc(prop, schema)
				if sn := SchemaNameFromPointer(pointer); sn != "" {
					r.At(loc, "Property '%s' in Schema Object '%s' must include an example", propName, sn)
				} else {
					r.At(loc, "Property '%s' at %s must include an example", propName, pointerForProperty(pointer, propName))
				}
			}
		})
	}).Fix(fixes.PropertyExample).Register(reg)
}

// #116 (split 3/3) - Response payloads must include examples.
func registerSailpointResponseExample(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-response-example",
		barrelman.SeverityError,
		barrelman.CategoryDocumentation,
		"Response payloads must include examples.",
		"Add a representative example to every response media type.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					if !isErrorOrSuccessCode(code) {
						continue
					}
					for mediaType, media := range resp.Content {
						if media == nil || hasMediaExample(media) || (media.Schema != nil && media.Schema.Example != nil) {
							continue
						}
						r.At(resp.Loc, "Response %s for %s %s must include an example for %s", code, strings.ToUpper(mo.Method), path, mediaType)
					}
				}
			}
		}
	}).Fix(fixes.ResponseExample).Register(reg)
}
