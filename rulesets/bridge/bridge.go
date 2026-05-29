// Package bridge provides the single authoritative mapping between
// barrelman's canonical rule slugs, vacuum rule IDs, spectral:oas rule IDs,
// and legacy aliases.
//
// The bridge is the one place to add, rename, or remove a rule mapping.
// Every callsite that needs to normalize a rule ID coming from user config,
// a vacuum result, or a spectral result should route through here so that
// reports, LSP diagnostics, and CLI configs all converge on the same slug.
package bridge

import (
	"strconv"
	"strings"
	"sync"
)

// Entry describes a single canonical rule and its equivalents in other
// rule systems. Exactly one of the three ID fields is the canonical slug
// (Canonical); Vacuum and Spectral may be empty when no equivalent exists.
type Entry struct {
	Canonical    string
	Vacuum       string
	Spectral     string
	GuidelineIDs []int
}

// PrimaryGuideline returns the primary guideline number for this entry
// (first entry in GuidelineIDs) or 0 if none is declared.
func (e Entry) PrimaryGuideline() int {
	if len(e.GuidelineIDs) == 0 {
		return 0
	}
	return e.GuidelineIDs[0]
}

// entries is intentionally empty for the generic public rule pack. Private or
// organization-specific rule packs own their own bridge metadata.
var entries = []Entry{}

// legacyToCanonical maps retired kebab-case rule IDs (from prior telescope /
// spectral-alignment naming) to their new canonical slugs. This list shrinks
// over time; new deprecations should be added here rather than sprinkling
// ad-hoc aliases through config code.
var legacyToCanonical = map[string]string{
	// Legacy spectral-ish slugs kept working.
	"no-trailing-slash": "path-keys-no-trailing-slash",
	"template-valid":    "path-declarations-must-exist",
	"params-match":      "path-params",
	"servers-defined":   "oas3-api-servers",
	// Structural validation rebrand.
	"structural-validation": "oas3-schema",
}

var (
	buildOnce       sync.Once
	byCanonical     map[string]Entry
	byVacuum        map[string]Entry
	bySpectral      map[string]Entry
	byGuidelineID   map[int]Entry
	byGuidelineAll  map[int][]Entry
	vacuumToCanon   map[string]string
	spectralToCanon map[string]string
)

func build() {
	byCanonical = make(map[string]Entry, len(entries))
	byVacuum = make(map[string]Entry)
	bySpectral = make(map[string]Entry)
	byGuidelineID = make(map[int]Entry)
	byGuidelineAll = make(map[int][]Entry)
	vacuumToCanon = make(map[string]string)
	spectralToCanon = make(map[string]string)
	for _, e := range entries {
		byCanonical[e.Canonical] = e
		if e.Vacuum != "" {
			byVacuum[e.Vacuum] = e
			vacuumToCanon[e.Vacuum] = e.Canonical
		}
		if e.Spectral != "" {
			bySpectral[e.Spectral] = e
			spectralToCanon[e.Spectral] = e.Canonical
		}
		for _, g := range e.GuidelineIDs {
			byGuidelineAll[g] = append(byGuidelineAll[g], e)
			// Prefer the first-declared canonical for a numeric lookup to be
			// stable; the full slice is available via AllByGuideline.
			if _, ok := byGuidelineID[g]; !ok {
				byGuidelineID[g] = e
			}
		}
	}
}

func ensureBuilt() {
	buildOnce.Do(build)
}

// All returns a copy of every bridge entry in declaration order.
func All() []Entry {
	ensureBuilt()
	out := make([]Entry, len(entries))
	copy(out, entries)
	return out
}

// Canonical normalizes any recognized rule ID (canonical slug, vacuum id,
// spectral:oas id, legacy kebab-case id, or guideline number) into the
// canonical slug when one exists. Unrecognized IDs are returned unchanged so
// callers may still forward vacuum-only or user-defined rule IDs untouched.
func Canonical(id string) string {
	ensureBuilt()
	trimmed := strings.TrimSpace(id)
	if trimmed == "" {
		return ""
	}
	if _, ok := byCanonical[trimmed]; ok {
		return trimmed
	}
	lower := strings.ToLower(trimmed)
	if canonical, ok := legacyToCanonical[trimmed]; ok {
		return canonical
	}
	if canonical, ok := legacyToCanonical[lower]; ok {
		return canonical
	}
	if canonical, ok := vacuumToCanon[trimmed]; ok {
		return canonical
	}
	if canonical, ok := spectralToCanon[trimmed]; ok {
		return canonical
	}
	if n, ok := numericID(trimmed); ok {
		if entry, ok := byGuidelineID[n]; ok {
			return entry.Canonical
		}
	}
	return trimmed
}

// Vacuum returns the vacuum rule ID that corresponds to the given canonical
// slug, or "" when there is no vacuum equivalent.
func Vacuum(id string) string {
	ensureBuilt()
	if e, ok := byCanonical[id]; ok {
		return e.Vacuum
	}
	return ""
}

// Spectral returns the spectral:oas rule ID that corresponds to the given
// canonical slug, or "" when there is no spectral equivalent.
func Spectral(id string) string {
	ensureBuilt()
	if e, ok := byCanonical[id]; ok {
		return e.Spectral
	}
	return ""
}

// FromVacuum returns the bridge entry for a vacuum rule ID. The second
// return value is false when the vacuum rule has no canonical equivalent.
func FromVacuum(vacuumID string) (Entry, bool) {
	ensureBuilt()
	e, ok := byVacuum[vacuumID]
	return e, ok
}

// FromSpectral returns the bridge entry for a spectral rule ID.
func FromSpectral(spectralID string) (Entry, bool) {
	ensureBuilt()
	e, ok := bySpectral[spectralID]
	return e, ok
}

// FromCanonical returns the bridge entry for a canonical slug.
func FromCanonical(canonical string) (Entry, bool) {
	ensureBuilt()
	e, ok := byCanonical[canonical]
	return e, ok
}

// ForGuideline returns every canonical entry tied to the given guideline
// number. Many guideline numbers resolve to multiple rules (e.g. #403 maps
// to four). The returned slice is a copy.
func ForGuideline(id int) []Entry {
	ensureBuilt()
	src := byGuidelineAll[id]
	out := make([]Entry, len(src))
	copy(out, src)
	return out
}

// Canonicals returns every canonical slug in declaration order.
func Canonicals() []string {
	ensureBuilt()
	out := make([]string, len(entries))
	for i, e := range entries {
		out[i] = e.Canonical
	}
	return out
}

func numericID(raw string) (int, bool) {
	trimmed := strings.TrimSpace(raw)
	trimmed = strings.TrimPrefix(trimmed, "#")
	lower := strings.ToLower(trimmed)
	if lower == "" {
		return 0, false
	}
	n, err := strconv.Atoi(lower)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
