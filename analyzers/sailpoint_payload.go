package analyzers

import (
	"strings"

	"github.com/sailpoint-oss/barrelman"
	navigator "github.com/sailpoint-oss/navigator"
)

// #204 - Successful JSON responses must return top-level objects, not arrays.
func registerSailpointResponseTopLevelObject(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-response-top-level-object",
		barrelman.SeverityError,
		barrelman.CategoryStructure,
		"Successful JSON responses must return top-level objects, not arrays.",
		"Wrap collections in an object envelope instead of returning a top-level array.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for code, resp := range mo.Operation.Responses {
					if !strings.HasPrefix(code, "2") {
						continue
					}
					for mediaType, media := range resp.Content {
						if media == nil || !strings.Contains(strings.ToLower(mediaType), "json") {
							continue
						}
						schema := resolveSchema(idx, media.Schema)
						if schema == nil || schema.Type != "array" {
							continue
						}
						r.At(resp.Loc, "Response %s for %s %s must return a top-level object instead of an array", code, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Register(reg)
}

// #500 - Paths must not include an /api base prefix.
func registerSailpointPathNoAPIPrefix(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-path-no-api-prefix",
		barrelman.SeverityError,
		barrelman.CategoryPaths,
		"Paths must not include an /api base prefix.",
		"Publish resource paths directly instead of nesting them under /api.",
	)

	barrelman.Define(meta.ID, meta).Paths(func(path string, item *navigator.PathItem, r *barrelman.Reporter) {
		if path == "/api" || strings.HasPrefix(path, "/api/") {
			r.At(item.PathLoc, "Path '%s' must not include an /api base prefix", path)
		}
	}).Register(reg)
}

// #514 - Path parameters must not use numeric identifier schemas.
func registerSailpointPathParamNoNumericID(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-path-param-no-numeric-id",
		barrelman.SeverityError,
		barrelman.CategoryPaths,
		"Path parameters must not use numeric identifier schemas.",
		"Model path identifiers as strings rather than integers or numbers.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					if param == nil || param.Ref != "" || param.In != "path" || param.Schema == nil {
						continue
					}
					schema := resolveSchema(idx, param.Schema)
					if schema == nil {
						continue
					}
					if schema.Type == "integer" || schema.Type == "number" {
						r.At(param.NameLoc, "Path parameter '%s' in %s %s must use a string identifier schema", param.Name, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Register(reg)
}

// #701 - Boolean schemas must not be nullable.
func registerSailpointBooleanNotNullable(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-boolean-not-nullable",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Boolean schemas must not be nullable.",
		"Represent booleans as true/false only; remove nullable from boolean schemas.",
	)

	barrelman.Define(meta.ID, meta).RecursiveSchemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
		if schema.Type == "boolean" && schema.Nullable {
			if sn := SchemaNameFromPointer(pointer); sn != "" {
				r.At(schema.Loc, "Boolean property in Schema Object '%s' must not be nullable", sn)
			} else {
				r.At(schema.Loc, "Boolean schema at %s must not be nullable", pointer)
			}
		}
	}).Register(reg)
}

// #702 - Array schemas must not be nullable.
func registerSailpointArrayNotNullable(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-array-not-nullable",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Array schemas must not be nullable.",
		"Use empty arrays instead of nullable arrays.",
	)

	barrelman.Define(meta.ID, meta).RecursiveSchemas(func(name string, schema *navigator.Schema, pointer string, r *barrelman.Reporter) {
		if schema.Type == "array" && schema.Nullable {
			if sn := SchemaNameFromPointer(pointer); sn != "" {
				r.At(schema.Loc, "Array property in Schema Object '%s' must not be nullable", sn)
			} else {
				r.At(schema.Loc, "Array schema at %s must not be nullable", pointer)
			}
		}
	}).Register(reg)
}

