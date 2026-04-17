package codemod

import (
	"errors"
	"fmt"
	"sort"
)

// ErrPatchConflict is returned when two patches target overlapping
// byte ranges in the same document. The Driver requires disjoint
// ranges to guarantee deterministic, order-independent application.
var ErrPatchConflict = errors.New("codemod: overlapping patches")

// ErrParseRegression is returned by ApplyAndVerify when the supplied
// parse function reports an error on the post-patch bytes. When this
// error is returned, the document on disk (or held by the caller) is
// unchanged: rollback is automatic.
var ErrParseRegression = errors.New("codemod: patch introduced a parse regression")

// ErrShrinkDisallowed is returned when a patch would shrink or delete
// source bytes while the Driver is running with AllowShrink=false
// (the default). Phase 2 fixes are pure insertions, so this is free
// safety; callers that genuinely need to replace/shrink must opt in.
var ErrShrinkDisallowed = errors.New("codemod: shrinking patches require AllowShrink=true")

// Driver applies Patches to source documents.
//
// The zero value is a safe default: insertion-only, no parse
// verification, no shrink allowed. Callers that want parse
// verification supply a ParseFn; callers that need to delete or
// replace bytes set AllowShrink=true.
type Driver struct {
	// AllowShrink permits patches whose Replacement is shorter than
	// their target range. When false (the default) such patches are
	// rejected with ErrShrinkDisallowed; this makes the safe-only
	// Phase 2 behavior free.
	AllowShrink bool

	// ParseFn, when non-nil, is called on the post-patch bytes from
	// Apply. If it returns a non-nil error, Apply discards the patched
	// bytes and returns ErrParseRegression. Use this to guard against
	// malformed-YAML regressions that slip past patch-level
	// preconditions.
	ParseFn func([]byte) error
}

// Apply applies every patch in patches to src and returns the patched
// bytes. Patches must have pairwise-disjoint [StartByte, EndByte)
// ranges or the call fails with ErrPatchConflict. Application order is
// internally reverse-sorted (by StartByte descending) so earlier
// patches cannot invalidate later patch offsets.
//
// src is never mutated: a fresh byte slice is returned.
func (d *Driver) Apply(src []byte, patches []Patch) ([]byte, error) {
	if len(patches) == 0 {
		out := make([]byte, len(src))
		copy(out, src)
		return out, nil
	}

	// Copy so callers can pass slices they continue to own.
	sorted := make([]Patch, len(patches))
	copy(sorted, patches)

	// Validate: bounds + shrink policy.
	for i, p := range sorted {
		if p.StartByte > p.EndByte {
			return nil, fmt.Errorf("codemod: patch %d has StartByte %d > EndByte %d", i, p.StartByte, p.EndByte)
		}
		if p.EndByte > uint(len(src)) {
			return nil, fmt.Errorf("codemod: patch %d EndByte %d exceeds source length %d", i, p.EndByte, len(src))
		}
		if !d.AllowShrink && len(p.Replacement) < int(p.EndByte-p.StartByte) {
			return nil, fmt.Errorf("%w: rule %q at %d..%d",
				ErrShrinkDisallowed, p.RuleID, p.StartByte, p.EndByte)
		}
	}

	// Sort descending by StartByte so reverse-order application keeps
	// every patch's offsets valid. Secondary sort by EndByte keeps
	// the order deterministic when two patches share a StartByte (one
	// must be a pure insertion because we already reject overlap).
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].StartByte != sorted[j].StartByte {
			return sorted[i].StartByte > sorted[j].StartByte
		}
		return sorted[i].EndByte > sorted[j].EndByte
	})

	// Overlap detection: after descending sort, each patch's EndByte
	// must be <= the previous (larger-start) patch's StartByte. Equal
	// offsets are allowed only when both are pure insertions at the
	// same point (degenerate but legal; they will be applied in
	// descending-index order).
	for i := 0; i < len(sorted)-1; i++ {
		cur := sorted[i]
		next := sorted[i+1]
		if next.EndByte > cur.StartByte {
			return nil, fmt.Errorf("%w: %s overlaps %s",
				ErrPatchConflict, next.String(), cur.String())
		}
		if next.EndByte == cur.StartByte && !(next.IsInsertion() && cur.IsInsertion()) {
			return nil, fmt.Errorf("%w: %s abuts %s (both must be insertions to share an offset)",
				ErrPatchConflict, next.String(), cur.String())
		}
	}

	// Apply in descending-start order: splice Replacement into a
	// mutable buffer. Building the output incrementally in-place
	// preserves byte-exact fidelity outside each patch window.
	out := make([]byte, len(src))
	copy(out, src)
	for _, p := range sorted {
		out = spliceBytes(out, int(p.StartByte), int(p.EndByte), p.Replacement)
	}
	return out, nil
}

// ApplyAndVerify runs Apply and then calls d.ParseFn (or the optional
// verifyFn override) on the result. When the parse fails, the original
// src is returned along with ErrParseRegression so callers never have
// to deal with a half-broken file.
//
// If d.ParseFn is nil and verifyFn is nil, ApplyAndVerify behaves like
// Apply.
func (d *Driver) ApplyAndVerify(src []byte, patches []Patch, verifyFn func([]byte) error) ([]byte, error) {
	patched, err := d.Apply(src, patches)
	if err != nil {
		return nil, err
	}

	parse := verifyFn
	if parse == nil {
		parse = d.ParseFn
	}
	if parse == nil {
		return patched, nil
	}
	if parseErr := parse(patched); parseErr != nil {
		// Return a copy of the original src so callers never see the
		// partially-applied bytes even transiently.
		orig := make([]byte, len(src))
		copy(orig, src)
		return orig, fmt.Errorf("%w: %v", ErrParseRegression, parseErr)
	}
	return patched, nil
}

// GroupByURI partitions a mixed slice of patches into one slice per
// URI. Callers typically use this when fixing many files at once so
// they can run Apply per-file in parallel.
func GroupByURI(patches []Patch) map[string][]Patch {
	out := make(map[string][]Patch)
	for _, p := range patches {
		out[p.URI] = append(out[p.URI], p)
	}
	return out
}

// spliceBytes returns src with [start, end) replaced by replacement.
// It allocates a fresh slice sized to the exact final length so we
// never grow or shrink the underlying array unexpectedly.
func spliceBytes(src []byte, start, end int, replacement []byte) []byte {
	out := make([]byte, 0, len(src)-(end-start)+len(replacement))
	out = append(out, src[:start]...)
	out = append(out, replacement...)
	out = append(out, src[end:]...)
	return out
}
