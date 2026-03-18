package rulesets

import (
	"github.com/sailpoint-oss/barrelman"
)

// Built-in ruleset names.
const (
	Recommended = "telescope:recommended"
	All         = "telescope:all"
	OWASP       = "telescope:owasp"
	Strict      = "telescope:strict"
)

// GetBuiltin returns a resolved ruleset by its built-in name.
func GetBuiltin(name string) *RuleSet {
	switch name {
	case Recommended:
		return recommendedRuleSet()
	case All:
		return allRuleSet()
	case OWASP:
		return owaspRuleSet()
	case Strict:
		return strictRuleSet()
	default:
		return nil
	}
}

func recommendedRuleSet() *RuleSet {
	rs := &RuleSet{
		Name:        "Telescope Recommended",
		Description: "Curated set of the most important OpenAPI rules.",
		Rules:       make(map[string]RuleDefinition),
	}
	for _, meta := range barrelman.DefaultRegistry.All() {
		if meta.Recommended {
			rs.Rules[meta.ID] = RuleDefinition{Severity: severityString(meta.Severity)}
		}
	}
	return rs
}

func allRuleSet() *RuleSet {
	rs := &RuleSet{
		Name:        "Telescope All",
		Description: "All available OpenAPI rules.",
		Rules:       make(map[string]RuleDefinition),
	}
	for _, meta := range barrelman.DefaultRegistry.All() {
		if meta.Category != barrelman.CategoryOWASP {
			rs.Rules[meta.ID] = RuleDefinition{Severity: severityString(meta.Severity)}
		}
	}
	return rs
}

func owaspRuleSet() *RuleSet {
	rs := &RuleSet{
		Name:        "Telescope OWASP",
		Description: "OWASP API security rules.",
		Rules:       make(map[string]RuleDefinition),
	}
	for _, meta := range barrelman.DefaultRegistry.ByCategory(barrelman.CategoryOWASP) {
		rs.Rules[meta.ID] = RuleDefinition{Severity: severityString(meta.Severity)}
	}
	return rs
}

func strictRuleSet() *RuleSet {
	recommended := recommendedRuleSet()
	owasp := owaspRuleSet()
	rs := &RuleSet{
		Name:        "Telescope Strict",
		Description: "Recommended rules plus OWASP with stricter severities.",
		Rules:       make(map[string]RuleDefinition, len(recommended.Rules)+len(owasp.Rules)),
	}
	for id, def := range recommended.Rules {
		rs.Rules[id] = def
	}
	for id, def := range owasp.Rules {
		rs.Rules[id] = def
	}
	return rs
}

func severityString(s barrelman.Severity) string {
	switch s {
	case barrelman.SeverityError:
		return "error"
	case barrelman.SeverityWarning:
		return "warn"
	case barrelman.SeverityInfo:
		return "info"
	case barrelman.SeverityHint:
		return "hint"
	default:
		return "warn"
	}
}