// #710 (split 1/2) - Request and response object schemas must declare a required array.
func registerSailpointSchemaRequiredFields(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-schema-required-fields",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Request and response object schemas must declare a required array.",
		"Declare required arrays on object schemas used by request bodies and responses.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				if mo.Operation.RequestBody != nil {
					for mediaType, media := range mo.Operation.RequestBody.Content {
						schema := resolveSchema(idx, media.Schema)
						if !schemaNeedsRequiredArray(schema) {
							continue
						}
						r.At(mo.Operation.RequestBody.Loc,
							"Request body schema for %s %s (%s) must declare a required array",
							strings.ToUpper(mo.Method), path, mediaType)
					}
				}
				for code, resp := range mo.Operation.Responses {
					for mediaType, media := range resp.Content {
						schema := resolveSchema(idx, media.Schema)
						if !schemaNeedsRequiredArray(schema) {
							continue
						}
						r.At(resp.Loc,
							"Response schema for %s %s response %s (%s) must declare a required array",
							strings.ToUpper(mo.Method), path, code, mediaType)
					}
				}
			}
		}
	}).Register(reg)
}

// #710 (split 2/2) - Path parameters must be required.
func registerSailpointPathParamRequired(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-path-param-required",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Path parameters must be marked required.",
		"Set required: true on every path parameter.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		for path, item := range idx.Document.Paths {
			for _, mo := range item.Operations() {
				for _, param := range operationParameters(item, mo.Operation) {
					if param == nil || param.Ref != "" || param.In != "path" {
						continue
					}
					if !param.Required {
						r.At(param.NameLoc, "Path parameter '%s' in %s %s must set required: true", param.Name, strings.ToUpper(mo.Method), path)
					}
				}
			}
		}
	}).Register(reg)
}

// #804 (split 1/2) - Numeric schemas must declare approved formats.
func registerSailpointNumericFormatApproved(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-numeric-format-approved",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Numeric schemas must use approved numeric formats.",
		"Use int32/int64 for integers and float/double for numbers.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		walkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
			switch schema.Type {
			case "integer":
				if schema.Format != "int32" && schema.Format != "int64" {
					if sn := SchemaNameFromPointer(pointer); sn != "" {
						r.At(schema.Loc, "Integer property in Schema Object '%s' must declare format int32 or int64", sn)
					} else {
						r.At(schema.Loc, "Integer schema at %s must declare format int32 or int64", pointer)
					}
				}
			case "number":
				if schema.Format != "float" && schema.Format != "double" {
					if sn := SchemaNameFromPointer(pointer); sn != "" {
						r.At(schema.Loc, "Number property in Schema Object '%s' must declare format float or double", sn)
					} else {
						r.At(schema.Loc, "Number schema at %s must declare format float or double", pointer)
					}
				}
			}
		})
	}).Register(reg)
}

// #804 (split 2/2) - Identifier fields must use type string.
func registerSailpointIdentifierStringType(reg *barrelman.Registry) {
	meta := sailpointMeta(
		"sailpoint-identifier-string-type",
		barrelman.SeverityError,
		barrelman.CategoryTypes,
		"Identifier fields must use type string.",
		"Model identifier fields (name ending in 'Id' or '_id') as strings.",
	)

	barrelman.Define(meta.ID, meta).Custom(func(idx *navigator.Index, r *barrelman.Reporter) {
		walkAllSchemas(idx, func(name string, schema *navigator.Schema, pointer string) {
			schemaName := nameFromPointer(pointer, name)
			if looksLikeID(schemaName) && (schema.Type == "integer" || schema.Type == "number") {
				if sn := SchemaNameFromPointer(pointer); sn != "" {
					r.At(schema.Loc, "Identifier schema in Schema Object '%s' must use type string instead of %s", sn, schema.Type)
				} else {
					r.At(schema.Loc, "Identifier schema at %s must use type string instead of %s", pointer, schema.Type)
				}
			}
			for propName, prop := range schema.Properties {
				if prop == nil {
					continue
				}
				if looksLikeID(propName) && (prop.Type == "integer" || prop.Type == "number") {
					if sn := SchemaNameFromPointer(pointer); sn != "" {
						r.At(propertyLoc(prop, schema), "Identifier property '%s' in Schema Object '%s' must use type string", propName, sn)
					} else {
						r.At(propertyLoc(prop, schema), "Identifier property '%s' at %s must use type string", propName, pointerForProperty(pointer, propName))
					}
				}
			}
		})
	}).Register(reg)
}
