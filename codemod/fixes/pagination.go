package fixes

import (
	"bytes"
	"fmt"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/codemod"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// CollectionOffsetPagination implements the Fix for
// sailpoint-collection-offset-pagination: appends `limit` and
// `offset` query parameters to the operation's parameters list.
//
// When `parameters:` is absent, inserts it with both parameters.
// Emits zero patches when both limit and offset are already
// declared.
func CollectionOffsetPagination(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	opMapping := operationMappingFor(ctx, diag)
	if opMapping == nil {
		return nil, nil
	}
	// Fresh parameters: include both entries.
	paramsVal := codemod.MappingValueNode(opMapping, ctx.Source, "parameters")
	if paramsVal == nil {
		indent := codemod.ChildIndent(opMapping, ctx.Source)
		inner := indent + 2
		pad := bytes.Repeat([]byte(" "), inner)
		value := fmt.Sprintf("\n%s- %s\n%s- %s", pad, paginationLimitEntry(inner+2), pad, paginationOffsetEntry(inner+2))
		patch := codemod.InsertMappingPair(opMapping, ctx.Source, "parameters", value)
		patch.URI = ctx.URI
		patch.RuleID = diag.Code
		patch.Description = "insert parameters: limit + offset"
		return []codemod.Patch{patch}, nil
	}
	// parameters: exists; append only the missing entries.
	// parameters is a block_sequence whose items are block_mappings.
	// Detect existing limit/offset entries by scanning name: keys.
	seq := descendantBlockSequence(paramsVal)
	if seq == nil {
		return nil, nil
	}
	hasLimit := sequenceHasQueryParam(seq, ctx.Source, "limit")
	hasOffset := sequenceHasQueryParam(seq, ctx.Source, "offset")
	if hasLimit && hasOffset {
		return nil, nil
	}
	indent := codemod.IndentOf(seq, ctx.Source)
	inner := indent + 2
	var patches []codemod.Patch
	if !hasLimit {
		entry := fmt.Sprintf("{name: limit, in: query, description: Maximum items to return., schema: {type: integer, minimum: 0, maximum: 250, default: 25}}")
		p := codemod.InsertSequenceItem(seq, ctx.Source, entry)
		_ = inner
		p.URI = ctx.URI
		p.RuleID = diag.Code
		p.Description = "append limit query parameter"
		patches = append(patches, p)
	}
	if !hasOffset {
		entry := fmt.Sprintf("{name: offset, in: query, description: Zero-based offset of the first item to return., schema: {type: integer, minimum: 0, default: 0}}")
		p := codemod.InsertSequenceItem(seq, ctx.Source, entry)
		p.URI = ctx.URI
		p.RuleID = diag.Code
		p.Description = "append offset query parameter"
		patches = append(patches, p)
	}
	return patches, nil
}

// PathParamRequired implements the Fix for
// sailpoint-path-param-required: inserts `required: true` on the
// path parameter's block_mapping. Required by the OAS spec, so the
// fix is always safe.
func PathParamRequired(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	mapping := mappingForDiagnostic(ctx, diag)
	if mapping == nil {
		return nil, nil
	}
	if codemod.MappingHasKey(mapping, ctx.Source, "required") {
		return nil, nil
	}
	// Insert immediately after in: if present, otherwise at end.
	inPair := codemod.FindMappingPair(mapping, ctx.Source, "in")
	var patch codemod.Patch
	if inPair != nil {
		patch = codemod.InsertMappingPairAfter(inPair, ctx.Source, "required", "true")
	} else {
		patch = codemod.InsertMappingPair(mapping, ctx.Source, "required", "true")
	}
	patch.URI = ctx.URI
	patch.RuleID = diag.Code
	patch.Description = "insert required: true on path parameter"
	return []codemod.Patch{patch}, nil
}

// operationMappingFor resolves the operation body's block_mapping. The
// rule emits its diagnostic at locations that belong to the operation
// (responses loc, operation loc, parameters loc), so the enclosing
// block_mapping is the operation itself.
func operationMappingFor(ctx *codemod.FixContext, diag barrelman.Diagnostic) *ts.Node {
	return mappingForDiagnostic(ctx, diag)
}

// descendantBlockSequence returns the first block_sequence child of
// node (or node itself when it already is one).
func descendantBlockSequence(node *ts.Node) *ts.Node {
	if node == nil {
		return nil
	}
	if node.Kind() == "block_sequence" {
		return node
	}
	queue := []*ts.Node{node}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == nil {
			continue
		}
		if cur.Kind() == "block_sequence" {
			return cur
		}
		if cur == node || cur.Kind() == "block_node" || cur.Kind() == "flow_node" {
			for i := uint(0); i < cur.NamedChildCount(); i++ {
				queue = append(queue, cur.NamedChild(i))
			}
		}
	}
	return nil
}

// sequenceHasQueryParam returns true when any item in seq is a block
// mapping with name: matching paramName and in: query.
func sequenceHasQueryParam(seq *ts.Node, src []byte, paramName string) bool {
	for i := uint(0); i < seq.NamedChildCount(); i++ {
		item := seq.NamedChild(i)
		if item == nil {
			continue
		}
		inner := codemod.EnclosingBlockMapping(item)
		if inner == nil {
			continue
		}
		nameVal := codemod.MappingValueNode(inner, src, "name")
		inVal := codemod.MappingValueNode(inner, src, "in")
		if nameVal == nil || inVal == nil {
			continue
		}
		if string(nameVal.Utf8Text(src)) == paramName && string(inVal.Utf8Text(src)) == "query" {
			return true
		}
	}
	return false
}

// paginationLimitEntry / paginationOffsetEntry are left as inline
// flow-style mappings for compactness when we need to bootstrap the
// whole parameters: section. Indented is unused but kept for future
// multi-line rendering.
func paginationLimitEntry(_ int) string {
	return "{name: limit, in: query, description: Maximum items to return., schema: {type: integer, minimum: 0, maximum: 250, default: 25}}"
}

func paginationOffsetEntry(_ int) string {
	return "{name: offset, in: query, description: Zero-based offset of the first item to return., schema: {type: integer, minimum: 0, default: 0}}"
}
