package fixes

import (
	"bytes"
	"fmt"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/codemod"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// XRequestIDSharedComponent adds components.headers.X-Request-Id with the
// canonical shape when absent.
func XRequestIDSharedComponent(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	headers := componentsHeadersMapping(ctx)
	if headers == nil {
		return nil, nil
	}
	if codemod.MappingHasKey(headers, ctx.Source, "X-Request-Id") {
		return nil, nil
	}
	childIndent := codemod.ChildIndent(headers, ctx.Source)
	inner := childIndent + 2
	pad := bytes.Repeat([]byte(" "), inner)
	value := fmt.Sprintf("\n%sdescription: Unique request correlation identifier.\n%sschema:\n%s  type: string\n%s  format: uuid",
		pad, pad, pad, pad)
	patch := codemod.InsertMappingPair(headers, ctx.Source, "X-Request-Id", value)
	patch.URI = ctx.URI
	patch.RuleID = diag.Code
	patch.Description = "insert shared X-Request-Id header under components.headers"
	return []codemod.Patch{patch}, nil
}

// XRequestIDHeader adds X-Request-Id to the response's headers mapping as a
// $ref to the shared component.
//
// When the response has no headers: key, inserts one containing the
// X-Request-Id reference.
func XRequestIDHeader(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	response := mappingForDiagnostic(ctx, diag)
	if response == nil {
		return nil, nil
	}
	headersVal := codemod.MappingValueNode(response, ctx.Source, "headers")
	if headersVal == nil {
		childIndent := codemod.ChildIndent(response, ctx.Source)
		inner := childIndent + 2
		pad := bytes.Repeat([]byte(" "), inner)
		value := fmt.Sprintf("\n%sX-Request-Id:\n%s  $ref: '#/components/headers/X-Request-Id'", pad, pad)
		patch := codemod.InsertMappingPair(response, ctx.Source, "headers", value)
		patch.URI = ctx.URI
		patch.RuleID = diag.Code
		patch.Description = "insert headers.X-Request-Id on response"
		return []codemod.Patch{patch}, nil
	}
	headers := codemod.EnclosingBlockMapping(headersVal)
	if headers == nil || codemod.MappingHasKey(headers, ctx.Source, "X-Request-Id") {
		return nil, nil
	}
	childIndent := codemod.ChildIndent(headers, ctx.Source)
	inner := childIndent + 2
	pad := bytes.Repeat([]byte(" "), inner)
	value := fmt.Sprintf("\n%s$ref: '#/components/headers/X-Request-Id'", pad)
	patch := codemod.InsertMappingPair(headers, ctx.Source, "X-Request-Id", value)
	patch.URI = ctx.URI
	patch.RuleID = diag.Code
	patch.Description = "insert X-Request-Id reference on response.headers"
	return []codemod.Patch{patch}, nil
}

// XRequestIDUUID inserts `format: uuid` next to `type:` when the shared
// X-Request-Id header's schema is `type: string` without a format. Does not
// touch schemas that already declare any format.
func XRequestIDUUID(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	schemaMapping := xRequestIDSchemaMapping(ctx)
	if schemaMapping == nil {
		return nil, nil
	}
	if codemod.MappingHasKey(schemaMapping, ctx.Source, "format") {
		return nil, nil
	}
	// Insert format: uuid after the type: line.
	typePair := codemod.FindMappingPair(schemaMapping, ctx.Source, "type")
	if typePair == nil {
		return nil, nil
	}
	patch := codemod.InsertMappingPairAfter(typePair, ctx.Source, "format", "uuid")
	patch.URI = ctx.URI
	patch.RuleID = diag.Code
	patch.Description = "insert format: uuid on X-Request-Id schema"
	return []codemod.Patch{patch}, nil
}

// componentsHeadersMapping mirrors componentsSchemasMapping but
// navigates to components.headers.
func componentsHeadersMapping(ctx *codemod.FixContext) *ts.Node {
	root := rootMapping(ctx)
	if root == nil {
		return nil
	}
	componentsVal := codemod.MappingValueNode(root, ctx.Source, "components")
	if componentsVal == nil {
		return nil
	}
	components := codemod.EnclosingBlockMapping(componentsVal)
	if components == nil {
		return nil
	}
	headersVal := codemod.MappingValueNode(components, ctx.Source, "headers")
	if headersVal == nil {
		return nil
	}
	return codemod.EnclosingBlockMapping(headersVal)
}

// xRequestIDSchemaMapping finds components.headers.X-Request-Id.schema
// as a block_mapping.
func xRequestIDSchemaMapping(ctx *codemod.FixContext) *ts.Node {
	headers := componentsHeadersMapping(ctx)
	if headers == nil {
		return nil
	}
	headerVal := codemod.MappingValueNode(headers, ctx.Source, "X-Request-Id")
	if headerVal == nil {
		return nil
	}
	header := codemod.EnclosingBlockMapping(headerVal)
	if header == nil {
		return nil
	}
	schemaVal := codemod.MappingValueNode(header, ctx.Source, "schema")
	if schemaVal == nil {
		return nil
	}
	return codemod.EnclosingBlockMapping(schemaVal)
}
