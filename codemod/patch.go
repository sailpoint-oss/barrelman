// Package codemod provides the auto-fix framework for barrelman rules.
//
// A Patch is a byte-range edit against the raw source document a rule's
// diagnostic was produced from. The Driver composes Patches together so
// multiple fixes for the same file can apply in one pass with overlap
// detection and parse-regression rollback. Rules author Patches through
// the helpers in yaml.go (for block mappings/sequences) or by emitting
// raw byte-range edits directly.
//
// The framework is deliberately independent of any LSP or CLI layer:
// it produces Patches; callers decide whether to apply them to disk,
// render them as LSP WorkspaceEdits, or emit them as GitHub review
// suggestions.
package codemod

import "fmt"

// Patch is a single byte-range edit against a source document. A Patch
// with StartByte == EndByte is a pure insertion; otherwise the range
// [StartByte, EndByte) is replaced with Replacement.
//
// Patches are the lowest-level unit the Driver applies. Helpers in
// yaml.go build Patches from tree-sitter nodes and YAML conventions;
// rule authors can also construct Patches directly when they need
// fine-grained control over the emitted text.
type Patch struct {
	// URI identifies the document the patch targets. Multiple patches
	// with different URIs can be handed to the Driver in one call; the
	// Driver groups by URI before applying.
	URI string

	// StartByte is the inclusive byte offset of the patch's start.
	StartByte uint

	// EndByte is the exclusive byte offset of the patch's end. When
	// EndByte == StartByte the patch is a pure insertion at StartByte.
	EndByte uint

	// Replacement is the new bytes to place at [StartByte, EndByte).
	// May be empty to denote deletion (rejected by default unless the
	// Driver's AllowShrink is set; see driver.go).
	Replacement []byte

	// Description is a one-line human-readable summary used in CLI
	// dry-run output, LSP code-action titles, and PR suggestion
	// blocks. Example: "insert description: TODO on parameter 'limit'".
	Description string

	// RuleID is the canonical rule slug (for example
	// "sailpoint-parameter-description") the patch was produced for.
	// Used for waiver lookup, stats, and LSP action grouping.
	RuleID string
}

// IsInsertion reports whether the patch is a pure insertion (zero-width
// target range).
func (p Patch) IsInsertion() bool {
	return p.StartByte == p.EndByte
}

// Size returns the replacement's length minus the original range
// length. Positive means the patch grows the file; negative means it
// shrinks it.
func (p Patch) Size() int {
	return len(p.Replacement) - int(p.EndByte-p.StartByte)
}

// String renders a terse patch summary suitable for CLI logging.
func (p Patch) String() string {
	if p.Description != "" {
		return fmt.Sprintf("%s@%d..%d: %s", p.RuleID, p.StartByte, p.EndByte, p.Description)
	}
	return fmt.Sprintf("%s@%d..%d: %d bytes", p.RuleID, p.StartByte, p.EndByte, len(p.Replacement))
}
