package fixes

import (
	"bytes"
	"fmt"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/codemod"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// Operation4xxResponse inserts a 400 response referencing the shared
// ProblemDetails component when no 4xx code exists.
func Operation4xxResponse(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	return insertStatusResponse(ctx, diag, "400", "Bad request")
}

// Operation401Response inserts a 401 response when absent.
func Operation401Response(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	return insertStatusResponse(ctx, diag, "401", "Unauthorized")
}

// Operation403Response inserts a 403 response when absent.
func Operation403Response(ctx *codemod.FixContext, diag barrelman.Diagnostic) ([]codemod.Patch, error) {
	return insertStatusResponse(ctx, diag, "403", "Forbidden")
}

// insertStatusResponse is the shared worker for the 4xx/401/403
// response fixers. It locates the responses block_mapping inside the
// operation's body and appends a status-code entry that references
// the shared ProblemDetails component.
func insertStatusResponse(ctx *codemod.FixContext, diag barrelman.Diagnostic, code, reason string) ([]codemod.Patch, error) {
	mapping := responsesMappingFor(ctx, diag)
	if mapping == nil {
		return nil, nil
	}
	if codemod.MappingHasKey(mapping, ctx.Source, code) {
		return nil, nil
	}

	// Build the response entry with matching indent. Each inner line
	// is indented by (mapping child indent + 2) to land inside the
	// status key's value.
	childIndent := codemod.ChildIndent(mapping, ctx.Source)
	inner := childIndent + 2
	pad := bytes.Repeat([]byte(" "), inner)

	value := fmt.Sprintf("\n%sdescription: %s\n%scontent:\n%s  application/problem+json:\n%s    schema:\n%s      $ref: '#/components/schemas/ProblemDetails'",
		pad, reason,
		pad,
		pad,
		pad,
		pad,
	)

	patch := codemod.InsertMappingPair(mapping, ctx.Source, quotedStatus(code), value)
	patch.URI = ctx.URI
	patch.RuleID = diag.Code
	patch.Description = fmt.Sprintf("insert %q response referencing ProblemDetails", code)
	return []codemod.Patch{patch}, nil
}

// quotedStatus always quotes the status code in YAML to avoid parsers
// interpreting it as an integer key.
func quotedStatus(code string) string {
	return "\"" + code + "\""
}

// responsesMappingFor locates the responses block_mapping for the
// operation targeted by diag. The diagnostic is emitted at either
// op.ResponsesLoc (which IS the responses mapping) or op.Loc (when
// responses: is absent, which is rare for any real spec). In both
// cases, walk from the diagnostic's node to the nearest block_mapping
// and return the responses mapping nested inside it when the diag
// refers to an operation body, or the node itself when it already is
// the responses mapping.
func responsesMappingFor(ctx *codemod.FixContext, diag barrelman.Diagnostic) *ts.Node {
	if ctx == nil || ctx.Index == nil || diag.ByteRange.IsZero() {
		return nil
	}
	node := codemod.NodeAtByteRange(ctx.Index.Tree(), diag.ByteRange.StartByte, diag.ByteRange.EndByte)
	mapping := codemod.EnclosingBlockMapping(node)
	if mapping == nil {
		return nil
	}
	// Distinguish "mapping is the responses map" (keys look like status
	// codes) from "mapping is the operation body" (has a `responses`
	// child). Status-code heuristic: any key starting with a digit or
	// the literal "default".
	if looksLikeResponsesMap(mapping, ctx.Source) {
		return mapping
	}
	// Operation body: dive into the responses value.
	if value := codemod.MappingValueNode(mapping, ctx.Source, "responses"); value != nil {
		if inner := codemod.EnclosingBlockMapping(value); inner != nil {
			return inner
		}
	}
	return nil
}

// looksLikeResponsesMap returns true when mapping's first non-extension
// key starts with a digit or equals "default" -- characteristic of an
// OpenAPI responses mapping.
func looksLikeResponsesMap(mapping *ts.Node, src []byte) bool {
	if mapping == nil {
		return false
	}
	for i := uint(0); i < mapping.NamedChildCount(); i++ {
		pair := mapping.NamedChild(i)
		if pair == nil || pair.Kind() != "block_mapping_pair" {
			continue
		}
		keyNode := pair.ChildByFieldName("key")
		if keyNode == nil {
			continue
		}
		key := string(keyNode.Utf8Text(src))
		if len(key) >= 2 && (key[0] == '"' || key[0] == '\'') {
			key = key[1 : len(key)-1]
		}
		if key == "default" {
			return true
		}
		if len(key) > 0 && key[0] >= '0' && key[0] <= '9' {
			return true
		}
		return false
	}
	return false
}
