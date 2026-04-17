package rulesets

import "github.com/sailpoint-oss/barrelman/rulesets/bridge"

// SpectralOAS is the name of the Spectral OpenAPI built-in ruleset.
const SpectralOAS = "spectral:oas"

// spectralToTelescope maps spectral:oas rule IDs to native barrelman rule
// IDs for rules implemented by existing analyzers but not in the SailPoint
// bridge (for example generic `info-*` checks). The SailPoint bridge is
// consulted first for spectral IDs that map onto a canonical SailPoint rule.
var spectralToTelescope = map[string]string{
	"info-contact":                "info-contact",
	"info-description":            "info-description",
	"info-license":                "info-license",
	"operation-description":       "operation-description",
	"path-keys-no-trailing-slash": "path-keys-no-trailing-slash",
	"path-declarations-must-exist": "path-declarations-must-exist",
	"path-params":                 "path-params",
	"no-eval-in-markdown":         "description-markdown",
	"no-script-tags-in-markdown":  "description-html",
	"oas3-api-servers":            "oas3-api-servers",
	"oas3-schema":                 "oas3-schema",
	"oas3-unused-component":       "unused-component",
	"contact-properties":          "contact-properties",
	"license-url":                 "license-url",
	"oas3-valid-media-example":    "oas3-valid-media-example",
	"oas3-valid-schema-example":   "oas3-valid-schema-example",
}

// telescopeToSpectral is the reverse mapping.
var telescopeToSpectral = func() map[string]string {
	m := make(map[string]string, len(spectralToTelescope))
	for spectral, telescope := range spectralToTelescope {
		m[telescope] = spectral
	}
	return m
}()

// SpectralToTelescopeID returns the native barrelman rule ID for a Spectral
// OAS rule, or the original ID if no mapping exists. The SailPoint bridge
// is consulted first so spectral rules that map to a canonical SailPoint
// slug (for example operation-operationId -> sailpoint-operation-id-camel-case)
// resolve to the SailPoint rule.
func SpectralToTelescopeID(spectralID string) string {
	if entry, ok := bridge.FromSpectral(spectralID); ok {
		return entry.Canonical
	}
	if tid, ok := spectralToTelescope[spectralID]; ok {
		return tid
	}
	return spectralID
}

// TelescopeToSpectralID returns the Spectral OAS rule ID for a native
// barrelman rule, or the original ID if no mapping exists.
func TelescopeToSpectralID(telescopeID string) string {
	if sid := bridge.Spectral(telescopeID); sid != "" {
		return sid
	}
	if sid, ok := telescopeToSpectral[telescopeID]; ok {
		return sid
	}
	return telescopeID
}

// IsNativeRule reports whether the given Spectral rule ID has a native
// barrelman implementation.
func IsNativeRule(spectralID string) bool {
	if _, ok := bridge.FromSpectral(spectralID); ok {
		return true
	}
	_, ok := spectralToTelescope[spectralID]
	return ok
}

// spectralOASDefaults defines the default severity for each spectral:oas
// rule. Keys are spectral rule IDs; they are normalized to canonical
// SailPoint slugs (via the bridge) when building the ruleset so that
// reports and config overrides always speak the same canonical names.
var spectralOASDefaults = map[string]string{
	"info-contact":                    "warn",
	"info-description":                "warn",
	"info-license":                    "warn",
	"operation-description":           "warn",
	"operation-operationId":           "warn",
	"operation-operationId-unique":    "error",
	"operation-tags":                  "warn",
	"path-keys-no-trailing-slash":     "warn",
	"path-declarations-must-exist":    "error",
	"path-params":                     "error",
	"no-eval-in-markdown":             "warn",
	"no-script-tags-in-markdown":      "warn",
	"oas3-api-servers":                "warn",
	"oas3-schema":                     "error",
	"tag-description":                 "warn",
	"parameter-description":           "warn",
	"contact-properties":              "warn",
	"duplicated-entry-in-enum":        "warn",
	"license-url":                     "warn",
	"oas3-operation-security-defined": "warn",
	"oas3-valid-media-example":        "warn",
	"oas3-valid-schema-example":       "warn",
	"oas3-unused-component":           "warn",
	"typed-enum":                      "warn",
}

// GetSpectralBuiltin returns a RuleSet for the given Spectral built-in name.
// Currently only "spectral:oas" is supported.
func GetSpectralBuiltin(name string) *RuleSet {
	if name != SpectralOAS {
		return nil
	}

	rs := &RuleSet{
		Name:        "Spectral OAS",
		Description: "Spectral OpenAPI ruleset mapped to canonical SailPoint rule IDs.",
		Rules:       make(map[string]RuleDefinition, len(spectralOASDefaults)),
	}

	for spectralID, sev := range spectralOASDefaults {
		canonical := SpectralToTelescopeID(spectralID)
		rs.Rules[canonical] = RuleDefinition{Severity: sev}
	}

	return rs
}
