// Package hints provides implementations of codemod.SourceHintProvider
// that supply non-sentinel values for fixes. The schema-synth provider
// picks plausible example values based on property-name heuristics;
// the cartographer provider (in cartographer.go) is wired separately
// because it needs access to cartographer's source map and is
// therefore configured at telescope startup.
package hints

import (
	"fmt"
	"strings"

	"github.com/sailpoint-oss/barrelman/codemod"
)

// Synth is a SourceHintProvider that emits realistic example values
// for common OpenAPI property patterns (email, id, timestamp,
// pagination counts). Phase 4 ships this as the default --source-hints
// backend; more sophisticated generation (feature #6 Example Synthesis
// Engine) can extend it by adding more heuristics or plugging in a
// schema-aware generator.
//
// Synth is stateless and safe for concurrent use.
type Synth struct{}

var _ codemod.SourceHintProvider = Synth{}

// DescriptionFor always returns no hint; Synth only supplies example
// values. Callers that want descriptions should combine Synth with a
// cartographer-backed DescriptionFor provider via the Composite
// helper below.
func (Synth) DescriptionFor(string) (string, bool) { return "", false }

// ExampleFor returns a plausible example based on the pointer's
// trailing component. When no heuristic matches, returns
// (nil, false) so the fix falls back to its sentinel TODO.
func (Synth) ExampleFor(pointer string) (any, bool) {
	if pointer == "" {
		return nil, false
	}
	name := strings.ToLower(lastSegment(pointer))
	switch {
	case strings.HasSuffix(name, "id"), name == "uuid":
		return "f4e0b80f-2e25-4b0e-8b3c-70a3dcb9b2f3", true
	case name == "email" || strings.HasSuffix(name, "email"):
		return "name@example.com", true
	case name == "username" || strings.HasSuffix(name, "name"):
		return "jane.doe", true
	case name == "createdat" || name == "created_at" || name == "updatedat" || name == "updated_at":
		return "2026-04-16T12:00:00Z", true
	case name == "limit":
		return 25, true
	case name == "offset":
		return 0, true
	case name == "count":
		return 100, true
	case name == "status":
		return "ACTIVE", true
	case name == "type":
		return "EXAMPLE_TYPE", true
	case name == "uri" || name == "url":
		return "https://example.com/resource", true
	case strings.HasPrefix(name, "is") || strings.HasPrefix(name, "has"):
		return true, true
	}
	return nil, false
}

// Composite combines two providers: the first is tried, then the
// second on miss. Useful for layering (cartographer descriptions +
// synth examples).
type Composite struct {
	Primary   codemod.SourceHintProvider
	Secondary codemod.SourceHintProvider
}

// DescriptionFor returns Primary's answer when available, else Secondary's.
func (c Composite) DescriptionFor(pointer string) (string, bool) {
	if c.Primary != nil {
		if v, ok := c.Primary.DescriptionFor(pointer); ok {
			return v, true
		}
	}
	if c.Secondary != nil {
		return c.Secondary.DescriptionFor(pointer)
	}
	return "", false
}

// ExampleFor returns Primary's answer when available, else Secondary's.
func (c Composite) ExampleFor(pointer string) (any, bool) {
	if c.Primary != nil {
		if v, ok := c.Primary.ExampleFor(pointer); ok {
			return v, true
		}
	}
	if c.Secondary != nil {
		return c.Secondary.ExampleFor(pointer)
	}
	return nil, false
}

// lastSegment returns the final slash-delimited segment of a JSON
// pointer, with common OpenAPI suffixes stripped (`/example`,
// `/schema`) so the heuristic matches the real field name.
func lastSegment(pointer string) string {
	trimmed := strings.TrimPrefix(pointer, "/")
	parts := strings.Split(trimmed, "/")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	// Stripping unhelpful trailing wrappers like "schema" or
	// "example" uncovers the actual property name.
	for i := len(parts) - 1; i >= 0; i-- {
		p := parts[i]
		if p == "schema" || p == "example" || p == "examples" || p == "properties" {
			continue
		}
		last = p
		break
	}
	return last
}

// LogDebug is a small helper for diagnostics during provider
// development; not used in production. Kept as an example of how
// providers can be wired up with observability.
func LogDebug(pointer string, got any, matched bool) string {
	return fmt.Sprintf("synth(%s) -> %v, matched=%v", pointer, got, matched)
}
