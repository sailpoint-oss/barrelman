// Package barrelman provides the core validation, linting, and rule engine for
// OpenAPI specifications. It is completely decoupled from any LSP protocol or
// gossip framework -- telescope stitches barrelman rules into its gossip-based
// LSP via a thin bridge layer.
package barrelman

import (
	navigator "github.com/sailpoint-oss/navigator"
)

// Severity indicates the severity level of a diagnostic.
type Severity int

const (
	SeverityError   Severity = 1
	SeverityWarning Severity = 2
	SeverityInfo    Severity = 3
	SeverityHint    Severity = 4
)

// DiagnosticTag adds semantic metadata to a diagnostic.
type DiagnosticTag int

const (
	DiagnosticTagUnnecessary DiagnosticTag = 1
	DiagnosticTagDeprecated  DiagnosticTag = 2
)

// Position is a zero-based line and character offset (from navigator).
type Position = navigator.Position

// Range is a start/end pair of positions (from navigator).
type Range = navigator.Range

// ContainsPosition delegates to navigator.ContainsPosition.
var ContainsPosition = navigator.ContainsPosition

// IsEmpty delegates to navigator.IsEmpty.
var IsEmpty = navigator.IsEmpty

// FileStartRange is a 1-character range at {0,0}->{0,1}, suitable for
// document-level diagnostics when no specific location applies.
var FileStartRange = Range{
	Start: Position{Line: 0, Character: 0},
	End:   Position{Line: 0, Character: 1},
}

// ByteRange is a byte-offset span into the source document. It is
// populated by Reporter.At whenever the supplied navigator.Loc carries
// a tree-sitter node, and is the preferred coordinate system for
// codemod patches (which operate on raw source bytes rather than
// editor-facing line/character positions).
//
// StartByte is inclusive, EndByte is exclusive; EndByte == StartByte
// means a zero-width position suitable for pure insertion.
type ByteRange struct {
	StartByte uint
	EndByte   uint
}

// IsZero reports whether the range has no associated byte offsets
// (both ends are zero). Zero-valued ByteRange signals that the
// diagnostic was not produced from a tree-sitter-backed parse.
func (b ByteRange) IsZero() bool {
	return b.StartByte == 0 && b.EndByte == 0
}

// Diagnostic represents an issue found during analysis.
type Diagnostic struct {
	URI             string
	Range           Range
	ByteRange       ByteRange
	Severity        Severity
	Code            string
	CodeDescription string // URL for documentation about this diagnostic code
	Source          string
	Message         string
	Tags            []DiagnosticTag
	Related         []RelatedInformation
	Fixes           []Fix
	Data            interface{}
}

// RelatedInformation represents a related message and source location.
type RelatedInformation struct {
	URI     string
	Range   Range
	Message string
}

// Fix describes a suggested code fix for a diagnostic.
type Fix struct {
	Description string
	Edits       []TextEdit
}

// TextEdit represents a text replacement in a document.
type TextEdit struct {
	Range   Range
	NewText string
}
