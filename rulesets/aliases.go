package rulesets

import "github.com/sailpoint-oss/barrelman/rulesets/bridge"

// NormalizeRuleID resolves any known alias (legacy kebab id, vacuum id,
// spectral id, SailPoint guideline number, or `sp-NNN`) into the canonical
// SailPoint slug. Unrecognized IDs are returned unchanged.
//
// This function is a thin wrapper around bridge.Canonical and exists for
// call-site stability; new code should depend on the bridge package
// directly.
func NormalizeRuleID(id string) string {
	return bridge.Canonical(id)
}
