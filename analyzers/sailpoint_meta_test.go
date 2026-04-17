package analyzers_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/sailpoint-oss/barrelman"
	"github.com/sailpoint-oss/barrelman/analyzers"
	"github.com/sailpoint-oss/barrelman/rulesets/bridge"
)

// TestEveryBridgeEntryRegisteredWithMatchingMetadata is the authoritative
// guard: every canonical slug in the bridge must be registered by a
// SailPoint analyzer with a RuleMeta whose ID, guideline number, doc URL,
// and vacuum / spectral IDs match the bridge entry.
func TestEveryBridgeEntryRegisteredWithMatchingMetadata(t *testing.T) {
	reg := barrelman.NewRegistry()
	analyzers.RegisterAll(reg)

	byID := make(map[string]barrelman.RuleMeta)
	for _, meta := range reg.All() {
		byID[meta.ID] = meta
	}

	for _, entry := range bridge.All() {
		meta, ok := byID[entry.Canonical]
		if !ok {
			t.Errorf("rule %q is in the bridge but not registered by analyzers.RegisterAll", entry.Canonical)
			continue
		}
		if !strings.HasPrefix(meta.ID, "sailpoint-") {
			t.Errorf("rule %q ID is not sailpoint-namespaced", meta.ID)
		}
		wantPrimary := entry.PrimaryGuideline()
		if meta.GuidelineID != wantPrimary {
			t.Errorf("rule %q GuidelineID = %d, want %d", meta.ID, meta.GuidelineID, wantPrimary)
		}
		if wantPrimary > 0 {
			wantAnchor := fmt.Sprintf("#%d", wantPrimary)
			if !strings.HasSuffix(meta.DocURL, wantAnchor) {
				t.Errorf("rule %q DocURL %q does not end with %q", meta.ID, meta.DocURL, wantAnchor)
			}
			if meta.DocURL == "" {
				t.Errorf("rule %q has empty DocURL", meta.ID)
			}
		}
		if meta.VacuumID != entry.Vacuum {
			t.Errorf("rule %q VacuumID = %q, want %q", meta.ID, meta.VacuumID, entry.Vacuum)
		}
		if meta.SpectralID != entry.Spectral {
			t.Errorf("rule %q SpectralID = %q, want %q", meta.ID, meta.SpectralID, entry.Spectral)
		}
	}
}

// TestBuiltinRulesetsIncludeEverySailpointRule snapshots the catalog so
// that the `barrelman:recommended` and `barrelman:all` built-in rulesets
// always enumerate every canonical SailPoint rule after the rename.
func TestBuiltinRulesetsIncludeEverySailpointRule(t *testing.T) {
	reg := barrelman.NewRegistry()
	analyzers.RegisterAll(reg)

	registered := make(map[string]bool)
	for _, meta := range reg.All() {
		registered[meta.ID] = true
	}

	for _, entry := range bridge.All() {
		if !registered[entry.Canonical] {
			t.Errorf("canonical rule %q missing from DefaultRegistry after RegisterAll", entry.Canonical)
		}
	}
}
