// Package bridge provides the single authoritative mapping between
// barrelman's canonical SailPoint rule slugs, SailPoint API Guidelines
// numbers, vacuum rule IDs, and spectral:oas rule IDs.
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

// entries is the authoritative bridge table. Keep it sorted by guideline
// number then by canonical slug for reviewability.
var entries = []Entry{
	// #104 - property naming
	{Canonical: "sailpoint-property-camel-case", GuidelineIDs: []int{104}},
	// #107 - path conventions (split into two)
	{Canonical: "sailpoint-path-kebab-case", Vacuum: "paths-kebab-case", GuidelineIDs: []int{107}},
	{Canonical: "sailpoint-path-param-camel-case", GuidelineIDs: []int{107}},
	// #108 - query parameter naming
	{Canonical: "sailpoint-query-param-camel-case", GuidelineIDs: []int{108}},
	// #111 - OAuth scope format
	{Canonical: "sailpoint-oauth-scope-format", GuidelineIDs: []int{111}},
	// #112 - enum casing
	{Canonical: "sailpoint-enum-screaming-snake-case", GuidelineIDs: []int{112}},
	// #115 - descriptions (split into parameter and property)
	{Canonical: "sailpoint-parameter-description", Vacuum: "parameter-description", Spectral: "parameter-description", GuidelineIDs: []int{115}},
	{Canonical: "sailpoint-property-description", GuidelineIDs: []int{115}},
	// #116 - examples (split three ways)
	{Canonical: "sailpoint-parameter-example", GuidelineIDs: []int{116}},
	{Canonical: "sailpoint-property-example", GuidelineIDs: []int{116}},
	{Canonical: "sailpoint-response-example", GuidelineIDs: []int{116}},
	// #122 - operationId requirements (split into casing and uniqueness)
	{Canonical: "sailpoint-operation-id-camel-case", Vacuum: "operation-operationId", Spectral: "operation-operationId", GuidelineIDs: []int{122}},
	{Canonical: "sailpoint-operation-id-unique", Vacuum: "operation-operationId-unique", Spectral: "operation-operationId-unique", GuidelineIDs: []int{122}},
	// #123 - tagging (split into single-tag and documented-tag)
	{Canonical: "sailpoint-operation-single-tag", Vacuum: "operation-tags", Spectral: "operation-tags", GuidelineIDs: []int{123}},
	{Canonical: "sailpoint-tag-documented", Vacuum: "tag-description", Spectral: "tag-description", GuidelineIDs: []int{123}},
	// #124 - resource operation id extension
	{Canonical: "sailpoint-path-param-resource-operation-link", GuidelineIDs: []int{124}},
	// #204 - top-level object responses
	{Canonical: "sailpoint-response-top-level-object", GuidelineIDs: []int{204}},
	// #300 - OAuth required (split into security defined and per-operation)
	{Canonical: "sailpoint-security-oauth2-required", GuidelineIDs: []int{300}},
	{Canonical: "sailpoint-operation-security-required", Vacuum: "oas3-operation-security-defined", Spectral: "oas3-operation-security-defined", GuidelineIDs: []int{300}},
	// #301 - OAuth scopes declared
	{Canonical: "sailpoint-oauth-scopes-declared", GuidelineIDs: []int{301}},
	// #304 - HTTPS servers
	{Canonical: "sailpoint-server-url-https", GuidelineIDs: []int{304}},
	// #403 - status codes (split four ways)
	{Canonical: "sailpoint-operation-method-status-codes", GuidelineIDs: []int{403}},
	{Canonical: "sailpoint-operation-4xx-response", Vacuum: "operation-4xx-response", GuidelineIDs: []int{403}},
	{Canonical: "sailpoint-operation-401-response", GuidelineIDs: []int{403}},
	{Canonical: "sailpoint-operation-403-response", GuidelineIDs: []int{403}},
	// #404 - Problem Details (split four ways)
	{Canonical: "sailpoint-error-problem-details-media-type", GuidelineIDs: []int{404}},
	{Canonical: "sailpoint-error-problem-details-schema", GuidelineIDs: []int{404}},
	{Canonical: "sailpoint-error-problem-details-shared-component", GuidelineIDs: []int{404}},
	{Canonical: "sailpoint-error-correlation-id", GuidelineIDs: []int{404}},
	// #500 - no /api base prefix
	{Canonical: "sailpoint-path-no-api-prefix", GuidelineIDs: []int{500}},
	// #514 - non-numeric path identifiers
	{Canonical: "sailpoint-path-param-no-numeric-id", Vacuum: "owasp-no-numeric-ids", GuidelineIDs: []int{514}},
	// #602 - pagination (split into pagination and wrapper)
	{Canonical: "sailpoint-collection-offset-pagination", GuidelineIDs: []int{602}},
	{Canonical: "sailpoint-collection-wrapped-response", GuidelineIDs: []int{602}},
	// #701 - booleans not nullable
	{Canonical: "sailpoint-boolean-not-nullable", GuidelineIDs: []int{701}},
	// #702 - arrays not nullable
	{Canonical: "sailpoint-array-not-nullable", GuidelineIDs: []int{702}},
	// #710 - required fields (split into schema and path parameter)
	{Canonical: "sailpoint-schema-required-fields", GuidelineIDs: []int{710}},
	{Canonical: "sailpoint-path-param-required", GuidelineIDs: []int{710}},
	// #804 - numeric formats and identifier types
	{Canonical: "sailpoint-numeric-format-approved", GuidelineIDs: []int{804}},
	{Canonical: "sailpoint-identifier-string-type", GuidelineIDs: []int{804}},
	// #903 - X-Request-Id header (split three ways)
	{Canonical: "sailpoint-x-request-id-header", GuidelineIDs: []int{903}},
	{Canonical: "sailpoint-x-request-id-shared-component", GuidelineIDs: []int{903}},
	{Canonical: "sailpoint-x-request-id-uuid", GuidelineIDs: []int{903}},
}

// legacyToCanonical maps retired kebab-case rule IDs (from prior telescope /
// spectral-alignment naming) to their new canonical slugs. This list shrinks
// over time; new deprecations should be added here rather than sprinkling
// ad-hoc aliases through config code.
var legacyToCanonical = map[string]string{
	// Previous telescope-native slugs that pre-date the sailpoint- namespace.
	"operation-operationid":        "sailpoint-operation-id-camel-case",
	"operation-operationId":        "sailpoint-operation-id-camel-case",
	"operationid-unique":           "sailpoint-operation-id-unique",
	"operation-operationId-unique": "sailpoint-operation-id-unique",
	"operation-tags":               "sailpoint-operation-single-tag",
	"parameter-description":        "sailpoint-parameter-description",
	"security-global-or-operation": "sailpoint-operation-security-required",
	"server-url-https":             "sailpoint-server-url-https",
	"missing-error-responses":      "sailpoint-operation-4xx-response",
	"missing-pagination":           "sailpoint-collection-offset-pagination",
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
// spectral:oas id, legacy kebab-case id, SailPoint guideline number, or
// `sp-NNN`) into the canonical SailPoint slug when one exists. When the
// input is already unrecognized it is returned unchanged so callers may
// still forward vacuum-only or user-defined rule IDs untouched.
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
	lower = strings.TrimPrefix(lower, "sp-")
	if lower == "" {
		return 0, false
	}
	n, err := strconv.Atoi(lower)
	if err != nil || n <= 0 {
		return 0, false
	}
	return n, true
}
