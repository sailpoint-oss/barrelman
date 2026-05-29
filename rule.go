package barrelman

import (
	"sync"

	"github.com/sailpoint-oss/barrelman/codemod"
	navigator "github.com/sailpoint-oss/navigator"
)

// Category groups related rules.
type Category string

const (
	CategoryNaming        Category = "naming"
	CategoryDocumentation Category = "documentation"
	CategoryStructure     Category = "structure"
	CategoryTypes         Category = "types"
	CategorySecurity      Category = "security"
	CategoryServers       Category = "servers"
	CategoryPaths         Category = "paths"
	CategoryReferences    Category = "references"
	CategorySyntax        Category = "syntax"
	CategoryOWASP         Category = "owasp"
)

// RuleMeta holds descriptive metadata for a rule.
//
// Rule IDs are stable kebab-case slugs. Rule packs may attach external
// guideline numbers in GuidelineID / GuidelineIDs when they need to link
// diagnostics back to a separate policy document.
type RuleMeta struct {
	ID           string
	Description  string
	Severity     Severity
	Category     Category
	Recommended  bool
	Formats      []navigator.Format
	HowToFix     string
	DocURL       string
	GuidelineID  int   // Primary external guideline number (0 = not linked).
	GuidelineIDs []int // Additional guideline numbers cited by this rule.
	VacuumID     string
	SpectralID   string
}

// FixFunc produces zero or more codemod Patches from a diagnostic.
// Rules attach a FixFunc via RuleBuilder.Fix; the framework calls it
// for every diagnostic the rule emitted that has an associated
// ByteRange and a source document available on the FixContext.
//
// Implementations should be idempotent: returning an empty slice when
// the underlying condition is already resolved keeps re-running the
// fixer safe.
type FixFunc func(ctx *codemod.FixContext, diag Diagnostic) ([]codemod.Patch, error)

// Rule is the self-contained unit of analysis. Every semantic analyzer, syntax
// check, spectral rule, and structural validator compiles down to this type.
//
// Fix is optional; rules without a Fix still run normally and emit
// diagnostics, they just do not contribute to auto-fix passes.
type Rule struct {
	ID   string
	Meta RuleMeta
	Run  func(ctx *AnalysisContext) []Diagnostic
	Fix  FixFunc

	// visitors is populated by RuleBuilder so lint can batch visitor-based rules
	// into a single document walk while preserving Rule.Run for compatibility.
	visitors Visitors
}

// Registry provides thread-safe storage for rule metadata AND built Rule
// instances. It replaces the old metadata-only registry.
type Registry struct {
	mu    sync.RWMutex
	metas map[string]RuleMeta
	rules []Rule
}

// NewRegistry creates an empty rule registry.
func NewRegistry() *Registry {
	return &Registry{
		metas: make(map[string]RuleMeta),
	}
}

// Register adds a fully built rule and its metadata to the registry.
func (r *Registry) Register(rule Rule) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metas[rule.ID] = rule.Meta
	r.rules = append(r.rules, rule)
}

// RegisterMeta adds only metadata (useful during migration or for rules not
// yet converted to barrelman.Rule).
func (r *Registry) RegisterMeta(meta RuleMeta) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.metas[meta.ID] = meta
}

// Get returns metadata for a rule, or false if not found.
func (r *Registry) Get(id string) (RuleMeta, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	meta, ok := r.metas[id]
	return meta, ok
}

// All returns all registered rule metadata.
func (r *Registry) All() []RuleMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]RuleMeta, 0, len(r.metas))
	for _, m := range r.metas {
		result = append(result, m)
	}
	return result
}

// AllRules returns all registered rules.
func (r *Registry) AllRules() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]Rule, len(r.rules))
	copy(out, r.rules)
	return out
}

// ByCategory returns all rule metadata in a given category.
func (r *Registry) ByCategory(cat Category) []RuleMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []RuleMeta
	for _, m := range r.metas {
		if m.Category == cat {
			result = append(result, m)
		}
	}
	return result
}

// Recommended returns only rules marked as recommended.
func (r *Registry) Recommended() []RuleMeta {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var result []RuleMeta
	for _, m := range r.metas {
		if m.Recommended {
			result = append(result, m)
		}
	}
	return result
}

// IDs returns all registered rule IDs.
func (r *Registry) IDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.metas))
	for id := range r.metas {
		ids = append(ids, id)
	}
	return ids
}

// DefaultRegistry is the global registry populated by init functions in the
// analyzers/ and checks/ sub-packages.
var DefaultRegistry = NewRegistry()

// Source is the diagnostic source string.
const Source = "telescope"

// DocBaseURL is the base URL for rule documentation.
const DocBaseURL = "https://telescope.dev/rules/"
