package hints

import "github.com/sailpoint-oss/barrelman/codemod"

// CartographerLookup is the narrow dependency the cartographer-backed
// SourceHintProvider needs from a cartographer source-map. It is an
// interface (rather than a direct cartographer import) so barrelman
// can stay free of a cartographer dependency while still accepting an
// implementation telescope/meridian inject at startup.
//
// Callers supply an implementation that, given a JSON Pointer into
// the OpenAPI document, looks up the corresponding source-code
// element and returns the leading doc-comment (for descriptions) or
// a literal / sample value (for examples).
type CartographerLookup interface {
	// DocCommentFor returns the source-level doc comment for the
	// element referenced by pointer, or ("", false) when no source
	// mapping exists for that pointer.
	DocCommentFor(pointer string) (string, bool)

	// SampleValueFor returns a realistic sample value pulled from
	// the source code (for example an example tag, a default, or
	// an observed literal). Returns (nil, false) when no source
	// information is available.
	SampleValueFor(pointer string) (any, bool)
}

// Cartographer is a SourceHintProvider that delegates to a
// CartographerLookup. telescope wires a concrete implementation
// backed by cartographer/extraction when --source-hints=cartographer
// is set; barrelman itself stays dependency-free.
type Cartographer struct {
	Lookup CartographerLookup
}

var _ codemod.SourceHintProvider = (*Cartographer)(nil)

// DescriptionFor returns the source-level doc comment when available.
func (c *Cartographer) DescriptionFor(pointer string) (string, bool) {
	if c == nil || c.Lookup == nil {
		return "", false
	}
	return c.Lookup.DocCommentFor(pointer)
}

// ExampleFor returns a source-level sample value when available.
func (c *Cartographer) ExampleFor(pointer string) (any, bool) {
	if c == nil || c.Lookup == nil {
		return nil, false
	}
	return c.Lookup.SampleValueFor(pointer)
}
