package analyzers

import (
	"github.com/sailpoint-oss/barrelman"
)

// RegisterGeneric registers every analyzer that is part of Barrelman's
// generic OpenAPI lint surface — naming, structure, OWASP, schema typing,
// references, markdown, etc. These rules are vendor-neutral and ship in
// the public Barrelman release.
//
// Org-specific rule packs are NOT registered here. Downstream consumers
// attach them via barrelman.RegisterPlugin and barrelman.ApplyPlugins.
func RegisterGeneric(reg *barrelman.Registry) {
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

// RegisterAll calls RegisterGeneric and then ApplyPlugins so callers that
// want the full "everything registered" behaviour continue to work when
// external rule packs are present.
//
// Downstream consumers register a RulePack via barrelman.RegisterPlugin and
// ApplyPlugins picks it up here.
func RegisterAll(reg *barrelman.Registry) {
	RegisterGeneric(reg)
	barrelman.ApplyPlugins(reg)
}
