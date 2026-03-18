// Package barrelman provides the core validation, linting, and rule engine for
// OpenAPI specifications. It is completely decoupled from any LSP protocol or
// gossip framework -- telescope stitches barrelman rules into its gossip-based
// LSP via a thin bridge layer.
package barrelman

import (
	navigator "github.com/LukasParke/navigator"
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

// Diagnostic represents an issue found during analysis.
type Diagnostic struct {
	URI             string
	Range           Range
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
