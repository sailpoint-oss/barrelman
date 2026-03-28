package rulesets

type CatalogRule struct {
	ID          string
	Severity    Severity
	Category    string
	Recommended bool
}

var builtinCatalogProvider func() []CatalogRule

// SetBuiltinCatalogProvider installs the rule catalog used to materialize
// built-in rulesets. Barrelman wires this to its live registry; other callers
// may provide their own catalogs in tests or alternate integrations.
func SetBuiltinCatalogProvider(provider func() []CatalogRule) {
	builtinCatalogProvider = provider
}

// Built-in ruleset names.
const (
	Recommended = "barrelman:recommended"
	All         = "barrelman:all"
	OWASP       = "barrelman:owasp"
	Strict      = "barrelman:strict"

	LegacyRecommended = "telescope:recommended"
	LegacyAll         = "telescope:all"
	LegacyOWASP       = "telescope:owasp"
	LegacyStrict      = "telescope:strict"
)

// GetBuiltin returns a resolved ruleset by its built-in name.
func GetBuiltin(name string) *RuleSet {
	switch name {
	case Recommended, LegacyRecommended:
		return recommendedRuleSet()
	case All, LegacyAll:
		return allRuleSet()
	case OWASP, LegacyOWASP:
		return owaspRuleSet()
	case Strict, LegacyStrict:
		return strictRuleSet()
	default:
		return nil
	}
}

func recommendedRuleSet() *RuleSet {
	rs := &RuleSet{
		Name:        "Barrelman Recommended",
		Description: "Curated set of the most important API-description rules.",
		Rules:       make(map[string]RuleDefinition),
	}
	for _, meta := range builtinCatalog() {
		if meta.Recommended {
			rs.Rules[meta.ID] = RuleDefinition{Severity: severityString(meta.Severity)}
		}
	}
	return rs
}

func allRuleSet() *RuleSet {
	rs := &RuleSet{
		Name:        "Barrelman All",
		Description: "All available non-OWASP API-description rules.",
		Rules:       make(map[string]RuleDefinition),
	}
	for _, meta := range builtinCatalog() {
		if meta.Category != "owasp" {
			rs.Rules[meta.ID] = RuleDefinition{Severity: severityString(meta.Severity)}
		}
	}
	return rs
}

func owaspRuleSet() *RuleSet {
	rs := &RuleSet{
		Name:        "Barrelman OWASP",
		Description: "OWASP API security rules.",
		Rules:       make(map[string]RuleDefinition),
	}
	for _, meta := range builtinCatalog() {
		if meta.Category != "owasp" {
			continue
		}
		rs.Rules[meta.ID] = RuleDefinition{Severity: severityString(meta.Severity)}
	}
	return rs
}

func strictRuleSet() *RuleSet {
	recommended := recommendedRuleSet()
	owasp := owaspRuleSet()
	rs := &RuleSet{
		Name:        "Barrelman Strict",
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

func builtinCatalog() []CatalogRule {
	if builtinCatalogProvider == nil {
		return nil
	}
	return builtinCatalogProvider()
}

func severityString(s Severity) string {
	switch s {
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warn"
	case SeverityInfo:
		return "info"
	case SeverityHint:
		return "hint"
	default:
		return "warn"
	}
}
