// Package engine is the unified OpenAPI-aware JSON Schema validation engine
// that replaces the fragmented stack of gossip/jsonschema, spectral's
// hand-rolled `schema` function, barrelman's examples heuristics, and
// meridian's string-based gates.
//
// The engine is a thin contextual layer on top of
// github.com/santhosh-tekuri/jsonschema/v6 (already a dependency). Santhosh
// provides JSON Schema 2020-12 conformance; this package adds:
//
//   - A single structured Issue type with a rule code, JSON pointer, human
//     path, expected/received snapshots, did-you-mean suggestions, and
//     optional tree-sitter range.
//   - OpenAPI semantic adapters (discriminator, OAS 3.0 `nullable`, format
//     handling) layered via compile-time schema rewriting.
//   - A Zod-style message formatter with terse / structured / LSP tiers.
//
// See validation/engine/openapi and validation/engine/message for the
// layered adapters; this file owns the engine-wide public types.
package engine

import (
	"fmt"
	"strings"
)

// Severity mirrors barrelman.Severity values but is expressed here so the
// engine has no hard dependency on any downstream diagnostic type.
type Severity int

const (
	SeverityError   Severity = 1
	SeverityWarning Severity = 2
	SeverityInfo    Severity = 3
	SeverityHint    Severity = 4
)

// Source names the validator that produced an issue. Populated so the
// engine's issues can be distinguished from adjacent toolchain issues in
// merged streams (e.g. mixed with barrelman analyzer diagnostics).
type Source string

const (
	SourceEngine  Source = "validation-engine"
	SourceExample Source = "validation-engine/example"
)

// Range describes a file span in editor coordinates (zero-based lines,
// UTF-16 characters). The engine treats Range as an opaque carrier so
// tree-sitter-backed callers can populate it without pulling in
// navigator as a dependency here.
type Range struct {
	Start Position
	End   Position
}

// Position is a zero-based (line, character) pair.
type Position struct {
	Line      int
	Character int
}

// Issue is the engine's unified diagnostic shape. Every validation path
// produces Issues; adapters (LSP, CLI, gate report) consume Issues and
// render them to their own output formats.
type Issue struct {
	// Code is the stable identifier for the kind of issue (for example
	// "type", "enum", "required", or a SailPoint rule slug like
	// "sailpoint-error-problem-details-shared-component").
	Code string

	// Severity drives display grouping and gate pass/fail decisions.
	Severity Severity

	// Source names the validator that produced this issue.
	Source Source

	// Message is a short, human-readable single-sentence description.
	// Message is always populated; callers render multi-line forms by
	// combining Message with Expected / Received / Suggestion.
	Message string

	// Pointer is the RFC 6901 JSON Pointer into the instance that
	// contains the failure (for example "/components/schemas/User/age").
	Pointer string

	// Path is the Zod-style dot/bracket path rendered from Pointer, for
	// example `components.schemas.User.properties.age`. Present when the
	// engine can derive a friendly path; otherwise empty.
	Path string

	// Range is the tree-sitter span of the failing instance node when
	// known. Zero value means the caller did not provide a node-aware
	// instance.
	Range Range

	// Expected and Received are optional human-readable snapshots used by
	// the Zod-style formatter ("expected integer, received string").
	Expected string
	Received string

	// Suggestion is a did-you-mean hint (for example for misspelled
	// property names or enum values). Empty when not applicable.
	Suggestion string

	// Causes lists nested issues. Composition keywords like oneOf and
	// anyOf populate Causes so UIs can present drill-down context.
	Causes []*Issue

	// Data carries structured metadata consumers may render
	// (for example {"variant": "Dog"} for discriminator issues).
	Data map[string]any
}

// HumanPath returns Path when non-empty, or a best-effort rendering of
// Pointer otherwise. Never returns an empty string for a non-root issue.
func (i Issue) HumanPath() string {
	if i.Path != "" {
		return i.Path
	}
	if i.Pointer == "" || i.Pointer == "/" {
		return "<root>"
	}
	segs := strings.Split(strings.TrimPrefix(i.Pointer, "/"), "/")
	for j, seg := range segs {
		seg = strings.ReplaceAll(seg, "~1", "/")
		seg = strings.ReplaceAll(seg, "~0", "~")
		segs[j] = seg
	}
	return strings.Join(segs, ".")
}

// String produces a terse single-line rendering suitable for CLI log
// lines and test failure messages.
func (i Issue) String() string {
	var b strings.Builder
	b.WriteString("at ")
	b.WriteString(i.HumanPath())
	b.WriteString(": ")
	b.WriteString(i.Message)
	if i.Expected != "" && i.Received != "" {
		fmt.Fprintf(&b, " (expected %s, received %s)", i.Expected, i.Received)
	}
	if i.Suggestion != "" {
		fmt.Fprintf(&b, " — did you mean %q?", i.Suggestion)
	}
	return b.String()
}
