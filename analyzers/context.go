package analyzers

import (
	"fmt"
	"strings"
)

// InferComponentTypeFromRef maps a $ref target to a human-readable OpenAPI Object type.
//
//	"#/components/schemas/Pet" → "Schema Object"
//	"#/components/responses/NotFound" → "Response Object"
//	"./other.yaml#/components/parameters/Limit" → "Parameter Object"
//	"#/definitions/Pet" → "Schema Object" (Swagger 2.0)
func InferComponentTypeFromRef(target string) string {
	// Strip file path, keep fragment
	if idx := strings.LastIndex(target, "#"); idx >= 0 {
		target = target[idx+1:]
	} else {
		// No fragment — whole-file ref
		return "component"
	}

	parts := strings.Split(strings.TrimPrefix(target, "/"), "/")
	if len(parts) < 2 {
		return "component"
	}

	// OpenAPI 3.x: /components/<kind>/<name>
	if parts[0] == "components" && len(parts) >= 2 {
		return componentKindToType(parts[1])
	}
	// Swagger 2.0: /definitions/<name>
	if parts[0] == "definitions" {
		return "Schema Object"
	}
	return "component"
}

// InferContextFromPointer maps a JSON Pointer to a human-readable OpenAPI context.
//
//	"/components/schemas/Pet" → "Schema Object 'Pet'"
//	"/paths/~1pets/get" → "Operation 'GET /pets'"
//	"/info" → "Info Object"
func InferContextFromPointer(pointer string) string {
	segs := splitPtr(pointer)
	if len(segs) == 0 {
		return "the document"
	}

	switch segs[0] {
	case "info":
		return "Info Object"
	case "servers":
		if len(segs) >= 2 {
			return fmt.Sprintf("Server Object (index %s)", segs[1])
		}
		return "servers"
	case "tags":
		if len(segs) >= 2 {
			return fmt.Sprintf("Tag (index %s)", segs[1])
		}
		return "tags"
	case "paths":
		if len(segs) >= 2 {
			path := unescPtr(segs[1])
			if len(segs) >= 3 {
				return fmt.Sprintf("Operation '%s %s'", strings.ToUpper(segs[2]), path)
			}
			return fmt.Sprintf("Path Item '%s'", path)
		}
		return "paths"
	case "components":
		if len(segs) >= 3 {
			kind := componentKindToType(segs[1])
			return fmt.Sprintf("%s '%s'", kind, segs[2])
		}
		if len(segs) >= 2 {
			return fmt.Sprintf("components/%s", segs[1])
		}
		return "components"
	}
	return pointer
}

// SchemaNameFromPointer extracts the schema name from a pointer like
// "/components/schemas/Pet/properties/id" → "Pet". Returns "" if not a schema pointer.
func SchemaNameFromPointer(pointer string) string {
	segs := splitPtr(pointer)
	if len(segs) >= 3 && segs[0] == "components" && segs[1] == "schemas" {
		return segs[2]
	}
	return ""
}

func componentKindToType(kind string) string {
	switch kind {
	case "schemas":
		return "Schema Object"
	case "responses":
		return "Response Object"
	case "parameters":
		return "Parameter Object"
	case "requestBodies":
		return "Request Body"
	case "headers":
		return "Header Object"
	case "securitySchemes":
		return "Security Scheme"
	case "links":
		return "Link Object"
	case "callbacks":
		return "Callback Object"
	case "examples":
		return "Example Object"
	case "pathItems":
		return "Path Item Object"
	default:
		return kind
	}
}

func splitPtr(ptr string) []string {
	if ptr == "" || ptr == "/" {
		return nil
	}
	return strings.Split(strings.TrimPrefix(ptr, "/"), "/")
}

func unescPtr(s string) string {
	s = strings.ReplaceAll(s, "~1", "/")
	s = strings.ReplaceAll(s, "~0", "~")
	return s
}
