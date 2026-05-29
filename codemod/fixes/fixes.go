// Package fixes implements reusable auto-fix hooks. Each Fix function is
// deterministic, insertion-only, and idempotent: running it twice produces
// the same output as running it once, and it emits zero patches when the
// target fix is already present.
//
// The helpers here (FindMapping, TodoDescription, etc.) are shared
// across fix implementations to keep the per-rule files compact.
package fixes

import (
	"fmt"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/codemod"
	ts "github.com/tree-sitter/go-tree-sitter"
)

// TodoDescription is the sentinel value used when no source-aware
// hint is available for a `description:` insertion. Chosen so
// reviewers can grep for stale values post-apply.
const TodoDescription = "TODO"

// TodoExample is the sentinel value used when no hint is available
// for an `example:` insertion.
const TodoExample = "TODO"

// mappingForDiagnostic finds the block_mapping the fix should insert
// into. It uses the diagnostic's ByteRange to locate the source node
// (diagnostics populated from a navigator.Loc.Node always carry a
// valid ByteRange) and walks to the nearest block_mapping.
//
// Returns nil when no mapping is reachable; callers should treat that
// as "no fix available" and return an empty patch list.
func mappingForDiagnostic(ctx *codemod.FixContext, diag barrelman.Diagnostic) *ts.Node {
	if ctx == nil || ctx.Index == nil || diag.ByteRange.IsZero() {
		return nil
	}
	node := codemod.NodeAtByteRange(ctx.Index.Tree(), diag.ByteRange.StartByte, diag.ByteRange.EndByte)
	return codemod.EnclosingBlockMapping(node)
}

// hintOrDefault returns the provided hint when available, otherwise
// defaultValue. The returned string is already YAML-safe for plain
// scalars; callers inserting multi-line or special-character values
// should quote explicitly (see quoted below).
func hintDescription(ctx *codemod.FixContext, pointer, defaultValue string) string {
	if ctx == nil {
		return defaultValue
	}
	if hint, ok := ctx.HintsOrNop().DescriptionFor(pointer); ok && hint != "" {
		return quoted(hint)
	}
	return defaultValue
}

// quoted produces a YAML-safe scalar rendering of s. For simple
// strings (no colon, quote, newline, leading whitespace) it returns
// the string unchanged; otherwise it double-quotes and escapes. Good
// enough for the sentinel values and short hint strings Phase 2 emits.
func quoted(s string) string {
	needsQuote := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c == ':' || c == '#' || c == '"' || c == '\'' || c == '\n' || c == '\r' || c == '\t' {
			needsQuote = true
			break
		}
	}
	if !needsQuote && len(s) > 0 && s[0] != ' ' {
		return s
	}
	// Escape backslashes and double-quotes; leave everything else as-is.
	escaped := make([]byte, 0, len(s)+2)
	escaped = append(escaped, '"')
	for i := 0; i < len(s); i++ {
		switch s[i] {
		case '\\', '"':
			escaped = append(escaped, '\\', s[i])
		case '\n':
			escaped = append(escaped, '\\', 'n')
		case '\r':
			escaped = append(escaped, '\\', 'r')
		case '\t':
			escaped = append(escaped, '\\', 't')
		default:
			escaped = append(escaped, s[i])
		}
	}
	escaped = append(escaped, '"')
	return string(escaped)
}

// pointerFromData extracts a JSON Pointer string from a Diagnostic's
// Data field when present. Used to key hint lookups to a specific
// element. Returns empty string when the pointer is unavailable.
func pointerFromData(diag barrelman.Diagnostic) string {
	switch d := diag.Data.(type) {
	case map[string]any:
		if v, ok := d["pointer"].(string); ok {
			return v
		}
	case map[string]string:
		if v, ok := d["pointer"]; ok {
			return v
		}
	}
	return ""
}

// buildInsertDescriptionPatch emits a patch that inserts
// `description: <value>` into the block_mapping at mapping. Returns
// an empty slice (nil) when the mapping already has a `description`
// key, satisfying the idempotency property.
func buildInsertDescriptionPatch(mapping *ts.Node, ctx *codemod.FixContext, diag barrelman.Diagnostic) []codemod.Patch {
	if mapping == nil {
		return nil
	}
	if codemod.MappingHasKey(mapping, ctx.Source, "description") {
		return nil
	}
	value := hintDescription(ctx, pointerFromData(diag), TodoDescription)
	patch := codemod.InsertMappingPair(mapping, ctx.Source, "description", value)
	patch.URI = ctx.URI
	patch.RuleID = diag.Code
	patch.Description = fmt.Sprintf("insert description on %s", diag.Code)
	return []codemod.Patch{patch}
}

// buildInsertExamplePatch emits a patch that inserts
// `example: <value>` into the block_mapping at mapping. Idempotent.
func buildInsertExamplePatch(mapping *ts.Node, ctx *codemod.FixContext, diag barrelman.Diagnostic) []codemod.Patch {
	if mapping == nil {
		return nil
	}
	if codemod.MappingHasKey(mapping, ctx.Source, "example") ||
		codemod.MappingHasKey(mapping, ctx.Source, "examples") {
		return nil
	}
	value := TodoExample
	if ctx != nil {
		if hint, ok := ctx.HintsOrNop().ExampleFor(pointerFromData(diag)); ok {
			value = quoted(fmt.Sprintf("%v", hint))
		}
	}
	patch := codemod.InsertMappingPair(mapping, ctx.Source, "example", value)
	patch.URI = ctx.URI
	patch.RuleID = diag.Code
	patch.Description = fmt.Sprintf("insert example on %s", diag.Code)
	return []codemod.Patch{patch}
}
