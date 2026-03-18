package analyzers

import (
	"github.com/sailpoint-oss/barrelman"
)

// RegisterAll registers all semantic analyzers on the given registry. Each rule's
// Define().Register() call handles both metadata registration and analyzer
// registration. All rules are registered unconditionally; filtering is handled
// by the DiagnosticTransformer.
func RegisterAll(reg *barrelman.Registry) {
	registerUnresolvedRef(reg)
	registerNamingAnalyzers(reg)
	registerDocumentationAnalyzers(reg)
	registerStructureAnalyzers(reg)
	registerTypesAnalyzers(reg)
	registerSecurityAnalyzers(reg)
	registerServersAnalyzers(reg)
	registerPathsAnalyzers(reg)
	registerOWASPAnalyzers(reg)
	registerExtendedAnalyzers(reg)
	registerStructuralValidation(reg)
	registerMarkdownAnalyzers(reg)
	registerUnusedComponentAnalyzers(reg)
	registerCompletenessAnalyzers(reg)
	registerExampleValidationAnalyzers(reg)
	registerMigrationAnalyzers(reg)
}
