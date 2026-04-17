package codemod

import (
	navigator "github.com/sailpoint-oss/navigator"
)

// FixContext is handed to each rule's Fix function. It carries
// everything a fix implementation might need: the raw source bytes the
// tree-sitter parse was built from, the full navigator.Index (semantic
// model), and the URI identifying the document.
//
// Fix implementations are expected to treat FixContext as read-only;
// they should emit one or more Patches and return. The Driver is
// responsible for ordering and applying them.
type FixContext struct {
	// Index is the semantic model barrelman rules walk during Run. Fix
	// authors use it to locate neighboring nodes (for example to find
	// the shared ProblemDetails component when fixing a response).
	Index *navigator.Index

	// Source is the raw bytes of the document at URI. Tree-sitter byte
	// offsets on navigator.Loc.Node are indices into this slice.
	Source []byte

	// URI identifies the document the Fix is targeting.
	URI string

	// Hints is an optional source-aware hint provider used by Phase 4
	// integrations (cartographer-backed descriptions, schema-synth
	// examples). When nil, Fix authors should fall back to their
	// sentinel values (TODO, reasonable defaults).
	Hints SourceHintProvider
}

// SourceHintProvider supplies optional non-sentinel values for a Fix.
// Implementations look up the JSON Pointer of the diagnostic and
// return a realistic description or example value when one is
// available (for example from the service's source-code doc comments
// or a property-name-driven synth engine).
//
// Zero-value implementations return (_, false); Fix authors fall back
// to TODO sentinels when no hint is available. Phase 4 wires real
// providers in behind the --source-hints flag.
type SourceHintProvider interface {
	DescriptionFor(pointer string) (string, bool)
	ExampleFor(pointer string) (any, bool)
}

// NopHintProvider is a SourceHintProvider that never returns a hint.
// Used as the default when callers do not set FixContext.Hints.
type NopHintProvider struct{}

// DescriptionFor always returns ("", false).
func (NopHintProvider) DescriptionFor(string) (string, bool) { return "", false }

// ExampleFor always returns (nil, false).
func (NopHintProvider) ExampleFor(string) (any, bool) { return nil, false }

// HintsOrNop returns ctx.Hints when non-nil, or a NopHintProvider.
func (c *FixContext) HintsOrNop() SourceHintProvider {
	if c == nil || c.Hints == nil {
		return NopHintProvider{}
	}
	return c.Hints
}
